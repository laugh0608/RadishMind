package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"radishmind.local/services/platform/internal/bridge"
	"radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/workspacepolicy"
)

// Current owners remain authoritative; a task bearer cannot supply these facts.
type developmentQualificationReader struct{ server *Server }

func (reader developmentQualificationReader) Resolve(ctx context.Context, identity developmentcontrol.Identity, plan developmentcontrol.Plan) (developmentcontrol.Selection, error) {
	s := reader.server
	if s.localIdentityRepository == nil || s.gatewayModelPricingRepository == nil || s.bridge == nil || plan.Validate() != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	now := time.Now().UTC()
	auth, err := s.localIdentityRepository.AuthorizeWorkspace(ctx, identity.AccountID, identity.TenantRef, identity.WorkspaceID, []string{"applications:read", workspacepolicy.PermissionRolesAssign}, now)
	if err != nil {
		if errors.Is(err, errLocalIdentityStoreUnavailable) || errors.Is(err, errLocalIdentityContractMismatch) {
			return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
		}
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	canManage := false
	for _, role := range auth.RoleAssignments {
		canManage = canManage || localIdentityAssignmentCanManage(role, now)
	}
	if !canManage {
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	// Only authorization facts enter the fingerprint, never login identifiers,
	// display names, credentials, sessions or provider secrets.
	roles := slices.Clone(auth.RoleAssignments)
	slices.SortFunc(roles, func(a, b LocalRoleAssignment) int {
		if a.AssignmentID < b.AssignmentID {
			return -1
		}
		if a.AssignmentID > b.AssignmentID {
			return 1
		}
		return 0
	})
	for i := range roles {
		roles[i].PermissionGrants = slices.Sorted(slices.Values(roles[i].PermissionGrants))
	}
	identityBytes, err := json.Marshal(struct {
		UserID         string
		AccountVersion int
		AccountState   string
		Membership     WorkspaceMembership
		Roles          []LocalRoleAssignment
	}{auth.Account.UserID, auth.Account.RecordVersion, auth.Account.LifecycleState, auth.Membership, roles})
	if err != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	actor := "user:" + identity.AccountID
	authority, failure := s.applicationInteractionAuthorityResolver().Resolve(ApplicationInteractionContext{RequestContext: ctx, RequestID: "development-qualification", TenantRef: identity.TenantRef, WorkspaceID: identity.WorkspaceID, ApplicationID: plan.ApplicationID, ActorRef: actor, OwnerSubjectRef: actor, AuditRef: "audit:development-qualification", WriteEnabled: true}, ApplicationInteractionProfileBinding{ExecutionProfile: plan.ExecutionProfile, DefinitionID: plan.DefinitionID})
	if failure != "" {
		if failure == ApplicationInteractionFailureStoreUnavailable {
			return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
		}
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	inventory, err := s.bridge.DescribeInventory(ctx)
	if err != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	var profile bridge.ProviderProfileDescription
	found := false
	for _, candidate := range inventory.Profiles {
		if buildNorthboundProfileModelID(candidate.ProviderID, candidate.Profile) == plan.ModelRef {
			if found {
				return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
			}
			profile, found = candidate, true
		}
	}
	if !found || !profile.Enabled || profile.ResolvedModel == "" || profile.APIStyle != "openai-compatible" || profile.CredentialState != "configured" {
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	profileBytes, err := json.Marshal(profile)
	if err != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	pricingContext := GatewayModelPricingContext{RequestContext: ctx, TenantRef: identity.TenantRef, WorkspaceID: identity.WorkspaceID, Environment: "development", ProviderID: profile.ProviderID, ProfileID: profile.Profile, ModelID: profile.ResolvedModel, ActorRef: actor, RequestID: "development-qualification", AuditRef: "audit:development-qualification"}
	pricing, exists, err := s.gatewayModelPricingRepository.ReadCurrent(pricingContext)
	if err != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	if !exists {
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	if !validGatewayModelPricingPolicy(pricingContext, pricing) {
		return developmentcontrol.Selection{}, developmentcontrol.ErrUnavailable
	}
	selected := developmentcontrol.Selection{ApplicationAuthorityDigest: authority.AuthorityDigest, IdentityDigest: developmentcontrol.PayloadDigest(identityBytes), ModelDigest: developmentcontrol.PayloadDigest(profileBytes), WireModel: profile.ResolvedModel, PricingDigest: pricing.PolicyDigest, TokenUnit: pricing.TokenUnit, InputPriceMicros: pricing.InputPriceMicrosPerTokenUnit, OutputPriceMicros: pricing.OutputPriceMicrosPerTokenUnit}
	if selected.Validate() != nil {
		return developmentcontrol.Selection{}, developmentcontrol.ErrIneligible
	}
	return selected, nil
}
