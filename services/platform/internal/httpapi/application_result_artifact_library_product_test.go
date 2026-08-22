package httpapi

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplicationResultArtifactLibraryDevFixtureOptionRequiresSQLiteSessionProductGates(t *testing.T) {
	cfg := aggregateSQLiteDevServerConfig(filepath.Join(t.TempDir(), "result-library-fixture-gate.db"))
	server, err := NewServerWithError(cfg, Options{
		BuildVersion: "sqlite-result-library-fixture-gate",
		ApplicationResultArtifactLibraryDevFixture: true,
	})
	if server != nil || err == nil || !strings.Contains(err.Error(), "requires the explicit SQLite local-product session gates") {
		if server != nil {
			server.Close()
		}
		t.Fatalf("result artifact library fixture accepted incomplete product gates: server=%v err=%v", server != nil, err)
	}
}

func TestApplicationResultArtifactLibraryDevFixtureSeedsMemoryRepositoriesIdempotently(t *testing.T) {
	server := &Server{
		applicationCatalogRepository:        newMemoryApplicationCatalogRepository(),
		applicationResultArtifactRepository: newMemoryApplicationResultArtifactRepository(),
	}
	fixture, err := seedApplicationResultArtifactLibraryDevFixture(server)
	if err != nil {
		t.Fatalf("seed memory result artifact library fixture: %v", err)
	}
	if _, err = seedApplicationResultArtifactLibraryDevFixture(server); err != nil {
		t.Fatalf("reseed memory result artifact library fixture: %v", err)
	}
	assertApplicationResultArtifactLibraryFixture(t, server, fixture)
}

func TestSQLiteApplicationResultArtifactLibraryDevFixtureRestartsWithoutResettingLifecycle(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "result-library-fixture.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	cfg.ApplicationSessionDevEnabled = true
	options := Options{
		BuildVersion: "sqlite-result-library-fixture-first",
		ApplicationResultArtifactLibraryDevFixture: true,
	}
	first, err := NewServerWithError(cfg, options)
	if err != nil {
		t.Fatalf("start SQLite result artifact library fixture server: %v", err)
	}
	fixture := applicationResultArtifactLibraryFixtureDefinition()
	assertApplicationResultArtifactLibraryFixture(t, first, fixture)

	service := first.applicationResultArtifactService()
	transitioned := service.Archive(fixture.ArtifactContext, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: fixture.Artifacts[0].SessionID, ArtifactID: fixture.Artifacts[0].ArtifactID, ExpectedLifecycleVersion: 1,
	})
	if transitioned.FailureCode != "" || transitioned.Lifecycle == nil || transitioned.Lifecycle.LifecycleVersion != 2 {
		first.Close()
		t.Fatalf("archive SQLite result artifact library fixture: %#v", transitioned)
	}
	restoredActive := service.Unarchive(fixture.ArtifactContext, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: fixture.Artifacts[0].SessionID, ArtifactID: fixture.Artifacts[0].ArtifactID, ExpectedLifecycleVersion: 2,
	})
	if restoredActive.FailureCode != "" || restoredActive.Lifecycle == nil ||
		restoredActive.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive ||
		restoredActive.Lifecycle.LifecycleVersion != 3 {
		first.Close()
		t.Fatalf("restore SQLite result artifact library fixture lifecycle: %#v", restoredActive)
	}
	var exportTableCount int
	if err = first.localPersistenceRuntime.DB().QueryRowContext(context.Background(),
		`SELECT count(*) FROM sqlite_master WHERE type='table' AND name LIKE 'application_result_artifact_export%'`,
	).Scan(&exportTableCount); err != nil || exportTableCount != 0 {
		first.Close()
		t.Fatalf("result artifact export unexpectedly persisted: count=%d err=%v", exportTableCount, err)
	}
	first.Close()

	options.BuildVersion = "sqlite-result-library-fixture-restarted"
	restarted, err := NewServerWithError(cfg, options)
	if err != nil {
		t.Fatalf("restart SQLite result artifact library fixture server: %v", err)
	}
	t.Cleanup(restarted.Close)
	assertApplicationResultArtifactLibraryFixture(t, restarted, fixture)
	read := restarted.applicationResultArtifactService().Read(fixture.ArtifactContext, fixture.Artifacts[0].ArtifactID)
	if read.FailureCode != "" || read.Artifact == nil || read.Lifecycle == nil ||
		read.Artifact.ContentDigest != fixture.Artifacts[0].ContentDigest ||
		read.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive || read.Lifecycle.LifecycleVersion != 3 {
		t.Fatalf("SQLite result artifact lifecycle was reset during fixture restart: %#v", read)
	}
}

func assertApplicationResultArtifactLibraryFixture(
	t *testing.T,
	server *Server,
	fixture applicationResultArtifactLibraryFixture,
) {
	t.Helper()
	catalog := newApplicationCatalogService(server.applicationCatalogRepository).Read(
		fixture.CatalogContext,
		fixture.Application.ApplicationID,
	)
	if catalog.FailureCode != "" || catalog.Record == nil ||
		catalog.Record.DisplayName != fixture.Application.DisplayName ||
		catalog.Record.LifecycleState != applicationCatalogLifecycleActive {
		t.Fatalf("result artifact library fixture application drifted: %#v", catalog)
	}

	service := server.applicationResultArtifactService()
	active := service.ListApplication(fixture.ArtifactContext, ApplicationResultArtifactListInput{})
	archived := service.ListApplication(fixture.ArtifactContext, ApplicationResultArtifactListInput{
		LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	promptJSON := service.ListApplication(fixture.ArtifactContext, ApplicationResultArtifactListInput{
		ExecutionProfile: applicationInteractionProfilePrompt, ContentType: "application/json",
	})
	workflowArchivedJSON := service.ListApplication(fixture.ArtifactContext, ApplicationResultArtifactListInput{
		LifecycleState:   ApplicationResultArtifactLifecycleArchived,
		ExecutionProfile: applicationInteractionProfileWorkflow, ContentType: "application/json",
	})
	if active.FailureCode != "" || len(active.Items) != 2 ||
		archived.FailureCode != "" || len(archived.Items) != 2 ||
		promptJSON.FailureCode != "" || len(promptJSON.Items) != 1 ||
		promptJSON.Items[0].ArtifactID != fixture.Artifacts[2].ArtifactID ||
		workflowArchivedJSON.FailureCode != "" || len(workflowArchivedJSON.Items) != 1 ||
		workflowArchivedJSON.Items[0].ArtifactID != fixture.Artifacts[1].ArtifactID {
		t.Fatalf("result artifact library fixture filters drifted: active=%#v archived=%#v prompt_json=%#v workflow_archived_json=%#v",
			active, archived, promptJSON, workflowArchivedJSON)
	}

	read := service.Read(fixture.ArtifactContext, fixture.Artifacts[2].ArtifactID)
	exported := service.Export(fixture.ArtifactContext, fixture.Artifacts[2].ArtifactID)
	if read.FailureCode != "" || read.Artifact == nil || read.Lifecycle == nil ||
		read.Artifact.SessionID != fixture.Artifacts[2].SessionID || read.Artifact.Content != fixture.Artifacts[2].Content ||
		exported.FailureCode != "" || exported.Export == nil ||
		exported.Export.Artifact.ContentDigest != fixture.Artifacts[2].ContentDigest ||
		exported.Export.ExportDigest != applicationResultArtifactExportDigest(*exported.Export) {
		t.Fatalf("result artifact library fixture exact read/export drifted: read=%#v export=%#v", read, exported)
	}
}
