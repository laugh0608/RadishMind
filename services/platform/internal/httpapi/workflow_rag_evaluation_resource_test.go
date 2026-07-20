package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestWorkflowRAGEvaluationResourceMemoryLifecycleAndCandidateReview(t *testing.T) {
	ctx := workflowRAGTestContext()
	snapshots := newMemoryWorkflowRAGSnapshotRepository(nil)
	baseline := createWorkflowRAGEvaluationTestSnapshot(t, snapshots, ctx, "rags_aaaaaaaaaaaaaaaa", "baseline_knowledge", "public", []WorkflowRAGFragmentInput{
		{FragmentRef: "official_guide", SourceType: "manual", SourceRef: "manual.baseline", PageSlug: "baseline/guide", Title: "Baseline guide", IsOfficial: true, Content: "approved pressure relief procedure"},
	})
	candidate := createWorkflowRAGEvaluationTestSnapshot(t, snapshots, ctx, "rags_bbbbbbbbbbbbbbbb", "candidate_knowledge", "public", []WorkflowRAGFragmentInput{
		{FragmentRef: "candidate_note", SourceType: "wiki", SourceRef: "wiki.candidate", PageSlug: "candidate/note", Title: "Candidate note", Content: "unrelated maintenance schedule"},
	})

	repository := newMemoryWorkflowRAGEvaluationDatasetRepository(nil)
	service := newWorkflowRAGEvaluationDatasetService(repository, snapshots)
	now := time.Date(2026, 7, 18, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.newID = workflowRAGEvaluationTestIDGenerator()
	created := service.Create(ctx, WorkflowRAGEvaluationDatasetCreateInput{
		DatasetKey: "pressure_review", DisplayName: "Pressure review", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "Human-reviewed pressure evidence.", Samples: []WorkflowRAGEvaluationSample{{
			SampleID: "pressure_relief", QueryText: "approved pressure relief procedure", Expectation: "evidence_required",
			ExpectedCitationRefs: []string{"official_guide"}, RequiredOfficialRefs: []string{"official_guide"}, TopK: 1, ReviewNote: "The official guide is required.",
		}},
	})
	if created.FailureCode != "" || created.Resource == nil || created.Version == nil || created.Resource.LatestVersion != 1 || !workflowRAGDigestPattern.MatchString(created.Version.Dataset.DatasetDigest) {
		t.Fatalf("create evaluation dataset: %#v", created)
	}
	listed := service.List(ctx, WorkflowRAGEvaluationListInput{})
	if listed.FailureCode != "" || len(listed.Resources) != 1 || listed.Resources[0].DatasetKey != "pressure_review" {
		t.Fatalf("list evaluation datasets: %#v", listed)
	}
	listPayload, _ := json.Marshal(listed)
	if strings.Contains(string(listPayload), "pressure relief procedure") {
		t.Fatalf("dataset list leaked query content: %s", listPayload)
	}

	rankerCalls := 0
	service.ranker = func(query string, fragments []WorkflowRAGFragment, topK int) WorkflowRAGRankingResult {
		rankerCalls++
		return RankWorkflowRAGFragments(query, fragments, topK)
	}
	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_candidate_review", "audit_candidate_review"
	reviewed := service.CreateCandidateReview(ctx, created.Resource.DatasetID, WorkflowRAGCandidateReviewInput{
		DatasetVersion: 1, DatasetDigest: created.Version.Dataset.DatasetDigest, CandidateSnapshot: workflowRAGEvaluationSnapshotBinding(candidate),
	})
	if reviewed.FailureCode != "" || reviewed.Review == nil || reviewed.Review.Conclusion != "regressed" || rankerCalls != 2 || reviewed.Review.Candidate.Metrics.SamplePassRate != 0 {
		t.Fatalf("candidate review did not detect regression or rank exactly once per side: calls=%d result=%#v", rankerCalls, reviewed)
	}
	reviewPayload, _ := json.Marshal(reviewed.Review)
	for _, forbidden := range []string{"approved pressure relief procedure", "unrelated maintenance schedule", "The official guide is required."} {
		if strings.Contains(string(reviewPayload), forbidden) {
			t.Fatalf("candidate review leaked sensitive content %q: %s", forbidden, reviewPayload)
		}
	}

	restored := service.ReadCandidateReview(ctx, created.Resource.DatasetID, reviewed.Review.ReviewID)
	if restored.FailureCode != "" || restored.Review == nil || restored.Review.Conclusion != "regressed" {
		t.Fatalf("read candidate review: %#v", restored)
	}

	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_dataset_version", "audit_dataset_version"
	versioned := service.Version(ctx, created.Resource.DatasetID, WorkflowRAGEvaluationDatasetVersionInput{
		ExpectedLatestVersion: 1, DisplayName: "Pressure review v2", ContentClassification: "synthetic_public",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(baseline), Thresholds: workflowRAGEvaluationPerfectThresholds(),
		ReviewSummary: "Second human review.", Samples: created.Version.Dataset.Samples,
	})
	if versioned.FailureCode != "" || versioned.Version == nil || versioned.Version.Dataset.DatasetVersion != 2 {
		t.Fatalf("version evaluation dataset: %#v", versioned)
	}
	stale := service.Version(ctx, created.Resource.DatasetID, WorkflowRAGEvaluationDatasetVersionInput{ExpectedLatestVersion: 1})
	if stale.FailureCode != WorkflowRAGEvaluationFailureVersionConflict || stale.CurrentLatestVersion != 2 {
		t.Fatalf("stale dataset CAS did not fail closed: %#v", stale)
	}

	now = now.Add(time.Minute)
	ctx.RequestID, ctx.AuditRef = "request_dataset_archive", "audit_dataset_archive"
	archived := service.Archive(ctx, created.Resource.DatasetID, 2)
	if archived.FailureCode != "" || archived.Resource == nil || archived.Resource.LifecycleState != workflowRAGEvaluationArchived {
		t.Fatalf("archive evaluation dataset: %#v", archived)
	}
	if old := service.Read(ctx, created.Resource.DatasetID, 1); old.FailureCode != "" || old.Version == nil || old.Version.Dataset.DatasetVersion != 1 {
		t.Fatalf("immutable dataset v1 was not readable after archive: %#v", old)
	}
}

func TestWorkflowRAGEvaluationResourceClassificationScopeAndBinding(t *testing.T) {
	ctx := workflowRAGTestContext()
	snapshots := newMemoryWorkflowRAGSnapshotRepository(nil)
	internal := createWorkflowRAGEvaluationTestSnapshot(t, snapshots, ctx, "rags_cccccccccccccccc", "internal_knowledge", "workspace_internal", workflowRAGTestFragments())
	service := newWorkflowRAGEvaluationDatasetService(newMemoryWorkflowRAGEvaluationDatasetRepository(nil), snapshots)
	service.newID = workflowRAGEvaluationTestIDGenerator()
	input := WorkflowRAGEvaluationDatasetCreateInput{
		DatasetKey: "internal_review", DisplayName: "Internal review", ContentClassification: "workspace_internal",
		BaselineSnapshot: workflowRAGEvaluationSnapshotBinding(internal), Thresholds: workflowRAGEvaluationPerfectThresholds(), ReviewSummary: "Authorized internal query review.",
		Samples: []WorkflowRAGEvaluationSample{{SampleID: "internal_notes", QueryText: "internal operator notes", Expectation: "evidence_required", ExpectedCitationRefs: []string{"internal_notes"}, RequiredOfficialRefs: []string{}, TopK: 2, ReviewNote: "Workspace-only evidence."}},
	}
	created := service.Create(ctx, input)
	if created.FailureCode != "" {
		t.Fatalf("workspace-internal dataset rejected: %#v", created)
	}
	publicMismatch := input
	publicMismatch.DatasetKey, publicMismatch.ContentClassification = "public_mismatch", "synthetic_public"
	if result := service.Create(ctx, publicMismatch); result.FailureCode != WorkflowRAGEvaluationFailureClassificationMismatch {
		t.Fatalf("classification mismatch accepted: %#v", result)
	}
	tampered := input
	tampered.DatasetKey, tampered.BaselineSnapshot.SnapshotDigest = "binding_mismatch", workflowRAGSHA256("tampered")
	if result := service.Create(ctx, tampered); result.FailureCode != WorkflowRAGEvaluationFailureBindingInvalid {
		t.Fatalf("tampered snapshot digest accepted: %#v", result)
	}
	other := ctx
	other.ApplicationID = "app_other"
	if result := service.Read(other, created.Resource.DatasetID, 1); result.FailureCode != WorkflowRAGEvaluationFailureScopeDenied {
		t.Fatalf("cross-application dataset read accepted: %#v", result)
	}
	secret := input
	secret.DatasetKey, secret.Samples[0].QueryText = "secret_review", "Authorization: Bearer not-allowed"
	if result := service.Create(ctx, secret); result.FailureCode != WorkflowRAGEvaluationFailurePayloadInvalid {
		t.Fatalf("secret-bearing query accepted: %#v", result)
	}
}

func createWorkflowRAGEvaluationTestSnapshot(t *testing.T, repository workflowRAGSnapshotRepository, ctx WorkflowRAGSnapshotContext, snapshotID, key, classification string, fragments []WorkflowRAGFragmentInput) WorkflowRAGSnapshotRecord {
	t.Helper()
	service := newWorkflowRAGSnapshotService(repository)
	service.newID = func(prefix string) (string, error) {
		if prefix == "rags_" {
			return snapshotID, nil
		}
		return newWorkflowRAGStableID(prefix)
	}
	result := service.Create(ctx, WorkflowRAGSnapshotCreateInput{SnapshotKey: key, DisplayName: strings.ReplaceAll(key, "_", " "), ContentClassification: classification, Fragments: fragments})
	if result.FailureCode != "" || result.Record == nil {
		t.Fatalf("create evaluation test snapshot: %#v", result)
	}
	return *result.Record
}

func workflowRAGEvaluationPerfectThresholds() WorkflowRAGEvaluationThresholds {
	return WorkflowRAGEvaluationThresholds{HitAtK: 1, ExpectedRecallAtK: 1, RequiredOfficialRecallAtK: 1, MeanReciprocalRank: 1, NoEvidenceAccuracy: 1, SamplePassRate: 1}
}

func workflowRAGEvaluationTestIDGenerator() func(string) (string, error) {
	counters := map[string]int{}
	return func(prefix string) (string, error) {
		counters[prefix]++
		letter := byte('a' + counters[prefix] - 1)
		return prefix + strings.Repeat(string(letter), 16), nil
	}
}
