package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkflowDefinitionReleaseReviewMaterializesImmutableVersionWithoutActivation(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	ctx := workflowDefinitionTestContext()
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	candidate, err := store.CreateCandidate(ctx, "candidate-one", "definition-one", workflowDefinitionTestDraft(), now)
	if err != nil || candidate.State != workflowDefinitionStatePending {
		t.Fatalf("create: %#v %v", candidate, err)
	}
	approved, version, err := store.Review(ctx, candidate.CandidateID, 0, "approve", "reviewed and approved", candidate.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || approved.State != workflowDefinitionStateApproved || version == nil || version.Version != 1 {
		t.Fatalf("review: %#v %#v %v", approved, version, err)
	}
	if _, exists := store.activations[workflowDefinitionScopeKey(ctx, candidate.DefinitionID)]; exists {
		t.Fatal("approval must not activate")
	}
	if version.DefinitionDigest != candidate.DefinitionDigest || version.SourceDraftVersion != 1 {
		t.Fatal("version authority drift")
	}
}

func TestWorkflowDefinitionReleaseCASAllowsOneReviewAndOneActivation(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	ctx := workflowDefinitionTestContext()
	now := time.Now().UTC()
	candidate, _ := store.CreateCandidate(ctx, "candidate-cas", "definition-cas", workflowDefinitionTestDraft(), now)
	var wg sync.WaitGroup
	successes, conflicts := 0, 0
	var mu sync.Mutex
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := store.Review(ctx, candidate.CandidateID, 0, "approve", "concurrent review", candidate.SourceDraftDigest, now)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successes++
			} else if errors.Is(err, errWorkflowDefinitionConflict) || errors.Is(err, errWorkflowDefinitionInvalidState) {
				conflicts++
			}
		}()
	}
	wg.Wait()
	if successes != 1 || conflicts != 1 {
		t.Fatalf("review successes=%d conflicts=%d", successes, conflicts)
	}
	successes, conflicts = 0, 0
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.Activate(ctx, "definition-cas", 0, 1, now)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successes++
			} else if errors.Is(err, errWorkflowDefinitionConflict) {
				conflicts++
			}
		}()
	}
	wg.Wait()
	if successes != 1 || conflicts != 1 {
		t.Fatalf("activation successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestWorkflowDefinitionReleaseRejectsBlockedDraftAndScopeIsolation(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	ctx := workflowDefinitionTestContext()
	draft := workflowDefinitionTestDraft()
	draft.BlockedCapabilitySummary = []SavedWorkflowDraftBlockedCapability{{CapabilityID: "http_tool"}}
	if _, err := store.CreateCandidate(ctx, "candidate-blocked", "definition-blocked", draft, time.Now()); !errors.Is(err, errWorkflowDefinitionInvalidState) {
		t.Fatalf("expected invalid state, got %v", err)
	}
	draft = workflowDefinitionTestDraft()
	candidate, _ := store.CreateCandidate(ctx, "candidate-scope", "definition-scope", draft, time.Now())
	other := ctx
	other.ApplicationID = "app_other"
	if _, _, err := store.Review(other, candidate.CandidateID, 0, "approve", "wrong scope", candidate.SourceDraftDigest, time.Now()); !errors.Is(err, errWorkflowDefinitionNotFound) {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}

func TestWorkflowDefinitionReleaseRejectsSensitiveSnapshotMaterial(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	draft := workflowDefinitionTestDraft()
	draft.Description = "authorization: bearer secret-value"
	if _, err := store.CreateCandidate(workflowDefinitionTestContext(), "candidate-sensitive", "definition-sensitive", draft, time.Now()); !errors.Is(err, errWorkflowDefinitionPayloadInvalid) {
		t.Fatalf("sensitive snapshot material must be rejected, got %v", err)
	}
}

func TestWorkflowDefinitionReleaseServiceReloadsExactDraftAndRejectsDrift(t *testing.T) {
	draftStore := newMemorySavedWorkflowDraftStore()
	draftService := newSavedWorkflowDraftService(draftStore)
	draftContext := savedWorkflowDraftTestContext()
	payload := validSavedWorkflowDraftPayload()
	if result := draftService.SaveDraft(draftContext, SaveWorkflowDraftRequest{Payload: payload}); result.Draft == nil {
		t.Fatalf("seed draft: %#v", result)
	}
	store := newWorkflowDefinitionReleaseStore()
	service := newWorkflowDefinitionReleaseService(draftStore, store)
	service.now = func() time.Time { return time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC) }
	ctx := workflowDefinitionTestContext()
	ctx.ApplicationID = payload.ApplicationID
	created := service.Create(ctx, WorkflowDefinitionCandidateCreateInput{
		CandidateID:          "candidate-authority",
		DefinitionID:         "definition-authority",
		DraftID:              payload.DraftID,
		ExpectedDraftVersion: 1,
	})
	if created.Candidate == nil || created.FailureCode != "" {
		t.Fatalf("create candidate: %#v", created)
	}
	payload.Description = "A newer saved draft version must invalidate the pending authority."
	if result := draftService.SaveDraft(draftContext, SaveWorkflowDraftRequest{Payload: payload, ExpectedDraftVersion: 1}); result.Draft == nil || result.Draft.DraftVersion != 2 {
		t.Fatalf("update draft: %#v", result)
	}
	reviewed := service.Review(ctx, created.Candidate.CandidateID, WorkflowDefinitionReviewInput{
		ExpectedReviewVersion: 0,
		Decision:              "approve",
		Reason:                "reviewed authority evidence",
	})
	if reviewed.FailureCode != workflowDefinitionFailureSourceDigestDrift || reviewed.Version != nil {
		t.Fatalf("drift must fail before version materialization: %#v", reviewed)
	}
	if versions, failure := service.ListVersions(ctx, created.Candidate.DefinitionID); failure != "" || len(versions) != 0 {
		t.Fatalf("drift left a partial version: %#v failure=%s", versions, failure)
	}
}

func TestWorkflowDefinitionReleaseStoreReturnsImmutableCopiesAndAtomicActivationEvidence(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	ctx := workflowDefinitionTestContext()
	now := time.Date(2026, 7, 19, 13, 0, 0, 0, time.UTC)
	candidate, err := store.CreateCandidate(ctx, "candidate-copy", "definition-copy", workflowDefinitionTestDraft(), now)
	if err != nil {
		t.Fatal(err)
	}
	candidate.Snapshot.Nodes[0].NodeID = "mutated"
	stored, err := store.ReadCandidate(ctx, candidate.CandidateID)
	if err != nil || stored.Snapshot.Nodes[0].NodeID == "mutated" {
		t.Fatalf("candidate repository leaked mutable storage: %#v err=%v", stored, err)
	}
	approved, version, err := store.Review(ctx, stored.CandidateID, 0, "approve", "reviewed immutable copy", stored.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || version == nil {
		t.Fatalf("approve: %#v %#v %v", approved, version, err)
	}
	activation, err := store.DecideActivation(ctx, stored.DefinitionID, 0, "activate", 1, "activate reviewed version", now.Add(2*time.Minute))
	if err != nil || activation.PointerVersion != 1 || len(activation.Events) != 1 {
		t.Fatalf("activate: %#v %v", activation, err)
	}
	if len(store.audits[workflowDefinitionScopeKey(ctx, "activation")]) != 1 {
		t.Fatalf("activation audit missing: %#v", store.audits)
	}
	deactivated, err := store.DecideActivation(ctx, stored.DefinitionID, 1, "deactivate", 0, "deactivate reviewed version", now.Add(3*time.Minute))
	if err != nil || deactivated.State != workflowDefinitionActivationInactive || deactivated.ActiveVersion != 0 || len(deactivated.Events) != 2 {
		t.Fatalf("deactivate: %#v %v", deactivated, err)
	}
}

func TestWorkflowDefinitionReleaseCodecUsesStrictCanonicalFieldNames(t *testing.T) {
	store := newWorkflowDefinitionReleaseStore()
	ctx := workflowDefinitionTestContext()
	candidate, err := store.CreateCandidate(ctx, "candidate-codec", "definition-codec", workflowDefinitionTestDraft(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"candidate_id", "definition_id", "source_draft_digest", "definition_digest", "activation_eligible", "eligibility_blockers"} {
		if _, ok := document[key]; !ok {
			t.Fatalf("canonical candidate field %q missing: %s", key, payload)
		}
	}
	snapshot := document["snapshot"].(map[string]any)
	node := snapshot["nodes"].([]any)[0].(map[string]any)
	if _, ok := node["node_id"]; !ok {
		t.Fatalf("canonical node_id missing: %s", payload)
	}
	for _, forbidden := range []string{"NodeID", "InputContractFields", "Authorization", "credential"} {
		if bytes.Contains(payload, []byte(`"`+forbidden+`"`)) {
			t.Fatalf("forbidden or non-canonical field %q in candidate: %s", forbidden, payload)
		}
	}
}

func TestWorkflowDefinitionReleaseHTTPAuthorityReviewVersionAndActivation(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:      true,
		WorkflowSavedDraftDevHTTPEnabled:    true,
		WorkflowSavedDraftDevWriteEnabled:   true,
		WorkflowDefinitionReleaseDevEnabled: true,
	}, Options{BuildVersion: "test"})
	draftContext := savedWorkflowDraftTestContext()
	payload := validSavedWorkflowDraftPayload()
	payload.ToolRefs = []string{}
	payload.RAGRefs = []string{}
	if result := server.savedWorkflowDraftService().SaveDraft(draftContext, SaveWorkflowDraftRequest{Payload: payload}); result.Draft == nil {
		t.Fatalf("seed draft: %#v", result)
	}
	created := performWorkflowDefinitionRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-definition-candidates", workflowDefinitionCandidateCreateBody{
		CandidateID:          "candidate-http",
		DefinitionID:         "definition-http",
		DraftID:              payload.DraftID,
		ExpectedDraftVersion: 1,
	}, "workflow_definitions:write")
	if created.Candidate == nil || created.FailureCode != nil {
		t.Fatalf("create: %#v", created)
	}
	detail := performWorkflowDefinitionRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definition-candidates/candidate-http?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_definitions:read")
	if detail.Candidate == nil || detail.Candidate.CandidateID != created.Candidate.CandidateID {
		t.Fatalf("candidate detail: %#v", detail)
	}
	listRecorder := performWorkflowDefinitionRawRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definition-candidates?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_definitions:read")
	var candidateList workflowDefinitionCandidateListEnvelope
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &candidateList); err != nil || len(candidateList.Candidates) != 1 || candidateList.FailureCode != nil {
		t.Fatalf("candidate list: %s err=%v", listRecorder.Body.String(), err)
	}
	approved := performWorkflowDefinitionRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-definition-candidates/candidate-http/decisions", workflowDefinitionCandidateDecisionBody{
		ExpectedReviewVersion: 0,
		Decision:              "approve",
		Reason:                "reviewed through verified actor",
	}, "workflow_definitions:review")
	if approved.Candidate == nil || approved.Version == nil || approved.Activation != nil || approved.Candidate.State != workflowDefinitionStateApproved {
		t.Fatalf("approval must only materialize version: %#v", approved)
	}
	versionsRecorder := performWorkflowDefinitionRawRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definitions/definition-http/versions?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_definitions:read")
	var versionList workflowDefinitionVersionListEnvelope
	if err := json.Unmarshal(versionsRecorder.Body.Bytes(), &versionList); err != nil || len(versionList.Versions) != 1 || versionList.FailureCode != nil {
		t.Fatalf("version list: %s err=%v", versionsRecorder.Body.String(), err)
	}
	activated := performWorkflowDefinitionRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-definitions/definition-http/activation-decisions", workflowDefinitionActivationDecisionBody{
		ExpectedPointerVersion: 0,
		Decision:               "activate",
		Version:                1,
		Reason:                 "activate reviewed version",
	}, "workflow_definitions:activate")
	if activated.Activation == nil || activated.Activation.ActiveVersion != 1 || len(activated.Activation.Events) != 1 {
		t.Fatalf("activate: %#v", activated)
	}
	activationRead := performWorkflowDefinitionRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definitions/definition-http/activation?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_definitions:read")
	if activationRead.Activation == nil || activationRead.Activation.PointerVersion != 1 {
		t.Fatalf("activation read: %#v", activationRead)
	}
	read := performWorkflowDefinitionRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definitions/definition-http/versions/1?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_definitions:read")
	if read.Version == nil || read.Version.DefinitionDigest != approved.Version.DefinitionDigest {
		t.Fatalf("read immutable version: %#v", read)
	}
}

func TestWorkflowDefinitionReleaseHTTPRejectsUnknownFieldAndMissingScope(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:      true,
		WorkflowSavedDraftDevHTTPEnabled:    true,
		WorkflowSavedDraftDevWriteEnabled:   true,
		WorkflowDefinitionReleaseDevEnabled: true,
	}, Options{BuildVersion: "test"})
	unknown := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-definition-candidates", bytes.NewBufferString(`{"candidate_id":"candidate-http","definition_id":"definition-http","draft_id":"draft-http","expected_draft_version":1,"state":"approved"}`))
	setWorkflowDefinitionHeaders(unknown, "workflow_definitions:write")
	unknownRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownRecorder, unknown)
	if unknownRecorder.Code != http.StatusBadRequest {
		t.Fatalf("unknown field must be rejected: %d %s", unknownRecorder.Code, unknownRecorder.Body.String())
	}
	denied := performWorkflowDefinitionRequest(t, server, http.MethodGet, "/v1/user-workspace/workflow-definition-candidates/candidate-http?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "workflow_drafts:read")
	if denied.FailureCode == nil || *denied.FailureCode != workflowDefinitionFailureScopeDenied {
		t.Fatalf("missing workflow definition scope must fail closed: %#v", denied)
	}
}

func TestWorkflowDefinitionReleaseHTTPIsDisabledByDefault(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
	}, Options{BuildVersion: "test"})
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-definition-candidates?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setWorkflowDefinitionHeaders(request, "workflow_definitions:read")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code == http.StatusOK || !strings.Contains(recorder.Body.String(), "WORKFLOW_DEFINITION_RELEASE_DEV_HTTP_DISABLED") {
		t.Fatalf("workflow definition release route must default closed: %d %s", recorder.Code, recorder.Body.String())
	}
}

func performWorkflowDefinitionRequest(t *testing.T, server *Server, method, target string, body any, scopes string) workflowDefinitionReleaseEnvelope {
	t.Helper()
	recorder := performWorkflowDefinitionRawRequest(t, server, method, target, body, scopes)
	var envelope workflowDefinitionReleaseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope
}

func performWorkflowDefinitionRawRequest(t *testing.T, server *Server, method, target string, body any, scopes string) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(raw))
	setWorkflowDefinitionHeaders(request, scopes)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
	}
	return recorder
}

func setWorkflowDefinitionHeaders(request *http.Request, scopes string) {
	setSavedWorkflowDraftDevHeaders(request, scopes)
}

func workflowDefinitionTestContext() WorkflowDefinitionReleaseContext {
	return WorkflowDefinitionReleaseContext{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: "app_demo", OwnerSubjectRef: "subject_demo", ActorRef: "subject_demo", RequestID: "request_demo", AuditRef: "audit_demo"}
}

func workflowDefinitionTestDraft() SavedWorkflowDraft {
	return SavedWorkflowDraft{DraftID: "draft_demo", WorkspaceID: "workspace_demo", ApplicationID: "app_demo", DraftVersion: 1, SchemaVersion: savedWorkflowDraftSchemaVersion,
		DraftStatus: SavedWorkflowDraftStatusValidForReview, Name: "Reviewed workflow", Nodes: []SavedWorkflowDraftNode{{NodeID: "prompt", NodeType: "prompt"}, {NodeID: "output", NodeType: "output"}},
		Edges: []SavedWorkflowDraftEdge{{EdgeID: "edge-one", FromNodeID: "prompt", ToNodeID: "output"}}, ValidationSummary: SavedWorkflowDraftValidationSummary{ValidationState: SavedWorkflowDraftStatusValidForReview, ValidForReview: true},
		BlockedCapabilitySummary: []SavedWorkflowDraftBlockedCapability{}, ProviderRefs: []string{}, ToolRefs: []string{}, RAGRefs: []string{}, RequestedCapabilities: []string{}}
}
