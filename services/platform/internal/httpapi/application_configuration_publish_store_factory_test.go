package httpapi

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	applicationdraftmigrations "radishmind.local/services/platform/migrations/sqlite/application_configuration_drafts"
	applicationpublishmigrations "radishmind.local/services/platform/migrations/sqlite/application_publish_candidates"
)

func TestApplicationConfigurationAndPublishFactoriesRequireSharedSQLiteRuntime(t *testing.T) {
	cfg := applicationConfigurationPublishSQLiteFactoryConfig()
	if _, _, err := newApplicationConfigurationDraftRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev application draft requires the shared SQLite runtime" {
		t.Fatalf("application draft must reject a missing shared SQLite runtime, got %v", err)
	}
	if _, _, err := newApplicationPublishCandidateRepositoryFromConfig(cfg); err == nil || err.Error() != "sqlite_dev application publish candidate requires the shared SQLite runtime" {
		t.Fatalf("publish candidate must reject a missing shared SQLite runtime, got %v", err)
	}

	withoutMigrations, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "without-migrations.db"),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime without component migrations: %v", err)
	}
	t.Cleanup(func() { _ = withoutMigrations.Close() })
	if _, _, err := newApplicationConfigurationDraftRepositoryFromConfigWithSQLiteRuntime(cfg, withoutMigrations); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("application draft must reject a missing migration, got %v", err)
	}

	draftOnlyRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "draft-only.db"),
		Migrations:   applicationdraftmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime with only application draft migration: %v", err)
	}
	t.Cleanup(func() { _ = draftOnlyRuntime.Close() })
	if repository, closeRepository, err := newApplicationConfigurationDraftRepositoryFromConfigWithSQLiteRuntime(cfg, draftOnlyRuntime); err != nil || repository == nil || closeRepository == nil {
		t.Fatalf("construct application draft repository from migrated runtime: repository=%T err=%v", repository, err)
	}
	if _, _, err := newApplicationPublishCandidateRepositoryFromConfigWithSQLiteRuntime(cfg, draftOnlyRuntime); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("publish candidate must reject a missing component migration, got %v", err)
	}

	publishOnlyRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "publish-only.db"),
		Migrations:   applicationpublishmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite runtime with only publish candidate migration: %v", err)
	}
	t.Cleanup(func() { _ = publishOnlyRuntime.Close() })
	if _, _, err := newApplicationPublishCandidateRepositoryFromConfigWithSQLiteRuntime(cfg, publishOnlyRuntime); err == nil || err.Error() != "SQLite development component migration is not applied" {
		t.Fatalf("publish candidate must reject a missing application draft migration, got %v", err)
	}
}

func TestApplicationConfigurationAndPublishFactoriesShareRuntimeOwnership(t *testing.T) {
	runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	cfg := applicationConfigurationPublishSQLiteFactoryConfig()
	draftRepository, closeDraftRepository, err := newApplicationConfigurationDraftRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || closeDraftRepository == nil {
		t.Fatalf("construct application draft SQLite repository: repository=%T err=%v", draftRepository, err)
	}
	publishRepository, closePublishRepository, err := newApplicationPublishCandidateRepositoryFromConfigWithSQLiteRuntime(cfg, runtime)
	if err != nil || closePublishRepository == nil {
		t.Fatalf("construct publish candidate SQLite repository: repository=%T err=%v", publishRepository, err)
	}
	draftSQLiteRepository, ok := draftRepository.(*sqliteApplicationConfigurationDraftRepository)
	if !ok || draftSQLiteRepository.database != runtime.DB() {
		t.Fatalf("application draft repository did not receive the shared runtime: %T", draftRepository)
	}
	publishSQLiteRepository, ok := publishRepository.(*sqliteApplicationPublishCandidateRepository)
	if !ok || publishSQLiteRepository.database != runtime.DB() {
		t.Fatalf("publish candidate repository did not receive the shared runtime: %T", publishRepository)
	}
	closePublishRepository()
	closeDraftRepository()
	if err := runtime.DB().PingContext(context.Background()); err != nil {
		t.Fatalf("component close callbacks must not close the shared SQLite runtime: %v", err)
	}
}

func TestApplicationPublishSQLiteFactoryRequiresSQLiteDraftStore(t *testing.T) {
	cfg := applicationConfigurationPublishSQLiteFactoryConfig()
	cfg.ApplicationDraftStoreMode = "memory_dev"
	if _, _, err := newApplicationPublishCandidateRepositoryFromConfigWithSQLiteRuntime(cfg, nil); err == nil || err.Error() != "sqlite_dev application publish candidate config is incomplete" {
		t.Fatalf("publish candidate must reject a different application draft store mode, got %v", err)
	}
}

func applicationConfigurationPublishSQLiteFactoryConfig() config.Config {
	return config.Config{
		ControlPlaneReadDevAuthEnabled:    true,
		ApplicationDraftDevHTTPEnabled:    true,
		ApplicationDraftDevWriteEnabled:   true,
		ApplicationDraftStoreMode:         "sqlite_dev",
		ApplicationPublishDevHTTPEnabled:  true,
		ApplicationPublishDevWriteEnabled: true,
		ApplicationPublishStoreMode:       "sqlite_dev",
	}
}
