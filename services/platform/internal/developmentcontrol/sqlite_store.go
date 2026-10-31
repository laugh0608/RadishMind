package developmentcontrol

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
)

type SQLiteStore struct {
	db     *sql.DB
	writes chan struct{}
}

func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {
	if db == nil {
		return nil, ErrUnavailable
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='development_attempts'`).Scan(&name); err != nil {
		return nil, ErrUnavailable
	}
	return &SQLiteStore{db: db, writes: make(chan struct{}, 1)}, nil
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (s *SQLiteStore) transaction(ctx context.Context, apply func(*sql.Conn) error) error {
	// Wait without occupying a pool connection: qualification owners may need
	// the same pool while the single SQLite writer is inside a transaction.
	select {
	case s.writes <- struct{}{}:
	case <-ctx.Done():
		return ErrUnavailable
	}
	defer func() { <-s.writes }()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return ErrUnavailable
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return ErrUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	if err = apply(conn); err != nil {
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return ErrUnavailable
	}
	committed = true
	return nil
}

func (s *SQLiteStore) Read(ctx context.Context, identity Identity, attemptID string) (*Record, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, ErrUnavailable
	}
	defer tx.Rollback()
	r, err := readRecord(ctx, tx, identity, attemptID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, ErrUnavailable
	}
	return r, nil
}

func readRecord(ctx context.Context, q queryer, identity Identity, attemptID string) (*Record, error) {
	var state, binding, stateSHA, bindingSHA string
	var revoked int
	r := &Record{Offers: map[string]Offer{}}
	err := q.QueryRowContext(ctx, `SELECT owner_instance_id,record_version,credential_revoked,state_record,state_digest,binding_record,binding_digest,context_digest,context_count,pending_tool_request_id
 FROM development_attempts WHERE tenant_ref=? AND workspace_id=? AND account_id=? AND attempt_id=?`, identity.TenantRef, identity.WorkspaceID, identity.AccountID, attemptID).Scan(&r.OwnerInstanceID, &r.Version, &revoked, &state, &stateSHA, &binding, &bindingSHA, &r.ContextDigest, &r.ContextCount, &r.PendingToolRequestID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrUnavailable
	}
	if PayloadDigest([]byte(state)) != stateSHA || PayloadDigest([]byte(binding)) != bindingSHA {
		return nil, ErrUnavailable
	}
	if err = json.Unmarshal([]byte(binding), &r.Binding); err != nil || r.Binding.Validate() != nil || r.Binding.Identity != identity {
		return nil, ErrUnavailable
	}
	events, err := readEvents(ctx, q, identity, attemptID)
	if err != nil {
		return nil, err
	}
	r.Control, err = decodeState([]byte(state), events)
	if err != nil {
		return nil, ErrUnavailable
	}
	d := r.Control.delegation
	planSHA, err := r.Binding.Plan.Digest()
	if err != nil {
		return nil, ErrUnavailable
	}
	policyJSON, err := json.Marshal(r.Binding.Plan.Tools)
	if err != nil {
		return nil, ErrUnavailable
	}
	if d.Scope.AccountID != identity.AccountID || d.Scope.WorkspaceID != identity.WorkspaceID || d.Scope.AttemptID != attemptID || d.Scope.ApplicationID != r.Binding.Plan.ApplicationID || d.Scope.ApplicationAuthorityDigest != r.Binding.Selection.ApplicationAuthorityDigest || d.Scope.PricingDigest != r.Binding.Selection.PricingDigest || d.Scope.ModelRef != r.Binding.Plan.ModelRef || d.Scope.PlanDigest != planSHA || d.Scope.BaselineDigest != r.Binding.Plan.BaselineDigest || d.Scope.ToolPolicyDigest != PayloadDigest(policyJSON) || d.Limits != r.Binding.Plan.Limits {
		return nil, ErrUnavailable
	}
	r.Revoked = revoked == 1
	rows, err := q.QueryContext(ctx, `SELECT request_id,parent_request_id,tool_call_id,payload_digest,arguments_digest,state,observation_digest FROM development_tool_offers WHERE tenant_ref=? AND workspace_id=? AND attempt_id=? ORDER BY request_id`, identity.TenantRef, identity.WorkspaceID, attemptID)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer rows.Close()
	for rows.Next() {
		var offer Offer
		if err := rows.Scan(&offer.RequestID, &offer.ParentRequestID, &offer.ToolCallID, &offer.PayloadDigest, &offer.ArgumentsDigest, &offer.State, &offer.ObservationDigest); err != nil {
			return nil, ErrUnavailable
		}
		if !idPattern.MatchString(offer.RequestID) || !digestPattern.MatchString(offer.PayloadDigest) || !digestPattern.MatchString(offer.ArgumentsDigest) {
			return nil, ErrUnavailable
		}
		parent := r.Control.operations[offer.ParentRequestID]
		if parent == nil || parent.request.Kind != "model" || !toolCallIDPattern.MatchString(offer.ToolCallID) || len(r.Offers) >= 8 {
			return nil, ErrUnavailable
		}
		op := r.Control.operations[offer.RequestID]
		if offer.State == "pending" {
			if op != nil || offer.ObservationDigest != "" {
				return nil, ErrUnavailable
			}
		} else if op == nil || op.request.Kind != "tool" || op.request.PayloadDigest != offer.PayloadDigest {
			return nil, ErrUnavailable
		} else if offer.State == "completed" {
			if op.state != "settled" || !digestPattern.MatchString(offer.ObservationDigest) {
				return nil, ErrUnavailable
			}
		} else if offer.State != "queued" || offer.ObservationDigest != "" {
			return nil, ErrUnavailable
		}
		r.Offers[offer.RequestID] = offer
	}
	if rows.Err() != nil {
		return nil, ErrUnavailable
	}
	if err := rows.Close(); err != nil {
		return nil, ErrUnavailable
	}
	if r.ContextCount == 0 {
		if r.ContextDigest != "" || r.PendingToolRequestID != "" || len(r.Offers) != 0 {
			return nil, ErrUnavailable
		}
	} else if r.ContextCount < 2 || r.ContextCount > 63 || !digestPattern.MatchString(r.ContextDigest) || r.Offers[r.PendingToolRequestID].RequestID == "" {
		return nil, ErrUnavailable
	}
	requests, err := q.QueryContext(ctx, `SELECT request_id,kind,payload_digest FROM development_requests WHERE tenant_ref=? AND workspace_id=? AND attempt_id=?`, identity.TenantRef, identity.WorkspaceID, attemptID)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer requests.Close()
	count := 0
	for requests.Next() {
		var id, kind, digest string
		if err := requests.Scan(&id, &kind, &digest); err != nil {
			return nil, ErrUnavailable
		}
		op := r.Control.operations[id]
		if op == nil || op.request.Kind != kind || op.request.PayloadDigest != digest {
			return nil, ErrUnavailable
		}
		count++
	}
	if requests.Err() != nil || count != len(r.Control.operations) {
		return nil, ErrUnavailable
	}
	return r, nil
}

func readEvents(ctx context.Context, q queryer, identity Identity, attemptID string) ([]Event, error) {
	rows, err := q.QueryContext(ctx, `SELECT sequence,event_record,event_digest FROM development_events WHERE tenant_ref=? AND workspace_id=? AND attempt_id=? ORDER BY sequence`, identity.TenantRef, identity.WorkspaceID, attemptID)
	if err != nil {
		return nil, ErrUnavailable
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var sequence int64
		var body, sha string
		if err := rows.Scan(&sequence, &body, &sha); err != nil || PayloadDigest([]byte(body)) != sha {
			return nil, ErrUnavailable
		}
		var event Event
		if err := json.Unmarshal([]byte(body), &event); err != nil || event.Sequence != sequence {
			return nil, ErrUnavailable
		}
		events = append(events, event)
	}
	if rows.Err() != nil {
		return nil, ErrUnavailable
	}
	return events, nil
}

func (s *SQLiteStore) create(ctx context.Context, confirmationID, ownerID, credentialSHA string, binding Binding, c *AttemptControl) error {
	if !idPattern.MatchString(confirmationID) || binding.Validate() != nil || !digestPattern.MatchString(credentialSHA) {
		return ErrIneligible
	}
	return s.transaction(ctx, func(conn *sql.Conn) error {
		state, err := encodeState(c)
		if err != nil {
			return ErrUnavailable
		}
		body, err := json.Marshal(binding)
		if err != nil {
			return ErrUnavailable
		}
		result, err := conn.ExecContext(ctx, `INSERT INTO development_attempts(tenant_ref,workspace_id,attempt_id,account_id,confirmation_id,owner_instance_id,record_version,credential_digest,credential_revoked,state_record,state_digest,binding_record,binding_digest,context_digest,context_count,pending_tool_request_id) VALUES(?,?,?,?,?,?,1,?,0,?,?,?,?,?,0,'') ON CONFLICT DO NOTHING`, binding.Identity.TenantRef, binding.Identity.WorkspaceID, c.delegation.Scope.AttemptID, binding.Identity.AccountID, confirmationID, ownerID, credentialSHA, string(state), PayloadDigest(state), string(body), PayloadDigest(body), "")
		if err != nil {
			return ErrUnavailable
		}
		n, err := result.RowsAffected()
		if err != nil {
			return ErrUnavailable
		}
		if n != 1 {
			return ErrConflict
		}
		return persistChildren(ctx, conn, binding.Identity, c.delegation.Scope.AttemptID, c, 0, nil)
	})
}

// change always loads a fresh private snapshot inside the write transaction.
// Failure discards that snapshot, so an unsuccessful commit cannot mutate a live
// controller or leave a consumed proposal without its reservation/event.
func (s *SQLiteStore) change(ctx context.Context, identity Identity, attemptID string, expected int64, ownerID string, apply func(*Record) error) (*Record, error) {
	var result *Record
	err := s.transaction(ctx, func(conn *sql.Conn) error {
		r, err := readRecord(ctx, conn, identity, attemptID)
		if err != nil {
			return err
		}
		if expected > 0 && expected != r.Version {
			return ErrConflict
		}
		if ownerID != "" && r.OwnerInstanceID != ownerID {
			return ErrCredential
		}
		beforeState, err := encodeState(r.Control)
		if err != nil {
			return ErrUnavailable
		}
		beforeRevoked, beforeContext, beforeCount, beforePending := r.Revoked, r.ContextDigest, r.ContextCount, r.PendingToolRequestID
		beforeOffers := maps.Clone(r.Offers)
		count := len(r.Control.events)
		if err := apply(r); err != nil {
			return err
		}
		state, err := encodeState(r.Control)
		if err != nil {
			return ErrUnavailable
		}
		if bytes.Equal(beforeState, state) && r.Revoked == beforeRevoked && r.ContextDigest == beforeContext && r.ContextCount == beforeCount && r.PendingToolRequestID == beforePending && maps.Equal(beforeOffers, r.Offers) {
			result = r
			return nil
		}
		changed, err := conn.ExecContext(ctx, `UPDATE development_attempts SET record_version=record_version+1,credential_revoked=?,state_record=?,state_digest=?,context_digest=?,context_count=?,pending_tool_request_id=? WHERE tenant_ref=? AND workspace_id=? AND attempt_id=? AND record_version=?`, r.Revoked, string(state), PayloadDigest(state), r.ContextDigest, r.ContextCount, r.PendingToolRequestID, identity.TenantRef, identity.WorkspaceID, attemptID, r.Version)
		if err != nil {
			return ErrUnavailable
		}
		n, err := changed.RowsAffected()
		if err != nil {
			return ErrUnavailable
		}
		if n != 1 {
			return ErrConflict
		}
		if err := persistChildren(ctx, conn, identity, attemptID, r.Control, count, r.Offers); err != nil {
			return err
		}
		r.Version++
		result = r
		return nil
	})
	return result, err
}

func persistChildren(ctx context.Context, conn *sql.Conn, identity Identity, attemptID string, c *AttemptControl, oldCount int, offers map[string]Offer) error {
	for _, op := range c.operations {
		_, err := conn.ExecContext(ctx, `INSERT INTO development_requests(tenant_ref,workspace_id,attempt_id,request_id,kind,payload_digest) VALUES(?,?,?,?,?,?) ON CONFLICT DO NOTHING`, identity.TenantRef, identity.WorkspaceID, attemptID, op.request.RequestID, op.request.Kind, op.request.PayloadDigest)
		if err != nil {
			return ErrUnavailable
		}
	}
	for _, offer := range offers {
		_, err := conn.ExecContext(ctx, `INSERT INTO development_tool_offers(tenant_ref,workspace_id,attempt_id,request_id,parent_request_id,tool_call_id,payload_digest,arguments_digest,state,observation_digest) VALUES(?,?,?,?,?,?,?,?,?,?) ON CONFLICT(tenant_ref,workspace_id,attempt_id,request_id) DO UPDATE SET state=excluded.state,observation_digest=excluded.observation_digest`, identity.TenantRef, identity.WorkspaceID, attemptID, offer.RequestID, offer.ParentRequestID, offer.ToolCallID, offer.PayloadDigest, offer.ArgumentsDigest, offer.State, offer.ObservationDigest)
		if err != nil {
			return ErrUnavailable
		}
	}
	for _, event := range c.events[oldCount:] {
		body, err := json.Marshal(event)
		if err != nil {
			return ErrUnavailable
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO development_events(tenant_ref,workspace_id,attempt_id,sequence,event_record,event_digest) VALUES(?,?,?,?,?,?)`, identity.TenantRef, identity.WorkspaceID, attemptID, event.Sequence, string(body), PayloadDigest(body))
		if err != nil {
			return ErrUnavailable
		}
	}
	return nil
}

func (s *SQLiteStore) authenticate(ctx context.Context, ownerID, attemptID, credential string) (*Record, error) {
	var identity Identity
	err := s.db.QueryRowContext(ctx, `SELECT tenant_ref,workspace_id,account_id FROM development_attempts WHERE attempt_id=? AND owner_instance_id=? AND credential_digest=? AND credential_revoked=0`, attemptID, ownerID, PayloadDigest([]byte(credential))).Scan(&identity.TenantRef, &identity.WorkspaceID, &identity.AccountID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCredential
	}
	if err != nil {
		return nil, ErrUnavailable
	}
	return s.Read(ctx, identity, attemptID)
}
