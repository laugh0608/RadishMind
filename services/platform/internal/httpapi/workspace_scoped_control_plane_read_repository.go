package httpapi

import (
	"context"
	"strings"
)

type workspaceScopedControlPlaneReadRepository struct {
	ControlPlaneReadRepository
	applications applicationCatalogRepository
	apiKeys      apiKeyRepository
	definitions  workflowDefinitionReleaseRepository
	runs         workflowRunStore
}

func newWorkspaceScopedControlPlaneReadRepository(
	base ControlPlaneReadRepository,
	applications applicationCatalogRepository,
	apiKeys apiKeyRepository,
	definitions workflowDefinitionReleaseRepository,
	runs workflowRunStore,
) ControlPlaneReadRepository {
	return workspaceScopedControlPlaneReadRepository{
		ControlPlaneReadRepository: base,
		applications:               applications,
		apiKeys:                    apiKeys,
		definitions:                definitions,
		runs:                       runs,
	}
}

func (repository workspaceScopedControlPlaneReadRepository) ListApplicationSummaries(
	readContext ReadRepositoryContext,
	request ListApplicationSummariesRequest,
) ListApplicationSummariesResult {
	result := ListApplicationSummariesResult{
		TenantRef: readContext.TenantRef, Items: []ApplicationSummary{}, AuditRef: readContext.AuditRef,
	}
	if failure := validateWorkspaceReadRepositoryContext(readContext); failure != "" {
		result.FailureCode = failure
		return result
	}
	if repository.applications == nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	ownerFilter := strings.TrimSpace(request.Filters["owner_subject_ref"])
	if ownerFilter != "" && ownerFilter != readContext.SubjectRef {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	if strings.TrimSpace(request.Filters["last_run_status"]) != "" ||
		(request.Sort != "" && request.Sort != "updated_at_desc") {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	serviceResult := newApplicationCatalogService(repository.applications).List(
		applicationCatalogContextFromReadRepositoryContext(readContext),
		ApplicationCatalogListInput{
			ApplicationKind: strings.TrimSpace(request.Filters["application_kind"]),
			Limit:           request.Limit, Cursor: request.Cursor,
		},
	)
	if serviceResult.FailureCode != "" {
		result.FailureCode = translateApplicationCatalogReadFailure(serviceResult.FailureCode)
		return result
	}
	result.Items = make([]ApplicationSummary, 0, len(serviceResult.Records))
	for _, record := range serviceResult.Records {
		result.Items = append(result.Items, ApplicationSummary{
			ApplicationRef: record.ApplicationID, TenantRef: record.TenantRef,
			ApplicationKind: record.ApplicationKind, DisplayName: record.DisplayName,
			OwnerSubjectRef: record.OwnerSubjectRef, UpdatedAt: record.UpdatedAt,
		})
	}
	result.NextCursor = serviceResult.NextCursor
	return result
}

func (repository workspaceScopedControlPlaneReadRepository) ListAPIKeySummaries(
	readContext ReadRepositoryContext,
	request ListAPIKeySummariesRequest,
) ListAPIKeySummariesResult {
	result := ListAPIKeySummariesResult{
		TenantRef: readContext.TenantRef, Items: []APIKeySummary{}, AuditRef: readContext.AuditRef,
	}
	if failure := validateWorkspaceReadRepositoryContext(readContext); failure != "" {
		result.FailureCode = failure
		return result
	}
	if repository.apiKeys == nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	ownerFilter := strings.TrimSpace(request.Filters["owner_subject_ref"])
	if ownerFilter != "" && ownerFilter != readContext.SubjectRef {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	if request.Sort != "" && request.Sort != "created_at_desc" {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	serviceResult := newAPIKeyService(repository.apiKeys, repository.applications).List(
		apiKeyContextFromReadRepositoryContext(readContext),
		APIKeyListInput{
			EffectiveState: strings.TrimSpace(request.Filters["state"]),
			Limit:          request.Limit, Cursor: request.Cursor,
		},
	)
	if serviceResult.FailureCode != "" {
		result.FailureCode = translateAPIKeyReadFailure(serviceResult.FailureCode)
		return result
	}
	requiredScope := strings.TrimSpace(request.Filters["scope"])
	result.Items = make([]APIKeySummary, 0, len(serviceResult.Records))
	for _, record := range serviceResult.Records {
		if requiredScope != "" && !controlPlaneReadHasScope(record.Scopes, requiredScope) {
			continue
		}
		expiresAt := record.ExpiresAt
		result.Items = append(result.Items, APIKeySummary{
			APIKeyID: record.APIKeyID, TenantRef: record.TenantRef,
			OwnerSubjectRef: record.OwnerSubjectRef, Scopes: append([]string{}, record.Scopes...),
			State: record.EffectiveState, CreatedAt: record.CreatedAt,
			ExpiresAt: &expiresAt, LastUsedAt: record.LastUsedAt,
		})
	}
	result.NextCursor = serviceResult.NextCursor
	return result
}

func (repository workspaceScopedControlPlaneReadRepository) ReadQuotaSummary(
	readContext ReadRepositoryContext,
	_ ReadQuotaSummaryRequest,
) ReadQuotaSummaryResult {
	return ReadQuotaSummaryResult{
		TenantRef: readContext.TenantRef, Items: []QuotaSummary{},
		FailureCode: ReadRepositoryFailureCode("quota_policy_unavailable"), AuditRef: readContext.AuditRef,
	}
}

func (repository workspaceScopedControlPlaneReadRepository) ListWorkflowDefinitionSummaries(
	readContext ReadRepositoryContext,
	request ListWorkflowDefinitionSummariesRequest,
) ListWorkflowDefinitionSummariesResult {
	if failure := validateWorkspaceReadRepositoryContext(readContext); failure != "" {
		return workflowDefinitionSummaryFailure(readContext, failure)
	}
	if repository.definitions == nil {
		return workflowDefinitionSummaryFailure(readContext, ReadRepositoryFailureStoreUnavailable)
	}
	return repository.definitions.ListSummaries(readContext, request)
}

func (repository workspaceScopedControlPlaneReadRepository) ListRunRecordSummaries(
	readContext ReadRepositoryContext,
	request ListRunRecordSummariesRequest,
) ListRunRecordSummariesResult {
	result := ListRunRecordSummariesResult{
		TenantRef: readContext.TenantRef, Items: []RunRecordSummary{}, AuditRef: readContext.AuditRef,
	}
	if failure := validateWorkspaceReadRepositoryContext(readContext); failure != "" {
		result.FailureCode = failure
		return result
	}
	if repository.runs == nil {
		result.FailureCode = ReadRepositoryFailureStoreUnavailable
		return result
	}
	applicationID := strings.TrimSpace(request.Filters["application_ref"])
	if applicationID == "" {
		result.FailureCode = ReadRepositoryFailureCode("workspace_application_selection_required")
		return result
	}
	workflowDefinitionID := strings.TrimSpace(request.Filters["workflow_definition_id"])
	executionSourceKind := ""
	if workflowDefinitionID != "" {
		executionSourceKind = workflowDefinitionExecutionSourceKind
	}
	if request.Sort != "" && request.Sort != "started_at_desc" {
		result.FailureCode = ReadRepositoryFailureInvalidFilter
		return result
	}
	serviceResult := (workflowExecutorService{store: repository.runs}).ListRuns(
		WorkflowRunContext{
			RequestContext: readContext.RequestContext, RequestID: readContext.RequestID,
			TenantRef: readContext.TenantRef, WorkspaceID: readContext.WorkspaceID,
			ApplicationID: applicationID, ActorRef: readContext.SubjectRef,
			ScopeGrants: append([]string{}, readContext.ScopeGrants...), AuditRef: readContext.AuditRef,
		},
		WorkflowRunListRequest{
			Limit: request.Limit, Cursor: request.Cursor,
			Status:              WorkflowRunStatus(strings.TrimSpace(request.Filters["status"])),
			ExecutionSourceKind: executionSourceKind,
			ExecutionSourceID:   workflowDefinitionID,
			FailureCode:         WorkflowRunFailureCode(strings.TrimSpace(request.Filters["failure_code"])),
		},
	)
	if serviceResult.FailureCode != "" {
		result.FailureCode = translateWorkflowRunReadFailure(serviceResult.FailureCode)
		return result
	}
	result.Items = make([]RunRecordSummary, 0, len(serviceResult.Runs))
	for _, summary := range serviceResult.Runs {
		var failureCode *ReadRepositoryFailureCode
		if summary.FailureCode != "" {
			value := ReadRepositoryFailureCode(summary.FailureCode)
			failureCode = &value
		}
		result.Items = append(result.Items, RunRecordSummary{
			RunID: summary.RunID, TenantRef: readContext.TenantRef,
			WorkflowDefinitionID: summary.ExecutionSourceID, ApplicationRef: summary.ApplicationID,
			Status: string(summary.Status), FailureCode: failureCode,
			TraceID: summary.RequestID, StartedAt: summary.StartedAt, CompletedAt: summary.CompletedAt,
		})
	}
	if strings.TrimSpace(serviceResult.NextCursor) != "" {
		result.NextCursor = &serviceResult.NextCursor
	}
	return result
}

func validateWorkspaceReadRepositoryContext(readContext ReadRepositoryContext) ReadRepositoryFailureCode {
	if strings.TrimSpace(readContext.TenantRef) == "" || strings.TrimSpace(readContext.SubjectRef) == "" ||
		strings.TrimSpace(readContext.WorkspaceID) == "" {
		return ReadRepositoryFailureContractMismatch
	}
	return ""
}

func applicationCatalogContextFromReadRepositoryContext(readContext ReadRepositoryContext) ApplicationCatalogContext {
	requestContext := readContext.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	return ApplicationCatalogContext{
		RequestContext: requestContext, RequestID: readContext.RequestID,
		TenantRef: readContext.TenantRef, WorkspaceID: readContext.WorkspaceID,
		ActorRef: readContext.SubjectRef, OwnerSubjectRef: readContext.SubjectRef,
		AuditRef: readContext.AuditRef,
	}
}

func apiKeyContextFromReadRepositoryContext(readContext ReadRepositoryContext) APIKeyContext {
	appContext := applicationCatalogContextFromReadRepositoryContext(readContext)
	return APIKeyContext{
		RequestContext: appContext.RequestContext, RequestID: appContext.RequestID,
		TenantRef: appContext.TenantRef, WorkspaceID: appContext.WorkspaceID,
		ActorRef: appContext.ActorRef, OwnerSubjectRef: appContext.OwnerSubjectRef,
		AuditRef: appContext.AuditRef,
	}
}

func translateApplicationCatalogReadFailure(code string) ReadRepositoryFailureCode {
	switch code {
	case ApplicationCatalogFailurePayloadInvalid, ApplicationCatalogFailureCursorInvalid:
		return ReadRepositoryFailureInvalidFilter
	case ApplicationCatalogFailureStoreUnavailable:
		return ReadRepositoryFailureStoreUnavailable
	default:
		return ReadRepositoryFailureContractMismatch
	}
}

func translateAPIKeyReadFailure(code string) ReadRepositoryFailureCode {
	switch code {
	case APIKeyFailurePayloadInvalid, APIKeyFailureCursorInvalid:
		return ReadRepositoryFailureInvalidFilter
	case APIKeyFailureStoreUnavailable:
		return ReadRepositoryFailureStoreUnavailable
	default:
		return ReadRepositoryFailureContractMismatch
	}
}

func translateWorkflowRunReadFailure(code WorkflowRunFailureCode) ReadRepositoryFailureCode {
	switch code {
	case WorkflowRunFailureFilterInvalid, WorkflowRunFailureCursorInvalid:
		return ReadRepositoryFailureInvalidFilter
	case WorkflowRunFailureStoreUnavailable:
		return ReadRepositoryFailureStoreUnavailable
	default:
		return ReadRepositoryFailureContractMismatch
	}
}
