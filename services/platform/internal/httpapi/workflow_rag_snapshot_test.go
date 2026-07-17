package httpapi

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/sqlitedev"
	sqliteworkflowrunmigrations "radishmind.local/services/platform/migrations/sqlite/workflow_runs"
)

func TestWorkflowRAGSnapshotMemoryLifecycle(t *testing.T) {
	repository := newMemoryWorkflowRAGSnapshotRepository(nil)
	runWorkflowRAGSnapshotLifecycle(t, repository)
	entry := repository.entries[workflowRAGStoreKey(workflowRAGTestContext(), "rags_aaaaaaaaaaaaaaaa")]
	auditPayload, err := json.Marshal(entry.audits)
	if err != nil {
		t.Fatalf("marshal memory audits: %v", err)
	}
	if strings.Contains(string(auditPayload), "official retrieval guidance") || strings.Contains(string(auditPayload), "internal operator notes") {
		t.Fatalf("snapshot audit leaked fragment content: %s", auditPayload)
	}
}

func TestWorkflowRAGSnapshotSQLiteLifecycleAndRestart(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "workflow-rag.db")
	runtime, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("open workflow RAG SQLite runtime: %v", err)
	}
	repository := newSQLiteWorkflowRAGSnapshotRepository(runtime.DB())
	runWorkflowRAGSnapshotLifecycle(t, repository)
	if err = runtime.Close(); err != nil {
		t.Fatalf("close workflow RAG SQLite runtime: %v", err)
	}

	restarted, err := sqlitedev.Open(context.Background(), sqlitedev.Options{
		DatabasePath: databasePath,
		Migrations:   sqliteworkflowrunmigrations.Migrations(),
	})
	if err != nil {
		t.Fatalf("restart workflow RAG SQLite runtime: %v", err)
	}
	t.Cleanup(func() { _ = restarted.Close() })
	restored := newWorkflowRAGSnapshotService(newSQLiteWorkflowRAGSnapshotRepository(restarted.DB())).Read(
		workflowRAGTestContext(), "rags_aaaaaaaaaaaaaaaa", 2,
	)
	if restored.FailureCode != "" || restored.Record == nil || restored.Record.SnapshotVersion != 2 ||
		restored.Record.LifecycleState != workflowRAGSnapshotArchived || len(restored.Record.Fragments) != 1 ||
		restored.Record.Fragments[0].Content != "version two replacement content" {
		t.Fatalf("SQLite restart did not restore exact immutable snapshot version: %#v", restored)
	}
	var versionCount, fragmentCount, auditCount int
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_snapshot_versions`).Scan(&versionCount); err != nil || versionCount != 2 {
		t.Fatalf("unexpected SQLite snapshot version count: count=%d err=%v", versionCount, err)
	}
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_snapshot_fragments`).Scan(&fragmentCount); err != nil || fragmentCount != 3 {
		t.Fatalf("unexpected SQLite snapshot fragment count: count=%d err=%v", fragmentCount, err)
	}
	if err = restarted.DB().QueryRow(`SELECT count(*) FROM workflow_rag_execution_audits`).Scan(&auditCount); err != nil || auditCount != 3 {
		t.Fatalf("unexpected SQLite snapshot audit count: count=%d err=%v", auditCount, err)
	}
	if _, err = restarted.DB().Exec(`UPDATE workflow_rag_snapshot_versions SET snapshot_version=3 WHERE snapshot_version=2`); err == nil {
		t.Fatal("SQLite accepted mutation of immutable snapshot version")
	}
	if _, err = restarted.DB().Exec(`DELETE FROM workflow_rag_snapshot_fragments`); err == nil {
		t.Fatal("SQLite accepted deletion of immutable snapshot fragments")
	}
	if _, err = restarted.DB().Exec(`UPDATE workflow_rag_execution_audits SET event_kind='snapshot_created'`); err == nil {
		t.Fatal("SQLite accepted mutation of append-only snapshot audit")
	}
}

func TestWorkflowRAGSnapshotValidationAndDigest(t *testing.T) {
	ctx := workflowRAGTestContext()
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	build := func(inputs []WorkflowRAGFragmentInput) WorkflowRAGSnapshotRecord {
		service := newWorkflowRAGSnapshotService(newMemoryWorkflowRAGSnapshotRepository(nil))
		service.now = func() time.Time { return now }
		service.newID = func(string) (string, error) { return "rags_bbbbbbbbbbbbbbbb", nil }
		result := service.Create(ctx, WorkflowRAGSnapshotCreateInput{
			SnapshotKey: "operator_manual", DisplayName: "Operator manual", ContentClassification: "workspace_internal", Fragments: inputs,
		})
		if result.FailureCode != "" || result.Record == nil {
			t.Fatalf("build deterministic snapshot: %#v", result)
		}
		return *result.Record
	}
	first := workflowRAGTestFragments()
	second := []WorkflowRAGFragmentInput{first[1], first[0]}
	if left, right := build(first), build(second); left.SnapshotDigest != right.SnapshotDigest {
		t.Fatalf("snapshot digest depends on client fragment order: %s != %s", left.SnapshotDigest, right.SnapshotDigest)
	}

	service := newWorkflowRAGSnapshotService(newMemoryWorkflowRAGSnapshotRepository(nil))
	service.newID = func(string) (string, error) { return "rags_cccccccccccccccc", nil }
	secret := service.Create(ctx, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "secret_manual", DisplayName: "Secret manual", ContentClassification: "workspace_internal",
		Fragments: []WorkflowRAGFragmentInput{{
			FragmentRef: "secret_fragment", SourceType: "manual", SourceRef: "manual.secret",
			PageSlug: "manual/secret", Content: "Authorization: Bearer should-not-be-stored",
		}},
	})
	if secret.FailureCode != WorkflowRAGFailureSecretForbidden {
		t.Fatalf("secret-bearing fragment was not rejected: %#v", secret)
	}
	oversized := strings.Repeat("x", workflowRAGMaxFragmentBytes+1)
	budget := service.Create(ctx, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "large_manual", DisplayName: "Large manual", ContentClassification: "public",
		Fragments: []WorkflowRAGFragmentInput{{FragmentRef: "large_fragment", SourceType: "manual", SourceRef: "manual.large", PageSlug: "large", Content: oversized}},
	})
	if budget.FailureCode != WorkflowRAGFailureBudgetExceeded {
		t.Fatalf("oversized fragment was not rejected: %#v", budget)
	}
}

func runWorkflowRAGSnapshotLifecycle(t *testing.T, repository workflowRAGSnapshotRepository) {
	t.Helper()
	ctx := workflowRAGTestContext()
	service := newWorkflowRAGSnapshotService(repository)
	now := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.newID = func(prefix string) (string, error) {
		if prefix == "rags_" {
			return "rags_aaaaaaaaaaaaaaaa", nil
		}
		return newWorkflowRAGStableID(prefix)
	}
	created := service.Create(ctx, WorkflowRAGSnapshotCreateInput{
		SnapshotKey: "operator_manual", DisplayName: "Operator manual",
		ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if created.FailureCode != "" || created.Record == nil || created.Record.SnapshotVersion != 1 ||
		created.Record.RAGRef != "workflow.rag.operator_manual.v1" || created.Record.FragmentCount != 2 ||
		!workflowRAGDigestPattern.MatchString(created.Record.SnapshotDigest) {
		t.Fatalf("create workflow RAG snapshot: %#v", created)
	}
	listed := service.List(ctx, WorkflowRAGSnapshotListInput{})
	if listed.FailureCode != "" || len(listed.Records) != 1 || listed.Records[0].SnapshotKey != "operator_manual" {
		t.Fatalf("list workflow RAG snapshot metadata: %#v", listed)
	}
	listJSON, _ := json.Marshal(listed)
	if strings.Contains(string(listJSON), "official retrieval guidance") {
		t.Fatalf("snapshot list returned fragment content: %s", listJSON)
	}
	readV1 := service.Read(ctx, created.Record.SnapshotID, 1)
	if readV1.FailureCode != "" || readV1.Record == nil || len(readV1.Record.Fragments) != 2 {
		t.Fatalf("read exact workflow RAG snapshot version: %#v", readV1)
	}

	now = now.Add(time.Minute)
	ctx.RequestID = "request_rag_version"
	ctx.AuditRef = "audit_rag_version"
	versioned := service.Version(ctx, created.Record.SnapshotID, WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: 1, DisplayName: "Operator manual v2", ContentClassification: "workspace_internal",
		Fragments: []WorkflowRAGFragmentInput{{
			FragmentRef: "replacement", SourceType: "manual", SourceRef: "manual.operator.v2",
			PageSlug: "operator/v2", Title: "Version two", IsOfficial: true, Content: "version two replacement content",
		}},
	})
	if versioned.FailureCode != "" || versioned.Record == nil || versioned.Record.SnapshotVersion != 2 ||
		versioned.Record.RAGRef != "workflow.rag.operator_manual.v2" {
		t.Fatalf("create immutable workflow RAG snapshot version: %#v", versioned)
	}
	stale := service.Version(ctx, created.Record.SnapshotID, WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: 1, DisplayName: "Stale", ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if stale.FailureCode != WorkflowRAGFailureVersionConflict || stale.CurrentLatestVersion != 2 {
		t.Fatalf("stale workflow RAG version CAS did not fail closed: %#v", stale)
	}
	future := service.Version(ctx, created.Record.SnapshotID, WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: 3, DisplayName: "Future", ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if future.FailureCode != WorkflowRAGFailureVersionConflict || future.CurrentLatestVersion != 2 {
		t.Fatalf("future workflow RAG version CAS did not report the authoritative latest version: %#v", future)
	}
	oldVersion := service.Read(ctx, created.Record.SnapshotID, 1)
	if oldVersion.FailureCode != "" || oldVersion.Record == nil || oldVersion.Record.Fragments[0].Content == "version two replacement content" {
		t.Fatalf("old workflow RAG snapshot version changed: %#v", oldVersion)
	}

	otherScope := ctx
	otherScope.ApplicationID = "app_other"
	if denied := service.Read(otherScope, created.Record.SnapshotID, 1); denied.FailureCode != WorkflowRAGFailureScopeDenied {
		t.Fatalf("cross-application snapshot read did not fail closed: %#v", denied)
	}
	now = now.Add(time.Minute)
	ctx.RequestID = "request_rag_archive"
	ctx.AuditRef = "audit_rag_archive"
	archived := service.Archive(ctx, created.Record.SnapshotID, 2)
	if archived.FailureCode != "" || archived.Record == nil || archived.Record.LifecycleState != workflowRAGSnapshotArchived {
		t.Fatalf("archive workflow RAG snapshot: %#v", archived)
	}
	blocked := service.Version(ctx, created.Record.SnapshotID, WorkflowRAGSnapshotVersionInput{
		ExpectedLatestVersion: 2, DisplayName: "Blocked", ContentClassification: "workspace_internal", Fragments: workflowRAGTestFragments(),
	})
	if blocked.FailureCode != WorkflowRAGFailureArchived {
		t.Fatalf("archived snapshot accepted a new version: %#v", blocked)
	}
}

func workflowRAGTestContext() WorkflowRAGSnapshotContext {
	return WorkflowRAGSnapshotContext{
		RequestContext: context.Background(), RequestID: "request_rag_test", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_flow_copilot", ActorRef: "subject_owner", AuditRef: "audit_rag_test",
	}
}

func workflowRAGTestFragments() []WorkflowRAGFragmentInput {
	return []WorkflowRAGFragmentInput{
		{FragmentRef: "official_guide", SourceType: "manual", SourceRef: "manual.operator", PageSlug: "operator/guide", Title: "Official guide", IsOfficial: true, Content: "official retrieval guidance"},
		{FragmentRef: "internal_notes", SourceType: "wiki", SourceRef: "wiki.operator", PageSlug: "operator/notes", Title: "Operator notes", Content: "internal operator notes"},
	}
}
