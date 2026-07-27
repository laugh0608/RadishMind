package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	savedWorkflowDraftRevisionSchemaVersion = "saved_workflow_draft_revision.v1"
	defaultSavedWorkflowDraftRevisionLimit  = 20
	maxSavedWorkflowDraftRevisionLimit      = 50
)

const (
	SavedWorkflowDraftFailureRevisionNotFound SavedWorkflowDraftFailureCode = "draft_revision_not_found"
	SavedWorkflowDraftFailureRevisionCursor   SavedWorkflowDraftFailureCode = "draft_revision_cursor_invalid"
	SavedWorkflowDraftFailureRevisionRestore  SavedWorkflowDraftFailureCode = "draft_revision_restore_invalid"
)

type SavedWorkflowDraftRevisionKind string

const (
	SavedWorkflowDraftRevisionKindSaved             SavedWorkflowDraftRevisionKind = "saved"
	SavedWorkflowDraftRevisionKindRestored          SavedWorkflowDraftRevisionKind = "restored"
	SavedWorkflowDraftRevisionKindBackfilledCurrent SavedWorkflowDraftRevisionKind = "backfilled_current"
)

type SavedWorkflowDraftRevision struct {
	SchemaVersion       string
	Draft               SavedWorkflowDraft
	RevisionKind        SavedWorkflowDraftRevisionKind
	RestoredFromVersion int
}

type SavedWorkflowDraftRevisionSummary struct {
	SchemaVersion       string
	DraftID             string
	DraftVersion        int
	RevisionKind        SavedWorkflowDraftRevisionKind
	RestoredFromVersion int
	DraftStatus         SavedWorkflowDraftStatus
	Name                string
	UpdatedAt           string
	UpdatedByActorRef   string
	NodeCount           int
	EdgeCount           int
	BlockedCount        int
}

type ListSavedWorkflowDraftRevisionsRequest struct {
	DraftID string
	Limit   int
	Cursor  string
}

type ReadSavedWorkflowDraftRevisionRequest struct {
	DraftID      string
	DraftVersion int
}

type RestoreSavedWorkflowDraftRevisionRequest struct {
	DraftID                     string
	SourceDraftVersion          int
	ExpectedCurrentDraftVersion int
}

type SavedWorkflowDraftRevisionResult struct {
	Revision             *SavedWorkflowDraftRevision
	Draft                *SavedWorkflowDraft
	FailureCode          SavedWorkflowDraftFailureCode
	CurrentDraftVersion  int
	ValidationSummary    SavedWorkflowDraftValidationSummary
	RequestAuditMetadata SavedWorkflowDraftAuditMetadata
}

type SavedWorkflowDraftRevisionListResult struct {
	Revisions            []SavedWorkflowDraftRevisionSummary
	NextCursor           string
	HasMore              bool
	FailureCode          SavedWorkflowDraftFailureCode
	RequestAuditMetadata SavedWorkflowDraftAuditMetadata
}

type savedWorkflowDraftRevisionListFilter struct {
	BeforeVersion int
	Limit         int
}

type savedWorkflowDraftRevisionPage struct {
	Revisions []SavedWorkflowDraftRevisionSummary
	HasMore   bool
}

type savedWorkflowDraftRevisionStore interface {
	ReadDraftRevision(
		SavedWorkflowDraftContext,
		string,
		int,
	) (SavedWorkflowDraftRevision, bool, error)
	ListDraftRevisions(
		SavedWorkflowDraftContext,
		string,
		savedWorkflowDraftRevisionListFilter,
	) (savedWorkflowDraftRevisionPage, error)
	WriteRestoredDraft(
		SavedWorkflowDraftContext,
		SavedWorkflowDraft,
		int,
		int,
	) (int, error)
}

func (service savedWorkflowDraftService) ReadDraftRevision(
	context SavedWorkflowDraftContext,
	request ReadSavedWorkflowDraftRevisionRequest,
) SavedWorkflowDraftRevisionResult {
	audit := savedWorkflowDraftAuditMetadata(context)
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" || request.DraftVersion < 1 {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailurePayloadInvalid, audit)
	}
	store, ok := service.store.(savedWorkflowDraftRevisionStore)
	if !ok {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureStoreUnavailable, audit)
	}
	revision, found, err := store.ReadDraftRevision(context, draftID, request.DraftVersion)
	if err != nil {
		return savedWorkflowDraftRevisionFailure(savedWorkflowDraftStoreFailureCode(err), audit)
	}
	if !found {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureRevisionNotFound, audit)
	}
	if failure := validateSavedWorkflowDraftRevisionScope(context, revision, draftID); failure != "" {
		return savedWorkflowDraftRevisionFailure(failure, audit)
	}
	return SavedWorkflowDraftRevisionResult{
		Revision:             cloneSavedWorkflowDraftRevisionPointer(revision),
		CurrentDraftVersion:  revision.Draft.DraftVersion,
		ValidationSummary:    cloneSavedWorkflowDraftValidationSummary(revision.Draft.ValidationSummary),
		RequestAuditMetadata: audit,
	}
}

func (service savedWorkflowDraftService) ListDraftRevisions(
	context SavedWorkflowDraftContext,
	request ListSavedWorkflowDraftRevisionsRequest,
) SavedWorkflowDraftRevisionListResult {
	audit := savedWorkflowDraftAuditMetadata(context)
	draftID := strings.TrimSpace(request.DraftID)
	limit := request.Limit
	if limit == 0 {
		limit = defaultSavedWorkflowDraftRevisionLimit
	}
	if draftID == "" || limit < 1 || limit > maxSavedWorkflowDraftRevisionLimit {
		return savedWorkflowDraftRevisionListFailure(SavedWorkflowDraftFailurePayloadInvalid, audit)
	}
	beforeVersion := 0
	if strings.TrimSpace(request.Cursor) != "" {
		var err error
		beforeVersion, err = decodeSavedWorkflowDraftRevisionCursor(request.Cursor, draftID, limit)
		if err != nil {
			return savedWorkflowDraftRevisionListFailure(SavedWorkflowDraftFailureRevisionCursor, audit)
		}
	}
	store, ok := service.store.(savedWorkflowDraftRevisionStore)
	if !ok {
		return savedWorkflowDraftRevisionListFailure(SavedWorkflowDraftFailureStoreUnavailable, audit)
	}
	page, err := store.ListDraftRevisions(
		context,
		draftID,
		savedWorkflowDraftRevisionListFilter{BeforeVersion: beforeVersion, Limit: limit},
	)
	if err != nil {
		return savedWorkflowDraftRevisionListFailure(savedWorkflowDraftStoreFailureCode(err), audit)
	}
	for _, revision := range page.Revisions {
		if revision.DraftID != draftID || revision.DraftVersion < 1 {
			return savedWorkflowDraftRevisionListFailure(SavedWorkflowDraftFailureStoreContractMismatch, audit)
		}
	}
	next := ""
	if page.HasMore && len(page.Revisions) > 0 {
		next, err = encodeSavedWorkflowDraftRevisionCursor(
			draftID,
			page.Revisions[len(page.Revisions)-1].DraftVersion,
			limit,
		)
		if err != nil {
			return savedWorkflowDraftRevisionListFailure(SavedWorkflowDraftFailureStoreContractMismatch, audit)
		}
	}
	return SavedWorkflowDraftRevisionListResult{
		Revisions:            append([]SavedWorkflowDraftRevisionSummary{}, page.Revisions...),
		NextCursor:           next,
		HasMore:              page.HasMore,
		RequestAuditMetadata: audit,
	}
}

func (service savedWorkflowDraftService) RestoreDraftRevision(
	context SavedWorkflowDraftContext,
	request RestoreSavedWorkflowDraftRevisionRequest,
) SavedWorkflowDraftRevisionResult {
	audit := savedWorkflowDraftAuditMetadata(context)
	if !context.WriteEnabled {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureWriteDisabled, audit)
	}
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" || request.SourceDraftVersion < 1 || request.ExpectedCurrentDraftVersion < 1 {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureRevisionRestore, audit)
	}
	store, ok := service.store.(savedWorkflowDraftRevisionStore)
	if !ok {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureStoreUnavailable, audit)
	}
	source, found, err := store.ReadDraftRevision(context, draftID, request.SourceDraftVersion)
	if err != nil {
		return savedWorkflowDraftRevisionFailure(savedWorkflowDraftStoreFailureCode(err), audit)
	}
	if !found {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureRevisionNotFound, audit)
	}
	if failure := validateSavedWorkflowDraftRevisionScope(context, source, draftID); failure != "" {
		return savedWorkflowDraftRevisionFailure(failure, audit)
	}
	current, found, err := service.store.ReadDraftByID(context, draftID)
	if err != nil {
		return savedWorkflowDraftRevisionFailure(savedWorkflowDraftStoreFailureCode(err), audit)
	}
	if !found {
		return savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureNotFound, audit)
	}
	if current.DraftVersion != request.ExpectedCurrentDraftVersion {
		result := savedWorkflowDraftRevisionFailure(SavedWorkflowDraftFailureVersionConflict, audit)
		result.CurrentDraftVersion = current.DraftVersion
		return result
	}
	payload := savedWorkflowDraftPayloadFromDraft(source.Draft)
	normalized, validation := service.validatePayload(context, payload)
	if validation.FailureCode != "" {
		return savedWorkflowDraftRevisionFailure(validation.FailureCode, audit)
	}
	now := service.now().UTC().Format(time.RFC3339)
	restored := SavedWorkflowDraft{
		DraftID:                    normalized.DraftID,
		WorkspaceID:                normalized.WorkspaceID,
		ApplicationID:              normalized.ApplicationID,
		SourceDefinitionID:         normalized.SourceDefinitionID,
		BaseDefinitionVersion:      normalized.BaseDefinitionVersion,
		DraftVersion:               current.DraftVersion + 1,
		SchemaVersion:              normalized.SchemaVersion,
		DraftStatus:                validation.ValidationSummary.ValidationState,
		CreatedAt:                  current.CreatedAt,
		UpdatedAt:                  now,
		CreatedByActorRef:          current.CreatedByActorRef,
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
		ValidationSummary:          cloneSavedWorkflowDraftValidationSummary(validation.ValidationSummary),
		BlockedCapabilitySummary:   cloneSavedWorkflowDraftBlockedCapabilities(validation.BlockedCapabilities),
		RequestAuditMetadata:       audit,
		SampleOrUnsavedDraftStatus: "saved_draft_record",
	}
	currentVersion, err := store.WriteRestoredDraft(
		context,
		restored,
		request.ExpectedCurrentDraftVersion,
		request.SourceDraftVersion,
	)
	if err != nil {
		result := savedWorkflowDraftRevisionFailure(savedWorkflowDraftStoreFailureCode(err), audit)
		result.CurrentDraftVersion = currentVersion
		return result
	}
	return SavedWorkflowDraftRevisionResult{
		Draft:                cloneSavedWorkflowDraftPointer(restored),
		CurrentDraftVersion:  restored.DraftVersion,
		ValidationSummary:    cloneSavedWorkflowDraftValidationSummary(restored.ValidationSummary),
		RequestAuditMetadata: audit,
	}
}

func (store *memorySavedWorkflowDraftStore) ReadDraftRevision(
	context SavedWorkflowDraftContext,
	draftID string,
	version int,
) (SavedWorkflowDraftRevision, bool, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return SavedWorkflowDraftRevision{}, false, errors.New("saved workflow draft store unavailable")
	}
	revision, found := store.revisions[draftID][version]
	if !found {
		return SavedWorkflowDraftRevision{}, false, nil
	}
	if !savedWorkflowDraftMatchesScope(revision.Draft, context.WorkspaceID, context.ApplicationID) {
		return SavedWorkflowDraftRevision{}, false, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureScopeDenied,
		)
	}
	return cloneSavedWorkflowDraftRevision(revision), true, nil
}

func (store *memorySavedWorkflowDraftStore) ListDraftRevisions(
	context SavedWorkflowDraftContext,
	draftID string,
	filter savedWorkflowDraftRevisionListFilter,
) (savedWorkflowDraftRevisionPage, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if store.unavailable {
		return savedWorkflowDraftRevisionPage{}, errors.New("saved workflow draft store unavailable")
	}
	current, found := store.drafts[draftID]
	if !found {
		return savedWorkflowDraftRevisionPage{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureNotFound,
		)
	}
	if !savedWorkflowDraftMatchesScope(current, context.WorkspaceID, context.ApplicationID) {
		return savedWorkflowDraftRevisionPage{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureScopeDenied,
		)
	}
	values := make([]SavedWorkflowDraftRevisionSummary, 0, len(store.revisions[draftID]))
	for _, revision := range store.revisions[draftID] {
		if filter.BeforeVersion > 0 && revision.Draft.DraftVersion >= filter.BeforeVersion {
			continue
		}
		values = append(values, savedWorkflowDraftRevisionSummary(revision))
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].DraftVersion > values[j].DraftVersion
	})
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return savedWorkflowDraftRevisionPage{Revisions: values, HasMore: hasMore}, nil
}

func (store *memorySavedWorkflowDraftStore) WriteRestoredDraft(
	_ SavedWorkflowDraftContext,
	draft SavedWorkflowDraft,
	expectedDraftVersion int,
	restoredFromVersion int,
) (int, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, found := store.revisions[draft.DraftID][restoredFromVersion]; !found {
		return expectedDraftVersion, savedWorkflowDraftStoreWriteFailure(
			SavedWorkflowDraftFailureRevisionNotFound,
		)
	}
	return store.writeDraftLocked(
		draft,
		expectedDraftVersion,
		SavedWorkflowDraftRevisionKindRestored,
		restoredFromVersion,
	)
}

func newSavedWorkflowDraftRevision(
	draft SavedWorkflowDraft,
	kind SavedWorkflowDraftRevisionKind,
	restoredFromVersion int,
) SavedWorkflowDraftRevision {
	if kind == "" {
		kind = SavedWorkflowDraftRevisionKindSaved
	}
	return SavedWorkflowDraftRevision{
		SchemaVersion:       savedWorkflowDraftRevisionSchemaVersion,
		Draft:               cloneSavedWorkflowDraft(draft),
		RevisionKind:        kind,
		RestoredFromVersion: restoredFromVersion,
	}
}

func validSavedWorkflowDraftRevisionWriteMetadata(
	kind SavedWorkflowDraftRevisionKind,
	restoredFromVersion int,
) bool {
	return (kind == SavedWorkflowDraftRevisionKindSaved && restoredFromVersion == 0) ||
		(kind == SavedWorkflowDraftRevisionKindRestored && restoredFromVersion > 0)
}

func savedWorkflowDraftRevisionSummary(
	revision SavedWorkflowDraftRevision,
) SavedWorkflowDraftRevisionSummary {
	return SavedWorkflowDraftRevisionSummary{
		SchemaVersion:       revision.SchemaVersion,
		DraftID:             revision.Draft.DraftID,
		DraftVersion:        revision.Draft.DraftVersion,
		RevisionKind:        revision.RevisionKind,
		RestoredFromVersion: revision.RestoredFromVersion,
		DraftStatus:         revision.Draft.DraftStatus,
		Name:                revision.Draft.Name,
		UpdatedAt:           revision.Draft.UpdatedAt,
		UpdatedByActorRef:   revision.Draft.UpdatedByActorRef,
		NodeCount:           len(revision.Draft.Nodes),
		EdgeCount:           len(revision.Draft.Edges),
		BlockedCount:        len(revision.Draft.BlockedCapabilitySummary),
	}
}

func validateSavedWorkflowDraftRevisionScope(
	context SavedWorkflowDraftContext,
	revision SavedWorkflowDraftRevision,
	draftID string,
) SavedWorkflowDraftFailureCode {
	if revision.SchemaVersion != savedWorkflowDraftRevisionSchemaVersion ||
		revision.Draft.DraftID != draftID ||
		revision.Draft.DraftVersion < 1 {
		return SavedWorkflowDraftFailureStoreContractMismatch
	}
	if !savedWorkflowDraftMatchesScope(revision.Draft, context.WorkspaceID, context.ApplicationID) {
		return SavedWorkflowDraftFailureScopeDenied
	}
	if revision.RevisionKind == SavedWorkflowDraftRevisionKindRestored {
		if revision.RestoredFromVersion < 1 {
			return SavedWorkflowDraftFailureStoreContractMismatch
		}
	} else if revision.RestoredFromVersion != 0 {
		return SavedWorkflowDraftFailureStoreContractMismatch
	}
	return ""
}

func cloneSavedWorkflowDraftRevision(
	revision SavedWorkflowDraftRevision,
) SavedWorkflowDraftRevision {
	revision.Draft = cloneSavedWorkflowDraft(revision.Draft)
	return revision
}

func cloneSavedWorkflowDraftRevisionPointer(
	revision SavedWorkflowDraftRevision,
) *SavedWorkflowDraftRevision {
	cloned := cloneSavedWorkflowDraftRevision(revision)
	return &cloned
}

func savedWorkflowDraftRevisionFailure(
	code SavedWorkflowDraftFailureCode,
	audit SavedWorkflowDraftAuditMetadata,
) SavedWorkflowDraftRevisionResult {
	return SavedWorkflowDraftRevisionResult{FailureCode: code, RequestAuditMetadata: audit}
}

func savedWorkflowDraftRevisionListFailure(
	code SavedWorkflowDraftFailureCode,
	audit SavedWorkflowDraftAuditMetadata,
) SavedWorkflowDraftRevisionListResult {
	return SavedWorkflowDraftRevisionListResult{FailureCode: code, RequestAuditMetadata: audit}
}

type savedWorkflowDraftRevisionCursor struct {
	Version       int    `json:"version"`
	BeforeVersion int    `json:"before_version"`
	Digest        string `json:"digest"`
}

func encodeSavedWorkflowDraftRevisionCursor(draftID string, beforeVersion int, limit int) (string, error) {
	document := savedWorkflowDraftRevisionCursor{
		Version:       1,
		BeforeVersion: beforeVersion,
		Digest:        savedWorkflowDraftRevisionCursorDigest(draftID, limit),
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeSavedWorkflowDraftRevisionCursor(value string, draftID string, limit int) (int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return 0, errors.New("invalid saved workflow draft revision cursor")
	}
	var document savedWorkflowDraftRevisionCursor
	if json.Unmarshal(raw, &document) != nil ||
		document.Version != 1 ||
		document.BeforeVersion < 1 ||
		document.Digest != savedWorkflowDraftRevisionCursorDigest(draftID, limit) {
		return 0, errors.New("invalid saved workflow draft revision cursor")
	}
	return document.BeforeVersion, nil
}

func savedWorkflowDraftRevisionCursorDigest(draftID string, limit int) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(
		fmt.Sprintf("saved-workflow-draft-revisions:%s:%d", strings.TrimSpace(draftID), limit),
	)))
}

var _ savedWorkflowDraftRevisionStore = (*memorySavedWorkflowDraftStore)(nil)
