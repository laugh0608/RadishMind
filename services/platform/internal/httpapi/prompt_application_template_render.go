package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	PromptApplicationTemplateFailurePayloadInvalid        = "prompt_template_payload_invalid"
	PromptApplicationTemplateFailureSecretForbidden       = "prompt_template_secret_material_forbidden"
	PromptApplicationTemplateFailureSyntaxInvalid         = "prompt_template_syntax_invalid"
	PromptApplicationTemplateFailureVariableInvalid       = "prompt_template_variable_invalid"
	PromptApplicationTemplateFailureOutputContractInvalid = "prompt_template_output_contract_invalid"
	PromptApplicationInvocationFailureOutputContract      = "prompt_invocation_output_contract_failed"
)

const (
	promptApplicationTemplateMaximumMessages        = 16
	promptApplicationTemplateMaximumMessageBytes    = 16 * 1024
	promptApplicationTemplateMaximumSourceBytes     = 64 * 1024
	promptApplicationTemplateMaximumVariables       = 64
	promptApplicationTemplateMaximumVariableBytes   = 16 * 1024
	promptApplicationTemplateMaximumStringListItems = 128
	promptApplicationTemplateMaximumStringListBytes = 32 * 1024
	promptApplicationTemplateMaximumRenderedBytes   = 128 * 1024
	promptApplicationOutputMaximumBytes             = 64 * 1024
	promptApplicationOutputSchemaMaximumBytes       = 32 * 1024
	promptApplicationOutputSchemaMaximumDepth       = 8
	promptApplicationOutputSchemaMaximumProperties  = 128
)

const (
	promptApplicationVariableString     = "string"
	promptApplicationVariableInteger    = "integer"
	promptApplicationVariableNumber     = "number"
	promptApplicationVariableBoolean    = "boolean"
	promptApplicationVariableStringList = "string_list"
	promptApplicationOutputText         = "text"
	promptApplicationOutputJSONObject   = "json_object"
)

var promptApplicationVariableNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)

type PromptApplicationTemplateMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptApplicationTemplateVariable struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Required     bool   `json:"required"`
	Description  string `json:"description"`
	DefaultValue any    `json:"default_value,omitempty"`
}

type PromptApplicationJSONSchema struct {
	Type                 string                                 `json:"type"`
	Properties           map[string]PromptApplicationJSONSchema `json:"properties,omitempty"`
	Required             []string                               `json:"required,omitempty"`
	AdditionalProperties bool                                   `json:"additionalProperties"`
	Items                *PromptApplicationJSONSchema           `json:"items,omitempty"`
}

type PromptApplicationOutputContract struct {
	Kind       string                       `json:"kind"`
	AllowEmpty bool                         `json:"allow_empty"`
	MaxBytes   int                          `json:"max_bytes"`
	JSONSchema *PromptApplicationJSONSchema `json:"json_schema,omitempty"`
}

type PromptApplicationTemplateSource struct {
	Messages       []PromptApplicationTemplateMessage  `json:"messages"`
	Variables      []PromptApplicationTemplateVariable `json:"variables"`
	OutputContract PromptApplicationOutputContract     `json:"output_contract"`
}

type PromptApplicationTemplateFinding struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Summary string `json:"summary"`
}

type PromptApplicationTemplateRenderResult struct {
	Messages       []PromptApplicationTemplateMessage `json:"messages"`
	RenderedDigest string                             `json:"rendered_digest"`
	FailureCode    string                             `json:"failure_code,omitempty"`
	Findings       []PromptApplicationTemplateFinding `json:"findings"`
}

type promptApplicationTemplateToken struct {
	literal      string
	variableName string
}

func ValidatePromptApplicationTemplateSource(source PromptApplicationTemplateSource) []PromptApplicationTemplateFinding {
	findings := make([]PromptApplicationTemplateFinding, 0)
	if len(source.Messages) == 0 || len(source.Messages) > promptApplicationTemplateMaximumMessages {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, "messages", "message count is outside the supported budget")
	}
	if len(source.Variables) > promptApplicationTemplateMaximumVariables {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, "variables", "variable count exceeds the supported budget")
	}

	variables, variableFindings := validatePromptApplicationVariables(source.Variables)
	findings = append(findings, variableFindings...)
	totalSourceBytes := 0
	usedVariables := make(map[string]struct{})
	for index, message := range source.Messages {
		field := fmt.Sprintf("messages[%d]", index)
		role := strings.TrimSpace(message.Role)
		if role != "system" && role != "developer" && role != "user" {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, field+".role", "message role is not supported")
		}
		if !utf8.ValidString(message.Content) || message.Content == "" || len([]byte(message.Content)) > promptApplicationTemplateMaximumMessageBytes {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, field+".content", "message content is empty, invalid UTF-8, or over budget")
		}
		totalSourceBytes += len([]byte(message.Content))
		if promptApplicationContainsSecretMaterial(message.Content) {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureSecretForbidden, field+".content", "message content contains credential-like material")
		}
		tokens, err := parsePromptApplicationTemplate(message.Content)
		if err != nil {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureSyntaxInvalid, field+".content", "message contains unsupported template syntax")
			continue
		}
		for _, token := range tokens {
			if token.variableName == "" {
				continue
			}
			if _, declared := variables[token.variableName]; !declared {
				findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, field+".content", "message references an undeclared variable")
				continue
			}
			usedVariables[token.variableName] = struct{}{}
		}
	}
	if totalSourceBytes > promptApplicationTemplateMaximumSourceBytes {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailurePayloadInvalid, "messages", "template source exceeds the supported budget")
	}
	for _, name := range sortedPromptApplicationVariableNames(variables) {
		if _, used := usedVariables[name]; !used {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, "variables."+name, "declared variable is not referenced by any message")
		}
	}
	findings = append(findings, validatePromptApplicationOutputContract(source.OutputContract)...)
	return findings
}

func NormalizePromptApplicationTemplateSource(source PromptApplicationTemplateSource) (PromptApplicationTemplateSource, error) {
	if findings := ValidatePromptApplicationTemplateSource(source); len(findings) != 0 {
		return PromptApplicationTemplateSource{}, errors.New(findings[0].Code)
	}
	normalized := PromptApplicationTemplateSource{
		Messages:  make([]PromptApplicationTemplateMessage, 0, len(source.Messages)),
		Variables: make([]PromptApplicationTemplateVariable, 0, len(source.Variables)),
		OutputContract: PromptApplicationOutputContract{
			Kind:       strings.TrimSpace(source.OutputContract.Kind),
			AllowEmpty: source.OutputContract.AllowEmpty,
			MaxBytes:   normalizedPromptApplicationOutputMaxBytes(source.OutputContract.MaxBytes),
			JSONSchema: normalizePromptApplicationJSONSchema(source.OutputContract.JSONSchema),
		},
	}
	for _, message := range source.Messages {
		normalized.Messages = append(normalized.Messages, PromptApplicationTemplateMessage{Role: strings.TrimSpace(message.Role), Content: message.Content})
	}
	for _, variable := range source.Variables {
		value, err := normalizePromptApplicationDefaultValue(variable.Type, variable.DefaultValue)
		if err != nil {
			return PromptApplicationTemplateSource{}, err
		}
		normalized.Variables = append(normalized.Variables, PromptApplicationTemplateVariable{
			Name: strings.TrimSpace(variable.Name), Type: strings.TrimSpace(variable.Type), Required: variable.Required,
			Description: strings.TrimSpace(variable.Description), DefaultValue: value,
		})
	}
	sort.Slice(normalized.Variables, func(left, right int) bool {
		return normalized.Variables[left].Name < normalized.Variables[right].Name
	})
	return normalized, nil
}

func RenderPromptApplicationTemplate(source PromptApplicationTemplateSource, input map[string]any) PromptApplicationTemplateRenderResult {
	findings := ValidatePromptApplicationTemplateSource(source)
	if len(findings) != 0 {
		return promptApplicationRenderFailure(findings)
	}
	variables := make(map[string]PromptApplicationTemplateVariable, len(source.Variables))
	for _, variable := range source.Variables {
		variable.Name = strings.TrimSpace(variable.Name)
		variable.Type = strings.TrimSpace(variable.Type)
		variables[variable.Name] = variable
	}
	for _, name := range sortedPromptApplicationInputNames(input) {
		if _, declared := variables[name]; !declared {
			return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureVariableInvalid, Field: "input." + name, Summary: "input contains an undeclared variable"}})
		}
	}

	canonicalValues := make(map[string]string, len(variables))
	for _, name := range sortedPromptApplicationVariableNames(variables) {
		variable := variables[name]
		value, provided := input[name]
		if !provided && variable.DefaultValue != nil {
			value, provided = variable.DefaultValue, true
		}
		if !provided {
			if variable.Required {
				return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureVariableInvalid, Field: "input." + name, Summary: "required variable is missing"}})
			}
			canonicalValues[name] = ""
			continue
		}
		canonical, err := canonicalPromptApplicationVariableValue(variable.Type, value)
		if err != nil {
			return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureVariableInvalid, Field: "input." + name, Summary: "variable value does not match its declared type or budget"}})
		}
		if promptApplicationContainsSecretMaterial(canonical) {
			return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureSecretForbidden, Field: "input." + name, Summary: "variable value contains credential-like material"}})
		}
		canonicalValues[name] = canonical
	}

	rendered := make([]PromptApplicationTemplateMessage, 0, len(source.Messages))
	totalBytes := 0
	for _, message := range source.Messages {
		tokens, err := parsePromptApplicationTemplate(message.Content)
		if err != nil {
			return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureSyntaxInvalid, Field: "messages", Summary: "message contains unsupported template syntax"}})
		}
		content := renderPromptApplicationTokens(tokens, canonicalValues)
		totalBytes += len([]byte(content))
		rendered = append(rendered, PromptApplicationTemplateMessage{Role: strings.TrimSpace(message.Role), Content: content})
	}
	if totalBytes > promptApplicationTemplateMaximumRenderedBytes {
		return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailureVariableInvalid, Field: "input", Summary: "rendered messages exceed the supported budget"}})
	}
	digest, err := promptApplicationRenderedMessagesDigest(rendered)
	if err != nil {
		return promptApplicationRenderFailure([]PromptApplicationTemplateFinding{{Code: PromptApplicationTemplateFailurePayloadInvalid, Field: "messages", Summary: "rendered messages could not be canonicalized"}})
	}
	return PromptApplicationTemplateRenderResult{Messages: rendered, RenderedDigest: digest, Findings: []PromptApplicationTemplateFinding{}}
}

func ValidatePromptApplicationOutput(contract PromptApplicationOutputContract, output string) string {
	if len(validatePromptApplicationOutputContract(contract)) != 0 || !utf8.ValidString(output) {
		return PromptApplicationInvocationFailureOutputContract
	}
	maxBytes := normalizedPromptApplicationOutputMaxBytes(contract.MaxBytes)
	if len([]byte(output)) > maxBytes || !contract.AllowEmpty && strings.TrimSpace(output) == "" {
		return PromptApplicationInvocationFailureOutputContract
	}
	if contract.Kind == promptApplicationOutputText {
		return ""
	}
	value, err := decodePromptApplicationJSONObject(output)
	if err != nil || validatePromptApplicationJSONValue(*contract.JSONSchema, value) != nil {
		return PromptApplicationInvocationFailureOutputContract
	}
	return ""
}

func validatePromptApplicationVariables(variables []PromptApplicationTemplateVariable) (map[string]PromptApplicationTemplateVariable, []PromptApplicationTemplateFinding) {
	byName := make(map[string]PromptApplicationTemplateVariable, len(variables))
	findings := make([]PromptApplicationTemplateFinding, 0)
	for index, variable := range variables {
		field := fmt.Sprintf("variables[%d]", index)
		variable.Name = strings.TrimSpace(variable.Name)
		variable.Type = strings.TrimSpace(variable.Type)
		variable.Description = strings.TrimSpace(variable.Description)
		if !promptApplicationVariableNamePattern.MatchString(variable.Name) || !supportedPromptApplicationVariableType(variable.Type) {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, field, "variable name or type is invalid")
			continue
		}
		if _, duplicate := byName[variable.Name]; duplicate {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, field+".name", "variable name is duplicated")
			continue
		}
		if !utf8.ValidString(variable.Description) || len([]byte(variable.Description)) > 512 || promptApplicationContainsSecretMaterial(variable.Description) {
			code := PromptApplicationTemplateFailureVariableInvalid
			if promptApplicationContainsSecretMaterial(variable.Description) {
				code = PromptApplicationTemplateFailureSecretForbidden
			}
			findings = appendPromptApplicationFinding(findings, code, field+".description", "variable description is invalid or contains credential-like material")
		}
		if variable.DefaultValue != nil {
			if variable.Required {
				findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, field+".default_value", "required variables cannot declare a default value")
			}
			canonical, err := canonicalPromptApplicationVariableValue(variable.Type, variable.DefaultValue)
			if err != nil {
				findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureVariableInvalid, field+".default_value", "default value does not match the declared type or budget")
			} else if promptApplicationContainsSecretMaterial(canonical) {
				findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureSecretForbidden, field+".default_value", "default value contains credential-like material")
			}
		}
		byName[variable.Name] = variable
	}
	return byName, findings
}

func validatePromptApplicationOutputContract(contract PromptApplicationOutputContract) []PromptApplicationTemplateFinding {
	findings := make([]PromptApplicationTemplateFinding, 0)
	if contract.Kind != promptApplicationOutputText && contract.Kind != promptApplicationOutputJSONObject {
		return appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.kind", "output contract kind is not supported")
	}
	if contract.MaxBytes < 0 || contract.MaxBytes > promptApplicationOutputMaximumBytes {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.max_bytes", "output byte budget is invalid")
	}
	if contract.Kind == promptApplicationOutputText {
		if contract.JSONSchema != nil {
			findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.json_schema", "text output cannot declare a JSON schema")
		}
		return findings
	}
	if contract.JSONSchema == nil {
		return appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.json_schema", "json_object output requires a schema")
	}
	if contract.JSONSchema.Type != "object" {
		return appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.json_schema", "json_object output requires an object root schema")
	}
	encoded, err := json.Marshal(contract.JSONSchema)
	if err != nil || len(encoded) > promptApplicationOutputSchemaMaximumBytes {
		return appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.json_schema", "JSON schema cannot be canonicalized within the supported budget")
	}
	propertyCount := 0
	if err := validatePromptApplicationJSONSchema(*contract.JSONSchema, 1, &propertyCount); err != nil {
		findings = appendPromptApplicationFinding(findings, PromptApplicationTemplateFailureOutputContractInvalid, "output_contract.json_schema", "JSON schema uses an unsupported or invalid shape")
	}
	return findings
}

func validatePromptApplicationJSONSchema(schema PromptApplicationJSONSchema, depth int, propertyCount *int) error {
	if depth > promptApplicationOutputSchemaMaximumDepth {
		return errors.New("schema depth exceeds budget")
	}
	switch schema.Type {
	case "object":
		if schema.Items != nil {
			return errors.New("object schema cannot declare items")
		}
		*propertyCount += len(schema.Properties)
		if *propertyCount > promptApplicationOutputSchemaMaximumProperties {
			return errors.New("schema properties exceed budget")
		}
		required := make(map[string]struct{}, len(schema.Required))
		for _, name := range schema.Required {
			if !promptApplicationVariableNamePattern.MatchString(name) {
				return errors.New("required property name is invalid")
			}
			if _, duplicate := required[name]; duplicate {
				return errors.New("required property is duplicated")
			}
			if _, declared := schema.Properties[name]; !declared {
				return errors.New("required property is not declared")
			}
			required[name] = struct{}{}
		}
		for _, name := range sortedPromptApplicationSchemaPropertyNames(schema.Properties) {
			property := schema.Properties[name]
			if !promptApplicationVariableNamePattern.MatchString(name) || validatePromptApplicationJSONSchema(property, depth+1, propertyCount) != nil {
				return errors.New("object property schema is invalid")
			}
		}
	case "array":
		if schema.Items == nil || len(schema.Properties) != 0 || len(schema.Required) != 0 || schema.AdditionalProperties {
			return errors.New("array schema is invalid")
		}
		return validatePromptApplicationJSONSchema(*schema.Items, depth+1, propertyCount)
	case "string", "integer", "number", "boolean":
		if len(schema.Properties) != 0 || len(schema.Required) != 0 || schema.Items != nil || schema.AdditionalProperties {
			return errors.New("scalar schema contains object or array fields")
		}
	default:
		return errors.New("schema type is unsupported")
	}
	return nil
}

func validatePromptApplicationJSONValue(schema PromptApplicationJSONSchema, value any) error {
	switch schema.Type {
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return errors.New("value is not an object")
		}
		for _, name := range schema.Required {
			if _, found := object[name]; !found {
				return errors.New("required property is missing")
			}
		}
		for _, name := range sortedPromptApplicationValueNames(object) {
			propertyValue := object[name]
			propertySchema, declared := schema.Properties[name]
			if !declared {
				if schema.AdditionalProperties {
					continue
				}
				return errors.New("additional property is forbidden")
			}
			if err := validatePromptApplicationJSONValue(propertySchema, propertyValue); err != nil {
				return err
			}
		}
	case "array":
		values, ok := value.([]any)
		if !ok || schema.Items == nil {
			return errors.New("value is not an array")
		}
		for _, item := range values {
			if err := validatePromptApplicationJSONValue(*schema.Items, item); err != nil {
				return err
			}
		}
	case "string":
		if _, ok := value.(string); !ok {
			return errors.New("value is not a string")
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return errors.New("value is not a boolean")
		}
	case "number":
		if _, ok := value.(json.Number); !ok {
			return errors.New("value is not a number")
		}
	case "integer":
		number, ok := value.(json.Number)
		if !ok || !promptApplicationJSONNumberIsInteger(number) {
			return errors.New("value is not an integer")
		}
	default:
		return errors.New("schema type is unsupported")
	}
	return nil
}

func parsePromptApplicationTemplate(content string) ([]promptApplicationTemplateToken, error) {
	tokens := make([]promptApplicationTemplateToken, 0)
	remaining := content
	for remaining != "" {
		opening := strings.Index(remaining, "{{")
		closing := strings.Index(remaining, "}}")
		if closing >= 0 && (opening < 0 || closing < opening) {
			return nil, errors.New("unmatched closing delimiter")
		}
		if opening < 0 {
			tokens = append(tokens, promptApplicationTemplateToken{literal: remaining})
			break
		}
		if opening > 0 {
			tokens = append(tokens, promptApplicationTemplateToken{literal: remaining[:opening]})
		}
		afterOpening := remaining[opening+2:]
		closing = strings.Index(afterOpening, "}}")
		if closing < 0 {
			return nil, errors.New("unmatched opening delimiter")
		}
		name := strings.TrimSpace(afterOpening[:closing])
		if !promptApplicationVariableNamePattern.MatchString(name) {
			return nil, errors.New("unsupported placeholder")
		}
		tokens = append(tokens, promptApplicationTemplateToken{variableName: name})
		remaining = afterOpening[closing+2:]
	}
	return tokens, nil
}

func renderPromptApplicationTokens(tokens []promptApplicationTemplateToken, values map[string]string) string {
	var builder strings.Builder
	for _, token := range tokens {
		if token.variableName == "" {
			builder.WriteString(token.literal)
			continue
		}
		builder.WriteString(values[token.variableName])
	}
	return builder.String()
}

func canonicalPromptApplicationVariableValue(variableType string, value any) (string, error) {
	switch variableType {
	case promptApplicationVariableString:
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) || len([]byte(text)) > promptApplicationTemplateMaximumVariableBytes {
			return "", errors.New("invalid string variable")
		}
		return text, nil
	case promptApplicationVariableInteger:
		return canonicalPromptApplicationInteger(value)
	case promptApplicationVariableNumber:
		return canonicalPromptApplicationNumber(value)
	case promptApplicationVariableBoolean:
		boolean, ok := value.(bool)
		if !ok {
			return "", errors.New("invalid boolean variable")
		}
		return strconv.FormatBool(boolean), nil
	case promptApplicationVariableStringList:
		return canonicalPromptApplicationStringList(value)
	default:
		return "", errors.New("unsupported variable type")
	}
}

func canonicalPromptApplicationInteger(value any) (string, error) {
	switch number := value.(type) {
	case int:
		return strconv.FormatInt(int64(number), 10), nil
	case int8:
		return strconv.FormatInt(int64(number), 10), nil
	case int16:
		return strconv.FormatInt(int64(number), 10), nil
	case int32:
		return strconv.FormatInt(int64(number), 10), nil
	case int64:
		return strconv.FormatInt(number, 10), nil
	case uint:
		return strconv.FormatUint(uint64(number), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(number), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(number), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(number), 10), nil
	case uint64:
		return strconv.FormatUint(number, 10), nil
	case float32:
		value := float64(number)
		if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
			return "", errors.New("number is not an integer")
		}
		return strconv.FormatFloat(value, 'f', -1, 32), nil
	case float64:
		if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
			return "", errors.New("number is not an integer")
		}
		return strconv.FormatFloat(number, 'f', -1, 64), nil
	case json.Number:
		if !promptApplicationJSONNumberIsInteger(number) {
			return "", errors.New("number is not an integer")
		}
		integer := new(big.Int)
		if _, ok := integer.SetString(number.String(), 10); ok {
			return integer.String(), nil
		}
		rational, ok := new(big.Rat).SetString(number.String())
		if !ok || !rational.IsInt() {
			return "", errors.New("number is not an integer")
		}
		return rational.Num().String(), nil
	default:
		return "", errors.New("invalid integer variable")
	}
}

func canonicalPromptApplicationNumber(value any) (string, error) {
	var number string
	switch typed := value.(type) {
	case json.Number:
		number = typed.String()
	case float32:
		number = strconv.FormatFloat(float64(typed), 'g', -1, 32)
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return "", errors.New("number must be finite")
		}
		number = strconv.FormatFloat(typed, 'g', -1, 64)
	default:
		return canonicalPromptApplicationInteger(value)
	}
	if _, ok := new(big.Rat).SetString(number); !ok {
		return "", errors.New("invalid number variable")
	}
	parsed, err := strconv.ParseFloat(number, 64)
	if err != nil || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
		return "", errors.New("number must be finite")
	}
	if parsed == 0 {
		return "0", nil
	}
	return strconv.FormatFloat(parsed, 'g', -1, 64), nil
}

func canonicalPromptApplicationStringList(value any) (string, error) {
	items := make([]string, 0)
	switch list := value.(type) {
	case []string:
		items = append(items, list...)
	case []any:
		for _, item := range list {
			text, ok := item.(string)
			if !ok {
				return "", errors.New("string_list contains a non-string item")
			}
			items = append(items, text)
		}
	default:
		return "", errors.New("invalid string_list variable")
	}
	if len(items) > promptApplicationTemplateMaximumStringListItems {
		return "", errors.New("string_list exceeds item budget")
	}
	for _, item := range items {
		if !utf8.ValidString(item) || len([]byte(item)) > promptApplicationTemplateMaximumVariableBytes {
			return "", errors.New("string_list item is invalid")
		}
	}
	encoded, err := json.Marshal(items)
	if err != nil || len(encoded) > promptApplicationTemplateMaximumStringListBytes {
		return "", errors.New("string_list exceeds byte budget")
	}
	return string(encoded), nil
}

func promptApplicationRenderedMessagesDigest(messages []PromptApplicationTemplateMessage) (string, error) {
	encoded, err := json.Marshal(messages)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func decodePromptApplicationJSONObject(output string) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(output))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil || value == nil {
		return nil, errors.New("output is not a JSON object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("output contains trailing JSON")
	}
	return value, nil
}

func promptApplicationJSONNumberIsInteger(number json.Number) bool {
	rational, ok := new(big.Rat).SetString(number.String())
	return ok && rational.IsInt()
}

func normalizedPromptApplicationOutputMaxBytes(maxBytes int) int {
	if maxBytes == 0 {
		return promptApplicationOutputMaximumBytes
	}
	return maxBytes
}

func normalizePromptApplicationDefaultValue(variableType string, value any) (any, error) {
	if value == nil {
		return nil, nil
	}
	canonical, err := canonicalPromptApplicationVariableValue(strings.TrimSpace(variableType), value)
	if err != nil {
		return nil, err
	}
	switch strings.TrimSpace(variableType) {
	case promptApplicationVariableString:
		return canonical, nil
	case promptApplicationVariableInteger, promptApplicationVariableNumber:
		return json.Number(canonical), nil
	case promptApplicationVariableBoolean:
		return strconv.ParseBool(canonical)
	case promptApplicationVariableStringList:
		var items []string
		if err := json.Unmarshal([]byte(canonical), &items); err != nil {
			return nil, err
		}
		return items, nil
	default:
		return nil, errors.New("unsupported variable type")
	}
}

func normalizePromptApplicationJSONSchema(schema *PromptApplicationJSONSchema) *PromptApplicationJSONSchema {
	if schema == nil {
		return nil
	}
	normalized := &PromptApplicationJSONSchema{
		Type: strings.TrimSpace(schema.Type), AdditionalProperties: schema.AdditionalProperties,
		Properties: make(map[string]PromptApplicationJSONSchema, len(schema.Properties)),
		Required:   append([]string{}, schema.Required...),
	}
	sort.Strings(normalized.Required)
	for name, property := range schema.Properties {
		propertyCopy := normalizePromptApplicationJSONSchema(&property)
		if propertyCopy != nil {
			normalized.Properties[name] = *propertyCopy
		}
	}
	normalized.Items = normalizePromptApplicationJSONSchema(schema.Items)
	return normalized
}

func supportedPromptApplicationVariableType(variableType string) bool {
	switch variableType {
	case promptApplicationVariableString, promptApplicationVariableInteger, promptApplicationVariableNumber, promptApplicationVariableBoolean, promptApplicationVariableStringList:
		return true
	default:
		return false
	}
}

func promptApplicationContainsSecretMaterial(value string) bool {
	lower := strings.ToLower(value)
	markers := []string{
		"authorization:", "bearer ", "api_key=", "api-key=", "x-api-key:", "cookie:", "set-cookie:",
		"dsn=", "database_url=", "postgres://", "postgresql://", "mysql://", "mongodb://", "redis://", "sk-",
	}
	for _, marker := range markers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func appendPromptApplicationFinding(findings []PromptApplicationTemplateFinding, code, field, summary string) []PromptApplicationTemplateFinding {
	return append(findings, PromptApplicationTemplateFinding{Code: code, Field: field, Summary: summary})
}

func promptApplicationRenderFailure(findings []PromptApplicationTemplateFinding) PromptApplicationTemplateRenderResult {
	code := PromptApplicationTemplateFailurePayloadInvalid
	if len(findings) != 0 {
		code = findings[0].Code
	}
	return PromptApplicationTemplateRenderResult{Messages: []PromptApplicationTemplateMessage{}, FailureCode: code, Findings: findings}
}

func sortedPromptApplicationSchemaPropertyNames(properties map[string]PromptApplicationJSONSchema) []string {
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedPromptApplicationVariableNames(variables map[string]PromptApplicationTemplateVariable) []string {
	names := make([]string, 0, len(variables))
	for name := range variables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedPromptApplicationInputNames(input map[string]any) []string {
	names := make([]string, 0, len(input))
	for name := range input {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedPromptApplicationValueNames(value map[string]any) []string {
	names := make([]string, 0, len(value))
	for name := range value {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
