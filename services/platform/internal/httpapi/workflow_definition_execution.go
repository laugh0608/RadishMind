package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const (
	workflowDefinitionExecutorProfile     = "workflow_definition_executor_v1"
	workflowDefinitionEvaluationProfile   = "workflow_definition_executor.v1"
	workflowDefinitionExecutionKind       = "workflow_definition_execution"
	workflowDefinitionExecutionSourceKind = "workflow_definition"
)

type WorkflowDefinitionRunRequest struct {
	DefinitionID              string
	ExpectedPointerVersion    int
	ExpectedDefinitionVersion int
	ExpectedDefinitionDigest  string
	InputText                 string
	ConditionValues           map[string]bool
	Model                     string
	Temperature               *float64
}

type workflowDefinitionExecutionAuthority struct {
	version     WorkflowDefinitionVersion
	activation  WorkflowDefinitionActivation
	application ApplicationCatalogRecord
	plan        workflowExecutionPlan
	draft       SavedWorkflowDraft
}

type workflowDefinitionExecutionService struct {
	repository   workflowDefinitionReleaseRepository
	applications applicationCatalogRepository
	executor     workflowExecutorService
}

func newWorkflowDefinitionExecutionService(repository workflowDefinitionReleaseRepository, applications applicationCatalogRepository, executor workflowExecutorService) workflowDefinitionExecutionService {
	return workflowDefinitionExecutionService{repository: repository, applications: applications, executor: executor}
}

func (service workflowDefinitionExecutionService) StartRun(runContext WorkflowRunContext, request WorkflowDefinitionRunRequest) WorkflowRunResult {
	normalized, code, summary := normalizeWorkflowDefinitionRunRequest(request)
	if code != "" {
		return workflowRunFailure(code, summary)
	}
	authority, code, summary := service.resolveAuthority(runContext, normalized)
	if code != "" {
		return workflowRunFailure(code, summary)
	}
	runID, err := service.executor.newRunID()
	if err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Unable to allocate a workflow definition run identifier.")
	}
	requestContext := runContext.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	maxRuntime := service.executor.maxRuntime
	if maxRuntime <= 0 {
		maxRuntime = workflowExecutorDefaultMaxRuntime
	}
	executionContext, cancel := context.WithTimeout(requestContext, maxRuntime)
	defer cancel()
	selection := resolveWorkflowRunSelection(service.executor, executionContext, normalized.Model)
	legacyRequest := WorkflowRunRequest{DraftID: authority.draft.DraftID, InputText: normalized.InputText, ConditionValues: normalized.ConditionValues, Model: normalized.Model, Temperature: normalized.Temperature}
	record := newWorkflowRunRecord(runContext, legacyRequest, authority.draft, authority.plan, selection, runID)
	record.SchemaVersion = workflowRunRecordDefinitionSchemaVersion
	record.DraftID = ""
	record.DraftVersion = 0
	record.DraftDigest = ""
	record.ExecutionKind = workflowDefinitionExecutionKind
	record.ExecutionSourceKind = workflowDefinitionExecutionSourceKind
	record.ExecutionSourceID = authority.version.DefinitionID
	record.ExecutionSourceVersion = authority.version.Version
	record.ExecutionProfile = workflowDefinitionExecutorProfile
	record.InputDigest = workflowDefinitionInputDigest(normalized.InputText)
	record.ExecutionSource = &workflowRunExecutionSource{Kind: workflowDefinitionExecutionKind, SourceKind: workflowDefinitionExecutionSourceKind, ID: authority.version.DefinitionID, Version: authority.version.Version}
	record.DefinitionAuthority = workflowDefinitionRunAuthority(authority)
	if err := service.executor.store.UpsertRun(runContext, &record); err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow definition run record storage is unavailable.")
	}
	checkpointRequest := normalized
	service.executor.beforeProviderCall = func(context.Context) (WorkflowRunFailureCode, string) {
		_, checkpointCode, checkpointSummary := service.resolveAuthority(runContext, checkpointRequest)
		return checkpointCode, checkpointSummary
	}
	return service.executor.executePlan(executionContext, runContext, legacyRequest, authority.draft, authority.plan, selection, record)
}

func (service workflowDefinitionExecutionService) resolveAuthority(runContext WorkflowRunContext, request WorkflowDefinitionRunRequest) (workflowDefinitionExecutionAuthority, WorkflowRunFailureCode, string) {
	ctx := WorkflowDefinitionReleaseContext{RequestContext: runContext.RequestContext, TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ApplicationID: runContext.ApplicationID, OwnerSubjectRef: runContext.ActorRef, ActorRef: runContext.ActorRef, RequestID: runContext.RequestID, AuditRef: runContext.AuditRef}
	activation, err := service.repository.ReadActivation(ctx, request.DefinitionID)
	if err != nil || activation.State != workflowDefinitionActivationActive || activation.PointerVersion != request.ExpectedPointerVersion || activation.ActiveVersion != request.ExpectedDefinitionVersion || activation.ActiveDefinitionDigest != request.ExpectedDefinitionDigest {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionAuthority, "Active workflow definition authority changed before provider execution."
	}
	version, err := service.repository.ReadVersion(ctx, request.DefinitionID, request.ExpectedDefinitionVersion)
	if err != nil || validateStoredWorkflowDefinitionVersion(version) != nil {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionAuthority, "Active workflow definition version is unavailable or invalid."
	}
	digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
	if err != nil || digest != version.DefinitionDigest || digest != activation.ActiveDefinitionDigest || digest != request.ExpectedDefinitionDigest {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionAuthority, "Active workflow definition digest changed before provider execution."
	}
	if !version.ActivationEligible || len(version.EligibilityBlockers) != 0 || version.Snapshot.ExecutionProfile != workflowDefinitionExecutorProfile {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionIncompatible, "Active workflow definition is not eligible for workflow_definition_executor_v1."
	}
	applicationContext := ApplicationCatalogContext{RequestContext: runContext.RequestContext, RequestID: runContext.RequestID, TenantRef: runContext.TenantRef, WorkspaceID: runContext.WorkspaceID, ActorRef: runContext.ActorRef, OwnerSubjectRef: runContext.ActorRef, AuditRef: runContext.AuditRef}
	application, err := service.applications.RequireActive(applicationContext, runContext.ApplicationID)
	if err != nil || application.LifecycleState != applicationCatalogLifecycleActive {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionAuthority, "Application lifecycle changed before provider execution."
	}
	draft := workflowDefinitionSnapshotAsDraft(runContext, version)
	plan, planCode, planSummary := buildWorkflowExecutionPlan(draft, request.ConditionValues)
	if planCode != "" {
		return workflowDefinitionExecutionAuthority{}, WorkflowRunFailureDefinitionIncompatible, planSummary
	}
	return workflowDefinitionExecutionAuthority{version: version, activation: activation, application: application, plan: plan, draft: draft}, "", ""
}

func normalizeWorkflowDefinitionRunRequest(request WorkflowDefinitionRunRequest) (WorkflowDefinitionRunRequest, WorkflowRunFailureCode, string) {
	request.DefinitionID = strings.TrimSpace(request.DefinitionID)
	request.ExpectedDefinitionDigest = strings.TrimSpace(request.ExpectedDefinitionDigest)
	legacy, code, summary := normalizeWorkflowRunRequest(WorkflowRunRequest{DraftID: request.DefinitionID, InputText: request.InputText, ConditionValues: request.ConditionValues, Model: request.Model, Temperature: request.Temperature})
	if code != "" || request.ExpectedPointerVersion < 1 || request.ExpectedDefinitionVersion < 1 || !workflowRAGDigestPattern.MatchString(request.ExpectedDefinitionDigest) {
		if code == WorkflowRunFailureBudgetExceeded {
			return WorkflowDefinitionRunRequest{}, code, summary
		}
		return WorkflowDefinitionRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow definition run authority and input are invalid."
	}
	if strings.Contains(legacy.Model, "://") || workflowRAGContainsForbiddenMaterial(legacy.Model) || applicationDraftStringContainsSecret(legacy.Model) {
		return WorkflowDefinitionRunRequest{}, WorkflowRunFailureInputInvalid, "Workflow definition run model selector is invalid."
	}
	request.InputText, request.ConditionValues, request.Model, request.Temperature = legacy.InputText, legacy.ConditionValues, legacy.Model, legacy.Temperature
	return request, "", ""
}

func workflowDefinitionSnapshotDigest(snapshot WorkflowDefinitionSnapshot) (string, error) {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func workflowDefinitionInputDigest(input string) string {
	digest := sha256.Sum256([]byte(input))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func workflowDefinitionSnapshotAsDraft(ctx WorkflowRunContext, version WorkflowDefinitionVersion) SavedWorkflowDraft {
	nodes := make([]SavedWorkflowDraftNode, 0, len(version.Snapshot.Nodes))
	for _, node := range version.Snapshot.Nodes {
		nodes = append(nodes, SavedWorkflowDraftNode{NodeID: node.NodeID, NodeType: node.NodeType, Label: node.Label, InputSummary: node.InputSummary, OutputSummary: node.OutputSummary, InputContractRef: node.InputContractRef, OutputContractRef: node.OutputContractRef, InputContractFields: cloneStringSlice(node.InputContractFields), OutputContractFields: cloneStringSlice(node.OutputContractFields), OutputMappingSummary: node.OutputMappingSummary, ProviderRef: node.ProviderRef, ToolRef: node.ToolRef, RAGRef: node.RAGRef, RiskLevel: node.RiskLevel, RequiresConfirmation: node.RequiresConfirmation})
	}
	edges := make([]SavedWorkflowDraftEdge, 0, len(version.Snapshot.Edges))
	for _, edge := range version.Snapshot.Edges {
		edges = append(edges, SavedWorkflowDraftEdge{EdgeID: edge.EdgeID, FromNodeID: edge.FromNodeID, ToNodeID: edge.ToNodeID, ConditionSummary: edge.ConditionSummary})
	}
	return SavedWorkflowDraft{DraftID: version.SourceDraftID, DraftVersion: version.SourceDraftVersion, WorkspaceID: ctx.WorkspaceID, ApplicationID: ctx.ApplicationID, SchemaVersion: savedWorkflowDraftSchemaVersion, DraftStatus: SavedWorkflowDraftStatusValidForReview, Nodes: nodes, Edges: edges, ProviderRefs: cloneStringSlice(version.Snapshot.ProviderRefs), ToolRefs: cloneStringSlice(version.Snapshot.ToolRefs), RAGRefs: cloneStringSlice(version.Snapshot.RAGRefs), RequestedCapabilities: cloneStringSlice(version.Snapshot.RequestedCapabilities), ValidationSummary: SavedWorkflowDraftValidationSummary{ValidationState: SavedWorkflowDraftStatusValidForReview, ValidForReview: true}}
}

func workflowDefinitionRunAuthority(value workflowDefinitionExecutionAuthority) *WorkflowDefinitionRunAuthority {
	return &WorkflowDefinitionRunAuthority{DefinitionID: value.version.DefinitionID, DefinitionVersion: value.version.Version, DefinitionDigest: value.version.DefinitionDigest, ActivationPointerVersion: value.activation.PointerVersion, CandidateID: value.version.CandidateID, CandidateReviewVersion: value.version.CandidateReviewVersion, SourceDraftID: value.version.SourceDraftID, SourceDraftVersion: value.version.SourceDraftVersion, SourceDraftDigest: value.version.SourceDraftDigest, ApplicationRecordVersion: value.application.RecordVersion, ApplicationLifecycle: value.application.LifecycleState}
}

func validateWorkflowDefinitionRunStoreRecord(runContext WorkflowRunContext, record *WorkflowRunRecord) error {
	if record == nil || record.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || record.DefinitionAuthority == nil ||
		record.ExecutionKind != workflowDefinitionExecutionKind || record.ExecutionSourceKind != workflowDefinitionExecutionSourceKind ||
		record.ExecutionProfile != workflowDefinitionExecutorProfile || record.ExecutionSourceID == "" || record.ExecutionSourceVersion < 1 ||
		record.Status == WorkflowRunStatusOutcomeUnknown ||
		record.ExecutionSource == nil || record.ExecutionSource.Kind != record.ExecutionKind || record.ExecutionSource.SourceKind != record.ExecutionSourceKind ||
		record.ExecutionSource.ID != record.ExecutionSourceID || record.ExecutionSource.Version != record.ExecutionSourceVersion ||
		record.DraftID != "" || record.DraftVersion != 0 || record.DraftDigest != "" || record.Output != "" ||
		!workflowRAGDigestPattern.MatchString(record.InputDigest) || record.InputBytes < 1 || record.InputBytes > workflowExecutorMaxInputBytes ||
		record.SideEffects.RetrievalCalls != 0 || record.SideEffects.ToolCalls != 0 || record.SideEffects.ConfirmationCalls != 0 ||
		record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 || record.SideEffects.ProviderCalls < 0 || record.SideEffects.ProviderCalls > workflowExecutorMaxLLMCalls ||
		record.PlanID != "" || record.ConfirmationID != "" || record.ToolAttempt != nil || record.RAGSnapshot != nil || record.RetrievalAttempt != nil || record.RAGAnswer != nil || record.RAGApplication != nil ||
		!validWorkflowRunDiagnostic(record.Diagnostic, isTerminalWorkflowRunStatus(record.Status)) {
		return errWorkflowRunStoreContract
	}
	authority := record.DefinitionAuthority
	if authority.DefinitionID != record.ExecutionSourceID || authority.DefinitionVersion != record.ExecutionSourceVersion ||
		!applicationDraftIdentifierPattern.MatchString(authority.DefinitionID) || !workflowRAGDigestPattern.MatchString(authority.DefinitionDigest) ||
		authority.ActivationPointerVersion < 1 || !applicationDraftIdentifierPattern.MatchString(authority.CandidateID) || authority.CandidateReviewVersion < 1 ||
		!applicationDraftIdentifierPattern.MatchString(authority.SourceDraftID) || authority.SourceDraftVersion < 1 ||
		!workflowRAGDigestPattern.MatchString(authority.SourceDraftDigest) || authority.SourceDraftDigest != authority.DefinitionDigest ||
		authority.ApplicationRecordVersion < 1 || authority.ApplicationLifecycle != applicationCatalogLifecycleActive {
		return errWorkflowRunStoreContract
	}
	for _, node := range record.Nodes {
		if node.OutputPreview != "" {
			return errWorkflowRunStoreContract
		}
	}
	return nil
}

func (service workflowDefinitionExecutionService) ReconcileStale(runContext WorkflowRunContext) WorkflowRunResult {
	stale := true
	page, err := service.executor.store.ListRuns(runContext, WorkflowRunListFilter{ExecutionSourceKind: workflowDefinitionExecutionSourceKind, StaleRunning: &stale, Limit: workflowRunListMaxLimit})
	if err != nil {
		return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow definition stale run reconciliation is unavailable.")
	}
	for _, record := range page.Records {
		if record.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || record.Status != WorkflowRunStatusRunning {
			continue
		}
		result := service.executor.finishFailedRun(runContext, record, WorkflowRunFailureDefinitionInterrupted, "Workflow definition execution was interrupted before a terminal record was stored.", false)
		if result.FailureCode != WorkflowRunFailureDefinitionInterrupted || result.Record == nil {
			return workflowRunFailure(WorkflowRunFailureStoreUnavailable, "Workflow definition stale run reconciliation is unavailable.")
		}
	}
	return WorkflowRunResult{}
}
