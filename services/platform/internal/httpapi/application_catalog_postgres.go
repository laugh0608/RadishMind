package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresApplicationCatalogRepository struct {
	pool *pgxpool.Pool
}

type applicationCatalogRow interface {
	Scan(...any) error
}

func newPostgresApplicationCatalogRepository(pool *pgxpool.Pool) *postgresApplicationCatalogRepository {
	return &postgresApplicationCatalogRepository{pool: pool}
}

func (repository *postgresApplicationCatalogRepository) Create(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	created, err := scanApplicationCatalogRecord(repository.pool.QueryRow(applicationCatalogDatabaseContext(requestContext), `INSERT INTO application_catalog_records
        (tenant_ref, workspace_id, application_id, owner_subject_ref, schema_version, display_name, description,
         application_kind, lifecycle_state, record_version, sanitized_record_payload, created_at, updated_at,
         archived_at, created_by_actor_ref, updated_by_actor_ref, request_id, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
        ON CONFLICT DO NOTHING RETURNING sanitized_record_payload`, record.TenantRef, record.WorkspaceID, record.ApplicationID,
		record.OwnerSubjectRef, record.SchemaVersion, record.DisplayName, record.Description, record.ApplicationKind,
		record.LifecycleState, record.RecordVersion, payload, record.CreatedAt, record.UpdatedAt, record.ArchivedAt,
		record.CreatedByActorRef, record.UpdatedByActorRef, record.RequestID, record.AuditRef))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{}
	}
	return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
}

func (repository *postgresApplicationCatalogRepository) Read(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	if repository == nil || repository.pool == nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	record, err := scanApplicationCatalogRecord(repository.pool.QueryRow(applicationCatalogDatabaseContext(requestContext), `SELECT sanitized_record_payload
        FROM application_catalog_records WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND owner_subject_ref=$4`,
		requestContext.TenantRef, requestContext.WorkspaceID, applicationID, requestContext.OwnerSubjectRef))
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplicationCatalogRecord{}, errApplicationCatalogNotFound
	}
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	return record, nil
}

func (repository *postgresApplicationCatalogRepository) List(requestContext ApplicationCatalogContext, query applicationCatalogListQuery) ([]ApplicationCatalogRecord, error) {
	if repository == nil || repository.pool == nil {
		return nil, errApplicationCatalogStoreUnavailable
	}
	ctx := applicationCatalogDatabaseContext(requestContext)
	var rows pgx.Rows
	var err error
	if query.AfterUpdatedAt == "" {
		rows, err = repository.pool.Query(ctx, `SELECT sanitized_record_payload FROM application_catalog_records
            WHERE tenant_ref=$1 AND workspace_id=$2 AND owner_subject_ref=$3 AND lifecycle_state=$4
              AND ($5='' OR application_kind=$5)
            ORDER BY updated_at DESC, application_id DESC LIMIT $6`, requestContext.TenantRef, requestContext.WorkspaceID,
			requestContext.OwnerSubjectRef, query.LifecycleState, query.ApplicationKind, query.Limit)
	} else {
		rows, err = repository.pool.Query(ctx, `SELECT sanitized_record_payload FROM application_catalog_records
            WHERE tenant_ref=$1 AND workspace_id=$2 AND owner_subject_ref=$3 AND lifecycle_state=$4
              AND ($5='' OR application_kind=$5) AND (updated_at, application_id) < ($6::timestamptz, $7)
            ORDER BY updated_at DESC, application_id DESC LIMIT $8`, requestContext.TenantRef, requestContext.WorkspaceID,
			requestContext.OwnerSubjectRef, query.LifecycleState, query.ApplicationKind, query.AfterUpdatedAt,
			query.AfterApplicationID, query.Limit)
	}
	if err != nil {
		return nil, errApplicationCatalogStoreUnavailable
	}
	defer rows.Close()
	records := make([]ApplicationCatalogRecord, 0)
	for rows.Next() {
		record, scanErr := scanApplicationCatalogRecord(rows)
		if scanErr != nil {
			return nil, errApplicationCatalogStoreUnavailable
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return nil, errApplicationCatalogStoreUnavailable
	}
	return records, nil
}

func (repository *postgresApplicationCatalogRepository) UpdateMetadata(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	current, err := repository.Read(requestContext, applicationID)
	if err != nil {
		return ApplicationCatalogRecord{}, err
	}
	if current.RecordVersion != expectedVersion {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == applicationCatalogLifecycleArchived {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	current.DisplayName = update.DisplayName
	current.Description = update.Description
	current.ApplicationKind = update.ApplicationKind
	current.RecordVersion++
	current.UpdatedAt = update.UpdatedAt
	current.UpdatedByActorRef = update.UpdatedByActorRef
	current.RequestID = update.RequestID
	current.AuditRef = update.AuditRef
	return repository.persistMutation(requestContext, current, expectedVersion, applicationCatalogLifecycleActive)
}

func (repository *postgresApplicationCatalogRepository) Archive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	current, err := repository.Read(requestContext, applicationID)
	if err != nil {
		return ApplicationCatalogRecord{}, err
	}
	if current.RecordVersion != expectedVersion {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState != applicationCatalogLifecycleActive {
		return ApplicationCatalogRecord{}, errApplicationCatalogTransitionInvalid
	}
	current.LifecycleState = applicationCatalogLifecycleArchived
	current.RecordVersion++
	current.UpdatedAt = update.UpdatedAt
	current.ArchivedAt = update.ArchivedAt
	current.UpdatedByActorRef = update.UpdatedByActorRef
	current.RequestID = update.RequestID
	current.AuditRef = update.AuditRef
	return repository.persistMutation(requestContext, current, expectedVersion, applicationCatalogLifecycleActive)
}

func (repository *postgresApplicationCatalogRepository) RequireActive(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	record, err := repository.Read(requestContext, applicationID)
	if err != nil {
		return ApplicationCatalogRecord{}, err
	}
	if record.LifecycleState != applicationCatalogLifecycleActive {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	return record, nil
}

func (repository *postgresApplicationCatalogRepository) persistMutation(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord, expectedVersion int, expectedLifecycle string) (ApplicationCatalogRecord, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	updated, err := scanApplicationCatalogRecord(repository.pool.QueryRow(applicationCatalogDatabaseContext(requestContext), `UPDATE application_catalog_records SET
        display_name=$1, description=$2, application_kind=$3, lifecycle_state=$4, record_version=$5,
        sanitized_record_payload=$6, updated_at=$7, archived_at=$8, updated_by_actor_ref=$9, request_id=$10, audit_ref=$11
        WHERE tenant_ref=$12 AND workspace_id=$13 AND application_id=$14 AND owner_subject_ref=$15
          AND record_version=$16 AND lifecycle_state=$17 RETURNING sanitized_record_payload`, record.DisplayName, record.Description,
		record.ApplicationKind, record.LifecycleState, record.RecordVersion, payload, record.UpdatedAt, record.ArchivedAt,
		record.UpdatedByActorRef, record.RequestID, record.AuditRef, requestContext.TenantRef, requestContext.WorkspaceID,
		record.ApplicationID, requestContext.OwnerSubjectRef, expectedVersion, expectedLifecycle))
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	latest, readErr := repository.Read(requestContext, record.ApplicationID)
	if readErr != nil {
		return ApplicationCatalogRecord{}, readErr
	}
	if latest.RecordVersion != expectedVersion {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: latest.RecordVersion, CurrentState: latest.LifecycleState}
	}
	if latest.LifecycleState == applicationCatalogLifecycleArchived {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
}

func scanApplicationCatalogRecord(row applicationCatalogRow) (ApplicationCatalogRecord, error) {
	var payload []byte
	if err := row.Scan(&payload); err != nil {
		return ApplicationCatalogRecord{}, err
	}
	var record ApplicationCatalogRecord
	if err := json.Unmarshal(payload, &record); err != nil || record.SchemaVersion != applicationCatalogSchemaVersion ||
		!applicationCatalogIDPattern.MatchString(record.ApplicationID) || record.RecordVersion < 1 ||
		(record.LifecycleState != applicationCatalogLifecycleActive && record.LifecycleState != applicationCatalogLifecycleArchived) ||
		strings.TrimSpace(record.TenantRef) == "" || strings.TrimSpace(record.WorkspaceID) == "" || strings.TrimSpace(record.OwnerSubjectRef) == "" {
		return ApplicationCatalogRecord{}, errors.New("stored application catalog record contract mismatch")
	}
	return record, nil
}

func applicationCatalogDatabaseContext(requestContext ApplicationCatalogContext) context.Context {
	if requestContext.RequestContext != nil {
		return requestContext.RequestContext
	}
	return context.Background()
}
