package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestSQLiteApplicationEvaluationScheduleRestartSingleWinnerAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "application-evaluation-schedules.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newSQLiteApplicationEvaluationScheduleRepository(runtime.DB())
	if err = repository.CreateSchedule(ctx, schedule, version); err != nil {
		_ = runtime.Close()
		t.Fatalf("create SQLite schedule: %v", err)
	}
	active := applicationEvaluationActivateScheduleRepository(t, repository, ctx, schedule, "2026-08-30T09:30:00Z")
	if active.RecordVersion != 2 || active.LifecycleState != applicationEvaluationScheduleStateActive {
		_ = runtime.Close()
		t.Fatalf("activate SQLite schedule: %+v", active)
	}
	duePage, err := repository.ListDueSchedules(context.Background(), "2026-08-30T09:30:00Z", 1)
	if err != nil || len(duePage.Schedules) != 1 || duePage.Schedules[0].ScheduleID != schedule.ScheduleID || duePage.HasMore {
		_ = runtime.Close()
		t.Fatalf("list SQLite runner due schedules: page=%+v err=%v", duePage, err)
	}

	systemCtx, due, claimed := applicationEvaluationScheduleOccurrenceTestRecords(ctx, version, "2026-08-30T09:30:00Z")
	if !validApplicationEvaluationScheduleClaim(systemCtx, due, claimed) {
		_ = runtime.Close()
		t.Fatal("test occurrence claim records do not satisfy the canonical contract")
	}
	var winners atomic.Int32
	var unexpected atomic.Int32
	var wait sync.WaitGroup
	var firstUnexpected error
	var firstUnexpectedOnce sync.Once
	for range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, won, claimErr := repository.ClaimOccurrence(systemCtx, due, claimed)
			if claimErr == nil && won {
				winners.Add(1)
				return
			}
			if !errors.Is(claimErr, errApplicationEvaluationScheduleClaimConflict) || won {
				unexpected.Add(1)
				firstUnexpectedOnce.Do(func() { firstUnexpected = claimErr })
			}
		}()
	}
	wait.Wait()
	if winners.Load() != 1 || unexpected.Load() != 0 {
		var occurrenceCount int
		var rawOccurrence string
		_ = runtime.DB().QueryRowContext(context.Background(), `SELECT count(*) FROM application_evaluation_schedule_occurrences`).Scan(&occurrenceCount)
		if occurrenceCount > 0 {
			_ = runtime.DB().QueryRowContext(context.Background(), `SELECT sanitized_occurrence_record FROM application_evaluation_schedule_occurrences LIMIT 1`).Scan(&rawOccurrence)
		}
		_ = runtime.Close()
		t.Fatalf("SQLite occurrence claim must have one winner: winners=%d unexpected=%d first_error=%v rows=%d payload=%s", winners.Load(), unexpected.Load(), firstUnexpected, occurrenceCount, rawOccurrence)
	}
	openPage, err := repository.ListOpenOccurrences(context.Background(), 1)
	if err != nil || len(openPage.Occurrences) != 1 || openPage.Occurrences[0].ClientCampaignKey != due.ClientCampaignKey || openPage.HasMore {
		_ = runtime.Close()
		t.Fatalf("list SQLite runner open occurrences: page=%+v err=%v", openPage, err)
	}
	if err = runtime.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatal(err)
	}
	restartedRepository := newSQLiteApplicationEvaluationScheduleRepository(restarted.DB())
	storedSchedule, found, err := restartedRepository.ReadSchedule(ctx, schedule.ScheduleID)
	if err != nil || !found || storedSchedule.RecordVersion != 2 || storedSchedule.NextDueAt == nil ||
		*storedSchedule.NextDueAt != due.ScheduledForUTC {
		_ = restarted.Close()
		t.Fatalf("read restarted SQLite schedule: found=%v value=%+v err=%v", found, storedSchedule, err)
	}
	storedVersion, found, err := restartedRepository.ReadScheduleVersion(ctx, schedule.ScheduleID, 1)
	if err != nil || !found || storedVersion.ScheduleDigest != version.ScheduleDigest {
		_ = restarted.Close()
		t.Fatalf("read restarted immutable schedule version: found=%v value=%+v err=%v", found, storedVersion, err)
	}
	storedOccurrence, found, err := restartedRepository.ReadOccurrence(systemCtx, schedule.ScheduleID, 1, due.ScheduledForUTC)
	if err != nil || !found || storedOccurrence.State != applicationEvaluationScheduleOccurrenceStateClaimed || storedOccurrence.RecordVersion != 2 {
		_ = restarted.Close()
		t.Fatalf("read restarted SQLite occurrence: found=%v value=%+v err=%v", found, storedOccurrence, err)
	}
	restartedOpenPage, err := restartedRepository.ListOpenOccurrences(context.Background(), 1)
	if err != nil || len(restartedOpenPage.Occurrences) != 1 || restartedOpenPage.Occurrences[0].ClientCampaignKey != due.ClientCampaignKey || restartedOpenPage.HasMore {
		_ = restarted.Close()
		t.Fatalf("list restarted SQLite runner open occurrences: page=%+v err=%v", restartedOpenPage, err)
	}

	if _, err = restarted.DB().ExecContext(context.Background(), `UPDATE application_evaluation_schedule_versions
SET sanitized_schedule_version_record=sanitized_schedule_version_record WHERE schedule_id=?`, schedule.ScheduleID); err == nil {
		_ = restarted.Close()
		t.Fatal("SQLite mutated an immutable schedule version")
	}
	if _, err = restarted.DB().ExecContext(context.Background(), `DELETE FROM application_evaluation_schedule_occurrences
WHERE schedule_id=?`, schedule.ScheduleID); err == nil {
		_ = restarted.Close()
		t.Fatal("SQLite deleted an immutable schedule occurrence")
	}
	if err = restarted.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = restartedRepository.ReadSchedule(ctx, schedule.ScheduleID); !errors.Is(err, errApplicationEvaluationScheduleStoreUnavailable) {
		t.Fatalf("closed SQLite schedule store fell back instead of failing closed: %v", err)
	}
}

func TestSQLiteApplicationEvaluationScheduleCorruptionCloses(t *testing.T) {
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "application-evaluation-schedule-corruption.db"),
		Migrations:   sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer runtime.Close()
	ctx, schedule, version := applicationEvaluationScheduleTestRecords(t)
	repository := newSQLiteApplicationEvaluationScheduleRepository(runtime.DB())
	if err = repository.CreateSchedule(ctx, schedule, version); err != nil {
		t.Fatalf("create SQLite schedule: %v", err)
	}
	applicationEvaluationActivateScheduleRepository(t, repository, ctx, schedule, "2026-08-30T09:30:00Z")
	if _, err = runtime.DB().ExecContext(context.Background(), `PRAGMA ignore_check_constraints=ON`); err != nil {
		t.Fatal(err)
	}
	if _, err = runtime.DB().ExecContext(context.Background(), `UPDATE application_evaluation_schedules
SET record_version=record_version+1,
    sanitized_schedule_record=json_set(sanitized_schedule_record,'$.record_version',record_version+1,'$.unexpected',1)
WHERE schedule_id=?`, schedule.ScheduleID); err != nil {
		t.Fatalf("inject corrupted SQLite schedule projection: %v", err)
	}
	if _, found, readErr := repository.ReadSchedule(ctx, schedule.ScheduleID); found || !errors.Is(readErr, errApplicationEvaluationScheduleStoreContract) {
		t.Fatalf("corrupted SQLite schedule did not fail closed: found=%v err=%v", found, readErr)
	}
}

func applicationEvaluationActivateScheduleRepository(
	t *testing.T,
	repository applicationEvaluationScheduleRepository,
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	nextDueAt string,
) ApplicationEvaluationSchedule {
	t.Helper()
	active := schedule
	active.RecordVersion++
	active.LifecycleState = applicationEvaluationScheduleStateActive
	active.NextDueAt = &nextDueAt
	active.UpdatedAt = "2026-08-30T08:01:00Z"
	active.RequestID = "request-schedule-activate"
	active.AuditRef = "audit:schedule-activate"
	activationContext := ctx
	activationContext.RequestID = active.RequestID
	activationContext.AuditRef = active.AuditRef
	stored, updated, err := repository.UpdateSchedule(activationContext, schedule.RecordVersion, active)
	if err != nil || !updated {
		t.Fatalf("activate schedule repository: updated=%v value=%#v err=%v", updated, stored, err)
	}
	return stored
}
