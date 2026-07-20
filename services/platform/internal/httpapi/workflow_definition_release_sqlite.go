package httpapi

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"time"
)

type sqliteWorkflowDefinitionReleaseRepository struct{ database *sql.DB }

func newSQLiteWorkflowDefinitionReleaseRepository(database *sql.DB) *sqliteWorkflowDefinitionReleaseRepository {
	return &sqliteWorkflowDefinitionReleaseRepository{database: database}
}

type sqliteWorkflowDefinitionQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) CreateCandidate(ctx WorkflowDefinitionReleaseContext, candidateID, definitionID string, draft SavedWorkflowDraft, now time.Time) (WorkflowDefinitionReleaseCandidate, error) {
	var output WorkflowDefinitionReleaseCandidate
	err := repository.mutate(ctx, func(connection *sql.Conn, store *workflowDefinitionReleaseStore) error {
		candidate, err := store.CreateCandidate(ctx, candidateID, definitionID, draft, now)
		if err != nil {
			return err
		}
		candidatePayload, err := encodeWorkflowDefinitionRecord(candidate)
		if err != nil {
			return err
		}
		createdAt, err := workflowDefinitionUnixNano(candidate.CreatedAt)
		if err != nil {
			return errWorkflowDefinitionStore
		}
		if _, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_candidates
(tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,definition_id,candidate_state,review_version,source_draft_id,source_draft_version,source_draft_digest,definition_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_candidate_payload)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidate.CandidateID, candidate.DefinitionID, candidate.State, candidate.ReviewVersion, candidate.SourceDraftID, candidate.SourceDraftVersion, candidate.SourceDraftDigest, candidate.DefinitionDigest, createdAt, createdAt, string(candidatePayload)); err != nil {
			return sqliteWorkflowDefinitionMutationError(err)
		}
		audit := store.audits[workflowDefinitionScopeKey(ctx, "candidate")][0]
		if err = insertSQLiteWorkflowDefinitionAudit(connection, ctx, audit); err != nil {
			return err
		}
		output = candidate
		return nil
	})
	return output, err
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) Review(ctx WorkflowDefinitionReleaseContext, candidateID string, expected int, decision, reason, sourceDigest string, now time.Time) (WorkflowDefinitionReleaseCandidate, *WorkflowDefinitionVersion, error) {
	var output WorkflowDefinitionReleaseCandidate
	var outputVersion *WorkflowDefinitionVersion
	err := repository.mutate(ctx, func(connection *sql.Conn, store *workflowDefinitionReleaseStore) error {
		candidate, version, err := store.Review(ctx, candidateID, expected, decision, reason, sourceDigest, now)
		if err != nil {
			return err
		}
		candidatePayload, err := encodeWorkflowDefinitionRecord(candidate)
		if err != nil {
			return err
		}
		updatedAt, err := workflowDefinitionUnixNano(candidate.UpdatedAt)
		if err != nil {
			return errWorkflowDefinitionStore
		}
		result, err := connection.ExecContext(workflowDefinitionRequestContext(ctx), `UPDATE workflow_definition_release_candidates SET candidate_state=?,review_version=?,updated_at_unix_nano=?,sanitized_candidate_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND candidate_id=? AND review_version=?`, candidate.State, candidate.ReviewVersion, updatedAt, string(candidatePayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, expected)
		if err != nil {
			return sqliteWorkflowDefinitionMutationError(err)
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return errWorkflowDefinitionConflict
		}
		review := candidate.Reviews[len(candidate.Reviews)-1]
		reviewPayload, err := encodeWorkflowDefinitionRecord(review)
		if err != nil {
			return err
		}
		reviewedAt, _ := workflowDefinitionUnixNano(review.ReviewedAt)
		if _, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_decisions
(tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,review_version,decision,reviewed_at_unix_nano,sanitized_decision_payload) VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, candidateID, review.ReviewVersion, review.Decision, reviewedAt, string(reviewPayload)); err != nil {
			return sqliteWorkflowDefinitionMutationError(err)
		}
		candidateAudits := store.audits[workflowDefinitionScopeKey(ctx, "candidate")]
		if err = insertSQLiteWorkflowDefinitionAudit(connection, ctx, candidateAudits[len(candidateAudits)-1]); err != nil {
			return err
		}
		if version != nil {
			versionPayload, encodeErr := encodeWorkflowDefinitionRecord(*version)
			if encodeErr != nil {
				return encodeErr
			}
			createdAt, _ := workflowDefinitionUnixNano(version.CreatedAt)
			if _, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_versions
(tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at_unix_nano,sanitized_version_payload) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, version.DefinitionID, version.Version, version.DefinitionDigest, version.CandidateID, version.CandidateReviewVersion, createdAt, string(versionPayload)); err != nil {
				return sqliteWorkflowDefinitionMutationError(err)
			}
			versionAudits := store.audits[workflowDefinitionScopeKey(ctx, "version")]
			if err = insertSQLiteWorkflowDefinitionAudit(connection, ctx, versionAudits[len(versionAudits)-1]); err != nil {
				return err
			}
			copyVersion := *version
			outputVersion = &copyVersion
		}
		output = candidate
		return nil
	})
	return output, outputVersion, err
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) DecideActivation(ctx WorkflowDefinitionReleaseContext, definitionID string, expected int, decision string, version int, reason string, now time.Time) (WorkflowDefinitionActivation, error) {
	var output WorkflowDefinitionActivation
	err := repository.mutate(ctx, func(connection *sql.Conn, store *workflowDefinitionReleaseStore) error {
		activation, err := store.DecideActivation(ctx, definitionID, expected, decision, version, reason, now)
		if err != nil {
			return err
		}
		payload, err := encodeWorkflowDefinitionRecord(activation)
		if err != nil {
			return err
		}
		updatedAt, _ := workflowDefinitionUnixNano(activation.UpdatedAt)
		var result sql.Result
		if expected == 0 {
			result, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_activations
(tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at_unix_nano,sanitized_activation_payload) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, activation.PointerVersion, activation.State, activation.ActiveVersion, activation.ActiveDefinitionDigest, updatedAt, string(payload))
		} else {
			result, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `UPDATE workflow_definition_activations SET pointer_version=?,activation_state=?,active_version=?,active_definition_digest=?,updated_at_unix_nano=?,sanitized_activation_payload=?
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? AND definition_id=? AND pointer_version=?`, activation.PointerVersion, activation.State, activation.ActiveVersion, activation.ActiveDefinitionDigest, updatedAt, string(payload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, expected)
		}
		if err != nil {
			return sqliteWorkflowDefinitionMutationError(err)
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			return errWorkflowDefinitionConflict
		}
		event := activation.Events[len(activation.Events)-1]
		eventPayload, _ := encodeWorkflowDefinitionRecord(event)
		eventAt, _ := workflowDefinitionUnixNano(event.CreatedAt)
		if _, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_activation_events
(tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,event_id,after_pointer_version,occurred_at_unix_nano,sanitized_event_payload) VALUES (?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, definitionID, event.EventID, event.AfterPointerVersion, eventAt, string(eventPayload)); err != nil {
			return sqliteWorkflowDefinitionMutationError(err)
		}
		audits := store.audits[workflowDefinitionScopeKey(ctx, "activation")]
		if err = insertSQLiteWorkflowDefinitionAudit(connection, ctx, audits[len(audits)-1]); err != nil {
			return err
		}
		output = activation
		return nil
	})
	return output, err
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ReadCandidate(ctx WorkflowDefinitionReleaseContext, candidateID string) (WorkflowDefinitionReleaseCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionReleaseCandidate{}, err
	}
	defer done()
	return store.ReadCandidate(ctx, candidateID)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ListCandidates(ctx WorkflowDefinitionReleaseContext) ([]WorkflowDefinitionReleaseCandidate, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListCandidates(ctx)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ListVersions(ctx WorkflowDefinitionReleaseContext, definitionID string) ([]WorkflowDefinitionVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return nil, err
	}
	defer done()
	return store.ListVersions(ctx, definitionID)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ReadVersion(ctx WorkflowDefinitionReleaseContext, definitionID string, version int) (WorkflowDefinitionVersion, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionVersion{}, err
	}
	defer done()
	return store.ReadVersion(ctx, definitionID, version)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ReadActivation(ctx WorkflowDefinitionReleaseContext, definitionID string) (WorkflowDefinitionActivation, error) {
	store, done, err := repository.readStore(ctx)
	if err != nil {
		return WorkflowDefinitionActivation{}, err
	}
	defer done()
	return store.ReadActivation(ctx, definitionID)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) ListSummaries(ctx ReadRepositoryContext, request ListWorkflowDefinitionSummariesRequest) ListWorkflowDefinitionSummariesResult {
	if repository == nil || repository.database == nil || strings.TrimSpace(ctx.TenantRef) == "" || strings.TrimSpace(ctx.SubjectRef) == "" {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	requestContext := ctx.RequestContext
	if requestContext == nil {
		requestContext = context.Background()
	}
	tx, err := repository.database.BeginTx(requestContext, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	defer tx.Rollback()
	versions := map[string][]WorkflowDefinitionVersion{}
	meta := map[string]workflowDefinitionSummarySource{}
	rows, err := tx.QueryContext(requestContext, `SELECT workspace_id,application_id,owner_subject_ref,definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at_unix_nano,sanitized_version_payload FROM workflow_definition_versions WHERE tenant_ref=? AND owner_subject_ref=? ORDER BY workspace_id,application_id,definition_id,definition_version`, ctx.TenantRef, ctx.SubjectRef)
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	for rows.Next() {
		var workspaceID, applicationID, owner, definitionID, digest, candidateID string
		var version, reviewVersion int
		var createdAt int64
		var payload []byte
		if rows.Scan(&workspaceID, &applicationID, &owner, &definitionID, &version, &digest, &candidateID, &reviewVersion, &createdAt, &payload) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
		}
		var value WorkflowDefinitionVersion
		if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionVersion(value) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		decodedAt, _ := workflowDefinitionUnixNano(value.CreatedAt)
		if value.DefinitionID != definitionID || value.Version != version || value.DefinitionDigest != digest || value.CandidateID != candidateID || value.CandidateReviewVersion != reviewVersion || decodedAt != createdAt {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		key := strings.Join([]string{ctx.TenantRef, workspaceID, applicationID, owner, definitionID}, "\x00")
		if value.Version != len(versions[key])+1 {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		versions[key] = append(versions[key], value)
		meta[key] = workflowDefinitionSummarySource{TenantRef: ctx.TenantRef, WorkspaceID: workspaceID, ApplicationID: applicationID, OwnerSubjectRef: owner}
	}
	if rows.Err() != nil {
		rows.Close()
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	rows.Close()
	activations := map[string]WorkflowDefinitionActivation{}
	rows, err = tx.QueryContext(requestContext, `SELECT workspace_id,application_id,owner_subject_ref,definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at_unix_nano,sanitized_activation_payload FROM workflow_definition_activations WHERE tenant_ref=? AND owner_subject_ref=? ORDER BY workspace_id,application_id,definition_id`, ctx.TenantRef, ctx.SubjectRef)
	if err != nil {
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	for rows.Next() {
		var workspaceID, applicationID, owner, definitionID, state, digest string
		var pointer, active int
		var updatedAt int64
		var payload []byte
		if rows.Scan(&workspaceID, &applicationID, &owner, &definitionID, &pointer, &state, &active, &digest, &updatedAt, &payload) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
		}
		var value WorkflowDefinitionActivation
		if decodeWorkflowDefinitionRecord(payload, &value) != nil || validateStoredWorkflowDefinitionActivation(value) != nil {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		decodedAt, _ := workflowDefinitionUnixNano(value.UpdatedAt)
		if value.DefinitionID != definitionID || value.PointerVersion != pointer || value.State != state || value.ActiveVersion != active || value.ActiveDefinitionDigest != digest || decodedAt != updatedAt {
			rows.Close()
			return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
		}
		key := strings.Join([]string{ctx.TenantRef, workspaceID, applicationID, owner, definitionID}, "\x00")
		activations[key] = value
	}
	if rows.Err() != nil {
		rows.Close()
		return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureStoreUnavailable)
	}
	rows.Close()
	sources := make([]workflowDefinitionSummarySource, 0, len(versions))
	for key, list := range versions {
		source := meta[key]
		selected := list[len(list)-1]
		if activation, ok := activations[key]; ok {
			copied := activation
			source.Activation = &copied
			if activation.State == workflowDefinitionActivationActive {
				if activation.ActiveVersion < 1 || activation.ActiveVersion > len(list) {
					return workflowDefinitionSummaryFailure(ctx, ReadRepositoryFailureContractMismatch)
				}
				selected = list[activation.ActiveVersion-1]
			}
		}
		source.Version = selected
		sources = append(sources, source)
	}
	return projectWorkflowDefinitionSummaries(ctx, request, sources)
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) mutate(ctx WorkflowDefinitionReleaseContext, operation func(*sql.Conn, *workflowDefinitionReleaseStore) error) error {
	if repository == nil || repository.database == nil || !validWorkflowDefinitionContext(ctx) {
		return errWorkflowDefinitionStore
	}
	requestContext := workflowDefinitionRequestContext(ctx)
	connection, err := repository.database.Conn(requestContext)
	if err != nil {
		return errWorkflowDefinitionStore
	}
	defer connection.Close()
	if _, err = connection.ExecContext(requestContext, "BEGIN IMMEDIATE"); err != nil {
		return errWorkflowDefinitionStore
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()
	store, err := loadSQLiteWorkflowDefinitionStore(ctx, connection)
	if err != nil {
		return err
	}
	if err = operation(connection, store); err != nil {
		return err
	}
	if _, err = connection.ExecContext(requestContext, "COMMIT"); err != nil {
		return errWorkflowDefinitionStore
	}
	committed = true
	return nil
}

func (repository *sqliteWorkflowDefinitionReleaseRepository) readStore(ctx WorkflowDefinitionReleaseContext) (*workflowDefinitionReleaseStore, func(), error) {
	if repository == nil || repository.database == nil || !validWorkflowDefinitionContext(ctx) {
		return nil, func() {}, errWorkflowDefinitionStore
	}
	tx, err := repository.database.BeginTx(workflowDefinitionRequestContext(ctx), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, func() {}, errWorkflowDefinitionStore
	}
	store, err := loadSQLiteWorkflowDefinitionStore(ctx, tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, func() {}, err
	}
	return store, func() { _ = tx.Rollback() }, nil
}

func loadSQLiteWorkflowDefinitionStore(ctx WorkflowDefinitionReleaseContext, query sqliteWorkflowDefinitionQueryer) (*workflowDefinitionReleaseStore, error) {
	store := newWorkflowDefinitionReleaseStore()
	requestContext := workflowDefinitionRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}
	rows, err := query.QueryContext(requestContext, `SELECT candidate_id,definition_id,candidate_state,review_version,source_draft_id,source_draft_version,source_draft_digest,definition_digest,created_at_unix_nano,updated_at_unix_nano,sanitized_candidate_payload FROM workflow_definition_release_candidates WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY created_at_unix_nano,candidate_id`, scope...)
	if err != nil {
		return nil, errWorkflowDefinitionStore
	}
	for rows.Next() {
		var candidateID, definitionID, state, draftID, draftDigest, definitionDigest string
		var reviewVersion, draftVersion int
		var createdAt, updatedAt int64
		var payload []byte
		if rows.Scan(&candidateID, &definitionID, &state, &reviewVersion, &draftID, &draftVersion, &draftDigest, &definitionDigest, &createdAt, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionCandidateIntoStore(store, ctx, payload)
		decodedCreated, _ := workflowDefinitionUnixNano(value.CreatedAt)
		decodedUpdated, _ := workflowDefinitionUnixNano(value.UpdatedAt)
		if loadErr != nil || value.CandidateID != candidateID || value.DefinitionID != definitionID || value.State != state || value.ReviewVersion != reviewVersion || value.SourceDraftID != draftID || value.SourceDraftVersion != draftVersion || value.SourceDraftDigest != draftDigest || value.DefinitionDigest != definitionDigest || decodedCreated != createdAt || decodedUpdated != updatedAt {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowDefinitionStore
	}
	rows, err = query.QueryContext(requestContext, `SELECT definition_id,definition_version,definition_digest,candidate_id,candidate_review_version,created_at_unix_nano,sanitized_version_payload FROM workflow_definition_versions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY definition_id,definition_version`, scope...)
	if err != nil {
		return nil, errWorkflowDefinitionStore
	}
	for rows.Next() {
		var definitionID, digest, candidateID string
		var version, reviewVersion int
		var createdAt int64
		var payload []byte
		if rows.Scan(&definitionID, &version, &digest, &candidateID, &reviewVersion, &createdAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionVersionIntoStore(store, ctx, payload)
		decodedAt, _ := workflowDefinitionUnixNano(value.CreatedAt)
		if loadErr != nil || value.DefinitionID != definitionID || value.Version != version || value.DefinitionDigest != digest || value.CandidateID != candidateID || value.CandidateReviewVersion != reviewVersion || decodedAt != createdAt {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowDefinitionStore
	}
	rows, err = query.QueryContext(requestContext, `SELECT definition_id,pointer_version,activation_state,active_version,active_definition_digest,updated_at_unix_nano,sanitized_activation_payload FROM workflow_definition_activations WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY definition_id`, scope...)
	if err != nil {
		return nil, errWorkflowDefinitionStore
	}
	for rows.Next() {
		var definitionID, state, digest string
		var pointer, active int
		var updatedAt int64
		var payload []byte
		if rows.Scan(&definitionID, &pointer, &state, &active, &digest, &updatedAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionActivationIntoStore(store, ctx, payload)
		decodedAt, _ := workflowDefinitionUnixNano(value.UpdatedAt)
		if loadErr != nil || value.DefinitionID != definitionID || value.PointerVersion != pointer || value.State != state || value.ActiveVersion != active || value.ActiveDefinitionDigest != digest || decodedAt != updatedAt {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowDefinitionStore
	}
	rows, err = query.QueryContext(requestContext, `SELECT audit_id,resource_kind,resource_id,action,occurred_at_unix_nano,sanitized_audit_payload FROM workflow_definition_release_audits WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY occurred_at_unix_nano,audit_id`, scope...)
	if err != nil {
		return nil, errWorkflowDefinitionStore
	}
	for rows.Next() {
		var auditID, kind, resourceID, action string
		var occurredAt int64
		var payload []byte
		if rows.Scan(&auditID, &kind, &resourceID, &action, &occurredAt, &payload) != nil {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
		value, loadErr := loadWorkflowDefinitionAuditIntoStore(store, ctx, payload)
		decodedAt, _ := workflowDefinitionUnixNano(value.CreatedAt)
		if loadErr != nil || value.AuditID != auditID || value.ResourceKind != kind || value.ResourceID != resourceID || value.Action != action || decodedAt != occurredAt {
			rows.Close()
			return nil, errWorkflowDefinitionStore
		}
	}
	if rows.Close() != nil || rows.Err() != nil {
		return nil, errWorkflowDefinitionStore
	}
	if err = validateSQLiteWorkflowDefinitionEvidence(ctx, query, store); err != nil {
		return nil, err
	}
	return store, nil
}

func validateSQLiteWorkflowDefinitionEvidence(ctx WorkflowDefinitionReleaseContext, query sqliteWorkflowDefinitionQueryer, store *workflowDefinitionReleaseStore) error {
	requestContext := workflowDefinitionRequestContext(ctx)
	scope := []any{ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef}
	decisionCounts := map[string]int{}
	rows, err := query.QueryContext(requestContext, `SELECT candidate_id,review_version,sanitized_decision_payload FROM workflow_definition_release_decisions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY candidate_id,review_version`, scope...)
	if err != nil {
		return errWorkflowDefinitionStore
	}
	for rows.Next() {
		var candidateID string
		var version int
		var payload []byte
		if rows.Scan(&candidateID, &version, &payload) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		var review WorkflowDefinitionReview
		if decodeWorkflowDefinitionRecord(payload, &review) != nil {
			return errWorkflowDefinitionStore
		}
		candidate, ok := store.candidates[workflowDefinitionScopeKey(ctx, candidateID)]
		if !ok || version < 1 || version > len(candidate.Reviews) || !reflect.DeepEqual(review, candidate.Reviews[version-1]) {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		decisionCounts[candidateID]++
	}
	if rows.Close() != nil || rows.Err() != nil {
		return errWorkflowDefinitionStore
	}
	for _, candidate := range store.candidates {
		if decisionCounts[candidate.CandidateID] != len(candidate.Reviews) {
			return errWorkflowDefinitionStore
		}
	}
	eventCounts := map[string]int{}
	rows, err = query.QueryContext(requestContext, `SELECT definition_id,after_pointer_version,sanitized_event_payload FROM workflow_definition_activation_events WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND owner_subject_ref=? ORDER BY definition_id,after_pointer_version`, scope...)
	if err != nil {
		return errWorkflowDefinitionStore
	}
	for rows.Next() {
		var definitionID string
		var version int
		var payload []byte
		if rows.Scan(&definitionID, &version, &payload) != nil {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		var event WorkflowDefinitionActivationEvent
		if decodeWorkflowDefinitionRecord(payload, &event) != nil {
			return errWorkflowDefinitionStore
		}
		activation, ok := store.activations[workflowDefinitionScopeKey(ctx, definitionID)]
		if !ok || version < 1 || version > len(activation.Events) || !reflect.DeepEqual(event, activation.Events[version-1]) {
			rows.Close()
			return errWorkflowDefinitionStore
		}
		eventCounts[definitionID]++
	}
	if rows.Close() != nil || rows.Err() != nil {
		return errWorkflowDefinitionStore
	}
	for _, activation := range store.activations {
		if eventCounts[activation.DefinitionID] != len(activation.Events) {
			return errWorkflowDefinitionStore
		}
	}
	return nil
}

func insertSQLiteWorkflowDefinitionAudit(connection *sql.Conn, ctx WorkflowDefinitionReleaseContext, audit WorkflowDefinitionReleaseAudit) error {
	payload, err := encodeWorkflowDefinitionRecord(audit)
	if err != nil {
		return err
	}
	occurredAt, _ := workflowDefinitionUnixNano(audit.CreatedAt)
	if _, err = connection.ExecContext(workflowDefinitionRequestContext(ctx), `INSERT INTO workflow_definition_release_audits (tenant_ref,workspace_id,application_id,owner_subject_ref,audit_id,resource_kind,resource_id,action,occurred_at_unix_nano,sanitized_audit_payload) VALUES (?,?,?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, ctx.OwnerSubjectRef, audit.AuditID, audit.ResourceKind, audit.ResourceID, audit.Action, occurredAt, string(payload)); err != nil {
		return sqliteWorkflowDefinitionMutationError(err)
	}
	return nil
}

func sqliteWorkflowDefinitionMutationError(err error) error {
	if err == nil {
		return nil
	}
	if stringsContainsAny(err.Error(), "UNIQUE constraint failed", "database is locked") {
		return errWorkflowDefinitionConflict
	}
	return errWorkflowDefinitionStore
}

func stringsContainsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

var _ workflowDefinitionReleaseRepository = (*sqliteWorkflowDefinitionReleaseRepository)(nil)
