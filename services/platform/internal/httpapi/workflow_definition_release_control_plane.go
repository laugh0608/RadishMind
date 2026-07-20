package httpapi

type liveWorkflowDefinitionControlPlaneReadRepository struct {
	ControlPlaneReadRepository
	definitions workflowDefinitionReleaseRepository
}

func (repository liveWorkflowDefinitionControlPlaneReadRepository) ListWorkflowDefinitionSummaries(ctx ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	if repository.definitions == nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	return repository.definitions.ListSummaries(ctx, request)
}
