package httpapi

import (
	"context"
	"database/sql"
	"time"
)

type sqliteWorkflowTemplateCatalogRepository struct{ database *sql.DB }

func newSQLiteWorkflowTemplateCatalogRepository(database *sql.DB) *sqliteWorkflowTemplateCatalogRepository {
	return &sqliteWorkflowTemplateCatalogRepository{database: database}
}

type sqliteWorkflowTemplateQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) CreateCandidate(ctx WorkflowTemplateCatalogContext, candidate WorkflowTemplateCandidate, now time.Time) (WorkflowTemplateCandidate, error) {
	var output WorkflowTemplateCandidate
	err := repository.mutate(ctx, func(connection *sql.Conn, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		created, err := store.CreateCandidate(ctx, candidate, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(created)
		if err != nil {
			return err
		}
		createdAt, err := workflowTemplateUnixNano(created.CreatedAt)
		if err != nil {
			return err
		}
		if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_candidates
(tenant_ref,workspace_id,candidate_id,template_id,candidate_state,review_version,source_application_id,source_owner_subject_ref,source_definition_id,source_definition_version,source_definition_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_candidate_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, created.CandidateID, created.TemplateID, created.State, created.ReviewVersion, created.SourceApplicationID, created.SourceOwnerSubjectRef, created.SourceDefinitionID, created.SourceDefinitionVersion, created.SourceDefinitionDigest, createdAt, createdAt, string(payload)); err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if err = insertSQLiteWorkflowTemplateNewAudits(connection, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = created
		return nil
	})
	return output, err
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ReviewCandidate(ctx WorkflowTemplateCatalogContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowTemplateCandidate, *WorkflowTemplateVersion, error) {
	var output WorkflowTemplateCandidate
	var outputVersion *WorkflowTemplateVersion
	err := repository.mutate(ctx, func(connection *sql.Conn, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		updated, version, err := store.ReviewCandidate(ctx, candidateID, expected, decision, reason, sourceDigest, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(updated)
		if err != nil {
			return err
		}
		updatedAt, err := workflowTemplateUnixNano(updated.UpdatedAt)
		if err != nil {
			return err
		}
		result, err := connection.ExecContext(workflowTemplateRequestContext(ctx), `UPDATE workflow_template_candidates
SET candidate_state=?,review_version=?,updated_at_unix_nano=?,sanitized_candidate_payload=?
WHERE tenant_ref=? AND workspace_id=? AND candidate_id=? AND review_version=? AND candidate_state='pending'`, updated.State, updated.ReviewVersion, updatedAt, string(payload), ctx.TenantRef, ctx.WorkspaceID, candidateID, expected)
		if err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return errWorkflowTemplateCandidateConflict
		}
		review := updated.Decisions[len(updated.Decisions)-1]
		reviewPayload, err := encodeWorkflowTemplateRecord(review)
		if err != nil {
			return err
		}
		decidedAt, err := workflowTemplateUnixNano(review.DecidedAt)
		if err != nil {
			return err
		}
		if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_decisions
(tenant_ref,workspace_id,candidate_id,review_version,decision,decided_at_unix_nano,sanitized_decision_payload)
VALUES (?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, candidateID, review.ReviewVersion, review.Decision, decidedAt, string(reviewPayload)); err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
		}
		if version != nil {
			lineageKey := workflowTemplateScopeKey(ctx, version.TemplateID)
			versionPayload, encodeErr := encodeWorkflowTemplateRecord(*version)
			if encodeErr != nil {
				return encodeErr
			}
			createdAt, timeErr := workflowTemplateUnixNano(version.CreatedAt)
			if timeErr != nil {
				return timeErr
			}
			if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_versions
(tenant_ref,workspace_id,template_id,template_version,template_digest,candidate_id,candidate_review_version,source_definition_id,source_definition_version,source_definition_digest,created_at_unix_nano,sanitized_version_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, version.TemplateID, version.Version, version.TemplateDigest, version.CandidateID, version.CandidateReviewVersion, version.SourceDefinitionID, version.SourceDefinitionVersion, version.SourceDefinitionDigest, createdAt, string(versionPayload)); err != nil {
				return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
			}
			if version.Version == 1 {
				lineage := store.lineages[lineageKey]
				lineagePayload, encodeErr := encodeWorkflowTemplateRecord(lineage)
				if encodeErr != nil {
					return encodeErr
				}
				lineageCreatedAt, timeErr := workflowTemplateUnixNano(lineage.CreatedAt)
				if timeErr != nil {
					return timeErr
				}
				if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_lineages
(tenant_ref,workspace_id,template_id,pointer_version,lifecycle,listed_version,listed_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_lineage_payload)
VALUES (?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, lineage.TemplateID, lineage.PointerVersion, lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest, lineageCreatedAt, lineageCreatedAt, string(lineagePayload)); err != nil {
					return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateCandidateConflict)
				}
			}
			outputVersion = cloneWorkflowTemplateVersionPointer(*version)
		}
		if err = insertSQLiteWorkflowTemplateNewAudits(connection, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = updated
		return nil
	})
	return output, outputVersion, err
}

func (repository *sqliteWorkflowTemplateCatalogRepository) DecideListing(ctx WorkflowTemplateCatalogContext, templateID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowTemplateLineage, error) {
	var output WorkflowTemplateLineage
	err := repository.mutate(ctx, func(connection *sql.Conn, store *memoryWorkflowTemplateCatalogRepository) error {
		beforeAuditCount := len(store.audits[workflowTemplateScopeKey(ctx, "audits")])
		lineage, err := store.DecideListing(ctx, templateID, expected, decision, version, reason, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowTemplateRecord(lineage)
		if err != nil {
			return err
		}
		updatedAt, err := workflowTemplateUnixNano(lineage.UpdatedAt)
		if err != nil {
			return err
		}
		result, err := connection.ExecContext(workflowTemplateRequestContext(ctx), `UPDATE workflow_template_lineages
SET pointer_version=?,lifecycle=?,listed_version=?,listed_digest=?,updated_at_unix_nano=?,sanitized_lineage_payload=?
WHERE tenant_ref=? AND workspace_id=? AND template_id=? AND pointer_version=?`, lineage.PointerVersion, lineage.Lifecycle, lineage.ListedVersion, lineage.ListedDigest, updatedAt, string(payload), ctx.TenantRef, ctx.WorkspaceID, templateID, expected)
		if err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplatePointerConflict)
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return errWorkflowTemplatePointerConflict
		}
		event := lineage.Events[len(lineage.Events)-1]
		eventPayload, err := encodeWorkflowTemplateRecord(event)
		if err != nil {
			return err
		}
		occurredAt, err := workflowTemplateUnixNano(event.CreatedAt)
		if err != nil {
			return err
		}
		if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_listing_events
(tenant_ref,workspace_id,template_id,event_id,after_pointer_version,decision,occurred_at_unix_nano,sanitized_event_payload)
VALUES (?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, templateID, event.EventID, event.AfterPointerVersion, event.Decision, occurredAt, string(eventPayload)); err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplatePointerConflict)
		}
		if err = insertSQLiteWorkflowTemplateNewAudits(connection, ctx, store, beforeAuditCount); err != nil {
			return err
		}
		output = lineage
		return nil
	})
	return output, err
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ReadCandidate(ctx WorkflowTemplateCatalogContext, candidateID string) (WorkflowTemplateCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateCandidate{}, err
	}
	defer done()
	return store.ReadCandidate(ctx, candidateID)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ListCandidates(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListCandidates(ctx)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ReadLineage(ctx WorkflowTemplateCatalogContext, templateID string) (WorkflowTemplateLineage, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateLineage{}, err
	}
	defer done()
	return store.ReadLineage(ctx, templateID)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ListLineages(ctx WorkflowTemplateCatalogContext) ([]WorkflowTemplateLineage, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListLineages(ctx)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ReadVersion(ctx WorkflowTemplateCatalogContext, templateID string, version int) (WorkflowTemplateVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowTemplateVersion{}, err
	}
	defer done()
	return store.ReadVersion(ctx, templateID, version)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) ListVersions(ctx WorkflowTemplateCatalogContext, templateID string) ([]WorkflowTemplateVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListVersions(ctx, templateID)
}

func (repository *sqliteWorkflowTemplateCatalogRepository) mutate(ctx WorkflowTemplateCatalogContext, operation func(*sql.Conn, *memoryWorkflowTemplateCatalogRepository) error) error {
	if repository == nil || repository.database == nil || !validWorkflowTemplateContext(ctx) {
		return errWorkflowTemplateStoreUnavailable
	}
	requestContext := workflowTemplateRequestContext(ctx)
	connection, err := repository.database.Conn(requestContext)
	if err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	defer connection.Close()
	if _, err = connection.ExecContext(requestContext, "BEGIN IMMEDIATE"); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	store, err := loadSQLiteWorkflowTemplateCatalogStore(ctx, connection)
	if err != nil {
		return err
	}
	if err = operation(connection, store); err != nil {
		return err
	}
	if _, err = connection.ExecContext(requestContext, "COMMIT"); err != nil {
		return errWorkflowTemplateStoreUnavailable
	}
	committed = true
	return nil
}

func (repository *sqliteWorkflowTemplateCatalogRepository) readStore(ctx WorkflowTemplateCatalogContext) (*memoryWorkflowTemplateCatalogRepository, func(), error) {
	if repository == nil || repository.database == nil || !validWorkflowTemplateContext(ctx) {
		return nil, func() {}, errWorkflowTemplateStoreUnavailable
	}
	tx, err := repository.database.BeginTx(workflowTemplateRequestContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, func() {}, errWorkflowTemplateStoreUnavailable
	}
	store, err := loadSQLiteWorkflowTemplateCatalogStore(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, func() {}, err
	}
	return store, func() { _ = tx.Rollback() }, nil
}

func loadSQLiteWorkflowTemplateCatalogStore(ctx WorkflowTemplateCatalogContext, query sqliteWorkflowTemplateQueryer) (*memoryWorkflowTemplateCatalogRepository, error) {
	store := newMemoryWorkflowTemplateCatalogRepository()
	requestContext := workflowTemplateRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID}
	rows, err := query.QueryContext(requestContext, `SELECT candidate_id,template_id,candidate_state,review_version,source_application_id,source_owner_subject_ref,source_definition_id,source_definition_version,source_definition_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_candidate_payload
FROM workflow_template_candidates WHERE tenant_ref=? AND workspace_id=? ORDER BY created_at_unix_nano,candidate_id`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var candidateID, templateID, state, applicationID, owner, definitionID, definitionDigest string
		var reviewVersion, definitionVersion int
		var createdAt, updatedAt int64
		var payload []byte
		if rows.Scan(&candidateID, &templateID, &state, &reviewVersion, &applicationID, &owner, &definitionID, &definitionVersion, &definitionDigest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateCandidateRecord(store, ctx, payload)
		decodedCreated, _ := workflowTemplateUnixNano(value.CreatedAt)
		decodedUpdated, _ := workflowTemplateUnixNano(value.UpdatedAt)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.CandidateID == candidateID, value.TemplateID == templateID, value.State == state, value.ReviewVersion == reviewVersion, value.SourceApplicationID == applicationID, value.SourceOwnerSubjectRef == owner, value.SourceDefinitionID == definitionID, value.SourceDefinitionVersion == definitionVersion, value.SourceDefinitionDigest == definitionDigest, decodedCreated == createdAt, decodedUpdated == updatedAt) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows, err = query.QueryContext(requestContext, `SELECT template_id,template_version,template_digest,candidate_id,candidate_review_version,source_definition_id,source_definition_version,source_definition_digest,created_at_unix_nano,sanitized_version_payload
FROM workflow_template_versions WHERE tenant_ref=? AND workspace_id=? ORDER BY template_id,template_version`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var templateID, digest, candidateID, definitionID, definitionDigest string
		var version, reviewVersion, definitionVersion int
		var createdAt int64
		var payload []byte
		if rows.Scan(&templateID, &version, &digest, &candidateID, &reviewVersion, &definitionID, &definitionVersion, &definitionDigest, &createdAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateVersionRecord(store, ctx, payload)
		decodedAt, _ := workflowTemplateUnixNano(value.CreatedAt)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.TemplateID == templateID, value.Version == version, value.TemplateDigest == digest, value.CandidateID == candidateID, value.CandidateReviewVersion == reviewVersion, value.SourceDefinitionID == definitionID, value.SourceDefinitionVersion == definitionVersion, value.SourceDefinitionDigest == definitionDigest, decodedAt == createdAt) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows, err = query.QueryContext(requestContext, `SELECT template_id,pointer_version,lifecycle,listed_version,listed_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_lineage_payload
FROM workflow_template_lineages WHERE tenant_ref=? AND workspace_id=? ORDER BY template_id`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var templateID, lifecycle, digest string
		var pointerVersion, listedVersion int
		var createdAt, updatedAt int64
		var payload []byte
		if rows.Scan(&templateID, &pointerVersion, &lifecycle, &listedVersion, &digest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateLineageRecord(store, ctx, payload)
		decodedCreated, _ := workflowTemplateUnixNano(value.CreatedAt)
		decodedUpdated, _ := workflowTemplateUnixNano(value.UpdatedAt)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.TemplateID == templateID, value.PointerVersion == pointerVersion, value.Lifecycle == lifecycle, value.ListedVersion == listedVersion, value.ListedDigest == digest, decodedCreated == createdAt, decodedUpdated == updatedAt) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	rows, err = query.QueryContext(requestContext, `SELECT audit_id,audit_sequence,resource_kind,resource_id,action,occurred_at_unix_nano,sanitized_audit_payload
FROM workflow_template_audits WHERE tenant_ref=? AND workspace_id=? ORDER BY audit_sequence`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var auditID, kind, resourceID, action string
		var sequence int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&auditID, &sequence, &kind, &resourceID, &action, &occurredAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		value, loadErr := loadWorkflowTemplateAuditRecord(store, ctx, sequence, payload)
		decodedAt, _ := workflowTemplateUnixNano(value.CreatedAt)
		if loadErr != nil || workflowTemplateStoredFieldMismatch(value.AuditID == auditID, value.ResourceKind == kind, value.ResourceID == resourceID, value.Action == action, decodedAt == occurredAt) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	decisions := map[string]int{}
	rows, err = query.QueryContext(requestContext, `SELECT candidate_id,review_version,decision,decided_at_unix_nano,sanitized_decision_payload
FROM workflow_template_decisions WHERE tenant_ref=? AND workspace_id=? ORDER BY candidate_id,review_version`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var candidateID, decision string
		var reviewVersion int
		var decidedAt int64
		var payload []byte
		if rows.Scan(&candidateID, &reviewVersion, &decision, &decidedAt, &payload) != nil || validateWorkflowTemplateDecisionEvidence(store, ctx, candidateID, reviewVersion, payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		var decoded WorkflowTemplateReviewDecision
		_ = decodeWorkflowTemplateRecord(payload, &decoded)
		decodedAt, _ := workflowTemplateUnixNano(decoded.DecidedAt)
		if decoded.Decision != decision || decodedAt != decidedAt {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		decisions[candidateID]++
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	events := map[string]int{}
	rows, err = query.QueryContext(requestContext, `SELECT template_id,event_id,after_pointer_version,decision,occurred_at_unix_nano,sanitized_event_payload
FROM workflow_template_listing_events WHERE tenant_ref=? AND workspace_id=? ORDER BY template_id,after_pointer_version`, scope...)
	if err != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	for rows.Next() {
		var templateID, eventID, decision string
		var pointerVersion int
		var occurredAt int64
		var payload []byte
		if rows.Scan(&templateID, &eventID, &pointerVersion, &decision, &occurredAt, &payload) != nil || validateWorkflowTemplateListingEventEvidence(store, ctx, templateID, pointerVersion, payload) != nil {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		var decoded WorkflowTemplateListingEvent
		_ = decodeWorkflowTemplateRecord(payload, &decoded)
		decodedAt, _ := workflowTemplateUnixNano(decoded.CreatedAt)
		if decoded.EventID != eventID || decoded.Decision != decision || decodedAt != occurredAt {
			rows.Close()
			return nil, errWorkflowTemplateStoreUnavailable
		}
		events[templateID]++
	}
	if rows.Close() != nil || rows.Err() != nil || validateWorkflowTemplateEvidenceCounts(store, ctx, decisions, events) != nil {
		return nil, errWorkflowTemplateStoreUnavailable
	}
	return store, nil
}

func insertSQLiteWorkflowTemplateNewAudits(connection *sql.Conn, ctx WorkflowTemplateCatalogContext, store *memoryWorkflowTemplateCatalogRepository, before int) error {
	values := store.audits[workflowTemplateScopeKey(ctx, "audits")]
	for index := before; index < len(values); index++ {
		value := values[index]
		payload, err := encodeWorkflowTemplateRecord(value)
		if err != nil {
			return err
		}
		occurredAt, err := workflowTemplateUnixNano(value.CreatedAt)
		if err != nil {
			return err
		}
		sequence, err := workflowTemplateAuditSequence(value)
		if err != nil || sequence != index+1 {
			return errWorkflowTemplateStoreUnavailable
		}
		if _, err = connection.ExecContext(workflowTemplateRequestContext(ctx), `INSERT INTO workflow_template_audits
(tenant_ref,workspace_id,audit_id,audit_sequence,resource_kind,resource_id,action,occurred_at_unix_nano,sanitized_audit_payload)
VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, value.AuditID, sequence, value.ResourceKind, value.ResourceID, value.Action, occurredAt, string(payload)); err != nil {
			return sqliteWorkflowTemplateMutationError(err, errWorkflowTemplateStoreUnavailable)
		}
	}
	return nil
}

func sqliteWorkflowTemplateMutationError(err error, conflict error) error {
	if err == nil {
		return nil
	}
	if stringsContainsAny(err.Error(), "UNIQUE constraint failed", "database is locked") && conflict != nil {
		return conflict
	}
	return errWorkflowTemplateStoreUnavailable
}

var _ workflowTemplateCatalogRepository = (*sqliteWorkflowTemplateCatalogRepository)(nil)
