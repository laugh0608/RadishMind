package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	AgentCopilotProfileFailurePayloadInvalid          = "agent_copilot_profile_payload_invalid"
	AgentCopilotProfileFailureSecretMaterialForbidden = "agent_copilot_profile_secret_material_forbidden"
	AgentCopilotProfileFailureProjectTaskInvalid      = "agent_copilot_profile_project_task_invalid"
	AgentCopilotProfileFailurePolicyInvalid           = "agent_copilot_profile_policy_invalid"

	agentCopilotMaximumProfileNameCharacters = 80
	agentCopilotMaximumDescriptionCharacters = 512
	agentCopilotMaximumLocaleBytes           = 32
	agentCopilotMaximumAllowedTasks          = 16
	agentCopilotMaximumAllowedLocales        = 8
	agentCopilotMaximumContextFields         = 16
	agentCopilotMaximumArtifacts             = 16
	agentCopilotMaximumResponseItems         = 64
)

var (
	// These orders are the Go projection of the canonical CopilotRequest and
	// CopilotResponse schemas. The adjacent schema-alignment test prevents this
	// runtime projection from becoming an independent registry.
	agentCopilotCanonicalRadishFlowTasks = [...]string{
		"explain_diagnostics",
		"suggest_flowsheet_edits",
		"suggest_ghost_completion",
		"summarize_selection",
		"explain_control_plane_state",
		"inspect_canvas_snapshot",
	}
	agentCopilotCanonicalRadishTasks = [...]string{
		"answer_docs_question",
		"summarize_doc_or_thread",
		"suggest_forum_metadata",
		"explain_console_capability",
		"interpret_attachment",
	}
	agentCopilotCanonicalRadishFlowContextFields = [...]string{
		"document_revision",
		"selected_unit_ids",
		"selected_unit",
		"selected_stream_ids",
		"diagnostic_summary",
		"diagnostics",
		"solve_session",
		"latest_snapshot",
		"unconnected_ports",
		"missing_canonical_ports",
		"nearby_nodes",
		"cursor_context",
		"legal_candidate_completions",
		"naming_hints",
		"topology_pattern_hints",
		"control_plane_state",
	}
	agentCopilotCanonicalRadishContextFields = [...]string{
		"current_app",
		"route",
		"resource",
		"viewer",
		"attachment_refs",
		"search_scope",
	}
	agentCopilotCanonicalArtifactKinds = [...]string{
		"json",
		"markdown",
		"text",
		"image",
		"attachment_ref",
	}
	agentCopilotCanonicalArtifactRoles = [...]string{
		"primary",
		"supporting",
		"reference",
	}
	agentCopilotCanonicalActionKinds = [...]string{
		"candidate_edit",
		"candidate_operation",
		"read_only_check",
		"ghost_completion",
	}
	agentCopilotConfirmationActionKinds = [...]string{
		"candidate_edit",
		"candidate_operation",
		"ghost_completion",
	}
)

type AgentCopilotProfileFinding struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Summary string `json:"summary"`
}

type AgentCopilotCompiledProfile struct {
	Source             AgentCopilotProfileSource
	ProfileDigest      string
	PolicyDigest       string
	AllowedTasksDigest string
}

type AgentCopilotCandidateActionSafety struct {
	Kind                 string
	RequiresConfirmation bool
}

type agentCopilotDigestPolicy struct {
	Project         string                      `json:"project"`
	AllowedTasks    []string                    `json:"allowed_tasks"`
	DefaultLocale   string                      `json:"default_locale"`
	AllowedLocales  []string                    `json:"allowed_locales"`
	ContextPolicy   AgentCopilotContextPolicy   `json:"context_policy"`
	ArtifactPolicy  AgentCopilotArtifactPolicy  `json:"artifact_policy"`
	ResponsePolicy  AgentCopilotResponsePolicy  `json:"response_policy"`
	RiskPolicy      AgentCopilotRiskPolicy      `json:"risk_policy"`
	ToolHintsPolicy AgentCopilotToolHintsPolicy `json:"tool_hints_policy"`
}

type agentCopilotAllowedTasksPolicy struct {
	Project      string   `json:"project"`
	AllowedTasks []string `json:"allowed_tasks"`
}

// CompileAgentCopilotProfileSource is a deterministic, side-effect-free
// normalization and policy compilation boundary. It never calls a repository,
// Gateway, provider, retrieval service, tool executor, Session, or Run owner.
func CompileAgentCopilotProfileSource(source AgentCopilotProfileSource) (AgentCopilotCompiledProfile, []AgentCopilotProfileFinding) {
	findings := validateAgentCopilotProfileText(source)
	if sourceBytes, err := json.Marshal(source); err != nil || len(sourceBytes) > agentCopilotMaximumProfileSourceBytes {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePayloadInvalid, "source", "profile source exceeds the supported UTF-8 byte budget")
	}
	if agentCopilotProfileSourceContainsForbiddenMaterial(source) {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailureSecretMaterialForbidden, "source", "profile source contains sensitive or runtime configuration material")
	}

	normalized := AgentCopilotProfileSource{
		ProfileName:    strings.TrimSpace(source.ProfileName),
		Description:    strings.TrimSpace(source.Description),
		Project:        strings.TrimSpace(source.Project),
		DefaultLocale:  normalizeAgentCopilotLocale(source.DefaultLocale),
		ContextPolicy:  source.ContextPolicy,
		ArtifactPolicy: source.ArtifactPolicy,
		ResponsePolicy: source.ResponsePolicy,
		RiskPolicy:     source.RiskPolicy,
		ToolHintsPolicy: AgentCopilotToolHintsPolicy{
			AllowRetrieval:      source.ToolHintsPolicy.AllowRetrieval,
			AllowToolCalls:      source.ToolHintsPolicy.AllowToolCalls,
			AllowImageReasoning: source.ToolHintsPolicy.AllowImageReasoning,
		},
	}

	tasks, taskOK := normalizeAgentCopilotCanonicalValues(
		source.AllowedTasks,
		agentCopilotCanonicalTasks(normalized.Project),
		agentCopilotMaximumAllowedTasks,
		false,
	)
	if !taskOK {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailureProjectTaskInvalid, "allowed_tasks", "allowed tasks must belong to the selected canonical project")
	}
	normalized.AllowedTasks = tasks

	locales, localeOK := normalizeAgentCopilotLocales(source.AllowedLocales)
	if !localeOK || normalized.DefaultLocale == "" || !agentCopilotContainsString(locales, normalized.DefaultLocale) {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "allowed_locales", "locales must be canonical, bounded, unique after normalization, and include the default locale")
	}
	normalized.AllowedLocales = locales

	contextFields, contextOK := normalizeAgentCopilotCanonicalValues(
		source.ContextPolicy.AllowedFields,
		agentCopilotCanonicalContextFields(normalized.Project),
		agentCopilotMaximumContextFields,
		false,
	)
	if !contextOK || !agentCopilotTaskContextPolicySatisfiable(tasks, contextFields) {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "context_policy.allowed_fields", "context fields must be canonical for the selected project and satisfy every allowed task")
	}
	normalized.ContextPolicy.AllowedFields = contextFields
	if source.ContextPolicy.MaxBytes < 1 || source.ContextPolicy.MaxBytes > agentCopilotMaximumContextBytes {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "context_policy.max_bytes", "context byte budget is outside the supported range")
	}

	artifactKinds, artifactKindsOK := normalizeAgentCopilotCanonicalValues(
		source.ArtifactPolicy.AllowedKinds,
		agentCopilotCanonicalArtifactKinds[:],
		len(agentCopilotCanonicalArtifactKinds),
		false,
	)
	artifactRoles, artifactRolesOK := normalizeAgentCopilotCanonicalValues(
		source.ArtifactPolicy.AllowedRoles,
		agentCopilotCanonicalArtifactRoles[:],
		len(agentCopilotCanonicalArtifactRoles),
		false,
	)
	normalized.ArtifactPolicy.AllowedKinds = artifactKinds
	normalized.ArtifactPolicy.AllowedRoles = artifactRoles
	if !artifactKindsOK || !artifactRolesOK {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "artifact_policy", "artifact kinds and roles must use canonical values")
	}
	if source.ArtifactPolicy.MaxCount < 0 || source.ArtifactPolicy.MaxCount > agentCopilotMaximumArtifacts ||
		source.ArtifactPolicy.MaxItemBytes < 1 || source.ArtifactPolicy.MaxItemBytes > agentCopilotMaximumArtifactItemBytes ||
		source.ArtifactPolicy.MaxTotalBytes < 1 || source.ArtifactPolicy.MaxTotalBytes > agentCopilotMaximumArtifactTotalBytes ||
		source.ArtifactPolicy.MaxItemBytes > source.ArtifactPolicy.MaxTotalBytes {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "artifact_policy", "artifact count or UTF-8 byte budget is outside the supported range")
	}

	actionKinds, actionKindsOK := normalizeAgentCopilotCanonicalValues(
		source.ResponsePolicy.AllowedActionKinds,
		agentCopilotCanonicalActionKinds[:],
		len(agentCopilotCanonicalActionKinds),
		true,
	)
	normalized.ResponsePolicy.AllowedActionKinds = actionKinds
	if !actionKindsOK || !validAgentCopilotResponsePolicyBudgets(source.ResponsePolicy) {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "response_policy", "response kinds or item and text budgets are outside the supported range")
	}

	confirmationKinds, confirmationKindsOK := normalizeAgentCopilotCanonicalValues(
		source.RiskPolicy.ConfirmationActionKinds,
		agentCopilotConfirmationActionKinds[:],
		len(agentCopilotConfirmationActionKinds),
		false,
	)
	normalized.RiskPolicy.Mode = strings.TrimSpace(source.RiskPolicy.Mode)
	normalized.RiskPolicy.ConfirmationActionKinds = confirmationKinds
	if normalized.RiskPolicy.Mode != "advisory" || !source.RiskPolicy.RequiresConfirmationForActions ||
		!confirmationKindsOK || len(confirmationKinds) != len(agentCopilotConfirmationActionKinds) {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "risk_policy", "advisory mode and confirmation for every candidate action are mandatory")
	}
	if source.ToolHintsPolicy.AllowRetrieval || source.ToolHintsPolicy.AllowToolCalls || source.ToolHintsPolicy.AllowImageReasoning {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePolicyInvalid, "tool_hints_policy", "all canonical tool hints must remain false")
	}

	if normalizedBytes, err := json.Marshal(normalized); err != nil || len(normalizedBytes) > agentCopilotMaximumProfileSourceBytes {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePayloadInvalid, "source", "normalized profile source exceeds the supported UTF-8 byte budget")
	}
	if len(findings) != 0 {
		return AgentCopilotCompiledProfile{}, findings
	}

	profileDigest, profileErr := agentCopilotStableDigest(normalized)
	policyDigest, policyErr := agentCopilotStableDigest(agentCopilotDigestPolicy{
		Project: normalized.Project, AllowedTasks: normalized.AllowedTasks,
		DefaultLocale: normalized.DefaultLocale, AllowedLocales: normalized.AllowedLocales,
		ContextPolicy: normalized.ContextPolicy, ArtifactPolicy: normalized.ArtifactPolicy,
		ResponsePolicy: normalized.ResponsePolicy, RiskPolicy: normalized.RiskPolicy,
		ToolHintsPolicy: normalized.ToolHintsPolicy,
	})
	allowedTasksDigest, tasksErr := agentCopilotStableDigest(agentCopilotAllowedTasksPolicy{
		Project: normalized.Project, AllowedTasks: normalized.AllowedTasks,
	})
	if profileErr != nil || policyErr != nil || tasksErr != nil {
		return AgentCopilotCompiledProfile{}, []AgentCopilotProfileFinding{{
			Code: AgentCopilotProfileFailurePayloadInvalid, Field: "source", Summary: "profile source could not be encoded canonically",
		}}
	}
	return AgentCopilotCompiledProfile{
		Source: normalized, ProfileDigest: profileDigest, PolicyDigest: policyDigest, AllowedTasksDigest: allowedTasksDigest,
	}, nil
}

// CompileAgentCopilotProfileSourceJSON gives the future owner/API boundary a
// strict raw-source entry point without registering any route or runtime.
func CompileAgentCopilotProfileSourceJSON(payload []byte) (AgentCopilotCompiledProfile, []AgentCopilotProfileFinding) {
	if len(payload) == 0 || len(payload) > agentCopilotMaximumProfileSourceBytes || !utf8.Valid(payload) {
		return AgentCopilotCompiledProfile{}, []AgentCopilotProfileFinding{{
			Code: AgentCopilotProfileFailurePayloadInvalid, Field: "source", Summary: "profile source JSON is empty, invalid UTF-8, or exceeds the supported byte budget",
		}}
	}
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return AgentCopilotCompiledProfile{}, []AgentCopilotProfileFinding{{
			Code: AgentCopilotProfileFailurePayloadInvalid, Field: "source", Summary: "profile source JSON is invalid",
		}}
	}
	if agentCopilotRawSourceContainsForbiddenField(raw) {
		return AgentCopilotCompiledProfile{}, []AgentCopilotProfileFinding{{
			Code: AgentCopilotProfileFailureSecretMaterialForbidden, Field: "source", Summary: "profile source contains a forbidden sensitive or runtime configuration field",
		}}
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var source AgentCopilotProfileSource
	if err := decoder.Decode(&source); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return AgentCopilotCompiledProfile{}, []AgentCopilotProfileFinding{{
			Code: AgentCopilotProfileFailurePayloadInvalid, Field: "source", Summary: "profile source must match the strict structured contract",
		}}
	}
	return CompileAgentCopilotProfileSource(source)
}

// ValidateAgentCopilotResponseConfirmation applies the Profile confirmation
// invariant to an already canonical CopilotResponse projection. It does not
// execute, mutate, or otherwise authorize any candidate action.
func ValidateAgentCopilotResponseConfirmation(actions []AgentCopilotCandidateActionSafety, responseRequiresConfirmation bool) bool {
	containsConfirmationAction := false
	for _, action := range actions {
		if !agentCopilotContainsString(agentCopilotCanonicalActionKinds[:], action.Kind) {
			return false
		}
		if agentCopilotContainsString(agentCopilotConfirmationActionKinds[:], action.Kind) {
			containsConfirmationAction = true
			if !action.RequiresConfirmation {
				return false
			}
		}
	}
	return !containsConfirmationAction || responseRequiresConfirmation
}

func validateAgentCopilotProfileText(source AgentCopilotProfileSource) []AgentCopilotProfileFinding {
	findings := []AgentCopilotProfileFinding{}
	name := strings.TrimSpace(source.ProfileName)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > agentCopilotMaximumProfileNameCharacters {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePayloadInvalid, "profile_name", "profile name must contain 2 to 80 valid UTF-8 characters")
	}
	if !utf8.ValidString(source.Description) || utf8.RuneCountInString(source.Description) > agentCopilotMaximumDescriptionCharacters {
		findings = appendAgentCopilotProfileFinding(findings, AgentCopilotProfileFailurePayloadInvalid, "description", "description exceeds the supported character budget")
	}
	return findings
}

func normalizeAgentCopilotCanonicalValues(values, canonical []string, maximum int, allowEmpty bool) ([]string, bool) {
	if len(values) > maximum || len(canonical) == 0 && len(values) != 0 {
		return []string{}, false
	}
	selected := make(map[string]bool, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if !agentCopilotContainsString(canonical, value) {
			return []string{}, false
		}
		selected[value] = true
	}
	normalized := make([]string, 0, len(selected))
	for _, value := range canonical {
		if selected[value] {
			normalized = append(normalized, value)
		}
	}
	return normalized, allowEmpty || len(normalized) != 0
}

func normalizeAgentCopilotLocales(values []string) ([]string, bool) {
	if len(values) == 0 || len(values) > agentCopilotMaximumAllowedLocales {
		return []string{}, false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		normalized := normalizeAgentCopilotLocale(value)
		if normalized == "" {
			return []string{}, false
		}
		seen[normalized] = true
	}
	locales := make([]string, 0, len(seen))
	for locale := range seen {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	return locales, true
}

func normalizeAgentCopilotLocale(value string) string {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || len([]byte(value)) > agentCopilotMaximumLocaleBytes || !agentCopilotLocalePattern.MatchString(value) {
		return ""
	}
	parts := strings.Split(value, "-")
	parts[0] = strings.ToLower(parts[0])
	for index := 1; index < len(parts); index++ {
		switch {
		case len(parts[index]) == 2 && allAgentCopilotLetters(parts[index]):
			parts[index] = strings.ToUpper(parts[index])
		case len(parts[index]) == 4 && allAgentCopilotLetters(parts[index]):
			runes := []rune(strings.ToLower(parts[index]))
			runes[0] = unicode.ToUpper(runes[0])
			parts[index] = string(runes)
		default:
			parts[index] = strings.ToLower(parts[index])
		}
	}
	return strings.Join(parts, "-")
}

func allAgentCopilotLetters(value string) bool {
	for _, character := range value {
		if !unicode.IsLetter(character) {
			return false
		}
	}
	return value != ""
}

func validAgentCopilotResponsePolicyBudgets(policy AgentCopilotResponsePolicy) bool {
	return policy.MaxAnswers >= 0 && policy.MaxAnswers <= agentCopilotMaximumResponseItems &&
		policy.MaxIssues >= 0 && policy.MaxIssues <= agentCopilotMaximumResponseItems &&
		policy.MaxActions >= 0 && policy.MaxActions <= agentCopilotMaximumResponseItems &&
		policy.MaxCitations >= 0 && policy.MaxCitations <= agentCopilotMaximumResponseItems &&
		policy.MaxVisibleTextBytes >= 1 && policy.MaxVisibleTextBytes <= agentCopilotMaximumVisibleResponseTextByte
}

func agentCopilotTaskContextPolicySatisfiable(tasks, fields []string) bool {
	if !agentCopilotContainsString(tasks, "suggest_ghost_completion") {
		return true
	}
	for _, required := range []string{"document_revision", "selected_unit_ids", "legal_candidate_completions"} {
		if !agentCopilotContainsString(fields, required) {
			return false
		}
	}
	return agentCopilotContainsString(fields, "unconnected_ports") || agentCopilotContainsString(fields, "missing_canonical_ports")
}

func agentCopilotCanonicalTasks(project string) []string {
	switch project {
	case "radishflow":
		return agentCopilotCanonicalRadishFlowTasks[:]
	case "radish":
		return agentCopilotCanonicalRadishTasks[:]
	default:
		return nil
	}
}

func agentCopilotCanonicalContextFields(project string) []string {
	switch project {
	case "radishflow":
		return agentCopilotCanonicalRadishFlowContextFields[:]
	case "radish":
		return agentCopilotCanonicalRadishContextFields[:]
	default:
		return nil
	}
}

func agentCopilotStableDigest(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func agentCopilotProfileSourceIsCanonical(source, canonical AgentCopilotProfileSource) bool {
	sourcePayload, sourceErr := json.Marshal(source)
	canonicalPayload, canonicalErr := json.Marshal(canonical)
	return sourceErr == nil && canonicalErr == nil && bytes.Equal(sourcePayload, canonicalPayload)
}

func agentCopilotProfileSourceContainsForbiddenMaterial(source AgentCopilotProfileSource) bool {
	return agentCopilotTextContainsForbiddenProfileMaterial(source.ProfileName) ||
		agentCopilotTextContainsForbiddenProfileMaterial(source.Description)
}

func agentCopilotTextContainsForbiddenProfileMaterial(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	for _, marker := range []string{
		"authorization:",
		"bearer ",
		"api_key=",
		"api-key=",
		"x-api-key",
		"credential=",
		"credential:",
		"password=",
		"passwd=",
		"secret=",
		"token=",
		"token:",
		"header=",
		"headers=",
		"endpoint=",
		"endpoint:",
		"dsn=",
		"http://",
		"https://",
		"grpc://",
		"postgres://",
		"postgresql://",
		"mysql://",
		"mongodb://",
		"redis://",
		"system_prompt",
		"system prompt",
		"system-message",
		"provider=",
		"provider:",
		"provider_config",
		"provider config",
		"model=",
		"model:",
		"model_config",
		"model config",
		"runtime=",
		"runtime_config",
		"runtime config",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return applicationDraftStringContainsSecret(value)
}

func agentCopilotRawSourceContainsForbiddenField(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if agentCopilotForbiddenProfileFieldName(key) || agentCopilotRawSourceContainsForbiddenField(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if agentCopilotRawSourceContainsForbiddenField(child) {
				return true
			}
		}
	case string:
		return agentCopilotTextContainsForbiddenProfileMaterial(typed)
	}
	return false
}

func agentCopilotForbiddenProfileFieldName(value string) bool {
	normalized := strings.NewReplacer("-", "_", " ", "_").Replace(strings.ToLower(strings.TrimSpace(value)))
	for _, forbidden := range []string{
		"credential",
		"credentials",
		"authorization",
		"api_key",
		"cookie",
		"password",
		"passwd",
		"secret",
		"token",
		"headers",
		"header",
		"endpoint",
		"dsn",
		"system_prompt",
		"system_message",
		"system_instructions",
		"instructions",
		"prompt",
		"messages",
		"base_url",
		"api_url",
		"provider",
		"provider_config",
		"model",
		"model_config",
		"runtime",
		"runtime_config",
	} {
		if normalized == forbidden || strings.HasSuffix(normalized, "_"+forbidden) {
			return true
		}
	}
	return strings.HasPrefix(normalized, "provider_") ||
		strings.HasPrefix(normalized, "model_") ||
		strings.HasPrefix(normalized, "runtime_")
}

func appendAgentCopilotProfileFinding(findings []AgentCopilotProfileFinding, code, field, summary string) []AgentCopilotProfileFinding {
	for _, finding := range findings {
		if finding.Code == code && finding.Field == field {
			return findings
		}
	}
	return append(findings, AgentCopilotProfileFinding{Code: code, Field: field, Summary: summary})
}
