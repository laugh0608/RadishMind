package developmentcontrol

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	artifacts "radishmind.local/services/platform/migrations/sqlite/development_artifacts"
	attempts "radishmind.local/services/platform/migrations/sqlite/development_attempts"
	tasks "radishmind.local/services/platform/migrations/sqlite/development_tasks"
)

type taskFixture struct {
	*TaskCoordinator
	db         *sqlitedev.Runtime
	path       string
	scope      TaskScope
	choice     TaskChoice
	profile    TaskProfile
	now        time.Time
	validUntil time.Time
	deny       TaskOperation
	reader     *currentQualification
}

func (f *taskFixture) AuthorizeTask(_ context.Context, _ TaskScope, op TaskOperation) error {
	if f.deny == op {
		return ErrIneligible
	}
	return nil
}
func (f *taskFixture) FixedProfile(_ context.Context, s TaskScope, choice TaskChoice) (TaskProfile, time.Time, error) {
	p := f.profile
	if choice.ModelRef != p.AttemptPlan.ModelRef {
		return TaskProfile{}, time.Time{}, ErrIneligible
	}
	p.AttemptPlan.ApplicationID = s.ApplicationID
	p.AttemptPlan.Limits = choice.Limits
	return p, f.validUntil, nil
}

func newTaskFixture(t *testing.T, taskMigrations ...sqlitedev.Migration) *taskFixture {
	t.Helper()
	sha := PayloadDigest([]byte("test fixed response"))
	choice := TaskChoice{"profile:fixture", Limits{4, 3, 1000, 1000}}
	plan := Plan{"development_attempt_plan.v1", "application_fixture", "prompt_invoke_v2", "", choice.ModelRef, sha, sha, ToolPolicy{[]string{"go.mod", "list.go", "list_test.go"}, []string{"list.go"}, "go_test_v1"}, choice.Limits, 60000, 50}
	f := &taskFixture{scope: TaskScope{Identity{"tenant_fixture", "account_fixture", "workspace_fixture"}, plan.ApplicationID}, choice: choice, profile: TaskProfile{plan, taskSource, "source_v1", "fixed_response_only", sha, sha, sha, "retention_v1"}, now: time.Now(), path: filepath.Join(t.TempDir(), "tasks.db"), reader: &currentQualification{selection: Selection{sha, sha, sha, "fixture-model", sha, 1000000, 100, 100}}}
	f.validUntil = f.now.Add(time.Hour)
	f.open(t, taskMigrations...)
	t.Cleanup(func() {
		if err := f.db.Close(); err != nil {
			t.Error(err)
		}
	})
	return f
}

func (f *taskFixture) open(t *testing.T, taskMigrations ...sqlitedev.Migration) {
	t.Helper()
	if len(taskMigrations) == 0 {
		taskMigrations = tasks.Migrations()
	}
	ms := append(attempts.Migrations(), artifacts.Migrations()...)
	db, err := sqlitedev.Open(context.Background(), sqlitedev.Options{DatabasePath: f.path, Migrations: append(ms, taskMigrations...)})
	if err != nil {
		t.Fatal(err)
	}
	f.db = db
	store, err := NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewRuntime(store, f.reader, &fixedNative{})
	if err != nil {
		t.Fatal(err)
	}
	r.now = func() time.Time { return f.now }
	f.TaskCoordinator, err = NewTaskCoordinator(r, f)
	if err != nil {
		t.Fatal(err)
	}
}

func (f *taskFixture) create(t *testing.T, key string) TaskReceipt {
	t.Helper()
	r, err := f.Create(t.Context(), f.scope, key, taskSource, f.choice)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func (f *taskFixture) preview(t *testing.T, task, key string) TaskPreparation {
	t.Helper()
	ctx := t.Context()
	current, err := f.ReadTask(ctx, f.scope, task)
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.ReadPlan(ctx, f.scope, task, current.Revision)
	if err != nil {
		t.Fatal(err)
	}
	r, err := f.Prepare(ctx, f.scope, task, key, current.Version, p.Digest)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := f.ReadPreparation(ctx, f.scope, task, r.PreparationID)
	if err != nil {
		t.Fatal(err)
	}
	return preview
}
func (f *taskFixture) start(ctx context.Context, p TaskPreparation, req string) (TaskReceipt, *Issued, error) {
	return f.Start(ctx, f.scope, p.TaskID, req, p.TaskVersion, p.PreparationID, p.ConfirmationDigest)
}
func (f *taskFixture) execSQL(t *testing.T, sql string) {
	t.Helper()
	if _, err := f.db.DB().Exec(sql); err != nil {
		t.Fatal(err)
	}
}
func (f *taskFixture) count(t *testing.T, sql string) int {
	t.Helper()
	var n int
	if err := f.db.DB().QueryRow(sql).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTaskPlanRevisionScopeAndConfirmation(t *testing.T) {
	f := newTaskFixture(t)
	ctx := t.Context()
	r := f.create(t, "create_one")
	if replay := f.create(t, "create_one"); replay != r {
		t.Fatal("create replay changed the receipt")
	}
	f.validUntil = f.now.Add(time.Minute)
	old := f.preview(t, r.TaskID, "prepare_one")
	if old.ExpiresAtMS != f.validUntil.UnixMilli() {
		t.Fatal("confirmation exceeds qualification lifetime")
	}
	choice := f.choice
	choice.Limits.MaxModelCalls--
	if _, err := f.Create(ctx, f.scope, "create_one", taskSource, choice); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if _, err := f.Revise(ctx, f.scope, r.TaskID, "create_one", 1, choice); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	rev, err := f.Revise(ctx, f.scope, r.TaskID, "revise_one", 1, choice)
	if err != nil || rev.Revision != 2 {
		t.Fatal(rev, err)
	}
	if _, _, err = f.start(ctx, old, "stale_start"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if again, err := f.Revise(ctx, f.scope, r.TaskID, "revise_one", 1, choice); err != nil || again != rev {
		t.Fatal(again, err)
	}
	first, err := f.ReadPlan(ctx, f.scope, r.TaskID, 1)
	if err != nil || first.AttemptPlan.Limits != f.choice.Limits {
		t.Fatal(first, err)
	}
	for _, other := range []TaskScope{
		{Identity{"tenant_other", f.scope.AccountID, f.scope.WorkspaceID}, f.scope.ApplicationID},
		{Identity{f.scope.TenantRef, "account_other", f.scope.WorkspaceID}, f.scope.ApplicationID},
		{Identity{f.scope.TenantRef, f.scope.AccountID, "workspace_other"}, f.scope.ApplicationID},
		{f.scope.Identity, "application_other"},
	} {
		if _, err := f.ReadTask(ctx, other, r.TaskID); !errors.Is(err, ErrNotFound) {
			t.Fatal(err)
		}
		if _, err := f.ReadPlan(ctx, other, r.TaskID, 1); !errors.Is(err, ErrNotFound) {
			t.Fatal(err)
		}
		if _, err := f.ReadPreparation(ctx, other, r.TaskID, old.PreparationID); !errors.Is(err, ErrNotFound) {
			t.Fatal(err)
		}
		if _, err := f.ReadReceipt(ctx, other, r.RequestID); !errors.Is(err, ErrNotFound) {
			t.Fatal(err)
		}
	}
	if _, err := f.ReadAttempt(ctx, f.scope, r.TaskID, "attempt_other"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	fresh := f.preview(t, r.TaskID, "prepare_new")
	f.profile.GoImageDigest = PayloadDigest([]byte("changed"))
	if _, _, err := f.start(ctx, fresh, "start_changed"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	f.profile.GoImageDigest = fresh.Plan.GoImageDigest
	f.reader.selection.PricingDigest = PayloadDigest([]byte("changed pricing"))
	if _, _, err := f.start(ctx, fresh, "start_changed"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	f.now = f.validUntil
	if _, _, err := f.start(ctx, fresh, "start_expired"); !errors.Is(err, ErrTaskExpired) {
		t.Fatal(err)
	}
	if f.count(t, "SELECT count(*) FROM development_attempts") != 0 {
		t.Fatal("invalid confirmations issued an attempt")
	}
}

func TestTaskStartAtomicCommitConcurrencyAndCancellation(t *testing.T) {
	f := newTaskFixture(t)
	ctx := t.Context()
	p := f.preview(t, f.create(t, "create_one").TaskID, "prepare_one")
	f.execSQL(t, `CREATE TRIGGER reject_receipt BEFORE INSERT ON development_task_receipts WHEN NEW.operation='start' BEGIN SELECT RAISE(ABORT,'fault'); END`)
	if _, issued, err := f.start(ctx, p, "start_one"); !errors.Is(err, ErrUnavailable) || issued != nil {
		t.Fatal(issued != nil, err)
	}
	if f.count(t, "SELECT count(*) FROM development_attempts") != 0 || f.count(t, "SELECT count(*) FROM development_events") != 0 || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1 OR attempt_id IS NOT NULL") != 0 || f.count(t, "SELECT sum(consumed) FROM development_task_preparations") != 0 || len(f.runtime.deadlines) != 0 {
		t.Fatal("partial startup survived rollback")
	}
	f.execSQL(t, "DROP TRIGGER reject_receipt")
	var issuedCount atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, issued, err := f.start(ctx, p, "start_one")
			if err != nil || r.TaskID != p.TaskID || r.AttemptID == "" {
				t.Error(r, err)
				return
			}
			if issued != nil {
				issuedCount.Add(1)
				s := issued.Delegation.Scope
				if s.RunID != p.TaskID || s.TaskID != p.TaskID || s.AttemptID != r.AttemptID || s.AttemptID == s.TaskID || issued.Credential == "" {
					t.Error("wrong issuance binding")
				}
			}
		}()
	}
	wg.Wait()
	if issuedCount.Load() != 1 || f.count(t, "SELECT count(*) FROM development_events") != 1 {
		t.Fatal("multiple issuances")
	}
	if _, _, err := f.start(ctx, p, "start_other"); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	other := f.preview(t, f.create(t, "create_two").TaskID, "prepare_two")
	if _, _, err := f.start(ctx, other, "start_two"); !errors.Is(err, ErrSlotBusy) {
		t.Fatal(err)
	}
	f.now = f.now.Add(10 * time.Minute)
	f.reader.deny(ErrIneligible)
	replay, issued, err := f.start(ctx, p, "start_one")
	if err != nil || issued != nil {
		t.Fatal("lost response required a live model or new credential", err)
	}
	f.deny = TaskStart
	if _, _, err := f.start(ctx, p, "start_one"); !errors.Is(err, ErrIneligible) {
		t.Fatal("revoked write replayed", err)
	}
	f.deny = TaskRead
	if _, err := f.ReadReceipt(ctx, f.scope, "start_one"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	f.execSQL(t, `CREATE TRIGGER reject_cancel BEFORE INSERT ON development_task_receipts WHEN NEW.operation='cancel' BEGIN SELECT RAISE(ABORT,'fault'); END`)
	if _, err := f.Cancel(ctx, f.scope, p.TaskID, replay.AttemptID, "cancel_one", 1); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	record, err := f.runtime.Read(ctx, f.scope.Identity, replay.AttemptID)
	if err != nil || record.Revoked || record.Version != 1 {
		t.Fatal("partial cancel", err)
	}
	f.execSQL(t, "DROP TRIGGER reject_cancel")
	r, err := f.Cancel(ctx, f.scope, p.TaskID, replay.AttemptID, "cancel_one", 1)
	if err != nil {
		t.Fatal(err)
	}
	if again, err := f.Cancel(ctx, f.scope, p.TaskID, replay.AttemptID, "cancel_one", 1); err != nil || again != r {
		t.Fatal(again, err)
	}
	if err := f.ReconcileSlot(ctx); err != nil || f.count(t, "SELECT count(*) FROM development_tasks WHERE execution_slot=1") != 1 {
		t.Fatal("cancel intention released slot", err)
	}
}

func TestTaskCrashRecoveryDoesNotReissueOrFreeUnknownResources(t *testing.T) {
	f := newTaskFixture(t, tasks.Migrations()[:1]...)
	ctx := t.Context()
	p := f.preview(t, f.create(t, "create_one").TaskID, "prepare_one")
	rec, issued, err := f.start(ctx, p, "start_one")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "http://127.0.0.1/", nil)
	req.SetPathValue("attempt_id", rec.AttemptID)
	req.Header.Set("Authorization", "Bearer "+issued.Credential)
	if _, err := f.runtime.requestRecord(req); !errors.Is(err, ErrCredential) {
		t.Fatal("channel accepted bearer before worker claim", err)
	}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	f.open(t)
	err = RecoverWorkbenchTasks(ctx, f.store, func(context.Context, TaskResourceOwner) (bool, error) {
		t.Fatal("pre-migration attempt has no ledger to inspect")
		return false, ErrUnavailable
	})
	if !errors.Is(err, ErrTaskTerminationUnconfirmed) {
		t.Fatal("upgrade released an unverified execution slot", err)
	}
	status, err := f.ReadTaskResources(ctx, f.scope, rec.TaskID, rec.AttemptID)
	if err != nil || status != (TaskResourceStatus{"unconfirmed", "unconfirmed", "unconfirmed"}) {
		t.Fatal(status, err)
	}
	record, err := f.runtime.Read(ctx, f.scope.Identity, rec.AttemptID)
	if err != nil || record.Control.phase != "outcome_unknown" || !record.Revoked || record.OwnerInstanceID == f.runtime.instanceID {
		t.Fatal("old active record not fenced", err)
	}
	if _, err := f.runtime.store.authenticate(ctx, f.runtime.instanceID, rec.AttemptID, issued.Credential); !errors.Is(err, ErrCredential) {
		t.Fatal(err)
	}
	if rr, ii, err := f.start(ctx, p, "start_one"); err != nil || ii != nil || rr != rec {
		t.Fatal("recovery repeated work", err)
	}
	other := f.preview(t, f.create(t, "create_two").TaskID, "prepare_two")
	if _, _, err := f.start(ctx, other, "start_two"); !errors.Is(err, ErrSlotBusy) {
		t.Fatal(err)
	}
	current, err := f.ReadAttempt(ctx, f.scope, rec.TaskID, rec.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Cancel(ctx, f.scope, rec.TaskID, rec.AttemptID, "cancel_after_restart", current.Version); err != nil {
		t.Fatal("terminal cancellation", err)
	}
	badOwner := func(_ context.Context, owner TaskResourceOwner) (TaskResourceReport, error) {
		owner.OwnerInstanceID = f.runtime.instanceID
		return TaskResourceReport{owner, true, true, true}, nil
	}
	if err := f.RecoverPreviousSlot(ctx, badOwner); !errors.Is(err, ErrTaskTerminationUnconfirmed) {
		t.Fatal(err)
	}
	for _, ended := range []bool{false, false, true} {
		err := f.RecoverPreviousSlot(ctx, func(_ context.Context, owner TaskResourceOwner) (TaskResourceReport, error) {
			if owner.Scope != f.scope || owner.TaskID != rec.TaskID || owner.OwnerInstanceID != record.OwnerInstanceID || owner.Sandbox != SandboxBindingFor(issued.Delegation) || owner.GoImageDigest != p.Plan.GoImageDigest {
				t.Fatal("wrong cleanup target")
			}
			return TaskResourceReport{owner, true, ended, true}, nil
		})
		if ended && err != nil || !ended && !errors.Is(err, ErrTaskTerminationUnconfirmed) {
			t.Fatal(ended, err)
		}
		if _, _, err = f.start(ctx, other, "start_two"); (!ended && !errors.Is(err, ErrSlotBusy)) || (ended && err != nil) {
			t.Fatal(ended, err)
		}
	}
}

func TestTaskPlanAndReceiptCapacity(t *testing.T) {
	f := newTaskFixture(t)
	ctx := t.Context()
	r := f.create(t, "create_one")
	for revision := 1; revision <= 16; revision++ {
		_, err := f.Revise(ctx, f.scope, r.TaskID, fmt.Sprintf("rev_%d", revision), int64(revision), f.choice)
		if revision < 16 && err != nil || revision == 16 && !errors.Is(err, ErrTaskCapacity) {
			t.Fatal(revision, err)
		}
	}
	p := f.preview(t, r.TaskID, "first_preview")
	for i := 17; i < 64; i++ {
		f.preview(t, r.TaskID, fmt.Sprintf("prep_%d", i))
	}
	if _, err := f.Prepare(ctx, f.scope, r.TaskID, "overflow", 16, p.Plan.Digest); !errors.Is(err, ErrTaskCapacity) {
		t.Fatal(err)
	}
	if _, err := f.Prepare(ctx, f.scope, r.TaskID, "first_preview", 16, p.Plan.Digest); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < 256; i++ {
		f.create(t, fmt.Sprintf("other_%d", i))
	}
	otherApp := TaskScope{f.scope.Identity, "application_other"}
	if _, err := f.Create(ctx, otherApp, "new_app", taskSource, f.choice); !errors.Is(err, ErrTaskCapacity) {
		t.Fatal(err)
	}
}

func TestTaskSlotSerializesIndependentWriters(t *testing.T) {
	f := newTaskFixture(t)
	ctx := t.Context()
	first := f.preview(t, f.create(t, "create_first").TaskID, "preview_first")
	second := f.preview(t, f.create(t, "create_second").TaskID, "preview_second")
	store, err := NewSQLiteStore(f.db.DB())
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntime(store, f.reader, &fixedNative{})
	if err != nil {
		t.Fatal(err)
	}
	runtime.now = f.runtime.now
	other, err := NewTaskCoordinator(runtime, f)
	if err != nil {
		t.Fatal(err)
	}
	var wins, busy atomic.Int32
	var wg sync.WaitGroup
	gate := make(chan struct{})
	for i, c := range []*TaskCoordinator{f.TaskCoordinator, other} {
		p := []TaskPreparation{first, second}[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			_, issued, err := c.Start(ctx, f.scope, p.TaskID, "start_"+p.TaskID, p.TaskVersion, p.PreparationID, p.ConfirmationDigest)
			switch {
			case err == nil && issued != nil:
				wins.Add(1)
			case errors.Is(err, ErrSlotBusy) && issued == nil:
				busy.Add(1)
			default:
				t.Error(err)
			}
		}()
	}
	close(gate)
	wg.Wait()
	if wins.Load() != 1 || busy.Load() != 1 || f.count(t, "SELECT sum(consumed) FROM development_task_preparations") != 1 {
		t.Fatal("SQLite did not serialize independent writers")
	}
}
