package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteApplicationResultArtifactLifecycleMigrationBackfillsExistingArtifacts(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application-result-artifact-lifecycle-backfill.db")
	migrations := sqliteworkflowrunmigrations.Migrations()
	legacyRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: migrations[:len(migrations)-1],
	})
	if err != nil {
		t.Fatalf("open pre-lifecycle SQLite runtime: %v", err)
	}
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	payload, err := encodeApplicationResultArtifact(ctx, artifact)
	if err != nil {
		_ = legacyRuntime.Close()
		t.Fatalf("encode pre-lifecycle SQLite artifact: %v", err)
	}
	createdAt := parseApplicationInteractionTimestamp(artifact.CreatedAt)
	if createdAt == nil {
		_ = legacyRuntime.Close()
		t.Fatal("pre-lifecycle SQLite artifact timestamp is invalid")
	}
	if _, err = legacyRuntime.DB().ExecContext(context.Background(), `INSERT INTO application_result_artifacts
(tenant_ref,workspace_id,application_id,owner_subject_ref,artifact_id,session_id,turn_id,client_turn_key,
execution_profile,run_id,run_schema_version,content_type,content_bytes,content_digest,created_at_unix_nano,artifact_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, artifact.ArtifactID,
		artifact.SessionID, artifact.TurnID, artifact.ClientTurnKey, artifact.ExecutionProfile,
		artifact.RunRef.RunID, artifact.RunRef.SchemaVersion, artifact.ContentType, artifact.ContentBytes,
		artifact.ContentDigest, createdAt.UnixNano(), string(payload)); err != nil {
		_ = legacyRuntime.Close()
		t.Fatalf("seed pre-lifecycle SQLite artifact: %v", err)
	}
	if err = legacyRuntime.Close(); err != nil {
		t.Fatalf("close pre-lifecycle SQLite runtime: %v", err)
	}

	upgradedRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: migrations,
	})
	if err != nil {
		t.Fatalf("upgrade SQLite artifact lifecycle: %v", err)
	}
	defer upgradedRuntime.Close()
	repository := newSQLiteApplicationResultArtifactRepository(upgradedRuntime.DB())
	lifecycle, err := repository.ReadLifecycle(ctx, artifact.ArtifactID)
	if err != nil || lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive ||
		lifecycle.LifecycleVersion != 1 || lifecycle.UpdatedAt != artifact.CreatedAt ||
		lifecycle.UpdatedByActorRef != artifact.CreatedByActorRef {
		t.Fatalf("SQLite lifecycle backfill drifted: lifecycle=%#v err=%v", lifecycle, err)
	}
}

func TestSQLiteApplicationResultArtifactPersistsAcrossRestartAndIsImmutable(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application-result-artifacts.db")
	runtime := openApplicationInteractionSQLiteRuntime(t, databasePath)
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifact := applicationResultArtifactPersistenceFixture()

	created, replay, err := repository.Create(ctx, artifact)
	if err != nil || replay || !applicationResultArtifactsEquivalent(created, artifact) {
		t.Fatalf("create SQLite result artifact: created=%#v replay=%v err=%v", created, replay, err)
	}
	replayed, replay, err := repository.Create(ctx, artifact)
	if err != nil || !replay || replayed.ArtifactID != artifact.ArtifactID {
		t.Fatalf("replay SQLite result artifact: replayed=%#v replay=%v err=%v", replayed, replay, err)
	}
	conflict := artifact
	conflict.ArtifactID = "appres_bbbbbbbbbbbbbbbb"
	conflict.Content = "different durable result"
	conflict.ContentBytes = len(conflict.Content)
	conflict.ContentDigest = applicationResultArtifactContentDigest(conflict.ContentType, conflict.Content)
	if _, _, err = repository.Create(ctx, conflict); !errors.Is(err, errApplicationResultArtifactConflict) {
		t.Fatalf("same-turn conflict was not rejected: %v", err)
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `UPDATE application_result_artifacts SET content_bytes=content_bytes WHERE artifact_id=?`, artifact.ArtifactID); err == nil {
		t.Fatal("SQLite result artifact accepted an update")
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `DELETE FROM application_result_artifacts WHERE artifact_id=?`, artifact.ArtifactID); err == nil {
		t.Fatal("SQLite result artifact accepted a delete")
	}
	initialLifecycle, err := repository.ReadLifecycle(ctx, artifact.ArtifactID)
	if err != nil || initialLifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive || initialLifecycle.LifecycleVersion != 1 {
		t.Fatalf("read initial SQLite result artifact lifecycle: lifecycle=%#v err=%v", initialLifecycle, err)
	}
	archived, event, err := repository.TransitionLifecycle(
		ctx, artifact.ArtifactID, ApplicationResultArtifactLifecycleArchived, 1,
		time.Date(2026, 8, 17, 9, 0, 0, 123456789, time.UTC),
	)
	if err != nil || archived.LifecycleState != ApplicationResultArtifactLifecycleArchived ||
		archived.LifecycleVersion != 2 || event.TransitionKind != ApplicationResultArtifactLifecycleTransitionArchived {
		t.Fatalf("archive SQLite result artifact: lifecycle=%#v event=%#v err=%v", archived, event, err)
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `UPDATE application_result_artifact_lifecycle_events SET actor_ref=actor_ref WHERE artifact_id=?`, artifact.ArtifactID); err == nil {
		t.Fatal("SQLite lifecycle event accepted an update")
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `DELETE FROM application_result_artifact_lifecycle_events WHERE artifact_id=?`, artifact.ArtifactID); err == nil {
		t.Fatal("SQLite lifecycle event accepted a delete")
	}
	if err = runtime.Close(); err != nil {
		t.Fatalf("close SQLite result artifact runtime: %v", err)
	}

	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restoredRepository := newSQLiteApplicationResultArtifactRepository(restarted.DB())
	restored, err := restoredRepository.Read(ctx, artifact.ArtifactID)
	byTurn, turnErr := restoredRepository.ReadByTurn(ctx, artifact.SessionID, artifact.TurnID)
	listed, listErr := restoredRepository.List(ctx, artifact.SessionID)
	restoredLifecycle, lifecycleErr := restoredRepository.ReadLifecycle(ctx, artifact.ArtifactID)
	archivedRecords, archivedListErr := restoredRepository.ListByLifecycle(ctx, artifact.SessionID, ApplicationResultArtifactLifecycleArchived)
	if err != nil || turnErr != nil || listErr != nil || !applicationResultArtifactsEquivalent(restored, artifact) ||
		lifecycleErr != nil || archivedListErr != nil || byTurn.ArtifactID != artifact.ArtifactID ||
		len(listed) != 1 || listed[0].ArtifactID != artifact.ArtifactID ||
		restoredLifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived ||
		len(archivedRecords) != 1 || archivedRecords[0].Artifact.ArtifactID != artifact.ArtifactID {
		t.Fatalf("restore SQLite result artifact: restored=%#v lifecycle=%#v by_turn=%#v listed=%#v archived=%#v errors=%v/%v/%v/%v/%v", restored, restoredLifecycle, byTurn, listed, archivedRecords, err, turnErr, listErr, lifecycleErr, archivedListErr)
	}
	otherScope := ctx
	otherScope.OwnerSubjectRef = "subject_other"
	if _, err = restoredRepository.Read(otherScope, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactNotFound) {
		t.Fatalf("cross-scope SQLite result artifact read did not fail closed: %v", err)
	}
}

func TestSQLiteApplicationResultArtifactLifecycleConcurrentCASConverges(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-lifecycle-concurrent.db"))
	defer runtime.Close()
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, _, err := repository.Create(ctx, artifact); err != nil {
		t.Fatalf("create SQLite lifecycle concurrency fixture: %v", err)
	}
	type outcome struct{ err error }
	outcomes := make(chan outcome, 8)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _, err := repository.TransitionLifecycle(
				ctx, artifact.ArtifactID, ApplicationResultArtifactLifecycleArchived, 1,
				time.Date(2026, 8, 17, 9, 30, 0, 0, time.UTC),
			)
			outcomes <- outcome{err: err}
		}()
	}
	wait.Wait()
	close(outcomes)
	successes, conflicts := 0, 0
	for result := range outcomes {
		switch {
		case result.err == nil:
			successes++
		case errors.Is(result.err, errApplicationResultArtifactLifecycleVersion):
			conflicts++
		default:
			t.Fatalf("unexpected SQLite lifecycle CAS error: %v", result.err)
		}
	}
	if successes != 1 || conflicts != 7 {
		t.Fatalf("SQLite lifecycle CAS did not converge: successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestSQLiteApplicationResultArtifactLifecycleCorruptionFailsClosed(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-lifecycle-corrupt.db"))
	defer runtime.Close()
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, _, err := repository.Create(ctx, artifact); err != nil {
		t.Fatalf("create SQLite lifecycle corruption fixture: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), "DROP TRIGGER application_result_artifact_lifecycles_controlled_update"); err != nil {
		t.Fatalf("drop test-only SQLite lifecycle trigger: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE application_result_artifact_lifecycles
SET lifecycle_payload=json_set(lifecycle_payload,'$.unexpected_projection','corrupt') WHERE artifact_id=?`, artifact.ArtifactID); err != nil {
		t.Fatalf("corrupt SQLite lifecycle payload: %v", err)
	}
	if _, err := repository.ReadLifecycle(ctx, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactContract) {
		t.Fatalf("corrupt SQLite lifecycle did not fail closed: %v", err)
	}
}

func TestSQLiteApplicationResultArtifactConcurrentCreateConverges(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-concurrent.db"))
	defer runtime.Close()
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifact := applicationResultArtifactPersistenceFixture()

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
			_, replay, err := repository.Create(ctx, artifact)
			outcomes <- outcome{replay: replay, err: err}
		}()
	}
	wait.Wait()
	close(outcomes)
	creates, replays := 0, 0
	for result := range outcomes {
		if result.err != nil {
			t.Fatalf("concurrent SQLite result artifact create failed: %v", result.err)
		}
		if result.replay {
			replays++
		} else {
			creates++
		}
	}
	if creates != 1 || replays != 7 {
		t.Fatalf("unexpected concurrent SQLite result artifact outcomes: creates=%d replays=%d", creates, replays)
	}
}

func TestSQLiteApplicationResultArtifactSupportsEverySessionProfile(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-profiles.db"))
	defer runtime.Close()
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifacts := applicationResultArtifactPersistenceProfileFixtures()

	for _, artifact := range artifacts {
		created, replay, err := repository.Create(ctx, artifact)
		if err != nil || replay || !applicationResultArtifactsEquivalent(created, artifact) {
			t.Fatalf("create SQLite %s result artifact: created=%#v replay=%v err=%v", artifact.ExecutionProfile, created, replay, err)
		}
		restored, err := repository.Read(ctx, artifact.ArtifactID)
		if err != nil || !applicationResultArtifactsEquivalent(restored, artifact) {
			t.Fatalf("read SQLite %s result artifact: restored=%#v err=%v", artifact.ExecutionProfile, restored, err)
		}
	}
}

func TestSQLiteApplicationResultArtifactCorruptionAndClosedStoreFailClosed(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-corrupt.db"))
	repository := newSQLiteApplicationResultArtifactRepository(runtime.DB())
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, _, err := repository.Create(ctx, artifact); err != nil {
		t.Fatalf("create SQLite corruption fixture: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), "DROP TRIGGER application_result_artifacts_no_update"); err != nil {
		t.Fatalf("remove test-only SQLite immutability trigger: %v", err)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE application_result_artifacts
SET artifact_payload=json_set(artifact_payload,'$.unexpected_projection','corrupt') WHERE artifact_id=?`, artifact.ArtifactID); err != nil {
		t.Fatalf("corrupt SQLite result artifact fixture: %v", err)
	}
	if _, err := repository.Read(ctx, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactContract) {
		t.Fatalf("corrupt SQLite result artifact did not fail closed: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite corruption runtime: %v", err)
	}
	if _, err := repository.Read(ctx, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactStore) || strings.Contains(err.Error(), artifact.Content) {
		t.Fatalf("closed SQLite result artifact store leaked or fell back: %v", err)
	}
}

func TestApplicationResultArtifactRepositoryFactoryFollowsRunStoreBackend(t *testing.T) {
	memory, err := newApplicationResultArtifactRepositoryForRunStore(newMemoryWorkflowRunStore(10))
	if err != nil {
		t.Fatalf("create memory result artifact repository: %v", err)
	}
	if _, ok := memory.(*memoryApplicationResultArtifactRepository); !ok {
		t.Fatalf("unexpected memory result artifact repository: %T", memory)
	}
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-result-artifact-factory.db"))
	defer runtime.Close()
	sqliteRepository, err := newApplicationResultArtifactRepositoryForRunStore(newSQLiteWorkflowRunStore(runtime.DB()))
	if err != nil {
		t.Fatalf("create SQLite result artifact repository: %v", err)
	}
	if _, ok := sqliteRepository.(*sqliteApplicationResultArtifactRepository); !ok {
		t.Fatalf("unexpected SQLite result artifact repository: %T", sqliteRepository)
	}
	if _, err = newApplicationResultArtifactRepositoryForRunStore(&sqliteWorkflowRunStore{}); err == nil {
		t.Fatal("result artifact repository accepted a missing shared SQLite database")
	}
	if _, err = newApplicationResultArtifactRepositoryForRunStore(&postgresWorkflowRunStore{}); err == nil {
		t.Fatal("result artifact repository accepted a missing workflow PostgreSQL pool")
	}
}

func applicationResultArtifactPersistenceFixture() (ApplicationInteractionContext, ApplicationResultArtifact) {
	ctx := ApplicationInteractionContext{
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: "application_demo",
		OwnerSubjectRef: "subject_demo", ActorRef: "actor_demo", RequestID: "request_demo", AuditRef: "audit_demo",
		WriteEnabled: true,
	}
	content := "# Durable result\n\nRestart-safe user-owned content."
	artifact := ApplicationResultArtifact{
		SchemaVersion: applicationResultArtifactSchemaVersion, ArtifactID: "appres_aaaaaaaaaaaaaaaa", RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, SessionID: "appsess_aaaaaaaaaaaaaaaa", TurnID: "appturn_aaaaaaaaaaaaaaaa",
		ClientTurnKey: "client_turn_1", ExecutionProfile: applicationInteractionProfileWorkflow,
		RunRef:      ApplicationInteractionRunRef{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: workflowRunRecordDefinitionSchemaVersion},
		ContentType: "text/markdown", Content: content, ContentBytes: len([]byte(content)),
		ContentDigest: applicationResultArtifactContentDigest("text/markdown", content), CreatedAt: time.Date(2026, 8, 15, 9, 0, 0, 123456789, time.UTC).Format(time.RFC3339Nano),
		CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	return ctx, artifact
}

func applicationResultArtifactPersistenceProfileFixtures() (ApplicationInteractionContext, []ApplicationResultArtifact) {
	ctx, workflowArtifact := applicationResultArtifactPersistenceFixture()
	artifacts := []ApplicationResultArtifact{workflowArtifact}
	profiles := []struct {
		suffix        string
		profile       string
		schemaVersion string
	}{
		{suffix: "d", profile: applicationInteractionProfileWorkflowStructured, schemaVersion: workflowRunRecordDefinitionStructuredSchemaVersion},
		{suffix: "e", profile: applicationInteractionProfileRAG, schemaVersion: workflowRunRecordAppRAGSchemaVersion},
		{suffix: "f", profile: applicationInteractionProfilePrompt, schemaVersion: workflowRunRecordPromptSchemaVersion},
		{suffix: "g", profile: applicationInteractionProfileAgentCopilot, schemaVersion: agentCopilotRunV7Schema},
	}
	for index, profile := range profiles {
		artifact := workflowArtifact
		artifact.ArtifactID = "appres_" + strings.Repeat(profile.suffix, 16)
		artifact.SessionID = "appsess_" + strings.Repeat(profile.suffix, 16)
		artifact.TurnID = "appturn_" + strings.Repeat(profile.suffix, 16)
		artifact.ClientTurnKey = "client_turn_" + profile.suffix
		artifact.ExecutionProfile = profile.profile
		artifact.RunRef = ApplicationInteractionRunRef{RunID: "run_" + strings.Repeat(profile.suffix, 16), SchemaVersion: profile.schemaVersion}
		artifact.Content = "Durable result for " + profile.profile
		artifact.ContentBytes = len([]byte(artifact.Content))
		artifact.ContentDigest = applicationResultArtifactContentDigest(artifact.ContentType, artifact.Content)
		artifact.CreatedAt = time.Date(2026, 8, 15, 9, 0, index+1, 123456789, time.UTC).Format(time.RFC3339Nano)
		artifacts = append(artifacts, artifact)
	}
	return ctx, artifacts
}
