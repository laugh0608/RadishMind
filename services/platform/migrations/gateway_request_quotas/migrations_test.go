package gatewayrequestquotamigrations

import (
	"strings"
	"testing"
)

func TestGatewayRequestQuotaMigrationContract(t *testing.T) {
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
	for _, expected := range []string{
		"gateway_request_quota_policies", "gateway_request_quota_usage",
		"gateway_request_quota_admissions", "sanitized_admission_record",
	} {
		if !strings.Contains(upSQL, expected) {
			t.Fatalf("migration is missing %s", expected)
		}
	}
}
