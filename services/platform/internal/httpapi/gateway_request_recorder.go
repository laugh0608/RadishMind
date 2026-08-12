package httpapi

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
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
		RequestContext: request.Context(),
		TenantRef:      strings.TrimSpace(request.Header.Get(gatewayRequestDevTenantHeader)),
		WorkspaceID:    strings.TrimSpace(request.Header.Get(gatewayRequestDevWorkspaceHeader)),
		ConsumerRef:    strings.TrimSpace(request.Header.Get(gatewayRequestDevConsumerHeader)),
		ApplicationID:  strings.TrimSpace(request.Header.Get(gatewayRequestDevApplicationHeader)),
		SubjectRef:     strings.TrimSpace(request.Header.Get(gatewayRequestDevSubjectHeader)),
		ScopeGrants:    splitControlPlaneReadDevScopes(request.Header.Get(gatewayRequestDevScopesHeader)),
		AuditContext:   strings.TrimSpace(request.Header.Get(gatewayRequestDevAuditHeader)),
		Source:         "dev_headers",
		RequestID:      trace.requestID,
		AuditRef:       "audit_" + trace.requestID + "_gateway-request",
	}
	return requestContext, validGatewayRequestContext(requestContext) && len(requestContext.ScopeGrants) > 0
}

func (s *Server) startGatewayRequestTrace(
	request *http.Request,
	trace *requestTrace,
	protocol string,
) error {
	if trace == nil || !s.config.GatewayRequestHistoryDevEnabled {
		return nil
	}
	requestContext := trace.gatewayRequestContext
	ok := false
	if config.EffectiveGatewayAuthMode(s.config) == gatewayAPIKeyAuthenticationSource {
		ok = requestContext.Source == gatewayAPIKeyAuthenticationSource && validGatewayRequestContext(requestContext)
		if !ok {
			return errGatewayRequestStoreContract
		}
	} else {
		requestContext, ok = gatewayRequestContextFromDevHeaders(request, *trace)
	}
	if !ok {
		log.Printf(
			"radishmind_gateway_request_history request_id=%s route=%s outcome=caller_context_unavailable",
			trace.requestID,
			trace.route,
		)
		return nil
	}
	schemaVersion := gatewayRequestRecordSchemaVersionV1
	costEstimate := GatewayRequestCostEstimate{}
	if s.config.GatewayModelPricingCaptureDevEnabled {
		schemaVersion = gatewayRequestRecordSchemaVersionV2
		costEstimate = gatewayRequestCostUnavailable(GatewayRequestCostNotApplicable, "provider_not_attempted")
	}
	record := GatewayRequestRecord{
		SchemaVersion: schemaVersion,
		StoreMode:     s.gatewayRequestHistoryStoreMode,
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
		CostEstimate:  costEstimate,
	}
	if err := s.gatewayRequestStore().CreateRequest(requestContext, &record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "create_failed")
		if requestContext.Source == gatewayAPIKeyAuthenticationSource {
			return err
		}
		return nil
	}
	trace.gatewayRequestContext = requestContext
	trace.gatewayRequest = &record
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "started")
	return nil
}

func (s *Server) checkpointGatewayRequestTrace(trace *requestTrace, stream bool) {
	if trace == nil || trace.gatewayRequest == nil {
		return
	}
	record := trace.gatewayRequest
	record.Stream = stream
	applyGatewayRequestSelection(record, *trace)
	s.bindGatewayModelPricingSnapshot(trace)
	if err := s.gatewayRequestStore().UpdateRequest(trace.gatewayRequestContext, record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "checkpoint_failed")
		return
	}
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "checkpoint_stored")
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
	trace.gatewayRequest.Usage = gatewayUsageFromEnvelope(envelope)
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
	if record.SchemaVersion == gatewayRequestRecordSchemaVersionV2 {
		if !trace.providerAttempted {
			record.Usage = GatewayRequestUsage{Availability: GatewayRequestUsageNotApplicable}
		}
		record.CostEstimate = buildGatewayRequestCostEstimate(
			trace.providerAttempted,
			record.Usage,
			trace.gatewayPricingSnapshot,
		)
	}
	requestContext, cancel := s.gatewayRequestTerminalStoreContext(trace.gatewayRequestContext)
	defer cancel()
	if err := s.gatewayRequestStore().UpdateRequest(requestContext, record); err != nil {
		logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "terminal_failed")
		return
	}
	logGatewayRequestStoreOutcome(trace.requestID, trace.route, record.StoreMode, "terminal_stored")
}

func (s *Server) gatewayRequestTerminalStoreContext(requestContext GatewayRequestContext) (GatewayRequestContext, context.CancelFunc) {
	timeout := s.config.GatewayRequestDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	parent := requestContext.RequestContext
	if parent == nil {
		parent = context.Background()
	} else {
		parent = context.WithoutCancel(parent)
	}
	terminalContext, cancel := context.WithTimeout(parent, timeout)
	requestContext.RequestContext = terminalContext
	return requestContext, cancel
}

func applyGatewayRequestSelection(record *GatewayRequestRecord, trace requestTrace) {
	if record == nil || !trace.hasSelection {
		return
	}
	record.SelectionSource = strings.TrimSpace(trace.selection.source)
	record.SelectedProvider = strings.TrimSpace(trace.selection.provider)
	record.SelectedProfile = strings.TrimSpace(trace.selection.providerProfile)
	record.SelectedModel = strings.TrimSpace(trace.selection.model)
	record.ProviderRouteConfigurationID = strings.TrimSpace(trace.selection.routeConfigurationID)
	record.ProviderRouteGeneration = trace.selection.routeGeneration
	record.ProviderRouteSnapshotDigest = strings.TrimSpace(trace.selection.routeSnapshotDigest)
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

func gatewayUsageFromEnvelope(envelope bridge.GatewayEnvelope) GatewayRequestUsage {
	notReported := GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}
	rawUsage, ok := envelope.Metadata["usage"].(map[string]any)
	if !ok {
		return notReported
	}
	availability, _ := rawUsage["availability"].(string)
	if GatewayRequestUsageAvailability(strings.TrimSpace(availability)) != GatewayRequestUsageReported {
		return notReported
	}
	source, _ := rawUsage["source"].(string)
	source = strings.TrimSpace(source)
	switch source {
	case "openai_compatible_usage", "gemini_usage_metadata", "anthropic_usage",
		"huggingface_usage", "ollama_usage", "ollama_eval_counts":
	default:
		return notReported
	}
	inputTokens, inputOK := gatewayMetadataInt64(rawUsage, "input_tokens")
	outputTokens, outputOK := gatewayMetadataInt64(rawUsage, "output_tokens")
	totalTokens, totalOK := gatewayMetadataInt64(rawUsage, "total_tokens")
	maxInt := int64(^uint(0) >> 1)
	if !inputOK || !outputOK || !totalOK ||
		inputTokens > maxInt || outputTokens > maxInt || totalTokens > maxInt {
		return notReported
	}
	usage := GatewayRequestUsage{
		Availability: GatewayRequestUsageReported,
		Source:       source,
		InputTokens:  int(inputTokens),
		OutputTokens: int(outputTokens),
		TotalTokens:  int(totalTokens),
	}
	if !validGatewayRequestUsage(usage) {
		return notReported
	}
	return usage
}

func logGatewayRequestStoreOutcome(requestID string, route string, storeMode string, outcome string) {
	log.Printf(
		"radishmind_gateway_request_history request_id=%s route=%s store_mode=%s outcome=%s",
		strings.TrimSpace(requestID),
		strings.TrimSpace(route),
		strings.TrimSpace(storeMode),
		strings.TrimSpace(outcome),
	)
}
