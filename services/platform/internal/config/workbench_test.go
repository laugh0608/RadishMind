package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func workbenchFile(t *testing.T, document string) string {
	t.Helper()
	for _, entry := range os.Environ() {
		name := strings.SplitN(entry, "=", 2)[0]
		if strings.HasPrefix(name, "RADISHMIND_") {
			t.Setenv(name, "")
		}
	}
	path := filepath.Join(t.TempDir(), "platform.json")
	if err := os.WriteFile(path, []byte(document), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RADISHMIND_PLATFORM_CONFIG", path)
	// An absent profile override and an empty override intentionally differ.
	os.Unsetenv("RADISHMIND_RUNTIME_PROFILE")
	return path
}

func TestWorkbenchProfileAndOverrides(t *testing.T) {
	path := workbenchFile(t, `{"runtime_profile":"local-workbench","model":"explicit-mock"}`)
	t.Setenv("RADISHMIND_PLATFORM_LISTEN_ADDR", "127.0.0.1:14100")
	t.Setenv("RADISHMIND_WORKBENCH_PUBLIC_ORIGIN", "http://127.0.0.1:14100")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model != "explicit-mock" || cfg.SQLiteDevDatabasePath != filepath.Join(filepath.Dir(path), "data/radishmind.db") ||
		cfg.LocalIdentityAllowedOrigin != "http://127.0.0.1:14100" || !cfg.WorkflowExecutorDevEnabled {
		t.Fatalf("incorrect workbench configuration: %#v", cfg)
	}
	if err := ValidateServerStart(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Dir(cfg.SQLiteDevDatabasePath)); !os.IsNotExist(err) {
		t.Fatal("config load created data")
	}
}

func TestWorkbenchRejectsInvalidDocuments(t *testing.T) {
	for name, document := range map[string]string{
		"unknown":           `{"runtime_profile":"local-workbench","mdoel":"typo"}`,
		"duplicate":         `{"runtime_profile":"local-workbench","provider":"mock","provider":"mock"}`,
		"nested duplicate":  `{"runtime_profile":"local-workbench","workbench":{"public_origin":"x","public_origin":"y"}}`,
		"null path":         `{"runtime_profile":"local-workbench","sqlite_dev_database_path":null}`,
		"wrong type":        `{"runtime_profile":"local-workbench","sqlite_dev_database_path":42}`,
		"secret":            `{"runtime_profile":"local-workbench","api_key":"do-not-print"}`,
		"memory":            `{"runtime_profile":"local-workbench","local_persistence_mode":"memory_dev"}`,
		"remote origin":     `{"runtime_profile":"local-workbench","workbench":{"public_origin":"http://example.com:4100"}}`,
		"remote incomplete": `{"runtime_profile":"local-workbench","provider":"openai-compatible"}`,
	} {
		t.Run(name, func(t *testing.T) {
			workbenchFile(t, document)
			if _, err := LoadFromEnv(); err == nil {
				t.Fatal("accepted invalid configuration")
			} else if strings.Contains(err.Error(), "do-not-print") {
				t.Fatal("secret exposed")
			}
		})
	}
}

func TestWorkbenchRejectsGateAndScopeOverrides(t *testing.T) {
	for name, value := range map[string]string{
		"RADISHMIND_LOCAL_IDENTITY_DEV_HTTP":      "false",
		"RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE": "dev_headers",
		"RADISHMIND_WORKFLOW_RAG_EXECUTION_DEV":   "true",
		"RADISHMIND_WORKBENCH_PUBLIC_ORIGIN":      "http://127.0.0.1:7000",
		"RADISHMIND_WORKBENCH_TENANT_REF":         "",
	} {
		t.Run(name, func(t *testing.T) {
			workbenchFile(t, `{"runtime_profile":"local-workbench"}`)
			t.Setenv(name, value)
			if _, err := LoadFromEnv(); err == nil {
				t.Fatal("accepted invalid override")
			}
		})
	}
}

func TestWorkbenchCredentialReference(t *testing.T) {
	path := workbenchFile(t, `{"runtime_profile":"local-workbench","provider":"openai-compatible","model":"example","base_url":"http://127.0.0.1:18100/v1","api_key_file":"key"}`)
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "key"), []byte("secret-value\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "" || cfg.SanitizedSummary().CredentialState != "configured" {
		t.Fatal("credential loaded early or summary incorrect")
	}
	if err := ResolveWorkbenchCredential(&cfg); err != nil || cfg.APIKey != "secret-value" {
		t.Fatal("credential not resolved")
	}
	t.Setenv("RADISHMIND_PLATFORM_API_KEY", "conflicting-secret")
	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("accepted conflicting credential")
	}
}

func TestWorkbenchRejectsNonCanonicalBrowserOriginPorts(t *testing.T) {
	for _, port := range []string{"80", "04100"} {
		t.Run(port, func(t *testing.T) {
			workbenchFile(t, `{"runtime_profile":"local-workbench"}`)
			t.Setenv("RADISHMIND_PLATFORM_LISTEN_ADDR", "127.0.0.1:"+port)
			t.Setenv("RADISHMIND_WORKBENCH_PUBLIC_ORIGIN", "http://127.0.0.1:"+port)
			if _, err := LoadFromEnv(); err == nil {
				t.Fatal("accepted an origin the browser would normalize to a different value")
			}
		})
	}
}
