package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListenAddr        = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 30 * time.Second
	defaultBridgeTimeout     = 30 * time.Second
	defaultPythonBinary      = "python3"
	defaultBridgeScript      = "scripts/run-platform-bridge.py"
	defaultProvider          = "mock"
)

const (
	configSourceDefault = "default"
	configSourceFile    = "file"
	configSourceEnv     = "env"
)

type Config struct {
	ListenAddr        string
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	BridgeTimeout     time.Duration
	PythonBinary      string
	BridgeScript      string
	Provider          string
	ProviderProfile   string
	Model             string
	BaseURL           string
	APIKey            string
	Temperature       float64
	ConfigFile        string
	FieldSources      map[string]string
}

type ConfigSummary struct {
	ListenAddr            string            `json:"listen_addr"`
	Provider              string            `json:"provider"`
	Profile               string            `json:"profile"`
	Model                 string            `json:"model"`
	ModelConfigured       bool              `json:"model_configured"`
	BaseURLConfigured     bool              `json:"base_url_configured"`
	CredentialState       string            `json:"credential_state"`
	Timeouts              map[string]string `json:"timeouts"`
	PythonBridge          PythonBridge      `json:"python_bridge"`
	Temperature           float64           `json:"temperature"`
	RequiredFields        []string          `json:"required_fields"`
	MissingRequiredFields []string          `json:"missing_required_fields"`
	SecretFields          []string          `json:"secret_fields"`
	ConfigFile            ConfigFileSummary `json:"config_file"`
	FieldSources          map[string]string `json:"field_sources"`
	Sanitized             bool              `json:"sanitized"`
}

type PythonBridge struct {
	PythonBinary string `json:"python_binary"`
	Script       string `json:"script"`
}

type ConfigFileSummary struct {
	Path       string `json:"path"`
	Configured bool   `json:"configured"`
	Loaded     bool   `json:"loaded"`
	Source     string `json:"source"`
}

type fileConfig struct {
	ListenAddr        *string          `json:"listen_addr"`
	ReadHeaderTimeout *string          `json:"read_header_timeout"`
	WriteTimeout      *string          `json:"write_timeout"`
	BridgeTimeout     *string          `json:"bridge_timeout"`
	PythonBinary      *string          `json:"python_binary"`
	BridgeScript      *string          `json:"bridge_script"`
	Provider          *string          `json:"provider"`
	ProviderProfile   *string          `json:"provider_profile"`
	Model             *string          `json:"model"`
	BaseURL           *string          `json:"base_url"`
	APIKey            *string          `json:"api_key"`
	Temperature       *json.RawMessage `json:"temperature"`
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

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		ListenAddr:        defaultListenAddr,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		WriteTimeout:      defaultWriteTimeout,
		BridgeTimeout:     defaultBridgeTimeout,
		PythonBinary:      defaultPythonBinary,
		BridgeScript:      defaultBridgeScript,
		Provider:          defaultProvider,
		ProviderProfile:   "",
		Model:             "",
		BaseURL:           "",
		APIKey:            "",
		Temperature:       0,
		FieldSources: map[string]string{
			"listen_addr":         configSourceDefault,
			"read_header_timeout": configSourceDefault,
			"write_timeout":       configSourceDefault,
			"bridge_timeout":      configSourceDefault,
			"python_binary":       configSourceDefault,
			"bridge_script":       configSourceDefault,
			"provider":            configSourceDefault,
			"profile":             configSourceDefault,
			"model":               configSourceDefault,
			"base_url":            configSourceDefault,
			"credential":          configSourceDefault,
			"temperature":         configSourceDefault,
			"config_file":         configSourceDefault,
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
	return nil
}

func (cfg Config) SanitizedSummary() ConfigSummary {
	provider := strings.TrimSpace(cfg.Provider)
	if provider == "" {
		provider = defaultProvider
	}
	profile := strings.TrimSpace(cfg.ProviderProfile)
	model := strings.TrimSpace(cfg.Model)
	baseURLConfigured := strings.TrimSpace(cfg.BaseURL) != ""
	credentialState := credentialState(provider, strings.TrimSpace(cfg.APIKey) != "")
	requiredFields := requiredConfigFields(provider)
	missingRequiredFields := missingRequiredConfigFields(cfg, requiredFields)

	return ConfigSummary{
		ListenAddr:        strings.TrimSpace(cfg.ListenAddr),
		Provider:          provider,
		Profile:           profile,
		Model:             model,
		ModelConfigured:   model != "",
		BaseURLConfigured: baseURLConfigured,
		CredentialState:   credentialState,
		Timeouts: map[string]string{
			"read_header": cfg.ReadHeaderTimeout.String(),
			"write":       cfg.WriteTimeout.String(),
			"bridge":      cfg.BridgeTimeout.String(),
		},
		PythonBridge: PythonBridge{
			PythonBinary: strings.TrimSpace(cfg.PythonBinary),
			Script:       strings.TrimSpace(cfg.BridgeScript),
		},
		Temperature:           cfg.Temperature,
		RequiredFields:        requiredFields,
		MissingRequiredFields: missingRequiredFields,
		SecretFields:          []string{"RADISHMIND_PLATFORM_API_KEY"},
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
		}
	}
	return missing
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
