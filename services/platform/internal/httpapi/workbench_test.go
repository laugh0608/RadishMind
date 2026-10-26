package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/workbench"
)

func TestWorkbenchEmptyRegistrationExplicitBootstrapAndRestart(t *testing.T) {
	for _, entry := range os.Environ() {
		name := strings.SplitN(entry, "=", 2)[0]
		if strings.HasPrefix(name, "RADISHMIND_") {
			t.Setenv(name, "")
		}
	}
	os.Unsetenv("RADISHMIND_RUNTIME_PROFILE")
	directory := t.TempDir()
	assets := filepath.Join(directory, "web")
	if err := os.Mkdir(assets, 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"index.html": "workbench", "workbench-build.json": `{"runtime_profile":"local-workbench"}`} {
		if err := os.WriteFile(filepath.Join(assets, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	document := map[string]any{"runtime_profile": "local-workbench", "listen_addr": "127.0.0.1:14100", "bridge_mode": "process_per_request",
		"workbench": map[string]string{"public_origin": "http://127.0.0.1:14100", "tenant_ref": "tenant_local", "workspace_id": "workspace_local", "environment": "test", "web_assets_path": assets}}
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	configurationPath := filepath.Join(directory, "platform.json")
	if err := os.WriteFile(configurationPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", configurationPath)
	cfg, err := config.LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewServerWithError(cfg, Options{TestOnly: true}); err == nil {
		t.Fatal("serve created missing database")
	}
	if err := workbench.Initialize(context.Background(), cfg.SQLiteDevDatabasePath); err != nil {
		t.Fatal(err)
	}
	server, err := NewServerWithError(cfg, Options{TestOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if _, err := NewServerWithError(cfg, Options{TestOnly: true}); err == nil {
		t.Fatal("second workbench opened active database")
	}
	register := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityRegisterRoute, map[string]any{
		"login_identifier": "workbench@example.test", "display_name": "Workbench User", "password": "a sufficiently long test password",
	}, nil)
	register.Host = "127.0.0.1:14100"
	register.Header.Set("Origin", cfg.Workbench.PublicOrigin)
	response := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, register)
	if response.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", response.Code, response.Body.String())
	}
	var registered localIdentityAuthenticationDocument
	decodeLocalIdentityHTTPResponse(t, response, &registered)
	cookies := activeLocalIdentityCookies(response.Result().Cookies())
	options := LocalIdentityDevTestBootstrapOptions{Workbench: true, StoreMode: "sqlite_dev", SQLiteDatabasePath: cfg.SQLiteDevDatabasePath, DatabaseTimeout: time.Second,
		TenantRef: cfg.Workbench.TenantRef, WorkspaceID: cfg.Workbench.WorkspaceID, UserID: registered.Account.UserID, AuditRef: "audit:workbench-bootstrap"}
	if _, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(context.Background(), options); err == nil {
		t.Fatal("bootstrap bypassed active service lock")
	}
	server.Close()
	if _, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if _, err := BootstrapLocalIdentityWorkspaceAdministratorDevTest(context.Background(), options); err == nil {
		t.Fatal("bootstrap repeated")
	}
	server, err = NewServerWithError(cfg, Options{TestOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	request := httptest.NewRequest(http.MethodGet, cfg.Workbench.PublicOrigin+"/v1/user-workspace/applications", nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	request.Header.Set(localIdentityActiveTenantHeader, cfg.Workbench.TenantRef)
	request.Header.Set(activeWorkspaceHeader, cfg.Workbench.WorkspaceID)
	response = httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), "app_flow_copilot") {
		t.Fatalf("workspace after restart: %d %s", response.Code, response.Body.String())
	}
	var body struct {
		Items       []any   `json:"items"`
		FailureCode *string `json:"failure_code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || len(body.Items) != 0 || body.FailureCode != nil {
		t.Fatalf("empty real catalog: %s", response.Body.String())
	}
	server.Close()
	legacy := cfg
	legacy.RuntimeProfile = ""
	if _, err := NewServerWithError(legacy, Options{TestOnly: true}); err == nil {
		t.Fatal("legacy mode opened workbench directory")
	}
}
