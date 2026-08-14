package httpapi

import (
	"bytes"
	"encoding/json"
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
	InputText                 json.RawMessage `json:"input_text,omitempty"`
	Inputs                    json.RawMessage `json:"inputs,omitempty"`
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
	auth, failure, status := server.authorizeWorkspaceScopedPermissions(request, "workflow_runs:execute", "workflow_definitions:read")
	runContext := workflowRunMutationContext(request, trace, auth, "", "definition-start")
	if failure != "" {
		writeWorkflowRunResultWithStatus(writer, status, trace, runContext, workflowRunFailure(WorkflowRunFailureCode(failure), "Workflow definition run authorization is denied."))
		return
	}
	var body workflowDefinitionRunHTTPBody
	if !server.decodeJSONRequestBody(writer, request, trace, &body, jsonRequestBodyOptions{maxBytes: maxControlJSONRequestBodyBytes, rejectUnknownFields: true}) {
		return
	}
	runContext = workflowRunMutationContext(request, trace, auth, body.ApplicationID, "definition-start")
	if !workflowMutationBindingMatches(request, auth, body.WorkspaceID, runContext.ApplicationID) {
		writeWorkflowRunResultWithStatus(writer, http.StatusForbidden, trace, runContext, workflowRunFailure(WorkflowRunFailureCode("workspace_binding_mismatch"), "Workflow definition run workspace binding is denied."))
		return
	}
	runRequest, code, summary := workflowDefinitionRunRequestFromHTTPBody(body)
	if code != "" {
		writeWorkflowRunResult(writer, trace, runContext, workflowRunFailure(code, summary))
		return
	}
	result := server.workflowDefinitionExecutionService().StartRun(runContext, runRequest)
	writeWorkflowRunResult(writer, trace, runContext, result)
}

func workflowDefinitionRunRequestFromHTTPBody(body workflowDefinitionRunHTTPBody) (WorkflowDefinitionRunRequest, WorkflowRunFailureCode, string) {
	request := WorkflowDefinitionRunRequest{
		DefinitionID:              strings.TrimSpace(body.DefinitionID),
		ExpectedPointerVersion:    body.ExpectedPointerVersion,
		ExpectedDefinitionVersion: body.ExpectedDefinitionVersion,
		ExpectedDefinitionDigest:  strings.TrimSpace(body.ExpectedDefinitionDigest),
		ConditionValues:           body.ConditionValues,
		Model:                     body.Model,
		Temperature:               body.Temperature,
		inputTextProvided:         body.InputText != nil,
		inputsProvided:            body.Inputs != nil,
	}
	if body.InputText != nil {
		if err := json.Unmarshal(body.InputText, &request.InputText); err != nil {
			return WorkflowDefinitionRunRequest{}, WorkflowRunFailureInputValueTypeInvalid, "Workflow definition input_text must be a JSON string."
		}
	}
	if body.Inputs != nil {
		decoder := json.NewDecoder(bytes.NewReader(body.Inputs))
		decoder.UseNumber()
		if err := decoder.Decode(&request.Inputs); err != nil || request.Inputs == nil {
			return WorkflowDefinitionRunRequest{}, WorkflowRunFailureInputContractMismatch, "Workflow structured inputs must be a JSON object."
		}
	}
	return request, "", ""
}

func (server *Server) workflowDefinitionExecutionService() workflowDefinitionExecutionService {
	return newWorkflowDefinitionExecutionService(server.workflowDefinitionReleaseRepository, server.applicationCatalogRepository, server.workflowExecutorService())
}
