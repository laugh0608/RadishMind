package workbench

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"radishmind.local/services/platform/internal/config"
)

func TestWorkbenchWebRoutingAndFileBoundary(t *testing.T) {
	assets := t.TempDir()
	if err := os.Mkdir(filepath.Join(assets, "assets"), 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"index.html": "<main>workbench</main>", "workbench-build.json": `{"runtime_profile":"local-workbench"}`, "assets/app-123.js": "export {}", ".secret": "private"} {
		if err := os.WriteFile(filepath.Join(assets, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	out := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(out, []byte("must-not-be-served"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(out, filepath.Join(assets, "assets", "escape.js")); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Provider: "mock", APIKey: "private-key", SQLiteDevDatabasePath: "private-db", Workbench: config.WorkbenchConfig{
		PublicOrigin: "http://127.0.0.1:14100", TenantRef: "tenant_local", WorkspaceID: "workspace_local", Environment: "test", WebAssetsPath: assets}}
	apiCalls := 0
	handler, closeAssets, err := WebHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiCalls++; http.Error(w, "api-route", 404) }))
	if err != nil {
		t.Fatal(err)
	}
	defer closeAssets()
	for _, test := range []struct {
		path, origin string
		status       int
		body         string
	}{
		{"/", "", 200, "workbench"}, {"/applications", "", 200, "workbench"}, {"/assets/app-123.js", "", 200, "export"},
		{"/v1/unknown", "", 404, "api-route"}, {"/workbench-config.json", "", 200, "schema_version"},
		{"/.secret", "", 404, "not_found"}, {"/assets/../.secret", "", 404, "not_found"},
		{"/assets/escape.js", "", 404, "not_found"}, {"/workbench-build.json", "", 404, "not_found"},
		{"/v1/control-plane/audit", "", 404, "workbench_surface_unavailable"},
		{"/v1/user-workspace/usage/quota-summary", "", 404, "workbench_surface_unavailable"},
		{"/workbench-config.json", "http://localhost:14100", 403, "workbench_origin_mismatch"},
	} {
		request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:14100"+test.path, nil)
		request.Header.Set("Origin", test.origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status || !strings.Contains(response.Body.String(), test.body) {
			t.Fatalf("%s: %d %s", test.path, response.Code, response.Body.String())
		}
		if test.path == "/workbench-config.json" && test.status == 200 {
			if strings.Contains(response.Body.String(), "private") || strings.Contains(response.Body.String(), assets) {
				t.Fatal("public configuration leaked private values")
			}
			var public PublicConfig
			if err := json.Unmarshal(response.Body.Bytes(), &public); err != nil || public.WorkspaceID != "workspace_local" {
				t.Fatal("invalid public configuration")
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("configuration cached")
			}
		}
	}
	if apiCalls != 1 {
		t.Fatalf("unexpected API forwarding: %d", apiCalls)
	}
	request := httptest.NewRequest(http.MethodGet, "http://attacker.invalid:14100/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatal("accepted wrong Host")
	}
}

func TestWorkbenchRejectsLegacyWebBuild(t *testing.T) {
	assets := t.TempDir()
	if err := os.WriteFile(filepath.Join(assets, "index.html"), []byte("legacy"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := CheckAssets(assets); err == nil {
		t.Fatal("accepted legacy build without runtime configuration")
	}
}
