package httpapi

import (
	"log"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	gatewayRequestDevTenantHeader      = "X-RadishMind-Dev-Gateway-Tenant"
	gatewayRequestDevWorkspaceHeader   = "X-RadishMind-Dev-Gateway-Workspace"
	gatewayRequestDevConsumerHeader    = "X-RadishMind-Dev-Gateway-Consumer"
	gatewayRequestDevApplicationHeader = "X-RadishMind-Dev-Gateway-Application"
	gatewayRequestDevSubjectHeader     = "X-RadishMind-Dev-Gateway-Subject"
	gatewayRequestDevScopesHeader      = "X-RadishMind-Dev-Gateway-Scopes"
	gatewayRequestDevAuditHeader       = "X-RadishMind-Dev-Gateway-Audit"
)

func gatewayRequestContextFromDevHeaders(
	request *http.Request,
	trace requestTrace,
) (GatewayRequestContext, bool) {
	requestContext := GatewayRequestContext{
		TenantRef:     strings.TrimSpace(request.Header.Get(gatewayRequestDevTenantHeader)),
		WorkspaceID:   strings.TrimSpace(request.Header.Get(gatewayRequestDevWorkspaceHeader)),
		ConsumerRef:   strings.TrimSpace(request.Header.Get(gatewayRequestDevConsumerHeader)),
		ApplicationID: strings.TrimSpace(request.Header.Get(gatewayRequestDevApplicationHeader)),
		SubjectRef:    strings.TrimSpace(request.Header.Get(gatewayRequestDevSubjectHeader)),
		ScopeGrants:   splitControlPlaneReadDevScopes(request.Header.Get(gatewayRequestDevScopesHeader)),
		AuditContext:  strings.TrimSpace(request.Header.Get(gatewayRequestDevAuditHeader)),
		Source:        "dev_headers",
		RequestID:     trace.requestID,
		AuditRef:      "audit_" + trace.requestID + "_gateway-request",
	}
	return requestContext, validGatewayRequestContext(requestContext) && len(requestContext.ScopeGrants) > 0
}

func (s *Server) startGatewayRequestTrace(
	request *http.Request,
	trace *requestTrace,
	protocol string,
) {
	if trace == nil || !s.config.GatewayRequestHistoryDevEnabled {
		return
	}
	requestContext, ok := gatewayRequestContextFromDevHeaders(request, *trace)
	if !ok {
		log.Printf(
			"radishmind_gateway_request_history request_id=%s route=%s outcome=caller_context_unavailable",
			trace.requestID,
			trace.route,
		)
		return
	}
	record := GatewayRequestRecord{
		SchemaVersion: gatewayRequestRecordSchemaVersion,
		RequestID:     trace.requestID,
		AuditRef:      requestContext.AuditRef,
		TenantRef:     requestContext.TenantRef,
		WorkspaceID:   requestContext.WorkspaceID,
		ConsumerRef:   requestContext.ConsumerRef,
		ApplicationID: requestContext.ApplicationID,
		SubjectRef:    requestContext.SubjectRef,
		Route:         trace.route,
		Protocol:      protocol,
		Status:        GatewayRequestStatusStarted,
		StartedAt:     trace.startedAt.UTC().Format(time.RFC3339Nano),
		Usage:         GatewayRequestUsage{Availability: GatewayRequestUsageNotReported},
	}
	if err := s.gatewayRequestStore().CreateRequest(requestContext, &record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, "create_failed")
		return
	}
	trace.gatewayRequestContext = requestContext
	trace.gatewayRequest = &record
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, "started")
}

func (s *Server) checkpointGatewayRequestTrace(trace *requestTrace, stream bool) {
	if trace == nil || trace.gatewayRequest == nil {
		return
	}
	record := trace.gatewayRequest
	record.Stream = stream
	applyGatewayRequestSelection(record, *trace)
	if err := s.gatewayRequestStore().UpdateRequest(trace.gatewayRequestContext, record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, "checkpoint_failed")
		return
	}
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, "checkpoint_stored")
}

func (s *Server) applyGatewayEnvelopeToTrace(trace *requestTrace, envelope bridge.GatewayEnvelope) {
	if trace == nil || trace.gatewayRequest == nil {
		return
	}
	if value, ok := gatewayMetadataInt64(envelope.Metadata, "duration_ms"); ok {
		trace.gatewayRequest.GatewayDurationMS = value
		trace.gatewayRequest.GatewayDurationAvailable = true
	}
	if value, ok := gatewayMetadataInt64(envelope.Metadata, "provider_duration_ms"); ok {
		trace.gatewayRequest.ProviderDurationMS = value
		trace.gatewayRequest.ProviderDurationAvailable = true
	}
}

func (s *Server) finishGatewayRequestTrace(
	trace *requestTrace,
	status GatewayRequestStatus,
	httpStatusCode int,
	failureCode string,
	failureBoundary string,
) {
	if trace == nil || trace.gatewayRequest == nil || isTerminalGatewayRequestStatus(trace.gatewayRequest.Status) {
		return
	}
	record := trace.gatewayRequest
	applyGatewayRequestSelection(record, *trace)
	record.Status = status
	record.CompletedAt = time.Now().UTC().Format(time.RFC3339Nano)
	record.DurationMS = trace.latencyMilliseconds()
	record.HTTPStatusCode = httpStatusCode
	record.FailureCode = strings.TrimSpace(failureCode)
	record.FailureBoundary = strings.TrimSpace(failureBoundary)
	if err := s.gatewayRequestStore().UpdateRequest(trace.gatewayRequestContext, record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, "terminal_failed")
		return
	}
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, "terminal_stored")
}

func applyGatewayRequestSelection(record *GatewayRequestRecord, trace requestTrace) {
	if record == nil || !trace.hasSelection {
		return
	}
	record.SelectionSource = strings.TrimSpace(trace.selection.source)
	record.SelectedProvider = strings.TrimSpace(trace.selection.provider)
	record.SelectedProfile = strings.TrimSpace(trace.selection.providerProfile)
	record.SelectedModel = strings.TrimSpace(trace.selection.model)
}

func gatewayMetadataInt64(metadata map[string]any, key string) (int64, bool) {
	value, ok := metadata[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case int:
		return int64(typed), typed >= 0
	case int64:
		return typed, typed >= 0
	case float64:
		converted := int64(typed)
		return converted, typed >= 0 && float64(converted) == typed
	default:
		return 0, false
	}
}

func logGatewayRequestStoreOutcome(requestID string, route string, outcome string) {
	log.Printf(
		"radishmind_gateway_request_history request_id=%s route=%s store_mode=memory_dev outcome=%s",
		strings.TrimSpace(requestID),
		strings.TrimSpace(route),
		strings.TrimSpace(outcome),
	)
}
