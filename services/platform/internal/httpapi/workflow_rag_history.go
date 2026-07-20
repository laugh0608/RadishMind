package httpapi

import "strings"

const workflowRAGHistoryPreviewCharacters = 512

type WorkflowRAGFragmentPreview struct {
	FragmentRef string `json:"fragment_ref"`
	Preview     string `json:"preview"`
	Truncated   bool   `json:"truncated"`
}

func (server *Server) workflowRAGHistoryPreviews(runContext WorkflowRunContext, record WorkflowRunRecord) ([]WorkflowRAGFragmentPreview, string) {
	if record.SchemaVersion != workflowRunRecordRAGSchemaVersion || record.RAGSnapshot == nil || record.RetrievalAttempt == nil {
		return []WorkflowRAGFragmentPreview{}, ""
	}
	snapshotContext := workflowRAGSnapshotContextFromRun(runContext, runContext.AuditRef+"-fragment-preview")
	result := server.workflowRAGSnapshotService().Read(snapshotContext, record.RAGSnapshot.SnapshotID, record.RAGSnapshot.SnapshotVersion)
	if result.FailureCode != "" || result.Record == nil {
		return []WorkflowRAGFragmentPreview{}, WorkflowRAGFailureStoreUnavailable
	}
	if result.Record.SnapshotDigest != record.RAGSnapshot.SnapshotDigest || result.Record.RAGRef != record.RAGSnapshot.RAGRef {
		return []WorkflowRAGFragmentPreview{}, WorkflowRAGFailureStoreUnavailable
	}
	fragments := make(map[string]WorkflowRAGFragment, len(result.Record.Fragments))
	for _, fragment := range result.Record.Fragments {
		fragments[fragment.FragmentRef] = fragment
	}
	previews := make([]WorkflowRAGFragmentPreview, 0, len(record.RetrievalAttempt.SelectedFragments))
	for _, selected := range record.RetrievalAttempt.SelectedFragments {
		fragment, ok := fragments[selected.FragmentRef]
		if !ok || fragment.ContentDigest != selected.ContentDigest {
			return []WorkflowRAGFragmentPreview{}, WorkflowRAGFailureStoreUnavailable
		}
		preview, truncated := workflowRAGHistoryPreview(fragment.Content)
		previews = append(previews, WorkflowRAGFragmentPreview{FragmentRef: selected.FragmentRef, Preview: preview, Truncated: truncated})
	}
	return previews, ""
}

func workflowRAGHistoryPreview(content string) (string, bool) {
	content = strings.TrimSpace(content)
	runes := []rune(content)
	if len(runes) <= workflowRAGHistoryPreviewCharacters {
		return content, false
	}
	return string(runes[:workflowRAGHistoryPreviewCharacters]), true
}
