package httpapi

import "strings"

type AdminProviderRouteExecutionMode string

const (
	AdminProviderRouteExecutionSingleAttempt      AdminProviderRouteExecutionMode = "single_attempt"
	AdminProviderRouteExecutionSequentialFallback AdminProviderRouteExecutionMode = "sequential_fallback"
)

type AdminProviderRouteAttemptTarget struct {
	Ordinal           int    `json:"ordinal"`
	ProviderProfileID string `json:"provider_profile_id"`
}

func normalizeAdminModelRouteContract(route AdminModelRouteDefinition) (AdminModelRouteDefinition, string) {
	route.RouteID = strings.TrimSpace(route.RouteID)
	route.Protocol = strings.TrimSpace(route.Protocol)
	route.ModelID = strings.TrimSpace(route.ModelID)
	route.ProviderProfileID = strings.TrimSpace(route.ProviderProfileID)
	if !adminProviderRouteIdentifierPattern.MatchString(route.RouteID) ||
		!isApplicationDraftProtocol(route.Protocol) ||
		!adminProviderRouteModelPattern.MatchString(route.ModelID) {
		return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
	}

	hasV1Target := route.ProviderProfileID != ""
	hasV2Contract := route.ExecutionMode != "" || len(route.AttemptTargets) > 0
	if hasV1Target == hasV2Contract {
		return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
	}
	if hasV1Target {
		if !adminProviderRouteIdentifierPattern.MatchString(route.ProviderProfileID) {
			return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
		}
		route.AttemptTargets = nil
		return route, ""
	}

	wantTargets := 0
	switch route.ExecutionMode {
	case AdminProviderRouteExecutionSingleAttempt:
		wantTargets = 1
	case AdminProviderRouteExecutionSequentialFallback:
		wantTargets = 2
	default:
		return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
	}
	if len(route.AttemptTargets) != wantTargets {
		return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
	}
	seenProfiles := make(map[string]bool, len(route.AttemptTargets))
	for index := range route.AttemptTargets {
		target := &route.AttemptTargets[index]
		target.ProviderProfileID = strings.TrimSpace(target.ProviderProfileID)
		if target.Ordinal != index+1 ||
			!adminProviderRouteIdentifierPattern.MatchString(target.ProviderProfileID) ||
			seenProfiles[target.ProviderProfileID] {
			return AdminModelRouteDefinition{}, AdminProviderRouteFailurePayloadInvalid
		}
		seenProfiles[target.ProviderProfileID] = true
	}
	return route, ""
}

func adminProviderRouteProfileIDs(route AdminModelRouteDefinition) []string {
	if route.ProviderProfileID != "" {
		return []string{route.ProviderProfileID}
	}
	profiles := make([]string, 0, len(route.AttemptTargets))
	for _, target := range route.AttemptTargets {
		profiles = append(profiles, target.ProviderProfileID)
	}
	return profiles
}

func adminProviderRoutePrimaryProfileID(route AdminModelRouteDefinition) string {
	profiles := adminProviderRouteProfileIDs(route)
	if len(profiles) == 0 {
		return ""
	}
	return profiles[0]
}

func adminProviderRouteContractVersion(routes []AdminModelRouteDefinition) int {
	version := 0
	for _, route := range routes {
		current := 1
		if route.ExecutionMode != "" || len(route.AttemptTargets) > 0 {
			current = 2
		}
		if version != 0 && version != current {
			return 0
		}
		version = current
	}
	return version
}

func adminProviderRouteDraftSchemaVersionForRoutes(routes []AdminModelRouteDefinition) string {
	if adminProviderRouteContractVersion(routes) == 2 {
		return adminProviderRouteDraftSchemaVersionV2
	}
	return adminProviderRouteDraftSchemaVersion
}

func adminProviderRouteCandidateSchemaVersionForRoutes(routes []AdminModelRouteDefinition) string {
	if adminProviderRouteContractVersion(routes) == 2 {
		return adminProviderRouteCandidateSchemaVersionV2
	}
	return adminProviderRouteCandidateSchemaVersion
}

func adminProviderRouteSnapshotSchemaVersionForRoutes(routes []AdminModelRouteDefinition) string {
	if adminProviderRouteContractVersion(routes) == 2 {
		return adminProviderRouteSnapshotSchemaVersionV2
	}
	return adminProviderRouteSnapshotSchemaVersion
}
