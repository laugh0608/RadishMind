package httpapi

import (
	"context"
	"crypto/rand"
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

const applicationCatalogSchemaVersion = "application_catalog_record.v1"

const (
	ApplicationCatalogFailureScopeDenied       = "application_catalog_scope_denied"
	ApplicationCatalogFailureNotFound          = "application_catalog_not_found"
	ApplicationCatalogFailurePayloadInvalid    = "application_catalog_payload_invalid"
	ApplicationCatalogFailureSecretForbidden   = "application_catalog_secret_material_forbidden"
	ApplicationCatalogFailureVersionConflict   = "application_catalog_version_conflict"
	ApplicationCatalogFailureArchived          = "application_catalog_archived"
	ApplicationCatalogFailureTransitionInvalid = "application_catalog_transition_invalid"
	ApplicationCatalogFailureCursorInvalid     = "application_catalog_cursor_invalid"
	ApplicationCatalogFailureStoreUnavailable  = "application_catalog_store_unavailable"
	ApplicationCatalogFailureWriteDisabled     = "application_catalog_write_disabled"
)

const (
	applicationCatalogLifecycleActive   = "active"
	applicationCatalogLifecycleArchived = "archived"
	applicationCatalogDefaultListLimit  = 50
	applicationCatalogMaximumListLimit  = 200
)

var (
	applicationCatalogIDPattern            = regexp.MustCompile(`^app_[a-z2-7]{16}$`)
	errApplicationCatalogNotFound          = errors.New(ApplicationCatalogFailureNotFound)
	errApplicationCatalogVersionConflict   = errors.New(ApplicationCatalogFailureVersionConflict)
	errApplicationCatalogArchived          = errors.New(ApplicationCatalogFailureArchived)
	errApplicationCatalogTransitionInvalid = errors.New(ApplicationCatalogFailureTransitionInvalid)
	errApplicationCatalogStoreUnavailable  = errors.New(ApplicationCatalogFailureStoreUnavailable)
)

type applicationCatalogVersionConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure applicationCatalogVersionConflictError) Error() string {
	return ApplicationCatalogFailureVersionConflict
}

func (failure applicationCatalogVersionConflictError) Is(target error) bool {
	return target == errApplicationCatalogVersionConflict
}

type ApplicationCatalogContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ActorRef        string
	OwnerSubjectRef string
	AuditRef        string
	WriteEnabled    bool
}

type ApplicationCatalogRecord struct {
	SchemaVersion     string  `json:"schema_version"`
	ApplicationID     string  `json:"application_id"`
	TenantRef         string  `json:"tenant_ref"`
	WorkspaceID       string  `json:"workspace_id"`
	OwnerSubjectRef   string  `json:"owner_subject_ref"`
	DisplayName       string  `json:"display_name"`
	Description       string  `json:"description"`
	ApplicationKind   string  `json:"application_kind"`
	LifecycleState    string  `json:"lifecycle_state"`
	RecordVersion     int     `json:"record_version"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ArchivedAt        *string `json:"archived_at"`
	CreatedByActorRef string  `json:"created_by_actor_ref"`
	UpdatedByActorRef string  `json:"updated_by_actor_ref"`
	RequestID         string  `json:"request_id"`
	AuditRef          string  `json:"audit_ref"`
}

type ApplicationCatalogCreateInput struct {
	DisplayName     string
	Description     string
	ApplicationKind string
}

type ApplicationCatalogUpdateInput struct {
	ExpectedVersion int
	DisplayName     string
	Description     string
	ApplicationKind string
}

type ApplicationCatalogListInput struct {
	LifecycleState  string
	ApplicationKind string
	Limit           int
	Cursor          string
}

type ApplicationCatalogResult struct {
	Record                *ApplicationCatalogRecord `json:"record"`
	FailureCode           string                    `json:"failure_code,omitempty"`
	CurrentRecordVersion  int                       `json:"current_record_version"`
	CurrentLifecycleState string                    `json:"current_lifecycle_state"`
}

type ApplicationCatalogListResult struct {
	Records     []ApplicationCatalogRecord
	NextCursor  *string
	FailureCode string
}

type applicationCatalogListQuery struct {
	LifecycleState     string
	ApplicationKind    string
	Limit              int
	AfterUpdatedAt     string
	AfterApplicationID string
}

type applicationCatalogCursor struct {
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	OwnerSubjectRef string `json:"owner_subject_ref"`
	LifecycleState  string `json:"lifecycle_state"`
	ApplicationKind string `json:"application_kind"`
	UpdatedAt       string `json:"updated_at"`
	ApplicationID   string `json:"application_id"`
}

type applicationCatalogRepository interface {
	Create(ApplicationCatalogContext, ApplicationCatalogRecord) (ApplicationCatalogRecord, error)
	Read(ApplicationCatalogContext, string) (ApplicationCatalogRecord, error)
	List(ApplicationCatalogContext, applicationCatalogListQuery) ([]ApplicationCatalogRecord, error)
	UpdateMetadata(ApplicationCatalogContext, string, int, ApplicationCatalogRecord) (ApplicationCatalogRecord, error)
	Archive(ApplicationCatalogContext, string, int, ApplicationCatalogRecord) (ApplicationCatalogRecord, error)
	RequireActive(ApplicationCatalogContext, string) (ApplicationCatalogRecord, error)
}

type memoryApplicationCatalogRepository struct {
	mu          sync.RWMutex
	records     map[string]ApplicationCatalogRecord
	unavailable bool
}

type applicationCatalogService struct {
	repository applicationCatalogRepository
	now        func() time.Time
	newID      func() (string, error)
}

func newMemoryApplicationCatalogRepository() *memoryApplicationCatalogRepository {
	return &memoryApplicationCatalogRepository{records: make(map[string]ApplicationCatalogRecord)}
}

func newApplicationCatalogService(repository applicationCatalogRepository) applicationCatalogService {
	return applicationCatalogService{
		repository: repository,
		now:        func() time.Time { return time.Now().UTC() },
		newID:      newApplicationCatalogID,
	}
}

func (service applicationCatalogService) Create(requestContext ApplicationCatalogContext, input ApplicationCatalogCreateInput) ApplicationCatalogResult {
	if !requestContext.WriteEnabled {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureWriteDisabled}
	}
	displayName, description, applicationKind, failure := normalizeApplicationCatalogMetadata(input.DisplayName, input.Description, input.ApplicationKind)
	if failure != "" {
		return ApplicationCatalogResult{FailureCode: failure}
	}
	for attempt := 0; attempt < 3; attempt++ {
		applicationID, err := service.newID()
		if err != nil || !applicationCatalogIDPattern.MatchString(applicationID) {
			return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureStoreUnavailable}
		}
		now := service.now().Format(time.RFC3339Nano)
		record := ApplicationCatalogRecord{
			SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: applicationID,
			TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
			OwnerSubjectRef: requestContext.OwnerSubjectRef, DisplayName: displayName,
			Description: description, ApplicationKind: applicationKind,
			LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1,
			CreatedAt: now, UpdatedAt: now, CreatedByActorRef: requestContext.ActorRef,
			UpdatedByActorRef: requestContext.ActorRef, RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
		}
		created, createErr := service.repository.Create(requestContext, record)
		if errors.Is(createErr, errApplicationCatalogVersionConflict) {
			continue
		}
		if createErr != nil {
			return applicationCatalogRepositoryFailure(createErr)
		}
		return ApplicationCatalogResult{Record: &created, CurrentRecordVersion: created.RecordVersion, CurrentLifecycleState: created.LifecycleState}
	}
	return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureStoreUnavailable}
}

func (service applicationCatalogService) Read(requestContext ApplicationCatalogContext, applicationID string) ApplicationCatalogResult {
	applicationID = strings.TrimSpace(applicationID)
	if !applicationCatalogIDPattern.MatchString(applicationID) {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	record, err := service.repository.Read(requestContext, applicationID)
	if err != nil {
		return applicationCatalogRepositoryFailure(err)
	}
	return ApplicationCatalogResult{Record: &record, CurrentRecordVersion: record.RecordVersion, CurrentLifecycleState: record.LifecycleState}
}

func (service applicationCatalogService) List(requestContext ApplicationCatalogContext, input ApplicationCatalogListInput) ApplicationCatalogListResult {
	lifecycle := strings.TrimSpace(input.LifecycleState)
	if lifecycle == "" {
		lifecycle = applicationCatalogLifecycleActive
	}
	if lifecycle != applicationCatalogLifecycleActive && lifecycle != applicationCatalogLifecycleArchived {
		return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	kind := strings.TrimSpace(input.ApplicationKind)
	if kind != "" && !isApplicationCatalogKind(kind) {
		return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	limit := input.Limit
	if limit == 0 {
		limit = applicationCatalogDefaultListLimit
	}
	if limit < 1 || limit > applicationCatalogMaximumListLimit {
		return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	query := applicationCatalogListQuery{LifecycleState: lifecycle, ApplicationKind: kind, Limit: limit + 1}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationCatalogCursor(input.Cursor)
		if err != nil || cursor.TenantRef != requestContext.TenantRef || cursor.WorkspaceID != requestContext.WorkspaceID ||
			cursor.OwnerSubjectRef != requestContext.OwnerSubjectRef || cursor.LifecycleState != lifecycle || cursor.ApplicationKind != kind ||
			!applicationCatalogIDPattern.MatchString(cursor.ApplicationID) {
			return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailureCursorInvalid}
		}
		if _, err := time.Parse(time.RFC3339Nano, cursor.UpdatedAt); err != nil {
			return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailureCursorInvalid}
		}
		query.AfterUpdatedAt = cursor.UpdatedAt
		query.AfterApplicationID = cursor.ApplicationID
	}
	records, err := service.repository.List(requestContext, query)
	if err != nil {
		return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailureStoreUnavailable}
	}
	result := ApplicationCatalogListResult{Records: records}
	if len(records) > limit {
		last := records[limit-1]
		result.Records = records[:limit]
		cursor, cursorErr := encodeApplicationCatalogCursor(applicationCatalogCursor{
			TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID, OwnerSubjectRef: requestContext.OwnerSubjectRef,
			LifecycleState: lifecycle, ApplicationKind: kind, UpdatedAt: last.UpdatedAt, ApplicationID: last.ApplicationID,
		})
		if cursorErr != nil {
			return ApplicationCatalogListResult{Records: []ApplicationCatalogRecord{}, FailureCode: ApplicationCatalogFailureStoreUnavailable}
		}
		result.NextCursor = &cursor
	}
	return result
}

func (service applicationCatalogService) Update(requestContext ApplicationCatalogContext, applicationID string, input ApplicationCatalogUpdateInput) ApplicationCatalogResult {
	if !requestContext.WriteEnabled {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureWriteDisabled}
	}
	applicationID = strings.TrimSpace(applicationID)
	if !applicationCatalogIDPattern.MatchString(applicationID) || input.ExpectedVersion < 1 {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	displayName, description, applicationKind, failure := normalizeApplicationCatalogMetadata(input.DisplayName, input.Description, input.ApplicationKind)
	if failure != "" {
		return ApplicationCatalogResult{FailureCode: failure}
	}
	updated, err := service.repository.UpdateMetadata(requestContext, applicationID, input.ExpectedVersion, ApplicationCatalogRecord{
		DisplayName: displayName, Description: description, ApplicationKind: applicationKind,
		UpdatedAt: service.now().Format(time.RFC3339Nano), UpdatedByActorRef: requestContext.ActorRef,
		RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
	})
	if err != nil {
		return applicationCatalogRepositoryFailure(err)
	}
	return ApplicationCatalogResult{Record: &updated, CurrentRecordVersion: updated.RecordVersion, CurrentLifecycleState: updated.LifecycleState}
}

func (service applicationCatalogService) Archive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int) ApplicationCatalogResult {
	if !requestContext.WriteEnabled {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureWriteDisabled}
	}
	applicationID = strings.TrimSpace(applicationID)
	if !applicationCatalogIDPattern.MatchString(applicationID) || expectedVersion < 1 {
		return ApplicationCatalogResult{FailureCode: ApplicationCatalogFailurePayloadInvalid}
	}
	now := service.now().Format(time.RFC3339Nano)
	archived, err := service.repository.Archive(requestContext, applicationID, expectedVersion, ApplicationCatalogRecord{
		LifecycleState: applicationCatalogLifecycleArchived, UpdatedAt: now, ArchivedAt: &now,
		UpdatedByActorRef: requestContext.ActorRef, RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
	})
	if err != nil {
		return applicationCatalogRepositoryFailure(err)
	}
	return ApplicationCatalogResult{Record: &archived, CurrentRecordVersion: archived.RecordVersion, CurrentLifecycleState: archived.LifecycleState}
}

func (service applicationCatalogService) RequireActive(requestContext ApplicationCatalogContext, applicationID string) ApplicationCatalogResult {
	record, err := service.repository.RequireActive(requestContext, strings.TrimSpace(applicationID))
	if err != nil {
		return applicationCatalogRepositoryFailure(err)
	}
	return ApplicationCatalogResult{Record: &record, CurrentRecordVersion: record.RecordVersion, CurrentLifecycleState: record.LifecycleState}
}

func newApplicationCatalogID() (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "app_" + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func normalizeApplicationCatalogMetadata(displayName, description, applicationKind string) (string, string, string, string) {
	displayName = strings.TrimSpace(displayName)
	description = strings.TrimSpace(description)
	applicationKind = strings.TrimSpace(applicationKind)
	if !utf8.ValidString(displayName) || !utf8.ValidString(description) || utf8.RuneCountInString(displayName) < 2 ||
		utf8.RuneCountInString(displayName) > 120 || utf8.RuneCountInString(description) > 1000 || !isApplicationCatalogKind(applicationKind) {
		return "", "", "", ApplicationCatalogFailurePayloadInvalid
	}
	if applicationDraftStringContainsSecret(displayName) || applicationDraftStringContainsSecret(description) {
		return "", "", "", ApplicationCatalogFailureSecretForbidden
	}
	return displayName, description, applicationKind, ""
}

func isApplicationCatalogKind(kind string) bool {
	switch kind {
	case "workflow_copilot", "docs_qa", "agent", "prompt_application":
		return true
	default:
		return false
	}
}

func applicationCatalogRepositoryKey(requestContext ApplicationCatalogContext, applicationID string) string {
	return strings.Join([]string{requestContext.TenantRef, requestContext.WorkspaceID, requestContext.OwnerSubjectRef, applicationID}, "\x00")
}

func (repository *memoryApplicationCatalogRepository) Create(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	key := applicationCatalogRepositoryKey(requestContext, record.ApplicationID)
	if current, exists := repository.records[key]; exists {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	repository.records[key] = record
	return record, nil
}

func (repository *memoryApplicationCatalogRepository) Read(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	record, exists := repository.records[applicationCatalogRepositoryKey(requestContext, applicationID)]
	if !exists {
		return ApplicationCatalogRecord{}, errApplicationCatalogNotFound
	}
	return record, nil
}

func (repository *memoryApplicationCatalogRepository) List(requestContext ApplicationCatalogContext, query applicationCatalogListQuery) ([]ApplicationCatalogRecord, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationCatalogStoreUnavailable
	}
	prefix := applicationCatalogRepositoryKey(requestContext, "")
	records := make([]ApplicationCatalogRecord, 0)
	for key, record := range repository.records {
		if !strings.HasPrefix(key, prefix) || record.LifecycleState != query.LifecycleState ||
			(query.ApplicationKind != "" && record.ApplicationKind != query.ApplicationKind) {
			continue
		}
		if query.AfterUpdatedAt != "" && (record.UpdatedAt > query.AfterUpdatedAt ||
			(record.UpdatedAt == query.AfterUpdatedAt && record.ApplicationID >= query.AfterApplicationID)) {
			continue
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].UpdatedAt == records[j].UpdatedAt {
			return records[i].ApplicationID > records[j].ApplicationID
		}
		return records[i].UpdatedAt > records[j].UpdatedAt
	})
	if len(records) > query.Limit {
		records = records[:query.Limit]
	}
	return records, nil
}

func (repository *memoryApplicationCatalogRepository) UpdateMetadata(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	key := applicationCatalogRepositoryKey(requestContext, applicationID)
	record, exists := repository.records[key]
	if !exists {
		return ApplicationCatalogRecord{}, errApplicationCatalogNotFound
	}
	if record.RecordVersion != expectedVersion {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: record.RecordVersion, CurrentState: record.LifecycleState}
	}
	if record.LifecycleState == applicationCatalogLifecycleArchived {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	record.DisplayName = update.DisplayName
	record.Description = update.Description
	record.ApplicationKind = update.ApplicationKind
	record.RecordVersion++
	record.UpdatedAt = update.UpdatedAt
	record.UpdatedByActorRef = update.UpdatedByActorRef
	record.RequestID = update.RequestID
	record.AuditRef = update.AuditRef
	repository.records[key] = record
	return record, nil
}

func (repository *memoryApplicationCatalogRepository) Archive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationCatalogRecord{}, errApplicationCatalogStoreUnavailable
	}
	key := applicationCatalogRepositoryKey(requestContext, applicationID)
	record, exists := repository.records[key]
	if !exists {
		return ApplicationCatalogRecord{}, errApplicationCatalogNotFound
	}
	if record.RecordVersion != expectedVersion {
		return ApplicationCatalogRecord{}, applicationCatalogVersionConflictError{CurrentVersion: record.RecordVersion, CurrentState: record.LifecycleState}
	}
	if record.LifecycleState != applicationCatalogLifecycleActive {
		return ApplicationCatalogRecord{}, errApplicationCatalogTransitionInvalid
	}
	record.LifecycleState = update.LifecycleState
	record.RecordVersion++
	record.UpdatedAt = update.UpdatedAt
	record.ArchivedAt = update.ArchivedAt
	record.UpdatedByActorRef = update.UpdatedByActorRef
	record.RequestID = update.RequestID
	record.AuditRef = update.AuditRef
	repository.records[key] = record
	return record, nil
}

func (repository *memoryApplicationCatalogRepository) RequireActive(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	record, err := repository.Read(requestContext, applicationID)
	if err != nil {
		return ApplicationCatalogRecord{}, err
	}
	if record.LifecycleState != applicationCatalogLifecycleActive {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	return record, nil
}

func applicationCatalogRepositoryFailure(err error) ApplicationCatalogResult {
	result := ApplicationCatalogResult{FailureCode: ApplicationCatalogFailureStoreUnavailable}
	switch {
	case errors.Is(err, errApplicationCatalogNotFound):
		result.FailureCode = ApplicationCatalogFailureNotFound
	case errors.Is(err, errApplicationCatalogVersionConflict):
		result.FailureCode = ApplicationCatalogFailureVersionConflict
	case errors.Is(err, errApplicationCatalogArchived):
		result.FailureCode = ApplicationCatalogFailureArchived
	case errors.Is(err, errApplicationCatalogTransitionInvalid):
		result.FailureCode = ApplicationCatalogFailureTransitionInvalid
	}
	var conflict applicationCatalogVersionConflictError
	if errors.As(err, &conflict) {
		result.CurrentRecordVersion = conflict.CurrentVersion
		result.CurrentLifecycleState = conflict.CurrentState
	}
	return result
}

func encodeApplicationCatalogCursor(cursor applicationCatalogCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeApplicationCatalogCursor(value string) (applicationCatalogCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return applicationCatalogCursor{}, err
	}
	var cursor applicationCatalogCursor
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil {
		return applicationCatalogCursor{}, err
	}
	return cursor, nil
}
