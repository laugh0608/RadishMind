package httpapi

func (server *Server) applicationEvaluationCampaignService() applicationEvaluationCampaignService {
	return newApplicationEvaluationCampaignService(
		server.applicationEvaluationRepository,
		server.resolveApplicationEvaluationCampaignAuthority,
		server.invokeApplicationEvaluationCampaignItem,
		server.readApplicationEvaluationCampaignRun,
	)
}

func (server *Server) applicationEvaluationHandoffService() applicationEvaluationHandoffService {
	evaluation := server.workflowEvaluationService()
	suite := server.workflowEvaluationSuiteService()
	return newApplicationEvaluationHandoffService(
		server.applicationEvaluationRepository,
		server.effectiveApplicationRunStore(),
		evaluation,
		suite,
	)
}

func (server *Server) resolveApplicationEvaluationCampaignAuthority(ctx ApplicationEvaluationContext, version ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string) {
	switch version.ExecutionProfile {
	case applicationInteractionProfilePrompt:
		authority, failure := server.promptApplicationInvocationService().resolveAuthority(promptApplicationEvaluationRuntimeContext(ctx))
		if failure != "" {
			return ApplicationEvaluationCampaignAuthority{}, applicationEvaluationAuthorityFailure(failure)
		}
		return applicationEvaluationCampaignAuthorityPayload(version.ExecutionProfile, authority.Snapshot, authority.Snapshot.AuthorityDigest)
	case applicationInteractionProfileAgentCopilot:
		authority, failure := server.agentCopilotInvocationService().resolveAuthority(agentCopilotApplicationEvaluationRuntimeContext(ctx))
		if failure != "" {
			return ApplicationEvaluationCampaignAuthority{}, applicationEvaluationAuthorityFailure(failure)
		}
		return applicationEvaluationCampaignAuthorityPayload(version.ExecutionProfile, authority.Snapshot, authority.Snapshot.AuthorityDigest)
	case applicationInteractionProfileWorkflow, applicationInteractionProfileRAG:
		binding := ApplicationInteractionProfileBinding{ExecutionProfile: version.ExecutionProfile}
		if version.ExecutionProfile == applicationInteractionProfileWorkflow && version.Target.WorkflowDefinition != nil {
			binding.DefinitionID = version.Target.WorkflowDefinition.DefinitionID
		}
		snapshot, failure := server.applicationInteractionAuthorityResolver().Resolve(applicationEvaluationInteractionContext(ctx), binding)
		if failure != "" {
			return ApplicationEvaluationCampaignAuthority{}, applicationEvaluationAuthorityFailure(failure)
		}
		if version.ExecutionProfile == applicationInteractionProfileWorkflow && !applicationEvaluationWorkflowTargetMatchesAuthority(version.Target.WorkflowDefinition, snapshot.WorkflowDefinition) {
			return ApplicationEvaluationCampaignAuthority{}, ApplicationEvaluationFailureAuthorityChanged
		}
		return applicationEvaluationCampaignAuthorityPayload(version.ExecutionProfile, snapshot, snapshot.AuthorityDigest)
	default:
		return ApplicationEvaluationCampaignAuthority{}, ApplicationEvaluationFailureProfileIneligible
	}
}

func (server *Server) invokeApplicationEvaluationCampaignItem(
	ctx ApplicationEvaluationContext,
	version ApplicationEvaluationPlanVersion,
	item ApplicationEvaluationPlanItem,
	runID string,
) (*WorkflowRunRecord, string, string) {
	switch version.ExecutionProfile {
	case applicationInteractionProfileWorkflow:
		if version.Target.WorkflowDefinition == nil || item.WorkflowDefinition == nil {
			return nil, ApplicationEvaluationFailureStoreContract, applicationEvaluationFailureSummary(ApplicationEvaluationFailureStoreContract)
		}
		target, fixture := version.Target.WorkflowDefinition, item.WorkflowDefinition
		service := server.workflowDefinitionExecutionService()
		service.executor.newRunID = func() (string, error) { return runID, nil }
		result := service.StartRun(applicationEvaluationWorkflowRunContext(ctx), WorkflowDefinitionRunRequest{
			DefinitionID: target.DefinitionID, ExpectedPointerVersion: target.ExpectedPointerVersion,
			ExpectedDefinitionVersion: target.ExpectedDefinitionVersion, ExpectedDefinitionDigest: target.ExpectedDefinitionDigest,
			InputText: fixture.InputText, ConditionValues: fixture.ConditionValues, Model: fixture.Model, Temperature: fixture.Temperature,
		})
		return result.Record, string(result.FailureCode), result.FailureSummary
	case applicationInteractionProfileRAG:
		if item.ApplicationRAG == nil {
			return nil, ApplicationEvaluationFailureStoreContract, applicationEvaluationFailureSummary(ApplicationEvaluationFailureStoreContract)
		}
		service := server.workflowRAGApplicationInvocationService()
		service.newRunID = func() (string, error) { return runID, nil }
		result := service.Invoke(workflowRAGApplicationEvaluationRuntimeContext(ctx), WorkflowRAGApplicationInvocationInput{Input: item.ApplicationRAG.Input})
		return result.Run, result.FailureCode, result.FailureSummary
	case applicationInteractionProfilePrompt:
		if item.PromptApplication == nil {
			return nil, ApplicationEvaluationFailureStoreContract, applicationEvaluationFailureSummary(ApplicationEvaluationFailureStoreContract)
		}
		result := server.promptApplicationInvocationService().Invoke(promptApplicationEvaluationRuntimeContext(ctx), PromptApplicationInvocationInput{
			Variables: item.PromptApplication.Variables, ClientInvocationKey: runID, RunID: runID,
		})
		return result.Run, result.FailureCode, result.FailureSummary
	case applicationInteractionProfileAgentCopilot:
		if item.AgentCopilot == nil {
			return nil, ApplicationEvaluationFailureStoreContract, applicationEvaluationFailureSummary(ApplicationEvaluationFailureStoreContract)
		}
		fixture := item.AgentCopilot
		result := server.agentCopilotInvocationService().Invoke(agentCopilotApplicationEvaluationRuntimeContext(ctx), AgentCopilotInvocationInput{
			Task: fixture.Task, Locale: fixture.Locale, ConversationID: fixture.ConversationID,
			Artifacts: fixture.Artifacts, Context: fixture.Context, ClientInvocationKey: runID, RunID: runID,
		})
		return result.Run, result.FailureCode, result.FailureSummary
	default:
		return nil, ApplicationEvaluationFailureProfileIneligible, applicationEvaluationFailureSummary(ApplicationEvaluationFailureProfileIneligible)
	}
}

func (server *Server) readApplicationEvaluationCampaignRun(ctx ApplicationEvaluationContext, runID string) (WorkflowRunRecord, bool, error) {
	return server.effectiveApplicationRunStore().ReadRun(applicationEvaluationWorkflowRunContext(ctx), runID)
}

func applicationEvaluationInteractionContext(ctx ApplicationEvaluationContext) ApplicationInteractionContext {
	return ApplicationInteractionContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled,
	}
}

func applicationEvaluationWorkflowRunContext(ctx ApplicationEvaluationContext) WorkflowRunContext {
	return WorkflowRunContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, AuditRef: ctx.AuditRef,
	}
}

func workflowRAGApplicationEvaluationRuntimeContext(ctx ApplicationEvaluationContext) WorkflowRAGApplicationRuntimeContext {
	return WorkflowRAGApplicationRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled,
	}
}

func promptApplicationEvaluationRuntimeContext(ctx ApplicationEvaluationContext) PromptApplicationRuntimeContext {
	return PromptApplicationRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled,
	}
}

func agentCopilotApplicationEvaluationRuntimeContext(ctx ApplicationEvaluationContext) AgentCopilotRuntimeContext {
	return AgentCopilotRuntimeContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef, WriteEnabled: ctx.WriteEnabled,
	}
}

func applicationEvaluationWorkflowTargetMatchesAuthority(target *ApplicationEvaluationDefinitionTarget, authority *ApplicationInteractionWorkflowAuthority) bool {
	return target != nil && authority != nil && target.DefinitionID == authority.DefinitionID &&
		target.ExpectedPointerVersion == authority.ActivationPointerVersion && target.ExpectedDefinitionVersion == authority.DefinitionVersion &&
		target.ExpectedDefinitionDigest == authority.DefinitionDigest
}

func applicationEvaluationAuthorityFailure(code string) string {
	switch code {
	case ApplicationInteractionFailureApplicationMissing, ApplicationInteractionFailureAuthorityMissing,
		PromptApplicationRuntimeFailureNotFound, AgentCopilotRuntimeFailureNotFound:
		return ApplicationEvaluationFailureNotFound
	case ApplicationInteractionFailureApplicationArchived:
		return ApplicationEvaluationFailureArchived
	case ApplicationInteractionFailureProfileIneligible:
		return ApplicationEvaluationFailureProfileIneligible
	case ApplicationInteractionFailureStoreUnavailable, PromptApplicationRuntimeFailureStoreUnavailable,
		AgentCopilotRuntimeFailureStoreUnavailable:
		return ApplicationEvaluationFailureStoreUnavailable
	default:
		return ApplicationEvaluationFailureAuthorityChanged
	}
}
