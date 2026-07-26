package httpapi

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteagentcopilotprofilemigrations "radishmind.local/services/platform/migrations/sqlite/agent_copilot_profiles"
)

func TestAgentCopilotProfileSQLiteLifecycleRestartAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	first := openAgentCopilotProfileSQLiteRuntime(t, databasePath)
	service := newAgentCopilotProfileService(newSQLiteAgentCopilotProfileRepository(first.DB()))
	service.now = func() time.Time { return time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC) }
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	created := service.SaveDraft(ctx, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("create SQLite profile draft: %#v", created)
	}
	version := service.CreateVersion(ctx, input.ProfileID, 1)
	if version.FailureCode != "" || version.Version == nil || version.Version.ProfileVersion != 1 {
		t.Fatalf("create SQLite immutable profile version: %#v", version)
	}
	if _, err := first.DB().ExecContext(context.Background(), `UPDATE agent_copilot_profile_versions SET profile_digest=profile_digest
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=? AND profile_version=1`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, input.ProfileID); err == nil {
		t.Fatal("SQLite immutable profile version accepted an update")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite profile runtime: %v", err)
	}
	if read := service.ReadDraft(ctx, input.ProfileID); read.FailureCode != AgentCopilotProfileFailureStoreUnavailable || read.Draft != nil {
		t.Fatalf("closed SQLite profile owner fell back to memory: %#v", read)
	}

	restarted := openAgentCopilotProfileSQLiteRuntime(t, databasePath)
	restartedService := newAgentCopilotProfileService(newSQLiteAgentCopilotProfileRepository(restarted.DB()))
	restored := restartedService.ReadVersion(ctx, input.ProfileID, 1)
	if restored.FailureCode != "" || restored.Version == nil || restored.Version.ProfileDigest != created.Draft.ProfileDigest {
		t.Fatalf("restore profile version after SQLite restart: %#v", restored)
	}
	otherWorkspace := ctx
	otherWorkspace.WorkspaceID = "workspace_other"
	if read := restartedService.ReadDraft(otherWorkspace, input.ProfileID); read.FailureCode != AgentCopilotProfileFailureNotFound {
		t.Fatalf("SQLite repository leaked profile across workspace scope: %#v", read)
	}
	if _, err := restarted.DB().ExecContext(context.Background(), `UPDATE agent_copilot_profile_drafts SET
		draft_version=draft_version+1,
		updated_at_unix_nano=updated_at_unix_nano+1,
		sanitized_draft_payload=json_set(sanitized_draft_payload,'$.draft_version',draft_version+1,'$.description','corrupted without digest update')
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND profile_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, input.ProfileID); err != nil {
		t.Fatalf("inject structurally valid SQLite profile digest drift: %v", err)
	}
	if read := restartedService.ReadDraft(ctx, input.ProfileID); read.FailureCode != AgentCopilotProfileFailureDigestDrift {
		t.Fatalf("SQLite read accepted profile digest drift: %#v", read)
	}
}

func TestAgentCopilotProfileSQLiteConcurrentCASAndSensitiveMaterial(t *testing.T) {
	runtime := openAgentCopilotProfileSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	service := newAgentCopilotProfileService(newSQLiteAgentCopilotProfileRepository(runtime.DB()))
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	if seed := service.SaveDraft(ctx, input, 0); seed.FailureCode != "" {
		t.Fatalf("seed SQLite profile draft: %#v", seed)
	}

	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			candidate := input
			candidate.Description = "concurrent SQLite profile update"
			result := service.SaveDraft(ctx, candidate, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case AgentCopilotProfileFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected SQLite profile CAS failure for writer %d: %#v", index, result)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("SQLite profile CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}

	forbidden := agentCopilotProfileTestInput("acpf_bbbbbbbbbbbbbbbb")
	forbidden.Description = "Authorization: Bearer sqlite-profile-secret-sentinel"
	if result := service.SaveDraft(ctx, forbidden, 0); result.FailureCode != AgentCopilotProfileFailureSecretMaterialForbidden {
		t.Fatalf("SQLite profile owner accepted credential-like material: %#v", result)
	}
	var storedPayloads string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT COALESCE(group_concat(sanitized_draft_payload,''),'') FROM agent_copilot_profile_drafts`).Scan(&storedPayloads); err != nil {
		t.Fatalf("scan SQLite profile payloads: %v", err)
	}
	if strings.Contains(storedPayloads, "sqlite-profile-secret-sentinel") {
		t.Fatal("SQLite profile owner persisted forbidden sensitive material")
	}
}

func TestAgentCopilotProfileSQLiteFactoryRequiresSharedMigratedRuntime(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, AgentCopilotProfileDevHTTPEnabled: true, AgentCopilotProfileDevWriteEnabled: true,
		AgentCopilotProfileStoreMode: "sqlite_dev", AgentCopilotProfileDatabaseTimeout: time.Second,
	}
	if _, _, err := newAgentCopilotProfileRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev agent copilot profile requires the shared SQLite runtime" {
		t.Fatalf("profile factory accepted a missing shared SQLite runtime: %v", err)
	}
	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "without-migration.db")})
	if err != nil {
		t.Fatalf("open SQLite runtime without profile migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if _, _, err := newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigration); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("profile factory accepted a missing migration: %v", err)
	}
	runtime := openAgentCopilotProfileSQLiteRuntime(t, filepath.Join(t.TempDir(), "migrated.db"))
	repository, closeRepository, err := newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || closeRepository == nil {
		t.Fatalf("construct profile SQLite repository: repository=%T err=%v", repository, err)
	}
	sqliteRepository, ok := repository.(*sqliteAgentCopilotProfileRepository)
	if !ok || sqliteRepository.database != runtime.DB() {
		t.Fatalf("profile repository did not share the aggregate SQLite runtime: %T", repository)
	}
	closeRepository()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component closer closed aggregate SQLite runtime: %v", err)
	}
	migration := sqliteagentcopilotprofilemigrations.Migrations()[0]
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations SET migration_checksum=? WHERE component=? AND migration_id=?`,
		"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", migration.Component, migration.ID); err != nil {
		t.Fatalf("inject SQLite profile migration checksum mismatch: %v", err)
	}
	if _, _, err := newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(cfg, runtime); err == nil || err.Error() != "SQLite development component migration marker mismatch" {
		t.Fatalf("profile factory accepted a mismatched marker: %v", err)
	}
}

func TestAgentCopilotProfileFactoryDoesNotFallbackFromPostgresConfig(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, AgentCopilotProfileDevHTTPEnabled: true, AgentCopilotProfileDevWriteEnabled: true,
		AgentCopilotProfileStoreMode: "postgres_dev_test", AgentCopilotProfileDatabaseTimeout: time.Second,
	}
	repository, closeRepository, err := newAgentCopilotProfileRepositoryFromConfig(cfg)
	if err == nil || err.Error() != "postgres_dev_test agent copilot profile config is incomplete" || repository != nil || closeRepository != nil {
		t.Fatalf("incomplete PostgreSQL profile config fell back to another owner: repository=%T closer_present=%t err=%v", repository, closeRepository != nil, err)
	}
}

func openAgentCopilotProfileSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteagentcopilotprofilemigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open agent copilot profile SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
