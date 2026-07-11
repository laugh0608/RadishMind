package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	workflowEvaluationCaseSchema      = "workflow_evaluation_case.v1"
	workflowEvaluationDefaultLimit    = 25
	workflowEvaluationMaxLimit        = 100
	workflowEvaluationMaxCandidates   = 20
	defaultWorkflowEvaluationCapacity = 100
)

type WorkflowEvaluationFailureCode string

const (
	WorkflowEvaluationFailureInvalid          WorkflowEvaluationFailureCode = "workflow_evaluation_invalid"
	WorkflowEvaluationFailureRunNotEligible   WorkflowEvaluationFailureCode = "workflow_evaluation_run_not_eligible"
	WorkflowEvaluationFailureNotFound         WorkflowEvaluationFailureCode = "workflow_evaluation_not_found"
	WorkflowEvaluationFailureCursorInvalid    WorkflowEvaluationFailureCode = "workflow_evaluation_cursor_invalid"
	WorkflowEvaluationFailureStoreUnavailable WorkflowEvaluationFailureCode = "workflow_evaluation_store_unavailable"
	WorkflowEvaluationFailureStoreContract    WorkflowEvaluationFailureCode = "workflow_evaluation_store_contract_mismatch"
)

type WorkflowEvaluationExpectation struct {
	CandidateRunID         string                              `json:"candidate_run_id"`
	ExpectedClassification WorkflowRunComparisonClassification `json:"expected_classification"`
}

type WorkflowEvaluationCase struct {
	SchemaVersion string                          `json:"schema_version"`
	CaseID        string                          `json:"case_id"`
	Name          string                          `json:"name"`
	WorkspaceID   string                          `json:"workspace_id"`
	ApplicationID string                          `json:"application_id"`
	BaselineRunID string                          `json:"baseline_run_id"`
	Expectations  []WorkflowEvaluationExpectation `json:"expectations"`
	CreatedAt     string                          `json:"created_at"`
	ActorRef      string                          `json:"actor_ref"`
	RequestID     string                          `json:"request_id"`
	AuditRef      string                          `json:"audit_ref"`
}

type WorkflowEvaluationListFilter struct {
	BaselineRunID string
	BeforeTime    *time.Time
	BeforeCaseID  string
	Limit         int
}
type WorkflowEvaluationListPage struct {
	Cases   []WorkflowEvaluationCase
	HasMore bool
}

var errWorkflowEvaluationStoreContract = errors.New("workflow evaluation store contract mismatch")

type workflowEvaluationStore interface {
	CreateCase(WorkflowRunContext, WorkflowEvaluationCase) error
	ReadCase(WorkflowRunContext, string) (WorkflowEvaluationCase, bool, error)
	ListCases(WorkflowRunContext, WorkflowEvaluationListFilter) (WorkflowEvaluationListPage, error)
}

type memoryWorkflowEvaluationStore struct {
	mu       sync.RWMutex
	capacity int
	records  map[string]WorkflowEvaluationCase
	order    []string
}

func newMemoryWorkflowEvaluationStore(capacity int) *memoryWorkflowEvaluationStore {
	if capacity <= 0 {
		capacity = defaultWorkflowEvaluationCapacity
	}
	return &memoryWorkflowEvaluationStore{capacity: capacity, records: make(map[string]WorkflowEvaluationCase, capacity)}
}

func (s *memoryWorkflowEvaluationStore) CreateCase(ctx WorkflowRunContext, value WorkflowEvaluationCase) error {
	if validateWorkflowEvaluationCase(ctx, value) != nil {
		return errWorkflowEvaluationStoreContract
	}
	key := workflowEvaluationKey(ctx, value.CaseID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[key]; exists {
		return errWorkflowEvaluationStoreContract
	}
	s.records[key] = cloneWorkflowEvaluationCase(value)
	s.order = append(s.order, key)
	for len(s.order) > s.capacity {
		delete(s.records, s.order[0])
		s.order = s.order[1:]
	}
	return nil
}
func (s *memoryWorkflowEvaluationStore) ReadCase(ctx WorkflowRunContext, id string) (WorkflowEvaluationCase, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.records[workflowEvaluationKey(ctx, id)]
	return cloneWorkflowEvaluationCase(value), ok, nil
}
func (s *memoryWorkflowEvaluationStore) ListCases(ctx WorkflowRunContext, filter WorkflowEvaluationListFilter) (WorkflowEvaluationListPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prefix := workflowEvaluationKey(ctx, "")
	values := make([]WorkflowEvaluationCase, 0)
	for key, value := range s.records {
		if strings.HasPrefix(key, prefix) && (filter.BaselineRunID == "" || value.BaselineRunID == filter.BaselineRunID) {
			created, _ := time.Parse(time.RFC3339Nano, value.CreatedAt)
			if filter.BeforeTime != nil && (created.After(*filter.BeforeTime) || (created.Equal(*filter.BeforeTime) && value.CaseID >= filter.BeforeCaseID)) {
				continue
			}
			values = append(values, cloneWorkflowEvaluationCase(value))
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].CreatedAt == values[j].CreatedAt {
			return values[i].CaseID > values[j].CaseID
		}
		return values[i].CreatedAt > values[j].CreatedAt
	})
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return WorkflowEvaluationListPage{Cases: values, HasMore: hasMore}, nil
}

type WorkflowEvaluationCreateRequest struct {
	Name          string
	BaselineRunID string
	Expectations  []WorkflowEvaluationExpectation
}
type WorkflowEvaluationListRequest struct {
	Limit         int
	Cursor        string
	BaselineRunID string
}
type WorkflowEvaluationResult struct {
	Case           *WorkflowEvaluationCase
	FailureCode    WorkflowEvaluationFailureCode
	FailureSummary string
}
type WorkflowEvaluationListResult struct {
	Cases          []WorkflowEvaluationCase
	NextCursor     string
	HasMore        bool
	FailureCode    WorkflowEvaluationFailureCode
	FailureSummary string
}

type WorkflowEvaluationReviewItem struct {
	CandidateRunID          string                              `json:"candidate_run_id"`
	ExpectedClassification  WorkflowRunComparisonClassification `json:"expected_classification"`
	ActualClassification    WorkflowRunComparisonClassification `json:"actual_classification"`
	ComparisonState         WorkflowRunComparisonState          `json:"comparison_state"`
	Outcome                 string                              `json:"outcome"`
	FindingCodes            []string                            `json:"finding_codes"`
	RecommendedReviewAction WorkflowRunReviewAction             `json:"recommended_review_action"`
}
type WorkflowEvaluationReview struct {
	CaseID       string                         `json:"case_id"`
	Outcome      string                         `json:"outcome"`
	Matched      int                            `json:"matched"`
	Mismatched   int                            `json:"mismatched"`
	Inconclusive int                            `json:"inconclusive"`
	Unavailable  int                            `json:"unavailable"`
	Items        []WorkflowEvaluationReviewItem `json:"items"`
}
type WorkflowEvaluationReviewResult struct {
	Review         *WorkflowEvaluationReview
	FailureCode    WorkflowEvaluationFailureCode
	FailureSummary string
}

type workflowEvaluationService struct {
	store     workflowEvaluationStore
	runStore  workflowRunStore
	newCaseID func() (string, error)
	now       func() time.Time
}

func newWorkflowEvaluationService(store workflowEvaluationStore, runStore workflowRunStore) workflowEvaluationService {
	return workflowEvaluationService{store: store, runStore: runStore, newCaseID: newWorkflowEvaluationCaseID, now: func() time.Time { return time.Now().UTC() }}
}

func (s workflowEvaluationService) Create(ctx WorkflowRunContext, request WorkflowEvaluationCreateRequest) WorkflowEvaluationResult {
	name := strings.TrimSpace(request.Name)
	baseline := strings.TrimSpace(request.BaselineRunID)
	if !validWorkflowEvaluationName(name) || baseline == "" || len(request.Expectations) < 1 || len(request.Expectations) > workflowEvaluationMaxCandidates {
		return workflowEvaluationFailure(WorkflowEvaluationFailureInvalid)
	}
	seen := map[string]bool{baseline: true}
	normalized := make([]WorkflowEvaluationExpectation, 0, len(request.Expectations))
	for _, item := range request.Expectations {
		id := strings.TrimSpace(item.CandidateRunID)
		if id == "" || seen[id] || !validExpectedClassification(item.ExpectedClassification) {
			return workflowEvaluationFailure(WorkflowEvaluationFailureInvalid)
		}
		seen[id] = true
		normalized = append(normalized, WorkflowEvaluationExpectation{CandidateRunID: id, ExpectedClassification: item.ExpectedClassification})
	}
	for id := range seen {
		record, found, err := s.runStore.ReadRun(ctx, id)
		if err != nil {
			return workflowEvaluationFailure(mapWorkflowEvaluationStoreError(err))
		}
		if !found {
			return workflowEvaluationFailure(WorkflowEvaluationFailureNotFound)
		}
		if !workflowEvaluationRunEligible(record, s.now()) {
			return workflowEvaluationFailure(WorkflowEvaluationFailureRunNotEligible)
		}
	}
	id, err := s.newCaseID()
	if err != nil {
		return workflowEvaluationFailure(WorkflowEvaluationFailureStoreUnavailable)
	}
	value := WorkflowEvaluationCase{SchemaVersion: workflowEvaluationCaseSchema, CaseID: id, Name: name, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, BaselineRunID: baseline, Expectations: normalized, CreatedAt: workflowRunTimestamp(s.now()), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	if err = s.store.CreateCase(ctx, value); err != nil {
		return workflowEvaluationFailure(mapWorkflowEvaluationStoreError(err))
	}
	return WorkflowEvaluationResult{Case: workflowEvaluationPointer(value)}
}

func (s workflowEvaluationService) Read(ctx WorkflowRunContext, id string) WorkflowEvaluationResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return workflowEvaluationFailure(WorkflowEvaluationFailureInvalid)
	}
	value, found, err := s.store.ReadCase(ctx, id)
	if err != nil {
		return workflowEvaluationFailure(mapWorkflowEvaluationStoreError(err))
	}
	if !found {
		return workflowEvaluationFailure(WorkflowEvaluationFailureNotFound)
	}
	return WorkflowEvaluationResult{Case: workflowEvaluationPointer(value)}
}

func (s workflowEvaluationService) List(ctx WorkflowRunContext, request WorkflowEvaluationListRequest) WorkflowEvaluationListResult {
	limit := request.Limit
	if limit == 0 {
		limit = workflowEvaluationDefaultLimit
	}
	baseline := strings.TrimSpace(request.BaselineRunID)
	if limit < 1 || limit > workflowEvaluationMaxLimit || len([]rune(baseline)) > 160 {
		return workflowEvaluationListFailure(WorkflowEvaluationFailureInvalid)
	}
	filter := WorkflowEvaluationListFilter{BaselineRunID: baseline, Limit: limit}
	if request.Cursor != "" {
		cursor, err := decodeWorkflowEvaluationCursor(request.Cursor, baseline, limit)
		if err != nil {
			return workflowEvaluationListFailure(WorkflowEvaluationFailureCursorInvalid)
		}
		filter.BeforeTime = &cursor.Time
		filter.BeforeCaseID = cursor.CaseID
	}
	page, err := s.store.ListCases(ctx, filter)
	if err != nil {
		return workflowEvaluationListFailure(mapWorkflowEvaluationStoreError(err))
	}
	next := ""
	if page.HasMore && len(page.Cases) > 0 {
		next, _ = encodeWorkflowEvaluationCursor(page.Cases[len(page.Cases)-1], baseline, limit)
	}
	return WorkflowEvaluationListResult{Cases: page.Cases, NextCursor: next, HasMore: page.HasMore}
}

func (s workflowEvaluationService) Review(ctx WorkflowRunContext, id string) WorkflowEvaluationReviewResult {
	read := s.Read(ctx, id)
	if read.FailureCode != "" {
		return WorkflowEvaluationReviewResult{FailureCode: read.FailureCode, FailureSummary: read.FailureSummary}
	}
	review := WorkflowEvaluationReview{CaseID: read.Case.CaseID, Items: make([]WorkflowEvaluationReviewItem, 0, len(read.Case.Expectations))}
	comparisonService := newWorkflowExecutorService(nil, nil, s.runStore)
	for _, expected := range read.Case.Expectations {
		result := comparisonService.CompareRuns(ctx, read.Case.BaselineRunID, expected.CandidateRunID)
		item := WorkflowEvaluationReviewItem{CandidateRunID: expected.CandidateRunID, ExpectedClassification: expected.ExpectedClassification, FindingCodes: []string{}}
		if result.FailureCode == WorkflowRunFailureRecordNotFound {
			item.Outcome = "unavailable"
			review.Unavailable++
			review.Items = append(review.Items, item)
			continue
		}
		if result.FailureCode != "" {
			code := WorkflowEvaluationFailureStoreUnavailable
			if result.FailureCode == WorkflowRunFailureStoreContractMismatch {
				code = WorkflowEvaluationFailureStoreContract
			}
			return WorkflowEvaluationReviewResult{FailureCode: code, FailureSummary: "Workflow evaluation review storage is unavailable."}
		}
		item.ActualClassification = result.Comparison.Classification
		item.ComparisonState = result.Comparison.ComparisonState
		item.RecommendedReviewAction = result.Comparison.RecommendedReviewAction
		for _, finding := range result.Comparison.Findings {
			item.FindingCodes = append(item.FindingCodes, finding.Code)
		}
		if item.ActualClassification == WorkflowRunComparisonInconclusive {
			item.Outcome = "inconclusive"
			review.Inconclusive++
		} else if item.ActualClassification == item.ExpectedClassification {
			item.Outcome = "matched"
			review.Matched++
		} else {
			item.Outcome = "mismatched"
			review.Mismatched++
		}
		review.Items = append(review.Items, item)
	}
	review.Outcome = "passed"
	if review.Mismatched > 0 {
		review.Outcome = "mismatch"
	} else if review.Inconclusive+review.Unavailable > 0 {
		review.Outcome = "inconclusive"
	}
	return WorkflowEvaluationReviewResult{Review: &review}
}

type workflowEvaluationCursor struct {
	Version   int    `json:"v"`
	CreatedAt string `json:"created_at"`
	CaseID    string `json:"case_id"`
	Digest    string `json:"digest"`
}
type decodedWorkflowEvaluationCursor struct {
	Time   time.Time
	CaseID string
}

func encodeWorkflowEvaluationCursor(value WorkflowEvaluationCase, baseline string, limit int) (string, error) {
	doc, err := json.Marshal(workflowEvaluationCursor{Version: 1, CreatedAt: value.CreatedAt, CaseID: value.CaseID, Digest: workflowEvaluationCursorDigest(baseline, limit)})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(doc), nil
}
func decodeWorkflowEvaluationCursor(value, baseline string, limit int) (decodedWorkflowEvaluationCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(raw) > 1024 {
		return decodedWorkflowEvaluationCursor{}, errors.New("invalid cursor")
	}
	var doc workflowEvaluationCursor
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&doc) != nil || doc.Version != 1 || doc.CaseID == "" || doc.Digest != workflowEvaluationCursorDigest(baseline, limit) {
		return decodedWorkflowEvaluationCursor{}, errors.New("invalid cursor")
	}
	parsed, err := time.Parse(time.RFC3339Nano, doc.CreatedAt)
	return decodedWorkflowEvaluationCursor{Time: parsed, CaseID: doc.CaseID}, err
}
func workflowEvaluationCursorDigest(baseline string, limit int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("v1\n%s\n%d", baseline, limit)))
	return hex.EncodeToString(sum[:16])
}

func workflowEvaluationRunEligible(record WorkflowRunRecord, now time.Time) bool {
	if !workflowRunComparisonSideEffectsValid(record.SideEffects) {
		return false
	}
	if isTerminalWorkflowRunStatus(record.Status) {
		return true
	}
	started, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	return err == nil && record.Status == WorkflowRunStatusRunning && now.Sub(started) > workflowExecutorDefaultMaxRuntime
}
func validExpectedClassification(v WorkflowRunComparisonClassification) bool {
	return v == WorkflowRunComparisonRegression || v == WorkflowRunComparisonImprovement || v == WorkflowRunComparisonChanged || v == WorkflowRunComparisonUnchanged
}
func validWorkflowEvaluationName(v string) bool {
	lower := strings.ToLower(v)
	return v != "" && len([]rune(v)) <= 96 && !strings.ContainsAny(v, "\r\n") && !strings.Contains(v, "://") && !strings.Contains(lower, "api_key") && !strings.Contains(lower, "token=") && !strings.Contains(lower, "secret=")
}
func validateWorkflowEvaluationCase(ctx WorkflowRunContext, v WorkflowEvaluationCase) error {
	if v.SchemaVersion != workflowEvaluationCaseSchema || v.CaseID == "" || !validWorkflowEvaluationName(v.Name) || v.WorkspaceID != ctx.WorkspaceID || v.ApplicationID != ctx.ApplicationID || v.BaselineRunID == "" || len(v.Expectations) < 1 || len(v.Expectations) > workflowEvaluationMaxCandidates || v.ActorRef == "" || v.RequestID == "" || v.AuditRef == "" {
		return errWorkflowEvaluationStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, v.CreatedAt); err != nil {
		return errWorkflowEvaluationStoreContract
	}
	seen := map[string]bool{v.BaselineRunID: true}
	for _, item := range v.Expectations {
		if item.CandidateRunID == "" || seen[item.CandidateRunID] || !validExpectedClassification(item.ExpectedClassification) {
			return errWorkflowEvaluationStoreContract
		}
		seen[item.CandidateRunID] = true
	}
	return nil
}
func newWorkflowEvaluationCaseID() (string, error) {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "eval_" + hex.EncodeToString(raw), nil
}
func workflowEvaluationKey(ctx WorkflowRunContext, id string) string {
	return ctx.TenantRef + "\x00" + ctx.WorkspaceID + "\x00" + ctx.ApplicationID + "\x00" + id
}
func cloneWorkflowEvaluationCase(v WorkflowEvaluationCase) WorkflowEvaluationCase {
	v.Expectations = append([]WorkflowEvaluationExpectation(nil), v.Expectations...)
	return v
}
func workflowEvaluationPointer(v WorkflowEvaluationCase) *WorkflowEvaluationCase {
	copy := cloneWorkflowEvaluationCase(v)
	return &copy
}
func workflowEvaluationFailure(code WorkflowEvaluationFailureCode) WorkflowEvaluationResult {
	summary := "Workflow evaluation request is invalid."
	if code == WorkflowEvaluationFailureNotFound {
		summary = "Workflow evaluation record was not found in the current scope."
	} else if code == WorkflowEvaluationFailureRunNotEligible {
		summary = "Workflow evaluation requires terminal or stale zero-side-effect runs."
	} else if code == WorkflowEvaluationFailureStoreUnavailable || code == WorkflowEvaluationFailureStoreContract {
		summary = "Workflow evaluation storage is unavailable."
	}
	return WorkflowEvaluationResult{FailureCode: code, FailureSummary: summary}
}
func workflowEvaluationListFailure(code WorkflowEvaluationFailureCode) WorkflowEvaluationListResult {
	value := workflowEvaluationFailure(code)
	return WorkflowEvaluationListResult{FailureCode: code, FailureSummary: value.FailureSummary}
}
func mapWorkflowEvaluationStoreError(err error) WorkflowEvaluationFailureCode {
	if errors.Is(err, errWorkflowEvaluationStoreContract) || errors.Is(err, errWorkflowRunStoreContract) {
		return WorkflowEvaluationFailureStoreContract
	}
	return WorkflowEvaluationFailureStoreUnavailable
}

var _ workflowEvaluationStore = (*memoryWorkflowEvaluationStore)(nil)
