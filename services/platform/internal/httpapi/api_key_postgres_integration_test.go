//go:build postgres_integration

package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"radishmind.local/services/platform/internal/config"
	apikeymigrations "radishmind.local/services/platform/migrations/api_key_records"
)

func TestAPIKeyPostgresLifecycleAuthenticationAndRestart(t *testing.T) {
	databaseURL := postgresIntegrationDatabaseURL(t)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" {
		t.Fatal("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER is required")
	}
	runtimeDatabaseURL := postgresIntegrationDatabaseURLForCredentials(t, runtimeUser, os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adminPool, err := apikeymigrations.OpenPool(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, adminPool)
	resetPostgresAPIKeySchema(t, ctx, adminPool)
	preparePostgresIntegrationRuntimeRole(t, ctx, adminPool, runtimeUser)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		resetPostgresAPIKeySchema(t, cleanupContext, adminPool)
		adminPool.Close()
	})
	state, err := apikeymigrations.Apply(ctx, adminPool)
	if err != nil || state.MigrationState != apikeymigrations.MigrationStateApplied {
		t.Fatalf("apply API key migration: state=%#v err=%v", state, err)
	}
	if repeated, repeatErr := apikeymigrations.Apply(ctx, adminPool); repeatErr != nil || repeated.MigrationChecksum != state.MigrationChecksum {
		t.Fatalf("repeat API key migration: state=%#v err=%v", repeated, repeatErr)
	}

	runtimePool, err := apikeymigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, ddlErr := runtimePool.Exec(ctx, "CREATE TABLE api_key_runtime_must_not_create (id integer)"); ddlErr == nil {
		t.Fatal("API key runtime role must not have schema DDL permission")
	}
	applicationRepository := newMemoryApplicationCatalogRepository()
	managementContext := apiKeyTestContext("subject_postgres_owner")
	managementContext.RequestContext = ctx
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	repository := newPostgresAPIKeyRepository(runtimePool)
	service := newAPIKeyService(repository, applicationRepository)
	service.newID = func() (string, error) { return "key_2222222222222222", nil }
	issued := service.Create(managementContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "PostgreSQL Gateway key",
		Scopes: []string{"models:read", "responses:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil || issued.CredentialToken == "" {
		t.Fatalf("issue PostgreSQL API key: %#v", issued)
	}
	var payloadText string
	var digestLength int
	if err := adminPool.QueryRow(ctx, "SELECT sanitized_record_payload::text, octet_length(credential_digest) FROM api_key_records WHERE api_key_id=$1", issued.Record.APIKeyID).Scan(&payloadText, &digestLength); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(payloadText, issued.CredentialToken) || strings.Contains(payloadText, "credential_digest") || digestLength != 32 {
		t.Fatalf("PostgreSQL API key material boundary failed: digest_length=%d payload=%s", digestLength, payloadText)
	}

	server := &Server{
		config:                       config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource},
		applicationCatalogRepository: applicationRepository, apiKeyRepository: repository,
	}
	authenticate := func() gatewayAPIKeyAuthenticationResult {
		request := newAPIKeyGatewayIntegrationRequest(issued.CredentialToken)
		return server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read")
	}
	if result := authenticate(); result.FailureCode != "" || result.RequestContext.ConsumerRef != "api_key:"+issued.Record.APIKeyID {
		t.Fatalf("authenticate PostgreSQL API key: %#v", result)
	}
	read := service.Read(managementContext, issued.Record.APIKeyID)
	if read.Record == nil || read.Record.LastUsedAt == nil || read.Record.RecordVersion != 1 {
		t.Fatalf("PostgreSQL authentication must atomically update last_used_at without lifecycle version drift: %#v", read)
	}

	var wait sync.WaitGroup
	results := make(chan string, 8)
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- authenticate().FailureCode
		}()
	}
	revoked := service.Revoke(managementContext, issued.Record.APIKeyID, 1)
	wait.Wait()
	close(results)
	if revoked.FailureCode != "" || revoked.Record == nil || revoked.Record.RecordVersion != 2 {
		t.Fatalf("revoke PostgreSQL API key: %#v", revoked)
	}
	for failure := range results {
		if failure != "" && failure != APIKeyFailureRevoked {
			t.Fatalf("PostgreSQL authentication/revoke race produced unstable failure: %s", failure)
		}
	}
	if result := authenticate(); result.FailureCode != APIKeyFailureRevoked {
		t.Fatalf("revoked PostgreSQL API key was accepted: %#v", result)
	}

	runtimePool.Close()
	reopenedPool, err := apikeymigrations.OpenPool(ctx, runtimeDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	restarted := newAPIKeyService(newPostgresAPIKeyRepository(reopenedPool), applicationRepository)
	restored := restarted.Read(managementContext, issued.Record.APIKeyID)
	if restored.Record == nil || restored.Record.EffectiveState != apiKeyLifecycleRevoked || restored.Record.LastUsedAt == nil {
		t.Fatalf("API key restart recovery failed: %#v", restored)
	}
	reopenedPool.Close()

	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true, APIKeyLifecycleDevHTTPEnabled: true, APIKeyLifecycleDevWriteEnabled: true,
		APIKeyStoreMode: "postgres_dev_test", APIKeyDatabaseURL: runtimeDatabaseURL, APIKeyDatabaseTimeout: 5 * time.Second,
	}
	selected, closeSelected, err := newAPIKeyRepositoryFromConfig(cfg)
	if err != nil || selected == nil || closeSelected == nil {
		t.Fatalf("API key factory did not select PostgreSQL: repository=%T err=%v", selected, err)
	}
	closeSelected()

	if _, err := adminPool.Exec(ctx, "UPDATE api_key_schema_versions SET migration_checksum='sha256:mismatch' WHERE component=$1", apikeymigrations.Component); err != nil {
		t.Fatal(err)
	}
	if selected, closeStore, factoryErr := newAPIKeyRepositoryFromConfig(cfg); factoryErr == nil || selected != nil || closeStore != nil {
		t.Fatalf("API key factory accepted marker mismatch: repository=%T close=%v err=%v", selected, closeStore != nil, factoryErr)
	}
	if _, err := adminPool.Exec(ctx, "UPDATE api_key_schema_versions SET migration_checksum=$1 WHERE component=$2", apikeymigrations.ExpectedChecksum(), apikeymigrations.Component); err != nil {
		t.Fatal(err)
	}
	if _, err := apikeymigrations.RollbackForDevTest(ctx, adminPool); err != nil {
		t.Fatal(err)
	}
	if reapplied, reapplyErr := apikeymigrations.Apply(ctx, adminPool); reapplyErr != nil || reapplied.MigrationState != apikeymigrations.MigrationStateApplied {
		t.Fatalf("reapply API key migration: state=%#v err=%v", reapplied, reapplyErr)
	}
}

func newAPIKeyGatewayIntegrationRequest(token string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func resetPostgresAPIKeySchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS api_key_records"); err != nil {
		t.Fatalf("reset API key records table: %v", err)
	}
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS api_key_schema_versions"); err != nil {
		t.Fatalf("reset API key migration marker: %v", err)
	}
}
