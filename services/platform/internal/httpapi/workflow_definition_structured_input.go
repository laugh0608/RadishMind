package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	savedWorkflowDraftStructuredSchemaVersion              = "saved_workflow_draft.v2"
	workflowDefinitionCandidateStructuredSchemaVersion     = "workflow_definition_release_candidate.v2"
	workflowDefinitionVersionStructuredSchemaVersion       = "workflow_definition_version.v2"
	workflowDefinitionStructuredExecutorProfile            = "workflow_definition_executor_v2"
	workflowDefinitionStructuredEvaluationProfile          = "workflow_definition_executor.v2"
	workflowRunRecordDefinitionStructuredSchemaVersion     = "workflow_run_record.v8"
	workflowDefinitionStructuredRunComparisonSchemaVersion = "workflow_run_comparison.v7"

	workflowStructuredInputMaxFields      = 16
	workflowStructuredInputMaxStringBytes = 4 * 1024
	workflowStructuredInputMaxBytes       = 8 * 1024
	workflowStructuredInputSafeInteger    = int64(9007199254740991)
)

var workflowStructuredInputFieldNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

type WorkflowStructuredInputValueType string

const (
	WorkflowStructuredInputString  WorkflowStructuredInputValueType = "string"
	WorkflowStructuredInputInteger WorkflowStructuredInputValueType = "integer"
	WorkflowStructuredInputNumber  WorkflowStructuredInputValueType = "number"
	WorkflowStructuredInputBoolean WorkflowStructuredInputValueType = "boolean"
)

type WorkflowStructuredInputField struct {
	Name        string                           `json:"name"`
	ValueType   WorkflowStructuredInputValueType `json:"value_type"`
	Required    bool                             `json:"required"`
	Label       string                           `json:"label"`
	Description string                           `json:"description"`
}

type WorkflowStructuredInputMetadataField struct {
	Name      string                           `json:"name"`
	ValueType WorkflowStructuredInputValueType `json:"value_type"`
}

type workflowStructuredInputNormalization struct {
	Contract         SavedWorkflowDraftContract
	CanonicalPayload []byte
	InputDigest      string
	InputBytes       int
	Fields           []WorkflowStructuredInputMetadataField
}

func supportedSavedWorkflowDraftSchemaVersion(value string) bool {
	return value == savedWorkflowDraftSchemaVersion || value == savedWorkflowDraftStructuredSchemaVersion
}

func workflowDefinitionSchemaIdentityForDraft(schemaVersion string) (candidateSchema, versionSchema, executorProfile string, ok bool) {
	switch schemaVersion {
	case savedWorkflowDraftSchemaVersion:
		return workflowDefinitionCandidateSchemaVersion, workflowDefinitionVersionSchemaVersion, workflowDefinitionExecutorProfile, true
	case savedWorkflowDraftStructuredSchemaVersion:
		return workflowDefinitionCandidateStructuredSchemaVersion, workflowDefinitionVersionStructuredSchemaVersion, workflowDefinitionStructuredExecutorProfile, true
	default:
		return "", "", "", false
	}
}

func workflowDefinitionCandidateSchemaMatchesSnapshot(candidateSchema, snapshotSchema string) bool {
	expected, _, _, ok := workflowDefinitionSchemaIdentityForDraft(snapshotSchema)
	return ok && candidateSchema == expected
}

func workflowDefinitionCandidateSchemaForDraft(snapshotSchema string) string {
	candidateSchema, _, _, _ := workflowDefinitionSchemaIdentityForDraft(snapshotSchema)
	return candidateSchema
}

func workflowDefinitionVersionSchemaMatchesSnapshot(versionSchema, snapshotSchema string) bool {
	_, expected, _, ok := workflowDefinitionSchemaIdentityForDraft(snapshotSchema)
	return ok && versionSchema == expected
}

func workflowDefinitionVersionSchemaForCandidate(candidateSchema string) (string, bool) {
	switch candidateSchema {
	case workflowDefinitionCandidateSchemaVersion:
		return workflowDefinitionVersionSchemaVersion, true
	case workflowDefinitionCandidateStructuredSchemaVersion:
		return workflowDefinitionVersionStructuredSchemaVersion, true
	default:
		return "", false
	}
}

func savedWorkflowDraftContractFromDefinition(contract WorkflowDefinitionContract) SavedWorkflowDraftContract {
	return SavedWorkflowDraftContract{
		ContractID: contract.ContractID, RequiredFields: cloneStringSlice(contract.RequiredFields), Summary: contract.Summary,
		Fields: cloneWorkflowStructuredInputFields(contract.Fields), ContractDigest: contract.ContractDigest,
	}
}

func validWorkflowDefinitionSnapshotVersionIdentity(snapshot WorkflowDefinitionSnapshot) bool {
	_, _, executorProfile, ok := workflowDefinitionSchemaIdentityForDraft(snapshot.SchemaVersion)
	if !ok || snapshot.ExecutionProfile != executorProfile {
		return false
	}
	if snapshot.OutputContract.ContractID == "" || len(snapshot.OutputContract.RequiredFields) == 0 ||
		len(snapshot.OutputContract.Fields) != 0 || snapshot.OutputContract.ContractDigest != "" {
		return false
	}
	if snapshot.SchemaVersion == savedWorkflowDraftSchemaVersion {
		return snapshot.InputContract.ContractID != "" && len(snapshot.InputContract.RequiredFields) > 0 &&
			len(snapshot.InputContract.Fields) == 0 && snapshot.InputContract.ContractDigest == ""
	}
	contract := savedWorkflowDraftContractFromDefinition(snapshot.InputContract)
	normalized, failureCode, _ := normalizeWorkflowStructuredInputContract(contract)
	return failureCode == "" && normalized.ContractDigest == snapshot.InputContract.ContractDigest
}

func normalizeWorkflowStructuredInputContract(contract SavedWorkflowDraftContract) (SavedWorkflowDraftContract, WorkflowRunFailureCode, string) {
	normalized := SavedWorkflowDraftContract{
		ContractID:     strings.TrimSpace(contract.ContractID),
		RequiredFields: normalizedStringSet(contract.RequiredFields),
		Summary:        strings.TrimSpace(contract.Summary),
		ContractDigest: strings.TrimSpace(contract.ContractDigest),
		Fields:         cloneWorkflowStructuredInputFields(contract.Fields),
	}
	if !applicationDraftIdentifierPattern.MatchString(normalized.ContractID) || len(normalized.RequiredFields) != 0 ||
		!utf8.ValidString(normalized.Summary) || len([]byte(normalized.Summary)) > maxSavedWorkflowDraftTextLength ||
		len(normalized.Fields) < 1 || len(normalized.Fields) > workflowStructuredInputMaxFields {
		return SavedWorkflowDraftContract{}, WorkflowRunFailureInputContractMismatch, "Workflow structured input contract is invalid."
	}
	if workflowStructuredInputMetadataContainsSecret(normalized.ContractID) || workflowStructuredInputMetadataContainsSecret(normalized.Summary) {
		return SavedWorkflowDraftContract{}, WorkflowRunFailureInputSecretMaterialForbidden, "Workflow structured input contract contains forbidden secret material."
	}
	seen := make(map[string]struct{}, len(normalized.Fields))
	for index := range normalized.Fields {
		field := &normalized.Fields[index]
		field.Name = strings.TrimSpace(field.Name)
		field.ValueType = WorkflowStructuredInputValueType(strings.TrimSpace(string(field.ValueType)))
		field.Label = strings.TrimSpace(field.Label)
		field.Description = strings.TrimSpace(field.Description)
		if !workflowStructuredInputFieldNamePattern.MatchString(field.Name) || !supportedWorkflowStructuredInputValueType(field.ValueType) ||
			field.Label == "" || !utf8.ValidString(field.Label) || len([]byte(field.Label)) > maxSavedWorkflowDraftLabelLength ||
			!utf8.ValidString(field.Description) || len([]byte(field.Description)) > maxSavedWorkflowDraftTextLength {
			return SavedWorkflowDraftContract{}, WorkflowRunFailureInputContractMismatch, "Workflow structured input contract field is invalid."
		}
		if _, exists := seen[field.Name]; exists {
			return SavedWorkflowDraftContract{}, WorkflowRunFailureInputContractMismatch, "Workflow structured input contract contains duplicate fields."
		}
		seen[field.Name] = struct{}{}
		if workflowStructuredInputMetadataContainsSecret(field.Name) || workflowStructuredInputMetadataContainsSecret(field.Label) ||
			workflowStructuredInputMetadataContainsSecret(field.Description) {
			return SavedWorkflowDraftContract{}, WorkflowRunFailureInputSecretMaterialForbidden, "Workflow structured input contract contains forbidden secret material."
		}
	}
	digest, err := workflowStructuredInputContractDigest(normalized)
	if err != nil {
		return SavedWorkflowDraftContract{}, WorkflowRunFailureInputContractMismatch, "Workflow structured input contract could not be canonicalized."
	}
	if normalized.ContractDigest != "" && normalized.ContractDigest != digest {
		return SavedWorkflowDraftContract{}, WorkflowRunFailureInputContractMismatch, "Workflow structured input contract digest does not match its fields."
	}
	normalized.ContractDigest = digest
	return normalized, "", ""
}

func normalizeWorkflowStructuredInputValues(contract SavedWorkflowDraftContract, inputs map[string]any) (workflowStructuredInputNormalization, WorkflowRunFailureCode, string) {
	normalizedContract, code, summary := normalizeWorkflowStructuredInputContract(contract)
	if code != "" {
		return workflowStructuredInputNormalization{}, code, summary
	}
	if inputs == nil {
		return workflowStructuredInputNormalization{}, WorkflowRunFailureInputContractMismatch, "Workflow structured inputs must be a JSON object."
	}
	contractFields := make(map[string]WorkflowStructuredInputField, len(normalizedContract.Fields))
	for _, field := range normalizedContract.Fields {
		contractFields[field.Name] = field
		if field.Required {
			if _, found := inputs[field.Name]; !found {
				return workflowStructuredInputNormalization{}, WorkflowRunFailureInputRequiredFieldMissing, "Workflow structured input is missing a required field."
			}
		}
	}
	names := make([]string, 0, len(inputs))
	for name := range inputs {
		if _, found := contractFields[name]; !found {
			return workflowStructuredInputNormalization{}, WorkflowRunFailureInputUnknownField, "Workflow structured input contains an unknown field."
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var payload bytes.Buffer
	payload.WriteByte('{')
	metadata := make([]WorkflowStructuredInputMetadataField, 0, len(names))
	for index, name := range names {
		if index > 0 {
			payload.WriteByte(',')
		}
		encodedName, _ := json.Marshal(name)
		payload.Write(encodedName)
		payload.WriteByte(':')
		field := contractFields[name]
		encodedValue, failure := canonicalWorkflowStructuredInputValue(field, inputs[name])
		if failure != "" {
			return workflowStructuredInputNormalization{}, failure, workflowStructuredInputFailureSummary(failure)
		}
		payload.Write(encodedValue)
		metadata = append(metadata, WorkflowStructuredInputMetadataField{Name: name, ValueType: field.ValueType})
	}
	payload.WriteByte('}')
	if payload.Len() > workflowStructuredInputMaxBytes {
		return workflowStructuredInputNormalization{}, WorkflowRunFailureInputBudgetExceeded, "Workflow structured input exceeded the execution budget."
	}
	digest := sha256.Sum256(payload.Bytes())
	return workflowStructuredInputNormalization{
		Contract:         normalizedContract,
		CanonicalPayload: append([]byte(nil), payload.Bytes()...),
		InputDigest:      "sha256:" + hex.EncodeToString(digest[:]),
		InputBytes:       payload.Len(),
		Fields:           metadata,
	}, "", ""
}

func workflowStructuredInputContractDigest(contract SavedWorkflowDraftContract) (string, error) {
	document := struct {
		ContractID string                         `json:"contract_id"`
		Fields     []WorkflowStructuredInputField `json:"fields"`
		Summary    string                         `json:"summary"`
	}{ContractID: contract.ContractID, Fields: contract.Fields, Summary: contract.Summary}
	payload, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func canonicalWorkflowStructuredInputValue(field WorkflowStructuredInputField, value any) ([]byte, WorkflowRunFailureCode) {
	if value == nil {
		return nil, WorkflowRunFailureInputValueTypeInvalid
	}
	switch field.ValueType {
	case WorkflowStructuredInputString:
		text, ok := value.(string)
		if !ok || !utf8.ValidString(text) {
			return nil, WorkflowRunFailureInputValueTypeInvalid
		}
		if len([]byte(text)) > workflowStructuredInputMaxStringBytes {
			return nil, WorkflowRunFailureInputBudgetExceeded
		}
		if workflowStructuredInputValueContainsSecret(text) {
			return nil, WorkflowRunFailureInputSecretMaterialForbidden
		}
		encoded, _ := json.Marshal(text)
		return encoded, ""
	case WorkflowStructuredInputInteger:
		integer, ok := workflowStructuredInputInteger(value)
		if !ok {
			return nil, WorkflowRunFailureInputValueTypeInvalid
		}
		return []byte(strconv.FormatInt(integer, 10)), ""
	case WorkflowStructuredInputNumber:
		number, ok := workflowStructuredInputNumber(value)
		if !ok {
			return nil, WorkflowRunFailureInputValueTypeInvalid
		}
		if number == 0 {
			return []byte("0"), ""
		}
		return []byte(strconv.FormatFloat(number, 'g', -1, 64)), ""
	case WorkflowStructuredInputBoolean:
		boolean, ok := value.(bool)
		if !ok {
			return nil, WorkflowRunFailureInputValueTypeInvalid
		}
		return []byte(strconv.FormatBool(boolean)), ""
	default:
		return nil, WorkflowRunFailureInputValueTypeInvalid
	}
}

func workflowStructuredInputInteger(value any) (int64, bool) {
	var result int64
	switch typed := value.(type) {
	case int:
		result = int64(typed)
	case int8:
		result = int64(typed)
	case int16:
		result = int64(typed)
	case int32:
		result = int64(typed)
	case int64:
		result = typed
	case uint:
		if uint64(typed) > uint64(workflowStructuredInputSafeInteger) {
			return 0, false
		}
		result = int64(typed)
	case uint8:
		result = int64(typed)
	case uint16:
		result = int64(typed)
	case uint32:
		result = int64(typed)
	case uint64:
		if typed > uint64(workflowStructuredInputSafeInteger) {
			return 0, false
		}
		result = int64(typed)
	case json.Number:
		if integer, err := typed.Int64(); err == nil {
			result = integer
		} else {
			parsed, err := typed.Float64()
			if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || math.Trunc(parsed) != parsed {
				return 0, false
			}
			result = int64(parsed)
		}
	case float32:
		parsed := float64(typed)
		if math.Trunc(parsed) != parsed {
			return 0, false
		}
		result = int64(parsed)
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || math.Trunc(typed) != typed {
			return 0, false
		}
		result = int64(typed)
	default:
		return 0, false
	}
	return result, result >= -workflowStructuredInputSafeInteger && result <= workflowStructuredInputSafeInteger
}

func workflowStructuredInputNumber(value any) (float64, bool) {
	var number float64
	switch typed := value.(type) {
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		number = parsed
	case float32:
		number = float64(typed)
	case float64:
		number = typed
	case int:
		number = float64(typed)
	case int8:
		number = float64(typed)
	case int16:
		number = float64(typed)
	case int32:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case uint:
		number = float64(typed)
	case uint8:
		number = float64(typed)
	case uint16:
		number = float64(typed)
	case uint32:
		number = float64(typed)
	case uint64:
		number = float64(typed)
	default:
		return 0, false
	}
	return number, !math.IsNaN(number) && !math.IsInf(number, 0)
}

func supportedWorkflowStructuredInputValueType(value WorkflowStructuredInputValueType) bool {
	return value == WorkflowStructuredInputString || value == WorkflowStructuredInputInteger ||
		value == WorkflowStructuredInputNumber || value == WorkflowStructuredInputBoolean
}

func workflowStructuredInputMetadataContainsSecret(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return false
	}
	for _, marker := range []string{"authorization", "api_key", "api-key", "api key", "token", "cookie", "dsn", "password", "private key", "private_key", "credential", "secret"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return workflowStructuredInputValueContainsSecret(value)
}

func workflowStructuredInputValueContainsSecret(value string) bool {
	lower := strings.ToLower(value)
	if promptApplicationContainsSecretMaterial(value) || strings.Contains(lower, "-----begin private key-----") {
		return true
	}
	for _, marker := range []string{"password=", "client_secret=", "access_token=", "refresh_token="} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func workflowStructuredInputFailureSummary(code WorkflowRunFailureCode) string {
	switch code {
	case WorkflowRunFailureInputBudgetExceeded:
		return "Workflow structured input exceeded the execution budget."
	case WorkflowRunFailureInputSecretMaterialForbidden:
		return "Workflow structured input contains forbidden secret material."
	default:
		return "Workflow structured input value does not match the immutable contract."
	}
}

func cloneWorkflowStructuredInputFields(fields []WorkflowStructuredInputField) []WorkflowStructuredInputField {
	if fields == nil {
		return nil
	}
	cloned := make([]WorkflowStructuredInputField, len(fields))
	copy(cloned, fields)
	return cloned
}

func cloneWorkflowStructuredInputMetadataFields(fields []WorkflowStructuredInputMetadataField) []WorkflowStructuredInputMetadataField {
	if fields == nil {
		return nil
	}
	cloned := make([]WorkflowStructuredInputMetadataField, len(fields))
	copy(cloned, fields)
	return cloned
}

func cloneWorkflowStructuredInputValues(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}
	cloned := make(map[string]any, len(values))
	for name, value := range values {
		cloned[name] = value
	}
	return cloned
}

func workflowDefinitionStructuredPromptInput(input workflowStructuredInputNormalization) (string, error) {
	document := struct {
		SchemaVersion  string                                 `json:"schema_version"`
		ContractID     string                                 `json:"contract_id"`
		ContractDigest string                                 `json:"contract_digest"`
		Fields         []WorkflowStructuredInputMetadataField `json:"fields"`
		Values         json.RawMessage                        `json:"values"`
	}{
		SchemaVersion:  "workflow_structured_input_packet.v1",
		ContractID:     input.Contract.ContractID,
		ContractDigest: input.Contract.ContractDigest,
		Fields:         cloneWorkflowStructuredInputMetadataFields(input.Fields),
		Values:         json.RawMessage(append([]byte(nil), input.CanonicalPayload...)),
	}
	payload, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func validateWorkflowDefinitionStructuredRunStoreRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || record.SchemaVersion != workflowRunRecordDefinitionStructuredSchemaVersion || record.DefinitionAuthority == nil ||
		record.ExecutionKind != workflowDefinitionExecutionKind || record.ExecutionSourceKind != workflowDefinitionExecutionSourceKind ||
		record.ExecutionProfile != workflowDefinitionStructuredExecutorProfile || record.ExecutionSourceID == "" || record.ExecutionSourceVersion < 1 ||
		record.Status == WorkflowRunStatusOutcomeUnknown || record.ExecutionSource == nil || record.ExecutionSource.Kind != record.ExecutionKind ||
		record.ExecutionSource.SourceKind != record.ExecutionSourceKind || record.ExecutionSource.ID != record.ExecutionSourceID ||
		record.ExecutionSource.Version != record.ExecutionSourceVersion || record.DraftID != "" || record.DraftVersion != 0 || record.DraftDigest != "" ||
		record.Output != "" || !applicationDraftIdentifierPattern.MatchString(record.InputContractID) ||
		!workflowRAGDigestPattern.MatchString(record.InputContractDigest) || !workflowRAGDigestPattern.MatchString(record.InputDigest) ||
		record.InputBytes < 2 || record.InputBytes > workflowStructuredInputMaxBytes || len(record.InputFields) > workflowStructuredInputMaxFields ||
		record.SideEffects.RetrievalCalls != 0 || record.SideEffects.ToolCalls != 0 || record.SideEffects.ConfirmationCalls != 0 ||
		record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 || record.SideEffects.ProviderCalls < 0 ||
		record.SideEffects.ProviderCalls > workflowExecutorMaxLLMCalls || record.PlanID != "" || record.ConfirmationID != "" ||
		record.ToolAttempt != nil || record.RAGSnapshot != nil || record.RetrievalAttempt != nil || record.RAGAnswer != nil || record.RAGApplication != nil ||
		!validWorkflowRunDiagnostic(record.Diagnostic, isTerminalWorkflowRunStatus(record.Status)) {
		return errWorkflowRunStoreContract
	}
	previousName := ""
	for _, field := range record.InputFields {
		if !workflowStructuredInputFieldNamePattern.MatchString(field.Name) || !supportedWorkflowStructuredInputValueType(field.ValueType) ||
			(previousName != "" && field.Name <= previousName) {
			return errWorkflowRunStoreContract
		}
		previousName = field.Name
	}
	authority := record.DefinitionAuthority
	if authority.DefinitionID != record.ExecutionSourceID || authority.DefinitionVersion != record.ExecutionSourceVersion ||
		!applicationDraftIdentifierPattern.MatchString(authority.DefinitionID) || !workflowRAGDigestPattern.MatchString(authority.DefinitionDigest) ||
		authority.ActivationPointerVersion < 1 || !applicationDraftIdentifierPattern.MatchString(authority.CandidateID) || authority.CandidateReviewVersion < 1 ||
		!applicationDraftIdentifierPattern.MatchString(authority.SourceDraftID) || authority.SourceDraftVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(authority.SourceDraftDigest) || authority.SourceDraftDigest != authority.DefinitionDigest ||
		authority.ApplicationRecordVersion < 1 || authority.ApplicationLifecycle != applicationCatalogLifecycleActive {
		return errWorkflowRunStoreContract
	}
	for _, node := range record.Nodes {
		if node.OutputPreview != "" {
			return errWorkflowRunStoreContract
		}
	}
	return nil
}
