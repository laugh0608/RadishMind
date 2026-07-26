package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationPublishCandidateLifecycleAndEligibility(t *testing.T) {
	runApplicationPublishCandidateLifecycleAndEligibility(t, newMemoryApplicationConfigurationDraftRepository(), newMemoryApplicationPublishCandidateRepository())
}

func TestPromptApplicationPublishCandidateV3ExactTemplateReview(t *testing.T) {
	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	candidateRepository := newMemoryApplicationPublishCandidateRepository()
	templateRepository := newMemoryPromptApplicationTemplateRepository()
	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = "app_aaaaaaaaaaaaaaaa"
	draftContext.TemplateBindingEnabled = true
	templateContext := validPromptApplicationTemplateContext()
	templateContext.ActorRef = draftContext.ActorRef
	templateContext.OwnerSubjectRef = draftContext.OwnerSubjectRef
	templateInput := validPromptApplicationTemplateDraftInput()
	templateService := newPromptApplicationTemplateService(templateRepository)
	if saved := templateService.SaveDraft(templateContext, templateInput, 0); saved.Draft == nil {
		t.Fatalf("seed prompt template draft: %#v", saved)
	}
	version := templateService.CreateVersion(templateContext, templateInput.TemplateID, 1)
	if version.Version == nil {
		t.Fatalf("seed prompt template version: %#v", version)
	}
	payload := validApplicationDraftPayload()
	payload.ApplicationID = draftContext.ApplicationID
	payload.DraftID = "app-config-prompt-v3"
	payload.ApplicationKind = "prompt_application"
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.readPromptTemplateVersion = func(ctx ApplicationConfigurationDraftContext, templateID string, templateVersion int) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, templateID, templateVersion)
		if err != nil {
			return PromptApplicationTemplateVersion{}, promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{}).FailureCode
		}
		return value, ""
	}
	if created := draftService.Save(draftContext, payload, 0); created.Draft == nil {
		t.Fatalf("seed prompt application draft: %#v", created)
	}
	bound := draftService.BindPromptTemplate(draftContext, payload.DraftID, PromptApplicationTemplateBindingInput{ExpectedDraftVersion: 1, TemplateID: templateInput.TemplateID, TemplateVersion: 1})
	if bound.Draft == nil {
		t.Fatalf("bind prompt template: %#v", bound)
	}
	publishContext := validApplicationPublishContext()
	publishContext.ApplicationID = draftContext.ApplicationID
	publishContext.PromptTemplateSourceReadEnabled = true
	service := newApplicationPublishCandidateService(draftRepository, candidateRepository, func(ctx ApplicationPublishContext) (ApplicationSummary, error) {
		return ApplicationSummary{ApplicationRef: ctx.ApplicationID, ApplicationKind: "prompt_application", UpdatedAt: payload.BaseApplicationUpdatedAt}, nil
	})
	service.readPromptTemplateVersion = func(ctx ApplicationPublishContext, ref PromptApplicationTemplateRef) (PromptApplicationTemplateVersion, string) {
		value, err := templateRepository.ReadVersion(templateContext, ref.TemplateID, ref.TemplateVersion)
		if err != nil {
			return PromptApplicationTemplateVersion{}, promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{}).FailureCode
		}
		return value, ""
	}
	created := service.Create(publishContext, ApplicationPublishCreateInput{CandidateID: "candidate-prompt-v3", DraftID: payload.DraftID, ExpectedDraftVersion: 2})
	if created.Candidate == nil || created.Candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV3 || created.Candidate.Configuration.PromptTemplateRef == nil || created.Candidate.Configuration.PromptTemplateRef.TemplateDigest != version.Version.TemplateDigest {
		t.Fatalf("create Prompt publish candidate v3: %#v", created)
	}
	approved := service.Review(publishContext, created.Candidate.CandidateID, ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "Reviewed exact Prompt template source and configuration lineage."})
	if approved.Candidate == nil || approved.Candidate.CandidateState != applicationPublishStateApproved || approved.Candidate.ReviewVersion != 1 {
		t.Fatalf("approve Prompt publish candidate v3: %#v", approved)
	}
	deniedContext := publishContext
	deniedContext.PromptTemplateSourceReadEnabled = false
	denied := service.Create(deniedContext, ApplicationPublishCreateInput{CandidateID: "candidate-prompt-denied", DraftID: payload.DraftID, ExpectedDraftVersion: 2})
	if denied.FailureCode != ApplicationPublishFailureScopeDenied {
		t.Fatalf("Prompt candidate creation without source review permission must fail: %#v", denied)
	}
}

func runApplicationPublishCandidateLifecycleAndEligibility(t *testing.T, draftRepository applicationConfigurationDraftRepository, candidateRepository applicationPublishCandidateRepository) {
	t.Helper()
	draftContext := validApplicationDraftContext()
	draft := newApplicationConfigurationDraftService(draftRepository).Save(draftContext, validApplicationDraftPayload(), 0)
	if draft.Draft == nil {
		t.Fatalf("seed application draft: %#v", draft)
	}
	baselineUpdatedAt := draft.Draft.BaseApplicationUpdatedAt
	service := newApplicationPublishCandidateService(draftRepository, candidateRepository, func(requestContext ApplicationPublishContext) (ApplicationSummary, error) {
		return ApplicationSummary{ApplicationRef: requestContext.ApplicationID, UpdatedAt: baselineUpdatedAt}, nil
	})
	service.now = func() time.Time { return time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC) }
	requestContext := validApplicationPublishContext()
	created := service.Create(requestContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-app-flow-v1", DraftID: draft.Draft.DraftID, ExpectedDraftVersion: 1,
		EvidenceRequestIDs: []string{"playground-request-0002", "playground-request-0001", "playground-request-0001"},
	})
	if created.Candidate == nil || created.FailureCode != "" || created.Candidate.CandidateState != applicationPublishStatePending {
		t.Fatalf("create publish candidate: %#v", created)
	}
	if !strings.HasPrefix(created.Candidate.DraftDigest, "sha256:") || len(created.Candidate.DraftDigest) != 71 {
		t.Fatalf("candidate digest is invalid: %s", created.Candidate.DraftDigest)
	}
	if len(created.Candidate.EvidenceRequestIDs) != 2 || created.Candidate.EvidenceRequestIDs[0] != "playground-request-0001" {
		t.Fatalf("evidence refs were not normalized: %#v", created.Candidate.EvidenceRequestIDs)
	}
	if created.Candidate.PromotionEligibility.Eligible || !hasPromotionBlocker(created.Candidate.PromotionEligibility, "publish_review_required") || !hasPromotionBlocker(created.Candidate.PromotionEligibility, "promotion_disabled") {
		t.Fatalf("pending candidate eligibility is incorrect: %#v", created.Candidate.PromotionEligibility)
	}
	duplicate := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-app-flow-v1", DraftID: draft.Draft.DraftID, ExpectedDraftVersion: 1})
	if duplicate.FailureCode != ApplicationPublishFailureImmutableConflict {
		t.Fatalf("candidate create must be immutable: %#v", duplicate)
	}
	approved := service.Review(requestContext, "candidate-app-flow-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "Validated configuration and scoped invocation evidence reviewed."})
	if approved.Candidate == nil || approved.Candidate.ReviewVersion != 1 || approved.Candidate.CandidateState != applicationPublishStateApproved || approved.Candidate.PromotionEligibility.Eligible {
		t.Fatalf("approve review failed or enabled promotion: %#v", approved)
	}
	conflict := service.Review(requestContext, "candidate-app-flow-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "reject", Reason: "Stale reviewer must not overwrite the accepted decision."})
	if conflict.FailureCode != ApplicationPublishFailureReviewVersionConflict || conflict.CurrentReviewVersion != 1 || conflict.CurrentCandidateState != applicationPublishStateApproved {
		t.Fatalf("review CAS conflict failed: %#v", conflict)
	}
	invalidTransition := service.Review(requestContext, "candidate-app-flow-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 1, Decision: "approve", Reason: "Repeated approval is not a valid transition."})
	if invalidTransition.FailureCode != ApplicationPublishFailureReviewTransitionInvalid {
		t.Fatalf("invalid review transition was accepted: %#v", invalidTransition)
	}
	withdrawn := service.Review(requestContext, "candidate-app-flow-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 1, Decision: "withdraw", Reason: "Owner withdrew the approved development candidate."})
	if withdrawn.Candidate == nil || withdrawn.Candidate.CandidateState != applicationPublishStateWithdrawn || withdrawn.Candidate.ReviewVersion != 2 {
		t.Fatalf("approved candidate withdraw failed: %#v", withdrawn)
	}
	baselineUpdatedAt = "2026-07-12T11:00:00Z"
	drifted := service.Read(requestContext, "candidate-app-flow-v1")
	if drifted.Candidate == nil || !hasPromotionBlocker(drifted.Candidate.PromotionEligibility, ApplicationPublishFailureBaseRevisionChanged) {
		t.Fatalf("baseline drift was not reported: %#v", drifted)
	}
}

func TestApplicationPublishCandidateDraftBindingAndScope(t *testing.T) {
	runApplicationPublishCandidateDraftBindingAndScope(t, newMemoryApplicationConfigurationDraftRepository(), newMemoryApplicationPublishCandidateRepository())
}

func runApplicationPublishCandidateDraftBindingAndScope(t *testing.T, draftRepository applicationConfigurationDraftRepository, candidateRepository applicationPublishCandidateRepository) {
	t.Helper()
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftContext := validApplicationDraftContext()
	createdDraft := draftService.Save(draftContext, validApplicationDraftPayload(), 0)
	service := newApplicationPublishCandidateService(draftRepository, candidateRepository, validApplicationPublishBaseline)
	requestContext := validApplicationPublishContext()
	wrongVersion := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-wrong-version", DraftID: createdDraft.Draft.DraftID, ExpectedDraftVersion: 2})
	if wrongVersion.FailureCode != ApplicationPublishFailureDraftVersionConflict || wrongVersion.CurrentDraftVersion != 1 {
		t.Fatalf("draft version mismatch was not rejected: %#v", wrongVersion)
	}
	requestContext.ApplicationID = "app_docs_assistant"
	notFound := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-cross-application", DraftID: createdDraft.Draft.DraftID, ExpectedDraftVersion: 1})
	if notFound.FailureCode != ApplicationPublishFailureDraftNotFound {
		t.Fatalf("cross-application draft read must fail closed: %#v", notFound)
	}
	requestContext = validApplicationPublishContext()
	secret := service.Review(requestContext, "candidate-missing", ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "Authorization: Bearer hidden"})
	if secret.FailureCode != ApplicationPublishFailureSecretForbidden {
		t.Fatalf("review secret was not rejected: %#v", secret)
	}
}

func TestApplicationPublishCandidateDraftDriftAndSupersede(t *testing.T) {
	runApplicationPublishCandidateDraftDriftAndSupersede(t, newMemoryApplicationConfigurationDraftRepository(), newMemoryApplicationPublishCandidateRepository())
}

func runApplicationPublishCandidateDraftDriftAndSupersede(t *testing.T, draftRepository applicationConfigurationDraftRepository, repository applicationPublishCandidateRepository) {
	t.Helper()
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftContext := validApplicationDraftContext()
	payload := validApplicationDraftPayload()
	if seeded := draftService.Save(draftContext, payload, 0); seeded.Draft == nil {
		t.Fatalf("seed draft: %#v", seeded)
	}
	service := newApplicationPublishCandidateService(draftRepository, repository, validApplicationPublishBaseline)
	sequence := 0
	service.now = func() time.Time { sequence++; return time.Date(2026, 7, 12, 12, sequence, 0, 0, time.UTC) }
	requestContext := validApplicationPublishContext()
	first := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-v1", DraftID: payload.DraftID, ExpectedDraftVersion: 1})
	if first.Candidate == nil {
		t.Fatalf("create first candidate: %#v", first)
	}
	payload.Description = "Changed sanitized configuration"
	if updated := draftService.Save(draftContext, payload, 1); updated.Draft == nil || updated.Draft.DraftVersion != 2 {
		t.Fatalf("update draft: %#v", updated)
	}
	second := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-v2", DraftID: payload.DraftID, ExpectedDraftVersion: 2})
	if second.Candidate == nil {
		t.Fatalf("create second candidate: %#v", second)
	}
	readFirst := service.Read(requestContext, "candidate-v1")
	if readFirst.Candidate == nil || !hasPromotionBlocker(readFirst.Candidate.PromotionEligibility, "publish_candidate_draft_changed") || !hasPromotionBlocker(readFirst.Candidate.PromotionEligibility, "publish_candidate_superseded") {
		t.Fatalf("old candidate drift/supersede blockers missing: %#v", readFirst)
	}
	candidates, err := repository.List(requestContext)
	if err != nil || len(candidates) != 2 || candidates[0].CandidateID != "candidate-v2" || candidates[1].CandidateID != "candidate-v1" {
		t.Fatalf("publish candidate list order is unstable: %#v err=%v", candidates, err)
	}
}

func TestApplicationPublishCandidateHTTPRoutes(t *testing.T) {
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, ApplicationDraftDevHTTPEnabled: true, ApplicationDraftDevWriteEnabled: true,
		ApplicationPublishDevHTTPEnabled: true, ApplicationPublishDevWriteEnabled: true,
	}, Options{BuildVersion: "test"})
	draftPayload := validApplicationDraftPayload()
	draftCreate := performApplicationDraftRequest(t, server, http.MethodPost, "/v1/user-workspace/application-drafts", applicationConfigurationDraftSaveBody{Draft: draftPayload}, "application_drafts:read,application_drafts:write", draftPayload.ApplicationID)
	if draftCreate.Draft == nil {
		t.Fatalf("seed draft over HTTP: %#v", draftCreate)
	}
	created := performApplicationPublishRequest(t, server, http.MethodPost, "/v1/user-workspace/application-publish-candidates", applicationPublishCandidateCreateBody{
		CandidateID: "candidate-http-v1", DraftID: draftPayload.DraftID, ExpectedDraftVersion: 1,
	}, "application_publish_candidates:write", draftPayload.ApplicationID)
	if created.FailureCode != nil || created.Candidate == nil || created.Candidate.CandidateID != "candidate-http-v1" {
		t.Fatalf("create candidate route failed: %#v", created)
	}
	approved := performApplicationPublishRequest(t, server, http.MethodPost, "/v1/user-workspace/application-publish-candidates/candidate-http-v1/reviews", applicationPublishCandidateReviewBody{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "Reviewed through the scoped development route.",
	}, "application_publish_candidates:review", draftPayload.ApplicationID)
	if approved.Candidate == nil || approved.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("review route failed: %#v", approved)
	}
	read := performApplicationPublishRequest(t, server, http.MethodGet, "/v1/user-workspace/application-publish-candidates/candidate-http-v1?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "application_publish_candidates:read", draftPayload.ApplicationID)
	if read.Candidate == nil || read.Candidate.ReviewVersion != 1 {
		t.Fatalf("read route failed: %#v", read)
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/application-publish-candidates?workspace_id=workspace_demo&application_id=app_flow_copilot", nil)
	setApplicationPublishHeaders(request, "application_publish_candidates:read", draftPayload.ApplicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	var list applicationPublishCandidateListEnvelope
	if recorder.Code != http.StatusOK || json.Unmarshal(recorder.Body.Bytes(), &list) != nil || len(list.CandidateSummaries) != 1 || list.FailureCode != nil {
		t.Fatalf("list route failed: %d %s", recorder.Code, recorder.Body.String())
	}
	denied := performApplicationPublishRequest(t, server, http.MethodGet, "/v1/user-workspace/application-publish-candidates/candidate-http-v1?workspace_id=workspace_demo&application_id=app_flow_copilot", nil, "application_publish_candidates:read", "app_docs_assistant")
	if denied.FailureCode == nil || *denied.FailureCode != ApplicationPublishFailureScopeDenied || denied.Candidate != nil {
		t.Fatalf("publish header scope mismatch must fail closed: %#v", denied)
	}
}

func validApplicationPublishContext() ApplicationPublishContext {
	return ApplicationPublishContext{
		RequestContext: context.Background(), RequestID: "request-application-publish-test", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ActorRef: "subject_platform_ops",
		OwnerSubjectRef: "subject_platform_ops", AuditRef: "audit_application_publish_test", WriteEnabled: true,
	}
}

func validApplicationPublishBaseline(requestContext ApplicationPublishContext) (ApplicationSummary, error) {
	return ApplicationSummary{ApplicationRef: requestContext.ApplicationID, UpdatedAt: "2026-05-31T10:20:00Z"}, nil
}

func hasPromotionBlocker(eligibility ApplicationPromotionEligibility, code string) bool {
	for _, blocker := range eligibility.Blockers {
		if blocker.Code == code {
			return true
		}
	}
	return false
}

func performApplicationPublishRequest(t *testing.T, server *Server, method, target string, body any, scopes, applicationID string) applicationPublishCandidateEnvelope {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal application publish body: %v", err)
		}
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(raw))
	setApplicationPublishHeaders(request, scopes, applicationID)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope applicationPublishCandidateEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode application publish response: %v", err)
	}
	return envelope
}

func setApplicationPublishHeaders(request *http.Request, scopes, applicationID string) {
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, scopes)
	request.Header.Set(applicationPublishDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(applicationPublishDevApplicationHeader, applicationID)
}
