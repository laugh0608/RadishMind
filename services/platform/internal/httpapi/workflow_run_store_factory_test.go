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
