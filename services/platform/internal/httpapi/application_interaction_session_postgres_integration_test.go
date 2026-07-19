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

func TestPostgresApplicationInteractionSessionRestartCASPrivacyAndNoFallback(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := workflowrunmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, admin)
	resetPostgresWorkflowRunSchema(t, ctx, admin)
	preparePostgresIntegrationRuntimeRole(t, ctx, admin, runtimeUser)
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanup, admin)
		admin.Close()
	})
	state, err := workflowrunmigrations.Apply(ctx, admin)
	if err != nil || state.MigrationState != workflowrunmigrations.MigrationStateApplied || state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply application interaction session migration: %#v %v", state, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	service, interactionContext, binding, bridge := applicationInteractionSessionTestFixture(t)
	interactionContext.RequestContext = ctx
	service.repository = newPostgresApplicationInteractionSessionRepository(runtimePool)
	ids := []string{"appsess_aaaaaaaaaaaaaaaa", "appturn_aaaaaaaaaaaaaaaa", "appsess_bbbbbbbbbbbbbbbb", "appturn_bbbbbbbbbbbbbbbb", "appturn_cccccccccccccccc"}
	var idLock sync.Mutex
	service.newID = func(string) (string, error) {
		idLock.Lock()
		defer idLock.Unlock()
		value := ids[0]
		ids = ids[1:]
		return value, nil
	}
	now := time.Date(2026, 7, 19, 14, 0, 0, 321, time.UTC)
	service.now = func() time.Time { return now }
	created := service.Create(interactionContext, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create PostgreSQL application session: %#v", created)
	}
	sensitiveInput := "postgres-private-session-input-must-not-persist"
	reserved := service.ReserveTurn(interactionContext, created.Session.SessionID, ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: 1, ClientTurnKey: "postgres_turn_1", InputDigest: workflowDefinitionInputDigest(sensitiveInput), InputBytes: len(sensitiveInput), StartedAt: now.Add(time.Second)})
	if reserved.FailureCode != "" || reserved.Turn == nil || reserved.Session == nil || reserved.Session.RecordVersion != 2 {
		t.Fatalf("reserve PostgreSQL application turn: %#v", reserved)
	}
	completed := service.CompleteTurn(interactionContext, created.Session.SessionID, reserved.Turn.TurnID, ApplicationInteractionTurnCompletionInput{Status: string(WorkflowRunStatusSucceeded), RunRef: &ApplicationInteractionRunRef{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: workflowRunRecordDefinitionSchemaVersion}, CompletedAt: now.Add(2 * time.Second)})
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("complete PostgreSQL application turn: %#v", completed)
	}
	if bridge.callCount() != 0 {
		t.Fatalf("PostgreSQL session management called provider: %d", bridge.callCount())
	}
	var sessionPayload, turnPayload string
	if err = runtimePool.QueryRow(ctx, `SELECT sanitized_session_payload::text FROM application_interaction_sessions WHERE session_id=$1`, created.Session.SessionID).Scan(&sessionPayload); err != nil {
		t.Fatal(err)
	}
	if err = runtimePool.QueryRow(ctx, `SELECT sanitized_turn_payload::text FROM application_interaction_session_turns WHERE turn_id=$1`, reserved.Turn.TurnID).Scan(&turnPayload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sessionPayload+turnPayload, sensitiveInput) {
		t.Fatal("PostgreSQL application session payload persisted raw input")
	}
	if _, err = runtimePool.Exec(ctx, `DELETE FROM application_interaction_session_turns WHERE turn_id=$1`, reserved.Turn.TurnID); err == nil {
		t.Fatal("PostgreSQL application session turn accepted DELETE")
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE application_interaction_session_turns SET client_turn_key='changed' WHERE turn_id=$1`, reserved.Turn.TurnID); err == nil {
		t.Fatal("PostgreSQL application session turn accepted identity mutation")
	}

	casCreated := service.Create(interactionContext, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if casCreated.FailureCode != "" || casCreated.Session == nil {
		t.Fatalf("create PostgreSQL CAS session: %#v", casCreated)
	}
	results := make(chan ApplicationInteractionSessionResult, 2)
	var wait sync.WaitGroup
	for _, key := range []string{"postgres_cas_a", "postgres_cas_b"} {
		wait.Add(1)
		go func(clientKey string) {
			defer wait.Done()
			results <- service.ReserveTurn(interactionContext, casCreated.Session.SessionID, ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: 1, ClientTurnKey: clientKey, InputDigest: workflowDefinitionInputDigest(clientKey), InputBytes: len(clientKey), StartedAt: now.Add(3 * time.Second)})
		}(key)
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		if result.FailureCode == "" {
			successes++
		} else if result.FailureCode == ApplicationInteractionFailureVersionConflict {
			conflicts++
		} else {
			t.Fatalf("unexpected PostgreSQL CAS result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("PostgreSQL CAS outcomes: successes=%d conflicts=%d", successes, conflicts)
	}

	runtimePool.Close()
	if _, err = service.repository.Read(interactionContext, created.Session.SessionID); !errors.Is(err, errApplicationSessionStore) {
		t.Fatalf("closed PostgreSQL session repository fell back: %v", err)
	}
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restarted := newPostgresApplicationInteractionSessionRepository(reopened)
	stored, err := restarted.Read(interactionContext, created.Session.SessionID)
	turns, turnsErr := restarted.ListTurns(interactionContext, created.Session.SessionID)
	if err != nil || turnsErr != nil || stored.RecordVersion != 2 || len(turns) != 1 || turns[0].Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("restart PostgreSQL application session: session=%#v turns=%#v err=%v turnsErr=%v", stored, turns, err, turnsErr)
	}
}
