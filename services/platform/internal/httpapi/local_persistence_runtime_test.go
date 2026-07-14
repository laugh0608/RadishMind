package httpapi

import (
	"context"
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
	).Scan(&migrationCount); err != nil || migrationCount != 7 {
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
