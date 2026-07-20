package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

type sqliteAPIKeyRepository struct {
	database *sql.DB
}

func newSQLiteAPIKeyRepository(database *sql.DB) *sqliteAPIKeyRepository {
	return &sqliteAPIKeyRepository{database: database}
}

func (repository *sqliteAPIKeyRepository) Create(requestContext APIKeyContext, record APIKeyRecord) (APIKeyRecord, error) {
	if repository == nil || repository.database == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	payload, scopes, createdAt, expiresAt, lastUsedAt, revokedAt, err := sqliteAPIKeyStorageValues(record)
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	created, err := scanSQLiteAPIKeyRecord(repository.database.QueryRowContext(apiKeyDatabaseContext(requestContext.RequestContext), `INSERT INTO api_key_records
        (tenant_ref, workspace_id, api_key_id, application_id, owner_subject_ref, schema_version, display_name,
         scopes_json, lifecycle_state, record_version, credential_digest, sanitized_record_payload,
         created_at_unix_nano, expires_at_unix_nano, last_used_at_unix_nano, revoked_at_unix_nano,
         created_by_actor_ref, revoked_by_actor_ref, request_id, audit_ref)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT DO NOTHING
        RETURNING sanitized_record_payload, credential_digest, last_used_at_unix_nano`, record.TenantRef,
		record.WorkspaceID, record.APIKeyID, record.ApplicationID, record.OwnerSubjectRef, record.SchemaVersion,
		record.DisplayName, scopes, record.LifecycleState, record.RecordVersion, record.credentialDigest[:], payload,
		createdAt, expiresAt, lastUsedAt, revokedAt, record.CreatedByActorRef, record.RevokedByActorRef,
		record.RequestID, record.AuditRef))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyIdentifierCollision
	}
	return APIKeyRecord{}, errAPIKeyStoreUnavailable
}

func (repository *sqliteAPIKeyRepository) Read(requestContext APIKeyContext, apiKeyID string) (APIKeyRecord, error) {
	if repository == nil || repository.database == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	record, err := scanSQLiteAPIKeyRecord(repository.database.QueryRowContext(apiKeyDatabaseContext(requestContext.RequestContext), `SELECT sanitized_record_payload, credential_digest, last_used_at_unix_nano
        FROM api_key_records WHERE tenant_ref=? AND workspace_id=? AND api_key_id=? AND owner_subject_ref=?`,
		requestContext.TenantRef, requestContext.WorkspaceID, apiKeyID, requestContext.OwnerSubjectRef))
	if errors.Is(err, sql.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	return record, nil
}

func (repository *sqliteAPIKeyRepository) List(requestContext APIKeyContext, query apiKeyListQuery) ([]APIKeyRecord, error) {
	if repository == nil || repository.database == nil {
		return nil, errAPIKeyStoreUnavailable
	}
	ctx := apiKeyDatabaseContext(requestContext.RequestContext)
	now := query.Now.UTC().UnixNano()
	var rows *sql.Rows
	var err error
	if query.AfterCreatedAt == "" {
		rows, err = repository.database.QueryContext(ctx, `SELECT sanitized_record_payload, credential_digest, last_used_at_unix_nano
            FROM api_key_records
            WHERE tenant_ref=? AND workspace_id=? AND owner_subject_ref=?
              AND (?='' OR application_id=?)
              AND (?='' OR (CASE WHEN lifecycle_state='revoked' THEN 'revoked'
                  WHEN expires_at_unix_nano <= ? THEN 'expired' ELSE 'active' END)=?)
            ORDER BY created_at_unix_nano DESC, api_key_id DESC LIMIT ?`, requestContext.TenantRef,
			requestContext.WorkspaceID, requestContext.OwnerSubjectRef, query.ApplicationID, query.ApplicationID,
			query.EffectiveState, now, query.EffectiveState, query.Limit)
	} else {
		afterCreatedAt, parseErr := apiKeyUnixNano(query.AfterCreatedAt)
		if parseErr != nil {
			return nil, errAPIKeyStoreUnavailable
		}
		rows, err = repository.database.QueryContext(ctx, `SELECT sanitized_record_payload, credential_digest, last_used_at_unix_nano
            FROM api_key_records
            WHERE tenant_ref=? AND workspace_id=? AND owner_subject_ref=?
              AND (?='' OR application_id=?)
              AND (?='' OR (CASE WHEN lifecycle_state='revoked' THEN 'revoked'
                  WHEN expires_at_unix_nano <= ? THEN 'expired' ELSE 'active' END)=?)
              AND (created_at_unix_nano < ? OR (created_at_unix_nano = ? AND api_key_id < ?))
            ORDER BY created_at_unix_nano DESC, api_key_id DESC LIMIT ?`, requestContext.TenantRef,
			requestContext.WorkspaceID, requestContext.OwnerSubjectRef, query.ApplicationID, query.ApplicationID,
			query.EffectiveState, now, query.EffectiveState, afterCreatedAt, afterCreatedAt, query.AfterAPIKeyID,
			query.Limit)
	}
	if err != nil {
		return nil, errAPIKeyStoreUnavailable
	}
	defer rows.Close()
	records := make([]APIKeyRecord, 0)
	for rows.Next() {
		record, scanErr := scanSQLiteAPIKeyRecord(rows)
		if scanErr != nil {
			return nil, errAPIKeyStoreUnavailable
		}
		records = append(records, record)
	}
	if rows.Err() != nil {
		return nil, errAPIKeyStoreUnavailable
	}
	return records, nil
}

func (repository *sqliteAPIKeyRepository) FindCredential(requestContext context.Context, apiKeyID string) (APIKeyRecord, error) {
	if repository == nil || repository.database == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	record, err := scanSQLiteAPIKeyRecord(repository.database.QueryRowContext(apiKeyDatabaseContext(requestContext), `SELECT sanitized_record_payload, credential_digest, last_used_at_unix_nano
        FROM api_key_records WHERE api_key_id=?`, apiKeyID))
	if errors.Is(err, sql.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	return record, nil
}

func (repository *sqliteAPIKeyRepository) RecordSuccessfulAuthentication(requestContext context.Context, apiKeyID string, expectedVersion int, usedAt time.Time) (APIKeyRecord, error) {
	if repository == nil || repository.database == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	usedAt = usedAt.UTC()
	usedAtUnixNano := usedAt.UnixNano()
	record, err := scanSQLiteAPIKeyRecord(repository.database.QueryRowContext(apiKeyDatabaseContext(requestContext), `UPDATE api_key_records
        SET last_used_at_unix_nano=CASE WHEN last_used_at_unix_nano IS NULL OR last_used_at_unix_nano < ?
            THEN ? ELSE last_used_at_unix_nano END
        WHERE api_key_id=? AND record_version=? AND lifecycle_state='active' AND expires_at_unix_nano > ?
        RETURNING sanitized_record_payload, credential_digest, last_used_at_unix_nano`, usedAtUnixNano,
		usedAtUnixNano, apiKeyID, expectedVersion, usedAtUnixNano))
	if err == nil {
		return record, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	latest, findErr := repository.FindCredential(requestContext, apiKeyID)
	if findErr != nil {
		return APIKeyRecord{}, findErr
	}
	if latest.LifecycleState == apiKeyLifecycleRevoked {
		return APIKeyRecord{}, errAPIKeyRevoked
	}
	if effectiveAPIKeyState(latest, usedAt) == apiKeyEffectiveExpired {
		return APIKeyRecord{}, errAPIKeyExpired
	}
	if latest.RecordVersion != expectedVersion {
		return APIKeyRecord{}, apiKeyVersionConflictError{CurrentVersion: latest.RecordVersion, CurrentState: latest.LifecycleState}
	}
	return APIKeyRecord{}, errAPIKeyStoreUnavailable
}

func (repository *sqliteAPIKeyRepository) Revoke(requestContext APIKeyContext, apiKeyID string, expectedVersion int, update APIKeyRecord) (APIKeyRecord, error) {
	current, err := repository.Read(requestContext, apiKeyID)
	if err != nil {
		return APIKeyRecord{}, err
	}
	if current.RecordVersion != expectedVersion {
		return APIKeyRecord{}, apiKeyVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == apiKeyLifecycleRevoked {
		return APIKeyRecord{}, errAPIKeyRevoked
	}
	if update.LifecycleState != apiKeyLifecycleRevoked || update.RevokedAt == nil || update.RevokedByActorRef == nil {
		return APIKeyRecord{}, errAPIKeyTransitionInvalid
	}
	current.LifecycleState = apiKeyLifecycleRevoked
	current.EffectiveState = apiKeyLifecycleRevoked
	current.RecordVersion++
	current.RevokedAt = update.RevokedAt
	current.RevokedByActorRef = update.RevokedByActorRef
	current.RequestID = update.RequestID
	current.AuditRef = update.AuditRef
	payload, err := json.Marshal(current)
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	revokedAt, err := optionalAPIKeyUnixNano(current.RevokedAt)
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	revoked, err := scanSQLiteAPIKeyRecord(repository.database.QueryRowContext(apiKeyDatabaseContext(requestContext.RequestContext), `UPDATE api_key_records SET
        lifecycle_state='revoked', record_version=?, sanitized_record_payload=?, revoked_at_unix_nano=?,
        revoked_by_actor_ref=?, request_id=?, audit_ref=?
        WHERE tenant_ref=? AND workspace_id=? AND api_key_id=? AND owner_subject_ref=?
          AND record_version=? AND lifecycle_state='active'
        RETURNING sanitized_record_payload, credential_digest, last_used_at_unix_nano`, current.RecordVersion,
		string(payload), revokedAt, current.RevokedByActorRef, current.RequestID, current.AuditRef,
		requestContext.TenantRef, requestContext.WorkspaceID, apiKeyID, requestContext.OwnerSubjectRef, expectedVersion))
	if err == nil {
		return revoked, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	latest, readErr := repository.Read(requestContext, apiKeyID)
	if readErr != nil {
		return APIKeyRecord{}, readErr
	}
	if latest.RecordVersion != expectedVersion {
		return APIKeyRecord{}, apiKeyVersionConflictError{CurrentVersion: latest.RecordVersion, CurrentState: latest.LifecycleState}
	}
	if latest.LifecycleState == apiKeyLifecycleRevoked {
		return APIKeyRecord{}, errAPIKeyRevoked
	}
	return APIKeyRecord{}, errAPIKeyStoreUnavailable
}

func scanSQLiteAPIKeyRecord(row apiKeyRow) (APIKeyRecord, error) {
	var payload []byte
	var digest []byte
	var lastUsedAt sql.NullInt64
	if err := row.Scan(&payload, &digest, &lastUsedAt); err != nil {
		return APIKeyRecord{}, err
	}
	var record APIKeyRecord
	if err := json.Unmarshal(payload, &record); err != nil || len(digest) != sha256.Size || !validStoredAPIKeyRecord(record) {
		return APIKeyRecord{}, errors.New("stored API key record contract mismatch")
	}
	copy(record.credentialDigest[:], digest)
	if record.credentialDigest == ([sha256.Size]byte{}) {
		return APIKeyRecord{}, errors.New("stored API key credential digest is empty")
	}
	if !lastUsedAt.Valid {
		record.LastUsedAt = nil
	} else {
		value := time.Unix(0, lastUsedAt.Int64).UTC().Format(time.RFC3339Nano)
		record.LastUsedAt = &value
	}
	return record, nil
}

func sqliteAPIKeyStorageValues(record APIKeyRecord) (string, string, int64, int64, any, any, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return "", "", 0, 0, nil, nil, err
	}
	scopes := append([]string{}, record.Scopes...)
	sort.Strings(scopes)
	scopesPayload, err := json.Marshal(scopes)
	if err != nil {
		return "", "", 0, 0, nil, nil, err
	}
	createdAt, err := apiKeyUnixNano(record.CreatedAt)
	if err != nil {
		return "", "", 0, 0, nil, nil, err
	}
	expiresAt, err := apiKeyUnixNano(record.ExpiresAt)
	if err != nil || expiresAt <= createdAt {
		return "", "", 0, 0, nil, nil, errors.New("invalid API key storage time range")
	}
	lastUsedAt, err := optionalAPIKeyUnixNano(record.LastUsedAt)
	if err != nil {
		return "", "", 0, 0, nil, nil, err
	}
	revokedAt, err := optionalAPIKeyUnixNano(record.RevokedAt)
	if err != nil {
		return "", "", 0, 0, nil, nil, err
	}
	return string(payload), string(scopesPayload), createdAt, expiresAt, lastUsedAt, revokedAt, nil
}

func apiKeyUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	parsed = parsed.UTC()
	unixNano := parsed.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(parsed) {
		return 0, errors.New("API key time is outside SQLite nanosecond range")
	}
	return unixNano, nil
}

func optionalAPIKeyUnixNano(value *string) (any, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := apiKeyUnixNano(*value)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}
