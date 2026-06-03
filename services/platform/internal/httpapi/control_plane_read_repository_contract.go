package httpapi

type ReadRepositoryContext struct {
	RequestID   string
	TenantRef   string
	SubjectRef  string
	ScopeGrants []string
	AuditRef    string
	IssuerRef   string
	SessionRef  string
}

type ReadRepositoryFilters map[string]string

type ReadRepositorySort string

type ReadRepositoryProjection string

type ReadRepositoryRequest struct {
	Limit      int
	Cursor     string
	Filters    ReadRepositoryFilters
	Sort       ReadRepositorySort
	Projection ReadRepositoryProjection
}

type ReadRepositoryFailureCode string

const (
	ReadRepositoryFailureTenantBindingMissing      ReadRepositoryFailureCode = "tenant_binding_missing"
	ReadRepositoryFailureScopeDenied               ReadRepositoryFailureCode = "scope_denied"
	ReadRepositoryFailureInvalidFilter             ReadRepositoryFailureCode = "invalid_filter"
	ReadRepositoryFailureStoreUnavailable          ReadRepositoryFailureCode = "read_store_unavailable"
	ReadRepositoryFailureContractMismatch          ReadRepositoryFailureCode = "read_store_contract_mismatch"
	ReadRepositoryFailureDatabaseReadDisabled      ReadRepositoryFailureCode = "database_read_disabled"
	ReadRepositoryFailureAuthContextMismatch       ReadRepositoryFailureCode = "auth_context_contract_mismatch"
	ReadRepositoryFailureSchemaMigrationNotApplied ReadRepositoryFailureCode = "schema_migration_not_applied"
	ReadRepositoryFailureSchemaVersionMismatch     ReadRepositoryFailureCode = "schema_version_mismatch"
)

type ReadRepositoryResult[T any] struct {
	TenantRef   string
	Items       []T
	NextCursor  *string
	FailureCode ReadRepositoryFailureCode
	AuditRef    string
}

type TenantSummary struct {
	TenantRef         string
	TenantDisplayName string
	State             string
	Plan              string
	UserCount         int
	ApplicationCount  int
	AuditRef          string
}

type ApplicationSummary struct {
	ApplicationRef  string
	ApplicationKind string
	DisplayName     string
	OwnerSubjectRef string
	Status          string
	LastRunStatus   string
	UpdatedAt       string
}

type APIKeySummary struct {
	APIKeyID        string
	OwnerSubjectRef string
	Scope           string
	State           string
	CreatedAt       string
	ExpiresAt       string
}

type QuotaSummary struct {
	QuotaID          string
	Period           string
	RequestLimit     int
	TokenLimit       int
	CostLimit        string
	RequestUsage     int
	TokenUsage       int
	CostUsage        string
	OverQuotaFailure ReadRepositoryFailureCode
}

type WorkflowDefinitionSummary struct {
	WorkflowDefinitionID        string
	ApplicationRef              string
	Version                     string
	DefinitionStatus            string
	NodeCount                   int
	RiskLevel                   string
	RequiresConfirmationCapable bool
	UpdatedAt                   string
}

type RunRecordSummary struct {
	RunID                 string
	WorkflowDefinitionRef string
	ApplicationRef        string
	Status                string
	FailureCode           ReadRepositoryFailureCode
	CostSummary           string
	TraceID               string
	StartedAt             string
	CompletedAt           string
}

type AuditSummary struct {
	AuditRef        string
	ActorSubjectRef string
	EventKind       string
	ResourceKind    string
	ResourceRef     string
	Decision        string
	FailureCode     ReadRepositoryFailureCode
	TraceID         string
	RecordedAt      string
}

type ReadTenantSummaryRequest struct {
	ReadRepositoryRequest
}

type ListApplicationSummariesRequest struct {
	ReadRepositoryRequest
}

type ListAPIKeySummariesRequest struct {
	ReadRepositoryRequest
}

type ReadQuotaSummaryRequest struct {
	ReadRepositoryRequest
}

type ListWorkflowDefinitionSummariesRequest struct {
	ReadRepositoryRequest
}

type ListRunRecordSummariesRequest struct {
	ReadRepositoryRequest
}

type ListAuditSummariesRequest struct {
	ReadRepositoryRequest
}

type ReadTenantSummaryResult = ReadRepositoryResult[TenantSummary]
type ListApplicationSummariesResult = ReadRepositoryResult[ApplicationSummary]
type ListAPIKeySummariesResult = ReadRepositoryResult[APIKeySummary]
type ReadQuotaSummaryResult = ReadRepositoryResult[QuotaSummary]
type ListWorkflowDefinitionSummariesResult = ReadRepositoryResult[WorkflowDefinitionSummary]
type ListRunRecordSummariesResult = ReadRepositoryResult[RunRecordSummary]
type ListAuditSummariesResult = ReadRepositoryResult[AuditSummary]

type ReadRepositoryRouteTypeContract struct {
	RouteID        string
	Operation      string
	Mode           string
	RequestType    string
	ResultType     string
	SummaryType    string
	ProjectionType string
	AllowedFilters []string
	AllowedSort    []string
}

func controlPlaneReadRepositoryRouteTypeContracts() []ReadRepositoryRouteTypeContract {
	return []ReadRepositoryRouteTypeContract{
		{
			RouteID:        "tenant-summary-route",
			Operation:      "ReadTenantSummary",
			Mode:           "single_resource_read",
			RequestType:    "ReadTenantSummaryRequest",
			ResultType:     "ReadTenantSummaryResult",
			SummaryType:    "TenantSummary",
			ProjectionType: "TenantSummaryProjection",
		},
		{
			RouteID:        "application-summary-list-route",
			Operation:      "ListApplicationSummaries",
			Mode:           "cursor_list_read",
			RequestType:    "ListApplicationSummariesRequest",
			ResultType:     "ListApplicationSummariesResult",
			SummaryType:    "ApplicationSummary",
			ProjectionType: "ApplicationSummaryProjection",
			AllowedFilters: []string{"status", "kind"},
			AllowedSort:    []string{"updated_at_desc"},
		},
		{
			RouteID:        "api-key-summary-list-route",
			Operation:      "ListAPIKeySummaries",
			Mode:           "cursor_list_read",
			RequestType:    "ListAPIKeySummariesRequest",
			ResultType:     "ListAPIKeySummariesResult",
			SummaryType:    "APIKeySummary",
			ProjectionType: "APIKeySummaryProjection",
			AllowedFilters: []string{"state", "owner_subject_ref"},
			AllowedSort:    []string{"created_at_desc"},
		},
		{
			RouteID:        "quota-summary-route",
			Operation:      "ReadQuotaSummary",
			Mode:           "single_resource_read",
			RequestType:    "ReadQuotaSummaryRequest",
			ResultType:     "ReadQuotaSummaryResult",
			SummaryType:    "QuotaSummary",
			ProjectionType: "QuotaSummaryProjection",
			AllowedFilters: []string{"period"},
		},
		{
			RouteID:        "workflow-definition-summary-list-route",
			Operation:      "ListWorkflowDefinitionSummaries",
			Mode:           "cursor_list_read",
			RequestType:    "ListWorkflowDefinitionSummariesRequest",
			ResultType:     "ListWorkflowDefinitionSummariesResult",
			SummaryType:    "WorkflowDefinitionSummary",
			ProjectionType: "WorkflowDefinitionSummaryProjection",
			AllowedFilters: []string{"application_ref", "definition_status", "risk_level"},
			AllowedSort:    []string{"updated_at_desc"},
		},
		{
			RouteID:        "run-record-summary-list-route",
			Operation:      "ListRunRecordSummaries",
			Mode:           "cursor_list_read",
			RequestType:    "ListRunRecordSummariesRequest",
			ResultType:     "ListRunRecordSummariesResult",
			SummaryType:    "RunRecordSummary",
			ProjectionType: "RunRecordSummaryProjection",
			AllowedFilters: []string{"application_ref", "workflow_definition_ref", "status", "failure_code"},
			AllowedSort:    []string{"started_at_desc"},
		},
		{
			RouteID:        "audit-summary-list-route",
			Operation:      "ListAuditSummaries",
			Mode:           "cursor_list_read",
			RequestType:    "ListAuditSummariesRequest",
			ResultType:     "ListAuditSummariesResult",
			SummaryType:    "AuditSummary",
			ProjectionType: "AuditSummaryProjection",
			AllowedFilters: []string{"actor_subject_ref", "event_kind", "resource_kind", "decision", "failure_code"},
			AllowedSort:    []string{"recorded_at_desc"},
		},
	}
}
