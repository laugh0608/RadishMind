// Package strictjson validates bounded, concrete-value internal JSON contracts.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

func ObjectFields(data []byte, names ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || len(fields) != len(names) {
		return nil, errors.New("contract fields must match the canonical contract exactly")
	}
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return nil, errors.New("contract field is missing or incorrectly cased")
		}
	}
	return fields, nil
}

// Decode rejects duplicate keys, nulls, unknown fields and trailing values.
// All consumers use concrete values, so null is never an implicit default.
func Decode(data []byte, result any) error {
	if len(data) > 262144 || !utf8.Valid(data) {
		return errors.New("contract JSON exceeds size limit or is not UTF-8")
	}
	tokens := json.NewDecoder(bytes.NewReader(data))
	if err := inspectJSONValue(tokens); err != nil {
		return err
	}
	if _, err := tokens.Token(); err != io.EOF {
		return errors.New("contract JSON has trailing content")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(result)
}

func inspectJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("null is not allowed in contract JSON")
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter != '{' && delimiter != '[' {
		return errors.New("unexpected JSON delimiter")
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("duplicate or invalid contract JSON key")
			}
			seen[name] = true
		}
		if err := inspectJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
