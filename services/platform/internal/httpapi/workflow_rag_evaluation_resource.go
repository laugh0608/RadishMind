package httpapi

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	workflowRAGEvaluationResourceSchemaVersion = "workflow_rag_evaluation_dataset_resource.v1"
	workflowRAGCandidateReviewSchemaVersion    = "workflow_rag_candidate_snapshot_review.v1"
	workflowRAGEvaluationAuditSchemaVersion    = "workflow_rag_evaluation_audit.v1"

	workflowRAGEvaluationActive   = "active"
	workflowRAGEvaluationArchived = "archived"

	workflowRAGEvaluationDefaultListLimit = 50
	workflowRAGEvaluationMaxListLimit     = 200
	workflowRAGEvaluationMaxResources     = 1024
	workflowRAGEvaluationMaxReviews       = 4096
)

const (
	WorkflowRAGEvaluationFailureScopeDenied            = "workflow_rag_evaluation_scope_denied"
	WorkflowRAGEvaluationFailureNotFound               = "workflow_rag_evaluation_not_found"
	WorkflowRAGEvaluationFailurePayloadInvalid         = "workflow_rag_evaluation_payload_invalid"
	WorkflowRAGEvaluationFailureBindingInvalid         = "workflow_rag_evaluation_binding_invalid"
	WorkflowRAGEvaluationFailureClassificationMismatch = "workflow_rag_evaluation_classification_mismatch"
	WorkflowRAGEvaluationFailureVersionConflict        = "workflow_rag_evaluation_version_conflict"
	WorkflowRAGEvaluationFailureArchived               = "workflow_rag_evaluation_archived"
	WorkflowRAGEvaluationFailureReviewFailed           = "workflow_rag_evaluation_review_failed"
	WorkflowRAGEvaluationFailureStoreUnavailable       = "workflow_rag_evaluation_store_unavailable"
)

var (
	workflowRAGEvaluationReviewIDPattern = regexp.MustCompile(`^wragr_[a-z2-7]{16}$`)
	workflowRAGEvaluationAuditIDPattern  = regexp.MustCompile(`^wraga_[a-z2-7]{16}$`)
	errWorkflowRAGEvaluationNotFound     = errors.New(WorkflowRAGEvaluationFailureNotFound)
	errWorkflowRAGEvaluationScopeDenied  = errors.New(WorkflowRAGEvaluationFailureScopeDenied)
	errWorkflowRAGEvaluationConflict     = errors.New(WorkflowRAGEvaluationFailureVersionConflict)
	errWorkflowRAGEvaluationArchived     = errors.New(WorkflowRAGEvaluationFailureArchived)
	errWorkflowRAGEvaluationStore        = errors.New(WorkflowRAGEvaluationFailureStoreUnavailable)
	errWorkflowRAGEvaluationContract     = errors.New("workflow RAG evaluation resource store contract mismatch")
)

type workflowRAGEvaluationConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure workflowRAGEvaluationConflictError) Error() string {
	return WorkflowRAGEvaluationFailureVersionConflict
}
func (failure workflowRAGEvaluationConflictError) Is(target error) bool {
	return target == errWorkflowRAGEvaluationConflict
}

type WorkflowRAGEvaluationDatasetResource struct {
	SchemaVersion         string                               `json:"schema_version"`
	DatasetID             string                               `json:"dataset_id"`
	TenantRef             string                               `json:"tenant_ref"`
	WorkspaceID           string                               `json:"workspace_id"`
	ApplicationID         string                               `json:"application_id"`
	DatasetKey            string                               `json:"dataset_key"`
	DisplayName           string                               `json:"display_name"`
	LifecycleState        string                               `json:"lifecycle_state"`
	ContentClassification string                               `json:"content_classification"`
	LatestVersion         int                                  `json:"latest_version"`
	LatestDigest          string                               `json:"latest_digest"`
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding `json:"baseline_snapshot"`
	SampleCount           int                                  `json:"sample_count"`
	CreatedAt             string                               `json:"created_at"`
	UpdatedAt             string                               `json:"updated_at"`
	ArchivedAt            *string                              `json:"archived_at"`
}

type WorkflowRAGEvaluationDatasetVersion struct {
	SchemaVersion     string                       `json:"schema_version"`
	DatasetKey        string                       `json:"dataset_key"`
	DisplayName       string                       `json:"display_name"`
	LifecycleState    string                       `json:"lifecycle_state"`
	CreatedAt         string                       `json:"created_at"`
	CreatedByActorRef string                       `json:"created_by_actor_ref"`
	RequestID         string                       `json:"request_id"`
	AuditRef          string                       `json:"audit_ref"`
	Dataset           WorkflowRAGEvaluationDataset `json:"dataset"`
}

type WorkflowRAGEvaluationAudit struct {
	SchemaVersion         string                               `json:"schema_version"`
	EventID               string                               `json:"event_id"`
	EventKind             string                               `json:"event_kind"`
	TenantRef             string                               `json:"tenant_ref"`
	WorkspaceID           string                               `json:"workspace_id"`
	ApplicationID         string                               `json:"application_id"`
	DatasetID             string                               `json:"dataset_id"`
	DatasetVersion        int                                  `json:"dataset_version"`
	DatasetDigest         string                               `json:"dataset_digest"`
	ContentClassification string                               `json:"content_classification"`
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding `json:"baseline_snapshot"`
	SampleCount           int                                  `json:"sample_count"`
	ActorRef              string                               `json:"actor_ref"`
	RequestID             string                               `json:"request_id"`
	AuditRef              string                               `json:"audit_ref"`
	OccurredAt            string                               `json:"occurred_at"`
}

type WorkflowRAGMetricDelta struct {
	HitAtK                    float64 `json:"hit_at_k"`
	ExpectedRecallAtK         float64 `json:"expected_recall_at_k"`
	RequiredOfficialRecallAtK float64 `json:"required_official_recall_at_k"`
	MeanReciprocalRank        float64 `json:"mean_reciprocal_rank"`
	NoEvidenceAccuracy        float64 `json:"no_evidence_accuracy"`
	SamplePassRate            float64 `json:"sample_pass_rate"`
}

type WorkflowRAGCandidateSampleComparison struct {
	SampleID                          string `json:"sample_id"`
	BaselinePassed                    bool   `json:"baseline_passed"`
	CandidatePassed                   bool   `json:"candidate_passed"`
	BaselineExpectedHitCount          int    `json:"baseline_expected_hit_count"`
	CandidateExpectedHitCount         int    `json:"candidate_expected_hit_count"`
	BaselineRequiredOfficialHitCount  int    `json:"baseline_required_official_hit_count"`
	CandidateRequiredOfficialHitCount int    `json:"candidate_required_official_hit_count"`
	BaselineFirstExpectedRank         int    `json:"baseline_first_expected_rank"`
	CandidateFirstExpectedRank        int    `json:"candidate_first_expected_rank"`
	Change                            string `json:"change"`
}

type WorkflowRAGCandidateSnapshotReview struct {
	SchemaVersion       string                                 `json:"schema_version"`
	ReviewID            string                                 `json:"review_id"`
	TenantRef           string                                 `json:"tenant_ref"`
	WorkspaceID         string                                 `json:"workspace_id"`
	ApplicationID       string                                 `json:"application_id"`
	Dataset             WorkflowRAGQualityDatasetBinding       `json:"dataset"`
	Profile             WorkflowRAGEvaluationProfileBinding    `json:"profile"`
	BaselineSnapshot    WorkflowRAGEvaluationSnapshotBinding   `json:"baseline_snapshot"`
	CandidateSnapshot   WorkflowRAGEvaluationSnapshotBinding   `json:"candidate_snapshot"`
	BaselineLifecycle   string                                 `json:"baseline_lifecycle"`
	CandidateLifecycle  string                                 `json:"candidate_lifecycle"`
	Conclusion          string                                 `json:"conclusion"`
	Baseline            WorkflowRAGQualityReview               `json:"baseline"`
	Candidate           WorkflowRAGQualityReview               `json:"candidate"`
	MetricDelta         WorkflowRAGMetricDelta                 `json:"metric_delta"`
	Samples             []WorkflowRAGCandidateSampleComparison `json:"samples"`
	AddedFindingCodes   []string                               `json:"added_finding_codes"`
	RemovedFindingCodes []string                               `json:"removed_finding_codes"`
	CreatedAt           string                                 `json:"created_at"`
	CreatedByActorRef   string                                 `json:"created_by_actor_ref"`
	RequestID           string                                 `json:"request_id"`
	AuditRef            string                                 `json:"audit_ref"`
}

type WorkflowRAGEvaluationDatasetCreateInput struct {
	DatasetKey            string
	DisplayName           string
	ContentClassification string
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding
	Thresholds            WorkflowRAGEvaluationThresholds
	ReviewSummary         string
	Samples               []WorkflowRAGEvaluationSample
}

type WorkflowRAGEvaluationDatasetVersionInput struct {
	ExpectedLatestVersion int
	DisplayName           string
	ContentClassification string
	BaselineSnapshot      WorkflowRAGEvaluationSnapshotBinding
	Thresholds            WorkflowRAGEvaluationThresholds
	ReviewSummary         string
	Samples               []WorkflowRAGEvaluationSample
}

type WorkflowRAGEvaluationListInput struct {
	LifecycleState string
	Limit          int
	Cursor         string
}

type WorkflowRAGCandidateReviewInput struct {
	DatasetVersion    int
	DatasetDigest     string
	CandidateSnapshot WorkflowRAGEvaluationSnapshotBinding
}

type WorkflowRAGEvaluationResult struct {
	Resource             *WorkflowRAGEvaluationDatasetResource
	Version              *WorkflowRAGEvaluationDatasetVersion
	Review               *WorkflowRAGCandidateSnapshotReview
	FailureCode          string
	CurrentLatestVersion int
	CurrentLifecycle     string
}

type WorkflowRAGEvaluationListResult struct {
	Resources   []WorkflowRAGEvaluationDatasetResource
	NextCursor  *string
	FailureCode string
}

type WorkflowRAGCandidateReviewListResult struct {
	Reviews     []WorkflowRAGCandidateSnapshotReview
	NextCursor  *string
	FailureCode string
}

type workflowRAGEvaluationListQuery struct {
	LifecycleState  string
	Limit           int
	AfterDatasetKey string
}

type workflowRAGCandidateReviewListQuery struct {
	Limit          int
	BeforeAt       string
	BeforeReviewID string
}

type workflowRAGEvaluationDatasetRepository interface {
	Create(WorkflowRAGSnapshotContext, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, WorkflowRAGEvaluationAudit) error
	List(WorkflowRAGSnapshotContext, workflowRAGEvaluationListQuery) ([]WorkflowRAGEvaluationDatasetResource, error)
	ReadLatest(WorkflowRAGSnapshotContext, string) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error)
	ReadVersion(WorkflowRAGSnapshotContext, string, int) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error)
	CreateVersion(WorkflowRAGSnapshotContext, string, int, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, WorkflowRAGEvaluationAudit) error
	Archive(WorkflowRAGSnapshotContext, string, int, WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationAudit) error
	CreateReview(WorkflowRAGSnapshotContext, WorkflowRAGCandidateSnapshotReview, WorkflowRAGEvaluationAudit) error
	ReadReview(WorkflowRAGSnapshotContext, string, string) (WorkflowRAGCandidateSnapshotReview, error)
	ListReviews(WorkflowRAGSnapshotContext, string, workflowRAGCandidateReviewListQuery) ([]WorkflowRAGCandidateSnapshotReview, error)
}

type workflowRAGEvaluationMemoryEntry struct {
	resource WorkflowRAGEvaluationDatasetResource
	versions map[int]WorkflowRAGEvaluationDatasetVersion
	reviews  map[string]WorkflowRAGCandidateSnapshotReview
	audits   []WorkflowRAGEvaluationAudit
}

type memoryWorkflowRAGEvaluationDatasetRepository struct {
	mu        *sync.RWMutex
	entries   map[string]workflowRAGEvaluationMemoryEntry
	capacity  int
	available bool
}

type workflowRAGEvaluationDatasetService struct {
	repository         workflowRAGEvaluationDatasetRepository
	snapshotRepository workflowRAGSnapshotRepository
	ranker             workflowRAGQualityRanker
	now                func() time.Time
	newID              func(string) (string, error)
}

func newMemoryWorkflowRAGEvaluationDatasetRepository(ownerLock *sync.RWMutex) *memoryWorkflowRAGEvaluationDatasetRepository {
	if ownerLock == nil {
		ownerLock = &sync.RWMutex{}
	}
	return &memoryWorkflowRAGEvaluationDatasetRepository{mu: ownerLock, entries: make(map[string]workflowRAGEvaluationMemoryEntry), capacity: workflowRAGEvaluationMaxResources, available: true}
}

func newWorkflowRAGEvaluationDatasetService(repository workflowRAGEvaluationDatasetRepository, snapshots workflowRAGSnapshotRepository) workflowRAGEvaluationDatasetService {
	return workflowRAGEvaluationDatasetService{repository: repository, snapshotRepository: snapshots, ranker: RankWorkflowRAGFragments, now: func() time.Time { return time.Now().UTC() }, newID: newWorkflowRAGStableID}
}

func (service workflowRAGEvaluationDatasetService) Create(ctx WorkflowRAGSnapshotContext, input WorkflowRAGEvaluationDatasetCreateInput) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	key, name, failure := normalizeWorkflowRAGEvaluationResourceMetadata(input.DatasetKey, input.DisplayName)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	baseline, failure := service.readBoundSnapshot(ctx, input.BaselineSnapshot, input.ContentClassification)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	datasetID, err := service.newID("wragd_")
	if err != nil || !workflowRAGDatasetIDPattern.MatchString(datasetID) {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	version, err := buildWorkflowRAGEvaluationDatasetVersion(ctx, datasetID, key, name, input.ContentClassification, 1, input.Thresholds, input.ReviewSummary, input.Samples, baseline, at)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	resource := workflowRAGEvaluationResourceFromVersion(ctx, version, at)
	audit, err := service.buildAudit(ctx, "dataset_created", version, at)
	if err != nil || service.repository == nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	if err = service.repository.Create(ctx, resource, version, audit); err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	return WorkflowRAGEvaluationResult{Resource: &resource, Version: &version, CurrentLatestVersion: 1, CurrentLifecycle: workflowRAGEvaluationActive}
}

func (service workflowRAGEvaluationDatasetService) List(ctx WorkflowRAGSnapshotContext, input WorkflowRAGEvaluationListInput) WorkflowRAGEvaluationListResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: WorkflowRAGEvaluationFailureScopeDenied}
	}
	lifecycle := strings.TrimSpace(input.LifecycleState)
	if lifecycle == "" {
		lifecycle = workflowRAGEvaluationActive
	}
	limit := input.Limit
	if limit == 0 {
		limit = workflowRAGEvaluationDefaultListLimit
	}
	cursor := strings.TrimSpace(input.Cursor)
	if (lifecycle != workflowRAGEvaluationActive && lifecycle != workflowRAGEvaluationArchived) || limit < 1 || limit > workflowRAGEvaluationMaxListLimit || (cursor != "" && !workflowRAGSnapshotKeyPattern.MatchString(cursor)) {
		return WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: WorkflowRAGEvaluationFailurePayloadInvalid}
	}
	if service.repository == nil {
		return WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: WorkflowRAGEvaluationFailureStoreUnavailable}
	}
	resources, err := service.repository.List(ctx, workflowRAGEvaluationListQuery{LifecycleState: lifecycle, Limit: limit + 1, AfterDatasetKey: cursor})
	if err != nil {
		return WorkflowRAGEvaluationListResult{Resources: []WorkflowRAGEvaluationDatasetResource{}, FailureCode: workflowRAGEvaluationRepositoryFailure(err).FailureCode}
	}
	result := WorkflowRAGEvaluationListResult{Resources: resources}
	if len(resources) > limit {
		result.Resources = resources[:limit]
		next := result.Resources[len(result.Resources)-1].DatasetKey
		result.NextCursor = &next
	}
	return result
}

func (service workflowRAGEvaluationDatasetService) Read(ctx WorkflowRAGSnapshotContext, datasetID string, version int) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	if !workflowRAGDatasetIDPattern.MatchString(strings.TrimSpace(datasetID)) || version < 1 {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	if service.repository == nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	resource, record, err := service.repository.ReadVersion(ctx, datasetID, version)
	if err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	record.LifecycleState = resource.LifecycleState
	return WorkflowRAGEvaluationResult{Resource: &resource, Version: &record, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
}

func (service workflowRAGEvaluationDatasetService) Version(ctx WorkflowRAGSnapshotContext, datasetID string, input WorkflowRAGEvaluationDatasetVersionInput) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	datasetID = strings.TrimSpace(datasetID)
	if !workflowRAGDatasetIDPattern.MatchString(datasetID) || input.ExpectedLatestVersion < 1 || service.repository == nil {
		if service.repository == nil {
			return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
		}
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	resource, _, err := service.repository.ReadLatest(ctx, datasetID)
	if err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	if resource.LatestVersion != input.ExpectedLatestVersion {
		return WorkflowRAGEvaluationResult{FailureCode: WorkflowRAGEvaluationFailureVersionConflict, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGEvaluationActive {
		return WorkflowRAGEvaluationResult{FailureCode: WorkflowRAGEvaluationFailureArchived, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	_, name, failure := normalizeWorkflowRAGEvaluationResourceMetadata(resource.DatasetKey, input.DisplayName)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	baseline, failure := service.readBoundSnapshot(ctx, input.BaselineSnapshot, input.ContentClassification)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	version, err := buildWorkflowRAGEvaluationDatasetVersion(ctx, datasetID, resource.DatasetKey, name, input.ContentClassification, resource.LatestVersion+1, input.Thresholds, input.ReviewSummary, input.Samples, baseline, at)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	updated := resource
	updated.DisplayName, updated.ContentClassification = name, version.Dataset.ContentClassification
	updated.LatestVersion, updated.LatestDigest = version.Dataset.DatasetVersion, version.Dataset.DatasetDigest
	updated.BaselineSnapshot, updated.SampleCount, updated.UpdatedAt = version.Dataset.Snapshot, len(version.Dataset.Samples), at
	audit, err := service.buildAudit(ctx, "dataset_versioned", version, at)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	if err = service.repository.CreateVersion(ctx, datasetID, input.ExpectedLatestVersion, updated, version, audit); err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	return WorkflowRAGEvaluationResult{Resource: &updated, Version: &version, CurrentLatestVersion: updated.LatestVersion, CurrentLifecycle: updated.LifecycleState}
}

func (service workflowRAGEvaluationDatasetService) Archive(ctx WorkflowRAGSnapshotContext, datasetID string, expectedVersion int) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	datasetID = strings.TrimSpace(datasetID)
	if !workflowRAGDatasetIDPattern.MatchString(datasetID) || expectedVersion < 1 || service.repository == nil {
		if service.repository == nil {
			return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
		}
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	resource, version, err := service.repository.ReadLatest(ctx, datasetID)
	if err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	if resource.LatestVersion != expectedVersion {
		return WorkflowRAGEvaluationResult{FailureCode: WorkflowRAGEvaluationFailureVersionConflict, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	if resource.LifecycleState != workflowRAGEvaluationActive {
		return WorkflowRAGEvaluationResult{FailureCode: WorkflowRAGEvaluationFailureArchived, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	resource.LifecycleState, resource.UpdatedAt, resource.ArchivedAt = workflowRAGEvaluationArchived, at, &at
	version.LifecycleState = workflowRAGEvaluationArchived
	audit, err := service.buildAudit(ctx, "dataset_archived", version, at)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	if err = service.repository.Archive(ctx, datasetID, expectedVersion, resource, audit); err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	return WorkflowRAGEvaluationResult{Resource: &resource, Version: &version, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
}

func (service workflowRAGEvaluationDatasetService) CreateCandidateReview(ctx WorkflowRAGSnapshotContext, datasetID string, input WorkflowRAGCandidateReviewInput) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	if !workflowRAGDatasetIDPattern.MatchString(strings.TrimSpace(datasetID)) || input.DatasetVersion < 1 || !workflowRAGDigestPattern.MatchString(input.DatasetDigest) || service.repository == nil || service.ranker == nil {
		if service.repository == nil {
			return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
		}
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	resource, version, err := service.repository.ReadVersion(ctx, datasetID, input.DatasetVersion)
	if err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	if version.Dataset.DatasetDigest != input.DatasetDigest {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureBindingInvalid)
	}
	baseline, failure := service.readBoundSnapshot(ctx, version.Dataset.Snapshot, version.Dataset.ContentClassification)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	candidate, failure := service.readBoundSnapshot(ctx, input.CandidateSnapshot, version.Dataset.ContentClassification)
	if failure != "" {
		return workflowRAGEvaluationFailure(failure)
	}
	baselineReview, err := reviewWorkflowRAGApplicationDataset(baseline, version.Dataset, service.ranker)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureReviewFailed)
	}
	candidateReview, err := reviewWorkflowRAGCandidateSnapshot(candidate, version.Dataset, service.ranker)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureReviewFailed)
	}
	reviewID, err := service.newID("wragr_")
	if err != nil || !workflowRAGEvaluationReviewIDPattern.MatchString(reviewID) {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	at := service.now().UTC().Format(time.RFC3339Nano)
	review := buildWorkflowRAGCandidateReview(ctx, reviewID, at, baselineReview, candidateReview, baseline.LifecycleState, candidate.LifecycleState)
	if validateStoredWorkflowRAGCandidateReview(review, ctx) != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureReviewFailed)
	}
	audit, err := service.buildAudit(ctx, "candidate_review_created", version, at)
	if err != nil {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	}
	if err = service.repository.CreateReview(ctx, review, audit); err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	return WorkflowRAGEvaluationResult{Resource: &resource, Review: &review, CurrentLatestVersion: resource.LatestVersion, CurrentLifecycle: resource.LifecycleState}
}

func (service workflowRAGEvaluationDatasetService) ReadCandidateReview(ctx WorkflowRAGSnapshotContext, datasetID, reviewID string) WorkflowRAGEvaluationResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureScopeDenied)
	}
	if !workflowRAGDatasetIDPattern.MatchString(strings.TrimSpace(datasetID)) || !workflowRAGEvaluationReviewIDPattern.MatchString(strings.TrimSpace(reviewID)) || service.repository == nil {
		if service.repository == nil {
			return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
		}
		return workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailurePayloadInvalid)
	}
	review, err := service.repository.ReadReview(ctx, datasetID, reviewID)
	if err != nil {
		return workflowRAGEvaluationRepositoryFailure(err)
	}
	return WorkflowRAGEvaluationResult{Review: &review}
}

func (service workflowRAGEvaluationDatasetService) ListCandidateReviews(ctx WorkflowRAGSnapshotContext, datasetID string, limit int, cursor string) WorkflowRAGCandidateReviewListResult {
	if validateWorkflowRAGContext(ctx) != "" {
		return WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: WorkflowRAGEvaluationFailureScopeDenied}
	}
	if limit == 0 {
		limit = workflowRAGEvaluationDefaultListLimit
	}
	query, failure := parseWorkflowRAGCandidateReviewCursor(limit, cursor)
	if !workflowRAGDatasetIDPattern.MatchString(strings.TrimSpace(datasetID)) || failure != "" || service.repository == nil {
		if service.repository == nil {
			return WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: WorkflowRAGEvaluationFailureStoreUnavailable}
		}
		return WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: WorkflowRAGEvaluationFailurePayloadInvalid}
	}
	query.Limit++
	reviews, err := service.repository.ListReviews(ctx, datasetID, query)
	if err != nil {
		return WorkflowRAGCandidateReviewListResult{Reviews: []WorkflowRAGCandidateSnapshotReview{}, FailureCode: workflowRAGEvaluationRepositoryFailure(err).FailureCode}
	}
	result := WorkflowRAGCandidateReviewListResult{Reviews: reviews}
	if len(reviews) >= query.Limit {
		result.Reviews = reviews[:query.Limit-1]
		last := result.Reviews[len(result.Reviews)-1]
		next := last.CreatedAt + "|" + last.ReviewID
		result.NextCursor = &next
	}
	return result
}

func (service workflowRAGEvaluationDatasetService) readBoundSnapshot(ctx WorkflowRAGSnapshotContext, binding WorkflowRAGEvaluationSnapshotBinding, classification string) (WorkflowRAGSnapshotRecord, string) {
	if service.snapshotRepository == nil || !validateWorkflowRAGEvaluationSnapshotBinding(binding) || binding.TenantRef != ctx.TenantRef || binding.WorkspaceID != ctx.WorkspaceID || binding.ApplicationID != ctx.ApplicationID {
		if service.snapshotRepository == nil {
			return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureStoreUnavailable
		}
		return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureBindingInvalid
	}
	resource, snapshot, err := service.snapshotRepository.ReadVersion(ctx, binding.SnapshotID, binding.SnapshotVersion)
	if err != nil {
		if errors.Is(err, errWorkflowRAGScopeDenied) {
			return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureScopeDenied
		}
		if errors.Is(err, errWorkflowRAGNotFound) {
			return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureNotFound
		}
		return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureStoreUnavailable
	}
	if workflowRAGEvaluationSnapshotBinding(snapshot) != binding {
		return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureBindingInvalid
	}
	expected := "public"
	if classification == "workspace_internal" {
		expected = "workspace_internal"
	} else if classification != workflowRAGEvaluationClassification {
		return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailurePayloadInvalid
	}
	if snapshot.ContentClassification != expected {
		return WorkflowRAGSnapshotRecord{}, WorkflowRAGEvaluationFailureClassificationMismatch
	}
	snapshot.LifecycleState = resource.LifecycleState
	return snapshot, ""
}

func (service workflowRAGEvaluationDatasetService) buildAudit(ctx WorkflowRAGSnapshotContext, kind string, version WorkflowRAGEvaluationDatasetVersion, at string) (WorkflowRAGEvaluationAudit, error) {
	id, err := service.newID("wraga_")
	if err != nil {
		return WorkflowRAGEvaluationAudit{}, err
	}
	return WorkflowRAGEvaluationAudit{SchemaVersion: workflowRAGEvaluationAuditSchemaVersion, EventID: id, EventKind: kind, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, DatasetID: version.Dataset.DatasetID, DatasetVersion: version.Dataset.DatasetVersion, DatasetDigest: version.Dataset.DatasetDigest, ContentClassification: version.Dataset.ContentClassification, BaselineSnapshot: version.Dataset.Snapshot, SampleCount: len(version.Dataset.Samples), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, OccurredAt: at}, nil
}

func buildWorkflowRAGEvaluationDatasetVersion(ctx WorkflowRAGSnapshotContext, datasetID, key, name, classification string, version int, thresholds WorkflowRAGEvaluationThresholds, reviewSummary string, samples []WorkflowRAGEvaluationSample, baseline WorkflowRAGSnapshotRecord, at string) (WorkflowRAGEvaluationDatasetVersion, error) {
	reviewSummary = strings.TrimSpace(reviewSummary)
	if reviewSummary == "" || utf8.RuneCountInString(reviewSummary) > 512 || workflowRAGContainsForbiddenMaterial(reviewSummary) {
		return WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	profile := workflowRAGLexicalProfile()
	dataset := WorkflowRAGEvaluationDataset{SchemaVersion: workflowRAGEvaluationDatasetSchemaVersion, DatasetID: datasetID, DatasetVersion: version, ContentClassification: classification, Snapshot: workflowRAGEvaluationSnapshotBinding(baseline), Profile: WorkflowRAGEvaluationProfileBinding{ProfileID: profile.ProfileID, ProfileVersion: profile.ProfileVersion, ProfileDigest: profile.ProfileDigest}, Thresholds: thresholds, Review: WorkflowRAGEvaluationReviewMetadata{ReviewerRef: ctx.ActorRef, ReviewedAt: at, ReviewSummary: reviewSummary}, Samples: cloneWorkflowRAGEvaluationSamples(samples)}
	digest, err := WorkflowRAGEvaluationDatasetDigest(dataset)
	if err != nil {
		return WorkflowRAGEvaluationDatasetVersion{}, err
	}
	dataset.DatasetDigest = digest
	if validateWorkflowRAGApplicationEvaluationDataset(dataset, &baseline) != nil {
		return WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationContract
	}
	return WorkflowRAGEvaluationDatasetVersion{SchemaVersion: workflowRAGEvaluationResourceSchemaVersion, DatasetKey: key, DisplayName: name, LifecycleState: workflowRAGEvaluationActive, CreatedAt: at, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, Dataset: dataset}, nil
}

func normalizeWorkflowRAGEvaluationResourceMetadata(key, name string) (string, string, string) {
	key, name = strings.TrimSpace(key), strings.TrimSpace(name)
	if !workflowRAGSnapshotKeyPattern.MatchString(key) || !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 120 || workflowRAGContainsForbiddenMaterial(name) {
		return "", "", WorkflowRAGEvaluationFailurePayloadInvalid
	}
	return key, name, ""
}

func workflowRAGEvaluationResourceFromVersion(ctx WorkflowRAGSnapshotContext, version WorkflowRAGEvaluationDatasetVersion, at string) WorkflowRAGEvaluationDatasetResource {
	return WorkflowRAGEvaluationDatasetResource{SchemaVersion: workflowRAGEvaluationResourceSchemaVersion, DatasetID: version.Dataset.DatasetID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, DatasetKey: version.DatasetKey, DisplayName: version.DisplayName, LifecycleState: workflowRAGEvaluationActive, ContentClassification: version.Dataset.ContentClassification, LatestVersion: version.Dataset.DatasetVersion, LatestDigest: version.Dataset.DatasetDigest, BaselineSnapshot: version.Dataset.Snapshot, SampleCount: len(version.Dataset.Samples), CreatedAt: at, UpdatedAt: at}
}

func cloneWorkflowRAGEvaluationSamples(samples []WorkflowRAGEvaluationSample) []WorkflowRAGEvaluationSample {
	if samples == nil {
		return nil
	}
	cloned := make([]WorkflowRAGEvaluationSample, len(samples))
	for index, sample := range samples {
		cloned[index] = sample
		cloned[index].ExpectedCitationRefs = cloneStringSlice(sample.ExpectedCitationRefs)
		cloned[index].RequiredOfficialRefs = cloneStringSlice(sample.RequiredOfficialRefs)
	}
	return cloned
}

func parseWorkflowRAGCandidateReviewCursor(limit int, cursor string) (workflowRAGCandidateReviewListQuery, string) {
	if limit < 1 || limit > workflowRAGEvaluationMaxListLimit {
		return workflowRAGCandidateReviewListQuery{}, WorkflowRAGEvaluationFailurePayloadInvalid
	}
	query := workflowRAGCandidateReviewListQuery{Limit: limit}
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return query, ""
	}
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 || !workflowRAGEvaluationReviewIDPattern.MatchString(parts[1]) {
		return workflowRAGCandidateReviewListQuery{}, WorkflowRAGEvaluationFailurePayloadInvalid
	}
	if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
		return workflowRAGCandidateReviewListQuery{}, WorkflowRAGEvaluationFailurePayloadInvalid
	}
	query.BeforeAt, query.BeforeReviewID = parts[0], parts[1]
	return query, ""
}

func workflowRAGEvaluationFailure(code string) WorkflowRAGEvaluationResult {
	return WorkflowRAGEvaluationResult{FailureCode: code}
}

func workflowRAGEvaluationRepositoryFailure(err error) WorkflowRAGEvaluationResult {
	result := workflowRAGEvaluationFailure(WorkflowRAGEvaluationFailureStoreUnavailable)
	var conflict workflowRAGEvaluationConflictError
	switch {
	case errors.As(err, &conflict):
		result.FailureCode, result.CurrentLatestVersion, result.CurrentLifecycle = WorkflowRAGEvaluationFailureVersionConflict, conflict.CurrentVersion, conflict.CurrentState
	case errors.Is(err, errWorkflowRAGEvaluationScopeDenied):
		result.FailureCode = WorkflowRAGEvaluationFailureScopeDenied
	case errors.Is(err, errWorkflowRAGEvaluationNotFound):
		result.FailureCode = WorkflowRAGEvaluationFailureNotFound
	case errors.Is(err, errWorkflowRAGEvaluationArchived):
		result.FailureCode = WorkflowRAGEvaluationFailureArchived
	}
	return result
}

func workflowRAGEvaluationStoreKey(ctx WorkflowRAGSnapshotContext, datasetID string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, datasetID}, "\x00")
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) Create(ctx WorkflowRAGSnapshotContext, resource WorkflowRAGEvaluationDatasetResource, version WorkflowRAGEvaluationDatasetVersion, audit WorkflowRAGEvaluationAudit) error {
	if !repository.available || validateStoredWorkflowRAGEvaluationResource(resource, ctx) != nil || validateStoredWorkflowRAGEvaluationVersion(version, ctx) != nil || validateStoredWorkflowRAGEvaluationAudit(audit, ctx) != nil {
		return errWorkflowRAGEvaluationStore
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := workflowRAGEvaluationStoreKey(ctx, resource.DatasetID)
	if _, exists := repository.entries[key]; exists {
		return errWorkflowRAGEvaluationContract
	}
	for _, entry := range repository.entries {
		if entry.resource.TenantRef == ctx.TenantRef && entry.resource.WorkspaceID == ctx.WorkspaceID && entry.resource.ApplicationID == ctx.ApplicationID && entry.resource.DatasetKey == resource.DatasetKey {
			return errWorkflowRAGEvaluationContract
		}
	}
	if len(repository.entries) >= repository.capacity {
		return errWorkflowRAGEvaluationStore
	}
	repository.entries[key] = workflowRAGEvaluationMemoryEntry{resource: cloneWorkflowRAGEvaluationResource(resource), versions: map[int]WorkflowRAGEvaluationDatasetVersion{1: cloneWorkflowRAGEvaluationVersion(version)}, reviews: make(map[string]WorkflowRAGCandidateSnapshotReview), audits: []WorkflowRAGEvaluationAudit{audit}}
	return nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) List(ctx WorkflowRAGSnapshotContext, query workflowRAGEvaluationListQuery) ([]WorkflowRAGEvaluationDatasetResource, error) {
	if !repository.available {
		return nil, errWorkflowRAGEvaluationStore
	}
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	resources := make([]WorkflowRAGEvaluationDatasetResource, 0)
	for _, entry := range repository.entries {
		resource := entry.resource
		if resource.TenantRef == ctx.TenantRef && resource.WorkspaceID == ctx.WorkspaceID && resource.ApplicationID == ctx.ApplicationID && resource.LifecycleState == query.LifecycleState && (query.AfterDatasetKey == "" || resource.DatasetKey > query.AfterDatasetKey) {
			resources = append(resources, cloneWorkflowRAGEvaluationResource(resource))
		}
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].DatasetKey < resources[j].DatasetKey })
	if len(resources) > query.Limit {
		resources = resources[:query.Limit]
	}
	return resources, nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) ReadLatest(ctx WorkflowRAGSnapshotContext, datasetID string) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	return cloneWorkflowRAGEvaluationResource(entry.resource), cloneWorkflowRAGEvaluationVersion(entry.versions[entry.resource.LatestVersion]), nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) ReadVersion(ctx WorkflowRAGSnapshotContext, datasetID string, version int) (WorkflowRAGEvaluationDatasetResource, WorkflowRAGEvaluationDatasetVersion, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, err
	}
	record, found := entry.versions[version]
	if !found {
		return WorkflowRAGEvaluationDatasetResource{}, WorkflowRAGEvaluationDatasetVersion{}, errWorkflowRAGEvaluationNotFound
	}
	return cloneWorkflowRAGEvaluationResource(entry.resource), cloneWorkflowRAGEvaluationVersion(record), nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) CreateVersion(ctx WorkflowRAGSnapshotContext, datasetID string, expected int, resource WorkflowRAGEvaluationDatasetResource, version WorkflowRAGEvaluationDatasetVersion, audit WorkflowRAGEvaluationAudit) error {
	if validateStoredWorkflowRAGEvaluationResource(resource, ctx) != nil || validateStoredWorkflowRAGEvaluationVersion(version, ctx) != nil || validateStoredWorkflowRAGEvaluationAudit(audit, ctx) != nil {
		return errWorkflowRAGEvaluationContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := workflowRAGEvaluationStoreKey(ctx, datasetID)
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return err
	}
	if entry.resource.LatestVersion != expected {
		return workflowRAGEvaluationConflictError{CurrentVersion: entry.resource.LatestVersion, CurrentState: entry.resource.LifecycleState}
	}
	if entry.resource.LifecycleState != workflowRAGEvaluationActive {
		return errWorkflowRAGEvaluationArchived
	}
	entry.resource, entry.versions[version.Dataset.DatasetVersion], entry.audits = cloneWorkflowRAGEvaluationResource(resource), cloneWorkflowRAGEvaluationVersion(version), append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) Archive(ctx WorkflowRAGSnapshotContext, datasetID string, expected int, resource WorkflowRAGEvaluationDatasetResource, audit WorkflowRAGEvaluationAudit) error {
	if validateStoredWorkflowRAGEvaluationResource(resource, ctx) != nil || resource.LifecycleState != workflowRAGEvaluationArchived || validateStoredWorkflowRAGEvaluationAudit(audit, ctx) != nil {
		return errWorkflowRAGEvaluationContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := workflowRAGEvaluationStoreKey(ctx, datasetID)
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return err
	}
	if entry.resource.LatestVersion != expected {
		return workflowRAGEvaluationConflictError{CurrentVersion: entry.resource.LatestVersion, CurrentState: entry.resource.LifecycleState}
	}
	if entry.resource.LifecycleState != workflowRAGEvaluationActive {
		return errWorkflowRAGEvaluationArchived
	}
	entry.resource, entry.audits = cloneWorkflowRAGEvaluationResource(resource), append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) CreateReview(ctx WorkflowRAGSnapshotContext, review WorkflowRAGCandidateSnapshotReview, audit WorkflowRAGEvaluationAudit) error {
	if validateStoredWorkflowRAGCandidateReview(review, ctx) != nil || validateStoredWorkflowRAGEvaluationAudit(audit, ctx) != nil {
		return errWorkflowRAGEvaluationContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	key := workflowRAGEvaluationStoreKey(ctx, review.Dataset.DatasetID)
	entry, err := repository.readEntryLocked(ctx, review.Dataset.DatasetID)
	if err != nil {
		return err
	}
	if _, exists := entry.reviews[review.ReviewID]; exists || len(entry.reviews) >= workflowRAGEvaluationMaxReviews {
		return errWorkflowRAGEvaluationStore
	}
	entry.reviews[review.ReviewID], entry.audits = cloneWorkflowRAGCandidateReview(review), append(entry.audits, audit)
	repository.entries[key] = entry
	return nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) ReadReview(ctx WorkflowRAGSnapshotContext, datasetID, reviewID string) (WorkflowRAGCandidateSnapshotReview, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return WorkflowRAGCandidateSnapshotReview{}, err
	}
	review, found := entry.reviews[reviewID]
	if !found {
		return WorkflowRAGCandidateSnapshotReview{}, errWorkflowRAGEvaluationNotFound
	}
	return cloneWorkflowRAGCandidateReview(review), nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) ListReviews(ctx WorkflowRAGSnapshotContext, datasetID string, query workflowRAGCandidateReviewListQuery) ([]WorkflowRAGCandidateSnapshotReview, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	entry, err := repository.readEntryLocked(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	reviews := make([]WorkflowRAGCandidateSnapshotReview, 0, len(entry.reviews))
	for _, review := range entry.reviews {
		if query.BeforeAt != "" && (review.CreatedAt > query.BeforeAt || (review.CreatedAt == query.BeforeAt && review.ReviewID >= query.BeforeReviewID)) {
			continue
		}
		reviews = append(reviews, cloneWorkflowRAGCandidateReview(review))
	}
	sort.Slice(reviews, func(i, j int) bool {
		if reviews[i].CreatedAt == reviews[j].CreatedAt {
			return reviews[i].ReviewID > reviews[j].ReviewID
		}
		return reviews[i].CreatedAt > reviews[j].CreatedAt
	})
	if len(reviews) > query.Limit {
		reviews = reviews[:query.Limit]
	}
	return reviews, nil
}

func (repository *memoryWorkflowRAGEvaluationDatasetRepository) readEntryLocked(ctx WorkflowRAGSnapshotContext, datasetID string) (workflowRAGEvaluationMemoryEntry, error) {
	if !repository.available {
		return workflowRAGEvaluationMemoryEntry{}, errWorkflowRAGEvaluationStore
	}
	entry, found := repository.entries[workflowRAGEvaluationStoreKey(ctx, datasetID)]
	if found {
		return entry, nil
	}
	for _, candidate := range repository.entries {
		if candidate.resource.DatasetID == datasetID {
			return workflowRAGEvaluationMemoryEntry{}, errWorkflowRAGEvaluationScopeDenied
		}
	}
	return workflowRAGEvaluationMemoryEntry{}, errWorkflowRAGEvaluationNotFound
}

func cloneWorkflowRAGEvaluationResource(value WorkflowRAGEvaluationDatasetResource) WorkflowRAGEvaluationDatasetResource {
	if value.ArchivedAt != nil {
		archived := *value.ArchivedAt
		value.ArchivedAt = &archived
	}
	return value
}

func cloneWorkflowRAGEvaluationVersion(value WorkflowRAGEvaluationDatasetVersion) WorkflowRAGEvaluationDatasetVersion {
	value.Dataset.Samples = cloneWorkflowRAGEvaluationSamples(value.Dataset.Samples)
	return value
}

func cloneWorkflowRAGCandidateReview(value WorkflowRAGCandidateSnapshotReview) WorkflowRAGCandidateSnapshotReview {
	value.Baseline = cloneWorkflowRAGQualityReview(value.Baseline)
	value.Candidate = cloneWorkflowRAGQualityReview(value.Candidate)
	comparisons := make([]WorkflowRAGCandidateSampleComparison, len(value.Samples))
	copy(comparisons, value.Samples)
	value.Samples = comparisons
	value.AddedFindingCodes = cloneStringSlice(value.AddedFindingCodes)
	value.RemovedFindingCodes = cloneStringSlice(value.RemovedFindingCodes)
	return value
}

func cloneWorkflowRAGQualityReview(value WorkflowRAGQualityReview) WorkflowRAGQualityReview {
	samples := make([]WorkflowRAGQualitySampleResult, len(value.Samples))
	for index, sample := range value.Samples {
		samples[index] = sample
		samples[index].Selected = make([]WorkflowRAGQualitySelectedFragment, len(sample.Selected))
		copy(samples[index].Selected, sample.Selected)
	}
	findings := make([]WorkflowRAGQualityFinding, len(value.Findings))
	for index, finding := range value.Findings {
		findings[index] = finding
		findings[index].FragmentRefs = cloneStringSlice(finding.FragmentRefs)
	}
	distribution := make([]WorkflowRAGQualitySourceTypeCount, len(value.KnowledgeSummary.SourceTypeDistribution))
	copy(distribution, value.KnowledgeSummary.SourceTypeDistribution)
	value.Samples = samples
	value.Findings = findings
	value.KnowledgeSummary.SourceTypeDistribution = distribution
	return value
}
