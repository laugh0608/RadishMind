package httpapi

type controlPlaneReadRepository interface {
	ReadTenantSummary(ReadRepositoryContext, ReadTenantSummaryRequest) ReadTenantSummaryResult
	ListApplicationSummaries(ReadRepositoryContext, ListApplicationSummariesRequest) ListApplicationSummariesResult
	ListAPIKeySummaries(ReadRepositoryContext, ListAPIKeySummariesRequest) ListAPIKeySummariesResult
	ReadQuotaSummary(ReadRepositoryContext, ReadQuotaSummaryRequest) ReadQuotaSummaryResult
	ListWorkflowDefinitionSummaries(ReadRepositoryContext, ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult
	ListRunRecordSummaries(ReadRepositoryContext, ListRunRecordSummariesRequest) ListRunRecordSummariesResult
	ListAuditSummaries(ReadRepositoryContext, ListAuditSummariesRequest) ListAuditSummariesResult
	SideEffects() controlPlaneReadSideEffects
}

type ControlPlaneReadRepository = controlPlaneReadRepository

type fakeStoreControlPlaneReadRepository struct {
	store controlPlaneReadStore
}

func newControlPlaneReadRepository(store controlPlaneReadStore) ControlPlaneReadRepository {
	return fakeStoreControlPlaneReadRepository{store: store}
}

func (s *Server) controlPlaneReadRepository() ControlPlaneReadRepository {
	if s.controlPlaneReadRepo != nil {
		return s.controlPlaneReadRepo
	}
	return newControlPlaneReadRepository(s.controlPlaneReadDataStore())
}

func (repo fakeStoreControlPlaneReadRepository) ReadTenantSummary(
	context ReadRepositoryContext,
	request ReadTenantSummaryRequest,
) ReadTenantSummaryResult {
	item, found := repo.store.TenantSummary(context.TenantRef)
	if !found {
		return ReadTenantSummaryResult{TenantRef: context.TenantRef, Items: []TenantSummary{}, AuditRef: context.AuditRef}
	}
	return ReadTenantSummaryResult{
		TenantRef: context.TenantRef,
		Items:     []TenantSummary{tenantSummaryFromControlPlaneReadMap(item)},
		AuditRef:  context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ListApplicationSummaries(
	context ReadRepositoryContext,
	request ListApplicationSummariesRequest,
) ListApplicationSummariesResult {
	items, nextCursor := repo.store.ApplicationSummaries(context.TenantRef, mapFromReadRepositoryFilters(request.Filters))
	return ListApplicationSummariesResult{
		TenantRef:   context.TenantRef,
		Items:       applicationSummariesFromControlPlaneReadMaps(items),
		NextCursor:  nextCursor,
		FailureCode: "",
		AuditRef:    context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ListAPIKeySummaries(
	context ReadRepositoryContext,
	request ListAPIKeySummariesRequest,
) ListAPIKeySummariesResult {
	items, nextCursor := repo.store.APIKeySummaries(context.TenantRef, mapFromReadRepositoryFilters(request.Filters))
	return ListAPIKeySummariesResult{
		TenantRef:   context.TenantRef,
		Items:       apiKeySummariesFromControlPlaneReadMaps(items),
		NextCursor:  nextCursor,
		FailureCode: "",
		AuditRef:    context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ReadQuotaSummary(
	context ReadRepositoryContext,
	request ReadQuotaSummaryRequest,
) ReadQuotaSummaryResult {
	item, found := repo.store.QuotaSummary(context.TenantRef)
	if !found {
		return ReadQuotaSummaryResult{TenantRef: context.TenantRef, Items: []QuotaSummary{}, AuditRef: context.AuditRef}
	}
	return ReadQuotaSummaryResult{
		TenantRef: context.TenantRef,
		Items:     []QuotaSummary{quotaSummaryFromControlPlaneReadMap(item)},
		AuditRef:  context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ListWorkflowDefinitionSummaries(
	context ReadRepositoryContext,
	request ListWorkflowDefinitionSummariesRequest,
) ListWorkflowDefinitionSummariesResult {
	items, nextCursor := repo.store.WorkflowDefinitionSummaries(context.TenantRef, mapFromReadRepositoryFilters(request.Filters))
	return ListWorkflowDefinitionSummariesResult{
		TenantRef:   context.TenantRef,
		Items:       workflowDefinitionSummariesFromControlPlaneReadMaps(items),
		NextCursor:  nextCursor,
		FailureCode: "",
		AuditRef:    context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ListRunRecordSummaries(
	context ReadRepositoryContext,
	request ListRunRecordSummariesRequest,
) ListRunRecordSummariesResult {
	items, nextCursor := repo.store.RunRecordSummaries(context.TenantRef, mapFromReadRepositoryFilters(request.Filters))
	return ListRunRecordSummariesResult{
		TenantRef:   context.TenantRef,
		Items:       runRecordSummariesFromControlPlaneReadMaps(items),
		NextCursor:  nextCursor,
		FailureCode: "",
		AuditRef:    context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) ListAuditSummaries(
	context ReadRepositoryContext,
	request ListAuditSummariesRequest,
) ListAuditSummariesResult {
	items, nextCursor := repo.store.AuditSummaries(context.TenantRef, mapFromReadRepositoryFilters(request.Filters))
	return ListAuditSummariesResult{
		TenantRef:   context.TenantRef,
		Items:       auditSummariesFromControlPlaneReadMaps(items),
		NextCursor:  nextCursor,
		FailureCode: "",
		AuditRef:    context.AuditRef,
	}
}

func (repo fakeStoreControlPlaneReadRepository) SideEffects() controlPlaneReadSideEffects {
	return repo.store.SideEffects()
}

func mapFromReadRepositoryFilters(filters ReadRepositoryFilters) map[string]string {
	output := make(map[string]string, len(filters))
	for key, value := range filters {
		output[key] = value
	}
	return output
}

func tenantSummaryFromControlPlaneReadMap(item map[string]any) TenantSummary {
	return TenantSummary{
		TenantRef:           stringFromControlPlaneReadMap(item, "tenant_ref"),
		TenantDisplayName:   stringFromControlPlaneReadMap(item, "tenant_display_name"),
		TenantState:         stringFromControlPlaneReadMap(item, "tenant_state"),
		PlanRef:             stringFromControlPlaneReadMap(item, "plan_ref"),
		QuotaSummaryRef:     stringFromControlPlaneReadMap(item, "quota_summary_ref"),
		DeploymentStatusRef: stringFromControlPlaneReadMap(item, "deployment_status_ref"),
		AuditSummaryRef:     stringFromControlPlaneReadMap(item, "audit_summary_ref"),
	}
}

func applicationSummariesFromControlPlaneReadMaps(items []map[string]any) []ApplicationSummary {
	output := make([]ApplicationSummary, 0, len(items))
	for _, item := range items {
		output = append(output, ApplicationSummary{
			ApplicationRef:              stringFromControlPlaneReadMap(item, "application_ref"),
			TenantRef:                   stringFromControlPlaneReadMap(item, "tenant_ref"),
			ApplicationKind:             stringFromControlPlaneReadMap(item, "application_kind"),
			DisplayName:                 stringFromControlPlaneReadMap(item, "display_name"),
			OwnerSubjectRef:             stringFromControlPlaneReadMap(item, "owner_subject_ref"),
			LatestWorkflowDefinitionRef: stringFromControlPlaneReadMap(item, "latest_workflow_definition_ref"),
			LastRunStatus:               stringFromControlPlaneReadMap(item, "last_run_status"),
			UpdatedAt:                   stringFromControlPlaneReadMap(item, "updated_at"),
		})
	}
	return output
}

func apiKeySummariesFromControlPlaneReadMaps(items []map[string]any) []APIKeySummary {
	output := make([]APIKeySummary, 0, len(items))
	for _, item := range items {
		output = append(output, APIKeySummary{
			APIKeyID:        stringFromControlPlaneReadMap(item, "api_key_id"),
			TenantRef:       stringFromControlPlaneReadMap(item, "tenant_ref"),
			OwnerSubjectRef: stringFromControlPlaneReadMap(item, "owner_subject_ref"),
			Scopes:          stringSliceFromControlPlaneReadMap(item, "scopes"),
			State:           stringFromControlPlaneReadMap(item, "state"),
			CreatedAt:       stringFromControlPlaneReadMap(item, "created_at"),
			ExpiresAt:       optionalStringFromControlPlaneReadMap(item, "expires_at"),
			LastUsedAt:      optionalStringFromControlPlaneReadMap(item, "last_used_at"),
		})
	}
	return output
}

func quotaSummaryFromControlPlaneReadMap(item map[string]any) QuotaSummary {
	return QuotaSummary{
		QuotaID:              stringFromControlPlaneReadMap(item, "quota_id"),
		TenantRef:            stringFromControlPlaneReadMap(item, "tenant_ref"),
		Period:               stringFromControlPlaneReadMap(item, "period"),
		RequestLimit:         intFromControlPlaneReadMap(item, "request_limit"),
		TokenLimit:           intFromControlPlaneReadMap(item, "token_limit"),
		CostLimit:            floatFromControlPlaneReadMap(item, "cost_limit"),
		UsageSnapshot:        mapValueFromControlPlaneReadMap(item, "usage_snapshot"),
		OverQuotaFailureCode: stringFromControlPlaneReadMap(item, "over_quota_failure_code"),
	}
}

func workflowDefinitionSummariesFromControlPlaneReadMaps(items []map[string]any) []WorkflowDefinitionSummary {
	output := make([]WorkflowDefinitionSummary, 0, len(items))
	for _, item := range items {
		output = append(output, WorkflowDefinitionSummary{
			WorkflowDefinitionID:        stringFromControlPlaneReadMap(item, "workflow_definition_id"),
			TenantRef:                   stringFromControlPlaneReadMap(item, "tenant_ref"),
			ApplicationRef:              stringFromControlPlaneReadMap(item, "application_ref"),
			Version:                     intFromControlPlaneReadMap(item, "version"),
			DefinitionStatus:            stringFromControlPlaneReadMap(item, "definition_status"),
			NodeCount:                   intFromControlPlaneReadMap(item, "node_count"),
			RiskLevel:                   stringFromControlPlaneReadMap(item, "risk_level"),
			RequiresConfirmationCapable: boolFromControlPlaneReadMap(item, "requires_confirmation_capable"),
			UpdatedAt:                   stringFromControlPlaneReadMap(item, "updated_at"),
		})
	}
	return output
}

func runRecordSummariesFromControlPlaneReadMaps(items []map[string]any) []RunRecordSummary {
	output := make([]RunRecordSummary, 0, len(items))
	for _, item := range items {
		output = append(output, RunRecordSummary{
			RunID:                stringFromControlPlaneReadMap(item, "run_id"),
			TenantRef:            stringFromControlPlaneReadMap(item, "tenant_ref"),
			WorkflowDefinitionID: stringFromControlPlaneReadMap(item, "workflow_definition_id"),
			ApplicationRef:       stringFromControlPlaneReadMap(item, "application_ref"),
			Status:               stringFromControlPlaneReadMap(item, "status"),
			FailureCode:          failureCodeFromControlPlaneReadMap(item, "failure_code"),
			CostSummary:          mapValueFromControlPlaneReadMap(item, "cost_summary"),
			TraceID:              stringFromControlPlaneReadMap(item, "trace_id"),
			StartedAt:            stringFromControlPlaneReadMap(item, "started_at"),
			CompletedAt:          stringFromControlPlaneReadMap(item, "completed_at"),
		})
	}
	return output
}

func auditSummariesFromControlPlaneReadMaps(items []map[string]any) []AuditSummary {
	output := make([]AuditSummary, 0, len(items))
	for _, item := range items {
		output = append(output, AuditSummary{
			AuditRef:        stringFromControlPlaneReadMap(item, "audit_ref"),
			TenantRef:       stringFromControlPlaneReadMap(item, "tenant_ref"),
			ActorSubjectRef: stringFromControlPlaneReadMap(item, "actor_subject_ref"),
			EventKind:       stringFromControlPlaneReadMap(item, "event_kind"),
			ResourceRef:     stringFromControlPlaneReadMap(item, "resource_ref"),
			Decision:        stringFromControlPlaneReadMap(item, "decision"),
			FailureCode:     failureCodeFromControlPlaneReadMap(item, "failure_code"),
			TraceID:         stringFromControlPlaneReadMap(item, "trace_id"),
			RecordedAt:      stringFromControlPlaneReadMap(item, "recorded_at"),
		})
	}
	return output
}

func stringFromControlPlaneReadMap(item map[string]any, key string) string {
	value, _ := item[key].(string)
	return value
}

func optionalStringFromControlPlaneReadMap(item map[string]any, key string) *string {
	value, ok := item[key].(string)
	if !ok {
		return nil
	}
	return &value
}

func stringSliceFromControlPlaneReadMap(item map[string]any, key string) []string {
	switch values := item[key].(type) {
	case []string:
		output := make([]string, len(values))
		copy(output, values)
		return output
	case []any:
		output := make([]string, 0, len(values))
		for _, value := range values {
			if item, ok := value.(string); ok {
				output = append(output, item)
			}
		}
		return output
	default:
		return nil
	}
}

func intFromControlPlaneReadMap(item map[string]any, key string) int {
	switch value := item[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func floatFromControlPlaneReadMap(item map[string]any, key string) float64 {
	switch value := item[key].(type) {
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case float64:
		return value
	default:
		return 0
	}
}

func boolFromControlPlaneReadMap(item map[string]any, key string) bool {
	value, _ := item[key].(bool)
	return value
}

func mapValueFromControlPlaneReadMap(item map[string]any, key string) map[string]any {
	value, ok := item[key].(map[string]any)
	if !ok {
		return nil
	}
	return cloneControlPlaneReadMap(value)
}

func failureCodeFromControlPlaneReadMap(item map[string]any, key string) *ReadRepositoryFailureCode {
	value, ok := item[key].(string)
	if !ok {
		return nil
	}
	failureCode := ReadRepositoryFailureCode(value)
	return &failureCode
}

func tenantSummaryToControlPlaneReadMap(summary TenantSummary) map[string]any {
	return map[string]any{
		"tenant_ref":            summary.TenantRef,
		"tenant_display_name":   summary.TenantDisplayName,
		"tenant_state":          summary.TenantState,
		"plan_ref":              summary.PlanRef,
		"quota_summary_ref":     summary.QuotaSummaryRef,
		"deployment_status_ref": summary.DeploymentStatusRef,
		"audit_summary_ref":     summary.AuditSummaryRef,
	}
}

func applicationSummariesToControlPlaneReadMaps(summaries []ApplicationSummary) []map[string]any {
	output := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		output = append(output, map[string]any{
			"application_ref":                summary.ApplicationRef,
			"tenant_ref":                     summary.TenantRef,
			"application_kind":               summary.ApplicationKind,
			"display_name":                   summary.DisplayName,
			"owner_subject_ref":              summary.OwnerSubjectRef,
			"latest_workflow_definition_ref": summary.LatestWorkflowDefinitionRef,
			"last_run_status":                summary.LastRunStatus,
			"updated_at":                     summary.UpdatedAt,
		})
	}
	return output
}

func apiKeySummariesToControlPlaneReadMaps(summaries []APIKeySummary) []map[string]any {
	output := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		output = append(output, map[string]any{
			"api_key_id":        summary.APIKeyID,
			"tenant_ref":        summary.TenantRef,
			"owner_subject_ref": summary.OwnerSubjectRef,
			"scopes":            append([]string{}, summary.Scopes...),
			"state":             summary.State,
			"created_at":        summary.CreatedAt,
			"expires_at":        optionalStringToControlPlaneReadValue(summary.ExpiresAt),
			"last_used_at":      optionalStringToControlPlaneReadValue(summary.LastUsedAt),
		})
	}
	return output
}

func quotaSummaryToControlPlaneReadMap(summary QuotaSummary) map[string]any {
	return map[string]any{
		"quota_id":                summary.QuotaID,
		"tenant_ref":              summary.TenantRef,
		"period":                  summary.Period,
		"request_limit":           summary.RequestLimit,
		"token_limit":             summary.TokenLimit,
		"cost_limit":              summary.CostLimit,
		"usage_snapshot":          cloneControlPlaneReadMap(summary.UsageSnapshot),
		"over_quota_failure_code": summary.OverQuotaFailureCode,
	}
}

func workflowDefinitionSummariesToControlPlaneReadMaps(summaries []WorkflowDefinitionSummary) []map[string]any {
	output := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		output = append(output, map[string]any{
			"workflow_definition_id":        summary.WorkflowDefinitionID,
			"tenant_ref":                    summary.TenantRef,
			"application_ref":               summary.ApplicationRef,
			"version":                       summary.Version,
			"definition_status":             summary.DefinitionStatus,
			"node_count":                    summary.NodeCount,
			"risk_level":                    summary.RiskLevel,
			"requires_confirmation_capable": summary.RequiresConfirmationCapable,
			"updated_at":                    summary.UpdatedAt,
		})
	}
	return output
}

func runRecordSummariesToControlPlaneReadMaps(summaries []RunRecordSummary) []map[string]any {
	output := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		output = append(output, map[string]any{
			"run_id":                 summary.RunID,
			"tenant_ref":             summary.TenantRef,
			"workflow_definition_id": summary.WorkflowDefinitionID,
			"application_ref":        summary.ApplicationRef,
			"status":                 summary.Status,
			"failure_code":           optionalFailureCodeToControlPlaneReadValue(summary.FailureCode),
			"cost_summary":           cloneControlPlaneReadMap(summary.CostSummary),
			"trace_id":               summary.TraceID,
			"started_at":             summary.StartedAt,
			"completed_at":           summary.CompletedAt,
		})
	}
	return output
}

func auditSummariesToControlPlaneReadMaps(summaries []AuditSummary) []map[string]any {
	output := make([]map[string]any, 0, len(summaries))
	for _, summary := range summaries {
		output = append(output, map[string]any{
			"audit_ref":         summary.AuditRef,
			"tenant_ref":        summary.TenantRef,
			"actor_subject_ref": summary.ActorSubjectRef,
			"event_kind":        summary.EventKind,
			"resource_ref":      summary.ResourceRef,
			"decision":          summary.Decision,
			"failure_code":      optionalFailureCodeToControlPlaneReadValue(summary.FailureCode),
			"trace_id":          summary.TraceID,
			"recorded_at":       summary.RecordedAt,
		})
	}
	return output
}

func optionalStringToControlPlaneReadValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func optionalFailureCodeToControlPlaneReadValue(value *ReadRepositoryFailureCode) any {
	if value == nil {
		return nil
	}
	return string(*value)
}
