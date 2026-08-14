package httpapi

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type applicationEvaluationRepository interface {
	CreatePlan(ApplicationEvaluationContext, ApplicationEvaluationPlan, ApplicationEvaluationPlanVersion) error
	RevisePlan(ApplicationEvaluationContext, int, ApplicationEvaluationPlan, ApplicationEvaluationPlanVersion) (ApplicationEvaluationPlan, bool, error)
	ArchivePlan(ApplicationEvaluationContext, int, ApplicationEvaluationPlan) (ApplicationEvaluationPlan, bool, error)
	ReadPlan(ApplicationEvaluationContext, string) (ApplicationEvaluationPlan, bool, error)
	ListPlans(ApplicationEvaluationContext, ApplicationEvaluationPlanListFilter) (ApplicationEvaluationPlanListPage, error)
	ReadPlanVersion(ApplicationEvaluationContext, string, int) (ApplicationEvaluationPlanVersion, bool, error)
	ListPlanVersions(ApplicationEvaluationContext, string, ApplicationEvaluationVersionListFilter) (ApplicationEvaluationVersionListPage, error)
	CreateCampaign(ApplicationEvaluationContext, ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error)
	UpdateCampaign(ApplicationEvaluationContext, int, ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error)
	ReadCampaign(ApplicationEvaluationContext, string) (ApplicationEvaluationCampaign, bool, error)
	ListCampaigns(ApplicationEvaluationContext, ApplicationEvaluationCampaignListFilter) (ApplicationEvaluationCampaignListPage, error)
}

type memoryApplicationEvaluationRepository struct {
	mu               sync.RWMutex
	planCapacity     int
	campaignCapacity int
	plans            map[string]ApplicationEvaluationPlan
	versions         map[string]map[int]ApplicationEvaluationPlanVersion
	planOrder        []string
	campaigns        map[string]ApplicationEvaluationCampaign
	campaignKeys     map[string]string
	campaignOrder    []string
	unavailable      bool
}

func newMemoryApplicationEvaluationRepository(planCapacity, campaignCapacity int) *memoryApplicationEvaluationRepository {
	if planCapacity <= 0 {
		planCapacity = applicationEvaluationMemoryPlanCapacity
	}
	if campaignCapacity <= 0 {
		campaignCapacity = applicationEvaluationMemoryRunCapacity
	}
	return &memoryApplicationEvaluationRepository{
		planCapacity: planCapacity, campaignCapacity: campaignCapacity,
		plans:        make(map[string]ApplicationEvaluationPlan, planCapacity),
		versions:     make(map[string]map[int]ApplicationEvaluationPlanVersion, planCapacity),
		campaigns:    make(map[string]ApplicationEvaluationCampaign, campaignCapacity),
		campaignKeys: make(map[string]string, campaignCapacity),
	}
}

func (repository *memoryApplicationEvaluationRepository) CreatePlan(
	ctx ApplicationEvaluationContext,
	plan ApplicationEvaluationPlan,
	version ApplicationEvaluationPlanVersion,
) error {
	if validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.PlanID != version.PlanID || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest ||
		plan.Name != version.Name || plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return errApplicationEvaluationStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationPlanKey(ctx, plan.PlanID)
	if _, exists := repository.plans[key]; exists {
		return applicationEvaluationVersionConflictError{CurrentVersion: repository.plans[key].RecordVersion, CurrentState: repository.plans[key].LifecycleState}
	}
	repository.plans[key] = cloneApplicationEvaluationPlan(plan)
	repository.versions[key] = map[int]ApplicationEvaluationPlanVersion{version.PlanVersion: cloneApplicationEvaluationPlanVersion(version)}
	repository.planOrder = append(repository.planOrder, key)
	for len(repository.planOrder) > repository.planCapacity {
		evicted := repository.planOrder[0]
		delete(repository.plans, evicted)
		delete(repository.versions, evicted)
		repository.planOrder = repository.planOrder[1:]
	}
	return nil
}

func (repository *memoryApplicationEvaluationRepository) RevisePlan(
	ctx ApplicationEvaluationContext,
	expected int,
	plan ApplicationEvaluationPlan,
	version ApplicationEvaluationPlanVersion,
) (ApplicationEvaluationPlan, bool, error) {
	if expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || validateApplicationEvaluationPlanVersion(ctx, version) != nil ||
		plan.RecordVersion != expected+1 || plan.LatestPlanVersion != version.PlanVersion || plan.LatestPlanDigest != version.PlanDigest ||
		plan.PlanID != version.PlanID || plan.Name != version.Name || plan.ExecutionProfile != version.ExecutionProfile || plan.ItemCount != len(version.Items) {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationPlanKey(ctx, plan.PlanID)
	current, found := repository.plans[key]
	if !found {
		return ApplicationEvaluationPlan{}, false, nil
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationPlan(current), false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return cloneApplicationEvaluationPlan(current), false, errApplicationEvaluationArchived
	}
	if version.PlanVersion != current.LatestPlanVersion+1 || version.PreviousPlanVersion != current.LatestPlanVersion ||
		plan.CreatedAt != current.CreatedAt || plan.CreatedByActorRef != current.CreatedByActorRef || plan.LifecycleState != applicationEvaluationPlanStateActive {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	if _, exists := repository.versions[key][version.PlanVersion]; exists {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	repository.plans[key] = cloneApplicationEvaluationPlan(plan)
	repository.versions[key][version.PlanVersion] = cloneApplicationEvaluationPlanVersion(version)
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *memoryApplicationEvaluationRepository) ArchivePlan(
	ctx ApplicationEvaluationContext,
	expected int,
	plan ApplicationEvaluationPlan,
) (ApplicationEvaluationPlan, bool, error) {
	if expected < 1 || validateApplicationEvaluationPlan(ctx, plan) != nil || plan.RecordVersion != expected+1 || plan.LifecycleState != applicationEvaluationPlanStateArchived {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationPlanKey(ctx, plan.PlanID)
	current, found := repository.plans[key]
	if !found {
		return ApplicationEvaluationPlan{}, false, nil
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationPlan(current), false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.LifecycleState}
	}
	if current.LifecycleState == applicationEvaluationPlanStateArchived {
		return cloneApplicationEvaluationPlan(current), false, errApplicationEvaluationArchived
	}
	if plan.LatestPlanVersion != current.LatestPlanVersion || plan.LatestPlanDigest != current.LatestPlanDigest ||
		plan.Name != current.Name || plan.ExecutionProfile != current.ExecutionProfile || plan.ItemCount != current.ItemCount ||
		plan.CreatedAt != current.CreatedAt || plan.CreatedByActorRef != current.CreatedByActorRef {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	repository.plans[key] = cloneApplicationEvaluationPlan(plan)
	return cloneApplicationEvaluationPlan(plan), true, nil
}

func (repository *memoryApplicationEvaluationRepository) ReadPlan(ctx ApplicationEvaluationContext, planID string) (ApplicationEvaluationPlan, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreUnavailable
	}
	plan, found := repository.plans[applicationEvaluationPlanKey(ctx, planID)]
	if found && validateApplicationEvaluationPlan(ctx, plan) != nil {
		return ApplicationEvaluationPlan{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationPlan(plan), found, nil
}

func (repository *memoryApplicationEvaluationRepository) ListPlans(ctx ApplicationEvaluationContext, filter ApplicationEvaluationPlanListFilter) (ApplicationEvaluationPlanListPage, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreUnavailable
	}
	prefix := applicationEvaluationScopeKey(ctx) + "\x00"
	values := make([]ApplicationEvaluationPlan, 0)
	for key, plan := range repository.plans {
		if !strings.HasPrefix(key, prefix) || plan.LifecycleState != filter.LifecycleState {
			continue
		}
		if filter.BeforeUpdatedAt != "" && (plan.UpdatedAt > filter.BeforeUpdatedAt || plan.UpdatedAt == filter.BeforeUpdatedAt && plan.PlanID >= filter.BeforePlanID) {
			continue
		}
		if validateApplicationEvaluationPlan(ctx, plan) != nil {
			return ApplicationEvaluationPlanListPage{}, errApplicationEvaluationStoreContract
		}
		values = append(values, cloneApplicationEvaluationPlan(plan))
	}
	sortApplicationEvaluationPlans(values)
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationPlanListPage{Plans: values, HasMore: hasMore}, nil
}

func (repository *memoryApplicationEvaluationRepository) ReadPlanVersion(ctx ApplicationEvaluationContext, planID string, version int) (ApplicationEvaluationPlanVersion, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreUnavailable
	}
	value, found := repository.versions[applicationEvaluationPlanKey(ctx, planID)][version]
	if found && validateApplicationEvaluationPlanVersion(ctx, value) != nil {
		return ApplicationEvaluationPlanVersion{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationPlanVersion(value), found, nil
}

func (repository *memoryApplicationEvaluationRepository) ListPlanVersions(ctx ApplicationEvaluationContext, planID string, filter ApplicationEvaluationVersionListFilter) (ApplicationEvaluationVersionListPage, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationPlanKey(ctx, planID)
	if _, found := repository.plans[key]; !found {
		return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationNotFound
	}
	values := make([]ApplicationEvaluationPlanVersion, 0, len(repository.versions[key]))
	for version, value := range repository.versions[key] {
		if filter.BeforeVersion != 0 && version >= filter.BeforeVersion {
			continue
		}
		if validateApplicationEvaluationPlanVersion(ctx, value) != nil {
			return ApplicationEvaluationVersionListPage{}, errApplicationEvaluationStoreContract
		}
		values = append(values, cloneApplicationEvaluationPlanVersion(value))
	}
	sort.Slice(values, func(i, j int) bool { return values[i].PlanVersion > values[j].PlanVersion })
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationVersionListPage{Versions: values, HasMore: hasMore}, nil
}

func (repository *memoryApplicationEvaluationRepository) CreateCampaign(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if validateApplicationEvaluationCampaign(ctx, campaign) != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationCampaignKey(ctx, campaign.CampaignID)
	clientKey := applicationEvaluationCampaignClientKey(ctx, campaign.ClientCampaignKey)
	if existingID, exists := repository.campaignKeys[clientKey]; exists {
		existing := repository.campaigns[applicationEvaluationCampaignKey(ctx, existingID)]
		if existing.PlanID != campaign.PlanID || existing.PlanVersion != campaign.PlanVersion || existing.PlanDigest != campaign.PlanDigest || existing.QuotaAPIKeyID != campaign.QuotaAPIKeyID {
			return cloneApplicationEvaluationCampaign(existing), false, errApplicationEvaluationCampaignConflict
		}
		return cloneApplicationEvaluationCampaign(existing), false, nil
	}
	if _, exists := repository.campaigns[key]; exists {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	repository.campaigns[key] = cloneApplicationEvaluationCampaign(campaign)
	repository.campaignKeys[clientKey] = campaign.CampaignID
	repository.campaignOrder = append(repository.campaignOrder, key)
	for len(repository.campaignOrder) > repository.campaignCapacity {
		evictedKey := repository.campaignOrder[0]
		evicted := repository.campaigns[evictedKey]
		delete(repository.campaignKeys, applicationEvaluationCampaignClientKeyFromRecord(evicted))
		delete(repository.campaigns, evictedKey)
		repository.campaignOrder = repository.campaignOrder[1:]
	}
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *memoryApplicationEvaluationRepository) UpdateCampaign(ctx ApplicationEvaluationContext, expected int, campaign ApplicationEvaluationCampaign) (ApplicationEvaluationCampaign, bool, error) {
	if expected < 1 || validateApplicationEvaluationCampaign(ctx, campaign) != nil || campaign.RecordVersion != expected+1 {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.unavailable {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	key := applicationEvaluationCampaignKey(ctx, campaign.CampaignID)
	current, found := repository.campaigns[key]
	if !found {
		return ApplicationEvaluationCampaign{}, false, nil
	}
	if current.RecordVersion != expected {
		return cloneApplicationEvaluationCampaign(current), false, applicationEvaluationVersionConflictError{CurrentVersion: current.RecordVersion, CurrentState: current.State}
	}
	if current.PlanID != campaign.PlanID || current.PlanVersion != campaign.PlanVersion || current.PlanDigest != campaign.PlanDigest ||
		current.ExecutionProfile != campaign.ExecutionProfile || current.QuotaAPIKeyID != campaign.QuotaAPIKeyID || current.ClientCampaignKey != campaign.ClientCampaignKey ||
		current.CreatedAt != campaign.CreatedAt || current.CreatedByActorRef != campaign.CreatedByActorRef || !validApplicationEvaluationCampaignUpdate(current, campaign) {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	repository.campaigns[key] = cloneApplicationEvaluationCampaign(campaign)
	return cloneApplicationEvaluationCampaign(campaign), true, nil
}

func (repository *memoryApplicationEvaluationRepository) ReadCampaign(ctx ApplicationEvaluationContext, campaignID string) (ApplicationEvaluationCampaign, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreUnavailable
	}
	campaign, found := repository.campaigns[applicationEvaluationCampaignKey(ctx, campaignID)]
	if found && validateApplicationEvaluationCampaign(ctx, campaign) != nil {
		return ApplicationEvaluationCampaign{}, false, errApplicationEvaluationStoreContract
	}
	return cloneApplicationEvaluationCampaign(campaign), found, nil
}

func (repository *memoryApplicationEvaluationRepository) ListCampaigns(ctx ApplicationEvaluationContext, filter ApplicationEvaluationCampaignListFilter) (ApplicationEvaluationCampaignListPage, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	if repository.unavailable {
		return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreUnavailable
	}
	prefix := applicationEvaluationScopeKey(ctx) + "\x00"
	values := make([]ApplicationEvaluationCampaign, 0)
	for key, campaign := range repository.campaigns {
		if !strings.HasPrefix(key, prefix) || filter.PlanID != "" && campaign.PlanID != filter.PlanID {
			continue
		}
		if filter.BeforeCreatedAt != "" && (campaign.CreatedAt > filter.BeforeCreatedAt || campaign.CreatedAt == filter.BeforeCreatedAt && campaign.CampaignID >= filter.BeforeCampaignID) {
			continue
		}
		if validateApplicationEvaluationCampaign(ctx, campaign) != nil {
			return ApplicationEvaluationCampaignListPage{}, errApplicationEvaluationStoreContract
		}
		values = append(values, cloneApplicationEvaluationCampaign(campaign))
	}
	sortApplicationEvaluationCampaigns(values)
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return ApplicationEvaluationCampaignListPage{Campaigns: values, HasMore: hasMore}, nil
}

func validApplicationEvaluationCampaignTransition(before, after string) bool {
	if before == after {
		return before == applicationEvaluationCampaignStateRunning
	}
	switch before {
	case applicationEvaluationCampaignStatePending:
		return after == applicationEvaluationCampaignStateRunning
	case applicationEvaluationCampaignStateRunning:
		return after == applicationEvaluationCampaignStateSucceeded || after == applicationEvaluationCampaignStateFailed || after == applicationEvaluationCampaignStateInterrupted
	default:
		return false
	}
}

func validApplicationEvaluationCampaignUpdate(current, next ApplicationEvaluationCampaign) bool {
	if current.State != next.State || current.State == applicationEvaluationCampaignStateRunning {
		return validApplicationEvaluationCampaignTransition(current.State, next.State)
	}
	if current.State != applicationEvaluationCampaignStateSucceeded && current.State != applicationEvaluationCampaignStateFailed && current.State != applicationEvaluationCampaignStateInterrupted {
		return false
	}
	if !validApplicationEvaluationHandoffTransition(current.Handoff, next.Handoff) {
		return false
	}
	currentCopy, nextCopy := cloneApplicationEvaluationCampaign(current), cloneApplicationEvaluationCampaign(next)
	if (currentCopy.Authority == nil) != (nextCopy.Authority == nil) {
		return false
	}
	if currentCopy.Authority != nil {
		if !applicationEvaluationJSONValuesEqual(currentCopy.Authority.Snapshot, nextCopy.Authority.Snapshot) {
			return false
		}
		currentCopy.Authority.Snapshot, nextCopy.Authority.Snapshot = nil, nil
	}
	currentCopy.RecordVersion, nextCopy.RecordVersion = 0, 0
	currentCopy.Handoff, nextCopy.Handoff = nil, nil
	currentCopy.UpdatedByActorRef, nextCopy.UpdatedByActorRef = "", ""
	currentCopy.RequestID, nextCopy.RequestID = "", ""
	currentCopy.AuditRef, nextCopy.AuditRef = "", ""
	return reflect.DeepEqual(currentCopy, nextCopy)
}

func applicationEvaluationJSONValuesEqual(current, next []byte) bool {
	var currentValue, nextValue any
	if decodeStrictApplicationEvaluationJSON(current, &currentValue) != nil || decodeStrictApplicationEvaluationJSON(next, &nextValue) != nil {
		return false
	}
	return reflect.DeepEqual(currentValue, nextValue)
}

func validApplicationEvaluationHandoffTransition(current, next *ApplicationEvaluationHandoffRef) bool {
	if next == nil || validateApplicationEvaluationHandoff(*next) != nil {
		return false
	}
	if current == nil {
		return true
	}
	if current.BaselineCampaignID != next.BaselineCampaignID || current.CandidateCampaignID != next.CandidateCampaignID ||
		current.AuditRef != next.AuditRef || current.State == "complete" || len(next.CaseRefs) < len(current.CaseRefs) {
		return false
	}
	for index := range current.CaseRefs {
		if current.CaseRefs[index] != next.CaseRefs[index] {
			return false
		}
	}
	return next.State == "partial" || next.State == "complete"
}

func applicationEvaluationScopeKey(ctx ApplicationEvaluationContext) string {
	return strings.Join([]string{ctx.TenantRef, ctx.WorkspaceID, ctx.Environment, ctx.ApplicationID}, "\x00")
}

func applicationEvaluationPlanKey(ctx ApplicationEvaluationContext, planID string) string {
	return applicationEvaluationScopeKey(ctx) + "\x00" + planID
}

func applicationEvaluationCampaignKey(ctx ApplicationEvaluationContext, campaignID string) string {
	return applicationEvaluationScopeKey(ctx) + "\x00" + campaignID
}

func applicationEvaluationCampaignClientKey(ctx ApplicationEvaluationContext, clientKey string) string {
	return applicationEvaluationScopeKey(ctx) + "\x00" + clientKey
}

func applicationEvaluationCampaignClientKeyFromRecord(campaign ApplicationEvaluationCampaign) string {
	return strings.Join([]string{campaign.TenantRef, campaign.WorkspaceID, campaign.Environment, campaign.ApplicationID, campaign.ClientCampaignKey}, "\x00")
}

func applicationEvaluationRepositoryFailure(err error) string {
	switch {
	case errors.Is(err, errApplicationEvaluationNotFound):
		return ApplicationEvaluationFailureNotFound
	case errors.Is(err, errApplicationEvaluationVersionConflict):
		return ApplicationEvaluationFailureVersionConflict
	case errors.Is(err, errApplicationEvaluationArchived):
		return ApplicationEvaluationFailureArchived
	case errors.Is(err, errApplicationEvaluationCampaignConflict):
		return ApplicationEvaluationFailureCampaignConflict
	case errors.Is(err, errApplicationEvaluationStoreContract):
		return ApplicationEvaluationFailureStoreContract
	default:
		return ApplicationEvaluationFailureStoreUnavailable
	}
}

var _ applicationEvaluationRepository = (*memoryApplicationEvaluationRepository)(nil)
