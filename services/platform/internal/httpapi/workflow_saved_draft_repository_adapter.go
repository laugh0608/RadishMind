package httpapi

import (
	"context"
	"sort"
	"strings"
)

const (
	savedWorkflowDraftRepositoryMigrationApplied     = "applied"
	savedWorkflowDraftRepositoryMigrationNotApplied  = "not_applied"
	savedWorkflowDraftRepositoryMigrationUnavailable = "unavailable"
)

type SavedWorkflowDraftRepositoryAdapter struct {
	queryExecutor   SavedWorkflowDraftRepositoryQueryExecutor
	schemaPreflight SavedWorkflowDraftRepositorySchemaPreflight
}

func NewSavedWorkflowDraftRepositoryAdapter(
	config SavedWorkflowDraftRepositoryAdapterConfig,
) SavedWorkflowDraftRepositoryAdapter {
	return SavedWorkflowDraftRepositoryAdapter{
		queryExecutor:   config.QueryExecutor,
		schemaPreflight: normalizedSavedWorkflowDraftRepositorySchemaPreflight(config.SchemaPreflight),
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) SaveWorkflowDraftRecord(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request SaveWorkflowDraftRecordRequest,
) SaveWorkflowDraftRecordResult {
	if ctx == nil {
		return SaveWorkflowDraftRecordResult{FailureCode: SavedWorkflowDraftFailureAuthContextMismatch}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failureCode := savedWorkflowDraftRepositoryActorFailure(actor, "workflow_drafts:write"); failureCode != "" {
		return SaveWorkflowDraftRecordResult{FailureCode: failureCode}
	}
	draft := sanitizeSavedWorkflowDraftRepositoryDraft(request.Draft)
	if failureCode := savedWorkflowDraftRepositoryDraftFailure(actor, draft); failureCode != "" {
		return SaveWorkflowDraftRecordResult{FailureCode: failureCode}
	}
	if failureCode := adapter.schemaPreflight.failureCodeFor(draft.SchemaVersion); failureCode != "" {
		return SaveWorkflowDraftRecordResult{FailureCode: failureCode}
	}
	if adapter.queryExecutor == nil {
		return SaveWorkflowDraftRecordResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}

	queryResult := adapter.queryExecutor.SaveWorkflowDraftRecord(ctx, savedWorkflowDraftRepositorySaveQuery{
		ActorContext:         actor,
		ExpectedDraftVersion: request.ExpectedDraftVersion,
		Record:               savedWorkflowDraftRepositoryRecordFromDraft(actor, draft, adapter.schemaPreflight),
	})
	if queryResult.FailureCode != "" {
		return SaveWorkflowDraftRecordResult{
			FailureCode:         queryResult.FailureCode,
			CurrentDraftVersion: queryResult.CurrentDraftVersion,
		}
	}
	return SaveWorkflowDraftRecordResult{
		Draft:               cloneSavedWorkflowDraftPointer(queryResult.Record.Draft),
		CurrentDraftVersion: queryResult.Record.Draft.DraftVersion,
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) ReadWorkflowDraftRecord(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	request ReadWorkflowDraftRecordRequest,
) ReadWorkflowDraftRecordResult {
	if ctx == nil {
		return ReadWorkflowDraftRecordResult{FailureCode: SavedWorkflowDraftFailureAuthContextMismatch}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failureCode := savedWorkflowDraftRepositoryActorFailure(actor, "workflow_drafts:read"); failureCode != "" {
		return ReadWorkflowDraftRecordResult{FailureCode: failureCode}
	}
	draftID := strings.TrimSpace(request.DraftID)
	if draftID == "" {
		return ReadWorkflowDraftRecordResult{FailureCode: SavedWorkflowDraftFailurePayloadInvalid}
	}
	if failureCode := adapter.schemaPreflight.failureCodeFor(savedWorkflowDraftSchemaVersion); failureCode != "" {
		return ReadWorkflowDraftRecordResult{FailureCode: failureCode}
	}
	if adapter.queryExecutor == nil {
		return ReadWorkflowDraftRecordResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}

	queryResult := adapter.queryExecutor.ReadWorkflowDraftRecord(ctx, savedWorkflowDraftRepositoryReadQuery{
		ActorContext: actor,
		DraftID:      draftID,
	})
	if queryResult.FailureCode != "" {
		return ReadWorkflowDraftRecordResult{
			FailureCode:         queryResult.FailureCode,
			CurrentDraftVersion: queryResult.CurrentDraftVersion,
		}
	}
	if failureCode := adapter.validateStoredRecord(actor, queryResult.Record); failureCode != "" {
		return ReadWorkflowDraftRecordResult{
			FailureCode:         failureCode,
			CurrentDraftVersion: queryResult.Record.Draft.DraftVersion,
		}
	}
	return ReadWorkflowDraftRecordResult{
		Draft:               cloneSavedWorkflowDraftPointer(queryResult.Record.Draft),
		CurrentDraftVersion: queryResult.Record.Draft.DraftVersion,
	}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) ListWorkflowDraftRecords(
	ctx context.Context,
	actor SavedWorkflowDraftRepositoryActorContext,
	_ ListWorkflowDraftRecordsRequest,
) ListWorkflowDraftRecordsResult {
	if ctx == nil {
		return ListWorkflowDraftRecordsResult{FailureCode: SavedWorkflowDraftFailureAuthContextMismatch}
	}
	actor = normalizeSavedWorkflowDraftRepositoryActorContext(actor)
	if failureCode := savedWorkflowDraftRepositoryActorFailure(actor, "workflow_drafts:read"); failureCode != "" {
		return ListWorkflowDraftRecordsResult{FailureCode: failureCode}
	}
	if failureCode := adapter.schemaPreflight.failureCodeFor(savedWorkflowDraftSchemaVersion); failureCode != "" {
		return ListWorkflowDraftRecordsResult{FailureCode: failureCode}
	}
	if adapter.queryExecutor == nil {
		return ListWorkflowDraftRecordsResult{FailureCode: SavedWorkflowDraftFailureStoreUnavailable}
	}

	queryResult := adapter.queryExecutor.ListWorkflowDraftRecords(ctx, savedWorkflowDraftRepositoryListQuery{
		ActorContext: actor,
	})
	if queryResult.FailureCode != "" {
		return ListWorkflowDraftRecordsResult{FailureCode: queryResult.FailureCode}
	}
	summaries := make([]SavedWorkflowDraftSummary, 0, len(queryResult.Records))
	for _, record := range queryResult.Records {
		if failureCode := adapter.validateStoredRecord(actor, record); failureCode != "" {
			return ListWorkflowDraftRecordsResult{FailureCode: failureCode}
		}
		summaries = append(summaries, savedWorkflowDraftSummaryFromDraft(record.Draft))
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].UpdatedAt == summaries[j].UpdatedAt {
			return summaries[i].DraftID < summaries[j].DraftID
		}
		return summaries[i].UpdatedAt > summaries[j].UpdatedAt
	})
	return ListWorkflowDraftRecordsResult{Summaries: summaries}
}

func (adapter SavedWorkflowDraftRepositoryAdapter) validateStoredRecord(
	actor SavedWorkflowDraftRepositoryActorContext,
	record SavedWorkflowDraftRepositoryStoredRecord,
) SavedWorkflowDraftFailureCode {
	if strings.TrimSpace(record.StoreSchemaVersion) != adapter.schemaPreflight.StoreSchemaVersion {
		return SavedWorkflowDraftFailureStoreSchemaVersionMismatch
	}
	if record.TenantRef != actor.TenantRef ||
		record.WorkspaceID != actor.WorkspaceID ||
		record.ApplicationID != actor.ApplicationID ||
		record.OwnerSubjectRef != actor.OwnerSubjectRef {
		return SavedWorkflowDraftFailureScopeDenied
	}
	if record.Draft.DraftID == "" ||
		record.DraftID != record.Draft.DraftID ||
		record.Draft.WorkspaceID != actor.WorkspaceID ||
		record.Draft.ApplicationID != actor.ApplicationID {
		return SavedWorkflowDraftFailureStoreContractMismatch
	}
	if record.Draft.SchemaVersion != savedWorkflowDraftSchemaVersion {
		return SavedWorkflowDraftFailureSchemaVersionUnsupported
	}
	return ""
}

func normalizeSavedWorkflowDraftRepositoryActorContext(
	actor SavedWorkflowDraftRepositoryActorContext,
) SavedWorkflowDraftRepositoryActorContext {
	actor.RequestID = strings.TrimSpace(actor.RequestID)
	actor.TenantRef = strings.TrimSpace(actor.TenantRef)
	actor.WorkspaceID = strings.TrimSpace(actor.WorkspaceID)
	actor.ApplicationID = strings.TrimSpace(actor.ApplicationID)
	actor.ActorSubjectRef = strings.TrimSpace(actor.ActorSubjectRef)
	actor.OwnerSubjectRef = strings.TrimSpace(actor.OwnerSubjectRef)
	actor.AuditRef = strings.TrimSpace(actor.AuditRef)
	actor.ScopeGrants = normalizedStringSet(actor.ScopeGrants)
	return actor
}

func savedWorkflowDraftRepositoryActorFailure(
	actor SavedWorkflowDraftRepositoryActorContext,
	requiredScope string,
) SavedWorkflowDraftFailureCode {
	if actor.ActorSubjectRef == "" {
		return SavedWorkflowDraftFailureIdentityContextMissing
	}
	if actor.TenantRef == "" {
		return SavedWorkflowDraftFailureTenantBindingMissing
	}
	if actor.WorkspaceID == "" || actor.ApplicationID == "" || actor.OwnerSubjectRef == "" {
		return SavedWorkflowDraftFailureWorkspaceMembershipDenied
	}
	if actor.RequestID == "" || actor.AuditRef == "" {
		return SavedWorkflowDraftFailureAuthContextMismatch
	}
	if !controlPlaneReadHasScope(actor.ScopeGrants, requiredScope) {
		return SavedWorkflowDraftFailureScopeGrantMissing
	}
	return ""
}

func savedWorkflowDraftRepositoryDraftFailure(
	actor SavedWorkflowDraftRepositoryActorContext,
	draft SavedWorkflowDraft,
) SavedWorkflowDraftFailureCode {
	if draft.DraftID == "" || draft.Name == "" {
		return SavedWorkflowDraftFailurePayloadInvalid
	}
	if draft.WorkspaceID != actor.WorkspaceID || draft.ApplicationID != actor.ApplicationID {
		return SavedWorkflowDraftFailureScopeDenied
	}
	if draft.SchemaVersion != savedWorkflowDraftSchemaVersion {
		return SavedWorkflowDraftFailureSchemaVersionUnsupported
	}
	return ""
}

func savedWorkflowDraftRepositoryRecordFromDraft(
	actor SavedWorkflowDraftRepositoryActorContext,
	draft SavedWorkflowDraft,
	preflight SavedWorkflowDraftRepositorySchemaPreflight,
) SavedWorkflowDraftRepositoryStoredRecord {
	return SavedWorkflowDraftRepositoryStoredRecord{
		TenantRef:          actor.TenantRef,
		WorkspaceID:        draft.WorkspaceID,
		ApplicationID:      draft.ApplicationID,
		DraftID:            draft.DraftID,
		OwnerSubjectRef:    actor.OwnerSubjectRef,
		StoreSchemaVersion: preflight.StoreSchemaVersion,
		Draft:              sanitizeSavedWorkflowDraftRepositoryDraft(draft),
	}
}

func sanitizeSavedWorkflowDraftRepositoryDraft(draft SavedWorkflowDraft) SavedWorkflowDraft {
	sanitized := cloneSavedWorkflowDraft(draft)
	sanitized.WorkspaceID = strings.TrimSpace(sanitized.WorkspaceID)
	sanitized.ApplicationID = strings.TrimSpace(sanitized.ApplicationID)
	sanitized.DraftID = strings.TrimSpace(sanitized.DraftID)
	sanitized.SchemaVersion = strings.TrimSpace(sanitized.SchemaVersion)
	sanitized.Name = strings.TrimSpace(sanitized.Name)
	sanitized.SampleOrUnsavedDraftStatus = "saved_draft_record"
	return sanitized
}

func normalizedSavedWorkflowDraftRepositorySchemaPreflight(
	preflight SavedWorkflowDraftRepositorySchemaPreflight,
) SavedWorkflowDraftRepositorySchemaPreflight {
	preflight.StoreSchemaVersion = strings.TrimSpace(preflight.StoreSchemaVersion)
	if preflight.StoreSchemaVersion == "" {
		preflight.StoreSchemaVersion = savedWorkflowDraftRepositoryStoreSchemaVersion
	}
	preflight.MigrationState = strings.TrimSpace(preflight.MigrationState)
	if preflight.MigrationState == "" {
		preflight.MigrationState = savedWorkflowDraftRepositoryMigrationApplied
	}
	return preflight
}

func (preflight SavedWorkflowDraftRepositorySchemaPreflight) failureCodeFor(
	payloadSchemaVersion string,
) SavedWorkflowDraftFailureCode {
	switch preflight.MigrationState {
	case savedWorkflowDraftRepositoryMigrationNotApplied:
		return SavedWorkflowDraftFailureSchemaMigrationNotApplied
	case savedWorkflowDraftRepositoryMigrationUnavailable:
		return SavedWorkflowDraftFailureStoreMigrationUnavailable
	}
	if preflight.StoreSchemaVersion != savedWorkflowDraftRepositoryStoreSchemaVersion {
		return SavedWorkflowDraftFailureStoreSchemaVersionMismatch
	}
	if strings.TrimSpace(payloadSchemaVersion) != savedWorkflowDraftSchemaVersion {
		return SavedWorkflowDraftFailureSchemaVersionUnsupported
	}
	return ""
}
