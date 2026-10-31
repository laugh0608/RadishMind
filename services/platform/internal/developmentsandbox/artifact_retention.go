package developmentsandbox

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	"radishmind.local/services/platform/internal/strictjson"
)

// Retention is fixed before execution. This metadata contains no source/logs.
func (s *Session) saveArtifactRetention() error {
	return atomicJSON(filepath.Join(s.config.Directory, "artifact-retention.json"), struct {
		SchemaVersion string `json:"schema_version"`
		SessionID     string `json:"session_id"`
		CreatedAt     string `json:"created_at"`
		ExpiresAt     string `json:"expires_at"`
	}{"development_artifact_retention.v1", s.ledger.SessionID, s.artifactCreatedAt.Format(time.RFC3339Nano), s.artifactCreatedAt.Add(artifact.TemporaryLifetime).Format(time.RFC3339Nano)})
}

// RemoveExpiredArtifacts is an explicit host operation, never a background job.
// It preserves resource/retention metadata. It does not call Docker and refuses
// unknown/live ownership; recovered sessions first need resource reconciliation.
func (s *Session) RemoveExpiredArtifacts(ctx context.Context, now time.Time) error {
	select {
	case s.operation <- struct{}{}:
		defer func() { <-s.operation }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if !s.ledger.Closed || s.artifactCreatedAt.IsZero() || now.Before(s.artifactCreatedAt.Add(artifact.TemporaryLifetime)) {
		return artifact.ErrConflict
	}
	info, err := os.Lstat(s.config.Directory)
	if err != nil || !info.IsDir() || !ownedFile(info) || info.Mode().Perm() != 0700 {
		return ErrSandbox
	}
	root, err := os.OpenRoot(s.config.Directory)
	if err != nil {
		return ErrSandbox
	}
	defer root.Close()
	f, err := root.Open(".")
	if err != nil {
		return ErrSandbox
	}
	opened, statErr := f.Stat()
	if statErr != nil || !os.SameFile(info, opened) {
		_ = f.Close()
		return ErrSandbox
	}
	entries, err := f.ReadDir(-1)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return ErrSandbox
	}
	// Resource identity must still be this session's persisted closed ledger.
	recordFile, err := root.Open("resources.json")
	if err != nil {
		return ErrSandbox
	}
	recordInfo, statErr := recordFile.Stat()
	if statErr != nil || !recordInfo.Mode().IsRegular() || !singleLink(recordInfo) || !ownedFile(recordInfo) || recordInfo.Size() > 65536 {
		_ = recordFile.Close()
		return ErrSandbox
	}
	raw, err := io.ReadAll(io.LimitReader(recordFile, 65537))
	recordCloseErr := recordFile.Close()
	if err != nil || recordCloseErr != nil || len(raw) > 65536 {
		return ErrSandbox
	}
	var actual ledger
	if strictjson.Decode(raw, &actual) != nil || !reflect.DeepEqual(actual, s.ledger) || !actual.Closed {
		return ErrSandbox
	}
	for _, entry := range entries {
		if entry.Name() == "resources.json" || entry.Name() == "artifact-retention.json" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if root.RemoveAll(entry.Name()) != nil {
			return ErrSandbox
		}
	}
	return nil
}
