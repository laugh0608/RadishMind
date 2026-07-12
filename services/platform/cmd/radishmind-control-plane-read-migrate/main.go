package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	controlplanereadmigrations "radishmind.local/services/platform/migrations/control_plane_admin_read"
)

const migrationDatabaseURLEnv = "RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL"

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
	if config.EffectiveControlPlaneReadAuthMode(cfg) != "signed_test_token" ||
		config.EffectiveControlPlaneReadStoreMode(cfg) != "postgres_dev_test" {
		fail(action, strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL) != "", "control plane read migration requires signed_test_token and postgres_dev_test modes")
	}
	databaseURL := strings.TrimSpace(os.Getenv(migrationDatabaseURLEnv))
	if action == "status" && databaseURL == "" {
		databaseURL = strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL)
	}
	if databaseURL == "" {
		fail(action, false, "control plane read migration database URL is missing")
	}
	timeout := cfg.ControlPlaneReadDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := controlplanereadmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		fail(action, true, err.Error())
	}
	defer pool.Close()
	var state controlplanereadmigrations.State
	if action == "up" {
		state, err = controlplanereadmigrations.Apply(ctx, pool)
	} else {
		state, err = controlplanereadmigrations.Inspect(ctx, pool)
	}
	if err != nil {
		fail(action, true, err.Error())
	}
	status := "ok"
	if state.MigrationState != controlplanereadmigrations.MigrationStateApplied {
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
		fmt.Fprintln(os.Stderr, "write control plane read migration output failed")
		os.Exit(1)
	}
}
