package httpapi

import (
	"sort"
	"time"
)

const (
	workflowRAGComparisonProfile            = "workflow_rag_retrieval.v1"
	workflowRAGApplicationComparisonProfile = "workflow_rag_application_invocation.v1"
)

type WorkflowRunRetrievalFragmentComparison struct {
	FragmentRef   string `json:"fragment_ref"`
	ContentDigest string `json:"content_digest"`
	SourceType    string `json:"source_type"`
	IsOfficial    bool   `json:"is_official"`
	BaselineRank  int    `json:"baseline_rank"`
	CandidateRank int    `json:"candidate_rank"`
	Change        string `json:"change"`
}

type WorkflowRunApplicationRAGAuthorityComparison struct {
	AssignmentID         string                           `json:"assignment_id"`
	AssignmentVersion    int                              `json:"assignment_version"`
	AssignmentDigest     string                           `json:"assignment_digest"`
	PublishCandidateID   string                           `json:"publish_candidate_id"`
	PublishReviewVersion int                              `json:"publish_review_version"`
	DraftID              string                           `json:"draft_id"`
	DraftVersion         int                              `json:"draft_version"`
	DraftDigest          string                           `json:"draft_digest"`
	BindingRef           WorkflowRAGApplicationBindingRef `json:"binding_ref"`
	SnapshotID           string                           `json:"snapshot_id"`
	SnapshotVersion      int                              `json:"snapshot_version"`
	SnapshotDigest       string                           `json:"snapshot_digest"`
	RAGRef               string                           `json:"rag_ref"`
	ProfileID            string                           `json:"profile_id"`
	ProfileVersion       int                              `json:"profile_version"`
	ProfileDigest        string                           `json:"profile_digest"`
	ConfiguredProtocol   string                           `json:"configured_protocol"`
	ConfiguredModel      string                           `json:"configured_model"`
}

type WorkflowRunRetrievalComparison struct {
	RunProfile              string                                        `json:"run_profile"`
	SnapshotID              string                                        `json:"snapshot_id,omitempty"`
	SnapshotVersion         int                                           `json:"snapshot_version,omitempty"`
	SnapshotDigest          string                                        `json:"snapshot_digest,omitempty"`
	RAGRef                  string                                        `json:"rag_ref,omitempty"`
	ProfileID               string                                        `json:"profile_id,omitempty"`
	ProfileVersion          int                                           `json:"profile_version,omitempty"`
	ProfileDigest           string                                        `json:"profile_digest,omitempty"`
	BaselineAuthority       *WorkflowRunApplicationRAGAuthorityComparison `json:"baseline_authority,omitempty"`
	CandidateAuthority      *WorkflowRunApplicationRAGAuthorityComparison `json:"candidate_authority,omitempty"`
	AuthorityChanged        bool                                          `json:"-"`
	AuthorityChangedValue   *bool                                         `json:"authority_changed,omitempty"`
	QueryDigest             string                                        `json:"query_digest"`
	QueryBytes              int                                           `json:"query_bytes"`
	RetrievalNodeID         string                                        `json:"retrieval_node_id"`
	BaselineAttemptStatus   string                                        `json:"baseline_attempt_status"`
	CandidateAttemptStatus  string                                        `json:"candidate_attempt_status"`
	BaselineCandidateCount  int                                           `json:"baseline_candidate_count"`
	CandidateCandidateCount int                                           `json:"candidate_candidate_count"`
	CandidateCountDelta     int                                           `json:"candidate_count_delta"`
	BaselineSelectedCount   int                                           `json:"baseline_selected_count"`
	CandidateSelectedCount  int                                           `json:"candidate_selected_count"`
	BaselineCitationCount   int                                           `json:"baseline_citation_count"`
	CandidateCitationCount  int                                           `json:"candidate_citation_count"`
	ContextBytesDelta       int                                           `json:"context_bytes_delta"`
	LatencyDeltaMS          int                                           `json:"latency_delta_ms"`
	EvidenceChanged         bool                                          `json:"evidence_changed"`
	RankingChanged          bool                                          `json:"ranking_changed"`
	CitationChanged         bool                                          `json:"citation_changed"`
	CitationAddedRefs       []string                                      `json:"citation_added_refs"`
	CitationRemovedRefs     []string                                      `json:"citation_removed_refs"`
	Fragments               []WorkflowRunRetrievalFragmentComparison      `json:"fragments"`
}

func workflowRAGRunsComparable(baseline, candidate WorkflowRunRecord) bool {
	if !workflowRunRecordUsesRetrievalComparison(baseline) || baseline.SchemaVersion != candidate.SchemaVersion ||
		baseline.RAGSnapshot == nil || candidate.RAGSnapshot == nil || baseline.RetrievalAttempt == nil || candidate.RetrievalAttempt == nil {
		return false
	}
	if baseline.SchemaVersion == workflowRunRecordAppRAGSchemaVersion {
		return workflowRAGApplicationRunsComparable(baseline, candidate)
	}
	leftSnapshot, rightSnapshot := baseline.RAGSnapshot, candidate.RAGSnapshot
	leftAttempt, rightAttempt := baseline.RetrievalAttempt, candidate.RetrievalAttempt
	return leftSnapshot.SnapshotID == rightSnapshot.SnapshotID &&
		leftSnapshot.SnapshotVersion == rightSnapshot.SnapshotVersion &&
		leftSnapshot.SnapshotDigest == rightSnapshot.SnapshotDigest &&
		leftSnapshot.RAGRef == rightSnapshot.RAGRef &&
		leftAttempt.ProfileID == rightAttempt.ProfileID &&
		leftAttempt.ProfileVersion == rightAttempt.ProfileVersion &&
		leftAttempt.ProfileDigest == rightAttempt.ProfileDigest &&
		leftAttempt.QueryDigest == rightAttempt.QueryDigest &&
		leftAttempt.QueryBytes == rightAttempt.QueryBytes &&
		leftAttempt.NodeID == rightAttempt.NodeID
}

func workflowRAGApplicationRunsComparable(baseline, candidate WorkflowRunRecord) bool {
	if baseline.RAGApplication == nil || candidate.RAGApplication == nil || baseline.ExecutionSource == nil || candidate.ExecutionSource == nil {
		return false
	}
	leftAttempt, rightAttempt := baseline.RetrievalAttempt, candidate.RetrievalAttempt
	return baseline.ApplicationID == candidate.ApplicationID &&
		leftAttempt.QueryDigest == rightAttempt.QueryDigest &&
		leftAttempt.QueryBytes == rightAttempt.QueryBytes &&
		leftAttempt.NodeID == rightAttempt.NodeID &&
		leftAttempt.ProfileID == rightAttempt.ProfileID &&
		leftAttempt.ProfileVersion == rightAttempt.ProfileVersion &&
		leftAttempt.ProfileDigest == rightAttempt.ProfileDigest &&
		baseline.RAGApplication.ConfiguredProtocol == candidate.RAGApplication.ConfiguredProtocol
}

func buildWorkflowRAGRunComparison(baseline, candidate WorkflowRunRecord, now time.Time) WorkflowRunComparison {
	comparison := buildWorkflowRunComparison(baseline, candidate, now)
	comparison.SchemaVersion = workflowRAGRunComparisonSchemaVersion
	retrieval := compareWorkflowRAGRetrievalAttempts(*baseline.RAGSnapshot, *baseline.RetrievalAttempt, *candidate.RetrievalAttempt)
	if baseline.SchemaVersion == workflowRunRecordAppRAGSchemaVersion {
		comparison.SchemaVersion = workflowRAGAppRunComparisonSchemaVersion
		retrieval.RunProfile = workflowRAGApplicationComparisonProfile
		baselineAuthority := workflowRAGApplicationComparisonAuthority(*baseline.RAGApplication, *baseline.RAGSnapshot)
		candidateAuthority := workflowRAGApplicationComparisonAuthority(*candidate.RAGApplication, *candidate.RAGSnapshot)
		retrieval.SnapshotID, retrieval.SnapshotVersion, retrieval.SnapshotDigest, retrieval.RAGRef = "", 0, "", ""
		retrieval.ProfileID, retrieval.ProfileVersion, retrieval.ProfileDigest = "", 0, ""
		retrieval.BaselineAuthority = &baselineAuthority
		retrieval.CandidateAuthority = &candidateAuthority
		retrieval.AuthorityChanged = baselineAuthority != candidateAuthority
		retrieval.AuthorityChangedValue = &retrieval.AuthorityChanged
	}
	comparison.Retrieval = &retrieval
	comparison.Classification = classifyWorkflowRunComparison(comparison)
	comparison.Findings = workflowRunComparisonFindings(comparison)
	comparison.RecommendedReviewAction = comparisonReviewAction(comparison, candidate)
	return comparison
}

func workflowRAGApplicationComparisonAuthority(
	authority workflowRAGApplicationRunAuthority,
	snapshot workflowRAGRunSnapshotBinding,
) WorkflowRunApplicationRAGAuthorityComparison {
	return WorkflowRunApplicationRAGAuthorityComparison{
		AssignmentID: authority.AssignmentID, AssignmentVersion: authority.AssignmentVersion, AssignmentDigest: authority.AssignmentDigest,
		PublishCandidateID: authority.PublishCandidateID, PublishReviewVersion: authority.PublishReviewVersion,
		DraftID: authority.DraftID, DraftVersion: authority.DraftVersion, DraftDigest: authority.DraftDigest,
		BindingRef: authority.BindingRef,
		SnapshotID: snapshot.SnapshotID, SnapshotVersion: snapshot.SnapshotVersion, SnapshotDigest: snapshot.SnapshotDigest, RAGRef: snapshot.RAGRef,
		ProfileID: authority.Profile.ProfileID, ProfileVersion: authority.Profile.ProfileVersion, ProfileDigest: authority.Profile.ProfileDigest,
		ConfiguredProtocol: authority.ConfiguredProtocol, ConfiguredModel: authority.ConfiguredModel,
	}
}

func compareWorkflowRAGRetrievalAttempts(
	snapshot workflowRAGRunSnapshotBinding,
	baseline workflowRAGRunRetrievalAttempt,
	candidate workflowRAGRunRetrievalAttempt,
) WorkflowRunRetrievalComparison {
	baselineFragments := make(map[string]workflowRAGRunSelectedFragment, len(baseline.SelectedFragments))
	candidateFragments := make(map[string]workflowRAGRunSelectedFragment, len(candidate.SelectedFragments))
	refs := make([]string, 0, len(baseline.SelectedFragments)+len(candidate.SelectedFragments))
	for _, fragment := range baseline.SelectedFragments {
		baselineFragments[fragment.FragmentRef] = fragment
		refs = append(refs, fragment.FragmentRef)
	}
	for _, fragment := range candidate.SelectedFragments {
		candidateFragments[fragment.FragmentRef] = fragment
		if _, exists := baselineFragments[fragment.FragmentRef]; !exists {
			refs = append(refs, fragment.FragmentRef)
		}
	}
	fragments := make([]WorkflowRunRetrievalFragmentComparison, 0, len(refs))
	evidenceChanged, rankingChanged := false, false
	for _, ref := range refs {
		left, hasLeft := baselineFragments[ref]
		right, hasRight := candidateFragments[ref]
		change := "unchanged"
		switch {
		case !hasLeft:
			change, evidenceChanged = "added", true
		case !hasRight:
			change, evidenceChanged = "removed", true
		case left.ContentDigest != right.ContentDigest || left.SourceType != right.SourceType || left.IsOfficial != right.IsOfficial:
			change, evidenceChanged = "changed", true
		case left.Rank != right.Rank:
			change, rankingChanged = "moved", true
		}
		metadata := right
		if !hasRight {
			metadata = left
		}
		fragments = append(fragments, WorkflowRunRetrievalFragmentComparison{
			FragmentRef: ref, ContentDigest: metadata.ContentDigest, SourceType: metadata.SourceType,
			IsOfficial: metadata.IsOfficial, BaselineRank: left.Rank, CandidateRank: right.Rank, Change: change,
		})
	}
	addedCitations, removedCitations := compareWorkflowRAGCitationRefs(baseline.CitationRefs, candidate.CitationRefs)
	return WorkflowRunRetrievalComparison{
		RunProfile: workflowRAGComparisonProfile,
		SnapshotID: snapshot.SnapshotID, SnapshotVersion: snapshot.SnapshotVersion,
		SnapshotDigest: snapshot.SnapshotDigest, RAGRef: snapshot.RAGRef,
		ProfileID: baseline.ProfileID, ProfileVersion: baseline.ProfileVersion, ProfileDigest: baseline.ProfileDigest,
		QueryDigest: baseline.QueryDigest, QueryBytes: baseline.QueryBytes, RetrievalNodeID: baseline.NodeID,
		BaselineAttemptStatus: baseline.Status, CandidateAttemptStatus: candidate.Status,
		BaselineCandidateCount: baseline.CandidateCount, CandidateCandidateCount: candidate.CandidateCount,
		CandidateCountDelta:   candidate.CandidateCount - baseline.CandidateCount,
		BaselineSelectedCount: len(baseline.SelectedFragments), CandidateSelectedCount: len(candidate.SelectedFragments),
		BaselineCitationCount: len(baseline.CitationRefs), CandidateCitationCount: len(candidate.CitationRefs),
		ContextBytesDelta: candidate.ContextBytes - baseline.ContextBytes,
		LatencyDeltaMS:    candidate.RetrievalLatencyMS - baseline.RetrievalLatencyMS,
		EvidenceChanged:   evidenceChanged, RankingChanged: rankingChanged,
		CitationChanged:   len(addedCitations)+len(removedCitations) > 0,
		CitationAddedRefs: addedCitations, CitationRemovedRefs: removedCitations, Fragments: fragments,
	}
}

func compareWorkflowRAGCitationRefs(baseline, candidate []string) ([]string, []string) {
	left, right := make(map[string]bool, len(baseline)), make(map[string]bool, len(candidate))
	for _, ref := range baseline {
		left[ref] = true
	}
	for _, ref := range candidate {
		right[ref] = true
	}
	added, removed := make([]string, 0), make([]string, 0)
	for ref := range right {
		if !left[ref] {
			added = append(added, ref)
		}
	}
	for ref := range left {
		if !right[ref] {
			removed = append(removed, ref)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func workflowRAGComparisonMateriallyChanged(retrieval *WorkflowRunRetrievalComparison) bool {
	if retrieval == nil {
		return false
	}
	return retrieval.BaselineAttemptStatus != retrieval.CandidateAttemptStatus ||
		retrieval.CandidateCountDelta != 0 || retrieval.ContextBytesDelta != 0 ||
		retrieval.EvidenceChanged || retrieval.RankingChanged || retrieval.CitationChanged || retrieval.AuthorityChanged
}
