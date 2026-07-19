package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

func encodeWorkflowRAGApplicationRuntimeRecord(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, errWorkflowRAGApplicationStoreContract
	}
	return payload, nil
}

func decodeWorkflowRAGApplicationRuntimeRecord(payload []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errWorkflowRAGApplicationStoreContract
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return errWorkflowRAGApplicationStoreContract
	}
	return nil
}

func workflowRAGApplicationRuntimeUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, errWorkflowRAGApplicationStoreContract
	}
	valueUnixNano, err := workflowRunUnixNano(parsed)
	if err != nil {
		return 0, errWorkflowRAGApplicationStoreContract
	}
	return valueUnixNano, nil
}
