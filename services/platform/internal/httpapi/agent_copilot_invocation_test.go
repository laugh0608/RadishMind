package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type agentCopilotInvocationFixture struct {
	service  agentCopilotInvocationService
	ctx      AgentCopilotRuntimeContext
	runStore workflowRunStore
	runtime  *memoryAgentCopilotRuntimeRepository
	bridge   *workflowExecutorTestBridge
	catalog  *memoryApplicationCatalogRepository
	batchC   *agentCopilotBatchCFixture
}

func TestAgentCopilotInvocationUsesCanonicalContractAndCallsGatewayOnce(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	input := validAgentCopilotInvocationInput()
	result := fixture.service.Invoke(fixture.ctx, input)
	if result.FailureCode != "" || result.Run == nil || result.Response == nil ||
		result.Run.SchemaVersion != agentCopilotRunV7Schema || result.Run.Status != WorkflowRunStatusSucceeded ||
		result.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent Copilot invocation did not complete exactly once: result=%#v run=%#v calls=%d", result, result.Run, fixture.bridge.callCount())
	}
	document, err := agentCopilotRunDocument(*result.Run)
	if err != nil || validateAgentCopilotRun(document) != nil || document.Output != "" ||
		document.Project != "radishflow" || document.Task != input.Task || document.ResponseDigest == "" {
		t.Fatalf("Agent Copilot v7 metadata contract drifted: document=%#v err=%v", document, err)
	}
	request := string(fixture.bridge.lastRequest())
	if !strings.Contains(request, `"project":"radishflow"`) ||
		!strings.Contains(request, `"mode":"advisory"`) ||
		!strings.Contains(request, `"allow_tool_calls":false`) ||
		strings.Contains(request, `"northbound"`) ||
		strings.Contains(request, input.ClientInvocationKey) {
		t.Fatalf("Gateway packet is not canonical or leaks the idempotency key: %s", request)
	}
	replayed := fixture.service.Invoke(fixture.ctx, input)
	if !replayed.IdempotentReplay || replayed.Response != nil || replayed.FailureCode != "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("terminal retry replayed provider or fabricated response: %#v calls=%d", replayed, fixture.bridge.callCount())
	}
	conflict := input
	conflict.Context = map[string]any{"selected_unit_ids": []any{"unit-2"}}
	if result := fixture.service.Invoke(fixture.ctx, conflict); result.FailureCode != AgentCopilotInvocationFailureInputInvalid || fixture.bridge.callCount() != 1 {
		t.Fatalf("same idempotency key accepted different canonical input: %#v", result)
	}
}

func TestAgentCopilotInvocationQuotaRejectsBeforeProviderAndReplayDoesNotConsume(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	repository := newMemoryGatewayRequestQuotaRepository()
	quotaContext := GatewayRequestQuotaContext{
		RequestContext: context.Background(), TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID,
		Environment: "test", ApplicationID: fixture.ctx.ApplicationID, ActorRef: fixture.ctx.ActorRef,
		RequestID: "request-agent-quota-policy", AuditRef: "audit-agent-quota-policy",
	}
	now := time.Date(2026, 8, 9, 17, 30, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 1, now); err != nil {
		t.Fatalf("put Agent quota policy: %v", err)
	}
	quotaBridge := newGatewayRequestQuotaBridgeClient(fixture.bridge, repository)
	quotaBridge.now = func() time.Time { return now }
	fixture.service.bridge = quotaBridge
	input := validAgentCopilotInvocationInput()
	input.ClientInvocationKey = "agent-quota-first"
	firstContext := fixture.ctx
	firstContext.RequestContext = gatewayRequestQuotaBridgeTestContextForRoute(quotaContext, "request-agent-quota-first", "POST "+agentCopilotInvocationRoute)
	first := fixture.service.Invoke(firstContext, input)
	if first.FailureCode != "" || first.Run == nil || first.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent quota first admission drifted: result=%+v calls=%d", first, fixture.bridge.callCount())
	}
	replayContext := fixture.ctx
	replayContext.RequestContext = gatewayRequestQuotaBridgeTestContextForRoute(quotaContext, "request-agent-quota-replay", "POST "+agentCopilotInvocationRoute)
	replay := fixture.service.Invoke(replayContext, input)
	if !replay.IdempotentReplay || replay.FailureCode != "" || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent replay consumed quota or provider: result=%+v calls=%d", replay, fixture.bridge.callCount())
	}
	overInput := validAgentCopilotInvocationInput()
	overInput.ClientInvocationKey = "agent-quota-over"
	overContext := fixture.ctx
	overContext.RequestContext = gatewayRequestQuotaBridgeTestContextForRoute(quotaContext, "request-agent-quota-over", "POST "+agentCopilotInvocationRoute)
	over := fixture.service.Invoke(overContext, overInput)
	if over.FailureCode != GatewayRequestQuotaFailureExceeded || over.Run == nil ||
		over.Run.SideEffects.ProviderCalls != 0 || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent quota rejection crossed provider: result=%+v calls=%d", over, fixture.bridge.callCount())
	}
}

func TestAgentCopilotInvocationAcceptsCanonicalPartialGatewayResponse(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return agentCopilotGatewayEnvelope(map[string]any{
			"schema_version": 1, "status": "partial", "project": "radishflow", "task": "explain_diagnostics",
			"summary": "A constrained candidate edit is available.", "answers": []any{},
			"issues": []any{map[string]any{"code": "not_converged", "message": "The selected unit did not converge.", "severity": "warning"}},
			"proposed_actions": []any{map[string]any{
				"kind": "candidate_edit", "title": "Review a constrained edit", "rationale": "Address the diagnostic.",
				"risk_level": "medium", "requires_confirmation": true,
			}},
			"citations": []any{}, "confidence": 0.7, "risk_level": "medium", "requires_confirmation": true,
		}), nil
	}

	result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
	if result.Run == nil {
		t.Fatalf("canonical partial response did not create a Run: %#v", result)
	}
	document, err := agentCopilotRunDocument(*result.Run)
	if result.FailureCode != "" || result.Response == nil ||
		err != nil || result.Run.Status != WorkflowRunStatusSucceeded || document.ResponseStatus != "partial" ||
		result.Response.Status != "partial" || fixture.bridge.callCount() != 1 {
		t.Fatalf("canonical partial response did not complete: result=%#v run=%#v calls=%d", result, result.Run, fixture.bridge.callCount())
	}
}

func TestAgentCopilotInvocationRejectsPolicyAndResponseRelaxationBeforeEffects(t *testing.T) {
	t.Run("input policy", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		for name, mutate := range map[string]func(*AgentCopilotInvocationInput){
			"cross project task": func(input *AgentCopilotInvocationInput) { input.Task = "answer_docs_question" },
			"locale":             func(input *AgentCopilotInvocationInput) { input.Locale = "fr-FR" },
			"context field":      func(input *AgentCopilotInvocationInput) { input.Context["credential"] = "secret" },
			"artifact kind": func(input *AgentCopilotInvocationInput) {
				input.Artifacts = []AgentCopilotArtifact{{Kind: "binary", Role: "primary", Name: "bad", MIMEType: "x", Content: "bad"}}
			},
			"artifact budget": func(input *AgentCopilotInvocationInput) {
				input.Artifacts = []AgentCopilotArtifact{{Kind: "text", Role: "primary", Name: "large", MIMEType: "text/plain", Content: strings.Repeat("x", agentCopilotMaximumArtifactItemBytes+1)}}
			},
		} {
			t.Run(name, func(t *testing.T) {
				input := validAgentCopilotInvocationInput()
				input.ClientInvocationKey += "-" + strings.ReplaceAll(name, " ", "-")
				mutate(&input)
				result := fixture.service.Invoke(fixture.ctx, input)
				if result.FailureCode != AgentCopilotInvocationFailureInputInvalid || fixture.bridge.callCount() != 0 {
					t.Fatalf("invalid input crossed provider boundary: %#v calls=%d", result, fixture.bridge.callCount())
				}
			})
		}
	})
	t.Run("response confirmation", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			return agentCopilotGatewayEnvelope(map[string]any{
				"schema_version": 1, "status": "ok", "project": "radishflow", "task": "explain_diagnostics",
				"summary": "unsafe", "answers": []any{}, "issues": []any{},
				"proposed_actions": []any{map[string]any{
					"kind": "candidate_edit", "title": "edit", "rationale": "change",
					"risk_level": "medium", "requires_confirmation": false,
				}},
				"citations": []any{}, "confidence": 0.5, "risk_level": "medium", "requires_confirmation": false,
			}), nil
		}
		result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if result.FailureCode != AgentCopilotInvocationFailureResponseContract || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusFailed || fixture.bridge.callCount() != 1 {
			t.Fatalf("unsafe response was accepted: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
}

func TestAgentCopilotInvocationAuthorityDriftAndConcurrentDuplicateFailClosed(t *testing.T) {
	t.Run("authority drift", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		fixture.service.runStore = authorityDriftAgentCopilotRunStore{
			workflowRunStore: fixture.runStore,
			drift: func() {
				assignment, events, err := fixture.runtime.Read(fixture.ctx)
				if err != nil {
					t.Fatalf("read assignment: %v", err)
				}
				assignment.State = "revoked"
				revoked := "2026-07-25T16:00:00Z"
				assignment.RevokedAt, assignment.AssignmentVersion, assignment.UpdatedAt = &revoked, assignment.AssignmentVersion+1, revoked
				assignment.AssignmentDigest, _ = agentCopilotRuntimeAssignmentDigest(assignment)
				event := events[len(events)-1]
				event.EventID, event.EventSequence = "acrae_bbbbbbbbbbbbbbbb", len(events)+1
				event.Action, event.ExpectedAssignmentVersion, event.ResultingAssignmentVersion = "revoke", assignment.AssignmentVersion-1, assignment.AssignmentVersion
				event.AssignmentDigest, event.OccurredAt = assignment.AssignmentDigest, revoked
				if err = fixture.runtime.Apply(fixture.ctx, event.ExpectedAssignmentVersion, assignment, event); err != nil {
					t.Fatalf("revoke assignment: %v", err)
				}
			},
		}
		result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if result.FailureCode != AgentCopilotRuntimeFailureAuthorityChanged || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusFailed || fixture.bridge.callCount() != 0 {
			t.Fatalf("authority drift crossed provider checkpoint: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	t.Run("candidate supersede then explicit replacement", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		replacement := fixture.batchC.createApprovedCandidate(t, "candidate-agent-invocation-replacement")
		stale := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if stale.FailureCode != AgentCopilotRuntimeFailureCandidate || stale.Run != nil || fixture.bridge.callCount() != 0 {
			t.Fatalf("superseded candidate crossed provider boundary: %#v calls=%d", stale, fixture.bridge.callCount())
		}
		replaced := fixture.batchC.runtimeService.Decide(fixture.ctx, AgentCopilotRuntimeDecisionInput{
			ExpectedAssignmentVersion: 1, Action: "replace", CandidateID: replacement.CandidateID,
		})
		if replaced.FailureCode != "" || replaced.Assignment == nil || replaced.Assignment.AssignmentVersion != 2 {
			t.Fatalf("replace superseded Agent Copilot assignment: %#v", replaced)
		}
		input := validAgentCopilotInvocationInput()
		input.ClientInvocationKey = "client-agent-invocation-replaced"
		result := fixture.service.Invoke(fixture.ctx, input)
		if result.FailureCode != "" || result.Run == nil || result.Run.Status != WorkflowRunStatusSucceeded ||
			fixture.bridge.callCount() != 1 {
			t.Fatalf("explicit replacement did not establish new exact authority: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	t.Run("concurrent duplicate", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		entered, release := make(chan struct{}), make(chan struct{})
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			close(entered)
			<-release
			return successfulAgentCopilotGatewayEnvelope(), nil
		}
		firstResult := make(chan AgentCopilotInvocationResult, 1)
		go func() { firstResult <- fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput()) }()
		<-entered
		duplicate := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if duplicate.FailureCode != AgentCopilotInvocationFailureDuplicateRunning || fixture.bridge.callCount() != 1 {
			t.Fatalf("concurrent duplicate crossed provider boundary: %#v calls=%d", duplicate, fixture.bridge.callCount())
		}
		close(release)
		if first := <-firstResult; first.FailureCode != "" || fixture.bridge.callCount() != 1 {
			t.Fatalf("first invocation did not finish exactly once: %#v calls=%d", first, fixture.bridge.callCount())
		}
	})
}

func TestAgentCopilotInvocationCancellationOutcomeAndResponseContractAreTerminal(t *testing.T) {
	t.Run("canceled", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		requestContext, cancel := context.WithCancel(context.Background())
		cancel()
		fixture.ctx.RequestContext = requestContext
		fixture.bridge.handle = func(ctx context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			return bridge.GatewayEnvelope{}, ctx.Err()
		}
		result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if result.FailureCode != AgentCopilotInvocationFailureCanceled || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusCanceled || result.Run.SideEffects.ProviderCalls != 1 ||
			fixture.bridge.callCount() != 1 {
			t.Fatalf("cancellation was not recorded terminally: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	t.Run("transport outcome unknown", func(t *testing.T) {
		fixture := newAgentCopilotInvocationFixture(t)
		fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
			return bridge.GatewayEnvelope{}, errors.New("private provider transport detail")
		}
		result := fixture.service.Invoke(fixture.ctx, validAgentCopilotInvocationInput())
		if result.FailureCode != AgentCopilotInvocationFailureOutcomeUnknown || result.Run == nil ||
			result.Run.Status != WorkflowRunStatusOutcomeUnknown || fixture.bridge.callCount() != 1 ||
			strings.Contains(result.FailureSummary, "private provider") {
			t.Fatalf("transport ambiguity was not sanitized and terminal: %#v calls=%d", result, fixture.bridge.callCount())
		}
	})
	for name, response := range map[string]map[string]any{
		"unknown field": {
			"schema_version": 1, "status": "ok", "project": "radishflow", "task": "explain_diagnostics",
			"summary": "invalid", "answers": []any{}, "issues": []any{}, "proposed_actions": []any{},
			"citations": []any{}, "confidence": 0.5, "risk_level": "low", "requires_confirmation": false,
			"provider_token": "forbidden",
		},
		"visible byte budget": {
			"schema_version": 1, "status": "ok", "project": "radishflow", "task": "explain_diagnostics",
			"summary": strings.Repeat("界", agentCopilotMaximumVisibleResponseTextByte), "answers": []any{},
			"issues": []any{}, "proposed_actions": []any{}, "citations": []any{},
			"confidence": 0.5, "risk_level": "low", "requires_confirmation": false,
		},
		"dangling citation": {
			"schema_version": 1, "status": "ok", "project": "radishflow", "task": "explain_diagnostics",
			"summary": "invalid", "answers": []any{map[string]any{"text": "answer", "citation_ids": []any{"missing"}}},
			"issues": []any{}, "proposed_actions": []any{}, "citations": []any{},
			"confidence": 0.5, "risk_level": "low", "requires_confirmation": false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := newAgentCopilotInvocationFixture(t)
			fixture.bridge.handle = func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
				return agentCopilotGatewayEnvelope(response), nil
			}
			input := validAgentCopilotInvocationInput()
			input.ClientInvocationKey += "-" + strings.ReplaceAll(name, " ", "-")
			result := fixture.service.Invoke(fixture.ctx, input)
			if result.FailureCode != AgentCopilotInvocationFailureResponseContract || result.Run == nil ||
				result.Run.Status != WorkflowRunStatusFailed || fixture.bridge.callCount() != 1 {
				t.Fatalf("invalid response crossed the terminal contract: %#v calls=%d", result, fixture.bridge.callCount())
			}
		})
	}
}

func TestAgentCopilotSessionV3DelegatesUniqueInvocationService(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	resolver := newExactApplicationInteractionAuthorityResolver(
		fixture.catalog, nil, nil, workflowRAGApplicationAuthorityResolver{},
	)
	resolver.resolveAgentCopilot = func(ctx AgentCopilotRuntimeContext) (AgentCopilotRuntimeAuthorityV3, string) {
		authority, failure := fixture.service.resolveAuthority(ctx)
		return authority.Snapshot, failure
	}
	legacy := newMemoryApplicationInteractionSessionRepository()
	prompt := newMemoryApplicationInteractionSessionRepository()
	agent := newMemoryApplicationInteractionSessionRepository()
	sessions := newApplicationInteractionSessionService(
		newCombinedApplicationInteractionSessionRepositoryWithAgent(legacy, prompt, agent), resolver,
	)
	sessions.now = func() time.Time { return time.Date(2026, 7, 25, 15, 31, 0, 0, time.UTC) }
	sessionContext := ApplicationInteractionContext{
		RequestContext: context.Background(), RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: fixture.ctx.AuditRef, WriteEnabled: true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileAgentCopilot},
	})
	if created.Session == nil || created.Session.SchemaVersion != agentCopilotSessionV3Schema {
		t.Fatalf("create Agent Copilot session v3: %#v", created)
	}
	payload, err := json.Marshal(created.Session)
	if err != nil {
		t.Fatalf("marshal Agent Copilot session: %v", err)
	}
	if _, err = decodeAgentCopilotContract(agentCopilotSessionV3Schema, payload); err != nil {
		t.Fatalf("Agent Copilot session is not strict v3: %v payload=%s", err, payload)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil).withAgentCopilot(fixture.service.Invoke)
	artifactService := newApplicationResultArtifactService(newMemoryApplicationResultArtifactRepository(), sessions.repository)
	artifactService.newID = applicationResultArtifactStableIDGenerator()
	coordinator = coordinator.withResultArtifacts(artifactService)
	coordinator.now = func() time.Time { return time.Date(2026, 7, 25, 15, 32, 0, 0, time.UTC) }
	input := validAgentCopilotInvocationInput()
	result := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "agent-turn-1",
		SaveResult: true,
		AgentTask:  input.Task, AgentLocale: input.Locale, AgentConversationID: input.ConversationID,
		AgentArtifacts: input.Artifacts, AgentContext: input.Context,
	})
	if result.FailureCode != "" || result.Session == nil || result.Turn == nil || result.AgentResponse == nil ||
		result.Turn.SchemaVersion != agentCopilotSessionTurnV3Schema ||
		result.Turn.RunRef == nil || result.Turn.RunRef.SchemaVersion != agentCopilotRunV7Schema || result.ResultArtifact == nil || result.ResultArtifactFailureCode != "" ||
		fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent Copilot session did not delegate v7 invocation: %#v calls=%d", result, fixture.bridge.callCount())
	}
	payload, err = json.Marshal(result.Turn)
	if err != nil {
		t.Fatalf("marshal Agent Copilot turn: %v", err)
	}
	if _, err = decodeAgentCopilotContract(agentCopilotSessionTurnV3Schema, payload); err != nil {
		t.Fatalf("Agent Copilot turn is not strict v3: %v payload=%s", err, payload)
	}
	replayed := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "agent-turn-1",
		SaveResult: true,
		AgentTask:  input.Task, AgentLocale: input.Locale, AgentConversationID: input.ConversationID,
		AgentArtifacts: input.Artifacts, AgentContext: input.Context,
	})
	if !replayed.IdempotentReplay || replayed.AgentResponse != nil || replayed.ResultArtifact == nil || replayed.ResultArtifact.ArtifactID != result.ResultArtifact.ArtifactID || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent Copilot session retry replayed provider or response: %#v calls=%d", replayed, fixture.bridge.callCount())
	}
}

func TestAgentCopilotRunFeedsHistoryComparisonEvaluationAndOperationsMetadata(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	firstInput := validAgentCopilotInvocationInput()
	firstInput.ClientInvocationKey = "agent-metadata-a"
	secondInput := validAgentCopilotInvocationInput()
	secondInput.ClientInvocationKey = "agent-metadata-b"
	first := fixture.service.Invoke(fixture.ctx, firstInput)
	second := fixture.service.Invoke(fixture.ctx, secondInput)
	if first.Run == nil || second.Run == nil || first.FailureCode != "" || second.FailureCode != "" {
		t.Fatalf("seed Agent Copilot metadata runs: first=%#v second=%#v", first, second)
	}
	runContext := agentCopilotWorkflowRunContext(fixture.ctx)
	runService := newWorkflowExecutorService(nil, nil, fixture.runStore)
	history := runService.ListRuns(runContext, WorkflowRunListRequest{
		ExecutionSourceKind: agentCopilotExecutionSourceKind,
		ExecutionSourceID:   first.Run.ExecutionSource.ID,
		Limit:               10,
	})
	if history.FailureCode != "" || len(history.Runs) != 2 ||
		history.Runs[0].ExecutionProfile != agentCopilotSuggestionProfile ||
		history.Runs[0].ProfileDigest == "" || history.Runs[0].PolicyDigest == "" ||
		history.Runs[0].AllowedTasksDigest == "" || history.Runs[0].Project != "radishflow" ||
		history.Runs[0].Task != firstInput.Task || history.Runs[0].ResponseDigest == "" ||
		history.Runs[0].SideEffects.ProviderCalls != 1 {
		t.Fatalf("Agent Copilot History/Operations metadata projection drifted: %#v", history)
	}
	comparison := runService.CompareRuns(runContext, first.Run.RunID, second.Run.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil ||
		comparison.Comparison.SchemaVersion != agentCopilotRunComparisonSchemaVersion ||
		comparison.Comparison.RunProfile != agentCopilotSuggestionProfile ||
		comparison.Comparison.Classification != WorkflowRunComparisonUnchanged {
		t.Fatalf("Agent Copilot comparison did not recognize v7 lineage: %#v", comparison)
	}
	evaluationStore := newMemoryWorkflowEvaluationStore(10)
	evaluations := newWorkflowEvaluationService(evaluationStore, fixture.runStore)
	created := evaluations.Create(runContext, WorkflowEvaluationCreateRequest{
		Name: "Agent Copilot metadata regression", BaselineRunID: first.Run.RunID,
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID: second.Run.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged,
		}},
	})
	if created.FailureCode != "" || created.Case == nil {
		t.Fatalf("Agent Copilot evaluation did not accept compatible v7 runs: %#v", created)
	}
	review := evaluations.Review(runContext, created.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.Outcome != "passed" ||
		review.Review.RunProfile != agentCopilotSuggestionProfile {
		t.Fatalf("Agent Copilot evaluation review drifted: %#v", review)
	}
	suites := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluations)
	suites.newSuiteID = func() (string, error) { return "suite_agentmetadata01", nil }
	suites.newDecisionID = func() (string, error) { return "decision_agentmetadata01", nil }
	suite := suites.Create(runContext, WorkflowEvaluationSuiteCreateRequest{
		Name: "Agent Copilot release review",
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{{
			CaseID: created.Case.CaseID, Version: created.Case.Version,
		}},
	})
	if suite.FailureCode != "" || suite.Suite == nil {
		t.Fatalf("Agent Copilot suite creation drifted: %#v", suite)
	}
	suiteReview := suites.Review(runContext, suite.Suite.SuiteID)
	if suiteReview.FailureCode != "" || suiteReview.Review == nil ||
		suiteReview.Review.Outcome != "passed" || len(suiteReview.Review.Items) != 1 ||
		suiteReview.Review.Items[0].RunProfile != agentCopilotSuggestionProfile {
		t.Fatalf("Agent Copilot suite review drifted: %#v", suiteReview)
	}
	decision := suites.Decide(runContext, suite.Suite.SuiteID, WorkflowEvaluationDecisionRequest{
		ExpectedDecisionVersion: 0, Decision: "approved", ReviewDigest: suiteReview.Review.ReviewDigest,
	})
	if decision.FailureCode != "" || decision.Decision == nil || decision.Decision.Decision != "approved" {
		t.Fatalf("Agent Copilot suite decision drifted: %#v", decision)
	}
	incompatibleRun := *second.Run
	incompatibleRun.RunID = "run_agenttaskdrift0001"
	incompatibleRun.AgentTask = "suggest_flowsheet_edits"
	incompatibleRun.RequestID = "request_agent_task_drift"
	incompatibleRun.AuditRef = "audit_agent_task_drift"
	if err := validateWorkflowRunStoreRecord(runContext, &incompatibleRun); err != nil {
		t.Fatalf("incompatible Agent task fixture is invalid: %v", err)
	}
	memoryRunStore, ok := fixture.runStore.(*memoryWorkflowRunStore)
	if !ok {
		t.Fatalf("expected memory run store, got %T", fixture.runStore)
	}
	incompatibleKey := workflowRunStoreKey(runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, incompatibleRun.RunID)
	memoryRunStore.mu.Lock()
	memoryRunStore.records[incompatibleKey] = cloneWorkflowRunRecord(incompatibleRun)
	memoryRunStore.order = append(memoryRunStore.order, incompatibleKey)
	memoryRunStore.mu.Unlock()
	incompatibleCreate := evaluations.Create(runContext, WorkflowEvaluationCreateRequest{
		Name: "Agent task drift", BaselineRunID: first.Run.RunID,
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID: incompatibleRun.RunID, ExpectedClassification: WorkflowRunComparisonChanged,
		}},
	})
	if incompatibleCreate.FailureCode != WorkflowEvaluationFailureAgentCopilotIncompatible {
		t.Fatalf("Agent task drift was accepted: %#v", incompatibleCreate)
	}
	incompatibleCase := *created.Case
	incompatibleCase.CaseID = "eval_agent_task_drift"
	incompatibleCase.Name = "Agent task drift review"
	incompatibleCase.Expectations = []WorkflowEvaluationExpectation{{
		CandidateRunID: incompatibleRun.RunID, ExpectedClassification: WorkflowRunComparisonChanged,
	}}
	if err := evaluationStore.CreateCase(runContext, incompatibleCase); err != nil {
		t.Fatalf("seed incompatible evaluation case: %v", err)
	}
	incompatibleReview := evaluations.Review(runContext, incompatibleCase.CaseID)
	if incompatibleReview.FailureCode != WorkflowEvaluationFailureAgentCopilotIncompatible ||
		!strings.Contains(incompatibleReview.FailureSummary, "profile, project, and task") {
		t.Fatalf("Agent review incompatibility was not mapped: %#v", incompatibleReview)
	}
	payload, err := json.Marshal(first.Run)
	if err != nil {
		t.Fatalf("marshal strict Agent Copilot run: %v", err)
	}
	if _, err = decodeAgentCopilotContract(agentCopilotRunV7Schema, payload); err != nil {
		t.Fatalf("History detail is not strict workflow_run_record.v7: %v payload=%s", err, payload)
	}
	for _, forbidden := range []string{"private selection", "selected_unit_ids", "provider_token", "structured_answer"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("Agent Copilot metadata projection leaked %q: %s", forbidden, payload)
		}
	}
	incompatible := *second.Run
	incompatible.AgentTask = "suggest_flowsheet_edits"
	if err = fixture.runStore.UpsertRun(runContext, &incompatible); !errors.Is(err, errWorkflowRunStoreConflict) {
		t.Fatalf("terminal Agent Copilot run must be immutable, got %v", err)
	}
}

func TestAgentCopilotRunAndSessionPersistAcrossSQLiteRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "agent-copilot-invocation.db")
	runtime := openApplicationInteractionSQLiteRuntime(t, databasePath)
	fixture := newAgentCopilotInvocationFixture(t)
	legacyRuns := newSQLiteWorkflowRunStore(runtime.DB())
	promptRuns, err := newPromptApplicationRunStoreForWorkflowRunStore(legacyRuns)
	if err != nil {
		t.Fatalf("create SQLite Prompt run store: %v", err)
	}
	agentRuns, err := newAgentCopilotRunStoreForWorkflowRunStore(legacyRuns)
	if err != nil {
		t.Fatalf("create SQLite Agent Copilot run store: %v", err)
	}
	fixture.service.runStore = newCombinedWorkflowRunStoreWithAgent(legacyRuns, promptRuns, agentRuns)
	legacySessions := newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	promptSessions, err := newPromptApplicationSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create SQLite Prompt session repository: %v", err)
	}
	agentSessions, err := newAgentCopilotSessionRepositoryForLegacy(legacySessions)
	if err != nil {
		t.Fatalf("create SQLite Agent Copilot session repository: %v", err)
	}
	sessionRepository := newCombinedApplicationInteractionSessionRepositoryWithAgent(legacySessions, promptSessions, agentSessions)
	resolver := newExactApplicationInteractionAuthorityResolver(fixture.catalog, nil, nil, workflowRAGApplicationAuthorityResolver{})
	resolver.resolveAgentCopilot = func(ctx AgentCopilotRuntimeContext) (AgentCopilotRuntimeAuthorityV3, string) {
		authority, failure := fixture.service.resolveAuthority(ctx)
		return authority.Snapshot, failure
	}
	sessions := newApplicationInteractionSessionService(sessionRepository, resolver)
	sessionContext := ApplicationInteractionContext{
		RequestContext: context.Background(), RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: fixture.ctx.AuditRef, WriteEnabled: true,
	}
	created := sessions.Create(sessionContext, ApplicationInteractionSessionCreateInput{
		ProfileBinding: ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileAgentCopilot},
	})
	if created.Session == nil {
		t.Fatalf("create SQLite Agent Copilot session: %#v", created)
	}
	coordinator := newApplicationInteractionTurnCoordinator(sessions, resolver, nil, nil).withAgentCopilot(fixture.service.Invoke)
	input := validAgentCopilotInvocationInput()
	completed := coordinator.Execute(sessionContext, created.Session.SessionID, ApplicationInteractionTurnExecutionInput{
		ExpectedSessionVersion: created.Session.RecordVersion, ClientTurnKey: "sqlite-agent-turn",
		AgentTask: input.Task, AgentLocale: input.Locale, AgentConversationID: input.ConversationID,
		AgentArtifacts: input.Artifacts, AgentContext: input.Context,
	})
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.RunRef == nil {
		t.Fatalf("complete SQLite Agent Copilot turn: %#v", completed)
	}
	runID := completed.Turn.RunRef.RunID
	if err = runtime.Close(); err != nil {
		t.Fatalf("close SQLite Agent Copilot runtime: %v", err)
	}

	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restartedLegacyRuns := newSQLiteWorkflowRunStore(restarted.DB())
	restartedPromptRuns, _ := newPromptApplicationRunStoreForWorkflowRunStore(restartedLegacyRuns)
	restartedAgentRuns, _ := newAgentCopilotRunStoreForWorkflowRunStore(restartedLegacyRuns)
	restartedRuns := newCombinedWorkflowRunStoreWithAgent(restartedLegacyRuns, restartedPromptRuns, restartedAgentRuns)
	restoredRun, found, err := restartedRuns.ReadRun(agentCopilotWorkflowRunContext(fixture.ctx), runID)
	if err != nil || !found || restoredRun.SchemaVersion != agentCopilotRunV7Schema ||
		restoredRun.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restore SQLite Agent Copilot run: found=%v err=%v run=%#v", found, err, restoredRun)
	}
	restartedLegacySessions := newSQLiteApplicationInteractionSessionRepository(restarted.DB())
	restartedPromptSessions, _ := newPromptApplicationSessionRepositoryForLegacy(restartedLegacySessions)
	restartedAgentSessions, _ := newAgentCopilotSessionRepositoryForLegacy(restartedLegacySessions)
	restartedSessionService := newApplicationInteractionSessionService(
		newCombinedApplicationInteractionSessionRepositoryWithAgent(restartedLegacySessions, restartedPromptSessions, restartedAgentSessions), resolver,
	)
	restoredSession := restartedSessionService.Read(sessionContext, created.Session.SessionID)
	restoredTurns, failure := restartedSessionService.ListTurns(sessionContext, created.Session.SessionID)
	if restoredSession.FailureCode != "" || restoredSession.Session == nil ||
		restoredSession.Session.SchemaVersion != agentCopilotSessionV3Schema ||
		failure != "" || len(restoredTurns) != 1 || restoredTurns[0].RunRef == nil ||
		restoredTurns[0].RunRef.RunID != runID {
		t.Fatalf("restore SQLite Agent Copilot session: session=%#v turns=%#v failure=%s", restoredSession, restoredTurns, failure)
	}
	var payloads string
	if err = restarted.DB().QueryRowContext(context.Background(), `SELECT sanitized_session_payload || sanitized_turn_payload
FROM agent_copilot_sessions JOIN agent_copilot_session_turns USING (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
WHERE session_id=?`, created.Session.SessionID).Scan(&payloads); err != nil {
		t.Fatalf("read SQLite Agent Copilot sanitized payloads: %v", err)
	}
	if strings.Contains(payloads, "private selection") || strings.Contains(payloads, "selected_unit_ids") {
		t.Fatalf("SQLite Agent Copilot persistence leaked transient content: %s", payloads)
	}
}

func TestAgentCopilotInvocationHTTPUsesDedicatedAPIKeyScopeAndStrictBody(t *testing.T) {
	fixture := newAgentCopilotInvocationFixture(t)
	apiKeys := newMemoryAPIKeyRepository()
	invokeToken := seedAgentCopilotInvocationAPIKey(t, apiKeys, fixture, "key_agentinvokeaaaaa", []string{agentCopilotInvokeScope})
	chatToken := seedAgentCopilotInvocationAPIKey(t, apiKeys, fixture, "key_agentchatonlyaaa", []string{"chat:invoke"})
	server := &Server{
		config: config.Config{
			AgentCopilotRuntimeDevHTTPEnabled: true, ApplicationCatalogDevHTTPEnabled: true,
			GatewayAuthMode: gatewayAPIKeyAuthenticationSource, Provider: "mock",
		},
		bridge: fixture.bridge, applicationCatalogRepository: fixture.catalog,
		applicationDraftRepository:            fixture.service.authorityResolver.draftRepository,
		applicationPublishCandidateRepository: fixture.service.authorityResolver.publishRepository,
		agentCopilotProfileRepository:         fixture.service.authorityResolver.profileRepository,
		agentCopilotRuntimeRepository:         fixture.runtime, apiKeyRepository: apiKeys,
		workflowRunStore: fixture.runStore, applicationRunStore: fixture.runStore,
	}
	body := `{"task":"explain_diagnostics","locale":"zh-CN","conversation_id":"http-agent-conversation","artifacts":[{"kind":"text","role":"primary","name":"selection.txt","mime_type":"text/plain","content":"private HTTP selection"}],"context":{"selected_unit_ids":["unit-1"]},"client_invocation_key":"http-agent-success"}`
	denied := httptest.NewRequest(http.MethodPost, agentCopilotInvocationRoute, strings.NewReader(body))
	denied.Header.Set("Authorization", "Bearer "+chatToken)
	deniedRecorder := httptest.NewRecorder()
	server.handleAgentCopilotInvocation(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden || !strings.Contains(deniedRecorder.Body.String(), APIKeyFailureScopeDenied) ||
		fixture.bridge.callCount() != 0 {
		t.Fatalf("non-Agent API key crossed invocation scope: status=%d body=%s calls=%d", deniedRecorder.Code, deniedRecorder.Body.String(), fixture.bridge.callCount())
	}
	unknown := httptest.NewRequest(http.MethodPost, agentCopilotInvocationRoute, strings.NewReader(strings.TrimSuffix(body, "}")+`,"model":"forbidden"}`))
	unknown.Header.Set("Authorization", "Bearer "+invokeToken)
	unknownRecorder := httptest.NewRecorder()
	server.handleAgentCopilotInvocation(unknownRecorder, unknown)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), "INVALID_JSON") ||
		fixture.bridge.callCount() != 0 {
		t.Fatalf("Agent Copilot invocation accepted client authority: status=%d body=%s calls=%d", unknownRecorder.Code, unknownRecorder.Body.String(), fixture.bridge.callCount())
	}
	success := httptest.NewRequest(http.MethodPost, agentCopilotInvocationRoute, strings.NewReader(body))
	success.Header.Set("Authorization", "Bearer "+invokeToken)
	successRecorder := httptest.NewRecorder()
	server.handleAgentCopilotInvocation(successRecorder, success)
	var envelope agentCopilotInvocationEnvelope
	if successRecorder.Code != http.StatusOK || json.Unmarshal(successRecorder.Body.Bytes(), &envelope) != nil ||
		envelope.FailureCode != nil || envelope.Run == nil || envelope.Run.SchemaVersion != agentCopilotRunV7Schema ||
		envelope.Response == nil || envelope.ActionSafety == nil || envelope.ActionSafety.Status != actionSafetyReadStatusRecorded ||
		envelope.ActionSafety.Owner.Kind != "agent_copilot_response" ||
		envelope.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
		t.Fatalf("Agent Copilot API key invocation failed: status=%d body=%s calls=%d", successRecorder.Code, successRecorder.Body.String(), fixture.bridge.callCount())
	}
	for _, forbidden := range []string{"private HTTP selection", invokeToken, "selected_unit_ids", "structured_answer"} {
		if strings.Contains(successRecorder.Body.String(), forbidden) {
			t.Fatalf("Agent Copilot invocation response leaked %q: %s", forbidden, successRecorder.Body.String())
		}
	}
}

func seedAgentCopilotInvocationAPIKey(
	t *testing.T,
	repository apiKeyRepository,
	fixture agentCopilotInvocationFixture,
	apiKeyID string,
	scopes []string,
) string {
	t.Helper()
	token, digest, err := newAPIKeyCredential(apiKeyID)
	if err != nil {
		t.Fatalf("create Agent Copilot API key credential: %v", err)
	}
	now := time.Now().UTC()
	ctx := APIKeyContext{
		RequestContext: context.Background(), RequestID: "request_agent_api_key",
		TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		AuditRef: "audit_agent_api_key", WriteEnabled: true,
	}
	record := APIKeyRecord{
		SchemaVersion: apiKeyRecordSchemaVersion, APIKeyID: apiKeyID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, DisplayName: "Agent Copilot invocation key", Scopes: scopes,
		LifecycleState: apiKeyLifecycleActive, EffectiveState: apiKeyLifecycleActive, RecordVersion: 1,
		CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339Nano),
		CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		credentialDigest: digest,
	}
	if _, err = repository.Create(ctx, record); err != nil {
		t.Fatalf("seed Agent Copilot API key: %v", err)
	}
	return token
}

type authorityDriftAgentCopilotRunStore struct {
	workflowRunStore
	drift func()
}

func (store authorityDriftAgentCopilotRunStore) UpsertRun(ctx WorkflowRunContext, record *WorkflowRunRecord) error {
	err := store.workflowRunStore.UpsertRun(ctx, record)
	if err == nil && record.RecordVersion == 1 && store.drift != nil {
		store.drift()
		store.drift = nil
	}
	return err
}

func newAgentCopilotInvocationFixture(t *testing.T) agentCopilotInvocationFixture {
	t.Helper()
	base := newAgentCopilotBatchCFixture(t)
	activated := base.runtimeService.Decide(base.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: base.firstCandidate.CandidateID,
	})
	if activated.Assignment == nil {
		t.Fatalf("activate Agent Copilot assignment: %#v", activated)
	}
	catalog := newMemoryApplicationCatalogRepository()
	catalogContext := ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_agent_invocation_catalog",
		TenantRef: base.runtimeContext.TenantRef, WorkspaceID: base.runtimeContext.WorkspaceID,
		ActorRef: base.runtimeContext.ActorRef, OwnerSubjectRef: base.runtimeContext.OwnerSubjectRef,
		AuditRef: "audit_agent_invocation_catalog", WriteEnabled: true,
	}
	record := ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: base.runtimeContext.ApplicationID,
		TenantRef: catalogContext.TenantRef, WorkspaceID: catalogContext.WorkspaceID,
		OwnerSubjectRef: catalogContext.OwnerSubjectRef, DisplayName: "Agent Copilot",
		Description: "Controlled Agent Copilot.", ApplicationKind: "agent",
		LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1,
		CreatedAt: "2026-07-25T12:00:00Z", UpdatedAt: "2026-07-25T12:00:00Z",
		CreatedByActorRef: catalogContext.ActorRef, UpdatedByActorRef: catalogContext.ActorRef,
		RequestID: catalogContext.RequestID, AuditRef: catalogContext.AuditRef,
	}
	if _, err := catalog.Create(catalogContext, record); err != nil {
		t.Fatalf("seed Agent application catalog: %v", err)
	}
	resolver := base.runtimeService.resolver
	runStore := newMemoryWorkflowRunStore(20)
	testBridge := &workflowExecutorTestBridge{handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return successfulAgentCopilotGatewayEnvelope(), nil
	}}
	service := newAgentCopilotInvocationService(base.runtimeRepository, resolver, catalog, runStore, testBridge)
	service.resolveSelection = func(context.Context, string) northboundSelection {
		return northboundSelection{
			provider: "mock", providerProfile: "default", model: "profile:local-dev",
			upstreamModel: "profile:local-dev", source: "test_selection",
		}
	}
	service.now = func() time.Time { return time.Date(2026, 7, 25, 15, 30, 0, 0, time.UTC) }
	return agentCopilotInvocationFixture{
		service: service, ctx: base.runtimeContext, runStore: runStore,
		runtime: base.runtimeRepository, bridge: testBridge, catalog: catalog, batchC: base,
	}
}

func validAgentCopilotInvocationInput() AgentCopilotInvocationInput {
	return AgentCopilotInvocationInput{
		Task: "explain_diagnostics", Locale: "zh-CN", ConversationID: "conversation-agent-1",
		Artifacts: []AgentCopilotArtifact{{
			Kind: "text", Role: "primary", Name: "selection.txt", MIMEType: "text/plain", Content: "private selection",
		}},
		Context:             map[string]any{"selected_unit_ids": []any{"unit-1"}},
		ClientInvocationKey: "client-agent-invocation-1",
	}
}

func successfulAgentCopilotGatewayEnvelope() bridge.GatewayEnvelope {
	return agentCopilotGatewayEnvelope(map[string]any{
		"schema_version": 1, "status": "ok", "project": "radishflow", "task": "explain_diagnostics",
		"summary": "Selection summarized.", "answers": []any{map[string]any{"text": "One selected unit."}},
		"issues": []any{}, "proposed_actions": []any{}, "citations": []any{},
		"confidence": 0.9, "risk_level": "low", "requires_confirmation": false,
	})
}

func agentCopilotGatewayEnvelope(response map[string]any) bridge.GatewayEnvelope {
	payload, _ := json.Marshal(response)
	var canonical map[string]any
	_ = json.Unmarshal(payload, &canonical)
	status, _ := canonical["status"].(string)
	return bridge.GatewayEnvelope{Status: status, Response: canonical}
}
