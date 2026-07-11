package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type workflowExecutorTestBridge struct {
	mu           sync.Mutex
	handleCalls  int
	lastRequests [][]byte
	handle       func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error)
}

func (testBridge *workflowExecutorTestBridge) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return []bridge.ProviderDescription{{ProviderID: "mock"}}, nil
}

func (testBridge *workflowExecutorTestBridge) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	return bridge.ProviderInventory{
		Providers: []bridge.ProviderDescription{{ProviderID: "mock"}},
	}, nil
}

func (testBridge *workflowExecutorTestBridge) HandleEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options bridge.EnvelopeOptions,
) (bridge.GatewayEnvelope, error) {
	testBridge.mu.Lock()
	testBridge.handleCalls++
	testBridge.lastRequests = append(testBridge.lastRequests, append([]byte(nil), canonicalRequest...))
	handle := testBridge.handle
	testBridge.mu.Unlock()
	if handle != nil {
		return handle(ctx, canonicalRequest, options)
	}
	return successfulWorkflowExecutorEnvelope("reviewable workflow answer"), nil
}

func (testBridge *workflowExecutorTestBridge) StreamEnvelope(
	context.Context,
	[]byte,
	bridge.EnvelopeOptions,
	func(bridge.StreamEvent) error,
) error {
	return errors.New("workflow executor v0 does not use streaming")
}

func (testBridge *workflowExecutorTestBridge) callCount() int {
	testBridge.mu.Lock()
	defer testBridge.mu.Unlock()
	return testBridge.handleCalls
}

func (testBridge *workflowExecutorTestBridge) lastRequest() []byte {
	testBridge.mu.Lock()
	defer testBridge.mu.Unlock()
	if len(testBridge.lastRequests) == 0 {
		return nil
	}
	return append([]byte(nil), testBridge.lastRequests[len(testBridge.lastRequests)-1]...)
}

func TestWorkflowExecutorRunsSavedLinearDraftAndStoresBoundedRecord(t *testing.T) {
	draft := executableWorkflowDraftForTest()
	testBridge := &workflowExecutorTestBridge{}
	store := newMemoryWorkflowRunStore(10)
	service := workflowExecutorTestService(draft, testBridge, store)
	runContext := workflowExecutorTestContext()
	rawInput := "sensitive operator input must not enter the run record"

	result := service.StartRun(runContext, WorkflowRunRequest{
		DraftID:   draft.DraftID,
		InputText: rawInput,
		Model:     "mock",
	})

	if result.FailureCode != "" || result.Record == nil {
		t.Fatalf("expected successful workflow run: %#v", result)
	}
	record := *result.Record
	if record.Status != WorkflowRunStatusSucceeded || record.Output != "reviewable workflow answer" {
		t.Fatalf("unexpected terminal workflow run record: %#v", record)
	}
	if record.DraftVersion != draft.DraftVersion || record.SchemaVersion != workflowRunRecordSchemaVersion {
		t.Fatalf("run record lost saved draft version evidence: %#v", record)
	}
	if record.SideEffects.ProviderCalls != 1 || record.SideEffects.ToolCalls != 0 ||
		record.SideEffects.ConfirmationCalls != 0 || record.SideEffects.BusinessWrites != 0 ||
		record.SideEffects.ReplayWrites != 0 {
		t.Fatalf("unexpected workflow run side effects: %#v", record.SideEffects)
	}
	if testBridge.callCount() != 1 {
		t.Fatalf("expected one Gateway call, got %d", testBridge.callCount())
	}
	serialized, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal workflow run record: %v", err)
	}
	if strings.Contains(string(serialized), rawInput) {
		t.Fatalf("run record must not retain raw input: %s", serialized)
	}
	promptRecord := workflowRunNodeRecordForTest(t, record, "node_context")
	if promptRecord.OutputPreview != "input packet accepted; raw input not retained" {
		t.Fatalf("prompt record should retain only bounded evidence: %#v", promptRecord)
	}

	stored := service.ReadRun(runContext, record.RunID)
	if stored.FailureCode != "" || stored.Record == nil || stored.Record.Output != record.Output {
		t.Fatalf("stored run record should be scoped and readable: %#v", stored)
	}
	otherScope := runContext
	otherScope.WorkspaceID = "workspace_other"
	if scoped := service.ReadRun(otherScope, record.RunID); scoped.FailureCode != WorkflowRunFailureRecordNotFound {
		t.Fatalf("cross-workspace read must fail closed: %#v", scoped)
	}
	otherTenant := runContext
	otherTenant.TenantRef = "tenant_other"
	if scoped := service.ReadRun(otherTenant, record.RunID); scoped.FailureCode != WorkflowRunFailureRecordNotFound {
		t.Fatalf("cross-tenant read must fail closed: %#v", scoped)
	}

	canonicalRequest := decodeCanonicalRequest(t, testBridge.lastRequest())
	toolHints, ok := canonicalRequest["tool_hints"].(map[string]any)
	if !ok || toolHints["allow_tool_calls"] != false || toolHints["allow_retrieval"] != false {
		t.Fatalf("workflow executor Gateway request must keep tool and retrieval disabled: %#v", canonicalRequest)
	}
	safety, ok := canonicalRequest["safety"].(map[string]any)
	if !ok || safety["mode"] != "advisory" {
		t.Fatalf("workflow executor Gateway request must remain advisory: %#v", canonicalRequest)
	}
}

func TestWorkflowExecutorConditionBranchRecordsSkippedNode(t *testing.T) {
	draft := executableConditionalWorkflowDraftForTest()
	testBridge := &workflowExecutorTestBridge{}
	service := workflowExecutorTestService(draft, testBridge, newMemoryWorkflowRunStore(10))

	result := service.StartRun(workflowExecutorTestContext(), WorkflowRunRequest{
		DraftID:         draft.DraftID,
		InputText:       "Evaluate the explicitly selected false branch.",
		ConditionValues: map[string]bool{"node_condition": false},
	})

	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("expected successful conditional workflow run: %#v", result)
	}
	if testBridge.callCount() != 1 {
		t.Fatalf("only the active LLM branch should call Gateway, got %d", testBridge.callCount())
	}
	if got := workflowRunNodeRecordForTest(t, *result.Record, "node_model_true").Status; got != WorkflowRunNodeStatusSkipped {
		t.Fatalf("true branch should be skipped, got %s", got)
	}
	if got := workflowRunNodeRecordForTest(t, *result.Record, "node_model_false").Status; got != WorkflowRunNodeStatusSucceeded {
		t.Fatalf("false branch should succeed, got %s", got)
	}
	conditionRecord := workflowRunNodeRecordForTest(t, *result.Record, "node_condition")
	if strings.Contains(conditionRecord.OutputPreview, "false") {
		t.Fatalf("condition value must not be retained in run record: %#v", conditionRecord)
	}
}

func TestWorkflowExecutorRejectsUnsafeOrInvalidDraftWithoutGatewaySideEffects(t *testing.T) {
	testCases := []struct {
		name            string
		mutate          func(*SavedWorkflowDraft)
		expectedFailure WorkflowRunFailureCode
	}{
		{
			name: "tool ref",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Nodes[1].ToolRef = "tool:blocked"
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
		{
			name: "rag ref",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Nodes[1].RAGRef = "rag:blocked"
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
		{
			name: "confirmation required",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Nodes[1].RequiresConfirmation = true
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
		{
			name: "medium risk",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Nodes[1].RiskLevel = "medium"
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
		{
			name: "requested capability",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.RequestedCapabilities = []string{"tool_executor"}
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
		{
			name: "cycle",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Edges = append(draft.Edges, SavedWorkflowDraftEdge{
					EdgeID:     "edge_output_context",
					FromNodeID: "node_output",
					ToNodeID:   "node_context",
				})
			},
			expectedFailure: WorkflowRunFailureGraphInvalid,
		},
		{
			name: "disallowed node type",
			mutate: func(draft *SavedWorkflowDraft) {
				draft.Nodes[1].NodeType = "http_tool"
			},
			expectedFailure: WorkflowRunFailureDraftNotEligible,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			draft := executableWorkflowDraftForTest()
			testCase.mutate(&draft)
			testBridge := &workflowExecutorTestBridge{}
			service := workflowExecutorTestService(draft, testBridge, newMemoryWorkflowRunStore(10))

			result := service.StartRun(workflowExecutorTestContext(), WorkflowRunRequest{
				DraftID:   draft.DraftID,
				InputText: "safe input",
			})

			if result.FailureCode != testCase.expectedFailure || result.Record != nil {
				t.Fatalf("unexpected unsafe draft result: %#v", result)
			}
			if testBridge.callCount() != 0 {
				t.Fatalf("rejected draft must not call Gateway, got %d", testBridge.callCount())
			}
		})
	}
}

func TestWorkflowExecutorInputBudgetRejectsBeforeDraftOrGatewayExecution(t *testing.T) {
	draft := executableWorkflowDraftForTest()
	testBridge := &workflowExecutorTestBridge{}
	draftReads := 0
	service := workflowExecutorTestService(draft, testBridge, newMemoryWorkflowRunStore(10))
	service.draftReader = func(context SavedWorkflowDraftContext, request ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
		draftReads++
		return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(draft)}
	}

	result := service.StartRun(workflowExecutorTestContext(), WorkflowRunRequest{
		DraftID:   draft.DraftID,
		InputText: strings.Repeat("x", workflowExecutorMaxInputBytes+1),
	})

	if result.FailureCode != WorkflowRunFailureBudgetExceeded || result.Record != nil {
		t.Fatalf("oversized input should fail before execution: %#v", result)
	}
	if draftReads != 0 || testBridge.callCount() != 0 {
		t.Fatalf("oversized input must not read draft or call Gateway: draft_reads=%d gateway_calls=%d", draftReads, testBridge.callCount())
	}
}

func TestWorkflowExecutorGatewayFailureAndTimeoutProduceTerminalRecords(t *testing.T) {
	t.Run("gateway failure", func(t *testing.T) {
		draft := executableWorkflowDraftForTest()
		testBridge := &workflowExecutorTestBridge{
			handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
				return bridge.GatewayEnvelope{
					Status: "failed",
					Error:  &bridge.GatewayError{Code: "provider_failed", Message: "raw provider failure"},
				}, nil
			},
		}
		store := newMemoryWorkflowRunStore(10)
		service := workflowExecutorTestService(draft, testBridge, store)

		result := service.StartRun(workflowExecutorTestContext(), WorkflowRunRequest{
			DraftID:   draft.DraftID,
			InputText: "safe input",
		})

		if result.FailureCode != WorkflowRunFailureGatewayFailed || result.Record == nil ||
			result.Record.Status != WorkflowRunStatusFailed || result.Record.SideEffects.ProviderCalls != 1 {
			t.Fatalf("unexpected Gateway failure record: %#v", result)
		}
		serialized, _ := json.Marshal(result.Record)
		if strings.Contains(string(serialized), "raw provider failure") {
			t.Fatalf("raw provider error must not enter run record: %s", serialized)
		}
	})

	t.Run("execution timeout", func(t *testing.T) {
		draft := executableWorkflowDraftForTest()
		testBridge := &workflowExecutorTestBridge{
			handle: func(ctx context.Context, _ []byte, _ bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
				<-ctx.Done()
				return bridge.GatewayEnvelope{}, ctx.Err()
			},
		}
		service := workflowExecutorTestService(draft, testBridge, newMemoryWorkflowRunStore(10))
		service.maxRuntime = 5 * time.Millisecond

		result := service.StartRun(workflowExecutorTestContext(), WorkflowRunRequest{
			DraftID:   draft.DraftID,
			InputText: "safe input",
		})

		if result.FailureCode != WorkflowRunFailureCanceled || result.Record == nil ||
			result.Record.Status != WorkflowRunStatusCanceled {
			t.Fatalf("unexpected timeout record: %#v", result)
		}
	})
}

func TestMemoryWorkflowRunStoreEvictsOldestAndClonesRecords(t *testing.T) {
	store := newMemoryWorkflowRunStore(2)
	runContext := workflowExecutorTestContext()
	for _, runID := range []string{"run_1", "run_2", "run_3"} {
		if err := store.UpsertRun(runContext, WorkflowRunRecord{
			RunID:            runID,
			WorkspaceID:      runContext.WorkspaceID,
			ApplicationID:    runContext.ApplicationID,
			ConditionNodeIDs: []string{"node_condition"},
		}); err != nil {
			t.Fatalf("upsert run record: %v", err)
		}
	}
	if _, found, _ := store.ReadRun(runContext, "run_1"); found {
		t.Fatalf("oldest run record should be evicted")
	}
	record, found, err := store.ReadRun(runContext, "run_3")
	if err != nil || !found {
		t.Fatalf("newest run record should remain: found=%v err=%v", found, err)
	}
	record.ConditionNodeIDs[0] = "mutated"
	reloaded, _, _ := store.ReadRun(runContext, "run_3")
	if reloaded.ConditionNodeIDs[0] != "node_condition" {
		t.Fatalf("run store must return cloned records: %#v", reloaded)
	}
}

func workflowExecutorTestService(
	draft SavedWorkflowDraft,
	testBridge bridgeClient,
	store workflowRunStore,
) workflowExecutorService {
	service := newWorkflowExecutorService(
		func(context SavedWorkflowDraftContext, request ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
			if request.DraftID != draft.DraftID {
				return savedWorkflowDraftFailure(SavedWorkflowDraftFailureNotFound, savedWorkflowDraftAuditMetadata(context))
			}
			return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(draft)}
		},
		testBridge,
		store,
	)
	service.resolveSelection = func(context.Context, string) northboundSelection {
		return northboundSelection{
			provider:      "mock",
			model:         "mock",
			upstreamModel: "mock",
			source:        "workflow_executor_test",
		}
	}
	service.envelopeOptions = func(selection northboundSelection, temperature float64) bridge.EnvelopeOptions {
		return bridge.EnvelopeOptions{
			Provider:       selection.provider,
			Model:          selection.upstreamModel,
			Temperature:    temperature,
			RequestTimeout: time.Second,
		}
	}
	service.newRunID = func() (string, error) { return "run_workflow_executor_test", nil }
	return service
}

func executableWorkflowDraftForTest() SavedWorkflowDraft {
	draft := savedWorkflowDraftFromPayload(validSavedWorkflowDraftPayload())
	draft.DraftVersion = 1
	draft.DraftStatus = SavedWorkflowDraftStatusValidForReview
	draft.ValidationSummary = SavedWorkflowDraftValidationSummary{
		ValidationState: SavedWorkflowDraftStatusValidForReview,
		ValidForReview:  true,
		Findings:        []SavedWorkflowDraftValidationFinding{},
	}
	return draft
}

func executableConditionalWorkflowDraftForTest() SavedWorkflowDraft {
	base := executableWorkflowDraftForTest()
	promptNode := base.Nodes[0]
	trueModelNode := base.Nodes[1]
	trueModelNode.NodeID = "node_model_true"
	trueModelNode.Label = "True branch model"
	falseModelNode := trueModelNode
	falseModelNode.NodeID = "node_model_false"
	falseModelNode.Label = "False branch model"
	outputNode := base.Nodes[2]
	conditionNode := SavedWorkflowDraftNode{
		NodeID:        "node_condition",
		NodeType:      "condition",
		Label:         "Explicit branch condition",
		InputSummary:  "Use the explicit run condition value.",
		OutputSummary: "Select one bounded advisory branch.",
		RiskLevel:     "low",
	}
	base.Nodes = []SavedWorkflowDraftNode{promptNode, conditionNode, trueModelNode, falseModelNode, outputNode}
	base.Edges = []SavedWorkflowDraftEdge{
		{EdgeID: "edge_prompt_condition", FromNodeID: promptNode.NodeID, ToNodeID: conditionNode.NodeID},
		{EdgeID: "edge_condition_true", FromNodeID: conditionNode.NodeID, ToNodeID: trueModelNode.NodeID, ConditionSummary: "when:true"},
		{EdgeID: "edge_condition_false", FromNodeID: conditionNode.NodeID, ToNodeID: falseModelNode.NodeID, ConditionSummary: "when:false"},
		{EdgeID: "edge_true_output", FromNodeID: trueModelNode.NodeID, ToNodeID: outputNode.NodeID},
		{EdgeID: "edge_false_output", FromNodeID: falseModelNode.NodeID, ToNodeID: outputNode.NodeID},
	}
	return base
}

func workflowExecutorTestContext() WorkflowRunContext {
	return WorkflowRunContext{
		RequestContext: context.Background(),
		RequestID:      "req_workflow_executor_test",
		TenantRef:      "tenant_demo",
		WorkspaceID:    "workspace_demo",
		ApplicationID:  "app_flow_copilot",
		ActorRef:       "subject_demo_user",
		ScopeGrants:    []string{"workflow_drafts:read", "workflow_runs:execute", "workflow_runs:read"},
		AuditRef:       "audit_workflow_executor_test",
	}
}

func successfulWorkflowExecutorEnvelope(summary string) bridge.GatewayEnvelope {
	return bridge.GatewayEnvelope{
		SchemaVersion: 1,
		Status:        "ok",
		RequestID:     "req_workflow_executor_gateway",
		Project:       "radish",
		Task:          "answer_docs_question",
		Response:      map[string]any{"summary": summary},
		Metadata:      map[string]any{},
	}
}

func workflowRunNodeRecordForTest(
	t *testing.T,
	record WorkflowRunRecord,
	nodeID string,
) WorkflowRunNodeRecord {
	t.Helper()
	for _, nodeRecord := range record.Nodes {
		if nodeRecord.NodeID == nodeID {
			return nodeRecord
		}
	}
	t.Fatalf("workflow run record missing node %s: %#v", nodeID, record.Nodes)
	return WorkflowRunNodeRecord{}
}
