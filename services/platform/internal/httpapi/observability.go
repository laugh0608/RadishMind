package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	errorBoundaryNorthboundRequest  = "northbound_request"
	errorBoundaryCanonicalRequest   = "canonical_request"
	errorBoundaryProviderInventory  = "provider_inventory"
	errorBoundaryPythonBridge       = "python_bridge"
	errorBoundaryPlatformResponse   = "platform_response"
	errorBoundarySouthboundProvider = "southbound_provider"
	errorBoundaryConfiguration      = "configuration"
	errorBoundaryUnknown            = "unknown"
)

type requestTrace struct {
	requestID    string
	route        string
	startedAt    time.Time
	selection    northboundSelection
	hasSelection bool
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
		"CHECKPOINT_NOT_FOUND": {
			statusCode:      http.StatusNotFound,
			errorType:       "invalid_request_error",
			failureBoundary: errorBoundaryNorthboundRequest,
			defaultMessage:  "checkpoint not found",
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
