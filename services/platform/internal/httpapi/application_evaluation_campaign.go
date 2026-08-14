package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	applicationEvaluationPlanSchemaVersion                  = "application_evaluation_plan.v1"
	applicationEvaluationPlanVersionSchemaVersion           = "application_evaluation_plan_version.v1"
	applicationEvaluationCampaignSchemaVersion              = "application_evaluation_campaign.v1"
	applicationEvaluationStructuredPlanSchemaVersion        = "application_evaluation_plan.v2"
	applicationEvaluationStructuredPlanVersionSchemaVersion = "application_evaluation_plan_version.v2"
	applicationEvaluationStructuredCampaignSchemaVersion    = "application_evaluation_campaign.v2"

	applicationEvaluationEnvironmentDevelopment = "development"
	applicationEvaluationEnvironmentTest        = "test"

	applicationEvaluationPlanStateActive   = "active"
	applicationEvaluationPlanStateArchived = "archived"

	applicationEvaluationCampaignStatePending     = "pending"
	applicationEvaluationCampaignStateRunning     = "running"
	applicationEvaluationCampaignStateSucceeded   = "succeeded"
	applicationEvaluationCampaignStateFailed      = "failed"
	applicationEvaluationCampaignStateInterrupted = "interrupted"

	applicationEvaluationCampaignItemPending   = "pending"
	applicationEvaluationCampaignItemRunning   = "running"
	applicationEvaluationCampaignItemSucceeded = "succeeded"
	applicationEvaluationCampaignItemFailed    = "failed"

	applicationEvaluationMaximumItems       = 20
	applicationEvaluationDefaultListLimit   = 25
	applicationEvaluationMaximumListLimit   = 100
	applicationEvaluationMemoryPlanCapacity = 200
	applicationEvaluationMemoryRunCapacity  = 500
)

const (
	ApplicationEvaluationFailureScopeDenied          = "application_evaluation_scope_denied"
	ApplicationEvaluationFailureEnvironmentDenied    = "application_evaluation_environment_denied"
	ApplicationEvaluationFailureNotFound             = "application_evaluation_not_found"
	ApplicationEvaluationFailurePayloadInvalid       = "application_evaluation_payload_invalid"
	ApplicationEvaluationFailureSecretForbidden      = "application_evaluation_secret_material_forbidden"
	ApplicationEvaluationFailureProfileIneligible    = "application_evaluation_profile_ineligible"
	ApplicationEvaluationFailureVersionConflict      = "application_evaluation_version_conflict"
	ApplicationEvaluationFailureArchived             = "application_evaluation_archived"
	ApplicationEvaluationFailureCursorInvalid        = "application_evaluation_cursor_invalid"
	ApplicationEvaluationFailureCampaignConflict     = "application_evaluation_campaign_conflict"
	ApplicationEvaluationFailureAuthorityChanged     = "application_evaluation_authority_changed"
	ApplicationEvaluationFailureRunUnavailable       = "application_evaluation_run_unavailable"
	ApplicationEvaluationFailureQuotaConsumerInvalid = "application_evaluation_quota_consumer_invalid"
	ApplicationEvaluationFailureHandoffPartial       = "application_evaluation_handoff_partial"
	ApplicationEvaluationFailureStoreUnavailable     = "application_evaluation_store_unavailable"
	ApplicationEvaluationFailureStoreContract        = "application_evaluation_store_contract_mismatch"
	ApplicationEvaluationFailureWriteDisabled        = "application_evaluation_write_disabled"
)

var (
	applicationEvaluationPlanIDPattern     = regexp.MustCompile(`^aeplan_[a-z2-7]{16}$`)
	applicationEvaluationCampaignIDPattern = regexp.MustCompile(`^aecamp_[a-z2-7]{16}$`)
	applicationEvaluationItemKeyPattern    = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

	errApplicationEvaluationNotFound         = errors.New(ApplicationEvaluationFailureNotFound)
	errApplicationEvaluationVersionConflict  = errors.New(ApplicationEvaluationFailureVersionConflict)
	errApplicationEvaluationArchived         = errors.New(ApplicationEvaluationFailureArchived)
	errApplicationEvaluationCampaignConflict = errors.New(ApplicationEvaluationFailureCampaignConflict)
	errApplicationEvaluationStoreUnavailable = errors.New(ApplicationEvaluationFailureStoreUnavailable)
	errApplicationEvaluationStoreContract    = errors.New(ApplicationEvaluationFailureStoreContract)
)

type ApplicationEvaluationContext struct {
	RequestContext context.Context
	RequestID      string
	TenantRef      string
	WorkspaceID    string
	Environment    string
	ApplicationID  string
	ActorRef       string
	AuditRef       string
	WriteEnabled   bool
}

type ApplicationEvaluationDefinitionTarget struct {
	DefinitionID              string                      `json:"definition_id"`
	ExpectedPointerVersion    int                         `json:"expected_pointer_version"`
	ExpectedDefinitionVersion int                         `json:"expected_definition_version"`
	ExpectedDefinitionDigest  string                      `json:"expected_definition_digest"`
	InputContract             *WorkflowDefinitionContract `json:"input_contract,omitempty"`
}

type ApplicationEvaluationPlanTarget struct {
	WorkflowDefinition *ApplicationEvaluationDefinitionTarget `json:"workflow_definition"`
}

type ApplicationEvaluationDefinitionFixture struct {
	InputText       string          `json:"input_text"`
	ConditionValues map[string]bool `json:"condition_values"`
	Model           string          `json:"model"`
	Temperature     *float64        `json:"temperature"`
	Inputs          map[string]any  `json:"-"`
}

func (fixture ApplicationEvaluationDefinitionFixture) MarshalJSON() ([]byte, error) {
	if fixture.Inputs != nil {
		return json.Marshal(struct {
			Inputs map[string]any `json:"inputs"`
		}{Inputs: fixture.Inputs})
	}
	type legacyFixture struct {
		InputText       string          `json:"input_text"`
		ConditionValues map[string]bool `json:"condition_values"`
		Model           string          `json:"model"`
		Temperature     *float64        `json:"temperature"`
	}
	return json.Marshal(legacyFixture{
		InputText: fixture.InputText, ConditionValues: fixture.ConditionValues,
		Model: fixture.Model, Temperature: fixture.Temperature,
	})
}

func (fixture *ApplicationEvaluationDefinitionFixture) UnmarshalJSON(payload []byte) error {
	var fields map[string]json.RawMessage
	if err := decodeStrictApplicationEvaluationJSON(payload, &fields); err != nil {
		return err
	}
	if raw, present := fields["inputs"]; present {
		if len(fields) != 1 {
			return errApplicationEvaluationStoreContract
		}
		var inputs map[string]any
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&inputs); err != nil || inputs == nil || decoder.Decode(&struct{}{}) == nil {
			return errApplicationEvaluationStoreContract
		}
		*fixture = ApplicationEvaluationDefinitionFixture{Inputs: inputs}
		return nil
	}
	if len(fields) != 4 {
		return errApplicationEvaluationStoreContract
	}
	type legacyFixture struct {
		InputText       string          `json:"input_text"`
		ConditionValues map[string]bool `json:"condition_values"`
		Model           string          `json:"model"`
		Temperature     *float64        `json:"temperature"`
	}
	var legacy legacyFixture
	if err := decodeStrictApplicationEvaluationJSON(payload, &legacy); err != nil {
		return err
	}
	*fixture = ApplicationEvaluationDefinitionFixture{
		InputText: legacy.InputText, ConditionValues: legacy.ConditionValues,
		Model: legacy.Model, Temperature: legacy.Temperature,
	}
	return nil
}

type ApplicationEvaluationRAGFixture struct {
	Input string `json:"input"`
}

type ApplicationEvaluationPromptFixture struct {
	Variables map[string]any `json:"variables"`
}

type ApplicationEvaluationAgentFixture struct {
	Task           string                 `json:"task"`
	Locale         string                 `json:"locale"`
	ConversationID string                 `json:"conversation_id"`
	Artifacts      []AgentCopilotArtifact `json:"artifacts"`
	Context        map[string]any         `json:"context"`
}

type ApplicationEvaluationPlanItem struct {
	ItemKey                string                                  `json:"item_key"`
	Name                   string                                  `json:"name"`
	ExpectedClassification WorkflowRunComparisonClassification     `json:"expected_classification"`
	WorkflowDefinition     *ApplicationEvaluationDefinitionFixture `json:"workflow_definition"`
	ApplicationRAG         *ApplicationEvaluationRAGFixture        `json:"application_rag"`
	PromptApplication      *ApplicationEvaluationPromptFixture     `json:"prompt_application"`
	AgentCopilot           *ApplicationEvaluationAgentFixture      `json:"agent_copilot"`
}

type ApplicationEvaluationPlan struct {
	SchemaVersion     string `json:"schema_version"`
	PlanID            string `json:"plan_id"`
	RecordVersion     int    `json:"record_version"`
	LatestPlanVersion int    `json:"latest_plan_version"`
	LatestPlanDigest  string `json:"latest_plan_digest"`
	TenantRef         string `json:"tenant_ref"`
	WorkspaceID       string `json:"workspace_id"`
	Environment       string `json:"environment"`
	ApplicationID     string `json:"application_id"`
	Name              string `json:"name"`
	ExecutionProfile  string `json:"execution_profile"`
	ItemCount         int    `json:"item_count"`
	LifecycleState    string `json:"lifecycle_state"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	CreatedByActorRef string `json:"created_by_actor_ref"`
	UpdatedByActorRef string `json:"updated_by_actor_ref"`
	RequestID         string `json:"request_id"`
	AuditRef          string `json:"audit_ref"`
}

type ApplicationEvaluationPlanVersion struct {
	SchemaVersion       string                          `json:"schema_version"`
	PlanID              string                          `json:"plan_id"`
	PlanVersion         int                             `json:"plan_version"`
	PreviousPlanVersion int                             `json:"previous_plan_version"`
	PlanDigest          string                          `json:"plan_digest"`
	TenantRef           string                          `json:"tenant_ref"`
	WorkspaceID         string                          `json:"workspace_id"`
	Environment         string                          `json:"environment"`
	ApplicationID       string                          `json:"application_id"`
	Name                string                          `json:"name"`
	ExecutionProfile    string                          `json:"execution_profile"`
	Target              ApplicationEvaluationPlanTarget `json:"target"`
	Items               []ApplicationEvaluationPlanItem `json:"items"`
	CreatedAt           string                          `json:"created_at"`
	CreatedByActorRef   string                          `json:"created_by_actor_ref"`
	RequestID           string                          `json:"request_id"`
	AuditRef            string                          `json:"audit_ref"`
}

type ApplicationEvaluationCampaignAuthority struct {
	ExecutionProfile string          `json:"execution_profile"`
	AuthorityDigest  string          `json:"authority_digest"`
	Snapshot         json.RawMessage `json:"snapshot"`
}

type ApplicationEvaluationCampaignItem struct {
	ItemKey          string `json:"item_key"`
	RunID            string `json:"run_id"`
	State            string `json:"state"`
	RunSchemaVersion string `json:"run_schema_version"`
	RunProfile       string `json:"run_profile"`
	AuthorityDigest  string `json:"authority_digest"`
	FailureCode      string `json:"failure_code"`
	FailureBoundary  string `json:"failure_boundary"`
	StartedAt        string `json:"started_at"`
	CompletedAt      string `json:"completed_at"`
}

type ApplicationEvaluationHandoffRef struct {
	BaselineCampaignID  string                           `json:"baseline_campaign_id"`
	CandidateCampaignID string                           `json:"candidate_campaign_id"`
	CaseRefs            []WorkflowEvaluationSuiteCaseRef `json:"case_refs"`
	SuiteID             string                           `json:"suite_id"`
	State               string                           `json:"state"`
	AuditRef            string                           `json:"audit_ref"`
}

type ApplicationEvaluationCampaign struct {
	SchemaVersion     string                                  `json:"schema_version"`
	CampaignID        string                                  `json:"campaign_id"`
	ClientCampaignKey string                                  `json:"client_campaign_key"`
	RecordVersion     int                                     `json:"record_version"`
	TenantRef         string                                  `json:"tenant_ref"`
	WorkspaceID       string                                  `json:"workspace_id"`
	Environment       string                                  `json:"environment"`
	ApplicationID     string                                  `json:"application_id"`
	PlanID            string                                  `json:"plan_id"`
	PlanVersion       int                                     `json:"plan_version"`
	PlanDigest        string                                  `json:"plan_digest"`
	ExecutionProfile  string                                  `json:"execution_profile"`
	QuotaAPIKeyID     string                                  `json:"quota_api_key_id"`
	Authority         *ApplicationEvaluationCampaignAuthority `json:"authority"`
	State             string                                  `json:"state"`
	CurrentItemIndex  int                                     `json:"current_item_index"`
	SucceededItems    int                                     `json:"succeeded_items"`
	FailedItems       int                                     `json:"failed_items"`
	FailureCode       string                                  `json:"failure_code"`
	FailureSummary    string                                  `json:"failure_summary"`
	Items             []ApplicationEvaluationCampaignItem     `json:"items"`
	Handoff           *ApplicationEvaluationHandoffRef        `json:"handoff"`
	CreatedAt         string                                  `json:"created_at"`
	StartedAt         string                                  `json:"started_at"`
	CompletedAt       string                                  `json:"completed_at"`
	CreatedByActorRef string                                  `json:"created_by_actor_ref"`
	UpdatedByActorRef string                                  `json:"updated_by_actor_ref"`
	RequestID         string                                  `json:"request_id"`
	AuditRef          string                                  `json:"audit_ref"`
}

type ApplicationEvaluationPlanListFilter struct {
	LifecycleState  string
	BeforeUpdatedAt string
	BeforePlanID    string
	Limit           int
}

type ApplicationEvaluationPlanListPage struct {
	Plans   []ApplicationEvaluationPlan
	HasMore bool
}

type ApplicationEvaluationVersionListFilter struct {
	BeforeVersion int
	Limit         int
}

type ApplicationEvaluationVersionListPage struct {
	Versions []ApplicationEvaluationPlanVersion
	HasMore  bool
}

type ApplicationEvaluationCampaignListFilter struct {
	PlanID           string
	BeforeCreatedAt  string
	BeforeCampaignID string
	Limit            int
}

type ApplicationEvaluationCampaignListPage struct {
	Campaigns []ApplicationEvaluationCampaign
	HasMore   bool
}

type applicationEvaluationVersionConflictError struct {
	CurrentVersion int
	CurrentState   string
}

func (failure applicationEvaluationVersionConflictError) Error() string {
	return ApplicationEvaluationFailureVersionConflict
}

func (failure applicationEvaluationVersionConflictError) Is(target error) bool {
	return target == errApplicationEvaluationVersionConflict
}

func validApplicationEvaluationContext(ctx ApplicationEvaluationContext) bool {
	return ctx.RequestContext != nil && strings.TrimSpace(ctx.RequestID) != "" &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.TenantRef)) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.WorkspaceID)) &&
		validApplicationEvaluationEnvironment(ctx.Environment) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.ApplicationID)) &&
		applicationDraftIdentifierPattern.MatchString(strings.TrimSpace(ctx.ActorRef)) &&
		strings.TrimSpace(ctx.AuditRef) != ""
}

func validApplicationEvaluationEnvironment(value string) bool {
	value = strings.TrimSpace(value)
	return value == applicationEvaluationEnvironmentDevelopment || value == applicationEvaluationEnvironmentTest
}

func validApplicationEvaluationExecutionProfile(value string) bool {
	switch strings.TrimSpace(value) {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured, applicationInteractionProfileRAG,
		applicationInteractionProfilePrompt, applicationInteractionProfileAgentCopilot:
		return true
	default:
		return false
	}
}

func applicationEvaluationSchemaVersions(profile string) (planSchema, versionSchema, campaignSchema string, ok bool) {
	if profile == applicationInteractionProfileWorkflowStructured {
		return applicationEvaluationStructuredPlanSchemaVersion, applicationEvaluationStructuredPlanVersionSchemaVersion,
			applicationEvaluationStructuredCampaignSchemaVersion, true
	}
	if validApplicationEvaluationExecutionProfile(profile) {
		return applicationEvaluationPlanSchemaVersion, applicationEvaluationPlanVersionSchemaVersion,
			applicationEvaluationCampaignSchemaVersion, true
	}
	return "", "", "", false
}

func validApplicationEvaluationClassification(value WorkflowRunComparisonClassification) bool {
	switch value {
	case WorkflowRunComparisonRegression, WorkflowRunComparisonImprovement, WorkflowRunComparisonChanged,
		WorkflowRunComparisonUnchanged, WorkflowRunComparisonInconclusive:
		return true
	default:
		return false
	}
}

func normalizeApplicationEvaluationPlanDefinition(
	name string,
	profile string,
	target ApplicationEvaluationPlanTarget,
	items []ApplicationEvaluationPlanItem,
) (string, string, ApplicationEvaluationPlanTarget, []ApplicationEvaluationPlanItem, string) {
	name = strings.TrimSpace(name)
	profile = strings.TrimSpace(profile)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 2 || utf8.RuneCountInString(name) > 120 ||
		!validApplicationEvaluationExecutionProfile(profile) || len(items) < 1 || len(items) > applicationEvaluationMaximumItems {
		return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
	}
	if applicationDraftStringContainsSecret(name) {
		return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailureSecretForbidden
	}
	if profile == applicationInteractionProfileWorkflow || profile == applicationInteractionProfileWorkflowStructured {
		if target.WorkflowDefinition == nil {
			return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
		}
		definition := *target.WorkflowDefinition
		definition.DefinitionID = strings.TrimSpace(definition.DefinitionID)
		definition.ExpectedDefinitionDigest = strings.TrimSpace(definition.ExpectedDefinitionDigest)
		if !applicationDraftIdentifierPattern.MatchString(definition.DefinitionID) || definition.ExpectedPointerVersion < 1 ||
			definition.ExpectedDefinitionVersion < 1 || !workflowRAGDigestPattern.MatchString(definition.ExpectedDefinitionDigest) {
			return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
		}
		if profile == applicationInteractionProfileWorkflowStructured {
			if definition.InputContract == nil {
				return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
			}
			normalized, code, _ := normalizeWorkflowStructuredInputContract(savedWorkflowDraftContractFromDefinition(*definition.InputContract))
			if code != "" || normalized.ContractDigest != definition.InputContract.ContractDigest {
				return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
			}
			definition.InputContract = &WorkflowDefinitionContract{
				ContractID: normalized.ContractID, Fields: cloneWorkflowStructuredInputFields(normalized.Fields),
				Summary: normalized.Summary, ContractDigest: normalized.ContractDigest,
			}
		} else if definition.InputContract != nil {
			return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
		}
		target.WorkflowDefinition = &definition
	} else if target.WorkflowDefinition != nil {
		return "", "", ApplicationEvaluationPlanTarget{}, nil, ApplicationEvaluationFailurePayloadInvalid
	}

	normalizedItems := make([]ApplicationEvaluationPlanItem, 0, len(items))
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		normalized, failure := normalizeApplicationEvaluationPlanItem(profile, target, item)
		if failure != "" || seen[normalized.ItemKey] {
			if failure == "" {
				failure = ApplicationEvaluationFailurePayloadInvalid
			}
			return "", "", ApplicationEvaluationPlanTarget{}, nil, failure
		}
		seen[normalized.ItemKey] = true
		normalizedItems = append(normalizedItems, normalized)
	}
	return name, profile, target, normalizedItems, ""
}

func normalizeApplicationEvaluationPlanItem(
	profile string,
	target ApplicationEvaluationPlanTarget,
	item ApplicationEvaluationPlanItem,
) (ApplicationEvaluationPlanItem, string) {
	item.ItemKey = strings.TrimSpace(item.ItemKey)
	item.Name = strings.TrimSpace(item.Name)
	if !applicationEvaluationItemKeyPattern.MatchString(item.ItemKey) || !utf8.ValidString(item.Name) ||
		utf8.RuneCountInString(item.Name) < 1 || utf8.RuneCountInString(item.Name) > 120 ||
		!validApplicationEvaluationClassification(item.ExpectedClassification) {
		return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
	}
	if applicationDraftStringContainsSecret(item.Name) {
		return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailureSecretForbidden
	}
	fixtureCount := 0
	for _, present := range []bool{item.WorkflowDefinition != nil, item.ApplicationRAG != nil, item.PromptApplication != nil, item.AgentCopilot != nil} {
		if present {
			fixtureCount++
		}
	}
	if fixtureCount != 1 {
		return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
	}
	switch profile {
	case applicationInteractionProfileWorkflow:
		if item.WorkflowDefinition == nil || target.WorkflowDefinition == nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture := *item.WorkflowDefinition
		normalized, code, _ := normalizeWorkflowDefinitionRunRequest(WorkflowDefinitionRunRequest{
			DefinitionID: target.WorkflowDefinition.DefinitionID, ExpectedPointerVersion: target.WorkflowDefinition.ExpectedPointerVersion,
			ExpectedDefinitionVersion: target.WorkflowDefinition.ExpectedDefinitionVersion, ExpectedDefinitionDigest: target.WorkflowDefinition.ExpectedDefinitionDigest,
			InputText: fixture.InputText, ConditionValues: fixture.ConditionValues, Model: fixture.Model, Temperature: fixture.Temperature,
		})
		if code != "" {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture.InputText, fixture.ConditionValues, fixture.Model, fixture.Temperature = normalized.InputText, normalized.ConditionValues, normalized.Model, normalized.Temperature
		item.WorkflowDefinition = &fixture
	case applicationInteractionProfileWorkflowStructured:
		if item.WorkflowDefinition == nil || target.WorkflowDefinition == nil || target.WorkflowDefinition.InputContract == nil || item.WorkflowDefinition.Inputs == nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture := *item.WorkflowDefinition
		if fixture.InputText != "" || fixture.ConditionValues != nil || fixture.Model != "" || fixture.Temperature != nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		normalized, code, _ := normalizeWorkflowStructuredInputValues(savedWorkflowDraftContractFromDefinition(*target.WorkflowDefinition.InputContract), fixture.Inputs)
		if code != "" {
			if code == WorkflowRunFailureInputSecretMaterialForbidden {
				return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailureSecretForbidden
			}
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		var inputs map[string]any
		if err := json.Unmarshal(normalized.CanonicalPayload, &inputs); err != nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture.Inputs = inputs
		item.WorkflowDefinition = &fixture
	case applicationInteractionProfileRAG:
		if item.ApplicationRAG == nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture := *item.ApplicationRAG
		fixture.Input = strings.TrimSpace(fixture.Input)
		if fixture.Input == "" || len([]byte(fixture.Input)) > workflowRAGApplicationInvocationMaxBytes || !utf8Safe(fixture.Input) {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		item.ApplicationRAG = &fixture
	case applicationInteractionProfilePrompt:
		if item.PromptApplication == nil || item.PromptApplication.Variables == nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		if _, _, err := canonicalPromptApplicationInvocationInput(item.PromptApplication.Variables); err != nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
	case applicationInteractionProfileAgentCopilot:
		if item.AgentCopilot == nil {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		fixture := *item.AgentCopilot
		fixture.Task = strings.TrimSpace(fixture.Task)
		fixture.Locale = normalizeAgentCopilotLocale(fixture.Locale)
		fixture.ConversationID = strings.TrimSpace(fixture.ConversationID)
		contextPayload, contextErr := json.Marshal(fixture.Context)
		artifactPayload, artifactErr := json.Marshal(fixture.Artifacts)
		if fixture.Task == "" || fixture.Locale == "" || fixture.Context == nil ||
			(fixture.ConversationID != "" && !validPromptApplicationRef(fixture.ConversationID)) ||
			contextErr != nil || artifactErr != nil || len(contextPayload) > agentCopilotMaximumContextBytes ||
			len(artifactPayload) > agentCopilotMaximumArtifactTotalBytes+16384 || len(fixture.Artifacts) > agentCopilotMaximumArtifacts {
			return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
		}
		item.AgentCopilot = &fixture
	default:
		return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailurePayloadInvalid
	}
	if applicationEvaluationValueContainsSecret(item) {
		return ApplicationEvaluationPlanItem{}, ApplicationEvaluationFailureSecretForbidden
	}
	return item, ""
}

func applicationEvaluationValueContainsSecret(value any) bool {
	payload, err := json.Marshal(value)
	if err != nil || !utf8.Valid(payload) {
		return true
	}
	text := string(payload)
	return workflowRAGContainsForbiddenMaterial(text) || applicationDraftStringContainsSecret(text)
}

func applicationEvaluationPlanDigest(version ApplicationEvaluationPlanVersion) (string, error) {
	payload, err := json.Marshal(struct {
		PlanID           string                          `json:"plan_id"`
		PlanVersion      int                             `json:"plan_version"`
		Name             string                          `json:"name"`
		ExecutionProfile string                          `json:"execution_profile"`
		Target           ApplicationEvaluationPlanTarget `json:"target"`
		Items            []ApplicationEvaluationPlanItem `json:"items"`
	}{version.PlanID, version.PlanVersion, version.Name, version.ExecutionProfile, version.Target, version.Items})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validateApplicationEvaluationPlan(ctx ApplicationEvaluationContext, plan ApplicationEvaluationPlan) error {
	createdAt, createdErr := time.Parse(time.RFC3339Nano, plan.CreatedAt)
	updatedAt, updatedErr := time.Parse(time.RFC3339Nano, plan.UpdatedAt)
	expectedSchema, _, _, supported := applicationEvaluationSchemaVersions(plan.ExecutionProfile)
	if !validApplicationEvaluationContext(ctx) || !supported || plan.SchemaVersion != expectedSchema ||
		!applicationEvaluationPlanIDPattern.MatchString(plan.PlanID) || plan.RecordVersion < 1 || plan.LatestPlanVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(plan.LatestPlanDigest) || plan.TenantRef != ctx.TenantRef ||
		plan.WorkspaceID != ctx.WorkspaceID || plan.Environment != ctx.Environment || plan.ApplicationID != ctx.ApplicationID ||
		!validApplicationEvaluationName(plan.Name) || !validApplicationEvaluationExecutionProfile(plan.ExecutionProfile) ||
		plan.ItemCount < 1 || plan.ItemCount > applicationEvaluationMaximumItems ||
		(plan.LifecycleState != applicationEvaluationPlanStateActive && plan.LifecycleState != applicationEvaluationPlanStateArchived) ||
		createdErr != nil || updatedErr != nil || updatedAt.Before(createdAt) ||
		strings.TrimSpace(plan.CreatedByActorRef) == "" || strings.TrimSpace(plan.UpdatedByActorRef) == "" ||
		strings.TrimSpace(plan.RequestID) == "" || strings.TrimSpace(plan.AuditRef) == "" {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validateApplicationEvaluationPlanVersion(ctx ApplicationEvaluationContext, version ApplicationEvaluationPlanVersion) error {
	_, expectedSchema, _, supported := applicationEvaluationSchemaVersions(version.ExecutionProfile)
	if !validApplicationEvaluationContext(ctx) || !supported || version.SchemaVersion != expectedSchema ||
		!applicationEvaluationPlanIDPattern.MatchString(version.PlanID) || version.PlanVersion < 1 || version.PreviousPlanVersion != version.PlanVersion-1 ||
		version.TenantRef != ctx.TenantRef || version.WorkspaceID != ctx.WorkspaceID || version.Environment != ctx.Environment || version.ApplicationID != ctx.ApplicationID ||
		strings.TrimSpace(version.CreatedByActorRef) == "" || strings.TrimSpace(version.RequestID) == "" || strings.TrimSpace(version.AuditRef) == "" {
		return errApplicationEvaluationStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, version.CreatedAt); err != nil {
		return errApplicationEvaluationStoreContract
	}
	name, profile, target, items, failure := normalizeApplicationEvaluationPlanDefinition(version.Name, version.ExecutionProfile, version.Target, version.Items)
	if failure != "" || name != version.Name || profile != version.ExecutionProfile || !applicationEvaluationDefinitionsEqual(target, version.Target) ||
		!applicationEvaluationItemsEqual(items, version.Items) {
		return errApplicationEvaluationStoreContract
	}
	digest, err := applicationEvaluationPlanDigest(version)
	if err != nil || digest != version.PlanDigest {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validateApplicationEvaluationCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign) error {
	createdAt, createdErr := time.Parse(time.RFC3339Nano, campaign.CreatedAt)
	_ = createdAt
	_, _, expectedSchema, supported := applicationEvaluationSchemaVersions(campaign.ExecutionProfile)
	if !validApplicationEvaluationContext(ctx) || !supported || campaign.SchemaVersion != expectedSchema ||
		!applicationEvaluationCampaignIDPattern.MatchString(campaign.CampaignID) ||
		!applicationDraftIdentifierPattern.MatchString(campaign.ClientCampaignKey) || campaign.RecordVersion < 1 ||
		campaign.TenantRef != ctx.TenantRef || campaign.WorkspaceID != ctx.WorkspaceID || campaign.Environment != ctx.Environment || campaign.ApplicationID != ctx.ApplicationID ||
		!applicationEvaluationPlanIDPattern.MatchString(campaign.PlanID) || campaign.PlanVersion < 1 || !workflowRAGDigestPattern.MatchString(campaign.PlanDigest) ||
		!validApplicationEvaluationExecutionProfile(campaign.ExecutionProfile) || !apiKeyIDPattern.MatchString(campaign.QuotaAPIKeyID) || createdErr != nil ||
		strings.TrimSpace(campaign.CreatedByActorRef) == "" || strings.TrimSpace(campaign.UpdatedByActorRef) == "" ||
		strings.TrimSpace(campaign.RequestID) == "" || strings.TrimSpace(campaign.AuditRef) == "" ||
		len(campaign.Items) < 1 || len(campaign.Items) > applicationEvaluationMaximumItems || campaign.CurrentItemIndex < 0 || campaign.CurrentItemIndex > len(campaign.Items) ||
		campaign.SucceededItems < 0 || campaign.FailedItems < 0 || campaign.SucceededItems+campaign.FailedItems > len(campaign.Items) {
		return errApplicationEvaluationStoreContract
	}
	if !validApplicationEvaluationCampaignState(campaign.State) || !validApplicationEvaluationCampaignTiming(campaign) {
		return errApplicationEvaluationStoreContract
	}
	seen := make(map[string]bool, len(campaign.Items))
	for _, item := range campaign.Items {
		if seen[item.ItemKey] || validateApplicationEvaluationCampaignItem(campaign, item) != nil {
			return errApplicationEvaluationStoreContract
		}
		seen[item.ItemKey] = true
	}
	if campaign.Authority != nil && validateApplicationEvaluationCampaignAuthority(*campaign.Authority) != nil {
		return errApplicationEvaluationStoreContract
	}
	if campaign.Handoff != nil && validateApplicationEvaluationHandoff(*campaign.Handoff) != nil {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validateApplicationEvaluationCampaignAuthority(authority ApplicationEvaluationCampaignAuthority) error {
	if !validApplicationEvaluationExecutionProfile(authority.ExecutionProfile) || !workflowRAGDigestPattern.MatchString(authority.AuthorityDigest) || len(authority.Snapshot) == 0 {
		return errApplicationEvaluationStoreContract
	}
	var digest string
	switch authority.ExecutionProfile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured, applicationInteractionProfileRAG:
		var snapshot ApplicationInteractionAuthoritySnapshot
		if decodeStrictApplicationEvaluationJSON(authority.Snapshot, &snapshot) != nil || validateApplicationInteractionAuthority(snapshot) != nil {
			return errApplicationEvaluationStoreContract
		}
		digest = snapshot.AuthorityDigest
	case applicationInteractionProfilePrompt:
		var snapshot PromptApplicationRuntimeAuthorityV2
		if decodeStrictApplicationEvaluationJSON(authority.Snapshot, &snapshot) != nil || validatePromptApplicationRuntimeAuthorityV2(snapshot) != nil {
			return errApplicationEvaluationStoreContract
		}
		digest = snapshot.AuthorityDigest
	case applicationInteractionProfileAgentCopilot:
		var snapshot AgentCopilotRuntimeAuthorityV3
		if decodeStrictApplicationEvaluationJSON(authority.Snapshot, &snapshot) != nil || validateAgentCopilotRuntimeAuthority(snapshot) != nil {
			return errApplicationEvaluationStoreContract
		}
		digest = snapshot.AuthorityDigest
	}
	if digest != authority.AuthorityDigest {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validateApplicationEvaluationCampaignItem(campaign ApplicationEvaluationCampaign, item ApplicationEvaluationCampaignItem) error {
	if !applicationEvaluationItemKeyPattern.MatchString(item.ItemKey) || !workflowRAGRunIDPattern.MatchString(item.RunID) ||
		!validApplicationEvaluationCampaignItemState(item.State) || len(item.FailureCode) > 160 || len(item.FailureBoundary) > 160 {
		return errApplicationEvaluationStoreContract
	}
	if item.AuthorityDigest != "" && !workflowRAGDigestPattern.MatchString(item.AuthorityDigest) {
		return errApplicationEvaluationStoreContract
	}
	if item.State == applicationEvaluationCampaignItemPending {
		if item.StartedAt != "" || item.CompletedAt != "" || item.RunSchemaVersion != "" || item.RunProfile != "" || item.FailureCode != "" {
			return errApplicationEvaluationStoreContract
		}
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, item.StartedAt); err != nil {
		return errApplicationEvaluationStoreContract
	}
	if item.State == applicationEvaluationCampaignItemRunning {
		if item.CompletedAt != "" || item.RunSchemaVersion != "" || item.RunProfile != "" {
			return errApplicationEvaluationStoreContract
		}
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, item.CompletedAt); err != nil || item.RunSchemaVersion == "" || item.RunProfile == "" || item.AuthorityDigest == "" {
		return errApplicationEvaluationStoreContract
	}
	if item.State == applicationEvaluationCampaignItemSucceeded && item.FailureCode != "" || item.State == applicationEvaluationCampaignItemFailed && item.FailureCode == "" {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validateApplicationEvaluationHandoff(handoff ApplicationEvaluationHandoffRef) error {
	if !applicationEvaluationCampaignIDPattern.MatchString(handoff.BaselineCampaignID) || !applicationEvaluationCampaignIDPattern.MatchString(handoff.CandidateCampaignID) ||
		(handoff.State != "complete" && handoff.State != "partial") || strings.TrimSpace(handoff.AuditRef) == "" {
		return errApplicationEvaluationStoreContract
	}
	seen := make(map[string]bool, len(handoff.CaseRefs))
	for _, ref := range handoff.CaseRefs {
		key := ref.CaseID + "\x00" + strconv.Itoa(ref.Version)
		if !strings.HasPrefix(ref.CaseID, "eval_") || ref.Version < 1 || seen[key] {
			return errApplicationEvaluationStoreContract
		}
		seen[key] = true
	}
	if handoff.State == "complete" && (len(handoff.CaseRefs) == 0 || !strings.HasPrefix(handoff.SuiteID, "suite_")) {
		return errApplicationEvaluationStoreContract
	}
	if handoff.State == "partial" && handoff.SuiteID != "" {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func validApplicationEvaluationCampaignTiming(campaign ApplicationEvaluationCampaign) bool {
	switch campaign.State {
	case applicationEvaluationCampaignStatePending:
		return campaign.StartedAt == "" && campaign.CompletedAt == "" && campaign.Authority == nil && campaign.FailureCode == ""
	case applicationEvaluationCampaignStateRunning:
		return parseApplicationEvaluationTimestamp(campaign.StartedAt) != nil && campaign.CompletedAt == "" && campaign.Authority != nil && campaign.FailureCode == ""
	case applicationEvaluationCampaignStateSucceeded:
		return parseApplicationEvaluationTimestamp(campaign.StartedAt) != nil && parseApplicationEvaluationTimestamp(campaign.CompletedAt) != nil && campaign.Authority != nil && campaign.FailureCode == "" && campaign.SucceededItems == len(campaign.Items)
	case applicationEvaluationCampaignStateFailed, applicationEvaluationCampaignStateInterrupted:
		return parseApplicationEvaluationTimestamp(campaign.StartedAt) != nil && parseApplicationEvaluationTimestamp(campaign.CompletedAt) != nil && campaign.Authority != nil && campaign.FailureCode != ""
	default:
		return false
	}
}

func validApplicationEvaluationCampaignState(value string) bool {
	switch value {
	case applicationEvaluationCampaignStatePending, applicationEvaluationCampaignStateRunning,
		applicationEvaluationCampaignStateSucceeded, applicationEvaluationCampaignStateFailed,
		applicationEvaluationCampaignStateInterrupted:
		return true
	default:
		return false
	}
}

func validApplicationEvaluationCampaignItemState(value string) bool {
	switch value {
	case applicationEvaluationCampaignItemPending, applicationEvaluationCampaignItemRunning,
		applicationEvaluationCampaignItemSucceeded, applicationEvaluationCampaignItemFailed:
		return true
	default:
		return false
	}
}

func validApplicationEvaluationName(value string) bool {
	return strings.TrimSpace(value) == value && utf8.ValidString(value) && utf8.RuneCountInString(value) >= 2 && utf8.RuneCountInString(value) <= 120 && !applicationDraftStringContainsSecret(value)
}

func parseApplicationEvaluationTimestamp(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func decodeStrictApplicationEvaluationJSON(payload []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) == nil {
		return errApplicationEvaluationStoreContract
	}
	return nil
}

func applicationEvaluationDefinitionsEqual(left, right ApplicationEvaluationPlanTarget) bool {
	leftPayload, leftErr := json.Marshal(left)
	rightPayload, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftPayload) == string(rightPayload)
}

func applicationEvaluationItemsEqual(left, right []ApplicationEvaluationPlanItem) bool {
	leftPayload, leftErr := json.Marshal(left)
	rightPayload, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftPayload) == string(rightPayload)
}

func newApplicationEvaluationID(prefix string) (string, error) {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return prefix + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)), nil
}

func applicationEvaluationDeterministicCampaignID(ctx ApplicationEvaluationContext, clientKey string) string {
	digest := sha256.Sum256([]byte(strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID, clientKey}, "\x00")))
	encoded := strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(digest[:10]))
	return "aecamp_" + encoded
}

func applicationEvaluationDeterministicRunID(campaignID, itemKey string) string {
	digest := sha256.Sum256([]byte(campaignID + "\x00" + itemKey))
	return "run_" + hex.EncodeToString(digest[:])
}

func cloneApplicationEvaluationPlan(value ApplicationEvaluationPlan) ApplicationEvaluationPlan {
	return value
}

func cloneApplicationEvaluationPlanVersion(value ApplicationEvaluationPlanVersion) ApplicationEvaluationPlanVersion {
	payload, _ := json.Marshal(value)
	var clone ApplicationEvaluationPlanVersion
	_ = json.Unmarshal(payload, &clone)
	return clone
}

func cloneApplicationEvaluationCampaign(value ApplicationEvaluationCampaign) ApplicationEvaluationCampaign {
	payload, _ := json.Marshal(value)
	var clone ApplicationEvaluationCampaign
	_ = json.Unmarshal(payload, &clone)
	return clone
}

type applicationEvaluationCursor struct {
	Version       int    `json:"version"`
	Kind          string `json:"kind"`
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	Environment   string `json:"environment"`
	ApplicationID string `json:"application_id"`
	Filter        string `json:"filter"`
	BeforeTime    string `json:"before_time"`
	BeforeID      string `json:"before_id"`
	BeforeVersion int    `json:"before_version"`
	Limit         int    `json:"limit"`
}

func encodeApplicationEvaluationCursor(cursor applicationEvaluationCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeApplicationEvaluationCursor(value string) (applicationEvaluationCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return applicationEvaluationCursor{}, err
	}
	var cursor applicationEvaluationCursor
	if err = decodeStrictApplicationEvaluationJSON(payload, &cursor); err != nil || cursor.Version != 1 {
		return applicationEvaluationCursor{}, errApplicationEvaluationStoreContract
	}
	return cursor, nil
}

func sortApplicationEvaluationPlans(values []ApplicationEvaluationPlan) {
	sort.Slice(values, func(i, j int) bool {
		if values[i].UpdatedAt == values[j].UpdatedAt {
			return values[i].PlanID > values[j].PlanID
		}
		return values[i].UpdatedAt > values[j].UpdatedAt
	})
}

func sortApplicationEvaluationCampaigns(values []ApplicationEvaluationCampaign) {
	sort.Slice(values, func(i, j int) bool {
		if values[i].CreatedAt == values[j].CreatedAt {
			return values[i].CampaignID > values[j].CampaignID
		}
		return values[i].CreatedAt > values[j].CreatedAt
	})
}
