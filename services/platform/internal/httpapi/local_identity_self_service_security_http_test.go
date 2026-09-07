package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLocalIdentitySelfServiceHTTPSessionListExactRevokeAndScope(t *testing.T) {
	now := time.Date(2026, 8, 25, 9, 0, 0, 0, time.UTC)
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityHTTPTestFixture(repository, now)
	password := "current password value"
	_, currentCookies, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "self-service@example.com", password)

	loginOne := performLocalIdentityLogin(t, fixture, "self-service@example.com", password)
	loginTwo := performLocalIdentityLogin(t, fixture, "self-service@example.com", password)
	if loginOne.Code != http.StatusOK || loginTwo.Code != http.StatusOK {
		t.Fatalf("create additional sessions: one=%d two=%d", loginOne.Code, loginTwo.Code)
	}
	var loginOneDocument localIdentityAuthenticationDocument
	decodeLocalIdentityHTTPResponse(t, loginOne, &loginOneDocument)

	list := httptest.NewRequest(http.MethodGet, localIdentitySelfServiceSessionListRoute+"?state=active&limit=2", nil)
	addLocalIdentityCookies(list, currentCookies)
	listResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(listResponse, list)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list self-service sessions: status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	var firstPage LocalIdentitySelfServiceSessionPage
	decodeLocalIdentityHTTPResponse(t, listResponse, &firstPage)
	if len(firstPage.Sessions) != 2 || firstPage.NextCursor == "" || firstPage.SnapshotAt != now {
		t.Fatalf("unexpected first session page: %#v", firstPage)
	}
	assertLocalIdentitySelfServiceHTTPBodySafe(t, listResponse.Body.String(), password,
		localIdentityCookieValue(t, currentCookies, fixture.service.sessionCookieName()), "self-service@example.com")

	second := httptest.NewRequest(http.MethodGet, localIdentitySelfServiceSessionListRoute+"?state=active&limit=2&cursor="+firstPage.NextCursor, nil)
	addLocalIdentityCookies(second, currentCookies)
	secondResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(secondResponse, second)
	if secondResponse.Code != http.StatusOK {
		t.Fatalf("list second self-service session page: status=%d body=%s", secondResponse.Code, secondResponse.Body.String())
	}
	var secondPage LocalIdentitySelfServiceSessionPage
	decodeLocalIdentityHTTPResponse(t, secondResponse, &secondPage)
	if len(secondPage.Sessions) != 1 || secondPage.NextCursor != "" || !secondPage.SnapshotAt.Equal(firstPage.SnapshotAt) {
		t.Fatalf("unexpected second session page: %#v", secondPage)
	}
	tampered := httptest.NewRequest(
		http.MethodGet,
		localIdentitySelfServiceSessionListRoute+"?state=active&limit=2&cursor="+tamperLocalIdentitySelfServiceCursor(firstPage.NextCursor),
		nil,
	)
	addLocalIdentityCookies(tampered, currentCookies)
	tamperedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(tamperedResponse, tampered)
	assertLocalIdentityError(t, tamperedResponse, http.StatusBadRequest, LocalIdentityFailureSessionCursorInvalid)

	revokeOther := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(loginOneDocument.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, currentCookies)
	revokeOtherResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(revokeOtherResponse, revokeOther)
	if revokeOtherResponse.Code != http.StatusOK {
		t.Fatalf("revoke owned session: status=%d body=%s", revokeOtherResponse.Code, revokeOtherResponse.Body.String())
	}
	var revocation LocalIdentitySelfServiceSessionRevocation
	decodeLocalIdentityHTTPResponse(t, revokeOtherResponse, &revocation)
	if revocation.CurrentSessionRevoked || revocation.Session.SessionID != loginOneDocument.Session.SessionID ||
		revocation.Session.EffectiveState != localIdentityStateRevoked || revocation.Session.RecordVersion != 2 {
		t.Fatalf("unexpected exact revocation: %#v", revocation)
	}
	assertLocalIdentitySelfServiceCurrentCookieUnchanged(t, fixture, revokeOtherResponse)
	assertLocalIdentitySelfServiceHTTPBodySafe(t, revokeOtherResponse.Body.String(), password)

	duplicate := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(loginOneDocument.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, currentCookies)
	duplicateResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(duplicateResponse, duplicate)
	assertLocalIdentityError(t, duplicateResponse, http.StatusConflict, LocalIdentityFailureSessionVersionConflict)

	otherAccount, otherCredential := localIdentityTestAccount(
		"usr_bbbbbbbbbbbbbbbb", "cred_bbbbbbbbbbbbbbbb", "cross-account@example.com", now.Add(-time.Hour),
	)
	if err := repository.CreateAccount(context.Background(), otherAccount, otherCredential); err != nil {
		t.Fatalf("create cross-account fixture: %v", err)
	}
	otherSession := createLocalIdentitySelfServiceTestSession(
		t, repository, otherAccount.UserID, "ses_bbbbbbbbbbbbbbbb", localAuthenticationMethodPassword,
		"credential:"+otherCredential.CredentialID, now.Add(-time.Minute), now.Add(time.Hour),
	)
	crossAccount := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(otherSession.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, currentCookies)
	crossAccountResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(crossAccountResponse, crossAccount)
	assertLocalIdentityError(t, crossAccountResponse, http.StatusForbidden, LocalIdentityFailureSessionScopeDenied)
	if stored := localIdentitySelfServiceMemorySession(t, repository, otherSession.SessionID); stored.LifecycleState != localIdentityStateActive {
		t.Fatal("cross-account failure changed the target session")
	}

	revokeCurrent := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(registered.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, currentCookies)
	revokeCurrentResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(revokeCurrentResponse, revokeCurrent)
	if revokeCurrentResponse.Code != http.StatusOK {
		t.Fatalf("revoke current session: status=%d body=%s", revokeCurrentResponse.Code, revokeCurrentResponse.Body.String())
	}
	decodeLocalIdentityHTTPResponse(t, revokeCurrentResponse, &revocation)
	if !revocation.CurrentSessionRevoked {
		t.Fatalf("current session revocation was not declared: %#v", revocation)
	}
	assertLocalIdentitySelfServiceCookiesCleared(t, fixture, revokeCurrentResponse)
}

func TestLocalIdentitySelfServiceHTTPBulkAndCredentialRotation(t *testing.T) {
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	password := "current password value"
	newPassword := "replacement password value"

	bulkRepository := newMemoryLocalIdentityRepository()
	bulkFixture := newLocalIdentityHTTPTestFixture(bulkRepository, now)
	_, bulkCookies, _ := registerLocalIdentityHTTPTestAccount(t, bulkFixture, "bulk@example.com", password)
	for range 2 {
		response := performLocalIdentityLogin(t, bulkFixture, "bulk@example.com", password)
		if response.Code != http.StatusOK {
			t.Fatalf("create bulk target session: status=%d body=%s", response.Code, response.Body.String())
		}
	}
	bulk := localIdentitySelfServiceMutationRequest(t, bulkFixture, localIdentitySelfServiceSessionRevokeOthersRoute, map[string]any{
		"confirmed": true,
	}, bulkCookies)
	bulkResponse := httptest.NewRecorder()
	bulkFixture.handler.ServeHTTP(bulkResponse, bulk)
	if bulkResponse.Code != http.StatusOK {
		t.Fatalf("revoke other sessions: status=%d body=%s", bulkResponse.Code, bulkResponse.Body.String())
	}
	var bulkResult LocalIdentitySelfServiceBulkSessionRevocation
	decodeLocalIdentityHTTPResponse(t, bulkResponse, &bulkResult)
	if bulkResult.RevokedCount != 2 {
		t.Fatalf("unexpected bulk revoke result: %#v", bulkResult)
	}
	assertLocalIdentitySelfServiceCurrentCookieUnchanged(t, bulkFixture, bulkResponse)

	repeatBulk := localIdentitySelfServiceMutationRequest(t, bulkFixture, localIdentitySelfServiceSessionRevokeOthersRoute, map[string]any{
		"confirmed": true,
	}, bulkCookies)
	repeatBulkResponse := httptest.NewRecorder()
	bulkFixture.handler.ServeHTTP(repeatBulkResponse, repeatBulk)
	if repeatBulkResponse.Code != http.StatusOK {
		t.Fatalf("repeat bulk revoke: status=%d body=%s", repeatBulkResponse.Code, repeatBulkResponse.Body.String())
	}
	decodeLocalIdentityHTTPResponse(t, repeatBulkResponse, &bulkResult)
	if bulkResult.RevokedCount != 0 {
		t.Fatalf("repeat bulk revoke was not zero-effect: %#v", bulkResult)
	}

	rotationRepository := newMemoryLocalIdentityRepository()
	rotationFixture := newLocalIdentityHTTPTestFixture(rotationRepository, now)
	_, rotationCookies, registered := registerLocalIdentityHTTPTestAccount(t, rotationFixture, "rotate@example.com", password)
	rotationFixture.setNow(now.Add(time.Minute))
	credentialBefore, err := rotationRepository.ReadActiveCredential(context.Background(), registered.Account.UserID)
	if err != nil {
		t.Fatalf("read credential before rotation: %v", err)
	}
	wrongCurrent := localIdentitySelfServiceMutationRequest(t, rotationFixture, localIdentitySelfServiceCredentialRotateRoute, map[string]any{
		"current_password": "wrong current password", "new_password": newPassword, "session_impact_confirmed": true,
	}, rotationCookies)
	wrongCurrentResponse := httptest.NewRecorder()
	rotationFixture.handler.ServeHTTP(wrongCurrentResponse, wrongCurrent)
	assertLocalIdentityError(t, wrongCurrentResponse, http.StatusUnauthorized, LocalIdentityFailureCredentialCurrentInvalid)
	credentialAfterFailure, err := rotationRepository.ReadActiveCredential(context.Background(), registered.Account.UserID)
	if err != nil || credentialAfterFailure.CredentialID != credentialBefore.CredentialID || credentialAfterFailure.RecordVersion != credentialBefore.RecordVersion {
		t.Fatalf("failed credential proof changed active credential: credential=%#v err=%v", credentialAfterFailure, err)
	}

	reuse := localIdentitySelfServiceMutationRequest(t, rotationFixture, localIdentitySelfServiceCredentialRotateRoute, map[string]any{
		"current_password": password, "new_password": password, "session_impact_confirmed": true,
	}, rotationCookies)
	reuseResponse := httptest.NewRecorder()
	rotationFixture.handler.ServeHTTP(reuseResponse, reuse)
	assertLocalIdentityError(t, reuseResponse, http.StatusConflict, LocalIdentityFailureCredentialReuseDenied)

	previousLogWriter := log.Writer()
	var observedLogs bytes.Buffer
	log.SetOutput(&observedLogs)
	rotate := localIdentitySelfServiceMutationRequest(t, rotationFixture, localIdentitySelfServiceCredentialRotateRoute, map[string]any{
		"current_password": password, "new_password": newPassword, "session_impact_confirmed": true,
	}, rotationCookies)
	rotateResponse := httptest.NewRecorder()
	rotationFixture.handler.ServeHTTP(rotateResponse, rotate)
	log.SetOutput(previousLogWriter)
	if rotateResponse.Code != http.StatusOK {
		t.Fatalf("rotate credential: status=%d body=%s", rotateResponse.Code, rotateResponse.Body.String())
	}
	var rotation LocalIdentitySelfServiceCredentialRotation
	decodeLocalIdentityHTTPResponse(t, rotateResponse, &rotation)
	if !rotation.CurrentSessionRevoked || rotation.RevokedSessionCount != 1 || rotation.PolicyVersion != localPasswordPolicyVersion {
		t.Fatalf("unexpected credential rotation result: %#v", rotation)
	}
	assertLocalIdentitySelfServiceCookiesCleared(t, rotationFixture, rotateResponse)
	assertLocalIdentitySelfServiceHTTPBodySafe(t, rotateResponse.Body.String(), password, newPassword)
	assertLocalIdentitySelfServiceHTTPBodySafe(t, observedLogs.String(), password, newPassword)

	oldLogin := performLocalIdentityLogin(t, rotationFixture, "rotate@example.com", password)
	newLogin := performLocalIdentityLogin(t, rotationFixture, "rotate@example.com", newPassword)
	assertLocalIdentityError(t, oldLogin, http.StatusUnauthorized, localIdentityAuthenticationFailed)
	if newLogin.Code != http.StatusOK {
		t.Fatalf("replacement password did not authenticate: status=%d body=%s", newLogin.Code, newLogin.Body.String())
	}
}

func TestLocalIdentitySelfServiceHTTPOIDCCurrentSessionSurvivesLocalCredentialRotation(t *testing.T) {
	now := time.Date(2026, 8, 25, 11, 0, 0, 0, time.UTC)
	password := "current password value"
	newPassword := "replacement password value"
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityHTTPTestFixture(repository, now)
	_, _, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "oidc-current@example.com", password)

	rawCredential := "oidc-current-session-credential-material"
	digest, err := DigestWebSessionCredential(rawCredential)
	if err != nil {
		t.Fatalf("digest OIDC current session credential: %v", err)
	}
	oidcSession := WebSession{
		SchemaVersion: localIdentitySchemaVersion, SessionID: "ses_cccccccccccccccc", UserID: registered.Account.UserID,
		credentialDigest: digest[:], AuthenticationMethod: localAuthenticationMethodOIDC,
		AuthenticationSourceRef: "binding:xid_cccccccccccccccc", PolicyVersion: localSessionPolicyVersion,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		LastVerifiedAt: now, ExpiresAt: now.Add(time.Hour), AuditRef: "audit:oidc-current-session",
	}
	if err := repository.CreateWebSession(context.Background(), oidcSession); err != nil {
		t.Fatalf("create OIDC current session: %v", err)
	}
	fixture.setNow(now.Add(time.Minute))
	oidcCookies := []*http.Cookie{
		{Name: fixture.service.sessionCookieName(), Value: rawCredential},
		{Name: fixture.service.csrfCookieName(), Value: deriveLocalIdentityCSRFToken(rawCredential)},
	}
	rotate := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceCredentialRotateRoute, map[string]any{
		"current_password": password, "new_password": newPassword, "session_impact_confirmed": true,
	}, oidcCookies)
	response := httptest.NewRecorder()
	fixture.handler.ServeHTTP(response, rotate)
	if response.Code != http.StatusOK {
		t.Fatalf("rotate credential from OIDC session: status=%d body=%s", response.Code, response.Body.String())
	}
	var rotation LocalIdentitySelfServiceCredentialRotation
	decodeLocalIdentityHTTPResponse(t, response, &rotation)
	if rotation.CurrentSessionRevoked || rotation.RevokedSessionCount != 1 {
		t.Fatalf("OIDC session impact mismatch: %#v", rotation)
	}
	assertLocalIdentitySelfServiceCurrentCookieUnchanged(t, fixture, response)

	current := httptest.NewRequest(http.MethodGet, localIdentityCurrentSessionRoute, nil)
	addLocalIdentityCookies(current, oidcCookies)
	currentResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(currentResponse, current)
	if currentResponse.Code != http.StatusOK {
		t.Fatalf("OIDC current session did not survive local rotation: status=%d body=%s", currentResponse.Code, currentResponse.Body.String())
	}
	if passwordLogin := performLocalIdentityLogin(t, fixture, "oidc-current@example.com", password); passwordLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password survived OIDC-origin rotation: status=%d body=%s", passwordLogin.Code, passwordLogin.Body.String())
	}
}

func TestLocalIdentitySelfServiceHTTPAuthorizationStrictnessAndZeroSideEffects(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	repository := newMemoryLocalIdentityRepository()
	fixture := newLocalIdentityHTTPTestFixture(repository, now)
	password := "current password value"
	_, cookies, _ := registerLocalIdentityHTTPTestAccount(t, fixture, "strict@example.com", password)
	login := performLocalIdentityLogin(t, fixture, "strict@example.com", password)
	var target localIdentityAuthenticationDocument
	decodeLocalIdentityHTTPResponse(t, login, &target)

	unauthenticated := httptest.NewRequest(http.MethodGet, localIdentitySelfServiceSessionListRoute, nil)
	unauthenticatedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(unauthenticatedResponse, unauthenticated)
	assertLocalIdentityError(t, unauthenticatedResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	for name, header := range map[string][2]string{
		"bearer":     {"Authorization", "Bearer signed-test-token"},
		"dev header": {controlPlaneReadDevIdentityHeader, "dev:test"},
	} {
		request := httptest.NewRequest(http.MethodGet, localIdentitySelfServiceSessionListRoute, nil)
		addLocalIdentityCookies(request, cookies)
		request.Header.Set(header[0], header[1])
		response := httptest.NewRecorder()
		fixture.handler.ServeHTTP(response, request)
		assertLocalIdentityError(t, response, http.StatusUnauthorized, localIdentityAuthenticationRequired)
		if strings.Contains(response.Body.String(), target.Account.UserID) {
			t.Fatalf("%s response leaked local actor: %s", name, response.Body.String())
		}
	}

	for _, rawQuery := range []string{"unknown=value", "limit=1&limit=2", "state=", "limit=101"} {
		request := httptest.NewRequest(http.MethodGet, localIdentitySelfServiceSessionListRoute+"?"+rawQuery, nil)
		addLocalIdentityCookies(request, cookies)
		response := httptest.NewRecorder()
		fixture.handler.ServeHTTP(response, request)
		assertLocalIdentityError(t, response, http.StatusBadRequest, LocalIdentityFailureSessionCursorInvalid)
	}

	wrongMethod := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, cookies)
	wrongMethod.Method = http.MethodPut
	wrongMethodResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(wrongMethodResponse, wrongMethod)
	if wrongMethodResponse.Code != http.StatusMethodNotAllowed {
		t.Fatalf("wrong method was not rejected: status=%d body=%s", wrongMethodResponse.Code, wrongMethodResponse.Body.String())
	}

	missingOrigin := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, cookies)
	missingOrigin.Header.Del("Origin")
	missingOriginResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(missingOriginResponse, missingOrigin)
	assertLocalIdentityError(t, missingOriginResponse, http.StatusForbidden, localIdentityOriginForbidden)

	missingCSRF := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, cookies)
	missingCSRF.Header.Del(localIdentityCSRFHeader)
	missingCSRFResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(missingCSRFResponse, missingCSRF)
	assertLocalIdentityError(t, missingCSRFResponse, http.StatusForbidden, localIdentityCSRFInvalid)

	for name, body := range map[string]string{
		"unknown field":   `{"expected_record_version":1,"confirmed":true,"user_id":"usr_attack"}`,
		"duplicate field": `{"expected_record_version":1,"expected_record_version":2,"confirmed":true}`,
		"multiple values": `{"expected_record_version":1,"confirmed":true}{"expected_record_version":1,"confirmed":true}`,
	} {
		request := httptest.NewRequest(http.MethodPost, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", localIdentityHTTPTestOrigin)
		request.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName()))
		addLocalIdentityCookies(request, cookies)
		response := httptest.NewRecorder()
		fixture.handler.ServeHTTP(response, request)
		assertLocalIdentityError(t, response, http.StatusBadRequest, "INVALID_JSON")
		if stored := localIdentitySelfServiceMemorySession(t, repository, target.Session.SessionID); stored.LifecycleState != localIdentityStateActive {
			t.Fatalf("%s changed target session", name)
		}
	}

	unconfirmed := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": false,
	}, cookies)
	unconfirmedResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(unconfirmedResponse, unconfirmed)
	assertLocalIdentityError(t, unconfirmedResponse, http.StatusBadRequest, localIdentityPayloadInvalid)

	currentRaw := localIdentityCookieValue(t, cookies, fixture.service.sessionCookieName())
	currentDigest, err := DigestWebSessionCredential(currentRaw)
	if err != nil {
		t.Fatalf("digest current session credential: %v", err)
	}
	currentSession, _, err := repository.ResolveWebSession(context.Background(), currentDigest, now)
	if err != nil {
		t.Fatalf("resolve current session: %v", err)
	}
	repository.mu.Lock()
	stale := repository.sessions[currentSession.SessionID]
	stale.LastVerifiedAt = now.Add(-localIdentitySelfServiceRecentAuthenticationAge - time.Second)
	repository.sessions[currentSession.SessionID] = stale
	repository.mu.Unlock()

	staleRequest := localIdentitySelfServiceMutationRequest(t, fixture, localIdentitySelfServiceSessionRevokeRouteFor(target.Session.SessionID), map[string]any{
		"expected_record_version": 1, "confirmed": true,
	}, cookies)
	staleResponse := httptest.NewRecorder()
	fixture.handler.ServeHTTP(staleResponse, staleRequest)
	assertLocalIdentityError(t, staleResponse, http.StatusUnauthorized, LocalIdentityFailureSessionRecentAuthentication)
	if stored := localIdentitySelfServiceMemorySession(t, repository, target.Session.SessionID); stored.LifecycleState != localIdentityStateActive {
		t.Fatal("recent-auth failure changed target session")
	}
}

func TestLocalIdentitySelfServiceHTTPStableFailureMapping(t *testing.T) {
	for _, testCase := range []struct {
		code   string
		status int
	}{
		{LocalIdentityFailureSessionCursorInvalid, http.StatusBadRequest},
		{LocalIdentityFailureSessionScopeDenied, http.StatusForbidden},
		{LocalIdentityFailureSessionVersionConflict, http.StatusConflict},
		{LocalIdentityFailureSessionRecentAuthentication, http.StatusUnauthorized},
		{LocalIdentityFailureSessionBulkRevokeConflict, http.StatusConflict},
		{LocalIdentityFailureCredentialUnavailable, http.StatusConflict},
		{LocalIdentityFailureCredentialCurrentInvalid, http.StatusUnauthorized},
		{LocalIdentityFailureCredentialPolicyRejected, http.StatusBadRequest},
		{LocalIdentityFailureCredentialReuseDenied, http.StatusConflict},
		{LocalIdentityFailureCredentialRotationConflict, http.StatusConflict},
		{LocalIdentityFailureSelfServiceUnavailable, http.StatusServiceUnavailable},
	} {
		status, errorType, message := localIdentitySelfServiceSecurityHTTPErrorDefinition(testCase.code)
		if status != testCase.status || strings.TrimSpace(errorType) == "" || strings.TrimSpace(message) == "" {
			t.Fatalf("failure mapping drifted for %s: status=%d type=%q message=%q", testCase.code, status, errorType, message)
		}
		assertLocalIdentitySelfServiceHTTPBodySafe(t, message)
	}
}

func localIdentitySelfServiceMutationRequest(
	t *testing.T,
	fixture *localIdentityHTTPTestFixture,
	path string,
	body any,
	cookies []*http.Cookie,
) *http.Request {
	t.Helper()
	request := localIdentityHTTPJSONRequest(t, http.MethodPost, path, body, cookies)
	request.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, cookies, fixture.service.csrfCookieName()))
	return request
}

func localIdentitySelfServiceSessionRevokeRouteFor(sessionID string) string {
	return "/v1/auth/sessions/" + sessionID + "/revoke"
}

func localIdentitySelfServiceMemorySession(t *testing.T, repository *memoryLocalIdentityRepository, sessionID string) WebSession {
	t.Helper()
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	return cloneWebSession(repository.sessions[sessionID])
}

func assertLocalIdentitySelfServiceCurrentCookieUnchanged(
	t *testing.T,
	fixture *localIdentityHTTPTestFixture,
	response *httptest.ResponseRecorder,
) {
	t.Helper()
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == fixture.service.sessionCookieName() {
			t.Fatalf("non-current mutation rewrote the current session cookie: %#v", cookie)
		}
	}
}

func assertLocalIdentitySelfServiceCookiesCleared(
	t *testing.T,
	fixture *localIdentityHTTPTestFixture,
	response *httptest.ResponseRecorder,
) {
	t.Helper()
	cleared := map[string]bool{}
	for _, cookie := range response.Result().Cookies() {
		if cookie.MaxAge < 0 {
			cleared[cookie.Name] = true
		}
	}
	if !cleared[fixture.service.sessionCookieName()] || !cleared[fixture.service.csrfCookieName()] {
		t.Fatalf("current-session mutation did not clear both authentication cookies: %#v", response.Result().Cookies())
	}
}

func assertLocalIdentitySelfServiceHTTPBodySafe(t *testing.T, body string, secrets ...string) {
	t.Helper()
	for _, forbidden := range append(secrets,
		"credential_digest", "credential_id", "authentication_source_ref", "audit_ref", "login_identifier", "password_hash", "current_password", "new_password",
	) {
		if forbidden != "" && strings.Contains(body, forbidden) {
			t.Fatalf("self-service HTTP material leaked %q: %s", forbidden, body)
		}
	}
	var document any
	if strings.HasPrefix(strings.TrimSpace(body), "{") && json.Unmarshal([]byte(body), &document) != nil {
		t.Fatalf("self-service HTTP body is not valid JSON: %s", body)
	}
}
