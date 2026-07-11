package bridge

import (
	"fmt"
	"strings"
	"time"
)

type Mode string

const (
	ModeProcessPerRequest Mode = "process_per_request"
	ModeStdioPool         Mode = "stdio_pool"
)

const (
	DefaultWorkerCount      = 4
	DefaultQueueCapacity    = 64
	DefaultHandshakeTimeout = 5 * time.Second
	defaultShutdownTimeout  = 2 * time.Second
	maximumWorkerCount      = 32
	maximumQueueCapacity    = 1024
)

type ClientOptions struct {
	Mode             Mode
	WorkerCount      int
	QueueCapacity    int
	HandshakeTimeout time.Duration
}

func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.TrimSpace(value))
	if mode == "" {
		return ModeProcessPerRequest, nil
	}
	switch mode {
	case ModeProcessPerRequest, ModeStdioPool:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported bridge mode")
	}
}

func normalizeClientOptions(options ClientOptions) (ClientOptions, error) {
	mode, err := ParseMode(string(options.Mode))
	if err != nil {
		return ClientOptions{}, err
	}
	options.Mode = mode
	if options.WorkerCount == 0 {
		options.WorkerCount = DefaultWorkerCount
	}
	if options.QueueCapacity == 0 {
		options.QueueCapacity = DefaultQueueCapacity
	}
	if options.HandshakeTimeout == 0 {
		options.HandshakeTimeout = DefaultHandshakeTimeout
	}
	if options.WorkerCount < 1 || options.WorkerCount > maximumWorkerCount {
		return ClientOptions{}, fmt.Errorf("bridge worker count must be between 1 and %d", maximumWorkerCount)
	}
	if options.QueueCapacity < 0 || options.QueueCapacity > maximumQueueCapacity {
		return ClientOptions{}, fmt.Errorf("bridge queue capacity must be between 0 and %d", maximumQueueCapacity)
	}
	if options.HandshakeTimeout <= 0 {
		return ClientOptions{}, fmt.Errorf("bridge handshake timeout must be positive")
	}
	return options, nil
}
