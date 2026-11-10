package developmentsandbox

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	"testing"
	"time"
)

func TestDeliverySnapshotRequiresClosedOwnerAndRetention(t *testing.T) {
	config := fixture(t)
	raw, _ := json.Marshal(config.Plan.Tools)
	config.Delegation.Scope.ToolPolicyDigest = control.PayloadDigest(raw)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	scope := artifact.Scope{TenantRef: "tenant_fixture", WorkspaceID: config.Delegation.Scope.WorkspaceID, AccountID: config.Delegation.Scope.AccountID, ApplicationID: config.Plan.ApplicationID}
	if _, err = s.deliveryInput(context.Background(), scope, "completed", config.Plan); err != artifact.ErrConflict {
		t.Fatal("live snapshot accepted")
	}
	// Synthetic terminal fixture: no Docker is invoked. Production obtains this
	// state only through Close and persisted attempt state in DeliverySource.
	s.ledger.Closed = true
	s.closing = true
	s.artifactSnapshotAt = time.Now().UTC()
	if err = s.save(); err != nil {
		t.Fatal(err)
	}
	in, err := s.deliveryInput(context.Background(), scope, "failed", config.Plan)
	if err != nil {
		t.Fatal(err)
	}
	b, err := artifact.Build(in)
	if err != nil || b.VerificationState != "missing" {
		t.Fatal(err)
	}
	again, err := s.deliveryInput(context.Background(), scope, "failed", config.Plan)
	if err != nil || !again.SourceCreatedAt.Equal(in.SourceCreatedAt) || !again.SnapshotAt.Equal(in.SnapshotAt) {
		t.Fatal("retry renewed retention")
	}
	if err = s.RemoveExpiredArtifacts(context.Background(), in.SourceCreatedAt.Add(artifact.TemporaryLifetime-time.Second)); err != artifact.ErrConflict {
		t.Fatal("early cleanup")
	}
	other := t.TempDir()
	sentinel := filepath.Join(other, "sentinel")
	if err = os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(other, filepath.Join(config.Directory, "external")); err != nil {
		t.Fatal(err)
	}
	if err = s.RemoveExpiredArtifacts(context.Background(), in.SourceCreatedAt.Add(artifact.TemporaryLifetime)); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(sentinel); err != nil {
		t.Fatal("cleanup escaped ownership")
	}
	if _, err = os.Stat(filepath.Join(config.Directory, "candidate")); !os.IsNotExist(err) {
		t.Fatal("candidate retained after cleanup")
	}
	if _, err = os.Stat(filepath.Join(config.Directory, "resources.json")); err != nil {
		t.Fatal("resource audit missing")
	}
	if _, err = s.deliveryInput(context.Background(), scope, "failed", config.Plan); err == nil {
		t.Fatal("recreated missing source")
	}
}
func TestDeliveryVerificationUsesTrustedResultAndMarksUnknown(t *testing.T) {
	config := fixture(t)
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	s.retainVerification("tool_unknown", config.Plan.BaselineDigest, Verification{Status: "command_passed"}, ErrUnknown)
	if len(s.artifactTests) != 1 || s.artifactTests[0].ExitKnown {
		t.Fatal("partial execution promoted")
	}
	s.ledger.Closed = false
	if err = s.RemoveExpiredArtifacts(context.Background(), s.artifactCreatedAt.Add(48*time.Hour)); err != artifact.ErrConflict {
		t.Fatal("unconfirmed resources cleaned")
	}
}

func TestClosePersistenceFailureCannotAuthorizeDelivery(t *testing.T) {
	config := fixture(t)
	// This local no-op program represents empty resource queries; no Docker
	// executable or daemon is invoked by the test.
	config.DockerPath = "/usr/bin/true"
	s, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	recordPath := filepath.Join(config.Directory, "resources.json")
	if err = os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(recordPath, 0700); err != nil {
		t.Fatal(err)
	}
	if closed, err := s.Close(context.Background()); err == nil || closed || s.ledger.Closed || !s.artifactSnapshotAt.IsZero() {
		t.Fatal("failed ledger save authorized snapshot")
	}
	if err = os.Remove(recordPath); err != nil {
		t.Fatal(err)
	}
	if closed, err := s.Close(context.Background()); err != nil || !closed || s.artifactSnapshotAt.IsZero() {
		t.Fatal("explicit retry could not close", err)
	}
	before := s.artifactSnapshotAt
	if _, err = s.Close(context.Background()); err != nil || !s.artifactSnapshotAt.Equal(before) {
		t.Fatal("repeat close changed snapshot identity", err)
	}
}
