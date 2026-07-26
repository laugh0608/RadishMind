package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	agentcopilotprofilemigrations "radishmind.local/services/platform/migrations/agent_copilot_profiles"
)

const migrationDatabaseURLEnv = "RADISHMIND_AGENT_COPILOT_PROFILE_DEV_TEST_MIGRATION_DATABASE_URL"

type migrationOutput struct {
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
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fail(action, false, "load platform configuration failed")
	}
	if strings.TrimSpace(cfg.AgentCopilotProfileStoreMode) != "postgres_dev_test" || !cfg.ControlPlaneReadDevAuthEnabled ||
		!cfg.AgentCopilotProfileDevHTTPEnabled || !cfg.AgentCopilotProfileDevWriteEnabled {
		fail(action, strings.TrimSpace(cfg.AgentCopilotProfileDatabaseURL) != "", "agent copilot profile migration requires complete postgres_dev_test development gates")
	}
	databaseURL := strings.TrimSpace(os.Getenv(migrationDatabaseURLEnv))
	if action == "status" && databaseURL == "" {
		databaseURL = strings.TrimSpace(cfg.AgentCopilotProfileDatabaseURL)
	}
	if databaseURL == "" {
		fail(action, false, "agent copilot profile migration database URL is missing")
	}
	timeout := cfg.AgentCopilotProfileDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := agentcopilotprofilemigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		fail(action, true, err.Error())
	}
	defer pool.Close()
	var state agentcopilotprofilemigrations.State
	if action == "up" {
		state, err = agentcopilotprofilemigrations.Apply(ctx, pool)
	} else {
		state, err = agentcopilotprofilemigrations.Inspect(ctx, pool)
	}
	if err != nil {
		fail(action, true, err.Error())
	}
	status := "ok"
	if state.MigrationState != agentcopilotprofilemigrations.MigrationStateApplied {
		status = "not_ready"
	}
	writeOutput(migrationOutput{
		Status: status, Action: action, MigrationState: state.MigrationState, MigrationID: state.MigrationID,
		StoreSchemaVersion: state.StoreSchemaVersion, MigrationChecksum: state.MigrationChecksum,
		DatabaseConfigured: true, Sanitized: true,
	})
	if status != "ok" {
		os.Exit(1)
	}
}

func fail(action string, databaseConfigured bool, message string) {
	writeOutput(migrationOutput{Status: "error", Action: action, MigrationState: "unavailable", DatabaseConfigured: databaseConfigured, Sanitized: true})
	fmt.Fprintln(os.Stderr, strings.TrimSpace(message))
	os.Exit(1)
}

func writeOutput(output migrationOutput) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, "write agent copilot profile migration output failed")
		os.Exit(1)
	}
}
