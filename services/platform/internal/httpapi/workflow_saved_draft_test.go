package httpapi

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSavedWorkflowDraftSaveReadValidateContracts(t *testing.T) {
	t.Run("save and read sanitized draft record", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		service.now = fixedSavedWorkflowDraftClock()
		context := savedWorkflowDraftTestContext()

		saveResult := service.SaveDraft(context, SaveWorkflowDraftRequest{
			ExpectedDraftVersion: 0,
			Payload:              validSavedWorkflowDraftPayload(),
		})

		if saveResult.FailureCode != "" {
			t.Fatalf("save should succeed: %#v", saveResult)
		}
		if saveResult.Draft == nil {
			t.Fatalf("save should return a draft")
		}
		if saveResult.Draft.DraftVersion != 1 || saveResult.Draft.DraftStatus != SavedWorkflowDraftStatusValidForReview {
			t.Fatalf("unexpected saved draft version/status: %#v", saveResult.Draft)
		}
		if !saveResult.ValidationSummary.ValidForReview {
			t.Fatalf("saved draft should be valid for review: %#v", saveResult.ValidationSummary)
		}
		if saveResult.Draft.SampleOrUnsavedDraftStatus != "saved_draft_record" {
			t.Fatalf("saved draft must be distinguishable from sample state: %#v", saveResult.Draft.SampleOrUnsavedDraftStatus)
		}
		if saveResult.Draft.RequestAuditMetadata.RequestID != context.RequestID {
			t.Fatalf("saved draft did not preserve audit metadata: %#v", saveResult.Draft.RequestAuditMetadata)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("unexpected side effects after save: %#v", got)
		}

		readResult := service.ReadDraft(context, ReadWorkflowDraftRequest{DraftID: saveResult.Draft.DraftID})

		if readResult.FailureCode != "" || readResult.Draft == nil {
			t.Fatalf("read should succeed: %#v", readResult)
		}
		if readResult.Draft.DraftID != saveResult.Draft.DraftID ||
			readResult.CurrentDraftVersion != saveResult.Draft.DraftVersion {
			t.Fatalf("read returned unexpected draft: %#v", readResult)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("read should not add side effects: %#v", got)
		}
	})

	t.Run("validate does not write or unlock execution", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)

		result := service.ValidateDraft(savedWorkflowDraftTestContext(), ValidateWorkflowDraftRequest{
			Payload: validSavedWorkflowDraftPayload(),
		})

		if result.FailureCode != "" {
			t.Fatalf("validate should succeed: %#v", result)
		}
		if result.ValidationSummary.ValidationState != SavedWorkflowDraftStatusValidForReview ||
			!result.ValidationSummary.ValidForReview {
			t.Fatalf("unexpected validation summary: %#v", result.ValidationSummary)
		}
		if result.Draft != nil {
			t.Fatalf("validate should not materialize a saved draft: %#v", result.Draft)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("validate should not write or execute: %#v", got)
		}
	})

	t.Run("contract invalid draft can be saved as review finding", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		payload := validSavedWorkflowDraftPayload()
		payload.OutputContract.RequiredFields = nil

		result := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{Payload: payload})

		if result.FailureCode != "" || result.Draft == nil {
			t.Fatalf("contract findings should be saved as invalid draft: %#v", result)
		}
		if result.Draft.DraftStatus != SavedWorkflowDraftStatusInvalidDraft {
			t.Fatalf("unexpected draft status: %#v", result.Draft.DraftStatus)
		}
		if !hasSavedWorkflowDraftFinding(result.ValidationSummary, SavedWorkflowDraftFailureContractInvalid) {
			t.Fatalf("missing contract finding: %#v", result.ValidationSummary)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("invalid draft save should only write the draft: %#v", got)
		}
	})

	t.Run("blocked capability can be saved but cannot unlock runtime", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		payload := validSavedWorkflowDraftPayload()
		payload.RequestedCapabilities = []string{"executor", "writeback", "replay"}

		result := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{Payload: payload})

		if result.FailureCode != "" || result.Draft == nil {
			t.Fatalf("blocked capability should be saved as review finding: %#v", result)
		}
		if result.Draft.DraftStatus != SavedWorkflowDraftStatusBlockedCapability {
			t.Fatalf("unexpected blocked draft status: %#v", result.Draft.DraftStatus)
		}
		if len(result.BlockedCapabilities) != 3 {
			t.Fatalf("unexpected blocked capability summary: %#v", result.BlockedCapabilities)
		}
		readResult := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{
			DraftID: result.Draft.DraftID,
		})
		if readResult.FailureCode != "" ||
			readResult.Draft == nil ||
			readResult.Draft.DraftStatus != SavedWorkflowDraftStatusBlockedCapability ||
			len(readResult.BlockedCapabilities) != 3 {
			t.Fatalf("read should preserve blocked capability review summary: %#v", readResult)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("blocked draft save should not execute runtime capabilities: %#v", got)
		}
	})

	t.Run("list returns sanitized summaries for current scope", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		service.now = fixedSavedWorkflowDraftClock()
		context := savedWorkflowDraftTestContext()
		firstPayload := validSavedWorkflowDraftPayload()
		secondPayload := validSavedWorkflowDraftPayload()
		secondPayload.DraftID = "draft_radishflow_copilot_saved_v2"
		secondPayload.Name = "RadishFlow copilot saved draft v2"
		outOfScopeDraft := savedWorkflowDraftFromPayload(validSavedWorkflowDraftPayload())
		outOfScopeDraft.DraftID = "draft_other_scope"
		outOfScopeDraft.ApplicationID = "app_other"

		if first := service.SaveDraft(context, SaveWorkflowDraftRequest{Payload: firstPayload}); first.FailureCode != "" {
			t.Fatalf("first save should succeed: %#v", first)
		}
		if second := service.SaveDraft(context, SaveWorkflowDraftRequest{Payload: secondPayload}); second.FailureCode != "" {
			t.Fatalf("second save should succeed: %#v", second)
		}
		if err := store.WriteDraft(outOfScopeDraft); err != nil {
			t.Fatalf("failed to seed out-of-scope draft: %v", err)
		}

		result := service.ListDrafts(context, ListWorkflowDraftsRequest{})

		if result.FailureCode != "" {
			t.Fatalf("list should succeed: %#v", result)
		}
		if len(result.Summaries) != 2 {
			t.Fatalf("list should return only current scope summaries: %#v", result.Summaries)
		}
		for _, summary := range result.Summaries {
			if summary.ApplicationID != context.ApplicationID ||
				summary.SampleOrUnsavedDraftStatus != "saved_draft_record" ||
				summary.NodeCount == 0 ||
				summary.EdgeCount == 0 {
				t.Fatalf("list summary drifted or exposed wrong scope: %#v", summary)
			}
		}
		if got := store.SideEffects(); got.DraftWriteCount != 3 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("list should not add side effects: %#v", got)
		}
	})
}

func TestSavedWorkflowDraftFailureSemantics(t *testing.T) {
	t.Run("version conflict rejects stale overwrite without partial write", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		context := savedWorkflowDraftTestContext()
		first := service.SaveDraft(context, SaveWorkflowDraftRequest{Payload: validSavedWorkflowDraftPayload()})
		if first.FailureCode != "" || first.Draft == nil {
			t.Fatalf("initial save failed: %#v", first)
		}
		updatePayload := validSavedWorkflowDraftPayload()
		updatePayload.Description = "updated by stale client"

		conflict := service.SaveDraft(context, SaveWorkflowDraftRequest{
			ExpectedDraftVersion: 0,
			Payload:              updatePayload,
		})

		if conflict.FailureCode != SavedWorkflowDraftFailureVersionConflict ||
			conflict.CurrentDraftVersion != first.Draft.DraftVersion {
			t.Fatalf("unexpected conflict result: %#v", conflict)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("conflict should not write or execute: %#v", got)
		}
	})

	t.Run("not found read does not return sample fallback", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)

		result := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{
			DraftID: "missing_draft",
		})

		if result.FailureCode != SavedWorkflowDraftFailureNotFound || result.Draft != nil {
			t.Fatalf("read should fail closed without sample fallback: %#v", result)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("not found read should not write or execute: %#v", got)
		}
	})

	t.Run("scope mismatch fails closed", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		saveResult := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{
			Payload: validSavedWorkflowDraftPayload(),
		})
		if saveResult.FailureCode != "" || saveResult.Draft == nil {
			t.Fatalf("save failed: %#v", saveResult)
		}
		wrongScope := savedWorkflowDraftTestContext()
		wrongScope.ApplicationID = "app_other"

		result := service.ReadDraft(wrongScope, ReadWorkflowDraftRequest{DraftID: saveResult.Draft.DraftID})

		if result.FailureCode != SavedWorkflowDraftFailureScopeDenied || result.Draft != nil {
			t.Fatalf("read should fail closed on scope mismatch: %#v", result)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 1 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("scope mismatch should not add side effects: %#v", got)
		}
	})

	t.Run("unsupported schema is rejected on save and reviewable on read", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		payload := validSavedWorkflowDraftPayload()
		payload.SchemaVersion = "saved_workflow_draft.v99"

		saveResult := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{Payload: payload})
		if saveResult.FailureCode != SavedWorkflowDraftFailureSchemaVersionUnsupported || saveResult.Draft != nil {
			t.Fatalf("save should reject unsupported schema: %#v", saveResult)
		}

		legacyDraft := savedWorkflowDraftFromPayload(validSavedWorkflowDraftPayload())
		legacyDraft.SchemaVersion = "saved_workflow_draft.v0"
		legacyDraft.DraftVersion = 7
		if err := store.WriteDraft(legacyDraft); err != nil {
			t.Fatalf("failed to seed legacy draft: %v", err)
		}

		readResult := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{DraftID: legacyDraft.DraftID})
		if readResult.FailureCode != SavedWorkflowDraftFailureSchemaVersionUnsupported ||
			readResult.Draft != nil ||
			readResult.ValidationSummary.ValidationState != SavedWorkflowDraftStatusSchemaUnsupported {
			t.Fatalf("read should return unsupported schema metadata only: %#v", readResult)
		}
	})

	t.Run("hard failures do not produce partial writes", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			mutate      func(SavedWorkflowDraftPayload) SavedWorkflowDraftPayload
			context     SavedWorkflowDraftContext
			failureCode SavedWorkflowDraftFailureCode
		}{
			{
				name: "forbidden sensitive field",
				mutate: func(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
					payload.AdditionalFields = map[string]any{"secret_value": "do-not-save"}
					return payload
				},
				context:     savedWorkflowDraftTestContext(),
				failureCode: SavedWorkflowDraftFailurePayloadInvalid,
			},
			{
				name: "graph endpoint missing",
				mutate: func(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
					payload.Edges[0].ToNodeID = "missing_node"
					return payload
				},
				context:     savedWorkflowDraftTestContext(),
				failureCode: SavedWorkflowDraftFailureGraphInvalid,
			},
			{
				name: "payload too large",
				mutate: func(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
					payload.Description = strings.Repeat("x", maxSavedWorkflowDraftTextLength+1)
					return payload
				},
				context:     savedWorkflowDraftTestContext(),
				failureCode: SavedWorkflowDraftFailurePayloadTooLarge,
			},
			{
				name: "scope denied",
				mutate: func(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
					payload.WorkspaceID = "workspace_other"
					return payload
				},
				context:     savedWorkflowDraftTestContext(),
				failureCode: SavedWorkflowDraftFailureScopeDenied,
			},
			{
				name: "write disabled",
				mutate: func(payload SavedWorkflowDraftPayload) SavedWorkflowDraftPayload {
					return payload
				},
				context: func() SavedWorkflowDraftContext {
					context := savedWorkflowDraftTestContext()
					context.WriteEnabled = false
					return context
				}(),
				failureCode: SavedWorkflowDraftFailureWriteDisabled,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				store := newMemorySavedWorkflowDraftStore()
				service := newSavedWorkflowDraftService(store)
				payload := tc.mutate(validSavedWorkflowDraftPayload())

				result := service.SaveDraft(tc.context, SaveWorkflowDraftRequest{Payload: payload})

				if result.FailureCode != tc.failureCode || result.Draft != nil {
					t.Fatalf("unexpected failure result: %#v", result)
				}
				readResult := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{DraftID: payload.DraftID})
				if readResult.FailureCode != SavedWorkflowDraftFailureNotFound || readResult.Draft != nil {
					t.Fatalf("hard failure should not leave a partial draft: %#v", readResult)
				}
				if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
					t.Fatalf("hard failure should not write or execute: %#v", got)
				}
			})
		}
	})

	t.Run("store unavailable fails closed", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		store.unavailable = true
		service := newSavedWorkflowDraftService(store)

		saveResult := service.SaveDraft(savedWorkflowDraftTestContext(), SaveWorkflowDraftRequest{
			Payload: validSavedWorkflowDraftPayload(),
		})
		if saveResult.FailureCode != SavedWorkflowDraftFailureStoreUnavailable || saveResult.Draft != nil {
			t.Fatalf("save should fail closed on store unavailable: %#v", saveResult)
		}
		readResult := service.ReadDraft(savedWorkflowDraftTestContext(), ReadWorkflowDraftRequest{
			DraftID: "draft_radishflow_copilot_saved_v1",
		})
		if readResult.FailureCode != SavedWorkflowDraftFailureStoreUnavailable || readResult.Draft != nil {
			t.Fatalf("read should fail closed on store unavailable: %#v", readResult)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("store unavailable should not write or execute: %#v", got)
		}
	})

	t.Run("list fails closed without sample fallback", func(t *testing.T) {
		store := newMemorySavedWorkflowDraftStore()
		service := newSavedWorkflowDraftService(store)
		context := savedWorkflowDraftTestContext()
		empty := service.ListDrafts(context, ListWorkflowDraftsRequest{})
		if empty.FailureCode != "" || len(empty.Summaries) != 0 {
			t.Fatalf("empty list should return no saved summaries without fallback: %#v", empty)
		}

		scopeDeniedContext := context
		scopeDeniedContext.ApplicationID = ""
		scopeDenied := service.ListDrafts(scopeDeniedContext, ListWorkflowDraftsRequest{})
		if scopeDenied.FailureCode != SavedWorkflowDraftFailureScopeDenied ||
			len(scopeDenied.Summaries) != 0 {
			t.Fatalf("list should fail closed on missing scope: %#v", scopeDenied)
		}

		store.unavailable = true
		unavailable := service.ListDrafts(context, ListWorkflowDraftsRequest{})
		if unavailable.FailureCode != SavedWorkflowDraftFailureStoreUnavailable ||
			len(unavailable.Summaries) != 0 {
			t.Fatalf("list should fail closed on store unavailable: %#v", unavailable)
		}
		if got := store.SideEffects(); got.DraftWriteCount != 0 || hasSavedWorkflowDraftRuntimeSideEffect(got) {
			t.Fatalf("list failure should not write or execute: %#v", got)
		}
	})
}

func TestSavedWorkflowDraftTypeBoundary(t *testing.T) {
	draftType := reflect.TypeOf(SavedWorkflowDraft{})
	for _, field := range []string{
		"DraftID",
		"WorkspaceID",
		"ApplicationID",
		"SourceDefinitionID",
		"BaseDefinitionVersion",
		"DraftVersion",
		"SchemaVersion",
		"DraftStatus",
		"Nodes",
		"Edges",
		"InputContract",
		"OutputContract",
		"ProviderRefs",
		"ToolRefs",
		"RAGRefs",
		"RequestedCapabilities",
		"ValidationSummary",
		"BlockedCapabilitySummary",
		"RequestAuditMetadata",
	} {
		if _, found := draftType.FieldByName(field); !found {
			t.Fatalf("SavedWorkflowDraft missing field %s", field)
		}
	}
	for _, forbidden := range []string{
		"SecretValue",
		"APIKeyValue",
		"Token",
		"OIDCToken",
		"UserClaims",
		"ToolExecutionResult",
		"MaterializedResult",
		"ConfirmationDecision",
		"RunInput",
		"RunOutput",
		"WritebackPayload",
		"ReplayState",
	} {
		if _, found := draftType.FieldByName(forbidden); found {
			t.Fatalf("SavedWorkflowDraft must not include %s", forbidden)
		}
	}
}

func savedWorkflowDraftTestContext() SavedWorkflowDraftContext {
	return SavedWorkflowDraftContext{
		RequestID:     "req_saved_workflow_draft_v1",
		WorkspaceID:   "workspace_demo",
		ApplicationID: "app_flow_copilot",
		ActorRef:      "subject_platform_ops",
		AuditRef:      "audit_saved_workflow_draft_v1",
		WriteEnabled:  true,
	}
}

func validSavedWorkflowDraftPayload() SavedWorkflowDraftPayload {
	return SavedWorkflowDraftPayload{
		DraftID:               "draft_radishflow_copilot_saved_v1",
		WorkspaceID:           "workspace_demo",
		ApplicationID:         "app_flow_copilot",
		SourceDefinitionID:    "wf_radishflow_copilot_latest",
		BaseDefinitionVersion: 3,
		SchemaVersion:         savedWorkflowDraftSchemaVersion,
		Name:                  "RadishFlow copilot saved draft",
		Description:           "Saved draft for reviewable workflow design.",
		Nodes: []SavedWorkflowDraftNode{
			{
				NodeID:            "node_context",
				NodeType:          "prompt",
				Label:             "Collect context",
				InputContractRef:  "contract_input",
				OutputContractRef: "contract_model_input",
				RiskLevel:         "low",
			},
			{
				NodeID:            "node_model",
				NodeType:          "llm",
				Label:             "Draft advisory response",
				InputContractRef:  "contract_model_input",
				OutputContractRef: "contract_output",
				ProviderRef:       "profile:radishmind-default-workflow",
				RiskLevel:         "low",
			},
			{
				NodeID:            "node_output",
				NodeType:          "output",
				Label:             "Return reviewable draft output",
				InputContractRef:  "contract_output",
				OutputContractRef: "contract_output",
				RiskLevel:         "low",
			},
		},
		Edges: []SavedWorkflowDraftEdge{
			{
				EdgeID:     "edge_context_model",
				FromNodeID: "node_context",
				ToNodeID:   "node_model",
			},
			{
				EdgeID:     "edge_model_output",
				FromNodeID: "node_model",
				ToNodeID:   "node_output",
			},
		},
		InputContract: SavedWorkflowDraftContract{
			ContractID:     "contract_input",
			RequiredFields: []string{"workspace_id", "application_id", "selection_summary"},
			Summary:        "Workspace-scoped input contract.",
		},
		OutputContract: SavedWorkflowDraftContract{
			ContractID:     "contract_output",
			RequiredFields: []string{"answer_summary", "risk_summary", "audit_refs"},
			Summary:        "Reviewable advisory output contract.",
		},
		ProviderRefs: []string{"profile:radishmind-default-workflow"},
		ToolRefs:     []string{"tool:radishflow-preview-readonly"},
		RAGRefs:      []string{"rag:radishflow-docs"},
	}
}

func savedWorkflowDraftFromPayload(payload SavedWorkflowDraftPayload) SavedWorkflowDraft {
	return SavedWorkflowDraft{
		DraftID:               payload.DraftID,
		WorkspaceID:           payload.WorkspaceID,
		ApplicationID:         payload.ApplicationID,
		SourceDefinitionID:    payload.SourceDefinitionID,
		BaseDefinitionVersion: payload.BaseDefinitionVersion,
		DraftVersion:          1,
		SchemaVersion:         payload.SchemaVersion,
		DraftStatus:           SavedWorkflowDraftStatusValidForReview,
		CreatedAt:             "2026-06-14T00:00:00Z",
		UpdatedAt:             "2026-06-14T00:00:00Z",
		CreatedByActorRef:     "subject_platform_ops",
		UpdatedByActorRef:     "subject_platform_ops",
		Name:                  payload.Name,
		Description:           payload.Description,
		Nodes:                 payload.Nodes,
		Edges:                 payload.Edges,
		InputContract:         payload.InputContract,
		OutputContract:        payload.OutputContract,
		ProviderRefs:          payload.ProviderRefs,
		ToolRefs:              payload.ToolRefs,
		RAGRefs:               payload.RAGRefs,
		RequestedCapabilities: payload.RequestedCapabilities,
	}
}

func fixedSavedWorkflowDraftClock() func() time.Time {
	return func() time.Time {
		return time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	}
}

func hasSavedWorkflowDraftFinding(
	summary SavedWorkflowDraftValidationSummary,
	code SavedWorkflowDraftFailureCode,
) bool {
	for _, finding := range summary.Findings {
		if finding.Code == code {
			return true
		}
	}
	return false
}

func hasSavedWorkflowDraftRuntimeSideEffect(sideEffects SavedWorkflowDraftSideEffects) bool {
	return sideEffects.ExecutorCallCount != 0 ||
		sideEffects.ConfirmationCallCount != 0 ||
		sideEffects.BusinessWritebackCount != 0 ||
		sideEffects.ReplayCallCount != 0 ||
		sideEffects.MaterializedResultReads != 0 ||
		sideEffects.ExternalRepositoryWrites != 0
}
