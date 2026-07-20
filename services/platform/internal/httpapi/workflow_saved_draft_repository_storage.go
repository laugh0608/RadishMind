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

func savedWorkflowDraftRecordValues(
	record SavedWorkflowDraftRepositoryStoredRecord,
) ([]byte, []byte, []byte, time.Time, time.Time, SavedWorkflowDraftFailureCode) {
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
