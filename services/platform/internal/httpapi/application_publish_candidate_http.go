package httpapi

import (
	"net/http"
	"strings"
)

const (
	applicationPublishCandidateCreateRoute = "POST /v1/user-workspace/application-publish-candidates"
	applicationPublishCandidateListRoute   = "GET /v1/user-workspace/application-publish-candidates"
	applicationPublishCandidateReadRoute   = "GET /v1/user-workspace/application-publish-candidates/{candidate_id}"
	applicationPublishCandidateReviewRoute = "POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews"

	applicationPublishDevWorkspaceHeader   = "X-RadishMind-Dev-Application-Publish-Workspace"
	applicationPublishDevApplicationHeader = "X-RadishMind-Dev-Application-Publish-Application"
)

type applicationPublishCandidateCreateBody struct {
	CandidateID          string   `json:"candidate_id"`
	DraftID              string   `json:"draft_id"`
	ExpectedDraftVersion int      `json:"expected_draft_version"`
	EvidenceRequestIDs   []string `json:"evidence_request_ids"`
}

type applicationPublishCandidateReviewBody struct {
	ExpectedReviewVersion int    `json:"expected_review_version"`
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
}

type applicationPublishCandidateEnvelope struct {
	RequestID             string                       `json:"request_id"`
	WorkspaceID           string                       `json:"workspace_id"`
	ApplicationID         string                       `json:"application_id"`
	Candidate             *ApplicationPublishCandidate `json:"candidate"`
	FailureCode           *string                      `json:"failure_code"`
	CurrentReviewVersion  int                          `json:"current_review_version"`
	CurrentCandidateState string                       `json:"current_candidate_state"`
	CurrentDraftVersion   int                          `json:"current_draft_version"`
	AuditRef              string                       `json:"audit_ref"`
}

type applicationPublishCandidateListEnvelope struct {
	RequestID          string                               `json:"request_id"`
	WorkspaceID        string                               `json:"workspace_id"`
	ApplicationID      string                               `json:"application_id"`
	CandidateSummaries []ApplicationPublishCandidateSummary `json:"candidate_summaries"`
	FailureCode        *string                              `json:"failure_code"`
	AuditRef           string                               `json:"audit_ref"`
}

func (server *Server) handleCreateApplicationPublishCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationPublishCandidateCreateRoute)
	if !server.allowApplicationPublishDevHTTP(writer, request, trace) {
		return
	}
	var body applicationPublishCandidateCreateBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode := applicationPublishContextFromRequest(request, trace, request.Header.Get(applicationPublishDevWorkspaceHeader), request.Header.Get(applicationPublishDevApplicationHeader), "application_publish_candidates:write", server.config.ApplicationPublishDevWriteEnabled, "create")
	if failureCode != "" {
		writeApplicationPublishCandidateResult(writer, trace, requestContext, ApplicationPublishResult{FailureCode: failureCode})
		return
	}
	result := server.applicationPublishCandidateService().Create(requestContext, ApplicationPublishCreateInput{
		CandidateID: body.CandidateID, DraftID: body.DraftID, ExpectedDraftVersion: body.ExpectedDraftVersion,
		EvidenceRequestIDs: body.EvidenceRequestIDs,
	})
	writeApplicationPublishCandidateResult(writer, trace, requestContext, result)
}

func (server *Server) handleListApplicationPublishCandidates(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationPublishCandidateListRoute)
	if !server.allowApplicationPublishDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	requestContext, failureCode := applicationPublishContextFromRequest(request, trace, workspaceID, applicationID, "application_publish_candidates:read", false, "list")
	if failureCode != "" {
		writeApplicationPublishCandidateListResult(writer, trace, requestContext, nil, failureCode)
		return
	}
	summaries, failureCode := server.applicationPublishCandidateService().List(requestContext)
	writeApplicationPublishCandidateListResult(writer, trace, requestContext, summaries, failureCode)
}

func (server *Server) handleReadApplicationPublishCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationPublishCandidateReadRoute)
	if !server.allowApplicationPublishDevHTTP(writer, request, trace) {
		return
	}
	workspaceID := strings.TrimSpace(request.URL.Query().Get("workspace_id"))
	applicationID := strings.TrimSpace(request.URL.Query().Get("application_id"))
	requestContext, failureCode := applicationPublishContextFromRequest(request, trace, workspaceID, applicationID, "application_publish_candidates:read", false, "read")
	if failureCode != "" {
		writeApplicationPublishCandidateResult(writer, trace, requestContext, ApplicationPublishResult{FailureCode: failureCode})
		return
	}
	writeApplicationPublishCandidateResult(writer, trace, requestContext, server.applicationPublishCandidateService().Read(requestContext, request.PathValue("candidate_id")))
}

func (server *Server) handleReviewApplicationPublishCandidate(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, applicationPublishCandidateReviewRoute)
	if !server.allowApplicationPublishDevHTTP(writer, request, trace) {
		return
	}
	var body applicationPublishCandidateReviewBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	requestContext, failureCode := applicationPublishContextFromRequest(request, trace, request.Header.Get(applicationPublishDevWorkspaceHeader), request.Header.Get(applicationPublishDevApplicationHeader), "application_publish_candidates:review", server.config.ApplicationPublishDevWriteEnabled, "review")
	if failureCode != "" {
		writeApplicationPublishCandidateResult(writer, trace, requestContext, ApplicationPublishResult{FailureCode: failureCode})
		return
	}
	result := server.applicationPublishCandidateService().Review(requestContext, request.PathValue("candidate_id"), ApplicationPublishReviewInput{
		ExpectedReviewVersion: body.ExpectedReviewVersion, Decision: body.Decision, Reason: body.Reason,
	})
	writeApplicationPublishCandidateResult(writer, trace, requestContext, result)
}

func (server *Server) allowApplicationPublishDevHTTP(writer http.ResponseWriter, request *http.Request, trace requestTrace) bool {
	if server.config.ApplicationPublishDevHTTPEnabled {
		return true
	}
	server.writePlatformError(writer, trace, "APPLICATION_PUBLISH_DEV_HTTP_DISABLED", "application publish candidate dev route requires explicit opt-in")
	return false
}

func (server *Server) applicationPublishCandidateService() applicationPublishCandidateService {
	if server.applicationPublishCandidateRepository == nil {
		server.applicationPublishCandidateRepository = &memoryApplicationPublishCandidateRepository{candidates: make(map[string]ApplicationPublishCandidate), unavailable: true}
	}
	service := newApplicationPublishCandidateService(server.applicationDraftRepository, server.applicationPublishCandidateRepository, server.readApplicationPublishBaseline)
	service.validateBinding = func(publishContext ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return server.workflowRAGPromotionService().resolveEligibleBinding(workflowRAGPromotionContextFromPublish(publishContext), ref, false)
	}
	service.readPromptTemplateVersion = func(publishContext ApplicationPublishContext, ref PromptApplicationTemplateRef) (PromptApplicationTemplateVersion, string) {
		if server.promptApplicationTemplateRepository == nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureStoreUnavailable
		}
		templateContext := PromptApplicationTemplateContext{
			RequestContext: publishContext.RequestContext, RequestID: publishContext.RequestID, TenantRef: publishContext.TenantRef,
			WorkspaceID: publishContext.WorkspaceID, ApplicationID: publishContext.ApplicationID, ActorRef: publishContext.ActorRef,
			OwnerSubjectRef: publishContext.OwnerSubjectRef, AuditRef: publishContext.AuditRef,
		}
		version, err := server.promptApplicationTemplateRepository.ReadVersion(templateContext, ref.TemplateID, ref.TemplateVersion)
		if err != nil {
			return PromptApplicationTemplateVersion{}, promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{}).FailureCode
		}
		if validateStoredPromptApplicationTemplateVersion(templateContext, version) != nil {
			return PromptApplicationTemplateVersion{}, PromptApplicationTemplateFailureStoreContract
		}
		return version, ""
	}
	return service
}

func (server *Server) readApplicationPublishBaseline(requestContext ApplicationPublishContext) (ApplicationSummary, error) {
	if server.config.ApplicationCatalogDevHTTPEnabled {
		result := server.applicationCatalogService().RequireActive(ApplicationCatalogContext{
			RequestContext: requestContext.RequestContext, RequestID: requestContext.RequestID, TenantRef: requestContext.TenantRef,
			WorkspaceID: requestContext.WorkspaceID, ActorRef: requestContext.ActorRef, OwnerSubjectRef: requestContext.OwnerSubjectRef,
			AuditRef: requestContext.AuditRef,
		}, requestContext.ApplicationID)
		if result.FailureCode == ApplicationCatalogFailureArchived {
			return ApplicationSummary{}, errApplicationCatalogArchived
		}
		if result.FailureCode == ApplicationCatalogFailureNotFound {
			return ApplicationSummary{}, errApplicationPublishBaselineNotFound
		}
		if result.FailureCode != "" || result.Record == nil {
			return ApplicationSummary{}, errApplicationPublishBaselineUnavailable
		}
		return ApplicationSummary{
			ApplicationRef: result.Record.ApplicationID, TenantRef: result.Record.TenantRef,
			ApplicationKind: result.Record.ApplicationKind, DisplayName: result.Record.DisplayName,
			OwnerSubjectRef: result.Record.OwnerSubjectRef, LastRunStatus: "not_available", UpdatedAt: result.Record.UpdatedAt,
		}, nil
	}
	result := server.controlPlaneReadRepository().ListApplicationSummaries(ReadRepositoryContext{
		RequestID: requestContext.RequestID, TenantRef: requestContext.TenantRef, SubjectRef: requestContext.ActorRef,
		ScopeGrants: []string{"applications:read"}, AuditRef: requestContext.AuditRef,
	}, ListApplicationSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{Limit: 200}})
	if result.FailureCode != "" {
		return ApplicationSummary{}, errApplicationPublishBaselineUnavailable
	}
	for _, application := range result.Items {
		if application.ApplicationRef == requestContext.ApplicationID {
			return application, nil
		}
	}
	return ApplicationSummary{}, errApplicationPublishBaselineNotFound
}

func applicationPublishContextFromRequest(request *http.Request, trace requestTrace, workspaceID, applicationID, requiredScope string, writeEnabled bool, auditSuffix string) (ApplicationPublishContext, string) {
	auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
	requestContext := ApplicationPublishContext{
		RequestContext: request.Context(), RequestID: trace.requestID, WorkspaceID: strings.TrimSpace(workspaceID),
		ApplicationID: strings.TrimSpace(applicationID), WriteEnabled: writeEnabled,
		AuditRef: "audit_" + trace.requestID + "_application-publish-" + auditSuffix,
	}
	if !ok || strings.TrimSpace(auth.IdentityContext) == "" || strings.TrimSpace(auth.SubjectBinding) == "" || !controlPlaneReadHasScope(auth.ScopeGrants, requiredScope) {
		return requestContext, ApplicationPublishFailureScopeDenied
	}
	requestContext.TenantRef = strings.TrimSpace(auth.TenantBinding)
	requestContext.ActorRef = strings.TrimSpace(auth.SubjectBinding)
	requestContext.OwnerSubjectRef = requestContext.ActorRef
	requestContext.RAGPromotionReadEnabled = controlPlaneReadHasScope(auth.ScopeGrants, "workflow_rag_promotions:read")
	requestContext.PromptTemplateSourceReadEnabled = controlPlaneReadHasScope(auth.ScopeGrants, "prompt_application_templates:read_source")
	if requestContext.TenantRef == "" || requestContext.WorkspaceID == "" || requestContext.ApplicationID == "" ||
		strings.TrimSpace(request.Header.Get(applicationPublishDevWorkspaceHeader)) != requestContext.WorkspaceID ||
		strings.TrimSpace(request.Header.Get(applicationPublishDevApplicationHeader)) != requestContext.ApplicationID {
		return requestContext, ApplicationPublishFailureScopeDenied
	}
	return requestContext, ""
}

func workflowRAGPromotionContextFromPublish(ctx ApplicationPublishContext) WorkflowRAGPromotionContext {
	return WorkflowRAGPromotionContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled,
	}
}

func writeApplicationPublishCandidateResult(writer http.ResponseWriter, trace requestTrace, requestContext ApplicationPublishContext, result ApplicationPublishResult) {
	writeObservedJSON(writer, http.StatusOK, trace, applicationPublishCandidateEnvelope{
		RequestID: trace.requestID, WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		Candidate: result.Candidate, FailureCode: optionalApplicationDraftFailure(result.FailureCode),
		CurrentReviewVersion: result.CurrentReviewVersion, CurrentCandidateState: result.CurrentCandidateState,
		CurrentDraftVersion: result.CurrentDraftVersion, AuditRef: requestContext.AuditRef,
	})
}

func writeApplicationPublishCandidateListResult(writer http.ResponseWriter, trace requestTrace, requestContext ApplicationPublishContext, summaries []ApplicationPublishCandidateSummary, failureCode string) {
	if summaries == nil {
		summaries = []ApplicationPublishCandidateSummary{}
	}
	writeObservedJSON(writer, http.StatusOK, trace, applicationPublishCandidateListEnvelope{
		RequestID: trace.requestID, WorkspaceID: requestContext.WorkspaceID, ApplicationID: requestContext.ApplicationID,
		CandidateSummaries: summaries, FailureCode: optionalApplicationDraftFailure(failureCode), AuditRef: requestContext.AuditRef,
	})
}
