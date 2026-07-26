package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	promptApplicationTemplateValidateRoute      = "POST /v1/user-workspace/prompt-application-templates/validate"
	promptApplicationTemplateSaveRoute          = "POST /v1/user-workspace/prompt-application-templates"
	promptApplicationTemplateListRoute          = "GET /v1/user-workspace/prompt-application-templates"
	promptApplicationTemplateReadRoute          = "GET /v1/user-workspace/prompt-application-templates/{template_id}"
	promptApplicationTemplateVersionCreateRoute = "POST /v1/user-workspace/prompt-application-templates/{template_id}/versions"
	promptApplicationTemplateVersionListRoute   = "GET /v1/user-workspace/prompt-application-templates/{template_id}/versions"
	promptApplicationTemplateVersionReadRoute   = "GET /v1/user-workspace/prompt-application-templates/{template_id}/versions/{template_version}"

	promptApplicationTemplateDevWorkspaceHeader   = "X-RadishMind-Dev-Prompt-Template-Workspace"
	promptApplicationTemplateDevApplicationHeader = "X-RadishMind-Dev-Prompt-Template-Application"
)

type promptApplicationTemplateValidateBody struct {
	Template PromptApplicationTemplateDraftInput `json:"template"`
}

type promptApplicationTemplateSaveBody struct {
	ExpectedDraftVersion int                                 `json:"expected_draft_version"`
	Template             PromptApplicationTemplateDraftInput `json:"template"`
}

type promptApplicationTemplateVersionCreateBody struct {
	WorkspaceID        string `json:"workspace_id"`
	ApplicationID      string `json:"application_id"`
	SourceDraftVersion int    `json:"source_draft_version"`
}

type promptApplicationTemplateEnvelope struct {
	RequestID              string                              `json:"request_id"`
	WorkspaceID            string                              `json:"workspace_id"`
	ApplicationID          string                              `json:"application_id"`
	Draft                  *PromptApplicationTemplateDraft     `json:"draft"`
	Version                *PromptApplicationTemplateVersion   `json:"version"`
	FailureCode            *string                             `json:"failure_code"`
	CurrentDraftVersion    int                                 `json:"current_draft_version"`
	CurrentTemplateVersion int                                 `json:"current_template_version"`
	ValidationSummary      PromptApplicationTemplateValidation `json:"validation_summary"`
	AuditRef               string                              `json:"audit_ref"`
}

type promptApplicationTemplateListEnvelope struct {
	RequestID      string                                  `json:"request_id"`
	WorkspaceID    string                                  `json:"workspace_id"`
	ApplicationID  string                                  `json:"application_id"`
	DraftSummaries []PromptApplicationTemplateDraftSummary `json:"draft_summaries"`
	FailureCode    *string                                 `json:"failure_code"`
	AuditRef       string                                  `json:"audit_ref"`
}

type promptApplicationTemplateVersionListEnvelope struct {
	RequestID        string                                    `json:"request_id"`
	WorkspaceID      string                                    `json:"workspace_id"`
	ApplicationID    string                                    `json:"application_id"`
	TemplateID       string                                    `json:"template_id"`
	VersionSummaries []PromptApplicationTemplateVersionSummary `json:"version_summaries"`
	FailureCode      *string                                   `json:"failure_code"`
	AuditRef         string                                    `json:"audit_ref"`
}

func (server *Server) handleValidatePromptApplicationTemplate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateValidateRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	var body promptApplicationTemplateValidateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, body.Template.WorkspaceID, body.Template.ApplicationID, "prompt_application_templates:write", false, "validate")
	if failure != "" {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: failure})
		return
	}
	writePromptApplicationTemplateResult(writer, trace, ctx, server.promptApplicationTemplateService().Validate(ctx, body.Template))
}

func (server *Server) handleSavePromptApplicationTemplate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateSaveRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	var body promptApplicationTemplateSaveBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, body.Template.WorkspaceID, body.Template.ApplicationID, "prompt_application_templates:write", server.config.PromptTemplateDevWriteEnabled, "save")
	if failure != "" {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: failure})
		return
	}
	writePromptApplicationTemplateResult(writer, trace, ctx, server.promptApplicationTemplateService().SaveDraft(ctx, body.Template, body.ExpectedDraftVersion))
}

func (server *Server) handleListPromptApplicationTemplates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateListRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	if !promptApplicationTemplateQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, PromptApplicationTemplateFailurePayloadInvalid, "prompt template list query is invalid")
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "prompt_application_templates:read", false, "list")
	if failure != "" {
		writePromptApplicationTemplateListResult(writer, trace, ctx, nil, failure)
		return
	}
	summaries, failure := server.promptApplicationTemplateService().ListDrafts(ctx)
	writePromptApplicationTemplateListResult(writer, trace, ctx, summaries, failure)
}

func (server *Server) handleReadPromptApplicationTemplate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateReadRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	if !promptApplicationTemplateQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, PromptApplicationTemplateFailurePayloadInvalid, "prompt template detail query is invalid")
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "prompt_application_templates:read_source", false, "read")
	if failure != "" {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: failure})
		return
	}
	writePromptApplicationTemplateResult(writer, trace, ctx, server.promptApplicationTemplateService().ReadDraft(ctx, request.PathValue("template_id")))
}

func (server *Server) handleCreatePromptApplicationTemplateVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateVersionCreateRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	var body promptApplicationTemplateVersionCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "prompt_application_templates:version", server.config.PromptTemplateDevWriteEnabled, "version")
	if failure != "" {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: failure})
		return
	}
	writePromptApplicationTemplateResult(writer, trace, ctx, server.promptApplicationTemplateService().CreateVersion(ctx, request.PathValue("template_id"), body.SourceDraftVersion))
}

func (server *Server) handleListPromptApplicationTemplateVersions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateVersionListRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	if !promptApplicationTemplateQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, PromptApplicationTemplateFailurePayloadInvalid, "prompt template version list query is invalid")
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "prompt_application_templates:read", false, "version-list")
	if failure != "" {
		writePromptApplicationTemplateVersionListResult(writer, trace, ctx, request.PathValue("template_id"), nil, failure)
		return
	}
	summaries, failure := server.promptApplicationTemplateService().ListVersions(ctx, request.PathValue("template_id"))
	writePromptApplicationTemplateVersionListResult(writer, trace, ctx, request.PathValue("template_id"), summaries, failure)
}

func (server *Server) handleReadPromptApplicationTemplateVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, promptApplicationTemplateVersionReadRoute)
	if !server.allowPromptApplicationTemplateDevHTTP(writer, trace) {
		return
	}
	if !promptApplicationTemplateQueryAllowed(request, "workspace_id", "application_id") {
		server.writePlatformError(writer, trace, PromptApplicationTemplateFailurePayloadInvalid, "prompt template version detail query is invalid")
		return
	}
	ctx, failure := promptApplicationTemplateContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "prompt_application_templates:read_source", false, "version-read")
	if failure != "" {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: failure})
		return
	}
	version, err := strconv.Atoi(strings.TrimSpace(request.PathValue("template_version")))
	if err != nil || version < 1 {
		writePromptApplicationTemplateResult(writer, trace, ctx, PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailurePayloadInvalid})
		return
	}
	writePromptApplicationTemplateResult(writer, trace, ctx, server.promptApplicationTemplateService().ReadVersion(ctx, request.PathValue("template_id"), version))
}

func (server *Server) promptApplicationTemplateService() promptApplicationTemplateService {
	if server.promptApplicationTemplateRepository == nil {
		server.promptApplicationTemplateRepository = &memoryPromptApplicationTemplateRepository{drafts: make(map[string]PromptApplicationTemplateDraft), versions: make(map[string]map[int]PromptApplicationTemplateVersion), unavailable: true}
	}
	service := newPromptApplicationTemplateService(server.promptApplicationTemplateRepository)
	service.requirePromptApplication = func(ctx PromptApplicationTemplateContext) string {
		result := server.applicationCatalogService().RequireActive(ApplicationCatalogContext{
			RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef,
		}, ctx.ApplicationID)
		if result.FailureCode != "" {
			switch result.FailureCode {
			case ApplicationCatalogFailureNotFound:
				return PromptApplicationTemplateFailureApplicationMissing
			case ApplicationCatalogFailureArchived:
				return PromptApplicationTemplateFailureApplicationArchived
			default:
				return PromptApplicationTemplateFailureStoreUnavailable
			}
		}
		if result.Record == nil || result.Record.ApplicationKind != "prompt_application" {
			return PromptApplicationTemplateFailureApplicationKind
		}
		return ""
	}
	return service
}

func promptApplicationTemplateContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, requiredScope string, writeEnabled bool, auditSuffix string) (PromptApplicationTemplateContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	ctx := PromptApplicationTemplateContext{
		RequestContext: request.Context(), RequestID: trace.requestID, WorkspaceID: strings.TrimSpace(workspaceID),
		ApplicationID: strings.TrimSpace(applicationID), WriteEnabled: writeEnabled,
		AuditRef: "audit_" + trace.requestID + "_prompt-template-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" || !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return ctx, PromptApplicationTemplateFailureScopeDenied
	}
	ctx.TenantRef = strings.TrimSpace(auth.TenantBinding)
	ctx.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	ctx.OwnerSubjectRef = ctx.ActorRef
	if ctx.TenantRef == "" || ctx.WorkspaceID == "" || ctx.ApplicationID == "" ||
		strings.TrimSpace(request.Header.Get(promptApplicationTemplateDevWorkspaceHeader)) != ctx.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(promptApplicationTemplateDevApplicationHeader)) != ctx.ApplicationID {
		return ctx, PromptApplicationTemplateFailureScopeDenied
	}
	return ctx, ""
}

func (server *Server) allowPromptApplicationTemplateDevHTTP(writer http.ResponseWriter, trace requestTrace) bool {
	if server.config.PromptTemplateDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "PROMPT_APPLICATION_TEMPLATE_DEV_HTTP_DISABLED", "prompt application template route requires explicit development opt-in")
	return false
}

func promptApplicationTemplateQueryAllowed(request *http.Request, allowed ...string) bool {
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

func writePromptApplicationTemplateResult(writer http.ResponseWriter, trace requestTrace, ctx PromptApplicationTemplateContext, result PromptApplicationTemplateResult) {
	validation := result.ValidationSummary
	if validation.Findings == nil {
		validation.Findings = []PromptApplicationTemplateFinding{}
	}
	if validation.State == "" {
		validation.State = promptApplicationTemplateValidationInvalid
	}
	writeObservedJSON(writer, http.StatusOK, trace, promptApplicationTemplateEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Draft: result.Draft, Version: result.Version, FailureCode: optionalPromptApplicationTemplateFailure(result.FailureCode),
		CurrentDraftVersion: result.CurrentDraftVersion, CurrentTemplateVersion: result.CurrentTemplateVersion,
		ValidationSummary: validation, AuditRef: ctx.AuditRef,
	})
}

func writePromptApplicationTemplateListResult(writer http.ResponseWriter, trace requestTrace, ctx PromptApplicationTemplateContext, summaries []PromptApplicationTemplateDraftSummary, failure string) {
	if summaries == nil {
		summaries = []PromptApplicationTemplateDraftSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, promptApplicationTemplateListEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		DraftSummaries: summaries, FailureCode: optionalPromptApplicationTemplateFailure(failure), AuditRef: ctx.AuditRef,
	})
}

func writePromptApplicationTemplateVersionListResult(writer http.ResponseWriter, trace requestTrace, ctx PromptApplicationTemplateContext, templateID string, summaries []PromptApplicationTemplateVersionSummary, failure string) {
	if summaries == nil {
		summaries = []PromptApplicationTemplateVersionSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, promptApplicationTemplateVersionListEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, TemplateID: strings.TrimSpace(templateID),
		VersionSummaries: summaries, FailureCode: optionalPromptApplicationTemplateFailure(failure), AuditRef: ctx.AuditRef,
	})
}

func optionalPromptApplicationTemplateFailure(failure string) *string {
	failure = strings.TrimSpace(failure)
	if failure == "" {
		return nil
	}
	return &failure
}
