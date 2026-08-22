package httpapi

import (
	"errors"
	"fmt"
	"time"
)

const applicationResultArtifactLibraryFixtureApplicationID = "app_rrrrrrrrrrrrrrrr"

type applicationResultArtifactLibraryFixture struct {
	CatalogContext               ApplicationCatalogContext
	ArtifactContext              ApplicationInteractionContext
	Application                  ApplicationCatalogRecord
	Artifacts                    []ApplicationResultArtifact
	InitiallyArchivedArtifactIDs map[string]struct{}
}

func applicationResultArtifactLibraryFixtureDefinition() applicationResultArtifactLibraryFixture {
	const (
		tenantRef       = "tenant_demo"
		workspaceID     = "workspace_demo"
		ownerSubjectRef = "subject_demo_user"
		requestID       = "request_result_library_fixture"
		auditRef        = "audit_result_library_fixture"
	)
	createdAt := time.Date(2026, 8, 17, 10, 0, 0, 123456000, time.UTC)
	catalogContext := ApplicationCatalogContext{
		RequestID: requestID, TenantRef: tenantRef, WorkspaceID: workspaceID,
		ActorRef: ownerSubjectRef, OwnerSubjectRef: ownerSubjectRef, AuditRef: auditRef, WriteEnabled: true,
	}
	artifactContext := ApplicationInteractionContext{
		RequestID: requestID, TenantRef: tenantRef, WorkspaceID: workspaceID,
		ApplicationID: applicationResultArtifactLibraryFixtureApplicationID,
		ActorRef:      ownerSubjectRef, OwnerSubjectRef: ownerSubjectRef, AuditRef: auditRef, WriteEnabled: true,
	}
	application := ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion,
		ApplicationID: applicationResultArtifactLibraryFixtureApplicationID,
		TenantRef:     tenantRef, WorkspaceID: workspaceID, OwnerSubjectRef: ownerSubjectRef,
		DisplayName:     "Result Library Fixture",
		Description:     "Stable SQLite development fixture for result artifact library validation.",
		ApplicationKind: "workflow_copilot", LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1,
		CreatedAt:         createdAt.Add(-time.Minute).Format(time.RFC3339Nano),
		UpdatedAt:         createdAt.Add(-time.Minute).Format(time.RFC3339Nano),
		CreatedByActorRef: ownerSubjectRef, UpdatedByActorRef: ownerSubjectRef,
		RequestID: requestID, AuditRef: auditRef,
	}
	artifacts := []ApplicationResultArtifact{
		newApplicationResultArtifactLibraryFixtureArtifact(
			artifactContext, "appres_rrrrrrrrrrrrrrrr", "appsess_rrrrrrrrrrrrrrrr", "appturn_rrrrrrrrrrrrrrrr",
			"result_library_fixture_1", applicationInteractionProfileWorkflow,
			"run_rrrrrrrrrrrrrrrr", workflowRunRecordDefinitionSchemaVersion,
			"text/markdown", "# Stable workflow result\n\nSQLite exact-read and export evidence.", createdAt,
		),
		newApplicationResultArtifactLibraryFixtureArtifact(
			artifactContext, "appres_ssssssssssssssss", "appsess_rrrrrrrrrrrrrrrr", "appturn_ssssssssssssssss",
			"result_library_fixture_2", applicationInteractionProfileWorkflow,
			"run_ssssssssssssssss", workflowRunRecordDefinitionSchemaVersion,
			"application/json", `{"fixture":"workflow","state":"archived"}`, createdAt.Add(time.Second),
		),
		newApplicationResultArtifactLibraryFixtureArtifact(
			artifactContext, "appres_tttttttttttttttt", "appsess_ssssssssssssssss", "appturn_tttttttttttttttt",
			"result_library_fixture_3", applicationInteractionProfilePrompt,
			"run_tttttttttttttttt", workflowRunRecordPromptSchemaVersion,
			"application/json", `{"fixture":"prompt","state":"active"}`, createdAt.Add(2*time.Second),
		),
		newApplicationResultArtifactLibraryFixtureArtifact(
			artifactContext, "appres_uuuuuuuuuuuuuuuu", "appsess_ssssssssssssssss", "appturn_uuuuuuuuuuuuuuuu",
			"result_library_fixture_4", applicationInteractionProfilePrompt,
			"run_uuuuuuuuuuuuuuuu", workflowRunRecordPromptSchemaVersion,
			"text/markdown", "# Archived prompt result\n\nStable lifecycle filter evidence.", createdAt.Add(3*time.Second),
		),
	}
	return applicationResultArtifactLibraryFixture{
		CatalogContext: catalogContext, ArtifactContext: artifactContext, Application: application, Artifacts: artifacts,
		InitiallyArchivedArtifactIDs: map[string]struct{}{
			"appres_ssssssssssssssss": {},
			"appres_uuuuuuuuuuuuuuuu": {},
		},
	}
}

func newApplicationResultArtifactLibraryFixtureArtifact(
	ctx ApplicationInteractionContext,
	artifactID string,
	sessionID string,
	turnID string,
	clientTurnKey string,
	executionProfile string,
	runID string,
	runSchemaVersion string,
	contentType string,
	content string,
	createdAt time.Time,
) ApplicationResultArtifact {
	return ApplicationResultArtifact{
		SchemaVersion: applicationResultArtifactSchemaVersion, ArtifactID: artifactID, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, SessionID: sessionID, TurnID: turnID,
		ClientTurnKey: clientTurnKey, ExecutionProfile: executionProfile,
		RunRef:      ApplicationInteractionRunRef{RunID: runID, SchemaVersion: runSchemaVersion},
		ContentType: contentType, Content: content, ContentBytes: len([]byte(content)),
		ContentDigest: applicationResultArtifactContentDigest(contentType, content),
		CreatedAt:     createdAt.Format(time.RFC3339Nano), CreatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func seedApplicationResultArtifactLibraryDevFixture(server *Server) (applicationResultArtifactLibraryFixture, error) {
	fixture := applicationResultArtifactLibraryFixtureDefinition()
	if server == nil || server.applicationCatalogRepository == nil || server.applicationResultArtifactRepository == nil {
		return fixture, errors.New("result artifact library fixture requires catalog and artifact repositories")
	}
	if err := seedApplicationResultArtifactLibraryCatalog(server.applicationCatalogRepository, fixture); err != nil {
		return fixture, err
	}
	for _, artifact := range fixture.Artifacts {
		created, _, err := server.applicationResultArtifactRepository.Create(fixture.ArtifactContext, artifact)
		if err != nil {
			return fixture, fmt.Errorf("create result artifact library fixture %s: %w", artifact.ArtifactID, err)
		}
		if !applicationResultArtifactsEquivalent(created, artifact) {
			return fixture, fmt.Errorf("result artifact library fixture %s conflicts with existing data", artifact.ArtifactID)
		}
		lifecycle, err := server.applicationResultArtifactRepository.ReadLifecycle(fixture.ArtifactContext, artifact.ArtifactID)
		if err != nil {
			return fixture, fmt.Errorf("read result artifact library fixture lifecycle %s: %w", artifact.ArtifactID, err)
		}
		if _, shouldArchive := fixture.InitiallyArchivedArtifactIDs[artifact.ArtifactID]; shouldArchive &&
			lifecycle.LifecycleVersion == 1 && lifecycle.LifecycleState == ApplicationResultArtifactLifecycleActive {
			if _, _, err = server.applicationResultArtifactRepository.TransitionLifecycle(
				fixture.ArtifactContext, artifact.ArtifactID, ApplicationResultArtifactLifecycleArchived,
				lifecycle.LifecycleVersion, time.Date(2026, 8, 17, 10, 5, 0, 123456000, time.UTC),
			); err != nil {
				return fixture, fmt.Errorf("archive result artifact library fixture %s: %w", artifact.ArtifactID, err)
			}
		}
	}
	return fixture, nil
}

func seedApplicationResultArtifactLibraryCatalog(
	repository applicationCatalogRepository,
	fixture applicationResultArtifactLibraryFixture,
) error {
	existing, err := repository.Read(fixture.CatalogContext, fixture.Application.ApplicationID)
	if errors.Is(err, errApplicationCatalogNotFound) {
		if _, createErr := repository.Create(fixture.CatalogContext, fixture.Application); createErr != nil {
			return fmt.Errorf("create result artifact library fixture application: %w", createErr)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read result artifact library fixture application: %w", err)
	}
	if existing.SchemaVersion != fixture.Application.SchemaVersion ||
		existing.ApplicationID != fixture.Application.ApplicationID ||
		existing.TenantRef != fixture.Application.TenantRef ||
		existing.WorkspaceID != fixture.Application.WorkspaceID ||
		existing.OwnerSubjectRef != fixture.Application.OwnerSubjectRef ||
		existing.DisplayName != fixture.Application.DisplayName ||
		existing.Description != fixture.Application.Description ||
		existing.ApplicationKind != fixture.Application.ApplicationKind ||
		existing.LifecycleState != applicationCatalogLifecycleActive {
		return errors.New("result artifact library fixture application conflicts with existing data")
	}
	return nil
}
