package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	workflowRAGSnapshotSchemaVersion = "workflow_rag_snapshot.v1"
	workflowRAGFragmentSchemaVersion = "workflow_rag_fragment.v1"
	workflowRAGAuditSchemaVersion    = "workflow_rag_execution_audit.v1"
	workflowRAGProfileID             = "workflow.rag.lexical-ngram-dev.v1"
	workflowRAGProfileVersion        = 1

	workflowRAGSnapshotActive   = "active"
	workflowRAGSnapshotArchived = "archived"

	workflowRAGMaxSnapshots        = 1024
	workflowRAGMaxFragments        = 256
	workflowRAGMaxFragmentBytes    = 8 * 1024
	workflowRAGMaxSnapshotBytes    = 1 * 1024 * 1024
	workflowRAGMaxQueryBytes       = 4 * 1024
	workflowRAGDefaultTopK         = 5
	workflowRAGMaxTopK             = 8
	workflowRAGMaxExcerptBytes     = 4 * 1024
	workflowRAGMaxContextBytes     = 32 * 1024
	workflowRAGDefaultListLimit    = 50
	workflowRAGMaximumListLimit    = 200
	workflowRAGOfficialScoreBoost  = 1.10
	workflowRAGLexicalK1           = 1.20
	workflowRAGLexicalB            = 0.75
	workflowRAGNormalizationPolicy = "unicode-lower-latin-word-cjk-bigram.v1"
	workflowRAGStopwordPolicy      = "none.v1"
)

const (
	WorkflowRAGFailureScopeDenied      = "workflow_rag_snapshot_scope_denied"
	WorkflowRAGFailureNotFound         = "workflow_rag_snapshot_not_found"
	WorkflowRAGFailurePayloadInvalid   = "workflow_rag_snapshot_payload_invalid"
	WorkflowRAGFailureFragmentInvalid  = "workflow_rag_fragment_invalid"
	WorkflowRAGFailureSecretForbidden  = "workflow_rag_secret_material_forbidden"
	WorkflowRAGFailureVersionConflict  = "workflow_rag_snapshot_version_conflict"
	WorkflowRAGFailureArchived         = "workflow_rag_snapshot_archived"
	WorkflowRAGFailureBudgetExceeded   = "workflow_rag_budget_exceeded"
	WorkflowRAGFailureQueryInvalid     = "workflow_rag_query_invalid"
	WorkflowRAGFailureNoEvidence       = "workflow_rag_no_evidence"
	WorkflowRAGFailureProfileDisabled  = "workflow_rag_profile_disabled"
	WorkflowRAGFailureDraftIneligible  = "workflow_rag_draft_ineligible"
	WorkflowRAGFailureRetrievalFailed  = "workflow_rag_retrieval_failed"
	WorkflowRAGFailureGatewayFailed    = "workflow_rag_gateway_failed"
	WorkflowRAGFailureCanceled         = "workflow_rag_canceled"
	WorkflowRAGFailureAnswerInvalid    = "workflow_rag_answer_invalid"
	WorkflowRAGFailureCitationInvalid  = "workflow_rag_citation_invalid"
	WorkflowRAGFailureInterrupted      = "workflow_rag_execution_interrupted"
	WorkflowRAGFailureStoreUnavailable = "workflow_rag_store_unavailable"
)

var (
	workflowRAGSnapshotIDPattern   = regexp.MustCompile(`^rags_[a-z2-7]{16}$`)
	workflowRAGSnapshotKeyPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]{2,47}$`)
	workflowRAGFragmentRefPattern  = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)
	workflowRAGScopedIDPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{2,119}$`)
	workflowRAGReferencePattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]{2,159}$`)
	workflowRAGPageSlugPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]{0,119}$`)
	workflowRAGDigestPattern       = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
	workflowRAGRAGRefPattern       = regexp.MustCompile(`^workflow\.rag\.[a-z][a-z0-9_]{2,47}\.v[1-9][0-9]*$`)
	workflowRAGRunIDPattern        = regexp.MustCompile(`^run_[a-z0-9]{16,64}$`)
	workflowRAGRunFailurePattern   = regexp.MustCompile(`^workflow_rag_[a-z_]{3,80}$`)
	workflowRAGDatasetIDPattern    = regexp.MustCompile(`^wragd_[a-z0-9_]{3,47}$`)
	workflowRAGSampleIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]{2,47}$`)
	errWorkflowRAGNotFound         = errors.New(WorkflowRAGFailureNotFound)
	errWorkflowRAGScopeDenied      = errors.New(WorkflowRAGFailureScopeDenied)
	errWorkflowRAGVersionConflict  = errors.New(WorkflowRAGFailureVersionConflict)
	errWorkflowRAGArchived         = errors.New(WorkflowRAGFailureArchived)
	errWorkflowRAGStoreUnavailable = errors.New(WorkflowRAGFailureStoreUnavailable)
	errWorkflowRAGStoreContract    = errors.New("workflow RAG snapshot store contract mismatch")
)

type workflowRAGVersionConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure workflowRAGVersionConflictError) Error() string {
	return WorkflowRAGFailureVersionConflict
}
func (failure workflowRAGVersionConflictError) Is(target error) bool {
	return target == errWorkflowRAGVersionConflict
}

type WorkflowRAGSnapshotContext struct {
	RequestContext context.Context
	RequestID      string
	TenantRef      string
	WorkspaceID    string
	ApplicationID  string
	ActorRef       string
	AuditRef       string
}

type WorkflowRAGFragmentInput struct {
	FragmentRef string `json:"fragment_ref"`
	SourceType  string `json:"source_type"`
	SourceRef   string `json:"source_ref"`
	PageSlug    string `json:"page_slug"`
	Title       string `json:"title"`
	IsOfficial  bool   `json:"is_official"`
	Content     string `json:"content"`
}

type WorkflowRAGFragment struct {
	SchemaVersion         string `json:"schema_version"`
	FragmentRef           string `json:"fragment_ref"`
	SourceType            string `json:"source_type"`
	SourceRef             string `json:"source_ref"`
	PageSlug              string `json:"page_slug"`
	Title                 string `json:"title"`
	IsOfficial            bool   `json:"is_official"`
	Content               string `json:"content"`
	ContentClassification string `json:"content_classification"`
	ContentBytes          int    `json:"content_bytes"`
	ContentDigest         string `json:"content_digest"`
}

type WorkflowRAGSnapshotRecord struct {
	SchemaVersion         string                `json:"schema_version"`
	SnapshotID            string                `json:"snapshot_id"`
	TenantRef             string                `json:"tenant_ref"`
	WorkspaceID           string                `json:"workspace_id"`
	ApplicationID         string                `json:"application_id"`
	SnapshotKey           string                `json:"snapshot_key"`
	RAGRef                string                `json:"rag_ref"`
	SnapshotVersion       int                   `json:"snapshot_version"`
	DisplayName           string                `json:"display_name"`
	LifecycleState        string                `json:"lifecycle_state"`
	ContentClassification string                `json:"content_classification"`
	ProfileRef            string                `json:"profile_ref"`
	FragmentCount         int                   `json:"fragment_count"`
	TotalContentBytes     int                   `json:"total_content_bytes"`
	SnapshotDigest        string                `json:"snapshot_digest"`
	CreatedAt             string                `json:"created_at"`
	CreatedByActorRef     string                `json:"created_by_actor_ref"`
	RequestID             string                `json:"request_id"`
	AuditRef              string                `json:"audit_ref"`
	Fragments             []WorkflowRAGFragment `json:"fragments"`
}

type WorkflowRAGSnapshotResource struct {
	SnapshotID        string  `json:"snapshot_id"`
	TenantRef         string  `json:"tenant_ref"`
	WorkspaceID       string  `json:"workspace_id"`
	ApplicationID     string  `json:"application_id"`
	SnapshotKey       string  `json:"snapshot_key"`
	DisplayName       string  `json:"display_name"`
	LifecycleState    string  `json:"lifecycle_state"`
	LatestVersion     int     `json:"latest_version"`
	LatestRAGRef      string  `json:"latest_rag_ref"`
	LatestDigest      string  `json:"latest_digest"`
	FragmentCount     int     `json:"fragment_count"`
	TotalContentBytes int     `json:"total_content_bytes"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	ArchivedAt        *string `json:"archived_at"`
}

type WorkflowRAGExecutionAudit struct {
	SchemaVersion     string `json:"schema_version"`
	EventID           string `json:"event_id"`
	EventKind         string `json:"event_kind"`
	TenantRef         string `json:"tenant_ref"`
	WorkspaceID       string `json:"workspace_id"`
	ApplicationID     string `json:"application_id"`
	SnapshotID        string `json:"snapshot_id"`
	SnapshotKey       string `json:"snapshot_key"`
	SnapshotVersion   int    `json:"snapshot_version"`
	SnapshotDigest    string `json:"snapshot_digest"`
	FragmentCount     int    `json:"fragment_count"`
	TotalContentBytes int    `json:"total_content_bytes"`
	ActorRef          string `json:"actor_ref"`
	RequestID         string `json:"request_id"`
	AuditRef          string `json:"audit_ref"`
	OccurredAt        string `json:"occurred_at"`
}

type WorkflowRAGExecutionProfile struct {
	SchemaVersion       string  `json:"schema_version"`
	ProfileID           string  `json:"profile_id"`
	ProfileVersion      int     `json:"profile_version"`
	Algorithm           string  `json:"algorithm"`
	NormalizationPolicy string  `json:"normalization_policy"`
	StopwordPolicy      string  `json:"stopword_policy"`
	DefaultTopK         int     `json:"default_top_k"`
	MaxTopK             int     `json:"max_top_k"`
	MaxQueryBytes       int     `json:"max_query_bytes"`
	MaxFragmentBytes    int     `json:"max_fragment_bytes"`
	MaxExcerptBytes     int     `json:"max_excerpt_bytes"`
	MaxContextBytes     int     `json:"max_context_bytes"`
	OfficialScoreBoost  float64 `json:"official_score_boost"`
	NetworkAccess       bool    `json:"network_access"`
	EmbeddingProvider   string  `json:"embedding_provider"`
	ProfileDigest       string  `json:"profile_digest"`
}

type WorkflowRAGSnapshotCreateInput struct {
	SnapshotKey           string
	DisplayName           string
	ContentClassification string
	Fragments             []WorkflowRAGFragmentInput
}

type WorkflowRAGSnapshotVersionInput struct {
	ExpectedLatestVersion int
	DisplayName           string
	ContentClassification string
	Fragments             []WorkflowRAGFragmentInput
}

type WorkflowRAGSnapshotListInput struct {
	LifecycleState string
	Limit          int
	Cursor         string
}

type WorkflowRAGSnapshotResult struct {
	Record               *WorkflowRAGSnapshotRecord
	FailureCode          string
	CurrentLatestVersion int
	CurrentLifecycle     string
}

type WorkflowRAGSnapshotListResult struct {
	Records     []WorkflowRAGSnapshotResource
	NextCursor  *string
	FailureCode string
}

type workflowRAGSnapshotListQuery struct {
	LifecycleState   string
	Limit            int
	AfterSnapshotKey string
}

type workflowRAGSnapshotRepository interface {
	Create(WorkflowRAGSnapshotContext, WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, WorkflowRAGExecutionAudit) error
	List(WorkflowRAGSnapshotContext, workflowRAGSnapshotListQuery) ([]WorkflowRAGSnapshotResource, error)
	ReadLatest(WorkflowRAGSnapshotContext, string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error)
	ReadVersion(WorkflowRAGSnapshotContext, string, int) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error)
	ReadByRAGRef(WorkflowRAGSnapshotContext, string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error)
	CreateVersion(WorkflowRAGSnapshotContext, string, int, WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, WorkflowRAGExecutionAudit) error
	Archive(WorkflowRAGSnapshotContext, string, int, WorkflowRAGSnapshotResource, WorkflowRAGExecutionAudit) error
	AppendAudit(WorkflowRAGSnapshotContext, WorkflowRAGExecutionAudit) error
}

type workflowRAGMemoryEntry struct {
	resource WorkflowRAGSnapshotResource
	versions map[int]WorkflowRAGSnapshotRecord
	audits   []WorkflowRAGExecutionAudit
}

type memoryWorkflowRAGSnapshotRepository struct {
	ownerLock *sync.RWMutex
	entries   map[string]workflowRAGMemoryEntry
	capacity  int
	available bool
}

type workflowRAGSnapshotService struct {
	repository workflowRAGSnapshotRepository
	now        func() time.Time
	newID      func(string) (string, error)
}

func newMemoryWorkflowRAGSnapshotRepository(ownerLock *sync.RWMutex) *memoryWorkflowRAGSnapshotRepository {
	if ownerLock == nil {
		ownerLock = &sync.RWMutex{}
	}
	return &memoryWorkflowRAGSnapshotRepository{
		ownerLock: ownerLock,
		entries:   make(map[string]workflowRAGMemoryEntry),
		capacity:  workflowRAGMaxSnapshots,
		available: true,
	}
}

func newWorkflowRAGSnapshotService(repository workflowRAGSnapshotRepository) workflowRAGSnapshotService {
	return workflowRAGSnapshotService{
		repository: repository,
		now:        func() time.Time { return time.Now().UTC() },
		newID:      newWorkflowRAGStableID,
	}
}

func (service workflowRAGSnapshotService) Create(ctx WorkflowRAGSnapshotContext, input WorkflowRAGSnapshotCreateInput) WorkflowRAGSnapshotResult {
	if failure := validateWorkflowRAGContext(ctx); failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	key, displayName, classification, fragments, totalBytes, failure := normalizeWorkflowRAGSnapshotInput(
		input.SnapshotKey, input.DisplayName, input.ContentClassification, input.Fragments,
	)
	if failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	snapshotID, err := service.newID("rags_")
	if err != nil || !workflowRAGSnapshotIDPattern.MatchString(snapshotID) {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	now := service.now().Format(time.RFC3339Nano)
	record, err := buildWorkflowRAGSnapshotRecord(ctx, snapshotID, key, displayName, classification, 1, now, fragments, totalBytes)
	if err != nil {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	resource := workflowRAGResourceFromRecord(record, now)
	audit, err := buildWorkflowRAGSnapshotAudit(ctx, "snapshot_created", record, now)
	if err != nil {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	if err = service.repository.Create(ctx, resource, record, audit); err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	return WorkflowRAGSnapshotResult{Record: &record, CurrentLatestVersion: 1, CurrentLifecycle: resource.LifecycleState}
}

func (service workflowRAGSnapshotService) List(ctx WorkflowRAGSnapshotContext, input WorkflowRAGSnapshotListInput) WorkflowRAGSnapshotListResult {
	if failure := validateWorkflowRAGContext(ctx); failure != "" {
		return WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: failure}
	}
	lifecycle := strings.TrimSpace(input.LifecycleState)
	if lifecycle == "" {
		lifecycle = workflowRAGSnapshotActive
	}
	if lifecycle != workflowRAGSnapshotActive && lifecycle != workflowRAGSnapshotArchived {
		return WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	limit := input.Limit
	if limit == 0 {
		limit = workflowRAGDefaultListLimit
	}
	if limit < 1 || limit > workflowRAGMaximumListLimit {
		return WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	afterKey := strings.TrimSpace(input.Cursor)
	if afterKey != "" && !workflowRAGSnapshotKeyPattern.MatchString(afterKey) {
		return WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	records, err := service.repository.List(ctx, workflowRAGSnapshotListQuery{LifecycleState: lifecycle, Limit: limit + 1, AfterSnapshotKey: afterKey})
	if err != nil {
		return WorkflowRAGSnapshotListResult{Records: []WorkflowRAGSnapshotResource{}, FailureCode: workflowRAGRepositoryFailure(err).FailureCode}
	}
	result := WorkflowRAGSnapshotListResult{Records: records}
	if len(records) > limit {
		result.Records = records[:limit]
		cursor := result.Records[len(result.Records)-1].SnapshotKey
		result.NextCursor = &cursor
	}
	return result
}

func (service workflowRAGSnapshotService) Read(ctx WorkflowRAGSnapshotContext, snapshotID string, snapshotVersion int) WorkflowRAGSnapshotResult {
	if failure := validateWorkflowRAGContext(ctx); failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	snapshotID = strings.TrimSpace(snapshotID)
	if !workflowRAGSnapshotIDPattern.MatchString(snapshotID) || snapshotVersion < 1 {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	resource, record, err := service.repository.ReadVersion(ctx, snapshotID, snapshotVersion)
	if err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	record.LifecycleState = resource.LifecycleState
	return WorkflowRAGSnapshotResult{Record: &record, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
}

func (service workflowRAGSnapshotService) Version(ctx WorkflowRAGSnapshotContext, snapshotID string, input WorkflowRAGSnapshotVersionInput) WorkflowRAGSnapshotResult {
	if failure := validateWorkflowRAGContext(ctx); failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	snapshotID = strings.TrimSpace(snapshotID)
	if !workflowRAGSnapshotIDPattern.MatchString(snapshotID) || input.ExpectedLatestVersion < 1 {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	resource, _, err := service.repository.ReadLatest(ctx, snapshotID)
	if err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	if resource.LatestVersion != input.ExpectedLatestVersion {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureVersionConflict, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGSnapshotActive {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureArchived, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	_, displayName, classification, fragments, totalBytes, failure := normalizeWorkflowRAGSnapshotInput(
		resource.SnapshotKey, input.DisplayName, input.ContentClassification, input.Fragments,
	)
	if failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	now := service.now().Format(time.RFC3339Nano)
	record, err := buildWorkflowRAGSnapshotRecord(ctx, snapshotID, resource.SnapshotKey, displayName, classification, resource.LatestVersion+1, now, fragments, totalBytes)
	if err != nil {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	updated := resource
	updated.DisplayName = record.DisplayName
	updated.LatestVersion = record.SnapshotVersion
	updated.LatestRAGRef = record.RAGRef
	updated.LatestDigest = record.SnapshotDigest
	updated.FragmentCount = record.FragmentCount
	updated.TotalContentBytes = record.TotalContentBytes
	updated.UpdatedAt = now
	audit, err := buildWorkflowRAGSnapshotAudit(ctx, "snapshot_versioned", record, now)
	if err != nil {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	if err = service.repository.CreateVersion(ctx, snapshotID, input.ExpectedLatestVersion, updated, record, audit); err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	return WorkflowRAGSnapshotResult{Record: &record, CurrentLatestVersion: updated.LatestVersion, CurrentLifecycle: updated.LifecycleState}
}

func (service workflowRAGSnapshotService) Archive(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedLatestVersion int) WorkflowRAGSnapshotResult {
	if failure := validateWorkflowRAGContext(ctx); failure != "" {
		return WorkflowRAGSnapshotResult{FailureCode: failure}
	}
	snapshotID = strings.TrimSpace(snapshotID)
	if !workflowRAGSnapshotIDPattern.MatchString(snapshotID) || expectedLatestVersion < 1 {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailurePayloadInvalid}
	}
	resource, record, err := service.repository.ReadLatest(ctx, snapshotID)
	if err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	if resource.LatestVersion != expectedLatestVersion {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureVersionConflict, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGSnapshotActive {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureArchived, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	now := service.now().Format(time.RFC3339Nano)
	resource.LifecycleState = workflowRAGSnapshotArchived
	resource.UpdatedAt = now
	resource.ArchivedAt = &now
	record.LifecycleState = workflowRAGSnapshotArchived
	audit, err := buildWorkflowRAGSnapshotAudit(ctx, "snapshot_archived", record, now)
	if err != nil {
		return WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	}
	if err = service.repository.Archive(ctx, snapshotID, expectedLatestVersion, resource, audit); err != nil {
		return workflowRAGRepositoryFailure(err)
	}
	return WorkflowRAGSnapshotResult{Record: &record, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
}

func normalizeWorkflowRAGSnapshotInput(snapshotKey, displayName, classification string, inputs []WorkflowRAGFragmentInput) (string, string, string, []WorkflowRAGFragment, int, string) {
	snapshotKey = strings.TrimSpace(snapshotKey)
	displayName = strings.TrimSpace(displayName)
	classification = strings.TrimSpace(classification)
	if !workflowRAGSnapshotKeyPattern.MatchString(snapshotKey) || !utf8.ValidString(displayName) || utf8.RuneCountInString(displayName) < 2 || utf8.RuneCountInString(displayName) > 120 || !workflowRAGClassificationAllowed(classification) {
		return "", "", "", nil, 0, WorkflowRAGFailurePayloadInvalid
	}
	if workflowRAGContainsForbiddenMaterial(displayName) {
		return "", "", "", nil, 0, WorkflowRAGFailureSecretForbidden
	}
	if len(inputs) < 1 || len(inputs) > workflowRAGMaxFragments {
		return "", "", "", nil, 0, WorkflowRAGFailureBudgetExceeded
	}
	fragments := make([]WorkflowRAGFragment, 0, len(inputs))
	seenRefs := make(map[string]bool, len(inputs))
	totalBytes := 0
	for _, input := range inputs {
		fragment, failure := normalizeWorkflowRAGFragment(input, classification)
		if failure != "" {
			return "", "", "", nil, 0, failure
		}
		if seenRefs[fragment.FragmentRef] {
			return "", "", "", nil, 0, WorkflowRAGFailureFragmentInvalid
		}
		seenRefs[fragment.FragmentRef] = true
		totalBytes += fragment.ContentBytes
		if totalBytes > workflowRAGMaxSnapshotBytes {
			return "", "", "", nil, 0, WorkflowRAGFailureBudgetExceeded
		}
		fragments = append(fragments, fragment)
	}
	sort.Slice(fragments, func(i, j int) bool { return fragments[i].FragmentRef < fragments[j].FragmentRef })
	return snapshotKey, displayName, classification, fragments, totalBytes, ""
}

func normalizeWorkflowRAGFragment(input WorkflowRAGFragmentInput, classification string) (WorkflowRAGFragment, string) {
	input.FragmentRef = strings.TrimSpace(input.FragmentRef)
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceRef = strings.TrimSpace(input.SourceRef)
	input.PageSlug = strings.TrimSpace(input.PageSlug)
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	contentBytes := len([]byte(input.Content))
	if !workflowRAGFragmentRefPattern.MatchString(input.FragmentRef) || !workflowRAGSourceTypeAllowed(input.SourceType) ||
		!workflowRAGReferencePattern.MatchString(input.SourceRef) || strings.Contains(input.SourceRef, "://") ||
		!workflowRAGPageSlugPattern.MatchString(input.PageSlug) || !utf8.ValidString(input.Title) || utf8.RuneCountInString(input.Title) > 160 ||
		!utf8.ValidString(input.Content) || input.Content == "" {
		return WorkflowRAGFragment{}, WorkflowRAGFailureFragmentInvalid
	}
	if contentBytes > workflowRAGMaxFragmentBytes {
		return WorkflowRAGFragment{}, WorkflowRAGFailureBudgetExceeded
	}
	if workflowRAGContainsForbiddenMaterial(strings.Join([]string{input.SourceRef, input.PageSlug, input.Title, input.Content}, "\n")) {
		return WorkflowRAGFragment{}, WorkflowRAGFailureSecretForbidden
	}
	return WorkflowRAGFragment{
		SchemaVersion: workflowRAGFragmentSchemaVersion, FragmentRef: input.FragmentRef,
		SourceType: input.SourceType, SourceRef: input.SourceRef, PageSlug: input.PageSlug,
		Title: input.Title, IsOfficial: input.IsOfficial, Content: input.Content,
		ContentClassification: classification, ContentBytes: contentBytes, ContentDigest: workflowRAGSHA256(input.Content),
	}, ""
}

func buildWorkflowRAGSnapshotRecord(ctx WorkflowRAGSnapshotContext, snapshotID, key, displayName, classification string, version int, createdAt string, fragments []WorkflowRAGFragment, totalBytes int) (WorkflowRAGSnapshotRecord, error) {
	record := WorkflowRAGSnapshotRecord{
		SchemaVersion: workflowRAGSnapshotSchemaVersion, SnapshotID: snapshotID,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		SnapshotKey: key, RAGRef: fmt.Sprintf("workflow.rag.%s.v%d", key, version), SnapshotVersion: version,
		DisplayName: displayName, LifecycleState: workflowRAGSnapshotActive, ContentClassification: classification,
		ProfileRef: workflowRAGProfileID, FragmentCount: len(fragments), TotalContentBytes: totalBytes,
		CreatedAt: createdAt, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		Fragments: cloneWorkflowRAGFragments(fragments),
	}
	digest, err := workflowRAGSnapshotDigest(record)
	if err != nil {
		return WorkflowRAGSnapshotRecord{}, err
	}
	record.SnapshotDigest = digest
	return record, nil
}

func workflowRAGSnapshotDigest(record WorkflowRAGSnapshotRecord) (string, error) {
	type digestFragment struct {
		FragmentRef           string `json:"fragment_ref"`
		SourceType            string `json:"source_type"`
		SourceRef             string `json:"source_ref"`
		PageSlug              string `json:"page_slug"`
		Title                 string `json:"title"`
		IsOfficial            bool   `json:"is_official"`
		ContentClassification string `json:"content_classification"`
		ContentBytes          int    `json:"content_bytes"`
		ContentDigest         string `json:"content_digest"`
	}
	fragments := make([]digestFragment, 0, len(record.Fragments))
	for _, fragment := range record.Fragments {
		fragments = append(fragments, digestFragment{
			FragmentRef: fragment.FragmentRef, SourceType: fragment.SourceType, SourceRef: fragment.SourceRef,
			PageSlug: fragment.PageSlug, Title: fragment.Title, IsOfficial: fragment.IsOfficial,
			ContentClassification: fragment.ContentClassification, ContentBytes: fragment.ContentBytes,
			ContentDigest: fragment.ContentDigest,
		})
	}
	sort.Slice(fragments, func(i, j int) bool { return fragments[i].FragmentRef < fragments[j].FragmentRef })
	canonical := struct {
		SchemaVersion         string           `json:"schema_version"`
		SnapshotID            string           `json:"snapshot_id"`
		TenantRef             string           `json:"tenant_ref"`
		WorkspaceID           string           `json:"workspace_id"`
		ApplicationID         string           `json:"application_id"`
		SnapshotKey           string           `json:"snapshot_key"`
		SnapshotVersion       int              `json:"snapshot_version"`
		DisplayName           string           `json:"display_name"`
		ContentClassification string           `json:"content_classification"`
		ProfileRef            string           `json:"profile_ref"`
		Fragments             []digestFragment `json:"fragments"`
	}{
		SchemaVersion: record.SchemaVersion, SnapshotID: record.SnapshotID, TenantRef: record.TenantRef,
		WorkspaceID: record.WorkspaceID, ApplicationID: record.ApplicationID, SnapshotKey: record.SnapshotKey,
		SnapshotVersion: record.SnapshotVersion, DisplayName: record.DisplayName,
		ContentClassification: record.ContentClassification, ProfileRef: record.ProfileRef, Fragments: fragments,
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func workflowRAGResourceFromRecord(record WorkflowRAGSnapshotRecord, at string) WorkflowRAGSnapshotResource {
	return WorkflowRAGSnapshotResource{
		SnapshotID: record.SnapshotID, TenantRef: record.TenantRef, WorkspaceID: record.WorkspaceID,
		ApplicationID: record.ApplicationID, SnapshotKey: record.SnapshotKey, DisplayName: record.DisplayName,
		LifecycleState: workflowRAGSnapshotActive, LatestVersion: record.SnapshotVersion, LatestRAGRef: record.RAGRef,
		LatestDigest: record.SnapshotDigest, FragmentCount: record.FragmentCount, TotalContentBytes: record.TotalContentBytes,
		CreatedAt: at, UpdatedAt: at,
	}
}

func buildWorkflowRAGSnapshotAudit(ctx WorkflowRAGSnapshotContext, eventKind string, record WorkflowRAGSnapshotRecord, occurredAt string) (WorkflowRAGExecutionAudit, error) {
	eventID, err := newWorkflowRAGStableID("rage_")
	if err != nil {
		return WorkflowRAGExecutionAudit{}, err
	}
	return WorkflowRAGExecutionAudit{
		SchemaVersion: workflowRAGAuditSchemaVersion, EventID: eventID, EventKind: eventKind,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		SnapshotID: record.SnapshotID, SnapshotKey: record.SnapshotKey, SnapshotVersion: record.SnapshotVersion,
		SnapshotDigest: record.SnapshotDigest, FragmentCount: record.FragmentCount, TotalContentBytes: record.TotalContentBytes,
		ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, OccurredAt: occurredAt,
	}, nil
}

func validateWorkflowRAGContext(ctx WorkflowRAGSnapshotContext) string {
	if !workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.TenantRef)) ||
		!workflowRAGScopedIDPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) ||
		!workflowRAGScopedIDPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) ||
		!workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.ActorRef)) ||
		!workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.RequestID)) ||
		!workflowRAGReferencePattern.MatchString(strings.TrimSpace(ctx.AuditRef)) {
		return WorkflowRAGFailureScopeDenied
	}
	return ""
}

func validateStoredWorkflowRAGRecord(record WorkflowRAGSnapshotRecord, ctx WorkflowRAGSnapshotContext) error {
	if record.SchemaVersion != workflowRAGSnapshotSchemaVersion || !workflowRAGSnapshotIDPattern.MatchString(record.SnapshotID) ||
		record.TenantRef != ctx.TenantRef || record.WorkspaceID != ctx.WorkspaceID || record.ApplicationID != ctx.ApplicationID ||
		!workflowRAGSnapshotKeyPattern.MatchString(record.SnapshotKey) || record.SnapshotVersion < 1 ||
		record.RAGRef != fmt.Sprintf("workflow.rag.%s.v%d", record.SnapshotKey, record.SnapshotVersion) ||
		record.ProfileRef != workflowRAGProfileID || record.FragmentCount != len(record.Fragments) ||
		record.FragmentCount < 1 || record.FragmentCount > workflowRAGMaxFragments || record.TotalContentBytes < 1 ||
		record.TotalContentBytes > workflowRAGMaxSnapshotBytes || !workflowRAGDigestPattern.MatchString(record.SnapshotDigest) {
		return errWorkflowRAGStoreContract
	}
	total := 0
	for _, fragment := range record.Fragments {
		if validateStoredWorkflowRAGFragment(fragment, record.ContentClassification) != nil {
			return errWorkflowRAGStoreContract
		}
		total += fragment.ContentBytes
	}
	digest, err := workflowRAGSnapshotDigest(record)
	if err != nil || digest != record.SnapshotDigest || total != record.TotalContentBytes {
		return errWorkflowRAGStoreContract
	}
	return nil
}

func validateStoredWorkflowRAGFragment(fragment WorkflowRAGFragment, classification string) error {
	normalized, failure := normalizeWorkflowRAGFragment(WorkflowRAGFragmentInput{
		FragmentRef: fragment.FragmentRef, SourceType: fragment.SourceType, SourceRef: fragment.SourceRef,
		PageSlug: fragment.PageSlug, Title: fragment.Title, IsOfficial: fragment.IsOfficial, Content: fragment.Content,
	}, classification)
	if failure != "" || normalized != fragment {
		return errWorkflowRAGStoreContract
	}
	return nil
}

func workflowRAGRepositoryFailure(err error) WorkflowRAGSnapshotResult {
	result := WorkflowRAGSnapshotResult{FailureCode: WorkflowRAGFailureStoreUnavailable}
	var conflict workflowRAGVersionConflictError
	switch {
	case errors.As(err, &conflict):
		result.FailureCode = WorkflowRAGFailureVersionConflict
		result.CurrentLatestVersion = conflict.CurrentVersion
		result.CurrentLifecycle = conflict.CurrentState
	case errors.Is(err, errWorkflowRAGScopeDenied):
		result.FailureCode = WorkflowRAGFailureScopeDenied
	case errors.Is(err, errWorkflowRAGNotFound):
		result.FailureCode = WorkflowRAGFailureNotFound
	case errors.Is(err, errWorkflowRAGArchived):
		result.FailureCode = WorkflowRAGFailureArchived
	}
	return result
}

func workflowRAGClassificationAllowed(value string) bool {
	return value == "public" || value == "workspace_internal"
}

func workflowRAGSourceTypeAllowed(value string) bool {
	switch value {
	case "document", "wiki", "faq", "forum", "manual":
		return true
	default:
		return false
	}
}

func workflowRAGContainsForbiddenMaterial(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"authorization:", "bearer ", "api_key=", "api-key=", "x-api-key", "cookie:",
		"password=", "passwd=", "secret=", "token=", "sk-", "-----begin private key-----",
		"postgresql://", "postgres://", "mysql://", "mongodb://",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func newWorkflowRAGStableID(prefix string) (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func workflowRAGSHA256(value string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(value)))
}

func parseWorkflowRAGRAGRef(value string) (string, int, bool) {
	value = strings.TrimSpace(value)
	if !workflowRAGRAGRefPattern.MatchString(value) {
		return "", 0, false
	}
	marker := strings.LastIndex(value, ".v")
	if marker < 0 {
		return "", 0, false
	}
	version := 0
	if _, err := fmt.Sscanf(value[marker+2:], "%d", &version); err != nil || version < 1 {
		return "", 0, false
	}
	return strings.TrimPrefix(value[:marker], "workflow.rag."), version, true
}

func workflowRAGStoreKey(ctx WorkflowRAGSnapshotContext, snapshotID string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, snapshotID}, "\x00")
}

func cloneWorkflowRAGFragments(fragments []WorkflowRAGFragment) []WorkflowRAGFragment {
	return append([]WorkflowRAGFragment(nil), fragments...)
}

func cloneWorkflowRAGSnapshotRecord(record WorkflowRAGSnapshotRecord) WorkflowRAGSnapshotRecord {
	record.Fragments = cloneWorkflowRAGFragments(record.Fragments)
	return record
}

func cloneWorkflowRAGResource(resource WorkflowRAGSnapshotResource) WorkflowRAGSnapshotResource {
	if resource.ArchivedAt != nil {
		archivedAt := *resource.ArchivedAt
		resource.ArchivedAt = &archivedAt
	}
	return resource
}

func (repository *memoryWorkflowRAGSnapshotRepository) Create(ctx WorkflowRAGSnapshotContext, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	if repository == nil || repository.ownerLock == nil || !repository.available || validateStoredWorkflowRAGRecord(record, ctx) != nil {
		return errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	if len(repository.entries) >= repository.capacity {
		return errWorkflowRAGStoreUnavailable
	}
	for _, entry := range repository.entries {
		if entry.resource.TenantRef == ctx.TenantRef && entry.resource.WorkspaceID == ctx.WorkspaceID && entry.resource.ApplicationID == ctx.ApplicationID && entry.resource.SnapshotKey == resource.SnapshotKey {
			return workflowRAGVersionConflictError{CurrentVersion: entry.resource.LatestVersion, CurrentState: entry.resource.LifecycleState}
		}
	}
	key := workflowRAGStoreKey(ctx, resource.SnapshotID)
	if _, exists := repository.entries[key]; exists {
		return workflowRAGVersionConflictError{}
	}
	repository.entries[key] = workflowRAGMemoryEntry{
		resource: cloneWorkflowRAGResource(resource),
		versions: map[int]WorkflowRAGSnapshotRecord{record.SnapshotVersion: cloneWorkflowRAGSnapshotRecord(record)},
		audits:   []WorkflowRAGExecutionAudit{audit},
	}
	return nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGSnapshotListQuery) ([]WorkflowRAGSnapshotResource, error) {
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return nil, errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	records := make([]WorkflowRAGSnapshotResource, 0)
	for _, entry := range repository.entries {
		resource := entry.resource
		if resource.TenantRef != ctx.TenantRef || resource.WorkspaceID != ctx.WorkspaceID || resource.ApplicationID != ctx.ApplicationID || resource.LifecycleState != query.LifecycleState ||
			(query.AfterSnapshotKey != "" && resource.SnapshotKey <= query.AfterSnapshotKey) {
			continue
		}
		records = append(records, cloneWorkflowRAGResource(resource))
	}
	sort.Slice(records, func(i, j int) bool { return records[i].SnapshotKey < records[j].SnapshotKey })
	if len(records) > query.Limit {
		records = records[:query.Limit]
	}
	return records, nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, version int) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	entry, exists := repository.entries[workflowRAGStoreKey(ctx, snapshotID)]
	if !exists {
		for _, candidate := range repository.entries {
			if candidate.resource.SnapshotID == snapshotID {
				return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGScopeDenied
			}
		}
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	record, exists := entry.versions[version]
	if !exists {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	return cloneWorkflowRAGResource(entry.resource), cloneWorkflowRAGSnapshotRecord(record), nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) ReadByRAGRef(ctx WorkflowRAGSnapshotContext, ragRef string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	key, version, ok := parseWorkflowRAGRAGRef(ragRef)
	if !ok {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	for storageKey, entry := range repository.entries {
		if strings.HasPrefix(storageKey, workflowRAGStoreKey(ctx, "")) && entry.resource.SnapshotKey == key {
			record, found := entry.versions[version]
			if !found {
				return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
			}
			return cloneWorkflowRAGResource(entry.resource), cloneWorkflowRAGSnapshotRecord(record), nil
		}
	}
	for _, entry := range repository.entries {
		if entry.resource.SnapshotKey == key {
			return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGScopeDenied
		}
	}
	return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
}

func (repository *memoryWorkflowRAGSnapshotRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, snapshotID string) (WorkflowRAGSnapshotResource, WorkflowRAGSnapshotRecord, error) {
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.RLock()
	defer repository.ownerLock.RUnlock()
	entry, exists := repository.entries[workflowRAGStoreKey(ctx, snapshotID)]
	if !exists {
		for _, candidate := range repository.entries {
			if candidate.resource.SnapshotID == snapshotID {
				return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGScopeDenied
			}
		}
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGNotFound
	}
	record, exists := entry.versions[entry.resource.LatestVersion]
	if !exists {
		return WorkflowRAGSnapshotResource{}, WorkflowRAGSnapshotRecord{}, errWorkflowRAGStoreContract
	}
	return cloneWorkflowRAGResource(entry.resource), cloneWorkflowRAGSnapshotRecord(record), nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) CreateVersion(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, record WorkflowRAGSnapshotRecord, audit WorkflowRAGExecutionAudit) error {
	if repository == nil || repository.ownerLock == nil || !repository.available || validateStoredWorkflowRAGRecord(record, ctx) != nil {
		return errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	key := workflowRAGStoreKey(ctx, snapshotID)
	entry, exists := repository.entries[key]
	if !exists {
		return errWorkflowRAGNotFound
	}
	if entry.resource.LatestVersion != expectedVersion {
		return workflowRAGVersionConflictError{CurrentVersion: entry.resource.LatestVersion, CurrentState: entry.resource.LifecycleState}
	}
	if entry.resource.LifecycleState != workflowRAGSnapshotActive {
		return errWorkflowRAGArchived
	}
	if record.SnapshotVersion != expectedVersion+1 {
		return errWorkflowRAGStoreContract
	}
	entry.resource = cloneWorkflowRAGResource(resource)
	entry.versions[record.SnapshotVersion] = cloneWorkflowRAGSnapshotRecord(record)
	entry.audits = append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) Archive(ctx WorkflowRAGSnapshotContext, snapshotID string, expectedVersion int, resource WorkflowRAGSnapshotResource, audit WorkflowRAGExecutionAudit) error {
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return errWorkflowRAGStoreUnavailable
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	key := workflowRAGStoreKey(ctx, snapshotID)
	entry, exists := repository.entries[key]
	if !exists {
		return errWorkflowRAGNotFound
	}
	if entry.resource.LatestVersion != expectedVersion {
		return workflowRAGVersionConflictError{CurrentVersion: entry.resource.LatestVersion, CurrentState: entry.resource.LifecycleState}
	}
	if entry.resource.LifecycleState != workflowRAGSnapshotActive {
		return errWorkflowRAGArchived
	}
	entry.resource = cloneWorkflowRAGResource(resource)
	entry.audits = append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGSnapshotRepository) AppendAudit(ctx WorkflowRAGSnapshotContext, audit WorkflowRAGExecutionAudit) error {
	if repository == nil || repository.ownerLock == nil || !repository.available {
		return errWorkflowRAGStoreUnavailable
	}
	if _, err := encodeWorkflowRAGAudit(audit, ctx); err != nil {
		return err
	}
	repository.ownerLock.Lock()
	defer repository.ownerLock.Unlock()
	storageKey := workflowRAGStoreKey(ctx, audit.SnapshotID)
	entry, found := repository.entries[storageKey]
	if !found {
		return errWorkflowRAGNotFound
	}
	for _, existing := range entry.audits {
		if existing.EventID == audit.EventID || existing.AuditRef == audit.AuditRef {
			return errWorkflowRAGStoreContract
		}
	}
	entry.audits = append(entry.audits, audit)
	repository.entries[storageKey] = entry
	return nil
}

var _ workflowRAGSnapshotRepository = (*memoryWorkflowRAGSnapshotRepository)(nil)
