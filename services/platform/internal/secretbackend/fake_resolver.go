package secretbackend

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const DefaultFakeResolverPolicyVersion = "fake-resolver-runtime-v1"

type FakeResolverFailureCode string

const (
	FakeResolverFailureDisabled              FakeResolverFailureCode = "fake_resolver_runtime_disabled"
	FakeResolverFailureRequestAuditMissing   FakeResolverFailureCode = "fake_resolver_runtime_request_audit_missing"
	FakeResolverFailurePolicyMissing         FakeResolverFailureCode = "fake_resolver_runtime_policy_missing"
	FakeResolverFailureEnvironmentMissing    FakeResolverFailureCode = "fake_resolver_runtime_environment_binding_missing"
	FakeResolverFailureEnvironmentDenied     FakeResolverFailureCode = "fake_resolver_runtime_environment_denied"
	FakeResolverFailureProviderMissing       FakeResolverFailureCode = "fake_resolver_runtime_provider_binding_missing"
	FakeResolverFailureProviderDenied        FakeResolverFailureCode = "fake_resolver_runtime_provider_denied"
	FakeResolverFailureProfileDenied         FakeResolverFailureCode = "fake_resolver_runtime_provider_profile_denied"
	FakeResolverFailurePlaceholderMissing    FakeResolverFailureCode = "fake_resolver_runtime_placeholder_secret_ref_missing"
	FakeResolverFailureSecretLikeInput       FakeResolverFailureCode = "fake_resolver_runtime_secret_value_detected"
	FakeResolverFailurePurposeMissing        FakeResolverFailureCode = "fake_resolver_runtime_purpose_missing"
	FakeResolverFailureOpaqueMetadataMissing FakeResolverFailureCode = "fake_resolver_runtime_opaque_handle_metadata_missing"
)

type FakeResolverConfig struct {
	Enabled                 bool
	AllowedEnvironments     []string
	AllowedProviders        []string
	AllowedProviderProfiles []string
	AllowedSecretRefKeys    []string
	PolicyVersion           string
}

type FakeResolverInput struct {
	Environment      string `json:"environment"`
	Provider         string `json:"provider"`
	ProviderProfile  string `json:"provider_profile"`
	SecretRefKey     string `json:"secret_ref_key"`
	SecretRefVersion string `json:"secret_ref_version"`
	Purpose          string `json:"purpose"`
	RequestID        string `json:"request_id"`
	AuditRef         string `json:"audit_ref"`
	PolicyVersion    string `json:"policy_version"`
}

type FakeResolverResult struct {
	CredentialHandleID  string                  `json:"credential_handle_id,omitempty"`
	CredentialKind      string                  `json:"credential_kind,omitempty"`
	Environment         string                  `json:"environment,omitempty"`
	Provider            string                  `json:"provider,omitempty"`
	ProviderProfile     string                  `json:"provider_profile,omitempty"`
	SecretRefKey        string                  `json:"secret_ref_key,omitempty"`
	SecretRefVersion    string                  `json:"secret_ref_version,omitempty"`
	RequestID           string                  `json:"request_id,omitempty"`
	AuditRef            string                  `json:"audit_ref,omitempty"`
	PolicyVersion       string                  `json:"policy_version,omitempty"`
	FailureCode         FakeResolverFailureCode `json:"failure_code,omitempty"`
	SanitizedDiagnostic string                  `json:"sanitized_diagnostic,omitempty"`
	SideEffects         SideEffectCounters      `json:"side_effects"`
}

type SideEffectCounters struct {
	SecretResolverCallCount        int `json:"secret_resolver_call_count"`
	SecretValueReadCount           int `json:"secret_value_read_count"`
	CloudSecretCallCount           int `json:"cloud_secret_call_count"`
	ProviderCallCount              int `json:"provider_call_count"`
	CredentialHandleCreatedCount   int `json:"credential_handle_created_count"`
	DatabaseConnectionCount        int `json:"database_connection_count"`
	DriverOpenCount                int `json:"driver_open_count"`
	SQLExecutionCount              int `json:"sql_execution_count"`
	SchemaMarkerReadCount          int `json:"schema_marker_read_count"`
	SchemaMarkerWriteCount         int `json:"schema_marker_write_count"`
	RepositoryModeEnablementCount  int `json:"repository_mode_enablement_count"`
	ProductionAPICallCount         int `json:"production_api_call_count"`
	ProductionAuditStoreWriteCount int `json:"production_audit_store_write_count"`
}

type FakeResolver struct {
	config fakeResolverConfig
}

type fakeResolverConfig struct {
	enabled                 bool
	allowedEnvironments     map[string]struct{}
	allowedProviders        map[string]struct{}
	allowedProviderProfiles map[string]struct{}
	allowedSecretRefKeys    map[string]struct{}
	policyVersion           string
}

func NewFakeResolver(config FakeResolverConfig) FakeResolver {
	policyVersion := strings.TrimSpace(config.PolicyVersion)
	if policyVersion == "" {
		policyVersion = DefaultFakeResolverPolicyVersion
	}
	return FakeResolver{
		config: fakeResolverConfig{
			enabled:                 config.Enabled,
			allowedEnvironments:     normalizedStringSet(config.AllowedEnvironments),
			allowedProviders:        normalizedStringSet(config.AllowedProviders),
			allowedProviderProfiles: normalizedStringSet(config.AllowedProviderProfiles),
			allowedSecretRefKeys:    normalizedStringSet(config.AllowedSecretRefKeys),
			policyVersion:           policyVersion,
		},
	}
}

func NewDisabledFakeResolver() FakeResolver {
	return NewFakeResolver(FakeResolverConfig{})
}

func (resolver FakeResolver) Resolve(input FakeResolverInput) FakeResolverResult {
	input = normalizedFakeResolverInput(input)
	result := FakeResolverResult{
		RequestID:     input.RequestID,
		AuditRef:      input.AuditRef,
		PolicyVersion: resolver.policyVersionFor(input.PolicyVersion),
		SideEffects:   SideEffectCounters{},
	}

	if !resolver.config.enabled {
		return result.withFailure(FakeResolverFailureDisabled, "fake resolver runtime is disabled")
	}
	if input.RequestID == "" || input.AuditRef == "" {
		return result.withFailure(FakeResolverFailureRequestAuditMissing, "request audit metadata is missing")
	}
	if input.PolicyVersion == "" || input.PolicyVersion != resolver.config.policyVersion {
		return result.withFailure(FakeResolverFailurePolicyMissing, "policy version is missing or unsupported")
	}
	if hasSecretLikeValue(input.stringValues()) {
		return result.withFailure(FakeResolverFailureSecretLikeInput, "secret-looking input was rejected")
	}
	if input.Environment == "" {
		return result.withFailure(FakeResolverFailureEnvironmentMissing, "environment binding is missing")
	}
	if !containsString(resolver.config.allowedEnvironments, input.Environment) {
		return result.withFailure(FakeResolverFailureEnvironmentDenied, "environment binding is not allowed")
	}
	if input.Provider == "" {
		return result.withFailure(FakeResolverFailureProviderMissing, "provider binding is missing")
	}
	if !containsString(resolver.config.allowedProviders, input.Provider) {
		return result.withFailure(FakeResolverFailureProviderDenied, "provider binding is not allowed")
	}
	if !containsString(resolver.config.allowedProviderProfiles, input.ProviderProfile) {
		return result.withFailure(FakeResolverFailureProfileDenied, "provider profile binding is not allowed")
	}
	if input.SecretRefKey == "" || input.SecretRefVersion == "" {
		return result.withFailure(FakeResolverFailurePlaceholderMissing, "placeholder secret reference is missing")
	}
	if !containsString(resolver.config.allowedSecretRefKeys, input.SecretRefKey) {
		return result.withFailure(FakeResolverFailurePlaceholderMissing, "placeholder secret reference is not allowed")
	}
	if input.Purpose == "" {
		return result.withFailure(FakeResolverFailurePurposeMissing, "credential purpose is missing")
	}

	handleID := opaqueCredentialHandleID(input)
	if handleID == "" {
		return result.withFailure(FakeResolverFailureOpaqueMetadataMissing, "opaque credential metadata is missing")
	}
	result.CredentialHandleID = handleID
	result.CredentialKind = "opaque_test_credential_handle"
	result.Environment = input.Environment
	result.Provider = input.Provider
	result.ProviderProfile = input.ProviderProfile
	result.SecretRefKey = input.SecretRefKey
	result.SecretRefVersion = input.SecretRefVersion
	return result
}

func (result FakeResolverResult) withFailure(
	code FakeResolverFailureCode,
	diagnostic string,
) FakeResolverResult {
	result.FailureCode = code
	result.SanitizedDiagnostic = diagnostic
	result.CredentialHandleID = ""
	result.CredentialKind = ""
	result.Environment = ""
	result.Provider = ""
	result.ProviderProfile = ""
	result.SecretRefKey = ""
	result.SecretRefVersion = ""
	return result
}

func (resolver FakeResolver) policyVersionFor(requestPolicy string) string {
	if strings.TrimSpace(requestPolicy) != "" {
		return strings.TrimSpace(requestPolicy)
	}
	return resolver.config.policyVersion
}

func normalizedFakeResolverInput(input FakeResolverInput) FakeResolverInput {
	input.Environment = strings.TrimSpace(input.Environment)
	input.Provider = strings.TrimSpace(input.Provider)
	input.ProviderProfile = strings.TrimSpace(input.ProviderProfile)
	input.SecretRefKey = strings.TrimSpace(input.SecretRefKey)
	input.SecretRefVersion = strings.TrimSpace(input.SecretRefVersion)
	input.Purpose = strings.TrimSpace(input.Purpose)
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.AuditRef = strings.TrimSpace(input.AuditRef)
	input.PolicyVersion = strings.TrimSpace(input.PolicyVersion)
	return input
}

func (input FakeResolverInput) stringValues() []string {
	return []string{
		input.Environment,
		input.Provider,
		input.ProviderProfile,
		input.SecretRefKey,
		input.SecretRefVersion,
		input.Purpose,
		input.RequestID,
		input.AuditRef,
		input.PolicyVersion,
	}
}

func opaqueCredentialHandleID(input FakeResolverInput) string {
	parts := []string{
		input.Environment,
		input.Provider,
		input.ProviderProfile,
		input.SecretRefKey,
		input.SecretRefVersion,
		input.Purpose,
		input.RequestID,
		input.AuditRef,
		input.PolicyVersion,
	}
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "test-handle-" + hex.EncodeToString(hash[:])[:24]
}

func normalizedStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized != "" {
			set[normalized] = struct{}{}
		}
	}
	return set
}

func containsString(set map[string]struct{}, value string) bool {
	if len(set) == 0 {
		return false
	}
	_, ok := set[value]
	return ok
}

func hasSecretLikeValue(values []string) bool {
	for _, value := range values {
		lower := strings.ToLower(strings.TrimSpace(value))
		if lower == "" {
			continue
		}
		if strings.Contains(lower, "begin private key") ||
			strings.Contains(lower, "bearer ") ||
			strings.Contains(lower, "akia") ||
			strings.Contains(lower, "authorization"+":") ||
			strings.Contains(lower, "cookie"+":") ||
			strings.Contains(lower, "password=") ||
			strings.Contains(lower, "api_key=") ||
			strings.Contains(lower, "token=") ||
			strings.Contains(lower, "://") && strings.Contains(lower, "@") {
			return true
		}
		if strings.HasPrefix(lower, "sk-") && len(lower) >= 11 {
			return true
		}
	}
	return false
}
