package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	workflowRAGEvaluationDatasetSchemaVersion = "workflow_rag_evaluation_dataset.v1"
	workflowRAGQualityReviewSchemaVersion     = "workflow_rag_quality_review.v1"
	workflowRAGEvaluationClassification       = "synthetic_public"
)

var errWorkflowRAGQualityReviewContract = errors.New("workflow RAG quality review contract invalid")

type WorkflowRAGEvaluationSnapshotBinding struct {
	TenantRef       string `json:"tenant_ref"`
	WorkspaceID     string `json:"workspace_id"`
	ApplicationID   string `json:"application_id"`
	SnapshotID      string `json:"snapshot_id"`
	SnapshotVersion int    `json:"snapshot_version"`
	SnapshotDigest  string `json:"snapshot_digest"`
	RAGRef          string `json:"rag_ref"`
}

type WorkflowRAGEvaluationProfileBinding struct {
	ProfileID      string `json:"profile_id"`
	ProfileVersion int    `json:"profile_version"`
	ProfileDigest  string `json:"profile_digest"`
}

type WorkflowRAGEvaluationThresholds struct {
	HitAtK                    float64 `json:"hit_at_k"`
	ExpectedRecallAtK         float64 `json:"expected_recall_at_k"`
	RequiredOfficialRecallAtK float64 `json:"required_official_recall_at_k"`
	MeanReciprocalRank        float64 `json:"mean_reciprocal_rank"`
	NoEvidenceAccuracy        float64 `json:"no_evidence_accuracy"`
	SamplePassRate            float64 `json:"sample_pass_rate"`
}

type WorkflowRAGEvaluationReviewMetadata struct {
	ReviewerRef   string `json:"reviewer_ref"`
	ReviewedAt    string `json:"reviewed_at"`
	ReviewSummary string `json:"review_summary"`
}

type WorkflowRAGEvaluationSample struct {
	SampleID             string   `json:"sample_id"`
	QueryText            string   `json:"query_text"`
	Expectation          string   `json:"expectation"`
	ExpectedCitationRefs []string `json:"expected_citation_refs"`
	RequiredOfficialRefs []string `json:"required_official_refs"`
	TopK                 int      `json:"top_k"`
	ReviewNote           string   `json:"review_note"`
}

type WorkflowRAGEvaluationDataset struct {
	SchemaVersion         string                               `json:"schema_version"`
	DatasetID             string                               `json:"dataset_id"`
	DatasetVersion        int                                  `json:"dataset_version"`
	DatasetDigest         string                               `json:"dataset_digest"`
	ContentClassification string                               `json:"content_classification"`
	Snapshot              WorkflowRAGEvaluationSnapshotBinding `json:"snapshot"`
	Profile               WorkflowRAGEvaluationProfileBinding  `json:"profile"`
	Thresholds            WorkflowRAGEvaluationThresholds      `json:"thresholds"`
	Review                WorkflowRAGEvaluationReviewMetadata  `json:"review"`
	Samples               []WorkflowRAGEvaluationSample        `json:"samples"`
}

type WorkflowRAGQualitySelectedFragment struct {
	FragmentRef   string `json:"fragment_ref"`
	Rank          int    `json:"rank"`
	ContentDigest string `json:"content_digest"`
	SourceType    string `json:"source_type"`
	IsOfficial    bool   `json:"is_official"`
}

type WorkflowRAGQualitySampleResult struct {
	SampleID                 string                               `json:"sample_id"`
	Expectation              string                               `json:"expectation"`
	QueryDigest              string                               `json:"query_digest"`
	QueryBytes               int                                  `json:"query_bytes"`
	TopK                     int                                  `json:"top_k"`
	FailureCode              string                               `json:"failure_code"`
	Selected                 []WorkflowRAGQualitySelectedFragment `json:"selected"`
	ExpectedRefCount         int                                  `json:"expected_ref_count"`
	ExpectedHitCount         int                                  `json:"expected_hit_count"`
	RequiredOfficialRefCount int                                  `json:"required_official_ref_count"`
	RequiredOfficialHitCount int                                  `json:"required_official_hit_count"`
	FirstExpectedRank        int                                  `json:"first_expected_rank"`
	Passed                   bool                                 `json:"passed"`
}

type WorkflowRAGQualityMetrics struct {
	HitAtK                    float64 `json:"hit_at_k"`
	ExpectedRecallAtK         float64 `json:"expected_recall_at_k"`
	RequiredOfficialRecallAtK float64 `json:"required_official_recall_at_k"`
	MeanReciprocalRank        float64 `json:"mean_reciprocal_rank"`
	NoEvidenceAccuracy        float64 `json:"no_evidence_accuracy"`
	SamplePassRate            float64 `json:"sample_pass_rate"`
}

type WorkflowRAGQualitySourceTypeCount struct {
	SourceType string `json:"source_type"`
	Count      int    `json:"count"`
}

type WorkflowRAGQualityKnowledgeSummary struct {
	FragmentCount          int                                 `json:"fragment_count"`
	OfficialFragmentCount  int                                 `json:"official_fragment_count"`
	ExpectedFragmentCount  int                                 `json:"expected_fragment_count"`
	SourceTypeDistribution []WorkflowRAGQualitySourceTypeCount `json:"source_type_distribution"`
}

type WorkflowRAGQualityFinding struct {
	Code         string   `json:"code"`
	Severity     string   `json:"severity"`
	FragmentRefs []string `json:"fragment_refs"`
}

type WorkflowRAGQualityDatasetBinding struct {
	DatasetID      string `json:"dataset_id"`
	DatasetVersion int    `json:"dataset_version"`
	DatasetDigest  string `json:"dataset_digest"`
}

type WorkflowRAGQualityReview struct {
	SchemaVersion    string                               `json:"schema_version"`
	Dataset          WorkflowRAGQualityDatasetBinding     `json:"dataset"`
	Snapshot         WorkflowRAGEvaluationSnapshotBinding `json:"snapshot"`
	Profile          WorkflowRAGEvaluationProfileBinding  `json:"profile"`
	Status           string                               `json:"status"`
	SampleCount      int                                  `json:"sample_count"`
	Thresholds       WorkflowRAGEvaluationThresholds      `json:"thresholds"`
	Metrics          WorkflowRAGQualityMetrics            `json:"metrics"`
	KnowledgeSummary WorkflowRAGQualityKnowledgeSummary   `json:"knowledge_summary"`
	Samples          []WorkflowRAGQualitySampleResult     `json:"samples"`
	Findings         []WorkflowRAGQualityFinding          `json:"findings"`
}

type workflowRAGQualityRanker func(string, []WorkflowRAGFragment, int) WorkflowRAGRankingResult

func DecodeWorkflowRAGEvaluationDataset(payload []byte) (WorkflowRAGEvaluationDataset, error) {
	var dataset WorkflowRAGEvaluationDataset
	if err := decodeWorkflowRAGStrictJSON(payload, &dataset); err != nil {
		return WorkflowRAGEvaluationDataset{}, errWorkflowRAGQualityReviewContract
	}
	if err := validateWorkflowRAGEvaluationDataset(dataset, nil); err != nil {
		return WorkflowRAGEvaluationDataset{}, err
	}
	return dataset, nil
}

func DecodeWorkflowRAGQualityReview(payload []byte) (WorkflowRAGQualityReview, error) {
	var review WorkflowRAGQualityReview
	if err := decodeWorkflowRAGStrictJSON(payload, &review); err != nil {
		return WorkflowRAGQualityReview{}, errWorkflowRAGQualityReviewContract
	}
	if err := validateWorkflowRAGQualityReview(review); err != nil {
		return WorkflowRAGQualityReview{}, err
	}
	return review, nil
}

func BuildWorkflowRAGQualityReview(snapshotPayload, datasetPayload []byte) (WorkflowRAGQualityReview, error) {
	var snapshot WorkflowRAGSnapshotRecord
	if err := decodeWorkflowRAGStrictJSON(snapshotPayload, &snapshot); err != nil {
		return WorkflowRAGQualityReview{}, errWorkflowRAGQualityReviewContract
	}
	dataset, err := DecodeWorkflowRAGEvaluationDataset(datasetPayload)
	if err != nil {
		return WorkflowRAGQualityReview{}, err
	}
	return reviewWorkflowRAGDataset(snapshot, dataset, RankWorkflowRAGFragments)
}

func WorkflowRAGEvaluationDatasetDigest(dataset WorkflowRAGEvaluationDataset) (string, error) {
	canonical := struct {
		SchemaVersion         string                               `json:"schema_version"`
		DatasetID             string                               `json:"dataset_id"`
		DatasetVersion        int                                  `json:"dataset_version"`
		ContentClassification string                               `json:"content_classification"`
		Snapshot              WorkflowRAGEvaluationSnapshotBinding `json:"snapshot"`
		Profile               WorkflowRAGEvaluationProfileBinding  `json:"profile"`
		Thresholds            WorkflowRAGEvaluationThresholds      `json:"thresholds"`
		Review                WorkflowRAGEvaluationReviewMetadata  `json:"review"`
		Samples               []WorkflowRAGEvaluationSample        `json:"samples"`
	}{
		SchemaVersion: dataset.SchemaVersion, DatasetID: dataset.DatasetID, DatasetVersion: dataset.DatasetVersion,
		ContentClassification: dataset.ContentClassification, Snapshot: dataset.Snapshot, Profile: dataset.Profile,
		Thresholds: dataset.Thresholds, Review: dataset.Review, Samples: dataset.Samples,
	}
	payload, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func reviewWorkflowRAGDataset(snapshot WorkflowRAGSnapshotRecord, dataset WorkflowRAGEvaluationDataset, ranker workflowRAGQualityRanker) (WorkflowRAGQualityReview, error) {
	if ranker == nil || validateWorkflowRAGEvaluationSnapshot(snapshot) != nil {
		return WorkflowRAGQualityReview{}, errWorkflowRAGQualityReviewContract
	}
	if err := validateWorkflowRAGEvaluationDataset(dataset, &snapshot); err != nil {
		return WorkflowRAGQualityReview{}, err
	}
	review := WorkflowRAGQualityReview{
		SchemaVersion: workflowRAGQualityReviewSchemaVersion,
		Dataset:       WorkflowRAGQualityDatasetBinding{DatasetID: dataset.DatasetID, DatasetVersion: dataset.DatasetVersion, DatasetDigest: dataset.DatasetDigest},
		Snapshot:      dataset.Snapshot, Profile: dataset.Profile, SampleCount: len(dataset.Samples), Thresholds: dataset.Thresholds,
		Samples: make([]WorkflowRAGQualitySampleResult, 0, len(dataset.Samples)), Findings: []WorkflowRAGQualityFinding{},
	}
	accumulator := workflowRAGQualityMetricAccumulator{}
	for _, sample := range dataset.Samples {
		ranking := ranker(sample.QueryText, snapshot.Fragments, sample.TopK)
		result := buildWorkflowRAGQualitySampleResult(sample, ranking)
		review.Samples = append(review.Samples, result)
		accumulator.add(sample, result)
	}
	review.Metrics = accumulator.metrics(len(dataset.Samples))
	review.KnowledgeSummary, review.Findings = buildWorkflowRAGKnowledgeReview(snapshot, dataset)
	review.Status = workflowRAGQualityStatus(review)
	if err := validateWorkflowRAGQualityReview(review); err != nil {
		return WorkflowRAGQualityReview{}, err
	}
	return review, nil
}

func validateWorkflowRAGEvaluationSnapshot(snapshot WorkflowRAGSnapshotRecord) error {
	ctx := WorkflowRAGSnapshotContext{TenantRef: snapshot.TenantRef, WorkspaceID: snapshot.WorkspaceID, ApplicationID: snapshot.ApplicationID}
	if validateStoredWorkflowRAGRecord(snapshot, ctx) != nil || snapshot.ContentClassification != "public" ||
		(snapshot.LifecycleState != workflowRAGSnapshotActive && snapshot.LifecycleState != workflowRAGSnapshotArchived) ||
		!workflowRAGReferencePattern.MatchString(snapshot.CreatedByActorRef) || !workflowRAGReferencePattern.MatchString(snapshot.RequestID) ||
		!workflowRAGReferencePattern.MatchString(snapshot.AuditRef) {
		return errWorkflowRAGQualityReviewContract
	}
	if _, err := time.Parse(time.RFC3339Nano, snapshot.CreatedAt); err != nil {
		return errWorkflowRAGQualityReviewContract
	}
	return nil
}

func validateWorkflowRAGEvaluationDataset(dataset WorkflowRAGEvaluationDataset, snapshot *WorkflowRAGSnapshotRecord) error {
	if dataset.SchemaVersion != workflowRAGEvaluationDatasetSchemaVersion ||
		!workflowRAGDatasetIDPattern.MatchString(dataset.DatasetID) || dataset.DatasetVersion < 1 ||
		dataset.ContentClassification != workflowRAGEvaluationClassification ||
		!workflowRAGDigestPattern.MatchString(dataset.DatasetDigest) || dataset.Samples == nil || len(dataset.Samples) < 1 || len(dataset.Samples) > 200 ||
		!workflowRAGReferencePattern.MatchString(dataset.Review.ReviewerRef) || strings.TrimSpace(dataset.Review.ReviewSummary) != dataset.Review.ReviewSummary ||
		dataset.Review.ReviewSummary == "" || utf8.RuneCountInString(dataset.Review.ReviewSummary) > 512 ||
		workflowRAGContainsForbiddenMaterial(dataset.Review.ReviewSummary) || validateWorkflowRAGQualityThresholds(dataset.Thresholds) != nil {
		return errWorkflowRAGQualityReviewContract
	}
	if _, err := time.Parse(time.RFC3339Nano, dataset.Review.ReviewedAt); err != nil {
		return errWorkflowRAGQualityReviewContract
	}
	profile := workflowRAGLexicalProfile()
	if dataset.Profile.ProfileID != profile.ProfileID || dataset.Profile.ProfileVersion != profile.ProfileVersion ||
		dataset.Profile.ProfileDigest != profile.ProfileDigest || !validateWorkflowRAGEvaluationSnapshotBinding(dataset.Snapshot) {
		return errWorkflowRAGQualityReviewContract
	}
	digest, err := WorkflowRAGEvaluationDatasetDigest(dataset)
	if err != nil || digest != dataset.DatasetDigest {
		return fmt.Errorf("%w: dataset digest mismatch; expected %s", errWorkflowRAGQualityReviewContract, digest)
	}
	fragmentByRef := map[string]WorkflowRAGFragment(nil)
	if snapshot != nil {
		if dataset.Snapshot != workflowRAGEvaluationSnapshotBinding(*snapshot) {
			return errWorkflowRAGQualityReviewContract
		}
		fragmentByRef = make(map[string]WorkflowRAGFragment, len(snapshot.Fragments))
		for _, fragment := range snapshot.Fragments {
			fragmentByRef[fragment.FragmentRef] = fragment
		}
	}
	seen := make(map[string]bool, len(dataset.Samples))
	for _, sample := range dataset.Samples {
		if seen[sample.SampleID] || validateWorkflowRAGEvaluationSample(sample, fragmentByRef) != nil {
			return errWorkflowRAGQualityReviewContract
		}
		seen[sample.SampleID] = true
	}
	return nil
}

func validateWorkflowRAGEvaluationSnapshotBinding(binding WorkflowRAGEvaluationSnapshotBinding) bool {
	return workflowRAGReferencePattern.MatchString(binding.TenantRef) && workflowRAGScopedIDPattern.MatchString(binding.WorkspaceID) &&
		workflowRAGScopedIDPattern.MatchString(binding.ApplicationID) && workflowRAGSnapshotIDPattern.MatchString(binding.SnapshotID) &&
		binding.SnapshotVersion > 0 && workflowRAGDigestPattern.MatchString(binding.SnapshotDigest) && workflowRAGRAGRefPattern.MatchString(binding.RAGRef)
}

func workflowRAGEvaluationSnapshotBinding(snapshot WorkflowRAGSnapshotRecord) WorkflowRAGEvaluationSnapshotBinding {
	return WorkflowRAGEvaluationSnapshotBinding{
		TenantRef: snapshot.TenantRef, WorkspaceID: snapshot.WorkspaceID, ApplicationID: snapshot.ApplicationID,
		SnapshotID: snapshot.SnapshotID, SnapshotVersion: snapshot.SnapshotVersion, SnapshotDigest: snapshot.SnapshotDigest, RAGRef: snapshot.RAGRef,
	}
}

func validateWorkflowRAGEvaluationSample(sample WorkflowRAGEvaluationSample, fragments map[string]WorkflowRAGFragment) error {
	if !workflowRAGSampleIDPattern.MatchString(sample.SampleID) || strings.TrimSpace(sample.QueryText) != sample.QueryText ||
		sample.QueryText == "" || !utf8.ValidString(sample.QueryText) || len([]byte(sample.QueryText)) > workflowRAGMaxQueryBytes ||
		workflowRAGContainsForbiddenMaterial(sample.QueryText) || sample.TopK < 1 || sample.TopK > workflowRAGMaxTopK ||
		strings.TrimSpace(sample.ReviewNote) != sample.ReviewNote || sample.ReviewNote == "" || utf8.RuneCountInString(sample.ReviewNote) > 512 ||
		workflowRAGContainsForbiddenMaterial(sample.ReviewNote) || sample.ExpectedCitationRefs == nil || sample.RequiredOfficialRefs == nil ||
		!workflowRAGUniqueRefs(sample.ExpectedCitationRefs) || !workflowRAGUniqueRefs(sample.RequiredOfficialRefs) {
		return errWorkflowRAGQualityReviewContract
	}
	if sample.Expectation == "evidence_required" {
		if len(sample.ExpectedCitationRefs) == 0 {
			return errWorkflowRAGQualityReviewContract
		}
	} else if sample.Expectation == "no_evidence" {
		if len(sample.ExpectedCitationRefs) != 0 || len(sample.RequiredOfficialRefs) != 0 {
			return errWorkflowRAGQualityReviewContract
		}
	} else {
		return errWorkflowRAGQualityReviewContract
	}
	expected := make(map[string]bool, len(sample.ExpectedCitationRefs))
	for _, ref := range sample.ExpectedCitationRefs {
		if !workflowRAGFragmentRefPattern.MatchString(ref) {
			return errWorkflowRAGQualityReviewContract
		}
		expected[ref] = true
		if fragments != nil {
			if _, ok := fragments[ref]; !ok {
				return errWorkflowRAGQualityReviewContract
			}
		}
	}
	for _, ref := range sample.RequiredOfficialRefs {
		if !expected[ref] {
			return errWorkflowRAGQualityReviewContract
		}
		if fragments != nil && !fragments[ref].IsOfficial {
			return errWorkflowRAGQualityReviewContract
		}
	}
	return nil
}

func workflowRAGUniqueRefs(refs []string) bool {
	seen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if seen[ref] {
			return false
		}
		seen[ref] = true
	}
	return true
}

func validateWorkflowRAGQualityThresholds(thresholds WorkflowRAGEvaluationThresholds) error {
	for _, value := range []float64{
		thresholds.HitAtK, thresholds.ExpectedRecallAtK, thresholds.RequiredOfficialRecallAtK,
		thresholds.MeanReciprocalRank, thresholds.NoEvidenceAccuracy, thresholds.SamplePassRate,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return errWorkflowRAGQualityReviewContract
		}
	}
	return nil
}

func buildWorkflowRAGQualitySampleResult(sample WorkflowRAGEvaluationSample, ranking WorkflowRAGRankingResult) WorkflowRAGQualitySampleResult {
	result := WorkflowRAGQualitySampleResult{
		SampleID: sample.SampleID, Expectation: sample.Expectation, QueryDigest: ranking.QueryDigest, QueryBytes: ranking.QueryBytes,
		TopK: sample.TopK, FailureCode: ranking.FailureCode, Selected: make([]WorkflowRAGQualitySelectedFragment, 0, len(ranking.Selected)),
		ExpectedRefCount: len(sample.ExpectedCitationRefs), RequiredOfficialRefCount: len(sample.RequiredOfficialRefs),
	}
	selectedRank := make(map[string]int, len(ranking.Selected))
	for _, fragment := range ranking.Selected {
		result.Selected = append(result.Selected, WorkflowRAGQualitySelectedFragment{
			FragmentRef: fragment.FragmentRef, Rank: fragment.Rank, ContentDigest: fragment.ContentDigest,
			SourceType: fragment.SourceType, IsOfficial: fragment.IsOfficial,
		})
		selectedRank[fragment.FragmentRef] = fragment.Rank
	}
	for _, ref := range sample.ExpectedCitationRefs {
		if rank := selectedRank[ref]; rank > 0 {
			result.ExpectedHitCount++
			if result.FirstExpectedRank == 0 || rank < result.FirstExpectedRank {
				result.FirstExpectedRank = rank
			}
		}
	}
	for _, ref := range sample.RequiredOfficialRefs {
		if selectedRank[ref] > 0 {
			result.RequiredOfficialHitCount++
		}
	}
	if sample.Expectation == "no_evidence" {
		result.Passed = ranking.FailureCode == WorkflowRAGFailureNoEvidence && len(ranking.Selected) == 0
	} else {
		result.Passed = ranking.FailureCode == "" && result.ExpectedHitCount == result.ExpectedRefCount &&
			result.RequiredOfficialHitCount == result.RequiredOfficialRefCount
	}
	return result
}

type workflowRAGQualityMetricAccumulator struct {
	positiveCount       int
	negativeCount       int
	hitSamples          int
	expectedRefs        int
	expectedHits        int
	officialRefs        int
	officialHits        int
	noEvidencePasses    int
	passedSamples       int
	reciprocalRankTotal float64
}

func (accumulator *workflowRAGQualityMetricAccumulator) add(sample WorkflowRAGEvaluationSample, result WorkflowRAGQualitySampleResult) {
	if result.Passed {
		accumulator.passedSamples++
	}
	if sample.Expectation == "no_evidence" {
		accumulator.negativeCount++
		if result.Passed {
			accumulator.noEvidencePasses++
		}
		return
	}
	accumulator.positiveCount++
	accumulator.expectedRefs += result.ExpectedRefCount
	accumulator.expectedHits += result.ExpectedHitCount
	accumulator.officialRefs += result.RequiredOfficialRefCount
	accumulator.officialHits += result.RequiredOfficialHitCount
	if result.ExpectedHitCount > 0 {
		accumulator.hitSamples++
	}
	if result.FirstExpectedRank > 0 {
		accumulator.reciprocalRankTotal += 1 / float64(result.FirstExpectedRank)
	}
}

func (accumulator workflowRAGQualityMetricAccumulator) metrics(sampleCount int) WorkflowRAGQualityMetrics {
	return WorkflowRAGQualityMetrics{
		HitAtK:                    workflowRAGQualityRatio(accumulator.hitSamples, accumulator.positiveCount),
		ExpectedRecallAtK:         workflowRAGQualityRatio(accumulator.expectedHits, accumulator.expectedRefs),
		RequiredOfficialRecallAtK: workflowRAGQualityRatio(accumulator.officialHits, accumulator.officialRefs),
		MeanReciprocalRank:        workflowRAGQualityAverage(accumulator.reciprocalRankTotal, accumulator.positiveCount),
		NoEvidenceAccuracy:        workflowRAGQualityRatio(accumulator.noEvidencePasses, accumulator.negativeCount),
		SamplePassRate:            workflowRAGQualityRatio(accumulator.passedSamples, sampleCount),
	}
}

func workflowRAGQualityRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 1
	}
	return workflowRAGQualityRound(float64(numerator) / float64(denominator))
}

func workflowRAGQualityAverage(total float64, denominator int) float64 {
	if denominator == 0 {
		return 1
	}
	return workflowRAGQualityRound(total / float64(denominator))
}

func workflowRAGQualityRound(value float64) float64 {
	return math.Round(value*1_000_000) / 1_000_000
}

func buildWorkflowRAGKnowledgeReview(snapshot WorkflowRAGSnapshotRecord, dataset WorkflowRAGEvaluationDataset) (WorkflowRAGQualityKnowledgeSummary, []WorkflowRAGQualityFinding) {
	sourceCounts := make(map[string]int)
	digestRefs := make(map[string][]string)
	officialRefs := make([]string, 0)
	for _, fragment := range snapshot.Fragments {
		sourceCounts[fragment.SourceType]++
		digestRefs[fragment.ContentDigest] = append(digestRefs[fragment.ContentDigest], fragment.FragmentRef)
		if fragment.IsOfficial {
			officialRefs = append(officialRefs, fragment.FragmentRef)
		}
	}
	expected := make(map[string]bool)
	for _, sample := range dataset.Samples {
		if sample.Expectation == "evidence_required" {
			for _, ref := range sample.ExpectedCitationRefs {
				expected[ref] = true
			}
		}
	}
	sourceTypes := make([]string, 0, len(sourceCounts))
	for sourceType := range sourceCounts {
		sourceTypes = append(sourceTypes, sourceType)
	}
	sort.Strings(sourceTypes)
	summary := WorkflowRAGQualityKnowledgeSummary{
		FragmentCount: len(snapshot.Fragments), OfficialFragmentCount: len(officialRefs), ExpectedFragmentCount: len(expected),
		SourceTypeDistribution: make([]WorkflowRAGQualitySourceTypeCount, 0, len(sourceTypes)),
	}
	for _, sourceType := range sourceTypes {
		summary.SourceTypeDistribution = append(summary.SourceTypeDistribution, WorkflowRAGQualitySourceTypeCount{SourceType: sourceType, Count: sourceCounts[sourceType]})
	}
	findings := make([]WorkflowRAGQualityFinding, 0)
	digests := make([]string, 0, len(digestRefs))
	for digest := range digestRefs {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	for _, digest := range digests {
		refs := digestRefs[digest]
		if len(refs) > 1 {
			sort.Strings(refs)
			findings = append(findings, WorkflowRAGQualityFinding{Code: "duplicate_fragment_content", Severity: "review_required", FragmentRefs: refs})
		}
	}
	if len(officialRefs) == 0 {
		findings = append(findings, WorkflowRAGQualityFinding{Code: "official_evidence_absent", Severity: "review_required", FragmentRefs: []string{}})
	}
	uncovered := make([]string, 0)
	for _, fragment := range snapshot.Fragments {
		if !expected[fragment.FragmentRef] {
			uncovered = append(uncovered, fragment.FragmentRef)
		}
	}
	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		findings = append(findings, WorkflowRAGQualityFinding{Code: "fragment_expectation_uncovered", Severity: "info", FragmentRefs: uncovered})
	}
	partialOfficial := make([]string, 0)
	for _, ref := range officialRefs {
		if !expected[ref] {
			partialOfficial = append(partialOfficial, ref)
		}
	}
	if len(partialOfficial) > 0 {
		sort.Strings(partialOfficial)
		findings = append(findings, WorkflowRAGQualityFinding{Code: "official_expectation_coverage_partial", Severity: "info", FragmentRefs: partialOfficial})
	}
	return summary, findings
}

func workflowRAGQualityStatus(review WorkflowRAGQualityReview) string {
	metrics := review.Metrics
	thresholds := review.Thresholds
	if metrics.HitAtK < thresholds.HitAtK || metrics.ExpectedRecallAtK < thresholds.ExpectedRecallAtK ||
		metrics.RequiredOfficialRecallAtK < thresholds.RequiredOfficialRecallAtK || metrics.MeanReciprocalRank < thresholds.MeanReciprocalRank ||
		metrics.NoEvidenceAccuracy < thresholds.NoEvidenceAccuracy || metrics.SamplePassRate < thresholds.SamplePassRate {
		return "failed"
	}
	for _, sample := range review.Samples {
		if !sample.Passed {
			return "failed"
		}
	}
	for _, finding := range review.Findings {
		if finding.Severity == "review_required" {
			return "needs_review"
		}
	}
	return "passed"
}

func validateWorkflowRAGQualityReview(review WorkflowRAGQualityReview) error {
	if review.SchemaVersion != workflowRAGQualityReviewSchemaVersion || !workflowRAGDatasetIDPattern.MatchString(review.Dataset.DatasetID) ||
		review.Dataset.DatasetVersion < 1 || !workflowRAGDigestPattern.MatchString(review.Dataset.DatasetDigest) ||
		!validateWorkflowRAGEvaluationSnapshotBinding(review.Snapshot) || review.Profile.ProfileID != workflowRAGProfileID ||
		review.Profile.ProfileVersion != workflowRAGProfileVersion || review.Profile.ProfileDigest != workflowRAGLexicalProfile().ProfileDigest ||
		(review.Status != "passed" && review.Status != "needs_review" && review.Status != "failed") || review.Samples == nil || review.Findings == nil ||
		review.KnowledgeSummary.SourceTypeDistribution == nil || review.SampleCount != len(review.Samples) ||
		validateWorkflowRAGQualityThresholds(review.Thresholds) != nil || validateWorkflowRAGQualityMetrics(review.Metrics) != nil ||
		review.KnowledgeSummary.FragmentCount < 1 || review.KnowledgeSummary.OfficialFragmentCount < 0 ||
		review.KnowledgeSummary.OfficialFragmentCount > review.KnowledgeSummary.FragmentCount || review.KnowledgeSummary.ExpectedFragmentCount < 0 ||
		review.KnowledgeSummary.ExpectedFragmentCount > review.KnowledgeSummary.FragmentCount {
		return errWorkflowRAGQualityReviewContract
	}
	seen := make(map[string]bool, len(review.Samples))
	for _, sample := range review.Samples {
		if seen[sample.SampleID] || !workflowRAGSampleIDPattern.MatchString(sample.SampleID) ||
			(sample.Expectation != "evidence_required" && sample.Expectation != "no_evidence") || !workflowRAGDigestPattern.MatchString(sample.QueryDigest) ||
			sample.QueryBytes < 1 || sample.QueryBytes > workflowRAGMaxQueryBytes || sample.TopK < 1 || sample.TopK > workflowRAGMaxTopK ||
			sample.Selected == nil || len(sample.Selected) > sample.TopK || !workflowRAGQualityFailureCodeAllowed(sample.FailureCode) ||
			sample.ExpectedHitCount > sample.ExpectedRefCount || sample.RequiredOfficialHitCount > sample.RequiredOfficialRefCount ||
			sample.FirstExpectedRank < 0 || sample.FirstExpectedRank > sample.TopK {
			return errWorkflowRAGQualityReviewContract
		}
		if (sample.Expectation == "evidence_required" && sample.ExpectedRefCount < 1) ||
			(sample.Expectation == "no_evidence" && (sample.ExpectedRefCount != 0 || sample.RequiredOfficialRefCount != 0 || sample.FirstExpectedRank != 0)) {
			return errWorkflowRAGQualityReviewContract
		}
		seen[sample.SampleID] = true
		selectedRefs := make(map[string]bool, len(sample.Selected))
		for index, selected := range sample.Selected {
			if !workflowRAGFragmentRefPattern.MatchString(selected.FragmentRef) || selected.Rank != index+1 ||
				selectedRefs[selected.FragmentRef] || !workflowRAGDigestPattern.MatchString(selected.ContentDigest) || !workflowRAGSourceTypeAllowed(selected.SourceType) {
				return errWorkflowRAGQualityReviewContract
			}
			selectedRefs[selected.FragmentRef] = true
		}
	}
	sourceCount := 0
	seenSourceTypes := make(map[string]bool, len(review.KnowledgeSummary.SourceTypeDistribution))
	for _, distribution := range review.KnowledgeSummary.SourceTypeDistribution {
		if !workflowRAGSourceTypeAllowed(distribution.SourceType) || seenSourceTypes[distribution.SourceType] || distribution.Count < 1 {
			return errWorkflowRAGQualityReviewContract
		}
		seenSourceTypes[distribution.SourceType] = true
		sourceCount += distribution.Count
	}
	if sourceCount != review.KnowledgeSummary.FragmentCount {
		return errWorkflowRAGQualityReviewContract
	}
	for _, finding := range review.Findings {
		if finding.FragmentRefs == nil || !workflowRAGQualityFindingAllowed(finding.Code, finding.Severity) || !workflowRAGUniqueRefs(finding.FragmentRefs) {
			return errWorkflowRAGQualityReviewContract
		}
		for _, ref := range finding.FragmentRefs {
			if !workflowRAGFragmentRefPattern.MatchString(ref) {
				return errWorkflowRAGQualityReviewContract
			}
		}
	}
	if review.Status != workflowRAGQualityStatus(review) {
		return errWorkflowRAGQualityReviewContract
	}
	return nil
}

func validateWorkflowRAGQualityMetrics(metrics WorkflowRAGQualityMetrics) error {
	return validateWorkflowRAGQualityThresholds(WorkflowRAGEvaluationThresholds(metrics))
}

func workflowRAGQualityFindingAllowed(code, severity string) bool {
	switch code {
	case "duplicate_fragment_content", "official_evidence_absent":
		return severity == "review_required"
	case "fragment_expectation_uncovered", "official_expectation_coverage_partial":
		return severity == "info"
	default:
		return false
	}
}

func workflowRAGQualityFailureCodeAllowed(value string) bool {
	return value == "" || workflowRAGRunFailurePattern.MatchString(value)
}
