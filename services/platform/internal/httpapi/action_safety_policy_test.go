package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestActionSafetyCompilerPositiveCompatibilityMatrix(t *testing.T) {
	compiler := ActionSafetyPolicyCompiler{}
	tests := []struct {
		name      string
		configure func(*ActionSafetyCompileInput)
		requested ActionSafetyLevel
		maximum   ActionSafetyLevel
		effective ActionSafetyLevel
		budget    ActionSafetySideEffectBudget
	}{
		{
			name: "answer only",
			configure: func(input *ActionSafetyCompileInput) {
				input.Demand.ActionKind = actionSafetyActionNone
				input.Demand.RequiresConfirmation = false
				input.Demand.ConfirmationState = "not_required"
			},
			requested: ActionSafetyLevelAnswerOnly, maximum: ActionSafetyLevelAnswerOnly, effective: ActionSafetyLevelAnswerOnly,
			budget: ActionSafetySideEffectBudget{ProviderCalls: 1},
		},
		{
			name: "proposal only",
			configure: func(input *ActionSafetyCompileInput) {
				input.Source.Kind = actionSafetySourceAgentCopilotAction
				input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
				input.Demand.ActionKind = "candidate_edit"
				input.Demand.RequiresConfirmation = true
				input.Demand.ConfirmationState = "pending"
			},
			requested: ActionSafetyLevelProposalOnly, maximum: ActionSafetyLevelProposalOnly, effective: ActionSafetyLevelProposalOnly,
			budget: ActionSafetySideEffectBudget{ProviderCalls: 1},
		},
		{
			name: "handoff ready",
			configure: func(input *ActionSafetyCompileInput) {
				input.Source.Kind = actionSafetySourceAgentCopilotAction
				input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
				input.Demand.ActionKind = "candidate_operation"
				input.Demand.TargetKind = actionSafetyTargetHumanReviewOwner
				input.Demand.HandoffRefs = 1
				input.Demand.RequiresConfirmation = true
				input.Demand.ConfirmationState = "pending"
				input.Authority.HandoffTargetAccepted = true
				input.Authority.HandoffMetadataOnly = true
			},
			requested: ActionSafetyLevelHandoffReady, maximum: ActionSafetyLevelHandoffReady, effective: ActionSafetyLevelHandoffReady,
			budget: ActionSafetySideEffectBudget{HandoffRefs: 1},
		},
		{
			name: "tool callable",
			configure: func(input *ActionSafetyCompileInput) {
				configureActionSafetyToolInput(input)
			},
			requested: ActionSafetyLevelToolCallable, maximum: ActionSafetyLevelToolCallable, effective: ActionSafetyLevelToolCallable,
			budget: ActionSafetySideEffectBudget{ToolNetworkCalls: 1, ConfirmationConsumptions: 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := actionSafetyCompileTestInput()
			test.configure(&input)
			decision, failure := compiler.Compile(input)
			if failure != "" || decision.RequestedLevel != test.requested || decision.MaximumAllowedLevel != test.maximum ||
				decision.EffectiveLevel != test.effective || decision.SideEffectBudget != test.budget || len(decision.Blockers) != 0 {
				t.Fatalf("unexpected decision: failure=%s decision=%#v", failure, decision)
			}
			payload, err := encodeActionSafetyDecision(decision)
			if err != nil {
				t.Fatalf("encode decision: %v", err)
			}
			decoded, err := decodeActionSafetyDecision(payload)
			if err != nil || !reflect.DeepEqual(decoded, decision) {
				t.Fatalf("strict round trip failed: decoded=%#v err=%v payload=%s", decoded, err, payload)
			}
		})
	}
}

func TestActionSafetyCompilerBlocksEveryWriteClassWithZeroSideEffects(t *testing.T) {
	compiler := ActionSafetyPolicyCompiler{}
	targets := []string{
		actionSafetyTargetBusinessTruth,
		actionSafetyTargetShell,
		actionSafetyTargetCode,
		actionSafetyTargetSandbox,
		actionSafetyTargetAgentLoop,
		actionSafetyTargetConnectorMutation,
		actionSafetyTargetAutomaticApply,
	}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			input := actionSafetyCompileTestInput()
			input.Source.Kind = actionSafetySourceAgentCopilotAction
			input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
			input.Demand.ActionKind = "candidate_operation"
			input.Demand.TargetKind = target
			input.Demand.BusinessWrites = 1
			input.Demand.RequiresConfirmation = true
			input.Demand.ConfirmationState = "approved"
			input.Authority.HumanApproved = true
			input.Authority.CandidateApproved = true
			input.Authority.AssignmentActive = true
			decision, failure := compiler.Compile(input)
			if failure != "" || decision.RequestedLevel != ActionSafetyLevelWriteAllowedByPolicy ||
				decision.EffectiveLevel != ActionSafetyLevelWriteBlocked || decision.SideEffectBudget != (ActionSafetySideEffectBudget{}) ||
				!actionSafetyContainsBlocker(decision.Blockers, ActionSafetyFailureWriteBlocked) {
				t.Fatalf("write target escaped fail-closed policy: failure=%s decision=%#v", failure, decision)
			}
		})
	}
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			input := actionSafetyCompileTestInput()
			configureActionSafetyToolInput(&input)
			input.Demand.Method = method
			decision, failure := compiler.Compile(input)
			if failure != "" || decision.EffectiveLevel != ActionSafetyLevelWriteBlocked || decision.SideEffectBudget != (ActionSafetySideEffectBudget{}) ||
				!actionSafetyContainsBlocker(decision.Blockers, ActionSafetyFailureWriteBlocked) {
				t.Fatalf("write method escaped fail-closed policy: failure=%s decision=%#v", failure, decision)
			}
		})
	}
}

func TestActionSafetyCompilerDriftConfirmationAndNonGrantingApproval(t *testing.T) {
	compiler := ActionSafetyPolicyCompiler{}
	t.Run("scope source and policy drift fail closed in stable order", func(t *testing.T) {
		input := actionSafetyCompileTestInput()
		input.Source.Kind = actionSafetySourceAgentCopilotAction
		input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
		input.Demand.ActionKind = "candidate_edit"
		input.Demand.RequiresConfirmation = true
		input.Demand.ConfirmationState = "pending"
		input.CurrentScope.WorkspaceID = "workspace_other"
		input.CurrentSource.Digest = actionSafetyTestDigest("b")
		input.CurrentPolicy.Digest = actionSafetyTestDigest("c")
		decision, failure := compiler.Compile(input)
		expected := []ActionSafetyFailureCode{
			ActionSafetyFailureScopeDenied,
			ActionSafetyFailureSourceChanged,
			ActionSafetyFailurePolicyChanged,
			ActionSafetyFailureLevelEscalationDenied,
		}
		if failure != "" || decision.MaximumAllowedLevel != ActionSafetyLevelAnswerOnly || decision.EffectiveLevel != ActionSafetyLevelAnswerOnly ||
			!reflect.DeepEqual(decision.Blockers, expected) {
			t.Fatalf("drift did not fail closed canonically: failure=%s decision=%#v", failure, decision)
		}
	})

	t.Run("human candidate and assignment state cannot grant a tool", func(t *testing.T) {
		input := actionSafetyCompileTestInput()
		input.Source.Kind = actionSafetySourceAgentCopilotAction
		input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
		input.Demand.ActionKind = "candidate_operation"
		input.Demand.RequiresConfirmation = true
		input.Demand.ConfirmationState = "approved"
		input.Authority.HumanApproved = true
		input.Authority.CandidateApproved = true
		input.Authority.AssignmentActive = true
		decision, failure := compiler.Compile(input)
		if failure != "" || decision.MaximumAllowedLevel != ActionSafetyLevelProposalOnly || decision.EffectiveLevel != ActionSafetyLevelProposalOnly {
			t.Fatalf("non-granting states elevated the decision: failure=%s decision=%#v", failure, decision)
		}
	})

	t.Run("missing and changed confirmation cannot dispatch", func(t *testing.T) {
		for name, test := range map[string]struct {
			configure func(*ActionSafetyCompileInput)
			blocker   ActionSafetyFailureCode
		}{
			"missing": {configure: func(input *ActionSafetyCompileInput) { input.Demand.ConfirmationState = "pending" }, blocker: ActionSafetyFailureConfirmationRequired},
			"changed": {configure: func(input *ActionSafetyCompileInput) { input.Authority.ConfirmationPlanExact = false }, blocker: ActionSafetyFailureConfirmationChanged},
		} {
			t.Run(name, func(t *testing.T) {
				input := actionSafetyCompileTestInput()
				configureActionSafetyToolInput(&input)
				test.configure(&input)
				decision, failure := compiler.Compile(input)
				if failure != "" || decision.EffectiveLevel == ActionSafetyLevelToolCallable || decision.SideEffectBudget.ToolNetworkCalls != 0 ||
					!actionSafetyContainsBlocker(decision.Blockers, test.blocker) || !actionSafetyContainsBlocker(decision.Blockers, ActionSafetyFailureLevelEscalationDenied) {
					t.Fatalf("confirmation failure retained tool authority: failure=%s decision=%#v", failure, decision)
				}
			})
		}
	})
}

func TestActionSafetyStrictCodecRejectsUnknownDuplicateOrderingAndDigestDrift(t *testing.T) {
	input := actionSafetyCompileTestInput()
	input.Source.Kind = actionSafetySourceAgentCopilotAction
	input.CurrentSource.Kind = actionSafetySourceAgentCopilotAction
	input.Demand.ActionKind = "candidate_edit"
	input.Demand.RequiresConfirmation = true
	input.Demand.ConfirmationState = "pending"
	input.CurrentScope.WorkspaceID = "workspace_other"
	input.CurrentSource.Digest = actionSafetyTestDigest("b")
	decision, failure := (ActionSafetyPolicyCompiler{}).Compile(input)
	if failure != "" {
		t.Fatalf("compile strict codec fixture: %s", failure)
	}
	payload, err := encodeActionSafetyDecision(decision)
	if err != nil {
		t.Fatalf("encode strict codec fixture: %v", err)
	}

	unknown := append(append([]byte(nil), payload[:len(payload)-1]...), []byte(`,"credential":"forbidden"}`)...)
	if _, err := decodeActionSafetyDecision(unknown); err == nil {
		t.Fatal("strict codec accepted an unknown sensitive field")
	}
	duplicate := []byte(`{"schema_version":"action_safety_decision.v1","schema_version":"action_safety_decision.v1"}`)
	if _, err := decodeActionSafetyDecision(duplicate); err == nil {
		t.Fatal("strict codec accepted a duplicate field")
	}
	var missing map[string]any
	if err := json.Unmarshal(payload, &missing); err != nil {
		t.Fatalf("decode missing-field fixture: %v", err)
	}
	delete(missing, "side_effect_budget")
	missingPayload, _ := json.Marshal(missing)
	if _, err := decodeActionSafetyDecision(missingPayload); err == nil {
		t.Fatal("strict codec accepted a missing required zero-value field")
	}

	corrupted := decision
	corrupted.SourceDigest = actionSafetyTestDigest("d")
	corruptedPayload, _ := json.Marshal(corrupted)
	if _, err := decodeActionSafetyDecision(corruptedPayload); err == nil {
		t.Fatal("strict codec accepted source digest drift without a new decision digest")
	}
	corrupted = decision
	corrupted.PolicyDigest = actionSafetyTestDigest("e")
	corruptedPayload, _ = json.Marshal(corrupted)
	if _, err := decodeActionSafetyDecision(corruptedPayload); err == nil {
		t.Fatal("strict codec accepted policy digest drift without a new decision digest")
	}

	corrupted = decision
	corrupted.Blockers = append([]ActionSafetyFailureCode(nil), decision.Blockers...)
	corrupted.Blockers[0], corrupted.Blockers[1] = corrupted.Blockers[1], corrupted.Blockers[0]
	corrupted.DecisionDigest, _ = actionSafetyDecisionDigest(corrupted)
	corruptedPayload, _ = json.Marshal(corrupted)
	if _, err := decodeActionSafetyDecision(corruptedPayload); err == nil {
		t.Fatal("strict codec accepted non-canonical blocker order")
	}

	corrupted = decision
	corrupted.RequestedLevel = ActionSafetyLevelToolCallable
	corrupted.DecisionDigest, _ = actionSafetyDecisionDigest(corrupted)
	corruptedPayload, _ = json.Marshal(corrupted)
	if _, err := decodeActionSafetyDecision(corruptedPayload); err == nil {
		t.Fatal("strict codec accepted an invalid level transition")
	}

	for _, mutate := range []func(*ActionSafetyDecision){
		func(value *ActionSafetyDecision) { value.MaximumAllowedLevel = ActionSafetyLevelWriteAllowedByPolicy },
		func(value *ActionSafetyDecision) { value.EffectiveLevel = ActionSafetyLevelWriteAllowedByPolicy },
	} {
		corrupted = decision
		corrupted.Blockers = append([]ActionSafetyFailureCode(nil), decision.Blockers...)
		mutate(&corrupted)
		corrupted.DecisionDigest, _ = actionSafetyDecisionDigest(corrupted)
		corruptedPayload, _ = json.Marshal(corrupted)
		if _, err := decodeActionSafetyDecision(corruptedPayload); err == nil {
			t.Fatal("strict codec accepted write_allowed_by_policy as an allowed or effective result")
		}
	}
}

func TestActionSafetyCompilerRejectsInvalidSourceActionAndDemandShape(t *testing.T) {
	for name, mutate := range map[string]func(*ActionSafetyCompileInput){
		"unknown source":       func(input *ActionSafetyCompileInput) { input.Source.Kind = "model_reported_authority" },
		"unknown task":         func(input *ActionSafetyCompileInput) { input.Demand.Task = "execute_anything" },
		"unknown action":       func(input *ActionSafetyCompileInput) { input.Demand.ActionKind = "arbitrary_command" },
		"invalid risk":         func(input *ActionSafetyCompileInput) { input.Demand.RiskLevel = "critical" },
		"handoff without ref":  func(input *ActionSafetyCompileInput) { input.Demand.TargetKind = actionSafetyTargetHumanReviewOwner },
		"tool without network": func(input *ActionSafetyCompileInput) { input.Demand.TargetKind = actionSafetyTargetWorkflowHTTPTool },
		"invalid environment":  func(input *ActionSafetyCompileInput) { input.Scope.Environment = "production" },
		"multiple tool calls": func(input *ActionSafetyCompileInput) {
			configureActionSafetyToolInput(input)
			input.Demand.ToolNetworkCalls = 2
		},
	} {
		t.Run(name, func(t *testing.T) {
			input := actionSafetyCompileTestInput()
			mutate(&input)
			decision, failure := (ActionSafetyPolicyCompiler{}).Compile(input)
			if failure != ActionSafetyFailurePayloadInvalid || !reflect.DeepEqual(decision, ActionSafetyDecision{}) {
				t.Fatalf("invalid input did not fail before decision creation: failure=%s decision=%#v", failure, decision)
			}
		})
	}
}

func TestActionSafetyCompatibilityConsumesExistingCanonicalRegistries(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "contracts", "copilot-response.schema.json"))
	if err != nil {
		t.Fatalf("read CopilotResponse schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
		Definitions map[string]struct {
			Properties map[string]struct {
				Enum []string `json:"enum"`
			} `json:"properties"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatalf("decode CopilotResponse schema: %v", err)
	}
	if actions := schema.Definitions["candidate_action"].Properties["kind"].Enum; !reflect.DeepEqual(actions, agentCopilotCanonicalActionKinds[:]) {
		t.Fatalf("Action Safety action compatibility drifted from CopilotResponse: schema=%v runtime=%v", actions, agentCopilotCanonicalActionKinds)
	}
	if risks := schema.Properties["risk_level"].Enum; !reflect.DeepEqual(risks, []string{"low", "medium", "high"}) {
		t.Fatalf("Action Safety risk compatibility drifted from CopilotResponse: %v", risks)
	}
	for _, project := range []string{"radishflow", "radish"} {
		for _, task := range agentCopilotCanonicalTasks(project) {
			for _, action := range agentCopilotCanonicalActionKinds {
				if !validActionSafetySourceCompatibility(actionSafetySourceAgentCopilotAction, actionSafetyCopilotResponseSchemaVersion, project, task, action, "medium", actionSafetyTargetNone) {
					t.Fatalf("existing canonical combination was not consumed: project=%s task=%s action=%s", project, task, action)
				}
			}
		}
	}
	if validActionSafetySourceCompatibility(actionSafetySourceAgentCopilotAction, actionSafetyCopilotResponseSchemaVersion, "radish", "new_task", "candidate_edit", "medium", actionSafetyTargetNone) ||
		validActionSafetySourceCompatibility(actionSafetySourceWorkflowHTTPToolAction, workflowHTTPToolPlanSchemaV2, "radishmind", actionSafetyWorkflowHTTPToolTask, "another_tool", "medium", actionSafetyTargetWorkflowHTTPTool) {
		t.Fatal("Action Safety accepted a task or tool outside the existing canonical owners")
	}

	authorityType := reflect.TypeOf(ActionSafetyAuthoritySnapshot{})
	for index := 0; index < authorityType.NumField(); index++ {
		if authorityType.Field(index).Type.Kind() != reflect.Bool {
			t.Fatalf("authority projection created a parallel permission registry field: %s %s", authorityType.Field(index).Name, authorityType.Field(index).Type)
		}
	}
}

func TestActionSafetyGoContractMatchesCommittedSchema(t *testing.T) {
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "contracts", "action-safety-decision.schema.json"))
	if err != nil {
		t.Fatalf("read Action Safety schema: %v", err)
	}
	var schema struct {
		Required    []string `json:"required"`
		Definitions map[string]struct {
			Enum []string `json:"enum"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatalf("decode Action Safety schema: %v", err)
	}
	if !reflect.DeepEqual(schema.Required, actionSafetyDecisionRequiredFields) {
		t.Fatalf("Go required fields drifted from schema: schema=%v go=%v", schema.Required, actionSafetyDecisionRequiredFields)
	}
	expectedLevels := []string{
		string(ActionSafetyLevelAnswerOnly), string(ActionSafetyLevelProposalOnly), string(ActionSafetyLevelHandoffReady),
		string(ActionSafetyLevelToolCallable), string(ActionSafetyLevelWriteBlocked), string(ActionSafetyLevelWriteAllowedByPolicy),
	}
	if !reflect.DeepEqual(schema.Definitions["level"].Enum, expectedLevels) {
		t.Fatalf("Go level matrix drifted from schema: schema=%v go=%v", schema.Definitions["level"].Enum, expectedLevels)
	}
	expectedBlockers := make([]string, 0, len(actionSafetyBlockerOrder))
	for _, blocker := range actionSafetyBlockerOrder {
		expectedBlockers = append(expectedBlockers, string(blocker))
	}
	if !reflect.DeepEqual(schema.Definitions["blocker"].Enum, expectedBlockers) {
		t.Fatalf("Go blocker order drifted from schema: schema=%v go=%v", schema.Definitions["blocker"].Enum, expectedBlockers)
	}
}

func actionSafetyCompileTestInput() ActionSafetyCompileInput {
	scope := ActionSafetyScope{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", Environment: "development", ApplicationID: "app_abcdefghijklmnop"}
	source := ActionSafetySourceReference{
		Kind: actionSafetySourceCopilotResponse, SchemaVersion: actionSafetyCopilotResponseSchemaVersion,
		ID: "response_demo", Version: 1, Digest: actionSafetyTestDigest("a"),
	}
	policy := ActionSafetyPolicyReference{Version: "action_safety_policy.v1", Digest: actionSafetyTestDigest("f")}
	return ActionSafetyCompileInput{
		DecisionID: "asd_abcdefghijklmnop", Scope: scope, CurrentScope: scope,
		Source: source, CurrentSource: source, SourceAvailable: true,
		Policy: policy, CurrentPolicy: policy, PolicyAvailable: true,
		Demand: ActionSafetyCapabilityDemand{
			Project: "radish", Task: "answer_docs_question", ActionKind: actionSafetyActionNone, RiskLevel: "low", TargetKind: actionSafetyTargetNone,
			Method: "none", RequiresConfirmation: false, ConfirmationState: "not_required", ProviderCalls: 1,
		},
		Authority: ActionSafetyAuthoritySnapshot{
			ScopeAuthorized: true, MembershipAllowed: true, SourceCurrent: true, PolicyCurrent: true, ActionAllowed: true,
		},
		ActorRef: "subject_demo", RequestID: "request_demo", AuditRef: "audit_demo", CreatedAt: "2026-08-29T08:00:00Z",
	}
}

func configureActionSafetyToolInput(input *ActionSafetyCompileInput) {
	input.Source = ActionSafetySourceReference{
		Kind: actionSafetySourceWorkflowHTTPToolAction, SchemaVersion: workflowHTTPToolPlanSchemaV2,
		ID: "wtap_abcdefghijklmnop", Version: 1, Digest: actionSafetyTestDigest("b"),
	}
	input.CurrentSource = input.Source
	input.Demand = ActionSafetyCapabilityDemand{
		Project: "radishmind", Task: actionSafetyWorkflowHTTPToolTask, ActionKind: workflowHTTPToolID, RiskLevel: "medium",
		TargetKind: actionSafetyTargetWorkflowHTTPTool, Method: "GET", RequiresConfirmation: true, ConfirmationState: "approved",
		ToolNetworkCalls: 1, ConfirmationConsumptions: 1,
	}
	input.Authority.WorkflowDefinitionExact = true
	input.Authority.ToolDefinitionExact = true
	input.Authority.ExecutionProfileExact = true
	input.Authority.PlanDigestExact = true
	input.Authority.ExecutePermissionsAllowed = true
	input.Authority.DevelopmentGateEnabled = true
	input.Authority.ConfirmationScopeExact = true
	input.Authority.ConfirmationPlanExact = true
}

func actionSafetyTestDigest(character string) string {
	return "sha256:" + strings.Repeat(character, 64)
}
