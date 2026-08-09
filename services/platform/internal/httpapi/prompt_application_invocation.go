package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	promptApplicationInvocationRoute      = "/v1/prompt-applications/invocations"
	promptApplicationInvocationProtocol   = "prompt-application-invocation-v1"
	promptApplicationInvocationMaxRuntime = 30 * time.Second

	PromptApplicationInvocationFailureInputInvalid     = "prompt_invocation_input_invalid"
	PromptApplicationInvocationFailureDuplicateRunning = "prompt_invocation_duplicate_running"
	PromptApplicationInvocationFailureCanceled         = "prompt_invocation_canceled"
	PromptApplicationInvocationFailureOutcomeUnknown   = "prompt_invocation_outcome_unknown"
)

type PromptApplicationInvocationInput struct {
	Variables           map[string]any
	ClientInvocationKey string
}

type PromptApplicationInvocationResult struct {
	Run              *WorkflowRunRecord
	Output           string
	FailureCode      string
	FailureSummary   string
	IdempotentReplay bool
}

type promptApplicationInvocationAuthority struct {
	Assignment PromptApplicationRuntimeAssignment
	Resolved   promptApplicationRuntimeAuthority
	Catalog    ApplicationCatalogRecord
	Selection  northboundSelection
	Snapshot   PromptApplicationRuntimeAuthorityV2
}

type promptApplicationInvocationService struct {
	runtimeRepository promptApplicationRuntimeRepository
	authorityResolver promptApplicationRuntimeAuthorityResolver
	catalogRepository applicationCatalogRepository
	runStore          workflowRunStore
	bridge            bridgeClient
	resolveSelection  func(context.Context, string) northboundSelection
	envelopeOptions   func(northboundSelection, float64) bridge.EnvelopeOptions
	maxRuntime        time.Duration
	now               func() time.Time
}

func newPromptApplicationInvocationService(
	runtimeRepository promptApplicationRuntimeRepository,
	authorityResolver promptApplicationRuntimeAuthorityResolver,
	catalogRepository applicationCatalogRepository,
	runStore workflowRunStore,
	bridgeClient bridgeClient,
) promptApplicationInvocationService {
	return promptApplicationInvocationService{
		runtimeRepository: runtimeRepository, authorityResolver: authorityResolver,
		catalogRepository: catalogRepository, runStore: runStore, bridge: bridgeClient,
		maxRuntime: promptApplicationInvocationMaxRuntime, now: func() time.Time { return time.Now().UTC() },
	}
}

func (service promptApplicationInvocationService) Invoke(ctx PromptApplicationRuntimeContext, input PromptApplicationInvocationInput) PromptApplicationInvocationResult {
	if validatePromptApplicationRuntimeContext(ctx) != nil {
		return promptApplicationInvocationFailure(PromptApplicationRuntimeFailureScopeDenied)
	}
	clientKey := strings.TrimSpace(input.ClientInvocationKey)
	if !validPromptApplicationRef(clientKey) || len(clientKey) > 160 || input.Variables == nil {
		return promptApplicationInvocationFailure(PromptApplicationInvocationFailureInputInvalid)
	}
	variablePayload, names, err := canonicalPromptApplicationInvocationInput(input.Variables)
	if err != nil {
		return promptApplicationInvocationFailure(PromptApplicationInvocationFailureInputInvalid)
	}
	runContext := promptApplicationWorkflowRunContext(ctx)
	runID := promptApplicationInvocationRunID(ctx, clientKey)
	existing, found, err := service.runStore.ReadRun(runContext, runID)
	if err != nil {
		return promptApplicationInvocationFailure(PromptApplicationRuntimeFailureStoreUnavailable)
	}
	if found {
		return promptApplicationDuplicateInvocation(existing, workflowRAGSHA256(string(variablePayload)))
	}

	authority, failure := service.resolveAuthority(ctx)
	if failure != "" {
		return promptApplicationInvocationFailure(failure)
	}
	rendered := RenderPromptApplicationTemplate(authority.Resolved.Template.PromptApplicationTemplateSource, input.Variables)
	if rendered.FailureCode != "" {
		return promptApplicationInvocationFailure(PromptApplicationInvocationFailureInputInvalid)
	}
	namesDigest, err := calculatePromptApplicationVariableNamesDigest(names)
	if err != nil {
		return promptApplicationInvocationFailure(PromptApplicationRuntimeFailureStoreUnavailable)
	}
	startedAt := service.currentTime()
	record := newPromptApplicationRunRecord(
		ctx, runID, workflowRAGSHA256(string(variablePayload)), len(variablePayload), names, namesDigest, authority, startedAt,
	)
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		if errors.Is(err, errWorkflowRunStoreConflict) {
			existing, found, readErr := service.runStore.ReadRun(runContext, runID)
			if readErr == nil && found {
				return promptApplicationDuplicateInvocation(existing, record.InputDigest)
			}
		}
		return promptApplicationInvocationFailure(PromptApplicationRuntimeFailureStoreUnavailable)
	}
	checkpoint, failure := service.resolveAuthority(ctx)
	if failure != "" || checkpoint.Snapshot.AuthorityDigest != authority.Snapshot.AuthorityDigest {
		return service.complete(runContext, record, WorkflowRunStatusFailed, PromptApplicationRuntimeFailureAuthorityChanged, "authority", "authority", "none")
	}

	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = promptApplicationInvocationMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(ctx.RequestContext, maxRuntime)
	defer cancel()
	output, gatewayCategory, gatewayFailure := service.callGateway(executionContext, runID, checkpoint, rendered.Messages)
	if gatewayRequestQuotaFailureCodeFromValue(gatewayFailure) != "" {
		return service.complete(runContext, record, WorkflowRunStatusFailed, gatewayFailure, "quota_admission", "provider", "quota")
	}
	record.SideEffects.ProviderCalls = 1
	if gatewayFailure != "" {
		if gatewayCategory == "canceled" {
			return service.complete(runContext, record, WorkflowRunStatusCanceled, PromptApplicationInvocationFailureCanceled, "gateway", "provider", gatewayCategory)
		}
		if gatewayCategory == "unavailable" || gatewayCategory == "timeout" || gatewayCategory == "worker_crash" {
			return service.complete(runContext, record, WorkflowRunStatusOutcomeUnknown, PromptApplicationInvocationFailureOutcomeUnknown, "gateway", "provider", gatewayCategory)
		}
		return service.complete(runContext, record, WorkflowRunStatusFailed, PromptApplicationInvocationFailureOutcomeUnknown, "gateway", "provider", gatewayCategory)
	}
	if failure = ValidatePromptApplicationOutput(checkpoint.Resolved.Template.OutputContract, output); failure != "" {
		return service.complete(runContext, record, WorkflowRunStatusFailed, PromptApplicationInvocationFailureOutputContract, "output_contract", "output", "output_unavailable")
	}
	result := service.complete(runContext, record, WorkflowRunStatusSucceeded, "", "", "", "none")
	if result.FailureCode == "" {
		result.Output = output
	}
	return result
}

func (service promptApplicationInvocationService) resolveAuthority(ctx PromptApplicationRuntimeContext) (promptApplicationInvocationAuthority, string) {
	assignment, _, err := service.runtimeRepository.Read(ctx)
	if errors.Is(err, errPromptApplicationRuntimeNotFound) {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureNotFound
	}
	if err != nil {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureStoreUnavailable
	}
	if assignment.State != "active" {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	resolved, failure := service.authorityResolver.Resolve(ctx, assignment.CandidateID, &assignment)
	if failure != "" {
		return promptApplicationInvocationAuthority{}, failure
	}
	catalogContext := ApplicationCatalogContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
	}
	catalog, err := service.catalogRepository.Read(catalogContext, ctx.ApplicationID)
	if err != nil || catalog.ApplicationKind != "prompt_application" || catalog.LifecycleState != applicationCatalogLifecycleActive {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	selection := service.selection(ctx.RequestContext, resolved.Candidate.Configuration.DefaultModel)
	selection = normalizePromptApplicationSelection(selection)
	if !promptApplicationSelectionEligible(resolved.Candidate.Configuration.DefaultProtocol, resolved.Candidate.Configuration.DefaultModel, selection) {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureAuthorityChanged
	}
	protocolDigest, err := promptApplicationProtocolPolicyDigest(resolved.Candidate.Configuration.DefaultProtocol, resolved.Candidate.Configuration.AllowedProtocols)
	if err != nil {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureStoreContract
	}
	selectionDigest, err := promptApplicationModelEligibilityDigest(selection)
	if err != nil {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureStoreContract
	}
	snapshot := PromptApplicationRuntimeAuthorityV2{
		SchemaVersion: promptApplicationRuntimeAuthorityV2Schema, ExecutionProfile: promptApplicationInvocationProfile,
		ApplicationID: catalog.ApplicationID, ApplicationRecordVersion: catalog.RecordVersion,
		ApplicationLifecycle: catalog.LifecycleState,
		PromptApplication: PromptApplicationAuthorityV2{
			AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.AssignmentVersion,
			AssignmentDigest: assignment.AssignmentDigest, PublishCandidateID: resolved.Candidate.CandidateID,
			PublishReviewVersion: resolved.Candidate.ReviewVersion, DraftID: resolved.Draft.DraftID,
			DraftVersion: resolved.Draft.DraftVersion, DraftDigest: resolved.Candidate.DraftDigest,
			PromptTemplateRef: assignment.PromptTemplateRef, DefaultProtocol: resolved.Candidate.Configuration.DefaultProtocol,
			DefaultModel: resolved.Candidate.Configuration.DefaultModel, ProtocolPolicyDigest: protocolDigest,
			ModelEligibilityDigest: selectionDigest,
		},
	}
	snapshot.AuthorityDigest, err = promptApplicationRuntimeAuthorityV2Digest(snapshot)
	if err != nil || validatePromptApplicationRuntimeAuthorityV2(snapshot) != nil {
		return promptApplicationInvocationAuthority{}, PromptApplicationRuntimeFailureStoreContract
	}
	return promptApplicationInvocationAuthority{
		Assignment: assignment, Resolved: resolved, Catalog: catalog, Selection: selection, Snapshot: snapshot,
	}, ""
}

func (service promptApplicationInvocationService) callGateway(
	ctx context.Context,
	runID string,
	authority promptApplicationInvocationAuthority,
	messages []PromptApplicationTemplateMessage,
) (string, string, string) {
	packet, err := json.Marshal(messages)
	if err != nil || len(packet) > promptApplicationTemplateMaximumRenderedBytes {
		return "", "protocol", PromptApplicationInvocationFailureOutcomeUnknown
	}
	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		requestID: runID, route: promptApplicationInvocationRoute, protocol: promptApplicationInvocationProtocol,
		locale: "zh-CN", promptText: string(packet),
		northboundFields: map[string]any{
			"request_kind": promptApplicationInvocationProtocol, "workflow_run_id": runID,
			"application_id": authority.Catalog.ApplicationID, "template_id": authority.Resolved.Template.TemplateID,
			"template_version": authority.Resolved.Template.TemplateVersion, "allow_tool_calls": false,
			"allow_retrieval": false, "writes_business_truth": false,
		},
	})
	if err != nil {
		return "", "protocol", PromptApplicationInvocationFailureOutcomeUnknown
	}
	envelope, err := service.bridge.HandleEnvelope(ctx, canonicalRequest, service.gatewayOptions(authority.Selection))
	if err != nil {
		if quotaFailure := gatewayRequestQuotaFailureCode(err); quotaFailure != "" {
			return "", "quota", quotaFailure
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return "", "canceled", PromptApplicationInvocationFailureCanceled
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", "timeout", PromptApplicationInvocationFailureOutcomeUnknown
		}
		return "", "unavailable", PromptApplicationInvocationFailureOutcomeUnknown
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Status), "ok") || envelope.Error != nil || envelope.Response == nil {
		return "", "provider_failed", PromptApplicationInvocationFailureOutcomeUnknown
	}
	if structured, ok := envelope.Response["structured_answer"]; ok {
		payload, marshalErr := json.Marshal(structured)
		if marshalErr != nil {
			return "", "output_unavailable", PromptApplicationInvocationFailureOutputContract
		}
		return string(payload), "none", ""
	}
	output := strings.TrimSpace(buildNorthboundResponseContent(envelope))
	if output == "" || len([]byte(output)) > promptApplicationOutputMaximumBytes {
		return "", "output_unavailable", PromptApplicationInvocationFailureOutputContract
	}
	return output, "none", ""
}

func (service promptApplicationInvocationService) complete(
	runContext WorkflowRunContext,
	record WorkflowRunRecord,
	status WorkflowRunStatus,
	failureCode, boundary, stage, gatewayCategory string,
) PromptApplicationInvocationResult {
	completedAt := service.currentTime()
	record.Status, record.CompletedAt = status, workflowRunTimestamp(completedAt)
	record.FailureCode = WorkflowRunFailureCode(failureCode)
	record.FailureSummary = ""
	if failureCode != "" {
		record.FailureSummary = promptApplicationInvocationFailureSummary(failureCode)
	}
	record.PromptDiagnostic = &PromptApplicationRunDiagnosticV6{
		FailureBoundary: boundary, FailureStage: stage, TerminalWriteState: "stored",
		GatewayFailureCategory: gatewayCategory, Summary: record.FailureSummary,
		RecommendedReviewAction: promptApplicationInvocationReviewAction(failureCode),
		ObservedAt:              workflowRunTimestamp(completedAt),
	}
	if err := service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Status = WorkflowRunStatusOutcomeUnknown
		record.FailureCode = WorkflowRunFailureCode(PromptApplicationInvocationFailureOutcomeUnknown)
		record.FailureSummary = promptApplicationInvocationFailureSummary(PromptApplicationInvocationFailureOutcomeUnknown)
		record.PromptDiagnostic.TerminalWriteState = "unknown"
		record.PromptDiagnostic.Summary = record.FailureSummary
		return PromptApplicationInvocationResult{
			Run: &record, FailureCode: PromptApplicationInvocationFailureOutcomeUnknown,
			FailureSummary: record.FailureSummary,
		}
	}
	return PromptApplicationInvocationResult{
		Run: &record, FailureCode: failureCode, FailureSummary: record.FailureSummary,
	}
}

func newPromptApplicationRunRecord(
	ctx PromptApplicationRuntimeContext,
	runID, inputDigest string,
	inputBytes int,
	names []string,
	namesDigest string,
	authority promptApplicationInvocationAuthority,
	startedAt time.Time,
) WorkflowRunRecord {
	selection := authority.Selection
	return WorkflowRunRecord{
		SchemaVersion: workflowRunRecordPromptSchemaVersion, RunID: runID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ExecutionSource: &workflowRunExecutionSource{
			Kind: promptApplicationExecutionKind, SourceKind: promptApplicationExecutionSourceKind,
			ID: authority.Resolved.Template.TemplateID, Version: authority.Resolved.Template.TemplateVersion,
		},
		ExecutionProfile: promptApplicationInvocationProfile, PromptApplication: &authority.Snapshot,
		InputDigest: inputDigest, InputBytes: inputBytes, VariableNames: append([]string(nil), names...),
		VariableNamesDigest: namesDigest, RequestedProtocol: authority.Resolved.Candidate.Configuration.DefaultProtocol,
		SelectedProtocol: authority.Resolved.Candidate.Configuration.DefaultProtocol,
		RequestedModel:   authority.Resolved.Candidate.Configuration.DefaultModel,
		SelectedProvider: selection.provider, SelectedProfile: selection.providerProfile,
		SelectedModel: selection.model, UpstreamModel: selection.upstreamModel, SelectionSource: selection.source,
		Status: WorkflowRunStatusRunning, StartedAt: workflowRunTimestamp(startedAt),
		PromptUsage: PromptApplicationRunUsageV6{State: "unavailable"},
		PromptDiagnostic: &PromptApplicationRunDiagnosticV6{
			TerminalWriteState: "pending", GatewayFailureCategory: "none", ObservedAt: workflowRunTimestamp(startedAt),
		},
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, ActorRef: ctx.ActorRef,
	}
}

func canonicalPromptApplicationInvocationInput(values map[string]any) ([]byte, []string, error) {
	names := make([]string, 0, len(values))
	for name := range values {
		if !promptApplicationVariableNamePattern.MatchString(name) {
			return nil, nil, errPromptApplicationVNextContract
		}
		names = append(names, name)
	}
	sort.Strings(names)
	payload, err := json.Marshal(values)
	if err != nil || len(payload) == 0 || len(payload) > promptApplicationTemplateMaximumSourceBytes {
		return nil, nil, errPromptApplicationVNextContract
	}
	return payload, names, nil
}

func promptApplicationInvocationRunID(ctx PromptApplicationRuntimeContext, clientKey string) string {
	sum := sha256.Sum256([]byte(ctx.TenantRef + "\x00" + ctx.WorkspaceID + "\x00" + ctx.ApplicationID + "\x00" + clientKey))
	return "run_" + hex.EncodeToString(sum[:])
}

func promptApplicationDuplicateInvocation(record WorkflowRunRecord, inputDigest string) PromptApplicationInvocationResult {
	if record.SchemaVersion != workflowRunRecordPromptSchemaVersion || record.InputDigest != inputDigest {
		return promptApplicationInvocationFailure(PromptApplicationInvocationFailureInputInvalid)
	}
	if record.Status == WorkflowRunStatusRunning {
		result := promptApplicationInvocationFailure(PromptApplicationInvocationFailureDuplicateRunning)
		result.Run, result.IdempotentReplay = &record, true
		return result
	}
	result := PromptApplicationInvocationResult{Run: &record, IdempotentReplay: true}
	if record.Status != WorkflowRunStatusSucceeded {
		result.FailureCode, result.FailureSummary = string(record.FailureCode), record.FailureSummary
	}
	return result
}

func promptApplicationProtocolPolicyDigest(defaultProtocol string, allowed []string) (string, error) {
	payload, err := json.Marshal(struct {
		Default string   `json:"default"`
		Allowed []string `json:"allowed"`
	}{Default: strings.TrimSpace(defaultProtocol), Allowed: append([]string(nil), allowed...)})
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func promptApplicationModelEligibilityDigest(selection northboundSelection) (string, error) {
	payload, err := json.Marshal([]string{
		selection.provider, selection.providerProfile, selection.model, selection.upstreamModel, selection.source,
	})
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func normalizePromptApplicationSelection(selection northboundSelection) northboundSelection {
	selection.provider = promptApplicationSelectionValue(selection.provider, "unavailable")
	selection.providerProfile = promptApplicationSelectionValue(selection.providerProfile, "default")
	selection.model = promptApplicationSelectionValue(selection.model, "unavailable")
	selection.upstreamModel = promptApplicationSelectionValue(selection.upstreamModel, selection.model)
	selection.source = promptApplicationSelectionValue(selection.source, "unavailable")
	return selection
}

func promptApplicationSelectionValue(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func promptApplicationSelectionEligible(protocol, requestedModel string, selection northboundSelection) bool {
	return isApplicationDraftProtocol(protocol) && validPromptApplicationSafeSelection(
		requestedModel, selection.provider, selection.providerProfile, selection.model, selection.upstreamModel, selection.source,
	)
}

func (service promptApplicationInvocationService) selection(ctx context.Context, model string) northboundSelection {
	if service.resolveSelection != nil {
		return service.resolveSelection(ctx, model)
	}
	return (&Server{bridge: service.bridge}).resolveNorthboundSelection(ctx, model, nil)
}

func (service promptApplicationInvocationService) gatewayOptions(selection northboundSelection) bridge.EnvelopeOptions {
	if service.envelopeOptions != nil {
		return service.envelopeOptions(selection, 0)
	}
	return (&Server{bridge: service.bridge}).buildBridgeEnvelopeOptions(selection, 0)
}

func (service promptApplicationInvocationService) currentTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func promptApplicationWorkflowRunContext(ctx PromptApplicationRuntimeContext) WorkflowRunContext {
	return WorkflowRunContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef,
	}
}

func promptApplicationInvocationFailure(code string) PromptApplicationInvocationResult {
	return PromptApplicationInvocationResult{FailureCode: code, FailureSummary: promptApplicationInvocationFailureSummary(code)}
}

func promptApplicationInvocationFailureSummary(code string) string {
	switch code {
	case PromptApplicationRuntimeFailureScopeDenied:
		return "Prompt application invocation scope is denied."
	case PromptApplicationRuntimeFailureNotFound:
		return "No active Prompt application runtime assignment was found."
	case PromptApplicationRuntimeFailureAuthorityChanged, PromptApplicationRuntimeFailureCandidate:
		return "Prompt application runtime authority is no longer eligible."
	case PromptApplicationInvocationFailureInputInvalid:
		return "Prompt application invocation input is invalid."
	case PromptApplicationInvocationFailureDuplicateRunning:
		return "Prompt application invocation is already running."
	case PromptApplicationInvocationFailureCanceled:
		return "Prompt application invocation was canceled."
	case PromptApplicationInvocationFailureOutputContract:
		return "Prompt application output did not satisfy its contract."
	case PromptApplicationInvocationFailureOutcomeUnknown:
		return "Prompt application invocation outcome is unknown and was not replayed."
	case GatewayRequestQuotaFailureExceeded:
		return "Prompt application request quota is exhausted for the current UTC period."
	case GatewayRequestQuotaFailurePolicyNotFound, GatewayRequestQuotaFailureStoreUnavailable:
		return "Prompt application request quota is unavailable."
	case GatewayRequestQuotaFailureAttemptConflict:
		return "Prompt application request quota admission conflicts with an existing attempt."
	default:
		return "Prompt application runtime store is unavailable."
	}
}

func promptApplicationInvocationReviewAction(code string) string {
	switch code {
	case "":
		return ""
	case PromptApplicationRuntimeFailureAuthorityChanged, PromptApplicationRuntimeFailureCandidate:
		return "review_authority"
	case PromptApplicationInvocationFailureOutputContract:
		return "review_output_contract"
	case PromptApplicationInvocationFailureCanceled:
		return "review_cancellation"
	case GatewayRequestQuotaFailureExceeded, GatewayRequestQuotaFailurePolicyNotFound,
		GatewayRequestQuotaFailureAttemptConflict, GatewayRequestQuotaFailureStoreUnavailable:
		return "review_run"
	default:
		return "review_run"
	}
}
