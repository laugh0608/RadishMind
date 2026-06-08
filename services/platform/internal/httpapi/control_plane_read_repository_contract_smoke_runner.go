package httpapi

type ReadRepositoryContractSmokeRequest struct {
	Limit      *int
	Cursor     *string
	Filters    map[string]*string
	Sort       *ReadRepositorySort
	Projection ReadRepositoryProjection
}

type ReadRepositoryContractSmokeCase struct {
	RouteID                 string
	Method                  string
	Path                    string
	ReadModel               string
	RequiredScope           string
	Operation               string
	ReadMode                string
	Request                 ReadRepositoryContractSmokeRequest
	ExpectedItemsKind       string
	DatabaseModeFailureCode ReadRepositoryFailureCode
	FakeFallbackAllowed     bool
	SideEffectAllowed       bool
}

type ReadRepositoryContractSmokeFailureExpectation struct {
	FailureCode ReadRepositoryFailureCode
	Expectation string
}

type ReadRepositoryContractSmokeRouteResult struct {
	RouteID     string
	Operation   string
	RequestType string
	ResultType  string
	Passed      bool
	FailureCode ReadRepositoryFailureCode
	Message     string
}

type ReadRepositoryContractSmokeSideEffectReport struct {
	RepositoryWriteCount   int
	ExecutorCallCount      int
	ConfirmationCallCount  int
	BusinessWritebackCount int
	ReplayCallCount        int
	ForbiddenSideEffects   []string
}

type ReadRepositoryContractSmokeSummary struct {
	TotalRoutes         int
	PassedRoutes        int
	FailedRoutes        int
	FailureMappingCount int
	Passed              bool
}

type ReadRepositoryContractSmokeReport struct {
	RouteResults           []ReadRepositoryContractSmokeRouteResult
	FailureResults         []ReadRepositoryContractSmokeFailureExpectation
	ContractMismatchReport []string
	SideEffectReport       ReadRepositoryContractSmokeSideEffectReport
	Summary                ReadRepositoryContractSmokeSummary
}

type ControlPlaneReadRepositoryContractSmokeRunner struct {
	typeContracts []ReadRepositoryRouteTypeContract
	smokeCases    []ReadRepositoryContractSmokeCase
}

func NewReadRepositoryContractSmokeRunner() ControlPlaneReadRepositoryContractSmokeRunner {
	return ControlPlaneReadRepositoryContractSmokeRunner{
		typeContracts: controlPlaneReadRepositoryRouteTypeContracts(),
		smokeCases:    controlPlaneReadRepositoryContractSmokeCases(),
	}
}

func RunControlPlaneReadRepositoryContractSmoke(context ReadRepositoryContext) ReadRepositoryContractSmokeReport {
	return NewReadRepositoryContractSmokeRunner().Run(context)
}

func (runner ControlPlaneReadRepositoryContractSmokeRunner) Run(context ReadRepositoryContext) ReadRepositoryContractSmokeReport {
	report := ReadRepositoryContractSmokeReport{
		FailureResults:   controlPlaneReadRepositoryContractSmokeFailureExpectations(),
		SideEffectReport: controlPlaneReadRepositoryContractSmokeSideEffectReport(),
	}
	typeContracts, mismatch := readRepositoryRouteTypeContractsByRoute(runner.typeContracts)
	report.ContractMismatchReport = append(report.ContractMismatchReport, mismatch...)
	contextFailure := readRepositoryContractSmokeContextFailure(context)

	for _, smokeCase := range runner.smokeCases {
		result := ReadRepositoryContractSmokeRouteResult{
			RouteID:   smokeCase.RouteID,
			Operation: smokeCase.Operation,
			Passed:    true,
		}
		if contextFailure != "" {
			result.Passed = false
			result.FailureCode = contextFailure
			result.Message = "repository context missing required smoke field"
		} else if typeContract, found := typeContracts[smokeCase.RouteID]; !found {
			result.Passed = false
			result.FailureCode = ReadRepositoryFailureContractMismatch
			result.Message = "route is missing from repository type catalog"
		} else {
			result.RequestType = typeContract.RequestType
			result.ResultType = typeContract.ResultType
			if failure := validateReadRepositoryContractSmokeCase(smokeCase, typeContract); failure != "" {
				result.Passed = false
				result.FailureCode = failure
				result.Message = "route smoke case does not match repository type catalog"
			}
		}
		report.RouteResults = append(report.RouteResults, result)
	}

	report.Summary.TotalRoutes = len(report.RouteResults)
	report.Summary.FailureMappingCount = len(report.FailureResults)
	for _, result := range report.RouteResults {
		if result.Passed {
			report.Summary.PassedRoutes++
		}
	}
	report.Summary.FailedRoutes = report.Summary.TotalRoutes - report.Summary.PassedRoutes
	report.Summary.Passed = report.Summary.TotalRoutes == 7 &&
		report.Summary.FailedRoutes == 0 &&
		len(report.ContractMismatchReport) == 0 &&
		report.SideEffectReport.RepositoryWriteCount == 0 &&
		report.SideEffectReport.ExecutorCallCount == 0 &&
		report.SideEffectReport.ConfirmationCallCount == 0 &&
		report.SideEffectReport.BusinessWritebackCount == 0 &&
		report.SideEffectReport.ReplayCallCount == 0
	return report
}

func readRepositoryRouteTypeContractsByRoute(contracts []ReadRepositoryRouteTypeContract) (map[string]ReadRepositoryRouteTypeContract, []string) {
	byRoute := make(map[string]ReadRepositoryRouteTypeContract, len(contracts))
	var mismatch []string
	for _, contract := range contracts {
		if contract.RouteID == "" || contract.Operation == "" || contract.RequestType == "" || contract.ResultType == "" {
			mismatch = append(mismatch, "route type contract missing required identifiers")
			continue
		}
		if _, exists := byRoute[contract.RouteID]; exists {
			mismatch = append(mismatch, "duplicate route type contract: "+contract.RouteID)
			continue
		}
		byRoute[contract.RouteID] = contract
	}
	return byRoute, mismatch
}

func readRepositoryContractSmokeContextFailure(context ReadRepositoryContext) ReadRepositoryFailureCode {
	if context.RequestID == "" || context.TenantRef == "" || context.SubjectRef == "" || context.AuditRef == "" {
		return ReadRepositoryFailureAuthContextMismatch
	}
	if len(context.ScopeGrants) == 0 {
		return ReadRepositoryFailureScopeDenied
	}
	return ""
}

func validateReadRepositoryContractSmokeCase(
	smokeCase ReadRepositoryContractSmokeCase,
	typeContract ReadRepositoryRouteTypeContract,
) ReadRepositoryFailureCode {
	if smokeCase.Operation != typeContract.Operation {
		return ReadRepositoryFailureContractMismatch
	}
	if smokeCase.ReadMode != "future_repository_contract" {
		return ReadRepositoryFailureContractMismatch
	}
	if smokeCase.DatabaseModeFailureCode != ReadRepositoryFailureDatabaseReadDisabled {
		return ReadRepositoryFailureContractMismatch
	}
	if smokeCase.FakeFallbackAllowed || smokeCase.SideEffectAllowed {
		return ReadRepositoryFailureContractMismatch
	}
	if smokeCase.Request.Projection == "" {
		return ReadRepositoryFailureContractMismatch
	}
	if !readRepositorySmokeFiltersAllowed(smokeCase.Request.Filters, typeContract.AllowedFilters) {
		return ReadRepositoryFailureInvalidFilter
	}
	if !readRepositorySmokeSortAllowed(smokeCase.Request.Sort, typeContract.AllowedSort) {
		return ReadRepositoryFailureInvalidFilter
	}
	return ""
}

func readRepositorySmokeFiltersAllowed(filters map[string]*string, allowed []string) bool {
	allowedSet := make(map[string]bool, len(allowed))
	for _, filter := range allowed {
		allowedSet[filter] = true
	}
	for filter := range filters {
		if !allowedSet[filter] {
			return false
		}
	}
	return true
}

func readRepositorySmokeSortAllowed(sort *ReadRepositorySort, allowed []string) bool {
	if sort == nil {
		return true
	}
	for _, allowedSort := range allowed {
		if string(*sort) == allowedSort {
			return true
		}
	}
	return false
}

func controlPlaneReadRepositoryContractSmokeCases() []ReadRepositoryContractSmokeCase {
	return []ReadRepositoryContractSmokeCase{
		{
			RouteID:                 "tenant-summary-route",
			Method:                  "GET",
			Path:                    "/v1/control-plane/tenants/{tenant_ref}/summary",
			ReadModel:               "tenant-summary",
			RequiredScope:           "tenant:read",
			Operation:               "ReadTenantSummary",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(nil, nil, nil, nil, "tenant_summary_projection"),
			ExpectedItemsKind:       "single_sanitized_summary",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "application-summary-list-route",
			Method:                  "GET",
			Path:                    "/v1/user-workspace/applications",
			ReadModel:               "application-summary",
			RequiredScope:           "applications:read",
			Operation:               "ListApplicationSummaries",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(readRepositoryInt(25), nil, readRepositorySmokeFilters("status", "active", "kind", "prompt_app"), readRepositorySort("updated_at_desc"), "application_summary_projection"),
			ExpectedItemsKind:       "sanitized_summary_list",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "api-key-summary-list-route",
			Method:                  "GET",
			Path:                    "/v1/user-workspace/api-keys",
			ReadModel:               "api-key-summary",
			RequiredScope:           "api_keys:read",
			Operation:               "ListAPIKeySummaries",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(readRepositoryInt(25), nil, readRepositorySmokeFilters("state", "active", "owner_subject_ref", "subject:demo"), readRepositorySort("created_at_desc"), "api_key_summary_projection"),
			ExpectedItemsKind:       "sanitized_summary_list",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "quota-summary-route",
			Method:                  "GET",
			Path:                    "/v1/user-workspace/usage/quota-summary",
			ReadModel:               "quota-summary",
			RequiredScope:           "usage:read",
			Operation:               "ReadQuotaSummary",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(nil, nil, readRepositorySmokeFilters("period", "month"), nil, "quota_summary_projection"),
			ExpectedItemsKind:       "single_sanitized_summary",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "workflow-definition-summary-list-route",
			Method:                  "GET",
			Path:                    "/v1/user-workspace/workflow-definitions",
			ReadModel:               "workflow-definition-summary",
			RequiredScope:           "applications:read",
			Operation:               "ListWorkflowDefinitionSummaries",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(readRepositoryInt(25), nil, readRepositorySmokeFilters("application_ref", "app:demo", "definition_status", "published", "risk_level", "low"), readRepositorySort("updated_at_desc"), "workflow_definition_summary_projection"),
			ExpectedItemsKind:       "sanitized_summary_list",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "run-record-summary-list-route",
			Method:                  "GET",
			Path:                    "/v1/user-workspace/runs",
			ReadModel:               "run-record-summary",
			RequiredScope:           "runs:read",
			Operation:               "ListRunRecordSummaries",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(readRepositoryInt(25), nil, readRepositorySmokeFilters("application_ref", "app:demo", "workflow_definition_ref", "workflow:def:demo", "status", "completed", "failure_code", ""), readRepositorySort("started_at_desc"), "run_record_summary_projection"),
			ExpectedItemsKind:       "sanitized_summary_list",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
		{
			RouteID:                 "audit-summary-list-route",
			Method:                  "GET",
			Path:                    "/v1/control-plane/audit",
			ReadModel:               "audit-summary",
			RequiredScope:           "audit:read",
			Operation:               "ListAuditSummaries",
			ReadMode:                "future_repository_contract",
			Request:                 readRepositoryContractSmokeRequest(readRepositoryInt(25), nil, readRepositorySmokeFilters("actor_subject_ref", "subject:demo", "event_kind", "read", "resource_kind", "application", "decision", "allow", "failure_code", ""), readRepositorySort("recorded_at_desc"), "audit_summary_projection"),
			ExpectedItemsKind:       "sanitized_summary_list",
			DatabaseModeFailureCode: ReadRepositoryFailureDatabaseReadDisabled,
		},
	}
}

func controlPlaneReadRepositoryContractSmokeFailureExpectations() []ReadRepositoryContractSmokeFailureExpectation {
	return []ReadRepositoryContractSmokeFailureExpectation{
		{FailureCode: ReadRepositoryFailureTenantBindingMissing, Expectation: "must fail closed before repository result comparison"},
		{FailureCode: ReadRepositoryFailureScopeDenied, Expectation: "must fail closed before repository result comparison"},
		{FailureCode: ReadRepositoryFailureInvalidFilter, Expectation: "must reject filters outside the route type allowlist"},
		{FailureCode: ReadRepositoryFailureStoreUnavailable, Expectation: "must not fall back to fixture fake store"},
		{FailureCode: ReadRepositoryFailureContractMismatch, Expectation: "must report contract mismatch without database detail"},
		{FailureCode: ReadRepositoryFailureDatabaseReadDisabled, Expectation: "must preserve disabled database guard"},
		{FailureCode: ReadRepositoryFailureAuthContextMismatch, Expectation: "must reject malformed repository context"},
		{FailureCode: ReadRepositoryFailureSchemaMigrationNotApplied, Expectation: "must block repository smoke when schema gate is absent"},
		{FailureCode: ReadRepositoryFailureSchemaVersionMismatch, Expectation: "must block repository smoke when schema version drifts"},
	}
}

func controlPlaneReadRepositoryContractSmokeSideEffectReport() ReadRepositoryContractSmokeSideEffectReport {
	return ReadRepositoryContractSmokeSideEffectReport{
		ForbiddenSideEffects: []string{
			"database_write",
			"api_key_issue",
			"quota_mutation",
			"workflow_execution",
			"confirmation_decision",
			"business_writeback",
			"replay_execution",
		},
	}
}

func readRepositoryContractSmokeRequest(
	limit *int,
	cursor *string,
	filters map[string]*string,
	sort *ReadRepositorySort,
	projection string,
) ReadRepositoryContractSmokeRequest {
	if filters == nil {
		filters = map[string]*string{}
	}
	return ReadRepositoryContractSmokeRequest{
		Limit:      limit,
		Cursor:     cursor,
		Filters:    filters,
		Sort:       sort,
		Projection: ReadRepositoryProjection(projection),
	}
}

func readRepositorySmokeFilters(pairs ...string) map[string]*string {
	filters := make(map[string]*string, len(pairs)/2)
	for index := 0; index < len(pairs); index += 2 {
		value := pairs[index+1]
		if value == "" {
			filters[pairs[index]] = nil
			continue
		}
		filters[pairs[index]] = &value
	}
	return filters
}

func readRepositoryInt(value int) *int {
	return &value
}

func readRepositorySort(value string) *ReadRepositorySort {
	sort := ReadRepositorySort(value)
	return &sort
}
