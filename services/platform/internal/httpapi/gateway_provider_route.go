package httpapi

import (
	"context"
	"errors"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const (
	gatewayProviderRouteSelectionSource = "admin_provider_route_snapshot"

	gatewayProviderRouteFailureSnapshotUnavailable = "GATEWAY_PROVIDER_ROUTE_SNAPSHOT_UNAVAILABLE"
	gatewayProviderRouteFailureNotFound            = "GATEWAY_PROVIDER_ROUTE_NOT_FOUND"
	gatewayProviderRouteFailureInventoryMismatch   = "GATEWAY_PROVIDER_ROUTE_INVENTORY_MISMATCH"
	gatewayProviderRouteFailureOverrideForbidden   = "GATEWAY_PROVIDER_ROUTE_OVERRIDE_FORBIDDEN"
)

var errGatewayProviderRouteSnapshotUnavailable = errors.New(gatewayProviderRouteFailureSnapshotUnavailable)

type gatewayProviderRouteScope struct {
	RequestContext  context.Context
	RequestID       string
	TenantRef       string
	WorkspaceID     string
	Environment     string
	ConfigurationID string
	ActorRef        string
}

type gatewayProviderRouteSnapshotProvider interface {
	ReadActiveSnapshot(gatewayProviderRouteScope) (AdminProviderRouteSnapshot, bool, error)
}

type adminProviderRouteSnapshotProvider struct {
	repository adminProviderRouteRepository
}

func (provider adminProviderRouteSnapshotProvider) ReadActiveSnapshot(
	scope gatewayProviderRouteScope,
) (AdminProviderRouteSnapshot, bool, error) {
	if provider.repository == nil {
		return AdminProviderRouteSnapshot{}, false, errGatewayProviderRouteSnapshotUnavailable
	}
	result := newAdminProviderRouteService(provider.repository, nil).ReadActiveSnapshot(AdminProviderRouteContext{
		RequestContext: scope.RequestContext,
		RequestID:      scope.RequestID,
		TenantRef:      scope.TenantRef,
		WorkspaceID:    scope.WorkspaceID,
		Environment:    scope.Environment,
		ActorRef:       scope.ActorRef,
		AuditRef:       "audit_" + scope.RequestID + "_gateway-provider-route",
		ReadEnabled:    true,
	}, scope.ConfigurationID)
	if result.FailureCode != "" {
		return AdminProviderRouteSnapshot{}, false, errGatewayProviderRouteSnapshotUnavailable
	}
	if result.Snapshot == nil {
		return AdminProviderRouteSnapshot{}, false, nil
	}
	return *result.Snapshot, true, nil
}

func (s *Server) resolveGatewayNorthboundSelection(
	ctx context.Context,
	requestContext GatewayRequestContext,
	protocol string,
	requestedModel string,
	extension *chatCompletionExtension,
) (northboundSelection, string) {
	if config.EffectiveGatewayProviderRouteSource(s.config) != "admin_snapshot_dev_test" {
		return s.resolveNorthboundSelection(ctx, requestedModel, extension), ""
	}
	if extension != nil &&
		(strings.TrimSpace(extension.Provider) != "" || strings.TrimSpace(extension.ProviderProfile) != "") {
		return northboundSelection{}, gatewayProviderRouteFailureOverrideForbidden
	}
	if !validGatewayRequestContext(requestContext) {
		return northboundSelection{}, gatewayProviderRouteFailureSnapshotUnavailable
	}
	if s.providerRouteSnapshotProvider == nil {
		return northboundSelection{}, gatewayProviderRouteFailureSnapshotUnavailable
	}
	snapshot, found, err := s.providerRouteSnapshotProvider.ReadActiveSnapshot(gatewayProviderRouteScope{
		RequestContext:  ctx,
		RequestID:       requestContext.RequestID,
		TenantRef:       requestContext.TenantRef,
		WorkspaceID:     requestContext.WorkspaceID,
		Environment:     strings.TrimSpace(s.config.GatewayProviderRouteEnvironment),
		ConfigurationID: strings.TrimSpace(s.config.GatewayProviderRouteConfigurationID),
		ActorRef:        requestContext.SubjectRef,
	})
	if err != nil || !found {
		return northboundSelection{}, gatewayProviderRouteFailureSnapshotUnavailable
	}
	selection := northboundSelection{
		model:                strings.TrimSpace(requestedModel),
		source:               gatewayProviderRouteSelectionSource,
		inventoryKind:        "activated_provider_route_snapshot",
		routeConfigurationID: snapshot.ConfigurationID,
		routeGeneration:      snapshot.Generation,
		routeSnapshotDigest:  snapshot.SnapshotDigest,
	}
	routeProtocol, ok := gatewayProviderRouteProtocol(protocol)
	if !ok || selection.model == "" {
		return selection, gatewayProviderRouteFailureNotFound
	}
	var matchedRoute *AdminModelRouteDefinition
	for index := range snapshot.Configuration.ModelRoutes {
		route := snapshot.Configuration.ModelRoutes[index]
		if route.Protocol != routeProtocol || route.ModelID != selection.model {
			continue
		}
		if matchedRoute != nil {
			return selection, gatewayProviderRouteFailureSnapshotUnavailable
		}
		copied := route
		matchedRoute = &copied
	}
	if matchedRoute == nil {
		return selection, gatewayProviderRouteFailureNotFound
	}
	primaryProfileID := adminProviderRoutePrimaryProfileID(*matchedRoute)
	if primaryProfileID == "" {
		return selection, gatewayProviderRouteFailureSnapshotUnavailable
	}
	var assignment *AdminProviderProfileAssignment
	for index := range snapshot.Configuration.ProviderProfiles {
		profile := snapshot.Configuration.ProviderProfiles[index]
		if profile.ProfileID == primaryProfileID {
			copied := profile
			assignment = &copied
			break
		}
	}
	if assignment == nil {
		return selection, gatewayProviderRouteFailureSnapshotUnavailable
	}
	profileKey, ok := adminProviderRouteRuntimeProfileKey(snapshot.Environment, assignment.RuntimeProfileRef)
	if !ok {
		return selection, gatewayProviderRouteFailureSnapshotUnavailable
	}
	selection.provider = assignment.ProviderID
	selection.providerProfile = profileKey

	expectedBinding, ok := gatewayProviderRouteBinding(snapshot.InventoryBindings, assignment.ProfileID)
	if !ok || !expectedBinding.Enabled || !gatewayProviderRouteHasCapability(expectedBinding.Capabilities, routeProtocol) {
		return selection, gatewayProviderRouteFailureInventoryMismatch
	}
	profile, currentBinding, err := s.resolveGatewayProviderProfile(ctx, snapshot.Environment, *assignment)
	if err != nil || !gatewayProviderRouteBindingsEqual(expectedBinding, currentBinding) {
		return selection, gatewayProviderRouteFailureInventoryMismatch
	}
	selection.applyProfile(profile)
	selection.model = strings.TrimSpace(requestedModel)
	selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
	selection.source = gatewayProviderRouteSelectionSource
	selection.inventoryKind = "activated_provider_route_snapshot"
	selection.routeConfigurationID = snapshot.ConfigurationID
	selection.routeGeneration = snapshot.Generation
	selection.routeSnapshotDigest = snapshot.SnapshotDigest
	return selection, ""
}

func (s *Server) resolveGatewayProviderProfile(
	ctx context.Context,
	environment string,
	assignment AdminProviderProfileAssignment,
) (bridge.ProviderProfileDescription, AdminProviderInventoryBinding, error) {
	if s.bridge == nil {
		return bridge.ProviderProfileDescription{}, AdminProviderInventoryBinding{}, errGatewayProviderRouteSnapshotUnavailable
	}
	profileKey, ok := adminProviderRouteRuntimeProfileKey(environment, assignment.RuntimeProfileRef)
	if !ok {
		return bridge.ProviderProfileDescription{}, AdminProviderInventoryBinding{}, errGatewayProviderRouteSnapshotUnavailable
	}
	inventory, err := s.bridge.DescribeInventory(ctx)
	if err != nil {
		return bridge.ProviderProfileDescription{}, AdminProviderInventoryBinding{}, err
	}
	var matched *bridge.ProviderProfileDescription
	for index := range inventory.Profiles {
		profile := inventory.Profiles[index]
		if strings.TrimSpace(profile.ProviderID) != assignment.ProviderID ||
			strings.TrimSpace(profile.Profile) != profileKey {
			continue
		}
		if matched != nil {
			return bridge.ProviderProfileDescription{}, AdminProviderInventoryBinding{}, errGatewayProviderRouteSnapshotUnavailable
		}
		copied := profile
		matched = &copied
	}
	if matched == nil {
		return bridge.ProviderProfileDescription{}, AdminProviderInventoryBinding{}, errGatewayProviderRouteSnapshotUnavailable
	}
	binding, err := adminProviderRouteInventoryBindingFromProfile(
		environment,
		assignment.ProviderID,
		assignment.RuntimeProfileRef,
		*matched,
	)
	binding.ProfileID = assignment.ProfileID
	return *matched, binding, err
}

func gatewayProviderRouteProtocol(protocol string) (string, bool) {
	switch strings.TrimSpace(protocol) {
	case northboundProtocolChatCompletions:
		return "chat_completions", true
	case northboundProtocolResponses:
		return "responses", true
	case northboundProtocolMessages:
		return "messages", true
	default:
		return "", false
	}
}

func gatewayProviderRouteBinding(
	bindings []AdminProviderInventoryBinding,
	profileID string,
) (AdminProviderInventoryBinding, bool) {
	var matched *AdminProviderInventoryBinding
	for index := range bindings {
		if bindings[index].ProfileID != profileID {
			continue
		}
		if matched != nil {
			return AdminProviderInventoryBinding{}, false
		}
		copied := bindings[index]
		matched = &copied
	}
	if matched == nil {
		return AdminProviderInventoryBinding{}, false
	}
	return *matched, true
}

func gatewayProviderRouteBindingsEqual(
	expected AdminProviderInventoryBinding,
	current AdminProviderInventoryBinding,
) bool {
	return expected.ProfileID == current.ProfileID &&
		expected.ProviderID == current.ProviderID &&
		expected.RuntimeProfileRef == current.RuntimeProfileRef &&
		expected.Environment == current.Environment &&
		expected.InventoryDigest == current.InventoryDigest &&
		expected.Enabled == current.Enabled &&
		strings.Join(expected.Capabilities, "\x00") == strings.Join(current.Capabilities, "\x00")
}

func gatewayProviderRouteHasCapability(capabilities []string, expected string) bool {
	for _, capability := range capabilities {
		if capability == expected {
			return true
		}
	}
	return false
}

var _ gatewayProviderRouteSnapshotProvider = adminProviderRouteSnapshotProvider{}
