package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const apiKeyRecordSchemaVersion = "api_key_record.v1"

const (
	APIKeyFailureLifecycleDisabled      = "api_key_lifecycle_disabled"
	APIKeyFailureCatalogRequired        = "api_key_application_catalog_required"
	APIKeyFailureApplicationUnavailable = "api_key_application_unavailable"
	APIKeyFailureScopeDenied            = "api_key_scope_denied"
	APIKeyFailurePayloadInvalid         = "api_key_payload_invalid"
	APIKeyFailureSecretForbidden        = "api_key_secret_material_forbidden"
	APIKeyFailureNotFound               = "api_key_not_found"
	APIKeyFailureVersionConflict        = "api_key_version_conflict"
	APIKeyFailureRevoked                = "api_key_revoked"
	APIKeyFailureExpired                = "api_key_expired"
	APIKeyFailureTransitionInvalid      = "api_key_transition_invalid"
	APIKeyFailureCursorInvalid          = "api_key_cursor_invalid"
	APIKeyFailureStoreUnavailable       = "api_key_store_unavailable"
	APIKeyFailureWriteDisabled          = "api_key_write_disabled"
)

const (
	apiKeyLifecycleActive   = "active"
	apiKeyLifecycleRevoked  = "revoked"
	apiKeyEffectiveExpired  = "expired"
	apiKeyDefaultListLimit  = 50
	apiKeyMaximumListLimit  = 200
	apiKeyMinimumExpiryDays = 1
	apiKeyMaximumExpiryDays = 90
	apiKeyTokenPrefix       = "rmd_dev_"
)

var (
	apiKeyIDPattern              = regexp.MustCompile(`^key_[a-z2-7]{16}$`)
	apiKeySecretPattern          = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	errAPIKeyNotFound            = errors.New(APIKeyFailureNotFound)
	errAPIKeyVersionConflict     = errors.New(APIKeyFailureVersionConflict)
	errAPIKeyRevoked             = errors.New(APIKeyFailureRevoked)
	errAPIKeyTransitionInvalid   = errors.New(APIKeyFailureTransitionInvalid)
	errAPIKeyStoreUnavailable    = errors.New(APIKeyFailureStoreUnavailable)
	errAPIKeyIdentifierCollision = errors.New("api key identifier collision")
)

var apiKeyAllowedScopes = map[string]struct{}{
	"models:read":      {},
	"chat:invoke":      {},
	"responses:invoke": {},
	"messages:invoke":  {},
}

type apiKeyVersionConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure apiKeyVersionConflictError) Error() string {
	return APIKeyFailureVersionConflict
}

func (failure apiKeyVersionConflictError) Is(target error) bool {
	return target == errAPIKeyVersionConflict
}

type APIKeyContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ActorRef        string
	OwnerSubjectRef string
	AuditRef        string
	WriteEnabled    bool
}

type APIKeyRecord struct {
	SchemaVersion     string   `json:"schema_version"`
	APIKeyID          string   `json:"api_key_id"`
	TenantRef         string   `json:"tenant_ref"`
	WorkspaceID       string   `json:"workspace_id"`
	ApplicationID     string   `json:"application_id"`
	OwnerSubjectRef   string   `json:"owner_subject_ref"`
	DisplayName       string   `json:"display_name"`
	Scopes            []string `json:"scopes"`
	LifecycleState    string   `json:"lifecycle_state"`
	EffectiveState    string   `json:"effective_state"`
	RecordVersion     int      `json:"record_version"`
	CreatedAt         string   `json:"created_at"`
	ExpiresAt         string   `json:"expires_at"`
	LastUsedAt        *string  `json:"last_used_at"`
	RevokedAt         *string  `json:"revoked_at"`
	CreatedByActorRef string   `json:"created_by_actor_ref"`
	RevokedByActorRef *string  `json:"revoked_by_actor_ref"`
	RequestID         string   `json:"request_id"`
	AuditRef          string   `json:"audit_ref"`
	credentialDigest  [sha256.Size]byte
}

type APIKeyCreateInput struct {
	ApplicationID string
	DisplayName   string
	Scopes        []string
	ExpiresInDays int
}

type APIKeyListInput struct {
	ApplicationID  string
	EffectiveState string
	Limit          int
	Cursor         string
}

type APIKeyResult struct {
	Record                *APIKeyRecord
	CredentialToken       string
	FailureCode           string
	CurrentRecordVersion  int
	CurrentEffectiveState string
}

type APIKeyListResult struct {
	Records     []APIKeyRecord
	NextCursor  *string
	FailureCode string
}

type apiKeyListQuery struct {
	ApplicationID  string
	EffectiveState string
	Limit          int
	AfterCreatedAt string
	AfterAPIKeyID  string
	Now            time.Time
}

type apiKeyCursor struct {
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	OwnerSubjectRef string `json:"owner_subject_ref"`
	ApplicationID   string `json:"application_id"`
	EffectiveState  string `json:"effective_state"`
	CreatedAt       string `json:"created_at"`
	APIKeyID        string `json:"api_key_id"`
}

type apiKeyRepository interface {
	Create(APIKeyContext, APIKeyRecord) (APIKeyRecord, error)
	Read(APIKeyContext, string) (APIKeyRecord, error)
	List(APIKeyContext, apiKeyListQuery) ([]APIKeyRecord, error)
	Revoke(APIKeyContext, string, int, APIKeyRecord) (APIKeyRecord, error)
}

type memoryAPIKeyRepository struct {
	mu          sync.RWMutex
	records     map[string]APIKeyRecord
	unavailable bool
}

type apiKeyService struct {
	repository                   apiKeyRepository
	applicationCatalogRepository applicationCatalogRepository
	now                          func() time.Time
	newID                        func() (string, error)
	newCredential                func(string) (string, [sha256.Size]byte, error)
}

func newMemoryAPIKeyRepository() *memoryAPIKeyRepository {
	return &memoryAPIKeyRepository{records: make(map[string]APIKeyRecord)}
}

func newAPIKeyService(repository apiKeyRepository, applicationRepository applicationCatalogRepository) apiKeyService {
	return apiKeyService{
		repository:                   repository,
		applicationCatalogRepository: applicationRepository,
		now:                          func() time.Time { return time.Now().UTC() },
		newID:                        newAPIKeyID,
		newCredential:                newAPIKeyCredential,
	}
}

func (service apiKeyService) Create(requestContext APIKeyContext, input APIKeyCreateInput) APIKeyResult {
	if !requestContext.WriteEnabled {
		return APIKeyResult{FailureCode: APIKeyFailureWriteDisabled}
	}
	applicationID := strings.TrimSpace(input.ApplicationID)
	displayName, scopes, failure := normalizeAPIKeyCreateInput(input.DisplayName, input.Scopes, input.ExpiresInDays)
	if failure != "" || !applicationCatalogIDPattern.MatchString(applicationID) {
		if failure == "" {
			failure = APIKeyFailurePayloadInvalid
		}
		return APIKeyResult{FailureCode: failure}
	}
	if failure = service.requireActiveApplication(requestContext, applicationID); failure != "" {
		return APIKeyResult{FailureCode: failure}
	}

	for attempt := 0; attempt < 3; attempt++ {
		apiKeyID, err := service.newID()
		if err != nil || !apiKeyIDPattern.MatchString(apiKeyID) {
			return APIKeyResult{FailureCode: APIKeyFailureStoreUnavailable}
		}
		token, digest, err := service.newCredential(apiKeyID)
		if err != nil || !apiKeyCredentialMatches(token, digest) {
			return APIKeyResult{FailureCode: APIKeyFailureStoreUnavailable}
		}
		now := service.now().UTC()
		nowText := now.Format(time.RFC3339Nano)
		record := APIKeyRecord{
			SchemaVersion: apiKeyRecordSchemaVersion, APIKeyID: apiKeyID,
			TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
			ApplicationID: applicationID, OwnerSubjectRef: requestContext.OwnerSubjectRef,
			DisplayName: displayName, Scopes: scopes, LifecycleState: apiKeyLifecycleActive,
			EffectiveState: apiKeyLifecycleActive, RecordVersion: 1,
			CreatedAt: nowText, ExpiresAt: now.Add(time.Duration(input.ExpiresInDays) * 24 * time.Hour).Format(time.RFC3339Nano),
			CreatedByActorRef: requestContext.ActorRef, RequestID: requestContext.RequestID,
			AuditRef: requestContext.AuditRef, credentialDigest: digest,
		}
		created, createErr := service.repository.Create(requestContext, record)
		if errors.Is(createErr, errAPIKeyIdentifierCollision) {
			continue
		}
		if createErr != nil {
			return apiKeyRepositoryFailure(createErr)
		}
		created = projectAPIKeyRecord(created, service.now())
		return APIKeyResult{
			Record: &created, CredentialToken: token,
			CurrentRecordVersion: created.RecordVersion, CurrentEffectiveState: created.EffectiveState,
		}
	}
	return APIKeyResult{FailureCode: APIKeyFailureStoreUnavailable}
}

func (service apiKeyService) Read(requestContext APIKeyContext, apiKeyID string) APIKeyResult {
	apiKeyID = strings.TrimSpace(apiKeyID)
	if !apiKeyIDPattern.MatchString(apiKeyID) {
		return APIKeyResult{FailureCode: APIKeyFailurePayloadInvalid}
	}
	record, err := service.repository.Read(requestContext, apiKeyID)
	if err != nil {
		return apiKeyRepositoryFailure(err)
	}
	record = projectAPIKeyRecord(record, service.now())
	return APIKeyResult{Record: &record, CurrentRecordVersion: record.RecordVersion, CurrentEffectiveState: record.EffectiveState}
}

func (service apiKeyService) List(requestContext APIKeyContext, input APIKeyListInput) APIKeyListResult {
	applicationID := strings.TrimSpace(input.ApplicationID)
	if applicationID != "" && !applicationCatalogIDPattern.MatchString(applicationID) {
		return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailurePayloadInvalid}
	}
	effectiveState := strings.TrimSpace(input.EffectiveState)
	if effectiveState != "" && effectiveState != apiKeyLifecycleActive && effectiveState != apiKeyEffectiveExpired && effectiveState != apiKeyLifecycleRevoked {
		return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailurePayloadInvalid}
	}
	limit := input.Limit
	if limit == 0 {
		limit = apiKeyDefaultListLimit
	}
	if limit < 1 || limit > apiKeyMaximumListLimit {
		return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailurePayloadInvalid}
	}
	now := service.now().UTC()
	query := apiKeyListQuery{ApplicationID: applicationID, EffectiveState: effectiveState, Limit: limit + 1, Now: now}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeAPIKeyCursor(input.Cursor)
		if err != nil || cursor.TenantRef != requestContext.TenantRef || cursor.WorkspaceID != requestContext.WorkspaceID ||
			cursor.OwnerSubjectRef != requestContext.OwnerSubjectRef || cursor.ApplicationID != applicationID ||
			cursor.EffectiveState != effectiveState || !apiKeyIDPattern.MatchString(cursor.APIKeyID) {
			return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailureCursorInvalid}
		}
		if _, err := time.Parse(time.RFC3339Nano, cursor.CreatedAt); err != nil {
			return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailureCursorInvalid}
		}
		query.AfterCreatedAt = cursor.CreatedAt
		query.AfterAPIKeyID = cursor.APIKeyID
	}
	records, err := service.repository.List(requestContext, query)
	if err != nil {
		return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailureStoreUnavailable}
	}
	for index := range records {
		records[index] = projectAPIKeyRecord(records[index], now)
	}
	result := APIKeyListResult{Records: records}
	if len(records) > limit {
		last := records[limit-1]
		result.Records = records[:limit]
		cursor, cursorErr := encodeAPIKeyCursor(apiKeyCursor{
			TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
			OwnerSubjectRef: requestContext.OwnerSubjectRef, ApplicationID: applicationID,
			EffectiveState: effectiveState, CreatedAt: last.CreatedAt, APIKeyID: last.APIKeyID,
		})
		if cursorErr != nil {
			return APIKeyListResult{Records: []APIKeyRecord{}, FailureCode: APIKeyFailureStoreUnavailable}
		}
		result.NextCursor = &cursor
	}
	return result
}

func (service apiKeyService) Revoke(requestContext APIKeyContext, apiKeyID string, expectedVersion int) APIKeyResult {
	if !requestContext.WriteEnabled {
		return APIKeyResult{FailureCode: APIKeyFailureWriteDisabled}
	}
	apiKeyID = strings.TrimSpace(apiKeyID)
	if !apiKeyIDPattern.MatchString(apiKeyID) || expectedVersion < 1 {
		return APIKeyResult{FailureCode: APIKeyFailurePayloadInvalid}
	}
	now := service.now().UTC()
	nowText := now.Format(time.RFC3339Nano)
	revokedBy := requestContext.ActorRef
	revoked, err := service.repository.Revoke(requestContext, apiKeyID, expectedVersion, APIKeyRecord{
		LifecycleState: apiKeyLifecycleRevoked, EffectiveState: apiKeyLifecycleRevoked,
		RevokedAt: &nowText, RevokedByActorRef: &revokedBy,
		RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
	})
	if err != nil {
		return apiKeyRepositoryFailure(err)
	}
	revoked = projectAPIKeyRecord(revoked, now)
	return APIKeyResult{Record: &revoked, CurrentRecordVersion: revoked.RecordVersion, CurrentEffectiveState: revoked.EffectiveState}
}

func (service apiKeyService) requireActiveApplication(requestContext APIKeyContext, applicationID string) string {
	if service.applicationCatalogRepository == nil {
		return APIKeyFailureCatalogRequired
	}
	catalogContext := ApplicationCatalogContext{
		RequestContext: requestContext.RequestContext, RequestID: requestContext.RequestID,
		TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		ActorRef: requestContext.ActorRef, OwnerSubjectRef: requestContext.OwnerSubjectRef,
		AuditRef: requestContext.AuditRef,
	}
	result := newApplicationCatalogService(service.applicationCatalogRepository).RequireActive(catalogContext, applicationID)
	if result.FailureCode == "" && result.Record != nil {
		return ""
	}
	switch result.FailureCode {
	case ApplicationCatalogFailureStoreUnavailable:
		return APIKeyFailureStoreUnavailable
	case ApplicationCatalogFailureNotFound, ApplicationCatalogFailureArchived, ApplicationCatalogFailurePayloadInvalid:
		return APIKeyFailureApplicationUnavailable
	default:
		return APIKeyFailureApplicationUnavailable
	}
}

func normalizeAPIKeyCreateInput(displayName string, scopes []string, expiresInDays int) (string, []string, string) {
	displayName = strings.TrimSpace(displayName)
	if !utf8.ValidString(displayName) || utf8.RuneCountInString(displayName) < 2 || utf8.RuneCountInString(displayName) > 80 ||
		expiresInDays < apiKeyMinimumExpiryDays || expiresInDays > apiKeyMaximumExpiryDays {
		return "", nil, APIKeyFailurePayloadInvalid
	}
	if applicationDraftStringContainsSecret(displayName) {
		return "", nil, APIKeyFailureSecretForbidden
	}
	normalizedScopes := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, rawScope := range scopes {
		scope := strings.TrimSpace(rawScope)
		if _, allowed := apiKeyAllowedScopes[scope]; !allowed {
			return "", nil, APIKeyFailurePayloadInvalid
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		normalizedScopes = append(normalizedScopes, scope)
	}
	if len(normalizedScopes) == 0 {
		return "", nil, APIKeyFailurePayloadInvalid
	}
	sort.Strings(normalizedScopes)
	return displayName, normalizedScopes, ""
}

func newAPIKeyID() (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "key_" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func newAPIKeyCredential(apiKeyID string) (string, [sha256.Size]byte, error) {
	var empty [sha256.Size]byte
	if !apiKeyIDPattern.MatchString(apiKeyID) {
		return "", empty, errors.New("invalid API key identifier")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", empty, err
	}
	secret := base64.RawURLEncoding.EncodeToString(raw)
	if !apiKeySecretPattern.MatchString(secret) {
		return "", empty, errors.New("invalid generated API key secret")
	}
	token := apiKeyTokenPrefix + apiKeyID + "." + secret
	return token, sha256.Sum256([]byte(token)), nil
}

func parseAPIKeyCredential(token string) (string, bool) {
	token = strings.TrimSpace(token)
	if !strings.HasPrefix(token, apiKeyTokenPrefix) {
		return "", false
	}
	parts := strings.Split(strings.TrimPrefix(token, apiKeyTokenPrefix), ".")
	if len(parts) != 2 || !apiKeyIDPattern.MatchString(parts[0]) || !apiKeySecretPattern.MatchString(parts[1]) {
		return "", false
	}
	return parts[0], true
}

func apiKeyCredentialMatches(token string, expected [sha256.Size]byte) bool {
	if _, ok := parseAPIKeyCredential(token); !ok {
		return false
	}
	actual := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return subtle.ConstantTimeCompare(actual[:], expected[:]) == 1
}

func projectAPIKeyRecord(record APIKeyRecord, now time.Time) APIKeyRecord {
	record = cloneAPIKeyRecord(record)
	record.EffectiveState = effectiveAPIKeyState(record, now)
	return record
}

func effectiveAPIKeyState(record APIKeyRecord, now time.Time) string {
	if record.LifecycleState == apiKeyLifecycleRevoked {
		return apiKeyLifecycleRevoked
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, record.ExpiresAt)
	if err == nil && !now.UTC().Before(expiresAt.UTC()) {
		return apiKeyEffectiveExpired
	}
	return apiKeyLifecycleActive
}

func apiKeyRepositoryFailure(err error) APIKeyResult {
	var conflict apiKeyVersionConflictError
	switch {
	case errors.As(err, &conflict):
		return APIKeyResult{
			FailureCode: APIKeyFailureVersionConflict, CurrentRecordVersion: conflict.CurrentVersion,
			CurrentEffectiveState: conflict.CurrentState,
		}
	case errors.Is(err, errAPIKeyNotFound):
		return APIKeyResult{FailureCode: APIKeyFailureNotFound}
	case errors.Is(err, errAPIKeyRevoked):
		return APIKeyResult{FailureCode: APIKeyFailureRevoked, CurrentEffectiveState: apiKeyLifecycleRevoked}
	case errors.Is(err, errAPIKeyTransitionInvalid):
		return APIKeyResult{FailureCode: APIKeyFailureTransitionInvalid}
	default:
		return APIKeyResult{FailureCode: APIKeyFailureStoreUnavailable}
	}
}

func apiKeyRepositoryKey(requestContext APIKeyContext, apiKeyID string) string {
	return strings.Join([]string{requestContext.TenantRef, requestContext.WorkspaceID, requestContext.OwnerSubjectRef, apiKeyID}, "\x00")
}

func cloneAPIKeyRecord(record APIKeyRecord) APIKeyRecord {
	record.Scopes = append([]string{}, record.Scopes...)
	if record.LastUsedAt != nil {
		value := *record.LastUsedAt
		record.LastUsedAt = &value
	}
	if record.RevokedAt != nil {
		value := *record.RevokedAt
		record.RevokedAt = &value
	}
	if record.RevokedByActorRef != nil {
		value := *record.RevokedByActorRef
		record.RevokedByActorRef = &value
	}
	return record
}

func (repository *memoryAPIKeyRepository) Create(requestContext APIKeyContext, record APIKeyRecord) (APIKeyRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	key := apiKeyRepositoryKey(requestContext, record.APIKeyID)
	if _, exists := repository.records[key]; exists {
		return APIKeyRecord{}, errAPIKeyIdentifierCollision
	}
	record = cloneAPIKeyRecord(record)
	repository.records[key] = record
	return cloneAPIKeyRecord(record), nil
}

func (repository *memoryAPIKeyRepository) Read(requestContext APIKeyContext, apiKeyID string) (APIKeyRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	record, exists := repository.records[apiKeyRepositoryKey(requestContext, apiKeyID)]
	if !exists {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	return cloneAPIKeyRecord(record), nil
}

func (repository *memoryAPIKeyRepository) List(requestContext APIKeyContext, query apiKeyListQuery) ([]APIKeyRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errAPIKeyStoreUnavailable
	}
	prefix := apiKeyRepositoryKey(requestContext, "")
	records := make([]APIKeyRecord, 0)
	for key, record := range repository.records {
		if !strings.HasPrefix(key, prefix) || (query.ApplicationID != "" && record.ApplicationID != query.ApplicationID) ||
			(query.EffectiveState != "" && effectiveAPIKeyState(record, query.Now) != query.EffectiveState) {
			continue
		}
		if query.AfterCreatedAt != "" && (record.CreatedAt > query.AfterCreatedAt ||
			(record.CreatedAt == query.AfterCreatedAt && record.APIKeyID >= query.AfterAPIKeyID)) {
			continue
		}
		records = append(records, cloneAPIKeyRecord(record))
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].CreatedAt == records[j].CreatedAt {
			return records[i].APIKeyID > records[j].APIKeyID
		}
		return records[i].CreatedAt > records[j].CreatedAt
	})
	if len(records) > query.Limit {
		records = records[:query.Limit]
	}
	return records, nil
}

func (repository *memoryAPIKeyRepository) Revoke(requestContext APIKeyContext, apiKeyID string, expectedVersion int, update APIKeyRecord) (APIKeyRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return APIKeyRecord{}, errAPIKeyStoreUnavailable
	}
	key := apiKeyRepositoryKey(requestContext, apiKeyID)
	record, exists := repository.records[key]
	if !exists {
		return APIKeyRecord{}, errAPIKeyNotFound
	}
	if record.RecordVersion != expectedVersion {
		return APIKeyRecord{}, apiKeyVersionConflictError{CurrentVersion: record.RecordVersion, CurrentState: record.LifecycleState}
	}
	if record.LifecycleState == apiKeyLifecycleRevoked {
		return APIKeyRecord{}, errAPIKeyRevoked
	}
	if update.LifecycleState != apiKeyLifecycleRevoked || update.RevokedAt == nil || update.RevokedByActorRef == nil {
		return APIKeyRecord{}, errAPIKeyTransitionInvalid
	}
	record.LifecycleState = apiKeyLifecycleRevoked
	record.EffectiveState = apiKeyLifecycleRevoked
	record.RecordVersion++
	record.RevokedAt = update.RevokedAt
	record.RevokedByActorRef = update.RevokedByActorRef
	record.RequestID = update.RequestID
	record.AuditRef = update.AuditRef
	repository.records[key] = cloneAPIKeyRecord(record)
	return cloneAPIKeyRecord(record), nil
}

func encodeAPIKeyCursor(cursor apiKeyCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeAPIKeyCursor(value string) (apiKeyCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return apiKeyCursor{}, err
	}
	var cursor apiKeyCursor
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil {
		return apiKeyCursor{}, err
	}
	return cursor, nil
}
