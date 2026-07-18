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

func TestPostgresWorkflowRAGPromotionIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	databaseContext, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	adminPool, err := workflowrunmigrations.OpenPool(databaseContext, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, databaseContext, adminPool)
	resetPostgresWorkflowRunSchema(t, databaseContext, adminPool)
	preparePostgresIntegrationRuntimeRole(t, databaseContext, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresWorkflowRunSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	migration, err := workflowrunmigrations.Apply(databaseContext, adminPool)
	if err != nil || migration.MigrationState != workflowrunmigrations.MigrationStateApplied ||
		migration.MigrationID != workflowrunmigrations.MigrationID || migration.StoreSchemaVersion != workflowrunmigrations.StoreSchemaVersion {
		t.Fatalf("apply PostgreSQL workflow RAG promotion migration: state=%#v err=%v", migration, err)
	}
	runtimePool, err := workflowrunmigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(databaseContext, `CREATE TABLE workflow_rag_promotion_runtime_must_not_create (id integer)`); ddlErr == nil {
		t.Fatal("PostgreSQL workflow runtime role acquired schema DDL permission")
	}
	repository, err := newWorkflowRAGPromotionRepositoryForRunStore(newPostgresWorkflowRunStore(runtimePool))
	if err != nil {
		t.Fatalf("create shared PostgreSQL promotion repository: %v", err)
	}
	postgresRepository, ok := repository.(*postgresWorkflowRAGPromotionRepository)
	if !ok || postgresRepository.pool != runtimePool {
		t.Fatalf("promotion repository did not reuse the workflow PostgreSQL pool: %T", repository)
	}

	fixture := newWorkflowRAGPromotionTestFixture(t)
	fixture.ctx.RequestContext = databaseContext
	fixture.service.repository = repository
	created := fixture.createCandidate(t)
	fixture.advanceRequest("postgres_defer")
	deferred := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionDefer, Reason: "PostgreSQL 持久化延后复核",
	})
	fixture.advanceRequest("postgres_approve")
	approved := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 2, Decision: workflowRAGPromotionDecisionApprove, Reason: "PostgreSQL 持久化人工批准",
	})
	fixture.advanceRequest("postgres_cancel")
	canceled := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 3, Decision: workflowRAGPromotionDecisionCancel, Reason: "PostgreSQL 持久化撤销资格",
	})
	if deferred.FailureCode != "" || approved.FailureCode != "" || approved.Binding == nil || canceled.FailureCode != "" || canceled.Candidate.RecordVersion != 4 {
		t.Fatalf("PostgreSQL promotion lifecycle failed: deferred=%#v approved=%#v canceled=%#v", deferred, approved, canceled)
	}
	candidate, decisions, binding, audits, err := repository.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || len(decisions) != 3 || binding == nil || len(audits) != 5 {
		t.Fatalf("read PostgreSQL promotion chain: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", candidate, len(decisions), binding, len(audits), err)
	}
	assertWorkflowRAGPromotionContracts(t, candidate, decisions, *binding, audits)
	listed, err := repository.List(fixture.ctx, workflowRAGPromotionListQuery{Limit: 10})
	if err != nil || len(listed) != 1 || listed[0].CandidateID != candidate.CandidateID {
		t.Fatalf("list PostgreSQL promotions: candidates=%#v err=%v", listed, err)
	}
	otherOwner := fixture.ctx
	otherOwner.OwnerSubjectRef = "subject_other"
	if _, _, _, _, err = repository.Read(otherOwner, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionScopeDenied) {
		t.Fatalf("cross-owner PostgreSQL promotion read did not fail closed: %v", err)
	}

	for name, statement := range map[string]string{
		"decision update": `UPDATE workflow_rag_knowledge_promotion_decisions SET decision='reject' WHERE candidate_id=$1`,
		"decision delete": `DELETE FROM workflow_rag_knowledge_promotion_decisions WHERE candidate_id=$1`,
		"binding update":  `UPDATE workflow_rag_application_bindings SET binding_version=1 WHERE candidate_id=$1`,
		"binding delete":  `DELETE FROM workflow_rag_application_bindings WHERE candidate_id=$1`,
		"audit update":    `UPDATE workflow_rag_knowledge_promotion_audits SET event_sequence=event_sequence WHERE candidate_id=$1`,
		"audit delete":    `DELETE FROM workflow_rag_knowledge_promotion_audits WHERE candidate_id=$1`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, mutationErr := runtimePool.Exec(databaseContext, statement, candidate.CandidateID); mutationErr == nil {
				t.Fatalf("append-only PostgreSQL resource accepted %s", name)
			}
		})
	}

	rollbackCandidate := fixture.createCandidate(t)
	if _, err = adminPool.Exec(databaseContext, `CREATE FUNCTION reject_workflow_rag_promotion_test_insert() RETURNS trigger
		LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected promotion decision failure'; END $$`); err != nil {
		t.Fatalf("create PostgreSQL promotion failure function: %v", err)
	}
	if _, err = adminPool.Exec(databaseContext, `CREATE TRIGGER workflow_rag_promotion_test_reject_decision
		BEFORE INSERT ON workflow_rag_knowledge_promotion_decisions
		FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_promotion_test_insert()`); err != nil {
		t.Fatalf("create PostgreSQL promotion failure trigger: %v", err)
	}
	fixture.advanceRequest("postgres_rollback")
	rollbackResult := fixture.service.Decide(fixture.ctx, rollbackCandidate.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "验证 PostgreSQL 事务整笔回滚",
	})
	if rollbackResult.FailureCode != WorkflowRAGPromotionFailureStoreUnavailable {
		t.Fatalf("injected PostgreSQL decision failure did not fail closed: %#v", rollbackResult)
	}
	rolledBack, rollbackDecisions, rollbackBinding, rollbackAudits, err := repository.Read(fixture.ctx, rollbackCandidate.Candidate.CandidateID)
	if err != nil || rolledBack.RecordVersion != 1 || rolledBack.CandidateState != workflowRAGPromotionStatePending || len(rollbackDecisions) != 0 || rollbackBinding != nil || len(rollbackAudits) != 1 {
		t.Fatalf("PostgreSQL decision failure left partial state: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", rolledBack, len(rollbackDecisions), rollbackBinding, len(rollbackAudits), err)
	}
	if _, err = adminPool.Exec(databaseContext, `DROP TRIGGER workflow_rag_promotion_test_reject_decision ON workflow_rag_knowledge_promotion_decisions`); err != nil {
		t.Fatalf("drop PostgreSQL promotion failure trigger: %v", err)
	}
	if _, err = adminPool.Exec(databaseContext, `DROP FUNCTION reject_workflow_rag_promotion_test_insert()`); err != nil {
		t.Fatalf("drop PostgreSQL promotion failure function: %v", err)
	}

	concurrentCandidate := fixture.createCandidate(t)
	fixture.service.newID = newWorkflowRAGStableID
	fixture.advanceRequest("postgres_cas")
	results := make(chan WorkflowRAGPromotionResult, 2)
	var waitGroup sync.WaitGroup
	for _, decision := range []string{workflowRAGPromotionDecisionReject, workflowRAGPromotionDecisionCancel} {
		waitGroup.Add(1)
		go func(value string) {
			defer waitGroup.Done()
			results <- fixture.service.Decide(fixture.ctx, concurrentCandidate.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
				ExpectedRecordVersion: 1, Decision: value, Reason: "PostgreSQL CAS 只允许一个人工决定成功",
			})
		}(decision)
	}
	waitGroup.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case WorkflowRAGPromotionFailureRecordConflict:
			if result.CurrentRecordVersion != 2 {
				t.Fatalf("PostgreSQL CAS conflict omitted current version: %#v", result)
			}
			conflicts++
		default:
			t.Fatalf("unexpected PostgreSQL CAS result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("PostgreSQL CAS selected unexpected winners: successes=%d conflicts=%d", successes, conflicts)
	}

	runtimePool.Close()
	if _, _, _, _, err = repository.Read(fixture.ctx, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionStore) {
		t.Fatalf("closed PostgreSQL promotion store silently fell back: %v", err)
	}
	reopened, err := workflowrunmigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	restartedRepository := newPostgresWorkflowRAGPromotionRepository(reopened)
	recovered, recoveredDecisions, recoveredBinding, recoveredAudits, err := restartedRepository.Read(fixture.ctx, candidate.CandidateID)
	if err != nil || recovered.RecordVersion != 4 || len(recoveredDecisions) != 3 || recoveredBinding == nil || len(recoveredAudits) != 5 {
		t.Fatalf("recover PostgreSQL promotion after restart: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", recovered, len(recoveredDecisions), recoveredBinding, len(recoveredAudits), err)
	}
	assertWorkflowRAGPromotionContracts(t, recovered, recoveredDecisions, *recoveredBinding, recoveredAudits)
	if _, err = reopened.Exec(databaseContext, `UPDATE workflow_rag_knowledge_promotion_candidates SET sanitized_candidate_payload='{}'::jsonb WHERE candidate_id=$1`, candidate.CandidateID); err != nil {
		t.Fatalf("inject PostgreSQL promotion payload corruption: %v", err)
	}
	if _, _, _, _, err = restartedRepository.Read(fixture.ctx, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionStoreContract) {
		t.Fatalf("corrupted PostgreSQL promotion payload was accepted: %v", err)
	}
}
