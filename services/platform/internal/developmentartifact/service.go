package developmentartifact

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	probe "radishmind.local/services/platform/internal/developmentprobe"
)

type Service struct {
	store     *SQLiteStore
	authority Authorizer
	source    SourceReader
	now       func() time.Time
}

func NewService(store *SQLiteStore, authority Authorizer, source SourceReader) (*Service, error) {
	if store == nil || authority == nil {
		return nil, ErrUnavailable
	}
	return &Service{store: store, authority: authority, source: source, now: time.Now}, nil
}
func validTime(s string) bool {
	t, err := time.Parse(time.RFC3339Nano, s)
	return err == nil && !t.IsZero()
}
func (s *Service) authorize(ctx context.Context, scope Scope, permission string) error {
	if scope.Validate() != nil {
		return ErrInvalid
	}
	return s.authority.Authorize(ctx, scope, permission)
}
func (s *Service) preview(ctx context.Context, scope Scope, attempt string) (Bundle, error) {
	if !idPattern.MatchString(attempt) || s.source == nil {
		return Bundle{}, ErrInvalid
	}
	in, err := s.source.Snapshot(ctx, scope, attempt)
	if err != nil {
		return Bundle{}, ErrUnavailable
	}
	if in.Scope != scope || in.Origin.AttemptID != attempt {
		return Bundle{}, ErrDenied
	}
	now := s.now()
	if in.SnapshotAt.After(now) || !now.Before(in.SourceCreatedAt.Add(TemporaryLifetime)) {
		return Bundle{}, ErrExpired
	}
	return Build(in)
}
func (s *Service) Preview(ctx context.Context, scope Scope, attempt string) (Bundle, error) {
	if err := s.authorize(ctx, scope, PermissionRead); err != nil {
		return Bundle{}, err
	}
	return s.preview(ctx, scope, attempt)
}
func (s *Service) Save(ctx context.Context, scope Scope, attempt, request, confirmedDigest string) (Record, error) {
	if !idPattern.MatchString(attempt) || !idPattern.MatchString(request) || !digestPattern.MatchString(confirmedDigest) {
		return Record{}, ErrInvalid
	}
	if err := s.authorize(ctx, scope, PermissionRead); err != nil {
		return Record{}, err
	}
	if err := s.authorize(ctx, scope, PermissionSave); err != nil {
		return Record{}, err
	}
	// Acknowledgement loss must be recoverable even after temporary source expiry.
	old, err := s.store.readRequest(ctx, scope, request)
	if err == nil {
		if old.Bundle.Origin.AttemptID != attempt || old.Bundle.Digest != confirmedDigest {
			return Record{}, ErrConflict
		}
		return old, nil
	}
	if err != ErrNotFound {
		return Record{}, err
	}
	b, err := s.preview(ctx, scope, attempt)
	if err != nil {
		return Record{}, err
	}
	if b.Digest != confirmedDigest {
		return Record{}, ErrConflict
	}
	if err = s.authorize(ctx, scope, PermissionSave); err != nil {
		return Record{}, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return Record{}, ErrUnavailable
	}
	r := Record{SchemaVersion: "development_artifact_record.v1", ArtifactID: "devart_" + hex.EncodeToString(nonce[:]), RequestID: request, SavedAt: s.now().UTC().Format(time.RFC3339Nano), Bundle: b}
	return s.store.save(ctx, r)
}
func (s *Service) Read(ctx context.Context, scope Scope, id string) (Record, error) {
	if err := s.authorize(ctx, scope, PermissionRead); err != nil {
		return Record{}, err
	}
	if !idPattern.MatchString(id) {
		return Record{}, ErrInvalid
	}
	return s.store.read(ctx, scope, id)
}
func (s *Service) Export(ctx context.Context, scope Scope, id, request string) (Export, error) {
	if !idPattern.MatchString(id) || !idPattern.MatchString(request) {
		return Export{}, ErrInvalid
	}
	if err := s.authorize(ctx, scope, PermissionRead); err != nil {
		return Export{}, err
	}
	if err := s.authorize(ctx, scope, PermissionExport); err != nil {
		return Export{}, err
	}
	r, err := s.store.read(ctx, scope, id)
	if err != nil {
		return Export{}, err
	}
	e := Export{SchemaVersion: ExportSchema, Record: r, ExportedAt: s.now().UTC().Format(time.RFC3339Nano), RequestID: request}
	raw, err := json.Marshal(e)
	if err != nil {
		return Export{}, ErrInvalid
	}
	e.Digest = probe.Digest(raw)
	encoded, err := json.Marshal(e)
	if err != nil {
		return Export{}, ErrInvalid
	}
	return ParseExport(encoded)
}
