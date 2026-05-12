package diagnostics

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type fakeBridgeClient struct {
	providers    []bridge.ProviderDescription
	inventory    bridge.ProviderInventory
	providersErr error
	inventoryErr error
}

func (client fakeBridgeClient) DescribeProviders(context.Context) ([]bridge.ProviderDescription, error) {
	return client.providers, client.providersErr
}

func (client fakeBridgeClient) DescribeInventory(context.Context) (bridge.ProviderInventory, error) {
	return client.inventory, client.inventoryErr
}

func TestBuildReportReturnsSanitizedOKDiagnostics(t *testing.T) {
	cfg := config.Config{
		ListenAddr:    "127.0.0.1:8080",
		BridgeTimeout: time.Second,
		PythonBinary:  "python3",
		BridgeScript:  "scripts/run-platform-bridge.py",
		Provider:      "mock",
	}
	client := fakeBridgeClient{
		providers: []bridge.ProviderDescription{
			{ProviderID: "mock"},
			{ProviderID: "huggingface"},
		},
		inventory: bridge.ProviderInventory{
			Profiles: []bridge.ProviderProfileDescription{
				{CredentialState: "configured", DeploymentMode: "remote_api"},
				{CredentialState: "optional_missing", DeploymentMode: "local_daemon"},
			},
			ActiveProfileChain: []string{"provider:huggingface:profile:hf-chat"},
		},
	}

	report := BuildReport(context.Background(), cfg, client, time.Date(2026, 5, 12, 1, 2, 3, 0, time.UTC))

	if report.Status != "ok" {
		t.Fatalf("unexpected status: %s", report.Status)
	}
	if !report.Sanitized || !report.Config.Sanitized {
		t.Fatalf("expected sanitized report")
	}
	if len(report.FailureCodes) != 0 || report.Failure != nil {
		t.Fatalf("unexpected failure: %#v", report)
	}
	if !report.Bridge.ProviderRegistryOK || !report.Bridge.InventoryOK {
		t.Fatalf("expected bridge checks to pass: %#v", report.Bridge)
	}
	if !reflect.DeepEqual(report.Providers.RegistryProviderIDs, []string{"mock", "huggingface"}) {
		t.Fatalf("unexpected provider ids: %#v", report.Providers.RegistryProviderIDs)
	}
	if report.Providers.ConfiguredCredentialCount != 1 || report.Providers.OptionalCredentialCount != 1 {
		t.Fatalf("unexpected credential counts: %#v", report.Providers)
	}
}

func TestBuildReportSurfacesFailureBoundary(t *testing.T) {
	cfg := config.Config{
		ListenAddr:    "127.0.0.1:8080",
		BridgeTimeout: time.Second,
		PythonBinary:  "python3",
		BridgeScript:  "missing.py",
		Provider:      "huggingface",
	}
	client := fakeBridgeClient{
		providersErr: errors.New("boom"),
		inventoryErr: errors.New("boom"),
	}

	report := BuildReport(context.Background(), cfg, client, time.Date(2026, 5, 12, 1, 2, 3, 0, time.UTC))

	if report.Status != "error" {
		t.Fatalf("unexpected status: %s", report.Status)
	}
	if report.Failure == nil {
		t.Fatalf("expected failure boundary")
	}
	if report.Failure.Code != "CONFIG_REQUIRED_FIELDS_MISSING" {
		t.Fatalf("unexpected failure code: %#v", report.Failure)
	}
	if !reflect.DeepEqual(report.Bridge.FailureCodes, []string{"PROVIDER_REGISTRY_UNAVAILABLE", "PROVIDER_INVENTORY_UNAVAILABLE"}) {
		t.Fatalf("unexpected bridge failure codes: %#v", report.Bridge.FailureCodes)
	}
}
