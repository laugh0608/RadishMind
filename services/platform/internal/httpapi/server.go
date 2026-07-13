package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const serviceName = "radishmind-platform"

type Options struct {
	BuildVersion string
}

type bridgeClient interface {
	DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error)
	DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error)
	HandleEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error)
	StreamEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions, handleEvent func(bridge.StreamEvent) error) error
}

type Server struct {
	httpServer            *http.Server
	options               Options
	bridge                bridgeClient
	config                config.Config
	controlPlaneReadStore controlPlaneReadStore
	controlPlaneReadRepo  ControlPlaneReadRepository

	savedWorkflowDraftStore               savedWorkflowDraftStore
	applicationDraftRepository            applicationConfigurationDraftRepository
	applicationPublishCandidateRepository applicationPublishCandidateRepository
	applicationCatalogRepository          applicationCatalogRepository
	apiKeyRepository                      apiKeyRepository
	workflowRunStore                      workflowRunStore
	workflowEvaluationStore               workflowEvaluationStore
	workflowEvaluationSuiteStore          workflowEvaluationSuiteStore
	gatewayRequestHistoryStore            gatewayRequestStore
	gatewayRequestHistoryStoreMode        string
	closeSavedWorkflowDraftStore          func()
	closeApplicationDraftStore            func()
	closeApplicationPublishStore          func()
	closeApplicationCatalogStore          func()
	closeWorkflowRunStore                 func()
	closeGatewayRequestStore              func()
	closeControlPlaneReadRepository       func()
	closeOnce                             sync.Once
}

type errorDocument struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message         string         `json:"message"`
	Type            string         `json:"type"`
	Code            string         `json:"code"`
	RequestID       string         `json:"request_id,omitempty"`
	Route           string         `json:"route,omitempty"`
	FailureBoundary string         `json:"failure_boundary,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

func NewServer(cfg config.Config, options Options) *Server {
	server, err := NewServerWithError(cfg, options)
	if err != nil {
		panic(err)
	}
	return server
}

func NewServerWithError(cfg config.Config, options Options) (*Server, error) {
	authenticator, err := newControlPlaneReadAuthenticator(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	controlPlaneReadRepository, closeControlPlaneReadRepository, err := newControlPlaneReadRepositoryFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	savedWorkflowDraftStore, closeSavedWorkflowDraftStore, err := newSavedWorkflowDraftStoreFromConfig(cfg)
	if err != nil {
		closeControlPlaneReadRepository()
		return nil, err
	}
	applicationDraftRepository, closeApplicationDraftStore, err := newApplicationConfigurationDraftRepositoryFromConfig(cfg)
	if err != nil {
		closeSavedWorkflowDraftStore()
		closeControlPlaneReadRepository()
		return nil, err
	}
	applicationPublishRepository, closeApplicationPublishStore, err := newApplicationPublishCandidateRepositoryFromConfig(cfg)
	if err != nil {
		closeApplicationDraftStore()
		closeSavedWorkflowDraftStore()
		closeControlPlaneReadRepository()
		return nil, err
	}
	applicationCatalogRepository, closeApplicationCatalogStore, err := newApplicationCatalogRepositoryFromConfig(cfg)
	if err != nil {
		closeApplicationPublishStore()
		closeApplicationDraftStore()
		closeSavedWorkflowDraftStore()
		closeControlPlaneReadRepository()
		return nil, err
	}
	workflowRunStore, closeWorkflowRunStore, err := newWorkflowRunStoreFromConfig(cfg)
	if err != nil {
		closeApplicationCatalogStore()
		closeApplicationPublishStore()
		closeApplicationDraftStore()
		closeSavedWorkflowDraftStore()
		closeControlPlaneReadRepository()
		return nil, err
	}
	gatewayRequestStore, gatewayRequestStoreMode, closeGatewayRequestStore, err := newGatewayRequestStoreFromConfig(cfg)
	if err != nil {
		closeWorkflowRunStore()
		closeApplicationCatalogStore()
		closeApplicationPublishStore()
		closeApplicationDraftStore()
		closeSavedWorkflowDraftStore()
		closeControlPlaneReadRepository()
		return nil, err
	}
	platformBridge, err := newPlatformBridgeClient(cfg)
	if err != nil {
		closeGatewayRequestStore()
		closeWorkflowRunStore()
		closeApplicationCatalogStore()
		closeApplicationPublishStore()
		closeApplicationDraftStore()
		if closeSavedWorkflowDraftStore != nil {
			closeSavedWorkflowDraftStore()
		}
		closeControlPlaneReadRepository()
		return nil, err
	}
	mux := http.NewServeMux()
	server := &Server{
		options:                               options,
		bridge:                                platformBridge,
		config:                                cfg,
		controlPlaneReadRepo:                  controlPlaneReadRepository,
		savedWorkflowDraftStore:               savedWorkflowDraftStore,
		applicationDraftRepository:            applicationDraftRepository,
		applicationPublishCandidateRepository: applicationPublishRepository,
		applicationCatalogRepository:          applicationCatalogRepository,
		apiKeyRepository:                      newMemoryAPIKeyRepository(),
		workflowRunStore:                      workflowRunStore,
		workflowEvaluationStore:               newWorkflowEvaluationStoreForRunStore(workflowRunStore),
		workflowEvaluationSuiteStore:          newWorkflowEvaluationSuiteStoreForRunStore(workflowRunStore),
		gatewayRequestHistoryStore:            gatewayRequestStore,
		gatewayRequestHistoryStoreMode:        gatewayRequestStoreMode,
		closeSavedWorkflowDraftStore:          closeSavedWorkflowDraftStore,
		closeApplicationDraftStore:            closeApplicationDraftStore,
		closeApplicationPublishStore:          closeApplicationPublishStore,
		closeApplicationCatalogStore:          closeApplicationCatalogStore,
		closeWorkflowRunStore:                 closeWorkflowRunStore,
		closeGatewayRequestStore:              closeGatewayRequestStore,
		closeControlPlaneReadRepository:       closeControlPlaneReadRepository,
	}

	mux.HandleFunc("GET /healthz", server.handleHealthz)
	mux.HandleFunc("GET /v1/platform/overview", server.handlePlatformOverview)
	mux.HandleFunc("GET /v1/platform/local-smoke", server.handlePlatformLocalSmoke)
	mux.HandleFunc("GET /v1/models", server.handleModels)
	mux.HandleFunc("GET /v1/models/{id}", server.handleModel)
	mux.HandleFunc("POST /v1/chat/completions", server.handleChatCompletions)
	mux.HandleFunc("POST /v1/responses", server.handleResponses)
	mux.HandleFunc("POST /v1/messages", server.handleMessages)
	mux.HandleFunc("GET /v1/session/metadata", server.handleSessionMetadata)
	mux.HandleFunc("GET /v1/session/recovery/checkpoints/{checkpoint_id}", server.handleSessionRecoveryCheckpoint)
	mux.HandleFunc("GET /v1/tools/metadata", server.handleToolsMetadata)
	mux.HandleFunc("POST /v1/tools/actions", server.handleToolAction)
	mux.HandleFunc(controlPlaneTenantSummaryRoute, server.handleControlPlaneTenantSummary)
	mux.HandleFunc(controlPlaneApplicationSummaryListRoute, server.handleUserWorkspaceApplicationSummaryList)
	mux.HandleFunc(applicationCatalogCreateRoute, server.handleCreateApplicationCatalogRecord)
	mux.HandleFunc(applicationCatalogReadRoute, server.handleReadApplicationCatalogRecord)
	mux.HandleFunc(applicationCatalogUpdateRoute, server.handleUpdateApplicationCatalogRecord)
	mux.HandleFunc(applicationCatalogArchiveRoute, server.handleArchiveApplicationCatalogRecord)
	mux.HandleFunc(controlPlaneAPIKeySummaryListRoute, server.handleUserWorkspaceAPIKeySummaryList)
	mux.HandleFunc(apiKeyCreateRoute, server.handleCreateAPIKey)
	mux.HandleFunc(apiKeyReadRoute, server.handleReadAPIKey)
	mux.HandleFunc(apiKeyRevokeRoute, server.handleRevokeAPIKey)
	mux.HandleFunc(controlPlaneQuotaSummaryRoute, server.handleUserWorkspaceQuotaSummary)
	mux.HandleFunc(controlPlaneWorkflowDefinitionSummaryListRoute, server.handleUserWorkspaceWorkflowDefinitionSummaryList)
	mux.HandleFunc(controlPlaneRunRecordSummaryListRoute, server.handleUserWorkspaceRunRecordSummaryList)
	mux.HandleFunc(controlPlaneAuditSummaryListRoute, server.handleControlPlaneAuditSummaryList)
	mux.HandleFunc(savedWorkflowDraftSaveRoute, server.handleSaveWorkflowDraft)
	mux.HandleFunc(savedWorkflowDraftListRoute, server.handleListWorkflowDrafts)
	mux.HandleFunc(savedWorkflowDraftReadRoute, server.handleReadWorkflowDraft)
	mux.HandleFunc(savedWorkflowDraftValidateRoute, server.handleValidateWorkflowDraft)
	mux.HandleFunc(applicationDraftSaveRoute, server.handleSaveApplicationConfigurationDraft)
	mux.HandleFunc(applicationDraftListRoute, server.handleListApplicationConfigurationDrafts)
	mux.HandleFunc(applicationDraftReadRoute, server.handleReadApplicationConfigurationDraft)
	mux.HandleFunc(applicationDraftValidateRoute, server.handleValidateApplicationConfigurationDraft)
	mux.HandleFunc(applicationPublishCandidateCreateRoute, server.handleCreateApplicationPublishCandidate)
	mux.HandleFunc(applicationPublishCandidateListRoute, server.handleListApplicationPublishCandidates)
	mux.HandleFunc(applicationPublishCandidateReadRoute, server.handleReadApplicationPublishCandidate)
	mux.HandleFunc(applicationPublishCandidateReviewRoute, server.handleReviewApplicationPublishCandidate)
	mux.HandleFunc(workflowExecutorStartRoute, server.handleStartWorkflowRun)
	mux.HandleFunc(workflowRunListRoute, server.handleListWorkflowRuns)
	mux.HandleFunc(workflowRunReadRoute, server.handleReadWorkflowRun)
	mux.HandleFunc(workflowRunComparisonRoute, server.handleCompareWorkflowRuns)
	mux.HandleFunc(workflowEvaluationCreateRoute, server.handleCreateWorkflowEvaluation)
	mux.HandleFunc(workflowEvaluationListRoute, server.handleListWorkflowEvaluations)
	mux.HandleFunc(workflowEvaluationReadRoute, server.handleReadWorkflowEvaluation)
	mux.HandleFunc(workflowEvaluationReviewRoute, server.handleReviewWorkflowEvaluation)
	mux.HandleFunc(workflowEvaluationRevisionCreateRoute, server.handleCreateWorkflowEvaluationRevision)
	mux.HandleFunc(workflowEvaluationRevisionListRoute, server.handleListWorkflowEvaluationRevisions)
	mux.HandleFunc(workflowEvaluationRevisionReadRoute, server.handleReadWorkflowEvaluationRevision)
	mux.HandleFunc(workflowEvaluationSuiteCreateRoute, server.handleCreateWorkflowEvaluationSuite)
	mux.HandleFunc(workflowEvaluationSuiteListRoute, server.handleListWorkflowEvaluationSuites)
	mux.HandleFunc(workflowEvaluationSuiteReadRoute, server.handleReadWorkflowEvaluationSuite)
	mux.HandleFunc(workflowEvaluationSuiteReviewRoute, server.handleReviewWorkflowEvaluationSuite)
	mux.HandleFunc(workflowEvaluationDecisionCreateRoute, server.handleCreateWorkflowEvaluationDecision)
	mux.HandleFunc(workflowEvaluationDecisionListRoute, server.handleListWorkflowEvaluationDecisions)
	mux.HandleFunc(gatewayRequestListRoute, server.handleListGatewayRequests)
	mux.HandleFunc(gatewayRequestReadRoute, server.handleReadGatewayRequest)

	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           withLocalConsoleCORS(withControlPlaneReadAuthenticator(mux, authenticator)),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}
	return server, nil
}

func newPlatformBridgeClient(cfg config.Config) (*bridge.Client, error) {
	modeText := strings.TrimSpace(cfg.BridgeMode)
	if modeText == "" {
		modeText = string(bridge.ModeProcessPerRequest)
	}
	mode, err := bridge.ParseMode(modeText)
	if err != nil {
		return nil, err
	}
	client, err := bridge.NewClientWithOptions(cfg.PythonBinary, cfg.BridgeScript, bridge.ClientOptions{
		Mode:             mode,
		WorkerCount:      cfg.BridgeWorkerCount,
		QueueCapacity:    cfg.BridgeQueueCapacity,
		HandshakeTimeout: cfg.BridgeHandshakeTimeout,
	})
	if err != nil {
		return nil, err
	}
	if mode == bridge.ModeStdioPool {
		startupTimeout := cfg.BridgeHandshakeTimeout
		if startupTimeout <= 0 {
			startupTimeout = bridge.DefaultHandshakeTimeout
		}
		workerCount := cfg.BridgeWorkerCount
		if workerCount <= 0 {
			workerCount = bridge.DefaultWorkerCount
		}
		startupTimeout *= time.Duration(workerCount)
		ctx, cancel := context.WithTimeout(context.Background(), startupTimeout)
		defer cancel()
		if err := client.Start(ctx); err != nil {
			client.Close()
			return nil, err
		}
	}
	return client, nil
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}
	err := s.httpServer.Shutdown(ctx)
	s.Close()
	return err
}

func (s *Server) Close() {
	if s == nil {
		return
	}
	s.closeOnce.Do(func() {
		if closer, ok := s.bridge.(interface{ Close() }); ok {
			closer.Close()
		}
		if s.closeSavedWorkflowDraftStore != nil {
			s.closeSavedWorkflowDraftStore()
		}
		if s.closeApplicationDraftStore != nil {
			s.closeApplicationDraftStore()
		}
		if s.closeApplicationPublishStore != nil {
			s.closeApplicationPublishStore()
		}
		if s.closeApplicationCatalogStore != nil {
			s.closeApplicationCatalogStore()
		}
		if s.closeWorkflowRunStore != nil {
			s.closeWorkflowRunStore()
		}
		if s.closeGatewayRequestStore != nil {
			s.closeGatewayRequestStore()
		}
		if s.closeControlPlaneReadRepository != nil {
			s.closeControlPlaneReadRepository()
		}
	})
}

func (s *Server) handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": serviceName,
		"version": s.options.BuildVersion,
		"path":    request.URL.Path,
	})
}

func withLocalConsoleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if applyLocalConsoleCORS(writer, request) && request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func applyLocalConsoleCORS(writer http.ResponseWriter, request *http.Request) bool {
	origin := strings.TrimSpace(request.Header.Get("Origin"))
	if !isAllowedLocalConsoleOrigin(origin) {
		return false
	}
	headers := writer.Header()
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", strings.Join(localConsoleAllowedHeaders(), ", "))
	headers.Set("Vary", "Origin")
	return true
}

func isAllowedLocalConsoleOrigin(origin string) bool {
	for _, allowedOrigin := range localConsoleAllowedOrigins() {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func localConsoleAllowedOrigins() []string {
	return []string{"http://127.0.0.1:4000", "http://localhost:4000", "http://127.0.0.1:4100", "http://localhost:4100"}
}

func localConsoleAllowedHeaders() []string {
	return []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"X-Request-Id",
		controlPlaneReadDevIdentityHeader,
		controlPlaneReadDevTenantHeader,
		controlPlaneReadDevSubjectHeader,
		controlPlaneReadDevScopesHeader,
		controlPlaneReadDevAuditHeader,
		savedWorkflowDraftDevWorkspaceHeader,
		savedWorkflowDraftDevApplicationHeader,
		applicationDraftDevWorkspaceHeader,
		applicationDraftDevApplicationHeader,
		applicationPublishDevWorkspaceHeader,
		applicationPublishDevApplicationHeader,
		gatewayRequestDevTenantHeader,
		gatewayRequestDevWorkspaceHeader,
		gatewayRequestDevConsumerHeader,
		gatewayRequestDevApplicationHeader,
		gatewayRequestDevSubjectHeader,
		gatewayRequestDevScopesHeader,
		gatewayRequestDevAuditHeader,
	}
}

func (s *Server) handleModels(writer http.ResponseWriter, request *http.Request) {
	handleModels(writer, request, s)
}

func (s *Server) handleModel(writer http.ResponseWriter, request *http.Request) {
	handleModel(writer, request, s)
}

func writeJSON(writer http.ResponseWriter, statusCode int, document any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(document)
}
