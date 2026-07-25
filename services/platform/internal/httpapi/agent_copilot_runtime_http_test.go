package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestAgentCopilotRuntimeHTTPDefaultClosedScopeAndWriteGate(t *testing.T) {
	server := &Server{}
	disabled := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications/app_aaaaaaaaaaaaaaaa/agent-copilot-runtime-assignment?workspace_id=workspace_one", nil)
	disabled.SetPathValue("application_id", "app_aaaaaaaaaaaaaaaa")
	recorder := httptest.NewRecorder()
	server.handleReadAgentCopilotRuntimeAssignment(recorder, disabled)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "AGENT_COPILOT_RUNTIME_DEV_HTTP_DISABLED") {
		t.Fatalf("Agent Copilot runtime HTTP was not default closed: %d body=%s", recorder.Code, recorder.Body.String())
	}

	server.config = config.Config{AgentCopilotRuntimeDevHTTPEnabled: true}
	server.agentCopilotRuntimeRepository = newMemoryAgentCopilotRuntimeRepository()
	read := agentCopilotRuntimeHTTPRequest(http.MethodGet, "agent_copilot_runtime:read", nil)
	readRecorder := httptest.NewRecorder()
	server.handleReadAgentCopilotRuntimeAssignment(readRecorder, read)
	if readRecorder.Code != http.StatusOK || !strings.Contains(readRecorder.Body.String(), AgentCopilotRuntimeFailureNotFound) {
		t.Fatalf("Agent Copilot runtime read scope failed: %d body=%s", readRecorder.Code, readRecorder.Body.String())
	}

	write := agentCopilotRuntimeHTTPRequest(http.MethodPost, "agent_copilot_runtime:write", strings.NewReader(`{"workspace_id":"workspace_one","expected_assignment_version":0,"action":"activate","candidate_id":"candidate-agent"}`))
	writeRecorder := httptest.NewRecorder()
	server.handleDecideAgentCopilotRuntimeAssignment(writeRecorder, write)
	if writeRecorder.Code != http.StatusOK || !strings.Contains(writeRecorder.Body.String(), AgentCopilotRuntimeFailureWriteDisabled) {
		t.Fatalf("Agent Copilot runtime write gate failed: %d body=%s", writeRecorder.Code, writeRecorder.Body.String())
	}
}

func TestAgentCopilotRuntimePermissionsDoNotImplyInvocation(t *testing.T) {
	grants := projectControlPlaneReadPermissions([]string{
		"radishmind.agent-copilot-runtime.read",
		"radishmind.agent-copilot-runtime.write",
	})
	for _, expected := range []string{"agent_copilot_runtime:read", "agent_copilot_runtime:write"} {
		if !controlPlaneReadHasScope(grants, expected) {
			t.Fatalf("upstream Agent Copilot runtime permission did not project %q: %#v", expected, grants)
		}
	}
	if controlPlaneReadHasScope(grants, "agent_copilot:invoke") ||
		controlPlaneReadHasScope(grants, "application_sessions:execute") {
		t.Fatalf("Agent Copilot runtime permission implied invocation authority: %#v", grants)
	}
}

func agentCopilotRuntimeHTTPRequest(method, scope string, body *strings.Reader) *http.Request {
	target := "/v1/user-workspace/applications/app_aaaaaaaaaaaaaaaa/agent-copilot-runtime-assignment"
	if method == http.MethodGet {
		target += "?workspace_id=workspace_one"
	}
	var request *http.Request
	if body == nil {
		request = httptest.NewRequest(method, target, nil)
	} else {
		request = httptest.NewRequest(method, target, body)
	}
	request.SetPathValue("application_id", "app_aaaaaaaaaaaaaaaa")
	auth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:dev",
		TenantBinding: "tenant:one", SubjectBinding: "subject:owner", ScopeGrants: []string{scope},
	}
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	request.Header.Set(agentCopilotRuntimeWorkspaceHeader, "workspace_one")
	request.Header.Set(agentCopilotRuntimeApplicationHeader, "app_aaaaaaaaaaaaaaaa")
	return request
}
