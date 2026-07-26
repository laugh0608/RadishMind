package httpapi

import (
	"context"
	"testing"
	"time"
)

func TestWorkspaceScopedReadRepositoryUsesDurableApplicationAndAPIKeyOwners(t *testing.T) {
	applications := newMemoryApplicationCatalogRepository()
	apiKeys := newMemoryAPIKeyRepository()
	owner := "subject_owner"
	appContext := applicationCatalogTestContext(owner)
	appRecord := ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: "app_aaaaaaaaaaaaaaaa",
		TenantRef: appContext.TenantRef, WorkspaceID: appContext.WorkspaceID, OwnerSubjectRef: owner,
		DisplayName: "Workspace application", ApplicationKind: "agent", LifecycleState: applicationCatalogLifecycleActive,
		RecordVersion: 1, CreatedAt: "2026-07-26T08:00:00Z", UpdatedAt: "2026-07-26T08:00:00Z",
		CreatedByActorRef: owner, UpdatedByActorRef: owner, RequestID: appContext.RequestID, AuditRef: appContext.AuditRef,
	}
	if _, err := applications.Create(appContext, appRecord); err != nil {
		t.Fatalf("seed application owner: %v", err)
	}
	keyContext := apiKeyTestContext(owner)
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)
	keyRecord := APIKeyRecord{
		SchemaVersion: apiKeyRecordSchemaVersion, APIKeyID: "key_aaaaaaaaaaaaaaaa",
		TenantRef: keyContext.TenantRef, WorkspaceID: keyContext.WorkspaceID,
		ApplicationID: appRecord.ApplicationID, OwnerSubjectRef: owner, DisplayName: "Read key",
		Scopes: []string{"models:read"}, LifecycleState: apiKeyLifecycleActive, EffectiveState: apiKeyLifecycleActive,
		RecordVersion: 1, CreatedAt: "2026-07-26T08:00:00Z", ExpiresAt: expiresAt,
		CreatedByActorRef: owner, RequestID: keyContext.RequestID, AuditRef: keyContext.AuditRef,
	}
	if _, err := apiKeys.Create(keyContext, keyRecord); err != nil {
		t.Fatalf("seed API key owner: %v", err)
	}
	repository := newWorkspaceScopedControlPlaneReadRepository(
		newControlPlaneReadRepository(newControlPlaneReadFakeStore()),
		applications, apiKeys, nil, nil,
	)
	readContext := ReadRepositoryContext{
		RequestContext: context.Background(), RequestID: "request_workspace_projection",
		TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", SubjectRef: owner,
		AuditRef: "audit_workspace_projection",
	}

	applicationResult := repository.ListApplicationSummaries(
		readContext,
		ListApplicationSummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{
			Filters: ReadRepositoryFilters{"application_kind": "agent"}, Sort: "updated_at_desc",
		}},
	)
	if applicationResult.FailureCode != "" || len(applicationResult.Items) != 1 ||
		applicationResult.Items[0].ApplicationRef != appRecord.ApplicationID {
		t.Fatalf("durable application projection mismatch: %#v", applicationResult)
	}

	apiKeyResult := repository.ListAPIKeySummaries(
		readContext,
		ListAPIKeySummariesRequest{ReadRepositoryRequest: ReadRepositoryRequest{
			Filters: ReadRepositoryFilters{"state": apiKeyLifecycleActive, "scope": "models:read"},
			Sort:    "created_at_desc",
		}},
	)
	if apiKeyResult.FailureCode != "" || len(apiKeyResult.Items) != 1 ||
		apiKeyResult.Items[0].APIKeyID != keyRecord.APIKeyID || apiKeyResult.Items[0].ExpiresAt == nil {
		t.Fatalf("durable API key projection mismatch: %#v", apiKeyResult)
	}

	otherWorkspace := readContext
	otherWorkspace.WorkspaceID = "workspace_other"
	if result := repository.ListApplicationSummaries(otherWorkspace, ListApplicationSummariesRequest{}); result.FailureCode != "" || len(result.Items) != 0 {
		t.Fatalf("cross-workspace application projection leaked records: %#v", result)
	}
	if result := repository.ListAPIKeySummaries(otherWorkspace, ListAPIKeySummariesRequest{}); result.FailureCode != "" || len(result.Items) != 0 {
		t.Fatalf("cross-workspace API key projection leaked records: %#v", result)
	}
}

func TestWorkspaceScopedRunProjectionDoesNotRequireApplicationSelection(t *testing.T) {
	store := &countingWorkspaceRunStore{}
	repository := newWorkspaceScopedControlPlaneReadRepository(
		newControlPlaneReadRepository(newControlPlaneReadFakeStore()),
		newMemoryApplicationCatalogRepository(), newMemoryAPIKeyRepository(), nil, store,
	)
	result := repository.ListRunRecordSummaries(
		ReadRepositoryContext{
			RequestContext: context.Background(), TenantRef: "tenant_demo",
			WorkspaceID: "workspace_demo", SubjectRef: "subject_owner", AuditRef: "audit_run_projection",
		},
		ListRunRecordSummariesRequest{},
	)
	if result.FailureCode != "" || len(result.Items) != 0 {
		t.Fatalf("unexpected workspace-wide run projection result: %#v", result)
	}
	if store.workspaceListCalls != 1 || store.applicationID != "" {
		t.Fatalf("workspace projection was not queried without an application filter: %#v", store)
	}
}

type countingWorkspaceRunStore struct {
	listCalls          int
	workspaceListCalls int
	applicationID      string
}

func (store *countingWorkspaceRunStore) UpsertRun(WorkflowRunContext, *WorkflowRunRecord) error {
	return nil
}

func (store *countingWorkspaceRunStore) ReadRun(WorkflowRunContext, string) (WorkflowRunRecord, bool, error) {
	return WorkflowRunRecord{}, false, nil
}

func (store *countingWorkspaceRunStore) ListRuns(WorkflowRunContext, WorkflowRunListFilter) (WorkflowRunListPage, error) {
	store.listCalls++
	return WorkflowRunListPage{Records: []WorkflowRunRecord{}}, nil
}

func (store *countingWorkspaceRunStore) ListWorkspaceRuns(
	_ WorkflowWorkspaceRunListContext,
	filter WorkflowWorkspaceRunListFilter,
) (WorkflowRunListPage, error) {
	store.workspaceListCalls++
	store.applicationID = filter.ApplicationID
	return WorkflowRunListPage{Records: []WorkflowRunRecord{}}, nil
}
