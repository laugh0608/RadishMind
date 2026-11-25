package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/strictjson"
)

const taskRequestLimit = 16 << 10

var errTaskPayloadLarge = errors.New("development task payload too large")

type developmentTaskRequest struct {
	request *http.Request
	host    developmentTaskHost
	access  developmentTaskAccess
	query   url.Values
	limit   int
}

type developmentTaskAction func(developmentTaskRequest) (any, error)

func registerDevelopmentTaskRoutes(mux *http.ServeMux, s *Server) {
	const base = "/v1/user-workspace/applications/{application_id}/"
	register := func(method, path string, status, maxBytes int, action developmentTaskAction, queryKeys ...string) {
		mux.HandleFunc(method+" "+base+path, func(w http.ResponseWriter, r *http.Request) {
			trace := newServerGeneratedRequestTrace(r, base+path)
			w.Header().Set("Cache-Control", "no-store")
			if r.Method != method {
				w.Header().Set("Allow", method)
				writeObservedJSON(w, http.StatusMethodNotAllowed, trace, errorDocument{Error: errorBody{Message: "Method not allowed", Type: "invalid_request_error", Code: "development_task_invalid", RequestID: trace.requestID}})
				return
			}
			input, err := s.developmentTaskRequest(r, queryKeys, path == "development-requests/{request_id}")
			if strings.HasPrefix(path, "development-artifacts") && err == control.ErrNotFound {
				err = artifact.ErrNotFound
			}
			var doc any
			if err == nil {
				doc, err = action(input)
			}
			var body []byte
			if err == nil {
				body, err = json.Marshal(doc)
				if len(body)+1 > maxBytes {
					err = control.ErrTaskCorrupt
				}
			}
			if err != nil {
				writeDevelopmentTaskFailure(w, trace, err)
				return
			}
			writeTraceHeaders(w, trace)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			_, _ = w.Write(append(body, '\n'))
			logRequestTrace(trace, status, "", "")
		})
	}
	const metadataLimit = 64 << 10
	register("GET", "development-capabilities", 200, metadataLimit, developmentTaskRequest.capabilities)
	register("POST", "development-tasks", 201, metadataLimit, developmentTaskRequest.create)
	register("GET", "development-tasks", 200, metadataLimit, developmentTaskRequest.tasks, "cursor", "limit")
	register("GET", "development-tasks/{task_id}", 200, metadataLimit, developmentTaskRequest.detail)
	register("POST", "development-tasks/{task_id}/plan-revisions", 201, metadataLimit, developmentTaskRequest.revise)
	register("GET", "development-requests/{request_id}", 200, metadataLimit, developmentTaskRequest.receipt)
	register("GET", "development-artifacts", 200, metadataLimit, developmentTaskRequest.artifacts, "cursor", "limit")
	register("GET", "development-artifacts/{artifact_id}", 200, 2<<20, developmentTaskRequest.artifact)
	register("GET", "development-artifacts/{artifact_id}/export", 200, 2<<20, developmentTaskRequest.export, "request_id")
}

func (s *Server) developmentTaskRequest(r *http.Request, keys []string, receipt bool) (developmentTaskRequest, error) {
	var input developmentTaskRequest
	session, ok := requireLocalIdentityRequestSession(r)
	if !ok {
		if session.failure == localIdentityServiceUnavailable {
			return input, control.ErrUnavailable
		}
		return input, control.ErrCredential
	}
	if s.localIdentityHTTPService == nil || s.localPersistenceRuntime == nil {
		return input, control.ErrUnavailable
	}
	if r.Method == http.MethodPost {
		if len(r.Header.Values("Content-Type")) != 1 {
			return input, control.ErrTaskInvalid
		}
		switch s.localIdentityHTTPService.authenticatedWriteFailure(r, session.csrfToken, session.csrfCookieValid) {
		case "":
		case localIdentityPayloadInvalid:
			return input, control.ErrTaskInvalid
		default:
			return input, control.ErrIneligible
		}
	} else {
		if origins := r.Header.Values("Origin"); len(origins) > 0 && (len(origins) != 1 || origins[0] != s.localIdentityHTTPService.allowedOrigin) {
			return input, control.ErrIneligible
		}
		if r.ContentLength != 0 || len(r.TransferEncoding) > 0 {
			return input, control.ErrTaskInvalid
		}
	}
	app := r.PathValue("application_id")
	if !applicationCatalogIDPattern.MatchString(app) || len(r.URL.RawQuery) > 4096 {
		return input, control.ErrTaskInvalid
	}
	query, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || query.Get("workspace_id") == "" {
		return input, control.ErrTaskInvalid
	}
	for k, values := range query {
		if len(values) != 1 || values[0] == "" || k != "workspace_id" && !slices.Contains(keys, k) {
			return input, control.ErrTaskInvalid
		}
	}
	limit := 20
	if value := query.Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 50 || strconv.Itoa(limit) != value {
			return input, control.ErrTaskInvalid
		}
	}
	auth, failure, status := s.authorizeWorkspaceScopedPermissions(r, "applications:read")
	if failure != "" {
		switch {
		case status == http.StatusServiceUnavailable:
			return input, control.ErrUnavailable
		case status == http.StatusUnauthorized:
			return input, control.ErrCredential
		case status == http.StatusBadRequest:
			return input, control.ErrTaskInvalid
		case receipt && failure == "workspace_permission_denied", failure == "workspace_membership_denied" || failure == "workspace_binding_mismatch":
			return input, control.ErrNotFound
		default:
			return input, control.ErrIneligible
		}
	}
	if auth.AuthMode != localIdentityAuthMode || auth.SubjectBinding != "user:"+session.userID {
		return input, control.ErrCredential
	}
	if auth.ResourceBinding.WorkspaceID != query.Get("workspace_id") {
		return input, control.ErrNotFound
	}
	scope := control.TaskScope{Identity: control.Identity{TenantRef: auth.ResourceBinding.TenantRef, AccountID: session.userID, WorkspaceID: auth.ResourceBinding.WorkspaceID}, ApplicationID: app}
	host := developmentTaskHost{s, session.sessionID, scope}
	access, err := host.access(r.Context())
	if receipt && err == control.ErrIneligible {
		err = control.ErrNotFound
	}
	return developmentTaskRequest{r, host, access, query, limit}, err
}

func (d developmentTaskRequest) metadata() (*control.TaskMetadata, error) {
	s := d.host.server
	s.developmentMu.Lock()
	store, err := s.developmentStoreLocked()
	s.developmentMu.Unlock()
	if err != nil {
		return nil, err
	}
	return control.NewTaskMetadata(store, d.host)
}

func decodeDevelopmentPlanRequest(r *http.Request, target any, fields ...string) error {
	if r.ContentLength > taskRequestLimit {
		return errTaskPayloadLarge
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, taskRequestLimit+1))
	if err != nil {
		return control.ErrTaskInvalid
	}
	if len(raw) > taskRequestLimit {
		return errTaskPayloadLarge
	}
	if strictjson.DecodeBounded(raw, target, taskRequestLimit) != nil {
		return control.ErrTaskInvalid
	}
	values, err := strictjson.ObjectFields(raw, fields...)
	if err != nil {
		return control.ErrTaskInvalid
	}
	if _, err = strictjson.ObjectFields(values["limits"], "max_model_calls", "max_tool_calls", "max_tokens", "max_cost_microusd"); err != nil {
		return control.ErrTaskInvalid
	}
	return nil
}

func writeDevelopmentTaskFailure(w http.ResponseWriter, trace requestTrace, err error) {
	status := http.StatusServiceUnavailable
	code := "development_task_unavailable"
	switch err {
	case control.ErrTaskInvalid:
		status, code = 400, "development_task_invalid"
	case control.ErrCredential:
		status, code = 401, "development_task_denied"
	case control.ErrIneligible:
		status, code = 403, "development_task_denied"
	case control.ErrNotFound:
		status, code = 404, "development_task_not_found"
	case control.ErrConflict:
		status, code = 409, "development_task_conflict"
	case errTaskPayloadLarge:
		status, code = 413, "development_task_invalid"
	case control.ErrTaskCapacity:
		status, code = 429, "development_task_capacity_exceeded"
	case control.ErrTaskCorrupt:
		status, code = 503, "development_task_corrupt"
	case artifact.ErrInvalid:
		status, code = 400, artifact.ErrInvalid.Error()
	case artifact.ErrDenied:
		status, code = 403, artifact.ErrDenied.Error()
	case artifact.ErrNotFound:
		status, code = 404, artifact.ErrNotFound.Error()
	case artifact.ErrCorrupt, artifact.ErrUnavailable:
		code = err.Error()
	}
	writeObservedJSON(w, status, trace, errorDocument{Error: errorBody{Message: http.StatusText(status), Type: "development_task_error", Code: code, RequestID: trace.requestID, Route: trace.route}})
}
