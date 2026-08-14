//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/config"
	gatewaymodelpricingmigrations "radishmind.local/services/platform/migrations/gateway_model_pricing"
	gatewayrequestmigrations "radishmind.local/services/platform/migrations/gateway_requests"
)

func TestGatewayModelPricingPostgresIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(
		t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := gatewaymodelpricingmigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresGatewayModelPricingSchema(t, ctx, adminPool)
	resetPostgresGatewayRequestSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresGatewayModelPricingSchema(t, cleanupContext, adminPool)
		resetPostgresGatewayRequestSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})

	missing, closeMissing, missingErr := newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayModelPricingStoreMode:       gatewayModelPricingStoreModePostgresDevTest,
		GatewayModelPricingDatabaseURL:     runtimeURL,
		GatewayModelPricingDatabaseTimeout: 5 * time.Second,
	}, nil)
	if missingErr == nil || missing != nil || closeMissing != nil {
		if closeMissing != nil {
			closeMissing()
		}
		t.Fatalf("PostgreSQL pricing selector fell back before migration: repository=%T err=%v", missing, missingErr)
	}
	state, err := gatewaymodelpricingmigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != gatewaymodelpricingmigrations.MigrationStateApplied {
		t.Fatalf("apply gateway model pricing migration: state=%+v err=%v", state, err)
	}
	if repeated, repeatErr := gatewaymodelpricingmigrations.Apply(ctx, adminPool); repeatErr != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat gateway model pricing migration: state=%+v err=%v", repeated, repeatErr)
	}
	requestState, err := gatewayrequestmigrations.Apply(ctx, adminPool)
	if err != nil || requestState.MigrationState != gatewayrequestmigrations.MigrationStateApplied {
		t.Fatalf("apply gateway request cost migration: state=%+v err=%v", requestState, err)
	}
	constrainPostgresGatewayModelPricingRuntimePrivileges(t, ctx, adminPool, runtimeUser)
	runtimePool, err := gatewaymodelpricingmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE gateway_model_pricing_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("gateway model pricing runtime role must not have schema DDL permission")
	}
	repository := newPostgresGatewayModelPricingRepository(runtimePool)
	pricingContext := gatewayModelPricingTestContext("model_postgres")
	pricingContext.RequestContext = ctx
	first := putGatewayModelPricingTestPolicy(t, repository, pricingContext, 0, 10, 20, "first PostgreSQL evidence")

	start := make(chan struct{})
	results := make(chan error, 8)
	var waitGroup sync.WaitGroup
	for index := 0; index < 8; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, updateErr := repository.PutRevision(pricingContext, GatewayModelPricingPolicyInput{
				ExpectedVersion: 1, Currency: GatewayModelPricingCurrency,
				InputPriceMicrosPerTokenUnit: 30, OutputPriceMicrosPerTokenUnit: 40,
				Reason: "concurrent PostgreSQL evidence",
			}, time.Now().UTC())
			results <- updateErr
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)
	winners, conflicts := 0, 0
	for updateErr := range results {
		switch {
		case updateErr == nil:
			winners++
		case errors.Is(updateErr, errGatewayModelPricingVersionConflict):
			conflicts++
		default:
			t.Fatalf("unexpected PostgreSQL pricing update error: %v", updateErr)
		}
	}
	if winners != 1 || conflicts != 7 {
		t.Fatalf("PostgreSQL pricing CAS was not linearized: winners=%d conflicts=%d", winners, conflicts)
	}
	if _, updateErr := runtimePool.Exec(ctx, `UPDATE gateway_model_pricing_revisions SET policy_digest='sha256:tampered'
        WHERE tenant_ref=$1 AND policy_id=$2 AND record_version=1`, pricingContext.TenantRef, first.PolicyID); updateErr == nil {
		t.Fatal("PostgreSQL append-only revision accepted an update")
	}
	requestContext := gatewayRequestTestContext()
	requestContext.RequestContext = ctx
	requestStore := newPostgresGatewayRequestStore(runtimePool)
	startedAt := time.Date(2026, 8, 12, 8, 0, 0, 0, time.UTC)
	record := gatewayRequestTestRecord(requestContext, "request_postgres_pricing_lineage", startedAt)
	record.SchemaVersion = gatewayRequestRecordSchemaVersionV2
	record.CostEstimate = gatewayRequestCostUnavailable(GatewayRequestCostNotApplicable, "provider_not_attempted")
	if err = requestStore.CreateRequest(requestContext, &record); err != nil {
		t.Fatalf("create PostgreSQL priced request: %v", err)
	}
	record.SelectionSource = "configured_provider"
	record.SelectedProvider = pricingContext.ProviderID
	record.SelectedProfile = pricingContext.ProfileID
	record.SelectedModel = pricingContext.ModelID
	record.Usage = gatewayModelPricingTestUsage(1_000_000, 500_000)
	record.CostEstimate = buildGatewayRequestCostEstimate(
		true, record.Usage, gatewayModelPricingSnapshotFromPolicy(first),
	)
	record.Status = GatewayRequestStatusSucceeded
	record.CompletedAt = startedAt.Add(time.Second).Format(time.RFC3339Nano)
	record.DurationMS = 1_000
	record.HTTPStatusCode = http.StatusOK
	if err = requestStore.UpdateRequest(requestContext, &record); err != nil {
		t.Fatalf("complete PostgreSQL priced request: %v", err)
	}
	runtimePool.Close()

	configured, closeRepository, err := newGatewayModelPricingRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayModelPricingStoreMode:       gatewayModelPricingStoreModePostgresDevTest,
		GatewayModelPricingDatabaseURL:     runtimeURL,
		GatewayModelPricingDatabaseTimeout: 5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("reopen configured PostgreSQL pricing repository: %v", err)
	}
	defer closeRepository()
	current, found, err := configured.ReadCurrent(pricingContext)
	if err != nil || !found || current.RecordVersion != 2 {
		t.Fatalf("restore PostgreSQL pricing current: policy=%+v found=%v err=%v", current, found, err)
	}
	retained, found, err := configured.ReadRevision(pricingContext, 1)
	if err != nil || !found || retained != first {
		t.Fatalf("restore PostgreSQL pricing revision: policy=%+v found=%v err=%v", retained, found, err)
	}
	reopenedRequests, err := gatewayrequestmigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatalf("reopen PostgreSQL request repository: %v", err)
	}
	defer reopenedRequests.Close()
	recoveredRequest, found, err := newPostgresGatewayRequestStore(reopenedRequests).ReadRequest(
		requestContext, "request_postgres_pricing_lineage",
	)
	if err != nil || !found || recoveredRequest.SchemaVersion != gatewayRequestRecordSchemaVersionV2 ||
		recoveredRequest.CostEstimate.Availability != GatewayRequestCostEstimated ||
		recoveredRequest.CostEstimate.EstimatedCostMicros == nil || *recoveredRequest.CostEstimate.EstimatedCostMicros != 20 ||
		recoveredRequest.CostEstimate.PricingPolicyVersion == nil || *recoveredRequest.CostEstimate.PricingPolicyVersion != 1 ||
		recoveredRequest.CostEstimate.PricingPolicyDigest != first.PolicyDigest || current.RecordVersion != 2 {
		t.Fatalf("PostgreSQL request-local pricing lineage was recalculated: found=%v record=%+v current=%+v err=%v", found, recoveredRequest, current, err)
	}
}

func constrainPostgresGatewayModelPricingRuntimePrivileges(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeUser string,
) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	for _, table := range []string{
		"gateway_model_pricing_schema_versions", "gateway_model_pricing_revisions", "gateway_model_pricing_current",
	} {
		if _, err := adminPool.Exec(ctx, "REVOKE ALL ON TABLE "+pgx.Identifier{table}.Sanitize()+" FROM "+quotedRuntimeUser); err != nil {
			t.Fatalf("revoke gateway model pricing runtime privileges: %v", err)
		}
	}
	statements := []string{
		"GRANT SELECT ON TABLE gateway_model_pricing_schema_versions TO " + quotedRuntimeUser,
		"GRANT SELECT, INSERT ON TABLE gateway_model_pricing_revisions TO " + quotedRuntimeUser,
		"GRANT SELECT, INSERT, UPDATE ON TABLE gateway_model_pricing_current TO " + quotedRuntimeUser,
	}
	for _, statement := range statements {
		if _, err := adminPool.Exec(ctx, statement); err != nil {
			t.Fatalf("grant gateway model pricing runtime privileges: %v", err)
		}
	}
}

func resetPostgresGatewayModelPricingSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, statement := range []string{
		"DROP TABLE IF EXISTS gateway_model_pricing_current",
		"DROP TABLE IF EXISTS gateway_model_pricing_revisions CASCADE",
		"DROP FUNCTION IF EXISTS gateway_model_pricing_revisions_append_only()",
		"DROP TABLE IF EXISTS gateway_model_pricing_schema_versions",
	} {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatalf("reset gateway model pricing schema: %v", err)
		}
	}
}
