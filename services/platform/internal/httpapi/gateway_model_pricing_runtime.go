package httpapi

import (
	"context"
	"log"
	"strings"

	"radishmind.local/services/platform/internal/bridge"
)

type gatewayProviderAttemptMarkerKey struct{}

type gatewayProviderAttemptBridgeClient struct {
	inner bridgeClient
}

func newGatewayProviderAttemptBridgeClient(inner bridgeClient) *gatewayProviderAttemptBridgeClient {
	return &gatewayProviderAttemptBridgeClient{inner: inner}
}

func (client *gatewayProviderAttemptBridgeClient) DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error) {
	return client.inner.DescribeProviders(ctx)
}

func (client *gatewayProviderAttemptBridgeClient) DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error) {
	return client.inner.DescribeInventory(ctx)
}

func (client *gatewayProviderAttemptBridgeClient) HandleEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options bridge.EnvelopeOptions,
) (bridge.GatewayEnvelope, error) {
	markGatewayProviderAttempt(ctx)
	return client.inner.HandleEnvelope(ctx, canonicalRequest, options)
}

func (client *gatewayProviderAttemptBridgeClient) StreamEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options bridge.EnvelopeOptions,
	handleEvent func(bridge.StreamEvent) error,
) error {
	markGatewayProviderAttempt(ctx)
	return client.inner.StreamEnvelope(ctx, canonicalRequest, options, handleEvent)
}

func (client *gatewayProviderAttemptBridgeClient) Close() {
	if closer, ok := client.inner.(interface{ Close() }); ok {
		closer.Close()
	}
}

func withGatewayProviderAttemptMarker(ctx context.Context, trace *requestTrace) context.Context {
	if ctx == nil || trace == nil {
		return ctx
	}
	return context.WithValue(ctx, gatewayProviderAttemptMarkerKey{}, &trace.providerAttempted)
}

func markGatewayProviderAttempt(ctx context.Context) {
	if ctx == nil {
		return
	}
	marker, ok := ctx.Value(gatewayProviderAttemptMarkerKey{}).(*bool)
	if ok && marker != nil {
		*marker = true
	}
}

func (server *Server) bindGatewayModelPricingSnapshot(trace *requestTrace) {
	if trace == nil || !server.config.GatewayModelPricingCaptureDevEnabled {
		return
	}
	snapshot := server.gatewayModelPricingSnapshotForSelection(
		trace.gatewayRequestContext, trace.requestID, trace.selection,
	)
	trace.gatewayPricingSnapshot = snapshot
	log.Printf(
		"radishmind_gateway_pricing_snapshot request_id=%s availability=%s",
		strings.TrimSpace(trace.requestID), strings.TrimSpace(snapshot.Availability),
	)
}

func (server *Server) gatewayModelPricingSnapshotForSelection(
	requestContext GatewayRequestContext,
	requestID string,
	selection northboundSelection,
) GatewayModelPricingSnapshot {
	snapshot := gatewayModelPricingUnavailableSnapshot(GatewayPricingSnapshotUnavailable, "pricing_selection_unavailable")
	if !server.config.GatewayModelPricingCaptureDevEnabled ||
		strings.TrimSpace(selection.provider) == "" || strings.TrimSpace(selection.providerProfile) == "" ||
		strings.TrimSpace(selection.model) == "" {
		return snapshot
	}
	pricingContext := GatewayModelPricingContext{
		RequestContext: requestContext.RequestContext,
		TenantRef:      requestContext.TenantRef,
		WorkspaceID:    requestContext.WorkspaceID,
		Environment:    strings.TrimSpace(server.config.GatewayModelPricingEnvironment),
		ProviderID:     strings.TrimSpace(selection.provider),
		ProfileID:      strings.TrimSpace(selection.providerProfile),
		ModelID:        strings.TrimSpace(selection.model),
		ActorRef:       requestContext.SubjectRef,
		RequestID:      strings.TrimSpace(requestID),
		AuditRef:       "audit_" + strings.TrimSpace(requestID) + "_gateway-pricing-snapshot",
	}
	result := newGatewayModelPricingService(server.gatewayModelPricingRepository).ReadCurrent(pricingContext)
	switch {
	case result.Policy != nil && result.FailureCode == "":
		return gatewayModelPricingSnapshotFromPolicy(*result.Policy)
	case result.FailureCode == GatewayModelPricingFailurePolicyNotFound:
		return gatewayModelPricingUnavailableSnapshot(GatewayPricingSnapshotNotConfigured, "pricing_policy_not_configured")
	default:
		return gatewayModelPricingUnavailableSnapshot(GatewayPricingSnapshotUnavailable, "pricing_policy_unavailable")
	}
}

var _ bridgeClient = (*gatewayProviderAttemptBridgeClient)(nil)
