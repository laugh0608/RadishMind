package httpapi

import "context"

type ReadRepositoryContext struct {
	RequestContext context.Context
	RequestID      string
	TenantRef      string
	SubjectRef     string
	ScopeGrants    []string
	AuditRef       string
	IssuerRef      string
	SessionRef     string
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
	TenantRef           string
	TenantDisplayName   string
	TenantState         string
	PlanRef             string
	QuotaSummaryRef     string
	DeploymentStatusRef string
	AuditSummaryRef     string
}

type ApplicationSummary struct {
	ApplicationRef              string
	TenantRef                   string
	ApplicationKind             string
	DisplayName                 string
	OwnerSubjectRef             string
	LatestWorkflowDefinitionRef string
	LastRunStatus               string
	UpdatedAt                   string
}

type APIKeySummary struct {
	APIKeyID        string
	TenantRef       string
	OwnerSubjectRef string
	Scopes          []string
	State           string
	CreatedAt       string
	ExpiresAt       *string
	LastUsedAt      *string
}

type QuotaSummary struct {
	QuotaID              string
	TenantRef            string
	Period               string
	RequestLimit         int
	TokenLimit           int
	CostLimit            float64
	UsageSnapshot        map[string]any
	OverQuotaFailureCode string
}

type WorkflowDefinitionSummary struct {
	WorkflowDefinitionID        string
	TenantRef                   string
	ApplicationRef              string
	Version                     int
	DefinitionStatus            string
	NodeCount                   int
	RiskLevel                   string
	RequiresConfirmationCapable bool
	UpdatedAt                   string
}

type RunRecordSummary struct {
	RunID                string
	TenantRef            string
	WorkflowDefinitionID string
	ApplicationRef       string
	Status               string
	FailureCode          *ReadRepositoryFailureCode
	CostSummary          map[string]any
	TraceID              string
	StartedAt            string
	CompletedAt          string
}

type AuditSummary struct {
	AuditRef        string
	TenantRef       string
	ActorSubjectRef string
	EventKind       string
	ResourceRef     string
	Decision        string
	FailureCode     *ReadRepositoryFailureCode
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
			AllowedFilters: []string{"event_kind", "resource_ref", "actor_subject_ref", "failure_code"},
			AllowedSort:    []string{"recorded_at_desc"},
		},
	}
}
