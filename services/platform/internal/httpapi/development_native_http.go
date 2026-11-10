package httpapi

import (
	"context"
	"strings"

	"radishmind.local/services/platform/internal/developmentcontrol"
)

// ConnectDevelopmentNativeHTTP is an explicit trusted-host integration point.
// Server construction never invokes it. It supports only the one connection
// already owned by platform configuration; ambient Python profiles cannot
// supply a second endpoint or credentials. It starts no listener or worker.
func (s *Server) ConnectDevelopmentNativeHTTP(ctx context.Context, policy developmentcontrol.NativePolicy) (*developmentcontrol.Runtime, error) {
	backend, err := developmentcontrol.NewNativeHTTPBackend(ctx, func(ctx context.Context) (developmentcontrol.NativeHTTPConnection, error) {
		return s.developmentNativeConnection(ctx, policy)
	}, s.options.TestOnly)
	if err != nil {
		return nil, err
	}
	return s.ConnectDevelopmentControl(backend)
}

func (s *Server) developmentNativeConnection(ctx context.Context, policy developmentcontrol.NativePolicy) (developmentcontrol.NativeHTTPConnection, error) {
	cfg := s.config
	if s.bridge == nil || cfg.Provider != "openai-compatible" || cfg.BaseURL == "" || cfg.APIKey == "" || cfg.Model == "" {
		return developmentcontrol.NativeHTTPConnection{}, developmentcontrol.ErrIneligible
	}
	ref := buildNorthboundProfileModelID(cfg.Provider, cfg.ProviderProfile)
	inventory, err := s.bridge.DescribeInventory(ctx)
	if err != nil {
		return developmentcontrol.NativeHTTPConnection{}, developmentcontrol.ErrUnavailable
	}
	found := false
	for _, p := range inventory.Profiles {
		if buildNorthboundProfileModelID(p.ProviderID, p.Profile) != ref {
			continue
		}
		if found || !p.Enabled || p.ProviderID != cfg.Provider || p.APIStyle != "openai-compatible" || p.ResolvedModel != cfg.Model || p.CredentialState != "configured" {
			return developmentcontrol.NativeHTTPConnection{}, developmentcontrol.ErrIneligible
		}
		found = true
	}
	if !found {
		return developmentcontrol.NativeHTTPConnection{}, developmentcontrol.ErrIneligible
	}
	// Reuse the same explicit-config selection and credential path as Gateway.
	options := s.buildBridgeEnvelopeOptions(northboundSelection{provider: cfg.Provider, providerProfile: cfg.ProviderProfile, upstreamModel: cfg.Model}, 0)
	if options.APIKey == "" || options.BaseURL == "" || options.Model != strings.TrimSpace(cfg.Model) {
		return developmentcontrol.NativeHTTPConnection{}, developmentcontrol.ErrIneligible
	}
	return developmentcontrol.NativeHTTPConnection{ModelRef: ref, Model: options.Model, BaseURL: options.BaseURL, APIKey: options.APIKey, Policy: policy}, nil
}
