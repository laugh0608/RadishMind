package httpapi

import (
	"strings"
	"sync"
	"time"
)

const (
	workflowHTTPToolPlanV1TimeoutMS        = 5000
	workflowHTTPToolPlanV1MaxResponseBytes = 64 * 1024
	workflowHTTPToolPlanV1MaxOutputBytes   = 16 * 1024
)

type memoryWorkflowHTTPToolActionStore struct {
	ownerLock *sync.RWMutex
	plans     map[string]WorkflowHTTPToolActionPlan
	decisions []WorkflowHTTPToolConfirmationDecision
	audits    []WorkflowHTTPToolExecutionAudit
}

func newMemoryWorkflowHTTPToolActionStore(ownerLock *sync.RWMutex) *memoryWorkflowHTTPToolActionStore {
	if ownerLock == nil {
		ownerLock = &sync.RWMutex{}
	}
	return &memoryWorkflowHTTPToolActionStore{
		ownerLock: ownerLock,
		plans:     make(map[string]WorkflowHTTPToolActionPlan),
		decisions: make([]WorkflowHTTPToolConfirmationDecision, 0),
		audits:    make([]WorkflowHTTPToolExecutionAudit, 0),
	}
}

func (store *memoryWorkflowHTTPToolActionStore) CreatePlan(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.ownerLock == nil || validateWorkflowHTTPToolActionContext(ctx) != "" ||
		plan == nil || !workflowHTTPToolPlanMatchesContext(*plan, ctx) || plan.RecordVersion != 1 ||
		plan.Status != WorkflowHTTPToolActionStatusPending || !workflowHTTPToolCreateAuditMatchesContext(audit, *plan, ctx) {
		return errWorkflowHTTPToolActionContract
	}
	key := workflowHTTPToolActionStoreKey(ctx, plan.PlanID)
	store.ownerLock.Lock()
	defer store.ownerLock.Unlock()
	if _, exists := store.plans[key]; exists {
		return errWorkflowHTTPToolActionConflict
	}
	store.plans[key] = cloneWorkflowHTTPToolActionPlan(*plan)
	store.audits = append(store.audits, audit)
	return nil
}

func (store *memoryWorkflowHTTPToolActionStore) ReadPlan(
	ctx WorkflowHTTPToolActionContext,
	planID string,
) (WorkflowHTTPToolActionPlan, bool, error) {
	if store == nil || store.ownerLock == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || strings.TrimSpace(planID) == "" {
		return WorkflowHTTPToolActionPlan{}, false, errWorkflowHTTPToolActionContract
	}
	store.ownerLock.RLock()
	defer store.ownerLock.RUnlock()
	plan, found := store.plans[workflowHTTPToolActionStoreKey(ctx, strings.TrimSpace(planID))]
	return cloneWorkflowHTTPToolActionPlan(plan), found, nil
}

func (store *memoryWorkflowHTTPToolActionStore) DecidePlan(
	ctx WorkflowHTTPToolActionContext,
	plan *WorkflowHTTPToolActionPlan,
	decision WorkflowHTTPToolConfirmationDecision,
	audit WorkflowHTTPToolExecutionAudit,
) error {
	if store == nil || store.ownerLock == nil || validateWorkflowHTTPToolActionContext(ctx) != "" || plan == nil ||
		!workflowHTTPToolPlanMatchesContext(*plan, ctx) || !workflowHTTPToolDecisionMatchesPlan(decision, *plan) ||
		!workflowHTTPToolDecisionAuditMatchesContext(audit, decision, *plan, ctx) || plan.RecordVersion != decision.ResultingRecordVersion ||
		decision.ExpectedRecordVersion+1 != decision.ResultingRecordVersion {
		return errWorkflowHTTPToolActionContract
	}
	key := workflowHTTPToolActionStoreKey(ctx, plan.PlanID)
	store.ownerLock.Lock()
	defer store.ownerLock.Unlock()
	stored, found := store.plans[key]
	if !found || stored.RecordVersion != decision.ExpectedRecordVersion || stored.ToolPlanDigest != plan.ToolPlanDigest ||
		stored.AuditRef != plan.AuditRef ||
		!workflowHTTPToolDecisionAllowedFrom(stored.Status, decision.Outcome) {
		return errWorkflowHTTPToolActionConflict
	}
	store.plans[key] = cloneWorkflowHTTPToolActionPlan(*plan)
	store.decisions = append(store.decisions, decision)
	store.audits = append(store.audits, audit)
	return nil
}

func workflowHTTPToolActionStoreKey(ctx WorkflowHTTPToolActionContext, planID string) string {
	return strings.TrimSpace(ctx.TenantRef) + "\x00" + strings.TrimSpace(ctx.WorkspaceID) + "\x00" +
		strings.TrimSpace(ctx.ApplicationID) + "\x00" + strings.TrimSpace(planID)
}

func workflowHTTPToolPlanMatchesContext(plan WorkflowHTTPToolActionPlan, ctx WorkflowHTTPToolActionContext) bool {
	createdAt, createdErr := time.Parse(time.RFC3339Nano, plan.CreatedAt)
	expiresAt, expiresErr := time.Parse(time.RFC3339Nano, plan.ExpiresAt)
	computedDigest, digestErr := workflowHTTPToolPlanDigest(plan)
	argumentsValid := plan.PublicArguments.ResourceKey != "" &&
		len(plan.PublicArguments.ResourceKey) <= workflowHTTPToolMaxResourceKeySize &&
		!strings.Contains(plan.PublicArguments.ResourceKey, "://") &&
		workflowHTTPToolResourceKeyPattern.MatchString(plan.PublicArguments.ResourceKey) &&
		(plan.PublicArguments.Locale == "" || workflowHTTPToolLocalePattern.MatchString(plan.PublicArguments.Locale))
	decisionStateValid := plan.Status == WorkflowHTTPToolActionStatusPending && plan.LastDecisionByActorRef == nil && plan.LastDecisionAt == nil ||
		plan.Status != WorkflowHTTPToolActionStatusPending && plan.LastDecisionByActorRef != nil && plan.LastDecisionAt != nil
	lastDecisionValid := plan.LastDecisionByActorRef == nil || plan.LastDecisionAt == nil
	if plan.LastDecisionByActorRef != nil && plan.LastDecisionAt != nil {
		_, lastDecisionAtErr := time.Parse(time.RFC3339Nano, *plan.LastDecisionAt)
		lastDecisionValid = workflowHTTPToolReferencePattern.MatchString(*plan.LastDecisionByActorRef) && lastDecisionAtErr == nil
	}
	return workflowHTTPToolPlanIDPattern.MatchString(plan.PlanID) && plan.SchemaVersion == workflowHTTPToolPlanSchema && plan.RecordVersion > 0 &&
		plan.TenantRef == strings.TrimSpace(ctx.TenantRef) && workflowHTTPToolReferencePattern.MatchString(plan.TenantRef) &&
		plan.WorkspaceID == strings.TrimSpace(ctx.WorkspaceID) && workflowHTTPToolScopedIDPattern.MatchString(plan.WorkspaceID) &&
		plan.ApplicationID == strings.TrimSpace(ctx.ApplicationID) && workflowHTTPToolScopedIDPattern.MatchString(plan.ApplicationID) &&
		workflowHTTPToolScopedIDPattern.MatchString(plan.DraftID) && plan.DraftVersion > 0 &&
		workflowHTTPToolScopedIDPattern.MatchString(plan.NodeID) && workflowHTTPToolIDPattern.MatchString(plan.ToolID) &&
		plan.ToolVersion == workflowHTTPToolVersion &&
		workflowHTTPToolDigestPattern.MatchString(plan.DefinitionDigest) && workflowHTTPToolProfileIDPattern.MatchString(plan.ProfileID) &&
		plan.ProfileVersion > 0 && workflowHTTPToolDigestPattern.MatchString(plan.ProfileDigest) &&
		plan.Method == "GET" && workflowHTTPToolTargetPolicyKeyPattern.MatchString(plan.TargetPolicyKey) && argumentsValid &&
		workflowHTTPToolStringSlicesEqual(plan.OutputFields, workflowHTTPToolOutputFieldsV1()) &&
		workflowHTTPToolDigestPattern.MatchString(plan.OutputSchemaDigest) &&
		plan.CredentialPolicy == "none" && plan.TimeoutMS == workflowHTTPToolPlanV1TimeoutMS &&
		plan.MaxResponseBytes == workflowHTTPToolPlanV1MaxResponseBytes && plan.MaxOutputBytes == workflowHTTPToolPlanV1MaxOutputBytes &&
		workflowHTTPToolReferencePattern.MatchString(plan.PlannedByActorRef) && createdErr == nil && expiresErr == nil && expiresAt.After(createdAt) &&
		workflowHTTPToolReferencePattern.MatchString(plan.AuditRef) && digestErr == nil &&
		workflowHTTPToolDigestPattern.MatchString(plan.ToolPlanDigest) && plan.ToolPlanDigest == computedDigest && decisionStateValid && lastDecisionValid
}

func workflowHTTPToolStringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func workflowHTTPToolDecisionMatchesPlan(decision WorkflowHTTPToolConfirmationDecision, plan WorkflowHTTPToolActionPlan) bool {
	decidedAt, decidedAtErr := time.Parse(time.RFC3339Nano, decision.DecidedAt)
	createdAt, createdAtErr := time.Parse(time.RFC3339Nano, plan.CreatedAt)
	expiresAt, expiresAtErr := time.Parse(time.RFC3339Nano, plan.ExpiresAt)
	decisionTimeValid := decidedAtErr == nil && createdAtErr == nil && expiresAtErr == nil && !decidedAt.Before(createdAt)
	if decision.ActorSource == "human" || decision.Outcome == workflowHTTPToolConfirmationInvalidate {
		decisionTimeValid = decisionTimeValid && decidedAt.Before(expiresAt)
	} else if decision.Outcome == workflowHTTPToolConfirmationExpire {
		decisionTimeValid = decisionTimeValid && !decidedAt.Before(expiresAt)
	}
	return decision.SchemaVersion == workflowHTTPToolDecisionSchema && workflowHTTPToolConfirmationIDPattern.MatchString(decision.ConfirmationID) &&
		decision.PlanID == plan.PlanID && decision.TenantRef == plan.TenantRef && decision.WorkspaceID == plan.WorkspaceID &&
		decision.ApplicationID == plan.ApplicationID && decision.DraftID == plan.DraftID &&
		decision.DraftVersion == plan.DraftVersion && decision.NodeID == plan.NodeID && decision.ToolID == plan.ToolID &&
		decision.ToolVersion == plan.ToolVersion &&
		decision.ToolPlanDigest == plan.ToolPlanDigest && decision.ResultingRecordVersion == plan.RecordVersion &&
		decision.ExpectedRecordVersion+1 == decision.ResultingRecordVersion && plan.Status == workflowHTTPToolStatusForOutcome(decision.Outcome) &&
		decision.ReasonCode == workflowHTTPToolConfirmationReason(decision.Outcome) && decisionTimeValid &&
		workflowHTTPToolReferencePattern.MatchString(decision.DecidedByActorRef) && workflowHTTPToolReferencePattern.MatchString(decision.AuditRef) &&
		((decision.ActorSource == "human" && workflowHTTPToolHumanDecisionAllowed(decision.Outcome)) ||
			(decision.ActorSource == "system" && (decision.Outcome == workflowHTTPToolConfirmationExpire || decision.Outcome == workflowHTTPToolConfirmationInvalidate))) &&
		plan.LastDecisionByActorRef != nil && *plan.LastDecisionByActorRef == decision.DecidedByActorRef &&
		plan.LastDecisionAt != nil && *plan.LastDecisionAt == decision.DecidedAt
}

func workflowHTTPToolAuditMatchesPlan(audit WorkflowHTTPToolExecutionAudit, plan WorkflowHTTPToolActionPlan) bool {
	_, occurredAtErr := time.Parse(time.RFC3339Nano, audit.OccurredAt)
	return audit.SchemaVersion == workflowHTTPToolAuditSchema && workflowHTTPToolAuditIDPattern.MatchString(audit.EventID) &&
		audit.PlanID == plan.PlanID && audit.TenantRef == plan.TenantRef && audit.WorkspaceID == plan.WorkspaceID &&
		audit.ApplicationID == plan.ApplicationID && audit.DraftID == plan.DraftID && audit.DraftVersion == plan.DraftVersion &&
		audit.NodeID == plan.NodeID && audit.ToolID == plan.ToolID && audit.ToolVersion == plan.ToolVersion &&
		audit.DefinitionDigest == plan.DefinitionDigest &&
		audit.ProfileID == plan.ProfileID && audit.ProfileDigest == plan.ProfileDigest && audit.ToolPlanDigest == plan.ToolPlanDigest &&
		workflowHTTPToolAuditEventAllowed(audit.EventKind) && (audit.ActorSource == "human" || audit.ActorSource == "system") &&
		workflowHTTPToolReferencePattern.MatchString(audit.ActorRef) && workflowHTTPToolReferencePattern.MatchString(audit.RequestID) &&
		workflowHTTPToolReferencePattern.MatchString(audit.AuditRef) && occurredAtErr == nil &&
		audit.AttemptStatus == "not_claimed" && audit.ExecutionAttemptID == nil && audit.RunID == nil && audit.FailureCode == nil &&
		audit.FailureBoundary == nil && audit.HTTPStatusClass == nil && audit.ResponseBytes == 0 && audit.DurationMS == 0 &&
		audit.SideEffects == (WorkflowHTTPToolAuditSideEffects{})
}

func workflowHTTPToolCreateAuditMatchesContext(
	audit WorkflowHTTPToolExecutionAudit,
	plan WorkflowHTTPToolActionPlan,
	ctx WorkflowHTTPToolActionContext,
) bool {
	return workflowHTTPToolAuditMatchesPlan(audit, plan) && audit.EventKind == "confirmation_requested" &&
		audit.ConfirmationID == nil && audit.ActorSource == "human" && audit.ActorRef == plan.PlannedByActorRef &&
		audit.ActorRef == ctx.ActorRef && audit.RequestID == ctx.RequestID && audit.AuditRef == ctx.AuditRef &&
		audit.OccurredAt == plan.CreatedAt
}

func workflowHTTPToolDecisionAuditMatchesContext(
	audit WorkflowHTTPToolExecutionAudit,
	decision WorkflowHTTPToolConfirmationDecision,
	plan WorkflowHTTPToolActionPlan,
	ctx WorkflowHTTPToolActionContext,
) bool {
	actorMatchesContext := decision.ActorSource == "human" && decision.DecidedByActorRef == ctx.ActorRef ||
		decision.ActorSource == "system" && decision.DecidedByActorRef == "system:workflow_tool_action_policy"
	return workflowHTTPToolAuditMatchesPlan(audit, plan) && audit.ConfirmationID != nil &&
		*audit.ConfirmationID == decision.ConfirmationID && audit.EventKind == workflowHTTPToolAuditEventForOutcome(decision.Outcome) &&
		audit.ActorRef == decision.DecidedByActorRef && audit.ActorSource == decision.ActorSource &&
		audit.RequestID == ctx.RequestID && audit.AuditRef == decision.AuditRef && audit.AuditRef == ctx.AuditRef &&
		audit.OccurredAt == decision.DecidedAt && actorMatchesContext
}

func workflowHTTPToolAuditEventAllowed(event string) bool {
	switch event {
	case "confirmation_requested", "confirmation_recorded", "confirmation_rejected", "confirmation_deferred",
		"confirmation_canceled", "confirmation_expired", "confirmation_invalidated":
		return true
	default:
		return false
	}
}

var _ workflowHTTPToolActionStore = (*memoryWorkflowHTTPToolActionStore)(nil)
