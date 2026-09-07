package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestActionSafetyReadProjectionUsesStrictMetadataOnlyAllowlist(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	authority, failure := fixture.service.resolveAuthority(fixture.ctx)
	if failure != "" {
		t.Fatalf("resolve Agent Copilot authority: %s", failure)
	}
	runtime := newActionSafetyRuntimeV1("development")
	runtime.now = func() time.Time { return time.Date(2026, 8, 30, 2, 0, 0, 0, time.UTC) }
	response := AgentCopilotResponse{
		SchemaVersion: 1, Status: "partial", Project: "radishflow", Task: "explain_diagnostics",
		Summary: "Candidate available.", Answers: []AgentCopilotResponseAnswer{}, Issues: []AgentCopilotResponseIssue{},
		ProposedActions: []AgentCopilotResponseAction{{
			Kind: "candidate_edit", Title: "Review edit", Rationale: "Address the diagnostic.",
			RiskLevel: "medium", RequiresConfirmation: true,
		}},
		Citations: []AgentCopilotResponseCitation{}, Confidence: 0.7,
		RiskLevel: "medium", RequiresConfirmation: true,
	}
	projection, safetyFailure := runtime.NormalizeAgentCopilotResponse(
		fixture.ctx, "run_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", response, authority,
	)
	view := actionSafetyReadFromResponse(&projection)
	if safetyFailure != "" || view == nil || validateActionSafetyReadProjection(view) != nil ||
		view.Status != actionSafetyReadStatusRecorded || view.Owner.Kind != "agent_copilot_response" ||
		len(view.Decisions) != 1 || view.Decisions[0].EffectiveLevel != ActionSafetyLevelProposalOnly {
		t.Fatalf("project response read view: failure=%s view=%#v", safetyFailure, view)
	}
	payload, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal response read view: %v", err)
	}
	for _, forbidden := range []string{
		`"actor_ref"`, `"request_id"`, `"audit_ref"`, `"summary"`, `"rationale"`,
		`"public_arguments"`, `"url"`, `"headers"`, `"credential"`, `"provider_raw_response"`,
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("read projection leaked forbidden field %s: %s", forbidden, payload)
		}
	}
}

func TestActionSafetyReadProjectionDistinguishesLegacyAndNotApplicable(t *testing.T) {
	legacyAssignment := AgentCopilotRuntimeAssignmentV1{
		AssignmentID: "acra_aaaaaaaaaaaaaaaa", AssignmentVersion: 1,
		AssignmentDigest: workflowRAGSHA256("legacy assignment"),
	}
	assignmentView := actionSafetyReadFromAssignment(&legacyAssignment)
	if assignmentView == nil || assignmentView.Status != actionSafetyReadStatusLegacy ||
		len(assignmentView.Decisions) != 0 || assignmentView.ProjectionDigest != "" || assignmentView.ObservedSideEffects != nil {
		t.Fatalf("legacy assignment was not explicit: %#v", assignmentView)
	}
	legacyPlan := WorkflowHTTPToolActionPlan{
		PlanID: "wtap_aaaaaaaaaaaaaaaa", RecordVersion: 1,
		ToolPlanDigest: workflowRAGSHA256("legacy plan"),
	}
	planView := actionSafetyReadFromPlan(&legacyPlan)
	if planView == nil || planView.Status != actionSafetyReadStatusLegacy || planView.Owner.Kind != "workflow_http_tool_action_plan" {
		t.Fatalf("legacy plan was not explicit: %#v", planView)
	}
	legacyAgentRun := WorkflowRunRecord{SchemaVersion: agentCopilotRunV7Schema, RunID: "run_legacy_agent", RecordVersion: 1}
	if runView := actionSafetyReadFromRun(&legacyAgentRun); runView == nil || runView.Status != actionSafetyReadStatusLegacy {
		t.Fatalf("legacy eligible Run was not explicit: %#v", runView)
	}
	standardRun := WorkflowRunRecord{SchemaVersion: workflowRunRecordSchemaVersion, RunID: "run_standard", RecordVersion: 1}
	if view := actionSafetyReadFromRun(&standardRun); view != nil {
		t.Fatalf("non-Action-Safety Run received a legacy marker: %#v", view)
	}
}

func TestActionSafetyRunReadProjectionBindsObservedSideEffects(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
	view := actionSafetyReadFromRun(result.Run)
	if result.FailureCode != "" || view == nil || view.Status != actionSafetyReadStatusRecorded ||
		view.ObservedSideEffects == nil || view.ObservedSideEffects.ProviderCalls != 1 ||
		view.ObservedSideEffects.ToolCalls != 0 || view.ObservedSideEffects.BusinessWrites != 0 ||
		view.ObservedSideEffects.ReplayWrites != 0 || len(view.Decisions) != 1 ||
		view.Decisions[0].EffectiveLevel != ActionSafetyLevelAnswerOnly {
		t.Fatalf("Run read projection did not bind frozen side effects: result=%#v view=%#v", result, view)
	}
	corrupted := *result.Run
	corrupted.ActionSafety = cloneActionSafetyRunProjection(result.Run.ActionSafety)
	corrupted.ActionSafety.ProjectionDigest = workflowRAGSHA256("corrupted read projection")
	if view = actionSafetyReadFromRun(&corrupted); view != nil {
		t.Fatalf("corrupted Run projection was exposed as a read view: %#v", view)
	}
}
