package gatewayrequestmigrations

import (
	"strings"
	"testing"
)

func TestGatewayRequestMigrationContract(t *testing.T) {
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
	for _, expected := range []string{"gateway_request_records", "sanitized_request_record", "postgres_dev_test", "usage_availability"} {
		if !strings.Contains(upSQL, expected) {
			t.Fatalf("migration is missing %s", expected)
		}
	}
}
