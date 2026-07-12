package httpapi

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultGatewayRequestStoreCapacity = 500

var (
	errGatewayRequestStoreContract = errors.New("gateway request store contract mismatch")
	errGatewayRequestStoreConflict = errors.New("gateway request store record version conflict")
)

type GatewayRequestListFilter struct {
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
	BeforeTime        *time.Time
	BeforeRequestID   string
	Limit             int
}

type GatewayRequestListPage struct {
	Records []GatewayRequestRecord
	HasMore bool
}

type gatewayRequestStore interface {
	CreateRequest(context GatewayRequestContext, record *GatewayRequestRecord) error
	UpdateRequest(context GatewayRequestContext, record *GatewayRequestRecord) error
	ReadRequest(context GatewayRequestContext, requestID string) (GatewayRequestRecord, bool, error)
	ListRequests(context GatewayRequestContext, filter GatewayRequestListFilter) (GatewayRequestListPage, error)
}

type memoryGatewayRequestStore struct {
	mu       sync.RWMutex
	capacity int
	records  map[string]GatewayRequestRecord
	order    []string
}

func newMemoryGatewayRequestStore(capacity int) *memoryGatewayRequestStore {
	if capacity <= 0 {
		capacity = defaultGatewayRequestStoreCapacity
	}
	return &memoryGatewayRequestStore{
		capacity: capacity,
		records:  make(map[string]GatewayRequestRecord, capacity),
		order:    make([]string, 0, capacity),
	}
}

func (store *memoryGatewayRequestStore) CreateRequest(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil || record.RecordVersion != 0 || record.Status != GatewayRequestStatusStarted {
		return errGatewayRequestStoreContract
	}
	key := gatewayRequestStoreKey(requestContext, record.RequestID)
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, found := store.records[key]; found {
		return errGatewayRequestStoreConflict
	}
	record.RecordVersion = 1
	store.records[key] = cloneGatewayRequestRecord(*record)
	store.order = append(store.order, key)
	for len(store.order) > store.capacity {
		oldestKey := store.order[0]
		store.order = store.order[1:]
		delete(store.records, oldestKey)
	}
	return nil
}

func (store *memoryGatewayRequestStore) UpdateRequest(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if err := validateGatewayRequestStoreRecord(requestContext, record); err != nil || record.RecordVersion < 1 {
		return errGatewayRequestStoreContract
	}
	key := gatewayRequestStoreKey(requestContext, record.RequestID)
	store.mu.Lock()
	defer store.mu.Unlock()
	current, found := store.records[key]
	if !found || current.RecordVersion != record.RecordVersion || isTerminalGatewayRequestStatus(current.Status) ||
		(current.Status != GatewayRequestStatusStarted) {
		return errGatewayRequestStoreConflict
	}
	record.RecordVersion++
	store.records[key] = cloneGatewayRequestRecord(*record)
	return nil
}

func (store *memoryGatewayRequestStore) ReadRequest(
	requestContext GatewayRequestContext,
	requestID string,
) (GatewayRequestRecord, bool, error) {
	key := gatewayRequestStoreKey(requestContext, requestID)
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, found := store.records[key]
	if !found {
		return GatewayRequestRecord{}, false, nil
	}
	return cloneGatewayRequestRecord(record), true, nil
}

func (store *memoryGatewayRequestStore) ListRequests(
	requestContext GatewayRequestContext,
	filter GatewayRequestListFilter,
) (GatewayRequestListPage, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	records := make([]GatewayRequestRecord, 0)
	prefix := gatewayRequestStoreKey(requestContext, "")
	for key, record := range store.records {
		if strings.HasPrefix(key, prefix) && gatewayRequestMatchesFilter(record, filter) {
			records = append(records, cloneGatewayRequestRecord(record))
		}
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].StartedAt == records[right].StartedAt {
			return records[left].RequestID > records[right].RequestID
		}
		return records[left].StartedAt > records[right].StartedAt
	})
	limit := filter.Limit
	if limit <= 0 {
		limit = gatewayRequestListDefaultLimit
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return GatewayRequestListPage{Records: records, HasMore: hasMore}, nil
}

func gatewayRequestMatchesFilter(record GatewayRequestRecord, filter GatewayRequestListFilter) bool {
	if filter.Route != "" && record.Route != filter.Route ||
		filter.Protocol != "" && record.Protocol != filter.Protocol ||
		filter.Provider != "" && record.SelectedProvider != filter.Provider ||
		filter.Profile != "" && record.SelectedProfile != filter.Profile ||
		filter.Model != "" && record.SelectedModel != filter.Model ||
		filter.Status != "" && record.Status != filter.Status ||
		filter.FailureBoundary != "" && record.FailureBoundary != filter.FailureBoundary ||
		filter.UsageAvailability != "" && record.Usage.Availability != filter.UsageAvailability {
		return false
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil || filter.StartedFrom != nil && startedAt.Before(*filter.StartedFrom) ||
		filter.StartedTo != nil && startedAt.After(*filter.StartedTo) {
		return false
	}
	if filter.BeforeTime != nil &&
		(startedAt.After(*filter.BeforeTime) || startedAt.Equal(*filter.BeforeTime) && record.RequestID >= filter.BeforeRequestID) {
		return false
	}
	return true
}

func validateGatewayRequestStoreRecord(
	requestContext GatewayRequestContext,
	record *GatewayRequestRecord,
) error {
	if record == nil || !validGatewayRequestContext(requestContext) ||
		record.SchemaVersion != gatewayRequestRecordSchemaVersion || record.RecordVersion < 0 ||
		record.TenantRef != requestContext.TenantRef || record.WorkspaceID != requestContext.WorkspaceID ||
		record.ConsumerRef != requestContext.ConsumerRef || record.ApplicationID != requestContext.ApplicationID ||
		record.SubjectRef != requestContext.SubjectRef || !validGatewayRequestReference(record.RequestID, 160) ||
		!validGatewayRequestReference(record.AuditRef, 256) || !validGatewayRequestReference(record.Route, 128) ||
		!validGatewayRequestProtocol(record.Protocol) || !validGatewayRequestStatus(record.Status) ||
		!validGatewayRequestUsage(record.Usage) || record.DurationMS < 0 || record.GatewayDurationMS < 0 ||
		record.ProviderDurationMS < 0 || record.HTTPStatusCode < 0 || record.HTTPStatusCode > 599 {
		return errGatewayRequestStoreContract
	}
	for _, reference := range []string{record.SelectionSource, record.SelectedProvider, record.SelectedProfile, record.SelectedModel} {
		if reference != "" && !validGatewayRequestReference(reference, 256) {
			return errGatewayRequestStoreContract
		}
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil || startedAt.IsZero() {
		return errGatewayRequestStoreContract
	}
	if record.Status == GatewayRequestStatusStarted {
		if record.CompletedAt != "" || record.FailureCode != "" || record.FailureBoundary != "" {
			return errGatewayRequestStoreContract
		}
		return nil
	}
	completedAt, err := time.Parse(time.RFC3339Nano, record.CompletedAt)
	if err != nil || completedAt.Before(startedAt) {
		return errGatewayRequestStoreContract
	}
	if record.Status == GatewayRequestStatusSucceeded {
		if record.FailureCode != "" || record.FailureBoundary != "" || record.HTTPStatusCode < 200 || record.HTTPStatusCode >= 400 {
			return errGatewayRequestStoreContract
		}
	} else if !validGatewayRequestReference(record.FailureCode, 160) || !validGatewayRequestFailureBoundary(record.FailureBoundary) {
		return errGatewayRequestStoreContract
	}
	return nil
}

func validGatewayRequestContext(requestContext GatewayRequestContext) bool {
	for _, reference := range []string{
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef,
		requestContext.SubjectRef, requestContext.AuditContext, requestContext.Source,
	} {
		if !validGatewayRequestReference(reference, 256) {
			return false
		}
	}
	return requestContext.ApplicationID == "" || validGatewayRequestReference(requestContext.ApplicationID, 256)
}

func validGatewayRequestUsage(usage GatewayRequestUsage) bool {
	if !validGatewayRequestUsageAvailability(usage.Availability) || usage.InputTokens < 0 || usage.OutputTokens < 0 || usage.TotalTokens < 0 {
		return false
	}
	if usage.Availability == GatewayRequestUsageReported {
		return validGatewayRequestReference(usage.Source, 128) && usage.TotalTokens == usage.InputTokens+usage.OutputTokens
	}
	return usage.Source == "" && usage.InputTokens == 0 && usage.OutputTokens == 0 && usage.TotalTokens == 0
}

func gatewayRequestStoreKey(requestContext GatewayRequestContext, requestID string) string {
	return strings.Join([]string{
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ConsumerRef,
		requestContext.ApplicationID, requestID,
	}, "\x00")
}

func cloneGatewayRequestRecord(record GatewayRequestRecord) GatewayRequestRecord {
	return record
}
