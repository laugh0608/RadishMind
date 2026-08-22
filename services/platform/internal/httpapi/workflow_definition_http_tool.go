package httpapi

import "strings"

const (
	workflowDefinitionHTTPToolProfile                = "workflow_definition_http_tool_v1"
	workflowDefinitionHTTPToolCandidateSchemaVersion = "workflow_definition_release_candidate.v3"
	workflowDefinitionHTTPToolVersionSchemaVersion   = "workflow_definition_version.v3"
)

func workflowDefinitionHTTPToolEligibility(draft SavedWorkflowDraft) (bool, []string) {
	registry, err := newWorkflowHTTPToolRegistry()
	if err != nil || !registry.profile.Enabled || registry.definitionDigest == "" || registry.profileDigest == "" {
		return false, []string{"workflow_http_tool_registry_unavailable"}
	}
	nodeID := ""
	for _, node := range draft.Nodes {
		if strings.EqualFold(strings.TrimSpace(node.NodeType), "http_tool") {
			if nodeID != "" {
				return false, []string{"execution_profile_incompatible:workflow_http_tool_v1"}
			}
			nodeID = strings.TrimSpace(node.NodeID)
		}
	}
	if nodeID == "" || validateWorkflowHTTPToolDraft(draft, nodeID, registry.definition) != nil {
		return false, []string{"execution_profile_incompatible:workflow_http_tool_v1"}
	}
	return true, nil
}

func workflowDefinitionExecutionProfileForDraft(draft SavedWorkflowDraft, requested string) (string, bool) {
	requested = strings.TrimSpace(requested)
	_, _, defaultProfile, ok := workflowDefinitionSchemaIdentityForDraft(draft.SchemaVersion)
	if !ok {
		return "", false
	}
	if requested == "" {
		return defaultProfile, true
	}
	_, _, supported := workflowDefinitionReleaseIdentityForDraft(draft.SchemaVersion, requested)
	return requested, supported
}

func workflowDefinitionDraftEligibleForProfile(draft SavedWorkflowDraft, executionProfile string) (bool, []string) {
	if executionProfile == workflowDefinitionHTTPToolProfile {
		return workflowDefinitionHTTPToolEligibility(draft)
	}
	return workflowDefinitionDefaultEligibility(draft)
}

func workflowDefinitionDraftMatchesProfile(draft SavedWorkflowDraft, executionProfile string) bool {
	if executionProfile == workflowDefinitionHTTPToolProfile {
		eligible, _ := workflowDefinitionHTTPToolEligibility(draft)
		return eligible
	}
	return workflowDefinitionDraftMatchesDefaultContract(draft)
}

func workflowDefinitionActivationBlocker(draft SavedWorkflowDraft, executionProfile string) string {
	if executionProfile == workflowDefinitionHTTPToolProfile {
		eligible, blockers := workflowDefinitionHTTPToolEligibility(draft)
		if !eligible && len(blockers) > 0 {
			return blockers[0]
		}
		return ""
	}
	return workflowDefinitionExecutionBlocker(draft)
}
