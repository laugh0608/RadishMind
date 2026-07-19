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
	now := time.Date(2026, 7, 19, 17, 0, 0, 0, time.UTC)
	candidate, err := repository.CreateCandidate(releaseCtx, "candidate-postgres", "definition-postgres", workflowDefinitionTestDraft(), now)
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
	if err != nil || current.ActiveVersion != 1 {
		t.Fatalf("restart activation: %#v %v", current, err)
	}
	if _, err = reopened.Exec(ctx, `UPDATE workflow_definition_release_candidates SET sanitized_candidate_payload=jsonb_set(sanitized_candidate_payload,'{definition_digest}',to_jsonb($1::text)) WHERE candidate_id=$2`, `sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, candidate.CandidateID); err != nil {
		t.Fatal(err)
	}
	if _, err = restarted.ReadCandidate(releaseCtx, candidate.CandidateID); !errors.Is(err, errWorkflowDefinitionStore) {
		t.Fatalf("corrupt PostgreSQL projection must fail closed: %v", err)
	}
}
