package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	apikeymigrations "radishmind.local/services/platform/migrations/sqlite/api_key_records"
	applicationcatalogmigrations "radishmind.local/services/platform/migrations/sqlite/application_catalog_records"
)

func TestAPIKeySQLiteRepositoryContract(t *testing.T) {
	t.Run("issue read list and revoke", func(t *testing.T) {
		runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runAPIKeyServiceIssueReadListAndRevoke(t,
			newSQLiteAPIKeyRepository(runtime.DB()),
			newSQLiteApplicationCatalogRepository(runtime.DB()),
		)
	})
	t.Run("expiry filtering pagination and owner isolation", func(t *testing.T) {
		runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runAPIKeyServiceExpiryFilteringPaginationAndOwnerIsolation(t,
			newSQLiteAPIKeyRepository(runtime.DB()),
			newSQLiteApplicationCatalogRepository(runtime.DB()),
		)
	})
	t.Run("concurrent revoke has one winner", func(t *testing.T) {
		runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runAPIKeyConcurrentRevokeHasSingleWinner(t,
			newSQLiteAPIKeyRepository(runtime.DB()),
			newSQLiteApplicationCatalogRepository(runtime.DB()),
		)
	})
}

func TestAPIKeySQLiteGatewayAuthenticationAndRevokeRace(t *testing.T) {
	t.Run("northbound authentication and request history", func(t *testing.T) {
		runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runGatewayAPIKeyAuthenticationProtectsNorthboundAndWritesHistory(t,
			newSQLiteAPIKeyRepository(runtime.DB()),
			newSQLiteApplicationCatalogRepository(runtime.DB()),
		)
	})
	t.Run("expiry application archive and revoke race", func(t *testing.T) {
		runtime := openAPIKeySQLiteRuntime(t, filepath.Join(t.TempDir(), "radishmind.db"))
		runGatewayAPIKeyExpiryApplicationArchiveAndRevokeRace(t,
			newSQLiteAPIKeyRepository(runtime.DB()),
			newSQLiteApplicationCatalogRepository(runtime.DB()),
		)
	})
}

func TestAPIKeySQLiteRejectsTimeOutsideNanosecondRange(t *testing.T) {
	if _, err := apiKeyUnixNano("2500-01-01T00:00:00Z"); err == nil {
		t.Fatal("SQLite API key time outside the nanosecond range must be rejected")
	}
}

func TestAPIKeySQLiteRestartRecoveryNoFallbackAndCredentialBoundary(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "radishmind.db")
	firstRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   apiKeySQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("open first API key SQLite runtime: %v", err)
	}
	applicationRepository := newSQLiteApplicationCatalogRepository(firstRuntime.DB())
	keyRepository := newSQLiteAPIKeyRepository(firstRuntime.DB())
	managementContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, managementContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	issuedAt := time.Now().UTC().Add(-time.Hour)
	service := newAPIKeyService(keyRepository, applicationRepository)
	service.now = func() time.Time { return issuedAt }
	service.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := service.Create(managementContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "SQLite restart key",
		Scopes: []string{"models:read", "responses:invoke"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil || issued.CredentialToken == "" {
		t.Fatalf("issue SQLite API key: failure=%s record_present=%t credential_present=%t", issued.FailureCode, issued.Record != nil, issued.CredentialToken != "")
	}
	var payload string
	var digestLength int
	if err := firstRuntime.DB().QueryRowContext(context.Background(), `SELECT sanitized_record_payload, length(credential_digest)
        FROM api_key_records WHERE api_key_id=?`, issued.Record.APIKeyID).Scan(&payload, &digestLength); err != nil {
		t.Fatalf("inspect SQLite API key material boundary: %v", err)
	}
	if strings.Contains(payload, issued.CredentialToken) || strings.Contains(payload, "credential_digest") || digestLength != 32 {
		t.Fatalf("SQLite API key material boundary failed: digest_length=%d payload=%s", digestLength, payload)
	}

	server := &Server{
		config:                       config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource},
		applicationCatalogRepository: applicationRepository,
		apiKeyRepository:             keyRepository,
	}
	request := newAPIKeySQLiteGatewayRequest(issued.CredentialToken)
	authenticated := server.authenticateGatewayAPIKey(request, newRequestTrace(request, "/v1/models"), "models:read")
	if authenticated.FailureCode != "" || authenticated.RequestContext.ConsumerRef != "api_key:"+issued.Record.APIKeyID {
		t.Fatalf("authenticate SQLite API key: %#v", authenticated)
	}
	used := service.Read(managementContext, issued.Record.APIKeyID)
	if used.Record == nil || used.Record.LastUsedAt == nil || used.Record.RecordVersion != 1 {
		t.Fatalf("SQLite authentication must update last_used_at without version drift: %#v", used)
	}
	lastUsedAt, err := time.Parse(time.RFC3339Nano, *used.Record.LastUsedAt)
	if err != nil {
		t.Fatalf("parse SQLite last_used_at: %v", err)
	}
	unchanged, err := keyRepository.RecordSuccessfulAuthentication(context.Background(), issued.Record.APIKeyID, 1, lastUsedAt.Add(-time.Minute))
	if err != nil || unchanged.LastUsedAt == nil || *unchanged.LastUsedAt != *used.Record.LastUsedAt {
		t.Fatalf("SQLite last_used_at must be monotonic: record=%#v err=%v", unchanged, err)
	}

	if err := firstRuntime.Close(); err != nil {
		t.Fatalf("close first API key SQLite runtime: %v", err)
	}
	if result := service.Read(managementContext, issued.Record.APIKeyID); result.FailureCode != APIKeyFailureStoreUnavailable {
		t.Fatalf("closed API key store must not fall back to memory: %#v", result)
	}
	closedRequest := newAPIKeySQLiteGatewayRequest(issued.CredentialToken)
	if result := server.authenticateGatewayAPIKey(closedRequest, newRequestTrace(closedRequest, "/v1/models"), "models:read"); result.FailureCode != APIKeyFailureStoreUnavailable {
		t.Fatalf("closed API key store must stop Gateway authentication: %#v", result)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, issued.CredentialToken)

	secondRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   apiKeySQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("reopen API key SQLite runtime: %v", err)
	}
	restartedApplicationRepository := newSQLiteApplicationCatalogRepository(secondRuntime.DB())
	restartedKeyRepository := newSQLiteAPIKeyRepository(secondRuntime.DB())
	restartedService := newAPIKeyService(restartedKeyRepository, restartedApplicationRepository)
	restored := restartedService.Read(managementContext, issued.Record.APIKeyID)
	if restored.Record == nil || restored.Record.LastUsedAt == nil || restored.Record.RecordVersion != 1 || restored.Record.EffectiveState != apiKeyLifecycleActive {
		t.Fatalf("restore active API key after SQLite restart: %#v", restored)
	}
	otherOwner := managementContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if result := restartedService.Read(otherOwner, issued.Record.APIKeyID); result.FailureCode != APIKeyFailureNotFound {
		t.Fatalf("restart must preserve API key owner isolation: %#v", result)
	}
	restartedServer := &Server{
		config:                       config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource},
		applicationCatalogRepository: restartedApplicationRepository,
		apiKeyRepository:             restartedKeyRepository,
	}
	restartedRequest := newAPIKeySQLiteGatewayRequest(issued.CredentialToken)
	if result := restartedServer.authenticateGatewayAPIKey(restartedRequest, newRequestTrace(restartedRequest, "/v1/models"), "models:read"); result.FailureCode != "" {
		t.Fatalf("restarted SQLite API key was not accepted: %#v", result)
	}
	revoked := restartedService.Revoke(managementContext, issued.Record.APIKeyID, 1)
	if revoked.Record == nil || revoked.Record.RecordVersion != 2 || revoked.Record.EffectiveState != apiKeyLifecycleRevoked {
		t.Fatalf("revoke restored SQLite API key: %#v", revoked)
	}
	if err := secondRuntime.Close(); err != nil {
		t.Fatalf("close second API key SQLite runtime: %v", err)
	}
	assertSQLiteFilesExcludeMarker(t, databasePath, issued.CredentialToken)

	thirdRuntime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   apiKeySQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("reopen revoked API key SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = thirdRuntime.Close() })
	finalApplicationRepository := newSQLiteApplicationCatalogRepository(thirdRuntime.DB())
	finalKeyRepository := newSQLiteAPIKeyRepository(thirdRuntime.DB())
	finalService := newAPIKeyService(finalKeyRepository, finalApplicationRepository)
	finalRecord := finalService.Read(managementContext, issued.Record.APIKeyID)
	if finalRecord.Record == nil || finalRecord.Record.RecordVersion != 2 || finalRecord.Record.EffectiveState != apiKeyLifecycleRevoked || finalRecord.Record.LastUsedAt == nil {
		t.Fatalf("restore revoked API key after second SQLite restart: %#v", finalRecord)
	}
	finalServer := &Server{
		config:                       config.Config{GatewayAuthMode: gatewayAPIKeyAuthenticationSource},
		applicationCatalogRepository: finalApplicationRepository,
		apiKeyRepository:             finalKeyRepository,
	}
	finalRequest := newAPIKeySQLiteGatewayRequest(issued.CredentialToken)
	if result := finalServer.authenticateGatewayAPIKey(finalRequest, newRequestTrace(finalRequest, "/v1/models"), "models:read"); result.FailureCode != APIKeyFailureRevoked {
		t.Fatalf("revoked SQLite API key was accepted after restart: %#v", result)
	}
}

func openAPIKeySQLiteRuntime(t *testing.T, databasePath string) *sqlitedev.Runtime {
	t.Helper()
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   apiKeySQLiteMigrations(),
	})
	if err != nil {
		t.Fatalf("open API key SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	return runtime
}

func apiKeySQLiteMigrations() []sqlitedev.Migration {
	migrations := applicationcatalogmigrations.Migrations()
	return append(migrations, apikeymigrations.Migrations()...)
}

func newAPIKeySQLiteGatewayRequest(token string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}
