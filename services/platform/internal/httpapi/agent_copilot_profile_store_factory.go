package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	agentcopilotprofilemigrations "radishmind.local/services/platform/migrations/agent_copilot_profiles"
	sqliteagentcopilotprofilemigrations "radishmind.local/services/platform/migrations/sqlite/agent_copilot_profiles"
)

func newAgentCopilotProfileRepositoryFromConfig(cfg config.Config) (agentCopilotProfileRepository, func(), error) {
	return newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(cfg, nil)
}

func newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(cfg config.Config, sqliteRuntime *sqlitedev.Runtime) (agentCopilotProfileRepository, func(), error) {
	mode := strings.TrimSpace(cfg.AgentCopilotProfileStoreMode)
	if mode == "" || mode == "memory_dev" {
		return newMemoryAgentCopilotProfileRepository(), func() {}, nil
	}
	if mode == "sqlite_dev" {
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.AgentCopilotProfileDevHTTPEnabled || !cfg.AgentCopilotProfileDevWriteEnabled {
			return nil, nil, errors.New("sqlite_dev agent copilot profile config is incomplete")
		}
		if sqliteRuntime == nil || sqliteRuntime.DB() == nil {
			return nil, nil, errors.New("sqlite_dev agent copilot profile requires the shared SQLite runtime")
		}
		timeout := cfg.AgentCopilotProfileDatabaseTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := sqliteRuntime.VerifyMigrations(ctx, sqliteagentcopilotprofilemigrations.Migrations()); err != nil {
			return nil, nil, err
		}
		return newSQLiteAgentCopilotProfileRepository(sqliteRuntime.DB()), func() {}, nil
	}
	if mode != "postgres_dev_test" {
		return nil, nil, errors.New("unsupported agent copilot profile store mode")
	}
	if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.AgentCopilotProfileDevHTTPEnabled || !cfg.AgentCopilotProfileDevWriteEnabled || strings.TrimSpace(cfg.AgentCopilotProfileDatabaseURL) == "" {
		return nil, nil, errors.New("postgres_dev_test agent copilot profile config is incomplete")
	}
	timeout := cfg.AgentCopilotProfileDatabaseTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pool, err := agentcopilotprofilemigrations.OpenPool(ctx, cfg.AgentCopilotProfileDatabaseURL)
	if err != nil {
		return nil, nil, err
	}
	closePool := func() { pool.Close() }
	state, err := agentcopilotprofilemigrations.Inspect(ctx, pool)
	if err != nil {
		closePool()
		return nil, nil, err
	}
	if state.MigrationState != agentcopilotprofilemigrations.MigrationStateApplied {
		closePool()
		return nil, nil, errors.New("agent copilot profile PostgreSQL migration is not applied")
	}
	return newPostgresAgentCopilotProfileRepository(pool), closePool, nil
}
