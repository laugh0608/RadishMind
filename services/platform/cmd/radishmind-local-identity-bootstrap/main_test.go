package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/httpapi"
)

func TestBootstrapCommandPropagatesExactPostgresScopeWithoutLeakingDatabaseURL(t *testing.T) {
	const databaseURL = "postgres://bootstrap:secret@localhost/radishmind"
	var received httpapi.LocalIdentityDevTestBootstrapOptions
	runner := func(
		_ context.Context,
		options httpapi.LocalIdentityDevTestBootstrapOptions,
	) (httpapi.LocalIdentityDevTestBootstrapResult, error) {
		received = options
		return httpapi.LocalIdentityDevTestBootstrapResult{
			StoreMode: "postgres_dev_test", TenantRef: options.TenantRef, WorkspaceID: options.WorkspaceID,
			UserID: options.UserID, MembershipID: "mbr_0000000000000001",
			RoleAssignmentID: "rla_0000000000000001", RoleKey: "workspace_admin",
			RoleCatalogVersion:   "local_identity_builtin_roles_v1",
			RoleDefinitionDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			AuditRef:             options.AuditRef,
		}, nil
	}
	getenv := func(key string) string {
		switch key {
		case postgresDatabaseURLEnv:
			return databaseURL
		case databaseTimeoutEnv:
			return "7s"
		default:
			return ""
		}
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{
		"--store", "postgres_dev_test", "--tenant-ref", "tenant_exact", "--workspace-id", "workspace_exact",
		"--user-id", "usr_0000000000000001", "--audit-ref", "audit:bootstrap-exact",
	}, getenv, stdout, stderr, runner)
	if exitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("bootstrap command failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	if received.StoreMode != "postgres_dev_test" || received.PostgresDatabaseURL != databaseURL ||
		received.SQLiteDatabasePath != "" || received.TenantRef != "tenant_exact" ||
		received.WorkspaceID != "workspace_exact" || received.UserID != "usr_0000000000000001" ||
		received.AuditRef != "audit:bootstrap-exact" || received.DatabaseTimeout != 7*time.Second {
		t.Fatalf("bootstrap command changed exact input: %#v", received)
	}
	if strings.Contains(stdout.String(), databaseURL) || strings.Contains(stdout.String(), "secret") {
		t.Fatalf("bootstrap output leaked database configuration: %s", stdout.String())
	}
	var output bootstrapOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil || output.Status != "ok" ||
		!output.DatabaseConfigured || !output.Sanitized || output.Result == nil ||
		output.Result.AuditRef != "audit:bootstrap-exact" {
		t.Fatalf("unexpected bootstrap output: output=%#v err=%v", output, err)
	}
}

func TestBootstrapCommandRejectsMemoryMissingDatabaseAndSanitizesFailure(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
		getenv    func(string) string
		wantCode  string
	}{
		{
			name:      "memory is never a bootstrap store",
			arguments: []string{"--store", "memory_dev"},
			getenv:    func(string) string { return "" },
			wantCode:  httpapi.LocalIdentityFailureContractMismatch,
		},
		{
			name:      "SQLite database must be explicit",
			arguments: []string{"--store", "sqlite_dev"},
			getenv:    func(string) string { return "" },
			wantCode:  httpapi.LocalIdentityFailureStoreUnavailable,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			called := false
			exitCode := run(test.arguments, test.getenv, stdout, stderr, func(
				context.Context,
				httpapi.LocalIdentityDevTestBootstrapOptions,
			) (httpapi.LocalIdentityDevTestBootstrapResult, error) {
				called = true
				return httpapi.LocalIdentityDevTestBootstrapResult{}, nil
			})
			if exitCode != 1 || called || !strings.Contains(stderr.String(), test.wantCode) {
				t.Fatalf("unexpected failure: exit=%d called=%t stderr=%q", exitCode, called, stderr.String())
			}
			var output bootstrapOutput
			if err := json.Unmarshal(stdout.Bytes(), &output); err != nil || output.FailureCode != test.wantCode || !output.Sanitized {
				t.Fatalf("unexpected sanitized failure: output=%#v err=%v", output, err)
			}
		})
	}
}

func TestBootstrapCommandDoesNotPrintRunnerError(t *testing.T) {
	const sensitiveMessage = "postgres://user:secret@example.invalid/database"
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := run([]string{
		"--store", "sqlite_dev", "--tenant-ref", "tenant_demo", "--workspace-id", "workspace_demo",
		"--user-id", "usr_0000000000000001", "--audit-ref", "audit:bootstrap",
	}, func(key string) string {
		if key == sqliteDatabasePathEnv {
			return "/tmp/radishmind-bootstrap.db"
		}
		return ""
	}, stdout, stderr, func(
		context.Context,
		httpapi.LocalIdentityDevTestBootstrapOptions,
	) (httpapi.LocalIdentityDevTestBootstrapResult, error) {
		return httpapi.LocalIdentityDevTestBootstrapResult{}, errors.New(sensitiveMessage)
	})
	if exitCode != 1 || strings.Contains(stdout.String(), sensitiveMessage) || strings.Contains(stderr.String(), sensitiveMessage) {
		t.Fatalf("runner failure leaked: exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
}
