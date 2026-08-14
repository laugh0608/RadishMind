package httpapi

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSavedWorkflowDraftRevisionHistoryAndRestore(t *testing.T) {
	testCases := []struct {
		name     string
		newStore func(*testing.T) savedWorkflowDraftStore
	}{
		{
			name: "memory_dev",
			newStore: func(*testing.T) savedWorkflowDraftStore {
				return newMemorySavedWorkflowDraftStore()
			},
		},
		{
			name: "sqlite_dev",
			newStore: func(t *testing.T) savedWorkflowDraftStore {
				runtime := openSavedWorkflowDraftSQLiteRuntime(
					t,
					filepath.Join(t.TempDir(), "revision-history.db"),
				)
				return newSQLiteSavedWorkflowDraftStore(runtime.DB())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store := testCase.newStore(t)
			service := newSavedWorkflowDraftService(store)
			tick := 0
			service.now = func() time.Time {
				tick++
				return time.Date(2026, 7, 27, 13, 0, tick, 0, time.UTC)
			}
			requestContext := savedWorkflowDraftSQLiteContext()
			payload := validSavedWorkflowDraftPayload()
			payload.Name = "Revision one"

			first := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{Payload: payload})
			if first.FailureCode != "" || first.Draft == nil || first.Draft.DraftVersion != 1 {
				t.Fatalf("save revision one: %#v", first)
			}
			payload.Name = "Revision two"
			second := service.SaveDraft(requestContext, SaveWorkflowDraftRequest{
				ExpectedDraftVersion:     1,
				ExpectedLifecycleVersion: 1,
				Payload:                  payload,
			})
			if second.FailureCode != "" || second.Draft == nil || second.Draft.DraftVersion != 2 {
				t.Fatalf("save revision two: %#v", second)
			}

			firstPage := service.ListDraftRevisions(requestContext, ListSavedWorkflowDraftRevisionsRequest{
				DraftID: payload.DraftID,
				Limit:   1,
			})
			if firstPage.FailureCode != "" || len(firstPage.Revisions) != 1 ||
				firstPage.Revisions[0].DraftVersion != 2 || !firstPage.HasMore ||
				firstPage.NextCursor == "" {
				t.Fatalf("list first revision page: %#v", firstPage)
			}
			secondPage := service.ListDraftRevisions(requestContext, ListSavedWorkflowDraftRevisionsRequest{
				DraftID: payload.DraftID,
				Limit:   1,
				Cursor:  firstPage.NextCursor,
			})
			if secondPage.FailureCode != "" || len(secondPage.Revisions) != 1 ||
				secondPage.Revisions[0].DraftVersion != 1 || secondPage.HasMore {
				t.Fatalf("list second revision page: %#v", secondPage)
			}

			historical := service.ReadDraftRevision(
				requestContext,
				ReadSavedWorkflowDraftRevisionRequest{
					DraftID:      payload.DraftID,
					DraftVersion: 1,
				},
			)
			if historical.FailureCode != "" || historical.Revision == nil ||
				historical.Revision.Draft.Name != "Revision one" ||
				historical.Revision.RevisionKind != SavedWorkflowDraftRevisionKindSaved {
				t.Fatalf("read historical revision: %#v", historical)
			}

			restored := service.RestoreDraftRevision(
				requestContext,
				RestoreSavedWorkflowDraftRevisionRequest{
					DraftID:                     payload.DraftID,
					SourceDraftVersion:          1,
					ExpectedCurrentDraftVersion: 2,
					ExpectedLifecycleVersion:    1,
				},
			)
			if restored.FailureCode != "" || restored.Draft == nil ||
				restored.Draft.DraftVersion != 3 || restored.Draft.Name != "Revision one" {
				t.Fatalf("restore historical revision: %#v", restored)
			}
			restoredRevision := service.ReadDraftRevision(
				requestContext,
				ReadSavedWorkflowDraftRevisionRequest{
					DraftID:      payload.DraftID,
					DraftVersion: 3,
				},
			)
			if restoredRevision.FailureCode != "" || restoredRevision.Revision == nil ||
				restoredRevision.Revision.RevisionKind != SavedWorkflowDraftRevisionKindRestored ||
				restoredRevision.Revision.RestoredFromVersion != 1 {
				t.Fatalf("restored revision metadata drifted: %#v", restoredRevision)
			}
			unchanged := service.ReadDraftRevision(
				requestContext,
				ReadSavedWorkflowDraftRevisionRequest{
					DraftID:      payload.DraftID,
					DraftVersion: 2,
				},
			)
			if unchanged.FailureCode != "" || unchanged.Revision == nil ||
				unchanged.Revision.Draft.Name != "Revision two" {
				t.Fatalf("restore mutated an existing revision: %#v", unchanged)
			}

			stale := service.RestoreDraftRevision(
				requestContext,
				RestoreSavedWorkflowDraftRevisionRequest{
					DraftID:                     payload.DraftID,
					SourceDraftVersion:          2,
					ExpectedCurrentDraftVersion: 2,
					ExpectedLifecycleVersion:    1,
				},
			)
			if stale.FailureCode != SavedWorkflowDraftFailureVersionConflict ||
				stale.CurrentDraftVersion != 3 || stale.Draft != nil {
				t.Fatalf("stale restore did not fail closed: %#v", stale)
			}
			invalidCursor := service.ListDraftRevisions(
				requestContext,
				ListSavedWorkflowDraftRevisionsRequest{
					DraftID: payload.DraftID,
					Limit:   1,
					Cursor:  "not-a-cursor",
				},
			)
			if invalidCursor.FailureCode != SavedWorkflowDraftFailureRevisionCursor {
				t.Fatalf("invalid cursor did not fail closed: %#v", invalidCursor)
			}
		})
	}
}
