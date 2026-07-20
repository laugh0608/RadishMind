package config

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var oidcConfigReferencePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,159}$`)
var oidcClaimNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:-]{0,79}$`)

func applyControlPlaneReadOIDCEnvOverrides(cfg *Config) error {
	for key, target := range map[string]*string{
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ISSUER":            &cfg.ControlPlaneReadOIDCIssuer,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_DISCOVERY_URL":     &cfg.ControlPlaneReadOIDCDiscoveryURL,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_AUDIENCE":          &cfg.ControlPlaneReadOIDCAudience,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAPPING_VERSION":   &cfg.ControlPlaneReadOIDCMappingVersion,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_EVIDENCE_REF":      &cfg.ControlPlaneReadOIDCEvidenceRef,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_SUBJECT_CLAIM":     &cfg.ControlPlaneReadOIDCSubjectClaim,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_TENANT_CLAIM":      &cfg.ControlPlaneReadOIDCTenantClaim,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_PERMISSION_CLAIM":  &cfg.ControlPlaneReadOIDCPermissionClaim,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_TENANT_PERMISSION": &cfg.ControlPlaneReadOIDCTenantPermission,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_AUDIT_PERMISSION":  &cfg.ControlPlaneReadOIDCAuditPermission,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ALGORITHMS":        &cfg.ControlPlaneReadOIDCAlgorithms,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_ORIGIN":       &cfg.ControlPlaneReadOIDCJWKSOrigin,
	} {
		if value, ok := stringEnv(key); ok {
			*target = strings.TrimSpace(value)
		}
	}
	for key, target := range map[string]*time.Duration{
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_DISCOVERY_TIMEOUT":  &cfg.ControlPlaneReadOIDCDiscoveryTimeout,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_MAX_AGE":       &cfg.ControlPlaneReadOIDCJWKSMaxAge,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_JWKS_HARD_EXPIRY":   &cfg.ControlPlaneReadOIDCJWKSHardExpiry,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_ROTATION_OVERLAP":   &cfg.ControlPlaneReadOIDCRotationOverlap,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_CLOCK_SKEW":         &cfg.ControlPlaneReadOIDCClockSkew,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_TOKEN_LIFETIME": &cfg.ControlPlaneReadOIDCMaxTokenLifetime,
	} {
		if value, ok := stringEnv(key); ok {
			parsed, err := parseDurationValue(key, value)
			if err != nil {
				return err
			}
			*target = parsed
		}
	}
	for key, target := range map[string]*int{
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_RESPONSE_BYTES": &cfg.ControlPlaneReadOIDCMaxResponseBytes,
		"RADISHMIND_CONTROL_PLANE_READ_OIDC_INTEGRATION_MAX_KEYS":           &cfg.ControlPlaneReadOIDCMaxKeys,
	} {
		if value, ok := stringEnv(key); ok {
			parsed, err := parseIntValue(key, value)
			if err != nil {
				return err
			}
			*target = parsed
		}
	}
	return nil
}

func validateControlPlaneReadOIDCIntegrationConfig(cfg Config) error {
	required := map[string]string{
		"issuer": cfg.ControlPlaneReadOIDCIssuer, "discovery URL": cfg.ControlPlaneReadOIDCDiscoveryURL,
		"audience": cfg.ControlPlaneReadOIDCAudience, "mapping version": cfg.ControlPlaneReadOIDCMappingVersion,
		"evidence ref": cfg.ControlPlaneReadOIDCEvidenceRef, "subject claim": cfg.ControlPlaneReadOIDCSubjectClaim,
		"tenant claim": cfg.ControlPlaneReadOIDCTenantClaim, "permission claim": cfg.ControlPlaneReadOIDCPermissionClaim,
		"tenant permission": cfg.ControlPlaneReadOIDCTenantPermission, "audit permission": cfg.ControlPlaneReadOIDCAuditPermission,
		"algorithm allowlist": cfg.ControlPlaneReadOIDCAlgorithms, "JWKS origin": cfg.ControlPlaneReadOIDCJWKSOrigin,
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("control plane read OIDC integration requires reviewed %s", name)
		}
	}
	if !oidcConfigReferencePattern.MatchString(strings.TrimSpace(cfg.ControlPlaneReadOIDCMappingVersion)) ||
		!oidcConfigReferencePattern.MatchString(strings.TrimSpace(cfg.ControlPlaneReadOIDCEvidenceRef)) ||
		!oidcConfigReferencePattern.MatchString(strings.TrimSpace(cfg.ControlPlaneReadOIDCTenantPermission)) ||
		!oidcConfigReferencePattern.MatchString(strings.TrimSpace(cfg.ControlPlaneReadOIDCAuditPermission)) ||
		strings.TrimSpace(cfg.ControlPlaneReadOIDCTenantPermission) == strings.TrimSpace(cfg.ControlPlaneReadOIDCAuditPermission) ||
		len(strings.TrimSpace(cfg.ControlPlaneReadOIDCAudience)) > 256 {
		return fmt.Errorf("control plane read OIDC integration reviewed mapping identifiers are invalid")
	}
	claimNames := []string{cfg.ControlPlaneReadOIDCSubjectClaim, cfg.ControlPlaneReadOIDCTenantClaim, cfg.ControlPlaneReadOIDCPermissionClaim}
	seenClaims := map[string]bool{}
	for _, claimName := range claimNames {
		claimName = strings.TrimSpace(claimName)
		if !oidcClaimNamePattern.MatchString(claimName) || seenClaims[claimName] {
			return fmt.Errorf("control plane read OIDC integration claim mapping is invalid")
		}
		seenClaims[claimName] = true
	}
	issuer, err := url.Parse(strings.TrimSpace(cfg.ControlPlaneReadOIDCIssuer))
	if err != nil || issuer.RawQuery != "" || issuer.Fragment != "" || issuer.User != nil || !validOIDCIntegrationScheme(issuer) {
		return fmt.Errorf("control plane read OIDC integration issuer must be an exact HTTPS URL or test-only loopback HTTP URL")
	}
	discovery, err := url.Parse(strings.TrimSpace(cfg.ControlPlaneReadOIDCDiscoveryURL))
	if err != nil || discovery.RawQuery != "" || discovery.Fragment != "" || discovery.User != nil || !sameURLOrigin(issuer, discovery) {
		return fmt.Errorf("control plane read OIDC integration discovery URL must use the exact issuer origin")
	}
	origin, err := url.Parse(strings.TrimSpace(cfg.ControlPlaneReadOIDCJWKSOrigin))
	if err != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || origin.User != nil || !validOIDCIntegrationScheme(origin) {
		return fmt.Errorf("control plane read OIDC integration JWKS origin must be an exact origin")
	}
	allowed := map[string]bool{"RS256": true, "RS384": true, "RS512": true, "ES256": true, "ES384": true, "ES512": true}
	seen := map[string]bool{}
	for _, algorithm := range strings.Split(cfg.ControlPlaneReadOIDCAlgorithms, ",") {
		algorithm = strings.TrimSpace(algorithm)
		if !allowed[algorithm] || seen[algorithm] {
			return fmt.Errorf("control plane read OIDC integration algorithm allowlist is invalid")
		}
		seen[algorithm] = true
	}
	if cfg.ControlPlaneReadOIDCDiscoveryTimeout <= 0 || cfg.ControlPlaneReadOIDCJWKSMaxAge <= 0 ||
		cfg.ControlPlaneReadOIDCJWKSHardExpiry <= cfg.ControlPlaneReadOIDCJWKSMaxAge || cfg.ControlPlaneReadOIDCRotationOverlap <= 0 ||
		cfg.ControlPlaneReadOIDCRotationOverlap > cfg.ControlPlaneReadOIDCJWKSHardExpiry || cfg.ControlPlaneReadOIDCClockSkew < 0 ||
		cfg.ControlPlaneReadOIDCMaxTokenLifetime <= 0 || cfg.ControlPlaneReadOIDCMaxResponseBytes < 1024 ||
		cfg.ControlPlaneReadOIDCMaxResponseBytes > 4*1024*1024 || cfg.ControlPlaneReadOIDCMaxKeys < 1 || cfg.ControlPlaneReadOIDCMaxKeys > 128 {
		return fmt.Errorf("control plane read OIDC integration timeout, cache, token, or response policy is invalid")
	}
	if EffectiveControlPlaneReadStoreMode(cfg) != "postgres_dev_test" {
		return fmt.Errorf("control plane read radish_oidc_integration_test requires postgres_dev_test store mode")
	}
	return nil
}

func ValidateControlPlaneReadOIDCIntegration(cfg Config) error {
	if EffectiveControlPlaneReadAuthMode(cfg) != "radish_oidc_integration_test" {
		return fmt.Errorf("control plane read OIDC integration validator requires radish_oidc_integration_test auth mode")
	}
	return validateControlPlaneReadOIDCIntegrationConfig(cfg)
}

func validOIDCIntegrationScheme(value *url.URL) bool {
	if value == nil || value.Host == "" {
		return false
	}
	if value.Scheme == "https" {
		return true
	}
	host := value.Hostname()
	return value.Scheme == "http" && (host == "localhost" || net.ParseIP(host) != nil && net.ParseIP(host).IsLoopback())
}

func sameURLOrigin(left *url.URL, right *url.URL) bool {
	return left != nil && right != nil && left.Scheme == right.Scheme && strings.EqualFold(left.Host, right.Host)
}
