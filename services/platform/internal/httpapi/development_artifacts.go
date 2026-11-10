package httpapi

import (
	"context"
	"errors"
	"net/http"
	artifact "radishmind.local/services/platform/internal/developmentartifact"
	"time"
)

// ConnectDevelopmentArtifacts is a trusted-host, request-bound integration.
// It registers no route, listener or default grant. source must be trusted.
func (s *Server) ConnectDevelopmentArtifacts(request *http.Request, source artifact.SourceReader) (*artifact.Service, error) {
	session, ok := requireLocalIdentityRequestSession(request)
	if !ok || s.localPersistenceRuntime == nil {
		return nil, artifact.ErrDenied
	}
	auth, failure, _ := s.authorizeWorkspaceScopedPermissions(request, "applications:read")
	if failure != "" || auth.AuthMode != localIdentityAuthMode || auth.SubjectBinding != "user:"+session.userID {
		return nil, artifact.ErrDenied
	}
	store, err := artifact.NewSQLiteStore(s.localPersistenceRuntime.DB())
	if err != nil {
		return nil, err
	}
	authority := developmentArtifactAuthority{s, session.sessionID, session.userID, auth.ResourceBinding.TenantRef, auth.ResourceBinding.WorkspaceID}
	return artifact.NewService(store, authority, source)
}

type developmentArtifactAuthority struct {
	server                                       *Server
	sessionID, accountID, tenantRef, workspaceID string
}

func (a developmentArtifactAuthority) Authorize(ctx context.Context, scope artifact.Scope, permission string) error {
	if scope.Validate() != nil || scope.AccountID != a.accountID || scope.TenantRef != a.tenantRef || scope.WorkspaceID != a.workspaceID {
		return artifact.ErrDenied
	}
	switch permission {
	case artifact.PermissionRead, artifact.PermissionSave, artifact.PermissionExport:
	default:
		return artifact.ErrDenied
	}
	s := a.server
	if s.localIdentityRepository == nil || s.applicationCatalogRepository == nil {
		return artifact.ErrUnavailable
	}
	now := time.Now().UTC()
	_, account, err := s.localIdentityRepository.ReadWebSession(ctx, a.sessionID, now)
	if err != nil {
		return developmentArtifactIdentityError(err)
	}
	if account.UserID != a.accountID {
		return artifact.ErrDenied
	}
	_, err = s.localIdentityRepository.AuthorizeWorkspace(ctx, a.accountID, a.tenantRef, a.workspaceID, []string{"applications:read", artifact.PermissionRead, permission}, now)
	if err != nil {
		return developmentArtifactIdentityError(err)
	}
	actor := "user:" + a.accountID
	app, err := s.applicationCatalogRepository.Read(ApplicationCatalogContext{RequestContext: ctx, RequestID: "development-artifact-authorize", TenantRef: a.tenantRef, WorkspaceID: a.workspaceID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:development-artifact-authorize"}, scope.ApplicationID)
	if err != nil {
		if errors.Is(err, errApplicationCatalogStoreUnavailable) {
			return artifact.ErrUnavailable
		}
		return artifact.ErrDenied
	}
	if app.OwnerSubjectRef != actor || app.ApplicationID != scope.ApplicationID {
		return artifact.ErrDenied
	}
	// Archived apps retain historical reads; new capture requires active state.
	if permission == artifact.PermissionSave && app.LifecycleState != "active" {
		return artifact.ErrDenied
	}
	return nil
}
func developmentArtifactIdentityError(err error) error {
	if errors.Is(err, errLocalIdentityStoreUnavailable) || errors.Is(err, errLocalIdentityContractMismatch) {
		return artifact.ErrUnavailable
	}
	return artifact.ErrDenied
}
