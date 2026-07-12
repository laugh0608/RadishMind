package httpapi

import (
	"radishmind.local/services/platform/internal/config"
	"testing"
	"time"
)

func TestWorkflowRunStoreFactoryNoFallback(t *testing.T) {
	cfg := config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowExecutorDevEnabled: true, WorkflowRunStoreMode: "postgres_dev_test", WorkflowRunDatabaseURL: "postgresql://invalid:invalid@127.0.0.1:1/invalid", WorkflowRunDatabaseTimeout: 50 * time.Millisecond}
	store, closeStore, err := newWorkflowRunStoreFromConfig(cfg)
	if err == nil || store != nil || closeStore != nil {
		t.Fatalf("database failure must not select memory: store=%T err=%v", store, err)
	}
}

func TestWorkflowRunStoreFactoryModesFailClosed(t *testing.T) {
	for _, mode := range []string{"repository", "repository_disabled", "unknown"} {
		store, closeStore, err := newWorkflowRunStoreFromConfig(config.Config{WorkflowRunStoreMode: mode})
		if err == nil || store != nil || closeStore != nil {
			t.Fatalf("mode %s must fail closed", mode)
		}
	}
}

func TestWorkflowEvaluationStoresFollowExplicitRunStoreSelection(t *testing.T) {
	memoryRunStore := newMemoryWorkflowRunStore(10)
	if _, ok := newWorkflowEvaluationStoreForRunStore(memoryRunStore).(*memoryWorkflowEvaluationStore); !ok {
		t.Fatal("memory_dev case store selection drifted")
	}
	if _, ok := newWorkflowEvaluationSuiteStoreForRunStore(memoryRunStore).(*memoryWorkflowEvaluationSuiteStore); !ok {
		t.Fatal("memory_dev suite store selection drifted")
	}
	postgresRunStore := newPostgresWorkflowRunStore(nil)
	if _, ok := newWorkflowEvaluationStoreForRunStore(postgresRunStore).(*postgresWorkflowEvaluationStore); !ok {
		t.Fatal("postgres_dev_test case store selection drifted")
	}
	if _, ok := newWorkflowEvaluationSuiteStoreForRunStore(postgresRunStore).(*postgresWorkflowEvaluationSuiteStore); !ok {
		t.Fatal("postgres_dev_test suite store selection drifted")
	}
}
