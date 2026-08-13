package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	gatewayProviderFallbackFailureDisabled          = "GATEWAY_PROVIDER_FALLBACK_DEV_DISABLED"
	gatewayProviderFallbackFailureModeInvalid       = "GATEWAY_PROVIDER_FALLBACK_MODE_INVALID"
	gatewayProviderFallbackFailureStreamUnsupported = "GATEWAY_PROVIDER_FALLBACK_STREAM_UNSUPPORTED"
	gatewayProviderAttemptFailurePlanUnavailable    = "GATEWAY_PROVIDER_ATTEMPT_PLAN_UNAVAILABLE"
	gatewayProviderAttemptFailureHistoryUnavailable = "GATEWAY_PROVIDER_ATTEMPT_HISTORY_UNAVAILABLE"
)

type gatewayProviderAttemptExecution struct {
	plan       GatewayProviderAttemptPlan
	selections []northboundSelection
}

type gatewayProviderAttemptExecutionResult struct {
	Envelope      bridge.GatewayEnvelope
	Selection     northboundSelection
	FailureCode   string
	FailureDetail string
}

type gatewayProviderAttemptPlanError struct {
	failureCode string
}

func (failure gatewayProviderAttemptPlanError) Error() string {
	return failure.failureCode
}

func gatewayProviderFallbackMode(
	extension *chatCompletionExtension,
) (GatewayProviderFallbackMode, string) {
	if extension == nil || strings.TrimSpace(string(extension.FallbackMode)) == "" {
		return GatewayProviderFallbackDisabled, ""
	}
	mode := GatewayProviderFallbackMode(strings.TrimSpace(string(extension.FallbackMode)))
	if mode != GatewayProviderFallbackDisabled && mode != GatewayProviderFallbackAllowConfigured {
		return "", gatewayProviderFallbackFailureModeInvalid
	}
	return mode, ""
}

func (server *Server) prepareGatewayProviderAttemptExecution(
	ctx context.Context,
	trace *requestTrace,
	protocol string,
	requestedModel string,
	extension *chatCompletionExtension,
	stream bool,
) (gatewayProviderAttemptExecution, northboundSelection, string) {
	mode, failureCode := gatewayProviderFallbackMode(extension)
	if failureCode != "" {
		return gatewayProviderAttemptExecution{}, northboundSelection{}, failureCode
	}
	if stream && mode == GatewayProviderFallbackAllowConfigured {
		return gatewayProviderAttemptExecution{}, northboundSelection{}, gatewayProviderFallbackFailureStreamUnsupported
	}
	if stream {
		selection, selectionFailure := server.resolveGatewayNorthboundSelection(
			ctx, trace.gatewayRequestContext, protocol, requestedModel, extension,
		)
		return gatewayProviderAttemptExecution{}, selection, selectionFailure
	}
	if !server.config.GatewayProviderFallbackDevEnabled {
		if mode == GatewayProviderFallbackAllowConfigured {
			return gatewayProviderAttemptExecution{}, northboundSelection{}, gatewayProviderFallbackFailureDisabled
		}
		selection, selectionFailure := server.resolveGatewayNorthboundSelection(
			ctx, trace.gatewayRequestContext, protocol, requestedModel, extension,
		)
		return gatewayProviderAttemptExecution{}, selection, selectionFailure
	}
	if stream || trace == nil || trace.gatewayRequestContext.Source != gatewayAPIKeyAuthenticationSource ||
		server.gatewayRequestHistoryStore == nil || server.gatewayRequestQuotaRepository == nil ||
		server.gatewayModelPricingRepository == nil || server.providerRouteSnapshotProvider == nil {
		return gatewayProviderAttemptExecution{}, northboundSelection{}, gatewayProviderAttemptFailurePlanUnavailable
	}

	execution, err := server.compileGatewayProviderAttemptExecution(
		ctx, *trace, protocol, requestedModel, mode, extension,
	)
	if err != nil || len(execution.selections) == 0 {
		var planFailure gatewayProviderAttemptPlanError
		if errors.As(err, &planFailure) {
			return gatewayProviderAttemptExecution{}, northboundSelection{}, planFailure.failureCode
		}
		return gatewayProviderAttemptExecution{}, northboundSelection{}, gatewayProviderAttemptFailurePlanUnavailable
	}
	return execution, execution.selections[0], ""
}

func (server *Server) compileGatewayProviderAttemptExecution(
	ctx context.Context,
	trace requestTrace,
	protocol string,
	requestedModel string,
	mode GatewayProviderFallbackMode,
	extension *chatCompletionExtension,
) (gatewayProviderAttemptExecution, error) {
	if extension != nil &&
		(strings.TrimSpace(extension.Provider) != "" || strings.TrimSpace(extension.ProviderProfile) != "") {
		return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureOverrideForbidden}
	}
	requestContext := trace.gatewayRequestContext
	if !validGatewayRequestContext(requestContext) {
		return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureSnapshotUnavailable}
	}
	snapshot, found, err := server.providerRouteSnapshotProvider.ReadActiveSnapshot(gatewayProviderRouteScope{
		RequestContext: ctx, RequestID: requestContext.RequestID,
		TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		Environment:     strings.TrimSpace(server.config.GatewayProviderRouteEnvironment),
		ConfigurationID: strings.TrimSpace(server.config.GatewayProviderRouteConfigurationID),
		ActorRef:        requestContext.SubjectRef,
	})
	if err != nil || !found ||
		snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersion &&
			snapshot.SchemaVersion != adminProviderRouteSnapshotSchemaVersionV2 {
		return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureSnapshotUnavailable}
	}
	routeProtocol, ok := gatewayProviderRouteProtocol(protocol)
	if !ok || strings.TrimSpace(requestedModel) == "" {
		return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureNotFound}
	}
	modelRoute, found := gatewayProviderAttemptModelRoute(snapshot, routeProtocol, requestedModel)
	if !found {
		return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureNotFound}
	}
	profileIDs := adminProviderRouteProfileIDs(modelRoute)
	selections := make([]northboundSelection, 0, len(profileIDs))
	targets := make([]GatewayProviderAttemptPlanTargetInput, 0, len(profileIDs))
	for index, profileID := range profileIDs {
		assignment, assignmentFound := gatewayProviderAttemptAssignment(snapshot, profileID)
		expectedBinding, bindingFound := gatewayProviderRouteBinding(snapshot.InventoryBindings, profileID)
		if !assignmentFound || !bindingFound || !expectedBinding.Enabled ||
			!gatewayProviderRouteHasCapability(expectedBinding.Capabilities, routeProtocol) {
			return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureInventoryMismatch}
		}
		profile, currentBinding, resolveErr := server.resolveGatewayProviderProfile(ctx, snapshot.Environment, assignment)
		if resolveErr != nil || !gatewayProviderRouteBindingsEqual(expectedBinding, currentBinding) {
			return gatewayProviderAttemptExecution{}, gatewayProviderAttemptPlanError{gatewayProviderRouteFailureInventoryMismatch}
		}
		selection := northboundSelection{
			model: strings.TrimSpace(requestedModel), source: gatewayProviderRouteSelectionSource,
			inventoryKind: "activated_provider_route_snapshot", routeConfigurationID: snapshot.ConfigurationID,
			routeGeneration: snapshot.Generation, routeSnapshotDigest: snapshot.SnapshotDigest,
		}
		selection.applyProfile(profile)
		selection.model = strings.TrimSpace(requestedModel)
		selection.upstreamModel = strings.TrimSpace(profile.ResolvedModel)
		selection.source = gatewayProviderRouteSelectionSource
		selection.inventoryKind = "activated_provider_route_snapshot"
		selection.routeConfigurationID = snapshot.ConfigurationID
		selection.routeGeneration = snapshot.Generation
		selection.routeSnapshotDigest = snapshot.SnapshotDigest
		pricingSnapshot := server.gatewayModelPricingSnapshotForSelection(requestContext, trace.requestID, selection)
		selections = append(selections, selection)
		targets = append(targets, GatewayProviderAttemptPlanTargetInput{
			Ordinal: index + 1, ProviderProfileID: profileID,
			ProviderID: selection.provider, RuntimeProfile: selection.providerProfile,
			SelectedModel: selection.model, UpstreamModel: selection.upstreamModel,
			InventoryDigest: expectedBinding.InventoryDigest, PricingSnapshot: pricingSnapshot,
		})
	}
	plan, err := buildGatewayProviderAttemptPlan(
		trace.requestID, trace.route, protocol, requestedModel, snapshot, modelRoute, mode, targets,
	)
	if err != nil {
		return gatewayProviderAttemptExecution{}, err
	}
	return gatewayProviderAttemptExecution{plan: plan, selections: selections}, nil
}

func gatewayProviderAttemptModelRoute(
	snapshot AdminProviderRouteSnapshot,
	protocol string,
	model string,
) (AdminModelRouteDefinition, bool) {
	var matched *AdminModelRouteDefinition
	for index := range snapshot.Configuration.ModelRoutes {
		route := snapshot.Configuration.ModelRoutes[index]
		if route.Protocol != protocol || route.ModelID != strings.TrimSpace(model) {
			continue
		}
		if matched != nil {
			return AdminModelRouteDefinition{}, false
		}
		copied := route
		matched = &copied
	}
	if matched == nil {
		return AdminModelRouteDefinition{}, false
	}
	return *matched, true
}

func gatewayProviderAttemptAssignment(
	snapshot AdminProviderRouteSnapshot,
	profileID string,
) (AdminProviderProfileAssignment, bool) {
	var matched *AdminProviderProfileAssignment
	for index := range snapshot.Configuration.ProviderProfiles {
		assignment := snapshot.Configuration.ProviderProfiles[index]
		if assignment.ProfileID != profileID {
			continue
		}
		if matched != nil {
			return AdminProviderProfileAssignment{}, false
		}
		copied := assignment
		matched = &copied
	}
	if matched == nil {
		return AdminProviderProfileAssignment{}, false
	}
	return *matched, true
}

func (server *Server) invokeGatewayProviderAttempts(
	ctx context.Context,
	trace *requestTrace,
	execution gatewayProviderAttemptExecution,
	canonicalRequests [][]byte,
	temperature float64,
) gatewayProviderAttemptExecutionResult {
	if trace == nil || len(canonicalRequests) != len(execution.selections) ||
		len(canonicalRequests) != len(execution.plan.Targets) || trace.gatewayRequest == nil {
		return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailurePlanUnavailable}
	}
	if ctx.Err() != nil {
		return gatewayProviderAttemptExecutionResult{FailureCode: bridge.ErrorCodeWorkerCanceled}
	}
	root, err := newGatewayProviderAttemptHistoryRecord(*trace.gatewayRequest, execution.plan)
	if err != nil || server.gatewayRequestHistoryStore.UpdateRequest(trace.gatewayRequestContext, &root) != nil {
		return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
	}
	trace.gatewayRequest = &root
	history := newGatewayProviderAttemptHistoryService(server.gatewayRequestHistoryStore)

	for index := 0; index < execution.plan.MaxAttempts; index++ {
		target := execution.plan.Targets[index]
		selection := execution.selections[index]
		trace.terminalSelection = &execution.selections[index]
		if index == 1 && ctx.Err() != nil {
			canceled, checkpointErr := history.CancelPendingFallback(
				trace.gatewayRequestContext, trace.requestID, target, time.Now().UTC(),
			)
			if checkpointErr != nil {
				return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
			}
			trace.gatewayRequest = &canceled
			server.finalizeGatewayProviderAttemptFailure(trace, history, bridge.ErrorCodeWorkerCanceled)
			return gatewayProviderAttemptExecutionResult{Selection: selection, FailureCode: bridge.ErrorCodeWorkerCanceled}
		}
		decision, quotaFailure := server.admitGatewayProviderAttempt(ctx, target.AttemptID)
		if quotaFailure != "" {
			rejected, checkpointErr := history.RejectAttemptQuota(
				trace.gatewayRequestContext, trace.requestID, target, quotaFailure, time.Now().UTC(),
			)
			if checkpointErr != nil {
				return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
			}
			trace.gatewayRequest = &rejected
			server.finalizeGatewayProviderAttemptFailure(trace, history, quotaFailure)
			return gatewayProviderAttemptExecutionResult{Selection: selection, FailureCode: quotaFailure}
		}
		started, checkpointErr := history.StartAttempt(
			trace.gatewayRequestContext, trace.requestID, target, decision.AdmissionID, time.Now().UTC(),
		)
		if checkpointErr != nil {
			return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
		}
		trace.gatewayRequest = &started
		trace.providerAttemptCount++
		trace.providerAttempted = true
		trace.providerAttemptHeaders = true
		if index == 1 {
			trace.fallbackUsed = true
		}

		envelope, bridgeErr := server.bridge.HandleEnvelope(
			withoutGatewayRequestQuotaBinding(ctx), canonicalRequests[index],
			server.buildBridgeEnvelopeOptions(selection, temperature),
		)
		if bridgeErr != nil {
			code := bridgeFailureCode(bridgeErr)
			if !server.interruptGatewayProviderAttempt(trace, history, target, code) {
				return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
			}
			server.finalizeGatewayProviderAttemptFailure(trace, history, code)
			return gatewayProviderAttemptExecutionResult{Selection: selection, FailureCode: code, FailureDetail: bridgeErr.Error()}
		}
		usage := gatewayUsageFromEnvelope(envelope)
		cost := buildGatewayRequestCostEstimate(
			true, usage, gatewayModelPricingSnapshotFromAttempt(target.PricingSnapshot),
		)
		if !strings.EqualFold(envelope.Status, "failed") && envelope.ProviderAttemptFailure == nil && envelope.Response != nil {
			completed, completeErr := history.CompleteAttempt(
				trace.gatewayRequestContext, trace.requestID, target.AttemptID,
				usage, cost, nil, false, time.Now().UTC(),
			)
			if completeErr == nil {
				trace.gatewayRequest = &completed
				finalized, finalizeErr := history.Finalize(
					trace.gatewayRequestContext, trace.requestID, GatewayRequestStatusSucceeded,
					http.StatusOK, "", "", time.Now().UTC(),
				)
				if finalizeErr == nil {
					trace.gatewayRequest = &finalized
				}
			}
			return gatewayProviderAttemptExecutionResult{Envelope: envelope, Selection: selection}
		}

		failure := envelope.ProviderAttemptFailure
		if failure == nil || !bridge.ValidProviderAttemptFailure(*failure) {
			if !server.interruptGatewayProviderAttempt(trace, history, target, "GATEWAY_INFERENCE_FAILED") {
				return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
			}
			server.finalizeGatewayProviderAttemptFailure(trace, history, "GATEWAY_INFERENCE_FAILED")
			return gatewayProviderAttemptExecutionResult{
				Selection: selection, FailureCode: "GATEWAY_INFERENCE_FAILED",
				FailureDetail: gatewayErrorMessage(envelope.Error),
			}
		}
		prepareFallback := index == 0 && execution.plan.FallbackAllowed &&
			gatewayProviderAttemptFailureEligible(*failure) && ctx.Err() == nil
		completed, completeErr := history.CompleteAttempt(
			trace.gatewayRequestContext, trace.requestID, target.AttemptID,
			usage, cost, failure, prepareFallback, time.Now().UTC(),
		)
		if completeErr != nil {
			return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailureHistoryUnavailable}
		}
		trace.gatewayRequest = &completed
		if prepareFallback {
			continue
		}
		code := gatewayErrorCode(envelope.Error)
		if ctx.Err() != nil {
			code = bridge.ErrorCodeWorkerCanceled
		}
		server.finalizeGatewayProviderAttemptFailure(trace, history, code)
		return gatewayProviderAttemptExecutionResult{
			Envelope: envelope, Selection: selection, FailureCode: code,
			FailureDetail: gatewayErrorMessage(envelope.Error),
		}
	}
	return gatewayProviderAttemptExecutionResult{FailureCode: gatewayProviderAttemptFailurePlanUnavailable}
}

func (server *Server) interruptGatewayProviderAttempt(
	trace *requestTrace,
	history gatewayProviderAttemptHistoryService,
	target GatewayProviderAttemptPlanTarget,
	failureCode string,
) bool {
	definition := lookupPlatformErrorDefinition(failureCode)
	usage := GatewayRequestUsage{Availability: GatewayRequestUsageNotReported}
	cost := buildGatewayRequestCostEstimate(
		true, usage, gatewayModelPricingSnapshotFromAttempt(target.PricingSnapshot),
	)
	interrupted, err := history.InterruptAttempt(
		trace.gatewayRequestContext, trace.requestID, target.AttemptID,
		usage, cost, definition.failureBoundary, time.Now().UTC(),
	)
	if err != nil {
		return false
	}
	trace.gatewayRequest = &interrupted
	return true
}

func (server *Server) finalizeGatewayProviderAttemptFailure(
	trace *requestTrace,
	history gatewayProviderAttemptHistoryService,
	failureCode string,
) {
	definition := lookupPlatformErrorDefinition(failureCode)
	status := GatewayRequestStatusFailed
	if strings.TrimSpace(failureCode) == bridge.ErrorCodeWorkerCanceled {
		status = GatewayRequestStatusCanceled
	}
	finalized, err := history.Finalize(
		trace.gatewayRequestContext, trace.requestID, status,
		definition.statusCode, strings.TrimSpace(failureCode), definition.failureBoundary, time.Now().UTC(),
	)
	if err == nil {
		trace.gatewayRequest = &finalized
	}
}
