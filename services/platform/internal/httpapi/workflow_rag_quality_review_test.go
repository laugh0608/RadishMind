package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWorkflowRAGQualityReviewCommittedAssets(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve quality review test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../.."))
	snapshotPayload, err := os.ReadFile(filepath.Join(root, "datasets/eval/workflow-rag/snapshots/dev_core_v1.json"))
	if err != nil {
		t.Fatalf("read committed snapshot: %v", err)
	}
	var snapshot WorkflowRAGSnapshotRecord
	if err = decodeWorkflowRAGStrictJSON(snapshotPayload, &snapshot); err != nil {
		t.Fatalf("decode committed snapshot: %v", err)
	}
	expectedSnapshotDigest, err := workflowRAGSnapshotDigest(snapshot)
	if err != nil || snapshot.SnapshotDigest != expectedSnapshotDigest {
		t.Fatalf("committed snapshot digest drifted: got=%s expected=%s err=%v", snapshot.SnapshotDigest, expectedSnapshotDigest, err)
	}
	datasetPayload, err := os.ReadFile(filepath.Join(root, "datasets/eval/workflow-rag/datasets/dev_core_v1.json"))
	if err != nil {
		t.Fatalf("read committed dataset: %v", err)
	}
	var dataset WorkflowRAGEvaluationDataset
	if err = decodeWorkflowRAGStrictJSON(datasetPayload, &dataset); err != nil {
		t.Fatalf("decode committed dataset: %v", err)
	}
	expectedDatasetDigest, err := WorkflowRAGEvaluationDatasetDigest(dataset)
	if err != nil || dataset.DatasetDigest != expectedDatasetDigest {
		t.Fatalf("committed dataset digest drifted: got=%s expected=%s err=%v", dataset.DatasetDigest, expectedDatasetDigest, err)
	}
	review, err := BuildWorkflowRAGQualityReview(snapshotPayload, datasetPayload)
	if err != nil || review.Status != "passed" || review.SampleCount != 7 {
		t.Fatalf("review committed assets: status=%s samples=%d err=%v", review.Status, review.SampleCount, err)
	}
	committedReport, err := os.ReadFile(filepath.Join(root, "datasets/eval/workflow-rag/reports/dev_core_v1.review.json"))
	if err != nil {
		t.Fatalf("read committed quality review: %v", err)
	}
	if _, err = DecodeWorkflowRAGQualityReview(committedReport); err != nil {
		t.Fatalf("validate committed quality review: %v", err)
	}
	expectedReport, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		t.Fatalf("encode expected quality review: %v", err)
	}
	expectedReport = append(expectedReport, '\n')
	if string(committedReport) != string(expectedReport) {
		t.Fatal("committed quality review does not match deterministic evaluator output")
	}
}

func TestWorkflowRAGQualityReviewIsDeterministicMetadataOnlyAndRanksOnce(t *testing.T) {
	snapshot := workflowRAGQualityTestSnapshot(t, workflowRAGQualityTestFragments())
	dataset := workflowRAGQualityTestDataset(t, snapshot, []WorkflowRAGEvaluationSample{
		{
			SampleID: "latin_official", QueryText: "retrieval ranking policy", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"official_retrieval"}, RequiredOfficialRefs: []string{"official_retrieval"},
			TopK: 2, ReviewNote: "The official retrieval policy is required.",
		},
		{
			SampleID: "cjk_guide", QueryText: "知识引用规则", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"official_chinese"}, RequiredOfficialRefs: []string{"official_chinese"},
			TopK: 1, ReviewNote: "The Chinese official guide is the accepted evidence.",
		},
		{
			SampleID: "supplemental_forum", QueryText: "community troubleshooting example", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"community_example"}, RequiredOfficialRefs: []string{},
			TopK: 1, ReviewNote: "The supplemental example is intentionally reviewable.",
		},
		{
			SampleID: "no_evidence", QueryText: "quantum orchard invoice", Expectation: "no_evidence",
			ExpectedCitationRefs: []string{}, RequiredOfficialRefs: []string{}, TopK: 2,
			ReviewNote: "No fragment discusses this synthetic topic.",
		},
	})
	calls := 0
	ranker := func(query string, fragments []WorkflowRAGFragment, topK int) WorkflowRAGRankingResult {
		calls++
		return RankWorkflowRAGFragments(query, fragments, topK)
	}
	first, err := reviewWorkflowRAGDataset(snapshot, dataset, ranker)
	if err != nil {
		t.Fatalf("review deterministic dataset: %v", err)
	}
	second, err := reviewWorkflowRAGDataset(snapshot, dataset, RankWorkflowRAGFragments)
	if err != nil {
		t.Fatalf("repeat deterministic dataset: %v", err)
	}
	if calls != len(dataset.Samples) {
		t.Fatalf("ranker calls=%d, want exactly one per sample (%d)", calls, len(dataset.Samples))
	}
	left, _ := json.Marshal(first)
	right, _ := json.Marshal(second)
	if string(left) != string(right) || first.Status != "passed" || first.Metrics != (WorkflowRAGQualityMetrics{
		HitAtK: 1, ExpectedRecallAtK: 1, RequiredOfficialRecallAtK: 1,
		MeanReciprocalRank: 1, NoEvidenceAccuracy: 1, SamplePassRate: 1,
	}) {
		t.Fatalf("quality review drifted:\nfirst=%s\nsecond=%s", left, right)
	}
	for _, forbidden := range []string{
		"retrieval ranking policy", "知识引用规则", "community troubleshooting example", "quantum orchard invoice",
		"Official retrieval policy requires deterministic lexical ranking.", "The official retrieval policy is required.",
	} {
		if strings.Contains(string(left), forbidden) {
			t.Fatalf("metadata-only report leaked %q: %s", forbidden, left)
		}
	}
}

func TestWorkflowRAGQualityDatasetRejectsStrictJSONDigestBindingAndSamples(t *testing.T) {
	snapshot := workflowRAGQualityTestSnapshot(t, workflowRAGQualityTestFragments())
	base := workflowRAGQualityTestDataset(t, snapshot, []WorkflowRAGEvaluationSample{{
		SampleID: "official_sample", QueryText: "retrieval ranking policy", Expectation: "evidence_required",
		ExpectedCitationRefs: []string{"official_retrieval"}, RequiredOfficialRefs: []string{"official_retrieval"},
		TopK: 2, ReviewNote: "Official evidence is required.",
	}})
	payload, _ := json.Marshal(base)
	unknown := strings.TrimSuffix(string(payload), "}") + `,"unknown":true}`
	if _, err := DecodeWorkflowRAGEvaluationDataset([]byte(unknown)); err == nil {
		t.Fatal("dataset accepted an unknown field")
	}
	if _, err := DecodeWorkflowRAGEvaluationDataset(append(payload, []byte(` {}`)...)); err == nil {
		t.Fatal("dataset accepted trailing JSON")
	}
	digestDrift := base
	digestDrift.Samples[0].QueryText = "changed query"
	if err := validateWorkflowRAGEvaluationDataset(digestDrift, &snapshot); err == nil {
		t.Fatal("dataset accepted digest drift")
	}
	tests := []struct {
		name   string
		mutate func(*WorkflowRAGEvaluationDataset)
	}{
		{name: "scope", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Snapshot.ApplicationID = "app_other" }},
		{name: "profile", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Profile.ProfileDigest = workflowRAGSHA256("other") }},
		{name: "empty query", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples[0].QueryText = "" }},
		{name: "query whitespace", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples[0].QueryText = " query" }},
		{name: "query secret", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples[0].QueryText = "password=forbidden" }},
		{name: "missing ref", mutate: func(value *WorkflowRAGEvaluationDataset) {
			value.Samples[0].ExpectedCitationRefs = []string{"missing_fragment"}
			value.Samples[0].RequiredOfficialRefs = nil
		}},
		{name: "nonofficial required", mutate: func(value *WorkflowRAGEvaluationDataset) {
			value.Samples[0].ExpectedCitationRefs = []string{"community_example"}
			value.Samples[0].RequiredOfficialRefs = []string{"community_example"}
		}},
		{name: "duplicate sample", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples = append(value.Samples, value.Samples[0]) }},
		{name: "negative with refs", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples[0].Expectation = "no_evidence" }},
		{name: "top k", mutate: func(value *WorkflowRAGEvaluationDataset) { value.Samples[0].TopK = workflowRAGMaxTopK + 1 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := workflowRAGCloneEvaluationDataset(base)
			test.mutate(&candidate)
			workflowRAGRefreshEvaluationDatasetDigest(t, &candidate)
			if err := validateWorkflowRAGEvaluationDataset(candidate, &snapshot); err == nil {
				t.Fatalf("dataset accepted invalid %s", test.name)
			}
		})
	}
	tamperedSnapshot := snapshot
	tamperedSnapshot.Fragments[0].Content = "tampered content"
	if _, err := reviewWorkflowRAGDataset(tamperedSnapshot, base, RankWorkflowRAGFragments); err == nil {
		t.Fatal("quality review accepted a tampered snapshot digest")
	}
}

func TestWorkflowRAGQualityReviewMetricsThresholdsAndFindings(t *testing.T) {
	snapshot := workflowRAGQualityTestSnapshot(t, workflowRAGQualityTestFragments())
	dataset := workflowRAGQualityTestDataset(t, snapshot, []WorkflowRAGEvaluationSample{{
		SampleID: "multi_expected", QueryText: "retrieval policy community example", Expectation: "evidence_required",
		ExpectedCitationRefs: []string{"official_retrieval", "community_example"}, RequiredOfficialRefs: []string{"official_retrieval"},
		TopK: 1, ReviewNote: "Both accepted sources must be retrieved for this sample.",
	}})
	review, err := reviewWorkflowRAGDataset(snapshot, dataset, RankWorkflowRAGFragments)
	if err != nil {
		t.Fatalf("review threshold failure: %v", err)
	}
	if review.Status != "failed" || review.Metrics.HitAtK != 1 || review.Metrics.ExpectedRecallAtK != 0.5 ||
		review.Metrics.SamplePassRate != 0 || review.Samples[0].Passed {
		t.Fatalf("unexpected partial-recall result: %#v", review)
	}

	duplicateInputs := workflowRAGQualityTestFragments()
	duplicateInputs[1].Content = duplicateInputs[0].Content
	duplicate := workflowRAGQualityTestSnapshot(t, duplicateInputs)
	duplicateDataset := workflowRAGQualityTestDataset(t, duplicate, []WorkflowRAGEvaluationSample{
		{SampleID: "duplicate_coverage", QueryText: "retrieval ranking policy and community example", Expectation: "evidence_required", ExpectedCitationRefs: []string{"community_example", "official_chinese", "official_retrieval"}, RequiredOfficialRefs: []string{"official_chinese", "official_retrieval"}, TopK: 3, ReviewNote: "Every fragment remains expected while duplicate content requires review."},
	})
	duplicateReview, err := reviewWorkflowRAGDataset(duplicate, duplicateDataset, func(query string, fragments []WorkflowRAGFragment, topK int) WorkflowRAGRankingResult {
		selected := make([]WorkflowRAGRankedFragment, 0, topK)
		for index, fragment := range fragments {
			if len(selected) == topK {
				break
			}
			selected = append(selected, WorkflowRAGRankedFragment{FragmentRef: fragment.FragmentRef, Rank: index + 1, ContentDigest: fragment.ContentDigest, SourceType: fragment.SourceType, IsOfficial: fragment.IsOfficial})
		}
		return WorkflowRAGRankingResult{Profile: workflowRAGLexicalProfile(), QueryDigest: workflowRAGSHA256(query), QueryBytes: len([]byte(query)), Selected: selected}
	})
	if err != nil {
		t.Fatalf("review duplicate knowledge: %v", err)
	}
	if duplicateReview.Status != "needs_review" || len(duplicateReview.Findings) == 0 || duplicateReview.Findings[0].Code != "duplicate_fragment_content" {
		t.Fatalf("duplicate content did not require review: %#v", duplicateReview)
	}
}

func TestWorkflowRAGQualityReviewStrictDecoderRejectsReportDrift(t *testing.T) {
	snapshot := workflowRAGQualityTestSnapshot(t, workflowRAGQualityTestFragments())
	dataset := workflowRAGQualityTestDataset(t, snapshot, []WorkflowRAGEvaluationSample{{
		SampleID: "official_sample", QueryText: "retrieval ranking policy", Expectation: "evidence_required",
		ExpectedCitationRefs: []string{"official_retrieval"}, RequiredOfficialRefs: []string{"official_retrieval"},
		TopK: 1, ReviewNote: "Official evidence is required.",
	}})
	review, err := reviewWorkflowRAGDataset(snapshot, dataset, RankWorkflowRAGFragments)
	if err != nil {
		t.Fatalf("review report fixture: %v", err)
	}
	payload, _ := json.Marshal(review)
	if _, err = DecodeWorkflowRAGQualityReview(payload); err != nil {
		t.Fatalf("strict decoder rejected valid report: %v", err)
	}
	unknown := strings.TrimSuffix(string(payload), "}") + `,"query_text":"must-not-exist"}`
	if _, err = DecodeWorkflowRAGQualityReview([]byte(unknown)); err == nil {
		t.Fatal("quality report accepted raw query field")
	}
}

func workflowRAGQualityTestSnapshot(t *testing.T, inputs []WorkflowRAGFragmentInput) WorkflowRAGSnapshotRecord {
	t.Helper()
	key, displayName, classification, fragments, totalBytes, failure := normalizeWorkflowRAGSnapshotInput(
		"quality_review", "Synthetic quality review knowledge", "public", inputs,
	)
	if failure != "" {
		t.Fatalf("normalize quality snapshot: %s", failure)
	}
	ctx := WorkflowRAGSnapshotContext{
		TenantRef: "tenant_eval", WorkspaceID: "workspace_eval", ApplicationID: "app_eval",
		ActorRef: "subject_quality_reviewer", RequestID: "request_quality_fixture", AuditRef: "audit_quality_fixture",
	}
	record, err := buildWorkflowRAGSnapshotRecord(
		ctx, "rags_eeeeeeeeeeeeeeee", key, displayName, classification, 1,
		time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC).Format(time.RFC3339), fragments, totalBytes,
	)
	if err != nil {
		t.Fatalf("build quality snapshot: %v", err)
	}
	return record
}

func workflowRAGQualityTestFragments() []WorkflowRAGFragmentInput {
	return []WorkflowRAGFragmentInput{
		{FragmentRef: "official_retrieval", SourceType: "manual", SourceRef: "manual.quality", PageSlug: "quality/retrieval", Title: "Official retrieval policy", IsOfficial: true, Content: "Official retrieval policy requires deterministic lexical ranking."},
		{FragmentRef: "official_chinese", SourceType: "wiki", SourceRef: "wiki.quality", PageSlug: "quality/chinese", Title: "知识引用规则", IsOfficial: true, Content: "知识引用规则要求检索结果引用明确的证据片段。"},
		{FragmentRef: "community_example", SourceType: "forum", SourceRef: "forum.quality", PageSlug: "quality/community", Title: "Community troubleshooting example", Content: "Community troubleshooting example supplements the official manual."},
	}
}

func workflowRAGQualityTestDataset(t *testing.T, snapshot WorkflowRAGSnapshotRecord, samples []WorkflowRAGEvaluationSample) WorkflowRAGEvaluationDataset {
	t.Helper()
	profile := workflowRAGLexicalProfile()
	dataset := WorkflowRAGEvaluationDataset{
		SchemaVersion: workflowRAGEvaluationDatasetSchemaVersion, DatasetID: "wragd_quality_review", DatasetVersion: 1,
		ContentClassification: workflowRAGEvaluationClassification, Snapshot: workflowRAGEvaluationSnapshotBinding(snapshot),
		Profile:    WorkflowRAGEvaluationProfileBinding{ProfileID: profile.ProfileID, ProfileVersion: profile.ProfileVersion, ProfileDigest: profile.ProfileDigest},
		Thresholds: WorkflowRAGEvaluationThresholds{HitAtK: 1, ExpectedRecallAtK: 1, RequiredOfficialRecallAtK: 1, MeanReciprocalRank: 1, NoEvidenceAccuracy: 1, SamplePassRate: 1},
		Review:     WorkflowRAGEvaluationReviewMetadata{ReviewerRef: "subject_quality_reviewer", ReviewedAt: "2026-07-18T08:05:00Z", ReviewSummary: "Synthetic-public evidence expectations were reviewed manually."},
		Samples:    samples,
	}
	workflowRAGRefreshEvaluationDatasetDigest(t, &dataset)
	return dataset
}

func workflowRAGRefreshEvaluationDatasetDigest(t *testing.T, dataset *WorkflowRAGEvaluationDataset) {
	t.Helper()
	digest, err := WorkflowRAGEvaluationDatasetDigest(*dataset)
	if err != nil {
		t.Fatalf("compute evaluation dataset digest: %v", err)
	}
	dataset.DatasetDigest = digest
}

func workflowRAGCloneEvaluationDataset(dataset WorkflowRAGEvaluationDataset) WorkflowRAGEvaluationDataset {
	payload, _ := json.Marshal(dataset)
	var clone WorkflowRAGEvaluationDataset
	_ = json.Unmarshal(payload, &clone)
	return clone
}
