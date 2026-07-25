package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	AgentCopilotProfileFailureScopeDenied         = "agent_copilot_profile_scope_denied"
	AgentCopilotProfileFailureNotFound            = "agent_copilot_profile_not_found"
	AgentCopilotProfileFailureVersionConflict     = "agent_copilot_profile_version_conflict"
	AgentCopilotProfileFailureStoreUnavailable    = "agent_copilot_profile_store_unavailable"
	AgentCopilotProfileFailureWriteDisabled       = "agent_copilot_profile_write_disabled"
	AgentCopilotProfileFailureImmutableConflict   = "agent_copilot_profile_immutable_conflict"
	AgentCopilotProfileFailureDigestDrift         = "agent_copilot_profile_digest_drift"
	AgentCopilotProfileFailureBindingIneligible   = "agent_copilot_profile_binding_ineligible"
	AgentCopilotProfileFailureApplicationMissing  = "agent_copilot_profile_application_not_found"
	AgentCopilotProfileFailureApplicationArchived = "agent_copilot_profile_application_archived"
	AgentCopilotProfileFailureApplicationKind     = "agent_copilot_profile_application_kind_mismatch"
)

var (
	errAgentCopilotProfileNotFound  = errors.New(AgentCopilotProfileFailureNotFound)
	errAgentCopilotProfileConflict  = errors.New(AgentCopilotProfileFailureVersionConflict)
	errAgentCopilotProfileStore     = errors.New(AgentCopilotProfileFailureStoreUnavailable)
	errAgentCopilotProfileImmutable = errors.New(AgentCopilotProfileFailureImmutableConflict)
	errAgentCopilotProfileDigest    = errors.New(AgentCopilotProfileFailureDigestDrift)
	errAgentCopilotProfileOwner     = errors.New("agent copilot profile owner contract mismatch")
)

type agentCopilotProfileVersionConflictError struct {
	CurrentDraftVersion int
}

func (failure agentCopilotProfileVersionConflictError) Error() string {
	return AgentCopilotProfileFailureVersionConflict
}

func (failure agentCopilotProfileVersionConflictError) Is(target error) bool {
	return target == errAgentCopilotProfileConflict
}

type AgentCopilotProfileContext struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	ActorRef        string
	OwnerSubjectRef string
	AuditRef        string
	WriteEnabled    bool
}

type AgentCopilotProfileDraftInput struct {
	SchemaVersion string `json:"schema_version"`
	ProfileID     string `json:"profile_id"`
	WorkspaceID   string `json:"workspace_id"`
	ApplicationID string `json:"application_id"`
	AgentCopilotProfileSource
}

type AgentCopilotProfileDraftSummary struct {
	SchemaVersion      string   `json:"schema_version"`
	ProfileID          string   `json:"profile_id"`
	ApplicationID      string   `json:"application_id"`
	ProfileName        string   `json:"profile_name"`
	Description        string   `json:"description"`
	Project            string   `json:"project"`
	AllowedTasks       []string `json:"allowed_tasks"`
	DefaultLocale      string   `json:"default_locale"`
	DraftVersion       int      `json:"draft_version"`
	ProfileDigest      string   `json:"profile_digest"`
	PolicyDigest       string   `json:"policy_digest"`
	AllowedTasksDigest string   `json:"allowed_tasks_digest"`
	ValidationState    string   `json:"validation_state"`
	UpdatedAt          string   `json:"updated_at"`
	UpdatedByActorRef  string   `json:"updated_by_actor_ref"`
}

type AgentCopilotProfileVersionSummary struct {
	SchemaVersion       string `json:"schema_version"`
	ProfileID           string `json:"profile_id"`
	ProfileVersion      int    `json:"profile_version"`
	SourceDraftVersion  int    `json:"source_draft_version"`
	ProfileName         string `json:"profile_name"`
	Project             string `json:"project"`
	DefaultLocale       string `json:"default_locale"`
	ProfileDigest       string `json:"profile_digest"`
	PolicyDigest        string `json:"policy_digest"`
	AllowedTasksDigest  string `json:"allowed_tasks_digest"`
	PublishedAt         string `json:"published_at"`
	PublishedByActorRef string `json:"published_by_actor_ref"`
}

type AgentCopilotProfileResult struct {
	Draft                 *AgentCopilotProfileDraftV1
	Version               *AgentCopilotProfileVersionV1
	ValidationSummary     ApplicationConfigurationDraftValidation
	FailureCode           string
	CurrentDraftVersion   int
	CurrentProfileVersion int
}

type agentCopilotProfileRepository interface {
	SaveDraft(AgentCopilotProfileContext, AgentCopilotProfileDraftV1, int) (AgentCopilotProfileDraftV1, error)
	ReadDraft(AgentCopilotProfileContext, string) (AgentCopilotProfileDraftV1, error)
	ListDrafts(AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, error)
	CreateVersion(AgentCopilotProfileContext, AgentCopilotProfileVersionV1) (AgentCopilotProfileVersionV1, error)
	ReadVersion(AgentCopilotProfileContext, string, int) (AgentCopilotProfileVersionV1, error)
	ListVersions(AgentCopilotProfileContext, string) ([]AgentCopilotProfileVersionSummary, error)
}

type memoryAgentCopilotProfileRepository struct {
	mu          sync.RWMutex
	drafts      map[string]AgentCopilotProfileDraftV1
	versions    map[string]map[int]AgentCopilotProfileVersionV1
	unavailable bool
}

type agentCopilotProfileService struct {
	repository              agentCopilotProfileRepository
	requireAgentApplication func(AgentCopilotProfileContext) string
	now                     func() time.Time
}

func newMemoryAgentCopilotProfileRepository() *memoryAgentCopilotProfileRepository {
	return &memoryAgentCopilotProfileRepository{
		drafts:   make(map[string]AgentCopilotProfileDraftV1),
		versions: make(map[string]map[int]AgentCopilotProfileVersionV1),
	}
}

func newAgentCopilotProfileService(repository agentCopilotProfileRepository) agentCopilotProfileService {
	return agentCopilotProfileService{repository: repository, now: func() time.Time { return time.Now().UTC() }}
}

func (service agentCopilotProfileService) Validate(ctx AgentCopilotProfileContext, input AgentCopilotProfileDraftInput) AgentCopilotProfileResult {
	_, validation := compileAgentCopilotProfileDraftInput(ctx, input)
	return AgentCopilotProfileResult{ValidationSummary: validation}
}

func (service agentCopilotProfileService) SaveDraft(ctx AgentCopilotProfileContext, input AgentCopilotProfileDraftInput, expectedVersion int) AgentCopilotProfileResult {
	compiled, validation := compileAgentCopilotProfileDraftInput(ctx, input)
	if !validation.IsValid {
		return AgentCopilotProfileResult{FailureCode: agentCopilotProfileValidationFailure(validation), ValidationSummary: validation}
	}
	if !ctx.WriteEnabled {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureWriteDisabled, ValidationSummary: validation}
	}
	if expectedVersion < 0 {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailurePayloadInvalid, ValidationSummary: validation}
	}
	if service.requireAgentApplication != nil {
		if failure := service.requireAgentApplication(ctx); failure != "" {
			return AgentCopilotProfileResult{FailureCode: failure, ValidationSummary: validation}
		}
	}
	now := service.now().UTC().Format(time.RFC3339Nano)
	draft := AgentCopilotProfileDraftV1{
		SchemaVersion: strings.TrimSpace(input.SchemaVersion), ProfileID: strings.TrimSpace(input.ProfileID),
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, AgentCopilotProfileSource: compiled.Source,
		DraftVersion: expectedVersion + 1, ProfileDigest: compiled.ProfileDigest, PolicyDigest: compiled.PolicyDigest,
		ValidationSummary: validation, CreatedAt: now, UpdatedAt: now,
		CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	saved, err := service.repository.SaveDraft(ctx, draft, expectedVersion)
	if err != nil {
		return agentCopilotProfileRepositoryFailure(err, validation)
	}
	return AgentCopilotProfileResult{Draft: &saved, ValidationSummary: saved.ValidationSummary, CurrentDraftVersion: saved.DraftVersion}
}

func (service agentCopilotProfileService) ReadDraft(ctx AgentCopilotProfileContext, profileID string) AgentCopilotProfileResult {
	if validateAgentCopilotProfileContext(ctx) != nil || !agentCopilotProfileIDPattern.MatchString(strings.TrimSpace(profileID)) {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureScopeDenied}
	}
	draft, err := service.repository.ReadDraft(ctx, strings.TrimSpace(profileID))
	if err != nil {
		return agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{})
	}
	return AgentCopilotProfileResult{Draft: &draft, ValidationSummary: draft.ValidationSummary, CurrentDraftVersion: draft.DraftVersion}
}

func (service agentCopilotProfileService) ListDrafts(ctx AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, string) {
	if validateAgentCopilotProfileContext(ctx) != nil {
		return []AgentCopilotProfileDraftSummary{}, AgentCopilotProfileFailureScopeDenied
	}
	summaries, err := service.repository.ListDrafts(ctx)
	if err != nil {
		return []AgentCopilotProfileDraftSummary{}, agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{}).FailureCode
	}
	return summaries, ""
}

func (service agentCopilotProfileService) CreateVersion(ctx AgentCopilotProfileContext, profileID string, sourceDraftVersion int) AgentCopilotProfileResult {
	if !ctx.WriteEnabled {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureWriteDisabled}
	}
	profileID = strings.TrimSpace(profileID)
	if validateAgentCopilotProfileContext(ctx) != nil || !agentCopilotProfileIDPattern.MatchString(profileID) || sourceDraftVersion < 1 {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailurePayloadInvalid}
	}
	if service.requireAgentApplication != nil {
		if failure := service.requireAgentApplication(ctx); failure != "" {
			return AgentCopilotProfileResult{FailureCode: failure}
		}
	}
	draft, err := service.repository.ReadDraft(ctx, profileID)
	if err != nil {
		return agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{})
	}
	if draft.DraftVersion != sourceDraftVersion {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureVersionConflict, CurrentDraftVersion: draft.DraftVersion}
	}
	version := AgentCopilotProfileVersionV1{
		SchemaVersion: agentCopilotProfileVersionSchema, ProfileID: profileID,
		SourceDraftVersion: draft.DraftVersion, TenantRef: draft.TenantRef, WorkspaceID: draft.WorkspaceID,
		ApplicationID: draft.ApplicationID, OwnerSubjectRef: draft.OwnerSubjectRef,
		AgentCopilotProfileSource: cloneAgentCopilotProfileSource(draft.AgentCopilotProfileSource),
		ProfileDigest:             draft.ProfileDigest, PolicyDigest: draft.PolicyDigest,
		PublishedAt: service.now().UTC().Format(time.RFC3339Nano), PublishedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	created, err := service.repository.CreateVersion(ctx, version)
	if err != nil {
		return agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{})
	}
	return AgentCopilotProfileResult{Version: &created, CurrentDraftVersion: draft.DraftVersion, CurrentProfileVersion: created.ProfileVersion}
}

func (service agentCopilotProfileService) ReadVersion(ctx AgentCopilotProfileContext, profileID string, profileVersion int) AgentCopilotProfileResult {
	profileID = strings.TrimSpace(profileID)
	if validateAgentCopilotProfileContext(ctx) != nil || !agentCopilotProfileIDPattern.MatchString(profileID) || profileVersion < 1 {
		return AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureScopeDenied}
	}
	version, err := service.repository.ReadVersion(ctx, profileID, profileVersion)
	if err != nil {
		return agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{})
	}
	return AgentCopilotProfileResult{Version: &version, CurrentProfileVersion: version.ProfileVersion}
}

func (service agentCopilotProfileService) ListVersions(ctx AgentCopilotProfileContext, profileID string) ([]AgentCopilotProfileVersionSummary, string) {
	profileID = strings.TrimSpace(profileID)
	if validateAgentCopilotProfileContext(ctx) != nil || !agentCopilotProfileIDPattern.MatchString(profileID) {
		return []AgentCopilotProfileVersionSummary{}, AgentCopilotProfileFailureScopeDenied
	}
	summaries, err := service.repository.ListVersions(ctx, profileID)
	if err != nil {
		return []AgentCopilotProfileVersionSummary{}, agentCopilotProfileRepositoryFailure(err, ApplicationConfigurationDraftValidation{}).FailureCode
	}
	return summaries, ""
}

func compileAgentCopilotProfileDraftInput(ctx AgentCopilotProfileContext, input AgentCopilotProfileDraftInput) (AgentCopilotCompiledProfile, ApplicationConfigurationDraftValidation) {
	findings := make([]ApplicationConfigurationDraftValidationFinding, 0)
	appendFinding := func(code, field, summary string) {
		findings = append(findings, ApplicationConfigurationDraftValidationFinding{Code: code, Field: field, Summary: summary})
	}
	if validateAgentCopilotProfileContext(ctx) != nil || strings.TrimSpace(input.WorkspaceID) != ctx.WorkspaceID || strings.TrimSpace(input.ApplicationID) != ctx.ApplicationID {
		appendFinding(AgentCopilotProfileFailureScopeDenied, "scope", "profile scope does not match the authenticated context")
	}
	if strings.TrimSpace(input.SchemaVersion) != agentCopilotProfileDraftSchema || !agentCopilotProfileIDPattern.MatchString(strings.TrimSpace(input.ProfileID)) {
		appendFinding(AgentCopilotProfileFailurePayloadInvalid, "schema_version", "profile schema version or identifier is invalid")
	}
	if !utf8.ValidString(input.ProfileName) || !utf8.ValidString(input.Description) {
		appendFinding(AgentCopilotProfileFailurePayloadInvalid, "source", "profile source must contain valid UTF-8 text")
	}
	compiled, sourceFindings := CompileAgentCopilotProfileSource(input.AgentCopilotProfileSource)
	for _, finding := range sourceFindings {
		appendFinding(finding.Code, finding.Field, finding.Summary)
	}
	state := applicationDraftValidationValid
	if len(findings) != 0 {
		state = applicationDraftValidationInvalid
	}
	return compiled, ApplicationConfigurationDraftValidation{State: state, IsValid: len(findings) == 0, Findings: findings}
}

func validateAgentCopilotProfileContext(ctx AgentCopilotProfileContext) error {
	if ctx.RequestContext == nil || !controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.TenantRef)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) ||
		!applicationCatalogIDPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.ActorRef)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.OwnerSubjectRef)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.RequestID)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.AuditRef)) {
		return errAgentCopilotProfileOwner
	}
	return nil
}

func validateStoredAgentCopilotProfileDraft(ctx AgentCopilotProfileContext, draft AgentCopilotProfileDraftV1) error {
	if draft.TenantRef != ctx.TenantRef || draft.WorkspaceID != ctx.WorkspaceID || draft.ApplicationID != ctx.ApplicationID ||
		draft.OwnerSubjectRef != ctx.OwnerSubjectRef || validateAgentCopilotProfileDraft(draft) != nil {
		compiled, findings := CompileAgentCopilotProfileSource(draft.AgentCopilotProfileSource)
		if len(findings) == 0 && (draft.ProfileDigest != compiled.ProfileDigest || draft.PolicyDigest != compiled.PolicyDigest) {
			return errAgentCopilotProfileDigest
		}
		return errAgentCopilotProfileOwner
	}
	return nil
}

func validateStoredAgentCopilotProfileVersion(ctx AgentCopilotProfileContext, version AgentCopilotProfileVersionV1) error {
	if version.TenantRef != ctx.TenantRef || version.WorkspaceID != ctx.WorkspaceID || version.ApplicationID != ctx.ApplicationID ||
		version.OwnerSubjectRef != ctx.OwnerSubjectRef || validateAgentCopilotProfileVersion(version) != nil {
		compiled, findings := CompileAgentCopilotProfileSource(version.AgentCopilotProfileSource)
		if len(findings) == 0 && (version.ProfileDigest != compiled.ProfileDigest || version.PolicyDigest != compiled.PolicyDigest) {
			return errAgentCopilotProfileDigest
		}
		return errAgentCopilotProfileOwner
	}
	return nil
}

func (repository *memoryAgentCopilotProfileRepository) SaveDraft(ctx AgentCopilotProfileContext, draft AgentCopilotProfileDraftV1, expectedVersion int) (AgentCopilotProfileDraftV1, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return AgentCopilotProfileDraftV1{}, err
	}
	key := agentCopilotProfileRepositoryKey(ctx, draft.ProfileID)
	current, exists := repository.drafts[key]
	if exists {
		if err := validateStoredAgentCopilotProfileDraft(ctx, current); err != nil {
			return AgentCopilotProfileDraftV1{}, err
		}
	}
	if (!exists && expectedVersion != 0) || (exists && current.DraftVersion != expectedVersion) {
		currentVersion := 0
		if exists {
			currentVersion = current.DraftVersion
		}
		return AgentCopilotProfileDraftV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: currentVersion}
	}
	if exists {
		draft.CreatedAt = current.CreatedAt
		draft.CreatedByActorRef = current.CreatedByActorRef
	}
	repository.drafts[key] = cloneAgentCopilotProfileDraft(draft)
	return cloneAgentCopilotProfileDraft(draft), nil
}

func (repository *memoryAgentCopilotProfileRepository) ReadDraft(ctx AgentCopilotProfileContext, profileID string) (AgentCopilotProfileDraftV1, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileStore
	}
	draft, exists := repository.drafts[agentCopilotProfileRepositoryKey(ctx, profileID)]
	if !exists {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileNotFound
	}
	if draft.ProfileID != strings.TrimSpace(profileID) {
		return AgentCopilotProfileDraftV1{}, errAgentCopilotProfileOwner
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return AgentCopilotProfileDraftV1{}, err
	}
	return cloneAgentCopilotProfileDraft(draft), nil
}

func (repository *memoryAgentCopilotProfileRepository) ListDrafts(ctx AgentCopilotProfileContext) ([]AgentCopilotProfileDraftSummary, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errAgentCopilotProfileStore
	}
	prefix := agentCopilotProfileRepositoryPrefix(ctx)
	summaries := make([]AgentCopilotProfileDraftSummary, 0)
	for key, draft := range repository.drafts {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		if key != agentCopilotProfileRepositoryKey(ctx, draft.ProfileID) {
			return nil, errAgentCopilotProfileOwner
		}
		if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
			return nil, err
		}
		summaries = append(summaries, agentCopilotProfileDraftSummary(draft))
	}
	sort.Slice(summaries, func(left, right int) bool {
		if summaries[left].UpdatedAt == summaries[right].UpdatedAt {
			return summaries[left].ProfileID < summaries[right].ProfileID
		}
		return summaries[left].UpdatedAt > summaries[right].UpdatedAt
	})
	return summaries, nil
}

func (repository *memoryAgentCopilotProfileRepository) CreateVersion(ctx AgentCopilotProfileContext, version AgentCopilotProfileVersionV1) (AgentCopilotProfileVersionV1, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	key := agentCopilotProfileRepositoryKey(ctx, version.ProfileID)
	draft, exists := repository.drafts[key]
	if !exists {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.ProfileDigest != version.ProfileDigest || draft.PolicyDigest != version.PolicyDigest {
		return AgentCopilotProfileVersionV1{}, agentCopilotProfileVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	versions := repository.versions[key]
	if versions == nil {
		versions = make(map[int]AgentCopilotProfileVersionV1)
		repository.versions[key] = versions
	}
	for storedVersion, existing := range versions {
		if storedVersion != existing.ProfileVersion || storedVersion < 1 || storedVersion > len(versions) {
			return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileOwner
		}
		if err := validateStoredAgentCopilotProfileVersion(ctx, existing); err != nil {
			return AgentCopilotProfileVersionV1{}, err
		}
		if existing.SourceDraftVersion == version.SourceDraftVersion {
			return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileImmutable
		}
	}
	version.ProfileVersion = len(versions) + 1
	if _, exists := versions[version.ProfileVersion]; exists {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileOwner
	}
	if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	versions[version.ProfileVersion] = cloneAgentCopilotProfileVersion(version)
	return cloneAgentCopilotProfileVersion(version), nil
}

func (repository *memoryAgentCopilotProfileRepository) ReadVersion(ctx AgentCopilotProfileContext, profileID string, profileVersion int) (AgentCopilotProfileVersionV1, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileStore
	}
	version, exists := repository.versions[agentCopilotProfileRepositoryKey(ctx, profileID)][profileVersion]
	if !exists {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileNotFound
	}
	if version.ProfileID != strings.TrimSpace(profileID) || version.ProfileVersion != profileVersion {
		return AgentCopilotProfileVersionV1{}, errAgentCopilotProfileOwner
	}
	if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
		return AgentCopilotProfileVersionV1{}, err
	}
	return cloneAgentCopilotProfileVersion(version), nil
}

func (repository *memoryAgentCopilotProfileRepository) ListVersions(ctx AgentCopilotProfileContext, profileID string) ([]AgentCopilotProfileVersionSummary, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errAgentCopilotProfileStore
	}
	key := agentCopilotProfileRepositoryKey(ctx, profileID)
	draft, exists := repository.drafts[key]
	if !exists {
		return nil, errAgentCopilotProfileNotFound
	}
	if err := validateStoredAgentCopilotProfileDraft(ctx, draft); err != nil {
		return nil, err
	}
	summaries := make([]AgentCopilotProfileVersionSummary, 0, len(repository.versions[key]))
	for storedVersion, version := range repository.versions[key] {
		if storedVersion != version.ProfileVersion || storedVersion < 1 || storedVersion > len(repository.versions[key]) {
			return nil, errAgentCopilotProfileOwner
		}
		if err := validateStoredAgentCopilotProfileVersion(ctx, version); err != nil {
			return nil, err
		}
		summaries = append(summaries, agentCopilotProfileVersionSummary(version))
	}
	sort.Slice(summaries, func(left, right int) bool { return summaries[left].ProfileVersion > summaries[right].ProfileVersion })
	return summaries, nil
}

func agentCopilotProfileDraftSummary(draft AgentCopilotProfileDraftV1) AgentCopilotProfileDraftSummary {
	compiled, _ := CompileAgentCopilotProfileSource(draft.AgentCopilotProfileSource)
	return AgentCopilotProfileDraftSummary{
		SchemaVersion: draft.SchemaVersion, ProfileID: draft.ProfileID, ApplicationID: draft.ApplicationID,
		ProfileName: draft.ProfileName, Description: draft.Description, Project: draft.Project,
		AllowedTasks: append([]string{}, draft.AllowedTasks...), DefaultLocale: draft.DefaultLocale,
		DraftVersion: draft.DraftVersion, ProfileDigest: draft.ProfileDigest, PolicyDigest: draft.PolicyDigest,
		AllowedTasksDigest: compiled.AllowedTasksDigest, ValidationState: draft.ValidationSummary.State,
		UpdatedAt: draft.UpdatedAt, UpdatedByActorRef: draft.UpdatedByActorRef,
	}
}

func agentCopilotProfileVersionSummary(version AgentCopilotProfileVersionV1) AgentCopilotProfileVersionSummary {
	compiled, _ := CompileAgentCopilotProfileSource(version.AgentCopilotProfileSource)
	return AgentCopilotProfileVersionSummary{
		SchemaVersion: version.SchemaVersion, ProfileID: version.ProfileID, ProfileVersion: version.ProfileVersion,
		SourceDraftVersion: version.SourceDraftVersion, ProfileName: version.ProfileName, Project: version.Project,
		DefaultLocale: version.DefaultLocale, ProfileDigest: version.ProfileDigest, PolicyDigest: version.PolicyDigest,
		AllowedTasksDigest: compiled.AllowedTasksDigest, PublishedAt: version.PublishedAt,
		PublishedByActorRef: version.PublishedByActorRef,
	}
}

func agentCopilotProfileRepositoryFailure(err error, validation ApplicationConfigurationDraftValidation) AgentCopilotProfileResult {
	result := AgentCopilotProfileResult{FailureCode: AgentCopilotProfileFailureStoreUnavailable, ValidationSummary: validation}
	switch {
	case errors.Is(err, errAgentCopilotProfileNotFound):
		result.FailureCode = AgentCopilotProfileFailureNotFound
	case errors.Is(err, errAgentCopilotProfileConflict):
		result.FailureCode = AgentCopilotProfileFailureVersionConflict
	case errors.Is(err, errAgentCopilotProfileImmutable):
		result.FailureCode = AgentCopilotProfileFailureImmutableConflict
	case errors.Is(err, errAgentCopilotProfileDigest):
		result.FailureCode = AgentCopilotProfileFailureDigestDrift
	}
	var conflict agentCopilotProfileVersionConflictError
	if errors.As(err, &conflict) {
		result.CurrentDraftVersion = conflict.CurrentDraftVersion
	}
	return result
}

func agentCopilotProfileValidationFailure(validation ApplicationConfigurationDraftValidation) string {
	if len(validation.Findings) == 0 {
		return AgentCopilotProfileFailurePayloadInvalid
	}
	return validation.Findings[0].Code
}

func agentCopilotProfileRepositoryKey(ctx AgentCopilotProfileContext, profileID string) string {
	return agentCopilotProfileRepositoryPrefix(ctx) + strings.TrimSpace(profileID)
}

func agentCopilotProfileRepositoryPrefix(ctx AgentCopilotProfileContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, ""}, "\x00")
}

func cloneAgentCopilotProfileSource(source AgentCopilotProfileSource) AgentCopilotProfileSource {
	payload, err := json.Marshal(source)
	if err != nil {
		return AgentCopilotProfileSource{}
	}
	var cloned AgentCopilotProfileSource
	if json.Unmarshal(payload, &cloned) != nil {
		return AgentCopilotProfileSource{}
	}
	return cloned
}

func cloneAgentCopilotProfileDraft(draft AgentCopilotProfileDraftV1) AgentCopilotProfileDraftV1 {
	draft.AgentCopilotProfileSource = cloneAgentCopilotProfileSource(draft.AgentCopilotProfileSource)
	draft.ValidationSummary.Findings = append([]ApplicationConfigurationDraftValidationFinding{}, draft.ValidationSummary.Findings...)
	return draft
}

func cloneAgentCopilotProfileVersion(version AgentCopilotProfileVersionV1) AgentCopilotProfileVersionV1 {
	version.AgentCopilotProfileSource = cloneAgentCopilotProfileSource(version.AgentCopilotProfileSource)
	return version
}
