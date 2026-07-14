package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	gatewayRequestRecordSchemaVersion      = "gateway_request_record.v1"
	gatewayRequestStoreModeMemoryDev       = "memory_dev"
	gatewayRequestStoreModeSQLiteDev       = "sqlite_dev"
	gatewayRequestStoreModePostgresDevTest = "postgres_dev_test"
	gatewayRequestListDefaultLimit         = 25
	gatewayRequestListMaxLimit             = 100
	gatewayRequestListMaxRange             = 31 * 24 * time.Hour
	gatewayRequestListFutureSkew           = 5 * time.Minute
	gatewayRequestStaleThreshold           = 5 * time.Minute
)

type GatewayRequestStatus string

const (
	GatewayRequestStatusStarted   GatewayRequestStatus = "started"
	GatewayRequestStatusSucceeded GatewayRequestStatus = "succeeded"
	GatewayRequestStatusFailed    GatewayRequestStatus = "failed"
	GatewayRequestStatusCanceled  GatewayRequestStatus = "canceled"
)

type GatewayRequestUsageAvailability string

const (
	GatewayRequestUsageReported      GatewayRequestUsageAvailability = "reported"
	GatewayRequestUsageNotReported   GatewayRequestUsageAvailability = "not_reported"
	GatewayRequestUsageNotApplicable GatewayRequestUsageAvailability = "not_applicable"
)

type GatewayRequestHistoryFailureCode string

const (
	GatewayRequestHistoryFailureDisabled          GatewayRequestHistoryFailureCode = "gateway_request_history_disabled"
	GatewayRequestHistoryFailureScopeDenied       GatewayRequestHistoryFailureCode = "gateway_request_scope_denied"
	GatewayRequestHistoryFailureNotFound          GatewayRequestHistoryFailureCode = "gateway_request_record_not_found"
	GatewayRequestHistoryFailureFilterInvalid     GatewayRequestHistoryFailureCode = "gateway_request_filter_invalid"
	GatewayRequestHistoryFailureCursorInvalid     GatewayRequestHistoryFailureCode = "gateway_request_cursor_invalid"
	GatewayRequestHistoryFailureStore             GatewayRequestHistoryFailureCode = "gateway_request_store_unavailable"
	GatewayRequestHistoryFailureContract          GatewayRequestHistoryFailureCode = "gateway_request_store_contract_mismatch"
	GatewayRequestHistoryFailureStoreModeInvalid  GatewayRequestHistoryFailureCode = "gateway_request_store_mode_invalid"
	GatewayRequestHistoryFailureStoreModeDisabled GatewayRequestHistoryFailureCode = "gateway_request_store_mode_disabled"
)

type GatewayRequestContext struct {
	RequestContext context.Context
	TenantRef      string
	WorkspaceID    string
	ConsumerRef    string
	ApplicationID  string
	SubjectRef     string
	ScopeGrants    []string
	AuditContext   string
	Source         string
	RequestID      string
	AuditRef       string
}

type GatewayRequestUsage struct {
	Availability GatewayRequestUsageAvailability `json:"availability"`
	Source       string                          `json:"source"`
	InputTokens  int                             `json:"input_tokens"`
	OutputTokens int                             `json:"output_tokens"`
	TotalTokens  int                             `json:"total_tokens"`
}

type GatewayRequestRecord struct {
	SchemaVersion             string               `json:"schema_version"`
	RecordVersion             int                  `json:"record_version"`
	StoreMode                 string               `json:"store_mode"`
	RequestID                 string               `json:"request_id"`
	AuditRef                  string               `json:"audit_ref"`
	TenantRef                 string               `json:"tenant_ref"`
	WorkspaceID               string               `json:"workspace_id"`
	ConsumerRef               string               `json:"consumer_ref"`
	ApplicationID             string               `json:"application_id,omitempty"`
	SubjectRef                string               `json:"subject_ref"`
	Route                     string               `json:"route"`
	Protocol                  string               `json:"protocol"`
	Stream                    bool                 `json:"stream"`
	Status                    GatewayRequestStatus `json:"status"`
	StartedAt                 string               `json:"started_at"`
	CompletedAt               string               `json:"completed_at"`
	DurationMS                int64                `json:"duration_ms"`
	GatewayDurationMS         int64                `json:"gateway_duration_ms"`
	GatewayDurationAvailable  bool                 `json:"gateway_duration_available"`
	ProviderDurationMS        int64                `json:"provider_duration_ms"`
	ProviderDurationAvailable bool                 `json:"provider_duration_available"`
	SelectionSource           string               `json:"selection_source"`
	SelectedProvider          string               `json:"selected_provider"`
	SelectedProfile           string               `json:"selected_profile"`
	SelectedModel             string               `json:"selected_model"`
	HTTPStatusCode            int                  `json:"http_status_code"`
	FailureCode               string               `json:"failure_code"`
	FailureBoundary           string               `json:"failure_boundary"`
	Usage                     GatewayRequestUsage  `json:"usage"`
}

type GatewayRequestListRequest struct {
	Limit             int
	Cursor            string
	Route             string
	Protocol          string
	Provider          string
	Profile           string
	Model             string
	Status            GatewayRequestStatus
	FailureBoundary   string
	UsageAvailability GatewayRequestUsageAvailability
	StartedFrom       *time.Time
	StartedTo         *time.Time
}

type GatewayRequestListResult struct {
	Records        []GatewayRequestRecord
	NextCursor     string
	HasMore        bool
	FailureCode    GatewayRequestHistoryFailureCode
	FailureSummary string
}

type GatewayRequestReadResult struct {
	Record         *GatewayRequestRecord
	FailureCode    GatewayRequestHistoryFailureCode
	FailureSummary string
}

type gatewayRequestHistoryService struct {
	store gatewayRequestStore
}

func newGatewayRequestHistoryService(store gatewayRequestStore) gatewayRequestHistoryService {
	return gatewayRequestHistoryService{store: store}
}

func (service gatewayRequestHistoryService) Read(
	requestContext GatewayRequestContext,
	requestID string,
) GatewayRequestReadResult {
	requestID = strings.TrimSpace(requestID)
	if !validGatewayRequestReference(requestID, 160) {
		return gatewayRequestReadFailure(GatewayRequestHistoryFailureNotFound)
	}
	record, found, err := service.store.ReadRequest(requestContext, requestID)
	if err != nil {
		return gatewayRequestReadFailure(gatewayRequestStoreFailureCode(err))
	}
	if !found {
		return gatewayRequestReadFailure(GatewayRequestHistoryFailureNotFound)
	}
	return GatewayRequestReadResult{Record: &record}
}

func (service gatewayRequestHistoryService) List(
	requestContext GatewayRequestContext,
	request GatewayRequestListRequest,
) GatewayRequestListResult {
	filter, failureCode := normalizeGatewayRequestListRequest(request)
	if failureCode != "" {
		return gatewayRequestListFailure(failureCode)
	}
	if request.Cursor != "" {
		cursor, err := decodeGatewayRequestCursor(request.Cursor, requestContext, filter)
		if err != nil {
			return gatewayRequestListFailure(GatewayRequestHistoryFailureCursorInvalid)
		}
		filter.BeforeTime = &cursor.StartedAt
		filter.BeforeRequestID = cursor.RequestID
	}
	page, err := service.store.ListRequests(requestContext, filter)
	if err != nil {
		return gatewayRequestListFailure(gatewayRequestStoreFailureCode(err))
	}
	nextCursor := ""
	if page.HasMore && len(page.Records) > 0 {
		nextCursor, err = encodeGatewayRequestCursor(page.Records[len(page.Records)-1], requestContext, filter)
		if err != nil {
			return gatewayRequestListFailure(GatewayRequestHistoryFailureContract)
		}
	}
	return GatewayRequestListResult{Records: page.Records, NextCursor: nextCursor, HasMore: page.HasMore}
}

func normalizeGatewayRequestListRequest(
	request GatewayRequestListRequest,
) (GatewayRequestListFilter, GatewayRequestHistoryFailureCode) {
	limit := request.Limit
	if limit == 0 {
		limit = gatewayRequestListDefaultLimit
	}
	if limit < 1 || limit > gatewayRequestListMaxLimit ||
		(request.Status != "" && !validGatewayRequestStatus(request.Status)) ||
		(request.UsageAvailability != "" && !validGatewayRequestUsageAvailability(request.UsageAvailability)) ||
		(request.Protocol != "" && !validGatewayRequestProtocol(request.Protocol)) ||
		(request.FailureBoundary != "" && !validGatewayRequestFailureBoundary(request.FailureBoundary)) {
		return GatewayRequestListFilter{}, GatewayRequestHistoryFailureFilterInvalid
	}
	references := []string{request.Route, request.Provider, request.Profile, request.Model}
	for _, reference := range references {
		if strings.TrimSpace(reference) != "" && !validGatewayRequestReference(reference, 256) {
			return GatewayRequestListFilter{}, GatewayRequestHistoryFailureFilterInvalid
		}
	}
	if request.StartedFrom != nil && request.StartedTo != nil &&
		(request.StartedFrom.After(*request.StartedTo) || request.StartedTo.Sub(*request.StartedFrom) > gatewayRequestListMaxRange) {
		return GatewayRequestListFilter{}, GatewayRequestHistoryFailureFilterInvalid
	}
	nowLimit := time.Now().UTC().Add(gatewayRequestListFutureSkew)
	if (request.StartedFrom != nil && request.StartedFrom.After(nowLimit)) ||
		(request.StartedTo != nil && request.StartedTo.After(nowLimit)) {
		return GatewayRequestListFilter{}, GatewayRequestHistoryFailureFilterInvalid
	}
	return GatewayRequestListFilter{
		Limit: limit, Route: strings.TrimSpace(request.Route), Protocol: strings.TrimSpace(request.Protocol),
		Provider: strings.TrimSpace(request.Provider), Profile: strings.TrimSpace(request.Profile),
		Model: strings.TrimSpace(request.Model), Status: request.Status,
		FailureBoundary: strings.TrimSpace(request.FailureBoundary), UsageAvailability: request.UsageAvailability,
		StartedFrom: request.StartedFrom, StartedTo: request.StartedTo,
	}, ""
}

type gatewayRequestCursor struct {
	Version      int    `json:"v"`
	StartedAt    string `json:"started_at"`
	RequestID    string `json:"request_id"`
	ScopeDigest  string `json:"scope_digest"`
	FilterDigest string `json:"filter_digest"`
}

type decodedGatewayRequestCursor struct {
	StartedAt time.Time
	RequestID string
}

func encodeGatewayRequestCursor(
	record GatewayRequestRecord,
	requestContext GatewayRequestContext,
	filter GatewayRequestListFilter,
) (string, error) {
	if _, err := time.Parse(time.RFC3339Nano, record.StartedAt); err != nil {
		return "", err
	}
	document, err := json.Marshal(gatewayRequestCursor{
		Version: 1, StartedAt: record.StartedAt, RequestID: record.RequestID,
		ScopeDigest: gatewayRequestScopeDigest(requestContext), FilterDigest: gatewayRequestFilterDigest(filter),
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(document), nil
}

func decodeGatewayRequestCursor(
	value string,
	requestContext GatewayRequestContext,
	filter GatewayRequestListFilter,
) (decodedGatewayRequestCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(decoded) > 1024 {
		return decodedGatewayRequestCursor{}, fmt.Errorf("invalid gateway request cursor")
	}
	var cursor gatewayRequestCursor
	decoder := json.NewDecoder(strings.NewReader(string(decoded)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil || cursor.Version != 1 ||
		!validGatewayRequestReference(cursor.RequestID, 160) ||
		cursor.ScopeDigest != gatewayRequestScopeDigest(requestContext) ||
		cursor.FilterDigest != gatewayRequestFilterDigest(filter) {
		return decodedGatewayRequestCursor{}, fmt.Errorf("invalid gateway request cursor")
	}
	startedAt, err := time.Parse(time.RFC3339Nano, cursor.StartedAt)
	if err != nil {
		return decodedGatewayRequestCursor{}, fmt.Errorf("invalid gateway request cursor")
	}
	return decodedGatewayRequestCursor{StartedAt: startedAt, RequestID: cursor.RequestID}, nil
}

func gatewayRequestScopeDigest(requestContext GatewayRequestContext) string {
	return gatewayRequestShortDigest(strings.Join([]string{
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef, requestContext.ApplicationID,
	}, "\n"))
}

func gatewayRequestFilterDigest(filter GatewayRequestListFilter) string {
	return gatewayRequestShortDigest(strings.Join([]string{
		filter.Route, filter.Protocol, filter.Provider, filter.Profile, filter.Model, string(filter.Status),
		filter.FailureBoundary, string(filter.UsageAvailability), gatewayRequestFilterTime(filter.StartedFrom),
		gatewayRequestFilterTime(filter.StartedTo), strconv.Itoa(filter.Limit),
	}, "\n"))
}

func gatewayRequestShortDigest(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:16])
}

func gatewayRequestFilterTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func gatewayRequestReadFailure(code GatewayRequestHistoryFailureCode) GatewayRequestReadResult {
	return GatewayRequestReadResult{FailureCode: code, FailureSummary: gatewayRequestFailureSummary(code)}
}

func gatewayRequestListFailure(code GatewayRequestHistoryFailureCode) GatewayRequestListResult {
	return GatewayRequestListResult{FailureCode: code, FailureSummary: gatewayRequestFailureSummary(code)}
}

func gatewayRequestFailureSummary(code GatewayRequestHistoryFailureCode) string {
	switch code {
	case GatewayRequestHistoryFailureDisabled:
		return "Gateway request history is disabled."
	case GatewayRequestHistoryFailureScopeDenied:
		return "Gateway request history scope is denied."
	case GatewayRequestHistoryFailureNotFound:
		return "Gateway request record was not found."
	case GatewayRequestHistoryFailureCursorInvalid:
		return "Gateway request history cursor is invalid."
	case GatewayRequestHistoryFailureStore, GatewayRequestHistoryFailureContract:
		return "Gateway request history storage is unavailable."
	default:
		return "Gateway request history filter is invalid."
	}
}

func gatewayRequestStoreFailureCode(err error) GatewayRequestHistoryFailureCode {
	if errors.Is(err, errGatewayRequestStoreContract) || errors.Is(err, errGatewayRequestStoreConflict) {
		return GatewayRequestHistoryFailureContract
	}
	return GatewayRequestHistoryFailureStore
}

func validGatewayRequestStatus(status GatewayRequestStatus) bool {
	return status == GatewayRequestStatusStarted || isTerminalGatewayRequestStatus(status)
}

func isTerminalGatewayRequestStatus(status GatewayRequestStatus) bool {
	return status == GatewayRequestStatusSucceeded || status == GatewayRequestStatusFailed || status == GatewayRequestStatusCanceled
}

func validGatewayRequestUsageAvailability(value GatewayRequestUsageAvailability) bool {
	return value == GatewayRequestUsageReported || value == GatewayRequestUsageNotReported || value == GatewayRequestUsageNotApplicable
}

func validGatewayRequestProtocol(value string) bool {
	return value == northboundProtocolModels || value == northboundProtocolChatCompletions || value == northboundProtocolResponses || value == northboundProtocolMessages
}

func validGatewayRequestFailureBoundary(value string) bool {
	switch value {
	case errorBoundaryNorthboundRequest, errorBoundaryCanonicalRequest, errorBoundaryProviderInventory,
		errorBoundaryPythonBridge, errorBoundaryPlatformResponse, errorBoundarySouthboundProvider,
		errorBoundaryGatewayAuth, errorBoundaryConfiguration, errorBoundaryUnknown:
		return true
	default:
		return false
	}
}

func validGatewayRequestReference(value string, maxRunes int) bool {
	value = strings.TrimSpace(value)
	if value == "" || len([]rune(value)) > maxRunes || strings.ContainsAny(value, "\r\n") || strings.Contains(value, "://") {
		return false
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{"authorization", "api_key", "apikey", "token=", "secret=", "password="} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}
