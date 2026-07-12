package httpapi

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const applicationConfigurationDraftSchemaVersion = "application_configuration_draft.v1"

const (
	ApplicationDraftFailureScopeDenied      = "application_draft_scope_denied"
	ApplicationDraftFailureNotFound         = "application_draft_not_found"
	ApplicationDraftFailurePayloadInvalid   = "application_draft_payload_invalid"
	ApplicationDraftFailureSecretForbidden  = "application_draft_secret_material_forbidden"
	ApplicationDraftFailureVersionConflict  = "application_draft_version_conflict"
	ApplicationDraftFailureStoreUnavailable = "application_draft_store_unavailable"
	ApplicationDraftFailureWriteDisabled    = "application_draft_write_disabled"
)

const (
	applicationDraftValidationValid   = "valid"
	applicationDraftValidationInvalid = "invalid"
)

var (
	applicationDraftIdentifierPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$`)
	errApplicationDraftNotFound         = errors.New(ApplicationDraftFailureNotFound)
	errApplicationDraftVersionConflict  = errors.New(ApplicationDraftFailureVersionConflict)
	errApplicationDraftStoreUnavailable = errors.New(ApplicationDraftFailureStoreUnavailable)
)

type applicationDraftVersionConflictError struct {
	CurrentVersion int
}

func (err applicationDraftVersionConflictError) Error() string {
	return ApplicationDraftFailureVersionConflict
}

func (err applicationDraftVersionConflictError) Is(target error) bool {
	return target == errApplicationDraftVersionConflict
}

type ApplicationConfigurationDraftContext struct {
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

type ApplicationConfigurationDraftPayload struct {
	DraftID                  string   `json:"draft_id"`
	WorkspaceID              string   `json:"workspace_id"`
	ApplicationID            string   `json:"application_id"`
	BaseApplicationUpdatedAt string   `json:"base_application_updated_at"`
	SchemaVersion            string   `json:"schema_version"`
	DisplayName              string   `json:"display_name"`
	Description              string   `json:"description"`
	ApplicationKind          string   `json:"application_kind"`
	DefaultProtocol          string   `json:"default_protocol"`
	DefaultModel             string   `json:"default_model"`
	AllowedProtocols         []string `json:"allowed_protocols"`
}

type ApplicationConfigurationDraft struct {
	ApplicationConfigurationDraftPayload
	DraftVersion      int                                     `json:"draft_version"`
	ValidationSummary ApplicationConfigurationDraftValidation `json:"validation_summary"`
	CreatedAt         string                                  `json:"created_at"`
	UpdatedAt         string                                  `json:"updated_at"`
	CreatedByActorRef string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef string                                  `json:"updated_by_actor_ref"`
	RequestID         string                                  `json:"request_id"`
	AuditRef          string                                  `json:"audit_ref"`
}

type ApplicationConfigurationDraftValidation struct {
	State    string                                           `json:"state"`
	IsValid  bool                                             `json:"is_valid"`
	Findings []ApplicationConfigurationDraftValidationFinding `json:"findings"`
}

type ApplicationConfigurationDraftValidationFinding struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Summary string `json:"summary"`
}

type ApplicationConfigurationDraftSummary struct {
	DraftID           string `json:"draft_id"`
	ApplicationID     string `json:"application_id"`
	DraftVersion      int    `json:"draft_version"`
	DisplayName       string `json:"display_name"`
	ApplicationKind   string `json:"application_kind"`
	DefaultProtocol   string `json:"default_protocol"`
	DefaultModel      string `json:"default_model"`
	ValidationState   string `json:"validation_state"`
	UpdatedAt         string `json:"updated_at"`
	UpdatedByActorRef string `json:"updated_by_actor_ref"`
}

type ApplicationConfigurationDraftResult struct {
	Draft               *ApplicationConfigurationDraft          `json:"draft"`
	FailureCode         string                                  `json:"failure_code,omitempty"`
	CurrentDraftVersion int                                     `json:"current_draft_version"`
	ValidationSummary   ApplicationConfigurationDraftValidation `json:"validation_summary"`
}

type applicationConfigurationDraftRepository interface {
	Save(ApplicationConfigurationDraftContext, ApplicationConfigurationDraft, int) (ApplicationConfigurationDraft, error)
	Read(ApplicationConfigurationDraftContext, string) (ApplicationConfigurationDraft, error)
	List(ApplicationConfigurationDraftContext) ([]ApplicationConfigurationDraftSummary, error)
}

type memoryApplicationConfigurationDraftRepository struct {
	mu          sync.RWMutex
	drafts      map[string]ApplicationConfigurationDraft
	unavailable bool
}

type applicationConfigurationDraftService struct {
	repository applicationConfigurationDraftRepository
	now        func() time.Time
}

func newApplicationConfigurationDraftService(repository applicationConfigurationDraftRepository) applicationConfigurationDraftService {
	return applicationConfigurationDraftService{
		repository: repository,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func newMemoryApplicationConfigurationDraftRepository() *memoryApplicationConfigurationDraftRepository {
	return &memoryApplicationConfigurationDraftRepository{drafts: make(map[string]ApplicationConfigurationDraft)}
}

func (service applicationConfigurationDraftService) Validate(
	requestContext ApplicationConfigurationDraftContext,
	payload ApplicationConfigurationDraftPayload,
) ApplicationConfigurationDraftResult {
	validation := validateApplicationConfigurationDraftPayload(requestContext, payload)
	return ApplicationConfigurationDraftResult{ValidationSummary: validation}
}

func (service applicationConfigurationDraftService) Save(
	requestContext ApplicationConfigurationDraftContext,
	payload ApplicationConfigurationDraftPayload,
	expectedVersion int,
) ApplicationConfigurationDraftResult {
	validation := validateApplicationConfigurationDraftPayload(requestContext, payload)
	if !validation.IsValid {
		failure := ApplicationDraftFailurePayloadInvalid
		for _, finding := range validation.Findings {
			if finding.Code == ApplicationDraftFailureScopeDenied {
				failure = ApplicationDraftFailureScopeDenied
				break
			}
			if finding.Code == ApplicationDraftFailureSecretForbidden {
				failure = ApplicationDraftFailureSecretForbidden
				break
			}
		}
		return ApplicationConfigurationDraftResult{FailureCode: failure, ValidationSummary: validation}
	}
	if !requestContext.WriteEnabled {
		return ApplicationConfigurationDraftResult{FailureCode: ApplicationDraftFailureWriteDisabled, ValidationSummary: validation}
	}
	if expectedVersion < 0 {
		return ApplicationConfigurationDraftResult{FailureCode: ApplicationDraftFailurePayloadInvalid, ValidationSummary: validation}
	}
	now := service.now().Format(time.RFC3339)
	draft := ApplicationConfigurationDraft{
		ApplicationConfigurationDraftPayload: normalizeApplicationConfigurationDraftPayload(payload),
		DraftVersion:                         expectedVersion + 1,
		ValidationSummary:                    validation,
		CreatedAt:                            now,
		UpdatedAt:                            now,
		CreatedByActorRef:                    requestContext.ActorRef,
		UpdatedByActorRef:                    requestContext.ActorRef,
		RequestID:                            requestContext.RequestID,
		AuditRef:                             requestContext.AuditRef,
	}
	saved, err := service.repository.Save(requestContext, draft, expectedVersion)
	if err != nil {
		return applicationConfigurationDraftRepositoryFailure(err, validation)
	}
	return ApplicationConfigurationDraftResult{
		Draft:               &saved,
		CurrentDraftVersion: saved.DraftVersion,
		ValidationSummary:   saved.ValidationSummary,
	}
}

func (service applicationConfigurationDraftService) Read(
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) ApplicationConfigurationDraftResult {
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(draftID)) {
		return ApplicationConfigurationDraftResult{FailureCode: ApplicationDraftFailurePayloadInvalid, ValidationSummary: invalidApplicationDraftValidation("draft_id", "Draft id is invalid.")}
	}
	draft, err := service.repository.Read(requestContext, strings.TrimSpace(draftID))
	if err != nil {
		return applicationConfigurationDraftRepositoryFailure(err, ApplicationConfigurationDraftValidation{})
	}
	return ApplicationConfigurationDraftResult{Draft: &draft, CurrentDraftVersion: draft.DraftVersion, ValidationSummary: draft.ValidationSummary}
}

func (service applicationConfigurationDraftService) List(
	requestContext ApplicationConfigurationDraftContext,
) ([]ApplicationConfigurationDraftSummary, string) {
	summaries, err := service.repository.List(requestContext)
	if err != nil {
		return []ApplicationConfigurationDraftSummary{}, ApplicationDraftFailureStoreUnavailable
	}
	return summaries, ""
}

func validateApplicationConfigurationDraftPayload(
	requestContext ApplicationConfigurationDraftContext,
	payload ApplicationConfigurationDraftPayload,
) ApplicationConfigurationDraftValidation {
	findings := make([]ApplicationConfigurationDraftValidationFinding, 0)
	add := func(code, field, summary string) {
		findings = append(findings, ApplicationConfigurationDraftValidationFinding{Code: code, Field: field, Summary: summary})
	}
	if strings.TrimSpace(payload.WorkspaceID) == "" || strings.TrimSpace(payload.ApplicationID) == "" ||
		strings.TrimSpace(payload.WorkspaceID) != requestContext.WorkspaceID || strings.TrimSpace(payload.ApplicationID) != requestContext.ApplicationID {
		add(ApplicationDraftFailureScopeDenied, "scope", "Draft workspace and application must match the current request scope.")
	}
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(payload.DraftID)) {
		add(ApplicationDraftFailurePayloadInvalid, "draft_id", "Draft id must be a stable identifier up to 160 characters.")
	}
	if strings.TrimSpace(payload.SchemaVersion) != applicationConfigurationDraftSchemaVersion {
		add(ApplicationDraftFailurePayloadInvalid, "schema_version", "Application draft schema version is unsupported.")
	}
	if length := len(strings.TrimSpace(payload.DisplayName)); length < 2 || length > 120 {
		add(ApplicationDraftFailurePayloadInvalid, "display_name", "Display name must contain 2 to 120 characters.")
	}
	if len(strings.TrimSpace(payload.Description)) > 1000 {
		add(ApplicationDraftFailurePayloadInvalid, "description", "Description must not exceed 1000 characters.")
	}
	allowedKinds := map[string]bool{"workflow_copilot": true, "docs_qa": true, "agent": true, "prompt_application": true}
	if !allowedKinds[strings.TrimSpace(payload.ApplicationKind)] {
		add(ApplicationDraftFailurePayloadInvalid, "application_kind", "Application kind is not supported by the current workspace.")
	}
	protocols := normalizeApplicationDraftProtocols(payload.AllowedProtocols)
	if len(protocols) == 0 || len(protocols) != len(payload.AllowedProtocols) {
		add(ApplicationDraftFailurePayloadInvalid, "allowed_protocols", "Allowed protocols must be unique supported protocol identifiers.")
	}
	defaultProtocol := strings.TrimSpace(payload.DefaultProtocol)
	if !containsApplicationDraftProtocol(protocols, defaultProtocol) {
		add(ApplicationDraftFailurePayloadInvalid, "default_protocol", "Default protocol must be included in allowed protocols.")
	}
	if !applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(payload.DefaultModel)) {
		add(ApplicationDraftFailurePayloadInvalid, "default_model", "Default model must be a validated model identifier.")
	}
	for field, value := range map[string]string{
		"display_name":     payload.DisplayName,
		"description":      payload.Description,
		"application_kind": payload.ApplicationKind,
		"default_model":    payload.DefaultModel,
	} {
		if applicationDraftStringContainsSecret(value) {
			add(ApplicationDraftFailureSecretForbidden, field, "Secret or internal caller material is forbidden in application drafts.")
		}
	}
	state := applicationDraftValidationValid
	if len(findings) > 0 {
		state = applicationDraftValidationInvalid
	}
	return ApplicationConfigurationDraftValidation{State: state, IsValid: len(findings) == 0, Findings: findings}
}

func normalizeApplicationConfigurationDraftPayload(payload ApplicationConfigurationDraftPayload) ApplicationConfigurationDraftPayload {
	payload.DraftID = strings.TrimSpace(payload.DraftID)
	payload.WorkspaceID = strings.TrimSpace(payload.WorkspaceID)
	payload.ApplicationID = strings.TrimSpace(payload.ApplicationID)
	payload.BaseApplicationUpdatedAt = strings.TrimSpace(payload.BaseApplicationUpdatedAt)
	payload.SchemaVersion = strings.TrimSpace(payload.SchemaVersion)
	payload.DisplayName = strings.TrimSpace(payload.DisplayName)
	payload.Description = strings.TrimSpace(payload.Description)
	payload.ApplicationKind = strings.TrimSpace(payload.ApplicationKind)
	payload.DefaultProtocol = strings.TrimSpace(payload.DefaultProtocol)
	payload.DefaultModel = strings.TrimSpace(payload.DefaultModel)
	payload.AllowedProtocols = normalizeApplicationDraftProtocols(payload.AllowedProtocols)
	return payload
}

func normalizeApplicationDraftProtocols(protocols []string) []string {
	normalized := make([]string, 0, len(protocols))
	seen := make(map[string]bool)
	for _, protocol := range protocols {
		value := strings.TrimSpace(protocol)
		if !isApplicationDraftProtocol(value) || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized
}

func isApplicationDraftProtocol(protocol string) bool {
	return protocol == "chat_completions" || protocol == "responses" || protocol == "messages"
}

func containsApplicationDraftProtocol(protocols []string, protocol string) bool {
	for _, candidate := range protocols {
		if candidate == protocol {
			return true
		}
	}
	return false
}

func applicationDraftStringContainsSecret(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{"authorization:", "bearer ", "api_key=", "api-key=", "x-radishmind-dev-", "sk-"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func invalidApplicationDraftValidation(field, summary string) ApplicationConfigurationDraftValidation {
	return ApplicationConfigurationDraftValidation{
		State:    applicationDraftValidationInvalid,
		Findings: []ApplicationConfigurationDraftValidationFinding{{Code: ApplicationDraftFailurePayloadInvalid, Field: field, Summary: summary}},
	}
}

func applicationConfigurationDraftRepositoryFailure(err error, validation ApplicationConfigurationDraftValidation) ApplicationConfigurationDraftResult {
	failureCode := ApplicationDraftFailureStoreUnavailable
	if errors.Is(err, errApplicationDraftNotFound) {
		failureCode = ApplicationDraftFailureNotFound
	}
	if errors.Is(err, errApplicationDraftVersionConflict) {
		failureCode = ApplicationDraftFailureVersionConflict
	}
	result := ApplicationConfigurationDraftResult{FailureCode: failureCode, ValidationSummary: validation}
	var conflict applicationDraftVersionConflictError
	if errors.As(err, &conflict) {
		result.CurrentDraftVersion = conflict.CurrentVersion
	}
	return result
}

func applicationConfigurationDraftRepositoryKey(context ApplicationConfigurationDraftContext, draftID string) string {
	return strings.Join([]string{context.TenantRef, context.WorkspaceID, context.ApplicationID, context.OwnerSubjectRef, draftID}, "\x00")
}

func (repository *memoryApplicationConfigurationDraftRepository) Save(
	requestContext ApplicationConfigurationDraftContext,
	draft ApplicationConfigurationDraft,
	expectedVersion int,
) (ApplicationConfigurationDraft, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	key := applicationConfigurationDraftRepositoryKey(requestContext, draft.DraftID)
	current, exists := repository.drafts[key]
	if expectedVersion > 0 && !exists {
		return ApplicationConfigurationDraft{}, errApplicationDraftNotFound
	}
	if expectedVersion == 0 && exists || expectedVersion > 0 && current.DraftVersion != expectedVersion {
		return ApplicationConfigurationDraft{}, applicationDraftVersionConflictError{CurrentVersion: current.DraftVersion}
	}
	if exists {
		draft.CreatedAt = current.CreatedAt
		draft.CreatedByActorRef = current.CreatedByActorRef
	}
	repository.drafts[key] = draft
	return draft, nil
}

func (repository *memoryApplicationConfigurationDraftRepository) Read(
	requestContext ApplicationConfigurationDraftContext,
	draftID string,
) (ApplicationConfigurationDraft, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationConfigurationDraft{}, errApplicationDraftStoreUnavailable
	}
	draft, exists := repository.drafts[applicationConfigurationDraftRepositoryKey(requestContext, draftID)]
	if !exists {
		return ApplicationConfigurationDraft{}, errApplicationDraftNotFound
	}
	return draft, nil
}

func (repository *memoryApplicationConfigurationDraftRepository) List(
	requestContext ApplicationConfigurationDraftContext,
) ([]ApplicationConfigurationDraftSummary, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return nil, errApplicationDraftStoreUnavailable
	}
	prefix := applicationConfigurationDraftRepositoryKey(requestContext, "")
	summaries := make([]ApplicationConfigurationDraftSummary, 0)
	for key, draft := range repository.drafts {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		summaries = append(summaries, applicationConfigurationDraftSummary(draft))
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].UpdatedAt == summaries[j].UpdatedAt {
			return summaries[i].DraftID < summaries[j].DraftID
		}
		return summaries[i].UpdatedAt > summaries[j].UpdatedAt
	})
	return summaries, nil
}

func applicationConfigurationDraftSummary(draft ApplicationConfigurationDraft) ApplicationConfigurationDraftSummary {
	return ApplicationConfigurationDraftSummary{
		DraftID: draft.DraftID, ApplicationID: draft.ApplicationID, DraftVersion: draft.DraftVersion,
		DisplayName: draft.DisplayName, ApplicationKind: draft.ApplicationKind, DefaultProtocol: draft.DefaultProtocol,
		DefaultModel: draft.DefaultModel, ValidationState: draft.ValidationSummary.State, UpdatedAt: draft.UpdatedAt,
		UpdatedByActorRef: draft.UpdatedByActorRef,
	}
}
