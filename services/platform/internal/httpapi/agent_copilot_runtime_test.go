package httpapi

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAgentCopilotBatchCConfigurationCandidateAndAssignmentLifecycle(t *testing.T) {
	fixture := newAgentCopilotBatchCFixture(t)

	if _, _, err := fixture.runtimeRepository.Read(fixture.runtimeContext); !errorsIsAgentCopilotRuntimeNotFound(err) {
		t.Fatalf("candidate approval must not auto-activate an assignment: %v", err)
	}
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 0, Action: "activate", CandidateID: fixture.firstCandidate.CandidateID,
	})
	if activated.FailureCode != "" || activated.Assignment == nil || activated.Assignment.AssignmentVersion != 1 ||
		activated.Assignment.State != "active" || len(activated.Events) != 1 ||
		activated.Assignment.AgentCopilotProfileRef != *fixture.boundDraft.AgentCopilotProfileRef {
		t.Fatalf("activate Agent Copilot assignment: %#v", activated)
	}

	secondCandidate := fixture.createApprovedCandidate(t, "candidate-agent-copilot-2")
	stale := fixture.runtimeService.Read(fixture.runtimeContext)
	if stale.FailureCode != AgentCopilotRuntimeFailureCandidate || stale.Assignment == nil {
		t.Fatalf("superseded assignment must fail read-time eligibility: %#v", stale)
	}
	replaced := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 1, Action: "replace", CandidateID: secondCandidate.CandidateID,
	})
	if replaced.FailureCode != "" || replaced.Assignment == nil || replaced.Assignment.AssignmentVersion != 2 ||
		replaced.Assignment.CandidateID != secondCandidate.CandidateID || len(replaced.Events) != 2 {
		t.Fatalf("replace Agent Copilot assignment: %#v", replaced)
	}
	revoked := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 2, Action: "revoke",
	})
	if revoked.FailureCode != "" || revoked.Assignment == nil || revoked.Assignment.State != "revoked" ||
		revoked.Assignment.AssignmentVersion != 3 || revoked.Assignment.RevokedAt == nil || len(revoked.Events) != 3 {
		t.Fatalf("revoke Agent Copilot assignment: %#v", revoked)
	}
	if retry := fixture.runtimeService.Decide(fixture.runtimeContext, AgentCopilotRuntimeDecisionInput{
		ExpectedAssignmentVersion: 2, Action: "revoke",
	}); retry.FailureCode != AgentCopilotRuntimeFailureVersionConflict {
		t.Fatalf("stale assignment mutation must fail CAS: %#v", retry)
	}
}

func TestAgentCopilotBatchCExactProfileReloadAndPermissions(t *testing.T) {
	profileRepository := newMemoryAgentCopilotProfileRepository()
	profileContext := agentCopilotBatchCProfileContext()
	profileVersion := createAgentCopilotBatchCProfileVersion(t, profileRepository, profileContext)

	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	draftContext := agentCopilotBatchCDraftContext()
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.readAgentProfileVersion = func(_ ApplicationConfigurationDraftContext, profileID string, profileVersionNumber int) (AgentCopilotProfileVersionV1, string) {
		version, err := profileRepository.ReadVersion(profileContext, profileID, profileVersionNumber)
		if err != nil {
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureNotFound
		}
		return version, ""
	}
	base := draftService.Save(draftContext, agentCopilotBatchCDraftPayload(), 0)
	if base.FailureCode != "" || base.Draft == nil {
		t.Fatalf("save base Agent draft: %#v", base)
	}
	deniedContext := draftContext
	deniedContext.ProfileBindingEnabled = false
	if denied := draftService.BindAgentCopilotProfile(deniedContext, base.Draft.DraftID, AgentCopilotProfileBindingInput{
		ExpectedDraftVersion: 1, ProfileID: profileVersion.ProfileID, ProfileVersion: 1,
	}); denied.FailureCode != ApplicationDraftFailureScopeDenied {
		t.Fatalf("profile binding without bind permission must fail closed: %#v", denied)
	}
	bound := draftService.BindAgentCopilotProfile(draftContext, base.Draft.DraftID, AgentCopilotProfileBindingInput{
		ExpectedDraftVersion: 1, ProfileID: profileVersion.ProfileID, ProfileVersion: 1,
	})
	if bound.FailureCode != "" || bound.Draft == nil || bound.Draft.SchemaVersion != applicationConfigurationDraftSchemaVersionV4 ||
		bound.Draft.AgentCopilotProfileRef == nil || bound.Draft.AgentCopilotProfileRef.ProfileDigest != profileVersion.ProfileDigest {
		t.Fatalf("bind exact Agent Copilot profile version: %#v", bound)
	}

	publishContext := agentCopilotBatchCPublishContext()
	publishService := newApplicationPublishCandidateService(draftRepository, newMemoryApplicationPublishCandidateRepository(), agentCopilotBatchCBaseline)
	publishService.readAgentProfileVersion = func(_ ApplicationPublishContext, ref AgentCopilotProfileRef) (AgentCopilotProfileVersionV1, string) {
		version, err := profileRepository.ReadVersion(profileContext, ref.ProfileID, ref.ProfileVersion)
		if err != nil {
			if err == errAgentCopilotProfileDigest {
				return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureDigestDrift
			}
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureNotFound
		}
		if validateStoredAgentCopilotProfileVersion(profileContext, version) != nil {
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureDigestDrift
		}
		return version, ""
	}
	deniedPublish := publishContext
	deniedPublish.AgentProfileSourceReadEnabled = false
	if denied := publishService.Create(deniedPublish, ApplicationPublishCreateInput{
		CandidateID: "candidate-agent-denied", DraftID: bound.Draft.DraftID, ExpectedDraftVersion: bound.Draft.DraftVersion,
	}); denied.FailureCode != ApplicationPublishFailureScopeDenied {
		t.Fatalf("v4 candidate creation without profile source permission must fail closed: %#v", denied)
	}
	created := publishService.Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-agent-exact", DraftID: bound.Draft.DraftID, ExpectedDraftVersion: bound.Draft.DraftVersion,
	})
	if created.FailureCode != "" || created.Candidate == nil ||
		created.Candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV4 ||
		created.Candidate.Configuration.AgentCopilotProfileRef == nil {
		t.Fatalf("create v4 Agent candidate: %#v", created)
	}
	profileRepository.mu.Lock()
	key := agentCopilotProfileRepositoryKey(profileContext, profileVersion.ProfileID)
	corrupt := profileRepository.versions[key][1]
	corrupt.PolicyDigest = "sha256:" + strings.Repeat("f", 64)
	profileRepository.versions[key][1] = corrupt
	profileRepository.mu.Unlock()
	if review := publishService.Review(publishContext, created.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "Reviewed exact Agent Copilot profile source.",
	}); review.FailureCode != ApplicationPublishFailureStoreUnavailable {
		t.Fatalf("profile digest drift must block candidate approval: %#v", review)
	}
}

func TestMemoryAgentCopilotRuntimeConcurrentCAS(t *testing.T) {
	ctx := agentCopilotBatchCRuntimeContext()
	repository := newMemoryAgentCopilotRuntimeRepository()
	assignment, event := validAgentCopilotRuntimeMutation(t, ctx)
	const writers = 8
	var wait sync.WaitGroup
	results := make(chan error, writers)
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- repository.Apply(ctx, 0, assignment, event)
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errorsIsAgentCopilotRuntimeVersionConflict(err):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent CAS result: %v", err)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("Agent Copilot assignment CAS accepted %d successes and %d conflicts", successes, conflicts)
	}
}

type agentCopilotBatchCFixture struct {
	profileRepository *memoryAgentCopilotProfileRepository
	draftRepository   *memoryApplicationConfigurationDraftRepository
	publishRepository *memoryApplicationPublishCandidateRepository
	runtimeRepository *memoryAgentCopilotRuntimeRepository
	runtimeService    agentCopilotRuntimeService
	runtimeContext    AgentCopilotRuntimeContext
	boundDraft        ApplicationConfigurationDraft
	firstCandidate    ApplicationPublishCandidate
	publishService    applicationPublishCandidateService
	nowSequence       int
}

func newAgentCopilotBatchCFixture(t *testing.T) *agentCopilotBatchCFixture {
	t.Helper()
	profileRepository := newMemoryAgentCopilotProfileRepository()
	profileContext := agentCopilotBatchCProfileContext()
	createAgentCopilotBatchCProfileVersion(t, profileRepository, profileContext)

	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	draftContext := agentCopilotBatchCDraftContext()
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftService.readAgentProfileVersion = func(_ ApplicationConfigurationDraftContext, profileID string, profileVersion int) (AgentCopilotProfileVersionV1, string) {
		version, err := profileRepository.ReadVersion(profileContext, profileID, profileVersion)
		if err != nil {
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureNotFound
		}
		return version, ""
	}
	base := draftService.Save(draftContext, agentCopilotBatchCDraftPayload(), 0)
	if base.FailureCode != "" || base.Draft == nil {
		t.Fatalf("save Agent configuration draft: %#v", base)
	}
	bound := draftService.BindAgentCopilotProfile(draftContext, base.Draft.DraftID, AgentCopilotProfileBindingInput{
		ExpectedDraftVersion: 1, ProfileID: "acpf_aaaaaaaaaaaaaaaa", ProfileVersion: 1,
	})
	if bound.FailureCode != "" || bound.Draft == nil {
		t.Fatalf("bind Agent profile: %#v", bound)
	}

	publishRepository := newMemoryApplicationPublishCandidateRepository()
	publishService := newApplicationPublishCandidateService(draftRepository, publishRepository, agentCopilotBatchCBaseline)
	publishService.readAgentProfileVersion = func(_ ApplicationPublishContext, ref AgentCopilotProfileRef) (AgentCopilotProfileVersionV1, string) {
		version, err := profileRepository.ReadVersion(profileContext, ref.ProfileID, ref.ProfileVersion)
		if err != nil {
			if err == errAgentCopilotProfileDigest {
				return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureDigestDrift
			}
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureNotFound
		}
		if validateStoredAgentCopilotProfileVersion(profileContext, version) != nil {
			return AgentCopilotProfileVersionV1{}, AgentCopilotProfileFailureDigestDrift
		}
		return version, ""
	}
	fixture := &agentCopilotBatchCFixture{
		profileRepository: profileRepository, draftRepository: draftRepository,
		publishRepository: publishRepository, runtimeRepository: newMemoryAgentCopilotRuntimeRepository(),
		runtimeContext: agentCopilotBatchCRuntimeContext(), boundDraft: *bound.Draft, publishService: publishService,
	}
	fixture.publishService.now = func() time.Time {
		fixture.nowSequence++
		return time.Date(2026, 7, 25, 14, fixture.nowSequence, 0, 0, time.UTC)
	}
	fixture.firstCandidate = fixture.createApprovedCandidate(t, "candidate-agent-copilot-1")
	fixture.runtimeService = newAgentCopilotRuntimeService(fixture.runtimeRepository, agentCopilotRuntimeAuthorityResolver{
		publishRepository: publishRepository, draftRepository: draftRepository,
		profileRepository: profileRepository, readApplication: agentCopilotBatchCBaseline,
	})
	fixture.runtimeService.now = func() time.Time {
		fixture.nowSequence++
		return time.Date(2026, 7, 25, 15, fixture.nowSequence, 0, 0, time.UTC)
	}
	var ids = map[string]int{}
	fixture.runtimeService.newID = func(prefix string) (string, error) {
		ids[prefix]++
		letter := "a"
		if ids[prefix] == 2 {
			letter = "b"
		} else if ids[prefix] == 3 {
			letter = "c"
		}
		return prefix + strings.Repeat(letter, 16), nil
	}
	return fixture
}

func (fixture *agentCopilotBatchCFixture) createApprovedCandidate(t *testing.T, candidateID string) ApplicationPublishCandidate {
	t.Helper()
	ctx := agentCopilotBatchCPublishContext()
	created := fixture.publishService.Create(ctx, ApplicationPublishCreateInput{
		CandidateID: candidateID, DraftID: fixture.boundDraft.DraftID,
		ExpectedDraftVersion: fixture.boundDraft.DraftVersion,
	})
	if created.FailureCode != "" || created.Candidate == nil {
		t.Fatalf("create Agent candidate %s: %#v", candidateID, created)
	}
	approved := fixture.publishService.Review(ctx, candidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove,
		Reason: "Reviewed exact Agent Copilot profile source and configuration lineage.",
	})
	if approved.FailureCode != "" || approved.Candidate == nil || approved.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("approve Agent candidate %s: %#v", candidateID, approved)
	}
	return *approved.Candidate
}

func createAgentCopilotBatchCProfileVersion(
	t *testing.T,
	repository *memoryAgentCopilotProfileRepository,
	ctx AgentCopilotProfileContext,
) AgentCopilotProfileVersionV1 {
	t.Helper()
	service := newAgentCopilotProfileService(repository)
	service.requireAgentApplication = func(AgentCopilotProfileContext) string { return "" }
	input := AgentCopilotProfileDraftInput{
		SchemaVersion: agentCopilotProfileDraftSchema, ProfileID: "acpf_aaaaaaaaaaaaaaaa",
		WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		AgentCopilotProfileSource: validAgentCopilotProfileSourceFixture(),
	}
	if saved := service.SaveDraft(ctx, input, 0); saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save Agent Copilot profile draft: %#v", saved)
	}
	version := service.CreateVersion(ctx, input.ProfileID, 1)
	if version.FailureCode != "" || version.Version == nil {
		t.Fatalf("create Agent Copilot profile version: %#v", version)
	}
	return *version.Version
}

func agentCopilotBatchCProfileContext() AgentCopilotProfileContext {
	return AgentCopilotProfileContext{
		RequestContext: context.Background(), RequestID: "request:agent-batch-c", TenantRef: "tenant:one",
		WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		ActorRef: "subject:owner", OwnerSubjectRef: "subject:owner",
		AuditRef: "audit:agent-batch-c", WriteEnabled: true,
	}
}

func agentCopilotBatchCDraftContext() ApplicationConfigurationDraftContext {
	return ApplicationConfigurationDraftContext{
		RequestContext: context.Background(), RequestID: "request:agent-batch-c", TenantRef: "tenant:one",
		WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		ActorRef: "subject:owner", OwnerSubjectRef: "subject:owner",
		AuditRef: "audit:agent-batch-c", WriteEnabled: true, ProfileBindingEnabled: true,
	}
}

func agentCopilotBatchCPublishContext() ApplicationPublishContext {
	return ApplicationPublishContext{
		RequestContext: context.Background(), RequestID: "request:agent-batch-c", TenantRef: "tenant:one",
		WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		ActorRef: "subject:owner", OwnerSubjectRef: "subject:owner",
		AuditRef: "audit:agent-batch-c", WriteEnabled: true, AgentProfileSourceReadEnabled: true,
	}
}

func agentCopilotBatchCRuntimeContext() AgentCopilotRuntimeContext {
	return AgentCopilotRuntimeContext{
		RequestContext: context.Background(), RequestID: "request:agent-batch-c", TenantRef: "tenant:one",
		WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		ActorRef: "subject:owner", OwnerSubjectRef: "subject:owner",
		AuditRef: "audit:agent-batch-c", WriteEnabled: true,
	}
}

func agentCopilotBatchCDraftPayload() ApplicationConfigurationDraftPayload {
	return ApplicationConfigurationDraftPayload{
		DraftID: "draft-agent-copilot", WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		BaseApplicationUpdatedAt: "2026-07-25T12:00:00Z", SchemaVersion: applicationConfigurationDraftSchemaVersionV1,
		DisplayName: "Agent Copilot", Description: "Advisory Agent Copilot application.",
		ApplicationKind: "agent", DefaultProtocol: "responses", DefaultModel: "profile:local-dev",
		AllowedProtocols: []string{"responses"},
	}
}

func agentCopilotBatchCBaseline(requestContext ApplicationPublishContext) (ApplicationSummary, error) {
	return ApplicationSummary{
		ApplicationRef: requestContext.ApplicationID, ApplicationKind: "agent",
		UpdatedAt: "2026-07-25T12:00:00Z",
	}, nil
}

func validAgentCopilotRuntimeMutation(
	t *testing.T,
	ctx AgentCopilotRuntimeContext,
) (AgentCopilotRuntimeAssignmentV1, AgentCopilotRuntimeAssignmentEventV1) {
	t.Helper()
	at := "2026-07-25T15:00:00Z"
	ref := AgentCopilotProfileRef{
		ProfileID: "acpf_aaaaaaaaaaaaaaaa", ProfileVersion: 1,
		ProfileDigest: "sha256:" + strings.Repeat("a", 64),
		PolicyDigest:  "sha256:" + strings.Repeat("b", 64),
	}
	assignment := AgentCopilotRuntimeAssignmentV1{
		SchemaVersion: agentCopilotRuntimeAssignmentSchema, AssignmentID: "acra_aaaaaaaaaaaaaaaa",
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AssignmentVersion: 1, State: "active",
		CandidateID: "candidate-agent-runtime", CandidateReviewVersion: 1,
		DraftID: "draft-agent-runtime", DraftVersion: 2,
		DraftDigest: "sha256:" + strings.Repeat("c", 64), AgentCopilotProfileRef: ref,
		ActivatedAt: at, UpdatedAt: at, ActivatedByActorRef: ctx.ActorRef,
		UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	digest, err := agentCopilotRuntimeAssignmentDigest(assignment)
	if err != nil {
		t.Fatalf("digest Agent Copilot assignment: %v", err)
	}
	assignment.AssignmentDigest = digest
	event := AgentCopilotRuntimeAssignmentEventV1{
		SchemaVersion: agentCopilotRuntimeAssignmentEventSchema,
		EventID:       "acrae_aaaaaaaaaaaaaaaa", AssignmentID: assignment.AssignmentID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, EventSequence: 1, Action: "activate",
		ExpectedAssignmentVersion: 0, ResultingAssignmentVersion: 1,
		CandidateID: assignment.CandidateID, CandidateReviewVersion: assignment.CandidateReviewVersion,
		DraftID: assignment.DraftID, DraftVersion: assignment.DraftVersion,
		DraftDigest: assignment.DraftDigest, AgentCopilotProfileRef: ref,
		AssignmentDigest: assignment.AssignmentDigest, OccurredAt: at,
		ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	return assignment, event
}

func errorsIsAgentCopilotRuntimeNotFound(err error) bool {
	return errorsIsAgentCopilotRuntime(err, errAgentCopilotRuntimeNotFound)
}

func errorsIsAgentCopilotRuntimeVersionConflict(err error) bool {
	return errorsIsAgentCopilotRuntime(err, errAgentCopilotRuntimeVersionConflict)
}

func errorsIsAgentCopilotRuntime(err, target error) bool {
	return err == target
}
