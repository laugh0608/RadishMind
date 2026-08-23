package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

var localIdentityTestNow = time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

func TestLocalPasswordDerivationVerificationAndSanitizedJSON(t *testing.T) {
	credential, err := DeriveLocalCredential(
		"correct horse battery staple", "cred_aaaaaaaaaaaaaaaa", "usr_aaaaaaaaaaaaaaaa",
		localIdentityTestNow, "audit:credential-create",
	)
	if err != nil {
		t.Fatalf("derive local credential: %v", err)
	}
	if credential.Algorithm != localPasswordAlgorithmPBKDF2SHA256 || credential.Iterations != localPasswordIterations ||
		len(credential.salt) != localPasswordSaltLength || len(credential.derivedKey) != localPasswordKeyLength {
		t.Fatalf("credential parameters drifted: %#v", credential)
	}
	if !VerifyLocalPassword("correct horse battery staple", credential) || VerifyLocalPassword("wrong password", credential) {
		t.Fatal("password verification boundary failed")
	}
	payload, err := json.Marshal(credential)
	if err != nil {
		t.Fatalf("marshal sanitized credential: %v", err)
	}
	for _, forbidden := range [][]byte{
		[]byte("correct horse battery staple"), credential.salt, credential.derivedKey, []byte("derived_key"), []byte("salt"),
	} {
		if len(forbidden) > 0 && bytes.Contains(payload, forbidden) {
			t.Fatalf("sanitized credential JSON leaked material: %s", payload)
		}
	}
	session := localIdentityTestSession(t, credential.UserID)
	sessionPayload, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("marshal sanitized Web session: %v", err)
	}
	if bytes.Contains(sessionPayload, session.credentialDigest) || bytes.Contains(sessionPayload, []byte("credential_digest")) {
		t.Fatalf("sanitized Web session JSON leaked credential digest: %s", sessionPayload)
	}
}

func TestLocalIdentityNormalizationBoundaries(t *testing.T) {
	identifier, err := NormalizeLocalLoginIdentifier("  Alice.Example+Dev@Example.COM ")
	if err != nil || identifier != "alice.example+dev@example.com" {
		t.Fatalf("normalize login identifier: value=%q err=%v", identifier, err)
	}
	issuer, err := NormalizeExternalIssuer("HTTPS://RADISH.EXAMPLE.COM/oidc/")
	if err != nil || issuer != "https://radish.example.com/oidc" {
		t.Fatalf("normalize issuer: value=%q err=%v", issuer, err)
	}
	loopback, err := NormalizeExternalIssuer("http://127.0.0.1:8080/oidc/")
	if err != nil || loopback != "http://127.0.0.1:8080/oidc" {
		t.Fatalf("normalize loopback issuer: value=%q err=%v", loopback, err)
	}
	for _, invalid := range []string{
		"HTTP://radish.example.com/oidc", "https://radish.example.com/oidc?tenant=other", "https://user@radish.example.com/oidc",
	} {
		if _, err := NormalizeExternalIssuer(invalid); !errors.Is(err, errLocalIdentityContractMismatch) {
			t.Fatalf("invalid issuer %q was accepted: %v", invalid, err)
		}
	}
	actorRef, err := LocalUserActorRef("usr_aaaaaaaaaaaaaaaa")
	if err != nil || actorRef != "user:usr_aaaaaaaaaaaaaaaa" {
		t.Fatalf("local actor projection mismatch: value=%q err=%v", actorRef, err)
	}
}

func TestMemoryLocalIdentityRepositoryContract(t *testing.T) {
	runLocalIdentityRepositoryContract(t, newMemoryLocalIdentityRepository())
}

func TestSQLiteLocalIdentityRepositoryContractAndRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity", "radishmind.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite identity runtime: %v", err)
	}
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	runLocalIdentityRepositoryContract(t, repository)
	pending := localIdentityTestOIDCTransaction("oat_restartaaaaaaaaaaa", "restart-state-with-more-than-thirty-two-characters", localIdentityTestNow.Add(3*time.Hour))
	if err := repository.CreateOIDCAuthorizationTransaction(context.Background(), pending); err != nil {
		t.Fatalf("create restart OIDC authorization transaction: %v", err)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite identity runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("restart SQLite identity runtime: %v", err)
	}
	t.Cleanup(func() { _ = restarted.Close() })
	restored := newSQLiteLocalIdentityRepository(restarted.DB())
	account, err := restored.ReadAccount(context.Background(), "usr_aaaaaaaaaaaaaaaa")
	if err != nil || account.LifecycleState != localIdentityStateDisabled || account.RecordVersion != 2 {
		t.Fatalf("restored disabled account mismatch: account=%#v err=%v", account, err)
	}
	if _, _, err := restored.ResolveWebSession(context.Background(), localIdentityTestSessionDigest(t), localIdentityTestNow.Add(time.Hour)); !errors.Is(err, errLocalIdentitySessionInvalid) {
		t.Fatalf("restored revoked session must remain invalid: %v", err)
	}
	restartStateDigest, _ := localIdentityOIDCStateDigest("restart-state-with-more-than-thirty-two-characters")
	if restoredTransaction, err := restored.ConsumeOIDCAuthorizationTransaction(context.Background(), restartStateDigest, localIdentityTestNow.Add(3*time.Hour+time.Minute)); err != nil ||
		restoredTransaction.TransactionID != pending.TransactionID || restoredTransaction.codeVerifier == "" {
		t.Fatalf("restored OIDC authorization transaction mismatch: transaction=%#v err=%v", restoredTransaction, err)
	}
}

func TestLocalIdentityRepositoryFactoryFailsClosed(t *testing.T) {
	if _, _, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{Mode: "future_backend"}); err == nil || err.Error() != "unsupported local identity store mode" {
		t.Fatalf("unsupported identity store mode did not fail closed: %v", err)
	}
	if _, _, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{Mode: localIdentityStoreModeSQLiteDev}); err == nil || err.Error() != "sqlite_dev local identity requires the shared SQLite runtime" {
		t.Fatalf("SQLite identity without runtime did not fail closed: %v", err)
	}
	if _, _, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{Mode: localIdentityStoreModePostgresDevTest}); err == nil || err.Error() != "postgres_dev_test local identity database URL is missing" {
		t.Fatalf("PostgreSQL identity without URL did not fail closed: %v", err)
	}

	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"),
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite identity factory runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	repository, closeRepository, err := newLocalIdentityRepositoryFromOptions(localIdentityStoreOptions{
		Mode: localIdentityStoreModeSQLiteDev, SQLiteRuntime: runtime,
	})
	if err != nil {
		t.Fatalf("select SQLite identity repository: %v", err)
	}
	defer closeRepository()
	if _, ok := repository.(*sqliteLocalIdentityRepository); !ok {
		t.Fatalf("unexpected SQLite identity repository: %T", repository)
	}
}

func TestMemoryLocalIdentityConcurrentIdentifierAndBindingSingleWinner(t *testing.T) {
	repository := newMemoryLocalIdentityRepository()
	runConcurrentLocalIdentityAccountCreate(t, repository, 12)
	runConcurrentLocalIdentityBindingCreate(t, repository)
}

func TestSQLiteLocalIdentityConcurrentIdentifierSingleWinner(t *testing.T) {
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: filepath.Join(t.TempDir(), "radishmind.db"),
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open concurrent SQLite identity runtime: %v", err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	runConcurrentLocalIdentityAccountCreate(t, repository, 6)
	runConcurrentLocalIdentityBindingCreate(t, repository)
}

func runConcurrentLocalIdentityAccountCreate(t *testing.T, repository localIdentityRepository, contenders int) {
	t.Helper()
	results := make(chan error, contenders)
	var wait sync.WaitGroup
	for index := 0; index < contenders; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			userID := fmt.Sprintf("usr_%016x", index+1)
			credentialID := fmt.Sprintf("cred_%016x", index+1)
			account, credential := localIdentityTestAccount(userID, credentialID, "shared@example.com", localIdentityTestNow)
			results <- repository.CreateAccount(context.Background(), account, credential)
		}(index)
	}
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, errLocalIdentityIdentifierConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent account result: %v", err)
		}
	}
	if winners != 1 || conflicts != contenders-1 {
		t.Fatalf("concurrent account single-winner mismatch: winners=%d conflicts=%d", winners, conflicts)
	}
}

func runConcurrentLocalIdentityBindingCreate(t *testing.T, repository localIdentityRepository) {
	t.Helper()
	ctx := context.Background()
	first, firstCredential := localIdentityTestAccount(
		"usr_iiiiiiiiiiiiiiii", "cred_iiiiiiiiiiiiiiii", "concurrent-binding-one@example.com", localIdentityTestNow,
	)
	second, secondCredential := localIdentityTestAccount(
		"usr_jjjjjjjjjjjjjjjj", "cred_jjjjjjjjjjjjjjjj", "concurrent-binding-two@example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(ctx, first, firstCredential); err != nil {
		t.Fatalf("create first concurrent binding account: %v", err)
	}
	if err := repository.CreateAccount(ctx, second, secondCredential); err != nil {
		t.Fatalf("create second concurrent binding account: %v", err)
	}
	bindings := []ExternalIdentityBinding{
		localIdentityTestBinding("xid_iiiiiiiiiiiiiiii", first.UserID, "https://radish.example.com/concurrent", "shared-concurrent-subject"),
		localIdentityTestBinding("xid_jjjjjjjjjjjjjjjj", second.UserID, "https://radish.example.com/concurrent", "shared-concurrent-subject"),
	}
	start := make(chan struct{})
	results := make(chan error, len(bindings))
	var wait sync.WaitGroup
	for _, binding := range bindings {
		wait.Add(1)
		go func(candidate ExternalIdentityBinding) {
			defer wait.Done()
			<-start
			results <- repository.BindExternalIdentity(ctx, candidate)
		}(binding)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, errLocalIdentityExternalConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent external binding result: %v", err)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("concurrent external binding single-winner mismatch: winners=%d conflicts=%d", winners, conflicts)
	}
}

type unavailableLocalIdentityRepository struct {
	localIdentityRepository
}

func (unavailableLocalIdentityRepository) AuthorizeWorkspace(
	context.Context,
	string,
	string,
	string,
	[]string,
	time.Time,
) (LocalWorkspaceAuthorization, error) {
	return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
}

func TestLocalWorkspaceMembershipProviderDoesNotFallback(t *testing.T) {
	provider := newLocalWorkspaceMembershipProvider(unavailableLocalIdentityRepository{})
	decision := provider.AuthorizeWorkspace(context.Background(), WorkspaceMembershipRequest{
		Auth: controlPlaneReadAuthContext{
			TenantBinding: "tenant_demo", SubjectBinding: "user:usr_aaaaaaaaaaaaaaaa",
			ResourceBinding: ControlPlaneResourceBinding{TenantRef: "tenant_demo", TenantVerified: true},
		},
		ActiveWorkspaceID: "workspace_demo", RequiredPermissions: []string{"applications:read"},
	})
	if decision.HTTPStatus != 503 || decision.FailureCode != "workspace_membership_unavailable" ||
		decision.Binding.WorkspaceMembershipVerified {
		t.Fatalf("unavailable local membership repository must fail without fallback: %#v", decision)
	}
}

func runLocalIdentityRepositoryContract(t *testing.T, repository localIdentityRepository) {
	t.Helper()
	ctx := context.Background()
	account, credential := localIdentityTestAccount(
		"usr_aaaaaaaaaaaaaaaa", "cred_aaaaaaaaaaaaaaaa", "Alice@Example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(ctx, account, credential); err != nil {
		t.Fatalf("create local account: %v", err)
	}
	read, err := repository.FindAccountByLoginIdentifier(ctx, " alice@example.COM ")
	if err != nil || read.UserID != account.UserID || read.NormalizedLoginIdentifier != "alice@example.com" {
		t.Fatalf("read normalized local account: account=%#v err=%v", read, err)
	}
	duplicate, duplicateCredential := localIdentityTestAccount(
		"usr_bbbbbbbbbbbbbbbb", "cred_bbbbbbbbbbbbbbbb", "alice@example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(ctx, duplicate, duplicateCredential); !errors.Is(err, errLocalIdentityIdentifierConflict) {
		t.Fatalf("duplicate identifier must conflict: %v", err)
	}
	second, secondCredential := localIdentityTestAccount(
		"usr_bbbbbbbbbbbbbbbb", "cred_bbbbbbbbbbbbbbbb", "bob@example.com", localIdentityTestNow,
	)
	if err := repository.CreateAccount(ctx, second, secondCredential); err != nil {
		t.Fatalf("create second local account: %v", err)
	}
	registered, registeredCredential := localIdentityTestAccount(
		"usr_cccccccccccccccc", "cred_dddddddddddddddd", "registered@example.com", localIdentityTestNow,
	)
	registeredDigest, digestErr := DigestWebSessionCredential("a-distinct-registration-session-credential")
	if digestErr != nil {
		t.Fatalf("digest atomic registration session: %v", digestErr)
	}
	registeredSession := localIdentityTestSession(t, registered.UserID)
	registeredSession.SessionID = "ses_cccccccccccccccc"
	registeredSession.credentialDigest = registeredDigest[:]
	registeredSession.AuthenticationSourceRef = "credential:" + registeredCredential.CredentialID
	if err := repository.CreateAccountAndWebSession(ctx, registered, registeredCredential, registeredSession); err != nil {
		t.Fatalf("create atomic account and session: %v", err)
	}
	if resolvedSession, resolvedAccount, err := repository.ResolveWebSession(ctx, registeredDigest, localIdentityTestNow); err != nil ||
		resolvedSession.SessionID != registeredSession.SessionID || resolvedAccount.UserID != registered.UserID {
		t.Fatalf("resolve atomic registration: session=%#v account=%#v err=%v", resolvedSession, resolvedAccount, err)
	}

	oidcTransaction := localIdentityTestOIDCTransaction(
		"oat_aaaaaaaaaaaaaaaa", "authorization-state-with-more-than-thirty-two-characters", localIdentityTestNow,
	)
	if err := repository.CreateOIDCAuthorizationTransaction(ctx, oidcTransaction); err != nil {
		t.Fatalf("create OIDC authorization transaction: %v", err)
	}
	stateDigest, _ := localIdentityOIDCStateDigest("authorization-state-with-more-than-thirty-two-characters")
	runConcurrentLocalIdentitySingleWinner(t, "OIDC authorization transaction consume", errLocalIdentityOIDCStateInvalid, []func() error{
		func() error {
			_, consumeErr := repository.ConsumeOIDCAuthorizationTransaction(ctx, stateDigest, localIdentityTestNow.Add(time.Minute))
			return consumeErr
		},
		func() error {
			_, consumeErr := repository.ConsumeOIDCAuthorizationTransaction(ctx, stateDigest, localIdentityTestNow.Add(time.Minute))
			return consumeErr
		},
	})
	expiredTransaction := localIdentityTestOIDCTransaction(
		"oat_bbbbbbbbbbbbbbbb", "expired-authorization-state-with-more-than-thirty-two-characters", localIdentityTestNow,
	)
	expiredTransaction.ExpiresAt = localIdentityTestNow.Add(time.Minute)
	if err := repository.CreateOIDCAuthorizationTransaction(ctx, expiredTransaction); err != nil {
		t.Fatalf("create expiring OIDC authorization transaction: %v", err)
	}
	expiredStateDigest, _ := localIdentityOIDCStateDigest("expired-authorization-state-with-more-than-thirty-two-characters")
	if _, err := repository.ConsumeOIDCAuthorizationTransaction(ctx, expiredStateDigest, localIdentityTestNow.Add(2*time.Minute)); !errors.Is(err, errLocalIdentityOIDCStateExpired) {
		t.Fatalf("expired OIDC authorization transaction must fail closed: %v", err)
	}

	oidcAccount := UserAccount{
		SchemaVersion: localIdentitySchemaVersion, UserID: "usr_oidconlyaaaaaaaa", LoginIdentifier: "oidc.only.account",
		NormalizedLoginIdentifier: "oidc.only.account", DisplayName: "OIDC only", LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: localIdentityTestNow, UpdatedAt: localIdentityTestNow, AuditRef: "audit:oidc-account",
	}
	oidcBinding := localIdentityTestBinding("xid_oidconlyaaaaaaaa", oidcAccount.UserID, "https://radish.example.com/oidc", "oidc-only-subject")
	oidcSession := localIdentityTestSession(t, oidcAccount.UserID)
	oidcSession.SessionID = "ses_oidconlyaaaaaaaa"
	oidcSession.AuthenticationMethod = localAuthenticationMethodOIDC
	oidcSession.AuthenticationSourceRef = "external:" + oidcBinding.BindingID
	oidcDigest, digestErr := DigestWebSessionCredential("oidc-only-session-credential-with-enough-entropy")
	if digestErr != nil {
		t.Fatalf("digest OIDC-only session: %v", digestErr)
	}
	oidcSession.credentialDigest = oidcDigest[:]
	if err := repository.CreateOIDCAccountAndWebSession(ctx, oidcAccount, oidcBinding, oidcSession); err != nil {
		t.Fatalf("create atomic OIDC account, binding, and session: %v", err)
	}
	if _, _, err := repository.ResolveWebSession(ctx, oidcDigest, localIdentityTestNow.Add(time.Minute)); err != nil {
		t.Fatalf("resolve OIDC-only session: %v", err)
	}
	if _, err := repository.RevokeExternalIdentity(ctx, oidcBinding.BindingID, 1, localIdentityTestNow.Add(time.Minute), "audit:oidc-orphan-denied"); !errors.Is(err, errLocalIdentityLastLoginMethodRemoval) {
		t.Fatalf("OIDC-only last login method removal must be denied: %v", err)
	}

	replacement := localIdentityTestCredential("cred_cccccccccccccccc", account.UserID, localIdentityTestNow.Add(time.Minute))
	if err := repository.ReplaceCredential(ctx, account.UserID, credential.CredentialID, 2, replacement); !errors.Is(err, errLocalIdentityVersionConflict) {
		t.Fatalf("stale credential replacement must conflict: %v", err)
	}
	if err := repository.ReplaceCredential(ctx, account.UserID, credential.CredentialID, 1, replacement); err != nil {
		t.Fatalf("replace local credential: %v", err)
	}
	wrongExpectedCredential := localIdentityTestCredential("cred_ffffffffffffffff", account.UserID, localIdentityTestNow.Add(2*time.Minute))
	if err := repository.ReplaceCredential(ctx, account.UserID, credential.CredentialID, 1, wrongExpectedCredential); !errors.Is(err, errLocalIdentityVersionConflict) {
		t.Fatalf("superseded expected credential must conflict: %v", err)
	}
	activeCredential, err := repository.ReadActiveCredential(ctx, account.UserID)
	if err != nil || activeCredential.CredentialID != replacement.CredentialID {
		t.Fatalf("active credential mismatch: credential=%#v err=%v", activeCredential, err)
	}
	firstConcurrentCredential := localIdentityTestCredential(
		"cred_gggggggggggggggg", account.UserID, localIdentityTestNow.Add(2*time.Minute),
	)
	secondConcurrentCredential := localIdentityTestCredential(
		"cred_hhhhhhhhhhhhhhhh", account.UserID, localIdentityTestNow.Add(3*time.Minute),
	)
	runConcurrentLocalIdentitySingleWinner(t, "credential replacement", errLocalIdentityVersionConflict, []func() error{
		func() error {
			return repository.ReplaceCredential(ctx, account.UserID, replacement.CredentialID, 1, firstConcurrentCredential)
		},
		func() error {
			return repository.ReplaceCredential(ctx, account.UserID, replacement.CredentialID, 1, secondConcurrentCredential)
		},
	})
	activeCredential, err = repository.ReadActiveCredential(ctx, account.UserID)
	if err != nil || activeCredential.CredentialID != firstConcurrentCredential.CredentialID &&
		activeCredential.CredentialID != secondConcurrentCredential.CredentialID {
		t.Fatalf("concurrent active credential mismatch: credential=%#v err=%v", activeCredential, err)
	}

	binding := localIdentityTestBinding("xid_aaaaaaaaaaaaaaaa", account.UserID, "https://radish.example.com/oidc", "shared-subject")
	if err := repository.BindExternalIdentity(ctx, binding); err != nil {
		t.Fatalf("bind external identity: %v", err)
	}
	conflictingBinding := localIdentityTestBinding("xid_bbbbbbbbbbbbbbbb", second.UserID, binding.Issuer, binding.Subject)
	if err := repository.BindExternalIdentity(ctx, conflictingBinding); !errors.Is(err, errLocalIdentityExternalConflict) {
		t.Fatalf("same issuer and subject must conflict: %v", err)
	}
	differentIssuerBinding := localIdentityTestBinding(
		"xid_cccccccccccccccc", second.UserID, "https://accounts.radish.example.com/oidc", binding.Subject,
	)
	if err := repository.BindExternalIdentity(ctx, differentIssuerBinding); err != nil {
		t.Fatalf("different issuer with same subject must remain isolated: %v", err)
	}
	resolvedBinding, err := repository.ResolveExternalIdentity(ctx, binding.Issuer, binding.Subject)
	if err != nil || resolvedBinding.UserID != account.UserID {
		t.Fatalf("resolve external identity: binding=%#v err=%v", resolvedBinding, err)
	}
	runConcurrentLocalIdentitySingleWinner(t, "external identity revocation", errLocalIdentityVersionConflict, []func() error{
		func() error {
			_, revokeErr := repository.RevokeExternalIdentity(
				ctx, differentIssuerBinding.BindingID, 1, localIdentityTestNow.Add(time.Hour), "audit:binding-revoke-one",
			)
			return revokeErr
		},
		func() error {
			_, revokeErr := repository.RevokeExternalIdentity(
				ctx, differentIssuerBinding.BindingID, 1, localIdentityTestNow.Add(time.Hour), "audit:binding-revoke-two",
			)
			return revokeErr
		},
	})
	if _, err := repository.ResolveExternalIdentity(ctx, differentIssuerBinding.Issuer, differentIssuerBinding.Subject); !errors.Is(err, errLocalIdentityNotFound) {
		t.Fatalf("revoked external identity must not resolve: %v", err)
	}

	session := localIdentityTestSession(t, account.UserID)
	session.AuthenticationSourceRef = "credential:" + activeCredential.CredentialID
	if err := repository.CreateWebSession(ctx, session); err != nil {
		t.Fatalf("create Web session: %v", err)
	}
	resolvedSession, resolvedAccount, err := repository.ResolveWebSession(ctx, localIdentityTestSessionDigest(t), localIdentityTestNow.Add(time.Hour))
	if err != nil || resolvedSession.SessionID != session.SessionID || resolvedAccount.UserID != account.UserID {
		t.Fatalf("resolve Web session: session=%#v account=%#v err=%v", resolvedSession, resolvedAccount, err)
	}
	if _, _, err := repository.ResolveWebSession(ctx, localIdentityTestSessionDigest(t), session.ExpiresAt); !errors.Is(err, errLocalIdentitySessionExpired) {
		t.Fatalf("expired Web session must fail closed: %v", err)
	}
	secondarySession := localIdentityTestSession(t, second.UserID)
	secondarySession.SessionID = "ses_bbbbbbbbbbbbbbbb"
	secondaryDigest, digestErr := DigestWebSessionCredential("secondary-session-credential-with-at-least-thirty-two-bytes")
	if digestErr != nil {
		t.Fatalf("digest secondary Web session credential: %v", digestErr)
	}
	secondarySession.credentialDigest = secondaryDigest[:]
	secondarySession.AuthenticationSourceRef = "credential:" + secondCredential.CredentialID
	if err := repository.CreateWebSession(ctx, secondarySession); err != nil {
		t.Fatalf("create secondary Web session: %v", err)
	}
	runConcurrentLocalIdentitySingleWinner(t, "Web session revocation", errLocalIdentityVersionConflict, []func() error{
		func() error {
			_, revokeErr := repository.RevokeWebSession(
				ctx, secondarySession.SessionID, 1, localIdentityTestNow.Add(90*time.Minute), "audit:session-revoke-one",
			)
			return revokeErr
		},
		func() error {
			_, revokeErr := repository.RevokeWebSession(
				ctx, secondarySession.SessionID, 1, localIdentityTestNow.Add(90*time.Minute), "audit:session-revoke-two",
			)
			return revokeErr
		},
	})
	if _, _, err := repository.ResolveWebSession(ctx, secondaryDigest, localIdentityTestNow.Add(90*time.Minute)); !errors.Is(err, errLocalIdentitySessionInvalid) {
		t.Fatalf("revoked secondary Web session must remain invalid: %v", err)
	}

	membership := localIdentityTestMembership(account.UserID)
	if err := repository.CreateWorkspaceMembership(ctx, membership); err != nil {
		t.Fatalf("create workspace membership: %v", err)
	}
	role := localIdentityTestRoleAssignment(account.UserID)
	if err := repository.CreateRoleAssignment(ctx, role); err != nil {
		t.Fatalf("create local role assignment: %v", err)
	}
	profile, err := repository.ReadAccountAccessProfile(ctx, account.UserID)
	if err != nil || profile.Account.UserID != account.UserID || !profile.HasActiveLocalCredential ||
		len(profile.ExternalIdentities) != 1 || profile.ExternalIdentities[0].BindingID != binding.BindingID ||
		len(profile.RoleAssignments) != 1 || profile.RoleAssignments[0].AssignmentID != role.AssignmentID ||
		len(profile.WorkspaceMemberships) != 1 || profile.WorkspaceMemberships[0].MembershipID != membership.MembershipID {
		t.Fatalf("account access profile mismatch: profile=%#v err=%v", profile, err)
	}
	if _, err := repository.ReadAccountAccessProfile(ctx, "usr_missingaaaaaaaaa"); !errors.Is(err, errLocalIdentityNotFound) {
		t.Fatalf("missing account access profile must remain not found: %v", err)
	}
	secondaryRole := role
	secondaryRole.AssignmentID = "rla_bbbbbbbbbbbbbbbb"
	secondaryRole.UserID = second.UserID
	secondaryRole.RoleKey = "workspace_auditor"
	secondaryRole.PermissionGrants = []string{"runs:read"}
	if err := repository.CreateRoleAssignment(ctx, secondaryRole); err != nil {
		t.Fatalf("create secondary local role assignment: %v", err)
	}
	runConcurrentLocalIdentitySingleWinner(t, "role revocation", errLocalIdentityVersionConflict, []func() error{
		func() error {
			_, revokeErr := repository.RevokeRoleAssignment(
				ctx, secondaryRole.AssignmentID, 1, localIdentityTestNow.Add(90*time.Minute), "audit:role-revoke-one",
			)
			return revokeErr
		},
		func() error {
			_, revokeErr := repository.RevokeRoleAssignment(
				ctx, secondaryRole.AssignmentID, 1, localIdentityTestNow.Add(90*time.Minute), "audit:role-revoke-two",
			)
			return revokeErr
		},
	})
	authorization, err := repository.AuthorizeWorkspace(
		ctx, account.UserID, membership.TenantRef, membership.WorkspaceID, []string{"applications:read"}, localIdentityTestNow,
	)
	if err != nil || !slicesEqual(authorization.PermissionGrants, []string{"applications:read", "runs:read"}) {
		t.Fatalf("authorize local workspace: authorization=%#v err=%v", authorization, err)
	}
	if _, err := repository.AuthorizeWorkspace(
		ctx, account.UserID, membership.TenantRef, membership.WorkspaceID, []string{"api_keys:write"}, localIdentityTestNow,
	); !errors.Is(err, errLocalIdentityPermissionDenied) {
		t.Fatalf("missing local permission must fail closed: %v", err)
	}

	provider := newLocalWorkspaceMembershipProvider(repository).(*localWorkspaceMembershipProvider)
	provider.now = func() time.Time { return localIdentityTestNow }
	decision := provider.AuthorizeWorkspace(ctx, WorkspaceMembershipRequest{
		Auth: controlPlaneReadAuthContext{
			TenantBinding: membership.TenantRef, SubjectBinding: "user:" + account.UserID,
			ResourceBinding: ControlPlaneResourceBinding{TenantRef: membership.TenantRef, TenantVerified: true},
		},
		ActiveWorkspaceID: membership.WorkspaceID, RequiredPermissions: []string{"applications:read"},
	})
	if decision.HTTPStatus != 200 || !decision.Binding.WorkspaceMembershipVerified ||
		decision.Binding.WorkspaceSourceRef != "membership:local:"+membership.MembershipID {
		t.Fatalf("local membership adapter decision mismatch: %#v", decision)
	}

	runConcurrentLocalIdentitySingleWinner(t, "workspace membership revocation", errLocalIdentityVersionConflict, []func() error{
		func() error {
			_, revokeErr := repository.RevokeWorkspaceMembership(
				ctx, membership.MembershipID, membership.RecordVersion, localIdentityTestNow.Add(2*time.Hour), "audit:membership-revoke-one",
			)
			return revokeErr
		},
		func() error {
			_, revokeErr := repository.RevokeWorkspaceMembership(
				ctx, membership.MembershipID, membership.RecordVersion, localIdentityTestNow.Add(2*time.Hour), "audit:membership-revoke-two",
			)
			return revokeErr
		},
	})
	if _, err := repository.AuthorizeWorkspace(
		ctx, account.UserID, membership.TenantRef, membership.WorkspaceID, []string{"applications:read"}, localIdentityTestNow.Add(2*time.Hour),
	); !errors.Is(err, errLocalIdentityMembershipDenied) {
		t.Fatalf("revoked membership must fail immediately: %v", err)
	}

	if _, err := repository.DisableAccount(ctx, account.UserID, 2, localIdentityTestNow.Add(3*time.Hour), "audit:stale-disable"); !errors.Is(err, errLocalIdentityVersionConflict) {
		t.Fatalf("stale account disable must conflict: %v", err)
	}
	runConcurrentLocalIdentitySingleWinner(t, "account disable", errLocalIdentityVersionConflict, []func() error{
		func() error {
			_, disableErr := repository.DisableAccount(
				ctx, account.UserID, 1, localIdentityTestNow.Add(3*time.Hour), "audit:disable-one",
			)
			return disableErr
		},
		func() error {
			_, disableErr := repository.DisableAccount(
				ctx, account.UserID, 1, localIdentityTestNow.Add(3*time.Hour), "audit:disable-two",
			)
			return disableErr
		},
	})
	disabled, err := repository.ReadAccount(ctx, account.UserID)
	if err != nil || disabled.LifecycleState != localIdentityStateDisabled || disabled.RecordVersion != 2 {
		t.Fatalf("disable local account: account=%#v err=%v", disabled, err)
	}
	if _, _, err := repository.ResolveWebSession(ctx, localIdentityTestSessionDigest(t), localIdentityTestNow.Add(3*time.Hour)); !errors.Is(err, errLocalIdentitySessionInvalid) {
		t.Fatalf("account disable must revoke active sessions: %v", err)
	}
}

func localIdentityTestOIDCTransaction(transactionID string, rawState string, now time.Time) LocalIdentityOIDCAuthorizationTransaction {
	stateDigest := sha256.Sum256([]byte(rawState))
	nonceDigest := sha256.Sum256([]byte("nonce-" + rawState))
	policyDigest := sha256.Sum256([]byte("policy-" + rawState))
	return LocalIdentityOIDCAuthorizationTransaction{
		SchemaVersion: localIdentitySchemaVersion, TransactionID: transactionID, Intent: localIdentityOIDCIntentLogin,
		ReturnTo: "/workspace", stateDigest: stateDigest[:], nonceDigest: nonceDigest[:], policyDigest: policyDigest[:],
		codeVerifier:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~",
		LifecycleState: localIdentityOIDCTransactionPending, RecordVersion: 1, CreatedAt: now, UpdatedAt: now,
		ExpiresAt: now.Add(5 * time.Minute), AuditRef: "audit:oidc-authorization",
	}
}

func localIdentityTestAccount(userID string, credentialID string, identifier string, now time.Time) (UserAccount, LocalCredential) {
	normalized, err := NormalizeLocalLoginIdentifier(identifier)
	if err != nil {
		panic(err)
	}
	account := UserAccount{
		SchemaVersion: localIdentitySchemaVersion, UserID: userID, LoginIdentifier: identifier,
		NormalizedLoginIdentifier: normalized, DisplayName: strings.Split(normalized, "@")[0],
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
		AuditRef: "audit:account-create",
	}
	return account, localIdentityTestCredential(credentialID, userID, now)
}

func localIdentityTestCredential(credentialID string, userID string, now time.Time) LocalCredential {
	return LocalCredential{
		SchemaVersion: localIdentitySchemaVersion, CredentialID: credentialID, UserID: userID,
		Algorithm: localPasswordAlgorithmPBKDF2SHA256, PolicyVersion: localPasswordPolicyVersion,
		Iterations: localPasswordIterations, KeyLength: localPasswordKeyLength,
		salt: bytes.Repeat([]byte{0x11}, localPasswordSaltLength), derivedKey: bytes.Repeat([]byte{0x22}, localPasswordKeyLength),
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: now.UTC(), UpdatedAt: now.UTC(),
		AuditRef: "audit:credential-create",
	}
}

func localIdentityTestBinding(bindingID string, userID string, issuer string, subject string) ExternalIdentityBinding {
	return ExternalIdentityBinding{
		SchemaVersion: localIdentitySchemaVersion, BindingID: bindingID, UserID: userID, Issuer: issuer, Subject: subject,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: localIdentityTestNow,
		UpdatedAt: localIdentityTestNow, AuditRef: "audit:binding-create",
	}
}

func localIdentityTestSession(t *testing.T, userID string) WebSession {
	t.Helper()
	digest := localIdentityTestSessionDigest(t)
	return WebSession{
		SchemaVersion: localIdentitySchemaVersion, SessionID: "ses_aaaaaaaaaaaaaaaa", UserID: userID,
		credentialDigest: digest[:], AuthenticationMethod: localAuthenticationMethodPassword,
		AuthenticationSourceRef: "credential:cred_cccccccccccccccc", PolicyVersion: localSessionPolicyVersion,
		LifecycleState: localIdentityStateActive, RecordVersion: 1, CreatedAt: localIdentityTestNow,
		UpdatedAt: localIdentityTestNow, LastVerifiedAt: localIdentityTestNow,
		ExpiresAt: localIdentityTestNow.Add(8 * time.Hour), AuditRef: "audit:session-create",
	}
}

func localIdentityTestSessionDigest(t *testing.T) [32]byte {
	t.Helper()
	digest, err := DigestWebSessionCredential("session-credential-with-at-least-thirty-two-bytes")
	if err != nil {
		t.Fatalf("digest test session credential: %v", err)
	}
	return digest
}

func localIdentityTestMembership(userID string) WorkspaceMembership {
	return WorkspaceMembership{
		SchemaVersion: localIdentitySchemaVersion, MembershipID: "mbr_aaaaaaaaaaaaaaaa", UserID: userID,
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: localIdentityTestNow, UpdatedAt: localIdentityTestNow,
		AuditRef: "audit:membership-create",
	}
}

func localIdentityTestRoleAssignment(userID string) LocalRoleAssignment {
	return LocalRoleAssignment{
		SchemaVersion: localIdentitySchemaVersion, AssignmentID: "rla_aaaaaaaaaaaaaaaa", UserID: userID,
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", RoleKey: "workspace_reader",
		PermissionGrants: []string{"runs:read", "applications:read"}, LifecycleState: localIdentityStateActive,
		RecordVersion: 1, CreatedAt: localIdentityTestNow, UpdatedAt: localIdentityTestNow,
		AuditRef: "audit:role-create",
	}
}

func slicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func runConcurrentLocalIdentitySingleWinner(t *testing.T, name string, conflict error, operations []func() error) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan error, len(operations))
	var wait sync.WaitGroup
	for _, operation := range operations {
		wait.Add(1)
		go func(candidate func() error) {
			defer wait.Done()
			<-start
			results <- candidate()
		}(operation)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for err := range results {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, conflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent %s result: %v", name, err)
		}
	}
	if winners != 1 || conflicts != len(operations)-1 {
		t.Fatalf("concurrent %s single-winner mismatch: winners=%d conflicts=%d", name, winners, conflicts)
	}
}
