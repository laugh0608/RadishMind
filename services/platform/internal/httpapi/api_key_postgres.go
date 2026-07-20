package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresAPIKeyRepository struct {
	pool *pgxpool.Pool
}

type apiKeyRow interface {
	Scan(...any) error
}

func newPostgresAPIKeyRepository(pool *pgxpool.Pool) *postgresAPIKeyRepository {
	return &postgresAPIKeyRepository{pool: pool}
}

func (repository *postgresAPIKeyRepository) Create(requestContext APIKeyContext, record APIKeyRecord) (APIKeyRecord, error) {
	if repository == nil || repository.pool == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	created, err := scanPostgresAPIKeyRecord(repository.pool.QueryRow(apiKeyDatabaseContext(requestContext.RequestContext), `INSERT INTO api_key_records
        (tenant_ref, workspace_id, api_key_id, application_id, owner_subject_ref, schema_version, display_name,
         scopes, lifecycle_state, record_version, credential_digest, sanitized_record_payload, created_at,
         expires_at, last_used_at, revoked_at, created_by_actor_ref, revoked_by_actor_ref, request_id, audit_ref)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
        ON CONFLICT DO NOTHING
        RETURNING sanitized_record_payload, credential_digest, last_used_at`, record.TenantRef, record.WorkspaceID,
		record.APIKeyID, record.ApplicationID, record.OwnerSubjectRef, record.SchemaVersion, record.DisplayName,
		record.Scopes, record.LifecycleState, record.RecordVersion, record.credentialDigest[:], payload,
		record.CreatedAt, record.ExpiresAt, record.LastUsedAt, record.RevokedAt, record.CreatedByActorRef,
		record.RevokedByActorRef, record.RequestID, record.AuditRef))
	if err == nil {
		return created, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyIdentifierCollision
	}
	return APIKeyRecord{}, errAPIKeyStoreUnavailable
}

func (repository *postgresAPIKeyRepository) Read(requestContext APIKeyContext, apiKeyID string) (APIKeyRecord, error) {
	if repository == nil || repository.pool == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	record, err := scanPostgresAPIKeyRecord(repository.pool.QueryRow(apiKeyDatabaseContext(requestContext.RequestContext), `SELECT sanitized_record_payload, credential_digest, last_used_at
        FROM api_key_records WHERE tenant_ref=$1 AND workspace_id=$2 AND api_key_id=$3 AND owner_subject_ref=$4`,
		requestContext.TenantRef, requestContext.WorkspaceID, apiKeyID, requestContext.OwnerSubjectRef))
	if errors.Is(err, pgx.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	return record, nil
}

func (repository *postgresAPIKeyRepository) List(requestContext APIKeyContext, query apiKeyListQuery) ([]APIKeyRecord, error) {
	if repository == nil || repository.pool == nil {
		return nil, errAPIKeyStoreUnavailable
	}
	rows, err := repository.pool.Query(apiKeyDatabaseContext(requestContext.RequestContext), `SELECT sanitized_record_payload, credential_digest, last_used_at
        FROM api_key_records
        WHERE tenant_ref=$1 AND workspace_id=$2 AND owner_subject_ref=$3
          AND ($4='' OR application_id=$4)
          AND ($5='' OR (CASE WHEN lifecycle_state='revoked' THEN 'revoked' WHEN expires_at <= $6 THEN 'expired' ELSE 'active' END)=$5)
          AND ($7='' OR (created_at, api_key_id) < ($7::timestamptz, $8))
        ORDER BY created_at DESC, api_key_id DESC LIMIT $9`, requestContext.TenantRef, requestContext.WorkspaceID,
		requestContext.OwnerSubjectRef, query.ApplicationID, query.EffectiveState, query.Now.UTC(), query.AfterCreatedAt,
		query.AfterAPIKeyID, query.Limit)
	if err != nil {
		return nil, errAPIKeyStoreUnavailable
	}
	defer rows.Close()
	records := make([]APIKeyRecord, 0)
	for rows.Next() {
		record, scanErr := scanPostgresAPIKeyRecord(rows)
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

func (repository *postgresAPIKeyRepository) FindCredential(requestContext context.Context, apiKeyID string) (APIKeyRecord, error) {
	if repository == nil || repository.pool == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	record, err := scanPostgresAPIKeyRecord(repository.pool.QueryRow(apiKeyDatabaseContext(requestContext), `SELECT sanitized_record_payload, credential_digest, last_used_at
        FROM api_key_records WHERE api_key_id=$1`, apiKeyID))
	if errors.Is(err, pgx.ErrNoRows) {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	if err != nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	return record, nil
}

func (repository *postgresAPIKeyRepository) RecordSuccessfulAuthentication(requestContext context.Context, apiKeyID string, expectedVersion int, usedAt time.Time) (APIKeyRecord, error) {
	if repository == nil || repository.pool == nil {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	usedAt = usedAt.UTC()
	record, err := scanPostgresAPIKeyRecord(repository.pool.QueryRow(apiKeyDatabaseContext(requestContext), `UPDATE api_key_records
        SET last_used_at=CASE WHEN last_used_at IS NULL OR last_used_at < $3 THEN $3 ELSE last_used_at END
        WHERE api_key_id=$1 AND record_version=$2 AND lifecycle_state='active' AND expires_at > $3
        RETURNING sanitized_record_payload, credential_digest, last_used_at`, apiKeyID, expectedVersion, usedAt))
	if err == nil {
		return record, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
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

func (repository *postgresAPIKeyRepository) Revoke(requestContext APIKeyContext, apiKeyID string, expectedVersion int, update APIKeyRecord) (APIKeyRecord, error) {
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
	revoked, err := scanPostgresAPIKeyRecord(repository.pool.QueryRow(apiKeyDatabaseContext(requestContext.RequestContext), `UPDATE api_key_records SET
        lifecycle_state='revoked', record_version=$1, sanitized_record_payload=$2, revoked_at=$3,
        revoked_by_actor_ref=$4, request_id=$5, audit_ref=$6
        WHERE tenant_ref=$7 AND workspace_id=$8 AND api_key_id=$9 AND owner_subject_ref=$10
          AND record_version=$11 AND lifecycle_state='active'
        RETURNING sanitized_record_payload, credential_digest, last_used_at`, current.RecordVersion, payload,
		current.RevokedAt, current.RevokedByActorRef, current.RequestID, current.AuditRef, requestContext.TenantRef,
		requestContext.WorkspaceID, apiKeyID, requestContext.OwnerSubjectRef, expectedVersion))
	if err == nil {
		return revoked, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
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

func scanPostgresAPIKeyRecord(row apiKeyRow) (APIKeyRecord, error) {
	var payload []byte
	var digest []byte
	var lastUsedAt *time.Time
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
	if lastUsedAt == nil {
		record.LastUsedAt = nil
	} else {
		value := lastUsedAt.UTC().Format(time.RFC3339Nano)
		record.LastUsedAt = &value
	}
	return record, nil
}

func validStoredAPIKeyRecord(record APIKeyRecord) bool {
	if record.SchemaVersion != apiKeyRecordSchemaVersion || !apiKeyIDPattern.MatchString(record.APIKeyID) ||
		record.RecordVersion < 1 || strings.TrimSpace(record.TenantRef) == "" || strings.TrimSpace(record.WorkspaceID) == "" ||
		!applicationCatalogIDPattern.MatchString(record.ApplicationID) || strings.TrimSpace(record.OwnerSubjectRef) == "" ||
		(record.LifecycleState != apiKeyLifecycleActive && record.LifecycleState != apiKeyLifecycleRevoked) {
		return false
	}
	createdAt, createdErr := time.Parse(time.RFC3339Nano, record.CreatedAt)
	expiresAt, expiresErr := time.Parse(time.RFC3339Nano, record.ExpiresAt)
	if createdErr != nil || expiresErr != nil || !expiresAt.After(createdAt) {
		return false
	}
	scopes := append([]string{}, record.Scopes...)
	sort.Strings(scopes)
	if len(scopes) == 0 {
		return false
	}
	for index, scope := range scopes {
		if _, ok := apiKeyAllowedScopes[scope]; !ok || index > 0 && scope == scopes[index-1] {
			return false
		}
	}
	return true
}

func apiKeyDatabaseContext(requestContext context.Context) context.Context {
	if requestContext != nil {
		return requestContext
	}
	return context.Background()
}
