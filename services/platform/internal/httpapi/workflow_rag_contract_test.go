package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWorkflowRAGContractValidatorsAcceptVersionedShapes(t *testing.T) {
	service := newWorkflowRAGSnapshotService(newMemoryWorkflowRAGSnapshotRepository(nil))
	service.now = func() time.Time { return time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC) }
	service.newID = func(string) (string, error) { return "rags_dddddddddddddddd", nil }
	created := service.Create(workflowRAGTestContext(), WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "contract_manual", DisplayName: "Contract manual",
		ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("create contract fixture: %#v", created)
	}
	assertWorkflowRAGContractValid(t, workflowRAGSnapshotSchemaVersion, *created.Record)
	assertWorkflowRAGContractValid(t, workflowRAGFragmentSchemaVersion, created.Record.Fragments[0])
	assertWorkflowRAGContractValid(t, "workflow_rag_execution_profile.v1", workflowRAGLexicalProfile())
	answer := WorkflowRAGAnswer{
		SchemaVersion: "workflow_rag_answer.v1", Answer: "The evidence supports the advisory answer.",
		Citations:   []WorkflowRAGCitation{{FragmentRef: "official_guide", ClaimSummary: "Official guidance supports the answer."}},
		Limitations: []string{"Batch A validates this shape but does not call a provider."}, Confidence: "medium",
	}
	assertWorkflowRAGContractValid(t, "workflow_rag_answer.v1", answer)
	audit, err := buildWorkflowRAGSnapshotAudit(workflowRAGTestContext(), "snapshot_created", *created.Record, created.Record.CreatedAt)
	if err != nil {
		t.Fatalf("build contract audit: %v", err)
	}
	assertWorkflowRAGContractValid(t, workflowRAGAuditSchemaVersion, audit)

	completedAt := "2026-07-17T12:00:01Z"
	run := workflowRAGRunRecordV3{
		SchemaVersion: "workflow_run_record.v3", RecordVersion: 1, RunID: "run_aaaaaaaaaaaaaaaa",
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		DraftID: "draft_contract", DraftVersion: 1, DraftDigest: workflowRAGSHA256("draft"), Status: "succeeded", FailureSummary: "",
		StartedAt: "2026-07-17T12:00:00Z", CompletedAt: &completedAt,
		Snapshot: workflowRAGRunSnapshotBinding{
			SnapshotID: created.Record.SnapshotID, SnapshotVersion: 1,
			SnapshotDigest: created.Record.SnapshotDigest, RAGRef: created.Record.RAGRef,
		},
		RetrievalAttempt: workflowRAGRunRetrievalAttempt{
			NodeID: "node_retrieval", Status: "succeeded", ProfileID: workflowRAGProfileID,
			ProfileVersion: 1, ProfileDigest: workflowRAGLexicalProfile().ProfileDigest,
			QueryDigest: workflowRAGSHA256("query"), QueryBytes: 5, CandidateCount: 2,
			SelectedFragments: []workflowRAGRunSelectedFragment{{
				FragmentRef: "official_guide", ContentDigest: created.Record.Fragments[0].ContentDigest,
				Rank: 1, SourceType: "manual", IsOfficial: true,
			}},
			RetrievalLatencyMS: 2, ContextBytes: 27, CitationRefs: []string{"official_guide"},
		},
		Answer: nil, SelectedProvider: "mock", SelectedModel: "mock-rag",
		RequestID: "request_run_v3", AuditRef: "audit_run_v3", ActorRef: "subject_owner",
		SideEffects: workflowRAGRunSideEffects{RetrievalCalls: 1, ProviderCalls: 1},
		Diagnostic:  workflowRAGRunDiagnostic{FailureBoundary: "none", RetrievalFailureCategory: "none"},
	}
	assertWorkflowRAGContractValid(t, "workflow_run_record.v3", run)
}

func TestWorkflowRAGContractValidatorsRejectUnknownForbiddenAndVersionDrift(t *testing.T) {
	profilePayload, _ := json.Marshal(workflowRAGLexicalProfile())
	unknown := strings.TrimSuffix(string(profilePayload), "}") + `,"endpoint":"https://forbidden.invalid"}`
	if err := validateWorkflowRAGContractJSON("workflow_rag_execution_profile.v1", []byte(unknown)); err == nil {
		t.Fatal("workflow RAG profile accepted unknown network material")
	}
	if err := validateWorkflowRAGContractJSON("workflow_run_record.v3", []byte(`{"schema_version":"workflow_run_record.v2"}`)); err == nil {
		t.Fatal("workflow run v3 validator accepted v2 schema drift")
	}
	secretFragment := WorkflowRAGFragment{
		SchemaVersion: workflowRAGFragmentSchemaVersion, FragmentRef: "secret_fragment", SourceType: "manual",
		SourceRef: "manual.secret", PageSlug: "manual/secret", Content: "password=not-allowed",
		ContentClassification: "workspace_internal", ContentBytes: len("password=not-allowed"), ContentDigest: workflowRAGSHA256("password=not-allowed"),
	}
	payload, _ := json.Marshal(secretFragment)
	if err := validateWorkflowRAGContractJSON(workflowRAGFragmentSchemaVersion, payload); err == nil {
		t.Fatal("workflow RAG fragment validator accepted forbidden secret material")
	}
	answer := WorkflowRAGAnswer{
		SchemaVersion: "workflow_rag_answer.v1", Answer: "answer",
		Citations: []WorkflowRAGCitation{
			{FragmentRef: "same_fragment", ClaimSummary: "first"},
			{FragmentRef: "same_fragment", ClaimSummary: "duplicate"},
		},
		Limitations: []string{}, Confidence: "high",
	}
	payload, _ = json.Marshal(answer)
	if err := validateWorkflowRAGContractJSON("workflow_rag_answer.v1", payload); err == nil {
		t.Fatal("workflow RAG answer validator accepted duplicate citations")
	}
	invalidRun := workflowRAGRunRecordV3{
		SchemaVersion: "workflow_run_record.v3", RecordVersion: 1, RunID: "run_aaaaaaaaaaaaaaaa",
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		DraftID: "draft_contract", DraftVersion: 1, DraftDigest: workflowRAGSHA256("draft"), Status: "succeeded", StartedAt: "2026-07-17T12:00:00Z",
		Snapshot:         workflowRAGRunSnapshotBinding{SnapshotID: "rags_dddddddddddddddd", SnapshotVersion: 1, SnapshotDigest: workflowRAGSHA256("snapshot"), RAGRef: "workflow.rag.contract_manual.v1"},
		RetrievalAttempt: workflowRAGRunRetrievalAttempt{NodeID: "node_retrieval", Status: "succeeded", ProfileID: workflowRAGProfileID, ProfileVersion: 1, ProfileDigest: workflowRAGSHA256("profile"), QueryDigest: workflowRAGSHA256("query"), CandidateCount: 1},
		SelectedProvider: "https://forbidden.invalid", SelectedModel: "mock-rag", RequestID: "request_run_v3", AuditRef: "audit_run_v3", ActorRef: "subject_owner",
		SideEffects: workflowRAGRunSideEffects{RetrievalCalls: 1, ProviderCalls: 1}, Diagnostic: workflowRAGRunDiagnostic{FailureBoundary: "none", RetrievalFailureCategory: "none"},
	}
	payload, _ = json.Marshal(invalidRun)
	if err := validateWorkflowRAGContractJSON("workflow_run_record.v3", payload); err == nil {
		t.Fatal("workflow run v3 validator accepted unsafe provider material and incomplete success evidence")
	}
}

func assertWorkflowRAGContractValid(t *testing.T, contract string, document any) {
	t.Helper()
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal %s contract fixture: %v", contract, err)
	}
	if err = validateWorkflowRAGContractJSON(contract, payload); err != nil {
		t.Fatalf("validate %s contract fixture: %v\n%s", contract, err, payload)
	}
}
