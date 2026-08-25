package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

var localIdentitySelfServiceTestNow = time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)

func TestMemoryLocalIdentitySelfServiceSessionPaginationAndCursorBinding(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	createdAt := fixture.now.Add(-time.Hour)
	for index := 1; index <= 120; index++ {
		sessionID := fmt.Sprintf("ses_%016x", index)
		expiresAt := fixture.now.Add(8 * time.Hour)
		if index == 2 {
			expiresAt = fixture.now.Add(30 * time.Second)
		}
		createLocalIdentitySelfServiceTestSession(
			t,
			fixture.repository,
			fixture.userID,
			sessionID,
			localAuthenticationMethodPassword,
			"credential:"+fixture.credentialID,
			createdAt,
			expiresAt,
		)
	}
	ctx := context.Background()
	query := LocalIdentitySelfServiceSessionListQuery{Limit: 50}
	firstPage, err := fixture.service.ListSessions(ctx, fixture.actor, query)
	if err != nil {
		t.Fatalf("list first self-service session page: %v", err)
	}
	if len(firstPage.Sessions) != 50 || firstPage.NextCursor == "" || !firstPage.SnapshotAt.Equal(fixture.now) {
		t.Fatalf("first session page mismatch: %#v", firstPage)
	}
	firstCursor := firstPage.NextCursor
	fixture.service.now = func() time.Time { return fixture.now.Add(time.Minute) }
	if _, err := fixture.repository.RevokeWebSession(
		ctx,
		"ses_0000000000000003",
		1,
		fixture.now.Add(30*time.Second),
		"audit:post-snapshot-revoke",
	); err != nil {
		t.Fatalf("prepare post-snapshot revoke: %v", err)
	}
	createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_ffffffffffffffff",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(45*time.Second),
		fixture.now.Add(9*time.Hour),
	)

	seen := append([]LocalIdentitySelfServiceSessionSummary(nil), firstPage.Sessions...)
	cursor := firstCursor
	for cursor != "" {
		page, listErr := fixture.service.ListSessions(ctx, fixture.actor, LocalIdentitySelfServiceSessionListQuery{
			Limit:  50,
			Cursor: cursor,
		})
		if listErr != nil {
			t.Fatalf("list subsequent self-service session page: %v", listErr)
		}
		if !page.SnapshotAt.Equal(firstPage.SnapshotAt) {
			t.Fatalf("session snapshot drifted: first=%s next=%s", firstPage.SnapshotAt, page.SnapshotAt)
		}
		seen = append(seen, page.Sessions...)
		cursor = page.NextCursor
		if len(seen) > 121 {
			t.Fatal("session pagination did not terminate")
		}
	}
	if len(seen) != 121 {
		t.Fatalf("session pagination count mismatch: %d", len(seen))
	}
	ids := make([]string, 0, len(seen))
	seenIDs := make(map[string]struct{}, len(seen))
	foundExpiryBoundary := false
	foundPostSnapshotRevoke := false
	foundCurrent := false
	for _, summary := range seen {
		if _, duplicate := seenIDs[summary.SessionID]; duplicate {
			t.Fatalf("session pagination repeated %s", summary.SessionID)
		}
		seenIDs[summary.SessionID] = struct{}{}
		ids = append(ids, summary.SessionID)
		foundExpiryBoundary = foundExpiryBoundary || summary.SessionID == "ses_0000000000000002" && summary.EffectiveState == localIdentityStateActive
		foundPostSnapshotRevoke = foundPostSnapshotRevoke ||
			summary.SessionID == "ses_0000000000000003" && summary.EffectiveState == localIdentityStateActive && summary.RevokedAt == nil
		foundCurrent = foundCurrent || summary.SessionID == fixture.actor.CurrentSessionID && summary.CurrentSession
	}
	if !slices.IsSortedFunc(ids, func(left, right string) int { return strings.Compare(right, left) }) {
		t.Fatalf("same-timestamp sessions are not ordered by session_id DESC: %v", ids)
	}
	if !foundExpiryBoundary || !foundPostSnapshotRevoke || !foundCurrent {
		t.Fatalf(
			"snapshot/current projection mismatch: expiry=%t post_snapshot_revoke=%t current=%t",
			foundExpiryBoundary,
			foundPostSnapshotRevoke,
			foundCurrent,
		)
	}
	if _, exists := seenIDs["ses_ffffffffffffffff"]; exists {
		t.Fatal("session created after snapshot leaked into later page")
	}

	for name, invalidQuery := range map[string]LocalIdentitySelfServiceSessionListQuery{
		"limit drift": {Limit: 25, Cursor: firstCursor},
		"state drift": {State: localIdentityStateRevoked, Limit: 50, Cursor: firstCursor},
		"tampered":    {Limit: 50, Cursor: tamperLocalIdentitySelfServiceCursor(firstCursor)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := fixture.service.ListSessions(ctx, fixture.actor, invalidQuery); !errors.Is(err, errLocalIdentitySessionCursorInvalid) {
				t.Fatalf("session cursor drift was accepted: %v", err)
			}
		})
	}

	other := newLocalIdentitySelfServiceTestFixtureWithIDs(
		t,
		"usr_bbbbbbbbbbbbbbbb",
		"cred_bbbbbbbbbbbbbbbb",
		"ses_bbbbbbbbbbbbbbbb",
		localAuthenticationMethodPassword,
	)
	if _, err := other.service.ListSessions(ctx, other.actor, LocalIdentitySelfServiceSessionListQuery{
		Limit: 50, Cursor: firstCursor,
	}); !errors.Is(err, errLocalIdentitySessionCursorInvalid) {
		t.Fatalf("cross-owner session cursor was accepted: %v", err)
	}
}

func TestMemoryLocalIdentitySelfServiceSessionStateProjectionAndSanitizedJSON(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	active := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodOIDC,
		"binding:xid_1111111111111111",
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_2222222222222222",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-2*time.Hour),
		fixture.now.Add(-time.Minute),
	)
	revoked := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_3333333333333333",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-3*time.Hour),
		fixture.now.Add(time.Hour),
	)
	if _, err := fixture.repository.RevokeWebSession(
		context.Background(), revoked.SessionID, revoked.RecordVersion, fixture.now.Add(-30*time.Second), "audit:test-revoke",
	); err != nil {
		t.Fatalf("prepare revoked session: %v", err)
	}

	for state, expectedSessionID := range map[string]string{
		localIdentitySessionEffectiveStateExpired: "ses_2222222222222222",
		localIdentityStateRevoked:                 "ses_3333333333333333",
	} {
		page, err := fixture.service.ListSessions(context.Background(), fixture.actor, LocalIdentitySelfServiceSessionListQuery{
			State: state,
		})
		if err != nil || len(page.Sessions) != 1 || page.Sessions[0].SessionID != expectedSessionID ||
			page.Sessions[0].EffectiveState != state {
			t.Fatalf("session state projection mismatch for %s: page=%#v err=%v", state, page, err)
		}
	}
	all, err := fixture.service.ListSessions(context.Background(), fixture.actor, LocalIdentitySelfServiceSessionListQuery{
		State: localIdentitySessionStateFilterAll,
	})
	if err != nil || len(all.Sessions) != 4 {
		t.Fatalf("all session projection mismatch: page=%#v err=%v", all, err)
	}
	payload, err := json.Marshal(all)
	if err != nil {
		t.Fatalf("marshal self-service session page: %v", err)
	}
	for _, forbidden := range []string{
		fixture.userID,
		fixture.credentialID,
		active.AuthenticationSourceRef,
		"authentication_source_ref",
		"credential_digest",
		"audit_ref",
		"cookie",
		"user_agent",
		"ip_address",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("self-service session page leaked %q: %s", forbidden, payload)
		}
	}
}

func TestLocalIdentitySelfServiceStableFailureCodes(t *testing.T) {
	for _, testCase := range []struct {
		err  error
		code string
	}{
		{errLocalIdentitySessionCursorInvalid, LocalIdentityFailureSessionCursorInvalid},
		{errLocalIdentitySessionScopeDenied, LocalIdentityFailureSessionScopeDenied},
		{errLocalIdentitySessionVersionConflict, LocalIdentityFailureSessionVersionConflict},
		{errLocalIdentitySessionRecentAuthentication, LocalIdentityFailureSessionRecentAuthentication},
		{errLocalIdentitySessionBulkRevokeConflict, LocalIdentityFailureSessionBulkRevokeConflict},
		{errLocalIdentityCredentialUnavailable, LocalIdentityFailureCredentialUnavailable},
		{errLocalIdentityCredentialCurrentInvalid, LocalIdentityFailureCredentialCurrentInvalid},
		{errLocalIdentityCredentialPolicyRejected, LocalIdentityFailureCredentialPolicyRejected},
		{errLocalIdentityCredentialReuseDenied, LocalIdentityFailureCredentialReuseDenied},
		{errLocalIdentityCredentialRotationConflict, LocalIdentityFailureCredentialRotationConflict},
		{errLocalIdentitySelfServiceUnavailable, LocalIdentityFailureSelfServiceUnavailable},
	} {
		if got := localIdentityRepositoryError(testCase.err); got != testCase.code {
			t.Fatalf("self-service failure code mismatch: got=%q want=%q", got, testCase.code)
		}
	}
}

func TestMemoryLocalIdentitySelfServiceExactRevokeScopeRecentAuthenticationAndCAS(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	target := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodOIDC,
		"binding:xid_1111111111111111",
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	otherAccount, otherCredential := localIdentityTestAccount(
		"usr_bbbbbbbbbbbbbbbb",
		"cred_bbbbbbbbbbbbbbbb",
		"other-self-service@example.com",
		fixture.now.Add(-2*time.Hour),
	)
	if err := fixture.repository.CreateAccount(context.Background(), otherAccount, otherCredential); err != nil {
		t.Fatalf("create cross-owner account: %v", err)
	}
	otherSession := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		otherAccount.UserID,
		"ses_bbbbbbbbbbbbbbbb",
		localAuthenticationMethodPassword,
		"credential:"+otherCredential.CredentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	ctx := context.Background()
	staleActor := fixture.actor
	staleActor.AuthenticatedAt = fixture.now.Add(-localIdentitySelfServiceRecentAuthenticationAge - time.Second)
	if _, err := fixture.service.RevokeSession(ctx, staleActor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: target.SessionID, ExpectedVersion: target.RecordVersion, Confirmed: true, AuditRef: "audit:self-revoke",
	}); !errors.Is(err, errLocalIdentitySessionRecentAuthentication) {
		t.Fatalf("stale actor exact revoke mismatch: %v", err)
	}
	if stored := fixture.repository.sessions[target.SessionID]; stored.LifecycleState != localIdentityStateActive {
		t.Fatal("stale actor changed target session")
	}
	if _, err := fixture.service.RevokeSession(ctx, fixture.actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: otherSession.SessionID, ExpectedVersion: 1, Confirmed: true, AuditRef: "audit:cross-scope-revoke",
	}); !errors.Is(err, errLocalIdentitySessionScopeDenied) {
		t.Fatalf("cross-owner exact revoke mismatch: %v", err)
	}
	if fixture.repository.sessions[otherSession.SessionID].LifecycleState != localIdentityStateActive {
		t.Fatal("cross-owner exact revoke changed the target")
	}
	result, err := fixture.service.RevokeSession(ctx, fixture.actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: target.SessionID, ExpectedVersion: target.RecordVersion, Confirmed: true, AuditRef: "audit:self-revoke",
	})
	if err != nil || result.CurrentSessionRevoked || result.Session.EffectiveState != localIdentityStateRevoked ||
		result.Session.RecordVersion != target.RecordVersion+1 {
		t.Fatalf("exact session revoke mismatch: result=%#v err=%v", result, err)
	}
	if _, err := fixture.service.RevokeSession(ctx, fixture.actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: target.SessionID, ExpectedVersion: target.RecordVersion, Confirmed: true, AuditRef: "audit:stale-revoke",
	}); !errors.Is(err, errLocalIdentitySessionVersionConflict) {
		t.Fatalf("stale exact session revoke mismatch: %v", err)
	}

	current, _, err := fixture.repository.ReadWebSession(ctx, fixture.actor.CurrentSessionID, fixture.now)
	if err != nil {
		t.Fatalf("read current session before revoke: %v", err)
	}
	currentResult, err := fixture.service.RevokeSession(ctx, fixture.actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: current.SessionID, ExpectedVersion: current.RecordVersion, Confirmed: true, AuditRef: "audit:current-revoke",
	})
	if err != nil || !currentResult.CurrentSessionRevoked {
		t.Fatalf("current session revoke mismatch: result=%#v err=%v", currentResult, err)
	}
	if _, err := fixture.service.ListSessions(ctx, fixture.actor, LocalIdentitySelfServiceSessionListQuery{}); !errors.Is(err, errLocalIdentitySessionScopeDenied) {
		t.Fatalf("revoked current session remained an actor: %v", err)
	}
}

func TestMemoryLocalIdentitySelfServiceBulkRevokeIsScopedAndAtomic(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	activeIDs := []string{"ses_1111111111111111", "ses_2222222222222222", "ses_3333333333333333"}
	for index, sessionID := range activeIDs {
		method := localAuthenticationMethodPassword
		source := "credential:" + fixture.credentialID
		if index == len(activeIDs)-1 {
			method = localAuthenticationMethodOIDC
			source = "binding:xid_3333333333333333"
		}
		createLocalIdentitySelfServiceTestSession(
			t, fixture.repository, fixture.userID, sessionID, method, source,
			fixture.now.Add(-time.Hour), fixture.now.Add(time.Hour),
		)
	}
	expired := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_4444444444444444",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-2*time.Hour),
		fixture.now.Add(-time.Minute),
	)
	result, err := fixture.service.RevokeOtherSessions(context.Background(), fixture.actor, LocalIdentityRevokeOtherSessionsInput{
		Confirmed: true, AuditRef: "audit:bulk-revoke",
	})
	if err != nil || result.RevokedCount != len(activeIDs) {
		t.Fatalf("bulk session revoke mismatch: result=%#v err=%v", result, err)
	}
	if fixture.repository.sessions[fixture.actor.CurrentSessionID].LifecycleState != localIdentityStateActive ||
		fixture.repository.sessions[expired.SessionID].LifecycleState != localIdentityStateActive {
		t.Fatal("bulk revoke changed current or effective-expired session")
	}
	for _, sessionID := range activeIDs {
		if fixture.repository.sessions[sessionID].LifecycleState != localIdentityStateRevoked {
			t.Fatalf("bulk revoke left %s active", sessionID)
		}
	}

	atomic := newLocalIdentitySelfServiceTestFixtureWithIDs(
		t,
		"usr_cccccccccccccccc",
		"cred_cccccccccccccccc",
		"ses_cccccccccccccccc",
		localAuthenticationMethodPassword,
	)
	valid := createLocalIdentitySelfServiceTestSession(
		t,
		atomic.repository,
		atomic.userID,
		"ses_dddddddddddddddd",
		localAuthenticationMethodPassword,
		"credential:"+atomic.credentialID,
		atomic.now.Add(-time.Hour),
		atomic.now.Add(time.Hour),
	)
	invalid := createLocalIdentitySelfServiceTestSession(
		t,
		atomic.repository,
		atomic.userID,
		"ses_eeeeeeeeeeeeeeee",
		localAuthenticationMethodOIDC,
		"binding:xid_eeeeeeeeeeeeeeee",
		atomic.now.Add(-time.Hour),
		atomic.now.Add(time.Hour),
	)
	corrupted := atomic.repository.sessions[invalid.SessionID]
	corrupted.AuditRef = ""
	atomic.repository.sessions[invalid.SessionID] = corrupted
	if _, err := atomic.service.RevokeOtherSessions(context.Background(), atomic.actor, LocalIdentityRevokeOtherSessionsInput{
		Confirmed: true, AuditRef: "audit:atomic-bulk-revoke",
	}); !errors.Is(err, errLocalIdentitySessionBulkRevokeConflict) {
		t.Fatalf("invalid bulk target did not fail closed: %v", err)
	}
	if atomic.repository.sessions[valid.SessionID].LifecycleState != localIdentityStateActive ||
		atomic.repository.sessions[valid.SessionID].RecordVersion != valid.RecordVersion {
		t.Fatal("failed bulk revoke partially changed a valid target")
	}
}

type localIdentitySelfServiceTestFixture struct {
	repository   *memoryLocalIdentityRepository
	service      *localIdentitySelfServiceSecurityService
	actor        LocalIdentitySelfServiceActor
	now          time.Time
	userID       string
	credentialID string
	password     string
}

func newLocalIdentitySelfServiceTestFixture(
	t *testing.T,
	currentAuthenticationMethod string,
) localIdentitySelfServiceTestFixture {
	t.Helper()
	return newLocalIdentitySelfServiceTestFixtureWithIDs(
		t,
		"usr_aaaaaaaaaaaaaaaa",
		"cred_aaaaaaaaaaaaaaaa",
		"ses_aaaaaaaaaaaaaaaa",
		currentAuthenticationMethod,
	)
}

func newLocalIdentitySelfServiceTestFixtureWithIDs(
	t *testing.T,
	userID string,
	credentialID string,
	currentSessionID string,
	currentAuthenticationMethod string,
) localIdentitySelfServiceTestFixture {
	t.Helper()
	repository := newMemoryLocalIdentityRepository()
	password := "current password value"
	account, _ := localIdentityTestAccount(userID, credentialID, userID+"@example.com", localIdentitySelfServiceTestNow.Add(-2*time.Hour))
	credential, err := DeriveLocalCredential(
		password,
		credentialID,
		userID,
		localIdentitySelfServiceTestNow.Add(-2*time.Hour),
		"audit:self-service-account",
	)
	if err != nil {
		t.Fatalf("derive self-service fixture credential: %v", err)
	}
	if err := repository.CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create self-service fixture account: %v", err)
	}
	source := "credential:" + credentialID
	if currentAuthenticationMethod == localAuthenticationMethodOIDC {
		source = "binding:xid_aaaaaaaaaaaaaaaa"
	}
	createLocalIdentitySelfServiceTestSession(
		t,
		repository,
		userID,
		currentSessionID,
		currentAuthenticationMethod,
		source,
		localIdentitySelfServiceTestNow.Add(-time.Hour),
		localIdentitySelfServiceTestNow.Add(8*time.Hour),
	)
	service := newLocalIdentitySelfServiceSecurityService(repository)
	service.now = func() time.Time { return localIdentitySelfServiceTestNow }
	service.newID = func(string) (string, error) { return "cred_ffffffffffffffff", nil }
	return localIdentitySelfServiceTestFixture{
		repository: repository,
		service:    service,
		actor: LocalIdentitySelfServiceActor{
			UserID:           userID,
			CurrentSessionID: currentSessionID,
			AuthenticatedAt:  localIdentitySelfServiceTestNow.Add(-time.Minute),
		},
		now:          localIdentitySelfServiceTestNow,
		userID:       userID,
		credentialID: credentialID,
		password:     password,
	}
}

func createLocalIdentitySelfServiceTestSession(
	t *testing.T,
	repository localIdentityRepository,
	userID string,
	sessionID string,
	authenticationMethod string,
	authenticationSourceRef string,
	createdAt time.Time,
	expiresAt time.Time,
) WebSession {
	t.Helper()
	rawCredential := "self-service-session-credential-material-" + sessionID
	digest, err := DigestWebSessionCredential(rawCredential)
	if err != nil {
		t.Fatalf("digest self-service test session: %v", err)
	}
	session := WebSession{
		SchemaVersion:           localIdentitySchemaVersion,
		SessionID:               sessionID,
		UserID:                  userID,
		credentialDigest:        digest[:],
		AuthenticationMethod:    authenticationMethod,
		AuthenticationSourceRef: authenticationSourceRef,
		PolicyVersion:           localSessionPolicyVersion,
		LifecycleState:          localIdentityStateActive,
		RecordVersion:           1,
		CreatedAt:               createdAt.UTC(),
		UpdatedAt:               createdAt.UTC(),
		LastVerifiedAt:          createdAt.UTC(),
		ExpiresAt:               expiresAt.UTC(),
		AuditRef:                "audit:self-service-session",
	}
	if err := repository.CreateWebSession(context.Background(), session); err != nil {
		t.Fatalf("create self-service test session %s: %v", sessionID, err)
	}
	return session
}

func tamperLocalIdentitySelfServiceCursor(cursor string) string {
	if cursor == "" {
		return "A"
	}
	replacement := byte('A')
	if cursor[0] == replacement {
		replacement = 'B'
	}
	return string(replacement) + cursor[1:]
}
