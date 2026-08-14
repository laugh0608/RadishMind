package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

var errSavedWorkflowDraftStoredRecordContract = errors.New(
	"stored saved workflow draft does not match the repository contract",
)

var errSavedWorkflowDraftStoredLibraryProjection = errors.New(
	"stored saved workflow draft library projection does not match the payload",
)

func savedWorkflowDraftRecordValues(
	record SavedWorkflowDraftRepositoryStoredRecord,
) ([]byte, []byte, []byte, time.Time, time.Time, SavedWorkflowDraftFailureCode) {
	normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(record.Draft)
	if !ok {
		return nil, nil, nil, time.Time{}, time.Time{},
			SavedWorkflowDraftFailureLifecycleStoreContract
	}
	record.Draft = normalized
	document := savedWorkflowDraftDocumentPointer(&record.Draft)
	if document == nil {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	payload, err := json.Marshal(document)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	validation, err := json.Marshal(document.ValidationSummary)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	blocked, err := json.Marshal(document.BlockedCapabilitySummary)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	createdAt, err := time.Parse(time.RFC3339, record.Draft.CreatedAt)
	if err != nil {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	updatedAt, err := time.Parse(time.RFC3339, record.Draft.UpdatedAt)
	if err != nil || updatedAt.Before(createdAt) {
		return nil, nil, nil, time.Time{}, time.Time{}, SavedWorkflowDraftFailureStoreContractMismatch
	}
	return payload, validation, blocked, createdAt.UTC(), updatedAt.UTC(), ""
}

func applySavedWorkflowDraftStoredLibraryProjection(
	record SavedWorkflowDraftRepositoryStoredRecord,
	lifecycleState string,
	lifecycleVersion int,
	archivedAt string,
	libraryUpdatedAt string,
	lifecycleUpdatedByActorRef string,
	draftName string,
	validationState string,
	provenanceKind string,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	record.Draft.LifecycleState = SavedWorkflowDraftLifecycleState(lifecycleState)
	record.Draft.LifecycleVersion = lifecycleVersion
	record.Draft.ArchivedAt = archivedAt
	record.Draft.LibraryUpdatedAt = libraryUpdatedAt
	record.Draft.LifecycleUpdatedByActorRef = lifecycleUpdatedByActorRef
	record.Draft.ProvenanceKind = SavedWorkflowDraftProvenanceKind(provenanceKind)
	normalized, ok := normalizeAndValidateSavedWorkflowDraftLifecycle(record.Draft)
	if !ok ||
		normalized.Name != draftName ||
		string(normalized.ValidationSummary.ValidationState) != validationState ||
		string(normalized.ProvenanceKind) != provenanceKind {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredLibraryProjection
	}
	record.Draft = normalized
	return record, nil
}

func decodeSavedWorkflowDraftStoredRecord(
	record SavedWorkflowDraftRepositoryStoredRecord,
	payload []byte,
) (SavedWorkflowDraftRepositoryStoredRecord, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var document savedWorkflowDraftDocument
	if err := decoder.Decode(&document); err != nil {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	record.Draft = savedWorkflowDraftFromDocument(document)
	if record.Draft.DraftID != record.DraftID ||
		record.Draft.WorkspaceID != record.WorkspaceID ||
		record.Draft.ApplicationID != record.ApplicationID ||
		record.Draft.SchemaVersion != strings.TrimSpace(record.Draft.SchemaVersion) ||
		record.Draft.DraftVersion < 1 ||
		strings.TrimSpace(record.Draft.SampleOrUnsavedDraftStatus) != "saved_draft_record" {
		return SavedWorkflowDraftRepositoryStoredRecord{}, errSavedWorkflowDraftStoredRecordContract
	}
	return record, nil
}

func savedWorkflowDraftUnixNano(value time.Time) (int64, error) {
	value = value.UTC()
	unixNano := value.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(value) {
		return 0, errors.New("saved workflow draft time is outside SQLite nanosecond range")
	}
	return unixNano, nil
}
