package gatewaymodelpricingmigrations

import (
	"strings"
	"testing"
)

func TestGatewayModelPricingMigrationContract(t *testing.T) {
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
	for _, expected := range []string{
		"gateway_model_pricing_revisions", "gateway_model_pricing_current",
		"sanitized_policy_record", "append-only",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Fatalf("migration is missing %s", expected)
		}
	}
}
