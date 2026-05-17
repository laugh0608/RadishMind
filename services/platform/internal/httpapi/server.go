package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/config"
)

const serviceName = "radishmind-platform"

type Options struct {
	BuildVersion string
}

type bridgeClient interface {
	DescribeProviders(ctx context.Context) ([]bridge.ProviderDescription, error)
	DescribeInventory(ctx context.Context) (bridge.ProviderInventory, error)
	HandleEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions) (bridge.GatewayEnvelope, error)
	StreamEnvelope(ctx context.Context, canonicalRequest []byte, options bridge.EnvelopeOptions, handleEvent func(bridge.StreamEvent) error) error
}

type Server struct {
	httpServer *http.Server
	options    Options
	bridge     bridgeClient
	config     config.Config
}

type errorDocument struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Message         string         `json:"message"`
	Type            string         `json:"type"`
	Code            string         `json:"code"`
	RequestID       string         `json:"request_id,omitempty"`
	Route           string         `json:"route,omitempty"`
	FailureBoundary string         `json:"failure_boundary,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
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
	mux.HandleFunc("GET /v1/models/{id}", server.handleModel)
	mux.HandleFunc("POST /v1/chat/completions", server.handleChatCompletions)
	mux.HandleFunc("POST /v1/responses", server.handleResponses)
	mux.HandleFunc("POST /v1/messages", server.handleMessages)
	mux.HandleFunc("GET /v1/session/recovery/checkpoints/{checkpoint_id}", server.handleSessionRecoveryCheckpoint)

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

func (s *Server) handleModels(writer http.ResponseWriter, request *http.Request) {
	handleModels(writer, request, s)
}

func (s *Server) handleModel(writer http.ResponseWriter, request *http.Request) {
	handleModel(writer, request, s)
}

func writeJSON(writer http.ResponseWriter, statusCode int, document any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(document)
}
