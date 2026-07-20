package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type workflowRAGApplicationRuntimeTestFixture struct {
	promotionFixture  *workflowRAGPromotionTestFixture
	runtimeContext    WorkflowRAGApplicationRuntimeContext
	publishRepository *memoryApplicationPublishCandidateRepository
	runtimeRepository *memoryWorkflowRAGApplicationRuntimeRepository
	runtimeService    workflowRAGApplicationRuntimeService
	resolver          workflowRAGApplicationAuthorityResolver
	runStore          *memoryWorkflowRunStore
	publishCandidate  ApplicationPublishCandidate
	bridge            *workflowExecutorTestBridge
	now               time.Time
}

func TestWorkflowRAGApplicationRuntimeAssignmentLifecycleCASAndMetadataOnly(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "人工激活已批准的开发测试候选"})
	if activated.FailureCode != "" || activated.Assignment == nil || activated.Assignment.State != workflowRAGApplicationRuntimeStateActive || activated.Assignment.RecordVersion != 1 || len(activated.Events) != 1 || len(activated.Audits) != 1 {
		t.Fatalf("activate runtime assignment: %#v", activated)
	}
	if activated.Assignment.PublishCandidateID != fixture.publishCandidate.CandidateID || activated.Assignment.DraftDigest != fixture.publishCandidate.DraftDigest || activated.Assignment.BindingRef != *fixture.publishCandidate.Configuration.WorkflowRAGBindingRef || !workflowRAGDigestPattern.MatchString(activated.Assignment.AssignmentDigest) {
		t.Fatalf("assignment did not seal exact publish authority: %#v", activated.Assignment)
	}
	for contract, value := range map[string]any{workflowRAGApplicationRuntimeAssignmentSchemaVersion: activated.Assignment, workflowRAGApplicationRuntimeEventSchemaVersion: activated.Events[0], workflowRAGApplicationRuntimeAuditSchemaVersion: activated.Audits[0]} {
		payload, marshalErr := json.Marshal(value)
		if marshalErr != nil || validateWorkflowRAGContractJSON(contract, payload) != nil {
			t.Fatalf("strict contract %s rejected its stored value: payload=%s err=%v", contract, payload, marshalErr)
		}
	}
	assignmentPayload, _ := json.Marshal(activated.Assignment)
	assignmentPayload = append(assignmentPayload[:len(assignmentPayload)-1], []byte(`,"unknown_field":true}`)...)
	if validateWorkflowRAGContractJSON(workflowRAGApplicationRuntimeAssignmentSchemaVersion, assignmentPayload) == nil {
		t.Fatal("assignment contract accepted an unknown field")
	}
	read := fixture.runtimeService.Read(fixture.runtimeContext)
	if read.FailureCode != "" || read.Assignment == nil || read.Assignment.AssignmentDigest != activated.Assignment.AssignmentDigest {
		t.Fatalf("read active assignment: %#v", read)
	}
	conflict := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "过期版本不得撤销当前 assignment"})
	if conflict.FailureCode != WorkflowRAGApplicationFailureVersionConflict || conflict.CurrentRecordVersion != 1 {
		t.Fatalf("stale assignment decision was accepted: %#v", conflict)
	}
	revoked := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "人工撤销后续开发测试调用资格"})
	if revoked.FailureCode != "" || revoked.Assignment == nil || revoked.Assignment.State != workflowRAGApplicationRuntimeStateRevoked || revoked.Assignment.RecordVersion != 2 || len(revoked.Events) != 2 || len(revoked.Audits) != 2 {
		t.Fatalf("revoke runtime assignment: %#v", revoked)
	}
	duplicate := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 2, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "重复撤销必须稳定失败关闭"})
	if duplicate.FailureCode != WorkflowRAGApplicationFailureTransitionInvalid {
		t.Fatalf("duplicate revoke was accepted: %#v", duplicate)
	}
	payload, err := json.Marshal(revoked)
	if err != nil {
		t.Fatalf("marshal assignment result: %v", err)
	}
	for _, forbidden := range []string{"promotion authority query", "promotion authority fragment", "promotion review note", "Authorization:", "prompt"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("assignment metadata leaked %q: %s", forbidden, payload)
		}
	}
}

func TestWorkflowRAGApplicationRuntimeConcurrentCASAllowsOneActivation(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	fixture.runtimeService.newID = newWorkflowRAGStableID
	results := make(chan WorkflowRAGApplicationRuntimeResult, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "并发激活只允许一个决定成功"})
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		if result.FailureCode == "" {
			successes++
		} else if result.FailureCode == WorkflowRAGApplicationFailureVersionConflict {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent result: %#v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("CAS winners drifted: successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestWorkflowRAGApplicationRuntimeRejectsUnsupportedStoreWithoutFallback(t *testing.T) {
	if repository, err := newWorkflowRAGApplicationRuntimeRepositoryForRunStore(nil); err == nil || repository != nil {
		t.Fatalf("unsupported store unexpectedly fell back to a runtime repository: repository=%T err=%v", repository, err)
	}
}

func TestWorkflowRAGApplicationInvocationSucceedsOnceAndStoresV4MetadataOnly(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "人工激活候选用于一次受控调用"})
	if activated.FailureCode != "" {
		t.Fatalf("activate before invocation: %#v", activated)
	}
	query := "promotion authority query-secret-needle"
	service := newWorkflowRAGApplicationInvocationService(fixture.runtimeRepository, fixture.resolver, fixture.runStore, fixture.bridge)
	service.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	service.envelopeOptions = func(northboundSelection, float64) bridge.EnvelopeOptions { return bridge.EnvelopeOptions{} }
	service.newRunID = func() (string, error) { return "run_applicationrag01", nil }
	service.now = func() time.Time { return fixture.now }
	result := service.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: query})
	if result.FailureCode != "" || result.Run == nil || result.Answer == nil {
		t.Fatalf("application RAG invocation failed: %#v", result)
	}
	record := *result.Run
	if record.SchemaVersion != workflowRunRecordAppRAGSchemaVersion || record.ExecutionSource == nil || record.ExecutionSource.Kind != workflowRAGApplicationExecutionKind || record.ExecutionSource.SourceKind != workflowRAGApplicationExecutionSourceKind || record.DraftID != "" || record.Status != WorkflowRunStatusSucceeded || record.SideEffects.RetrievalCalls != 1 || record.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 || record.RAGAnswer != nil || record.RAGApplication == nil {
		t.Fatalf("unexpected v4 record: %#v bridge_calls=%d", record, fixture.bridge.callCount())
	}
	stored, found, err := fixture.runStore.ReadRun(workflowRAGApplicationRunContext(fixture.runtimeContext), record.RunID)
	if err != nil || !found || stored.RecordVersion < 4 {
		t.Fatalf("read stored v4 run: %#v found=%t err=%v", stored, found, err)
	}
	payload, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("marshal stored v4 run: %v", err)
	}
	if validateWorkflowRAGContractJSON(workflowRunRecordAppRAGSchemaVersion, payload) != nil {
		t.Fatalf("strict v4 contract rejected stored run: %s", payload)
	}
	oversized := cloneWorkflowRunRecord(stored)
	oversized.InputBytes = workflowRAGApplicationInvocationMaxBytes + 1
	oversized.RetrievalAttempt.QueryBytes = oversized.InputBytes
	oversizedPayload, err := json.Marshal(oversized)
	if err != nil {
		t.Fatalf("marshal oversized v4 run: %v", err)
	}
	if validateWorkflowRAGContractJSON(workflowRunRecordAppRAGSchemaVersion, oversizedPayload) == nil {
		t.Fatalf("strict v4 contract accepted oversized invocation metadata: %s", oversizedPayload)
	}
	duplicateCitation := cloneWorkflowRunRecord(stored)
	duplicateCitation.RetrievalAttempt.CitationRefs = append(duplicateCitation.RetrievalAttempt.CitationRefs, duplicateCitation.RetrievalAttempt.CitationRefs[0])
	duplicateCitationPayload, err := json.Marshal(duplicateCitation)
	if err != nil {
		t.Fatalf("marshal duplicate-citation v4 run: %v", err)
	}
	if validateWorkflowRAGContractJSON(workflowRunRecordAppRAGSchemaVersion, duplicateCitationPayload) == nil {
		t.Fatalf("strict v4 contract accepted duplicate citation refs: %s", duplicateCitationPayload)
	}
	answerPayload, _ := json.Marshal(result.Answer)
	if validateWorkflowRAGContractJSON(workflowRAGApplicationAnswerSchemaVersion, answerPayload) != nil {
		t.Fatalf("strict application answer contract rejected response: %s", answerPayload)
	}
	for _, forbidden := range []string{query, "query-secret-needle", "promotion authority fragment", result.Answer.Answer, "用户问题", "credential"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("metadata-only v4 run leaked %q: %s", forbidden, payload)
		}
	}
	if !strings.Contains(string(fixture.bridge.lastRequest()), `"allow_retrieval":false`) || !strings.Contains(string(fixture.bridge.lastRequest()), query) {
		t.Fatalf("canonical gateway request drifted: %s", fixture.bridge.lastRequest())
	}
	fixture.advanceRuntimeRequest("revoke")
	revoked := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 1, Decision: workflowRAGApplicationRuntimeDecisionRevoke, Reason: "人工撤销后必须阻止新调用"})
	if revoked.FailureCode != "" {
		t.Fatalf("revoke after invocation: %#v", revoked)
	}
	blocked := service.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "new invocation after revoke"})
	if blocked.FailureCode != WorkflowRAGApplicationFailureAssignmentRevoked || fixture.bridge.callCount() != 1 {
		t.Fatalf("revoked assignment did not fail before side effects: %#v calls=%d", blocked, fixture.bridge.callCount())
	}
}

func TestWorkflowRAGApplicationRunComparisonAllowsReviewedAuthorityChange(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "人工激活候选用于验证知识晋级比较"})
	if activated.FailureCode != "" {
		t.Fatalf("activate before comparison: %#v", activated)
	}
	service := newWorkflowRAGApplicationInvocationService(fixture.runtimeRepository, fixture.resolver, fixture.runStore, fixture.bridge)
	service.resolveSelection = func(context.Context, string) northboundSelection { return workflowRAGTestSelection() }
	service.newRunID = func() (string, error) { return "run_applicationragcompare01", nil }
	service.now = func() time.Time { return fixture.now }
	result := service.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "same application input for reviewed knowledge promotion"})
	if result.FailureCode != "" || result.Run == nil {
		t.Fatalf("invoke comparison baseline: %#v", result)
	}

	baseline := cloneWorkflowRunRecord(*result.Run)
	candidate := cloneWorkflowRunRecord(baseline)
	candidate.RunID = "run_applicationragcompare02"
	candidate.RAGApplication.AssignmentID = "wragra_bbbbbbbbbbbbbbbb"
	candidate.RAGApplication.AssignmentVersion++
	candidate.RAGApplication.AssignmentDigest = workflowRAGSHA256("replacement assignment")
	candidate.RAGApplication.BindingRef = WorkflowRAGApplicationBindingRef{BindingID: "wragb_bbbbbbbbbbbbbbbb", BindingVersion: 1, BindingDigest: workflowRAGSHA256("replacement binding")}
	candidate.RAGSnapshot.SnapshotID = "rags_bbbbbbbbbbbbbbbb"
	candidate.RAGSnapshot.SnapshotVersion = 2
	candidate.RAGSnapshot.SnapshotDigest = workflowRAGSHA256("promoted candidate snapshot")
	candidate.RAGApplication.CandidateSnapshot.SnapshotID = candidate.RAGSnapshot.SnapshotID
	candidate.RAGApplication.CandidateSnapshot.SnapshotVersion = candidate.RAGSnapshot.SnapshotVersion
	candidate.RAGApplication.CandidateSnapshot.SnapshotDigest = candidate.RAGSnapshot.SnapshotDigest
	candidate.RAGApplication.CandidateSnapshot.RAGRef = candidate.RAGSnapshot.RAGRef
	if err := validateWorkflowRunStoreRecord(workflowRAGApplicationRunContext(fixture.runtimeContext), &candidate); err != nil {
		t.Fatalf("candidate authority fixture is not a valid v4 record: %v", err)
	}
	if !workflowRAGRunsComparable(baseline, candidate) {
		t.Fatal("v4 comparison rejected the same input and answer contract after reviewed authority replacement")
	}
	comparison := buildWorkflowRAGRunComparison(baseline, candidate, fixture.now.Add(time.Second))
	if comparison.SchemaVersion != workflowRAGAppRunComparisonSchemaVersion || comparison.Retrieval == nil || !comparison.Retrieval.AuthorityChanged || comparison.Retrieval.BaselineAuthority == nil || comparison.Retrieval.CandidateAuthority == nil || comparison.Retrieval.CandidateAuthority.SnapshotID != candidate.RAGSnapshot.SnapshotID || comparison.Retrieval.SnapshotID != "" {
		t.Fatalf("v4 authority comparison evidence drifted: %#v", comparison)
	}
	unchangedCandidate := cloneWorkflowRunRecord(baseline)
	unchangedCandidate.RunID = "run_applicationragcompare03"
	unchanged := buildWorkflowRAGRunComparison(baseline, unchangedCandidate, fixture.now.Add(time.Second))
	payload, err := json.Marshal(unchanged)
	if err != nil {
		t.Fatalf("marshal unchanged v4 comparison: %v", err)
	}
	if !strings.Contains(string(payload), `"authority_changed":false`) {
		t.Fatalf("v4 comparison omitted explicit unchanged authority projection: %s", payload)
	}
	candidate.RAGApplication.ConfiguredProtocol = "different-answer-contract"
	if workflowRAGRunsComparable(baseline, candidate) {
		t.Fatal("v4 comparison accepted different configured answer contracts")
	}
}

func TestWorkflowRAGApplicationInvocationFailsBeforeSideEffectsOnAuthorityDrift(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	activated := fixture.runtimeService.Decide(fixture.runtimeContext, WorkflowRAGApplicationRuntimeDecisionInput{ExpectedRecordVersion: 0, Decision: workflowRAGApplicationRuntimeDecisionActivate, PublishCandidateID: fixture.publishCandidate.CandidateID, Reason: "激活后验证权威漂移失败关闭"})
	if activated.FailureCode != "" {
		t.Fatalf("activate before drift: %#v", activated)
	}
	fixture.promotionFixture.archiveDataset(t)
	service := newWorkflowRAGApplicationInvocationService(fixture.runtimeRepository, fixture.resolver, fixture.runStore, fixture.bridge)
	result := service.Invoke(fixture.runtimeContext, WorkflowRAGApplicationInvocationInput{Input: "authority drift query"})
	if result.FailureCode != WorkflowRAGApplicationFailureBindingNotEligible || result.Run != nil || fixture.bridge.callCount() != 0 {
		t.Fatalf("authority drift did not fail before side effects: %#v calls=%d", result, fixture.bridge.callCount())
	}
	fixture.runStore.mu.RLock()
	defer fixture.runStore.mu.RUnlock()
	if len(fixture.runStore.records) != 0 {
		t.Fatalf("precondition failure created %d runs", len(fixture.runStore.records))
	}
}

func TestWorkflowRAGApplicationRuntimeHTTPManagementAndAPIKeyInvocationBoundaries(t *testing.T) {
	fixture := newWorkflowRAGApplicationRuntimeTestFixture(t)
	applicationRepository := newMemoryApplicationCatalogRepository()
	apiKeyRepository := newMemoryAPIKeyRepository()
	seedWorkflowRAGApplicationRuntimeCatalogRecord(t, applicationRepository, fixture)
	invokeToken := seedWorkflowRAGApplicationRuntimeAPIKey(t, apiKeyRepository, fixture, "key_aaaaaaaaaaaaaaaa", []string{"application_rag:invoke"})
	chatOnlyToken := seedWorkflowRAGApplicationRuntimeAPIKey(t, apiKeyRepository, fixture, "key_bbbbbbbbbbbbbbbb", []string{"chat:invoke"})
	server := &Server{
		config:                                 config.Config{WorkflowRAGAppInvocationDevEnabled: true, ApplicationCatalogDevHTTPEnabled: true, Provider: "mock"},
		bridge:                                 fixture.bridge,
		applicationCatalogRepository:           applicationRepository,
		apiKeyRepository:                       apiKeyRepository,
		applicationDraftRepository:             fixture.promotionFixture.drafts,
		applicationPublishCandidateRepository:  fixture.publishRepository,
		workflowRunStore:                       fixture.runStore,
		workflowRAGSnapshotRepository:          fixture.promotionFixture.snapshots,
		workflowRAGEvaluationDatasetRepository: fixture.promotionFixture.evaluations,
		workflowRAGPromotionRepository:         fixture.promotionFixture.promotions,
		workflowRAGAppRuntimeRepository:        fixture.runtimeRepository,
	}
	if controlPlaneReadPermissionGrants["radishmind.workflow-rag-runtime.read"] != "workflow_rag_runtime:read" || controlPlaneReadPermissionGrants["radishmind.workflow-rag-runtime.write"] != "workflow_rag_runtime:write" {
		t.Fatalf("runtime management permission projection drifted: %#v", controlPlaneReadPermissionGrants)
	}

	decisionRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+fixture.runtimeContext.ApplicationID+"/workflow-rag-runtime-assignment/decisions", strings.NewReader(`{"workspace_id":"workspace_demo","expected_record_version":0,"decision":"activate","publish_candidate_id":"candidate-rag-runtime-v2","reason":"人工通过管理 API 激活已批准候选"}`))
	decisionRequest.SetPathValue("application_id", fixture.runtimeContext.ApplicationID)
	decisionRequest.Header.Set(savedWorkflowDraftDevWorkspaceHeader, fixture.runtimeContext.WorkspaceID)
	decisionRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, fixture.runtimeContext.ApplicationID)
	decisionRequest = decisionRequest.WithContext(withControlPlaneReadFakeAuthContext(decisionRequest.Context(), workflowRAGApplicationRuntimeHTTPAuth(fixture, "workflow_rag_runtime:write")))
	decisionRecorder := httptest.NewRecorder()
	server.handleDecideWorkflowRAGApplicationRuntimeAssignment(decisionRecorder, decisionRequest)
	var decisionEnvelope workflowRAGApplicationRuntimeEnvelope
	if decisionRecorder.Code != http.StatusOK || json.Unmarshal(decisionRecorder.Body.Bytes(), &decisionEnvelope) != nil || decisionEnvelope.FailureCode != nil || decisionEnvelope.Assignment == nil || decisionEnvelope.Assignment.State != workflowRAGApplicationRuntimeStateActive {
		t.Fatalf("activate assignment over management API: status=%d body=%s", decisionRecorder.Code, decisionRecorder.Body.String())
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications/"+fixture.runtimeContext.ApplicationID+"/workflow-rag-runtime-assignment?workspace_id="+fixture.runtimeContext.WorkspaceID, nil)
	readRequest.SetPathValue("application_id", fixture.runtimeContext.ApplicationID)
	readRequest.Header.Set(savedWorkflowDraftDevWorkspaceHeader, fixture.runtimeContext.WorkspaceID)
	readRequest.Header.Set(savedWorkflowDraftDevApplicationHeader, fixture.runtimeContext.ApplicationID)
	readRequest = readRequest.WithContext(withControlPlaneReadFakeAuthContext(readRequest.Context(), workflowRAGApplicationRuntimeHTTPAuth(fixture, "workflow_rag_runtime:read")))
	readRecorder := httptest.NewRecorder()
	server.handleReadWorkflowRAGApplicationRuntimeAssignment(readRecorder, readRequest)
	if readRecorder.Code != http.StatusOK || !strings.Contains(readRecorder.Body.String(), `"state":"active"`) || strings.Contains(readRecorder.Body.String(), "promotion authority fragment") {
		t.Fatalf("read assignment over management API: status=%d body=%s", readRecorder.Code, readRecorder.Body.String())
	}

	deniedRequest := httptest.NewRequest(http.MethodPost, workflowRAGApplicationInvocationRoute, strings.NewReader(`{"input":"scope denied query"}`))
	deniedRequest.Header.Set("Authorization", "Bearer "+chatOnlyToken)
	deniedRecorder := httptest.NewRecorder()
	server.handleWorkflowRAGApplicationInvocation(deniedRecorder, deniedRequest)
	if deniedRecorder.Code != http.StatusForbidden || !strings.Contains(deniedRecorder.Body.String(), APIKeyFailureScopeDenied) || fixture.bridge.callCount() != 0 {
		t.Fatalf("chat-only API key crossed application RAG scope: status=%d body=%s calls=%d", deniedRecorder.Code, deniedRecorder.Body.String(), fixture.bridge.callCount())
	}

	unknownRequest := httptest.NewRequest(http.MethodPost, workflowRAGApplicationInvocationRoute, strings.NewReader(`{"input":"strict query","model":"client-choice-forbidden"}`))
	unknownRequest.Header.Set("Authorization", "Bearer "+invokeToken)
	unknownRecorder := httptest.NewRecorder()
	server.handleWorkflowRAGApplicationInvocation(unknownRecorder, unknownRequest)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), "INVALID_JSON") || fixture.bridge.callCount() != 0 {
		t.Fatalf("invocation accepted client authority field: status=%d body=%s calls=%d", unknownRecorder.Code, unknownRecorder.Body.String(), fixture.bridge.callCount())
	}

	successRequest := httptest.NewRequest(http.MethodPost, workflowRAGApplicationInvocationRoute, strings.NewReader(`{"input":"promotion authority query"}`))
	successRequest.Header.Set("Authorization", "Bearer "+invokeToken)
	successRecorder := httptest.NewRecorder()
	server.handleWorkflowRAGApplicationInvocation(successRecorder, successRequest)
	var successEnvelope workflowRAGApplicationInvocationEnvelope
	if successRecorder.Code != http.StatusOK || json.Unmarshal(successRecorder.Body.Bytes(), &successEnvelope) != nil || successEnvelope.FailureCode != nil || successEnvelope.Run == nil || successEnvelope.Answer == nil || successEnvelope.Run.SideEffects.RetrievalCalls != 1 || successEnvelope.Run.SideEffects.ProviderCalls != 1 || fixture.bridge.callCount() != 1 {
		t.Fatalf("application RAG invocation over API key route: status=%d body=%s calls=%d", successRecorder.Code, successRecorder.Body.String(), fixture.bridge.callCount())
	}
	if strings.Contains(successRecorder.Body.String(), `"input":"promotion authority query"`) || strings.Contains(successRecorder.Body.String(), "promotion authority fragment") || strings.Contains(successRecorder.Body.String(), invokeToken) {
		t.Fatalf("application invocation response leaked protected material: %s", successRecorder.Body.String())
	}
}

func workflowRAGApplicationRuntimeHTTPAuth(fixture *workflowRAGApplicationRuntimeTestFixture, scopes ...string) controlPlaneReadAuthContext {
	return controlPlaneReadAuthContext{
		AuthMode:        controlPlaneReadAuthModeDevHeaders,
		IdentityContext: "dev:workflow-rag-runtime-test",
		TenantBinding:   fixture.runtimeContext.TenantRef,
		SubjectBinding:  fixture.runtimeContext.ActorRef,
		ScopeGrants:     scopes,
		AuditContext:    "audit_workflow_rag_runtime_http",
		VerifiedIdentity: &VerifiedControlPlaneIdentity{
			SubjectRef: fixture.runtimeContext.ActorRef,
			TenantRef:  fixture.runtimeContext.TenantRef,
		},
		ResourceBinding: ControlPlaneResourceBinding{TenantRef: fixture.runtimeContext.TenantRef, TenantVerified: true},
	}
}

func seedWorkflowRAGApplicationRuntimeCatalogRecord(t *testing.T, repository applicationCatalogRepository, fixture *workflowRAGApplicationRuntimeTestFixture) {
	t.Helper()
	ctx := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_runtime_catalog", TenantRef: fixture.runtimeContext.TenantRef, WorkspaceID: fixture.runtimeContext.WorkspaceID, ActorRef: fixture.runtimeContext.ActorRef, OwnerSubjectRef: fixture.runtimeContext.OwnerSubjectRef, AuditRef: "audit_runtime_catalog"}
	record := ApplicationCatalogRecord{SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: fixture.runtimeContext.ApplicationID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, OwnerSubjectRef: ctx.OwnerSubjectRef, DisplayName: fixture.promotionFixture.application.DisplayName, ApplicationKind: fixture.promotionFixture.application.ApplicationKind, LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1, CreatedAt: fixture.promotionFixture.application.UpdatedAt, UpdatedAt: fixture.promotionFixture.application.UpdatedAt, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	if _, err := repository.Create(ctx, record); err != nil {
		t.Fatalf("seed runtime application catalog: %v", err)
	}
}

func seedWorkflowRAGApplicationRuntimeAPIKey(t *testing.T, repository apiKeyRepository, fixture *workflowRAGApplicationRuntimeTestFixture, apiKeyID string, scopes []string) string {
	t.Helper()
	token, digest, err := newAPIKeyCredential(apiKeyID)
	if err != nil {
		t.Fatalf("create runtime API key credential: %v", err)
	}
	now := time.Now().UTC()
	ctx := APIKeyContext{RequestContext: context.Background(), RequestID: "request_runtime_api_key", TenantRef: fixture.runtimeContext.TenantRef, WorkspaceID: fixture.runtimeContext.WorkspaceID, ActorRef: fixture.runtimeContext.ActorRef, OwnerSubjectRef: fixture.runtimeContext.OwnerSubjectRef, AuditRef: "audit_runtime_api_key", WriteEnabled: true}
	record := APIKeyRecord{SchemaVersion: apiKeyRecordSchemaVersion, APIKeyID: apiKeyID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: fixture.runtimeContext.ApplicationID, OwnerSubjectRef: ctx.OwnerSubjectRef, DisplayName: "Application RAG invocation key", Scopes: scopes, LifecycleState: apiKeyLifecycleActive, EffectiveState: apiKeyLifecycleActive, RecordVersion: 1, CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(24 * time.Hour).Format(time.RFC3339Nano), CreatedByActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef, credentialDigest: digest}
	if _, err := repository.Create(ctx, record); err != nil {
		t.Fatalf("seed runtime API key record: %v", err)
	}
	return token
}

func newWorkflowRAGApplicationRuntimeTestFixture(t *testing.T) *workflowRAGApplicationRuntimeTestFixture {
	t.Helper()
	promotionFixture, approval := approvedWorkflowRAGPromotionFixture(t)
	draftService, draftContext := workflowRAGBindingDraftService(promotionFixture)
	payload := promotionFixture.draft.ApplicationConfigurationDraftPayload
	payload.SchemaVersion = applicationConfigurationDraftSchemaVersionV2
	payload.WorkflowRAGBindingRef = cloneWorkflowRAGApplicationBindingRef(&approval.Binding.WorkflowRAGApplicationBindingRef)
	attached := draftService.Save(draftContext, payload, promotionFixture.draft.DraftVersion)
	if attached.FailureCode != "" || attached.Draft == nil {
		t.Fatalf("attach binding for runtime fixture: %#v", attached)
	}
	publishRepository := newMemoryApplicationPublishCandidateRepository()
	publishService := newApplicationPublishCandidateService(promotionFixture.drafts, publishRepository, func(ApplicationPublishContext) (ApplicationSummary, error) {
		return promotionFixture.application, promotionFixture.applicationErr
	})
	publishService.validateBinding = func(ctx ApplicationPublishContext, ref WorkflowRAGApplicationBindingRef) (WorkflowRAGApplicationBinding, string) {
		return promotionFixture.service.resolveEligibleBinding(workflowRAGPromotionContextFromPublish(ctx), ref, false)
	}
	publishService.now = func() time.Time { return promotionFixture.now.Add(time.Hour) }
	publishContext := workflowRAGBindingPublishContext(promotionFixture)
	created := publishService.Create(publishContext, ApplicationPublishCreateInput{CandidateID: "candidate-rag-runtime-v2", DraftID: attached.Draft.DraftID, ExpectedDraftVersion: attached.Draft.DraftVersion})
	if created.FailureCode != "" || created.Candidate == nil {
		t.Fatalf("create runtime publish candidate: %#v", created)
	}
	approved := publishService.Review(publishContext, created.Candidate.CandidateID, ApplicationPublishReviewInput{ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "人工批准开发测试态运行候选"})
	if approved.FailureCode != "" || approved.Candidate == nil {
		t.Fatalf("approve runtime publish candidate: %#v", approved)
	}
	runtimeRepository, err := newWorkflowRAGApplicationRuntimeRepositoryForRunStore(promotionFixture.runStore)
	if err != nil {
		t.Fatalf("create runtime repository: %v", err)
	}
	runtimeContext := WorkflowRAGApplicationRuntimeContext{RequestContext: context.Background(), RequestID: "request_runtime_activate", TenantRef: promotionFixture.ctx.TenantRef, WorkspaceID: promotionFixture.ctx.WorkspaceID, ApplicationID: promotionFixture.ctx.ApplicationID, ActorRef: promotionFixture.ctx.ActorRef, OwnerSubjectRef: promotionFixture.ctx.OwnerSubjectRef, AuditRef: "audit_runtime_activate", WriteEnabled: true}
	resolver := workflowRAGApplicationAuthorityResolver{publishRepository: publishRepository, draftRepository: promotionFixture.drafts, promotionService: promotionFixture.service, readApplication: promotionFixture.service.readApplication}
	runtimeService := newWorkflowRAGApplicationRuntimeService(runtimeRepository, resolver)
	runtimeService.newID = workflowRAGEvaluationTestIDGenerator()
	now := time.Date(2026, 7, 19, 9, 0, 0, 0, time.UTC)
	runtimeService.now = func() time.Time { return now }
	bridgeClient := &workflowExecutorTestBridge{handle: func(context.Context, []byte, bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error) {
		return bridge.GatewayEnvelope{Status: "ok", Response: map[string]any{"structured_answer": map[string]any{"schema_version": workflowRAGApplicationAnswerSchemaVersion, "answer": "Use the approved promotion authority guidance.", "citations": []map[string]any{{"fragment_ref": "promotion_authority", "claim_summary": "Approved authority evidence supports the answer."}}, "limitations": []string{}, "confidence": "high"}}}, nil
	}}
	return &workflowRAGApplicationRuntimeTestFixture{promotionFixture: promotionFixture, runtimeContext: runtimeContext, publishRepository: publishRepository, runtimeRepository: runtimeRepository.(*memoryWorkflowRAGApplicationRuntimeRepository), runtimeService: runtimeService, resolver: resolver, runStore: promotionFixture.runStore, publishCandidate: *approved.Candidate, bridge: bridgeClient, now: now}
}

func (fixture *workflowRAGApplicationRuntimeTestFixture) advanceRuntimeRequest(suffix string) {
	fixture.runtimeContext.RequestID = "request_runtime_" + suffix
	fixture.runtimeContext.AuditRef = "audit_runtime_" + suffix
}
