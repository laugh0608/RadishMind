package httpapi

const (
	actionSafetySourceCopilotResponse        = "copilot_response"
	actionSafetySourceAgentCopilotAction     = "agent_copilot_proposed_action"
	actionSafetySourceWorkflowHTTPToolAction = "workflow_http_tool_action"
	actionSafetyCopilotResponseSchemaVersion = "copilot_response.v1"
	actionSafetyWorkflowHTTPToolTask         = "workflow_http_tool"
	actionSafetyActionNone                   = "none"
	actionSafetyTargetNone                   = "none"
	actionSafetyTargetHumanReviewOwner       = "human_review_owner"
	actionSafetyTargetWorkflowHTTPTool       = "workflow_http_tool"
	actionSafetyTargetBusinessTruth          = "business_truth"
	actionSafetyTargetShell                  = "shell"
	actionSafetyTargetCode                   = "code"
	actionSafetyTargetSandbox                = "sandbox"
	actionSafetyTargetAgentLoop              = "agent_loop"
	actionSafetyTargetConnectorMutation      = "connector_mutation"
	actionSafetyTargetAutomaticApply         = "automatic_apply"
)

type ActionSafetyScope struct {
	TenantRef     string
	WorkspaceID   string
	Environment   string
	ApplicationID string
}

type ActionSafetySourceReference struct {
	Kind          string
	SchemaVersion string
	ID            string
	Version       int
	Digest        string
}

type ActionSafetyPolicyReference struct {
	Version string
	Digest  string
}

type ActionSafetyCapabilityDemand struct {
	Project                  string
	Task                     string
	ActionKind               string
	RiskLevel                string
	TargetKind               string
	Method                   string
	RequiresConfirmation     bool
	ConfirmationState        string
	ProviderCalls            int
	HandoffRefs              int
	ToolNetworkCalls         int
	ConfirmationConsumptions int
	BusinessWrites           int
	ReplayWrites             int
}

// ActionSafetyAuthoritySnapshot contains only caller-projected, metadata-only
// facts read from existing owners. Review and assignment states are explicitly
// non-granting facts; they are retained here so tests can prove that they never
// elevate the compiler result by themselves.
type ActionSafetyAuthoritySnapshot struct {
	ScopeAuthorized           bool
	MembershipAllowed         bool
	SourceCurrent             bool
	PolicyCurrent             bool
	ActionAllowed             bool
	HandoffTargetAccepted     bool
	HandoffMetadataOnly       bool
	WorkflowDefinitionExact   bool
	ToolDefinitionExact       bool
	ExecutionProfileExact     bool
	PlanDigestExact           bool
	ExecutePermissionsAllowed bool
	DevelopmentGateEnabled    bool
	ConfirmationScopeExact    bool
	ConfirmationPlanExact     bool
	HumanApproved             bool
	CandidateApproved         bool
	AssignmentActive          bool
}

type ActionSafetyCompileInput struct {
	DecisionID      string
	Scope           ActionSafetyScope
	CurrentScope    ActionSafetyScope
	Source          ActionSafetySourceReference
	CurrentSource   ActionSafetySourceReference
	SourceAvailable bool
	Policy          ActionSafetyPolicyReference
	CurrentPolicy   ActionSafetyPolicyReference
	PolicyAvailable bool
	Demand          ActionSafetyCapabilityDemand
	Authority       ActionSafetyAuthoritySnapshot
	ActorRef        string
	RequestID       string
	AuditRef        string
	CreatedAt       string
}

// ActionSafetyPolicyCompiler is deterministic and side-effect free. It does not
// read a store, call a provider, create a handoff or plan, consume a confirmation,
// dispatch a tool, write a Run, or mutate business truth.
type ActionSafetyPolicyCompiler struct{}

func (ActionSafetyPolicyCompiler) Compile(input ActionSafetyCompileInput) (ActionSafetyDecision, ActionSafetyFailureCode) {
	if !validActionSafetyCompileInput(input) {
		return ActionSafetyDecision{}, ActionSafetyFailurePayloadInvalid
	}

	selected := map[ActionSafetyFailureCode]bool{}
	if !input.Authority.ScopeAuthorized || !input.Authority.MembershipAllowed || input.Scope != input.CurrentScope {
		selected[ActionSafetyFailureScopeDenied] = true
	}
	if !input.SourceAvailable {
		selected[ActionSafetyFailureSourceUnavailable] = true
	} else if !input.Authority.SourceCurrent || input.Source != input.CurrentSource {
		selected[ActionSafetyFailureSourceChanged] = true
	}
	if !input.PolicyAvailable {
		selected[ActionSafetyFailurePolicyUnavailable] = true
	} else if !input.Authority.PolicyCurrent || input.Policy != input.CurrentPolicy {
		selected[ActionSafetyFailurePolicyChanged] = true
	}

	requested, writeRequest := actionSafetyRequestedLevel(input.Demand)
	maximum := actionSafetyMaximumAllowedLevel(input, selected)
	if requested == ActionSafetyLevelToolCallable {
		if !actionSafetyToolAuthorityComplete(input) {
			selected[ActionSafetyFailureToolAuthorityUnavailable] = true
		}
		if input.Demand.ConfirmationState != "approved" {
			selected[ActionSafetyFailureConfirmationRequired] = true
		} else if !input.Authority.ConfirmationScopeExact || !input.Authority.ConfirmationPlanExact {
			selected[ActionSafetyFailureConfirmationChanged] = true
		}
	}
	effective, ok := actionSafetyTransition(requested, maximum, writeRequest)
	if !ok {
		return ActionSafetyDecision{}, ActionSafetyFailurePayloadInvalid
	}
	if writeRequest {
		selected[ActionSafetyFailureWriteBlocked] = true
	} else if effective != requested {
		selected[ActionSafetyFailureLevelEscalationDenied] = true
	}

	decision := ActionSafetyDecision{
		SchemaVersion: actionSafetyDecisionSchemaVersion, DecisionID: input.DecisionID,
		TenantRef: input.Scope.TenantRef, WorkspaceID: input.Scope.WorkspaceID, Environment: input.Scope.Environment, ApplicationID: input.Scope.ApplicationID,
		SourceKind: input.Source.Kind, SourceSchemaVersion: input.Source.SchemaVersion, SourceID: input.Source.ID, SourceVersion: input.Source.Version, SourceDigest: input.Source.Digest,
		Project: input.Demand.Project, Task: input.Demand.Task, ActionKind: input.Demand.ActionKind, RiskLevel: input.Demand.RiskLevel, TargetKind: input.Demand.TargetKind, Method: input.Demand.Method,
		RequestedLevel: requested, MaximumAllowedLevel: maximum, EffectiveLevel: effective,
		RequiresConfirmation: input.Demand.RequiresConfirmation, ConfirmationState: input.Demand.ConfirmationState,
		WritesBusinessTruth: input.Demand.BusinessWrites > 0, SideEffectBudget: actionSafetySideEffectBudget(effective),
		Blockers: normalizeActionSafetyBlockers(selected), PolicyVersion: input.Policy.Version, PolicyDigest: input.Policy.Digest,
		ActorRef: input.ActorRef, RequestID: input.RequestID, AuditRef: input.AuditRef, CreatedAt: input.CreatedAt,
	}
	digest, err := actionSafetyDecisionDigest(decision)
	if err != nil {
		return ActionSafetyDecision{}, ActionSafetyFailurePayloadInvalid
	}
	decision.DecisionDigest = digest
	if validateActionSafetyDecision(decision) != nil {
		return ActionSafetyDecision{}, ActionSafetyFailurePayloadInvalid
	}
	return decision, ""
}

func actionSafetyRequestedLevel(demand ActionSafetyCapabilityDemand) (ActionSafetyLevel, bool) {
	writeRequest := demand.BusinessWrites > 0 || demand.ReplayWrites > 0 || actionSafetyTargetIsWriteBlocked(demand.TargetKind) ||
		(demand.ToolNetworkCalls > 0 && demand.Method != "GET")
	if writeRequest {
		return ActionSafetyLevelWriteAllowedByPolicy, true
	}
	if demand.ToolNetworkCalls > 0 || demand.TargetKind == actionSafetyTargetWorkflowHTTPTool {
		return ActionSafetyLevelToolCallable, false
	}
	if demand.HandoffRefs > 0 || demand.TargetKind == actionSafetyTargetHumanReviewOwner {
		return ActionSafetyLevelHandoffReady, false
	}
	if demand.ActionKind != actionSafetyActionNone {
		return ActionSafetyLevelProposalOnly, false
	}
	return ActionSafetyLevelAnswerOnly, false
}

func actionSafetyMaximumAllowedLevel(input ActionSafetyCompileInput, blockers map[ActionSafetyFailureCode]bool) ActionSafetyLevel {
	if blockers[ActionSafetyFailureScopeDenied] || blockers[ActionSafetyFailureSourceUnavailable] || blockers[ActionSafetyFailureSourceChanged] ||
		blockers[ActionSafetyFailurePolicyUnavailable] || blockers[ActionSafetyFailurePolicyChanged] || !input.Authority.ActionAllowed {
		return ActionSafetyLevelAnswerOnly
	}
	maximum := ActionSafetyLevelAnswerOnly
	if input.Demand.ActionKind != actionSafetyActionNone {
		maximum = ActionSafetyLevelProposalOnly
	}
	if input.Demand.TargetKind == actionSafetyTargetHumanReviewOwner && input.Authority.HandoffTargetAccepted && input.Authority.HandoffMetadataOnly {
		maximum = ActionSafetyLevelHandoffReady
	}
	if input.Demand.TargetKind == actionSafetyTargetWorkflowHTTPTool && actionSafetyToolAuthorityComplete(input) &&
		input.Demand.ConfirmationState == "approved" && input.Authority.ConfirmationScopeExact && input.Authority.ConfirmationPlanExact {
		maximum = ActionSafetyLevelToolCallable
	}
	return maximum
}

func actionSafetyToolAuthorityComplete(input ActionSafetyCompileInput) bool {
	return input.Source.Kind == actionSafetySourceWorkflowHTTPToolAction && input.Demand.ActionKind == workflowHTTPToolID &&
		input.Demand.Method == "GET" && input.Demand.RequiresConfirmation && input.Demand.ToolNetworkCalls == 1 &&
		input.Demand.ConfirmationConsumptions == 1 && input.Demand.BusinessWrites == 0 && input.Demand.ReplayWrites == 0 &&
		input.Authority.WorkflowDefinitionExact && input.Authority.ToolDefinitionExact && input.Authority.ExecutionProfileExact &&
		input.Authority.PlanDigestExact && input.Authority.ExecutePermissionsAllowed && input.Authority.DevelopmentGateEnabled
}

func actionSafetyTransition(requested, maximum ActionSafetyLevel, writeRequest bool) (ActionSafetyLevel, bool) {
	if writeRequest {
		if requested != ActionSafetyLevelWriteAllowedByPolicy {
			return "", false
		}
		return ActionSafetyLevelWriteBlocked, true
	}
	switch requested {
	case ActionSafetyLevelAnswerOnly:
		switch maximum {
		case ActionSafetyLevelAnswerOnly, ActionSafetyLevelProposalOnly, ActionSafetyLevelHandoffReady, ActionSafetyLevelToolCallable:
			return ActionSafetyLevelAnswerOnly, true
		}
	case ActionSafetyLevelProposalOnly:
		switch maximum {
		case ActionSafetyLevelAnswerOnly:
			return ActionSafetyLevelAnswerOnly, true
		case ActionSafetyLevelProposalOnly, ActionSafetyLevelHandoffReady, ActionSafetyLevelToolCallable:
			return ActionSafetyLevelProposalOnly, true
		}
	case ActionSafetyLevelHandoffReady:
		switch maximum {
		case ActionSafetyLevelAnswerOnly:
			return ActionSafetyLevelAnswerOnly, true
		case ActionSafetyLevelProposalOnly:
			return ActionSafetyLevelProposalOnly, true
		case ActionSafetyLevelHandoffReady, ActionSafetyLevelToolCallable:
			return ActionSafetyLevelHandoffReady, true
		}
	case ActionSafetyLevelToolCallable:
		switch maximum {
		case ActionSafetyLevelAnswerOnly:
			return ActionSafetyLevelAnswerOnly, true
		case ActionSafetyLevelProposalOnly:
			return ActionSafetyLevelProposalOnly, true
		case ActionSafetyLevelHandoffReady:
			return ActionSafetyLevelHandoffReady, true
		case ActionSafetyLevelToolCallable:
			return ActionSafetyLevelToolCallable, true
		}
	}
	return "", false
}

func validActionSafetyCompileInput(input ActionSafetyCompileInput) bool {
	if !actionSafetyDecisionIDPattern.MatchString(input.DecisionID) || !validActionSafetyScope(input.Scope) || !validActionSafetyScope(input.CurrentScope) ||
		!validActionSafetySourceReference(input.Source) || input.SourceAvailable && !validActionSafetySourceReference(input.CurrentSource) ||
		!validActionSafetyPolicyReference(input.Policy) || input.PolicyAvailable && !validActionSafetyPolicyReference(input.CurrentPolicy) ||
		!validActionSafetySourceCompatibility(input.Source.Kind, input.Source.SchemaVersion, input.Demand.Project, input.Demand.Task, input.Demand.ActionKind, input.Demand.RiskLevel, input.Demand.TargetKind) ||
		!validActionSafetyConfirmation(input.Demand.RequiresConfirmation, input.Demand.ConfirmationState) || !validActionSafetyDemandShape(input.Demand) ||
		!validActionSafetyReference(input.ActorRef) || !validActionSafetyReference(input.RequestID) || !validActionSafetyReference(input.AuditRef) || !validActionSafetyTimestamp(input.CreatedAt) {
		return false
	}
	return true
}

func validActionSafetyScope(value ActionSafetyScope) bool {
	return validActionSafetyReference(value.TenantRef) && validActionSafetyReference(value.WorkspaceID) &&
		applicationCatalogIDPattern.MatchString(value.ApplicationID) && (value.Environment == "development" || value.Environment == "test")
}

func validActionSafetySourceReference(value ActionSafetySourceReference) bool {
	return validActionSafetySourceKind(value.Kind) && actionSafetyVersionPattern.MatchString(value.SchemaVersion) &&
		validActionSafetyReference(value.ID) && value.Version > 0 && workflowRAGDigestPattern.MatchString(value.Digest)
}

func validActionSafetyPolicyReference(value ActionSafetyPolicyReference) bool {
	return actionSafetyVersionPattern.MatchString(value.Version) && workflowRAGDigestPattern.MatchString(value.Digest)
}

func validActionSafetySourceKind(value string) bool {
	switch value {
	case actionSafetySourceCopilotResponse, actionSafetySourceAgentCopilotAction, actionSafetySourceWorkflowHTTPToolAction:
		return true
	default:
		return false
	}
}

func validActionSafetySourceCompatibility(sourceKind, sourceSchema, project, task, actionKind, riskLevel, targetKind string) bool {
	if !actionSafetyTokenPattern.MatchString(task) || !actionSafetyTokenPattern.MatchString(actionKind) ||
		!agentCopilotContainsString([]string{"low", "medium", "high"}, riskLevel) || !validActionSafetyTargetKind(targetKind) {
		return false
	}
	switch sourceKind {
	case actionSafetySourceCopilotResponse:
		return sourceSchema == actionSafetyCopilotResponseSchemaVersion && (project == "radishflow" || project == "radish") &&
			agentCopilotContainsString(agentCopilotCanonicalTasks(project), task) &&
			(actionKind == actionSafetyActionNone || agentCopilotContainsString(agentCopilotCanonicalActionKinds[:], actionKind)) &&
			(targetKind == actionSafetyTargetNone || targetKind == actionSafetyTargetHumanReviewOwner && actionKind != actionSafetyActionNone || actionSafetyTargetIsWriteBlocked(targetKind))
	case actionSafetySourceAgentCopilotAction:
		return sourceSchema == actionSafetyCopilotResponseSchemaVersion && (project == "radishflow" || project == "radish") &&
			agentCopilotContainsString(agentCopilotCanonicalTasks(project), task) && agentCopilotContainsString(agentCopilotCanonicalActionKinds[:], actionKind) &&
			(targetKind == actionSafetyTargetNone || targetKind == actionSafetyTargetHumanReviewOwner || actionSafetyTargetIsWriteBlocked(targetKind))
	case actionSafetySourceWorkflowHTTPToolAction:
		return sourceSchema == workflowHTTPToolPlanSchemaV2 && project == "radishmind" && task == actionSafetyWorkflowHTTPToolTask &&
			actionKind == workflowHTTPToolID && riskLevel == "medium" && (targetKind == actionSafetyTargetWorkflowHTTPTool || actionSafetyTargetIsWriteBlocked(targetKind))
	default:
		return false
	}
}

func validActionSafetyDemandShape(value ActionSafetyCapabilityDemand) bool {
	counts := []int{value.ProviderCalls, value.HandoffRefs, value.ToolNetworkCalls, value.ConfirmationConsumptions, value.BusinessWrites, value.ReplayWrites}
	for _, count := range counts {
		if count < 0 || count > 1 {
			return false
		}
	}
	if value.TargetKind == actionSafetyTargetNone && (value.HandoffRefs != 0 || value.ToolNetworkCalls != 0) ||
		value.TargetKind == actionSafetyTargetHumanReviewOwner && value.HandoffRefs != 1 ||
		value.TargetKind == actionSafetyTargetWorkflowHTTPTool && (value.ToolNetworkCalls != 1 || value.ConfirmationConsumptions != 1 || !agentCopilotContainsString([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, value.Method)) ||
		value.TargetKind != actionSafetyTargetWorkflowHTTPTool && (value.ToolNetworkCalls != 0 || value.ConfirmationConsumptions != 0 || value.Method != "none") ||
		(value.BusinessWrites > 0 || value.ReplayWrites > 0) && !actionSafetyTargetIsWriteBlocked(value.TargetKind) {
		return false
	}
	return true
}

func validActionSafetyTargetKind(value string) bool {
	switch value {
	case actionSafetyTargetNone, actionSafetyTargetHumanReviewOwner, actionSafetyTargetWorkflowHTTPTool,
		actionSafetyTargetBusinessTruth, actionSafetyTargetShell, actionSafetyTargetCode, actionSafetyTargetSandbox,
		actionSafetyTargetAgentLoop, actionSafetyTargetConnectorMutation, actionSafetyTargetAutomaticApply:
		return true
	default:
		return false
	}
}

func actionSafetyTargetIsWriteBlocked(value string) bool {
	switch value {
	case actionSafetyTargetBusinessTruth, actionSafetyTargetShell, actionSafetyTargetCode, actionSafetyTargetSandbox,
		actionSafetyTargetAgentLoop, actionSafetyTargetConnectorMutation, actionSafetyTargetAutomaticApply:
		return true
	default:
		return false
	}
}
