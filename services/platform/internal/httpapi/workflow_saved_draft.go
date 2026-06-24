package httpapi

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	savedWorkflowDraftSchemaVersion                 = "saved_workflow_draft.v1"
	savedWorkflowDraftDesignerLayoutAdditionalField = "designer_layout_v1"
	savedWorkflowDraftDesignerLayoutVersion         = "designer_layout_v1"
	savedWorkflowDraftDesignerLayoutSource          = "workflow_node_designer"
	savedWorkflowDraftDesignerLayoutPersistence     = "saved_draft_metadata"

	maxSavedWorkflowDraftNodes            = 50
	maxSavedWorkflowDraftEdges            = 120
	maxSavedWorkflowDraftTextLength       = 4000
	maxSavedWorkflowDraftLabelLength      = 160
	maxSavedWorkflowDraftLayoutCoordinate = 10000
)

type SavedWorkflowDraftFailureCode string

const (
	SavedWorkflowDraftFailureScopeDenied                SavedWorkflowDraftFailureCode = "draft_scope_denied"
	SavedWorkflowDraftFailureNotFound                   SavedWorkflowDraftFailureCode = "draft_not_found"
	SavedWorkflowDraftFailureSchemaVersionUnsupported   SavedWorkflowDraftFailureCode = "draft_schema_version_unsupported"
	SavedWorkflowDraftFailurePayloadInvalid             SavedWorkflowDraftFailureCode = "draft_payload_invalid"
	SavedWorkflowDraftFailureGraphInvalid               SavedWorkflowDraftFailureCode = "draft_graph_invalid"
	SavedWorkflowDraftFailureContractInvalid            SavedWorkflowDraftFailureCode = "draft_contract_invalid"
	SavedWorkflowDraftFailureBlockedCapability          SavedWorkflowDraftFailureCode = "draft_blocked_capability"
	SavedWorkflowDraftFailureVersionConflict            SavedWorkflowDraftFailureCode = "draft_version_conflict"
	SavedWorkflowDraftFailurePayloadTooLarge            SavedWorkflowDraftFailureCode = "draft_payload_too_large"
	SavedWorkflowDraftFailureStoreUnavailable           SavedWorkflowDraftFailureCode = "draft_store_unavailable"
	SavedWorkflowDraftFailureStoreContractMismatch      SavedWorkflowDraftFailureCode = "draft_store_contract_mismatch"
	SavedWorkflowDraftFailureWriteDisabled              SavedWorkflowDraftFailureCode = "draft_write_disabled"
	SavedWorkflowDraftFailureRepositoryStoreDisabled    SavedWorkflowDraftFailureCode = "repository_store_disabled"
	SavedWorkflowDraftFailureInvalidStoreMode           SavedWorkflowDraftFailureCode = "invalid_draft_store_mode"
	SavedWorkflowDraftFailureAuthContextMismatch        SavedWorkflowDraftFailureCode = "draft_auth_context_contract_mismatch"
	SavedWorkflowDraftFailureSchemaMigrationNotApplied  SavedWorkflowDraftFailureCode = "draft_schema_migration_not_applied"
	SavedWorkflowDraftFailureStoreSchemaVersionMismatch SavedWorkflowDraftFailureCode = "draft_store_schema_version_mismatch"
	SavedWorkflowDraftFailureStoreMigrationUnavailable  SavedWorkflowDraftFailureCode = "draft_store_migration_unavailable"
	SavedWorkflowDraftFailureIdentityContextMissing     SavedWorkflowDraftFailureCode = "draft_identity_context_missing"
	SavedWorkflowDraftFailureTenantBindingMissing       SavedWorkflowDraftFailureCode = "draft_tenant_binding_missing"
	SavedWorkflowDraftFailureWorkspaceMembershipDenied  SavedWorkflowDraftFailureCode = "draft_workspace_membership_denied"
	SavedWorkflowDraftFailureApplicationScopeDenied     SavedWorkflowDraftFailureCode = "draft_application_scope_denied"
	SavedWorkflowDraftFailureOwnerScopeDenied           SavedWorkflowDraftFailureCode = "draft_owner_scope_denied"
	SavedWorkflowDraftFailureScopeGrantMissing          SavedWorkflowDraftFailureCode = "draft_scope_grant_missing"
	SavedWorkflowDraftFailureAuditContextMissing        SavedWorkflowDraftFailureCode = "draft_audit_context_missing"
)

type SavedWorkflowDraftStatus string

const (
	SavedWorkflowDraftStatusValidForReview    SavedWorkflowDraftStatus = "valid_for_review"
	SavedWorkflowDraftStatusInvalidDraft      SavedWorkflowDraftStatus = "invalid_draft"
	SavedWorkflowDraftStatusBlockedCapability SavedWorkflowDraftStatus = "blocked_capability"
	SavedWorkflowDraftStatusSchemaUnsupported SavedWorkflowDraftStatus = "schema_unsupported"
)

type SavedWorkflowDraftValidationSeverity string

const (
	SavedWorkflowDraftValidationInfo     SavedWorkflowDraftValidationSeverity = "info"
	SavedWorkflowDraftValidationWarning  SavedWorkflowDraftValidationSeverity = "warning"
	SavedWorkflowDraftValidationBlocking SavedWorkflowDraftValidationSeverity = "blocking"
)

type SavedWorkflowDraftContext struct {
	RequestID     string
	WorkspaceID   string
	ApplicationID string
	ActorRef      string
	AuditRef      string
	WriteEnabled  bool
}

type SaveWorkflowDraftRequest struct {
	ExpectedDraftVersion int
	Payload              SavedWorkflowDraftPayload
}

type ReadWorkflowDraftRequest struct {
	DraftID string
}

type ListWorkflowDraftsRequest struct{}

type ValidateWorkflowDraftRequest struct {
	Payload SavedWorkflowDraftPayload
}

type SavedWorkflowDraftPayload struct {
	DraftID               string
	WorkspaceID           string
	ApplicationID         string
	SourceDefinitionID    string
	BaseDefinitionVersion int
	SchemaVersion         string
	DraftStatus           SavedWorkflowDraftStatus
	Name                  string
	Description           string
	Nodes                 []SavedWorkflowDraftNode
	Edges                 []SavedWorkflowDraftEdge
	InputContract         SavedWorkflowDraftContract
	OutputContract        SavedWorkflowDraftContract
	ProviderRefs          []string
	ToolRefs              []string
	RAGRefs               []string
	RequestedCapabilities []string
	AdditionalFields      map[string]any
}

type SavedWorkflowDraftNode struct {
	NodeID               string
	NodeType             string
	Label                string
	InputSummary         string
	OutputSummary        string
	InputContractRef     string
	OutputContractRef    string
	InputContractFields  []string
	OutputContractFields []string
	OutputMappingSummary string
	ProviderRef          string
	ToolRef              string
	RAGRef               string
	RiskLevel            string
	RequiresConfirmation bool
}

type SavedWorkflowDraftEdge struct {
	EdgeID           string
	FromNodeID       string
	ToNodeID         string
	ConditionSummary string
}

type SavedWorkflowDraftContract struct {
	ContractID     string
	RequiredFields []string
	Summary        string
}

type SavedWorkflowDraft struct {
	DraftID                    string
	WorkspaceID                string
	ApplicationID              string
	SourceDefinitionID         string
	BaseDefinitionVersion      int
	DraftVersion               int
	SchemaVersion              string
	DraftStatus                SavedWorkflowDraftStatus
	CreatedAt                  string
	UpdatedAt                  string
	CreatedByActorRef          string
	UpdatedByActorRef          string
	Name                       string
	Description                string
	Nodes                      []SavedWorkflowDraftNode
	Edges                      []SavedWorkflowDraftEdge
	InputContract              SavedWorkflowDraftContract
	OutputContract             SavedWorkflowDraftContract
	ProviderRefs               []string
	ToolRefs                   []string
	RAGRefs                    []string
	RequestedCapabilities      []string
	AdditionalFields           map[string]any
	ValidationSummary          SavedWorkflowDraftValidationSummary
	BlockedCapabilitySummary   []SavedWorkflowDraftBlockedCapability
	RequestAuditMetadata       SavedWorkflowDraftAuditMetadata
	SampleOrUnsavedDraftStatus string
}

type SavedWorkflowDraftValidationSummary struct {
	ValidationState SavedWorkflowDraftStatus
	ValidForReview  bool
	Findings        []SavedWorkflowDraftValidationFinding
}

type SavedWorkflowDraftValidationFinding struct {
	Code       SavedWorkflowDraftFailureCode
	Severity   SavedWorkflowDraftValidationSeverity
	Field      string
	Summary    string
	EvidenceID string
}

type SavedWorkflowDraftBlockedCapability struct {
	CapabilityID        string
	MissingPrerequisite string
	Summary             string
}

type SavedWorkflowDraftAuditMetadata struct {
	RequestID string
	AuditRef  string
	ActorRef  string
}

type SavedWorkflowDraftResult struct {
	Draft                *SavedWorkflowDraft
	FailureCode          SavedWorkflowDraftFailureCode
	CurrentDraftVersion  int
	ValidationSummary    SavedWorkflowDraftValidationSummary
	BlockedCapabilities  []SavedWorkflowDraftBlockedCapability
	RequestAuditMetadata SavedWorkflowDraftAuditMetadata
}

type SavedWorkflowDraftSummary struct {
	DraftID                    string
	WorkspaceID                string
	ApplicationID              string
	SourceDefinitionID         string
	DraftVersion               int
	SchemaVersion              string
	DraftStatus                SavedWorkflowDraftStatus
	Name                       string
	Description                string
	UpdatedAt                  string
	UpdatedByActorRef          string
	NodeCount                  int
	EdgeCount                  int
	BlockedCapabilityCount     int
	ValidationState            SavedWorkflowDraftStatus
	ValidForReview             bool
	SampleOrUnsavedDraftStatus string
}

type SavedWorkflowDraftListResult struct {
	Summaries            []SavedWorkflowDraftSummary
	FailureCode          SavedWorkflowDraftFailureCode
	RequestAuditMetadata SavedWorkflowDraftAuditMetadata
}

type savedWorkflowDraftStore interface {
	ReadDraftByID(draftID string) (SavedWorkflowDraft, bool, error)
	ListDraftsByScope(workspaceID string, applicationID string) ([]SavedWorkflowDraft, error)
	WriteDraft(draft SavedWorkflowDraft) error
	SideEffects() SavedWorkflowDraftSideEffects
}

type SavedWorkflowDraftSideEffects struct {
	DraftWriteCount          int
	ExecutorCallCount        int
	ConfirmationCallCount    int
	BusinessWritebackCount   int
	ReplayCallCount          int
	MaterializedResultReads  int
	ExternalRepositoryWrites int
}

type memorySavedWorkflowDraftStore struct {
	drafts      map[string]SavedWorkflowDraft
	sideEffects SavedWorkflowDraftSideEffects
	unavailable bool
}

type savedWorkflowDraftService struct {
	store savedWorkflowDraftStore
	now   func() time.Time
}

func newSavedWorkflowDraftService(store savedWorkflowDraftStore) savedWorkflowDraftService {
	return savedWorkflowDraftService{
		store: store,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func newMemorySavedWorkflowDraftStore() *memorySavedWorkflowDraftStore {
	return &memorySavedWorkflowDraftStore{
		drafts: make(map[string]SavedWorkflowDraft),
	}
}

func (service savedWorkflowDraftService) SaveDraft(
	context SavedWorkflowDraftContext,
	request SaveWorkflowDraftRequest,
) SavedWorkflowDraftResult {
	auditMetadata := savedWorkflowDraftAuditMetadata(context)
	if !context.WriteEnabled {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureWriteDisabled, auditMetadata)
	}

	normalized, validationResult := service.validatePayload(context, request.Payload)
	if validationResult.FailureCode != "" {
		validationResult.RequestAuditMetadata = auditMetadata
		return validationResult
	}

	existing, found, err := service.store.ReadDraftByID(normalized.DraftID)
	if err != nil {
		return savedWorkflowDraftFailure(savedWorkflowDraftStoreFailureCode(err), auditMetadata)
	}
	if found && !savedWorkflowDraftMatchesScope(existing, context.WorkspaceID, context.ApplicationID) {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
	}
	if found && request.ExpectedDraftVersion != existing.DraftVersion {
		result := savedWorkflowDraftFailure(SavedWorkflowDraftFailureVersionConflict, auditMetadata)
		result.CurrentDraftVersion = existing.DraftVersion
		return result
	}
	if !found && request.ExpectedDraftVersion != 0 {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureNotFound, auditMetadata)
	}

	now := service.now().UTC().Format(time.RFC3339)
	draftVersion := 1
	createdAt := now
	createdByActorRef := context.ActorRef
	if found {
		draftVersion = existing.DraftVersion + 1
		createdAt = existing.CreatedAt
		createdByActorRef = existing.CreatedByActorRef
	}

	draft := SavedWorkflowDraft{
		DraftID:                    normalized.DraftID,
		WorkspaceID:                normalized.WorkspaceID,
		ApplicationID:              normalized.ApplicationID,
		SourceDefinitionID:         normalized.SourceDefinitionID,
		BaseDefinitionVersion:      normalized.BaseDefinitionVersion,
		DraftVersion:               draftVersion,
		SchemaVersion:              normalized.SchemaVersion,
		DraftStatus:                validationResult.ValidationSummary.ValidationState,
		CreatedAt:                  createdAt,
		UpdatedAt:                  now,
		CreatedByActorRef:          createdByActorRef,
		UpdatedByActorRef:          context.ActorRef,
		Name:                       normalized.Name,
		Description:                normalized.Description,
		Nodes:                      cloneSavedWorkflowDraftNodes(normalized.Nodes),
		Edges:                      cloneSavedWorkflowDraftEdges(normalized.Edges),
		InputContract:              cloneSavedWorkflowDraftContract(normalized.InputContract),
		OutputContract:             cloneSavedWorkflowDraftContract(normalized.OutputContract),
		ProviderRefs:               cloneStringSlice(normalized.ProviderRefs),
		ToolRefs:                   cloneStringSlice(normalized.ToolRefs),
		RAGRefs:                    cloneStringSlice(normalized.RAGRefs),
		RequestedCapabilities:      cloneStringSlice(normalized.RequestedCapabilities),
		AdditionalFields:           cloneSavedWorkflowDraftAdditionalFields(normalized.AdditionalFields),
		ValidationSummary:          validationResult.ValidationSummary,
		BlockedCapabilitySummary:   cloneSavedWorkflowDraftBlockedCapabilities(validationResult.BlockedCapabilities),
		RequestAuditMetadata:       auditMetadata,
		SampleOrUnsavedDraftStatus: "saved_draft_record",
	}
	if err := service.store.WriteDraft(draft); err != nil {
		return savedWorkflowDraftFailure(savedWorkflowDraftStoreFailureCode(err), auditMetadata)
	}

	return SavedWorkflowDraftResult{
		Draft:                cloneSavedWorkflowDraftPointer(draft),
		CurrentDraftVersion:  draft.DraftVersion,
		ValidationSummary:    draft.ValidationSummary,
		BlockedCapabilities:  cloneSavedWorkflowDraftBlockedCapabilities(draft.BlockedCapabilitySummary),
		RequestAuditMetadata: auditMetadata,
	}
}

func (service savedWorkflowDraftService) ReadDraft(
	context SavedWorkflowDraftContext,
	request ReadWorkflowDraftRequest,
) SavedWorkflowDraftResult {
	auditMetadata := savedWorkflowDraftAuditMetadata(context)
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailurePayloadInvalid, auditMetadata)
	}
	if strings.TrimSpace(context.WorkspaceID) == "" || strings.TrimSpace(context.ApplicationID) == "" {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
	}

	draft, found, err := service.store.ReadDraftByID(draftID)
	if err != nil {
		return savedWorkflowDraftFailure(savedWorkflowDraftStoreFailureCode(err), auditMetadata)
	}
	if !found {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureNotFound, auditMetadata)
	}
	if !savedWorkflowDraftMatchesScope(draft, context.WorkspaceID, context.ApplicationID) {
		return savedWorkflowDraftFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
	}
	if draft.SchemaVersion != savedWorkflowDraftSchemaVersion {
		result := savedWorkflowDraftFailure(SavedWorkflowDraftFailureSchemaVersionUnsupported, auditMetadata)
		result.CurrentDraftVersion = draft.DraftVersion
		result.ValidationSummary = SavedWorkflowDraftValidationSummary{
			ValidationState: SavedWorkflowDraftStatusSchemaUnsupported,
			ValidForReview:  false,
			Findings: []SavedWorkflowDraftValidationFinding{
				{
					Code:     SavedWorkflowDraftFailureSchemaVersionUnsupported,
					Severity: SavedWorkflowDraftValidationBlocking,
					Field:    "schema_version",
					Summary:  "Saved workflow draft schema version is not supported by the current policy.",
				},
			},
		}
		return result
	}

	payload := savedWorkflowDraftPayloadFromDraft(draft)
	_, validationResult := service.validatePayload(context, payload)
	if validationResult.FailureCode != "" {
		validationResult.RequestAuditMetadata = auditMetadata
		validationResult.CurrentDraftVersion = draft.DraftVersion
		return validationResult
	}
	draft.ValidationSummary = validationResult.ValidationSummary
	draft.BlockedCapabilitySummary = cloneSavedWorkflowDraftBlockedCapabilities(validationResult.BlockedCapabilities)
	draft.RequestAuditMetadata = auditMetadata
	draft.SampleOrUnsavedDraftStatus = "saved_draft_record"

	return SavedWorkflowDraftResult{
		Draft:                cloneSavedWorkflowDraftPointer(draft),
		CurrentDraftVersion:  draft.DraftVersion,
		ValidationSummary:    draft.ValidationSummary,
		BlockedCapabilities:  cloneSavedWorkflowDraftBlockedCapabilities(draft.BlockedCapabilitySummary),
		RequestAuditMetadata: auditMetadata,
	}
}

func (service savedWorkflowDraftService) ListDrafts(
	context SavedWorkflowDraftContext,
	_ ListWorkflowDraftsRequest,
) SavedWorkflowDraftListResult {
	auditMetadata := savedWorkflowDraftAuditMetadata(context)
	if strings.TrimSpace(context.WorkspaceID) == "" || strings.TrimSpace(context.ApplicationID) == "" {
		return savedWorkflowDraftListFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
	}

	drafts, err := service.store.ListDraftsByScope(context.WorkspaceID, context.ApplicationID)
	if err != nil {
		return savedWorkflowDraftListFailure(savedWorkflowDraftStoreFailureCode(err), auditMetadata)
	}

	summaries := make([]SavedWorkflowDraftSummary, 0, len(drafts))
	for _, draft := range drafts {
		if !savedWorkflowDraftMatchesScope(draft, context.WorkspaceID, context.ApplicationID) {
			return savedWorkflowDraftListFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
		}
		summaries = append(summaries, savedWorkflowDraftSummaryFromDraft(draft))
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].UpdatedAt == summaries[j].UpdatedAt {
			return summaries[i].DraftID < summaries[j].DraftID
		}
		return summaries[i].UpdatedAt > summaries[j].UpdatedAt
	})

	return SavedWorkflowDraftListResult{
		Summaries:            summaries,
		RequestAuditMetadata: auditMetadata,
	}
}

func (service savedWorkflowDraftService) ValidateDraft(
	context SavedWorkflowDraftContext,
	request ValidateWorkflowDraftRequest,
) SavedWorkflowDraftResult {
	auditMetadata := savedWorkflowDraftAuditMetadata(context)
	_, result := service.validatePayload(context, request.Payload)
	result.RequestAuditMetadata = auditMetadata
	return result
}

func (service savedWorkflowDraftService) validatePayload(
	context SavedWorkflowDraftContext,
	payload SavedWorkflowDraftPayload,
) (SavedWorkflowDraftPayload, SavedWorkflowDraftResult) {
	if forbiddenField := firstSavedWorkflowDraftForbiddenField(payload.AdditionalFields); forbiddenField != "" {
		auditMetadata := savedWorkflowDraftAuditMetadata(context)
		result := savedWorkflowDraftFailure(SavedWorkflowDraftFailurePayloadInvalid, auditMetadata)
		result.ValidationSummary = SavedWorkflowDraftValidationSummary{
			ValidationState: SavedWorkflowDraftStatusInvalidDraft,
			Findings: []SavedWorkflowDraftValidationFinding{
				{
					Code:     SavedWorkflowDraftFailurePayloadInvalid,
					Severity: SavedWorkflowDraftValidationBlocking,
					Field:    forbiddenField,
					Summary:  "Saved draft payload contains a forbidden field that cannot be sanitized safely.",
				},
			},
		}
		return normalizeSavedWorkflowDraftPayload(payload), result
	}

	normalized := normalizeSavedWorkflowDraftPayload(payload)
	auditMetadata := savedWorkflowDraftAuditMetadata(context)
	findings := make([]SavedWorkflowDraftValidationFinding, 0)
	blockedCapabilities := make([]SavedWorkflowDraftBlockedCapability, 0)

	if strings.TrimSpace(context.WorkspaceID) == "" ||
		strings.TrimSpace(context.ApplicationID) == "" ||
		strings.TrimSpace(normalized.WorkspaceID) != strings.TrimSpace(context.WorkspaceID) ||
		strings.TrimSpace(normalized.ApplicationID) != strings.TrimSpace(context.ApplicationID) {
		return normalized, savedWorkflowDraftFailure(SavedWorkflowDraftFailureScopeDenied, auditMetadata)
	}
	if normalized.DraftID == "" || normalized.Name == "" {
		return normalized, savedWorkflowDraftFailure(SavedWorkflowDraftFailurePayloadInvalid, auditMetadata)
	}
	if normalized.SchemaVersion != savedWorkflowDraftSchemaVersion {
		return normalized, savedWorkflowDraftFailure(SavedWorkflowDraftFailureSchemaVersionUnsupported, auditMetadata)
	}
	if savedWorkflowDraftPayloadTooLarge(normalized) {
		return normalized, savedWorkflowDraftFailure(SavedWorkflowDraftFailurePayloadTooLarge, auditMetadata)
	}
	if len(normalized.Nodes) == 0 || len(normalized.Edges) == 0 {
		return normalized, savedWorkflowDraftFailure(SavedWorkflowDraftFailureGraphInvalid, auditMetadata)
	}

	graphFindings, graphBlockedCapabilities, graphHardFailure := validateSavedWorkflowDraftGraph(normalized)
	if graphHardFailure != "" {
		return normalized, savedWorkflowDraftFailure(graphHardFailure, auditMetadata)
	}
	findings = append(findings, graphFindings...)
	blockedCapabilities = append(blockedCapabilities, graphBlockedCapabilities...)

	contractFindings := validateSavedWorkflowDraftContracts(normalized)
	findings = append(findings, contractFindings...)

	capabilityFindings, capabilityBlockedCapabilities := validateSavedWorkflowDraftCapabilities(normalized)
	findings = append(findings, capabilityFindings...)
	blockedCapabilities = append(blockedCapabilities, capabilityBlockedCapabilities...)

	riskFindings := validateSavedWorkflowDraftRisk(normalized)
	findings = append(findings, riskFindings...)

	summary := SavedWorkflowDraftValidationSummary{
		ValidationState: deriveSavedWorkflowDraftValidationState(findings, blockedCapabilities),
		Findings:        findings,
	}
	summary.ValidForReview = summary.ValidationState == SavedWorkflowDraftStatusValidForReview

	return normalized, SavedWorkflowDraftResult{
		ValidationSummary:    summary,
		BlockedCapabilities:  cloneSavedWorkflowDraftBlockedCapabilities(blockedCapabilities),
		RequestAuditMetadata: auditMetadata,
	}
}

func (store *memorySavedWorkflowDraftStore) ReadDraftByID(draftID string) (SavedWorkflowDraft, bool, error) {
	if store.unavailable {
		return SavedWorkflowDraft{}, false, errors.New("saved workflow draft store unavailable")
	}
	draft, found := store.drafts[draftID]
	if !found {
		return SavedWorkflowDraft{}, false, nil
	}
	return cloneSavedWorkflowDraft(draft), true, nil
}

func (store *memorySavedWorkflowDraftStore) ListDraftsByScope(
	workspaceID string,
	applicationID string,
) ([]SavedWorkflowDraft, error) {
	if store.unavailable {
		return nil, errors.New("saved workflow draft store unavailable")
	}
	drafts := make([]SavedWorkflowDraft, 0)
	for _, draft := range store.drafts {
		if savedWorkflowDraftMatchesScope(draft, workspaceID, applicationID) {
			drafts = append(drafts, cloneSavedWorkflowDraft(draft))
		}
	}
	sort.Slice(drafts, func(i, j int) bool {
		if drafts[i].UpdatedAt == drafts[j].UpdatedAt {
			return drafts[i].DraftID < drafts[j].DraftID
		}
		return drafts[i].UpdatedAt > drafts[j].UpdatedAt
	})
	return drafts, nil
}

func (store *memorySavedWorkflowDraftStore) WriteDraft(draft SavedWorkflowDraft) error {
	if store.unavailable {
		return errors.New("saved workflow draft store unavailable")
	}
	store.drafts[draft.DraftID] = cloneSavedWorkflowDraft(draft)
	store.sideEffects.DraftWriteCount++
	return nil
}

func (store *memorySavedWorkflowDraftStore) SideEffects() SavedWorkflowDraftSideEffects {
	return store.sideEffects
}

func savedWorkflowDraftAuditMetadata(context SavedWorkflowDraftContext) SavedWorkflowDraftAuditMetadata {
	return SavedWorkflowDraftAuditMetadata{
		RequestID: strings.TrimSpace(context.RequestID),
		AuditRef:  strings.TrimSpace(context.AuditRef),
		ActorRef:  strings.TrimSpace(context.ActorRef),
	}
}

func savedWorkflowDraftFailure(
	failureCode SavedWorkflowDraftFailureCode,
	auditMetadata SavedWorkflowDraftAuditMetadata,
) SavedWorkflowDraftResult {
	return SavedWorkflowDraftResult{
		FailureCode:          failureCode,
		RequestAuditMetadata: auditMetadata,
	}
}

func savedWorkflowDraftListFailure(
	failureCode SavedWorkflowDraftFailureCode,
	auditMetadata SavedWorkflowDraftAuditMetadata,
) SavedWorkflowDraftListResult {
	return SavedWorkflowDraftListResult{
		FailureCode:          failureCode,
		RequestAuditMetadata: auditMetadata,
	}
}

func normalizeSavedWorkflowDraftPayload(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
	normalized := payload
	normalized.DraftID = strings.TrimSpace(payload.DraftID)
	normalized.WorkspaceID = strings.TrimSpace(payload.WorkspaceID)
	normalized.ApplicationID = strings.TrimSpace(payload.ApplicationID)
	normalized.SourceDefinitionID = strings.TrimSpace(payload.SourceDefinitionID)
	normalized.SchemaVersion = strings.TrimSpace(payload.SchemaVersion)
	normalized.Name = strings.TrimSpace(payload.Name)
	normalized.Description = strings.TrimSpace(payload.Description)
	normalized.Nodes = cloneSavedWorkflowDraftNodes(payload.Nodes)
	normalized.Edges = cloneSavedWorkflowDraftEdges(payload.Edges)
	normalized.InputContract = cloneSavedWorkflowDraftContract(payload.InputContract)
	normalized.OutputContract = cloneSavedWorkflowDraftContract(payload.OutputContract)
	normalized.ProviderRefs = normalizedStringSet(payload.ProviderRefs)
	normalized.ToolRefs = normalizedStringSet(payload.ToolRefs)
	normalized.RAGRefs = normalizedStringSet(payload.RAGRefs)
	normalized.RequestedCapabilities = normalizedStringSet(payload.RequestedCapabilities)
	normalized.AdditionalFields = cloneSavedWorkflowDraftAdditionalFields(payload.AdditionalFields)
	sort.Slice(normalized.Nodes, func(i, j int) bool { return normalized.Nodes[i].NodeID < normalized.Nodes[j].NodeID })
	sort.Slice(normalized.Edges, func(i, j int) bool { return normalized.Edges[i].EdgeID < normalized.Edges[j].EdgeID })
	normalized.AdditionalFields = normalizeSavedWorkflowDraftAdditionalFields(
		normalized.AdditionalFields,
		normalized.Nodes,
	)
	return normalized
}

func savedWorkflowDraftPayloadTooLarge(payload SavedWorkflowDraftPayload) bool {
	if len(payload.Nodes) > maxSavedWorkflowDraftNodes || len(payload.Edges) > maxSavedWorkflowDraftEdges {
		return true
	}
	for _, text := range []string{payload.Name, payload.Description, payload.InputContract.Summary, payload.OutputContract.Summary} {
		if len(text) > maxSavedWorkflowDraftTextLength {
			return true
		}
	}
	for _, node := range payload.Nodes {
		if len(node.Label) > maxSavedWorkflowDraftLabelLength {
			return true
		}
		for _, text := range []string{node.InputSummary, node.OutputSummary, node.OutputMappingSummary} {
			if len(text) > maxSavedWorkflowDraftTextLength {
				return true
			}
		}
	}
	return false
}

func firstSavedWorkflowDraftForbiddenField(fields map[string]any) string {
	return firstSavedWorkflowDraftForbiddenFieldAtPath("additional_fields", fields)
}

func firstSavedWorkflowDraftForbiddenFieldAtPath(path string, fields map[string]any) string {
	for _, key := range []string{
		"secret_value",
		"api_key_value",
		"token",
		"oauth_token",
		"oidc_token",
		"confirmation_decision",
		"run_input",
		"run_output",
		"materialized_result",
		"writeback_payload",
		"executor_result",
	} {
		if _, found := fields[key]; found {
			return path + "." + key
		}
	}
	for key, value := range fields {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			continue
		}
		nestedPath := path + "." + normalizedKey
		switch typed := value.(type) {
		case map[string]any:
			if forbiddenField := firstSavedWorkflowDraftForbiddenFieldAtPath(nestedPath, typed); forbiddenField != "" {
				return forbiddenField
			}
		case []any:
			for _, item := range typed {
				if itemMap, ok := item.(map[string]any); ok {
					if forbiddenField := firstSavedWorkflowDraftForbiddenFieldAtPath(nestedPath, itemMap); forbiddenField != "" {
						return forbiddenField
					}
				}
			}
		}
	}
	return ""
}

func normalizeSavedWorkflowDraftAdditionalFields(
	fields map[string]any,
	nodes []SavedWorkflowDraftNode,
) map[string]any {
	normalized := cloneSavedWorkflowDraftAdditionalFields(fields)
	if len(normalized) == 0 {
		return nil
	}
	if layout, found := normalizeSavedWorkflowDraftDesignerLayout(
		normalized[savedWorkflowDraftDesignerLayoutAdditionalField],
		nodes,
	); found {
		normalized[savedWorkflowDraftDesignerLayoutAdditionalField] = layout
	} else {
		delete(normalized, savedWorkflowDraftDesignerLayoutAdditionalField)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeSavedWorkflowDraftDesignerLayout(
	value any,
	nodes []SavedWorkflowDraftNode,
) (map[string]any, bool) {
	layout, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}
	if savedWorkflowDraftAdditionalString(layout["layout_version"]) != savedWorkflowDraftDesignerLayoutVersion ||
		savedWorkflowDraftAdditionalString(layout["source"]) != savedWorkflowDraftDesignerLayoutSource ||
		savedWorkflowDraftAdditionalString(layout["persistence"]) != savedWorkflowDraftDesignerLayoutPersistence {
		return nil, false
	}
	rawNodes, ok := layout["nodes"].([]any)
	if !ok {
		return nil, false
	}
	knownNodeIDs := make(map[string]bool, len(nodes))
	for _, node := range nodes {
		knownNodeIDs[node.NodeID] = true
	}
	positionsByNodeID := make(map[string]map[string]any, len(rawNodes))
	for _, rawNode := range rawNodes {
		nodeLayout, ok := rawNode.(map[string]any)
		if !ok {
			continue
		}
		nodeID := savedWorkflowDraftAdditionalString(nodeLayout["node_id"])
		if nodeID == "" || !knownNodeIDs[nodeID] {
			continue
		}
		x, xOK := savedWorkflowDraftAdditionalNumber(nodeLayout["x"])
		y, yOK := savedWorkflowDraftAdditionalNumber(nodeLayout["y"])
		if !xOK || !yOK {
			continue
		}
		positionsByNodeID[nodeID] = map[string]any{
			"node_id": nodeID,
			"x":       savedWorkflowDraftLayoutCoordinate(x),
			"y":       savedWorkflowDraftLayoutCoordinate(y),
			"pinned":  false,
		}
	}
	if len(positionsByNodeID) == 0 {
		return nil, false
	}
	normalizedNodes := make([]any, 0, len(positionsByNodeID))
	for _, node := range nodes {
		if position, found := positionsByNodeID[node.NodeID]; found {
			normalizedNodes = append(normalizedNodes, position)
		}
	}
	return map[string]any{
		"layout_version": savedWorkflowDraftDesignerLayoutVersion,
		"source":         savedWorkflowDraftDesignerLayoutSource,
		"persistence":    savedWorkflowDraftDesignerLayoutPersistence,
		"nodes":          normalizedNodes,
	}, true
}

func savedWorkflowDraftAdditionalString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func savedWorkflowDraftAdditionalNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case float32:
		value := float64(typed)
		return value, !math.IsNaN(value) && !math.IsInf(value, 0)
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	default:
		return 0, false
	}
}

func savedWorkflowDraftLayoutCoordinate(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Max(-maxSavedWorkflowDraftLayoutCoordinate, math.Min(maxSavedWorkflowDraftLayoutCoordinate, math.Round(value)))
}

func validateSavedWorkflowDraftGraph(
	payload SavedWorkflowDraftPayload,
) ([]SavedWorkflowDraftValidationFinding, []SavedWorkflowDraftBlockedCapability, SavedWorkflowDraftFailureCode) {
	findings := make([]SavedWorkflowDraftValidationFinding, 0)
	blockedCapabilities := make([]SavedWorkflowDraftBlockedCapability, 0)
	nodeIDs := make(map[string]bool, len(payload.Nodes))
	outputNodeFound := false
	for _, node := range payload.Nodes {
		if node.NodeID == "" || node.NodeType == "" {
			return nil, nil, SavedWorkflowDraftFailureGraphInvalid
		}
		if nodeIDs[node.NodeID] {
			return nil, nil, SavedWorkflowDraftFailureGraphInvalid
		}
		nodeIDs[node.NodeID] = true
		if node.NodeType == "output" {
			outputNodeFound = true
		}
		if savedWorkflowDraftBlockedNodeType(node.NodeType) {
			findings = append(findings, SavedWorkflowDraftValidationFinding{
				Code:       SavedWorkflowDraftFailureBlockedCapability,
				Severity:   SavedWorkflowDraftValidationBlocking,
				Field:      "nodes.node_type",
				Summary:    "Reserved workflow node type is blocked in Saved Workflow Draft v1.",
				EvidenceID: node.NodeID,
			})
			blockedCapabilities = append(blockedCapabilities, SavedWorkflowDraftBlockedCapability{
				CapabilityID:        node.NodeType,
				MissingPrerequisite: "independent runtime capability task card",
				Summary:             "Reserved capability can be saved only as a blocked review finding.",
			})
			continue
		}
		if !savedWorkflowDraftAllowedNodeType(node.NodeType) {
			return nil, nil, SavedWorkflowDraftFailureGraphInvalid
		}
	}
	for _, edge := range payload.Edges {
		if edge.EdgeID == "" || !nodeIDs[edge.FromNodeID] || !nodeIDs[edge.ToNodeID] {
			return nil, nil, SavedWorkflowDraftFailureGraphInvalid
		}
	}
	if !outputNodeFound {
		findings = append(findings, SavedWorkflowDraftValidationFinding{
			Code:     SavedWorkflowDraftFailureGraphInvalid,
			Severity: SavedWorkflowDraftValidationBlocking,
			Field:    "nodes",
			Summary:  "Saved workflow draft must include an output node before review.",
		})
	}
	return findings, blockedCapabilities, ""
}

func validateSavedWorkflowDraftContracts(payload SavedWorkflowDraftPayload) []SavedWorkflowDraftValidationFinding {
	findings := make([]SavedWorkflowDraftValidationFinding, 0)
	if payload.InputContract.ContractID == "" || len(payload.InputContract.RequiredFields) == 0 {
		findings = append(findings, SavedWorkflowDraftValidationFinding{
			Code:     SavedWorkflowDraftFailureContractInvalid,
			Severity: SavedWorkflowDraftValidationBlocking,
			Field:    "input_contract",
			Summary:  "Input contract must define a contract id and required fields.",
		})
	}
	if payload.OutputContract.ContractID == "" || len(payload.OutputContract.RequiredFields) == 0 {
		findings = append(findings, SavedWorkflowDraftValidationFinding{
			Code:     SavedWorkflowDraftFailureContractInvalid,
			Severity: SavedWorkflowDraftValidationBlocking,
			Field:    "output_contract",
			Summary:  "Output contract must define a contract id and required fields.",
		})
	}
	for _, node := range payload.Nodes {
		if node.InputContractRef == "" || node.OutputContractRef == "" ||
			len(node.InputContractFields) == 0 || len(node.OutputContractFields) == 0 {
			findings = append(findings, SavedWorkflowDraftValidationFinding{
				Code:       SavedWorkflowDraftFailureContractInvalid,
				Severity:   SavedWorkflowDraftValidationBlocking,
				Field:      "nodes.contract_fields",
				Summary:    "Saved workflow draft nodes must define contract refs and explicit contract fields.",
				EvidenceID: node.NodeID,
			})
		}
	}
	return findings
}

func validateSavedWorkflowDraftCapabilities(
	payload SavedWorkflowDraftPayload,
) ([]SavedWorkflowDraftValidationFinding, []SavedWorkflowDraftBlockedCapability) {
	findings := make([]SavedWorkflowDraftValidationFinding, 0)
	blockedCapabilities := make([]SavedWorkflowDraftBlockedCapability, 0)
	for _, capability := range payload.RequestedCapabilities {
		if !savedWorkflowDraftBlockedCapability(capability) {
			continue
		}
		findings = append(findings, SavedWorkflowDraftValidationFinding{
			Code:       SavedWorkflowDraftFailureBlockedCapability,
			Severity:   SavedWorkflowDraftValidationBlocking,
			Field:      "requested_capabilities",
			Summary:    "Requested capability is blocked in Saved Workflow Draft v1.",
			EvidenceID: capability,
		})
		blockedCapabilities = append(blockedCapabilities, SavedWorkflowDraftBlockedCapability{
			CapabilityID:        capability,
			MissingPrerequisite: "independent workflow runtime implementation target",
			Summary:             "Saved draft can preserve this as a blocked review finding but cannot unlock it.",
		})
	}
	return findings, blockedCapabilities
}

func validateSavedWorkflowDraftRisk(payload SavedWorkflowDraftPayload) []SavedWorkflowDraftValidationFinding {
	findings := make([]SavedWorkflowDraftValidationFinding, 0)
	for _, node := range payload.Nodes {
		if node.NodeType != "http_tool" && node.RiskLevel != "high" {
			continue
		}
		if node.RequiresConfirmation {
			findings = append(findings, SavedWorkflowDraftValidationFinding{
				Code:       SavedWorkflowDraftFailureContractInvalid,
				Severity:   SavedWorkflowDraftValidationInfo,
				Field:      "nodes.requires_confirmation",
				Summary:    "Risk-bearing node is marked for future human confirmation; v1 does not store a decision.",
				EvidenceID: node.NodeID,
			})
			continue
		}
		findings = append(findings, SavedWorkflowDraftValidationFinding{
			Code:       SavedWorkflowDraftFailureContractInvalid,
			Severity:   SavedWorkflowDraftValidationBlocking,
			Field:      "nodes.requires_confirmation",
			Summary:    "Risk-bearing node should require confirmation before any future publish or run target.",
			EvidenceID: node.NodeID,
		})
	}
	return findings
}

func deriveSavedWorkflowDraftValidationState(
	findings []SavedWorkflowDraftValidationFinding,
	blockedCapabilities []SavedWorkflowDraftBlockedCapability,
) SavedWorkflowDraftStatus {
	if len(blockedCapabilities) > 0 {
		return SavedWorkflowDraftStatusBlockedCapability
	}
	for _, finding := range findings {
		if finding.Severity == SavedWorkflowDraftValidationBlocking {
			return SavedWorkflowDraftStatusInvalidDraft
		}
	}
	if len(findings) > 0 {
		return SavedWorkflowDraftStatusInvalidDraft
	}
	return SavedWorkflowDraftStatusValidForReview
}

func savedWorkflowDraftMatchesScope(draft SavedWorkflowDraft, workspaceID string, applicationID string) bool {
	return draft.WorkspaceID == strings.TrimSpace(workspaceID) && draft.ApplicationID == strings.TrimSpace(applicationID)
}

func savedWorkflowDraftAllowedNodeType(nodeType string) bool {
	switch nodeType {
	case "prompt", "llm", "http_tool", "rag_retrieval", "condition", "output":
		return true
	default:
		return false
	}
}

func savedWorkflowDraftBlockedNodeType(nodeType string) bool {
	switch nodeType {
	case "code", "sandbox", "agent_loop":
		return true
	default:
		return false
	}
}

func savedWorkflowDraftBlockedCapability(capability string) bool {
	switch capability {
	case "executor", "node_executor", "tool_executor", "agent_loop", "publish", "run",
		"confirmation_decision", "decision_store", "execution_unlock", "writeback",
		"business_writeback", "replay", "resume", "materialized_result_reader",
		"real_database", "oidc", "repository_adapter", "public_production_api":
		return true
	default:
		return false
	}
}

func savedWorkflowDraftPayloadFromDraft(draft SavedWorkflowDraft) SavedWorkflowDraftPayload {
	return SavedWorkflowDraftPayload{
		DraftID:               draft.DraftID,
		WorkspaceID:           draft.WorkspaceID,
		ApplicationID:         draft.ApplicationID,
		SourceDefinitionID:    draft.SourceDefinitionID,
		BaseDefinitionVersion: draft.BaseDefinitionVersion,
		SchemaVersion:         draft.SchemaVersion,
		DraftStatus:           draft.DraftStatus,
		Name:                  draft.Name,
		Description:           draft.Description,
		Nodes:                 cloneSavedWorkflowDraftNodes(draft.Nodes),
		Edges:                 cloneSavedWorkflowDraftEdges(draft.Edges),
		InputContract:         cloneSavedWorkflowDraftContract(draft.InputContract),
		OutputContract:        cloneSavedWorkflowDraftContract(draft.OutputContract),
		ProviderRefs:          cloneStringSlice(draft.ProviderRefs),
		ToolRefs:              cloneStringSlice(draft.ToolRefs),
		RAGRefs:               cloneStringSlice(draft.RAGRefs),
		RequestedCapabilities: cloneStringSlice(draft.RequestedCapabilities),
		AdditionalFields:      cloneSavedWorkflowDraftAdditionalFields(draft.AdditionalFields),
	}
}

func savedWorkflowDraftSummaryFromDraft(draft SavedWorkflowDraft) SavedWorkflowDraftSummary {
	return SavedWorkflowDraftSummary{
		DraftID:                    draft.DraftID,
		WorkspaceID:                draft.WorkspaceID,
		ApplicationID:              draft.ApplicationID,
		SourceDefinitionID:         draft.SourceDefinitionID,
		DraftVersion:               draft.DraftVersion,
		SchemaVersion:              draft.SchemaVersion,
		DraftStatus:                draft.DraftStatus,
		Name:                       draft.Name,
		Description:                draft.Description,
		UpdatedAt:                  draft.UpdatedAt,
		UpdatedByActorRef:          draft.UpdatedByActorRef,
		NodeCount:                  len(draft.Nodes),
		EdgeCount:                  len(draft.Edges),
		BlockedCapabilityCount:     len(draft.BlockedCapabilitySummary),
		ValidationState:            draft.ValidationSummary.ValidationState,
		ValidForReview:             draft.ValidationSummary.ValidForReview,
		SampleOrUnsavedDraftStatus: draft.SampleOrUnsavedDraftStatus,
	}
}

func normalizedStringSet(values []string) []string {
	seen := make(map[string]bool, len(values))
	output := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		output = append(output, normalized)
	}
	sort.Strings(output)
	return output
}

func cloneSavedWorkflowDraftPointer(draft SavedWorkflowDraft) *SavedWorkflowDraft {
	clone := cloneSavedWorkflowDraft(draft)
	return &clone
}

func cloneSavedWorkflowDraft(draft SavedWorkflowDraft) SavedWorkflowDraft {
	clone := draft
	clone.Nodes = cloneSavedWorkflowDraftNodes(draft.Nodes)
	clone.Edges = cloneSavedWorkflowDraftEdges(draft.Edges)
	clone.InputContract = cloneSavedWorkflowDraftContract(draft.InputContract)
	clone.OutputContract = cloneSavedWorkflowDraftContract(draft.OutputContract)
	clone.ProviderRefs = cloneStringSlice(draft.ProviderRefs)
	clone.ToolRefs = cloneStringSlice(draft.ToolRefs)
	clone.RAGRefs = cloneStringSlice(draft.RAGRefs)
	clone.RequestedCapabilities = cloneStringSlice(draft.RequestedCapabilities)
	clone.AdditionalFields = cloneSavedWorkflowDraftAdditionalFields(draft.AdditionalFields)
	clone.ValidationSummary = cloneSavedWorkflowDraftValidationSummary(draft.ValidationSummary)
	clone.BlockedCapabilitySummary = cloneSavedWorkflowDraftBlockedCapabilities(draft.BlockedCapabilitySummary)
	return clone
}

func cloneSavedWorkflowDraftNodes(nodes []SavedWorkflowDraftNode) []SavedWorkflowDraftNode {
	output := make([]SavedWorkflowDraftNode, 0, len(nodes))
	for _, node := range nodes {
		node.NodeID = strings.TrimSpace(node.NodeID)
		node.NodeType = strings.TrimSpace(node.NodeType)
		node.Label = strings.TrimSpace(node.Label)
		node.InputSummary = strings.TrimSpace(node.InputSummary)
		node.OutputSummary = strings.TrimSpace(node.OutputSummary)
		node.InputContractRef = strings.TrimSpace(node.InputContractRef)
		node.OutputContractRef = strings.TrimSpace(node.OutputContractRef)
		node.InputContractFields = normalizedStringSet(node.InputContractFields)
		node.OutputContractFields = normalizedStringSet(node.OutputContractFields)
		node.OutputMappingSummary = strings.TrimSpace(node.OutputMappingSummary)
		node.ProviderRef = strings.TrimSpace(node.ProviderRef)
		node.ToolRef = strings.TrimSpace(node.ToolRef)
		node.RAGRef = strings.TrimSpace(node.RAGRef)
		node.RiskLevel = strings.TrimSpace(node.RiskLevel)
		output = append(output, node)
	}
	return output
}

func cloneSavedWorkflowDraftEdges(edges []SavedWorkflowDraftEdge) []SavedWorkflowDraftEdge {
	output := make([]SavedWorkflowDraftEdge, 0, len(edges))
	for _, edge := range edges {
		edge.EdgeID = strings.TrimSpace(edge.EdgeID)
		edge.FromNodeID = strings.TrimSpace(edge.FromNodeID)
		edge.ToNodeID = strings.TrimSpace(edge.ToNodeID)
		edge.ConditionSummary = strings.TrimSpace(edge.ConditionSummary)
		output = append(output, edge)
	}
	return output
}

func cloneSavedWorkflowDraftContract(contract SavedWorkflowDraftContract) SavedWorkflowDraftContract {
	return SavedWorkflowDraftContract{
		ContractID:     strings.TrimSpace(contract.ContractID),
		RequiredFields: normalizedStringSet(contract.RequiredFields),
		Summary:        strings.TrimSpace(contract.Summary),
	}
}

func cloneSavedWorkflowDraftValidationSummary(summary SavedWorkflowDraftValidationSummary) SavedWorkflowDraftValidationSummary {
	return SavedWorkflowDraftValidationSummary{
		ValidationState: summary.ValidationState,
		ValidForReview:  summary.ValidForReview,
		Findings:        cloneSavedWorkflowDraftFindings(summary.Findings),
	}
}

func cloneSavedWorkflowDraftFindings(findings []SavedWorkflowDraftValidationFinding) []SavedWorkflowDraftValidationFinding {
	output := make([]SavedWorkflowDraftValidationFinding, len(findings))
	copy(output, findings)
	return output
}

func cloneSavedWorkflowDraftBlockedCapabilities(
	capabilities []SavedWorkflowDraftBlockedCapability,
) []SavedWorkflowDraftBlockedCapability {
	output := make([]SavedWorkflowDraftBlockedCapability, len(capabilities))
	copy(output, capabilities)
	return output
}

func cloneStringSlice(values []string) []string {
	output := make([]string, len(values))
	copy(output, values)
	return output
}

func cloneSavedWorkflowDraftAdditionalFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return nil
	}
	output := make(map[string]any, len(fields))
	for key, value := range fields {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" {
			continue
		}
		output[normalizedKey] = cloneSavedWorkflowDraftAdditionalValue(value)
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func cloneSavedWorkflowDraftAdditionalValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneSavedWorkflowDraftAdditionalFields(typed)
	case []any:
		output := make([]any, len(typed))
		for index, item := range typed {
			output[index] = cloneSavedWorkflowDraftAdditionalValue(item)
		}
		return output
	case []map[string]any:
		output := make([]any, len(typed))
		for index, item := range typed {
			output[index] = cloneSavedWorkflowDraftAdditionalFields(item)
		}
		return output
	case []string:
		return cloneStringSlice(typed)
	default:
		return typed
	}
}
