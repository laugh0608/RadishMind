package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const LocalWorkbenchProfile = "local-workbench"

type WorkbenchConfig struct {
	PublicOrigin  string `json:"public_origin"`
	TenantRef     string `json:"tenant_ref"`
	WorkspaceID   string `json:"workspace_id"`
	Environment   string `json:"environment"`
	WebAssetsPath string `json:"web_assets_path"`
}

func applyRuntimeProfile(cfg *Config, fileProfile *string) error {
	profile := ""
	if fileProfile != nil {
		profile = strings.TrimSpace(*fileProfile)
		if profile == "" {
			return errors.New("runtime_profile cannot be empty")
		}
	}
	if value, exists := os.LookupEnv("RADISHMIND_RUNTIME_PROFILE"); exists {
		if profile != "" && profile != value {
			return errors.New("runtime profile conflicts with configuration file")
		}
		profile = strings.TrimSpace(value)
	}
	if profile == "" {
		return nil
	}
	if profile != LocalWorkbenchProfile {
		return errors.New("unsupported runtime_profile")
	}
	cfg.RuntimeProfile = profile
	cfg.ListenAddr = "127.0.0.1:4100"
	cfg.LocalPersistenceMode = "sqlite_dev"
	cfg.SQLiteDevDatabasePath = "data/radishmind.db"
	cfg.Model = "radishmind-workbench-mock"
	cfg.ControlPlaneReadDevAuthEnabled = true
	cfg.ControlPlaneReadAuthMode = "local_session_dev_test"
	cfg.LocalIdentityDevHTTPEnabled = true
	cfg.LocalIdentityCookieSecure = false
	cfg.GatewayAuthMode = "api_key_dev_test"
	cfg.GatewayProviderRouteSource = "static_config"
	cfg.Workbench = WorkbenchConfig{
		PublicOrigin: "http://127.0.0.1:4100", TenantRef: "tenant_local", WorkspaceID: "workspace_local",
		Environment: "development", WebAssetsPath: "../../apps/radishmind-web/dist",
	}
	for _, enabled := range workbenchRequiredGates(cfg) {
		*enabled = true
	}
	return nil
}

func workbenchRequiredGates(cfg *Config) map[string]*bool {
	return map[string]*bool{
		"local identity": &cfg.LocalIdentityDevHTTPEnabled, "session auth": &cfg.ControlPlaneReadDevAuthEnabled,
		"catalog HTTP": &cfg.ApplicationCatalogDevHTTPEnabled, "catalog write": &cfg.ApplicationCatalogDevWriteEnabled,
		"draft HTTP": &cfg.ApplicationDraftDevHTTPEnabled, "draft write": &cfg.ApplicationDraftDevWriteEnabled,
		"publish HTTP": &cfg.ApplicationPublishDevHTTPEnabled, "publish write": &cfg.ApplicationPublishDevWriteEnabled,
		"template HTTP": &cfg.PromptTemplateDevHTTPEnabled, "template write": &cfg.PromptTemplateDevWriteEnabled,
		"prompt HTTP": &cfg.PromptApplicationRuntimeDevHTTPEnabled, "prompt write": &cfg.PromptApplicationRuntimeDevWriteEnabled,
		"workflow draft HTTP": &cfg.WorkflowSavedDraftDevHTTPEnabled, "workflow draft write": &cfg.WorkflowSavedDraftDevWriteEnabled,
		"workflow release": &cfg.WorkflowDefinitionReleaseDevEnabled, "workflow executor": &cfg.WorkflowExecutorDevEnabled,
		"application session": &cfg.ApplicationSessionDevEnabled,
		"gateway history":     &cfg.GatewayRequestHistoryDevEnabled,
		"API key HTTP":        &cfg.APIKeyLifecycleDevHTTPEnabled, "API key write": &cfg.APIKeyLifecycleDevWriteEnabled,
	}
}

func finalizeWorkbenchConfig(cfg *Config) error {
	if cfg.RuntimeProfile != LocalWorkbenchProfile {
		if cfg.APIKeyFile != "" || cfg.Workbench.PublicOrigin != "" {
			return errors.New("workbench fields require local-workbench runtime_profile")
		}
		return nil
	}
	if cfg.ConfigFile == "" {
		return errors.New("local-workbench requires an explicit RADISHMIND_PLATFORM_CONFIG file")
	}
	for name, target := range map[string]*string{
		"RADISHMIND_WORKBENCH_PUBLIC_ORIGIN":   &cfg.Workbench.PublicOrigin,
		"RADISHMIND_WORKBENCH_TENANT_REF":      &cfg.Workbench.TenantRef,
		"RADISHMIND_WORKBENCH_WORKSPACE_ID":    &cfg.Workbench.WorkspaceID,
		"RADISHMIND_WORKBENCH_ENVIRONMENT":     &cfg.Workbench.Environment,
		"RADISHMIND_WORKBENCH_WEB_ASSETS_PATH": &cfg.Workbench.WebAssetsPath,
		"RADISHMIND_PLATFORM_API_KEY_FILE":     &cfg.APIKeyFile,
	} {
		if value, exists := os.LookupEnv(name); exists {
			*target = strings.TrimSpace(value)
		}
	}
	if value, exists := os.LookupEnv("RADISHMIND_WORKBENCH_CONTAINER"); exists && value != "" {
		if value != "1" {
			return errors.New("RADISHMIND_WORKBENCH_CONTAINER must be 1 when explicitly selected")
		}
		cfg.WorkbenchContainer = true
	}
	absoluteConfig, err := filepath.Abs(cfg.ConfigFile)
	if err != nil {
		return errors.New("resolve workbench configuration path")
	}
	cfg.ConfigFile = absoluteConfig
	if strings.ContainsAny(cfg.PythonBinary, "/\\") && !filepath.IsAbs(cfg.PythonBinary) {
		cfg.PythonBinary = filepath.Join(filepath.Dir(absoluteConfig), cfg.PythonBinary)
	}
	for _, path := range []*string{&cfg.SQLiteDevDatabasePath, &cfg.Workbench.WebAssetsPath, &cfg.APIKeyFile} {
		if *path != "" && !filepath.IsAbs(*path) {
			*path = filepath.Join(filepath.Dir(absoluteConfig), *path)
		}
	}
	if cfg.APIKeyFile != "" {
		if cfg.APIKey != "" {
			return errors.New("api_key_file conflicts with RADISHMIND_PLATFORM_API_KEY")
		}
		info, err := os.Stat(cfg.APIKeyFile)
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("api_key_file must reference an accessible regular file")
		}
		file, err := os.Open(cfg.APIKeyFile)
		if err != nil {
			return errors.New("api_key_file is not readable")
		}
		if err := file.Close(); err != nil {
			return errors.New("close api_key_file")
		}
	}
	if cfg.FieldSources["local_identity_allowed_origin"] == configSourceEnv && cfg.LocalIdentityAllowedOrigin != cfg.Workbench.PublicOrigin {
		return errors.New("local identity allowed origin must match workbench public_origin")
	}
	cfg.LocalIdentityAllowedOrigin = cfg.Workbench.PublicOrigin
	return ValidateWorkbench(*cfg)
}

var workbenchScopePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

func ValidateWorkbench(cfg Config) error {
	if cfg.RuntimeProfile != LocalWorkbenchProfile {
		return nil
	}
	if cfg.LocalIdentityCookieSecure {
		return errors.New("loopback HTTP workbench requires the existing non-secure development cookie setting")
	}
	if cfg.LocalPersistenceMode != "sqlite_dev" || cfg.ControlPlaneReadAuthMode != "local_session_dev_test" ||
		cfg.GatewayAuthMode != "api_key_dev_test" || cfg.GatewayProviderRouteSource != "static_config" {
		return errors.New("local-workbench requires SQLite, local Session, API key gateway and static provider routing")
	}
	for name, enabled := range workbenchRequiredGates(&cfg) {
		if !*enabled {
			return fmt.Errorf("local-workbench requires %s", name)
		}
	}
	if cfg.LocalIdentityOIDCEnabled || cfg.AdminProviderRouteDevHTTPEnabled || cfg.AgentCopilotRuntimeDevHTTPEnabled ||
		cfg.AgentCopilotProfileDevHTTPEnabled || cfg.WorkflowToolActionDevEnabled || cfg.WorkflowHTTPToolExecutionDevEnabled ||
		cfg.WorkflowRAGSnapshotDevEnabled || cfg.WorkflowRAGExecutionDevEnabled || cfg.WorkflowRAGEvaluationDevEnabled ||
		cfg.WorkflowRAGPromotionDevEnabled || cfg.WorkflowRAGAppInvocationDevEnabled || cfg.WorkflowTemplateCatalogDevEnabled ||
		cfg.ApplicationEvaluationCampaignDevEnabled || cfg.ApplicationEvaluationScheduleRunnerDevEnabled ||
		cfg.GatewayProviderFallbackDevEnabled || cfg.GatewayRequestQuotaDevHTTPEnabled || cfg.GatewayRequestQuotaEnforcementDevEnabled ||
		cfg.GatewayModelPricingDevHTTPEnabled || cfg.GatewayModelPricingCaptureDevEnabled {
		return errors.New("local-workbench has an unsupported feature gate enabled")
	}
	host, port, err := net.SplitHostPort(cfg.ListenAddr)
	if err != nil || port == "" || (host != "127.0.0.1" && host != "::1" && host != "0.0.0.0") {
		return errors.New("workbench listen_addr must be loopback or the explicit container address 0.0.0.0")
	}
	if portNumber, err := strconv.Atoi(port); err != nil || portNumber < 1 || portNumber > 65535 {
		return errors.New("invalid workbench listen port")
	}
	if host == "0.0.0.0" && !cfg.WorkbenchContainer {
		return errors.New("0.0.0.0 is restricted to the explicit workbench container mode")
	}
	origin, err := url.Parse(cfg.Workbench.PublicOrigin)
	if err != nil || origin.Scheme != "http" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" ||
		(origin.Hostname() != "127.0.0.1" && origin.Hostname() != "::1") || origin.Port() == "" {
		return errors.New("workbench public_origin must be an HTTP loopback origin with an explicit port")
	}
	if portNumber, err := strconv.Atoi(origin.Port()); err != nil || portNumber < 1 || portNumber > 65535 || portNumber == 80 || strconv.Itoa(portNumber) != origin.Port() {
		return errors.New("invalid workbench public port")
	}
	if host != "0.0.0.0" && (origin.Hostname() != host || origin.Port() != port) {
		return errors.New("workbench public_origin must match listen_addr")
	}
	if !workbenchScopePattern.MatchString(cfg.Workbench.TenantRef) || !workbenchScopePattern.MatchString(cfg.Workbench.WorkspaceID) ||
		(cfg.Workbench.Environment != "development" && cfg.Workbench.Environment != "test") || cfg.Workbench.WebAssetsPath == "" || cfg.SQLiteDevDatabasePath == "" {
		return errors.New("workbench scope, environment, database and web assets paths are required")
	}
	if cfg.ProviderProfile != "" {
		return errors.New("workbench provider_profile is unsupported; configure provider/model explicitly")
	}
	switch cfg.Provider {
	case "mock":
		if cfg.APIKey != "" || cfg.APIKeyFile != "" || cfg.BaseURL != "" {
			return errors.New("mock workbench must not configure remote credentials or base_url")
		}
	case "openai-compatible":
		endpoint, err := url.Parse(cfg.BaseURL)
		if err != nil || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" ||
			cfg.Model == "" || cfg.Model == "radishmind-workbench-mock" || (cfg.APIKey == "" && cfg.APIKeyFile == "") {
			return errors.New("openai-compatible workbench requires explicit model, base_url and credential")
		}
	default:
		return errors.New("workbench provider must be mock or openai-compatible")
	}
	return nil
}

// ResolveWorkbenchCredential is called only immediately before starting the runtime.
func ResolveWorkbenchCredential(cfg *Config) error {
	if cfg.APIKeyFile == "" {
		return nil
	}
	file, err := os.Open(cfg.APIKeyFile)
	if err != nil {
		return errors.New("read workbench credential file")
	}
	defer file.Close()
	value, err := io.ReadAll(io.LimitReader(file, 16385))
	if err != nil || len(value) > 16384 || strings.TrimSpace(string(value)) == "" {
		return errors.New("workbench credential file is empty, unreadable or too large")
	}
	cfg.APIKey = strings.TrimSpace(string(value))
	return nil
}

func decodeWorkbenchDocument(content []byte, document *fileConfig) error {
	if err := rejectDuplicateJSONKeys(json.NewDecoder(bytes.NewReader(content))); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(document); err != nil {
		return fmt.Errorf("parse workbench configuration: %w", err)
	}
	return nil
}

func rejectDuplicateJSONKeys(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return errors.New("workbench configuration fields cannot be null")
	}
	delimiter, nested := token.(json.Delim)
	if !nested {
		return nil
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delimiter == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return errors.New("workbench configuration contains a duplicate JSON key")
			}
			seen[name] = true
		}
		if err := rejectDuplicateJSONKeys(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
