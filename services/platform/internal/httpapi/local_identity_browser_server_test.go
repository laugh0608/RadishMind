//go:build identity_browser_test

package httpapi

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	sqlitelocalidentitymigrations "radishmind.local/services/platform/migrations/sqlite/local_identity_records"
)

// TestLocalIdentityBrowserServer is an opt-in host for the existing identity
// handlers, middleware and SQLite repository. It does not enable unrelated
// product gates or change the platform's production startup configuration.
func TestLocalIdentityBrowserServer(t *testing.T) {
	if os.Getenv("RADISHMIND_IDENTITY_BROWSER_TEST") != "1" {
		t.Skip("started only by the isolated identity browser runner")
	}
	databasePath := os.Getenv("RADISHMIND_SQLITE_DEV_DATABASE_PATH")
	if !filepath.IsAbs(databasePath) {
		t.Fatal("identity browser host requires an absolute temporary database path")
	}
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqlitelocalidentitymigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open identity browser database: %v", err)
	}
	t.Cleanup(func() {
		if err := runtime.Close(); err != nil {
			t.Errorf("close identity browser database: %v", err)
		}
	})
	repository := newSQLiteLocalIdentityRepository(runtime.DB())
	cfg := config.Config{
		ControlPlaneReadDevAuthEnabled: true,
		ControlPlaneReadAuthMode:       localIdentityAuthMode,
		LocalIdentityDevHTTPEnabled:    true,
		LocalIdentityStoreMode:         localIdentityStoreModeSQLiteDev,
		LocalIdentityAllowedOrigin:     "http://127.0.0.1:4100",
		LocalIdentitySessionTTL:        time.Hour,
	}
	identity := newLocalIdentityHTTPService(cfg, repository)
	server := &Server{
		config:                                  cfg,
		localIdentityHTTPService:                identity,
		localIdentityAdministrationService:      newLocalIdentityAdministrationService(repository),
		localIdentitySelfServiceSecurityService: newLocalIdentitySelfServiceSecurityService(repository),
		workspaceInvitationService:              newWorkspaceInvitationService(repository),
	}
	mux := http.NewServeMux()
	registerLocalIdentityHTTPRoutes(mux, server)
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		if _, err := writer.Write([]byte(`{"status":"ok","profile":"identity-browser-test"}`)); err != nil {
			t.Errorf("write identity browser readiness: %v", err)
		}
	})
	listener := &http.Server{
		Addr: "127.0.0.1:17000",
		Handler: withLocalConsoleCORS(
			withLocalIdentitySessionAuthentication(withControlPlaneReadAuthenticator(mux, &controlPlaneReadAuthenticator{mode: localIdentityAuthMode}), identity), cfg,
		),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	stopped, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- listener.ListenAndServe() }()
	select {
	case err := <-serveErrors:
		if err != nil {
			t.Fatalf("identity browser listener: %v", err)
		}
	case <-stopped.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := listener.Shutdown(ctx); err != nil {
			t.Errorf("stop identity browser listener: %v", err)
		}
		if err := <-serveErrors; !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("identity browser listener stopped unexpectedly: %v", err)
		}
	}
}
