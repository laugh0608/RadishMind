package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	workflowRAGApplicationInvocationProtocol   = "workflow-rag-application-invocation-v1"
	workflowRAGApplicationInvocationRoute      = "/v1/application-rag/invocations"
	workflowRAGApplicationInvocationMaxBytes   = 4096
	workflowRAGApplicationInvocationMaxRuntime = 30 * time.Second
)

type WorkflowRAGApplicationInvocationInput struct {
	Input string
}

type WorkflowRAGApplicationInvocationResult struct {
	Run            *WorkflowRunRecord
	Answer         *WorkflowRAGApplicationAnswer
	FailureCode    string
	FailureSummary string
}

type workflowRAGApplicationInvocationService struct {
	runtimeRepository  workflowRAGApplicationRuntimeRepository
	resolver           workflowRAGApplicationAuthorityResolver
	runStore           workflowRunStore
	bridge             bridgeClient
	rank               func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult
	resolveSelection   func(context.Context, string) northboundSelection
	envelopeOptions    func(northboundSelection, float64) bridge.EnvelopeOptions
	defaultTemperature float64
	maxRuntime         time.Duration
	newRunID           func() (string, error)
	now                func() time.Time
}

func newWorkflowRAGApplicationInvocationService(repository workflowRAGApplicationRuntimeRepository, resolver workflowRAGApplicationAuthorityResolver, runStore workflowRunStore, bridgeClient bridgeClient) workflowRAGApplicationInvocationService {
	return workflowRAGApplicationInvocationService{runtimeRepository: repository, resolver: resolver, runStore: runStore, bridge: bridgeClient, rank: RankWorkflowRAGFragments, maxRuntime: workflowRAGApplicationInvocationMaxRuntime, newRunID: newWorkflowRunID, now: func() time.Time { return time.Now().UTC() }}
}

func (service workflowRAGApplicationInvocationService) Invoke(ctx WorkflowRAGApplicationRuntimeContext, input WorkflowRAGApplicationInvocationInput) WorkflowRAGApplicationInvocationResult {
	normalized := strings.TrimSpace(input.Input)
	if validateWorkflowRAGApplicationRuntimeContext(ctx) != nil {
		return workflowRAGApplicationInvocationFailure(WorkflowRAGApplicationFailureScopeDenied)
	}
	if normalized == "" || len([]byte(normalized)) > workflowRAGApplicationInvocationMaxBytes || !utf8Safe(normalized) {
		return workflowRAGApplicationInvocationFailure(WorkflowRAGApplicationFailurePayloadInvalid)
	}
	if workflowRAGContainsForbiddenMaterial(normalized) || applicationDraftStringContainsSecret(normalized) {
		return workflowRAGApplicationInvocationFailure(WorkflowRAGApplicationFailureSecretForbidden)
	}
	assignment, authority, failure := service.loadAuthority(ctx)
	if failure != "" {
		return workflowRAGApplicationInvocationFailure(failure)
	}
	runID, err := service.newRunID()
	if err != nil || !workflowRAGRunIDPattern.MatchString(runID) {
		return workflowRAGApplicationInvocationFailure(WorkflowRAGApplicationFailureStoreUnavailable)
	}
	requestContext := ctx.RequestContext
	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = workflowRAGApplicationInvocationMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(requestContext, maxRuntime)
	defer cancel()
	selection := service.selection(executionContext, authority.Candidate.Configuration.DefaultModel)
	runContext := workflowRAGApplicationRunContext(ctx)
	record := newWorkflowRAGApplicationRunRecord(ctx, normalized, assignment, authority, selection, runID, service.now())
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGApplicationInvocationFailure(WorkflowRAGApplicationFailureStoreUnavailable)
	}
	if failure = service.revalidate(ctx, assignment); failure != "" {
		return service.completeFailure(runContext, record, failure, WorkflowRunFailureBoundaryRetrievalPolicy, "scope")
	}

	retrievalStarted := service.now()
	ranking := service.rank(normalized, authority.Snapshot.Fragments, authority.Profile.DefaultTopK)
	retrievalDuration := service.now().Sub(retrievalStarted)
	record.SideEffects.RetrievalCalls = 1
	record.RetrievalAttempt.RetrievalLatencyMS = max(0, int(retrievalDuration.Milliseconds()))
	record.RetrievalAttempt.QueryDigest = ranking.QueryDigest
	record.RetrievalAttempt.QueryBytes = ranking.QueryBytes
	record.RetrievalAttempt.CandidateCount = ranking.CandidateCount
	if retrievalDuration > 2*time.Second || ranking.FailureCode == WorkflowRAGFailureBudgetExceeded {
		record.RetrievalAttempt.Status = "failed"
		return service.completeFailure(runContext, record, WorkflowRAGApplicationFailureBudgetExceeded, WorkflowRunFailureBoundaryRetrievalContext, "budget")
	}
	if ranking.FailureCode != "" {
		record.RetrievalAttempt.Status = "failed"
		failure = WorkflowRAGApplicationFailureStoreUnavailable
		category := "store"
		if ranking.FailureCode == WorkflowRAGFailureNoEvidence {
			failure, category = WorkflowRAGApplicationFailureNoEvidence, "no_evidence"
		} else if ranking.FailureCode == WorkflowRAGFailureQueryInvalid {
			failure, category = WorkflowRAGApplicationFailurePayloadInvalid, "query"
		}
		return service.completeFailure(runContext, record, failure, WorkflowRunFailureBoundaryRetrievalContext, category)
	}
	record.RetrievalAttempt.Status = "succeeded"
	record.RetrievalAttempt.ContextBytes = ranking.ContextBytes
	record.RetrievalAttempt.SelectedFragments = workflowRAGSelectedFragmentRecords(ranking.Selected)
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGApplicationInvocationCheckpointFailure(record)
	}
	if failure = service.revalidate(ctx, assignment); failure != "" {
		return service.completeFailure(runContext, record, failure, WorkflowRunFailureBoundaryProviderSelection, "scope")
	}
	record.SideEffects.ProviderCalls = 1
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGApplicationInvocationCheckpointFailure(record)
	}
	rawAnswer, gatewayCategory, failure := service.callGateway(executionContext, normalized, authority, runID, selection, ranking)
	if failure != "" {
		record.Diagnostic.GatewayFailureCategory = gatewayCategory
		return service.completeFailure(runContext, record, failure, WorkflowRunFailureBoundaryProviderCall, "provider")
	}
	answer, failure := parseWorkflowRAGApplicationAnswer(rawAnswer, ranking.Selected)
	if failure != "" {
		boundary, category := WorkflowRunFailureBoundaryRetrievalContext, "answer"
		if failure == WorkflowRAGApplicationFailureCitationInvalid {
			boundary, category = WorkflowRunFailureBoundaryRetrievalCitation, "citation"
		}
		return service.completeFailure(runContext, record, failure, boundary, category)
	}
	record.RetrievalAttempt.CitationRefs = workflowRAGApplicationCitationRefs(answer)
	record.Status = WorkflowRunStatusSucceeded
	record.CompletedAt = workflowRunTimestamp(service.now())
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	record.Diagnostic.ObservedAt = record.CompletedAt
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
		return workflowRAGApplicationInvocationCheckpointFailure(record)
	}
	return WorkflowRAGApplicationInvocationResult{Run: workflowRunRecordPointer(record), Answer: &answer}
}

func (service workflowRAGApplicationInvocationService) loadAuthority(ctx WorkflowRAGApplicationRuntimeContext) (WorkflowRAGApplicationRuntimeAssignment, workflowRAGApplicationAuthority, string) {
	assignment, _, _, err := service.runtimeRepository.Read(ctx)
	if errors.Is(err, errWorkflowRAGApplicationAssignmentNotFound) {
		return WorkflowRAGApplicationRuntimeAssignment{}, workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureAssignmentNotFound
	}
	if err != nil {
		return WorkflowRAGApplicationRuntimeAssignment{}, workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureStoreUnavailable
	}
	if assignment.State != workflowRAGApplicationRuntimeStateActive {
		return WorkflowRAGApplicationRuntimeAssignment{}, workflowRAGApplicationAuthority{}, WorkflowRAGApplicationFailureAssignmentRevoked
	}
	authority, failure := service.resolver.Resolve(ctx, assignment.PublishCandidateID, &assignment)
	return assignment, authority, failure
}

func (service workflowRAGApplicationInvocationService) revalidate(ctx WorkflowRAGApplicationRuntimeContext, expected WorkflowRAGApplicationRuntimeAssignment) string {
	current, _, _, err := service.runtimeRepository.Read(ctx)
	if err != nil {
		return WorkflowRAGApplicationFailureStoreUnavailable
	}
	if current.State != workflowRAGApplicationRuntimeStateActive {
		return WorkflowRAGApplicationFailureAssignmentRevoked
	}
	if current.AssignmentID != expected.AssignmentID || current.RecordVersion != expected.RecordVersion || current.AssignmentDigest != expected.AssignmentDigest {
		return WorkflowRAGApplicationFailureVersionConflict
	}
	_, failure := service.resolver.Resolve(ctx, expected.PublishCandidateID, &expected)
	return failure
}

func (service workflowRAGApplicationInvocationService) completeFailure(runContext WorkflowRunContext, record WorkflowRunRecord, failure string, boundary WorkflowRunFailureBoundary, category string) WorkflowRAGApplicationInvocationResult {
	record.Status = WorkflowRunStatusFailed
	record.FailureCode = WorkflowRunFailureCode(failure)
	record.FailureSummary = workflowRAGApplicationFailureSummary(failure)
	record.CompletedAt = workflowRunTimestamp(service.now())
	record.RAGAnswer = nil
	record.RetrievalAttempt.CitationRefs = []string{}
	record.Diagnostic.FailureBoundary = boundary
	record.Diagnostic.RetrievalFailureCategory = category
	record.Diagnostic.FailureStage = string(boundary)
	record.Diagnostic.Summary = record.FailureSummary
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	record.Diagnostic.ObservedAt = record.CompletedAt
	if err := service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
		return workflowRAGApplicationInvocationCheckpointFailure(record)
	}
	return WorkflowRAGApplicationInvocationResult{Run: workflowRunRecordPointer(record), FailureCode: failure, FailureSummary: record.FailureSummary}
}

func (service workflowRAGApplicationInvocationService) callGateway(ctx context.Context, input string, authority workflowRAGApplicationAuthority, runID string, selection northboundSelection, ranking WorkflowRAGRankingResult) (string, WorkflowRunGatewayFailureCategory, string) {
	promptNode := SavedWorkflowDraftNode{InputSummary: "请严格基于已批准应用知识证据回答用户问题。"}
	packet, err := buildWorkflowRAGGatewayPacket(input, promptNode, ranking.Selected)
	if err != nil {
		return "", WorkflowRunGatewayFailureProtocol, WorkflowRAGApplicationFailureGatewayFailed
	}
	packet = strings.Replace(packet, "workflow_rag_answer.v1", workflowRAGApplicationAnswerSchemaVersion, 1)
	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{requestID: runID + "-application-rag", route: workflowRAGApplicationInvocationRoute, protocol: workflowRAGApplicationInvocationProtocol, locale: "zh-CN", promptText: packet, northboundFields: map[string]any{"request_kind": workflowRAGApplicationInvocationProtocol, "workflow_run_id": runID, "application_id": authority.Candidate.ApplicationID, "application_configuration_draft_id": authority.Draft.DraftID, "allow_tool_calls": false, "allow_retrieval": false, "writes_business_truth": false}})
	if err != nil {
		return "", WorkflowRunGatewayFailureProtocol, WorkflowRAGApplicationFailureGatewayFailed
	}
	options := service.gatewayOptions(selection, service.defaultTemperature)
	envelope, err := service.bridge.HandleEnvelope(ctx, canonicalRequest, options)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", WorkflowRunGatewayFailureCanceled, WorkflowRAGApplicationFailureGatewayFailed
		}
		return "", workflowRunGatewayCategory(err), WorkflowRAGApplicationFailureGatewayFailed
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Status), "ok") || envelope.Error != nil || envelope.Response == nil {
		return "", WorkflowRunGatewayFailureProviderFailed, WorkflowRAGApplicationFailureGatewayFailed
	}
	if structured, ok := envelope.Response["structured_answer"]; ok {
		payload, marshalErr := json.Marshal(structured)
		if marshalErr != nil {
			return "", WorkflowRunGatewayFailureOutputUnavailable, WorkflowRAGApplicationFailureAnswerInvalid
		}
		return string(payload), WorkflowRunGatewayFailureNone, ""
	}
	output := strings.TrimSpace(buildNorthboundResponseContent(envelope))
	if output == "" || len([]byte(output)) > workflowExecutorMaxOutputBytes {
		return "", WorkflowRunGatewayFailureOutputUnavailable, WorkflowRAGApplicationFailureAnswerInvalid
	}
	return output, WorkflowRunGatewayFailureNone, ""
}

func (service workflowRAGApplicationInvocationService) selection(ctx context.Context, model string) northboundSelection {
	if service.resolveSelection != nil {
		return service.resolveSelection(ctx, model)
	}
	return (&Server{bridge: service.bridge}).resolveNorthboundSelection(ctx, model, nil)
}

func (service workflowRAGApplicationInvocationService) gatewayOptions(selection northboundSelection, temperature float64) bridge.EnvelopeOptions {
	if service.envelopeOptions != nil {
		return service.envelopeOptions(selection, temperature)
	}
	return (&Server{bridge: service.bridge}).buildBridgeEnvelopeOptions(selection, temperature)
}

func newWorkflowRAGApplicationRunRecord(ctx WorkflowRAGApplicationRuntimeContext, input string, assignment WorkflowRAGApplicationRuntimeAssignment, authority workflowRAGApplicationAuthority, selection northboundSelection, runID string, startedAt time.Time) WorkflowRunRecord {
	evidence := authority.Binding.Evidence
	return WorkflowRunRecord{SchemaVersion: workflowRunRecordAppRAGSchemaVersion, RunID: runID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ExecutionSource: &workflowRunExecutionSource{Kind: workflowRAGApplicationExecutionKind, SourceKind: workflowRAGApplicationExecutionSourceKind, ID: authority.Draft.DraftID, Version: authority.Draft.DraftVersion}, Status: WorkflowRunStatusRunning, StartedAt: workflowRunTimestamp(startedAt), InputBytes: len([]byte(input)), RequestedModel: authority.Candidate.Configuration.DefaultModel, SelectedProvider: strings.TrimSpace(selection.provider), SelectedProfile: strings.TrimSpace(selection.providerProfile), SelectedModel: strings.TrimSpace(selection.model), UpstreamModel: strings.TrimSpace(selection.upstreamModel), SelectionSource: strings.TrimSpace(selection.source), RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, ActorRef: ctx.ActorRef, RAGSnapshot: &workflowRAGRunSnapshotBinding{SnapshotID: authority.Snapshot.SnapshotID, SnapshotVersion: authority.Snapshot.SnapshotVersion, SnapshotDigest: authority.Snapshot.SnapshotDigest, RAGRef: authority.Snapshot.RAGRef}, RetrievalAttempt: &workflowRAGRunRetrievalAttempt{NodeID: "application_rag_retrieval", Status: "not_started", ProfileID: authority.Profile.ProfileID, ProfileVersion: authority.Profile.ProfileVersion, ProfileDigest: authority.Profile.ProfileDigest, QueryDigest: workflowRAGSHA256(input), QueryBytes: len([]byte(input)), SelectedFragments: []workflowRAGRunSelectedFragment{}, CitationRefs: []string{}}, RAGApplication: &workflowRAGApplicationRunAuthority{AssignmentID: assignment.AssignmentID, AssignmentVersion: assignment.RecordVersion, AssignmentDigest: assignment.AssignmentDigest, PublishCandidateID: authority.Candidate.CandidateID, PublishReviewVersion: authority.Candidate.ReviewVersion, PublishCandidateState: authority.Candidate.CandidateState, DraftID: authority.Draft.DraftID, DraftVersion: authority.Draft.DraftVersion, DraftDigest: authority.Candidate.DraftDigest, BindingRef: authority.Binding.WorkflowRAGApplicationBindingRef, Dataset: evidence.Dataset, CandidateReviewID: evidence.CandidateReviewID, BaselineSnapshot: evidence.BaselineSnapshot, CandidateSnapshot: evidence.CandidateSnapshot, EffectiveSnapshotRole: workflowRAGApplicationEffectiveSnapshotRole, Profile: evidence.Profile, ConfiguredProtocol: authority.Candidate.Configuration.DefaultProtocol, ConfiguredModel: authority.Candidate.Configuration.DefaultModel}, Diagnostic: newWorkflowRunDiagnostic()}
}

func parseWorkflowRAGApplicationAnswer(payload string, selected []WorkflowRAGRankedFragment) (WorkflowRAGApplicationAnswer, string) {
	var answer WorkflowRAGApplicationAnswer
	decoder := json.NewDecoder(bytes.NewReader([]byte(strings.TrimSpace(payload))))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&answer); err != nil || decoder.Decode(&struct{}{}) == nil || validateWorkflowRAGApplicationAnswer(answer, selected) != nil {
		return WorkflowRAGApplicationAnswer{}, WorkflowRAGApplicationFailureAnswerInvalid
	}
	var selectedRefs map[string]bool
	if selected != nil {
		selectedRefs = make(map[string]bool, len(selected))
		for _, fragment := range selected {
			selectedRefs[fragment.FragmentRef] = true
		}
	}
	for _, citation := range answer.Citations {
		if !selectedRefs[citation.FragmentRef] {
			return WorkflowRAGApplicationAnswer{}, WorkflowRAGApplicationFailureCitationInvalid
		}
	}
	return answer, ""
}

func validateWorkflowRAGApplicationAnswer(answer WorkflowRAGApplicationAnswer, selected []WorkflowRAGRankedFragment) error {
	converted := WorkflowRAGAnswer{SchemaVersion: "workflow_rag_answer.v1", Answer: answer.Answer, Citations: answer.Citations, Limitations: answer.Limitations, Confidence: answer.Confidence}
	var selectedRefs map[string]bool
	if selected != nil {
		selectedRefs = make(map[string]bool, len(selected))
		for _, fragment := range selected {
			selectedRefs[fragment.FragmentRef] = true
		}
	}
	if answer.SchemaVersion != workflowRAGApplicationAnswerSchemaVersion || validateWorkflowRAGAnswer(converted, selectedRefs) != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func workflowRAGApplicationCitationRefs(answer WorkflowRAGApplicationAnswer) []string {
	refs := make([]string, 0, len(answer.Citations))
	for _, citation := range answer.Citations {
		refs = append(refs, citation.FragmentRef)
	}
	return refs
}

func workflowRAGApplicationRunContext(ctx WorkflowRAGApplicationRuntimeContext) WorkflowRunContext {
	return WorkflowRunContext{RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef}
}

func workflowRAGApplicationInvocationFailure(code string) WorkflowRAGApplicationInvocationResult {
	return WorkflowRAGApplicationInvocationResult{FailureCode: code, FailureSummary: workflowRAGApplicationFailureSummary(code)}
}

func workflowRAGApplicationInvocationCheckpointFailure(record WorkflowRunRecord) WorkflowRAGApplicationInvocationResult {
	return WorkflowRAGApplicationInvocationResult{Run: workflowRunRecordPointer(record), FailureCode: WorkflowRAGApplicationFailureStoreUnavailable, FailureSummary: workflowRAGApplicationFailureSummary(WorkflowRAGApplicationFailureStoreUnavailable)}
}

func workflowRAGApplicationFailureSummary(code string) string {
	switch code {
	case WorkflowRAGApplicationFailureAssignmentNotFound:
		return "No active application RAG runtime assignment was found."
	case WorkflowRAGApplicationFailureAssignmentRevoked:
		return "The application RAG runtime assignment is revoked."
	case WorkflowRAGApplicationFailureVersionConflict:
		return "The application RAG runtime assignment changed during invocation."
	case WorkflowRAGApplicationFailureCandidateNotApproved, WorkflowRAGApplicationFailureCandidateSuperseded:
		return "The selected application publish candidate is no longer eligible."
	case WorkflowRAGApplicationFailureConfigurationChanged, WorkflowRAGApplicationFailureConfigurationInvalid:
		return "The selected application configuration is no longer eligible."
	case WorkflowRAGApplicationFailureBindingNotEligible:
		return "The selected application RAG binding is no longer eligible."
	case WorkflowRAGApplicationFailureApplicationArchived:
		return "The application is archived."
	case WorkflowRAGApplicationFailureScopeDenied:
		return "The application RAG invocation scope is denied."
	case WorkflowRAGApplicationFailurePayloadInvalid, WorkflowRAGApplicationFailureSecretForbidden:
		return "The application RAG invocation input is invalid."
	case WorkflowRAGApplicationFailureNoEvidence:
		return "The application RAG invocation found no eligible evidence."
	case WorkflowRAGApplicationFailureBudgetExceeded:
		return "The application RAG retrieval budget was exceeded."
	case WorkflowRAGApplicationFailureGatewayFailed:
		return "Gateway could not complete the application RAG invocation."
	case WorkflowRAGApplicationFailureAnswerInvalid, WorkflowRAGApplicationFailureCitationInvalid:
		return "The application RAG answer or citations are invalid."
	default:
		return "The application RAG runtime store is unavailable."
	}
}

func utf8Safe(value string) bool {
	return strings.ToValidUTF8(value, "") == value
}
