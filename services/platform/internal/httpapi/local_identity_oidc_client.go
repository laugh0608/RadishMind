package httpapi

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"radishmind.local/services/platform/internal/config"
)

var (
	errLocalIdentityOIDCProviderUnavailable = errors.New("OIDC provider unavailable")
	errLocalIdentityOIDCCallbackMismatch    = errors.New("OIDC callback contract mismatch")
	errLocalIdentityOIDCTokenExchange       = errors.New("OIDC token exchange failed")
	errLocalIdentityOIDCIdentityMismatch    = errors.New("OIDC identity contract mismatch")
)

type localIdentityOIDCClientPolicy struct {
	issuer            string
	discoveryURL      string
	clientID          string
	redirectURI       string
	scopes            []string
	algorithms        map[string]bool
	jwksOrigin        string
	transactionTTL    time.Duration
	firstLoginEnabled bool
}

type localIdentityOIDCDiscoveryDocument struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	JWKSURI               string   `json:"jwks_uri"`
	SigningAlgorithms     []string `json:"id_token_signing_alg_values_supported"`
	ResponseTypes         []string `json:"response_types_supported"`
	CodeChallengeMethods  []string `json:"code_challenge_methods_supported"`
}

type localIdentityOIDCClient struct {
	policy                localIdentityOIDCClientPolicy
	authorizationEndpoint string
	tokenEndpoint         string
	verifier              *oidcTokenVerifier
	httpClient            *http.Client
}

type localIdentityOIDCVerifiedIdentity struct {
	Issuer  string
	Subject string
}

type localIdentityOIDCTokenResponse struct {
	IDToken string `json:"id_token"`
}

func newLocalIdentityOIDCClient(ctx context.Context, cfg config.Config) (*localIdentityOIDCClient, error) {
	if !cfg.LocalIdentityOIDCEnabled {
		return nil, nil
	}
	policy, err := localIdentityOIDCClientPolicyFromConfig(cfg)
	if err != nil {
		return nil, errLocalIdentityOIDCCallbackMismatch
	}
	verifierPolicy := oidcVerifierPolicy{
		issuer: policy.issuer, discoveryURL: policy.discoveryURL, audience: policy.clientID,
		mappingVersion: localIdentityOIDCPolicyVersion, evidenceRef: "issuer:local-oidc-client", subjectClaim: "sub",
		algorithms: policy.algorithms, jwksOrigin: policy.jwksOrigin,
		discoveryTimeout: 3 * time.Second, jwksMaxAge: 5 * time.Minute, jwksHardExpiry: 15 * time.Minute,
		rotationOverlap: 2 * time.Minute, clockSkew: 30 * time.Second, maxTokenLifetime: 15 * time.Minute,
		maxResponseBytes: 64 * 1024, maxKeys: 32,
	}
	httpClient := &http.Client{
		Timeout: verifierPolicy.discoveryTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("OIDC redirects are disabled")
		},
	}
	verifier := &oidcTokenVerifier{policy: verifierPolicy, httpClient: httpClient, now: time.Now}
	var discovery localIdentityOIDCDiscoveryDocument
	if err := verifier.fetchJSON(ctx, policy.discoveryURL, &discovery); err != nil {
		return nil, errLocalIdentityOIDCProviderUnavailable
	}
	if !validLocalIdentityOIDCDiscovery(policy, discovery) {
		return nil, errLocalIdentityOIDCCallbackMismatch
	}
	verifier.jwksURI = discovery.JWKSURI
	if err := verifier.refreshKeys(ctx); err != nil {
		return nil, errLocalIdentityOIDCProviderUnavailable
	}
	return &localIdentityOIDCClient{
		policy: policy, authorizationEndpoint: discovery.AuthorizationEndpoint,
		tokenEndpoint: discovery.TokenEndpoint, verifier: verifier, httpClient: httpClient,
	}, nil
}

func localIdentityOIDCClientPolicyFromConfig(cfg config.Config) (localIdentityOIDCClientPolicy, error) {
	issuer, err := NormalizeExternalIssuer(cfg.LocalIdentityOIDCIssuer)
	if err != nil {
		return localIdentityOIDCClientPolicy{}, err
	}
	scopes := splitTrimmedUnique(cfg.LocalIdentityOIDCScopes)
	algorithms := map[string]bool{}
	for _, algorithm := range splitTrimmedUnique(cfg.LocalIdentityOIDCAlgorithms) {
		algorithms[algorithm] = true
	}
	if len(scopes) == 0 || len(algorithms) == 0 || !slices.Contains(scopes, "openid") {
		return localIdentityOIDCClientPolicy{}, errLocalIdentityOIDCCallbackMismatch
	}
	return localIdentityOIDCClientPolicy{
		issuer: issuer, discoveryURL: strings.TrimSpace(cfg.LocalIdentityOIDCDiscoveryURL),
		clientID: strings.TrimSpace(cfg.LocalIdentityOIDCClientID), redirectURI: strings.TrimSpace(cfg.LocalIdentityOIDCRedirectURI),
		scopes: scopes, algorithms: algorithms, jwksOrigin: strings.TrimSuffix(strings.TrimSpace(cfg.LocalIdentityOIDCJWKSOrigin), "/"),
		transactionTTL: cfg.LocalIdentityOIDCTransactionTTL, firstLoginEnabled: cfg.LocalIdentityOIDCFirstLoginEnabled,
	}, nil
}

func splitTrimmedUnique(raw string) []string {
	values := strings.Split(raw, ",")
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, rawValue := range values {
		value := strings.TrimSpace(rawValue)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func validLocalIdentityOIDCDiscovery(policy localIdentityOIDCClientPolicy, discovery localIdentityOIDCDiscoveryDocument) bool {
	issuer, err := NormalizeExternalIssuer(discovery.Issuer)
	if err != nil || issuer != policy.issuer || !sameOIDCOrigin(discovery.JWKSURI, policy.jwksOrigin) ||
		!sameOIDCOrigin(discovery.AuthorizationEndpoint, policy.jwksOrigin) ||
		!sameOIDCOrigin(discovery.TokenEndpoint, policy.jwksOrigin) ||
		!containsAllowedOIDCAlgorithm(discovery.SigningAlgorithms, policy.algorithms) ||
		!containsExactString(discovery.ResponseTypes, "code") ||
		!containsExactString(discovery.CodeChallengeMethods, "S256") {
		return false
	}
	for _, endpoint := range []string{discovery.AuthorizationEndpoint, discovery.TokenEndpoint, discovery.JWKSURI} {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.User != nil || parsed.Fragment != "" || parsed.RawQuery != "" {
			return false
		}
	}
	return true
}

func (client *localIdentityOIDCClient) policyDigest() [sha256.Size]byte {
	canonical := strings.Join([]string{
		localIdentityOIDCPolicyVersion, client.policy.issuer, client.policy.discoveryURL, client.policy.clientID,
		client.policy.redirectURI, strings.Join(client.policy.scopes, " "), strings.Join(sortedMapKeys(client.policy.algorithms), ","),
		fmt.Sprintf("first_login=%t", client.policy.firstLoginEnabled),
	}, "\n")
	return sha256.Sum256([]byte(canonical))
}

func sortedMapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key, enabled := range values {
		if enabled {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}

func (client *localIdentityOIDCClient) authorizationURL(state string, nonce string, codeChallenge string) (string, error) {
	endpoint, err := url.Parse(client.authorizationEndpoint)
	if err != nil {
		return "", errLocalIdentityOIDCCallbackMismatch
	}
	query := endpoint.Query()
	query.Set("response_type", "code")
	query.Set("client_id", client.policy.clientID)
	query.Set("redirect_uri", client.policy.redirectURI)
	query.Set("scope", strings.Join(client.policy.scopes, " "))
	query.Set("state", state)
	query.Set("nonce", nonce)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	endpoint.RawQuery = query.Encode()
	return endpoint.String(), nil
}

func (client *localIdentityOIDCClient) exchangeCode(
	ctx context.Context,
	code string,
	codeVerifier string,
) (string, error) {
	if len(code) < 8 || len(code) > 4096 || strings.ContainsAny(code, "\x00\r\n") || !validOIDCCodeVerifier(codeVerifier) {
		return "", errLocalIdentityOIDCTokenExchange
	}
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {client.policy.redirectURI},
		"client_id":     {client.policy.clientID},
		"code_verifier": {codeVerifier},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", errLocalIdentityOIDCTokenExchange
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return "", errLocalIdentityOIDCProviderUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errLocalIdentityOIDCTokenExchange
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return "", errLocalIdentityOIDCTokenExchange
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, client.verifier.policy.maxResponseBytes+1))
	if err != nil || int64(len(body)) > client.verifier.policy.maxResponseBytes || jsonNestingDepth(body) > 6 {
		return "", errLocalIdentityOIDCTokenExchange
	}
	if err := validateNoDuplicateJSONFields(body); err != nil {
		return "", errLocalIdentityOIDCTokenExchange
	}
	var tokenResponse localIdentityOIDCTokenResponse
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	if err := decoder.Decode(&tokenResponse); err != nil || len(tokenResponse.IDToken) < 32 || len(tokenResponse.IDToken) > 64*1024 {
		return "", errLocalIdentityOIDCTokenExchange
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return "", errLocalIdentityOIDCTokenExchange
	}
	return tokenResponse.IDToken, nil
}

func (client *localIdentityOIDCClient) validateIDToken(
	ctx context.Context,
	rawToken string,
	expectedNonceDigest []byte,
) (localIdentityOIDCVerifiedIdentity, error) {
	segments := strings.Split(rawToken, ".")
	if len(segments) != 3 || len(rawToken) > 64*1024 || len(expectedNonceDigest) != sha256.Size {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil || validateNoDuplicateJSONFields(headerBytes) != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	var header signedTestTokenHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Type != "JWT" ||
		!client.policy.algorithms[header.Algorithm] || !validControlPlaneReadAuthReference(header.KeyID, false) {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	key, err := client.verifier.verificationKey(ctx, header.KeyID, header.Algorithm)
	if errors.Is(err, errOIDCIdentityProviderUnavailable) {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCProviderUnavailable
	}
	if err != nil || verifyOIDCSignature(key.key, header.Algorithm, segments[0]+"."+segments[1], segments[2]) != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	claimBytes, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil || jsonNestingDepth(claimBytes) > 8 || validateNoDuplicateJSONFields(claimBytes) != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	claims := map[string]json.RawMessage{}
	decoder := json.NewDecoder(strings.NewReader(string(claimBytes)))
	if err := decoder.Decode(&claims); err != nil || len(claims) > 64 {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	issuer, err := requiredOIDCStringClaim(claims, "iss")
	if err != nil || issuer != client.policy.issuer {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	audiences, err := requiredOIDCAudiences(claims["aud"])
	if err != nil || !containsExactString(audiences, client.policy.clientID) {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	if _, hasAuthorizedParty := claims["azp"]; len(audiences) > 1 || hasAuthorizedParty {
		authorizedParty, err := requiredOIDCStringClaim(claims, "azp")
		if err != nil || authorizedParty != client.policy.clientID {
			return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
		}
	}
	subject, err := requiredOIDCStringClaim(claims, "sub")
	if err != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	subject, err = NormalizeExternalSubject(subject)
	if err != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	nonce, err := requiredOIDCStringClaim(claims, "nonce")
	nonceDigest := sha256.Sum256([]byte(nonce))
	if err != nil || subtle.ConstantTimeCompare(nonceDigest[:], expectedNonceDigest) != 1 {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	iat, err := requiredOIDCNumericDate(claims, "iat")
	if err != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	exp, err := requiredOIDCNumericDate(claims, "exp")
	if err != nil {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	now := client.verifier.now().UTC()
	if !exp.After(iat) || exp.Sub(iat) > client.verifier.policy.maxTokenLifetime ||
		iat.After(now.Add(client.verifier.policy.clockSkew)) || !exp.After(now.Add(-client.verifier.policy.clockSkew)) {
		return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
	}
	if rawNBF, present := claims["nbf"]; present {
		nbf, err := requiredOIDCNumericDate(map[string]json.RawMessage{"nbf": rawNBF}, "nbf")
		if err != nil || nbf.After(now.Add(client.verifier.policy.clockSkew)) || !exp.After(nbf) {
			return localIdentityOIDCVerifiedIdentity{}, errLocalIdentityOIDCIdentityMismatch
		}
	}
	return localIdentityOIDCVerifiedIdentity{Issuer: issuer, Subject: subject}, nil
}

func localIdentityOIDCSanitizedStartupError(err error) error {
	switch {
	case errors.Is(err, errLocalIdentityOIDCProviderUnavailable):
		return fmt.Errorf("initialize local identity OIDC client: %w", errLocalIdentityOIDCProviderUnavailable)
	default:
		return fmt.Errorf("initialize local identity OIDC client: %w", errLocalIdentityOIDCCallbackMismatch)
	}
}
