//go:build postgres_integration

package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	localidentitymigrations "radishmind.local/services/platform/migrations/local_identity_records"
)

func TestWorkspaceInvitationConfiguredPostgresProduct(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	admin, err := localidentitymigrations.OpenPool(ctx, postgresIntegrationDatabaseURL(t))
	if err != nil {
		t.Fatal("open invitation migration pool failed")
	}
	assertPostgresIntegrationDatabaseIsDisposable(t, ctx, admin)
	resetPostgresLocalIdentitySchema(t, ctx, admin)
	runtimeUser := strings.TrimSpace(os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_USER"))
	if runtimeUser == "" || runtimeUser == os.Getenv("PGUSER") {
		t.Fatal("distinct runtime role is required")
	}
	preparePostgresIntegrationRuntimeRole(t, ctx, admin, runtimeUser)
	t.Cleanup(func() {
		cleanup, finish := context.WithTimeout(context.Background(), 15*time.Second)
		defer finish()
		resetPostgresLocalIdentitySchema(t, cleanup, admin)
		admin.Close()
	})
	if _, err := localidentitymigrations.Apply(ctx, admin); err != nil {
		t.Fatalf("migrate identity owner: %v", err)
	}
	cfg := workspaceInvitationProductConfig()
	cfg.LocalIdentityStoreMode = "postgres_dev_test"
	cfg.LocalIdentityDatabaseTimeout = time.Second
	cfg.LocalIdentityDatabaseURL = postgresIntegrationDatabaseURLForCredentials(t, runtimeUser,
		os.Getenv("RADISHMIND_POSTGRES_INTEGRATION_RUNTIME_PASSWORD"))
	server, fixture := newWorkspaceInvitationConfiguredProduct(t, cfg)
	repository, ok := server.localIdentityRepository.(*postgresLocalIdentityRepository)
	if !ok {
		t.Fatal("configured product did not select PostgreSQL identity owner")
	}
	if _, err := repository.pool.Exec(ctx, "CREATE TABLE invitation_runtime_ddl_forbidden (id integer)"); err == nil {
		t.Fatal("runtime role unexpectedly accepted DDL")
	}
	creation := runWorkspaceInvitationHTTPVerticalChain(t, fixture)
	var persisted string
	if err := admin.QueryRow(ctx, "SELECT row_to_json(i)::text FROM local_workspace_invitations i WHERE invitation_id=$1",
		creation.Invitation.InvitationID).Scan(&persisted); err != nil {
		t.Fatal("read durable invitation audit failed")
	}
	if strings.Contains(persisted, creation.InvitationCode) || strings.Contains(persisted, strings.Split(creation.InvitationCode, ".")[1]) {
		t.Fatal("durable invitation leaked raw credential")
	}
	server.Close()
	closed := fixture.request(t, http.MethodGet, strings.Replace(workspaceInvitationAdminListRoute,
		"{workspace_id}", "workspace_demo", 1), nil, fixture.adminCookies)
	// The protected invitation route cannot establish a Session while its owner is closed.
	assertLocalIdentityError(t, closed, http.StatusUnauthorized, localIdentityAuthenticationRequired)
	restarted, err := NewServerWithError(cfg, Options{TestOnly: true})
	if err != nil {
		t.Fatal("reconnect configured PostgreSQL server failed")
	}
	defer restarted.Close()
	assertWorkspaceInvitationProductRestart(t, restarted, fixture, creation)
	runWorkspaceInvitationConfiguredHTTPRace(t, fixture)
}

func runWorkspaceInvitationConfiguredHTTPRace(t *testing.T, fixture localIdentityAdministrationHTTPFixture) {
	t.Helper()
	role := roleDefinitionByKey(t, LocalIdentityBuiltInRoleCatalog(), localIdentityRoleWorkspaceReader)
	creation := createWorkspaceInvitationOverHTTP(t, fixture, role, workspaceInvitationTTL1Hour)
	const competitors = 8
	requests := make([]*http.Request, competitors)
	for index := range competitors {
		_, cookies, _ := registerLocalIdentityHTTPTestAccount(t, fixture.identity,
			"race-"+string(rune('a'+index))+"@example.test", "concurrent claimant password with entropy")
		requests[index] = workspaceInvitationClaimantRequest(t, fixture, http.MethodPost, workspaceInvitationClaimRoute,
			map[string]any{"invitation_code": creation.document.InvitationCode, "expected_record_version": 1, "confirmed": true},
			cookies, "tenant_demo")
		requests[index].Header.Del(activeWorkspaceHeader)
	}
	start := make(chan struct{})
	statuses := make(chan int, competitors)
	var group sync.WaitGroup
	for _, request := range requests {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			response := httptest.NewRecorder()
			fixture.identity.handler.ServeHTTP(response, request)
			statuses <- response.Code
		}()
	}
	close(start)
	group.Wait()
	close(statuses)
	winners := 0
	for status := range statuses {
		if status == http.StatusOK {
			winners++
		} else if status != http.StatusConflict {
			t.Fatalf("unexpected race status: %d", status)
		}
	}
	if winners != 1 {
		t.Fatalf("configured HTTP race winners: %d", winners)
	}
}
