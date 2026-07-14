package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteApplicationCatalogRepository struct {
	database *sql.DB
}

func newSQLiteApplicationCatalogRepository(database *sql.DB) *sqliteApplicationCatalogRepository {
	return &sqliteApplicationCatalogRepository{database: database}
}

func (repository *sqliteApplicationCatalogRepository) Create(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	if repository == nil || repository.database == nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	created, err := scanApplicationCatalogRecord(repository.database.QueryRowContext(applicationCatalogDatabaseContext(requestContext), `INSERT INTO application_catalog_records
        (tenant_ref, workspace_id, application_id, owner_subject_ref, schema_version, display_name, description,
         application_kind, lifecycle_state, record_version, sanitized_record_payload, created_at, updated_at,
         archived_at, created_by_actor_ref, updated_by_actor_ref, request_id, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT DO NOTHING RETURNING sanitized_record_payload`, record.TenantRef, record.WorkspaceID, record.ApplicationID,
		record.OwnerSubjectRef, record.SchemaVersion, record.DisplayName, record.Description, record.ApplicationKind,
		record.LifecycleState, record.RecordVersion, string(payload), record.CreatedAt, record.UpdatedAt, record.ArchivedAt,
		record.CreatedByActorRef, record.UpdatedByActorRef, record.RequestID, record.AuditRef))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{}
	}
	return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
}

func (repository *sqliteApplicationCatalogRepository) Read(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	if repository == nil || repository.database == nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	record, err := scanApplicationCatalogRecord(repository.database.QueryRowContext(applicationCatalogDatabaseContext(requestContext), `SELECT sanitized_record_payload
        FROM application_catalog_records WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, applicationID, requestContext.OwnerSubjectRef))
	if errors.Is(err, sql.ErrNoRows) {
		return ApplicationCatalogRecord{}, errApplicationCatalogNotFound
	}
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	return record, nil
}

func (repository *sqliteApplicationCatalogRepository) List(requestContext ApplicationCatalogContext, query applicationCatalogListQuery) ([]ApplicationCatalogRecord, error) {
	if repository == nil || repository.database == nil {
		return nil, errApplicationCatalogStoreUnavailable
	}
	ctx := applicationCatalogDatabaseContext(requestContext)
	var rows *sql.Rows
	var err error
	if query.AfterUpdatedAt == "" {
		rows, err = repository.database.QueryContext(ctx, `SELECT sanitized_record_payload FROM application_catalog_records
            WHERE tenant_ref=? AND workspace_id=? AND owner_subject_ref=? AND lifecycle_state=?
              AND (?='' OR application_kind=?)
            ORDER BY updated_at DESC, application_id DESC LIMIT ?`, requestContext.TenantRef, requestContext.WorkspaceID,
			requestContext.OwnerSubjectRef, query.LifecycleState, query.ApplicationKind, query.ApplicationKind, query.Limit)
	} else {
		rows, err = repository.database.QueryContext(ctx, `SELECT sanitized_record_payload FROM application_catalog_records
            WHERE tenant_ref=? AND workspace_id=? AND owner_subject_ref=? AND lifecycle_state=?
              AND (?='' OR application_kind=?)
              AND (updated_at < ? OR (updated_at = ? AND application_id < ?))
            ORDER BY updated_at DESC, application_id DESC LIMIT ?`, requestContext.TenantRef, requestContext.WorkspaceID,
			requestContext.OwnerSubjectRef, query.LifecycleState, query.ApplicationKind, query.ApplicationKind,
			query.AfterUpdatedAt, query.AfterUpdatedAt, query.AfterApplicationID, query.Limit)
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

func (repository *sqliteApplicationCatalogRepository) UpdateMetadata(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
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

func (repository *sqliteApplicationCatalogRepository) Archive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
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

func (repository *sqliteApplicationCatalogRepository) RequireActive(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	record, err := repository.Read(requestContext, applicationID)
	if err != nil {
		return ApplicationCatalogRecord{}, err
	}
	if record.LifecycleState != applicationCatalogLifecycleActive {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	return record, nil
}

func (repository *sqliteApplicationCatalogRepository) persistMutation(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord, expectedVersion int, expectedLifecycle string) (ApplicationCatalogRecord, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	updated, err := scanApplicationCatalogRecord(repository.database.QueryRowContext(applicationCatalogDatabaseContext(requestContext), `UPDATE application_catalog_records SET
        display_name=?, description=?, application_kind=?, lifecycle_state=?, record_version=?,
        sanitized_record_payload=?, updated_at=?, archived_at=?, updated_by_actor_ref=?, request_id=?, audit_ref=?
        WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=?
          AND record_version=? AND lifecycle_state=? RETURNING sanitized_record_payload`, record.DisplayName, record.Description,
		record.ApplicationKind, record.LifecycleState, record.RecordVersion, string(payload), record.UpdatedAt, record.ArchivedAt,
		record.UpdatedByActorRef, record.RequestID, record.AuditRef, requestContext.TenantRef, requestContext.WorkspaceID,
		record.ApplicationID, requestContext.OwnerSubjectRef, expectedVersion, expectedLifecycle))
	if err == nil {
		return updated, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
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
