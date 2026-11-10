package workbench

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"radishmind.local/services/platform/internal/config"
)

type PublicConfig struct {
	SchemaVersion int      `json:"schema_version"`
	PublicOrigin  string   `json:"public_origin"`
	TenantRef     string   `json:"tenant_ref"`
	WorkspaceID   string   `json:"workspace_id"`
	Environment   string   `json:"environment"`
	ProviderMode  string   `json:"provider_mode"`
	Features      []string `json:"features"`
}

func PublicConfiguration(cfg config.Config) PublicConfig {
	return PublicConfig{SchemaVersion: 1, PublicOrigin: cfg.Workbench.PublicOrigin, TenantRef: cfg.Workbench.TenantRef,
		WorkspaceID: cfg.Workbench.WorkspaceID, Environment: cfg.Workbench.Environment, ProviderMode: cfg.Provider,
		Features: []string{"applications", "prompt", "workflow", "sessions", "results", "api_keys", "local_identity"}}
}

// CheckAssets verifies the build profile as well as the entry point. Serving a
// legacy Vite build here would silently restore fixture defaults in the browser.
func CheckAssets(directory string) error {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return errors.New("workbench web assets unavailable; run workbench build")
	}
	defer root.Close()
	file, err := root.Open("workbench-build.json")
	if err != nil {
		return errors.New("workbench build marker missing; rebuild Web with the local-workbench profile")
	}
	defer file.Close()
	var marker struct {
		RuntimeProfile string `json:"runtime_profile"`
	}
	decoder := json.NewDecoder(io.LimitReader(file, 1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&marker); err != nil || marker.RuntimeProfile != config.LocalWorkbenchProfile {
		return errors.New("incorrect workbench Web build profile")
	}
	info, err := root.Stat("index.html")
	if err != nil || !info.Mode().IsRegular() {
		return errors.New("workbench index.html missing or invalid")
	}
	return nil
}

func WebHandler(cfg config.Config, api http.Handler) (http.Handler, func() error, error) {
	if err := CheckAssets(cfg.Workbench.WebAssetsPath); err != nil {
		return nil, nil, err
	}
	root, err := os.OpenRoot(cfg.Workbench.WebAssetsPath)
	if err != nil {
		return nil, nil, err
	}
	origin, err := url.Parse(cfg.Workbench.PublicOrigin)
	if err != nil {
		root.Close()
		return nil, nil, err
	}
	public := PublicConfiguration(cfg)
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		if request.URL.Path == "/healthz" {
			api.ServeHTTP(writer, request)
			return
		}
		if request.Host != origin.Host || (request.Header.Get("Origin") != "" && request.Header.Get("Origin") != cfg.Workbench.PublicOrigin) {
			webFailure(writer, http.StatusForbidden, "workbench_origin_mismatch")
			return
		}
		if !safeWebPath(request.URL.Path) {
			webFailure(writer, http.StatusNotFound, "not_found")
			return
		}
		if request.URL.Path == "/workbench-config.json" {
			if request.Method != http.MethodGet {
				webFailure(writer, http.StatusMethodNotAllowed, "method_not_allowed")
				return
			}
			writer.Header().Set("Cache-Control", "no-store")
			writer.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(writer).Encode(public)
			return
		}
		if strings.HasPrefix(request.URL.Path, "/v1/") || request.URL.Path == "/v1" {
			if unsupportedWorkbenchRead(request.URL.Path) {
				webFailure(writer, http.StatusNotFound, "workbench_surface_unavailable")
				return
			}
			api.ServeHTTP(writer, request)
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			webFailure(writer, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		name := strings.TrimPrefix(request.URL.Path, "/")
		if name == "" || (!strings.Contains(path.Base(name), ".") && !strings.HasPrefix(name, "assets/")) {
			name = "index.html"
		}
		if name != "index.html" && name != "favicon.svg" && !strings.HasPrefix(name, "assets/") {
			webFailure(writer, http.StatusNotFound, "not_found")
			return
		}
		file, err := root.Open(name)
		if err != nil {
			webFailure(writer, http.StatusNotFound, "not_found")
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() {
			webFailure(writer, http.StatusNotFound, "not_found")
			return
		}
		writer.Header().Set("Cache-Control", "no-store")
		if strings.HasPrefix(name, "assets/") {
			writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeContent(writer, request, name, info.ModTime(), file)
	})
	return handler, root.Close, nil
}

func safeWebPath(value string) bool {
	if strings.ContainsAny(value, "\\\x00") {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if strings.HasPrefix(part, ".") {
			return false
		}
	}
	return true
}

func unsupportedWorkbenchRead(route string) bool {
	return strings.HasPrefix(route, "/v1/control-plane/") || route == "/v1/user-workspace/usage/quota-summary" ||
		strings.HasPrefix(route, "/v1/platform/") || strings.HasPrefix(route, "/v1/session/") || strings.HasPrefix(route, "/v1/tools/")
}

func webFailure(writer http.ResponseWriter, status int, code string) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(map[string]any{"error": map[string]string{"code": code, "message": code}})
}
