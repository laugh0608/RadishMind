package bridge

import (
	"regexp"
	"strings"
)

const ProviderAttemptFailureSchemaVersion = "gateway_provider_attempt_failure.v1"

type ProviderAttemptFailureClass string
type ProviderAttemptFallbackDisposition string
type ProviderAttemptOutcome string

const (
	ProviderFailureRateLimited                ProviderAttemptFailureClass = "provider_rate_limited"
	ProviderFailureTemporarilyUnavailable     ProviderAttemptFailureClass = "provider_temporarily_unavailable"
	ProviderFailureUpstreamGatewayUnavailable ProviderAttemptFailureClass = "provider_upstream_gateway_unavailable"
	ProviderFailureAuthentication             ProviderAttemptFailureClass = "provider_authentication_failed"
	ProviderFailureInvalidRequest             ProviderAttemptFailureClass = "provider_invalid_request"
	ProviderFailureModelNotFound              ProviderAttemptFailureClass = "provider_model_not_found"
	ProviderFailureUnsupported                ProviderAttemptFailureClass = "provider_unsupported"
	ProviderFailureSafety                     ProviderAttemptFailureClass = "provider_safety_rejected"
)

const (
	ProviderFallbackEligible   ProviderAttemptFallbackDisposition = "eligible"
	ProviderFallbackIneligible ProviderAttemptFallbackDisposition = "ineligible"
	ProviderAttemptFailed      ProviderAttemptOutcome             = "failed"
	ProviderAttemptUnknown     ProviderAttemptOutcome             = "unknown"
)

type ProviderAttemptFailure struct {
	SchemaVersion           string                             `json:"schema_version"`
	FailureClass            ProviderAttemptFailureClass        `json:"failure_class"`
	FallbackDisposition     ProviderAttemptFallbackDisposition `json:"fallback_disposition"`
	ProviderResponseStarted bool                               `json:"provider_response_started"`
	Outcome                 ProviderAttemptOutcome             `json:"outcome"`
	Code                    string                             `json:"code"`
	HTTPStatusClass         string                             `json:"http_status_class,omitempty"`
}

var providerAttemptFailureCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{2,95}$`)

func NewProviderAttemptFailure(
	failureClass ProviderAttemptFailureClass,
	disposition ProviderAttemptFallbackDisposition,
	responseStarted bool,
	outcome ProviderAttemptOutcome,
	code string,
	httpStatusClass string,
) (ProviderAttemptFailure, bool) {
	failure := ProviderAttemptFailure{
		SchemaVersion: ProviderAttemptFailureSchemaVersion,
		FailureClass:  failureClass, FallbackDisposition: disposition,
		ProviderResponseStarted: responseStarted, Outcome: outcome,
		Code: strings.TrimSpace(code), HTTPStatusClass: strings.TrimSpace(httpStatusClass),
	}
	return failure, ValidProviderAttemptFailure(failure)
}

func ValidProviderAttemptFailure(failure ProviderAttemptFailure) bool {
	if failure.SchemaVersion != ProviderAttemptFailureSchemaVersion ||
		!validProviderAttemptFailureClass(failure.FailureClass) ||
		(failure.FallbackDisposition != ProviderFallbackEligible && failure.FallbackDisposition != ProviderFallbackIneligible) ||
		(failure.Outcome != ProviderAttemptFailed && failure.Outcome != ProviderAttemptUnknown) ||
		!providerAttemptFailureCodePattern.MatchString(failure.Code) ||
		(failure.HTTPStatusClass != "" && failure.HTTPStatusClass != "4xx" && failure.HTTPStatusClass != "5xx") ||
		providerAttemptFailureContainsSensitiveMaterial(failure.Code) {
		return false
	}
	if failure.FallbackDisposition == ProviderFallbackEligible {
		return providerAttemptFailureClassEligible(failure.FailureClass) &&
			failure.Outcome == ProviderAttemptFailed && !failure.ProviderResponseStarted
	}
	return true
}

func ProviderAttemptFailureEligible(failure ProviderAttemptFailure) bool {
	return ValidProviderAttemptFailure(failure) &&
		failure.FallbackDisposition == ProviderFallbackEligible &&
		providerAttemptFailureClassEligible(failure.FailureClass) &&
		failure.Outcome == ProviderAttemptFailed &&
		!failure.ProviderResponseStarted
}

func validProviderAttemptFailureClass(value ProviderAttemptFailureClass) bool {
	switch value {
	case ProviderFailureRateLimited, ProviderFailureTemporarilyUnavailable,
		ProviderFailureUpstreamGatewayUnavailable, ProviderFailureAuthentication,
		ProviderFailureInvalidRequest, ProviderFailureModelNotFound,
		ProviderFailureUnsupported, ProviderFailureSafety:
		return true
	default:
		return false
	}
}

func providerAttemptFailureClassEligible(value ProviderAttemptFailureClass) bool {
	return value == ProviderFailureRateLimited || value == ProviderFailureTemporarilyUnavailable ||
		value == ProviderFailureUpstreamGatewayUnavailable
}

func providerAttemptFailureContainsSensitiveMaterial(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"http://", "https://", "authorization", "bearer", "api_key", "token", "secret", "password", "credential"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
