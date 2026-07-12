package applicationpublishmigrations

import (
	"strings"
	"testing"
)

func TestApplicationPublishMigrationContract(t *testing.T) {
	for _, fragment := range []string{"CREATE TABLE application_publish_candidates", "sanitized_candidate_payload jsonb", "review_version bigint", "application_publish_candidates_scope_list_idx"} {
		if !strings.Contains(upSQL, fragment) {
			t.Fatalf("up migration missing %q", fragment)
		}
	}
	if !strings.Contains(downSQL, "DROP TABLE IF EXISTS application_publish_candidates") {
		t.Fatal("down migration does not remove candidate table")
	}
	if !strings.HasPrefix(ExpectedChecksum(), "sha256:") || len(ExpectedChecksum()) != 71 {
		t.Fatalf("unexpected migration checksum: %s", ExpectedChecksum())
	}
}
