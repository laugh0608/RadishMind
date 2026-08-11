package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWorkflowStructuredInputCanonicalizationIsDeterministicAndMetadataOnly(t *testing.T) {
	contract := workflowStructuredInputContractForTest()
	first, code, summary := normalizeWorkflowStructuredInputValues(contract, map[string]any{
		"score":         json.Number("1.50"),
		"customer_name": "Ada",
		"dry_run":       true,
		"retry_count":   json.Number("2"),
	})
	if code != "" {
		t.Fatalf("normalize first input: code=%s summary=%s", code, summary)
	}
	second, code, summary := normalizeWorkflowStructuredInputValues(contract, map[string]any{
		"retry_count":   int64(2),
		"dry_run":       true,
		"customer_name": "Ada",
		"score":         1.5,
	})
	if code != "" {
		t.Fatalf("normalize second input: code=%s summary=%s", code, summary)
	}
	wantCanonical := `{"customer_name":"Ada","dry_run":true,"retry_count":2,"score":1.5}`
	if string(first.CanonicalPayload) != wantCanonical || string(second.CanonicalPayload) != wantCanonical {
		t.Fatalf("canonical payload drift: first=%s second=%s", first.CanonicalPayload, second.CanonicalPayload)
	}
	if first.InputDigest != second.InputDigest || first.InputBytes != len(wantCanonical) || first.Contract.ContractDigest == "" || first.Contract.ContractDigest != second.Contract.ContractDigest {
		t.Fatalf("unstable normalization metadata: first=%#v second=%#v", first, second)
	}
	reorderedContract := contract
	reorderedContract.Fields = cloneWorkflowStructuredInputFields(contract.Fields)
	reorderedContract.Fields[0], reorderedContract.Fields[1] = reorderedContract.Fields[1], reorderedContract.Fields[0]
	reordered, failureCode, failureSummary := normalizeWorkflowStructuredInputContract(reorderedContract)
	if failureCode != "" || reordered.ContractDigest == first.Contract.ContractDigest {
		t.Fatalf("ordered contract identity was not preserved: code=%s summary=%s digest=%s", failureCode, failureSummary, reordered.ContractDigest)
	}
	wantFields := []WorkflowStructuredInputMetadataField{
		{Name: "customer_name", ValueType: WorkflowStructuredInputString},
		{Name: "dry_run", ValueType: WorkflowStructuredInputBoolean},
		{Name: "retry_count", ValueType: WorkflowStructuredInputInteger},
		{Name: "score", ValueType: WorkflowStructuredInputNumber},
	}
	if payload, _ := json.Marshal(first.Fields); string(payload) != mustJSONForStructuredInputTest(t, wantFields) {
		t.Fatalf("input field metadata is not sorted: %s", payload)
	}

	record := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_metadata", first)
	persisted, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(persisted), "Ada") || strings.Contains(string(persisted), wantCanonical) || !strings.Contains(string(persisted), first.InputDigest) {
		t.Fatalf("Run v8 privacy metadata contract violated: %s", persisted)
	}

	optionalContract := SavedWorkflowDraftContract{
		ContractID: "contract_optional_input",
		Summary:    "Optional workflow runtime input.",
		Fields: []WorkflowStructuredInputField{
			{Name: "note", ValueType: WorkflowStructuredInputString, Label: "Note", Description: "Optional advisory note."},
		},
	}
	empty, code, summary := normalizeWorkflowStructuredInputValues(optionalContract, map[string]any{})
	if code != "" || string(empty.CanonicalPayload) != `{}` || len(empty.Fields) != 0 {
		t.Fatalf("optional empty input was not preserved: input=%#v code=%s summary=%s", empty, code, summary)
	}
	emptyRecord := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_empty", empty)
	emptyPersisted, err := json.Marshal(emptyRecord)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(emptyPersisted), `"input_fields"`) {
		t.Fatalf("zero provided fields should use the optional Run v8 property: %s", emptyPersisted)
	}
}

func TestWorkflowStructuredInputValidationUsesStableFailureCodes(t *testing.T) {
	contract := workflowStructuredInputContractForTest()
	tests := []struct {
		name   string
		input  map[string]any
		code   WorkflowRunFailureCode
		mutate func(SavedWorkflowDraftContract) SavedWorkflowDraftContract
	}{
		{name: "object required", input: nil, code: WorkflowRunFailureInputContractMismatch},
		{name: "required field", input: map[string]any{}, code: WorkflowRunFailureInputRequiredFieldMissing},
		{name: "unknown field", input: map[string]any{"customer_name": "Ada", "unexpected": true}, code: WorkflowRunFailureInputUnknownField},
		{name: "string type", input: map[string]any{"customer_name": 12}, code: WorkflowRunFailureInputValueTypeInvalid},
		{name: "null", input: map[string]any{"customer_name": nil}, code: WorkflowRunFailureInputValueTypeInvalid},
		{name: "string budget", input: map[string]any{"customer_name": strings.Repeat("a", workflowStructuredInputMaxStringBytes+1)}, code: WorkflowRunFailureInputBudgetExceeded},
		{name: "secret value", input: map[string]any{"customer_name": "password=hunter2"}, code: WorkflowRunFailureInputSecretMaterialForbidden},
		{name: "unsafe integer", input: map[string]any{"customer_name": "Ada", "retry_count": json.Number("9007199254740992")}, code: WorkflowRunFailureInputValueTypeInvalid},
		{name: "contract digest mismatch", input: map[string]any{"customer_name": "Ada"}, code: WorkflowRunFailureInputContractMismatch, mutate: func(value SavedWorkflowDraftContract) SavedWorkflowDraftContract {
			value.ContractDigest = "sha256:" + strings.Repeat("f", 64)
			return value
		}},
		{name: "secret field metadata", input: map[string]any{"customer_name": "Ada"}, code: WorkflowRunFailureInputSecretMaterialForbidden, mutate: func(value SavedWorkflowDraftContract) SavedWorkflowDraftContract {
			value.Fields[0].Name = "api_key"
			return value
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := contract
			current.Fields = cloneWorkflowStructuredInputFields(contract.Fields)
			if test.mutate != nil {
				current = test.mutate(current)
			}
			_, code, _ := normalizeWorkflowStructuredInputValues(current, test.input)
			if code != test.code {
				t.Fatalf("failure code=%s want=%s", code, test.code)
			}
		})
	}
}

func TestSavedWorkflowDraftV2ValidationAndV1Isolation(t *testing.T) {
	payload := validSavedWorkflowDraftPayload()
	payload.SchemaVersion = savedWorkflowDraftStructuredSchemaVersion
	payload.InputContract = workflowStructuredInputContractForTest()
	service := newSavedWorkflowDraftService(newMemorySavedWorkflowDraftStore())
	context := SavedWorkflowDraftContext{WorkspaceID: payload.WorkspaceID, ApplicationID: payload.ApplicationID, RequestID: "request_structured_validation", AuditRef: "audit_structured_validation"}
	result := service.ValidateDraft(context, ValidateWorkflowDraftRequest{Payload: payload})
	if result.FailureCode != "" || !result.ValidationSummary.ValidForReview || result.ValidationSummary.ValidationState != SavedWorkflowDraftStatusValidForReview {
		t.Fatalf("v2 draft validation failed: %#v", result)
	}

	legacy := validSavedWorkflowDraftPayload()
	legacy.InputContract.Fields = cloneWorkflowStructuredInputFields(payload.InputContract.Fields)
	legacyResult := service.ValidateDraft(context, ValidateWorkflowDraftRequest{Payload: legacy})
	if legacyResult.FailureCode != "" || legacyResult.ValidationSummary.ValidForReview || !savedWorkflowDraftValidationHasFinding(legacyResult.ValidationSummary, SavedWorkflowDraftFailureContractInvalid) {
		t.Fatalf("v1 accepted v2 contract fields: %#v", legacyResult)
	}

	unsupported := payload
	unsupported.SchemaVersion = "saved_workflow_draft.v3"
	unsupportedResult := service.ValidateDraft(context, ValidateWorkflowDraftRequest{Payload: unsupported})
	if unsupportedResult.FailureCode != SavedWorkflowDraftFailureSchemaVersionUnsupported {
		t.Fatalf("unsupported schema did not fail closed: %#v", unsupportedResult)
	}
}

func TestWorkflowDefinitionStructuredReleaseUsesExactVersionIdentity(t *testing.T) {
	context := workflowDefinitionTestContext()
	draft := workflowDefinitionTestDraft()
	draft.SchemaVersion = savedWorkflowDraftStructuredSchemaVersion
	draft.InputContract = workflowStructuredInputContractForTest()
	draft.OutputContract = SavedWorkflowDraftContract{ContractID: "contract_output", RequiredFields: []string{"answer_summary"}, Summary: "Reviewable advisory output."}
	store := newWorkflowDefinitionReleaseStore()
	now := time.Date(2026, 8, 10, 6, 0, 0, 0, time.UTC)
	candidate, err := store.CreateCandidate(context, "candidate_structured", "definition_structured", draft, now)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.SchemaVersion != workflowDefinitionCandidateStructuredSchemaVersion || candidate.Snapshot.SchemaVersion != savedWorkflowDraftStructuredSchemaVersion || candidate.Snapshot.ExecutionProfile != workflowDefinitionStructuredExecutorProfile || candidate.Snapshot.InputContract.ContractDigest == "" {
		t.Fatalf("candidate v2 identity drift: %#v", candidate)
	}
	approved, version, err := store.Review(context, candidate.CandidateID, 0, "approve", "approve exact structured contract", candidate.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || version == nil || approved.State != workflowDefinitionStateApproved {
		t.Fatalf("review v2 candidate: approved=%#v version=%#v err=%v", approved, version, err)
	}
	if version.SchemaVersion != workflowDefinitionVersionStructuredSchemaVersion || version.Snapshot.ExecutionProfile != workflowDefinitionStructuredExecutorProfile || validateStoredWorkflowDefinitionVersion(*version) != nil {
		t.Fatalf("definition v2 identity drift: %#v", version)
	}
	if _, failureCode, _ := buildWorkflowExecutionPlan(draft, map[string]bool{}); failureCode != WorkflowRunFailureDraftNotEligible {
		t.Fatalf("legacy executor accepted v2 draft: %s", failureCode)
	}
	if _, failureCode, failureSummary := buildWorkflowDefinitionExecutionPlan(draft, map[string]bool{}); failureCode != "" {
		t.Fatalf("definition executor rejected exact v2 draft: code=%s summary=%s", failureCode, failureSummary)
	}
	unsupportedDraft := draft
	unsupportedDraft.SchemaVersion = "saved_workflow_draft.v3"
	if _, failureCode, _ := buildWorkflowDefinitionExecutionPlan(unsupportedDraft, map[string]bool{}); failureCode != WorkflowRunFailureInputSchemaUnsupported {
		t.Fatalf("unsupported definition schema did not fail closed: %s", failureCode)
	}

	mixedCandidate := candidate
	mixedCandidate.SchemaVersion = workflowDefinitionCandidateSchemaVersion
	if validateStoredWorkflowDefinitionCandidate(mixedCandidate) == nil {
		t.Fatal("v1 candidate schema accepted a v2 snapshot")
	}
	mixedVersion := *version
	mixedVersion.SchemaVersion = workflowDefinitionVersionSchemaVersion
	if validateStoredWorkflowDefinitionVersion(mixedVersion) == nil {
		t.Fatal("v1 definition schema accepted a v2 snapshot")
	}
	corruptedVersion := *version
	corruptedVersion.Snapshot.Name = "corrupted without digest update"
	if validateStoredWorkflowDefinitionVersion(corruptedVersion) == nil {
		t.Fatal("v2 definition accepted snapshot corruption without a digest update")
	}

	legacy := workflowDefinitionTestDraft()
	legacyCandidate, err := store.CreateCandidate(context, "candidate_legacy", "definition_legacy", legacy, now)
	if err != nil {
		t.Fatal(err)
	}
	_, legacyVersion, err := store.Review(context, legacyCandidate.CandidateID, 0, "approve", "preserve exact legacy identity", legacyCandidate.SourceDraftDigest, now.Add(time.Minute))
	if err != nil || legacyVersion == nil || legacyCandidate.SchemaVersion != workflowDefinitionCandidateSchemaVersion || legacyVersion.SchemaVersion != workflowDefinitionVersionSchemaVersion || legacyCandidate.Snapshot.ExecutionProfile != workflowDefinitionExecutorProfile {
		t.Fatalf("legacy identity changed or auto-migrated: candidate=%#v version=%#v err=%v", legacyCandidate, legacyVersion, err)
	}
}

func TestWorkflowDefinitionStructuredRunAndComparisonCompatibility(t *testing.T) {
	contract := workflowStructuredInputContractForTest()
	normalized, code, summary := normalizeWorkflowStructuredInputValues(contract, map[string]any{"customer_name": "Ada"})
	if code != "" {
		t.Fatalf("normalize comparison input: code=%s summary=%s", code, summary)
	}
	context := workflowExecutorTestContext()
	store := newMemoryWorkflowRunStore(10)
	baseline := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_baseline", normalized)
	candidate := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_candidate", normalized)
	persistTerminalStructuredRunForTest(t, store, context, &baseline)
	persistTerminalStructuredRunForTest(t, store, context, &candidate)
	service := workflowExecutorService{store: store}
	result := service.CompareRuns(context, baseline.RunID, candidate.RunID)
	if result.FailureCode != "" || result.Comparison == nil || result.Comparison.SchemaVersion != workflowDefinitionStructuredRunComparisonSchemaVersion || result.Comparison.RunProfile != workflowDefinitionStructuredEvaluationProfile {
		t.Fatalf("Run v8 comparison failed: %#v", result)
	}

	drifted := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_drifted", normalized)
	drifted.InputContractDigest = "sha256:" + strings.Repeat("b", 64)
	persistTerminalStructuredRunForTest(t, store, context, &drifted)
	if drift := service.CompareRuns(context, baseline.RunID, drifted.RunID); drift.FailureCode != WorkflowRunFailureDefinitionIncompatible {
		t.Fatalf("contract drift did not fail closed: %#v", drift)
	}

	legacy := workflowDefinitionRunRecordForStoreTest(context, "run_structured_legacy")
	persistTerminalStructuredRunForTest(t, store, context, &legacy)
	if mixed := service.CompareRuns(context, baseline.RunID, legacy.RunID); mixed.FailureCode != WorkflowRunFailureDefinitionIncompatible {
		t.Fatalf("mixed v5/v8 comparison did not fail closed: %#v", mixed)
	}

	running := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_running", normalized)
	if err := store.UpsertRun(context, &running); err != nil {
		t.Fatal(err)
	}
	if nonterminal := service.CompareRuns(context, baseline.RunID, running.RunID); nonterminal.FailureCode != WorkflowRunFailureDefinitionIncompatible {
		t.Fatalf("nonterminal Run v8 comparison did not fail closed: %#v", nonterminal)
	}
}

func TestWorkflowDefinitionStructuredEvaluationCaseAndSuiteUseExactProfile(t *testing.T) {
	contract := workflowStructuredInputContractForTest()
	normalized, code, summary := normalizeWorkflowStructuredInputValues(contract, map[string]any{"customer_name": "Ada"})
	if code != "" {
		t.Fatalf("normalize evaluation input: code=%s summary=%s", code, summary)
	}
	context := workflowExecutorTestContext()
	runStore := newMemoryWorkflowRunStore(10)
	baseline := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_eval_base", normalized)
	candidate := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_eval_candidate", normalized)
	persistTerminalStructuredRunForTest(t, runStore, context, &baseline)
	persistTerminalStructuredRunForTest(t, runStore, context, &candidate)

	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), runStore)
	evaluation.newCaseID = func() (string, error) { return "eval_structured_definition", nil }
	created := evaluation.Create(context, WorkflowEvaluationCreateRequest{
		Name:          "Structured definition evaluation",
		BaselineRunID: baseline.RunID,
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID: candidate.RunID, ExpectedClassification: WorkflowRunComparisonUnchanged,
		}},
	})
	if created.FailureCode != "" || created.Case == nil {
		t.Fatalf("create structured evaluation case: %#v", created)
	}
	review := evaluation.Review(context, created.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.RunProfile != workflowDefinitionStructuredEvaluationProfile ||
		len(review.Review.Items) != 1 || review.Review.Items[0].ComparisonSchemaVersion != workflowDefinitionStructuredRunComparisonSchemaVersion {
		t.Fatalf("review structured evaluation case: %#v", review)
	}

	suite := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suite.newSuiteID = func() (string, error) { return "suite_structured_definition", nil }
	createdSuite := suite.Create(context, WorkflowEvaluationSuiteCreateRequest{
		Name: "Structured definition suite",
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{{
			CaseID: created.Case.CaseID, Version: created.Case.Version,
		}},
	})
	if createdSuite.FailureCode != "" || createdSuite.Suite == nil {
		t.Fatalf("create structured evaluation suite: %#v", createdSuite)
	}
	suiteReview := suite.Review(context, createdSuite.Suite.SuiteID)
	if suiteReview.FailureCode != "" || suiteReview.Review == nil || len(suiteReview.Review.Items) != 1 ||
		suiteReview.Review.Items[0].RunProfile != workflowDefinitionStructuredEvaluationProfile {
		t.Fatalf("review structured evaluation suite: %#v", suiteReview)
	}

	drifted := workflowDefinitionStructuredRunRecordForTest(t, "run_structured_eval_drift", normalized)
	drifted.InputContractDigest = "sha256:" + strings.Repeat("c", 64)
	persistTerminalStructuredRunForTest(t, runStore, context, &drifted)
	rejected := evaluation.Create(context, WorkflowEvaluationCreateRequest{
		Name:          "Structured contract drift",
		BaselineRunID: baseline.RunID,
		Expectations: []WorkflowEvaluationExpectation{{
			CandidateRunID: drifted.RunID, ExpectedClassification: WorkflowRunComparisonChanged,
		}},
	})
	if rejected.FailureCode != WorkflowEvaluationFailureDefinitionIncompatible {
		t.Fatalf("structured evaluation accepted contract drift: %#v", rejected)
	}
}

func workflowStructuredInputContractForTest() SavedWorkflowDraftContract {
	return SavedWorkflowDraftContract{
		ContractID: "contract_structured_input",
		Summary:    "Bounded workflow runtime inputs.",
		Fields: []WorkflowStructuredInputField{
			{Name: "customer_name", ValueType: WorkflowStructuredInputString, Required: true, Label: "Customer name", Description: "Name used in the advisory request."},
			{Name: "retry_count", ValueType: WorkflowStructuredInputInteger, Label: "Retry count", Description: "Bounded retry count."},
			{Name: "score", ValueType: WorkflowStructuredInputNumber, Label: "Score", Description: "Finite evaluation score."},
			{Name: "dry_run", ValueType: WorkflowStructuredInputBoolean, Label: "Dry run", Description: "Whether to produce advisory output only."},
		},
	}
}

func workflowDefinitionStructuredRunRecordForTest(t *testing.T, runID string, input workflowStructuredInputNormalization) WorkflowRunRecord {
	t.Helper()
	context := workflowExecutorTestContext()
	record := workflowDefinitionRunRecordForStoreTest(context, runID)
	record.SchemaVersion = workflowRunRecordDefinitionStructuredSchemaVersion
	record.ExecutionProfile = workflowDefinitionStructuredExecutorProfile
	record.InputContractID = input.Contract.ContractID
	record.InputContractDigest = input.Contract.ContractDigest
	record.InputDigest = input.InputDigest
	record.InputBytes = input.InputBytes
	record.InputFields = cloneWorkflowStructuredInputMetadataFields(input.Fields)
	return record
}

func persistTerminalStructuredRunForTest(t *testing.T, store *memoryWorkflowRunStore, context WorkflowRunContext, record *WorkflowRunRecord) {
	t.Helper()
	if err := store.UpsertRun(context, record); err != nil {
		t.Fatalf("persist running run %s: %v", record.RunID, err)
	}
	record.Status = WorkflowRunStatusSucceeded
	record.CompletedAt = workflowRunTimestamp(time.Now().UTC())
	record.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err := store.UpsertRun(context, record); err != nil {
		t.Fatalf("persist terminal run %s: %v", record.RunID, err)
	}
}

func savedWorkflowDraftValidationHasFinding(summary SavedWorkflowDraftValidationSummary, code SavedWorkflowDraftFailureCode) bool {
	for _, finding := range summary.Findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func mustJSONForStructuredInputTest(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
