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
		if !strings.Contains(initialUpSQL, expected) {
			t.Fatalf("migration is missing %s", expected)
		}
	}
	for _, expected := range []string{"cost_availability", "legacy_not_captured", "price_unavailable"} {
		if !strings.Contains(costEstimateUpSQL, expected) {
			t.Fatalf("cost estimate migration is missing %s", expected)
		}
	}
}
