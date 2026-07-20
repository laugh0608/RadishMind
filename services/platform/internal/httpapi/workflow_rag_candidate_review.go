package httpapi

import (
	"reflect"
	"sort"
	"strings"
	"time"
)

func buildWorkflowRAGCandidateReview(
	ctx WorkflowRAGSnapshotContext,
	reviewID string,
	createdAt string,
	baseline WorkflowRAGQualityReview,
	candidate WorkflowRAGQualityReview,
	baselineLifecycle string,
	candidateLifecycle string,
) WorkflowRAGCandidateSnapshotReview {
	samples, regressed, improved := compareWorkflowRAGQualitySamples(baseline.Samples, candidate.Samples)
	addedFindings, removedFindings := compareWorkflowRAGFindingCodes(baseline.Findings, candidate.Findings)
	delta := workflowRAGQualityMetricDelta(baseline.Metrics, candidate.Metrics)
	if workflowRAGMetricDeltaHasRegression(delta) || workflowRAGHasReviewRequiredFinding(candidate.Findings, addedFindings) {
		regressed = true
	}
	if workflowRAGMetricDeltaHasImprovement(delta) || workflowRAGHasRemovedReviewRequiredFinding(baseline.Findings, removedFindings) {
		improved = true
	}
	conclusion := "unchanged"
	switch {
	case regressed:
		conclusion = "regressed"
	case candidate.Status == "needs_review":
		conclusion = "needs_review"
	case improved:
		conclusion = "improved"
	}
	return WorkflowRAGCandidateSnapshotReview{
		SchemaVersion: workflowRAGCandidateReviewSchemaVersion,
		ReviewID:      reviewID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID,
		Dataset: baseline.Dataset, Profile: baseline.Profile, BaselineSnapshot: baseline.Snapshot, CandidateSnapshot: candidate.Snapshot,
		BaselineLifecycle: baselineLifecycle, CandidateLifecycle: candidateLifecycle, Conclusion: conclusion,
		Baseline: baseline, Candidate: candidate, MetricDelta: delta, Samples: samples,
		AddedFindingCodes: addedFindings, RemovedFindingCodes: removedFindings,
		CreatedAt: createdAt, CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
}

func compareWorkflowRAGQualitySamples(baseline, candidate []WorkflowRAGQualitySampleResult) ([]WorkflowRAGCandidateSampleComparison, bool, bool) {
	candidates := make(map[string]WorkflowRAGQualitySampleResult, len(candidate))
	for _, sample := range candidate {
		candidates[sample.SampleID] = sample
	}
	comparisons := make([]WorkflowRAGCandidateSampleComparison, 0, len(baseline))
	regressed, improved := false, false
	for _, before := range baseline {
		after := candidates[before.SampleID]
		change := "unchanged"
		if before.Passed && !after.Passed || after.ExpectedHitCount < before.ExpectedHitCount || after.RequiredOfficialHitCount < before.RequiredOfficialHitCount || workflowRAGRankRegressed(before.FirstExpectedRank, after.FirstExpectedRank) {
			change, regressed = "regressed", true
		} else if !before.Passed && after.Passed || after.ExpectedHitCount > before.ExpectedHitCount || after.RequiredOfficialHitCount > before.RequiredOfficialHitCount || workflowRAGRankImproved(before.FirstExpectedRank, after.FirstExpectedRank) {
			change, improved = "improved", true
		}
		comparisons = append(comparisons, WorkflowRAGCandidateSampleComparison{
			SampleID: before.SampleID, BaselinePassed: before.Passed, CandidatePassed: after.Passed,
			BaselineExpectedHitCount: before.ExpectedHitCount, CandidateExpectedHitCount: after.ExpectedHitCount,
			BaselineRequiredOfficialHitCount: before.RequiredOfficialHitCount, CandidateRequiredOfficialHitCount: after.RequiredOfficialHitCount,
			BaselineFirstExpectedRank: before.FirstExpectedRank, CandidateFirstExpectedRank: after.FirstExpectedRank, Change: change,
		})
	}
	return comparisons, regressed, improved
}

func workflowRAGRankRegressed(before, after int) bool {
	return before > 0 && (after == 0 || after > before)
}

func workflowRAGRankImproved(before, after int) bool {
	return after > 0 && (before == 0 || after < before)
}

func workflowRAGQualityMetricDelta(baseline, candidate WorkflowRAGQualityMetrics) WorkflowRAGMetricDelta {
	return WorkflowRAGMetricDelta{
		HitAtK:                    workflowRAGQualityRound(candidate.HitAtK - baseline.HitAtK),
		ExpectedRecallAtK:         workflowRAGQualityRound(candidate.ExpectedRecallAtK - baseline.ExpectedRecallAtK),
		RequiredOfficialRecallAtK: workflowRAGQualityRound(candidate.RequiredOfficialRecallAtK - baseline.RequiredOfficialRecallAtK),
		MeanReciprocalRank:        workflowRAGQualityRound(candidate.MeanReciprocalRank - baseline.MeanReciprocalRank),
		NoEvidenceAccuracy:        workflowRAGQualityRound(candidate.NoEvidenceAccuracy - baseline.NoEvidenceAccuracy),
		SamplePassRate:            workflowRAGQualityRound(candidate.SamplePassRate - baseline.SamplePassRate),
	}
}

func workflowRAGMetricDeltaValues(delta WorkflowRAGMetricDelta) []float64 {
	return []float64{delta.HitAtK, delta.ExpectedRecallAtK, delta.RequiredOfficialRecallAtK, delta.MeanReciprocalRank, delta.NoEvidenceAccuracy, delta.SamplePassRate}
}

func workflowRAGMetricDeltaHasRegression(delta WorkflowRAGMetricDelta) bool {
	for _, value := range workflowRAGMetricDeltaValues(delta) {
		if value < 0 {
			return true
		}
	}
	return false
}

func workflowRAGMetricDeltaHasImprovement(delta WorkflowRAGMetricDelta) bool {
	for _, value := range workflowRAGMetricDeltaValues(delta) {
		if value > 0 {
			return true
		}
	}
	return false
}

func compareWorkflowRAGFindingCodes(baseline, candidate []WorkflowRAGQualityFinding) ([]string, []string) {
	before, after := workflowRAGFindingCodeSet(baseline), workflowRAGFindingCodeSet(candidate)
	added, removed := make([]string, 0), make([]string, 0)
	for code := range after {
		if !before[code] {
			added = append(added, code)
		}
	}
	for code := range before {
		if !after[code] {
			removed = append(removed, code)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func workflowRAGFindingCodeSet(findings []WorkflowRAGQualityFinding) map[string]bool {
	result := make(map[string]bool, len(findings))
	for _, finding := range findings {
		result[finding.Code] = true
	}
	return result
}

func workflowRAGHasReviewRequiredFinding(findings []WorkflowRAGQualityFinding, selected []string) bool {
	set := make(map[string]bool, len(selected))
	for _, code := range selected {
		set[code] = true
	}
	for _, finding := range findings {
		if set[finding.Code] && finding.Severity == "review_required" {
			return true
		}
	}
	return false
}

func workflowRAGHasRemovedReviewRequiredFinding(findings []WorkflowRAGQualityFinding, selected []string) bool {
	return workflowRAGHasReviewRequiredFinding(findings, selected)
}

func validateStoredWorkflowRAGEvaluationResource(resource WorkflowRAGEvaluationDatasetResource, ctx WorkflowRAGSnapshotContext) error {
	if resource.SchemaVersion != workflowRAGEvaluationResourceSchemaVersion || !workflowRAGDatasetIDPattern.MatchString(resource.DatasetID) ||
		resource.TenantRef != ctx.TenantRef || resource.WorkspaceID != ctx.WorkspaceID || resource.ApplicationID != ctx.ApplicationID ||
		!workflowRAGSnapshotKeyPattern.MatchString(resource.DatasetKey) || strings.TrimSpace(resource.DisplayName) != resource.DisplayName ||
		resource.DisplayName == "" || !workflowRAGEvaluationClassificationAllowed(resource.ContentClassification, false) ||
		(resource.LifecycleState != workflowRAGEvaluationActive && resource.LifecycleState != workflowRAGEvaluationArchived) ||
		resource.LatestVersion < 1 || !workflowRAGDigestPattern.MatchString(resource.LatestDigest) || resource.SampleCount < 1 || resource.SampleCount > 200 ||
		!validateWorkflowRAGEvaluationSnapshotBinding(resource.BaselineSnapshot) {
		return errWorkflowRAGEvaluationContract
	}
	created, err := time.Parse(time.RFC3339Nano, resource.CreatedAt)
	if err != nil {
		return errWorkflowRAGEvaluationContract
	}
	updated, err := time.Parse(time.RFC3339Nano, resource.UpdatedAt)
	if err != nil || updated.Before(created) {
		return errWorkflowRAGEvaluationContract
	}
	if resource.LifecycleState == workflowRAGEvaluationActive && resource.ArchivedAt != nil {
		return errWorkflowRAGEvaluationContract
	}
	if resource.LifecycleState == workflowRAGEvaluationArchived {
		if resource.ArchivedAt == nil {
			return errWorkflowRAGEvaluationContract
		}
		archived, parseErr := time.Parse(time.RFC3339Nano, *resource.ArchivedAt)
		if parseErr != nil || archived.Before(created) {
			return errWorkflowRAGEvaluationContract
		}
	}
	return nil
}

func validateStoredWorkflowRAGEvaluationVersion(version WorkflowRAGEvaluationDatasetVersion, ctx WorkflowRAGSnapshotContext) error {
	if version.SchemaVersion != workflowRAGEvaluationResourceSchemaVersion || !workflowRAGSnapshotKeyPattern.MatchString(version.DatasetKey) ||
		strings.TrimSpace(version.DisplayName) != version.DisplayName || version.DisplayName == "" ||
		(version.LifecycleState != workflowRAGEvaluationActive && version.LifecycleState != workflowRAGEvaluationArchived) ||
		version.CreatedByActorRef == "" || version.RequestID == "" || version.AuditRef == "" ||
		version.Dataset.Snapshot.TenantRef != ctx.TenantRef || version.Dataset.Snapshot.WorkspaceID != ctx.WorkspaceID || version.Dataset.Snapshot.ApplicationID != ctx.ApplicationID ||
		validateWorkflowRAGApplicationEvaluationDataset(version.Dataset, nil) != nil {
		return errWorkflowRAGEvaluationContract
	}
	if _, err := time.Parse(time.RFC3339Nano, version.CreatedAt); err != nil {
		return errWorkflowRAGEvaluationContract
	}
	return nil
}

func validateStoredWorkflowRAGEvaluationAudit(audit WorkflowRAGEvaluationAudit, ctx WorkflowRAGSnapshotContext) error {
	if audit.SchemaVersion != workflowRAGEvaluationAuditSchemaVersion || !workflowRAGEvaluationAuditIDPattern.MatchString(audit.EventID) ||
		!workflowRAGEvaluationAuditKindAllowed(audit.EventKind) || audit.TenantRef != ctx.TenantRef || audit.WorkspaceID != ctx.WorkspaceID || audit.ApplicationID != ctx.ApplicationID ||
		!workflowRAGDatasetIDPattern.MatchString(audit.DatasetID) || audit.DatasetVersion < 1 || !workflowRAGDigestPattern.MatchString(audit.DatasetDigest) ||
		!workflowRAGEvaluationClassificationAllowed(audit.ContentClassification, false) || !validateWorkflowRAGEvaluationSnapshotBinding(audit.BaselineSnapshot) ||
		audit.SampleCount < 1 || audit.SampleCount > 200 || audit.ActorRef != ctx.ActorRef || audit.RequestID != ctx.RequestID || audit.AuditRef != ctx.AuditRef {
		return errWorkflowRAGEvaluationContract
	}
	if _, err := time.Parse(time.RFC3339Nano, audit.OccurredAt); err != nil {
		return errWorkflowRAGEvaluationContract
	}
	return nil
}

func workflowRAGEvaluationAuditKindAllowed(kind string) bool {
	switch kind {
	case "dataset_created", "dataset_versioned", "dataset_archived", "candidate_review_created":
		return true
	default:
		return false
	}
}

func validateStoredWorkflowRAGCandidateReview(review WorkflowRAGCandidateSnapshotReview, ctx WorkflowRAGSnapshotContext) error {
	if review.SchemaVersion != workflowRAGCandidateReviewSchemaVersion || !workflowRAGEvaluationReviewIDPattern.MatchString(review.ReviewID) ||
		review.TenantRef != ctx.TenantRef || review.WorkspaceID != ctx.WorkspaceID || review.ApplicationID != ctx.ApplicationID ||
		review.BaselineLifecycle != workflowRAGEvaluationActive && review.BaselineLifecycle != workflowRAGEvaluationArchived ||
		review.CandidateLifecycle != workflowRAGEvaluationActive && review.CandidateLifecycle != workflowRAGEvaluationArchived ||
		(review.Conclusion != "improved" && review.Conclusion != "unchanged" && review.Conclusion != "regressed" && review.Conclusion != "needs_review") ||
		review.CreatedByActorRef != ctx.ActorRef || review.RequestID != ctx.RequestID || review.AuditRef != ctx.AuditRef ||
		validateWorkflowRAGQualityReview(review.Baseline) != nil || validateWorkflowRAGQualityReview(review.Candidate) != nil ||
		review.Dataset != review.Baseline.Dataset || review.Dataset != review.Candidate.Dataset || review.Profile != review.Baseline.Profile || review.Profile != review.Candidate.Profile ||
		review.BaselineSnapshot != review.Baseline.Snapshot || review.CandidateSnapshot != review.Candidate.Snapshot || review.Samples == nil || review.AddedFindingCodes == nil || review.RemovedFindingCodes == nil {
		return errWorkflowRAGEvaluationContract
	}
	if _, err := time.Parse(time.RFC3339Nano, review.CreatedAt); err != nil {
		return errWorkflowRAGEvaluationContract
	}
	rebuilt := buildWorkflowRAGCandidateReview(ctx, review.ReviewID, review.CreatedAt, review.Baseline, review.Candidate, review.BaselineLifecycle, review.CandidateLifecycle)
	if review.Conclusion != rebuilt.Conclusion || review.MetricDelta != rebuilt.MetricDelta || !reflect.DeepEqual(review.Samples, rebuilt.Samples) ||
		!reflect.DeepEqual(review.AddedFindingCodes, rebuilt.AddedFindingCodes) || !reflect.DeepEqual(review.RemovedFindingCodes, rebuilt.RemovedFindingCodes) {
		return errWorkflowRAGEvaluationContract
	}
	return nil
}
