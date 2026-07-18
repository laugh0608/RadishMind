package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type sqliteWorkflowEvaluationStore struct{ database *sql.DB }

func newSQLiteWorkflowEvaluationStore(database *sql.DB) *sqliteWorkflowEvaluationStore {
	return &sqliteWorkflowEvaluationStore{database: database}
}

func (s *sqliteWorkflowEvaluationStore) CreateCase(ctx WorkflowRunContext, value WorkflowEvaluationCase) error {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil || validateWorkflowEvaluationCase(ctx, value) != nil {
		return errWorkflowEvaluationStoreContract
	}
	payload, createdAt, revisedAt, err := encodeSQLiteWorkflowEvaluationCase(value)
	if err != nil {
		return errWorkflowEvaluationStoreContract
	}
	tx, err := s.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	result, err := tx.ExecContext(ctx.RequestContext, `INSERT OR IGNORE INTO workflow_evaluation_cases
		(tenant_ref,workspace_id,application_id,case_id,baseline_run_id,created_at_unix_nano,current_version,sanitized_case_record)
		VALUES(?,?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.BaselineRunID, createdAt, value.Version, payload)
	if err != nil {
		return err
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return errWorkflowEvaluationStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_evaluation_case_revisions
		(tenant_ref,workspace_id,application_id,case_id,version,revised_at_unix_nano,sanitized_revision_record)
		VALUES(?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.Version, revisedAt, payload); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *sqliteWorkflowEvaluationStore) ReviseCase(ctx WorkflowRunContext, expected int, value WorkflowEvaluationCase) (WorkflowEvaluationCase, bool, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil || expected < 1 || validateWorkflowEvaluationCase(ctx, value) != nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	payload, _, revisedAt, err := encodeSQLiteWorkflowEvaluationCase(value)
	if err != nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	tx, err := s.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	defer func() { _ = tx.Rollback() }()
	var currentVersion int
	var currentPayload string
	err = tx.QueryRowContext(ctx.RequestContext, `SELECT current_version,sanitized_case_record FROM workflow_evaluation_cases
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND case_id=?`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID).Scan(&currentVersion, &currentPayload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	if currentVersion != expected {
		current, decodeErr := decodeWorkflowEvaluationCase(ctx, []byte(currentPayload))
		return current, false, decodeErr
	}
	if value.Version != expected+1 || value.PreviousVersion != expected {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_evaluation_case_revisions
		(tenant_ref,workspace_id,application_id,case_id,version,revised_at_unix_nano,sanitized_revision_record)
		VALUES(?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, value.Version, revisedAt, payload); err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE workflow_evaluation_cases
		SET baseline_run_id=?,current_version=?,sanitized_case_record=?
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND case_id=? AND current_version=?`,
		value.BaselineRunID, value.Version, payload, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.CaseID, expected)
	if err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	if err = tx.Commit(); err != nil {
		return WorkflowEvaluationCase{}, false, err
	}
	return cloneWorkflowEvaluationCase(value), true, nil
}

func (s *sqliteWorkflowEvaluationStore) ReadCase(ctx WorkflowRunContext, id string) (WorkflowEvaluationCase, bool, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	value, err := scanSQLiteWorkflowEvaluationCase(ctx, s.database.QueryRowContext(ctx.RequestContext, sqliteWorkflowEvaluationCaseReadSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, normalizeSQLiteEvaluationStoreError(err)
	}
	return value, true, nil
}

func (s *sqliteWorkflowEvaluationStore) ReadRevision(ctx WorkflowRunContext, id string, version int) (WorkflowEvaluationCase, bool, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil || version < 1 {
		return WorkflowEvaluationCase{}, false, errWorkflowEvaluationStoreContract
	}
	value, err := scanSQLiteWorkflowEvaluationRevision(ctx, s.database.QueryRowContext(ctx.RequestContext, sqliteWorkflowEvaluationRevisionReadSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, version))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowEvaluationCase{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationCase{}, false, normalizeSQLiteEvaluationStoreError(err)
	}
	return value, true, nil
}

func (s *sqliteWorkflowEvaluationStore) ListCases(ctx WorkflowRunContext, filter WorkflowEvaluationListFilter) (WorkflowEvaluationListPage, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return WorkflowEvaluationListPage{}, errWorkflowEvaluationStoreContract
	}
	before, err := optionalWorkflowRunUnixNano(filter.BeforeTime)
	if err != nil {
		return WorkflowEvaluationListPage{}, errWorkflowEvaluationStoreContract
	}
	rows, err := s.database.QueryContext(ctx.RequestContext, sqliteWorkflowEvaluationCaseListSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, filter.BaselineRunID, filter.BaselineRunID,
		before, before, before, filter.BeforeCaseID, filter.Limit+1)
	if err != nil {
		return WorkflowEvaluationListPage{}, normalizeSQLiteEvaluationStoreError(err)
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationCase, 0, filter.Limit+1)
	for rows.Next() {
		value, scanErr := scanSQLiteWorkflowEvaluationCase(ctx, rows)
		if scanErr != nil {
			return WorkflowEvaluationListPage{}, normalizeSQLiteEvaluationStoreError(scanErr)
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return WorkflowEvaluationListPage{}, normalizeSQLiteEvaluationStoreError(rows.Err())
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return WorkflowEvaluationListPage{Cases: values, HasMore: hasMore}, nil
}

func (s *sqliteWorkflowEvaluationStore) ListRevisions(ctx WorkflowRunContext, id string, filter WorkflowEvaluationRevisionListFilter) (WorkflowEvaluationRevisionListPage, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return WorkflowEvaluationRevisionListPage{}, errWorkflowEvaluationStoreContract
	}
	rows, err := s.database.QueryContext(ctx.RequestContext, sqliteWorkflowEvaluationRevisionListSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, filter.BeforeVersion, filter.BeforeVersion, filter.Limit+1)
	if err != nil {
		return WorkflowEvaluationRevisionListPage{}, normalizeSQLiteEvaluationStoreError(err)
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationCase, 0, filter.Limit+1)
	for rows.Next() {
		value, scanErr := scanSQLiteWorkflowEvaluationRevision(ctx, rows)
		if scanErr != nil {
			return WorkflowEvaluationRevisionListPage{}, normalizeSQLiteEvaluationStoreError(scanErr)
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return WorkflowEvaluationRevisionListPage{}, normalizeSQLiteEvaluationStoreError(rows.Err())
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return WorkflowEvaluationRevisionListPage{Revisions: values, HasMore: hasMore}, nil
}

type sqliteEvaluationScanner interface{ Scan(...any) error }

func scanSQLiteWorkflowEvaluationCase(ctx WorkflowRunContext, row sqliteEvaluationScanner) (WorkflowEvaluationCase, error) {
	var tenant, workspace, application, caseID, baseline, payload string
	var createdAt int64
	var currentVersion int
	if err := row.Scan(&tenant, &workspace, &application, &caseID, &baseline, &createdAt, &currentVersion, &payload); err != nil {
		return WorkflowEvaluationCase{}, err
	}
	value, err := decodeWorkflowEvaluationCase(ctx, []byte(payload))
	if err != nil {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	decodedCreatedAt, err := workflowEvaluationUnixNano(value.CreatedAt)
	if err != nil || tenant != ctx.TenantRef || workspace != ctx.WorkspaceID || application != ctx.ApplicationID ||
		caseID != value.CaseID || baseline != value.BaselineRunID || createdAt != decodedCreatedAt || currentVersion != value.Version {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	return value, nil
}

func scanSQLiteWorkflowEvaluationRevision(ctx WorkflowRunContext, row sqliteEvaluationScanner) (WorkflowEvaluationCase, error) {
	var tenant, workspace, application, caseID, payload string
	var version int
	var revisedAt int64
	if err := row.Scan(&tenant, &workspace, &application, &caseID, &version, &revisedAt, &payload); err != nil {
		return WorkflowEvaluationCase{}, err
	}
	value, err := decodeWorkflowEvaluationCase(ctx, []byte(payload))
	if err != nil {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	decodedRevisedAt, err := workflowEvaluationUnixNano(value.RevisedAt)
	if err != nil || tenant != ctx.TenantRef || workspace != ctx.WorkspaceID || application != ctx.ApplicationID ||
		caseID != value.CaseID || version != value.Version || revisedAt != decodedRevisedAt {
		return WorkflowEvaluationCase{}, errWorkflowEvaluationStoreContract
	}
	return value, nil
}

func encodeSQLiteWorkflowEvaluationCase(value WorkflowEvaluationCase) (string, int64, int64, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", 0, 0, err
	}
	createdAt, err := workflowEvaluationUnixNano(value.CreatedAt)
	if err != nil {
		return "", 0, 0, err
	}
	revisedAt, err := workflowEvaluationUnixNano(value.RevisedAt)
	return string(payload), createdAt, revisedAt, err
}

func workflowEvaluationUnixNano(value string) (int64, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0, err
	}
	return workflowRunUnixNano(parsed.UTC())
}

func validSQLiteEvaluationContext(ctx WorkflowRunContext) bool {
	return ctx.RequestContext != nil && ctx.TenantRef != "" && ctx.WorkspaceID != "" && ctx.ApplicationID != ""
}

func normalizeSQLiteEvaluationStoreError(err error) error {
	if errors.Is(err, errWorkflowEvaluationStoreContract) {
		return errWorkflowEvaluationStoreContract
	}
	return err
}

const sqliteWorkflowEvaluationCaseColumns = `tenant_ref,workspace_id,application_id,case_id,baseline_run_id,created_at_unix_nano,current_version,sanitized_case_record`
const sqliteWorkflowEvaluationRevisionColumns = `tenant_ref,workspace_id,application_id,case_id,version,revised_at_unix_nano,sanitized_revision_record`
const sqliteWorkflowEvaluationCaseReadSQL = `SELECT ` + sqliteWorkflowEvaluationCaseColumns + ` FROM workflow_evaluation_cases WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND case_id=?`
const sqliteWorkflowEvaluationRevisionReadSQL = `SELECT ` + sqliteWorkflowEvaluationRevisionColumns + ` FROM workflow_evaluation_case_revisions WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND case_id=? AND version=?`
const sqliteWorkflowEvaluationCaseListSQL = `SELECT ` + sqliteWorkflowEvaluationCaseColumns + ` FROM workflow_evaluation_cases
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND (?='' OR baseline_run_id=?)
AND (? IS NULL OR created_at_unix_nano<? OR (created_at_unix_nano=? AND case_id<?))
ORDER BY created_at_unix_nano DESC,case_id DESC LIMIT ?`
const sqliteWorkflowEvaluationRevisionListSQL = `SELECT ` + sqliteWorkflowEvaluationRevisionColumns + ` FROM workflow_evaluation_case_revisions
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND case_id=? AND (?=0 OR version<?)
ORDER BY version DESC LIMIT ?`

var _ workflowEvaluationStore = (*sqliteWorkflowEvaluationStore)(nil)
