package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	maxNorthboundJSONRequestBodyBytes int64 = 4 << 20
	maxControlJSONRequestBodyBytes    int64 = 1 << 20
)

type jsonRequestBodyOptions struct {
	maxBytes            int64
	rejectUnknownFields bool
}

func (s *Server) decodeJSONRequestBody(
	writer http.ResponseWriter,
	request *http.Request,
	trace requestTrace,
	target any,
	options jsonRequestBodyOptions,
) bool {
	if request.ContentLength > options.maxBytes {
		s.writePlatformError(writer, trace, "REQUEST_BODY_TOO_LARGE", "request body exceeds the endpoint limit")
		return false
	}

	request.Body = http.MaxBytesReader(writer, request.Body, options.maxBytes)
	decoder := json.NewDecoder(request.Body)
	if options.rejectUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(target); err != nil {
		s.writeJSONRequestBodyError(writer, trace, err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("request body contains multiple JSON documents")
		}
		s.writeJSONRequestBodyError(writer, trace, err)
		return false
	}
	return true
}

func (s *Server) writeJSONRequestBodyError(writer http.ResponseWriter, trace requestTrace, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		s.writePlatformError(writer, trace, "REQUEST_BODY_TOO_LARGE", "request body exceeds the endpoint limit")
		return
	}
	s.writePlatformError(writer, trace, "INVALID_JSON", "request body must contain exactly one valid JSON document")
}
