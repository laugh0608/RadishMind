package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const (
	errorBoundaryNorthboundRequest  = "northbound_request"
	errorBoundaryCanonicalRequest   = "canonical_request"
	errorBoundaryProviderInventory  = "provider_inventory"
	errorBoundaryPythonBridge       = "python_bridge"
	errorBoundaryPlatformResponse   = "platform_response"
	errorBoundarySouthboundProvider = "southbound_provider"
	errorBoundaryGatewayAuth        = "gateway_authentication"
	errorBoundaryConfiguration      = "configuration"
	errorBoundaryUnknown            = "unknown"
)

type requestTrace struct {
	requestID             string
	route                 string
	startedAt             time.Time
	selection             northboundSelection
	hasSelection          bool
	gatewayRequestContext GatewayRequestContext
	gatewayRequest        *GatewayRequestRecord
}

type platformErrorDefinition struct {
	statusCode      int
	errorType       string
	failureBoundary string
	defaultMessage  string
}

func newRequestTrace(request *http.Request, route string) requestTrace {
	requestID := strings.TrimSpace(request.Header.Get("X-Request-Id"))
	if requestID == "" {
		requestID = strings.TrimSpace(request.Header.Get("X-Request-ID"))
	}
	if requestID == "" {
		requestID = strings.TrimSpace(request.Header.Get("OpenAI-Request-ID"))
	}
	if requestID == "" {
		requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return requestTrace{
		requestID: requestID,
		route:     strings.TrimSpace(route),
		startedAt: time.Now(),
	}
}

func (trace *requestTrace) applySelection(selection northboundSelection) {
	trace.selection = selection
	trace.hasSelection = true
}

func (trace requestTrace) latencyMilliseconds() int64 {
	if trace.startedAt.IsZero() {
		return 0
	}
	return time.Since(trace.startedAt).Milliseconds()
}

func writeTraceHeaders(writer http.ResponseWriter, trace requestTrace) {
	if trace.requestID != "" {
		writer.Header().Set("X-Request-Id", trace.requestID)
	}
	if trace.route != "" {
		writer.Header().Set("X-RadishMind-Route", trace.route)
	}
}

func writeObservedJSON(writer http.ResponseWriter, statusCode int, trace requestTrace, document any) {
	writeTraceHeaders(writer, trace)
	writeJSON(writer, statusCode, document)
	logRequestTrace(trace, statusCode, "", "")
}

func (s *Server) writePlatformError(writer http.ResponseWriter, trace requestTrace, code string, detail string) {
	definition := lookupPlatformErrorDefinition(code)
	message := sanitizePlatformErrorDetail(s.config, detail)
	if message == "" {
		message = definition.defaultMessage
	}
	if message == "" {
		message = "platform request failed"
	}

	writeTraceHeaders(writer, trace)
	writeJSON(writer, definition.statusCode, errorDocument{
		Error: errorBody{
			Message:         message,
			Type:            definition.errorType,
			Code:            strings.TrimSpace(code),
			RequestID:       trace.requestID,
			Route:           trace.route,
			FailureBoundary: definition.failureBoundary,
			Metadata:        buildTraceErrorMetadata(trace),
		},
	})
	logRequestTrace(trace, definition.statusCode, strings.TrimSpace(code), definition.failureBoundary)
	status := GatewayRequestStatusFailed
	if strings.TrimSpace(code) == bridge.ErrorCodeWorkerCanceled {
		status = GatewayRequestStatusCanceled
	}
	s.finishGatewayRequestTrace(&trace, status, definition.statusCode, strings.TrimSpace(code), definition.failureBoundary)
}

func lookupPlatformErrorDefinition(code string) platformErrorDefinition {
	normalizedCode := strings.TrimSpace(code)
	definitions := map[string]platformErrorDefinition{
		"INVALID_JSON": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "request body must be valid JSON",
		},
		"REQUEST_BODY_TOO_LARGE": {
			statusCode:      http.StatusRequestEntityTooLarge,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "request body exceeds the endpoint limit",
		},
		"MISSING_MESSAGES": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "messages must contain at least one item",
		},
		"INVALID_CHAT_MESSAGES": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "chat messages could not be translated",
		},
		"INVALID_RESPONSES_REQUEST": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "responses request could not be translated",
		},
		"INVALID_MESSAGES_REQUEST": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "messages request could not be translated",
		},
		"REQUEST_SCHEMA_INVALID": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryCanonicalRequest,
			defaultMessage:  "canonical request schema validation failed",
		},
		"MODEL_NOT_FOUND": {
			statusCode:      http.StatusNotFound,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryProviderInventory,
			defaultMessage:  "model not found",
		},
		"MISSING_MODEL_ID": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "model id is required",
		},
		"MISSING_CHECKPOINT_ID": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint id is required",
		},
		"MISSING_TOOL_ID": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "tool_id is required",
		},
		"CHECKPOINT_NOT_FOUND": {
			statusCode:      http.StatusNotFound,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint not found",
		},
		"CHECKPOINT_MATERIALIZED_RESULTS_DISABLED": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint read is metadata-only; materialized results are disabled",
		},
		"CHECKPOINT_AUTO_REPLAY_DISABLED": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint read does not support automatic replay",
		},
		"CHECKPOINT_DURABLE_MEMORY_DISABLED": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint read does not support durable memory",
		},
		"CONTROL_PLANE_READ_METHOD_NOT_ALLOWED": {
			statusCode:      http.StatusMethodNotAllowed,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "control plane read route only supports GET",
		},
		"CONTROL_PLANE_READ_QUERY_FORBIDDEN": {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "control plane read route rejected a forbidden query parameter",
		},
		"SAVED_WORKFLOW_DRAFT_DEV_HTTP_DISABLED": {
			statusCode:      http.StatusForbidden,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "saved workflow draft dev HTTP route is disabled",
		},
		"API_KEY_LIFECYCLE_DEV_HTTP_DISABLED": {
			statusCode:      http.StatusForbidden,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "API key lifecycle dev HTTP route is disabled",
		},
		APIKeyFailureMissing: {
			statusCode: http.StatusUnauthorized, errorType: "authentication_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key is required",
		},
		APIKeyFailureInvalid: {
			statusCode: http.StatusUnauthorized, errorType: "authentication_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key is invalid",
		},
		APIKeyFailureCredentialConflict: {
			statusCode: http.StatusBadRequest, errorType: "invalid_request_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "conflicting Gateway credentials are not allowed",
		},
		APIKeyFailureRevoked: {
			statusCode: http.StatusForbidden, errorType: "authentication_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key is revoked",
		},
		APIKeyFailureExpired: {
			statusCode: http.StatusForbidden, errorType: "authentication_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key is expired",
		},
		APIKeyFailureScopeDenied: {
			statusCode: http.StatusForbidden, errorType: "permission_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key scope is denied",
		},
		APIKeyFailureApplicationUnavailable: {
			statusCode: http.StatusForbidden, errorType: "authentication_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key application is unavailable",
		},
		APIKeyFailureStoreUnavailable: {
			statusCode: http.StatusServiceUnavailable, errorType: "service_unavailable_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "API key store is unavailable",
		},
		string(GatewayRequestHistoryFailureStore): {
			statusCode: http.StatusServiceUnavailable, errorType: "service_unavailable_error",
			failureBoundary: errorBoundaryGatewayAuth, defaultMessage: "Gateway request history store is unavailable",
		},
		"WORKFLOW_EXECUTOR_DEV_DISABLED": {
			statusCode:      http.StatusForbidden,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "workflow executor dev route is disabled",
		},
		"WORKFLOW_RAG_SNAPSHOT_DEV_DISABLED": {
			statusCode:      http.StatusForbidden,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "workflow RAG snapshot dev route is disabled",
		},
		WorkflowRAGFailurePayloadInvalid: {
			statusCode:      http.StatusBadRequest,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "workflow RAG snapshot request is invalid",
		},
		"GATEWAY_REQUEST_HISTORY_DEV_DISABLED": {
			statusCode:      http.StatusForbidden,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "gateway request history dev route is disabled",
		},
		"CONFIG_REQUIRED_FIELDS_MISSING": {
			statusCode:      http.StatusServiceUnavailable,
			errorType:       "configuration_error",
			failureBoundary: errorBoundaryConfiguration,
			defaultMessage:  "required platform config fields are missing",
		},
		"PROVIDER_REGISTRY_UNAVAILABLE": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "python bridge provider registry is unavailable",
		},
		"PROVIDER_INVENTORY_UNAVAILABLE": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "python bridge provider inventory is unavailable",
		},
		"PLATFORM_BRIDGE_FAILED": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "platform bridge failed",
		},
		"PLATFORM_GATEWAY_FAILED": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "platform gateway failed",
		},
		bridge.ErrorCodeWorkerQueueFull: {
			statusCode:      http.StatusServiceUnavailable,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker queue is full",
		},
		bridge.ErrorCodeWorkerTimeout: {
			statusCode:      http.StatusGatewayTimeout,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker request timed out",
		},
		bridge.ErrorCodeWorkerCanceled: {
			statusCode:      http.StatusRequestTimeout,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker request was canceled",
		},
		bridge.ErrorCodeWorkerExited: {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker exited before completing request",
		},
		bridge.ErrorCodeWorkerProtocol: {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker protocol failed",
		},
		bridge.ErrorCodeWorkerUnavailable: {
			statusCode:      http.StatusServiceUnavailable,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker is unavailable",
		},
		bridge.ErrorCodeWorkerRequestFailed: {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge worker request failed",
		},
		bridge.ErrorCodeClientClosed: {
			statusCode:      http.StatusServiceUnavailable,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "bridge client is closed",
		},
		bridge.ErrorCodeProcessFailed: {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPythonBridge,
			defaultMessage:  "platform bridge process failed",
		},
		"PLATFORM_RESPONSE_INVALID": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryPlatformResponse,
			defaultMessage:  "platform response could not be translated",
		},
		"UNSUPPORTED_PROVIDER": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundarySouthboundProvider,
			defaultMessage:  "selected provider is unsupported",
		},
		"UNSUPPORTED_TASK": {
			statusCode:      http.StatusBadGateway,
			errorType:       "platform_error",
			failureBoundary: errorBoundaryCanonicalRequest,
			defaultMessage:  "canonical task is unsupported",
		},
	}
	if definition, ok := definitions[normalizedCode]; ok {
		return definition
	}
	return platformErrorDefinition{
		statusCode:      http.StatusBadGateway,
		errorType:       "platform_error",
		failureBoundary: errorBoundaryUnknown,
		defaultMessage:  "platform request failed",
	}
}

func bridgeFailureCode(err error) string {
	if code := bridge.ErrorCode(err); code != "" {
		return code
	}
	return "PLATFORM_BRIDGE_FAILED"
}

func buildTraceErrorMetadata(trace requestTrace) map[string]any {
	metadata := map[string]any{
		"latency_ms": trace.latencyMilliseconds(),
	}
	if trace.hasSelection {
		for key, value := range buildNorthboundSelectionMetadata(trace.selection) {
			metadata[key] = value
		}
	}
	return metadata
}

func logRequestTrace(trace requestTrace, statusCode int, errorCode string, failureBoundary string) {
	fields := []string{
		"request_id=" + trace.requestID,
		"route=" + trace.route,
		fmt.Sprintf("status=%d", statusCode),
		fmt.Sprintf("latency_ms=%d", trace.latencyMilliseconds()),
	}
	if trace.hasSelection {
		fields = append(fields,
			"provider="+strings.TrimSpace(trace.selection.provider),
			"profile="+strings.TrimSpace(trace.selection.providerProfile),
			"model="+strings.TrimSpace(trace.selection.model),
			"selection_source="+strings.TrimSpace(trace.selection.source),
		)
	}
	if errorCode != "" {
		fields = append(fields, "error_code="+errorCode)
	}
	if failureBoundary != "" {
		fields = append(fields, "failure_boundary="+failureBoundary)
	}
	log.Printf("radishmind_platform_request %s", strings.Join(fields, " "))
}

func sanitizePlatformErrorDetail(cfg config.Config, detail string) string {
	message := strings.TrimSpace(detail)
	if message == "" {
		return ""
	}
	for _, secret := range []string{cfg.APIKey, cfg.BaseURL} {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		message = strings.ReplaceAll(message, secret, "[redacted]")
	}
	if len(message) > 320 {
		return message[:317] + "..."
	}
	return message
}
