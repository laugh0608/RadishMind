package httpapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

type workflowHTTPToolRoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn workflowHTTPToolRoundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestWorkflowHTTPToolTransportSuccessUsesReviewedTargetAndHeaders(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	var captured *http.Request
	transport := workflowHTTPToolTestTransport(func(request *http.Request) (*http.Response, error) {
		captured = request.Clone(request.Context())
		return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/alpha","title":"Alpha","summary":"Reviewed summary","updated_at":"2026-07-16T10:00:00Z"}`), nil
	})
	result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_001")
	if result.FailureCode != "" || result.FailureCategory != WorkflowHTTPToolFailureNone {
		t.Fatalf("expected success, got %#v", result)
	}
	if captured == nil || captured.Method != http.MethodGet || captured.URL.Scheme != "https" ||
		captured.URL.Host != "api.dev.example.invalid:443" || captured.URL.Path != "/v1/reviewed-resources" ||
		captured.URL.Query().Get("resource_key") != "docs/alpha" || captured.URL.Query().Get("locale") != "zh-CN" {
		t.Fatalf("unexpected reviewed request: %#v", captured)
	}
	if captured.Header.Get("Accept") != "application/json" || captured.Header.Get("User-Agent") != workflowHTTPToolUserAgent ||
		captured.Header.Get("X-RadishMind-Trace-ID") != "req_workflow_tool_transport_001" || len(captured.Header) != 3 {
		t.Fatalf("unexpected request headers: %#v", captured.Header)
	}
	if result.OutputProjection["resource_key"] != "docs/alpha" || result.HTTPStatusClass != "2xx" || result.ResponseBytes == 0 {
		t.Fatalf("unexpected transport result: %#v", result)
	}
}

func TestWorkflowHTTPToolTransportTestOnlyLoopbackRequiresIndependentGate(t *testing.T) {
	networkAttempts := 0
	loopbackServer := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		networkAttempts++
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"resource_key":"docs/alpha","title":"Alpha","summary":"Loopback test projection","updated_at":"2026-07-17T10:00:00Z"}`))
	}))
	t.Cleanup(loopbackServer.Close)
	serverURL, err := url.Parse(loopbackServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, portText, err := net.SplitHostPort(serverURL.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}

	plan, profile := workflowHTTPToolTransportFixture(t)
	profile.TestOnlyLoopback = true
	profile.TargetPolicy.Host = "workflow-tool.test"
	profile.TargetPolicy.Port = port
	profileDigest, err := canonicalSHA256(profile)
	if err != nil {
		t.Fatal(err)
	}
	plan.ProfileDigest = profileDigest
	plan.ToolPlanDigest, err = workflowHTTPToolPlanDigest(plan)
	if err != nil {
		t.Fatal(err)
	}

	transport := newWorkflowHTTPToolTransport()
	transport.lookupNetIP = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
	}
	loopbackClient := loopbackServer.Client()
	transport.newClient = func(string, int, []netip.Addr, time.Duration) (*http.Client, error) {
		return &http.Client{Transport: workflowHTTPToolRoundTripperFunc(func(request *http.Request) (*http.Response, error) {
			forwarded := request.Clone(request.Context())
			forwarded.URL.Scheme = serverURL.Scheme
			forwarded.URL.Host = serverURL.Host
			return loopbackClient.Transport.RoundTrip(forwarded)
		})}, nil
	}

	denied := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_loopback_denied")
	if denied.FailureCode != WorkflowRunFailureToolPolicy || denied.NetworkAttempted || networkAttempts != 0 {
		t.Fatalf("test-only loopback profile was accepted without the independent gate: %#v network=%d", denied, networkAttempts)
	}
	transport.allowTestOnlyLoopback = true
	allowed := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_loopback_allowed")
	if allowed.FailureCode != "" || !allowed.NetworkAttempted || networkAttempts != 1 ||
		allowed.OutputProjection["resource_key"] != "docs/alpha" {
		t.Fatalf("independently gated test-only loopback profile did not execute once: %#v network=%d", allowed, networkAttempts)
	}
}

func TestWorkflowHTTPToolTransportRejectsEveryForbiddenAddressBeforeClientCreation(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	for _, rawAddress := range []string{
		"0.0.0.0", "10.0.0.1", "100.64.0.1", "127.0.0.1", "169.254.169.254", "172.16.0.1",
		"192.0.2.1", "192.168.1.1", "198.18.0.1", "198.51.100.1", "203.0.113.1", "224.0.0.1",
		"::", "::1", "100::1", "2001:db8::1", "fc00::1", "fe80::1", "ff00::1",
	} {
		t.Run(rawAddress, func(t *testing.T) {
			clientCreated := false
			transport := newWorkflowHTTPToolTransport()
			transport.lookupNetIP = func(context.Context, string, string) ([]netip.Addr, error) {
				return []netip.Addr{netip.MustParseAddr(rawAddress)}, nil
			}
			transport.newClient = func(string, int, []netip.Addr, time.Duration) (*http.Client, error) {
				clientCreated = true
				return nil, errors.New("must not be called")
			}
			result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_002")
			if result.FailureCode != WorkflowRunFailureToolPolicy || result.FailureCategory != WorkflowHTTPToolFailurePolicy || clientCreated {
				t.Fatalf("expected policy rejection before client creation, got %#v created=%t", result, clientCreated)
			}
		})
	}
}

func TestWorkflowHTTPToolTransportRejectsMixedPublicAndPrivateResolution(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	transport := newWorkflowHTTPToolTransport()
	transport.lookupNetIP = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")}, nil
	}
	transport.newClient = func(string, int, []netip.Addr, time.Duration) (*http.Client, error) {
		t.Fatal("client must not be created for mixed DNS results")
		return nil, nil
	}
	result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_003")
	if result.FailureCode != WorkflowRunFailureToolPolicy {
		t.Fatalf("expected policy failure, got %#v", result)
	}
}

func TestWorkflowHTTPToolTransportRejectsBindingDrift(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	mutations := map[string]func(*WorkflowHTTPToolActionPlan, *WorkflowHTTPToolExecutionProfile){
		"profile digest": func(plan *WorkflowHTTPToolActionPlan, _ *WorkflowHTTPToolExecutionProfile) {
			plan.ProfileDigest = "sha256:" + strings.Repeat("0", 64)
		},
		"plan digest": func(plan *WorkflowHTTPToolActionPlan, _ *WorkflowHTTPToolExecutionProfile) {
			plan.PublicArguments.ResourceKey = "docs/drift"
		},
		"redirect": func(_ *WorkflowHTTPToolActionPlan, profile *WorkflowHTTPToolExecutionProfile) {
			profile.NetworkPolicy.FollowRedirects = true
		},
		"proxy": func(_ *WorkflowHTTPToolActionPlan, profile *WorkflowHTTPToolExecutionProfile) {
			profile.NetworkPolicy.UseAmbientProxy = true
		},
		"retry": func(_ *WorkflowHTTPToolActionPlan, profile *WorkflowHTTPToolExecutionProfile) {
			profile.NetworkPolicy.AutomaticRetry = true
		},
		"credential": func(plan *WorkflowHTTPToolActionPlan, _ *WorkflowHTTPToolExecutionProfile) {
			plan.CredentialPolicy = "secret_ref"
		},
		"not approved": func(plan *WorkflowHTTPToolActionPlan, _ *WorkflowHTTPToolExecutionProfile) {
			plan.Status = WorkflowHTTPToolActionStatusPending
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			mutatedPlan, mutatedProfile := plan, profile
			mutate(&mutatedPlan, &mutatedProfile)
			transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
				t.Fatal("network must not be reached after binding drift")
				return nil, nil
			})
			result := transport.Execute(context.Background(), mutatedPlan, mutatedProfile, "req_workflow_tool_transport_004")
			if result.FailureCode != WorkflowRunFailureToolPolicy {
				t.Fatalf("expected policy failure, got %#v", result)
			}
		})
	}
}

func TestWorkflowHTTPToolTransportResponsePolicies(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	tests := []struct {
		name     string
		response func() *http.Response
		code     WorkflowRunFailureCode
		category WorkflowHTTPToolFailureCategory
	}{
		{name: "redirect", response: func() *http.Response { return workflowHTTPToolJSONResponse(http.StatusFound, `{}`) }, code: WorkflowRunFailureToolResponseStatus, category: WorkflowHTTPToolFailureResponseStatus},
		{name: "server error", response: func() *http.Response { return workflowHTTPToolJSONResponse(http.StatusServiceUnavailable, `{}`) }, code: WorkflowRunFailureToolResponseStatus, category: WorkflowHTTPToolFailureResponseStatus},
		{name: "wrong content type", response: func() *http.Response {
			response := workflowHTTPToolJSONResponse(http.StatusOK, `{}`)
			response.Header.Set("Content-Type", "text/plain")
			return response
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "unknown field", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/alpha","title":"Alpha","summary":"Summary","updated_at":"2026-07-16T10:00:00Z","endpoint":"hidden"}`)
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "resource mismatch", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/beta","title":"Alpha","summary":"Summary","updated_at":"2026-07-16T10:00:00Z"}`)
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "endpoint leakage", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/alpha","title":"https://internal.invalid","summary":"Summary","updated_at":"2026-07-16T10:00:00Z"}`)
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "invalid timestamp", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/alpha","title":"Alpha","summary":"Summary","updated_at":"yesterday"}`)
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "trailing document", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, `{"resource_key":"docs/alpha","title":"Alpha","summary":"Summary","updated_at":"2026-07-16T10:00:00Z"} {}`)
		}, code: WorkflowRunFailureToolResponseInvalid, category: WorkflowHTTPToolFailureResponseInvalid},
		{name: "too large", response: func() *http.Response {
			return workflowHTTPToolJSONResponse(http.StatusOK, strings.Repeat("x", workflowHTTPToolMaxResponseBytes+1))
		}, code: WorkflowRunFailureToolResponseTooLarge, category: WorkflowHTTPToolFailureResponseTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) { return test.response(), nil })
			result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_005")
			if result.FailureCode != test.code || result.FailureCategory != test.category || result.OutputProjection != nil {
				t.Fatalf("unexpected result: %#v", result)
			}
		})
	}
}

func TestWorkflowHTTPToolTransportSupportsBoundedGzipResponse(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, _ = writer.Write([]byte(`{"resource_key":"docs/alpha","title":"Alpha","summary":"Summary","updated_at":"2026-07-16T10:00:00Z"}`))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		response := workflowHTTPToolJSONResponse(http.StatusOK, compressed.String())
		response.Header.Set("Content-Encoding", "gzip")
		return response, nil
	})
	result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_006")
	if result.FailureCode != "" || result.OutputProjection["title"] != "Alpha" {
		t.Fatalf("unexpected gzip result: %#v", result)
	}
}

func TestWorkflowHTTPToolTransportMapsTimeoutWithoutRetry(t *testing.T) {
	plan, profile := workflowHTTPToolTransportFixture(t)
	attempts := 0
	transport := workflowHTTPToolTestTransport(func(*http.Request) (*http.Response, error) {
		attempts++
		return nil, context.DeadlineExceeded
	})
	result := transport.Execute(context.Background(), plan, profile, "req_workflow_tool_transport_007")
	if result.FailureCode != WorkflowRunFailureToolTimeout || result.FailureCategory != WorkflowHTTPToolFailureTimeout || attempts != 1 {
		t.Fatalf("expected one timeout attempt, got %#v attempts=%d", result, attempts)
	}
}

func TestNewPinnedWorkflowHTTPToolClientDisablesProxyReuseCompressionAndRedirects(t *testing.T) {
	client, err := newPinnedWorkflowHTTPToolClient("api.dev.example.invalid", 443, []netip.Addr{netip.MustParseAddr("8.8.8.8")}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy != nil || !transport.DisableKeepAlives || !transport.DisableCompression ||
		transport.MaxConnsPerHost != 1 || transport.TLSClientConfig == nil ||
		transport.TLSClientConfig.ServerName != "api.dev.example.invalid" {
		t.Fatalf("unexpected pinned transport: %#v", transport)
	}
	if err := client.CheckRedirect(&http.Request{}, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("expected redirects disabled, got %v", err)
	}
}

func workflowHTTPToolTransportFixture(t *testing.T) (WorkflowHTTPToolActionPlan, WorkflowHTTPToolExecutionProfile) {
	t.Helper()
	registry, err := newWorkflowHTTPToolRegistry()
	if err != nil {
		t.Fatal(err)
	}
	plan := WorkflowHTTPToolActionPlan{
		SchemaVersion: workflowHTTPToolPlanSchema, PlanID: "wtap_0123456789abcdef", RecordVersion: 2,
		TenantRef: "tenant:alpha", WorkspaceID: "workspace-alpha", ApplicationID: "application-alpha",
		DraftID: "draft-alpha", DraftVersion: 3, NodeID: "tool-node",
		ToolID: registry.definition.ToolID, ToolVersion: registry.definition.ToolVersion,
		DefinitionDigest: registry.definitionDigest, ProfileID: registry.profile.ProfileID,
		ProfileVersion: registry.profile.ProfileVersion, ProfileDigest: registry.profileDigest,
		Method: registry.profile.Method, TargetPolicyKey: registry.profile.TargetPolicyKey,
		PublicArguments: WorkflowHTTPToolPublicArguments{ResourceKey: "docs/alpha", Locale: "zh-CN"},
		OutputFields:    append([]string(nil), registry.definition.OutputFields...), OutputSchemaDigest: registry.outputSchemaDigest,
		CredentialPolicy: registry.profile.CredentialPolicy, TimeoutMS: registry.profile.TimeoutMS,
		MaxResponseBytes: registry.profile.MaxResponseBytes, MaxOutputBytes: registry.profile.MaxOutputBytes,
		PlannedByActorRef: "actor:planner", CreatedAt: "2026-07-16T09:00:00Z", ExpiresAt: "2026-07-16T09:15:00Z",
		Status: WorkflowHTTPToolActionStatusApproved, AuditRef: "audit:workflow-tool",
	}
	plan.ToolPlanDigest, err = workflowHTTPToolPlanDigest(plan)
	if err != nil {
		t.Fatal(err)
	}
	return plan, registry.profile
}

func workflowHTTPToolTestTransport(roundTrip workflowHTTPToolRoundTripperFunc) workflowHTTPToolTransport {
	transport := newWorkflowHTTPToolTransport()
	transport.lookupNetIP = func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("2001:4860:4860::8888")}, nil
	}
	transport.newClient = func(string, int, []netip.Addr, time.Duration) (*http.Client, error) {
		return &http.Client{Transport: roundTrip}, nil
	}
	return transport
}

func workflowHTTPToolJSONResponse(status int, payload string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(payload)),
	}
}
