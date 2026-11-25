package developmentartifact

import (
	"context"
	"errors"
	"testing"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/sqlitedev"
	artifacts "radishmind.local/services/platform/migrations/sqlite/development_artifacts"
	attempts "radishmind.local/services/platform/migrations/sqlite/development_attempts"
	tasks "radishmind.local/services/platform/migrations/sqlite/development_tasks"
)

type fixedTaskHost struct {
	profile control.TaskProfile
	deny    control.TaskOperation
}

func (h *fixedTaskHost) AuthorizeTask(_ context.Context, _ control.TaskScope, op control.TaskOperation) error {
	if op == h.deny {
		return control.ErrIneligible
	}
	return nil
}
func (h *fixedTaskHost) FixedProfile(context.Context, control.TaskScope, control.TaskChoice) (control.TaskProfile, time.Time, error) {
	return h.profile, time.Now().Add(time.Hour), nil
}
func (h *fixedTaskHost) Resolve(context.Context, control.Identity, control.Plan) (control.Selection, error) {
	d := h.profile.SDKImageDigest
	return control.Selection{ApplicationAuthorityDigest: d, IdentityDigest: d, ModelDigest: d, WireModel: "fixture", PricingDigest: d, TokenUnit: 1000000, InputPriceMicros: 100, OutputPriceMicros: 100}, nil
}
func (h *fixedTaskHost) Policy(context.Context, string) (control.NativePolicy, error) {
	return control.NativePolicy{InputTokenCeiling: 100, MaxOutputTokens: 100, Timeout: time.Second}, nil
}
func (h *fixedTaskHost) Complete(context.Context, string, control.NativeRequest) (control.NativeResult, error) {
	panic("saving a task must never dispatch the model")
}

func TestTaskSaveAndReceiptCommitTogether(t *testing.T) {
	ctx := t.Context()
	svc, _, src, old, path := newServiceFixture(t)
	if err := old.Close(); err != nil {
		t.Fatal(err)
	}
	ms := append(attempts.Migrations(), artifacts.Migrations()...)
	db, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: path, Migrations: append(ms, tasks.Migrations()...)})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	svc.store, err = NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	store, err := control.NewSQLiteStore(db.DB())
	if err != nil {
		t.Fatal(err)
	}
	src.in.Plan.Limits.MaxModelCalls = 4
	src.in.Plan.Limits.MaxToolCalls = 3
	d := src.in.Plan.InitialMessagesDigest
	host := &fixedTaskHost{profile: control.TaskProfile{AttemptPlan: src.in.Plan, SourceProfileID: "probe_list_filter", SourceVersion: "source_v1", ExecutionMode: "fixed_response_only", SDKImageDigest: d, GoImageDigest: d, ExecutorDigest: d, RetentionPolicyVersion: "retention_v1"}}
	coordinator := func() *control.TaskCoordinator {
		r, err := control.NewRuntime(store, host, host)
		if err != nil {
			t.Fatal(err)
		}
		c, err := control.NewTaskCoordinator(r, host)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	c := coordinator()
	s := control.TaskScope{Identity: control.Identity{TenantRef: src.in.Scope.TenantRef, AccountID: src.in.Scope.AccountID, WorkspaceID: src.in.Scope.WorkspaceID}, ApplicationID: src.in.Scope.ApplicationID}
	created, err := c.Create(ctx, s, "create_save", "probe_list_filter", control.TaskChoice{ModelRef: src.in.Plan.ModelRef, Limits: src.in.Plan.Limits})
	if err != nil {
		t.Fatal(err)
	}
	task := created.TaskID
	plan, err := c.ReadPlan(ctx, s, task, 1)
	if err != nil {
		t.Fatal(err)
	}
	confirmation, err := c.Prepare(ctx, s, task, "prepare_save", 1, plan.Digest)
	if err != nil {
		t.Fatal(err)
	}
	prep, err := c.ReadPreparation(ctx, s, task, confirmation.PreparationID)
	if err != nil {
		t.Fatal(err)
	}
	start, issued, err := c.Start(ctx, s, task, "start_save", 1, prep.PreparationID, prep.ConfirmationDigest)
	if err != nil {
		t.Fatal(err)
	}
	src.in.Origin = issued.Delegation.Scope
	src.in.AttemptState = "outcome_unknown"
	bundle := mustBuild(t, src.in)
	save := func() (control.TaskReceipt, error) {
		return c.SaveArtifact(ctx, s, task, start.AttemptID, "save_task", bundle.Digest, svc)
	}
	if _, err := save(); !errors.Is(err, control.ErrConflict) {
		t.Fatal("live resources allowed save", err)
	}
	c = coordinator()
	if err := c.RecoverPreviousSlot(ctx, func(_ context.Context, o control.TaskResourceOwner) (control.TaskResourceReport, error) {
		return control.TaskResourceReport{Owner: o, SDKExited: true, CodeEnvironmentExited: true, ModelConnectionClosed: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().Exec(`CREATE TRIGGER reject_save_receipt BEFORE INSERT ON development_task_receipts WHEN NEW.operation='save_artifact' BEGIN SELECT RAISE(ABORT,'fault'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := save(); !errors.Is(err, control.ErrUnavailable) {
		t.Fatal("receipt fault", err)
	}
	var count int
	if err := db.DB().QueryRow(`SELECT count(*) FROM development_artifacts`).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial package insert", count, err)
	}
	if _, err := db.DB().Exec(`DROP TRIGGER reject_save_receipt`); err != nil {
		t.Fatal(err)
	}
	r, err := save()
	if err != nil {
		t.Fatal(err)
	}
	saved, err := svc.Read(ctx, src.in.Scope, r.ArtifactID)
	if err != nil || saved.Bundle.Digest != bundle.Digest || saved.Bundle.Origin != issued.Delegation.Scope {
		t.Fatal("incorrect saved package", err)
	}
	calls := src.calls
	src.failure = ErrExpired
	svc.now = func() time.Time { return src.in.SourceCreatedAt.Add(48 * time.Hour) }
	if again, err := save(); err != nil || again != r || src.calls != calls {
		t.Fatal("save replay read expired source", err)
	}
	host.deny = control.TaskSaveArtifact
	if _, err := save(); !errors.Is(err, control.ErrIneligible) {
		t.Fatal("revoked write replayed", err)
	}
	host.deny = control.TaskReadArtifact
	if _, err := c.ReadReceipt(ctx, s, "save_task"); !errors.Is(err, control.ErrNotFound) {
		t.Fatal("receipt readable without artifact grant", err)
	}
	host.deny = control.TaskRead
	if again, err := c.ReadReceipt(ctx, s, "save_task"); err != nil || again != r {
		t.Fatal("artifact read required task grant", err)
	}
}
