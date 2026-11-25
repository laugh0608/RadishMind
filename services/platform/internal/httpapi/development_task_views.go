package httpapi

import (
	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
)

func (d developmentTaskRequest) capabilities() (any, error) {
	if !d.access.eligibility(control.TaskRead).Allowed {
		return nil, control.ErrIneligible
	}
	create := d.access.eligibility(control.TaskCreate)
	if create.Allowed {
		m, err := d.metadata()
		if err != nil {
			return nil, err
		}
		available, err := m.HasCreateCapacity(d.request.Context(), d.host.scope)
		if err != nil {
			return nil, err
		}
		if !available {
			create = taskEligibility{false, "capacity_exceeded"}
		}
	}
	return map[string]any{
		"source_profile_id": fixedTaskSource, "source_version": fixedTaskVersion, "baseline_digest": fixedTaskBaseline, "mode": "fixed_response_only",
		"model_refs": []string{fixedTaskModel}, "limits": fixedTaskLimits(), "max_command_timeout_ms": 120000, "max_attempt_lifetime_ms": 600000,
		"eligibility": map[string]taskEligibility{"create_task": create, "read_tasks": d.access.eligibility(control.TaskRead), "read_artifacts": d.access.eligibility(control.TaskReadArtifact)},
	}, nil
}

type developmentTaskDetail struct {
	control.TaskSnapshot
	Eligibility map[string]taskEligibility `json:"eligibility"`
}

func (d developmentTaskRequest) taskDetail(m *control.TaskMetadata, taskID string, creation bool) (developmentTaskDetail, error) {
	ctx := d.request.Context()
	snapshot, err := m.Snapshot(ctx, d.host.scope, taskID)
	if err != nil {
		return developmentTaskDetail{}, err
	}
	revision := d.access.eligibility(control.TaskRevise)
	prepare := d.access.eligibility(control.TaskPrepare)
	start := d.access.eligibility(control.TaskStart)
	cancel := d.access.eligibility(control.TaskCancel)
	preview := d.access.eligibility(control.TaskReadArtifact)
	save := d.access.eligibility(control.TaskSaveArtifact)
	disable := func(e *taskEligibility, reason string) {
		if e.Allowed {
			*e = taskEligibility{false, reason}
		}
	}
	if creation && snapshot.Task.Version != 1 {
		// A lost creation response refers to revision one even after later writes.
		snapshot.Plan, err = m.ReadPlan(ctx, d.host.scope, taskID, 1)
		if err != nil {
			return developmentTaskDetail{}, err
		}
		snapshot.Task.Revision, snapshot.Task.Version = 1, 1
		snapshot.Task.UpdatedAtMS, snapshot.Task.AttemptID = snapshot.Task.CreatedAtMS, ""
		snapshot.SavedArtifactIDs = []string{}
		for _, e := range []*taskEligibility{&revision, &prepare, &start} {
			disable(e, "plan_changed")
		}
	}
	if snapshot.Task.AttemptID != "" {
		for _, e := range []*taskEligibility{&revision, &prepare, &start} {
			disable(e, "task_already_started")
		}
	} else {
		disable(&cancel, "attempt_not_active")
	}
	if snapshot.Task.Revision >= 16 || snapshot.ReceiptCount >= 64 {
		disable(&revision, "capacity_exceeded")
	}
	for _, e := range []*taskEligibility{&prepare, &start, &cancel} {
		disable(e, "execution_disabled")
	}
	disable(&preview, "artifact_unavailable")
	disable(&save, "artifact_unavailable")
	return developmentTaskDetail{snapshot, map[string]taskEligibility{"revise_plan": revision, "prepare": prepare, "start": start, "cancel": cancel, "preview_artifact": preview, "save_artifact": save}}, nil
}

func (d developmentTaskRequest) create() (any, error) {
	if !d.access.eligibility(control.TaskCreate).Allowed {
		return nil, control.ErrIneligible
	}
	var in struct {
		control.TaskChoice
		RequestID string `json:"request_id"`
		Source    string `json:"source_profile_id"`
	}
	if err := decodeDevelopmentPlanRequest(d.request, &in, "request_id", "source_profile_id", "model_ref", "limits"); err != nil {
		return nil, err
	}
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	receipt, err := m.Create(d.request.Context(), d.host.scope, in.RequestID, in.Source, in.TaskChoice)
	if err != nil {
		return nil, err
	}
	return d.taskDetail(m, receipt.TaskID, true)
}

func (d developmentTaskRequest) revise() (any, error) {
	if !d.access.eligibility(control.TaskRevise).Allowed {
		return nil, control.ErrIneligible
	}
	var in struct {
		control.TaskChoice
		RequestID string `json:"request_id"`
		Version   int64  `json:"expected_version"`
	}
	if err := decodeDevelopmentPlanRequest(d.request, &in, "request_id", "expected_version", "model_ref", "limits"); err != nil {
		return nil, err
	}
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	ctx := d.request.Context()
	receipt, err := m.Revise(ctx, d.host.scope, d.request.PathValue("task_id"), in.RequestID, in.Version, in.TaskChoice)
	if err != nil {
		return nil, err
	}
	return m.ReadPlan(ctx, d.host.scope, receipt.TaskID, receipt.Revision)
}

func (d developmentTaskRequest) tasks() (any, error) {
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	return m.List(d.request.Context(), d.host.scope, d.query.Get("cursor"), d.limit)
}

func (d developmentTaskRequest) detail() (any, error) {
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	return d.taskDetail(m, d.request.PathValue("task_id"), false)
}

func (d developmentTaskRequest) receipt() (any, error) {
	m, err := d.metadata()
	if err != nil {
		return nil, err
	}
	return m.ReadReceipt(d.request.Context(), d.host.scope, d.request.PathValue("request_id"))
}

func (d developmentTaskRequest) artifactService() (*artifact.Service, artifact.Scope, error) {
	s := d.host.scope
	scope := artifact.Scope{TenantRef: s.TenantRef, WorkspaceID: s.WorkspaceID, AccountID: s.AccountID, ApplicationID: s.ApplicationID}
	service, err := d.host.server.ConnectDevelopmentArtifacts(d.request, nil)
	return service, scope, err
}
func (d developmentTaskRequest) artifacts() (any, error) {
	s, scope, err := d.artifactService()
	if err != nil {
		return nil, err
	}
	return s.List(d.request.Context(), scope, d.query.Get("cursor"), d.limit)
}
func (d developmentTaskRequest) artifact() (any, error) {
	s, scope, err := d.artifactService()
	if err != nil {
		return nil, err
	}
	return s.Read(d.request.Context(), scope, d.request.PathValue("artifact_id"))
}
func (d developmentTaskRequest) export() (any, error) {
	s, scope, err := d.artifactService()
	if err != nil {
		return nil, err
	}
	return s.Export(d.request.Context(), scope, d.request.PathValue("artifact_id"), d.query.Get("request_id"))
}
