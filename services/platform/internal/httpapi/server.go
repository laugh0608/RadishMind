package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
)

const serviceName = "radishmind-platform"

type Options struct {
	BuildVersion                               string
	TestOnly                                   bool
	ApplicationResultArtifactLibraryDevFixture bool
}

type bridgeClient interface {
	DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error)
	DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error)
	HandleEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error)
	StreamEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions, handleEvent func(bridge.StreamEvent) error) error
}

type Server struct {
	httpServer                    *http.Server
	options                       Options
	bridge                        bridgeClient
	config                        config.Config
	controlPlaneReadStore         controlPlaneReadStore
	controlPlaneReadRepo          ControlPlaneReadRepository
	workspaceControlPlaneReadRepo ControlPlaneReadRepository
	workspaceMembershipProvider   WorkspaceMembershipProvider

	savedWorkflowDraftStore                 savedWorkflowDraftStore
	applicationDraftRepository              applicationConfigurationDraftRepository
	applicationPublishCandidateRepository   applicationPublishCandidateRepository
	applicationCatalogRepository            applicationCatalogRepository
	promptApplicationTemplateRepository     promptApplicationTemplateRepository
	agentCopilotProfileRepository           agentCopilotProfileRepository
	adminProviderRouteRepository            adminProviderRouteRepository
	providerRouteSnapshotProvider           gatewayProviderRouteSnapshotProvider
	applicationInteractionSessionRepository applicationInteractionSessionRepository
	applicationSessionRepository            applicationInteractionSessionRepository
	applicationResultArtifactRepository     applicationResultArtifactRepository
	apiKeyRepository                        apiKeyRepository
	workflowRunStore                        workflowRunStore
	applicationRunStore                     workflowRunStore
	workflowDefinitionReleaseRepository     workflowDefinitionReleaseRepository
	workflowTemplateCatalogRepository       workflowTemplateCatalogRepository
	workflowTemplateTargetBindingValidator  workflowTemplateTargetBindingValidator
	workflowRAGSnapshotRepository           workflowRAGSnapshotRepository
	workflowRAGEvaluationDatasetRepository  workflowRAGEvaluationDatasetRepository
	workflowRAGPromotionRepository          workflowRAGPromotionRepository
	workflowRAGAppRuntimeRepository         workflowRAGApplicationRuntimeRepository
	promptApplicationRuntimeRepository      promptApplicationRuntimeRepository
	agentCopilotRuntimeRepository           agentCopilotRuntimeRepository
	workflowHTTPToolActionStore             workflowHTTPToolActionStore
	workflowHTTPToolExecutionStore          workflowHTTPToolExecutionStore
	workflowHTTPToolExecutionTransport      *workflowHTTPToolTransport
	workflowEvaluationStore                 workflowEvaluationStore
	workflowEvaluationSuiteStore            workflowEvaluationSuiteStore
	applicationEvaluationRepository         applicationEvaluationRepository
	applicationEvaluationScheduleRepository applicationEvaluationScheduleRepository
	gatewayRequestHistoryStore              gatewayRequestStore
	gatewayRequestHistoryStoreMode          string
	gatewayRequestQuotaRepository           GatewayRequestQuotaRepository
	gatewayModelPricingRepository           GatewayModelPricingRepository
	localIdentityHTTPService                *localIdentityHTTPService
	localIdentityAdministrationService      *localIdentityAdministrationService
	localIdentitySelfServiceSecurityService *localIdentitySelfServiceSecurityService
	closeSavedWorkflowDraftStore            func()
	closeApplicationDraftStore              func()
	closeApplicationPublishStore            func()
	closeApplicationCatalogStore            func()
	closePromptApplicationTemplateStore     func()
	closeAgentCopilotProfileStore           func()
	closeAdminProviderRouteStore            func()
	closeAPIKeyStore                        func()
	closeWorkflowRunStore                   func()
	closeGatewayRequestStore                func()
	closeGatewayRequestQuotaStore           func()
	closeGatewayModelPricingStore           func()
	closeLocalIdentityRepository            func()
	localPersistenceRuntime                 *sqlitedev.Runtime
	closeControlPlaneReadRepository         func()
	closeOnce                               sync.Once
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
	if err := config.ValidateServerStart(cfg); err != nil {
		return nil, err
	}
	if cfg.WorkflowHTTPToolTestLoopbackEnabled && !options.TestOnly {
		return nil, fmt.Errorf("workflow HTTP tool test loopback is restricted to explicit test servers")
	}
	runtimeConfig := config.EffectiveLocalPersistenceConfig(cfg)
	if options.ApplicationResultArtifactLibraryDevFixture &&
		(!runtimeConfig.ApplicationSessionDevEnabled || !runtimeConfig.ControlPlaneReadDevAuthEnabled ||
			!runtimeConfig.ApplicationCatalogDevHTTPEnabled || !runtimeConfig.ApplicationCatalogDevWriteEnabled ||
			config.EffectiveLocalPersistenceMode(runtimeConfig) != "sqlite_dev") {
		return nil, fmt.Errorf("application result artifact library fixture requires the explicit SQLite local-product session gates")
	}
	authenticator, err := newControlPlaneReadAuthenticator(context.Background(), runtimeConfig)
	if err != nil {
		return nil, err
	}
	controlPlaneReadRepository, closeControlPlaneReadRepository, err := newControlPlaneReadRepositoryFromConfig(runtimeConfig)
	if err != nil {
		return nil, err
	}
	localPersistenceRuntime, err := openLocalPersistenceRuntime(runtimeConfig)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository)
		return nil, err
	}
	closeLocalPersistenceRuntime := sqliteRuntimeCloser(localPersistenceRuntime)
	savedWorkflowDraftStore, closeSavedWorkflowDraftStore, err := newSavedWorkflowDraftStoreFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime)
		return nil, err
	}
	applicationDraftRepository, closeApplicationDraftStore, err := newApplicationConfigurationDraftRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore)
		return nil, err
	}
	applicationPublishRepository, closeApplicationPublishStore, err := newApplicationPublishCandidateRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore)
		return nil, err
	}
	applicationCatalogRepository, closeApplicationCatalogStore, err := newApplicationCatalogRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore)
		return nil, err
	}
	apiKeyRepository, closeAPIKeyStore, err := newAPIKeyRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore)
		return nil, err
	}
	workflowRunStore, closeWorkflowRunStore, err := newWorkflowRunStoreFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore)
		return nil, err
	}
	applicationEvaluationRepository := newApplicationEvaluationRepositoryForRunStore(workflowRunStore)
	if runtimeConfig.ApplicationEvaluationCampaignDevEnabled && applicationEvaluationRepository == nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, fmt.Errorf("application evaluation campaign requires a supported workflow runtime backend")
	}
	applicationEvaluationScheduleRepository, err := newApplicationEvaluationScheduleRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	applicationInteractionSessionRepository, err := newApplicationInteractionSessionRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	promptApplicationSessionRepository, err := newPromptApplicationSessionRepositoryForLegacy(applicationInteractionSessionRepository)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	agentCopilotSessionRepository, err := newAgentCopilotSessionRepositoryForLegacy(applicationInteractionSessionRepository)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	combinedApplicationSessionRepository := newCombinedApplicationInteractionSessionRepositoryWithAgent(
		applicationInteractionSessionRepository, promptApplicationSessionRepository, agentCopilotSessionRepository,
	)
	applicationResultArtifactRepository, err := newApplicationResultArtifactRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	var workflowDefinitionReleaseRepository workflowDefinitionReleaseRepository
	if runtimeConfig.WorkflowDefinitionReleaseDevEnabled {
		workflowDefinitionReleaseRepository, err = newWorkflowDefinitionReleaseRepositoryForRunStore(workflowRunStore)
		if err != nil {
			closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
			return nil, err
		}
		controlPlaneReadRepository = liveWorkflowDefinitionControlPlaneReadRepository{ControlPlaneReadRepository: controlPlaneReadRepository, definitions: workflowDefinitionReleaseRepository}
	}
	var workflowTemplateCatalogRepository workflowTemplateCatalogRepository
	if runtimeConfig.WorkflowTemplateCatalogDevEnabled {
		workflowTemplateCatalogRepository, err = newWorkflowTemplateCatalogRepositoryForSavedDraftStore(savedWorkflowDraftStore)
		if err != nil {
			closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
			return nil, err
		}
	}
	workflowRAGSnapshotRepository, err := newWorkflowRAGSnapshotRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	workflowRAGEvaluationDatasetRepository := newWorkflowRAGEvaluationDatasetRepositoryForRunStore(workflowRunStore)
	var workflowRAGPromotionRepository workflowRAGPromotionRepository
	if runtimeConfig.WorkflowRAGPromotionDevEnabled {
		workflowRAGPromotionRepository, err = newWorkflowRAGPromotionRepositoryForRunStore(workflowRunStore)
		if err != nil {
			closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
			return nil, err
		}
	}
	var workflowRAGApplicationRuntimeRepository workflowRAGApplicationRuntimeRepository
	if runtimeConfig.WorkflowRAGAppInvocationDevEnabled {
		workflowRAGApplicationRuntimeRepository, err = newWorkflowRAGApplicationRuntimeRepositoryForRunStore(workflowRunStore)
		if err != nil {
			closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
			return nil, err
		}
	}
	promptApplicationRuntimeRepository, err := newPromptApplicationRuntimeRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	agentCopilotRuntimeRepository, err := newAgentCopilotRuntimeRepositoryForRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	promptApplicationRunStore, err := newPromptApplicationRunStoreForWorkflowRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	agentCopilotRunStore, err := newAgentCopilotRunStoreForWorkflowRunStore(workflowRunStore)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	combinedRunStore := newCombinedWorkflowRunStoreWithAgent(workflowRunStore, promptApplicationRunStore, agentCopilotRunStore)
	workspaceControlPlaneReadRepository := newWorkspaceScopedControlPlaneReadRepository(
		controlPlaneReadRepository,
		applicationCatalogRepository,
		apiKeyRepository,
		workflowDefinitionReleaseRepository,
		combinedRunStore,
	)
	gatewayRequestStore, gatewayRequestStoreMode, closeGatewayRequestStore, err := newGatewayRequestStoreFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore)
		return nil, err
	}
	promptApplicationTemplateRepository, closePromptApplicationTemplateStore, err := newPromptApplicationTemplateRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore)
		return nil, err
	}
	agentCopilotProfileRepository, closeAgentCopilotProfileStore, err := newAgentCopilotProfileRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore)
		return nil, err
	}
	adminProviderRouteRepository, closeAdminProviderRouteStore, err := newAdminProviderRouteRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore)
		return nil, err
	}
	gatewayRequestQuotaRepository, closeGatewayRequestQuotaStore, err := newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore)
		return nil, err
	}
	gatewayModelPricingRepository, closeGatewayModelPricingStore, err := newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(runtimeConfig, localPersistenceRuntime)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore)
		return nil, err
	}
	localIdentityRepository, closeLocalIdentityRepository, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{
		Mode: runtimeConfig.LocalIdentityStoreMode, SQLiteRuntime: localPersistenceRuntime,
		PostgresDatabaseURL: runtimeConfig.LocalIdentityDatabaseURL, DatabaseTimeout: runtimeConfig.LocalIdentityDatabaseTimeout,
	})
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore, closeGatewayModelPricingStore)
		return nil, err
	}
	localIdentityHTTPService := newLocalIdentityHTTPService(runtimeConfig, localIdentityRepository)
	localIdentityAdministrationRepository, ok := localIdentityRepository.(localIdentityAdministrationRepository)
	if !ok {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore, closeGatewayModelPricingStore, closeLocalIdentityRepository)
		return nil, errors.New("local identity administration repository is unavailable")
	}
	localIdentityAdministrationService := newLocalIdentityAdministrationService(localIdentityAdministrationRepository)
	localIdentitySelfServiceSecurityRepository, ok := localIdentityRepository.(localIdentitySelfServiceSecurityRepository)
	if !ok {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore, closeGatewayModelPricingStore, closeLocalIdentityRepository)
		return nil, errors.New("local identity self-service security repository is unavailable")
	}
	localIdentitySelfServiceSecurityService := newLocalIdentitySelfServiceSecurityService(localIdentitySelfServiceSecurityRepository)
	if err := localIdentityHTTPService.configureOIDC(context.Background(), runtimeConfig); err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore, closeGatewayModelPricingStore, closeLocalIdentityRepository)
		return nil, err
	}
	rawPlatformBridge, err := newPlatformBridgeClient(runtimeConfig)
	if err != nil {
		closeServerStartupResources(closeControlPlaneReadRepository, closeLocalPersistenceRuntime, closeSavedWorkflowDraftStore, closeApplicationDraftStore, closeApplicationPublishStore, closeApplicationCatalogStore, closeAPIKeyStore, closeWorkflowRunStore, closeGatewayRequestStore, closePromptApplicationTemplateStore, closeAgentCopilotProfileStore, closeAdminProviderRouteStore, closeGatewayRequestQuotaStore, closeGatewayModelPricingStore, closeLocalIdentityRepository)
		return nil, err
	}
	var platformBridge bridgeClient = newGatewayProviderAttemptBridgeClient(rawPlatformBridge)
	if runtimeConfig.GatewayRequestQuotaEnforcementDevEnabled {
		platformBridge = newGatewayRequestQuotaBridgeClient(platformBridge, gatewayRequestQuotaRepository)
	}
	mux := http.NewServeMux()
	workflowHTTPToolActionStore := newWorkflowHTTPToolActionStoreForRunStore(workflowRunStore)
	server := &Server{
		options:                                 options,
		bridge:                                  platformBridge,
		config:                                  runtimeConfig,
		controlPlaneReadRepo:                    controlPlaneReadRepository,
		workspaceControlPlaneReadRepo:           workspaceControlPlaneReadRepository,
		workspaceMembershipProvider:             newDeterministicDevTestWorkspaceMembershipProvider(),
		savedWorkflowDraftStore:                 savedWorkflowDraftStore,
		applicationDraftRepository:              applicationDraftRepository,
		applicationPublishCandidateRepository:   applicationPublishRepository,
		applicationCatalogRepository:            applicationCatalogRepository,
		promptApplicationTemplateRepository:     promptApplicationTemplateRepository,
		agentCopilotProfileRepository:           agentCopilotProfileRepository,
		adminProviderRouteRepository:            adminProviderRouteRepository,
		providerRouteSnapshotProvider:           adminProviderRouteSnapshotProvider{repository: adminProviderRouteRepository},
		applicationInteractionSessionRepository: applicationInteractionSessionRepository,
		applicationSessionRepository:            combinedApplicationSessionRepository,
		applicationResultArtifactRepository:     applicationResultArtifactRepository,
		apiKeyRepository:                        apiKeyRepository,
		workflowRunStore:                        workflowRunStore,
		applicationRunStore:                     combinedRunStore,
		workflowDefinitionReleaseRepository:     workflowDefinitionReleaseRepository,
		workflowTemplateCatalogRepository:       workflowTemplateCatalogRepository,
		workflowTemplateTargetBindingValidator: configuredWorkflowTemplateTargetBindingValidator{
			providerRouteSource:      config.EffectiveGatewayProviderRouteSource(runtimeConfig),
			providerRouteEnvironment: runtimeConfig.GatewayProviderRouteEnvironment,
			providerRouteConfigID:    runtimeConfig.GatewayProviderRouteConfigurationID,
			snapshotProvider:         adminProviderRouteSnapshotProvider{repository: adminProviderRouteRepository},
			bridge:                   platformBridge,
		},
		workflowRAGSnapshotRepository:           workflowRAGSnapshotRepository,
		workflowRAGEvaluationDatasetRepository:  workflowRAGEvaluationDatasetRepository,
		workflowRAGPromotionRepository:          workflowRAGPromotionRepository,
		workflowRAGAppRuntimeRepository:         workflowRAGApplicationRuntimeRepository,
		promptApplicationRuntimeRepository:      promptApplicationRuntimeRepository,
		agentCopilotRuntimeRepository:           agentCopilotRuntimeRepository,
		workflowHTTPToolActionStore:             workflowHTTPToolActionStore,
		workflowHTTPToolExecutionStore:          newWorkflowHTTPToolExecutionStoreForRunStore(workflowRunStore, workflowHTTPToolActionStore),
		workflowEvaluationStore:                 newWorkflowEvaluationStoreForRunStore(workflowRunStore),
		workflowEvaluationSuiteStore:            newWorkflowEvaluationSuiteStoreForRunStore(workflowRunStore),
		applicationEvaluationRepository:         applicationEvaluationRepository,
		applicationEvaluationScheduleRepository: applicationEvaluationScheduleRepository,
		gatewayRequestHistoryStore:              gatewayRequestStore,
		gatewayRequestHistoryStoreMode:          gatewayRequestStoreMode,
		gatewayRequestQuotaRepository:           gatewayRequestQuotaRepository,
		gatewayModelPricingRepository:           gatewayModelPricingRepository,
		localIdentityHTTPService:                localIdentityHTTPService,
		localIdentityAdministrationService:      localIdentityAdministrationService,
		localIdentitySelfServiceSecurityService: localIdentitySelfServiceSecurityService,
		closeSavedWorkflowDraftStore:            closeSavedWorkflowDraftStore,
		closeApplicationDraftStore:              closeApplicationDraftStore,
		closeApplicationPublishStore:            closeApplicationPublishStore,
		closeApplicationCatalogStore:            closeApplicationCatalogStore,
		closePromptApplicationTemplateStore:     closePromptApplicationTemplateStore,
		closeAgentCopilotProfileStore:           closeAgentCopilotProfileStore,
		closeAdminProviderRouteStore:            closeAdminProviderRouteStore,
		closeAPIKeyStore:                        closeAPIKeyStore,
		closeWorkflowRunStore:                   closeWorkflowRunStore,
		closeGatewayRequestStore:                closeGatewayRequestStore,
		closeGatewayRequestQuotaStore:           closeGatewayRequestQuotaStore,
		closeGatewayModelPricingStore:           closeGatewayModelPricingStore,
		closeLocalIdentityRepository:            closeLocalIdentityRepository,
		localPersistenceRuntime:                 localPersistenceRuntime,
		closeControlPlaneReadRepository:         closeControlPlaneReadRepository,
	}
	if config.EffectiveControlPlaneReadAuthMode(runtimeConfig) == controlPlaneReadAuthModeLocalSessionDevTest {
		server.workspaceMembershipProvider = newLocalWorkspaceMembershipProvider(localIdentityRepository)
	}
	if options.ApplicationResultArtifactLibraryDevFixture {
		if _, err = seedApplicationResultArtifactLibraryDevFixture(server); err != nil {
			server.Close()
			return nil, fmt.Errorf("seed application result artifact library fixture: %w", err)
		}
	}
	registerLocalIdentityHTTPRoutes(mux, server)

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
	mux.HandleFunc(applicationCatalogUnarchiveRoute, server.handleUnarchiveApplicationCatalogRecord)
	mux.HandleFunc(applicationSessionCreateRoute, server.handleCreateApplicationInteractionSession)
	mux.HandleFunc(applicationSessionListRoute, server.handleListApplicationInteractionSessions)
	mux.HandleFunc(applicationSessionReadRoute, server.handleReadApplicationInteractionSession)
	mux.HandleFunc(applicationSessionCloseRoute, server.handleCloseApplicationInteractionSession)
	mux.HandleFunc(applicationSessionTurnListRoute, server.handleListApplicationInteractionTurns)
	mux.HandleFunc(applicationSessionTurnRoute, server.handleExecuteApplicationInteractionTurn)
	mux.HandleFunc(applicationResultArtifactListRoute, server.handleListApplicationResultArtifacts)
	mux.HandleFunc(applicationResultArtifactReadRoute, server.handleReadApplicationResultArtifact)
	mux.HandleFunc(applicationResultArtifactArchiveRoute, server.handleArchiveApplicationResultArtifact)
	mux.HandleFunc(applicationResultArtifactUnarchiveRoute, server.handleUnarchiveApplicationResultArtifact)
	mux.HandleFunc(applicationResultArtifactApplicationListRoute, server.handleListApplicationResultArtifactsByApplication)
	mux.HandleFunc(applicationResultArtifactExportRoute, server.handleExportApplicationResultArtifact)
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
	mux.HandleFunc(savedWorkflowDraftArchiveRoute, server.handleArchiveWorkflowDraft)
	mux.HandleFunc(savedWorkflowDraftUnarchiveRoute, server.handleUnarchiveWorkflowDraft)
	mux.HandleFunc(savedWorkflowDraftRevisionListRoute, server.handleListWorkflowDraftRevisions)
	mux.HandleFunc(savedWorkflowDraftRevisionReadRoute, server.handleReadWorkflowDraftRevision)
	mux.HandleFunc(savedWorkflowDraftRevisionRestoreRoute, server.handleRestoreWorkflowDraftRevision)
	mux.HandleFunc(applicationDraftSaveRoute, server.handleSaveApplicationConfigurationDraft)
	mux.HandleFunc(applicationDraftListRoute, server.handleListApplicationConfigurationDrafts)
	mux.HandleFunc(applicationDraftReadRoute, server.handleReadApplicationConfigurationDraft)
	mux.HandleFunc(applicationDraftValidateRoute, server.handleValidateApplicationConfigurationDraft)
	mux.HandleFunc(applicationDraftPromptTemplateBindingRoute, server.handleBindApplicationConfigurationDraftPromptTemplate)
	mux.HandleFunc(applicationDraftAgentProfileBindingRoute, server.handleBindApplicationConfigurationDraftAgentProfile)
	mux.HandleFunc(promptApplicationRuntimeReadRoute, server.handleReadPromptApplicationRuntimeAssignment)
	mux.HandleFunc(promptApplicationRuntimeEventsRoute, server.handleReadPromptApplicationRuntimeEvents)
	mux.HandleFunc(promptApplicationRuntimeDecisionRoute, server.handleDecidePromptApplicationRuntimeAssignment)
	mux.HandleFunc(agentCopilotRuntimeReadRoute, server.handleReadAgentCopilotRuntimeAssignment)
	mux.HandleFunc(agentCopilotRuntimeEventsRoute, server.handleReadAgentCopilotRuntimeEvents)
	mux.HandleFunc(agentCopilotRuntimeDecisionRoute, server.handleDecideAgentCopilotRuntimeAssignment)
	mux.HandleFunc(promptApplicationTemplateValidateRoute, server.handleValidatePromptApplicationTemplate)
	mux.HandleFunc(promptApplicationTemplateSaveRoute, server.handleSavePromptApplicationTemplate)
	mux.HandleFunc(promptApplicationTemplateListRoute, server.handleListPromptApplicationTemplates)
	mux.HandleFunc(promptApplicationTemplateReadRoute, server.handleReadPromptApplicationTemplate)
	mux.HandleFunc(promptApplicationTemplateVersionCreateRoute, server.handleCreatePromptApplicationTemplateVersion)
	mux.HandleFunc(promptApplicationTemplateVersionListRoute, server.handleListPromptApplicationTemplateVersions)
	mux.HandleFunc(promptApplicationTemplateVersionReadRoute, server.handleReadPromptApplicationTemplateVersion)
	mux.HandleFunc(agentCopilotProfileValidateRoute, server.handleValidateAgentCopilotProfile)
	mux.HandleFunc(agentCopilotProfileSaveRoute, server.handleSaveAgentCopilotProfile)
	mux.HandleFunc(agentCopilotProfileListRoute, server.handleListAgentCopilotProfiles)
	mux.HandleFunc(agentCopilotProfileReadRoute, server.handleReadAgentCopilotProfile)
	mux.HandleFunc(agentCopilotProfileVersionCreateRoute, server.handleCreateAgentCopilotProfileVersion)
	mux.HandleFunc(agentCopilotProfileVersionListRoute, server.handleListAgentCopilotProfileVersions)
	mux.HandleFunc(agentCopilotProfileVersionReadRoute, server.handleReadAgentCopilotProfileVersion)
	mux.HandleFunc(adminProviderRouteDraftReadRoute, server.handleReadAdminProviderRouteDraft)
	mux.HandleFunc(adminProviderRouteDraftPutRoute, server.handlePutAdminProviderRouteDraft)
	mux.HandleFunc(adminProviderRouteCandidateCreateRoute, server.handleCreateAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteCandidateReadRoute, server.handleReadAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteReviewRoute, server.handleReviewAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteActivationRoute, server.handleActivateAdminProviderRouteCandidate)
	mux.HandleFunc(adminProviderRouteActiveSnapshotRoute, server.handleReadAdminProviderRouteActiveSnapshot)
	mux.HandleFunc(adminProviderRouteActivationHistoryRoute, server.handleListAdminProviderRouteActivations)
	mux.HandleFunc(gatewayRequestQuotaAdminReadRoute, server.handleReadGatewayRequestQuota)
	mux.HandleFunc(gatewayRequestQuotaAdminPutRoute, server.handlePutGatewayRequestQuota)
	mux.HandleFunc(gatewayModelPricingAdminReadRoute, server.handleReadGatewayModelPricing)
	mux.HandleFunc(gatewayModelPricingAdminPutRoute, server.handlePutGatewayModelPricing)
	mux.HandleFunc(applicationPublishCandidateCreateRoute, server.handleCreateApplicationPublishCandidate)
	mux.HandleFunc(applicationPublishCandidateListRoute, server.handleListApplicationPublishCandidates)
	mux.HandleFunc(applicationPublishCandidateReadRoute, server.handleReadApplicationPublishCandidate)
	mux.HandleFunc(applicationPublishCandidateReviewRoute, server.handleReviewApplicationPublishCandidate)
	mux.HandleFunc(workflowDefinitionCandidateCreateRoute, server.handleCreateWorkflowDefinitionCandidate)
	mux.HandleFunc(workflowDefinitionCandidateListRoute, server.handleListWorkflowDefinitionCandidates)
	mux.HandleFunc(workflowDefinitionCandidateReadRoute, server.handleReadWorkflowDefinitionCandidate)
	mux.HandleFunc(workflowDefinitionCandidateDecisionRoute, server.handleDecideWorkflowDefinitionCandidate)
	mux.HandleFunc(workflowDefinitionVersionListRoute, server.handleListWorkflowDefinitionVersions)
	mux.HandleFunc(workflowDefinitionVersionReadRoute, server.handleReadWorkflowDefinitionVersion)
	mux.HandleFunc(workflowDefinitionActivationReadRoute, server.handleReadWorkflowDefinitionActivation)
	mux.HandleFunc(workflowDefinitionActivationDecisionRoute, server.handleDecideWorkflowDefinitionActivation)
	mux.HandleFunc(workflowDefinitionRunCreateRoute, server.handleStartWorkflowDefinitionRun)
	mux.HandleFunc(workflowTemplateCandidateCreateRoute, server.handleCreateWorkflowTemplateCandidate)
	mux.HandleFunc(workflowTemplateCandidateListRoute, server.handleListWorkflowTemplateCandidates)
	mux.HandleFunc(workflowTemplateCandidateReadRoute, server.handleReadWorkflowTemplateCandidate)
	mux.HandleFunc(workflowTemplateCandidateDecisionRoute, server.handleDecideWorkflowTemplateCandidate)
	mux.HandleFunc(workflowTemplateListRoute, server.handleListWorkflowTemplates)
	mux.HandleFunc(workflowTemplateReadRoute, server.handleReadWorkflowTemplate)
	mux.HandleFunc(workflowTemplateVersionListRoute, server.handleListWorkflowTemplateVersions)
	mux.HandleFunc(workflowTemplateVersionReadRoute, server.handleReadWorkflowTemplateVersion)
	mux.HandleFunc(workflowTemplateListingDecisionRoute, server.handleDecideWorkflowTemplateListing)
	mux.HandleFunc(workflowTemplateDerivationRoute, server.handleDeriveWorkflowTemplate)
	mux.HandleFunc(workflowExecutorStartRoute, server.handleStartWorkflowRun)
	mux.HandleFunc("POST "+workflowRAGExecutionRoute, server.handleWorkflowRAGExecution)
	mux.HandleFunc(workflowRAGSnapshotCreateRoute, server.handleCreateWorkflowRAGSnapshot)
	mux.HandleFunc(workflowRAGSnapshotListRoute, server.handleListWorkflowRAGSnapshots)
	mux.HandleFunc(workflowRAGSnapshotReadRoute, server.handleReadWorkflowRAGSnapshot)
	mux.HandleFunc(workflowRAGSnapshotVersionRoute, server.handleVersionWorkflowRAGSnapshot)
	mux.HandleFunc(workflowRAGSnapshotArchiveRoute, server.handleArchiveWorkflowRAGSnapshot)
	mux.HandleFunc(workflowRAGEvaluationDatasetCreateRoute, server.handleCreateWorkflowRAGEvaluationDataset)
	mux.HandleFunc(workflowRAGEvaluationDatasetListRoute, server.handleListWorkflowRAGEvaluationDatasets)
	mux.HandleFunc(workflowRAGEvaluationDatasetReadRoute, server.handleReadWorkflowRAGEvaluationDataset)
	mux.HandleFunc(workflowRAGEvaluationDatasetVersionRoute, server.handleVersionWorkflowRAGEvaluationDataset)
	mux.HandleFunc(workflowRAGEvaluationDatasetArchiveRoute, server.handleArchiveWorkflowRAGEvaluationDataset)
	mux.HandleFunc(workflowRAGCandidateReviewCreateRoute, server.handleCreateWorkflowRAGCandidateReview)
	mux.HandleFunc(workflowRAGCandidateReviewListRoute, server.handleListWorkflowRAGCandidateReviews)
	mux.HandleFunc(workflowRAGCandidateReviewReadRoute, server.handleReadWorkflowRAGCandidateReview)
	mux.HandleFunc(workflowRAGPromotionCandidateCreateRoute, server.handleCreateWorkflowRAGPromotionCandidate)
	mux.HandleFunc(workflowRAGPromotionCandidateListRoute, server.handleListWorkflowRAGPromotionCandidates)
	mux.HandleFunc(workflowRAGPromotionCandidateReadRoute, server.handleReadWorkflowRAGPromotionCandidate)
	mux.HandleFunc(workflowRAGPromotionDecisionRoute, server.handleDecideWorkflowRAGPromotionCandidate)
	mux.HandleFunc(workflowRAGApplicationRuntimeAssignmentReadRoute, server.handleReadWorkflowRAGApplicationRuntimeAssignment)
	mux.HandleFunc(workflowRAGApplicationRuntimeAssignmentDecisionRoute, server.handleDecideWorkflowRAGApplicationRuntimeAssignment)
	mux.HandleFunc("POST "+workflowRAGApplicationInvocationRoute, server.handleWorkflowRAGApplicationInvocation)
	mux.HandleFunc("POST "+promptApplicationInvocationRoute, server.handlePromptApplicationInvocation)
	mux.HandleFunc("POST "+agentCopilotInvocationRoute, server.handleAgentCopilotInvocation)
	mux.HandleFunc(workflowHTTPToolPlanCreateRoute, server.handleCreateWorkflowHTTPToolActionPlan)
	mux.HandleFunc(workflowDefinitionHTTPToolPlanCreateRoute, server.handleCreateWorkflowDefinitionHTTPToolActionPlan)
	mux.HandleFunc(workflowHTTPToolPlanReadRoute, server.handleReadWorkflowHTTPToolActionPlan)
	mux.HandleFunc(workflowHTTPToolDecisionRoute, server.handleDecideWorkflowHTTPToolActionPlan)
	mux.HandleFunc(workflowHTTPToolExecutionRoute, server.handleExecuteWorkflowHTTPToolActionPlan)
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
	mux.HandleFunc(applicationEvaluationPlanCreateRoute, server.handleCreateApplicationEvaluationPlan)
	mux.HandleFunc(applicationEvaluationPlanListRoute, server.handleListApplicationEvaluationPlans)
	mux.HandleFunc(applicationEvaluationPlanReadRoute, server.handleReadApplicationEvaluationPlan)
	mux.HandleFunc(applicationEvaluationPlanReviseRoute, server.handleReviseApplicationEvaluationPlan)
	mux.HandleFunc(applicationEvaluationPlanArchiveRoute, server.handleArchiveApplicationEvaluationPlan)
	mux.HandleFunc(applicationEvaluationVersionListRoute, server.handleListApplicationEvaluationPlanVersions)
	mux.HandleFunc(applicationEvaluationVersionReadRoute, server.handleReadApplicationEvaluationPlanVersion)
	mux.HandleFunc(applicationEvaluationCampaignExecuteRoute, server.handleExecuteApplicationEvaluationCampaign)
	mux.HandleFunc(applicationEvaluationCampaignListRoute, server.handleListApplicationEvaluationCampaigns)
	mux.HandleFunc(applicationEvaluationCampaignReadRoute, server.handleReadApplicationEvaluationCampaign)
	mux.HandleFunc(applicationEvaluationCampaignReconcileRoute, server.handleReconcileApplicationEvaluationCampaign)
	mux.HandleFunc(applicationEvaluationScheduleCreateRoute, server.handleCreateApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleListRoute, server.handleListApplicationEvaluationSchedules)
	mux.HandleFunc(applicationEvaluationScheduleReadRoute, server.handleReadApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleReviseRoute, server.handleReviseApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleActivateRoute, server.handleActivateApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationSchedulePauseRoute, server.handlePauseApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleResumeRoute, server.handleResumeApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleArchiveRoute, server.handleArchiveApplicationEvaluationSchedule)
	mux.HandleFunc(applicationEvaluationScheduleVersionReadRoute, server.handleReadApplicationEvaluationScheduleVersion)
	mux.HandleFunc(applicationEvaluationScheduleOccurrenceReadRoute, server.handleReadApplicationEvaluationScheduleOccurrence)
	mux.HandleFunc(applicationEvaluationPairPreviewRoute, server.handlePreviewApplicationEvaluationCampaignPair)
	mux.HandleFunc(applicationEvaluationHandoffRoute, server.handleMaterializeApplicationEvaluationHandoff)
	mux.HandleFunc(gatewayRequestListRoute, server.handleListGatewayRequests)
	mux.HandleFunc(gatewayRequestReadRoute, server.handleReadGatewayRequest)

	server.httpServer = &http.Server{
		Addr: runtimeConfig.ListenAddr,
		Handler: withLocalConsoleCORS(
			withLocalIdentitySessionAuthentication(withControlPlaneReadAuthenticator(mux, authenticator), server.localIdentityHTTPService),
			runtimeConfig,
		),
		ReadHeaderTimeout: runtimeConfig.ReadHeaderTimeout,
		WriteTimeout:      runtimeConfig.WriteTimeout,
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
		if s.closeGatewayRequestStore != nil {
			s.closeGatewayRequestStore()
		}
		if s.closeGatewayRequestQuotaStore != nil {
			s.closeGatewayRequestQuotaStore()
		}
		if s.closeGatewayModelPricingStore != nil {
			s.closeGatewayModelPricingStore()
		}
		if s.closeWorkflowRunStore != nil {
			s.closeWorkflowRunStore()
		}
		if s.closeAPIKeyStore != nil {
			s.closeAPIKeyStore()
		}
		if s.closeApplicationCatalogStore != nil {
			s.closeApplicationCatalogStore()
		}
		if s.closeAgentCopilotProfileStore != nil {
			s.closeAgentCopilotProfileStore()
		}
		if s.closeAdminProviderRouteStore != nil {
			s.closeAdminProviderRouteStore()
		}
		if s.closePromptApplicationTemplateStore != nil {
			s.closePromptApplicationTemplateStore()
		}
		if s.closeApplicationPublishStore != nil {
			s.closeApplicationPublishStore()
		}
		if s.closeApplicationDraftStore != nil {
			s.closeApplicationDraftStore()
		}
		if s.closeSavedWorkflowDraftStore != nil {
			s.closeSavedWorkflowDraftStore()
		}
		if s.closeLocalIdentityRepository != nil {
			s.closeLocalIdentityRepository()
		}
		if s.localPersistenceRuntime != nil {
			_ = s.localPersistenceRuntime.Close()
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

func withLocalConsoleCORS(next http.Handler, cfg config.Config) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if applyLocalConsoleCORS(writer, request, cfg) && request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func applyLocalConsoleCORS(writer http.ResponseWriter, request *http.Request, cfg config.Config) bool {
	origin := strings.TrimSpace(request.Header.Get("Origin"))
	if !isAllowedLocalConsoleOrigin(origin, cfg) {
		return false
	}
	headers := writer.Header()
	headers.Set("Access-Control-Allow-Origin", origin)
	if cfg.LocalIdentityDevHTTPEnabled && origin == strings.TrimSpace(cfg.LocalIdentityAllowedOrigin) {
		headers.Set("Access-Control-Allow-Credentials", "true")
	} else {
		headers.Del("Access-Control-Allow-Credentials")
	}
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", strings.Join(localConsoleAllowedHeaders(), ", "))
	headers.Set("Access-Control-Expose-Headers", strings.Join(localConsoleExposedHeaders(), ", "))
	headers.Set("Vary", "Origin")
	return true
}

func isAllowedLocalConsoleOrigin(origin string, cfg config.Config) bool {
	for _, allowedOrigin := range localConsoleAllowedOrigins(cfg) {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func localConsoleAllowedOrigins(configs ...config.Config) []string {
	origins := []string{"http://127.0.0.1:4000", "http://localhost:4000", "http://127.0.0.1:4100", "http://localhost:4100"}
	if len(configs) == 1 && configs[0].LocalIdentityDevHTTPEnabled {
		origin := strings.TrimSpace(configs[0].LocalIdentityAllowedOrigin)
		if origin != "" && !slices.Contains(origins, origin) {
			origins = append(origins, origin)
		}
	}
	return origins
}

func localConsoleExposedHeaders() []string {
	return []string{
		"X-Request-Id",
		"X-RadishMind-Route",
		"X-RadishMind-Provider-Attempts",
		"X-RadishMind-Fallback-Used",
	}
}

func localConsoleAllowedHeaders() []string {
	return []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"X-Request-Id",
		localIdentityActiveTenantHeader,
		localIdentityCSRFHeader,
		controlPlaneReadDevIdentityHeader,
		controlPlaneReadDevTenantHeader,
		controlPlaneReadDevSubjectHeader,
		controlPlaneReadDevScopesHeader,
		controlPlaneReadDevAuditHeader,
		savedWorkflowDraftDevWorkspaceHeader,
		activeWorkspaceHeader,
		controlPlaneReadDevMembershipHeader,
		controlPlaneReadDevMembershipPermHeader,
		savedWorkflowDraftDevApplicationHeader,
		applicationDraftDevWorkspaceHeader,
		applicationDraftDevApplicationHeader,
		promptApplicationTemplateDevWorkspaceHeader,
		promptApplicationTemplateDevApplicationHeader,
		agentCopilotProfileDevWorkspaceHeader,
		agentCopilotProfileDevApplicationHeader,
		agentCopilotRuntimeWorkspaceHeader,
		agentCopilotRuntimeApplicationHeader,
		promptApplicationRuntimeWorkspaceHeader,
		promptApplicationRuntimeApplicationHeader,
		applicationPublishDevWorkspaceHeader,
		applicationPublishDevApplicationHeader,
		adminProviderRouteDevWorkspaceHeader,
		adminProviderRouteDevEnvironmentHeader,
		gatewayModelPricingEnvironmentHeader,
		gatewayRequestQuotaEnvironmentHeader,
		applicationEvaluationEnvironmentHeader,
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
