package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	agentCopilotProfileValidateRoute      = "POST /v1/user-workspace/agent-copilot-profiles/validate"
	agentCopilotProfileSaveRoute          = "POST /v1/user-workspace/agent-copilot-profiles"
	agentCopilotProfileListRoute          = "GET /v1/user-workspace/agent-copilot-profiles"
	agentCopilotProfileReadRoute          = "GET /v1/user-workspace/agent-copilot-profiles/{profile_id}"
	agentCopilotProfileVersionCreateRoute = "POST /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions"
	agentCopilotProfileVersionListRoute   = "GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions"
	agentCopilotProfileVersionReadRoute   = "GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions/{profile_version}"

	agentCopilotProfileDevWorkspaceHeader   = "X-RadishMind-Dev-Agent-Copilot-Profile-Workspace"
	agentCopilotProfileDevApplicationHeader = "X-RadishMind-Dev-Agent-Copilot-Profile-Application"
)

type agentCopilotProfileValidateBody struct {
	Profile AgentCopilotProfileDraftInput `json:"profile"`
}

type agentCopilotProfileSaveBody struct {
	ExpectedDraftVersion int                           `json:"expected_draft_version"`
	Profile              AgentCopilotProfileDraftInput `json:"profile"`
}

type agentCopilotProfileVersionCreateBody struct {
	WorkspaceID        string `json:"workspace_id"`
	ApplicationID      string `json:"application_id"`
	SourceDraftVersion int    `json:"source_draft_version"`
}

type agentCopilotProfileEnvelope struct {
	RequestID             string                                  `json:"request_id"`
	WorkspaceID           string                                  `json:"workspace_id"`
	ApplicationID         string                                  `json:"application_id"`
	Draft                 *AgentCopilotProfileDraftV1             `json:"draft"`
	Version               *AgentCopilotProfileVersionV1           `json:"version"`
	FailureCode           *string                                 `json:"failure_code"`
	CurrentDraftVersion   int                                     `json:"current_draft_version"`
	CurrentProfileVersion int                                     `json:"current_profile_version"`
	ValidationSummary     ApplicationConfigurationDraftValidation `json:"validation_summary"`
	AuditRef              string                                  `json:"audit_ref"`
}

type agentCopilotProfileListEnvelope struct {
	RequestID      string                            `json:"request_id"`
	WorkspaceID    string                            `json:"workspace_id"`
	ApplicationID  string                            `json:"application_id"`
	DraftSummaries []AgentCopilotProfileDraftSummary `json:"draft_summaries"`
	FailureCode    *string                           `json:"failure_code"`
	AuditRef       string                            `json:"audit_ref"`
}

type agentCopilotProfileVersionListEnvelope struct {
	RequestID        string                              `json:"request_id"`
	WorkspaceID      string                              `json:"workspace_id"`
	ApplicationID    string                              `json:"application_id"`
	ProfileID        string                              `json:"profile_id"`
	VersionSummaries []AgentCopilotProfileVersionSummary `json:"version_summaries"`
	FailureCode      *string                             `json:"failure_code"`
	AuditRef         string                              `json:"audit_ref"`
}

func (server *Server) handleValidateAgentCopilotProfile(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileValidateRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	var body agentCopilotProfileValidateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, body.Profile.WorkspaceID, body.Profile.ApplicationID, "agent_copilot_profiles:write", false, "validate")
	if failure != "" {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: failure})
		return
	}
	writeAgentCopilotProfileResult(writer, trace, ctx, server.agentCopilotProfileService().Validate(ctx, body.Profile))
}

func (server *Server) handleSaveAgentCopilotProfile(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileSaveRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	var body agentCopilotProfileSaveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, body.Profile.WorkspaceID, body.Profile.ApplicationID, "agent_copilot_profiles:write", server.config.AgentCopilotProfileDevWriteEnabled, "save")
	if failure != "" {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: failure})
		return
	}
	writeAgentCopilotProfileResult(writer, trace, ctx, server.agentCopilotProfileService().SaveDraft(ctx, body.Profile, body.ExpectedDraftVersion))
}

func (server *Server) handleListAgentCopilotProfiles(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileListRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	if !agentCopilotProfileQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, AgentCopilotProfileFailurePayloadInvalid, "agent copilot profile list query is invalid")
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "agent_copilot_profiles:read", false, "list")
	if failure != "" {
		writeAgentCopilotProfileListResult(writer, trace, ctx, nil, failure)
		return
	}
	summaries, failure := server.agentCopilotProfileService().ListDrafts(ctx)
	writeAgentCopilotProfileListResult(writer, trace, ctx, summaries, failure)
}

func (server *Server) handleReadAgentCopilotProfile(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileReadRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	if !agentCopilotProfileQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, AgentCopilotProfileFailurePayloadInvalid, "agent copilot profile detail query is invalid")
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "agent_copilot_profiles:read_source", false, "read")
	if failure != "" {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: failure})
		return
	}
	writeAgentCopilotProfileResult(writer, trace, ctx, server.agentCopilotProfileService().ReadDraft(ctx, request.PathValue("profile_id")))
}

func (server *Server) handleCreateAgentCopilotProfileVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileVersionCreateRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	var body agentCopilotProfileVersionCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "agent_copilot_profiles:version", server.config.AgentCopilotProfileDevWriteEnabled, "version")
	if failure != "" {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: failure})
		return
	}
	writeAgentCopilotProfileResult(writer, trace, ctx, server.agentCopilotProfileService().CreateVersion(ctx, request.PathValue("profile_id"), body.SourceDraftVersion))
}

func (server *Server) handleListAgentCopilotProfileVersions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileVersionListRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	if !agentCopilotProfileQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, AgentCopilotProfileFailurePayloadInvalid, "agent copilot profile version list query is invalid")
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "agent_copilot_profiles:read", false, "version-list")
	if failure != "" {
		writeAgentCopilotProfileVersionListResult(writer, trace, ctx, request.PathValue("profile_id"), nil, failure)
		return
	}
	summaries, failure := server.agentCopilotProfileService().ListVersions(ctx, request.PathValue("profile_id"))
	writeAgentCopilotProfileVersionListResult(writer, trace, ctx, request.PathValue("profile_id"), summaries, failure)
}

func (server *Server) handleReadAgentCopilotProfileVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, agentCopilotProfileVersionReadRoute)
	if !server.allowAgentCopilotProfileDevHTTP(writer, trace) {
		return
	}
	if !agentCopilotProfileQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, AgentCopilotProfileFailurePayloadInvalid, "agent copilot profile version detail query is invalid")
		return
	}
	ctx, failure := agentCopilotProfileContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "agent_copilot_profiles:read_source", false, "version-read")
	if failure != "" {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: failure})
		return
	}
	version, err := strconv.Atoi(strings.TrimSpace(request.PathValue("profile_version")))
	if err != nil || version < 1 {
		writeAgentCopilotProfileResult(writer, trace, ctx, AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailurePayloadInvalid})
		return
	}
	writeAgentCopilotProfileResult(writer, trace, ctx, server.agentCopilotProfileService().ReadVersion(ctx, request.PathValue("profile_id"), version))
}

func (server *Server) agentCopilotProfileService() agentCopilotProfileService {
	if server.agentCopilotProfileRepository == nil {
		server.agentCopilotProfileRepository = &memoryAgentCopilotProfileRepository{
			drafts: make(map[string]AgentCopilotProfileDraftV1), versions: make(map[string]map[int]AgentCopilotProfileVersionV1), unavailable: true,
		}
	}
	service := newAgentCopilotProfileService(server.agentCopilotProfileRepository)
	service.requireAgentApplication = func(ctx AgentCopilotProfileContext) string {
		result := server.applicationCatalogService().RequireActive(ApplicationCatalogContext{
			RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
			WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
		}, ctx.ApplicationID)
		if result.FailureCode != "" {
			switch result.FailureCode {
			case ApplicationCatalogFailureNotFound:
				return AgentCopilotProfileFailureApplicationMissing
			case ApplicationCatalogFailureArchived:
				return AgentCopilotProfileFailureApplicationArchived
			default:
				return AgentCopilotProfileFailureStoreUnavailable
			}
		}
		if result.Record == nil || result.Record.ApplicationKind != "agent" {
			return AgentCopilotProfileFailureApplicationKind
		}
		return ""
	}
	return service
}

func agentCopilotProfileContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, requiredScope string, writeEnabled bool, auditSuffix string) (AgentCopilotProfileContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	ctx := AgentCopilotProfileContext{
		RequestContext: request.Context(), RequestID: trace.requestID, WorkspaceID: strings.TrimSpace(workspaceID),
		ApplicationID: strings.TrimSpace(applicationID), WriteEnabled: writeEnabled,
		AuditRef: "audit_" + trace.requestID + "_agent-copilot-profile-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" ||
		!controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return ctx, AgentCopilotProfileFailureScopeDenied
	}
	ctx.TenantRef = strings.TrimSpace(auth.TenantBinding)
	ctx.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	ctx.OwnerSubjectRef = ctx.ActorRef
	if ctx.TenantRef == "" || ctx.WorkspaceID == "" || ctx.ApplicationID == "" ||
		strings.TrimSpace(request.Header.Get(agentCopilotProfileDevWorkspaceHeader)) != ctx.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(agentCopilotProfileDevApplicationHeader)) != ctx.ApplicationID {
		return ctx, AgentCopilotProfileFailureScopeDenied
	}
	return ctx, ""
}

func (server *Server) allowAgentCopilotProfileDevHTTP(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.AgentCopilotProfileDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "AGENT_COPILOT_PROFILE_DEV_HTTP_DISABLED", "agent copilot profile route requires explicit development opt-in")
	return false
}

func agentCopilotProfileQueryAllowed(request *http.Request, allowed ...string) bool {
	allowedKeys := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedKeys[key] = struct{}{}
	}
	for key := range request.URL.Query() {
		if _, ok := allowedKeys[key]; !ok {
			return false
		}
	}
	return true
}

func writeAgentCopilotProfileResult(writer http.ResponseWriter, trace requestTrace, ctx AgentCopilotProfileContext, result AgentCopilotProfileResult) {
	validation := result.ValidationSummary
	if validation.Findings == nil {
		validation.Findings = []ApplicationConfigurationDraftValidationFinding{}
	}
	if validation.State == "" {
		validation.State = applicationDraftValidationInvalid
	}
	writeObservedJSON(writer, http.StatusOK, trace, agentCopilotProfileEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Draft: result.Draft, Version: result.Version, FailureCode: optionalAgentCopilotProfileFailure(result.FailureCode),
		CurrentDraftVersion: result.CurrentDraftVersion, CurrentProfileVersion: result.CurrentProfileVersion,
		ValidationSummary: validation, AuditRef: ctx.AuditRef,
	})
}

func writeAgentCopilotProfileListResult(writer http.ResponseWriter, trace requestTrace, ctx AgentCopilotProfileContext, summaries []AgentCopilotProfileDraftSummary, failure string) {
	if summaries == nil {
		summaries = []AgentCopilotProfileDraftSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, agentCopilotProfileListEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftSummaries: summaries, FailureCode: optionalAgentCopilotProfileFailure(failure), AuditRef: ctx.AuditRef,
	})
}

func writeAgentCopilotProfileVersionListResult(writer http.ResponseWriter, trace requestTrace, ctx AgentCopilotProfileContext, profileID string, summaries []AgentCopilotProfileVersionSummary, failure string) {
	if summaries == nil {
		summaries = []AgentCopilotProfileVersionSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, agentCopilotProfileVersionListEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		ProfileID: strings.TrimSpace(profileID), VersionSummaries: summaries,
		FailureCode: optionalAgentCopilotProfileFailure(failure), AuditRef: ctx.AuditRef,
	})
}

func optionalAgentCopilotProfileFailure(failure string) *string {
	failure = strings.TrimSpace(failure)
	if failure == "" {
		return nil
	}
	return &failure
}
