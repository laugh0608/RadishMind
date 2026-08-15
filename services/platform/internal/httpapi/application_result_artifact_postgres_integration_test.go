//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	workflowrunmigrations "radishmind.local/services/platform/migrations/workflow_runs"
)

func TestPostgresApplicationResultArtifactRestartConcurrencyRoleAndMigrationLifecycle(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	adminPool, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresWorkflowRunSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, adminPool)
		adminPool.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied ||
		state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply result artifact migration: state=%#v err=%v", state, err)
	}

	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = runtimePool.Exec(ctx, "CREATE TABLE application_result_artifact_runtime_must_not_create (id integer)"); err == nil {
		runtimePool.Close()
		t.Fatal("application result artifact runtime role unexpectedly received DDL permission")
	}
	repository := newPostgresApplicationResultArtifactRepository(runtimePool)
	interactionContext, artifact := applicationResultArtifactPersistenceFixture()
	interactionContext.RequestContext = ctx
	if _, replay, createErr := repository.Create(interactionContext, artifact); createErr != nil || replay {
		runtimePool.Close()
		t.Fatalf("create PostgreSQL result artifact preflight: replay=%v err=%v", replay, createErr)
	}
	concurrentArtifact := artifact
	concurrentArtifact.ArtifactID = "appres_bbbbbbbbbbbbbbbb"
	concurrentArtifact.TurnID = "appturn_bbbbbbbbbbbbbbbb"
	concurrentArtifact.ClientTurnKey = "client_turn_2"

	type outcome struct {
		replay bool
		err    error
	}
	outcomes := make(chan outcome, 8)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, replay, createErr := repository.Create(interactionContext, concurrentArtifact)
			outcomes <- outcome{replay: replay, err: createErr}
		}()
	}
	wait.Wait()
	close(outcomes)
	creates, replays := 0, 0
	for result := range outcomes {
		if result.err != nil {
			runtimePool.Close()
			t.Fatalf("concurrent PostgreSQL result artifact create failed: %v", result.err)
		}
		if result.replay {
			replays++
		} else {
			creates++
		}
	}
	if creates != 1 || replays != 7 {
		runtimePool.Close()
		t.Fatalf("unexpected PostgreSQL result artifact outcomes: creates=%d replays=%d", creates, replays)
	}
	conflict := artifact
	conflict.ArtifactID = "appres_cccccccccccccccc"
	conflict.Content = "different PostgreSQL durable result"
	conflict.ContentBytes = len(conflict.Content)
	conflict.ContentDigest = applicationResultArtifactContentDigest(conflict.ContentType, conflict.Content)
	if _, _, err = repository.Create(interactionContext, conflict); !errors.Is(err, errApplicationResultArtifactConflict) {
		runtimePool.Close()
		t.Fatalf("same-turn PostgreSQL result artifact conflict was not rejected: %v", err)
	}
	_, profileArtifacts := applicationResultArtifactPersistenceProfileFixtures()
	for _, profileArtifact := range profileArtifacts[1:] {
		if _, replay, createErr := repository.Create(interactionContext, profileArtifact); createErr != nil || replay {
			runtimePool.Close()
			t.Fatalf("create PostgreSQL %s result artifact: replay=%v err=%v", profileArtifact.ExecutionProfile, replay, createErr)
		}
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE application_result_artifacts SET content_bytes=content_bytes WHERE artifact_id=$1`, artifact.ArtifactID); err == nil {
		runtimePool.Close()
		t.Fatal("PostgreSQL runtime role mutated an immutable result artifact")
	}
	if _, err = runtimePool.Exec(ctx, `DELETE FROM application_result_artifacts WHERE artifact_id=$1`, artifact.ArtifactID); err == nil {
		runtimePool.Close()
		t.Fatal("PostgreSQL runtime role deleted an immutable result artifact")
	}
	runtimePool.Close()

	restartedPool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	restartedRepository := newPostgresApplicationResultArtifactRepository(restartedPool)
	restored, readErr := restartedRepository.Read(interactionContext, artifact.ArtifactID)
	listed, listErr := restartedRepository.List(interactionContext, artifact.SessionID)
	if readErr != nil || listErr != nil || !applicationResultArtifactsEquivalent(restored, artifact) ||
		len(listed) != 2 || listed[1].ArtifactID != artifact.ArtifactID {
		restartedPool.Close()
		t.Fatalf("restart PostgreSQL result artifact: restored=%#v listed=%#v errors=%v/%v", restored, listed, readErr, listErr)
	}
	for _, profileArtifact := range profileArtifacts[1:] {
		profileRestored, profileReadErr := restartedRepository.Read(interactionContext, profileArtifact.ArtifactID)
		if profileReadErr != nil || !applicationResultArtifactsEquivalent(profileRestored, profileArtifact) {
			restartedPool.Close()
			t.Fatalf("restart PostgreSQL %s result artifact: restored=%#v err=%v", profileArtifact.ExecutionProfile, profileRestored, profileReadErr)
		}
	}
	otherScope := interactionContext
	otherScope.OwnerSubjectRef = "subject_other"
	if _, err = restartedRepository.Read(otherScope, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactNotFound) {
		restartedPool.Close()
		t.Fatalf("cross-scope PostgreSQL result artifact read did not fail closed: %v", err)
	}
	restartedPool.Close()

	rolledBack, err := workflowrunmigrations.RollbackForDevTest(ctx, adminPool)
	if err != nil || rolledBack.MigrationState != workflowrunmigrations.MigrationStateNotApplied {
		t.Fatalf("rollback result artifact migration chain: state=%#v err=%v", rolledBack, err)
	}
	reapplied, err := workflowrunmigrations.Apply(ctx, adminPool)
	if err != nil || reapplied.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("reapply result artifact migration chain: state=%#v err=%v", reapplied, err)
	}
}
