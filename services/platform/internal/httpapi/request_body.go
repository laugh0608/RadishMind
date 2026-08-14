package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	maxNorthboundJSONRequestBodyBytes int64 = 4 << 20
	maxControlJSONRequestBodyBytes    int64 = 1 << 20
)

type jsonRequestBodyOptions struct {
	maxBytes              int64
	rejectUnknownFields   bool
	rejectDuplicateFields bool
}

func (s *Server) decodeJSONRequestBody(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	target any,
	options jsonRequestBodyOptions,
) bool {
	err := decodeJSONRequestBodyValue(writer, request, target, options)
	if err != nil {
		s.writeJSONRequestBodyError(writer, trace, err)
		return false
	}
	return true
}

func decodeJSONRequestBodyValue(
	writer http.ResponseWriter,
	request *http.Request,
	target any,
	options jsonRequestBodyOptions,
) error {
	if request.ContentLength > options.maxBytes {
		return &http.MaxBytesError{Limit: options.maxBytes}
	}

	request.Body = http.MaxBytesReader(writer, request.Body, options.maxBytes)
	var decoder *json.Decoder
	if options.rejectDuplicateFields {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			return err
		}
		if err := validateNoDuplicateJSONFields(payload); err != nil {
			return err
		}
		decoder = json.NewDecoder(bytes.NewReader(payload))
	} else {
		decoder = json.NewDecoder(request.Body)
	}
	if options.rejectUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("request body contains multiple JSON documents")
		}
		return err
	}
	return nil
}

func validateNoDuplicateJSONFields(payload []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	if err := validateNoDuplicateJSONValue(decoder); err != nil {
		return err
	}
	return nil
}

func validateNoDuplicateJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		fields := make(map[string]struct{})
		for decoder.More() {
			nameToken, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := nameToken.(string)
			if !ok {
				return errors.New("JSON object field name must be a string")
			}
			if _, exists := fields[name]; exists {
				return fmt.Errorf("JSON object field %q is duplicated", name)
			}
			fields[name] = struct{}{}
			if err := validateNoDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case '[':
		for decoder.More() {
			if err := validateNoDuplicateJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	default:
		return errors.New("invalid JSON delimiter")
	}
}

func (s *Server) writeJSONRequestBodyError(writer http.ResponseWriter, trace requestTrace, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		s.writePlatformError(writer, trace, "REQUEST_BODY_TOO_LARGE", "request body exceeds the endpoint limit")
		return
	}
	s.writePlatformError(writer, trace, "INVALID_JSON", "request body must contain exactly one valid JSON document")
}
