package httpapi

import (
	"errors"
	"strings"
)

func (service workflowHTTPToolActionService) readExactEligibleDefinition(
	ctx WorkflowHTTPToolActionContext,
	definitionOwnerRef string,
	definitionID string,
	nodeID string,
) (WorkflowDefinitionVersion, WorkflowDefinitionActivation, WorkflowHTTPToolActionResult) {
	if service.definitions == nil {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureStoreUnavailable,
			"Workflow definition authority is unavailable.",
		)
	}
	releaseContext := WorkflowDefinitionReleaseContext{
		RequestContext: ctx.RequestContext, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		ApplicationID: ctx.ApplicationID, OwnerSubjectRef: strings.TrimSpace(definitionOwnerRef), ActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	}
	activation, err := service.definitions.ReadActivation(releaseContext, strings.TrimSpace(definitionID))
	if errors.Is(err, errWorkflowDefinitionNotFound) {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureDefinitionNotFound,
			"Workflow definition was not found.",
		)
	}
	if err != nil {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureStoreUnavailable,
			"Workflow definition activation authority is unavailable.",
		)
	}
	if activation.DefinitionID != strings.TrimSpace(definitionID) || activation.State != workflowDefinitionActivationActive ||
		activation.ActiveVersion < 1 || activation.PointerVersion < 1 ||
		!workflowHTTPToolDigestPattern.MatchString(activation.ActiveDefinitionDigest) {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureDefinitionInactive,
			"Workflow definition has no active version.",
		)
	}
	version, err := service.definitions.ReadVersion(releaseContext, activation.DefinitionID, activation.ActiveVersion)
	if errors.Is(err, errWorkflowDefinitionNotFound) {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureDefinitionDrift,
			"Active workflow definition version could not be resolved.",
		)
	}
	if err != nil {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureStoreUnavailable,
			"Workflow definition version authority is unavailable.",
		)
	}
	if version.DefinitionID != activation.DefinitionID || version.Version != activation.ActiveVersion ||
		version.DefinitionDigest != activation.ActiveDefinitionDigest || version.SchemaVersion != workflowDefinitionHTTPToolVersionSchemaVersion ||
		version.Snapshot.ExecutionProfile != workflowDefinitionHTTPToolProfile || !version.ActivationEligible ||
		len(version.EligibilityBlockers) != 0 {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureDefinitionDrift,
			"Active workflow definition authority drifted before planning.",
		)
	}
	draft := workflowDefinitionSnapshotAsDraft(WorkflowRunContext{WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID}, version)
	if err := validateWorkflowHTTPToolGraph(draft, strings.TrimSpace(nodeID), service.registry.definition); err != nil {
		return WorkflowDefinitionVersion{}, WorkflowDefinitionActivation{}, workflowHTTPToolActionFailure(
			WorkflowHTTPToolActionFailureDefinitionIneligible,
			err.Error(),
		)
	}
	return version, activation, WorkflowHTTPToolActionResult{}
}

func workflowHTTPToolDefinitionPlanDigest(plan WorkflowHTTPToolActionPlan) (string, error) {
	payload := struct {
		SchemaVersion             string                          `json:"schema_version"`
		TenantRef                 string                          `json:"tenant_ref"`
		WorkspaceID               string                          `json:"workspace_id"`
		ApplicationID             string                          `json:"application_id"`
		SourceKind                string                          `json:"source_kind"`
		WorkflowDefinitionID      string                          `json:"workflow_definition_id"`
		WorkflowDefinitionVersion int                             `json:"workflow_definition_version"`
		WorkflowDefinitionDigest  string                          `json:"workflow_definition_digest"`
		ActivationPointerVersion  int                             `json:"activation_pointer_version"`
		NodeID                    string                          `json:"node_id"`
		ToolID                    string                          `json:"tool_id"`
		ToolVersion               int                             `json:"tool_version"`
		DefinitionDigest          string                          `json:"definition_digest"`
		ProfileID                 string                          `json:"profile_id"`
		ProfileVersion            int                             `json:"profile_version"`
		ProfileDigest             string                          `json:"profile_digest"`
		Method                    string                          `json:"method"`
		TargetPolicyKey           string                          `json:"target_policy_key"`
		PublicArguments           WorkflowHTTPToolPublicArguments `json:"public_arguments"`
		OutputFields              []string                        `json:"output_fields"`
		OutputSchemaDigest        string                          `json:"output_schema_digest"`
		CredentialPolicy          string                          `json:"credential_policy"`
		TimeoutMS                 int                             `json:"timeout_ms"`
		MaxResponseBytes          int                             `json:"max_response_bytes"`
		MaxOutputBytes            int                             `json:"max_output_bytes"`
		PlannedByActorRef         string                          `json:"planned_by_actor_ref"`
		CreatedAt                 string                          `json:"created_at"`
		ExpiresAt                 string                          `json:"expires_at"`
	}{
		plan.SchemaVersion, plan.TenantRef, plan.WorkspaceID, plan.ApplicationID, plan.SourceKind,
		plan.WorkflowDefinitionID, plan.WorkflowDefinitionVersion, plan.WorkflowDefinitionDigest,
		plan.ActivationPointerVersion, plan.NodeID, plan.ToolID, plan.ToolVersion, plan.DefinitionDigest,
		plan.ProfileID, plan.ProfileVersion, plan.ProfileDigest, plan.Method, plan.TargetPolicyKey,
		plan.PublicArguments, append([]string(nil), plan.OutputFields...), plan.OutputSchemaDigest,
		plan.CredentialPolicy, plan.TimeoutMS, plan.MaxResponseBytes, plan.MaxOutputBytes,
		plan.PlannedByActorRef, plan.CreatedAt, plan.ExpiresAt,
	}
	return canonicalSHA256(payload)
}
