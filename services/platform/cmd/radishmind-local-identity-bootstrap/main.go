package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/httpapi"
)

const (
	sqliteDatabasePathEnv  = "RADISHMIND_SQLITE_DEV_DATABASE_PATH"
	postgresDatabaseURLEnv = "RADISHMIND_LOCAL_IDENTITY_DEV_TEST_DATABASE_URL"
	databaseTimeoutEnv     = "RADISHMIND_LOCAL_IDENTITY_DATABASE_TIMEOUT"
	defaultDatabaseTimeout = 30 * time.Second
)

type bootstrapRunner func(
	context.Context,
	httpapi.LocalIdentityDevTestBootstrapOptions,
) (httpapi.LocalIdentityDevTestBootstrapResult, error)

type bootstrapOutput struct {
	Status             string                                       `json:"status"`
	FailureCode        string                                       `json:"failure_code,omitempty"`
	DatabaseConfigured bool                                         `json:"database_configured"`
	Sanitized          bool                                         `json:"sanitized"`
	Result             *httpapi.LocalIdentityDevTestBootstrapResult `json:"result,omitempty"`
}

func main() {
	os.Exit(run(
		os.Args[1:],
		os.Getenv,
		os.Stdout,
		os.Stderr,
		httpapi.BootstrapLocalIdentityWorkspaceAdministratorDevTest,
	))
}

func run(
	arguments []string,
	getenv func(string) string,
	stdout io.Writer,
	stderr io.Writer,
	runner bootstrapRunner,
) int {
	flags := flag.NewFlagSet("radishmind-local-identity-bootstrap", flag.ContinueOnError)
	flags.SetOutput(stderr)
	storeMode := flags.String("store", "", "development store mode: sqlite_dev or postgres_dev_test")
	tenantRef := flags.String("tenant-ref", "", "exact tenant reference")
	workspaceID := flags.String("workspace-id", "", "exact workspace identifier")
	userID := flags.String("user-id", "", "exact existing active local user identifier")
	auditRef := flags.String("audit-ref", "", "required audit reference")
	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return writeFailure(stdout, stderr, false, httpapi.LocalIdentityFailureContractMismatch)
	}
	if flags.NArg() != 0 {
		return writeFailure(stdout, stderr, false, httpapi.LocalIdentityFailureContractMismatch)
	}

	mode := strings.TrimSpace(*storeMode)
	databaseLocation := ""
	switch mode {
	case "sqlite_dev":
		databaseLocation = strings.TrimSpace(getenv(sqliteDatabasePathEnv))
	case "postgres_dev_test":
		databaseLocation = strings.TrimSpace(getenv(postgresDatabaseURLEnv))
	default:
		return writeFailure(stdout, stderr, false, httpapi.LocalIdentityFailureContractMismatch)
	}
	if databaseLocation == "" {
		return writeFailure(stdout, stderr, false, httpapi.LocalIdentityFailureStoreUnavailable)
	}
	databaseTimeout, err := parseDatabaseTimeout(getenv(databaseTimeoutEnv))
	if err != nil {
		return writeFailure(stdout, stderr, true, httpapi.LocalIdentityFailureContractMismatch)
	}

	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()
	result, err := runner(ctx, httpapi.LocalIdentityDevTestBootstrapOptions{
		StoreMode:           mode,
		SQLiteDatabasePath:  mapDatabaseLocation(mode, "sqlite_dev", databaseLocation),
		PostgresDatabaseURL: mapDatabaseLocation(mode, "postgres_dev_test", databaseLocation),
		DatabaseTimeout:     databaseTimeout,
		TenantRef:           strings.TrimSpace(*tenantRef),
		WorkspaceID:         strings.TrimSpace(*workspaceID),
		UserID:              strings.TrimSpace(*userID),
		AuditRef:            strings.TrimSpace(*auditRef),
	})
	if err != nil {
		return writeFailure(stdout, stderr, true, httpapi.LocalIdentityFailureCode(err))
	}
	if err := writeJSON(stdout, bootstrapOutput{
		Status: "ok", DatabaseConfigured: true, Sanitized: true, Result: &result,
	}); err != nil {
		fmt.Fprintln(stderr, "write local identity bootstrap output failed")
		return 1
	}
	return 0
}

func parseDatabaseTimeout(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultDatabaseTimeout, nil
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout < time.Millisecond {
		return 0, errors.New("invalid local identity database timeout")
	}
	return timeout, nil
}

func mapDatabaseLocation(mode string, expectedMode string, location string) string {
	if mode != expectedMode {
		return ""
	}
	return location
}

func writeFailure(stdout io.Writer, stderr io.Writer, databaseConfigured bool, failureCode string) int {
	if err := writeJSON(stdout, bootstrapOutput{
		Status: "error", FailureCode: failureCode, DatabaseConfigured: databaseConfigured, Sanitized: true,
	}); err != nil {
		fmt.Fprintln(stderr, "write local identity bootstrap output failed")
		return 1
	}
	fmt.Fprintf(stderr, "local identity bootstrap failed: %s\n", failureCode)
	return 1
}

func writeJSON(writer io.Writer, output bootstrapOutput) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
