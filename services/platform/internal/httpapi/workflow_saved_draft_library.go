package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultSavedWorkflowDraftListLimit = 25
	maxSavedWorkflowDraftListLimit     = 100
	maxSavedWorkflowDraftNamePrefix    = 80

	savedWorkflowDraftListCursorSchemaVersion     = "saved_workflow_draft_list_cursor.v1"
	savedWorkflowDraftLifecycleEventSchemaVersion = "saved_workflow_draft_lifecycle_event.v1"
)

type SavedWorkflowDraftLifecycleState string

const (
	SavedWorkflowDraftLifecycleActive   SavedWorkflowDraftLifecycleState = "active"
	SavedWorkflowDraftLifecycleArchived SavedWorkflowDraftLifecycleState = "archived"
)

type SavedWorkflowDraftProvenanceKind string

const (
	SavedWorkflowDraftProvenanceUnversioned  SavedWorkflowDraftProvenanceKind = "unversioned"
	SavedWorkflowDraftProvenanceDefinition   SavedWorkflowDraftProvenanceKind = "workflow_definition"
	SavedWorkflowDraftProvenanceDraftDerived SavedWorkflowDraftProvenanceKind = "saved_draft_derivation"
	SavedWorkflowDraftProvenanceTemplate     SavedWorkflowDraftProvenanceKind = "workspace_template_derivation"
)

type SavedWorkflowDraftLifecycleTransitionKind string

const (
	SavedWorkflowDraftLifecycleTransitionArchived   SavedWorkflowDraftLifecycleTransitionKind = "archived"
	SavedWorkflowDraftLifecycleTransitionUnarchived SavedWorkflowDraftLifecycleTransitionKind = "unarchived"
)

type TransitionSavedWorkflowDraftLifecycleRequest struct {
	DraftID                  string
	ExpectedDraftVersion     int
	ExpectedLifecycleVersion int
}

type SavedWorkflowDraftLifecycleEvent struct {
	SchemaVersion    string
	TenantRef        string
	WorkspaceID      string
	ApplicationID    string
	OwnerSubjectRef  string
	DraftID          string
	LifecycleVersion int
	FromState        SavedWorkflowDraftLifecycleState
	ToState          SavedWorkflowDraftLifecycleState
	TransitionKind   SavedWorkflowDraftLifecycleTransitionKind
	OccurredAt       string
	ActorRef         string
	RequestID        string
	AuditRef         string
}

type SavedWorkflowDraftLifecycleTransitionResult struct {
	Draft                   *SavedWorkflowDraft
	Event                   *SavedWorkflowDraftLifecycleEvent
	FailureCode             SavedWorkflowDraftFailureCode
	CurrentDraftVersion     int
	CurrentLifecycleVersion int
	CurrentLifecycleState   SavedWorkflowDraftLifecycleState
	RequestAuditMetadata    SavedWorkflowDraftAuditMetadata
}

type savedWorkflowDraftLibraryListFilter struct {
	LifecycleState  SavedWorkflowDraftLifecycleState
	Limit           int
	NamePrefix      string
	ValidationState SavedWorkflowDraftStatus
	ProvenanceKind  SavedWorkflowDraftProvenanceKind
	AfterUpdatedAt  string
	AfterDraftID    string
}

type savedWorkflowDraftLibraryPage struct {
	Summaries []SavedWorkflowDraftSummary
	HasMore   bool
}

type savedWorkflowDraftLibraryStore interface {
	ListDraftLibraryPage(
		SavedWorkflowDraftContext,
		savedWorkflowDraftLibraryListFilter,
	) (savedWorkflowDraftLibraryPage, error)
	TransitionDraftLifecycle(
		SavedWorkflowDraftContext,
		string,
		SavedWorkflowDraftLifecycleState,
		int,
		int,
		time.Time,
	) (SavedWorkflowDraft, SavedWorkflowDraftLifecycleEvent, error)
}

type savedWorkflowDraftListCursor struct {
	SchemaVersion string `json:"schema_version"`
	LastUpdatedAt string `json:"last_library_updated_at"`
	LastDraftID   string `json:"last_draft_id"`
	BindingDigest string `json:"binding_digest"`
}

func (service savedWorkflowDraftService) listDraftLibrary(
	requestContext SavedWorkflowDraftContext,
	request ListWorkflowDraftsRequest,
	store savedWorkflowDraftLibraryStore,
) SavedWorkflowDraftListResult {
	audit := savedWorkflowDraftAuditMetadata(requestContext)
	filter, failureCode := normalizeSavedWorkflowDraftLibraryListRequest(requestContext, request)
	if failureCode != "" {
		return savedWorkflowDraftListFailure(failureCode, audit)
	}
	page, err := store.ListDraftLibraryPage(requestContext, filter)
	if err != nil {
		return savedWorkflowDraftListFailure(savedWorkflowDraftStoreFailureCode(err), audit)
	}
	for _, summary := range page.Summaries {
		if summary.WorkspaceID != requestContext.WorkspaceID ||
			summary.ApplicationID != requestContext.ApplicationID ||
			summary.LifecycleState != filter.LifecycleState ||
			!validSavedWorkflowDraftSummaryLifecycle(summary) {
			return savedWorkflowDraftListFailure(SavedWorkflowDraftFailureLifecycleStoreContract, audit)
		}
	}
	nextCursor := ""
	if page.HasMore {
		if len(page.Summaries) == 0 {
			return savedWorkflowDraftListFailure(SavedWorkflowDraftFailureLifecycleStoreContract, audit)
		}
		last := page.Summaries[len(page.Summaries)-1]
		nextCursor, err = encodeSavedWorkflowDraftListCursor(requestContext, request, filter, last)
		if err != nil {
			return savedWorkflowDraftListFailure(SavedWorkflowDraftFailureLifecycleStoreContract, audit)
		}
	}
	return SavedWorkflowDraftListResult{
		Summaries:            append([]SavedWorkflowDraftSummary{}, page.Summaries...),
		NextCursor:           nextCursor,
		HasMore:              page.HasMore,
		RequestAuditMetadata: audit,
	}
}

func (service savedWorkflowDraftService) ArchiveDraft(
	requestContext SavedWorkflowDraftContext,
	request TransitionSavedWorkflowDraftLifecycleRequest,
) SavedWorkflowDraftLifecycleTransitionResult {
	return service.transitionDraftLifecycle(requestContext, request, SavedWorkflowDraftLifecycleArchived)
}

func (service savedWorkflowDraftService) UnarchiveDraft(
	requestContext SavedWorkflowDraftContext,
	request TransitionSavedWorkflowDraftLifecycleRequest,
) SavedWorkflowDraftLifecycleTransitionResult {
	return service.transitionDraftLifecycle(requestContext, request, SavedWorkflowDraftLifecycleActive)
}

func (service savedWorkflowDraftService) transitionDraftLifecycle(
	requestContext SavedWorkflowDraftContext,
	request TransitionSavedWorkflowDraftLifecycleRequest,
	target SavedWorkflowDraftLifecycleState,
) SavedWorkflowDraftLifecycleTransitionResult {
	audit := savedWorkflowDraftAuditMetadata(requestContext)
	if !requestContext.WriteEnabled {
		return savedWorkflowDraftLifecycleFailure(SavedWorkflowDraftFailureWriteDisabled, audit)
	}
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" || request.ExpectedDraftVersion < 1 || request.ExpectedLifecycleVersion < 1 {
		return savedWorkflowDraftLifecycleFailure(SavedWorkflowDraftFailurePayloadInvalid, audit)
	}
	store, ok := service.store.(savedWorkflowDraftLibraryStore)
	if !ok {
		return savedWorkflowDraftLifecycleFailure(SavedWorkflowDraftFailureStoreUnavailable, audit)
	}
	draft, event, err := store.TransitionDraftLifecycle(
		requestContext,
		draftID,
		target,
		request.ExpectedDraftVersion,
		request.ExpectedLifecycleVersion,
		service.now().UTC().Truncate(time.Microsecond),
	)
	result := SavedWorkflowDraftLifecycleTransitionResult{
		CurrentDraftVersion:     draft.DraftVersion,
		CurrentLifecycleVersion: draft.LifecycleVersion,
		CurrentLifecycleState:   draft.LifecycleState,
		RequestAuditMetadata:    audit,
	}
	if err != nil {
		result.FailureCode = savedWorkflowDraftStoreFailureCode(err)
		return result
	}
	if !validSavedWorkflowDraftLifecycleTransition(
		requestContext,
		draft,
		event,
		target,
		request.ExpectedDraftVersion,
		request.ExpectedLifecycleVersion,
	) {
		result.FailureCode = SavedWorkflowDraftFailureLifecycleStoreContract
		return result
	}
	result.Draft = cloneSavedWorkflowDraftPointer(draft)
	result.Event = cloneSavedWorkflowDraftLifecycleEventPointer(event)
	return result
}

func (store *memorySavedWorkflowDraftStore) ListDraftLibraryPage(
	requestContext SavedWorkflowDraftContext,
	filter savedWorkflowDraftLibraryListFilter,
) (savedWorkflowDraftLibraryPage, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if store.unavailable {
		return savedWorkflowDraftLibraryPage{}, errors.New("saved workflow draft store unavailable")
	}
	drafts := make([]SavedWorkflowDraft, 0, len(store.drafts))
	for _, stored := range store.drafts {
		if !savedWorkflowDraftMatchesScope(stored, requestContext.WorkspaceID, requestContext.ApplicationID) {
			continue
		}
		draft, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(stored)
		if !ok {
			return savedWorkflowDraftLibraryPage{}, savedWorkflowDraftStoreOperationFailure(
				SavedWorkflowDraftFailureLifecycleStoreContract,
			)
		}
		if !savedWorkflowDraftMatchesLibraryFilter(draft, filter) ||
			!savedWorkflowDraftIsAfterLibraryCursor(draft, filter) {
			continue
		}
		drafts = append(drafts, draft)
	}
	sort.Slice(drafts, func(i, j int) bool {
		left := mustParseSavedWorkflowDraftLibraryTime(drafts[i].LibraryUpdatedAt)
		right := mustParseSavedWorkflowDraftLibraryTime(drafts[j].LibraryUpdatedAt)
		if left.Equal(right) {
			return drafts[i].DraftID < drafts[j].DraftID
		}
		return left.After(right)
	})
	if len(drafts) > filter.Limit+1 {
		drafts = drafts[:filter.Limit+1]
	}
	hasMore := len(drafts) > filter.Limit
	if hasMore {
		drafts = drafts[:filter.Limit]
	}
	summaries := make([]SavedWorkflowDraftSummary, 0, len(drafts))
	for _, draft := range drafts {
		summaries = append(summaries, savedWorkflowDraftSummaryFromDraft(draft))
	}
	return savedWorkflowDraftLibraryPage{Summaries: summaries, HasMore: hasMore}, nil
}

func (store *memorySavedWorkflowDraftStore) TransitionDraftLifecycle(
	requestContext SavedWorkflowDraftContext,
	draftID string,
	target SavedWorkflowDraftLifecycleState,
	expectedDraftVersion int,
	expectedLifecycleVersion int,
	now time.Time,
) (SavedWorkflowDraft, SavedWorkflowDraftLifecycleEvent, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if store.unavailable {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{}, errors.New(
			"saved workflow draft store unavailable",
		)
	}
	stored, found := store.drafts[draftID]
	if !found {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureNotFound,
		)
	}
	if !savedWorkflowDraftMatchesScope(stored, requestContext.WorkspaceID, requestContext.ApplicationID) {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureScopeDenied,
		)
	}
	current, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(stored)
	if !ok {
		return SavedWorkflowDraft{}, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleStoreContract,
		)
	}
	if current.DraftVersion != expectedDraftVersion {
		return current, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureVersionConflict,
		)
	}
	if current.LifecycleVersion != expectedLifecycleVersion {
		return current, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleVersionConflict,
		)
	}
	if target == current.LifecycleState ||
		(target != SavedWorkflowDraftLifecycleActive && target != SavedWorkflowDraftLifecycleArchived) {
		return current, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleStateConflict,
		)
	}
	occurredAt := now.UTC().Format(time.RFC3339Nano)
	updated := cloneSavedWorkflowDraft(current)
	updated.LifecycleState = target
	updated.LifecycleVersion++
	updated.LibraryUpdatedAt = occurredAt
	updated.LifecycleUpdatedByActorRef = strings.TrimSpace(requestContext.ActorRef)
	if target == SavedWorkflowDraftLifecycleArchived {
		updated.ArchivedAt = occurredAt
	} else {
		updated.ArchivedAt = ""
	}
	transitionKind := SavedWorkflowDraftLifecycleTransitionArchived
	if target == SavedWorkflowDraftLifecycleActive {
		transitionKind = SavedWorkflowDraftLifecycleTransitionUnarchived
	}
	event := SavedWorkflowDraftLifecycleEvent{
		SchemaVersion:    savedWorkflowDraftLifecycleEventSchemaVersion,
		TenantRef:        strings.TrimSpace(requestContext.TenantRef),
		WorkspaceID:      strings.TrimSpace(requestContext.WorkspaceID),
		ApplicationID:    strings.TrimSpace(requestContext.ApplicationID),
		OwnerSubjectRef:  strings.TrimSpace(requestContext.OwnerSubjectRef),
		DraftID:          updated.DraftID,
		LifecycleVersion: updated.LifecycleVersion,
		FromState:        current.LifecycleState,
		ToState:          updated.LifecycleState,
		TransitionKind:   transitionKind,
		OccurredAt:       occurredAt,
		ActorRef:         strings.TrimSpace(requestContext.ActorRef),
		RequestID:        strings.TrimSpace(requestContext.RequestID),
		AuditRef:         strings.TrimSpace(requestContext.AuditRef),
	}
	events := store.lifecycleEvents[draftID]
	if events == nil {
		events = make(map[int]SavedWorkflowDraftLifecycleEvent)
	}
	if _, duplicate := events[event.LifecycleVersion]; duplicate {
		return current, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleStoreContract,
		)
	}
	if store.lifecycleEventWriteFailure {
		return current, SavedWorkflowDraftLifecycleEvent{}, savedWorkflowDraftStoreOperationFailure(
			SavedWorkflowDraftFailureLifecycleEventWrite,
		)
	}
	store.drafts[draftID] = cloneSavedWorkflowDraft(updated)
	events[event.LifecycleVersion] = event
	store.lifecycleEvents[draftID] = events
	store.sideEffects.LifecycleTransitionCount++
	store.sideEffects.LifecycleEventWriteCount++
	return cloneSavedWorkflowDraft(updated), event, nil
}

func normalizeSavedWorkflowDraftLibraryListRequest(
	requestContext SavedWorkflowDraftContext,
	request ListWorkflowDraftsRequest,
) (savedWorkflowDraftLibraryListFilter, SavedWorkflowDraftFailureCode) {
	filter := savedWorkflowDraftLibraryListFilter{
		LifecycleState:  request.LifecycleState,
		Limit:           request.Limit,
		NamePrefix:      strings.TrimSpace(request.NamePrefix),
		ValidationState: request.ValidationState,
		ProvenanceKind:  request.ProvenanceKind,
	}
	if filter.LifecycleState == "" {
		filter.LifecycleState = SavedWorkflowDraftLifecycleActive
	}
	if filter.Limit == 0 {
		filter.Limit = defaultSavedWorkflowDraftListLimit
	}
	if !validSavedWorkflowDraftLifecycleState(filter.LifecycleState) ||
		filter.Limit < 1 ||
		filter.Limit > maxSavedWorkflowDraftListLimit ||
		!utf8.ValidString(filter.NamePrefix) ||
		utf8.RuneCountInString(filter.NamePrefix) > maxSavedWorkflowDraftNamePrefix ||
		!validSavedWorkflowDraftValidationFilter(filter.ValidationState) ||
		!validSavedWorkflowDraftProvenanceFilter(filter.ProvenanceKind) {
		return savedWorkflowDraftLibraryListFilter{}, SavedWorkflowDraftFailureListFilterInvalid
	}
	if strings.TrimSpace(request.Cursor) != "" {
		cursor, err := decodeSavedWorkflowDraftListCursor(requestContext, request, filter)
		if err != nil {
			return savedWorkflowDraftLibraryListFilter{}, SavedWorkflowDraftFailureListCursorInvalid
		}
		filter.AfterUpdatedAt = cursor.LastUpdatedAt
		filter.AfterDraftID = cursor.LastDraftID
	}
	return filter, ""
}

func encodeSavedWorkflowDraftListCursor(
	requestContext SavedWorkflowDraftContext,
	request ListWorkflowDraftsRequest,
	filter savedWorkflowDraftLibraryListFilter,
	last SavedWorkflowDraftSummary,
) (string, error) {
	if !validSavedWorkflowDraftSummaryLifecycle(last) {
		return "", errors.New("invalid saved workflow draft list anchor")
	}
	document := savedWorkflowDraftListCursor{
		SchemaVersion: savedWorkflowDraftListCursorSchemaVersion,
		LastUpdatedAt: last.LibraryUpdatedAt,
		LastDraftID:   last.DraftID,
		BindingDigest: savedWorkflowDraftListCursorDigest(
			requestContext,
			request,
			filter,
			last.LibraryUpdatedAt,
			last.DraftID,
		),
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeSavedWorkflowDraftListCursor(
	requestContext SavedWorkflowDraftContext,
	request ListWorkflowDraftsRequest,
	filter savedWorkflowDraftLibraryListFilter,
) (savedWorkflowDraftListCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(request.Cursor))
	if err != nil {
		return savedWorkflowDraftListCursor{}, errors.New("invalid saved workflow draft list cursor")
	}
	var document savedWorkflowDraftListCursor
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil ||
		decoder.Decode(&struct{}{}) != io.EOF ||
		document.SchemaVersion != savedWorkflowDraftListCursorSchemaVersion ||
		strings.TrimSpace(document.LastDraftID) == "" ||
		strings.TrimSpace(document.LastDraftID) != document.LastDraftID ||
		document.BindingDigest != savedWorkflowDraftListCursorDigest(
			requestContext,
			request,
			filter,
			document.LastUpdatedAt,
			document.LastDraftID,
		) {
		return savedWorkflowDraftListCursor{}, errors.New("invalid saved workflow draft list cursor")
	}
	parsed, err := time.Parse(time.RFC3339Nano, document.LastUpdatedAt)
	if err != nil || parsed.UTC().Format(time.RFC3339Nano) != document.LastUpdatedAt {
		return savedWorkflowDraftListCursor{}, errors.New("invalid saved workflow draft list cursor")
	}
	return document, nil
}

func savedWorkflowDraftListCursorDigest(
	requestContext SavedWorkflowDraftContext,
	_ ListWorkflowDraftsRequest,
	filter savedWorkflowDraftLibraryListFilter,
	lastUpdatedAt string,
	lastDraftID string,
) string {
	binding := strings.Join([]string{
		savedWorkflowDraftListCursorSchemaVersion,
		strings.TrimSpace(requestContext.TenantRef),
		strings.TrimSpace(requestContext.WorkspaceID),
		strings.TrimSpace(requestContext.ApplicationID),
		strings.TrimSpace(requestContext.OwnerSubjectRef),
		string(filter.LifecycleState),
		filter.NamePrefix,
		string(filter.ValidationState),
		string(filter.ProvenanceKind),
		fmt.Sprintf("%d", filter.Limit),
		strings.TrimSpace(lastUpdatedAt),
		strings.TrimSpace(lastDraftID),
	}, "\x00")
	return fmt.Sprintf("%x", sha256.Sum256([]byte(binding)))
}

func normalizeAndValidateSavedWorkflowDraftLifecycle(
	draft SavedWorkflowDraft,
) (SavedWorkflowDraft, bool) {
	normalized := cloneSavedWorkflowDraft(draft)
	legacy := savedWorkflowDraftHasLegacyLifecycle(normalized)
	if legacy {
		normalized.LifecycleState = SavedWorkflowDraftLifecycleActive
		normalized.LifecycleVersion = 1
		normalized.LibraryUpdatedAt = strings.TrimSpace(normalized.UpdatedAt)
		normalized.ProvenanceKind = savedWorkflowDraftProvenanceKind(normalized)
	}
	rawLibraryUpdatedAt := strings.TrimSpace(normalized.LibraryUpdatedAt)
	updatedAt, err := time.Parse(time.RFC3339Nano, rawLibraryUpdatedAt)
	if err != nil {
		return SavedWorkflowDraft{}, false
	}
	normalized.LibraryUpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
	if !legacy && rawLibraryUpdatedAt != normalized.LibraryUpdatedAt {
		return SavedWorkflowDraft{}, false
	}
	normalized.ArchivedAt = strings.TrimSpace(normalized.ArchivedAt)
	if normalized.ArchivedAt != "" {
		rawArchivedAt := normalized.ArchivedAt
		archivedAt, parseErr := time.Parse(time.RFC3339Nano, rawArchivedAt)
		if parseErr != nil {
			return SavedWorkflowDraft{}, false
		}
		normalized.ArchivedAt = archivedAt.UTC().Format(time.RFC3339Nano)
		if rawArchivedAt != normalized.ArchivedAt {
			return SavedWorkflowDraft{}, false
		}
	}
	normalized.LifecycleUpdatedByActorRef = strings.TrimSpace(normalized.LifecycleUpdatedByActorRef)
	if !validSavedWorkflowDraftLifecycleState(normalized.LifecycleState) ||
		normalized.LifecycleVersion < 1 ||
		(normalized.LifecycleState == SavedWorkflowDraftLifecycleActive && normalized.ArchivedAt != "") ||
		(normalized.LifecycleState == SavedWorkflowDraftLifecycleArchived && normalized.ArchivedAt == "") ||
		normalized.ProvenanceKind != savedWorkflowDraftProvenanceKind(normalized) {
		return SavedWorkflowDraft{}, false
	}
	return normalized, true
}

func savedWorkflowDraftProvenanceKind(draft SavedWorkflowDraft) SavedWorkflowDraftProvenanceKind {
	if _, found := normalizeSavedWorkflowTemplateDerivation(
		draft.AdditionalFields[savedWorkflowTemplateDerivationAdditionalField],
	); found {
		return SavedWorkflowDraftProvenanceTemplate
	}
	if _, found := normalizeSavedWorkflowDraftDerivation(
		draft.AdditionalFields[savedWorkflowDraftDerivationAdditionalField],
		draft.DraftID,
	); found {
		return SavedWorkflowDraftProvenanceDraftDerived
	}
	if draft.BaseDefinitionVersion > 0 {
		return SavedWorkflowDraftProvenanceDefinition
	}
	return SavedWorkflowDraftProvenanceUnversioned
}

func activeSavedWorkflowDraftForConsumption(
	draft SavedWorkflowDraft,
) (SavedWorkflowDraft, SavedWorkflowDraftFailureCode) {
	if savedWorkflowDraftHasLegacyLifecycle(draft) {
		legacy := cloneSavedWorkflowDraft(draft)
		legacy.LifecycleState = SavedWorkflowDraftLifecycleActive
		legacy.LifecycleVersion = 1
		legacy.LibraryUpdatedAt = strings.TrimSpace(legacy.UpdatedAt)
		legacy.ProvenanceKind = savedWorkflowDraftProvenanceKind(legacy)
		return legacy, ""
	}
	normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(draft)
	if !ok {
		return SavedWorkflowDraft{}, SavedWorkflowDraftFailureLifecycleStoreContract
	}
	if normalized.LifecycleState != SavedWorkflowDraftLifecycleActive {
		return SavedWorkflowDraft{}, SavedWorkflowDraftFailureArchived
	}
	return normalized, ""
}

func savedWorkflowDraftHasLegacyLifecycle(draft SavedWorkflowDraft) bool {
	return draft.LifecycleState == "" &&
		draft.LifecycleVersion == 0 &&
		strings.TrimSpace(draft.ArchivedAt) == "" &&
		strings.TrimSpace(draft.LibraryUpdatedAt) == "" &&
		strings.TrimSpace(draft.LifecycleUpdatedByActorRef) == "" &&
		draft.ProvenanceKind == ""
}

func savedWorkflowDraftMatchesLibraryFilter(
	draft SavedWorkflowDraft,
	filter savedWorkflowDraftLibraryListFilter,
) bool {
	return draft.LifecycleState == filter.LifecycleState &&
		(filter.NamePrefix == "" || strings.HasPrefix(draft.Name, filter.NamePrefix)) &&
		(filter.ValidationState == "" || draft.ValidationSummary.ValidationState == filter.ValidationState) &&
		(filter.ProvenanceKind == "" || draft.ProvenanceKind == filter.ProvenanceKind)
}

func savedWorkflowDraftIsAfterLibraryCursor(
	draft SavedWorkflowDraft,
	filter savedWorkflowDraftLibraryListFilter,
) bool {
	if filter.AfterUpdatedAt == "" {
		return true
	}
	draftTime := mustParseSavedWorkflowDraftLibraryTime(draft.LibraryUpdatedAt)
	anchorTime := mustParseSavedWorkflowDraftLibraryTime(filter.AfterUpdatedAt)
	return draftTime.Before(anchorTime) ||
		(draftTime.Equal(anchorTime) && draft.DraftID > filter.AfterDraftID)
}

func validSavedWorkflowDraftSummaryLifecycle(summary SavedWorkflowDraftSummary) bool {
	if strings.TrimSpace(summary.DraftID) == "" ||
		!validSavedWorkflowDraftLifecycleState(summary.LifecycleState) ||
		summary.LifecycleVersion < 1 ||
		!validSavedWorkflowDraftProvenanceFilter(summary.ProvenanceKind) ||
		summary.ProvenanceKind == "" {
		return false
	}
	if _, err := time.Parse(time.RFC3339Nano, summary.LibraryUpdatedAt); err != nil {
		return false
	}
	if summary.LifecycleState == SavedWorkflowDraftLifecycleActive {
		return summary.ArchivedAt == ""
	}
	_, err := time.Parse(time.RFC3339Nano, summary.ArchivedAt)
	return err == nil
}

func validSavedWorkflowDraftLifecycleState(state SavedWorkflowDraftLifecycleState) bool {
	return state == SavedWorkflowDraftLifecycleActive || state == SavedWorkflowDraftLifecycleArchived
}

func validSavedWorkflowDraftValidationFilter(state SavedWorkflowDraftStatus) bool {
	switch state {
	case "",
		SavedWorkflowDraftStatusValidForReview,
		SavedWorkflowDraftStatusInvalidDraft,
		SavedWorkflowDraftStatusBlockedCapability,
		SavedWorkflowDraftStatusSchemaUnsupported:
		return true
	default:
		return false
	}
}

func validSavedWorkflowDraftProvenanceFilter(kind SavedWorkflowDraftProvenanceKind) bool {
	switch kind {
	case "",
		SavedWorkflowDraftProvenanceUnversioned,
		SavedWorkflowDraftProvenanceDefinition,
		SavedWorkflowDraftProvenanceDraftDerived,
		SavedWorkflowDraftProvenanceTemplate:
		return true
	default:
		return false
	}
}

func mustParseSavedWorkflowDraftLibraryTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic("validated saved workflow draft library time became invalid")
	}
	return parsed
}

func savedWorkflowDraftListRequestUsesLibraryContract(request ListWorkflowDraftsRequest) bool {
	return request.LifecycleState != "" ||
		request.Limit != 0 ||
		strings.TrimSpace(request.Cursor) != "" ||
		strings.TrimSpace(request.NamePrefix) != "" ||
		request.ValidationState != "" ||
		request.ProvenanceKind != ""
}

func savedWorkflowDraftLifecycleFailure(
	code SavedWorkflowDraftFailureCode,
	audit SavedWorkflowDraftAuditMetadata,
) SavedWorkflowDraftLifecycleTransitionResult {
	return SavedWorkflowDraftLifecycleTransitionResult{
		FailureCode:          code,
		RequestAuditMetadata: audit,
	}
}

func validSavedWorkflowDraftLifecycleTransition(
	requestContext SavedWorkflowDraftContext,
	draft SavedWorkflowDraft,
	event SavedWorkflowDraftLifecycleEvent,
	target SavedWorkflowDraftLifecycleState,
	expectedDraftVersion int,
	expectedLifecycleVersion int,
) bool {
	normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(draft)
	if !ok ||
		normalized.DraftVersion != expectedDraftVersion ||
		normalized.LifecycleVersion != expectedLifecycleVersion+1 ||
		normalized.LifecycleState != target ||
		event.SchemaVersion != savedWorkflowDraftLifecycleEventSchemaVersion ||
		event.TenantRef != strings.TrimSpace(requestContext.TenantRef) ||
		event.WorkspaceID != strings.TrimSpace(requestContext.WorkspaceID) ||
		event.ApplicationID != strings.TrimSpace(requestContext.ApplicationID) ||
		event.OwnerSubjectRef != strings.TrimSpace(requestContext.OwnerSubjectRef) ||
		event.DraftID != normalized.DraftID ||
		event.LifecycleVersion != normalized.LifecycleVersion ||
		event.ToState != normalized.LifecycleState ||
		event.OccurredAt != normalized.LibraryUpdatedAt ||
		event.ActorRef != strings.TrimSpace(requestContext.ActorRef) ||
		event.RequestID != strings.TrimSpace(requestContext.RequestID) ||
		event.AuditRef != strings.TrimSpace(requestContext.AuditRef) {
		return false
	}
	if target == SavedWorkflowDraftLifecycleArchived {
		return event.FromState == SavedWorkflowDraftLifecycleActive &&
			event.TransitionKind == SavedWorkflowDraftLifecycleTransitionArchived &&
			normalized.ArchivedAt == event.OccurredAt
	}
	return event.FromState == SavedWorkflowDraftLifecycleArchived &&
		event.TransitionKind == SavedWorkflowDraftLifecycleTransitionUnarchived &&
		normalized.ArchivedAt == ""
}

func cloneSavedWorkflowDraftLifecycleEventPointer(
	event SavedWorkflowDraftLifecycleEvent,
) *SavedWorkflowDraftLifecycleEvent {
	clone := event
	return &clone
}

var _ savedWorkflowDraftLibraryStore = (*memorySavedWorkflowDraftStore)(nil)
