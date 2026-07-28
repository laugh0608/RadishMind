package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSavedWorkflowDraftMemoryLifecycleRoundTrip(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	tick := 0
	service.now = func() time.Time {
		tick++
		return time.Date(2026, 7, 28, 12, 0, tick, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftTestContext()
	payload := validSavedWorkflowDraftPayload()

	saved := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("save active draft: %#v", saved)
	}
	if saved.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		saved.Draft.LifecycleVersion != 1 ||
		saved.Draft.ArchivedAt != "" ||
		saved.Draft.LibraryUpdatedAt != saved.Draft.UpdatedAt ||
		saved.Draft.LifecycleUpdatedByActorRef != "" ||
		saved.Draft.ProvenanceKind != SavedWorkflowDraftProvenanceDefinition {
		t.Fatalf("new draft lifecycle defaults drifted: %#v", saved.Draft)
	}

	archived := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  payload.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" || archived.Draft == nil || archived.Event == nil {
		t.Fatalf("archive active draft: %#v", archived)
	}
	if archived.Draft.DraftVersion != 1 ||
		archived.Draft.LifecycleState != SavedWorkflowDraftLifecycleArchived ||
		archived.Draft.LifecycleVersion != 2 ||
		archived.Draft.ArchivedAt == "" ||
		archived.Draft.LibraryUpdatedAt != archived.Draft.ArchivedAt {
		t.Fatalf("archive changed the wrong version or timestamps: %#v", archived.Draft)
	}
	if archived.Event.FromState != SavedWorkflowDraftLifecycleActive ||
		archived.Event.ToState != SavedWorkflowDraftLifecycleArchived ||
		archived.Event.TransitionKind != SavedWorkflowDraftLifecycleTransitionArchived ||
		archived.Event.LifecycleVersion != 2 {
		t.Fatalf("archive event drifted: %#v", archived.Event)
	}

	activePage := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{})
	if activePage.FailureCode != "" || len(activePage.Summaries) != 0 {
		t.Fatalf("default list must exclude archived drafts: %#v", activePage)
	}
	archivedPage := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{
		LifecycleState: SavedWorkflowDraftLifecycleArchived,
	})
	if archivedPage.FailureCode != "" || len(archivedPage.Summaries) != 1 ||
		archivedPage.Summaries[0].LifecycleVersion != 2 {
		t.Fatalf("archived list should expose lifecycle projection: %#v", archivedPage)
	}
	read := service.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: payload.DraftID})
	if read.FailureCode != "" || read.Draft == nil ||
		read.Draft.LifecycleState != SavedWorkflowDraftLifecycleArchived {
		t.Fatalf("archived draft must remain readable: %#v", read)
	}

	saveBlocked := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 2,
		Payload:                  payload,
	})
	if saveBlocked.FailureCode != SavedWorkflowDraftFailureArchived || saveBlocked.Draft != nil {
		t.Fatalf("archived draft save must fail closed: %#v", saveBlocked)
	}
	restoreBlocked := service.RestoreDraftRevision(requestContext, RestoreSavedWorkflowDraftRevisionRequest{
		DraftID:                     payload.DraftID,
		SourceDraftVersion:          1,
		ExpectedCurrentDraftVersion: 1,
		ExpectedLifecycleVersion:    2,
	})
	if restoreBlocked.FailureCode != SavedWorkflowDraftFailureArchived || restoreBlocked.Draft != nil {
		t.Fatalf("archived draft restore must fail closed: %#v", restoreBlocked)
	}
	staleLifecycle := service.UnarchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  payload.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 1,
	})
	if staleLifecycle.FailureCode != SavedWorkflowDraftFailureLifecycleVersionConflict ||
		staleLifecycle.CurrentLifecycleVersion != 2 ||
		staleLifecycle.CurrentLifecycleState != SavedWorkflowDraftLifecycleArchived {
		t.Fatalf("stale lifecycle CAS must return current metadata: %#v", staleLifecycle)
	}
	repeatedArchive := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  payload.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 2,
	})
	if repeatedArchive.FailureCode != SavedWorkflowDraftFailureLifecycleStateConflict {
		t.Fatalf("repeated archive must report state conflict: %#v", repeatedArchive)
	}

	unarchived := service.UnarchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  payload.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 2,
	})
	if unarchived.FailureCode != "" || unarchived.Draft == nil || unarchived.Event == nil ||
		unarchived.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
		unarchived.Draft.LifecycleVersion != 3 ||
		unarchived.Draft.ArchivedAt != "" ||
		unarchived.Event.TransitionKind != SavedWorkflowDraftLifecycleTransitionUnarchived {
		t.Fatalf("unarchive round trip drifted: %#v", unarchived)
	}

	store.mu.RLock()
	eventCount := len(store.lifecycleEvents[payload.DraftID])
	revisionCount := len(store.revisions[payload.DraftID])
	store.mu.RUnlock()
	if eventCount != 2 || revisionCount != 1 {
		t.Fatalf("lifecycle transitions must append events without content revisions: events=%d revisions=%d", eventCount, revisionCount)
	}
	sideEffects := store.SideEffects()
	if sideEffects.DraftWriteCount != 1 ||
		sideEffects.LifecycleTransitionCount != 2 ||
		sideEffects.LifecycleEventWriteCount != 2 ||
		hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
		t.Fatalf("unexpected lifecycle side effects: %#v", sideEffects)
	}
}

func TestSavedWorkflowDraftMemoryLifecycleLegacyAndAtomicEvent(t *testing.T) {
	t.Run("legacy fixture defaults without fake event", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		legacy := savedWorkflowDraftFromPayload(validSavedWorkflowDraftPayload())
		store.drafts[legacy.DraftID] = legacy

		service := newSavedWorkflowDraftService(store)
		read := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{
			DraftID: legacy.DraftID,
		})
		if read.FailureCode != "" || read.Draft == nil ||
			read.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
			read.Draft.LifecycleVersion != 1 ||
			read.Draft.LibraryUpdatedAt != legacy.UpdatedAt {
			t.Fatalf("legacy fixture lifecycle was not normalized: %#v", read)
		}
		if len(store.lifecycleEvents[legacy.DraftID]) != 0 {
			t.Fatalf("legacy normalization must not invent lifecycle events: %#v", store.lifecycleEvents)
		}
	})

	t.Run("event write failure rolls back current state", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		service.now = fixedSavedWorkflowDraftClock()
		requestContext := savedWorkflowDraftTestContext()
		saved := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
			Payload: validSavedWorkflowDraftPayload(),
		})
		if saved.FailureCode != "" || saved.Draft == nil {
			t.Fatalf("seed draft: %#v", saved)
		}
		store.lifecycleEventWriteFailure = true

		result := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
			DraftID:                  saved.Draft.DraftID,
			ExpectedDraftVersion:     1,
			ExpectedLifecycleVersion: 1,
		})
		if result.FailureCode != SavedWorkflowDraftFailureLifecycleEventWrite {
			t.Fatalf("event failure should be stable: %#v", result)
		}
		read := service.ReadDraft(requestContext, ReadWorkflowDraftRequest{DraftID: saved.Draft.DraftID})
		if read.FailureCode != "" || read.Draft == nil ||
			read.Draft.LifecycleState != SavedWorkflowDraftLifecycleActive ||
			read.Draft.LifecycleVersion != 1 ||
			len(store.lifecycleEvents[saved.Draft.DraftID]) != 0 {
			t.Fatalf("event failure left a partial transition: read=%#v events=%#v", read, store.lifecycleEvents)
		}
		if sideEffects := store.SideEffects(); sideEffects.LifecycleTransitionCount != 0 ||
			sideEffects.LifecycleEventWriteCount != 0 {
			t.Fatalf("rolled back transition changed side effects: %#v", sideEffects)
		}
	})
}

func TestSavedWorkflowDraftMemoryLibraryPaginationBeyondRepositoryWindow(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	service.now = func() time.Time {
		return time.Date(2026, 7, 28, 13, 0, 0, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftTestContext()
	const draftCount = 231
	for index := 0; index < draftCount; index++ {
		payload := validSavedWorkflowDraftPayload()
		payload.DraftID = fmt.Sprintf("draft_library_%03d", index)
		payload.Name = fmt.Sprintf("Library draft %03d", index)
		result := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
		if result.FailureCode != "" {
			t.Fatalf("seed draft %d: %#v", index, result)
		}
	}
	defaultPage := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{})
	if defaultPage.FailureCode != "" ||
		len(defaultPage.Summaries) != defaultSavedWorkflowDraftListLimit ||
		!defaultPage.HasMore ||
		defaultPage.NextCursor == "" {
		t.Fatalf("default page size drifted: %#v", defaultPage)
	}

	seen := make([]string, 0, draftCount)
	cursor := ""
	for {
		page := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{
			Limit:  100,
			Cursor: cursor,
		})
		if page.FailureCode != "" {
			t.Fatalf("list page after %q: %#v", cursor, page)
		}
		for _, summary := range page.Summaries {
			seen = append(seen, summary.DraftID)
		}
		if !page.HasMore {
			if page.NextCursor != "" {
				t.Fatalf("terminal page returned cursor: %#v", page)
			}
			break
		}
		if page.NextCursor == "" || len(page.Summaries) != 100 {
			t.Fatalf("non-terminal page contract drifted: %#v", page)
		}
		cursor = page.NextCursor
	}
	if len(seen) != draftCount {
		t.Fatalf("pagination truncated the library: got %d want %d", len(seen), draftCount)
	}
	for index, draftID := range seen {
		want := fmt.Sprintf("draft_library_%03d", index)
		if draftID != want {
			t.Fatalf("same-time draft_id tie-break drifted at %d: got %q want %q", index, draftID, want)
		}
	}
}

func TestSavedWorkflowDraftMemoryLibraryFilters(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	tick := 0
	service.now = func() time.Time {
		tick++
		return time.Date(2026, 7, 28, 14, 0, tick, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftTestContext()

	unversioned := validSavedWorkflowDraftPayload()
	unversioned.DraftID = "draft_filter_unversioned"
	unversioned.Name = "萝卜 Alpha"
	unversioned.BaseDefinitionVersion = 0
	saveWorkflowDraftLibraryFixture(t, service, requestContext, unversioned)

	definition := validSavedWorkflowDraftPayload()
	definition.DraftID = "draft_filter_definition"
	definition.Name = "萝卜 Definition"
	definition.BaseDefinitionVersion = 4
	saveWorkflowDraftLibraryFixture(t, service, requestContext, definition)

	derived := validSavedWorkflowDraftPayload()
	derived.DraftID = "draft_filter_derived"
	derived.Name = "Derived Alpha"
	derived.BaseDefinitionVersion = 8
	derived.AdditionalFields = map[string]any{
		savedWorkflowDraftDerivationAdditionalField: map[string]any{
			"version":              1,
			"source_kind":          savedWorkflowDraftDerivationSourceKind,
			"source_draft_id":      unversioned.DraftID,
			"source_draft_version": 1,
		},
	}
	saveWorkflowDraftLibraryFixture(t, service, requestContext, derived)

	invalid := validSavedWorkflowDraftPayload()
	invalid.DraftID = "draft_filter_invalid"
	invalid.Name = "萝卜 Invalid"
	invalid.OutputContract.RequiredFields = nil
	saveWorkflowDraftLibraryFixture(t, service, requestContext, invalid)

	archived := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  definition.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" {
		t.Fatalf("archive definition fixture: %#v", archived)
	}

	testCases := []struct {
		name    string
		request ListWorkflowDraftsRequest
		wantIDs []string
	}{
		{
			name:    "default active",
			request: ListWorkflowDraftsRequest{},
			wantIDs: []string{invalid.DraftID, derived.DraftID, unversioned.DraftID},
		},
		{
			name: "archived",
			request: ListWorkflowDraftsRequest{
				LifecycleState: SavedWorkflowDraftLifecycleArchived,
			},
			wantIDs: []string{definition.DraftID},
		},
		{
			name: "name prefix",
			request: ListWorkflowDraftsRequest{
				NamePrefix: "萝卜",
			},
			wantIDs: []string{invalid.DraftID, unversioned.DraftID},
		},
		{
			name: "validation",
			request: ListWorkflowDraftsRequest{
				ValidationState: SavedWorkflowDraftStatusInvalidDraft,
			},
			wantIDs: []string{invalid.DraftID},
		},
		{
			name: "provenance precedence",
			request: ListWorkflowDraftsRequest{
				ProvenanceKind: SavedWorkflowDraftProvenanceDraftDerived,
			},
			wantIDs: []string{derived.DraftID},
		},
		{
			name: "combined",
			request: ListWorkflowDraftsRequest{
				NamePrefix:      "萝卜",
				ValidationState: SavedWorkflowDraftStatusValidForReview,
				ProvenanceKind:  SavedWorkflowDraftProvenanceUnversioned,
			},
			wantIDs: []string{unversioned.DraftID},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := service.ListDrafts(requestContext, testCase.request)
			if result.FailureCode != "" {
				t.Fatalf("list filters: %#v", result)
			}
			got := make([]string, 0, len(result.Summaries))
			for _, summary := range result.Summaries {
				got = append(got, summary.DraftID)
			}
			if fmt.Sprint(got) != fmt.Sprint(testCase.wantIDs) {
				t.Fatalf("filter result drifted: got %v want %v", got, testCase.wantIDs)
			}
		})
	}
}

func TestSavedWorkflowDraftLibraryCursorBindingAndFilterValidation(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	tick := 0
	service.now = func() time.Time {
		tick++
		return time.Date(2026, 7, 28, 15, 0, tick, 0, time.UTC)
	}
	requestContext := savedWorkflowDraftTestContext()
	for index := 0; index < 2; index++ {
		payload := validSavedWorkflowDraftPayload()
		payload.DraftID = fmt.Sprintf("draft_cursor_%d", index)
		payload.Name = fmt.Sprintf("Cursor name %d", index)
		saveWorkflowDraftLibraryFixture(t, service, requestContext, payload)
	}
	first := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{Limit: 1})
	if first.FailureCode != "" || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("seed cursor page: %#v", first)
	}
	raw, err := base64.RawURLEncoding.DecodeString(first.NextCursor)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	for _, forbidden := range []string{
		"Cursor name",
		requestContext.ActorRef,
		requestContext.RequestID,
		requestContext.AuditRef,
	} {
		if forbidden != "" && strings.Contains(string(raw), forbidden) {
			t.Fatalf("cursor leaked forbidden value %q: %s", forbidden, raw)
		}
	}

	baseRequest := ListWorkflowDraftsRequest{Limit: 1, Cursor: first.NextCursor}
	driftCases := []struct {
		name    string
		context SavedWorkflowDraftContext
		request ListWorkflowDraftsRequest
	}{
		{name: "tenant", context: changedSavedWorkflowDraftCursorContext(requestContext, "tenant"), request: baseRequest},
		{name: "workspace", context: changedSavedWorkflowDraftCursorContext(requestContext, "workspace"), request: baseRequest},
		{name: "application", context: changedSavedWorkflowDraftCursorContext(requestContext, "application"), request: baseRequest},
		{name: "owner", context: changedSavedWorkflowDraftCursorContext(requestContext, "owner"), request: baseRequest},
		{name: "lifecycle", context: requestContext, request: ListWorkflowDraftsRequest{Limit: 1, Cursor: first.NextCursor, LifecycleState: SavedWorkflowDraftLifecycleArchived}},
		{name: "limit", context: requestContext, request: ListWorkflowDraftsRequest{Limit: 2, Cursor: first.NextCursor}},
		{name: "name", context: requestContext, request: ListWorkflowDraftsRequest{Limit: 1, Cursor: first.NextCursor, NamePrefix: "Cursor"}},
		{name: "validation", context: requestContext, request: ListWorkflowDraftsRequest{Limit: 1, Cursor: first.NextCursor, ValidationState: SavedWorkflowDraftStatusValidForReview}},
		{name: "provenance", context: requestContext, request: ListWorkflowDraftsRequest{Limit: 1, Cursor: first.NextCursor, ProvenanceKind: SavedWorkflowDraftProvenanceUnversioned}},
	}
	for _, testCase := range driftCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := service.ListDrafts(testCase.context, testCase.request)
			if result.FailureCode != SavedWorkflowDraftFailureListCursorInvalid {
				t.Fatalf("cursor drift must fail closed: %#v", result)
			}
		})
	}

	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("unmarshal cursor: %v", err)
	}
	document["schema_version"] = "saved_workflow_draft_list_cursor.v999"
	driftedRaw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal drifted cursor: %v", err)
	}
	schemaDrift := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{
		Limit:  1,
		Cursor: base64.RawURLEncoding.EncodeToString(driftedRaw),
	})
	if schemaDrift.FailureCode != SavedWorkflowDraftFailureListCursorInvalid {
		t.Fatalf("cursor schema drift must fail closed: %#v", schemaDrift)
	}
	document["schema_version"] = savedWorkflowDraftListCursorSchemaVersion
	document["last_draft_id"] = "draft_cursor_tampered"
	tamperedRaw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal tampered cursor: %v", err)
	}
	tampered := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{
		Limit:  1,
		Cursor: base64.RawURLEncoding.EncodeToString(tamperedRaw),
	})
	if tampered.FailureCode != SavedWorkflowDraftFailureListCursorInvalid {
		t.Fatalf("cursor anchor tampering must fail closed: %#v", tampered)
	}
	document["last_draft_id"] = first.Summaries[0].DraftID
	document["unknown_field"] = "must fail closed"
	unknownRaw, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal cursor with unknown field: %v", err)
	}
	unknownField := service.ListDrafts(requestContext, ListWorkflowDraftsRequest{
		Limit:  1,
		Cursor: base64.RawURLEncoding.EncodeToString(unknownRaw),
	})
	if unknownField.FailureCode != SavedWorkflowDraftFailureListCursorInvalid {
		t.Fatalf("cursor structure drift must fail closed: %#v", unknownField)
	}

	filterFailures := []ListWorkflowDraftsRequest{
		{LifecycleState: "deleted"},
		{Limit: maxSavedWorkflowDraftListLimit + 1},
		{NamePrefix: strings.Repeat("萝", maxSavedWorkflowDraftNamePrefix+1)},
		{NamePrefix: string([]byte{0xff})},
		{ValidationState: "unknown"},
		{ProvenanceKind: "recursive"},
	}
	for _, request := range filterFailures {
		result := service.ListDrafts(requestContext, request)
		if result.FailureCode != SavedWorkflowDraftFailureListFilterInvalid {
			t.Fatalf("invalid filter must fail closed: request=%#v result=%#v", request, result)
		}
	}
}

func TestSavedWorkflowDraftArchiveSaveConcurrency(t *testing.T) {
	for iteration := 0; iteration < 32; iteration++ {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		service.now = fixedSavedWorkflowDraftClock()
		requestContext := savedWorkflowDraftTestContext()
		payload := validSavedWorkflowDraftPayload()
		saved := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
		if saved.FailureCode != "" || saved.Draft == nil {
			t.Fatalf("seed concurrent draft: %#v", saved)
		}
		payload.Name = "Concurrent save"

		start := make(chan struct{})
		var wait sync.WaitGroup
		wait.Add(2)
		var saveResult SavedWorkflowDraftResult
		var archiveResult SavedWorkflowDraftLifecycleTransitionResult
		go func() {
			defer wait.Done()
			<-start
			saveResult = service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
				ExpectedDraftVersion:     1,
				ExpectedLifecycleVersion: 1,
				Payload:                  payload,
			})
		}()
		go func() {
			defer wait.Done()
			<-start
			archiveResult = service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
				DraftID:                  payload.DraftID,
				ExpectedDraftVersion:     1,
				ExpectedLifecycleVersion: 1,
			})
		}()
		close(start)
		wait.Wait()

		saveSucceeded := saveResult.FailureCode == ""
		archiveSucceeded := archiveResult.FailureCode == ""
		if saveSucceeded == archiveSucceeded {
			t.Fatalf("exactly one concurrent mutation must succeed: save=%#v archive=%#v", saveResult, archiveResult)
		}
		if !saveSucceeded && saveResult.FailureCode != SavedWorkflowDraftFailureArchived {
			t.Fatalf("losing save returned unexpected failure: %#v", saveResult)
		}
		if !archiveSucceeded && archiveResult.FailureCode != SavedWorkflowDraftFailureVersionConflict {
			t.Fatalf("losing archive returned unexpected failure: %#v", archiveResult)
		}
		sideEffects := store.SideEffects()
		if hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
			t.Fatalf("concurrent lifecycle mutation unlocked runtime side effects: %#v", sideEffects)
		}
		if saveSucceeded && (sideEffects.DraftWriteCount != 2 || sideEffects.LifecycleTransitionCount != 0) {
			t.Fatalf("save winner side effects drifted: %#v", sideEffects)
		}
		if archiveSucceeded && (sideEffects.DraftWriteCount != 1 || sideEffects.LifecycleTransitionCount != 1) {
			t.Fatalf("archive winner side effects drifted: %#v", sideEffects)
		}
	}
}

func TestSavedWorkflowDraftArchiveRestoreConcurrency(t *testing.T) {
	for iteration := 0; iteration < 16; iteration++ {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		service.now = fixedSavedWorkflowDraftClock()
		requestContext := savedWorkflowDraftTestContext()
		payload := validSavedWorkflowDraftPayload()
		first := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
		if first.FailureCode != "" {
			t.Fatalf("seed first revision: %#v", first)
		}
		payload.Name = "Second revision"
		second := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
			ExpectedDraftVersion:     1,
			ExpectedLifecycleVersion: 1,
			Payload:                  payload,
		})
		if second.FailureCode != "" || second.Draft == nil {
			t.Fatalf("seed second revision: %#v", second)
		}

		start := make(chan struct{})
		var wait sync.WaitGroup
		wait.Add(2)
		var restoreResult SavedWorkflowDraftRevisionResult
		var archiveResult SavedWorkflowDraftLifecycleTransitionResult
		go func() {
			defer wait.Done()
			<-start
			restoreResult = service.RestoreDraftRevision(requestContext, RestoreSavedWorkflowDraftRevisionRequest{
				DraftID:                     payload.DraftID,
				SourceDraftVersion:          1,
				ExpectedCurrentDraftVersion: 2,
				ExpectedLifecycleVersion:    1,
			})
		}()
		go func() {
			defer wait.Done()
			<-start
			archiveResult = service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
				DraftID:                  payload.DraftID,
				ExpectedDraftVersion:     2,
				ExpectedLifecycleVersion: 1,
			})
		}()
		close(start)
		wait.Wait()

		restoreSucceeded := restoreResult.FailureCode == ""
		archiveSucceeded := archiveResult.FailureCode == ""
		if restoreSucceeded == archiveSucceeded {
			t.Fatalf("exactly one restore/archive mutation must succeed: restore=%#v archive=%#v", restoreResult, archiveResult)
		}
		if !restoreSucceeded && restoreResult.FailureCode != SavedWorkflowDraftFailureArchived {
			t.Fatalf("losing restore returned unexpected failure: %#v", restoreResult)
		}
		if !archiveSucceeded && archiveResult.FailureCode != SavedWorkflowDraftFailureVersionConflict {
			t.Fatalf("losing archive returned unexpected failure: %#v", archiveResult)
		}
		if sideEffects := store.SideEffects(); hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
			t.Fatalf("restore/archive race unlocked runtime side effects: %#v", sideEffects)
		}
	}
}

func TestArchivedSavedWorkflowDraftIsRejectedByDirectConsumers(t *testing.T) {
	store := newMemorySavedWorkflowDraftStore()
	service := newSavedWorkflowDraftService(store)
	service.now = fixedSavedWorkflowDraftClock()
	requestContext := savedWorkflowDraftTestContext()
	saved := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		Payload: validSavedWorkflowDraftPayload(),
	})
	if saved.FailureCode != "" || saved.Draft == nil {
		t.Fatalf("seed direct-consumer draft: %#v", saved)
	}
	archived := service.ArchiveDraft(requestContext, TransitionSavedWorkflowDraftLifecycleRequest{
		DraftID:                  saved.Draft.DraftID,
		ExpectedDraftVersion:     1,
		ExpectedLifecycleVersion: 1,
	})
	if archived.FailureCode != "" || archived.Draft == nil {
		t.Fatalf("archive direct-consumer draft: %#v", archived)
	}

	if _, failureCode, _ := buildWorkflowExecutionPlan(*archived.Draft, nil); failureCode != WorkflowRunFailureDraftNotEligible {
		t.Fatalf("executor accepted archived draft: %s", failureCode)
	}
	if _, failureCode := buildWorkflowRAGExecutionPlan(*archived.Draft, archived.Draft.DraftVersion); failureCode != WorkflowRAGFailureDraftIneligible {
		t.Fatalf("RAG execution accepted archived draft: %s", failureCode)
	}
	if err := validateWorkflowHTTPToolDraft(*archived.Draft, "node_tool", WorkflowHTTPToolDefinition{}); err == nil ||
		err.Error() != "Saved workflow draft is not active." {
		t.Fatalf("HTTP tool planning accepted archived draft: %v", err)
	}
	derivedPayload := validSavedWorkflowDraftPayload()
	derivedPayload.DraftID = "draft_archived_source_derivation"
	derivedPayload.AdditionalFields = map[string]any{
		savedWorkflowDraftDerivationAdditionalField: map[string]any{
			"version":              1,
			"source_kind":          savedWorkflowDraftDerivationSourceKind,
			"source_draft_id":      archived.Draft.DraftID,
			"source_draft_version": archived.Draft.DraftVersion,
		},
	}
	derived := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		Payload:                  derivedPayload,
		ExpectedLifecycleVersion: archived.Draft.LifecycleVersion,
	})
	if derived.FailureCode != SavedWorkflowDraftFailureArchived || derived.Draft != nil ||
		derived.CurrentDraftVersion != archived.Draft.DraftVersion ||
		derived.CurrentLifecycleVersion != archived.Draft.LifecycleVersion {
		t.Fatalf("derived first save accepted archived source: %#v", derived)
	}
	if sideEffects := store.SideEffects(); sideEffects.DraftWriteCount != 1 ||
		sideEffects.LifecycleTransitionCount != 1 ||
		hasSavedWorkflowDraftRuntimeSideEffect(sideEffects) {
		t.Fatalf("direct-consumer rejection changed runtime side effects: %#v", sideEffects)
	}
}

func saveWorkflowDraftLibraryFixture(
	t *testing.T,
	service savedWorkflowDraftService,
	requestContext SavedWorkflowDraftContext,
	payload SavedWorkflowDraftPayload,
) SavedWorkflowDraft {
	t.Helper()
	expectedLifecycleVersion := 0
	if _, derived := normalizeSavedWorkflowDraftDerivation(
		payload.AdditionalFields[savedWorkflowDraftDerivationAdditionalField],
		payload.DraftID,
	); derived {
		expectedLifecycleVersion = 1
	}
	result := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
		Payload:                  payload,
		ExpectedLifecycleVersion: expectedLifecycleVersion,
	})
	if result.FailureCode != "" || result.Draft == nil {
		t.Fatalf("save library fixture %s: %#v", payload.DraftID, result)
	}
	return *result.Draft
}

func changedSavedWorkflowDraftCursorContext(
	requestContext SavedWorkflowDraftContext,
	field string,
) SavedWorkflowDraftContext {
	changed := requestContext
	switch field {
	case "tenant":
		changed.TenantRef = "tenant_other"
	case "workspace":
		changed.WorkspaceID = "workspace_other"
	case "application":
		changed.ApplicationID = "application_other"
	case "owner":
		changed.OwnerSubjectRef = "subject_other"
	}
	return changed
}
