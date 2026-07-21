package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	promptapplicationtemplatemigrations "radishmind.local/services/platform/migrations/prompt_application_templates"
	sqlitepromptapplicationtemplatemigrations "radishmind.local/services/platform/migrations/sqlite/prompt_application_templates"
)

func newPromptApplicationTemplateRepositoryFromConfig(cfg config.Config) (promptApplicationTemplateRepository, func(), error) {
	return newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(cfg, nil)
}

func newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(cfg config.Config, sqliteRuntime *sqlitedev.Runtime) (promptApplicationTemplateRepository, func(), error) {
	mode := strings.TrimSpace(cfg.PromptTemplateStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryPromptApplicationTemplateRepository(), func() {}, nil
	}
	if mode == "sqlite_dev" {
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.PromptTemplateDevHTTPEnabled || !cfg.PromptTemplateDevWriteEnabled {
			return nil, nil, errors.New("sqlite_dev prompt application template config is incomplete")
		}
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev prompt application template requires the shared SQLite runtime")
		}
		timeout := cfg.PromptTemplateDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sqliteRuntime.VerifyMigrations(ctx, sqlitepromptapplicationtemplatemigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLitePromptApplicationTemplateRepository(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported prompt application template store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.PromptTemplateDevHTTPEnabled || !cfg.PromptTemplateDevWriteEnabled || strings.TrimSpace(cfg.PromptTemplateDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test prompt application template config is incomplete")
	}
	timeout := cfg.PromptTemplateDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := promptapplicationtemplatemigrations.OpenPool(ctx, cfg.PromptTemplateDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := promptapplicationtemplatemigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != promptapplicationtemplatemigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("prompt application template PostgreSQL migration is not applied")
	}
	return newPostgresPromptApplicationTemplateRepository(pool), closePool, nil
}
