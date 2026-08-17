package httpapi

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestApplicationResultArtifactLifecycleDefaultsFiltersAndTransitions(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, replay, err := repository.Create(ctx, artifact); err != nil || replay {
		t.Fatalf("create lifecycle fixture: replay=%v err=%v", replay, err)
	}
	service := newApplicationResultArtifactService(repository)
	service.now = func() time.Time { return time.Date(2026, 8, 17, 8, 30, 0, 123456789, time.UTC) }

	active := service.List(ctx, ApplicationResultArtifactListInput{SessionID: artifact.SessionID})
	if active.FailureCode != "" || len(active.Items) != 1 ||
		active.Items[0].SchemaVersion != applicationResultArtifactSummarySchemaVersion ||
		active.Items[0].LifecycleState != ApplicationResultArtifactLifecycleActive ||
		active.Items[0].LifecycleVersion != 1 || active.Items[0].ArchivedAt != nil {
		t.Fatalf("initial active lifecycle list drifted: %#v", active)
	}
	archived := service.Archive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: artifact.SessionID, ArtifactID: artifact.ArtifactID, ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" || archived.Lifecycle == nil || archived.Event == nil ||
		archived.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived ||
		archived.Lifecycle.LifecycleVersion != 2 || archived.Lifecycle.ArchivedAt == nil ||
		archived.Event.TransitionKind != ApplicationResultArtifactLifecycleTransitionArchived {
		t.Fatalf("archive lifecycle drifted: %#v", archived)
	}
	if page := service.List(ctx, ApplicationResultArtifactListInput{SessionID: artifact.SessionID}); page.FailureCode != "" || len(page.Items) != 0 {
		t.Fatalf("archived artifact remained in default active list: %#v", page)
	}
	archivedPage := service.List(ctx, ApplicationResultArtifactListInput{
		SessionID: artifact.SessionID, LifecycleState: ApplicationResultArtifactLifecycleArchived,
	})
	if archivedPage.FailureCode != "" || len(archivedPage.Items) != 1 ||
		archivedPage.Items[0].LifecycleVersion != 2 || archivedPage.Items[0].ArchivedAt == nil {
		t.Fatalf("archived artifact list drifted: %#v", archivedPage)
	}
	read := service.Read(ctx, artifact.ArtifactID)
	if read.FailureCode != "" || read.Artifact == nil || read.Lifecycle == nil ||
		read.Artifact.Content != artifact.Content || read.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleArchived {
		t.Fatalf("exact archived artifact read drifted: %#v", read)
	}

	repeated := service.Archive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: artifact.SessionID, ArtifactID: artifact.ArtifactID, ExpectedLifecycleVersion: 2,
	})
	if repeated.FailureCode != ApplicationResultArtifactFailureLifecycleState ||
		repeated.CurrentLifecycleVersion != 2 || repeated.CurrentLifecycleState != ApplicationResultArtifactLifecycleArchived {
		t.Fatalf("repeated archive did not fail closed: %#v", repeated)
	}
	stale := service.Unarchive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: artifact.SessionID, ArtifactID: artifact.ArtifactID, ExpectedLifecycleVersion: 1,
	})
	if stale.FailureCode != ApplicationResultArtifactFailureLifecycleVersion || stale.CurrentLifecycleVersion != 2 {
		t.Fatalf("stale unarchive did not expose sanitized current version: %#v", stale)
	}
	unarchived := service.Unarchive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
		SessionID: artifact.SessionID, ArtifactID: artifact.ArtifactID, ExpectedLifecycleVersion: 2,
	})
	if unarchived.FailureCode != "" || unarchived.Lifecycle == nil || unarchived.Event == nil ||
		unarchived.Lifecycle.LifecycleState != ApplicationResultArtifactLifecycleActive ||
		unarchived.Lifecycle.LifecycleVersion != 3 || unarchived.Lifecycle.ArchivedAt != nil ||
		unarchived.Event.TransitionKind != ApplicationResultArtifactLifecycleTransitionUnarchived {
		t.Fatalf("unarchive lifecycle drifted: %#v", unarchived)
	}
}

func TestApplicationResultArtifactLifecycleCursorBindsStateAndConcurrentCAS(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, artifacts := applicationResultArtifactPersistenceProfileFixtures()
	artifacts[1].SessionID = artifacts[0].SessionID
	for _, artifact := range artifacts[:2] {
		if _, _, err := repository.Create(ctx, artifact); err != nil {
			t.Fatalf("create lifecycle cursor fixture: %v", err)
		}
	}
	service := newApplicationResultArtifactService(repository)
	page := service.List(ctx, ApplicationResultArtifactListInput{SessionID: artifacts[0].SessionID, Limit: 1})
	if page.FailureCode != "" || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatalf("lifecycle cursor page drifted: %#v", page)
	}
	drifted := service.List(ctx, ApplicationResultArtifactListInput{
		SessionID: artifacts[0].SessionID, LifecycleState: ApplicationResultArtifactLifecycleArchived,
		Limit: 1, Cursor: *page.NextCursor,
	})
	if drifted.FailureCode != ApplicationResultArtifactFailurePayloadInvalid {
		t.Fatalf("lifecycle state cursor drift did not fail closed: %#v", drifted)
	}

	outcomes := make(chan ApplicationResultArtifactLifecycleTransitionResult, 8)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			outcomes <- service.Archive(ctx, ApplicationResultArtifactLifecycleTransitionInput{
				SessionID: artifacts[0].SessionID, ArtifactID: artifacts[0].ArtifactID, ExpectedLifecycleVersion: 1,
			})
		}()
	}
	wait.Wait()
	close(outcomes)
	successes, conflicts := 0, 0
	for outcome := range outcomes {
		switch outcome.FailureCode {
		case "":
			successes++
		case ApplicationResultArtifactFailureLifecycleVersion:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent lifecycle result: %#v", outcome)
		}
	}
	if successes != 1 || conflicts != 7 {
		t.Fatalf("concurrent lifecycle CAS did not converge: successes=%d conflicts=%d", successes, conflicts)
	}

	otherScope := ctx
	otherScope.OwnerSubjectRef = "subject_other"
	if _, err := repository.ReadLifecycle(otherScope, artifacts[0].ArtifactID); !errors.Is(err, errApplicationResultArtifactNotFound) {
		t.Fatalf("cross-owner lifecycle read did not fail closed: %v", err)
	}
}

func TestApplicationResultArtifactLifecycleCorruptionFailsClosed(t *testing.T) {
	repository := newMemoryApplicationResultArtifactRepository()
	ctx, artifact := applicationResultArtifactPersistenceFixture()
	if _, _, err := repository.Create(ctx, artifact); err != nil {
		t.Fatalf("create lifecycle corruption fixture: %v", err)
	}
	repository.lifecycleByID[artifact.ArtifactID] = ApplicationResultArtifactLifecycle{
		SchemaVersion: applicationResultArtifactLifecycleSchemaVersion, ArtifactID: artifact.ArtifactID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, LifecycleState: ApplicationResultArtifactLifecycleActive,
		LifecycleVersion: 0,
	}
	if _, err := repository.ReadLifecycle(ctx, artifact.ArtifactID); !errors.Is(err, errApplicationResultArtifactContract) {
		t.Fatalf("corrupt lifecycle did not fail closed: %v", err)
	}
}
