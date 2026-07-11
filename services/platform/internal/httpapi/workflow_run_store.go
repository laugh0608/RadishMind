package httpapi

import (
	"sync"
)

const defaultWorkflowRunStoreCapacity = 100

type workflowRunStore interface {
	UpsertRun(context WorkflowRunContext, record WorkflowRunRecord) error
	ReadRun(context WorkflowRunContext, runID string) (WorkflowRunRecord, bool, error)
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
	context WorkflowRunContext,
	record WorkflowRunRecord,
) error {
	key := workflowRunStoreKey(context.TenantRef, context.WorkspaceID, context.ApplicationID, record.RunID)
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, found := store.records[key]; !found {
		store.order = append(store.order, key)
	}
	store.records[key] = cloneWorkflowRunRecord(record)
	for len(store.order) > store.capacity {
		oldestKey := store.order[0]
		store.order = store.order[1:]
		delete(store.records, oldestKey)
	}
	return nil
}

func (store *memoryWorkflowRunStore) ReadRun(
	context WorkflowRunContext,
	runID string,
) (WorkflowRunRecord, bool, error) {
	key := workflowRunStoreKey(context.TenantRef, context.WorkspaceID, context.ApplicationID, runID)
	store.mu.RLock()
	defer store.mu.RUnlock()
	record, found := store.records[key]
	if !found {
		return WorkflowRunRecord{}, false, nil
	}
	return cloneWorkflowRunRecord(record), true, nil
}

func workflowRunStoreKey(tenantRef string, workspaceID string, applicationID string, runID string) string {
	return tenantRef + "\x00" + workspaceID + "\x00" + applicationID + "\x00" + runID
}

func cloneWorkflowRunRecord(record WorkflowRunRecord) WorkflowRunRecord {
	cloned := record
	cloned.ConditionNodeIDs = cloneStringSlice(record.ConditionNodeIDs)
	cloned.Nodes = make([]WorkflowRunNodeRecord, 0, len(record.Nodes))
	for _, node := range record.Nodes {
		clonedNode := node
		clonedNode.PredecessorNodeIDs = cloneStringSlice(node.PredecessorNodeIDs)
		cloned.Nodes = append(cloned.Nodes, clonedNode)
	}
	return cloned
}
