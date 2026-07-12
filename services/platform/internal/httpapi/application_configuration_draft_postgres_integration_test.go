//go:build postgres_integration

package httpapi

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	applicationdraftmigrations "radishmind.local/services/platform/migrations/application_configuration_drafts"
)

func TestApplicationConfigurationDraftPostgresLifecycle(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := applicationdraftmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open application draft test database: %v", err)
	}
	defer pool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, pool)
	if _, err := applicationdraftmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatalf("reset application draft migration: %v", err)
	}
	state, err := applicationdraftmigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != applicationdraftmigrations.MigrationStateApplied {
		t.Fatalf("apply application draft migration: %#v %v", state, err)
	}
	runtimePool, err := applicationdraftmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open application draft runtime pool: %v", err)
	}
	defer runtimePool.Close()
	defer func() { _, _ = applicationdraftmigrations.RollbackForDevTest(context.Background(), pool) }()

	requestContext := validApplicationDraftContext()
	payload := validApplicationDraftPayload()
	service := newApplicationConfigurationDraftService(newPostgresApplicationConfigurationDraftRepository(runtimePool))
	created := service.Save(requestContext, payload, 0)
	if created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("create application draft in PostgreSQL: %#v", created)
	}

	restartedService := newApplicationConfigurationDraftService(newPostgresApplicationConfigurationDraftRepository(runtimePool))
	restored := restartedService.Read(requestContext, payload.DraftID)
	if restored.Draft == nil || restored.Draft.DraftVersion != 1 || restored.Draft.DisplayName != payload.DisplayName {
		t.Fatalf("restore application draft after service reconstruction: %#v", restored)
	}
	payload.Description = "PostgreSQL CAS update"
	updated := restartedService.Save(requestContext, payload, 1)
	if updated.Draft == nil || updated.Draft.DraftVersion != 2 {
		t.Fatalf("update application draft in PostgreSQL: %#v", updated)
	}
	conflict := restartedService.Save(requestContext, payload, 1)
	if conflict.FailureCode != ApplicationDraftFailureVersionConflict || conflict.CurrentDraftVersion != 2 {
		t.Fatalf("PostgreSQL CAS conflict did not preserve current version: %#v", conflict)
	}
	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if denied := restartedService.Read(otherOwner, payload.DraftID); denied.FailureCode != ApplicationDraftFailureNotFound {
		t.Fatalf("PostgreSQL owner isolation failed: %#v", denied)
	}

	if _, err := applicationdraftmigrations.RollbackForDevTest(ctx, pool); err != nil {
		t.Fatalf("rollback application draft migration: %v", err)
	}
	state, err = applicationdraftmigrations.Apply(ctx, pool)
	if err != nil || state.MigrationState != applicationdraftmigrations.MigrationStateApplied {
		t.Fatalf("reapply application draft migration: %#v %v", state, err)
	}
}
