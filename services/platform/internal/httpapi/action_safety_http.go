package httpapi

import "strings"

const (
	actionSafetyReadProjectionSchema = "action_safety_read_projection.v1"
	actionSafetyReadStatusRecorded   = "recorded"
	actionSafetyReadStatusLegacy     = "not_recorded_legacy"
)

type ActionSafetyReadOwnerRefV1 struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version int    `json:"version"`
	Digest  string `json:"digest,omitempty"`
}

type ActionSafetyReadSourceRefV1 struct {
	Kind          string `json:"kind"`
	SchemaVersion string `json:"schema_version"`
	ID            string `json:"id"`
	Version       int    `json:"version"`
	Digest        string `json:"digest"`
}

type ActionSafetyReadScopeV1 struct {
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	Environment   string `json:"environment"`
	ApplicationID string `json:"application_id"`
}

type ActionSafetyReadPolicyRefV1 struct {
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type ActionSafetyReadDecisionV1 struct {
	DecisionID           string                       `json:"decision_id"`
	DecisionDigest       string                       `json:"decision_digest"`
	Scope                ActionSafetyReadScopeV1      `json:"scope"`
	Source               ActionSafetyReadSourceRefV1  `json:"source"`
	Project              string                       `json:"project"`
	Task                 string                       `json:"task"`
	ActionKind           string                       `json:"action_kind"`
	RiskLevel            string                       `json:"risk_level"`
	TargetKind           string                       `json:"target_kind"`
	Method               string                       `json:"method"`
	RequestedLevel       ActionSafetyLevel            `json:"requested_level"`
	MaximumAllowedLevel  ActionSafetyLevel            `json:"maximum_allowed_level"`
	EffectiveLevel       ActionSafetyLevel            `json:"effective_level"`
	RequiresConfirmation bool                         `json:"requires_confirmation"`
	ConfirmationState    string                       `json:"confirmation_state"`
	WritesBusinessTruth  bool                         `json:"writes_business_truth"`
	SideEffectBudget     ActionSafetySideEffectBudget `json:"side_effect_budget"`
	Blockers             []ActionSafetyFailureCode    `json:"blockers"`
	Policy               ActionSafetyReadPolicyRefV1  `json:"policy"`
	CreatedAt            string                       `json:"created_at"`
}

type ActionSafetyReadProjectionV1 struct {
	SchemaVersion       string                       `json:"schema_version"`
	Status              string                       `json:"status"`
	Owner               ActionSafetyReadOwnerRefV1   `json:"owner"`
	ProjectionVersion   string                       `json:"projection_version,omitempty"`
	ProjectionDigest    string                       `json:"projection_digest,omitempty"`
	Decisions           []ActionSafetyReadDecisionV1 `json:"decisions"`
	ObservedSideEffects *WorkflowRunSideEffects      `json:"observed_side_effects,omitempty"`
}

func actionSafetyReadFromResponse(value *ActionSafetyResponseProjectionV1) *ActionSafetyReadProjectionV1 {
	if value == nil || validateActionSafetyResponseProjection(*value) != nil {
		return nil
	}
	decisions := actionSafetyResponseDecisions(*value)
	return actionSafetyRecordedReadProjection(
		ActionSafetyReadOwnerRefV1{Kind: "agent_copilot_response", ID: value.SourceRunID, Version: 1, Digest: value.ResponseDigest},
		value.SchemaVersion, value.ProjectionDigest, decisions, nil,
	)
}

func actionSafetyReadFromAssignment(value *AgentCopilotRuntimeAssignmentV1) *ActionSafetyReadProjectionV1 {
	if value == nil {
		return nil
	}
	owner := ActionSafetyReadOwnerRefV1{
		Kind: "agent_copilot_runtime_assignment", ID: value.AssignmentID,
		Version: value.AssignmentVersion, Digest: value.AssignmentDigest,
	}
	if value.ActionSafety == nil {
		return actionSafetyLegacyReadProjection(owner)
	}
	if validateActionSafetyAssignmentProjection(*value.ActionSafety) != nil {
		return nil
	}
	return actionSafetyRecordedReadProjection(owner, value.ActionSafety.SchemaVersion,
		value.ActionSafety.ProjectionDigest, []ActionSafetyDecision{value.ActionSafety.Decision}, nil)
}

func actionSafetyReadFromPlan(value *WorkflowHTTPToolActionPlan) *ActionSafetyReadProjectionV1 {
	if value == nil {
		return nil
	}
	owner := ActionSafetyReadOwnerRefV1{
		Kind: "workflow_http_tool_action_plan", ID: value.PlanID,
		Version: value.RecordVersion, Digest: value.ToolPlanDigest,
	}
	if value.ActionSafety == nil {
		return actionSafetyLegacyReadProjection(owner)
	}
	if validateActionSafetyPlanProjection(*value.ActionSafety) != nil {
		return nil
	}
	return actionSafetyRecordedReadProjection(owner, value.ActionSafety.SchemaVersion,
		value.ActionSafety.ProjectionDigest, []ActionSafetyDecision{value.ActionSafety.Decision}, nil)
}

func actionSafetyReadFromRun(value *WorkflowRunRecord) *ActionSafetyReadProjectionV1 {
	if value == nil || !actionSafetyRunEligible(value.SchemaVersion) {
		return nil
	}
	owner := ActionSafetyReadOwnerRefV1{Kind: "workflow_run", ID: value.RunID, Version: value.RecordVersion}
	if value.ActionSafety == nil {
		return actionSafetyLegacyReadProjection(owner)
	}
	if validateActionSafetyRunProjection(*value.ActionSafety) != nil {
		return nil
	}
	owner.Digest = value.ActionSafety.ProjectionDigest
	observed := value.ActionSafety.SideEffects
	return actionSafetyRecordedReadProjection(owner, value.ActionSafety.SchemaVersion,
		value.ActionSafety.ProjectionDigest, value.ActionSafety.Decisions, &observed)
}

func actionSafetyRunEligible(schemaVersion string) bool {
	return schemaVersion == workflowRunRecordToolSchemaVersion || schemaVersion == workflowRunRecordDefinitionToolSchemaVersion ||
		schemaVersion == agentCopilotRunV7Schema
}

func actionSafetyRecordedReadProjection(
	owner ActionSafetyReadOwnerRefV1,
	projectionVersion string,
	projectionDigest string,
	decisions []ActionSafetyDecision,
	observed *WorkflowRunSideEffects,
) *ActionSafetyReadProjectionV1 {
	view := &ActionSafetyReadProjectionV1{
		SchemaVersion: actionSafetyReadProjectionSchema, Status: actionSafetyReadStatusRecorded,
		Owner: owner, ProjectionVersion: projectionVersion, ProjectionDigest: projectionDigest,
		Decisions: make([]ActionSafetyReadDecisionV1, 0, len(decisions)), ObservedSideEffects: observed,
	}
	for _, decision := range decisions {
		view.Decisions = append(view.Decisions, actionSafetyReadDecision(decision))
	}
	if validateActionSafetyReadProjection(view) != nil {
		return nil
	}
	return view
}

func actionSafetyLegacyReadProjection(owner ActionSafetyReadOwnerRefV1) *ActionSafetyReadProjectionV1 {
	view := &ActionSafetyReadProjectionV1{
		SchemaVersion: actionSafetyReadProjectionSchema, Status: actionSafetyReadStatusLegacy,
		Owner: owner, Decisions: []ActionSafetyReadDecisionV1{},
	}
	if validateActionSafetyReadProjection(view) != nil {
		return nil
	}
	return view
}

func actionSafetyReadDecision(value ActionSafetyDecision) ActionSafetyReadDecisionV1 {
	return ActionSafetyReadDecisionV1{
		DecisionID: value.DecisionID, DecisionDigest: value.DecisionDigest,
		Scope: ActionSafetyReadScopeV1{
			TenantRef: value.TenantRef, WorkspaceID: value.WorkspaceID,
			Environment: value.Environment, ApplicationID: value.ApplicationID,
		},
		Source: ActionSafetyReadSourceRefV1{
			Kind: value.SourceKind, SchemaVersion: value.SourceSchemaVersion,
			ID: value.SourceID, Version: value.SourceVersion, Digest: value.SourceDigest,
		},
		Project: value.Project, Task: value.Task, ActionKind: value.ActionKind,
		RiskLevel: value.RiskLevel, TargetKind: value.TargetKind, Method: value.Method,
		RequestedLevel: value.RequestedLevel, MaximumAllowedLevel: value.MaximumAllowedLevel,
		EffectiveLevel: value.EffectiveLevel, RequiresConfirmation: value.RequiresConfirmation,
		ConfirmationState: value.ConfirmationState, WritesBusinessTruth: value.WritesBusinessTruth,
		SideEffectBudget: value.SideEffectBudget,
		Blockers:         append([]ActionSafetyFailureCode{}, value.Blockers...),
		Policy:           ActionSafetyReadPolicyRefV1{Version: value.PolicyVersion, Digest: value.PolicyDigest},
		CreatedAt:        value.CreatedAt,
	}
}

func validateActionSafetyReadProjection(value *ActionSafetyReadProjectionV1) error {
	if value == nil || value.SchemaVersion != actionSafetyReadProjectionSchema ||
		strings.TrimSpace(value.Owner.Kind) == "" || strings.TrimSpace(value.Owner.ID) == "" || value.Owner.Version < 1 ||
		value.Decisions == nil {
		return errActionSafetyProjectionContract
	}
	if value.Status == actionSafetyReadStatusLegacy {
		if value.ProjectionVersion != "" || value.ProjectionDigest != "" || len(value.Decisions) != 0 || value.ObservedSideEffects != nil {
			return errActionSafetyProjectionContract
		}
		return nil
	}
	if value.Status != actionSafetyReadStatusRecorded || !actionSafetyVersionPattern.MatchString(value.ProjectionVersion) ||
		!workflowRAGDigestPattern.MatchString(value.ProjectionDigest) || len(value.Decisions) == 0 {
		return errActionSafetyProjectionContract
	}
	for _, decision := range value.Decisions {
		if !validActionSafetyReadDecision(decision) {
			return errActionSafetyProjectionContract
		}
	}
	if value.ObservedSideEffects != nil && value.ObservedSideEffects.BusinessWrites != 0 ||
		value.ObservedSideEffects != nil && value.ObservedSideEffects.ReplayWrites != 0 {
		return errActionSafetyProjectionContract
	}
	return nil
}

func validActionSafetyReadDecision(value ActionSafetyReadDecisionV1) bool {
	return validActionSafetyReference(value.DecisionID) && workflowRAGDigestPattern.MatchString(value.DecisionDigest) &&
		validActionSafetyScope(ActionSafetyScope{
			TenantRef: value.Scope.TenantRef, WorkspaceID: value.Scope.WorkspaceID,
			Environment: value.Scope.Environment, ApplicationID: value.Scope.ApplicationID,
		}) && validActionSafetySourceReference(ActionSafetySourceReference{
		Kind: value.Source.Kind, SchemaVersion: value.Source.SchemaVersion,
		ID: value.Source.ID, Version: value.Source.Version, Digest: value.Source.Digest,
	}) && validActionSafetyPolicyReference(ActionSafetyPolicyReference{Version: value.Policy.Version, Digest: value.Policy.Digest}) &&
		validActionSafetyLevel(value.RequestedLevel) && validActionSafetyLevel(value.MaximumAllowedLevel) &&
		validActionSafetyLevel(value.EffectiveLevel) && value.EffectiveLevel != ActionSafetyLevelWriteAllowedByPolicy &&
		canonicalActionSafetyBlockers(value.Blockers) && parsePromptApplicationTemplateTimestamp(value.CreatedAt) != nil
}
