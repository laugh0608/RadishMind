package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPromptApplicationTemplateRendererIsDeterministicAndDoesNotReinterpretValues(t *testing.T) {
	source := PromptApplicationTemplateSource{
		Messages: []PromptApplicationTemplateMessage{
			{Role: "system", Content: "Answer for {{ audience }}."},
			{Role: "user", Content: "count={{ count }}; ratio={{ ratio }}; enabled={{ enabled }}; tags={{ tags }}; literal={{ literal }}"},
		},
		Variables: []PromptApplicationTemplateVariable{
			{Name: "audience", Type: promptApplicationVariableString, Required: true},
			{Name: "count", Type: promptApplicationVariableInteger, Required: true},
			{Name: "ratio", Type: promptApplicationVariableNumber, Required: true},
			{Name: "enabled", Type: promptApplicationVariableBoolean, Required: true},
			{Name: "tags", Type: promptApplicationVariableStringList, Required: true},
			{Name: "literal", Type: promptApplicationVariableString, Required: true},
		},
		OutputContract: PromptApplicationOutputContract{Kind: promptApplicationOutputText, MaxBytes: 1024},
	}
	input := map[string]any{
		"audience": "developers", "count": json.Number("42"), "ratio": json.Number("1.2500"),
		"enabled": true, "tags": []any{"stable", "strict"}, "literal": "{{ audience }}",
	}

	first := RenderPromptApplicationTemplate(source, input)
	second := RenderPromptApplicationTemplate(source, map[string]any{
		"literal": "{{ audience }}", "tags": []string{"stable", "strict"}, "enabled": true,
		"ratio": json.Number("1.25"), "count": int64(42), "audience": "developers",
	})
	if first.FailureCode != "" || second.FailureCode != "" {
		t.Fatalf("valid rendering failed: first=%#v second=%#v", first, second)
	}
	if first.RenderedDigest != second.RenderedDigest || !workflowRAGDigestPattern.MatchString(first.RenderedDigest) {
		t.Fatalf("canonical inputs must produce one stable digest: %q %q", first.RenderedDigest, second.RenderedDigest)
	}
	want := `count=42; ratio=1.25; enabled=true; tags=["stable","strict"]; literal={{ audience }}`
	if first.Messages[1].Content != want {
		t.Fatalf("renderer changed canonical output or recursively interpreted a value: %q", first.Messages[1].Content)
	}
}

func TestPromptApplicationTemplateValidationRejectsSyntaxVariablesSecretsAndBudgets(t *testing.T) {
	tests := []struct {
		name       string
		source     PromptApplicationTemplateSource
		wantedCode string
	}{
		{
			name:       "expression syntax",
			source:     promptApplicationTextTemplate("Hello {{ user.name }}", nil),
			wantedCode: PromptApplicationTemplateFailureSyntaxInvalid,
		},
		{
			name:       "undeclared variable",
			source:     promptApplicationTextTemplate("Hello {{ name }}", nil),
			wantedCode: PromptApplicationTemplateFailureVariableInvalid,
		},
		{
			name: "duplicate variable",
			source: promptApplicationTextTemplate("Hello {{ name }}", []PromptApplicationTemplateVariable{
				{Name: "name", Type: promptApplicationVariableString},
				{Name: "name", Type: promptApplicationVariableString},
			}),
			wantedCode: PromptApplicationTemplateFailureVariableInvalid,
		},
		{
			name:       "unused variable",
			source:     promptApplicationTextTemplate("Static", []PromptApplicationTemplateVariable{{Name: "unused", Type: promptApplicationVariableString}}),
			wantedCode: PromptApplicationTemplateFailureVariableInvalid,
		},
		{
			name:       "secret default",
			source:     promptApplicationTextTemplate("Value {{ value }}", []PromptApplicationTemplateVariable{{Name: "value", Type: promptApplicationVariableString, DefaultValue: "Authorization: Bearer forbidden"}}),
			wantedCode: PromptApplicationTemplateFailureSecretForbidden,
		},
		{
			name:       "required variable with default",
			source:     promptApplicationTextTemplate("Value {{ value }}", []PromptApplicationTemplateVariable{{Name: "value", Type: promptApplicationVariableString, Required: true, DefaultValue: "fallback"}}),
			wantedCode: PromptApplicationTemplateFailureVariableInvalid,
		},
		{
			name:       "message over budget",
			source:     promptApplicationTextTemplate(strings.Repeat("a", promptApplicationTemplateMaximumMessageBytes+1), nil),
			wantedCode: PromptApplicationTemplateFailurePayloadInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			findings := ValidatePromptApplicationTemplateSource(test.source)
			if !promptApplicationFindingsContain(findings, test.wantedCode) {
				t.Fatalf("wanted finding %q, got %#v", test.wantedCode, findings)
			}
		})
	}
}

func TestPromptApplicationTemplateRendererFailsClosedForInputBoundary(t *testing.T) {
	source := promptApplicationTextTemplate("Required={{ required }} Optional={{ optional }}", []PromptApplicationTemplateVariable{
		{Name: "required", Type: promptApplicationVariableInteger, Required: true},
		{Name: "optional", Type: promptApplicationVariableString, DefaultValue: "safe"},
	})

	valid := RenderPromptApplicationTemplate(source, map[string]any{"required": 7})
	if valid.FailureCode != "" || valid.Messages[0].Content != "Required=7 Optional=safe" {
		t.Fatalf("safe default did not render deterministically: %#v", valid)
	}
	for name, input := range map[string]map[string]any{
		"missing required": {},
		"extra variable":   {"required": 7, "extra": "forbidden"},
		"wrong type":       {"required": 7.5},
		"secret input":     {"required": 7, "optional": "dsn=postgres://secret"},
	} {
		t.Run(name, func(t *testing.T) {
			result := RenderPromptApplicationTemplate(source, input)
			if result.FailureCode == "" || len(result.Messages) != 0 || result.RenderedDigest != "" {
				t.Fatalf("invalid input did not fail closed: %#v", result)
			}
		})
	}
}

func TestPromptApplicationTemplateRendererEnforcesFinalRenderedBudget(t *testing.T) {
	placeholder := strings.Repeat("{{ value }}", 9)
	source := promptApplicationTextTemplate(placeholder, []PromptApplicationTemplateVariable{{Name: "value", Type: promptApplicationVariableString, Required: true}})
	result := RenderPromptApplicationTemplate(source, map[string]any{"value": strings.Repeat("x", promptApplicationTemplateMaximumVariableBytes)})
	if result.FailureCode != PromptApplicationTemplateFailureVariableInvalid || !promptApplicationFindingsContain(result.Findings, PromptApplicationTemplateFailureVariableInvalid) {
		t.Fatalf("rendered output over budget did not fail closed: %#v", result)
	}
}

func TestPromptApplicationOutputContractValidatesTextAndControlledJSONObject(t *testing.T) {
	schema := &PromptApplicationJSONSchema{
		Type: "object",
		Properties: map[string]PromptApplicationJSONSchema{
			"answer": {Type: "string"},
			"count":  {Type: "integer"},
			"active": {Type: "boolean"},
			"tags":   {Type: "array", Items: &PromptApplicationJSONSchema{Type: "string"}},
			"meta": {
				Type:       "object",
				Properties: map[string]PromptApplicationJSONSchema{"score": {Type: "number"}},
				Required:   []string{"score"},
			},
		},
		Required: []string{"answer", "count", "active", "tags"},
	}
	contract := PromptApplicationOutputContract{Kind: promptApplicationOutputJSONObject, MaxBytes: 4096, JSONSchema: schema}
	valid := `{"answer":"ok","count":2,"active":true,"tags":["a","b"],"meta":{"score":0.5}}`
	if failure := ValidatePromptApplicationOutput(contract, valid); failure != "" {
		t.Fatalf("valid JSON object failed output contract: %s", failure)
	}
	invalid := []string{
		`{"answer":"ok","count":2.5,"active":true,"tags":[]}`,
		`{"answer":"ok","count":2,"active":true,"tags":[],"extra":true}`,
		`{"answer":"ok","count":2,"active":true}`,
		`[]`,
		valid + ` {}`,
	}
	for _, output := range invalid {
		if failure := ValidatePromptApplicationOutput(contract, output); failure != PromptApplicationInvocationFailureOutputContract {
			t.Fatalf("invalid output did not fail contract: output=%q failure=%q", output, failure)
		}
	}

	textContract := PromptApplicationOutputContract{Kind: promptApplicationOutputText, MaxBytes: 8}
	if failure := ValidatePromptApplicationOutput(textContract, "advisory"); failure != "" {
		t.Fatalf("valid text output failed: %s", failure)
	}
	if failure := ValidatePromptApplicationOutput(textContract, ""); failure != PromptApplicationInvocationFailureOutputContract {
		t.Fatalf("empty text output must fail when allow_empty is false: %s", failure)
	}
	if failure := ValidatePromptApplicationOutput(textContract, "too-large"); failure != PromptApplicationInvocationFailureOutputContract {
		t.Fatalf("oversized text output must fail: %s", failure)
	}
}

func TestPromptApplicationOutputContractRejectsUnsupportedSchema(t *testing.T) {
	unknownRequired := PromptApplicationTemplateSource{
		Messages:  []PromptApplicationTemplateMessage{{Role: "user", Content: "Static"}},
		Variables: []PromptApplicationTemplateVariable{},
		OutputContract: PromptApplicationOutputContract{Kind: promptApplicationOutputJSONObject, JSONSchema: &PromptApplicationJSONSchema{
			Type: "object", Properties: map[string]PromptApplicationJSONSchema{"answer": {Type: "string"}}, Required: []string{"missing"},
		}},
	}
	if findings := ValidatePromptApplicationTemplateSource(unknownRequired); !promptApplicationFindingsContain(findings, PromptApplicationTemplateFailureOutputContractInvalid) {
		t.Fatalf("schema with unknown required property was accepted: %#v", findings)
	}

	unsupported := unknownRequired
	unsupported.OutputContract.JSONSchema = &PromptApplicationJSONSchema{Type: "null"}
	if findings := ValidatePromptApplicationTemplateSource(unsupported); !promptApplicationFindingsContain(findings, PromptApplicationTemplateFailureOutputContractInvalid) {
		t.Fatalf("unsupported JSON schema type was accepted: %#v", findings)
	}

	arrayRoot := unknownRequired
	arrayRoot.OutputContract.JSONSchema = &PromptApplicationJSONSchema{Type: "array", Items: &PromptApplicationJSONSchema{Type: "string"}}
	if findings := ValidatePromptApplicationTemplateSource(arrayRoot); !promptApplicationFindingsContain(findings, PromptApplicationTemplateFailureOutputContractInvalid) {
		t.Fatalf("json_object contract accepted an array root schema: %#v", findings)
	}
}

func promptApplicationTextTemplate(content string, variables []PromptApplicationTemplateVariable) PromptApplicationTemplateSource {
	if variables == nil {
		variables = []PromptApplicationTemplateVariable{}
	}
	return PromptApplicationTemplateSource{
		Messages:       []PromptApplicationTemplateMessage{{Role: "user", Content: content}},
		Variables:      variables,
		OutputContract: PromptApplicationOutputContract{Kind: promptApplicationOutputText, MaxBytes: 1024},
	}
}

func promptApplicationFindingsContain(findings []PromptApplicationTemplateFinding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}
