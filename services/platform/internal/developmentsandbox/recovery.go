package developmentsandbox

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"radishmind.local/services/platform/internal/strictjson"
)

// CleanupRecovered explicitly reconciles resources only. It never starts a
// container, SDK, model request or command, or resumes the old attempt.
func CleanupRecovered(ctx context.Context, dockerPath, directory string) (bool, error) {
	if !filepath.IsAbs(dockerPath) || !filepath.IsAbs(directory) {
		return false, ErrSandbox
	}
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || !ownedFile(info) || info.Mode().Perm() != 0700 {
		return false, ErrSandbox
	}
	path := filepath.Join(directory, "resources.json")
	info, err = os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || !singleLink(info) || !ownedFile(info) || info.Size() > 65536 || info.Mode().Perm() != 0600 {
		return false, ErrSandbox
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, ErrSandbox
	}
	var record ledger
	if strictjson.Decode(raw, &record) != nil || record.SchemaVersion != "development_sandbox_resources.v1" || !sessionPattern.MatchString(record.SessionID) || !idPattern.MatchString(record.Binding.AttemptID) || !idPattern.MatchString(record.Binding.ExecutorInstanceID) || !imagePattern.MatchString(record.Binding.PlanDigest) || !imagePattern.MatchString(record.Binding.BaselineDigest) || !imagePattern.MatchString(record.Image) || !imagePattern.MatchString(record.RunnerDigest) || record.Resources == nil || len(record.Resources) > 32 {
		return false, ErrSandbox
	}
	fields, err := strictjson.ObjectFields(raw, "schema_version", "session_id", "binding", "image", "runner_digest", "resources", "closed")
	if err != nil {
		return false, ErrSandbox
	}
	if _, err = strictjson.ObjectFields(fields["binding"], "attempt_id", "executor_instance_id", "plan_digest", "baseline_digest"); err != nil {
		return false, ErrSandbox
	}
	var resources []json.RawMessage
	if json.Unmarshal(fields["resources"], &resources) != nil {
		return false, ErrSandbox
	}
	for _, raw := range resources {
		if _, err = strictjson.ObjectFields(raw, "request_id", "name", "container_id", "candidate_digest", "state"); err != nil {
			return false, ErrSandbox
		}
	}
	seen := map[string]bool{}
	for _, r := range record.Resources {
		if !idPattern.MatchString(r.RequestID) || seen[r.RequestID] || r.Name != "radishmind-code-"+record.SessionID+"-"+r.RequestID || !imagePattern.MatchString(r.CandidateDigest) || r.ContainerID != "" && !containerPattern.MatchString(r.ContainerID) {
			return false, ErrSandbox
		}
		if r.State != "removed" && r.State != "dispatched" && r.State != "intent" && r.State != "created" && r.State != "finished" && r.State != "unknown" && r.State != "create_unknown" {
			return false, ErrSandbox
		}
		seen[r.RequestID] = true
	}
	s := &Session{config: Config{DockerPath: dockerPath, Directory: directory, Image: record.Image}, ledger: record}
	empty, err := s.cleanup(ctx)
	if err != nil || !empty {
		return false, ErrUnknown
	}
	s.ledger.Closed = true
	if err = s.save(); err != nil {
		return false, err
	}
	return true, nil
}
