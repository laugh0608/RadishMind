package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestPromptApplicationRuntimeAssignmentLifecycleAndExactAuthority(t *testing.T) {
	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	candidateRepository := newMemoryApplicationPublishCandidateRepository()
	templateRepository := newMemoryPromptApplicationTemplateRepository()
	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	draftContext.TemplateBindingEnabled = true
	templateContext := validPromptApplicationTemplateContext()
	templateContext.ActorRef = draftContext.ActorRef
	templateContext.OwnerSubjectRef = draftContext.OwnerSubjectRef
	templateInput := validPromptApplicationTemplateDraftInput()
	templateService := newPromptApplicationTemplateService(templateRepository)
	if result := templateService.SaveDraft(templateContext, templateInput, 0); result.Draft == nil {
		t.Fatalf("seed template draft: %#v", result)
	}
	if result := templateService.CreateVersion(templateContext, templateInput.TemplateID, 1); result.Version == nil {
		t.Fatalf("seed template version: %#v", result)
	}
	payload := validApplicationDraftPayload()
	payload.ApplicationID = draftContext.ApplicationID
	payload.DraftID = "app-config-prompt-runtime"
	payload.ApplicationKind = "prompt_application"
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.readPromptTemplateVersion = func(ctx ApplicationConfigurationDraftContext, templateID string, version int) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, templateID, version)
		if err != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureVersionNotFound
		}
		return value, ""
	}
	if result := draftService.Save(draftContext, payload, 0); result.Draft == nil {
		t.Fatalf("seed configuration draft: %#v", result)
	}
	bound := draftService.BindPromptTemplate(draftContext, payload.DraftID, PromptApplicationTemplateBindingInput{ExpectedDraftVersion: 1, TemplateID: templateInput.TemplateID, TemplateVersion: 1})
	if bound.Draft == nil {
		t.Fatalf("bind template version: %#v", bound)
	}
	publishContext := validApplicationPublishContext()
	publishContext.ApplicationID = draftContext.ApplicationID
	publishContext.PromptTemplateSourceReadEnabled = true
	readBaseline := func(ctx ApplicationPublishContext) (ApplicationSummary, error) {
		return ApplicationSummary{ApplicationRef: ctx.ApplicationID, ApplicationKind: "prompt_application", UpdatedAt: payload.BaseApplicationUpdatedAt}, nil
	}
	publishService := newApplicationPublishCandidateService(draftRepository, candidateRepository, readBaseline)
	publishService.readPromptTemplateVersion = func(ctx ApplicationPublishContext, ref PromptApplicationTemplateRef) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, ref.TemplateID, ref.TemplateVersion)
		if err != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureVersionNotFound
		}
		return value, ""
	}
	createAndApprove := func(candidateID string) {
		t.Helper()
		created := publishService.Create(publishContext, ApplicationPublishCreateInput{CandidateID: candidateID, DraftID: payload.DraftID, ExpectedDraftVersion: 2})
		if created.Candidate == nil {
			t.Fatalf("create candidate %s: %#v", candidateID, created)
		}
		approved := publishService.Review(publishContext, candidateID, ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed exact Prompt runtime authority lineage."})
		if approved.Candidate == nil || approved.Candidate.CandidateState != applicationPublishStateApproved {
			t.Fatalf("approve candidate %s: %#v", candidateID, approved)
		}
	}
	createAndApprove("candidate-prompt-runtime-v1")

	runtimeContext := PromptApplicationRuntimeContext{
		RequestContext: draftContext.RequestContext, RequestID: "request_prompt_runtime", TenantRef: draftContext.TenantRef,
		WorkspaceID: draftContext.WorkspaceID, ApplicationID: draftContext.ApplicationID, ActorRef: draftContext.ActorRef,
		OwnerSubjectRef: draftContext.OwnerSubjectRef, AuditRef: "audit_prompt_runtime", WriteEnabled: true,
	}
	runtimeRepository := newMemoryPromptApplicationRuntimeRepository()
	service := newPromptApplicationRuntimeService(runtimeRepository, promptApplicationRuntimeAuthorityResolver{
		publishRepository: candidateRepository, draftRepository: draftRepository, templateRepository: templateRepository, readApplication: readBaseline,
	})
	service.now = func() time.Time { return time.Date(2026, 7, 21, 13, 0, 0, 0, time.UTC) }
	eventSequence := 0
	service.newID = func(prefix string) (string, error) {
		if prefix == "ptra_" {
			return "ptra_aaaaaaaaaaaaaaaa", nil
		}
		eventSequence++
		values := []string{"ptrae_aaaaaaaaaaaaaaaa", "ptrae_bbbbbbbbbbbbbbbb", "ptrae_cccccccccccccccc"}
		return values[eventSequence-1], nil
	}
	if before := service.Read(runtimeContext); before.FailureCode != PromptApplicationRuntimeFailureNotFound {
		t.Fatalf("candidate approval must not auto-activate assignment: %#v", before)
	}
	activated := service.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: "candidate-prompt-runtime-v1"})
	if activated.Assignment == nil || activated.Assignment.State != "active" || activated.Assignment.AssignmentVersion != 1 || len(activated.Events) != 1 {
		t.Fatalf("activate Prompt assignment: %#v", activated)
	}
	createAndApprove("candidate-prompt-runtime-v2")
	if drifted := service.Read(runtimeContext); drifted.Assignment == nil || drifted.FailureCode != PromptApplicationRuntimeFailureAuthorityChanged {
		t.Fatalf("superseded active assignment must fail read-time authority: %#v", drifted)
	}
	conflict := service.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{ExpectedAssignmentVersion: 0, Action: "replace", CandidateID: "candidate-prompt-runtime-v2"})
	if conflict.FailureCode != PromptApplicationRuntimeFailureVersionConflict || conflict.CurrentAssignmentVersion != 1 {
		t.Fatalf("stale replacement must fail CAS: %#v", conflict)
	}
	replaced := service.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{ExpectedAssignmentVersion: 1, Action: "replace", CandidateID: "candidate-prompt-runtime-v2"})
	if replaced.Assignment == nil || replaced.Assignment.CandidateID != "candidate-prompt-runtime-v2" || replaced.Assignment.AssignmentVersion != 2 || len(replaced.Events) != 2 {
		t.Fatalf("replace Prompt assignment: %#v", replaced)
	}
	revoked := service.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{ExpectedAssignmentVersion: 2, Action: "revoke"})
	if revoked.Assignment == nil || revoked.Assignment.State != "revoked" || revoked.Assignment.RevokedAt == nil || revoked.Assignment.AssignmentVersion != 3 || len(revoked.Events) != 3 {
		t.Fatalf("revoke Prompt assignment: %#v", revoked)
	}
	invalid := service.Decide(runtimeContext, PromptApplicationRuntimeDecisionInput{ExpectedAssignmentVersion: 3, Action: "replace", CandidateID: "candidate-prompt-runtime-v1"})
	if invalid.FailureCode != PromptApplicationRuntimeFailureTransition {
		t.Fatalf("revoked assignment must not be restored: %#v", invalid)
	}
}

func TestPromptApplicationRuntimeHTTPGatesAndScopes(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationDraftDevHTTPEnabled: true, ApplicationPublishDevHTTPEnabled: true,
		PromptTemplateDevHTTPEnabled: true, PromptApplicationRuntimeDevHTTPEnabled: true, PromptApplicationRuntimeDevWriteEnabled: true,
	}, Options{BuildVersion: "test"})
	applicationID := "app_aaaaaaaaaaaaaaaa"
	body, err := json.Marshal(promptApplicationRuntimeDecisionBody{WorkspaceID: "workspace_demo", ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: "candidate-missing"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+applicationID+"/prompt-runtime-assignment/decisions", bytes.NewReader(body))
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, "prompt_application_runtime:write")
	request.Header.Set(promptApplicationRuntimeWorkspaceHeader, "workspace_demo")
	request.Header.Set(promptApplicationRuntimeApplicationHeader, applicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	var result promptApplicationRuntimeEnvelope
	if recorder.Code != http.StatusOK || json.Unmarshal(recorder.Body.Bytes(), &result) != nil || result.FailureCode == nil || *result.FailureCode != PromptApplicationRuntimeFailureCandidate {
		t.Fatalf("Prompt runtime decision route did not fail closed on missing authority: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	read := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications/"+applicationID+"/prompt-runtime-assignment?workspace_id=workspace_demo", nil)
	setControlPlaneReadDevAuthHeaders(read)
	read.Header.Set(controlPlaneReadDevScopesHeader, "prompt_application_runtime:read")
	read.Header.Set(promptApplicationRuntimeWorkspaceHeader, "workspace_demo")
	read.Header.Set(promptApplicationRuntimeApplicationHeader, applicationID)
	readRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(readRecorder, read)
	if readRecorder.Code != http.StatusOK || json.Unmarshal(readRecorder.Body.Bytes(), &result) != nil || result.FailureCode == nil || *result.FailureCode != PromptApplicationRuntimeFailureNotFound {
		t.Fatalf("Prompt runtime read route did not return stable not-found: status=%d body=%s", readRecorder.Code, readRecorder.Body.String())
	}
}
