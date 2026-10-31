package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/workspacepolicy"
)

// ConnectDevelopmentControl is an explicit trusted-host integration point.
// Normal server construction does not enable execution or start a listener.
func (s *Server) ConnectDevelopmentControl(backend developmentcontrol.NativeBackend) (*developmentcontrol.Runtime, error) {
	s.developmentMu.Lock()
	defer s.developmentMu.Unlock()
	if s.developmentClosed || s.localPersistenceRuntime == nil || backend == nil {
		return nil, developmentcontrol.ErrUnavailable
	}
	if s.developmentRuntime != nil {
		return nil, developmentcontrol.ErrConflict
	}
	store, err := developmentcontrol.NewSQLiteStore(s.localPersistenceRuntime.DB())
	if err != nil {
		return nil, err
	}
	runtime, err := developmentcontrol.NewRuntime(store, developmentQualificationReader{s}, backend)
	if err != nil {
		return nil, err
	}
	s.developmentRuntime = runtime
	return runtime, nil
}

func (s *Server) developmentIssuer(request *http.Request) (developmentcontrol.Identity, *developmentcontrol.Runtime, error) {
	session, valid := requireLocalIdentityRequestSession(request)
	if !valid || !session.expiresAt.After(time.Now()) {
		return developmentcontrol.Identity{}, nil, developmentcontrol.ErrCredential
	}
	auth, failure, _ := s.authorizeWorkspaceScopedPermissions(request, "applications:read", workspacepolicy.PermissionRolesAssign)
	if failure != "" || auth.AuthMode != localIdentityAuthMode || auth.SubjectBinding != "user:"+session.userID {
		return developmentcontrol.Identity{}, nil, developmentcontrol.ErrIneligible
	}
	s.developmentMu.Lock()
	runtime, closed := s.developmentRuntime, s.developmentClosed
	s.developmentMu.Unlock()
	if runtime == nil || closed {
		return developmentcontrol.Identity{}, nil, developmentcontrol.ErrUnavailable
	}
	return developmentcontrol.Identity{TenantRef: auth.ResourceBinding.TenantRef, AccountID: session.userID, WorkspaceID: auth.ResourceBinding.WorkspaceID}, runtime, nil
}

// These Go methods must be called after the host's explicit task confirmation.
// Issued credentials are for trusted SDK handoff, never a browser response.
func (s *Server) PrepareDevelopmentAttempt(request *http.Request, plan developmentcontrol.Plan) (developmentcontrol.Prepared, error) {
	identity, runtime, err := s.developmentIssuer(request)
	if err != nil {
		return developmentcontrol.Prepared{}, err
	}
	return runtime.Prepare(request.Context(), identity, plan)
}
func (s *Server) IssueDevelopmentAttempt(request *http.Request, plan developmentcontrol.Plan, confirmationID string, confirmed developmentcontrol.Prepared) (developmentcontrol.Issued, error) {
	identity, runtime, err := s.developmentIssuer(request)
	if err != nil {
		return developmentcontrol.Issued{}, err
	}
	return runtime.Issue(request.Context(), identity, plan, confirmationID, confirmed.PlanDigest, confirmed.Selection)
}

// StartDevelopmentChannel starts only a dedicated loopback listener. The caller
// owns the host lifecycle and must observe the returned Serve result.
func (s *Server) StartDevelopmentChannel(ctx context.Context) (string, <-chan error, error) {
	s.developmentMu.Lock()
	defer s.developmentMu.Unlock()
	if s.developmentClosed || s.developmentRuntime == nil {
		return "", nil, developmentcontrol.ErrUnavailable
	}
	if ctx.Err() != nil {
		return "", nil, ctx.Err()
	}
	if s.developmentHTTP != nil {
		return "", nil, developmentcontrol.ErrConflict
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return "", nil, developmentcontrol.ErrUnavailable
	}
	server := &http.Server{Handler: s.developmentRuntime.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 135 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 16384, BaseContext: func(net.Listener) context.Context { return ctx }}
	s.developmentHTTP = server
	done := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		done <- err
		close(done)
	}()
	return "http://" + listener.Addr().String(), done, nil
}

func (s *Server) CloseDevelopmentControl(ctx context.Context) error {
	s.developmentMu.Lock()
	s.developmentClosed = true
	server, runtime := s.developmentHTTP, s.developmentRuntime
	s.developmentMu.Unlock()
	var err error
	// Fence/cancel model calls before waiting for HTTP handlers to drain.
	if runtime != nil {
		err = runtime.Close(ctx)
	}
	if server != nil {
		if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
			err = errors.Join(err, shutdownErr, server.Close())
		}
	}
	if err != nil {
		return fmt.Errorf("development channel shutdown: %w", err)
	}
	return nil
}
