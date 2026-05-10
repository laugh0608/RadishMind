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
