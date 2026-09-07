package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
)

const actionSafetyDecisionSchemaVersion = "action_safety_decision.v1"

type ActionSafetyLevel string

const (
	ActionSafetyLevelAnswerOnly           ActionSafetyLevel = "answer_only"
	ActionSafetyLevelProposalOnly         ActionSafetyLevel = "proposal_only"
	ActionSafetyLevelHandoffReady         ActionSafetyLevel = "handoff_ready"
	ActionSafetyLevelToolCallable         ActionSafetyLevel = "tool_callable"
	ActionSafetyLevelWriteBlocked         ActionSafetyLevel = "write_blocked"
	ActionSafetyLevelWriteAllowedByPolicy ActionSafetyLevel = "write_allowed_by_policy"
)

type ActionSafetyFailureCode string

const (
	ActionSafetyFailureScopeDenied              ActionSafetyFailureCode = "action_safety_scope_denied"
	ActionSafetyFailurePayloadInvalid           ActionSafetyFailureCode = "action_safety_payload_invalid"
	ActionSafetyFailureSourceUnavailable        ActionSafetyFailureCode = "action_safety_source_unavailable"
	ActionSafetyFailureSourceChanged            ActionSafetyFailureCode = "action_safety_source_changed"
	ActionSafetyFailurePolicyUnavailable        ActionSafetyFailureCode = "action_safety_policy_unavailable"
	ActionSafetyFailurePolicyChanged            ActionSafetyFailureCode = "action_safety_policy_changed"
	ActionSafetyFailureLevelEscalationDenied    ActionSafetyFailureCode = "action_safety_level_escalation_denied"
	ActionSafetyFailureConfirmationRequired     ActionSafetyFailureCode = "action_safety_confirmation_required"
	ActionSafetyFailureConfirmationChanged      ActionSafetyFailureCode = "action_safety_confirmation_changed"
	ActionSafetyFailureToolAuthorityUnavailable ActionSafetyFailureCode = "action_safety_tool_authority_unavailable"
	ActionSafetyFailureWriteBlocked             ActionSafetyFailureCode = "action_safety_write_blocked"
	ActionSafetyFailureStoreContractMismatch    ActionSafetyFailureCode = "action_safety_store_contract_mismatch"
)

type ActionSafetySideEffectBudget struct {
	ProviderCalls            int `json:"provider_calls"`
	HandoffRefs              int `json:"handoff_refs"`
	ToolNetworkCalls         int `json:"tool_network_calls"`
	ConfirmationConsumptions int `json:"confirmation_consumptions"`
	BusinessWrites           int `json:"business_writes"`
	ReplayWrites             int `json:"replay_writes"`
}

type ActionSafetyDecision struct {
	SchemaVersion        string                       `json:"schema_version"`
	DecisionID           string                       `json:"decision_id"`
	DecisionDigest       string                       `json:"decision_digest"`
	TenantRef            string                       `json:"tenant_ref"`
	WorkspaceID          string                       `json:"workspace_id"`
	Environment          string                       `json:"environment"`
	ApplicationID        string                       `json:"application_id"`
	SourceKind           string                       `json:"source_kind"`
	SourceSchemaVersion  string                       `json:"source_schema_version"`
	SourceID             string                       `json:"source_id"`
	SourceVersion        int                          `json:"source_version"`
	SourceDigest         string                       `json:"source_digest"`
	Project              string                       `json:"project"`
	Task                 string                       `json:"task"`
	ActionKind           string                       `json:"action_kind"`
	RiskLevel            string                       `json:"risk_level"`
	TargetKind           string                       `json:"target_kind"`
	Method               string                       `json:"method"`
	RequestedLevel       ActionSafetyLevel            `json:"requested_level"`
	MaximumAllowedLevel  ActionSafetyLevel            `json:"maximum_allowed_level"`
	EffectiveLevel       ActionSafetyLevel            `json:"effective_level"`
	RequiresConfirmation bool                         `json:"requires_confirmation"`
	ConfirmationState    string                       `json:"confirmation_state"`
	WritesBusinessTruth  bool                         `json:"writes_business_truth"`
	SideEffectBudget     ActionSafetySideEffectBudget `json:"side_effect_budget"`
	Blockers             []ActionSafetyFailureCode    `json:"blockers"`
	PolicyVersion        string                       `json:"policy_version"`
	PolicyDigest         string                       `json:"policy_digest"`
	ActorRef             string                       `json:"actor_ref"`
	RequestID            string                       `json:"request_id"`
	AuditRef             string                       `json:"audit_ref"`
	CreatedAt            string                       `json:"created_at"`
}

var actionSafetyDecisionIDPattern = regexp.MustCompile(`^asd_[a-z2-7]{16}$`)
var actionSafetyVersionPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]*\.v[0-9]+$`)
var actionSafetyTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,119}$`)

var actionSafetyBlockerOrder = []ActionSafetyFailureCode{
	ActionSafetyFailureScopeDenied,
	ActionSafetyFailurePayloadInvalid,
	ActionSafetyFailureSourceUnavailable,
	ActionSafetyFailureSourceChanged,
	ActionSafetyFailurePolicyUnavailable,
	ActionSafetyFailurePolicyChanged,
	ActionSafetyFailureLevelEscalationDenied,
	ActionSafetyFailureConfirmationRequired,
	ActionSafetyFailureConfirmationChanged,
	ActionSafetyFailureToolAuthorityUnavailable,
	ActionSafetyFailureWriteBlocked,
	ActionSafetyFailureStoreContractMismatch,
}

var actionSafetyDecisionRequiredFields = []string{
	"schema_version", "decision_id", "decision_digest", "tenant_ref", "workspace_id", "environment", "application_id",
	"source_kind", "source_schema_version", "source_id", "source_version", "source_digest", "project", "task", "action_kind",
	"risk_level", "target_kind", "method", "requested_level", "maximum_allowed_level", "effective_level", "requires_confirmation",
	"confirmation_state", "writes_business_truth", "side_effect_budget", "blockers", "policy_version", "policy_digest", "actor_ref",
	"request_id", "audit_ref", "created_at",
}

var errActionSafetyContract = errors.New("action safety contract is invalid")

func encodeActionSafetyDecision(value ActionSafetyDecision) ([]byte, error) {
	if validateActionSafetyDecision(value) != nil {
		return nil, errActionSafetyContract
	}
	payload, err := json.Marshal(value)
	if err != nil || applicationDraftStringContainsSecret(string(payload)) {
		return nil, errActionSafetyContract
	}
	return payload, nil
}

func decodeActionSafetyDecision(payload []byte) (ActionSafetyDecision, error) {
	if len(payload) == 0 || len(payload) > 16<<10 || validateNoDuplicateJSONFields(payload) != nil {
		return ActionSafetyDecision{}, errActionSafetyContract
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil || len(raw) != len(actionSafetyDecisionRequiredFields) {
		return ActionSafetyDecision{}, errActionSafetyContract
	}
	for _, field := range actionSafetyDecisionRequiredFields {
		if _, ok := raw[field]; !ok {
			return ActionSafetyDecision{}, errActionSafetyContract
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var value ActionSafetyDecision
	if err := decoder.Decode(&value); err != nil || decoder.Decode(&struct{}{}) != io.EOF || validateActionSafetyDecision(value) != nil {
		return ActionSafetyDecision{}, errActionSafetyContract
	}
	return value, nil
}

func validateActionSafetyDecision(value ActionSafetyDecision) error {
	if value.SchemaVersion != actionSafetyDecisionSchemaVersion || !actionSafetyDecisionIDPattern.MatchString(value.DecisionID) ||
		!workflowRAGDigestPattern.MatchString(value.DecisionDigest) || !validActionSafetyScope(ActionSafetyScope{
		TenantRef: value.TenantRef, WorkspaceID: value.WorkspaceID, Environment: value.Environment, ApplicationID: value.ApplicationID,
	}) || !validActionSafetySourceReference(ActionSafetySourceReference{
		Kind: value.SourceKind, SchemaVersion: value.SourceSchemaVersion, ID: value.SourceID, Version: value.SourceVersion, Digest: value.SourceDigest,
	}) || !validActionSafetyPolicyReference(ActionSafetyPolicyReference{Version: value.PolicyVersion, Digest: value.PolicyDigest}) ||
		!validPromptApplicationRef(value.ActorRef) || !validPromptApplicationRef(value.RequestID) || !validPromptApplicationRef(value.AuditRef) ||
		parsePromptApplicationTemplateTimestamp(value.CreatedAt) == nil || !validActionSafetySourceCompatibility(value.SourceKind, value.SourceSchemaVersion, value.Project, value.Task, value.ActionKind, value.RiskLevel, value.TargetKind) ||
		!validActionSafetyDecisionMethod(value.TargetKind, value.Method) ||
		!validActionSafetyLevel(value.RequestedLevel) || !validActionSafetyLevel(value.MaximumAllowedLevel) || !validActionSafetyLevel(value.EffectiveLevel) ||
		value.RequestedLevel == ActionSafetyLevelWriteBlocked || value.MaximumAllowedLevel == ActionSafetyLevelWriteBlocked ||
		value.MaximumAllowedLevel == ActionSafetyLevelWriteAllowedByPolicy || value.EffectiveLevel == ActionSafetyLevelWriteAllowedByPolicy ||
		!validActionSafetyConfirmation(value.RequiresConfirmation, value.ConfirmationState) || value.Blockers == nil || !canonicalActionSafetyBlockers(value.Blockers) {
		return errActionSafetyContract
	}
	writeRequest := actionSafetyTargetIsWriteBlocked(value.TargetKind) || value.WritesBusinessTruth || value.Method != "none" && value.Method != "GET"
	expectedLevel, ok := actionSafetyTransition(value.RequestedLevel, value.MaximumAllowedLevel, writeRequest)
	if !ok || expectedLevel != value.EffectiveLevel || value.SideEffectBudget != actionSafetySideEffectBudget(value.EffectiveLevel) {
		return errActionSafetyContract
	}
	escalated := !writeRequest && value.EffectiveLevel != value.RequestedLevel
	fundamentalBlocker := actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureScopeDenied) ||
		actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureSourceUnavailable) || actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureSourceChanged) ||
		actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailurePolicyUnavailable) || actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailurePolicyChanged)
	if writeRequest != actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureWriteBlocked) ||
		escalated != actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureLevelEscalationDenied) ||
		fundamentalBlocker && value.MaximumAllowedLevel != ActionSafetyLevelAnswerOnly ||
		actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureToolAuthorityUnavailable) && value.MaximumAllowedLevel == ActionSafetyLevelToolCallable ||
		actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureConfirmationRequired) && value.ConfirmationState == "approved" ||
		(value.RequestedLevel == ActionSafetyLevelToolCallable && value.ConfirmationState != "approved") != actionSafetyContainsBlocker(value.Blockers, ActionSafetyFailureConfirmationRequired) ||
		(value.EffectiveLevel == ActionSafetyLevelToolCallable && (!value.RequiresConfirmation || value.ConfirmationState != "approved" || len(value.Blockers) != 0)) ||
		(value.EffectiveLevel != ActionSafetyLevelToolCallable && value.SideEffectBudget.ToolNetworkCalls != 0) {
		return errActionSafetyContract
	}
	digest, err := actionSafetyDecisionDigest(value)
	if err != nil || digest != value.DecisionDigest {
		return errActionSafetyContract
	}
	return nil
}

func actionSafetyDecisionDigest(value ActionSafetyDecision) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return "", err
	}
	delete(document, "decision_digest")
	return canonicalSHA256(document)
}

func validActionSafetyLevel(value ActionSafetyLevel) bool {
	switch value {
	case ActionSafetyLevelAnswerOnly, ActionSafetyLevelProposalOnly, ActionSafetyLevelHandoffReady,
		ActionSafetyLevelToolCallable, ActionSafetyLevelWriteBlocked, ActionSafetyLevelWriteAllowedByPolicy:
		return true
	default:
		return false
	}
}

func validActionSafetyConfirmation(required bool, state string) bool {
	switch state {
	case "not_required":
		return !required
	case "pending", "approved", "rejected", "changed":
		return required
	default:
		return false
	}
}

func validActionSafetyDecisionMethod(targetKind, method string) bool {
	if targetKind == actionSafetyTargetWorkflowHTTPTool {
		return agentCopilotContainsString([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, method)
	}
	return method == "none"
}

func canonicalActionSafetyBlockers(values []ActionSafetyFailureCode) bool {
	position := -1
	for _, value := range values {
		index := -1
		for candidateIndex, candidate := range actionSafetyBlockerOrder {
			if value == candidate {
				index = candidateIndex
				break
			}
		}
		if index <= position {
			return false
		}
		position = index
	}
	return true
}

func normalizeActionSafetyBlockers(selected map[ActionSafetyFailureCode]bool) []ActionSafetyFailureCode {
	blockers := make([]ActionSafetyFailureCode, 0, len(selected))
	for _, code := range actionSafetyBlockerOrder {
		if selected[code] {
			blockers = append(blockers, code)
		}
	}
	return blockers
}

func actionSafetyContainsBlocker(values []ActionSafetyFailureCode, target ActionSafetyFailureCode) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func actionSafetySideEffectBudget(level ActionSafetyLevel) ActionSafetySideEffectBudget {
	switch level {
	case ActionSafetyLevelAnswerOnly, ActionSafetyLevelProposalOnly:
		return ActionSafetySideEffectBudget{ProviderCalls: 1}
	case ActionSafetyLevelHandoffReady:
		return ActionSafetySideEffectBudget{HandoffRefs: 1}
	case ActionSafetyLevelToolCallable:
		return ActionSafetySideEffectBudget{ToolNetworkCalls: 1, ConfirmationConsumptions: 1}
	default:
		return ActionSafetySideEffectBudget{}
	}
}

func validActionSafetyReference(value string) bool {
	return validPromptApplicationRef(value) && strings.TrimSpace(value) == value
}

func validActionSafetyTimestamp(value string) bool {
	_, err := time.Parse(time.RFC3339Nano, value)
	return err == nil
}
