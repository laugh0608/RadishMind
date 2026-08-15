package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

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
	if err = runtime.Close(); err != nil {
		t.Fatalf("close SQLite result artifact runtime: %v", err)
	}

	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restoredRepository := newSQLiteApplicationResultArtifactRepository(restarted.DB())
	restored, err := restoredRepository.Read(ctx, artifact.ArtifactID)
	byTurn, turnErr := restoredRepository.ReadByTurn(ctx, artifact.SessionID, artifact.TurnID)
	listed, listErr := restoredRepository.List(ctx, artifact.SessionID)
	if err != nil || turnErr != nil || listErr != nil || !applicationResultArtifactsEquivalent(restored, artifact) ||
		byTurn.ArtifactID != artifact.ArtifactID || len(listed) != 1 || listed[0].ArtifactID != artifact.ArtifactID {
		t.Fatalf("restore SQLite result artifact: restored=%#v by_turn=%#v listed=%#v errors=%v/%v/%v", restored, byTurn, listed, err, turnErr, listErr)
	}
	otherScope := ctx
	otherScope.OwnerSubjectRef = "subject_other"
	if _, err = restoredRepository.Read(otherScope, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactNotFound) {
		t.Fatalf("cross-scope SQLite result artifact read did not fail closed: %v", err)
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
