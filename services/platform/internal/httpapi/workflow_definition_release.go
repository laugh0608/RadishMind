package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	workflowDefinitionCandidateSchemaVersion       = "workflow_definition_release_candidate.v1"
	workflowDefinitionDecisionSchemaVersion        = "workflow_definition_release_decision.v1"
	workflowDefinitionVersionSchemaVersion         = "workflow_definition_version.v1"
	workflowDefinitionActivationSchemaVersion      = "workflow_definition_activation.v1"
	workflowDefinitionActivationEventSchemaVersion = "workflow_definition_activation_event.v1"
	workflowDefinitionAuditSchemaVersion           = "workflow_definition_release_audit.v1"

	workflowDefinitionStatePending       = "pending"
	workflowDefinitionStateApproved      = "approved"
	workflowDefinitionStateRejected      = "rejected"
	workflowDefinitionActivationActive   = "active"
	workflowDefinitionActivationInactive = "inactive"
)

var (
	errWorkflowDefinitionNotFound       = errors.New("workflow_definition_not_found")
	errWorkflowDefinitionConflict       = errors.New("workflow_definition_version_conflict")
	errWorkflowDefinitionInvalidState   = errors.New("workflow_definition_transition_invalid")
	errWorkflowDefinitionPayloadInvalid = errors.New("workflow_definition_payload_invalid")
	errWorkflowDefinitionSourceDrift    = errors.New("workflow_definition_source_drift")
	errWorkflowDefinitionStore          = errors.New("workflow_definition_store_unavailable")
)

type WorkflowDefinitionReleaseContext struct {
	RequestContext  context.Context
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	OwnerSubjectRef string
	ActorRef        string
	RequestID       string
	AuditRef        string
}

type workflowDefinitionReleaseRepository interface {
	CreateCandidate(WorkflowDefinitionReleaseContext, string, string, SavedWorkflowDraft, time.Time) (WorkflowDefinitionReleaseCandidate, error)
	Review(WorkflowDefinitionReleaseContext, string, int, string, string, string, time.Time) (WorkflowDefinitionReleaseCandidate, *WorkflowDefinitionVersion, error)
	DecideActivation(WorkflowDefinitionReleaseContext, string, int, string, int, string, time.Time) (WorkflowDefinitionActivation, error)
	ReadCandidate(WorkflowDefinitionReleaseContext, string) (WorkflowDefinitionReleaseCandidate, error)
	ListCandidates(WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, error)
	ListVersions(WorkflowDefinitionReleaseContext, string) ([]WorkflowDefinitionVersion, error)
	ReadVersion(WorkflowDefinitionReleaseContext, string, int) (WorkflowDefinitionVersion, error)
	ReadActivation(WorkflowDefinitionReleaseContext, string) (WorkflowDefinitionActivation, error)
	ListSummaries(ReadRepositoryContext, ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult
}

type WorkflowDefinitionSnapshot struct {
	SchemaVersion         string                     `json:"schema_version"`
	Name                  string                     `json:"name"`
	Description           string                     `json:"description"`
	Nodes                 []WorkflowDefinitionNode   `json:"nodes"`
	Edges                 []WorkflowDefinitionEdge   `json:"edges"`
	InputContract         WorkflowDefinitionContract `json:"input_contract"`
	OutputContract        WorkflowDefinitionContract `json:"output_contract"`
	ProviderRefs          []string                   `json:"provider_refs"`
	ToolRefs              []string                   `json:"tool_refs"`
	RAGRefs               []string                   `json:"rag_refs"`
	RequestedCapabilities []string                   `json:"requested_capabilities"`
	ExecutionProfile      string                     `json:"execution_profile"`
}

type WorkflowDefinitionNode struct {
	NodeID               string   `json:"node_id"`
	NodeType             string   `json:"node_type"`
	Label                string   `json:"label"`
	InputSummary         string   `json:"input_summary"`
	OutputSummary        string   `json:"output_summary"`
	InputContractRef     string   `json:"input_contract_ref"`
	OutputContractRef    string   `json:"output_contract_ref"`
	InputContractFields  []string `json:"input_contract_fields"`
	OutputContractFields []string `json:"output_contract_fields"`
	OutputMappingSummary string   `json:"output_mapping_summary"`
	ProviderRef          string   `json:"provider_ref"`
	ToolRef              string   `json:"tool_ref"`
	RAGRef               string   `json:"rag_ref"`
	RiskLevel            string   `json:"risk_level"`
	RequiresConfirmation bool     `json:"requires_confirmation"`
}

type WorkflowDefinitionEdge struct {
	EdgeID           string `json:"edge_id"`
	FromNodeID       string `json:"from_node_id"`
	ToNodeID         string `json:"to_node_id"`
	ConditionSummary string `json:"condition_summary"`
}

type WorkflowDefinitionContract struct {
	ContractID     string   `json:"contract_id"`
	RequiredFields []string `json:"required_fields"`
	Summary        string   `json:"summary"`
}

type WorkflowDefinitionReview struct {
	SchemaVersion string `json:"schema_version"`
	ReviewVersion int    `json:"review_version"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	ReviewerRef   string `json:"reviewer_ref"`
	ReviewedAt    string `json:"reviewed_at"`
	RequestID     string `json:"request_id"`
	AuditRef      string `json:"audit_ref"`
}

type WorkflowDefinitionReleaseCandidate struct {
	SchemaVersion       string                     `json:"schema_version"`
	CandidateID         string                     `json:"candidate_id"`
	DefinitionID        string                     `json:"definition_id"`
	SourceDraftID       string                     `json:"source_draft_id"`
	SourceDraftVersion  int                        `json:"source_draft_version"`
	SourceDraftDigest   string                     `json:"source_draft_digest"`
	DefinitionDigest    string                     `json:"definition_digest"`
	Snapshot            WorkflowDefinitionSnapshot `json:"snapshot"`
	ActivationEligible  bool                       `json:"activation_eligible"`
	EligibilityBlockers []string                   `json:"eligibility_blockers"`
	State               string                     `json:"state"`
	ReviewVersion       int                        `json:"review_version"`
	Reviews             []WorkflowDefinitionReview `json:"reviews"`
	CreatedAt           string                     `json:"created_at"`
	UpdatedAt           string                     `json:"updated_at"`
	CreatedByActorRef   string                     `json:"created_by_actor_ref"`
	UpdatedByActorRef   string                     `json:"updated_by_actor_ref"`
	RequestID           string                     `json:"request_id"`
	AuditRef            string                     `json:"audit_ref"`
}

type WorkflowDefinitionVersion struct {
	SchemaVersion          string                     `json:"schema_version"`
	DefinitionID           string                     `json:"definition_id"`
	Version                int                        `json:"version"`
	DefinitionDigest       string                     `json:"definition_digest"`
	CandidateID            string                     `json:"candidate_id"`
	CandidateReviewVersion int                        `json:"candidate_review_version"`
	SourceDraftID          string                     `json:"source_draft_id"`
	SourceDraftVersion     int                        `json:"source_draft_version"`
	SourceDraftDigest      string                     `json:"source_draft_digest"`
	Snapshot               WorkflowDefinitionSnapshot `json:"snapshot"`
	ActivationEligible     bool                       `json:"activation_eligible"`
	EligibilityBlockers    []string                   `json:"eligibility_blockers"`
	CreatedAt              string                     `json:"created_at"`
	CreatedByActorRef      string                     `json:"created_by_actor_ref"`
	RequestID              string                     `json:"request_id"`
	AuditRef               string                     `json:"audit_ref"`
}

type WorkflowDefinitionActivationEvent struct {
	SchemaVersion        string `json:"schema_version"`
	EventID              string `json:"event_id"`
	DefinitionID         string `json:"definition_id"`
	Decision             string `json:"decision"`
	Reason               string `json:"reason"`
	BeforePointerVersion int    `json:"before_pointer_version"`
	AfterPointerVersion  int    `json:"after_pointer_version"`
	BeforeActiveVersion  int    `json:"before_active_version"`
	AfterActiveVersion   int    `json:"after_active_version"`
	ActorRef             string `json:"actor_ref"`
	CreatedAt            string `json:"created_at"`
	RequestID            string `json:"request_id"`
	AuditRef             string `json:"audit_ref"`
}

type WorkflowDefinitionActivation struct {
	SchemaVersion          string                              `json:"schema_version"`
	DefinitionID           string                              `json:"definition_id"`
	PointerVersion         int                                 `json:"pointer_version"`
	State                  string                              `json:"state"`
	ActiveVersion          int                                 `json:"active_version"`
	ActiveDefinitionDigest string                              `json:"active_definition_digest"`
	Events                 []WorkflowDefinitionActivationEvent `json:"events"`
	UpdatedAt              string                              `json:"updated_at"`
	UpdatedByActorRef      string                              `json:"updated_by_actor_ref"`
	RequestID              string                              `json:"request_id"`
	AuditRef               string                              `json:"audit_ref"`
}

type WorkflowDefinitionReleaseAudit struct {
	SchemaVersion string `json:"schema_version"`
	AuditID       string `json:"audit_id"`
	ResourceKind  string `json:"resource_kind"`
	ResourceID    string `json:"resource_id"`
	Action        string `json:"action"`
	ActorRef      string `json:"actor_ref"`
	CreatedAt     string `json:"created_at"`
	RequestID     string `json:"request_id"`
	AuditRef      string `json:"audit_ref"`
}

type workflowDefinitionReleaseStore struct {
	mu          sync.RWMutex
	candidates  map[string]WorkflowDefinitionReleaseCandidate
	versions    map[string][]WorkflowDefinitionVersion
	activations map[string]WorkflowDefinitionActivation
	audits      map[string][]WorkflowDefinitionReleaseAudit
	unavailable bool
}

func newWorkflowDefinitionReleaseStore() *workflowDefinitionReleaseStore {
	return &workflowDefinitionReleaseStore{
		candidates:  map[string]WorkflowDefinitionReleaseCandidate{},
		versions:    map[string][]WorkflowDefinitionVersion{},
		activations: map[string]WorkflowDefinitionActivation{},
		audits:      map[string][]WorkflowDefinitionReleaseAudit{},
	}
}

func workflowDefinitionScopeKey(ctx WorkflowDefinitionReleaseContext, id string) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, id}, "\x00")
}

func workflowDefinitionSnapshotFromDraft(draft SavedWorkflowDraft) (WorkflowDefinitionSnapshot, string, error) {
	snapshot := WorkflowDefinitionSnapshot{
		SchemaVersion:         draft.SchemaVersion,
		Name:                  strings.TrimSpace(draft.Name),
		Description:           strings.TrimSpace(draft.Description),
		Nodes:                 workflowDefinitionNodesFromDraft(draft.Nodes),
		Edges:                 workflowDefinitionEdgesFromDraft(draft.Edges),
		InputContract:         workflowDefinitionContractFromDraft(draft.InputContract),
		OutputContract:        workflowDefinitionContractFromDraft(draft.OutputContract),
		ProviderRefs:          cloneStringSlice(draft.ProviderRefs),
		ToolRefs:              cloneStringSlice(draft.ToolRefs),
		RAGRefs:               cloneStringSlice(draft.RAGRefs),
		RequestedCapabilities: cloneStringSlice(draft.RequestedCapabilities),
		ExecutionProfile:      "workflow_definition_executor_v1",
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return WorkflowDefinitionSnapshot{}, "", err
	}
	digest := sha256.Sum256(payload)
	return snapshot, "sha256:" + hex.EncodeToString(digest[:]), nil
}

func workflowDefinitionEligibility(draft SavedWorkflowDraft) (bool, []string) {
	blockers := make([]string, 0, 4)
	if len(draft.ToolRefs) > 0 {
		blockers = append(blockers, "tool_refs_unsupported")
	}
	if len(draft.RAGRefs) > 0 {
		blockers = append(blockers, "rag_refs_unsupported")
	}
	for _, capability := range draft.RequestedCapabilities {
		normalized := strings.ToLower(strings.TrimSpace(capability))
		if normalized != "" {
			blockers = append(blockers, "capability_unsupported:"+normalized)
		}
	}
	for _, node := range draft.Nodes {
		if node.RequiresConfirmation {
			blockers = append(blockers, "confirmation_authority_required:"+node.NodeID)
		}
	}
	sort.Strings(blockers)
	return len(blockers) == 0, blockers
}

func (store *workflowDefinitionReleaseStore) CreateCandidate(ctx WorkflowDefinitionReleaseContext, candidateID, definitionID string, draft SavedWorkflowDraft, now time.Time) (WorkflowDefinitionReleaseCandidate, error) {
	if !validWorkflowDefinitionContext(ctx) || !applicationDraftIdentifierPattern.MatchString(candidateID) || !applicationDraftIdentifierPattern.MatchString(definitionID) ||
		!applicationDraftIdentifierPattern.MatchString(draft.DraftID) || draft.DraftVersion < 1 {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionPayloadInvalid
	}
	if draft.DraftStatus != SavedWorkflowDraftStatusValidForReview || !draft.ValidationSummary.ValidForReview || len(draft.BlockedCapabilitySummary) > 0 {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionInvalidState
	}
	snapshot, digest, err := workflowDefinitionSnapshotFromDraft(draft)
	if err != nil {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionStore
	}
	if workflowDefinitionSnapshotContainsForbiddenMaterial(snapshot) {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionPayloadInvalid
	}
	eligible, blockers := workflowDefinitionEligibility(draft)
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.unavailable {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionStore
	}
	key := workflowDefinitionScopeKey(ctx, candidateID)
	if _, exists := store.candidates[key]; exists {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionConflict
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	candidate := WorkflowDefinitionReleaseCandidate{
		SchemaVersion:       workflowDefinitionCandidateSchemaVersion,
		CandidateID:         candidateID,
		DefinitionID:        definitionID,
		SourceDraftID:       draft.DraftID,
		SourceDraftVersion:  draft.DraftVersion,
		SourceDraftDigest:   digest,
		DefinitionDigest:    digest,
		Snapshot:            snapshot,
		ActivationEligible:  eligible,
		EligibilityBlockers: blockers,
		State:               workflowDefinitionStatePending,
		Reviews:             []WorkflowDefinitionReview{},
		CreatedAt:           timestamp,
		UpdatedAt:           timestamp,
		CreatedByActorRef:   ctx.ActorRef,
		UpdatedByActorRef:   ctx.ActorRef,
		RequestID:           ctx.RequestID,
		AuditRef:            ctx.AuditRef,
	}
	store.candidates[key] = cloneWorkflowDefinitionCandidate(candidate)
	store.appendAuditLocked(ctx, "candidate", candidateID, "create", timestamp)
	return cloneWorkflowDefinitionCandidate(candidate), nil
}

func (store *workflowDefinitionReleaseStore) Review(ctx WorkflowDefinitionReleaseContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowDefinitionReleaseCandidate, *WorkflowDefinitionVersion, error) {
	decision = strings.TrimSpace(decision)
	reason = strings.TrimSpace(reason)
	if !validWorkflowDefinitionContext(ctx) || !applicationDraftIdentifierPattern.MatchString(candidateID) || expected < 0 ||
		(decision != "approve" && decision != "reject") || !validWorkflowDefinitionReason(reason) {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionPayloadInvalid
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.unavailable {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionStore
	}
	key := workflowDefinitionScopeKey(ctx, candidateID)
	candidate, ok := store.candidates[key]
	if !ok {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionNotFound
	}
	if candidate.State != workflowDefinitionStatePending {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionInvalidState
	}
	if candidate.ReviewVersion != expected {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionConflict
	}
	if sourceDigest != candidate.SourceDraftDigest {
		return WorkflowDefinitionReleaseCandidate{}, nil, errWorkflowDefinitionSourceDrift
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	candidate.ReviewVersion++
	if decision == "approve" {
		candidate.State = workflowDefinitionStateApproved
	} else {
		candidate.State = workflowDefinitionStateRejected
	}
	candidate.UpdatedAt = timestamp
	candidate.UpdatedByActorRef = ctx.ActorRef
	candidate.Reviews = append(candidate.Reviews, WorkflowDefinitionReview{
		SchemaVersion: workflowDefinitionDecisionSchemaVersion,
		ReviewVersion: candidate.ReviewVersion,
		Decision:      decision,
		Reason:        reason,
		ReviewerRef:   ctx.ActorRef,
		ReviewedAt:    timestamp,
		RequestID:     ctx.RequestID,
		AuditRef:      ctx.AuditRef,
	})
	store.candidates[key] = cloneWorkflowDefinitionCandidate(candidate)
	store.appendAuditLocked(ctx, "candidate", candidateID, "review_"+decision, timestamp)
	if decision != "approve" {
		return cloneWorkflowDefinitionCandidate(candidate), nil, nil
	}
	definitionKey := workflowDefinitionScopeKey(ctx, candidate.DefinitionID)
	versionNumber := len(store.versions[definitionKey]) + 1
	version := WorkflowDefinitionVersion{
		SchemaVersion:          workflowDefinitionVersionSchemaVersion,
		DefinitionID:           candidate.DefinitionID,
		Version:                versionNumber,
		DefinitionDigest:       candidate.DefinitionDigest,
		CandidateID:            candidate.CandidateID,
		CandidateReviewVersion: candidate.ReviewVersion,
		SourceDraftID:          candidate.SourceDraftID,
		SourceDraftVersion:     candidate.SourceDraftVersion,
		SourceDraftDigest:      candidate.SourceDraftDigest,
		Snapshot:               cloneWorkflowDefinitionSnapshot(candidate.Snapshot),
		ActivationEligible:     candidate.ActivationEligible,
		EligibilityBlockers:    cloneStringSlice(candidate.EligibilityBlockers),
		CreatedAt:              timestamp,
		CreatedByActorRef:      ctx.ActorRef,
		RequestID:              ctx.RequestID,
		AuditRef:               ctx.AuditRef,
	}
	store.versions[definitionKey] = append(store.versions[definitionKey], cloneWorkflowDefinitionVersion(version))
	store.appendAuditLocked(ctx, "version", fmt.Sprintf("%s:%d", candidate.DefinitionID, versionNumber), "materialize", timestamp)
	copyVersion := cloneWorkflowDefinitionVersion(version)
	return cloneWorkflowDefinitionCandidate(candidate), &copyVersion, nil
}

func (store *workflowDefinitionReleaseStore) DecideActivation(ctx WorkflowDefinitionReleaseContext, definitionID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowDefinitionActivation, error) {
	decision = strings.TrimSpace(decision)
	reason = strings.TrimSpace(reason)
	if !validWorkflowDefinitionContext(ctx) || !applicationDraftIdentifierPattern.MatchString(definitionID) || expected < 0 ||
		(decision != "activate" && decision != "replace" && decision != "deactivate") || !validWorkflowDefinitionReason(reason) ||
		(decision == "deactivate" && version != 0) || (decision != "deactivate" && version < 1) {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionPayloadInvalid
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.unavailable {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionStore
	}
	key := workflowDefinitionScopeKey(ctx, definitionID)
	current, exists := store.activations[key]
	if !exists {
		current = WorkflowDefinitionActivation{SchemaVersion: workflowDefinitionActivationSchemaVersion, DefinitionID: definitionID, State: workflowDefinitionActivationInactive, Events: []WorkflowDefinitionActivationEvent{}}
	}
	if current.PointerVersion != expected {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionConflict
	}
	if (decision == "activate" && current.State != workflowDefinitionActivationInactive) ||
		((decision == "replace" || decision == "deactivate") && current.State != workflowDefinitionActivationActive) {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionInvalidState
	}
	beforeActiveVersion := current.ActiveVersion
	activeDigest := ""
	if decision != "deactivate" {
		versions := store.versions[key]
		if version > len(versions) {
			return WorkflowDefinitionActivation{}, errWorkflowDefinitionNotFound
		}
		selected := versions[version-1]
		if !selected.ActivationEligible || len(selected.EligibilityBlockers) > 0 {
			return WorkflowDefinitionActivation{}, errWorkflowDefinitionInvalidState
		}
		activeDigest = selected.DefinitionDigest
	} else {
		version = 0
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	event := WorkflowDefinitionActivationEvent{
		SchemaVersion:        workflowDefinitionActivationEventSchemaVersion,
		EventID:              fmt.Sprintf("activation-event-%d", current.PointerVersion+1),
		DefinitionID:         definitionID,
		Decision:             decision,
		Reason:               reason,
		BeforePointerVersion: current.PointerVersion,
		AfterPointerVersion:  current.PointerVersion + 1,
		BeforeActiveVersion:  beforeActiveVersion,
		AfterActiveVersion:   version,
		ActorRef:             ctx.ActorRef,
		CreatedAt:            timestamp,
		RequestID:            ctx.RequestID,
		AuditRef:             ctx.AuditRef,
	}
	current.PointerVersion++
	current.State = workflowDefinitionActivationActive
	if decision == "deactivate" {
		current.State = workflowDefinitionActivationInactive
	}
	current.ActiveVersion = version
	current.ActiveDefinitionDigest = activeDigest
	current.Events = append(current.Events, event)
	current.UpdatedAt = timestamp
	current.UpdatedByActorRef = ctx.ActorRef
	current.RequestID = ctx.RequestID
	current.AuditRef = ctx.AuditRef
	store.activations[key] = cloneWorkflowDefinitionActivation(current)
	store.appendAuditLocked(ctx, "activation", definitionID, decision, timestamp)
	return cloneWorkflowDefinitionActivation(current), nil
}

func (store *workflowDefinitionReleaseStore) Activate(ctx WorkflowDefinitionReleaseContext, definitionID string, expected, version int, now time.Time) (WorkflowDefinitionActivation, error) {
	return store.DecideActivation(ctx, definitionID, expected, "activate", version, "activate reviewed definition version", now)
}

func (store *workflowDefinitionReleaseStore) ReadCandidate(ctx WorkflowDefinitionReleaseContext, candidateID string) (WorkflowDefinitionReleaseCandidate, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionStore
	}
	value, ok := store.candidates[workflowDefinitionScopeKey(ctx, candidateID)]
	if !ok {
		return WorkflowDefinitionReleaseCandidate{}, errWorkflowDefinitionNotFound
	}
	return cloneWorkflowDefinitionCandidate(value), nil
}

func (store *workflowDefinitionReleaseStore) ListCandidates(ctx WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return nil, errWorkflowDefinitionStore
	}
	prefix := workflowDefinitionScopeKey(ctx, "")
	values := make([]WorkflowDefinitionReleaseCandidate, 0)
	for key, value := range store.candidates {
		if strings.HasPrefix(key, prefix) {
			values = append(values, cloneWorkflowDefinitionCandidate(value))
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].CreatedAt > values[j].CreatedAt })
	return values, nil
}

func (store *workflowDefinitionReleaseStore) ListVersions(ctx WorkflowDefinitionReleaseContext, definitionID string) ([]WorkflowDefinitionVersion, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return nil, errWorkflowDefinitionStore
	}
	stored := store.versions[workflowDefinitionScopeKey(ctx, definitionID)]
	values := make([]WorkflowDefinitionVersion, len(stored))
	for index := range stored {
		values[index] = cloneWorkflowDefinitionVersion(stored[index])
	}
	return values, nil
}

func (store *workflowDefinitionReleaseStore) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) (WorkflowDefinitionVersion, error) {
	values, err := store.ListVersions(ctx, definitionID)
	if err != nil {
		return WorkflowDefinitionVersion{}, err
	}
	if version < 1 || version > len(values) {
		return WorkflowDefinitionVersion{}, errWorkflowDefinitionNotFound
	}
	return values[version-1], nil
}

func (store *workflowDefinitionReleaseStore) ReadActivation(ctx WorkflowDefinitionReleaseContext, definitionID string) (WorkflowDefinitionActivation, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return WorkflowDefinitionActivation{}, errWorkflowDefinitionStore
	}
	value, ok := store.activations[workflowDefinitionScopeKey(ctx, definitionID)]
	if !ok {
		return WorkflowDefinitionActivation{
			SchemaVersion:     workflowDefinitionActivationSchemaVersion,
			DefinitionID:      definitionID,
			State:             workflowDefinitionActivationInactive,
			Events:            []WorkflowDefinitionActivationEvent{},
			UpdatedAt:         time.Unix(0, 0).UTC().Format(time.RFC3339),
			UpdatedByActorRef: "system",
			RequestID:         ctx.RequestID,
			AuditRef:          ctx.AuditRef,
		}, nil
	}
	return cloneWorkflowDefinitionActivation(value), nil
}

func (store *workflowDefinitionReleaseStore) appendAuditLocked(ctx WorkflowDefinitionReleaseContext, resourceKind, resourceID, action, timestamp string) {
	key := workflowDefinitionScopeKey(ctx, resourceKind)
	sequence := len(store.audits[key]) + 1
	store.audits[key] = append(store.audits[key], WorkflowDefinitionReleaseAudit{
		SchemaVersion: workflowDefinitionAuditSchemaVersion,
		AuditID:       fmt.Sprintf("release-audit-%s-%d", resourceKind, sequence),
		ResourceKind:  resourceKind,
		ResourceID:    resourceID,
		Action:        action,
		ActorRef:      ctx.ActorRef,
		CreatedAt:     timestamp,
		RequestID:     ctx.RequestID,
		AuditRef:      ctx.AuditRef,
	})
}

func validWorkflowDefinitionContext(ctx WorkflowDefinitionReleaseContext) bool {
	return strings.TrimSpace(ctx.TenantRef) != "" && applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) && strings.TrimSpace(ctx.OwnerSubjectRef) != "" &&
		strings.TrimSpace(ctx.ActorRef) != "" && strings.TrimSpace(ctx.RequestID) != "" && strings.TrimSpace(ctx.AuditRef) != ""
}

func validWorkflowDefinitionReason(reason string) bool {
	return utf8.ValidString(reason) && utf8.RuneCountInString(reason) >= 4 && utf8.RuneCountInString(reason) <= 500 && !applicationDraftStringContainsSecret(reason)
}

func workflowDefinitionSnapshotContainsForbiddenMaterial(snapshot WorkflowDefinitionSnapshot) bool {
	payload, err := json.Marshal(snapshot)
	return err != nil || applicationDraftStringContainsSecret(string(payload))
}

func cloneWorkflowDefinitionSnapshot(value WorkflowDefinitionSnapshot) WorkflowDefinitionSnapshot {
	value.Nodes = append([]WorkflowDefinitionNode(nil), value.Nodes...)
	for index := range value.Nodes {
		value.Nodes[index].InputContractFields = cloneStringSlice(value.Nodes[index].InputContractFields)
		value.Nodes[index].OutputContractFields = cloneStringSlice(value.Nodes[index].OutputContractFields)
	}
	value.Edges = append([]WorkflowDefinitionEdge(nil), value.Edges...)
	value.InputContract.RequiredFields = cloneStringSlice(value.InputContract.RequiredFields)
	value.OutputContract.RequiredFields = cloneStringSlice(value.OutputContract.RequiredFields)
	value.ProviderRefs = cloneStringSlice(value.ProviderRefs)
	value.ToolRefs = cloneStringSlice(value.ToolRefs)
	value.RAGRefs = cloneStringSlice(value.RAGRefs)
	value.RequestedCapabilities = cloneStringSlice(value.RequestedCapabilities)
	return value
}

func workflowDefinitionNodesFromDraft(nodes []SavedWorkflowDraftNode) []WorkflowDefinitionNode {
	values := make([]WorkflowDefinitionNode, len(nodes))
	for index, node := range nodes {
		values[index] = WorkflowDefinitionNode{
			NodeID: node.NodeID, NodeType: node.NodeType, Label: node.Label,
			InputSummary: node.InputSummary, OutputSummary: node.OutputSummary,
			InputContractRef: node.InputContractRef, OutputContractRef: node.OutputContractRef,
			InputContractFields: cloneStringSlice(node.InputContractFields), OutputContractFields: cloneStringSlice(node.OutputContractFields),
			OutputMappingSummary: node.OutputMappingSummary, ProviderRef: node.ProviderRef, ToolRef: node.ToolRef, RAGRef: node.RAGRef,
			RiskLevel: node.RiskLevel, RequiresConfirmation: node.RequiresConfirmation,
		}
	}
	return values
}

func workflowDefinitionEdgesFromDraft(edges []SavedWorkflowDraftEdge) []WorkflowDefinitionEdge {
	values := make([]WorkflowDefinitionEdge, len(edges))
	for index, edge := range edges {
		values[index] = WorkflowDefinitionEdge{EdgeID: edge.EdgeID, FromNodeID: edge.FromNodeID, ToNodeID: edge.ToNodeID, ConditionSummary: edge.ConditionSummary}
	}
	return values
}

func workflowDefinitionContractFromDraft(contract SavedWorkflowDraftContract) WorkflowDefinitionContract {
	return WorkflowDefinitionContract{ContractID: contract.ContractID, RequiredFields: cloneStringSlice(contract.RequiredFields), Summary: contract.Summary}
}

func cloneWorkflowDefinitionCandidate(value WorkflowDefinitionReleaseCandidate) WorkflowDefinitionReleaseCandidate {
	value.Snapshot = cloneWorkflowDefinitionSnapshot(value.Snapshot)
	value.EligibilityBlockers = cloneStringSlice(value.EligibilityBlockers)
	value.Reviews = append([]WorkflowDefinitionReview(nil), value.Reviews...)
	return value
}

func cloneWorkflowDefinitionVersion(value WorkflowDefinitionVersion) WorkflowDefinitionVersion {
	value.Snapshot = cloneWorkflowDefinitionSnapshot(value.Snapshot)
	value.EligibilityBlockers = cloneStringSlice(value.EligibilityBlockers)
	return value
}

func cloneWorkflowDefinitionActivation(value WorkflowDefinitionActivation) WorkflowDefinitionActivation {
	value.Events = append([]WorkflowDefinitionActivationEvent(nil), value.Events...)
	return value
}
