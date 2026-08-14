package httpapi

import (
	"context"
	"net/http"
	"strings"
)

type promptApplicationInvocationBody struct {
	Variables           map[string]any `json:"variables"`
	ClientInvocationKey string         `json:"client_invocation_key"`
}

type promptApplicationInvocationEnvelope struct {
	RequestID        string                        `json:"request_id"`
	TenantRef        string                        `json:"tenant_ref"`
	WorkspaceID      string                        `json:"workspace_id"`
	ApplicationID    string                        `json:"application_id"`
	Run              *PromptApplicationRunRecordV6 `json:"run"`
	Output           string                        `json:"output"`
	FailureCode      *string                       `json:"failure_code"`
	FailureSummary   string                        `json:"failure_summary"`
	IdempotentReplay bool                          `json:"idempotent_replay"`
	AuditRef         string                        `json:"audit_ref"`
}

func (server *Server) handlePromptApplicationInvocation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, "POST "+promptApplicationInvocationRoute)
	if !server.config.PromptApplicationRuntimeDevHTTPEnabled {
		server.writePlatformError(writer, trace, "PROMPT_APPLICATION_INVOCATION_DEV_DISABLED", "Prompt application invocation requires explicit development opt-in")
		return
	}
	authentication := server.authenticateGatewayAPIKey(request, trace, "prompt_application:invoke")
	if authentication.FailureCode != "" {
		server.writePlatformError(writer, trace, authentication.FailureCode, "")
		return
	}
	var body promptApplicationInvocationBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes: promptApplicationTemplateMaximumSourceBytes + 512, rejectUnknownFields: true,
	}) {
		return
	}
	gatewayContext := authentication.RequestContext
	ctx := PromptApplicationRuntimeContext{
		RequestContext: gatewayContext.RequestContext, RequestID: trace.requestID,
		TenantRef: gatewayContext.TenantRef, WorkspaceID: gatewayContext.WorkspaceID,
		ApplicationID: gatewayContext.ApplicationID, ActorRef: gatewayContext.SubjectRef,
		OwnerSubjectRef: gatewayContext.SubjectRef,
		AuditRef:        "audit_" + trace.requestID + "_prompt-application-invocation",
	}
	result := server.promptApplicationInvocationService().Invoke(ctx, PromptApplicationInvocationInput{
		Variables: body.Variables, ClientInvocationKey: body.ClientInvocationKey,
	})
	var run *PromptApplicationRunRecordV6
	if result.Run != nil {
		document, err := promptApplicationRunDocument(*result.Run)
		if err != nil {
			result = promptApplicationInvocationFailure(PromptApplicationRuntimeFailureStoreContract)
		} else {
			run = &document
		}
	}
	status := http.StatusOK
	if gatewayRequestQuotaFailureCodeFromValue(result.FailureCode) != "" {
		status = gatewayRequestQuotaHTTPStatus(result.FailureCode)
	}
	writeObservedJSON(writer, status, trace, promptApplicationInvocationEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, Run: run, Output: result.Output,
		FailureCode:    optionalApplicationDraftFailure(strings.TrimSpace(result.FailureCode)),
		FailureSummary: result.FailureSummary, IdempotentReplay: result.IdempotentReplay, AuditRef: ctx.AuditRef,
	})
}

func (server *Server) promptApplicationInvocationService() promptApplicationInvocationService {
	resolver := promptApplicationRuntimeAuthorityResolver{
		publishRepository:  server.applicationPublishCandidateRepository,
		draftRepository:    server.applicationDraftRepository,
		templateRepository: server.promptApplicationTemplateRepository,
		readApplication:    server.readApplicationPublishBaseline,
	}
	service := newPromptApplicationInvocationService(
		server.promptApplicationRuntimeRepository, resolver, server.applicationCatalogRepository,
		server.effectiveApplicationRunStore(), server.bridge,
	)
	service.resolveSelection = func(ctx context.Context, model string) northboundSelection {
		return server.resolveNorthboundSelection(ctx, model, nil)
	}
	service.envelopeOptions = server.buildBridgeEnvelopeOptions
	return service
}
