package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const workflowRAGExecutionMaxRuntime = 30 * time.Second

const workflowRAGReconcilerActorRef = "system:workflow_rag_reconciler"

type workflowRAGExecutionService struct {
	draftReader        workflowSavedDraftReader
	snapshotRepository workflowRAGSnapshotRepository
	bridge             bridgeClient
	runStore           workflowRunStore
	maxRuntime         time.Duration
	defaultTemperature float64
	resolveSelection   func(context.Context, string) northboundSelection
	envelopeOptions    func(northboundSelection, float64) bridge.EnvelopeOptions
	rank               func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult
	profile            func() WorkflowRAGExecutionProfile
	newRunID           func() (string, error)
	now                func() time.Time
}

type WorkflowRAGReconciliationResult struct {
	Reconciled  int
	FailureCode string
}

func newWorkflowRAGExecutionService(
	draftReader workflowSavedDraftReader,
	snapshotRepository workflowRAGSnapshotRepository,
	bridgeClient bridgeClient,
	runStore workflowRunStore,
) workflowRAGExecutionService {
	return workflowRAGExecutionService{
		draftReader: draftReader, snapshotRepository: snapshotRepository, bridge: bridgeClient, runStore: runStore,
		maxRuntime: workflowRAGExecutionMaxRuntime, rank: RankWorkflowRAGFragments,
		profile: workflowRAGLexicalProfile, newRunID: newWorkflowRunID, now: func() time.Time { return time.Now().UTC() },
	}
}

func (service workflowRAGExecutionService) Execute(runContext WorkflowRunContext, request WorkflowRAGExecutionRequest) WorkflowRunResult {
	normalized, failure := normalizeWorkflowRAGExecutionRequest(request)
	if failure != "" {
		return workflowRAGExecutionFailure(failure)
	}
	draftResult := service.draftReader(workflowRAGSavedDraftContext(runContext), ReadWorkflowDraftRequest{DraftID: normalized.DraftID})
	if draftResult.FailureCode != "" || draftResult.Draft == nil {
		return workflowRAGExecutionFailure(workflowRAGFailureForDraftRead(draftResult.FailureCode))
	}
	draft := *draftResult.Draft
	plan, failure := buildWorkflowRAGExecutionPlan(draft, normalized.DraftVersion)
	if failure != "" {
		return workflowRAGExecutionFailure(failure)
	}
	snapshotContext := workflowRAGSnapshotContextFromRun(runContext, runContext.AuditRef)
	resource, snapshot, err := service.snapshotRepository.ReadByRAGRef(snapshotContext, plan.ragRef)
	if err != nil {
		return workflowRAGExecutionFailure(workflowRAGFailureForSnapshotRead(err))
	}
	if resource.LifecycleState != workflowRAGSnapshotActive || resource.ArchivedAt != nil {
		return workflowRAGExecutionFailure(WorkflowRAGFailureArchived)
	}
	if snapshot.RAGRef != plan.ragRef || snapshot.SnapshotVersion < 1 || validateStoredWorkflowRAGRecord(snapshot, snapshotContext) != nil {
		return workflowRAGExecutionFailure(WorkflowRAGFailureStoreUnavailable)
	}
	if snapshot.ProfileRef != workflowRAGProfileID {
		return workflowRAGExecutionFailure(WorkflowRAGFailureProfileDisabled)
	}
	profile := service.profile()
	if profile != workflowRAGLexicalProfile() || profile.ProfileID != snapshot.ProfileRef || profile.ProfileVersion != workflowRAGProfileVersion {
		return workflowRAGExecutionFailure(WorkflowRAGFailureProfileDisabled)
	}
	draftDigest, err := workflowRAGDraftDigest(draft)
	if err != nil {
		return workflowRAGExecutionFailure(WorkflowRAGFailureStoreUnavailable)
	}
	runID, err := service.newRunID()
	if err != nil || !workflowRAGRunIDPattern.MatchString(runID) {
		return workflowRAGExecutionFailure(WorkflowRAGFailureStoreUnavailable)
	}
	requestContext := runContext.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	maxRuntime := service.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = workflowRAGExecutionMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(requestContext, maxRuntime)
	defer cancel()
	selection := service.selection(executionContext, normalized.Model)
	record := newWorkflowRAGRunRecord(runContext, normalized, draft, draftDigest, plan, snapshot, profile, selection, runID, service.now())
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGExecutionFailure(WorkflowRAGFailureStoreUnavailable)
	}
	if err = service.appendAudit(snapshotContext, snapshot, runID, "retrieval_started", service.now()); err != nil {
		return service.completeFailure(runContext, snapshotContext, snapshot, record, WorkflowRAGFailureStoreUnavailable,
			WorkflowRunFailureBoundaryRetrievalStore, "store", plan.retrievalNode.NodeID)
	}

	retrievalStarted := service.now()
	ranking := service.rank(normalized.InputText, snapshot.Fragments, profile.DefaultTopK)
	retrievalDuration := service.now().Sub(retrievalStarted)
	record.SideEffects.RetrievalCalls = 1
	record.RetrievalAttempt.RetrievalLatencyMS = int(retrievalDuration.Milliseconds())
	if record.RetrievalAttempt.RetrievalLatencyMS < 0 {
		record.RetrievalAttempt.RetrievalLatencyMS = 0
	}
	if retrievalDuration > 2*time.Second {
		record.RetrievalAttempt.Status = "failed"
		return service.completeFailure(runContext, snapshotContext, snapshot, record, WorkflowRAGFailureBudgetExceeded,
			WorkflowRunFailureBoundaryRetrievalRank, "budget", plan.retrievalNode.NodeID)
	}
	record.RetrievalAttempt.QueryDigest = ranking.QueryDigest
	record.RetrievalAttempt.QueryBytes = ranking.QueryBytes
	record.RetrievalAttempt.CandidateCount = ranking.CandidateCount
	if ranking.FailureCode != "" {
		record.RetrievalAttempt.Status = "failed"
		boundary, category := WorkflowRunFailureBoundaryRetrievalRank, "store"
		if ranking.FailureCode == WorkflowRAGFailureNoEvidence {
			boundary, category = WorkflowRunFailureBoundaryRetrievalContext, "no_evidence"
		} else if ranking.FailureCode == WorkflowRAGFailureBudgetExceeded {
			boundary, category = WorkflowRunFailureBoundaryRetrievalContext, "budget"
		} else if ranking.FailureCode == WorkflowRAGFailureQueryInvalid {
			boundary, category = WorkflowRunFailureBoundaryRetrievalPolicy, "query"
		}
		return service.completeFailure(runContext, snapshotContext, snapshot, record, ranking.FailureCode, boundary, category, plan.retrievalNode.NodeID)
	}
	record.RetrievalAttempt.Status = "succeeded"
	record.RetrievalAttempt.ContextBytes = ranking.ContextBytes
	record.RetrievalAttempt.SelectedFragments = workflowRAGSelectedFragmentRecords(ranking.Selected)
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGCheckpointFailure(record)
	}

	record.SideEffects.ProviderCalls = 1
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		return workflowRAGCheckpointFailure(record)
	}
	rawAnswer, gatewayCategory, failure := service.callGateway(executionContext, normalized, draft, plan, runID, selection, ranking)
	if failure != "" {
		boundary := WorkflowRunFailureBoundaryProviderCall
		category := "provider"
		if failure == WorkflowRAGFailureCanceled {
			boundary, category = WorkflowRunFailureBoundaryProviderCall, "provider"
		}
		record.Diagnostic.GatewayFailureCategory = gatewayCategory
		return service.completeFailure(runContext, snapshotContext, snapshot, record, failure, boundary, category, plan.llmNode.NodeID)
	}
	answer, failure := parseWorkflowRAGAnswer(rawAnswer, ranking.Selected)
	if failure != "" {
		boundary, category := WorkflowRunFailureBoundaryRetrievalContext, "answer"
		if failure == WorkflowRAGFailureCitationInvalid {
			boundary, category = WorkflowRunFailureBoundaryRetrievalCitation, "citation"
		}
		return service.completeFailure(runContext, snapshotContext, snapshot, record, failure, boundary, category, plan.llmNode.NodeID)
	}
	record.RetrievalAttempt.CitationRefs = workflowRAGCitationRefs(answer)
	record.Status = WorkflowRunStatusSucceeded
	record.CompletedAt = workflowRunTimestamp(service.now())
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	record.Diagnostic.ObservedAt = record.CompletedAt
	if err = service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
		return workflowRAGCheckpointFailure(record)
	}
	if err = service.appendAudit(snapshotContext, snapshot, runID, "retrieval_succeeded", service.now()); err != nil {
		return WorkflowRunResult{Record: workflowRunRecordPointer(record), FailureCode: WorkflowRunFailureCode(WorkflowRAGFailureStoreUnavailable), FailureSummary: workflowRAGFailureSummary(WorkflowRAGFailureStoreUnavailable)}
	}
	return WorkflowRunResult{Record: workflowRunRecordPointer(record), RetrievalAnswer: &answer}
}

func (service workflowRAGExecutionService) completeFailure(
	runContext WorkflowRunContext,
	snapshotContext WorkflowRAGSnapshotContext,
	snapshot WorkflowRAGSnapshotRecord,
	record WorkflowRunRecord,
	failure string,
	boundary WorkflowRunFailureBoundary,
	category string,
	failedNodeID string,
) WorkflowRunResult {
	record.Status = WorkflowRunStatusFailed
	if failure == WorkflowRAGFailureCanceled {
		record.Status = WorkflowRunStatusCanceled
	}
	record.FailureCode = WorkflowRunFailureCode(failure)
	record.FailureSummary = workflowRAGFailureSummary(failure)
	record.CompletedAt = workflowRunTimestamp(service.now())
	record.RAGAnswer = nil
	record.RetrievalAttempt.CitationRefs = []string{}
	record.Diagnostic.FailureBoundary = boundary
	record.Diagnostic.RetrievalFailureCategory = category
	record.Diagnostic.FailedNodeID = strings.TrimSpace(failedNodeID)
	record.Diagnostic.FailureStage = string(boundary)
	record.Diagnostic.Summary = record.FailureSummary
	record.Diagnostic.RecommendedReviewAction = workflowRAGReviewAction(boundary)
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	record.Diagnostic.ObservedAt = record.CompletedAt
	if err := service.runStore.UpsertRun(runContext, &record); err != nil {
		record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWritePending
		return workflowRAGCheckpointFailure(record)
	}
	if err := service.appendAudit(snapshotContext, snapshot, record.RunID, "retrieval_failed", service.now()); err != nil {
		return WorkflowRunResult{Record: workflowRunRecordPointer(record), FailureCode: WorkflowRunFailureCode(WorkflowRAGFailureStoreUnavailable), FailureSummary: workflowRAGFailureSummary(WorkflowRAGFailureStoreUnavailable)}
	}
	return WorkflowRunResult{Record: workflowRunRecordPointer(record), FailureCode: record.FailureCode, FailureSummary: record.FailureSummary}
}

func (service workflowRAGExecutionService) ReconcileStale(runContext WorkflowRunContext) WorkflowRAGReconciliationResult {
	stale := true
	page, err := service.runStore.ListRuns(runContext, WorkflowRunListFilter{StaleRunning: &stale, Limit: workflowRunListMaxLimit})
	if err != nil {
		return WorkflowRAGReconciliationResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	result := WorkflowRAGReconciliationResult{}
	for _, candidate := range page.Records {
		if candidate.SchemaVersion != workflowRunRecordRAGSchemaVersion || candidate.Status != WorkflowRunStatusRunning || candidate.RAGSnapshot == nil {
			continue
		}
		snapshotContext := workflowRAGSnapshotContextFromRun(runContext, "audit_"+candidate.RunID+"_reconciled")
		snapshotContext.ActorRef = workflowRAGReconcilerActorRef
		_, snapshot, readErr := service.snapshotRepository.ReadByRAGRef(snapshotContext, candidate.RAGSnapshot.RAGRef)
		if readErr != nil || snapshot.SnapshotID != candidate.RAGSnapshot.SnapshotID || snapshot.SnapshotDigest != candidate.RAGSnapshot.SnapshotDigest {
			return WorkflowRAGReconciliationResult{Reconciled: result.Reconciled, FailureCode: WorkflowRAGFailureStoreUnavailable}
		}
		terminal := service.completeFailure(runContext, snapshotContext, snapshot, candidate, WorkflowRAGFailureInterrupted,
			WorkflowRunFailureBoundaryRunStore, "store", candidate.RetrievalAttempt.NodeID)
		if terminal.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureInterrupted) || terminal.Record == nil {
			return WorkflowRAGReconciliationResult{Reconciled: result.Reconciled, FailureCode: WorkflowRAGFailureStoreUnavailable}
		}
		result.Reconciled++
	}
	return result
}

func (service workflowRAGExecutionService) selection(ctx context.Context, model string) northboundSelection {
	if service.resolveSelection != nil {
		return service.resolveSelection(ctx, model)
	}
	server := &Server{bridge: service.bridge}
	return server.resolveNorthboundSelection(ctx, model, nil)
}

func (service workflowRAGExecutionService) appendAudit(ctx WorkflowRAGSnapshotContext, snapshot WorkflowRAGSnapshotRecord, runID, eventKind string, at time.Time) error {
	auditContext := ctx
	auditContext.AuditRef = "audit_" + runID + "_" + eventKind
	audit, err := buildWorkflowRAGSnapshotAudit(auditContext, eventKind, snapshot, workflowRunTimestamp(at))
	if err != nil {
		return err
	}
	return service.snapshotRepository.AppendAudit(auditContext, audit)
}

func workflowRAGSavedDraftContext(runContext WorkflowRunContext) SavedWorkflowDraftContext {
	return SavedWorkflowDraftContext{
		RequestContext: runContext.RequestContext, RequestID: runContext.RequestID, TenantRef: runContext.TenantRef,
		WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef,
		OwnerSubjectRef: runContext.ActorRef, ScopeGrants: cloneStringSlice(runContext.ScopeGrants), AuditRef: runContext.AuditRef,
	}
}

func workflowRAGSnapshotContextFromRun(runContext WorkflowRunContext, auditRef string) WorkflowRAGSnapshotContext {
	return WorkflowRAGSnapshotContext{
		RequestContext: runContext.RequestContext, RequestID: runContext.RequestID, TenantRef: runContext.TenantRef,
		WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, AuditRef: auditRef,
	}
}

func workflowRAGFailureForDraftRead(code SavedWorkflowDraftFailureCode) string {
	switch code {
	case SavedWorkflowDraftFailureScopeDenied, SavedWorkflowDraftFailureOwnerScopeDenied,
		SavedWorkflowDraftFailureApplicationScopeDenied, SavedWorkflowDraftFailureWorkspaceMembershipDenied:
		return WorkflowRAGFailureScopeDenied
	case SavedWorkflowDraftFailureStoreUnavailable, SavedWorkflowDraftFailureStoreContractMismatch:
		return WorkflowRAGFailureStoreUnavailable
	default:
		return WorkflowRAGFailureDraftIneligible
	}
}

func workflowRAGFailureForSnapshotRead(err error) string {
	switch {
	case errors.Is(err, errWorkflowRAGScopeDenied):
		return WorkflowRAGFailureScopeDenied
	case errors.Is(err, errWorkflowRAGNotFound):
		return WorkflowRAGFailureNotFound
	case errors.Is(err, errWorkflowRAGArchived):
		return WorkflowRAGFailureArchived
	default:
		return WorkflowRAGFailureStoreUnavailable
	}
}

func newWorkflowRAGRunRecord(
	ctx WorkflowRunContext,
	request WorkflowRAGExecutionRequest,
	draft SavedWorkflowDraft,
	draftDigest string,
	plan workflowRAGExecutionPlan,
	snapshot WorkflowRAGSnapshotRecord,
	profile WorkflowRAGExecutionProfile,
	selection northboundSelection,
	runID string,
	startedAt time.Time,
) WorkflowRunRecord {
	return WorkflowRunRecord{
		SchemaVersion: workflowRunRecordRAGSchemaVersion, RunID: runID, TenantRef: ctx.TenantRef,
		DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, DraftDigest: draftDigest,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, Status: WorkflowRunStatusRunning,
		StartedAt: workflowRunTimestamp(startedAt), InputBytes: len([]byte(request.InputText)),
		RequestedModel: request.Model, SelectedProvider: strings.TrimSpace(selection.provider),
		SelectedProfile: strings.TrimSpace(selection.providerProfile), SelectedModel: strings.TrimSpace(selection.model),
		UpstreamModel: strings.TrimSpace(selection.upstreamModel), SelectionSource: strings.TrimSpace(selection.source),
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, ActorRef: ctx.ActorRef,
		RAGSnapshot: &workflowRAGRunSnapshotBinding{
			SnapshotID: snapshot.SnapshotID, SnapshotVersion: snapshot.SnapshotVersion,
			SnapshotDigest: snapshot.SnapshotDigest, RAGRef: snapshot.RAGRef,
		},
		RetrievalAttempt: &workflowRAGRunRetrievalAttempt{
			NodeID: plan.retrievalNode.NodeID, Status: "not_started", ProfileID: profile.ProfileID,
			ProfileVersion: profile.ProfileVersion, ProfileDigest: profile.ProfileDigest,
			QueryDigest: workflowRAGSHA256(request.InputText), QueryBytes: len([]byte(request.InputText)),
			SelectedFragments: []workflowRAGRunSelectedFragment{}, CitationRefs: []string{},
		},
		Diagnostic: newWorkflowRunDiagnostic(),
	}
}

func workflowRAGSelectedFragmentRecords(selected []WorkflowRAGRankedFragment) []workflowRAGRunSelectedFragment {
	result := make([]workflowRAGRunSelectedFragment, 0, len(selected))
	for _, fragment := range selected {
		result = append(result, workflowRAGRunSelectedFragment{
			FragmentRef: fragment.FragmentRef, ContentDigest: fragment.ContentDigest, Rank: fragment.Rank,
			SourceType: fragment.SourceType, IsOfficial: fragment.IsOfficial, ExcerptTruncated: fragment.ExcerptTruncated,
		})
	}
	return result
}

func workflowRAGCitationRefs(answer WorkflowRAGAnswer) []string {
	result := make([]string, 0, len(answer.Citations))
	for _, citation := range answer.Citations {
		result = append(result, citation.FragmentRef)
	}
	return result
}

func workflowRAGExecutionFailure(code string) WorkflowRunResult {
	return WorkflowRunResult{FailureCode: WorkflowRunFailureCode(code), FailureSummary: workflowRAGFailureSummary(code)}
}

func workflowRAGCheckpointFailure(record WorkflowRunRecord) WorkflowRunResult {
	return WorkflowRunResult{
		Record: workflowRunRecordPointer(record), FailureCode: WorkflowRunFailureCode(WorkflowRAGFailureStoreUnavailable),
		FailureSummary: workflowRAGFailureSummary(WorkflowRAGFailureStoreUnavailable),
	}
}

func workflowRAGFailureSummary(code string) string {
	switch code {
	case WorkflowRAGFailureScopeDenied:
		return "Workflow retrieval scope is denied."
	case WorkflowRAGFailureNotFound:
		return "The exact workflow retrieval snapshot was not found."
	case WorkflowRAGFailureArchived:
		return "The workflow retrieval snapshot is archived."
	case WorkflowRAGFailureProfileDisabled:
		return "The exact workflow retrieval profile is unavailable."
	case WorkflowRAGFailureDraftIneligible:
		return "The exact workflow draft is not eligible for retrieval execution."
	case WorkflowRAGFailureQueryInvalid:
		return "The workflow retrieval query is invalid."
	case WorkflowRAGFailureBudgetExceeded:
		return "The workflow retrieval execution exceeded its budget."
	case WorkflowRAGFailureNoEvidence:
		return "The workflow retrieval execution found no eligible evidence."
	case WorkflowRAGFailureGatewayFailed:
		return "Gateway could not complete the workflow retrieval model call."
	case WorkflowRAGFailureAnswerInvalid:
		return "The workflow retrieval answer is invalid."
	case WorkflowRAGFailureCitationInvalid:
		return "The workflow retrieval citations are invalid."
	case WorkflowRAGFailureCanceled:
		return "The workflow retrieval execution was canceled or timed out."
	case WorkflowRAGFailureInterrupted:
		return "The workflow retrieval execution was interrupted and was not replayed."
	default:
		return "Workflow retrieval storage is unavailable."
	}
}

func workflowRAGReviewAction(boundary WorkflowRunFailureBoundary) WorkflowRunReviewAction {
	switch boundary {
	case WorkflowRunFailureBoundaryProviderCall, WorkflowRunFailureBoundaryProviderSelection:
		return WorkflowRunReviewProviderConfiguration
	case WorkflowRunFailureBoundaryRunStore, WorkflowRunFailureBoundaryRetrievalStore:
		return WorkflowRunReviewRunStore
	default:
		return WorkflowRunReviewDraft
	}
}
