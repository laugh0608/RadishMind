package httpapi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestCompileAgentCopilotProfileSourceNormalizesDeterministically(t *testing.T) {
	unordered := validAgentCopilotProfileSourceFixture()
	unordered.ProfileName = "  RadishFlow diagnostics advisor  "
	unordered.Description = "  Review diagnostics and candidate actions.  "
	unordered.AllowedTasks = []string{"suggest_flowsheet_edits", "explain_diagnostics", "suggest_flowsheet_edits"}
	unordered.DefaultLocale = " ZH-cn "
	unordered.AllowedLocales = []string{"zh-cn", "en-us", "ZH-CN"}
	unordered.ContextPolicy.AllowedFields = []string{"diagnostics", "selected_unit_ids", "diagnostics"}
	unordered.ArtifactPolicy.AllowedKinds = []string{"text", "json", "text"}
	unordered.ArtifactPolicy.AllowedRoles = []string{"supporting", "primary", "supporting"}
	unordered.ResponsePolicy.AllowedActionKinds = []string{"read_only_check", "candidate_edit", "read_only_check"}
	unordered.RiskPolicy.ConfirmationActionKinds = []string{"ghost_completion", "candidate_operation", "candidate_edit"}

	canonical := validAgentCopilotProfileSourceFixture()
	canonical.Description = "Review diagnostics and candidate actions."
	first, findings := CompileAgentCopilotProfileSource(unordered)
	if len(findings) != 0 {
		t.Fatalf("compile unordered profile: %#v", findings)
	}
	second, findings := CompileAgentCopilotProfileSource(canonical)
	if len(findings) != 0 {
		t.Fatalf("compile canonical profile: %#v", findings)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("equivalent profiles must normalize identically:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if !reflect.DeepEqual(first.Source.AllowedTasks, []string{"explain_diagnostics", "suggest_flowsheet_edits"}) ||
		!reflect.DeepEqual(first.Source.AllowedLocales, []string{"en-US", "zh-CN"}) ||
		!reflect.DeepEqual(first.Source.ContextPolicy.AllowedFields, []string{"selected_unit_ids", "diagnostics"}) ||
		!reflect.DeepEqual(first.Source.ArtifactPolicy.AllowedKinds, []string{"json", "text"}) ||
		!reflect.DeepEqual(first.Source.ArtifactPolicy.AllowedRoles, []string{"primary", "supporting"}) ||
		!reflect.DeepEqual(first.Source.ResponsePolicy.AllowedActionKinds, []string{"candidate_edit", "read_only_check"}) ||
		!reflect.DeepEqual(first.Source.RiskPolicy.ConfirmationActionKinds, []string{"candidate_edit", "candidate_operation", "ghost_completion"}) {
		t.Fatalf("canonical order mismatch: %#v", first.Source)
	}
	for name, digest := range map[string]string{
		"profile": first.ProfileDigest, "policy": first.PolicyDigest, "allowed tasks": first.AllowedTasksDigest,
	} {
		if !workflowRAGDigestPattern.MatchString(digest) {
			t.Fatalf("%s digest is unstable or malformed: %q", name, digest)
		}
	}

	descriptionOnly := canonical
	descriptionOnly.Description = "A different display description."
	third, findings := CompileAgentCopilotProfileSource(descriptionOnly)
	if len(findings) != 0 {
		t.Fatalf("compile description-only change: %#v", findings)
	}
	if third.ProfileDigest == first.ProfileDigest {
		t.Fatal("profile digest must cover descriptive source")
	}
	if third.PolicyDigest != first.PolicyDigest || third.AllowedTasksDigest != first.AllowedTasksDigest {
		t.Fatal("descriptive source must not drift policy or allowed-task digests")
	}
}

func TestCompileAgentCopilotProfileSourceRejectsProjectTaskAndContextDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*AgentCopilotProfileSource)
		code   string
		field  string
	}{
		{name: "unknown project", code: AgentCopilotProfileFailureProjectTaskInvalid, field: "allowed_tasks", mutate: func(source *AgentCopilotProfileSource) {
			source.Project = "unknown"
		}},
		{name: "cross project task", code: AgentCopilotProfileFailureProjectTaskInvalid, field: "allowed_tasks", mutate: func(source *AgentCopilotProfileSource) {
			source.AllowedTasks = []string{"answer_docs_question"}
		}},
		{name: "cross project context", code: AgentCopilotProfileFailurePolicyInvalid, field: "context_policy.allowed_fields", mutate: func(source *AgentCopilotProfileSource) {
			source.ContextPolicy.AllowedFields = []string{"current_app"}
		}},
		{name: "ghost task missing required context", code: AgentCopilotProfileFailurePolicyInvalid, field: "context_policy.allowed_fields", mutate: func(source *AgentCopilotProfileSource) {
			source.AllowedTasks = []string{"suggest_ghost_completion"}
			source.ContextPolicy.AllowedFields = []string{"document_revision", "selected_unit_ids", "legal_candidate_completions"}
		}},
		{name: "unknown artifact kind", code: AgentCopilotProfileFailurePolicyInvalid, field: "artifact_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ArtifactPolicy.AllowedKinds = []string{"binary"}
		}},
		{name: "unknown artifact role", code: AgentCopilotProfileFailurePolicyInvalid, field: "artifact_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ArtifactPolicy.AllowedRoles = []string{"system"}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := validAgentCopilotProfileSourceFixture()
			test.mutate(&source)
			_, findings := CompileAgentCopilotProfileSource(source)
			requireAgentCopilotFinding(t, findings, test.code, test.field)
		})
	}
}

func TestCompileAgentCopilotProfileSourceEnforcesBudgetsAndEntryCounts(t *testing.T) {
	boundary := validAgentCopilotProfileSourceFixture()
	boundary.ProfileName = strings.Repeat("萝", agentCopilotMaximumProfileNameCharacters)
	boundary.Description = strings.Repeat("界", agentCopilotMaximumDescriptionCharacters)
	boundary.ContextPolicy.MaxBytes = agentCopilotMaximumContextBytes
	boundary.ArtifactPolicy.MaxCount = agentCopilotMaximumArtifacts
	boundary.ArtifactPolicy.MaxItemBytes = agentCopilotMaximumArtifactItemBytes
	boundary.ArtifactPolicy.MaxTotalBytes = agentCopilotMaximumArtifactTotalBytes
	boundary.ResponsePolicy.MaxAnswers = agentCopilotMaximumResponseItems
	boundary.ResponsePolicy.MaxIssues = agentCopilotMaximumResponseItems
	boundary.ResponsePolicy.MaxActions = agentCopilotMaximumResponseItems
	boundary.ResponsePolicy.MaxCitations = agentCopilotMaximumResponseItems
	boundary.ResponsePolicy.MaxVisibleTextBytes = agentCopilotMaximumVisibleResponseTextByte
	if _, findings := CompileAgentCopilotProfileSource(boundary); len(findings) != 0 {
		t.Fatalf("exact boundaries must remain valid: %#v", findings)
	}

	tests := []struct {
		name   string
		mutate func(*AgentCopilotProfileSource)
		code   string
		field  string
	}{
		{name: "profile name characters", code: AgentCopilotProfileFailurePayloadInvalid, field: "profile_name", mutate: func(source *AgentCopilotProfileSource) {
			source.ProfileName = strings.Repeat("萝", agentCopilotMaximumProfileNameCharacters+1)
		}},
		{name: "description characters", code: AgentCopilotProfileFailurePayloadInvalid, field: "description", mutate: func(source *AgentCopilotProfileSource) {
			source.Description = strings.Repeat("界", agentCopilotMaximumDescriptionCharacters+1)
		}},
		{name: "allowed task raw entries", code: AgentCopilotProfileFailureProjectTaskInvalid, field: "allowed_tasks", mutate: func(source *AgentCopilotProfileSource) {
			source.AllowedTasks = make([]string, agentCopilotMaximumAllowedTasks+1)
			for index := range source.AllowedTasks {
				source.AllowedTasks[index] = "explain_diagnostics"
			}
		}},
		{name: "locale entries", code: AgentCopilotProfileFailurePolicyInvalid, field: "allowed_locales", mutate: func(source *AgentCopilotProfileSource) {
			source.AllowedLocales = []string{"en", "fr", "de", "es", "it", "pt", "ja", "ko", "zh"}
		}},
		{name: "locale bytes", code: AgentCopilotProfileFailurePolicyInvalid, field: "allowed_locales", mutate: func(source *AgentCopilotProfileSource) {
			source.AllowedLocales = []string{"en-abcdefgh-abcdefgh-abcdefgh-abcdefgh"}
			source.DefaultLocale = source.AllowedLocales[0]
		}},
		{name: "context bytes", code: AgentCopilotProfileFailurePolicyInvalid, field: "context_policy.max_bytes", mutate: func(source *AgentCopilotProfileSource) {
			source.ContextPolicy.MaxBytes = agentCopilotMaximumContextBytes + 1
		}},
		{name: "artifact count", code: AgentCopilotProfileFailurePolicyInvalid, field: "artifact_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ArtifactPolicy.MaxCount = agentCopilotMaximumArtifacts + 1
		}},
		{name: "artifact item bytes", code: AgentCopilotProfileFailurePolicyInvalid, field: "artifact_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ArtifactPolicy.MaxItemBytes = agentCopilotMaximumArtifactItemBytes + 1
		}},
		{name: "artifact total bytes", code: AgentCopilotProfileFailurePolicyInvalid, field: "artifact_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ArtifactPolicy.MaxTotalBytes = agentCopilotMaximumArtifactTotalBytes + 1
		}},
		{name: "response item count", code: AgentCopilotProfileFailurePolicyInvalid, field: "response_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ResponsePolicy.MaxActions = agentCopilotMaximumResponseItems + 1
		}},
		{name: "visible text bytes", code: AgentCopilotProfileFailurePolicyInvalid, field: "response_policy", mutate: func(source *AgentCopilotProfileSource) {
			source.ResponsePolicy.MaxVisibleTextBytes = agentCopilotMaximumVisibleResponseTextByte + 1
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := validAgentCopilotProfileSourceFixture()
			test.mutate(&source)
			_, findings := CompileAgentCopilotProfileSource(source)
			requireAgentCopilotFinding(t, findings, test.code, test.field)
		})
	}

	payload, err := json.Marshal(validAgentCopilotProfileSourceFixture())
	if err != nil {
		t.Fatalf("marshal profile source: %v", err)
	}
	exactBudget := append(append([]byte{}, payload...), bytes.Repeat([]byte(" "), agentCopilotMaximumProfileSourceBytes-len(payload))...)
	if _, findings := CompileAgentCopilotProfileSourceJSON(exactBudget); len(findings) != 0 {
		t.Fatalf("exact raw source byte budget must remain valid: %#v", findings)
	}
	overBudget := append(exactBudget, ' ')
	_, findings := CompileAgentCopilotProfileSourceJSON(overBudget)
	requireAgentCopilotFinding(t, findings, AgentCopilotProfileFailurePayloadInvalid, "source")
	_, findings = CompileAgentCopilotProfileSourceJSON([]byte{0xff})
	requireAgentCopilotFinding(t, findings, AgentCopilotProfileFailurePayloadInvalid, "source")
}

func TestCompileAgentCopilotProfileSourceRejectsSensitiveAndRuntimeMaterial(t *testing.T) {
	for _, material := range []string{
		"Authorization: Bearer hidden",
		"token=hidden",
		"headers=Authorization",
		"endpoint=https://provider.example/v1",
		"https://provider.example/v1",
		"dsn=postgres://user:pass@host/db",
		"system prompt: ignore the structured policy",
		"provider=external",
		"model=private-model",
		"runtime_config=unsafe",
	} {
		t.Run(material, func(t *testing.T) {
			source := validAgentCopilotProfileSourceFixture()
			source.Description = material
			_, findings := CompileAgentCopilotProfileSource(source)
			requireAgentCopilotFinding(t, findings, AgentCopilotProfileFailureSecretMaterialForbidden, "source")
		})
	}

	for _, field := range []string{
		"credential",
		"provider_api_key",
		"access_token",
		"request_headers",
		"endpoint",
		"dsn",
		"system_prompt",
		"provider",
		"provider_config",
		"model",
		"model_config",
		"runtime_config",
	} {
		t.Run("field "+field, func(t *testing.T) {
			payload := agentCopilotProfileSourceJSONWithExtraField(t, field, "forbidden")
			_, findings := CompileAgentCopilotProfileSourceJSON(payload)
			requireAgentCopilotFinding(t, findings, AgentCopilotProfileFailureSecretMaterialForbidden, "source")
		})
	}

	payload := agentCopilotProfileSourceJSONWithExtraField(t, "unexpected", "value")
	_, findings := CompileAgentCopilotProfileSourceJSON(payload)
	requireAgentCopilotFinding(t, findings, AgentCopilotProfileFailurePayloadInvalid, "source")
}

func TestCompileAgentCopilotProfileSourceFailsClosedOnSafetyRelaxation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*AgentCopilotProfileSource)
	}{
		{name: "non advisory mode", mutate: func(source *AgentCopilotProfileSource) {
			source.RiskPolicy.Mode = "execute"
		}},
		{name: "confirmation disabled", mutate: func(source *AgentCopilotProfileSource) {
			source.RiskPolicy.RequiresConfirmationForActions = false
		}},
		{name: "confirmation kind omitted", mutate: func(source *AgentCopilotProfileSource) {
			source.RiskPolicy.ConfirmationActionKinds = []string{"candidate_edit", "candidate_operation"}
		}},
		{name: "confirmation kind duplicated", mutate: func(source *AgentCopilotProfileSource) {
			source.RiskPolicy.ConfirmationActionKinds = []string{"candidate_edit", "candidate_edit", "ghost_completion"}
		}},
		{name: "retrieval hint enabled", mutate: func(source *AgentCopilotProfileSource) {
			source.ToolHintsPolicy.AllowRetrieval = true
		}},
		{name: "tool call hint enabled", mutate: func(source *AgentCopilotProfileSource) {
			source.ToolHintsPolicy.AllowToolCalls = true
		}},
		{name: "image reasoning hint enabled", mutate: func(source *AgentCopilotProfileSource) {
			source.ToolHintsPolicy.AllowImageReasoning = true
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := validAgentCopilotProfileSourceFixture()
			test.mutate(&source)
			_, findings := CompileAgentCopilotProfileSource(source)
			if !agentCopilotFindingsContainCode(findings, AgentCopilotProfileFailurePolicyInvalid) {
				t.Fatalf("safety relaxation must fail closed: %#v", findings)
			}
		})
	}
}

func TestValidateAgentCopilotResponseConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		actions  []AgentCopilotCandidateActionSafety
		response bool
		valid    bool
	}{
		{name: "no actions", actions: []AgentCopilotCandidateActionSafety{}, response: false, valid: true},
		{name: "read only without confirmation", actions: []AgentCopilotCandidateActionSafety{{Kind: "read_only_check"}}, response: false, valid: true},
		{name: "candidate edit confirmed", actions: []AgentCopilotCandidateActionSafety{{Kind: "candidate_edit", RequiresConfirmation: true}}, response: true, valid: true},
		{name: "candidate operation action relaxed", actions: []AgentCopilotCandidateActionSafety{{Kind: "candidate_operation"}}, response: true, valid: false},
		{name: "ghost response relaxed", actions: []AgentCopilotCandidateActionSafety{{Kind: "ghost_completion", RequiresConfirmation: true}}, response: false, valid: false},
		{name: "unknown action", actions: []AgentCopilotCandidateActionSafety{{Kind: "execute"}}, response: true, valid: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ValidateAgentCopilotResponseConfirmation(test.actions, test.response); got != test.valid {
				t.Fatalf("confirmation result=%t want=%t", got, test.valid)
			}
		})
	}
}

func TestAgentCopilotCanonicalProjectionMatchesCopilotSchemas(t *testing.T) {
	requestSchema := readAgentCopilotSchema(t, "copilot-request.schema.json")
	responseSchema := readAgentCopilotSchema(t, "copilot-response.schema.json")

	requestAllOf := requestSchema["allOf"].([]any)
	tasksByProject := map[string][]string{}
	contextFieldsByProject := map[string][]string{}
	requestDefs := requestSchema["$defs"].(map[string]any)
	for _, rawRule := range requestAllOf {
		rule := rawRule.(map[string]any)
		condition := rule["if"].(map[string]any)["properties"].(map[string]any)
		projectValue, exists := condition["project"]
		if !exists {
			continue
		}
		project := projectValue.(map[string]any)["const"].(string)
		thenProperties := rule["then"].(map[string]any)["properties"].(map[string]any)
		taskPolicy, exists := thenProperties["task"]
		if !exists {
			continue
		}
		tasksByProject[project] = agentCopilotSchemaStrings(taskPolicy.(map[string]any)["enum"])
		contextRef := thenProperties["context"].(map[string]any)["$ref"].(string)
		definitionName := strings.TrimPrefix(contextRef, "#/$defs/")
		contextProperties := requestDefs[definitionName].(map[string]any)["properties"].(map[string]any)
		fields := make([]string, 0, len(contextProperties))
		for field := range contextProperties {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		contextFieldsByProject[project] = fields
	}
	if !reflect.DeepEqual(tasksByProject["radishflow"], agentCopilotCanonicalRadishFlowTasks[:]) ||
		!reflect.DeepEqual(tasksByProject["radish"], agentCopilotCanonicalRadishTasks[:]) {
		t.Fatalf("canonical task projection drifted from CopilotRequest: %#v", tasksByProject)
	}
	requireAgentCopilotSameStrings(t, contextFieldsByProject["radishflow"], agentCopilotCanonicalRadishFlowContextFields[:])
	requireAgentCopilotSameStrings(t, contextFieldsByProject["radish"], agentCopilotCanonicalRadishContextFields[:])

	artifactProperties := requestDefs["artifact"].(map[string]any)["properties"].(map[string]any)
	if !reflect.DeepEqual(agentCopilotSchemaStrings(artifactProperties["kind"].(map[string]any)["enum"]), agentCopilotCanonicalArtifactKinds[:]) ||
		!reflect.DeepEqual(agentCopilotSchemaStrings(artifactProperties["role"].(map[string]any)["enum"]), agentCopilotCanonicalArtifactRoles[:]) {
		t.Fatal("canonical artifact projection drifted from CopilotRequest")
	}
	responseDefs := responseSchema["$defs"].(map[string]any)
	actionProperties := responseDefs["candidate_action"].(map[string]any)["properties"].(map[string]any)
	if !reflect.DeepEqual(agentCopilotSchemaStrings(actionProperties["kind"].(map[string]any)["enum"]), agentCopilotCanonicalActionKinds[:]) {
		t.Fatal("canonical action projection drifted from CopilotResponse")
	}
}

func validAgentCopilotProfileSourceFixture() AgentCopilotProfileSource {
	return AgentCopilotProfileSource{
		ProfileName:    "RadishFlow diagnostics advisor",
		Description:    "Review diagnostics and candidate actions.",
		Project:        "radishflow",
		AllowedTasks:   []string{"explain_diagnostics", "suggest_flowsheet_edits"},
		DefaultLocale:  "zh-CN",
		AllowedLocales: []string{"en-US", "zh-CN"},
		ContextPolicy: AgentCopilotContextPolicy{
			AllowedFields: []string{"selected_unit_ids", "diagnostics"}, MaxBytes: 65536, RequireTaskContext: true,
		},
		ArtifactPolicy: AgentCopilotArtifactPolicy{
			AllowedKinds: []string{"json", "text"}, AllowedRoles: []string{"primary", "supporting"},
			MaxCount: 8, MaxItemBytes: 65536, MaxTotalBytes: 131072,
		},
		ResponsePolicy: AgentCopilotResponsePolicy{
			AllowedActionKinds: []string{"candidate_edit", "read_only_check"},
			MaxAnswers:         8, MaxIssues: 16, MaxActions: 8, MaxCitations: 16, MaxVisibleTextBytes: 8192,
		},
		RiskPolicy: AgentCopilotRiskPolicy{
			Mode: "advisory", RequiresConfirmationForActions: true,
			ConfirmationActionKinds: []string{"candidate_edit", "candidate_operation", "ghost_completion"},
		},
		ToolHintsPolicy: AgentCopilotToolHintsPolicy{},
	}
}

func requireAgentCopilotFinding(t *testing.T, findings []AgentCopilotProfileFinding, code, field string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Code == code && finding.Field == field {
			return
		}
	}
	t.Fatalf("missing finding code=%s field=%s in %#v", code, field, findings)
}

func agentCopilotFindingsContainCode(findings []AgentCopilotProfileFinding, code string) bool {
	for _, finding := range findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func agentCopilotProfileSourceJSONWithExtraField(t *testing.T, field string, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(validAgentCopilotProfileSourceFixture())
	if err != nil {
		t.Fatalf("marshal source: %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatalf("decode source: %v", err)
	}
	object[field] = value
	payload, err = json.Marshal(object)
	if err != nil {
		t.Fatalf("marshal source with field %s: %v", field, err)
	}
	return payload
}

func readAgentCopilotSchema(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "contracts", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var schema map[string]any
	if err := json.Unmarshal(payload, &schema); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return schema
}

func agentCopilotSchemaStrings(value any) []string {
	raw := value.([]any)
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		result = append(result, item.(string))
	}
	return result
}

func requireAgentCopilotSameStrings(t *testing.T, left, right []string) {
	t.Helper()
	left = append([]string{}, left...)
	right = append([]string{}, right...)
	sort.Strings(left)
	sort.Strings(right)
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("canonical schema projection mismatch: left=%#v right=%#v", left, right)
	}
}
