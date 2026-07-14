package httpapi

import (
	"bytes"
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

func TestAPIKeyServiceIssueReadListAndRevoke(t *testing.T) {
	runAPIKeyServiceIssueReadListAndRevoke(t, newMemoryAPIKeyRepository(), newMemoryApplicationCatalogRepository())
}

func runAPIKeyServiceIssueReadListAndRevoke(t *testing.T, repository apiKeyRepository, applicationRepository applicationCatalogRepository) {
	t.Helper()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	service := newAPIKeyService(repository, applicationRepository)
	service.now = func() time.Time { return time.Date(2026, 7, 13, 8, 0, 0, 0, time.UTC) }
	service.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }

	issued := service.Create(requestContext, APIKeyCreateInput{
		ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "Primary SDK key",
		Scopes: []string{"responses:invoke", "models:read", "models:read"}, ExpiresInDays: 30,
	})
	if issued.FailureCode != "" || issued.Record == nil || issued.CredentialToken == "" {
		t.Fatalf("issue API key: failure=%s record_present=%t credential_present=%t", issued.FailureCode, issued.Record != nil, issued.CredentialToken != "")
	}
	if issued.Record.SchemaVersion != apiKeyRecordSchemaVersion || issued.Record.RecordVersion != 1 ||
		issued.Record.EffectiveState != apiKeyLifecycleActive || issued.Record.ExpiresAt != "2026-08-12T08:00:00Z" {
		t.Fatalf("unexpected issued record: %#v", issued.Record)
	}
	if strings.Join(issued.Record.Scopes, ",") != "models:read,responses:invoke" {
		t.Fatalf("scopes must be normalized: %#v", issued.Record.Scopes)
	}
	if id, ok := parseAPIKeyCredential(issued.CredentialToken); !ok || id != issued.Record.APIKeyID {
		t.Fatalf("issued credential format is invalid")
	}
	stored, err := repository.Read(requestContext, issued.Record.APIKeyID)
	if err != nil || !apiKeyCredentialMatches(issued.CredentialToken, stored.credentialDigest) {
		t.Fatalf("stored digest must validate the issued credential: %v", err)
	}
	encoded, err := json.Marshal(issued.Record)
	if err != nil {
		t.Fatalf("marshal issued record: %v", err)
	}
	if bytes.Contains(encoded, []byte(issued.CredentialToken)) || bytes.Contains(encoded, []byte("credential_digest")) || bytes.Contains(encoded, []byte("credentialDigest")) {
		t.Fatalf("public record leaked credential material: %s", encoded)
	}

	read := service.Read(requestContext, issued.Record.APIKeyID)
	if read.FailureCode != "" || read.Record == nil || read.CredentialToken != "" {
		t.Fatalf("read must return only the record: %#v", read)
	}
	listed := service.List(requestContext, APIKeyListInput{ApplicationID: issued.Record.ApplicationID, EffectiveState: apiKeyLifecycleActive})
	if listed.FailureCode != "" || len(listed.Records) != 1 {
		t.Fatalf("list active API keys: %#v", listed)
	}

	revoked := service.Revoke(requestContext, issued.Record.APIKeyID, 1)
	if revoked.FailureCode != "" || revoked.Record == nil || revoked.Record.RecordVersion != 2 ||
		revoked.Record.EffectiveState != apiKeyLifecycleRevoked || revoked.Record.RevokedAt == nil {
		t.Fatalf("revoke API key: %#v", revoked)
	}
	if second := service.Revoke(requestContext, issued.Record.APIKeyID, 2); second.FailureCode != APIKeyFailureRevoked {
		t.Fatalf("revoked key must not be revoked again: %#v", second)
	}
	if conflict := service.Revoke(requestContext, issued.Record.APIKeyID, 1); conflict.FailureCode != APIKeyFailureVersionConflict || conflict.CurrentRecordVersion != 2 {
		t.Fatalf("stale revoke must expose sanitized current version: %#v", conflict)
	}
}

func TestAPIKeyServiceRejectsInvalidAndInactiveApplicationBeforeCredentialGeneration(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleArchived)
	repository := newMemoryAPIKeyRepository()
	service := newAPIKeyService(repository, applicationRepository)
	credentialCalls := 0
	service.newCredential = func(string) (string, [32]byte, error) {
		credentialCalls++
		return "", [32]byte{}, nil
	}

	tests := []struct {
		name     string
		input    APIKeyCreateInput
		expected string
	}{
		{name: "expired duration", input: validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 0), expected: APIKeyFailurePayloadInvalid},
		{name: "unknown scope", input: APIKeyCreateInput{ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "SDK key", Scopes: []string{"admin:write"}, ExpiresInDays: 30}, expected: APIKeyFailurePayloadInvalid},
		{name: "secret in label", input: APIKeyCreateInput{ApplicationID: "app_aaaaaaaaaaaaaaaa", DisplayName: "Authorization: Bearer hidden", Scopes: []string{"models:read"}, ExpiresInDays: 30}, expected: APIKeyFailureSecretForbidden},
		{name: "archived application", input: validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30), expected: APIKeyFailureApplicationUnavailable},
		{name: "missing application", input: validAPIKeyCreateInput("app_bbbbbbbbbbbbbbbb", 30), expected: APIKeyFailureApplicationUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if result := service.Create(requestContext, test.input); result.FailureCode != test.expected || result.CredentialToken != "" {
				t.Fatalf("unexpected rejection: failure=%s record_present=%t credential_present=%t", result.FailureCode, result.Record != nil, result.CredentialToken != "")
			}
		})
	}
	if credentialCalls != 0 {
		t.Fatalf("invalid or inactive applications must be rejected before credential generation: %d", credentialCalls)
	}

	withoutCatalog := newAPIKeyService(repository, nil).Create(requestContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if withoutCatalog.FailureCode != APIKeyFailureCatalogRequired {
		t.Fatalf("missing application catalog must fail closed: failure=%s credential_present=%t", withoutCatalog.FailureCode, withoutCatalog.CredentialToken != "")
	}
	otherOwner := requestContext
	otherOwner.OwnerSubjectRef = "subject_other"
	otherOwner.ActorRef = "subject_other"
	if result := service.Create(otherOwner, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30)); result.FailureCode != APIKeyFailureApplicationUnavailable {
		t.Fatalf("cross-owner application must not be visible: failure=%s credential_present=%t", result.FailureCode, result.CredentialToken != "")
	}
}

func TestAPIKeyServiceExpiryFilteringPaginationAndOwnerIsolation(t *testing.T) {
	runAPIKeyServiceExpiryFilteringPaginationAndOwnerIsolation(t, newMemoryAPIKeyRepository(), newMemoryApplicationCatalogRepository())
}

func runAPIKeyServiceExpiryFilteringPaginationAndOwnerIsolation(t *testing.T, repository apiKeyRepository, applicationRepository applicationCatalogRepository) {
	t.Helper()
	owner := apiKeyTestContext("subject_owner")
	other := apiKeyTestContext("subject_other")
	seedAPIKeyTestApplication(t, applicationRepository, owner, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	seedAPIKeyTestApplication(t, applicationRepository, other, "app_bbbbbbbbbbbbbbbb", applicationCatalogLifecycleActive)
	now := time.Date(2026, 7, 13, 8, 0, 0, 0, time.UTC)
	service := newAPIKeyService(repository, applicationRepository)
	service.now = func() time.Time { return now }
	ids := []string{"key_aaaaaaaaaaaaaaaa", "key_bbbbbbbbbbbbbbbb", "key_cccccccccccccccc"}
	service.newID = func() (string, error) {
		id := ids[0]
		ids = ids[1:]
		return id, nil
	}
	first := service.Create(owner, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 1))
	now = now.Add(time.Second)
	second := service.Create(owner, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	now = now.Add(time.Second)
	third := service.Create(other, validAPIKeyCreateInput("app_bbbbbbbbbbbbbbbb", 30))
	if first.FailureCode != "" || second.FailureCode != "" || third.FailureCode != "" {
		t.Fatalf("seed API keys: first=%s second=%s third=%s", first.FailureCode, second.FailureCode, third.FailureCode)
	}

	now = time.Date(2026, 7, 15, 8, 0, 0, 0, time.UTC)
	expired := service.List(owner, APIKeyListInput{EffectiveState: apiKeyEffectiveExpired})
	if expired.FailureCode != "" || len(expired.Records) != 1 || expired.Records[0].APIKeyID != first.Record.APIKeyID {
		t.Fatalf("expired filter must use server time: %#v", expired)
	}
	pageOne := service.List(owner, APIKeyListInput{Limit: 1})
	if len(pageOne.Records) != 1 || pageOne.NextCursor == nil || pageOne.Records[0].APIKeyID != second.Record.APIKeyID {
		t.Fatalf("unexpected first page: %#v", pageOne)
	}
	pageTwo := service.List(owner, APIKeyListInput{Limit: 1, Cursor: *pageOne.NextCursor})
	if len(pageTwo.Records) != 1 || pageTwo.Records[0].APIKeyID != first.Record.APIKeyID {
		t.Fatalf("unexpected second page: %#v", pageTwo)
	}
	if otherRecords := service.List(other, APIKeyListInput{}); len(otherRecords.Records) != 1 || otherRecords.Records[0].APIKeyID != third.Record.APIKeyID {
		t.Fatalf("owner isolation failed: %#v", otherRecords)
	}
	if drift := service.List(owner, APIKeyListInput{ApplicationID: "app_aaaaaaaaaaaaaaaa", Limit: 1, Cursor: *pageOne.NextCursor}); drift.FailureCode != APIKeyFailureCursorInvalid {
		t.Fatalf("cursor must bind filters: %#v", drift)
	}
}

func TestAPIKeyMemoryRepositoryConcurrentRevokeHasSingleWinner(t *testing.T) {
	runAPIKeyConcurrentRevokeHasSingleWinner(t, newMemoryAPIKeyRepository(), newMemoryApplicationCatalogRepository())
}

func runAPIKeyConcurrentRevokeHasSingleWinner(t *testing.T, repository apiKeyRepository, applicationRepository applicationCatalogRepository) {
	t.Helper()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	service := newAPIKeyService(repository, applicationRepository)
	service.newID = func() (string, error) { return "key_aaaaaaaaaaaaaaaa", nil }
	issued := service.Create(requestContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30))
	if issued.FailureCode != "" {
		t.Fatalf("issue API key: failure=%s credential_present=%t", issued.FailureCode, issued.CredentialToken != "")
	}

	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result := service.Revoke(requestContext, issued.Record.APIKeyID, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case APIKeyFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected concurrent revoke result: %#v", result)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("expected one CAS winner: success=%d conflict=%d", successes.Load(), conflicts.Load())
	}
}

func TestAPIKeyServiceStoreUnavailableDoesNotFallback(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	repository := &memoryAPIKeyRepository{records: make(map[string]APIKeyRecord), unavailable: true}
	service := newAPIKeyService(repository, applicationRepository)
	if result := service.Create(requestContext, validAPIKeyCreateInput("app_aaaaaaaaaaaaaaaa", 30)); result.FailureCode != APIKeyFailureStoreUnavailable || result.CredentialToken != "" {
		t.Fatalf("create must fail without returning a credential: failure=%s credential_present=%t", result.FailureCode, result.CredentialToken != "")
	}
	if result := service.Read(requestContext, "key_aaaaaaaaaaaaaaaa"); result.FailureCode != APIKeyFailureStoreUnavailable {
		t.Fatalf("read must expose store failure: %#v", result)
	}
	if result := service.List(requestContext, APIKeyListInput{}); result.FailureCode != APIKeyFailureStoreUnavailable || len(result.Records) != 0 {
		t.Fatalf("list must expose store failure without fallback: %#v", result)
	}
}

func TestAPIKeyLifecycleHTTPCreateListReadRevokeAndNoLeakage(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	server := &Server{
		config: config.Config{
			ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true,
			APIKeyLifecycleDevHTTPEnabled: true, APIKeyLifecycleDevWriteEnabled: true,
		},
		applicationCatalogRepository: applicationRepository,
		apiKeyRepository:             newMemoryAPIKeyRepository(),
	}
	auth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "dev:test",
		TenantBinding: "tenant_demo", SubjectBinding: "subject_owner",
		ScopeGrants: []string{"api_keys:read", "api_keys:write", "api_keys:revoke"},
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{
  "workspace_id":"workspace_demo",
  "application_id":"app_aaaaaaaaaaaaaaaa",
  "display_name":"Primary SDK key",
  "scopes":["models:read","responses:invoke"],
  "expires_in_days":30
}`))
	createRequest = createRequest.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), auth))
	createRecorder := httptest.NewRecorder()
	server.handleCreateAPIKey(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated || createRecorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected create response: %d headers=%v body=%s", createRecorder.Code, createRecorder.Header(), createRecorder.Body.String())
	}
	var issued apiKeyEnvelope
	if err := json.Unmarshal(createRecorder.Body.Bytes(), &issued); err != nil || issued.Record == nil || issued.Credential == nil || issued.Credential.Token == "" {
		t.Fatalf("decode issue response: err=%v record_present=%t credential_present=%t", err, issued.Record != nil, issued.Credential != nil && issued.Credential.Token != "")
	}
	token := issued.Credential.Token
	if strings.Count(createRecorder.Body.String(), token) != 1 || strings.Contains(createRecorder.Body.String(), "credential_digest") {
		t.Fatalf("issue response must contain the token exactly once and no digest: %s", createRecorder.Body.String())
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/api-keys?workspace_id=workspace_demo&application_id=app_aaaaaaaaaaaaaaaa", nil)
	listRequest = listRequest.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), auth))
	listRecorder := httptest.NewRecorder()
	server.handleListAPIKeys(listRecorder, listRequest)
	if listRecorder.Code != http.StatusOK || strings.Contains(listRecorder.Body.String(), token) || strings.Contains(listRecorder.Body.String(), "credential") {
		t.Fatalf("list response leaked credential or failed: %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}

	readRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/api-keys/"+issued.Record.APIKeyID+"?workspace_id=workspace_demo", nil)
	readRequest.SetPathValue("api_key_id", issued.Record.APIKeyID)
	readRequest = readRequest.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), auth))
	readRecorder := httptest.NewRecorder()
	server.handleReadAPIKey(readRecorder, readRequest)
	if readRecorder.Code != http.StatusOK || strings.Contains(readRecorder.Body.String(), token) || strings.Contains(readRecorder.Body.String(), "credential") {
		t.Fatalf("read response leaked credential or failed: %d body=%s", readRecorder.Code, readRecorder.Body.String())
	}

	revokeRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys/"+issued.Record.APIKeyID+"/revoke", strings.NewReader(`{"workspace_id":"workspace_demo","expected_version":1}`))
	revokeRequest.SetPathValue("api_key_id", issued.Record.APIKeyID)
	revokeRequest = revokeRequest.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), auth))
	revokeRecorder := httptest.NewRecorder()
	server.handleRevokeAPIKey(revokeRecorder, revokeRequest)
	if revokeRecorder.Code != http.StatusOK || strings.Contains(revokeRecorder.Body.String(), token) || !strings.Contains(revokeRecorder.Body.String(), `"effective_state":"revoked"`) {
		t.Fatalf("unexpected revoke response: %d body=%s", revokeRecorder.Code, revokeRecorder.Body.String())
	}
}

func TestAPIKeyLifecycleHTTPPermissionsUnknownFieldsAndOIDCZeroQuery(t *testing.T) {
	applicationRepository := newMemoryApplicationCatalogRepository()
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, applicationRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)
	counting := &countingAPIKeyRepository{apiKeyRepository: newMemoryAPIKeyRepository()}
	server := &Server{
		config: config.Config{
			ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true,
			APIKeyLifecycleDevHTTPEnabled: true, APIKeyLifecycleDevWriteEnabled: true,
		},
		applicationCatalogRepository: applicationRepository,
		apiKeyRepository:             counting,
	}
	readOnlyAuth := controlPlaneReadAuthContext{
		AuthMode: controlPlaneReadAuthModeDevHeaders, IdentityContext: "dev:test",
		TenantBinding: "tenant_demo", SubjectBinding: "subject_owner", ScopeGrants: []string{"api_keys:read"},
	}
	denied := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_aaaaaaaaaaaaaaaa","display_name":"SDK key","scopes":["models:read"],"expires_in_days":30}`))
	denied = denied.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), readOnlyAuth))
	deniedRecorder := httptest.NewRecorder()
	server.handleCreateAPIKey(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden || !strings.Contains(deniedRecorder.Body.String(), APIKeyFailureScopeDenied) ||
		strings.Contains(deniedRecorder.Body.String(), "credential") || counting.createCalls.Load() != 0 {
		t.Fatalf("write scope must be independent and query-free: %d body=%s calls=%d", deniedRecorder.Code, deniedRecorder.Body.String(), counting.createCalls.Load())
	}

	unknown := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_aaaaaaaaaaaaaaaa","display_name":"SDK key","scopes":["models:read"],"expires_in_days":30,"token":"forbidden"}`))
	unknown = unknown.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), readOnlyAuth))
	unknownRecorder := httptest.NewRecorder()
	server.handleCreateAPIKey(unknownRecorder, unknown)
	if unknownRecorder.Code != http.StatusBadRequest || !strings.Contains(unknownRecorder.Body.String(), "INVALID_JSON") || counting.createCalls.Load() != 0 {
		t.Fatalf("unknown sensitive field must be rejected before repository: %d body=%s", unknownRecorder.Code, unknownRecorder.Body.String())
	}

	writeAuth := readOnlyAuth
	writeAuth.ScopeGrants = []string{"api_keys:write"}
	apiKeyCredential := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_aaaaaaaaaaaaaaaa","display_name":"SDK key","scopes":["models:read"],"expires_in_days":30}`))
	apiKeyCredential.Header.Set("Authorization", "Bearer rmd_dev_key_aaaaaaaaaaaaaaaa.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	apiKeyCredential = apiKeyCredential.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), writeAuth))
	apiKeyCredentialRecorder := httptest.NewRecorder()
	server.handleCreateAPIKey(apiKeyCredentialRecorder, apiKeyCredential)
	if apiKeyCredentialRecorder.Code != http.StatusBadRequest || !strings.Contains(apiKeyCredentialRecorder.Body.String(), APIKeyFailureCredentialConflict) || counting.createCalls.Load() != 0 {
		t.Fatalf("API key credential must not authorize its management route: status=%d body=%s calls=%d", apiKeyCredentialRecorder.Code, apiKeyCredentialRecorder.Body.String(), counting.createCalls.Load())
	}

	oidcAuth := readOnlyAuth
	oidcAuth.AuthMode = controlPlaneReadAuthModeRadishOIDCIntegrationTest
	list := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/api-keys?workspace_id=workspace_demo", nil)
	list = list.WithContext(withControlPlaneReadFakeAuthContext(context.Background(), oidcAuth))
	listRecorder := httptest.NewRecorder()
	server.handleListAPIKeys(listRecorder, list)
	if listRecorder.Code != http.StatusServiceUnavailable || !strings.Contains(listRecorder.Body.String(), "workspace_membership_unavailable") || counting.listCalls.Load() != 0 {
		t.Fatalf("OIDC membership gate must fail closed before repository: %d body=%s calls=%d", listRecorder.Code, listRecorder.Body.String(), counting.listCalls.Load())
	}

	disabled := &Server{config: config.Config{ApplicationCatalogDevHTTPEnabled: true}}
	disabledRequest := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{}`))
	disabledRecorder := httptest.NewRecorder()
	disabled.handleCreateAPIKey(disabledRecorder, disabledRequest)
	if disabledRecorder.Code != http.StatusForbidden || !strings.Contains(disabledRecorder.Body.String(), "API_KEY_LIFECYCLE_DEV_HTTP_DISABLED") {
		t.Fatalf("disabled lifecycle must fail closed: %d body=%s", disabledRecorder.Code, disabledRecorder.Body.String())
	}
}

func TestAPIKeyLifecycleRoutesAndPermissionProjection(t *testing.T) {
	if controlPlaneReadPermissionGrants["radishmind.api-keys.write"] != "api_keys:write" ||
		controlPlaneReadPermissionGrants["radishmind.api-keys.revoke"] != "api_keys:revoke" {
		t.Fatalf("API key management permissions are not projected independently: %#v", controlPlaneReadPermissionGrants)
	}
	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled:   true,
		ApplicationCatalogDevHTTPEnabled: true, ApplicationCatalogDevWriteEnabled: true,
		APIKeyLifecycleDevHTTPEnabled: true, APIKeyLifecycleDevWriteEnabled: true,
	}, Options{BuildVersion: "test"})
	t.Cleanup(server.Close)
	requestContext := apiKeyTestContext("subject_owner")
	seedAPIKeyTestApplication(t, server.applicationCatalogRepository, requestContext, "app_aaaaaaaaaaaaaaaa", applicationCatalogLifecycleActive)

	request := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/api-keys", strings.NewReader(`{"workspace_id":"workspace_demo","application_id":"app_aaaaaaaaaaaaaaaa","display_name":"SDK key","scopes":["models:read"],"expires_in_days":30}`))
	request.Header.Set(controlPlaneReadDevIdentityHeader, "dev:test")
	request.Header.Set(controlPlaneReadDevTenantHeader, "tenant_demo")
	request.Header.Set(controlPlaneReadDevSubjectHeader, "subject_owner")
	request.Header.Set(controlPlaneReadDevScopesHeader, "api_keys:write")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || !strings.Contains(recorder.Body.String(), `"credential":{"token":"`) {
		t.Fatalf("registered POST route did not issue a key: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

type countingAPIKeyRepository struct {
	apiKeyRepository
	createCalls atomic.Int32
	listCalls   atomic.Int32
}

func (repository *countingAPIKeyRepository) Create(requestContext APIKeyContext, record APIKeyRecord) (APIKeyRecord, error) {
	repository.createCalls.Add(1)
	return repository.apiKeyRepository.Create(requestContext, record)
}

func (repository *countingAPIKeyRepository) List(requestContext APIKeyContext, query apiKeyListQuery) ([]APIKeyRecord, error) {
	repository.listCalls.Add(1)
	return repository.apiKeyRepository.List(requestContext, query)
}

func apiKeyTestContext(owner string) APIKeyContext {
	return APIKeyContext{
		RequestContext: context.Background(), RequestID: "request_api_key_test", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ActorRef: owner, OwnerSubjectRef: owner,
		AuditRef: "audit_api_key_test", WriteEnabled: true,
	}
}

func validAPIKeyCreateInput(applicationID string, expiresInDays int) APIKeyCreateInput {
	return APIKeyCreateInput{
		ApplicationID: applicationID, DisplayName: "SDK key", Scopes: []string{"models:read"}, ExpiresInDays: expiresInDays,
	}
}

func seedAPIKeyTestApplication(t *testing.T, repository applicationCatalogRepository, requestContext APIKeyContext, applicationID, lifecycleState string) {
	t.Helper()
	catalogContext := ApplicationCatalogContext{
		RequestContext: requestContext.RequestContext, RequestID: requestContext.RequestID,
		TenantRef: requestContext.TenantRef, WorkspaceID: requestContext.WorkspaceID,
		ActorRef: requestContext.ActorRef, OwnerSubjectRef: requestContext.OwnerSubjectRef, AuditRef: requestContext.AuditRef,
	}
	createdAt := "2026-07-13T07:00:00Z"
	record := ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: applicationID,
		TenantRef: catalogContext.TenantRef, WorkspaceID: catalogContext.WorkspaceID, OwnerSubjectRef: catalogContext.OwnerSubjectRef,
		DisplayName: "API key test app", Description: "Active application for API key tests", ApplicationKind: "agent",
		LifecycleState: lifecycleState, RecordVersion: 1, CreatedAt: createdAt, UpdatedAt: createdAt,
		CreatedByActorRef: catalogContext.ActorRef, UpdatedByActorRef: catalogContext.ActorRef,
		RequestID: catalogContext.RequestID, AuditRef: catalogContext.AuditRef,
	}
	if lifecycleState == applicationCatalogLifecycleArchived {
		record.ArchivedAt = &createdAt
	}
	if _, err := repository.Create(catalogContext, record); err != nil {
		t.Fatalf("seed application catalog record: %v", err)
	}
}
