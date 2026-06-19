package httpapi

type SavedWorkflowDraftRepositoryActorContext struct {
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	ApplicationID   string
	ActorSubjectRef string
	OwnerSubjectRef string
	ScopeGrants     []string
	AuditRef        string
}

type SavedWorkflowDraftRepositoryContractSmokeFailureCode string

const (
	SavedWorkflowDraftRepositorySmokeFailureScopeDenied               SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_scope_denied"
	SavedWorkflowDraftRepositorySmokeFailureNotFound                  SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_not_found"
	SavedWorkflowDraftRepositorySmokeFailureSchemaVersionUnsupported  SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_schema_version_unsupported"
	SavedWorkflowDraftRepositorySmokeFailurePayloadInvalid            SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_payload_invalid"
	SavedWorkflowDraftRepositorySmokeFailureVersionConflict           SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_version_conflict"
	SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable          SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_store_unavailable"
	SavedWorkflowDraftRepositorySmokeFailureContractMismatch          SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_store_contract_mismatch"
	SavedWorkflowDraftRepositorySmokeFailureRepositoryStoreDisabled   SavedWorkflowDraftRepositoryContractSmokeFailureCode = "repository_store_disabled"
	SavedWorkflowDraftRepositorySmokeFailureInvalidStoreMode          SavedWorkflowDraftRepositoryContractSmokeFailureCode = "invalid_draft_store_mode"
	SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch       SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_auth_context_contract_mismatch"
	SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing    SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_schema_migration_not_applied"
	SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch       SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_store_schema_version_mismatch"
	SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable SavedWorkflowDraftRepositoryContractSmokeFailureCode = "draft_store_migration_unavailable"
)

type SavedWorkflowDraftRepositoryOperationContract struct {
	Operation     string
	RequiredScope string
	RequestType   string
	ResultType    string
}

type SavedWorkflowDraftRepositoryContractSmokeCase struct {
	Operation                                string
	RequiredScope                            string
	RequestType                              string
	ResultType                               string
	ExpectedSuccessProjection                string
	RequiredFailureCodes                     []SavedWorkflowDraftRepositoryContractSmokeFailureCode
	SmokeMode                                string
	UsesRepositoryOperationContract          bool
	UsesContractSmokeFixture                 bool
	UsesAuthContextContract                  bool
	UsesSchemaArtifactGate                   bool
	UsesSelectorSmokeGate                    bool
	RepositoryContractSmokeRunnerImplemented bool
	RepositoryAdapterDependencySatisfied     bool
	FallbackToMemoryDevStoreAllowed          bool
	FallbackToSampleAllowed                  bool
	FallbackToFixtureAllowed                 bool
	FallbackToDevHTTPRouteAllowed            bool
	SideEffectAllowed                        bool
}

type SavedWorkflowDraftRepositoryContractSmokeFailureExpectation struct {
	FailureCode SavedWorkflowDraftRepositoryContractSmokeFailureCode
	Expectation string
}

type SavedWorkflowDraftRepositoryContractSmokeOperationResult struct {
	Operation   string
	RequestType string
	ResultType  string
	Passed      bool
	FailureCode SavedWorkflowDraftRepositoryContractSmokeFailureCode
	Message     string
}

type SavedWorkflowDraftRepositoryContractSmokeSideEffectReport struct {
	RepositoryWriteCount   int
	ExecutorCallCount      int
	ConfirmationCallCount  int
	BusinessWritebackCount int
	ReplayCallCount        int
	ForbiddenSideEffects   []string
}

type SavedWorkflowDraftRepositoryContractSmokeSummary struct {
	TotalOperations     int
	PassedOperations    int
	FailedOperations    int
	FailureMappingCount int
	Passed              bool
}

type SavedWorkflowDraftRepositoryContractSmokeReport struct {
	OperationResults       []SavedWorkflowDraftRepositoryContractSmokeOperationResult
	FailureResults         []SavedWorkflowDraftRepositoryContractSmokeFailureExpectation
	ContractMismatchReport []string
	FallbackReport         []string
	SideEffectReport       SavedWorkflowDraftRepositoryContractSmokeSideEffectReport
	Summary                SavedWorkflowDraftRepositoryContractSmokeSummary
}

type SavedWorkflowDraftRepositoryContractSmokeRunner struct {
	operationContracts []SavedWorkflowDraftRepositoryOperationContract
	smokeCases         []SavedWorkflowDraftRepositoryContractSmokeCase
}

func NewSavedWorkflowDraftRepositoryContractSmokeRunner() SavedWorkflowDraftRepositoryContractSmokeRunner {
	return SavedWorkflowDraftRepositoryContractSmokeRunner{
		operationContracts: savedWorkflowDraftRepositoryOperationContracts(),
		smokeCases:         savedWorkflowDraftRepositoryContractSmokeCases(),
	}
}

func RunSavedWorkflowDraftRepositoryContractSmoke(
	context SavedWorkflowDraftRepositoryActorContext,
) SavedWorkflowDraftRepositoryContractSmokeReport {
	return NewSavedWorkflowDraftRepositoryContractSmokeRunner().Run(context)
}

func (runner SavedWorkflowDraftRepositoryContractSmokeRunner) Run(
	context SavedWorkflowDraftRepositoryActorContext,
) SavedWorkflowDraftRepositoryContractSmokeReport {
	report := SavedWorkflowDraftRepositoryContractSmokeReport{
		FailureResults:   savedWorkflowDraftRepositoryContractSmokeFailureExpectations(),
		SideEffectReport: savedWorkflowDraftRepositoryContractSmokeSideEffectReport(),
	}
	operationContracts, mismatch := savedWorkflowDraftRepositoryOperationContractsByOperation(runner.operationContracts)
	report.ContractMismatchReport = append(report.ContractMismatchReport, mismatch...)
	contextFailure := savedWorkflowDraftRepositoryContractSmokeContextFailure(context)
	scopeGrants := savedWorkflowDraftRepositorySmokeScopeGrants(context.ScopeGrants)

	for _, smokeCase := range runner.smokeCases {
		result := SavedWorkflowDraftRepositoryContractSmokeOperationResult{
			Operation: smokeCase.Operation,
			Passed:    true,
		}
		if contextFailure != "" {
			result.Passed = false
			result.FailureCode = contextFailure
			result.Message = "repository actor context missing required smoke field"
		} else if !scopeGrants[smokeCase.RequiredScope] {
			result.Passed = false
			result.FailureCode = SavedWorkflowDraftRepositorySmokeFailureScopeDenied
			result.Message = "repository actor context missing operation scope grant"
		} else if operationContract, found := operationContracts[smokeCase.Operation]; !found {
			result.Passed = false
			result.FailureCode = SavedWorkflowDraftRepositorySmokeFailureContractMismatch
			result.Message = "operation is missing from saved draft repository operation contract"
		} else {
			result.RequestType = operationContract.RequestType
			result.ResultType = operationContract.ResultType
			if failure := validateSavedWorkflowDraftRepositoryContractSmokeCase(smokeCase, operationContract); failure != "" {
				result.Passed = false
				result.FailureCode = failure
				result.Message = "operation smoke case does not match repository operation contract"
			}
		}
		report.OperationResults = append(report.OperationResults, result)
	}

	report.Summary.TotalOperations = len(report.OperationResults)
	report.Summary.FailureMappingCount = len(report.FailureResults)
	for _, result := range report.OperationResults {
		if result.Passed {
			report.Summary.PassedOperations++
		}
	}
	report.Summary.FailedOperations = report.Summary.TotalOperations - report.Summary.PassedOperations
	report.Summary.Passed = report.Summary.TotalOperations == 3 &&
		report.Summary.FailedOperations == 0 &&
		len(report.ContractMismatchReport) == 0 &&
		len(report.FallbackReport) == 0 &&
		report.SideEffectReport.RepositoryWriteCount == 0 &&
		report.SideEffectReport.ExecutorCallCount == 0 &&
		report.SideEffectReport.ConfirmationCallCount == 0 &&
		report.SideEffectReport.BusinessWritebackCount == 0 &&
		report.SideEffectReport.ReplayCallCount == 0
	return report
}

func savedWorkflowDraftRepositoryOperationContracts() []SavedWorkflowDraftRepositoryOperationContract {
	return []SavedWorkflowDraftRepositoryOperationContract{
		{
			Operation:     "SaveWorkflowDraftRecord",
			RequiredScope: "workflow_drafts:write",
			RequestType:   "SaveWorkflowDraftRecordRequest",
			ResultType:    "SaveWorkflowDraftRecordResult",
		},
		{
			Operation:     "ReadWorkflowDraftRecord",
			RequiredScope: "workflow_drafts:read",
			RequestType:   "ReadWorkflowDraftRecordRequest",
			ResultType:    "ReadWorkflowDraftRecordResult",
		},
		{
			Operation:     "ListWorkflowDraftRecords",
			RequiredScope: "workflow_drafts:read",
			RequestType:   "ListWorkflowDraftRecordsRequest",
			ResultType:    "ListWorkflowDraftRecordsResult",
		},
	}
}

func savedWorkflowDraftRepositoryOperationContractsByOperation(
	contracts []SavedWorkflowDraftRepositoryOperationContract,
) (map[string]SavedWorkflowDraftRepositoryOperationContract, []string) {
	byOperation := make(map[string]SavedWorkflowDraftRepositoryOperationContract, len(contracts))
	var mismatch []string
	for _, contract := range contracts {
		if contract.Operation == "" || contract.RequiredScope == "" || contract.RequestType == "" || contract.ResultType == "" {
			mismatch = append(mismatch, "saved draft repository operation contract missing required identifiers")
			continue
		}
		if _, exists := byOperation[contract.Operation]; exists {
			mismatch = append(mismatch, "duplicate saved draft repository operation contract: "+contract.Operation)
			continue
		}
		byOperation[contract.Operation] = contract
	}
	return byOperation, mismatch
}

func savedWorkflowDraftRepositoryContractSmokeContextFailure(
	context SavedWorkflowDraftRepositoryActorContext,
) SavedWorkflowDraftRepositoryContractSmokeFailureCode {
	if context.RequestID == "" ||
		context.TenantRef == "" ||
		context.WorkspaceID == "" ||
		context.ApplicationID == "" ||
		context.ActorSubjectRef == "" ||
		context.OwnerSubjectRef == "" ||
		context.AuditRef == "" {
		return SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch
	}
	if len(context.ScopeGrants) == 0 {
		return SavedWorkflowDraftRepositorySmokeFailureScopeDenied
	}
	return ""
}

func validateSavedWorkflowDraftRepositoryContractSmokeCase(
	smokeCase SavedWorkflowDraftRepositoryContractSmokeCase,
	operationContract SavedWorkflowDraftRepositoryOperationContract,
) SavedWorkflowDraftRepositoryContractSmokeFailureCode {
	if smokeCase.Operation != operationContract.Operation ||
		smokeCase.RequiredScope != operationContract.RequiredScope ||
		smokeCase.RequestType != operationContract.RequestType ||
		smokeCase.ResultType != operationContract.ResultType {
		return SavedWorkflowDraftRepositorySmokeFailureContractMismatch
	}
	if smokeCase.SmokeMode != "future_repository_contract" || smokeCase.ExpectedSuccessProjection == "" {
		return SavedWorkflowDraftRepositorySmokeFailureContractMismatch
	}
	if !smokeCase.UsesRepositoryOperationContract ||
		!smokeCase.UsesContractSmokeFixture ||
		!smokeCase.UsesAuthContextContract ||
		!smokeCase.UsesSchemaArtifactGate ||
		!smokeCase.UsesSelectorSmokeGate {
		return SavedWorkflowDraftRepositorySmokeFailureContractMismatch
	}
	if smokeCase.RepositoryContractSmokeRunnerImplemented ||
		smokeCase.RepositoryAdapterDependencySatisfied ||
		smokeCase.FallbackToMemoryDevStoreAllowed ||
		smokeCase.FallbackToSampleAllowed ||
		smokeCase.FallbackToFixtureAllowed ||
		smokeCase.FallbackToDevHTTPRouteAllowed ||
		smokeCase.SideEffectAllowed {
		return SavedWorkflowDraftRepositorySmokeFailureContractMismatch
	}
	if !savedWorkflowDraftRepositorySmokeFailureSetContains(
		smokeCase.RequiredFailureCodes,
		SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable,
		SavedWorkflowDraftRepositorySmokeFailureContractMismatch,
		SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch,
		SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing,
		SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch,
		SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable,
	) {
		return SavedWorkflowDraftRepositorySmokeFailureContractMismatch
	}
	return ""
}

func savedWorkflowDraftRepositoryContractSmokeCases() []SavedWorkflowDraftRepositoryContractSmokeCase {
	return []SavedWorkflowDraftRepositoryContractSmokeCase{
		savedWorkflowDraftRepositoryContractSmokeCase(
			"SaveWorkflowDraftRecord",
			"workflow_drafts:write",
			"SaveWorkflowDraftRecordRequest",
			"SaveWorkflowDraftRecordResult",
			"sanitized_saved_draft_record_with_current_version_metadata",
			[]SavedWorkflowDraftRepositoryContractSmokeFailureCode{
				SavedWorkflowDraftRepositorySmokeFailureScopeDenied,
				SavedWorkflowDraftRepositorySmokeFailureSchemaVersionUnsupported,
				SavedWorkflowDraftRepositorySmokeFailurePayloadInvalid,
				SavedWorkflowDraftRepositorySmokeFailureVersionConflict,
				SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable,
				SavedWorkflowDraftRepositorySmokeFailureContractMismatch,
				SavedWorkflowDraftRepositorySmokeFailureRepositoryStoreDisabled,
				SavedWorkflowDraftRepositorySmokeFailureInvalidStoreMode,
				SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch,
				SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing,
				SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch,
				SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable,
			},
		),
		savedWorkflowDraftRepositoryContractSmokeCase(
			"ReadWorkflowDraftRecord",
			"workflow_drafts:read",
			"ReadWorkflowDraftRecordRequest",
			"ReadWorkflowDraftRecordResult",
			"sanitized_saved_draft_record",
			[]SavedWorkflowDraftRepositoryContractSmokeFailureCode{
				SavedWorkflowDraftRepositorySmokeFailureScopeDenied,
				SavedWorkflowDraftRepositorySmokeFailureNotFound,
				SavedWorkflowDraftRepositorySmokeFailureSchemaVersionUnsupported,
				SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable,
				SavedWorkflowDraftRepositorySmokeFailureContractMismatch,
				SavedWorkflowDraftRepositorySmokeFailureRepositoryStoreDisabled,
				SavedWorkflowDraftRepositorySmokeFailureInvalidStoreMode,
				SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch,
				SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing,
				SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch,
				SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable,
			},
		),
		savedWorkflowDraftRepositoryContractSmokeCase(
			"ListWorkflowDraftRecords",
			"workflow_drafts:read",
			"ListWorkflowDraftRecordsRequest",
			"ListWorkflowDraftRecordsResult",
			"sanitized_saved_draft_summary_list",
			[]SavedWorkflowDraftRepositoryContractSmokeFailureCode{
				SavedWorkflowDraftRepositorySmokeFailureScopeDenied,
				SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable,
				SavedWorkflowDraftRepositorySmokeFailureContractMismatch,
				SavedWorkflowDraftRepositorySmokeFailureRepositoryStoreDisabled,
				SavedWorkflowDraftRepositorySmokeFailureInvalidStoreMode,
				SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch,
				SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing,
				SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch,
				SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable,
			},
		),
	}
}

func savedWorkflowDraftRepositoryContractSmokeCase(
	operation string,
	requiredScope string,
	requestType string,
	resultType string,
	expectedSuccessProjection string,
	requiredFailureCodes []SavedWorkflowDraftRepositoryContractSmokeFailureCode,
) SavedWorkflowDraftRepositoryContractSmokeCase {
	return SavedWorkflowDraftRepositoryContractSmokeCase{
		Operation:                       operation,
		RequiredScope:                   requiredScope,
		RequestType:                     requestType,
		ResultType:                      resultType,
		ExpectedSuccessProjection:       expectedSuccessProjection,
		RequiredFailureCodes:            requiredFailureCodes,
		SmokeMode:                       "future_repository_contract",
		UsesRepositoryOperationContract: true,
		UsesContractSmokeFixture:        true,
		UsesAuthContextContract:         true,
		UsesSchemaArtifactGate:          true,
		UsesSelectorSmokeGate:           true,
	}
}

func savedWorkflowDraftRepositoryContractSmokeFailureExpectations() []SavedWorkflowDraftRepositoryContractSmokeFailureExpectation {
	return []SavedWorkflowDraftRepositoryContractSmokeFailureExpectation{
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureScopeDenied, Expectation: "must fail closed before repository result comparison"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureNotFound, Expectation: "must not fall back to sample or dev route"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureSchemaVersionUnsupported, Expectation: "must reject unsupported schema versions"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailurePayloadInvalid, Expectation: "must reject invalid saved draft payloads"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureVersionConflict, Expectation: "must preserve current version metadata without overwrite"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureStoreUnavailable, Expectation: "must not fall back to memory dev store"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureContractMismatch, Expectation: "must report contract mismatch without database detail"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureRepositoryStoreDisabled, Expectation: "must preserve disabled repository mode"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureInvalidStoreMode, Expectation: "must reject unknown store mode"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureAuthContextMismatch, Expectation: "must reject malformed repository actor context"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureSchemaMigrationMissing, Expectation: "must block repository smoke when schema gate is absent"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureStoreSchemaMismatch, Expectation: "must block repository smoke when schema version drifts"},
		{FailureCode: SavedWorkflowDraftRepositorySmokeFailureStoreMigrationUnavailable, Expectation: "must block repository smoke when migration evidence is unavailable"},
	}
}

func savedWorkflowDraftRepositoryContractSmokeSideEffectReport() SavedWorkflowDraftRepositoryContractSmokeSideEffectReport {
	return SavedWorkflowDraftRepositoryContractSmokeSideEffectReport{
		ForbiddenSideEffects: []string{
			"database_write",
			"workflow_execution",
			"confirmation_decision",
			"business_writeback",
			"replay_execution",
		},
	}
}

func savedWorkflowDraftRepositorySmokeScopeGrants(grants []string) map[string]bool {
	scopeGrants := make(map[string]bool, len(grants))
	for _, grant := range grants {
		scopeGrants[grant] = true
	}
	return scopeGrants
}

func savedWorkflowDraftRepositorySmokeFailureSetContains(
	codes []SavedWorkflowDraftRepositoryContractSmokeFailureCode,
	required ...SavedWorkflowDraftRepositoryContractSmokeFailureCode,
) bool {
	present := make(map[SavedWorkflowDraftRepositoryContractSmokeFailureCode]bool, len(codes))
	for _, code := range codes {
		present[code] = true
	}
	for _, code := range required {
		if !present[code] {
			return false
		}
	}
	return true
}
