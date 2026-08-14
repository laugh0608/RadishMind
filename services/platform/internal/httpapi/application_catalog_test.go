package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

func TestApplicationCatalogLifecycleAndOwnerIsolation(t *testing.T) {
	runApplicationCatalogLifecycleAndOwnerIsolation(t, newMemoryApplicationCatalogRepository())
}

func runApplicationCatalogLifecycleAndOwnerIsolation(t *testing.T, repository applicationCatalogRepository) {
	t.Helper()
	service := newApplicationCatalogService(repository)
	service.newID = func() (string, error) { return "app_aaaaaaaaaaaaaaaa", nil }
	service.now = func() time.Time { return time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC) }
	owner := applicationCatalogTestContext("subject_owner")

	created := service.Create(owner, ApplicationCatalogCreateInput{DisplayName: "Docs Assistant", Description: "Internal documentation assistant", ApplicationKind: "docs_qa"})
	if created.FailureCode != "" || created.Record == nil || created.Record.RecordVersion != 1 || created.Record.LifecycleState != applicationCatalogLifecycleActive {
		t.Fatalf("unexpected create result: %#v", created)
	}
	if created.Record.ApplicationID != "app_aaaaaaaaaaaaaaaa" || created.Record.OwnerSubjectRef != owner.OwnerSubjectRef {
		t.Fatalf("unexpected generated identity: %#v", created.Record)
	}

	otherOwner := service.Read(applicationCatalogTestContext("subject_other"), created.Record.ApplicationID)
	if otherOwner.FailureCode != ApplicationCatalogFailureNotFound {
		t.Fatalf("cross-owner read must not reveal the record: %#v", otherOwner)
	}

	updated := service.Update(owner, created.Record.ApplicationID, ApplicationCatalogUpdateInput{
		ExpectedVersion: 1, DisplayName: "Docs Assistant v2", Description: "Updated metadata", ApplicationKind: "agent",
	})
	if updated.FailureCode != "" || updated.Record == nil || updated.Record.RecordVersion != 2 || updated.Record.ApplicationKind != "agent" {
		t.Fatalf("unexpected update result: %#v", updated)
	}
	conflict := service.Update(owner, created.Record.ApplicationID, ApplicationCatalogUpdateInput{
		ExpectedVersion: 1, DisplayName: "Stale edit", ApplicationKind: "agent",
	})
	if conflict.FailureCode != ApplicationCatalogFailureVersionConflict || conflict.CurrentRecordVersion != 2 || conflict.CurrentLifecycleState != applicationCatalogLifecycleActive {
		t.Fatalf("stale update must expose only current version and state: %#v", conflict)
	}

	archived := service.Archive(owner, created.Record.ApplicationID, 2)
	if archived.FailureCode != "" || archived.Record == nil || archived.Record.RecordVersion != 3 || archived.Record.ArchivedAt == nil {
		t.Fatalf("unexpected archive result: %#v", archived)
	}
	if active := service.RequireActive(owner, created.Record.ApplicationID); active.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived record must fail active requirement: %#v", active)
	}
	if update := service.Update(owner, created.Record.ApplicationID, ApplicationCatalogUpdateInput{ExpectedVersion: 3, DisplayName: "No update", ApplicationKind: "agent"}); update.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived record must reject metadata update: %#v", update)
	}
	if secondArchive := service.Archive(owner, created.Record.ApplicationID, 3); secondArchive.FailureCode != ApplicationCatalogFailureTransitionInvalid {
		t.Fatalf("archived record must reject a second transition: %#v", secondArchive)
	}
	staleUnarchive := service.Unarchive(owner, created.Record.ApplicationID, ApplicationCatalogUnarchiveInput{
		ExpectedVersion: 2, AcknowledgeExistingAccessReactivation: true,
	})
	if staleUnarchive.FailureCode != ApplicationCatalogFailureVersionConflict ||
		staleUnarchive.CurrentRecordVersion != 3 ||
		staleUnarchive.CurrentLifecycleState != applicationCatalogLifecycleArchived {
		t.Fatalf("stale unarchive must expose current archived version: %#v", staleUnarchive)
	}
	missingAcknowledgement := service.Unarchive(owner, created.Record.ApplicationID, ApplicationCatalogUnarchiveInput{ExpectedVersion: 3})
	if missingAcknowledgement.FailureCode != ApplicationCatalogFailurePayloadInvalid {
		t.Fatalf("unarchive without access reactivation acknowledgement must fail: %#v", missingAcknowledgement)
	}
	unarchived := service.Unarchive(owner, created.Record.ApplicationID, ApplicationCatalogUnarchiveInput{
		ExpectedVersion: 3, AcknowledgeExistingAccessReactivation: true,
	})
	if unarchived.FailureCode != "" || unarchived.Record == nil || unarchived.Record.RecordVersion != 4 ||
		unarchived.Record.LifecycleState != applicationCatalogLifecycleActive || unarchived.Record.ArchivedAt != nil {
		t.Fatalf("unexpected unarchive result: %#v", unarchived)
	}
	if active := service.RequireActive(owner, created.Record.ApplicationID); active.FailureCode != "" || active.Record == nil {
		t.Fatalf("unarchived record must satisfy active requirement: %#v", active)
	}
	if secondUnarchive := service.Unarchive(owner, created.Record.ApplicationID, ApplicationCatalogUnarchiveInput{
		ExpectedVersion: 4, AcknowledgeExistingAccessReactivation: true,
	}); secondUnarchive.FailureCode != ApplicationCatalogFailureTransitionInvalid {
		t.Fatalf("active record must reject a second unarchive: %#v", secondUnarchive)
	}
}

func TestApplicationCatalogValidationPaginationAndCAS(t *testing.T) {
	runApplicationCatalogValidationPaginationAndCAS(t, newMemoryApplicationCatalogRepository())
}

func runApplicationCatalogValidationPaginationAndCAS(t *testing.T, repository applicationCatalogRepository) {
	t.Helper()
	service := newApplicationCatalogService(repository)
	identifiers := []string{"app_aaaaaaaaaaaaaaaa", "app_bbbbbbbbbbbbbbbb", "app_cccccccccccccccc"}
	var identifierIndex int
	service.newID = func() (string, error) {
		identifier := identifiers[identifierIndex]
		identifierIndex++
		return identifier, nil
	}
	clock := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}
	requestContext := applicationCatalogTestContext("subject_owner")

	for index := 0; index < 3; index++ {
		result := service.Create(requestContext, ApplicationCatalogCreateInput{DisplayName: "App " + string(rune('A'+index)), ApplicationKind: "agent"})
		if result.FailureCode != "" {
			t.Fatalf("create %d failed: %#v", index, result)
		}
	}
	firstPage := service.List(requestContext, ApplicationCatalogListInput{Limit: 2, ApplicationKind: "agent"})
	if firstPage.FailureCode != "" || len(firstPage.Records) != 2 || firstPage.NextCursor == nil || firstPage.Records[0].ApplicationID != identifiers[2] {
		t.Fatalf("unexpected first page: %#v", firstPage)
	}
	secondPage := service.List(requestContext, ApplicationCatalogListInput{Limit: 2, ApplicationKind: "agent", Cursor: *firstPage.NextCursor})
	if secondPage.FailureCode != "" || len(secondPage.Records) != 1 || secondPage.Records[0].ApplicationID != identifiers[0] {
		t.Fatalf("unexpected second page: %#v", secondPage)
	}
	wrongOwner := service.List(applicationCatalogTestContext("subject_other"), ApplicationCatalogListInput{Limit: 2, ApplicationKind: "agent", Cursor: *firstPage.NextCursor})
	if wrongOwner.FailureCode != ApplicationCatalogFailureCursorInvalid {
		t.Fatalf("cursor binding must reject a different owner: %#v", wrongOwner)
	}
	secret := service.Create(requestContext, ApplicationCatalogCreateInput{DisplayName: "Unsafe App", Description: "Authorization: Bearer hidden", ApplicationKind: "agent"})
	if secret.FailureCode != ApplicationCatalogFailureSecretForbidden {
		t.Fatalf("secret material must be rejected: %#v", secret)
	}

	applicationID := identifiers[0]
	service.now = func() time.Time { return time.Date(2026, 7, 13, 13, 0, 0, 0, time.UTC) }
	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := service.Update(requestContext, applicationID, ApplicationCatalogUpdateInput{ExpectedVersion: 1, DisplayName: "Concurrent App", ApplicationKind: "agent"})
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case ApplicationCatalogFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected concurrent update: %#v", result)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}

	archived := service.Archive(requestContext, applicationID, 2)
	if archived.FailureCode != "" || archived.Record == nil {
		t.Fatalf("archive before concurrent unarchive: %#v", archived)
	}
	successes.Store(0)
	conflicts.Store(0)
	wait = sync.WaitGroup{}
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := service.Unarchive(requestContext, applicationID, ApplicationCatalogUnarchiveInput{
				ExpectedVersion: 3, AcknowledgeExistingAccessReactivation: true,
			})
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case ApplicationCatalogFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected concurrent unarchive: %#v", result)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("unarchive CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}
}

func TestApplicationCatalogHTTPScopesUnknownFieldsAndOIDCZeroQuery(t *testing.T) {
	repository := newMemoryApplicationCatalogRepository()
	server := &Server{
		config:                       config.Config{ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true},
		applicationCatalogRepository: repository,
		workspaceMembershipProvider:  newDeterministicDevTestWorkspaceMembershipProvider(),
	}
	auth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:dev", TenantBinding: "tenant_demo",
		SubjectBinding: "subject_owner", ScopeGrants: []string{"applications:read", "applications:write", "applications:archive"},
		ResourceBinding: ControlPlaneResourceBinding{TenantRef: "tenant_demo", TenantVerified: true},
		WorkspaceMemberships: []VerifiedWorkspaceMembershipAssertion{{
			TenantRef: "tenant_demo", SubjectRef: "subject_owner", WorkspaceID: "workspace_demo",
			PermissionGrants: []string{"applications:read", "applications:write", "applications:archive"},
			SourceRef:        "membership:test", PolicyVersion: workspaceMembershipPolicyVersion,
			ExpiresAt: time.Now().Add(time.Hour),
		}},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(`{"workspace_id":"workspace_demo","display_name":"Catalog App","description":"Owned app","application_kind":"agent"}`))
	request.Header.Set(activeWorkspaceHeader, "workspace_demo")
	request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
	recorder := httptest.NewRecorder()
	server.handleCreateApplicationCatalogRecord(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected create status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	var created applicationCatalogEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil || created.Record == nil || created.FailureCode != nil {
		t.Fatalf("unexpected create envelope: %#v err=%v", created, err)
	}

	applicationID := created.Record.ApplicationID
	read := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications/"+applicationID+"?workspace_id=workspace_demo", nil)
	read.SetPathValue("application_id", applicationID)
	read = read.WithContext(withControlPlaneReadFakeAuthContext(read.Context(), auth))
	readRecorder := httptest.NewRecorder()
	server.handleReadApplicationCatalogRecord(readRecorder, read)
	if readRecorder.Code != http.StatusOK || !strings.Contains(readRecorder.Body.String(), `"application_id":"`+applicationID+`"`) {
		t.Fatalf("unexpected read response: %d body=%s", readRecorder.Code, readRecorder.Body.String())
	}

	update := httptest.NewRequest(http.MethodPut, "/v1/user-workspace/applications/"+applicationID, strings.NewReader(`{"workspace_id":"workspace_demo","expected_version":1,"display_name":"Catalog App v2","description":"Updated app","application_kind":"agent"}`))
	update.SetPathValue("application_id", applicationID)
	update.Header.Set(activeWorkspaceHeader, "workspace_demo")
	update = update.WithContext(withControlPlaneReadFakeAuthContext(update.Context(), auth))
	updateRecorder := httptest.NewRecorder()
	server.handleUpdateApplicationCatalogRecord(updateRecorder, update)
	if updateRecorder.Code != http.StatusOK || !strings.Contains(updateRecorder.Body.String(), `"record_version":2`) {
		t.Fatalf("unexpected update response: %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}

	activeList := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications?workspace_id=workspace_demo&lifecycle_state=active&limit=10", nil)
	activeList = activeList.WithContext(withControlPlaneReadFakeAuthContext(activeList.Context(), auth))
	activeListRecorder := httptest.NewRecorder()
	server.handleListApplicationCatalogRecords(activeListRecorder, activeList)
	if activeListRecorder.Code != http.StatusOK || !strings.Contains(activeListRecorder.Body.String(), `"application_ref":"`+applicationID+`"`) {
		t.Fatalf("unexpected list response: %d body=%s", activeListRecorder.Code, activeListRecorder.Body.String())
	}

	archive := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+applicationID+"/archive", strings.NewReader(`{"workspace_id":"workspace_demo","expected_version":2}`))
	archive.SetPathValue("application_id", applicationID)
	archive.Header.Set(activeWorkspaceHeader, "workspace_demo")
	archive = archive.WithContext(withControlPlaneReadFakeAuthContext(archive.Context(), auth))
	archiveRecorder := httptest.NewRecorder()
	server.handleArchiveApplicationCatalogRecord(archiveRecorder, archive)
	if archiveRecorder.Code != http.StatusOK || !strings.Contains(archiveRecorder.Body.String(), `"lifecycle_state":"archived"`) {
		t.Fatalf("unexpected archive response: %d body=%s", archiveRecorder.Code, archiveRecorder.Body.String())
	}

	unarchive := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/"+applicationID+"/unarchive", strings.NewReader(
		`{"workspace_id":"workspace_demo","expected_version":3,"acknowledge_existing_access_reactivation":true}`,
	))
	unarchive.SetPathValue("application_id", applicationID)
	unarchive.Header.Set(activeWorkspaceHeader, "workspace_demo")
	unarchive = unarchive.WithContext(withControlPlaneReadFakeAuthContext(unarchive.Context(), auth))
	unarchiveRecorder := httptest.NewRecorder()
	server.handleUnarchiveApplicationCatalogRecord(unarchiveRecorder, unarchive)
	if unarchiveRecorder.Code != http.StatusOK ||
		!strings.Contains(unarchiveRecorder.Body.String(), `"lifecycle_state":"active"`) ||
		!strings.Contains(unarchiveRecorder.Body.String(), `"record_version":4`) {
		t.Fatalf("unexpected unarchive response: %d body=%s", unarchiveRecorder.Code, unarchiveRecorder.Body.String())
	}

	unknown := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(`{"workspace_id":"workspace_demo","display_name":"Catalog App","application_kind":"agent","owner_subject_ref":"subject_other"}`))
	unknown.Header.Set(activeWorkspaceHeader, "workspace_demo")
	unknown = unknown.WithContext(withControlPlaneReadFakeAuthContext(unknown.Context(), auth))
	unknownRecorder := httptest.NewRecorder()
	server.handleCreateApplicationCatalogRecord(unknownRecorder, unknown)
	if unknownRecorder.Code != http.StatusBadRequest {
		t.Fatalf("client-owned identity field must be rejected: %d body=%s", unknownRecorder.Code, unknownRecorder.Body.String())
	}

	readOnlyAuth := auth
	readOnlyAuth.ScopeGrants = []string{"applications:read"}
	denied := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(`{"workspace_id":"workspace_demo","display_name":"Denied App","application_kind":"agent"}`))
	denied.Header.Set(activeWorkspaceHeader, "workspace_demo")
	denied = denied.WithContext(withControlPlaneReadFakeAuthContext(denied.Context(), readOnlyAuth))
	deniedRecorder := httptest.NewRecorder()
	server.handleCreateApplicationCatalogRecord(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden || !strings.Contains(deniedRecorder.Body.String(), "scope_denied") {
		t.Fatalf("write scope must be independent: %d body=%s", deniedRecorder.Code, deniedRecorder.Body.String())
	}

	counting := &countingApplicationCatalogRepository{applicationCatalogRepository: repository}
	server.applicationCatalogRepository = counting
	oidcAuth := auth
	oidcAuth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
	list := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/applications?workspace_id=workspace_demo", nil)
	list = list.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), oidcAuth))
	listRecorder := httptest.NewRecorder()
	server.handleListApplicationCatalogRecords(listRecorder, list)
	if listRecorder.Code != http.StatusServiceUnavailable || !strings.Contains(listRecorder.Body.String(), "workspace_membership_unavailable") {
		t.Fatalf("OIDC membership gate must fail closed: %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	if counting.listCalls.Load() != 0 {
		t.Fatalf("OIDC membership failure must not query repository: %d", counting.listCalls.Load())
	}
}

func TestApplicationCatalogUnarchiveRequiresAcknowledgementAndCombinedPermissionsWithoutRepositoryAccess(t *testing.T) {
	now := time.Now().UTC()
	broadAuth := applicationCatalogMutationTestAuth(now)
	broadAuth.ScopeGrants = []string{"applications:archive", "applications:write"}
	broadAuth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:archive", "applications:write"}
	tests := []struct {
		name          string
		body          string
		mutate        func(*controlPlaneReadAuthContext)
		expectedCode  int
		expectedError string
	}{
		{
			name:         "acknowledgement missing",
			body:         `{"workspace_id":"workspace_demo","expected_version":2}`,
			expectedCode: http.StatusBadRequest, expectedError: ApplicationCatalogFailurePayloadInvalid,
		},
		{
			name:         "acknowledgement false",
			body:         `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":false}`,
			expectedCode: http.StatusBadRequest, expectedError: ApplicationCatalogFailurePayloadInvalid,
		},
		{
			name:         "unknown field",
			body:         `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true,"reactivate_everything":true}`,
			expectedCode: http.StatusBadRequest, expectedError: "INVALID_JSON",
		},
		{
			name: "write scope missing",
			body: `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true}`,
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.ScopeGrants = []string{"applications:archive"}
			},
			expectedCode: http.StatusForbidden, expectedError: "scope_denied",
		},
		{
			name: "archive scope missing",
			body: `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true}`,
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.ScopeGrants = []string{"applications:write"}
			},
			expectedCode: http.StatusForbidden, expectedError: "scope_denied",
		},
		{
			name: "write membership missing",
			body: `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true}`,
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:archive"}
			},
			expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied",
		},
		{
			name: "archive membership missing",
			body: `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true}`,
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:write"}
			},
			expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied",
		},
		{
			name: "OIDC membership unavailable",
			body: `{"workspace_id":"workspace_demo","expected_version":2,"acknowledge_existing_access_reactivation":true}`,
			mutate: func(auth *controlPlaneReadAuthContext) {
				auth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
			},
			expectedCode: http.StatusServiceUnavailable, expectedError: "workspace_membership_unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			baseRepository := newMemoryApplicationCatalogRepository()
			requestContext := applicationCatalogTestContext("subject_owner")
			record := ApplicationCatalogRecord{
				SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: "app_aaaaaaaaaaaaaaaa",
				TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID, OwnerSubjectRef: requestContext.OwnerSubjectRef,
				DisplayName: "Archived App", ApplicationKind: "agent", LifecycleState: applicationCatalogLifecycleArchived,
				RecordVersion: 2, CreatedAt: now.Format(time.RFC3339Nano), UpdatedAt: now.Format(time.RFC3339Nano),
				ArchivedAt:        optionalApplicationCatalogTimestamp(now.Format(time.RFC3339Nano)),
				CreatedByActorRef: requestContext.ActorRef, UpdatedByActorRef: requestContext.ActorRef,
				RequestID: requestContext.RequestID, AuditRef: requestContext.AuditRef,
			}
			if _, err := baseRepository.Create(requestContext, record); err != nil {
				t.Fatalf("seed archived application: %v", err)
			}
			repository := &countingApplicationCatalogRepository{applicationCatalogRepository: baseRepository}
			server := &Server{
				config:                       config.Config{ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true},
				applicationCatalogRepository: repository,
				workspaceMembershipProvider:  newDeterministicDevTestWorkspaceMembershipProvider(),
			}
			auth := cloneApplicationCatalogMutationTestAuth(broadAuth)
			if test.mutate != nil {
				test.mutate(&auth)
			}
			request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications/app_aaaaaaaaaaaaaaaa/unarchive", strings.NewReader(test.body))
			request.SetPathValue("application_id", "app_aaaaaaaaaaaaaaaa")
			request.Header.Set(activeWorkspaceHeader, "workspace_demo")
			request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			recorder := httptest.NewRecorder()

			server.handleUnarchiveApplicationCatalogRecord(recorder, request)

			if recorder.Code != test.expectedCode || !strings.Contains(recorder.Body.String(), test.expectedError) {
				t.Fatalf("unexpected unarchive denial: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if repository.totalCalls.Load() != 0 {
				t.Fatalf("unarchive denial reached repository %d times", repository.totalCalls.Load())
			}
		})
	}
}

func TestApplicationCatalogMutationAuthorizationDenialsDoNotQueryRepository(t *testing.T) {
	now := time.Now().UTC()
	baseAuth := applicationCatalogMutationTestAuth(now)
	tests := []struct {
		name          string
		body          string
		active        []string
		mutate        func(*controlPlaneReadAuthContext)
		omitAuth      bool
		expectedCode  int
		expectedError string
	}{
		{
			name: "identity missing", active: []string{"workspace_demo"}, omitAuth: true,
			expectedCode: http.StatusUnauthorized, expectedError: "identity_context_missing",
		},
		{
			name: "identity invalid", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.FailureCode = "auth_context_contract_mismatch"
				auth.FailureStatus = http.StatusUnauthorized
			}, expectedCode: http.StatusUnauthorized, expectedError: "auth_context_contract_mismatch",
		},
		{
			name: "identity permission denied", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.ScopeGrants = []string{"applications:read"}
			}, expectedCode: http.StatusForbidden, expectedError: "scope_denied",
		},
		{
			name: "selection missing before malformed payload", body: `{`,
			expectedCode: http.StatusBadRequest, expectedError: "workspace_selection_missing",
		},
		{
			name: "selection repeated", active: []string{"workspace_demo", "workspace_demo"},
			expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "selection invalid", active: []string{"workspace demo"},
			expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "membership missing", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships = nil
			}, expectedCode: http.StatusForbidden, expectedError: "workspace_membership_denied",
		},
		{
			name: "membership expired", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].ExpiresAt = now.Add(-time.Second)
			}, expectedCode: http.StatusForbidden, expectedError: "workspace_membership_expired",
		},
		{
			name: "membership tenant mismatch", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].TenantRef = "tenant_other"
			}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "membership subject mismatch", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].SubjectRef = "subject_other"
			}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "membership workspace mismatch", active: []string{"workspace_other"},
			expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "membership permission denied", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.WorkspaceMemberships[0].PermissionGrants = []string{"applications:read"}
			}, expectedCode: http.StatusForbidden, expectedError: "workspace_permission_denied",
		},
		{
			name: "payload workspace mismatch", body: `{"workspace_id":"workspace_other","display_name":"Denied App","application_kind":"agent"}`,
			active: []string{"workspace_demo"}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "payload workspace is not normalized", body: `{"workspace_id":" workspace_demo","display_name":"Denied App","application_kind":"agent"}`,
			active: []string{"workspace_demo"}, expectedCode: http.StatusForbidden, expectedError: "workspace_binding_mismatch",
		},
		{
			name: "OIDC membership unavailable", active: []string{"workspace_demo"}, mutate: func(auth *controlPlaneReadAuthContext) {
				auth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
			}, expectedCode: http.StatusServiceUnavailable, expectedError: "workspace_membership_unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &countingApplicationCatalogRepository{applicationCatalogRepository: newMemoryApplicationCatalogRepository()}
			server := &Server{
				config:                       config.Config{ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true},
				applicationCatalogRepository: repository,
				workspaceMembershipProvider:  newDeterministicDevTestWorkspaceMembershipProvider(),
			}
			body := test.body
			if body == "" {
				body = `{"workspace_id":"workspace_demo","display_name":"Denied App","application_kind":"agent"}`
			}
			request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(body))
			for _, active := range test.active {
				request.Header.Add(activeWorkspaceHeader, active)
			}
			if !test.omitAuth {
				auth := cloneApplicationCatalogMutationTestAuth(baseAuth)
				if test.mutate != nil {
					test.mutate(&auth)
				}
				request = request.WithContext(withControlPlaneReadFakeAuthContext(request.Context(), auth))
			}
			recorder := httptest.NewRecorder()

			server.handleCreateApplicationCatalogRecord(recorder, request)

			if recorder.Code != test.expectedCode || !strings.Contains(recorder.Body.String(), test.expectedError) {
				t.Fatalf("unexpected denial: status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			if repository.totalCalls.Load() != 0 {
				t.Fatalf("authorization denial reached repository %d times", repository.totalCalls.Load())
			}
		})
	}
}

func TestApplicationCatalogMutationDevAndSignedAuthorization(t *testing.T) {
	t.Run("dev headers", func(t *testing.T) {
		server := NewServer(config.Config{
			ControlPlaneReadDevAuthEnabled:    true,
			ApplicationCatalogDevHTTPEnabled:  true,
			ApplicationCatalogDevWriteEnabled: true,
		}, Options{BuildVersion: "test"})
		repository := &countingApplicationCatalogRepository{applicationCatalogRepository: newMemoryApplicationCatalogRepository()}
		server.applicationCatalogRepository = repository
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(
			`{"workspace_id":"workspace_demo","display_name":"Dev App","application_kind":"agent"}`,
		))
		request.Header.Set(controlPlaneReadDevIdentityHeader, "dev-mutation-consumer")
		request.Header.Set(controlPlaneReadDevTenantHeader, "tenant_demo")
		request.Header.Set(controlPlaneReadDevSubjectHeader, "subject_owner")
		request.Header.Set(controlPlaneReadDevScopesHeader, "applications:write")
		request.Header.Set(controlPlaneReadDevAuditHeader, "audit_dev_mutation_consumer")
		request.Header.Set(activeWorkspaceHeader, "workspace_demo")
		request.Header.Set(controlPlaneReadDevMembershipHeader, "workspace_demo")
		request.Header.Set(controlPlaneReadDevMembershipPermHeader, "applications:write")
		recorder := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK || repository.totalCalls.Load() != 1 {
			t.Fatalf("dev mutation authorization failed: status=%d calls=%d body=%s", recorder.Code, repository.totalCalls.Load(), recorder.Body.String())
		}
	})

	t.Run("signed test token", func(t *testing.T) {
		privateKey := generateSignedTestPrivateKey(t)
		server := newSignedTestControlPlaneReadServer(t, privateKey)
		server.config.ApplicationCatalogDevHTTPEnabled = true
		server.config.ApplicationCatalogDevWriteEnabled = true
		repository := &countingApplicationCatalogRepository{applicationCatalogRepository: newMemoryApplicationCatalogRepository()}
		server.applicationCatalogRepository = repository
		claims := validSignedTestClaims()
		claims["permissions"] = []string{"radishmind.applications.write"}
		claims["workspace_memberships"] = []map[string]any{{
			"workspace_id": "workspace_demo",
			"permissions":  []string{"applications:write"},
		}}
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(
			`{"workspace_id":"workspace_demo","display_name":"Signed App","application_kind":"agent"}`,
		))
		request.Header.Set(activeWorkspaceHeader, "workspace_demo")
		request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
		recorder := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK || repository.totalCalls.Load() != 1 {
			t.Fatalf("signed mutation authorization failed: status=%d calls=%d body=%s", recorder.Code, repository.totalCalls.Load(), recorder.Body.String())
		}
	})

	t.Run("expired signed identity", func(t *testing.T) {
		privateKey := generateSignedTestPrivateKey(t)
		server := newSignedTestControlPlaneReadServer(t, privateKey)
		server.config.ApplicationCatalogDevHTTPEnabled = true
		server.config.ApplicationCatalogDevWriteEnabled = true
		repository := &countingApplicationCatalogRepository{applicationCatalogRepository: newMemoryApplicationCatalogRepository()}
		server.applicationCatalogRepository = repository
		claims := validSignedTestClaims()
		claims["permissions"] = []string{"radishmind.applications.write"}
		claims["exp"] = time.Now().Add(-time.Minute).Unix()
		request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/applications", strings.NewReader(
			`{"workspace_id":"workspace_demo","display_name":"Expired App","application_kind":"agent"}`,
		))
		request.Header.Set(activeWorkspaceHeader, "workspace_demo")
		request.Header.Set("Authorization", "Bearer "+signControlPlaneReadTestToken(t, privateKey, "RS256", claims))
		recorder := httptest.NewRecorder()

		server.httpServer.Handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized || !strings.Contains(recorder.Body.String(), "auth_context_contract_mismatch") {
			t.Fatalf("expired signed identity was not rejected: status=%d body=%s", recorder.Code, recorder.Body.String())
		}
		if repository.totalCalls.Load() != 0 {
			t.Fatalf("expired signed identity reached repository %d times", repository.totalCalls.Load())
		}
	})
}

func TestApplicationCatalogArchiveBlocksDraftSaveAndPublishReview(t *testing.T) {
	catalogRepository := newMemoryApplicationCatalogRepository()
	draftRepository := newMemoryApplicationConfigurationDraftRepository()
	publishRepository := newMemoryApplicationPublishCandidateRepository()
	server := &Server{
		config:                       config.Config{ApplicationCatalogDevHTTPEnabled: true},
		applicationCatalogRepository: catalogRepository, applicationDraftRepository: draftRepository,
		applicationPublishCandidateRepository: publishRepository,
	}
	applicationID := "app_aaaaaaaaaaaaaaaa"
	catalogContext := applicationCatalogTestContext("subject_platform_ops")
	createdAt := "2026-07-13T12:00:00Z"
	if _, err := catalogRepository.Create(catalogContext, ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: applicationID, TenantRef: catalogContext.TenantRef,
		WorkspaceID: catalogContext.WorkspaceID, OwnerSubjectRef: catalogContext.OwnerSubjectRef,
		DisplayName: "Catalog App", Description: "Lifecycle integration", ApplicationKind: "workflow_copilot",
		LifecycleState: applicationCatalogLifecycleActive, RecordVersion: 1, CreatedAt: createdAt, UpdatedAt: createdAt,
		CreatedByActorRef: catalogContext.ActorRef, UpdatedByActorRef: catalogContext.ActorRef,
		RequestID: catalogContext.RequestID, AuditRef: catalogContext.AuditRef,
	}); err != nil {
		t.Fatalf("seed active catalog record: %v", err)
	}

	draftContext := validApplicationDraftContext()
	draftContext.ApplicationID = applicationID
	payload := validApplicationDraftPayload()
	payload.ApplicationID = applicationID
	payload.DraftID = "app-config-catalog-lifecycle"
	payload.BaseApplicationUpdatedAt = createdAt
	saved := server.applicationConfigurationDraftService().Save(draftContext, payload, 0)
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("active application must allow draft save: %#v", saved)
	}

	publishContext := validApplicationPublishContext()
	publishContext.ApplicationID = applicationID
	publishContext.OwnerSubjectRef = catalogContext.OwnerSubjectRef
	publishContext.ActorRef = catalogContext.ActorRef
	publishContext.WorkspaceID = catalogContext.WorkspaceID
	publishContext.TenantRef = catalogContext.TenantRef
	candidate := server.applicationPublishCandidateService().Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "publish-catalog-lifecycle", DraftID: payload.DraftID, ExpectedDraftVersion: 1,
	})
	if candidate.FailureCode != "" || candidate.Candidate == nil {
		t.Fatalf("active application must allow publish candidate creation: %#v", candidate)
	}

	archiveService := newApplicationCatalogService(catalogRepository)
	archived := archiveService.Archive(catalogContext, applicationID, 1)
	if archived.FailureCode != "" {
		t.Fatalf("archive application: %#v", archived)
	}
	blockedDraft := server.applicationConfigurationDraftService().Save(draftContext, payload, 1)
	if blockedDraft.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived application must block draft writes: %#v", blockedDraft)
	}
	blockedCreate := server.applicationPublishCandidateService().Create(publishContext, ApplicationPublishCreateInput{
		CandidateID: "publish-catalog-after-archive", DraftID: payload.DraftID, ExpectedDraftVersion: 1,
	})
	if blockedCreate.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived application must block candidate creation: %#v", blockedCreate)
	}
	blockedReview := server.applicationPublishCandidateService().Review(publishContext, candidate.Candidate.CandidateID, ApplicationPublishReviewInput{
		ExpectedReviewVersion: 0, Decision: applicationPublishDecisionApprove, Reason: "Evidence reviewed and approved.",
	})
	if blockedReview.FailureCode != ApplicationCatalogFailureArchived {
		t.Fatalf("archived application must block candidate review: %#v", blockedReview)
	}

	unarchived := archiveService.Unarchive(catalogContext, applicationID, ApplicationCatalogUnarchiveInput{
		ExpectedVersion: 2, AcknowledgeExistingAccessReactivation: true,
	})
	if unarchived.FailureCode != "" || unarchived.Record == nil {
		t.Fatalf("unarchive application: %#v", unarchived)
	}
	payload.BaseApplicationUpdatedAt = unarchived.Record.UpdatedAt
	resumedDraft := server.applicationConfigurationDraftService().Save(draftContext, payload, 1)
	if resumedDraft.FailureCode != "" || resumedDraft.Draft == nil || resumedDraft.Draft.DraftVersion != 2 {
		t.Fatalf("unarchived application must allow a baseline-refreshed draft save: %#v", resumedDraft)
	}
}

type countingApplicationCatalogRepository struct {
	applicationCatalogRepository
	totalCalls atomic.Int32
	listCalls  atomic.Int32
}

func (repository *countingApplicationCatalogRepository) Create(requestContext ApplicationCatalogContext, record ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.Create(requestContext, record)
}

func (repository *countingApplicationCatalogRepository) Read(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.Read(requestContext, applicationID)
}

func (repository *countingApplicationCatalogRepository) List(requestContext ApplicationCatalogContext, query applicationCatalogListQuery) ([]ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	repository.listCalls.Add(1)
	return repository.applicationCatalogRepository.List(requestContext, query)
}

func (repository *countingApplicationCatalogRepository) UpdateMetadata(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.UpdateMetadata(requestContext, applicationID, expectedVersion, update)
}

func (repository *countingApplicationCatalogRepository) Archive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.Archive(requestContext, applicationID, expectedVersion, update)
}

func (repository *countingApplicationCatalogRepository) Unarchive(requestContext ApplicationCatalogContext, applicationID string, expectedVersion int, update ApplicationCatalogRecord) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.Unarchive(requestContext, applicationID, expectedVersion, update)
}

func (repository *countingApplicationCatalogRepository) RequireActive(requestContext ApplicationCatalogContext, applicationID string) (ApplicationCatalogRecord, error) {
	repository.totalCalls.Add(1)
	return repository.applicationCatalogRepository.RequireActive(requestContext, applicationID)
}

func applicationCatalogMutationTestAuth(now time.Time) controlPlaneReadAuthContext {
	return controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "verified:mutation-test",
		TenantBinding: "tenant_demo", SubjectBinding: "subject_owner",
		ScopeGrants: []string{"applications:write"},
		ResourceBinding: ControlPlaneResourceBinding{
			TenantRef: "tenant_demo", TenantVerified: true,
		},
		WorkspaceMemberships: []VerifiedWorkspaceMembershipAssertion{{
			TenantRef: "tenant_demo", SubjectRef: "subject_owner", WorkspaceID: "workspace_demo",
			PermissionGrants: []string{"applications:write"}, SourceRef: "membership:test",
			PolicyVersion: workspaceMembershipPolicyVersion, ExpiresAt: now.Add(time.Hour),
		}},
	}
}

func cloneApplicationCatalogMutationTestAuth(auth controlPlaneReadAuthContext) controlPlaneReadAuthContext {
	cloned := auth
	cloned.ScopeGrants = append([]string{}, auth.ScopeGrants...)
	cloned.WorkspaceMemberships = append([]VerifiedWorkspaceMembershipAssertion{}, auth.WorkspaceMemberships...)
	for index := range cloned.WorkspaceMemberships {
		cloned.WorkspaceMemberships[index].PermissionGrants = append(
			[]string{},
			auth.WorkspaceMemberships[index].PermissionGrants...,
		)
	}
	return cloned
}

func applicationCatalogTestContext(owner string) ApplicationCatalogContext {
	return ApplicationCatalogContext{
		RequestContext: context.Background(), RequestID: "request_catalog_test", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner,
		AuditRef: "audit_catalog_test", WriteEnabled: true,
	}
}

func optionalApplicationCatalogTimestamp(value string) *string {
	return &value
}
