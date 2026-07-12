package httpapi

import (
	"net/http"
	"strings"
)

const (
	applicationDraftSaveRoute     = "POST /v1/user-workspace/application-drafts"
	applicationDraftListRoute     = "GET /v1/user-workspace/application-drafts"
	applicationDraftReadRoute     = "GET /v1/user-workspace/application-drafts/{draft_id}"
	applicationDraftValidateRoute = "POST /v1/user-workspace/application-drafts/validate"

	applicationDraftDevWorkspaceHeader   = "X-RadishMind-Dev-Application-Draft-Workspace"
	applicationDraftDevApplicationHeader = "X-RadishMind-Dev-Application-Draft-Application"
)

type applicationConfigurationDraftSaveBody struct {
	ExpectedDraftVersion int                                  `json:"expected_draft_version"`
	Draft                ApplicationConfigurationDraftPayload `json:"draft"`
}

type applicationConfigurationDraftValidateBody struct {
	Draft ApplicationConfigurationDraftPayload `json:"draft"`
}

type applicationConfigurationDraftEnvelope struct {
	RequestID           string                                  `json:"request_id"`
	WorkspaceID         string                                  `json:"workspace_id"`
	ApplicationID       string                                  `json:"application_id"`
	Draft               *ApplicationConfigurationDraft          `json:"draft"`
	FailureCode         *string                                 `json:"failure_code"`
	CurrentDraftVersion int                                     `json:"current_draft_version"`
	ValidationSummary   ApplicationConfigurationDraftValidation `json:"validation_summary"`
	AuditRef            string                                  `json:"audit_ref"`
}

type applicationConfigurationDraftListEnvelope struct {
	RequestID      string                                 `json:"request_id"`
	WorkspaceID    string                                 `json:"workspace_id"`
	ApplicationID  string                                 `json:"application_id"`
	DraftSummaries []ApplicationConfigurationDraftSummary `json:"draft_summaries"`
	FailureCode    *string                                `json:"failure_code"`
	AuditRef       string                                 `json:"audit_ref"`
}

func (server *Server) handleValidateApplicationConfigurationDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationDraftValidateRoute)
	if !server.allowApplicationDraftDevHTTP(writer, request, trace) {
		return
	}
	var body applicationConfigurationDraftValidateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode := applicationConfigurationDraftContextFromRequest(request, trace, body.Draft.WorkspaceID, body.Draft.ApplicationID, "application_drafts:write", false, "validate")
	if failureCode != "" {
		writeApplicationConfigurationDraftResult(writer, trace, requestContext, ApplicationConfigurationDraftResult{FailureCode: failureCode})
		return
	}
	writeApplicationConfigurationDraftResult(writer, trace, requestContext, server.applicationConfigurationDraftService().Validate(requestContext, body.Draft))
}

func (server *Server) handleSaveApplicationConfigurationDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationDraftSaveRoute)
	if !server.allowApplicationDraftDevHTTP(writer, request, trace) {
		return
	}
	var body applicationConfigurationDraftSaveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode := applicationConfigurationDraftContextFromRequest(request, trace, body.Draft.WorkspaceID, body.Draft.ApplicationID, "application_drafts:write", server.config.ApplicationDraftDevWriteEnabled, "save")
	if failureCode != "" {
		writeApplicationConfigurationDraftResult(writer, trace, requestContext, ApplicationConfigurationDraftResult{FailureCode: failureCode})
		return
	}
	result := server.applicationConfigurationDraftService().Save(requestContext, body.Draft, body.ExpectedDraftVersion)
	writeApplicationConfigurationDraftResult(writer, trace, requestContext, result)
}

func (server *Server) handleReadApplicationConfigurationDraft(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationDraftReadRoute)
	if !server.allowApplicationDraftDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	requestContext, failureCode := applicationConfigurationDraftContextFromRequest(request, trace, workspaceID, applicationID, "application_drafts:read", false, "read")
	if failureCode != "" {
		writeApplicationConfigurationDraftResult(writer, trace, requestContext, ApplicationConfigurationDraftResult{FailureCode: failureCode})
		return
	}
	result := server.applicationConfigurationDraftService().Read(requestContext, request.PathValue("draft_id"))
	writeApplicationConfigurationDraftResult(writer, trace, requestContext, result)
}

func (server *Server) handleListApplicationConfigurationDrafts(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationDraftListRoute)
	if !server.allowApplicationDraftDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	requestContext, failureCode := applicationConfigurationDraftContextFromRequest(request, trace, workspaceID, applicationID, "application_drafts:read", false, "list")
	if failureCode != "" {
		writeApplicationConfigurationDraftListResult(writer, trace, requestContext, nil, failureCode)
		return
	}
	summaries, failureCode := server.applicationConfigurationDraftService().List(requestContext)
	writeApplicationConfigurationDraftListResult(writer, trace, requestContext, summaries, failureCode)
}

func (server *Server) allowApplicationDraftDevHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if server.config.ApplicationDraftDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "APPLICATION_DRAFT_DEV_HTTP_DISABLED", "application draft dev route requires explicit opt-in")
	return false
}

func (server *Server) applicationConfigurationDraftService() applicationConfigurationDraftService {
	if server.applicationDraftRepository == nil {
		server.applicationDraftRepository = &memoryApplicationConfigurationDraftRepository{drafts: make(map[string]ApplicationConfigurationDraft), unavailable: true}
	}
	return newApplicationConfigurationDraftService(server.applicationDraftRepository)
}

func applicationConfigurationDraftContextFromRequest(
	request *http.Request,
	trace requestTrace,
	workspaceID string,
	applicationID string,
	requiredScope string,
	writeEnabled bool,
	auditSuffix string,
) (ApplicationConfigurationDraftContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	requestContext := ApplicationConfigurationDraftContext{
		RequestContext: request.Context(), RequestID: trace.requestID,
		WorkspaceID: strings.TrimSpace(workspaceID), ApplicationID: strings.TrimSpace(applicationID),
		WriteEnabled: writeEnabled, AuditRef: "audit_" + trace.requestID + "_application-draft-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" || !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return requestContext, ApplicationDraftFailureScopeDenied
	}
	requestContext.TenantRef = strings.TrimSpace(auth.TenantBinding)
	requestContext.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	requestContext.OwnerSubjectRef = requestContext.ActorRef
	headerWorkspaceID := strings.TrimSpace(request.Header.Get(applicationDraftDevWorkspaceHeader))
	headerApplicationID := strings.TrimSpace(request.Header.Get(applicationDraftDevApplicationHeader))
	if requestContext.TenantRef == "" || requestContext.WorkspaceID == "" || requestContext.ApplicationID == "" ||
		headerWorkspaceID != requestContext.WorkspaceID || headerApplicationID != requestContext.ApplicationID {
		return requestContext, ApplicationDraftFailureScopeDenied
	}
	return requestContext, ""
}

func writeApplicationConfigurationDraftResult(
	writer http.ResponseWriter,
	trace requestTrace,
	requestContext ApplicationConfigurationDraftContext,
	result ApplicationConfigurationDraftResult,
) {
	validation := result.ValidationSummary
	if validation.Findings == nil {
		validation.Findings = []ApplicationConfigurationDraftValidationFinding{}
	}
	if validation.State == "" {
		validation.State = applicationDraftValidationInvalid
	}
	writeObservedJSON(writer, http.StatusOK, trace, applicationConfigurationDraftEnvelope{
		RequestID: trace.requestID, WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		Draft: result.Draft, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentDraftVersion: result.CurrentDraftVersion, ValidationSummary: validation, AuditRef: requestContext.AuditRef,
	})
}

func writeApplicationConfigurationDraftListResult(
	writer http.ResponseWriter,
	trace requestTrace,
	requestContext ApplicationConfigurationDraftContext,
	summaries []ApplicationConfigurationDraftSummary,
	failureCode string,
) {
	if summaries == nil {
		summaries = []ApplicationConfigurationDraftSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, applicationConfigurationDraftListEnvelope{
		RequestID: trace.requestID, WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		DraftSummaries: summaries, FailureCode: optionalApplicationDraftFailure(failureCode), AuditRef: requestContext.AuditRef,
	})
}

func optionalApplicationDraftFailure(failureCode string) *string {
	if strings.TrimSpace(failureCode) == "" {
		return nil
	}
	value := strings.TrimSpace(failureCode)
	return &value
}
