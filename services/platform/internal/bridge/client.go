package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPythonBinary = "python3"
	defaultScriptPath   = "scripts/run-platform-bridge.py"
)

type EnvelopeOptions struct {
	Provider        string
	ProviderProfile string
	Model           string
	BaseURL         string
	APIKey          string
	Temperature     float64
	RequestTimeout  time.Duration
}

type ProviderDescription struct {
	ProviderID         string         `json:"provider_id"`
	DisplayName        string         `json:"display_name"`
	DefaultAPIStyle    string         `json:"default_api_style"`
	SupportedAPIStyles []string       `json:"supported_api_styles"`
	ProfileDriven      bool           `json:"profile_driven"`
	Notes              string         `json:"notes"`
	Capabilities       map[string]any `json:"capabilities"`
}

type ProviderProfileDescription struct {
	Profile               string  `json:"profile"`
	NormalizedProfile     string  `json:"normalized_profile"`
	ProviderID            string  `json:"provider_id"`
	ResolvedModel         string  `json:"resolved_model"`
	APIStyle              string  `json:"api_style"`
	HasBaseURL            bool    `json:"has_base_url"`
	HasAPIKey             bool    `json:"has_api_key"`
	RequestTimeoutSeconds float64 `json:"request_timeout_seconds"`
	Active                bool    `json:"active"`
	Fallback              bool    `json:"fallback"`
	ChainIndex            int     `json:"chain_index"`
}

type ProviderInventory struct {
	Providers          []ProviderDescription        `json:"providers"`
	Profiles           []ProviderProfileDescription `json:"profiles"`
	ActiveProfileChain []string                     `json:"active_profile_chain"`
}

type GatewayEnvelope struct {
	SchemaVersion int            `json:"schema_version"`
	Status        string         `json:"status"`
	RequestID     string         `json:"request_id"`
	Project       string         `json:"project"`
	Task          string         `json:"task"`
	Response      map[string]any `json:"response"`
	Error         *GatewayError  `json:"error"`
	Metadata      map[string]any `json:"metadata"`
}

type GatewayError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	pythonBinary string
	scriptPath   string
}

func NewClient(pythonBinary string, scriptPath string) *Client {
	normalizedPythonBinary := strings.TrimSpace(pythonBinary)
	if normalizedPythonBinary == "" {
		normalizedPythonBinary = defaultPythonBinary
	}
	normalizedScriptPath := strings.TrimSpace(scriptPath)
	if normalizedScriptPath == "" {
		normalizedScriptPath = defaultScriptPath
	}
	return &Client{
		pythonBinary: normalizedPythonBinary,
		scriptPath:   normalizedScriptPath,
	}
}

func (c *Client) DescribeProviders(ctx context.Context) ([]ProviderDescription, error) {
	stdout, err := c.run(ctx, []string{"providers"}, nil)
	if err != nil {
		return nil, err
	}
	var descriptions []ProviderDescription
	if err := json.Unmarshal(stdout, &descriptions); err != nil {
		return nil, fmt.Errorf("decode provider registry: %w", err)
	}
	return descriptions, nil
}

func (c *Client) DescribeInventory(ctx context.Context) (ProviderInventory, error) {
	stdout, err := c.run(ctx, []string{"inventory"}, nil)
	if err != nil {
		return ProviderInventory{}, err
	}
	var inventory ProviderInventory
	if err := json.Unmarshal(stdout, &inventory); err != nil {
		return ProviderInventory{}, fmt.Errorf("decode provider inventory: %w", err)
	}
	return inventory, nil
}

func (c *Client) HandleEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options EnvelopeOptions,
) (GatewayEnvelope, error) {
	args := []string{
		"envelope",
		"--provider", strings.TrimSpace(options.Provider),
		"--provider-profile", strings.TrimSpace(options.ProviderProfile),
		"--model", strings.TrimSpace(options.Model),
		"--base-url", strings.TrimSpace(options.BaseURL),
		"--api-key", strings.TrimSpace(options.APIKey),
		"--temperature", strconv.FormatFloat(options.Temperature, 'f', -1, 64),
	}
	if options.RequestTimeout > 0 {
		args = append(
			args,
			"--request-timeout-seconds",
			strconv.FormatFloat(options.RequestTimeout.Seconds(), 'f', -1, 64),
		)
	}
	stdout, err := c.run(ctx, args, canonicalRequest)
	if err != nil {
		return GatewayEnvelope{}, err
	}
	var envelope GatewayEnvelope
	if err := json.Unmarshal(stdout, &envelope); err != nil {
		return GatewayEnvelope{}, fmt.Errorf("decode gateway envelope: %w", err)
	}
	return envelope, nil
}

func (c *Client) run(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	scriptPath := c.scriptPath
	if !filepath.IsAbs(scriptPath) {
		resolvedPath, err := filepath.Abs(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("resolve bridge script path: %w", err)
		}
		scriptPath = resolvedPath
	}

	commandArgs := append([]string{scriptPath}, args...)
	command := exec.CommandContext(ctx, c.pythonBinary, commandArgs...)
	if len(stdin) > 0 {
		command.Stdin = bytes.NewReader(stdin)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return nil, fmt.Errorf("python bridge failed: %w: %s", err, stderrText)
		}
		return nil, fmt.Errorf("python bridge failed: %w", err)
	}
	return stdout.Bytes(), nil
}
