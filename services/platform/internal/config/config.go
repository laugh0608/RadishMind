package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListenAddr                  = ":7000"
	defaultReadHeaderTimeout           = 5 * time.Second
	defaultWriteTimeout                = 30 * time.Second
	defaultBridgeTimeout               = 30 * time.Second
	defaultBridgeMode                  = "stdio_pool"
	defaultBridgeWorkerCount           = 4
	defaultBridgeQueueSize             = 64
	defaultBridgeHandshake             = 5 * time.Second
	defaultDraftDBTimeout              = 5 * time.Second
	defaultApplicationDraftDBTimeout   = 5 * time.Second
	defaultApplicationPublishDBTimeout = 5 * time.Second
	defaultApplicationCatalogDBTimeout = 5 * time.Second
	defaultAPIKeyDBTimeout             = 5 * time.Second
	defaultRunDBTimeout                = 5 * time.Second
	defaultGatewayRequestDBTimeout     = 5 * time.Second
	defaultPythonBinary                = "python3"
	defaultBridgeScript                = "scripts/run-platform-bridge.py"
	defaultProvider                    = "mock"
	defaultDraftStoreMode              = "memory_dev"
	defaultApplicationDraftStoreMode   = "memory_dev"
	defaultApplicationPublishStoreMode = "memory_dev"
	defaultApplicationCatalogStoreMode = "memory_dev"
	defaultAPIKeyStoreMode             = "memory_dev"
	defaultGatewayAuthMode             = "dev_headers"
	defaultRunStoreMode                = "memory_dev"
	defaultGatewayRequestStoreMode     = "memory_dev"
	defaultLocalPersistenceMode        = "memory_dev"
	defaultSQLiteDevDatabasePath       = "var/sqlite-dev/radishmind.db"
	defaultControlPlaneReadAuthMode    = ""
	defaultControlPlaneReadStoreMode   = "fake_store_dev"
	defaultControlPlaneReadDBTimeout   = 5 * time.Second
	defaultOIDCDiscoveryTimeout        = 3 * time.Second
	defaultOIDCJWKSMaxAge              = 5 * time.Minute
	defaultOIDCJWKSHardExpiry          = 15 * time.Minute
	defaultOIDCRotationOverlap         = 5 * time.Minute
	defaultOIDCClockSkew               = 30 * time.Second
	defaultOIDCMaxTokenLifetime        = 10 * time.Minute
	defaultOIDCMaxResponseBytes        = 256 * 1024
	defaultOIDCMaxKeys                 = 32
)

const (
	configSourceDefault = "default"
	configSourceFile    = "file"
	configSourceEnv     = "env"
)

type Config struct {
	ListenAddr                           string
	ReadHeaderTimeout                    time.Duration
	WriteTimeout                         time.Duration
	BridgeTimeout                        time.Duration
	BridgeMode                           string
	BridgeWorkerCount                    int
	BridgeQueueCapacity                  int
	BridgeHandshakeTimeout               time.Duration
	ControlPlaneReadDevAuthEnabled       bool
	ControlPlaneReadAuthMode             string
	ControlPlaneReadTestIssuer           string
	ControlPlaneReadTestAudience         string
	ControlPlaneReadTestPublicKeyPEM     string
	ControlPlaneReadOIDCIssuer           string
	ControlPlaneReadOIDCDiscoveryURL     string
	ControlPlaneReadOIDCAudience         string
	ControlPlaneReadOIDCMappingVersion   string
	ControlPlaneReadOIDCEvidenceRef      string
	ControlPlaneReadOIDCSubjectClaim     string
	ControlPlaneReadOIDCTenantClaim      string
	ControlPlaneReadOIDCPermissionClaim  string
	ControlPlaneReadOIDCTenantPermission string
	ControlPlaneReadOIDCAuditPermission  string
	ControlPlaneReadOIDCAlgorithms       string
	ControlPlaneReadOIDCJWKSOrigin       string
	ControlPlaneReadOIDCDiscoveryTimeout time.Duration
	ControlPlaneReadOIDCJWKSMaxAge       time.Duration
	ControlPlaneReadOIDCJWKSHardExpiry   time.Duration
	ControlPlaneReadOIDCRotationOverlap  time.Duration
	ControlPlaneReadOIDCClockSkew        time.Duration
	ControlPlaneReadOIDCMaxTokenLifetime time.Duration
	ControlPlaneReadOIDCMaxResponseBytes int
	ControlPlaneReadOIDCMaxKeys          int
	ControlPlaneReadStoreMode            string
	ControlPlaneReadDatabaseURL          string
	ControlPlaneReadDatabaseTimeout      time.Duration
	WorkflowSavedDraftDevHTTPEnabled     bool
	WorkflowSavedDraftDevWriteEnabled    bool
	ApplicationDraftDevHTTPEnabled       bool
	ApplicationDraftDevWriteEnabled      bool
	ApplicationDraftStoreMode            string
	ApplicationDraftDatabaseURL          string
	ApplicationDraftDatabaseTimeout      time.Duration
	ApplicationPublishDevHTTPEnabled     bool
	ApplicationPublishDevWriteEnabled    bool
	ApplicationPublishStoreMode          string
	ApplicationPublishDatabaseURL        string
	ApplicationPublishDatabaseTimeout    time.Duration
	ApplicationCatalogDevHTTPEnabled     bool
	ApplicationCatalogDevWriteEnabled    bool
	ApplicationCatalogStoreMode          string
	ApplicationCatalogDatabaseURL        string
	ApplicationCatalogDatabaseTimeout    time.Duration
	APIKeyLifecycleDevHTTPEnabled        bool
	APIKeyLifecycleDevWriteEnabled       bool
	APIKeyStoreMode                      string
	APIKeyDatabaseURL                    string
	APIKeyDatabaseTimeout                time.Duration
	GatewayAuthMode                      string
	WorkflowExecutorDevEnabled           bool
	WorkflowDiagnosticsDevEnabled        bool
	GatewayRequestHistoryDevEnabled      bool
	GatewayRequestStoreMode              string
	GatewayRequestDatabaseURL            string
	GatewayRequestDatabaseTimeout        time.Duration
	LocalPersistenceMode                 string
	SQLiteDevDatabasePath                string
	WorkflowSavedDraftStoreMode          string
	WorkflowSavedDraftDatabaseURL        string
	WorkflowSavedDraftDatabaseTimeout    time.Duration
	WorkflowRunStoreMode                 string
	WorkflowRunDatabaseURL               string
	WorkflowRunDatabaseTimeout           time.Duration
	PythonBinary                         string
	BridgeScript                         string
	Provider                             string
	ProviderProfile                      string
	Model                                string
	BaseURL                              string
	APIKey                               string
	Temperature                          float64
	ConfigFile                           string
	FieldSources                         map[string]string
}

type ConfigSummary struct {
	ListenAddr                           string            `json:"listen_addr"`
	ControlPlaneReadDevAuthEnabled       bool              `json:"control_plane_read_dev_auth_enabled"`
	ControlPlaneReadAuthMode             string            `json:"control_plane_read_auth_mode"`
	ControlPlaneReadTestConfigured       bool              `json:"control_plane_read_signed_test_configured"`
	ControlPlaneReadOIDCConfigured       bool              `json:"control_plane_read_oidc_integration_configured"`
	ControlPlaneReadOIDCMappingVersion   string            `json:"control_plane_read_oidc_mapping_version,omitempty"`
	ControlPlaneReadOIDCEvidenceRef      string            `json:"control_plane_read_oidc_evidence_ref,omitempty"`
	ControlPlaneReadStoreMode            string            `json:"control_plane_read_store_mode"`
	ControlPlaneReadDatabaseConfigured   bool              `json:"control_plane_read_database_configured"`
	WorkflowSavedDraftDevHTTPEnabled     bool              `json:"workflow_saved_draft_dev_http_enabled"`
	WorkflowSavedDraftDevWriteEnabled    bool              `json:"workflow_saved_draft_dev_write_enabled"`
	ApplicationDraftDevHTTPEnabled       bool              `json:"application_draft_dev_http_enabled"`
	ApplicationDraftDevWriteEnabled      bool              `json:"application_draft_dev_write_enabled"`
	ApplicationDraftStoreMode            string            `json:"application_draft_store_mode"`
	ApplicationDraftDatabaseConfigured   bool              `json:"application_draft_database_configured"`
	ApplicationPublishDevHTTPEnabled     bool              `json:"application_publish_dev_http_enabled"`
	ApplicationPublishDevWriteEnabled    bool              `json:"application_publish_dev_write_enabled"`
	ApplicationPublishStoreMode          string            `json:"application_publish_store_mode"`
	ApplicationPublishDatabaseConfigured bool              `json:"application_publish_database_configured"`
	ApplicationCatalogDevHTTPEnabled     bool              `json:"application_catalog_dev_http_enabled"`
	ApplicationCatalogDevWriteEnabled    bool              `json:"application_catalog_dev_write_enabled"`
	ApplicationCatalogStoreMode          string            `json:"application_catalog_store_mode"`
	ApplicationCatalogDatabaseConfigured bool              `json:"application_catalog_database_configured"`
	APIKeyLifecycleDevHTTPEnabled        bool              `json:"api_key_lifecycle_dev_http_enabled"`
	APIKeyLifecycleDevWriteEnabled       bool              `json:"api_key_lifecycle_dev_write_enabled"`
	APIKeyStoreMode                      string            `json:"api_key_store_mode"`
	APIKeyDatabaseConfigured             bool              `json:"api_key_database_configured"`
	GatewayAuthMode                      string            `json:"gateway_auth_mode"`
	WorkflowExecutorDevEnabled           bool              `json:"workflow_executor_dev_enabled"`
	WorkflowDiagnosticsDevEnabled        bool              `json:"workflow_diagnostics_dev_enabled"`
	GatewayRequestHistoryDevEnabled      bool              `json:"gateway_request_history_dev_enabled"`
	GatewayRequestStoreMode              string            `json:"gateway_request_store_mode"`
	GatewayRequestDatabaseConfigured     bool              `json:"gateway_request_database_configured"`
	LocalPersistenceMode                 string            `json:"local_persistence_mode"`
	LocalPersistenceConfigured           bool              `json:"local_persistence_configured"`
	SQLiteDevDatabaseConfigured          bool              `json:"sqlite_dev_database_configured"`
	SQLiteDevSchemaStatus                string            `json:"sqlite_dev_schema_status"`
	LocalPersistenceComponentsConsistent bool              `json:"local_persistence_components_consistent"`
	WorkflowSavedDraftStoreMode          string            `json:"workflow_saved_draft_store_mode"`
	WorkflowSavedDraftDatabaseConfigured bool              `json:"workflow_saved_draft_database_configured"`
	WorkflowRunStoreMode                 string            `json:"workflow_run_store_mode"`
	WorkflowRunDatabaseConfigured        bool              `json:"workflow_run_database_configured"`
	Provider                             string            `json:"provider"`
	Profile                              string            `json:"profile"`
	Model                                string            `json:"model"`
	ModelConfigured                      bool              `json:"model_configured"`
	BaseURLConfigured                    bool              `json:"base_url_configured"`
	CredentialState                      string            `json:"credential_state"`
	Timeouts                             map[string]string `json:"timeouts"`
	PythonBridge                         PythonBridge      `json:"python_bridge"`
	Temperature                          float64           `json:"temperature"`
	RequiredFields                       []string          `json:"required_fields"`
	MissingRequiredFields                []string          `json:"missing_required_fields"`
	SecretFields                         []string          `json:"secret_fields"`
	ConfigFile                           ConfigFileSummary `json:"config_file"`
	FieldSources                         map[string]string `json:"field_sources"`
	Sanitized                            bool              `json:"sanitized"`
}

type PythonBridge struct {
	PythonBinary     string `json:"python_binary"`
	Script           string `json:"script"`
	Mode             string `json:"mode"`
	WorkerCount      int    `json:"worker_count"`
	QueueCapacity    int    `json:"queue_capacity"`
	HandshakeTimeout string `json:"handshake_timeout"`
}

type ConfigFileSummary struct {
	Path       string `json:"path"`
	Configured bool   `json:"configured"`
	Loaded     bool   `json:"loaded"`
	Source     string `json:"source"`
}

type fileConfig struct {
	ListenAddr         *string          `json:"listen_addr"`
	ReadHeaderTimeout  *string          `json:"read_header_timeout"`
	WriteTimeout       *string          `json:"write_timeout"`
	BridgeTimeout      *string          `json:"bridge_timeout"`
	BridgeMode         *string          `json:"bridge_mode"`
	BridgeWorkerCount  *int             `json:"bridge_worker_count"`
	BridgeQueueSize    *int             `json:"bridge_queue_capacity"`
	BridgeHandshake    *string          `json:"bridge_handshake_timeout"`
	PythonBinary       *string          `json:"python_binary"`
	BridgeScript       *string          `json:"bridge_script"`
	Provider           *string          `json:"provider"`
	ProviderProfile    *string          `json:"provider_profile"`
	Model              *string          `json:"model"`
	BaseURL            *string          `json:"base_url"`
	APIKey             *string          `json:"api_key"`
	Temperature        *json.RawMessage `json:"temperature"`
	WorkflowDraftStore *string          `json:"workflow_saved_draft_store"`
	LocalPersistence   *string          `json:"local_persistence_mode"`
	SQLiteDatabasePath *string          `json:"sqlite_dev_database_path"`
}

func LoadFromEnv() (Config, error) {
	cfg := defaultConfig()

	configFile := strings.TrimSpace(os.Getenv("RADISHMIND_PLATFORM_CONFIG"))
	if configFile != "" {
		if err := applyConfigFile(&cfg, configFile); err != nil {
			return Config{}, err
		}
		cfg.ConfigFile = configFile
		cfg.FieldSources["config_file"] = configSourceEnv
	}

	if err := applyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	if err := validateBridgeRuntimeConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		ListenAddr:                           defaultListenAddr,
		ReadHeaderTimeout:                    defaultReadHeaderTimeout,
		WriteTimeout:                         defaultWriteTimeout,
		BridgeTimeout:                        defaultBridgeTimeout,
		BridgeMode:                           defaultBridgeMode,
		BridgeWorkerCount:                    defaultBridgeWorkerCount,
		BridgeQueueCapacity:                  defaultBridgeQueueSize,
		BridgeHandshakeTimeout:               defaultBridgeHandshake,
		ControlPlaneReadAuthMode:             defaultControlPlaneReadAuthMode,
		ControlPlaneReadStoreMode:            defaultControlPlaneReadStoreMode,
		ControlPlaneReadDatabaseTimeout:      defaultControlPlaneReadDBTimeout,
		ControlPlaneReadOIDCDiscoveryTimeout: defaultOIDCDiscoveryTimeout,
		ControlPlaneReadOIDCJWKSMaxAge:       defaultOIDCJWKSMaxAge,
		ControlPlaneReadOIDCJWKSHardExpiry:   defaultOIDCJWKSHardExpiry,
		ControlPlaneReadOIDCRotationOverlap:  defaultOIDCRotationOverlap,
		ControlPlaneReadOIDCClockSkew:        defaultOIDCClockSkew,
		ControlPlaneReadOIDCMaxTokenLifetime: defaultOIDCMaxTokenLifetime,
		ControlPlaneReadOIDCMaxResponseBytes: defaultOIDCMaxResponseBytes,
		ControlPlaneReadOIDCMaxKeys:          defaultOIDCMaxKeys,
		PythonBinary:                         defaultPythonBinary,
		BridgeScript:                         defaultBridgeScript,
		Provider:                             defaultProvider,
		ProviderProfile:                      "",
		Model:                                "",
		BaseURL:                              "",
		APIKey:                               "",
		Temperature:                          0,
		WorkflowSavedDraftStoreMode:          defaultDraftStoreMode,
		WorkflowSavedDraftDatabaseTimeout:    defaultDraftDBTimeout,
		ApplicationDraftStoreMode:            defaultApplicationDraftStoreMode,
		ApplicationDraftDatabaseTimeout:      defaultApplicationDraftDBTimeout,
		ApplicationPublishStoreMode:          defaultApplicationPublishStoreMode,
		ApplicationPublishDatabaseTimeout:    defaultApplicationPublishDBTimeout,
		ApplicationCatalogStoreMode:          defaultApplicationCatalogStoreMode,
		ApplicationCatalogDatabaseTimeout:    defaultApplicationCatalogDBTimeout,
		APIKeyStoreMode:                      defaultAPIKeyStoreMode,
		APIKeyDatabaseTimeout:                defaultAPIKeyDBTimeout,
		GatewayAuthMode:                      defaultGatewayAuthMode,
		WorkflowRunStoreMode:                 defaultRunStoreMode,
		WorkflowRunDatabaseTimeout:           defaultRunDBTimeout,
		GatewayRequestStoreMode:              defaultGatewayRequestStoreMode,
		GatewayRequestDatabaseTimeout:        defaultGatewayRequestDBTimeout,
		SQLiteDevDatabasePath:                defaultSQLiteDevDatabasePath,
		FieldSources: map[string]string{
			"listen_addr":                           configSourceDefault,
			"read_header_timeout":                   configSourceDefault,
			"write_timeout":                         configSourceDefault,
			"bridge_timeout":                        configSourceDefault,
			"bridge_mode":                           configSourceDefault,
			"bridge_worker_count":                   configSourceDefault,
			"bridge_queue_capacity":                 configSourceDefault,
			"bridge_handshake_timeout":              configSourceDefault,
			"control_plane_read_dev_auth":           configSourceDefault,
			"control_plane_read_auth_mode":          configSourceDefault,
			"control_plane_read_store":              configSourceDefault,
			"control_plane_read_database":           configSourceDefault,
			"control_plane_read_database_timeout":   configSourceDefault,
			"workflow_saved_draft_dev_http":         configSourceDefault,
			"workflow_saved_draft_dev_write":        configSourceDefault,
			"application_draft_dev_http":            configSourceDefault,
			"application_draft_dev_write":           configSourceDefault,
			"application_draft_store":               configSourceDefault,
			"application_draft_database":            configSourceDefault,
			"application_draft_database_timeout":    configSourceDefault,
			"application_publish_dev_http":          configSourceDefault,
			"application_publish_dev_write":         configSourceDefault,
			"application_publish_store":             configSourceDefault,
			"application_publish_database":          configSourceDefault,
			"application_publish_database_timeout":  configSourceDefault,
			"application_catalog_dev_http":          configSourceDefault,
			"application_catalog_dev_write":         configSourceDefault,
			"application_catalog_store":             configSourceDefault,
			"application_catalog_database":          configSourceDefault,
			"application_catalog_database_timeout":  configSourceDefault,
			"api_key_store":                         configSourceDefault,
			"api_key_database":                      configSourceDefault,
			"api_key_database_timeout":              configSourceDefault,
			"gateway_auth_mode":                     configSourceDefault,
			"workflow_executor_dev":                 configSourceDefault,
			"workflow_diagnostics_dev":              configSourceDefault,
			"gateway_request_history_dev":           configSourceDefault,
			"gateway_request_store":                 configSourceDefault,
			"gateway_request_database":              configSourceDefault,
			"gateway_request_database_timeout":      configSourceDefault,
			"local_persistence_mode":                configSourceDefault,
			"sqlite_dev_database_path":              configSourceDefault,
			"workflow_saved_draft_store":            configSourceDefault,
			"workflow_saved_draft_database":         configSourceDefault,
			"workflow_saved_draft_database_timeout": configSourceDefault,
			"workflow_run_store":                    configSourceDefault,
			"workflow_run_database":                 configSourceDefault,
			"workflow_run_database_timeout":         configSourceDefault,
			"python_binary":                         configSourceDefault,
			"bridge_script":                         configSourceDefault,
			"provider":                              configSourceDefault,
			"profile":                               configSourceDefault,
			"model":                                 configSourceDefault,
			"base_url":                              configSourceDefault,
			"credential":                            configSourceDefault,
			"temperature":                           configSourceDefault,
			"config_file":                           configSourceDefault,
		},
	}
}

func applyConfigFile(cfg *Config, pathText string) error {
	content, err := os.ReadFile(pathText)
	if err != nil {
		return fmt.Errorf("read RADISHMIND_PLATFORM_CONFIG: %w", err)
	}
	var document fileConfig
	if err := json.Unmarshal(content, &document); err != nil {
		return fmt.Errorf("parse RADISHMIND_PLATFORM_CONFIG as json: %w", err)
	}
	if document.ListenAddr != nil {
		applyStringValue(&cfg.ListenAddr, *document.ListenAddr, cfg.FieldSources, "listen_addr", configSourceFile)
	}
	if document.ReadHeaderTimeout != nil {
		parsed, err := parseDurationValue("read_header_timeout", *document.ReadHeaderTimeout)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ReadHeaderTimeout, parsed, cfg.FieldSources, "read_header_timeout", configSourceFile)
	}
	if document.WriteTimeout != nil {
		parsed, err := parseDurationValue("write_timeout", *document.WriteTimeout)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.WriteTimeout, parsed, cfg.FieldSources, "write_timeout", configSourceFile)
	}
	if document.BridgeTimeout != nil {
		parsed, err := parseDurationValue("bridge_timeout", *document.BridgeTimeout)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.BridgeTimeout, parsed, cfg.FieldSources, "bridge_timeout", configSourceFile)
	}
	if document.BridgeMode != nil {
		applyStringValue(&cfg.BridgeMode, *document.BridgeMode, cfg.FieldSources, "bridge_mode", configSourceFile)
	}
	if document.BridgeWorkerCount != nil {
		applyIntValue(
			&cfg.BridgeWorkerCount,
			*document.BridgeWorkerCount,
			cfg.FieldSources,
			"bridge_worker_count",
			configSourceFile,
		)
	}
	if document.BridgeQueueSize != nil {
		applyIntValue(
			&cfg.BridgeQueueCapacity,
			*document.BridgeQueueSize,
			cfg.FieldSources,
			"bridge_queue_capacity",
			configSourceFile,
		)
	}
	if document.BridgeHandshake != nil {
		parsed, err := parseDurationValue("bridge_handshake_timeout", *document.BridgeHandshake)
		if err != nil {
			return err
		}
		applyDurationValue(
			&cfg.BridgeHandshakeTimeout,
			parsed,
			cfg.FieldSources,
			"bridge_handshake_timeout",
			configSourceFile,
		)
	}
	if document.PythonBinary != nil {
		applyStringValue(&cfg.PythonBinary, *document.PythonBinary, cfg.FieldSources, "python_binary", configSourceFile)
	}
	if document.BridgeScript != nil {
		applyStringValue(&cfg.BridgeScript, *document.BridgeScript, cfg.FieldSources, "bridge_script", configSourceFile)
	}
	if document.Provider != nil {
		applyStringValue(&cfg.Provider, *document.Provider, cfg.FieldSources, "provider", configSourceFile)
	}
	if document.ProviderProfile != nil {
		applyStringValue(&cfg.ProviderProfile, *document.ProviderProfile, cfg.FieldSources, "profile", configSourceFile)
	}
	if document.Model != nil {
		applyStringValue(&cfg.Model, *document.Model, cfg.FieldSources, "model", configSourceFile)
	}
	if document.BaseURL != nil {
		applyStringValue(&cfg.BaseURL, *document.BaseURL, cfg.FieldSources, "base_url", configSourceFile)
	}
	if document.APIKey != nil {
		applyStringValue(&cfg.APIKey, *document.APIKey, cfg.FieldSources, "credential", configSourceFile)
	}
	if document.Temperature != nil {
		parsed, err := parseJSONFloatValue("temperature", *document.Temperature)
		if err != nil {
			return err
		}
		cfg.Temperature = parsed
		cfg.FieldSources["temperature"] = configSourceFile
	}
	if document.WorkflowDraftStore != nil {
		applyStringValue(
			&cfg.WorkflowSavedDraftStoreMode,
			*document.WorkflowDraftStore,
			cfg.FieldSources,
			"workflow_saved_draft_store",
			configSourceFile,
		)
	}
	if document.LocalPersistence != nil {
		applyStringValue(
			&cfg.LocalPersistenceMode,
			*document.LocalPersistence,
			cfg.FieldSources,
			"local_persistence_mode",
			configSourceFile,
		)
	}
	if document.SQLiteDatabasePath != nil {
		applyStringValue(
			&cfg.SQLiteDevDatabasePath,
			*document.SQLiteDatabasePath,
			cfg.FieldSources,
			"sqlite_dev_database_path",
			configSourceFile,
		)
	}
	return nil
}

func applyEnvOverrides(cfg *Config) error {
	if value, ok := stringEnv("RADISHMIND_PLATFORM_LISTEN_ADDR"); ok {
		applyStringValue(&cfg.ListenAddr, value, cfg.FieldSources, "listen_addr", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ReadHeaderTimeout, parsed, cfg.FieldSources, "read_header_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_WRITE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_PLATFORM_WRITE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.WriteTimeout, parsed, cfg.FieldSources, "write_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_PLATFORM_BRIDGE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.BridgeTimeout, parsed, cfg.FieldSources, "bridge_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_MODE"); ok {
		applyStringValue(&cfg.BridgeMode, value, cfg.FieldSources, "bridge_mode", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT"); ok {
		parsed, err := parseIntValue("RADISHMIND_PLATFORM_BRIDGE_WORKER_COUNT", value)
		if err != nil {
			return err
		}
		applyIntValue(&cfg.BridgeWorkerCount, parsed, cfg.FieldSources, "bridge_worker_count", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY"); ok {
		parsed, err := parseIntValue("RADISHMIND_PLATFORM_BRIDGE_QUEUE_CAPACITY", value)
		if err != nil {
			return err
		}
		applyIntValue(&cfg.BridgeQueueCapacity, parsed, cfg.FieldSources, "bridge_queue_capacity", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_PLATFORM_BRIDGE_HANDSHAKE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(
			&cfg.BridgeHandshakeTimeout,
			parsed,
			cfg.FieldSources,
			"bridge_handshake_timeout",
			configSourceEnv,
		)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_PYTHON_BIN"); ok {
		applyStringValue(&cfg.PythonBinary, value, cfg.FieldSources, "python_binary", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BRIDGE_SCRIPT"); ok {
		applyStringValue(&cfg.BridgeScript, value, cfg.FieldSources, "bridge_script", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_PROVIDER"); ok {
		applyStringValue(&cfg.Provider, value, cfg.FieldSources, "provider", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_PROVIDER_PROFILE"); ok {
		applyStringValue(&cfg.ProviderProfile, value, cfg.FieldSources, "profile", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_MODEL"); ok {
		applyStringValue(&cfg.Model, value, cfg.FieldSources, "model", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_BASE_URL"); ok {
		applyStringValue(&cfg.BaseURL, value, cfg.FieldSources, "base_url", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_API_KEY"); ok {
		applyStringValue(&cfg.APIKey, value, cfg.FieldSources, "credential", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_PLATFORM_TEMPERATURE"); ok {
		parsed, err := parseFloatValue("RADISHMIND_PLATFORM_TEMPERATURE", value)
		if err != nil {
			return err
		}
		cfg.Temperature = parsed
		cfg.FieldSources["temperature"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH"); ok {
		parsed, err := parseBoolValue("RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH", value)
		if err != nil {
			return err
		}
		cfg.ControlPlaneReadDevAuthEnabled = parsed
		cfg.FieldSources["control_plane_read_dev_auth"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_AUTH_MODE"); ok {
		applyStringValue(&cfg.ControlPlaneReadAuthMode, value, cfg.FieldSources, "control_plane_read_auth_mode", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_ISSUER"); ok {
		cfg.ControlPlaneReadTestIssuer = strings.TrimSpace(value)
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_AUDIENCE"); ok {
		cfg.ControlPlaneReadTestAudience = strings.TrimSpace(value)
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_SIGNED_TEST_PUBLIC_KEY_PEM"); ok {
		cfg.ControlPlaneReadTestPublicKeyPEM = strings.TrimSpace(value)
	}
	if err := applyControlPlaneReadOIDCEnvOverrides(cfg); err != nil {
		return err
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_STORE"); ok {
		applyStringValue(&cfg.ControlPlaneReadStoreMode, value, cfg.FieldSources, "control_plane_read_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.ControlPlaneReadDatabaseURL, value, cfg.FieldSources, "control_plane_read_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_CONTROL_PLANE_READ_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ControlPlaneReadDatabaseTimeout, parsed, cfg.FieldSources, "control_plane_read_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP"); ok {
		parsed, err := parseBoolValue("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP", value)
		if err != nil {
			return err
		}
		cfg.WorkflowSavedDraftDevHTTPEnabled = parsed
		cfg.FieldSources["workflow_saved_draft_dev_http"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE"); ok {
		parsed, err := parseBoolValue("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE", value)
		if err != nil {
			return err
		}
		cfg.WorkflowSavedDraftDevWriteEnabled = parsed
		cfg.FieldSources["workflow_saved_draft_dev_write"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_DRAFT_DEV_HTTP"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_DRAFT_DEV_HTTP", value)
		if err != nil {
			return err
		}
		cfg.ApplicationDraftDevHTTPEnabled = parsed
		cfg.FieldSources["application_draft_dev_http"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_DRAFT_DEV_WRITE"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_DRAFT_DEV_WRITE", value)
		if err != nil {
			return err
		}
		cfg.ApplicationDraftDevWriteEnabled = parsed
		cfg.FieldSources["application_draft_dev_write"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_DRAFT_STORE"); ok {
		applyStringValue(&cfg.ApplicationDraftStoreMode, value, cfg.FieldSources, "application_draft_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.ApplicationDraftDatabaseURL, value, cfg.FieldSources, "application_draft_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_DRAFT_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_APPLICATION_DRAFT_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ApplicationDraftDatabaseTimeout, parsed, cfg.FieldSources, "application_draft_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_PUBLISH_DEV_HTTP", value)
		if err != nil {
			return err
		}
		cfg.ApplicationPublishDevHTTPEnabled = parsed
		cfg.FieldSources["application_publish_dev_http"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_PUBLISH_DEV_WRITE", value)
		if err != nil {
			return err
		}
		cfg.ApplicationPublishDevWriteEnabled = parsed
		cfg.FieldSources["application_publish_dev_write"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_PUBLISH_STORE"); ok {
		applyStringValue(&cfg.ApplicationPublishStoreMode, value, cfg.FieldSources, "application_publish_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.ApplicationPublishDatabaseURL, value, cfg.FieldSources, "application_publish_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_PUBLISH_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_APPLICATION_PUBLISH_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ApplicationPublishDatabaseTimeout, parsed, cfg.FieldSources, "application_publish_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_CATALOG_DEV_HTTP"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_CATALOG_DEV_HTTP", value)
		if err != nil {
			return err
		}
		cfg.ApplicationCatalogDevHTTPEnabled = parsed
		cfg.FieldSources["application_catalog_dev_http"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_CATALOG_DEV_WRITE"); ok {
		parsed, err := parseBoolValue("RADISHMIND_APPLICATION_CATALOG_DEV_WRITE", value)
		if err != nil {
			return err
		}
		cfg.ApplicationCatalogDevWriteEnabled = parsed
		cfg.FieldSources["application_catalog_dev_write"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_CATALOG_STORE"); ok {
		applyStringValue(&cfg.ApplicationCatalogStoreMode, value, cfg.FieldSources, "application_catalog_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.ApplicationCatalogDatabaseURL, value, cfg.FieldSources, "application_catalog_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_APPLICATION_CATALOG_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_APPLICATION_CATALOG_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.ApplicationCatalogDatabaseTimeout, parsed, cfg.FieldSources, "application_catalog_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP"); ok {
		parsed, err := parseBoolValue("RADISHMIND_API_KEY_LIFECYCLE_DEV_HTTP", value)
		if err != nil {
			return err
		}
		cfg.APIKeyLifecycleDevHTTPEnabled = parsed
		cfg.FieldSources["api_key_lifecycle_dev_http"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE"); ok {
		parsed, err := parseBoolValue("RADISHMIND_API_KEY_LIFECYCLE_DEV_WRITE", value)
		if err != nil {
			return err
		}
		cfg.APIKeyLifecycleDevWriteEnabled = parsed
		cfg.FieldSources["api_key_lifecycle_dev_write"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_API_KEY_STORE"); ok {
		applyStringValue(&cfg.APIKeyStoreMode, value, cfg.FieldSources, "api_key_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_API_KEY_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.APIKeyDatabaseURL, value, cfg.FieldSources, "api_key_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_API_KEY_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_API_KEY_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.APIKeyDatabaseTimeout, parsed, cfg.FieldSources, "api_key_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_GATEWAY_AUTH_MODE"); ok {
		applyStringValue(&cfg.GatewayAuthMode, value, cfg.FieldSources, "gateway_auth_mode", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_EXECUTOR_DEV"); ok {
		parsed, err := parseBoolValue("RADISHMIND_WORKFLOW_EXECUTOR_DEV", value)
		if err != nil {
			return err
		}
		cfg.WorkflowExecutorDevEnabled = parsed
		cfg.FieldSources["workflow_executor_dev"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV"); ok {
		parsed, err := parseBoolValue("RADISHMIND_WORKFLOW_DIAGNOSTICS_DEV", value)
		if err != nil {
			return err
		}
		cfg.WorkflowDiagnosticsDevEnabled = parsed
		cfg.FieldSources["workflow_diagnostics_dev"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV"); ok {
		parsed, err := parseBoolValue("RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV", value)
		if err != nil {
			return err
		}
		cfg.GatewayRequestHistoryDevEnabled = parsed
		cfg.FieldSources["gateway_request_history_dev"] = configSourceEnv
	}
	if value, ok := stringEnv("RADISHMIND_GATEWAY_REQUEST_STORE"); ok {
		applyStringValue(&cfg.GatewayRequestStoreMode, value, cfg.FieldSources, "gateway_request_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.GatewayRequestDatabaseURL, value, cfg.FieldSources, "gateway_request_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.GatewayRequestDatabaseTimeout, parsed, cfg.FieldSources, "gateway_request_database_timeout", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_LOCAL_PERSISTENCE_MODE"); ok {
		applyStringValue(&cfg.LocalPersistenceMode, value, cfg.FieldSources, "local_persistence_mode", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_SQLITE_DEV_DATABASE_PATH"); ok {
		applyStringValue(&cfg.SQLiteDevDatabasePath, value, cfg.FieldSources, "sqlite_dev_database_path", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE"); ok {
		applyStringValue(
			&cfg.WorkflowSavedDraftStoreMode,
			value,
			cfg.FieldSources,
			"workflow_saved_draft_store",
			configSourceEnv,
		)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(
			&cfg.WorkflowSavedDraftDatabaseURL,
			value,
			cfg.FieldSources,
			"workflow_saved_draft_database",
			configSourceEnv,
		)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(
			&cfg.WorkflowSavedDraftDatabaseTimeout,
			parsed,
			cfg.FieldSources,
			"workflow_saved_draft_database_timeout",
			configSourceEnv,
		)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_RUN_STORE"); ok {
		applyStringValue(&cfg.WorkflowRunStoreMode, value, cfg.FieldSources, "workflow_run_store", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL"); ok {
		applyStringValue(&cfg.WorkflowRunDatabaseURL, value, cfg.FieldSources, "workflow_run_database", configSourceEnv)
	}
	if value, ok := stringEnv("RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT"); ok {
		parsed, err := parseDurationValue("RADISHMIND_WORKFLOW_RUN_DATABASE_TIMEOUT", value)
		if err != nil {
			return err
		}
		applyDurationValue(&cfg.WorkflowRunDatabaseTimeout, parsed, cfg.FieldSources, "workflow_run_database_timeout", configSourceEnv)
	}
	return nil
}

func (cfg Config) SanitizedSummary() ConfigSummary {
	cfg = EffectiveLocalPersistenceConfig(cfg)
	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		provider = defaultProvider
	}
	profile := strings.TrimSpace(cfg.ProviderProfile)
	model := strings.TrimSpace(cfg.Model)
	baseURLConfigured := strings.TrimSpace(cfg.BaseURL) != ""
	bridgeMode := strings.TrimSpace(cfg.BridgeMode)
	if bridgeMode == "" {
		bridgeMode = "process_per_request"
	}
	bridgeWorkerCount := cfg.BridgeWorkerCount
	if bridgeWorkerCount <= 0 {
		bridgeWorkerCount = defaultBridgeWorkerCount
	}
	bridgeQueueCapacity := cfg.BridgeQueueCapacity
	if bridgeQueueCapacity <= 0 {
		bridgeQueueCapacity = defaultBridgeQueueSize
	}
	bridgeHandshakeTimeout := cfg.BridgeHandshakeTimeout
	if bridgeHandshakeTimeout <= 0 {
		bridgeHandshakeTimeout = defaultBridgeHandshake
	}
	workflowSavedDraftStoreMode := strings.TrimSpace(cfg.WorkflowSavedDraftStoreMode)
	if workflowSavedDraftStoreMode == "" {
		workflowSavedDraftStoreMode = defaultDraftStoreMode
	}
	applicationDraftStoreMode := strings.TrimSpace(cfg.ApplicationDraftStoreMode)
	if applicationDraftStoreMode == "" {
		applicationDraftStoreMode = defaultApplicationDraftStoreMode
	}
	applicationPublishStoreMode := strings.TrimSpace(cfg.ApplicationPublishStoreMode)
	if applicationPublishStoreMode == "" {
		applicationPublishStoreMode = defaultApplicationPublishStoreMode
	}
	applicationCatalogStoreMode := strings.TrimSpace(cfg.ApplicationCatalogStoreMode)
	if applicationCatalogStoreMode == "" {
		applicationCatalogStoreMode = defaultApplicationCatalogStoreMode
	}
	apiKeyStoreMode := strings.TrimSpace(cfg.APIKeyStoreMode)
	if apiKeyStoreMode == "" {
		apiKeyStoreMode = defaultAPIKeyStoreMode
	}
	gatewayAuthMode := EffectiveGatewayAuthMode(cfg)
	workflowRunStoreMode := strings.TrimSpace(cfg.WorkflowRunStoreMode)
	if workflowRunStoreMode == "" {
		workflowRunStoreMode = defaultRunStoreMode
	}
	gatewayRequestStoreMode := strings.TrimSpace(cfg.GatewayRequestStoreMode)
	if gatewayRequestStoreMode == "" {
		gatewayRequestStoreMode = defaultGatewayRequestStoreMode
	}
	localPersistenceMode := EffectiveLocalPersistenceMode(cfg)
	controlPlaneReadStoreMode := EffectiveControlPlaneReadStoreMode(cfg)
	credentialState := credentialState(provider, strings.TrimSpace(cfg.APIKey) != "")
	requiredFields := requiredConfigFields(provider)
	if workflowSavedDraftStoreMode == "postgres_dev_test" {
		requiredFields = append(
			requiredFields,
			"control_plane_read_dev_auth",
			"workflow_saved_draft_dev_http",
			"workflow_saved_draft_dev_write",
			"workflow_saved_draft_database",
		)
	}
	if workflowSavedDraftStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_saved_draft_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_saved_draft_dev_write")
	}
	if applicationDraftStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_database")
	}
	if applicationDraftStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_write")
	}
	if applicationPublishStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_store_postgres_dev_test")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_database")
		requiredFields = appendRequiredConfigField(requiredFields, "application_publish_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_publish_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_publish_database")
	}
	if applicationPublishStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_draft_store_sqlite_dev")
		requiredFields = appendRequiredConfigField(requiredFields, "application_publish_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_publish_dev_write")
	}
	if applicationCatalogStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_database")
	}
	if apiKeyStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_lifecycle_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_lifecycle_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_database")
	}
	if apiKeyStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_dev_write")
		requiredFields = appendRequiredConfigField(requiredFields, "application_catalog_store_sqlite_dev")
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_lifecycle_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_lifecycle_dev_write")
	}
	if gatewayAuthMode == "api_key_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "api_key_lifecycle_dev_http")
		requiredFields = appendRequiredConfigField(requiredFields, "gateway_request_history_dev")
	}
	if cfg.WorkflowExecutorDevEnabled {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_saved_draft_dev_http")
	}
	if cfg.GatewayRequestHistoryDevEnabled {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
	}
	if gatewayRequestStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "gateway_request_history_dev")
		requiredFields = appendRequiredConfigField(requiredFields, "gateway_request_database")
	}
	if gatewayRequestStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "gateway_request_history_dev")
	}
	if workflowRunStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_executor_dev")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_run_database")
	}
	if workflowRunStoreMode == "sqlite_dev" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_dev_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "workflow_executor_dev")
	}
	if controlPlaneReadStoreMode == "postgres_dev_test" {
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_verified_auth")
		requiredFields = appendRequiredConfigField(requiredFields, "control_plane_read_database")
	}
	missingRequiredFields := missingRequiredConfigFields(cfg, requiredFields)

	return ConfigSummary{
		ListenAddr:                           strings.TrimSpace(cfg.ListenAddr),
		ControlPlaneReadDevAuthEnabled:       cfg.ControlPlaneReadDevAuthEnabled,
		ControlPlaneReadAuthMode:             EffectiveControlPlaneReadAuthMode(cfg),
		ControlPlaneReadTestConfigured:       strings.TrimSpace(cfg.ControlPlaneReadTestIssuer) != "" && strings.TrimSpace(cfg.ControlPlaneReadTestAudience) != "" && strings.TrimSpace(cfg.ControlPlaneReadTestPublicKeyPEM) != "",
		ControlPlaneReadOIDCConfigured:       strings.TrimSpace(cfg.ControlPlaneReadOIDCIssuer) != "" && strings.TrimSpace(cfg.ControlPlaneReadOIDCAudience) != "" && strings.TrimSpace(cfg.ControlPlaneReadOIDCMappingVersion) != "",
		ControlPlaneReadOIDCMappingVersion:   strings.TrimSpace(cfg.ControlPlaneReadOIDCMappingVersion),
		ControlPlaneReadOIDCEvidenceRef:      strings.TrimSpace(cfg.ControlPlaneReadOIDCEvidenceRef),
		ControlPlaneReadStoreMode:            controlPlaneReadStoreMode,
		ControlPlaneReadDatabaseConfigured:   strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL) != "",
		WorkflowSavedDraftDevHTTPEnabled:     cfg.WorkflowSavedDraftDevHTTPEnabled,
		WorkflowSavedDraftDevWriteEnabled:    cfg.WorkflowSavedDraftDevWriteEnabled,
		ApplicationDraftDevHTTPEnabled:       cfg.ApplicationDraftDevHTTPEnabled,
		ApplicationDraftDevWriteEnabled:      cfg.ApplicationDraftDevWriteEnabled,
		ApplicationDraftStoreMode:            applicationDraftStoreMode,
		ApplicationDraftDatabaseConfigured:   strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) != "",
		ApplicationPublishDevHTTPEnabled:     cfg.ApplicationPublishDevHTTPEnabled,
		ApplicationPublishDevWriteEnabled:    cfg.ApplicationPublishDevWriteEnabled,
		ApplicationPublishStoreMode:          applicationPublishStoreMode,
		ApplicationPublishDatabaseConfigured: strings.TrimSpace(cfg.ApplicationPublishDatabaseURL) != "",
		ApplicationCatalogDevHTTPEnabled:     cfg.ApplicationCatalogDevHTTPEnabled,
		ApplicationCatalogDevWriteEnabled:    cfg.ApplicationCatalogDevWriteEnabled,
		ApplicationCatalogStoreMode:          applicationCatalogStoreMode,
		ApplicationCatalogDatabaseConfigured: strings.TrimSpace(cfg.ApplicationCatalogDatabaseURL) != "",
		APIKeyLifecycleDevHTTPEnabled:        cfg.APIKeyLifecycleDevHTTPEnabled,
		APIKeyLifecycleDevWriteEnabled:       cfg.APIKeyLifecycleDevWriteEnabled,
		APIKeyStoreMode:                      apiKeyStoreMode,
		APIKeyDatabaseConfigured:             strings.TrimSpace(cfg.APIKeyDatabaseURL) != "",
		GatewayAuthMode:                      gatewayAuthMode,
		WorkflowExecutorDevEnabled:           cfg.WorkflowExecutorDevEnabled,
		WorkflowDiagnosticsDevEnabled:        cfg.WorkflowDiagnosticsDevEnabled,
		GatewayRequestHistoryDevEnabled:      cfg.GatewayRequestHistoryDevEnabled,
		GatewayRequestStoreMode:              gatewayRequestStoreMode,
		GatewayRequestDatabaseConfigured:     strings.TrimSpace(cfg.GatewayRequestDatabaseURL) != "",
		LocalPersistenceMode:                 localPersistenceMode,
		LocalPersistenceConfigured:           strings.TrimSpace(cfg.LocalPersistenceMode) != "",
		SQLiteDevDatabaseConfigured:          strings.TrimSpace(cfg.SQLiteDevDatabasePath) != "",
		SQLiteDevSchemaStatus:                sqliteDevSchemaStatus(localPersistenceMode),
		LocalPersistenceComponentsConsistent: localPersistenceComponentsConsistent(cfg),
		WorkflowSavedDraftStoreMode:          workflowSavedDraftStoreMode,
		WorkflowSavedDraftDatabaseConfigured: strings.TrimSpace(cfg.WorkflowSavedDraftDatabaseURL) != "",
		WorkflowRunStoreMode:                 workflowRunStoreMode,
		WorkflowRunDatabaseConfigured:        strings.TrimSpace(cfg.WorkflowRunDatabaseURL) != "",
		Provider:                             provider,
		Profile:                              profile,
		Model:                                model,
		ModelConfigured:                      model != "",
		BaseURLConfigured:                    baseURLConfigured,
		CredentialState:                      credentialState,
		Timeouts: map[string]string{
			"read_header":                          cfg.ReadHeaderTimeout.String(),
			"write":                                cfg.WriteTimeout.String(),
			"bridge":                               cfg.BridgeTimeout.String(),
			"bridge_handshake":                     bridgeHandshakeTimeout.String(),
			"control_plane_read_database":          cfg.ControlPlaneReadDatabaseTimeout.String(),
			"control_plane_read_oidc_discovery":    cfg.ControlPlaneReadOIDCDiscoveryTimeout.String(),
			"control_plane_read_oidc_jwks_max_age": cfg.ControlPlaneReadOIDCJWKSMaxAge.String(),
			"workflow_saved_draft_database":        cfg.WorkflowSavedDraftDatabaseTimeout.String(),
			"application_draft_database":           cfg.ApplicationDraftDatabaseTimeout.String(),
			"application_publish_database":         cfg.ApplicationPublishDatabaseTimeout.String(),
			"application_catalog_database":         cfg.ApplicationCatalogDatabaseTimeout.String(),
			"api_key_database":                     cfg.APIKeyDatabaseTimeout.String(),
			"workflow_run_database":                cfg.WorkflowRunDatabaseTimeout.String(),
		},
		PythonBridge: PythonBridge{
			PythonBinary:     strings.TrimSpace(cfg.PythonBinary),
			Script:           strings.TrimSpace(cfg.BridgeScript),
			Mode:             bridgeMode,
			WorkerCount:      bridgeWorkerCount,
			QueueCapacity:    bridgeQueueCapacity,
			HandshakeTimeout: bridgeHandshakeTimeout.String(),
		},
		Temperature:           cfg.Temperature,
		RequiredFields:        requiredFields,
		MissingRequiredFields: missingRequiredFields,
		SecretFields: []string{
			"RADISHMIND_PLATFORM_API_KEY",
			"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL",
			"RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_DATABASE_URL",
			"RADISHMIND_APPLICATION_DRAFT_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_DATABASE_URL",
			"RADISHMIND_APPLICATION_PUBLISH_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_DATABASE_URL",
			"RADISHMIND_APPLICATION_CATALOG_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_API_KEY_DEV_TEST_DATABASE_URL",
			"RADISHMIND_API_KEY_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_WORKFLOW_RUN_DEV_TEST_DATABASE_URL",
			"RADISHMIND_WORKFLOW_RUN_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL",
			"RADISHMIND_GATEWAY_REQUEST_DEV_TEST_MIGRATION_DATABASE_URL",
			"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_DATABASE_URL",
			"RADISHMIND_CONTROL_PLANE_READ_DEV_TEST_MIGRATION_DATABASE_URL",
		},
		ConfigFile: ConfigFileSummary{
			Path:       strings.TrimSpace(cfg.ConfigFile),
			Configured: strings.TrimSpace(cfg.ConfigFile) != "",
			Loaded:     strings.TrimSpace(cfg.ConfigFile) != "",
			Source:     fieldSource(cfg.FieldSources, "config_file"),
		},
		FieldSources: sanitizedFieldSources(cfg.FieldSources),
		Sanitized:    true,
	}
}

func (cfg Config) Check() []string {
	return cfg.SanitizedSummary().MissingRequiredFields
}

func credentialState(provider string, hasAPIKey bool) string {
	switch strings.TrimSpace(provider) {
	case "", "mock":
		return "not_required"
	case "ollama":
		if hasAPIKey {
			return "configured"
		}
		return "optional_missing"
	default:
		if hasAPIKey {
			return "configured"
		}
		return "missing"
	}
}

func requiredConfigFields(provider string) []string {
	switch strings.TrimSpace(provider) {
	case "", "mock":
		return []string{"listen_addr", "provider"}
	case "ollama":
		return []string{"listen_addr", "provider", "model", "base_url"}
	default:
		return []string{"listen_addr", "provider", "model", "base_url", "credential"}
	}
}

func missingRequiredConfigFields(cfg Config, requiredFields []string) []string {
	missing := make([]string, 0)
	for _, field := range requiredFields {
		switch field {
		case "listen_addr":
			if strings.TrimSpace(cfg.ListenAddr) == "" {
				missing = append(missing, field)
			}
		case "provider":
			if strings.TrimSpace(cfg.Provider) == "" {
				missing = append(missing, field)
			}
		case "profile":
			if strings.TrimSpace(cfg.ProviderProfile) == "" {
				missing = append(missing, field)
			}
		case "model":
			if strings.TrimSpace(cfg.Model) == "" {
				missing = append(missing, field)
			}
		case "base_url":
			if strings.TrimSpace(cfg.BaseURL) == "" {
				missing = append(missing, field)
			}
		case "credential":
			if credentialState(cfg.Provider, strings.TrimSpace(cfg.APIKey) != "") == "missing" {
				missing = append(missing, field)
			}
		case "control_plane_read_dev_auth":
			if !cfg.ControlPlaneReadDevAuthEnabled {
				missing = append(missing, field)
			}
		case "control_plane_read_verified_auth":
			mode := EffectiveControlPlaneReadAuthMode(cfg)
			if mode != "signed_test_token" && mode != "radish_oidc_integration_test" {
				missing = append(missing, field)
			}
		case "control_plane_read_database":
			if strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "workflow_saved_draft_dev_http":
			if !cfg.WorkflowSavedDraftDevHTTPEnabled {
				missing = append(missing, field)
			}
		case "workflow_saved_draft_dev_write":
			if !cfg.WorkflowSavedDraftDevWriteEnabled {
				missing = append(missing, field)
			}
		case "workflow_saved_draft_database":
			if strings.TrimSpace(cfg.WorkflowSavedDraftDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "application_draft_dev_http":
			if !cfg.ApplicationDraftDevHTTPEnabled {
				missing = append(missing, field)
			}
		case "application_draft_dev_write":
			if !cfg.ApplicationDraftDevWriteEnabled {
				missing = append(missing, field)
			}
		case "application_draft_database":
			if strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "application_draft_store_postgres_dev_test":
			if strings.TrimSpace(cfg.ApplicationDraftStoreMode) != "postgres_dev_test" {
				missing = append(missing, field)
			}
		case "application_draft_store_sqlite_dev":
			if strings.TrimSpace(cfg.ApplicationDraftStoreMode) != "sqlite_dev" {
				missing = append(missing, field)
			}
		case "application_publish_dev_http":
			if !cfg.ApplicationPublishDevHTTPEnabled {
				missing = append(missing, field)
			}
		case "application_publish_dev_write":
			if !cfg.ApplicationPublishDevWriteEnabled {
				missing = append(missing, field)
			}
		case "application_publish_database":
			if strings.TrimSpace(cfg.ApplicationPublishDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "application_catalog_dev_http":
			if !cfg.ApplicationCatalogDevHTTPEnabled {
				missing = append(missing, field)
			}
		case "application_catalog_dev_write":
			if !cfg.ApplicationCatalogDevWriteEnabled {
				missing = append(missing, field)
			}
		case "application_catalog_database":
			if strings.TrimSpace(cfg.ApplicationCatalogDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "application_catalog_store_sqlite_dev":
			if strings.TrimSpace(cfg.ApplicationCatalogStoreMode) != "sqlite_dev" {
				missing = append(missing, field)
			}
		case "api_key_lifecycle_dev_http":
			if !cfg.APIKeyLifecycleDevHTTPEnabled {
				missing = append(missing, field)
			}
		case "api_key_lifecycle_dev_write":
			if !cfg.APIKeyLifecycleDevWriteEnabled {
				missing = append(missing, field)
			}
		case "api_key_database":
			if strings.TrimSpace(cfg.APIKeyDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "workflow_executor_dev":
			if !cfg.WorkflowExecutorDevEnabled {
				missing = append(missing, field)
			}
		case "workflow_run_database":
			if strings.TrimSpace(cfg.WorkflowRunDatabaseURL) == "" {
				missing = append(missing, field)
			}
		case "gateway_request_history_dev":
			if !cfg.GatewayRequestHistoryDevEnabled {
				missing = append(missing, field)
			}
		case "gateway_request_database":
			if strings.TrimSpace(cfg.GatewayRequestDatabaseURL) == "" {
				missing = append(missing, field)
			}
		}
	}
	return missing
}

func appendRequiredConfigField(fields []string, field string) []string {
	for _, existing := range fields {
		if existing == field {
			return fields
		}
	}
	return append(fields, field)
}

func sanitizedFieldSources(fieldSources map[string]string) map[string]string {
	summarySources := make(map[string]string, len(fieldSources)+1)
	for key, source := range fieldSources {
		summarySources[key] = strings.TrimSpace(source)
	}
	return summarySources
}

func fieldSource(fieldSources map[string]string, key string) string {
	if fieldSources == nil {
		return configSourceDefault
	}
	source := strings.TrimSpace(fieldSources[key])
	if source == "" {
		return configSourceDefault
	}
	return source
}

func stringEnv(key string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", false
	}
	return value, true
}

func applyStringValue(target *string, value string, fieldSources map[string]string, fieldName string, source string) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return
	}
	*target = normalizedValue
	fieldSources[fieldName] = source
}

func applyDurationValue(target *time.Duration, value time.Duration, fieldSources map[string]string, fieldName string, source string) {
	*target = value
	fieldSources[fieldName] = source
}

func applyIntValue(target *int, value int, fieldSources map[string]string, fieldName string, source string) {
	*target = value
	fieldSources[fieldName] = source
}

func parseDurationValue(key string, rawValue string) (time.Duration, error) {
	parsed, err := time.ParseDuration(strings.TrimSpace(rawValue))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return parsed, nil
}

func parseFloatValue(key string, rawValue string) (float64, error) {
	parsed, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid float: %w", key, err)
	}
	return parsed, nil
}

func parseIntValue(key string, rawValue string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(rawValue))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	return parsed, nil
}

func validateBridgeRuntimeConfig(cfg Config) error {
	if err := validateLocalPersistenceConfig(cfg); err != nil {
		return err
	}
	switch EffectiveControlPlaneReadAuthMode(cfg) {
	case "disabled":
	case "dev_headers":
		if !cfg.ControlPlaneReadDevAuthEnabled {
			return fmt.Errorf("control plane read dev_headers auth mode requires control plane read dev auth")
		}
	case "signed_test_token":
		if strings.TrimSpace(cfg.ControlPlaneReadTestIssuer) == "" || strings.TrimSpace(cfg.ControlPlaneReadTestAudience) == "" || strings.TrimSpace(cfg.ControlPlaneReadTestPublicKeyPEM) == "" {
			return fmt.Errorf("control plane read signed_test_token auth mode requires issuer, audience, and public key PEM")
		}
		if !isRSAPublicKeyPEM(cfg.ControlPlaneReadTestPublicKeyPEM) {
			return fmt.Errorf("control plane read signed_test_token public key PEM must contain an RSA public key")
		}
	case "radish_oidc_integration_test":
		if err := validateControlPlaneReadOIDCIntegrationConfig(cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("control plane read auth mode must be disabled, dev_headers, signed_test_token, or radish_oidc_integration_test")
	}
	switch EffectiveControlPlaneReadStoreMode(cfg) {
	case "fake_store_dev":
	case "postgres_dev_test":
		mode := EffectiveControlPlaneReadAuthMode(cfg)
		if mode != "signed_test_token" && mode != "radish_oidc_integration_test" {
			return fmt.Errorf("control plane read postgres_dev_test store requires signed_test_token or radish_oidc_integration_test auth mode")
		}
		if strings.TrimSpace(cfg.ControlPlaneReadDatabaseURL) == "" {
			return fmt.Errorf("control plane read postgres_dev_test store requires a database URL")
		}
	default:
		return fmt.Errorf("control plane read store must be fake_store_dev or postgres_dev_test")
	}
	if cfg.ControlPlaneReadDatabaseTimeout <= 0 {
		return fmt.Errorf("control plane read database timeout must be positive")
	}
	if cfg.WorkflowSavedDraftDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("saved workflow draft dev HTTP requires control plane read dev auth")
	}
	if cfg.WorkflowSavedDraftDevWriteEnabled && !cfg.WorkflowSavedDraftDevHTTPEnabled {
		return fmt.Errorf("saved workflow draft dev write requires saved workflow draft dev HTTP")
	}
	switch strings.TrimSpace(cfg.WorkflowSavedDraftStoreMode) {
	case "", "memory_dev", "repository_disabled", "repository":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			!cfg.WorkflowSavedDraftDevWriteEnabled {
			return fmt.Errorf("saved workflow draft sqlite_dev store requires complete development gates")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			!cfg.WorkflowSavedDraftDevWriteEnabled || strings.TrimSpace(cfg.WorkflowSavedDraftDatabaseURL) == "" {
			return fmt.Errorf("saved workflow draft postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("saved workflow draft store must be memory_dev, sqlite_dev, postgres_dev_test, repository_disabled, or repository")
	}
	if cfg.WorkflowSavedDraftDatabaseTimeout <= 0 {
		return fmt.Errorf("saved workflow draft database timeout must be positive")
	}
	if cfg.GatewayRequestHistoryDevEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("gateway request history dev requires control plane read dev auth")
	}
	switch strings.TrimSpace(cfg.GatewayRequestStoreMode) {
	case "", "memory_dev":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.GatewayRequestHistoryDevEnabled {
			return fmt.Errorf("Gateway request sqlite_dev store requires complete development gates")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.GatewayRequestHistoryDevEnabled ||
			strings.TrimSpace(cfg.GatewayRequestDatabaseURL) == "" {
			return fmt.Errorf("Gateway request postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("Gateway request store must be memory_dev, sqlite_dev, or postgres_dev_test")
	}
	if cfg.GatewayRequestDatabaseTimeout <= 0 {
		return fmt.Errorf("Gateway request database timeout must be positive")
	}
	if cfg.ApplicationDraftDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("application draft dev HTTP requires control plane read dev auth")
	}
	if cfg.ApplicationDraftDevWriteEnabled && !cfg.ApplicationDraftDevHTTPEnabled {
		return fmt.Errorf("application draft dev write requires application draft dev HTTP")
	}
	switch strings.TrimSpace(cfg.ApplicationDraftStoreMode) {
	case "", "memory_dev":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled {
			return fmt.Errorf("application draft sqlite_dev store requires complete development gates")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled || strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) == "" {
			return fmt.Errorf("application draft postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("application draft store must be memory_dev, sqlite_dev, or postgres_dev_test")
	}
	if cfg.ApplicationDraftDatabaseTimeout <= 0 {
		return fmt.Errorf("application draft database timeout must be positive")
	}
	if cfg.ApplicationPublishDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("application publish dev HTTP requires control plane read dev auth")
	}
	if cfg.ApplicationPublishDevHTTPEnabled && !cfg.ApplicationDraftDevHTTPEnabled {
		return fmt.Errorf("application publish dev HTTP requires application draft dev HTTP")
	}
	if cfg.ApplicationPublishDevWriteEnabled && !cfg.ApplicationPublishDevHTTPEnabled {
		return fmt.Errorf("application publish dev write requires application publish dev HTTP")
	}
	if cfg.ApplicationPublishDevWriteEnabled && !cfg.ApplicationDraftDevWriteEnabled {
		return fmt.Errorf("application publish dev write requires application draft dev write")
	}
	switch strings.TrimSpace(cfg.ApplicationPublishStoreMode) {
	case "", "memory_dev":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled ||
			strings.TrimSpace(cfg.ApplicationDraftStoreMode) != "sqlite_dev" || !cfg.ApplicationPublishDevHTTPEnabled || !cfg.ApplicationPublishDevWriteEnabled {
			return fmt.Errorf("application publish sqlite_dev store requires complete development gates and sqlite_dev application draft store")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationDraftDevHTTPEnabled || !cfg.ApplicationDraftDevWriteEnabled ||
			strings.TrimSpace(cfg.ApplicationDraftStoreMode) != "postgres_dev_test" || strings.TrimSpace(cfg.ApplicationDraftDatabaseURL) == "" ||
			!cfg.ApplicationPublishDevHTTPEnabled || !cfg.ApplicationPublishDevWriteEnabled || strings.TrimSpace(cfg.ApplicationPublishDatabaseURL) == "" {
			return fmt.Errorf("application publish postgres_dev_test store requires complete development gates and postgres_dev_test application draft store")
		}
	default:
		return fmt.Errorf("application publish store must be memory_dev, sqlite_dev, or postgres_dev_test")
	}
	if cfg.ApplicationPublishDatabaseTimeout <= 0 {
		return fmt.Errorf("application publish database timeout must be positive")
	}
	if cfg.ApplicationCatalogDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("application catalog dev HTTP requires control plane read dev auth")
	}
	if cfg.ApplicationCatalogDevWriteEnabled && !cfg.ApplicationCatalogDevHTTPEnabled {
		return fmt.Errorf("application catalog dev write requires application catalog dev HTTP")
	}
	switch strings.TrimSpace(cfg.ApplicationCatalogStoreMode) {
	case "", "memory_dev":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.ApplicationCatalogDevWriteEnabled {
			return fmt.Errorf("application catalog sqlite_dev store requires complete development gates")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.ApplicationCatalogDevWriteEnabled || strings.TrimSpace(cfg.ApplicationCatalogDatabaseURL) == "" {
			return fmt.Errorf("application catalog postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("application catalog store must be memory_dev, sqlite_dev, or postgres_dev_test")
	}
	if cfg.ApplicationCatalogDatabaseTimeout <= 0 {
		return fmt.Errorf("application catalog database timeout must be positive")
	}
	if cfg.APIKeyLifecycleDevHTTPEnabled && !cfg.ControlPlaneReadDevAuthEnabled {
		return fmt.Errorf("API key lifecycle dev HTTP requires control plane read dev auth")
	}
	if cfg.APIKeyLifecycleDevHTTPEnabled && !cfg.ApplicationCatalogDevHTTPEnabled {
		return fmt.Errorf("API key lifecycle dev HTTP requires application catalog dev HTTP")
	}
	if cfg.APIKeyLifecycleDevWriteEnabled && !cfg.APIKeyLifecycleDevHTTPEnabled {
		return fmt.Errorf("API key lifecycle dev write requires API key lifecycle dev HTTP")
	}
	if cfg.APIKeyLifecycleDevWriteEnabled && !cfg.ApplicationCatalogDevWriteEnabled {
		return fmt.Errorf("API key lifecycle dev write requires application catalog dev write")
	}
	switch strings.TrimSpace(cfg.APIKeyStoreMode) {
	case "", "memory_dev":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.ApplicationCatalogDevHTTPEnabled || !cfg.ApplicationCatalogDevWriteEnabled ||
			strings.TrimSpace(cfg.ApplicationCatalogStoreMode) != "sqlite_dev" ||
			!cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.APIKeyLifecycleDevWriteEnabled {
			return fmt.Errorf("API key sqlite_dev store requires complete development gates and sqlite_dev application catalog store")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.APIKeyLifecycleDevWriteEnabled || strings.TrimSpace(cfg.APIKeyDatabaseURL) == "" {
			return fmt.Errorf("API key postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("API key store must be memory_dev, sqlite_dev, or postgres_dev_test")
	}
	if cfg.APIKeyDatabaseTimeout <= 0 {
		return fmt.Errorf("API key database timeout must be positive")
	}
	switch EffectiveGatewayAuthMode(cfg) {
	case "dev_headers":
	case "api_key_dev_test":
		if !cfg.APIKeyLifecycleDevHTTPEnabled || !cfg.GatewayRequestHistoryDevEnabled {
			return fmt.Errorf("Gateway api_key_dev_test auth requires API key lifecycle HTTP and Gateway request history")
		}
	default:
		return fmt.Errorf("Gateway auth mode must be dev_headers or api_key_dev_test")
	}
	if cfg.WorkflowDiagnosticsDevEnabled {
		if !cfg.WorkflowExecutorDevEnabled {
			return fmt.Errorf("workflow diagnostics dev requires workflow executor dev")
		}
		if !strings.EqualFold(strings.TrimSpace(cfg.Provider), "mock") {
			return fmt.Errorf("workflow diagnostics dev requires mock provider")
		}
	}
	switch strings.TrimSpace(cfg.WorkflowRunStoreMode) {
	case "", "memory_dev", "repository_disabled", "repository":
	case "sqlite_dev":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			!cfg.WorkflowExecutorDevEnabled {
			return fmt.Errorf("workflow run sqlite_dev store requires control plane read dev auth, saved workflow draft dev HTTP, and workflow executor dev")
		}
	case "postgres_dev_test":
		if !cfg.ControlPlaneReadDevAuthEnabled || !cfg.WorkflowSavedDraftDevHTTPEnabled ||
			!cfg.WorkflowExecutorDevEnabled || strings.TrimSpace(cfg.WorkflowRunDatabaseURL) == "" {
			return fmt.Errorf("workflow run postgres_dev_test store requires complete development gates and a database URL")
		}
	default:
		return fmt.Errorf("workflow run store must be memory_dev, sqlite_dev, postgres_dev_test, repository_disabled, or repository")
	}
	if cfg.WorkflowRunDatabaseTimeout <= 0 {
		return fmt.Errorf("workflow run database timeout must be positive")
	}
	switch strings.TrimSpace(cfg.BridgeMode) {
	case "process_per_request", "stdio_pool":
	default:
		return fmt.Errorf("bridge_mode must be process_per_request or stdio_pool")
	}
	if cfg.BridgeWorkerCount < 1 || cfg.BridgeWorkerCount > 32 {
		return fmt.Errorf("bridge_worker_count must be between 1 and 32")
	}
	if cfg.BridgeQueueCapacity < 1 || cfg.BridgeQueueCapacity > 1024 {
		return fmt.Errorf("bridge_queue_capacity must be between 1 and 1024")
	}
	if cfg.BridgeHandshakeTimeout <= 0 {
		return fmt.Errorf("bridge_handshake_timeout must be positive")
	}
	return nil
}

func EffectiveControlPlaneReadAuthMode(cfg Config) string {
	mode := strings.TrimSpace(cfg.ControlPlaneReadAuthMode)
	if mode == "" && cfg.ControlPlaneReadDevAuthEnabled {
		return "dev_headers"
	}
	if mode == "" {
		return "disabled"
	}
	return mode
}

func EffectiveGatewayAuthMode(cfg Config) string {
	mode := strings.TrimSpace(cfg.GatewayAuthMode)
	if mode == "" {
		return defaultGatewayAuthMode
	}
	return mode
}

func EffectiveControlPlaneReadStoreMode(cfg Config) string {
	mode := strings.TrimSpace(cfg.ControlPlaneReadStoreMode)
	if mode == "" {
		return defaultControlPlaneReadStoreMode
	}
	return mode
}

func isRSAPublicKeyPEM(rawPEM string) bool {
	block, _ := pem.Decode([]byte(strings.TrimSpace(rawPEM)))
	if block == nil {
		return false
	}
	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		_, ok := parsed.(*rsa.PublicKey)
		return ok
	}
	_, err := x509.ParsePKCS1PublicKey(block.Bytes)
	return err == nil
}

func parseBoolValue(key string, rawValue string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(rawValue)) {
	case "1", "true", "yes", "on", "enabled":
		return true, nil
	case "0", "false", "no", "off", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be a boolean flag", key)
	}
}

func parseJSONFloatValue(key string, rawValue json.RawMessage) (float64, error) {
	var numericValue float64
	if err := json.Unmarshal(rawValue, &numericValue); err == nil {
		return numericValue, nil
	}
	var stringValue string
	if err := json.Unmarshal(rawValue, &stringValue); err == nil {
		return parseFloatValue(key, strings.TrimSpace(stringValue))
	}
	return 0, fmt.Errorf("%s must be a number or numeric string", key)
}
