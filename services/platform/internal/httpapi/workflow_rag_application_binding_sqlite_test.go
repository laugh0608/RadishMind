package httpapi

import (
	"context"
	"path/filepath"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
)

func TestWorkflowRAGPromotionBindingPublishSQLiteContinuousRestartChain(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-rag-promotion-chain.db")
	first, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: localPersistenceSQLiteMigrations()})
	if err != nil {
		t.Fatalf("open shared SQLite promotion chain: %v", err)
	}
	draftRepository := newSQLiteApplicationConfigurationDraftRepository(first.DB())
	publishRepository := newSQLiteApplicationPublishCandidateRepository(first.DB())
	promotionRepository := newSQLiteWorkflowRAGPromotionRepository(first.DB())
	fixture := newWorkflowRAGPromotionTestFixture(t)
	fixture.ctx.RequestContext = context.Background()
	fixture.service.draftRepository = draftRepository
	fixture.service.repository = promotionRepository
	draftContext := workflowRAGPromotionDraftContext(fixture.ctx)
	draftContext.WriteEnabled = true
	draftContext.BindingEnabled = true
	createdDraft := newApplicationConfigurationDraftService(draftRepository).Save(draftContext, fixture.draft.ApplicationConfigurationDraftPayload, 0)
	if createdDraft.Draft == nil {
		t.Fatalf("create SQLite source draft: %#v", createdDraft)
	}
	fixture.draft = *createdDraft.Draft
	fixture.createInput.DraftID = createdDraft.Draft.DraftID
	fixture.createInput.ExpectedDraftVersion = createdDraft.Draft.DraftVersion
	createdPromotion := fixture.createCandidate(t)
	fixture.advanceRequest("sqlite_chain_approve")
	approvedPromotion := fixture.service.Decide(fixture.ctx, createdPromotion.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "批准 SQLite 连续链的不可变绑定资格",
	})
	if approvedPromotion.Binding == nil {
		t.Fatalf("approve SQLite promotion binding: %#v", approvedPromotion)
	}

	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.validateBinding = func(ctx ApplicationConfigurationDraftContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return fixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromDraft(ctx), ref, true)
	}
	payload := createdDraft.Draft.ApplicationConfigurationDraftPayload
	payload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
	payload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&approvedPromotion.Binding.WorkflowRAGApplicationBindingRef)
	attached := draftService.Save(draftContext, payload, 1)
	if attached.Draft == nil || attached.Draft.DraftVersion != 2 {
		t.Fatalf("attach SQLite binding to exact source draft: %#v", attached)
	}
	publishContext := workflowRAGBindingPublishContext(fixture)
	publishService := newApplicationPublishCandidateService(draftRepository, publishRepository, func(ApplicationPublishContext) (ApplicationSummary, error) {
		return fixture.application, fixture.applicationErr
	})
	publishService.validateBinding = func(ctx ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return fixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromPublish(ctx), ref, false)
	}
	createdPublish := publishService.Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-sqlite-rag-continuous", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: attached.Draft.DraftVersion,
	})
	if createdPublish.Candidate == nil || createdPublish.Candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV2 {
		t.Fatalf("create SQLite bound publish candidate: %#v", createdPublish)
	}
	approvedPublish := publishService.Review(publishContext, createdPublish.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "发布治理重新校验 SQLite 不可变绑定",
	})
	if approvedPublish.Candidate == nil || approvedPublish.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("approve SQLite bound publish candidate: %#v", approvedPublish)
	}

	if err := first.Close(); err != nil {
		t.Fatalf("close shared SQLite promotion chain: %v", err)
	}
	if _, failure := fixture.service.resolveEligibleBinding(fixture.ctx, approvedPromotion.Binding.WorkflowRAGApplicationBindingRef, false); failure != WorkflowRAGPromotionFailureStoreUnavailable {
		t.Fatalf("closed SQLite promotion store fell back: %s", failure)
	}
	if result := draftService.Read(draftContext, attached.Draft.DraftID); result.FailureCode != ApplicationDraftFailureStoreUnavailable {
		t.Fatalf("closed SQLite draft store fell back: %#v", result)
	}
	if result := publishService.Read(publishContext, createdPublish.Candidate.CandidateID); result.FailureCode != ApplicationPublishFailureStoreUnavailable {
		t.Fatalf("closed SQLite publish store fell back: %#v", result)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: databasePath, Migrations: localPersistenceSQLiteMigrations()})
	if err != nil {
		t.Fatalf("restart shared SQLite promotion chain: %v", err)
	}
	t.Cleanup(func() { _ = restarted.Close() })
	restartedDraftRepository := newSQLiteApplicationConfigurationDraftRepository(restarted.DB())
	restartedPromotionService := fixture.service
	restartedPromotionService.repository = newSQLiteWorkflowRAGPromotionRepository(restarted.DB())
	restartedPromotionService.draftRepository = restartedDraftRepository
	if _, failure := restartedPromotionService.resolveEligibleBinding(fixture.ctx, approvedPromotion.Binding.WorkflowRAGApplicationBindingRef, false); failure != "" {
		t.Fatalf("restore exact SQLite binding after restart: %s", failure)
	}
	restoredDraft := newApplicationConfigurationDraftService(restartedDraftRepository).Read(draftContext, attached.Draft.DraftID)
	restartedPublishService := newApplicationPublishCandidateService(restartedDraftRepository, newSQLiteApplicationPublishCandidateRepository(restarted.DB()), func(ApplicationPublishContext) (ApplicationSummary, error) {
		return fixture.application, fixture.applicationErr
	})
	restartedPublishService.validateBinding = func(ctx ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return restartedPromotionService.resolveEligibleBinding(workflowRAGPromotionContextFromPublish(ctx), ref, false)
	}
	restoredPublish := restartedPublishService.Read(publishContext, createdPublish.Candidate.CandidateID)
	if restoredDraft.Draft == nil || restoredDraft.Draft.WorkflowRAGBindingRef == nil || restoredPublish.Candidate == nil ||
		restoredPublish.Candidate.CandidateState != applicationPublishStateApproved || restoredPublish.Candidate.Configuration.WorkflowRAGBindingRef == nil {
		t.Fatalf("SQLite continuous chain did not recover: draft=%#v publish=%#v", restoredDraft, restoredPublish)
	}
}
