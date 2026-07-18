package httpapi

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type workflowRAGPromotionTestFixture struct {
	ctx             WorkflowRAGPromotionContext
	runStore        *memoryWorkflowRunStore
	snapshots       *memoryWorkflowRAGSnapshotRepository
	evaluations     *memoryWorkflowRAGEvaluationDatasetRepository
	drafts          *memoryApplicationConfigurationDraftRepository
	promotions      *memoryWorkflowRAGPromotionRepository
	service         workflowRAGPromotionService
	application     ApplicationSummary
	applicationErr  error
	profile         WorkflowRAGExecutionProfile
	now             time.Time
	baseline        WorkflowRAGSnapshotRecord
	candidate       WorkflowRAGSnapshotRecord
	datasetResource WorkflowRAGEvaluationDatasetResource
	datasetVersion  WorkflowRAGEvaluationDatasetVersion
	review          WorkflowRAGCandidateSnapshotReview
	draft           ApplicationConfigurationDraft
	createInput     WorkflowRAGPromotionCreateInput
}

func TestWorkflowRAGPromotionLifecycleContractsAndMetadataOnlyProjection(t *testing.T) {
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)
	if created.Candidate.CandidateState != workflowRAGPromotionStatePending || created.Candidate.RecordVersion != 1 ||
		created.Candidate.BindingRef != nil || created.Eligibility.Eligible {
		t.Fatalf("unexpected pending candidate: %#v", created)
	}
	if created.Candidate.Evidence.Dataset.DatasetID != fixture.datasetResource.DatasetID ||
		created.Candidate.Evidence.Dataset.DatasetVersion != 1 || created.Candidate.Evidence.Dataset.DatasetDigest != fixture.datasetVersion.Dataset.DatasetDigest ||
		created.Candidate.Evidence.CandidateReviewID != fixture.review.ReviewID ||
		created.Candidate.Evidence.BaselineSnapshot != workflowRAGEvaluationSnapshotBinding(fixture.baseline) ||
		created.Candidate.Evidence.CandidateSnapshot != workflowRAGEvaluationSnapshotBinding(fixture.candidate) ||
		created.Candidate.Evidence.Profile.ProfileVersion != fixture.profile.ProfileVersion ||
		created.Candidate.Evidence.Profile.ProfileDigest != fixture.profile.ProfileDigest {
		t.Fatalf("candidate did not preserve exact server-side evidence: %#v", created.Candidate.Evidence)
	}
	if !workflowRAGDigestPattern.MatchString(created.Candidate.CandidateDigest) || !workflowRAGDigestPattern.MatchString(created.Candidate.Evidence.SourceDraft.DraftDigest) {
		t.Fatalf("candidate digests are invalid: %#v", created.Candidate)
	}

	fixture.advanceRequest("defer")
	deferred := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionDefer, Reason: "等待维护窗口复核",
	})
	if deferred.FailureCode != "" || deferred.Candidate.CandidateState != workflowRAGPromotionStateDeferred || deferred.Candidate.RecordVersion != 2 || deferred.Binding != nil {
		t.Fatalf("defer decision failed: %#v", deferred)
	}

	fixture.advanceRequest("approve")
	approved := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 2, Decision: workflowRAGPromotionDecisionApprove, Reason: "权威证据已由人工确认",
	})
	if approved.FailureCode != "" || approved.Candidate.CandidateState != workflowRAGPromotionStateApproved || approved.Candidate.RecordVersion != 3 ||
		approved.Binding == nil || approved.Candidate.BindingRef == nil || *approved.Candidate.BindingRef != approved.Binding.WorkflowRAGApplicationBindingRef ||
		!approved.Eligibility.Eligible || approved.Eligibility.Status != "eligible" || len(approved.Decisions) != 2 {
		t.Fatalf("approve did not atomically issue an eligible immutable binding: %#v", approved)
	}
	if approved.Binding.Evidence != approved.Candidate.Evidence || approved.Binding.CandidateDigest != approved.Candidate.CandidateDigest ||
		approved.Binding.ApprovedDecisionID != approved.Decisions[1].DecisionID || approved.Binding.ApprovedRecordVersion != 3 {
		t.Fatalf("binding does not match the exact approval evidence: %#v", approved.Binding)
	}

	fixture.advanceRequest("cancel")
	canceled := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 3, Decision: workflowRAGPromotionDecisionCancel, Reason: "撤销当前配置绑定资格",
	})
	if canceled.FailureCode != "" || canceled.Candidate.CandidateState != workflowRAGPromotionStateCanceled || canceled.Candidate.RecordVersion != 4 ||
		canceled.Binding == nil || canceled.Candidate.BindingRef == nil || canceled.Eligibility.Eligible || len(canceled.Decisions) != 3 {
		t.Fatalf("cancel did not retain immutable history while revoking eligibility: %#v", canceled)
	}
	invalid := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 4, Decision: workflowRAGPromotionDecisionApprove, Reason: "不得重复批准终态候选",
	})
	if invalid.FailureCode != WorkflowRAGPromotionFailureTransitionInvalid || invalid.CurrentRecordVersion != 4 || invalid.CurrentState != workflowRAGPromotionStateCanceled {
		t.Fatalf("terminal transition did not fail closed: %#v", invalid)
	}

	candidate, decisions, binding, audits, err := fixture.promotions.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || len(decisions) != 3 || binding == nil || len(audits) != 5 {
		t.Fatalf("append-only promotion history drifted: decisions=%d audits=%d binding=%#v err=%v", len(decisions), len(audits), binding, err)
	}
	assertWorkflowRAGPromotionContracts(t, candidate, decisions, *binding, audits)

	listed := fixture.service.List(fixture.ctx, WorkflowRAGPromotionListInput{})
	read := fixture.service.Read(fixture.ctx, candidate.CandidateID)
	if listed.FailureCode != "" || len(listed.Items) != 1 || read.FailureCode != "" {
		t.Fatalf("promotion list/detail failed: list=%#v read=%#v", listed, read)
	}
	metadata, _ := json.Marshal(struct {
		List WorkflowRAGPromotionListResult
		Read WorkflowRAGPromotionResult
	}{listed, read})
	for _, forbidden := range []string{"promotion authority query", "promotion authority fragment", "promotion review note", "Authorization:", "prompt"} {
		if strings.Contains(string(metadata), forbidden) {
			t.Fatalf("promotion projection leaked forbidden body %q: %s", forbidden, metadata)
		}
	}
}

func TestWorkflowRAGPromotionApproveFailsClosedOnAuthorityDriftAndArchive(t *testing.T) {
	tests := []struct {
		name    string
		failure string
		mutate  func(*testing.T, *workflowRAGPromotionTestFixture)
	}{
		{"dataset version", WorkflowRAGPromotionFailureDatasetChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.versionDataset(t) }},
		{"dataset archive", WorkflowRAGPromotionFailureDatasetArchived, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.archiveDataset(t) }},
		{"baseline snapshot version", WorkflowRAGPromotionFailureSnapshotChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.versionSnapshot(t, f.baseline) }},
		{"candidate snapshot version", WorkflowRAGPromotionFailureSnapshotChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.versionSnapshot(t, f.candidate) }},
		{"baseline snapshot archive", WorkflowRAGPromotionFailureSnapshotArchived, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.archiveSnapshot(t, f.baseline) }},
		{"candidate snapshot archive", WorkflowRAGPromotionFailureSnapshotArchived, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.archiveSnapshot(t, f.candidate) }},
		{"review contract", WorkflowRAGPromotionFailureReviewInvalid, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.corruptReview(t) }},
		{"lexical profile", WorkflowRAGPromotionFailureProfileChanged, func(_ *testing.T, f *workflowRAGPromotionTestFixture) { f.profile.ProfileVersion++ }},
		{"draft version", WorkflowRAGPromotionFailureDraftChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) { f.versionDraft(t) }},
		{"draft digest", WorkflowRAGPromotionFailureDraftChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) {
			f.mutateDraft(t, func(draft *ApplicationConfigurationDraft) { draft.DisplayName = "Changed promotion draft" })
		}},
		{"draft validation", WorkflowRAGPromotionFailureDraftInvalid, func(t *testing.T, f *workflowRAGPromotionTestFixture) {
			f.mutateDraft(t, func(draft *ApplicationConfigurationDraft) {
				draft.ValidationSummary = invalidApplicationDraftValidation("display_name", "invalid")
			})
		}},
		{"draft base revision", WorkflowRAGPromotionFailureDraftChanged, func(t *testing.T, f *workflowRAGPromotionTestFixture) {
			f.mutateDraft(t, func(draft *ApplicationConfigurationDraft) { draft.BaseApplicationUpdatedAt = "2026-05-31T10:21:00Z" })
		}},
		{"application archive", WorkflowRAGPromotionFailureApplicationArchived, func(_ *testing.T, f *workflowRAGPromotionTestFixture) {
			f.applicationErr = errApplicationCatalogArchived
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newWorkflowRAGPromotionTestFixture(t)
			created := fixture.createCandidate(t)
			test.mutate(t, fixture)
			fixture.advanceRequest("drift_approve")
			result := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
				ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "尝试批准已漂移的权威证据",
			})
			if result.FailureCode != test.failure || result.CurrentRecordVersion != 1 || result.CurrentState != workflowRAGPromotionStatePending {
				t.Fatalf("expected %s, got %#v", test.failure, result)
			}
			candidate, decisions, binding, audits, err := fixture.promotions.Read(fixture.ctx, created.Candidate.CandidateID)
			if err != nil || candidate.RecordVersion != 1 || candidate.CandidateState != workflowRAGPromotionStatePending || len(decisions) != 0 || binding != nil || len(audits) != 1 {
				t.Fatalf("failed approval left partial state: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", candidate, len(decisions), binding, len(audits), err)
			}
		})
	}
}

func TestWorkflowRAGPromotionNonAuthorizingDecisionCanCloseDriftedCandidate(t *testing.T) {
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)
	fixture.archiveDataset(t)
	fixture.advanceRequest("reject_drifted")
	result := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionReject, Reason: "来源已漂移，人工拒绝候选",
	})
	if result.FailureCode != "" || result.Candidate.CandidateState != workflowRAGPromotionStateRejected || result.Binding != nil || result.Eligibility.Eligible {
		t.Fatalf("safe rejection was blocked by source drift: %#v", result)
	}
}

func TestWorkflowRAGPromotionConcurrentCASAllowsSingleDecision(t *testing.T) {
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)
	fixture.service.newID = newWorkflowRAGStableID
	fixture.advanceRequest("concurrent")

	results := make(chan WorkflowRAGPromotionResult, 2)
	var wait sync.WaitGroup
	for _, decision := range []string{workflowRAGPromotionDecisionReject, workflowRAGPromotionDecisionCancel} {
		wait.Add(1)
		go func(value string) {
			defer wait.Done()
			results <- fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
				ExpectedRecordVersion: 1, Decision: value, Reason: "并发人工决定只允许一个成功",
			})
		}(decision)
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		if result.FailureCode == "" {
			successes++
		} else if result.FailureCode == WorkflowRAGPromotionFailureRecordConflict && result.CurrentRecordVersion == 2 {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS did not select one winner: successes=%d conflicts=%d", successes, conflicts)
	}
	_, decisions, binding, audits, err := fixture.promotions.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || len(decisions) != 1 || binding != nil || len(audits) != 2 {
		t.Fatalf("concurrent CAS history is not append-only: decisions=%d binding=%#v audits=%d err=%v", len(decisions), binding, len(audits), err)
	}
}

func TestWorkflowRAGPromotionScopeSecretRollbackAndNoFallback(t *testing.T) {
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)

	other := fixture.ctx
	other.OwnerSubjectRef, other.ActorRef = "subject_other", "subject_other"
	if result := fixture.service.Read(other, created.Candidate.CandidateID); result.FailureCode != WorkflowRAGPromotionFailureScopeDenied {
		t.Fatalf("cross-owner read was accepted: %#v", result)
	}
	disabled := fixture.ctx
	disabled.WriteEnabled = false
	if result := fixture.service.Decide(disabled, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionReject, Reason: "写入门禁必须生效"}); result.FailureCode != WorkflowRAGPromotionFailureWriteDisabled {
		t.Fatalf("disabled write was accepted: %#v", result)
	}
	if result := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "Authorization: Bearer forbidden"}); result.FailureCode != WorkflowRAGPromotionFailureSecretForbidden {
		t.Fatalf("secret-bearing decision was accepted: %#v", result)
	}
	secretCreate := fixture.createInput
	secretCreate.DraftID = "sk-forbidden-draft"
	if result := fixture.service.Create(fixture.ctx, secretCreate); result.FailureCode != WorkflowRAGPromotionFailureSecretForbidden || result.Candidate != nil {
		t.Fatalf("secret-bearing candidate reference was accepted: %#v", result)
	}
	tainted := *created.Candidate
	tainted.UpdatedByActorRef = "sk-forbidden-actor"
	taintedPayload, err := json.Marshal(tainted)
	if err != nil {
		t.Fatalf("marshal secret-bearing stored candidate: %v", err)
	}
	if _, err = decodeWorkflowRAGPromotionCandidate(taintedPayload); err == nil {
		t.Fatal("secret-bearing stored candidate was accepted")
	}

	failing := fixture.service
	failing.repository = workflowRAGPromotionFailingAppendRepository{workflowRAGPromotionRepository: fixture.promotions}
	fixture.advanceRequest("rollback")
	result := failing.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "验证原子回滚不留下部分资源",
	})
	if result.FailureCode != WorkflowRAGPromotionFailureStoreUnavailable {
		t.Fatalf("injected append failure did not fail closed: %#v", result)
	}
	candidate, decisions, binding, audits, err := fixture.promotions.Read(fixture.ctx, created.Candidate.CandidateID)
	if err != nil || candidate.RecordVersion != 1 || len(decisions) != 0 || binding != nil || len(audits) != 1 {
		t.Fatalf("append failure left partial writes: candidate=%#v decisions=%d binding=%#v audits=%d err=%v", candidate, len(decisions), binding, len(audits), err)
	}

	if repository, err := newWorkflowRAGPromotionRepositoryForRunStore(&sqliteWorkflowRunStore{}); err == nil || repository != nil {
		t.Fatalf("batch A silently fell back from SQLite: repository=%T err=%v", repository, err)
	}
	if repository, err := newWorkflowRAGPromotionRepositoryForRunStore(&postgresWorkflowRunStore{}); err == nil || repository != nil {
		t.Fatalf("batch A silently fell back from PostgreSQL: repository=%T err=%v", repository, err)
	}
}

func TestWorkflowRAGPromotionApprovalDoesNotMutateSourcesOrExecuteWorkflow(t *testing.T) {
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)
	snapshotContext := workflowRAGPromotionSnapshotContext(fixture.ctx)
	draftContext := workflowRAGPromotionDraftContext(fixture.ctx)
	beforeBaselineResource, beforeBaseline, _ := fixture.snapshots.ReadVersion(snapshotContext, fixture.baseline.SnapshotID, 1)
	beforeCandidateResource, beforeCandidate, _ := fixture.snapshots.ReadVersion(snapshotContext, fixture.candidate.SnapshotID, 1)
	beforeDatasetResource, beforeDataset, _ := fixture.evaluations.ReadVersion(snapshotContext, fixture.datasetResource.DatasetID, 1)
	beforeDraft, _ := fixture.drafts.Read(draftContext, fixture.draft.DraftID)
	beforeApplication := fixture.application

	fixture.advanceRequest("no_side_effects")
	approved := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "只签发配置绑定资格",
	})
	if approved.FailureCode != "" || approved.Binding == nil {
		t.Fatalf("approve failed: %#v", approved)
	}
	afterBaselineResource, afterBaseline, _ := fixture.snapshots.ReadVersion(snapshotContext, fixture.baseline.SnapshotID, 1)
	afterCandidateResource, afterCandidate, _ := fixture.snapshots.ReadVersion(snapshotContext, fixture.candidate.SnapshotID, 1)
	afterDatasetResource, afterDataset, _ := fixture.evaluations.ReadVersion(snapshotContext, fixture.datasetResource.DatasetID, 1)
	afterDraft, _ := fixture.drafts.Read(draftContext, fixture.draft.DraftID)
	if !reflect.DeepEqual(beforeBaselineResource, afterBaselineResource) || !reflect.DeepEqual(beforeBaseline, afterBaseline) ||
		!reflect.DeepEqual(beforeCandidateResource, afterCandidateResource) || !reflect.DeepEqual(beforeCandidate, afterCandidate) ||
		!reflect.DeepEqual(beforeDatasetResource, afterDatasetResource) || !reflect.DeepEqual(beforeDataset, afterDataset) ||
		!reflect.DeepEqual(beforeDraft, afterDraft) || !reflect.DeepEqual(beforeApplication, fixture.application) {
		t.Fatal("promotion approval mutated an authoritative source")
	}
	fixture.runStore.mu.RLock()
	defer fixture.runStore.mu.RUnlock()
	if len(fixture.runStore.records) != 0 {
		t.Fatalf("promotion approval created %d workflow runs", len(fixture.runStore.records))
	}
}

type workflowRAGPromotionFailingAppendRepository struct {
	workflowRAGPromotionRepository
}

func (workflowRAGPromotionFailingAppendRepository) AppendDecision(
	WorkflowRAGPromotionContext,
	string,
	int,
	WorkflowRAGKnowledgePromotionCandidate,
	WorkflowRAGKnowledgePromotionDecision,
	*WorkflowRAGApplicationBinding,
	[]WorkflowRAGPromotionAudit,
) error {
	return errWorkflowRAGPromotionStore
}

func newWorkflowRAGPromotionTestFixture(t *testing.T) *workflowRAGPromotionTestFixture {
	t.Helper()
	fixture := &workflowRAGPromotionTestFixture{
		runStore: newMemoryWorkflowRunStore(defaultWorkflowRunStoreCapacity),
		profile:  workflowRAGLexicalProfile(),
		now:      time.Date(2026, 7, 18, 9, 0, 0, 0, time.UTC),
	}
	base := workflowRAGTestContext()
	fixture.ctx = WorkflowRAGPromotionContext{
		RequestContext: base.RequestContext, RequestID: base.RequestID, TenantRef: base.TenantRef,
		WorkspaceID: base.WorkspaceID, ApplicationID: base.ApplicationID, ActorRef: base.ActorRef,
		OwnerSubjectRef: base.ActorRef, AuditRef: base.AuditRef, WriteEnabled: true,
	}
	fixture.application = ApplicationSummary{
		ApplicationRef: fixture.ctx.ApplicationID, TenantRef: fixture.ctx.TenantRef, ApplicationKind: "workflow_copilot",
		DisplayName: "RadishFlow Copilot", OwnerSubjectRef: fixture.ctx.OwnerSubjectRef, UpdatedAt: "2026-05-31T10:20:00Z",
	}

	snapshotRepository, err := newWorkflowRAGSnapshotRepositoryForRunStore(fixture.runStore)
	if err != nil {
		t.Fatalf("create snapshot repository: %v", err)
	}
	fixture.snapshots = snapshotRepository.(*memoryWorkflowRAGSnapshotRepository)
	fixture.evaluations = newWorkflowRAGEvaluationDatasetRepositoryForRunStore(fixture.runStore).(*memoryWorkflowRAGEvaluationDatasetRepository)
	promotionRepository, err := newWorkflowRAGPromotionRepositoryForRunStore(fixture.runStore)
	if err != nil {
		t.Fatalf("create promotion repository: %v", err)
	}
	fixture.promotions = promotionRepository.(*memoryWorkflowRAGPromotionRepository)
	if fixture.snapshots.ownerLock != &fixture.runStore.mu || fixture.evaluations.mu != &fixture.runStore.mu || fixture.promotions.ownerLock != &fixture.runStore.mu {
		t.Fatal("workflow memory repositories do not share the workflow owner lock")
	}

	snapshotContext := workflowRAGPromotionSnapshotContext(fixture.ctx)
	fragments := []WorkflowRAGFragmentInput{{
		FragmentRef: "promotion_authority", SourceType: "manual", SourceRef: "manual.promotion", PageSlug: "promotion/authority",
		Title: "Promotion authority", IsOfficial: true, Content: "promotion authority fragment",
	}}
	fixture.baseline = createWorkflowRAGEvaluationTestSnapshot(t, fixture.snapshots, snapshotContext, "rags_aaaaaaaaaaaaaaaa", "promotion_baseline", "public", fragments)
	fixture.candidate = createWorkflowRAGEvaluationTestSnapshot(t, fixture.snapshots, snapshotContext, "rags_bbbbbbbbbbbbbbbb", "promotion_candidate", "public", fragments)

	evaluationService := newWorkflowRAGEvaluationDatasetService(fixture.evaluations, fixture.snapshots)
	evaluationService.now = func() time.Time { return fixture.now }
	evaluationService.newID = workflowRAGEvaluationTestIDGenerator()
	created := evaluationService.Create(snapshotContext, WorkflowRAGEvaluationDatasetCreateInput{
		DatasetKey: "promotion_authority", DisplayName: "Promotion authority", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(fixture.baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "Metadata-only reviewed promotion evidence.", Samples: []WorkflowRAGEvaluationSample{{
			SampleID: "promotion_authority", QueryText: "promotion authority query", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"promotion_authority"}, RequiredOfficialRefs: []string{"promotion_authority"}, TopK: 1,
			ReviewNote: "promotion review note",
		}},
	})
	if created.FailureCode != "" || created.Resource == nil || created.Version == nil {
		t.Fatalf("create promotion dataset: %#v", created)
	}
	fixture.datasetResource, fixture.datasetVersion = *created.Resource, *created.Version
	fixture.now = fixture.now.Add(time.Minute)
	reviewed := evaluationService.CreateCandidateReview(snapshotContext, fixture.datasetResource.DatasetID, WorkflowRAGCandidateReviewInput{
		DatasetVersion: 1, DatasetDigest: fixture.datasetVersion.Dataset.DatasetDigest,
		CandidateSnapshot: workflowRAGEvaluationSnapshotBinding(fixture.candidate),
	})
	if reviewed.FailureCode != "" || reviewed.Review == nil || reviewed.Review.Candidate.Status != "passed" || reviewed.Review.Conclusion != "unchanged" {
		t.Fatalf("create eligible promotion review: %#v", reviewed)
	}
	fixture.review = *reviewed.Review
	reviewContext := snapshotContext
	reviewContext.ActorRef, reviewContext.RequestID, reviewContext.AuditRef = fixture.review.CreatedByActorRef, fixture.review.RequestID, fixture.review.AuditRef
	if err := validateStoredWorkflowRAGCandidateReview(fixture.review, reviewContext); err != nil {
		t.Fatalf("promotion review fixture contract: %v", err)
	}
	if fixture.review.Dataset != fixture.datasetVersion.Dataset.DatasetBinding() || fixture.review.BaselineSnapshot != fixture.datasetVersion.Dataset.Snapshot || fixture.review.Profile != fixture.datasetVersion.Dataset.Profile {
		t.Fatalf("promotion review fixture binding mismatch: review=%#v dataset=%#v", fixture.review, fixture.datasetVersion.Dataset)
	}
	storedResource, storedVersion, storedErr := fixture.evaluations.ReadVersion(snapshotContext, fixture.datasetResource.DatasetID, 1)
	storedReview, storedReviewErr := fixture.evaluations.ReadReview(snapshotContext, fixture.datasetResource.DatasetID, fixture.review.ReviewID)
	storedReviewContext := snapshotContext
	storedReviewContext.ActorRef, storedReviewContext.RequestID, storedReviewContext.AuditRef = storedReview.CreatedByActorRef, storedReview.RequestID, storedReview.AuditRef
	if storedErr != nil || storedReviewErr != nil || validateStoredWorkflowRAGCandidateReview(storedReview, storedReviewContext) != nil ||
		storedReview.Dataset != storedVersion.Dataset.DatasetBinding() || storedReview.BaselineSnapshot != storedVersion.Dataset.Snapshot || storedReview.Profile != storedVersion.Dataset.Profile || storedResource.LatestDigest != storedVersion.Dataset.DatasetDigest {
		t.Fatalf("stored promotion review fixture mismatch: resource=%#v version=%#v review=%#v version_err=%v review_err=%v", storedResource, storedVersion, storedReview, storedErr, storedReviewErr)
	}

	fixture.drafts = newMemoryApplicationConfigurationDraftRepository()
	draftService := newApplicationConfigurationDraftService(fixture.drafts)
	draftService.now = func() time.Time { return fixture.now }
	payload := validApplicationDraftPayload()
	payload.WorkspaceID, payload.ApplicationID = fixture.ctx.WorkspaceID, fixture.ctx.ApplicationID
	payload.BaseApplicationUpdatedAt = fixture.application.UpdatedAt
	draftContext := workflowRAGPromotionDraftContext(fixture.ctx)
	draftContext.WriteEnabled = true
	draftResult := draftService.Save(draftContext, payload, 0)
	if draftResult.FailureCode != "" || draftResult.Draft == nil {
		t.Fatalf("create promotion source draft: %#v", draftResult)
	}
	fixture.draft = *draftResult.Draft
	fixture.createInput = WorkflowRAGPromotionCreateInput{
		DatasetID: fixture.datasetResource.DatasetID, DatasetVersion: 1, DatasetDigest: fixture.datasetVersion.Dataset.DatasetDigest,
		CandidateReviewID: fixture.review.ReviewID, DraftID: fixture.draft.DraftID, ExpectedDraftVersion: 1,
	}
	fixture.service = newWorkflowRAGPromotionService(fixture.promotions, fixture.evaluations, fixture.snapshots, fixture.drafts, func(WorkflowRAGPromotionContext) (ApplicationSummary, error) {
		if fixture.applicationErr != nil {
			return ApplicationSummary{}, fixture.applicationErr
		}
		return fixture.application, nil
	})
	fixture.service.currentProfile = func() WorkflowRAGExecutionProfile { return fixture.profile }
	fixture.service.now = func() time.Time { return fixture.now }
	fixture.service.newID = workflowRAGEvaluationTestIDGenerator()
	if _, failure := fixture.service.loadPromotionEvidence(fixture.ctx, fixture.createInput); failure != "" {
		_, version, _ := fixture.evaluations.ReadVersion(snapshotContext, fixture.createInput.DatasetID, fixture.createInput.DatasetVersion)
		review, reviewErr := fixture.evaluations.ReadReview(snapshotContext, fixture.createInput.DatasetID, fixture.createInput.CandidateReviewID)
		storedContext := snapshotContext
		storedContext.ActorRef, storedContext.RequestID, storedContext.AuditRef = review.CreatedByActorRef, review.RequestID, review.AuditRef
		t.Fatalf("promotion authority fixture reload failed: failure=%s review_err=%v contract_err=%v dataset_equal=%t baseline_equal=%t profile_equal=%t input=%#v review=%#v", failure, reviewErr,
			validateStoredWorkflowRAGCandidateReview(review, storedContext), review.Dataset == version.Dataset.DatasetBinding(),
			review.BaselineSnapshot == version.Dataset.Snapshot, review.Profile == version.Dataset.Profile, fixture.createInput, review)
	}
	return fixture
}

func (fixture *workflowRAGPromotionTestFixture) createCandidate(t *testing.T) WorkflowRAGPromotionResult {
	t.Helper()
	result := fixture.service.Create(fixture.ctx, fixture.createInput)
	if result.FailureCode != "" || result.Candidate == nil {
		t.Fatalf("create promotion candidate: %#v", result)
	}
	return result
}

func (fixture *workflowRAGPromotionTestFixture) advanceRequest(suffix string) {
	fixture.now = fixture.now.Add(time.Minute)
	fixture.ctx.RequestID = "request_promotion_" + suffix
	fixture.ctx.AuditRef = "audit_promotion_" + suffix
}

func (fixture *workflowRAGPromotionTestFixture) versionDataset(t *testing.T) {
	t.Helper()
	service := newWorkflowRAGEvaluationDatasetService(fixture.evaluations, fixture.snapshots)
	service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	result := service.Version(workflowRAGPromotionSnapshotContext(fixture.ctx), fixture.datasetResource.DatasetID, WorkflowRAGEvaluationDatasetVersionInput{
		ExpectedLatestVersion: 1, DisplayName: fixture.datasetVersion.DisplayName + " v2", ContentClassification: fixture.datasetVersion.Dataset.ContentClassification,
		BaselineSnapshot: fixture.datasetVersion.Dataset.Snapshot, Thresholds: fixture.datasetVersion.Dataset.Thresholds,
		ReviewSummary: "Versioned promotion evidence.", Samples: fixture.datasetVersion.Dataset.Samples,
	})
	if result.FailureCode != "" {
		t.Fatalf("version promotion dataset: %#v", result)
	}
}

func (fixture *workflowRAGPromotionTestFixture) archiveDataset(t *testing.T) {
	t.Helper()
	service := newWorkflowRAGEvaluationDatasetService(fixture.evaluations, fixture.snapshots)
	service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	result := service.Archive(workflowRAGPromotionSnapshotContext(fixture.ctx), fixture.datasetResource.DatasetID, 1)
	if result.FailureCode != "" {
		t.Fatalf("archive promotion dataset: %#v", result)
	}
}

func (fixture *workflowRAGPromotionTestFixture) versionSnapshot(t *testing.T, record WorkflowRAGSnapshotRecord) {
	t.Helper()
	service := newWorkflowRAGSnapshotService(fixture.snapshots)
	service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	inputs := make([]WorkflowRAGFragmentInput, 0, len(record.Fragments))
	for _, fragment := range record.Fragments {
		inputs = append(inputs, WorkflowRAGFragmentInput{
			FragmentRef: fragment.FragmentRef, SourceType: fragment.SourceType, SourceRef: fragment.SourceRef, PageSlug: fragment.PageSlug,
			Title: fragment.Title, IsOfficial: fragment.IsOfficial, Content: fragment.Content,
		})
	}
	result := service.Version(workflowRAGPromotionSnapshotContext(fixture.ctx), record.SnapshotID, WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: 1, DisplayName: record.DisplayName + " v2", ContentClassification: record.ContentClassification, Fragments: inputs,
	})
	if result.FailureCode != "" {
		t.Fatalf("version promotion snapshot: %#v", result)
	}
}

func (fixture *workflowRAGPromotionTestFixture) archiveSnapshot(t *testing.T, record WorkflowRAGSnapshotRecord) {
	t.Helper()
	service := newWorkflowRAGSnapshotService(fixture.snapshots)
	service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	result := service.Archive(workflowRAGPromotionSnapshotContext(fixture.ctx), record.SnapshotID, 1)
	if result.FailureCode != "" {
		t.Fatalf("archive promotion snapshot: %#v", result)
	}
}

func (fixture *workflowRAGPromotionTestFixture) corruptReview(t *testing.T) {
	t.Helper()
	ctx := workflowRAGPromotionSnapshotContext(fixture.ctx)
	fixture.runStore.mu.Lock()
	defer fixture.runStore.mu.Unlock()
	key := workflowRAGEvaluationStoreKey(ctx, fixture.datasetResource.DatasetID)
	entry := fixture.evaluations.entries[key]
	review := entry.reviews[fixture.review.ReviewID]
	review.Candidate.Status = "corrupted"
	entry.reviews[fixture.review.ReviewID] = review
	fixture.evaluations.entries[key] = entry
}

func (fixture *workflowRAGPromotionTestFixture) versionDraft(t *testing.T) {
	t.Helper()
	service := newApplicationConfigurationDraftService(fixture.drafts)
	service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	ctx := workflowRAGPromotionDraftContext(fixture.ctx)
	ctx.WriteEnabled = true
	result := service.Save(ctx, fixture.draft.ApplicationConfigurationDraftPayload, 1)
	if result.FailureCode != "" {
		t.Fatalf("version promotion draft: %#v", result)
	}
}

func (fixture *workflowRAGPromotionTestFixture) mutateDraft(t *testing.T, mutate func(*ApplicationConfigurationDraft)) {
	t.Helper()
	ctx := workflowRAGPromotionDraftContext(fixture.ctx)
	fixture.drafts.mu.Lock()
	defer fixture.drafts.mu.Unlock()
	key := applicationConfigurationDraftRepositoryKey(ctx, fixture.draft.DraftID)
	draft, found := fixture.drafts.drafts[key]
	if !found {
		t.Fatal("promotion draft fixture is missing")
	}
	mutate(&draft)
	fixture.drafts.drafts[key] = draft
}

func assertWorkflowRAGPromotionContracts(
	t *testing.T,
	candidate WorkflowRAGKnowledgePromotionCandidate,
	decisions []WorkflowRAGKnowledgePromotionDecision,
	binding WorkflowRAGApplicationBinding,
	audits []WorkflowRAGPromotionAudit,
) {
	t.Helper()
	assertWorkflowRAGContractValid(t, workflowRAGPromotionCandidateSchemaVersion, candidate)
	for _, decision := range decisions {
		assertWorkflowRAGContractValid(t, workflowRAGPromotionDecisionSchemaVersion, decision)
	}
	assertWorkflowRAGContractValid(t, workflowRAGApplicationBindingSchemaVersion, binding)
	for _, audit := range audits {
		assertWorkflowRAGContractValid(t, workflowRAGPromotionAuditSchemaVersion, audit)
	}
	payload, err := encodeWorkflowRAGPromotionCandidate(candidate)
	if err != nil {
		t.Fatalf("encode promotion candidate: %v", err)
	}
	if _, err = decodeWorkflowRAGPromotionCandidate(payload); err != nil {
		t.Fatalf("decode promotion candidate: %v", err)
	}
	unknown := append(append([]byte(nil), payload[:len(payload)-1]...), []byte(`,"unknown":true}`)...)
	if _, err = decodeWorkflowRAGPromotionCandidate(unknown); err == nil {
		t.Fatal("promotion candidate decoder accepted an unknown field")
	}
}

var _ workflowRAGPromotionRepository = workflowRAGPromotionFailingAppendRepository{}
