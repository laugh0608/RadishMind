package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	apikeymigrations "radishmind.local/services/platform/migrations/api_key_records"
)

const migrationDatabaseURLEnv = "RADISHMIND_API_KEY_DEV_TEST_MIGRATION_DATABASE_URL"

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
	if strings.TrimSpace(cfg.APIKeyStoreMode) != "postgres_dev_test" || !cfg.ControlPlaneReadDevAuthEnabled ||
		!cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.APIKeyLifecycleDevWriteEnabled {
		fail(action, strings.TrimSpace(cfg.APIKeyDatabaseURL) != "", "API key migration requires complete postgres_dev_test development gates")
	}
	databaseURL := strings.TrimSpace(os.Getenv(migrationDatabaseURLEnv))
	if action == "status" && databaseURL == "" {
		databaseURL = strings.TrimSpace(cfg.APIKeyDatabaseURL)
	}
	if databaseURL == "" {
		fail(action, false, "API key migration database URL is missing")
	}
	timeout := cfg.APIKeyDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := apikeymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		fail(action, true, err.Error())
	}
	defer pool.Close()
	var state apikeymigrations.State
	if action == "up" {
		state, err = apikeymigrations.Apply(ctx, pool)
	} else {
		state, err = apikeymigrations.Inspect(ctx, pool)
	}
	if err != nil {
		fail(action, true, err.Error())
	}
	status := "ok"
	if state.MigrationState != apikeymigrations.MigrationStateApplied {
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
		fmt.Fprintln(os.Stderr, "write API key migration output failed")
		os.Exit(1)
	}
}
