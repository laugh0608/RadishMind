package developmentartifact

import (
	"encoding/json"
	"strings"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"radishmind.local/services/platform/internal/strictjson"
)

func fields(data []byte, names string) (map[string]json.RawMessage, error) {
	return strictjson.ObjectFields(data, strings.Fields(names)...)
}
func Parse(data []byte) (Bundle, error) {
	var b Bundle
	if strictjson.DecodeBounded(data, &b, MaxPackageBytes) != nil {
		return Bundle{}, ErrInvalid
	}
	f, err := fields(data, "source_created_at schema_version scope origin plan probe source_kind attempt_state snapshot_at expires_at files candidate_digest patch patch_digest tests verification_state review_state acceptance_state sensitive_policy digest")
	if err != nil {
		return Bundle{}, ErrInvalid
	}
	if _, err = fields(f["scope"], "tenant_ref workspace_id account_id application_id"); err != nil {
		return Bundle{}, ErrInvalid
	}
	if _, err = fields(f["origin"], "account_id workspace_id application_id application_authority_digest binding_id binding_revision authorization_revision run_id task_id attempt_id plan_digest baseline_digest model_ref pricing_digest tool_policy_digest"); err != nil {
		return Bundle{}, ErrInvalid
	}
	if _, err = control.ParsePlan(f["plan"]); err != nil {
		return Bundle{}, ErrInvalid
	}
	if _, err = probe.ParsePlan(f["probe"]); err != nil {
		return Bundle{}, ErrInvalid
	}
	for _, entry := range []struct{ key, names string }{{"files", "path base base_digest candidate candidate_digest"}, {"tests", "request_id candidate_digest profile argv image_digest runner_digest started_at finished_at exit_known exit_code oom_killed output_truncated log log_digest log_redacted status"}} {
		var items []json.RawMessage
		if json.Unmarshal(f[entry.key], &items) != nil {
			return Bundle{}, ErrInvalid
		}
		for _, item := range items {
			if _, err = fields(item, entry.names); err != nil {
				return Bundle{}, ErrInvalid
			}
		}
	}
	return b, b.Validate()
}
func parseRecord(data []byte) (Record, error) {
	var r Record
	if strictjson.DecodeBounded(data, &r, MaxPackageBytes+4096) != nil {
		return r, ErrCorrupt
	}
	f, err := fields(data, "schema_version artifact_id request_id saved_at bundle")
	if err != nil {
		return r, ErrCorrupt
	}
	if _, err = Parse(f["bundle"]); err != nil {
		return r, ErrCorrupt
	}
	if r.SchemaVersion != "development_artifact_record.v1" || !idPattern.MatchString(r.ArtifactID) || !idPattern.MatchString(r.RequestID) || !validTime(r.SavedAt) {
		return r, ErrCorrupt
	}
	saved, _ := time.Parse(time.RFC3339Nano, r.SavedAt)
	snapshot, _ := time.Parse(time.RFC3339Nano, r.Bundle.SnapshotAt)
	expiry, _ := time.Parse(time.RFC3339Nano, r.Bundle.ExpiresAt)
	if saved.Before(snapshot) || !saved.Before(expiry) {
		return r, ErrCorrupt
	}
	return r, nil
}

// ParseExport validates the downloaded projection without importing or saving it.
func ParseExport(data []byte) (Export, error) {
	var e Export
	if strictjson.DecodeBounded(data, &e, MaxPackageBytes+8192) != nil {
		return e, ErrInvalid
	}
	f, err := fields(data, "schema_version record exported_at request_id digest")
	if err != nil {
		return e, ErrInvalid
	}
	r, err := parseRecord(f["record"])
	if err != nil {
		return e, err
	}
	if e.SchemaVersion != ExportSchema || !validTime(e.ExportedAt) || !idPattern.MatchString(e.RequestID) {
		return e, ErrInvalid
	}
	at, _ := time.Parse(time.RFC3339Nano, e.ExportedAt)
	saved, _ := time.Parse(time.RFC3339Nano, r.SavedAt)
	if at.Before(saved) {
		return e, ErrInvalid
	}
	digest := e.Digest
	e.Digest = ""
	raw, err := json.Marshal(e)
	e.Digest = digest
	if err != nil || probe.Digest(raw) != digest {
		return e, ErrInvalid
	}
	return e, nil
}
