package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	adminProviderRouteDraftSchemaVersion       = "admin_provider_route_configuration_draft.v1"
	adminProviderRouteDraftSchemaVersionV2     = "admin_provider_route_configuration_draft.v2"
	adminProviderRouteCandidateSchemaVersion   = "admin_provider_route_candidate.v1"
	adminProviderRouteCandidateSchemaVersionV2 = "admin_provider_route_candidate.v2"
	adminProviderRouteReviewSchemaVersion      = "admin_provider_route_review.v1"
	adminProviderRouteSnapshotSchemaVersion    = "admin_provider_route_snapshot.v1"
	adminProviderRouteSnapshotSchemaVersionV2  = "admin_provider_route_snapshot.v2"
	adminProviderRouteActivationSchemaVersion  = "admin_provider_route_activation_record.v1"
)

const (
	AdminProviderRouteFailureDisabled                = "admin_provider_route_disabled"
	AdminProviderRouteFailureScopeDenied             = "admin_provider_route_scope_denied"
	AdminProviderRouteFailurePayloadInvalid          = "admin_provider_route_payload_invalid"
	AdminProviderRouteFailureSensitiveForbidden      = "admin_provider_route_sensitive_material_forbidden"
	AdminProviderRouteFailureEnvironmentForbidden    = "admin_provider_route_environment_forbidden"
	AdminProviderRouteFailureDraftNotFound           = "admin_provider_route_draft_not_found"
	AdminProviderRouteFailureDraftRevisionConflict   = "admin_provider_route_draft_revision_conflict"
	AdminProviderRouteFailureCandidateNotFound       = "admin_provider_route_candidate_not_found"
	AdminProviderRouteFailureCandidateConflict       = "admin_provider_route_candidate_conflict"
	AdminProviderRouteFailureReviewVersionConflict   = "admin_provider_route_review_version_conflict"
	AdminProviderRouteFailureReviewTransitionInvalid = "admin_provider_route_review_transition_invalid"
	AdminProviderRouteFailureCandidateNotApproved    = "admin_provider_route_candidate_not_approved"
	AdminProviderRouteFailureInventoryNotFound       = "admin_provider_route_inventory_not_found"
	AdminProviderRouteFailureInventoryMismatch       = "admin_provider_route_inventory_mismatch"
	AdminProviderRouteFailureInventoryUnavailable    = "admin_provider_route_inventory_unavailable"
	AdminProviderRouteFailureGenerationConflict      = "admin_provider_route_generation_conflict"
	AdminProviderRouteFailureAlreadyActive           = "admin_provider_route_already_active"
	AdminProviderRouteFailureRollbackTargetInvalid   = "admin_provider_route_rollback_target_invalid"
	AdminProviderRouteFailureStoreUnavailable        = "admin_provider_route_store_unavailable"
)

const (
	adminProviderRouteEnvironmentDevelopment = "development"
	adminProviderRouteEnvironmentTest        = "test"
	adminProviderRouteCandidatePending       = "pending_review"
	adminProviderRouteCandidateApproved      = "approved"
	adminProviderRouteCandidateRejected      = "rejected"
	adminProviderRouteDecisionApprove        = "approve"
	adminProviderRouteDecisionReject         = "reject"
	adminProviderRouteActionActivate         = "activate"
	adminProviderRouteActionRollback         = "rollback"
	adminProviderRouteMaxProfiles            = 32
	adminProviderRouteMaxRoutes              = 128
	adminProviderRouteMaxCapabilities        = 8
)

var (
	errAdminProviderRouteDraftNotFound         = errors.New(AdminProviderRouteFailureDraftNotFound)
	errAdminProviderRouteDraftConflict         = errors.New(AdminProviderRouteFailureDraftRevisionConflict)
	errAdminProviderRouteCandidateNotFound     = errors.New(AdminProviderRouteFailureCandidateNotFound)
	errAdminProviderRouteCandidateConflict     = errors.New(AdminProviderRouteFailureCandidateConflict)
	errAdminProviderRouteReviewConflict        = errors.New(AdminProviderRouteFailureReviewVersionConflict)
	errAdminProviderRouteReviewTransition      = errors.New(AdminProviderRouteFailureReviewTransitionInvalid)
	errAdminProviderRouteGenerationConflict    = errors.New(AdminProviderRouteFailureGenerationConflict)
	errAdminProviderRouteAlreadyActive         = errors.New(AdminProviderRouteFailureAlreadyActive)
	errAdminProviderRouteRollbackTargetInvalid = errors.New(AdminProviderRouteFailureRollbackTargetInvalid)
	errAdminProviderRouteStoreUnavailable      = errors.New(AdminProviderRouteFailureStoreUnavailable)
	errAdminProviderRouteInventoryNotFound     = errors.New(AdminProviderRouteFailureInventoryNotFound)
	errAdminProviderRouteInventoryUnavailable  = errors.New(AdminProviderRouteFailureInventoryUnavailable)

	adminProviderRouteIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$`)
	adminProviderRouteModelPattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,159}$`)
	adminProviderRouteDigestPattern     = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

type adminProviderRouteDraftConflictError struct {
	CurrentRevision int
}

func (err adminProviderRouteDraftConflictError) Error() string {
	return AdminProviderRouteFailureDraftRevisionConflict
}

func (err adminProviderRouteDraftConflictError) Is(target error) bool {
	return target == errAdminProviderRouteDraftConflict
}

type adminProviderRouteReviewConflictError struct {
	CurrentReviewVersion int
	CurrentState         string
}

func (err adminProviderRouteReviewConflictError) Error() string {
	return AdminProviderRouteFailureReviewVersionConflict
}

func (err adminProviderRouteReviewConflictError) Is(target error) bool {
	return target == errAdminProviderRouteReviewConflict
}

type adminProviderRouteGenerationConflictError struct {
	CurrentGeneration int
}

func (err adminProviderRouteGenerationConflictError) Error() string {
	return AdminProviderRouteFailureGenerationConflict
}

func (err adminProviderRouteGenerationConflictError) Is(target error) bool {
	return target == errAdminProviderRouteGenerationConflict
}

type AdminProviderRouteContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	Environment     string
	ActorRef        string
	AuditRef        string
	ReadEnabled     bool
	DraftEnabled    bool
	ReviewEnabled   bool
	ActivateEnabled bool
}

type AdminProviderProfileAssignment struct {
	ProfileID         string   `json:"profile_id"`
	DisplayName       string   `json:"display_name"`
	ProviderID        string   `json:"provider_id"`
	RuntimeProfileRef string   `json:"runtime_profile_ref"`
	Capabilities      []string `json:"capabilities"`
}

type AdminModelRouteDefinition struct {
	RouteID           string                            `json:"route_id"`
	Protocol          string                            `json:"protocol"`
	ModelID           string                            `json:"model_id"`
	ProviderProfileID string                            `json:"provider_profile_id,omitempty"`
	ExecutionMode     AdminProviderRouteExecutionMode   `json:"execution_mode,omitempty"`
	AttemptTargets    []AdminProviderRouteAttemptTarget `json:"attempt_targets,omitempty"`
}

type AdminProviderRouteConfigurationSnapshot struct {
	DisplayName      string                           `json:"display_name"`
	ProviderProfiles []AdminProviderProfileAssignment `json:"provider_profiles"`
	ModelRoutes      []AdminModelRouteDefinition      `json:"model_routes"`
}

type AdminProviderRouteConfigurationDraft struct {
	SchemaVersion     string                           `json:"schema_version"`
	TenantRef         string                           `json:"tenant_ref"`
	WorkspaceID       string                           `json:"workspace_id"`
	Environment       string                           `json:"environment"`
	ConfigurationID   string                           `json:"configuration_id"`
	DraftRevision     int                              `json:"draft_revision"`
	DisplayName       string                           `json:"display_name"`
	ProviderProfiles  []AdminProviderProfileAssignment `json:"provider_profiles"`
	ModelRoutes       []AdminModelRouteDefinition      `json:"model_routes"`
	DraftDigest       string                           `json:"draft_digest"`
	CreatedAt         string                           `json:"created_at"`
	UpdatedAt         string                           `json:"updated_at"`
	CreatedByActorRef string                           `json:"created_by_actor_ref"`
	UpdatedByActorRef string                           `json:"updated_by_actor_ref"`
	RequestID         string                           `json:"request_id"`
	AuditRef          string                           `json:"audit_ref"`
}

type AdminProviderInventoryBinding struct {
	ProfileID         string   `json:"profile_id"`
	ProviderID        string   `json:"provider_id"`
	RuntimeProfileRef string   `json:"runtime_profile_ref"`
	Environment       string   `json:"environment"`
	Capabilities      []string `json:"capabilities"`
	InventoryDigest   string   `json:"inventory_digest"`
	Enabled           bool     `json:"enabled"`
}

type AdminProviderRouteReview struct {
	SchemaVersion  string `json:"schema_version"`
	ReviewVersion  int    `json:"review_version"`
	Decision       string `json:"decision"`
	Reason         string `json:"reason"`
	ResultingState string `json:"resulting_state"`
	ReviewedAt     string `json:"reviewed_at"`
	ReviewerRef    string `json:"reviewer_ref"`
	RequestID      string `json:"request_id"`
	AuditRef       string `json:"audit_ref"`
}

type AdminProviderRouteCandidate struct {
	SchemaVersion       string                                  `json:"schema_version"`
	TenantRef           string                                  `json:"tenant_ref"`
	WorkspaceID         string                                  `json:"workspace_id"`
	Environment         string                                  `json:"environment"`
	ConfigurationID     string                                  `json:"configuration_id"`
	CandidateID         string                                  `json:"candidate_id"`
	SourceDraftRevision int                                     `json:"source_draft_revision"`
	SourceDraftDigest   string                                  `json:"source_draft_digest"`
	Configuration       AdminProviderRouteConfigurationSnapshot `json:"configuration"`
	InventoryBindings   []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	CandidateDigest     string                                  `json:"candidate_digest"`
	CandidateState      string                                  `json:"candidate_state"`
	ReviewVersion       int                                     `json:"review_version"`
	Review              *AdminProviderRouteReview               `json:"review,omitempty"`
	CreatedAt           string                                  `json:"created_at"`
	CreatedByActorRef   string                                  `json:"created_by_actor_ref"`
	RequestID           string                                  `json:"request_id"`
	AuditRef            string                                  `json:"audit_ref"`
}

type AdminProviderRouteSnapshot struct {
	SchemaVersion       string                                  `json:"schema_version"`
	TenantRef           string                                  `json:"tenant_ref"`
	WorkspaceID         string                                  `json:"workspace_id"`
	Environment         string                                  `json:"environment"`
	ConfigurationID     string                                  `json:"configuration_id"`
	Generation          int                                     `json:"generation"`
	CandidateID         string                                  `json:"candidate_id"`
	CandidateDigest     string                                  `json:"candidate_digest"`
	Configuration       AdminProviderRouteConfigurationSnapshot `json:"configuration"`
	InventoryBindings   []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	SnapshotDigest      string                                  `json:"snapshot_digest"`
	ActivatedAt         string                                  `json:"activated_at"`
	ActivatedByActorRef string                                  `json:"activated_by_actor_ref"`
	RequestID           string                                  `json:"request_id"`
	AuditRef            string                                  `json:"audit_ref"`
}

type AdminProviderRouteActivationRecord struct {
	SchemaVersion        string `json:"schema_version"`
	ActivationID         string `json:"activation_id"`
	ConfigurationID      string `json:"configuration_id"`
	Action               string `json:"action"`
	Reason               string `json:"reason"`
	BeforeGeneration     int    `json:"before_generation"`
	AfterGeneration      int    `json:"after_generation"`
	BeforeCandidateID    string `json:"before_candidate_id,omitempty"`
	BeforeSnapshotDigest string `json:"before_snapshot_digest,omitempty"`
	AfterCandidateID     string `json:"after_candidate_id"`
	AfterSnapshotDigest  string `json:"after_snapshot_digest"`
	PreviousRecordDigest string `json:"previous_record_digest,omitempty"`
	RecordDigest         string `json:"record_digest"`
	CreatedAt            string `json:"created_at"`
	ActorRef             string `json:"actor_ref"`
	RequestID            string `json:"request_id"`
	AuditRef             string `json:"audit_ref"`
}

type AdminProviderRouteDraftInput struct {
	ConfigurationID  string
	ExpectedRevision int
	DisplayName      string
	ProviderProfiles []AdminProviderProfileAssignment
	ModelRoutes      []AdminModelRouteDefinition
}

type AdminProviderRouteCandidateInput struct {
	ConfigurationID       string
	CandidateID           string
	ExpectedDraftRevision int
}

type AdminProviderRouteReviewInput struct {
	ExpectedReviewVersion int
	Decision              string
	Reason                string
}

type AdminProviderRouteActivationInput struct {
	ConfigurationID    string
	CandidateID        string
	ExpectedGeneration int
	Action             string
	Reason             string
}

type AdminProviderRouteResult struct {
	Draft                 *AdminProviderRouteConfigurationDraft
	Candidate             *AdminProviderRouteCandidate
	Snapshot              *AdminProviderRouteSnapshot
	Activation            *AdminProviderRouteActivationRecord
	ActivationHistory     []AdminProviderRouteActivationRecord
	FailureCode           string
	CurrentDraftRevision  int
	CurrentReviewVersion  int
	CurrentCandidateState string
	CurrentGeneration     int
}

type AdminProviderInventoryResolver interface {
	ResolveProviderProfile(context.Context, string, string, string) (AdminProviderInventoryBinding, error)
}

type adminProviderRouteRepository interface {
	PutDraft(AdminProviderRouteContext, int, AdminProviderRouteConfigurationDraft, time.Time) (AdminProviderRouteConfigurationDraft, error)
	ReadDraft(AdminProviderRouteContext, string) (AdminProviderRouteConfigurationDraft, error)
	CreateCandidate(AdminProviderRouteContext, AdminProviderRouteCandidate) (AdminProviderRouteCandidate, error)
	ReadCandidate(AdminProviderRouteContext, string, string) (AdminProviderRouteCandidate, error)
	ReviewCandidate(AdminProviderRouteContext, string, string, int, AdminProviderRouteReview) (AdminProviderRouteCandidate, error)
	CommitActivation(AdminProviderRouteContext, string, string, int, string, string, AdminProviderRouteSnapshot, time.Time) (AdminProviderRouteSnapshot, AdminProviderRouteActivationRecord, error)
	ReadActiveSnapshot(AdminProviderRouteContext, string) (AdminProviderRouteSnapshot, error)
	ListActivations(AdminProviderRouteContext, string) ([]AdminProviderRouteActivationRecord, error)
}

type adminProviderRouteService struct {
	repository        adminProviderRouteRepository
	inventoryResolver AdminProviderInventoryResolver
	now               func() time.Time
}

func newAdminProviderRouteService(repository adminProviderRouteRepository, resolver AdminProviderInventoryResolver) adminProviderRouteService {
	return adminProviderRouteService{
		repository: repository, inventoryResolver: resolver,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (service adminProviderRouteService) PutDraft(ctx AdminProviderRouteContext, input AdminProviderRouteDraftInput) AdminProviderRouteResult {
	if !ctx.DraftEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(input.ConfigurationID)) || input.ExpectedRevision < 0 {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	configuration, failureCode := normalizeAdminProviderRouteConfiguration(input.DisplayName, input.ProviderProfiles, input.ModelRoutes, ctx.Environment)
	if failureCode != "" {
		return AdminProviderRouteResult{FailureCode: failureCode}
	}
	digest, err := adminProviderRouteCanonicalDigest(configuration)
	if err != nil {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailurePayloadInvalid}
	}
	draft := AdminProviderRouteConfigurationDraft{
		SchemaVersion: adminProviderRouteDraftSchemaVersionForRoutes(configuration.ModelRoutes),
		TenantRef:     ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ConfigurationID: strings.TrimSpace(input.ConfigurationID), DisplayName: configuration.DisplayName,
		ProviderProfiles: configuration.ProviderProfiles, ModelRoutes: configuration.ModelRoutes,
		DraftDigest: digest, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	stored, err := service.repository.PutDraft(ctx, input.ExpectedRevision, draft, service.now())
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteDraftIntegrity(ctx, stored) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{Draft: &stored, CurrentDraftRevision: stored.DraftRevision}
}

func (service adminProviderRouteService) ReadDraft(ctx AdminProviderRouteContext, configurationID string) AdminProviderRouteResult {
	if !ctx.ReadEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(configurationID)) {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	draft, err := service.repository.ReadDraft(ctx, strings.TrimSpace(configurationID))
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteDraftIntegrity(ctx, draft) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{Draft: &draft, CurrentDraftRevision: draft.DraftRevision}
}

func (service adminProviderRouteService) CreateCandidate(ctx AdminProviderRouteContext, input AdminProviderRouteCandidateInput) AdminProviderRouteResult {
	if !ctx.DraftEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	configurationID := strings.TrimSpace(input.ConfigurationID)
	candidateID := strings.TrimSpace(input.CandidateID)
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(configurationID) ||
		!adminProviderRouteIdentifierPattern.MatchString(candidateID) || input.ExpectedDraftRevision < 1 {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	draft, err := service.repository.ReadDraft(ctx, configurationID)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if draft.DraftRevision != input.ExpectedDraftRevision {
		return AdminProviderRouteResult{
			FailureCode:          AdminProviderRouteFailureDraftRevisionConflict,
			CurrentDraftRevision: draft.DraftRevision,
		}
	}
	configuration := adminProviderRouteConfigurationFromDraft(draft)
	draftDigest, err := adminProviderRouteCanonicalDigest(configuration)
	if err != nil || draftDigest != draft.DraftDigest {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	bindings, failureCode := service.resolveInventory(ctx, configuration.ProviderProfiles)
	if failureCode != "" {
		return AdminProviderRouteResult{FailureCode: failureCode}
	}
	candidateDigest, err := adminProviderRouteCandidateDigest(configurationID, draft.DraftRevision, draft.DraftDigest, configuration, bindings)
	if err != nil {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailurePayloadInvalid}
	}
	candidate := AdminProviderRouteCandidate{
		SchemaVersion: adminProviderRouteCandidateSchemaVersionForRoutes(configuration.ModelRoutes),
		TenantRef:     ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ConfigurationID: configurationID, CandidateID: candidateID,
		SourceDraftRevision: draft.DraftRevision, SourceDraftDigest: draft.DraftDigest,
		Configuration: configuration, InventoryBindings: bindings, CandidateDigest: candidateDigest,
		CandidateState: adminProviderRouteCandidatePending, ReviewVersion: 0,
		CreatedAt: service.now().Format(time.RFC3339Nano), CreatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	created, err := service.repository.CreateCandidate(ctx, candidate)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteCandidateIntegrity(created) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{Candidate: &created, CurrentCandidateState: created.CandidateState}
}

func (service adminProviderRouteService) ReadCandidate(ctx AdminProviderRouteContext, configurationID, candidateID string) AdminProviderRouteResult {
	if !ctx.ReadEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(configurationID)) ||
		!adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(candidateID)) {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	candidate, err := service.repository.ReadCandidate(ctx, strings.TrimSpace(configurationID), strings.TrimSpace(candidateID))
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteCandidateIntegrity(candidate) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{
		Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion,
		CurrentCandidateState: candidate.CandidateState,
	}
}

func (service adminProviderRouteService) ReviewCandidate(ctx AdminProviderRouteContext, configurationID, candidateID string, input AdminProviderRouteReviewInput) AdminProviderRouteResult {
	if !ctx.ReviewEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	configurationID = strings.TrimSpace(configurationID)
	candidateID = strings.TrimSpace(candidateID)
	decision := strings.TrimSpace(input.Decision)
	reason := strings.TrimSpace(input.Reason)
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(configurationID) ||
		!adminProviderRouteIdentifierPattern.MatchString(candidateID) || input.ExpectedReviewVersion < 0 ||
		(decision != adminProviderRouteDecisionApprove && decision != adminProviderRouteDecisionReject) ||
		!validAdminProviderRouteReason(reason) {
		return AdminProviderRouteResult{FailureCode: adminProviderRoutePayloadFailure(reason)}
	}
	current, err := service.repository.ReadCandidate(ctx, configurationID, candidateID)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteCandidateIntegrity(current) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	state := adminProviderRouteCandidateRejected
	if decision == adminProviderRouteDecisionApprove {
		state = adminProviderRouteCandidateApproved
	}
	review := AdminProviderRouteReview{
		SchemaVersion: adminProviderRouteReviewSchemaVersion,
		ReviewVersion: input.ExpectedReviewVersion + 1,
		Decision:      decision, Reason: reason, ResultingState: state,
		ReviewedAt: service.now().Format(time.RFC3339Nano), ReviewerRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	candidate, err := service.repository.ReviewCandidate(ctx, configurationID, candidateID, input.ExpectedReviewVersion, review)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteCandidateIntegrity(candidate) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{
		Candidate: &candidate, CurrentReviewVersion: candidate.ReviewVersion,
		CurrentCandidateState: candidate.CandidateState,
	}
}

func (service adminProviderRouteService) Activate(ctx AdminProviderRouteContext, input AdminProviderRouteActivationInput) AdminProviderRouteResult {
	if !ctx.ActivateEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	configurationID := strings.TrimSpace(input.ConfigurationID)
	candidateID := strings.TrimSpace(input.CandidateID)
	action := strings.TrimSpace(input.Action)
	reason := strings.TrimSpace(input.Reason)
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(configurationID) ||
		!adminProviderRouteIdentifierPattern.MatchString(candidateID) || input.ExpectedGeneration < 0 ||
		(action != adminProviderRouteActionActivate && action != adminProviderRouteActionRollback) ||
		!validAdminProviderRouteReason(reason) {
		return AdminProviderRouteResult{FailureCode: adminProviderRoutePayloadFailure(reason)}
	}
	candidate, err := service.repository.ReadCandidate(ctx, configurationID, candidateID)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if candidate.CandidateState != adminProviderRouteCandidateApproved || candidate.Review == nil {
		return AdminProviderRouteResult{
			FailureCode:          AdminProviderRouteFailureCandidateNotApproved,
			CurrentReviewVersion: candidate.ReviewVersion, CurrentCandidateState: candidate.CandidateState,
		}
	}
	if !validAdminProviderRouteCandidateIntegrity(candidate) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	bindings, failureCode := service.resolveInventory(ctx, candidate.Configuration.ProviderProfiles)
	if failureCode != "" {
		return AdminProviderRouteResult{FailureCode: failureCode}
	}
	if !adminProviderRouteBindingsEqual(bindings, candidate.InventoryBindings) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureInventoryMismatch}
	}
	snapshotDigest, err := adminProviderRouteSnapshotDigest(candidate, bindings)
	if err != nil {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	snapshot := AdminProviderRouteSnapshot{
		SchemaVersion: adminProviderRouteSnapshotSchemaVersionForRoutes(candidate.Configuration.ModelRoutes),
		TenantRef:     ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
		ConfigurationID: configurationID, CandidateID: candidate.CandidateID,
		CandidateDigest: candidate.CandidateDigest, Configuration: cloneAdminProviderRouteConfiguration(candidate.Configuration),
		InventoryBindings: cloneAdminProviderInventoryBindings(bindings), SnapshotDigest: snapshotDigest,
		ActivatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	active, activation, err := service.repository.CommitActivation(
		ctx, configurationID, candidateID, input.ExpectedGeneration, action, reason, snapshot, service.now(),
	)
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteSnapshotIntegrity(ctx, active) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{
		Snapshot: &active, Activation: &activation, CurrentGeneration: active.Generation,
	}
}

func (service adminProviderRouteService) ReadActiveSnapshot(ctx AdminProviderRouteContext, configurationID string) AdminProviderRouteResult {
	if !ctx.ReadEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(configurationID)) {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	snapshot, err := service.repository.ReadActiveSnapshot(ctx, strings.TrimSpace(configurationID))
	if errors.Is(err, errAdminProviderRouteCandidateNotFound) {
		return AdminProviderRouteResult{CurrentGeneration: 0}
	}
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteSnapshotIntegrity(ctx, snapshot) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{Snapshot: &snapshot, CurrentGeneration: snapshot.Generation}
}

func (service adminProviderRouteService) ListActivations(ctx AdminProviderRouteContext, configurationID string) AdminProviderRouteResult {
	if !ctx.ReadEnabled {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureScopeDenied}
	}
	if !validAdminProviderRouteContext(ctx) || !adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(configurationID)) {
		return AdminProviderRouteResult{FailureCode: adminProviderRouteContextFailure(ctx)}
	}
	history, err := service.repository.ListActivations(ctx, strings.TrimSpace(configurationID))
	if err != nil {
		return adminProviderRouteRepositoryFailure(err)
	}
	if !validAdminProviderRouteActivationHistory(strings.TrimSpace(configurationID), history) {
		return AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	}
	return AdminProviderRouteResult{ActivationHistory: history}
}

func (service adminProviderRouteService) resolveInventory(ctx AdminProviderRouteContext, profiles []AdminProviderProfileAssignment) ([]AdminProviderInventoryBinding, string) {
	if service.inventoryResolver == nil {
		return nil, AdminProviderRouteFailureInventoryUnavailable
	}
	bindings := make([]AdminProviderInventoryBinding, 0, len(profiles))
	for _, profile := range profiles {
		binding, err := service.inventoryResolver.ResolveProviderProfile(
			ctx.RequestContext, ctx.Environment, profile.ProviderID, profile.RuntimeProfileRef,
		)
		if errors.Is(err, errAdminProviderRouteInventoryNotFound) {
			return nil, AdminProviderRouteFailureInventoryNotFound
		}
		if err != nil {
			return nil, AdminProviderRouteFailureInventoryUnavailable
		}
		binding.ProfileID = profile.ProfileID
		normalized, ok := normalizeAdminProviderInventoryBinding(binding)
		if !ok || !normalized.Enabled || normalized.ProviderID != profile.ProviderID ||
			normalized.RuntimeProfileRef != profile.RuntimeProfileRef || normalized.Environment != ctx.Environment ||
			!adminProviderRouteCapabilitiesContain(normalized.Capabilities, profile.Capabilities) {
			return nil, AdminProviderRouteFailureInventoryMismatch
		}
		bindings = append(bindings, normalized)
	}
	sort.Slice(bindings, func(i, j int) bool { return bindings[i].ProfileID < bindings[j].ProfileID })
	return bindings, ""
}

func normalizeAdminProviderRouteConfiguration(
	displayName string,
	profiles []AdminProviderProfileAssignment,
	routes []AdminModelRouteDefinition,
	environment string,
) (AdminProviderRouteConfigurationSnapshot, string) {
	displayName = strings.TrimSpace(displayName)
	if !validAdminProviderRouteDisplayName(displayName) || len(profiles) < 1 || len(profiles) > adminProviderRouteMaxProfiles ||
		len(routes) < 1 || len(routes) > adminProviderRouteMaxRoutes {
		return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
	}
	normalizedProfiles := make([]AdminProviderProfileAssignment, 0, len(profiles))
	profilesByID := make(map[string]AdminProviderProfileAssignment, len(profiles))
	for _, profile := range profiles {
		normalized, failureCode := normalizeAdminProviderProfile(profile, environment)
		if failureCode != "" {
			return AdminProviderRouteConfigurationSnapshot{}, failureCode
		}
		if _, exists := profilesByID[normalized.ProfileID]; exists {
			return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
		}
		profilesByID[normalized.ProfileID] = normalized
		normalizedProfiles = append(normalizedProfiles, normalized)
	}
	normalizedRoutes := make([]AdminModelRouteDefinition, 0, len(routes))
	routeIDs := make(map[string]bool, len(routes))
	routeBindings := make(map[string]bool, len(routes))
	for _, route := range routes {
		normalized, failureCode := normalizeAdminModelRoute(route)
		if failureCode != "" {
			return AdminProviderRouteConfigurationSnapshot{}, failureCode
		}
		profileIDs := adminProviderRouteProfileIDs(normalized)
		for _, profileID := range profileIDs {
			profile, exists := profilesByID[profileID]
			if !exists || !adminProviderRouteCapabilitiesContain(profile.Capabilities, []string{normalized.Protocol}) {
				return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
			}
		}
		bindingKey := normalized.Protocol + "\x00" + normalized.ModelID
		if routeIDs[normalized.RouteID] || routeBindings[bindingKey] {
			return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
		}
		routeIDs[normalized.RouteID] = true
		routeBindings[bindingKey] = true
		normalizedRoutes = append(normalizedRoutes, normalized)
	}
	if adminProviderRouteContractVersion(normalizedRoutes) == 0 {
		return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
	}
	sort.Slice(normalizedProfiles, func(i, j int) bool { return normalizedProfiles[i].ProfileID < normalizedProfiles[j].ProfileID })
	sort.Slice(normalizedRoutes, func(i, j int) bool {
		if normalizedRoutes[i].Protocol != normalizedRoutes[j].Protocol {
			return normalizedRoutes[i].Protocol < normalizedRoutes[j].Protocol
		}
		if normalizedRoutes[i].ModelID != normalizedRoutes[j].ModelID {
			return normalizedRoutes[i].ModelID < normalizedRoutes[j].ModelID
		}
		return normalizedRoutes[i].RouteID < normalizedRoutes[j].RouteID
	})
	configuration := AdminProviderRouteConfigurationSnapshot{
		DisplayName: displayName, ProviderProfiles: normalizedProfiles, ModelRoutes: normalizedRoutes,
	}
	payload, err := json.Marshal(configuration)
	if err != nil || len(payload) > 64*1024 {
		return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailurePayloadInvalid
	}
	if adminProviderRouteContainsSensitiveMaterial(string(payload)) {
		return AdminProviderRouteConfigurationSnapshot{}, AdminProviderRouteFailureSensitiveForbidden
	}
	return configuration, ""
}

func normalizeAdminProviderProfile(profile AdminProviderProfileAssignment, environment string) (AdminProviderProfileAssignment, string) {
	profile.ProfileID = strings.TrimSpace(profile.ProfileID)
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.ProviderID = strings.TrimSpace(profile.ProviderID)
	profile.RuntimeProfileRef = strings.TrimSpace(profile.RuntimeProfileRef)
	if !adminProviderRouteIdentifierPattern.MatchString(profile.ProfileID) ||
		!validAdminProviderRouteDisplayName(profile.DisplayName) ||
		!adminProviderRouteIdentifierPattern.MatchString(profile.ProviderID) {
		return AdminProviderProfileAssignment{}, AdminProviderRouteFailurePayloadInvalid
	}
	expectedPrefix := "ref:radishmind/" + environment + "/provider-profiles/"
	if !strings.HasPrefix(profile.RuntimeProfileRef, expectedPrefix) {
		if strings.HasPrefix(profile.RuntimeProfileRef, "ref:radishmind/") {
			return AdminProviderProfileAssignment{}, AdminProviderRouteFailureEnvironmentForbidden
		}
		return AdminProviderProfileAssignment{}, AdminProviderRouteFailurePayloadInvalid
	}
	refKey := strings.TrimPrefix(profile.RuntimeProfileRef, expectedPrefix)
	if !adminProviderRouteIdentifierPattern.MatchString(refKey) || strings.Contains(profile.RuntimeProfileRef, "..") {
		return AdminProviderProfileAssignment{}, AdminProviderRouteFailurePayloadInvalid
	}
	capabilities, ok := normalizeAdminProviderRouteCapabilities(profile.Capabilities)
	if !ok {
		return AdminProviderProfileAssignment{}, AdminProviderRouteFailurePayloadInvalid
	}
	profile.Capabilities = capabilities
	return profile, ""
}

func normalizeAdminModelRoute(route AdminModelRouteDefinition) (AdminModelRouteDefinition, string) {
	return normalizeAdminModelRouteContract(route)
}

func normalizeAdminProviderRouteCapabilities(values []string) ([]string, bool) {
	if len(values) < 1 || len(values) > adminProviderRouteMaxCapabilities {
		return nil, false
	}
	seen := make(map[string]bool, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if !isApplicationDraftProtocol(value) || seen[value] {
			return nil, false
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized, true
}

func normalizeAdminProviderInventoryBinding(binding AdminProviderInventoryBinding) (AdminProviderInventoryBinding, bool) {
	binding.ProfileID = strings.TrimSpace(binding.ProfileID)
	binding.ProviderID = strings.TrimSpace(binding.ProviderID)
	binding.RuntimeProfileRef = strings.TrimSpace(binding.RuntimeProfileRef)
	binding.Environment = strings.TrimSpace(binding.Environment)
	binding.InventoryDigest = strings.TrimSpace(binding.InventoryDigest)
	capabilities, ok := normalizeAdminProviderRouteCapabilities(binding.Capabilities)
	if !ok || !adminProviderRouteIdentifierPattern.MatchString(binding.ProfileID) ||
		!adminProviderRouteIdentifierPattern.MatchString(binding.ProviderID) ||
		!adminProviderRouteDigestPattern.MatchString(binding.InventoryDigest) {
		return AdminProviderInventoryBinding{}, false
	}
	binding.Capabilities = capabilities
	return binding, true
}

func adminProviderRouteCanonicalDigest(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func adminProviderRouteCandidateDigest(
	configurationID string,
	revision int,
	draftDigest string,
	configuration AdminProviderRouteConfigurationSnapshot,
	bindings []AdminProviderInventoryBinding,
) (string, error) {
	return adminProviderRouteCanonicalDigest(struct {
		ConfigurationID string                                  `json:"configuration_id"`
		DraftRevision   int                                     `json:"draft_revision"`
		DraftDigest     string                                  `json:"draft_digest"`
		Configuration   AdminProviderRouteConfigurationSnapshot `json:"configuration"`
		Bindings        []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	}{
		ConfigurationID: configurationID, DraftRevision: revision, DraftDigest: draftDigest,
		Configuration: configuration, Bindings: bindings,
	})
}

func adminProviderRouteSnapshotDigest(candidate AdminProviderRouteCandidate, bindings []AdminProviderInventoryBinding) (string, error) {
	return adminProviderRouteCanonicalDigest(struct {
		ConfigurationID string                                  `json:"configuration_id"`
		CandidateID     string                                  `json:"candidate_id"`
		CandidateDigest string                                  `json:"candidate_digest"`
		Configuration   AdminProviderRouteConfigurationSnapshot `json:"configuration"`
		Bindings        []AdminProviderInventoryBinding         `json:"inventory_bindings"`
	}{
		ConfigurationID: candidate.ConfigurationID, CandidateID: candidate.CandidateID,
		CandidateDigest: candidate.CandidateDigest, Configuration: candidate.Configuration, Bindings: bindings,
	})
}

func adminProviderRouteRepositoryFailure(err error) AdminProviderRouteResult {
	result := AdminProviderRouteResult{FailureCode: AdminProviderRouteFailureStoreUnavailable}
	var draftConflict adminProviderRouteDraftConflictError
	var reviewConflict adminProviderRouteReviewConflictError
	var generationConflict adminProviderRouteGenerationConflictError
	switch {
	case errors.As(err, &draftConflict):
		result.FailureCode = AdminProviderRouteFailureDraftRevisionConflict
		result.CurrentDraftRevision = draftConflict.CurrentRevision
	case errors.Is(err, errAdminProviderRouteDraftNotFound):
		result.FailureCode = AdminProviderRouteFailureDraftNotFound
	case errors.Is(err, errAdminProviderRouteCandidateNotFound):
		result.FailureCode = AdminProviderRouteFailureCandidateNotFound
	case errors.Is(err, errAdminProviderRouteCandidateConflict):
		result.FailureCode = AdminProviderRouteFailureCandidateConflict
	case errors.As(err, &reviewConflict):
		result.FailureCode = AdminProviderRouteFailureReviewVersionConflict
		result.CurrentReviewVersion = reviewConflict.CurrentReviewVersion
		result.CurrentCandidateState = reviewConflict.CurrentState
	case errors.Is(err, errAdminProviderRouteReviewTransition):
		result.FailureCode = AdminProviderRouteFailureReviewTransitionInvalid
	case errors.As(err, &generationConflict):
		result.FailureCode = AdminProviderRouteFailureGenerationConflict
		result.CurrentGeneration = generationConflict.CurrentGeneration
	case errors.Is(err, errAdminProviderRouteAlreadyActive):
		result.FailureCode = AdminProviderRouteFailureAlreadyActive
	case errors.Is(err, errAdminProviderRouteRollbackTargetInvalid):
		result.FailureCode = AdminProviderRouteFailureRollbackTargetInvalid
	}
	return result
}

func validAdminProviderRouteContext(ctx AdminProviderRouteContext) bool {
	return ctx.RequestContext != nil &&
		validAdminProviderRouteScopeToken(ctx.TenantRef, 160) &&
		adminProviderRouteIdentifierPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) &&
		(ctx.Environment == adminProviderRouteEnvironmentDevelopment || ctx.Environment == adminProviderRouteEnvironmentTest) &&
		validAdminProviderRouteScopeToken(ctx.ActorRef, 256) &&
		validAdminProviderRouteScopeToken(ctx.RequestID, 256) &&
		validAdminProviderRouteScopeToken(ctx.AuditRef, 256)
}

func adminProviderRouteContextFailure(ctx AdminProviderRouteContext) string {
	if ctx.Environment != adminProviderRouteEnvironmentDevelopment && ctx.Environment != adminProviderRouteEnvironmentTest {
		return AdminProviderRouteFailureEnvironmentForbidden
	}
	return AdminProviderRouteFailurePayloadInvalid
}

func validAdminProviderRouteDisplayName(value string) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= 2 && utf8.RuneCountInString(value) <= 80
}

func validAdminProviderRouteScopeToken(value string, maxRunes int) bool {
	value = strings.TrimSpace(value)
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= 1 &&
		utf8.RuneCountInString(value) <= maxRunes && !strings.ContainsRune(value, '\x00')
}

func validAdminProviderRouteReason(value string) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= 4 &&
		utf8.RuneCountInString(value) <= 500 && !adminProviderRouteContainsSensitiveMaterial(value)
}

func adminProviderRoutePayloadFailure(value string) string {
	if adminProviderRouteContainsSensitiveMaterial(value) {
		return AdminProviderRouteFailureSensitiveForbidden
	}
	return AdminProviderRouteFailurePayloadInvalid
}

func adminProviderRouteContainsSensitiveMaterial(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{
		"http://", "https://", "authorization", "bearer ", "api_key", "api-key",
		"password", "passwd", "token=", "cookie", "dsn=", "database_url",
		"-----begin", "sk-", "secret=", "credential=",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func adminProviderRouteCapabilitiesContain(available, required []string) bool {
	values := make(map[string]bool, len(available))
	for _, capability := range available {
		values[capability] = true
	}
	for _, capability := range required {
		if !values[capability] {
			return false
		}
	}
	return true
}

func adminProviderRouteBindingsEqual(left, right []AdminProviderInventoryBinding) bool {
	leftPayload, leftErr := json.Marshal(left)
	rightPayload, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftPayload) == string(rightPayload)
}

func adminProviderRouteConfigurationFromDraft(draft AdminProviderRouteConfigurationDraft) AdminProviderRouteConfigurationSnapshot {
	return AdminProviderRouteConfigurationSnapshot{
		DisplayName:      draft.DisplayName,
		ProviderProfiles: cloneAdminProviderProfiles(draft.ProviderProfiles),
		ModelRoutes:      cloneAdminModelRoutes(draft.ModelRoutes),
	}
}

func cloneAdminProviderRouteConfiguration(value AdminProviderRouteConfigurationSnapshot) AdminProviderRouteConfigurationSnapshot {
	value.ProviderProfiles = cloneAdminProviderProfiles(value.ProviderProfiles)
	value.ModelRoutes = cloneAdminModelRoutes(value.ModelRoutes)
	return value
}

func cloneAdminProviderProfiles(values []AdminProviderProfileAssignment) []AdminProviderProfileAssignment {
	cloned := append([]AdminProviderProfileAssignment(nil), values...)
	for index := range cloned {
		cloned[index].Capabilities = cloneStringSlice(cloned[index].Capabilities)
	}
	return cloned
}

func cloneAdminModelRoutes(values []AdminModelRouteDefinition) []AdminModelRouteDefinition {
	cloned := append([]AdminModelRouteDefinition(nil), values...)
	for index := range cloned {
		cloned[index].AttemptTargets = append([]AdminProviderRouteAttemptTarget(nil), cloned[index].AttemptTargets...)
	}
	return cloned
}

func cloneAdminProviderInventoryBindings(values []AdminProviderInventoryBinding) []AdminProviderInventoryBinding {
	cloned := append([]AdminProviderInventoryBinding(nil), values...)
	for index := range cloned {
		cloned[index].Capabilities = cloneStringSlice(cloned[index].Capabilities)
	}
	return cloned
}

func cloneAdminProviderRouteDraft(value AdminProviderRouteConfigurationDraft) AdminProviderRouteConfigurationDraft {
	value.ProviderProfiles = cloneAdminProviderProfiles(value.ProviderProfiles)
	value.ModelRoutes = cloneAdminModelRoutes(value.ModelRoutes)
	return value
}

func cloneAdminProviderRouteCandidate(value AdminProviderRouteCandidate) AdminProviderRouteCandidate {
	value.Configuration = cloneAdminProviderRouteConfiguration(value.Configuration)
	value.InventoryBindings = cloneAdminProviderInventoryBindings(value.InventoryBindings)
	if value.Review != nil {
		review := *value.Review
		value.Review = &review
	}
	return value
}

func cloneAdminProviderRouteSnapshot(value AdminProviderRouteSnapshot) AdminProviderRouteSnapshot {
	value.Configuration = cloneAdminProviderRouteConfiguration(value.Configuration)
	value.InventoryBindings = cloneAdminProviderInventoryBindings(value.InventoryBindings)
	return value
}
