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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	workflowdraftmigrations "radishmind.local/services/platform/migrations/workflow_saved_drafts"
)

func TestWorkflowTemplateCatalogPostgresDurabilityAtomicityCASAndCorruption(t *testing.T) {
	migrationDatabaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required for PostgreSQL integration tests")
	}
	if runtimeUser == strings.TrimSpace(os.Getenv("PGUSER")) {
		t.Fatal("PostgreSQL integration runtime user must differ from the migration user")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(
		t,
		runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	databaseContext, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	adminPool, err := workflowdraftmigrations.OpenPool(databaseContext, migrationDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL workflow template integration database: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, databaseContext, adminPool)
	resetPostgresSavedWorkflowDraftSchema(t, databaseContext, adminPool)
	preparePostgresIntegrationRuntimeRole(t, databaseContext, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresSavedWorkflowDraftSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	if _, err = workflowdraftmigrations.Apply(databaseContext, adminPool); err != nil {
		t.Fatalf("apply PostgreSQL workflow template migration: %v", err)
	}
	constrainPostgresWorkflowTemplateHistoryPrivileges(t, databaseContext, adminPool, runtimeUser)
	assertPostgresIntegrationRuntimeRoleCannotMigrate(t, databaseContext, adminPool, runtimeDatabaseURL)

	runtimePool, err := workflowdraftmigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL workflow template runtime pool: %v", err)
	}
	t.Cleanup(runtimePool.Close)
	repository, draftStore := newPostgresWorkflowTemplateIntegrationStores(runtimePool)
	fixture := newWorkflowTemplateTestFixture(t)
	fixture.ctx.RequestContext = databaseContext
	service := newWorkflowTemplateCatalogService(repository, fixture.definitions, fixture.applications, draftStore)
	service.now = func() time.Time { return time.Date(2026, 8, 27, 13, 0, 0, 123456000, time.UTC) }

	if result := service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput()); result.FailureCode != "" {
		portability, _ := validateWorkflowTemplatePortability(fixture.definition.Snapshot)
		_, repositoryErr := repository.CreateCandidate(fixture.ctx, WorkflowTemplateCandidate{
			CandidateID: workflowTemplateTestCandidate, TemplateID: workflowTemplateTestTemplate,
			SourceApplicationID: workflowTemplateTestSourceApplication, SourceOwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
			SourceDefinitionID: workflowTemplateTestDefinition, SourceDefinitionVersion: 1,
			SourceDefinitionDigest: fixture.definition.DefinitionDigest, Title: "团队问答工作流",
			Summary: "供工作区成员复用的受控问答流程。", UsageNotes: "派生后重新检查目标应用模型绑定。",
			Labels: []string{"team", "qa"}, Portability: portability,
		}, service.now())
		t.Fatalf("create PostgreSQL workflow template candidate: result=%#v repository_error=%v", result, repositoryErr)
	}
	revokePostgresWorkflowTemplatePrivilege(t, databaseContext, adminPool, runtimeUser, "INSERT", "workflow_template_versions")
	failedReview := service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "验证 PostgreSQL 评审事务回滚",
	})
	if failedReview.FailureCode != WorkflowTemplateFailureStoreUnavailable {
		t.Fatalf("PostgreSQL version insert failure did not fail closed: %#v", failedReview)
	}
	assertPostgresWorkflowTemplateCandidatePending(t, databaseContext, adminPool, workflowTemplateTestCandidate)
	grantPostgresWorkflowTemplatePrivilege(t, databaseContext, adminPool, runtimeUser, "INSERT", "workflow_template_versions")
	reviewed := service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准 PostgreSQL 模板候选",
	})
	if reviewed.FailureCode != "" || reviewed.Version == nil {
		t.Fatalf("approve PostgreSQL workflow template candidate: %#v", reviewed)
	}

	revokePostgresWorkflowTemplatePrivilege(t, databaseContext, adminPool, runtimeUser, "INSERT", "workflow_template_listing_events")
	failedListing := service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "验证 PostgreSQL 上架事务回滚",
	})
	if failedListing.FailureCode != WorkflowTemplateFailureStoreUnavailable {
		t.Fatalf("PostgreSQL listing event failure did not fail closed: %#v", failedListing)
	}
	assertPostgresWorkflowTemplateLineageUnlisted(t, databaseContext, adminPool, workflowTemplateTestTemplate)
	grantPostgresWorkflowTemplatePrivilege(t, databaseContext, adminPool, runtimeUser, "INSERT", "workflow_template_listing_events")
	listed := service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架 PostgreSQL 模板版本",
	})
	if listed.FailureCode != "" || listed.Lineage == nil {
		t.Fatalf("list PostgreSQL workflow template version: %#v", listed)
	}
	derived := service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
		ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
		DraftID: "draft_from_postgres_template", Name: "PostgreSQL 模板派生草案", Confirmed: true,
	})
	if derived.FailureCode != "" || derived.Draft == nil || derived.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("derive Saved Draft through PostgreSQL owner: %#v", derived)
	}

	assertPostgresWorkflowTemplateCAS(t, service, fixture.ctx)
	assertPostgresWorkflowTemplateHistoryIsAppendOnly(t, databaseContext, runtimePool)

	runtimePool.Close()
	if _, err := repository.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate); !errors.Is(err, errWorkflowTemplateStoreUnavailable) {
		t.Fatalf("closed PostgreSQL catalog fell back instead of failing: %v", err)
	}
	reconnectedPool, err := workflowdraftmigrations.OpenPool(databaseContext, runtimeDatabaseURL)
	if err != nil {
		t.Fatalf("reconnect PostgreSQL workflow template runtime pool: %v", err)
	}
	defer reconnectedPool.Close()
	reconnectedRepository, reconnectedDraftStore := newPostgresWorkflowTemplateIntegrationStores(reconnectedPool)
	lineage, err := reconnectedRepository.ReadLineage(fixture.ctx, workflowTemplateTestTemplate)
	if err != nil || lineage.PointerVersion != 1 || lineage.ListedDigest != reviewed.Version.TemplateDigest {
		t.Fatalf("restore PostgreSQL workflow template lineage: lineage=%#v err=%v", lineage, err)
	}
	draftService := newSavedWorkflowDraftService(reconnectedDraftStore)
	draft := draftService.ReadDraft(SavedWorkflowDraftContext{
		RequestContext: databaseContext, RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef,
		WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: workflowTemplateTestTargetApplication,
		ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef,
		ScopeGrants: []string{"workflow_drafts:read"}, AuditRef: fixture.ctx.AuditRef,
	}, ReadWorkflowDraftRequest{DraftID: "draft_from_postgres_template"})
	if draft.FailureCode != "" || draft.Draft == nil || draft.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("restore PostgreSQL derived Saved Draft: %#v", draft)
	}
	if _, err = adminPool.Exec(databaseContext, `UPDATE workflow_template_candidates
SET updated_at=updated_at+interval '1 microsecond' WHERE candidate_id=$1`, workflowTemplateTestCandidate); err != nil {
		t.Fatalf("prepare PostgreSQL workflow template corruption: %v", err)
	}
	if _, err = reconnectedRepository.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate); !errors.Is(err, errWorkflowTemplateStoreUnavailable) {
		t.Fatalf("PostgreSQL projection corruption did not fail closed: %v", err)
	}
	rollbackState, err := workflowdraftmigrations.RollbackForDevTest(databaseContext, adminPool)
	if err != nil || rollbackState.MigrationState != workflowdraftmigrations.MigrationStateNotApplied {
		t.Fatalf("rollback PostgreSQL workflow template migration after derivation: state=%#v err=%v", rollbackState, err)
	}
	reappliedState, err := workflowdraftmigrations.Apply(databaseContext, adminPool)
	if err != nil || reappliedState.MigrationState != workflowdraftmigrations.MigrationStateApplied {
		t.Fatalf("reapply PostgreSQL workflow template migration: state=%#v err=%v", reappliedState, err)
	}
}

func newPostgresWorkflowTemplateIntegrationStores(pool *pgxpool.Pool) (*postgresWorkflowTemplateCatalogRepository, savedWorkflowDraftStore) {
	executor := newPostgresSavedWorkflowDraftQueryExecutor(pool)
	adapter := NewSavedWorkflowDraftRepositoryAdapter(SavedWorkflowDraftRepositoryAdapterConfig{
		QueryExecutor: executor,
		SchemaPreflight: SavedWorkflowDraftRepositorySchemaPreflight{
			StoreSchemaVersion: savedWorkflowDraftRepositoryStoreSchemaVersion,
			MigrationState:     savedWorkflowDraftRepositoryMigrationApplied,
		},
	})
	draftStore := newRepositorySavedWorkflowDraftLibraryStore(newRepositorySavedWorkflowDraftStore(adapter))
	return newPostgresWorkflowTemplateCatalogRepository(pool), draftStore
}

func assertPostgresWorkflowTemplateCAS(t *testing.T, service workflowTemplateCatalogService, ctx WorkflowTemplateCatalogContext) {
	t.Helper()
	input := workflowTemplateCandidateTestInput()
	input.CandidateID = "candidate_postgres_cas"
	input.TemplateID = "template_postgres_cas"
	if result := service.CreateCandidate(ctx, input); result.FailureCode != "" {
		t.Fatalf("create PostgreSQL CAS candidate: %#v", result)
	}
	const writers = 8
	var group sync.WaitGroup
	reviewResults := make(chan WorkflowTemplateCatalogResult, writers)
	for range writers {
		group.Add(1)
		go func() {
			defer group.Done()
			reviewResults <- service.ReviewCandidate(ctx, input.CandidateID, WorkflowTemplateReviewInput{
				ExpectedReviewVersion: 0, Decision: "approve", Reason: "并发批准 PostgreSQL 模板候选",
			})
		}()
	}
	group.Wait()
	close(reviewResults)
	assertWorkflowTemplateSingleCASWinner(t, reviewResults, WorkflowTemplateFailureCandidateVersionConflict, WorkflowTemplateFailureReviewTransitionInvalid)

	listingResults := make(chan WorkflowTemplateCatalogResult, writers)
	for range writers {
		group.Add(1)
		go func() {
			defer group.Done()
			listingResults <- service.DecideListing(ctx, input.TemplateID, WorkflowTemplateListingInput{
				ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "并发上架 PostgreSQL 模板版本",
			})
		}()
	}
	group.Wait()
	close(listingResults)
	assertWorkflowTemplateSingleCASWinner(t, listingResults, WorkflowTemplateFailurePointerVersionConflict)
}

func assertWorkflowTemplateSingleCASWinner(t *testing.T, results <-chan WorkflowTemplateCatalogResult, allowedFailures ...string) {
	t.Helper()
	successes := 0
	for result := range results {
		if result.FailureCode == "" {
			successes++
			continue
		}
		allowed := false
		for _, failure := range allowedFailures {
			allowed = allowed || result.FailureCode == failure
		}
		if !allowed {
			t.Fatalf("unexpected PostgreSQL CAS result: %#v", result)
		}
	}
	if successes != 1 {
		t.Fatalf("PostgreSQL CAS winners drifted: %d", successes)
	}
}

func constrainPostgresWorkflowTemplateHistoryPrivileges(t *testing.T, ctx context.Context, pool *pgxpool.Pool, runtimeUser string) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	for _, table := range []string{
		"workflow_template_decisions",
		"workflow_template_versions",
		"workflow_template_listing_events",
		"workflow_template_audits",
	} {
		if _, err := pool.Exec(ctx, "REVOKE ALL ON TABLE "+table+" FROM "+quotedRuntimeUser); err != nil {
			t.Fatalf("revoke PostgreSQL workflow template history privileges: %v", err)
		}
		if _, err := pool.Exec(ctx, "GRANT SELECT, INSERT ON TABLE "+table+" TO "+quotedRuntimeUser); err != nil {
			t.Fatalf("grant PostgreSQL workflow template append-only privileges: %v", err)
		}
	}
}

func revokePostgresWorkflowTemplatePrivilege(t *testing.T, ctx context.Context, pool *pgxpool.Pool, runtimeUser, privilege, table string) {
	t.Helper()
	if _, err := pool.Exec(ctx, "REVOKE "+privilege+" ON TABLE "+pgx.Identifier{table}.Sanitize()+" FROM "+pgx.Identifier{runtimeUser}.Sanitize()); err != nil {
		t.Fatalf("revoke PostgreSQL workflow template privilege: %v", err)
	}
}

func grantPostgresWorkflowTemplatePrivilege(t *testing.T, ctx context.Context, pool *pgxpool.Pool, runtimeUser, privilege, table string) {
	t.Helper()
	if _, err := pool.Exec(ctx, "GRANT "+privilege+" ON TABLE "+pgx.Identifier{table}.Sanitize()+" TO "+pgx.Identifier{runtimeUser}.Sanitize()); err != nil {
		t.Fatalf("grant PostgreSQL workflow template privilege: %v", err)
	}
}

func assertPostgresWorkflowTemplateCandidatePending(t *testing.T, ctx context.Context, pool *pgxpool.Pool, candidateID string) {
	t.Helper()
	var state string
	var reviewVersion, decisions, versions, lineages int
	if err := pool.QueryRow(ctx, `SELECT candidate_state,review_version,
(SELECT count(*) FROM workflow_template_decisions WHERE candidate_id=$1),
(SELECT count(*) FROM workflow_template_versions WHERE candidate_id=$1),
(SELECT count(*) FROM workflow_template_lineages WHERE template_id=$2)
FROM workflow_template_candidates WHERE candidate_id=$1`, candidateID, workflowTemplateTestTemplate).Scan(&state, &reviewVersion, &decisions, &versions, &lineages); err != nil {
		t.Fatalf("inspect PostgreSQL failed review rollback: %v", err)
	}
	if state != workflowTemplateCandidatePending || reviewVersion != 0 || decisions != 0 || versions != 0 || lineages != 0 {
		t.Fatalf("PostgreSQL failed review committed partial state: state=%s review=%d decisions=%d versions=%d lineages=%d", state, reviewVersion, decisions, versions, lineages)
	}
}

func assertPostgresWorkflowTemplateLineageUnlisted(t *testing.T, ctx context.Context, pool *pgxpool.Pool, templateID string) {
	t.Helper()
	var lifecycle string
	var pointerVersion, eventCount int
	if err := pool.QueryRow(ctx, `SELECT lifecycle,pointer_version,
(SELECT count(*) FROM workflow_template_listing_events WHERE template_id=$1)
FROM workflow_template_lineages WHERE template_id=$1`, templateID).Scan(&lifecycle, &pointerVersion, &eventCount); err != nil {
		t.Fatalf("inspect PostgreSQL failed listing rollback: %v", err)
	}
	if lifecycle != workflowTemplateLineageUnlisted || pointerVersion != 0 || eventCount != 0 {
		t.Fatalf("PostgreSQL failed listing committed partial state: lifecycle=%s pointer=%d events=%d", lifecycle, pointerVersion, eventCount)
	}
}

func assertPostgresWorkflowTemplateHistoryIsAppendOnly(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, statement := range []string{
		"UPDATE workflow_template_decisions SET decision=decision",
		"DELETE FROM workflow_template_versions",
		"UPDATE workflow_template_listing_events SET decision=decision",
		"DELETE FROM workflow_template_audits",
	} {
		if _, err := pool.Exec(ctx, statement); err == nil {
			t.Fatalf("PostgreSQL runtime role unexpectedly mutated workflow template history: %s", statement)
		} else {
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
				t.Fatalf("workflow template append-only denial returned unexpected error: %v", err)
			}
		}
	}
}
