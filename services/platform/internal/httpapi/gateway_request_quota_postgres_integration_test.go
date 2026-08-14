//go:build postgres_integration

package httpapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/config"
	gatewayrequestquotamigrations "radishmind.local/services/platform/migrations/gateway_request_quotas"
)

func TestGatewayRequestQuotaPostgresIntegration(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := gatewayrequestquotamigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresGatewayRequestQuotaSchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresGatewayRequestQuotaSchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	missingRepository, closeMissingRepository, missingErr := newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayRequestQuotaStoreMode:       gatewayRequestQuotaStoreModePostgresDevTest,
		GatewayRequestQuotaDatabaseURL:     runtimeURL,
		GatewayRequestQuotaDatabaseTimeout: 5 * time.Second,
	}, nil)
	if missingErr == nil || missingRepository != nil || closeMissingRepository != nil {
		if closeMissingRepository != nil {
			closeMissingRepository()
		}
		t.Fatalf("PostgreSQL quota selector fell back before migration: repository=%T err=%v", missingRepository, missingErr)
	}
	state, err := gatewayrequestquotamigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != gatewayrequestquotamigrations.MigrationStateApplied {
		t.Fatalf("apply gateway request quota migration: state=%+v err=%v", state, err)
	}
	if repeated, repeatErr := gatewayrequestquotamigrations.Apply(ctx, adminPool); repeatErr != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat gateway request quota migration: state=%+v err=%v", repeated, repeatErr)
	}
	constrainPostgresGatewayRequestQuotaRuntimePrivileges(t, ctx, adminPool, runtimeUser)
	runtimePool, err := gatewayrequestquotamigrations.OpenPool(ctx, runtimeURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE gateway_request_quota_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("gateway request quota runtime role must not have schema DDL permission")
	}
	repository := newPostgresGatewayRequestQuotaRepository(runtimePool)
	quotaContext := testGatewayRequestQuotaContext("app-postgres")
	quotaContext.RequestContext = ctx
	now := time.Date(2026, 8, 9, 16, 0, 0, 0, time.UTC)
	if _, err := repository.PutPolicy(quotaContext, 0, 1, now); err != nil {
		t.Fatalf("put PostgreSQL quota policy: %v", err)
	}
	start := make(chan struct{})
	results := make(chan error, 8)
	var waitGroup sync.WaitGroup
	for attempt := 0; attempt < 8; attempt++ {
		waitGroup.Add(1)
		go func(attempt int) {
			defer waitGroup.Done()
			<-start
			_, admissionErr := repository.AdmitProviderAttempt(quotaContext, testQuotaAdmission(fmt.Sprintf("request-pg-%d", attempt), now))
			results <- admissionErr
		}(attempt)
	}
	close(start)
	waitGroup.Wait()
	close(results)
	admitted, exceeded := 0, 0
	for admissionErr := range results {
		switch {
		case admissionErr == nil:
			admitted++
		case errors.Is(admissionErr, errGatewayRequestQuotaExceeded):
			exceeded++
		default:
			t.Fatalf("unexpected PostgreSQL admission error: %v", admissionErr)
		}
	}
	if admitted != 1 || exceeded != 7 {
		t.Fatalf("PostgreSQL quota was not linearized: admitted=%d exceeded=%d", admitted, exceeded)
	}
	runtimePool.Close()

	configuredRepository, closeRepository, err := newGatewayRequestQuotaRepositoryFromConfigWithSQLiteRuntime(config.Config{
		GatewayRequestQuotaStoreMode: "postgres_dev_test", GatewayRequestQuotaDatabaseURL: runtimeURL,
		GatewayRequestQuotaDatabaseTimeout: 5 * time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("reopen configured PostgreSQL quota repository: %v", err)
	}
	defer closeRepository()
	usage, found, err := configuredRepository.ReadUsage(quotaContext, "2026-08-09")
	if err != nil || !found || usage.AdmittedRequestCount != 1 || usage.RemainingRequestCount != 0 {
		t.Fatalf("restore PostgreSQL quota usage: found=%t usage=%+v err=%v", found, usage, err)
	}
}

func constrainPostgresGatewayRequestQuotaRuntimePrivileges(
	t *testing.T,
	ctx context.Context,
	adminPool *pgxpool.Pool,
	runtimeUser string,
) {
	t.Helper()
	quotedRuntimeUser := pgx.Identifier{runtimeUser}.Sanitize()
	for _, table := range []string{
		"gateway_request_quota_schema_versions",
		"gateway_request_quota_policies",
		"gateway_request_quota_usage",
		"gateway_request_quota_admissions",
	} {
		if _, err := adminPool.Exec(ctx, "REVOKE ALL ON TABLE "+pgx.Identifier{table}.Sanitize()+" FROM "+quotedRuntimeUser); err != nil {
			t.Fatalf("revoke gateway request quota runtime privileges: %v", err)
		}
	}
	statements := []string{
		"GRANT SELECT ON TABLE gateway_request_quota_schema_versions TO " + quotedRuntimeUser,
		"GRANT SELECT, INSERT, UPDATE ON TABLE gateway_request_quota_policies TO " + quotedRuntimeUser,
		"GRANT SELECT, INSERT, UPDATE ON TABLE gateway_request_quota_usage TO " + quotedRuntimeUser,
		"GRANT SELECT, INSERT, UPDATE ON TABLE gateway_request_quota_admissions TO " + quotedRuntimeUser,
	}
	for _, statement := range statements {
		if _, err := adminPool.Exec(ctx, statement); err != nil {
			t.Fatalf("grant gateway request quota runtime privileges: %v", err)
		}
	}
}

func resetPostgresGatewayRequestQuotaSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, statement := range []string{
		"DROP TABLE IF EXISTS gateway_request_quota_admissions",
		"DROP TABLE IF EXISTS gateway_request_quota_usage",
		"DROP TABLE IF EXISTS gateway_request_quota_policies",
		"DROP TABLE IF EXISTS gateway_request_quota_schema_versions",
	} {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatalf("reset gateway request quota schema: %v", err)
		}
	}
}
