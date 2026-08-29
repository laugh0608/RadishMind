package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowTemplateCandidateCreateRoute   = "POST /v1/user-workspace/workflow-template-candidates"
	workflowTemplateCandidateListRoute     = "GET /v1/user-workspace/workflow-template-candidates"
	workflowTemplateCandidateReadRoute     = "GET /v1/user-workspace/workflow-template-candidates/{candidate_id}"
	workflowTemplateCandidateDecisionRoute = "POST /v1/user-workspace/workflow-template-candidates/{candidate_id}/decisions"
	workflowTemplateListRoute              = "GET /v1/user-workspace/workflow-templates"
	workflowTemplateReadRoute              = "GET /v1/user-workspace/workflow-templates/{template_id}"
	workflowTemplateVersionListRoute       = "GET /v1/user-workspace/workflow-templates/{template_id}/versions"
	workflowTemplateVersionReadRoute       = "GET /v1/user-workspace/workflow-templates/{template_id}/versions/{version}"
	workflowTemplateListingDecisionRoute   = "POST /v1/user-workspace/workflow-templates/{template_id}/listing-decisions"
	workflowTemplateDerivationRoute        = "POST /v1/user-workspace/workflow-templates/{template_id}/derivations"
)

type workflowTemplateCandidateCreateBody struct {
	CandidateID             string   `json:"candidate_id"`
	TemplateID              string   `json:"template_id"`
	SourceApplicationID     string   `json:"source_application_id"`
	SourceDefinitionID      string   `json:"source_definition_id"`
	SourceDefinitionVersion int      `json:"source_definition_version"`
	Title                   string   `json:"title"`
	Summary                 string   `json:"summary"`
	UsageNotes              string   `json:"usage_notes"`
	Labels                  []string `json:"labels"`
}

type workflowTemplateCandidateDecisionBody struct {
	ExpectedReviewVersion int    `json:"expected_review_version"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
}

type workflowTemplateListingDecisionBody struct {
	ExpectedPointerVersion int    `json:"expected_pointer_version"`
	Decision               string `json:"decision"`
	Version                int    `json:"version"`
	Reason                 string `json:"reason"`
}

type workflowTemplateDerivationBody struct {
	ExpectedPointerVersion int    `json:"expected_pointer_version"`
	TemplateVersion        int    `json:"template_version"`
	TargetApplicationID    string `json:"target_application_id"`
	DraftID                string `json:"draft_id"`
	Name                   string `json:"name"`
	Confirmed              bool   `json:"confirmed"`
}

type workflowTemplateCatalogEnvelope struct {
	RequestID             string                      `json:"request_id"`
	WorkspaceID           string                      `json:"workspace_id"`
	Candidate             *WorkflowTemplateCandidate  `json:"candidate"`
	Version               *WorkflowTemplateVersion    `json:"version"`
	Lineage               *WorkflowTemplateLineage    `json:"lineage"`
	Draft                 *savedWorkflowDraftDocument `json:"draft"`
	FailureCode           *string                     `json:"failure_code"`
	CurrentReviewVersion  int                         `json:"current_review_version"`
	CurrentPointerVersion int                         `json:"current_pointer_version"`
	AuditRef              string                      `json:"audit_ref"`
}

type workflowTemplateCandidateListEnvelope struct {
	RequestID   string                      `json:"request_id"`
	WorkspaceID string                      `json:"workspace_id"`
	Candidates  []WorkflowTemplateCandidate `json:"candidates"`
	NextCursor  string                      `json:"next_cursor"`
	FailureCode *string                     `json:"failure_code"`
	AuditRef    string                      `json:"audit_ref"`
}

type workflowTemplateLineageListEnvelope struct {
	RequestID   string                    `json:"request_id"`
	WorkspaceID string                    `json:"workspace_id"`
	Templates   []WorkflowTemplateLineage `json:"templates"`
	NextCursor  string                    `json:"next_cursor"`
	FailureCode *string                   `json:"failure_code"`
	AuditRef    string                    `json:"audit_ref"`
}

type workflowTemplateVersionListEnvelope struct {
	RequestID   string                    `json:"request_id"`
	WorkspaceID string                    `json:"workspace_id"`
	TemplateID  string                    `json:"template_id"`
	Versions    []WorkflowTemplateVersion `json:"versions"`
	NextCursor  string                    `json:"next_cursor"`
	FailureCode *string                   `json:"failure_code"`
	AuditRef    string                    `json:"audit_ref"`
}

func (server *Server) handleCreateWorkflowTemplateCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateCandidateCreateRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateMutationContext(request, trace, "candidate-create", "workflow_definitions:read", "workflow_definitions:write")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	var body workflowTemplateCandidateCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	result := server.workflowTemplateCatalogService().CreateCandidate(ctx, WorkflowTemplateCandidateCreateInput{
		CandidateID: body.CandidateID, TemplateID: body.TemplateID, SourceApplicationID: body.SourceApplicationID,
		SourceDefinitionID: body.SourceDefinitionID, SourceDefinitionVersion: body.SourceDefinitionVersion,
		Title: body.Title, Summary: body.Summary, UsageNotes: body.UsageNotes, Labels: body.Labels,
	})
	writeWorkflowTemplateResult(writer, trace, ctx, result)
}

func (server *Server) handleListWorkflowTemplateCandidates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateCandidateListRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "candidate-list", []string{"state", "limit", "cursor"}, "workflow_definitions:read")
	if failure != "" {
		writeWorkflowTemplateCandidateListWithStatus(writer, status, trace, ctx, WorkflowTemplateCandidateListResult{FailureCode: failure})
		return
	}
	limit, ok := workflowTemplateHTTPListLimit(request.URL.Query().Get("limit"))
	if !ok {
		writeWorkflowTemplateCandidateList(writer, trace, ctx, WorkflowTemplateCandidateListResult{FailureCode: WorkflowTemplateFailurePayloadInvalid})
		return
	}
	writeWorkflowTemplateCandidateList(writer, trace, ctx, server.workflowTemplateCatalogService().ListCandidates(ctx, WorkflowTemplateListInput{
		State: request.URL.Query().Get("state"), Limit: limit, Cursor: request.URL.Query().Get("cursor"),
	}))
}

func (server *Server) handleReadWorkflowTemplateCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateCandidateReadRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "candidate-read", nil, "workflow_definitions:read")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().ReadCandidate(ctx, request.PathValue("candidate_id")))
}

func (server *Server) handleDecideWorkflowTemplateCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateCandidateDecisionRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateMutationContext(request, trace, "candidate-decision", "workflow_definitions:read", "workflow_definitions:review")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	var body workflowTemplateCandidateDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().ReviewCandidate(ctx, request.PathValue("candidate_id"), WorkflowTemplateReviewInput{
		ExpectedReviewVersion: body.ExpectedReviewVersion, Decision: body.Decision, Reason: body.Reason,
	}))
}

func (server *Server) handleListWorkflowTemplates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateListRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "template-list", []string{"limit", "cursor"}, "workflow_definitions:read")
	if failure != "" {
		writeWorkflowTemplateLineageListWithStatus(writer, status, trace, ctx, WorkflowTemplateLineageListResult{FailureCode: failure})
		return
	}
	limit, ok := workflowTemplateHTTPListLimit(request.URL.Query().Get("limit"))
	if !ok {
		writeWorkflowTemplateLineageList(writer, trace, ctx, WorkflowTemplateLineageListResult{FailureCode: WorkflowTemplateFailurePayloadInvalid})
		return
	}
	writeWorkflowTemplateLineageList(writer, trace, ctx, server.workflowTemplateCatalogService().ListTemplates(ctx, WorkflowTemplateListInput{Limit: limit, Cursor: request.URL.Query().Get("cursor")}))
}

func (server *Server) handleReadWorkflowTemplate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateReadRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "template-read", nil, "workflow_definitions:read")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().ReadTemplate(ctx, request.PathValue("template_id")))
}

func (server *Server) handleListWorkflowTemplateVersions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateVersionListRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "version-list", []string{"limit", "cursor"}, "workflow_definitions:read")
	templateID := request.PathValue("template_id")
	if failure != "" {
		writeWorkflowTemplateVersionListWithStatus(writer, status, trace, ctx, templateID, WorkflowTemplateVersionListResult{FailureCode: failure})
		return
	}
	limit, ok := workflowTemplateHTTPListLimit(request.URL.Query().Get("limit"))
	if !ok {
		writeWorkflowTemplateVersionList(writer, trace, ctx, templateID, WorkflowTemplateVersionListResult{FailureCode: WorkflowTemplateFailurePayloadInvalid})
		return
	}
	writeWorkflowTemplateVersionList(writer, trace, ctx, templateID, server.workflowTemplateCatalogService().ListVersions(ctx, templateID, WorkflowTemplateListInput{Limit: limit, Cursor: request.URL.Query().Get("cursor")}))
}

func (server *Server) handleReadWorkflowTemplateVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateVersionReadRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateReadContext(request, trace, "version-read", nil, "workflow_definitions:read")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	version, err := strconv.Atoi(request.PathValue("version"))
	if err != nil {
		writeWorkflowTemplateResult(writer, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: WorkflowTemplateFailurePayloadInvalid})
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().ReadVersion(ctx, request.PathValue("template_id"), version))
}

func (server *Server) handleDecideWorkflowTemplateListing(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateListingDecisionRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateMutationContext(request, trace, "listing-decision", "workflow_definitions:read", "workflow_definitions:activate")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	var body workflowTemplateListingDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().DecideListing(ctx, request.PathValue("template_id"), WorkflowTemplateListingInput{
		ExpectedPointerVersion: body.ExpectedPointerVersion, Decision: body.Decision, Version: body.Version, Reason: body.Reason,
	}))
}

func (server *Server) handleDeriveWorkflowTemplate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowTemplateDerivationRoute)
	if !server.allowWorkflowTemplateCatalogHTTP(writer, request, trace) {
		return
	}
	ctx, failure, status := server.workflowTemplateMutationContext(request, trace, "derive", "workflow_definitions:read", "workflow_drafts:write")
	if failure != "" {
		writeWorkflowTemplateResultWithStatus(writer, status, trace, ctx, WorkflowTemplateCatalogResult{FailureCode: failure})
		return
	}
	var body workflowTemplateDerivationBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true, rejectDuplicateFields: true}) {
		return
	}
	writeWorkflowTemplateResult(writer, trace, ctx, server.workflowTemplateCatalogService().Derive(ctx, request.PathValue("template_id"), WorkflowTemplateDerivationInput{
		ExpectedPointerVersion: body.ExpectedPointerVersion, TemplateVersion: body.TemplateVersion,
		TargetApplicationID: body.TargetApplicationID, DraftID: body.DraftID, Name: body.Name, Confirmed: body.Confirmed,
	}))
}

func (server *Server) allowWorkflowTemplateCatalogHTTP(writer http.ResponseWriter, _ *http.Request, trace requestTrace) bool {
	if server.config.WorkflowTemplateCatalogDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_TEMPLATE_CATALOG_DEV_HTTP_DISABLED", "workflow template catalog route requires explicit development opt-in")
	return false
}

func (server *Server) workflowTemplateContext(request *http.Request, trace requestTrace, auditSuffix string, permissions ...string) (WorkflowTemplateCatalogContext, string, int) {
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, permissions...)
	ctx := WorkflowTemplateCatalogContext{
		RequestContext: request.Context(),
		TenantRef:      strings.TrimSpace(auth.TenantBinding), WorkspaceID: strings.TrimSpace(auth.ResourceBinding.WorkspaceID),
		OwnerSubjectRef: strings.TrimSpace(auth.SubjectBinding), ActorRef: strings.TrimSpace(auth.SubjectBinding),
		RequestID: trace.requestID, AuditRef: "audit_" + trace.requestID + "_workflow-template-" + auditSuffix,
	}
	if failure != "" || !validWorkflowTemplateContext(ctx) {
		return ctx, WorkflowTemplateFailureScopeDenied, status
	}
	return ctx, "", http.StatusOK
}

func (server *Server) workflowTemplateMutationContext(request *http.Request, trace requestTrace, auditSuffix string, permissions ...string) (WorkflowTemplateCatalogContext, string, int) {
	ctx, failure, status := server.workflowTemplateContext(request, trace, auditSuffix, permissions...)
	if failure != "" {
		return ctx, failure, status
	}
	if !workflowTemplateStrictQueryAllowed(request.URL.Query()) {
		return ctx, WorkflowTemplateFailurePayloadInvalid, http.StatusBadRequest
	}
	return ctx, "", http.StatusOK
}

func (server *Server) workflowTemplateReadContext(request *http.Request, trace requestTrace, auditSuffix string, optionalQueryKeys []string, permissions ...string) (WorkflowTemplateCatalogContext, string, int) {
	ctx, failure, status := server.workflowTemplateContext(request, trace, auditSuffix, permissions...)
	if failure != "" {
		return ctx, failure, status
	}
	allowed := append([]string{"workspace_id"}, optionalQueryKeys...)
	if !workflowTemplateStrictQueryAllowed(request.URL.Query(), allowed...) {
		return ctx, WorkflowTemplateFailurePayloadInvalid, http.StatusBadRequest
	}
	if strings.TrimSpace(request.URL.Query().Get("workspace_id")) != ctx.WorkspaceID {
		return ctx, WorkflowTemplateFailureScopeDenied, http.StatusForbidden
	}
	return ctx, "", http.StatusOK
}

func workflowTemplateStrictQueryAllowed(values map[string][]string, allowed ...string) bool {
	allowlist := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		allowlist[key] = true
	}
	for key, entries := range values {
		if !allowlist[key] || len(entries) != 1 {
			return false
		}
	}
	return true
}

func (server *Server) workflowTemplateCatalogService() workflowTemplateCatalogService {
	service := newWorkflowTemplateCatalogService(server.workflowTemplateCatalogRepository, server.workflowDefinitionReleaseRepository, server.applicationCatalogRepository, server.savedWorkflowDraftStore)
	if server.workflowTemplateTargetBindingValidator != nil {
		service.targetBinding = server.workflowTemplateTargetBindingValidator
	}
	return service
}

func workflowTemplateHTTPListLimit(value string) (int, bool) {
	if strings.TrimSpace(value) == "" {
		return 0, true
	}
	limit, err := strconv.Atoi(value)
	return limit, err == nil
}

func writeWorkflowTemplateResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateCatalogResult) {
	writeWorkflowTemplateResultWithStatus(writer, http.StatusOK, trace, ctx, result)
}

func writeWorkflowTemplateResultWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateCatalogResult) {
	writeObservedJSON(writer, status, trace, workflowTemplateCatalogEnvelope{
		RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, Candidate: result.Candidate, Version: result.Version,
		Lineage: result.Lineage, Draft: savedWorkflowDraftDocumentPointer(result.Draft), FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentReviewVersion: result.CurrentReviewVersion, CurrentPointerVersion: result.CurrentPointerVersion, AuditRef: ctx.AuditRef,
	})
}

func writeWorkflowTemplateCandidateList(writer http.ResponseWriter, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateCandidateListResult) {
	writeWorkflowTemplateCandidateListWithStatus(writer, http.StatusOK, trace, ctx, result)
}

func writeWorkflowTemplateCandidateListWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateCandidateListResult) {
	if result.Candidates == nil {
		result.Candidates = []WorkflowTemplateCandidate{}
	}
	writeObservedJSON(writer, status, trace, workflowTemplateCandidateListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, Candidates: result.Candidates, NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef})
}

func writeWorkflowTemplateLineageList(writer http.ResponseWriter, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateLineageListResult) {
	writeWorkflowTemplateLineageListWithStatus(writer, http.StatusOK, trace, ctx, result)
}

func writeWorkflowTemplateLineageListWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx WorkflowTemplateCatalogContext, result WorkflowTemplateLineageListResult) {
	if result.Lineages == nil {
		result.Lineages = []WorkflowTemplateLineage{}
	}
	writeObservedJSON(writer, status, trace, workflowTemplateLineageListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, Templates: result.Lineages, NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef})
}

func writeWorkflowTemplateVersionList(writer http.ResponseWriter, trace requestTrace, ctx WorkflowTemplateCatalogContext, templateID string, result WorkflowTemplateVersionListResult) {
	writeWorkflowTemplateVersionListWithStatus(writer, http.StatusOK, trace, ctx, templateID, result)
}

func writeWorkflowTemplateVersionListWithStatus(writer http.ResponseWriter, status int, trace requestTrace, ctx WorkflowTemplateCatalogContext, templateID string, result WorkflowTemplateVersionListResult) {
	if result.Versions == nil {
		result.Versions = []WorkflowTemplateVersion{}
	}
	writeObservedJSON(writer, status, trace, workflowTemplateVersionListEnvelope{RequestID: trace.requestID, WorkspaceID: ctx.WorkspaceID, TemplateID: templateID, Versions: result.Versions, NextCursor: result.NextCursor, FailureCode: optionalApplicationDraftFailure(result.FailureCode), AuditRef: ctx.AuditRef})
}
