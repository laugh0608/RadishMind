package httpapi

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	controlPlaneReadAuthModeDisabled                  = "disabled"
	controlPlaneReadAuthModeDevHeaders                = "dev_headers"
	controlPlaneReadAuthModeSignedTestToken           = "signed_test_token"
	controlPlaneReadAuthModeRadishOIDCIntegrationTest = "radish_oidc_integration_test"
	controlPlaneReadAuthPolicyVersion                 = "control_plane_read_auth_policy_v1"
)

var controlPlaneReadPermissionGrants = map[string]string{
	"radishmind.tenant.read":                    "tenant:read",
	"radishmind.applications.read":              "applications:read",
	"radishmind.applications.write":             "applications:write",
	"radishmind.applications.archive":           "applications:archive",
	"radishmind.api-keys.read":                  "api_keys:read",
	"radishmind.api-keys.write":                 "api_keys:write",
	"radishmind.api-keys.revoke":                "api_keys:revoke",
	"radishmind.usage.read":                     "usage:read",
	"radishmind.runs.read":                      "runs:read",
	"radishmind.audit.read":                     "audit:read",
	"radishmind.workflow-rag-snapshots.read":    "workflow_rag_snapshots:read",
	"radishmind.workflow-rag-snapshots.write":   "workflow_rag_snapshots:write",
	"radishmind.workflow-rag-snapshots.archive": "workflow_rag_snapshots:archive",
	"radishmind.workflow-rag-promotions.read":   "workflow_rag_promotions:read",
	"radishmind.workflow-rag-promotions.write":  "workflow_rag_promotions:write",
	"radishmind.workflow-rag-promotions.review": "workflow_rag_promotions:review",
	"radishmind.workflow-rag-promotions.bind":   "workflow_rag_promotions:bind",
	"radishmind.workflow-rag-runtime.read":      "workflow_rag_runtime:read",
	"radishmind.workflow-rag-runtime.write":     "workflow_rag_runtime:write",
	"radishmind.workflow-rag.execute":           "workflow_rag:execute",
	"radishmind.application-sessions.read":      "application_sessions:read",
	"radishmind.application-sessions.write":     "application_sessions:write",
	"radishmind.application-sessions.execute":   "application_sessions:execute",
}

var promptApplicationTemplatePermissionGrants = map[string]string{
	"radishmind.prompt-application-templates.read":        "prompt_application_templates:read",
	"radishmind.prompt-application-templates.read-source": "prompt_application_templates:read_source",
	"radishmind.prompt-application-templates.write":       "prompt_application_templates:write",
	"radishmind.prompt-application-templates.version":     "prompt_application_templates:version",
	"radishmind.prompt-application-templates.bind":        "prompt_application_templates:bind",
	"radishmind.prompt-application-runtime.read":          "prompt_application_runtime:read",
	"radishmind.prompt-application-runtime.write":         "prompt_application_runtime:write",
}

var controlPlaneReadAuthReferencePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]{0,159}$`)

type VerifiedControlPlaneIdentity struct {
	AuthSource             string
	IssuerRef              string
	SubjectRef             string
	TenantRef              string
	AudienceRefs           []string
	UpstreamPermissionRefs []string
	ScopeGrants            []string
	IssuedAt               time.Time
	ExpiresAt              time.Time
	AuthTime               time.Time
	PolicyVersion          string
	SessionRef             string
	KeyIDRef               string
	Algorithm              string
}

type ControlPlaneResourceBinding struct {
	TenantRef        string
	TenantVerified   bool
	PermissionGrants []string
	SourceRef        string
	PolicyVersion    string
	ExpiresAt        time.Time
}

type controlPlaneReadAuthContext struct {
	AuthMode         string
	IdentityContext  string
	TenantBinding    string
	SubjectBinding   string
	ScopeGrants      []string
	AuditContext     string
	IssuerRef        string
	SessionRef       string
	VerifiedIdentity *VerifiedControlPlaneIdentity
	ResourceBinding  ControlPlaneResourceBinding
	FailureCode      string
	FailureStatus    int
}

type controlPlaneReadAuthenticator struct {
	mode            string
	signedValidator *signedTestTokenValidator
	oidcVerifier    *oidcTokenVerifier
}

type controlPlaneReadAuthContextKey struct{}

type signedTestTokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

type signedTestTokenClaims struct {
	Issuer      string          `json:"iss"`
	Subject     string          `json:"sub"`
	TenantRef   string          `json:"tenant_id"`
	Audience    json.RawMessage `json:"aud"`
	Permissions []string        `json:"permissions"`
	IssuedAt    int64           `json:"iat"`
	NotBefore   int64           `json:"nbf"`
	ExpiresAt   int64           `json:"exp"`
	AuthTime    int64           `json:"auth_time"`
	SessionRef  string          `json:"sid"`
}

type signedTestTokenValidator struct {
	issuer    string
	audience  string
	publicKey *rsa.PublicKey
	now       func() time.Time
}

func withControlPlaneReadFakeAuthContext(ctx context.Context, auth controlPlaneReadAuthContext) context.Context {
	return context.WithValue(ctx, controlPlaneReadAuthContextKey{}, auth)
}

func withControlPlaneReadAuth(next http.Handler, cfg config.Config) http.Handler {
	authenticator, err := newControlPlaneReadAuthenticator(context.Background(), cfg)
	if err != nil {
		authenticator = &controlPlaneReadAuthenticator{mode: config.EffectiveControlPlaneReadAuthMode(cfg)}
	}
	return withControlPlaneReadAuthenticator(next, authenticator)
}

func newControlPlaneReadAuthenticator(ctx context.Context, cfg config.Config) (*controlPlaneReadAuthenticator, error) {
	mode := config.EffectiveControlPlaneReadAuthMode(cfg)
	authenticator := &controlPlaneReadAuthenticator{mode: mode}
	if mode == controlPlaneReadAuthModeSignedTestToken {
		if publicKey, err := parseSignedTestRSAPublicKey(cfg.ControlPlaneReadTestPublicKeyPEM); err == nil {
			authenticator.signedValidator = &signedTestTokenValidator{
				issuer:    strings.TrimSpace(cfg.ControlPlaneReadTestIssuer),
				audience:  strings.TrimSpace(cfg.ControlPlaneReadTestAudience),
				publicKey: publicKey,
				now:       time.Now,
			}
		}
	}
	if mode == controlPlaneReadAuthModeRadishOIDCIntegrationTest {
		if err := config.ValidateControlPlaneReadOIDCIntegration(cfg); err != nil {
			return nil, errors.New("control plane read OIDC auth policy mismatch")
		}
		verifier, err := newOIDCTokenVerifier(ctx, cfg)
		if err != nil {
			return nil, err
		}
		authenticator.oidcVerifier = verifier
	}
	return authenticator, nil
}

func withControlPlaneReadAuthenticator(next http.Handler, authenticator *controlPlaneReadAuthenticator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if authenticator == nil || isGatewayNorthboundRequest(request) {
			next.ServeHTTP(writer, request)
			return
		}
		switch authenticator.mode {
		case controlPlaneReadAuthModeDevHeaders:
			if auth, ok := controlPlaneReadDevAuthFromHeaders(request); ok {
				request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			}
		case controlPlaneReadAuthModeSignedTestToken:
			auth := controlPlaneReadSignedTestAuthFromRequest(request, authenticator.signedValidator)
			request = sanitizedControlPlaneReadAuthRequest(request, auth)
		case controlPlaneReadAuthModeRadishOIDCIntegrationTest:
			auth := controlPlaneReadOIDCAuthFromRequest(request, authenticator.oidcVerifier)
			request = sanitizedControlPlaneReadAuthRequest(request, auth)
		case controlPlaneReadAuthModeDisabled:
		}
		next.ServeHTTP(writer, request)
	})
}

func sanitizedControlPlaneReadAuthRequest(request *http.Request, auth controlPlaneReadAuthContext) *http.Request {
	cloned := request.Clone(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	cloned.Header.Del("Authorization")
	for _, header := range []string{controlPlaneReadDevIdentityHeader, controlPlaneReadDevTenantHeader, controlPlaneReadDevSubjectHeader, controlPlaneReadDevScopesHeader, controlPlaneReadDevAuditHeader} {
		cloned.Header.Del(header)
	}
	return cloned
}

func controlPlaneReadOIDCAuthFromRequest(request *http.Request, verifier *oidcTokenVerifier) controlPlaneReadAuthContext {
	if controlPlaneReadHasAnyDevHeader(request) {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "auth_context_contract_mismatch"}
	}
	values := request.Header.Values("Authorization")
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "identity_context_missing"}
	}
	parts := strings.Fields(values[0])
	if len(values) != 1 || len(parts) != 2 || parts[0] != "Bearer" || verifier == nil {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "auth_context_contract_mismatch"}
	}
	identity, err := verifier.Validate(request.Context(), parts[1])
	if errors.Is(err, errOIDCIdentityProviderUnavailable) {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "identity_provider_unavailable"}
	}
	if errors.Is(err, errOIDCTenantBindingMissing) {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "tenant_binding_missing", FailureStatus: http.StatusUnauthorized}
	}
	if err != nil {
		return controlPlaneReadAuthContext{AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, FailureCode: "auth_context_contract_mismatch"}
	}
	binding := ControlPlaneResourceBinding{
		TenantRef: identity.TenantRef, TenantVerified: strings.TrimSpace(identity.TenantRef) != "",
		PermissionGrants: append([]string{}, identity.ScopeGrants...), SourceRef: "binding:radish-oidc-integration-test",
		PolicyVersion: identity.PolicyVersion, ExpiresAt: identity.ExpiresAt,
	}
	return controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeRadishOIDCIntegrationTest, IdentityContext: "verified:radish-oidc-integration-test",
		TenantBinding: identity.TenantRef, SubjectBinding: identity.SubjectRef, ScopeGrants: append([]string{}, identity.ScopeGrants...),
		AuditContext: "audit:radish-oidc-integration-test", IssuerRef: identity.IssuerRef, SessionRef: identity.SessionRef,
		VerifiedIdentity: &identity, ResourceBinding: binding,
	}
}

func controlPlaneReadDevAuthFromHeaders(request *http.Request) (controlPlaneReadAuthContext, bool) {
	identity := strings.TrimSpace(request.Header.Get(controlPlaneReadDevIdentityHeader))
	tenantRef := strings.TrimSpace(request.Header.Get(controlPlaneReadDevTenantHeader))
	subjectRef := strings.TrimSpace(request.Header.Get(controlPlaneReadDevSubjectHeader))
	scopeHeader := strings.TrimSpace(request.Header.Get(controlPlaneReadDevScopesHeader))
	if identity == "" || tenantRef == "" || subjectRef == "" || scopeHeader == "" {
		return controlPlaneReadAuthContext{}, false
	}
	scopes := splitControlPlaneReadDevScopes(scopeHeader)
	identityProjection := &VerifiedControlPlaneIdentity{
		AuthSource:    controlPlaneReadAuthModeDevHeaders,
		IssuerRef:     "issuer:dev-headers",
		SubjectRef:    subjectRef,
		TenantRef:     tenantRef,
		ScopeGrants:   scopes,
		PolicyVersion: controlPlaneReadAuthPolicyVersion,
		SessionRef:    "session:dev-read",
	}
	return controlPlaneReadAuthContext{
		AuthMode:         controlPlaneReadAuthModeDevHeaders,
		IdentityContext:  identity,
		TenantBinding:    tenantRef,
		SubjectBinding:   subjectRef,
		ScopeGrants:      scopes,
		AuditContext:     strings.TrimSpace(request.Header.Get(controlPlaneReadDevAuditHeader)),
		IssuerRef:        identityProjection.IssuerRef,
		SessionRef:       identityProjection.SessionRef,
		VerifiedIdentity: identityProjection,
		ResourceBinding: ControlPlaneResourceBinding{
			TenantRef:        tenantRef,
			TenantVerified:   true,
			PermissionGrants: append([]string{}, scopes...),
			SourceRef:        "binding:dev-headers",
			PolicyVersion:    controlPlaneReadAuthPolicyVersion,
		},
	}, true
}

func splitControlPlaneReadDevScopes(rawScopes string) []string {
	scopes := make([]string, 0)
	for _, scope := range strings.Split(rawScopes, ",") {
		normalized := strings.TrimSpace(scope)
		if normalized != "" {
			scopes = append(scopes, normalized)
		}
	}
	return scopes
}

func controlPlaneReadSignedTestAuthFromRequest(request *http.Request, validator *signedTestTokenValidator) controlPlaneReadAuthContext {
	if controlPlaneReadHasAnyDevHeader(request) {
		return controlPlaneReadAuthContext{FailureCode: "auth_context_contract_mismatch"}
	}
	authorization := strings.TrimSpace(request.Header.Get("Authorization"))
	if authorization == "" {
		return controlPlaneReadAuthContext{FailureCode: "identity_context_missing"}
	}
	parts := strings.Fields(authorization)
	if len(parts) != 2 || parts[0] != "Bearer" || validator == nil {
		return controlPlaneReadAuthContext{FailureCode: "auth_context_contract_mismatch"}
	}
	identity, err := validator.Validate(parts[1])
	if err != nil {
		return controlPlaneReadAuthContext{FailureCode: "auth_context_contract_mismatch"}
	}
	binding := ControlPlaneResourceBinding{
		TenantRef:        identity.TenantRef,
		TenantVerified:   strings.TrimSpace(identity.TenantRef) != "",
		PermissionGrants: append([]string{}, identity.ScopeGrants...),
		SourceRef:        "binding:signed-test-token",
		PolicyVersion:    identity.PolicyVersion,
		ExpiresAt:        identity.ExpiresAt,
	}
	return controlPlaneReadAuthContext{
		AuthMode:         controlPlaneReadAuthModeSignedTestToken,
		IdentityContext:  "verified:signed-test-token",
		TenantBinding:    identity.TenantRef,
		SubjectBinding:   identity.SubjectRef,
		ScopeGrants:      append([]string{}, identity.ScopeGrants...),
		AuditContext:     "audit:signed-test-token",
		IssuerRef:        identity.IssuerRef,
		SessionRef:       identity.SessionRef,
		VerifiedIdentity: &identity,
		ResourceBinding:  binding,
	}
}

func (validator *signedTestTokenValidator) Validate(rawToken string) (VerifiedControlPlaneIdentity, error) {
	segments := strings.Split(rawToken, ".")
	if len(segments) != 3 || validator.publicKey == nil || validator.issuer == "" || validator.audience == "" {
		return VerifiedControlPlaneIdentity{}, errors.New("invalid signed test token contract")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	var header signedTestTokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "RS256" || header.Type != "JWT" || strings.TrimSpace(header.KeyID) == "" {
		return VerifiedControlPlaneIdentity{}, errors.New("invalid signed test token header")
	}
	signature, err := base64.RawURLEncoding.DecodeString(segments[2])
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	digest := sha256.Sum256([]byte(segments[0] + "." + segments[1]))
	if err := rsa.VerifyPKCS1v15(validator.publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return VerifiedControlPlaneIdentity{}, errors.New("invalid signed test token signature")
	}
	claimsBytes, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return VerifiedControlPlaneIdentity{}, err
	}
	var claims signedTestTokenClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return VerifiedControlPlaneIdentity{}, errors.New("invalid signed test token claims")
	}
	audiences, err := signedTestTokenAudiences(claims.Audience)
	if err != nil || !validControlPlaneReadAuthReference(claims.Subject, false) || !validControlPlaneReadAuthReference(claims.TenantRef, true) || !validControlPlaneReadAuthReference(claims.SessionRef, false) || !validControlPlaneReadPermissions(claims.Permissions) || !validControlPlaneReadAuthReference(header.KeyID, false) || claims.IssuedAt == 0 || claims.NotBefore == 0 || claims.ExpiresAt == 0 || claims.AuthTime == 0 {
		return VerifiedControlPlaneIdentity{}, errors.New("missing signed test token claims")
	}
	if claims.Issuer != validator.issuer || !containsExactString(audiences, validator.audience) {
		return VerifiedControlPlaneIdentity{}, errors.New("signed test token issuer or audience mismatch")
	}
	now := validator.now().UTC()
	if claims.ExpiresAt <= claims.IssuedAt || claims.NotBefore >= claims.ExpiresAt || claims.AuthTime > claims.IssuedAt || !time.Unix(claims.ExpiresAt, 0).After(now) || time.Unix(claims.NotBefore, 0).After(now) || time.Unix(claims.IssuedAt, 0).After(now) || time.Unix(claims.AuthTime, 0).After(now) {
		return VerifiedControlPlaneIdentity{}, errors.New("signed test token time validation failed")
	}
	grants := projectControlPlaneReadPermissions(claims.Permissions)
	return VerifiedControlPlaneIdentity{
		AuthSource:             controlPlaneReadAuthModeSignedTestToken,
		IssuerRef:              "issuer:signed-test",
		SubjectRef:             strings.TrimSpace(claims.Subject),
		TenantRef:              strings.TrimSpace(claims.TenantRef),
		AudienceRefs:           append([]string{}, audiences...),
		UpstreamPermissionRefs: append([]string{}, claims.Permissions...),
		ScopeGrants:            grants,
		IssuedAt:               time.Unix(claims.IssuedAt, 0).UTC(),
		ExpiresAt:              time.Unix(claims.ExpiresAt, 0).UTC(),
		AuthTime:               time.Unix(claims.AuthTime, 0).UTC(),
		PolicyVersion:          controlPlaneReadAuthPolicyVersion,
		SessionRef:             strings.TrimSpace(claims.SessionRef),
		KeyIDRef:               "key:" + strings.TrimSpace(header.KeyID),
		Algorithm:              header.Algorithm,
	}, nil
}

func parseSignedTestRSAPublicKey(rawPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(strings.TrimSpace(rawPEM)))
	if block == nil {
		return nil, errors.New("invalid RSA public key PEM")
	}
	if parsed, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if publicKey, ok := parsed.(*rsa.PublicKey); ok {
			return publicKey, nil
		}
	}
	if publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return publicKey, nil
	}
	return nil, errors.New("PEM does not contain an RSA public key")
}

func signedTestTokenAudiences(raw json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(raw, &single); err == nil && strings.TrimSpace(single) != "" {
		return []string{strings.TrimSpace(single)}, nil
	}
	var multiple []string
	if err := json.Unmarshal(raw, &multiple); err != nil || len(multiple) == 0 {
		return nil, errors.New("invalid audience")
	}
	for index := range multiple {
		multiple[index] = strings.TrimSpace(multiple[index])
		if multiple[index] == "" {
			return nil, errors.New("invalid audience")
		}
	}
	return multiple, nil
}

func projectControlPlaneReadPermissions(permissions []string) []string {
	grants := make([]string, 0, len(permissions))
	seen := map[string]bool{}
	for _, permission := range permissions {
		grant, ok := controlPlaneReadPermissionGrants[strings.TrimSpace(permission)]
		if !ok {
			grant, ok = promptApplicationTemplatePermissionGrants[strings.TrimSpace(permission)]
		}
		if ok && !seen[grant] {
			seen[grant] = true
			grants = append(grants, grant)
		}
	}
	return grants
}

func controlPlaneReadHasAnyDevHeader(request *http.Request) bool {
	for _, header := range []string{
		controlPlaneReadDevIdentityHeader,
		controlPlaneReadDevTenantHeader,
		controlPlaneReadDevSubjectHeader,
		controlPlaneReadDevScopesHeader,
		controlPlaneReadDevAuditHeader,
	} {
		if strings.TrimSpace(request.Header.Get(header)) != "" {
			return true
		}
	}
	return false
}

func containsExactString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func validControlPlaneReadAuthReference(value string, allowEmpty bool) bool {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return allowEmpty
	}
	return controlPlaneReadAuthReferencePattern.MatchString(normalized)
}

func validControlPlaneReadPermissions(permissions []string) bool {
	if len(permissions) == 0 {
		return false
	}
	for _, permission := range permissions {
		normalized := strings.TrimSpace(permission)
		if len(normalized) > 120 || !controlPlaneReadAuthReferencePattern.MatchString(normalized) {
			return false
		}
	}
	return true
}
