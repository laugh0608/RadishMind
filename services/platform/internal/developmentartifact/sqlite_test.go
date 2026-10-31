package developmentartifact

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"radishmind.local/services/platform/internal/sqlitedev"
	migrations "radishmind.local/services/platform/migrations/sqlite/development_artifacts"
	"sync"
	"testing"
	"time"
)

type fixtureAuthority struct {
	scope       Scope
	denied      bool
	permissions map[string]bool
}

func (a *fixtureAuthority) Authorize(_ context.Context, s Scope, p string) error {
	if a.denied || s != a.scope || !a.permissions[p] {
		return ErrDenied
	}
	return nil
}

type fixtureSource struct {
	in      Input
	calls   int
	failure error
}

func (s *fixtureSource) Snapshot(_ context.Context, _ Scope, _ string) (Input, error) {
	s.calls++
	return s.in, s.failure
}
func newServiceFixture(t *testing.T) (*Service, *fixtureAuthority, *fixtureSource, *sqlitedev.Runtime, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifacts.db")
	db, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: path, Migrations: migrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	store, err := NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	in := fixtureInput(t)
	a := &fixtureAuthority{scope: in.Scope, permissions: map[string]bool{PermissionRead: true, PermissionSave: true, PermissionExport: true}}
	src := &fixtureSource{in: in}
	s, err := NewService(store, a, src)
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return in.SnapshotAt.Add(time.Second) }
	return s, a, src, db, path
}
func TestExplicitSaveReopenExportAndNoReplay(t *testing.T) {
	s, a, src, db, path := newServiceFixture(t)
	ctx := context.Background()
	scope := src.in.Scope
	b, err := s.Preview(ctx, scope, "attempt_fixture")
	if err != nil {
		t.Fatal(err)
	}
	var count int
	db.DB().QueryRow(`SELECT COUNT(*) FROM development_artifacts`).Scan(&count)
	if count != 0 {
		t.Fatal("preview persisted content")
	}
	r, err := s.Save(ctx, scope, "attempt_fixture", "save_fixture", b.Digest)
	if err != nil {
		t.Fatal(err)
	}
	calls := src.calls
	src.failure = ErrUnavailable
	s.now = func() time.Time { return src.in.SnapshotAt.Add(48 * time.Hour) }
	retry, err := s.Save(ctx, scope, "attempt_fixture", "save_fixture", b.Digest)
	if err != nil || retry.ArtifactID != r.ArtifactID || src.calls != calls {
		t.Fatal("retry replayed source", err)
	}
	if _, err = s.Save(ctx, scope, "attempt_fixture", "save_fixture", probe.Digest([]byte("other"))); err != ErrConflict {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: path, Migrations: migrations.Migrations()})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	store, err := NewSQLiteStore(reopened.DB())
	if err != nil {
		t.Fatal(err)
	}
	readOnly, err := NewService(store, a, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := readOnly.Read(ctx, scope, r.ArtifactID)
	if err != nil || got.Bundle.Digest != r.Bundle.Digest {
		t.Fatal("reopen failed", err)
	}
	exported, err := readOnly.Export(ctx, scope, r.ArtifactID, "export_fixture")
	if err != nil {
		t.Fatal(err)
	}
	sha := exported.Digest
	exported.Digest = ""
	raw, _ := json.Marshal(exported)
	if probe.Digest(raw) != sha {
		t.Fatal("export not bound")
	}
	delete(a.permissions, PermissionExport)
	if _, err = readOnly.Export(ctx, scope, r.ArtifactID, "export_denied"); err != ErrDenied {
		t.Fatal("read implied export")
	}
	a.denied = true
	if _, err = readOnly.Read(ctx, scope, r.ArtifactID); err != ErrDenied {
		t.Fatal("revocation ignored")
	}
	a.denied = false
	scope.WorkspaceID = "workspace_other"
	if _, err = readOnly.Read(ctx, scope, r.ArtifactID); err != ErrDenied {
		t.Fatal("cross scope read")
	}
	for _, query := range []string{`UPDATE development_artifacts SET package_bytes=1`, `DELETE FROM development_artifacts`} {
		if _, err = reopened.DB().Exec(query); err == nil {
			t.Fatal("immutable record changed")
		}
	}
}
func TestPreviewChangeExpiryFailureAndCorruption(t *testing.T) {
	s, _, src, db, _ := newServiceFixture(t)
	ctx := context.Background()
	scope := src.in.Scope
	b := mustBuild(t, src.in)
	src.in.Files[1].Candidate += "\n// later candidate\n"
	if _, err := s.Save(ctx, scope, "attempt_fixture", "save_stale", b.Digest); err != ErrConflict {
		t.Fatal(err)
	}
	b = mustBuild(t, src.in)
	s.now = func() time.Time { return src.in.SourceCreatedAt.Add(TemporaryLifetime) }
	if _, err := s.Save(ctx, scope, "attempt_fixture", "save_expired", b.Digest); err != ErrExpired {
		t.Fatal(err)
	}
	s.now = func() time.Time { return src.in.SnapshotAt.Add(time.Second) }
	if _, err := db.DB().Exec(`CREATE TRIGGER fail_artifact AFTER INSERT ON development_artifacts BEGIN SELECT RAISE(ABORT,'synthetic write failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Save(ctx, scope, "attempt_fixture", "save_retry", b.Digest); err != ErrUnavailable {
		t.Fatal(err)
	}
	var count int
	if err := db.DB().QueryRow(`SELECT COUNT(*) FROM development_artifacts`).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial write")
	}
	if _, err := db.DB().Exec(`DROP TRIGGER fail_artifact`); err != nil {
		t.Fatal(err)
	}
	r, err := s.Save(ctx, scope, "attempt_fixture", "save_retry", b.Digest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.DB().Exec(`DROP TRIGGER development_artifacts_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.DB().Exec(`UPDATE development_artifacts SET record_digest='corrupt'`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Export(ctx, scope, r.ArtifactID, "export_corrupt"); err != ErrCorrupt {
		t.Fatal("corrupt export accepted", err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Read(ctx, scope, r.ArtifactID); err != ErrUnavailable {
		t.Fatal("closed store fallback", err)
	}
}
func TestConcurrentQuotaAndIdempotency(t *testing.T) {
	s, _, src, db, _ := newServiceFixture(t)
	ctx := context.Background()
	b := mustBuild(t, src.in)
	makeRecord := func(i int) Record {
		return Record{"development_artifact_record.v1", fmt.Sprintf("devart_%03d", i), fmt.Sprintf("save_%03d", i), src.in.SnapshotAt.Add(time.Second).Format(time.RFC3339Nano), b}
	}
	for i := 0; i < 63; i++ {
		if _, err := s.store.save(ctx, makeRecord(i)); err != nil {
			t.Fatal(err)
		}
	}
	second, err := NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	outcomes := make(chan error, 2)
	var wg sync.WaitGroup
	for i, store := range []*SQLiteStore{s.store, second} {
		wg.Add(1)
		go func(i int, store *SQLiteStore) {
			defer wg.Done()
			_, err := store.save(ctx, makeRecord(63+i))
			outcomes <- err
		}(i, store)
	}
	wg.Wait()
	close(outcomes)
	success, quota := 0, 0
	for err := range outcomes {
		if err == nil {
			success++
		} else if err == ErrQuota {
			quota++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || quota != 1 {
		t.Fatal(success, quota)
	}
	// Even at capacity, a repeated request resolves the existing immutable row.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r, err := second.save(ctx, makeRecord(0)); err != nil || r.ArtifactID != "devart_000" {
				t.Errorf("idempotency failed: %v", err)
			}
		}()
	}
	wg.Wait()
}
