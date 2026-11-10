package httpapi

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	artifact "radishmind.local/services/platform/internal/developmentartifact"
	"testing"
	"time"
)

type artifactSourceFixture struct{ input artifact.Input }

func (s artifactSourceFixture) Snapshot(context.Context, artifact.Scope, string) (artifact.Input, error) {
	return s.input, nil
}
func TestDevelopmentArtifactCurrentIdentityAndIndependentGrants(t *testing.T) {
	f := newDevelopmentServerFixture(t)
	ctx := context.Background()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "scripts", "eval", "fixtures", "development-control", "artifact.json"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := artifact.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	b.Scope = artifact.Scope{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, AccountID: f.identity.AccountID, ApplicationID: f.plan.ApplicationID}
	b.Origin.AccountID = b.Scope.AccountID
	b.Origin.WorkspaceID = b.Scope.WorkspaceID
	b.Origin.ApplicationID = b.Scope.ApplicationID
	b.Plan.ApplicationID = b.Scope.ApplicationID
	b.Origin.PlanDigest, err = b.Plan.Digest()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	in := artifact.Input{Scope: b.Scope, Origin: b.Origin, Plan: b.Plan, Probe: b.Probe, AttemptState: "completed", SourceCreatedAt: now.Add(-time.Minute), SnapshotAt: now, Files: b.Files, Tests: []artifact.Test{}}
	service, err := f.server.ConnectDevelopmentArtifacts(f.request, artifactSourceFixture{in})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Preview(ctx, b.Scope, b.Origin.AttemptID); err != artifact.ErrDenied {
		t.Fatal("existing admin implicitly gained permission", err)
	}
	role := f.role
	role.AssignmentID = "rla_aaaaaaaaaaaaaaab"
	role.RoleKey = "development_artifact_operator"
	role.RoleCatalogVersion = ""
	role.RoleDefinitionDigest = ""
	role.PermissionGrants = []string{artifact.PermissionRead, artifact.PermissionSave, artifact.PermissionExport}
	if err = f.identityRepository.CreateRoleAssignment(ctx, role); err != nil {
		t.Fatal(err)
	}
	preview, err := service.Preview(ctx, b.Scope, b.Origin.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := service.Save(ctx, b.Scope, b.Origin.AttemptID, "save_artifact_fixture", preview.Digest)
	if err != nil {
		t.Fatal(err)
	}
	// Saved content reading must not depend on a live model or price policy.
	f.server.bridge = nil
	f.server.gatewayModelPricingRepository = nil
	exported, err := service.Export(ctx, b.Scope, saved.ArtifactID, "export_artifact_fixture")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(exported)
	if _, err = artifact.ParseExport(encoded); err != nil {
		t.Fatal(err)
	}
	other := b.Scope
	other.AccountID = "account_other"
	if _, err = service.Read(ctx, other, saved.ArtifactID); err != artifact.ErrDenied {
		t.Fatal("cross owner read")
	}
	if _, err = f.identityRepository.RevokeRoleAssignment(ctx, role.AssignmentID, role.RecordVersion, now.Add(time.Second), "audit:artifact-revoke"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Read(ctx, b.Scope, saved.ArtifactID); err != artifact.ErrDenied {
		t.Fatal("role revocation ignored", err)
	}
	if _, err = f.server.ConnectDevelopmentArtifacts(httptest.NewRequest("POST", "/", nil), nil); err != artifact.ErrDenied {
		t.Fatal("unverified request accepted")
	}
	// A task credential cannot establish a local user session.
	request := httptest.NewRequest("POST", "/", nil)
	request.Header.Set("Authorization", "Bearer task_fixture")
	if _, err = f.server.ConnectDevelopmentArtifacts(request, nil); err != artifact.ErrDenied {
		t.Fatal("task credential granted content")
	}
}
func TestDevelopmentArtifactSessionAndMembershipRevocation(t *testing.T) {
	for _, kind := range []string{"session", "membership", "account"} {
		t.Run(kind, func(t *testing.T) {
			f := newDevelopmentServerFixture(t)
			ctx := context.Background()
			session, _ := requireLocalIdentityRequestSession(f.request)
			a := developmentArtifactAuthority{f.server, session.sessionID, f.identity.AccountID, f.identity.TenantRef, f.identity.WorkspaceID}
			scope := artifact.Scope{TenantRef: f.identity.TenantRef, WorkspaceID: f.identity.WorkspaceID, AccountID: f.identity.AccountID, ApplicationID: f.plan.ApplicationID}
			role := f.role
			role.AssignmentID = "rla_aaaaaaaaaaaaaaab"
			role.RoleKey = "development_artifact_operator"
			role.RoleCatalogVersion = ""
			role.RoleDefinitionDigest = ""
			role.PermissionGrants = []string{artifact.PermissionRead}
			if err := f.identityRepository.CreateRoleAssignment(ctx, role); err != nil {
				t.Fatal(err)
			}
			if err := a.Authorize(ctx, scope, artifact.PermissionRead); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "session":
				_, err := f.identityRepository.RevokeWebSession(ctx, session.sessionID, session.sessionVersion, time.Now(), "audit:artifact-test")
				if err != nil {
					t.Fatal(err)
				}
			case "membership":
				if _, err := revokeDevelopmentFixtureMembership(t, f); err != nil {
					t.Fatal(err)
				}
			case "account":
				account, err := f.identityRepository.ReadAccount(ctx, f.identity.AccountID)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.identityRepository.DisableAccount(ctx, account.UserID, account.RecordVersion, time.Now(), "audit:artifact-test")
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := a.Authorize(ctx, scope, artifact.PermissionRead); err != artifact.ErrDenied {
				t.Fatal("revocation ignored", err)
			}
		})
	}
}
