package httpapi

import (
	"strings"
	"testing"
)

func TestWorkflowRAGLexicalRankingIsDeterministic(t *testing.T) {
	fragments := workflowRAGLexicalTestFragments(t, []WorkflowRAGFragmentInput{
		{FragmentRef: "community_guide", SourceType: "forum", SourceRef: "forum.rag", PageSlug: "rag/community", Title: "Retrieval guide", Content: "retrieval ranking guidance"},
		{FragmentRef: "official_guide", SourceType: "manual", SourceRef: "manual.rag", PageSlug: "rag/official", Title: "Retrieval guide", IsOfficial: true, Content: "retrieval ranking guidance"},
		{FragmentRef: "chinese_guide", SourceType: "wiki", SourceRef: "wiki.rag", PageSlug: "rag/chinese", Title: "检索说明", Content: "知识检索采用稳定排序"},
	})
	latin := RankWorkflowRAGFragments("retrieval ranking", fragments, 2)
	if latin.FailureCode != "" || len(latin.Selected) != 2 || latin.Selected[0].FragmentRef != "official_guide" ||
		latin.Selected[0].Rank != 1 || latin.Profile.NetworkAccess || latin.Profile.EmbeddingProvider != "none" ||
		!workflowRAGDigestPattern.MatchString(latin.Profile.ProfileDigest) {
		t.Fatalf("Latin lexical ranking drifted: %#v", latin)
	}
	cjk := RankWorkflowRAGFragments("知识检索", fragments, 1)
	if cjk.FailureCode != "" || len(cjk.Selected) != 1 || cjk.Selected[0].FragmentRef != "chinese_guide" {
		t.Fatalf("CJK bigram ranking drifted: %#v", cjk)
	}
}

func TestWorkflowRAGLexicalRankingTieBreakAndBudgets(t *testing.T) {
	fragments := workflowRAGLexicalTestFragments(t, []WorkflowRAGFragmentInput{
		{FragmentRef: "alpha_fragment", SourceType: "faq", SourceRef: "faq.alpha", PageSlug: "faq/alpha", Content: "same evidence"},
		{FragmentRef: "beta_fragment", SourceType: "faq", SourceRef: "faq.beta", PageSlug: "faq/beta", Content: "same evidence"},
	})
	result := RankWorkflowRAGFragments("same evidence", fragments, 2)
	if result.FailureCode != "" || len(result.Selected) != 2 || result.Selected[0].FragmentRef != "alpha_fragment" || result.Selected[1].FragmentRef != "beta_fragment" {
		t.Fatalf("equal lexical score did not use fragment_ref tie-break: %#v", result)
	}
	if empty := RankWorkflowRAGFragments("   ", fragments, 1); empty.FailureCode != WorkflowRAGFailureQueryInvalid || len(empty.Selected) != 0 {
		t.Fatalf("empty query did not fail closed: %#v", empty)
	}
	if excessiveTopK := RankWorkflowRAGFragments("evidence", fragments, workflowRAGMaxTopK+1); excessiveTopK.FailureCode != WorkflowRAGFailureBudgetExceeded {
		t.Fatalf("excessive top_k did not fail closed: %#v", excessiveTopK)
	}

	long := workflowRAGLexicalTestFragments(t, []WorkflowRAGFragmentInput{{
		FragmentRef: "long_fragment", SourceType: "manual", SourceRef: "manual.long", PageSlug: "manual/long",
		Content: "needle " + strings.Repeat("x", workflowRAGMaxExcerptBytes+100),
	}})
	bounded := RankWorkflowRAGFragments("needle", long, 1)
	if bounded.FailureCode != "" || len(bounded.Selected) != 1 || !bounded.Selected[0].ExcerptTruncated ||
		bounded.Selected[0].ExcerptBytes > workflowRAGMaxExcerptBytes || bounded.ContextBytes > workflowRAGMaxContextBytes {
		t.Fatalf("lexical excerpt budget drifted: %#v", bounded)
	}
}

func workflowRAGLexicalTestFragments(t *testing.T, inputs []WorkflowRAGFragmentInput) []WorkflowRAGFragment {
	t.Helper()
	fragments := make([]WorkflowRAGFragment, 0, len(inputs))
	for _, input := range inputs {
		fragment, failure := normalizeWorkflowRAGFragment(input, "workspace_internal")
		if failure != "" {
			t.Fatalf("normalize lexical test fragment %s: %s", input.FragmentRef, failure)
		}
		fragments = append(fragments, fragment)
	}
	return fragments
}
