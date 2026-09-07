package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteApplicationEvaluationScheduleRepository struct{ database *sql.DB }

func newSQLiteApplicationEvaluationScheduleRepository(database *sql.DB) *sqliteApplicationEvaluationScheduleRepository {
	return &sqliteApplicationEvaluationScheduleRepository{database: database}
}

type applicationEvaluationScheduleSQLScanner interface{ Scan(...any) error }

func (repository *sqliteApplicationEvaluationScheduleRepository) CreateSchedule(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) error {
	if repository == nil || repository.database == nil || !validApplicationEvaluationScheduleCreate(ctx, schedule, version) {
		return errApplicationEvaluationScheduleStoreContract
	}
	schedulePayload, updatedAt, nextDueAt, err := encodeSQLiteApplicationEvaluationSchedule(schedule)
	if err != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	versionPayload, createdAt, err := encodeSQLiteApplicationEvaluationScheduleVersion(version)
	if err != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	if _, found, readErr := readSQLiteApplicationEvaluationSchedule(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID)); readErr != nil {
		return readErr
	} else if found {
		return errApplicationEvaluationScheduleVersionConflict
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_schedules
(tenant_ref,workspace_id,environment,application_id,schedule_id,record_version,latest_schedule_version,latest_schedule_digest,lifecycle_state,updated_at_unix_nano,next_due_at_unix_nano,sanitized_schedule_record)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID,
		schedule.RecordVersion, schedule.LatestScheduleVersion, schedule.LatestScheduleDigest, schedule.LifecycleState, updatedAt, nextDueAt, schedulePayload); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_versions
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest,created_at_unix_nano,sanitized_schedule_version_record)
VALUES(?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.ScheduleID,
		version.ScheduleVersion, version.ScheduleDigest, createdAt, versionPayload); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	if err = tx.Commit(); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	return nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ReviseSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationScheduleRevision(ctx, expected, schedule, version) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationSchedule(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, false, err
	}
	currentVersion, versionFound, err := readSQLiteApplicationEvaluationScheduleVersion(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, current.ScheduleID, current.LatestScheduleVersion))
	if err != nil || !versionFound {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	if !validApplicationEvaluationScheduleRevisionAgainstCurrent(current, currentVersion, schedule, version) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	schedulePayload, updatedAt, nextDueAt, err := encodeSQLiteApplicationEvaluationSchedule(schedule)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	versionPayload, createdAt, err := encodeSQLiteApplicationEvaluationScheduleVersion(version)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_versions
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest,created_at_unix_nano,sanitized_schedule_version_record)
VALUES(?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, version.ScheduleID,
		version.ScheduleVersion, version.ScheduleDigest, createdAt, versionPayload); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_schedules SET
record_version=?,latest_schedule_version=?,latest_schedule_digest=?,lifecycle_state=?,updated_at_unix_nano=?,next_due_at_unix_nano=?,sanitized_schedule_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND record_version=?`,
		schedule.RecordVersion, schedule.LatestScheduleVersion, schedule.LatestScheduleDigest, schedule.LifecycleState, updatedAt, nextDueAt, schedulePayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID, expected)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) UpdateSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationScheduleUpdateInput(ctx, expected, schedule) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationSchedule(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, false, err
	}
	version, versionFound, err := readSQLiteApplicationEvaluationScheduleVersion(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, current.ScheduleID, current.LatestScheduleVersion))
	if err != nil || !versionFound {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	if !validApplicationEvaluationScheduleUpdateAgainstCurrent(ctx, current, version, schedule) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	payload, updatedAt, nextDueAt, err := encodeSQLiteApplicationEvaluationSchedule(schedule)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_schedules SET
record_version=?,lifecycle_state=?,updated_at_unix_nano=?,next_due_at_unix_nano=?,sanitized_schedule_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND record_version=?`,
		schedule.RecordVersion, schedule.LifecycleState, updatedAt, nextDueAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID, expected)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ReadSchedule(
	ctx ApplicationEvaluationContext,
	scheduleID string,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	schedule, found, err := readSQLiteApplicationEvaluationSchedule(ctx, repository.database.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, scheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, found, err
	}
	version, versionFound, err := repository.ReadScheduleVersion(ctx, scheduleID, schedule.LatestScheduleVersion)
	if err != nil || !versionFound || !applicationEvaluationScheduleMatchesVersion(schedule, version) {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	return schedule, true, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ListSchedules(
	ctx ApplicationEvaluationContext,
	filter ApplicationEvaluationScheduleListFilter,
) (ApplicationEvaluationScheduleListPage, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) ||
		!validApplicationEvaluationScheduleState(filter.LifecycleState) || filter.Limit < 1 {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreContract
	}
	before, err := optionalApplicationEvaluationScheduleUnixNano(filter.BeforeUpdatedAt)
	if err != nil {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreContract
	}
	rows, err := repository.database.QueryContext(ctx.RequestContext, `SELECT schedules.sanitized_schedule_record,versions.sanitized_schedule_version_record
FROM application_evaluation_schedules schedules JOIN application_evaluation_schedule_versions versions
ON versions.tenant_ref=schedules.tenant_ref AND versions.workspace_id=schedules.workspace_id AND versions.environment=schedules.environment
AND versions.application_id=schedules.application_id AND versions.schedule_id=schedules.schedule_id AND versions.schedule_version=schedules.latest_schedule_version
WHERE schedules.tenant_ref=? AND schedules.workspace_id=? AND schedules.environment=? AND schedules.application_id=? AND schedules.lifecycle_state=?
AND (? IS NULL OR schedules.updated_at_unix_nano<? OR (schedules.updated_at_unix_nano=? AND schedules.schedule_id<?))
ORDER BY schedules.updated_at_unix_nano DESC,schedules.schedule_id DESC LIMIT ?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.LifecycleState,
		before, before, before, filter.BeforeScheduleID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationSchedule, 0, filter.Limit+1)
	for rows.Next() {
		schedule, version, scanErr := scanSQLiteApplicationEvaluationSchedulePair(ctx, rows)
		if scanErr != nil {
			return ApplicationEvaluationScheduleListPage{}, scanErr
		}
		if !applicationEvaluationScheduleMatchesVersion(schedule, version) {
			return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreContract
		}
		values = append(values, schedule)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationScheduleListPage{Schedules: values, HasMore: hasMore}, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ListDueSchedules(
	requestContext context.Context,
	dueThrough string,
	limit int,
) (ApplicationEvaluationScheduleRunnerPage, error) {
	if repository == nil || repository.database == nil || requestContext == nil || limit < 1 {
		return ApplicationEvaluationScheduleRunnerPage{}, errApplicationEvaluationScheduleStoreContract
	}
	dueTime, ok := parseApplicationEvaluationScheduleUTCTimestamp(dueThrough)
	if !ok {
		return ApplicationEvaluationScheduleRunnerPage{}, errApplicationEvaluationScheduleStoreContract
	}
	rows, err := repository.database.QueryContext(requestContext, `SELECT schedules.sanitized_schedule_record,versions.sanitized_schedule_version_record
FROM application_evaluation_schedules schedules JOIN application_evaluation_schedule_versions versions
ON versions.tenant_ref=schedules.tenant_ref AND versions.workspace_id=schedules.workspace_id AND versions.environment=schedules.environment
AND versions.application_id=schedules.application_id AND versions.schedule_id=schedules.schedule_id AND versions.schedule_version=schedules.latest_schedule_version
WHERE schedules.lifecycle_state=? AND schedules.next_due_at_unix_nano<=?
ORDER BY schedules.next_due_at_unix_nano ASC,schedules.schedule_id ASC LIMIT ?`,
		applicationEvaluationScheduleStateActive, dueTime.UnixNano(), limit+1)
	if err != nil {
		return ApplicationEvaluationScheduleRunnerPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationSchedule, 0, limit+1)
	for rows.Next() {
		schedule, version, scanErr := scanApplicationEvaluationScheduleRunnerPair(rows)
		if scanErr != nil || !applicationEvaluationScheduleMatchesVersion(schedule, version) {
			return ApplicationEvaluationScheduleRunnerPage{}, firstApplicationEvaluationScheduleStoreError(scanErr)
		}
		values = append(values, schedule)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationScheduleRunnerPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	hasMore := len(values) > limit
	if hasMore {
		values = values[:limit]
	}
	return ApplicationEvaluationScheduleRunnerPage{Schedules: values, HasMore: hasMore}, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ReadScheduleVersion(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	versionNumber int,
) (ApplicationEvaluationScheduleVersion, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) || versionNumber < 1 {
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return readSQLiteApplicationEvaluationScheduleVersion(ctx, repository.database.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, scheduleID, versionNumber))
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ClaimOccurrence(
	ctx ApplicationEvaluationContext,
	due ApplicationEvaluationScheduleOccurrence,
	claimed ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationScheduleClaim(ctx, due, claimed) {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	schedule, found, err := repository.ReadSchedule(ctx, due.ScheduleID)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, err
	}
	if !found {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleNotFound
	}
	version, found, err := repository.ReadScheduleVersion(ctx, due.ScheduleID, due.ScheduleVersion)
	if err != nil || !found || !validApplicationEvaluationScheduleClaimBinding(schedule, version, due) {
		return ApplicationEvaluationScheduleOccurrence{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	payload, scheduledAt, updatedAt, err := encodeSQLiteApplicationEvaluationScheduleOccurrence(claimed)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	result, err := repository.database.ExecContext(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_occurrences
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,scheduled_for_unix_nano,record_version,schedule_digest,occurrence_state,client_campaign_key,campaign_id,updated_at_unix_nano,sanitized_occurrence_record)
SELECT ?,?,?,?,?,?,?,?,?,?,?,?,?,?
WHERE EXISTS (SELECT 1 FROM application_evaluation_schedules WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=?
AND lifecycle_state='active' AND latest_schedule_version=? AND latest_schedule_digest=? AND next_due_at_unix_nano=?)
AND EXISTS (SELECT 1 FROM application_evaluation_schedule_versions WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND schedule_version=? AND schedule_digest=?)`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, claimed.ScheduleID, claimed.ScheduleVersion, scheduledAt,
		claimed.RecordVersion, claimed.ScheduleDigest, claimed.State, claimed.ClientCampaignKey, claimed.CampaignID, updatedAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, due.ScheduleID, due.ScheduleVersion, due.ScheduleDigest, scheduledAt,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, due.ScheduleID, due.ScheduleVersion, due.ScheduleDigest)
	if err == nil {
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil {
			return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
		}
		if affected == 1 {
			return cloneApplicationEvaluationScheduleOccurrence(claimed), true, nil
		}
		if affected != 0 {
			return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
		}
	}
	current, exists, readErr := repository.ReadOccurrence(ctx, due.ScheduleID, due.ScheduleVersion, due.ScheduledForUTC)
	if readErr != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, readErr
	}
	if exists {
		return current, false, errApplicationEvaluationScheduleClaimConflict
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
}

func (repository *sqliteApplicationEvaluationScheduleRepository) UpdateOccurrence(
	ctx ApplicationEvaluationContext,
	expected int,
	occurrence ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationScheduleOccurrenceUpdateInput(ctx, expected, occurrence) {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback() }()
	current, found, err := readSQLiteApplicationEvaluationScheduleOccurrence(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleOccurrenceReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		occurrence.ScheduleID, occurrence.ScheduleVersion, mustApplicationEvaluationScheduleUnixNano(occurrence.ScheduledForUTC)))
	if err != nil || !found {
		return ApplicationEvaluationScheduleOccurrence{}, false, err
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	version, versionFound, err := readSQLiteApplicationEvaluationScheduleVersion(ctx, tx.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		occurrence.ScheduleID, occurrence.ScheduleVersion))
	if err != nil || !versionFound || !validApplicationEvaluationScheduleOccurrenceAgainstVersion(current, occurrence, version) {
		return ApplicationEvaluationScheduleOccurrence{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	payload, _, updatedAt, err := encodeSQLiteApplicationEvaluationScheduleOccurrence(occurrence)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE application_evaluation_schedule_occurrences SET
record_version=?,occurrence_state=?,campaign_id=?,updated_at_unix_nano=?,sanitized_occurrence_record=?
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND schedule_version=? AND scheduled_for_unix_nano=? AND record_version=?`,
		occurrence.RecordVersion, occurrence.State, occurrence.CampaignID, updatedAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, occurrence.ScheduleID, occurrence.ScheduleVersion,
		mustApplicationEvaluationScheduleUnixNano(occurrence.ScheduledForUTC), expected)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(); err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationScheduleOccurrence(occurrence), true, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ReadOccurrence(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	scheduleVersion int,
	scheduledForUTC string,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.database == nil || !validApplicationEvaluationContext(ctx) || scheduleVersion < 1 {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	scheduledAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(scheduledForUTC)
	if !ok {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	occurrence, found, err := readSQLiteApplicationEvaluationScheduleOccurrence(ctx, repository.database.QueryRowContext(ctx.RequestContext,
		sqliteApplicationEvaluationScheduleOccurrenceReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		scheduleID, scheduleVersion, scheduledAt.UnixNano()))
	if err != nil || !found {
		return ApplicationEvaluationScheduleOccurrence{}, found, err
	}
	version, versionFound, err := repository.ReadScheduleVersion(ctx, scheduleID, scheduleVersion)
	if err != nil || !versionFound || version.ScheduleDigest != occurrence.ScheduleDigest ||
		version.Authorization.SystemActorRef != occurrence.SystemActorRef || version.Authorization.DelegatedByUserRef != occurrence.DelegatedByUserRef {
		return ApplicationEvaluationScheduleOccurrence{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	return occurrence, true, nil
}

func (repository *sqliteApplicationEvaluationScheduleRepository) ListOpenOccurrences(
	requestContext context.Context,
	limit int,
) (ApplicationEvaluationOccurrenceRunnerPage, error) {
	if repository == nil || repository.database == nil || requestContext == nil || limit < 1 {
		return ApplicationEvaluationOccurrenceRunnerPage{}, errApplicationEvaluationScheduleStoreContract
	}
	rows, err := repository.database.QueryContext(requestContext, `SELECT sanitized_occurrence_record
FROM application_evaluation_schedule_occurrences
WHERE occurrence_state IN (?,?,?)
ORDER BY updated_at_unix_nano ASC,client_campaign_key ASC LIMIT ?`,
		applicationEvaluationScheduleOccurrenceStateClaimed, applicationEvaluationScheduleOccurrenceStateCampaignCreated,
		applicationEvaluationScheduleOccurrenceStateObserving, limit+1)
	if err != nil {
		return ApplicationEvaluationOccurrenceRunnerPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationScheduleOccurrence, 0, limit+1)
	for rows.Next() {
		occurrence, scanErr := scanApplicationEvaluationScheduleRunnerOccurrence(rows)
		if scanErr != nil {
			return ApplicationEvaluationOccurrenceRunnerPage{}, scanErr
		}
		values = append(values, occurrence)
	}
	if rows.Err() != nil {
		return ApplicationEvaluationOccurrenceRunnerPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	hasMore := len(values) > limit
	if hasMore {
		values = values[:limit]
	}
	return ApplicationEvaluationOccurrenceRunnerPage{Occurrences: values, HasMore: hasMore}, nil
}

func encodeSQLiteApplicationEvaluationSchedule(schedule ApplicationEvaluationSchedule) ([]byte, int64, any, error) {
	payload, err := json.Marshal(schedule)
	if err != nil {
		return nil, 0, nil, err
	}
	updatedAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(schedule.UpdatedAt)
	if !ok {
		return nil, 0, nil, errApplicationEvaluationScheduleStoreContract
	}
	var nextDueAt any
	if schedule.NextDueAt != nil {
		parsed, parsedOK := parseApplicationEvaluationScheduleUTCTimestamp(*schedule.NextDueAt)
		if !parsedOK {
			return nil, 0, nil, errApplicationEvaluationScheduleStoreContract
		}
		nextDueAt = parsed.UnixNano()
	}
	return payload, updatedAt.UnixNano(), nextDueAt, nil
}

func encodeSQLiteApplicationEvaluationScheduleVersion(version ApplicationEvaluationScheduleVersion) ([]byte, int64, error) {
	payload, err := json.Marshal(version)
	if err != nil {
		return nil, 0, err
	}
	createdAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(version.CreatedAt)
	if !ok {
		return nil, 0, errApplicationEvaluationScheduleStoreContract
	}
	return payload, createdAt.UnixNano(), nil
}

func encodeSQLiteApplicationEvaluationScheduleOccurrence(occurrence ApplicationEvaluationScheduleOccurrence) ([]byte, int64, int64, error) {
	payload, err := json.Marshal(occurrence)
	if err != nil {
		return nil, 0, 0, err
	}
	scheduledAt, scheduledOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.ScheduledForUTC)
	updatedAt, updatedOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.UpdatedAt)
	if !scheduledOK || !updatedOK {
		return nil, 0, 0, errApplicationEvaluationScheduleStoreContract
	}
	return payload, scheduledAt.UnixNano(), updatedAt.UnixNano(), nil
}

func readSQLiteApplicationEvaluationSchedule(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationSchedule, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationSchedule{}, false, nil
		}
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	var value ApplicationEvaluationSchedule
	if decodeStrictApplicationEvaluationJSON(payload, &value) != nil || validateApplicationEvaluationSchedule(ctx, value) != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationSchedule(value), true, nil
}

func readSQLiteApplicationEvaluationScheduleVersion(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationScheduleVersion, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationScheduleVersion{}, false, nil
		}
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	var value ApplicationEvaluationScheduleVersion
	if decodeStrictApplicationEvaluationJSON(payload, &value) != nil || validateApplicationEvaluationScheduleVersion(ctx, value) != nil {
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationScheduleVersion(value), true, nil
}

func readSQLiteApplicationEvaluationScheduleOccurrence(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApplicationEvaluationScheduleOccurrence{}, false, nil
		}
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	var value ApplicationEvaluationScheduleOccurrence
	if decodeStrictApplicationEvaluationJSON(payload, &value) != nil || validateApplicationEvaluationScheduleOccurrence(ctx, value) != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationScheduleOccurrence(value), true, nil
}

func scanSQLiteApplicationEvaluationSchedulePair(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion, error) {
	var schedulePayload, versionPayload []byte
	if err := row.Scan(&schedulePayload, &versionPayload); err != nil {
		return ApplicationEvaluationSchedule{}, ApplicationEvaluationScheduleVersion{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	var schedule ApplicationEvaluationSchedule
	var version ApplicationEvaluationScheduleVersion
	if decodeStrictApplicationEvaluationJSON(schedulePayload, &schedule) != nil || decodeStrictApplicationEvaluationJSON(versionPayload, &version) != nil ||
		validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil {
		return ApplicationEvaluationSchedule{}, ApplicationEvaluationScheduleVersion{}, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationSchedule(schedule), cloneApplicationEvaluationScheduleVersion(version), nil
}

func scanApplicationEvaluationScheduleRunnerPair(row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion, error) {
	var schedulePayload, versionPayload []byte
	if err := row.Scan(&schedulePayload, &versionPayload); err != nil {
		return ApplicationEvaluationSchedule{}, ApplicationEvaluationScheduleVersion{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	var schedule ApplicationEvaluationSchedule
	var version ApplicationEvaluationScheduleVersion
	if decodeStrictApplicationEvaluationJSON(schedulePayload, &schedule) != nil || decodeStrictApplicationEvaluationJSON(versionPayload, &version) != nil {
		return ApplicationEvaluationSchedule{}, ApplicationEvaluationScheduleVersion{}, errApplicationEvaluationScheduleStoreContract
	}
	ctx := applicationEvaluationContextForSchedule(schedule)
	if validateApplicationEvaluationSchedule(ctx, schedule) != nil || validateApplicationEvaluationScheduleVersion(ctx, version) != nil {
		return ApplicationEvaluationSchedule{}, ApplicationEvaluationScheduleVersion{}, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationSchedule(schedule), cloneApplicationEvaluationScheduleVersion(version), nil
}

func scanApplicationEvaluationScheduleRunnerOccurrence(row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationScheduleOccurrence, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	var occurrence ApplicationEvaluationScheduleOccurrence
	if decodeStrictApplicationEvaluationJSON(payload, &occurrence) != nil ||
		validateApplicationEvaluationScheduleOccurrence(applicationEvaluationContextForOccurrence(occurrence), occurrence) != nil {
		return ApplicationEvaluationScheduleOccurrence{}, errApplicationEvaluationScheduleStoreContract
	}
	return cloneApplicationEvaluationScheduleOccurrence(occurrence), nil
}

func optionalApplicationEvaluationScheduleUnixNano(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	parsed, ok := parseApplicationEvaluationScheduleUTCTimestamp(value)
	if !ok {
		return nil, errApplicationEvaluationScheduleStoreContract
	}
	return parsed.UnixNano(), nil
}

func mustApplicationEvaluationScheduleUnixNano(value string) int64 {
	parsed, _ := time.Parse(time.RFC3339Nano, value)
	return parsed.UTC().UnixNano()
}

const sqliteApplicationEvaluationScheduleReadSQL = `SELECT sanitized_schedule_record FROM application_evaluation_schedules
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=?`

const sqliteApplicationEvaluationScheduleVersionReadSQL = `SELECT sanitized_schedule_version_record FROM application_evaluation_schedule_versions
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND schedule_version=?`

const sqliteApplicationEvaluationScheduleOccurrenceReadSQL = `SELECT sanitized_occurrence_record FROM application_evaluation_schedule_occurrences
WHERE tenant_ref=? AND workspace_id=? AND environment=? AND application_id=? AND schedule_id=? AND schedule_version=? AND scheduled_for_unix_nano=?`

var _ applicationEvaluationScheduleRepository = (*sqliteApplicationEvaluationScheduleRepository)(nil)
