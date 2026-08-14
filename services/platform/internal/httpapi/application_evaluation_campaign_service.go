package httpapi

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type ApplicationEvaluationCampaignExecuteInput struct {
	PlanID                         string
	PlanVersion                    int
	PlanDigest                     string
	ExpectedPlanRecordVersion      int
	ClientCampaignKey              string
	QuotaAPIKeyID                  string
	AcknowledgeSequentialExecution bool
	AcknowledgeQuotaConsumption    bool
}

type ApplicationEvaluationCampaignListInput struct {
	PlanID string
	Limit  int
	Cursor string
}

type ApplicationEvaluationCampaignResult struct {
	Campaign             *ApplicationEvaluationCampaign
	IdempotentReplay     bool
	FailureCode          string
	FailureSummary       string
	CurrentRecordVersion int
	CurrentState         string
}

type ApplicationEvaluationCampaignListResult struct {
	Campaigns      []ApplicationEvaluationCampaign
	NextCursor     string
	HasMore        bool
	FailureCode    string
	FailureSummary string
}

type applicationEvaluationCampaignAuthorityResolver func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion) (ApplicationEvaluationCampaignAuthority, string)
type applicationEvaluationCampaignInvoker func(ApplicationEvaluationContext, ApplicationEvaluationPlanVersion, ApplicationEvaluationPlanItem, string) (*WorkflowRunRecord, string, string)
type applicationEvaluationCampaignRunReader func(ApplicationEvaluationContext, string) (WorkflowRunRecord, bool, error)

type applicationEvaluationCampaignService struct {
	repository       applicationEvaluationRepository
	resolveAuthority applicationEvaluationCampaignAuthorityResolver
	invoke           applicationEvaluationCampaignInvoker
	readRun          applicationEvaluationCampaignRunReader
	now              func() time.Time
}

func newApplicationEvaluationCampaignService(
	repository applicationEvaluationRepository,
	resolveAuthority applicationEvaluationCampaignAuthorityResolver,
	invoke applicationEvaluationCampaignInvoker,
	readRun applicationEvaluationCampaignRunReader,
) applicationEvaluationCampaignService {
	return applicationEvaluationCampaignService{
		repository: repository, resolveAuthority: resolveAuthority, invoke: invoke, readRun: readRun,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (service applicationEvaluationCampaignService) Execute(ctx ApplicationEvaluationContext, input ApplicationEvaluationCampaignExecuteInput) ApplicationEvaluationCampaignResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationCampaignFailure(failure)
	}
	input.PlanID = strings.TrimSpace(input.PlanID)
	input.PlanDigest = strings.TrimSpace(input.PlanDigest)
	input.ClientCampaignKey = strings.TrimSpace(input.ClientCampaignKey)
	input.QuotaAPIKeyID = strings.TrimSpace(input.QuotaAPIKeyID)
	if !applicationEvaluationPlanIDPattern.MatchString(input.PlanID) || input.PlanVersion < 1 || input.ExpectedPlanRecordVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(input.PlanDigest) || !applicationDraftIdentifierPattern.MatchString(input.ClientCampaignKey) || !apiKeyIDPattern.MatchString(input.QuotaAPIKeyID) ||
		!input.AcknowledgeSequentialExecution || !input.AcknowledgeQuotaConsumption {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	plan, found, err := service.repository.ReadPlan(ctx, input.PlanID)
	if err != nil {
		return applicationEvaluationCampaignFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureNotFound)
	}
	if plan.RecordVersion != input.ExpectedPlanRecordVersion {
		return applicationEvaluationCampaignConflict(plan.RecordVersion, plan.LifecycleState)
	}
	if plan.LifecycleState != applicationEvaluationPlanStateActive {
		return applicationEvaluationCampaignConflictWithCode(plan.RecordVersion, plan.LifecycleState, ApplicationEvaluationFailureArchived)
	}
	version, found, err := service.repository.ReadPlanVersion(ctx, input.PlanID, input.PlanVersion)
	if err != nil {
		return applicationEvaluationCampaignFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureNotFound)
	}
	if version.PlanDigest != input.PlanDigest || version.ExecutionProfile != plan.ExecutionProfile {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureVersionConflict)
	}
	if service.resolveAuthority == nil || service.invoke == nil || service.readRun == nil {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureStoreUnavailable)
	}
	authority, failure := service.resolveAuthority(ctx, version)
	if failure != "" {
		return applicationEvaluationCampaignFailure(failure)
	}
	campaignID := applicationEvaluationDeterministicCampaignID(ctx, input.ClientCampaignKey)
	_, _, campaignSchema, supported := applicationEvaluationSchemaVersions(version.ExecutionProfile)
	if !supported {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureProfileIneligible)
	}
	campaign := ApplicationEvaluationCampaign{
		SchemaVersion: campaignSchema, CampaignID: campaignID,
		ClientCampaignKey: input.ClientCampaignKey, RecordVersion: 1,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		PlanID: version.PlanID, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest, ExecutionProfile: version.ExecutionProfile, QuotaAPIKeyID: input.QuotaAPIKeyID,
		State: applicationEvaluationCampaignStatePending, Items: make([]ApplicationEvaluationCampaignItem, 0, len(version.Items)),
		CreatedAt: service.currentTime().Format(time.RFC3339Nano), CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	for _, item := range version.Items {
		campaign.Items = append(campaign.Items, ApplicationEvaluationCampaignItem{
			ItemKey: item.ItemKey, RunID: applicationEvaluationDeterministicRunID(campaignID, item.ItemKey), State: applicationEvaluationCampaignItemPending,
		})
	}
	stored, inserted, err := service.repository.CreateCampaign(ctx, campaign)
	if err != nil {
		return applicationEvaluationCampaignFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !inserted {
		return applicationEvaluationCampaignSuccess(stored, true)
	}
	campaign = stored
	campaign.RecordVersion++
	campaign.State = applicationEvaluationCampaignStateRunning
	campaign.Authority = &authority
	campaign.StartedAt = service.currentTime().Format(time.RFC3339Nano)
	campaign.UpdatedByActorRef = ctx.ActorRef
	campaign.RequestID, campaign.AuditRef = ctx.RequestID, ctx.AuditRef
	campaign, ok, err := service.repository.UpdateCampaign(ctx, 1, campaign)
	if err != nil || !ok {
		return applicationEvaluationCampaignRepositoryResult(err)
	}
	for index, item := range version.Items {
		result, stop := service.executeItem(ctx, version, campaign, index, item)
		if result.Campaign == nil {
			return result
		}
		campaign = *result.Campaign
		if stop {
			return result
		}
	}
	return applicationEvaluationCampaignSuccess(campaign, false)
}

func (service applicationEvaluationCampaignService) executeItem(
	ctx ApplicationEvaluationContext,
	version ApplicationEvaluationPlanVersion,
	campaign ApplicationEvaluationCampaign,
	index int,
	item ApplicationEvaluationPlanItem,
) (ApplicationEvaluationCampaignResult, bool) {
	checkpoint, failure := service.resolveAuthority(ctx, version)
	if failure != "" || campaign.Authority == nil || checkpoint.AuthorityDigest != campaign.Authority.AuthorityDigest {
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateFailed, ApplicationEvaluationFailureAuthorityChanged, "Runtime authority changed before the next evaluation item."), true
	}
	if ctx.RequestContext.Err() != nil {
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateInterrupted, ApplicationEvaluationFailureRunUnavailable, "Evaluation request was canceled before the next item."), true
	}
	expected := campaign.RecordVersion
	campaign.RecordVersion++
	campaign.CurrentItemIndex = index
	campaign.Items[index].State = applicationEvaluationCampaignItemRunning
	campaign.Items[index].StartedAt = service.currentTime().Format(time.RFC3339Nano)
	stored, ok, err := service.repository.UpdateCampaign(ctx, expected, campaign)
	if err != nil || !ok {
		return applicationEvaluationCampaignRepositoryResult(err), true
	}
	campaign = stored
	invocationContext := ctx
	invocationContext.RequestContext = withGatewayRequestQuotaBinding(ctx.RequestContext, gatewayRequestQuotaBinding{
		QuotaContext: GatewayRequestQuotaContext{
			RequestContext: ctx.RequestContext, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment,
			ApplicationID: ctx.ApplicationID, ActorRef: ctx.ActorRef, RequestID: campaign.Items[index].RunID,
			AuditRef: "audit_" + campaign.Items[index].RunID + "_application-evaluation-quota",
		},
		APIKeyID: campaign.QuotaAPIKeyID, RequestID: campaign.Items[index].RunID, Route: applicationEvaluationInvocationRoute(version.ExecutionProfile),
	})
	run, invocationFailure, invocationSummary := service.invoke(invocationContext, version, item, campaign.Items[index].RunID)
	postCheckpoint, postFailure := service.resolveAuthority(ctx, version)
	if postFailure != "" || campaign.Authority == nil || postCheckpoint.AuthorityDigest != campaign.Authority.AuthorityDigest {
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateFailed, ApplicationEvaluationFailureAuthorityChanged, "Runtime authority changed while an evaluation item was executing."), true
	}
	if run == nil || run.RunID != campaign.Items[index].RunID || !applicationEvaluationRunMatchesCampaign(ctx, campaign, *run) {
		code := ApplicationEvaluationFailureRunUnavailable
		if invocationFailure != "" {
			code = invocationFailure
		}
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateFailed, code, invocationSummary), true
	}
	if run.Status == WorkflowRunStatusRunning {
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateInterrupted, ApplicationEvaluationFailureRunUnavailable, "Evaluation run did not reach a durable terminal state."), true
	}
	expected = campaign.RecordVersion
	campaign.RecordVersion++
	campaign.Items[index].RunSchemaVersion = run.SchemaVersion
	campaign.Items[index].RunProfile = campaign.ExecutionProfile
	campaign.Items[index].AuthorityDigest = campaign.Authority.AuthorityDigest
	campaign.Items[index].CompletedAt = run.CompletedAt
	if run.Status == WorkflowRunStatusSucceeded && invocationFailure == "" {
		campaign.Items[index].State = applicationEvaluationCampaignItemSucceeded
		campaign.SucceededItems++
		if index == len(version.Items)-1 {
			campaign.State = applicationEvaluationCampaignStateSucceeded
			campaign.CompletedAt = service.currentTime().Format(time.RFC3339Nano)
		}
	} else {
		campaign.Items[index].State = applicationEvaluationCampaignItemFailed
		campaign.Items[index].FailureCode = firstApplicationEvaluationFailure(invocationFailure, string(run.FailureCode), ApplicationEvaluationFailureRunUnavailable)
		campaign.Items[index].FailureBoundary = applicationEvaluationRunFailureBoundary(*run)
		campaign.FailedItems++
		campaign.State = applicationEvaluationCampaignStateFailed
		campaign.FailureCode = campaign.Items[index].FailureCode
		campaign.FailureSummary = firstApplicationEvaluationSummary(invocationSummary, run.FailureSummary)
		campaign.CompletedAt = service.currentTime().Format(time.RFC3339Nano)
	}
	stored, ok, err = service.repository.UpdateCampaign(ctx, expected, campaign)
	if err != nil || !ok {
		return applicationEvaluationCampaignRepositoryResult(err), true
	}
	result := applicationEvaluationCampaignSuccess(stored, false)
	if stored.State == applicationEvaluationCampaignStateFailed || stored.State == applicationEvaluationCampaignStateInterrupted {
		result.FailureCode, result.FailureSummary = stored.FailureCode, stored.FailureSummary
	}
	return result, stored.State != applicationEvaluationCampaignStateRunning
}

func (service applicationEvaluationCampaignService) Reconcile(ctx ApplicationEvaluationContext, campaignID string, expectedVersion int) ApplicationEvaluationCampaignResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationCampaignFailure(failure)
	}
	campaignID = strings.TrimSpace(campaignID)
	if !applicationEvaluationCampaignIDPattern.MatchString(campaignID) || expectedVersion < 1 {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	campaign, found, err := service.repository.ReadCampaign(ctx, campaignID)
	if err != nil {
		return applicationEvaluationCampaignFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureNotFound)
	}
	if campaign.RecordVersion != expectedVersion {
		return applicationEvaluationCampaignConflict(campaign.RecordVersion, campaign.State)
	}
	if campaign.State != applicationEvaluationCampaignStateRunning {
		return applicationEvaluationCampaignSuccess(campaign, true)
	}
	index := campaign.CurrentItemIndex
	if index < 0 || index >= len(campaign.Items) || campaign.Items[index].State != applicationEvaluationCampaignItemRunning {
		return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateInterrupted, ApplicationEvaluationFailureRunUnavailable, "Campaign checkpoint does not identify a running item.")
	}
	run, found, runErr := service.readRun(ctx, campaign.Items[index].RunID)
	if runErr != nil {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureStoreUnavailable)
	}
	if found && run.RunID == campaign.Items[index].RunID && run.Status != WorkflowRunStatusRunning && applicationEvaluationRunMatchesCampaign(ctx, campaign, run) {
		campaign.Items[index].RunSchemaVersion = run.SchemaVersion
		campaign.Items[index].RunProfile = campaign.ExecutionProfile
		campaign.Items[index].AuthorityDigest = campaign.Authority.AuthorityDigest
		campaign.Items[index].CompletedAt = run.CompletedAt
		campaign.Items[index].State = applicationEvaluationCampaignItemFailed
		campaign.Items[index].FailureCode = firstApplicationEvaluationFailure(string(run.FailureCode), ApplicationEvaluationFailureRunUnavailable)
		campaign.Items[index].FailureBoundary = applicationEvaluationRunFailureBoundary(run)
		campaign.FailedItems++
	}
	return service.finishCampaign(ctx, campaign, applicationEvaluationCampaignStateInterrupted, ApplicationEvaluationFailureRunUnavailable, "Interrupted campaign was reconciled without replaying provider execution.")
}

func (service applicationEvaluationCampaignService) Read(ctx ApplicationEvaluationContext, campaignID string) ApplicationEvaluationCampaignResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureScopeDenied)
	}
	campaignID = strings.TrimSpace(campaignID)
	if !applicationEvaluationCampaignIDPattern.MatchString(campaignID) {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	campaign, found, err := service.repository.ReadCampaign(ctx, campaignID)
	if err != nil {
		return applicationEvaluationCampaignFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationCampaignFailure(ApplicationEvaluationFailureNotFound)
	}
	return applicationEvaluationCampaignSuccess(campaign, false)
}

func (service applicationEvaluationCampaignService) List(ctx ApplicationEvaluationContext, input ApplicationEvaluationCampaignListInput) ApplicationEvaluationCampaignListResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationCampaignListFailure(ApplicationEvaluationFailureScopeDenied)
	}
	input.PlanID = strings.TrimSpace(input.PlanID)
	limit := input.Limit
	if limit == 0 {
		limit = applicationEvaluationDefaultListLimit
	}
	if input.PlanID != "" && !applicationEvaluationPlanIDPattern.MatchString(input.PlanID) || limit < 1 || limit > applicationEvaluationMaximumListLimit {
		return applicationEvaluationCampaignListFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	filter := ApplicationEvaluationCampaignListFilter{PlanID: input.PlanID, Limit: limit}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationEvaluationCursor(input.Cursor)
		if err != nil || !applicationEvaluationCursorMatches(ctx, cursor, "campaigns", input.PlanID, limit) ||
			!applicationEvaluationCampaignIDPattern.MatchString(cursor.BeforeID) || parseApplicationEvaluationTimestamp(cursor.BeforeTime) == nil {
			return applicationEvaluationCampaignListFailure(ApplicationEvaluationFailureCursorInvalid)
		}
		filter.BeforeCreatedAt, filter.BeforeCampaignID = cursor.BeforeTime, cursor.BeforeID
	}
	page, err := service.repository.ListCampaigns(ctx, filter)
	if err != nil {
		return applicationEvaluationCampaignListFailure(applicationEvaluationRepositoryFailure(err))
	}
	result := ApplicationEvaluationCampaignListResult{Campaigns: page.Campaigns, HasMore: page.HasMore}
	if page.HasMore && len(page.Campaigns) > 0 {
		last := page.Campaigns[len(page.Campaigns)-1]
		result.NextCursor, _ = encodeApplicationEvaluationCursor(applicationEvaluationCursor{
			Version: 1, Kind: "campaigns", TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			Environment: ctx.Environment, ApplicationID: ctx.ApplicationID, Filter: input.PlanID,
			BeforeTime: last.CreatedAt, BeforeID: last.CampaignID, Limit: limit,
		})
	}
	return result
}

func (service applicationEvaluationCampaignService) finishCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign, state, code, summary string) ApplicationEvaluationCampaignResult {
	expected := campaign.RecordVersion
	campaign.RecordVersion++
	campaign.State = state
	campaign.FailureCode = code
	campaign.FailureSummary = firstApplicationEvaluationSummary(summary, applicationEvaluationFailureSummary(code))
	campaign.CompletedAt = service.currentTime().Format(time.RFC3339Nano)
	campaign.UpdatedByActorRef, campaign.RequestID, campaign.AuditRef = ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	stored, ok, err := service.repository.UpdateCampaign(ctx, expected, campaign)
	if err != nil || !ok {
		return applicationEvaluationCampaignRepositoryResult(err)
	}
	result := applicationEvaluationCampaignSuccess(stored, false)
	result.FailureCode, result.FailureSummary = code, campaign.FailureSummary
	return result
}

func (service applicationEvaluationCampaignService) currentTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func applicationEvaluationRunMatchesCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign, run WorkflowRunRecord) bool {
	if run.RunID == "" || run.TenantRef != "" && run.TenantRef != ctx.TenantRef || run.WorkspaceID != ctx.WorkspaceID || run.ApplicationID != ctx.ApplicationID || campaign.Authority == nil {
		return false
	}
	switch campaign.ExecutionProfile {
	case applicationInteractionProfileWorkflow:
		var snapshot ApplicationInteractionAuthoritySnapshot
		return run.SchemaVersion == workflowRunRecordDefinitionSchemaVersion && run.ExecutionProfile == applicationInteractionProfileWorkflow && run.DefinitionAuthority != nil &&
			decodeStrictApplicationEvaluationJSON(campaign.Authority.Snapshot, &snapshot) == nil && snapshot.WorkflowDefinition != nil &&
			run.DefinitionAuthority.DefinitionID == snapshot.WorkflowDefinition.DefinitionID && run.DefinitionAuthority.DefinitionVersion == snapshot.WorkflowDefinition.DefinitionVersion &&
			run.DefinitionAuthority.DefinitionDigest == snapshot.WorkflowDefinition.DefinitionDigest && run.DefinitionAuthority.ActivationPointerVersion == snapshot.WorkflowDefinition.ActivationPointerVersion &&
			run.DefinitionAuthority.ApplicationRecordVersion == snapshot.ApplicationRecordVersion
	case applicationInteractionProfileWorkflowStructured:
		var snapshot ApplicationInteractionAuthoritySnapshot
		return run.SchemaVersion == workflowRunRecordDefinitionStructuredSchemaVersion && run.ExecutionProfile == applicationInteractionProfileWorkflowStructured &&
			run.DefinitionAuthority != nil && decodeStrictApplicationEvaluationJSON(campaign.Authority.Snapshot, &snapshot) == nil &&
			snapshot.WorkflowDefinition != nil && snapshot.WorkflowDefinition.InputContract != nil &&
			run.DefinitionAuthority.DefinitionID == snapshot.WorkflowDefinition.DefinitionID &&
			run.DefinitionAuthority.DefinitionVersion == snapshot.WorkflowDefinition.DefinitionVersion &&
			run.DefinitionAuthority.DefinitionDigest == snapshot.WorkflowDefinition.DefinitionDigest &&
			run.DefinitionAuthority.ActivationPointerVersion == snapshot.WorkflowDefinition.ActivationPointerVersion &&
			run.DefinitionAuthority.ApplicationRecordVersion == snapshot.ApplicationRecordVersion &&
			run.InputContractID == snapshot.WorkflowDefinition.InputContract.ContractID &&
			run.InputContractDigest == snapshot.WorkflowDefinition.InputContract.ContractDigest
	case applicationInteractionProfileRAG:
		var snapshot ApplicationInteractionAuthoritySnapshot
		return run.SchemaVersion == workflowRunRecordAppRAGSchemaVersion && run.RAGApplication != nil && run.RAGSnapshot != nil &&
			decodeStrictApplicationEvaluationJSON(campaign.Authority.Snapshot, &snapshot) == nil && snapshot.ApplicationRAG != nil &&
			run.RAGApplication.AssignmentID == snapshot.ApplicationRAG.AssignmentID && run.RAGApplication.AssignmentVersion == snapshot.ApplicationRAG.AssignmentVersion &&
			run.RAGApplication.AssignmentDigest == snapshot.ApplicationRAG.AssignmentDigest && run.RAGApplication.DraftID == snapshot.ApplicationRAG.DraftID &&
			run.RAGApplication.DraftVersion == snapshot.ApplicationRAG.DraftVersion && run.RAGApplication.DraftDigest == snapshot.ApplicationRAG.DraftDigest &&
			run.RAGSnapshot.SnapshotID == snapshot.ApplicationRAG.SnapshotID && run.RAGSnapshot.SnapshotVersion == snapshot.ApplicationRAG.SnapshotVersion &&
			run.RAGSnapshot.SnapshotDigest == snapshot.ApplicationRAG.SnapshotDigest
	case applicationInteractionProfilePrompt:
		var snapshot PromptApplicationRuntimeAuthorityV2
		return run.SchemaVersion == workflowRunRecordPromptSchemaVersion && run.ExecutionProfile == applicationInteractionProfilePrompt && run.PromptApplication != nil &&
			decodeStrictApplicationEvaluationJSON(campaign.Authority.Snapshot, &snapshot) == nil && run.PromptApplication.AuthorityDigest == snapshot.AuthorityDigest
	case applicationInteractionProfileAgentCopilot:
		var snapshot AgentCopilotRuntimeAuthorityV3
		return run.SchemaVersion == agentCopilotRunV7Schema && run.ExecutionProfile == applicationInteractionProfileAgentCopilot && run.AgentCopilotAuthority != nil &&
			decodeStrictApplicationEvaluationJSON(campaign.Authority.Snapshot, &snapshot) == nil && run.AgentCopilotAuthority.AuthorityDigest == snapshot.AuthorityDigest
	default:
		return false
	}
}

func applicationEvaluationRunFailureBoundary(run WorkflowRunRecord) string {
	if run.Diagnostic != nil {
		return string(run.Diagnostic.FailureBoundary)
	}
	return "run"
}

func firstApplicationEvaluationFailure(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ApplicationEvaluationFailureRunUnavailable
}

func firstApplicationEvaluationSummary(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return applicationEvaluationFailureSummary(ApplicationEvaluationFailureRunUnavailable)
}

func applicationEvaluationInvocationRoute(profile string) string {
	switch profile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileWorkflowStructured:
		return workflowDefinitionRunCreateRoute
	case applicationInteractionProfileRAG:
		return "POST " + workflowRAGApplicationInvocationRoute
	case applicationInteractionProfilePrompt:
		return "POST " + promptApplicationInvocationRoute
	case applicationInteractionProfileAgentCopilot:
		return "POST " + agentCopilotInvocationRoute
	default:
		return applicationEvaluationCampaignExecuteRoute
	}
}

func applicationEvaluationCampaignSuccess(campaign ApplicationEvaluationCampaign, replay bool) ApplicationEvaluationCampaignResult {
	copy := cloneApplicationEvaluationCampaign(campaign)
	return ApplicationEvaluationCampaignResult{Campaign: &copy, IdempotentReplay: replay, CurrentRecordVersion: copy.RecordVersion, CurrentState: copy.State}
}

func applicationEvaluationCampaignFailure(code string) ApplicationEvaluationCampaignResult {
	return ApplicationEvaluationCampaignResult{FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func applicationEvaluationCampaignConflict(version int, state string) ApplicationEvaluationCampaignResult {
	return applicationEvaluationCampaignConflictWithCode(version, state, ApplicationEvaluationFailureVersionConflict)
}

func applicationEvaluationCampaignConflictWithCode(version int, state, code string) ApplicationEvaluationCampaignResult {
	result := applicationEvaluationCampaignFailure(code)
	result.CurrentRecordVersion, result.CurrentState = version, state
	return result
}

func applicationEvaluationCampaignRepositoryResult(err error) ApplicationEvaluationCampaignResult {
	code := applicationEvaluationRepositoryFailure(err)
	result := applicationEvaluationCampaignFailure(code)
	var conflict applicationEvaluationVersionConflictError
	if errors.As(err, &conflict) {
		result.CurrentRecordVersion, result.CurrentState = conflict.CurrentVersion, conflict.CurrentState
	}
	return result
}

func applicationEvaluationCampaignListFailure(code string) ApplicationEvaluationCampaignListResult {
	return ApplicationEvaluationCampaignListResult{Campaigns: []ApplicationEvaluationCampaign{}, FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func applicationEvaluationCampaignAuthorityPayload(profile string, value any, digest string) (ApplicationEvaluationCampaignAuthority, string) {
	payload, err := json.Marshal(value)
	if err != nil || !workflowRAGDigestPattern.MatchString(digest) {
		return ApplicationEvaluationCampaignAuthority{}, ApplicationEvaluationFailureStoreContract
	}
	authority := ApplicationEvaluationCampaignAuthority{ExecutionProfile: profile, AuthorityDigest: digest, Snapshot: payload}
	if validateApplicationEvaluationCampaignAuthority(authority) != nil {
		return ApplicationEvaluationCampaignAuthority{}, ApplicationEvaluationFailureStoreContract
	}
	return authority, ""
}
