//go:build postgres_integration

package httpapi

import (
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	applicationcatalogmigrations "radishmind.local/services/platform/migrations/application_catalog_records"
)

func TestApplicationCatalogPostgresLifecycle(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	migrationPool, err := applicationcatalogmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open application catalog migration pool: %v", err)
	}
	defer migrationPool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, migrationPool)
	if _, err := applicationcatalogmigrations.RollbackForDevTest(ctx, migrationPool); err != nil {
		t.Fatalf("reset application catalog migration: %v", err)
	}
	state, err := applicationcatalogmigrations.Apply(ctx, migrationPool)
	if err != nil || state.MigrationState != applicationcatalogmigrations.MigrationStateApplied {
		t.Fatalf("apply application catalog migration: %#v %v", state, err)
	}
	if repeated, err := applicationcatalogmigrations.Apply(ctx, migrationPool); err != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat application catalog migration: %#v %v", repeated, err)
	}
	runtimePool, err := applicationcatalogmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open application catalog runtime pool: %v", err)
	}
	defer runtimePool.Close()
	defer func() { _, _ = applicationcatalogmigrations.RollbackForDevTest(context.Background(), migrationPool) }()

	requestContext := applicationCatalogTestContext("subject_catalog_owner")
	repository := newPostgresApplicationCatalogRepository(runtimePool)
	service := newApplicationCatalogService(repository)
	service.newID = func() (string, error) { return "app_aaaaaaaaaaaaaaaa", nil }
	service.now = func() time.Time { return time.Date(2026, 7, 13, 15, 0, 0, 0, time.UTC) }
	created := service.Create(requestContext, ApplicationCatalogCreateInput{DisplayName: "PostgreSQL App", Description: "Persistent catalog record", ApplicationKind: "agent"})
	if created.Record == nil || created.Record.RecordVersion != 1 {
		t.Fatalf("create application catalog record in PostgreSQL: %#v", created)
	}

	restarted := newApplicationCatalogService(newPostgresApplicationCatalogRepository(runtimePool))
	restarted.now = func() time.Time { return time.Date(2026, 7, 13, 15, 1, 0, 0, time.UTC) }
	restored := restarted.Read(requestContext, created.Record.ApplicationID)
	if restored.Record == nil || restored.Record.DisplayName != created.Record.DisplayName {
		t.Fatalf("restore application catalog record after service reconstruction: %#v", restored)
	}
	updated := restarted.Update(requestContext, created.Record.ApplicationID, ApplicationCatalogUpdateInput{
		ExpectedVersion: 1, DisplayName: "PostgreSQL App v2", Description: "Updated persistent metadata", ApplicationKind: "prompt_application",
	})
	if updated.Record == nil || updated.Record.RecordVersion != 2 {
		t.Fatalf("update application catalog record: %#v", updated)
	}

	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := restarted.Update(requestContext, created.Record.ApplicationID, ApplicationCatalogUpdateInput{
				ExpectedVersion: 2, DisplayName: "Concurrent PostgreSQL App", ApplicationKind: "agent",
			})
			if result.FailureCode == "" {
				successes.Add(1)
			} else if result.FailureCode == ApplicationCatalogFailureVersionConflict {
				conflicts.Add(1)
			} else {
				t.Errorf("unexpected PostgreSQL CAS result: %#v", result)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("PostgreSQL CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}
	current := restarted.Read(requestContext, created.Record.ApplicationID)
	if current.Record == nil || current.Record.RecordVersion != 3 {
		t.Fatalf("read current PostgreSQL record version: %#v", current)
	}
	archived := restarted.Archive(requestContext, created.Record.ApplicationID, 3)
	if archived.Record == nil || archived.Record.RecordVersion != 4 || archived.Record.ArchivedAt == nil {
		t.Fatalf("archive PostgreSQL application record: %#v", archived)
	}
	if active := restarted.RequireActive(requestContext, created.Record.ApplicationID); active.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived PostgreSQL record must fail active requirement: %#v", active)
	}
	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if denied := restarted.Read(otherOwner, created.Record.ApplicationID); denied.FailureCode != ApplicationCatalogFailureNotFound {
		t.Fatalf("PostgreSQL application owner isolation failed: %#v", denied)
	}
	if _, err := runtimePool.Exec(ctx, "CREATE TABLE application_catalog_runtime_ddl_denied(id integer)"); err == nil {
		t.Fatal("runtime role unexpectedly received DDL permission")
	}

	if _, err := applicationcatalogmigrations.RollbackForDevTest(ctx, migrationPool); err != nil {
		t.Fatalf("rollback application catalog migration: %v", err)
	}
	if state, err := applicationcatalogmigrations.Apply(ctx, migrationPool); err != nil || state.MigrationState != applicationcatalogmigrations.MigrationStateApplied {
		t.Fatalf("reapply application catalog migration: %#v %v", state, err)
	}
}
