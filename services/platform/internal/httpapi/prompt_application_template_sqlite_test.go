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
	sqlitepromptapplicationtemplatemigrations "radishmind.local/services/platform/migrations/sqlite/prompt_application_templates"
)

func TestPromptApplicationTemplateSQLiteLifecycleRestartAndCorruption(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	first := openPromptApplicationTemplateSQLiteRuntime(t, databasePath)
	service := newPromptApplicationTemplateService(newSQLitePromptApplicationTemplateRepository(first.DB()))
	service.now = func() time.Time { return time.Date(2026, 7, 21, 9, 0, 0, 0, time.UTC) }
	requestContext := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	created := service.SaveDraft(requestContext, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("create SQLite prompt application template draft: %#v", created)
	}
	version := service.CreateVersion(requestContext, input.TemplateID, 1)
	if version.FailureCode != "" || version.Version == nil || version.Version.TemplateVersion != 1 {
		t.Fatalf("create SQLite immutable prompt application template version: %#v", version)
	}
	if _, err := first.DB().ExecContext(context.Background(), `UPDATE prompt_application_template_versions SET template_digest=template_digest
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=? AND template_version=1`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, input.TemplateID); err == nil {
		t.Fatal("SQLite immutable version accepted an update")
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first SQLite prompt application template runtime: %v", err)
	}
	if read := service.ReadDraft(requestContext, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureStoreUnavailable || read.Draft != nil {
		t.Fatalf("closed SQLite owner fell back to memory state: %#v", read)
	}

	restarted := openPromptApplicationTemplateSQLiteRuntime(t, databasePath)
	restartedService := newPromptApplicationTemplateService(newSQLitePromptApplicationTemplateRepository(restarted.DB()))
	restored := restartedService.ReadVersion(requestContext, input.TemplateID, 1)
	if restored.FailureCode != "" || restored.Version == nil || restored.Version.TemplateDigest != created.Draft.TemplateDigest {
		t.Fatalf("restore immutable version after SQLite restart: %#v", restored)
	}
	otherWorkspace := requestContext
	otherWorkspace.WorkspaceID = "workspace_other"
	if read := restartedService.ReadDraft(otherWorkspace, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureNotFound {
		t.Fatalf("SQLite repository leaked a draft across workspace scope: %#v", read)
	}
	if _, err := restarted.DB().ExecContext(context.Background(), `UPDATE prompt_application_template_drafts SET
		draft_version=draft_version+1,
		updated_at_unix_nano=updated_at_unix_nano+1,
		sanitized_draft_payload=json_set(sanitized_draft_payload,'$.draft_version',draft_version+1,'$.messages[0].content','corrupted without digest update')
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND template_id=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, input.TemplateID); err != nil {
		t.Fatalf("inject structurally valid SQLite digest drift: %v", err)
	}
	if read := restartedService.ReadDraft(requestContext, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureStoreContract {
		t.Fatalf("SQLite read accepted stored digest drift: %#v", read)
	}
}

func TestPromptApplicationTemplateSQLiteConcurrentCASAndSensitiveMaterial(t *testing.T) {
	runtime := openPromptApplicationTemplateSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	service := newPromptApplicationTemplateService(newSQLitePromptApplicationTemplateRepository(runtime.DB()))
	requestContext := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	if seed := service.SaveDraft(requestContext, input, 0); seed.FailureCode != "" {
		t.Fatalf("seed SQLite prompt application template draft: %#v", seed)
	}

	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			candidate := input
			candidate.Description = "concurrent SQLite update"
			result := service.SaveDraft(requestContext, candidate, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case PromptApplicationTemplateFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected SQLite concurrent failure for writer %d: %#v", index, result)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("SQLite CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}

	forbidden := validPromptApplicationTemplateDraftInput()
	forbidden.TemplateID = "ptpl_bbbbbbbbbbbbbbbb"
	forbidden.Variables[1].DefaultValue = "Authorization: Bearer sqlite-secret-sentinel"
	if result := service.SaveDraft(requestContext, forbidden, 0); result.FailureCode != PromptApplicationTemplateFailureSecretForbidden {
		t.Fatalf("SQLite owner accepted credential-like template material: %#v", result)
	}
	var storedPayloads string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT COALESCE(group_concat(sanitized_draft_payload,''),'') FROM prompt_application_template_drafts`).Scan(&storedPayloads); err != nil {
		t.Fatalf("scan SQLite prompt application template payloads: %v", err)
	}
	for _, forbiddenMaterial := range []string{"sqlite-secret-sentinel", "rendered-output-sentinel", "runtime-variable-sentinel"} {
		if strings.Contains(storedPayloads, forbiddenMaterial) {
			t.Fatalf("SQLite prompt application template owner persisted forbidden material %q", forbiddenMaterial)
		}
	}
}

func TestPromptApplicationTemplateSQLiteFactoryRequiresSharedMigratedRuntime(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, PromptTemplateDevHTTPEnabled: true, PromptTemplateDevWriteEnabled: true,
		PromptTemplateStoreMode: "sqlite_dev", PromptTemplateDatabaseTimeout: time.Second,
	}
	if _, _, err := newPromptApplicationTemplateRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev prompt application template requires the shared SQLite runtime" {
		t.Fatalf("prompt application template factory accepted a missing shared SQLite runtime: %v", err)
	}
	withoutMigration, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: filepath.Join(t.TempDir(), "without-migration.db")})
	if err != nil {
		t.Fatalf("open SQLite runtime without template migration: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigration.Close() })
	if _, _, err := newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigration); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("prompt application template factory accepted a missing migration: %v", err)
	}
	runtime := openPromptApplicationTemplateSQLiteRuntime(t, filepath.Join(t.TempDir(), "migrated.db"))
	repository, closeRepository, err := newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || closeRepository == nil {
		t.Fatalf("construct prompt application template SQLite repository: repository=%T err=%v", repository, err)
	}
	sqliteRepository, ok := repository.(*sqlitePromptApplicationTemplateRepository)
	if !ok || sqliteRepository.database != runtime.DB() {
		t.Fatalf("prompt application template repository did not share the aggregate SQLite runtime: %T", repository)
	}
	closeRepository()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component closer closed the aggregate SQLite runtime: %v", err)
	}
	migration := sqlitepromptapplicationtemplatemigrations.Migrations()[0]
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations SET migration_checksum=? WHERE component=? AND migration_id=?`,
		"sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", migration.Component, migration.ID); err != nil {
		t.Fatalf("inject SQLite template migration checksum mismatch: %v", err)
	}
	if _, _, err := newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(cfg, runtime); err == nil || err.Error() != "SQLite development component migration marker mismatch" {
		t.Fatalf("prompt application template factory accepted a mismatched marker: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations SET migration_checksum=? WHERE component=? AND migration_id=?`,
		migration.Checksum(), migration.Component, migration.ID); err != nil {
		t.Fatalf("restore SQLite template migration checksum: %v", err)
	}
}

func TestPromptApplicationTemplateFactoryDoesNotFallbackFromPostgresConfig(t *testing.T) {
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, PromptTemplateDevHTTPEnabled: true, PromptTemplateDevWriteEnabled: true,
		PromptTemplateStoreMode: "postgres_dev_test", PromptTemplateDatabaseTimeout: time.Second,
	}
	repository, closeRepository, err := newPromptApplicationTemplateRepositoryFromConfig(cfg)
	if err == nil || err.Error() != "postgres_dev_test prompt application template config is incomplete" || repository != nil || closeRepository != nil {
		t.Fatalf("incomplete PostgreSQL prompt template config fell back to another owner: repository=%T closer_present=%t err=%v", repository, closeRepository != nil, err)
	}
}

func openPromptApplicationTemplateSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitepromptapplicationtemplatemigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open prompt application template SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}
