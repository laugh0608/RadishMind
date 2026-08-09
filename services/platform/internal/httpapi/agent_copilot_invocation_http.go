package httpapi

import (
	"context"
	"net/http"
	"strings"
)

type agentCopilotInvocationBody struct {
	Task                string                 `json:"task"`
	Locale              string                 `json:"locale"`
	ConversationID      string                 `json:"conversation_id,omitempty"`
	Artifacts           []AgentCopilotArtifact `json:"artifacts"`
	Context             map[string]any         `json:"context"`
	ClientInvocationKey string                 `json:"client_invocation_key"`
}

type agentCopilotInvocationEnvelope struct {
	RequestID        string                   `json:"request_id"`
	TenantRef        string                   `json:"tenant_ref"`
	WorkspaceID      string                   `json:"workspace_id"`
	ApplicationID    string                   `json:"application_id"`
	Run              *AgentCopilotRunRecordV7 `json:"run"`
	Response         *AgentCopilotResponse    `json:"response"`
	FailureCode      *string                  `json:"failure_code"`
	FailureSummary   string                   `json:"failure_summary"`
	IdempotentReplay bool                     `json:"idempotent_replay"`
	AuditRef         string                   `json:"audit_ref"`
}

func (server *Server) handleAgentCopilotInvocation(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, "POST "+agentCopilotInvocationRoute)
	if !server.config.AgentCopilotRuntimeDevHTTPEnabled {
		server.writePlatformError(writer, trace, "AGENT_COPILOT_INVOCATION_DEV_DISABLED", "Agent Copilot invocation requires explicit development opt-in")
		return
	}
	authentication := server.authenticateGatewayAPIKey(request, trace, agentCopilotInvokeScope)
	if authentication.FailureCode != "" {
		server.writePlatformError(writer, trace, authentication.FailureCode, "")
		return
	}
	var body agentCopilotInvocationBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{
		maxBytes: agentCopilotMaximumInvocationBytes + 512, rejectUnknownFields: true,
	}) {
		return
	}
	gatewayContext := authentication.RequestContext
	ctx := AgentCopilotRuntimeContext{
		RequestContext: gatewayContext.RequestContext, RequestID: trace.requestID,
		TenantRef: gatewayContext.TenantRef, WorkspaceID: gatewayContext.WorkspaceID,
		ApplicationID: gatewayContext.ApplicationID, ActorRef: gatewayContext.SubjectRef,
		OwnerSubjectRef: gatewayContext.SubjectRef,
		AuditRef:        "audit_" + trace.requestID + "_agent-copilot-invocation",
	}
	result := server.agentCopilotInvocationService().Invoke(ctx, AgentCopilotInvocationInput{
		Task: body.Task, Locale: body.Locale, ConversationID: body.ConversationID,
		Artifacts: body.Artifacts, Context: body.Context, ClientInvocationKey: body.ClientInvocationKey,
	})
	var run *AgentCopilotRunRecordV7
	if result.Run != nil {
		document, err := agentCopilotRunDocument(*result.Run)
		if err != nil {
			result = agentCopilotInvocationFailure(AgentCopilotRuntimeFailureStoreContract)
		} else {
			run = &document
		}
	}
	status := http.StatusOK
	if gatewayRequestQuotaFailureCodeFromValue(result.FailureCode) != "" {
		status = gatewayRequestQuotaHTTPStatus(result.FailureCode)
	}
	writeObservedJSON(writer, status, trace, agentCopilotInvocationEnvelope{
		RequestID: trace.requestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, Run: run, Response: result.Response,
		FailureCode:    optionalApplicationDraftFailure(strings.TrimSpace(result.FailureCode)),
		FailureSummary: result.FailureSummary, IdempotentReplay: result.IdempotentReplay, AuditRef: ctx.AuditRef,
	})
}

func (server *Server) agentCopilotInvocationService() agentCopilotInvocationService {
	resolver := agentCopilotRuntimeAuthorityResolver{
		publishRepository: server.applicationPublishCandidateRepository,
		draftRepository:   server.applicationDraftRepository, profileRepository: server.agentCopilotProfileRepository,
		readApplication: server.readApplicationPublishBaseline,
	}
	service := newAgentCopilotInvocationService(
		server.agentCopilotRuntimeRepository, resolver, server.applicationCatalogRepository,
		server.effectiveApplicationRunStore(), server.bridge,
	)
	service.resolveSelection = func(ctx context.Context, model string) northboundSelection {
		return server.resolveNorthboundSelection(ctx, model, nil)
	}
	service.envelopeOptions = server.buildBridgeEnvelopeOptions
	return service
}
