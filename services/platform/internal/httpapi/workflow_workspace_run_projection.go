package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

const workspaceRunCursorVersion = 1

type workspaceRunProjectionCursor struct {
	Version       int    `json:"v"`
	StartedAt     string `json:"started_at"`
	RunID         string `json:"run_id"`
	ApplicationID string `json:"application_id"`
	BindingDigest string `json:"binding_digest"`
}

type workspaceRunProjectionResult struct {
	Records    []WorkflowRunRecord
	NextCursor string
	HasMore    bool
	Failure    ReadRepositoryFailureCode
}

func listWorkspaceRunProjection(
	projection workflowWorkspaceRunProjection,
	readContext ReadRepositoryContext,
	request ListRunRecordSummariesRequest,
) workspaceRunProjectionResult {
	if projection == nil {
		return workspaceRunProjectionFailure(ReadRepositoryFailureStoreUnavailable)
	}
	applicationID := strings.TrimSpace(request.Filters["application_ref"])
	workflowDefinitionID := strings.TrimSpace(request.Filters["workflow_definition_id"])
	if (applicationID != "" && !validControlPlaneReadAuthReference(applicationID, false)) ||
		(workflowDefinitionID != "" && !validControlPlaneReadAuthReference(workflowDefinitionID, false)) ||
		(request.Sort != "" && request.Sort != "started_at_desc") {
		return workspaceRunProjectionFailure(ReadRepositoryFailureInvalidFilter)
	}
	executionSourceKind := ""
	if workflowDefinitionID != "" {
		executionSourceKind = workflowDefinitionExecutionSourceKind
	}
	baseFilter, failureCode := normalizeWorkflowRunListRequest(WorkflowRunListRequest{
		Limit: request.Limit, Status: WorkflowRunStatus(strings.TrimSpace(request.Filters["status"])),
		ExecutionSourceKind: executionSourceKind, ExecutionSourceID: workflowDefinitionID,
		FailureCode: WorkflowRunFailureCode(strings.TrimSpace(request.Filters["failure_code"])),
	})
	if failureCode != "" {
		return workspaceRunProjectionFailure(ReadRepositoryFailureInvalidFilter)
	}
	projectionContext := WorkflowWorkspaceRunListContext{
		RequestContext: readContext.RequestContext, TenantRef: readContext.TenantRef,
		WorkspaceID: readContext.WorkspaceID, OwnerSubjectRef: readContext.SubjectRef,
	}
	filter := WorkflowWorkspaceRunListFilter{
		WorkflowRunListFilter: baseFilter,
		ApplicationID:         applicationID,
	}
	if strings.TrimSpace(request.Cursor) != "" {
		cursor, err := decodeWorkspaceRunProjectionCursor(request.Cursor, projectionContext, filter)
		if err != nil {
			return workspaceRunProjectionFailure(ReadRepositoryFailureInvalidFilter)
		}
		filter.BeforeTime = &cursor.StartedAt
		filter.BeforeRunID = cursor.RunID
		filter.BeforeApplicationID = cursor.ApplicationID
	}
	page, err := projection.ListWorkspaceRuns(projectionContext, filter)
	if err != nil {
		return workspaceRunProjectionFailure(translateWorkflowRunReadFailure(workflowRunStoreFailureCode(err)))
	}
	result := workspaceRunProjectionResult{Records: page.Records, HasMore: page.HasMore}
	if page.HasMore && len(page.Records) > 0 {
		cursor, cursorErr := encodeWorkspaceRunProjectionCursor(
			page.Records[len(page.Records)-1],
			projectionContext,
			WorkflowWorkspaceRunListFilter{WorkflowRunListFilter: baseFilter, ApplicationID: applicationID},
		)
		if cursorErr != nil {
			return workspaceRunProjectionFailure(ReadRepositoryFailureContractMismatch)
		}
		result.NextCursor = cursor
	}
	return result
}

func workspaceRunProjectionFailure(code ReadRepositoryFailureCode) workspaceRunProjectionResult {
	return workspaceRunProjectionResult{Records: []WorkflowRunRecord{}, Failure: code}
}

func encodeWorkspaceRunProjectionCursor(
	record WorkflowRunRecord,
	runContext WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (string, error) {
	if strings.TrimSpace(record.RunID) == "" || strings.TrimSpace(record.ApplicationID) == "" {
		return "", errWorkflowRunStoreContract
	}
	if _, err := time.Parse(time.RFC3339Nano, record.StartedAt); err != nil {
		return "", errWorkflowRunStoreContract
	}
	bindingDigest, err := workspaceRunProjectionBindingDigest(runContext, filter)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(workspaceRunProjectionCursor{
		Version: workspaceRunCursorVersion, StartedAt: record.StartedAt,
		RunID: record.RunID, ApplicationID: record.ApplicationID, BindingDigest: bindingDigest,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeWorkspaceRunProjectionCursor(
	raw string,
	runContext WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (struct {
	StartedAt     time.Time
	RunID         string
	ApplicationID string
}, error) {
	var decoded struct {
		StartedAt     time.Time
		RunID         string
		ApplicationID string
	}
	if len(raw) > 1024 {
		return decoded, errWorkflowRunStoreContract
	}
	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return decoded, errWorkflowRunStoreContract
	}
	var cursor workspaceRunProjectionCursor
	if err = json.Unmarshal(payload, &cursor); err != nil {
		return decoded, errWorkflowRunStoreContract
	}
	bindingDigest, err := workspaceRunProjectionBindingDigest(runContext, filter)
	if err != nil || cursor.Version != workspaceRunCursorVersion || cursor.BindingDigest != bindingDigest ||
		!validControlPlaneReadAuthReference(cursor.RunID, false) ||
		!validControlPlaneReadAuthReference(cursor.ApplicationID, false) {
		return decoded, errWorkflowRunStoreContract
	}
	startedAt, err := time.Parse(time.RFC3339Nano, cursor.StartedAt)
	if err != nil {
		return decoded, errWorkflowRunStoreContract
	}
	decoded.StartedAt = startedAt
	decoded.RunID = cursor.RunID
	decoded.ApplicationID = cursor.ApplicationID
	return decoded, nil
}

func workspaceRunProjectionBindingDigest(
	runContext WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (string, error) {
	type cursorBinding struct {
		TenantRef              string                     `json:"tenant_ref"`
		WorkspaceID            string                     `json:"workspace_id"`
		OwnerSubjectRef        string                     `json:"owner_subject_ref"`
		ApplicationID          string                     `json:"application_id"`
		Status                 WorkflowRunStatus          `json:"status"`
		ExecutionSourceKind    string                     `json:"execution_source_kind"`
		ExecutionSourceID      string                     `json:"execution_source_id"`
		ExecutionSourceVersion int                        `json:"execution_source_version"`
		FailureCode            WorkflowRunFailureCode     `json:"failure_code"`
		FailureBoundary        WorkflowRunFailureBoundary `json:"failure_boundary"`
		Provider               string                     `json:"provider"`
		Model                  string                     `json:"model"`
		Limit                  int                        `json:"limit"`
		Sort                   string                     `json:"sort"`
	}
	payload, err := json.Marshal(cursorBinding{
		TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID,
		OwnerSubjectRef: runContext.OwnerSubjectRef, ApplicationID: filter.ApplicationID,
		Status: filter.Status, ExecutionSourceKind: filter.ExecutionSourceKind,
		ExecutionSourceID: filter.ExecutionSourceID, ExecutionSourceVersion: filter.ExecutionSourceVersion,
		FailureCode: filter.FailureCode, FailureBoundary: filter.FailureBoundary,
		Provider: filter.Provider, Model: filter.Model, Limit: filter.Limit, Sort: "started_at_desc",
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}
