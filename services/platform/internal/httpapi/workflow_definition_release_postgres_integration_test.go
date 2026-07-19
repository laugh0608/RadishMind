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

func TestPostgresWorkflowDefinitionReleaseLifecycleRestartCASAndCorruption(t *testing.T) {
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
	if err != nil || state.MigrationID != workflowrunmigrations.MigrationID || state.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply definition release migration: %#v %v", state, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	repository := newPostgresWorkflowDefinitionReleaseRepository(runtimePool)
	releaseCtx := workflowDefinitionTestContext()
	releaseCtx.RequestContext = ctx
	releaseCtx.ApplicationID = "app_bbbbbbbbbbbbbbbb"
	applicationRepository := newMemoryApplicationCatalogRepository()
	applicationService := newApplicationCatalogService(applicationRepository)
	applicationService.newID = func() (string, error) { return releaseCtx.ApplicationID, nil }
	applicationContext := ApplicationCatalogContext{RequestContext: ctx, RequestID: "request_postgres_definition_app", TenantRef: releaseCtx.TenantRef, WorkspaceID: releaseCtx.WorkspaceID, ActorRef: releaseCtx.ActorRef, OwnerSubjectRef: releaseCtx.OwnerSubjectRef, AuditRef: "audit_postgres_definition_app", WriteEnabled: true}
	if created := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "PostgreSQL definition product", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" {
		t.Fatalf("create application fixture: %#v", created)
	}
	now := time.Date(2026, 7, 19, 17, 0, 0, 0, time.UTC)
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID, draft.ToolRefs, draft.RAGRefs, draft.RequestedCapabilities = releaseCtx.ApplicationID, []string{}, []string{}, []string{}
	candidate, err := repository.CreateCandidate(releaseCtx, "candidate-postgres", "definition-postgres", draft, now)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	var lock sync.Mutex
	successes, conflicts := 0, 0
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _, reviewErr := repository.Review(releaseCtx, candidate.CandidateID, 0, "approve", "concurrent PostgreSQL review", candidate.SourceDraftDigest, now.Add(time.Minute))
			lock.Lock()
			defer lock.Unlock()
			if reviewErr == nil {
				successes++
			} else if errors.Is(reviewErr, errWorkflowDefinitionConflict) || errors.Is(reviewErr, errWorkflowDefinitionInvalidState) {
				conflicts++
			}
		}()
	}
	wait.Wait()
	if successes != 1 || conflicts != 1 {
		t.Fatalf("review successes=%d conflicts=%d", successes, conflicts)
	}
	activation, err := repository.DecideActivation(releaseCtx, candidate.DefinitionID, 0, "activate", 1, "activate PostgreSQL version", now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	summaries := repository.ListSummaries(ReadRepositoryContext{RequestContext: ctx, TenantRef: releaseCtx.TenantRef, SubjectRef: releaseCtx.OwnerSubjectRef, AuditRef: "audit_postgres_summary"}, ListWorkflowDefinitionSummariesRequest{})
	if summaries.FailureCode != "" || len(summaries.Items) != 1 || summaries.Items[0].DefinitionStatus != workflowDefinitionActivationActive {
		t.Fatalf("PostgreSQL live summary: %#v", summaries)
	}
	runStore := newPostgresWorkflowRunStore(runtimePool)
	bridgeClient := &workflowExecutorTestBridge{}
	executor := newWorkflowExecutorService(nil, bridgeClient, runStore)
	execution := newWorkflowDefinitionExecutionService(repository, applicationRepository, executor)
	runContext := WorkflowRunContext{RequestContext: ctx, RequestID: "request_postgres_definition_run", TenantRef: releaseCtx.TenantRef, WorkspaceID: releaseCtx.WorkspaceID, ApplicationID: releaseCtx.ApplicationID, ActorRef: releaseCtx.ActorRef, AuditRef: "audit_postgres_definition_run"}
	runRequest := WorkflowDefinitionRunRequest{DefinitionID: candidate.DefinitionID, ExpectedPointerVersion: activation.PointerVersion, ExpectedDefinitionVersion: 1, ExpectedDefinitionDigest: candidate.DefinitionDigest, InputText: "private PostgreSQL continuous-chain input", ConditionValues: map[string]bool{}}
	baseline := execution.StartRun(runContext, runRequest)
	candidateRun := execution.StartRun(runContext, runRequest)
	if baseline.Record == nil || candidateRun.Record == nil || baseline.FailureCode != "" || candidateRun.FailureCode != "" {
		t.Fatalf("execute PostgreSQL v5 runs: baseline=%#v candidate=%#v", baseline, candidateRun)
	}
	comparison := executor.CompareRuns(runContext, baseline.Record.RunID, candidateRun.Record.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.RunProfile != workflowDefinitionEvaluationProfile || bridgeClient.callCount() != 2 {
		t.Fatalf("read-only PostgreSQL comparison: %#v bridge=%d", comparison, bridgeClient.callCount())
	}
	deactivated, err := repository.DecideActivation(releaseCtx, candidate.DefinitionID, activation.PointerVersion, "deactivate", 0, "stop PostgreSQL authority", now.Add(3*time.Minute))
	if err != nil || deactivated.State != workflowDefinitionActivationInactive {
		t.Fatalf("deactivate PostgreSQL definition: %#v %v", deactivated, err)
	}
	if blocked := execution.StartRun(runContext, runRequest); blocked.FailureCode != WorkflowRunFailureDefinitionAuthority || bridgeClient.callCount() != 2 {
		t.Fatalf("deactivated PostgreSQL authority reached provider: %#v bridge=%d", blocked, bridgeClient.callCount())
	}
	var storedPayload string
	if err = runtimePool.QueryRow(ctx, `SELECT sanitized_run_record FROM workflow_run_records WHERE run_id=$1`, baseline.Record.RunID).Scan(&storedPayload); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(storedPayload, runRequest.InputText) || strings.Contains(storedPayload, baseline.AdvisoryOutput) {
		t.Fatal("PostgreSQL v5 payload persisted raw input or advisory output")
	}
	if _, err = runtimePool.Exec(ctx, `UPDATE workflow_definition_activation_events SET after_pointer_version=after_pointer_version WHERE event_id=$1`, activation.Events[0].EventID); err == nil {
		t.Fatal("PostgreSQL activation event accepted UPDATE")
	}
	runtimePool.Close()
	if _, err = repository.ReadCandidate(releaseCtx, candidate.CandidateID); !errors.Is(err, errWorkflowDefinitionStore) {
		t.Fatalf("closed PostgreSQL repository fell back: %v", err)
	}
	reopened, err := workflowrunmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	restarted := newPostgresWorkflowDefinitionReleaseRepository(reopened)
	stored, err := restarted.ReadCandidate(releaseCtx, candidate.CandidateID)
	if err != nil || stored.State != workflowDefinitionStateApproved {
		t.Fatalf("restart candidate: %#v %v", stored, err)
	}
	current, err := restarted.ReadActivation(releaseCtx, candidate.DefinitionID)
	if err != nil || current.State != workflowDefinitionActivationInactive || current.ActiveVersion != 0 || current.PointerVersion != deactivated.PointerVersion || len(current.Events) != 2 {
		t.Fatalf("restart activation: %#v %v", current, err)
	}
	restoredRun, found, err := newPostgresWorkflowRunStore(reopened).ReadRun(runContext, baseline.Record.RunID)
	if err != nil || !found || restoredRun.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || restoredRun.DefinitionAuthority == nil || restoredRun.Output != "" {
		t.Fatalf("restart PostgreSQL v5 run: %#v found=%t err=%v", restoredRun, found, err)
	}
	if _, err = reopened.Exec(ctx, `UPDATE workflow_definition_release_candidates SET sanitized_candidate_payload=jsonb_set(sanitized_candidate_payload,'{definition_digest}',to_jsonb($1::text)) WHERE candidate_id=$2`, `sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, candidate.CandidateID); err != nil {
		t.Fatal(err)
	}
	if _, err = restarted.ReadCandidate(releaseCtx, candidate.CandidateID); !errors.Is(err, errWorkflowDefinitionStore) {
		t.Fatalf("corrupt PostgreSQL projection must fail closed: %v", err)
	}
}
