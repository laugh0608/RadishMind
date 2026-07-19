package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestApplicationInteractionSessionMemoryLifecycleCASAndPrivacy(t *testing.T) {
	service, ctx, binding, bridgeClient := applicationInteractionSessionTestFixture(t)
	ids := []string{"appsess_aaaaaaaaaaaaaaaa", "appturn_aaaaaaaaaaaaaaaa", "appturn_bbbbbbbbbbbbbbbb", "appturn_cccccccccccccccc"}
	service.newID = func(string) (string, error) {
		value := ids[0]
		ids = ids[1:]
		return value, nil
	}
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil || created.Session.State != applicationSessionStateActive || created.Session.RecordVersion != 1 || created.Session.ContentRetention != applicationSessionRetentionPolicy || created.Session.TurnCount != 0 {
		t.Fatalf("create application session: %#v", created)
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("session create called provider: %d", bridgeClient.callCount())
	}

	listed := service.List(ctx, ApplicationInteractionSessionListInput{ExecutionProfile: applicationInteractionProfileWorkflow, Limit: 10})
	if listed.FailureCode != "" || len(listed.Sessions) != 1 || listed.Sessions[0].SessionID != created.Session.SessionID {
		t.Fatalf("list application sessions: %#v", listed)
	}

	turnInput := ApplicationInteractionTerminalTurnInput{ExpectedSessionVersion: 1, ClientTurnKey: "client_turn_1", Status: string(WorkflowRunStatusSucceeded), InputDigest: workflowDefinitionInputDigest("private user input"), InputBytes: len("private user input"), RunRef: &ApplicationInteractionRunRef{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: workflowRunRecordDefinitionSchemaVersion}, StartedAt: now.Add(time.Second), CompletedAt: now.Add(2 * time.Second)}
	appended := service.AppendTerminalTurn(ctx, created.Session.SessionID, turnInput)
	if appended.FailureCode != "" || appended.Session == nil || appended.Turn == nil || appended.Session.RecordVersion != 2 || appended.Session.TurnCount != 1 || appended.Turn.Sequence != 1 || appended.Turn.RunRef == nil || appended.IdempotentReplay {
		t.Fatalf("append terminal application turn: %#v", appended)
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("terminal metadata append called provider: %d", bridgeClient.callCount())
	}

	replayed := service.AppendTerminalTurn(ctx, created.Session.SessionID, turnInput)
	if replayed.FailureCode != "" || !replayed.IdempotentReplay || replayed.Turn == nil || replayed.Turn.TurnID != appended.Turn.TurnID || replayed.Session == nil || replayed.Session.RecordVersion != 2 {
		t.Fatalf("idempotent terminal turn replay: %#v", replayed)
	}
	conflictingInput := turnInput
	conflictingInput.InputDigest = workflowDefinitionInputDigest("different private input")
	if conflict := service.AppendTerminalTurn(ctx, created.Session.SessionID, conflictingInput); conflict.FailureCode != ApplicationInteractionFailureIdempotencyConflict {
		t.Fatalf("client turn key accepted different content: %#v", conflict)
	}

	turns, failure := service.ListTurns(ctx, created.Session.SessionID)
	if failure != "" || len(turns) != 1 || turns[0].TurnID != appended.Turn.TurnID {
		t.Fatalf("list terminal turns: turns=%#v failure=%s", turns, failure)
	}
	closed := service.Close(ctx, created.Session.SessionID, 2)
	if closed.FailureCode != "" || closed.Session == nil || closed.Session.State != applicationSessionStateClosed || closed.Session.RecordVersion != 3 || closed.Session.ClosedAt == nil {
		t.Fatalf("close application session: %#v", closed)
	}
	if result := service.AppendTerminalTurn(ctx, created.Session.SessionID, turnInput); result.FailureCode != "" || !result.IdempotentReplay || result.Turn == nil || result.Turn.TurnID != appended.Turn.TurnID {
		t.Fatalf("closed session did not replay an existing turn: %#v", result)
	}
	newTurnAfterClose := turnInput
	newTurnAfterClose.ClientTurnKey = "client_turn_after_close"
	newTurnAfterClose.InputDigest = workflowDefinitionInputDigest("new input after close")
	newTurnAfterClose.InputBytes = len("new input after close")
	if result := service.AppendTerminalTurn(ctx, created.Session.SessionID, newTurnAfterClose); result.FailureCode != ApplicationInteractionFailureSessionClosed {
		t.Fatalf("closed session accepted a new turn: %#v", result)
	}

	payload, err := json.Marshal(struct {
		Session *ApplicationInteractionSession `json:"session"`
		Turns   []ApplicationInteractionTurn   `json:"turns"`
	}{closed.Session, turns})
	if err != nil {
		t.Fatalf("marshal application session evidence: %v", err)
	}
	for _, forbidden := range []string{"private user input", "different private input", "answer", "prompt", "Authorization", "credential", "provider raw"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("application session evidence leaked %q: %s", forbidden, payload)
		}
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("session management lifecycle called provider: %d", bridgeClient.callCount())
	}
}

func TestApplicationInteractionSessionConcurrentAppendCASAllowsOne(t *testing.T) {
	service, ctx, binding, _ := applicationInteractionSessionTestFixture(t)
	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create concurrent session fixture: %#v", created)
	}
	now := time.Now().UTC()
	results := make(chan ApplicationInteractionSessionResult, 2)
	var wait sync.WaitGroup
	for index, key := range []string{"concurrent_turn_a", "concurrent_turn_b"} {
		wait.Add(1)
		go func(index int, key string) {
			defer wait.Done()
			results <- service.AppendTerminalTurn(ctx, created.Session.SessionID, ApplicationInteractionTerminalTurnInput{ExpectedSessionVersion: 1, ClientTurnKey: key, Status: string(WorkflowRunStatusSucceeded), InputDigest: workflowDefinitionInputDigest(key), InputBytes: len(key), RunRef: &ApplicationInteractionRunRef{RunID: []string{"run_aaaaaaaaaaaaaaaa", "run_bbbbbbbbbbbbbbbb"}[index], SchemaVersion: workflowRunRecordDefinitionSchemaVersion}, StartedAt: now, CompletedAt: now.Add(time.Millisecond)})
		}(index, key)
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
			t.Fatalf("unexpected concurrent append result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent append outcomes: success=%d conflict=%d", successes, conflicts)
	}
}

func TestApplicationInteractionSessionTurnReservationPrecedesTerminalWrite(t *testing.T) {
	service, ctx, binding, bridgeClient := applicationInteractionSessionTestFixture(t)
	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create reservation fixture: %#v", created)
	}
	now := time.Now().UTC()
	reservationInput := ApplicationInteractionTurnReservationInput{ExpectedSessionVersion: 1, ClientTurnKey: "reserved_turn_1", InputDigest: workflowDefinitionInputDigest("transient reservation input"), InputBytes: len("transient reservation input"), StartedAt: now}
	reserved := service.ReserveTurn(ctx, created.Session.SessionID, reservationInput)
	if reserved.FailureCode != "" || reserved.Turn == nil || reserved.Turn.Status != string(WorkflowRunStatusRunning) || reserved.Turn.CompletedAt != nil || reserved.Turn.RunRef != nil || reserved.Session == nil || reserved.Session.RecordVersion != 2 {
		t.Fatalf("reserve turn before execution: %#v", reserved)
	}
	replayed := service.ReserveTurn(ctx, created.Session.SessionID, reservationInput)
	if replayed.FailureCode != "" || !replayed.IdempotentReplay || replayed.Turn == nil || replayed.Turn.TurnID != reserved.Turn.TurnID {
		t.Fatalf("idempotent turn reservation: %#v", replayed)
	}
	completion := ApplicationInteractionTurnCompletionInput{Status: string(WorkflowRunStatusSucceeded), RunRef: &ApplicationInteractionRunRef{RunID: "run_aaaaaaaaaaaaaaaa", SchemaVersion: workflowRunRecordDefinitionSchemaVersion}, CompletedAt: now.Add(time.Second)}
	completed := service.CompleteTurn(ctx, created.Session.SessionID, reserved.Turn.TurnID, completion)
	if completed.FailureCode != "" || completed.Turn == nil || completed.Turn.Status != string(WorkflowRunStatusSucceeded) || completed.Turn.CompletedAt == nil || completed.Turn.RunRef == nil || completed.Session == nil || completed.Session.RecordVersion != 2 {
		t.Fatalf("complete reserved turn: %#v", completed)
	}
	duplicate := service.CompleteTurn(ctx, created.Session.SessionID, reserved.Turn.TurnID, completion)
	if duplicate.FailureCode != "" || !duplicate.IdempotentReplay || duplicate.Turn == nil || duplicate.Turn.TurnID != completed.Turn.TurnID {
		t.Fatalf("idempotent terminal write: %#v", duplicate)
	}
	changed := completion
	changed.RunRef = &ApplicationInteractionRunRef{RunID: "run_bbbbbbbbbbbbbbbb", SchemaVersion: workflowRunRecordDefinitionSchemaVersion}
	if conflict := service.CompleteTurn(ctx, created.Session.SessionID, reserved.Turn.TurnID, changed); conflict.FailureCode != ApplicationInteractionFailureIdempotencyConflict {
		t.Fatalf("terminal write accepted a different run: %#v", conflict)
	}
	if bridgeClient.callCount() != 0 {
		t.Fatalf("reservation lifecycle called provider: %d", bridgeClient.callCount())
	}
}

func TestApplicationInteractionSessionStrictContractsRejectUnknownAndRawContent(t *testing.T) {
	service, ctx, binding, _ := applicationInteractionSessionTestFixture(t)
	created := service.Create(ctx, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if created.FailureCode != "" || created.Session == nil {
		t.Fatalf("create contract fixture: %#v", created)
	}
	payload, err := json.Marshal(created.Session)
	if err != nil || validateApplicationInteractionContractJSON(applicationSessionSchemaVersion, payload) != nil {
		t.Fatalf("strict session contract rejected stored session: payload=%s err=%v", payload, err)
	}
	payload = append(payload[:len(payload)-1], []byte(`,"input":"forbidden raw input"}`)...)
	if validateApplicationInteractionContractJSON(applicationSessionSchemaVersion, payload) == nil {
		t.Fatal("strict session contract accepted a raw input field")
	}
}

func applicationInteractionSessionTestFixture(t *testing.T) (applicationInteractionSessionService, ApplicationInteractionContext, ApplicationInteractionProfileBinding, *workflowExecutorTestBridge) {
	t.Helper()
	definitionService, runContext, request, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
	ctx := ApplicationInteractionContext{RequestContext: context.Background(), RequestID: "request_application_session", TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: "audit_application_session", WriteEnabled: true}
	resolver := newExactApplicationInteractionAuthorityResolver(definitionService.applications, definitionService.repository, nil, workflowRAGApplicationAuthorityResolver{})
	service := newApplicationInteractionSessionService(newMemoryApplicationInteractionSessionRepository(), resolver)
	return service, ctx, ApplicationInteractionProfileBinding{ExecutionProfile: applicationInteractionProfileWorkflow, DefinitionID: request.DefinitionID}, bridgeClient
}
