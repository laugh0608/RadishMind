//go:build identity_browser_test

package httpapi

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	"radishmind.local/services/platform/internal/workspacepolicy"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

// TestLocalIdentityBrowserSeed is a one-shot test fixture, not an HTTP route or
// first-user rule. The browser runner supplies the exact registered user/scope
// and owns the temporary database and any one-time invitation material.
func TestLocalIdentityBrowserSeed(t *testing.T) {
	if os.Getenv("RADISHMIND_IDENTITY_BROWSER_TEST") != "1" {
		t.Skip("started only by the isolated identity browser runner")
	}
	var input struct {
		Kind        string `json:"kind"`
		UserID      string `json:"userId"`
		TenantRef   string `json:"tenantRef"`
		WorkspaceID string `json:"workspaceId"`
	}
	if err := json.Unmarshal([]byte(os.Getenv("RADISHMIND_IDENTITY_BROWSER_SEED")), &input); err != nil {
		t.Fatal("invalid identity browser seed request")
	}
	databasePath := os.Getenv("RADISHMIND_SQLITE_DEV_DATABASE_PATH")
	outputPath := os.Getenv("RADISHMIND_IDENTITY_BROWSER_SEED_OUTPUT")
	root := filepath.Dir(databasePath)
	if !filepath.IsAbs(databasePath) || !strings.HasPrefix(filepath.Base(root), "radishmind-workflow-e2e-") ||
		filepath.Dir(outputPath) != root || !localUserIDPattern.MatchString(input.UserID) ||
		!validControlPlaneReadAuthReference(input.TenantRef, false) || !validControlPlaneReadAuthReference(input.WorkspaceID, false) {
		t.Fatal("seed requires the owned temporary runtime and exact user/scope")
	}
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath, Migrations: sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open browser seed database: %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Errorf("close browser seed database: %v", err)
		}
	})
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	var result any
	switch input.Kind {
	case "bootstrap":
		result, err = newLocalIdentityAdministrationService(repository).BootstrapWorkspaceAdministrator(context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{
			TenantRef: input.TenantRef, WorkspaceID: input.WorkspaceID, UserID: input.UserID, AuditRef: "audit:identity-browser-bootstrap",
		})
	case "expired-invitation":
		// Create through the existing service with a historical test clock. HTTP
		// reads and claims use the real clock and must reject the expired record.
		past := time.Now().UTC().Add(-2 * time.Hour)
		service := newWorkspaceInvitationService(repository)
		service.now = func() time.Time { return past }
		role, exists := workspacepolicy.BuiltInRole(workspacepolicy.RoleWorkspaceReader)
		if !exists {
			t.Fatal("reader role missing")
		}
		result, err = service.Create(context.Background(), LocalIdentityAdministrationActor{
			UserID: input.UserID, TenantRef: input.TenantRef, WorkspaceID: input.WorkspaceID, AuthenticatedAt: past,
		}, WorkspaceInvitationCreateInput{
			TenantRef: input.TenantRef, WorkspaceID: input.WorkspaceID, RoleKey: role.RoleKey,
			ExpectedCatalogVersion: role.CatalogVersion, ExpectedRoleDefinitionDigest: role.DefinitionDigest,
			TTLPolicy: "1h", Confirmed: true, RequestRef: "request:identity-browser-expired", AuditRef: "audit:identity-browser-expired",
		})
	default:
		t.Fatal("unsupported identity browser seed operation")
	}
	if err != nil {
		t.Fatalf("identity browser seed failed: %v", err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal("could not encode browser seed result")
	}
	if err := os.WriteFile(outputPath, encoded, 0o600); err != nil {
		t.Fatal("could not save private browser seed result")
	}
}
