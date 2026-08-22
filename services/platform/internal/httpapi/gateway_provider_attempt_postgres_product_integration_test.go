//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	adminproviderroutemigrations "radishmind.local/services/platform/migrations/admin_provider_routes"
	apikeymigrations "radishmind.local/services/platform/migrations/api_key_records"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/application_catalog_records"
	gatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/gateway_model_pricing"
	gatewayrequestquotamigrations "radishmind.local/services/platform/migrations/gateway_request_quotas"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
)

func TestGatewayProviderAttemptPostgresProductThreeProtocolsRestartAndRuntimeRole(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	adminPool, err := gatewayrequestmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open Provider Attempt PostgreSQL migration pool: %v", err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresProviderAttemptProductSchemas(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		resetPostgresProviderAttemptProductSchemas(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	applyPostgresProviderAttemptProductMigrations(t, ctx, adminPool)
	constrainPostgresGatewayRequestQuotaRuntimePrivileges(t, ctx, adminPool, runtimeUser)
	constrainPostgresGatewayModelPricingRuntimePrivileges(t, ctx, adminPool, runtimeUser)

	runtimePool, err := gatewayrequestmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatalf("open Provider Attempt PostgreSQL runtime pool: %v", err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE provider_attempt_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("Provider Attempt PostgreSQL runtime role unexpectedly executed DDL")
	}

	catalogRepository := newPostgresApplicationCatalogRepository(runtimePool)
	apiKeyRepository := newPostgresAPIKeyRepository(runtimePool)
	routeRepository := newPostgresAdminProviderRouteRepository(runtimePool)
	historyStore := newPostgresGatewayRequestStore(runtimePool)
	quotaRepository := newPostgresGatewayRequestQuotaRepository(runtimePool)
	pricingRepository := newPostgresGatewayModelPricingRepository(runtimePool)

	const owner = "subject_provider_attempt_postgres"
	const applicationID = "app_dddddddddddddddd"
	catalogContext := applicationCatalogTestContext(owner)
	catalogContext.RequestContext = ctx
	catalogService := newApplicationCatalogService(catalogRepository)
	catalogService.newID = func() (string, error) { return applicationID, nil }
	createdApplication := catalogService.Create(catalogContext, ApplicationCatalogCreateInput{
		DisplayName:     "Provider Attempt PostgreSQL product",
		Description:     "PostgreSQL product continuity for explicit unary fallback.",
		ApplicationKind: "workflow_copilot",
	})
	if createdApplication.FailureCode != "" || createdApplication.Record == nil {
		t.Fatalf("create Provider Attempt PostgreSQL application: %#v", createdApplication)
	}

	apiKeyContext := apiKeyTestContext(owner)
	apiKeyContext.RequestContext = ctx
	keyService := newAPIKeyService(apiKeyRepository, catalogRepository)
	keyService.newID = func() (string, error) { return "key_dddddddddddddddd", nil }
	keyService.now = func() time.Time { return time.Now().UTC().Truncate(time.Second) }
	issued := keyService.Create(apiKeyContext, APIKeyCreateInput{
		ApplicationID: applicationID,
		DisplayName:   "Provider Attempt PostgreSQL key",
		Scopes:        []string{"chat:invoke", "responses:invoke", "messages:invoke"},
		ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil || issued.CredentialToken == "" {
		t.Fatalf("issue Provider Attempt PostgreSQL API key: failure=%s record_present=%t token_present=%t",
			issued.FailureCode, issued.Record != nil, issued.CredentialToken != "")
	}

	primary := adminProviderRouteBridgeProfile()
	primary.ResolvedModel = "mock-model-primary"
	secondary := primary
	secondary.Profile = "mock-secondary"
	secondary.NormalizedProfile = "mock-secondary"
	secondary.ResolvedModel = "mock-model-secondary"
	testBridge := &gatewayProviderAttemptScriptedBridge{inventory: bridge.ProviderInventory{
		Profiles: []bridge.ProviderProfileDescription{primary, secondary},
	}}
	routeService := newAdminProviderRouteService(
		routeRepository,
		bridgeAdminProviderInventoryResolver{bridge: testBridge},
	)
	routeContext := adminProviderRouteTestContext()
	routeContext.RequestContext = ctx
	routeContext.TenantRef = catalogContext.TenantRef
	routeContext.WorkspaceID = catalogContext.WorkspaceID
	routeContext.ActorRef = owner
	routeInput := adminProviderRouteV2TestDraftInput(0)
	routeInput.ConfigurationID = "gateway-default"
	routeInput.ModelRoutes = []AdminModelRouteDefinition{
		gatewayProviderAttemptExecutorRoute("chat", "chat_completions"),
		gatewayProviderAttemptExecutorRoute("responses", "responses"),
		gatewayProviderAttemptExecutorRoute("messages", "messages"),
	}
	if result := routeService.PutDraft(routeContext, routeInput); result.FailureCode != "" {
		t.Fatalf("create Provider Attempt PostgreSQL route draft: %#v", result)
	}
	if result := routeService.CreateCandidate(routeContext, AdminProviderRouteCandidateInput{
		ConfigurationID:       routeInput.ConfigurationID,
		CandidateID:           "candidate-provider-attempt-postgres",
		ExpectedDraftRevision: 1,
	}); result.FailureCode != "" {
		t.Fatalf("create Provider Attempt PostgreSQL route candidate: %#v", result)
	}
	if result := routeService.ReviewCandidate(
		routeContext,
		routeInput.ConfigurationID,
		"candidate-provider-attempt-postgres",
		AdminProviderRouteReviewInput{
			ExpectedReviewVersion: 0,
			Decision:              adminProviderRouteDecisionApprove,
			Reason:                "Approve the PostgreSQL three-protocol fallback route.",
		},
	); result.FailureCode != "" {
		t.Fatalf("review Provider Attempt PostgreSQL route candidate: %#v", result)
	}
	activated := routeService.Activate(routeContext, AdminProviderRouteActivationInput{
		ConfigurationID:    routeInput.ConfigurationID,
		CandidateID:        "candidate-provider-attempt-postgres",
		ExpectedGeneration: 0,
		Action:             adminProviderRouteActionActivate,
		Reason:             "Activate the PostgreSQL three-protocol fallback route.",
	})
	if activated.FailureCode != "" || activated.Snapshot == nil || activated.Snapshot.Generation != 1 {
		t.Fatalf("activate Provider Attempt PostgreSQL route: %#v", activated)
	}

	quotaContext := GatewayRequestQuotaContext{
		RequestContext: ctx,
		TenantRef:      catalogContext.TenantRef,
		WorkspaceID:    catalogContext.WorkspaceID,
		Environment:    "test",
		ApplicationID:  applicationID,
		ActorRef:       owner,
		RequestID:      "request-provider-attempt-postgres-quota",
		AuditRef:       "audit-provider-attempt-postgres-quota",
	}
	if _, err := quotaRepository.PutPolicy(quotaContext, 0, 20, time.Now().UTC()); err != nil {
		t.Fatalf("create Provider Attempt PostgreSQL quota policy: %v", err)
	}
	for _, profileID := range []string{"mock-primary", "mock-secondary"} {
		pricingContext := GatewayModelPricingContext{
			RequestContext: ctx,
			TenantRef:      catalogContext.TenantRef,
			WorkspaceID:    catalogContext.WorkspaceID,
			Environment:    "test",
			ProviderID:     "mock",
			ProfileID:      profileID,
			ModelID:        "mock-model",
			ActorRef:       owner,
			RequestID:      "request-pricing-" + profileID,
			AuditRef:       "audit-pricing-" + profileID,
		}
		putGatewayModelPricingTestPolicy(
			t, pricingRepository, pricingContext, 0, 10, 20,
			"Freeze PostgreSQL Provider Attempt pricing.",
		)
	}

	server := &Server{
		bridge: testBridge,
		config: config.Config{
			BridgeTimeout:                            time.Second,
			GatewayAuthMode:                          gatewayAPIKeyAuthenticationSource,
			GatewayRequestHistoryDevEnabled:          true,
			GatewayProviderRouteSource:               "admin_snapshot_dev_test",
			GatewayProviderRouteEnvironment:          "test",
			GatewayProviderRouteConfigurationID:      "gateway-default",
			GatewayProviderFallbackDevEnabled:        true,
			GatewayRequestQuotaEnforcementDevEnabled: true,
			GatewayRequestQuotaEnvironment:           "test",
			GatewayModelPricingCaptureDevEnabled:     true,
			GatewayModelPricingEnvironment:           "test",
		},
		applicationCatalogRepository:   catalogRepository,
		apiKeyRepository:               apiKeyRepository,
		providerRouteSnapshotProvider:  adminProviderRouteSnapshotProvider{repository: routeRepository},
		gatewayRequestHistoryStore:     historyStore,
		gatewayRequestHistoryStoreMode: gatewayRequestStoreModePostgresDevTest,
		gatewayRequestQuotaRepository:  quotaRepository,
		gatewayModelPricingRepository:  pricingRepository,
	}
	historyContext := GatewayRequestContext{
		RequestContext: ctx,
		TenantRef:      catalogContext.TenantRef,
		WorkspaceID:    catalogContext.WorkspaceID,
		ConsumerRef:    "api_key:" + issued.Record.APIKeyID,
		ApplicationID:  applicationID,
		SubjectRef:     owner,
		ScopeGrants:    []string{"gateway_requests:read"},
		AuditContext:   "api-key-dev-test",
		Source:         gatewayAPIKeyAuthenticationSource,
		RequestID:      "request-provider-attempt-postgres-history",
		AuditRef:       "audit-provider-attempt-postgres-history",
	}
	requestIDs := make([]string, 0, 3)
	for _, unaryRoute := range gatewayProviderAttemptUnaryRoutes(server) {
		requestID := "request-postgres-fallback-" + unaryRoute.name
		requestIDs = append(requestIDs, requestID)
		testBridge.reset(
			gatewayProviderAttemptFailedResult(t, bridge.ProviderFailureTemporarilyUnavailable, bridge.ProviderFallbackEligible),
			gatewayProviderAttemptSucceededResult("PostgreSQL fallback succeeded"),
		)
		recorder := invokePostgresProviderAttemptUnary(
			t, unaryRoute, issued.CredentialToken, requestID,
		)
		if recorder.Code != http.StatusOK ||
			recorder.Header().Get("X-RadishMind-Provider-Attempts") != "2" ||
			recorder.Header().Get("X-RadishMind-Fallback-Used") != "true" ||
			len(testBridge.callOptions()) != 2 {
			t.Fatalf("%s PostgreSQL fallback drifted: status=%d headers=%v calls=%d body=%s",
				unaryRoute.name, recorder.Code, recorder.Header(), len(testBridge.callOptions()), recorder.Body.String())
		}
		record, found, err := historyStore.ReadRequest(historyContext, requestID)
		if err != nil || !found || record.SchemaVersion != gatewayRequestRecordSchemaVersionV3 ||
			record.Status != GatewayRequestStatusSucceeded || record.ProviderAttemptCount != 2 ||
			!record.FallbackUsed || record.ProviderAttemptPlan.Protocol != unaryRoute.protocol ||
			len(record.ProviderAttempts) != 2 ||
			record.ProviderAttempts[0].Status != GatewayProviderAttemptStatusFailed ||
			record.ProviderAttempts[1].Status != GatewayProviderAttemptStatusSucceeded ||
			record.ProviderAttemptPlan.Targets[0].PricingSnapshot.PricingPolicyVersion != 1 ||
			record.ProviderAttemptPlan.Targets[1].PricingSnapshot.PricingPolicyVersion != 1 {
			t.Fatalf("%s PostgreSQL history drifted: found=%t record=%#v err=%v",
				unaryRoute.name, found, record, err)
		}
	}

	periodStart := time.Now().UTC().Format("2006-01-02")
	usage, found, err := quotaRepository.ReadUsage(quotaContext, periodStart)
	if err != nil || !found || usage.AdmittedRequestCount != 6 || usage.RemainingRequestCount != 14 {
		t.Fatalf("PostgreSQL per-attempt quota drifted: found=%t usage=%#v err=%v", found, usage, err)
	}
	var persistedHistory string
	if err := adminPool.QueryRow(ctx, `SELECT COALESCE(string_agg(sanitized_request_record::text, ''), '')
		FROM gateway_request_records`).Scan(&persistedHistory); err != nil {
		t.Fatalf("read Provider Attempt PostgreSQL privacy evidence: %v", err)
	}
	if strings.Contains(persistedHistory, "PostgreSQL private request body") ||
		strings.Contains(persistedHistory, issued.CredentialToken) {
		t.Fatal("Provider Attempt PostgreSQL history persisted request or credential material")
	}

	runtimePool.Close()
	if _, _, err := historyStore.ReadRequest(historyContext, requestIDs[0]); !errors.Is(err, errGatewayRequestStoreUnavailable) {
		t.Fatalf("closed PostgreSQL history store did not fail closed: %v", err)
	}
	reopened, err := gatewayrequestmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatalf("reopen Provider Attempt PostgreSQL runtime pool: %v", err)
	}
	defer reopened.Close()
	reopenedHistory := newPostgresGatewayRequestStore(reopened)
	for _, requestID := range requestIDs {
		restored, found, err := reopenedHistory.ReadRequest(historyContext, requestID)
		if err != nil || !found || restored.Status != GatewayRequestStatusSucceeded ||
			restored.ProviderAttemptCount != 2 || !restored.FallbackUsed {
			t.Fatalf("restore Provider Attempt PostgreSQL history %s: found=%t record=%#v err=%v",
				requestID, found, restored, err)
		}
	}
	restoredSnapshot, err := newPostgresAdminProviderRouteRepository(reopened).ReadActiveSnapshot(
		routeContext, routeInput.ConfigurationID,
	)
	if err != nil || restoredSnapshot.Generation != 1 ||
		len(restoredSnapshot.Configuration.ModelRoutes) != 3 {
		t.Fatalf("restore Provider Attempt PostgreSQL route: snapshot=%#v err=%v", restoredSnapshot, err)
	}
	restoredUsage, found, err := newPostgresGatewayRequestQuotaRepository(reopened).ReadUsage(
		quotaContext, periodStart,
	)
	if err != nil || !found || restoredUsage.AdmittedRequestCount != 6 {
		t.Fatalf("restore Provider Attempt PostgreSQL quota: found=%t usage=%#v err=%v",
			found, restoredUsage, err)
	}
}

func invokePostgresProviderAttemptUnary(
	t *testing.T,
	route gatewayProviderAttemptUnaryRoute,
	token string,
	requestID string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := strings.Replace(route.body, "hello", "PostgreSQL private request body", 1)
	body = strings.TrimSuffix(body, "}") + `,"radishmind":{"fallback_mode":"allow_configured"}}`
	request := httptest.NewRequest(http.MethodPost, route.path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Request-Id", requestID)
	recorder := httptest.NewRecorder()
	route.handle(recorder, request)
	return recorder
}

func applyPostgresProviderAttemptProductMigrations(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
) {
	t.Helper()
	if state, err := applicationcatalogmigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != applicationcatalogmigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt application catalog migration: state=%#v err=%v", state, err)
	}
	if state, err := apikeymigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != apikeymigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt API key migration: state=%#v err=%v", state, err)
	}
	if state, err := adminproviderroutemigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != adminproviderroutemigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt route migration: state=%#v err=%v", state, err)
	}
	if state, err := gatewayrequestmigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != gatewayrequestmigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt history migration: state=%#v err=%v", state, err)
	}
	if state, err := gatewayrequestquotamigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != gatewayrequestquotamigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt quota migration: state=%#v err=%v", state, err)
	}
	if state, err := gatewaymodelpricingmigrations.Apply(ctx, adminPool); err != nil ||
		state.MigrationState != gatewaymodelpricingmigrations.MigrationStateApplied {
		t.Fatalf("apply Provider Attempt pricing migration: state=%#v err=%v", state, err)
	}
}

func resetPostgresProviderAttemptProductSchemas(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
) {
	t.Helper()
	resetPostgresGatewayRequestQuotaSchema(t, ctx, adminPool)
	resetPostgresGatewayModelPricingSchema(t, ctx, adminPool)
	resetPostgresGatewayRequestSchema(t, ctx, adminPool)
	if _, err := adminproviderroutemigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatalf("reset Provider Attempt route schema: %v", err)
	}
	resetPostgresAPIKeySchema(t, ctx, adminPool)
	if _, err := applicationcatalogmigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatalf("reset Provider Attempt application catalog schema: %v", err)
	}
}
