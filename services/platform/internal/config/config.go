package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultListenAddr        = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 30 * time.Second
)

type Config struct {
	ListenAddr        string
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
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

	listenAddr := strings.TrimSpace(os.Getenv("RADISHMIND_PLATFORM_LISTEN_ADDR"))
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	return Config{
		ListenAddr:        listenAddr,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
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
