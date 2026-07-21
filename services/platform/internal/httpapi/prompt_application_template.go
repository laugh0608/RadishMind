package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	promptApplicationTemplateDraftSchemaVersion   = "prompt_application_template_draft.v1"
	promptApplicationTemplateVersionSchemaVersion = "prompt_application_template_version.v1"
	promptApplicationTemplateValidationValid      = "valid"
	promptApplicationTemplateValidationInvalid    = "invalid"
)

const (
	PromptApplicationTemplateFailureScopeDenied         = "prompt_template_scope_denied"
	PromptApplicationTemplateFailureNotFound            = "prompt_template_not_found"
	PromptApplicationTemplateFailureVersionNotFound     = "prompt_template_version_not_found"
	PromptApplicationTemplateFailureVersionConflict     = "prompt_template_version_conflict"
	PromptApplicationTemplateFailureStoreUnavailable    = "prompt_template_store_unavailable"
	PromptApplicationTemplateFailureWriteDisabled       = "prompt_template_write_disabled"
	PromptApplicationTemplateFailureImmutableConflict   = "prompt_template_immutable_conflict"
	PromptApplicationTemplateFailureDigestDrift         = "prompt_template_digest_drift"
	PromptApplicationTemplateFailureStoreContract       = "prompt_template_store_contract_mismatch"
	PromptApplicationTemplateFailureApplicationMissing  = "prompt_template_application_not_found"
	PromptApplicationTemplateFailureApplicationArchived = "prompt_template_application_archived"
	PromptApplicationTemplateFailureApplicationKind     = "prompt_template_application_kind_mismatch"
)

var (
	promptApplicationTemplateIDPattern          = regexp.MustCompile(`^ptpl_[a-z2-7]{16}$`)
	errPromptApplicationTemplateNotFound        = errors.New(PromptApplicationTemplateFailureNotFound)
	errPromptApplicationTemplateVersionNotFound = errors.New(PromptApplicationTemplateFailureVersionNotFound)
	errPromptApplicationTemplateVersionConflict = errors.New(PromptApplicationTemplateFailureVersionConflict)
	errPromptApplicationTemplateImmutable       = errors.New(PromptApplicationTemplateFailureImmutableConflict)
	errPromptApplicationTemplateStore           = errors.New(PromptApplicationTemplateFailureStoreUnavailable)
	errPromptApplicationTemplateContract        = errors.New(PromptApplicationTemplateFailureStoreContract)
)

type promptApplicationTemplateVersionConflictError struct {
	CurrentDraftVersion int
}

func (failure promptApplicationTemplateVersionConflictError) Error() string {
	return PromptApplicationTemplateFailureVersionConflict
}

func (failure promptApplicationTemplateVersionConflictError) Is(target error) bool {
	return target == errPromptApplicationTemplateVersionConflict
}

type PromptApplicationTemplateContext struct {
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

type PromptApplicationTemplateDraftInput struct {
	SchemaVersion string `json:"schema_version"`
	TemplateID    string `json:"template_id"`
	WorkspaceID   string `json:"workspace_id"`
	ApplicationID string `json:"application_id"`
	TemplateName  string `json:"template_name"`
	Description   string `json:"description"`
	PromptApplicationTemplateSource
}

type PromptApplicationTemplateValidation struct {
	State    string                             `json:"state"`
	IsValid  bool                               `json:"is_valid"`
	Findings []PromptApplicationTemplateFinding `json:"findings"`
}

type PromptApplicationTemplateDraft struct {
	PromptApplicationTemplateDraftInput
	TenantRef         string                              `json:"tenant_ref"`
	OwnerSubjectRef   string                              `json:"owner_subject_ref"`
	DraftVersion      int                                 `json:"draft_version"`
	TemplateDigest    string                              `json:"template_digest"`
	ValidationSummary PromptApplicationTemplateValidation `json:"validation_summary"`
	CreatedAt         string                              `json:"created_at"`
	UpdatedAt         string                              `json:"updated_at"`
	CreatedByActorRef string                              `json:"created_by_actor_ref"`
	UpdatedByActorRef string                              `json:"updated_by_actor_ref"`
	RequestID         string                              `json:"request_id"`
	AuditRef          string                              `json:"audit_ref"`
}

type PromptApplicationTemplateDraftSummary struct {
	SchemaVersion     string   `json:"schema_version"`
	TemplateID        string   `json:"template_id"`
	ApplicationID     string   `json:"application_id"`
	TemplateName      string   `json:"template_name"`
	Description       string   `json:"description"`
	DraftVersion      int      `json:"draft_version"`
	TemplateDigest    string   `json:"template_digest"`
	ValidationState   string   `json:"validation_state"`
	MessageRoles      []string `json:"message_roles"`
	VariableNames     []string `json:"variable_names"`
	OutputKind        string   `json:"output_kind"`
	UpdatedAt         string   `json:"updated_at"`
	UpdatedByActorRef string   `json:"updated_by_actor_ref"`
}

type PromptApplicationTemplateVersion struct {
	SchemaVersion      string `json:"schema_version"`
	TemplateID         string `json:"template_id"`
	TemplateVersion    int    `json:"template_version"`
	SourceDraftVersion int    `json:"source_draft_version"`
	TenantRef          string `json:"tenant_ref"`
	WorkspaceID        string `json:"workspace_id"`
	ApplicationID      string `json:"application_id"`
	OwnerSubjectRef    string `json:"owner_subject_ref"`
	TemplateName       string `json:"template_name"`
	Description        string `json:"description"`
	PromptApplicationTemplateSource
	TemplateDigest    string `json:"template_digest"`
	CreatedAt         string `json:"created_at"`
	CreatedByActorRef string `json:"created_by_actor_ref"`
	RequestID         string `json:"request_id"`
	AuditRef          string `json:"audit_ref"`
}

type PromptApplicationTemplateVersionSummary struct {
	SchemaVersion      string `json:"schema_version"`
	TemplateID         string `json:"template_id"`
	TemplateVersion    int    `json:"template_version"`
	SourceDraftVersion int    `json:"source_draft_version"`
	TemplateName       string `json:"template_name"`
	TemplateDigest     string `json:"template_digest"`
	OutputKind         string `json:"output_kind"`
	CreatedAt          string `json:"created_at"`
	CreatedByActorRef  string `json:"created_by_actor_ref"`
}

type PromptApplicationTemplateResult struct {
	Draft                  *PromptApplicationTemplateDraft
	Version                *PromptApplicationTemplateVersion
	ValidationSummary      PromptApplicationTemplateValidation
	FailureCode            string
	CurrentDraftVersion    int
	CurrentTemplateVersion int
}

type promptApplicationTemplateRepository interface {
	SaveDraft(PromptApplicationTemplateContext, PromptApplicationTemplateDraft, int) (PromptApplicationTemplateDraft, error)
	ReadDraft(PromptApplicationTemplateContext, string) (PromptApplicationTemplateDraft, error)
	ListDrafts(PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, error)
	CreateVersion(PromptApplicationTemplateContext, PromptApplicationTemplateVersion) (PromptApplicationTemplateVersion, error)
	ReadVersion(PromptApplicationTemplateContext, string, int) (PromptApplicationTemplateVersion, error)
	ListVersions(PromptApplicationTemplateContext, string) ([]PromptApplicationTemplateVersionSummary, error)
}

type memoryPromptApplicationTemplateRepository struct {
	mu          sync.RWMutex
	drafts      map[string]PromptApplicationTemplateDraft
	versions    map[string]map[int]PromptApplicationTemplateVersion
	unavailable bool
}

type promptApplicationTemplateService struct {
	repository               promptApplicationTemplateRepository
	requirePromptApplication func(PromptApplicationTemplateContext) string
	now                      func() time.Time
}

func newMemoryPromptApplicationTemplateRepository() *memoryPromptApplicationTemplateRepository {
	return &memoryPromptApplicationTemplateRepository{
		drafts: make(map[string]PromptApplicationTemplateDraft), versions: make(map[string]map[int]PromptApplicationTemplateVersion),
	}
}

func newPromptApplicationTemplateService(repository promptApplicationTemplateRepository) promptApplicationTemplateService {
	return promptApplicationTemplateService{repository: repository, now: func() time.Time { return time.Now().UTC() }}
}

func (service promptApplicationTemplateService) Validate(ctx PromptApplicationTemplateContext, input PromptApplicationTemplateDraftInput) PromptApplicationTemplateResult {
	validation := validatePromptApplicationTemplateDraftInput(ctx, input)
	return PromptApplicationTemplateResult{ValidationSummary: validation}
}

func (service promptApplicationTemplateService) SaveDraft(ctx PromptApplicationTemplateContext, input PromptApplicationTemplateDraftInput, expectedVersion int) PromptApplicationTemplateResult {
	validation := validatePromptApplicationTemplateDraftInput(ctx, input)
	if !validation.IsValid {
		return PromptApplicationTemplateResult{FailureCode: promptApplicationTemplateValidationFailure(validation), ValidationSummary: validation}
	}
	if !ctx.WriteEnabled {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureWriteDisabled, ValidationSummary: validation}
	}
	if expectedVersion < 0 {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailurePayloadInvalid, ValidationSummary: validation}
	}
	if service.requirePromptApplication != nil {
		if failure := service.requirePromptApplication(ctx); failure != "" {
			return PromptApplicationTemplateResult{FailureCode: failure, ValidationSummary: validation}
		}
	}
	normalizedSource, err := NormalizePromptApplicationTemplateSource(input.PromptApplicationTemplateSource)
	if err != nil {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailurePayloadInvalid, ValidationSummary: validation}
	}
	input = normalizePromptApplicationTemplateDraftInput(input, normalizedSource)
	digest, err := promptApplicationTemplateSourceDigest(normalizedSource)
	if err != nil {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureStoreUnavailable, ValidationSummary: validation}
	}
	now := service.now().Format(time.RFC3339Nano)
	draft := PromptApplicationTemplateDraft{
		PromptApplicationTemplateDraftInput: input, TenantRef: ctx.TenantRef, OwnerSubjectRef: ctx.OwnerSubjectRef,
		DraftVersion: expectedVersion + 1, TemplateDigest: digest, ValidationSummary: validation,
		CreatedAt: now, UpdatedAt: now, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	saved, err := service.repository.SaveDraft(ctx, draft, expectedVersion)
	if err != nil {
		return promptApplicationTemplateRepositoryFailure(err, validation)
	}
	return PromptApplicationTemplateResult{Draft: &saved, ValidationSummary: saved.ValidationSummary, CurrentDraftVersion: saved.DraftVersion}
}

func (service promptApplicationTemplateService) ReadDraft(ctx PromptApplicationTemplateContext, templateID string) PromptApplicationTemplateResult {
	if validatePromptApplicationTemplateContext(ctx) != nil || !promptApplicationTemplateIDPattern.MatchString(strings.TrimSpace(templateID)) {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureScopeDenied}
	}
	draft, err := service.repository.ReadDraft(ctx, strings.TrimSpace(templateID))
	if err != nil {
		return promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{})
	}
	return PromptApplicationTemplateResult{Draft: &draft, ValidationSummary: draft.ValidationSummary, CurrentDraftVersion: draft.DraftVersion}
}

func (service promptApplicationTemplateService) ListDrafts(ctx PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, string) {
	if validatePromptApplicationTemplateContext(ctx) != nil {
		return []PromptApplicationTemplateDraftSummary{}, PromptApplicationTemplateFailureScopeDenied
	}
	summaries, err := service.repository.ListDrafts(ctx)
	if err != nil {
		return []PromptApplicationTemplateDraftSummary{}, PromptApplicationTemplateFailureStoreUnavailable
	}
	return summaries, ""
}

func (service promptApplicationTemplateService) CreateVersion(ctx PromptApplicationTemplateContext, templateID string, sourceDraftVersion int) PromptApplicationTemplateResult {
	if !ctx.WriteEnabled {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureWriteDisabled}
	}
	templateID = strings.TrimSpace(templateID)
	if validatePromptApplicationTemplateContext(ctx) != nil || !promptApplicationTemplateIDPattern.MatchString(templateID) || sourceDraftVersion < 1 {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailurePayloadInvalid}
	}
	if service.requirePromptApplication != nil {
		if failure := service.requirePromptApplication(ctx); failure != "" {
			return PromptApplicationTemplateResult{FailureCode: failure}
		}
	}
	draft, err := service.repository.ReadDraft(ctx, templateID)
	if err != nil {
		return promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{})
	}
	if draft.DraftVersion != sourceDraftVersion {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureVersionConflict, CurrentDraftVersion: draft.DraftVersion}
	}
	if validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil || !draft.ValidationSummary.IsValid {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureDigestDrift, CurrentDraftVersion: draft.DraftVersion}
	}
	createdAt := service.now().Format(time.RFC3339Nano)
	version := PromptApplicationTemplateVersion{
		SchemaVersion: promptApplicationTemplateVersionSchemaVersion, TemplateID: draft.TemplateID,
		SourceDraftVersion: draft.DraftVersion, TenantRef: draft.TenantRef, WorkspaceID: draft.WorkspaceID,
		ApplicationID: draft.ApplicationID, OwnerSubjectRef: draft.OwnerSubjectRef, TemplateName: draft.TemplateName,
		Description: draft.Description, PromptApplicationTemplateSource: clonePromptApplicationTemplateSource(draft.PromptApplicationTemplateSource),
		TemplateDigest: draft.TemplateDigest, CreatedAt: createdAt, CreatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	created, err := service.repository.CreateVersion(ctx, version)
	if err != nil {
		return promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{})
	}
	return PromptApplicationTemplateResult{Version: &created, CurrentDraftVersion: draft.DraftVersion, CurrentTemplateVersion: created.TemplateVersion}
}

func (service promptApplicationTemplateService) ReadVersion(ctx PromptApplicationTemplateContext, templateID string, templateVersion int) PromptApplicationTemplateResult {
	templateID = strings.TrimSpace(templateID)
	if validatePromptApplicationTemplateContext(ctx) != nil || !promptApplicationTemplateIDPattern.MatchString(templateID) || templateVersion < 1 {
		return PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureScopeDenied}
	}
	version, err := service.repository.ReadVersion(ctx, templateID, templateVersion)
	if err != nil {
		return promptApplicationTemplateRepositoryFailure(err, PromptApplicationTemplateValidation{})
	}
	return PromptApplicationTemplateResult{Version: &version, CurrentTemplateVersion: version.TemplateVersion}
}

func (service promptApplicationTemplateService) ListVersions(ctx PromptApplicationTemplateContext, templateID string) ([]PromptApplicationTemplateVersionSummary, string) {
	templateID = strings.TrimSpace(templateID)
	if validatePromptApplicationTemplateContext(ctx) != nil || !promptApplicationTemplateIDPattern.MatchString(templateID) {
		return []PromptApplicationTemplateVersionSummary{}, PromptApplicationTemplateFailureScopeDenied
	}
	summaries, err := service.repository.ListVersions(ctx, templateID)
	if err != nil {
		if errors.Is(err, errPromptApplicationTemplateNotFound) {
			return []PromptApplicationTemplateVersionSummary{}, PromptApplicationTemplateFailureNotFound
		}
		return []PromptApplicationTemplateVersionSummary{}, PromptApplicationTemplateFailureStoreUnavailable
	}
	return summaries, ""
}

func validatePromptApplicationTemplateDraftInput(ctx PromptApplicationTemplateContext, input PromptApplicationTemplateDraftInput) PromptApplicationTemplateValidation {
	findings := make([]PromptApplicationTemplateFinding, 0)
	if validatePromptApplicationTemplateContext(ctx) != nil || strings.TrimSpace(input.WorkspaceID) != ctx.WorkspaceID || strings.TrimSpace(input.ApplicationID) != ctx.ApplicationID {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureScopeDenied, "scope", "template scope does not match the authenticated context")
	}
	if strings.TrimSpace(input.SchemaVersion) != promptApplicationTemplateDraftSchemaVersion || !promptApplicationTemplateIDPattern.MatchString(strings.TrimSpace(input.TemplateID)) {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, "schema_version", "template schema version or identifier is invalid")
	}
	name, description := strings.TrimSpace(input.TemplateName), strings.TrimSpace(input.Description)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 80 || !utf8.ValidString(description) || utf8.RuneCountInString(description) > 512 {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, "template_name", "template name or description is invalid")
	}
	if promptApplicationContainsSecretMaterial(name) || promptApplicationContainsSecretMaterial(description) {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureSecretForbidden, "template_name", "template metadata contains credential-like material")
	}
	findings = append(findings, ValidatePromptApplicationTemplateSource(input.PromptApplicationTemplateSource)...)
	state := promptApplicationTemplateValidationValid
	if len(findings) != 0 {
		state = promptApplicationTemplateValidationInvalid
	}
	return PromptApplicationTemplateValidation{State: state, IsValid: len(findings) == 0, Findings: findings}
}

func validatePromptApplicationTemplateContext(ctx PromptApplicationTemplateContext) error {
	if ctx.RequestContext == nil || !controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.TenantRef)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) || !applicationCatalogIDPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.ActorRef)) || !controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.OwnerSubjectRef)) ||
		!controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.RequestID)) || !controlPlaneReadAuthReferencePattern.MatchString(strings.TrimSpace(ctx.AuditRef)) {
		return errPromptApplicationTemplateContract
	}
	return nil
}

func validateStoredPromptApplicationTemplateDraft(ctx PromptApplicationTemplateContext, draft PromptApplicationTemplateDraft) error {
	createdAt := parsePromptApplicationTemplateTimestamp(draft.CreatedAt)
	updatedAt := parsePromptApplicationTemplateTimestamp(draft.UpdatedAt)
	if draft.SchemaVersion != promptApplicationTemplateDraftSchemaVersion || draft.TenantRef != ctx.TenantRef || draft.WorkspaceID != ctx.WorkspaceID ||
		draft.ApplicationID != ctx.ApplicationID || draft.OwnerSubjectRef != ctx.OwnerSubjectRef || draft.DraftVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(draft.TemplateDigest) || draft.ValidationSummary.State != promptApplicationTemplateValidationValid || !draft.ValidationSummary.IsValid ||
		len(draft.ValidationSummary.Findings) != 0 || createdAt == nil || updatedAt == nil || updatedAt.Before(*createdAt) ||
		!controlPlaneReadAuthReferencePattern.MatchString(draft.CreatedByActorRef) || !controlPlaneReadAuthReferencePattern.MatchString(draft.UpdatedByActorRef) ||
		!controlPlaneReadAuthReferencePattern.MatchString(draft.RequestID) || !controlPlaneReadAuthReferencePattern.MatchString(draft.AuditRef) {
		return errPromptApplicationTemplateContract
	}
	validation := validatePromptApplicationTemplateDraftInput(ctx, draft.PromptApplicationTemplateDraftInput)
	if !validation.IsValid {
		return errPromptApplicationTemplateContract
	}
	normalized, err := NormalizePromptApplicationTemplateSource(draft.PromptApplicationTemplateSource)
	if err != nil {
		return errPromptApplicationTemplateContract
	}
	digest, err := promptApplicationTemplateSourceDigest(normalized)
	if err != nil || digest != draft.TemplateDigest {
		return errPromptApplicationTemplateContract
	}
	return nil
}

func validateStoredPromptApplicationTemplateVersion(ctx PromptApplicationTemplateContext, version PromptApplicationTemplateVersion) error {
	if version.SchemaVersion != promptApplicationTemplateVersionSchemaVersion || !promptApplicationTemplateIDPattern.MatchString(version.TemplateID) ||
		version.TemplateVersion < 1 || version.SourceDraftVersion < 1 || version.TenantRef != ctx.TenantRef || version.WorkspaceID != ctx.WorkspaceID ||
		version.ApplicationID != ctx.ApplicationID || version.OwnerSubjectRef != ctx.OwnerSubjectRef || !workflowRAGDigestPattern.MatchString(version.TemplateDigest) ||
		parsePromptApplicationTemplateTimestamp(version.CreatedAt) == nil || !controlPlaneReadAuthReferencePattern.MatchString(version.CreatedByActorRef) ||
		!controlPlaneReadAuthReferencePattern.MatchString(version.RequestID) || !controlPlaneReadAuthReferencePattern.MatchString(version.AuditRef) {
		return errPromptApplicationTemplateContract
	}
	input := PromptApplicationTemplateDraftInput{SchemaVersion: promptApplicationTemplateDraftSchemaVersion, TemplateID: version.TemplateID, WorkspaceID: version.WorkspaceID, ApplicationID: version.ApplicationID, TemplateName: version.TemplateName, Description: version.Description, PromptApplicationTemplateSource: version.PromptApplicationTemplateSource}
	if !validatePromptApplicationTemplateDraftInput(ctx, input).IsValid {
		return errPromptApplicationTemplateContract
	}
	normalized, err := NormalizePromptApplicationTemplateSource(version.PromptApplicationTemplateSource)
	if err != nil {
		return errPromptApplicationTemplateContract
	}
	digest, err := promptApplicationTemplateSourceDigest(normalized)
	if err != nil || digest != version.TemplateDigest {
		return errPromptApplicationTemplateContract
	}
	return nil
}

func promptApplicationTemplateSourceDigest(source PromptApplicationTemplateSource) (string, error) {
	normalized, err := NormalizePromptApplicationTemplateSource(source)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func normalizePromptApplicationTemplateDraftInput(input PromptApplicationTemplateDraftInput, source PromptApplicationTemplateSource) PromptApplicationTemplateDraftInput {
	input.SchemaVersion = strings.TrimSpace(input.SchemaVersion)
	input.TemplateID = strings.TrimSpace(input.TemplateID)
	input.WorkspaceID = strings.TrimSpace(input.WorkspaceID)
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.TemplateName = strings.TrimSpace(input.TemplateName)
	input.Description = strings.TrimSpace(input.Description)
	input.PromptApplicationTemplateSource = source
	return input
}

func (repository *memoryPromptApplicationTemplateRepository) SaveDraft(ctx PromptApplicationTemplateContext, draft PromptApplicationTemplateDraft, expectedVersion int) (PromptApplicationTemplateDraft, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	if validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateContract
	}
	key := promptApplicationTemplateRepositoryKey(ctx, draft.TemplateID)
	current, exists := repository.drafts[key]
	if !exists && expectedVersion != 0 || exists && current.DraftVersion != expectedVersion {
		currentVersion := 0
		if exists {
			currentVersion = current.DraftVersion
		}
		return PromptApplicationTemplateDraft{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: currentVersion}
	}
	if exists {
		draft.CreatedAt = current.CreatedAt
		draft.CreatedByActorRef = current.CreatedByActorRef
	}
	repository.drafts[key] = clonePromptApplicationTemplateDraft(draft)
	return clonePromptApplicationTemplateDraft(draft), nil
}

func (repository *memoryPromptApplicationTemplateRepository) ReadDraft(ctx PromptApplicationTemplateContext, templateID string) (PromptApplicationTemplateDraft, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateStore
	}
	draft, exists := repository.drafts[promptApplicationTemplateRepositoryKey(ctx, templateID)]
	if !exists {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateNotFound
	}
	if validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil {
		return PromptApplicationTemplateDraft{}, errPromptApplicationTemplateContract
	}
	return clonePromptApplicationTemplateDraft(draft), nil
}

func (repository *memoryPromptApplicationTemplateRepository) ListDrafts(ctx PromptApplicationTemplateContext) ([]PromptApplicationTemplateDraftSummary, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errPromptApplicationTemplateStore
	}
	prefix := promptApplicationTemplateRepositoryPrefix(ctx)
	summaries := make([]PromptApplicationTemplateDraftSummary, 0)
	for key, draft := range repository.drafts {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		if validateStoredPromptApplicationTemplateDraft(ctx, draft) != nil {
			return nil, errPromptApplicationTemplateContract
		}
		summaries = append(summaries, promptApplicationTemplateDraftSummary(draft))
	}
	sort.Slice(summaries, func(left, right int) bool {
		if summaries[left].UpdatedAt == summaries[right].UpdatedAt {
			return summaries[left].TemplateID < summaries[right].TemplateID
		}
		return summaries[left].UpdatedAt > summaries[right].UpdatedAt
	})
	return summaries, nil
}

func (repository *memoryPromptApplicationTemplateRepository) CreateVersion(ctx PromptApplicationTemplateContext, version PromptApplicationTemplateVersion) (PromptApplicationTemplateVersion, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	key := promptApplicationTemplateRepositoryKey(ctx, version.TemplateID)
	draft, exists := repository.drafts[key]
	if !exists {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateNotFound
	}
	if draft.DraftVersion != version.SourceDraftVersion || draft.TemplateDigest != version.TemplateDigest {
		return PromptApplicationTemplateVersion{}, promptApplicationTemplateVersionConflictError{CurrentDraftVersion: draft.DraftVersion}
	}
	versions := repository.versions[key]
	if versions == nil {
		versions = make(map[int]PromptApplicationTemplateVersion)
		repository.versions[key] = versions
	}
	for _, existing := range versions {
		if existing.SourceDraftVersion == version.SourceDraftVersion {
			return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateImmutable
		}
	}
	version.TemplateVersion = len(versions) + 1
	if validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	versions[version.TemplateVersion] = clonePromptApplicationTemplateVersion(version)
	return clonePromptApplicationTemplateVersion(version), nil
}

func (repository *memoryPromptApplicationTemplateRepository) ReadVersion(ctx PromptApplicationTemplateContext, templateID string, templateVersion int) (PromptApplicationTemplateVersion, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateStore
	}
	version, exists := repository.versions[promptApplicationTemplateRepositoryKey(ctx, templateID)][templateVersion]
	if !exists {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateVersionNotFound
	}
	if validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
		return PromptApplicationTemplateVersion{}, errPromptApplicationTemplateContract
	}
	return clonePromptApplicationTemplateVersion(version), nil
}

func (repository *memoryPromptApplicationTemplateRepository) ListVersions(ctx PromptApplicationTemplateContext, templateID string) ([]PromptApplicationTemplateVersionSummary, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errPromptApplicationTemplateStore
	}
	key := promptApplicationTemplateRepositoryKey(ctx, templateID)
	if _, exists := repository.drafts[key]; !exists {
		return nil, errPromptApplicationTemplateNotFound
	}
	versions := repository.versions[key]
	summaries := make([]PromptApplicationTemplateVersionSummary, 0, len(versions))
	for _, version := range versions {
		if validateStoredPromptApplicationTemplateVersion(ctx, version) != nil {
			return nil, errPromptApplicationTemplateContract
		}
		summaries = append(summaries, promptApplicationTemplateVersionSummary(version))
	}
	sort.Slice(summaries, func(left, right int) bool { return summaries[left].TemplateVersion > summaries[right].TemplateVersion })
	return summaries, nil
}

func promptApplicationTemplateDraftSummary(draft PromptApplicationTemplateDraft) PromptApplicationTemplateDraftSummary {
	roles := make([]string, 0, len(draft.Messages))
	for _, message := range draft.Messages {
		roles = append(roles, message.Role)
	}
	names := make([]string, 0, len(draft.Variables))
	for _, variable := range draft.Variables {
		names = append(names, variable.Name)
	}
	return PromptApplicationTemplateDraftSummary{
		SchemaVersion: draft.SchemaVersion, TemplateID: draft.TemplateID, ApplicationID: draft.ApplicationID,
		TemplateName: draft.TemplateName, Description: draft.Description, DraftVersion: draft.DraftVersion,
		TemplateDigest: draft.TemplateDigest, ValidationState: draft.ValidationSummary.State, MessageRoles: roles,
		VariableNames: names, OutputKind: draft.OutputContract.Kind, UpdatedAt: draft.UpdatedAt, UpdatedByActorRef: draft.UpdatedByActorRef,
	}
}

func promptApplicationTemplateVersionSummary(version PromptApplicationTemplateVersion) PromptApplicationTemplateVersionSummary {
	return PromptApplicationTemplateVersionSummary{
		SchemaVersion: version.SchemaVersion, TemplateID: version.TemplateID, TemplateVersion: version.TemplateVersion,
		SourceDraftVersion: version.SourceDraftVersion, TemplateName: version.TemplateName, TemplateDigest: version.TemplateDigest,
		OutputKind: version.OutputContract.Kind, CreatedAt: version.CreatedAt, CreatedByActorRef: version.CreatedByActorRef,
	}
}

func promptApplicationTemplateRepositoryFailure(err error, validation PromptApplicationTemplateValidation) PromptApplicationTemplateResult {
	result := PromptApplicationTemplateResult{FailureCode: PromptApplicationTemplateFailureStoreUnavailable, ValidationSummary: validation}
	switch {
	case errors.Is(err, errPromptApplicationTemplateNotFound):
		result.FailureCode = PromptApplicationTemplateFailureNotFound
	case errors.Is(err, errPromptApplicationTemplateVersionNotFound):
		result.FailureCode = PromptApplicationTemplateFailureVersionNotFound
	case errors.Is(err, errPromptApplicationTemplateVersionConflict):
		result.FailureCode = PromptApplicationTemplateFailureVersionConflict
	case errors.Is(err, errPromptApplicationTemplateImmutable):
		result.FailureCode = PromptApplicationTemplateFailureImmutableConflict
	case errors.Is(err, errPromptApplicationTemplateContract):
		result.FailureCode = PromptApplicationTemplateFailureStoreContract
	}
	var conflict promptApplicationTemplateVersionConflictError
	if errors.As(err, &conflict) {
		result.CurrentDraftVersion = conflict.CurrentDraftVersion
	}
	return result
}

func promptApplicationTemplateValidationFailure(validation PromptApplicationTemplateValidation) string {
	if len(validation.Findings) == 0 {
		return PromptApplicationTemplateFailurePayloadInvalid
	}
	return validation.Findings[0].Code
}

func promptApplicationTemplateRepositoryKey(ctx PromptApplicationTemplateContext, templateID string) string {
	return promptApplicationTemplateRepositoryPrefix(ctx) + strings.TrimSpace(templateID)
}

func promptApplicationTemplateRepositoryPrefix(ctx PromptApplicationTemplateContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, ""}, "\x00")
}

func parsePromptApplicationTemplateTimestamp(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &parsed
}

func clonePromptApplicationTemplateSource(source PromptApplicationTemplateSource) PromptApplicationTemplateSource {
	payload, err := json.Marshal(source)
	if err != nil {
		return PromptApplicationTemplateSource{}
	}
	var cloned PromptApplicationTemplateSource
	if json.Unmarshal(payload, &cloned) != nil {
		return PromptApplicationTemplateSource{}
	}
	return cloned
}

func clonePromptApplicationTemplateDraft(draft PromptApplicationTemplateDraft) PromptApplicationTemplateDraft {
	draft.PromptApplicationTemplateSource = clonePromptApplicationTemplateSource(draft.PromptApplicationTemplateSource)
	draft.ValidationSummary.Findings = append([]PromptApplicationTemplateFinding{}, draft.ValidationSummary.Findings...)
	return draft
}

func clonePromptApplicationTemplateVersion(version PromptApplicationTemplateVersion) PromptApplicationTemplateVersion {
	version.PromptApplicationTemplateSource = clonePromptApplicationTemplateSource(version.PromptApplicationTemplateSource)
	return version
}

func validatePromptApplicationTemplateContractJSON(contract string, payload []byte) error {
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	switch contract {
	case promptApplicationTemplateDraftSchemaVersion:
		var draft PromptApplicationTemplateDraft
		if err := decoder.Decode(&draft); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return errPromptApplicationTemplateContract
		}
		ctx := PromptApplicationTemplateContext{RequestContext: context.Background(), RequestID: draft.RequestID, TenantRef: draft.TenantRef, WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID, ActorRef: draft.UpdatedByActorRef, OwnerSubjectRef: draft.OwnerSubjectRef, AuditRef: draft.AuditRef}
		return validateStoredPromptApplicationTemplateDraft(ctx, draft)
	case promptApplicationTemplateVersionSchemaVersion:
		var version PromptApplicationTemplateVersion
		if err := decoder.Decode(&version); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return errPromptApplicationTemplateContract
		}
		ctx := PromptApplicationTemplateContext{RequestContext: context.Background(), RequestID: version.RequestID, TenantRef: version.TenantRef, WorkspaceID: version.WorkspaceID, ApplicationID: version.ApplicationID, ActorRef: version.CreatedByActorRef, OwnerSubjectRef: version.OwnerSubjectRef, AuditRef: version.AuditRef}
		return validateStoredPromptApplicationTemplateVersion(ctx, version)
	default:
		return errPromptApplicationTemplateContract
	}
}
