package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

func TestWorkflowRAGExecutionSucceedsExactlyOnceAndStoresMetadataOnly(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	query := "official retrieval guidance query-secret-needle"
	fixture.bridge.envelope.Response["provider_debug"] = "model-raw-secret-needle"

	result := fixture.service.Execute(fixture.runContext, WorkflowRAGExecutionRequest{
		DraftID: fixture.draft.DraftID, DraftVersion: fixture.draft.DraftVersion, InputText: query, Model: "mock-rag",
	})
	if result.FailureCode != "" || result.Record == nil || result.RetrievalAnswer == nil {
		t.Fatalf("retrieval execution failed: %#v", result)
	}
	record := *result.Record
	if record.SchemaVersion != workflowRunRecordRAGSchemaVersion || record.Status != WorkflowRunStatusSucceeded ||
		record.SideEffects.RetrievalCalls != 1 || record.SideEffects.ProviderCalls != 1 || fixture.bridge.handleCalls != 1 ||
		record.RetrievalAttempt == nil || len(record.RetrievalAttempt.SelectedFragments) == 0 ||
		len(record.RetrievalAttempt.CitationRefs) != 1 || record.RAGAnswer != nil {
		t.Fatalf("unexpected successful retrieval run: %#v calls=%d", record, fixture.bridge.handleCalls)
	}
	if result.RetrievalAnswer.Answer != "Use the official guidance." || result.RetrievalAnswer.Citations[0].FragmentRef != "official_guide" {
		t.Fatalf("unexpected response-only structured answer: %#v", result.RetrievalAnswer)
	}
	if !strings.Contains(string(fixture.bridge.lastRequest), `"allow_retrieval":false`) ||
		!strings.Contains(string(fixture.bridge.lastRequest), query) {
		t.Fatalf("Gateway packet did not preserve the frozen retrieval boundary: %s", fixture.bridge.lastRequest)
	}
	stored, found, err := fixture.runStore.ReadRun(fixture.runContext, record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusSucceeded || stored.RecordVersion < 4 {
		t.Fatalf("successful v3 run was not durably checkpointed: %#v found=%t err=%v", stored, found, err)
	}
	payload, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("marshal stored v3 run: %v", err)
	}
	for _, forbidden := range []string{query, "query-secret-needle", "official retrieval guidance", "model-raw-secret-needle", "用户问题", "credential"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("metadata-only v3 run leaked %q: %s", forbidden, payload)
		}
	}
	entry := workflowRAGExecutionMemoryEntry(t, fixture.snapshotRepository, fixture.snapshot.SnapshotID, fixture.snapshotContext)
	if len(entry.audits) != 3 || entry.audits[1].EventKind != "retrieval_started" || entry.audits[2].EventKind != "retrieval_succeeded" {
		t.Fatalf("unexpected retrieval execution audit trail: %#v", entry.audits)
	}
	auditPayload, _ := json.Marshal(entry.audits)
	if strings.Contains(string(auditPayload), query) || strings.Contains(string(auditPayload), fixture.snapshot.Fragments[0].Content) {
		t.Fatalf("execution audit leaked query or fragment content: %s", auditPayload)
	}
}

func TestWorkflowRAGExecutionFailsClosedWithoutRetryOrFallback(t *testing.T) {
	tests := []struct {
		name            string
		prepare         func(*workflowRAGExecutionFixture)
		request         func(*workflowRAGExecutionFixture) WorkflowRAGExecutionRequest
		expectedFailure string
		expectedCalls   int
	}{
		{
			name: "exact draft version",
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion+1)
			},
			expectedFailure: WorkflowRAGFailureDraftIneligible,
		},
		{
			name: "illegal topology",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.draft.Edges[0].ConditionSummary = "condition is forbidden"
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureDraftIneligible,
		},
		{
			name: "empty query",
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "   ", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureQueryInvalid,
		},
		{
			name: "no evidence",
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "unmatched-token", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureNoEvidence,
		},
		{
			name: "retrieval budget",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.service.rank = func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult {
					return WorkflowRAGRankingResult{Profile: workflowRAGLexicalProfile(), QueryDigest: workflowRAGSHA256("query"), QueryBytes: 5, FailureCode: WorkflowRAGFailureBudgetExceeded}
				}
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureBudgetExceeded,
		},
		{
			name: "profile digest drift",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.service.profile = func() WorkflowRAGExecutionProfile {
					profile := workflowRAGLexicalProfile()
					profile.ProfileDigest = workflowRAGSHA256("drift")
					return profile
				}
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureProfileDisabled,
		},
		{
			name: "snapshot digest drift",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.snapshotRepository.ownerLock.Lock()
				key := workflowRAGStoreKey(f.snapshotContext, f.snapshot.SnapshotID)
				entry := f.snapshotRepository.entries[key]
				version := entry.versions[f.snapshot.SnapshotVersion]
				version.SnapshotDigest = workflowRAGSHA256("tampered snapshot")
				entry.versions[f.snapshot.SnapshotVersion] = version
				f.snapshotRepository.entries[key] = entry
				f.snapshotRepository.ownerLock.Unlock()
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureStoreUnavailable,
		},
		{
			name: "exact snapshot version",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.draft.RAGRefs[0] = strings.TrimSuffix(f.draft.RAGRefs[0], ".v1") + ".v2"
				for index := range f.draft.Nodes {
					if f.draft.Nodes[index].NodeType == "rag_retrieval" {
						f.draft.Nodes[index].RAGRef = f.draft.RAGRefs[0]
					}
				}
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureNotFound,
		},
		{
			name: "snapshot three-level scope",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.runContext.ApplicationID = "application_other"
				f.service.draftReader = func(SavedWorkflowDraftContext, ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
					return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(f.draft)}
				}
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureScopeDenied,
		},
		{
			name: "archived snapshot",
			prepare: func(f *workflowRAGExecutionFixture) {
				result := newWorkflowRAGSnapshotService(f.snapshotRepository).Archive(f.snapshotContext, f.snapshot.SnapshotID, 1)
				if result.FailureCode != "" {
					t.Fatalf("archive fixture snapshot: %#v", result)
				}
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureArchived,
		},
		{
			name:    "Gateway failure is not retried",
			prepare: func(f *workflowRAGExecutionFixture) { f.bridge.handleErr = errors.New("provider unavailable") },
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureGatewayFailed,
			expectedCalls:   1,
		},
		{
			name: "unselected citation",
			prepare: func(f *workflowRAGExecutionFixture) {
				f.bridge.envelope.Response["structured_answer"] = workflowRAGTestAnswer("internal_notes", "Unknown citation")
			},
			request: func(f *workflowRAGExecutionFixture) WorkflowRAGExecutionRequest {
				return workflowRAGExecutionFixtureRequest(f, "official retrieval guidance", f.draft.DraftVersion)
			},
			expectedFailure: WorkflowRAGFailureCitationInvalid,
			expectedCalls:   1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkflowRAGExecutionFixture(t)
			if test.prepare != nil {
				test.prepare(fixture)
			}
			result := fixture.service.Execute(fixture.runContext, test.request(fixture))
			if string(result.FailureCode) != test.expectedFailure || fixture.bridge.handleCalls != test.expectedCalls {
				t.Fatalf("unexpected fail-closed result: %#v calls=%d", result, fixture.bridge.handleCalls)
			}
			if result.Record != nil && (result.Record.Status != WorkflowRunStatusFailed || result.Record.RAGAnswer != nil) {
				t.Fatalf("failed run did not persist a metadata-only terminal state: %#v", result.Record)
			}
		})
	}
}

func TestWorkflowRAGAnswerValidatorAcceptsOnlySelectedUniqueCitations(t *testing.T) {
	selected := []WorkflowRAGRankedFragment{{FragmentRef: "official_guide"}, {FragmentRef: "internal_notes"}}
	valid, failure := parseWorkflowRAGAnswer(mustWorkflowRAGAnswerJSON(t, workflowRAGTestAnswer("official_guide", "Supported claim")), selected)
	if failure != "" || valid.Confidence != "medium" {
		t.Fatalf("valid workflow RAG answer was rejected: %#v failure=%s", valid, failure)
	}
	tests := []struct {
		name    string
		payload string
		failure string
	}{
		{name: "malformed", payload: `{`, failure: WorkflowRAGFailureAnswerInvalid},
		{name: "empty answer", payload: mustWorkflowRAGAnswerJSON(t, WorkflowRAGAnswer{SchemaVersion: "workflow_rag_answer.v1", Citations: []WorkflowRAGCitation{{FragmentRef: "official_guide", ClaimSummary: "claim"}}, Limitations: []string{}, Confidence: "medium"}), failure: WorkflowRAGFailureAnswerInvalid},
		{name: "unknown citation", payload: mustWorkflowRAGAnswerJSON(t, workflowRAGTestAnswer("other_fragment", "claim")), failure: WorkflowRAGFailureCitationInvalid},
		{name: "duplicate citation", payload: mustWorkflowRAGAnswerJSON(t, WorkflowRAGAnswer{SchemaVersion: "workflow_rag_answer.v1", Answer: "answer", Citations: []WorkflowRAGCitation{{FragmentRef: "official_guide", ClaimSummary: "one"}, {FragmentRef: "official_guide", ClaimSummary: "two"}}, Limitations: []string{}, Confidence: "medium"}), failure: WorkflowRAGFailureCitationInvalid},
		{name: "invalid confidence", payload: mustWorkflowRAGAnswerJSON(t, WorkflowRAGAnswer{SchemaVersion: "workflow_rag_answer.v1", Answer: "answer", Citations: []WorkflowRAGCitation{{FragmentRef: "official_guide", ClaimSummary: "claim"}}, Limitations: []string{}, Confidence: "certain"}), failure: WorkflowRAGFailureAnswerInvalid},
		{name: "empty limitation", payload: mustWorkflowRAGAnswerJSON(t, WorkflowRAGAnswer{SchemaVersion: "workflow_rag_answer.v1", Answer: "answer", Citations: []WorkflowRAGCitation{{FragmentRef: "official_guide", ClaimSummary: "claim"}}, Limitations: []string{" "}, Confidence: "low"}), failure: WorkflowRAGFailureAnswerInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, failure := parseWorkflowRAGAnswer(test.payload, selected); failure != test.failure {
				t.Fatalf("expected %s, got %s", test.failure, failure)
			}
		})
	}
}

func TestWorkflowRAGReconciliationClosesStaleRunningWithoutExecution(t *testing.T) {
	fixture := newWorkflowRAGExecutionFixture(t)
	plan, failure := buildWorkflowRAGExecutionPlan(fixture.draft, fixture.draft.DraftVersion)
	if failure != "" {
		t.Fatalf("build reconciliation plan: %s", failure)
	}
	digest, _ := workflowRAGDraftDigest(fixture.draft)
	run := newWorkflowRAGRunRecord(
		fixture.runContext,
		workflowRAGExecutionFixtureRequest(fixture, "official retrieval guidance", fixture.draft.DraftVersion),
		fixture.draft, digest, plan, fixture.snapshot, workflowRAGLexicalProfile(), workflowRAGTestSelection(),
		"run_reconcile0000001", time.Now().UTC().Add(-workflowExecutorDefaultMaxRuntime-time.Second),
	)
	if err := fixture.runStore.UpsertRun(fixture.runContext, &run); err != nil {
		t.Fatalf("seed stale v3 run: %v", err)
	}
	rankCalls := 0
	fixture.service.rank = func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult {
		rankCalls++
		return WorkflowRAGRankingResult{}
	}
	result := fixture.service.ReconcileStale(fixture.runContext)
	if result.FailureCode != "" || result.Reconciled != 1 || rankCalls != 0 || fixture.bridge.handleCalls != 0 {
		t.Fatalf("stale reconciliation executed work or failed: %#v rank=%d gateway=%d", result, rankCalls, fixture.bridge.handleCalls)
	}
	stored, found, err := fixture.runStore.ReadRun(fixture.runContext, run.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusFailed ||
		stored.FailureCode != WorkflowRunFailureCode(WorkflowRAGFailureInterrupted) ||
		stored.SideEffects.RetrievalCalls != 0 || stored.SideEffects.ProviderCalls != 0 {
		t.Fatalf("stale run did not converge to the real interrupted terminal state: %#v found=%t err=%v", stored, found, err)
	}
	entry := workflowRAGExecutionMemoryEntry(t, fixture.snapshotRepository, fixture.snapshot.SnapshotID, fixture.snapshotContext)
	lastAudit := entry.audits[len(entry.audits)-1]
	if lastAudit.EventKind != "retrieval_failed" || lastAudit.ActorRef != workflowRAGReconcilerActorRef {
		t.Fatalf("reconciliation audit is incomplete: %#v", lastAudit)
	}
	repeated := fixture.service.ReconcileStale(fixture.runContext)
	if repeated.FailureCode != "" || repeated.Reconciled != 0 {
		t.Fatalf("terminal v3 run reconciled more than once: %#v", repeated)
	}
}

type workflowRAGExecutionFixture struct {
	service            workflowRAGExecutionService
	bridge             *fakeBridge
	runStore           *memoryWorkflowRunStore
	snapshotRepository *memoryWorkflowRAGSnapshotRepository
	snapshotContext    WorkflowRAGSnapshotContext
	runContext         WorkflowRunContext
	snapshot           WorkflowRAGSnapshotRecord
	draft              SavedWorkflowDraft
}

func newWorkflowRAGExecutionFixture(t *testing.T) *workflowRAGExecutionFixture {
	t.Helper()
	runStore := newMemoryWorkflowRunStore(20)
	snapshotRepository := newMemoryWorkflowRAGSnapshotRepository(nil)
	snapshotContext := workflowRAGTestContext()
	snapshotService := newWorkflowRAGSnapshotService(snapshotRepository)
	snapshot := snapshotService.Create(snapshotContext, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "execution_manual", DisplayName: "Execution manual", ContentClassification: "workspace_internal",
		Fragments: workflowRAGTestFragments(),
	})
	if snapshot.FailureCode != "" || snapshot.Record == nil {
		t.Fatalf("create execution snapshot: %#v", snapshot)
	}
	draft := workflowRAGEligibleDraft(snapshot.Record.RAGRef)
	reader := func(ctx SavedWorkflowDraftContext, request ReadWorkflowDraftRequest) SavedWorkflowDraftResult {
		if request.DraftID != draft.DraftID || ctx.TenantRef != snapshotContext.TenantRef ||
			ctx.WorkspaceID != draft.WorkspaceID || ctx.ApplicationID != draft.ApplicationID {
			return SavedWorkflowDraftResult{FailureCode: SavedWorkflowDraftFailureScopeDenied}
		}
		return SavedWorkflowDraftResult{Draft: cloneSavedWorkflowDraftPointer(draft)}
	}
	testBridge := &fakeBridge{envelope: bridge.GatewayEnvelope{
		Status: "ok", Response: map[string]any{"structured_answer": workflowRAGTestAnswer("official_guide", "Official guidance supports the answer.")},
	}}
	service := newWorkflowRAGExecutionService(reader, snapshotRepository, testBridge, runStore)
	service.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	service.envelopeOptions = func(northboundSelection, float64) bridge.EnvelopeOptions { return bridge.EnvelopeOptions{} }
	service.newRunID = func() (string, error) { return "run_execution0000001", nil }
	fixed := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixed }
	return &workflowRAGExecutionFixture{
		service: service, bridge: testBridge, runStore: runStore, snapshotRepository: snapshotRepository,
		snapshotContext: snapshotContext,
		runContext: WorkflowRunContext{
			RequestContext: context.Background(), RequestID: "request_rag_execution", TenantRef: snapshotContext.TenantRef,
			WorkspaceID: snapshotContext.WorkspaceID, ApplicationID: snapshotContext.ApplicationID, ActorRef: snapshotContext.ActorRef,
			ScopeGrants: append([]string{}, workflowRAGExecutionRequiredScopes...), AuditRef: "audit_rag_execution",
		},
		snapshot: *snapshot.Record, draft: draft,
	}
}

func workflowRAGEligibleDraft(ragRef string) SavedWorkflowDraft {
	node := func(id, nodeType string) SavedWorkflowDraftNode {
		return SavedWorkflowDraftNode{
			NodeID: id, NodeType: nodeType, Label: id, InputSummary: "Answer using only selected evidence.",
			OutputSummary: "structured advisory answer", InputContractRef: "contract_input", OutputContractRef: "contract_output",
			InputContractFields: []string{"input_text"}, OutputContractFields: []string{"answer"},
			OutputMappingSummary: "map structured advisory output", RiskLevel: "low",
		}
	}
	prompt := node("node_prompt", "prompt")
	retrieval := node("node_retrieval", "rag_retrieval")
	retrieval.RAGRef = ragRef
	model := node("node_model", "llm")
	model.ProviderRef = "profile:mock-workflow"
	output := node("node_output", "output")
	return SavedWorkflowDraft{
		DraftID: "draft_rag_execution", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		SourceDefinitionID: "workflow_rag_execution", BaseDefinitionVersion: 1, DraftVersion: 1,
		SchemaVersion: savedWorkflowDraftSchemaVersion, DraftStatus: SavedWorkflowDraftStatusValidForReview,
		CreatedAt: "2026-07-18T08:00:00Z", UpdatedAt: "2026-07-18T08:00:00Z",
		CreatedByActorRef: "subject_owner", UpdatedByActorRef: "subject_owner", Name: "RAG execution workflow",
		Description: "A reviewable prompt, retrieval, model, and output workflow.",
		Nodes:       []SavedWorkflowDraftNode{prompt, retrieval, model, output},
		Edges: []SavedWorkflowDraftEdge{
			{EdgeID: "edge_prompt_retrieval", FromNodeID: prompt.NodeID, ToNodeID: retrieval.NodeID},
			{EdgeID: "edge_retrieval_model", FromNodeID: retrieval.NodeID, ToNodeID: model.NodeID},
			{EdgeID: "edge_model_output", FromNodeID: model.NodeID, ToNodeID: output.NodeID},
		},
		InputContract:  SavedWorkflowDraftContract{ContractID: "contract_input", RequiredFields: []string{"input_text"}, Summary: "bounded query"},
		OutputContract: SavedWorkflowDraftContract{ContractID: "contract_output", RequiredFields: []string{"answer", "citations"}, Summary: "structured answer"},
		ProviderRefs:   []string{"profile:mock-workflow"}, RAGRefs: []string{ragRef},
		ValidationSummary:          SavedWorkflowDraftValidationSummary{ValidationState: SavedWorkflowDraftStatusValidForReview, ValidForReview: true},
		SampleOrUnsavedDraftStatus: "saved_draft_record",
	}
}

func workflowRAGTestSelection() northboundSelection {
	return northboundSelection{
		provider: "mock", providerProfile: "mock-default", model: "mock-rag", upstreamModel: "mock-rag",
		source: "development-default", inventoryKind: "builtin", credentialState: "not_required", deploymentMode: "local", authMode: "none",
	}
}

func workflowRAGTestAnswer(fragmentRef, claim string) WorkflowRAGAnswer {
	return WorkflowRAGAnswer{
		SchemaVersion: "workflow_rag_answer.v1", Answer: "Use the official guidance.",
		Citations:   []WorkflowRAGCitation{{FragmentRef: fragmentRef, ClaimSummary: claim}},
		Limitations: []string{"The answer is limited to the selected snapshot."}, Confidence: "medium",
	}
}

func workflowRAGExecutionFixtureRequest(fixture *workflowRAGExecutionFixture, query string, version int) WorkflowRAGExecutionRequest {
	return WorkflowRAGExecutionRequest{DraftID: fixture.draft.DraftID, DraftVersion: version, InputText: query, Model: "mock-rag"}
}

func mustWorkflowRAGAnswerJSON(t *testing.T, answer WorkflowRAGAnswer) string {
	t.Helper()
	payload, err := json.Marshal(answer)
	if err != nil {
		t.Fatalf("marshal workflow RAG answer: %v", err)
	}
	return string(payload)
}

func workflowRAGExecutionMemoryEntry(
	t *testing.T,
	repository *memoryWorkflowRAGSnapshotRepository,
	snapshotID string,
	ctx WorkflowRAGSnapshotContext,
) workflowRAGMemoryEntry {
	t.Helper()
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	entry, found := repository.entries[workflowRAGStoreKey(ctx, snapshotID)]
	if !found {
		t.Fatalf("workflow RAG memory entry %s not found", snapshotID)
	}
	return entry
}
