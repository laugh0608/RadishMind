package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationConfigurationDraftLifecycleAndIsolation(t *testing.T) {
	runApplicationConfigurationDraftLifecycleAndIsolation(t, newMemoryApplicationConfigurationDraftRepository())
}

func TestApplicationConfigurationDraftPromptTemplateBinding(t *testing.T) {
	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	templateRepository := newMemoryPromptApplicationTemplateRepository()
	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	draftContext.ActorRef = "subject_demo_user"
	draftContext.OwnerSubjectRef = "subject_demo_user"
	draftContext.TemplateBindingEnabled = true
	payload := validApplicationDraftPayload()
	payload.ApplicationID = draftContext.ApplicationID
	payload.DraftID = "app-config-prompt-assistant"
	payload.ApplicationKind = "prompt_application"

	templateContext := validPromptApplicationTemplateContext()
	templateContext.ApplicationID = draftContext.ApplicationID
	templateContext.ActorRef = draftContext.ActorRef
	templateContext.OwnerSubjectRef = draftContext.OwnerSubjectRef
	templateInput := validPromptApplicationTemplateDraftInput()
	templateInput.ApplicationID = draftContext.ApplicationID
	templateService := newPromptApplicationTemplateService(templateRepository)
	templateDraft := templateService.SaveDraft(templateContext, templateInput, 0)
	if templateDraft.Draft == nil {
		t.Fatalf("seed prompt template draft: %#v", templateDraft)
	}
	templateVersion := templateService.CreateVersion(templateContext, templateInput.TemplateID, 1)
	if templateVersion.Version == nil {
		t.Fatalf("seed prompt template version: %#v", templateVersion)
	}

	service := newApplicationConfigurationDraftService(draftRepository)
	service.readPromptTemplateVersion = func(ctx ApplicationConfigurationDraftContext, templateID string, version int) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(PromptApplicationTemplateContext{
			RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
		}, templateID, version)
		if err != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureVersionNotFound
		}
		return value, ""
	}
	created := service.Save(draftContext, payload, 0)
	if created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("seed prompt application draft: %#v", created)
	}
	bound := service.BindPromptTemplate(draftContext, payload.DraftID, PromptApplicationTemplateBindingInput{
		ExpectedDraftVersion: 1, TemplateID: templateInput.TemplateID, TemplateVersion: 1,
	})
	if bound.Draft == nil || bound.Draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV3 || bound.Draft.DraftVersion != 2 ||
		bound.Draft.PromptTemplateRef == nil || bound.Draft.PromptTemplateRef.TemplateDigest != templateVersion.Version.TemplateDigest || bound.Draft.WorkflowRAGBindingRef != nil {
		t.Fatalf("bind prompt template version: %#v", bound)
	}
	conflict := service.BindPromptTemplate(draftContext, payload.DraftID, PromptApplicationTemplateBindingInput{
		ExpectedDraftVersion: 1, TemplateID: templateInput.TemplateID, TemplateVersion: 1,
	})
	if conflict.FailureCode != ApplicationDraftFailureVersionConflict || conflict.CurrentDraftVersion != 2 {
		t.Fatalf("stale prompt template binding must fail CAS: %#v", conflict)
	}
	forgedPayload := bound.Draft.ApplicationConfigurationDraftPayload
	forgedPayload.PromptTemplateRef = clonePromptApplicationTemplateRef(forgedPayload.PromptTemplateRef)
	forgedPayload.PromptTemplateRef.TemplateDigest = strings.Repeat("a", 71)
	forged := service.Save(draftContext, forgedPayload, 2)
	if forged.FailureCode != ApplicationDraftFailurePayloadInvalid && forged.FailureCode != ApplicationDraftFailureTemplateBinding {
		t.Fatalf("forged prompt template digest must fail closed: %#v", forged)
	}
}

func TestApplicationConfigurationDraftPromptTemplateBindingHTTPRoute(t *testing.T) {
	server := NewServer(config.Config{ControlPlaneReadDevAuthEnabled: true, ApplicationDraftDevHTTPEnabled: true, ApplicationDraftDevWriteEnabled: true}, Options{BuildVersion: "test"})
	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	draftContext.ActorRef = "subject_demo_user"
	draftContext.OwnerSubjectRef = "subject_demo_user"
	payload := validApplicationDraftPayload()
	payload.ApplicationID = draftContext.ApplicationID
	payload.DraftID = "app-config-prompt-http"
	payload.ApplicationKind = "prompt_application"
	if saved := newApplicationConfigurationDraftService(server.applicationDraftRepository).Save(draftContext, payload, 0); saved.Draft == nil {
		t.Fatalf("seed HTTP binding draft: %#v", saved)
	}
	templateContext := validPromptApplicationTemplateContext()
	templateContext.ActorRef = draftContext.ActorRef
	templateContext.OwnerSubjectRef = draftContext.OwnerSubjectRef
	templateInput := validPromptApplicationTemplateDraftInput()
	templateService := newPromptApplicationTemplateService(server.promptApplicationTemplateRepository)
	if saved := templateService.SaveDraft(templateContext, templateInput, 0); saved.Draft == nil {
		t.Fatalf("seed HTTP binding template: %#v", saved)
	}
	if version := templateService.CreateVersion(templateContext, templateInput.TemplateID, 1); version.Version == nil {
		t.Fatalf("seed HTTP binding template version: %#v", version)
	}
	body := applicationConfigurationDraftPromptTemplateBindingBody{
		WorkspaceID: draftContext.WorkspaceID, ApplicationID: draftContext.ApplicationID, ExpectedDraftVersion: 1,
		TemplateID: templateInput.TemplateID, TemplateVersion: 1,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/application-configuration-drafts/"+payload.DraftID+"/prompt-template-binding", bytes.NewReader(raw))
	setApplicationDraftHeaders(request, "application_drafts:write,prompt_application_templates:bind", draftContext.ApplicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	var result applicationConfigurationDraftEnvelope
	if recorder.Code != http.StatusOK || json.Unmarshal(recorder.Body.Bytes(), &result) != nil || result.FailureCode != nil || result.Draft == nil || result.Draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV3 || result.Draft.PromptTemplateRef == nil {
		t.Fatalf("Prompt template binding route failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func runApplicationConfigurationDraftLifecycleAndIsolation(t *testing.T, repository applicationConfigurationDraftRepository) {
	t.Helper()
	service := newApplicationConfigurationDraftService(repository)
	service.now = func() time.Time { return time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC) }
	requestContext := validApplicationDraftContext()
	payload := validApplicationDraftPayload()

	validated := service.Validate(requestContext, payload)
	if !validated.ValidationSummary.IsValid || validated.FailureCode != "" {
		t.Fatalf("valid draft was rejected: %#v", validated)
	}
	created := service.Save(requestContext, payload, 0)
	if created.Draft == nil || created.Draft.DraftVersion != 1 || created.FailureCode != "" {
		t.Fatalf("first save failed: %#v", created)
	}
	payload.Description = "Updated public description"
	updated := service.Save(requestContext, payload, 1)
	if updated.Draft == nil || updated.Draft.DraftVersion != 2 || updated.Draft.CreatedAt != created.Draft.CreatedAt {
		t.Fatalf("CAS update failed: %#v", updated)
	}
	stale := service.Save(requestContext, payload, 1)
	if stale.FailureCode != ApplicationDraftFailureVersionConflict || stale.CurrentDraftVersion != 2 {
		t.Fatalf("stale save must expose current version without overwriting: %#v", stale)
	}
	read := service.Read(requestContext, payload.DraftID)
	if read.Draft == nil || read.Draft.Description != "Updated public description" || read.Draft.DraftVersion != 2 {
		t.Fatalf("saved draft changed after conflict: %#v", read)
	}
	summaries, failure := service.List(requestContext)
	if failure != "" || len(summaries) != 1 || summaries[0].DraftID != payload.DraftID {
		t.Fatalf("scoped list failed: %#v %s", summaries, failure)
	}
	otherApplication := requestContext
	otherApplication.ApplicationID = "app_docs_assistant"
	if result := service.Read(otherApplication, payload.DraftID); result.FailureCode != ApplicationDraftFailureNotFound {
		t.Fatalf("cross application read must fail closed: %#v", result)
	}
}

func TestApplicationConfigurationDraftValidationRejectsScopeProtocolAndSecrets(t *testing.T) {
	service := newApplicationConfigurationDraftService(newMemoryApplicationConfigurationDraftRepository())
	requestContext := validApplicationDraftContext()
	cases := []struct {
		name     string
		mutate   func(*ApplicationConfigurationDraftPayload)
		expected string
	}{
		{name: "scope", mutate: func(payload *ApplicationConfigurationDraftPayload) { payload.ApplicationID = "other" }, expected: ApplicationDraftFailureScopeDenied},
		{name: "protocol", mutate: func(payload *ApplicationConfigurationDraftPayload) { payload.DefaultProtocol = "messages" }, expected: ApplicationDraftFailurePayloadInvalid},
		{name: "secret", mutate: func(payload *ApplicationConfigurationDraftPayload) {
			payload.Description = "Authorization: Bearer secret"
		}, expected: ApplicationDraftFailureSecretForbidden},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			payload := validApplicationDraftPayload()
			testCase.mutate(&payload)
			result := service.Save(requestContext, payload, 0)
			if result.FailureCode != testCase.expected || result.Draft != nil {
				t.Fatalf("invalid draft must fail closed: %#v", result)
			}
		})
	}
}

func TestApplicationConfigurationDraftStoreFailureDoesNotFallback(t *testing.T) {
	repository := newMemoryApplicationConfigurationDraftRepository()
	repository.unavailable = true
	service := newApplicationConfigurationDraftService(repository)
	result := service.Save(validApplicationDraftContext(), validApplicationDraftPayload(), 0)
	if result.FailureCode != ApplicationDraftFailureStoreUnavailable || result.Draft != nil {
		t.Fatalf("store failure must not materialize an in-memory fallback: %#v", result)
	}
	summaries, failureCode := service.List(validApplicationDraftContext())
	if failureCode != ApplicationDraftFailureStoreUnavailable || len(summaries) != 0 {
		t.Fatalf("list failure must not return fallback summaries: %#v %s", summaries, failureCode)
	}
}

func TestApplicationConfigurationDraftWriteDisabledHasNoSideEffect(t *testing.T) {
	repository := newMemoryApplicationConfigurationDraftRepository()
	service := newApplicationConfigurationDraftService(repository)
	requestContext := validApplicationDraftContext()
	requestContext.WriteEnabled = false
	result := service.Save(requestContext, validApplicationDraftPayload(), 0)
	if result.FailureCode != ApplicationDraftFailureWriteDisabled || result.Draft != nil {
		t.Fatalf("write-disabled save must fail: %#v", result)
	}
	requestContext.WriteEnabled = true
	if read := service.Read(requestContext, validApplicationDraftPayload().DraftID); read.FailureCode != ApplicationDraftFailureNotFound {
		t.Fatalf("write-disabled save left a partial record: %#v", read)
	}
}

func TestApplicationConfigurationDraftHTTPRoutes(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:  true,
		ApplicationDraftDevHTTPEnabled:  true,
		ApplicationDraftDevWriteEnabled: true,
	}, Options{BuildVersion: "test"})
	payload := validApplicationDraftPayload()

	validate := performApplicationDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/application-drafts/validate", applicationConfigurationDraftValidateBody{Draft: payload}, "application_drafts:read,application_drafts:write", payload.ApplicationID)
	if validate.FailureCode != nil || !validate.ValidationSummary.IsValid || validate.Draft != nil {
		t.Fatalf("validate route failed: %#v", validate)
	}
	created := performApplicationDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/application-drafts", applicationConfigurationDraftSaveBody{Draft: payload}, "application_drafts:read,application_drafts:write", payload.ApplicationID)
	if created.FailureCode != nil || created.Draft == nil || created.CurrentDraftVersion != 1 {
		t.Fatalf("save route failed: %#v", created)
	}
	conflict := performApplicationDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/application-drafts", applicationConfigurationDraftSaveBody{Draft: payload}, "application_drafts:read,application_drafts:write", payload.ApplicationID)
	if conflict.FailureCode == nil || *conflict.FailureCode != ApplicationDraftFailureVersionConflict || conflict.CurrentDraftVersion != 1 {
		t.Fatalf("conflict route failed: %#v", conflict)
	}

	read := performApplicationDraftRequest(t, server, http.MethodGet, "/v1/user-workspace/application-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID, nil, "application_drafts:read", payload.ApplicationID)
	if read.Draft == nil || read.Draft.DraftVersion != 1 {
		t.Fatalf("read route failed: %#v", read)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-drafts?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID, nil)
	setApplicationDraftHeaders(request, "application_drafts:read", payload.ApplicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	var list applicationConfigurationDraftListEnvelope
	if recorder.Code != http.StatusOK || json.Unmarshal(recorder.Body.Bytes(), &list) != nil || len(list.DraftSummaries) != 1 || list.FailureCode != nil {
		t.Fatalf("list route failed: %d %s", recorder.Code, recorder.Body.String())
	}

	denied := performApplicationDraftRequest(t, server, http.MethodGet, "/v1/user-workspace/application-drafts/"+payload.DraftID+"?workspace_id="+payload.WorkspaceID+"&application_id="+payload.ApplicationID, nil, "application_drafts:read", "app_docs_assistant")
	if denied.FailureCode == nil || *denied.FailureCode != ApplicationDraftFailureScopeDenied || denied.Draft != nil {
		t.Fatalf("header scope mismatch must fail closed: %#v", denied)
	}
}

func validApplicationDraftContext() ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{
		RequestContext: context.Background(), RequestID: "request-application-draft-test", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ActorRef: "subject_platform_ops",
		OwnerSubjectRef: "subject_platform_ops", AuditRef: "audit_application_draft_test", WriteEnabled: true,
	}
}

func validApplicationDraftPayload() ApplicationConfigurationDraftPayload {
	return ApplicationConfigurationDraftPayload{
		DraftID: "app-config-app-flow-copilot", WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot",
		BaseApplicationUpdatedAt: "2026-05-31T10:20:00Z", SchemaVersion: applicationConfigurationDraftSchemaVersion,
		DisplayName: "RadishFlow Copilot", Description: "Advisory application configuration draft.",
		ApplicationKind: "workflow_copilot", DefaultProtocol: "responses", DefaultModel: "profile:local-dev",
		AllowedProtocols: []string{"chat_completions", "responses"},
	}
}

func performApplicationDraftRequest(t *testing.T, server *Server, method, target string, body any, scopes, applicationID string) applicationConfigurationDraftEnvelope {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal application draft body: %v", err)
		}
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(raw))
	setApplicationDraftHeaders(request, scopes, applicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope applicationConfigurationDraftEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode application draft response: %v", err)
	}
	return envelope
}

func setApplicationDraftHeaders(request *http.Request, scopes, applicationID string) {
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, scopes)
	request.Header.Set(applicationDraftDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(applicationDraftDevApplicationHeader, applicationID)
}
