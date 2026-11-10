package developmentcontrol

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Prepared struct {
	PlanDigest string    `json:"plan_digest"`
	Selection  Selection `json:"selection"`
}
type Issued struct {
	Delegation    Delegation `json:"delegation"`
	Credential    string     `json:"-"`
	RecordVersion int64      `json:"record_version"`
}

type Runtime struct {
	store      *SQLiteStore
	reader     QualificationReader
	backend    NativeBackend
	instanceID string
	now        func() time.Time
	mu         sync.Mutex
	closed     bool
	active     map[string]context.CancelFunc
	workers    map[string]context.CancelFunc
	fenced     map[string]bool
	deadlines  map[string]time.Time
	work       sync.WaitGroup
}

func NewRuntime(store *SQLiteStore, reader QualificationReader, backend NativeBackend) (*Runtime, error) {
	if store == nil || reader == nil {
		return nil, ErrUnavailable
	}
	id, err := randomID("owner_")
	if err != nil {
		return nil, err
	}
	return &Runtime{store: store, reader: reader, backend: backend, instanceID: id, now: time.Now, active: map[string]context.CancelFunc{}, deadlines: map[string]time.Time{}, workers: map[string]context.CancelFunc{}, fenced: map[string]bool{}}, nil
}
func randomID(prefix string) (string, error) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", ErrUnavailable
	}
	return prefix + hex.EncodeToString(data[:]), nil
}
func (r *Runtime) Prepare(ctx context.Context, identity Identity, plan Plan) (Prepared, error) {
	if plan.Validate() != nil || r.backend == nil {
		return Prepared{}, ErrTransport
	}
	selected, err := r.reader.Resolve(ctx, identity, plan)
	if err != nil || selected.Validate() != nil {
		return Prepared{}, ErrIneligible
	}
	policy, err := r.backend.Policy(ctx, plan.ModelRef)
	if err != nil || policy.InputTokenCeiling < 1 || policy.InputTokenCeiling+plan.MaxOutputTokens > plan.Limits.MaxTokens || plan.MaxOutputTokens > policy.MaxOutputTokens || policy.Timeout <= 0 || policy.Timeout > 2*time.Minute {
		return Prepared{}, ErrTransport
	}
	sha, err := plan.Digest()
	if err != nil {
		return Prepared{}, err
	}
	return Prepared{sha, selected}, nil
}
func (r *Runtime) Issue(ctx context.Context, identity Identity, plan Plan, confirmationID, confirmedDigest string, confirmedSelection Selection) (Issued, error) {
	r.mu.Lock()
	closed := r.closed
	r.mu.Unlock()
	if closed {
		return Issued{}, ErrUnavailable
	}
	prepared, err := r.Prepare(ctx, identity, plan)
	if err != nil {
		return Issued{}, err
	}
	if prepared.PlanDigest != confirmedDigest || prepared.Selection != confirmedSelection || !idPattern.MatchString(confirmationID) {
		return Issued{}, ErrConflict
	}
	policy, err := r.backend.Policy(ctx, plan.ModelRef)
	if err != nil {
		return Issued{}, ErrTransport
	}
	binding := Binding{"development_attempt_binding.v1", identity, plan, prepared.Selection, policy.InputTokenCeiling}
	if binding.Validate() != nil {
		return Issued{}, ErrIneligible
	}
	attemptID, err := randomID("attempt_")
	if err != nil {
		return Issued{}, err
	}
	delegationID, err := randomID("delegate_")
	if err != nil {
		return Issued{}, err
	}
	executorID, err := randomID("executor_")
	if err != nil {
		return Issued{}, err
	}
	bindingID, err := randomID("binding_")
	if err != nil {
		return Issued{}, err
	}
	tools, err := json.Marshal(plan.Tools)
	if err != nil {
		return Issued{}, ErrUnavailable
	}
	issuedAt := r.now()
	now := issuedAt.UnixMilli()
	d := Delegation{SchemaVersion: "development_attempt_delegation.v1", DelegationID: delegationID, Revision: 1, Audience: Audience, ExecutorInstanceID: executorID, IssuedAtMS: now, ExpiresAtMS: now + plan.LifetimeMS, Limits: plan.Limits,
		Scope: Scope{AccountID: identity.AccountID, WorkspaceID: identity.WorkspaceID, ApplicationID: plan.ApplicationID, ApplicationAuthorityDigest: prepared.Selection.ApplicationAuthorityDigest, BindingID: bindingID, BindingRevision: 1, AuthorizationRevision: 1, RunID: attemptID, TaskID: attemptID, AttemptID: attemptID, PlanDigest: prepared.PlanDigest, BaselineDigest: plan.BaselineDigest, ModelRef: plan.ModelRef, PricingDigest: prepared.Selection.PricingDigest, ToolPolicyDigest: PayloadDigest(tools)}}
	c, err := NewAttemptControl(d, Authority{Scope: d.Scope, DelegationRevision: d.Revision, ExecutorInstanceID: d.ExecutorInstanceID}, now)
	if err != nil {
		return Issued{}, err
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return Issued{}, ErrUnavailable
	}
	credential := "devtask_" + base64.RawURLEncoding.EncodeToString(secret[:])
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return Issued{}, ErrUnavailable
	}
	if err := r.store.create(ctx, confirmationID, r.instanceID, PayloadDigest([]byte(credential)), binding, c); err != nil {
		return Issued{}, err
	}
	r.deadlines[attemptID] = issuedAt.Add(time.Duration(plan.LifetimeMS) * time.Millisecond)
	return Issued{d, credential, 1}, nil
}

func authorityFor(record *Record) Authority {
	d := record.Control.delegation
	return Authority{Scope: d.Scope, DelegationRevision: d.Revision, ExecutorInstanceID: d.ExecutorInstanceID}
}

// revalidate returns a refusal after recording the fence. Callers commit that
// refusal state, then return the error outside the SQL transaction.
func (r *Runtime) revalidate(ctx context.Context, record *Record) error {
	selected, err := r.reader.Resolve(ctx, record.Binding.Identity, record.Binding.Plan)
	now := r.now().UnixMilli()
	d := record.Control.delegation
	deadline, live := r.attemptDeadline(d.Scope.AttemptID)
	if !record.Revoked && live && r.now().Before(deadline) && err == nil && selected == record.Binding.Selection && now >= d.IssuedAtMS && now < d.ExpiresAtMS {
		return nil
	}
	record.Revoked = true
	if record.Control.phase == "active" || record.Control.phase == "cancel_requested" {
		if err != nil && !errors.Is(err, ErrIneligible) {
			if _, stopErr := record.Control.Interrupt(now); stopErr != nil {
				return stopErr
			}
		} else if record.Control.phase == "active" {
			a := authorityFor(record)
			a.Revoked = true
			if _, stopErr := record.Control.RefreshAuthority(a, now); stopErr != nil {
				return stopErr
			}
		}
	}
	if err != nil {
		return err
	}
	return ErrIneligible
}

func (r *Runtime) attemptDeadline(attemptID string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	deadline, found := r.deadlines[attemptID]
	return deadline, found && !r.closed && !r.fenced[attemptID]
}

func (r *Runtime) Read(ctx context.Context, identity Identity, attemptID string) (*Record, error) {
	return r.store.Read(ctx, identity, attemptID)
}
func (r *Runtime) Cancel(ctx context.Context, identity Identity, attemptID string, expected int64) (*Record, error) {
	record, err := r.store.change(ctx, identity, attemptID, expected, "", func(record *Record) error {
		if record.Control.phase != "active" {
			return AttemptStopped
		}
		if _, err := record.Control.RequestCancel("user_cancelled", r.now().UnixMilli()); err != nil {
			return err
		}
		record.Revoked = true
		return nil
	})
	if err == nil {
		r.cancelActive(attemptID)
	}
	return record, err
}
func (r *Runtime) Interrupt(ctx context.Context, identity Identity, attemptID string, expected int64) (*Record, error) {
	record, err := r.store.change(ctx, identity, attemptID, expected, "", func(record *Record) error {
		if _, err := record.Control.Interrupt(r.now().UnixMilli()); err != nil {
			return err
		}
		record.Revoked = true
		return nil
	})
	if err == nil {
		r.cancelActive(attemptID)
	}
	return record, err
}
func (r *Runtime) cancelActive(attemptID string) {
	r.mu.Lock()
	cancel := r.active[attemptID]
	workerCancel := r.workers[attemptID]
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if workerCancel != nil {
		workerCancel()
	}
}
func (r *Runtime) beginCall(attemptID string, cancel context.CancelFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrUnavailable
	}
	if r.active[attemptID] != nil || len(r.active) >= 16 {
		return ErrConflict
	}
	r.active[attemptID] = cancel
	r.work.Add(1)
	return nil
}
func (r *Runtime) endCall(attemptID string) {
	r.mu.Lock()
	delete(r.active, attemptID)
	r.mu.Unlock()
	r.work.Done()
}

func (r *Runtime) Close(ctx context.Context) error {
	r.mu.Lock()
	r.closed = true
	for _, cancel := range r.workers {
		cancel()
	}
	for _, cancel := range r.active {
		cancel()
	}
	r.mu.Unlock()
	done := make(chan struct{})
	go func() { r.work.Wait(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	rows, err := r.store.db.QueryContext(ctx, `SELECT tenant_ref,workspace_id,account_id,attempt_id FROM development_attempts WHERE owner_instance_id=? AND credential_revoked=0`, r.instanceID)
	if err != nil {
		return ErrUnavailable
	}
	type key struct {
		identity Identity
		attempt  string
	}
	keys := []key{}
	for rows.Next() {
		var k key
		if err := rows.Scan(&k.identity.TenantRef, &k.identity.WorkspaceID, &k.identity.AccountID, &k.attempt); err != nil {
			rows.Close()
			return ErrUnavailable
		}
		keys = append(keys, k)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return ErrUnavailable
	}
	var failure error
	for _, k := range keys {
		_, err := r.store.change(ctx, k.identity, k.attempt, 0, r.instanceID, func(record *Record) error {
			if record.Control.phase == "active" || record.Control.phase == "cancel_requested" {
				if _, err := record.Control.Interrupt(r.now().UnixMilli()); err != nil {
					return err
				}
			}
			record.Revoked = true
			return nil
		})
		failure = errors.Join(failure, err)
	}
	return failure
}
