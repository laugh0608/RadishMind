//go:build postgres_integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/config"
	workflowdraftmigrations "radishmind.local/services/platform/migrations/workflow_saved_drafts"
)

func TestSavedWorkflowDraftPostgresDevTestRepository(t *testing.T) {
	migrationDatabaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required for PostgreSQL integration tests")
	}
	if runtimeUser == strings.TrimSpace(os.Getenv("PGUSER")) {
		t.Fatal("PostgreSQL integration runtime user must differ from the migration user")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t,
		runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	databaseContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	adminPool, err := workflowdraftmigrations.OpenPool(databaseContext, migrationDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL integration database: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, databaseContext, adminPool)
	resetPostgresSavedWorkflowDraftSchema(t, databaseContext, adminPool)
	preparePostgresIntegrationRuntimeRole(t, databaseContext, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresSavedWorkflowDraftSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})

	cfg := postgresSavedWorkflowDraftIntegrationConfig(runtimeDatabaseURL)
	if _, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-integration"}); err == nil || !strings.Contains(err.Error(), "migration is not applied") {
		t.Fatalf("server must fail before explicit migration, got: %v", err)
	}

	firstMigration, err := workflowdraftmigrations.Apply(databaseContext, adminPool)
	if err != nil {
		t.Fatalf("apply first migration: %v", err)
	}
	secondMigration, err := workflowdraftmigrations.Apply(databaseContext, adminPool)
	if err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	if firstMigration.MigrationState != workflowdraftmigrations.MigrationStateApplied ||
		secondMigration.MigrationState != workflowdraftmigrations.MigrationStateApplied ||
		firstMigration.MigrationChecksum != secondMigration.MigrationChecksum ||
		firstMigration.MigrationID != secondMigration.MigrationID {
		t.Fatalf("migration apply must be idempotent: first=%#v second=%#v", firstMigration, secondMigration)
	}
	assertPostgresIntegrationRuntimeRoleCannotMigrate(
		t,
		databaseContext,
		adminPool,
		runtimeDatabaseURL,
	)

	firstServer := newPostgresSavedWorkflowDraftIntegrationServer(t, cfg)
	payload := validSavedWorkflowDraftPayload()
	firstSave := postPostgresSavedWorkflowDraft(t, firstServer, payload, 0, "subject_platform_ops", "tenant_demo")
	if firstSave.FailureCode != nil || firstSave.Draft == nil || firstSave.CurrentDraftVersion != 1 {
		t.Fatalf("first PostgreSQL save failed: %#v", firstSave)
	}
	firstServer.Close()

	secondServer := newPostgresSavedWorkflowDraftIntegrationServer(t, cfg)
	restored := readPostgresSavedWorkflowDraft(
		t,
		secondServer,
		payload.DraftID,
		"workspace_demo",
		"app_flow_copilot",
		"subject_platform_ops",
		"tenant_demo",
	)
	if restored.FailureCode != nil || restored.Draft == nil ||
		restored.Draft.DraftVersion != 1 || restored.Draft.Name != payload.Name {
		t.Fatalf("new server did not restore persisted draft: %#v", restored)
	}
	listed := listPostgresSavedWorkflowDrafts(
		t,
		secondServer,
		"workspace_demo",
		"app_flow_copilot",
		"subject_platform_ops",
		"tenant_demo",
	)
	if listed.FailureCode != nil || len(listed.DraftSummaries) != 1 ||
		listed.DraftSummaries[0].DraftID != payload.DraftID {
		t.Fatalf("new server did not list persisted draft: %#v", listed)
	}

	const concurrentWriters = 16
	results := make(chan savedWorkflowDraftEnvelope, concurrentWriters)
	var writers sync.WaitGroup
	for writerIndex := 0; writerIndex < concurrentWriters; writerIndex++ {
		writerIndex := writerIndex
		writers.Add(1)
		go func() {
			defer writers.Done()
			candidate := payload
			candidate.Name = fmt.Sprintf("concurrent candidate %02d", writerIndex)
			results <- postPostgresSavedWorkflowDraft(
				t,
				secondServer,
				candidate,
				1,
				"subject_platform_ops",
				"tenant_demo",
			)
		}()
	}
	writers.Wait()
	close(results)
	successCount := 0
	conflictCount := 0
	for result := range results {
		if result.FailureCode == nil {
			successCount++
			if result.CurrentDraftVersion != 2 {
				t.Fatalf("winning CAS must advance to version 2: %#v", result)
			}
			continue
		}
		if *result.FailureCode != string(SavedWorkflowDraftFailureVersionConflict) ||
			result.CurrentDraftVersion != 2 {
			t.Fatalf("losing CAS returned unexpected failure: %#v", result)
		}
		conflictCount++
	}
	if successCount != 1 || conflictCount != concurrentWriters-1 {
		t.Fatalf("atomic CAS drifted: successes=%d conflicts=%d", successCount, conflictCount)
	}

	ownerDenied := readPostgresSavedWorkflowDraft(
		t,
		secondServer,
		payload.DraftID,
		"workspace_demo",
		"app_flow_copilot",
		"subject_other_owner",
		"tenant_demo",
	)
	if ownerDenied.Draft != nil || ownerDenied.FailureCode == nil ||
		*ownerDenied.FailureCode != string(SavedWorkflowDraftFailureScopeDenied) {
		t.Fatalf("owner mismatch must fail closed: %#v", ownerDenied)
	}
	tenantDenied := readPostgresSavedWorkflowDraft(
		t,
		secondServer,
		payload.DraftID,
		"workspace_demo",
		"app_flow_copilot",
		"subject_platform_ops",
		"tenant_other",
	)
	if tenantDenied.Draft != nil || tenantDenied.FailureCode == nil {
		t.Fatalf("tenant mismatch must not return a draft: %#v", tenantDenied)
	}
	applicationDenied := readPostgresSavedWorkflowDraft(
		t,
		secondServer,
		payload.DraftID,
		"workspace_demo",
		"app_other",
		"subject_platform_ops",
		"tenant_demo",
	)
	if applicationDenied.Draft != nil || applicationDenied.FailureCode == nil {
		t.Fatalf("application mismatch must not return a draft: %#v", applicationDenied)
	}

	secondServer.Close()
	unavailable := readPostgresSavedWorkflowDraft(
		t,
		secondServer,
		payload.DraftID,
		"workspace_demo",
		"app_flow_copilot",
		"subject_platform_ops",
		"tenant_demo",
	)
	if unavailable.Draft != nil || unavailable.FailureCode == nil ||
		*unavailable.FailureCode != string(SavedWorkflowDraftFailureStoreUnavailable) {
		t.Fatalf("closed database pool must fail without fallback: %#v", unavailable)
	}

	if _, err := adminPool.Exec(
		databaseContext,
		"UPDATE workflow_saved_draft_schema_versions SET migration_checksum = $1 WHERE component = $2",
		"sha256:incompatible",
		workflowdraftmigrations.Component,
	); err != nil {
		t.Fatalf("corrupt migration marker for negative test: %v", err)
	}
	if _, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-integration"}); err == nil || !strings.Contains(err.Error(), "marker is incompatible") {
		t.Fatalf("server must fail on migration marker mismatch, got: %v", err)
	}
	if _, err := adminPool.Exec(
		databaseContext,
		"UPDATE workflow_saved_draft_schema_versions SET migration_checksum = $1 WHERE component = $2",
		workflowdraftmigrations.ExpectedChecksum(),
		workflowdraftmigrations.Component,
	); err != nil {
		t.Fatalf("restore migration marker before rollback: %v", err)
	}
	rollbackState, err := workflowdraftmigrations.RollbackForDevTest(databaseContext, adminPool)
	if err != nil || rollbackState.MigrationState != workflowdraftmigrations.MigrationStateNotApplied {
		t.Fatalf("reviewed down migration did not return not_applied: state=%#v error=%v", rollbackState, err)
	}
	reappliedState, err := workflowdraftmigrations.Apply(databaseContext, adminPool)
	if err != nil || reappliedState.MigrationState != workflowdraftmigrations.MigrationStateApplied {
		t.Fatalf("migration did not reapply after reviewed rollback: state=%#v error=%v", reappliedState, err)
	}
}

func postgresIntegrationDatabaseURL(t *testing.T) string {
	t.Helper()
	if os.Getenv("RADISHMIND_RUN_POSTGRES_INTEGRATION") != "1" {
		t.Skip("set RADISHMIND_RUN_POSTGRES_INTEGRATION=1 to run PostgreSQL integration tests")
	}
	userName := strings.TrimSpace(os.Getenv("PGUSER"))
	if userName == "" {
		t.Fatal("PGUSER is required for PostgreSQL integration tests")
	}
	return postgresIntegrationDatabaseURLForCredentials(t, userName, os.Getenv("PGPASSWORD"))
}

func postgresIntegrationDatabaseURLForCredentials(
	t *testing.T,
	userName string,
	password string,
) string {
	t.Helper()
	host := strings.TrimSpace(os.Getenv("PGHOST"))
	port := strings.TrimSpace(os.Getenv("PGPORT"))
	databaseName := strings.TrimSpace(os.Getenv("PGDATABASE"))
	if host == "" || port == "" || strings.TrimSpace(userName) == "" || databaseName == "" {
		t.Fatal("PGHOST, PGPORT, database user and PGDATABASE are required for PostgreSQL integration tests")
	}
	userInfo := url.User(userName)
	if password != "" {
		userInfo = url.UserPassword(userName, password)
	}
	databaseURL := &url.URL{
		Scheme: "postgresql",
		User:   userInfo,
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + databaseName,
	}
	query := databaseURL.Query()
	sslMode := strings.TrimSpace(os.Getenv("PGSSLMODE"))
	if sslMode == "" {
		sslMode = "disable"
	}
	query.Set("sslmode", sslMode)
	databaseURL.RawQuery = query.Encode()
	return databaseURL.String()
}

func preparePostgresIntegrationRuntimeRole(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeUser string,
) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	var roleExists bool
	if err := adminPool.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)",
		runtimeUser,
	).Scan(&roleExists); err != nil {
		t.Fatalf("inspect PostgreSQL integration runtime role: %v", err)
	}
	if !roleExists {
		if _, err := adminPool.Exec(
			ctx,
			"CREATE ROLE "+quotedRuntimeUser+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT",
		); err != nil {
			t.Fatalf("create PostgreSQL integration runtime role: %v", err)
		}
	}
	if _, err := adminPool.Exec(
		ctx,
		"ALTER ROLE "+quotedRuntimeUser+" LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT",
	); err != nil {
		t.Fatalf("constrain PostgreSQL integration runtime role: %v", err)
	}
	var databaseName string
	if err := adminPool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("read PostgreSQL integration database name: %v", err)
	}
	statements := []string{
		"REVOKE CREATE ON SCHEMA public FROM PUBLIC",
		"GRANT CONNECT ON DATABASE " + pgx.Identifier{databaseName}.Sanitize() + " TO " + quotedRuntimeUser,
		"GRANT USAGE ON SCHEMA public TO " + quotedRuntimeUser,
		"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO " + quotedRuntimeUser,
	}
	for _, statement := range statements {
		if _, err := adminPool.Exec(ctx, statement); err != nil {
			t.Fatalf("prepare PostgreSQL integration runtime privileges: %v", err)
		}
	}
}

func assertPostgresIntegrationRuntimeRoleCannotMigrate(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeDatabaseURL string,
) {
	t.Helper()
	runtimePool, err := workflowdraftmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL integration runtime pool: %v", err)
	}
	defer runtimePool.Close()
	_, err = runtimePool.Exec(ctx, "CREATE TABLE public.saved_workflow_draft_runtime_must_not_create (id bigint)")
	if err == nil {
		_, _ = adminPool.Exec(ctx, "DROP TABLE IF EXISTS public.saved_workflow_draft_runtime_must_not_create")
		t.Fatal("PostgreSQL integration runtime role unexpectedly created a table")
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("runtime DDL denial returned unexpected database error type")
	}
}

func assertPostgresIntegrationDatabaseIsDisposable(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()
	var databaseName string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("read integration database name: %v", err)
	}
	if !strings.Contains(strings.ToLower(databaseName), "test") {
		t.Fatalf("refusing destructive integration setup for non-test database")
	}
}

func resetPostgresSavedWorkflowDraftSchema(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
) {
	t.Helper()
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS saved_workflow_drafts"); err != nil {
		t.Fatalf("drop integration draft table: %v", err)
	}
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS workflow_saved_draft_schema_versions"); err != nil {
		t.Fatalf("drop integration migration marker: %v", err)
	}
}

func postgresSavedWorkflowDraftIntegrationConfig(databaseURL string) config.Config {
	return config.Config{
		ListenAddr:                        "127.0.0.1:0",
		Provider:                          "mock",
		ControlPlaneReadDevAuthEnabled:    true,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowSavedDraftStoreMode:       string(WorkflowSavedDraftStoreModePostgresDevTest),
		WorkflowSavedDraftDatabaseURL:     databaseURL,
		WorkflowSavedDraftDatabaseTimeout: 5 * time.Second,
	}
}

func newPostgresSavedWorkflowDraftIntegrationServer(t *testing.T, cfg config.Config) *Server {
	t.Helper()
	server, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-integration"})
	if err != nil {
		t.Fatalf("create PostgreSQL saved draft server: %v", err)
	}
	return server
}

func postPostgresSavedWorkflowDraft(
	t *testing.T,
	server *Server,
	payload SavedWorkflowDraftPayload,
	expectedVersion int,
	subjectRef string,
	tenantRef string,
) savedWorkflowDraftEnvelope {
	t.Helper()
	body, err := json.Marshal(savedWorkflowDraftSaveHTTPBody{
		ExpectedDraftVersion: expectedVersion,
		Draft:                savedWorkflowDraftPayloadDocumentFromDraftPayload(payload),
	})
	if err != nil {
		t.Fatalf("marshal PostgreSQL saved draft request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-drafts", bytes.NewReader(body))
	setPostgresSavedWorkflowDraftHeaders(request, payload.WorkspaceID, payload.ApplicationID, subjectRef, tenantRef)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	return decodeSavedWorkflowDraftEnvelope(t, recorder, http.StatusOK)
}

func readPostgresSavedWorkflowDraft(
	t *testing.T,
	server *Server,
	draftID string,
	workspaceID string,
	applicationID string,
	subjectRef string,
	tenantRef string,
) savedWorkflowDraftEnvelope {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts/"+url.PathEscape(draftID)+
			"?workspace_id="+url.QueryEscape(workspaceID)+
			"&application_id="+url.QueryEscape(applicationID),
		nil,
	)
	setPostgresSavedWorkflowDraftHeaders(request, workspaceID, applicationID, subjectRef, tenantRef)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	return decodeSavedWorkflowDraftEnvelope(t, recorder, http.StatusOK)
}

func listPostgresSavedWorkflowDrafts(
	t *testing.T,
	server *Server,
	workspaceID string,
	applicationID string,
	subjectRef string,
	tenantRef string,
) savedWorkflowDraftListEnvelope {
	t.Helper()
	request := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts?workspace_id="+url.QueryEscape(workspaceID)+
			"&application_id="+url.QueryEscape(applicationID),
		nil,
	)
	setPostgresSavedWorkflowDraftHeaders(request, workspaceID, applicationID, subjectRef, tenantRef)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	return decodeSavedWorkflowDraftListEnvelope(t, recorder, http.StatusOK)
}

func setPostgresSavedWorkflowDraftHeaders(
	request *http.Request,
	workspaceID string,
	applicationID string,
	subjectRef string,
	tenantRef string,
) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-Id", "postgres-saved-draft-integration")
	request.Header.Set(controlPlaneReadDevIdentityHeader, "postgres-integration")
	request.Header.Set(controlPlaneReadDevTenantHeader, tenantRef)
	request.Header.Set(controlPlaneReadDevSubjectHeader, subjectRef)
	request.Header.Set(
		controlPlaneReadDevScopesHeader,
		"workflow_drafts:read,workflow_drafts:write",
	)
	request.Header.Set(controlPlaneReadDevAuditHeader, "audit_postgres_saved_draft_integration")
	request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, workspaceID)
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
}
