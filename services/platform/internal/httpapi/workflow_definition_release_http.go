package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	workflowDefinitionCandidateCreateRoute    = "POST /v1/user-workspace/workflow-definition-candidates"
	workflowDefinitionCandidateListRoute      = "GET /v1/user-workspace/workflow-definition-candidates"
	workflowDefinitionCandidateReadRoute      = "GET /v1/user-workspace/workflow-definition-candidates/{candidate_id}"
	workflowDefinitionCandidateDecisionRoute  = "POST /v1/user-workspace/workflow-definition-candidates/{candidate_id}/decisions"
	workflowDefinitionVersionListRoute        = "GET /v1/user-workspace/workflow-definitions/{definition_id}/versions"
	workflowDefinitionVersionReadRoute        = "GET /v1/user-workspace/workflow-definitions/{definition_id}/versions/{version}"
	workflowDefinitionActivationReadRoute     = "GET /v1/user-workspace/workflow-definitions/{definition_id}/activation"
	workflowDefinitionActivationDecisionRoute = "POST /v1/user-workspace/workflow-definitions/{definition_id}/activation-decisions"
)

type workflowDefinitionCandidateCreateBody struct {
	CandidateID          string `json:"candidate_id"`
	DefinitionID         string `json:"definition_id"`
	DraftID              string `json:"draft_id"`
	ExpectedDraftVersion int    `json:"expected_draft_version"`
}

type workflowDefinitionCandidateDecisionBody struct {
	ExpectedReviewVersion int    `json:"expected_review_version"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
}

type workflowDefinitionActivationDecisionBody struct {
	ExpectedPointerVersion int    `json:"expected_pointer_version"`
	Decision               string `json:"decision"`
	Version                int    `json:"version"`
	Reason                 string `json:"reason"`
}

type workflowDefinitionReleaseEnvelope struct {
	RequestID             string                              `json:"request_id"`
	WorkspaceID           string                              `json:"workspace_id"`
	ApplicationID         string                              `json:"application_id"`
	Candidate             *WorkflowDefinitionReleaseCandidate `json:"candidate"`
	Version               *WorkflowDefinitionVersion          `json:"version"`
	Activation            *WorkflowDefinitionActivation       `json:"activation"`
	FailureCode           *string                             `json:"failure_code"`
	CurrentReviewVersion  int                                 `json:"current_review_version"`
	CurrentPointerVersion int                                 `json:"current_pointer_version"`
	AuditRef              string                              `json:"audit_ref"`
}

type workflowDefinitionCandidateListEnvelope struct {
	RequestID     string                               `json:"request_id"`
	WorkspaceID   string                               `json:"workspace_id"`
	ApplicationID string                               `json:"application_id"`
	Candidates    []WorkflowDefinitionReleaseCandidate `json:"candidates"`
	FailureCode   *string                              `json:"failure_code"`
	AuditRef      string                               `json:"audit_ref"`
}

type workflowDefinitionVersionListEnvelope struct {
	RequestID     string                      `json:"request_id"`
	WorkspaceID   string                      `json:"workspace_id"`
	ApplicationID string                      `json:"application_id"`
	DefinitionID  string                      `json:"definition_id"`
	Versions      []WorkflowDefinitionVersion `json:"versions"`
	FailureCode   *string                     `json:"failure_code"`
	AuditRef      string                      `json:"audit_ref"`
}

func (server *Server) handleCreateWorkflowDefinitionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionCandidateCreateRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	var body workflowDefinitionCandidateCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.Header.Get(savedWorkflowDraftDevWorkspaceHeader), request.Header.Get(savedWorkflowDraftDevApplicationHeader), "workflow_definitions:write", "candidate-create")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	result := server.workflowDefinitionReleaseService().Create(ctx, WorkflowDefinitionCandidateCreateInput{
		CandidateID:          body.CandidateID,
		DefinitionID:         body.DefinitionID,
		DraftID:              body.DraftID,
		ExpectedDraftVersion: body.ExpectedDraftVersion,
	})
	writeWorkflowDefinitionResult(writer, trace, ctx, result)
}

func (server *Server) handleListWorkflowDefinitionCandidates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionCandidateListRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "workflow_definitions:read", "candidate-list")
	if failure != "" {
		writeWorkflowDefinitionCandidateList(writer, trace, ctx, nil, failure)
		return
	}
	values, failure := server.workflowDefinitionReleaseService().ListCandidates(ctx)
	writeWorkflowDefinitionCandidateList(writer, trace, ctx, values, failure)
}

func (server *Server) handleReadWorkflowDefinitionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionCandidateReadRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "workflow_definitions:read", "candidate-read")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	writeWorkflowDefinitionResult(writer, trace, ctx, server.workflowDefinitionReleaseService().ReadCandidate(ctx, request.PathValue("candidate_id")))
}

func (server *Server) handleDecideWorkflowDefinitionCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionCandidateDecisionRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	var body workflowDefinitionCandidateDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.Header.Get(savedWorkflowDraftDevWorkspaceHeader), request.Header.Get(savedWorkflowDraftDevApplicationHeader), "workflow_definitions:review", "candidate-decision")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	result := server.workflowDefinitionReleaseService().Review(ctx, request.PathValue("candidate_id"), WorkflowDefinitionReviewInput{
		ExpectedReviewVersion: body.ExpectedReviewVersion,
		Decision:              body.Decision,
		Reason:                body.Reason,
	})
	writeWorkflowDefinitionResult(writer, trace, ctx, result)
}

func (server *Server) handleListWorkflowDefinitionVersions(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionVersionListRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "workflow_definitions:read", "version-list")
	definitionID := request.PathValue("definition_id")
	if failure != "" {
		writeWorkflowDefinitionVersionList(writer, trace, ctx, definitionID, nil, failure)
		return
	}
	values, failure := server.workflowDefinitionReleaseService().ListVersions(ctx, definitionID)
	writeWorkflowDefinitionVersionList(writer, trace, ctx, definitionID, values, failure)
}

func (server *Server) handleReadWorkflowDefinitionVersion(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionVersionReadRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "workflow_definitions:read", "version-read")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	version, err := strconv.Atoi(request.PathValue("version"))
	if err != nil {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: workflowDefinitionFailurePayloadInvalid})
		return
	}
	writeWorkflowDefinitionResult(writer, trace, ctx, server.workflowDefinitionReleaseService().ReadVersion(ctx, request.PathValue("definition_id"), version))
}

func (server *Server) handleReadWorkflowDefinitionActivation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionActivationReadRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.URL.Query().Get("workspace_id"), request.URL.Query().Get("application_id"), "workflow_definitions:read", "activation-read")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	writeWorkflowDefinitionResult(writer, trace, ctx, server.workflowDefinitionReleaseService().ReadActivation(ctx, request.PathValue("definition_id")))
}

func (server *Server) handleDecideWorkflowDefinitionActivation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionActivationDecisionRoute)
	if !server.allowWorkflowDefinitionReleaseHTTP(writer, request, trace) {
		return
	}
	var body workflowDefinitionActivationDecisionBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	ctx, failure := workflowDefinitionContextFromRequest(request, trace, request.Header.Get(savedWorkflowDraftDevWorkspaceHeader), request.Header.Get(savedWorkflowDraftDevApplicationHeader), "workflow_definitions:activate", "activation-decision")
	if failure != "" {
		writeWorkflowDefinitionResult(writer, trace, ctx, WorkflowDefinitionReleaseResult{FailureCode: failure})
		return
	}
	result := server.workflowDefinitionReleaseService().DecideActivation(ctx, request.PathValue("definition_id"), WorkflowDefinitionActivationInput{
		ExpectedPointerVersion: body.ExpectedPointerVersion,
		Decision:               body.Decision,
		Version:                body.Version,
		Reason:                 body.Reason,
	})
	writeWorkflowDefinitionResult(writer, trace, ctx, result)
}

func (server *Server) allowWorkflowDefinitionReleaseHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if server.config.WorkflowDefinitionReleaseDevEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "WORKFLOW_DEFINITION_RELEASE_DEV_HTTP_DISABLED", "workflow definition release route requires explicit development opt-in")
	return false
}

func (server *Server) workflowDefinitionReleaseService() workflowDefinitionReleaseService {
	return newWorkflowDefinitionReleaseService(server.savedWorkflowDraftStore, server.workflowDefinitionReleaseRepository)
}

func workflowDefinitionContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, requiredScope, auditSuffix string) (WorkflowDefinitionReleaseContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	ctx := WorkflowDefinitionReleaseContext{
		RequestContext: request.Context(),
		WorkspaceID:    strings.TrimSpace(workspaceID),
		ApplicationID:  strings.TrimSpace(applicationID),
		RequestID:      trace.requestID,
		AuditRef:       "audit_" + trace.requestID + "_workflow-definition-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" || !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return ctx, workflowDefinitionFailureScopeDenied
	}
	ctx.TenantRef = strings.TrimSpace(auth.TenantBinding)
	ctx.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	ctx.OwnerSubjectRef = ctx.ActorRef
	if ctx.TenantRef == "" || ctx.WorkspaceID == "" || ctx.ApplicationID == "" ||
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevWorkspaceHeader)) != ctx.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(savedWorkflowDraftDevApplicationHeader)) != ctx.ApplicationID {
		return ctx, workflowDefinitionFailureScopeDenied
	}
	return ctx, ""
}

func writeWorkflowDefinitionResult(writer http.ResponseWriter, trace requestTrace, ctx WorkflowDefinitionReleaseContext, result WorkflowDefinitionReleaseResult) {
	writeObservedJSON(writer, http.StatusOK, trace, workflowDefinitionReleaseEnvelope{
		RequestID:             trace.requestID,
		WorkspaceID:           ctx.WorkspaceID,
		ApplicationID:         ctx.ApplicationID,
		Candidate:             result.Candidate,
		Version:               result.Version,
		Activation:            result.Activation,
		FailureCode:           optionalApplicationDraftFailure(result.FailureCode),
		CurrentReviewVersion:  result.CurrentReviewVersion,
		CurrentPointerVersion: result.CurrentPointerVersion,
		AuditRef:              ctx.AuditRef,
	})
}

func writeWorkflowDefinitionCandidateList(writer http.ResponseWriter, trace requestTrace, ctx WorkflowDefinitionReleaseContext, values []WorkflowDefinitionReleaseCandidate, failure string) {
	if values == nil {
		values = []WorkflowDefinitionReleaseCandidate{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowDefinitionCandidateListEnvelope{
		RequestID:     trace.requestID,
		WorkspaceID:   ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID,
		Candidates:    values,
		FailureCode:   optionalApplicationDraftFailure(failure),
		AuditRef:      ctx.AuditRef,
	})
}

func writeWorkflowDefinitionVersionList(writer http.ResponseWriter, trace requestTrace, ctx WorkflowDefinitionReleaseContext, definitionID string, values []WorkflowDefinitionVersion, failure string) {
	if values == nil {
		values = []WorkflowDefinitionVersion{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, workflowDefinitionVersionListEnvelope{
		RequestID:     trace.requestID,
		WorkspaceID:   ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID,
		DefinitionID:  definitionID,
		Versions:      values,
		FailureCode:   optionalApplicationDraftFailure(failure),
		AuditRef:      ctx.AuditRef,
	})
}
