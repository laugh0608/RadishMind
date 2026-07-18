package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteWorkflowRAGPromotionLifecycleRestartAndAppendOnlyContracts(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-rag-promotions.db")
	runtime := openWorkflowRAGPromotionSQLiteRuntime(t, databasePath)
	runStore := newSQLiteWorkflowRunStore(runtime.DB())
	repository, err := newWorkflowRAGPromotionRepositoryForRunStore(runStore)
	if err != nil {
		t.Fatalf("create shared SQLite promotion repository: %v", err)
	}
	sqliteRepository, ok := repository.(*sqliteWorkflowRAGPromotionRepository)
	if !ok || sqliteRepository.database != runtime.DB() {
		t.Fatalf("promotion repository did not reuse the workflow SQLite database: %T", repository)
	}

	fixture := newWorkflowRAGPromotionTestFixture(t)
	fixture.service.repository = repository
	created := fixture.createCandidate(t)
	fixture.advanceRequest("sqlite_defer")
	deferred := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionDefer, Reason: "SQLite 持久化延后复核",
	})
	fixture.advanceRequest("sqlite_approve")
	approved := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 2, Decision: workflowRAGPromotionDecisionApprove, Reason: "SQLite 持久化人工批准",
	})
	fixture.advanceRequest("sqlite_cancel")
	canceled := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 3, Decision: workflowRAGPromotionDecisionCancel, Reason: "SQLite 持久化撤销资格",
	})
	if deferred.FailureCode != "" || approved.FailureCode != "" || approved.Binding == nil || canceled.FailureCode != "" || canceled.Candidate.RecordVersion != 4 {
		t.Fatalf("SQLite promotion lifecycle failed: deferred=%#v approved=%#v canceled=%#v", deferred, approved, canceled)
	}
	candidate, decisions, binding, audits, err := repository.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || len(decisions) != 3 || binding == nil || len(audits) != 5 {
		t.Fatalf("read SQLite promotion chain: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", candidate, len(decisions), binding, len(audits), err)
	}
	assertWorkflowRAGPromotionContracts(t, candidate, decisions, *binding, audits)
	boundCandidate, exactBinding, err := repository.ReadBinding(fixture.ctx, binding.WorkflowRAGApplicationBindingRef)
	if err != nil || boundCandidate.CandidateID != candidate.CandidateID || exactBinding != *binding {
		t.Fatalf("read exact SQLite promotion binding: candidate=%#v binding=%#v err=%v", boundCandidate, exactBinding, err)
	}
	listed, err := repository.List(fixture.ctx, workflowRAGPromotionListQuery{Limit: 10})
	if err != nil || len(listed) != 1 || listed[0].CandidateID != candidate.CandidateID {
		t.Fatalf("list SQLite promotions: candidates=%#v err=%v", listed, err)
	}
	otherOwner := fixture.ctx
	otherOwner.OwnerSubjectRef = "subject_other"
	if _, _, _, _, err = repository.Read(otherOwner, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionScopeDenied) {
		t.Fatalf("cross-owner SQLite promotion read did not fail closed: %v", err)
	}
	if _, _, err = repository.ReadBinding(otherOwner, binding.WorkflowRAGApplicationBindingRef); !errors.Is(err, errWorkflowRAGPromotionNotFound) {
		t.Fatalf("cross-owner SQLite binding read did not fail closed: %v", err)
	}

	for name, statement := range map[string]string{
		"decision update": `UPDATE workflow_rag_knowledge_promotion_decisions SET decision='reject' WHERE candidate_id=?`,
		"decision delete": `DELETE FROM workflow_rag_knowledge_promotion_decisions WHERE candidate_id=?`,
		"binding update":  `UPDATE workflow_rag_application_bindings SET binding_version=1 WHERE candidate_id=?`,
		"binding delete":  `DELETE FROM workflow_rag_application_bindings WHERE candidate_id=?`,
		"audit update":    `UPDATE workflow_rag_knowledge_promotion_audits SET event_sequence=event_sequence WHERE candidate_id=?`,
		"audit delete":    `DELETE FROM workflow_rag_knowledge_promotion_audits WHERE candidate_id=?`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, mutationErr := runtime.DB().ExecContext(context.Background(), statement, candidate.CandidateID); mutationErr == nil {
				t.Fatalf("append-only SQLite resource accepted %s", name)
			}
		})
	}

	if err = runtime.Close(); err != nil {
		t.Fatalf("close first SQLite promotion runtime: %v", err)
	}
	if _, _, _, _, err = repository.Read(fixture.ctx, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionStore) {
		t.Fatalf("closed SQLite promotion store silently fell back: %v", err)
	}

	restarted := openWorkflowRAGPromotionSQLiteRuntime(t, databasePath)
	t.Cleanup(func() { _ = restarted.Close() })
	restartedRepository := newSQLiteWorkflowRAGPromotionRepository(restarted.DB())
	recovered, recoveredDecisions, recoveredBinding, recoveredAudits, err := restartedRepository.Read(fixture.ctx, candidate.CandidateID)
	if err != nil || recovered.RecordVersion != 4 || len(recoveredDecisions) != 3 || recoveredBinding == nil || len(recoveredAudits) != 5 {
		t.Fatalf("recover SQLite promotion after restart: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", recovered, len(recoveredDecisions), recoveredBinding, len(recoveredAudits), err)
	}
	assertWorkflowRAGPromotionContracts(t, recovered, recoveredDecisions, *recoveredBinding, recoveredAudits)

	if _, err = restarted.DB().ExecContext(context.Background(), `UPDATE workflow_rag_knowledge_promotion_candidates
		SET sanitized_candidate_payload='{}' WHERE candidate_id=?`, candidate.CandidateID); err != nil {
		t.Fatalf("inject SQLite promotion payload corruption: %v", err)
	}
	if _, _, _, _, err = restartedRepository.Read(fixture.ctx, candidate.CandidateID); !errors.Is(err, errWorkflowRAGPromotionStoreContract) {
		t.Fatalf("corrupted SQLite promotion payload was accepted: %v", err)
	}
}

func TestSQLiteWorkflowRAGPromotionDecisionTransactionRollsBack(t *testing.T) {
	runtime := openWorkflowRAGPromotionSQLiteRuntime(t, filepath.Join(t.TempDir(), "workflow-rag-promotion-rollback.db"))
	t.Cleanup(func() { _ = runtime.Close() })
	repository := newSQLiteWorkflowRAGPromotionRepository(runtime.DB())
	fixture := newWorkflowRAGPromotionTestFixture(t)
	fixture.service.repository = repository
	created := fixture.createCandidate(t)
	if _, err := runtime.DB().ExecContext(context.Background(), `CREATE TRIGGER workflow_rag_promotion_test_reject_decision
		BEFORE INSERT ON workflow_rag_knowledge_promotion_decisions BEGIN
			SELECT RAISE(ABORT, 'injected decision failure');
		END`); err != nil {
		t.Fatalf("install SQLite promotion failure trigger: %v", err)
	}
	fixture.advanceRequest("sqlite_rollback")
	result := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "验证 SQLite 事务整笔回滚",
	})
	if result.FailureCode != WorkflowRAGPromotionFailureStoreUnavailable {
		t.Fatalf("injected SQLite decision failure did not fail closed: %#v", result)
	}
	candidate, decisions, binding, audits, err := repository.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || candidate.RecordVersion != 1 || candidate.CandidateState != workflowRAGPromotionStatePending || len(decisions) != 0 || binding != nil || len(audits) != 1 {
		t.Fatalf("SQLite decision failure left partial state: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", candidate, len(decisions), binding, len(audits), err)
	}
}

func TestSQLiteWorkflowRAGPromotionCASAllowsOneConcurrentWriter(t *testing.T) {
	runtime := openWorkflowRAGPromotionSQLiteRuntime(t, filepath.Join(t.TempDir(), "workflow-rag-promotion-cas.db"))
	t.Cleanup(func() { _ = runtime.Close() })
	repository := newSQLiteWorkflowRAGPromotionRepository(runtime.DB())
	fixture := newWorkflowRAGPromotionTestFixture(t)
	fixture.service.repository = repository
	fixture.service.newID = newWorkflowRAGStableID
	created := fixture.createCandidate(t)
	fixture.advanceRequest("sqlite_cas")

	results := make(chan WorkflowRAGPromotionResult, 2)
	var waitGroup sync.WaitGroup
	for _, decision := range []string{workflowRAGPromotionDecisionReject, workflowRAGPromotionDecisionCancel} {
		waitGroup.Add(1)
		go func(value string) {
			defer waitGroup.Done()
			results <- fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
				ExpectedRecordVersion: 1, Decision: value, Reason: "SQLite CAS 只允许一个人工决定成功",
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
				t.Fatalf("SQLite CAS conflict omitted current version: %#v", result)
			}
			conflicts++
		default:
			t.Fatalf("unexpected SQLite CAS result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("SQLite CAS selected unexpected winners: successes=%d conflicts=%d", successes, conflicts)
	}
	_, decisions, binding, audits, err := repository.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || len(decisions) != 1 || binding != nil || len(audits) != 2 {
		t.Fatalf("SQLite CAS history is inconsistent: decisions=%d binding=%#v audits=%d err=%v", len(decisions), binding, len(audits), err)
	}
}

func openWorkflowRAGPromotionSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open workflow RAG promotion SQLite runtime: %v", err)
	}
	return runtime
}
