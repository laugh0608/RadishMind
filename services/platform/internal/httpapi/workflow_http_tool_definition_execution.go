package httpapi

import (
	"errors"
	"strings"
)

const workflowDefinitionHTTPToolExecutionKind = "workflow_definition_http_tool_execution"

type workflowHTTPToolResolvedExecutionSource struct {
	draft       SavedWorkflowDraft
	version     WorkflowDefinitionVersion
	activation  WorkflowDefinitionActivation
	application ApplicationCatalogRecord
}

func (source workflowHTTPToolResolvedExecutionSource) definitionBound() bool {
	return source.version.DefinitionID != ""
}

func (service workflowHTTPToolExecutionService) resolveExecutionSource(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
) (workflowHTTPToolResolvedExecutionSource, WorkflowHTTPToolExecutionResult) {
	if plan.SchemaVersion == workflowHTTPToolPlanSchemaV2 && plan.SourceKind == workflowHTTPToolSourceDefinition {
		return service.resolveDefinitionExecutionSource(ctx, plan)
	}
	draft, actionResult := service.actions.readExactEligibleDraft(
		ctx,
		plan.PlannedByActorRef,
		plan.DraftID,
		plan.DraftVersion,
		plan.NodeID,
	)
	if actionResult.FailureCode != "" {
		return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailureForAction(actionResult)
	}
	return workflowHTTPToolResolvedExecutionSource{draft: draft}, WorkflowHTTPToolExecutionResult{}
}

func (service workflowHTTPToolExecutionService) resolveDefinitionExecutionSource(
	ctx WorkflowHTTPToolActionContext,
	plan WorkflowHTTPToolActionPlan,
) (workflowHTTPToolResolvedExecutionSource, WorkflowHTTPToolExecutionResult) {
	version, activation, actionResult := service.actions.readExactEligibleDefinition(
		ctx,
		plan.PlannedByActorRef,
		plan.WorkflowDefinitionID,
		plan.NodeID,
	)
	if actionResult.FailureCode != "" {
		return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailureForAction(actionResult)
	}
	if version.Version != plan.WorkflowDefinitionVersion || version.DefinitionDigest != plan.WorkflowDefinitionDigest ||
		activation.PointerVersion != plan.ActivationPointerVersion {
		return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailure(
			WorkflowRunFailureDefinitionAuthority,
			"Active workflow definition authority changed before the tool execution claim.",
		)
	}
	if service.applications == nil {
		return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailure(
			WorkflowRunFailureToolStore,
			"Workflow definition application authority is unavailable.",
		)
	}
	application, err := service.applications.RequireActive(ApplicationCatalogContext{
		RequestContext:  ctx.RequestContext,
		RequestID:       ctx.RequestID,
		TenantRef:       ctx.TenantRef,
		WorkspaceID:     ctx.WorkspaceID,
		ActorRef:        ctx.ActorRef,
		OwnerSubjectRef: strings.TrimSpace(plan.PlannedByActorRef),
		AuditRef:        ctx.AuditRef,
	}, ctx.ApplicationID)
	if err != nil {
		if errors.Is(err, errApplicationCatalogStoreUnavailable) {
			return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailure(
				WorkflowRunFailureToolStore,
				"Workflow definition application authority is unavailable.",
			)
		}
		return workflowHTTPToolResolvedExecutionSource{}, workflowHTTPToolExecutionFailure(
			WorkflowRunFailureDefinitionAuthority,
			"Application lifecycle changed before the tool execution claim.",
		)
	}
	draft := workflowDefinitionSnapshotAsDraft(
		WorkflowRunContext{WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID},
		version,
	)
	return workflowHTTPToolResolvedExecutionSource{
		draft: draft, version: version, activation: activation, application: application,
	}, WorkflowHTTPToolExecutionResult{}
}

func workflowHTTPToolDefinitionSourceMatches(
	left workflowHTTPToolResolvedExecutionSource,
	right workflowHTTPToolResolvedExecutionSource,
) bool {
	return left.definitionBound() && right.definitionBound() &&
		left.version.DefinitionID == right.version.DefinitionID &&
		left.version.Version == right.version.Version &&
		left.version.DefinitionDigest == right.version.DefinitionDigest &&
		left.activation.PointerVersion == right.activation.PointerVersion &&
		left.activation.ActiveVersion == right.activation.ActiveVersion &&
		left.activation.ActiveDefinitionDigest == right.activation.ActiveDefinitionDigest &&
		left.application.ApplicationID == right.application.ApplicationID &&
		left.application.RecordVersion == right.application.RecordVersion &&
		left.application.LifecycleState == applicationCatalogLifecycleActive &&
		right.application.LifecycleState == applicationCatalogLifecycleActive
}

func workflowDefinitionHTTPToolRunAuthority(source workflowHTTPToolResolvedExecutionSource) *WorkflowDefinitionRunAuthority {
	return &WorkflowDefinitionRunAuthority{
		DefinitionID:             source.version.DefinitionID,
		DefinitionVersion:        source.version.Version,
		DefinitionDigest:         source.version.DefinitionDigest,
		ActivationPointerVersion: source.activation.PointerVersion,
		CandidateID:              source.version.CandidateID,
		CandidateReviewVersion:   source.version.CandidateReviewVersion,
		SourceDraftID:            source.version.SourceDraftID,
		SourceDraftVersion:       source.version.SourceDraftVersion,
		SourceDraftDigest:        source.version.SourceDraftDigest,
		ApplicationRecordVersion: source.application.RecordVersion,
		ApplicationLifecycle:     source.application.LifecycleState,
	}
}

func validateWorkflowDefinitionHTTPToolRunStoreRecord(
	runContext WorkflowRunContext,
	record *WorkflowRunRecord,
) error {
	if record == nil || record.SchemaVersion != workflowRunRecordDefinitionToolSchemaVersion ||
		record.TenantRef != strings.TrimSpace(runContext.TenantRef) ||
		record.ExecutionKind != workflowDefinitionHTTPToolExecutionKind ||
		record.ExecutionSourceKind != workflowDefinitionExecutionSourceKind ||
		record.ExecutionProfile != workflowDefinitionHTTPToolProfile ||
		record.ExecutionSource == nil || record.ExecutionSource.Kind != record.ExecutionKind ||
		record.ExecutionSource.SourceKind != record.ExecutionSourceKind ||
		record.ExecutionSource.ID != record.ExecutionSourceID ||
		record.ExecutionSource.Version != record.ExecutionSourceVersion ||
		!applicationDraftIdentifierPattern.MatchString(record.ExecutionSourceID) || record.ExecutionSourceVersion < 1 ||
		record.DefinitionAuthority == nil ||
		record.DefinitionAuthority.DefinitionID != record.ExecutionSourceID ||
		record.DefinitionAuthority.DefinitionVersion != record.ExecutionSourceVersion ||
		!workflowHTTPToolDigestPattern.MatchString(record.DefinitionAuthority.DefinitionDigest) ||
		record.DefinitionAuthority.ActivationPointerVersion < 1 ||
		!applicationDraftIdentifierPattern.MatchString(record.DefinitionAuthority.CandidateID) ||
		record.DefinitionAuthority.CandidateReviewVersion < 1 ||
		!applicationDraftIdentifierPattern.MatchString(record.DefinitionAuthority.SourceDraftID) ||
		record.DefinitionAuthority.SourceDraftVersion < 1 ||
		!workflowHTTPToolDigestPattern.MatchString(record.DefinitionAuthority.SourceDraftDigest) ||
		record.DefinitionAuthority.SourceDraftDigest != record.DefinitionAuthority.DefinitionDigest ||
		record.DefinitionAuthority.ApplicationRecordVersion < 1 ||
		record.DefinitionAuthority.ApplicationLifecycle != applicationCatalogLifecycleActive ||
		record.DraftID != "" || record.DraftVersion != 0 || record.DraftDigest != "" ||
		!workflowHTTPToolDigestPattern.MatchString(record.InputDigest) ||
		record.InputBytes < 1 || record.InputBytes > workflowExecutorMaxInputBytes ||
		record.Output != "" || len(record.ConditionNodeIDs) != 0 ||
		record.RAGSnapshot != nil || record.RetrievalAttempt != nil || record.RAGAnswer != nil || record.RAGApplication != nil ||
		record.PromptApplication != nil || record.AgentCopilotAuthority != nil ||
		validateWorkflowRunToolEvidence(record) != nil ||
		!validWorkflowRunDiagnostic(record.Diagnostic, isTerminalWorkflowRunStatus(record.Status)) {
		return errWorkflowRunStoreContract
	}
	for _, node := range record.Nodes {
		if node.OutputPreview != "" {
			return errWorkflowRunStoreContract
		}
	}
	return nil
}
