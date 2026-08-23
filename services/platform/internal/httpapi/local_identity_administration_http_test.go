package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type localIdentityAdministrationHTTPFixture struct {
	identity      *localIdentityHTTPTestFixture
	administrator localIdentityAuthenticationDocument
	adminCookies  []*http.Cookie
	target        localIdentityAuthenticationDocument
	targetCookies []*http.Cookie
	bootstrap     LocalIdentityWorkspaceAdministratorBootstrap
}

func TestLocalIdentityAdministrationHTTPSevenRouteVerticalChain(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)

	catalogResponse := fixture.request(t, http.MethodGet, localIdentityAdminRoleCatalogPath, nil, fixture.adminCookies)
	if catalogResponse.Code != http.StatusOK {
		t.Fatalf("read role catalog: status=%d body=%s", catalogResponse.Code, catalogResponse.Body.String())
	}
	var catalogDocument localIdentityAdminRoleCatalogResponse
	decodeLocalIdentityHTTPResponse(t, catalogResponse, &catalogDocument)
	reader := roleDefinitionByKey(t, catalogDocument.Catalog, localIdentityRoleWorkspaceReader)

	membershipResponse := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminMembershipCreatePath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	if membershipResponse.Code != http.StatusCreated {
		t.Fatalf("create membership: status=%d body=%s", membershipResponse.Code, membershipResponse.Body.String())
	}
	var membershipDocument localIdentityAdminMembershipResponse
	decodeLocalIdentityHTTPResponse(t, membershipResponse, &membershipDocument)

	assignmentResponse := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminRoleAssignPath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "role_key": reader.RoleKey,
		"expected_catalog_version":        reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"confirmed":                       true,
	}, fixture.adminCookies)
	if assignmentResponse.Code != http.StatusCreated {
		t.Fatalf("assign role: status=%d body=%s", assignmentResponse.Code, assignmentResponse.Body.String())
	}
	var assignmentDocument localIdentityAdminRoleAssignmentResponse
	decodeLocalIdentityHTTPResponse(t, assignmentResponse, &assignmentDocument)
	if !assignmentDocument.RoleAssignment.Effective || assignmentDocument.RoleAssignment.RoleKey != reader.RoleKey {
		t.Fatalf("role assignment projection mismatch: %#v", assignmentDocument.RoleAssignment)
	}

	memberListPath := strings.Replace(localIdentityAdminMemberListPath, "{workspace_id}", "workspace_demo", 1) + "?limit=1"
	listResponse := fixture.request(t, http.MethodGet, memberListPath, nil, fixture.adminCookies)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list members: status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	var listDocument localIdentityAdminMemberListResponse
	decodeLocalIdentityHTTPResponse(t, listResponse, &listDocument)
	if len(listDocument.Members) != 1 || listDocument.NextCursor == "" {
		t.Fatalf("member directory pagination mismatch: %#v", listDocument)
	}

	detailPath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{user_id}", fixture.target.Account.UserID,
	).Replace(localIdentityAdminMemberReadPath)
	detailResponse := fixture.request(t, http.MethodGet, detailPath, nil, fixture.adminCookies)
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("read member detail: status=%d body=%s", detailResponse.Code, detailResponse.Body.String())
	}
	var detailDocument localIdentityAdminMemberReadResponse
	decodeLocalIdentityHTTPResponse(t, detailResponse, &detailDocument)
	if detailDocument.Member.UserID != fixture.target.Account.UserID || len(detailDocument.Member.RoleAssignments) != 1 {
		t.Fatalf("member detail mismatch: %#v", detailDocument.Member)
	}

	roleRevokePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{assignment_id}", assignmentDocument.RoleAssignment.AssignmentID,
	).Replace(localIdentityAdminRoleRevokePath)
	roleRevokeResponse := fixture.request(t, http.MethodPost, roleRevokePath, map[string]any{
		"expected_record_version": assignmentDocument.RoleAssignment.RecordVersion, "confirmed": true,
	}, fixture.adminCookies)
	if roleRevokeResponse.Code != http.StatusOK {
		t.Fatalf("revoke role: status=%d body=%s", roleRevokeResponse.Code, roleRevokeResponse.Body.String())
	}

	reassignmentResponse := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminRoleAssignPath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "role_key": reader.RoleKey,
		"expected_catalog_version":        reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"confirmed":                       true,
	}, fixture.adminCookies)
	if reassignmentResponse.Code != http.StatusCreated {
		t.Fatalf("reassign role before aggregate revoke: status=%d body=%s", reassignmentResponse.Code, reassignmentResponse.Body.String())
	}

	membershipRevokePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{membership_id}", membershipDocument.Membership.MembershipID,
	).Replace(localIdentityAdminMembershipRevokePath)
	membershipRevokeResponse := fixture.request(t, http.MethodPost, membershipRevokePath, map[string]any{
		"expected_record_version": membershipDocument.Membership.RecordVersion, "confirmed": true,
	}, fixture.adminCookies)
	if membershipRevokeResponse.Code != http.StatusOK {
		t.Fatalf("revoke membership: status=%d body=%s", membershipRevokeResponse.Code, membershipRevokeResponse.Body.String())
	}
	var revocationDocument localIdentityAdminMembershipResponse
	decodeLocalIdentityHTTPResponse(t, membershipRevokeResponse, &revocationDocument)
	if revocationDocument.Membership.Effective || len(revocationDocument.RevokedRoleAssignments) != 1 {
		t.Fatalf("aggregate revocation projection mismatch: %#v", revocationDocument)
	}

	for _, payload := range []string{
		catalogResponse.Body.String(), membershipResponse.Body.String(), assignmentResponse.Body.String(),
		detailResponse.Body.String(), membershipRevokeResponse.Body.String(),
	} {
		for _, forbidden := range []string{`"login_identifier"`, `"credential"`, `"session"`, `"audit_ref"`, `"password"`} {
			if strings.Contains(payload, forbidden) {
				t.Fatalf("administration response leaked %q: %s", forbidden, payload)
			}
		}
	}
}

func TestLocalIdentityAdministrationHTTPAuthenticationAndMutationGuards(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)
	catalog := LocalIdentityBuiltInRoleCatalog()
	reader := roleDefinitionByKey(t, catalog, localIdentityRoleWorkspaceReader)
	membershipPath := strings.Replace(localIdentityAdminMembershipCreatePath, "{workspace_id}", "workspace_demo", 1)

	unauthenticated := fixture.request(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, nil)
	assertLocalIdentityError(t, unauthenticated, http.StatusUnauthorized, localIdentityAuthenticationRequired)
	nonMember := fixture.request(t, http.MethodGet, localIdentityAdminRoleCatalogPath, nil, fixture.targetCookies)
	assertLocalIdentityError(t, nonMember, http.StatusForbidden, LocalIdentityFailureMembershipDenied)

	fallbackRequest := fixture.adminRequest(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	fallbackRequest.Header.Set("Authorization", "Bearer forbidden")
	fallbackResponse := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(fallbackResponse, fallbackRequest)
	assertLocalIdentityError(t, fallbackResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)
	devHeaderRequest := fixture.adminRequest(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	devHeaderRequest.Header.Set(controlPlaneReadDevIdentityHeader, "verified:dev")
	devHeaderResponse := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(devHeaderResponse, devHeaderRequest)
	assertLocalIdentityError(t, devHeaderResponse, http.StatusUnauthorized, localIdentityAuthenticationRequired)

	crossScopePath := strings.Replace(localIdentityAdminMembershipCreatePath, "{workspace_id}", "workspace_other", 1)
	crossScope := fixture.request(t, http.MethodPost, crossScopePath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, crossScope, http.StatusForbidden, LocalIdentityFailureAdminScopeMismatch)

	wrongOriginRequest := fixture.adminRequest(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	wrongOriginRequest.Header.Set("Origin", "https://evil.example")
	wrongOrigin := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(wrongOrigin, wrongOriginRequest)
	assertLocalIdentityError(t, wrongOrigin, http.StatusForbidden, localIdentityOriginForbidden)

	missingCSRFRequest := fixture.adminRequest(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	missingCSRFRequest.Header.Del(localIdentityCSRFHeader)
	missingCSRF := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(missingCSRF, missingCSRFRequest)
	assertLocalIdentityError(t, missingCSRF, http.StatusForbidden, localIdentityCSRFInvalid)

	grantsInjection := fixture.request(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
		"permission_grants": []string{localIdentityPermissionRolesAssign},
	}, fixture.adminCookies)
	assertLocalIdentityError(t, grantsInjection, http.StatusBadRequest, "INVALID_JSON")
	if _, err := fixture.identity.repository.AuthorizeWorkspace(
		context.Background(), fixture.target.Account.UserID, "tenant_demo", "workspace_demo",
		[]string{localIdentityPermissionMembersRead}, fixture.identity.service.nowUTC(),
	); err == nil {
		t.Fatal("grants injection created a membership")
	}
	missingConfirmation := fixture.request(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": false,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, missingConfirmation, http.StatusBadRequest, localIdentityPayloadInvalid)
	missingTarget := fixture.request(t, http.MethodPost, membershipPath, map[string]any{
		"user_id": "usr_00000000000000ff", "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, missingTarget, http.StatusNotFound, LocalIdentityFailureMemberUnavailable)

	fixture.createTargetMembership(t)
	unknownRole := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminRoleAssignPath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "role_key": "workspace_superuser",
		"expected_catalog_version":        catalog.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"confirmed":                       true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, unknownRole, http.StatusConflict, LocalIdentityFailureRoleCatalogMismatch)
	var catalogFailure errorDocument
	decodeLocalIdentityHTTPResponse(t, unknownRole, &catalogFailure)
	if catalogFailure.Error.Metadata["recovery"] != "refresh_role_catalog" {
		t.Fatalf("catalog failure recovery metadata mismatch: %#v", catalogFailure.Error.Metadata)
	}

	permissionDenied := fixture.request(t, http.MethodGet, localIdentityAdminRoleCatalogPath, nil, fixture.targetCookies)
	assertLocalIdentityError(t, permissionDenied, http.StatusForbidden, LocalIdentityFailurePermissionDenied)

	fixture.identity.setNow(fixture.identity.service.nowUTC().Add(localIdentityAdministrationRecentAuthenticationAge + time.Second))
	staleAuthentication := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminRoleAssignPath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "role_key": reader.RoleKey,
		"expected_catalog_version":        reader.CatalogVersion,
		"expected_role_definition_digest": reader.DefinitionDigest,
		"confirmed":                       true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, staleAuthentication, http.StatusUnauthorized, LocalIdentityFailureRecentAuthentication)

	invalidCursor := fixture.request(t, http.MethodGet, strings.Replace(
		localIdentityAdminMemberListPath, "{workspace_id}", "workspace_demo", 1,
	)+"?limit=0", nil, fixture.adminCookies)
	assertLocalIdentityError(t, invalidCursor, http.StatusBadRequest, LocalIdentityFailureMemberCursorInvalid)
	methodMismatch := fixture.request(t, http.MethodPut, membershipPath, map[string]any{}, fixture.adminCookies)
	if methodMismatch.Code != http.StatusMethodNotAllowed {
		t.Fatalf("strict route accepted wrong method: status=%d body=%s", methodMismatch.Code, methodMismatch.Body.String())
	}
	bootstrapHTTP := fixture.request(t, http.MethodPost, "/v1/admin/local-identity/bootstrap", map[string]any{}, fixture.adminCookies)
	if bootstrapHTTP.Code != http.StatusNotFound {
		t.Fatalf("first-admin bootstrap became an HTTP route: status=%d body=%s", bootstrapHTTP.Code, bootstrapHTTP.Body.String())
	}
}

func TestLocalIdentityAdministrationHTTPRequiresExactPermission(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)
	membership := fixture.createTargetMembership(t)
	underlying := fixture.identity.repository.(*memoryLocalIdentityRepository)
	repository := &localIdentityAdministrationHTTPExactPermissionRepository{
		localIdentityRepository:               underlying,
		localIdentityAdministrationRepository: underlying,
		underlying:                            underlying,
		userID:                                fixture.target.Account.UserID,
		permission:                            localIdentityPermissionMembersRead,
	}
	fixture.identity.service.repository = repository
	fixture.identity.server.localIdentityAdministrationService.repository = repository

	list := fixture.request(t, http.MethodGet, strings.Replace(
		localIdentityAdminMemberListPath, "{workspace_id}", "workspace_demo", 1,
	), nil, fixture.targetCookies)
	if list.Code != http.StatusOK {
		t.Fatalf("exact member read permission was denied: status=%d body=%s", list.Code, list.Body.String())
	}
	catalog := fixture.request(t, http.MethodGet, localIdentityAdminRoleCatalogPath, nil, fixture.targetCookies)
	assertLocalIdentityError(t, catalog, http.StatusForbidden, LocalIdentityFailurePermissionDenied)
	revokePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{membership_id}", membership.MembershipID,
	).Replace(localIdentityAdminMembershipRevokePath)
	revoke := fixture.request(t, http.MethodPost, revokePath, map[string]any{
		"expected_record_version": membership.RecordVersion, "confirmed": true,
	}, fixture.targetCookies)
	assertLocalIdentityError(t, revoke, http.StatusForbidden, LocalIdentityFailurePermissionDenied)
}

type localIdentityAdministrationHTTPExactPermissionRepository struct {
	localIdentityRepository
	localIdentityAdministrationRepository
	underlying *memoryLocalIdentityRepository
	userID     string
	permission string
}

func (repository *localIdentityAdministrationHTTPExactPermissionRepository) AuthorizeWorkspace(
	ctx context.Context,
	userID string,
	tenantRef string,
	workspaceID string,
	required []string,
	now time.Time,
) (LocalWorkspaceAuthorization, error) {
	if repository == nil || repository.underlying == nil {
		return LocalWorkspaceAuthorization{}, errLocalIdentityStoreUnavailable
	}
	if userID != repository.userID || len(required) != 1 || required[0] != repository.permission {
		return repository.underlying.AuthorizeWorkspace(ctx, userID, tenantRef, workspaceID, required, now)
	}
	repository.underlying.mu.RLock()
	defer repository.underlying.mu.RUnlock()
	account, exists := repository.underlying.accounts[userID]
	membershipKey := repository.underlying.activeMembershipByScope[localMembershipScopeKey(userID, tenantRef, workspaceID)]
	membership, member := repository.underlying.memberships[membershipKey]
	if !exists || account.LifecycleState != localIdentityStateActive || !member ||
		!localIdentityMembershipEffective(membership, now) {
		return LocalWorkspaceAuthorization{}, errLocalIdentityMembershipDenied
	}
	return LocalWorkspaceAuthorization{
		Account: cloneUserAccount(account), Membership: cloneWorkspaceMembership(membership),
		PermissionGrants: []string{repository.permission},
	}, nil
}

func TestLocalIdentityAdministrationHTTPCASAndAdministratorProtections(t *testing.T) {
	fixture := newLocalIdentityAdministrationHTTPFixture(t)

	selfMembershipPath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{membership_id}", fixture.bootstrap.Membership.MembershipID,
	).Replace(localIdentityAdminMembershipRevokePath)
	selfRevoke := fixture.request(t, http.MethodPost, selfMembershipPath, map[string]any{
		"expected_record_version": fixture.bootstrap.Membership.RecordVersion, "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, selfRevoke, http.StatusConflict, LocalIdentityFailureSelfMembershipRevoke)

	lastAdminRolePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{assignment_id}", fixture.bootstrap.RoleAssignment.AssignmentID,
	).Replace(localIdentityAdminRoleRevokePath)
	lastAdminRevoke := fixture.request(t, http.MethodPost, lastAdminRolePath, map[string]any{
		"expected_record_version": fixture.bootstrap.RoleAssignment.RecordVersion, "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, lastAdminRevoke, http.StatusConflict, LocalIdentityFailureLastAdminRemoval)

	membership := fixture.createTargetMembership(t)
	stalePath := strings.NewReplacer(
		"{workspace_id}", "workspace_demo", "{membership_id}", membership.MembershipID,
	).Replace(localIdentityAdminMembershipRevokePath)
	stale := fixture.request(t, http.MethodPost, stalePath, map[string]any{
		"expected_record_version": membership.RecordVersion + 1, "confirmed": true,
	}, fixture.adminCookies)
	assertLocalIdentityError(t, stale, http.StatusConflict, LocalIdentityFailureMembershipConflict)
	detail, err := fixture.identity.server.localIdentityAdministrationService.ReadWorkspaceMember(
		context.Background(), LocalIdentityAdministrationActor{
			UserID: fixture.administrator.Account.UserID, TenantRef: "tenant_demo", WorkspaceID: "workspace_demo",
			AuthenticatedAt: fixture.identity.service.nowUTC(),
		}, "tenant_demo", "workspace_demo", fixture.target.Account.UserID,
	)
	if err != nil || len(detail.Memberships) != 1 || !detail.Memberships[0].Effective {
		t.Fatalf("stale CAS changed membership: detail=%#v err=%v", detail, err)
	}
}

func newLocalIdentityAdministrationHTTPFixture(t *testing.T) localIdentityAdministrationHTTPFixture {
	t.Helper()
	now := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	identity := newLocalIdentityHTTPTestFixture(newMemoryLocalIdentityRepository(), now)
	_, adminCookies, administrator := registerLocalIdentityHTTPTestAccount(
		t, identity, "administrator@example.com", "administrator password with enough entropy",
	)
	_, targetCookies, target := registerLocalIdentityHTTPTestAccount(
		t, identity, "target@example.com", "target password with enough entropy",
	)
	bootstrap, err := identity.server.localIdentityAdministrationService.BootstrapWorkspaceAdministrator(
		context.Background(), LocalIdentityBootstrapWorkspaceAdministratorInput{
			TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", UserID: administrator.Account.UserID,
			AuditRef: "audit:http-admin-bootstrap",
		},
	)
	if err != nil {
		t.Fatalf("bootstrap HTTP administrator: %v", err)
	}
	return localIdentityAdministrationHTTPFixture{
		identity: identity, administrator: administrator, adminCookies: adminCookies,
		target: target, targetCookies: targetCookies, bootstrap: bootstrap,
	}
}

func (fixture localIdentityAdministrationHTTPFixture) adminRequest(
	t *testing.T,
	method string,
	path string,
	body any,
	cookies []*http.Cookie,
) *http.Request {
	t.Helper()
	var request *http.Request
	if method == http.MethodGet {
		request = httptest.NewRequest(method, path, nil)
	} else {
		request = localIdentityHTTPJSONRequest(t, method, path, body, cookies)
	}
	if method == http.MethodGet {
		addLocalIdentityCookies(request, cookies)
	}
	request.Header.Set(localIdentityActiveTenantHeader, "tenant_demo")
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	if len(cookies) > 0 {
		request.Header.Set(localIdentityCSRFHeader, localIdentityCookieValue(t, cookies, fixture.identity.service.csrfCookieName()))
	}
	return request
}

func (fixture localIdentityAdministrationHTTPFixture) request(
	t *testing.T,
	method string,
	path string,
	body any,
	cookies []*http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()
	request := fixture.adminRequest(t, method, path, body, cookies)
	response := httptest.NewRecorder()
	fixture.identity.handler.ServeHTTP(response, request)
	return response
}

func (fixture localIdentityAdministrationHTTPFixture) createTargetMembership(t *testing.T) LocalIdentityWorkspaceMembershipView {
	t.Helper()
	response := fixture.request(t, http.MethodPost, strings.Replace(
		localIdentityAdminMembershipCreatePath, "{workspace_id}", "workspace_demo", 1,
	), map[string]any{
		"user_id": fixture.target.Account.UserID, "confirmed": true,
	}, fixture.adminCookies)
	if response.Code != http.StatusCreated {
		t.Fatalf("create target membership: status=%d body=%s", response.Code, response.Body.String())
	}
	var document localIdentityAdminMembershipResponse
	decodeLocalIdentityHTTPResponse(t, response, &document)
	return document.Membership
}

func roleDefinitionByKey(t *testing.T, catalog LocalIdentityRoleCatalog, roleKey string) LocalIdentityRoleDefinition {
	t.Helper()
	for _, definition := range catalog.Roles {
		if definition.RoleKey == roleKey {
			return definition
		}
	}
	t.Fatalf("role %q is missing from catalog", roleKey)
	return LocalIdentityRoleDefinition{}
}
