package bridge

import "testing"

func TestProviderAttemptFailureEligibilityIsExplicitAndClosed(t *testing.T) {
	for _, failureClass := range []ProviderAttemptFailureClass{
		ProviderFailureRateLimited,
		ProviderFailureTemporarilyUnavailable,
		ProviderFailureUpstreamGatewayUnavailable,
	} {
		failure, ok := NewProviderAttemptFailure(
			failureClass, ProviderFallbackEligible, false, ProviderAttemptFailed,
			"PROVIDER_TEMPORARY_FAILURE", "5xx",
		)
		if !ok || !ProviderAttemptFailureEligible(failure) {
			t.Fatalf("eligible failure rejected: %#v", failure)
		}
	}

	for name, failure := range map[string]ProviderAttemptFailure{
		"auth cannot be eligible": {
			SchemaVersion: ProviderAttemptFailureSchemaVersion, FailureClass: ProviderFailureAuthentication,
			FallbackDisposition: ProviderFallbackEligible, Outcome: ProviderAttemptFailed, Code: "PROVIDER_AUTH_FAILED",
		},
		"response started": {
			SchemaVersion: ProviderAttemptFailureSchemaVersion, FailureClass: ProviderFailureRateLimited,
			FallbackDisposition: ProviderFallbackEligible, ProviderResponseStarted: true,
			Outcome: ProviderAttemptFailed, Code: "PROVIDER_RATE_LIMITED",
		},
		"unknown outcome": {
			SchemaVersion: ProviderAttemptFailureSchemaVersion, FailureClass: ProviderFailureTemporarilyUnavailable,
			FallbackDisposition: ProviderFallbackEligible, Outcome: ProviderAttemptUnknown,
			Code: "PROVIDER_OUTCOME_UNKNOWN",
		},
		"unknown class": {
			SchemaVersion: ProviderAttemptFailureSchemaVersion, FailureClass: "provider_other",
			FallbackDisposition: ProviderFallbackIneligible, Outcome: ProviderAttemptFailed, Code: "PROVIDER_OTHER",
		},
		"sensitive code": {
			SchemaVersion: ProviderAttemptFailureSchemaVersion, FailureClass: ProviderFailureAuthentication,
			FallbackDisposition: ProviderFallbackIneligible, Outcome: ProviderAttemptFailed,
			Code: "PROVIDER_TOKEN_LEAK",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if ValidProviderAttemptFailure(failure) || ProviderAttemptFailureEligible(failure) {
				t.Fatalf("unsafe failure accepted: %#v", failure)
			}
		})
	}
}

func TestProviderAttemptFailureAllowsTypedIneligibleTerminalFailure(t *testing.T) {
	failure, ok := NewProviderAttemptFailure(
		ProviderFailureInvalidRequest, ProviderFallbackIneligible, false, ProviderAttemptFailed,
		"PROVIDER_INVALID_REQUEST", "4xx",
	)
	if !ok || !ValidProviderAttemptFailure(failure) || ProviderAttemptFailureEligible(failure) {
		t.Fatalf("typed ineligible failure drifted: %#v", failure)
	}
}
