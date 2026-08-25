package httpapi

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

type durableLocalIdentitySelfServiceRepository interface {
	localIdentityRepository
	localIdentitySelfServiceSecurityRepository
}

type durableLocalIdentitySelfServiceState struct {
	Now                  time.Time
	Actor                LocalIdentitySelfServiceActor
	ReplacementID        string
	RevokedSourceSession string
}

func TestSQLiteLocalIdentitySelfServiceContractRestartQueryPlanAndNoFallback(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity-self-service.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open SQLite self-service runtime: %v", err)
	}
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	state := runDurableLocalIdentitySelfServiceContract(t, repository)
	assertSQLiteLocalIdentitySelfServiceQueryPlan(t, runtime)
	if err := runtime.Close(); err != nil {
		t.Fatalf("close SQLite self-service runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("restart SQLite self-service runtime: %v", err)
	}
	restored := newSQLiteLocalIdentityRepository(restarted.DB())
	assertDurableLocalIdentitySelfServiceRestart(t, restored, state)
	if err := restarted.Close(); err != nil {
		t.Fatalf("close restarted SQLite self-service runtime: %v", err)
	}
	unavailableService := newLocalIdentitySelfServiceSecurityService(restored)
	unavailableService.now = func() time.Time { return state.Now }
	if _, err := unavailableService.ListSessions(
		context.Background(), state.Actor, LocalIdentitySelfServiceSessionListQuery{},
	); !errors.Is(err, errLocalIdentitySelfServiceUnavailable) {
		t.Fatalf("closed SQLite self-service owner did not fail closed: %v", err)
	}
}

func TestSQLiteLocalIdentitySelfServiceMigrationUpgradeAndReapply(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "identity-self-service-upgrade.db")
	allMigrations := sqlitelocalidentitymigrations.Migrations()
	legacy, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   allMigrations[:3],
	})
	if err != nil {
		t.Fatalf("open pre-self-service SQLite runtime: %v", err)
	}
	account, credential := localIdentityTestAccount(
		"usr_90000000000000f1", "cred_90000000000000f1", "self-service-upgrade@example.com", localIdentityTestNow,
	)
	if err := newSQLiteLocalIdentityRepository(legacy.DB()).CreateAccount(context.Background(), account, credential); err != nil {
		t.Fatalf("create pre-self-service SQLite account: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close pre-self-service SQLite runtime: %v", err)
	}

	upgraded, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   allMigrations,
	})
	if err != nil {
		t.Fatalf("upgrade SQLite self-service migration: %v", err)
	}
	if _, err := newSQLiteLocalIdentityRepository(upgraded.DB()).ReadAccount(context.Background(), account.UserID); err != nil {
		t.Fatalf("read account after SQLite self-service upgrade: %v", err)
	}
	assertSQLiteLocalIdentitySelfServiceQueryPlan(t, upgraded)
	if err := upgraded.Close(); err != nil {
		t.Fatalf("close upgraded SQLite self-service runtime: %v", err)
	}
	reapplied, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   allMigrations,
	})
	if err != nil {
		t.Fatalf("reapply SQLite self-service migration: %v", err)
	}
	defer reapplied.Close()
	if err := reapplied.VerifyMigrations(context.Background(), allMigrations); err != nil {
		t.Fatalf("verify reapplied SQLite self-service migration: %v", err)
	}
}

func runDurableLocalIdentitySelfServiceContract(
	t *testing.T,
	repository durableLocalIdentitySelfServiceRepository,
) durableLocalIdentitySelfServiceState {
	t.Helper()
	ctx := context.Background()
	now := localIdentitySelfServiceTestNow.Add(24 * time.Hour)
	userID := "usr_9000000000000001"
	credentialID := "cred_9000000000000001"
	currentSessionID := "ses_9000000000000001"
	password := "durable current password value"
	createdAt := now.Add(-2 * time.Hour)
	account, _ := localIdentityTestAccount(userID, credentialID, "durable-self-service@example.com", createdAt)
	credential, err := DeriveLocalCredential(password, credentialID, userID, createdAt, "audit:durable-self-service")
	if err != nil {
		t.Fatalf("derive durable self-service credential: %v", err)
	}
	if err := repository.CreateAccount(ctx, account, credential); err != nil {
		t.Fatalf("create durable self-service account: %v", err)
	}
	createLocalIdentitySelfServiceTestSession(
		t, repository, userID, currentSessionID, localAuthenticationMethodOIDC,
		"binding:xid_9000000000000001", now.Add(-time.Hour), now.Add(8*time.Hour),
	)
	exact := createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000002", localAuthenticationMethodPassword,
		"credential:"+credentialID, now.Add(-30*time.Minute), now.Add(8*time.Hour),
	)
	concurrent := createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000003", localAuthenticationMethodOIDC,
		"binding:xid_9000000000000003", now.Add(-30*time.Minute), now.Add(8*time.Hour),
	)
	createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000004", localAuthenticationMethodPassword,
		"credential:"+credentialID, now.Add(-30*time.Minute), now.Add(8*time.Hour),
	)
	expired := createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000005", localAuthenticationMethodPassword,
		"credential:"+credentialID, now.Add(-time.Hour), now.Add(-time.Minute),
	)
	actor := LocalIdentitySelfServiceActor{
		UserID: userID, CurrentSessionID: currentSessionID, AuthenticatedAt: now.Add(-time.Minute),
	}
	service := newLocalIdentitySelfServiceSecurityService(repository)
	service.now = func() time.Time { return now }
	first, err := service.ListSessions(ctx, actor, LocalIdentitySelfServiceSessionListQuery{
		State: localIdentitySessionStateFilterAll, Limit: 2,
	})
	if err != nil || len(first.Sessions) != 2 || first.NextCursor == "" || !first.SnapshotAt.Equal(now) {
		t.Fatalf("durable self-service first page mismatch: page=%#v err=%v", first, err)
	}
	mutationTime := now.Add(time.Minute)
	service.now = func() time.Time { return mutationTime }
	if _, err := service.RevokeSession(ctx, actor, LocalIdentityRevokeOwnedSessionInput{
		SessionID: exact.SessionID, ExpectedVersion: exact.RecordVersion, Confirmed: true,
		AuditRef: "audit:durable-exact-revoke",
	}); err != nil {
		t.Fatalf("durable exact session revoke: %v", err)
	}
	createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_90000000000000ff", localAuthenticationMethodOIDC,
		"binding:xid_90000000000000ff", now.Add(30*time.Second), now.Add(8*time.Hour),
	)
	seen := append([]LocalIdentitySelfServiceSessionSummary(nil), first.Sessions...)
	cursor := first.NextCursor
	for cursor != "" {
		page, listErr := service.ListSessions(ctx, actor, LocalIdentitySelfServiceSessionListQuery{
			State: localIdentitySessionStateFilterAll, Limit: 2, Cursor: cursor,
		})
		if listErr != nil {
			t.Fatalf("list durable self-service continuation: %v", listErr)
		}
		seen = append(seen, page.Sessions...)
		cursor = page.NextCursor
	}
	if len(seen) != 5 {
		t.Fatalf("durable snapshot pagination count mismatch: %d", len(seen))
	}
	seenIDs := make([]string, 0, len(seen))
	seenSet := make(map[string]struct{}, len(seen))
	foundPostSnapshotRevoke := false
	for _, session := range seen {
		if _, exists := seenSet[session.SessionID]; exists {
			t.Fatalf("durable pagination repeated session %s", session.SessionID)
		}
		seenSet[session.SessionID] = struct{}{}
		seenIDs = append(seenIDs, session.SessionID)
		foundPostSnapshotRevoke = foundPostSnapshotRevoke ||
			session.SessionID == exact.SessionID && session.EffectiveState == localIdentityStateActive
	}
	if !foundPostSnapshotRevoke {
		t.Fatal("durable pagination did not preserve the pre-revoke snapshot")
	}
	if _, exists := seenSet["ses_90000000000000ff"]; exists {
		t.Fatal("post-snapshot durable session leaked into pagination")
	}
	if !slices.IsSortedFunc(seenIDs[:3], func(left, right string) int { return strings.Compare(right, left) }) {
		t.Fatalf("durable same-time session order mismatch: %v", seenIDs)
	}

	concurrentTime := now.Add(2 * time.Minute)
	service.now = func() time.Time { return concurrentTime }
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for index := 0; index < 2; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			_, revokeErr := service.RevokeSession(ctx, actor, LocalIdentityRevokeOwnedSessionInput{
				SessionID: concurrent.SessionID, ExpectedVersion: concurrent.RecordVersion, Confirmed: true,
				AuditRef: fmt.Sprintf("audit:durable-concurrent-revoke-%d", index),
			})
			results <- revokeErr
		}(index)
	}
	close(start)
	wait.Wait()
	close(results)
	winners := 0
	conflicts := 0
	for resultErr := range results {
		switch {
		case resultErr == nil:
			winners++
		case errors.Is(resultErr, errLocalIdentitySessionVersionConflict):
			conflicts++
		default:
			t.Fatalf("unexpected durable concurrent revoke result: %v", resultErr)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("durable concurrent revoke single-winner mismatch: winners=%d conflicts=%d", winners, conflicts)
	}

	bulkTime := now.Add(3 * time.Minute)
	service.now = func() time.Time { return bulkTime }
	bulk, err := service.RevokeOtherSessions(ctx, actor, LocalIdentityRevokeOtherSessionsInput{
		Confirmed: true, AuditRef: "audit:durable-bulk-revoke",
	})
	if err != nil || bulk.RevokedCount != 2 {
		t.Fatalf("durable bulk revoke mismatch: result=%#v err=%v", bulk, err)
	}
	if _, _, err := repository.ReadWebSession(ctx, currentSessionID, bulkTime); err != nil {
		t.Fatalf("durable bulk revoke changed current OIDC session: %v", err)
	}
	storedExpired, err := readLocalIdentitySessionWithoutEffectiveState(repository, ctx, expired.SessionID)
	if err != nil || storedExpired.LifecycleState != localIdentityStateActive {
		t.Fatalf("durable bulk revoke changed effective-expired session: session=%#v err=%v", storedExpired, err)
	}

	activeSource := createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000011", localAuthenticationMethodPassword,
		"credential:"+credentialID, now.Add(-time.Hour), now.Add(9*time.Hour),
	)
	expiredSource := createLocalIdentitySelfServiceTestSession(
		t, repository, userID, "ses_9000000000000012", localAuthenticationMethodPassword,
		"credential:"+credentialID, now.Add(-time.Hour), bulkTime.Add(-time.Second),
	)
	collisionID := "cred_90000000000000ff"
	otherAccount, otherCredential := localIdentityTestAccount(
		"usr_90000000000000ff", collisionID, "durable-self-service-collision@example.com", createdAt,
	)
	if err := repository.CreateAccount(ctx, otherAccount, otherCredential); err != nil {
		t.Fatalf("create durable credential collision owner: %v", err)
	}
	rotationTime := now.Add(4 * time.Minute)
	service.now = func() time.Time { return rotationTime }
	service.newID = func(string) (string, error) { return collisionID, nil }
	if _, err := service.RotateCredential(ctx, actor, LocalIdentityRotateCredentialInput{
		CurrentPassword: password, NewPassword: "durable replacement password value",
		SessionImpactConfirmed: true, AuditRef: "audit:durable-rotation-collision",
	}); !errors.Is(err, errLocalIdentityCredentialRotationConflict) {
		t.Fatalf("durable credential collision did not fail atomically: %v", err)
	}
	if current, readErr := repository.ReadActiveCredential(ctx, userID); readErr != nil || current.CredentialID != credentialID {
		t.Fatalf("failed durable rotation changed active credential: credential=%#v err=%v", current, readErr)
	}
	if stored, readErr := readLocalIdentitySessionWithoutEffectiveState(repository, ctx, activeSource.SessionID); readErr != nil || stored.LifecycleState != localIdentityStateActive {
		t.Fatalf("failed durable rotation partially revoked a session: session=%#v err=%v", stored, readErr)
	}

	replacementID := "cred_9000000000000002"
	service.newID = func(string) (string, error) { return replacementID, nil }
	rotation, err := service.RotateCredential(ctx, actor, LocalIdentityRotateCredentialInput{
		CurrentPassword: password, NewPassword: "durable replacement password value",
		SessionImpactConfirmed: true, AuditRef: "audit:durable-rotation",
	})
	if err != nil || rotation.RevokedSessionCount != 3 || rotation.CurrentSessionRevoked {
		t.Fatalf("durable credential rotation mismatch: result=%#v err=%v", rotation, err)
	}
	if current, readErr := repository.ReadActiveCredential(ctx, userID); readErr != nil ||
		current.CredentialID != replacementID || !VerifyLocalPassword("durable replacement password value", current) {
		t.Fatalf("durable replacement credential mismatch: credential=%#v err=%v", current, readErr)
	}
	for _, sessionID := range []string{expired.SessionID, activeSource.SessionID, expiredSource.SessionID} {
		stored, readErr := readLocalIdentitySessionWithoutEffectiveState(repository, ctx, sessionID)
		if readErr != nil || stored.LifecycleState != localIdentityStateRevoked {
			t.Fatalf("credential-bound durable session was not revoked: session=%#v err=%v", stored, readErr)
		}
	}
	if _, _, err := repository.ReadWebSession(ctx, currentSessionID, rotationTime); err != nil {
		t.Fatalf("durable credential rotation changed current OIDC session: %v", err)
	}
	return durableLocalIdentitySelfServiceState{
		Now: rotationTime, Actor: actor, ReplacementID: replacementID, RevokedSourceSession: expiredSource.SessionID,
	}
}

func assertDurableLocalIdentitySelfServiceRestart(
	t *testing.T,
	repository durableLocalIdentitySelfServiceRepository,
	state durableLocalIdentitySelfServiceState,
) {
	t.Helper()
	credential, err := repository.ReadActiveCredential(context.Background(), state.Actor.UserID)
	if err != nil || credential.CredentialID != state.ReplacementID {
		t.Fatalf("durable self-service restart credential mismatch: credential=%#v err=%v", credential, err)
	}
	service := newLocalIdentitySelfServiceSecurityService(repository)
	service.now = func() time.Time { return state.Now.Add(time.Minute) }
	page, err := service.ListSessions(context.Background(), state.Actor, LocalIdentitySelfServiceSessionListQuery{
		State: localIdentityStateRevoked,
	})
	if err != nil {
		t.Fatalf("list durable self-service sessions after restart: %v", err)
	}
	found := false
	for _, session := range page.Sessions {
		found = found || session.SessionID == state.RevokedSourceSession
	}
	if !found {
		t.Fatalf("durable self-service restart lost revoked session %s", state.RevokedSourceSession)
	}
}

func readLocalIdentitySessionWithoutEffectiveState(
	repository localIdentityRepository,
	ctx context.Context,
	sessionID string,
) (WebSession, error) {
	switch owner := repository.(type) {
	case *sqliteLocalIdentityRepository:
		if owner == nil || owner.database == nil {
			return WebSession{}, errLocalIdentityStoreUnavailable
		}
		return scanSQLiteWebSession(owner.database.QueryRowContext(ctx, sqliteSessionSelect+` WHERE session_id=?`, sessionID))
	case *postgresLocalIdentityRepository:
		if owner == nil || owner.pool == nil {
			return WebSession{}, errLocalIdentityStoreUnavailable
		}
		return scanPostgresWebSession(owner.pool.QueryRow(ctx, postgresSessionSelect+` WHERE session_id=$1`, sessionID))
	default:
		return WebSession{}, errLocalIdentityStoreUnavailable
	}
}

func assertSQLiteLocalIdentitySelfServiceQueryPlan(t *testing.T, runtime *sqlitedev.Runtime) {
	t.Helper()
	if _, err := runtime.DB().ExecContext(context.Background(), "ANALYZE local_web_sessions"); err != nil {
		t.Fatalf("analyze SQLite self-service session statistics: %v", err)
	}
	rows, err := runtime.DB().QueryContext(context.Background(), `EXPLAIN QUERY PLAN
        SELECT session_id FROM local_web_sessions
        WHERE user_id=? AND created_at_unix_nano<=?
        ORDER BY created_at_unix_nano DESC, session_id DESC LIMIT 51`,
		"usr_9000000000000001", localIdentitySelfServiceTestNow.UnixNano(),
	)
	if err != nil {
		t.Fatalf("explain SQLite self-service session query: %v", err)
	}
	defer rows.Close()
	plan := strings.Builder{}
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatalf("scan SQLite self-service query plan: %v", err)
		}
		plan.WriteString(detail)
	}
	if rows.Err() != nil || !strings.Contains(plan.String(), "local_web_sessions_self_service_list_idx") ||
		strings.Contains(plan.String(), "TEMP B-TREE") {
		t.Fatalf("SQLite self-service query did not use its ordered index: %s", plan.String())
	}
}
