package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryLocalIdentitySelfServiceCredentialRotationRevokesSourceBoundSessions(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	otherLocal := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	expiredLocal := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_2222222222222222",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-2*time.Hour),
		fixture.now.Add(-time.Minute),
	)
	oidc := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_3333333333333333",
		localAuthenticationMethodOIDC,
		"binding:xid_3333333333333333",
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	newPassword := "replacement password value"
	result, err := fixture.service.RotateCredential(context.Background(), fixture.actor, LocalIdentityRotateCredentialInput{
		CurrentPassword:        fixture.password,
		NewPassword:            newPassword,
		SessionImpactConfirmed: true,
		AuditRef:               "audit:credential-rotation",
	})
	if err != nil || result.PolicyVersion != localPasswordPolicyVersion || result.RevokedSessionCount != 3 ||
		!result.CurrentSessionRevoked {
		t.Fatalf("credential rotation result mismatch: result=%#v err=%v", result, err)
	}
	old := fixture.repository.credentials[fixture.credentialID]
	replacement := fixture.repository.credentials[fixture.repository.activeCredentialByUser[fixture.userID]]
	if old.LifecycleState != localIdentityStateSuperseded || old.RecordVersion != 2 ||
		replacement.CredentialID != "cred_ffffffffffffffff" || replacement.LifecycleState != localIdentityStateActive ||
		!VerifyLocalPassword(newPassword, replacement) || VerifyLocalPassword(fixture.password, replacement) {
		t.Fatalf("credential lifecycle mismatch: old=%#v replacement=%#v", old, replacement)
	}
	for _, sessionID := range []string{fixture.actor.CurrentSessionID, otherLocal.SessionID, expiredLocal.SessionID} {
		if session := fixture.repository.sessions[sessionID]; session.LifecycleState != localIdentityStateRevoked ||
			session.RecordVersion != 2 {
			t.Fatalf("source-bound session %s was not revoked: %#v", sessionID, session)
		}
	}
	if session := fixture.repository.sessions[oidc.SessionID]; session.LifecycleState != localIdentityStateActive || session.RecordVersion != 1 {
		t.Fatalf("OIDC session changed during local credential rotation: %#v", session)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal credential rotation result: %v", err)
	}
	for _, forbidden := range []string{
		fixture.credentialID,
		replacement.CredentialID,
		fixture.password,
		newPassword,
		"credential_id",
		"authentication_source_ref",
		"audit_ref",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("credential rotation result leaked %q: %s", forbidden, payload)
		}
	}
}

func TestMemoryLocalIdentitySelfServiceCredentialRotationRetainsCurrentOIDCSession(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodOIDC)
	local := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	result, err := fixture.service.RotateCredential(context.Background(), fixture.actor, LocalIdentityRotateCredentialInput{
		CurrentPassword:        fixture.password,
		NewPassword:            "replacement password value",
		SessionImpactConfirmed: true,
		AuditRef:               "audit:oidc-current-rotation",
	})
	if err != nil || result.CurrentSessionRevoked || result.RevokedSessionCount != 1 {
		t.Fatalf("OIDC-current credential rotation mismatch: result=%#v err=%v", result, err)
	}
	if fixture.repository.sessions[fixture.actor.CurrentSessionID].LifecycleState != localIdentityStateActive ||
		fixture.repository.sessions[local.SessionID].LifecycleState != localIdentityStateRevoked {
		t.Fatal("OIDC-current credential rotation changed the wrong session set")
	}
	if _, err := fixture.service.ListSessions(context.Background(), fixture.actor, LocalIdentitySelfServiceSessionListQuery{}); err != nil {
		t.Fatalf("retained OIDC current session could not continue self-service: %v", err)
	}
}

func TestMemoryLocalIdentitySelfServiceCredentialRotationRejectsInvalidProofReuseAndUnavailableCredential(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodPassword)
	ctx := context.Background()
	for _, testCase := range []struct {
		name     string
		input    LocalIdentityRotateCredentialInput
		expected error
	}{
		{
			name: "password policy",
			input: LocalIdentityRotateCredentialInput{
				CurrentPassword: fixture.password, NewPassword: "too-short", SessionImpactConfirmed: true,
				AuditRef: "audit:policy-reject",
			},
			expected: errLocalIdentityCredentialPolicyRejected,
		},
		{
			name: "current password",
			input: LocalIdentityRotateCredentialInput{
				CurrentPassword: "wrong current password", NewPassword: "replacement password value", SessionImpactConfirmed: true,
				AuditRef: "audit:proof-reject",
			},
			expected: errLocalIdentityCredentialCurrentInvalid,
		},
		{
			name: "password reuse",
			input: LocalIdentityRotateCredentialInput{
				CurrentPassword: fixture.password, NewPassword: fixture.password, SessionImpactConfirmed: true,
				AuditRef: "audit:reuse-reject",
			},
			expected: errLocalIdentityCredentialReuseDenied,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := fixture.service.RotateCredential(ctx, fixture.actor, testCase.input); !errors.Is(err, testCase.expected) {
				t.Fatalf("credential rotation failure mismatch: %v", err)
			}
			current := fixture.repository.credentials[fixture.credentialID]
			if current.LifecycleState != localIdentityStateActive || current.RecordVersion != 1 ||
				fixture.repository.sessions[fixture.actor.CurrentSessionID].LifecycleState != localIdentityStateActive ||
				fixture.repository.credentials["cred_ffffffffffffffff"].CredentialID != "" {
				t.Fatal("rejected credential rotation changed repository state")
			}
		})
	}

	unavailable := newLocalIdentitySelfServiceTestFixtureWithIDs(
		t,
		"usr_bbbbbbbbbbbbbbbb",
		"cred_bbbbbbbbbbbbbbbb",
		"ses_bbbbbbbbbbbbbbbb",
		localAuthenticationMethodOIDC,
	)
	delete(unavailable.repository.credentials, unavailable.credentialID)
	delete(unavailable.repository.activeCredentialByUser, unavailable.userID)
	if _, err := unavailable.service.RotateCredential(ctx, unavailable.actor, LocalIdentityRotateCredentialInput{
		CurrentPassword:        unavailable.password,
		NewPassword:            "replacement password value",
		SessionImpactConfirmed: true,
		AuditRef:               "audit:credential-unavailable",
	}); !errors.Is(err, errLocalIdentityCredentialUnavailable) {
		t.Fatalf("missing local credential did not fail closed: %v", err)
	}
}

func TestMemoryLocalIdentitySelfServiceCredentialRotationHasZeroPartialWrites(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodOIDC)
	valid := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	invalid := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_2222222222222222",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	corrupted := fixture.repository.sessions[invalid.SessionID]
	corrupted.AuditRef = ""
	fixture.repository.sessions[invalid.SessionID] = corrupted
	if _, err := fixture.service.RotateCredential(context.Background(), fixture.actor, LocalIdentityRotateCredentialInput{
		CurrentPassword:        fixture.password,
		NewPassword:            "replacement password value",
		SessionImpactConfirmed: true,
		AuditRef:               "audit:atomic-rotation",
	}); !errors.Is(err, errLocalIdentityCredentialRotationConflict) {
		t.Fatalf("invalid source-bound session did not fail rotation: %v", err)
	}
	old := fixture.repository.credentials[fixture.credentialID]
	if old.LifecycleState != localIdentityStateActive || old.RecordVersion != 1 ||
		fixture.repository.activeCredentialByUser[fixture.userID] != fixture.credentialID ||
		fixture.repository.credentials["cred_ffffffffffffffff"].CredentialID != "" ||
		fixture.repository.sessions[valid.SessionID].LifecycleState != localIdentityStateActive ||
		fixture.repository.sessions[valid.SessionID].RecordVersion != valid.RecordVersion {
		t.Fatal("failed credential rotation committed a partial write")
	}
}

func TestMemoryLocalIdentitySelfServiceConcurrentCredentialRotationHasSingleWinner(t *testing.T) {
	fixture := newLocalIdentitySelfServiceTestFixture(t, localAuthenticationMethodOIDC)
	local := createLocalIdentitySelfServiceTestSession(
		t,
		fixture.repository,
		fixture.userID,
		"ses_1111111111111111",
		localAuthenticationMethodPassword,
		"credential:"+fixture.credentialID,
		fixture.now.Add(-time.Hour),
		fixture.now.Add(time.Hour),
	)
	var identifierSequence atomic.Uint64
	fixture.service.newID = func(prefix string) (string, error) {
		return fmt.Sprintf("%s%016x", prefix, identifierSequence.Add(1)+0x100), nil
	}
	const contenders = 4
	results := make(chan error, contenders)
	var wait sync.WaitGroup
	for index := 0; index < contenders; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			_, err := fixture.service.RotateCredential(context.Background(), fixture.actor, LocalIdentityRotateCredentialInput{
				CurrentPassword:        fixture.password,
				NewPassword:            fmt.Sprintf("replacement password value %d", index),
				SessionImpactConfirmed: true,
				AuditRef:               fmt.Sprintf("audit:concurrent-rotation:%d", index),
			})
			results <- err
		}(index)
	}
	wait.Wait()
	close(results)
	winners := 0
	losers := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, errLocalIdentityCredentialCurrentInvalid), errors.Is(err, errLocalIdentityCredentialRotationConflict):
			losers++
		default:
			t.Fatalf("unexpected concurrent credential rotation result: %v", err)
		}
	}
	if winners != 1 || losers != contenders-1 {
		t.Fatalf("concurrent credential rotation mismatch: winners=%d losers=%d", winners, losers)
	}
	activeCredentials := 0
	for _, credential := range fixture.repository.credentials {
		if credential.UserID == fixture.userID && credential.LifecycleState == localIdentityStateActive {
			activeCredentials++
		}
	}
	if activeCredentials != 1 || fixture.repository.credentials[fixture.credentialID].LifecycleState != localIdentityStateSuperseded ||
		fixture.repository.sessions[local.SessionID].LifecycleState != localIdentityStateRevoked ||
		fixture.repository.sessions[local.SessionID].RecordVersion != 2 ||
		fixture.repository.sessions[fixture.actor.CurrentSessionID].LifecycleState != localIdentityStateActive {
		t.Fatalf("concurrent credential rotation repository mismatch: active_credentials=%d", activeCredentials)
	}
}
