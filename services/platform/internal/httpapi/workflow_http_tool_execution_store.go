package httpapi

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"
)

var (
	errWorkflowHTTPToolExecutionConflict    = errors.New("workflow HTTP tool execution conflict")
	errWorkflowHTTPToolExecutionContract    = errors.New("workflow HTTP tool execution store contract mismatch")
	errWorkflowHTTPToolExecutionUnavailable = errors.New("workflow HTTP tool execution store unavailable")
)

type workflowHTTPToolExecutionStore interface {
	ReadApprovedConfirmation(WorkflowHTTPToolActionContext, string, int) (WorkflowHTTPToolConfirmationDecision, bool, error)
	ReadClaimedExecution(WorkflowHTTPToolActionContext, string) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, bool, error)
	ClaimExecution(WorkflowHTTPToolActionContext, *WorkflowHTTPToolActionPlan, WorkflowHTTPToolConfirmationDecision, *WorkflowHTTPToolExecutionAttempt, *WorkflowRunRecord, WorkflowHTTPToolExecutionAudit) error
	CompleteExecution(WorkflowHTTPToolActionContext, *WorkflowHTTPToolExecutionAttempt, *WorkflowRunRecord, WorkflowHTTPToolExecutionAudit) error
}

type memoryWorkflowHTTPToolExecutionStore struct {
	actions  *memoryWorkflowHTTPToolActionStore
	runs     *memoryWorkflowRunStore
	attempts map[string]WorkflowHTTPToolExecutionAttempt
}

func newMemoryWorkflowHTTPToolExecutionStore(
	actions *memoryWorkflowHTTPToolActionStore,
	runs *memoryWorkflowRunStore,
) *memoryWorkflowHTTPToolExecutionStore {
	return &memoryWorkflowHTTPToolExecutionStore{
		actions:  actions,
		runs:     runs,
		attempts: make(map[string]WorkflowHTTPToolExecutionAttempt),
	}
}

func (store *memoryWorkflowHTTPToolExecutionStore) ReadApprovedConfirmation(
	ctx WorkflowHTTPToolActionContext,
	planID string,
	recordVersion int,
) (WorkflowHTTPToolConfirmationDecision, bool, error) {
	if !validMemoryWorkflowHTTPToolExecutionStore(store) || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		strings.TrimSpace(planID) == "" || recordVersion < 1 {
		return WorkflowHTTPToolConfirmationDecision{}, false, errWorkflowHTTPToolExecutionContract
	}
	store.actions.ownerLock.RLock()
	defer store.actions.ownerLock.RUnlock()
	for index := len(store.actions.decisions) - 1; index >= 0; index-- {
		decision := store.actions.decisions[index]
		if decision.PlanID == planID && decision.TenantRef == ctx.TenantRef &&
			decision.WorkspaceID == ctx.WorkspaceID && decision.ApplicationID == ctx.ApplicationID &&
			decision.Outcome == WorkflowHTTPToolConfirmationApprove && decision.ResultingRecordVersion == recordVersion {
			return decision, true, nil
		}
	}
	return WorkflowHTTPToolConfirmationDecision{}, false, nil
}

func (store *memoryWorkflowHTTPToolExecutionStore) ReadClaimedExecution(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) (WorkflowHTTPToolExecutionAttempt, WorkflowRunRecord, bool, error) {
	if !validMemoryWorkflowHTTPToolExecutionStore(store) || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		strings.TrimSpace(planID) == "" {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
	}
	planKey := workflowHTTPToolActionStoreKey(ctx, planID)
	store.actions.ownerLock.RLock()
	defer store.actions.ownerLock.RUnlock()
	attempt, found := store.attempts[planKey]
	if !found || attempt.Status != WorkflowHTTPToolAttemptClaimed {
		return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, nil
	}
	for key, run := range store.runs.records {
		if !strings.HasPrefix(key, workflowRunStoreKey(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, "")) ||
			run.PlanID != planID {
			continue
		}
		if run.Status != WorkflowRunStatusRunning || run.ToolAttempt == nil ||
			!reflect.DeepEqual(*run.ToolAttempt, attempt) {
			return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
		}
		return cloneWorkflowHTTPToolExecutionAttempt(attempt), cloneWorkflowRunRecord(run), true, nil
	}
	return WorkflowHTTPToolExecutionAttempt{}, WorkflowRunRecord{}, false, errWorkflowHTTPToolExecutionContract
}

func (store *memoryWorkflowHTTPToolExecutionStore) ClaimExecution(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if !validMemoryWorkflowHTTPToolExecutionStore(store) ||
		!validWorkflowHTTPToolExecutionClaim(ctx, plan, confirmation, attempt, run, audit) {
		return errWorkflowHTTPToolExecutionContract
	}
	planKey := workflowHTTPToolActionStoreKey(ctx, plan.PlanID)
	runKey := workflowRunStoreKey(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, run.RunID)
	store.actions.ownerLock.Lock()
	defer store.actions.ownerLock.Unlock()
	storedPlan, found := store.actions.plans[planKey]
	if !found || storedPlan.Status != WorkflowHTTPToolActionStatusApproved ||
		storedPlan.RecordVersion+1 != plan.RecordVersion || storedPlan.ToolPlanDigest != plan.ToolPlanDigest ||
		!memoryWorkflowHTTPToolApprovalExists(store.actions.decisions, confirmation, storedPlan) {
		return errWorkflowHTTPToolExecutionConflict
	}
	if _, found = store.attempts[planKey]; found {
		return errWorkflowHTTPToolExecutionConflict
	}
	if _, found = store.runs.records[runKey]; found {
		return errWorkflowHTTPToolExecutionConflict
	}
	claimedRun := cloneWorkflowRunRecord(*run)
	claimedRun.RecordVersion = 1
	store.actions.plans[planKey] = cloneWorkflowHTTPToolActionPlan(*plan)
	store.actions.audits = append(store.actions.audits, audit)
	store.attempts[planKey] = cloneWorkflowHTTPToolExecutionAttempt(*attempt)
	store.runs.records[runKey] = claimedRun
	store.runs.order = append(store.runs.order, runKey)
	for len(store.runs.order) > store.runs.capacity {
		oldestKey := store.runs.order[0]
		store.runs.order = store.runs.order[1:]
		delete(store.runs.records, oldestKey)
	}
	run.RecordVersion = 1
	return nil
}

func (store *memoryWorkflowHTTPToolExecutionStore) CompleteExecution(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if !validMemoryWorkflowHTTPToolExecutionStore(store) ||
		!validWorkflowHTTPToolExecutionCompletion(ctx, attempt, run, audit) {
		return errWorkflowHTTPToolExecutionContract
	}
	planKey := workflowHTTPToolActionStoreKey(ctx, run.PlanID)
	runKey := workflowRunStoreKey(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, run.RunID)
	store.actions.ownerLock.Lock()
	defer store.actions.ownerLock.Unlock()
	storedPlan, planFound := store.actions.plans[planKey]
	storedAttempt, attemptFound := store.attempts[planKey]
	storedRun, runFound := store.runs.records[runKey]
	if !planFound || storedPlan.Status != WorkflowHTTPToolActionStatusConsumed ||
		!attemptFound || storedAttempt.Status != WorkflowHTTPToolAttemptClaimed ||
		storedAttempt.AttemptID != attempt.AttemptID || !runFound || storedRun.Status != WorkflowRunStatusRunning ||
		storedRun.RecordVersion != run.RecordVersion || storedRun.ToolAttempt == nil ||
		storedRun.ToolAttempt.AttemptID != attempt.AttemptID {
		return errWorkflowHTTPToolExecutionConflict
	}
	completedRun := cloneWorkflowRunRecord(*run)
	completedRun.RecordVersion++
	store.attempts[planKey] = cloneWorkflowHTTPToolExecutionAttempt(*attempt)
	store.runs.records[runKey] = completedRun
	store.actions.audits = append(store.actions.audits, audit)
	run.RecordVersion = completedRun.RecordVersion
	return nil
}

func validMemoryWorkflowHTTPToolExecutionStore(store *memoryWorkflowHTTPToolExecutionStore) bool {
	return store != nil && store.actions != nil && store.runs != nil && store.actions.ownerLock != nil &&
		&store.runs.mu == store.actions.ownerLock && store.attempts != nil
}

func memoryWorkflowHTTPToolApprovalExists(
	decisions []WorkflowHTTPToolConfirmationDecision,
	confirmation WorkflowHTTPToolConfirmationDecision,
	plan WorkflowHTTPToolActionPlan,
) bool {
	for index := len(decisions) - 1; index >= 0; index-- {
		stored := decisions[index]
		if stored.ConfirmationID == confirmation.ConfirmationID &&
			stored.PlanID == plan.PlanID && stored.Outcome == WorkflowHTTPToolConfirmationApprove &&
			stored.ResultingRecordVersion == plan.RecordVersion && reflect.DeepEqual(stored, confirmation) {
			return true
		}
	}
	return false
}

func validWorkflowHTTPToolExecutionClaim(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) bool {
	if validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil || attempt == nil || run == nil ||
		ctx.RequestContext == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) ||
		plan.Status != WorkflowHTTPToolActionStatusConsumed ||
		confirmation.Outcome != WorkflowHTTPToolConfirmationApprove ||
		confirmation.ResultingRecordVersion+1 != plan.RecordVersion ||
		confirmation.ConfirmationID != attempt.ConfirmationID || confirmation.ConfirmationID != run.ConfirmationID ||
		confirmation.PlanID != plan.PlanID || run.PlanID != plan.PlanID ||
		attempt.Status != WorkflowHTTPToolAttemptClaimed || run.Status != WorkflowRunStatusRunning ||
		run.RecordVersion != 0 || run.ToolAttempt == nil || !reflect.DeepEqual(*run.ToolAttempt, *attempt) ||
		validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), run) != nil ||
		!workflowHTTPToolExecutionAuditMatches(ctx, *plan, confirmation, *attempt, *run, audit, true) {
		return false
	}
	return true
}

func validWorkflowHTTPToolExecutionCompletion(
	ctx WorkflowHTTPToolActionContext,
	attempt *WorkflowHTTPToolExecutionAttempt,
	run *WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) bool {
	if validateWorkflowHTTPToolActionContext(ctx) != "" || attempt == nil || run == nil ||
		ctx.RequestContext == nil || run.RecordVersion < 1 || run.ToolAttempt == nil ||
		!reflect.DeepEqual(*run.ToolAttempt, *attempt) || !isTerminalWorkflowRunStatus(run.Status) ||
		validateWorkflowRunStoreRecord(workflowRunContextFromToolAction(ctx), run) != nil {
		return false
	}
	return workflowHTTPToolCompletionAuditMatches(ctx, *attempt, *run, audit)
}

func workflowHTTPToolExecutionAuditMatches(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	confirmation WorkflowHTTPToolConfirmationDecision,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
	claim bool,
) bool {
	if !workflowHTTPToolExecutionAuditBaseMatches(ctx, plan, attempt, run, audit) ||
		audit.ConfirmationID == nil || *audit.ConfirmationID != confirmation.ConfirmationID ||
		audit.ActorSource != "human" {
		return false
	}
	if claim {
		return audit.EventKind == "tool_execution_started" && audit.AttemptStatus == "claimed" &&
			audit.FailureCode == nil && audit.FailureBoundary == nil && audit.HTTPStatusClass == nil &&
			audit.ResponseBytes == 0 && audit.DurationMS == 0 && audit.SideEffects == (WorkflowHTTPToolAuditSideEffects{})
	}
	return false
}

func workflowHTTPToolCompletionAuditMatches(
	ctx WorkflowHTTPToolActionContext,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) bool {
	plan := WorkflowHTTPToolActionPlan{
		PlanID: run.PlanID, TenantRef: run.TenantRef, WorkspaceID: run.WorkspaceID,
		ApplicationID: run.ApplicationID, DraftID: run.DraftID, DraftVersion: run.DraftVersion,
		NodeID: attempt.NodeID, ToolID: attempt.ToolID, DefinitionDigest: attempt.DefinitionDigest,
		ProfileID: attempt.ProfileID, ProfileDigest: attempt.ProfileDigest, ToolPlanDigest: attempt.ToolPlanDigest,
	}
	if !workflowHTTPToolExecutionAuditBaseMatches(ctx, plan, attempt, run, audit) ||
		audit.ConfirmationID == nil || *audit.ConfirmationID != attempt.ConfirmationID ||
		audit.SideEffects.ToolCalls != 1 || audit.SideEffects.NetworkAttempts < 0 ||
		audit.SideEffects.NetworkAttempts > 1 || audit.SideEffects.ProviderCalls != 0 ||
		audit.SideEffects.BusinessWrites != 0 || audit.SideEffects.ReplayWrites != 0 {
		return false
	}
	switch attempt.Status {
	case WorkflowHTTPToolAttemptSucceeded:
		return audit.EventKind == "tool_execution_succeeded" && audit.AttemptStatus == "succeeded" &&
			audit.FailureCode == nil && audit.ActorSource == "human"
	case WorkflowHTTPToolAttemptFailed:
		return audit.EventKind == "tool_execution_failed" && audit.AttemptStatus == "failed" &&
			audit.FailureCode != nil && *audit.FailureCode == string(attempt.FailureCode) && audit.ActorSource == "human"
	case WorkflowHTTPToolAttemptOutcomeUnknown:
		return audit.EventKind == "tool_execution_outcome_unknown" && audit.AttemptStatus == "outcome_unknown" &&
			audit.FailureCode != nil && *audit.FailureCode == string(WorkflowRunFailureToolOutcomeUnknown) &&
			(audit.ActorSource == "human" || audit.ActorSource == "system")
	default:
		return false
	}
}

func workflowHTTPToolExecutionAuditBaseMatches(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	attempt WorkflowHTTPToolExecutionAttempt,
	run WorkflowRunRecord,
	audit WorkflowHTTPToolExecutionAudit,
) bool {
	_, occurredAtErr := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	return audit.SchemaVersion == workflowHTTPToolAuditSchema && workflowHTTPToolAuditIDPattern.MatchString(audit.EventID) &&
		audit.TenantRef == ctx.TenantRef && audit.WorkspaceID == ctx.WorkspaceID && audit.ApplicationID == ctx.ApplicationID &&
		audit.PlanID == plan.PlanID && audit.DraftID == plan.DraftID && audit.DraftVersion == plan.DraftVersion &&
		audit.NodeID == attempt.NodeID && audit.ToolID == attempt.ToolID && audit.ToolVersion == workflowHTTPToolVersion &&
		audit.DefinitionDigest == attempt.DefinitionDigest && audit.ProfileID == attempt.ProfileID &&
		audit.ProfileDigest == attempt.ProfileDigest && audit.ToolPlanDigest == attempt.ToolPlanDigest &&
		audit.ExecutionAttemptID != nil && *audit.ExecutionAttemptID == attempt.AttemptID &&
		audit.RunID != nil && *audit.RunID == run.RunID && workflowHTTPToolExecutionAuditActorMatches(ctx, audit) &&
		audit.RequestID == ctx.RequestID && workflowHTTPToolReferencePattern.MatchString(audit.AuditRef) &&
		occurredAtErr == nil && audit.ResponseBytes == attempt.ResponseBytes && audit.DurationMS == attempt.DurationMS &&
		((audit.HTTPStatusClass == nil && attempt.HTTPStatusClass == "") ||
			(audit.HTTPStatusClass != nil && *audit.HTTPStatusClass == attempt.HTTPStatusClass))
}

func workflowHTTPToolExecutionAuditActorMatches(
	ctx WorkflowHTTPToolActionContext,
	audit WorkflowHTTPToolExecutionAudit,
) bool {
	if audit.ActorSource == "human" {
		return audit.ActorRef == ctx.ActorRef
	}
	return audit.ActorSource == "system" && audit.ActorRef == workflowHTTPToolReconcilerActorRef
}

func workflowRunContextFromToolAction(ctx WorkflowHTTPToolActionContext) WorkflowRunContext {
	return WorkflowRunContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		ScopeGrants: cloneStringSlice(ctx.ScopeGrants), AuditRef: ctx.AuditRef,
	}
}

func cloneWorkflowHTTPToolExecutionAttempt(attempt WorkflowHTTPToolExecutionAttempt) WorkflowHTTPToolExecutionAttempt {
	cloned := attempt
	cloned.OutputProjection = make(map[string]any, len(attempt.OutputProjection))
	for key, value := range attempt.OutputProjection {
		cloned.OutputProjection[key] = value
	}
	return cloned
}

func encodeWorkflowHTTPToolExecutionAttempt(attempt WorkflowHTTPToolExecutionAttempt) ([]byte, time.Time, *time.Time, error) {
	payload, err := json.Marshal(attempt)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowHTTPToolExecutionContract
	}
	claimedAt, err := time.Parse(time.RFC3339Nano, attempt.ClaimedAt)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowHTTPToolExecutionContract
	}
	var completedAt *time.Time
	if attempt.CompletedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, attempt.CompletedAt)
		if parseErr != nil {
			return nil, time.Time{}, nil, errWorkflowHTTPToolExecutionContract
		}
		parsed = parsed.UTC()
		completedAt = &parsed
	}
	return payload, claimedAt.UTC(), completedAt, nil
}

func decodeWorkflowHTTPToolConfirmationDecision(payload []byte) (WorkflowHTTPToolConfirmationDecision, error) {
	var decision WorkflowHTTPToolConfirmationDecision
	if err := json.Unmarshal(payload, &decision); err != nil ||
		decision.SchemaVersion != workflowHTTPToolDecisionSchema ||
		!workflowHTTPToolConfirmationIDPattern.MatchString(decision.ConfirmationID) ||
		!workflowHTTPToolPlanIDPattern.MatchString(decision.PlanID) ||
		decision.Outcome != WorkflowHTTPToolConfirmationApprove || decision.ActorSource != "human" ||
		decision.ExpectedRecordVersion < 1 || decision.ResultingRecordVersion != decision.ExpectedRecordVersion+1 ||
		!workflowHTTPToolDigestPattern.MatchString(decision.ToolPlanDigest) {
		return WorkflowHTTPToolConfirmationDecision{}, errWorkflowHTTPToolExecutionContract
	}
	return decision, nil
}

func decodeWorkflowHTTPToolExecutionAttempt(payload []byte) (WorkflowHTTPToolExecutionAttempt, error) {
	var attempt WorkflowHTTPToolExecutionAttempt
	if err := json.Unmarshal(payload, &attempt); err != nil || attempt.Status == "" ||
		!workflowHTTPToolAttemptIDPattern.MatchString(attempt.AttemptID) ||
		!workflowHTTPToolConfirmationIDPattern.MatchString(attempt.ConfirmationID) ||
		!workflowHTTPToolDigestPattern.MatchString(attempt.ToolPlanDigest) {
		return WorkflowHTTPToolExecutionAttempt{}, errWorkflowHTTPToolExecutionContract
	}
	return attempt, nil
}

var _ workflowHTTPToolExecutionStore = (*memoryWorkflowHTTPToolExecutionStore)(nil)
