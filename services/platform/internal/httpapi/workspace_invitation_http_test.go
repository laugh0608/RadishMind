package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestWorkspaceInvitationHTTPFiveRouteVerticalChain(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)
	builder := roleDefinitionByKey(t, LocalIdentityBuiltInRoleCatalog(), localIdentityRoleWorkspaceBuilder)
	creation := createWorkspaceInvitationOverHTTP(t, fixture, builder, workspaceInvitationTTL24Hours)
	if creation.response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("create response is cacheable: %q", creation.response.Header().Get("Cache-Control"))
	}
	if creation.document.InvitationCode == "" ||
		creation.document.Invitation.RoleKey != localIdentityRoleWorkspaceBuilder ||
		creation.document.RequestID == "" {
		t.Fatalf("invitation creation projection mismatch: %#v", creation.document)
	}
	assertWorkspaceInvitationHTTPSafePayload(t, creation.response.Body.String(), true, creation.document.InvitationCode)

	listPath := strings.Replace(workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1) +
		"?effective_state=pending&limit=1"
	listResponse := fixture.request(t, http.MethodGet, listPath, nil, fixture.adminCookies)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list workspace invitations: status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	var listed workspaceInvitationListHTTPResponse
	decodeLocalIdentityHTTPResponse(t, listResponse, &listed)
	if listed.TenantRef != "tenant_demo" || listed.WorkspaceID != "workspace_demo" ||
		len(listed.Invitations) != 1 || listed.Invitations[0].InvitationID != creation.document.Invitation.InvitationID ||
		listed.Invitations[0].EffectiveState != workspaceInvitationEffectivePending {
		t.Fatalf("invitation list projection mismatch: %#v", listed)
	}
	assertWorkspaceInvitationHTTPSafePayload(t, listResponse.Body.String(), false, creation.document.InvitationCode)

	previewResponse := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute,
		map[string]any{"invitation_code": creation.document.InvitationCode}, fixture.targetCookies, "tenant_demo",
	)
	if previewResponse.Code != http.StatusOK || previewResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("preview invitation: status=%d cache=%q body=%s", previewResponse.Code,
			previewResponse.Header().Get("Cache-Control"), previewResponse.Body.String())
	}
	var preview workspaceInvitationPreviewHTTPResponse
	decodeLocalIdentityHTTPResponse(t, previewResponse, &preview)
	if preview.InvitationID != creation.document.Invitation.InvitationID || preview.TenantRef != "tenant_demo" ||
		preview.WorkspaceID != "workspace_demo" || preview.Role.RoleKey != localIdentityRoleWorkspaceBuilder {
		t.Fatalf("invitation preview mismatch: %#v", preview)
	}
	assertWorkspaceInvitationHTTPSafePayload(t, previewResponse.Body.String(), false, creation.document.InvitationCode)

	claimResponse := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, map[string]any{
			"invitation_code": creation.document.InvitationCode, "expected_record_version": preview.RecordVersion,
			"confirmed": true,
		}, fixture.targetCookies, "tenant_demo",
	)
	if claimResponse.Code != http.StatusOK || claimResponse.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("claim invitation: status=%d cache=%q body=%s", claimResponse.Code,
			claimResponse.Header().Get("Cache-Control"), claimResponse.Body.String())
	}
	var claimed workspaceInvitationMutationHTTPResponse
	decodeLocalIdentityHTTPResponse(t, claimResponse, &claimed)
	if claimed.Invitation.EffectiveState != workspaceInvitationEffectiveClaimed || claimed.Membership == nil ||
		claimed.RoleAssignment == nil || claimed.Membership.UserID != fixture.target.Account.UserID ||
		claimed.RoleAssignment.RoleKey != localIdentityRoleWorkspaceBuilder {
		t.Fatalf("claim response did not return committed authorization refs: %#v", claimed)
	}
	assertWorkspaceInvitationHTTPSafePayload(t, claimResponse.Body.String(), false, creation.document.InvitationCode)
	if _, err := fixture.identity.repository.AuthorizeWorkspace(
		context.Background(), fixture.target.Account.UserID, "tenant_demo", "workspace_demo",
		[]string{"applications:write", "workflow_runs:execute"}, fixture.identity.service.nowUTC(),
	); err != nil {
		t.Fatalf("claimed permissions were not immediately effective: %v", err)
	}

	replay := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, map[string]any{
			"invitation_code":         creation.document.InvitationCode,
			"expected_record_version": preview.RecordVersion, "confirmed": true,
		}, fixture.targetCookies, "tenant_demo",
	)
	assertWorkspaceInvitationHTTPError(t, replay, http.StatusConflict, WorkspaceInvitationFailureNotClaimable,
		"discard_terminal_invitation")
	assertWorkspaceInvitationHTTPSafePayload(t, replay.Body.String(), false, creation.document.InvitationCode)

	claimedList := fixture.request(t, http.MethodGet, strings.Replace(
		workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1,
	)+"?effective_state=claimed", nil, fixture.adminCookies)
	if claimedList.Code != http.StatusOK {
		t.Fatalf("list claimed invitation: status=%d body=%s", claimedList.Code, claimedList.Body.String())
	}
	var claimedPage workspaceInvitationListHTTPResponse
	decodeLocalIdentityHTTPResponse(t, claimedList, &claimedPage)
	if len(claimedPage.Invitations) != 1 || claimedPage.Invitations[0].MembershipID != claimed.Membership.MembershipID ||
		claimedPage.Invitations[0].AssignmentID != claimed.RoleAssignment.AssignmentID {
		t.Fatalf("claimed directory did not use canonical terminal refs: %#v", claimedPage)
	}

	reviewer := roleDefinitionByKey(t, LocalIdentityBuiltInRoleCatalog(), localIdentityRoleWorkspaceReviewer)
	revocable := createWorkspaceInvitationOverHTTP(t, fixture, reviewer, workspaceInvitationTTL1Hour)
	revokePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{invitation_id}", revocable.document.Invitation.InvitationID,
	).Replace(workspaceInvitationAdminRevokeRoute)
	revokeResponse := fixture.request(t, http.MethodPost, revokePath, map[string]any{
		"expected_record_version": revocable.document.Invitation.RecordVersion, "confirmed": true,
	}, fixture.adminCookies)
	if revokeResponse.Code != http.StatusOK {
		t.Fatalf("revoke invitation: status=%d body=%s", revokeResponse.Code, revokeResponse.Body.String())
	}
	revokedPreview := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute,
		map[string]any{"invitation_code": revocable.document.InvitationCode}, fixture.targetCookies, "tenant_demo",
	)
	assertWorkspaceInvitationHTTPError(t, revokedPreview, http.StatusConflict, WorkspaceInvitationFailureNotClaimable,
		"discard_terminal_invitation")
}

func TestWorkspaceInvitationHTTPAuthenticationScopeCSRFAndEnumerationGuards(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)
	reader := roleDefinitionByKey(t, LocalIdentityBuiltInRoleCatalog(), localIdentityRoleWorkspaceReader)
	creation := createWorkspaceInvitationOverHTTP(t, fixture, reader, workspaceInvitationTTL24Hours)
	previewBody := map[string]any{"invitation_code": creation.document.InvitationCode}

	unauthenticated := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody, nil, "tenant_demo",
	)
	assertLocalIdentityError(t, unauthenticated, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	for name, header := range map[string][2]string{
		"bearer":        {"Authorization", "Bearer forbidden"},
		"dev-header":    {controlPlaneReadDevIdentityHeader, "verified:dev"},
		"signed-test":   {"Authorization", "Bearer signed-test-token"},
		"resource-oidc": {"Authorization", "Bearer resource-server-oidc-token"},
	} {
		t.Run(name, func(t *testing.T) {
			request := workspaceInvitationClaimantRequest(
				t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
				fixture.targetCookies, "tenant_demo",
			)
			request.Header.Set(header[0], header[1])
			response := httptest.NewRecorder()
			fixture.identity.handler.ServeHTTP(response, request)
			assertLocalIdentityError(t, response, http.StatusUnauthorized, localIdentityAuthenticationRequired)
		})
	}
	adminFallbackRequest := fixture.adminRequest(t, http.MethodGet, strings.Replace(
		workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1,
	), nil, fixture.adminCookies)
	adminFallbackRequest.Header.Set("Authorization", "Bearer resource-server-oidc-token")
	adminFallback := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(adminFallback, adminFallbackRequest)
	assertLocalIdentityError(t, adminFallback, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	for name, tenantRef := range map[string]string{"missing-tenant": "", "wrong-tenant": "tenant_other"} {
		t.Run(name, func(t *testing.T) {
			response := workspaceInvitationClaimantResponse(
				t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
				fixture.targetCookies, tenantRef,
			)
			assertWorkspaceInvitationHTTPError(t, response, http.StatusForbidden,
				WorkspaceInvitationFailureAccountIneligible, "select_matching_tenant_or_account")
		})
	}

	claimBody := map[string]any{
		"invitation_code":         creation.document.InvitationCode,
		"expected_record_version": creation.document.Invitation.RecordVersion, "confirmed": true,
	}
	missingCSRFRequest := workspaceInvitationClaimantRequest(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, claimBody, fixture.targetCookies, "tenant_demo",
	)
	missingCSRFRequest.Header.Del(localIdentityCSRFHeader)
	missingCSRF := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(missingCSRF, missingCSRFRequest)
	assertLocalIdentityError(t, missingCSRF, http.StatusForbidden, localIdentityCSRFInvalid)

	wrongOriginRequest := workspaceInvitationClaimantRequest(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, claimBody, fixture.targetCookies, "tenant_demo",
	)
	wrongOriginRequest.Header.Set("Origin", "https://evil.example")
	wrongOrigin := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(wrongOrigin, wrongOriginRequest)
	assertLocalIdentityError(t, wrongOrigin, http.StatusForbidden, localIdentityOriginForbidden)
	stillPending := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
		fixture.targetCookies, "tenant_demo",
	)
	if stillPending.Code != http.StatusOK {
		t.Fatalf("rejected CSRF/origin request consumed invitation: status=%d body=%s", stillPending.Code, stillPending.Body.String())
	}
	staleClaim := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, map[string]any{
			"invitation_code":         creation.document.InvitationCode,
			"expected_record_version": creation.document.Invitation.RecordVersion + 1, "confirmed": true,
		}, fixture.targetCookies, "tenant_demo",
	)
	assertWorkspaceInvitationHTTPError(t, staleClaim, http.StatusConflict,
		WorkspaceInvitationFailureVersionConflict, "refresh_invitation_state")
	stillPendingAfterCAS := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
		fixture.targetCookies, "tenant_demo",
	)
	if stillPendingAfterCAS.Code != http.StatusOK {
		t.Fatalf("stale claim changed invitation: status=%d body=%s",
			stillPendingAfterCAS.Code, stillPendingAfterCAS.Body.String())
	}
	redactedTraceRequest := workspaceInvitationClaimantRequest(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
		fixture.targetCookies, "tenant_demo",
	)
	redactedTraceRequest.Header.Set("X-Request-ID", creation.document.InvitationCode)
	redactedTrace := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(redactedTrace, redactedTraceRequest)
	if redactedTrace.Code != http.StatusOK ||
		redactedTrace.Header().Get("X-Request-Id") == creation.document.InvitationCode ||
		strings.Contains(redactedTrace.Body.String(), creation.document.InvitationCode) {
		t.Fatalf("invitation code entered trace metadata: status=%d request_id=%q body=%s",
			redactedTrace.Code, redactedTrace.Header().Get("X-Request-Id"), redactedTrace.Body.String())
	}

	parts := strings.Split(creation.document.InvitationCode, ".")
	unknownLocator := "rmi_ffffffffffffffffffffffffffffffff." + parts[1]
	wrongSecret := parts[0] + "." + strings.Repeat("A", 43)
	invalidResponses := []*httptest.ResponseRecorder{
		workspaceInvitationClaimantResponse(t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute,
			map[string]any{"invitation_code": "not-an-invitation"}, fixture.targetCookies, "tenant_demo"),
		workspaceInvitationClaimantResponse(t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute,
			map[string]any{"invitation_code": unknownLocator}, fixture.targetCookies, "tenant_demo"),
		workspaceInvitationClaimantResponse(t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute,
			map[string]any{"invitation_code": wrongSecret}, fixture.targetCookies, "tenant_demo"),
	}
	for _, response := range invalidResponses {
		assertWorkspaceInvitationHTTPError(t, response, http.StatusBadRequest,
			WorkspaceInvitationFailureInvalid, "reenter_invitation_code")
		assertWorkspaceInvitationHTTPSafePayload(t, response.Body.String(), false, creation.document.InvitationCode)
		for _, forbidden := range []string{"tenant_demo", "workspace_demo", localIdentityRoleWorkspaceReader, parts[0], parts[1]} {
			if strings.Contains(response.Body.String(), forbidden) {
				t.Fatalf("invalid-code response leaked %q: %s", forbidden, response.Body.String())
			}
		}
	}

	fixture.identity.setNow(fixture.identity.service.nowUTC().Add(workspaceInvitationRecentAuthAge + time.Second))
	stalePreview := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, previewBody,
		fixture.targetCookies, "tenant_demo",
	)
	assertWorkspaceInvitationHTTPError(t, stalePreview, http.StatusUnauthorized,
		LocalIdentityFailureRecentAuthentication, "reauthenticate")
	staleCreatePath := strings.Replace(workspaceInvitationAdminCreateRoute, "{workspace_id}", "workspace_demo", 1)
	staleCreate := fixture.request(t, http.MethodPost, staleCreatePath, map[string]any{
		"role_key": reader.RoleKey, "expected_catalog_version": reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"ttl_policy":                      workspaceInvitationTTL24Hours, "confirmed": true,
	}, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, staleCreate, http.StatusUnauthorized,
		LocalIdentityFailureRecentAuthentication, "reauthenticate")
}

func TestWorkspaceInvitationHTTPStrictPayloadQueryPermissionAndMethodBoundaries(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)
	reader := roleDefinitionByKey(t, LocalIdentityBuiltInRoleCatalog(), localIdentityRoleWorkspaceReader)
	creation := createWorkspaceInvitationOverHTTP(t, fixture, reader, workspaceInvitationTTL24Hours)

	injection := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute, map[string]any{
			"invitation_code": creation.document.InvitationCode, "tenant_ref": "tenant_demo",
		}, fixture.targetCookies, "tenant_demo",
	)
	assertLocalIdentityError(t, injection, http.StatusBadRequest, "INVALID_JSON")

	duplicatePayload := `{"invitation_code":"` + creation.document.InvitationCode + `","invitation_code":"` +
		creation.document.InvitationCode + `"}`
	duplicate := workspaceInvitationRawClaimantResponse(
		t, fixture, workspaceInvitationClaimPreviewRoute, duplicatePayload, fixture.targetCookies, "tenant_demo",
	)
	assertLocalIdentityError(t, duplicate, http.StatusBadRequest, "INVALID_JSON")
	multiple := workspaceInvitationRawClaimantResponse(
		t, fixture, workspaceInvitationClaimPreviewRoute,
		`{"invitation_code":"`+creation.document.InvitationCode+`"} {}`, fixture.targetCookies, "tenant_demo",
	)
	assertLocalIdentityError(t, multiple, http.StatusBadRequest, "INVALID_JSON")
	query := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimPreviewRoute+"?tenant_ref=tenant_demo",
		map[string]any{"invitation_code": creation.document.InvitationCode}, fixture.targetCookies, "tenant_demo",
	)
	assertLocalIdentityError(t, query, http.StatusBadRequest, localIdentityPayloadInvalid)

	unconfirmedClaim := workspaceInvitationClaimantResponse(
		t, fixture, http.MethodPost, workspaceInvitationClaimRoute, map[string]any{
			"invitation_code":         creation.document.InvitationCode,
			"expected_record_version": creation.document.Invitation.RecordVersion, "confirmed": false,
		}, fixture.targetCookies, "tenant_demo",
	)
	assertLocalIdentityError(t, unconfirmedClaim, http.StatusBadRequest, localIdentityPayloadInvalid)

	invalidList := fixture.request(t, http.MethodGet, strings.Replace(
		workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1,
	)+"?limit=0", nil, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, invalidList, http.StatusBadRequest,
		WorkspaceInvitationFailureCursorInvalid, "restart_invitation_list")
	unknownFilter := fixture.request(t, http.MethodGet, strings.Replace(
		workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1,
	)+"?email=someone%40example.com", nil, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, unknownFilter, http.StatusBadRequest,
		WorkspaceInvitationFailureCursorInvalid, "restart_invitation_list")
	revokePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{invitation_id}", creation.document.Invitation.InvitationID,
	).Replace(workspaceInvitationAdminRevokeRoute)
	staleRevoke := fixture.request(t, http.MethodPost, revokePath, map[string]any{
		"expected_record_version": creation.document.Invitation.RecordVersion + 1, "confirmed": true,
	}, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, staleRevoke, http.StatusConflict,
		WorkspaceInvitationFailureVersionConflict, "refresh_invitation_state")

	createPath := strings.Replace(workspaceInvitationAdminCreateRoute, "{workspace_id}", "workspace_demo", 1)
	adminInvitation := fixture.request(t, http.MethodPost, createPath, map[string]any{
		"role_key": localIdentityRoleWorkspaceAdmin, "expected_catalog_version": reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"ttl_policy":                      workspaceInvitationTTL24Hours, "confirmed": true,
	}, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, adminInvitation, http.StatusBadRequest,
		WorkspaceInvitationFailureRoleIneligible, "refresh_role_catalog")
	invalidTTL := fixture.request(t, http.MethodPost, createPath, map[string]any{
		"role_key": reader.RoleKey, "expected_catalog_version": reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"ttl_policy":                      "30d", "confirmed": true,
	}, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, invalidTTL, http.StatusConflict,
		WorkspaceInvitationFailureTransitionInvalid, "refresh_invitation_directory")

	methodPath := createPath
	methodMismatch := fixture.request(t, http.MethodPut, methodPath, map[string]any{}, fixture.adminCookies)
	if methodMismatch.Code != http.StatusMethodNotAllowed {
		t.Fatalf("strict invitation route accepted wrong method: status=%d body=%s",
			methodMismatch.Code, methodMismatch.Body.String())
	}
	crossScope := fixture.request(t, http.MethodPost, strings.Replace(
		workspaceInvitationAdminCreateRoute, "{workspace_id}", "workspace_other", 1,
	), map[string]any{
		"role_key": reader.RoleKey, "expected_catalog_version": reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"ttl_policy":                      workspaceInvitationTTL24Hours, "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, crossScope, http.StatusForbidden, LocalIdentityFailureAdminScopeMismatch)

	underlying := fixture.identity.repository.(*memoryLocalIdentityRepository)
	permissionRepository := &workspaceInvitationHTTPPermissionRepository{
		localIdentityAdministrationRepository: underlying, underlying: underlying,
		deniedPermission: localIdentityPermissionRolesRead,
	}
	fixture.identity.server.localIdentityAdministrationService.repository = permissionRepository
	permissionDeniedList := fixture.request(t, http.MethodGet, strings.Replace(
		workspaceInvitationAdminListRoute, "{workspace_id}", "workspace_demo", 1,
	), nil, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, permissionDeniedList, http.StatusForbidden,
		LocalIdentityFailurePermissionDenied, "refresh_active_workspace_authorization")
	if !slices.Equal(permissionRepository.lastRequired, []string{
		localIdentityPermissionMembersRead, localIdentityPermissionRolesRead,
	}) {
		t.Fatalf("invitation list did not preauthorize the exact permission combination: %#v",
			permissionRepository.lastRequired)
	}
	permissionRepository.deniedPermission = localIdentityPermissionRolesAssign
	permissionDeniedCreate := fixture.request(t, http.MethodPost, methodPath, map[string]any{
		"role_key": reader.RoleKey, "expected_catalog_version": reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"ttl_policy":                      workspaceInvitationTTL24Hours, "confirmed": true,
	}, fixture.adminCookies)
	assertWorkspaceInvitationHTTPError(t, permissionDeniedCreate, http.StatusForbidden,
		LocalIdentityFailurePermissionDenied, "refresh_active_workspace_authorization")
	if !slices.Equal(permissionRepository.lastRequired, []string{
		localIdentityPermissionMembershipsWrite, localIdentityPermissionRolesAssign,
	}) {
		t.Fatalf("invitation create did not preauthorize the exact permission combination: %#v",
			permissionRepository.lastRequired)
	}
}

type workspaceInvitationHTTPPermissionRepository struct {
	localIdentityAdministrationRepository
	underlying       *memoryLocalIdentityRepository
	deniedPermission string
	lastRequired     []string
}

func (repository *workspaceInvitationHTTPPermissionRepository) AuthorizeWorkspace(
	ctx context.Context,
	userID string,
	tenantRef string,
	workspaceID string,
	required []string,
	now time.Time,
) (LocalWorkspaceAuthorization, error) {
	repository.lastRequired = append([]string(nil), required...)
	if slices.Contains(required, repository.deniedPermission) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityPermissionDenied
	}
	return repository.underlying.AuthorizeWorkspace(ctx, userID, tenantRef, workspaceID, required, now)
}

type workspaceInvitationHTTPCreateResult struct {
	response *httptest.ResponseRecorder
	document workspaceInvitationCreationHTTPResponse
}

func createWorkspaceInvitationOverHTTP(
	t *testing.T,
	fixture localIdentityAdministrationHTTPFixture,
	role LocalIdentityRoleDefinition,
	ttlPolicy string,
) workspaceInvitationHTTPCreateResult {
	t.Helper()
	path := strings.Replace(workspaceInvitationAdminCreateRoute, "{workspace_id}", "workspace_demo", 1)
	response := fixture.request(t, http.MethodPost, path, map[string]any{
		"role_key": role.RoleKey, "expected_catalog_version": role.CatalogVersion,
		"expected_role_definition_digest": role.DefinitionDigest,
		"ttl_policy":                      ttlPolicy, "confirmed": true,
	}, fixture.adminCookies)
	if response.Code != http.StatusCreated {
		t.Fatalf("create workspace invitation: status=%d body=%s", response.Code, response.Body.String())
	}
	var document workspaceInvitationCreationHTTPResponse
	decodeLocalIdentityHTTPResponse(t, response, &document)
	return workspaceInvitationHTTPCreateResult{response: response, document: document}
}

func workspaceInvitationClaimantResponse(
	t *testing.T,
	fixture localIdentityAdministrationHTTPFixture,
	method string,
	path string,
	body any,
	cookies []*http.Cookie,
	tenantRef string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := workspaceInvitationClaimantRequest(t, fixture, method, path, body, cookies, tenantRef)
	response := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(response, request)
	return response
}

func workspaceInvitationClaimantRequest(
	t *testing.T,
	fixture localIdentityAdministrationHTTPFixture,
	method string,
	path string,
	body any,
	cookies []*http.Cookie,
	tenantRef string,
) *http.Request {
	t.Helper()
	request := localIdentityHTTPJSONRequest(t, method, path, body, cookies)
	if tenantRef != "" {
		request.Header.Set(localIdentityActiveTenantHeader, tenantRef)
	}
	request.Header.Set(activeWorkspaceHeader, "workspace_untrusted_claimant_header")
	if len(cookies) > 0 {
		request.Header.Set(localIdentityCSRFHeader,
			localIdentityCookieValue(t, cookies, fixture.identity.service.csrfCookieName()))
	}
	return request
}

func workspaceInvitationRawClaimantResponse(
	t *testing.T,
	fixture localIdentityAdministrationHTTPFixture,
	path string,
	payload string,
	cookies []*http.Cookie,
	tenantRef string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(payload))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", localIdentityHTTPTestOrigin)
	request.Header.Set(localIdentityActiveTenantHeader, tenantRef)
	addLocalIdentityCookies(request, cookies)
	response := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(response, request)
	return response
}

func assertWorkspaceInvitationHTTPError(
	t *testing.T,
	response *httptest.ResponseRecorder,
	status int,
	code string,
	recovery string,
) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("unexpected invitation HTTP status: got=%d want=%d body=%s", response.Code, status, response.Body.String())
	}
	var document errorDocument
	decodeLocalIdentityHTTPResponse(t, response, &document)
	if document.Error.Code != code || document.Error.RequestID == "" || document.Error.Metadata["recovery"] != recovery {
		t.Fatalf("unexpected invitation HTTP error: %#v", document.Error)
	}
}

func assertWorkspaceInvitationHTTPSafePayload(t *testing.T, payload string, allowCode bool, code string) {
	t.Helper()
	for _, forbidden := range []string{
		`"secret_digest"`, `"login_identifier"`, `"email"`, `"credential"`, `"session"`,
		`"cookie"`, `"issuer"`, `"subject"`, `"permission_grants"`,
	} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("workspace invitation response leaked %q: %s", forbidden, payload)
		}
	}
	if !allowCode && code != "" && strings.Contains(payload, code) {
		t.Fatalf("workspace invitation response repeated one-time code: %s", payload)
	}
}
