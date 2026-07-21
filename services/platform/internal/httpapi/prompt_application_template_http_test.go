package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestPromptApplicationTemplateHTTPDefaultClosedAndStrictJSON(t *testing.T) {
	server := &Server{}
	disabled := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/prompt-application-templates/validate", strings.NewReader(`{}`))
	disabledRecorder := httptest.NewRecorder()
	server.handleValidatePromptApplicationTemplate(disabledRecorder, disabled)
	if disabledRecorder.Code != http.StatusForbidden || !strings.Contains(disabledRecorder.Body.String(), "PROMPT_APPLICATION_TEMPLATE_DEV_HTTP_DISABLED") {
		t.Fatalf("prompt template HTTP was not default closed: %d body=%s", disabledRecorder.Code, disabledRecorder.Body.String())
	}

	fixture := newPromptApplicationTemplateHTTPFixture(t, "prompt_application")
	unknown := promptApplicationTemplateSaveBody{ExpectedDraftVersion: 0, Template: fixture.input}
	payload, _ := json.Marshal(unknown)
	payload = []byte(strings.Replace(string(payload), `"template":{`, `"template":{"provider_api_key":"forbidden",`, 1))
	request := fixture.request(http.MethodPost, "/v1/user-workspace/prompt-application-templates", payload, []string{"prompt_application_templates:write"})
	recorder := httptest.NewRecorder()
	fixture.server.handleSavePromptApplicationTemplate(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "INVALID_JSON") || fixture.bridge.handleCalls != 0 {
		t.Fatalf("unknown credential field did not fail before owner/provider: %d body=%s calls=%d", recorder.Code, recorder.Body.String(), fixture.bridge.handleCalls)
	}
}

func TestPromptApplicationTemplatePermissionsProjectWithoutBroaderApplicationGrant(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.prompt-application-templates.read",
		"radishmind.prompt-application-templates.read-source",
		"radishmind.prompt-application-templates.write",
		"radishmind.prompt-application-templates.version",
		"radishmind.prompt-application-templates.bind",
	})
	for _, expected := range []string{
		"prompt_application_templates:read", "prompt_application_templates:read_source", "prompt_application_templates:write",
		"prompt_application_templates:version", "prompt_application_templates:bind",
	} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("upstream Prompt Application permission did not project %q: %#v", expected, grants)
		}
	}
	if controlPlaneReadHasScope(grants, "applications:write") {
		t.Fatalf("Prompt Application template permission implied broader catalog write: %#v", grants)
	}
}

func TestPromptApplicationTemplateHTTPScopesSummarySourceAndVersion(t *testing.T) {
	fixture := newPromptApplicationTemplateHTTPFixture(t, "prompt_application")
	validatePayload, _ := json.Marshal(promptApplicationTemplateValidateBody{Template: fixture.input})
	validate := fixture.request(http.MethodPost, "/v1/user-workspace/prompt-application-templates/validate", validatePayload, []string{"prompt_application_templates:write"})
	validateRecorder := httptest.NewRecorder()
	fixture.server.handleValidatePromptApplicationTemplate(validateRecorder, validate)
	var validated promptApplicationTemplateEnvelope
	if validateRecorder.Code != http.StatusOK || json.Unmarshal(validateRecorder.Body.Bytes(), &validated) != nil || !validated.ValidationSummary.IsValid || validated.Draft != nil {
		t.Fatalf("validate must be read-only and valid: %d envelope=%#v body=%s", validateRecorder.Code, validated, validateRecorder.Body.String())
	}

	savePayload, _ := json.Marshal(promptApplicationTemplateSaveBody{ExpectedDraftVersion: 0, Template: fixture.input})
	save := fixture.request(http.MethodPost, "/v1/user-workspace/prompt-application-templates", savePayload, []string{"prompt_application_templates:write"})
	saveRecorder := httptest.NewRecorder()
	fixture.server.handleSavePromptApplicationTemplate(saveRecorder, save)
	var saved promptApplicationTemplateEnvelope
	if saveRecorder.Code != http.StatusOK || json.Unmarshal(saveRecorder.Body.Bytes(), &saved) != nil || saved.Draft == nil || saved.FailureCode != nil || saved.CurrentDraftVersion != 1 {
		t.Fatalf("save failed: %d envelope=%#v body=%s", saveRecorder.Code, saved, saveRecorder.Body.String())
	}

	list := fixture.request(http.MethodGet, fixture.collectionQuery(), nil, []string{"prompt_application_templates:read"})
	listRecorder := httptest.NewRecorder()
	fixture.server.handleListPromptApplicationTemplates(listRecorder, list)
	var listed promptApplicationTemplateListEnvelope
	if listRecorder.Code != http.StatusOK || json.Unmarshal(listRecorder.Body.Bytes(), &listed) != nil || len(listed.DraftSummaries) != 1 || strings.Contains(listRecorder.Body.String(), `"messages"`) || strings.Contains(listRecorder.Body.String(), `"content"`) {
		t.Fatalf("summary route leaked source or failed: %d envelope=%#v body=%s", listRecorder.Code, listed, listRecorder.Body.String())
	}

	readOnly := fixture.request(http.MethodGet, fixture.detailQuery(fixture.input.TemplateID), nil, []string{"prompt_application_templates:read"})
	readOnly.SetPathValue("template_id", fixture.input.TemplateID)
	readOnlyRecorder := httptest.NewRecorder()
	fixture.server.handleReadPromptApplicationTemplate(readOnlyRecorder, readOnly)
	if !strings.Contains(readOnlyRecorder.Body.String(), PromptApplicationTemplateFailureScopeDenied) || strings.Contains(readOnlyRecorder.Body.String(), `"messages"`) {
		t.Fatalf("summary scope read template source: %s", readOnlyRecorder.Body.String())
	}

	readSource := fixture.request(http.MethodGet, fixture.detailQuery(fixture.input.TemplateID), nil, []string{"prompt_application_templates:read_source"})
	readSource.SetPathValue("template_id", fixture.input.TemplateID)
	readSourceRecorder := httptest.NewRecorder()
	fixture.server.handleReadPromptApplicationTemplate(readSourceRecorder, readSource)
	if readSourceRecorder.Code != http.StatusOK || !strings.Contains(readSourceRecorder.Body.String(), `"messages"`) || !strings.Contains(readSourceRecorder.Body.String(), `"template_digest"`) {
		t.Fatalf("source scope could not read exact draft: %d body=%s", readSourceRecorder.Code, readSourceRecorder.Body.String())
	}

	versionPayload, _ := json.Marshal(promptApplicationTemplateVersionCreateBody{WorkspaceID: fixture.input.WorkspaceID, ApplicationID: fixture.input.ApplicationID, SourceDraftVersion: 1})
	versionRequest := fixture.request(http.MethodPost, "/v1/user-workspace/prompt-application-templates/"+fixture.input.TemplateID+"/versions", versionPayload, []string{"prompt_application_templates:version"})
	versionRequest.SetPathValue("template_id", fixture.input.TemplateID)
	versionRecorder := httptest.NewRecorder()
	fixture.server.handleCreatePromptApplicationTemplateVersion(versionRecorder, versionRequest)
	var versioned promptApplicationTemplateEnvelope
	if versionRecorder.Code != http.StatusOK || json.Unmarshal(versionRecorder.Body.Bytes(), &versioned) != nil || versioned.Version == nil || versioned.Version.TemplateVersion != 1 || versioned.Version.TemplateDigest != saved.Draft.TemplateDigest {
		t.Fatalf("version create failed exact lineage: %d envelope=%#v body=%s", versionRecorder.Code, versioned, versionRecorder.Body.String())
	}

	versionList := fixture.request(http.MethodGet, fixture.versionQuery(fixture.input.TemplateID, ""), nil, []string{"prompt_application_templates:read"})
	versionList.SetPathValue("template_id", fixture.input.TemplateID)
	versionListRecorder := httptest.NewRecorder()
	fixture.server.handleListPromptApplicationTemplateVersions(versionListRecorder, versionList)
	var versions promptApplicationTemplateVersionListEnvelope
	if versionListRecorder.Code != http.StatusOK || json.Unmarshal(versionListRecorder.Body.Bytes(), &versions) != nil || len(versions.VersionSummaries) != 1 || strings.Contains(versionListRecorder.Body.String(), `"messages"`) {
		t.Fatalf("version summary route leaked source or failed: %d envelope=%#v body=%s", versionListRecorder.Code, versions, versionListRecorder.Body.String())
	}

	versionRead := fixture.request(http.MethodGet, fixture.versionQuery(fixture.input.TemplateID, "1"), nil, []string{"prompt_application_templates:read_source"})
	versionRead.SetPathValue("template_id", fixture.input.TemplateID)
	versionRead.SetPathValue("template_version", "1")
	versionReadRecorder := httptest.NewRecorder()
	fixture.server.handleReadPromptApplicationTemplateVersion(versionReadRecorder, versionRead)
	if versionReadRecorder.Code != http.StatusOK || !strings.Contains(versionReadRecorder.Body.String(), `"source_draft_version":1`) || !strings.Contains(versionReadRecorder.Body.String(), `"messages"`) || fixture.bridge.handleCalls != 0 {
		t.Fatalf("exact version source read failed or called provider: %d body=%s calls=%d", versionReadRecorder.Code, versionReadRecorder.Body.String(), fixture.bridge.handleCalls)
	}
}

func TestPromptApplicationTemplateHTTPRequiresPromptApplicationAndExactQuery(t *testing.T) {
	fixture := newPromptApplicationTemplateHTTPFixture(t, "agent")
	payload, _ := json.Marshal(promptApplicationTemplateSaveBody{ExpectedDraftVersion: 0, Template: fixture.input})
	request := fixture.request(http.MethodPost, "/v1/user-workspace/prompt-application-templates", payload, []string{"prompt_application_templates:write"})
	recorder := httptest.NewRecorder()
	fixture.server.handleSavePromptApplicationTemplate(recorder, request)
	if !strings.Contains(recorder.Body.String(), PromptApplicationTemplateFailureApplicationKind) {
		t.Fatalf("non-prompt application accepted template: %s", recorder.Body.String())
	}
	if summaries, failure := fixture.server.promptApplicationTemplateService().ListDrafts(validPromptApplicationTemplateContextForHTTP(fixture.input.ApplicationID)); failure != "" || len(summaries) != 0 {
		t.Fatalf("ineligible application wrote owner state: failure=%q summaries=%#v", failure, summaries)
	}

	unknownQuery := fixture.request(http.MethodGet, fixture.collectionQuery()+"&provider=forbidden", nil, []string{"prompt_application_templates:read"})
	unknownRecorder := httptest.NewRecorder()
	fixture.server.handleListPromptApplicationTemplates(unknownRecorder, unknownQuery)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), PromptApplicationTemplateFailurePayloadInvalid) || fixture.bridge.handleCalls != 0 {
		t.Fatalf("unknown query did not fail before owner/provider: %d body=%s calls=%d", unknownRecorder.Code, unknownRecorder.Body.String(), fixture.bridge.handleCalls)
	}
}

type promptApplicationTemplateHTTPFixture struct {
	server *Server
	bridge *fakeBridge
	auth   controlPlaneReadAuthContext
	input  PromptApplicationTemplateDraftInput
}

func newPromptApplicationTemplateHTTPFixture(t *testing.T, applicationKind string) promptApplicationTemplateHTTPFixture {
	t.Helper()
	catalogRepository := newMemoryApplicationCatalogRepository()
	catalogContext := ApplicationCatalogContext{
		RequestContext: validPromptApplicationTemplateContext().RequestContext, RequestID: "request_catalog_seed", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: "subject_owner", OwnerSubjectRef: "subject_owner", AuditRef: "audit_catalog_seed", WriteEnabled: true,
	}
	created := newApplicationCatalogService(catalogRepository).Create(catalogContext, ApplicationCatalogCreateInput{
		DisplayName: "Prompt app", Description: "Template owner test", ApplicationKind: applicationKind,
	})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("seed application catalog: %#v", created)
	}
	input := validPromptApplicationTemplateDraftInput()
	input.ApplicationID = created.Record.ApplicationID
	bridgeClient := &fakeBridge{}
	server := &Server{
		config: config.Config{
			PromptTemplateDevHTTPEnabled: true, PromptTemplateDevWriteEnabled: true,
		},
		bridge: bridgeClient, applicationCatalogRepository: catalogRepository,
		promptApplicationTemplateRepository: newMemoryPromptApplicationTemplateRepository(),
	}
	return promptApplicationTemplateHTTPFixture{
		server: server, bridge: bridgeClient, input: input,
		auth: controlPlaneReadAuthContext{
			AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:dev", TenantBinding: "tenant_demo", SubjectBinding: "subject_owner",
		},
	}
}

func (fixture promptApplicationTemplateHTTPFixture) request(method, target string, payload []byte, scopes []string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(string(payload)))
	auth := fixture.auth
	auth.ScopeGrants = append([]string{}, scopes...)
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	request.Header.Set(promptApplicationTemplateDevWorkspaceHeader, fixture.input.WorkspaceID)
	request.Header.Set(promptApplicationTemplateDevApplicationHeader, fixture.input.ApplicationID)
	return request
}

func (fixture promptApplicationTemplateHTTPFixture) collectionQuery() string {
	return "/v1/user-workspace/prompt-application-templates?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}

func (fixture promptApplicationTemplateHTTPFixture) detailQuery(templateID string) string {
	return "/v1/user-workspace/prompt-application-templates/" + templateID + "?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}

func (fixture promptApplicationTemplateHTTPFixture) versionQuery(templateID, version string) string {
	path := "/v1/user-workspace/prompt-application-templates/" + templateID + "/versions"
	if version != "" {
		path += "/" + version
	}
	return path + "?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}

func validPromptApplicationTemplateContextForHTTP(applicationID string) PromptApplicationTemplateContext {
	ctx := validPromptApplicationTemplateContext()
	ctx.ApplicationID = applicationID
	return ctx
}
