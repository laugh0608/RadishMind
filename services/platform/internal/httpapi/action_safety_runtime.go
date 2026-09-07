package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	actionSafetyResponseProjectionSchema   = "action_safety_response_projection.v1"
	actionSafetyCandidateProjectionSchema  = "action_safety_candidate_projection.v1"
	actionSafetyAssignmentProjectionSchema = "action_safety_assignment_projection.v1"
	actionSafetyPlanProjectionSchema       = "action_safety_plan_projection.v1"
	actionSafetyRunProjectionSchema        = "action_safety_run_projection.v1"
	actionSafetyRuntimePolicyVersion       = "action_safety_runtime_policy.v1"

	actionSafetyCandidatePending  = "pending_review"
	actionSafetyCandidateApproved = "approved"
	actionSafetyCandidateRejected = "rejected"
)

var (
	errActionSafetyProjectionContract = errors.New("action safety projection contract mismatch")
	errActionSafetyOwnerConflict      = errors.New("action safety owner version conflict")
)

// ActionSafetyResponseProjectionV1 is an ephemeral projection owned by the
// canonical response path. It contains metadata-only candidate records and no
// answer, prompt, context, artifact, provider payload, or business data.
type ActionSafetyResponseProjectionV1 struct {
	SchemaVersion    string                              `json:"schema_version"`
	SourceRunID      string                              `json:"source_run_id"`
	ResponseDigest   string                              `json:"response_digest"`
	AnswerDecision   *ActionSafetyDecision               `json:"answer_decision,omitempty"`
	Candidates       []ActionSafetyCandidateProjectionV1 `json:"candidates"`
	ProjectionDigest string                              `json:"projection_digest"`
}

// ActionSafetyCandidateProjectionV1 belongs to the caller's candidate review
// owner. The runtime compiler never stores it independently.
type ActionSafetyCandidateProjectionV1 struct {
	SchemaVersion        string                      `json:"schema_version"`
	CandidateID          string                      `json:"candidate_id"`
	RecordVersion        int                         `json:"record_version"`
	SourceRunID          string                      `json:"source_run_id"`
	ActionIndex          int                         `json:"action_index"`
	Source               ActionSafetySourceReference `json:"source"`
	Project              string                      `json:"project"`
	Task                 string                      `json:"task"`
	ActionKind           string                      `json:"action_kind"`
	RiskLevel            string                      `json:"risk_level"`
	TargetKind           string                      `json:"target_kind"`
	RequiresConfirmation bool                        `json:"requires_confirmation"`
	ReviewState          string                      `json:"review_state"`
	Decision             ActionSafetyDecision        `json:"decision"`
	ProjectionDigest     string                      `json:"projection_digest"`
}

// ActionSafetyAssignmentProjectionV1 is embedded only in an existing runtime
// assignment owner. It is not an execution grant and cannot be submitted over
// the current HTTP contract.
type ActionSafetyAssignmentProjectionV1 struct {
	SchemaVersion     string               `json:"schema_version"`
	AssignmentVersion int                  `json:"assignment_version"`
	CandidateID       string               `json:"candidate_id"`
	CandidateVersion  int                  `json:"candidate_version"`
	Decision          ActionSafetyDecision `json:"decision"`
	ProjectionDigest  string               `json:"projection_digest"`
}

// ActionSafetyPlanProjectionV1 is an ephemeral snapshot carried by the
// existing memory action-plan owner. Batch C decides its durable representation.
type ActionSafetyPlanProjectionV1 struct {
	SchemaVersion    string               `json:"schema_version"`
	PlanID           string               `json:"plan_id"`
	PlanVersion      int                  `json:"plan_version"`
	ToolPlanDigest   string               `json:"tool_plan_digest"`
	Decision         ActionSafetyDecision `json:"decision"`
	ProjectionDigest string               `json:"projection_digest"`
}

// ActionSafetyRunProjectionV1 is an immutable Run projection of the decisions
// actually used and the side effects actually observed by the existing owner.
type ActionSafetyRunProjectionV1 struct {
	SchemaVersion    string                 `json:"schema_version"`
	Checkpoint       string                 `json:"checkpoint"`
	Decisions        []ActionSafetyDecision `json:"decisions"`
	SideEffects      WorkflowRunSideEffects `json:"side_effects"`
	ProjectionDigest string                 `json:"projection_digest"`
}

type actionSafetyRuntimeV1 struct {
	compiler    ActionSafetyPolicyCompiler
	environment string
	now         func() time.Time
	newID       func(string) (string, error)
	policyRef   func(string) (ActionSafetyPolicyReference, bool)
}

func newActionSafetyRuntimeV1(environment string) *actionSafetyRuntimeV1 {
	environment = strings.TrimSpace(environment)
	if environment != "development" && environment != "test" {
		environment = "development"
	}
	return &actionSafetyRuntimeV1{
		environment: environment,
		now:         func() time.Time { return time.Now().UTC() },
		newID:       newWorkflowRAGStableID,
		policyRef:   actionSafetyRuntimePolicyReference,
	}
}

func (runtime *actionSafetyRuntimeV1) NormalizeAgentCopilotResponse(
	ctx AgentCopilotRuntimeContext,
	runID string,
	response AgentCopilotResponse,
	authority agentCopilotInvocationAuthority,
) (ActionSafetyResponseProjectionV1, ActionSafetyFailureCode) {
	if runtime == nil || runtime.newID == nil || runtime.now == nil || runtime.policyRef == nil ||
		!workflowRAGRunIDPattern.MatchString(strings.TrimSpace(runID)) {
		return ActionSafetyResponseProjectionV1{}, ActionSafetyFailurePayloadInvalid
	}
	responsePayload, err := actionSafetySourceDigest(response)
	if err != nil {
		return ActionSafetyResponseProjectionV1{}, ActionSafetyFailurePayloadInvalid
	}
	policy, available := runtime.policyRef(authority.Resolved.Profile.PolicyDigest)
	projection := ActionSafetyResponseProjectionV1{
		SchemaVersion: actionSafetyResponseProjectionSchema,
		SourceRunID:   strings.TrimSpace(runID), ResponseDigest: responsePayload,
		Candidates: []ActionSafetyCandidateProjectionV1{},
	}
	if len(response.ProposedActions) == 0 {
		source := ActionSafetySourceReference{
			Kind: actionSafetySourceCopilotResponse, SchemaVersion: actionSafetyCopilotResponseSchemaVersion,
			ID: runID, Version: 1, Digest: responsePayload,
		}
		decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
			Scope: actionSafetyScopeFromAgentContext(ctx, runtime.environment), Source: source,
			CurrentScope: actionSafetyScopeFromAgentContext(ctx, runtime.environment), CurrentSource: source,
			SourceAvailable: true, Policy: policy, CurrentPolicy: policy, PolicyAvailable: available,
			Demand: ActionSafetyCapabilityDemand{
				Project: response.Project, Task: response.Task, ActionKind: actionSafetyActionNone,
				RiskLevel: response.RiskLevel, TargetKind: actionSafetyTargetNone, Method: "none",
				RequiresConfirmation: response.RequiresConfirmation,
				ConfirmationState:    actionSafetyConfirmationState(response.RequiresConfirmation, false), ProviderCalls: 1,
			},
			Authority: actionSafetyAgentAuthority(authority.Resolved.Profile, actionSafetyActionNone, true),
			ActorRef:  ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		})
		if failure != "" {
			return ActionSafetyResponseProjectionV1{}, failure
		}
		projection.AnswerDecision = &decision
	}
	for index, action := range response.ProposedActions {
		actionDigest, err := actionSafetySourceDigest(action)
		if err != nil {
			return ActionSafetyResponseProjectionV1{}, ActionSafetyFailurePayloadInvalid
		}
		targetKind, businessWrites := actionSafetyAgentActionTarget(action)
		source := ActionSafetySourceReference{
			Kind: actionSafetySourceAgentCopilotAction, SchemaVersion: actionSafetyCopilotResponseSchemaVersion,
			ID: fmt.Sprintf("%s.action.%d", runID, index+1), Version: 1, Digest: actionDigest,
		}
		decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
			Scope: actionSafetyScopeFromAgentContext(ctx, runtime.environment), Source: source,
			CurrentScope: actionSafetyScopeFromAgentContext(ctx, runtime.environment), CurrentSource: source,
			SourceAvailable: true, Policy: policy, CurrentPolicy: policy, PolicyAvailable: available,
			Demand: ActionSafetyCapabilityDemand{
				Project: response.Project, Task: response.Task, ActionKind: action.Kind, RiskLevel: action.RiskLevel,
				TargetKind: targetKind, Method: "none", RequiresConfirmation: action.RequiresConfirmation,
				ConfirmationState: actionSafetyConfirmationState(action.RequiresConfirmation, false), ProviderCalls: 1,
				BusinessWrites: businessWrites,
			},
			Authority: actionSafetyAgentAuthority(authority.Resolved.Profile, action.Kind, true),
			ActorRef:  ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		})
		if failure != "" {
			return ActionSafetyResponseProjectionV1{}, failure
		}
		candidateID, err := runtime.newID("asc_")
		if err != nil {
			return ActionSafetyResponseProjectionV1{}, ActionSafetyFailureSourceUnavailable
		}
		candidate := ActionSafetyCandidateProjectionV1{
			SchemaVersion: actionSafetyCandidateProjectionSchema, CandidateID: candidateID, RecordVersion: 1,
			SourceRunID: runID, ActionIndex: index, Source: source, Project: response.Project, Task: response.Task,
			ActionKind: action.Kind, RiskLevel: action.RiskLevel, TargetKind: targetKind,
			RequiresConfirmation: action.RequiresConfirmation, ReviewState: actionSafetyCandidatePending, Decision: decision,
		}
		if err := sealActionSafetyCandidateProjection(&candidate); err != nil {
			return ActionSafetyResponseProjectionV1{}, ActionSafetyFailurePayloadInvalid
		}
		projection.Candidates = append(projection.Candidates, candidate)
	}
	if err := sealActionSafetyResponseProjection(&projection); err != nil {
		return ActionSafetyResponseProjectionV1{}, ActionSafetyFailurePayloadInvalid
	}
	return projection, ""
}

func (runtime *actionSafetyRuntimeV1) ReviewCandidate(
	ctx AgentCopilotRuntimeContext,
	candidate ActionSafetyCandidateProjectionV1,
	expectedVersion int,
	currentSource ActionSafetySourceReference,
	profile AgentCopilotProfileVersionV1,
	approve bool,
	membershipAllowed bool,
	handoffTargetAccepted bool,
) (ActionSafetyCandidateProjectionV1, ActionSafetyFailureCode, error) {
	if validateActionSafetyCandidateProjection(candidate) != nil {
		return ActionSafetyCandidateProjectionV1{}, ActionSafetyFailureStoreContractMismatch, errActionSafetyProjectionContract
	}
	if candidate.RecordVersion != expectedVersion || candidate.ReviewState != actionSafetyCandidatePending {
		return ActionSafetyCandidateProjectionV1{}, "", errActionSafetyOwnerConflict
	}
	policy, available := runtime.policyRef(profile.PolicyDigest)
	targetKind, handoffRefs := actionSafetyTargetNone, 0
	if actionSafetyTargetIsWriteBlocked(candidate.TargetKind) {
		targetKind = candidate.TargetKind
	} else if approve {
		targetKind, handoffRefs = actionSafetyTargetHumanReviewOwner, 1
	}
	decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
		Previous: &candidate.Decision,
		Scope:    actionSafetyScopeFromAgentContext(ctx, runtime.environment), CurrentScope: actionSafetyScopeFromAgentContext(ctx, runtime.environment),
		Source: candidate.Source, CurrentSource: currentSource, SourceAvailable: true,
		Policy:        ActionSafetyPolicyReference{Version: candidate.Decision.PolicyVersion, Digest: candidate.Decision.PolicyDigest},
		CurrentPolicy: policy, PolicyAvailable: available,
		Demand: ActionSafetyCapabilityDemand{
			Project: candidate.Project, Task: candidate.Task, ActionKind: candidate.ActionKind, RiskLevel: candidate.RiskLevel,
			TargetKind: targetKind, Method: "none", RequiresConfirmation: candidate.RequiresConfirmation,
			ConfirmationState: actionSafetyConfirmationState(candidate.RequiresConfirmation, approve), HandoffRefs: handoffRefs,
			BusinessWrites: actionSafetyCandidateBusinessWrites(candidate),
		},
		Authority: actionSafetyReviewAuthority(profile, candidate.ActionKind, membershipAllowed, handoffTargetAccepted, approve),
		ActorRef:  ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if failure != "" {
		return ActionSafetyCandidateProjectionV1{}, failure, nil
	}
	if approve && decision.EffectiveLevel != ActionSafetyLevelHandoffReady {
		return ActionSafetyCandidateProjectionV1{}, actionSafetyPrimaryBlocker(decision), nil
	}
	candidate.RecordVersion++
	candidate.ReviewState = actionSafetyCandidateRejected
	if approve {
		candidate.ReviewState = actionSafetyCandidateApproved
	}
	candidate.Decision = decision
	if err := sealActionSafetyCandidateProjection(&candidate); err != nil {
		return ActionSafetyCandidateProjectionV1{}, ActionSafetyFailureStoreContractMismatch, errActionSafetyProjectionContract
	}
	return candidate, "", nil
}

func (runtime *actionSafetyRuntimeV1) ActivateCandidate(
	ctx AgentCopilotRuntimeContext,
	candidate ActionSafetyCandidateProjectionV1,
	profile AgentCopilotProfileVersionV1,
	assignmentVersion int,
	membershipAllowed bool,
) (ActionSafetyAssignmentProjectionV1, ActionSafetyFailureCode) {
	if validateActionSafetyCandidateProjection(candidate) != nil || candidate.ReviewState != actionSafetyCandidateApproved ||
		candidate.Decision.EffectiveLevel != ActionSafetyLevelHandoffReady || assignmentVersion < 1 {
		return ActionSafetyAssignmentProjectionV1{}, ActionSafetyFailureStoreContractMismatch
	}
	policy, available := runtime.policyRef(profile.PolicyDigest)
	decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
		Previous: &candidate.Decision,
		Scope:    actionSafetyScopeFromAgentContext(ctx, runtime.environment), CurrentScope: actionSafetyScopeFromAgentContext(ctx, runtime.environment),
		Source: candidate.Source, CurrentSource: candidate.Source, SourceAvailable: true,
		Policy:        ActionSafetyPolicyReference{Version: candidate.Decision.PolicyVersion, Digest: candidate.Decision.PolicyDigest},
		CurrentPolicy: policy, PolicyAvailable: available,
		Demand: ActionSafetyCapabilityDemand{
			Project: candidate.Project, Task: candidate.Task, ActionKind: candidate.ActionKind, RiskLevel: candidate.RiskLevel,
			TargetKind: actionSafetyTargetHumanReviewOwner, Method: "none", RequiresConfirmation: candidate.RequiresConfirmation,
			ConfirmationState: actionSafetyConfirmationState(candidate.RequiresConfirmation, true), HandoffRefs: 1,
		},
		Authority: actionSafetyReviewAuthority(profile, candidate.ActionKind, membershipAllowed, true, true),
		ActorRef:  ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if failure != "" {
		return ActionSafetyAssignmentProjectionV1{}, failure
	}
	if decision.EffectiveLevel != ActionSafetyLevelHandoffReady {
		return ActionSafetyAssignmentProjectionV1{}, actionSafetyPrimaryBlocker(decision)
	}
	projection := ActionSafetyAssignmentProjectionV1{
		SchemaVersion: actionSafetyAssignmentProjectionSchema, AssignmentVersion: assignmentVersion,
		CandidateID: candidate.CandidateID, CandidateVersion: candidate.RecordVersion, Decision: decision,
	}
	if err := sealActionSafetyAssignmentProjection(&projection); err != nil {
		return ActionSafetyAssignmentProjectionV1{}, ActionSafetyFailureStoreContractMismatch
	}
	return projection, ""
}

type actionSafetyRuntimeCompileRequest struct {
	Previous        *ActionSafetyDecision
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
}

func (runtime *actionSafetyRuntimeV1) compile(request actionSafetyRuntimeCompileRequest) (ActionSafetyDecision, ActionSafetyFailureCode) {
	if request.Previous != nil && !actionSafetyPreviousDecisionMatchesRequest(*request.Previous, request) {
		return ActionSafetyDecision{}, ActionSafetyFailureStoreContractMismatch
	}
	decisionID, err := runtime.newID("asd_")
	if err != nil {
		return ActionSafetyDecision{}, ActionSafetyFailureSourceUnavailable
	}
	return runtime.compiler.Compile(ActionSafetyCompileInput{
		DecisionID: decisionID, Scope: request.Scope, CurrentScope: request.CurrentScope,
		Source: request.Source, CurrentSource: request.CurrentSource, SourceAvailable: request.SourceAvailable,
		Policy: request.Policy, CurrentPolicy: request.CurrentPolicy, PolicyAvailable: request.PolicyAvailable,
		Demand: request.Demand, Authority: request.Authority,
		ActorRef: request.ActorRef, RequestID: request.RequestID, AuditRef: request.AuditRef,
		CreatedAt: runtime.now().UTC().Format(time.RFC3339Nano),
	})
}

func actionSafetyPreviousDecisionMatchesRequest(previous ActionSafetyDecision, request actionSafetyRuntimeCompileRequest) bool {
	return validateActionSafetyDecision(previous) == nil &&
		previous.TenantRef == request.Scope.TenantRef && previous.WorkspaceID == request.Scope.WorkspaceID &&
		previous.Environment == request.Scope.Environment && previous.ApplicationID == request.Scope.ApplicationID &&
		previous.SourceKind == request.Source.Kind && previous.SourceSchemaVersion == request.Source.SchemaVersion &&
		previous.SourceID == request.Source.ID && previous.SourceVersion == request.Source.Version &&
		previous.SourceDigest == request.Source.Digest && previous.PolicyVersion == request.Policy.Version &&
		previous.PolicyDigest == request.Policy.Digest
}

func actionSafetyRuntimePolicyReference(ownerPolicyDigest string) (ActionSafetyPolicyReference, bool) {
	if !workflowRAGDigestPattern.MatchString(strings.TrimSpace(ownerPolicyDigest)) {
		return ActionSafetyPolicyReference{}, false
	}
	digest, err := canonicalSHA256(struct {
		SchemaVersion     string                    `json:"schema_version"`
		DecisionSchema    string                    `json:"decision_schema"`
		OwnerPolicyDigest string                    `json:"owner_policy_digest"`
		Levels            []ActionSafetyLevel       `json:"levels"`
		Blockers          []ActionSafetyFailureCode `json:"blockers"`
	}{
		SchemaVersion: actionSafetyRuntimePolicyVersion, DecisionSchema: actionSafetyDecisionSchemaVersion,
		OwnerPolicyDigest: ownerPolicyDigest,
		Levels: []ActionSafetyLevel{
			ActionSafetyLevelAnswerOnly, ActionSafetyLevelProposalOnly, ActionSafetyLevelHandoffReady,
			ActionSafetyLevelToolCallable, ActionSafetyLevelWriteBlocked, ActionSafetyLevelWriteAllowedByPolicy,
		},
		Blockers: append([]ActionSafetyFailureCode(nil), actionSafetyBlockerOrder...),
	})
	if err != nil {
		return ActionSafetyPolicyReference{}, false
	}
	return ActionSafetyPolicyReference{Version: actionSafetyRuntimePolicyVersion, Digest: digest}, true
}

func actionSafetySourceDigest(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func actionSafetyScopeFromAgentContext(ctx AgentCopilotRuntimeContext, environment string) ActionSafetyScope {
	return ActionSafetyScope{TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: environment, ApplicationID: ctx.ApplicationID}
}

func actionSafetyAgentAuthority(profile AgentCopilotProfileVersionV1, actionKind string, membershipAllowed bool) ActionSafetyAuthoritySnapshot {
	allowed := actionKind == actionSafetyActionNone || agentCopilotContainsString(profile.ResponsePolicy.AllowedActionKinds, actionKind)
	return ActionSafetyAuthoritySnapshot{
		ScopeAuthorized: true, MembershipAllowed: membershipAllowed, SourceCurrent: true, PolicyCurrent: true,
		ActionAllowed: allowed, HandoffMetadataOnly: true,
	}
}

func actionSafetyReviewAuthority(profile AgentCopilotProfileVersionV1, actionKind string, membershipAllowed, targetAccepted, approved bool) ActionSafetyAuthoritySnapshot {
	authority := actionSafetyAgentAuthority(profile, actionKind, membershipAllowed)
	authority.HandoffTargetAccepted = targetAccepted
	authority.HandoffMetadataOnly = true
	authority.HumanApproved = approved
	authority.CandidateApproved = approved
	return authority
}

func actionSafetyConfirmationState(required, approved bool) string {
	if !required {
		return "not_required"
	}
	if approved {
		return "approved"
	}
	return "pending"
}

func actionSafetyAgentActionTarget(action AgentCopilotResponseAction) (string, int) {
	if value, ok := action.Apply["automatic"].(bool); ok && value {
		return actionSafetyTargetAutomaticApply, 1
	}
	if value, ok := action.Apply["execute"].(bool); ok && value {
		return actionSafetyTargetAutomaticApply, 1
	}
	if value, ok := action.Apply["mode"].(string); ok && strings.TrimSpace(value) == actionSafetyTargetAutomaticApply {
		return actionSafetyTargetAutomaticApply, 1
	}
	return actionSafetyTargetNone, 0
}

func actionSafetyCandidateBusinessWrites(candidate ActionSafetyCandidateProjectionV1) int {
	if actionSafetyTargetIsWriteBlocked(candidate.TargetKind) {
		return 1
	}
	return 0
}

func actionSafetyPrimaryBlocker(decision ActionSafetyDecision) ActionSafetyFailureCode {
	for _, blocker := range []ActionSafetyFailureCode{
		ActionSafetyFailureScopeDenied,
		ActionSafetyFailurePayloadInvalid,
		ActionSafetyFailureSourceUnavailable,
		ActionSafetyFailureSourceChanged,
		ActionSafetyFailurePolicyUnavailable,
		ActionSafetyFailurePolicyChanged,
		ActionSafetyFailureWriteBlocked,
		ActionSafetyFailureConfirmationChanged,
		ActionSafetyFailureConfirmationRequired,
		ActionSafetyFailureToolAuthorityUnavailable,
		ActionSafetyFailureStoreContractMismatch,
		ActionSafetyFailureLevelEscalationDenied,
	} {
		if actionSafetyDecisionHasFailure(decision, blocker) {
			return blocker
		}
	}
	return ActionSafetyFailureLevelEscalationDenied
}

func actionSafetyDecisionHasFailure(decision ActionSafetyDecision, failure ActionSafetyFailureCode) bool {
	for _, blocker := range decision.Blockers {
		if blocker == failure {
			return true
		}
	}
	return false
}

func sealActionSafetyResponseProjection(value *ActionSafetyResponseProjectionV1) error {
	value.ProjectionDigest = ""
	digest, err := canonicalSHA256(value)
	if err != nil {
		return err
	}
	value.ProjectionDigest = digest
	return validateActionSafetyResponseProjection(*value)
}

func validateActionSafetyResponseProjection(value ActionSafetyResponseProjectionV1) error {
	if value.SchemaVersion != actionSafetyResponseProjectionSchema || !workflowRAGRunIDPattern.MatchString(value.SourceRunID) ||
		!workflowRAGDigestPattern.MatchString(value.ResponseDigest) || !workflowRAGDigestPattern.MatchString(value.ProjectionDigest) ||
		(value.AnswerDecision == nil) == (len(value.Candidates) == 0) {
		return errActionSafetyProjectionContract
	}
	if value.AnswerDecision != nil && validateActionSafetyDecision(*value.AnswerDecision) != nil {
		return errActionSafetyProjectionContract
	}
	for _, candidate := range value.Candidates {
		if validateActionSafetyCandidateProjection(candidate) != nil || candidate.SourceRunID != value.SourceRunID {
			return errActionSafetyProjectionContract
		}
	}
	clone := value
	clone.ProjectionDigest = ""
	want, err := canonicalSHA256(clone)
	if err != nil || want != value.ProjectionDigest {
		return errActionSafetyProjectionContract
	}
	return nil
}

func sealActionSafetyCandidateProjection(value *ActionSafetyCandidateProjectionV1) error {
	value.ProjectionDigest = ""
	digest, err := canonicalSHA256(value)
	if err != nil {
		return err
	}
	value.ProjectionDigest = digest
	return validateActionSafetyCandidateProjection(*value)
}

func validateActionSafetyCandidateProjection(value ActionSafetyCandidateProjectionV1) error {
	if value.SchemaVersion != actionSafetyCandidateProjectionSchema || !validActionSafetyReference(value.CandidateID) ||
		value.RecordVersion < 1 || !workflowRAGRunIDPattern.MatchString(value.SourceRunID) || value.ActionIndex < 0 ||
		!validActionSafetySourceReference(value.Source) || value.Source.Kind != actionSafetySourceAgentCopilotAction ||
		value.Source.ID != fmt.Sprintf("%s.action.%d", value.SourceRunID, value.ActionIndex+1) ||
		!agentCopilotContainsString([]string{actionSafetyCandidatePending, actionSafetyCandidateApproved, actionSafetyCandidateRejected}, value.ReviewState) ||
		!validActionSafetyTargetKind(value.TargetKind) || validateActionSafetyDecision(value.Decision) != nil ||
		!workflowRAGDigestPattern.MatchString(value.ProjectionDigest) {
		return errActionSafetyProjectionContract
	}
	clone := value
	clone.ProjectionDigest = ""
	want, err := canonicalSHA256(clone)
	if err != nil || want != value.ProjectionDigest {
		return errActionSafetyProjectionContract
	}
	return nil
}

func sealActionSafetyAssignmentProjection(value *ActionSafetyAssignmentProjectionV1) error {
	value.ProjectionDigest = ""
	digest, err := canonicalSHA256(value)
	if err != nil {
		return err
	}
	value.ProjectionDigest = digest
	return validateActionSafetyAssignmentProjection(*value)
}

func validateActionSafetyAssignmentProjection(value ActionSafetyAssignmentProjectionV1) error {
	if value.SchemaVersion != actionSafetyAssignmentProjectionSchema || value.AssignmentVersion < 1 ||
		!validActionSafetyReference(value.CandidateID) || value.CandidateVersion < 1 ||
		validateActionSafetyDecision(value.Decision) != nil || value.Decision.EffectiveLevel != ActionSafetyLevelHandoffReady ||
		!workflowRAGDigestPattern.MatchString(value.ProjectionDigest) {
		return errActionSafetyProjectionContract
	}
	clone := value
	clone.ProjectionDigest = ""
	want, err := canonicalSHA256(clone)
	if err != nil || want != value.ProjectionDigest {
		return errActionSafetyProjectionContract
	}
	return nil
}

func (runtime *actionSafetyRuntimeV1) CompileToolPlanCreation(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	registry workflowHTTPToolRegistry,
	definitionExact bool,
	membershipAllowed bool,
	developmentGateEnabled bool,
) (ActionSafetyPlanProjectionV1, ActionSafetyFailureCode) {
	if runtime == nil || plan.SchemaVersion != workflowHTTPToolPlanSchemaV2 || plan.SourceKind != workflowHTTPToolSourceDefinition {
		return ActionSafetyPlanProjectionV1{}, ActionSafetyFailurePayloadInvalid
	}
	source := actionSafetySourceForToolPlan(plan)
	policy, available := actionSafetyToolPolicyReference(runtime, plan, registry)
	planDigest, digestErr := workflowHTTPToolPlanDigest(plan)
	authority := ActionSafetyAuthoritySnapshot{
		ScopeAuthorized: validateWorkflowHTTPToolActionContext(ctx) == "", MembershipAllowed: membershipAllowed,
		SourceCurrent: true, PolicyCurrent: true, ActionAllowed: true,
		WorkflowDefinitionExact: definitionExact, ToolDefinitionExact: registry.definitionDigest == plan.DefinitionDigest,
		ExecutionProfileExact:     registry.profileDigest == plan.ProfileDigest,
		PlanDigestExact:           digestErr == nil && planDigest == plan.ToolPlanDigest,
		ExecutePermissionsAllowed: false, DevelopmentGateEnabled: developmentGateEnabled,
	}
	decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
		Scope:        actionSafetyScopeFromToolContext(ctx, registry.profile.Environment),
		CurrentScope: actionSafetyScopeFromToolContext(ctx, registry.profile.Environment),
		Source:       source, CurrentSource: source, SourceAvailable: true,
		Policy: policy, CurrentPolicy: policy, PolicyAvailable: available,
		Demand: actionSafetyToolDemand(plan, "pending"), Authority: authority,
		ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if failure != "" {
		return ActionSafetyPlanProjectionV1{}, failure
	}
	projection := ActionSafetyPlanProjectionV1{
		SchemaVersion: actionSafetyPlanProjectionSchema, PlanID: plan.PlanID, PlanVersion: plan.RecordVersion,
		ToolPlanDigest: plan.ToolPlanDigest, Decision: decision,
	}
	if err := sealActionSafetyPlanProjection(&projection); err != nil {
		return ActionSafetyPlanProjectionV1{}, ActionSafetyFailureStoreContractMismatch
	}
	return projection, ""
}

func (runtime *actionSafetyRuntimeV1) RevalidateToolPreDispatch(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
	confirmation *WorkflowHTTPToolConfirmationDecision,
	registry workflowHTTPToolRegistry,
	definitionExact bool,
	membershipAllowed bool,
	developmentGateEnabled bool,
) (ActionSafetyDecision, ActionSafetyFailureCode) {
	if runtime == nil || plan.ActionSafety == nil || validateActionSafetyPlanProjection(*plan.ActionSafety) != nil {
		return ActionSafetyDecision{}, ActionSafetyFailureStoreContractMismatch
	}
	currentSource := actionSafetySourceForToolPlan(plan)
	currentPolicy, policyAvailable := actionSafetyToolPolicyReference(runtime, plan, registry)
	planDigest, digestErr := workflowHTTPToolPlanDigest(plan)
	confirmationState := "pending"
	confirmationScopeExact, confirmationPlanExact := false, false
	if confirmation != nil {
		if confirmation.Outcome == WorkflowHTTPToolConfirmationApprove {
			confirmationState = "approved"
		} else if confirmation.Outcome == WorkflowHTTPToolConfirmationReject {
			confirmationState = "rejected"
		}
		confirmationScopeExact = confirmation.TenantRef == ctx.TenantRef && confirmation.WorkspaceID == ctx.WorkspaceID &&
			confirmation.ApplicationID == ctx.ApplicationID
		confirmationPlanExact = workflowHTTPToolApprovalMatchesPlan(*confirmation, plan)
	}
	previous := plan.ActionSafety.Decision
	decision, failure := runtime.compile(actionSafetyRuntimeCompileRequest{
		Previous: &previous,
		Scope: ActionSafetyScope{
			TenantRef: previous.TenantRef, WorkspaceID: previous.WorkspaceID,
			Environment: previous.Environment, ApplicationID: previous.ApplicationID,
		},
		CurrentScope: actionSafetyScopeFromToolContext(ctx, registry.profile.Environment),
		Source: ActionSafetySourceReference{
			Kind: previous.SourceKind, SchemaVersion: previous.SourceSchemaVersion, ID: previous.SourceID,
			Version: previous.SourceVersion, Digest: previous.SourceDigest,
		},
		CurrentSource: currentSource, SourceAvailable: true,
		Policy:        ActionSafetyPolicyReference{Version: previous.PolicyVersion, Digest: previous.PolicyDigest},
		CurrentPolicy: currentPolicy, PolicyAvailable: policyAvailable,
		Demand: actionSafetyToolDemand(plan, confirmationState),
		Authority: ActionSafetyAuthoritySnapshot{
			ScopeAuthorized:   validateWorkflowHTTPToolActionContext(ctx) == "",
			MembershipAllowed: membershipAllowed, SourceCurrent: true, PolicyCurrent: true, ActionAllowed: true,
			WorkflowDefinitionExact: definitionExact, ToolDefinitionExact: registry.definitionDigest == plan.DefinitionDigest,
			ExecutionProfileExact:     registry.profileDigest == plan.ProfileDigest,
			PlanDigestExact:           digestErr == nil && planDigest == plan.ToolPlanDigest,
			ExecutePermissionsAllowed: workflowHTTPToolExecutionSourceScopeAllowed(ctx.ScopeGrants, plan),
			DevelopmentGateEnabled:    developmentGateEnabled,
			ConfirmationScopeExact:    confirmationScopeExact, ConfirmationPlanExact: confirmationPlanExact,
			HumanApproved: confirmationState == "approved", CandidateApproved: definitionExact, AssignmentActive: definitionExact,
		},
		ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if failure != "" {
		return ActionSafetyDecision{}, failure
	}
	if decision.EffectiveLevel != ActionSafetyLevelToolCallable {
		return decision, actionSafetyPrimaryBlocker(decision)
	}
	return decision, ""
}

func (runtime *actionSafetyRuntimeV1) ProjectRun(
	checkpoint string,
	decisions []ActionSafetyDecision,
	sideEffects WorkflowRunSideEffects,
) (ActionSafetyRunProjectionV1, error) {
	projection := ActionSafetyRunProjectionV1{
		SchemaVersion: actionSafetyRunProjectionSchema, Checkpoint: strings.TrimSpace(checkpoint),
		Decisions: cloneActionSafetyDecisions(decisions), SideEffects: sideEffects,
	}
	if err := sealActionSafetyRunProjection(&projection); err != nil {
		return ActionSafetyRunProjectionV1{}, err
	}
	return projection, nil
}

func actionSafetyScopeFromToolContext(ctx WorkflowHTTPToolActionContext, environment string) ActionSafetyScope {
	return ActionSafetyScope{
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: environment, ApplicationID: ctx.ApplicationID,
	}
}

func actionSafetySourceForToolPlan(plan WorkflowHTTPToolActionPlan) ActionSafetySourceReference {
	return ActionSafetySourceReference{
		Kind: actionSafetySourceWorkflowHTTPToolAction, SchemaVersion: workflowHTTPToolPlanSchemaV2,
		ID: plan.PlanID, Version: 1, Digest: plan.ToolPlanDigest,
	}
}

func actionSafetyToolPolicyReference(runtime *actionSafetyRuntimeV1, plan WorkflowHTTPToolActionPlan, registry workflowHTTPToolRegistry) (ActionSafetyPolicyReference, bool) {
	ownerDigest, err := canonicalSHA256(struct {
		DefinitionDigest         string `json:"definition_digest"`
		ProfileDigest            string `json:"profile_digest"`
		WorkflowDefinitionDigest string `json:"workflow_definition_digest"`
	}{
		DefinitionDigest: registry.definitionDigest, ProfileDigest: registry.profileDigest,
		WorkflowDefinitionDigest: plan.WorkflowDefinitionDigest,
	})
	if err != nil {
		return ActionSafetyPolicyReference{}, false
	}
	return runtime.policyRef(ownerDigest)
}

func actionSafetyToolDemand(plan WorkflowHTTPToolActionPlan, confirmationState string) ActionSafetyCapabilityDemand {
	return ActionSafetyCapabilityDemand{
		Project: "radishmind", Task: actionSafetyWorkflowHTTPToolTask, ActionKind: workflowHTTPToolID,
		RiskLevel: "medium", TargetKind: actionSafetyTargetWorkflowHTTPTool, Method: plan.Method,
		RequiresConfirmation: true, ConfirmationState: confirmationState,
		ToolNetworkCalls: 1, ConfirmationConsumptions: 1,
	}
}

func sealActionSafetyPlanProjection(value *ActionSafetyPlanProjectionV1) error {
	value.ProjectionDigest = ""
	digest, err := canonicalSHA256(value)
	if err != nil {
		return err
	}
	value.ProjectionDigest = digest
	return validateActionSafetyPlanProjection(*value)
}

func validateActionSafetyPlanProjection(value ActionSafetyPlanProjectionV1) error {
	if value.SchemaVersion != actionSafetyPlanProjectionSchema || !workflowHTTPToolPlanIDPattern.MatchString(value.PlanID) ||
		value.PlanVersion < 1 || !workflowRAGDigestPattern.MatchString(value.ToolPlanDigest) ||
		validateActionSafetyDecision(value.Decision) != nil || value.Decision.SourceKind != actionSafetySourceWorkflowHTTPToolAction ||
		value.Decision.SourceID != value.PlanID || value.Decision.SourceDigest != value.ToolPlanDigest ||
		!workflowRAGDigestPattern.MatchString(value.ProjectionDigest) {
		return errActionSafetyProjectionContract
	}
	clone := value
	clone.ProjectionDigest = ""
	want, err := canonicalSHA256(clone)
	if err != nil || want != value.ProjectionDigest {
		return errActionSafetyProjectionContract
	}
	return nil
}

func sealActionSafetyRunProjection(value *ActionSafetyRunProjectionV1) error {
	value.ProjectionDigest = ""
	digest, err := canonicalSHA256(value)
	if err != nil {
		return err
	}
	value.ProjectionDigest = digest
	return validateActionSafetyRunProjection(*value)
}

func validateActionSafetyRunProjection(value ActionSafetyRunProjectionV1) error {
	if value.SchemaVersion != actionSafetyRunProjectionSchema ||
		!agentCopilotContainsString([]string{"response_normalization", "pre_dispatch", "terminal"}, value.Checkpoint) ||
		len(value.Decisions) == 0 || !workflowRAGDigestPattern.MatchString(value.ProjectionDigest) ||
		value.SideEffects.RetrievalCalls < 0 || value.SideEffects.ProviderCalls < 0 || value.SideEffects.ToolCalls < 0 ||
		value.SideEffects.ConfirmationCalls < 0 || value.SideEffects.BusinessWrites < 0 || value.SideEffects.ReplayWrites < 0 {
		return errActionSafetyProjectionContract
	}
	for _, decision := range value.Decisions {
		if validateActionSafetyDecision(decision) != nil || decision.EffectiveLevel == ActionSafetyLevelWriteAllowedByPolicy ||
			!actionSafetyObservedSideEffectsAllowed(decision, value.SideEffects) {
			return errActionSafetyProjectionContract
		}
	}
	clone := value
	clone.ProjectionDigest = ""
	want, err := canonicalSHA256(clone)
	if err != nil || want != value.ProjectionDigest {
		return errActionSafetyProjectionContract
	}
	return nil
}

func actionSafetyObservedSideEffectsAllowed(decision ActionSafetyDecision, sideEffects WorkflowRunSideEffects) bool {
	if sideEffects.BusinessWrites != 0 || sideEffects.ReplayWrites != 0 {
		return false
	}
	switch decision.EffectiveLevel {
	case ActionSafetyLevelAnswerOnly, ActionSafetyLevelProposalOnly:
		return sideEffects.ProviderCalls <= 1 && sideEffects.ToolCalls == 0 && sideEffects.ConfirmationCalls == 0
	case ActionSafetyLevelHandoffReady:
		return sideEffects.ProviderCalls == 0 && sideEffects.ToolCalls == 0 && sideEffects.ConfirmationCalls == 0
	case ActionSafetyLevelToolCallable:
		return sideEffects.ProviderCalls <= 1 && sideEffects.ToolCalls <= 1 && sideEffects.ConfirmationCalls <= 1
	case ActionSafetyLevelWriteBlocked:
		providerCallsAllowed := sideEffects.ProviderCalls == 0
		if decision.SourceKind == actionSafetySourceAgentCopilotAction {
			providerCallsAllowed = sideEffects.ProviderCalls <= 1
		}
		return providerCallsAllowed && sideEffects.ToolCalls == 0 && sideEffects.ConfirmationCalls == 0
	default:
		return false
	}
}

func cloneActionSafetyDecisions(values []ActionSafetyDecision) []ActionSafetyDecision {
	cloned := make([]ActionSafetyDecision, len(values))
	for index, value := range values {
		cloned[index] = value
		cloned[index].Blockers = append([]ActionSafetyFailureCode{}, value.Blockers...)
	}
	return cloned
}

func cloneActionSafetyRunProjection(value *ActionSafetyRunProjectionV1) *ActionSafetyRunProjectionV1 {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.Decisions = cloneActionSafetyDecisions(value.Decisions)
	return &cloned
}

func cloneActionSafetyPlanProjection(value *ActionSafetyPlanProjectionV1) *ActionSafetyPlanProjectionV1 {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.Decision.Blockers = append([]ActionSafetyFailureCode{}, value.Decision.Blockers...)
	return &cloned
}

func cloneActionSafetyAssignmentProjection(value *ActionSafetyAssignmentProjectionV1) *ActionSafetyAssignmentProjectionV1 {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.Decision.Blockers = append([]ActionSafetyFailureCode{}, value.Decision.Blockers...)
	return &cloned
}

func actionSafetyResponseDecisions(value ActionSafetyResponseProjectionV1) []ActionSafetyDecision {
	if value.AnswerDecision != nil {
		return []ActionSafetyDecision{*value.AnswerDecision}
	}
	decisions := make([]ActionSafetyDecision, 0, len(value.Candidates))
	for _, candidate := range value.Candidates {
		decisions = append(decisions, candidate.Decision)
	}
	return decisions
}
