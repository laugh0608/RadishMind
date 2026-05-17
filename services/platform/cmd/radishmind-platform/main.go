package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/diagnostics"
	"radishmind.local/services/platform/internal/httpapi"
)

var buildVersion = "dev"

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config-summary":
			writeConfigSummary(cfg)
			return
		case "config-check":
			writeConfigCheck(cfg)
			return
		case "diagnostics":
			writeDiagnostics(cfg)
			return
		default:
			log.Fatalf("unsupported command: %s", os.Args[1])
		}
	}

	server := httpapi.NewServer(cfg, httpapi.Options{
		BuildVersion: buildVersion,
	})

	log.Printf("radishmind platform service listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve platform api: %v", err)
	}
}

func writeConfigSummary(cfg config.Config) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg.SanitizedSummary()); err != nil {
		log.Fatalf("write config summary: %v", err)
	}
}

func writeConfigCheck(cfg config.Config) {
	summary := cfg.SanitizedSummary()
	status := "ok"
	exitCode := 0
	if len(summary.MissingRequiredFields) > 0 {
		status = "error"
		exitCode = 1
	}

	document := map[string]any{
		"status":                  status,
		"missing_required_fields": summary.MissingRequiredFields,
		"config":                  summary,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(document); err != nil {
		log.Fatalf("write config check: %v", err)
	}
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func writeDiagnostics(cfg config.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.BridgeTimeout)
	defer cancel()
	report := diagnostics.BuildReport(
		ctx,
		cfg,
		bridge.NewClient(cfg.PythonBinary, cfg.BridgeScript),
		time.Now().UTC(),
	)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatalf("write diagnostics: %v", err)
	}
	if report.Status != "ok" {
		os.Exit(1)
	}
}
