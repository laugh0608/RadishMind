package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func encodeWorkflowRunStorageRecord(record WorkflowRunRecord) ([]byte, time.Time, *time.Time, error) {
	payload, err := json.Marshal(record)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowRunStoreContract
	}
	startedAt, err := time.Parse(time.RFC3339Nano, record.StartedAt)
	if err != nil {
		return nil, time.Time{}, nil, errWorkflowRunStoreContract
	}
	startedAt = startedAt.UTC()
	var completedAt *time.Time
	if record.CompletedAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, record.CompletedAt)
		if parseErr != nil {
			return nil, time.Time{}, nil, errWorkflowRunStoreContract
		}
		parsed = parsed.UTC()
		if parsed.Before(startedAt) {
			return nil, time.Time{}, nil, errWorkflowRunStoreContract
		}
		completedAt = &parsed
	}
	return payload, startedAt, completedAt, nil
}

func decodeWorkflowRunStorageRecord(
	runContext WorkflowRunContext,
	payload []byte,
) (WorkflowRunRecord, error) {
	var record WorkflowRunRecord
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if err := rejectTrailingWorkflowRunJSON(decoder); err != nil {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	if err := validateWorkflowRunStoreRecord(runContext, &record); err != nil || record.RecordVersion <= 0 {
		return WorkflowRunRecord{}, errWorkflowRunStoreContract
	}
	return record, nil
}

func rejectTrailingWorkflowRunJSON(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return errWorkflowRunStoreContract
}

func workflowRunUnixNano(value time.Time) (int64, error) {
	value = value.UTC()
	unixNano := value.UnixNano()
	if !time.Unix(0, unixNano).UTC().Equal(value) {
		return 0, errWorkflowRunStoreContract
	}
	return unixNano, nil
}

func optionalWorkflowRunUnixNano(value *time.Time) (any, error) {
	if value == nil {
		return nil, nil
	}
	return workflowRunUnixNano(value.UTC())
}
