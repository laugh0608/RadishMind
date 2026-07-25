package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestAgentCopilotProfileHTTPDefaultClosedStrictJSONAndWriteGate(t *testing.T) {
	server := &Server{}
	disabled := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles/validate", strings.NewReader(`{}`))
	disabledRecorder := httptest.NewRecorder()
	server.handleValidateAgentCopilotProfile(disabledRecorder, disabled)
	if disabledRecorder.Code != http.StatusForbidden || !strings.Contains(disabledRecorder.Body.String(), "AGENT_COPILOT_PROFILE_DEV_HTTP_DISABLED") {
		t.Fatalf("profile HTTP was not default closed: %d body=%s", disabledRecorder.Code, disabledRecorder.Body.String())
	}

	fixture := newAgentCopilotProfileHTTPFixture(t, "agent")
	saveBody := agentCopilotProfileSaveBody{ExpectedDraftVersion: 0, Profile: fixture.input}
	payload, _ := json.Marshal(saveBody)
	payload = []byte(strings.Replace(string(payload), `"profile":{`, `"profile":{"provider_api_key":"forbidden",`, 1))
	request := fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles", payload, []string{"agent_copilot_profiles:write"})
	recorder := httptest.NewRecorder()
	fixture.server.handleSaveAgentCopilotProfile(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "INVALID_JSON") ||
		len(fixture.server.agentCopilotProfileRepository.(*memoryAgentCopilotProfileRepository).drafts) != 0 || fixture.bridge.handleCalls != 0 {
		t.Fatalf("unknown credential field did not fail before owner/provider: %d body=%s calls=%d", recorder.Code, recorder.Body.String(), fixture.bridge.handleCalls)
	}

	fixture.server.config.AgentCopilotProfileDevWriteEnabled = false
	payload, _ = json.Marshal(saveBody)
	request = fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles", payload, []string{"agent_copilot_profiles:write"})
	recorder = httptest.NewRecorder()
	fixture.server.handleSaveAgentCopilotProfile(recorder, request)
	if !strings.Contains(recorder.Body.String(), AgentCopilotProfileFailureWriteDisabled) ||
		len(fixture.server.agentCopilotProfileRepository.(*memoryAgentCopilotProfileRepository).drafts) != 0 {
		t.Fatalf("write gate did not fail before mutation: %s", recorder.Body.String())
	}
}

func TestAgentCopilotProfilePermissionsProjectWithoutBroaderApplicationGrant(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.agent-copilot-profiles.read",
		"radishmind.agent-copilot-profiles.read-source",
		"radishmind.agent-copilot-profiles.write",
		"radishmind.agent-copilot-profiles.version",
		"radishmind.agent-copilot-profiles.bind",
	})
	for _, expected := range []string{
		"agent_copilot_profiles:read", "agent_copilot_profiles:read_source", "agent_copilot_profiles:write",
		"agent_copilot_profiles:version", "agent_copilot_profiles:bind",
	} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("upstream profile permission did not project %q: %#v", expected, grants)
		}
	}
	if controlPlaneReadHasScope(grants, "applications:write") || controlPlaneReadHasScope(grants, "agent_copilot:invoke") {
		t.Fatalf("profile permission implied broader application/runtime authority: %#v", grants)
	}
}

func TestAgentCopilotProfileHTTPScopesSummarySourceAndVersion(t *testing.T) {
	fixture := newAgentCopilotProfileHTTPFixture(t, "agent")
	validatePayload, _ := json.Marshal(agentCopilotProfileValidateBody{Profile: fixture.input})
	validate := fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles/validate", validatePayload, []string{"agent_copilot_profiles:write"})
	validateRecorder := httptest.NewRecorder()
	fixture.server.handleValidateAgentCopilotProfile(validateRecorder, validate)
	var validated agentCopilotProfileEnvelope
	if validateRecorder.Code != http.StatusOK || json.Unmarshal(validateRecorder.Body.Bytes(), &validated) != nil ||
		!validated.ValidationSummary.IsValid || validated.Draft != nil ||
		len(fixture.server.agentCopilotProfileRepository.(*memoryAgentCopilotProfileRepository).drafts) != 0 {
		t.Fatalf("validate must be read-only and valid: %d envelope=%#v body=%s", validateRecorder.Code, validated, validateRecorder.Body.String())
	}

	savePayload, _ := json.Marshal(agentCopilotProfileSaveBody{ExpectedDraftVersion: 0, Profile: fixture.input})
	save := fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles", savePayload, []string{"agent_copilot_profiles:write"})
	saveRecorder := httptest.NewRecorder()
	fixture.server.handleSaveAgentCopilotProfile(saveRecorder, save)
	var saved agentCopilotProfileEnvelope
	if saveRecorder.Code != http.StatusOK || json.Unmarshal(saveRecorder.Body.Bytes(), &saved) != nil ||
		saved.Draft == nil || saved.FailureCode != nil || saved.CurrentDraftVersion != 1 {
		t.Fatalf("save failed: %d envelope=%#v body=%s", saveRecorder.Code, saved, saveRecorder.Body.String())
	}

	list := fixture.request(http.MethodGet, fixture.collectionQuery(), nil, []string{"agent_copilot_profiles:read"})
	listRecorder := httptest.NewRecorder()
	fixture.server.handleListAgentCopilotProfiles(listRecorder, list)
	var listed agentCopilotProfileListEnvelope
	if listRecorder.Code != http.StatusOK || json.Unmarshal(listRecorder.Body.Bytes(), &listed) != nil ||
		len(listed.DraftSummaries) != 1 || listed.DraftSummaries[0].AllowedTasksDigest == "" ||
		strings.Contains(listRecorder.Body.String(), `"context_policy"`) || strings.Contains(listRecorder.Body.String(), `"tool_hints_policy"`) {
		t.Fatalf("summary route leaked source or failed: %d envelope=%#v body=%s", listRecorder.Code, listed, listRecorder.Body.String())
	}

	readOnly := fixture.request(http.MethodGet, fixture.detailQuery(fixture.input.ProfileID), nil, []string{"agent_copilot_profiles:read"})
	readOnly.SetPathValue("profile_id", fixture.input.ProfileID)
	readOnlyRecorder := httptest.NewRecorder()
	fixture.server.handleReadAgentCopilotProfile(readOnlyRecorder, readOnly)
	if !strings.Contains(readOnlyRecorder.Body.String(), AgentCopilotProfileFailureScopeDenied) ||
		strings.Contains(readOnlyRecorder.Body.String(), `"context_policy"`) {
		t.Fatalf("summary scope read profile source: %s", readOnlyRecorder.Body.String())
	}

	readSource := fixture.request(http.MethodGet, fixture.detailQuery(fixture.input.ProfileID), nil, []string{"agent_copilot_profiles:read_source"})
	readSource.SetPathValue("profile_id", fixture.input.ProfileID)
	readSourceRecorder := httptest.NewRecorder()
	fixture.server.handleReadAgentCopilotProfile(readSourceRecorder, readSource)
	if readSourceRecorder.Code != http.StatusOK || !strings.Contains(readSourceRecorder.Body.String(), `"context_policy"`) ||
		!strings.Contains(readSourceRecorder.Body.String(), `"profile_digest"`) {
		t.Fatalf("source scope could not read exact draft: %d body=%s", readSourceRecorder.Code, readSourceRecorder.Body.String())
	}

	versionPayload, _ := json.Marshal(agentCopilotProfileVersionCreateBody{
		WorkspaceID: fixture.input.WorkspaceID, ApplicationID: fixture.input.ApplicationID, SourceDraftVersion: 1,
	})
	versionRequest := fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles/"+fixture.input.ProfileID+"/versions", versionPayload, []string{"agent_copilot_profiles:version"})
	versionRequest.SetPathValue("profile_id", fixture.input.ProfileID)
	versionRecorder := httptest.NewRecorder()
	fixture.server.handleCreateAgentCopilotProfileVersion(versionRecorder, versionRequest)
	var versioned agentCopilotProfileEnvelope
	if versionRecorder.Code != http.StatusOK || json.Unmarshal(versionRecorder.Body.Bytes(), &versioned) != nil ||
		versioned.Version == nil || versioned.Version.ProfileVersion != 1 ||
		versioned.Version.ProfileDigest != saved.Draft.ProfileDigest || versioned.Version.PolicyDigest != saved.Draft.PolicyDigest {
		t.Fatalf("version create failed exact lineage: %d envelope=%#v body=%s", versionRecorder.Code, versioned, versionRecorder.Body.String())
	}

	versionList := fixture.request(http.MethodGet, fixture.versionQuery(fixture.input.ProfileID, ""), nil, []string{"agent_copilot_profiles:read"})
	versionList.SetPathValue("profile_id", fixture.input.ProfileID)
	versionListRecorder := httptest.NewRecorder()
	fixture.server.handleListAgentCopilotProfileVersions(versionListRecorder, versionList)
	var versions agentCopilotProfileVersionListEnvelope
	if versionListRecorder.Code != http.StatusOK || json.Unmarshal(versionListRecorder.Body.Bytes(), &versions) != nil ||
		len(versions.VersionSummaries) != 1 || versions.VersionSummaries[0].AllowedTasksDigest == "" ||
		strings.Contains(versionListRecorder.Body.String(), `"context_policy"`) {
		t.Fatalf("version summary route leaked source or failed: %d envelope=%#v body=%s", versionListRecorder.Code, versions, versionListRecorder.Body.String())
	}

	versionRead := fixture.request(http.MethodGet, fixture.versionQuery(fixture.input.ProfileID, "1"), nil, []string{"agent_copilot_profiles:read_source"})
	versionRead.SetPathValue("profile_id", fixture.input.ProfileID)
	versionRead.SetPathValue("profile_version", "1")
	versionReadRecorder := httptest.NewRecorder()
	fixture.server.handleReadAgentCopilotProfileVersion(versionReadRecorder, versionRead)
	if versionReadRecorder.Code != http.StatusOK || !strings.Contains(versionReadRecorder.Body.String(), `"source_draft_version":1`) ||
		!strings.Contains(versionReadRecorder.Body.String(), `"context_policy"`) || fixture.bridge.handleCalls != 0 {
		t.Fatalf("exact version source read failed or called provider: %d body=%s calls=%d", versionReadRecorder.Code, versionReadRecorder.Body.String(), fixture.bridge.handleCalls)
	}
}

func TestAgentCopilotProfileHTTPRequiresAgentApplicationAndExactQuery(t *testing.T) {
	fixture := newAgentCopilotProfileHTTPFixture(t, "prompt_application")
	payload, _ := json.Marshal(agentCopilotProfileSaveBody{ExpectedDraftVersion: 0, Profile: fixture.input})
	request := fixture.request(http.MethodPost, "/v1/user-workspace/agent-copilot-profiles", payload, []string{"agent_copilot_profiles:write"})
	recorder := httptest.NewRecorder()
	fixture.server.handleSaveAgentCopilotProfile(recorder, request)
	if !strings.Contains(recorder.Body.String(), AgentCopilotProfileFailureApplicationKind) {
		t.Fatalf("non-agent application accepted profile: %s", recorder.Body.String())
	}
	if len(fixture.server.agentCopilotProfileRepository.(*memoryAgentCopilotProfileRepository).drafts) != 0 {
		t.Fatal("ineligible application wrote profile owner state")
	}

	unknownQuery := fixture.request(http.MethodGet, fixture.collectionQuery()+"&provider=forbidden", nil, []string{"agent_copilot_profiles:read"})
	unknownRecorder := httptest.NewRecorder()
	fixture.server.handleListAgentCopilotProfiles(unknownRecorder, unknownQuery)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), AgentCopilotProfileFailurePayloadInvalid) ||
		fixture.bridge.handleCalls != 0 {
		t.Fatalf("unknown query did not fail before owner/provider: %d body=%s calls=%d", unknownRecorder.Code, unknownRecorder.Body.String(), fixture.bridge.handleCalls)
	}
}

type agentCopilotProfileHTTPFixture struct {
	server *Server
	bridge *fakeBridge
	auth   controlPlaneReadAuthContext
	input  AgentCopilotProfileDraftInput
}

func newAgentCopilotProfileHTTPFixture(t *testing.T, applicationKind string) agentCopilotProfileHTTPFixture {
	t.Helper()
	catalogRepository := newMemoryApplicationCatalogRepository()
	catalogContext := ApplicationCatalogContext{
		RequestContext: agentCopilotProfileTestContext("tenant_demo", "workspace_demo", "app_aaaaaaaaaaaaaaaa", "subject_owner").RequestContext,
		RequestID:      "request_catalog_seed", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
		ActorRef: "subject_owner", OwnerSubjectRef: "subject_owner", AuditRef: "audit_catalog_seed", WriteEnabled: true,
	}
	created := newApplicationCatalogService(catalogRepository).Create(catalogContext, ApplicationCatalogCreateInput{
		DisplayName: "Agent app", Description: "Profile owner test", ApplicationKind: applicationKind,
	})
	if created.FailureCode != "" || created.Record == nil {
		t.Fatalf("seed application catalog: %#v", created)
	}
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	input.WorkspaceID = "workspace_demo"
	input.ApplicationID = created.Record.ApplicationID
	bridgeClient := &fakeBridge{}
	server := &Server{
		config: config.Config{
			AgentCopilotProfileDevHTTPEnabled: true, AgentCopilotProfileDevWriteEnabled: true,
		},
		bridge: bridgeClient, applicationCatalogRepository: catalogRepository,
		agentCopilotProfileRepository: newMemoryAgentCopilotProfileRepository(),
	}
	return agentCopilotProfileHTTPFixture{
		server: server, bridge: bridgeClient, input: input,
		auth: controlPlaneReadAuthContext{
			AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:dev",
			TenantBinding: "tenant_demo", SubjectBinding: "subject_owner",
		},
	}
}

func (fixture agentCopilotProfileHTTPFixture) request(method, target string, payload []byte, scopes []string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(string(payload)))
	auth := fixture.auth
	auth.ScopeGrants = append([]string{}, scopes...)
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	request.Header.Set(agentCopilotProfileDevWorkspaceHeader, fixture.input.WorkspaceID)
	request.Header.Set(agentCopilotProfileDevApplicationHeader, fixture.input.ApplicationID)
	return request
}

func (fixture agentCopilotProfileHTTPFixture) collectionQuery() string {
	return "/v1/user-workspace/agent-copilot-profiles?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}

func (fixture agentCopilotProfileHTTPFixture) detailQuery(profileID string) string {
	return "/v1/user-workspace/agent-copilot-profiles/" + profileID + "?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}

func (fixture agentCopilotProfileHTTPFixture) versionQuery(profileID, version string) string {
	path := "/v1/user-workspace/agent-copilot-profiles/" + profileID + "/versions"
	if version != "" {
		path += "/" + version
	}
	return path + "?workspace_id=" + fixture.input.WorkspaceID + "&application_id=" + fixture.input.ApplicationID
}
