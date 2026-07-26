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
	promptapplicationtemplatemigrations "radishmind.local/services/platform/migrations/prompt_application_templates"
)

func TestPromptApplicationTemplatePostgresPersistenceContract(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	databaseContext, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	adminPool, err := promptapplicationtemplatemigrations.OpenPool(databaseContext, databaseURL)
	if err != nil {
		t.Fatalf("open prompt application template migration pool: %v", err)
	}
	defer adminPool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, databaseContext, adminPool)
	if _, err := promptapplicationtemplatemigrations.RollbackForDevTest(databaseContext, adminPool); err != nil {
		t.Fatalf("reset prompt application template migration: %v", err)
	}
	preparePostgresIntegrationRuntimeRole(t, databaseContext, adminPool, runtimeUser)
	state, err := promptapplicationtemplatemigrations.Apply(databaseContext, adminPool)
	if err != nil || state.MigrationState != promptapplicationtemplatemigrations.MigrationStateApplied || state.MigrationChecksum != promptapplicationtemplatemigrations.ExpectedChecksum() {
		t.Fatalf("apply prompt application template migration: state=%#v error=%v", state, err)
	}
	if repeated, err := promptapplicationtemplatemigrations.Apply(databaseContext, adminPool); err != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat prompt application template migration: state=%#v error=%v", repeated, err)
	}
	runtimePool, err := promptapplicationtemplatemigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open prompt application template runtime pool: %v", err)
	}
	defer func() {
		runtimePool.Close()
		_, _ = promptapplicationtemplatemigrations.RollbackForDevTest(context.Background(), adminPool)
	}()
	if _, err := runtimePool.Exec(databaseContext, "CREATE TABLE public.prompt_application_template_runtime_must_not_create (id bigint)"); err == nil {
		_, _ = adminPool.Exec(databaseContext, "DROP TABLE IF EXISTS public.prompt_application_template_runtime_must_not_create")
		t.Fatal("prompt application template runtime role unexpectedly created a table")
	} else {
		var postgresError *pgconn.PgError
		if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
			t.Fatalf("runtime DDL denial returned an unexpected database error")
		}
	}

	requestContext := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	service := newPromptApplicationTemplateService(newPostgresPromptApplicationTemplateRepository(runtimePool))
	service.now = func() time.Time { return time.Date(2026, 7, 21, 10, 0, 0, 0, time.UTC) }
	created := service.SaveDraft(requestContext, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("create PostgreSQL prompt application template draft: %#v", created)
	}
	immutable := service.CreateVersion(requestContext, input.TemplateID, 1)
	if immutable.FailureCode != "" || immutable.Version == nil || immutable.Version.TemplateVersion != 1 {
		t.Fatalf("create PostgreSQL immutable prompt application template version: %#v", immutable)
	}
	if _, err := runtimePool.Exec(databaseContext, `UPDATE prompt_application_template_versions SET template_digest=template_digest
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5 AND template_version=1`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, input.TemplateID); err == nil {
		t.Fatal("PostgreSQL immutable version accepted an update")
	}

	input.Description = "concurrent PostgreSQL update"
	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			result := service.SaveDraft(requestContext, input, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case PromptApplicationTemplateFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected PostgreSQL concurrent failure for writer %d: %#v", index, result)
			}
		}(index)
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("PostgreSQL CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}

	restarted := newPromptApplicationTemplateService(newPostgresPromptApplicationTemplateRepository(runtimePool))
	restored := restarted.ReadVersion(requestContext, input.TemplateID, 1)
	if restored.FailureCode != "" || restored.Version == nil || restored.Version.TemplateDigest != created.Draft.TemplateDigest {
		t.Fatalf("restore PostgreSQL prompt application template version after service reconstruction: %#v", restored)
	}
	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if read := restarted.ReadDraft(otherOwner, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureNotFound {
		t.Fatalf("PostgreSQL repository leaked a draft across owner scope: %#v", read)
	}
	forbidden := validPromptApplicationTemplateDraftInput()
	forbidden.TemplateID = "ptpl_bbbbbbbbbbbbbbbb"
	forbidden.Variables[1].DefaultValue = "Authorization: Bearer postgres-secret-sentinel"
	if result := restarted.SaveDraft(requestContext, forbidden, 0); result.FailureCode != PromptApplicationTemplateFailureSecretForbidden {
		t.Fatalf("PostgreSQL owner accepted credential-like template material: %#v", result)
	}
	var storedPayloads string
	if err := runtimePool.QueryRow(databaseContext, `SELECT COALESCE(string_agg(sanitized_draft_payload::text,''),'') FROM prompt_application_template_drafts`).Scan(&storedPayloads); err != nil {
		t.Fatalf("scan PostgreSQL prompt application template payloads: %v", err)
	}
	for _, forbiddenMaterial := range []string{"postgres-secret-sentinel", "rendered-output-sentinel", "runtime-variable-sentinel"} {
		if strings.Contains(storedPayloads, forbiddenMaterial) {
			t.Fatalf("PostgreSQL prompt application template owner persisted forbidden material %q", forbiddenMaterial)
		}
	}
	if _, err := adminPool.Exec(databaseContext, `UPDATE prompt_application_template_drafts SET
		draft_version=draft_version+1,
		updated_at=updated_at+interval '1 microsecond',
		sanitized_draft_payload=jsonb_set(jsonb_set(sanitized_draft_payload,'{draft_version}',to_jsonb(draft_version+1)),'{messages,0,content}','"corrupted without digest update"'::jsonb)
		WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND template_id=$5`,
		requestContext.TenantRef, requestContext.WorkspaceID, requestContext.ApplicationID, requestContext.OwnerSubjectRef, input.TemplateID); err != nil {
		t.Fatalf("inject structurally valid PostgreSQL digest drift: %v", err)
	}
	if read := restarted.ReadDraft(requestContext, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureStoreContract {
		t.Fatalf("PostgreSQL read accepted stored digest drift: %#v", read)
	}

	if _, err := adminPool.Exec(databaseContext, `UPDATE prompt_application_template_schema_versions SET migration_checksum='sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff' WHERE component=$1`, promptapplicationtemplatemigrations.Component); err != nil {
		t.Fatalf("inject prompt application template marker mismatch: %v", err)
	}
	if inspected, err := promptapplicationtemplatemigrations.Inspect(databaseContext, adminPool); err != nil || inspected.MigrationState != promptapplicationtemplatemigrations.MigrationStateMismatch {
		t.Fatalf("prompt application template checksum mismatch was not detected: state=%#v error=%v", inspected, err)
	}
	if _, err := adminPool.Exec(databaseContext, `UPDATE prompt_application_template_schema_versions SET migration_checksum=$1 WHERE component=$2`, promptapplicationtemplatemigrations.ExpectedChecksum(), promptapplicationtemplatemigrations.Component); err != nil {
		t.Fatalf("restore prompt application template marker checksum: %v", err)
	}

	runtimePool.Close()
	if read := restarted.ReadVersion(requestContext, input.TemplateID, 1); read.FailureCode != PromptApplicationTemplateFailureStoreUnavailable || read.Version != nil {
		t.Fatalf("closed PostgreSQL owner fell back to memory state: %#v", read)
	}
	if rolledBack, err := promptapplicationtemplatemigrations.RollbackForDevTest(databaseContext, adminPool); err != nil || rolledBack.MigrationState != promptapplicationtemplatemigrations.MigrationStateNotApplied {
		t.Fatalf("rollback prompt application template migration: state=%#v error=%v", rolledBack, err)
	}
	if reapplied, err := promptapplicationtemplatemigrations.Apply(databaseContext, adminPool); err != nil || reapplied.MigrationState != promptapplicationtemplatemigrations.MigrationStateApplied {
		t.Fatalf("reapply prompt application template migration: state=%#v error=%v", reapplied, err)
	}
}
