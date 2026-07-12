package httpapi

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	signedTestIssuer   = "https://radish.test/oidc"
	signedTestAudience = "radishmind-control-plane"
)

func TestControlPlaneReadSignedTestTokenAdminRoutes(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	server := newSignedTestControlPlaneReadServer(t, privateKey)
	repository := &recordingControlPlaneReadRepository{}
	server.controlPlaneReadRepo = repository

	tenantRequest := httptest.NewRequest(http.MethodGet, "/v1/control-plane/tenants/tenant_demo/summary", nil)
	tenantRequest.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", validSignedTestClaims()))
	tenantRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(tenantRecorder, tenantRequest)

	tenantEnvelope := decodeControlPlaneReadEnvelope(t, tenantRecorder, http.StatusOK)
	if tenantEnvelope.FailureCode != nil || len(tenantEnvelope.Items) != 1 {
		t.Fatalf("unexpected signed-token tenant response: %#v", tenantEnvelope)
	}
	if repository.totalCalls != 1 || repository.lastContext.IssuerRef != "issuer:signed-test" || repository.lastContext.SessionRef != "session:test-admin" {
		t.Fatalf("verified identity was not projected into repository context: calls=%d context=%#v", repository.totalCalls, repository.lastContext)
	}

	auditRequest := httptest.NewRequest(http.MethodGet, "/v1/control-plane/audit", nil)
	claims := validSignedTestClaims()
	claims["permissions"] = []string{"radishmind.audit.read"}
	auditRequest.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
	auditRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(auditRecorder, auditRequest)

	auditEnvelope := decodeControlPlaneReadEnvelope(t, auditRecorder, http.StatusOK)
	if auditEnvelope.FailureCode != nil {
		t.Fatalf("unexpected signed-token audit response: %#v", auditEnvelope)
	}
	if repository.totalCalls != 2 {
		t.Fatalf("expected both Admin reads to reach repository, got %d calls", repository.totalCalls)
	}
}

func TestControlPlaneReadSignedTestTokenCoversAllRouteGrants(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	server := newSignedTestControlPlaneReadServer(t, privateKey)
	for _, test := range []struct {
		name       string
		target     string
		permission string
	}{
		{name: "tenant", target: "/v1/control-plane/tenants/tenant_demo/summary", permission: "radishmind.tenant.read"},
		{name: "applications", target: "/v1/user-workspace/applications", permission: "radishmind.applications.read"},
		{name: "api keys", target: "/v1/user-workspace/api-keys", permission: "radishmind.api-keys.read"},
		{name: "quota", target: "/v1/user-workspace/usage/quota-summary", permission: "radishmind.usage.read"},
		{name: "workflow definitions", target: "/v1/user-workspace/workflow-definitions", permission: "radishmind.applications.read"},
		{name: "runs", target: "/v1/user-workspace/runs", permission: "radishmind.runs.read"},
		{name: "audit", target: "/v1/control-plane/audit", permission: "radishmind.audit.read"},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := validSignedTestClaims()
			claims["permissions"] = []string{test.permission}
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
			recorder := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(recorder, request)

			envelope := decodeControlPlaneReadEnvelope(t, recorder, http.StatusOK)
			if envelope.FailureCode != nil {
				t.Fatalf("route grant was denied: %#v", envelope)
			}
		})
	}
}

func TestControlPlaneReadSignedTestTokenNegativeAuthMatrix(t *testing.T) {
	privateKey := generateSignedTestPrivateKey(t)
	otherPrivateKey := generateSignedTestPrivateKey(t)

	tests := []struct {
		name            string
		target          string
		authorization   func() string
		addDevHeader    bool
		expectedCode    int
		expectedFailure string
	}{
		{name: "missing credential", target: "/v1/control-plane/tenants/tenant_demo/summary", expectedCode: http.StatusUnauthorized, expectedFailure: "identity_context_missing"},
		{name: "malformed scheme", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: func() string { return "Basic malformed" }, expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "disallowed algorithm", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "HS256", validSignedTestClaims()), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "invalid signature", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, otherPrivateKey, "RS256", validSignedTestClaims()), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "issuer mismatch", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("iss", "https://other.test/oidc")), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "audience mismatch", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("aud", "other-audience")), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "expired", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("exp", time.Now().Add(-time.Minute).Unix())), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "not yet valid", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("nbf", time.Now().Add(time.Minute).Unix())), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "required claim invalid", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("sub", "")), expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
		{name: "tenant binding missing", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("tenant_id", "")), expectedCode: http.StatusUnauthorized, expectedFailure: "tenant_binding_missing"},
		{name: "path tenant mismatch", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("tenant_id", "tenant_other")), expectedCode: http.StatusForbidden, expectedFailure: "tenant_binding_missing"},
		{name: "permission mapping denied", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", mutateSignedTestClaims("permissions", []string{"radishmind.unknown.read"})), expectedCode: http.StatusForbidden, expectedFailure: "scope_denied"},
		{name: "dev header credential conflict", target: "/v1/control-plane/tenants/tenant_demo/summary", authorization: signedTestAuthorization(t, privateKey, "RS256", validSignedTestClaims()), addDevHeader: true, expectedCode: http.StatusUnauthorized, expectedFailure: "auth_context_contract_mismatch"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newSignedTestControlPlaneReadServer(t, privateKey)
			repository := &recordingControlPlaneReadRepository{}
			server.controlPlaneReadRepo = repository
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			if test.authorization != nil {
				request.Header.Set("Authorization", test.authorization())
			}
			if test.addDevHeader {
				request.Header.Set(controlPlaneReadDevIdentityHeader, "must-not-be-consumed")
			}
			recorder := httptest.NewRecorder()

			server.httpServer.Handler.ServeHTTP(recorder, request)

			envelope := decodeControlPlaneReadEnvelope(t, recorder, test.expectedCode)
			assertControlPlaneReadFailure(t, envelope, test.expectedFailure)
			if repository.totalCalls != 0 {
				t.Fatalf("auth denial reached repository %d times", repository.totalCalls)
			}
			assertControlPlaneReadNoForbiddenPayload(t, recorder.Body.String())
		})
	}
}

func newSignedTestControlPlaneReadServer(t *testing.T, privateKey *rsa.PrivateKey) *Server {
	t.Helper()
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal signed test public key: %v", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})
	return NewServer(config.Config{
		ControlPlaneReadAuthMode:         controlPlaneReadAuthModeSignedTestToken,
		ControlPlaneReadTestIssuer:       signedTestIssuer,
		ControlPlaneReadTestAudience:     signedTestAudience,
		ControlPlaneReadTestPublicKeyPEM: string(publicKeyPEM),
	}, Options{BuildVersion: "test"})
}

func generateSignedTestPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate signed test RSA key: %v", err)
	}
	return privateKey
}

func validSignedTestClaims() map[string]any {
	now := time.Now().UTC()
	return map[string]any{
		"iss":         signedTestIssuer,
		"sub":         "subject:test-admin",
		"tenant_id":   "tenant_demo",
		"aud":         signedTestAudience,
		"permissions": []string{"radishmind.tenant.read"},
		"iat":         now.Add(-time.Minute).Unix(),
		"nbf":         now.Add(-time.Minute).Unix(),
		"exp":         now.Add(time.Hour).Unix(),
		"auth_time":   now.Add(-time.Minute).Unix(),
		"sid":         "session:test-admin",
	}
}

func mutateSignedTestClaims(key string, value any) map[string]any {
	claims := validSignedTestClaims()
	claims[key] = value
	return claims
}

func signedTestAuthorization(t *testing.T, privateKey *rsa.PrivateKey, algorithm string, claims map[string]any) func() string {
	t.Helper()
	return func() string {
		return "Bearer " + signControlPlaneReadTestToken(t, privateKey, algorithm, claims)
	}
}

func signControlPlaneReadTestToken(t *testing.T, privateKey *rsa.PrivateKey, algorithm string, claims map[string]any) string {
	t.Helper()
	headerBytes, err := json.Marshal(map[string]string{"alg": algorithm, "typ": "JWT", "kid": "signed-test-key"})
	if err != nil {
		t.Fatalf("marshal signed test header: %v", err)
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal signed test claims: %v", err)
	}
	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)
	digest := sha256.Sum256([]byte(header + "." + payload))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(signature)
}
