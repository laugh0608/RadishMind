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

func TestSQLiteApplicationInteractionSessionPersistsAcrossRestartAndRejectsRawContent(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application-sessions.db")
	runtime := openApplicationInteractionSQLiteRuntime(t, databasePath)
	service, ctx, binding, bridge := applicationInteractionSessionTestFixture(t)
	service.repository = newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	ids := []string{"appsess_aaaaaaaaaaaaaaaa", "appturn_aaaaaaaaaaaaaaaa"}
	service.newID = func(string) (string, error) {
		value := ids[0]
		ids = ids[1:]
		return value, nil
	}
	now := time.Date(2026, 7, 19, 12, 0, 0, 123, time.UTC)
	service.now = func() time.Time { return now }

	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create SQLite application session: %#v", created)
	}
	sensitiveInput := "private-session-input-must-not-persist"
	reserved := service.ReserveTurn(ctx, created.Session.SessionID, ApplicationInteractionTurnReservationInput{
		ExpectedSessionVersion: 1,
		ClientTurnKey:          "sqlite_turn_1",
		InputDigest:            workflowDefinitionInputDigest(sensitiveInput),
		InputBytes:             len(sensitiveInput),
		StartedAt:              now.Add(time.Second),
	})
	if reserved.FailureCode != "" || reserved.Turn == nil || reserved.Session == nil || reserved.Session.RecordVersion != 2 {
		t.Fatalf("reserve SQLite application turn: %#v", reserved)
	}
	completed := service.CompleteTurn(ctx, created.Session.SessionID, reserved.Turn.TurnID, ApplicationInteractionTurnCompletionInput{
		Status:      string(WorkflowRunStatusSucceeded),
		RunRef:      &ApplicationInteractionRunRef{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: workflowRunRecordDefinitionSchemaVersion},
		CompletedAt: now.Add(2 * time.Second),
	})
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("complete SQLite application turn: %#v", completed)
	}
	if bridge.callCount() != 0 {
		t.Fatalf("SQLite session management called provider: %d", bridge.callCount())
	}
	var sessionPayload, turnPayload string
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT sanitized_session_payload FROM application_interaction_sessions WHERE session_id=?`, created.Session.SessionID).Scan(&sessionPayload); err != nil {
		t.Fatalf("read SQLite session payload: %v", err)
	}
	if err := runtime.DB().QueryRowContext(context.Background(), `SELECT sanitized_turn_payload FROM application_interaction_session_turns WHERE turn_id=?`, reserved.Turn.TurnID).Scan(&turnPayload); err != nil {
		t.Fatalf("read SQLite turn payload: %v", err)
	}
	for _, forbidden := range []string{sensitiveInput, "prompt", "provider_raw_response", "credential", "authorization"} {
		if strings.Contains(strings.ToLower(sessionPayload+turnPayload), strings.ToLower(forbidden)) {
			t.Fatalf("SQLite application session payload persisted forbidden content %q", forbidden)
		}
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close first SQLite application session runtime: %v", err)
	}

	restarted := openApplicationInteractionSQLiteRuntime(t, databasePath)
	defer restarted.Close()
	restartedService := newApplicationInteractionSessionService(newSQLiteApplicationInteractionSessionRepository(restarted.DB()), service.resolver)
	read := restartedService.Read(ctx, created.Session.SessionID)
	turns, failure := restartedService.ListTurns(ctx, created.Session.SessionID)
	if read.FailureCode != "" || read.Session == nil || read.Session.RecordVersion != 2 || failure != "" || len(turns) != 1 || turns[0].Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("read SQLite session after restart: read=%#v turns=%#v failure=%s", read, turns, failure)
	}
}

func TestSQLiteApplicationInteractionSessionCASAllowsOneConcurrentReservation(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "application-session-cas.db"))
	defer runtime.Close()
	service, ctx, binding, _ := applicationInteractionSessionTestFixture(t)
	service.repository = newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create SQLite CAS fixture: %#v", created)
	}
	now := time.Now().UTC()
	results := make(chan ApplicationInteractionSessionResult, 2)
	var wait sync.WaitGroup
	for _, key := range []string{"sqlite_cas_a", "sqlite_cas_b"} {
		wait.Add(1)
		go func(clientKey string) {
			defer wait.Done()
			results <- service.ReserveTurn(ctx, created.Session.SessionID, ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: 1, ClientTurnKey: clientKey, InputDigest: workflowDefinitionInputDigest(clientKey), InputBytes: len(clientKey), StartedAt: now})
		}(key)
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case ApplicationInteractionFailureVersionConflict:
			conflicts++
		default:
			t.Fatalf("unexpected SQLite CAS result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("SQLite CAS outcomes: successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestApplicationInteractionSessionRepositoryFactoryFollowsRunStoreBackend(t *testing.T) {
	memory, err := newApplicationInteractionSessionRepositoryForRunStore(newMemoryWorkflowRunStore(10))
	if err != nil {
		t.Fatalf("create memory application session repository: %v", err)
	}
	if _, ok := memory.(*memoryApplicationInteractionSessionRepository); !ok {
		t.Fatalf("unexpected memory application session repository: %T", memory)
	}
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "factory.db"))
	defer runtime.Close()
	sqliteRepository, err := newApplicationInteractionSessionRepositoryForRunStore(newSQLiteWorkflowRunStore(runtime.DB()))
	if err != nil {
		t.Fatalf("create SQLite application session repository: %v", err)
	}
	if _, ok := sqliteRepository.(*sqliteApplicationInteractionSessionRepository); !ok {
		t.Fatalf("unexpected SQLite application session repository: %T", sqliteRepository)
	}
	if _, err = newApplicationInteractionSessionRepositoryForRunStore(&sqliteWorkflowRunStore{}); err == nil {
		t.Fatal("application session repository accepted a missing shared SQLite database")
	}
	if _, err = newApplicationInteractionSessionRepositoryForRunStore(&postgresWorkflowRunStore{}); err == nil {
		t.Fatal("application session repository accepted a missing workflow PostgreSQL pool")
	}
}

func TestSQLiteApplicationInteractionSessionCorruptionAndClosedStoreFailClosed(t *testing.T) {
	runtime := openApplicationInteractionSQLiteRuntime(t, filepath.Join(t.TempDir(), "corruption.db"))
	service, ctx, binding, _ := applicationInteractionSessionTestFixture(t)
	repository := newSQLiteApplicationInteractionSessionRepository(runtime.DB())
	service.repository = repository
	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create SQLite corruption fixture: %#v", created)
	}
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE application_interaction_sessions
SET record_version=record_version+1,
sanitized_session_payload=json_set(sanitized_session_payload,'$.record_version',record_version+1,'$.unexpected_projection','corrupt')
WHERE session_id=?`, created.Session.SessionID); err != nil {
		t.Fatalf("corrupt SQLite application session projection: %v", err)
	}
	if result := service.Read(ctx, created.Session.SessionID); result.FailureCode != ApplicationInteractionFailureStoreContract {
		t.Fatalf("corrupt SQLite application session did not fail closed: %#v", result)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite corruption runtime: %v", err)
	}
	if _, err := repository.Read(ctx, created.Session.SessionID); !errors.Is(err, errApplicationSessionStore) {
		t.Fatalf("closed SQLite application session repository fell back: %v", err)
	}
}

func openApplicationInteractionSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations()})
	if err != nil {
		t.Fatalf("open application interaction SQLite runtime: %v", err)
	}
	return runtime
}
