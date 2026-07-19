package httpapi

import (
	"net/http"
	"strings"
)

const workflowDefinitionRunCreateRoute = "POST /v1/user-workspace/workflow-definition-runs"

type workflowDefinitionRunHTTPBody struct {
	WorkspaceID               string          `json:"workspace_id"`
	ApplicationID             string          `json:"application_id"`
	DefinitionID              string          `json:"definition_id"`
	ExpectedPointerVersion    int             `json:"expected_pointer_version"`
	ExpectedDefinitionVersion int             `json:"expected_definition_version"`
	ExpectedDefinitionDigest  string          `json:"expected_definition_digest"`
	InputText                 string          `json:"input_text"`
	ConditionValues           map[string]bool `json:"condition_values"`
	Model                     string          `json:"model"`
	Temperature               *float64        `json:"temperature"`
}

func (server *Server) handleStartWorkflowDefinitionRun(writer http.ResponseWriter, request *http.Request) {
	trace := newRequestTrace(request, workflowDefinitionRunCreateRoute)
	if !server.config.WorkflowDefinitionReleaseDevEnabled || !server.config.WorkflowExecutorDevEnabled {
		server.writePlatformError(writer, trace, "WORKFLOW_DEFINITION_EXECUTOR_DEV_DISABLED", "workflow definition execution requires explicit development opt-in")
		return
	}
	var body workflowDefinitionRunHTTPBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	runContext, failure := workflowRunContextFromRequest(request, trace, body.WorkspaceID, body.ApplicationID, "definition-start", "workflow_runs:execute", "workflow_definitions:read")
	if failure != "" {
		writeWorkflowRunResult(writer, trace, runContext, workflowRunFailure(failure, "Workflow definition run scope is denied."))
		return
	}
	result := server.workflowDefinitionExecutionService().StartRun(runContext, WorkflowDefinitionRunRequest{DefinitionID: strings.TrimSpace(body.DefinitionID), ExpectedPointerVersion: body.ExpectedPointerVersion, ExpectedDefinitionVersion: body.ExpectedDefinitionVersion, ExpectedDefinitionDigest: strings.TrimSpace(body.ExpectedDefinitionDigest), InputText: body.InputText, ConditionValues: body.ConditionValues, Model: body.Model, Temperature: body.Temperature})
	writeWorkflowRunResult(writer, trace, runContext, result)
}

func (server *Server) workflowDefinitionExecutionService() workflowDefinitionExecutionService {
	return newWorkflowDefinitionExecutionService(server.workflowDefinitionReleaseRepository, server.applicationCatalogRepository, server.workflowExecutorService())
}
