package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	gatewayRequestListRoute = "GET /v1/model-gateway/requests"
	gatewayRequestReadRoute = "GET /v1/model-gateway/requests/{request_id}"
)

type gatewayRequestSummaryDocument struct {
	SchemaVersion             string                          `json:"schema_version"`
	RecordVersion             int                             `json:"record_version"`
	StoreMode                 string                          `json:"store_mode"`
	RequestID                 string                          `json:"request_id"`
	AuditRef                  string                          `json:"audit_ref"`
	Route                     string                          `json:"route"`
	Protocol                  string                          `json:"protocol"`
	Stream                    bool                            `json:"stream"`
	Status                    GatewayRequestStatus            `json:"status"`
	StartedAt                 string                          `json:"started_at"`
	CompletedAt               string                          `json:"completed_at"`
	DurationMS                int64                           `json:"duration_ms"`
	ProviderDurationMS        int64                           `json:"provider_duration_ms"`
	ProviderDurationAvailable bool                            `json:"provider_duration_available"`
	SelectionSource           string                          `json:"selection_source"`
	SelectedProvider          string                          `json:"selected_provider"`
	SelectedProfile           string                          `json:"selected_profile"`
	SelectedModel             string                          `json:"selected_model"`
	HTTPStatusCode            int                             `json:"http_status_code"`
	FailureCode               string                          `json:"failure_code"`
	FailureBoundary           string                          `json:"failure_boundary"`
	UsageAvailability         GatewayRequestUsageAvailability `json:"usage_availability"`
	StaleStarted              bool                            `json:"stale_started"`
}

type gatewayRequestDetailDocument struct {
	GatewayRequestRecord
	StaleStarted bool `json:"stale_started"`
}

type gatewayRequestListEnvelope struct {
	RequestID      string                          `json:"request_id"`
	TenantRef      string                          `json:"tenant_ref"`
	WorkspaceID    string                          `json:"workspace_id"`
	ConsumerRef    string                          `json:"consumer_ref"`
	ApplicationID  string                          `json:"application_id,omitempty"`
	Requests       []gatewayRequestSummaryDocument `json:"requests"`
	NextCursor     string                          `json:"next_cursor"`
	HasMore        bool                            `json:"has_more"`
	FailureCode    *string                         `json:"failure_code"`
	FailureSummary string                          `json:"failure_summary"`
	AuditRef       string                          `json:"audit_ref"`
}

type gatewayRequestReadEnvelope struct {
	RequestID      string                        `json:"request_id"`
	TenantRef      string                        `json:"tenant_ref"`
	WorkspaceID    string                        `json:"workspace_id"`
	ConsumerRef    string                        `json:"consumer_ref"`
	ApplicationID  string                        `json:"application_id,omitempty"`
	Request        *gatewayRequestDetailDocument `json:"request"`
	FailureCode    *string                       `json:"failure_code"`
	FailureSummary string                        `json:"failure_summary"`
	AuditRef       string                        `json:"audit_ref"`
}

func (s *Server) handleListGatewayRequests(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayRequestListRoute)
	if !s.allowGatewayRequestHistoryDev(writer, trace) {
		return
	}
	requestContext, failureCode := gatewayRequestReadContextFromRequest(request, trace, "list")
	if failureCode != "" {
		writeGatewayRequestListResult(writer, trace, requestContext, gatewayRequestListFailure(failureCode))
		return
	}
	listRequest, failureCode := parseGatewayRequestListRequest(request.URL.Query())
	if failureCode != "" {
		writeGatewayRequestListResult(writer, trace, requestContext, gatewayRequestListFailure(failureCode))
		return
	}
	result := newGatewayRequestHistoryService(s.gatewayRequestStore()).List(requestContext, listRequest)
	writeGatewayRequestListResult(writer, trace, requestContext, result)
}

func (s *Server) handleReadGatewayRequest(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, gatewayRequestReadRoute)
	if !s.allowGatewayRequestHistoryDev(writer, trace) {
		return
	}
	requestContext, failureCode := gatewayRequestReadContextFromRequest(request, trace, "read")
	if failureCode != "" {
		writeGatewayRequestReadResult(writer, trace, requestContext, gatewayRequestReadFailure(failureCode))
		return
	}
	result := newGatewayRequestHistoryService(s.gatewayRequestStore()).Read(
		requestContext,
		strings.TrimSpace(request.PathValue("request_id")),
	)
	writeGatewayRequestReadResult(writer, trace, requestContext, result)
}

func (s *Server) allowGatewayRequestHistoryDev(writer http.ResponseWriter, trace requestTrace) bool {
	if s.config.GatewayRequestHistoryDevEnabled {
		return true
	}
	s.writePlatformError(writer, trace, "GATEWAY_REQUEST_HISTORY_DEV_DISABLED", "gateway request history dev route requires explicit opt-in")
	return false
}

func (s *Server) gatewayRequestStore() gatewayRequestStore {
	if s.gatewayRequestHistoryStore == nil {
		s.gatewayRequestHistoryStore = newMemoryGatewayRequestStore(defaultGatewayRequestStoreCapacity)
	}
	return s.gatewayRequestHistoryStore
}

func gatewayRequestReadContextFromRequest(
	request *http.Request,
	trace requestTrace,
	auditSuffix string,
) (GatewayRequestContext, GatewayRequestHistoryFailureCode) {
	requestContext, ok := gatewayRequestContextFromDevHeaders(request, trace)
	requestContext.AuditRef = "audit_" + trace.requestID + "_gateway-request-" + auditSuffix
	if !ok || !controlPlaneReadHasScope(requestContext.ScopeGrants, "gateway_requests:read") {
		return requestContext, GatewayRequestHistoryFailureScopeDenied
	}
	values := request.URL.Query()
	workspaceID := strings.TrimSpace(values.Get("workspace_id"))
	consumerRef := strings.TrimSpace(values.Get("consumer_ref"))
	applicationID := strings.TrimSpace(values.Get("application_id"))
	if workspaceID == "" || consumerRef == "" || workspaceID != requestContext.WorkspaceID ||
		consumerRef != requestContext.ConsumerRef ||
		(applicationID != "" && applicationID != requestContext.ApplicationID) {
		return requestContext, GatewayRequestHistoryFailureScopeDenied
	}
	return requestContext, ""
}

func parseGatewayRequestListRequest(
	values map[string][]string,
) (GatewayRequestListRequest, GatewayRequestHistoryFailureCode) {
	allowed := map[string]bool{
		"workspace_id": true, "consumer_ref": true, "application_id": true,
		"limit": true, "cursor": true, "route": true, "protocol": true,
		"provider": true, "profile": true, "model": true, "status": true,
		"failure_boundary": true, "usage_availability": true,
		"started_from": true, "started_to": true,
	}
	for key, entries := range values {
		if !allowed[key] || len(entries) != 1 {
			return GatewayRequestListRequest{}, GatewayRequestHistoryFailureFilterInvalid
		}
	}
	request := GatewayRequestListRequest{
		Cursor:            strings.TrimSpace(firstQueryValue(values, "cursor")),
		Route:             strings.TrimSpace(firstQueryValue(values, "route")),
		Protocol:          strings.TrimSpace(firstQueryValue(values, "protocol")),
		Provider:          strings.TrimSpace(firstQueryValue(values, "provider")),
		Profile:           strings.TrimSpace(firstQueryValue(values, "profile")),
		Model:             strings.TrimSpace(firstQueryValue(values, "model")),
		Status:            GatewayRequestStatus(strings.TrimSpace(firstQueryValue(values, "status"))),
		FailureBoundary:   strings.TrimSpace(firstQueryValue(values, "failure_boundary")),
		UsageAvailability: GatewayRequestUsageAvailability(strings.TrimSpace(firstQueryValue(values, "usage_availability"))),
	}
	if raw := strings.TrimSpace(firstQueryValue(values, "limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil {
			return GatewayRequestListRequest{}, GatewayRequestHistoryFailureFilterInvalid
		}
		request.Limit = limit
	}
	for key, target := range map[string]**time.Time{
		"started_from": &request.StartedFrom,
		"started_to":   &request.StartedTo,
	} {
		raw := strings.TrimSpace(firstQueryValue(values, key))
		if raw == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return GatewayRequestListRequest{}, GatewayRequestHistoryFailureFilterInvalid
		}
		parsed = parsed.UTC()
		*target = &parsed
	}
	return request, ""
}

func writeGatewayRequestListResult(
	writer http.ResponseWriter,
	trace requestTrace,
	requestContext GatewayRequestContext,
	result GatewayRequestListResult,
) {
	requests := make([]gatewayRequestSummaryDocument, 0, len(result.Records))
	for _, record := range result.Records {
		requests = append(requests, gatewayRequestSummaryFromRecord(record, time.Now().UTC()))
	}
	writeObservedJSON(writer, http.StatusOK, trace, gatewayRequestListEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef,
		WorkspaceID: requestContext.WorkspaceID, ConsumerRef: requestContext.ConsumerRef,
		ApplicationID: requestContext.ApplicationID, Requests: requests,
		NextCursor: result.NextCursor, HasMore: result.HasMore,
		FailureCode: gatewayRequestFailureCodePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		AuditRef: requestContext.AuditRef,
	})
}

func writeGatewayRequestReadResult(
	writer http.ResponseWriter,
	trace requestTrace,
	requestContext GatewayRequestContext,
	result GatewayRequestReadResult,
) {
	var document *gatewayRequestDetailDocument
	if result.Record != nil {
		document = &gatewayRequestDetailDocument{
			GatewayRequestRecord: *result.Record,
			StaleStarted:         gatewayRequestIsStale(*result.Record, time.Now().UTC()),
		}
	}
	writeObservedJSON(writer, http.StatusOK, trace, gatewayRequestReadEnvelope{
		RequestID: trace.requestID, TenantRef: requestContext.TenantRef,
		WorkspaceID: requestContext.WorkspaceID, ConsumerRef: requestContext.ConsumerRef,
		ApplicationID: requestContext.ApplicationID, Request: document,
		FailureCode: gatewayRequestFailureCodePointer(result.FailureCode), FailureSummary: result.FailureSummary,
		AuditRef: requestContext.AuditRef,
	})
}

func gatewayRequestSummaryFromRecord(record GatewayRequestRecord, now time.Time) gatewayRequestSummaryDocument {
	return gatewayRequestSummaryDocument{
		SchemaVersion: record.SchemaVersion, RecordVersion: record.RecordVersion,
		StoreMode: record.StoreMode,
		RequestID: record.RequestID, AuditRef: record.AuditRef, Route: record.Route,
		Protocol: record.Protocol, Stream: record.Stream, Status: record.Status,
		StartedAt: record.StartedAt, CompletedAt: record.CompletedAt, DurationMS: record.DurationMS,
		ProviderDurationMS:        record.ProviderDurationMS,
		ProviderDurationAvailable: record.ProviderDurationAvailable,
		SelectionSource:           record.SelectionSource, SelectedProvider: record.SelectedProvider,
		SelectedProfile: record.SelectedProfile, SelectedModel: record.SelectedModel,
		HTTPStatusCode: record.HTTPStatusCode, FailureCode: record.FailureCode,
		FailureBoundary: record.FailureBoundary, UsageAvailability: record.Usage.Availability,
		StaleStarted: gatewayRequestIsStale(record, now),
	}
}

func gatewayRequestIsStale(record GatewayRequestRecord, now time.Time) bool {
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	return err == nil && record.Status == GatewayRequestStatusStarted && now.Sub(startedAt) > gatewayRequestStaleThreshold
}

func gatewayRequestFailureCodePointer(code GatewayRequestHistoryFailureCode) *string {
	if code == "" {
		return nil
	}
	value := string(code)
	return &value
}
