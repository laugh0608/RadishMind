package config

import (
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
	Sanitized             bool              `json:"sanitized"`
}

type PythonBridge struct {
	PythonBinary string `json:"python_binary"`
	Script       string `json:"script"`
}

func LoadFromEnv() (Config, error) {
	readHeaderTimeout, err := loadDurationEnv("RADISHMIND_PLATFORM_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := loadDurationEnv("RADISHMIND_PLATFORM_WRITE_TIMEOUT", defaultWriteTimeout)
	if err != nil {
		return Config{}, err
	}
	bridgeTimeout, err := loadDurationEnv("RADISHMIND_PLATFORM_BRIDGE_TIMEOUT", defaultBridgeTimeout)
	if err != nil {
		return Config{}, err
	}
	temperature, err := loadFloatEnv("RADISHMIND_PLATFORM_TEMPERATURE", 0)
	if err != nil {
		return Config{}, err
	}

	listenAddr := strings.TrimSpace(os.Getenv("RADISHMIND_PLATFORM_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	return Config{
		ListenAddr:        listenAddr,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		BridgeTimeout:     bridgeTimeout,
		PythonBinary:      loadStringEnv("RADISHMIND_PLATFORM_PYTHON_BIN", defaultPythonBinary),
		BridgeScript:      loadStringEnv("RADISHMIND_PLATFORM_BRIDGE_SCRIPT", defaultBridgeScript),
		Provider:          loadStringEnv("RADISHMIND_PLATFORM_PROVIDER", defaultProvider),
		ProviderProfile:   loadStringEnv("RADISHMIND_PLATFORM_PROVIDER_PROFILE", ""),
		Model:             loadStringEnv("RADISHMIND_PLATFORM_MODEL", ""),
		BaseURL:           loadStringEnv("RADISHMIND_PLATFORM_BASE_URL", ""),
		APIKey:            loadStringEnv("RADISHMIND_PLATFORM_API_KEY", ""),
		Temperature:       temperature,
	}, nil
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
		Sanitized:             true,
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

func loadDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(rawValue)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}
	return parsed, nil
}

func loadFloatEnv(key string, fallback float64) (float64, error) {
	rawValue := strings.TrimSpace(os.Getenv(key))
	if rawValue == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid float: %w", key, err)
	}
	return parsed, nil
}

func loadStringEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
