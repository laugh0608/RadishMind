package main

import (
	"errors"
	"log"
	"net/http"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/httpapi"
)

var buildVersion = "dev"

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server := httpapi.NewServer(cfg, httpapi.Options{
		BuildVersion: buildVersion,
	})

	log.Printf("radishmind platform service listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve platform api: %v", err)
	}
}
