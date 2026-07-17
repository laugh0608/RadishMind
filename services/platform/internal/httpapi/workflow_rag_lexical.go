package httpapi

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type WorkflowRAGRankedFragment struct {
	FragmentRef      string  `json:"fragment_ref"`
	Rank             int     `json:"rank"`
	Score            float64 `json:"score"`
	SourceType       string  `json:"source_type"`
	IsOfficial       bool    `json:"is_official"`
	ContentDigest    string  `json:"content_digest"`
	Excerpt          string  `json:"excerpt"`
	ExcerptBytes     int     `json:"excerpt_bytes"`
	ExcerptTruncated bool    `json:"excerpt_truncated"`
}

type WorkflowRAGRankingResult struct {
	Profile        WorkflowRAGExecutionProfile `json:"profile"`
	QueryDigest    string                      `json:"query_digest"`
	QueryBytes     int                         `json:"query_bytes"`
	CandidateCount int                         `json:"candidate_count"`
	Selected       []WorkflowRAGRankedFragment `json:"selected"`
	ContextBytes   int                         `json:"context_bytes"`
	FailureCode    string                      `json:"failure_code,omitempty"`
}

func workflowRAGLexicalProfile() WorkflowRAGExecutionProfile {
	profile := WorkflowRAGExecutionProfile{
		SchemaVersion: "workflow_rag_execution_profile.v1", ProfileID: workflowRAGProfileID,
		ProfileVersion: workflowRAGProfileVersion, Algorithm: "bm25_like.v1",
		NormalizationPolicy: workflowRAGNormalizationPolicy, StopwordPolicy: workflowRAGStopwordPolicy,
		DefaultTopK: workflowRAGDefaultTopK, MaxTopK: workflowRAGMaxTopK,
		MaxQueryBytes: workflowRAGMaxQueryBytes, MaxFragmentBytes: workflowRAGMaxFragmentBytes,
		MaxExcerptBytes: workflowRAGMaxExcerptBytes, MaxContextBytes: workflowRAGMaxContextBytes,
		OfficialScoreBoost: workflowRAGOfficialScoreBoost, NetworkAccess: false, EmbeddingProvider: "none",
	}
	canonical := profile
	canonical.ProfileDigest = ""
	payload, _ := json.Marshal(canonical)
	profile.ProfileDigest = workflowRAGSHA256(string(payload))
	return profile
}

func RankWorkflowRAGFragments(query string, fragments []WorkflowRAGFragment, topK int) WorkflowRAGRankingResult {
	profile := workflowRAGLexicalProfile()
	query = strings.TrimSpace(query)
	result := WorkflowRAGRankingResult{
		Profile: profile, QueryDigest: workflowRAGSHA256(query), QueryBytes: len([]byte(query)),
		CandidateCount: len(fragments), Selected: []WorkflowRAGRankedFragment{},
	}
	if query == "" || !utf8.ValidString(query) || result.QueryBytes > workflowRAGMaxQueryBytes {
		result.FailureCode = WorkflowRAGFailureQueryInvalid
		return result
	}
	if topK == 0 {
		topK = workflowRAGDefaultTopK
	}
	if topK < 1 || topK > workflowRAGMaxTopK || len(fragments) > workflowRAGMaxFragments {
		result.FailureCode = WorkflowRAGFailureBudgetExceeded
		return result
	}
	queryTokens := workflowRAGLexicalTokens(query)
	if len(queryTokens) == 0 {
		result.FailureCode = WorkflowRAGFailureQueryInvalid
		return result
	}
	type document struct {
		fragment WorkflowRAGFragment
		tokens   []string
		counts   map[string]int
		score    float64
	}
	documents := make([]document, 0, len(fragments))
	documentFrequency := make(map[string]int)
	totalLength := 0
	for _, fragment := range fragments {
		if fragment.ContentBytes > workflowRAGMaxFragmentBytes || fragment.ContentDigest != workflowRAGSHA256(fragment.Content) {
			result.FailureCode = WorkflowRAGFailureFragmentInvalid
			return result
		}
		tokens := workflowRAGLexicalTokens(fragment.Title + "\n" + fragment.Content)
		counts := make(map[string]int)
		for _, token := range tokens {
			counts[token]++
		}
		for token := range counts {
			documentFrequency[token]++
		}
		totalLength += len(tokens)
		documents = append(documents, document{fragment: fragment, tokens: tokens, counts: counts})
	}
	if len(documents) == 0 {
		result.FailureCode = WorkflowRAGFailureNoEvidence
		return result
	}
	averageLength := float64(totalLength) / float64(len(documents))
	if averageLength < 1 {
		averageLength = 1
	}
	queryCounts := make(map[string]int)
	for _, token := range queryTokens {
		queryCounts[token]++
	}
	for index := range documents {
		documentLength := float64(len(documents[index].tokens))
		for token, queryFrequency := range queryCounts {
			termFrequency := documents[index].counts[token]
			if termFrequency == 0 {
				continue
			}
			df := float64(documentFrequency[token])
			idf := math.Log(1 + (float64(len(documents))-df+0.5)/(df+0.5))
			frequency := float64(termFrequency)
			denominator := frequency + workflowRAGLexicalK1*(1-workflowRAGLexicalB+workflowRAGLexicalB*documentLength/averageLength)
			documents[index].score += float64(queryFrequency) * idf * (frequency * (workflowRAGLexicalK1 + 1) / denominator)
		}
		if documents[index].score > 0 && documents[index].fragment.IsOfficial {
			documents[index].score *= workflowRAGOfficialScoreBoost
		}
	}
	sort.Slice(documents, func(i, j int) bool {
		if math.Abs(documents[i].score-documents[j].score) < 1e-12 {
			return documents[i].fragment.FragmentRef < documents[j].fragment.FragmentRef
		}
		return documents[i].score > documents[j].score
	})
	for _, document := range documents {
		if document.score <= 0 || len(result.Selected) >= topK {
			break
		}
		excerpt, truncated := workflowRAGBoundedUTF8(document.fragment.Content, workflowRAGMaxExcerptBytes)
		if result.ContextBytes+len([]byte(excerpt)) > workflowRAGMaxContextBytes {
			break
		}
		result.ContextBytes += len([]byte(excerpt))
		result.Selected = append(result.Selected, WorkflowRAGRankedFragment{
			FragmentRef: document.fragment.FragmentRef, Rank: len(result.Selected) + 1,
			Score: document.score, SourceType: document.fragment.SourceType, IsOfficial: document.fragment.IsOfficial,
			ContentDigest: document.fragment.ContentDigest, Excerpt: excerpt, ExcerptBytes: len([]byte(excerpt)), ExcerptTruncated: truncated,
		})
	}
	if len(result.Selected) == 0 {
		result.FailureCode = WorkflowRAGFailureNoEvidence
	}
	return result
}

func workflowRAGLexicalTokens(value string) []string {
	runes := []rune(strings.ToLower(value))
	tokens := make([]string, 0)
	latin := make([]rune, 0)
	han := make([]rune, 0)
	for _, current := range runes {
		switch {
		case unicode.Is(unicode.Han, current):
			tokens = workflowRAGAppendLatinToken(tokens, latin)
			latin = latin[:0]
			han = append(han, current)
		case unicode.IsLetter(current) || unicode.IsDigit(current):
			tokens = workflowRAGAppendHanTokens(tokens, han)
			han = han[:0]
			latin = append(latin, current)
		default:
			tokens = workflowRAGAppendLatinToken(tokens, latin)
			tokens = workflowRAGAppendHanTokens(tokens, han)
			latin = latin[:0]
			han = han[:0]
		}
	}
	tokens = workflowRAGAppendLatinToken(tokens, latin)
	tokens = workflowRAGAppendHanTokens(tokens, han)
	return tokens
}

func workflowRAGAppendLatinToken(tokens []string, runes []rune) []string {
	if len(runes) == 0 {
		return tokens
	}
	return append(tokens, string(runes))
}

func workflowRAGAppendHanTokens(tokens []string, runes []rune) []string {
	if len(runes) == 1 {
		return append(tokens, string(runes))
	}
	for index := 0; index+1 < len(runes); index++ {
		tokens = append(tokens, string(runes[index:index+2]))
	}
	return tokens
}

func workflowRAGBoundedUTF8(value string, maximumBytes int) (string, bool) {
	if len([]byte(value)) <= maximumBytes {
		return value, false
	}
	bytes := []byte(value)
	end := maximumBytes
	for end > 0 && !utf8.Valid(bytes[:end]) {
		end--
	}
	return string(bytes[:end]), true
}
