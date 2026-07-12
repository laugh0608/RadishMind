package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

const (
	runtimeDatabaseURLEnv   = "RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL"
	migrationDatabaseURLEnv = "RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL"
)

type output struct {
	Status             string `json:"status"`
	Action             string `json:"action"`
	MigrationState     string `json:"migration_state"`
	MigrationID        string `json:"migration_id,omitempty"`
	StoreSchemaVersion string `json:"store_schema_version,omitempty"`
	MigrationChecksum  string `json:"migration_checksum,omitempty"`
	DatabaseConfigured bool   `json:"database_configured"`
	Sanitized          bool   `json:"sanitized"`
}

func main() {
	action := "status"
	if len(os.Args) > 1 {
		action = strings.TrimSpace(os.Args[1])
	}
	if action != "status" && action != "up" {
		fail(action, false, "unsupported migration action; expected status or up")
	}
	if strings.TrimSpace(os.Getenv("RADISHMIND_WORKFLOW_RUN_STORE")) != "postgres_dev_test" {
		fail(action, false, "workflow run migration requires postgres_dev_test store mode")
	}
	databaseURL := strings.TrimSpace(os.Getenv(runtimeDatabaseURLEnv))
	if action == "up" {
		databaseURL = strings.TrimSpace(os.Getenv(migrationDatabaseURLEnv))
	}
	if databaseURL == "" {
		fail(action, false, "workflow run migration database URL is missing")
	}
	timeout := 5 * time.Second
	if raw := strings.TrimSpace(os.Getenv("RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			fail(action, true, "workflow run database timeout is invalid")
		}
		timeout = parsed
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		fail(action, true, err.Error())
	}
	defer pool.Close()
	var state workflowrunmigrations.State
	if action == "up" {
		state, err = workflowrunmigrations.Apply(ctx, pool)
	} else {
		state, err = workflowrunmigrations.Inspect(ctx, pool)
	}
	if err != nil {
		fail(action, true, err.Error())
	}
	status := "ok"
	if state.MigrationState != workflowrunmigrations.MigrationStateApplied {
		status = "not_ready"
	}
	write(output{Status: status, Action: action, MigrationState: state.MigrationState, MigrationID: state.MigrationID, StoreSchemaVersion: state.StoreSchemaVersion, MigrationChecksum: state.MigrationChecksum, DatabaseConfigured: true, Sanitized: true})
	if status != "ok" {
		os.Exit(1)
	}
}

func fail(action string, configured bool, message string) {
	write(output{Status: "error", Action: action, MigrationState: "unavailable", DatabaseConfigured: configured, Sanitized: true})
	fmt.Fprintln(os.Stderr, strings.TrimSpace(message))
	os.Exit(1)
}
func write(value output) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		os.Exit(1)
	}
}
