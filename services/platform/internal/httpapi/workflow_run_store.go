package httpapi

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultWorkflowRunStoreCapacity = 100

var (
	errWorkflowRunStoreContract = errors.New("workflow run store contract mismatch")
	errWorkflowRunStoreConflict = errors.New("workflow run store record version conflict")
)

type WorkflowRunListFilter struct {
	Status          WorkflowRunStatus
	DraftID         string
	FailureCode     WorkflowRunFailureCode
	FailureBoundary WorkflowRunFailureBoundary
	Provider        string
	Model           string
	StaleRunning    *bool
	StartedFrom     *time.Time
	StartedTo       *time.Time
	BeforeTime      *time.Time
	BeforeRunID     string
	Limit           int
}

type WorkflowRunListPage struct {
	Records []WorkflowRunRecord
	HasMore bool
}

type workflowRunStore interface {
	UpsertRun(context WorkflowRunContext, record *WorkflowRunRecord) error
	ReadRun(context WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error)
	ListRuns(context WorkflowRunContext, filter WorkflowRunListFilter) (WorkflowRunListPage, error)
}

type memoryWorkflowRunStore struct {
	mu       sync.RWMutex
	capacity int
	records  map[string]WorkflowRunRecord
	order    []string
}

func newMemoryWorkflowRunStore(capacity int) *memoryWorkflowRunStore {
	if capacity <= 0 {
		capacity = defaultWorkflowRunStoreCapacity
	}
	return &memoryWorkflowRunStore{
		capacity: capacity,
		records:  make(map[string]WorkflowRunRecord, capacity),
		order:    make([]string, 0, capacity),
	}
}

func (store *memoryWorkflowRunStore) UpsertRun(
	runContext WorkflowRunContext,
	record *WorkflowRunRecord,
) error {
	if err := validateWorkflowRunStoreRecord(runContext, record); err != nil {
		return err
	}
	key := workflowRunStoreKey(runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, record.RunID)
	store.mu.Lock()
	defer store.mu.Unlock()
	current, found := store.records[key]
	if !found {
		if record.RecordVersion != 0 || record.Status != WorkflowRunStatusRunning {
			return errWorkflowRunStoreConflict
		}
		record.RecordVersion = 1
		store.order = append(store.order, key)
	} else {
		if record.RecordVersion != current.RecordVersion || isTerminalWorkflowRunStatus(current.Status) {
			return errWorkflowRunStoreConflict
		}
		record.RecordVersion++
	}
	store.records[key] = cloneWorkflowRunRecord(*record)
	for len(store.order) > store.capacity {
		oldestKey := store.order[0]
		store.order = store.order[1:]
		delete(store.records, oldestKey)
	}
	return nil
}

func (store *memoryWorkflowRunStore) ReadRun(
	runContext WorkflowRunContext,
	runID string,
) (WorkflowRunRecord, bool, error) {
	key := workflowRunStoreKey(runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, runID)
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, found := store.records[key]
	if !found {
		return WorkflowRunRecord{}, false, nil
	}
	return cloneWorkflowRunRecord(record), true, nil
}

func (store *memoryWorkflowRunStore) ListRuns(
	runContext WorkflowRunContext,
	filter WorkflowRunListFilter,
) (WorkflowRunListPage, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	records := make([]WorkflowRunRecord, 0)
	prefix := workflowRunStoreKey(runContext.TenantRef, runContext.WorkspaceID, runContext.ApplicationID, "")
	for key, record := range store.records {
		if strings.HasPrefix(key, prefix) && workflowRunMatchesFilter(record, filter) {
			records = append(records, cloneWorkflowRunRecord(record))
		}
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].StartedAt == records[right].StartedAt {
			return records[left].RunID > records[right].RunID
		}
		return records[left].StartedAt > records[right].StartedAt
	})
	limit := filter.Limit
	if limit <= 0 {
		limit = 25
	}
	hasMore := len(records) > limit
	if hasMore {
		records = records[:limit]
	}
	return WorkflowRunListPage{Records: records, HasMore: hasMore}, nil
}

func workflowRunMatchesFilter(record WorkflowRunRecord, filter WorkflowRunListFilter) bool {
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	if filter.DraftID != "" && record.DraftID != filter.DraftID {
		return false
	}
	if filter.FailureCode != "" && record.FailureCode != filter.FailureCode {
		return false
	}
	if filter.FailureBoundary != "" && (record.Diagnostic == nil || record.Diagnostic.FailureBoundary != filter.FailureBoundary) {
		return false
	}
	if filter.Provider != "" && record.SelectedProvider != filter.Provider {
		return false
	}
	if filter.Model != "" && record.SelectedModel != filter.Model {
		return false
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return false
	}
	if filter.StartedFrom != nil && startedAt.Before(*filter.StartedFrom) {
		return false
	}
	if filter.StartedTo != nil && startedAt.After(*filter.StartedTo) {
		return false
	}
	if filter.BeforeTime != nil {
		if startedAt.After(*filter.BeforeTime) || (startedAt.Equal(*filter.BeforeTime) && record.RunID >= filter.BeforeRunID) {
			return false
		}
	}
	if filter.StaleRunning != nil {
		stale := record.Status == WorkflowRunStatusRunning && time.Since(startedAt) > workflowExecutorDefaultMaxRuntime
		if stale != *filter.StaleRunning {
			return false
		}
	}
	return true
}

func validateWorkflowRunStoreRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || strings.TrimSpace(runContext.TenantRef) == "" ||
		strings.TrimSpace(runContext.WorkspaceID) == "" || strings.TrimSpace(runContext.ApplicationID) == "" ||
		!validWorkflowRunRecordSchema(record.SchemaVersion) || strings.TrimSpace(record.RunID) == "" ||
		strings.TrimSpace(record.DraftID) == "" || record.DraftVersion <= 0 || record.RecordVersion < 0 ||
		strings.TrimSpace(record.ActorRef) == "" || strings.TrimSpace(record.RequestID) == "" || strings.TrimSpace(record.AuditRef) == "" ||
		record.WorkspaceID != runContext.WorkspaceID || record.ApplicationID != runContext.ApplicationID ||
		!validWorkflowRunStatus(record.Status) || record.SideEffects.ToolCalls != 0 ||
		record.SideEffects.ConfirmationCalls != 0 || record.SideEffects.BusinessWrites != 0 ||
		record.SideEffects.ReplayWrites != 0 || len([]byte(record.Output)) > workflowExecutorMaxOutputBytes ||
		len([]rune(record.FailureSummary)) > 256 || workflowRunRecordContainsEndpoint(record) {
		return errWorkflowRunStoreContract
	}
	if record.SchemaVersion == workflowRunRecordSchemaVersion && !validWorkflowRunDiagnostic(record.Diagnostic, isTerminalWorkflowRunStatus(record.Status)) {
		return errWorkflowRunStoreContract
	}
	if record.SchemaVersion == workflowRunRecordLegacySchemaVersion && record.Diagnostic != nil {
		return errWorkflowRunStoreContract
	}
	for _, node := range record.Nodes {
		if len([]rune(node.OutputPreview)) > workflowExecutorNodePreviewRunes+1 || strings.Contains(node.ProviderRef, "://") {
			return errWorkflowRunStoreContract
		}
	}
	if _, err := time.Parse(time.RFC3339Nano, record.StartedAt); err != nil {
		return errWorkflowRunStoreContract
	}
	if isTerminalWorkflowRunStatus(record.Status) {
		if _, err := time.Parse(time.RFC3339Nano, record.CompletedAt); err != nil {
			return errWorkflowRunStoreContract
		}
	} else if record.CompletedAt != "" {
		return errWorkflowRunStoreContract
	}
	return nil
}

func workflowRunRecordContainsEndpoint(record *WorkflowRunRecord) bool {
	return strings.Contains(record.RequestedModel, "://") || strings.Contains(record.SelectedProvider, "://") ||
		strings.Contains(record.SelectedProfile, "://") || strings.Contains(record.SelectedModel, "://") ||
		strings.Contains(record.UpstreamModel, "://")
}

func validWorkflowRunRecordSchema(schemaVersion string) bool {
	return schemaVersion == workflowRunRecordSchemaVersion || schemaVersion == workflowRunRecordLegacySchemaVersion
}

func validWorkflowRunStatus(status WorkflowRunStatus) bool {
	return status == WorkflowRunStatusRunning || isTerminalWorkflowRunStatus(status)
}

func isTerminalWorkflowRunStatus(status WorkflowRunStatus) bool {
	return status == WorkflowRunStatusSucceeded || status == WorkflowRunStatusFailed || status == WorkflowRunStatusCanceled
}

func workflowRunStoreKey(tenantRef string, workspaceID string, applicationID string, runID string) string {
	return tenantRef + "\x00" + workspaceID + "\x00" + applicationID + "\x00" + runID
}

func cloneWorkflowRunRecord(record WorkflowRunRecord) WorkflowRunRecord {
	cloned := record
	if record.Diagnostic != nil {
		diagnostic := *record.Diagnostic
		cloned.Diagnostic = &diagnostic
	}
	cloned.ConditionNodeIDs = cloneStringSlice(record.ConditionNodeIDs)
	cloned.Nodes = make([]WorkflowRunNodeRecord, 0, len(record.Nodes))
	for _, node := range record.Nodes {
		clonedNode := node
		clonedNode.PredecessorNodeIDs = cloneStringSlice(node.PredecessorNodeIDs)
		cloned.Nodes = append(cloned.Nodes, clonedNode)
	}
	return cloned
}
