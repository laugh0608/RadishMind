package developmentcontrol

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const nativeHTTPBodyLimit = 262144

// NativeHTTPConnection is trusted host input, never an SDK/browser contract.
// The owner must resolve it from current configuration and credentials each time.
type NativeHTTPConnection struct {
	ModelRef string
	Model    string
	BaseURL  string
	APIKey   string
	Policy   NativePolicy
}

type nativeHTTPBackend struct {
	frozen   NativeHTTPConnection
	resolve  func(context.Context) (NativeHTTPConnection, error)
	endpoint string
	client   *http.Client
}

// Only these constant categories can leave the transport; underlying errors and
// provider bodies can contain credentials, URLs or sensitive prompt material.
type nativeHTTPError string

func (e nativeHTTPError) Error() string { return string(e) }
func (e nativeHTTPError) Unwrap() error { return ErrTransport }

func nativeHTTPFailure(err error) error {
	var failure nativeHTTPError
	if errors.As(err, &failure) {
		return failure
	}
	return ErrTransport
}

// NewNativeHTTPBackend does not dial or start a service. The loopback exception
// is exclusively for explicit fixed-upstream tests, never a provider fallback.
func NewNativeHTTPBackend(ctx context.Context, resolve func(context.Context) (NativeHTTPConnection, error), testOnlyLoopback bool) (NativeBackend, error) {
	if resolve == nil {
		return nil, ErrIneligible
	}
	c, err := resolve(ctx)
	if err != nil {
		return nil, ErrIneligible
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || u.RawPath != "" || strings.Contains(u.Path, "..") {
		return nil, ErrIneligible
	}
	loopback := testOnlyLoopback && u.Scheme == "http" && u.Hostname() == "127.0.0.1" && u.Port() != ""
	if !loopback && u.Scheme != "https" {
		return nil, ErrIneligible
	}
	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 || strconv.Itoa(port) != u.Port() {
			return nil, ErrIneligible
		}
	}
	if c.ModelRef == "" || c.Model == "" || len(c.Model) > 512 || c.APIKey == "" || len(c.APIKey) > 16384 || strings.ContainsAny(c.APIKey, "\r\n") || strings.TrimSpace(c.APIKey) != c.APIKey || c.Policy.InputTokenCeiling < 1 || c.Policy.InputTokenCeiling > 1000000000 || c.Policy.MaxOutputTokens < 1 || c.Policy.MaxOutputTokens > 262144 || c.Policy.Timeout <= 0 || c.Policy.Timeout > 2*time.Minute {
		return nil, ErrIneligible
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/chat/completions"
	transport := &http.Transport{
		Proxy: nil, DisableKeepAlives: true, DisableCompression: true,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: c.Policy.Timeout,
		MaxResponseHeaderBytes: 16384,
		DialContext:            nativeHTTPDialer(u.Hostname(), loopback),
	}
	return &nativeHTTPBackend{frozen: c, resolve: resolve, endpoint: u.String(), client: &http.Client{Transport: transport, Timeout: c.Policy.Timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func nativePublicAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, prefix := range []string{"0.0.0.0/8", "100.64.0.0/10", "192.0.0.0/24", "192.0.2.0/24", "198.18.0.0/15", "198.51.100.0/24", "203.0.113.0/24", "240.0.0.0/4", "2001:db8::/32", "64:ff9b::/96", "64:ff9b:1::/48", "2002::/16"} {
		if netip.MustParsePrefix(prefix).Contains(ip) {
			return false
		}
	}
	return true
}

func nativeHTTPDialer(host string, loopback bool) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		actual, port, err := net.SplitHostPort(address)
		if err != nil || actual != host {
			return nil, nativeHTTPError("native_endpoint_denied")
		}
		ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if err != nil || len(ips) == 0 {
			return nil, nativeHTTPError("native_connection_failed")
		}
		for _, ip := range ips {
			if !(loopback && ip.Unmap() == netip.MustParseAddr("127.0.0.1")) && !nativePublicAddress(ip) {
				return nil, nativeHTTPError("native_endpoint_denied")
			}
		}
		// Pin the validated address for this call; no second DNS lookup or retry.
		return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ips[0].String(), port))
	}
}

func (b *nativeHTTPBackend) current(ctx context.Context, ref string) error {
	c, err := b.resolve(ctx)
	if err != nil || c != b.frozen || ref != c.ModelRef {
		return nativeHTTPError("native_configuration_changed")
	}
	return nil
}

func (b *nativeHTTPBackend) Policy(ctx context.Context, ref string) (NativePolicy, error) {
	if err := b.current(ctx, ref); err != nil {
		return NativePolicy{}, err
	}
	return b.frozen.Policy, nil
}

func (b *nativeHTTPBackend) Complete(ctx context.Context, ref string, input NativeRequest) (NativeResult, error) {
	ctx, cancel := context.WithTimeout(ctx, b.frozen.Policy.Timeout)
	defer cancel()
	if err := b.current(ctx, ref); err != nil {
		return NativeResult{}, err
	}
	data, err := json.Marshal(input)
	if err != nil || input.Model != b.frozen.Model || input.MaxCompletionTokens > b.frozen.Policy.MaxOutputTokens {
		return NativeResult{}, nativeHTTPError("native_request_invalid")
	}
	canonical, err := ParseNativeRequest(data)
	if err != nil {
		return NativeResult{}, nativeHTTPError("native_request_invalid")
	}
	// Both false values must be present: some providers default parallel tools
	// to true when the field is omitted by the internal omitempty tags.
	data, err = json.Marshal(struct {
		NativeRequest
		Stream            bool `json:"stream"`
		ParallelToolCalls bool `json:"parallel_tool_calls"`
	}{NativeRequest: canonical})
	if err != nil || len(data) > nativeHTTPBodyLimit {
		return NativeResult{}, nativeHTTPError("native_request_invalid")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, b.endpoint, bytes.NewReader(data))
	if err != nil {
		return NativeResult{}, nativeHTTPError("native_request_invalid")
	}
	request.Header.Set("Authorization", "Bearer "+b.frozen.APIKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := b.client.Do(request)
	if err != nil {
		return NativeResult{}, nativeHTTPContextError(ctx, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		code := "native_http_rejected"
		switch {
		case response.StatusCode == 401 || response.StatusCode == 403:
			code = "native_authentication_failed"
		case response.StatusCode == 429:
			code = "native_rate_limited"
		case response.StatusCode >= 500:
			code = "native_upstream_failed"
		case response.StatusCode >= 300 && response.StatusCode < 400:
			code = "native_redirect_denied"
		}
		return NativeResult{}, nativeHTTPError(code)
	}
	media, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil || media != "application/json" || response.Header.Get("Content-Encoding") != "" {
		return NativeResult{}, nativeHTTPError("native_response_invalid")
	}
	if response.ContentLength > nativeHTTPBodyLimit {
		return NativeResult{}, nativeHTTPError("native_response_too_large")
	}
	data, err = io.ReadAll(io.LimitReader(response.Body, nativeHTTPBodyLimit+1))
	if err != nil {
		return NativeResult{}, nativeHTTPContextError(ctx, err)
	}
	if len(data) > nativeHTTPBodyLimit {
		return NativeResult{}, nativeHTTPError("native_response_too_large")
	}
	result, err := parseNativeHTTPResponse(data, b.frozen.Model)
	if err != nil {
		return NativeResult{}, nativeHTTPError("native_response_invalid")
	}
	return result, nil
}

func nativeHTTPContextError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.Canceled) {
		return nativeHTTPError("native_canceled")
	}
	var timeout net.Error
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.As(err, &timeout) && timeout.Timeout() {
		return nativeHTTPError("native_timeout")
	}
	var failure nativeHTTPError
	if errors.As(err, &failure) {
		return failure
	}
	return nativeHTTPError("native_connection_failed")
}
