//go:build postgres_integration

package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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
	profileArtifacts[1].ContentType = "application/json"
	profileArtifacts[1].Content = `{"result":"structured workflow"}`
	profileArtifacts[1].ContentBytes = len([]byte(profileArtifacts[1].Content))
	profileArtifacts[1].ContentDigest = applicationResultArtifactContentDigest(profileArtifacts[1].ContentType, profileArtifacts[1].Content)
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
	initialLifecycle, lifecycleErr := repository.ReadLifecycle(interactionContext, artifact.ArtifactID)
	if lifecycleErr != nil || initialLifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive || initialLifecycle.LifecycleVersion != 1 {
		runtimePool.Close()
		t.Fatalf("read initial PostgreSQL result artifact lifecycle: lifecycle=%#v err=%v", initialLifecycle, lifecycleErr)
	}
	type lifecycleOutcome struct{ err error }
	lifecycleOutcomes := make(chan lifecycleOutcome, 8)
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _, transitionErr := repository.TransitionLifecycle(
				interactionContext, artifact.ArtifactID, ApplicationResultArtifactLifecycleArchived, 1,
				time.Date(2026, 8, 17, 9, 30, 0, 123456000, time.UTC),
			)
			lifecycleOutcomes <- lifecycleOutcome{err: transitionErr}
		}()
	}
	wait.Wait()
	close(lifecycleOutcomes)
	lifecycleSuccesses, lifecycleConflicts := 0, 0
	for result := range lifecycleOutcomes {
		switch {
		case result.err == nil:
			lifecycleSuccesses++
		case errors.Is(result.err, errApplicationResultArtifactLifecycleVersion):
			lifecycleConflicts++
		default:
			runtimePool.Close()
			t.Fatalf("unexpected PostgreSQL lifecycle CAS error: %v", result.err)
		}
	}
	if lifecycleSuccesses != 1 || lifecycleConflicts != 7 {
		runtimePool.Close()
		t.Fatalf("PostgreSQL lifecycle CAS did not converge: successes=%d conflicts=%d", lifecycleSuccesses, lifecycleConflicts)
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE application_result_artifact_lifecycle_events SET actor_ref=actor_ref WHERE artifact_id=$1`, artifact.ArtifactID); err == nil {
		runtimePool.Close()
		t.Fatal("PostgreSQL lifecycle event accepted an update")
	}
	if _, err = runtimePool.Exec(ctx, `DELETE FROM application_result_artifact_lifecycle_events WHERE artifact_id=$1`, artifact.ArtifactID); err == nil {
		runtimePool.Close()
		t.Fatal("PostgreSQL lifecycle event accepted a delete")
	}
	runtimePool.Close()

	restartedPool, err := workflowrunmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	restartedRepository := newPostgresApplicationResultArtifactRepository(restartedPool)
	restartedService := newApplicationResultArtifactService(restartedRepository)
	restartedService.now = func() time.Time { return time.Date(2026, 8, 17, 10, 30, 0, 123456000, time.UTC) }
	restored, readErr := restartedRepository.Read(interactionContext, artifact.ArtifactID)
	listed, listErr := restartedRepository.List(interactionContext, artifact.SessionID)
	restoredLifecycle, lifecycleReadErr := restartedRepository.ReadLifecycle(interactionContext, artifact.ArtifactID)
	archivedRecords, archivedListErr := restartedRepository.ListByLifecycle(interactionContext, applicationResultArtifactRepositoryListFilter{
		SessionID:      artifact.SessionID,
		LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	if readErr != nil || listErr != nil || !applicationResultArtifactsEquivalent(restored, artifact) ||
		lifecycleReadErr != nil || archivedListErr != nil || len(listed) != 2 || listed[1].ArtifactID != artifact.ArtifactID ||
		restoredLifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived || restoredLifecycle.LifecycleVersion != 2 ||
		len(archivedRecords) != 1 || archivedRecords[0].Artifact.ArtifactID != artifact.ArtifactID {
		restartedPool.Close()
		t.Fatalf("restart PostgreSQL result artifact: restored=%#v lifecycle=%#v listed=%#v archived=%#v errors=%v/%v/%v/%v", restored, restoredLifecycle, listed, archivedRecords, readErr, listErr, lifecycleReadErr, archivedListErr)
	}
	applicationActive := restartedService.ListApplication(interactionContext, ApplicationResultArtifactListInput{})
	jsonResults := restartedService.ListApplication(interactionContext, ApplicationResultArtifactListInput{ContentType: "application/json"})
	applicationArchived := restartedService.ListApplication(interactionContext, ApplicationResultArtifactListInput{
		LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	exported := restartedService.Export(interactionContext, profileArtifacts[1].ArtifactID)
	if applicationActive.FailureCode != "" || len(applicationActive.Items) != 5 ||
		jsonResults.FailureCode != "" || len(jsonResults.Items) != 1 || jsonResults.Items[0].ArtifactID != profileArtifacts[1].ArtifactID ||
		applicationArchived.FailureCode != "" || len(applicationArchived.Items) != 1 || applicationArchived.Items[0].ArtifactID != artifact.ArtifactID ||
		exported.FailureCode != "" || exported.Export == nil || exported.Export.Artifact.Content != profileArtifacts[1].Content ||
		exported.Export.ExportDigest != applicationResultArtifactExportDigest(*exported.Export) {
		restartedPool.Close()
		t.Fatalf("restart PostgreSQL application library/export: active=%#v json=%#v archived=%#v export=%#v", applicationActive, jsonResults, applicationArchived, exported)
	}
	var applicationIndexExists bool
	if err = restartedPool.QueryRow(ctx, "SELECT to_regclass('public.application_result_artifacts_application_history_idx') IS NOT NULL").Scan(&applicationIndexExists); err != nil || !applicationIndexExists {
		restartedPool.Close()
		t.Fatalf("PostgreSQL application result history index is missing: exists=%v err=%v", applicationIndexExists, err)
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
	assertPostgresApplicationResultArtifactLifecycleBackfill(t, ctx, adminPool, interactionContext, artifact)
}

func assertPostgresApplicationResultArtifactLifecycleBackfill(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	interactionContext ApplicationInteractionContext,
	artifact ApplicationResultArtifact,
) {
	t.Helper()
	manifestPayload, err := os.ReadFile("../../migrations/workflow_runs/manifest.json")
	if err != nil {
		t.Fatalf("read workflow migration manifest for lifecycle backfill: %v", err)
	}
	var manifest struct {
		Up []string `json:"up"`
	}
	if err = json.Unmarshal(manifestPayload, &manifest); err != nil {
		t.Fatalf("decode workflow migration manifest for lifecycle backfill: %v", err)
	}
	throughV25 := make([]string, 0, len(manifest.Up)-1)
	for _, filename := range manifest.Up {
		if filename == "0026_application_result_artifact_lifecycle.up.sql" {
			break
		}
		migrationPayload, readErr := os.ReadFile("../../migrations/workflow_runs/" + filename)
		if readErr != nil {
			t.Fatalf("read workflow migration %s for lifecycle backfill: %v", filename, readErr)
		}
		throughV25 = append(throughV25, string(migrationPayload))
	}
	if len(throughV25) != 25 {
		t.Fatalf("unexpected pre-lifecycle PostgreSQL migration count: %d", len(throughV25))
	}
	preLifecycleSQL := strings.Join(throughV25, "\n")
	if _, err = adminPool.Exec(ctx, preLifecycleSQL); err != nil {
		t.Fatalf("apply pre-lifecycle PostgreSQL migrations: %v", err)
	}
	payload, err := encodeApplicationResultArtifact(interactionContext, artifact)
	if err != nil {
		t.Fatalf("encode pre-lifecycle PostgreSQL artifact: %v", err)
	}
	createdAt := parseApplicationInteractionTimestamp(artifact.CreatedAt)
	if createdAt == nil {
		t.Fatal("pre-lifecycle PostgreSQL artifact timestamp is invalid")
	}
	if _, err = adminPool.Exec(ctx, `INSERT INTO application_result_artifacts
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,session_id,turn_id,client_turn_key,
execution_profile,run_id,run_schema_version,content_type,content_bytes,content_digest,created_at,artifact_payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		interactionContext.TenantRef, interactionContext.WorkspaceID, interactionContext.ApplicationID,
		interactionContext.OwnerSubjectRef, artifact.ArtifactID, artifact.SessionID, artifact.TurnID,
		artifact.ClientTurnKey, artifact.ExecutionProfile, artifact.RunRef.RunID, artifact.RunRef.SchemaVersion,
		artifact.ContentType, artifact.ContentBytes, artifact.ContentDigest, createdAt.Round(time.Microsecond), payload); err != nil {
		t.Fatalf("seed pre-lifecycle PostgreSQL artifact: %v", err)
	}
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(preLifecycleSQL)))
	if _, err = adminPool.Exec(ctx, `CREATE TABLE IF NOT EXISTS workflow_run_schema_versions (
component text PRIMARY KEY, migration_id text NOT NULL, store_schema_version text NOT NULL,
migration_checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatalf("create pre-lifecycle PostgreSQL migration marker: %v", err)
	}
	if _, err = adminPool.Exec(ctx, `INSERT INTO workflow_run_schema_versions
(component,migration_id,store_schema_version,migration_checksum) VALUES ($1,$2,$3,$4)`,
		workflowrunmigrations.Component, "0025_application_result_artifacts", "workflow_run_store_v25", checksum); err != nil {
		t.Fatalf("write pre-lifecycle PostgreSQL migration marker: %v", err)
	}
	state, err := workflowrunmigrations.Inspect(ctx, adminPool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStatePending {
		t.Fatalf("pre-lifecycle PostgreSQL marker was not pending: state=%#v err=%v", state, err)
	}
	state, err = workflowrunmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied {
		t.Fatalf("upgrade PostgreSQL artifact lifecycle: state=%#v err=%v", state, err)
	}
	var applicationIndexExists bool
	if err = adminPool.QueryRow(ctx, "SELECT to_regclass('public.application_result_artifacts_application_history_idx') IS NOT NULL").Scan(&applicationIndexExists); err != nil || !applicationIndexExists {
		t.Fatalf("upgrade PostgreSQL application result history index: exists=%v err=%v", applicationIndexExists, err)
	}
	var lifecycleState string
	var lifecycleVersion int
	var lifecyclePayload []byte
	if err = adminPool.QueryRow(ctx, `SELECT lifecycle_state,lifecycle_version,lifecycle_payload
FROM application_result_artifact_lifecycles
WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4 AND artifact_id=$5`,
		interactionContext.TenantRef, interactionContext.WorkspaceID, interactionContext.ApplicationID,
		interactionContext.OwnerSubjectRef, artifact.ArtifactID).Scan(&lifecycleState, &lifecycleVersion, &lifecyclePayload); err != nil {
		t.Fatalf("read backfilled PostgreSQL artifact lifecycle: %v", err)
	}
	lifecycle, err := decodeApplicationResultArtifactLifecycle(interactionContext, lifecyclePayload)
	if err != nil || lifecycleState != string(ApplicationResultArtifactLifecycleActive) || lifecycleVersion != 1 ||
		lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive || lifecycle.LifecycleVersion != 1 ||
		lifecycle.UpdatedAt != artifact.CreatedAt || lifecycle.UpdatedByActorRef != artifact.CreatedByActorRef {
		t.Fatalf("PostgreSQL lifecycle backfill drifted: state=%s version=%d lifecycle=%#v err=%v", lifecycleState, lifecycleVersion, lifecycle, err)
	}
}
