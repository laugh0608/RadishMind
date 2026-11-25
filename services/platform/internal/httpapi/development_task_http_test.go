package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
	artifact "radishmind.local/services/platform/internal/developmentartifact"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/workspacepolicy"
)

type taskWebFixture struct {
	t                          *testing.T
	cfg                        config.Config
	s                          *Server
	cookies                    []*http.Cookie
	csrf, user, app, sessionID string
	actor                      LocalIdentityAdministrationActor
}

func (f *taskWebFixture) open() {
	s, err := NewServerWithError(f.cfg, Options{TestOnly: true})
	if err != nil {
		f.t.Fatal(err)
	}
	f.s = s
	f.t.Cleanup(s.Close)
}

func newTaskWeb(t *testing.T) *taskWebFixture {
	t.Helper()
	f := &taskWebFixture{t: t, cfg: aggregateSQLiteDevServerConfig(filepath.Join(t.TempDir(), "tasks.db"))}
	f.cfg.ControlPlaneReadAuthMode = localIdentityAuthMode
	f.cfg.LocalIdentityDevHTTPEnabled = true
	f.cfg.LocalIdentitySessionTTL = time.Hour
	f.cfg.LocalIdentityAllowedOrigin = localIdentityHTTPTestOrigin
	f.cfg.LocalIdentityStoreMode = "sqlite_dev"
	f.open()
	fixture := &localIdentityHTTPTestFixture{handler: f.s.httpServer.Handler}
	_, cookies, registered := registerLocalIdentityHTTPTestAccount(t, fixture, "developer@example.test", "a sufficiently long test password")
	f.cookies = cookies
	f.csrf = localIdentityCookieValue(t, cookies, f.s.localIdentityHTTPService.csrfCookieName())
	f.user = registered.Account.UserID
	f.sessionID = registered.Session.SessionID
	f.actor = LocalIdentityAdministrationActor{UserID: f.user, TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", AuthenticatedAt: time.Now().UTC()}
	_, err := f.s.localIdentityAdministrationService.BootstrapWorkspaceAdministrator(context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, UserID: f.user, AuditRef: "audit:task-bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	f.app = f.createApp()
	return f
}

func (f *taskWebFixture) createApp() string {
	ctx := ApplicationCatalogContext{RequestContext: context.Background(), RequestID: "request_application", TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, ActorRef: "user:" + f.user, OwnerSubjectRef: "user:" + f.user, AuditRef: "audit:task-app", WriteEnabled: true}
	result := newApplicationCatalogService(f.s.applicationCatalogRepository).Create(ctx, ApplicationCatalogCreateInput{DisplayName: "Development fixture", ApplicationKind: "workflow_copilot"})
	if result.Record == nil {
		f.t.Fatal(result.FailureCode)
	}
	return result.Record.ApplicationID
}

func (f *taskWebFixture) grant() LocalRoleAssignment {
	d, _ := workspacepolicy.BuiltInRole(workspacepolicy.RoleWorkspaceDeveloper)
	r, err := f.s.localIdentityAdministrationService.AssignWorkspaceRole(context.Background(), f.actor, LocalIdentityAssignWorkspaceRoleInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, UserID: f.user, RoleKey: d.RoleKey, ExpectedCatalogVersion: d.CatalogVersion, ExpectedRoleDefinitionDigest: d.DefinitionDigest, AuditRef: "audit:task-grant"})
	if err != nil {
		f.t.Fatal(err)
	}
	return r
}

func (f *taskWebFixture) request(method, path, body string) *http.Request {
	url := "/v1/user-workspace/applications/" + f.app + "/" + path + "?workspace_id=" + f.actor.WorkspaceID
	if before, after, ok := strings.Cut(path, "?"); ok {
		url = "/v1/user-workspace/applications/" + f.app + "/" + before + "?workspace_id=" + f.actor.WorkspaceID + "&" + after
	}
	r := httptest.NewRequest(method, url, strings.NewReader(body))
	r.Header.Set("Origin", localIdentityHTTPTestOrigin)
	r.Header.Set(localIdentityActiveTenantHeader, f.actor.TenantRef)
	r.Header.Set(activeWorkspaceHeader, f.actor.WorkspaceID)
	r.Header.Set("X-Request-ID", "untrusted-id")
	addLocalIdentityCookies(r, f.cookies)
	if method == "POST" {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set(localIdentityCSRFHeader, f.csrf)
	}
	return r
}
func (f *taskWebFixture) serve(r *http.Request, status int) []byte {
	f.t.Helper()
	w := httptest.NewRecorder()
	f.s.httpServer.Handler.ServeHTTP(w, r)
	if w.Code != status {
		f.t.Fatalf("%s %s: got %d want %d: %s", r.Method, r.URL, w.Code, status, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Request-ID") == "untrusted-id" {
		f.t.Fatal("unsafe response headers", w.Header())
	}
	return w.Body.Bytes()
}
func (f *taskWebFixture) call(method, path, body string, status int) []byte {
	f.t.Helper()
	return f.serve(f.request(method, path, body), status)
}
func taskWebJSON[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

const taskCreateJSON = `{"request_id":"req_create","source_profile_id":"probe_list_filter","model_ref":"fixed_response:probe_list_filter_v1","limits":{"max_model_calls":4,"max_tool_calls":3,"max_tokens":1000,"max_cost_microusd":1000}}`
const taskReviseJSON = `{"request_id":"req_revise","expected_version":1,"model_ref":"fixed_response:probe_list_filter_v1","limits":{"max_model_calls":2,"max_tool_calls":2,"max_tokens":500,"max_cost_microusd":500}}`

func TestDevelopmentTaskHTTPRoundTrip(t *testing.T) {
	f := newTaskWeb(t)
	f.call("GET", "development-capabilities", "", 403)
	grant := f.grant()
	capabilities := taskWebJSON[struct {
		Eligibility map[string]taskEligibility
		Limits      control.Limits
	}](t, f.call("GET", "development-capabilities", "", 200))
	if !capabilities.Eligibility["create_task"].Allowed || capabilities.Limits != fixedTaskLimits() {
		t.Fatal(capabilities)
	}
	created := taskWebJSON[developmentTaskDetail](t, f.call("POST", "development-tasks", taskCreateJSON, 201))
	if created.Task.Version != 1 || created.Task.AttemptID != "" || created.Plan.AttemptPlan.ApplicationID != f.app || created.Eligibility["start"].Reason != "execution_disabled" {
		t.Fatal(created)
	}
	path := "development-tasks/" + created.Task.TaskID
	plan := taskWebJSON[control.TaskPlan](t, f.call("POST", path+"/plan-revisions", taskReviseJSON, 201))
	if plan.Revision != 2 || plan.AttemptPlan.Limits.MaxTokens != 500 {
		t.Fatal(plan)
	}
	f.call("POST", path+"/plan-revisions", strings.Replace(taskReviseJSON, "req_revise", "req_create", 1), 409)
	f.call("POST", path+"/plan-revisions", strings.Replace(taskReviseJSON, "req_revise", "req_stale", 1), 409)
	f.call("POST", "development-tasks", strings.Replace(taskCreateJSON, "1000", "900", 1), 409)
	// Lose both POST acknowledgements, reopen the real server and use the original cookies.
	f.s.Close()
	f.open()
	r := taskWebJSON[control.TaskReceipt](t, f.call("GET", "development-requests/req_create", "", 200))
	if r.TaskID != created.Task.TaskID || r.Operation != control.TaskCreate || r.RequestDigest == "" {
		t.Fatal(r)
	}
	replay := taskWebJSON[developmentTaskDetail](t, f.call("POST", "development-tasks", taskCreateJSON, 201))
	if replay.Task != created.Task || replay.Plan.Digest != created.Plan.Digest || replay.Eligibility["revise_plan"].Reason != "plan_changed" {
		t.Fatal(replay)
	}
	revisedAgain := taskWebJSON[control.TaskPlan](t, f.call("POST", path+"/plan-revisions", taskReviseJSON, 201))
	current := taskWebJSON[developmentTaskDetail](t, f.call("GET", path, "", 200))
	if revisedAgain.Digest != plan.Digest || current.Task.Version != 2 || current.Plan.Digest != plan.Digest {
		t.Fatal(current)
	}
	f.call("POST", "development-tasks", strings.Replace(taskCreateJSON, "req_create", "req_another", 1), 201)
	page := taskWebJSON[control.TaskPage](t, f.call("GET", "development-tasks?limit=1", "", 200))
	if len(page.Tasks) != 1 || !page.HasMore || page.NextCursor == "" {
		t.Fatal(page)
	}
	last := taskWebJSON[control.TaskPage](t, f.call("GET", "development-tasks?limit=1&cursor="+url.QueryEscape(page.NextCursor), "", 200))
	if last.HasMore || len(last.Tasks) != 1 || last.Tasks[0].TaskID == page.Tasks[0].TaskID || last.NextCursor != "" {
		t.Fatal(last)
	}
	original := f.app
	f.app = f.createApp()
	f.call("GET", path, "", 404)
	f.call("GET", "development-requests/req_create", "", 404)
	f.call("GET", "development-tasks?cursor="+page.NextCursor, "", 400)
	f.app = original
	artifacts := taskWebJSON[artifact.Page](t, f.call("GET", "development-artifacts", "", 200))
	if artifacts.Artifacts == nil || len(artifacts.Artifacts) != 0 {
		t.Fatal(artifacts)
	}
	f.call("GET", "development-artifacts?cursor="+page.NextCursor, "", 400)
	var n int
	if err := f.s.localPersistenceRuntime.DB().QueryRow("SELECT count(*) FROM development_attempts").Scan(&n); err != nil || n != 0 || f.s.developmentRuntime != nil {
		t.Fatal("metadata initialized execution", n, err)
	}
	_, err := f.s.localIdentityAdministrationService.RevokeWorkspaceRole(context.Background(), f.actor, LocalIdentityRevokeWorkspaceRoleInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, AssignmentID: grant.AssignmentID, ExpectedVersion: grant.RecordVersion, Confirmed: true, AuditRef: "audit:task-revoke"})
	if err != nil {
		t.Fatal(err)
	}
	f.call("GET", "development-requests/req_create", "", 404)
	f.call("GET", "development-requests/req_missing", "", 404)
	f.call("POST", "development-tasks", taskCreateJSON, 403)
	f.call("GET", "development-artifacts", "", 403)
}

func TestDevelopmentTaskHTTPRejectsInvalidRequests(t *testing.T) {
	f := newTaskWeb(t)
	f.grant()
	for _, body := range []string{
		strings.Replace(taskCreateJSON, "request_id", "REQUEST_ID", 1), strings.Replace(taskCreateJSON, "max_tokens", "Max_Tokens", 1),
		strings.Replace(taskCreateJSON, `"max_tokens":1000`, `"max_tokens":null`, 1), strings.Replace(taskCreateJSON, `"max_tokens":1000`, `"max_tokens":1,"max_tokens":1000`, 1),
		strings.Replace(taskCreateJSON, `"max_tokens":1000,`, "", 1), strings.Replace(taskCreateJSON, "1000", "1001", 1),
		strings.Replace(taskCreateJSON, "probe_list_filter_v1", "other_v1", 1), taskCreateJSON + "{}", taskCreateJSON + string([]byte{0xff}),
	} {
		f.call("POST", "development-tasks", body, 400)
	}
	for _, path := range []string{"development-tasks?limit=51", "development-tasks?limit=01", "development-tasks?limit=1&limit=2", "development-tasks?cursor=", "development-tasks?unknown=true", "development-tasks?cursor=%xx", "development-tasks?workspace_id=workspace_demo", "development-artifacts/abc/export"} {
		f.call("GET", path, "", 400)
	}
	r := f.request("POST", "development-tasks", strings.Repeat("x", taskRequestLimit+1))
	r.ContentLength = -1
	f.serve(r, 413)
	for _, header := range []string{localIdentityCSRFHeader, "Origin"} {
		r = f.request("POST", "development-tasks", taskCreateJSON)
		r.Header.Del(header)
		f.serve(r, 403)
	}
	r = f.request("GET", "development-tasks", "")
	r.Header.Set("Origin", "https://different.invalid")
	f.serve(r, 403)
	f.call("GET", "development-tasks", "{}", 400)
	for _, header := range []string{"Authorization", controlPlaneReadDevSubjectHeader} {
		r = f.request("GET", "development-tasks", "")
		r.Header.Set(header, "Bearer other")
		f.serve(r, 401)
	}
	r = f.request("GET", "development-tasks", "")
	r.Header.Del("Cookie")
	f.serve(r, 401)
	r = f.request("GET", "development-tasks", "")
	r.URL.RawQuery = "workspace_id=workspace_other"
	f.serve(r, 404)
	f.call("HEAD", "development-tasks", "", 405)
}

func TestDevelopmentTaskHTTPSavedArtifactHistory(t *testing.T) {
	f := newTaskWeb(t)
	f.grant()
	ctx := context.Background()
	scope := artifact.Scope{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, AccountID: f.user, ApplicationID: f.app}
	in := developmentArtifactInput(t, scope)
	store, err := artifact.NewSQLiteStore(f.s.localPersistenceRuntime.DB())
	if err != nil {
		t.Fatal(err)
	}
	service, err := artifact.NewService(store, developmentArtifactAuthority{f.s, f.sessionID, f.user, scope.TenantRef, scope.WorkspaceID}, artifactSourceFixture{in})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := service.Preview(ctx, scope, in.Origin.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	records := map[string]bool{}
	for _, id := range []string{"saved_first", "saved_second"} {
		r, err := service.Save(ctx, scope, in.Origin.AttemptID, id, preview.Digest)
		if err != nil {
			t.Fatal(err)
		}
		records[r.ArtifactID] = true
	}
	appCtx := ApplicationCatalogContext{RequestContext: ctx, RequestID: "request_archive", TenantRef: scope.TenantRef, WorkspaceID: scope.WorkspaceID, ActorRef: "user:" + scope.AccountID, OwnerSubjectRef: "user:" + scope.AccountID, AuditRef: "audit:task-archive", WriteEnabled: true}
	if archived := newApplicationCatalogService(f.s.applicationCatalogRepository).Archive(appCtx, f.app, 1); archived.Record == nil {
		t.Fatal(archived.FailureCode)
	}
	f.call("POST", "development-tasks", taskCreateJSON, 403)
	f.s.Close()
	f.open()
	page := taskWebJSON[artifact.Page](t, f.call("GET", "development-artifacts?limit=1", "", 200))
	if len(page.Artifacts) != 1 || !page.HasMore || !records[page.Artifacts[0].ArtifactID] || page.Artifacts[0].Digest != preview.Digest {
		t.Fatal(page)
	}
	last := taskWebJSON[artifact.Page](t, f.call("GET", "development-artifacts?limit=1&cursor="+page.NextCursor, "", 200))
	if len(last.Artifacts) != 1 || last.HasMore || last.Artifacts[0].ArtifactID == page.Artifacts[0].ArtifactID || !records[last.Artifacts[0].ArtifactID] {
		t.Fatal(last)
	}
	path := "development-artifacts/" + page.Artifacts[0].ArtifactID
	rec := taskWebJSON[artifact.Record](t, f.call("GET", path, "", 200))
	if rec.Bundle.Digest != preview.Digest || rec.ArtifactID != page.Artifacts[0].ArtifactID {
		t.Fatal(rec.ArtifactID)
	}
	export, err := artifact.ParseExport(f.call("GET", path+"/export?request_id=export_again", "", 200))
	if err != nil || export.Record.ArtifactID != rec.ArtifactID || export.RequestID != "export_again" {
		t.Fatal("invalid export", err)
	}
	// Simulate at-rest damage in the isolated database, outside the immutable store.
	if _, err = f.s.localPersistenceRuntime.DB().Exec("DROP TRIGGER development_artifacts_no_update"); err != nil {
		t.Fatal(err)
	}
	if _, err = f.s.localPersistenceRuntime.DB().Exec("UPDATE development_artifacts SET record_digest=? WHERE artifact_id=?", control.PayloadDigest([]byte("damaged")), rec.ArtifactID); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"development-artifacts", path, path + "/export?request_id=export_again"} {
		result := taskWebJSON[errorDocument](t, f.call("GET", path, "", 503))
		if result.Error.Code != artifact.ErrCorrupt.Error() {
			t.Fatal(result)
		}
	}
}

func TestDevelopmentTaskHTTPOwnerIsolation(t *testing.T) {
	f := newTaskWeb(t)
	f.grant()
	created := taskWebJSON[developmentTaskDetail](t, f.call("POST", "development-tasks", taskCreateJSON, 201))
	_, cookies, other := registerLocalIdentityHTTPTestAccount(t, &localIdentityHTTPTestFixture{handler: f.s.httpServer.Handler}, "second@example.test", "another sufficiently long password")
	_, err := f.s.localIdentityAdministrationService.CreateWorkspaceMembership(context.Background(), f.actor, LocalIdentityCreateWorkspaceMembershipInput{TenantRef: f.actor.TenantRef, WorkspaceID: f.actor.WorkspaceID, UserID: other.Account.UserID, AuditRef: "audit:other-member"})
	if err != nil {
		t.Fatal(err)
	}
	f.user = other.Account.UserID
	f.grant()
	f.cookies = cookies
	f.csrf = localIdentityCookieValue(t, cookies, f.s.localIdentityHTTPService.csrfCookieName())
	for _, path := range []string{"development-tasks/" + created.Task.TaskID, "development-requests/req_create", "development-requests/req_missing", "development-artifacts"} {
		f.call("GET", path, "", 404)
	}
	f.call("POST", "development-tasks", taskCreateJSON, 404)
}
