package httpapi

import (
	"encoding/json"
	"math"
	"strings"
	"unicode/utf8"
)

type WorkflowRAGExecutionRequest struct {
	DraftID      string
	DraftVersion int
	InputText    string
	Model        string
	Temperature  *float64
}

type workflowRAGExecutionPlan struct {
	promptNode    SavedWorkflowDraftNode
	retrievalNode SavedWorkflowDraftNode
	llmNode       SavedWorkflowDraftNode
	outputNode    SavedWorkflowDraftNode
	ragRef        string
}

func normalizeWorkflowRAGExecutionRequest(request WorkflowRAGExecutionRequest) (WorkflowRAGExecutionRequest, string) {
	normalized := WorkflowRAGExecutionRequest{
		DraftID: strings.TrimSpace(request.DraftID), DraftVersion: request.DraftVersion,
		InputText: strings.TrimSpace(request.InputText), Model: strings.TrimSpace(request.Model),
		Temperature: request.Temperature,
	}
	if !workflowRAGScopedIDPattern.MatchString(normalized.DraftID) || normalized.DraftVersion < 1 {
		return WorkflowRAGExecutionRequest{}, WorkflowRAGFailureDraftIneligible
	}
	if normalized.InputText == "" || !utf8.ValidString(normalized.InputText) || len([]byte(normalized.InputText)) > workflowRAGMaxQueryBytes {
		return WorkflowRAGExecutionRequest{}, WorkflowRAGFailureQueryInvalid
	}
	if len([]rune(normalized.Model)) > workflowExecutorMaxModelChars || strings.Contains(normalized.Model, "://") || workflowRAGContainsForbiddenMaterial(normalized.Model) {
		return WorkflowRAGExecutionRequest{}, WorkflowRAGFailureDraftIneligible
	}
	if normalized.Temperature != nil {
		value := *normalized.Temperature
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 2 {
			return WorkflowRAGExecutionRequest{}, WorkflowRAGFailureDraftIneligible
		}
	}
	return normalized, ""
}

func buildWorkflowRAGExecutionPlan(draft SavedWorkflowDraft, expectedVersion int) (workflowRAGExecutionPlan, string) {
	if draft.SchemaVersion != savedWorkflowDraftSchemaVersion || draft.DraftVersion != expectedVersion ||
		draft.DraftStatus != SavedWorkflowDraftStatusValidForReview || !draft.ValidationSummary.ValidForReview ||
		draft.ValidationSummary.ValidationState != SavedWorkflowDraftStatusValidForReview ||
		len(draft.BlockedCapabilitySummary) != 0 || len(draft.Nodes) != 4 || len(draft.Edges) != 3 ||
		len(draft.RAGRefs) != 1 || len(draft.ToolRefs) != 0 {
		return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
	}
	plan := workflowRAGExecutionPlan{ragRef: strings.TrimSpace(draft.RAGRefs[0])}
	if !workflowRAGRAGRefPattern.MatchString(plan.ragRef) {
		return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
	}
	nodes := make(map[string]SavedWorkflowDraftNode, 4)
	counts := make(map[string]int, 4)
	for _, rawNode := range draft.Nodes {
		node := rawNode
		node.NodeID = strings.TrimSpace(node.NodeID)
		node.NodeType = strings.ToLower(strings.TrimSpace(node.NodeType))
		if !workflowRAGScopedIDPattern.MatchString(node.NodeID) || nodes[node.NodeID].NodeID != "" ||
			strings.TrimSpace(node.ToolRef) != "" || node.RequiresConfirmation {
			return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
		}
		counts[node.NodeType]++
		switch node.NodeType {
		case "prompt":
			if strings.TrimSpace(node.RAGRef) != "" {
				return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
			}
			plan.promptNode = node
		case "rag_retrieval":
			if strings.TrimSpace(node.RAGRef) != plan.ragRef || strings.ToLower(strings.TrimSpace(node.RiskLevel)) != "low" {
				return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
			}
			plan.retrievalNode = node
		case "llm":
			if strings.TrimSpace(node.RAGRef) != "" {
				return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
			}
			plan.llmNode = node
		case "output":
			if strings.TrimSpace(node.RAGRef) != "" {
				return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
			}
			plan.outputNode = node
		default:
			return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
		}
		nodes[node.NodeID] = node
	}
	for _, nodeType := range []string{"prompt", "rag_retrieval", "llm", "output"} {
		if counts[nodeType] != 1 {
			return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
		}
	}
	expectedEdges := map[string]bool{
		plan.promptNode.NodeID + "\x00" + plan.retrievalNode.NodeID: true,
		plan.retrievalNode.NodeID + "\x00" + plan.llmNode.NodeID:    true,
		plan.llmNode.NodeID + "\x00" + plan.outputNode.NodeID:       true,
	}
	seenEdges := make(map[string]bool, 3)
	for _, edge := range draft.Edges {
		key := strings.TrimSpace(edge.FromNodeID) + "\x00" + strings.TrimSpace(edge.ToNodeID)
		if strings.TrimSpace(edge.EdgeID) == "" || strings.TrimSpace(edge.ConditionSummary) != "" || !expectedEdges[key] || seenEdges[key] {
			return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
		}
		seenEdges[key] = true
	}
	if len(seenEdges) != 3 {
		return workflowRAGExecutionPlan{}, WorkflowRAGFailureDraftIneligible
	}
	return plan, ""
}

func workflowRAGDraftDigest(draft SavedWorkflowDraft) (string, error) {
	immutable := struct {
		DraftVersion int                       `json:"draft_version"`
		Payload      SavedWorkflowDraftPayload `json:"payload"`
	}{
		DraftVersion: draft.DraftVersion,
		Payload: SavedWorkflowDraftPayload{
			DraftID: draft.DraftID, WorkspaceID: draft.WorkspaceID, ApplicationID: draft.ApplicationID,
			SourceDefinitionID: draft.SourceDefinitionID, BaseDefinitionVersion: draft.BaseDefinitionVersion,
			SchemaVersion: draft.SchemaVersion, DraftStatus: draft.DraftStatus, Name: draft.Name,
			Description: draft.Description, Nodes: draft.Nodes, Edges: draft.Edges,
			InputContract: draft.InputContract, OutputContract: draft.OutputContract,
			ProviderRefs: draft.ProviderRefs, ToolRefs: draft.ToolRefs, RAGRefs: draft.RAGRefs,
			RequestedCapabilities: draft.RequestedCapabilities, AdditionalFields: draft.AdditionalFields,
		},
	}
	payload, err := json.Marshal(immutable)
	if err != nil {
		return "", err
	}
	return workflowRAGSHA256(string(payload)), nil
}

func parseWorkflowRAGAnswer(payload string, selected []WorkflowRAGRankedFragment) (WorkflowRAGAnswer, string) {
	var answer WorkflowRAGAnswer
	if err := decodeWorkflowRAGStrictJSON([]byte(strings.TrimSpace(payload)), &answer); err != nil {
		return WorkflowRAGAnswer{}, WorkflowRAGFailureAnswerInvalid
	}
	if answer.SchemaVersion != "workflow_rag_answer.v1" || strings.TrimSpace(answer.Answer) == "" ||
		len([]rune(answer.Answer)) > 16384 || len(answer.Citations) < 1 || len(answer.Citations) > workflowRAGMaxTopK ||
		len(answer.Limitations) > workflowRAGMaxTopK ||
		(answer.Confidence != "low" && answer.Confidence != "medium" && answer.Confidence != "high") ||
		workflowRAGContainsForbiddenMaterial(answer.Answer) {
		return WorkflowRAGAnswer{}, WorkflowRAGFailureAnswerInvalid
	}
	for _, limitation := range answer.Limitations {
		if strings.TrimSpace(limitation) == "" || len([]rune(limitation)) > 512 || workflowRAGContainsForbiddenMaterial(limitation) {
			return WorkflowRAGAnswer{}, WorkflowRAGFailureAnswerInvalid
		}
	}
	selectedRefs := make(map[string]bool, len(selected))
	for _, fragment := range selected {
		selectedRefs[fragment.FragmentRef] = true
	}
	seen := make(map[string]bool, len(answer.Citations))
	for _, citation := range answer.Citations {
		if !workflowRAGFragmentRefPattern.MatchString(citation.FragmentRef) || !selectedRefs[citation.FragmentRef] || seen[citation.FragmentRef] {
			return WorkflowRAGAnswer{}, WorkflowRAGFailureCitationInvalid
		}
		if strings.TrimSpace(citation.ClaimSummary) == "" || len([]rune(citation.ClaimSummary)) > 512 || workflowRAGContainsForbiddenMaterial(citation.ClaimSummary) {
			return WorkflowRAGAnswer{}, WorkflowRAGFailureAnswerInvalid
		}
		seen[citation.FragmentRef] = true
	}
	return answer, ""
}
