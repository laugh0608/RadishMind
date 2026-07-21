package httpapi

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPromptApplicationTemplateOwnerCASNormalizationAndImmutableVersion(t *testing.T) {
	repository := newMemoryPromptApplicationTemplateRepository()
	service := newPromptApplicationTemplateService(repository)
	service.now = func() time.Time { return time.Date(2026, 7, 21, 8, 0, 0, 0, time.UTC) }
	ctx := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	input.TemplateName = "  Support summary  "
	input.Variables = []PromptApplicationTemplateVariable{
		{Name: "tone", Type: promptApplicationVariableString, Description: "  response tone  ", DefaultValue: "clear"},
		{Name: "question", Type: promptApplicationVariableString, Required: true, Description: "  user question  "},
	}

	created := service.SaveDraft(ctx, input, 0)
	if created.FailureCode != "" || created.Draft == nil || created.Draft.DraftVersion != 1 {
		t.Fatalf("valid draft create failed: %#v", created)
	}
	if created.Draft.TemplateName != "Support summary" || created.Draft.Variables[0].Name != "question" || created.Draft.Variables[1].Description != "response tone" || !workflowRAGDigestPattern.MatchString(created.Draft.TemplateDigest) {
		t.Fatalf("draft was not normalized deterministically: %#v", created.Draft)
	}
	conflict := service.SaveDraft(ctx, input, 0)
	if conflict.FailureCode != PromptApplicationTemplateFailureVersionConflict || conflict.CurrentDraftVersion != 1 {
		t.Fatalf("stale draft CAS did not fail closed: %#v", conflict)
	}

	version := service.CreateVersion(ctx, input.TemplateID, 1)
	if version.FailureCode != "" || version.Version == nil || version.Version.TemplateVersion != 1 || version.Version.TemplateDigest != created.Draft.TemplateDigest {
		t.Fatalf("immutable version create failed: %#v", version)
	}
	duplicate := service.CreateVersion(ctx, input.TemplateID, 1)
	if duplicate.FailureCode != PromptApplicationTemplateFailureImmutableConflict {
		t.Fatalf("source draft version was published twice: %#v", duplicate)
	}

	input.Description = "second draft"
	updated := service.SaveDraft(ctx, input, 1)
	if updated.FailureCode != "" || updated.Draft == nil || updated.Draft.DraftVersion != 2 || updated.Draft.CreatedAt != created.Draft.CreatedAt {
		t.Fatalf("draft update did not preserve owner history: %#v", updated)
	}
	oldVersion := service.ReadVersion(ctx, input.TemplateID, 1)
	if oldVersion.FailureCode != "" || oldVersion.Version == nil || oldVersion.Version.Description != created.Draft.Description {
		t.Fatalf("immutable version changed with the draft: %#v", oldVersion)
	}
	staleVersion := service.CreateVersion(ctx, input.TemplateID, 1)
	if staleVersion.FailureCode != PromptApplicationTemplateFailureVersionConflict || staleVersion.CurrentDraftVersion != 2 {
		t.Fatalf("version create did not re-read exact draft: %#v", staleVersion)
	}
}

func TestPromptApplicationTemplateOwnerIsolationNoFallbackAndDigestDrift(t *testing.T) {
	repository := newMemoryPromptApplicationTemplateRepository()
	service := newPromptApplicationTemplateService(repository)
	ctx := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	created := service.SaveDraft(ctx, input, 0)
	if created.FailureCode != "" || created.Draft == nil {
		t.Fatalf("seed draft: %#v", created)
	}

	otherOwner := ctx
	otherOwner.ActorRef = "subject_other"
	otherOwner.OwnerSubjectRef = "subject_other"
	if read := service.ReadDraft(otherOwner, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureNotFound {
		t.Fatalf("owner scope leaked a draft: %#v", read)
	}
	otherWorkspace := ctx
	otherWorkspace.WorkspaceID = "workspace_other"
	if summaries, failure := service.ListDrafts(otherWorkspace); failure != "" || len(summaries) != 0 {
		t.Fatalf("workspace scope leaked summaries: failure=%q summaries=%#v", failure, summaries)
	}

	repository.mu.Lock()
	corrupted := repository.drafts[promptApplicationTemplateRepositoryKey(ctx, input.TemplateID)]
	corrupted.TemplateDigest = "sha256:" + string(make([]byte, 64))
	repository.drafts[promptApplicationTemplateRepositoryKey(ctx, input.TemplateID)] = corrupted
	repository.mu.Unlock()
	if read := service.ReadDraft(ctx, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureStoreContract {
		t.Fatalf("stored digest drift was accepted: %#v", read)
	}

	repository.mu.Lock()
	repository.unavailable = true
	repository.mu.Unlock()
	if read := service.ReadDraft(ctx, input.TemplateID); read.FailureCode != PromptApplicationTemplateFailureStoreUnavailable || read.Draft != nil {
		t.Fatalf("unavailable owner fell back to another state: %#v", read)
	}
}

func TestPromptApplicationTemplateOwnerConcurrentCASSelectsOneWriter(t *testing.T) {
	repository := newMemoryPromptApplicationTemplateRepository()
	service := newPromptApplicationTemplateService(repository)
	ctx := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	if seed := service.SaveDraft(ctx, input, 0); seed.FailureCode != "" {
		t.Fatalf("seed draft: %#v", seed)
	}

	var successes atomic.Int32
	var conflicts atomic.Int32
	var wait sync.WaitGroup
	for index := 0; index < 8; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			candidate := input
			candidate.Description = "concurrent update"
			result := service.SaveDraft(ctx, candidate, 1)
			switch result.FailureCode {
			case "":
				successes.Add(1)
			case PromptApplicationTemplateFailureVersionConflict:
				conflicts.Add(1)
			default:
				t.Errorf("unexpected concurrent failure: %#v", result)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || conflicts.Load() != 7 {
		t.Fatalf("CAS must select one writer: successes=%d conflicts=%d", successes.Load(), conflicts.Load())
	}
}

func TestPromptApplicationTemplateContractCodecRejectsUnknownAndCredentialFields(t *testing.T) {
	service := newPromptApplicationTemplateService(newMemoryPromptApplicationTemplateRepository())
	ctx := validPromptApplicationTemplateContext()
	created := service.SaveDraft(ctx, validPromptApplicationTemplateDraftInput(), 0)
	if created.FailureCode != "" || created.Draft == nil {
		t.Fatalf("seed draft: %#v", created)
	}
	payload, err := json.Marshal(created.Draft)
	if err != nil || validatePromptApplicationTemplateContractJSON(promptApplicationTemplateDraftSchemaVersion, payload) != nil {
		t.Fatalf("valid draft contract rejected: err=%v payload=%s", err, payload)
	}
	var unknown map[string]any
	if err := json.Unmarshal(payload, &unknown); err != nil {
		t.Fatalf("decode draft fixture: %v", err)
	}
	unknown["provider_api_key"] = "forbidden"
	unknownPayload, _ := json.Marshal(unknown)
	if validatePromptApplicationTemplateContractJSON(promptApplicationTemplateDraftSchemaVersion, unknownPayload) == nil {
		t.Fatal("unknown credential-bearing field was accepted")
	}

	version := service.CreateVersion(ctx, created.Draft.TemplateID, 1)
	versionPayload, err := json.Marshal(version.Version)
	if err != nil || validatePromptApplicationTemplateContractJSON(promptApplicationTemplateVersionSchemaVersion, versionPayload) != nil {
		t.Fatalf("valid version contract rejected: err=%v payload=%s", err, versionPayload)
	}
}

func TestPromptApplicationTemplateSaveRequiresEligibleApplicationBeforeWrite(t *testing.T) {
	repository := newMemoryPromptApplicationTemplateRepository()
	service := newPromptApplicationTemplateService(repository)
	var checks atomic.Int32
	service.requirePromptApplication = func(PromptApplicationTemplateContext) string {
		checks.Add(1)
		return PromptApplicationTemplateFailureApplicationKind
	}
	ctx := validPromptApplicationTemplateContext()
	input := validPromptApplicationTemplateDraftInput()
	result := service.SaveDraft(ctx, input, 0)
	if result.FailureCode != PromptApplicationTemplateFailureApplicationKind || checks.Load() != 1 {
		t.Fatalf("application eligibility did not fail closed: %#v checks=%d", result, checks.Load())
	}
	if _, err := repository.ReadDraft(ctx, input.TemplateID); err == nil {
		t.Fatal("ineligible application wrote a template draft")
	}
}

func validPromptApplicationTemplateContext() PromptApplicationTemplateContext {
	return PromptApplicationTemplateContext{
		RequestContext: context.Background(), RequestID: "request_prompt_template", TenantRef: "tenant_demo",
		WorkspaceID: "workspace_demo", ApplicationID: "app_aaaaaaaaaaaaaaaa", ActorRef: "subject_owner",
		OwnerSubjectRef: "subject_owner", AuditRef: "audit_prompt_template", WriteEnabled: true,
	}
}

func validPromptApplicationTemplateDraftInput() PromptApplicationTemplateDraftInput {
	return PromptApplicationTemplateDraftInput{
		SchemaVersion: promptApplicationTemplateDraftSchemaVersion, TemplateID: "ptpl_aaaaaaaaaaaaaaaa",
		WorkspaceID: "workspace_demo", ApplicationID: "app_aaaaaaaaaaaaaaaa", TemplateName: "Support summary",
		Description: "Summarize a support question", PromptApplicationTemplateSource: PromptApplicationTemplateSource{
			Messages: []PromptApplicationTemplateMessage{
				{Role: "system", Content: "Use a {{ tone }} tone."},
				{Role: "user", Content: "Question: {{ question }}"},
			},
			Variables: []PromptApplicationTemplateVariable{
				{Name: "question", Type: promptApplicationVariableString, Required: true, Description: "User question"},
				{Name: "tone", Type: promptApplicationVariableString, Description: "Response tone", DefaultValue: "clear"},
			},
			OutputContract: PromptApplicationOutputContract{Kind: promptApplicationOutputText, MaxBytes: 4096},
		},
	}
}
