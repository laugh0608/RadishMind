//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostgresConfiguredApplicationResultArtifactProductChainRestartsWithoutFallback(t *testing.T) {
	adminPool, runtimeDatabaseURL, ctx := newConfiguredPostgresTestDatabase(t)
	for _, gate := range configuredPostgresMigrationGates(adminPool) {
		state, _, err := gate.apply(ctx)
		if err != nil || state != "applied" {
			t.Fatalf("apply %s migration for result artifact product chain: state=%s err=%v", gate.name, state, err)
		}
	}

	cfg := configuredPostgresProductConfig(runtimeDatabaseURL)
	cfg.ApplicationSessionDevEnabled = true
	server, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-result-artifact-first"})
	if err != nil {
		t.Fatalf("start configured PostgreSQL result artifact platform: %v", err)
	}
	assertConfiguredPostgresRepositorySelection(t, server)
	if repository, ok := server.applicationResultArtifactRepository.(*postgresApplicationResultArtifactRepository); !ok ||
		repository.pool != server.workflowRunStore.(*postgresWorkflowRunStore).pool {
		server.Close()
		t.Fatalf("configured result artifact store did not share the PostgreSQL pool: %T", server.applicationResultArtifactRepository)
	}

	artifactContext, session, turn := createConfiguredPostgresResultArtifactSource(t, ctx, server)
	artifactService := server.applicationResultArtifactService()
	artifactService.newID = func(string) (string, error) { return "appres_dddddddddddddddd", nil }
	artifactService.now = func() time.Time { return time.Date(2026, 8, 17, 12, 0, 2, 123456000, time.UTC) }
	content := "# PostgreSQL product result\n\nRestart-safe explicit user-owned content."
	captured := artifactService.Capture(artifactContext, ApplicationResultArtifactCaptureInput{
		Turn: turn, ContentType: "text/markdown", Content: content,
	})
	if captured.FailureCode != "" || captured.Artifact == nil || captured.Summary == nil || captured.Lifecycle == nil ||
		captured.IdempotentReplay || captured.Artifact.SessionID != session.SessionID ||
		captured.Artifact.Content != content || captured.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive {
		server.Close()
		t.Fatalf("capture configured PostgreSQL result artifact: %#v", captured)
	}
	replayed := artifactService.Capture(artifactContext, ApplicationResultArtifactCaptureInput{
		Turn: turn, ContentType: "text/markdown", Content: content,
	})
	if replayed.FailureCode != "" || replayed.Artifact == nil || !replayed.IdempotentReplay ||
		replayed.Artifact.ArtifactID != captured.Artifact.ArtifactID {
		server.Close()
		t.Fatalf("configured PostgreSQL result artifact idempotency drifted: %#v", replayed)
	}
	active := artifactService.List(artifactContext, ApplicationResultArtifactListInput{SessionID: session.SessionID, Limit: 1})
	if active.FailureCode != "" || len(active.Items) != 1 || active.NextCursor != nil ||
		active.Items[0].ArtifactID != captured.Artifact.ArtifactID || active.Items[0].LifecycleState != ApplicationResultArtifactLifecycleActive {
		server.Close()
		t.Fatalf("list configured PostgreSQL active result artifact: %#v", active)
	}
	archived := artifactService.Archive(artifactContext, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: session.SessionID, ArtifactID: captured.Artifact.ArtifactID, ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" || archived.Lifecycle == nil || archived.Event == nil ||
		archived.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived || archived.Lifecycle.LifecycleVersion != 2 {
		server.Close()
		t.Fatalf("archive configured PostgreSQL result artifact: %#v", archived)
	}
	assertApplicationResultArtifactPurgeRouteAbsent(t, server, session.SessionID, captured.Artifact.ArtifactID)

	closedRepository := server.applicationResultArtifactRepository
	server.Close()
	if _, err = closedRepository.Read(artifactContext, captured.Artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactStore) {
		t.Fatalf("closed configured PostgreSQL result artifact store did not fail closed: %v", err)
	}

	restarted, err := NewServerWithError(cfg, Options{BuildVersion: "postgres-result-artifact-restarted"})
	if err != nil {
		t.Fatalf("restart configured PostgreSQL result artifact platform: %v", err)
	}
	t.Cleanup(restarted.Close)
	assertConfiguredPostgresRepositorySelection(t, restarted)
	restartedService := restarted.applicationResultArtifactService()
	restored := restartedService.Read(artifactContext, captured.Artifact.ArtifactID)
	archivedPage := restartedService.List(artifactContext, ApplicationResultArtifactListInput{
		SessionID: session.SessionID, LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	if restored.FailureCode != "" || restored.Artifact == nil || restored.Lifecycle == nil ||
		restored.Artifact.Content != content || restored.Artifact.ContentDigest != captured.Artifact.ContentDigest ||
		restored.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived || restored.Lifecycle.LifecycleVersion != 2 ||
		archivedPage.FailureCode != "" || len(archivedPage.Items) != 1 || archivedPage.Items[0].ArtifactID != captured.Artifact.ArtifactID {
		t.Fatalf("restore configured PostgreSQL result artifact after restart: restored=%#v archived=%#v", restored, archivedPage)
	}
	unarchived := restartedService.Unarchive(artifactContext, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: session.SessionID, ArtifactID: captured.Artifact.ArtifactID, ExpectedLifecycleVersion: 2,
	})
	if unarchived.FailureCode != "" || unarchived.Lifecycle == nil || unarchived.Event == nil ||
		unarchived.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive || unarchived.Lifecycle.LifecycleVersion != 3 {
		t.Fatalf("unarchive restored configured PostgreSQL result artifact: %#v", unarchived)
	}
	active = restartedService.List(artifactContext, ApplicationResultArtifactListInput{SessionID: session.SessionID})
	if active.FailureCode != "" || len(active.Items) != 1 || active.Items[0].LifecycleVersion != 3 {
		t.Fatalf("restored artifact did not return to the active list: %#v", active)
	}
}

func createConfiguredPostgresResultArtifactSource(
	t *testing.T,
	requestContext context.Context,
	server *Server,
) (ApplicationInteractionContext, ApplicationInteractionSession, ApplicationInteractionTurn) {
	t.Helper()
	fixtureService, fixtureContext, binding, _ := applicationInteractionSessionTestFixture(t)
	fixtureService.newID = func(string) (string, error) { return "appsess_dddddddddddddddd", nil }
	fixtureService.now = func() time.Time { return time.Date(2026, 8, 17, 12, 0, 0, 123456000, time.UTC) }
	fixture := fixtureService.Create(fixtureContext, ApplicationInteractionSessionCreateInput{ProfileBinding: binding})
	if fixture.FailureCode != "" || fixture.Session == nil {
		t.Fatalf("create valid result artifact session fixture: %#v", fixture)
	}

	context := fixtureContext
	context.RequestContext = requestContext
	context.RequestID = "request_postgres_result_artifact"
	context.ApplicationID = "app_dddddddddddddddd"
	context.ActorRef = "subject_platform_ops"
	context.OwnerSubjectRef = "subject_platform_ops"
	context.AuditRef = "audit_postgres_result_artifact"
	applicationContext := applicationCatalogTestContext(context.OwnerSubjectRef)
	applicationContext.RequestContext = requestContext
	catalogService := newApplicationCatalogService(server.applicationCatalogRepository)
	catalogService.newID = func() (string, error) { return context.ApplicationID, nil }
	catalogService.now = func() time.Time { return time.Date(2026, 8, 17, 11, 59, 59, 123456000, time.UTC) }
	application := catalogService.Create(applicationContext, ApplicationCatalogCreateInput{
		DisplayName:     "Result artifact PostgreSQL product chain",
		Description:     "Configured PostgreSQL result artifact restart evidence.",
		ApplicationKind: "workflow_copilot",
	})
	if application.FailureCode != "" || application.Record == nil {
		t.Fatalf("create result artifact product application: %#v", application)
	}
	context.TenantRef = application.Record.TenantRef
	context.WorkspaceID = application.Record.WorkspaceID

	session := *fixture.Session
	session.TenantRef = context.TenantRef
	session.WorkspaceID = context.WorkspaceID
	session.ApplicationID = context.ApplicationID
	session.OwnerSubjectRef = context.OwnerSubjectRef
	session.Authority.ApplicationID = context.ApplicationID
	session.Authority.ApplicationRecordVersion = application.Record.RecordVersion
	session.Authority.ApplicationLifecycle = applicationCatalogLifecycleActive
	session.Authority.AuthorityDigest = ""
	digest, err := applicationInteractionAuthorityDigest(session.Authority)
	if err != nil {
		t.Fatalf("digest configured PostgreSQL result artifact authority: %v", err)
	}
	session.Authority.AuthorityDigest = digest
	session.CreatedByActorRef = context.ActorRef
	session.UpdatedByActorRef = context.ActorRef
	session.RequestID = context.RequestID
	session.AuditRef = context.AuditRef
	storedSession, err := server.applicationInteractionSessionRepository.Create(context, session)
	if err != nil {
		t.Fatalf("store configured PostgreSQL result artifact session: %v", err)
	}

	resolver := applicationInteractionAuthorityResolverFunc(func(
		ApplicationInteractionContext,
		ApplicationInteractionProfileBinding,
	) (ApplicationInteractionAuthoritySnapshot, string) {
		return storedSession.Authority, ""
	})
	sessionService := newApplicationInteractionSessionService(server.applicationInteractionSessionRepository, resolver)
	sessionService.newID = func(string) (string, error) { return "appturn_dddddddddddddddd", nil }
	turnResult := sessionService.AppendTerminalTurn(context, storedSession.SessionID, ApplicationInteractionTerminalTurnInput{
		ExpectedSessionVersion: 1,
		ClientTurnKey:          "postgres_result_turn_1",
		Status:                 string(WorkflowRunStatusSucceeded),
		InputDigest:            workflowDefinitionInputDigest("private PostgreSQL product input"),
		InputBytes:             len("private PostgreSQL product input"),
		RunRef: &ApplicationInteractionRunRef{
			RunID: "run_dddddddddddddddd", SchemaVersion: workflowRunRecordDefinitionSchemaVersion,
		},
		StartedAt:   time.Date(2026, 8, 17, 12, 0, 1, 123456000, time.UTC),
		CompletedAt: time.Date(2026, 8, 17, 12, 0, 2, 123456000, time.UTC),
	})
	if turnResult.FailureCode != "" || turnResult.Turn == nil || turnResult.Turn.Status != string(WorkflowRunStatusSucceeded) {
		t.Fatalf("store configured PostgreSQL result artifact source turn: %#v", turnResult)
	}
	return context, storedSession, *turnResult.Turn
}

func assertApplicationResultArtifactPurgeRouteAbsent(
	t *testing.T,
	server *Server,
	sessionID string,
	artifactID string,
) {
	t.Helper()
	request := httptest.NewRequest(http.MethodDelete,
		"/v1/user-workspace/application-sessions/"+sessionID+"/result-artifacts/"+artifactID, nil)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("result artifact purge route unexpectedly exists: status=%d body=%s", response.Code, response.Body.String())
	}
}
