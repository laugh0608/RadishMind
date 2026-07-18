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
	workflowEvaluationSuiteSchema          = "workflow_evaluation_suite.v1"
	workflowEvaluationDecisionSchema       = "workflow_evaluation_release_decision.v1"
	workflowEvaluationSuiteMaxCases        = 50
	defaultWorkflowEvaluationSuiteCapacity = 100
)

type WorkflowEvaluationSuiteFailureCode string

const (
	WorkflowEvaluationSuiteFailureInvalid               WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_invalid"
	WorkflowEvaluationSuiteFailureNotFound              WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_not_found"
	WorkflowEvaluationSuiteFailureCaseNotEligible       WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_case_not_eligible"
	WorkflowEvaluationSuiteFailureCursor                WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_cursor_invalid"
	WorkflowEvaluationSuiteFailureReviewChanged         WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_review_changed"
	WorkflowEvaluationSuiteFailureDecisionConflict      WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_decision_conflict"
	WorkflowEvaluationSuiteFailureApprovalBlocked       WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_approval_blocked"
	WorkflowEvaluationSuiteFailureStoreUnavailable      WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_store_unavailable"
	WorkflowEvaluationSuiteFailureStoreContract         WorkflowEvaluationSuiteFailureCode = "workflow_evaluation_suite_store_contract_mismatch"
	WorkflowEvaluationSuiteFailureRetrievalProfile      WorkflowEvaluationSuiteFailureCode = "workflow_run_retrieval_profile_unsupported"
	WorkflowEvaluationSuiteFailureRetrievalIncompatible WorkflowEvaluationSuiteFailureCode = "workflow_run_retrieval_profile_incompatible"
)

type WorkflowEvaluationSuiteCaseRef struct {
	CaseID  string `json:"case_id"`
	Version int    `json:"version"`
}
type WorkflowEvaluationSuite struct {
	SchemaVersion          string                           `json:"schema_version"`
	SuiteID                string                           `json:"suite_id"`
	Name                   string                           `json:"name"`
	WorkspaceID            string                           `json:"workspace_id"`
	ApplicationID          string                           `json:"application_id"`
	CaseRefs               []WorkflowEvaluationSuiteCaseRef `json:"case_refs"`
	CurrentDecisionVersion int                              `json:"current_decision_version"`
	CurrentDecision        string                           `json:"current_decision"`
	CreatedAt              string                           `json:"created_at"`
	ActorRef               string                           `json:"actor_ref"`
	RequestID              string                           `json:"request_id"`
	AuditRef               string                           `json:"audit_ref"`
}
type WorkflowEvaluationReleaseDecision struct {
	SchemaVersion string `json:"schema_version"`
	DecisionID    string `json:"decision_id"`
	SuiteID       string `json:"suite_id"`
	Version       int    `json:"version"`
	Decision      string `json:"decision"`
	ReviewDigest  string `json:"review_digest"`
	ReviewOutcome string `json:"review_outcome"`
	Passed        int    `json:"passed"`
	Mismatch      int    `json:"mismatch"`
	Inconclusive  int    `json:"inconclusive"`
	Unavailable   int    `json:"unavailable"`
	CreatedAt     string `json:"created_at"`
	ActorRef      string `json:"actor_ref"`
	RequestID     string `json:"request_id"`
	AuditRef      string `json:"audit_ref"`
}
type WorkflowEvaluationSuiteReviewItem struct {
	CaseID       string `json:"case_id"`
	Version      int    `json:"version"`
	Name         string `json:"name"`
	Outcome      string `json:"outcome"`
	Matched      int    `json:"matched"`
	Mismatched   int    `json:"mismatched"`
	Inconclusive int    `json:"inconclusive"`
	Unavailable  int    `json:"unavailable"`
	AuditRef     string `json:"audit_ref"`
	RunProfile   string `json:"run_profile"`
}
type WorkflowEvaluationSuiteReview struct {
	SuiteID      string                              `json:"suite_id"`
	Outcome      string                              `json:"outcome"`
	ReviewDigest string                              `json:"review_digest"`
	Passed       int                                 `json:"passed"`
	Mismatch     int                                 `json:"mismatch"`
	Inconclusive int                                 `json:"inconclusive"`
	Unavailable  int                                 `json:"unavailable"`
	Items        []WorkflowEvaluationSuiteReviewItem `json:"items"`
}

type workflowEvaluationSuiteListFilter struct {
	BeforeTime    *time.Time
	BeforeSuiteID string
	Limit         int
}
type workflowEvaluationSuiteListPage struct {
	Suites  []WorkflowEvaluationSuite
	HasMore bool
}
type workflowEvaluationDecisionListFilter struct {
	BeforeVersion int
	Limit         int
}
type workflowEvaluationDecisionListPage struct {
	Decisions []WorkflowEvaluationReleaseDecision
	HasMore   bool
}

var errWorkflowEvaluationSuiteStoreContract = errors.New("workflow evaluation suite store contract mismatch")

type workflowEvaluationSuiteStore interface {
	CreateSuite(WorkflowRunContext, WorkflowEvaluationSuite) error
	ReadSuite(WorkflowRunContext, string) (WorkflowEvaluationSuite, bool, error)
	ListSuites(WorkflowRunContext, workflowEvaluationSuiteListFilter) (workflowEvaluationSuiteListPage, error)
	AppendDecision(WorkflowRunContext, int, WorkflowEvaluationReleaseDecision) (WorkflowEvaluationSuite, bool, error)
	ListDecisions(WorkflowRunContext, string, workflowEvaluationDecisionListFilter) (workflowEvaluationDecisionListPage, error)
}

type memoryWorkflowEvaluationSuiteStore struct {
	mu        sync.RWMutex
	capacity  int
	suites    map[string]WorkflowEvaluationSuite
	order     []string
	decisions map[string][]WorkflowEvaluationReleaseDecision
}

func newMemoryWorkflowEvaluationSuiteStore(capacity int) *memoryWorkflowEvaluationSuiteStore {
	if capacity <= 0 {
		capacity = defaultWorkflowEvaluationSuiteCapacity
	}
	return &memoryWorkflowEvaluationSuiteStore{capacity: capacity, suites: map[string]WorkflowEvaluationSuite{}, decisions: map[string][]WorkflowEvaluationReleaseDecision{}}
}
func (s *memoryWorkflowEvaluationSuiteStore) CreateSuite(ctx WorkflowRunContext, value WorkflowEvaluationSuite) error {
	if validateWorkflowEvaluationSuite(ctx, value) != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	key := workflowEvaluationKey(ctx, value.SuiteID)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suites[key]; ok {
		return errWorkflowEvaluationSuiteStoreContract
	}
	s.suites[key] = cloneWorkflowEvaluationSuite(value)
	s.decisions[key] = []WorkflowEvaluationReleaseDecision{}
	s.order = append(s.order, key)
	for len(s.order) > s.capacity {
		delete(s.suites, s.order[0])
		delete(s.decisions, s.order[0])
		s.order = s.order[1:]
	}
	return nil
}
func (s *memoryWorkflowEvaluationSuiteStore) ReadSuite(ctx WorkflowRunContext, id string) (WorkflowEvaluationSuite, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.suites[workflowEvaluationKey(ctx, id)]
	return cloneWorkflowEvaluationSuite(v), ok, nil
}
func (s *memoryWorkflowEvaluationSuiteStore) ListSuites(ctx WorkflowRunContext, f workflowEvaluationSuiteListFilter) (workflowEvaluationSuiteListPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	prefix := workflowEvaluationKey(ctx, "")
	values := []WorkflowEvaluationSuite{}
	for key, v := range s.suites {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		created, _ := time.Parse(time.RFC3339Nano, v.CreatedAt)
		if f.BeforeTime != nil && (created.After(*f.BeforeTime) || (created.Equal(*f.BeforeTime) && v.SuiteID >= f.BeforeSuiteID)) {
			continue
		}
		values = append(values, cloneWorkflowEvaluationSuite(v))
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].CreatedAt == values[j].CreatedAt {
			return values[i].SuiteID > values[j].SuiteID
		}
		return values[i].CreatedAt > values[j].CreatedAt
	})
	more := len(values) > f.Limit
	if more {
		values = values[:f.Limit]
	}
	return workflowEvaluationSuiteListPage{Suites: values, HasMore: more}, nil
}
func (s *memoryWorkflowEvaluationSuiteStore) AppendDecision(ctx WorkflowRunContext, expected int, d WorkflowEvaluationReleaseDecision) (WorkflowEvaluationSuite, bool, error) {
	if validateWorkflowEvaluationDecision(ctx, d) != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	key := workflowEvaluationKey(ctx, d.SuiteID)
	s.mu.Lock()
	defer s.mu.Unlock()
	suite, ok := s.suites[key]
	if !ok {
		return WorkflowEvaluationSuite{}, false, nil
	}
	if suite.CurrentDecisionVersion != expected {
		return cloneWorkflowEvaluationSuite(suite), false, nil
	}
	if d.Version != expected+1 {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	s.decisions[key] = append(s.decisions[key], d)
	suite.CurrentDecisionVersion = d.Version
	suite.CurrentDecision = d.Decision
	s.suites[key] = suite
	return cloneWorkflowEvaluationSuite(suite), true, nil
}
func (s *memoryWorkflowEvaluationSuiteStore) ListDecisions(ctx WorkflowRunContext, id string, f workflowEvaluationDecisionListFilter) (workflowEvaluationDecisionListPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := s.decisions[workflowEvaluationKey(ctx, id)]
	values := []WorkflowEvaluationReleaseDecision{}
	for i := len(all) - 1; i >= 0; i-- {
		if f.BeforeVersion == 0 || all[i].Version < f.BeforeVersion {
			values = append(values, all[i])
		}
	}
	more := len(values) > f.Limit
	if more {
		values = values[:f.Limit]
	}
	return workflowEvaluationDecisionListPage{Decisions: values, HasMore: more}, nil
}

type WorkflowEvaluationSuiteCreateRequest struct {
	Name     string
	CaseRefs []WorkflowEvaluationSuiteCaseRef
}
type WorkflowEvaluationSuiteListRequest struct {
	Limit  int
	Cursor string
}
type WorkflowEvaluationDecisionRequest struct {
	ExpectedDecisionVersion int
	Decision                string
	ReviewDigest            string
}
type WorkflowEvaluationDecisionListRequest struct {
	Limit  int
	Cursor string
}
type WorkflowEvaluationSuiteResult struct {
	Suite          *WorkflowEvaluationSuite
	Decision       *WorkflowEvaluationReleaseDecision
	FailureCode    WorkflowEvaluationSuiteFailureCode
	FailureSummary string
}
type WorkflowEvaluationSuiteListResult struct {
	Suites         []WorkflowEvaluationSuite
	NextCursor     string
	HasMore        bool
	FailureCode    WorkflowEvaluationSuiteFailureCode
	FailureSummary string
}
type WorkflowEvaluationDecisionListResult struct {
	Decisions      []WorkflowEvaluationReleaseDecision
	NextCursor     string
	HasMore        bool
	FailureCode    WorkflowEvaluationSuiteFailureCode
	FailureSummary string
}
type WorkflowEvaluationSuiteReviewResult struct {
	Review         *WorkflowEvaluationSuiteReview
	FailureCode    WorkflowEvaluationSuiteFailureCode
	FailureSummary string
}
type workflowEvaluationSuiteService struct {
	store         workflowEvaluationSuiteStore
	evaluation    workflowEvaluationService
	newSuiteID    func() (string, error)
	newDecisionID func() (string, error)
	now           func() time.Time
}

func newWorkflowEvaluationSuiteService(store workflowEvaluationSuiteStore, evaluation workflowEvaluationService) workflowEvaluationSuiteService {
	return workflowEvaluationSuiteService{store: store, evaluation: evaluation, newSuiteID: func() (string, error) { return newEvaluationID("suite_") }, newDecisionID: func() (string, error) { return newEvaluationID("decision_") }, now: func() time.Time { return time.Now().UTC() }}
}

func (s workflowEvaluationSuiteService) Create(ctx WorkflowRunContext, r WorkflowEvaluationSuiteCreateRequest) WorkflowEvaluationSuiteResult {
	name := strings.TrimSpace(r.Name)
	if !validWorkflowEvaluationName(name) || len(r.CaseRefs) < 1 || len(r.CaseRefs) > workflowEvaluationSuiteMaxCases {
		return suiteFailure(WorkflowEvaluationSuiteFailureInvalid)
	}
	refs := append([]WorkflowEvaluationSuiteCaseRef(nil), r.CaseRefs...)
	seen := map[string]bool{}
	for i := range refs {
		refs[i].CaseID = strings.TrimSpace(refs[i].CaseID)
		key := fmt.Sprintf("%s\x00%d", refs[i].CaseID, refs[i].Version)
		if refs[i].CaseID == "" || refs[i].Version < 1 || seen[key] {
			return suiteFailure(WorkflowEvaluationSuiteFailureInvalid)
		}
		seen[key] = true
		read := s.evaluation.ReadRevision(ctx, refs[i].CaseID, refs[i].Version)
		if read.FailureCode != "" {
			if read.FailureCode == WorkflowEvaluationFailureNotFound {
				return suiteFailure(WorkflowEvaluationSuiteFailureCaseNotEligible)
			}
			return suiteFailure(mapSuiteEvaluationFailure(read.FailureCode))
		}
	}
	id, err := s.newSuiteID()
	if err != nil {
		return suiteFailure(WorkflowEvaluationSuiteFailureStoreUnavailable)
	}
	value := WorkflowEvaluationSuite{SchemaVersion: workflowEvaluationSuiteSchema, SuiteID: id, Name: name, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, CaseRefs: refs, CreatedAt: workflowRunTimestamp(s.now()), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	if err = s.store.CreateSuite(ctx, value); err != nil {
		return suiteFailure(mapSuiteStoreError(err))
	}
	return WorkflowEvaluationSuiteResult{Suite: suitePointer(value)}
}
func (s workflowEvaluationSuiteService) Read(ctx WorkflowRunContext, id string) WorkflowEvaluationSuiteResult {
	id = strings.TrimSpace(id)
	if id == "" {
		return suiteFailure(WorkflowEvaluationSuiteFailureInvalid)
	}
	v, ok, err := s.store.ReadSuite(ctx, id)
	if err != nil {
		return suiteFailure(mapSuiteStoreError(err))
	}
	if !ok {
		return suiteFailure(WorkflowEvaluationSuiteFailureNotFound)
	}
	return WorkflowEvaluationSuiteResult{Suite: suitePointer(v)}
}
func (s workflowEvaluationSuiteService) Review(ctx WorkflowRunContext, id string) WorkflowEvaluationSuiteReviewResult {
	read := s.Read(ctx, id)
	if read.FailureCode != "" {
		return suiteReviewFailure(read.FailureCode)
	}
	review := WorkflowEvaluationSuiteReview{SuiteID: read.Suite.SuiteID, Items: []WorkflowEvaluationSuiteReviewItem{}}
	for _, ref := range read.Suite.CaseRefs {
		caseRead := s.evaluation.ReadRevision(ctx, ref.CaseID, ref.Version)
		item := WorkflowEvaluationSuiteReviewItem{CaseID: ref.CaseID, Version: ref.Version}
		if caseRead.FailureCode == WorkflowEvaluationFailureNotFound {
			item.Outcome = "unavailable"
			item.Unavailable = 1
			review.Unavailable++
			review.Items = append(review.Items, item)
			continue
		}
		if caseRead.FailureCode != "" {
			return suiteReviewFailure(mapSuiteEvaluationFailure(caseRead.FailureCode))
		}
		item.Name = caseRead.Case.Name
		item.AuditRef = caseRead.Case.AuditRef
		result := s.evaluation.ReviewVersion(ctx, ref.CaseID, ref.Version)
		if result.FailureCode == WorkflowEvaluationFailureNotFound {
			item.Outcome = "unavailable"
			item.Unavailable = 1
			review.Unavailable++
			review.Items = append(review.Items, item)
			continue
		}
		if result.FailureCode != "" {
			return suiteReviewFailure(mapSuiteEvaluationFailure(result.FailureCode))
		}
		item.Outcome = result.Review.Outcome
		item.RunProfile = result.Review.RunProfile
		item.Matched = result.Review.Matched
		item.Mismatched = result.Review.Mismatched
		item.Inconclusive = result.Review.Inconclusive
		item.Unavailable = result.Review.Unavailable
		switch item.Outcome {
		case "passed":
			review.Passed++
		case "mismatch":
			review.Mismatch++
		default:
			review.Inconclusive++
		}
		review.Items = append(review.Items, item)
	}
	review.Outcome = "passed"
	if review.Mismatch > 0 {
		review.Outcome = "mismatch"
	} else if review.Inconclusive+review.Unavailable > 0 {
		review.Outcome = "inconclusive"
	}
	review.ReviewDigest = workflowEvaluationSuiteReviewDigest(review, read.Suite.CaseRefs)
	return WorkflowEvaluationSuiteReviewResult{Review: &review}
}
func (s workflowEvaluationSuiteService) Decide(ctx WorkflowRunContext, id string, r WorkflowEvaluationDecisionRequest) WorkflowEvaluationSuiteResult {
	if r.ExpectedDecisionVersion < 0 || !validReleaseDecision(r.Decision) || len(r.ReviewDigest) != 64 {
		return suiteFailure(WorkflowEvaluationSuiteFailureInvalid)
	}
	review := s.Review(ctx, id)
	if review.FailureCode != "" {
		return suiteFailure(review.FailureCode)
	}
	if !strings.EqualFold(strings.TrimSpace(r.ReviewDigest), review.Review.ReviewDigest) {
		return suiteFailure(WorkflowEvaluationSuiteFailureReviewChanged)
	}
	if r.Decision == "approved" && (review.Review.Outcome != "passed" || review.Review.Unavailable > 0) {
		return suiteFailure(WorkflowEvaluationSuiteFailureApprovalBlocked)
	}
	decisionID, err := s.newDecisionID()
	if err != nil {
		return suiteFailure(WorkflowEvaluationSuiteFailureStoreUnavailable)
	}
	d := WorkflowEvaluationReleaseDecision{SchemaVersion: workflowEvaluationDecisionSchema, DecisionID: decisionID, SuiteID: id, Version: r.ExpectedDecisionVersion + 1, Decision: r.Decision, ReviewDigest: review.Review.ReviewDigest, ReviewOutcome: review.Review.Outcome, Passed: review.Review.Passed, Mismatch: review.Review.Mismatch, Inconclusive: review.Review.Inconclusive, Unavailable: review.Review.Unavailable, CreatedAt: workflowRunTimestamp(s.now()), ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	suite, applied, err := s.store.AppendDecision(ctx, r.ExpectedDecisionVersion, d)
	if err != nil {
		return suiteFailure(mapSuiteStoreError(err))
	}
	if !applied {
		if suite.SuiteID == "" {
			return suiteFailure(WorkflowEvaluationSuiteFailureNotFound)
		}
		result := suiteFailure(WorkflowEvaluationSuiteFailureDecisionConflict)
		result.Suite = suitePointer(suite)
		return result
	}
	return WorkflowEvaluationSuiteResult{Suite: suitePointer(suite), Decision: &d}
}

func (s workflowEvaluationSuiteService) List(ctx WorkflowRunContext, r WorkflowEvaluationSuiteListRequest) WorkflowEvaluationSuiteListResult {
	limit := r.Limit
	if limit == 0 {
		limit = workflowEvaluationDefaultLimit
	}
	if limit < 1 || limit > workflowEvaluationMaxLimit {
		return suiteListFailure(WorkflowEvaluationSuiteFailureInvalid)
	}
	f := workflowEvaluationSuiteListFilter{Limit: limit}
	if r.Cursor != "" {
		created, id, err := decodeSuiteCursor(r.Cursor, limit)
		if err != nil {
			return suiteListFailure(WorkflowEvaluationSuiteFailureCursor)
		}
		f.BeforeTime = &created
		f.BeforeSuiteID = id
	}
	page, err := s.store.ListSuites(ctx, f)
	if err != nil {
		return suiteListFailure(mapSuiteStoreError(err))
	}
	next := ""
	if page.HasMore {
		next, _ = encodeSuiteCursor(page.Suites[len(page.Suites)-1], limit)
	}
	return WorkflowEvaluationSuiteListResult{Suites: page.Suites, NextCursor: next, HasMore: page.HasMore}
}
func (s workflowEvaluationSuiteService) ListDecisions(ctx WorkflowRunContext, id string, r WorkflowEvaluationDecisionListRequest) WorkflowEvaluationDecisionListResult {
	limit := r.Limit
	if limit == 0 {
		limit = workflowEvaluationDefaultLimit
	}
	if strings.TrimSpace(id) == "" || limit < 1 || limit > workflowEvaluationMaxLimit {
		return decisionListFailure(WorkflowEvaluationSuiteFailureInvalid)
	}
	f := workflowEvaluationDecisionListFilter{Limit: limit}
	if r.Cursor != "" {
		v, err := decodeDecisionCursor(r.Cursor, id, limit)
		if err != nil {
			return decisionListFailure(WorkflowEvaluationSuiteFailureCursor)
		}
		f.BeforeVersion = v
	}
	page, err := s.store.ListDecisions(ctx, id, f)
	if err != nil {
		return decisionListFailure(mapSuiteStoreError(err))
	}
	if len(page.Decisions) == 0 && r.Cursor == "" {
		if read := s.Read(ctx, id); read.FailureCode != "" {
			return decisionListFailure(read.FailureCode)
		}
	}
	next := ""
	if page.HasMore {
		next, _ = encodeDecisionCursor(id, page.Decisions[len(page.Decisions)-1].Version, limit)
	}
	return WorkflowEvaluationDecisionListResult{Decisions: page.Decisions, NextCursor: next, HasMore: page.HasMore}
}

func workflowEvaluationSuiteReviewDigest(review WorkflowEvaluationSuiteReview, refs []WorkflowEvaluationSuiteCaseRef) string {
	doc := struct {
		Schema       string                              `json:"schema"`
		SuiteID      string                              `json:"suite_id"`
		Refs         []WorkflowEvaluationSuiteCaseRef    `json:"refs"`
		Outcome      string                              `json:"outcome"`
		Passed       int                                 `json:"passed"`
		Mismatch     int                                 `json:"mismatch"`
		Inconclusive int                                 `json:"inconclusive"`
		Unavailable  int                                 `json:"unavailable"`
		Items        []WorkflowEvaluationSuiteReviewItem `json:"items"`
	}{"workflow_evaluation_suite_review_digest.v1", review.SuiteID, refs, review.Outcome, review.Passed, review.Mismatch, review.Inconclusive, review.Unavailable, review.Items}
	raw, _ := json.Marshal(doc)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

type suiteCursor struct {
	Version   int    `json:"v"`
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Digest    string `json:"digest"`
}

func encodeSuiteCursor(v WorkflowEvaluationSuite, limit int) (string, error) {
	raw, err := json.Marshal(suiteCursor{1, v.CreatedAt, v.SuiteID, cursorDigest("suite", "", limit)})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func decodeSuiteCursor(value string, limit int) (time.Time, string, error) {
	var doc suiteCursor
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(raw) > 1024 {
		return time.Time{}, "", errors.New("invalid")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&doc) != nil || doc.Version != 1 || doc.ID == "" || doc.Digest != cursorDigest("suite", "", limit) {
		return time.Time{}, "", errors.New("invalid")
	}
	parsed, err := time.Parse(time.RFC3339Nano, doc.CreatedAt)
	return parsed, doc.ID, err
}

type decisionCursor struct {
	Version int    `json:"v"`
	Before  int    `json:"before"`
	Digest  string `json:"digest"`
}

func encodeDecisionCursor(id string, before, limit int) (string, error) {
	raw, err := json.Marshal(decisionCursor{1, before, cursorDigest("decision", id, limit)})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func decodeDecisionCursor(value, id string, limit int) (int, error) {
	var doc decisionCursor
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil || len(raw) > 1024 {
		return 0, errors.New("invalid")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&doc) != nil || doc.Version != 1 || doc.Before < 1 || doc.Digest != cursorDigest("decision", id, limit) {
		return 0, errors.New("invalid")
	}
	return doc.Before, nil
}
func cursorDigest(kind, id string, limit int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\n%s\n%d", kind, id, limit)))
	return hex.EncodeToString(sum[:16])
}
func newEvaluationID(prefix string) (string, error) {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(raw), nil
}
func validReleaseDecision(v string) bool {
	return v == "approved" || v == "rejected" || v == "needs_review"
}
func validateWorkflowEvaluationSuite(ctx WorkflowRunContext, v WorkflowEvaluationSuite) error {
	if v.SchemaVersion != workflowEvaluationSuiteSchema || v.SuiteID == "" || !validWorkflowEvaluationName(v.Name) || v.WorkspaceID != ctx.WorkspaceID || v.ApplicationID != ctx.ApplicationID || len(v.CaseRefs) < 1 || len(v.CaseRefs) > workflowEvaluationSuiteMaxCases || v.CurrentDecisionVersion < 0 || v.ActorRef == "" || v.RequestID == "" || v.AuditRef == "" || (v.CurrentDecisionVersion == 0) != (v.CurrentDecision == "") || (v.CurrentDecision != "" && !validReleaseDecision(v.CurrentDecision)) {
		return errWorkflowEvaluationSuiteStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, v.CreatedAt); err != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	seen := map[string]bool{}
	for _, ref := range v.CaseRefs {
		key := fmt.Sprintf("%s\x00%d", ref.CaseID, ref.Version)
		if ref.CaseID == "" || ref.Version < 1 || seen[key] {
			return errWorkflowEvaluationSuiteStoreContract
		}
		seen[key] = true
	}
	return nil
}
func validateWorkflowEvaluationDecision(ctx WorkflowRunContext, v WorkflowEvaluationReleaseDecision) error {
	if v.SchemaVersion != workflowEvaluationDecisionSchema || v.DecisionID == "" || v.SuiteID == "" || v.Version < 1 || !validReleaseDecision(v.Decision) || len(v.ReviewDigest) != 64 || v.ActorRef == "" || v.RequestID == "" || v.AuditRef == "" || (v.ReviewOutcome != "passed" && v.ReviewOutcome != "mismatch" && v.ReviewOutcome != "inconclusive") || v.Passed < 0 || v.Mismatch < 0 || v.Inconclusive < 0 || v.Unavailable < 0 {
		return errWorkflowEvaluationSuiteStoreContract
	}
	if _, err := hex.DecodeString(v.ReviewDigest); err != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	_, err := time.Parse(time.RFC3339Nano, v.CreatedAt)
	return err
}
func cloneWorkflowEvaluationSuite(v WorkflowEvaluationSuite) WorkflowEvaluationSuite {
	v.CaseRefs = append([]WorkflowEvaluationSuiteCaseRef(nil), v.CaseRefs...)
	return v
}
func suitePointer(v WorkflowEvaluationSuite) *WorkflowEvaluationSuite {
	copy := cloneWorkflowEvaluationSuite(v)
	return &copy
}
func suiteFailure(code WorkflowEvaluationSuiteFailureCode) WorkflowEvaluationSuiteResult {
	summary := "Workflow evaluation suite request is invalid."
	switch code {
	case WorkflowEvaluationSuiteFailureNotFound:
		summary = "Workflow evaluation suite was not found in the current scope."
	case WorkflowEvaluationSuiteFailureCaseNotEligible:
		summary = "Workflow evaluation suite requires readable exact case versions."
	case WorkflowEvaluationSuiteFailureReviewChanged:
		summary = "Workflow evaluation suite review changed; refresh before deciding."
	case WorkflowEvaluationSuiteFailureDecisionConflict:
		summary = "Workflow evaluation suite decision changed; refresh before deciding."
	case WorkflowEvaluationSuiteFailureApprovalBlocked:
		summary = "Workflow evaluation suite approval requires a fully passed review."
	case WorkflowEvaluationSuiteFailureStoreUnavailable, WorkflowEvaluationSuiteFailureStoreContract:
		summary = "Workflow evaluation suite storage is unavailable."
	case WorkflowEvaluationSuiteFailureRetrievalProfile:
		summary = "Workflow evaluation suite does not support workflow retrieval profiles."
	case WorkflowEvaluationSuiteFailureRetrievalIncompatible:
		summary = "Workflow evaluation suite contains incompatible retrieval bindings."
	}
	return WorkflowEvaluationSuiteResult{FailureCode: code, FailureSummary: summary}
}
func suiteListFailure(code WorkflowEvaluationSuiteFailureCode) WorkflowEvaluationSuiteListResult {
	v := suiteFailure(code)
	return WorkflowEvaluationSuiteListResult{FailureCode: code, FailureSummary: v.FailureSummary}
}
func decisionListFailure(code WorkflowEvaluationSuiteFailureCode) WorkflowEvaluationDecisionListResult {
	v := suiteFailure(code)
	return WorkflowEvaluationDecisionListResult{FailureCode: code, FailureSummary: v.FailureSummary}
}
func suiteReviewFailure(code WorkflowEvaluationSuiteFailureCode) WorkflowEvaluationSuiteReviewResult {
	v := suiteFailure(code)
	return WorkflowEvaluationSuiteReviewResult{FailureCode: code, FailureSummary: v.FailureSummary}
}
func mapSuiteStoreError(err error) WorkflowEvaluationSuiteFailureCode {
	if errors.Is(err, errWorkflowEvaluationSuiteStoreContract) {
		return WorkflowEvaluationSuiteFailureStoreContract
	}
	return WorkflowEvaluationSuiteFailureStoreUnavailable
}
func mapSuiteEvaluationFailure(code WorkflowEvaluationFailureCode) WorkflowEvaluationSuiteFailureCode {
	if code == WorkflowEvaluationFailureNotFound {
		return WorkflowEvaluationSuiteFailureNotFound
	}
	if code == WorkflowEvaluationFailureStoreContract {
		return WorkflowEvaluationSuiteFailureStoreContract
	}
	if code == WorkflowEvaluationFailureStoreUnavailable {
		return WorkflowEvaluationSuiteFailureStoreUnavailable
	}
	if code == WorkflowEvaluationFailureRetrievalProfile {
		return WorkflowEvaluationSuiteFailureRetrievalProfile
	}
	if code == WorkflowEvaluationFailureRetrievalIncompatible {
		return WorkflowEvaluationSuiteFailureRetrievalIncompatible
	}
	return WorkflowEvaluationSuiteFailureCaseNotEligible
}

var _ workflowEvaluationSuiteStore = (*memoryWorkflowEvaluationSuiteStore)(nil)
