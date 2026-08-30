package httpapi

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationEvaluationScheduleRepository struct{ pool *pgxpool.Pool }

func newPostgresApplicationEvaluationScheduleRepository(pool *pgxpool.Pool) *postgresApplicationEvaluationScheduleRepository {
	return &postgresApplicationEvaluationScheduleRepository{pool: pool}
}

func (repository *postgresApplicationEvaluationScheduleRepository) CreateSchedule(
	ctx ApplicationEvaluationContext,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) error {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationScheduleCreate(ctx, schedule, version) {
		return errApplicationEvaluationScheduleStoreContract
	}
	schedulePayload, updatedAt, nextDueAt, err := encodePostgresApplicationEvaluationSchedule(schedule)
	if err != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	versionPayload, createdAt, err := encodePostgresApplicationEvaluationScheduleVersion(version)
	if err != nil {
		return errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	if _, found, readErr := readPostgresApplicationEvaluationSchedule(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID)); readErr != nil {
		return readErr
	} else if found {
		return errApplicationEvaluationScheduleVersionConflict
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_schedules
(tenant_ref,workspace_id,environment,application_id,schedule_id,record_version,latest_schedule_version,latest_schedule_digest,lifecycle_state,updated_at,next_due_at,sanitized_schedule_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		schedule.ScheduleID, schedule.RecordVersion, schedule.LatestScheduleVersion, schedule.LatestScheduleDigest,
		schedule.LifecycleState, updatedAt, nextDueAt, schedulePayload); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_versions
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest,created_at,sanitized_schedule_version_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		version.ScheduleID, version.ScheduleVersion, version.ScheduleDigest, createdAt, versionPayload); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return errApplicationEvaluationScheduleStoreUnavailable
	}
	return nil
}

func (repository *postgresApplicationEvaluationScheduleRepository) ReviseSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
	version ApplicationEvaluationScheduleVersion,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationScheduleRevision(ctx, expected, schedule, version) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	current, found, err := readPostgresApplicationEvaluationSchedule(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleReadForUpdateSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, false, err
	}
	currentVersion, versionFound, err := readPostgresApplicationEvaluationScheduleVersion(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		current.ScheduleID, current.LatestScheduleVersion))
	if err != nil || !versionFound {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	if !validApplicationEvaluationScheduleRevisionAgainstCurrent(current, currentVersion, schedule, version) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	schedulePayload, updatedAt, nextDueAt, err := encodePostgresApplicationEvaluationSchedule(schedule)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	versionPayload, createdAt, err := encodePostgresApplicationEvaluationScheduleVersion(version)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_versions
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest,created_at,sanitized_schedule_version_record)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		version.ScheduleID, version.ScheduleVersion, version.ScheduleDigest, createdAt, versionPayload); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_schedules SET
record_version=$1,latest_schedule_version=$2,latest_schedule_digest=$3,lifecycle_state=$4,updated_at=$5,next_due_at=$6,sanitized_schedule_record=$7
WHERE tenant_ref=$8 AND workspace_id=$9 AND environment=$10 AND application_id=$11 AND schedule_id=$12 AND record_version=$13`,
		schedule.RecordVersion, schedule.LatestScheduleVersion, schedule.LatestScheduleDigest, schedule.LifecycleState, updatedAt, nextDueAt, schedulePayload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID, expected)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *postgresApplicationEvaluationScheduleRepository) UpdateSchedule(
	ctx ApplicationEvaluationContext,
	expected int,
	schedule ApplicationEvaluationSchedule,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationScheduleUpdateInput(ctx, expected, schedule) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	current, found, err := readPostgresApplicationEvaluationSchedule(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleReadForUpdateSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, false, err
	}
	version, versionFound, err := readPostgresApplicationEvaluationScheduleVersion(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		current.ScheduleID, current.LatestScheduleVersion))
	if err != nil || !versionFound {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	if !validApplicationEvaluationScheduleUpdateAgainstCurrent(ctx, current, version, schedule) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	payload, updatedAt, nextDueAt, err := encodePostgresApplicationEvaluationSchedule(schedule)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_schedules SET
record_version=$1,lifecycle_state=$2,updated_at=$3,next_due_at=$4,sanitized_schedule_record=$5
WHERE tenant_ref=$6 AND workspace_id=$7 AND environment=$8 AND application_id=$9 AND schedule_id=$10 AND record_version=$11`,
		schedule.RecordVersion, schedule.LifecycleState, updatedAt, nextDueAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, schedule.ScheduleID, expected)
	if err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationSchedule(schedule), true, nil
}

func (repository *postgresApplicationEvaluationScheduleRepository) ReadSchedule(
	ctx ApplicationEvaluationContext,
	scheduleID string,
) (ApplicationEvaluationSchedule, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) {
		return ApplicationEvaluationSchedule{}, false, errApplicationEvaluationScheduleStoreContract
	}
	schedule, found, err := readPostgresApplicationEvaluationSchedule(ctx, repository.pool.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, scheduleID))
	if err != nil || !found {
		return ApplicationEvaluationSchedule{}, found, err
	}
	version, versionFound, err := repository.ReadScheduleVersion(ctx, scheduleID, schedule.LatestScheduleVersion)
	if err != nil || !versionFound || !applicationEvaluationScheduleMatchesVersion(schedule, version) {
		return ApplicationEvaluationSchedule{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	return schedule, true, nil
}

func (repository *postgresApplicationEvaluationScheduleRepository) ListSchedules(
	ctx ApplicationEvaluationContext,
	filter ApplicationEvaluationScheduleListFilter,
) (ApplicationEvaluationScheduleListPage, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) ||
		!validApplicationEvaluationScheduleState(filter.LifecycleState) || filter.Limit < 1 {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreContract
	}
	before, err := optionalApplicationEvaluationScheduleTimestamp(filter.BeforeUpdatedAt)
	if err != nil {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreContract
	}
	rows, err := repository.pool.Query(ctx.RequestContext, `SELECT schedules.sanitized_schedule_record,versions.sanitized_schedule_version_record
FROM application_evaluation_schedules schedules JOIN application_evaluation_schedule_versions versions
ON versions.tenant_ref=schedules.tenant_ref AND versions.workspace_id=schedules.workspace_id AND versions.environment=schedules.environment
AND versions.application_id=schedules.application_id AND versions.schedule_id=schedules.schedule_id AND versions.schedule_version=schedules.latest_schedule_version
WHERE schedules.tenant_ref=$1 AND schedules.workspace_id=$2 AND schedules.environment=$3 AND schedules.application_id=$4 AND schedules.lifecycle_state=$5
AND ($6::timestamptz IS NULL OR schedules.updated_at<$6 OR (schedules.updated_at=$6 AND schedules.schedule_id<$7))
ORDER BY schedules.updated_at DESC,schedules.schedule_id DESC LIMIT $8`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, filter.LifecycleState,
		before, filter.BeforeScheduleID, filter.Limit+1)
	if err != nil {
		return ApplicationEvaluationScheduleListPage{}, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer rows.Close()
	values := make([]ApplicationEvaluationSchedule, 0, filter.Limit+1)
	for rows.Next() {
		schedule, version, scanErr := scanPostgresApplicationEvaluationSchedulePair(ctx, rows)
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

func (repository *postgresApplicationEvaluationScheduleRepository) ReadScheduleVersion(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	versionNumber int,
) (ApplicationEvaluationScheduleVersion, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) || versionNumber < 1 {
		return ApplicationEvaluationScheduleVersion{}, false, errApplicationEvaluationScheduleStoreContract
	}
	return readPostgresApplicationEvaluationScheduleVersion(ctx, repository.pool.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, scheduleID, versionNumber))
}

func (repository *postgresApplicationEvaluationScheduleRepository) ClaimOccurrence(
	ctx ApplicationEvaluationContext,
	due ApplicationEvaluationScheduleOccurrence,
	claimed ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationScheduleClaim(ctx, due, claimed) {
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
	payload, scheduledAt, updatedAt, err := encodePostgresApplicationEvaluationScheduleOccurrence(claimed)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	command, err := repository.pool.Exec(ctx.RequestContext, `INSERT INTO application_evaluation_schedule_occurrences
(tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,scheduled_for,record_version,schedule_digest,occurrence_state,client_campaign_key,campaign_id,updated_at,sanitized_occurrence_record)
SELECT $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14
WHERE EXISTS (SELECT 1 FROM application_evaluation_schedules WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND schedule_id=$5
AND lifecycle_state='active' AND latest_schedule_version=$6 AND latest_schedule_digest=$9 AND next_due_at=$7)
AND EXISTS (SELECT 1 FROM application_evaluation_schedule_versions WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4
AND schedule_id=$5 AND schedule_version=$6 AND schedule_digest=$9)
ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, claimed.ScheduleID,
		claimed.ScheduleVersion, scheduledAt, claimed.RecordVersion, claimed.ScheduleDigest, claimed.State,
		claimed.ClientCampaignKey, claimed.CampaignID, updatedAt, payload)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if command.RowsAffected() == 1 {
		return cloneApplicationEvaluationScheduleOccurrence(claimed), true, nil
	}
	current, exists, readErr := repository.ReadOccurrence(ctx, due.ScheduleID, due.ScheduleVersion, due.ScheduledForUTC)
	if readErr != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, readErr
	}
	if exists {
		return current, false, errApplicationEvaluationScheduleClaimConflict
	}
	return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
}

func (repository *postgresApplicationEvaluationScheduleRepository) UpdateOccurrence(
	ctx ApplicationEvaluationContext,
	expected int,
	occurrence ApplicationEvaluationScheduleOccurrence,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationScheduleOccurrenceUpdateInput(ctx, expected, occurrence) {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	scheduledAt, _ := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.ScheduledForUTC)
	tx, err := repository.pool.Begin(ctx.RequestContext)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	current, found, err := readPostgresApplicationEvaluationScheduleOccurrence(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleOccurrenceReadForUpdateSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		occurrence.ScheduleID, occurrence.ScheduleVersion, scheduledAt))
	if err != nil || !found {
		return ApplicationEvaluationScheduleOccurrence{}, false, err
	}
	if current.RecordVersion != expected {
		return current, false, errApplicationEvaluationScheduleVersionConflict
	}
	version, versionFound, err := readPostgresApplicationEvaluationScheduleVersion(ctx, tx.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleVersionReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		occurrence.ScheduleID, occurrence.ScheduleVersion))
	if err != nil || !versionFound || !validApplicationEvaluationScheduleOccurrenceAgainstVersion(current, occurrence, version) {
		return ApplicationEvaluationScheduleOccurrence{}, false, firstApplicationEvaluationScheduleStoreError(err)
	}
	payload, _, updatedAt, err := encodePostgresApplicationEvaluationScheduleOccurrence(occurrence)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE application_evaluation_schedule_occurrences SET
record_version=$1,occurrence_state=$2,campaign_id=$3,updated_at=$4,sanitized_occurrence_record=$5
WHERE tenant_ref=$6 AND workspace_id=$7 AND environment=$8 AND application_id=$9 AND schedule_id=$10 AND schedule_version=$11 AND scheduled_for=$12 AND record_version=$13`,
		occurrence.RecordVersion, occurrence.State, occurrence.CampaignID, updatedAt, payload,
		ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, occurrence.ScheduleID, occurrence.ScheduleVersion, scheduledAt, expected)
	if err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	if command.RowsAffected() != 1 {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreUnavailable
	}
	return cloneApplicationEvaluationScheduleOccurrence(occurrence), true, nil
}

func (repository *postgresApplicationEvaluationScheduleRepository) ReadOccurrence(
	ctx ApplicationEvaluationContext,
	scheduleID string,
	scheduleVersion int,
	scheduledForUTC string,
) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	if repository == nil || repository.pool == nil || !validApplicationEvaluationContext(ctx) || scheduleVersion < 1 {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	scheduledAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(scheduledForUTC)
	if !ok {
		return ApplicationEvaluationScheduleOccurrence{}, false, errApplicationEvaluationScheduleStoreContract
	}
	occurrence, found, err := readPostgresApplicationEvaluationScheduleOccurrence(ctx, repository.pool.QueryRow(ctx.RequestContext,
		postgresApplicationEvaluationScheduleOccurrenceReadSQL, ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID,
		scheduleID, scheduleVersion, scheduledAt))
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

func encodePostgresApplicationEvaluationSchedule(schedule ApplicationEvaluationSchedule) ([]byte, time.Time, *time.Time, error) {
	payload, err := json.Marshal(schedule)
	if err != nil {
		return nil, time.Time{}, nil, err
	}
	updatedAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(schedule.UpdatedAt)
	if !ok {
		return nil, time.Time{}, nil, errApplicationEvaluationScheduleStoreContract
	}
	var nextDueAt *time.Time
	if schedule.NextDueAt != nil {
		parsed, parsedOK := parseApplicationEvaluationScheduleUTCTimestamp(*schedule.NextDueAt)
		if !parsedOK {
			return nil, time.Time{}, nil, errApplicationEvaluationScheduleStoreContract
		}
		nextDueAt = &parsed
	}
	return payload, updatedAt, nextDueAt, nil
}

func encodePostgresApplicationEvaluationScheduleVersion(version ApplicationEvaluationScheduleVersion) ([]byte, time.Time, error) {
	payload, err := json.Marshal(version)
	if err != nil {
		return nil, time.Time{}, err
	}
	createdAt, ok := parseApplicationEvaluationScheduleUTCTimestamp(version.CreatedAt)
	if !ok {
		return nil, time.Time{}, errApplicationEvaluationScheduleStoreContract
	}
	return payload, createdAt, nil
}

func encodePostgresApplicationEvaluationScheduleOccurrence(occurrence ApplicationEvaluationScheduleOccurrence) ([]byte, time.Time, time.Time, error) {
	payload, err := json.Marshal(occurrence)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}
	scheduledAt, scheduledOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.ScheduledForUTC)
	updatedAt, updatedOK := parseApplicationEvaluationScheduleUTCTimestamp(occurrence.UpdatedAt)
	if !scheduledOK || !updatedOK {
		return nil, time.Time{}, time.Time{}, errApplicationEvaluationScheduleStoreContract
	}
	return payload, scheduledAt, updatedAt, nil
}

func readPostgresApplicationEvaluationSchedule(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationSchedule, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func readPostgresApplicationEvaluationScheduleVersion(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationScheduleVersion, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func readPostgresApplicationEvaluationScheduleOccurrence(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationScheduleOccurrence, bool, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func scanPostgresApplicationEvaluationSchedulePair(ctx ApplicationEvaluationContext, row applicationEvaluationScheduleSQLScanner) (ApplicationEvaluationSchedule, ApplicationEvaluationScheduleVersion, error) {
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

func optionalApplicationEvaluationScheduleTimestamp(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, ok := parseApplicationEvaluationScheduleUTCTimestamp(value)
	if !ok {
		return nil, errApplicationEvaluationScheduleStoreContract
	}
	return &parsed, nil
}

const postgresApplicationEvaluationScheduleReadSQL = `SELECT sanitized_schedule_record FROM application_evaluation_schedules
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND schedule_id=$5`

const postgresApplicationEvaluationScheduleReadForUpdateSQL = postgresApplicationEvaluationScheduleReadSQL + ` FOR UPDATE`

const postgresApplicationEvaluationScheduleVersionReadSQL = `SELECT sanitized_schedule_version_record FROM application_evaluation_schedule_versions
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND schedule_id=$5 AND schedule_version=$6`

const postgresApplicationEvaluationScheduleOccurrenceReadSQL = `SELECT sanitized_occurrence_record FROM application_evaluation_schedule_occurrences
WHERE tenant_ref=$1 AND workspace_id=$2 AND environment=$3 AND application_id=$4 AND schedule_id=$5 AND schedule_version=$6 AND scheduled_for=$7`

const postgresApplicationEvaluationScheduleOccurrenceReadForUpdateSQL = postgresApplicationEvaluationScheduleOccurrenceReadSQL + ` FOR UPDATE`

var _ applicationEvaluationScheduleRepository = (*postgresApplicationEvaluationScheduleRepository)(nil)
