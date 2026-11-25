package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/workspacepolicy"
)

type taskExecutionEnvironment interface {
	control.TaskWorkerEnvironment
	Check(context.Context) error
}

type developmentExecutionHost struct {
	developmentTaskHost
	env taskExecutionEnvironment
}

func (h developmentExecutionHost) CheckTaskEnvironment(ctx context.Context) error {
	if h.env == nil {
		return control.ErrUnavailable
	}
	return h.env.Check(ctx)
}
func (h developmentExecutionHost) Prepare(ctx context.Context, o control.TaskResourceOwner, p control.TaskPlan, d control.Delegation) (control.WorkerLaunch, func() error, error) {
	if h.env == nil {
		return control.WorkerLaunch{}, nil, control.ErrUnavailable
	}
	return h.env.Prepare(ctx, o, p, d)
}

// The fixed model has a separate runtime from the provider probe. Authorization
// is scoped to the user's app and developer grants; preparation only needs write.
type fixedTaskQualification struct{ s *Server }

func (q fixedTaskQualification) Resolve(ctx context.Context, id control.Identity, p control.Plan) (control.Selection, error) {
	if p.ModelRef != fixedTaskModel || p.BaselineDigest != fixedTaskBaseline || p.ExecutionProfile != "fixed_go_single_agent_v1" {
		return control.Selection{}, control.ErrIneligible
	}
	actor := "user:" + id.AccountID
	now := q.s.localIdentityHTTPService.nowUTC()
	auth, err := q.s.localIdentityRepository.AuthorizeWorkspace(ctx, id.AccountID, id.TenantRef, id.WorkspaceID, []string{"applications:read", workspacepolicy.PermissionDevelopmentTasksRead, workspacepolicy.PermissionDevelopmentTasksWrite}, now)
	if err != nil {
		if errors.Is(err, errLocalIdentityStoreUnavailable) || errors.Is(err, errLocalIdentityContractMismatch) {
			return control.Selection{}, control.ErrUnavailable
		}
		return control.Selection{}, control.ErrIneligible
	}
	app, err := q.s.applicationCatalogRepository.Read(ApplicationCatalogContext{RequestContext: ctx, RequestID: "fixed-task-qualification", TenantRef: id.TenantRef, WorkspaceID: id.WorkspaceID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:fixed-task-qualification"}, p.ApplicationID)
	if errors.Is(err, errApplicationCatalogNotFound) {
		return control.Selection{}, control.ErrNotFound
	}
	if err != nil {
		return control.Selection{}, control.ErrUnavailable
	}
	if app.OwnerSubjectRef != actor || app.ApplicationID != p.ApplicationID || app.WorkspaceID != id.WorkspaceID || app.TenantRef != id.TenantRef || app.LifecycleState != applicationCatalogLifecycleActive {
		return control.Selection{}, control.ErrIneligible
	}
	roles := slices.Clone(auth.RoleAssignments)
	slices.SortFunc(roles, compareDevelopmentRoles)
	for i := range roles {
		roles[i].PermissionGrants = slices.Sorted(slices.Values(roles[i].PermissionGrants))
	}
	identity, err := json.Marshal([]any{id, auth.Account.RecordVersion, auth.Membership, roles})
	if err != nil {
		return control.Selection{}, control.ErrUnavailable
	}
	application, err := json.Marshal([]any{app.ApplicationID, app.WorkspaceID, app.OwnerSubjectRef, app.RecordVersion, app.LifecycleState})
	if err != nil {
		return control.Selection{}, control.ErrUnavailable
	}
	return control.Selection{
		ApplicationAuthorityDigest: control.PayloadDigest(application), IdentityDigest: control.PayloadDigest(identity),
		ModelDigest: control.PayloadDigest([]byte(fixedTaskModel)), WireModel: fixedTaskModel,
		PricingDigest: control.PayloadDigest([]byte("fixed-response-v1:1000000/1000000/1000000")),
		TokenUnit:     1000000, InputPriceMicros: 1000000, OutputPriceMicros: 1000000,
	}, nil
}

func (s *Server) fixedTaskRuntimeLocked() (*control.Runtime, error) {
	store, err := s.developmentStoreLocked()
	if err != nil {
		return nil, err
	}
	if s.developmentTaskRuntime == nil {
		s.developmentTaskRuntime, err = control.NewRuntime(store, fixedTaskQualification{s}, fixedTaskBackend{})
	}
	return s.developmentTaskRuntime, err
}

func (d developmentTaskRequest) coordinator() (*control.TaskCoordinator, developmentExecutionHost, error) {
	s := d.host.server
	s.developmentMu.Lock()
	r, err := s.fixedTaskRuntimeLocked()
	h := developmentExecutionHost{d.host, s.taskEnvironment}
	s.developmentMu.Unlock()
	if err != nil {
		return nil, h, err
	}
	c, err := control.NewTaskCoordinator(r, h)
	return c, h, err
}

func (d developmentTaskRequest) prepare() (any, error) {
	var in struct {
		Request string `json:"request_id"`
		Version int64  `json:"expected_version"`
		Digest  string `json:"plan_digest"`
	}
	if err := d.executionInput(control.TaskPrepare, &in, "request_id", "expected_version", "plan_digest"); err != nil {
		return nil, err
	}
	c, _, err := d.coordinator()
	if err != nil {
		return nil, err
	}
	ctx, s, task := d.request.Context(), d.host.scope, d.request.PathValue("task_id")
	receipt, err := c.Prepare(ctx, s, task, in.Request, in.Version, in.Digest)
	if err != nil {
		return nil, err
	}
	return c.ReadPreparation(ctx, s, task, receipt.PreparationID)
}
func (d developmentTaskRequest) start() (any, error) {
	var in struct {
		Request     string `json:"request_id"`
		Version     int64  `json:"expected_version"`
		Preparation string `json:"preparation_id"`
		Digest      string `json:"confirmation_digest"`
	}
	if err := d.executionInput(control.TaskStart, &in, "request_id", "expected_version", "preparation_id", "confirmation_digest"); err != nil {
		return nil, err
	}
	c, host, err := d.coordinator()
	if err != nil {
		return nil, err
	}
	receipt, _, err := c.StartWorker(d.request.Context(), host.scope, d.request.PathValue("task_id"), in.Request, in.Version, in.Preparation, in.Digest, host)
	if err != nil {
		return nil, err
	}
	return map[string]string{"task_id": receipt.TaskID, "attempt_id": receipt.AttemptID}, nil
}
func (d developmentTaskRequest) executionInput(op control.TaskOperation, target any, fields ...string) error {
	if !d.access.eligibility(op).Allowed {
		return control.ErrIneligible
	}
	return decodeDevelopmentPlanRequest(d.request, target, fields...)
}
func (d developmentTaskRequest) attempt() (any, error) {
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	return m.AttemptView(d.request.Context(), d.host.scope, d.request.PathValue("task_id"), d.request.PathValue("attempt_id"))
}
func (d developmentTaskRequest) cancel() (any, error) {
	var in struct {
		Request string `json:"request_id"`
		Version int64  `json:"expected_version"`
	}
	if err := d.executionInput(control.TaskCancel, &in, "request_id", "expected_version"); err != nil {
		return nil, err
	}
	c, _, err := d.coordinator()
	if err != nil {
		return nil, err
	}
	ctx, s, task, attempt := d.request.Context(), d.host.scope, d.request.PathValue("task_id"), d.request.PathValue("attempt_id")
	if _, err = c.Cancel(ctx, s, task, attempt, in.Request, in.Version); err != nil {
		return nil, err
	}
	v, err := c.AttemptView(ctx, s, task, attempt)
	if err != nil {
		return nil, err
	}
	return map[string]any{"cancellation_requested": v.CancelRequested, "attempt": v}, nil
}

func (s *Server) enableWorkbenchTaskChannel(env *workbenchTaskEnvironment) error {
	s.developmentMu.Lock()
	defer s.developmentMu.Unlock()
	r, err := s.fixedTaskRuntimeLocked()
	if err != nil {
		return err
	}
	server, address, done, err := serveDevelopmentChannel(context.Background(), r)
	if err != nil {
		return err
	}
	env.workers.ChannelURL = address
	s.taskEnvironment = env
	s.developmentTaskHTTP = server
	watched := make(chan struct{})
	s.developmentTaskWatchDone = watched
	go func() {
		defer close(watched)
		if err := <-done; err == nil {
			return
		}
		s.developmentMu.Lock()
		s.taskEnvironment = nil
		s.developmentTaskRuntime = nil // Further requests can only replay, read or cancel.
		s.developmentMu.Unlock()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = r.Close(ctx)
	}()
	return nil
}
