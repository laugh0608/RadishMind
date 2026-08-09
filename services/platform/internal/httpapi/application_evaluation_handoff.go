package httpapi

import (
	"errors"
	"reflect"
	"strings"
)

type ApplicationEvaluationPairInput struct {
	BaselineCampaignID  string
	CandidateCampaignID string
}

type ApplicationEvaluationHandoffInput struct {
	BaselineCampaignID               string
	CandidateCampaignID              string
	ExpectedBaselineRecordVersion    int
	ExpectedCandidateRecordVersion   int
	AcknowledgeEvidenceMaterializing bool
}

type ApplicationEvaluationPairItem struct {
	ItemKey                string                              `json:"item_key"`
	Name                   string                              `json:"name"`
	BaselineRunID          string                              `json:"baseline_run_id"`
	CandidateRunID         string                              `json:"candidate_run_id"`
	ExpectedClassification WorkflowRunComparisonClassification `json:"expected_classification"`
	ActualClassification   WorkflowRunComparisonClassification `json:"actual_classification"`
	ExpectationMatched     bool                                `json:"expectation_matched"`
	Comparison             *WorkflowRunComparison              `json:"comparison"`
}

type ApplicationEvaluationPairReview struct {
	PlanID              string                           `json:"plan_id"`
	PlanName            string                           `json:"plan_name"`
	PlanVersion         int                              `json:"plan_version"`
	PlanDigest          string                           `json:"plan_digest"`
	ExecutionProfile    string                           `json:"execution_profile"`
	BaselineCampaignID  string                           `json:"baseline_campaign_id"`
	CandidateCampaignID string                           `json:"candidate_campaign_id"`
	ExpectedMatches     int                              `json:"expected_matches"`
	ExpectedMismatches  int                              `json:"expected_mismatches"`
	Items               []ApplicationEvaluationPairItem  `json:"items"`
	ExistingHandoff     *ApplicationEvaluationHandoffRef `json:"existing_handoff"`
}

type ApplicationEvaluationPairResult struct {
	Review                  *ApplicationEvaluationPairReview
	CandidateCampaign       *ApplicationEvaluationCampaign
	Handoff                 *ApplicationEvaluationHandoffRef
	IdempotentReplay        bool
	FailureCode             string
	FailureSummary          string
	CurrentBaselineVersion  int
	CurrentCandidateVersion int
}

type applicationEvaluationHandoffService struct {
	repository applicationEvaluationRepository
	comparison workflowExecutorService
	evaluation workflowEvaluationService
	suite      workflowEvaluationSuiteService
}

func newApplicationEvaluationHandoffService(
	repository applicationEvaluationRepository,
	runStore workflowRunStore,
	evaluation workflowEvaluationService,
	suite workflowEvaluationSuiteService,
) applicationEvaluationHandoffService {
	return applicationEvaluationHandoffService{
		repository: repository, comparison: newWorkflowExecutorService(nil, nil, runStore), evaluation: evaluation, suite: suite,
	}
}

func (service applicationEvaluationHandoffService) Preview(ctx ApplicationEvaluationContext, input ApplicationEvaluationPairInput) ApplicationEvaluationPairResult {
	return service.buildReview(ctx, input)
}

func (service applicationEvaluationHandoffService) Materialize(ctx ApplicationEvaluationContext, input ApplicationEvaluationHandoffInput) ApplicationEvaluationPairResult {
	if failure := validateApplicationEvaluationMutationContext(ctx); failure != "" {
		return applicationEvaluationPairFailure(failure)
	}
	if input.ExpectedBaselineRecordVersion < 1 || input.ExpectedCandidateRecordVersion < 1 || !input.AcknowledgeEvidenceMaterializing {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	pair := service.buildReview(ctx, ApplicationEvaluationPairInput{
		BaselineCampaignID: input.BaselineCampaignID, CandidateCampaignID: input.CandidateCampaignID,
	})
	if pair.FailureCode != "" || pair.Review == nil || pair.CandidateCampaign == nil {
		return pair
	}
	baseline, found, err := service.repository.ReadCampaign(ctx, strings.TrimSpace(input.BaselineCampaignID))
	if err != nil {
		return applicationEvaluationPairFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailureNotFound)
	}
	candidate := *pair.CandidateCampaign
	if candidate.Handoff != nil {
		pair.Handoff = cloneApplicationEvaluationHandoff(candidate.Handoff)
		pair.IdempotentReplay = true
		if candidate.Handoff.State == "partial" {
			pair.FailureCode = ApplicationEvaluationFailureHandoffPartial
			pair.FailureSummary = applicationEvaluationFailureSummary(pair.FailureCode)
		}
		return pair
	}
	if baseline.RecordVersion != input.ExpectedBaselineRecordVersion || candidate.RecordVersion != input.ExpectedCandidateRecordVersion {
		pair.FailureCode = ApplicationEvaluationFailureVersionConflict
		pair.FailureSummary = applicationEvaluationFailureSummary(pair.FailureCode)
		pair.CurrentBaselineVersion, pair.CurrentCandidateVersion = baseline.RecordVersion, candidate.RecordVersion
		return pair
	}
	runContext := applicationEvaluationWorkflowRunContext(ctx)
	handoff := ApplicationEvaluationHandoffRef{
		BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID,
		CaseRefs: []WorkflowEvaluationSuiteCaseRef{}, State: "partial", AuditRef: ctx.AuditRef + "_handoff",
	}
	for _, item := range pair.Review.Items {
		created := service.evaluation.Create(runContext, WorkflowEvaluationCreateRequest{
			Name: applicationEvaluationEvidenceName(item.Name, item.ItemKey), BaselineRunID: item.BaselineRunID,
			Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: item.CandidateRunID, ExpectedClassification: item.ExpectedClassification}},
		})
		if created.FailureCode != "" || created.Case == nil {
			result := service.persistPartialHandoff(ctx, candidate, handoff, applicationEvaluationWorkflowEvaluationFailure(created.FailureCode))
			result.Review = pair.Review
			return result
		}
		handoff.CaseRefs = append(handoff.CaseRefs, WorkflowEvaluationSuiteCaseRef{CaseID: created.Case.CaseID, Version: created.Case.Version})
		stored, result := service.checkpointHandoff(ctx, candidate, handoff)
		if result.FailureCode != "" {
			result.Review = pair.Review
			result.Handoff = cloneApplicationEvaluationHandoff(&handoff)
			return result
		}
		candidate = stored
	}

	createdSuite := service.suite.Create(runContext, WorkflowEvaluationSuiteCreateRequest{
		Name: applicationEvaluationEvidenceName(pair.Review.PlanName, "campaign-pair"), CaseRefs: handoff.CaseRefs,
	})
	if createdSuite.FailureCode != "" || createdSuite.Suite == nil {
		result := service.persistPartialHandoff(ctx, candidate, handoff, applicationEvaluationWorkflowSuiteFailure(createdSuite.FailureCode))
		result.Review = pair.Review
		return result
	}
	handoff.State, handoff.SuiteID = "complete", createdSuite.Suite.SuiteID
	stored, result := service.checkpointHandoff(ctx, candidate, handoff)
	if result.FailureCode != "" {
		result.Review = pair.Review
		result.Handoff = cloneApplicationEvaluationHandoff(&handoff)
		return result
	}
	result.Review, result.CandidateCampaign, result.Handoff = pair.Review, &stored, cloneApplicationEvaluationHandoff(stored.Handoff)
	return result
}

func (service applicationEvaluationHandoffService) buildReview(ctx ApplicationEvaluationContext, input ApplicationEvaluationPairInput) ApplicationEvaluationPairResult {
	if !validApplicationEvaluationContext(ctx) {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailureScopeDenied)
	}
	input.BaselineCampaignID, input.CandidateCampaignID = strings.TrimSpace(input.BaselineCampaignID), strings.TrimSpace(input.CandidateCampaignID)
	if !applicationEvaluationCampaignIDPattern.MatchString(input.BaselineCampaignID) || !applicationEvaluationCampaignIDPattern.MatchString(input.CandidateCampaignID) || input.BaselineCampaignID == input.CandidateCampaignID {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailurePayloadInvalid)
	}
	baseline, found, err := service.repository.ReadCampaign(ctx, input.BaselineCampaignID)
	if err != nil {
		return applicationEvaluationPairFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailureNotFound)
	}
	candidate, found, err := service.repository.ReadCampaign(ctx, input.CandidateCampaignID)
	if err != nil {
		return applicationEvaluationPairFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailureNotFound)
	}
	result := ApplicationEvaluationPairResult{CurrentBaselineVersion: baseline.RecordVersion, CurrentCandidateVersion: candidate.RecordVersion}
	if !applicationEvaluationCampaignsPairable(baseline, candidate) {
		result.FailureCode, result.FailureSummary = ApplicationEvaluationFailureRunUnavailable, applicationEvaluationFailureSummary(ApplicationEvaluationFailureRunUnavailable)
		return result
	}
	version, found, err := service.repository.ReadPlanVersion(ctx, baseline.PlanID, baseline.PlanVersion)
	if err != nil {
		return applicationEvaluationPairFailure(applicationEvaluationRepositoryFailure(err))
	}
	if !found || version.PlanDigest != baseline.PlanDigest || len(version.Items) != len(baseline.Items) || len(version.Items) != len(candidate.Items) {
		return applicationEvaluationPairFailure(ApplicationEvaluationFailureStoreContract)
	}
	review := ApplicationEvaluationPairReview{
		PlanID: version.PlanID, PlanName: version.Name, PlanVersion: version.PlanVersion, PlanDigest: version.PlanDigest, ExecutionProfile: version.ExecutionProfile,
		BaselineCampaignID: baseline.CampaignID, CandidateCampaignID: candidate.CampaignID,
		Items: make([]ApplicationEvaluationPairItem, 0, len(version.Items)), ExistingHandoff: cloneApplicationEvaluationHandoff(candidate.Handoff),
	}
	runContext := applicationEvaluationWorkflowRunContext(ctx)
	for index, planItem := range version.Items {
		baselineItem, candidateItem := baseline.Items[index], candidate.Items[index]
		if baselineItem.ItemKey != planItem.ItemKey || candidateItem.ItemKey != planItem.ItemKey ||
			baselineItem.State != applicationEvaluationCampaignItemSucceeded || candidateItem.State != applicationEvaluationCampaignItemSucceeded {
			return applicationEvaluationPairFailure(ApplicationEvaluationFailureRunUnavailable)
		}
		comparison := service.comparison.CompareRuns(runContext, baselineItem.RunID, candidateItem.RunID)
		if comparison.FailureCode != "" || comparison.Comparison == nil {
			return applicationEvaluationPairFailure(applicationEvaluationComparisonFailure(comparison.FailureCode))
		}
		if !applicationEvaluationComparisonMatchesProfile(*comparison.Comparison, version.ExecutionProfile) {
			return applicationEvaluationPairFailure(ApplicationEvaluationFailureProfileIneligible)
		}
		item := ApplicationEvaluationPairItem{
			ItemKey: planItem.ItemKey, Name: planItem.Name, BaselineRunID: baselineItem.RunID, CandidateRunID: candidateItem.RunID,
			ExpectedClassification: planItem.ExpectedClassification, ActualClassification: comparison.Comparison.Classification,
			ExpectationMatched: comparison.Comparison.Classification == planItem.ExpectedClassification, Comparison: comparison.Comparison,
		}
		if item.ExpectationMatched {
			review.ExpectedMatches++
		} else {
			review.ExpectedMismatches++
		}
		review.Items = append(review.Items, item)
	}
	result.Review, result.CandidateCampaign = &review, &candidate
	return result
}

func (service applicationEvaluationHandoffService) checkpointHandoff(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign, handoff ApplicationEvaluationHandoffRef) (ApplicationEvaluationCampaign, ApplicationEvaluationPairResult) {
	expected := campaign.RecordVersion
	campaign.RecordVersion++
	campaign.Handoff = cloneApplicationEvaluationHandoff(&handoff)
	campaign.UpdatedByActorRef, campaign.RequestID, campaign.AuditRef = ctx.ActorRef, ctx.RequestID, ctx.AuditRef
	stored, updated, err := service.repository.UpdateCampaign(ctx, expected, campaign)
	if err != nil || !updated {
		result := applicationEvaluationPairFailure(applicationEvaluationRepositoryFailure(err))
		if conflict := (applicationEvaluationVersionConflictError{}); errors.As(err, &conflict) {
			result.CurrentCandidateVersion = conflict.CurrentVersion
		}
		return ApplicationEvaluationCampaign{}, result
	}
	return stored, ApplicationEvaluationPairResult{}
}

func (service applicationEvaluationHandoffService) persistPartialHandoff(ctx ApplicationEvaluationContext, campaign ApplicationEvaluationCampaign, handoff ApplicationEvaluationHandoffRef, cause string) ApplicationEvaluationPairResult {
	handoff.State, handoff.SuiteID = "partial", ""
	if campaign.Handoff != nil && reflect.DeepEqual(campaign.Handoff.CaseRefs, handoff.CaseRefs) {
		result := ApplicationEvaluationPairResult{CandidateCampaign: &campaign, Handoff: cloneApplicationEvaluationHandoff(campaign.Handoff)}
		result.FailureCode = ApplicationEvaluationFailureHandoffPartial
		result.FailureSummary = applicationEvaluationFailureSummary(result.FailureCode)
		if strings.TrimSpace(cause) != "" {
			result.FailureSummary += " Cause: " + strings.TrimSpace(cause) + "."
		}
		return result
	}
	stored, result := service.checkpointHandoff(ctx, campaign, handoff)
	if result.FailureCode != "" {
		result.Handoff = cloneApplicationEvaluationHandoff(&handoff)
		return result
	}
	result.CandidateCampaign, result.Handoff = &stored, cloneApplicationEvaluationHandoff(stored.Handoff)
	result.FailureCode = ApplicationEvaluationFailureHandoffPartial
	result.FailureSummary = applicationEvaluationFailureSummary(result.FailureCode)
	if strings.TrimSpace(cause) != "" {
		result.FailureSummary += " Cause: " + strings.TrimSpace(cause) + "."
	}
	return result
}

func applicationEvaluationCampaignsPairable(baseline, candidate ApplicationEvaluationCampaign) bool {
	return baseline.State == applicationEvaluationCampaignStateSucceeded && candidate.State == applicationEvaluationCampaignStateSucceeded &&
		baseline.PlanID == candidate.PlanID && baseline.PlanVersion == candidate.PlanVersion && baseline.PlanDigest == candidate.PlanDigest &&
		baseline.ExecutionProfile == candidate.ExecutionProfile && baseline.TenantRef == candidate.TenantRef && baseline.WorkspaceID == candidate.WorkspaceID &&
		baseline.Environment == candidate.Environment && baseline.ApplicationID == candidate.ApplicationID
}

func applicationEvaluationEvidenceName(preferred, fallback string) string {
	value := strings.TrimSpace(preferred)
	runes := []rune(value)
	if len(runes) > 96 {
		value = string(runes[:96])
	}
	if validWorkflowEvaluationName(value) {
		return value
	}
	return "Application evaluation " + strings.TrimSpace(fallback)
}

func applicationEvaluationComparisonFailure(code WorkflowRunFailureCode) string {
	switch code {
	case WorkflowRunFailureRecordNotFound, WorkflowRunFailureComparisonInvalid:
		return ApplicationEvaluationFailureRunUnavailable
	case WorkflowRunFailureDefinitionIncompatible, WorkflowRunFailureRetrievalIncompatible,
		WorkflowRunFailurePromptIncompatible, WorkflowRunFailureAgentCopilotIncompatible,
		WorkflowRunFailureSideEffectUnsupported:
		return ApplicationEvaluationFailureProfileIneligible
	case WorkflowRunFailureStoreContractMismatch:
		return ApplicationEvaluationFailureStoreContract
	default:
		return ApplicationEvaluationFailureStoreUnavailable
	}
}

func applicationEvaluationComparisonMatchesProfile(comparison WorkflowRunComparison, profile string) bool {
	switch profile {
	case applicationInteractionProfileWorkflow:
		return comparison.SchemaVersion == workflowDefinitionRunComparisonSchemaVersion
	case applicationInteractionProfileRAG:
		return comparison.SchemaVersion == workflowRAGAppRunComparisonSchemaVersion
	case applicationInteractionProfilePrompt:
		return comparison.SchemaVersion == promptApplicationRunComparisonSchemaVersion
	case applicationInteractionProfileAgentCopilot:
		return comparison.SchemaVersion == agentCopilotRunComparisonSchemaVersion
	default:
		return false
	}
}

func applicationEvaluationWorkflowEvaluationFailure(code WorkflowEvaluationFailureCode) string {
	if code == "" {
		return ApplicationEvaluationFailureStoreUnavailable
	}
	return string(code)
}

func applicationEvaluationWorkflowSuiteFailure(code WorkflowEvaluationSuiteFailureCode) string {
	if code == "" {
		return ApplicationEvaluationFailureStoreUnavailable
	}
	return string(code)
}

func applicationEvaluationPairFailure(code string) ApplicationEvaluationPairResult {
	return ApplicationEvaluationPairResult{FailureCode: code, FailureSummary: applicationEvaluationFailureSummary(code)}
}

func cloneApplicationEvaluationHandoff(value *ApplicationEvaluationHandoffRef) *ApplicationEvaluationHandoffRef {
	if value == nil {
		return nil
	}
	copy := *value
	copy.CaseRefs = append([]WorkflowEvaluationSuiteCaseRef(nil), value.CaseRefs...)
	return &copy
}
