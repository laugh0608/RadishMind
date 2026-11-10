package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/httpapi"
	"radishmind.local/services/platform/internal/workbench"
)

func runWorkbench(arguments []string, stdout, stderr io.Writer) int {
	if err := executeWorkbench(arguments, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "workbench: %v\n", err)
		return 1
	}
	return 0
}

func executeWorkbench(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("expected check, init, serve, bootstrap-admin, backup, upgrade or restore")
	}
	action := arguments[0]
	flags := flag.NewFlagSet("workbench "+action, flag.ContinueOnError)
	flags.SetOutput(stderr)
	defaultConfig := os.Getenv("RADISHMIND_PLATFORM_CONFIG")
	if defaultConfig == "" {
		defaultConfig = "var/workbench/platform.json"
	}
	configPath := flags.String("config", defaultConfig, "explicit workbench JSON configuration")
	output := flags.String("output", "", "new backup or restore destination directory")
	source := flags.String("source", "", "verified backup directory to restore")
	userID := flags.String("user-id", "", "exact existing local account ID for administrator bootstrap")
	auditRef := flags.String("audit-ref", "", "explicit audit reference for administrator bootstrap")
	if err := flags.Parse(arguments[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	allowed := map[string]map[string]bool{
		"check": {"config": true}, "init": {"config": true}, "serve": {"config": true},
		"bootstrap-admin": {"config": true, "user-id": true, "audit-ref": true},
		"backup":          {"config": true, "output": true}, "upgrade": {"config": true, "output": true},
		"restore": {"source": true, "output": true},
	}
	if _, known := allowed[action]; !known {
		return errors.New("unknown workbench action")
	}
	var invalidFlag string
	flags.Visit(func(value *flag.Flag) {
		if !allowed[action][value.Name] {
			invalidFlag = value.Name
		}
	})
	if invalidFlag != "" {
		return fmt.Errorf("--%s is not valid for %s", invalidFlag, action)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if action == "restore" {
		manifest, err := workbench.Restore(ctx, *source, *output)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(map[string]any{"status": "restored", "backup_time": manifest.CreatedAt,
			"directory": *output, "next": "Review configuration, credentials, members and API keys as of the backup time; then explicitly check and serve. Historical tasks are not replayed."})
	}
	if err := os.Setenv("RADISHMIND_PLATFORM_CONFIG", *configPath); err != nil {
		return err
	}
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}
	if cfg.RuntimeProfile != config.LocalWorkbenchProfile {
		return errors.New("command requires runtime_profile local-workbench")
	}
	if err := config.ValidateServerStart(cfg); err != nil {
		return err
	}
	switch action {
	case "check":
		if err := workbench.CheckAssets(cfg.Workbench.WebAssetsPath); err != nil {
			return err
		}
		if _, err := exec.LookPath(cfg.PythonBinary); err != nil {
			return errors.New("configured Python executable is unavailable")
		}
		dependencyCheck := exec.CommandContext(ctx, cfg.PythonBinary, "-I", "-c", "import sys; import jsonschema; assert sys.version_info >= (3, 12)")
		if err := dependencyCheck.Run(); err != nil {
			return errors.New("workbench requires Python >=3.12 with jsonschema installed; select the prepared virtual environment")
		}
		if info, err := os.Stat(cfg.BridgeScript); err != nil || !info.Mode().IsRegular() {
			return errors.New("configured Python bridge script is unavailable")
		}
		return json.NewEncoder(stdout).Encode(map[string]any{"status": "configuration_valid", "runtime_profile": cfg.RuntimeProfile,
			"public_origin": cfg.Workbench.PublicOrigin, "tenant_ref": cfg.Workbench.TenantRef, "workspace_id": cfg.Workbench.WorkspaceID,
			"provider": cfg.Provider, "database_inspected": false, "workers_started": false})
	case "init":
		err = workbench.Initialize(ctx, cfg.SQLiteDevDatabasePath)
	case "backup":
		err = workbench.Backup(ctx, cfg, *output, buildVersion)
	case "upgrade":
		err = workbench.Upgrade(ctx, cfg, *output, buildVersion)
	case "bootstrap-admin":
		if *userID == "" || *auditRef == "" {
			return errors.New("bootstrap-admin requires --user-id and --audit-ref")
		}
		result, err := httpapi.BootstrapLocalIdentityWorkspaceAdministratorDevTest(ctx, httpapi.LocalIdentityDevTestBootstrapOptions{
			Workbench: true,
			StoreMode: "sqlite_dev", SQLiteDatabasePath: cfg.SQLiteDevDatabasePath, DatabaseTimeout: 30 * time.Second,
			TenantRef: cfg.Workbench.TenantRef, WorkspaceID: cfg.Workbench.WorkspaceID, UserID: *userID, AuditRef: *auditRef,
		})
		if err != nil {
			return fmt.Errorf("administrator bootstrap: %s", httpapi.LocalIdentityFailureCode(err))
		}
		return json.NewEncoder(stdout).Encode(result)
	case "serve":
		return serveWorkbench(cfg, stdout, stderr)
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(stdout).Encode(map[string]string{"status": "ok", "action": action})
}

func serveWorkbench(cfg config.Config, stdout, stderr io.Writer) error {
	server, err := httpapi.NewServerWithError(cfg, httpapi.Options{BuildVersion: buildVersion})
	if err != nil {
		return err
	}
	defer server.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serveErrors := make(chan error, 1)
	go func() { serveErrors <- server.ListenAndServe() }()
	fmt.Fprintf(stdout, "Workbench: %s (provider=%s); Ctrl+C stops HTTP and Python workers.\n", cfg.Workbench.PublicOrigin, cfg.Provider)
	select {
	case err := <-serveErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			fmt.Fprintln(stderr, "Workbench shutdown interrupted active work; restart never replays it.")
			return err
		}
	}
	return nil
}
