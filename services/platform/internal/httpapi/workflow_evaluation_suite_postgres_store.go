package httpapi

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresWorkflowEvaluationSuiteStore struct{ pool *pgxpool.Pool }

func newPostgresWorkflowEvaluationSuiteStore(pool *pgxpool.Pool) *postgresWorkflowEvaluationSuiteStore {
	return &postgresWorkflowEvaluationSuiteStore{pool: pool}
}
func (s *postgresWorkflowEvaluationSuiteStore) CreateSuite(ctx WorkflowRunContext, v WorkflowEvaluationSuite) error {
	if s == nil || s.pool == nil || ctx.RequestContext == nil || validateWorkflowEvaluationSuite(ctx, v) != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	command, err := s.pool.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_suites(tenant_ref,workspace_id,application_id,suite_id,created_at,current_decision_version,sanitized_suite_record) VALUES($1,$2,$3,$4,$5,0,$6) ON CONFLICT DO NOTHING`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, v.SuiteID, v.CreatedAt, payload)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return errWorkflowEvaluationSuiteStoreContract
	}
	return nil
}
func (s *postgresWorkflowEvaluationSuiteStore) ReadSuite(ctx WorkflowRunContext, id string) (WorkflowEvaluationSuite, bool, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	var payload []byte
	err := s.pool.QueryRow(ctx.RequestContext, `SELECT sanitized_suite_record FROM workflow_evaluation_suites WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND suite_id=$4`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowEvaluationSuite{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	v, err := decodeWorkflowEvaluationSuite(ctx, payload)
	return v, true, err
}
func (s *postgresWorkflowEvaluationSuiteStore) ListSuites(ctx WorkflowRunContext, f workflowEvaluationSuiteListFilter) (workflowEvaluationSuiteListPage, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return workflowEvaluationSuiteListPage{}, errWorkflowEvaluationSuiteStoreContract
	}
	rows, err := s.pool.Query(ctx.RequestContext, `SELECT sanitized_suite_record FROM workflow_evaluation_suites WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND ($4::timestamptz IS NULL OR (created_at,suite_id)<($4,$5)) ORDER BY created_at DESC,suite_id DESC LIMIT $6`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, f.BeforeTime, f.BeforeSuiteID, f.Limit+1)
	if err != nil {
		return workflowEvaluationSuiteListPage{}, err
	}
	defer rows.Close()
	values := []WorkflowEvaluationSuite{}
	for rows.Next() {
		var payload []byte
		if rows.Scan(&payload) != nil {
			return workflowEvaluationSuiteListPage{}, errWorkflowEvaluationSuiteStoreContract
		}
		v, decodeErr := decodeWorkflowEvaluationSuite(ctx, payload)
		if decodeErr != nil {
			return workflowEvaluationSuiteListPage{}, decodeErr
		}
		values = append(values, v)
	}
	if rows.Err() != nil {
		return workflowEvaluationSuiteListPage{}, rows.Err()
	}
	more := len(values) > f.Limit
	if more {
		values = values[:f.Limit]
	}
	return workflowEvaluationSuiteListPage{Suites: values, HasMore: more}, nil
}
func (s *postgresWorkflowEvaluationSuiteStore) AppendDecision(ctx WorkflowRunContext, expected int, d WorkflowEvaluationReleaseDecision) (WorkflowEvaluationSuite, bool, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil || expected < 0 || d.Version != expected+1 || validateWorkflowEvaluationDecision(ctx, d) != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	decisionPayload, err := json.Marshal(d)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	tx, err := s.pool.Begin(ctx.RequestContext)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx.RequestContext) }()
	var current int
	var suitePayload []byte
	err = tx.QueryRow(ctx.RequestContext, `SELECT current_decision_version,sanitized_suite_record FROM workflow_evaluation_suites WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND suite_id=$4 FOR UPDATE`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, d.SuiteID).Scan(&current, &suitePayload)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkflowEvaluationSuite{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	suite, err := decodeWorkflowEvaluationSuite(ctx, suitePayload)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	if current != expected {
		return suite, false, nil
	}
	suite.CurrentDecisionVersion = d.Version
	suite.CurrentDecision = d.Decision
	updatedPayload, err := json.Marshal(suite)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	if _, err = tx.Exec(ctx.RequestContext, `INSERT INTO workflow_evaluation_suite_decisions(tenant_ref,workspace_id,application_id,suite_id,version,created_at,sanitized_decision_record) VALUES($1,$2,$3,$4,$5,$6,$7)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, d.SuiteID, d.Version, d.CreatedAt, decisionPayload); err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	command, err := tx.Exec(ctx.RequestContext, `UPDATE workflow_evaluation_suites SET current_decision_version=$5,sanitized_suite_record=$6 WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND suite_id=$4 AND current_decision_version=$7`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, d.SuiteID, d.Version, updatedPayload, expected)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	if command.RowsAffected() != 1 {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	if err = tx.Commit(ctx.RequestContext); err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	return suite, true, nil
}
func (s *postgresWorkflowEvaluationSuiteStore) ListDecisions(ctx WorkflowRunContext, id string, f workflowEvaluationDecisionListFilter) (workflowEvaluationDecisionListPage, error) {
	if s == nil || s.pool == nil || ctx.RequestContext == nil {
		return workflowEvaluationDecisionListPage{}, errWorkflowEvaluationSuiteStoreContract
	}
	rows, err := s.pool.Query(ctx.RequestContext, `SELECT version,sanitized_decision_record FROM workflow_evaluation_suite_decisions WHERE tenant_ref=$1 AND workspace_id=$2 AND application_id=$3 AND suite_id=$4 AND ($5=0 OR version<$5) ORDER BY version DESC LIMIT $6`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, f.BeforeVersion, f.Limit+1)
	if err != nil {
		return workflowEvaluationDecisionListPage{}, err
	}
	defer rows.Close()
	values := []WorkflowEvaluationReleaseDecision{}
	for rows.Next() {
		var version int
		var payload []byte
		if rows.Scan(&version, &payload) != nil {
			return workflowEvaluationDecisionListPage{}, errWorkflowEvaluationSuiteStoreContract
		}
		v, decodeErr := decodeWorkflowEvaluationDecision(ctx, payload)
		if decodeErr != nil {
			return workflowEvaluationDecisionListPage{}, decodeErr
		}
		if v.SuiteID != id || v.Version != version {
			return workflowEvaluationDecisionListPage{}, errWorkflowEvaluationSuiteStoreContract
		}
		values = append(values, v)
	}
	if rows.Err() != nil {
		return workflowEvaluationDecisionListPage{}, rows.Err()
	}
	more := len(values) > f.Limit
	if more {
		values = values[:f.Limit]
	}
	return workflowEvaluationDecisionListPage{Decisions: values, HasMore: more}, nil
}
func decodeWorkflowEvaluationSuite(ctx WorkflowRunContext, payload []byte) (WorkflowEvaluationSuite, error) {
	var v WorkflowEvaluationSuite
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&v) != nil || validateWorkflowEvaluationSuite(ctx, v) != nil {
		return WorkflowEvaluationSuite{}, errWorkflowEvaluationSuiteStoreContract
	}
	return v, nil
}
func decodeWorkflowEvaluationDecision(ctx WorkflowRunContext, payload []byte) (WorkflowEvaluationReleaseDecision, error) {
	var v WorkflowEvaluationReleaseDecision
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&v) != nil || validateWorkflowEvaluationDecision(ctx, v) != nil {
		return WorkflowEvaluationReleaseDecision{}, errWorkflowEvaluationSuiteStoreContract
	}
	return v, nil
}

var _ workflowEvaluationSuiteStore = (*postgresWorkflowEvaluationSuiteStore)(nil)
