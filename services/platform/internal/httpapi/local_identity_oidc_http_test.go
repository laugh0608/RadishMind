package httpapi

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

const localIdentityOIDCTestClientID = "radishmind-loopback-client"

type localIdentityOIDCTestCode struct {
	subject       string
	nonce         string
	codeChallenge string
	used          bool
}

type localIdentityOIDCTestIssuer struct {
	testing *testing.T
	server  *httptest.Server
	mu      sync.Mutex
	keys    map[string]*rsa.PrivateKey
	keyID   string
	codes   map[string]*localIdentityOIDCTestCode
	next    int
	now     time.Time

	failToken          bool
	nonceOverride      string
	claimOverrides     map[string]any
	signingAlgorithm   string
	signingKeyOverride *rsa.PrivateKey
	tokenHits          int
}

func newLocalIdentityOIDCTestIssuer(t *testing.T, now time.Time) *localIdentityOIDCTestIssuer {
	t.Helper()
	key := generateSignedTestPrivateKey(t)
	issuer := &localIdentityOIDCTestIssuer{
		testing: t, keys: map[string]*rsa.PrivateKey{"oidc-login-key": key}, keyID: "oidc-login-key",
		codes: map[string]*localIdentityOIDCTestCode{}, now: now,
	}
	issuer.server = httptest.NewServer(http.HandlerFunc(issuer.handle))
	t.Cleanup(issuer.server.Close)
	return issuer
}

func (issuer *localIdentityOIDCTestIssuer) handle(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/.well-known/openid-configuration":
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"issuer": issuer.server.URL, "authorization_endpoint": issuer.server.URL + "/authorize",
			"token_endpoint": issuer.server.URL + "/token", "jwks_uri": issuer.server.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"response_types_supported":              []string{"code"}, "code_challenge_methods_supported": []string{"S256"},
		})
	case "/jwks":
		issuer.mu.Lock()
		keys := make(map[string]*rsa.PrivateKey, len(issuer.keys))
		for keyID, key := range issuer.keys {
			keys[keyID] = key
		}
		issuer.mu.Unlock()
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(marshalOIDCJWKS(keys))
	case "/token":
		issuer.handleToken(writer, request)
	default:
		http.NotFound(writer, request)
	}
}

func (issuer *localIdentityOIDCTestIssuer) handleToken(writer http.ResponseWriter, request *http.Request) {
	issuer.mu.Lock()
	defer issuer.mu.Unlock()
	issuer.tokenHits++
	if issuer.failToken {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if request.Method != http.MethodPost || request.ParseForm() != nil ||
		request.Form.Get("grant_type") != "authorization_code" ||
		request.Form.Get("client_id") != localIdentityOIDCTestClientID ||
		request.Form.Get("redirect_uri") != localIdentityHTTPTestOrigin+localIdentityOIDCCallbackRoute {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	code := issuer.codes[request.Form.Get("code")]
	if code == nil || code.used {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	challenge := sha256.Sum256([]byte(request.Form.Get("code_verifier")))
	if base64.RawURLEncoding.EncodeToString(challenge[:]) != code.codeChallenge {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	code.used = true
	nonce := code.nonce
	if issuer.nonceOverride != "" {
		nonce = issuer.nonceOverride
	}
	claims := map[string]any{
		"iss": issuer.server.URL, "aud": localIdentityOIDCTestClientID, "sub": code.subject,
		"nonce": nonce, "iat": issuer.now.Unix(), "exp": issuer.now.Add(5 * time.Minute).Unix(),
		"email": "shared@example.com", "roles": []string{"upstream-admin"}, "permissions": []string{"all"},
	}
	for name, value := range issuer.claimOverrides {
		claims[name] = value
	}
	algorithm := issuer.signingAlgorithm
	if algorithm == "" {
		algorithm = "RS256"
	}
	signingKey := issuer.keys[issuer.keyID]
	if issuer.signingKeyOverride != nil {
		signingKey = issuer.signingKeyOverride
	}
	idToken := signOIDCTestToken(issuer.testing, signingKey, issuer.keyID, algorithm, claims)
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(map[string]any{
		"id_token": idToken, "access_token": "raw-access-token-must-not-survive", "token_type": "Bearer",
	})
}

func (issuer *localIdentityOIDCTestIssuer) issueCode(t *testing.T, authorizationURL string, subject string) (string, string) {
	t.Helper()
	parsed, err := url.Parse(authorizationURL)
	if err != nil || parsed.Scheme+"://"+parsed.Host+parsed.Path != issuer.server.URL+"/authorize" {
		t.Fatalf("invalid authorization URL: %q err=%v", authorizationURL, err)
	}
	query := parsed.Query()
	if query.Get("response_type") != "code" || query.Get("client_id") != localIdentityOIDCTestClientID ||
		query.Get("redirect_uri") != localIdentityHTTPTestOrigin+localIdentityOIDCCallbackRoute ||
		query.Get("scope") != "openid profile" || query.Get("code_challenge_method") != "S256" ||
		query.Get("state") == "" || query.Get("nonce") == "" || query.Get("code_challenge") == "" {
		t.Fatalf("authorization URL contract mismatch: %s", parsed.RawQuery)
	}
	issuer.mu.Lock()
	defer issuer.mu.Unlock()
	issuer.next++
	code := fmt.Sprintf("authorization-code-%d-with-enough-length", issuer.next)
	issuer.codes[code] = &localIdentityOIDCTestCode{
		subject: subject, nonce: query.Get("nonce"), codeChallenge: query.Get("code_challenge"),
	}
	return code, query.Get("state")
}

func (issuer *localIdentityOIDCTestIssuer) rotateKey(t *testing.T) {
	t.Helper()
	newKey := generateSignedTestPrivateKey(t)
	issuer.mu.Lock()
	issuer.keys = map[string]*rsa.PrivateKey{"oidc-login-key-rotated": newKey}
	issuer.keyID = "oidc-login-key-rotated"
	issuer.mu.Unlock()
}

func (issuer *localIdentityOIDCTestIssuer) setTokenMutation(
	claims map[string]any,
	algorithm string,
	signingKey *rsa.PrivateKey,
	now time.Time,
) {
	issuer.mu.Lock()
	defer issuer.mu.Unlock()
	issuer.claimOverrides = claims
	issuer.signingAlgorithm = algorithm
	issuer.signingKeyOverride = signingKey
	issuer.now = now
}

func newLocalIdentityOIDCHTTPTestFixture(
	t *testing.T,
	issuer *localIdentityOIDCTestIssuer,
	repository localIdentityRepository,
	now time.Time,
	firstLogin bool,
) *localIdentityHTTPTestFixture {
	t.Helper()
	fixture := newLocalIdentityHTTPTestFixture(repository, now)
	cfg := fixture.server.config
	cfg.LocalIdentityOIDCEnabled = true
	cfg.LocalIdentityOIDCIssuer = issuer.server.URL
	cfg.LocalIdentityOIDCDiscoveryURL = issuer.server.URL + "/.well-known/openid-configuration"
	cfg.LocalIdentityOIDCClientID = localIdentityOIDCTestClientID
	cfg.LocalIdentityOIDCRedirectURI = localIdentityHTTPTestOrigin + localIdentityOIDCCallbackRoute
	cfg.LocalIdentityOIDCScopes = "openid,profile"
	cfg.LocalIdentityOIDCAlgorithms = "RS256"
	cfg.LocalIdentityOIDCJWKSOrigin = issuer.server.URL
	cfg.LocalIdentityOIDCTransactionTTL = 5 * time.Minute
	cfg.LocalIdentityOIDCFirstLoginEnabled = firstLogin
	if err := fixture.service.configureOIDC(t.Context(), cfg); err != nil {
		t.Fatalf("configure local identity OIDC test client: %v", err)
	}
	fixture.service.oidcClient.verifier.now = fixture.service.now
	return fixture
}

func TestLocalIdentityOIDCFirstLoginBoundLoginReplayRotationAndPrivacy(t *testing.T) {
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	issuer := newLocalIdentityOIDCTestIssuer(t, now)
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityOIDCHTTPTestFixture(t, issuer, repository, now, true)

	firstStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/workspace", nil)
	firstCode, firstState := issuer.issueCode(t, firstStart.AuthorizationURL, "radish-subject-one")
	firstCallback := performLocalIdentityOIDCCallback(fixture, firstState, firstCode)
	if firstCallback.Code != http.StatusSeeOther || firstCallback.Header().Get("Location") != "/workspace" {
		t.Fatalf("first OIDC callback failed: status=%d body=%s", firstCallback.Code, firstCallback.Body.String())
	}
	cookies := activeLocalIdentityCookies(firstCallback.Result().Cookies())
	if len(cookies) != 2 || strings.Contains(firstCallback.Body.String(), firstCode) ||
		strings.Contains(firstCallback.Body.String(), "raw-access-token") || strings.Contains(firstCallback.Body.String(), "radish-subject-one") {
		t.Fatalf("OIDC callback privacy or cookie contract failed: cookies=%#v body=%s", cookies, firstCallback.Body.String())
	}
	binding, err := repository.ResolveExternalIdentity(t.Context(), issuer.server.URL, "radish-subject-one")
	if err != nil {
		t.Fatalf("resolve first-login binding: %v", err)
	}
	account, err := repository.ReadAccount(t.Context(), binding.UserID)
	if err != nil || account.NormalizedLoginIdentifier == "shared@example.com" || account.DisplayName != "External user" {
		t.Fatalf("OIDC first-login account trusted upstream profile: account=%#v err=%v", account, err)
	}
	if len(repository.roleAssignments) != 0 || len(repository.memberships) != 0 {
		t.Fatalf("upstream authorization claims formed local grants: roles=%d memberships=%d",
			len(repository.roleAssignments), len(repository.memberships))
	}
	current := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(current, cookies)
	currentResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(currentResponse, current)
	if currentResponse.Code != http.StatusOK || !strings.Contains(currentResponse.Body.String(), `"authentication_method":"oidc"`) {
		t.Fatalf("OIDC Web Session was not restored: status=%d body=%s", currentResponse.Code, currentResponse.Body.String())
	}

	replay := performLocalIdentityOIDCCallback(fixture, firstState, firstCode)
	if replay.Code != http.StatusBadRequest || !strings.Contains(replay.Body.String(), localIdentityOIDCStateInvalid) || issuer.tokenHits != 1 {
		t.Fatalf("OIDC replay did not fail before token exchange: status=%d hits=%d body=%s", replay.Code, issuer.tokenHits, replay.Body.String())
	}

	issuer.rotateKey(t)
	secondStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/again", nil)
	secondCode, secondState := issuer.issueCode(t, secondStart.AuthorizationURL, "radish-subject-one")
	secondCallback := performLocalIdentityOIDCCallback(fixture, secondState, secondCode)
	if secondCallback.Code != http.StatusSeeOther || len(repository.accounts) != 1 || len(repository.externalBindings) != 1 || len(repository.sessions) != 2 {
		t.Fatalf("bound OIDC login or key rotation drifted: status=%d accounts=%d bindings=%d sessions=%d body=%s",
			secondCallback.Code, len(repository.accounts), len(repository.externalBindings), len(repository.sessions), secondCallback.Body.String())
	}
}

func TestLocalIdentityOIDCIdentityPolicyConcurrencyAndMalformedCallback(t *testing.T) {
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	issuer := newLocalIdentityOIDCTestIssuer(t, now)
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityOIDCHTTPTestFixture(t, issuer, repository, now, true)

	seedStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	seedCode, seedState := issuer.issueCode(t, seedStart.AuthorizationURL, "policy-subject")
	if response := performLocalIdentityOIDCCallback(fixture, seedState, seedCode); response.Code != http.StatusSeeOther {
		t.Fatalf("seed OIDC identity: status=%d body=%s", response.Code, response.Body.String())
	}
	baselineSessions := len(repository.sessions)
	invalidCases := []struct {
		name       string
		claims     map[string]any
		algorithm  string
		signingKey *rsa.PrivateKey
	}{
		{name: "issuer", claims: map[string]any{"iss": issuer.server.URL + "/different"}},
		{name: "audience", claims: map[string]any{"aud": "different-client"}},
		{name: "multi audience without azp", claims: map[string]any{"aud": []string{localIdentityOIDCTestClientID, "another-client"}}},
		{name: "authorized party", claims: map[string]any{"azp": "different-client"}},
		{name: "expired", claims: map[string]any{"iat": now.Add(-10 * time.Minute).Unix(), "exp": now.Add(-time.Minute).Unix()}},
		{name: "future issued at", claims: map[string]any{"iat": now.Add(time.Minute).Unix(), "exp": now.Add(5 * time.Minute).Unix()}},
		{name: "algorithm", algorithm: "RS384"},
		{name: "signature", signingKey: generateSignedTestPrivateKey(t)},
	}
	for _, testCase := range invalidCases {
		t.Run(testCase.name, func(t *testing.T) {
			start := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
			code, state := issuer.issueCode(t, start.AuthorizationURL, "policy-subject")
			issuer.setTokenMutation(testCase.claims, testCase.algorithm, testCase.signingKey, now)
			response := performLocalIdentityOIDCCallback(fixture, state, code)
			issuer.setTokenMutation(nil, "", nil, now)
			if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), localIdentityOIDCIdentityMismatch) {
				t.Fatalf("invalid ID token was accepted: status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	if len(repository.sessions) != baselineSessions {
		t.Fatalf("invalid ID tokens created sessions: before=%d after=%d", baselineSessions, len(repository.sessions))
	}

	malformedStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	malformedCode, malformedState := issuer.issueCode(t, malformedStart.AuthorizationURL, "policy-subject")
	malformedRequest := httptest.NewRequest(http.MethodGet,
		localIdentityOIDCCallbackRoute+"?state="+url.QueryEscape(malformedState)+"&error=access_denied", nil)
	malformedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(malformedResponse, malformedRequest)
	tokenHitsBeforeReplay := issuer.tokenHits
	malformedReplay := performLocalIdentityOIDCCallback(fixture, malformedState, malformedCode)
	if malformedResponse.Code != http.StatusBadRequest || malformedReplay.Code != http.StatusBadRequest ||
		!strings.Contains(malformedReplay.Body.String(), localIdentityOIDCStateInvalid) || issuer.tokenHits != tokenHitsBeforeReplay {
		t.Fatalf("malformed callback did not consume state before failure: malformed=%d replay=%d hits=%d body=%s",
			malformedResponse.Code, malformedReplay.Code, issuer.tokenHits, malformedReplay.Body.String())
	}

	concurrentRepository := newMemoryLocalIdentityRepository()
	concurrentAdmission := newLocalIdentityOIDCConcurrentAdmissionRepository(
		concurrentRepository,
		issuer.server.URL,
		"concurrent-subject",
	)
	concurrentFixture := newLocalIdentityOIDCHTTPTestFixture(t, issuer, concurrentAdmission, now, true)
	firstStart := startLocalIdentityOIDCTestFlow(t, concurrentFixture, localIdentityOIDCIntentLogin, "/", nil)
	secondStart := startLocalIdentityOIDCTestFlow(t, concurrentFixture, localIdentityOIDCIntentLogin, "/", nil)
	firstCode, firstState := issuer.issueCode(t, firstStart.AuthorizationURL, "concurrent-subject")
	secondCode, secondState := issuer.issueCode(t, secondStart.AuthorizationURL, "concurrent-subject")
	responses := make(chan *httptest.ResponseRecorder, 2)
	var wait sync.WaitGroup
	for _, callback := range [][2]string{{firstState, firstCode}, {secondState, secondCode}} {
		wait.Add(1)
		go func(state string, code string) {
			defer wait.Done()
			responses <- performLocalIdentityOIDCCallback(concurrentFixture, state, code)
		}(callback[0], callback[1])
	}
	wait.Wait()
	close(responses)
	statusCounts := map[int]int{}
	for response := range responses {
		statusCounts[response.Code]++
	}
	if statusCounts[http.StatusSeeOther] != 1 || statusCounts[http.StatusConflict] != 1 ||
		len(concurrentRepository.accounts) != 1 || len(concurrentRepository.externalBindings) != 1 || len(concurrentRepository.sessions) != 1 {
		t.Fatalf("concurrent first login did not converge atomically: statuses=%v accounts=%d bindings=%d sessions=%d",
			statusCounts, len(concurrentRepository.accounts), len(concurrentRepository.externalBindings), len(concurrentRepository.sessions))
	}
}

type localIdentityOIDCConcurrentAdmissionRepository struct {
	localIdentityRepository
	issuer  string
	subject string
	ready   chan struct{}
	mu      sync.Mutex
	arrived int
}

func newLocalIdentityOIDCConcurrentAdmissionRepository(
	repository localIdentityRepository,
	issuer string,
	subject string,
) *localIdentityOIDCConcurrentAdmissionRepository {
	return &localIdentityOIDCConcurrentAdmissionRepository{
		localIdentityRepository: repository,
		issuer:                  issuer,
		subject:                 subject,
		ready:                   make(chan struct{}),
	}
}

func (repository *localIdentityOIDCConcurrentAdmissionRepository) ResolveExternalIdentity(
	ctx context.Context,
	issuer string,
	subject string,
) (ExternalIdentityBinding, error) {
	if issuer != repository.issuer || subject != repository.subject {
		return repository.localIdentityRepository.ResolveExternalIdentity(ctx, issuer, subject)
	}
	repository.mu.Lock()
	if repository.arrived >= 2 {
		repository.mu.Unlock()
		return repository.localIdentityRepository.ResolveExternalIdentity(ctx, issuer, subject)
	}
	repository.arrived++
	if repository.arrived == 2 {
		close(repository.ready)
	}
	repository.mu.Unlock()
	select {
	case <-ctx.Done():
		return ExternalIdentityBinding{}, ctx.Err()
	case <-repository.ready:
		return ExternalIdentityBinding{}, errLocalIdentityNotFound
	}
}

func TestLocalIdentityOIDCAdmissionLinkConflictNonceAndProviderFailure(t *testing.T) {
	now := time.Date(2026, 8, 23, 9, 0, 0, 0, time.UTC)
	issuer := newLocalIdentityOIDCTestIssuer(t, now)
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityOIDCHTTPTestFixture(t, issuer, repository, now, false)

	deniedStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	deniedCode, deniedState := issuer.issueCode(t, deniedStart.AuthorizationURL, "unadmitted-subject")
	denied := performLocalIdentityOIDCCallback(fixture, deniedState, deniedCode)
	if denied.Code != http.StatusForbidden || !strings.Contains(denied.Body.String(), localIdentityOIDCAdmissionDenied) || len(repository.accounts) != 0 {
		t.Fatalf("closed first-login admission produced side effects: status=%d accounts=%d body=%s", denied.Code, len(repository.accounts), denied.Body.String())
	}

	_, firstCookies, firstAccount := registerLocalIdentityHTTPTestAccount(t, fixture, "first@example.com", "a sufficiently long password")
	linkStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLink, "/account", firstCookies)
	linkCode, linkState := issuer.issueCode(t, linkStart.AuthorizationURL, "linkable-subject")
	linked := performLocalIdentityOIDCCallback(fixture, linkState, linkCode)
	linkBinding, linkErr := repository.ResolveExternalIdentity(t.Context(), issuer.server.URL, "linkable-subject")
	if linked.Code != http.StatusSeeOther || linkErr != nil || linkBinding.UserID != firstAccount.Account.UserID {
		t.Fatalf("explicit OIDC link failed: status=%d binding=%#v err=%v body=%s", linked.Code, linkBinding, linkErr, linked.Body.String())
	}
	revokedStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLink, "/account", firstCookies)
	revokedCode, revokedState := issuer.issueCode(t, revokedStart.AuthorizationURL, "revoked-session-subject")
	if _, err := repository.RevokeWebSession(t.Context(), firstAccount.Session.SessionID, 1, now.Add(time.Second), "audit:revoke-before-link-callback"); err != nil {
		t.Fatalf("revoke initiating local session: %v", err)
	}
	revokedCallback := performLocalIdentityOIDCCallback(fixture, revokedState, revokedCode)
	if revokedCallback.Code != http.StatusUnauthorized {
		t.Fatalf("revoked initiating session still linked identity: status=%d body=%s", revokedCallback.Code, revokedCallback.Body.String())
	}
	if _, err := repository.ResolveExternalIdentity(t.Context(), issuer.server.URL, "revoked-session-subject"); !errors.Is(err, errLocalIdentityNotFound) {
		t.Fatalf("revoked initiating session created binding: %v", err)
	}

	_, secondCookies, _ := registerLocalIdentityHTTPTestAccount(t, fixture, "second@example.com", "a sufficiently long password")
	conflictStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLink, "/account", secondCookies)
	conflictCode, conflictState := issuer.issueCode(t, conflictStart.AuthorizationURL, "linkable-subject")
	conflict := performLocalIdentityOIDCCallback(fixture, conflictState, conflictCode)
	if conflict.Code != http.StatusConflict || !strings.Contains(conflict.Body.String(), localIdentityExternalIdentityConflict) {
		t.Fatalf("external identity conflict was not rejected: status=%d body=%s", conflict.Code, conflict.Body.String())
	}

	issuer.nonceOverride = "mismatched-nonce"
	nonceStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	nonceCode, nonceState := issuer.issueCode(t, nonceStart.AuthorizationURL, "linkable-subject")
	nonceFailure := performLocalIdentityOIDCCallback(fixture, nonceState, nonceCode)
	if nonceFailure.Code != http.StatusUnauthorized || !strings.Contains(nonceFailure.Body.String(), localIdentityOIDCIdentityMismatch) {
		t.Fatalf("nonce mismatch was accepted: status=%d body=%s", nonceFailure.Code, nonceFailure.Body.String())
	}
	issuer.nonceOverride = ""
	pkceStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	pkceCode, pkceState := issuer.issueCode(t, pkceStart.AuthorizationURL, "linkable-subject")
	pkceDigest, _ := localIdentityOIDCStateDigest(pkceState)
	repository.mu.Lock()
	pkceTransactionID := repository.oidcTransactionByStateDigest[string(pkceDigest[:])]
	pkceTransaction := repository.oidcAuthorizationTransactions[pkceTransactionID]
	pkceTransaction.codeVerifier = strings.Repeat("z", 43)
	repository.oidcAuthorizationTransactions[pkceTransactionID] = pkceTransaction
	repository.mu.Unlock()
	pkceFailure := performLocalIdentityOIDCCallback(fixture, pkceState, pkceCode)
	if pkceFailure.Code != http.StatusBadGateway || !strings.Contains(pkceFailure.Body.String(), localIdentityOIDCTokenExchangeFailed) {
		t.Fatalf("PKCE mismatch was accepted: status=%d body=%s", pkceFailure.Code, pkceFailure.Body.String())
	}

	fixture.setNow(now.Add(9 * time.Minute))
	issuer.setTokenMutation(nil, "", nil, now.Add(9*time.Minute))
	recentStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLink, "/account", secondCookies)
	recentCode, recentState := issuer.issueCode(t, recentStart.AuthorizationURL, "recent-auth-subject")
	fixture.setNow(now.Add(11 * time.Minute))
	issuer.setTokenMutation(nil, "", nil, now.Add(11*time.Minute))
	recentFailure := performLocalIdentityOIDCCallback(fixture, recentState, recentCode)
	if recentFailure.Code != http.StatusUnauthorized || !strings.Contains(recentFailure.Body.String(), localIdentityLinkRecentAuthRequired) {
		t.Fatalf("stale initiating authentication was accepted at callback: status=%d body=%s", recentFailure.Code, recentFailure.Body.String())
	}
	if _, err := repository.ResolveExternalIdentity(t.Context(), issuer.server.URL, "recent-auth-subject"); !errors.Is(err, errLocalIdentityNotFound) {
		t.Fatalf("stale initiating authentication created binding: %v", err)
	}
	staleStartRequest := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityOIDCStartRoute, map[string]any{
		"intent": localIdentityOIDCIntentLink, "return_to": "/account",
	}, secondCookies)
	staleStartRequest.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, secondCookies, fixture.service.csrfCookieName()))
	staleStartResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(staleStartResponse, staleStartRequest)
	if staleStartResponse.Code != http.StatusUnauthorized || !strings.Contains(staleStartResponse.Body.String(), localIdentityLinkRecentAuthRequired) {
		t.Fatalf("stale authentication started linking: status=%d body=%s", staleStartResponse.Code, staleStartResponse.Body.String())
	}

	issuer.failToken = true
	unavailableStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	unavailableCode, unavailableState := issuer.issueCode(t, unavailableStart.AuthorizationURL, "linkable-subject")
	unavailable := performLocalIdentityOIDCCallback(fixture, unavailableState, unavailableCode)
	if unavailable.Code != http.StatusBadGateway || !strings.Contains(unavailable.Body.String(), localIdentityOIDCTokenExchangeFailed) {
		t.Fatalf("provider token failure was not sanitized: status=%d body=%s", unavailable.Code, unavailable.Body.String())
	}
	issuer.failToken = false
	networkStart := startLocalIdentityOIDCTestFlow(t, fixture, localIdentityOIDCIntentLogin, "/", nil)
	networkCode, networkState := issuer.issueCode(t, networkStart.AuthorizationURL, "linkable-subject")
	issuer.server.Close()
	networkFailure := performLocalIdentityOIDCCallback(fixture, networkState, networkCode)
	if networkFailure.Code != http.StatusServiceUnavailable || !strings.Contains(networkFailure.Body.String(), localIdentityOIDCProviderUnavailable) {
		t.Fatalf("provider network failure was not failed closed: status=%d body=%s", networkFailure.Code, networkFailure.Body.String())
	}
}

func startLocalIdentityOIDCTestFlow(
	t *testing.T,
	fixture *localIdentityHTTPTestFixture,
	intent string,
	returnTo string,
	cookies []*http.Cookie,
) localIdentityOIDCStartDocument {
	t.Helper()
	request := localIdentityHTTPJSONRequest(t, http.MethodPost, localIdentityOIDCStartRoute, map[string]any{
		"intent": intent, "return_to": returnTo,
	}, cookies)
	if intent == localIdentityOIDCIntentLink {
		request.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName()))
	}
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("start OIDC %s flow: status=%d body=%s", intent, response.Code, response.Body.String())
	}
	var document localIdentityOIDCStartDocument
	decodeLocalIdentityHTTPResponse(t, response, &document)
	return document
}

func performLocalIdentityOIDCCallback(fixture *localIdentityHTTPTestFixture, state string, code string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, localIdentityOIDCCallbackRoute+"?state="+url.QueryEscape(state)+"&code="+url.QueryEscape(code), nil)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, request)
	return response
}
