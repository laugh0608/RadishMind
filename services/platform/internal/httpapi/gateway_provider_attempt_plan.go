package httpapi

import (
	"fmt"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

const gatewayProviderAttemptPlanSchemaVersion = "gateway_provider_attempt_plan.v1"

type GatewayProviderFallbackMode string

const (
	GatewayProviderFallbackDisabled        GatewayProviderFallbackMode = "disabled"
	GatewayProviderFallbackAllowConfigured GatewayProviderFallbackMode = "allow_configured"
)

type GatewayProviderAttemptPlanTargetInput struct {
	Ordinal           int
	ProviderProfileID string
	ProviderID        string
	RuntimeProfile    string
	SelectedModel     string
	UpstreamModel     string
	InventoryDigest   string
	PricingSnapshot   GatewayModelPricingSnapshot
}

type GatewayProviderAttemptPlanTarget struct {
	AttemptID         string                                `json:"attempt_id"`
	Ordinal           int                                   `json:"ordinal"`
	ProviderProfileID string                                `json:"provider_profile_id"`
	ProviderID        string                                `json:"provider_id"`
	RuntimeProfile    string                                `json:"runtime_profile"`
	SelectedModel     string                                `json:"selected_model"`
	UpstreamModel     string                                `json:"upstream_model"`
	InventoryDigest   string                                `json:"inventory_digest"`
	PricingSnapshot   GatewayProviderAttemptPricingSnapshot `json:"pricing_snapshot"`
}

type GatewayProviderAttemptPricingSnapshot struct {
	Availability                  string `json:"availability"`
	Reason                        string `json:"reason,omitempty"`
	Currency                      string `json:"currency,omitempty"`
	TokenUnit                     int64  `json:"token_unit,omitempty"`
	InputPriceMicrosPerTokenUnit  int64  `json:"input_price_micros_per_token_unit,omitempty"`
	OutputPriceMicrosPerTokenUnit int64  `json:"output_price_micros_per_token_unit,omitempty"`
	PricingPolicyID               string `json:"pricing_policy_id,omitempty"`
	PricingPolicyVersion          int64  `json:"pricing_policy_version,omitempty"`
	PricingPolicyDigest           string `json:"pricing_policy_digest,omitempty"`
	IntegrityDigest               string `json:"integrity_digest,omitempty"`
}

type GatewayProviderAttemptPlan struct {
	SchemaVersion       string                             `json:"schema_version"`
	RootRequestID       string                             `json:"root_request_id"`
	Route               string                             `json:"route"`
	Protocol            string                             `json:"protocol"`
	RequestedModel      string                             `json:"requested_model"`
	ConfigurationID     string                             `json:"configuration_id"`
	RouteGeneration     int                                `json:"route_generation"`
	RouteSnapshotDigest string                             `json:"route_snapshot_digest"`
	ExecutionMode       AdminProviderRouteExecutionMode    `json:"execution_mode"`
	FallbackMode        GatewayProviderFallbackMode        `json:"fallback_mode"`
	FallbackAllowed     bool                               `json:"fallback_allowed"`
	MaxAttempts         int                                `json:"max_attempts"`
	Targets             []GatewayProviderAttemptPlanTarget `json:"targets"`
}

func buildGatewayProviderAttemptPlan(
	rootRequestID string,
	route string,
	protocol string,
	requestedModel string,
	snapshot AdminProviderRouteSnapshot,
	modelRoute AdminModelRouteDefinition,
	fallbackMode GatewayProviderFallbackMode,
	targetInputs []GatewayProviderAttemptPlanTargetInput,
) (GatewayProviderAttemptPlan, error) {
	rootRequestID = strings.TrimSpace(rootRequestID)
	route = strings.TrimSpace(route)
	protocol = strings.TrimSpace(protocol)
	requestedModel = strings.TrimSpace(requestedModel)
	if !validGatewayRequestReference(rootRequestID, 150) ||
		!validGatewayRequestReference(route, 128) ||
		!validGatewayRequestProtocol(protocol) ||
		!adminProviderRouteModelPattern.MatchString(requestedModel) ||
		fallbackMode != GatewayProviderFallbackDisabled && fallbackMode != GatewayProviderFallbackAllowConfigured ||
		!adminProviderRouteIdentifierPattern.MatchString(snapshot.ConfigurationID) ||
		snapshot.Generation < 1 || !adminProviderRouteDigestPattern.MatchString(snapshot.SnapshotDigest) {
		return GatewayProviderAttemptPlan{}, fmt.Errorf("invalid gateway provider attempt plan input")
	}
	normalizedRoute, failureCode := normalizeAdminModelRouteContract(modelRoute)
	if failureCode != "" || normalizedRoute.Protocol != gatewayProviderRouteProtocolName(protocol) ||
		normalizedRoute.ModelID != requestedModel {
		return GatewayProviderAttemptPlan{}, fmt.Errorf("gateway provider attempt route mismatch")
	}
	profileIDs := adminProviderRouteProfileIDs(normalizedRoute)
	if len(profileIDs) != len(targetInputs) || len(profileIDs) < 1 || len(profileIDs) > 2 {
		return GatewayProviderAttemptPlan{}, fmt.Errorf("gateway provider attempt target mismatch")
	}
	executionMode := normalizedRoute.ExecutionMode
	if executionMode == "" {
		executionMode = AdminProviderRouteExecutionSingleAttempt
	}
	maxAttempts := 1
	fallbackAllowed := executionMode == AdminProviderRouteExecutionSequentialFallback &&
		fallbackMode == GatewayProviderFallbackAllowConfigured
	if fallbackAllowed {
		maxAttempts = 2
	}
	targets := make([]GatewayProviderAttemptPlanTarget, 0, len(targetInputs))
	for index, input := range targetInputs {
		profileID := strings.TrimSpace(input.ProviderProfileID)
		providerID := strings.TrimSpace(input.ProviderID)
		runtimeProfile := strings.TrimSpace(input.RuntimeProfile)
		selectedModel := strings.TrimSpace(input.SelectedModel)
		upstreamModel := strings.TrimSpace(input.UpstreamModel)
		inventoryDigest := strings.TrimSpace(input.InventoryDigest)
		if input.Ordinal != index+1 || input.Ordinal > 2 || profileID != profileIDs[index] ||
			!adminProviderRouteIdentifierPattern.MatchString(profileID) ||
			!adminProviderRouteIdentifierPattern.MatchString(providerID) ||
			!validGatewayRequestReference(runtimeProfile, 256) ||
			selectedModel != requestedModel || !adminProviderRouteModelPattern.MatchString(upstreamModel) ||
			!adminProviderRouteDigestPattern.MatchString(inventoryDigest) ||
			!validGatewayProviderAttemptPricingSnapshot(input.PricingSnapshot) {
			return GatewayProviderAttemptPlan{}, fmt.Errorf("invalid gateway provider attempt target")
		}
		targets = append(targets, GatewayProviderAttemptPlanTarget{
			AttemptID: rootRequestID + ".pa" + fmt.Sprint(input.Ordinal), Ordinal: input.Ordinal,
			ProviderProfileID: profileID, ProviderID: providerID, RuntimeProfile: runtimeProfile,
			SelectedModel: selectedModel, UpstreamModel: upstreamModel,
			InventoryDigest: inventoryDigest, PricingSnapshot: freezeGatewayProviderAttemptPricingSnapshot(input.PricingSnapshot),
		})
	}
	plan := GatewayProviderAttemptPlan{
		SchemaVersion: gatewayProviderAttemptPlanSchemaVersion,
		RootRequestID: rootRequestID, Route: route, Protocol: protocol, RequestedModel: requestedModel,
		ConfigurationID: snapshot.ConfigurationID, RouteGeneration: snapshot.Generation,
		RouteSnapshotDigest: snapshot.SnapshotDigest, ExecutionMode: executionMode,
		FallbackMode: fallbackMode, FallbackAllowed: fallbackAllowed, MaxAttempts: maxAttempts, Targets: targets,
	}
	if !validGatewayProviderAttemptPlan(plan) {
		return GatewayProviderAttemptPlan{}, fmt.Errorf("invalid gateway provider attempt plan")
	}
	return plan, nil
}

func gatewayProviderRouteProtocolName(protocol string) string {
	value, _ := gatewayProviderRouteProtocol(protocol)
	return value
}

func cloneGatewayProviderAttemptPlan(plan GatewayProviderAttemptPlan) GatewayProviderAttemptPlan {
	plan.Targets = append([]GatewayProviderAttemptPlanTarget(nil), plan.Targets...)
	return plan
}

func cloneGatewayModelPricingSnapshot(snapshot GatewayModelPricingSnapshot) GatewayModelPricingSnapshot {
	return snapshot
}

func validGatewayProviderAttemptPricingSnapshot(snapshot GatewayModelPricingSnapshot) bool {
	if snapshot.Availability == GatewayPricingSnapshotConfigured {
		return validGatewayModelPricingSnapshot(snapshot)
	}
	if snapshot.Availability != GatewayPricingSnapshotNotConfigured && snapshot.Availability != GatewayPricingSnapshotUnavailable {
		return false
	}
	return strings.TrimSpace(snapshot.Reason) != "" && len(snapshot.Reason) <= 160 &&
		snapshot.Currency == "" && snapshot.TokenUnit == 0 &&
		snapshot.InputPriceMicrosPerTokenUnit == 0 && snapshot.OutputPriceMicrosPerTokenUnit == 0 &&
		snapshot.PricingPolicyID == "" && snapshot.PricingPolicyVersion == 0 &&
		snapshot.PricingPolicyDigest == "" && snapshot.integrityDigest == ""
}

func freezeGatewayProviderAttemptPricingSnapshot(snapshot GatewayModelPricingSnapshot) GatewayProviderAttemptPricingSnapshot {
	return GatewayProviderAttemptPricingSnapshot{
		Availability: snapshot.Availability, Reason: snapshot.Reason, Currency: snapshot.Currency,
		TokenUnit: snapshot.TokenUnit, InputPriceMicrosPerTokenUnit: snapshot.InputPriceMicrosPerTokenUnit,
		OutputPriceMicrosPerTokenUnit: snapshot.OutputPriceMicrosPerTokenUnit,
		PricingPolicyID:               snapshot.PricingPolicyID, PricingPolicyVersion: snapshot.PricingPolicyVersion,
		PricingPolicyDigest: snapshot.PricingPolicyDigest, IntegrityDigest: snapshot.integrityDigest,
	}
}

func gatewayModelPricingSnapshotFromAttempt(snapshot GatewayProviderAttemptPricingSnapshot) GatewayModelPricingSnapshot {
	return GatewayModelPricingSnapshot{
		Availability: snapshot.Availability, Reason: snapshot.Reason, Currency: snapshot.Currency,
		TokenUnit: snapshot.TokenUnit, InputPriceMicrosPerTokenUnit: snapshot.InputPriceMicrosPerTokenUnit,
		OutputPriceMicrosPerTokenUnit: snapshot.OutputPriceMicrosPerTokenUnit,
		PricingPolicyID:               snapshot.PricingPolicyID, PricingPolicyVersion: snapshot.PricingPolicyVersion,
		PricingPolicyDigest: snapshot.PricingPolicyDigest, integrityDigest: snapshot.IntegrityDigest,
	}
}

func validGatewayProviderAttemptFrozenPricingSnapshot(snapshot GatewayProviderAttemptPricingSnapshot) bool {
	return validGatewayProviderAttemptPricingSnapshot(gatewayModelPricingSnapshotFromAttempt(snapshot))
}

func validGatewayProviderAttemptPlan(plan GatewayProviderAttemptPlan) bool {
	if plan.SchemaVersion != gatewayProviderAttemptPlanSchemaVersion ||
		!validGatewayRequestReference(plan.RootRequestID, 150) ||
		!validGatewayRequestReference(plan.Route, 128) || !validGatewayRequestProtocol(plan.Protocol) ||
		!adminProviderRouteModelPattern.MatchString(plan.RequestedModel) ||
		!adminProviderRouteIdentifierPattern.MatchString(plan.ConfigurationID) || plan.RouteGeneration < 1 ||
		!adminProviderRouteDigestPattern.MatchString(plan.RouteSnapshotDigest) ||
		(plan.FallbackMode != GatewayProviderFallbackDisabled && plan.FallbackMode != GatewayProviderFallbackAllowConfigured) ||
		len(plan.Targets) < 1 || len(plan.Targets) > 2 {
		return false
	}
	switch plan.ExecutionMode {
	case AdminProviderRouteExecutionSingleAttempt:
		if len(plan.Targets) != 1 {
			return false
		}
	case AdminProviderRouteExecutionSequentialFallback:
		if len(plan.Targets) != 2 {
			return false
		}
	default:
		return false
	}
	expectedFallback := plan.ExecutionMode == AdminProviderRouteExecutionSequentialFallback &&
		plan.FallbackMode == GatewayProviderFallbackAllowConfigured
	expectedMaxAttempts := 1
	if expectedFallback {
		expectedMaxAttempts = 2
	}
	if plan.FallbackAllowed != expectedFallback || plan.MaxAttempts != expectedMaxAttempts {
		return false
	}
	profileIDs := make(map[string]struct{}, len(plan.Targets))
	for index, target := range plan.Targets {
		if target.AttemptID != plan.RootRequestID+".pa"+fmt.Sprint(index+1) || target.Ordinal != index+1 ||
			!adminProviderRouteIdentifierPattern.MatchString(target.ProviderProfileID) ||
			!adminProviderRouteIdentifierPattern.MatchString(target.ProviderID) ||
			!validGatewayRequestReference(target.RuntimeProfile, 256) || target.SelectedModel != plan.RequestedModel ||
			!adminProviderRouteModelPattern.MatchString(target.UpstreamModel) ||
			!adminProviderRouteDigestPattern.MatchString(target.InventoryDigest) ||
			!validGatewayProviderAttemptFrozenPricingSnapshot(target.PricingSnapshot) {
			return false
		}
		if _, found := profileIDs[target.ProviderProfileID]; found {
			return false
		}
		profileIDs[target.ProviderProfileID] = struct{}{}
	}
	return true
}

func gatewayProviderAttemptFailureEligible(failure bridge.ProviderAttemptFailure) bool {
	return bridge.ProviderAttemptFailureEligible(failure)
}
