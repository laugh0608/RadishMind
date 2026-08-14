package httpapi

import (
	"context"
	"errors"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

type gatewayRequestQuotaBindingKey struct{}

type gatewayRequestQuotaBinding struct {
	QuotaContext GatewayRequestQuotaContext
	APIKeyID     string
	RequestID    string
	Route        string
}

type gatewayRequestQuotaAdmissionError struct {
	FailureCode string
}

func (failure gatewayRequestQuotaAdmissionError) Error() string {
	return failure.FailureCode
}

type gatewayRequestQuotaBridgeClient struct {
	inner      bridgeClient
	repository GatewayRequestQuotaRepository
	now        func() time.Time
}

func newGatewayRequestQuotaBridgeClient(
	inner bridgeClient,
	repository GatewayRequestQuotaRepository,
) *gatewayRequestQuotaBridgeClient {
	return &gatewayRequestQuotaBridgeClient{inner: inner, repository: repository, now: time.Now}
}

func (client *gatewayRequestQuotaBridgeClient) DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error) {
	return client.inner.DescribeProviders(ctx)
}

func (client *gatewayRequestQuotaBridgeClient) DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error) {
	return client.inner.DescribeInventory(ctx)
}

func (client *gatewayRequestQuotaBridgeClient) HandleEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options bridge.EnvelopeOptions,
) (bridge.GatewayEnvelope, error) {
	if err := client.admit(ctx); err != nil {
		return bridge.GatewayEnvelope{}, err
	}
	return client.inner.HandleEnvelope(ctx, canonicalRequest, options)
}

func (client *gatewayRequestQuotaBridgeClient) StreamEnvelope(
	ctx context.Context,
	canonicalRequest []byte,
	options bridge.EnvelopeOptions,
	handleEvent func(bridge.StreamEvent) error,
) error {
	if err := client.admit(ctx); err != nil {
		return err
	}
	return client.inner.StreamEnvelope(ctx, canonicalRequest, options, handleEvent)
}

func (client *gatewayRequestQuotaBridgeClient) Close() {
	if closer, ok := client.inner.(interface{ Close() }); ok {
		closer.Close()
	}
}

func (client *gatewayRequestQuotaBridgeClient) admit(ctx context.Context) error {
	binding, found := gatewayRequestQuotaBindingFromContext(ctx)
	if !found {
		return nil
	}
	if client.repository == nil {
		return gatewayRequestQuotaAdmissionError{FailureCode: GatewayRequestQuotaFailureStoreUnavailable}
	}
	binding.QuotaContext.RequestContext = ctx
	_, err := client.repository.AdmitProviderAttempt(binding.QuotaContext, GatewayRequestQuotaAdmissionInput{
		APIKeyID: binding.APIKeyID, RequestID: binding.RequestID, Route: binding.Route, AdmittedAt: client.now().UTC(),
	})
	if err == nil {
		return nil
	}
	return gatewayRequestQuotaAdmissionError{FailureCode: gatewayRequestQuotaFailureFromRepositoryError(err)}
}

func withGatewayRequestQuotaBinding(ctx context.Context, binding gatewayRequestQuotaBinding) context.Context {
	return context.WithValue(ctx, gatewayRequestQuotaBindingKey{}, binding)
}

func withGatewayRequestQuotaAttempt(ctx context.Context, requestID string) context.Context {
	binding, found := gatewayRequestQuotaBindingFromContext(ctx)
	if !found {
		return ctx
	}
	binding.RequestID = strings.TrimSpace(requestID)
	binding.QuotaContext.RequestID = binding.RequestID
	return withGatewayRequestQuotaBinding(ctx, binding)
}

func gatewayRequestQuotaBindingFromContext(ctx context.Context) (gatewayRequestQuotaBinding, bool) {
	if ctx == nil {
		return gatewayRequestQuotaBinding{}, false
	}
	binding, ok := ctx.Value(gatewayRequestQuotaBindingKey{}).(gatewayRequestQuotaBinding)
	return binding, ok && strings.TrimSpace(binding.APIKeyID) != ""
}

func withoutGatewayRequestQuotaBinding(ctx context.Context) context.Context {
	return context.WithValue(ctx, gatewayRequestQuotaBindingKey{}, gatewayRequestQuotaBinding{})
}

func (server *Server) admitGatewayProviderAttempt(
	ctx context.Context,
	attemptID string,
) (GatewayRequestQuotaAdmissionDecision, string) {
	binding, found := gatewayRequestQuotaBindingFromContext(ctx)
	if !found || server.gatewayRequestQuotaRepository == nil {
		return GatewayRequestQuotaAdmissionDecision{}, GatewayRequestQuotaFailureStoreUnavailable
	}
	attemptID = strings.TrimSpace(attemptID)
	binding.QuotaContext.RequestContext = ctx
	binding.QuotaContext.RequestID = attemptID
	decision, err := server.gatewayRequestQuotaRepository.AdmitProviderAttempt(
		binding.QuotaContext,
		GatewayRequestQuotaAdmissionInput{
			APIKeyID: binding.APIKeyID, RequestID: attemptID, Route: binding.Route, AdmittedAt: time.Now().UTC(),
		},
	)
	if err != nil {
		return GatewayRequestQuotaAdmissionDecision{}, gatewayRequestQuotaFailureFromRepositoryError(err)
	}
	return decision, ""
}

func gatewayRequestQuotaFailureFromRepositoryError(err error) string {
	switch {
	case errors.Is(err, errGatewayRequestQuotaPolicyNotFound):
		return GatewayRequestQuotaFailurePolicyNotFound
	case errors.Is(err, errGatewayRequestQuotaPolicyVersionConflict):
		return GatewayRequestQuotaFailurePolicyVersionConflict
	case errors.Is(err, errGatewayRequestQuotaAttemptConflict):
		return GatewayRequestQuotaFailureAttemptConflict
	case errors.Is(err, errGatewayRequestQuotaExceeded):
		return GatewayRequestQuotaFailureExceeded
	case errors.Is(err, errGatewayRequestQuotaContract):
		return GatewayRequestQuotaFailurePayloadInvalid
	default:
		return GatewayRequestQuotaFailureStoreUnavailable
	}
}

func gatewayRequestQuotaFailureCode(err error) string {
	var quotaFailure gatewayRequestQuotaAdmissionError
	if errors.As(err, &quotaFailure) {
		return strings.TrimSpace(quotaFailure.FailureCode)
	}
	return ""
}

func gatewayRequestQuotaFailureCodeFromValue(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case GatewayRequestQuotaFailureDisabled, GatewayRequestQuotaFailureScopeDenied,
		GatewayRequestQuotaFailureEnvironmentForbidden, GatewayRequestQuotaFailurePayloadInvalid,
		GatewayRequestQuotaFailurePolicyNotFound, GatewayRequestQuotaFailurePolicyVersionConflict,
		GatewayRequestQuotaFailureAttemptConflict, GatewayRequestQuotaFailureExceeded,
		GatewayRequestQuotaFailureStoreUnavailable:
		return value
	default:
		return ""
	}
}

func gatewayRequestQuotaHTTPStatus(failureCode string) int {
	switch failureCode {
	case GatewayRequestQuotaFailureExceeded:
		return 429
	case GatewayRequestQuotaFailureAttemptConflict, GatewayRequestQuotaFailurePolicyVersionConflict:
		return 409
	case GatewayRequestQuotaFailureDisabled, GatewayRequestQuotaFailureScopeDenied,
		GatewayRequestQuotaFailureEnvironmentForbidden:
		return 403
	case GatewayRequestQuotaFailurePolicyNotFound, GatewayRequestQuotaFailureStoreUnavailable:
		return 503
	default:
		return 400
	}
}
