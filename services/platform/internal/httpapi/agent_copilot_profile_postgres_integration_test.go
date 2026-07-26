//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"radishmind.local/services/platform/internal/config"
	agentcopilotprofilemigrations "radishmind.local/services/platform/migrations/agent_copilot_profiles"
)

func TestAgentCopilotProfilePostgresPersistenceContract(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	databaseContext, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	adminPool, err := agentcopilotprofilemigrations.OpenPool(databaseContext, databaseURL)
	if err != nil {
		t.Fatalf("open profile migration pool: %v", err)
	}
	defer adminPool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, databaseContext, adminPool)
	if _, err := agentcopilotprofilemigrations.RollbackForDevTest(databaseContext, adminPool); err != nil {
		t.Fatalf("reset profile migration: %v", err)
	}
	preparePostgresIntegrationRuntimeRole(t, databaseContext, adminPool, runtimeUser)
	state, err := agentcopilotprofilemigrations.Apply(databaseContext, adminPool)
	if err != nil || state.MigrationState != agentcopilotprofilemigrations.MigrationStateApplied || state.MigrationChecksum != agentcopilotprofilemigrations.ExpectedChecksum() {
		t.Fatalf("apply profile migration: state=%#v error=%v", state, err)
	}
	if repeated, err := agentcopilotprofilemigrations.Apply(databaseContext, adminPool); err != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat profile migration: state=%#v error=%v", repeated, err)
	}
	factoryRepository, closeFactoryRepository, err := newAgentCopilotProfileRepositoryFromConfig(config.Config{
		ControlPlaneReadDevAuthEnabled: true, AgentCopilotProfileDevHTTPEnabled: true, AgentCopilotProfileDevWriteEnabled: true,
		AgentCopilotProfileStoreMode: "postgres_dev_test", AgentCopilotProfileDatabaseURL: runtimeDatabaseURL,
		AgentCopilotProfileDatabaseTimeout: 5 * time.Second,
	})
	if err != nil || closeFactoryRepository == nil {
		t.Fatalf("construct configured PostgreSQL profile repository: repository=%T error=%v", factoryRepository, err)
	}
	if _, ok := factoryRepository.(*postgresAgentCopilotProfileRepository); !ok {
		t.Fatalf("configured profile repository selected unexpected owner: %T", factoryRepository)
	}
	closeFactoryRepository()
	runtimePool, err := agentcopilotprofilemigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open profile runtime pool: %v", err)
	}
	defer func() {
		runtimePool.Close()
		_, _ = agentcopilotprofilemigrations.RollbackForDevTest(context.Background(), adminPool)
	}()
	if _, err := runtimePool.Exec(databaseContext, "CREATE TABLE public.agent_copilot_profile_runtime_must_not_create (id bigint)"); err == nil {
		_, _ = adminPool.Exec(databaseContext, "DROP TABLE IF EXISTS public.agent_copilot_profile_runtime_must_not_create")
		t.Fatal("profile runtime role unexpectedly created a table")
	} else {
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
			t.Fatal("runtime DDL denial returned an unexpected database error")
		}
	}

	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	service := newAgentCopilotProfileService(newPostgresAgentCopilotProfileRepository(runtimePool))
	service.now = func() time.Time { return time.Date(2026, 7, 25, 13, 0, 0, 0, time.UTC) }
	created := service.SaveDraft(ctx, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("create PostgreSQL profile draft: %#v", created)
	}
	immutable := service.CreateVersion(ctx, input.ProfileID, 1)
	if immutable.FailureCode != "" || immutable.Version == nil || immutable.Version.ProfileVersion != 1 {
		t.Fatalf("create PostgreSQL profile version: %#v", immutable)
	}
	if _, err := runtimePool.Exec(databaseContext, `UPDATE agent_copilot_profile_versions SET profile_digest=profile_digest
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5 AND profile_version=1`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, input.ProfileID); err == nil {
		t.Fatal("PostgreSQL immutable profile version accepted an update")
	}

	input.Description = "concurrent PostgreSQL profile update"
	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			result := service.SaveDraft(ctx, input, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case AgentCopilotProfileFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected PostgreSQL profile CAS failure for writer %d: %#v", index, result)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("PostgreSQL profile CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}

	restarted := newAgentCopilotProfileService(newPostgresAgentCopilotProfileRepository(runtimePool))
	restored := restarted.ReadVersion(ctx, input.ProfileID, 1)
	if restored.FailureCode != "" || restored.Version == nil || restored.Version.ProfileDigest != created.Draft.ProfileDigest {
		t.Fatalf("restore profile version after service reconstruction: %#v", restored)
	}
	otherOwner := ctx
	otherOwner.OwnerSubjectRef = "subject:other"
	otherOwner.ActorRef = "subject:other"
	if read := restarted.ReadDraft(otherOwner, input.ProfileID); read.FailureCode != AgentCopilotProfileFailureNotFound {
		t.Fatalf("PostgreSQL repository leaked profile across owner scope: %#v", read)
	}
	forbidden := agentCopilotProfileTestInput("acpf_bbbbbbbbbbbbbbbb")
	forbidden.Description = "Authorization: Bearer postgres-profile-secret-sentinel"
	if result := restarted.SaveDraft(ctx, forbidden, 0); result.FailureCode != AgentCopilotProfileFailureSecretMaterialForbidden {
		t.Fatalf("PostgreSQL profile owner accepted credential-like material: %#v", result)
	}
	var storedPayloads string
	if err := runtimePool.QueryRow(databaseContext, `SELECT COALESCE(string_agg(sanitized_draft_payload::text,''),'') FROM agent_copilot_profile_drafts`).Scan(&storedPayloads); err != nil {
		t.Fatalf("scan PostgreSQL profile payloads: %v", err)
	}
	if strings.Contains(storedPayloads, "postgres-profile-secret-sentinel") {
		t.Fatal("PostgreSQL profile owner persisted forbidden sensitive material")
	}
	if _, err := adminPool.Exec(databaseContext, `UPDATE agent_copilot_profile_drafts SET
		draft_version=draft_version+1,
		updated_at=updated_at+interval '1 microsecond',
		sanitized_draft_payload=jsonb_set(jsonb_set(sanitized_draft_payload,'{draft_version}',to_jsonb(draft_version+1)),'{description}','"corrupted without digest update"'::jsonb)
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND profile_id=$5`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, input.ProfileID); err != nil {
		t.Fatalf("inject PostgreSQL profile digest drift: %v", err)
	}
	if read := restarted.ReadDraft(ctx, input.ProfileID); read.FailureCode != AgentCopilotProfileFailureDigestDrift {
		t.Fatalf("PostgreSQL read accepted profile digest drift: %#v", read)
	}

	if _, err := adminPool.Exec(databaseContext, `UPDATE agent_copilot_profile_schema_versions SET migration_checksum='sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff' WHERE component=$1`, agentcopilotprofilemigrations.Component); err != nil {
		t.Fatalf("inject profile migration marker mismatch: %v", err)
	}
	if inspected, err := agentcopilotprofilemigrations.Inspect(databaseContext, adminPool); err != nil || inspected.MigrationState != agentcopilotprofilemigrations.MigrationStateMismatch {
		t.Fatalf("profile checksum mismatch was not detected: state=%#v error=%v", inspected, err)
	}
	if _, err := adminPool.Exec(databaseContext, `UPDATE agent_copilot_profile_schema_versions SET migration_checksum=$1 WHERE component=$2`, agentcopilotprofilemigrations.ExpectedChecksum(), agentcopilotprofilemigrations.Component); err != nil {
		t.Fatalf("restore profile marker checksum: %v", err)
	}

	runtimePool.Close()
	if read := restarted.ReadVersion(ctx, input.ProfileID, 1); read.FailureCode != AgentCopilotProfileFailureStoreUnavailable || read.Version != nil {
		t.Fatalf("closed PostgreSQL profile owner fell back to memory: %#v", read)
	}
	if rolledBack, err := agentcopilotprofilemigrations.RollbackForDevTest(databaseContext, adminPool); err != nil || rolledBack.MigrationState != agentcopilotprofilemigrations.MigrationStateNotApplied {
		t.Fatalf("rollback profile migration: state=%#v error=%v", rolledBack, err)
	}
	if reapplied, err := agentcopilotprofilemigrations.Apply(databaseContext, adminPool); err != nil || reapplied.MigrationState != agentcopilotprofilemigrations.MigrationStateApplied {
		t.Fatalf("reapply profile migration: state=%#v error=%v", reapplied, err)
	}
}
