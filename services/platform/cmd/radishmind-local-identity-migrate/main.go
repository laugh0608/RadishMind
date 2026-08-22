package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	localidentitymigrations "radishmind.local/services/platform/migrations/local_identity_records"
)

const migrationDatabaseURLEnv = "RADISHMIND_LOCAL_IDENTITY_DEV_TEST_MIGRATION_DATABASE_URL"

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
	databaseURL := strings.TrimSpace(os.Getenv(migrationDatabaseURLEnv))
	if databaseURL == "" {
		fail(action, false, "local identity migration database URL is missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := localidentitymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		fail(action, true, err.Error())
	}
	defer pool.Close()
	var state localidentitymigrations.State
	if action == "up" {
		state, err = localidentitymigrations.Apply(ctx, pool)
	} else {
		state, err = localidentitymigrations.Inspect(ctx, pool)
	}
	if err != nil {
		fail(action, true, err.Error())
	}
	status := "ok"
	if state.MigrationState != localidentitymigrations.MigrationStateApplied {
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
	writeOutput(migrationOutput{
		Status: "error", Action: action, MigrationState: "unavailable",
		DatabaseConfigured: databaseConfigured, Sanitized: true,
	})
	fmt.Fprintln(os.Stderr, strings.TrimSpace(message))
	os.Exit(1)
}

func writeOutput(output migrationOutput) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, "write local identity migration output failed")
		os.Exit(1)
	}
}
