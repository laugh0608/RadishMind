package controlplanereadmigrations

import (
	"strings"
	"testing"
)

func TestControlPlaneAdminReadMigrationContract(t *testing.T) {
	if checksum := ExpectedChecksum(); !strings.HasPrefix(checksum, "sha256:") || len(checksum) != 71 {
		t.Fatalf("unexpected migration checksum: %s", checksum)
	}
	for _, expected := range []string{
		"control_plane_read_schema_versions",
		"control_plane_tenant_summary_projections",
		"control_plane_audit_summary_projections",
		"projection_version bigint NOT NULL CHECK (projection_version > 0)",
		"PRIMARY KEY (tenant_ref, audit_ref)",
		"recorded_at DESC, audit_ref DESC",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Fatalf("control plane read migration is missing %s", expected)
		}
	}
	for _, forbidden := range []string{"json", "authorization", "api_key", "raw_payload"} {
		if strings.Contains(strings.ToLower(upSQL), forbidden) {
			t.Fatalf("control plane read migration contains forbidden storage field %s", forbidden)
		}
	}
	for _, expected := range []string{
		"DROP TABLE IF EXISTS control_plane_audit_summary_projections",
		"DROP TABLE IF EXISTS control_plane_tenant_summary_projections",
		"DROP TABLE IF EXISTS control_plane_read_schema_versions",
	} {
		if !strings.Contains(downSQL, expected) {
			t.Fatalf("control plane read down migration is missing %s", expected)
		}
	}
}
