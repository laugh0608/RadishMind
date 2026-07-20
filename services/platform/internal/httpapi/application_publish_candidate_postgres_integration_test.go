//go:build postgres_integration

package httpapi

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	applicationdraftmigrations "radishmind.local/services/platform/migrations/application_configuration_drafts"
	applicationpublishmigrations "radishmind.local/services/platform/migrations/application_publish_candidates"
)

func TestApplicationPublishCandidatePostgresLifecycle(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	migrationPool, err := applicationpublishmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open application publish migration pool: %v", err)
	}
	defer migrationPool.Close()
	draftMigrationPool, err := applicationdraftmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open application draft migration pool: %v", err)
	}
	defer draftMigrationPool.Close()
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, migrationPool)
	if _, err := applicationpublishmigrations.RollbackForDevTest(ctx, migrationPool); err != nil {
		t.Fatalf("reset application publish migration: %v", err)
	}
	if _, err := applicationdraftmigrations.RollbackForDevTest(ctx, draftMigrationPool); err != nil {
		t.Fatalf("reset application draft migration: %v", err)
	}
	if state, err := applicationdraftmigrations.Apply(ctx, draftMigrationPool); err != nil || state.MigrationState != applicationdraftmigrations.MigrationStateApplied {
		t.Fatalf("apply application draft migration: %#v %v", state, err)
	}
	if state, err := applicationpublishmigrations.Apply(ctx, migrationPool); err != nil || state.MigrationState != applicationpublishmigrations.MigrationStateApplied {
		t.Fatalf("apply application publish migration: %#v %v", state, err)
	}
	runtimePool, err := applicationpublishmigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open application publish runtime pool: %v", err)
	}
	defer runtimePool.Close()
	defer func() {
		_, _ = applicationpublishmigrations.RollbackForDevTest(context.Background(), migrationPool)
		_, _ = applicationdraftmigrations.RollbackForDevTest(context.Background(), draftMigrationPool)
	}()

	draftContext := validApplicationDraftContext()
	draftRepository := newPostgresApplicationConfigurationDraftRepository(runtimePool)
	createdDraft := newApplicationConfigurationDraftService(draftRepository).Save(draftContext, validApplicationDraftPayload(), 0)
	if createdDraft.Draft == nil {
		t.Fatalf("create bound application draft: %#v", createdDraft)
	}
	requestContext := validApplicationPublishContext()
	repository := newPostgresApplicationPublishCandidateRepository(runtimePool)
	service := newApplicationPublishCandidateService(draftRepository, repository, validApplicationPublishBaseline)
	created := service.Create(requestContext, ApplicationPublishCreateInput{CandidateID: "candidate-postgres-v1", DraftID: validApplicationDraftPayload().DraftID, ExpectedDraftVersion: 1})
	if created.Candidate == nil || created.Candidate.ReviewVersion != 0 {
		t.Fatalf("create candidate in PostgreSQL: %#v", created)
	}
	restarted := newApplicationPublishCandidateService(draftRepository, newPostgresApplicationPublishCandidateRepository(runtimePool), validApplicationPublishBaseline)
	restored := restarted.Read(requestContext, "candidate-postgres-v1")
	if restored.Candidate == nil || restored.Candidate.DraftDigest != created.Candidate.DraftDigest {
		t.Fatalf("restore candidate after service reconstruction: %#v", restored)
	}
	approved := restarted.Review(requestContext, "candidate-postgres-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "PostgreSQL candidate review accepted for development verification."})
	if approved.Candidate == nil || approved.Candidate.ReviewVersion != 1 || approved.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("append candidate review: %#v", approved)
	}
	conflict := restarted.Review(requestContext, "candidate-postgres-v1", ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: "reject", Reason: "Stale review must not replace the accepted decision."})
	if conflict.FailureCode != ApplicationPublishFailureReviewVersionConflict || conflict.CurrentReviewVersion != 1 {
		t.Fatalf("PostgreSQL review CAS conflict failed: %#v", conflict)
	}
	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if denied := restarted.Read(otherOwner, "candidate-postgres-v1"); denied.FailureCode != ApplicationPublishFailureNotFound {
		t.Fatalf("PostgreSQL candidate owner isolation failed: %#v", denied)
	}
	if _, err := runtimePool.Exec(ctx, "CREATE TABLE application_publish_runtime_ddl_denied(id integer)"); err == nil {
		t.Fatal("runtime role unexpectedly received DDL permission")
	}

	binding := testWorkflowRAGApplicationBindingForDraft(*createdDraft.Draft)
	boundDraftContext := draftContext
	boundDraftContext.BindingEnabled = true
	boundDraftService := newApplicationConfigurationDraftService(draftRepository)
	boundDraftService.validateBinding = func(ApplicationConfigurationDraftContext, WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return binding, ""
	}
	boundPayload := createdDraft.Draft.ApplicationConfigurationDraftPayload
	boundPayload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
	boundPayload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&binding.WorkflowRAGApplicationBindingRef)
	boundDraft := boundDraftService.Save(boundDraftContext, boundPayload, 1)
	if boundDraft.Draft == nil || boundDraft.Draft.DraftVersion != 2 {
		t.Fatalf("save PostgreSQL v2 bound draft without DDL: %#v", boundDraft)
	}
	boundPublishContext := requestContext
	boundPublishContext.RAGPromotionReadEnabled = true
	boundPublishService := newApplicationPublishCandidateService(draftRepository, repository, validApplicationPublishBaseline)
	boundPublishService.validateBinding = func(ApplicationPublishContext, WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return binding, ""
	}
	boundCandidate := boundPublishService.Create(boundPublishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-postgres-v2", DraftID: boundDraft.Draft.DraftID, ExpectedDraftVersion: 2,
	})
	if boundCandidate.Candidate == nil || boundCandidate.Candidate.SchemaVersion != applicationPublishCandidateSchemaVersionV2 || boundCandidate.Candidate.Configuration.WorkflowRAGBindingRef == nil {
		t.Fatalf("save PostgreSQL v2 publish candidate without DDL: %#v", boundCandidate)
	}
	restoredBoundDraft := newApplicationConfigurationDraftService(newPostgresApplicationConfigurationDraftRepository(runtimePool)).Read(boundDraftContext, boundDraft.Draft.DraftID)
	restoredBoundCandidate := newApplicationPublishCandidateService(draftRepository, newPostgresApplicationPublishCandidateRepository(runtimePool), validApplicationPublishBaseline).Read(boundPublishContext, boundCandidate.Candidate.CandidateID)
	if restoredBoundDraft.Draft == nil || restoredBoundDraft.Draft.WorkflowRAGBindingRef == nil || restoredBoundCandidate.Candidate == nil || restoredBoundCandidate.Candidate.Configuration.WorkflowRAGBindingRef == nil {
		t.Fatalf("restore PostgreSQL v2 JSON payloads: draft=%#v candidate=%#v", restoredBoundDraft, restoredBoundCandidate)
	}

	if _, err := applicationpublishmigrations.RollbackForDevTest(ctx, migrationPool); err != nil {
		t.Fatalf("rollback application publish migration: %v", err)
	}
	if state, err := applicationpublishmigrations.Apply(ctx, migrationPool); err != nil || state.MigrationState != applicationpublishmigrations.MigrationStateApplied {
		t.Fatalf("reapply application publish migration: %#v %v", state, err)
	}
}
