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
	command, err := s.pool.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_cases (tenant_ref,workspace_id,application_id,case_id,baseline_run_id,created_at,sanitized_case_record) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.BaselineRunID, value.CreatedAt, payload)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return errWorkflowEvaluationStoreContract
	}
	return nil
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
	value, err := decodePostgresWorkflowEvaluationCase(ctx, payload)
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
		value, decodeErr := decodePostgresWorkflowEvaluationCase(ctx, payload)
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
func decodePostgresWorkflowEvaluationCase(ctx WorkflowRunContext, payload []byte) (WorkflowEvaluationCase, error) {
	var value WorkflowEvaluationCase
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || validateWorkflowEvaluationCase(ctx, value) != nil {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	return value, nil
}

var _ workflowEvaluationStore = (*postgresWorkflowEvaluationStore)(nil)
