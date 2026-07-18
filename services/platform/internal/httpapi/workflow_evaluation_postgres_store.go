package httpapi

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowEvaluationStore struct{ pool *pgxpool.Pool }

func newPostgresWorkflowEvaluationStore(pool *pgxpool.Pool) *postgresWorkflowEvaluationStore {
	return &postgresWorkflowEvaluationStore{pool: pool}
}

func (s *postgresWorkflowEvaluationStore) CreateCase(ctx WorkflowRunContext, value WorkflowEvaluationCase) error {
	if s == nil || s.pool == nil || ctx.RequestContext == nil || validateWorkflowEvaluationCase(ctx, value) != nil {
		return errWorkflowEvaluationStoreContract
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return errWorkflowEvaluationStoreContract
	}
	tx, err := s.pool.Begin(ctx.RequestContext)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	command, err := tx.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_cases (tenant_ref,workspace_id,application_id,case_id,baseline_run_id,created_at,current_version,sanitized_case_record) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.BaselineRunID, value.CreatedAt, value.Version, payload)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return errWorkflowEvaluationStoreContract
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_case_revisions (tenant_ref,workspace_id,application_id,case_id,version,created_at,sanitized_revision_record) VALUES ($1,$2,$3,$4,$5,$6,$7)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.Version, value.RevisedAt, payload); err != nil {
		return err
	}
	return tx.Commit(ctx.RequestContext)
}
func (s *postgresWorkflowEvaluationStore) ReviseCase(ctx WorkflowRunContext, expected int, value WorkflowEvaluationCase) (WorkflowEvaluationCase, bool, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil || expected < 1 || validateWorkflowEvaluationCase(ctx, value) != nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	tx, err := s.pool.Begin(ctx.RequestContext)
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	var currentVersion int
	var currentPayload []byte
	err = tx.QueryRow(ctx.RequestContext, `SELECT current_version,sanitized_case_record FROM workflow_evaluation_cases WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4 FOR UPDATE`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID).Scan(&currentVersion, &currentPayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	if currentVersion != expected {
		current, decodeErr := decodeWorkflowEvaluationCase(ctx, currentPayload)
		return current, false, decodeErr
	}
	if value.Version != expected+1 || value.PreviousVersion != expected {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_case_revisions (tenant_ref,workspace_id,application_id,case_id,version,created_at,sanitized_revision_record) VALUES ($1,$2,$3,$4,$5,$6,$7)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.Version, value.RevisedAt, payload); err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE workflow_evaluation_cases SET baseline_run_id=$5,current_version=$6,sanitized_case_record=$7 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4 AND current_version=$8`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.BaselineRunID, value.Version, payload, expected)
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	if command.RowsAffected() != 1 {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	return cloneWorkflowEvaluationCase(value), true, nil
}
func (s *postgresWorkflowEvaluationStore) ReadCase(ctx WorkflowRunContext, id string) (WorkflowEvaluationCase, bool, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	var payload []byte
	err := s.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_case_record FROM workflow_evaluation_cases WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	value, err := decodeWorkflowEvaluationCase(ctx, payload)
	return value, true, err
}
func (s *postgresWorkflowEvaluationStore) ReadRevision(ctx WorkflowRunContext, id string, version int) (WorkflowEvaluationCase, bool, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil || version < 1 {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	var payload []byte
	err := s.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_revision_record FROM workflow_evaluation_case_revisions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4 AND version=$5`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, version).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	value, err := decodeWorkflowEvaluationCase(ctx, payload)
	return value, true, err
}
func (s *postgresWorkflowEvaluationStore) ListCases(ctx WorkflowRunContext, filter WorkflowEvaluationListFilter) (WorkflowEvaluationListPage, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return WorkflowEvaluationListPage{}, errWorkflowEvaluationStoreContract
	}
	rows, err := s.pool.Query(ctx.RequestContext, `SELECT sanitized_case_record FROM workflow_evaluation_cases WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND ($4='' OR baseline_run_id=$4) AND ($5::timestamptz IS NULL OR (created_at,case_id)<($5,$6)) ORDER BY created_at DESC,case_id DESC LIMIT $7`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, filter.BaselineRunID, filter.BeforeTime, filter.BeforeCaseID, filter.Limit+1)
	if err != nil {
		return WorkflowEvaluationListPage{}, err
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationCase, 0, filter.Limit+1)
	for rows.Next() {
		var payload []byte
		if rows.Scan(&payload) != nil {
			return WorkflowEvaluationListPage{}, errWorkflowEvaluationStoreContract
		}
		value, decodeErr := decodeWorkflowEvaluationCase(ctx, payload)
		if decodeErr != nil {
			return WorkflowEvaluationListPage{}, decodeErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return WorkflowEvaluationListPage{}, rows.Err()
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return WorkflowEvaluationListPage{Cases: values, HasMore: hasMore}, nil
}
func (s *postgresWorkflowEvaluationStore) ListRevisions(ctx WorkflowRunContext, id string, filter WorkflowEvaluationRevisionListFilter) (WorkflowEvaluationRevisionListPage, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return WorkflowEvaluationRevisionListPage{}, errWorkflowEvaluationStoreContract
	}
	rows, err := s.pool.Query(ctx.RequestContext, `SELECT sanitized_revision_record FROM workflow_evaluation_case_revisions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND case_id=$4 AND ($5=0 OR version<$5) ORDER BY version DESC LIMIT $6`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, filter.BeforeVersion, filter.Limit+1)
	if err != nil {
		return WorkflowEvaluationRevisionListPage{}, err
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationCase, 0, filter.Limit+1)
	for rows.Next() {
		var payload []byte
		if rows.Scan(&payload) != nil {
			return WorkflowEvaluationRevisionListPage{}, errWorkflowEvaluationStoreContract
		}
		value, decodeErr := decodeWorkflowEvaluationCase(ctx, payload)
		if decodeErr != nil {
			return WorkflowEvaluationRevisionListPage{}, decodeErr
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return WorkflowEvaluationRevisionListPage{}, rows.Err()
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return WorkflowEvaluationRevisionListPage{Revisions: values, HasMore: hasMore}, nil
}
func decodeWorkflowEvaluationCase(ctx WorkflowRunContext, payload []byte) (WorkflowEvaluationCase, error) {
	var value WorkflowEvaluationCase
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	value = upgradeWorkflowEvaluationCase(value)
	if validateWorkflowEvaluationCase(ctx, value) != nil {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	return value, nil
}

var _ workflowEvaluationStore = (*postgresWorkflowEvaluationStore)(nil)
