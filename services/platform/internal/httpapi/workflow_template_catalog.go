package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	workflowTemplateCandidateSchemaVersion = "workspace_workflow_template_candidate.v1"
	workflowTemplateDecisionSchemaVersion  = "workspace_workflow_template_decision.v1"
	workflowTemplateVersionSchemaVersion   = "workspace_workflow_template_version.v1"
	workflowTemplateLineageSchemaVersion   = "workspace_workflow_template_lineage.v1"
	workflowTemplateListingEventSchema     = "workspace_workflow_template_listing_event.v1"
	workflowTemplateAuditSchemaVersion     = "workspace_workflow_template_audit.v1"

	workflowTemplateCandidatePending          = "pending"
	workflowTemplateCandidateApproved         = "approved"
	workflowTemplateCandidateRejected         = "rejected"
	workflowTemplateCandidateChangesRequested = "changes_requested"
	workflowTemplateCandidateWithdrawn        = "withdrawn"
	workflowTemplateLineageUnlisted           = "unlisted"
	workflowTemplateLineageListed             = "listed"
)

const (
	WorkflowTemplateFailureScopeDenied                  = "workflow_template_scope_denied"
	WorkflowTemplateFailureSourceApplicationUnavailable = "workflow_template_source_application_unavailable"
	WorkflowTemplateFailureSourceDefinitionNotFound     = "workflow_template_source_definition_not_found"
	WorkflowTemplateFailureSourceDefinitionChanged      = "workflow_template_source_definition_changed"
	WorkflowTemplateFailureSourceProfileUnsupported     = "workflow_template_source_profile_unsupported"
	WorkflowTemplateFailureForbiddenCapability          = "workflow_template_forbidden_capability"
	WorkflowTemplateFailurePayloadInvalid               = "workflow_template_payload_invalid"
	WorkflowTemplateFailureSecretMaterialForbidden      = "workflow_template_secret_material_forbidden"
	WorkflowTemplateFailureCandidateNotFound            = "workflow_template_candidate_not_found"
	WorkflowTemplateFailureCandidateVersionConflict     = "workflow_template_candidate_version_conflict"
	WorkflowTemplateFailureReviewTransitionInvalid      = "workflow_template_review_transition_invalid"
	WorkflowTemplateFailureVersionNotFound              = "workflow_template_version_not_found"
	WorkflowTemplateFailurePointerVersionConflict       = "workflow_template_pointer_version_conflict"
	WorkflowTemplateFailureNotListed                    = "workflow_template_not_listed"
	WorkflowTemplateFailureTargetApplicationUnavailable = "workflow_template_target_application_unavailable"
	WorkflowTemplateFailureTargetBindingUnavailable     = "workflow_template_target_binding_unavailable"
	WorkflowTemplateFailureDraftIDConflict              = "workflow_template_draft_id_conflict"
	WorkflowTemplateFailureCursorInvalid                = "workflow_template_cursor_invalid"
	WorkflowTemplateFailureStoreUnavailable             = "workflow_template_store_unavailable"
)

var (
	errWorkflowTemplateCandidateNotFound       = errors.New(WorkflowTemplateFailureCandidateNotFound)
	errWorkflowTemplateCandidateConflict       = errors.New(WorkflowTemplateFailureCandidateVersionConflict)
	errWorkflowTemplateReviewTransition        = errors.New(WorkflowTemplateFailureReviewTransitionInvalid)
	errWorkflowTemplateVersionNotFound         = errors.New(WorkflowTemplateFailureVersionNotFound)
	errWorkflowTemplatePointerConflict         = errors.New(WorkflowTemplateFailurePointerVersionConflict)
	errWorkflowTemplateNotListed               = errors.New(WorkflowTemplateFailureNotListed)
	errWorkflowTemplatePayloadInvalid          = errors.New(WorkflowTemplateFailurePayloadInvalid)
	errWorkflowTemplateSecretMaterialForbidden = errors.New(WorkflowTemplateFailureSecretMaterialForbidden)
	errWorkflowTemplateSourceDefinitionChanged = errors.New(WorkflowTemplateFailureSourceDefinitionChanged)
	errWorkflowTemplateStoreUnavailable        = errors.New(WorkflowTemplateFailureStoreUnavailable)
)

type WorkflowTemplateCatalogContext struct {
	RequestContext  context.Context
	TenantRef       string
	WorkspaceID     string
	OwnerSubjectRef string
	ActorRef        string
	RequestID       string
	AuditRef        string
}

type WorkflowTemplatePortabilitySummary struct {
	ExecutionProfile string   `json:"execution_profile"`
	NodeKinds        []string `json:"node_kinds"`
	ProviderRefs     []string `json:"provider_refs"`
	RiskLevel        string   `json:"risk_level"`
	Portable         bool     `json:"portable"`
	Blockers         []string `json:"blockers"`
}

type WorkflowTemplateReviewDecision struct {
	SchemaVersion string `json:"schema_version"`
	ReviewVersion int    `json:"review_version"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
	ReviewerRef   string `json:"reviewer_ref"`
	DecidedAt     string `json:"decided_at"`
	RequestID     string `json:"request_id"`
	AuditRef      string `json:"audit_ref"`
}

type WorkflowTemplateCandidate struct {
	SchemaVersion           string                             `json:"schema_version"`
	CandidateID             string                             `json:"candidate_id"`
	TemplateID              string                             `json:"template_id"`
	State                   string                             `json:"state"`
	ReviewVersion           int                                `json:"review_version"`
	SourceApplicationID     string                             `json:"source_application_id"`
	SourceOwnerSubjectRef   string                             `json:"source_owner_subject_ref"`
	SourceDefinitionID      string                             `json:"source_definition_id"`
	SourceDefinitionVersion int                                `json:"source_definition_version"`
	SourceDefinitionDigest  string                             `json:"source_definition_digest"`
	Title                   string                             `json:"title"`
	Summary                 string                             `json:"summary"`
	UsageNotes              string                             `json:"usage_notes"`
	Labels                  []string                           `json:"labels"`
	Portability             WorkflowTemplatePortabilitySummary `json:"portability"`
	Decisions               []WorkflowTemplateReviewDecision   `json:"decisions"`
	CreatedAt               string                             `json:"created_at"`
	UpdatedAt               string                             `json:"updated_at"`
	CreatedByActorRef       string                             `json:"created_by_actor_ref"`
	UpdatedByActorRef       string                             `json:"updated_by_actor_ref"`
	RequestID               string                             `json:"request_id"`
	AuditRef                string                             `json:"audit_ref"`
}

type WorkflowTemplateVersion struct {
	SchemaVersion           string                             `json:"schema_version"`
	TemplateID              string                             `json:"template_id"`
	Version                 int                                `json:"version"`
	TemplateDigest          string                             `json:"template_digest"`
	CandidateID             string                             `json:"candidate_id"`
	CandidateReviewVersion  int                                `json:"candidate_review_version"`
	SourceApplicationID     string                             `json:"source_application_id"`
	SourceOwnerSubjectRef   string                             `json:"source_owner_subject_ref"`
	SourceDefinitionID      string                             `json:"source_definition_id"`
	SourceDefinitionVersion int                                `json:"source_definition_version"`
	SourceDefinitionDigest  string                             `json:"source_definition_digest"`
	Title                   string                             `json:"title"`
	Summary                 string                             `json:"summary"`
	UsageNotes              string                             `json:"usage_notes"`
	Labels                  []string                           `json:"labels"`
	Portability             WorkflowTemplatePortabilitySummary `json:"portability"`
	CreatedAt               string                             `json:"created_at"`
	CreatedByActorRef       string                             `json:"created_by_actor_ref"`
	RequestID               string                             `json:"request_id"`
	AuditRef                string                             `json:"audit_ref"`
}

type WorkflowTemplateListingEvent struct {
	SchemaVersion        string `json:"schema_version"`
	EventID              string `json:"event_id"`
	TemplateID           string `json:"template_id"`
	Decision             string `json:"decision"`
	Reason               string `json:"reason"`
	BeforePointerVersion int    `json:"before_pointer_version"`
	AfterPointerVersion  int    `json:"after_pointer_version"`
	BeforeListedVersion  int    `json:"before_listed_version"`
	AfterListedVersion   int    `json:"after_listed_version"`
	ActorRef             string `json:"actor_ref"`
	CreatedAt            string `json:"created_at"`
	RequestID            string `json:"request_id"`
	AuditRef             string `json:"audit_ref"`
}

type WorkflowTemplateLineage struct {
	SchemaVersion     string                         `json:"schema_version"`
	TemplateID        string                         `json:"template_id"`
	TenantRef         string                         `json:"tenant_ref"`
	WorkspaceID       string                         `json:"workspace_id"`
	PointerVersion    int                            `json:"pointer_version"`
	Lifecycle         string                         `json:"lifecycle"`
	ListedVersion     int                            `json:"listed_version"`
	ListedDigest      string                         `json:"listed_digest"`
	Events            []WorkflowTemplateListingEvent `json:"events"`
	CreatedAt         string                         `json:"created_at"`
	UpdatedAt         string                         `json:"updated_at"`
	CreatedByActorRef string                         `json:"created_by_actor_ref"`
	UpdatedByActorRef string                         `json:"updated_by_actor_ref"`
	RequestID         string                         `json:"request_id"`
	AuditRef          string                         `json:"audit_ref"`
}

type WorkflowTemplateAudit struct {
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

type workflowTemplateCatalogRepository interface {
	CreateCandidate(WorkflowTemplateCatalogContext, WorkflowTemplateCandidate, time.Time) (WorkflowTemplateCandidate, error)
	ReviewCandidate(WorkflowTemplateCatalogContext, string, int, string, string, string, time.Time) (WorkflowTemplateCandidate, *WorkflowTemplateVersion, error)
	DecideListing(WorkflowTemplateCatalogContext, string, int, string, int, string, time.Time) (WorkflowTemplateLineage, error)
	ReadCandidate(WorkflowTemplateCatalogContext, string) (WorkflowTemplateCandidate, error)
	ListCandidates(WorkflowTemplateCatalogContext) ([]WorkflowTemplateCandidate, error)
	ReadLineage(WorkflowTemplateCatalogContext, string) (WorkflowTemplateLineage, error)
	ListLineages(WorkflowTemplateCatalogContext) ([]WorkflowTemplateLineage, error)
	ReadVersion(WorkflowTemplateCatalogContext, string, int) (WorkflowTemplateVersion, error)
	ListVersions(WorkflowTemplateCatalogContext, string) ([]WorkflowTemplateVersion, error)
}

type memoryWorkflowTemplateCatalogRepository struct {
	mu          sync.RWMutex
	candidates  map[string]WorkflowTemplateCandidate
	versions    map[string][]WorkflowTemplateVersion
	lineages    map[string]WorkflowTemplateLineage
	audits      map[string][]WorkflowTemplateAudit
	unavailable bool
}

func newMemoryWorkflowTemplateCatalogRepository() *memoryWorkflowTemplateCatalogRepository {
	return &memoryWorkflowTemplateCatalogRepository{
		candidates: map[string]WorkflowTemplateCandidate{}, versions: map[string][]WorkflowTemplateVersion{},
		lineages: map[string]WorkflowTemplateLineage{}, audits: map[string][]WorkflowTemplateAudit{},
	}
}

func workflowTemplateScopeKey(ctx WorkflowTemplateCatalogContext, id string) string {
	return strings.Join([]string{strings.TrimSpace(ctx.TenantRef), strings.TrimSpace(ctx.WorkspaceID), id}, "\x00")
}

func (repository *memoryWorkflowTemplateCatalogRepository) CreateCandidate(ctx WorkflowTemplateCatalogContext, candidate WorkflowTemplateCandidate, now time.Time) (WorkflowTemplateCandidate, error) {
	if !validWorkflowTemplateContext(ctx) || !validWorkflowTemplateCandidate(candidate) {
		return WorkflowTemplateCandidate{}, errWorkflowTemplatePayloadInvalid
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, candidate.CandidateID)
	if _, found := repository.candidates[key]; found {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateCandidateConflict
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	candidate.SchemaVersion = workflowTemplateCandidateSchemaVersion
	candidate.State = workflowTemplateCandidatePending
	candidate.ReviewVersion = 0
	candidate.Decisions = []WorkflowTemplateReviewDecision{}
	candidate.CreatedAt, candidate.UpdatedAt = timestamp, timestamp
	candidate.CreatedByActorRef, candidate.UpdatedByActorRef = ctx.ActorRef, ctx.ActorRef
	candidate.RequestID, candidate.AuditRef = ctx.RequestID, ctx.AuditRef
	repository.candidates[key] = cloneWorkflowTemplateCandidate(candidate)
	repository.appendAuditLocked(ctx, "candidate", candidate.CandidateID, "create", timestamp)
	return cloneWorkflowTemplateCandidate(candidate), nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ReviewCandidate(ctx WorkflowTemplateCatalogContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowTemplateCandidate, *WorkflowTemplateVersion, error) {
	if !validWorkflowTemplateContext(ctx) || !applicationDraftIdentifierPattern.MatchString(candidateID) || expected < 0 ||
		!validWorkflowTemplateReviewDecision(decision) || !validWorkflowTemplateReason(reason) {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplatePayloadInvalid
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, candidateID)
	candidate, found := repository.candidates[key]
	if !found {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateCandidateNotFound
	}
	if candidate.State != workflowTemplateCandidatePending {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateReviewTransition
	}
	if candidate.ReviewVersion != expected {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateCandidateConflict
	}
	if candidate.SourceDefinitionDigest != sourceDigest {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateSourceDefinitionChanged
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	candidate.ReviewVersion++
	candidate.State = workflowTemplateCandidateStateForDecision(decision)
	candidate.UpdatedAt, candidate.UpdatedByActorRef = timestamp, ctx.ActorRef
	candidate.RequestID, candidate.AuditRef = ctx.RequestID, ctx.AuditRef
	candidate.Decisions = append(candidate.Decisions, WorkflowTemplateReviewDecision{
		SchemaVersion: workflowTemplateDecisionSchemaVersion, ReviewVersion: candidate.ReviewVersion,
		Decision: decision, Reason: reason, ReviewerRef: ctx.ActorRef, DecidedAt: timestamp,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if decision != "approve" {
		repository.candidates[key] = cloneWorkflowTemplateCandidate(candidate)
		repository.appendAuditLocked(ctx, "candidate", candidateID, "review_"+decision, timestamp)
		return cloneWorkflowTemplateCandidate(candidate), nil, nil
	}
	versionNumber := len(repository.versions[workflowTemplateScopeKey(ctx, candidate.TemplateID)]) + 1
	version := workflowTemplateVersionFromCandidate(candidate, versionNumber, ctx, timestamp)
	digest, err := workflowTemplateVersionDigest(version)
	if err != nil {
		return WorkflowTemplateCandidate{}, nil, errWorkflowTemplateStoreUnavailable
	}
	version.TemplateDigest = digest
	versionKey := workflowTemplateScopeKey(ctx, candidate.TemplateID)
	repository.candidates[key] = cloneWorkflowTemplateCandidate(candidate)
	repository.appendAuditLocked(ctx, "candidate", candidateID, "review_"+decision, timestamp)
	repository.versions[versionKey] = append(repository.versions[versionKey], cloneWorkflowTemplateVersion(version))
	lineage, exists := repository.lineages[versionKey]
	if !exists {
		lineage = WorkflowTemplateLineage{
			SchemaVersion: workflowTemplateLineageSchemaVersion, TemplateID: candidate.TemplateID,
			TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Lifecycle: workflowTemplateLineageUnlisted,
			Events: []WorkflowTemplateListingEvent{}, CreatedAt: timestamp, UpdatedAt: timestamp,
			CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		}
		repository.lineages[versionKey] = lineage
	}
	repository.appendAuditLocked(ctx, "version", fmt.Sprintf("%s:%d", candidate.TemplateID, versionNumber), "materialize", timestamp)
	return cloneWorkflowTemplateCandidate(candidate), cloneWorkflowTemplateVersionPointer(version), nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) DecideListing(ctx WorkflowTemplateCatalogContext, templateID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowTemplateLineage, error) {
	if !validWorkflowTemplateContext(ctx) || !applicationDraftIdentifierPattern.MatchString(templateID) || expected < 0 ||
		!validWorkflowTemplateListingDecision(decision) || !validWorkflowTemplateReason(reason) {
		return WorkflowTemplateLineage{}, errWorkflowTemplatePayloadInvalid
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return WorkflowTemplateLineage{}, errWorkflowTemplateStoreUnavailable
	}
	key := workflowTemplateScopeKey(ctx, templateID)
	lineage, found := repository.lineages[key]
	if !found {
		return WorkflowTemplateLineage{}, errWorkflowTemplateVersionNotFound
	}
	if lineage.PointerVersion != expected {
		return WorkflowTemplateLineage{}, errWorkflowTemplatePointerConflict
	}
	beforeVersion := lineage.ListedVersion
	switch decision {
	case "list":
		if lineage.Lifecycle != workflowTemplateLineageUnlisted || version < 1 {
			return WorkflowTemplateLineage{}, errWorkflowTemplateNotListed
		}
	case "replace":
		if lineage.Lifecycle != workflowTemplateLineageListed || version < 1 || version == lineage.ListedVersion {
			return WorkflowTemplateLineage{}, errWorkflowTemplateNotListed
		}
	case "unlist":
		if lineage.Lifecycle != workflowTemplateLineageListed || version != 0 {
			return WorkflowTemplateLineage{}, errWorkflowTemplateNotListed
		}
	}
	var selected WorkflowTemplateVersion
	if decision != "unlist" {
		versions := repository.versions[key]
		if version > len(versions) {
			return WorkflowTemplateLineage{}, errWorkflowTemplateVersionNotFound
		}
		selected = versions[version-1]
	}
	timestamp := now.UTC().Format(time.RFC3339Nano)
	lineage.PointerVersion++
	lineage.UpdatedAt, lineage.UpdatedByActorRef = timestamp, ctx.ActorRef
	lineage.RequestID, lineage.AuditRef = ctx.RequestID, ctx.AuditRef
	if decision == "unlist" {
		lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest = workflowTemplateLineageUnlisted, 0, ""
	} else {
		lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest = workflowTemplateLineageListed, selected.Version, selected.TemplateDigest
	}
	lineage.Events = append(lineage.Events, WorkflowTemplateListingEvent{
		SchemaVersion: workflowTemplateListingEventSchema, EventID: fmt.Sprintf("template-listing-%d", lineage.PointerVersion),
		TemplateID: templateID, Decision: decision, Reason: reason, BeforePointerVersion: expected,
		AfterPointerVersion: lineage.PointerVersion, BeforeListedVersion: beforeVersion, AfterListedVersion: lineage.ListedVersion,
		ActorRef: ctx.ActorRef, CreatedAt: timestamp, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	repository.lineages[key] = cloneWorkflowTemplateLineage(lineage)
	repository.appendAuditLocked(ctx, "lineage", templateID, decision, timestamp)
	return cloneWorkflowTemplateLineage(lineage), nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ReadCandidate(ctx WorkflowTemplateCatalogContext, candidateID string) (WorkflowTemplateCandidate, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateStoreUnavailable
	}
	value, found := repository.candidates[workflowTemplateScopeKey(ctx, candidateID)]
	if !found {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateCandidateNotFound
	}
	if validateStoredWorkflowTemplateCandidate(value) != nil {
		return WorkflowTemplateCandidate{}, errWorkflowTemplateStoreUnavailable
	}
	return cloneWorkflowTemplateCandidate(value), nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ListCandidates(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateCandidate, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	prefix := workflowTemplateScopeKey(ctx, "")
	values := make([]WorkflowTemplateCandidate, 0)
	for key, value := range repository.candidates {
		if strings.HasPrefix(key, prefix) {
			if validateStoredWorkflowTemplateCandidate(value) != nil {
				return nil, errWorkflowTemplateStoreUnavailable
			}
			values = append(values, cloneWorkflowTemplateCandidate(value))
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].UpdatedAt == values[j].UpdatedAt {
			return values[i].CandidateID > values[j].CandidateID
		}
		return values[i].UpdatedAt > values[j].UpdatedAt
	})
	return values, nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ReadLineage(ctx WorkflowTemplateCatalogContext, templateID string) (WorkflowTemplateLineage, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return WorkflowTemplateLineage{}, errWorkflowTemplateStoreUnavailable
	}
	value, found := repository.lineages[workflowTemplateScopeKey(ctx, templateID)]
	if !found {
		return WorkflowTemplateLineage{}, errWorkflowTemplateVersionNotFound
	}
	if validateStoredWorkflowTemplateLineage(value) != nil ||
		validateWorkflowTemplateLineageVersion(value, repository.versions[workflowTemplateScopeKey(ctx, templateID)]) != nil {
		return WorkflowTemplateLineage{}, errWorkflowTemplateStoreUnavailable
	}
	return cloneWorkflowTemplateLineage(value), nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ListLineages(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateLineage, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	prefix := workflowTemplateScopeKey(ctx, "")
	values := make([]WorkflowTemplateLineage, 0)
	for key, value := range repository.lineages {
		if strings.HasPrefix(key, prefix) {
			if validateStoredWorkflowTemplateLineage(value) != nil || validateWorkflowTemplateLineageVersion(value, repository.versions[key]) != nil {
				return nil, errWorkflowTemplateStoreUnavailable
			}
			values = append(values, cloneWorkflowTemplateLineage(value))
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].UpdatedAt == values[j].UpdatedAt {
			return values[i].TemplateID > values[j].TemplateID
		}
		return values[i].UpdatedAt > values[j].UpdatedAt
	})
	return values, nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ReadVersion(ctx WorkflowTemplateCatalogContext, templateID string, version int) (WorkflowTemplateVersion, error) {
	values, err := repository.ListVersions(ctx, templateID)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	if version < 1 || version > len(values) {
		return WorkflowTemplateVersion{}, errWorkflowTemplateVersionNotFound
	}
	return values[version-1], nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) ListVersions(ctx WorkflowTemplateCatalogContext, templateID string) ([]WorkflowTemplateVersion, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	stored := repository.versions[workflowTemplateScopeKey(ctx, templateID)]
	values := make([]WorkflowTemplateVersion, len(stored))
	for index := range stored {
		if validateStoredWorkflowTemplateVersion(stored[index]) != nil || stored[index].Version != index+1 {
			return nil, errWorkflowTemplateStoreUnavailable
		}
		values[index] = cloneWorkflowTemplateVersion(stored[index])
	}
	return values, nil
}

func (repository *memoryWorkflowTemplateCatalogRepository) appendAuditLocked(ctx WorkflowTemplateCatalogContext, kind, id, action, timestamp string) {
	key := workflowTemplateScopeKey(ctx, "audits")
	repository.audits[key] = append(repository.audits[key], WorkflowTemplateAudit{
		SchemaVersion: workflowTemplateAuditSchemaVersion, AuditID: fmt.Sprintf("template-audit-%d", len(repository.audits[key])+1),
		ResourceKind: kind, ResourceID: id, Action: action, ActorRef: ctx.ActorRef,
		CreatedAt: timestamp, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
}

func workflowTemplateVersionFromCandidate(candidate WorkflowTemplateCandidate, version int, ctx WorkflowTemplateCatalogContext, timestamp string) WorkflowTemplateVersion {
	return WorkflowTemplateVersion{
		SchemaVersion: workflowTemplateVersionSchemaVersion, TemplateID: candidate.TemplateID, Version: version,
		CandidateID: candidate.CandidateID, CandidateReviewVersion: candidate.ReviewVersion,
		SourceApplicationID: candidate.SourceApplicationID, SourceOwnerSubjectRef: candidate.SourceOwnerSubjectRef,
		SourceDefinitionID: candidate.SourceDefinitionID, SourceDefinitionVersion: candidate.SourceDefinitionVersion,
		SourceDefinitionDigest: candidate.SourceDefinitionDigest, Title: candidate.Title, Summary: candidate.Summary,
		UsageNotes: candidate.UsageNotes, Labels: cloneStringSlice(candidate.Labels), Portability: cloneWorkflowTemplatePortability(candidate.Portability),
		CreatedAt: timestamp, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func workflowTemplateVersionDigest(version WorkflowTemplateVersion) (string, error) {
	document := struct {
		TemplateID              string                             `json:"template_id"`
		Version                 int                                `json:"version"`
		SourceApplicationID     string                             `json:"source_application_id"`
		SourceDefinitionID      string                             `json:"source_definition_id"`
		SourceDefinitionVersion int                                `json:"source_definition_version"`
		SourceDefinitionDigest  string                             `json:"source_definition_digest"`
		Title                   string                             `json:"title"`
		Summary                 string                             `json:"summary"`
		UsageNotes              string                             `json:"usage_notes"`
		Labels                  []string                           `json:"labels"`
		Portability             WorkflowTemplatePortabilitySummary `json:"portability"`
	}{
		TemplateID:              version.TemplateID,
		Version:                 version.Version,
		SourceApplicationID:     version.SourceApplicationID,
		SourceDefinitionID:      version.SourceDefinitionID,
		SourceDefinitionVersion: version.SourceDefinitionVersion,
		SourceDefinitionDigest:  version.SourceDefinitionDigest,
		Title:                   version.Title,
		Summary:                 version.Summary,
		UsageNotes:              version.UsageNotes,
		Labels:                  cloneStringSlice(version.Labels),
		Portability:             cloneWorkflowTemplatePortability(version.Portability),
	}
	payload, err := json.Marshal(document)
	if err != nil || applicationDraftStringContainsSecret(string(payload)) {
		return "", errWorkflowTemplateStoreUnavailable
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func encodeWorkflowTemplateRecord(value any) ([]byte, error) {
	if validateStoredWorkflowTemplateRecord(value) != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	payload, err := json.Marshal(value)
	if err != nil || applicationDraftStringContainsSecret(string(payload)) {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	return payload, nil
}

func decodeWorkflowTemplateRecord(payload []byte, target any) error {
	if validateNoDuplicateJSONFields(payload) != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errWorkflowTemplateStoreUnavailable
	}
	canonical, err := json.Marshal(target)
	if err != nil || applicationDraftStringContainsSecret(string(canonical)) {
		return errWorkflowTemplateStoreUnavailable
	}
	return validateStoredWorkflowTemplateRecord(target)
}

func validateStoredWorkflowTemplateRecord(value any) error {
	switch typed := value.(type) {
	case WorkflowTemplateReviewDecision:
		return validateStoredWorkflowTemplateReviewDecision(typed)
	case *WorkflowTemplateReviewDecision:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateReviewDecision(*typed)
	case WorkflowTemplateCandidate:
		return validateStoredWorkflowTemplateCandidate(typed)
	case *WorkflowTemplateCandidate:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateCandidate(*typed)
	case WorkflowTemplateVersion:
		return validateStoredWorkflowTemplateVersion(typed)
	case *WorkflowTemplateVersion:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateVersion(*typed)
	case WorkflowTemplateListingEvent:
		return validateStoredWorkflowTemplateListingEvent(typed)
	case *WorkflowTemplateListingEvent:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateListingEvent(*typed)
	case WorkflowTemplateLineage:
		return validateStoredWorkflowTemplateLineage(typed)
	case *WorkflowTemplateLineage:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateLineage(*typed)
	case WorkflowTemplateAudit:
		return validateStoredWorkflowTemplateAudit(typed)
	case *WorkflowTemplateAudit:
		if typed == nil {
			return errWorkflowTemplateStoreUnavailable
		}
		return validateStoredWorkflowTemplateAudit(*typed)
	default:
		return errWorkflowTemplateStoreUnavailable
	}
}

func validateStoredWorkflowTemplateCandidate(value WorkflowTemplateCandidate) error {
	created, createdErr := time.Parse(time.RFC3339Nano, value.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, value.UpdatedAt)
	if value.SchemaVersion != workflowTemplateCandidateSchemaVersion || !validWorkflowTemplateCandidate(value) ||
		value.ReviewVersion != len(value.Decisions) || createdErr != nil || updatedErr != nil || updated.Before(created) ||
		!validWorkflowTemplateReference(value.CreatedByActorRef, 240) || !validWorkflowTemplateReference(value.UpdatedByActorRef, 240) ||
		!validWorkflowTemplateReference(value.RequestID, 240) || !validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	if value.State == workflowTemplateCandidatePending && value.ReviewVersion != 0 ||
		value.State != workflowTemplateCandidatePending && value.ReviewVersion != 1 {
		return errWorkflowTemplateStoreUnavailable
	}
	for index, decision := range value.Decisions {
		if validateStoredWorkflowTemplateReviewDecision(decision) != nil || decision.ReviewVersion != index+1 ||
			workflowTemplateCandidateStateForDecision(decision.Decision) != value.State || strings.TrimSpace(decision.ReviewerRef) == "" ||
			strings.TrimSpace(decision.RequestID) == "" || strings.TrimSpace(decision.AuditRef) == "" {
			return errWorkflowTemplateStoreUnavailable
		}
	}
	return nil
}

func validateStoredWorkflowTemplateReviewDecision(value WorkflowTemplateReviewDecision) error {
	if value.SchemaVersion != workflowTemplateDecisionSchemaVersion || value.ReviewVersion != 1 ||
		!validWorkflowTemplateReviewDecision(value.Decision) || !validWorkflowTemplateReason(value.Reason) ||
		!validWorkflowTemplateReference(value.ReviewerRef, 240) || !validWorkflowTemplateReference(value.RequestID, 240) ||
		!validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	if _, err := time.Parse(time.RFC3339Nano, value.DecidedAt); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateStoredWorkflowTemplateVersion(value WorkflowTemplateVersion) error {
	if value.SchemaVersion != workflowTemplateVersionSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.TemplateID) || value.Version < 1 ||
		!workflowRAGDigestPattern.MatchString(value.TemplateDigest) || !applicationDraftIdentifierPattern.MatchString(value.CandidateID) || value.CandidateReviewVersion != 1 ||
		!applicationCatalogIDPattern.MatchString(value.SourceApplicationID) || !validWorkflowTemplateReference(value.SourceOwnerSubjectRef, 240) ||
		!applicationDraftIdentifierPattern.MatchString(value.SourceDefinitionID) || value.SourceDefinitionVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(value.SourceDefinitionDigest) || !validWorkflowTemplateMetadata(value.Title, value.Summary, value.UsageNotes, value.Labels) ||
		!validWorkflowTemplatePortabilitySummary(value.Portability) ||
		!validWorkflowTemplateReference(value.CreatedByActorRef, 240) || !validWorkflowTemplateReference(value.RequestID, 240) ||
		!validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	if _, err := time.Parse(time.RFC3339Nano, value.CreatedAt); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	digest, err := workflowTemplateVersionDigest(value)
	if err != nil || digest != value.TemplateDigest {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateStoredWorkflowTemplateLineage(value WorkflowTemplateLineage) error {
	created, createdErr := time.Parse(time.RFC3339Nano, value.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, value.UpdatedAt)
	if value.SchemaVersion != workflowTemplateLineageSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.TemplateID) ||
		!validWorkflowTemplateReference(value.TenantRef, 240) || !applicationDraftIdentifierPattern.MatchString(value.WorkspaceID) ||
		value.PointerVersion != len(value.Events) || createdErr != nil || updatedErr != nil || updated.Before(created) ||
		!validWorkflowTemplateReference(value.CreatedByActorRef, 240) || !validWorkflowTemplateReference(value.UpdatedByActorRef, 240) ||
		!validWorkflowTemplateReference(value.RequestID, 240) || !validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	if value.Lifecycle == workflowTemplateLineageListed {
		if value.ListedVersion < 1 || !workflowRAGDigestPattern.MatchString(value.ListedDigest) {
			return errWorkflowTemplateStoreUnavailable
		}
	} else if value.Lifecycle != workflowTemplateLineageUnlisted || value.ListedVersion != 0 || value.ListedDigest != "" {
		return errWorkflowTemplateStoreUnavailable
	}
	previousListedVersion := 0
	for index, event := range value.Events {
		if validateStoredWorkflowTemplateListingEvent(event) != nil || event.TemplateID != value.TemplateID ||
			event.BeforePointerVersion != index || event.AfterPointerVersion != index+1 || event.BeforeListedVersion != previousListedVersion {
			return errWorkflowTemplateStoreUnavailable
		}
		previousListedVersion = event.AfterListedVersion
	}
	if previousListedVersion != value.ListedVersion {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateStoredWorkflowTemplateListingEvent(value WorkflowTemplateListingEvent) error {
	if value.SchemaVersion != workflowTemplateListingEventSchema || strings.TrimSpace(value.EventID) == "" ||
		!applicationDraftIdentifierPattern.MatchString(value.TemplateID) || !validWorkflowTemplateListingDecision(value.Decision) ||
		!validWorkflowTemplateReason(value.Reason) || value.BeforePointerVersion < 0 || value.AfterPointerVersion != value.BeforePointerVersion+1 ||
		value.BeforeListedVersion < 0 || value.AfterListedVersion < 0 || !applicationDraftIdentifierPattern.MatchString(value.EventID) ||
		!validWorkflowTemplateReference(value.ActorRef, 240) || !validWorkflowTemplateReference(value.RequestID, 240) ||
		!validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	switch value.Decision {
	case "list":
		if value.BeforeListedVersion != 0 || value.AfterListedVersion < 1 {
			return errWorkflowTemplateStoreUnavailable
		}
	case "replace":
		if value.BeforeListedVersion < 1 || value.AfterListedVersion < 1 || value.AfterListedVersion == value.BeforeListedVersion {
			return errWorkflowTemplateStoreUnavailable
		}
	case "unlist":
		if value.BeforeListedVersion < 1 || value.AfterListedVersion != 0 {
			return errWorkflowTemplateStoreUnavailable
		}
	}
	if _, err := time.Parse(time.RFC3339Nano, value.CreatedAt); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateWorkflowTemplateLineageVersion(lineage WorkflowTemplateLineage, versions []WorkflowTemplateVersion) error {
	if lineage.Lifecycle == workflowTemplateLineageUnlisted {
		return nil
	}
	if lineage.ListedVersion < 1 || lineage.ListedVersion > len(versions) {
		return errWorkflowTemplateStoreUnavailable
	}
	version := versions[lineage.ListedVersion-1]
	if validateStoredWorkflowTemplateVersion(version) != nil || version.TemplateID != lineage.TemplateID ||
		version.Version != lineage.ListedVersion || version.TemplateDigest != lineage.ListedDigest {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validateStoredWorkflowTemplateAudit(value WorkflowTemplateAudit) error {
	if value.SchemaVersion != workflowTemplateAuditSchemaVersion || !applicationDraftIdentifierPattern.MatchString(value.AuditID) ||
		(value.ResourceKind != "candidate" && value.ResourceKind != "version" && value.ResourceKind != "lineage") ||
		!validWorkflowTemplateReference(value.ResourceID, 200) || !validWorkflowTemplateAuditAction(value.Action) ||
		!validWorkflowTemplateReference(value.ActorRef, 240) || !validWorkflowTemplateReference(value.RequestID, 240) ||
		!validWorkflowTemplateReference(value.AuditRef, 240) {
		return errWorkflowTemplateStoreUnavailable
	}
	if _, err := time.Parse(time.RFC3339Nano, value.CreatedAt); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	return nil
}

func validWorkflowTemplateAuditAction(value string) bool {
	switch value {
	case "create", "review_approve", "review_reject", "review_request_changes", "review_withdraw", "materialize", "list", "replace", "unlist":
		return true
	default:
		return false
	}
}

func validWorkflowTemplateContext(ctx WorkflowTemplateCatalogContext) bool {
	return validWorkflowTemplateReference(ctx.TenantRef, 240) && applicationDraftIdentifierPattern.MatchString(ctx.WorkspaceID) &&
		validWorkflowTemplateReference(ctx.OwnerSubjectRef, 240) && validWorkflowTemplateReference(ctx.ActorRef, 240) &&
		validWorkflowTemplateReference(ctx.RequestID, 240) && validWorkflowTemplateReference(ctx.AuditRef, 240)
}

func validWorkflowTemplateCandidate(candidate WorkflowTemplateCandidate) bool {
	return applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(candidate.CandidateID)) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(candidate.TemplateID)) &&
		applicationCatalogIDPattern.MatchString(strings.TrimSpace(candidate.SourceApplicationID)) &&
		validWorkflowTemplateReference(candidate.SourceOwnerSubjectRef, 240) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(candidate.SourceDefinitionID)) &&
		candidate.SourceDefinitionVersion > 0 && workflowRAGDigestPattern.MatchString(candidate.SourceDefinitionDigest) &&
		validWorkflowTemplateMetadata(candidate.Title, candidate.Summary, candidate.UsageNotes, candidate.Labels) &&
		!applicationDraftStringContainsSecret(candidate.Title) && !applicationDraftStringContainsSecret(candidate.Summary) &&
		!applicationDraftStringContainsSecret(candidate.UsageNotes) &&
		validWorkflowTemplatePortabilitySummary(candidate.Portability)
}

func validWorkflowTemplateReference(value string, maximum int) bool {
	return utf8.ValidString(value) && strings.TrimSpace(value) == value && utf8.RuneCountInString(value) >= 1 && utf8.RuneCountInString(value) <= maximum
}

func validWorkflowTemplatePortabilitySummary(value WorkflowTemplatePortabilitySummary) bool {
	if value.ExecutionProfile != workflowDefinitionExecutorProfile || !value.Portable || len(value.Blockers) != 0 ||
		(value.RiskLevel != "low" && value.RiskLevel != "medium" && value.RiskLevel != "high") ||
		len(value.NodeKinds) < 1 || len(value.NodeKinds) > 4 {
		return false
	}
	for index, kind := range value.NodeKinds {
		if kind != "prompt" && kind != "llm" && kind != "condition" && kind != "output" {
			return false
		}
		if index > 0 && value.NodeKinds[index-1] >= kind {
			return false
		}
	}
	for index, reference := range value.ProviderRefs {
		if strings.TrimSpace(reference) != reference || !validWorkflowTemplateProviderRef(reference) ||
			applicationDraftStringContainsSecret(reference) || index > 0 && value.ProviderRefs[index-1] >= reference {
			return false
		}
	}
	return true
}

func validWorkflowTemplateMetadata(title, summary, usageNotes string, labels []string) bool {
	title, summary, usageNotes = strings.TrimSpace(title), strings.TrimSpace(summary), strings.TrimSpace(usageNotes)
	if !utf8.ValidString(title) || !utf8.ValidString(summary) || !utf8.ValidString(usageNotes) ||
		utf8.RuneCountInString(title) < 2 || utf8.RuneCountInString(title) > 120 ||
		utf8.RuneCountInString(summary) < 4 || utf8.RuneCountInString(summary) > 1000 ||
		utf8.RuneCountInString(usageNotes) > 2000 || len(labels) > 8 {
		return false
	}
	seen := map[string]struct{}{}
	for _, label := range labels {
		if label != strings.ToLower(strings.TrimSpace(label)) || !applicationDraftIdentifierPattern.MatchString(label) || utf8.RuneCountInString(label) > 40 {
			return false
		}
		if _, duplicate := seen[label]; duplicate {
			return false
		}
		seen[label] = struct{}{}
	}
	if !sort.StringsAreSorted(labels) {
		return false
	}
	return true
}

func validWorkflowTemplateReason(reason string) bool {
	reason = strings.TrimSpace(reason)
	return utf8.ValidString(reason) && utf8.RuneCountInString(reason) >= 4 && utf8.RuneCountInString(reason) <= 500 &&
		!applicationDraftStringContainsSecret(reason)
}

func validWorkflowTemplateReviewDecision(decision string) bool {
	switch strings.TrimSpace(decision) {
	case "approve", "reject", "request_changes", "withdraw":
		return true
	default:
		return false
	}
}

func workflowTemplateCandidateStateForDecision(decision string) string {
	switch decision {
	case "approve":
		return workflowTemplateCandidateApproved
	case "reject":
		return workflowTemplateCandidateRejected
	case "request_changes":
		return workflowTemplateCandidateChangesRequested
	default:
		return workflowTemplateCandidateWithdrawn
	}
}

func validWorkflowTemplateListingDecision(decision string) bool {
	switch strings.TrimSpace(decision) {
	case "list", "replace", "unlist":
		return true
	default:
		return false
	}
}

func cloneWorkflowTemplatePortability(value WorkflowTemplatePortabilitySummary) WorkflowTemplatePortabilitySummary {
	value.NodeKinds = cloneStringSlice(value.NodeKinds)
	value.ProviderRefs = cloneStringSlice(value.ProviderRefs)
	value.Blockers = cloneStringSlice(value.Blockers)
	return value
}

func cloneWorkflowTemplateCandidate(value WorkflowTemplateCandidate) WorkflowTemplateCandidate {
	value.Labels = cloneStringSlice(value.Labels)
	value.Portability = cloneWorkflowTemplatePortability(value.Portability)
	decisions := make([]WorkflowTemplateReviewDecision, len(value.Decisions))
	copy(decisions, value.Decisions)
	value.Decisions = decisions
	return value
}

func cloneWorkflowTemplateVersion(value WorkflowTemplateVersion) WorkflowTemplateVersion {
	value.Labels = cloneStringSlice(value.Labels)
	value.Portability = cloneWorkflowTemplatePortability(value.Portability)
	return value
}

func cloneWorkflowTemplateVersionPointer(value WorkflowTemplateVersion) *WorkflowTemplateVersion {
	cloned := cloneWorkflowTemplateVersion(value)
	return &cloned
}

func cloneWorkflowTemplateLineage(value WorkflowTemplateLineage) WorkflowTemplateLineage {
	events := make([]WorkflowTemplateListingEvent, len(value.Events))
	copy(events, value.Events)
	value.Events = events
	return value
}
