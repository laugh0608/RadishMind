package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
)

const actionSafetySnapshotStatusNotRecordedLegacy = "not_recorded_legacy"

// actionSafetyStorageSnapshot is the common durable representation embedded in
// existing owners. An all-empty value is the explicit legacy representation;
// partial values are always corruption and never fall back to current policy.
type actionSafetyStorageSnapshot struct {
	SchemaVersion    string
	ProjectionDigest string
	Payload          []byte
}

func (snapshot actionSafetyStorageSnapshot) isLegacy() bool {
	return snapshot.SchemaVersion == "" && snapshot.ProjectionDigest == "" && len(snapshot.Payload) == 0
}

func (snapshot actionSafetyStorageSnapshot) columnValues() (any, any, any) {
	if snapshot.isLegacy() {
		return nil, nil, nil
	}
	return snapshot.SchemaVersion, snapshot.ProjectionDigest, snapshot.Payload
}

func (snapshot actionSafetyStorageSnapshot) sqliteColumnValues() (any, any, any) {
	if snapshot.isLegacy() {
		return nil, nil, nil
	}
	return snapshot.SchemaVersion, snapshot.ProjectionDigest, string(snapshot.Payload)
}

func encodeActionSafetyAssignmentSnapshot(value *ActionSafetyAssignmentProjectionV1) (actionSafetyStorageSnapshot, error) {
	if value == nil {
		return actionSafetyStorageSnapshot{}, nil
	}
	if validateActionSafetyAssignmentProjection(*value) != nil {
		return actionSafetyStorageSnapshot{}, errActionSafetyProjectionContract
	}
	return marshalActionSafetyStorageSnapshot(value.SchemaVersion, value.ProjectionDigest, value)
}

func encodeActionSafetyPlanSnapshot(value *ActionSafetyPlanProjectionV1) (actionSafetyStorageSnapshot, error) {
	if value == nil {
		return actionSafetyStorageSnapshot{}, nil
	}
	if validateActionSafetyPlanProjection(*value) != nil {
		return actionSafetyStorageSnapshot{}, errActionSafetyProjectionContract
	}
	return marshalActionSafetyStorageSnapshot(value.SchemaVersion, value.ProjectionDigest, value)
}

func encodeActionSafetyRunSnapshot(value *ActionSafetyRunProjectionV1) (actionSafetyStorageSnapshot, error) {
	if value == nil {
		return actionSafetyStorageSnapshot{}, nil
	}
	if validateActionSafetyRunProjection(*value) != nil {
		return actionSafetyStorageSnapshot{}, errActionSafetyProjectionContract
	}
	return marshalActionSafetyStorageSnapshot(value.SchemaVersion, value.ProjectionDigest, value)
}

func marshalActionSafetyStorageSnapshot(schemaVersion, projectionDigest string, value any) (actionSafetyStorageSnapshot, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return actionSafetyStorageSnapshot{}, errActionSafetyProjectionContract
	}
	return actionSafetyStorageSnapshot{
		SchemaVersion: schemaVersion, ProjectionDigest: projectionDigest, Payload: payload,
	}, nil
}

func decodeActionSafetyAssignmentSnapshot(snapshot actionSafetyStorageSnapshot) (*ActionSafetyAssignmentProjectionV1, error) {
	if snapshot.isLegacy() {
		return nil, nil
	}
	var value ActionSafetyAssignmentProjectionV1
	if snapshot.SchemaVersion != actionSafetyAssignmentProjectionSchema ||
		decodeStrictActionSafetySnapshot(snapshot.Payload, &value) != nil ||
		value.SchemaVersion != snapshot.SchemaVersion || value.ProjectionDigest != snapshot.ProjectionDigest ||
		validateActionSafetyAssignmentProjection(value) != nil {
		return nil, errActionSafetyProjectionContract
	}
	return &value, nil
}

func decodeActionSafetyPlanSnapshot(snapshot actionSafetyStorageSnapshot) (*ActionSafetyPlanProjectionV1, error) {
	if snapshot.isLegacy() {
		return nil, nil
	}
	var value ActionSafetyPlanProjectionV1
	if snapshot.SchemaVersion != actionSafetyPlanProjectionSchema ||
		decodeStrictActionSafetySnapshot(snapshot.Payload, &value) != nil ||
		value.SchemaVersion != snapshot.SchemaVersion || value.ProjectionDigest != snapshot.ProjectionDigest ||
		validateActionSafetyPlanProjection(value) != nil {
		return nil, errActionSafetyProjectionContract
	}
	return &value, nil
}

func decodeActionSafetyRunSnapshot(snapshot actionSafetyStorageSnapshot) (*ActionSafetyRunProjectionV1, error) {
	if snapshot.isLegacy() {
		return nil, nil
	}
	var value ActionSafetyRunProjectionV1
	if snapshot.SchemaVersion != actionSafetyRunProjectionSchema ||
		decodeStrictActionSafetySnapshot(snapshot.Payload, &value) != nil ||
		value.SchemaVersion != snapshot.SchemaVersion || value.ProjectionDigest != snapshot.ProjectionDigest ||
		validateActionSafetyRunProjection(value) != nil {
		return nil, errActionSafetyProjectionContract
	}
	return &value, nil
}

func decodeWorkflowRunStorageRecordWithActionSafety(
	ctx WorkflowRunContext,
	payload []byte,
	snapshot actionSafetyStorageSnapshot,
) (WorkflowRunRecord, error) {
	record, err := decodeWorkflowRunStorageRecord(ctx, payload)
	if err != nil {
		return WorkflowRunRecord{}, err
	}
	record.ActionSafety, err = decodeActionSafetyRunSnapshot(snapshot)
	if err != nil || validateWorkflowRunStoreRecord(ctx, &record) != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func decodeStrictActionSafetySnapshot(payload []byte, target any) error {
	if len(payload) == 0 || validateNoDuplicateJSONFields(payload) != nil {
		return errActionSafetyProjectionContract
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errActionSafetyProjectionContract
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errActionSafetyProjectionContract
	}
	return nil
}

func actionSafetySnapshotStatus(snapshot actionSafetyStorageSnapshot) string {
	if snapshot.isLegacy() {
		return actionSafetySnapshotStatusNotRecordedLegacy
	}
	return snapshot.SchemaVersion
}

func workflowRunStoreSupportsActionSafety(store workflowRunStore) bool {
	switch typed := store.(type) {
	case *memoryWorkflowRunStore, *sqliteWorkflowRunStore, *postgresWorkflowRunStore:
		return true
	case *sqlitePromptApplicationRunStore:
		return typed.hasActionSafetySnapshot()
	case *postgresPromptApplicationRunStore:
		return typed.hasActionSafetySnapshot()
	default:
		return false
	}
}
