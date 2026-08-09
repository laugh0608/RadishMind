package httpapi

import (
	"errors"
	"strings"
	"time"
)

type ApplicationEvaluationPlanCreateInput struct {
	Name             string
	ExecutionProfile string
	Target           ApplicationEvaluationPlanTarget
	Items            []ApplicationEvaluationPlanItem
}

type ApplicationEvaluationPlanReviseInput struct {
	ExpectedVersion  int
	Name             string
	ExecutionProfile string
	Target           ApplicationEvaluationPlanTarget
	Items            []ApplicationEvaluationPlanItem
}

type ApplicationEvaluationPlanArchiveInput struct {
	ExpectedVersion           int
	AcknowledgeNoNewCampaigns bool
}

type ApplicationEvaluationPlanListInput struct {
	LifecycleState string
	Limit          int
	Cursor         string
}

type ApplicationEvaluationVersionListInput struct {
	Limit  int
	Cursor string
}

type ApplicationEvaluationPlanResult struct {
	Plan                 *ApplicationEvaluationPlan
	Version              *ApplicationEvaluationPlanVersion
	FailureCode          string
	FailureSummary       string
	CurrentRecordVersion int
	CurrentState         string
}

type ApplicationEvaluationPlanListResult struct {
	Plans          []ApplicationEvaluationPlan
	NextCursor     string
	HasMore        bool
	FailureCode    string
	FailureSummary string
}

type ApplicationEvaluationVersionListResult struct {
	Versions       []ApplicationEvaluationPlanVersion
	NextCursor     string
	HasMore        bool
	FailureCode    string
	FailureSummary string
}

type applicationEvaluationPlanService struct {
	repository   applicationEvaluationRepository
	applications applicationCatalogRepository
	now          func() time.Time
	newPlanID    func() (string, error)
}

func newApplicationEvaluationPlanService(repository applicationEvaluationRepository, applications applicationCatalogRepository) applicationEvaluationPlanService {
	return applicationEvaluationPlanService{
		repository: repository, applications: applications,
		now:       func() time.Time { return time.Now().UTC() },
		newPlanID: func() (string, error) { return newApplicationEvaluationID("aeplan_") },
	}
}

func (service applicationEvaluationPlanService) Create(ctx ApplicationEvaluationContext, input ApplicationEvaluationPlanCreateInput) ApplicationEvaluationPlanResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	name, profile, target, items, failure := normalizeApplicationEvaluationPlanDefinition(input.Name, input.ExecutionProfile, input.Target, input.Items)
	if failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	if failure = service.requireApplicationProfile(ctx, profile); failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	for attempt := 0; attempt < 3; attempt++ {
		planID, err := service.newPlanID()
		if err != nil || !applicationEvaluationPlanIDPattern.MatchString(planID) {
			return applicationEvaluationPlanFailure(ApplicationEvaluationFailureStoreUnavailable)
		}
		now := service.currentTime().Format(time.RFC3339Nano)
		version := ApplicationEvaluationPlanVersion{
			SchemaVersion: applicationEvaluationPlanVersionSchemaVersion,
			PlanID:        planID, PlanVersion: 1, PreviousPlanVersion: 0,
			TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
			Name: name, ExecutionProfile: profile, Target: target, Items: items,
			CreatedAt: now, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		}
		version.PlanDigest, err = applicationEvaluationPlanDigest(version)
		if err != nil {
			return applicationEvaluationPlanFailure(ApplicationEvaluationFailureStoreUnavailable)
		}
		plan := ApplicationEvaluationPlan{
			SchemaVersion: applicationEvaluationPlanSchemaVersion, PlanID: planID, RecordVersion: 1,
			LatestPlanVersion: 1, LatestPlanDigest: version.PlanDigest,
			TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
			Name: name, ExecutionProfile: profile, ItemCount: len(items), LifecycleState: applicationEvaluationPlanStateActive,
			CreatedAt: now, UpdatedAt: now, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
			RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
		}
		err = service.repository.CreatePlan(ctx, plan, version)
		if errors.Is(err, errApplicationEvaluationVersionConflict) {
			continue
		}
		if err != nil {
			return applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
		}
		return applicationEvaluationPlanSuccess(plan, &version)
	}
	return applicationEvaluationPlanFailure(ApplicationEvaluationFailureStoreUnavailable)
}

func (service applicationEvaluationPlanService) Revise(ctx ApplicationEvaluationContext, planID string, input ApplicationEvaluationPlanReviseInput) ApplicationEvaluationPlanResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	planID = strings.TrimSpace(planID)
	if !applicationEvaluationPlanIDPattern.MatchString(planID) || input.ExpectedVersion < 1 {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	name, profile, target, items, failure := normalizeApplicationEvaluationPlanDefinition(input.Name, input.ExecutionProfile, input.Target, input.Items)
	if failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	if failure = service.requireApplicationProfile(ctx, profile); failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	current, found, err := service.repository.ReadPlan(ctx, planID)
	if err != nil {
		return applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	if current.RecordVersion != input.ExpectedVersion {
		return applicationEvaluationPlanConflict(current)
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return applicationEvaluationPlanConflictWithCode(current, ApplicationEvaluationFailureArchived)
	}
	now := service.currentTime().Format(time.RFC3339Nano)
	version := ApplicationEvaluationPlanVersion{
		SchemaVersion: applicationEvaluationPlanVersionSchemaVersion, PlanID: planID,
		PlanVersion: current.LatestPlanVersion + 1, PreviousPlanVersion: current.LatestPlanVersion,
		TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, Environment: ctx.Environment, ApplicationID: ctx.ApplicationID,
		Name: name, ExecutionProfile: profile, Target: target, Items: items,
		CreatedAt: now, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	version.PlanDigest, err = applicationEvaluationPlanDigest(version)
	if err != nil {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureStoreUnavailable)
	}
	updated := current
	updated.RecordVersion++
	updated.LatestPlanVersion = version.PlanVersion
	updated.LatestPlanDigest = version.PlanDigest
	updated.Name = name
	updated.ExecutionProfile = profile
	updated.ItemCount = len(items)
	updated.UpdatedAt = now
	updated.UpdatedByActorRef = ctx.ActorRef
	updated.RequestID = ctx.RequestID
	updated.AuditRef = ctx.AuditRef
	stored, ok, err := service.repository.RevisePlan(ctx, input.ExpectedVersion, updated, version)
	if err != nil {
		result := applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
		var conflict applicationEvaluationVersionConflictError
		if errors.As(err, &conflict) {
			result.CurrentRecordVersion, result.CurrentState = conflict.CurrentVersion, conflict.CurrentState
		}
		return result
	}
	if !ok {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	return applicationEvaluationPlanSuccess(stored, &version)
}

func (service applicationEvaluationPlanService) Archive(ctx ApplicationEvaluationContext, planID string, input ApplicationEvaluationPlanArchiveInput) ApplicationEvaluationPlanResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationPlanFailure(failure)
	}
	planID = strings.TrimSpace(planID)
	if !applicationEvaluationPlanIDPattern.MatchString(planID) || input.ExpectedVersion < 1 || !input.AcknowledgeNoNewCampaigns {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	current, found, err := service.repository.ReadPlan(ctx, planID)
	if err != nil {
		return applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	if current.RecordVersion != input.ExpectedVersion {
		return applicationEvaluationPlanConflict(current)
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return applicationEvaluationPlanConflictWithCode(current, ApplicationEvaluationFailureArchived)
	}
	updated := current
	updated.RecordVersion++
	updated.LifecycleState = applicationEvaluationPlanStateArchived
	updated.UpdatedAt = service.currentTime().Format(time.RFC3339Nano)
	updated.UpdatedByActorRef = ctx.ActorRef
	updated.RequestID = ctx.RequestID
	updated.AuditRef = ctx.AuditRef
	stored, ok, err := service.repository.ArchivePlan(ctx, input.ExpectedVersion, updated)
	if err != nil {
		result := applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
		var conflict applicationEvaluationVersionConflictError
		if errors.As(err, &conflict) {
			result.CurrentRecordVersion, result.CurrentState = conflict.CurrentVersion, conflict.CurrentState
		}
		return result
	}
	if !ok {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	return applicationEvaluationPlanSuccess(stored, nil)
}

func (service applicationEvaluationPlanService) Read(ctx ApplicationEvaluationContext, planID string) ApplicationEvaluationPlanResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureScopeDenied)
	}
	planID = strings.TrimSpace(planID)
	if !applicationEvaluationPlanIDPattern.MatchString(planID) {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	plan, found, err := service.repository.ReadPlan(ctx, planID)
	if err != nil {
		return applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	return applicationEvaluationPlanSuccess(plan, nil)
}

func (service applicationEvaluationPlanService) ReadVersion(ctx ApplicationEvaluationContext, planID string, version int) ApplicationEvaluationPlanResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureScopeDenied)
	}
	planID = strings.TrimSpace(planID)
	if !applicationEvaluationPlanIDPattern.MatchString(planID) || version < 1 {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	value, found, err := service.repository.ReadPlanVersion(ctx, planID, version)
	if err != nil {
		return applicationEvaluationPlanFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPlanFailure(ApplicationEvaluationFailureNotFound)
	}
	return ApplicationEvaluationPlanResult{Version: &value}
}

func (service applicationEvaluationPlanService) List(ctx ApplicationEvaluationContext, input ApplicationEvaluationPlanListInput) ApplicationEvaluationPlanListResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationPlanListFailure(ApplicationEvaluationFailureScopeDenied)
	}
	lifecycle := strings.TrimSpace(input.LifecycleState)
	if lifecycle == "" {
		lifecycle = applicationEvaluationPlanStateActive
	}
	limit := input.Limit
	if limit == 0 {
		limit = applicationEvaluationDefaultListLimit
	}
	if (lifecycle != applicationEvaluationPlanStateActive && lifecycle != applicationEvaluationPlanStateArchived) || limit < 1 || limit > applicationEvaluationMaximumListLimit {
		return applicationEvaluationPlanListFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	filter := ApplicationEvaluationPlanListFilter{LifecycleState: lifecycle, Limit: limit}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationEvaluationCursor(input.Cursor)
		if err != nil || !applicationEvaluationCursorMatches(ctx, cursor, "plans", lifecycle, limit) ||
			!applicationEvaluationPlanIDPattern.MatchString(cursor.BeforeID) || parseApplicationEvaluationTimestamp(cursor.BeforeTime) == nil {
			return applicationEvaluationPlanListFailure(ApplicationEvaluationFailureCursorInvalid)
		}
		filter.BeforeUpdatedAt, filter.BeforePlanID = cursor.BeforeTime, cursor.BeforeID
	}
	page, err := service.repository.ListPlans(ctx, filter)
	if err != nil {
		return applicationEvaluationPlanListFailure(applicationEvaluationRepositoryFailure(err))
	}
	result := ApplicationEvaluationPlanListResult{Plans: page.Plans, HasMore: page.HasMore}
	if page.HasMore && len(page.Plans) > 0 {
		last := page.Plans[len(page.Plans)-1]
		result.NextCursor, _ = encodeApplicationEvaluationCursor(applicationEvaluationCursor{
			Version: 1, Kind: "plans", TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			Environment: ctx.Environment, ApplicationID: ctx.ApplicationID, Filter: lifecycle,
			BeforeTime: last.UpdatedAt, BeforeID: last.PlanID, Limit: limit,
		})
	}
	return result
}

func (service applicationEvaluationPlanService) ListVersions(ctx ApplicationEvaluationContext, planID string, input ApplicationEvaluationVersionListInput) ApplicationEvaluationVersionListResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationVersionListFailure(ApplicationEvaluationFailureScopeDenied)
	}
	planID = strings.TrimSpace(planID)
	limit := input.Limit
	if limit == 0 {
		limit = applicationEvaluationDefaultListLimit
	}
	if !applicationEvaluationPlanIDPattern.MatchString(planID) || limit < 1 || limit > applicationEvaluationMaximumListLimit {
		return applicationEvaluationVersionListFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	filter := ApplicationEvaluationVersionListFilter{Limit: limit}
	if strings.TrimSpace(input.Cursor) != "" {
		cursor, err := decodeApplicationEvaluationCursor(input.Cursor)
		if err != nil || !applicationEvaluationCursorMatches(ctx, cursor, "versions", planID, limit) || cursor.BeforeVersion < 1 {
			return applicationEvaluationVersionListFailure(ApplicationEvaluationFailureCursorInvalid)
		}
		filter.BeforeVersion = cursor.BeforeVersion
	}
	page, err := service.repository.ListPlanVersions(ctx, planID, filter)
	if err != nil {
		return applicationEvaluationVersionListFailure(applicationEvaluationRepositoryFailure(err))
	}
	result := ApplicationEvaluationVersionListResult{Versions: page.Versions, HasMore: page.HasMore}
	if page.HasMore && len(page.Versions) > 0 {
		last := page.Versions[len(page.Versions)-1]
		result.NextCursor, _ = encodeApplicationEvaluationCursor(applicationEvaluationCursor{
			Version: 1, Kind: "versions", TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
			Environment: ctx.Environment, ApplicationID: ctx.ApplicationID, Filter: planID,
			BeforeVersion: last.PlanVersion, Limit: limit,
		})
	}
	return result
}

func (service applicationEvaluationPlanService) requireApplicationProfile(ctx ApplicationEvaluationContext, profile string) string {
	if service.applications == nil {
		return ApplicationEvaluationFailureStoreUnavailable
	}
	application, err := service.applications.RequireActive(ApplicationCatalogContext{
		RequestContext: ctx.RequestContext, RequestID: ctx.RequestID, TenantRef: ctx.TenantRef,
		WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.ActorRef, AuditRef: ctx.AuditRef,
	}, ctx.ApplicationID)
	if errors.Is(err, errApplicationCatalogNotFound) {
		return ApplicationEvaluationFailureNotFound
	}
	if errors.Is(err, errApplicationCatalogArchived) {
		return ApplicationEvaluationFailureArchived
	}
	if err != nil {
		return ApplicationEvaluationFailureStoreUnavailable
	}
	eligible := false
	switch profile {
	case applicationInteractionProfileWorkflow, applicationInteractionProfileRAG:
		eligible = application.ApplicationKind == "workflow_copilot" || application.ApplicationKind == "docs_qa"
	case applicationInteractionProfilePrompt:
		eligible = application.ApplicationKind == "prompt_application"
	case applicationInteractionProfileAgentCopilot:
		eligible = application.ApplicationKind == "agent"
	}
	if !eligible {
		return ApplicationEvaluationFailureProfileIneligible
	}
	return ""
}

func (service applicationEvaluationPlanService) currentTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func validateApplicationEvaluationMutationContext(ctx ApplicationEvaluationContext) string {
	if !validApplicationEvaluationContext(ctx) {
		if !validApplicationEvaluationEnvironment(ctx.Environment) {
			return ApplicationEvaluationFailureEnvironmentDenied
		}
		return ApplicationEvaluationFailureScopeDenied
	}
	if !ctx.WriteEnabled {
		return ApplicationEvaluationFailureWriteDisabled
	}
	return ""
}

func applicationEvaluationCursorMatches(ctx ApplicationEvaluationContext, cursor applicationEvaluationCursor, kind, filter string, limit int) bool {
	return cursor.Kind == kind && cursor.TenantRef == ctx.TenantRef && cursor.WorkspaceID == ctx.WorkspaceID &&
		cursor.Environment == ctx.Environment && cursor.ApplicationID == ctx.ApplicationID && cursor.Filter == filter && cursor.Limit == limit
}

func applicationEvaluationPlanSuccess(plan ApplicationEvaluationPlan, version *ApplicationEvaluationPlanVersion) ApplicationEvaluationPlanResult {
	result := ApplicationEvaluationPlanResult{Plan: &plan, CurrentRecordVersion: plan.RecordVersion, CurrentState: plan.LifecycleState}
	if version != nil {
		copy := cloneApplicationEvaluationPlanVersion(*version)
		result.Version = &copy
	}
	return result
}

func applicationEvaluationPlanConflict(plan ApplicationEvaluationPlan) ApplicationEvaluationPlanResult {
	return applicationEvaluationPlanConflictWithCode(plan, ApplicationEvaluationFailureVersionConflict)
}

func applicationEvaluationPlanConflictWithCode(plan ApplicationEvaluationPlan, code string) ApplicationEvaluationPlanResult {
	result := applicationEvaluationPlanFailure(code)
	result.CurrentRecordVersion, result.CurrentState = plan.RecordVersion, plan.LifecycleState
	return result
}

func applicationEvaluationPlanFailure(code string) ApplicationEvaluationPlanResult {
	return ApplicationEvaluationPlanResult{FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func applicationEvaluationPlanListFailure(code string) ApplicationEvaluationPlanListResult {
	return ApplicationEvaluationPlanListResult{Plans: []ApplicationEvaluationPlan{}, FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func applicationEvaluationVersionListFailure(code string) ApplicationEvaluationVersionListResult {
	return ApplicationEvaluationVersionListResult{Versions: []ApplicationEvaluationPlanVersion{}, FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func applicationEvaluationFailureSummary(code string) string {
	switch code {
	case ApplicationEvaluationFailureScopeDenied:
		return "Application evaluation scope is denied."
	case ApplicationEvaluationFailureEnvironmentDenied:
		return "Application evaluation is restricted to the configured development or test environment."
	case ApplicationEvaluationFailureNotFound:
		return "Application evaluation record was not found in the current scope."
	case ApplicationEvaluationFailurePayloadInvalid:
		return "Application evaluation input is invalid."
	case ApplicationEvaluationFailureSecretForbidden:
		return "Application evaluation fixtures must not contain secret material."
	case ApplicationEvaluationFailureProfileIneligible:
		return "Application kind or runtime authority is not eligible for the selected evaluation profile."
	case ApplicationEvaluationFailureVersionConflict:
		return "Application evaluation changed; reload the current version before updating."
	case ApplicationEvaluationFailureArchived:
		return "Archived application evaluation plans cannot be revised or executed."
	case ApplicationEvaluationFailureCursorInvalid:
		return "Application evaluation cursor is invalid for the current scope and filter."
	case ApplicationEvaluationFailureCampaignConflict:
		return "The campaign idempotency key is already bound to a different plan version."
	case ApplicationEvaluationFailureAuthorityChanged:
		return "Application runtime authority changed during the evaluation campaign."
	case ApplicationEvaluationFailureRunUnavailable:
		return "Application evaluation could not produce or read the expected durable run."
	case ApplicationEvaluationFailureQuotaConsumerInvalid:
		return "The selected application API key is not an active quota consumer in the current scope."
	case ApplicationEvaluationFailureHandoffPartial:
		return "Application evaluation handoff completed only partially; inspect the persisted exact references."
	case ApplicationEvaluationFailureWriteDisabled:
		return "Application evaluation writes require explicit development opt-in."
	case ApplicationEvaluationFailureStoreContract, ApplicationEvaluationFailureStoreUnavailable:
		return "Application evaluation storage is unavailable."
	default:
		return "Application evaluation request failed."
	}
}
