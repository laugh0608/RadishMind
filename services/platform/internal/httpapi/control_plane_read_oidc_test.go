package httpapi

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	testOIDCAudience         = "audience:radishmind-integration"
	testOIDCMappingVersion   = "mapping:radish-oidc-v1"
	testOIDCTenantPermission = "permission:tenant-summary-read"
	testOIDCAuditPermission  = "permission:audit-summary-read"
)

type deterministicOIDCIssuer struct {
	testing       *testing.T
	server        *httptest.Server
	mu            sync.RWMutex
	keys          map[string]*rsa.PrivateKey
	discoveryHits int
	jwksHits      int
	contentType   string
	discoveryBody []byte
	jwksBody      []byte
	redirectJWKS  string
	delay         time.Duration
}

func newDeterministicOIDCIssuer(t *testing.T, keys map[string]*rsa.PrivateKey) *deterministicOIDCIssuer {
	t.Helper()
	issuer := &deterministicOIDCIssuer{testing: t, keys: keys, contentType: "application/json"}
	issuer.server = httptest.NewServer(http.HandlerFunc(issuer.handle))
	t.Cleanup(issuer.server.Close)
	return issuer
}

func (issuer *deterministicOIDCIssuer) handle(writer http.ResponseWriter, request *http.Request) {
	issuer.mu.Lock()
	delay := issuer.delay
	contentType := issuer.contentType
	if request.URL.Path == "/.well-known/openid-configuration" {
		issuer.discoveryHits++
		body := issuer.discoveryBody
		if body == nil {
			body, _ = json.Marshal(map[string]any{
				"issuer": issuer.server.URL, "jwks_uri": issuer.server.URL + "/jwks",
				"id_token_signing_alg_values_supported": []string{"RS256"},
			})
		}
		issuer.mu.Unlock()
		time.Sleep(delay)
		writer.Header().Set("Content-Type", contentType)
		_, _ = writer.Write(body)
		return
	}
	if request.URL.Path == "/jwks" {
		issuer.jwksHits++
		redirect := issuer.redirectJWKS
		body := issuer.jwksBody
		keys := make(map[string]*rsa.PrivateKey, len(issuer.keys))
		for keyID, key := range issuer.keys {
			keys[keyID] = key
		}
		issuer.mu.Unlock()
		time.Sleep(delay)
		if redirect != "" {
			http.Redirect(writer, request, redirect, http.StatusFound)
			return
		}
		if body == nil {
			body = marshalOIDCJWKS(keys)
		}
		writer.Header().Set("Content-Type", contentType)
		_, _ = writer.Write(body)
		return
	}
	issuer.mu.Unlock()
	http.NotFound(writer, request)
}

func (issuer *deterministicOIDCIssuer) config() config.Config {
	return config.Config{
		ControlPlaneReadAuthMode:  controlPlaneReadAuthModeRadishOIDCIntegrationTest,
		ControlPlaneReadStoreMode: "postgres_dev_test", ControlPlaneReadDatabaseURL: "postgres://unused-in-verifier-test",
		ControlPlaneReadOIDCIssuer: issuer.server.URL, ControlPlaneReadOIDCDiscoveryURL: issuer.server.URL + "/.well-known/openid-configuration",
		ControlPlaneReadOIDCAudience: testOIDCAudience, ControlPlaneReadOIDCMappingVersion: testOIDCMappingVersion,
		ControlPlaneReadOIDCEvidenceRef: "issuer:test-evidence", ControlPlaneReadOIDCSubjectClaim: "sub",
		ControlPlaneReadOIDCTenantClaim: "tenant_id", ControlPlaneReadOIDCPermissionClaim: "permissions",
		ControlPlaneReadOIDCTenantPermission: testOIDCTenantPermission, ControlPlaneReadOIDCAuditPermission: testOIDCAuditPermission,
		ControlPlaneReadOIDCAlgorithms: "RS256", ControlPlaneReadOIDCJWKSOrigin: issuer.server.URL,
		ControlPlaneReadOIDCDiscoveryTimeout: time.Second, ControlPlaneReadOIDCJWKSMaxAge: time.Minute,
		ControlPlaneReadOIDCJWKSHardExpiry: 5 * time.Minute, ControlPlaneReadOIDCRotationOverlap: 2 * time.Minute,
		ControlPlaneReadOIDCClockSkew: 10 * time.Second, ControlPlaneReadOIDCMaxTokenLifetime: 5 * time.Minute,
		ControlPlaneReadOIDCMaxResponseBytes: 32 * 1024, ControlPlaneReadOIDCMaxKeys: 8,
	}
}

func TestOIDCVerifierDiscoveryAndJWKSPolicy(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	tests := []struct {
		name   string
		mutate func(*deterministicOIDCIssuer, *config.Config)
	}{
		{name: "exact issuer mismatch", mutate: func(issuer *deterministicOIDCIssuer, _ *config.Config) {
			issuer.discoveryBody, _ = json.Marshal(map[string]any{"issuer": "https://other.invalid", "jwks_uri": issuer.server.URL + "/jwks", "id_token_signing_alg_values_supported": []string{"RS256"}})
		}},
		{name: "JWKS cross origin", mutate: func(issuer *deterministicOIDCIssuer, _ *config.Config) {
			issuer.discoveryBody, _ = json.Marshal(map[string]any{"issuer": issuer.server.URL, "jwks_uri": "https://other.invalid/jwks", "id_token_signing_alg_values_supported": []string{"RS256"}})
		}},
		{name: "redirect denied", mutate: func(issuer *deterministicOIDCIssuer, _ *config.Config) {
			issuer.redirectJWKS = issuer.server.URL + "/jwks-target"
		}},
		{name: "content type denied", mutate: func(issuer *deterministicOIDCIssuer, _ *config.Config) { issuer.contentType = "text/plain" }},
		{name: "malformed JWKS", mutate: func(issuer *deterministicOIDCIssuer, _ *config.Config) { issuer.jwksBody = []byte(`{"keys":`) }},
		{name: "oversized response", mutate: func(issuer *deterministicOIDCIssuer, cfg *config.Config) {
			cfg.ControlPlaneReadOIDCMaxResponseBytes = 1024
			issuer.jwksBody = []byte(`{"keys":[] ,"padding":"` + strings.Repeat("x", 2048) + `"}`)
		}},
		{name: "timeout", mutate: func(issuer *deterministicOIDCIssuer, cfg *config.Config) {
			cfg.ControlPlaneReadOIDCDiscoveryTimeout = 5 * time.Millisecond
			issuer.delay = 25 * time.Millisecond
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
			cfg := issuer.config()
			test.mutate(issuer, &cfg)
			if _, err := newOIDCTokenVerifier(context.Background(), cfg); err == nil || strings.Contains(err.Error(), issuer.server.URL) {
				t.Fatalf("expected sanitized startup failure, got %v", err)
			}
		})
	}
}

func TestOIDCVerifierTokenValidationAndPermissionProjection(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
	verifier, err := newOIDCTokenVerifier(context.Background(), issuer.config())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	verifier.now = func() time.Time { return now }
	base := validOIDCClaims(issuer.server.URL, now)

	identity, err := verifier.Validate(context.Background(), signOIDCTestToken(t, key, "key-one", "RS256", base))
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if identity.IssuerRef != "issuer:test-evidence" || identity.PolicyVersion != testOIDCMappingVersion ||
		!containsExactString(identity.ScopeGrants, "tenant:read") || !containsExactString(identity.ScopeGrants, "audit:read") ||
		len(identity.UpstreamPermissionRefs) != 2 {
		t.Fatalf("unexpected sanitized identity: %#v", identity)
	}

	tests := []struct {
		name      string
		algorithm string
		mutate    func(map[string]any)
	}{
		{name: "algorithm confusion", algorithm: "HS256"},
		{name: "issuer mismatch", mutate: func(claims map[string]any) { claims["iss"] = "https://other.invalid" }},
		{name: "audience mismatch", mutate: func(claims map[string]any) { claims["aud"] = "audience:other" }},
		{name: "expired", mutate: func(claims map[string]any) { claims["exp"] = now.Add(-time.Minute).Unix() }},
		{name: "future nbf", mutate: func(claims map[string]any) { claims["nbf"] = now.Add(time.Minute).Unix() }},
		{name: "excess lifetime", mutate: func(claims map[string]any) { claims["exp"] = now.Add(time.Hour).Unix() }},
		{name: "missing subject", mutate: func(claims map[string]any) { delete(claims, "sub") }},
		{name: "permission cardinality", mutate: func(claims map[string]any) { claims["permissions"] = testOIDCTenantPermission }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := cloneOIDCClaims(base)
			if test.mutate != nil {
				test.mutate(claims)
			}
			algorithm := test.algorithm
			if algorithm == "" {
				algorithm = "RS256"
			}
			if _, err := verifier.Validate(context.Background(), signOIDCTestToken(t, key, "key-one", algorithm, claims)); err == nil {
				t.Fatal("expected token validation failure")
			}
		})
	}

	unknown := cloneOIDCClaims(base)
	unknown["permissions"] = []string{"permission:unknown"}
	identity, err = verifier.Validate(context.Background(), signOIDCTestToken(t, key, "key-one", "RS256", unknown))
	if err != nil || len(identity.ScopeGrants) != 0 || len(identity.UpstreamPermissionRefs) != 0 {
		t.Fatalf("unknown permissions must be ignored: identity=%#v err=%v", identity, err)
	}
}

func TestOIDCVerifierUnknownKidRotationOverlapAndHardExpiry(t *testing.T) {
	oldKey := generateSignedTestPrivateKey(t)
	newKey := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"old-key": oldKey})
	cfg := issuer.config()
	cfg.ControlPlaneReadOIDCJWKSMaxAge = time.Minute
	cfg.ControlPlaneReadOIDCRotationOverlap = 2 * time.Minute
	cfg.ControlPlaneReadOIDCJWKSHardExpiry = 5 * time.Minute
	verifier, err := newOIDCTokenVerifier(context.Background(), cfg)
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	verifier.now = func() time.Time { return now }
	claims := validOIDCClaims(issuer.server.URL, now)

	issuer.mu.Lock()
	issuer.keys = map[string]*rsa.PrivateKey{"new-key": newKey}
	issuer.mu.Unlock()
	if _, err := verifier.Validate(context.Background(), signOIDCTestToken(t, newKey, "new-key", "RS256", claims)); err != nil {
		t.Fatalf("unknown kid should refresh once and find rotated key: %v", err)
	}
	issuer.mu.RLock()
	hitsAfterRefresh := issuer.jwksHits
	issuer.mu.RUnlock()
	if hitsAfterRefresh != 2 {
		t.Fatalf("expected startup plus one unknown-kid refresh, got %d", hitsAfterRefresh)
	}
	if _, err := verifier.Validate(context.Background(), signOIDCTestToken(t, oldKey, "old-key", "RS256", claims)); err != nil {
		t.Fatalf("old key must remain valid in overlap: %v", err)
	}
	now = now.Add(3 * time.Minute)
	claims = validOIDCClaims(issuer.server.URL, now)
	if _, err := verifier.Validate(context.Background(), signOIDCTestToken(t, oldKey, "old-key", "RS256", claims)); err == nil {
		t.Fatal("old key must be rejected after overlap")
	}

	issuer.server.Close()
	now = now.Add(6 * time.Minute)
	claims = validOIDCClaims(issuer.server.URL, now)
	_, err = verifier.Validate(context.Background(), signOIDCTestToken(t, newKey, "new-key", "RS256", claims))
	if !errorsIs(err, errOIDCIdentityProviderUnavailable) {
		t.Fatalf("hard-expired cache must surface provider unavailable, got %v", err)
	}
}

func TestOIDCVerifierRefreshIsSingleFlight(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
	verifier, err := newOIDCTokenVerifier(context.Background(), issuer.config())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second).Add(2 * time.Minute)
	verifier.now = func() time.Time { return now }
	issuer.mu.Lock()
	issuer.delay = 50 * time.Millisecond
	issuer.mu.Unlock()
	token := signOIDCTestToken(t, key, "key-one", "RS256", validOIDCClaims(issuer.server.URL, now))
	start := make(chan struct{})
	validationErrors := make(chan error, 16)
	var wait sync.WaitGroup
	for index := 0; index < 16; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, validationErr := verifier.Validate(context.Background(), token)
			validationErrors <- validationErr
		}()
	}
	close(start)
	wait.Wait()
	close(validationErrors)
	for validationErr := range validationErrors {
		if validationErr != nil {
			t.Fatalf("concurrent validation failed: %v", validationErr)
		}
	}
	issuer.mu.RLock()
	hits := issuer.jwksHits
	issuer.mu.RUnlock()
	if hits != 2 {
		t.Fatalf("expected one startup and one single-flight refresh, got %d", hits)
	}
}

func TestOIDCAuthBoundaryWorkspaceGateAndZeroQuery(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
	verifier, err := newOIDCTokenVerifier(context.Background(), issuer.config())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	authenticator := &controlPlaneReadAuthenticator{mode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, oidcVerifier: verifier}
	server := NewServer(config.Config{}, Options{BuildVersion: "test"})
	repository := &recordingControlPlaneReadRepository{}
	server.controlPlaneReadRepo = repository
	server.httpServer.Handler = withControlPlaneReadAuthenticator(server.httpServer.Handler, authenticator)
	claims := validOIDCClaims(issuer.server.URL, time.Now().UTC())
	token := signOIDCTestToken(t, key, "key-one", "RS256", claims)

	for _, test := range []struct {
		name, target string
		status       int
		failure      string
	}{
		{name: "tenant ready", target: "/v1/control-plane/tenants/tenant_demo/summary", status: 200},
		{name: "audit ready", target: "/v1/control-plane/audit", status: 200},
		{name: "applications blocked", target: "/v1/user-workspace/applications", status: 503, failure: "workspace_membership_unavailable"},
		{name: "api keys blocked", target: "/v1/user-workspace/api-keys", status: 503, failure: "workspace_membership_unavailable"},
		{name: "quota blocked", target: "/v1/user-workspace/usage/quota-summary", status: 503, failure: "workspace_membership_unavailable"},
		{name: "workflow definitions blocked", target: "/v1/user-workspace/workflow-definitions", status: 503, failure: "workspace_membership_unavailable"},
		{name: "runs blocked", target: "/v1/user-workspace/runs", status: 503, failure: "workspace_membership_unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := repository.totalCalls
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(recorder, request)
			envelope := decodeControlPlaneReadEnvelope(t, recorder, test.status)
			if test.failure != "" {
				assertControlPlaneReadFailure(t, envelope, test.failure)
				if repository.totalCalls != before {
					t.Fatalf("membership denial reached repository")
				}
			} else if repository.totalCalls != before+1 {
				t.Fatalf("Admin operation did not reach repository exactly once")
			}
			assertControlPlaneReadNoForbiddenPayload(t, recorder.Body.String())
		})
	}

	for _, test := range []struct {
		name, target string
		mutate       func(map[string]any)
		status       int
		failure      string
	}{
		{name: "tenant claim missing", target: "/v1/control-plane/tenants/tenant_demo/summary", mutate: func(c map[string]any) { delete(c, "tenant_id") }, status: 401, failure: "tenant_binding_missing"},
		{name: "tenant mismatch", target: "/v1/control-plane/tenants/tenant_other/summary", status: 403, failure: "tenant_binding_missing"},
		{name: "permission denied", target: "/v1/control-plane/audit", mutate: func(c map[string]any) { c["permissions"] = []string{testOIDCTenantPermission} }, status: 403, failure: "scope_denied"},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := repository.totalCalls
			negativeClaims := cloneOIDCClaims(claims)
			if test.mutate != nil {
				test.mutate(negativeClaims)
			}
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.Header.Set("Authorization", "Bearer "+signOIDCTestToken(t, key, "key-one", "RS256", negativeClaims))
			recorder := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(recorder, request)
			envelope := decodeControlPlaneReadEnvelope(t, recorder, test.status)
			assertControlPlaneReadFailure(t, envelope, test.failure)
			if repository.totalCalls != before {
				t.Fatal("denial reached repository")
			}
		})
	}

	for _, test := range []struct {
		name      string
		configure func(*http.Request)
		failure   string
	}{
		{name: "missing credential", failure: "identity_context_missing"},
		{name: "dev header conflict has no fallback", failure: "auth_context_contract_mismatch", configure: func(request *http.Request) {
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set(controlPlaneReadDevIdentityHeader, "must-not-fallback")
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := repository.totalCalls
			request := httptest.NewRequest(http.MethodGet, "/v1/control-plane/tenants/tenant_demo/summary", nil)
			if test.configure != nil {
				test.configure(request)
			}
			recorder := httptest.NewRecorder()
			server.httpServer.Handler.ServeHTTP(recorder, request)
			envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusUnauthorized)
			assertControlPlaneReadFailure(t, envelope, test.failure)
			if repository.totalCalls != before {
				t.Fatal("authentication failure reached repository")
			}
			if strings.Contains(recorder.Body.String(), token) || strings.Contains(recorder.Body.String(), "must-not-fallback") {
				t.Fatal("authentication diagnostic leaked credential material")
			}
		})
	}

	providerFailureNow := time.Now().UTC().Add(10 * time.Minute)
	verifier.now = func() time.Time { return providerFailureNow }
	issuer.server.Close()
	providerFailureToken := signOIDCTestToken(t, key, "key-one", "RS256", validOIDCClaims(issuer.server.URL, providerFailureNow))
	before := repository.totalCalls
	request := httptest.NewRequest(http.MethodGet, "/v1/control-plane/audit", nil)
	request.Header.Set("Authorization", "Bearer "+providerFailureToken)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusServiceUnavailable)
	assertControlPlaneReadFailure(t, envelope, "identity_provider_unavailable")
	if repository.totalCalls != before || strings.Contains(recorder.Body.String(), issuer.server.URL) {
		t.Fatal("identity provider failure reached repository or leaked provider detail")
	}
}

func TestOIDCMiddlewareStripsCredentialBeforeDownstream(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
	verifier, err := newOIDCTokenVerifier(context.Background(), issuer.config())
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	authenticator := &controlPlaneReadAuthenticator{mode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, oidcVerifier: verifier}
	called := false
	handler := withControlPlaneReadAuthenticator(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		called = true
		if request.Header.Get("Authorization") != "" || controlPlaneReadHasAnyDevHeader(request) {
			t.Fatal("credential material reached downstream handler")
		}
		auth, ok := request.Context().Value(controlPlaneReadAuthContextKey{}).(controlPlaneReadAuthContext)
		if !ok || auth.VerifiedIdentity == nil || auth.VerifiedIdentity.SubjectRef != "subject:test-admin" {
			t.Fatalf("sanitized identity was not propagated: %#v", auth)
		}
	}), authenticator)
	request := httptest.NewRequest(http.MethodGet, "/v1/control-plane/audit", nil)
	request.Header.Set("Authorization", "Bearer "+signOIDCTestToken(t, key, "key-one", "RS256", validOIDCClaims(issuer.server.URL, time.Now().UTC())))
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if !called {
		t.Fatal("downstream handler was not called")
	}
}

func validOIDCClaims(issuer string, now time.Time) map[string]any {
	return map[string]any{
		"iss": issuer, "sub": "subject:test-admin", "tenant_id": "tenant_demo", "aud": []string{testOIDCAudience},
		"permissions": []string{testOIDCTenantPermission, testOIDCAuditPermission, "permission:unknown"},
		"iat":         now.Add(-time.Minute).Unix(), "nbf": now.Add(-time.Minute).Unix(), "exp": now.Add(3 * time.Minute).Unix(),
		"auth_time": now.Add(-time.Minute).Unix(),
	}
}

func cloneOIDCClaims(source map[string]any) map[string]any {
	encoded, _ := json.Marshal(source)
	result := map[string]any{}
	_ = json.Unmarshal(encoded, &result)
	return result
}

func signOIDCTestToken(t *testing.T, key *rsa.PrivateKey, keyID string, algorithm string, claims map[string]any) string {
	t.Helper()
	headerBytes, _ := json.Marshal(map[string]string{"alg": algorithm, "typ": "JWT", "kid": keyID})
	claimBytes, _ := json.Marshal(claims)
	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(claimBytes)
	digest := sha256.Sum256([]byte(header + "." + payload))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign OIDC token: %v", err)
	}
	return header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func marshalOIDCJWKS(keys map[string]*rsa.PrivateKey) []byte {
	documents := make([]map[string]string, 0, len(keys))
	for keyID, key := range keys {
		documents = append(documents, map[string]string{
			"kty": "RSA", "use": "sig", "kid": keyID, "alg": "RS256",
			"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
		})
	}
	body, _ := json.Marshal(map[string]any{"keys": documents})
	return body
}

func errorsIs(err error, target error) bool {
	return err != nil && (err == target || strings.Contains(err.Error(), target.Error()))
}

func TestOIDCDiagnosticsDoNotContainProviderMaterial(t *testing.T) {
	key := generateSignedTestPrivateKey(t)
	issuer := newDeterministicOIDCIssuer(t, map[string]*rsa.PrivateKey{"key-one": key})
	issuer.discoveryBody = []byte(fmt.Sprintf(`{"issuer":%q,"jwks_uri":"https://secret.invalid/jwks?token=must-not-leak","id_token_signing_alg_values_supported":["RS256"]}`, issuer.server.URL))
	_, err := newOIDCTokenVerifier(context.Background(), issuer.config())
	if err == nil || strings.Contains(err.Error(), "secret.invalid") || strings.Contains(err.Error(), "must-not-leak") {
		t.Fatalf("diagnostic leaked provider material: %v", err)
	}
}
