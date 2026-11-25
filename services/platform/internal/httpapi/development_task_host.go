package httpapi

import (
	"context"
	"errors"
	"slices"
	"time"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/workspacepolicy"
)

const (
	fixedTaskModel    = "fixed_response:probe_list_filter_v1"
	fixedTaskSource   = "probe_list_filter"
	fixedTaskVersion  = "task_list_filter_v1"
	fixedTaskBaseline = "sha256:a46359b152f0284e7973792632c1f88a216170485d9947e9d7aa72aff48e45d9"
)

func fixedTaskLimits() control.Limits {
	return control.Limits{MaxModelCalls: 4, MaxToolCalls: 3, MaxTokens: 1000, MaxCostMicrousd: 1000}
}

type developmentTaskHost struct {
	server    *Server
	sessionID string
	scope     control.TaskScope
}
type developmentTaskAccess struct {
	authorization LocalWorkspaceAuthorization
	app           ApplicationCatalogRecord
	expiresAt     time.Time
}

func (h developmentTaskHost) access(ctx context.Context) (developmentTaskAccess, error) {
	s := h.server
	if s.localIdentityRepository == nil || s.applicationCatalogRepository == nil {
		return developmentTaskAccess{}, control.ErrUnavailable
	}
	now := s.localIdentityHTTPService.nowUTC()
	session, acct, err := s.localIdentityRepository.ReadWebSession(ctx, h.sessionID, now)
	if err != nil {
		if errors.Is(err, errLocalIdentityStoreUnavailable) || errors.Is(err, errLocalIdentityContractMismatch) {
			return developmentTaskAccess{}, control.ErrUnavailable
		}
		return developmentTaskAccess{}, control.ErrCredential
	}
	if acct.UserID != h.scope.AccountID {
		return developmentTaskAccess{}, control.ErrCredential
	}
	auth, err := s.localIdentityRepository.AuthorizeWorkspace(ctx, acct.UserID, h.scope.TenantRef, h.scope.WorkspaceID, []string{"applications:read"}, now)
	if err != nil {
		switch {
		case errors.Is(err, errLocalIdentityPermissionDenied):
			return developmentTaskAccess{}, control.ErrIneligible
		case errors.Is(err, errLocalIdentityMembershipDenied), errors.Is(err, errLocalIdentityAccountInactive):
			return developmentTaskAccess{}, control.ErrNotFound
		default:
			return developmentTaskAccess{}, control.ErrUnavailable
		}
	}
	actor := "user:" + acct.UserID
	app, err := s.applicationCatalogRepository.Read(ApplicationCatalogContext{RequestContext: ctx, RequestID: "development-task-access", TenantRef: h.scope.TenantRef, WorkspaceID: h.scope.WorkspaceID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:development-task-access"}, h.scope.ApplicationID)
	if errors.Is(err, errApplicationCatalogNotFound) {
		return developmentTaskAccess{}, control.ErrNotFound
	}
	if err != nil {
		return developmentTaskAccess{}, control.ErrUnavailable
	}
	if app.ApplicationID != h.scope.ApplicationID || app.OwnerSubjectRef != actor || app.TenantRef != h.scope.TenantRef || app.WorkspaceID != h.scope.WorkspaceID {
		return developmentTaskAccess{}, control.ErrNotFound
	}
	expiry := session.ExpiresAt
	if end := auth.Membership.ExpiresAt; end != nil && end.Before(expiry) {
		expiry = *end
	}
	for _, r := range auth.RoleAssignments {
		if r.ExpiresAt != nil && r.ExpiresAt.Before(expiry) {
			expiry = *r.ExpiresAt
		}
	}
	return developmentTaskAccess{auth, app, expiry}, nil
}

type taskEligibility struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func (a developmentTaskAccess) eligibility(op control.TaskOperation) taskEligibility {
	read, write := workspacepolicy.PermissionDevelopmentTasksRead, workspacepolicy.PermissionDevelopmentTasksWrite
	var permissions []string
	switch op {
	case control.TaskRead:
		permissions = []string{read}
	case control.TaskCreate, control.TaskRevise, control.TaskPrepare:
		permissions = []string{read, write}
	case control.TaskStart:
		permissions = []string{read, write, workspacepolicy.PermissionDevelopmentTasksExecute}
	case control.TaskCancel:
		permissions = []string{read, workspacepolicy.PermissionDevelopmentTasksCancel}
	case control.TaskReadArtifact:
		permissions = []string{artifact.PermissionRead}
	case control.TaskSaveArtifact:
		permissions = []string{artifact.PermissionRead, artifact.PermissionSave}
	default:
		return taskEligibility{false, "permission_denied"}
	}
	for _, p := range permissions {
		if !slices.Contains(a.authorization.PermissionGrants, p) {
			return taskEligibility{false, "permission_denied"}
		}
	}
	if a.app.LifecycleState != applicationCatalogLifecycleActive && (op == control.TaskCreate || op == control.TaskRevise || op == control.TaskPrepare || op == control.TaskStart || op == control.TaskSaveArtifact) {
		return taskEligibility{false, "application_archived"}
	}
	return taskEligibility{true, "allowed"}
}

func (h developmentTaskHost) AuthorizeTask(ctx context.Context, s control.TaskScope, op control.TaskOperation) error {
	if h.scope != s {
		return control.ErrNotFound
	}
	a, err := h.access(ctx)
	if err != nil {
		return err
	}
	if !a.eligibility(op).Allowed {
		return control.ErrIneligible
	}
	return nil
}

// These are the November 21 ARM64 joint run's immutable inputs. Planning
// records the profile even on a workbench without the SDK or its images.
func (h developmentTaskHost) FixedProfile(ctx context.Context, s control.TaskScope, choice control.TaskChoice) (control.TaskProfile, time.Time, error) {
	if s != h.scope {
		return control.TaskProfile{}, time.Time{}, control.ErrNotFound
	}
	limits := fixedTaskLimits()
	if choice.ModelRef != fixedTaskModel || choice.Limits.MaxTokens > limits.MaxTokens || choice.Limits.MaxCostMicrousd > limits.MaxCostMicrousd {
		return control.TaskProfile{}, time.Time{}, control.ErrTaskInvalid
	}
	a, err := h.access(ctx)
	if err != nil {
		return control.TaskProfile{}, time.Time{}, err
	}
	plan := control.Plan{
		SchemaVersion: "development_attempt_plan.v1", ApplicationID: s.ApplicationID,
		ExecutionProfile: "fixed_go_single_agent_v1", ModelRef: fixedTaskModel,
		BaselineDigest: fixedTaskBaseline, InitialMessagesDigest: "sha256:23764171adc8a33c1ef4329df088c5b9c4849a30f90e825bd4c83fe30cace4d9",
		Tools:  control.ToolPolicy{ReadPaths: []string{"go.mod", "list.go", "list_test.go"}, WritePaths: []string{"list.go"}, VerificationProfile: "go_test_v1"},
		Limits: choice.Limits, LifetimeMS: 600000, MaxOutputTokens: 50,
	}
	return control.TaskProfile{
		AttemptPlan: plan, SourceProfileID: fixedTaskSource, SourceVersion: fixedTaskVersion, ExecutionMode: "fixed_response_only",
		SDKImageDigest:         "sha256:5e8c4c4473d5497febada6cb563b6630c495b1dcd54651de7c270df38bdd1daf",
		GoImageDigest:          "sha256:b17af760035fc2f338eed92d448a6c67f2d45438844fc6c60678fa5f99e44b57",
		ExecutorDigest:         "sha256:f009798f17e15fb1a9155282ac899e6dab0cb55d8382caa172883606d9b53597",
		RetentionPolicyVersion: "development_artifact_retention_v1",
	}, a.expiresAt, nil
}
