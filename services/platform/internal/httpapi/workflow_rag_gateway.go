package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	workflowRAGExecutionProtocol = "workflow-rag-retrieval-v1"
	workflowRAGExecutionRoute    = "/v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions"
)

type workflowRAGGatewayEvidence struct {
	FragmentRef      string `json:"fragment_ref"`
	Rank             int    `json:"rank"`
	SourceType       string `json:"source_type"`
	IsOfficial       bool   `json:"is_official"`
	Excerpt          string `json:"excerpt"`
	ExcerptTruncated bool   `json:"excerpt_truncated"`
}

func (service workflowRAGExecutionService) callGateway(
	ctx context.Context,
	request WorkflowRAGExecutionRequest,
	draft SavedWorkflowDraft,
	plan workflowRAGExecutionPlan,
	runID string,
	selection northboundSelection,
	ranking WorkflowRAGRankingResult,
) (string, WorkflowRunGatewayFailureCategory, string) {
	packet, err := buildWorkflowRAGGatewayPacket(request.InputText, plan.promptNode, ranking.Selected)
	if err != nil {
		return "", WorkflowRunGatewayFailureProtocol, WorkflowRAGFailureGatewayFailed
	}
	canonicalRequest, err := buildNorthboundCanonicalRequest(northboundCanonicalRequestOptions{
		requestID: runID + "-" + plan.llmNode.NodeID,
		route:     workflowRAGExecutionRoute, protocol: workflowRAGExecutionProtocol, locale: "zh-CN", promptText: packet,
		northboundFields: map[string]any{
			"request_kind": workflowRAGExecutionProtocol, "workflow_run_id": runID,
			"workflow_draft_id": draft.DraftID, "workflow_draft_version": draft.DraftVersion,
			"workflow_node_id": plan.llmNode.NodeID, "allow_tool_calls": false,
			"allow_retrieval": false, "writes_business_truth": false,
		},
	})
	if err != nil {
		return "", WorkflowRunGatewayFailureProtocol, WorkflowRAGFailureGatewayFailed
	}
	temperature := service.defaultTemperature
	if request.Temperature != nil {
		temperature = *request.Temperature
	}
	envelope, err := service.bridge.HandleEnvelope(ctx, canonicalRequest, service.gatewayOptions(selection, temperature))
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", WorkflowRunGatewayFailureCanceled, WorkflowRAGFailureCanceled
		}
		return "", workflowRunGatewayCategory(err), WorkflowRAGFailureGatewayFailed
	}
	if !strings.EqualFold(strings.TrimSpace(envelope.Status), "ok") || envelope.Error != nil || envelope.Response == nil {
		return "", WorkflowRunGatewayFailureProviderFailed, WorkflowRAGFailureGatewayFailed
	}
	if structured, ok := envelope.Response["structured_answer"]; ok {
		payload, marshalErr := json.Marshal(structured)
		if marshalErr != nil {
			return "", WorkflowRunGatewayFailureOutputUnavailable, WorkflowRAGFailureAnswerInvalid
		}
		return string(payload), WorkflowRunGatewayFailureNone, ""
	}
	output := strings.TrimSpace(buildNorthboundResponseContent(envelope))
	if output == "" || len([]byte(output)) > workflowExecutorMaxOutputBytes {
		return "", WorkflowRunGatewayFailureOutputUnavailable, WorkflowRAGFailureAnswerInvalid
	}
	return output, WorkflowRunGatewayFailureNone, ""
}

func (service workflowRAGExecutionService) gatewayOptions(selection northboundSelection, temperature float64) bridge.EnvelopeOptions {
	if service.envelopeOptions != nil {
		return service.envelopeOptions(selection, temperature)
	}
	server := &Server{bridge: service.bridge}
	return server.buildBridgeEnvelopeOptions(selection, temperature)
}

func buildWorkflowRAGGatewayPacket(query string, promptNode SavedWorkflowDraftNode, selected []WorkflowRAGRankedFragment) (string, error) {
	evidence := make([]workflowRAGGatewayEvidence, 0, len(selected))
	for _, fragment := range selected {
		evidence = append(evidence, workflowRAGGatewayEvidence{
			FragmentRef: fragment.FragmentRef, Rank: fragment.Rank, SourceType: fragment.SourceType,
			IsOfficial: fragment.IsOfficial, Excerpt: fragment.Excerpt, ExcerptTruncated: fragment.ExcerptTruncated,
		})
	}
	payload, err := json.Marshal(evidence)
	if err != nil || len(payload) > workflowRAGMaxContextBytes+8*1024 {
		return "", errWorkflowRAGStoreContract
	}
	return fmt.Sprintf(
		"%s\n\n用户问题：\n%s\n\n仅可使用以下已召回证据回答：\n%s\n\n输出且只输出 workflow_rag_answer.v1 JSON；至少一个 citation，fragment_ref 只能来自上述证据；limitations 为字符串数组；confidence 只能是 low、medium 或 high。",
		strings.TrimSpace(promptNode.InputSummary), strings.TrimSpace(query), string(payload),
	), nil
}
