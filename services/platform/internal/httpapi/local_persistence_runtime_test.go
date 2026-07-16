package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestSQLiteDevAggregateServerRestartRestoresAllRepositoryData(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	firstServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-aggregate-first"})
	if err != nil {
		t.Fatalf("start first aggregate SQLite server: %v", err)
	}
	if firstServer.localPersistenceRuntime == nil || firstServer.localPersistenceRuntime.DB() == nil {
		t.Fatal("aggregate SQLite server did not retain the shared runtime")
	}
	assertAggregateSQLiteRepositorySelection(t, firstServer)
	if err := firstServer.localPersistenceRuntime.VerifyMigrations(context.Background(), localPersistenceSQLiteMigrations()); err != nil {
		t.Fatalf("verify aggregate SQLite migrations: %v", err)
	}
	var migrationCount int
	if err := firstServer.localPersistenceRuntime.DB().QueryRowContext(
		context.Background(),
		"SELECT count(*) FROM radishmind_schema_migrations",
	).Scan(&migrationCount); err != nil || migrationCount != 9 {
		t.Fatalf("aggregate SQLite migration count drifted: count=%d err=%v", migrationCount, err)
	}

	catalogContext := applicationCatalogTestContext("subject_owner")
	catalogResult := newApplicationCatalogService(firstServer.applicationCatalogRepository).Create(catalogContext, ApplicationCatalogCreateInput{
		DisplayName: "Aggregate SQLite application", Description: "Persists across the platform lifecycle.", ApplicationKind: "agent",
	})
	if catalogResult.FailureCode != "" || catalogResult.Record == nil {
		t.Fatalf("create aggregate SQLite application: %#v", catalogResult)
	}

	draftContext := validApplicationDraftContext()
	draftResult := newApplicationConfigurationDraftService(firstServer.applicationDraftRepository).Save(
		draftContext,
		validApplicationDraftPayload(),
		0,
	)
	if draftResult.FailureCode != "" || draftResult.Draft == nil {
		t.Fatalf("save aggregate SQLite application draft: %#v", draftResult)
	}

	publishContext := validApplicationPublishContext()
	publishResult := newApplicationPublishCandidateService(
		firstServer.applicationDraftRepository,
		firstServer.applicationPublishCandidateRepository,
		validApplicationPublishBaseline,
	).Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "candidate-aggregate-restart", DraftID: draftResult.Draft.DraftID, ExpectedDraftVersion: 1,
		EvidenceRequestIDs: []string{"request-aggregate-evidence"},
	})
	if publishResult.FailureCode != "" || publishResult.Candidate == nil {
		t.Fatalf("create aggregate SQLite publish candidate: %#v", publishResult)
	}

	apiKeyContext := apiKeyTestContext("subject_owner")
	apiKeyResult := newAPIKeyService(firstServer.apiKeyRepository, firstServer.applicationCatalogRepository).Create(
		apiKeyContext,
		validAPIKeyCreateInput(catalogResult.Record.ApplicationID, 30),
	)
	if apiKeyResult.FailureCode != "" || apiKeyResult.Record == nil || apiKeyResult.CredentialToken == "" {
		t.Fatalf("create aggregate SQLite API key: failure=%s record_present=%t token_present=%t",
			apiKeyResult.FailureCode, apiKeyResult.Record != nil, apiKeyResult.CredentialToken != "")
	}

	gatewayContext := gatewayRequestTestContext()
	gatewayRecord := gatewayRequestTestRecord(gatewayContext, "request_aggregate_restart", time.Now().UTC())
	if err := firstServer.gatewayRequestHistoryStore.CreateRequest(gatewayContext, &gatewayRecord); err != nil {
		t.Fatalf("create aggregate SQLite Gateway request: %v", err)
	}

	savedDraftContext := savedWorkflowDraftSQLiteContext()
	savedDraftPayload := validSavedWorkflowDraftPayload()
	savedDraftResult := newSavedWorkflowDraftService(firstServer.savedWorkflowDraftStore).SaveDraft(
		savedDraftContext,
		SaveWorkflowDraftRequest{Payload: savedDraftPayload},
	)
	if savedDraftResult.FailureCode != "" || savedDraftResult.Draft == nil {
		t.Fatalf("save aggregate SQLite workflow draft: %#v", savedDraftResult)
	}

	runContext := workflowExecutorTestContext()
	runRecord := workflowRunHistoryTestRecord(runContext, "run_aggregate_restart", savedDraftPayload.DraftID, time.Now().UTC())
	if err := firstServer.workflowRunStore.UpsertRun(runContext, &runRecord); err != nil {
		t.Fatalf("create aggregate SQLite workflow run: %v", err)
	}

	firstRuntime := firstServer.localPersistenceRuntime
	firstServer.Close()
	firstServer.Close()
	if err := firstRuntime.DB().PingContext(context.Background()); err == nil {
		t.Fatal("aggregate SQLite runtime remained usable after Server.Close")
	}

	secondServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-aggregate-second"})
	if err != nil {
		t.Fatalf("restart aggregate SQLite server: %v", err)
	}
	t.Cleanup(secondServer.Close)
	assertAggregateSQLiteRepositorySelection(t, secondServer)

	restoredCatalog := newApplicationCatalogService(secondServer.applicationCatalogRepository).Read(catalogContext, catalogResult.Record.ApplicationID)
	if restoredCatalog.FailureCode != "" || restoredCatalog.Record == nil || restoredCatalog.Record.DisplayName != catalogResult.Record.DisplayName {
		t.Fatalf("restore aggregate SQLite application: %#v", restoredCatalog)
	}
	restoredDraft := newApplicationConfigurationDraftService(secondServer.applicationDraftRepository).Read(draftContext, draftResult.Draft.DraftID)
	if restoredDraft.FailureCode != "" || restoredDraft.Draft == nil || restoredDraft.Draft.DraftVersion != 1 {
		t.Fatalf("restore aggregate SQLite application draft: %#v", restoredDraft)
	}
	restoredPublish := newApplicationPublishCandidateService(
		secondServer.applicationDraftRepository,
		secondServer.applicationPublishCandidateRepository,
		validApplicationPublishBaseline,
	).Read(publishContext, publishResult.Candidate.CandidateID)
	if restoredPublish.FailureCode != "" || restoredPublish.Candidate == nil || restoredPublish.Candidate.ReviewVersion != 0 {
		t.Fatalf("restore aggregate SQLite publish candidate: %#v", restoredPublish)
	}
	restoredAPIKey := newAPIKeyService(secondServer.apiKeyRepository, secondServer.applicationCatalogRepository).Read(apiKeyContext, apiKeyResult.Record.APIKeyID)
	if restoredAPIKey.FailureCode != "" || restoredAPIKey.Record == nil || restoredAPIKey.Record.ApplicationID != catalogResult.Record.ApplicationID {
		t.Fatalf("restore aggregate SQLite API key: %#v", restoredAPIKey)
	}
	restoredGateway, found, err := secondServer.gatewayRequestHistoryStore.ReadRequest(gatewayContext, gatewayRecord.RequestID)
	if err != nil || !found || restoredGateway.RecordVersion != 1 {
		t.Fatalf("restore aggregate SQLite Gateway request: found=%t record=%#v err=%v", found, restoredGateway, err)
	}
	restoredSavedDraft := newSavedWorkflowDraftService(secondServer.savedWorkflowDraftStore).ReadDraft(
		savedDraftContext,
		ReadWorkflowDraftRequest{DraftID: savedDraftPayload.DraftID},
	)
	if restoredSavedDraft.FailureCode != "" || restoredSavedDraft.Draft == nil || restoredSavedDraft.Draft.DraftVersion != 1 {
		t.Fatalf("restore aggregate SQLite workflow draft: %#v", restoredSavedDraft)
	}
	restoredRun, found, err := secondServer.workflowRunStore.ReadRun(runContext, runRecord.RunID)
	if err != nil || !found || restoredRun.RecordVersion != 1 || restoredRun.Status != WorkflowRunStatusRunning {
		t.Fatalf("restore aggregate SQLite workflow run: found=%t record=%#v err=%v", found, restoredRun, err)
	}
}

func TestSQLiteDevLocalProductHTTPChainSurvivesServerRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	cfg.GatewayAuthMode = gatewayAPIKeyAuthenticationSource
	cfg.Model = "platform-model"

	firstServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-local-product-first"})
	if err != nil {
		t.Fatalf("start first local product server: %v", err)
	}
	t.Cleanup(firstServer.Close)
	firstBridge := &workflowExecutorTestBridge{}
	firstServer.bridge = firstBridge

	createApplicationRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/applications",
		bytes.NewReader(mustLocalProductJSON(t, applicationCatalogCreateBody{
			WorkspaceID: "workspace_demo", DisplayName: "SQLite local product", Description: "Continuous local product chain.", ApplicationKind: "workflow_copilot",
		})),
	)
	setLocalProductControlHeaders(createApplicationRequest, "applications:read,applications:write")
	createApplicationResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(createApplicationResponse, createApplicationRequest)
	var createdApplication applicationCatalogEnvelope
	decodeLocalProductResponse(t, createApplicationResponse, http.StatusOK, &createdApplication)
	if createdApplication.FailureCode != nil || createdApplication.Record == nil {
		t.Fatalf("create local product application: %#v", createdApplication)
	}
	applicationID := createdApplication.Record.ApplicationID

	applicationDraft := validApplicationDraftPayload()
	applicationDraft.DraftID = "app-config-local-product-chain"
	applicationDraft.ApplicationID = applicationID
	applicationDraft.BaseApplicationUpdatedAt = createdApplication.Record.UpdatedAt
	createdApplicationDraft := performApplicationDraftRequest(
		t,
		firstServer,
		http.MethodPost,
		"/v1/user-workspace/application-drafts",
		applicationConfigurationDraftSaveBody{Draft: applicationDraft},
		"application_drafts:read,application_drafts:write",
		applicationID,
	)
	if createdApplicationDraft.FailureCode != nil || createdApplicationDraft.Draft == nil || createdApplicationDraft.CurrentDraftVersion != 1 {
		t.Fatalf("save local product application draft: %#v", createdApplicationDraft)
	}

	createdCandidate := performApplicationPublishRequest(
		t,
		firstServer,
		http.MethodPost,
		"/v1/user-workspace/application-publish-candidates",
		applicationPublishCandidateCreateBody{
			CandidateID: "candidate-local-product-chain", DraftID: applicationDraft.DraftID, ExpectedDraftVersion: 1,
		},
		"application_publish_candidates:write",
		applicationID,
	)
	if createdCandidate.FailureCode != nil || createdCandidate.Candidate == nil {
		t.Fatalf("create local product publish candidate: %#v", createdCandidate)
	}
	approvedCandidate := performApplicationPublishRequest(
		t,
		firstServer,
		http.MethodPost,
		"/v1/user-workspace/application-publish-candidates/candidate-local-product-chain/reviews",
		applicationPublishCandidateReviewBody{
			ExpectedReviewVersion: 0, Decision: "approve", Reason: "Approved through the local SQLite product chain.",
		},
		"application_publish_candidates:review",
		applicationID,
	)
	if approvedCandidate.FailureCode != nil || approvedCandidate.Candidate == nil || approvedCandidate.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("approve local product publish candidate: %#v", approvedCandidate)
	}

	issueKeyRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/api-keys",
		bytes.NewReader(mustLocalProductJSON(t, apiKeyCreateBody{
			WorkspaceID: "workspace_demo", ApplicationID: applicationID, DisplayName: "Local product Gateway key",
			Scopes: []string{"models:read", "chat:invoke"}, ExpiresInDays: 30,
		})),
	)
	setLocalProductControlHeaders(issueKeyRequest, "api_keys:read,api_keys:write")
	issueKeyResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(issueKeyResponse, issueKeyRequest)
	var issuedKey apiKeyEnvelope
	decodeLocalProductResponse(t, issueKeyResponse, http.StatusCreated, &issuedKey)
	if issuedKey.FailureCode != nil || issuedKey.Record == nil || issuedKey.Credential == nil || issuedKey.Credential.Token == "" {
		t.Fatalf("issue local product API key: record=%#v credential_present=%t", issuedKey.Record, issuedKey.Credential != nil)
	}

	rawGatewayInput := "private Gateway prompt must not persist"
	gatewayRequestID := "request-local-product-chain"
	gatewayRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		bytes.NewReader(mustLocalProductJSON(t, map[string]any{
			"model": "platform-model", "messages": []map[string]string{{"role": "user", "content": rawGatewayInput}},
		})),
	)
	gatewayRequest.Header.Set("Authorization", "Bearer "+issuedKey.Credential.Token)
	gatewayRequest.Header.Set("X-Request-Id", gatewayRequestID)
	gatewayResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(gatewayResponse, gatewayRequest)
	if gatewayResponse.Code != http.StatusOK {
		t.Fatalf("call local product Gateway with issued key: %d %s", gatewayResponse.Code, gatewayResponse.Body.String())
	}

	workflowDraft := validSavedWorkflowDraftPayload()
	workflowDraft.DraftID = "draft_local_product_chain"
	workflowDraft.ApplicationID = applicationID
	saveWorkflowRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts",
		bytes.NewReader(mustSavedWorkflowDraftJSON(t, savedWorkflowDraftSaveHTTPBody{
			Draft: savedWorkflowDraftPayloadDocumentFromDraftPayload(workflowDraft),
		})),
	)
	setLocalProductWorkflowHeaders(saveWorkflowRequest, "workflow_drafts:read,workflow_drafts:write", applicationID)
	saveWorkflowResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(saveWorkflowResponse, saveWorkflowRequest)
	savedWorkflow := decodeSavedWorkflowDraftEnvelope(t, saveWorkflowResponse, http.StatusOK)
	if savedWorkflow.FailureCode != nil || savedWorkflow.Draft == nil || savedWorkflow.CurrentDraftVersion != 1 {
		t.Fatalf("save local product workflow draft: %#v", savedWorkflow)
	}

	rawWorkflowInput := "private workflow run input must not persist"
	startRunRequest := httptest.NewRequest(
		http.MethodPost,
		"/v1/user-workspace/workflow-drafts/"+workflowDraft.DraftID+"/runs",
		bytes.NewReader(mustWorkflowRunJSON(t, workflowRunStartHTTPBody{
			WorkspaceID: "workspace_demo", ApplicationID: applicationID, InputText: rawWorkflowInput, Model: "platform-model",
		})),
	)
	setLocalProductWorkflowHeaders(startRunRequest, "workflow_drafts:read,workflow_runs:execute,workflow_runs:read", applicationID)
	startRunResponse := httptest.NewRecorder()
	firstServer.httpServer.Handler.ServeHTTP(startRunResponse, startRunRequest)
	startedRun := decodeWorkflowRunEnvelope(t, startRunResponse, http.StatusOK)
	if startedRun.FailureCode != nil || startedRun.Run == nil || startedRun.Run.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("start local product workflow run: %#v", startedRun)
	}
	if firstBridge.callCount() != 2 {
		t.Fatalf("local product chain should call the bridge once for Gateway and once for Workflow, got %d", firstBridge.callCount())
	}

	firstServer.Close()

	restartedServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-local-product-restarted"})
	if err != nil {
		t.Fatalf("restart local product server: %v", err)
	}
	t.Cleanup(restartedServer.Close)
	restartedServer.bridge = &workflowExecutorTestBridge{}

	readApplicationRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/applications/"+applicationID+"?workspace_id=workspace_demo",
		nil,
	)
	setLocalProductControlHeaders(readApplicationRequest, "applications:read")
	readApplicationResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readApplicationResponse, readApplicationRequest)
	var restoredApplication applicationCatalogEnvelope
	decodeLocalProductResponse(t, readApplicationResponse, http.StatusOK, &restoredApplication)
	if restoredApplication.FailureCode != nil || restoredApplication.Record == nil || restoredApplication.Record.ApplicationID != applicationID {
		t.Fatalf("restore local product application over HTTP: %#v", restoredApplication)
	}

	restoredApplicationDraft := performApplicationDraftRequest(
		t,
		restartedServer,
		http.MethodGet,
		"/v1/user-workspace/application-drafts/"+applicationDraft.DraftID+"?workspace_id=workspace_demo&application_id="+applicationID,
		nil,
		"application_drafts:read",
		applicationID,
	)
	if restoredApplicationDraft.FailureCode != nil || restoredApplicationDraft.Draft == nil || restoredApplicationDraft.Draft.DraftVersion != 1 {
		t.Fatalf("restore local product application draft over HTTP: %#v", restoredApplicationDraft)
	}

	restoredCandidate := performApplicationPublishRequest(
		t,
		restartedServer,
		http.MethodGet,
		"/v1/user-workspace/application-publish-candidates/candidate-local-product-chain?workspace_id=workspace_demo&application_id="+applicationID,
		nil,
		"application_publish_candidates:read",
		applicationID,
	)
	if restoredCandidate.FailureCode != nil || restoredCandidate.Candidate == nil || restoredCandidate.Candidate.CandidateState != applicationPublishStateApproved {
		t.Fatalf("restore local product publish candidate over HTTP: %#v", restoredCandidate)
	}

	readKeyRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/api-keys/"+issuedKey.Record.APIKeyID+"?workspace_id=workspace_demo",
		nil,
	)
	setLocalProductControlHeaders(readKeyRequest, "api_keys:read")
	readKeyResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readKeyResponse, readKeyRequest)
	var restoredKey apiKeyEnvelope
	decodeLocalProductResponse(t, readKeyResponse, http.StatusOK, &restoredKey)
	if restoredKey.FailureCode != nil || restoredKey.Record == nil || restoredKey.Record.LastUsedAt == nil || restoredKey.Credential != nil {
		t.Fatalf("restore used local product API key without credential: %#v", restoredKey)
	}

	gatewayQuery := url.Values{
		"workspace_id": {"workspace_demo"}, "consumer_ref": {"api_key:" + issuedKey.Record.APIKeyID},
		"application_id": {applicationID}, "status": {string(GatewayRequestStatusSucceeded)},
	}
	readGatewayRequest := httptest.NewRequest(http.MethodGet, "/v1/model-gateway/requests?"+gatewayQuery.Encode(), nil)
	setGatewayRequestDevHeaders(readGatewayRequest, "gateway_requests:read")
	readGatewayRequest.Header.Set(gatewayRequestDevConsumerHeader, "api_key:"+issuedKey.Record.APIKeyID)
	readGatewayRequest.Header.Set(gatewayRequestDevApplicationHeader, applicationID)
	readGatewayRequest.Header.Set(gatewayRequestDevSubjectHeader, "subject_demo_user")
	readGatewayResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readGatewayResponse, readGatewayRequest)
	restoredGatewayHistory := decodeGatewayRequestListEnvelope(t, readGatewayResponse)
	if restoredGatewayHistory.FailureCode != nil || len(restoredGatewayHistory.Requests) != 1 || restoredGatewayHistory.Requests[0].RequestID != gatewayRequestID {
		t.Fatalf("restore local product Gateway request history: %#v", restoredGatewayHistory)
	}

	readWorkflowRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-drafts/"+workflowDraft.DraftID+"?workspace_id=workspace_demo&application_id="+applicationID,
		nil,
	)
	setLocalProductWorkflowHeaders(readWorkflowRequest, "workflow_drafts:read", applicationID)
	readWorkflowResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readWorkflowResponse, readWorkflowRequest)
	restoredWorkflow := decodeSavedWorkflowDraftEnvelope(t, readWorkflowResponse, http.StatusOK)
	if restoredWorkflow.FailureCode != nil || restoredWorkflow.Draft == nil || restoredWorkflow.Draft.DraftVersion != 1 {
		t.Fatalf("restore local product workflow draft over HTTP: %#v", restoredWorkflow)
	}

	readRunRequest := httptest.NewRequest(
		http.MethodGet,
		"/v1/user-workspace/workflow-runs/"+startedRun.Run.RunID+"?workspace_id=workspace_demo&application_id="+applicationID,
		nil,
	)
	setLocalProductWorkflowHeaders(readRunRequest, "workflow_runs:read", applicationID)
	readRunResponse := httptest.NewRecorder()
	restartedServer.httpServer.Handler.ServeHTTP(readRunResponse, readRunRequest)
	restoredRun := decodeWorkflowRunEnvelope(t, readRunResponse, http.StatusOK)
	if restoredRun.FailureCode != nil || restoredRun.Run == nil || restoredRun.Run.Status != WorkflowRunStatusSucceeded {
		t.Fatalf("restore local product workflow run over HTTP: %#v", restoredRun)
	}

	restartedServer.Close()
	assertLocalProductSQLiteFilesExclude(t, databasePath, issuedKey.Credential.Token, rawGatewayInput, rawWorkflowInput)
}

func mustLocalProductJSON(t *testing.T, document any) []byte {
	t.Helper()
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal local product request: %v", err)
	}
	return encoded
}

func decodeLocalProductResponse(t *testing.T, response *httptest.ResponseRecorder, expectedStatus int, destination any) {
	t.Helper()
	if response.Code != expectedStatus {
		t.Fatalf("unexpected local product response status: expected=%d actual=%d body=%s", expectedStatus, response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), destination); err != nil {
		t.Fatalf("decode local product response: %v body=%s", err, response.Body.String())
	}
}

func setLocalProductControlHeaders(request *http.Request, scopes string) {
	setControlPlaneReadDevAuthHeaders(request)
	request.Header.Set(controlPlaneReadDevScopesHeader, scopes)
}

func setLocalProductWorkflowHeaders(request *http.Request, scopes, applicationID string) {
	setSavedWorkflowDraftDevHeaders(request, scopes)
	request.Header.Set(savedWorkflowDraftDevApplicationHeader, applicationID)
}

func assertLocalProductSQLiteFilesExclude(t *testing.T, databasePath string, forbidden ...string) {
	t.Helper()
	for _, path := range []string{databasePath, databasePath + "-wal", databasePath + "-shm"} {
		content, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatalf("read local product SQLite artifact %s: %v", filepath.Base(path), err)
		}
		for _, value := range forbidden {
			if value != "" && bytes.Contains(content, []byte(value)) {
				t.Fatalf("local product SQLite artifact %s contains forbidden runtime input", filepath.Base(path))
			}
		}
	}
}

func TestSQLiteDevAggregateRejectsMigrationMarkerMismatch(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	cfg := aggregateSQLiteDevServerConfig(databasePath)
	server, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-aggregate-marker-seed"})
	if err != nil {
		t.Fatalf("seed aggregate SQLite migrations: %v", err)
	}
	runtime := server.localPersistenceRuntime
	if _, err := runtime.DB().ExecContext(context.Background(), `UPDATE radishmind_schema_migrations
		SET migration_checksum='sha256:incompatible'
		WHERE component=? AND migration_id=?`,
		localPersistenceSQLiteMigrations()[0].Component,
		localPersistenceSQLiteMigrations()[0].ID,
	); err != nil {
		t.Fatalf("inject aggregate SQLite marker mismatch: %v", err)
	}
	server.Close()

	failedServer, err := NewServerWithError(cfg, Options{BuildVersion: "sqlite-aggregate-marker-mismatch"})
	if failedServer != nil || err == nil || err.Error() != "SQLite development migration marker mismatch" {
		t.Fatalf("aggregate SQLite marker mismatch did not fail closed: server=%v err=%v", failedServer, err)
	}
}

func TestSQLiteDevAggregateBridgeStartupFailureAllowsCleanRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	failingConfig := aggregateSQLiteDevServerConfig(databasePath)
	failingConfig.BridgeMode = "stdio_pool"
	failingConfig.BridgeHandshakeTimeout = 100 * time.Millisecond
	failingConfig.PythonBinary = filepath.Join(t.TempDir(), "missing-python")
	failedServer, err := NewServerWithError(failingConfig, Options{BuildVersion: "sqlite-aggregate-bridge-failure"})
	if failedServer != nil || err == nil {
		t.Fatalf("aggregate SQLite bridge startup failure did not stop construction: server=%v err=%v", failedServer, err)
	}

	restartedServer, err := NewServerWithError(
		aggregateSQLiteDevServerConfig(databasePath),
		Options{BuildVersion: "sqlite-aggregate-after-bridge-failure"},
	)
	if err != nil {
		t.Fatalf("aggregate SQLite runtime was not reusable after bridge startup rollback: %v", err)
	}
	assertAggregateSQLiteRepositorySelection(t, restartedServer)
	restartedServer.Close()
}

func assertAggregateSQLiteRepositorySelection(t *testing.T, server *Server) {
	t.Helper()
	if _, ok := server.applicationCatalogRepository.(*sqliteApplicationCatalogRepository); !ok {
		t.Fatalf("application catalog did not select SQLite: %T", server.applicationCatalogRepository)
	}
	if _, ok := server.applicationDraftRepository.(*sqliteApplicationConfigurationDraftRepository); !ok {
		t.Fatalf("application draft did not select SQLite: %T", server.applicationDraftRepository)
	}
	if _, ok := server.applicationPublishCandidateRepository.(*sqliteApplicationPublishCandidateRepository); !ok {
		t.Fatalf("application publish did not select SQLite: %T", server.applicationPublishCandidateRepository)
	}
	if _, ok := server.apiKeyRepository.(*sqliteAPIKeyRepository); !ok {
		t.Fatalf("API key did not select SQLite: %T", server.apiKeyRepository)
	}
	if _, ok := server.gatewayRequestHistoryStore.(*sqliteGatewayRequestStore); !ok {
		t.Fatalf("Gateway request did not select SQLite: %T", server.gatewayRequestHistoryStore)
	}
	if _, ok := server.savedWorkflowDraftStore.(*repositorySavedWorkflowDraftStore); !ok {
		t.Fatalf("saved workflow draft did not select the SQLite repository adapter: %T", server.savedWorkflowDraftStore)
	}
	if _, ok := server.workflowRunStore.(*sqliteWorkflowRunStore); !ok {
		t.Fatalf("workflow run did not select SQLite: %T", server.workflowRunStore)
	}
	if actionStore, ok := server.workflowHTTPToolActionStore.(*sqliteWorkflowHTTPToolActionStore); !ok || actionStore.database != server.localPersistenceRuntime.DB() {
		t.Fatalf("workflow tool actions did not share the SQLite runtime: %T", server.workflowHTTPToolActionStore)
	}
	for name, mode := range map[string]string{
		"application_catalog": server.config.ApplicationCatalogStoreMode,
		"application_draft":   server.config.ApplicationDraftStoreMode,
		"application_publish": server.config.ApplicationPublishStoreMode,
		"api_key":             server.config.APIKeyStoreMode,
		"gateway_request":     server.config.GatewayRequestStoreMode,
		"workflow_draft":      server.config.WorkflowSavedDraftStoreMode,
		"workflow_run":        server.config.WorkflowRunStoreMode,
	} {
		if mode != "sqlite_dev" {
			t.Fatalf("aggregate store mode drifted: component=%s mode=%s", name, mode)
		}
	}
}

func aggregateSQLiteDevServerConfig(databasePath string) config.Config {
	return config.Config{
		ListenAddr:                        ":0",
		ReadHeaderTimeout:                 time.Second,
		WriteTimeout:                      time.Second,
		BridgeTimeout:                     time.Second,
		BridgeMode:                        "process_per_request",
		BridgeWorkerCount:                 1,
		BridgeQueueCapacity:               1,
		BridgeHandshakeTimeout:            time.Second,
		PythonBinary:                      "python3",
		BridgeScript:                      "scripts/run-platform-bridge.py",
		Provider:                          "mock",
		ControlPlaneReadDevAuthEnabled:    true,
		ControlPlaneReadStoreMode:         "fake_store_dev",
		ControlPlaneReadDatabaseTimeout:   time.Second,
		WorkflowSavedDraftDevHTTPEnabled:  true,
		WorkflowSavedDraftDevWriteEnabled: true,
		WorkflowSavedDraftDatabaseTimeout: time.Second,
		ApplicationDraftDevHTTPEnabled:    true,
		ApplicationDraftDevWriteEnabled:   true,
		ApplicationDraftDatabaseTimeout:   time.Second,
		ApplicationPublishDevHTTPEnabled:  true,
		ApplicationPublishDevWriteEnabled: true,
		ApplicationPublishDatabaseTimeout: time.Second,
		ApplicationCatalogDevHTTPEnabled:  true,
		ApplicationCatalogDevWriteEnabled: true,
		ApplicationCatalogDatabaseTimeout: time.Second,
		APIKeyLifecycleDevHTTPEnabled:     true,
		APIKeyLifecycleDevWriteEnabled:    true,
		APIKeyDatabaseTimeout:             time.Second,
		GatewayAuthMode:                   "dev_headers",
		WorkflowExecutorDevEnabled:        true,
		WorkflowRunDatabaseTimeout:        time.Second,
		GatewayRequestHistoryDevEnabled:   true,
		GatewayRequestDatabaseTimeout:     time.Second,
		LocalPersistenceMode:              "sqlite_dev",
		SQLiteDevDatabasePath:             databasePath,
	}
}
