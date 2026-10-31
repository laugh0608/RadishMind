package developmentartifact

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	probe "radishmind.local/services/platform/internal/developmentprobe"
)

type SQLiteStore struct{ db *sql.DB }

func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if db == nil {
		return nil, ErrUnavailable
	}
	var name string
	if db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='development_artifacts'`).Scan(&name) != nil {
		return nil, ErrUnavailable
	}
	return &SQLiteStore{db: db}, nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readRecord(ctx context.Context, q queryer, scope Scope, key, value string) (Record, error) {
	// key is a private constant selected by read/readRequest, never user SQL.
	var body, digest, bundleDigest string
	var size int
	err := q.QueryRowContext(ctx, `SELECT record,record_digest,bundle_digest,package_bytes FROM development_artifacts WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND application_id=? AND `+key+`=?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, value).Scan(&body, &digest, &bundleDigest, &size)
	if errors.Is(err, sql.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, ErrUnavailable
	}
	if len(body) > MaxPackageBytes+4096 || probe.Digest([]byte(body)) != digest {
		return Record{}, ErrCorrupt
	}
	r, err := parseRecord([]byte(body))
	if err != nil {
		return Record{}, ErrCorrupt
	}
	raw, _ := json.Marshal(r.Bundle)
	if r.Bundle.Scope != scope || r.Bundle.Digest != bundleDigest || len(raw) != size || (key == "artifact_id" && r.ArtifactID != value) || (key == "request_id" && r.RequestID != value) {
		return Record{}, ErrCorrupt
	}
	return r, nil
}
func (s *SQLiteStore) read(ctx context.Context, scope Scope, id string) (Record, error) {
	return readRecord(ctx, s.db, scope, "artifact_id", id)
}
func (s *SQLiteStore) readRequest(ctx context.Context, scope Scope, id string) (Record, error) {
	return readRecord(ctx, s.db, scope, "request_id", id)
}
func (s *SQLiteStore) save(ctx context.Context, r Record) (Record, error) {
	body, err := json.Marshal(r)
	if err != nil {
		return Record{}, ErrInvalid
	}
	if _, err = parseRecord(body); err != nil {
		return Record{}, ErrInvalid
	}
	bundle, _ := json.Marshal(r.Bundle)
	scope := r.Bundle.Scope
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return Record{}, ErrUnavailable
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return Record{}, ErrUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = conn.ExecContext(cleanup, "ROLLBACK")
		}
	}()
	old, err := readRecord(ctx, conn, scope, "request_id", r.RequestID)
	if err == nil {
		if old.Bundle.Digest != r.Bundle.Digest {
			return Record{}, ErrConflict
		}
		return old, nil
	}
	if err != ErrNotFound {
		return Record{}, err
	}
	var count, total int
	if conn.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(package_bytes),0) FROM development_artifacts WHERE tenant_ref=? AND workspace_id=? AND account_id=?`, scope.TenantRef, scope.WorkspaceID, scope.AccountID).Scan(&count, &total) != nil {
		return Record{}, ErrUnavailable
	}
	if count >= MaxSavedCount || total+len(bundle) > MaxSavedBytes {
		return Record{}, ErrQuota
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO development_artifacts(tenant_ref,workspace_id,account_id,application_id,artifact_id,request_id,bundle_digest,package_bytes,record,record_digest) VALUES(?,?,?,?,?,?,?,?,?,?)`, scope.TenantRef, scope.WorkspaceID, scope.AccountID, scope.ApplicationID, r.ArtifactID, r.RequestID, r.Bundle.Digest, len(bundle), string(body), probe.Digest(body))
	if err != nil {
		return Record{}, ErrUnavailable
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return Record{}, ErrUnavailable
	}
	committed = true
	return r, nil
}
