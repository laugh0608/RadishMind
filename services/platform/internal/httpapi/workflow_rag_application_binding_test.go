package httpapi

import (
	"testing"
	"time"
)

func TestApplicationDraftV2AttachesAndReplacesOnlyEligibleExactBinding(t *testing.T) {
	fixture, firstApproval := approvedWorkflowRAGPromotionFixture(t)
	draftService, draftContext := workflowRAGBindingDraftService(fixture)

	payload := fixture.draft.ApplicationConfigurationDraftPayload
	payload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
	payload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&firstApproval.Binding.WorkflowRAGApplicationBindingRef)

	withoutPermission := draftContext
	withoutPermission.BindingEnabled = false
	if result := draftService.Save(withoutPermission, payload, 1); result.FailureCode != ApplicationDraftFailureScopeDenied || result.Draft != nil {
		t.Fatalf("binding attach without bind permission was accepted: %#v", result)
	}

	mixed := payload
	mixed.Description = "Binding attach must not carry a configuration edit."
	if result := draftService.Save(draftContext, mixed, 1); result.FailureCode != WorkflowRAGPromotionFailureBindingNotEligible || result.Draft != nil {
		t.Fatalf("binding attach carried another configuration change: %#v", result)
	}

	attached := draftService.Save(draftContext, payload, 1)
	if attached.FailureCode != "" || attached.Draft == nil || attached.Draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV2 ||
		attached.Draft.DraftVersion != 2 || attached.Draft.WorkflowRAGBindingRef == nil || *attached.Draft.WorkflowRAGBindingRef != firstApproval.Binding.WorkflowRAGApplicationBindingRef ||
		attached.Draft.DraftDigest == fixture.draft.DraftDigest {
		t.Fatalf("eligible binding was not attached as a ref-only v2 draft: %#v", attached)
	}
	if _, failure := fixture.service.resolveEligibleBinding(fixture.ctx, firstApproval.Binding.WorkflowRAGApplicationBindingRef, false); failure != "" {
		t.Fatalf("attached binding incorrectly required its source draft to remain latest: %s", failure)
	}

	retainedPayload := attached.Draft.ApplicationConfigurationDraftPayload
	retainedPayload.Description = "A later draft version may retain the same immutable binding."
	retained := draftService.Save(draftContext, retainedPayload, 2)
	if retained.FailureCode != "" || retained.Draft == nil || retained.Draft.DraftVersion != 3 || retained.Draft.WorkflowRAGBindingRef == nil {
		t.Fatalf("normal edit with the same binding failed: %#v", retained)
	}

	fixture.createInput.ExpectedDraftVersion = retained.Draft.DraftVersion
	fixture.advanceRequest("replacement_create")
	replacementCandidate := fixture.createCandidate(t)
	fixture.advanceRequest("replacement_approve")
	replacementApproval := fixture.service.Decide(fixture.ctx, replacementCandidate.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "批准显式替换当前不可变绑定",
	})
	if replacementApproval.FailureCode != "" || replacementApproval.Binding == nil {
		t.Fatalf("create replacement binding: %#v", replacementApproval)
	}
	replacementPayload := retained.Draft.ApplicationConfigurationDraftPayload
	replacementPayload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&replacementApproval.Binding.WorkflowRAGApplicationBindingRef)
	replaced := draftService.Save(draftContext, replacementPayload, 3)
	if replaced.FailureCode != "" || replaced.Draft == nil || replaced.Draft.DraftVersion != 4 ||
		replaced.Draft.WorkflowRAGBindingRef == nil || *replaced.Draft.WorkflowRAGBindingRef != replacementApproval.Binding.WorkflowRAGApplicationBindingRef {
		t.Fatalf("eligible exact replacement binding was not saved: %#v", replaced)
	}
}

func TestApplicationDraftV1DigestCompatibilityAndV2Shape(t *testing.T) {
	service := newApplicationConfigurationDraftService(newMemoryApplicationConfigurationDraftRepository())
	created := service.Save(validApplicationDraftContext(), validApplicationDraftPayload(), 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV1 ||
		!workflowRAGDigestPattern.MatchString(created.Draft.DraftDigest) || created.Draft.WorkflowRAGBindingRef != nil {
		t.Fatalf("v1 application draft compatibility regressed: %#v", created)
	}
	invalid := validApplicationDraftPayload()
	invalid.WorkflowRAGBindingRef = &WorkflowRAGApplicationBindingRef{BindingID: "wragb_aaaaaaaaaaaaaaaa", BindingVersion: 1, BindingDigest: workflowRAGSHA256("binding")}
	if result := service.Validate(validApplicationDraftContext(), invalid); result.ValidationSummary.IsValid {
		t.Fatalf("v1 draft accepted a v2-only binding ref: %#v", result)
	}
}

func TestPublishCandidateV2RevalidatesBindingAtCreateApproveAndRead(t *testing.T) {
	fixture, approval := approvedWorkflowRAGPromotionFixture(t)
	draftService, draftContext := workflowRAGBindingDraftService(fixture)
	payload := fixture.draft.ApplicationConfigurationDraftPayload
	payload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
	payload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&approval.Binding.WorkflowRAGApplicationBindingRef)
	attached := draftService.Save(draftContext, payload, 1)
	if attached.Draft == nil {
		t.Fatalf("attach binding before publish review: %#v", attached)
	}

	repository := newMemoryApplicationPublishCandidateRepository()
	service := newApplicationPublishCandidateService(fixture.drafts, repository, func(ApplicationPublishContext) (ApplicationSummary, error) {
		return fixture.application, fixture.applicationErr
	})
	service.validateBinding = func(ctx ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return fixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromPublish(ctx), ref, false)
	}
	service.now = func() time.Time { return fixture.now.Add(time.Hour) }
	publishContext := workflowRAGBindingPublishContext(fixture)

	first := service.Create(publishContext, ApplicationPublishCreateInput{CandidateID: "candidate-rag-binding-v2-a", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: 2})
	second := service.Create(publishContext, ApplicationPublishCreateInput{CandidateID: "candidate-rag-binding-v2-b", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: 2})
	if first.FailureCode != "" || first.Candidate == nil || second.FailureCode != "" || second.Candidate == nil ||
		first.Candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV2 || first.Candidate.Configuration.WorkflowRAGBindingRef == nil ||
		first.Candidate.DraftDigest != attached.Draft.DraftDigest {
		t.Fatalf("publish v2 candidate did not retain the exact draft binding: first=%#v second=%#v", first, second)
	}
	approved := service.Review(publishContext, first.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "发布审查前重新校验不可变 RAG 绑定",
	})
	if approved.FailureCode != "" || approved.Candidate == nil || approved.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("eligible bound publish candidate could not be approved: %#v", approved)
	}

	fixture.advanceRequest("binding_cancel")
	canceled := fixture.service.Decide(fixture.ctx, approval.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 2, Decision: workflowRAGPromotionDecisionCancel, Reason: "撤销应用配置使用资格",
	})
	if canceled.FailureCode != "" {
		t.Fatalf("cancel approved promotion: %#v", canceled)
	}
	read := service.Read(publishContext, first.Candidate.CandidateID)
	if read.Candidate == nil || !hasPromotionBlocker(read.Candidate.PromotionEligibility, WorkflowRAGPromotionFailureBindingNotEligible) {
		t.Fatalf("canceled binding did not fail closed at read time: %#v", read)
	}
	blockedApproval := service.Review(publishContext, second.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "失效绑定不得继续批准",
	})
	if blockedApproval.FailureCode != ApplicationPublishFailureBindingNotEligible {
		t.Fatalf("approve accepted a canceled binding: %#v", blockedApproval)
	}
	storedSecond, err := repository.Read(publishContext, second.Candidate.CandidateID)
	if err != nil || storedSecond.ReviewVersion != 0 || len(storedSecond.Reviews) != 0 {
		t.Fatalf("failed approval appended partial review state: candidate=%#v err=%v", storedSecond, err)
	}
	if blockedCreate := service.Create(publishContext, ApplicationPublishCreateInput{CandidateID: "candidate-rag-binding-v2-c", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: 2}); blockedCreate.FailureCode != ApplicationPublishFailureBindingNotEligible {
		t.Fatalf("create accepted a canceled binding: %#v", blockedCreate)
	}
}

func TestPublishCandidateBindingEligibilityMapsAuthorityFailures(t *testing.T) {
	tests := []struct {
		name    string
		failure string
		mutate  func(*testing.T, *workflowRAGPromotionTestFixture)
	}{
		{"dataset archived", WorkflowRAGPromotionFailureDatasetArchived, func(t *testing.T, fixture *workflowRAGPromotionTestFixture) { fixture.archiveDataset(t) }},
		{"application archived", WorkflowRAGPromotionFailureApplicationArchived, func(_ *testing.T, fixture *workflowRAGPromotionTestFixture) {
			fixture.applicationErr = errApplicationCatalogArchived
		}},
		{"promotion store unavailable", WorkflowRAGPromotionFailureStoreUnavailable, func(_ *testing.T, fixture *workflowRAGPromotionTestFixture) { fixture.promotions.available = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture, approval := approvedWorkflowRAGPromotionFixture(t)
			draftService, draftContext := workflowRAGBindingDraftService(fixture)
			payload := fixture.draft.ApplicationConfigurationDraftPayload
			payload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
			payload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&approval.Binding.WorkflowRAGApplicationBindingRef)
			attached := draftService.Save(draftContext, payload, 1)
			if attached.Draft == nil {
				t.Fatalf("attach binding: %#v", attached)
			}
			service := newApplicationPublishCandidateService(fixture.drafts, newMemoryApplicationPublishCandidateRepository(), validApplicationPublishBaseline)
			service.validateBinding = func(ctx ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
				return fixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromPublish(ctx), ref, false)
			}
			ctx := workflowRAGBindingPublishContext(fixture)
			created := service.Create(ctx, ApplicationPublishCreateInput{CandidateID: "candidate-rag-drift-v2", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: 2})
			if created.Candidate == nil {
				t.Fatalf("create candidate before authority drift: %#v", created)
			}
			test.mutate(t, fixture)
			read := service.Read(ctx, created.Candidate.CandidateID)
			if read.Candidate == nil || !hasPromotionBlocker(read.Candidate.PromotionEligibility, test.failure) {
				t.Fatalf("expected blocker %s after authority failure: %#v", test.failure, read)
			}
		})
	}
}

func approvedWorkflowRAGPromotionFixture(t *testing.T) (*workflowRAGPromotionTestFixture, WorkflowRAGPromotionResult) {
	t.Helper()
	fixture := newWorkflowRAGPromotionTestFixture(t)
	created := fixture.createCandidate(t)
	fixture.advanceRequest("binding_approve")
	approved := fixture.service.Decide(fixture.ctx, created.Candidate.CandidateID, WorkflowRAGPromotionDecisionInput{
		ExpectedRecordVersion: 1, Decision: workflowRAGPromotionDecisionApprove, Reason: "批准不可变应用配置绑定资格",
	})
	if approved.FailureCode != "" || approved.Binding == nil {
		t.Fatalf("approve promotion binding: %#v", approved)
	}
	return fixture, approved
}

func workflowRAGBindingDraftService(fixture *workflowRAGPromotionTestFixture) (applicationConfigurationDraftService, ApplicationConfigurationDraftContext) {
	service := newApplicationConfigurationDraftService(fixture.drafts)
	service.now = func() time.Time { return fixture.now.Add(30 * time.Minute) }
	service.validateBinding = func(ctx ApplicationConfigurationDraftContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return fixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromDraft(ctx), ref, true)
	}
	ctx := workflowRAGPromotionDraftContext(fixture.ctx)
	ctx.WriteEnabled = true
	ctx.BindingEnabled = true
	return service, ctx
}

func workflowRAGBindingPublishContext(fixture *workflowRAGPromotionTestFixture) ApplicationPublishContext {
	return ApplicationPublishContext{
		RequestContext: fixture.ctx.RequestContext, RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: fixture.ctx.ApplicationID, ActorRef: fixture.ctx.ActorRef,
		OwnerSubjectRef: fixture.ctx.OwnerSubjectRef, AuditRef: fixture.ctx.AuditRef, WriteEnabled: true, RAGPromotionReadEnabled: true,
	}
}

func testWorkflowRAGApplicationBindingForDraft(draft ApplicationConfigurationDraft) WorkflowRAGApplicationBinding {
	ref := WorkflowRAGApplicationBindingRef{
		BindingID: "wragb_aaaaaaaaaaaaaaaa", BindingVersion: 1, BindingDigest: workflowRAGSHA256("test-application-binding"),
	}
	return WorkflowRAGApplicationBinding{
		SchemaVersion: workflowRAGApplicationBindingSchemaVersion, WorkflowRAGApplicationBindingRef: ref,
		Evidence: WorkflowRAGPromotionEvidenceBinding{SourceDraft: WorkflowRAGPromotionSourceDraftBinding{
			DraftID: draft.DraftID, DraftVersion: draft.DraftVersion, DraftDigest: draft.DraftDigest,
			BaseApplicationUpdatedAt: draft.BaseApplicationUpdatedAt,
		}},
	}
}
