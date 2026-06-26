package secretbackend

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFakeResolverZeroValueAndDisabledConstructorAreDisabled(t *testing.T) {
	cases := map[string]FakeResolver{
		"zero value":           {},
		"disabled constructor": NewDisabledFakeResolver(),
	}
	for name, resolver := range cases {
		t.Run(name, func(t *testing.T) {
			result := resolver.Resolve(validFakeResolverInput())

			if result.FailureCode != FakeResolverFailureDisabled {
				t.Fatalf("expected disabled failure, got %#v", result)
			}
			if result.CredentialHandleID != "" || result.CredentialKind != "" {
				t.Fatalf("disabled resolver must not return credential metadata: %#v", result)
			}
			assertNoSideEffects(t, result.SideEffects)
		})
	}
}

func TestFakeResolverReturnsOpaqueMetadataForAllowedTestFixture(t *testing.T) {
	result := validFakeResolver().Resolve(validFakeResolverInput())

	if result.FailureCode != "" {
		t.Fatalf("unexpected failure: %#v", result)
	}
	if !strings.HasPrefix(result.CredentialHandleID, "test-handle-") {
		t.Fatalf("expected opaque test handle id, got %q", result.CredentialHandleID)
	}
	if result.CredentialKind != "opaque_test_credential_handle" {
		t.Fatalf("unexpected credential kind: %s", result.CredentialKind)
	}
	if result.Environment != "test" || result.Provider != "mock" || result.ProviderProfile != "local-smoke" {
		t.Fatalf("unexpected binding metadata: %#v", result)
	}
	if result.SecretRefKey != "placeholder/provider/mock/local-smoke" {
		t.Fatalf("unexpected secret ref key: %s", result.SecretRefKey)
	}
	assertNoSideEffects(t, result.SideEffects)
	assertNoLeakage(t, result, "credential_payload", "raw_secret", "database_hostname")
}

func TestFakeResolverFailsClosedForUnsupportedEnvironment(t *testing.T) {
	input := validFakeResolverInput()
	input.Environment = "production"

	result := validFakeResolver().Resolve(input)

	if result.FailureCode != FakeResolverFailureEnvironmentDenied {
		t.Fatalf("expected environment denied, got %#v", result)
	}
	if result.CredentialHandleID != "" || result.SecretRefKey != "" {
		t.Fatalf("failure result must not expose handle or secret ref metadata: %#v", result)
	}
	assertNoSideEffects(t, result.SideEffects)
}

func TestFakeResolverRejectsSecretLookingInputWithoutEchoingIt(t *testing.T) {
	secretLikeValue := "sk-" + strings.Repeat("x", 12)
	input := validFakeResolverInput()
	input.SecretRefKey = secretLikeValue

	result := validFakeResolver().Resolve(input)

	if result.FailureCode != FakeResolverFailureSecretLikeInput {
		t.Fatalf("expected secret-looking input failure, got %#v", result)
	}
	assertNoSideEffects(t, result.SideEffects)
	assertNoLeakage(t, result, secretLikeValue)
}

func TestFakeResolverRequiresPlaceholderSecretRef(t *testing.T) {
	input := validFakeResolverInput()
	input.SecretRefKey = "placeholder/provider/mock/unknown"

	result := validFakeResolver().Resolve(input)

	if result.FailureCode != FakeResolverFailurePlaceholderMissing {
		t.Fatalf("expected placeholder failure, got %#v", result)
	}
	assertNoSideEffects(t, result.SideEffects)
}

func validFakeResolver() FakeResolver {
	return NewFakeResolver(FakeResolverConfig{
		Enabled:                 true,
		AllowedEnvironments:     []string{"test"},
		AllowedProviders:        []string{"mock"},
		AllowedProviderProfiles: []string{"local-smoke"},
		AllowedSecretRefKeys:    []string{"placeholder/provider/mock/local-smoke"},
		PolicyVersion:           DefaultFakeResolverPolicyVersion,
	})
}

func validFakeResolverInput() FakeResolverInput {
	return FakeResolverInput{
		Environment:      "test",
		Provider:         "mock",
		ProviderProfile:  "local-smoke",
		SecretRefKey:     "placeholder/provider/mock/local-smoke",
		SecretRefVersion: "v1",
		Purpose:          "workflow-saved-draft-database-smoke",
		RequestID:        "req-runtime-smoke-001",
		AuditRef:         "audit-runtime-smoke-001",
		PolicyVersion:    DefaultFakeResolverPolicyVersion,
	}
}

func assertNoSideEffects(t *testing.T, counters SideEffectCounters) {
	t.Helper()
	if counters != (SideEffectCounters{}) {
		t.Fatalf("expected zero side effects, got %#v", counters)
	}
}

func assertNoLeakage(t *testing.T, result FakeResolverResult, forbiddenValues ...string) {
	t.Helper()
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range forbiddenValues {
		if forbidden != "" && strings.Contains(text, forbidden) {
			t.Fatalf("result leaked %q in %s", forbidden, text)
		}
	}
}
