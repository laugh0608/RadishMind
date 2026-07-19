package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

type driftingWorkflowDefinitionRepository struct {
	workflowDefinitionReleaseRepository
	mu    sync.Mutex
	reads int
}

type driftingWorkflowDefinitionVersionRepository struct {
	workflowDefinitionReleaseRepository
	mu    sync.Mutex
	reads int
}

type ineligibleWorkflowDefinitionVersionRepository struct {
	workflowDefinitionReleaseRepository
	mu    sync.Mutex
	reads int
}

func (repository *driftingWorkflowDefinitionVersionRepository) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) (WorkflowDefinitionVersion, error) {
	value, err := repository.workflowDefinitionReleaseRepository.ReadVersion(ctx, definitionID, version)
	repository.mu.Lock()
	repository.reads++
	reads := repository.reads
	repository.mu.Unlock()
	if err == nil && reads > 1 {
		value.DefinitionDigest = "sha256:" + strings.Repeat("b", 64)
	}
	return value, err
}

func (repository *ineligibleWorkflowDefinitionVersionRepository) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) (WorkflowDefinitionVersion, error) {
	value, err := repository.workflowDefinitionReleaseRepository.ReadVersion(ctx, definitionID, version)
	repository.mu.Lock()
	repository.reads++
	reads := repository.reads
	repository.mu.Unlock()
	if err == nil && reads > 1 {
		value.ActivationEligible = false
		value.EligibilityBlockers = []string{"profile_drift"}
	}
	return value, err
}

type driftingApplicationCatalogRepository struct {
	applicationCatalogRepository
	mu    sync.Mutex
	reads int
}

func (repository *driftingApplicationCatalogRepository) RequireActive(ctx ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	value, err := repository.applicationCatalogRepository.RequireActive(ctx, applicationID)
	repository.mu.Lock()
	repository.reads++
	reads := repository.reads
	repository.mu.Unlock()
	if err == nil && reads > 1 {
		return ApplicationCatalogRecord{}, errApplicationCatalogArchived
	}
	return value, err
}

func (repository *driftingWorkflowDefinitionRepository) ReadActivation(ctx WorkflowDefinitionReleaseContext, definitionID string) (WorkflowDefinitionActivation, error) {
	activation, err := repository.workflowDefinitionReleaseRepository.ReadActivation(ctx, definitionID)
	repository.mu.Lock()
	repository.reads++
	reads := repository.reads
	repository.mu.Unlock()
	if err == nil && reads > 1 {
		activation.PointerVersion++
	}
	return activation, err
}

func TestWorkflowDefinitionExecutionUsesExactAuthorityAndMetadataOnlyV5(t *testing.T) {
	service, runContext, request, bridgeClient, store := workflowDefinitionExecutionFixture(t)
	input := "private workflow definition input"
	request.InputText = input
	result := service.StartRun(runContext, request)
	if result.FailureCode != "" || result.Record == nil || result.Record.Status != WorkflowRunStatusSucceeded || result.AdvisoryOutput == "" {
		t.Fatalf("definition execution failed: %#v", result)
	}
	record := result.Record
	if record.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || record.ExecutionSourceKind != workflowDefinitionExecutionSourceKind || record.ExecutionSourceID != request.DefinitionID || record.ExecutionSourceVersion != request.ExpectedDefinitionVersion || record.ExecutionProfile != workflowDefinitionExecutorProfile {
		t.Fatalf("definition execution identity drifted: %#v", record)
	}
	if record.DraftID != "" || record.DraftVersion != 0 || record.Output != "" || record.InputDigest == "" || record.InputBytes != len([]byte(input)) || record.DefinitionAuthority == nil || record.DefinitionAuthority.SourceDraftID == "" {
		t.Fatalf("v5 authority or privacy projection is invalid: %#v", record)
	}
	for _, node := range record.Nodes {
		if node.OutputPreview != "" {
			t.Fatalf("v5 persisted node output preview: %#v", node)
		}
	}
	if bridgeClient.callCount() != 1 || record.SideEffects.ProviderCalls != 1 || record.SideEffects.ToolCalls != 0 || record.SideEffects.RetrievalCalls != 0 || record.SideEffects.BusinessWrites != 0 || record.SideEffects.ReplayWrites != 0 {
		t.Fatalf("unexpected definition execution side effects: %#v bridge=%d", record.SideEffects, bridgeClient.callCount())
	}
	payload, err := json.Marshal(record)
	if err != nil || strings.Contains(string(payload), input) || strings.Contains(string(payload), result.AdvisoryOutput) {
		t.Fatalf("v5 persisted sensitive execution material: err=%v payload=%s", err, payload)
	}
	stored, found, err := store.ReadRun(runContext, record.RunID)
	if err != nil || !found || stored.Output != "" || stored.DefinitionAuthority == nil {
		t.Fatalf("stored v5 record is unavailable: found=%t err=%v record=%#v", found, err, stored)
	}
	filtered, err := store.ListRuns(runContext, WorkflowRunListFilter{ExecutionSourceKind: workflowDefinitionExecutionSourceKind, ExecutionSourceID: request.DefinitionID, ExecutionSourceVersion: request.ExpectedDefinitionVersion, Limit: 10})
	if err != nil || len(filtered.Records) != 1 || filtered.Records[0].RunID != record.RunID {
		t.Fatalf("v5 execution source filter failed: %#v err=%v", filtered, err)
	}
	legacy, err := store.ListRuns(runContext, WorkflowRunListFilter{DraftID: record.DefinitionAuthority.SourceDraftID, Limit: 10})
	if err != nil || len(legacy.Records) != 0 {
		t.Fatalf("source draft provenance leaked into legacy draft filter: %#v err=%v", legacy, err)
	}
}

func TestWorkflowDefinitionExecutionFailsAuthorityDriftBeforeProviderCall(t *testing.T) {
	service, runContext, request, bridgeClient, store := workflowDefinitionExecutionFixture(t)
	service.repository = &driftingWorkflowDefinitionRepository{workflowDefinitionReleaseRepository: service.repository}
	result := service.StartRun(runContext, request)
	if result.FailureCode != WorkflowRunFailureDefinitionAuthority || result.Record == nil || result.Record.Status != WorkflowRunStatusFailed || bridgeClient.callCount() != 0 || result.Record.SideEffects.ProviderCalls != 0 {
		t.Fatalf("authority drift did not fail closed before provider call: %#v bridge=%d", result, bridgeClient.callCount())
	}
	stored, found, err := store.ReadRun(runContext, result.Record.RunID)
	if err != nil || !found || stored.Status != WorkflowRunStatusFailed || stored.FailureCode != WorkflowRunFailureDefinitionAuthority {
		t.Fatalf("authority drift terminal evidence missing: %#v found=%t err=%v", stored, found, err)
	}
}

func TestWorkflowDefinitionExecutionRechecksVersionDigestAndApplicationLifecycle(t *testing.T) {
	for name, mutate := range map[string]func(*workflowDefinitionExecutionService){
		"definition_digest": func(service *workflowDefinitionExecutionService) {
			service.repository = &driftingWorkflowDefinitionVersionRepository{workflowDefinitionReleaseRepository: service.repository}
		},
		"application_lifecycle": func(service *workflowDefinitionExecutionService) {
			service.applications = &driftingApplicationCatalogRepository{applicationCatalogRepository: service.applications}
		},
		"profile_eligibility": func(service *workflowDefinitionExecutionService) {
			service.repository = &ineligibleWorkflowDefinitionVersionRepository{workflowDefinitionReleaseRepository: service.repository}
		},
	} {
		t.Run(name, func(t *testing.T) {
			service, runContext, request, bridgeClient, _ := workflowDefinitionExecutionFixture(t)
			mutate(&service)
			result := service.StartRun(runContext, request)
			expected := WorkflowRunFailureDefinitionAuthority
			if name == "profile_eligibility" {
				expected = WorkflowRunFailureDefinitionIncompatible
			}
			if result.FailureCode != expected || result.Record == nil || result.Record.SideEffects.ProviderCalls != 0 || bridgeClient.callCount() != 0 {
				t.Fatalf("%s drift did not fail closed before provider: %#v bridge=%d", name, result, bridgeClient.callCount())
			}
		})
	}
}

func TestWorkflowDefinitionV5ComparisonEvaluationAndStaleReconciliation(t *testing.T) {
	service, runContext, request, bridgeClient, store := workflowDefinitionExecutionFixture(t)
	baseline := service.StartRun(runContext, request)
	candidate := service.StartRun(runContext, request)
	if baseline.Record == nil || candidate.Record == nil {
		t.Fatalf("definition comparison fixture failed: baseline=%#v candidate=%#v", baseline, candidate)
	}
	comparison := service.executor.CompareRuns(runContext, baseline.Record.RunID, candidate.Record.RunID)
	if comparison.FailureCode != "" || comparison.Comparison == nil || comparison.Comparison.SchemaVersion != workflowDefinitionRunComparisonSchemaVersion || comparison.Comparison.RunProfile != workflowDefinitionEvaluationProfile {
		t.Fatalf("definition comparison profile failed: %#v", comparison)
	}
	otherLineage := workflowDefinitionRunRecordForStoreTest(runContext, "run_definition_otherlineage")
	if err := store.UpsertRun(runContext, &otherLineage); err != nil {
		t.Fatalf("seed other definition lineage: %v", err)
	}
	otherLineage.Status = WorkflowRunStatusSucceeded
	otherLineage.CompletedAt = workflowRunTimestamp(time.Now().UTC())
	otherLineage.Diagnostic.TerminalWriteState = WorkflowRunTerminalWriteStored
	if err := store.UpsertRun(runContext, &otherLineage); err != nil {
		t.Fatalf("complete other definition lineage: %v", err)
	}
	if incompatible := service.executor.CompareRuns(runContext, baseline.Record.RunID, otherLineage.RunID); incompatible.FailureCode != WorkflowRunFailureDefinitionIncompatible || incompatible.Comparison != nil {
		t.Fatalf("cross-lineage definition comparison did not fail closed: %#v", incompatible)
	}
	evaluation := newWorkflowEvaluationService(newMemoryWorkflowEvaluationStore(10), store)
	created := evaluation.Create(runContext, WorkflowEvaluationCreateRequest{Name: "Definition regression", BaselineRunID: baseline.Record.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: candidate.Record.RunID, ExpectedClassification: comparison.Comparison.Classification}}})
	if created.FailureCode != "" || created.Case == nil {
		t.Fatalf("definition evaluation create failed: %#v", created)
	}
	review := evaluation.Review(runContext, created.Case.CaseID)
	if review.FailureCode != "" || review.Review == nil || review.Review.RunProfile != workflowDefinitionEvaluationProfile {
		t.Fatalf("definition evaluation profile failed: %#v", review)
	}
	promotedBaseline := service.StartRun(runContext, request)
	if promotedBaseline.Record == nil {
		t.Fatalf("definition baseline promotion fixture failed: %#v", promotedBaseline)
	}
	promotedComparison := service.executor.CompareRuns(runContext, promotedBaseline.Record.RunID, candidate.Record.RunID)
	if promotedComparison.FailureCode != "" || promotedComparison.Comparison == nil {
		t.Fatalf("definition promoted baseline comparison failed: %#v", promotedComparison)
	}
	promoted := evaluation.Revise(runContext, created.Case.CaseID, WorkflowEvaluationRevisionRequest{ExpectedVersion: created.Case.Version, RevisionKind: WorkflowEvaluationRevisionBaselinePromotion, Name: created.Case.Name, BaselineRunID: promotedBaseline.Record.RunID, Expectations: []WorkflowEvaluationExpectation{{CandidateRunID: candidate.Record.RunID, ExpectedClassification: promotedComparison.Comparison.Classification}}})
	if promoted.FailureCode != "" || promoted.Case == nil || promoted.Case.RevisionKind != WorkflowEvaluationRevisionBaselinePromotion {
		t.Fatalf("definition baseline promotion failed: %#v", promoted)
	}
	suiteService := newWorkflowEvaluationSuiteService(newMemoryWorkflowEvaluationSuiteStore(10), evaluation)
	suite := suiteService.Create(runContext, WorkflowEvaluationSuiteCreateRequest{Name: "Definition release suite", CaseRefs: []WorkflowEvaluationSuiteCaseRef{{CaseID: promoted.Case.CaseID, Version: promoted.Case.Version}}})
	if suite.FailureCode != "" || suite.Suite == nil {
		t.Fatalf("definition evaluation suite create failed: %#v", suite)
	}
	suiteReview := suiteService.Review(runContext, suite.Suite.SuiteID)
	if suiteReview.FailureCode != "" || suiteReview.Review == nil || len(suiteReview.Review.Items) != 1 || suiteReview.Review.Items[0].RunProfile != workflowDefinitionEvaluationProfile {
		t.Fatalf("definition evaluation suite profile failed: %#v", suiteReview)
	}
	if bridgeClient.callCount() != 3 {
		t.Fatalf("comparison, evaluation, baseline, or suite re-executed the definition: bridge=%d", bridgeClient.callCount())
	}
	stale := cloneWorkflowRunRecord(*baseline.Record)
	stale.RunID = "run_definitionstale01"
	stale.RecordVersion = 0
	stale.Status = WorkflowRunStatusRunning
	stale.CompletedAt = ""
	stale.FailureCode = ""
	stale.FailureSummary = ""
	stale.SideEffects.ProviderCalls = 0
	stale.Diagnostic = newWorkflowRunDiagnostic()
	stale.StartedAt = workflowRunTimestamp(time.Now().Add(-workflowExecutorDefaultMaxRuntime - time.Second))
	if err := store.UpsertRun(runContext, &stale); err != nil {
		t.Fatalf("seed stale v5 run: %v", err)
	}
	if reconciled := service.ReconcileStale(runContext); reconciled.FailureCode != "" {
		t.Fatalf("reconcile stale v5 run: %#v", reconciled)
	}
	recovered, found, err := store.ReadRun(runContext, stale.RunID)
	if err != nil || !found || recovered.Status != WorkflowRunStatusFailed || recovered.FailureCode != WorkflowRunFailureDefinitionInterrupted || recovered.SideEffects.ProviderCalls != 0 {
		t.Fatalf("stale v5 run was replayed or not reconciled: %#v found=%t err=%v", recovered, found, err)
	}
}

func TestWorkflowDefinitionV5CodecRejectsUnknownAndSensitiveMaterial(t *testing.T) {
	runContext := workflowExecutorTestContext()
	record := workflowDefinitionRunRecordForStoreTest(runContext, "run_definition_codec")
	record.RecordVersion = 1
	payload, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeWorkflowRunStorageRecord(runContext, payload); err != nil {
		t.Fatalf("valid v5 codec rejected: %v", err)
	}
	unknown := append([]byte(nil), payload[:len(payload)-1]...)
	unknown = append(unknown, []byte(`,"prompt":"forbidden"}`)...)
	if _, err = decodeWorkflowRunStorageRecord(runContext, unknown); err == nil {
		t.Fatal("v5 codec accepted an unknown prompt field")
	}
	withOutput := record
	withOutput.Output = "forbidden answer"
	if validateWorkflowRunStoreRecord(runContext, &withOutput) == nil {
		t.Fatal("v5 store accepted a persisted answer")
	}
	if _, code, _ := normalizeWorkflowDefinitionRunRequest(WorkflowDefinitionRunRequest{DefinitionID: "definition_codec", ExpectedPointerVersion: 1, ExpectedDefinitionVersion: 1, ExpectedDefinitionDigest: "sha256:" + strings.Repeat("a", 64), InputText: "bounded input", ConditionValues: map[string]bool{}, Model: "token=forbidden"}); code != WorkflowRunFailureInputInvalid {
		t.Fatalf("v5 request accepted a credential-shaped model selector: %s", code)
	}
}

func TestWorkflowDefinitionExecutionHTTPStrictAuthorityAndScope(t *testing.T) {
	server := NewServer(config.Config{ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true, WorkflowDefinitionReleaseDevEnabled: true, WorkflowExecutorDevEnabled: true, Provider: "mock"}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	bridgeClient := &workflowExecutorTestBridge{}
	server.bridge = bridgeClient
	applicationID := "app_aaaaaaaaaaaaaaaa"
	actor := "subject_demo_user"
	catalog := server.applicationCatalogService()
	catalog.newID = func() (string, error) { return applicationID, nil }
	appContext := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_http_app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit_http_app", WriteEnabled: true}
	if created := catalog.Create(appContext, ApplicationCatalogCreateInput{DisplayName: "HTTP definition app", ApplicationKind: "workflow_copilot"}); created.FailureCode != "" {
		t.Fatalf("create HTTP application fixture: %#v", created)
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID, OwnerSubjectRef: actor, ActorRef: actor, RequestID: "request_http_definition", AuditRef: "audit_http_definition"}
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID, draft.ToolRefs, draft.RAGRefs, draft.RequestedCapabilities = applicationID, []string{}, []string{}, []string{}
	candidate, err := server.workflowDefinitionReleaseRepository.CreateCandidate(releaseContext, "candidate_http", "definition_http", draft, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	_, version, err := server.workflowDefinitionReleaseRepository.Review(releaseContext, candidate.CandidateID, 0, "approve", "approve HTTP definition", candidate.SourceDraftDigest, time.Now().UTC())
	if err != nil || version == nil {
		t.Fatalf("approve HTTP definition: %#v %v", version, err)
	}
	activation, err := server.workflowDefinitionReleaseRepository.DecideActivation(releaseContext, version.DefinitionID, 0, "activate", version.Version, "activate HTTP definition", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	body := workflowDefinitionRunHTTPBody{WorkspaceID: "workspace_demo", ApplicationID: applicationID, DefinitionID: version.DefinitionID, ExpectedPointerVersion: activation.PointerVersion, ExpectedDefinitionVersion: version.Version, ExpectedDefinitionDigest: version.DefinitionDigest, InputText: "HTTP bounded input", ConditionValues: map[string]bool{}}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-definition-runs", bytes.NewReader(mustWorkflowRunJSON(t, body)))
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, "workflow_runs:execute,workflow_definitions:read")
	request.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	envelope := decodeWorkflowRunEnvelope(t, response, http.StatusOK)
	if envelope.FailureCode != nil || envelope.Run == nil || envelope.Run.SchemaVersion != workflowRunRecordDefinitionSchemaVersion || envelope.AdvisoryOutput == "" || bridgeClient.callCount() != 1 {
		t.Fatalf("HTTP definition execution failed: %#v body=%s bridge=%d", envelope, response.Body.String(), bridgeClient.callCount())
	}
	unknownPayload := append(mustWorkflowRunJSON(t, body)[:len(mustWorkflowRunJSON(t, body))-1], []byte(`,"credential":"forbidden"}`)...)
	unknownRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-definition-runs", bytes.NewReader(unknownPayload))
	setControlPlaneReadDevAuthHeaders(unknownRequest)
	unknownRequest.Header.Set(controlPlaneReadDevScopesHeader, "workflow_runs:execute,workflow_definitions:read")
	unknownRequest.Header.Set(savedWorkflowDraftDevWorkspaceHeader, "workspace_demo")
	unknownRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
	unknownResponse := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusBadRequest || bridgeClient.callCount() != 1 {
		t.Fatalf("unknown definition execution field did not fail before provider: status=%d body=%s bridge=%d", unknownResponse.Code, unknownResponse.Body.String(), bridgeClient.callCount())
	}
}

func workflowDefinitionExecutionFixture(t *testing.T) (workflowDefinitionExecutionService, WorkflowRunContext, WorkflowDefinitionRunRequest, *workflowExecutorTestBridge, *memoryWorkflowRunStore) {
	t.Helper()
	applicationID := "app_aaaaaaaaaaaaaaaa"
	owner := "subject_demo"
	applicationRepository := newMemoryApplicationCatalogRepository()
	applicationService := newApplicationCatalogService(applicationRepository)
	applicationService.newID = func() (string, error) { return applicationID, nil }
	applicationContext := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_app", TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner, AuditRef: "audit_app", WriteEnabled: true}
	if result := applicationService.Create(applicationContext, ApplicationCatalogCreateInput{DisplayName: "Definition app", ApplicationKind: "workflow_copilot"}); result.FailureCode != "" || result.Record == nil {
		t.Fatalf("create application fixture: %#v", result)
	}
	releaseContext := WorkflowDefinitionReleaseContext{RequestContext: context.Background(), TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", ApplicationID: applicationID, OwnerSubjectRef: owner, ActorRef: owner, RequestID: "request_definition", AuditRef: "audit_definition"}
	draft := executableWorkflowDraftForTest()
	draft.ApplicationID = applicationID
	draft.ToolRefs = []string{}
	draft.RAGRefs = []string{}
	draft.RequestedCapabilities = []string{}
	repository := newWorkflowDefinitionReleaseStore()
	candidate, err := repository.CreateCandidate(releaseContext, "candidate_definition", "definition_runtime", draft, time.Now().UTC())
	if err != nil {
		t.Fatalf("create definition candidate fixture: %v", err)
	}
	_, version, err := repository.Review(releaseContext, candidate.CandidateID, 0, "approve", "approve exact definition runtime", candidate.SourceDraftDigest, time.Now().UTC())
	if err != nil || version == nil {
		t.Fatalf("approve definition fixture: version=%#v err=%v", version, err)
	}
	activation, err := repository.Activate(releaseContext, version.DefinitionID, 0, version.Version, time.Now().UTC())
	if err != nil {
		t.Fatalf("activate definition fixture: %v", err)
	}
	store := newMemoryWorkflowRunStore(20)
	bridgeClient := &workflowExecutorTestBridge{}
	executor := newWorkflowExecutorService(nil, bridgeClient, store)
	runContext := WorkflowRunContext{RequestContext: context.Background(), RequestID: "request_run", TenantRef: releaseContext.TenantRef, WorkspaceID: releaseContext.WorkspaceID, ApplicationID: applicationID, ActorRef: owner, AuditRef: "audit_run"}
	service := newWorkflowDefinitionExecutionService(repository, applicationRepository, executor)
	request := WorkflowDefinitionRunRequest{DefinitionID: version.DefinitionID, ExpectedPointerVersion: activation.PointerVersion, ExpectedDefinitionVersion: version.Version, ExpectedDefinitionDigest: version.DefinitionDigest, InputText: "bounded input", ConditionValues: map[string]bool{}}
	return service, runContext, request, bridgeClient, store
}

func workflowDefinitionRunRecordForStoreTest(ctx WorkflowRunContext, runID string) WorkflowRunRecord {
	record := workflowRunHistoryTestRecord(ctx, runID, "draft_definition_source", time.Now().UTC().Add(-time.Second))
	record.SchemaVersion = workflowRunRecordDefinitionSchemaVersion
	record.DraftID, record.DraftVersion, record.DraftDigest = "", 0, ""
	record.ExecutionKind = workflowDefinitionExecutionKind
	record.ExecutionSourceKind = workflowDefinitionExecutionSourceKind
	record.ExecutionSourceID = "definition_store"
	record.ExecutionSourceVersion = 2
	record.ExecutionProfile = workflowDefinitionExecutorProfile
	record.ExecutionSource = &workflowRunExecutionSource{Kind: record.ExecutionKind, SourceKind: record.ExecutionSourceKind, ID: record.ExecutionSourceID, Version: record.ExecutionSourceVersion}
	record.InputDigest = workflowDefinitionInputDigest("store test input")
	record.InputBytes = len("store test input")
	record.DefinitionAuthority = &WorkflowDefinitionRunAuthority{DefinitionID: record.ExecutionSourceID, DefinitionVersion: record.ExecutionSourceVersion, DefinitionDigest: "sha256:" + strings.Repeat("a", 64), ActivationPointerVersion: 3, CandidateID: "candidate_store", CandidateReviewVersion: 1, SourceDraftID: "draft_definition_source", SourceDraftVersion: 4, SourceDraftDigest: "sha256:" + strings.Repeat("a", 64), ApplicationRecordVersion: 1, ApplicationLifecycle: applicationCatalogLifecycleActive}
	record.Output = ""
	for index := range record.Nodes {
		record.Nodes[index].OutputPreview = ""
	}
	return record
}
