package httpapi

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"radishmind.local/services/platform/internal/config"
)

var errOIDCIdentityProviderUnavailable = errors.New("OIDC identity provider unavailable")
var errOIDCTenantBindingMissing = errors.New("OIDC tenant binding missing")

type oidcDiscoveryDocument struct {
	Issuer            string   `json:"issuer"`
	JWKSURI           string   `json:"jwks_uri"`
	SigningAlgorithms []string `json:"id_token_signing_alg_values_supported"`
}

type oidcJWKSet struct {
	Keys []oidcJWK `json:"keys"`
}

type oidcJWK struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
	Curve     string `json:"crv"`
	X         string `json:"x"`
	Y         string `json:"y"`
}

type oidcVerificationKey struct {
	key       crypto.PublicKey
	algorithm string
	expiresAt time.Time
}

type oidcKeyCache struct {
	keys         map[string]oidcVerificationKey
	refreshAfter time.Time
}

type oidcVerifierPolicy struct {
	issuer           string
	discoveryURL     string
	audience         string
	mappingVersion   string
	evidenceRef      string
	subjectClaim     string
	tenantClaim      string
	permissionClaim  string
	tenantPermission string
	auditPermission  string
	algorithms       map[string]bool
	jwksOrigin       string
	discoveryTimeout time.Duration
	jwksMaxAge       time.Duration
	jwksHardExpiry   time.Duration
	rotationOverlap  time.Duration
	clockSkew        time.Duration
	maxTokenLifetime time.Duration
	maxResponseBytes int64
	maxKeys          int
}

type oidcTokenVerifier struct {
	policy     oidcVerifierPolicy
	httpClient *http.Client
	now        func() time.Time
	jwksURI    string
	mu         sync.RWMutex
	cache      oidcKeyCache
	refresh    singleflight.Group
}

func newOIDCTokenVerifier(ctx context.Context, cfg config.Config) (*oidcTokenVerifier, error) {
	policy := oidcVerifierPolicyFromConfig(cfg)
	verifier := &oidcTokenVerifier{
		policy: policy,
		now:    time.Now,
		httpClient: &http.Client{
			Timeout: policy.discoveryTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return errors.New("OIDC redirects are disabled")
			},
		},
	}
	discovery, err := verifier.fetchDiscovery(ctx)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery preflight failed: %w", errOIDCIdentityProviderUnavailable)
	}
	verifier.jwksURI = discovery.JWKSURI
	if err := verifier.refreshKeys(ctx); err != nil {
		return nil, fmt.Errorf("OIDC JWKS preflight failed: %w", errOIDCIdentityProviderUnavailable)
	}
	return verifier, nil
}

func oidcVerifierPolicyFromConfig(cfg config.Config) oidcVerifierPolicy {
	algorithms := map[string]bool{}
	for _, algorithm := range strings.Split(cfg.ControlPlaneReadOIDCAlgorithms, ",") {
		algorithms[strings.TrimSpace(algorithm)] = true
	}
	return oidcVerifierPolicy{
		issuer: strings.TrimSpace(cfg.ControlPlaneReadOIDCIssuer), discoveryURL: strings.TrimSpace(cfg.ControlPlaneReadOIDCDiscoveryURL),
		audience: strings.TrimSpace(cfg.ControlPlaneReadOIDCAudience), mappingVersion: strings.TrimSpace(cfg.ControlPlaneReadOIDCMappingVersion),
		evidenceRef: strings.TrimSpace(cfg.ControlPlaneReadOIDCEvidenceRef), subjectClaim: strings.TrimSpace(cfg.ControlPlaneReadOIDCSubjectClaim),
		tenantClaim: strings.TrimSpace(cfg.ControlPlaneReadOIDCTenantClaim), permissionClaim: strings.TrimSpace(cfg.ControlPlaneReadOIDCPermissionClaim),
		tenantPermission: strings.TrimSpace(cfg.ControlPlaneReadOIDCTenantPermission), auditPermission: strings.TrimSpace(cfg.ControlPlaneReadOIDCAuditPermission),
		algorithms: algorithms, jwksOrigin: strings.TrimSuffix(strings.TrimSpace(cfg.ControlPlaneReadOIDCJWKSOrigin), "/"),
		discoveryTimeout: cfg.ControlPlaneReadOIDCDiscoveryTimeout, jwksMaxAge: cfg.ControlPlaneReadOIDCJWKSMaxAge,
		jwksHardExpiry: cfg.ControlPlaneReadOIDCJWKSHardExpiry, rotationOverlap: cfg.ControlPlaneReadOIDCRotationOverlap,
		clockSkew: cfg.ControlPlaneReadOIDCClockSkew, maxTokenLifetime: cfg.ControlPlaneReadOIDCMaxTokenLifetime,
		maxResponseBytes: int64(cfg.ControlPlaneReadOIDCMaxResponseBytes), maxKeys: cfg.ControlPlaneReadOIDCMaxKeys,
	}
}

func (verifier *oidcTokenVerifier) fetchDiscovery(ctx context.Context) (oidcDiscoveryDocument, error) {
	var document oidcDiscoveryDocument
	if err := verifier.fetchJSON(ctx, verifier.policy.discoveryURL, &document); err != nil {
		return document, err
	}
	if document.Issuer != verifier.policy.issuer || !sameOIDCOrigin(document.JWKSURI, verifier.policy.jwksOrigin) ||
		!containsAllowedOIDCAlgorithm(document.SigningAlgorithms, verifier.policy.algorithms) {
		return document, errors.New("OIDC discovery contract mismatch")
	}
	return document, nil
}

func (verifier *oidcTokenVerifier) fetchJSON(ctx context.Context, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	response, err := verifier.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("OIDC endpoint returned non-success status")
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || (mediaType != "application/json" && mediaType != "application/jwk-set+json") {
		return errors.New("OIDC endpoint returned invalid content type")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, verifier.policy.maxResponseBytes+1))
	if err != nil || int64(len(body)) > verifier.policy.maxResponseBytes || jsonNestingDepth(body) > 8 {
		return errors.New("OIDC endpoint response policy exceeded")
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	if err := decoder.Decode(target); err != nil {
		return errors.New("OIDC endpoint returned invalid JSON")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("OIDC endpoint returned trailing JSON")
	}
	return nil
}

func (verifier *oidcTokenVerifier) refreshKeys(ctx context.Context) error {
	_, err, _ := verifier.refresh.Do("jwks", func() (any, error) {
		var set oidcJWKSet
		if err := verifier.fetchJSON(ctx, verifier.jwksURI, &set); err != nil {
			return nil, errOIDCIdentityProviderUnavailable
		}
		if len(set.Keys) == 0 || len(set.Keys) > verifier.policy.maxKeys {
			return nil, errOIDCIdentityProviderUnavailable
		}
		now := verifier.now().UTC()
		keys := make(map[string]oidcVerificationKey, len(set.Keys))
		for _, keyDocument := range set.Keys {
			key, err := parseOIDCVerificationKey(keyDocument, verifier.policy.algorithms)
			if err != nil {
				return nil, errOIDCIdentityProviderUnavailable
			}
			if _, duplicate := keys[keyDocument.KeyID]; duplicate {
				return nil, errOIDCIdentityProviderUnavailable
			}
			key.expiresAt = now.Add(verifier.policy.jwksHardExpiry)
			keys[keyDocument.KeyID] = key
		}
		verifier.mu.Lock()
		for keyID, oldKey := range verifier.cache.keys {
			if _, present := keys[keyID]; present {
				continue
			}
			overlapExpiry := now.Add(verifier.policy.rotationOverlap)
			if oldKey.expiresAt.Before(overlapExpiry) {
				overlapExpiry = oldKey.expiresAt
			}
			if overlapExpiry.After(now) {
				if len(keys) >= verifier.policy.maxKeys {
					verifier.mu.Unlock()
					return nil, errOIDCIdentityProviderUnavailable
				}
				oldKey.expiresAt = overlapExpiry
				keys[keyID] = oldKey
			}
		}
		verifier.cache = oidcKeyCache{keys: keys, refreshAfter: now.Add(verifier.policy.jwksMaxAge)}
		verifier.mu.Unlock()
		return nil, nil
	})
	return err
}

func (verifier *oidcTokenVerifier) verificationKey(ctx context.Context, keyID string, algorithm string) (oidcVerificationKey, error) {
	now := verifier.now().UTC()
	refreshed := false
	verifier.mu.RLock()
	key, found := verifier.cache.keys[keyID]
	refreshNeeded := !verifier.cache.refreshAfter.After(now)
	verifier.mu.RUnlock()
	if refreshNeeded {
		if err := verifier.refreshKeys(ctx); err != nil && (!found || !key.expiresAt.After(now)) {
			return oidcVerificationKey{}, errOIDCIdentityProviderUnavailable
		}
		refreshed = true
		verifier.mu.RLock()
		key, found = verifier.cache.keys[keyID]
		verifier.mu.RUnlock()
	}
	if (!found || !key.expiresAt.After(now)) && !refreshed {
		if err := verifier.refreshKeys(ctx); err != nil {
			return oidcVerificationKey{}, errOIDCIdentityProviderUnavailable
		}
		verifier.mu.RLock()
		key, found = verifier.cache.keys[keyID]
		verifier.mu.RUnlock()
	}
	if !found || !key.expiresAt.After(now) || key.algorithm != algorithm {
		return oidcVerificationKey{}, errors.New("OIDC token key is invalid")
	}
	return key, nil
}

func (verifier *oidcTokenVerifier) Validate(ctx context.Context, rawToken string) (VerifiedControlPlaneIdentity, error) {
	segments := strings.Split(rawToken, ".")
	if len(segments) != 3 || len(rawToken) > 64*1024 {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC compact token is invalid")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token header is invalid")
	}
	var header signedTestTokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Type != "JWT" || !verifier.policy.algorithms[header.Algorithm] ||
		!validControlPlaneReadAuthReference(header.KeyID, false) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token header contract mismatch")
	}
	key, err := verifier.verificationKey(ctx, header.KeyID, header.Algorithm)
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	if err := verifyOIDCSignature(key.key, header.Algorithm, segments[0]+"."+segments[1], segments[2]); err != nil {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token signature is invalid")
	}
	claimBytes, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil || jsonNestingDepth(claimBytes) > 8 {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token claims are invalid")
	}
	claims := map[string]json.RawMessage{}
	decoder := json.NewDecoder(strings.NewReader(string(claimBytes)))
	if err := decoder.Decode(&claims); err != nil || len(claims) > 64 {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token claims are invalid")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token claims are invalid")
	}
	issuer, err := requiredOIDCStringClaim(claims, "iss")
	if err != nil || issuer != verifier.policy.issuer {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC issuer mismatch")
	}
	audiences, err := requiredOIDCAudiences(claims["aud"])
	if err != nil || !containsExactString(audiences, verifier.policy.audience) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC audience mismatch")
	}
	subject, err := requiredOIDCStringClaim(claims, verifier.policy.subjectClaim)
	if err != nil || !validControlPlaneReadAuthReference(subject, false) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC subject claim mismatch")
	}
	tenant, err := requiredOIDCStringClaim(claims, verifier.policy.tenantClaim)
	if err != nil || strings.TrimSpace(tenant) == "" {
		return VerifiedControlPlaneIdentity{}, errOIDCTenantBindingMissing
	}
	if !validControlPlaneReadAuthReference(tenant, false) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC tenant claim mismatch")
	}
	permissions, err := requiredOIDCStringArrayClaim(claims, verifier.policy.permissionClaim)
	if err != nil || !validControlPlaneReadPermissions(permissions) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC permission claim mismatch")
	}
	iat, err := requiredOIDCNumericDate(claims, "iat")
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	nbf, err := requiredOIDCNumericDate(claims, "nbf")
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	exp, err := requiredOIDCNumericDate(claims, "exp")
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	authTime, err := requiredOIDCNumericDate(claims, "auth_time")
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	now := verifier.now().UTC()
	if !exp.After(nbf) || exp.Sub(iat) > verifier.policy.maxTokenLifetime || authTime.After(iat) ||
		iat.After(now.Add(verifier.policy.clockSkew)) || nbf.After(now.Add(verifier.policy.clockSkew)) || !exp.After(now.Add(-verifier.policy.clockSkew)) {
		return VerifiedControlPlaneIdentity{}, errors.New("OIDC token time window is invalid")
	}
	grants := make([]string, 0, 2)
	knownPermissions := make([]string, 0, 2)
	if containsExactString(permissions, verifier.policy.tenantPermission) {
		grants = append(grants, "tenant:read")
		knownPermissions = append(knownPermissions, verifier.policy.tenantPermission)
	}
	if containsExactString(permissions, verifier.policy.auditPermission) {
		grants = append(grants, "audit:read")
		knownPermissions = append(knownPermissions, verifier.policy.auditPermission)
	}
	return VerifiedControlPlaneIdentity{
		AuthSource: controlPlaneReadAuthModeRadishOIDCIntegrationTest, IssuerRef: verifier.policy.evidenceRef,
		SubjectRef: subject, TenantRef: tenant, AudienceRefs: []string{verifier.policy.audience},
		UpstreamPermissionRefs: knownPermissions, ScopeGrants: grants, IssuedAt: iat, ExpiresAt: exp, AuthTime: authTime,
		PolicyVersion: verifier.policy.mappingVersion, SessionRef: "session:oidc-integration", KeyIDRef: "key:" + header.KeyID,
		Algorithm: header.Algorithm,
	}, nil
}

func parseOIDCVerificationKey(document oidcJWK, allowed map[string]bool) (oidcVerificationKey, error) {
	if !validControlPlaneReadAuthReference(document.KeyID, false) || !allowed[document.Algorithm] || document.Use != "sig" {
		return oidcVerificationKey{}, errors.New("invalid OIDC JWK metadata")
	}
	if strings.HasPrefix(document.Algorithm, "RS") && document.KeyType == "RSA" {
		modulusBytes, err := base64.RawURLEncoding.DecodeString(document.Modulus)
		if err != nil || len(modulusBytes) < 256 {
			return oidcVerificationKey{}, errors.New("invalid RSA modulus")
		}
		exponentBytes, err := base64.RawURLEncoding.DecodeString(document.Exponent)
		if err != nil || len(exponentBytes) == 0 || len(exponentBytes) > 4 {
			return oidcVerificationKey{}, errors.New("invalid RSA exponent")
		}
		exponent := 0
		for _, value := range exponentBytes {
			exponent = exponent<<8 + int(value)
		}
		if exponent < 3 || exponent%2 == 0 {
			return oidcVerificationKey{}, errors.New("invalid RSA exponent")
		}
		return oidcVerificationKey{key: &rsa.PublicKey{N: new(big.Int).SetBytes(modulusBytes), E: exponent}, algorithm: document.Algorithm}, nil
	}
	curveByAlgorithm := map[string]struct {
		name  string
		curve elliptic.Curve
	}{"ES256": {"P-256", elliptic.P256()}, "ES384": {"P-384", elliptic.P384()}, "ES512": {"P-521", elliptic.P521()}}
	curvePolicy, ok := curveByAlgorithm[document.Algorithm]
	if !ok || document.KeyType != "EC" || document.Curve != curvePolicy.name {
		return oidcVerificationKey{}, errors.New("invalid EC JWK metadata")
	}
	xBytes, errX := base64.RawURLEncoding.DecodeString(document.X)
	yBytes, errY := base64.RawURLEncoding.DecodeString(document.Y)
	x, y := new(big.Int).SetBytes(xBytes), new(big.Int).SetBytes(yBytes)
	if errX != nil || errY != nil || !curvePolicy.curve.IsOnCurve(x, y) {
		return oidcVerificationKey{}, errors.New("invalid EC point")
	}
	return oidcVerificationKey{key: &ecdsa.PublicKey{Curve: curvePolicy.curve, X: x, Y: y}, algorithm: document.Algorithm}, nil
}

func verifyOIDCSignature(key crypto.PublicKey, algorithm string, signed string, encodedSignature string) error {
	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return err
	}
	var digest []byte
	var hash crypto.Hash
	switch algorithm {
	case "RS256", "ES256":
		value := sha256.Sum256([]byte(signed))
		digest, hash = value[:], crypto.SHA256
	case "RS384", "ES384":
		value := sha512.Sum384([]byte(signed))
		digest, hash = value[:], crypto.SHA384
	case "RS512", "ES512":
		value := sha512.Sum512([]byte(signed))
		digest, hash = value[:], crypto.SHA512
	default:
		return errors.New("unsupported OIDC algorithm")
	}
	if rsaKey, ok := key.(*rsa.PublicKey); ok && strings.HasPrefix(algorithm, "RS") {
		return rsa.VerifyPKCS1v15(rsaKey, hash, digest, signature)
	}
	if ecKey, ok := key.(*ecdsa.PublicKey); ok && strings.HasPrefix(algorithm, "ES") {
		partSize := (ecKey.Curve.Params().BitSize + 7) / 8
		if len(signature) != partSize*2 || !ecdsa.Verify(ecKey, digest, new(big.Int).SetBytes(signature[:partSize]), new(big.Int).SetBytes(signature[partSize:])) {
			return errors.New("invalid ECDSA signature")
		}
		return nil
	}
	return errors.New("OIDC algorithm and key type mismatch")
}

func requiredOIDCStringClaim(claims map[string]json.RawMessage, name string) (string, error) {
	var value string
	if name == "" || json.Unmarshal(claims[name], &value) != nil || strings.TrimSpace(value) == "" {
		return "", errors.New("required OIDC string claim is invalid")
	}
	return strings.TrimSpace(value), nil
}

func requiredOIDCStringArrayClaim(claims map[string]json.RawMessage, name string) ([]string, error) {
	var values []string
	if name == "" || json.Unmarshal(claims[name], &values) != nil || len(values) == 0 || len(values) > 64 {
		return nil, errors.New("required OIDC array claim is invalid")
	}
	seen := map[string]bool{}
	for index, value := range values {
		values[index] = strings.TrimSpace(value)
		if values[index] == "" || seen[values[index]] {
			return nil, errors.New("required OIDC array claim is invalid")
		}
		seen[values[index]] = true
	}
	return values, nil
}

func requiredOIDCAudiences(raw json.RawMessage) ([]string, error) {
	var one string
	if json.Unmarshal(raw, &one) == nil && strings.TrimSpace(one) != "" {
		return []string{strings.TrimSpace(one)}, nil
	}
	claims := map[string]json.RawMessage{"aud": raw}
	return requiredOIDCStringArrayClaim(claims, "aud")
}

func requiredOIDCNumericDate(claims map[string]json.RawMessage, name string) (time.Time, error) {
	var value int64
	if json.Unmarshal(claims[name], &value) != nil || value <= 0 {
		return time.Time{}, errors.New("required OIDC numeric date is invalid")
	}
	return time.Unix(value, 0).UTC(), nil
}

func sameOIDCOrigin(endpoint string, expectedOrigin string) bool {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	origin := parsed.Scheme + "://" + parsed.Host
	return origin == expectedOrigin
}

func containsAllowedOIDCAlgorithm(discovered []string, allowed map[string]bool) bool {
	if len(discovered) == 0 {
		return false
	}
	for algorithm := range allowed {
		if !containsExactString(discovered, algorithm) {
			return false
		}
	}
	return true
}

func jsonNestingDepth(data []byte) int {
	depth, maximum := 0, 0
	inString, escaped := false, false
	for _, value := range data {
		if inString {
			if escaped {
				escaped = false
			} else if value == '\\' {
				escaped = true
			} else if value == '"' {
				inString = false
			}
			continue
		}
		if value == '"' {
			inString = true
			continue
		}
		if value == '{' || value == '[' {
			depth++
			if depth > maximum {
				maximum = depth
			}
		}
		if value == '}' || value == ']' {
			depth--
		}
	}
	return maximum
}
