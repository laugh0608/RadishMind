//go:build postgres_integration

package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
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

func TestSavedWorkflowDraftPostgresLegacy0002Upgrade(t *testing.T) {
	migrationDatabaseURL := postgresIntegrationDatabaseURL(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool, err := workflowdraftmigrations.OpenPool(ctx, migrationDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL legacy upgrade database: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	resetPostgresSavedWorkflowDraftSchema(t, ctx, pool)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresSavedWorkflowDraftSchema(t, cleanupContext, pool)
		pool.Close()
	})

	initialSQL, err := os.ReadFile("../../migrations/workflow_saved_drafts/0001_saved_workflow_drafts.up.sql")
	if err != nil {
		t.Fatalf("read PostgreSQL legacy 0001 migration: %v", err)
	}
	revisionSQL, err := os.ReadFile("../../migrations/workflow_saved_drafts/0002_saved_workflow_draft_revisions.up.sql")
	if err != nil {
		t.Fatalf("read PostgreSQL legacy 0002 migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(initialSQL)+"\n"+string(revisionSQL)); err != nil {
		t.Fatalf("apply PostgreSQL legacy migration chain: %v", err)
	}
	if _, err := pool.Exec(ctx, `
CREATE TABLE workflow_saved_draft_schema_versions (
    component text PRIMARY KEY,
    migration_id text NOT NULL,
    store_schema_version text NOT NULL,
    migration_checksum text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
)`); err != nil {
		t.Fatalf("create PostgreSQL legacy migration marker: %v", err)
	}
	legacyChecksum := fmt.Sprintf(
		"sha256:%x",
		sha256.Sum256([]byte(string(initialSQL)+"\n"+string(revisionSQL))),
	)
	if _, err := pool.Exec(ctx, `
INSERT INTO workflow_saved_draft_schema_versions (
    component, migration_id, store_schema_version, migration_checksum
) VALUES ($1, $2, $3, $4)`,
		workflowdraftmigrations.Component,
		"0002_saved_workflow_draft_revisions",
		workflowdraftmigrations.StoreSchemaVersion,
		legacyChecksum,
	); err != nil {
		t.Fatalf("write PostgreSQL legacy migration marker: %v", err)
	}

	legacyDraft := savedWorkflowDraftLegacyFixture(t)
	payload, err := json.Marshal(savedWorkflowDraftDocumentPointer(&legacyDraft))
	if err != nil {
		t.Fatalf("marshal PostgreSQL legacy draft: %v", err)
	}
	validation, err := json.Marshal(
		savedWorkflowDraftValidationToDocument(legacyDraft.ValidationSummary),
	)
	if err != nil {
		t.Fatalf("marshal PostgreSQL legacy validation: %v", err)
	}
	blocked, err := json.Marshal(
		savedWorkflowDraftBlockedToDocuments(legacyDraft.BlockedCapabilitySummary),
	)
	if err != nil {
		t.Fatalf("marshal PostgreSQL legacy blocked summary: %v", err)
	}
	requestContext := savedWorkflowDraftSQLiteContext()
	if _, err := pool.Exec(ctx, `
INSERT INTO saved_workflow_drafts (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    store_schema_version, schema_version, draft_version, draft_status,
    sanitized_draft_payload, validation_summary, blocked_capability_summary,
    created_at, updated_at, created_by_actor_ref, updated_by_actor_ref,
    request_id, audit_ref
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15, $16, $17, $18
)`,
		requestContext.TenantRef,
		legacyDraft.WorkspaceID,
		legacyDraft.ApplicationID,
		legacyDraft.DraftID,
		requestContext.OwnerSubjectRef,
		workflowdraftmigrations.StoreSchemaVersion,
		legacyDraft.SchemaVersion,
		legacyDraft.DraftVersion,
		legacyDraft.DraftStatus,
		payload,
		validation,
		blocked,
		legacyDraft.CreatedAt,
		legacyDraft.UpdatedAt,
		legacyDraft.CreatedByActorRef,
		legacyDraft.UpdatedByActorRef,
		legacyDraft.RequestAuditMetadata.RequestID,
		legacyDraft.RequestAuditMetadata.AuditRef,
	); err != nil {
		t.Fatalf("insert PostgreSQL legacy draft: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO saved_workflow_draft_revisions (
    tenant_ref, workspace_id, application_id, draft_id, owner_subject_ref,
    draft_version, revision_kind, restored_from_version, sanitized_revision_record
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		requestContext.TenantRef,
		legacyDraft.WorkspaceID,
		legacyDraft.ApplicationID,
		legacyDraft.DraftID,
		requestContext.OwnerSubjectRef,
		legacyDraft.DraftVersion,
		SavedWorkflowDraftRevisionKindSaved,
		0,
		payload,
	); err != nil {
		t.Fatalf("insert PostgreSQL legacy revision: %v", err)
	}
	pending, err := workflowdraftmigrations.Inspect(ctx, pool)
	if err != nil || pending.MigrationState != workflowdraftmigrations.MigrationStateNotApplied ||
		pending.MigrationID != "0002_saved_workflow_draft_revisions" {
		t.Fatalf("PostgreSQL legacy 0002 marker was not recognized: state=%#v err=%v", pending, err)
	}
	upgraded, err := workflowdraftmigrations.Apply(ctx, pool)
	if err != nil || upgraded.MigrationState != workflowdraftmigrations.MigrationStateApplied ||
		upgraded.MigrationID != workflowdraftmigrations.MigrationID {
		t.Fatalf("PostgreSQL legacy 0002 upgrade failed: state=%#v err=%v", upgraded, err)
	}
	var lifecycleState string
	var lifecycleVersion int
	var libraryUpdatedAt time.Time
	var draftName string
	var provenanceKind string
	var eventCount int
	var revisionCount int
	if err := pool.QueryRow(ctx, `
SELECT lifecycle_state, lifecycle_version, library_updated_at, draft_name, provenance_kind,
       (SELECT count(*) FROM saved_workflow_draft_lifecycle_events),
       (SELECT count(*) FROM saved_workflow_draft_revisions WHERE draft_id=$1)
  FROM saved_workflow_drafts
 WHERE draft_id=$1`,
		legacyDraft.DraftID,
	).Scan(
		&lifecycleState,
		&lifecycleVersion,
		&libraryUpdatedAt,
		&draftName,
		&provenanceKind,
		&eventCount,
		&revisionCount,
	); err != nil {
		t.Fatalf("inspect PostgreSQL legacy lifecycle backfill: %v", err)
	}
	expectedUpdatedAt, _ := time.Parse(time.RFC3339Nano, legacyDraft.UpdatedAt)
	if lifecycleState != string(SavedWorkflowDraftLifecycleActive) ||
		lifecycleVersion != 1 ||
		!libraryUpdatedAt.Equal(expectedUpdatedAt) ||
		draftName != legacyDraft.Name ||
		provenanceKind != string(SavedWorkflowDraftProvenanceDefinition) ||
		eventCount != 0 ||
		revisionCount != 1 {
		t.Fatalf(
			"PostgreSQL legacy lifecycle backfill drifted: state=%s version=%d library=%s name=%s provenance=%s events=%d revisions=%d",
			lifecycleState,
			lifecycleVersion,
			libraryUpdatedAt,
			draftName,
			provenanceKind,
			eventCount,
			revisionCount,
		)
	}
}

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
	constrainPostgresSavedWorkflowDraftRevisionPrivileges(
		t,
		databaseContext,
		adminPool,
		runtimeUser,
	)
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
	assertPostgresSavedWorkflowDraftRevisionIsAppendOnly(
		t,
		databaseContext,
		runtimeDatabaseURL,
		payload.DraftID,
	)
	assertPostgresSavedWorkflowDraftLifecycleEventsAreAppendOnly(
		t,
		databaseContext,
		runtimeDatabaseURL,
		payload.DraftID,
	)
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
	revisionContext := SavedWorkflowDraftContext{
		RequestContext:  databaseContext,
		RequestID:       "request-postgres-revision-history",
		TenantRef:       "tenant_demo",
		WorkspaceID:     "workspace_demo",
		ApplicationID:   "app_flow_copilot",
		ActorRef:        "subject_platform_ops",
		OwnerSubjectRef: "subject_platform_ops",
		ScopeGrants:     []string{"workflow_drafts:read", "workflow_drafts:write"},
		AuditRef:        "audit-postgres-revision-history",
		WriteEnabled:    true,
	}
	revisionService := secondServer.savedWorkflowDraftService()
	revisionHistory := revisionService.ListDraftRevisions(
		revisionContext,
		ListSavedWorkflowDraftRevisionsRequest{DraftID: payload.DraftID},
	)
	if revisionHistory.FailureCode != "" || len(revisionHistory.Revisions) != 2 ||
		revisionHistory.Revisions[0].DraftVersion != 2 ||
		revisionHistory.Revisions[1].DraftVersion != 1 {
		t.Fatalf("PostgreSQL revision history drifted: %#v", revisionHistory)
	}
	restoredRevision := revisionService.RestoreDraftRevision(
		revisionContext,
		RestoreSavedWorkflowDraftRevisionRequest{
			DraftID:                     payload.DraftID,
			SourceDraftVersion:          1,
			ExpectedCurrentDraftVersion: 2,
		},
	)
	if restoredRevision.FailureCode != "" || restoredRevision.Draft == nil ||
		restoredRevision.Draft.DraftVersion != 3 ||
		restoredRevision.Draft.Name != payload.Name {
		t.Fatalf("PostgreSQL revision restore drifted: %#v", restoredRevision)
	}
	repositoryStore, ok := secondServer.savedWorkflowDraftStore.(*repositorySavedWorkflowDraftStore)
	if !ok {
		t.Fatalf("PostgreSQL saved draft store does not expose repository contract: %T",
			secondServer.savedWorkflowDraftStore)
	}
	libraryService := newSavedWorkflowDraftService(
		newRepositorySavedWorkflowDraftLibraryStore(repositoryStore),
	)
	assertPostgresSavedWorkflowDraftLifecycleAtomicity(
		t,
		databaseContext,
		adminPool,
		runtimeUser,
		libraryService,
		revisionContext,
		restoredRevision.Draft,
	)
	assertPostgresSavedWorkflowDraftLibraryMatchesMemory(
		t,
		libraryService,
		revisionContext,
	)

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

func constrainPostgresSavedWorkflowDraftRevisionPrivileges(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeUser string,
) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	statements := []string{
		"REVOKE ALL ON TABLE saved_workflow_draft_revisions FROM " + quotedRuntimeUser,
		"GRANT SELECT, INSERT ON TABLE saved_workflow_draft_revisions TO " + quotedRuntimeUser,
		"REVOKE ALL ON TABLE saved_workflow_draft_lifecycle_events FROM " + quotedRuntimeUser,
		"GRANT SELECT, INSERT ON TABLE saved_workflow_draft_lifecycle_events TO " + quotedRuntimeUser,
	}
	for _, statement := range statements {
		if _, err := adminPool.Exec(ctx, statement); err != nil {
			t.Fatalf("constrain PostgreSQL saved draft revision privileges: %v", err)
		}
	}
}

func assertPostgresSavedWorkflowDraftLifecycleEventsAreAppendOnly(
	t *testing.T,
	ctx context.Context,
	runtimeDatabaseURL string,
	draftID string,
) {
	t.Helper()
	runtimePool, err := workflowdraftmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL runtime pool for lifecycle append-only check: %v", err)
	}
	defer runtimePool.Close()
	for _, statement := range []string{
		"UPDATE saved_workflow_draft_lifecycle_events SET actor_ref='forbidden' WHERE draft_id=$1",
		"DELETE FROM saved_workflow_draft_lifecycle_events WHERE draft_id=$1",
	} {
		if _, err := runtimePool.Exec(ctx, statement, draftID); err == nil {
			t.Fatalf("PostgreSQL runtime role unexpectedly mutated lifecycle event history")
		} else {
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
				t.Fatalf("lifecycle append-only denial returned unexpected database error")
			}
		}
	}
}

func assertPostgresSavedWorkflowDraftLifecycleAtomicity(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeUser string,
	service savedWorkflowDraftService,
	requestContext SavedWorkflowDraftContext,
	current *SavedWorkflowDraft,
) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	if _, err := adminPool.Exec(
		ctx,
		"REVOKE INSERT ON TABLE saved_workflow_draft_lifecycle_events FROM "+quotedRuntimeUser,
	); err != nil {
		t.Fatalf("revoke PostgreSQL lifecycle event insert: %v", err)
	}
	failed := service.ArchiveDraft(
		requestContext,
		TransitionSavedWorkflowDraftLifecycleRequest{
			DraftID:                  current.DraftID,
			ExpectedDraftVersion:     current.DraftVersion,
			ExpectedLifecycleVersion: current.LifecycleVersion,
		},
	)
	if failed.FailureCode != SavedWorkflowDraftFailureLifecycleEventWrite ||
		failed.Draft != nil || failed.Event != nil {
		t.Fatalf("PostgreSQL lifecycle event failure did not fail closed: %#v", failed)
	}
	read := service.ReadDraft(
		requestContext,
		ReadWorkflowDraftRequest{DraftID: current.DraftID},
	)
	if read.FailureCode != "" || read.Draft == nil ||
		read.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		read.Draft.LifecycleVersion != current.LifecycleVersion {
		t.Fatalf("PostgreSQL lifecycle event failure committed current state: %#v", read)
	}
	if _, err := adminPool.Exec(
		ctx,
		"GRANT INSERT ON TABLE saved_workflow_draft_lifecycle_events TO "+quotedRuntimeUser,
	); err != nil {
		t.Fatalf("restore PostgreSQL lifecycle event insert: %v", err)
	}

	const writers = 8
	results := make(chan SavedWorkflowDraftLifecycleTransitionResult, writers)
	var waitGroup sync.WaitGroup
	for index := 0; index < writers; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			results <- service.ArchiveDraft(
				requestContext,
				TransitionSavedWorkflowDraftLifecycleRequest{
					DraftID:                  current.DraftID,
					ExpectedDraftVersion:     current.DraftVersion,
					ExpectedLifecycleVersion: current.LifecycleVersion,
				},
			)
		}()
	}
	waitGroup.Wait()
	close(results)
	successes := 0
	conflicts := 0
	var archivedDraft *SavedWorkflowDraft
	for result := range results {
		switch result.FailureCode {
		case "":
			if result.Draft == nil {
				t.Fatal("PostgreSQL lifecycle CAS winner omitted draft")
			}
			successes++
			archivedDraft = cloneSavedWorkflowDraftPointer(*result.Draft)
		case SavedWorkflowDraftFailureLifecycleVersionConflict,
			SavedWorkflowDraftFailureLifecycleStateConflict:
			conflicts++
		default:
			t.Fatalf("PostgreSQL lifecycle CAS returned unexpected result: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("PostgreSQL lifecycle CAS drifted: successes=%d conflicts=%d", successes, conflicts)
	}
	archivedPage := service.ListDrafts(
		requestContext,
		ListWorkflowDraftsRequest{
			LifecycleState: SavedWorkflowDraftLifecycleArchived,
			Limit:          1,
		},
	)
	if archivedPage.FailureCode != "" || len(archivedPage.Summaries) != 1 ||
		archivedPage.Summaries[0].DraftID != current.DraftID {
		t.Fatalf("PostgreSQL archived library page drifted: %#v", archivedPage)
	}
	if _, err := adminPool.Exec(
		ctx,
		"UPDATE saved_workflow_drafts SET draft_name='projection drift' WHERE draft_id=$1",
		current.DraftID,
	); err != nil {
		t.Fatalf("corrupt PostgreSQL lifecycle projection: %v", err)
	}
	corrupted := service.ListDrafts(
		requestContext,
		ListWorkflowDraftsRequest{LifecycleState: SavedWorkflowDraftLifecycleArchived},
	)
	if corrupted.FailureCode != SavedWorkflowDraftFailureLifecycleStoreContract ||
		len(corrupted.Summaries) != 0 {
		t.Fatalf("PostgreSQL projection mismatch did not fail closed: %#v", corrupted)
	}
	if _, err := adminPool.Exec(
		ctx,
		"UPDATE saved_workflow_drafts SET draft_name=$1 WHERE draft_id=$2",
		current.Name,
		current.DraftID,
	); err != nil {
		t.Fatalf("restore PostgreSQL lifecycle projection: %v", err)
	}
	unarchived := service.UnarchiveDraft(
		requestContext,
		TransitionSavedWorkflowDraftLifecycleRequest{
			DraftID:                  archivedDraft.DraftID,
			ExpectedDraftVersion:     archivedDraft.DraftVersion,
			ExpectedLifecycleVersion: archivedDraft.LifecycleVersion,
		},
	)
	if unarchived.FailureCode != "" || unarchived.Draft == nil ||
		unarchived.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		unarchived.Draft.LifecycleVersion != archivedDraft.LifecycleVersion+1 {
		t.Fatalf("PostgreSQL lifecycle unarchive drifted: %#v", unarchived)
	}
	var eventCount int
	if err := adminPool.QueryRow(
		ctx,
		"SELECT count(*) FROM saved_workflow_draft_lifecycle_events WHERE draft_id=$1",
		current.DraftID,
	).Scan(&eventCount); err != nil || eventCount != 2 {
		t.Fatalf("PostgreSQL lifecycle event count drifted: count=%d err=%v", eventCount, err)
	}
}

func assertPostgresSavedWorkflowDraftLibraryMatchesMemory(
	t *testing.T,
	postgresService savedWorkflowDraftService,
	requestContext SavedWorkflowDraftContext,
) {
	t.Helper()
	requestContext.RequestID = "request-postgres-library-matrix"
	requestContext.ActorRef = "subject_postgres_library_matrix"
	requestContext.OwnerSubjectRef = requestContext.ActorRef
	requestContext.AuditRef = "audit-postgres-library-matrix"
	memoryService := newSavedWorkflowDraftService(newMemorySavedWorkflowDraftStore())
	populateSavedWorkflowDraftLibraryFixture(t, &memoryService, requestContext)
	populateSavedWorkflowDraftLibraryFixture(t, &postgresService, requestContext)
	for _, request := range []ListWorkflowDraftsRequest{
		{LifecycleState: SavedWorkflowDraftLifecycleActive, Limit: 4},
		{LifecycleState: SavedWorkflowDraftLifecycleArchived, Limit: 4},
		{
			LifecycleState: SavedWorkflowDraftLifecycleActive,
			Limit:          4,
			NamePrefix:     "Alpha",
		},
		{
			LifecycleState:  SavedWorkflowDraftLifecycleActive,
			Limit:           4,
			ValidationState: SavedWorkflowDraftStatusBlockedCapability,
		},
		{
			LifecycleState: SavedWorkflowDraftLifecycleActive,
			Limit:          4,
			ProvenanceKind: SavedWorkflowDraftProvenanceDraftDerived,
		},
	} {
		memoryIDs := collectSavedWorkflowDraftLibraryIDs(
			t,
			memoryService,
			requestContext,
			request,
		)
		postgresIDs := collectSavedWorkflowDraftLibraryIDs(
			t,
			postgresService,
			requestContext,
			request,
		)
		if !reflect.DeepEqual(postgresIDs, memoryIDs) {
			t.Fatalf("PostgreSQL library matrix drifted for %#v: postgres=%v memory=%v",
				request, postgresIDs, memoryIDs)
		}
	}
}

func assertPostgresSavedWorkflowDraftRevisionIsAppendOnly(
	t *testing.T,
	ctx context.Context,
	runtimeDatabaseURL string,
	draftID string,
) {
	t.Helper()
	runtimePool, err := workflowdraftmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL runtime pool for append-only check: %v", err)
	}
	defer runtimePool.Close()
	for _, statement := range []string{
		"UPDATE saved_workflow_draft_revisions SET revision_kind='saved' WHERE draft_id=$1",
		"DELETE FROM saved_workflow_draft_revisions WHERE draft_id=$1",
	} {
		if _, err := runtimePool.Exec(ctx, statement, draftID); err == nil {
			t.Fatalf("PostgreSQL runtime role unexpectedly mutated saved draft revision history")
		} else {
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
				t.Fatalf("append-only denial returned unexpected database error")
			}
		}
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
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS saved_workflow_draft_lifecycle_events"); err != nil {
		t.Fatalf("drop integration draft lifecycle event table: %v", err)
	}
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS saved_workflow_draft_revisions"); err != nil {
		t.Fatalf("drop integration draft revision table: %v", err)
	}
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS saved_workflow_drafts"); err != nil {
		t.Fatalf("drop integration draft table: %v", err)
	}
	if _, err := pool.Exec(
		ctx,
		"DROP FUNCTION IF EXISTS reject_saved_workflow_draft_lifecycle_event_mutation()",
	); err != nil {
		t.Fatalf("drop integration draft lifecycle event guard: %v", err)
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
	request.Header.Set(activeWorkspaceHeader, workspaceID)
	request.Header.Set(controlPlaneReadDevMembershipHeader, workspaceID)
	request.Header.Set(controlPlaneReadDevMembershipPermHeader, "workflow_drafts:write")
	request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, workspaceID)
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
}
