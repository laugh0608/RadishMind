package diagnostics

import (
	"context"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

type BridgeClient interface {
	DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error)
	DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error)
}

type Report struct {
	Status       string                 `json:"status"`
	Sanitized    bool                   `json:"sanitized"`
	Generated    string                 `json:"generated_at"`
	Config       config.ConfigSummary   `json:"config"`
	Checks       []CheckResult          `json:"checks"`
	FailureCodes []string               `json:"failure_codes"`
	Bridge       BridgeDiagnostics      `json:"bridge"`
	Providers    ProviderDiagnostics    `json:"providers"`
	Failure      *FailureBoundary       `json:"failure,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type BridgeDiagnostics struct {
	PythonBinary       string   `json:"python_binary"`
	Script             string   `json:"script"`
	ProviderRegistryOK bool     `json:"provider_registry_ok"`
	InventoryOK        bool     `json:"inventory_ok"`
	FailureCodes       []string `json:"failure_codes"`
}

type ProviderDiagnostics struct {
	RegistryCount             int      `json:"registry_count"`
	RegistryProviderIDs       []string `json:"registry_provider_ids"`
	ProfileCount              int      `json:"profile_count"`
	SelectableModelIDs        []string `json:"selectable_model_ids"`
	SelectableModelCount      int      `json:"selectable_model_count"`
	MissingCredentialModelIDs []string `json:"missing_credential_model_ids"`
	ActiveProfileChain        []string `json:"active_profile_chain"`
	MissingCredentialCount    int      `json:"missing_credential_count"`
	OptionalCredentialCount   int      `json:"optional_credential_count"`
	ConfiguredCredentialCount int      `json:"configured_credential_count"`
	RemoteProfileCount        int      `json:"remote_profile_count"`
	LocalProfileCount         int      `json:"local_profile_count"`
}

type FailureBoundary struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func BuildReport(ctx context.Context, cfg config.Config, client BridgeClient, generatedAt time.Time) Report {
	summary := cfg.SanitizedSummary()
	report := Report{
		Status:       "ok",
		Sanitized:    true,
		Generated:    generatedAt.UTC().Format(time.RFC3339),
		Config:       summary,
		Checks:       make([]CheckResult, 0, 4),
		FailureCodes: make([]string, 0),
		Bridge: BridgeDiagnostics{
			PythonBinary: summary.PythonBridge.PythonBinary,
			Script:       summary.PythonBridge.Script,
			FailureCodes: make([]string, 0),
		},
		Providers: ProviderDiagnostics{
			RegistryProviderIDs:       make([]string, 0),
			SelectableModelIDs:        make([]string, 0),
			MissingCredentialModelIDs: make([]string, 0),
			ActiveProfileChain:        make([]string, 0),
		},
		Metadata: map[string]interface{}{
			"diagnostics_version": 1,
		},
	}

	if len(summary.MissingRequiredFields) > 0 {
		report.addError("config_required_fields", "CONFIG_REQUIRED_FIELDS_MISSING", "required platform config fields are missing")
	} else {
		report.addOK("config_required_fields")
	}

	providers, err := client.DescribeProviders(ctx)
	if err != nil {
		report.addError("bridge_provider_registry", "PROVIDER_REGISTRY_UNAVAILABLE", "python bridge provider registry is unavailable")
	} else {
		report.addOK("bridge_provider_registry")
		report.Bridge.ProviderRegistryOK = true
		report.Providers.RegistryCount = len(providers)
		report.Providers.RegistryProviderIDs = providerIDs(providers)
	}

	inventory, err := client.DescribeInventory(ctx)
	if err != nil {
		report.addError("bridge_provider_inventory", "PROVIDER_INVENTORY_UNAVAILABLE", "python bridge provider inventory is unavailable")
	} else {
		report.addOK("bridge_provider_inventory")
		report.Bridge.InventoryOK = true
		report.Providers.ProfileCount = len(inventory.Profiles)
		report.Providers.ActiveProfileChain = append([]string(nil), inventory.ActiveProfileChain...)
		applyProfileCounts(&report.Providers, inventory.Profiles)
	}

	if len(report.Bridge.FailureCodes) > 0 {
		report.setFailure()
	}
	if report.Status != "ok" {
		report.setFailure()
		return report
	}
	report.Checks = append(report.Checks, CheckResult{
		Name:   "deployment_readiness",
		Status: "ok",
	})
	return report
}

func (report *Report) addOK(name string) {
	report.Checks = append(report.Checks, CheckResult{Name: name, Status: "ok"})
}

func (report *Report) addError(name string, code string, message string) {
	report.Status = "error"
	report.Checks = append(report.Checks, CheckResult{
		Name:    name,
		Status:  "error",
		Code:    code,
		Message: message,
	})
	report.FailureCodes = append(report.FailureCodes, code)
	if strings.HasPrefix(name, "bridge_") {
		report.Bridge.FailureCodes = append(report.Bridge.FailureCodes, code)
	}
}

func (report *Report) setFailure() {
	if report.Failure != nil || len(report.FailureCodes) == 0 {
		return
	}
	report.Failure = &FailureBoundary{
		Code:    report.FailureCodes[0],
		Message: "platform service is not ready for deployment; inspect checks for the failed boundary",
	}
}

func providerIDs(providers []bridge.ProviderDescription) []string {
	ids := make([]string, 0, len(providers))
	for _, provider := range providers {
		providerID := strings.TrimSpace(provider.ProviderID)
		if providerID == "" {
			continue
		}
		ids = append(ids, providerID)
	}
	return ids
}

func applyProfileCounts(diagnostics *ProviderDiagnostics, profiles []bridge.ProviderProfileDescription) {
	for _, profile := range profiles {
		modelID := buildProviderProfileModelID(profile.ProviderID, profile.Profile)
		if modelID != "" {
			diagnostics.SelectableModelIDs = append(diagnostics.SelectableModelIDs, modelID)
			diagnostics.SelectableModelCount++
		}
		switch profile.CredentialState {
		case "missing":
			diagnostics.MissingCredentialCount++
			if modelID != "" {
				diagnostics.MissingCredentialModelIDs = append(diagnostics.MissingCredentialModelIDs, modelID)
			}
		case "optional_missing":
			diagnostics.OptionalCredentialCount++
		case "configured":
			diagnostics.ConfiguredCredentialCount++
		}
		switch profile.DeploymentMode {
		case "remote_api":
			diagnostics.RemoteProfileCount++
		case "local_daemon", "embedded":
			diagnostics.LocalProfileCount++
		}
	}
}

func buildProviderProfileModelID(providerID string, profile string) string {
	normalizedProviderID := strings.TrimSpace(providerID)
	normalizedProfile := strings.TrimSpace(profile)
	if normalizedProviderID == "" && normalizedProfile == "" {
		return ""
	}
	if normalizedProviderID == "openai-compatible" {
		if normalizedProfile == "" {
			return "profile:default"
		}
		return "profile:" + normalizedProfile
	}
	if normalizedProviderID == "" {
		return "profile:" + normalizedProfile
	}
	if normalizedProfile == "" {
		normalizedProfile = "default"
	}
	return "provider:" + normalizedProviderID + ":profile:" + normalizedProfile
}
