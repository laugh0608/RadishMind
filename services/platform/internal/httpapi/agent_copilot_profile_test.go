package httpapi

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAgentCopilotProfileOwnerLifecycleAndCAS(t *testing.T) {
	repository := newMemoryAgentCopilotProfileRepository()
	service := newAgentCopilotProfileService(repository)
	nextTime := time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time {
		value := nextTime
		nextTime = nextTime.Add(time.Second)
		return value
	}
	catalogReads := 0
	service.requireAgentApplication = func(AgentCopilotProfileContext) string {
		catalogReads++
		return ""
	}
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")

	validated := service.Validate(ctx, input)
	if !validated.ValidationSummary.IsValid || len(repository.drafts) != 0 || catalogReads != 0 {
		t.Fatalf("validate must be valid and side-effect free: %#v drafts=%d catalog_reads=%d", validated, len(repository.drafts), catalogReads)
	}

	disabled := ctx
	disabled.WriteEnabled = false
	if result := service.SaveDraft(disabled, input, 0); result.FailureCode != AgentCopilotProfileFailureWriteDisabled || len(repository.drafts) != 0 {
		t.Fatalf("write-disabled save mutated owner: %#v", result)
	}

	created := service.SaveDraft(ctx, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 || created.Draft.AllowedTasks[0] != "explain_diagnostics" {
		t.Fatalf("create draft failed: %#v", created)
	}
	if catalogReads != 1 {
		t.Fatalf("save must reread application catalog once, got %d", catalogReads)
	}
	stale := service.SaveDraft(ctx, input, 0)
	if stale.FailureCode != AgentCopilotProfileFailureVersionConflict || stale.CurrentDraftVersion != 1 {
		t.Fatalf("stale CAS did not fail closed: %#v", stale)
	}

	input.Description = "Updated controlled profile."
	updated := service.SaveDraft(ctx, input, 1)
	if updated.FailureCode != "" || updated.Draft == nil || updated.Draft.DraftVersion != 2 ||
		updated.Draft.CreatedAt != created.Draft.CreatedAt || updated.Draft.UpdatedAt == created.Draft.UpdatedAt {
		t.Fatalf("draft update did not preserve creation metadata: %#v", updated)
	}

	version := service.CreateVersion(ctx, input.ProfileID, 2)
	if version.FailureCode != "" || version.Version == nil || version.Version.ProfileVersion != 1 ||
		version.Version.SourceDraftVersion != 2 || version.Version.ProfileDigest != updated.Draft.ProfileDigest {
		t.Fatalf("version creation failed: %#v", version)
	}
	if catalogReads != 4 {
		t.Fatalf("save/version must reread catalog; got %d reads", catalogReads)
	}
	if duplicate := service.CreateVersion(ctx, input.ProfileID, 2); duplicate.FailureCode != AgentCopilotProfileFailureImmutableConflict {
		t.Fatalf("same draft produced a second immutable version: %#v", duplicate)
	}
	if staleVersion := service.CreateVersion(ctx, input.ProfileID, 1); staleVersion.FailureCode != AgentCopilotProfileFailureVersionConflict || staleVersion.CurrentDraftVersion != 2 {
		t.Fatalf("stale source draft was accepted: %#v", staleVersion)
	}

	readDraft := service.ReadDraft(ctx, input.ProfileID)
	readVersion := service.ReadVersion(ctx, input.ProfileID, 1)
	if readDraft.Draft == nil || readVersion.Version == nil {
		t.Fatalf("saved sources are unreadable: draft=%#v version=%#v", readDraft, readVersion)
	}
	readDraft.Draft.AllowedTasks[0] = "mutated"
	readVersion.Version.AllowedTasks[0] = "mutated"
	if again := service.ReadDraft(ctx, input.ProfileID); again.Draft == nil || again.Draft.AllowedTasks[0] == "mutated" {
		t.Fatal("repository leaked mutable draft state")
	}
	if again := service.ReadVersion(ctx, input.ProfileID, 1); again.Version == nil || again.Version.AllowedTasks[0] == "mutated" {
		t.Fatal("repository leaked mutable version state")
	}
}

func TestAgentCopilotProfileOwnerScopeStableListsAndApplicationGate(t *testing.T) {
	repository := newMemoryAgentCopilotProfileRepository()
	service := newAgentCopilotProfileService(repository)
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	fixed := time.Date(2026, 7, 25, 11, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixed }
	service.requireAgentApplication = func(AgentCopilotProfileContext) string { return "" }

	for _, profileID := range []string{"acpf_bbbbbbbbbbbbbbbb", "acpf_aaaaaaaaaaaaaaaa"} {
		if result := service.SaveDraft(ctx, agentCopilotProfileTestInput(profileID), 0); result.FailureCode != "" {
			t.Fatalf("save %s: %#v", profileID, result)
		}
	}
	summaries, failure := service.ListDrafts(ctx)
	if failure != "" || len(summaries) != 2 || summaries[0].ProfileID != "acpf_aaaaaaaaaaaaaaaa" ||
		summaries[0].AllowedTasksDigest == "" || summaries[0].AllowedTasks == nil {
		t.Fatalf("draft list is not stable or complete: %#v failure=%s", summaries, failure)
	}

	otherOwner := agentCopilotProfileTestContext(ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, "subject:other")
	if result := service.ReadDraft(otherOwner, "acpf_aaaaaaaaaaaaaaaa"); result.FailureCode != AgentCopilotProfileFailureNotFound {
		t.Fatalf("cross-owner read leaked a profile: %#v", result)
	}
	otherWorkspace := agentCopilotProfileTestContext(ctx.TenantRef, "workspace_two", ctx.ApplicationID, ctx.OwnerSubjectRef)
	if summaries, failure := service.ListDrafts(otherWorkspace); failure != "" || len(summaries) != 0 {
		t.Fatalf("cross-workspace list leaked profiles: %#v failure=%s", summaries, failure)
	}

	service.requireAgentApplication = func(AgentCopilotProfileContext) string {
		return AgentCopilotProfileFailureApplicationKind
	}
	if result := service.SaveDraft(ctx, agentCopilotProfileTestInput("acpf_cccccccccccccccc"), 0); result.FailureCode != AgentCopilotProfileFailureApplicationKind {
		t.Fatalf("non-agent application was accepted: %#v", result)
	}
}

func TestAgentCopilotProfileOwnerCorruptionAndUnavailableFailClosed(t *testing.T) {
	repository := newMemoryAgentCopilotProfileRepository()
	service := newAgentCopilotProfileService(repository)
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	service.requireAgentApplication = func(AgentCopilotProfileContext) string { return "" }
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	if result := service.SaveDraft(ctx, input, 0); result.FailureCode != "" {
		t.Fatalf("save fixture: %#v", result)
	}
	if result := service.CreateVersion(ctx, input.ProfileID, 1); result.FailureCode != "" {
		t.Fatalf("version fixture: %#v", result)
	}

	key := agentCopilotProfileRepositoryKey(ctx, input.ProfileID)
	draft := repository.drafts[key]
	draft.ProfileDigest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	repository.drafts[key] = draft
	if result := service.ReadDraft(ctx, input.ProfileID); result.FailureCode != AgentCopilotProfileFailureDigestDrift {
		t.Fatalf("draft digest corruption did not fail closed: %#v", result)
	}
	if _, failure := service.ListDrafts(ctx); failure != AgentCopilotProfileFailureDigestDrift {
		t.Fatalf("list ignored corrupt draft: %s", failure)
	}

	compiled, findings := CompileAgentCopilotProfileSource(draft.AgentCopilotProfileSource)
	if len(findings) != 0 {
		t.Fatalf("recompile fixture: %#v", findings)
	}
	draft.ProfileDigest = compiled.ProfileDigest
	repository.drafts[key] = draft
	version := repository.versions[key][1]
	version.PolicyDigest = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	repository.versions[key][1] = version
	if result := service.ReadVersion(ctx, input.ProfileID, 1); result.FailureCode != AgentCopilotProfileFailureDigestDrift {
		t.Fatalf("version digest corruption did not fail closed: %#v", result)
	}

	repository.unavailable = true
	if result := service.ReadDraft(ctx, input.ProfileID); result.FailureCode != AgentCopilotProfileFailureStoreUnavailable {
		t.Fatalf("unavailable read did not preserve stable failure: %#v", result)
	}
	if result := service.SaveDraft(ctx, input, 1); result.FailureCode != AgentCopilotProfileFailureStoreUnavailable {
		t.Fatalf("unavailable save did not preserve stable failure: %#v", result)
	}
}

func TestAgentCopilotProfileOwnerConcurrentCAS(t *testing.T) {
	repository := newMemoryAgentCopilotProfileRepository()
	service := newAgentCopilotProfileService(repository)
	ctx := agentCopilotProfileTestContext("tenant:one", "workspace_one", "app_aaaaaaaaaaaaaaaa", "subject:owner")
	service.requireAgentApplication = func(AgentCopilotProfileContext) string { return "" }
	input := agentCopilotProfileTestInput("acpf_aaaaaaaaaaaaaaaa")
	if result := service.SaveDraft(ctx, input, 0); result.FailureCode != "" {
		t.Fatalf("save fixture: %#v", result)
	}

	const writers = 12
	var wait sync.WaitGroup
	wait.Add(writers)
	results := make(chan AgentCopilotProfileResult, writers)
	for index := 0; index < writers; index++ {
		go func() {
			defer wait.Done()
			results <- service.SaveDraft(ctx, input, 1)
		}()
	}
	wait.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		switch result.FailureCode {
		case "":
			successes++
		case AgentCopilotProfileFailureVersionConflict:
			conflicts++
		default:
			t.Fatalf("unexpected CAS result: %#v", result)
		}
	}
	if successes != 1 || conflicts != writers-1 {
		t.Fatalf("CAS accepted %d successes and %d conflicts", successes, conflicts)
	}
}

func agentCopilotProfileTestContext(tenantRef, workspaceID, applicationID, ownerRef string) AgentCopilotProfileContext {
	return AgentCopilotProfileContext{
		RequestContext: context.Background(), RequestID: "request:profile", TenantRef: tenantRef,
		WorkspaceID: workspaceID, ApplicationID: applicationID, ActorRef: ownerRef,
		OwnerSubjectRef: ownerRef, AuditRef: "audit:profile", WriteEnabled: true,
	}
}

func agentCopilotProfileTestInput(profileID string) AgentCopilotProfileDraftInput {
	return AgentCopilotProfileDraftInput{
		SchemaVersion: agentCopilotProfileDraftSchema, ProfileID: profileID,
		WorkspaceID: "workspace_one", ApplicationID: "app_aaaaaaaaaaaaaaaa",
		AgentCopilotProfileSource: validAgentCopilotProfileSourceFixture(),
	}
}
