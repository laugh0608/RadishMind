package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"radishmind.local/services/platform/internal/config"
)

const (
	workflowTemplateTestSourceApplication = "app_aaaaaaaaaaaaaaaa"
	workflowTemplateTestTargetApplication = "app_bbbbbbbbbbbbbbbb"
	workflowTemplateTestDefinition        = "definition_template_source"
	workflowTemplateTestCandidate         = "candidate_template_source"
	workflowTemplateTestTemplate          = "template_workspace_source"
)

type workflowTemplateTestFixture struct {
	ctx          WorkflowTemplateCatalogContext
	releaseCtx   WorkflowDefinitionReleaseContext
	store        *memoryWorkflowTemplateCatalogRepository
	definitions  *workflowDefinitionReleaseStore
	applications *memoryApplicationCatalogRepository
	drafts       *memorySavedWorkflowDraftStore
	service      workflowTemplateCatalogService
	definition   WorkflowDefinitionVersion
}

func TestWorkflowTemplateCatalogLifecycleDerivesExactSavedDraft(t *testing.T) {
	fixture := newWorkflowTemplateTestFixture(t)
	if blocker := workflowDefinitionExecutionBlocker(workflowDefinitionSnapshotAsDraft(WorkflowRunContext{WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: workflowTemplateTestSourceApplication}, fixture.definition)); blocker != "" {
		t.Fatalf("portable definition fixture is not executable: %s", blocker)
	}
	if _, failure := validateWorkflowTemplatePortability(fixture.definition.Snapshot); failure != "" {
		t.Fatalf("portable definition fixture failed portability: failure=%s provider_valid=%t snapshot=%#v", failure, validWorkflowTemplateProviderBindings(fixture.definition.Snapshot), fixture.definition.Snapshot)
	}
	created := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
	if created.FailureCode != "" || created.Candidate == nil || created.Candidate.SourceDefinitionDigest != fixture.definition.DefinitionDigest {
		t.Fatalf("create exact template candidate: %#v", created)
	}
	reviewed := fixture.service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准工作区模板版本",
	})
	if reviewed.FailureCode != "" || reviewed.Candidate == nil || reviewed.Version == nil || reviewed.Version.Version != 1 || reviewed.Version.TemplateDigest == "" {
		t.Fatalf("approve and materialize template version: %#v", reviewed)
	}
	listed := fixture.service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架已批准模板版本",
	})
	if listed.FailureCode != "" || listed.Lineage == nil || listed.Lineage.PointerVersion != 1 || listed.Lineage.ListedVersion != 1 {
		t.Fatalf("list exact template version: %#v", listed)
	}
	derived := fixture.service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
		ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
		DraftID: "draft_from_workspace_template", Name: "由团队模板派生的草案", Confirmed: true,
	})
	if derived.FailureCode != "" || derived.Draft == nil || derived.Draft.DraftVersion != 1 ||
		derived.Draft.ApplicationID != workflowTemplateTestTargetApplication || derived.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("derive exact saved draft: %#v", derived)
	}
	provenance, ok := derived.Draft.AdditionalFields[savedWorkflowTemplateDerivationAdditionalField].(map[string]any)
	if !ok || provenance["template_id"] != workflowTemplateTestTemplate || provenance["template_digest"] != reviewed.Version.TemplateDigest ||
		provenance["source_definition_digest"] != fixture.definition.DefinitionDigest {
		t.Fatalf("derived draft provenance drifted: %#v", derived.Draft.AdditionalFields)
	}
	afterDefinition, err := fixture.definitions.ReadVersion(fixture.releaseCtx, workflowTemplateTestDefinition, 1)
	if err != nil || afterDefinition.DefinitionDigest != fixture.definition.DefinitionDigest {
		t.Fatalf("source definition changed during derivation: %#v err=%v", afterDefinition, err)
	}
	if got := fixture.drafts.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
		t.Fatalf("template lifecycle crossed execution boundary: %#v", got)
	}
}

func TestWorkflowTemplateCatalogReviewAndListingCASHaveSingleWinners(t *testing.T) {
	fixture := newWorkflowTemplateTestFixture(t)
	created := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
	if created.FailureCode != "" {
		t.Fatalf("create candidate: %#v", created)
	}
	var reviewWG sync.WaitGroup
	reviewResults := make(chan WorkflowTemplateCatalogResult, 2)
	for range 2 {
		reviewWG.Add(1)
		go func() {
			defer reviewWG.Done()
			reviewResults <- fixture.service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
				ExpectedReviewVersion: 0, Decision: "approve", Reason: "并发批准模板候选",
			})
		}()
	}
	reviewWG.Wait()
	close(reviewResults)
	reviewSuccess, reviewConflict := 0, 0
	for result := range reviewResults {
		switch result.FailureCode {
		case "":
			reviewSuccess++
		case WorkflowTemplateFailureCandidateVersionConflict, WorkflowTemplateFailureReviewTransitionInvalid:
			reviewConflict++
		default:
			t.Fatalf("unexpected review CAS result: %#v", result)
		}
	}
	if reviewSuccess != 1 || reviewConflict != 1 {
		t.Fatalf("review CAS winners drifted: success=%d conflict=%d", reviewSuccess, reviewConflict)
	}
	candidate, err := fixture.store.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate)
	if err != nil || candidate.ReviewVersion != 1 || len(candidate.Decisions) != 1 {
		t.Fatalf("review CAS left non-atomic candidate state: candidate=%#v err=%v", candidate, err)
	}
	versions, err := fixture.store.ListVersions(fixture.ctx, workflowTemplateTestTemplate)
	if err != nil || len(versions) != 1 {
		t.Fatalf("review CAS left non-atomic version state: versions=%#v err=%v", versions, err)
	}

	var listingWG sync.WaitGroup
	listingResults := make(chan WorkflowTemplateCatalogResult, 2)
	for range 2 {
		listingWG.Add(1)
		go func() {
			defer listingWG.Done()
			listingResults <- fixture.service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
				ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "并发上架模板版本",
			})
		}()
	}
	listingWG.Wait()
	close(listingResults)
	listingSuccess, listingConflict := 0, 0
	for result := range listingResults {
		if result.FailureCode == "" {
			listingSuccess++
		} else if result.FailureCode == WorkflowTemplateFailurePointerVersionConflict {
			listingConflict++
		} else {
			t.Fatalf("unexpected listing CAS result: %#v", result)
		}
	}
	if listingSuccess != 1 || listingConflict != 1 {
		t.Fatalf("listing CAS winners drifted: success=%d conflict=%d", listingSuccess, listingConflict)
	}
	lineage, err := fixture.store.ReadLineage(fixture.ctx, workflowTemplateTestTemplate)
	if err != nil || lineage.PointerVersion != 1 || len(lineage.Events) != 1 {
		t.Fatalf("listing CAS left non-atomic pointer/event state: lineage=%#v err=%v", lineage, err)
	}
	fixture.store.mu.RLock()
	audits := append([]WorkflowTemplateAudit(nil), fixture.store.audits[workflowTemplateScopeKey(fixture.ctx, "audits")]...)
	fixture.store.mu.RUnlock()
	if len(audits) != 4 {
		t.Fatalf("review/listing CAS left duplicate or partial audit records: %#v", audits)
	}
}

func TestWorkflowTemplateCatalogFailsClosedBeforeWrites(t *testing.T) {
	t.Run("missing source definition", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		input := workflowTemplateCandidateTestInput()
		input.SourceDefinitionID = "definition_missing_source"
		result := fixture.service.CreateCandidate(fixture.ctx, input)
		if result.FailureCode != WorkflowTemplateFailureSourceDefinitionNotFound {
			t.Fatalf("missing source definition was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("missing source definition left a candidate: %#v", candidates)
		}
	})

	t.Run("source definition version drift", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		input := workflowTemplateCandidateTestInput()
		input.SourceDefinitionVersion = 2
		result := fixture.service.CreateCandidate(fixture.ctx, input)
		if result.FailureCode != WorkflowTemplateFailureSourceDefinitionNotFound {
			t.Fatalf("unavailable source definition version was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("source definition version drift left a candidate: %#v", candidates)
		}
	})

	t.Run("workspace scope mismatch", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		otherScope := fixture.ctx
		otherScope.WorkspaceID = "workspace_other"
		result := fixture.service.CreateCandidate(otherScope, workflowTemplateCandidateTestInput())
		if result.FailureCode != WorkflowTemplateFailureSourceApplicationUnavailable {
			t.Fatalf("workspace scope mismatch was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("workspace scope mismatch left a candidate: %#v", candidates)
		}
	})

	t.Run("source application archived", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		applicationContext := ApplicationCatalogContext{RequestID: fixture.ctx.RequestID, TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID, ActorRef: fixture.ctx.ActorRef, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef, AuditRef: fixture.ctx.AuditRef}
		if _, err := fixture.applications.Archive(applicationContext, workflowTemplateTestSourceApplication, 1, ApplicationCatalogRecord{LifecycleState: applicationCatalogLifecycleArchived, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano), UpdatedByActorRef: fixture.ctx.ActorRef, RequestID: fixture.ctx.RequestID, AuditRef: fixture.ctx.AuditRef}); err != nil {
			t.Fatal(err)
		}
		result := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
		if result.FailureCode != WorkflowTemplateFailureSourceApplicationUnavailable {
			t.Fatalf("archived source application was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("archived source application left a candidate: %#v", candidates)
		}
	})

	t.Run("forbidden capability", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		key := workflowDefinitionScopeKey(fixture.releaseCtx, workflowTemplateTestDefinition)
		fixture.definitions.mu.Lock()
		version := fixture.definitions.versions[key][0]
		version.Snapshot.RequestedCapabilities = []string{"background_execution"}
		digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
		if err != nil {
			t.Fatal(err)
		}
		version.DefinitionDigest, version.SourceDraftDigest = digest, digest
		fixture.definitions.versions[key][0] = version
		fixture.definitions.mu.Unlock()
		result := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
		if result.FailureCode != WorkflowTemplateFailureForbiddenCapability {
			t.Fatalf("forbidden source was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("forbidden source wrote candidate: %#v", candidates)
		}
	})

	t.Run("forbidden node", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		key := workflowDefinitionScopeKey(fixture.releaseCtx, workflowTemplateTestDefinition)
		fixture.definitions.mu.Lock()
		version := fixture.definitions.versions[key][0]
		version.Snapshot.Nodes[0].NodeType = "tool"
		digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
		if err != nil {
			fixture.definitions.mu.Unlock()
			t.Fatal(err)
		}
		version.DefinitionDigest, version.SourceDraftDigest = digest, digest
		fixture.definitions.versions[key][0] = version
		fixture.definitions.mu.Unlock()
		result := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
		if result.FailureCode != WorkflowTemplateFailureForbiddenCapability {
			t.Fatalf("forbidden node was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("forbidden node wrote candidate: %#v", candidates)
		}
	})

	t.Run("forbidden material", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		forbiddenSnapshot := cloneWorkflowDefinitionSnapshot(fixture.definition.Snapshot)
		forbiddenSnapshot.Description = "embedded bearer secret-value"
		if _, failure := validateWorkflowTemplatePortability(forbiddenSnapshot); failure != WorkflowTemplateFailureSecretMaterialForbidden {
			t.Fatalf("portability validator accepted forbidden material: %s", failure)
		}
		key := workflowDefinitionScopeKey(fixture.releaseCtx, workflowTemplateTestDefinition)
		fixture.definitions.mu.Lock()
		version := fixture.definitions.versions[key][0]
		version.Snapshot = forbiddenSnapshot
		digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
		if err != nil {
			fixture.definitions.mu.Unlock()
			t.Fatal(err)
		}
		version.DefinitionDigest, version.SourceDraftDigest = digest, digest
		fixture.definitions.versions[key][0] = version
		fixture.definitions.mu.Unlock()
		result := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
		if result.FailureCode != WorkflowTemplateFailureStoreUnavailable {
			t.Fatalf("strict Definition owner corruption was not failed closed: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("forbidden source material wrote candidate: %#v", candidates)
		}
	})

	t.Run("forbidden candidate metadata", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		input := workflowTemplateCandidateTestInput()
		input.UsageNotes = "Bearer secret-value"
		result := fixture.service.CreateCandidate(fixture.ctx, input)
		if result.FailureCode != WorkflowTemplateFailureSecretMaterialForbidden {
			t.Fatalf("forbidden candidate metadata was not rejected: %#v", result)
		}
		if candidates, _ := fixture.store.ListCandidates(fixture.ctx); len(candidates) != 0 {
			t.Fatalf("forbidden candidate metadata wrote candidate: %#v", candidates)
		}
	})

	t.Run("definition digest drift", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		created := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
		if created.FailureCode != "" {
			t.Fatalf("create candidate: %#v", created)
		}
		key := workflowDefinitionScopeKey(fixture.releaseCtx, workflowTemplateTestDefinition)
		fixture.definitions.mu.Lock()
		version := fixture.definitions.versions[key][0]
		version.Snapshot.Name = "Changed immutable authority"
		digest, err := workflowDefinitionSnapshotDigest(version.Snapshot)
		if err != nil {
			t.Fatal(err)
		}
		version.DefinitionDigest, version.SourceDraftDigest = digest, digest
		fixture.definitions.versions[key][0] = version
		fixture.definitions.mu.Unlock()
		result := fixture.service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
			ExpectedReviewVersion: 0, Decision: "approve", Reason: "检测来源摘要漂移",
		})
		if result.FailureCode != WorkflowTemplateFailureSourceDefinitionChanged {
			t.Fatalf("definition drift was not rejected: %#v", result)
		}
		candidate, _ := fixture.store.ReadCandidate(fixture.ctx, workflowTemplateTestCandidate)
		if candidate.State != workflowTemplateCandidatePending || candidate.ReviewVersion != 0 {
			t.Fatalf("definition drift left partial review: %#v", candidate)
		}
	})

	t.Run("target binding and draft conflict", func(t *testing.T) {
		fixture := newWorkflowTemplateTestFixture(t)
		prepareListedWorkflowTemplate(t, fixture)
		fixture.service.targetBinding = rejectingWorkflowTemplateBinding{}
		blocked := fixture.service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
			ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
			DraftID: "draft_binding_blocked", Name: "目标绑定不可用草案", Confirmed: true,
		})
		if blocked.FailureCode != WorkflowTemplateFailureTargetBindingUnavailable || fixture.drafts.SideEffects().DraftWriteCount != 0 {
			t.Fatalf("target binding failure crossed draft owner: %#v side=%#v", blocked, fixture.drafts.SideEffects())
		}
		fixture.service.targetBinding = strictWorkflowTemplateTargetBindingValidator{}
		first := fixture.service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
			ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
			DraftID: "draft_conflict", Name: "首次模板派生草案", Confirmed: true,
		})
		second := fixture.service.Derive(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateDerivationInput{
			ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication,
			DraftID: "draft_conflict", Name: "重复模板派生草案", Confirmed: true,
		})
		if first.FailureCode != "" || second.FailureCode != WorkflowTemplateFailureDraftIDConflict || fixture.drafts.SideEffects().DraftWriteCount != 1 {
			t.Fatalf("draft conflict semantics drifted: first=%#v second=%#v side=%#v", first, second, fixture.drafts.SideEffects())
		}
	})
}

func TestWorkflowTemplateStrictCodecAndCursorBindings(t *testing.T) {
	fixture := newWorkflowTemplateTestFixture(t)
	created := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput())
	if created.FailureCode != "" || created.Candidate == nil {
		t.Fatalf("create candidate: %#v", created)
	}
	payload, err := encodeWorkflowTemplateRecord(*created.Candidate)
	if err != nil {
		t.Fatalf("encode strict candidate: %v", err)
	}
	var decoded WorkflowTemplateCandidate
	if err = decodeWorkflowTemplateRecord(payload, &decoded); err != nil || decoded.CandidateID != created.Candidate.CandidateID {
		t.Fatalf("decode strict candidate: decoded=%#v err=%v", decoded, err)
	}
	unknown := append(payload[:len(payload)-1], []byte(`,"unknown":true}`)...)
	if err = decodeWorkflowTemplateRecord(unknown, &WorkflowTemplateCandidate{}); err == nil {
		t.Fatalf("strict candidate codec accepted unknown field: %s", unknown)
	}
	duplicate := []byte(`{"schema_version":"workspace_workflow_template_candidate.v1","schema_version":"workspace_workflow_template_candidate.v1"}`)
	if err = decodeWorkflowTemplateRecord(duplicate, &WorkflowTemplateCandidate{}); err == nil {
		t.Fatal("strict candidate codec accepted duplicate field")
	}
	reviewed := fixture.service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{
		ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准 codec 模板候选",
	})
	if reviewed.FailureCode != "" || reviewed.Candidate == nil || reviewed.Version == nil {
		t.Fatalf("prepare strict codec records: %#v", reviewed)
	}
	listed := fixture.service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{
		ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架 codec 模板版本",
	})
	if listed.FailureCode != "" || listed.Lineage == nil {
		t.Fatalf("prepare strict listing codec records: %#v", listed)
	}
	fixture.store.mu.RLock()
	audits := append([]WorkflowTemplateAudit(nil), fixture.store.audits[workflowTemplateScopeKey(fixture.ctx, "audits")]...)
	fixture.store.mu.RUnlock()
	if len(audits) == 0 {
		t.Fatal("strict audit codec fixture is empty")
	}
	assertWorkflowTemplateRecordRoundTrip(t, reviewed.Candidate.Decisions[0], &WorkflowTemplateReviewDecision{})
	assertWorkflowTemplateRecordRoundTrip(t, *reviewed.Version, &WorkflowTemplateVersion{})
	assertWorkflowTemplateRecordRoundTrip(t, listed.Lineage.Events[0], &WorkflowTemplateListingEvent{})
	assertWorkflowTemplateRecordRoundTrip(t, *listed.Lineage, &WorkflowTemplateLineage{})
	assertWorkflowTemplateRecordRoundTrip(t, audits[len(audits)-1], &WorkflowTemplateAudit{})
	corruptedLineage := *listed.Lineage
	corruptedLineage.Events = append([]WorkflowTemplateListingEvent(nil), listed.Lineage.Events...)
	corruptedLineage.Events[0].AfterListedVersion = 0
	if err := validateStoredWorkflowTemplateLineage(corruptedLineage); err == nil {
		t.Fatal("strict lineage codec accepted a listing transition mismatch")
	}

	first := fixture.service.ListCandidates(fixture.ctx, WorkflowTemplateListInput{Limit: 1})
	if first.FailureCode != "" || len(first.Candidates) != 1 {
		t.Fatalf("list first candidate page: %#v", first)
	}
	cursor := encodeWorkflowTemplateListCursor(fixture.ctx, "candidate", "", 1, time.Now().UTC().Add(time.Minute).Format(time.RFC3339Nano), first.Candidates[0].UpdatedAt, first.Candidates[0].CandidateID)
	otherScope := fixture.ctx
	otherScope.WorkspaceID = "workspace_other"
	if result := fixture.service.ListCandidates(otherScope, WorkflowTemplateListInput{Limit: 1, Cursor: cursor}); result.FailureCode != WorkflowTemplateFailureCursorInvalid {
		t.Fatalf("workspace scope drift did not invalidate cursor: %#v", result)
	}
}

func assertWorkflowTemplateRecordRoundTrip(t *testing.T, value, target any) {
	t.Helper()
	payload, err := encodeWorkflowTemplateRecord(value)
	if err != nil {
		t.Fatalf("encode strict %T: %v", value, err)
	}
	if err = decodeWorkflowTemplateRecord(payload, target); err != nil {
		t.Fatalf("decode strict %T: %v payload=%s", value, err, payload)
	}
}

func TestSavedWorkflowDraftTemplateDerivationIsStrictAndCompatible(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	ctx := savedWorkflowDraftTestContext()
	payload := portableWorkflowTemplateDraftPayload()
	ctx.WorkspaceID, ctx.ApplicationID = payload.WorkspaceID, payload.ApplicationID
	payload.DraftID = "draft_template_provenance"
	payload.AdditionalFields = map[string]any{savedWorkflowTemplateDerivationAdditionalField: validWorkflowTemplateDerivationTestDocument()}
	created := service.SaveDraft(ctx, SaveWorkflowDraftRequest{Payload: payload})
	if created.FailureCode != "" || created.Draft == nil || created.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceTemplate {
		t.Fatalf("strict derivation_v2 was not persisted: %#v", created)
	}
	invalid := portableWorkflowTemplateDraftPayload()
	invalid.DraftID = "draft_invalid_union"
	invalid.AdditionalFields = map[string]any{
		savedWorkflowTemplateDerivationAdditionalField: validWorkflowTemplateDerivationTestDocument(),
		savedWorkflowDraftDerivationAdditionalField:    map[string]any{"version": 1, "source_kind": "saved_workflow_draft", "source_draft_id": "draft_parent", "source_draft_version": 1},
	}
	if result := service.SaveDraft(ctx, SaveWorkflowDraftRequest{Payload: invalid}); result.FailureCode != SavedWorkflowDraftFailurePayloadInvalid {
		t.Fatalf("invalid provenance union was accepted: %#v", result)
	}
	legacy := portableWorkflowTemplateDraftPayload()
	legacy.DraftID = "draft_legacy_derivation"
	legacy.AdditionalFields = map[string]any{savedWorkflowDraftDerivationAdditionalField: map[string]any{
		"version": 1, "source_kind": "saved_workflow_draft", "source_draft_id": "draft_missing", "source_draft_version": 1,
	}}
	if result := service.SaveDraft(ctx, SaveWorkflowDraftRequest{Payload: legacy}); result.FailureCode != SavedWorkflowDraftFailureNotFound {
		t.Fatalf("legacy derivation_v1 authority behavior changed: %#v", result)
	}
}

func TestWorkflowTemplateCatalogStrictHTTPAndDefaultGate(t *testing.T) {
	disabled := NewServer(config.Config{}, Options{BuildVersion: "test"})
	defer disabled.Close()
	disabledRequest := httptest.NewRequest(http.MethodGet, "/v1/user-workspace/workflow-templates?workspace_id=workspace_demo", nil)
	setSavedWorkflowDraftDevHeaders(disabledRequest, "workflow_definitions:read")
	disabledRecorder := httptest.NewRecorder()
	disabled.httpServer.Handler.ServeHTTP(disabledRecorder, disabledRequest)
	if disabledRecorder.Code == http.StatusOK || !bytes.Contains(disabledRecorder.Body.Bytes(), []byte("WORKFLOW_TEMPLATE_CATALOG_DEV_HTTP_DISABLED")) {
		t.Fatalf("template catalog route was not disabled by default: status=%d body=%s", disabledRecorder.Code, disabledRecorder.Body.String())
	}

	server := NewServer(config.Config{
		ControlPlaneReadDevAuthEnabled: true, WorkflowSavedDraftDevHTTPEnabled: true, WorkflowSavedDraftDevWriteEnabled: true,
		ApplicationCatalogDevHTTPEnabled: true, WorkflowDefinitionReleaseDevEnabled: true, WorkflowTemplateCatalogDevEnabled: true,
	}, Options{BuildVersion: "test"})
	defer server.Close()
	seedWorkflowTemplateServerAuthorities(t, server)
	created := performWorkflowTemplateRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-template-candidates", workflowTemplateCandidateCreateBodyFromInput(workflowTemplateCandidateTestInput()), "workflow_definitions:read,workflow_definitions:write")
	if created.Candidate == nil || created.FailureCode != nil {
		t.Fatalf("HTTP create candidate: %#v", created)
	}
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-template-candidates?workspace_id=workspace_demo", workflowTemplateTestCandidate)
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-template-candidates/"+workflowTemplateTestCandidate+"?workspace_id=workspace_demo", workflowTemplateTestCandidate)
	reviewed := performWorkflowTemplateRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-template-candidates/"+workflowTemplateTestCandidate+"/decisions", workflowTemplateCandidateDecisionBody{ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准 HTTP 模板候选"}, "workflow_definitions:read,workflow_definitions:review")
	if reviewed.Version == nil || reviewed.FailureCode != nil {
		t.Fatalf("HTTP review candidate: %#v", reviewed)
	}
	listed := performWorkflowTemplateRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/listing-decisions", workflowTemplateListingDecisionBody{ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架 HTTP 模板版本"}, "workflow_definitions:read,workflow_definitions:activate")
	if listed.Lineage == nil || listed.FailureCode != nil {
		t.Fatalf("HTTP list template: %#v", listed)
	}
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-templates?workspace_id=workspace_demo", workflowTemplateTestTemplate)
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"?workspace_id=workspace_demo", workflowTemplateTestTemplate)
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/versions?workspace_id=workspace_demo", `"version":1`)
	assertWorkflowTemplateGET(t, server, "/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/versions/1?workspace_id=workspace_demo", `"version":1`)
	derived := performWorkflowTemplateRequest(t, server, http.MethodPost, "/v1/user-workspace/workflow-templates/"+workflowTemplateTestTemplate+"/derivations", workflowTemplateDerivationBody{ExpectedPointerVersion: 1, TemplateVersion: 1, TargetApplicationID: workflowTemplateTestTargetApplication, DraftID: "draft_http_template", Name: "HTTP 模板派生草案", Confirmed: true}, "workflow_definitions:read,workflow_drafts:write")
	if derived.Draft == nil || derived.Draft.ProvenanceKind != string(SavedWorkflowDraftProvenanceTemplate) || derived.FailureCode != nil {
		t.Fatalf("HTTP derive template: %#v", derived)
	}
	duplicate := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-template-candidates", bytes.NewBufferString(`{"candidate_id":"one","candidate_id":"two"}`))
	duplicate.Header.Set("Content-Type", "application/json")
	setSavedWorkflowDraftDevHeaders(duplicate, "workflow_definitions:read,workflow_definitions:write")
	duplicateRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(duplicateRecorder, duplicate)
	if duplicateRecorder.Code != http.StatusBadRequest || !bytes.Contains(duplicateRecorder.Body.Bytes(), []byte("INVALID_JSON")) {
		t.Fatalf("duplicate JSON field reached template owner: status=%d body=%s", duplicateRecorder.Code, duplicateRecorder.Body.String())
	}
	denied := httptest.NewRequest(http.MethodPost, "/v1/user-workspace/workflow-template-candidates", bytes.NewReader([]byte(`{}`)))
	denied.Header.Set("Content-Type", "application/json")
	setSavedWorkflowDraftDevHeaders(denied, "workflow_definitions:read")
	deniedRecorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(deniedRecorder, denied)
	if deniedRecorder.Code != http.StatusForbidden || !bytes.Contains(deniedRecorder.Body.Bytes(), []byte(WorkflowTemplateFailureScopeDenied)) {
		t.Fatalf("missing combined permission reached template owner: status=%d body=%s", deniedRecorder.Code, deniedRecorder.Body.String())
	}
}

func assertWorkflowTemplateGET(t *testing.T, server *Server, target, expected string) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, target, nil)
	setSavedWorkflowDraftDevHeaders(request, "workflow_definitions:read")
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !bytes.Contains(recorder.Body.Bytes(), []byte(expected)) || bytes.Contains(recorder.Body.Bytes(), []byte(`"failure_code":"`)) {
		t.Fatalf("GET %s failed: status=%d body=%s", target, recorder.Code, recorder.Body.String())
	}
}

type rejectingWorkflowTemplateBinding struct{}

func (rejectingWorkflowTemplateBinding) ValidateTargetBinding(ApplicationCatalogRecord, WorkflowDefinitionSnapshot) error {
	return errWorkflowTemplateStoreUnavailable
}

func newWorkflowTemplateTestFixture(t *testing.T) *workflowTemplateTestFixture {
	t.Helper()
	fixture := &workflowTemplateTestFixture{
		ctx:   WorkflowTemplateCatalogContext{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", OwnerSubjectRef: "subject_demo", ActorRef: "subject_demo", RequestID: "request_template", AuditRef: "audit_template"},
		store: newMemoryWorkflowTemplateCatalogRepository(), definitions: newWorkflowDefinitionReleaseStore(),
		applications: newMemoryApplicationCatalogRepository(), drafts: newMemorySavedWorkflowDraftStore(),
	}
	fixture.releaseCtx = WorkflowDefinitionReleaseContext{TenantRef: fixture.ctx.TenantRef, WorkspaceID: fixture.ctx.WorkspaceID, ApplicationID: workflowTemplateTestSourceApplication, OwnerSubjectRef: fixture.ctx.OwnerSubjectRef, ActorRef: fixture.ctx.ActorRef, RequestID: fixture.ctx.RequestID, AuditRef: fixture.ctx.AuditRef}
	seedWorkflowTemplateApplication(t, fixture.applications, fixture.ctx, workflowTemplateTestSourceApplication, "Source workflow application")
	seedWorkflowTemplateApplication(t, fixture.applications, fixture.ctx, workflowTemplateTestTargetApplication, "Target workflow application")
	draft := savedWorkflowDraftFromPayload(portableWorkflowTemplateDraftPayload())
	draft.ApplicationID, draft.WorkspaceID = workflowTemplateTestSourceApplication, fixture.ctx.WorkspaceID
	draft.ValidationSummary = SavedWorkflowDraftValidationSummary{ValidationState: SavedWorkflowDraftStatusValidForReview, ValidForReview: true}
	candidate, err := fixture.definitions.CreateCandidate(fixture.releaseCtx, "definition_candidate_template", workflowTemplateTestDefinition, workflowDefinitionExecutorProfile, draft, time.Date(2026, 8, 27, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("seed definition candidate: %v", err)
	}
	_, version, err := fixture.definitions.Review(fixture.releaseCtx, candidate.CandidateID, 0, "approve", "批准模板来源 Definition", candidate.SourceDraftDigest, time.Date(2026, 8, 27, 9, 1, 0, 0, time.UTC))
	if err != nil || version == nil {
		t.Fatalf("seed definition version: version=%#v err=%v", version, err)
	}
	fixture.definition = *version
	fixture.service = newWorkflowTemplateCatalogService(fixture.store, fixture.definitions, fixture.applications, fixture.drafts)
	return fixture
}

func portableWorkflowTemplateDraftPayload() SavedWorkflowDraftPayload {
	payload := validSavedWorkflowDraftPayload()
	payload.DraftID = "draft_template_definition_source"
	payload.WorkspaceID = "workspace_demo"
	payload.ApplicationID = workflowTemplateTestSourceApplication
	payload.ToolRefs = []string{}
	payload.RAGRefs = []string{}
	payload.RequestedCapabilities = []string{}
	return payload
}

func seedWorkflowTemplateApplication(t *testing.T, repository *memoryApplicationCatalogRepository, ctx WorkflowTemplateCatalogContext, applicationID, name string) {
	t.Helper()
	now := time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	applicationContext := ApplicationCatalogContext{RequestID: ctx.RequestID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ActorRef: ctx.ActorRef, OwnerSubjectRef: ctx.OwnerSubjectRef, AuditRef: ctx.AuditRef}
	_, err := repository.Create(applicationContext, ApplicationCatalogRecord{
		SchemaVersion: applicationCatalogSchemaVersion, ApplicationID: applicationID, TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID,
		OwnerSubjectRef: ctx.OwnerSubjectRef, DisplayName: name, ApplicationKind: "workflow_copilot", LifecycleState: applicationCatalogLifecycleActive,
		RecordVersion: 1, CreatedAt: now, UpdatedAt: now, CreatedByActorRef: ctx.ActorRef, UpdatedByActorRef: ctx.ActorRef,
		RequestID: ctx.RequestID, AuditRef: ctx.AuditRef,
	})
	if err != nil {
		t.Fatalf("seed application %s: %v", applicationID, err)
	}
}

func workflowTemplateCandidateTestInput() WorkflowTemplateCandidateCreateInput {
	return WorkflowTemplateCandidateCreateInput{
		CandidateID: workflowTemplateTestCandidate, TemplateID: workflowTemplateTestTemplate,
		SourceApplicationID: workflowTemplateTestSourceApplication, SourceDefinitionID: workflowTemplateTestDefinition,
		SourceDefinitionVersion: 1, Title: "团队问答工作流", Summary: "供工作区成员复用的受控问答流程。",
		UsageNotes: "派生后重新检查目标应用模型绑定。", Labels: []string{"team", "qa"},
	}
}

func prepareListedWorkflowTemplate(t *testing.T, fixture *workflowTemplateTestFixture) {
	t.Helper()
	if result := fixture.service.CreateCandidate(fixture.ctx, workflowTemplateCandidateTestInput()); result.FailureCode != "" {
		t.Fatalf("create candidate: %#v", result)
	}
	if result := fixture.service.ReviewCandidate(fixture.ctx, workflowTemplateTestCandidate, WorkflowTemplateReviewInput{ExpectedReviewVersion: 0, Decision: "approve", Reason: "批准模板候选版本"}); result.FailureCode != "" {
		t.Fatalf("review candidate: %#v", result)
	}
	if result := fixture.service.DecideListing(fixture.ctx, workflowTemplateTestTemplate, WorkflowTemplateListingInput{ExpectedPointerVersion: 0, Decision: "list", Version: 1, Reason: "上架模板候选版本"}); result.FailureCode != "" {
		t.Fatalf("list version: %#v", result)
	}
}

func validWorkflowTemplateDerivationTestDocument() map[string]any {
	return map[string]any{
		"version": 2, "source_kind": "workspace_workflow_template", "template_id": workflowTemplateTestTemplate,
		"template_version": 1, "template_digest": "sha256:" + string(bytes.Repeat([]byte{'a'}, 64)),
		"source_definition_id": workflowTemplateTestDefinition, "source_definition_version": 1,
		"source_definition_digest": "sha256:" + string(bytes.Repeat([]byte{'b'}, 64)),
	}
}

func seedWorkflowTemplateServerAuthorities(t *testing.T, server *Server) {
	t.Helper()
	ctx := WorkflowTemplateCatalogContext{TenantRef: "tenant_demo", WorkspaceID: "workspace_demo", OwnerSubjectRef: "subject_demo_user", ActorRef: "subject_demo_user", RequestID: "request_server_seed", AuditRef: "audit_server_seed"}
	applications := server.applicationCatalogRepository.(*memoryApplicationCatalogRepository)
	seedWorkflowTemplateApplication(t, applications, ctx, workflowTemplateTestSourceApplication, "Source workflow application")
	seedWorkflowTemplateApplication(t, applications, ctx, workflowTemplateTestTargetApplication, "Target workflow application")
	releaseCtx := WorkflowDefinitionReleaseContext{TenantRef: ctx.TenantRef, WorkspaceID: ctx.WorkspaceID, ApplicationID: workflowTemplateTestSourceApplication, OwnerSubjectRef: ctx.OwnerSubjectRef, ActorRef: ctx.ActorRef, RequestID: ctx.RequestID, AuditRef: ctx.AuditRef}
	draft := savedWorkflowDraftFromPayload(portableWorkflowTemplateDraftPayload())
	draft.ApplicationID = workflowTemplateTestSourceApplication
	draft.ValidationSummary = SavedWorkflowDraftValidationSummary{ValidationState: SavedWorkflowDraftStatusValidForReview, ValidForReview: true}
	candidate, err := server.workflowDefinitionReleaseRepository.CreateCandidate(releaseCtx, "definition_candidate_template", workflowTemplateTestDefinition, workflowDefinitionExecutorProfile, draft, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if _, version, err := server.workflowDefinitionReleaseRepository.Review(releaseCtx, candidate.CandidateID, 0, "approve", "批准 HTTP Definition 来源", candidate.SourceDraftDigest, time.Now().UTC()); err != nil || version == nil {
		t.Fatalf("seed HTTP definition: version=%#v err=%v", version, err)
	}
}

func workflowTemplateCandidateCreateBodyFromInput(input WorkflowTemplateCandidateCreateInput) workflowTemplateCandidateCreateBody {
	return workflowTemplateCandidateCreateBody{CandidateID: input.CandidateID, TemplateID: input.TemplateID, SourceApplicationID: input.SourceApplicationID, SourceDefinitionID: input.SourceDefinitionID, SourceDefinitionVersion: input.SourceDefinitionVersion, Title: input.Title, Summary: input.Summary, UsageNotes: input.UsageNotes, Labels: input.Labels}
}

func performWorkflowTemplateRequest(t *testing.T, server *Server, method, target string, body any, scopes string) workflowTemplateCatalogEnvelope {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(method, target, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	setSavedWorkflowDraftDevHeaders(request, scopes)
	recorder := httptest.NewRecorder()
	server.httpServer.Handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
	}
	var envelope workflowTemplateCatalogEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode template response: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}
