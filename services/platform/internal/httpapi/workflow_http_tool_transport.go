package httpapi

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const workflowHTTPToolUserAgent = "RadishMind-Workflow-HTTP-Tool/1"

type WorkflowHTTPToolFailureCategory string

const (
	WorkflowHTTPToolFailureNone             WorkflowHTTPToolFailureCategory = "none"
	WorkflowHTTPToolFailurePolicy           WorkflowHTTPToolFailureCategory = "policy"
	WorkflowHTTPToolFailureConfirmation     WorkflowHTTPToolFailureCategory = "confirmation"
	WorkflowHTTPToolFailureTransport        WorkflowHTTPToolFailureCategory = "transport"
	WorkflowHTTPToolFailureTimeout          WorkflowHTTPToolFailureCategory = "timeout"
	WorkflowHTTPToolFailureResponseStatus   WorkflowHTTPToolFailureCategory = "response_status"
	WorkflowHTTPToolFailureResponseTooLarge WorkflowHTTPToolFailureCategory = "response_too_large"
	WorkflowHTTPToolFailureResponseInvalid  WorkflowHTTPToolFailureCategory = "response_invalid"
	WorkflowHTTPToolFailureStore            WorkflowHTTPToolFailureCategory = "store"
	WorkflowHTTPToolFailureOutcomeUnknown   WorkflowHTTPToolFailureCategory = "outcome_unknown"
)

func validWorkflowHTTPToolFailureCategory(value WorkflowHTTPToolFailureCategory) bool {
	switch value {
	case WorkflowHTTPToolFailureNone, WorkflowHTTPToolFailurePolicy,
		WorkflowHTTPToolFailureConfirmation, WorkflowHTTPToolFailureTransport,
		WorkflowHTTPToolFailureTimeout, WorkflowHTTPToolFailureResponseStatus,
		WorkflowHTTPToolFailureResponseTooLarge, WorkflowHTTPToolFailureResponseInvalid,
		WorkflowHTTPToolFailureStore, WorkflowHTTPToolFailureOutcomeUnknown:
		return true
	default:
		return false
	}
}

type workflowHTTPToolTransportResult struct {
	OutputProjection map[string]any
	FailureCode      WorkflowRunFailureCode
	FailureCategory  WorkflowHTTPToolFailureCategory
	FailureSummary   string
	HTTPStatusClass  string
	ResponseBytes    int
	DurationMS       int
	NetworkAttempted bool
}

type workflowHTTPToolHTTPClientFactory func(
	host string,
	port int,
	addresses []netip.Addr,
	timeout time.Duration,
) (*http.Client, error)

type workflowHTTPToolTransport struct {
	lookupNetIP func(context.Context, string, string) ([]netip.Addr, error)
	newClient   workflowHTTPToolHTTPClientFactory
	now         func() time.Time
}

func newWorkflowHTTPToolTransport() workflowHTTPToolTransport {
	return workflowHTTPToolTransport{
		lookupNetIP: net.DefaultResolver.LookupNetIP,
		newClient:   newPinnedWorkflowHTTPToolClient,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (transport workflowHTTPToolTransport) Execute(
	ctx context.Context,
	plan WorkflowHTTPToolActionPlan,
	profile WorkflowHTTPToolExecutionProfile,
	requestID string,
) workflowHTTPToolTransportResult {
	startedAt := transportNow(transport)
	networkAttempted := false
	fail := func(code WorkflowRunFailureCode, category WorkflowHTTPToolFailureCategory, summary string) workflowHTTPToolTransportResult {
		return workflowHTTPToolTransportResult{
			FailureCode: code, FailureCategory: category, FailureSummary: summary,
			DurationMS: workflowHTTPToolDurationMS(startedAt, transportNow(transport)), NetworkAttempted: networkAttempted,
		}
	}
	if ctx == nil || transport.lookupNetIP == nil || transport.newClient == nil {
		return fail(WorkflowRunFailureToolPolicy, WorkflowHTTPToolFailurePolicy, "Workflow HTTP tool transport policy is unavailable.")
	}
	if err := validateWorkflowHTTPToolExecutionBinding(plan, profile, requestID); err != nil {
		return fail(WorkflowRunFailureToolPolicy, WorkflowHTTPToolFailurePolicy, "Workflow HTTP tool execution profile is invalid or no longer matches the approved plan.")
	}
	target, err := workflowHTTPToolTargetURL(plan, profile)
	if err != nil {
		return fail(WorkflowRunFailureToolPolicy, WorkflowHTTPToolFailurePolicy, "Workflow HTTP tool target policy is invalid.")
	}
	addresses, err := transport.lookupNetIP(ctx, "ip", profile.TargetPolicy.Host)
	if err != nil || len(addresses) == 0 {
		if workflowHTTPToolContextTimedOut(ctx, err) {
			return fail(WorkflowRunFailureToolTimeout, WorkflowHTTPToolFailureTimeout, "Workflow HTTP tool target resolution exceeded its deadline.")
		}
		return fail(WorkflowRunFailureToolTransport, WorkflowHTTPToolFailureTransport, "Workflow HTTP tool target could not be resolved.")
	}
	validatedAddresses, err := validateWorkflowHTTPToolResolvedAddresses(addresses)
	if err != nil {
		return fail(WorkflowRunFailureToolPolicy, WorkflowHTTPToolFailurePolicy, "Workflow HTTP tool target resolution was rejected by the network policy.")
	}
	timeout := time.Duration(plan.TimeoutMS) * time.Millisecond
	client, err := transport.newClient(profile.TargetPolicy.Host, profile.TargetPolicy.Port, validatedAddresses, timeout)
	if err != nil || client == nil {
		return fail(WorkflowRunFailureToolTransport, WorkflowHTTPToolFailureTransport, "Workflow HTTP tool client could not be initialized.")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return fail(WorkflowRunFailureToolPolicy, WorkflowHTTPToolFailurePolicy, "Workflow HTTP tool request could not be assembled.")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", workflowHTTPToolUserAgent)
	request.Header.Set("X-RadishMind-Trace-ID", requestID)
	networkAttempted = true
	response, err := client.Do(request)
	if err != nil {
		if workflowHTTPToolContextTimedOut(ctx, err) {
			return fail(WorkflowRunFailureToolTimeout, WorkflowHTTPToolFailureTimeout, "Workflow HTTP tool request exceeded its deadline.")
		}
		return fail(WorkflowRunFailureToolTransport, WorkflowHTTPToolFailureTransport, "Workflow HTTP tool request failed before a valid response was available.")
	}
	if response == nil || response.Body == nil {
		return fail(WorkflowRunFailureToolTransport, WorkflowHTTPToolFailureTransport, "Workflow HTTP tool returned an empty transport response.")
	}
	defer response.Body.Close()
	statusClass := workflowHTTPToolHTTPStatusClass(response.StatusCode)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		result := fail(WorkflowRunFailureToolResponseStatus, WorkflowHTTPToolFailureResponseStatus, "Workflow HTTP tool response status was rejected.")
		result.HTTPStatusClass = statusClass
		return result
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		result := fail(WorkflowRunFailureToolResponseInvalid, WorkflowHTTPToolFailureResponseInvalid, "Workflow HTTP tool response content type is invalid.")
		result.HTTPStatusClass = statusClass
		return result
	}
	bodyReader, closeBody, err := workflowHTTPToolDecodedBody(response)
	if err != nil {
		result := fail(WorkflowRunFailureToolResponseInvalid, WorkflowHTTPToolFailureResponseInvalid, "Workflow HTTP tool response content encoding is invalid.")
		result.HTTPStatusClass = statusClass
		return result
	}
	if closeBody != nil {
		defer closeBody()
	}
	payload, tooLarge, err := readWorkflowHTTPToolBody(bodyReader, plan.MaxResponseBytes)
	if err != nil {
		result := fail(WorkflowRunFailureToolTransport, WorkflowHTTPToolFailureTransport, "Workflow HTTP tool response body could not be read.")
		result.HTTPStatusClass = statusClass
		return result
	}
	if tooLarge {
		result := fail(WorkflowRunFailureToolResponseTooLarge, WorkflowHTTPToolFailureResponseTooLarge, "Workflow HTTP tool response exceeded its byte budget.")
		result.HTTPStatusClass = statusClass
		result.ResponseBytes = len(payload)
		return result
	}
	projection, err := projectWorkflowHTTPToolResponse(payload, plan, profile)
	if err != nil {
		result := fail(WorkflowRunFailureToolResponseInvalid, WorkflowHTTPToolFailureResponseInvalid, "Workflow HTTP tool response did not satisfy the reviewed output schema.")
		result.HTTPStatusClass = statusClass
		result.ResponseBytes = len(payload)
		return result
	}
	result := workflowHTTPToolTransportResult{
		OutputProjection: projection,
		FailureCategory:  WorkflowHTTPToolFailureNone,
		HTTPStatusClass:  statusClass,
		ResponseBytes:    len(payload),
		DurationMS:       workflowHTTPToolDurationMS(startedAt, transportNow(transport)),
		NetworkAttempted: true,
	}
	return result
}

func validateWorkflowHTTPToolExecutionBinding(
	plan WorkflowHTTPToolActionPlan,
	profile WorkflowHTTPToolExecutionProfile,
	requestID string,
) error {
	profileDigest, profileDigestErr := canonicalSHA256(profile)
	planDigest, planDigestErr := workflowHTTPToolPlanDigest(plan)
	if strings.TrimSpace(requestID) == "" || !workflowHTTPToolReferencePattern.MatchString(requestID) ||
		plan.SchemaVersion != workflowHTTPToolPlanSchema || plan.Status != WorkflowHTTPToolActionStatusApproved ||
		plan.ToolID != workflowHTTPToolID || plan.ToolVersion != workflowHTTPToolVersion ||
		profile.SchemaVersion != workflowHTTPToolProfileSchema || !profile.Enabled || profile.TestOnlyLoopback ||
		profile.ToolID != plan.ToolID || profile.ProfileID != plan.ProfileID ||
		profile.ProfileVersion != plan.ProfileVersion || profile.DefinitionDigest != plan.DefinitionDigest ||
		profileDigestErr != nil || profileDigest != plan.ProfileDigest || planDigestErr != nil || planDigest != plan.ToolPlanDigest ||
		profile.TargetPolicyKey != plan.TargetPolicyKey || profile.Method != plan.Method ||
		profile.CredentialPolicy != plan.CredentialPolicy || profile.TimeoutMS != plan.TimeoutMS ||
		profile.MaxResponseBytes != plan.MaxResponseBytes || profile.MaxOutputBytes != plan.MaxOutputBytes ||
		profile.MaxAttempts != 1 || plan.Method != http.MethodGet || plan.CredentialPolicy != "none" ||
		plan.TimeoutMS <= 0 || plan.TimeoutMS > workflowHTTPToolTimeoutMS ||
		plan.MaxResponseBytes <= 0 || plan.MaxResponseBytes > workflowHTTPToolMaxResponseBytes ||
		plan.MaxOutputBytes <= 0 || plan.MaxOutputBytes > workflowHTTPToolMaxOutputBytes {
		return errors.New("workflow HTTP tool execution binding mismatch")
	}
	policy := profile.NetworkPolicy
	if policy.FollowRedirects || policy.UseAmbientProxy || policy.AutomaticRetry || policy.FallbackEnabled ||
		policy.CrossProfileConnectionReuse || !policy.RequireAllResolvedAddressesPublic ||
		!policy.PinValidatedAddress || !policy.PreserveTLSServerName {
		return errors.New("workflow HTTP tool network policy mismatch")
	}
	return nil
}

func workflowHTTPToolTargetURL(
	plan WorkflowHTTPToolActionPlan,
	profile WorkflowHTTPToolExecutionProfile,
) (*url.URL, error) {
	targetPolicy := profile.TargetPolicy
	host := strings.ToLower(strings.TrimSpace(targetPolicy.Host))
	path := strings.TrimSpace(targetPolicy.Path)
	if targetPolicy.Scheme != "https" || host == "" || net.ParseIP(host) != nil ||
		targetPolicy.Port != 443 || !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "?#") ||
		strings.Contains(host, ":") || strings.Contains(host, "/") || strings.Contains(host, "@") {
		return nil, errors.New("invalid workflow HTTP tool target")
	}
	values := url.Values{}
	seenArguments := map[string]bool{}
	seenQueryKeys := map[string]bool{}
	for _, mapping := range targetPolicy.QueryMappings {
		argumentKey := strings.TrimSpace(mapping.ArgumentKey)
		queryKey := strings.TrimSpace(mapping.QueryKey)
		if seenArguments[argumentKey] || seenQueryKeys[queryKey] || queryKey == "" {
			return nil, errors.New("invalid workflow HTTP tool query mapping")
		}
		seenArguments[argumentKey], seenQueryKeys[queryKey] = true, true
		switch argumentKey {
		case "resource_key":
			values.Set(queryKey, plan.PublicArguments.ResourceKey)
		case "locale":
			if plan.PublicArguments.Locale != "" {
				values.Set(queryKey, plan.PublicArguments.Locale)
			}
		default:
			return nil, errors.New("invalid workflow HTTP tool query mapping")
		}
	}
	if !seenArguments["resource_key"] || !seenArguments["locale"] {
		return nil, errors.New("incomplete workflow HTTP tool query mapping")
	}
	return &url.URL{
		Scheme:   "https",
		Host:     net.JoinHostPort(host, strconv.Itoa(targetPolicy.Port)),
		Path:     path,
		RawQuery: values.Encode(),
	}, nil
}

func validateWorkflowHTTPToolResolvedAddresses(addresses []netip.Addr) ([]netip.Addr, error) {
	unique := make(map[netip.Addr]struct{}, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		if !address.IsValid() || !workflowHTTPToolAddressIsPublic(address) {
			return nil, errors.New("workflow HTTP tool resolved address is not public")
		}
		unique[address] = struct{}{}
	}
	if len(unique) == 0 {
		return nil, errors.New("workflow HTTP tool resolved no public addresses")
	}
	validated := make([]netip.Addr, 0, len(unique))
	for address := range unique {
		validated = append(validated, address)
	}
	sort.Slice(validated, func(left, right int) bool { return validated[left].Compare(validated[right]) < 0 })
	return validated, nil
}

var workflowHTTPToolForbiddenAddressPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("169.254.0.0/16"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

func workflowHTTPToolAddressIsPublic(address netip.Addr) bool {
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() ||
		address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() || address.IsMulticast() ||
		address.IsUnspecified() {
		return false
	}
	for _, prefix := range workflowHTTPToolForbiddenAddressPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func newPinnedWorkflowHTTPToolClient(
	host string,
	port int,
	addresses []netip.Addr,
	timeout time.Duration,
) (*http.Client, error) {
	if strings.TrimSpace(host) == "" || port != 443 || len(addresses) == 0 || timeout <= 0 {
		return nil, errors.New("invalid workflow HTTP tool client policy")
	}
	address := addresses[0]
	if !workflowHTTPToolAddressIsPublic(address) {
		return nil, errors.New("invalid workflow HTTP tool dial address")
	}
	dialer := &net.Dialer{Timeout: timeout, KeepAlive: -1}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, net.JoinHostPort(address.String(), strconv.Itoa(port)))
		},
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     true,
		DisableCompression:    true,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   0,
		MaxConnsPerHost:       1,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: 0,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: strings.TrimSpace(host),
		},
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func workflowHTTPToolDecodedBody(response *http.Response) (io.Reader, func() error, error) {
	encoding := strings.ToLower(strings.TrimSpace(response.Header.Get("Content-Encoding")))
	switch encoding {
	case "", "identity":
		return response.Body, nil, nil
	case "gzip":
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, nil, err
		}
		return reader, reader.Close, nil
	default:
		return nil, nil, errors.New("unsupported workflow HTTP tool content encoding")
	}
}

func readWorkflowHTTPToolBody(reader io.Reader, limit int) ([]byte, bool, error) {
	if reader == nil || limit <= 0 || limit > workflowHTTPToolMaxResponseBytes {
		return nil, false, errors.New("invalid workflow HTTP tool response budget")
	}
	payload, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, false, err
	}
	if len(payload) > limit {
		return payload, true, nil
	}
	return payload, false, nil
}

type workflowHTTPToolReviewedJSONResponse struct {
	ResourceKey string `json:"resource_key"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	UpdatedAt   string `json:"updated_at"`
}

func projectWorkflowHTTPToolResponse(
	payload []byte,
	plan WorkflowHTTPToolActionPlan,
	profile WorkflowHTTPToolExecutionProfile,
) (map[string]any, error) {
	var decoded workflowHTTPToolReviewedJSONResponse
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("workflow HTTP tool response has trailing data")
	}
	if decoded.ResourceKey != plan.PublicArguments.ResourceKey ||
		len(decoded.ResourceKey) == 0 || len(decoded.ResourceKey) > workflowHTTPToolMaxResourceKeySize ||
		len([]rune(decoded.Title)) == 0 || len([]rune(decoded.Title)) > 240 ||
		len([]rune(decoded.Summary)) == 0 || len([]rune(decoded.Summary)) > 4096 ||
		strings.Contains(decoded.ResourceKey, "://") || strings.Contains(decoded.Title, "://") ||
		strings.Contains(decoded.Summary, "://") {
		return nil, errors.New("workflow HTTP tool response projection is invalid")
	}
	if _, err := time.Parse(time.RFC3339Nano, decoded.UpdatedAt); err != nil {
		return nil, errors.New("workflow HTTP tool response timestamp is invalid")
	}
	projection := map[string]any{
		"resource_key": decoded.ResourceKey,
		"title":        decoded.Title,
		"summary":      decoded.Summary,
		"updated_at":   decoded.UpdatedAt,
	}
	encoded, err := json.Marshal(projection)
	if err != nil || len(encoded) > plan.MaxOutputBytes || len(encoded) > profile.MaxOutputBytes {
		return nil, errors.New("workflow HTTP tool response projection exceeds budget")
	}
	return projection, nil
}

func workflowHTTPToolContextTimedOut(ctx context.Context, err error) bool {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var networkError net.Error
	return errors.As(err, &networkError) && networkError.Timeout()
}

func workflowHTTPToolHTTPStatusClass(status int) string {
	if status < 100 || status > 599 {
		return ""
	}
	return fmt.Sprintf("%dxx", status/100)
}

func transportNow(transport workflowHTTPToolTransport) time.Time {
	if transport.now == nil {
		return time.Now().UTC()
	}
	return transport.now().UTC()
}

func workflowHTTPToolDurationMS(startedAt, completedAt time.Time) int {
	duration := completedAt.Sub(startedAt)
	if duration < 0 {
		return 0
	}
	if duration > workflowExecutorDefaultMaxRuntime {
		return int(workflowExecutorDefaultMaxRuntime.Milliseconds())
	}
	return int(duration.Milliseconds())
}
