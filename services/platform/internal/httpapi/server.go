package httpapi

import (
	"encoding/json"
	"net/http"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const serviceName = "radishmind-platform"

type Options struct {
	BuildVersion string
}

type Server struct {
	httpServer *http.Server
	options    Options
	bridge     *bridge.Client
	config     config.Config
}

type errorDocument struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func NewServer(cfg config.Config, options Options) *Server {
	mux := http.NewServeMux()
	server := &Server{
		options: options,
		bridge:  bridge.NewClient(cfg.PythonBinary, cfg.BridgeScript),
		config:  cfg,
	}

	mux.HandleFunc("GET /healthz", server.handleHealthz)
	mux.HandleFunc("GET /v1/models", server.handleModels)
	mux.HandleFunc("POST /v1/chat/completions", server.handleChatCompletions)

	server.httpServer = &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}
	return server
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": serviceName,
		"version": s.options.BuildVersion,
		"path":    request.URL.Path,
	})
}

func (s *Server) handleModels(writer http.ResponseWriter, _ *http.Request) {
	handleModels(writer, s)
}

func writeJSON(writer http.ResponseWriter, statusCode int, document any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(document)
}

func writeOpenAIError(writer http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(writer, statusCode, map[string]any{
		"error": map[string]any{
			"message": message,
			"type":    "invalid_request_error",
			"code":    code,
		},
	})
}
