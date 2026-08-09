package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	agentCopilotInvocationRoute      = "/v1/agent-copilot/invocations"
	agentCopilotInvokeScope          = "agent_copilot:invoke"
	agentCopilotInvocationMaxRuntime = 30 * time.Second
	agentCopilotMaximumResponseBytes = 256 * 1024

	AgentCopilotInvocationFailureInputInvalid     = "agent_copilot_invocation_input_invalid"
	AgentCopilotInvocationFailureDuplicateRunning = "agent_copilot_invocation_duplicate_running"
	AgentCopilotInvocationFailureCanceled         = "agent_copilot_invocation_canceled"
	AgentCopilotInvocationFailureOutcomeUnknown   = "agent_copilot_invocation_outcome_unknown"
	AgentCopilotInvocationFailureResponseContract = "agent_copilot_response_contract_failed"
)

type AgentCopilotArtifact struct {
	Kind     string         `json:"kind"`
	Role     string         `json:"role"`
	Name     string         `json:"name"`
	MIMEType string         `json:"mime_type"`
	URI      string         `json:"uri,omitempty"`
	Content  any            `json:"content,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type AgentCopilotInvocationInput struct {
	Task                string
	Locale              string
	ConversationID      string
	Artifacts           []AgentCopilotArtifact
	Context             map[string]any
	ClientInvocationKey string
}

type AgentCopilotInvocationResult struct {
	Run              *WorkflowRunRecord
	Response         *AgentCopilotResponse
	FailureCode      string
	FailureSummary   string
	IdempotentReplay bool
}

type AgentCopilotResponse struct {
	SchemaVersion        int                            `json:"schema_version"`
	Status               string                         `json:"status"`
	Project              string                         `json:"project"`
	Task                 string                         `json:"task"`
	Summary              string                         `json:"summary"`
	Answers              []AgentCopilotResponseAnswer   `json:"answers"`
	Issues               []AgentCopilotResponseIssue    `json:"issues"`
	ProposedActions      []AgentCopilotResponseAction   `json:"proposed_actions"`
	Citations            []AgentCopilotResponseCitation `json:"citations"`
	Confidence           float64                        `json:"confidence"`
	RiskLevel            string                         `json:"risk_level"`
	RequiresConfirmation bool                           `json:"requires_confirmation"`
}

type AgentCopilotResponseAnswer struct {
	Kind        string   `json:"kind,omitempty"`
	Text        string   `json:"text"`
	CitationIDs []string `json:"citation_ids,omitempty"`
}

type AgentCopilotResponseIssue struct {
	Code        string   `json:"code,omitempty"`
	Message     string   `json:"message"`
	Severity    string   `json:"severity"`
	CitationIDs []string `json:"citation_ids,omitempty"`
}

type AgentCopilotResponseAction struct {
	Kind                 string         `json:"kind"`
	Title                string         `json:"title"`
	Target               map[string]any `json:"target,omitempty"`
	Rationale            string         `json:"rationale"`
	Patch                map[string]any `json:"patch,omitempty"`
	Preview              map[string]any `json:"preview,omitempty"`
	Apply                map[string]any `json:"apply,omitempty"`
	RiskLevel            string         `json:"risk_level"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
	CitationIDs          []string       `json:"citation_ids,omitempty"`
}

type AgentCopilotResponseCitation struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Label     string `json:"label"`
	Locator   string `json:"locator,omitempty"`
	Excerpt   string `json:"excerpt,omitempty"`
	SourceURI string `json:"source_uri,omitempty"`
}

type agentCopilotCanonicalRequest struct {
	SchemaVersion  int                    `json:"schema_version"`
	RequestID      string                 `json:"request_id,omitempty"`
	Project        string                 `json:"project"`
	Task           string                 `json:"task"`
	Locale         string                 `json:"locale"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Artifacts      []AgentCopilotArtifact `json:"artifacts"`
	Context        map[string]any         `json:"context"`
	ToolHints      agentCopilotToolHints  `json:"tool_hints"`
	Safety         agentCopilotSafety     `json:"safety"`
}

type agentCopilotToolHints struct {
	AllowRetrieval      bool `json:"allow_retrieval"`
	AllowToolCalls      bool `json:"allow_tool_calls"`
	AllowImageReasoning bool `json:"allow_image_reasoning"`
}

type agentCopilotSafety struct {
	Mode                           string `json:"mode"`
	RequiresConfirmationForActions bool   `json:"requires_confirmation_for_actions"`
}

type agentCopilotInvocationAuthority struct {
	Assignment AgentCopilotRuntimeAssignmentV1
	Resolved   agentCopilotRuntimeAuthority
	Selection  northboundSelection
	Snapshot   AgentCopilotRuntimeAuthorityV3
}

type agentCopilotInvocationService struct {
	runtimeRepository agentCopilotRuntimeRepository
	authorityResolver agentCopilotRuntimeAuthorityResolver
	catalogRepository applicationCatalogRepository
	runStore          workflowRunStore
	bridge            bridgeClient
	resolveSelection  func(context.Context, string) northboundSelection
	envelopeOptions   func(northboundSelection, float64) bridge.EnvelopeOptions
	maxRuntime        time.Duration
	now               func() time.Time
}

func newAgentCopilotInvocationService(
	runtimeRepository agentCopilotRuntimeRepository,
	authorityResolver agentCopilotRuntimeAuthorityResolver,
	catalogRepository applicationCatalogRepository,
	runStore workflowRunStore,
	bridgeClient bridgeClient,
) agentCopilotInvocationService {
	return agentCopilotInvocationService{
		runtimeRepository: runtimeRepository, authorityResolver: authorityResolver, catalogRepository: catalogRepository,
		runStore: runStore, bridge: bridgeClient, maxRuntime: agentCopilotInvocationMaxRuntime,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (service agentCopilotInvocationService) Invoke(ctx AgentCopilotRuntimeContext, input AgentCopilotInvocationInput) AgentCopilotInvocationResult {
	if validateAgentCopilotRuntimeContext(ctx) != nil {
		return agentCopilotInvocationFailure(AgentCopilotRuntimeFailureScopeDenied)
	}
	clientKey := strings.TrimSpace(input.ClientInvocationKey)
	if !validPromptApplicationRef(clientKey) || len(clientKey) > 160 {
		return agentCopilotInvocationFailure(AgentCopilotInvocationFailureInputInvalid)
	}
	authority, failure := service.resolveAuthority(ctx)
	if failure != "" {
		return agentCopilotInvocationFailure(failure)
	}
	request, requestPayload, contextBytes, artifactBytes, failure := canonicalAgentCopilotInvocationRequest(ctx, authority.Resolved.Profile, input)
	if failure != "" {
		return agentCopilotInvocationFailure(failure)
	}
	runContext := agentCopilotWorkflowRunContext(ctx)
	runID := agentCopilotInvocationRunID(ctx, clientKey)
	existing, found, err := service.runStore.ReadRun(runContext, runID)
	if err != nil {
		return agentCopilotInvocationFailure(AgentCopilotRuntimeFailureStoreUnavailable)
	}
	inputDigest := workflowRAGSHA256(string(requestPayload))
	if found {
		return agentCopilotDuplicateInvocation(existing, inputDigest)
	}
	record := newAgentCopilotRunRecord(ctx, runID, request, inputDigest, len(requestPayload), contextBytes, artifactBytes, authority, service.currentTime())
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		if errors.Is(err, errWorkflowRunStoreConflict) {
			existing, found, readErr := service.runStore.ReadRun(runContext, runID)
			if readErr == nil && found {
				return agentCopilotDuplicateInvocation(existing, inputDigest)
			}
		}
		return agentCopilotInvocationFailure(AgentCopilotRuntimeFailureStoreUnavailable)
	}
	checkpoint, failure := service.resolveAuthority(ctx)
	if failure != "" || checkpoint.Snapshot.AuthorityDigest != authority.Snapshot.AuthorityDigest {
		return service.complete(runContext, record, nil, WorkflowRunStatusFailed, AgentCopilotRuntimeFailureAuthorityChanged, "authority", "authority", "none")
	}
	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = agentCopilotInvocationMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(ctx.RequestContext, maxRuntime)
	defer cancel()
	responsePayload, gatewayCategory, gatewayFailure := service.callGateway(executionContext, runID, checkpoint, requestPayload)
	if gatewayRequestQuotaFailureCodeFromValue(gatewayFailure) != "" {
		return service.complete(runContext, record, nil, WorkflowRunStatusFailed, gatewayFailure, "quota_admission", "provider", "quota")
	}
	record.SideEffects.ProviderCalls = 1
	if gatewayFailure != "" {
		if gatewayCategory == "canceled" {
			return service.complete(runContext, record, nil, WorkflowRunStatusCanceled, AgentCopilotInvocationFailureCanceled, "gateway", "provider", gatewayCategory)
		}
		if gatewayCategory == "unavailable" || gatewayCategory == "timeout" || gatewayCategory == "worker_crash" {
			return service.complete(runContext, record, nil, WorkflowRunStatusOutcomeUnknown, AgentCopilotInvocationFailureOutcomeUnknown, "gateway", "provider", gatewayCategory)
		}
		return service.complete(runContext, record, nil, WorkflowRunStatusFailed, AgentCopilotInvocationFailureOutcomeUnknown, "gateway", "provider", gatewayCategory)
	}
	response, failure := decodeAndValidateAgentCopilotResponse(responsePayload, request, checkpoint.Resolved.Profile.ResponsePolicy)
	if failure != "" {
		return service.complete(runContext, record, nil, WorkflowRunStatusFailed, failure, "response_contract", "output", "output_unavailable")
	}
	return service.complete(runContext, record, &response, WorkflowRunStatusSucceeded, "", "", "", "none")
}

func (service agentCopilotInvocationService) resolveAuthority(ctx AgentCopilotRuntimeContext) (agentCopilotInvocationAuthority, string) {
	assignment, _, err := service.runtimeRepository.Read(ctx)
	if errors.Is(err, errAgentCopilotRuntimeNotFound) {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureNotFound
	}
	if err != nil {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureStoreUnavailable
	}
	if assignment.State != "active" {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	resolved, failure := service.authorityResolver.Resolve(ctx, assignment.CandidateID, &assignment)
	if failure != "" {
		return agentCopilotInvocationAuthority{}, failure
	}
	catalog, err := service.catalogRepository.Read(ApplicationCatalogContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}, ctx.ApplicationID)
	if err != nil || catalog.ApplicationKind != "agent" || catalog.LifecycleState != applicationCatalogLifecycleActive {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	selection := normalizePromptApplicationSelection(service.selection(ctx.RequestContext, resolved.Candidate.Configuration.DefaultModel))
	if !promptApplicationSelectionEligible(resolved.Candidate.Configuration.DefaultProtocol, resolved.Candidate.Configuration.DefaultModel, selection) {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	protocolDigest, err := promptApplicationProtocolPolicyDigest(resolved.Candidate.Configuration.DefaultProtocol, resolved.Candidate.Configuration.AllowedProtocols)
	if err != nil {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureStoreContract
	}
	modelDigest, err := promptApplicationModelEligibilityDigest(selection)
	if err != nil {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureStoreContract
	}
	compiled, findings := CompileAgentCopilotProfileSource(resolved.Profile.AgentCopilotProfileSource)
	if len(findings) != 0 || compiled.ProfileDigest != resolved.Profile.ProfileDigest || compiled.PolicyDigest != resolved.Profile.PolicyDigest {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureAuthorityChanged
	}
	snapshot := AgentCopilotRuntimeAuthorityV3{
		SchemaVersion: agentCopilotRuntimeAuthorityV3Schema, ExecutionProfile: agentCopilotSuggestionProfile,
		ApplicationID: catalog.ApplicationID, ApplicationRecordVersion: catalog.RecordVersion,
		ApplicationLifecycle: catalog.LifecycleState,
		AgentCopilot: AgentCopilotAuthorityV3{
			AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.AssignmentVersion,
			AssignmentDigest: assignment.AssignmentDigest, PublishCandidateID: resolved.Candidate.CandidateID,
			PublishReviewVersion: resolved.Candidate.ReviewVersion, DraftID: resolved.Draft.DraftID,
			DraftVersion: resolved.Draft.DraftVersion, DraftDigest: resolved.Candidate.DraftDigest,
			AgentCopilotProfileRef: assignment.AgentCopilotProfileRef, Project: compiled.Source.Project,
			AllowedTasksDigest: compiled.AllowedTasksDigest, DefaultProtocol: resolved.Candidate.Configuration.DefaultProtocol,
			DefaultModel: resolved.Candidate.Configuration.DefaultModel, ProtocolPolicyDigest: protocolDigest,
			ModelEligibilityDigest: modelDigest,
		},
	}
	snapshot.AuthorityDigest, err = agentCopilotAuthorityDigest(snapshot)
	if err != nil || validateAgentCopilotRuntimeAuthority(snapshot) != nil {
		return agentCopilotInvocationAuthority{}, AgentCopilotRuntimeFailureStoreContract
	}
	return agentCopilotInvocationAuthority{Assignment: assignment, Resolved: resolved, Selection: selection, Snapshot: snapshot}, ""
}

func canonicalAgentCopilotInvocationRequest(
	ctx AgentCopilotRuntimeContext,
	profile AgentCopilotProfileVersionV1,
	input AgentCopilotInvocationInput,
) (agentCopilotCanonicalRequest, []byte, int, int, string) {
	input.Task = strings.TrimSpace(input.Task)
	input.Locale = normalizeAgentCopilotLocale(input.Locale)
	input.ConversationID = strings.TrimSpace(input.ConversationID)
	if !agentCopilotContainsString(profile.AllowedTasks, input.Task) ||
		!agentCopilotContainsString(profile.AllowedLocales, input.Locale) ||
		input.ConversationID != "" && !validPromptApplicationRef(input.ConversationID) {
		return agentCopilotCanonicalRequest{}, nil, 0, 0, AgentCopilotInvocationFailureInputInvalid
	}
	contextPayload, err := json.Marshal(input.Context)
	if err != nil || input.Context == nil || len(contextPayload) > profile.ContextPolicy.MaxBytes ||
		len(contextPayload) > agentCopilotMaximumContextBytes ||
		!validAgentCopilotInvocationContext(profile, input.Task, input.Context) {
		return agentCopilotCanonicalRequest{}, nil, 0, 0, AgentCopilotInvocationFailureInputInvalid
	}
	artifacts, artifactBytes, valid := canonicalAgentCopilotArtifacts(profile.ArtifactPolicy, input.Artifacts)
	if !valid {
		return agentCopilotCanonicalRequest{}, nil, 0, 0, AgentCopilotInvocationFailureInputInvalid
	}
	request := agentCopilotCanonicalRequest{
		SchemaVersion: 1, RequestID: ctx.RequestID, Project: profile.Project, Task: input.Task,
		Locale: input.Locale, ConversationID: input.ConversationID, Artifacts: artifacts, Context: input.Context,
		ToolHints: agentCopilotToolHints{}, Safety: agentCopilotSafety{Mode: "advisory", RequiresConfirmationForActions: true},
	}
	payload, err := json.Marshal(request)
	if err != nil || len(payload) > agentCopilotMaximumInvocationBytes {
		return agentCopilotCanonicalRequest{}, nil, 0, 0, AgentCopilotInvocationFailureInputInvalid
	}
	return request, payload, len(contextPayload), artifactBytes, ""
}

func validAgentCopilotInvocationContext(profile AgentCopilotProfileVersionV1, task string, value map[string]any) bool {
	if profile.ContextPolicy.RequireTaskContext && len(value) == 0 {
		return false
	}
	for field := range value {
		if !agentCopilotContainsString(profile.ContextPolicy.AllowedFields, field) {
			return false
		}
	}
	if task != "suggest_ghost_completion" {
		return true
	}
	for _, field := range []string{"document_revision", "selected_unit_ids", "legal_candidate_completions"} {
		if _, ok := value[field]; !ok {
			return false
		}
	}
	_, unconnected := value["unconnected_ports"]
	_, missing := value["missing_canonical_ports"]
	selected, selectedOK := value["selected_unit_ids"].([]any)
	return (unconnected || missing) && selectedOK && len(selected) == 1
}

func canonicalAgentCopilotArtifacts(policy AgentCopilotArtifactPolicy, values []AgentCopilotArtifact) ([]AgentCopilotArtifact, int, bool) {
	if len(values) > policy.MaxCount || len(values) > agentCopilotMaximumArtifacts {
		return nil, 0, false
	}
	result := make([]AgentCopilotArtifact, 0, len(values))
	total := 0
	for _, value := range values {
		value.Kind, value.Role = strings.TrimSpace(value.Kind), strings.TrimSpace(value.Role)
		value.Name, value.MIMEType, value.URI = strings.TrimSpace(value.Name), strings.TrimSpace(value.MIMEType), strings.TrimSpace(value.URI)
		if !agentCopilotContainsString(policy.AllowedKinds, value.Kind) || !agentCopilotContainsString(policy.AllowedRoles, value.Role) ||
			value.Name == "" || value.MIMEType == "" || value.URI == "" && value.Content == nil {
			return nil, 0, false
		}
		itemBytes := 0
		if value.Content != nil {
			payload, err := json.Marshal(value.Content)
			if err != nil {
				return nil, 0, false
			}
			itemBytes = len(payload)
		}
		if itemBytes > policy.MaxItemBytes || itemBytes > agentCopilotMaximumArtifactItemBytes {
			return nil, 0, false
		}
		total += itemBytes
		if total > policy.MaxTotalBytes || total > agentCopilotMaximumArtifactTotalBytes {
			return nil, 0, false
		}
		result = append(result, value)
	}
	return result, total, true
}

func decodeAndValidateAgentCopilotResponse(payload []byte, request agentCopilotCanonicalRequest, policy AgentCopilotResponsePolicy) (AgentCopilotResponse, string) {
	if len(payload) == 0 || len(payload) > agentCopilotMaximumResponseBytes {
		return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
	}
	var response AgentCopilotResponse
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		response.SchemaVersion != 1 || response.Project != request.Project || response.Task != request.Task ||
		!agentCopilotContainsString([]string{"ok", "partial"}, response.Status) ||
		!agentCopilotContainsString([]string{"low", "medium", "high"}, response.RiskLevel) ||
		response.Confidence < 0 || response.Confidence > 1 ||
		len(response.Answers) > policy.MaxAnswers || len(response.Issues) > policy.MaxIssues ||
		len(response.ProposedActions) > policy.MaxActions || len(response.Citations) > policy.MaxCitations ||
		!validAgentCopilotVisibleText(response.Summary, policy.MaxVisibleTextBytes) {
		return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
	}
	actions := make([]AgentCopilotCandidateActionSafety, 0, len(response.ProposedActions))
	citationIDs := make(map[string]struct{}, len(response.Citations))
	for _, citation := range response.Citations {
		citationID := strings.TrimSpace(citation.ID)
		if citationID == "" || !validAgentCopilotVisibleText(citation.Label, policy.MaxVisibleTextBytes) ||
			citation.Excerpt != "" && len([]byte(citation.Excerpt)) > policy.MaxVisibleTextBytes ||
			!agentCopilotContainsString([]string{"artifact", "context", "retrieval", "rule"}, citation.Kind) {
			return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
		}
		if _, duplicate := citationIDs[citationID]; duplicate {
			return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
		}
		citationIDs[citationID] = struct{}{}
	}
	for _, answer := range response.Answers {
		if !validAgentCopilotVisibleText(answer.Text, policy.MaxVisibleTextBytes) ||
			!validAgentCopilotResponseCitationRefs(answer.CitationIDs, citationIDs) {
			return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
		}
	}
	for _, issue := range response.Issues {
		if !validAgentCopilotVisibleText(issue.Message, policy.MaxVisibleTextBytes) ||
			!agentCopilotContainsString([]string{"info", "warning", "error"}, issue.Severity) ||
			!validAgentCopilotResponseCitationRefs(issue.CitationIDs, citationIDs) {
			return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
		}
	}
	for _, action := range response.ProposedActions {
		if !agentCopilotContainsString(policy.AllowedActionKinds, action.Kind) ||
			!validAgentCopilotVisibleText(action.Title, policy.MaxVisibleTextBytes) ||
			!validAgentCopilotVisibleText(action.Rationale, policy.MaxVisibleTextBytes) ||
			!agentCopilotContainsString([]string{"low", "medium", "high"}, action.RiskLevel) ||
			action.Kind == "ghost_completion" && (action.Patch == nil || action.Preview == nil || action.Apply == nil) ||
			!validAgentCopilotResponseCitationRefs(action.CitationIDs, citationIDs) {
			return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
		}
		actions = append(actions, AgentCopilotCandidateActionSafety{Kind: action.Kind, RequiresConfirmation: action.RequiresConfirmation})
	}
	if !ValidateAgentCopilotResponseConfirmation(actions, response.RequiresConfirmation) {
		return AgentCopilotResponse{}, AgentCopilotInvocationFailureResponseContract
	}
	return response, ""
}

func validAgentCopilotResponseCitationRefs(values []string, known map[string]struct{}) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return false
		}
		if _, exists := known[value]; !exists {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validAgentCopilotVisibleText(value string, maximum int) bool {
	return strings.TrimSpace(value) != "" && len([]byte(value)) <= maximum && len([]byte(value)) <= agentCopilotMaximumVisibleResponseTextByte
}

func (service agentCopilotInvocationService) callGateway(ctx context.Context, runID string, authority agentCopilotInvocationAuthority, requestPayload []byte) ([]byte, string, string) {
	if strings.TrimSpace(runID) == "" || len(requestPayload) == 0 {
		return nil, "protocol", AgentCopilotInvocationFailureOutcomeUnknown
	}
	envelope, err := service.bridge.HandleEnvelope(ctx, requestPayload, service.gatewayOptions(authority.Selection))
	if err != nil {
		if quotaFailure := gatewayRequestQuotaFailureCode(err); quotaFailure != "" {
			return nil, "quota", quotaFailure
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, "canceled", AgentCopilotInvocationFailureCanceled
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, "timeout", AgentCopilotInvocationFailureOutcomeUnknown
		}
		return nil, "unavailable", AgentCopilotInvocationFailureOutcomeUnknown
	}
	gatewayStatus := strings.ToLower(strings.TrimSpace(envelope.Status))
	if (gatewayStatus != "ok" && gatewayStatus != "partial") || envelope.Error != nil || envelope.Response == nil {
		return nil, "provider_failed", AgentCopilotInvocationFailureOutcomeUnknown
	}
	payload, marshalErr := json.Marshal(envelope.Response)
	if marshalErr != nil {
		return nil, "output_unavailable", AgentCopilotInvocationFailureResponseContract
	}
	return payload, "none", ""
}

func (service agentCopilotInvocationService) complete(
	runContext WorkflowRunContext,
	record WorkflowRunRecord,
	response *AgentCopilotResponse,
	status WorkflowRunStatus,
	failureCode, boundary, stage, gatewayCategory string,
) AgentCopilotInvocationResult {
	completedAt := service.currentTime()
	record.Status, record.CompletedAt = status, workflowRunTimestamp(completedAt)
	record.FailureCode, record.FailureSummary = WorkflowRunFailureCode(failureCode), ""
	if failureCode != "" {
		record.FailureSummary = agentCopilotInvocationFailureSummary(failureCode)
	}
	record.AgentResponseStatus, record.AgentResponseDigest = "unavailable", ""
	record.AgentAnswerCount, record.AgentIssueCount, record.AgentActionCount, record.AgentCitationCount = 0, 0, 0, 0
	record.AgentRiskLevel, record.AgentRequiresConfirmation = "low", false
	if response != nil {
		payload, _ := json.Marshal(response)
		record.AgentResponseStatus, record.AgentResponseDigest = response.Status, workflowRAGSHA256(string(payload))
		record.AgentAnswerCount, record.AgentIssueCount = len(response.Answers), len(response.Issues)
		record.AgentActionCount, record.AgentCitationCount = len(response.ProposedActions), len(response.Citations)
		record.AgentRiskLevel, record.AgentRequiresConfirmation = response.RiskLevel, response.RequiresConfirmation
	}
	record.PromptDiagnostic = &PromptApplicationRunDiagnosticV6{
		FailureBoundary: boundary, FailureStage: stage, TerminalWriteState: "stored",
		GatewayFailureCategory: gatewayCategory, Summary: record.FailureSummary,
		RecommendedReviewAction: agentCopilotInvocationReviewAction(failureCode), ObservedAt: workflowRunTimestamp(completedAt),
	}
	if err := service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Status = WorkflowRunStatusOutcomeUnknown
		record.FailureCode = WorkflowRunFailureCode(AgentCopilotInvocationFailureOutcomeUnknown)
		record.FailureSummary = agentCopilotInvocationFailureSummary(AgentCopilotInvocationFailureOutcomeUnknown)
		record.PromptDiagnostic.TerminalWriteState, record.PromptDiagnostic.Summary = "unknown", record.FailureSummary
		return AgentCopilotInvocationResult{Run: &record, FailureCode: AgentCopilotInvocationFailureOutcomeUnknown, FailureSummary: record.FailureSummary}
	}
	return AgentCopilotInvocationResult{Run: &record, Response: response, FailureCode: failureCode, FailureSummary: record.FailureSummary}
}

func newAgentCopilotRunRecord(
	ctx AgentCopilotRuntimeContext,
	runID string,
	request agentCopilotCanonicalRequest,
	inputDigest string,
	inputBytes, contextBytes, artifactBytes int,
	authority agentCopilotInvocationAuthority,
	startedAt time.Time,
) WorkflowRunRecord {
	selection := authority.Selection
	return WorkflowRunRecord{
		SchemaVersion: agentCopilotRunV7Schema, RunID: runID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ExecutionSource:  &workflowRunExecutionSource{Kind: "agent_copilot_suggestion", SourceKind: agentCopilotExecutionSourceKind, ID: authority.Resolved.Profile.ProfileID, Version: authority.Resolved.Profile.ProfileVersion},
		ExecutionProfile: agentCopilotSuggestionProfile, AgentCopilotAuthority: &authority.Snapshot,
		AgentProject: request.Project, AgentTask: request.Task, AgentLocale: request.Locale,
		InputDigest: inputDigest, InputBytes: inputBytes, AgentContextBytes: contextBytes,
		AgentArtifactCount: len(request.Artifacts), AgentArtifactBytes: artifactBytes,
		RequestedProtocol: authority.Resolved.Candidate.Configuration.DefaultProtocol,
		SelectedProtocol:  authority.Resolved.Candidate.Configuration.DefaultProtocol,
		RequestedModel:    authority.Resolved.Candidate.Configuration.DefaultModel,
		SelectedProvider:  selection.provider, SelectedProfile: selection.providerProfile,
		SelectedModel: selection.model, UpstreamModel: selection.upstreamModel, SelectionSource: selection.source,
		AgentResponseStatus: "unavailable", AgentRiskLevel: "low", Status: WorkflowRunStatusRunning,
		StartedAt: workflowRunTimestamp(startedAt), PromptUsage: PromptApplicationRunUsageV6{State: "unavailable"},
		PromptDiagnostic: &PromptApplicationRunDiagnosticV6{
			TerminalWriteState: "pending", GatewayFailureCategory: "none", ObservedAt: workflowRunTimestamp(startedAt),
		},
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, ActorRef: ctx.ActorRef,
	}
}

func agentCopilotInvocationRunID(ctx AgentCopilotRuntimeContext, clientKey string) string {
	sum := sha256.Sum256([]byte(ctx.TenantRef + "\x00" + ctx.WorkspaceID + "\x00" + ctx.ApplicationID + "\x00" + clientKey))
	return "run_" + hex.EncodeToString(sum[:])
}

func agentCopilotDuplicateInvocation(record WorkflowRunRecord, inputDigest string) AgentCopilotInvocationResult {
	if record.SchemaVersion != agentCopilotRunV7Schema || record.InputDigest != inputDigest {
		return agentCopilotInvocationFailure(AgentCopilotInvocationFailureInputInvalid)
	}
	if record.Status == WorkflowRunStatusRunning {
		result := agentCopilotInvocationFailure(AgentCopilotInvocationFailureDuplicateRunning)
		result.Run, result.IdempotentReplay = &record, true
		return result
	}
	result := AgentCopilotInvocationResult{Run: &record, IdempotentReplay: true}
	if record.Status != WorkflowRunStatusSucceeded {
		result.FailureCode, result.FailureSummary = string(record.FailureCode), record.FailureSummary
	}
	return result
}

func (service agentCopilotInvocationService) selection(ctx context.Context, model string) northboundSelection {
	if service.resolveSelection != nil {
		return service.resolveSelection(ctx, model)
	}
	return (&Server{bridge: service.bridge}).resolveNorthboundSelection(ctx, model, nil)
}

func (service agentCopilotInvocationService) gatewayOptions(selection northboundSelection) bridge.EnvelopeOptions {
	if service.envelopeOptions != nil {
		return service.envelopeOptions(selection, 0)
	}
	return (&Server{bridge: service.bridge}).buildBridgeEnvelopeOptions(selection, 0)
}

func (service agentCopilotInvocationService) currentTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func agentCopilotWorkflowRunContext(ctx AgentCopilotRuntimeContext) WorkflowRunContext {
	return WorkflowRunContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef}
}

func agentCopilotInvocationFailure(code string) AgentCopilotInvocationResult {
	return AgentCopilotInvocationResult{FailureCode: code, FailureSummary: agentCopilotInvocationFailureSummary(code)}
}

func agentCopilotInvocationFailureSummary(code string) string {
	switch code {
	case AgentCopilotRuntimeFailureScopeDenied:
		return "Agent Copilot invocation scope is denied."
	case AgentCopilotRuntimeFailureNotFound:
		return "No active Agent Copilot runtime assignment was found."
	case AgentCopilotRuntimeFailureAuthorityChanged, AgentCopilotRuntimeFailureCandidate:
		return "Agent Copilot runtime authority is no longer eligible."
	case AgentCopilotInvocationFailureInputInvalid:
		return "Agent Copilot invocation input is invalid."
	case AgentCopilotInvocationFailureDuplicateRunning:
		return "Agent Copilot invocation is already running."
	case AgentCopilotInvocationFailureCanceled:
		return "Agent Copilot invocation was canceled."
	case AgentCopilotInvocationFailureResponseContract:
		return "Agent Copilot response did not satisfy the canonical contract."
	case AgentCopilotInvocationFailureOutcomeUnknown:
		return "Agent Copilot invocation outcome is unknown and was not replayed."
	case GatewayRequestQuotaFailureExceeded:
		return "Agent Copilot request quota is exhausted for the current UTC period."
	case GatewayRequestQuotaFailurePolicyNotFound, GatewayRequestQuotaFailureStoreUnavailable:
		return "Agent Copilot request quota is unavailable."
	case GatewayRequestQuotaFailureAttemptConflict:
		return "Agent Copilot request quota admission conflicts with an existing attempt."
	default:
		return "Agent Copilot runtime store is unavailable."
	}
}

func agentCopilotInvocationReviewAction(code string) string {
	switch code {
	case "":
		return ""
	case AgentCopilotRuntimeFailureAuthorityChanged, AgentCopilotRuntimeFailureCandidate:
		return "review_authority"
	case AgentCopilotInvocationFailureResponseContract:
		return "review_response_contract"
	case AgentCopilotInvocationFailureCanceled:
		return "review_cancellation"
	case GatewayRequestQuotaFailureExceeded, GatewayRequestQuotaFailurePolicyNotFound,
		GatewayRequestQuotaFailureAttemptConflict, GatewayRequestQuotaFailureStoreUnavailable:
		return "review_run"
	default:
		return "review_run"
	}
}
