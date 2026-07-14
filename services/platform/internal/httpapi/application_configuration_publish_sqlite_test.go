package httpapi

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	applicationdraftmigrations "radishmind.local/services/platform/migrations/sqlite/application_configuration_drafts"
	applicationpublishmigrations "radishmind.local/services/platform/migrations/sqlite/application_publish_candidates"
)

func TestApplicationConfigurationDraftSQLiteRepositoryContract(t *testing.T) {
	runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	runApplicationConfigurationDraftLifecycleAndIsolation(t, newSQLiteApplicationConfigurationDraftRepository(runtime.DB()))
}

func TestApplicationPublishCandidateSQLiteRepositoryContract(t *testing.T) {
	t.Run("lifecycle and eligibility", func(t *testing.T) {
		runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runApplicationPublishCandidateLifecycleAndEligibility(t,
			newSQLiteApplicationConfigurationDraftRepository(runtime.DB()),
			newSQLiteApplicationPublishCandidateRepository(runtime.DB()),
		)
	})
	t.Run("draft binding and scope", func(t *testing.T) {
		runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runApplicationPublishCandidateDraftBindingAndScope(t,
			newSQLiteApplicationConfigurationDraftRepository(runtime.DB()),
			newSQLiteApplicationPublishCandidateRepository(runtime.DB()),
		)
	})
	t.Run("draft drift supersede and list order", func(t *testing.T) {
		runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runApplicationPublishCandidateDraftDriftAndSupersede(t,
			newSQLiteApplicationConfigurationDraftRepository(runtime.DB()),
			newSQLiteApplicationPublishCandidateRepository(runtime.DB()),
		)
	})
}

func TestApplicationConfigurationDraftSQLiteListOrder(t *testing.T) {
	runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	service := newApplicationConfigurationDraftService(newSQLiteApplicationConfigurationDraftRepository(runtime.DB()))
	service.now = func() time.Time { return time.Date(2026, 7, 14, 10, 0, 0, 0, time.UTC) }
	requestContext := validApplicationDraftContext()
	for _, draftID := range []string{"draft-b", "draft-a"} {
		payload := validApplicationDraftPayload()
		payload.DraftID = draftID
		if result := service.Save(requestContext, payload, 0); result.Draft == nil {
			t.Fatalf("save %s for list order: %#v", draftID, result)
		}
	}
	summaries, failure := service.List(requestContext)
	if failure != "" || len(summaries) != 2 || summaries[0].DraftID != "draft-a" || summaries[1].DraftID != "draft-b" {
		t.Fatalf("application draft list order is unstable: %#v failure=%s", summaries, failure)
	}
}

func TestApplicationConfigurationAndPublishSQLiteConcurrentCAS(t *testing.T) {
	runtime := openApplicationConfigurationPublishSQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
	draftRepository := newSQLiteApplicationConfigurationDraftRepository(runtime.DB())
	draftService := newApplicationConfigurationDraftService(draftRepository)
	requestContext := validApplicationDraftContext()
	payload := validApplicationDraftPayload()
	if result := draftService.Save(requestContext, payload, 0); result.Draft == nil {
		t.Fatalf("seed application draft: %#v", result)
	}
	payload.Description = "Concurrent SQLite configuration update"
	var draftSuccesses atomic.Int32
	var draftConflicts atomic.Int32
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := draftService.Save(requestContext, payload, 1)
			switch result.FailureCode {
			case "":
				draftSuccesses.Add(1)
			case ApplicationDraftFailureVersionConflict:
				draftConflicts.Add(1)
			default:
				t.Errorf("unexpected application draft CAS result: %#v", result)
			}
		}()
	}
	wait.Wait()
	if draftSuccesses.Load() != 1 || draftConflicts.Load() != 7 {
		t.Fatalf("application draft CAS must select one writer: successes=%d conflicts=%d", draftSuccesses.Load(), draftConflicts.Load())
	}

	publishRepository := newSQLiteApplicationPublishCandidateRepository(runtime.DB())
	publishService := newApplicationPublishCandidateService(draftRepository, publishRepository, validApplicationPublishBaseline)
	created := publishService.Create(validApplicationPublishContext(), ApplicationPublishCreateInput{
		CandidateID: "candidate-concurrent", DraftID: payload.DraftID, ExpectedDraftVersion: 2,
	})
	if created.Candidate == nil {
		t.Fatalf("create candidate for concurrent review: %#v", created)
	}
	var reviewSuccesses atomic.Int32
	var reviewConflicts atomic.Int32
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := publishService.Review(validApplicationPublishContext(), "candidate-concurrent", ApplicationPublishReviewInput{
				ExpectedReviewVersion: 0, Decision: "approve", Reason: "Concurrent SQLite review accepted for development verification.",
			})
			switch result.FailureCode {
			case "":
				reviewSuccesses.Add(1)
			case ApplicationPublishFailureReviewVersionConflict:
				reviewConflicts.Add(1)
			default:
				t.Errorf("unexpected publish review CAS result: %#v", result)
			}
		}()
	}
	wait.Wait()
	if reviewSuccesses.Load() != 1 || reviewConflicts.Load() != 7 {
		t.Fatalf("publish review CAS must select one writer: successes=%d conflicts=%d", reviewSuccesses.Load(), reviewConflicts.Load())
	}
}

func TestApplicationConfigurationAndPublishSQLiteRestartRecoveryAndSensitiveMaterialBoundary(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationConfigurationPublishSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("open first configuration and publish SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = firstRuntime.Close() })
	draftRepository := newSQLiteApplicationConfigurationDraftRepository(firstRuntime.DB())
	draftService := newApplicationConfigurationDraftService(draftRepository)
	draftContext := validApplicationDraftContext()
	secretMarker := "Authorization: Bearer forbidden-sqlite-marker"
	secretPayload := validApplicationDraftPayload()
	secretPayload.Description = secretMarker
	if result := draftService.Save(draftContext, secretPayload, 0); result.FailureCode != ApplicationDraftFailureSecretForbidden {
		t.Fatalf("secret-bearing draft must be rejected before persistence: %#v", result)
	}
	createdDraft := draftService.Save(draftContext, validApplicationDraftPayload(), 0)
	if createdDraft.Draft == nil {
		t.Fatalf("create restart recovery draft: %#v", createdDraft)
	}
	publishRepository := newSQLiteApplicationPublishCandidateRepository(firstRuntime.DB())
	publishService := newApplicationPublishCandidateService(draftRepository, publishRepository, validApplicationPublishBaseline)
	publishContext := validApplicationPublishContext()
	createdCandidate := publishService.Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-restart", DraftID: createdDraft.Draft.DraftID, ExpectedDraftVersion: 1,
		EvidenceRequestIDs: []string{"request-safe-evidence"},
	})
	if createdCandidate.Candidate == nil {
		t.Fatalf("create restart recovery candidate: %#v", createdCandidate)
	}
	if result := publishService.Review(publishContext, "candidate-restart", ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: secretMarker,
	}); result.FailureCode != ApplicationPublishFailureSecretForbidden {
		t.Fatalf("secret-bearing review must be rejected before persistence: %#v", result)
	}
	approved := publishService.Review(publishContext, "candidate-restart", ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "SQLite restart recovery review accepted.",
	})
	if approved.Candidate == nil || approved.Candidate.ReviewVersion != 1 {
		t.Fatalf("approve restart recovery candidate: %#v", approved)
	}
	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first configuration and publish SQLite runtime: %v", err)
	}
	if result := draftService.Read(draftContext, createdDraft.Draft.DraftID); result.FailureCode != ApplicationDraftFailureStoreUnavailable {
		t.Fatalf("closed draft store must not fall back to memory: %#v", result)
	}
	if result := publishService.Read(publishContext, createdCandidate.Candidate.CandidateID); result.FailureCode != ApplicationPublishFailureStoreUnavailable {
		t.Fatalf("closed publish store must not fall back to memory: %#v", result)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, secretMarker)

	secondRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationConfigurationPublishSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("reopen configuration and publish SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = secondRuntime.Close() })
	restartedDraftRepository := newSQLiteApplicationConfigurationDraftRepository(secondRuntime.DB())
	restoredDraft := newApplicationConfigurationDraftService(restartedDraftRepository).Read(draftContext, createdDraft.Draft.DraftID)
	if restoredDraft.Draft == nil || restoredDraft.Draft.DraftVersion != 1 {
		t.Fatalf("restore application draft after SQLite restart: %#v", restoredDraft)
	}
	restartedPublishService := newApplicationPublishCandidateService(
		restartedDraftRepository,
		newSQLiteApplicationPublishCandidateRepository(secondRuntime.DB()),
		validApplicationPublishBaseline,
	)
	restoredCandidate := restartedPublishService.Read(publishContext, createdCandidate.Candidate.CandidateID)
	if restoredCandidate.Candidate == nil || restoredCandidate.Candidate.ReviewVersion != 1 ||
		restoredCandidate.Candidate.CandidateState != applicationPublishStateApproved ||
		len(restoredCandidate.Candidate.Reviews) != 1 || restoredCandidate.Candidate.PromotionEligibility.Eligible ||
		!hasPromotionBlocker(restoredCandidate.Candidate.PromotionEligibility, "promotion_disabled") {
		t.Fatalf("restore publish candidate after SQLite restart: %#v", restoredCandidate)
	}
	otherOwnerDraftContext := draftContext
	otherOwnerDraftContext.OwnerSubjectRef = "subject_other"
	if result := newApplicationConfigurationDraftService(restartedDraftRepository).Read(otherOwnerDraftContext, createdDraft.Draft.DraftID); result.FailureCode != ApplicationDraftFailureNotFound {
		t.Fatalf("restart must preserve application draft owner isolation: %#v", result)
	}
	otherOwnerPublishContext := publishContext
	otherOwnerPublishContext.OwnerSubjectRef = "subject_other"
	if result := restartedPublishService.Read(otherOwnerPublishContext, createdCandidate.Candidate.CandidateID); result.FailureCode != ApplicationPublishFailureNotFound {
		t.Fatalf("restart must preserve publish candidate owner isolation: %#v", result)
	}
}

func openApplicationConfigurationPublishSQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   applicationConfigurationPublishSQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("open configuration and publish SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}

func applicationConfigurationPublishSQLiteMigrations() []sqlitedev.Migration {
	migrations := applicationdraftmigrations.Migrations()
	return append(migrations, applicationpublishmigrations.Migrations()...)
}

func assertSQLiteFilesExcludeMarker(t *testing.T, databasePath string, marker string) {
	t.Helper()
	for _, path := range []string{databasePath, databasePath + "-wal", databasePath + "-shm"} {
		payload, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("read SQLite persistence file %s: %v", filepath.Base(path), err)
		}
		if bytes.Contains(payload, []byte(marker)) {
			t.Fatalf("SQLite persistence file %s contains rejected sensitive material", filepath.Base(path))
		}
	}
}
