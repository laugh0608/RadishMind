package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
)

type sqliteWorkflowEvaluationSuiteStore struct{ database *sql.DB }

func newSQLiteWorkflowEvaluationSuiteStore(database *sql.DB) *sqliteWorkflowEvaluationSuiteStore {
	return &sqliteWorkflowEvaluationSuiteStore{database: database}
}

func (s *sqliteWorkflowEvaluationSuiteStore) CreateSuite(ctx WorkflowRunContext, value WorkflowEvaluationSuite) error {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil || validateWorkflowEvaluationSuite(ctx, value) != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	createdAt, err := workflowEvaluationUnixNano(value.CreatedAt)
	if err != nil {
		return errWorkflowEvaluationSuiteStoreContract
	}
	result, err := s.database.ExecContext(ctx.RequestContext, `INSERT OR IGNORE INTO workflow_evaluation_suites
		(tenant_ref,workspace_id,application_id,suite_id,created_at_unix_nano,current_decision_version,sanitized_suite_record)
		VALUES(?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, value.SuiteID, createdAt, 0, string(payload))
	if err != nil {
		return err
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return errWorkflowEvaluationSuiteStoreContract
	}
	return nil
}

func (s *sqliteWorkflowEvaluationSuiteStore) ReadSuite(ctx WorkflowRunContext, id string) (WorkflowEvaluationSuite, bool, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	value, err := scanSQLiteWorkflowEvaluationSuite(ctx, s.database.QueryRowContext(ctx.RequestContext, sqliteWorkflowEvaluationSuiteReadSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id))
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowEvaluationSuite{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationSuite{}, false, normalizeSQLiteEvaluationSuiteStoreError(err)
	}
	return value, true, nil
}

func (s *sqliteWorkflowEvaluationSuiteStore) ListSuites(ctx WorkflowRunContext, filter workflowEvaluationSuiteListFilter) (workflowEvaluationSuiteListPage, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return workflowEvaluationSuiteListPage{}, errWorkflowEvaluationSuiteStoreContract
	}
	before, err := optionalWorkflowRunUnixNano(filter.BeforeTime)
	if err != nil {
		return workflowEvaluationSuiteListPage{}, errWorkflowEvaluationSuiteStoreContract
	}
	rows, err := s.database.QueryContext(ctx.RequestContext, sqliteWorkflowEvaluationSuiteListSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, before, before, before, filter.BeforeSuiteID, filter.Limit+1)
	if err != nil {
		return workflowEvaluationSuiteListPage{}, normalizeSQLiteEvaluationSuiteStoreError(err)
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationSuite, 0, filter.Limit+1)
	for rows.Next() {
		value, scanErr := scanSQLiteWorkflowEvaluationSuite(ctx, rows)
		if scanErr != nil {
			return workflowEvaluationSuiteListPage{}, normalizeSQLiteEvaluationSuiteStoreError(scanErr)
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return workflowEvaluationSuiteListPage{}, normalizeSQLiteEvaluationSuiteStoreError(rows.Err())
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return workflowEvaluationSuiteListPage{Suites: values, HasMore: hasMore}, nil
}

func (s *sqliteWorkflowEvaluationSuiteStore) AppendDecision(ctx WorkflowRunContext, expected int, decision WorkflowEvaluationReleaseDecision) (WorkflowEvaluationSuite, bool, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil || expected < 0 || decision.Version != expected+1 || validateWorkflowEvaluationDecision(ctx, decision) != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	decisionPayload, err := json.Marshal(decision)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	createdAt, err := workflowEvaluationUnixNano(decision.CreatedAt)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	tx, err := s.database.BeginTx(ctx.RequestContext, nil)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	defer func() { _ = tx.Rollback() }()
	var current int
	var suitePayload string
	err = tx.QueryRowContext(ctx.RequestContext, `SELECT current_decision_version,sanitized_suite_record FROM workflow_evaluation_suites
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND suite_id=?`,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, decision.SuiteID).Scan(&current, &suitePayload)
	if errors.Is(err, sql.ErrNoRows) {
		return WorkflowEvaluationSuite{}, false, nil
	}
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	suite, err := decodeWorkflowEvaluationSuite(ctx, []byte(suitePayload))
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	if current != expected || suite.CurrentDecisionVersion != current {
		return suite, false, nil
	}
	suite.CurrentDecisionVersion = decision.Version
	suite.CurrentDecision = decision.Decision
	updatedPayload, err := json.Marshal(suite)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	if _, err = tx.ExecContext(ctx.RequestContext, `INSERT INTO workflow_evaluation_suite_decisions
		(tenant_ref,workspace_id,application_id,suite_id,version,created_at_unix_nano,sanitized_decision_record)
		VALUES(?,?,?,?,?,?,?)`, ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, decision.SuiteID, decision.Version, createdAt, string(decisionPayload)); err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	result, err := tx.ExecContext(ctx.RequestContext, `UPDATE workflow_evaluation_suites
		SET current_decision_version=?,sanitized_suite_record=?
		WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND suite_id=? AND current_decision_version=?`,
		decision.Version, string(updatedPayload), ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, decision.SuiteID, expected)
	if err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	if affected, affectedErr := result.RowsAffected(); affectedErr != nil || affected != 1 {
		return WorkflowEvaluationSuite{}, false, errWorkflowEvaluationSuiteStoreContract
	}
	if err = tx.Commit(); err != nil {
		return WorkflowEvaluationSuite{}, false, err
	}
	return suite, true, nil
}

func (s *sqliteWorkflowEvaluationSuiteStore) ListDecisions(ctx WorkflowRunContext, id string, filter workflowEvaluationDecisionListFilter) (workflowEvaluationDecisionListPage, error) {
	if !validSQLiteEvaluationContext(ctx) || s == nil || s.database == nil {
		return workflowEvaluationDecisionListPage{}, errWorkflowEvaluationSuiteStoreContract
	}
	rows, err := s.database.QueryContext(ctx.RequestContext, sqliteWorkflowEvaluationDecisionListSQL,
		ctx.TenantRef, ctx.WorkspaceID, ctx.ApplicationID, id, filter.BeforeVersion, filter.BeforeVersion, filter.Limit+1)
	if err != nil {
		return workflowEvaluationDecisionListPage{}, normalizeSQLiteEvaluationSuiteStoreError(err)
	}
	defer rows.Close()
	values := make([]WorkflowEvaluationReleaseDecision, 0, filter.Limit+1)
	for rows.Next() {
		value, scanErr := scanSQLiteWorkflowEvaluationDecision(ctx, rows)
		if scanErr != nil {
			return workflowEvaluationDecisionListPage{}, normalizeSQLiteEvaluationSuiteStoreError(scanErr)
		}
		values = append(values, value)
	}
	if rows.Err() != nil {
		return workflowEvaluationDecisionListPage{}, normalizeSQLiteEvaluationSuiteStoreError(rows.Err())
	}
	hasMore := len(values) > filter.Limit
	if hasMore {
		values = values[:filter.Limit]
	}
	return workflowEvaluationDecisionListPage{Decisions: values, HasMore: hasMore}, nil
}

func scanSQLiteWorkflowEvaluationSuite(ctx WorkflowRunContext, row sqliteEvaluationScanner) (WorkflowEvaluationSuite, error) {
	var tenant, workspace, application, suiteID, payload string
	var createdAt int64
	var currentVersion int
	if err := row.Scan(&tenant, &workspace, &application, &suiteID, &createdAt, &currentVersion, &payload); err != nil {
		return WorkflowEvaluationSuite{}, err
	}
	value, err := decodeWorkflowEvaluationSuite(ctx, []byte(payload))
	if err != nil {
		return WorkflowEvaluationSuite{}, errWorkflowEvaluationSuiteStoreContract
	}
	decodedCreatedAt, err := workflowEvaluationUnixNano(value.CreatedAt)
	if err != nil || tenant != ctx.TenantRef || workspace != ctx.WorkspaceID || application != ctx.ApplicationID ||
		suiteID != value.SuiteID || createdAt != decodedCreatedAt || currentVersion != value.CurrentDecisionVersion {
		return WorkflowEvaluationSuite{}, errWorkflowEvaluationSuiteStoreContract
	}
	return value, nil
}

func scanSQLiteWorkflowEvaluationDecision(ctx WorkflowRunContext, row sqliteEvaluationScanner) (WorkflowEvaluationReleaseDecision, error) {
	var tenant, workspace, application, suiteID, payload string
	var version int
	var createdAt int64
	if err := row.Scan(&tenant, &workspace, &application, &suiteID, &version, &createdAt, &payload); err != nil {
		return WorkflowEvaluationReleaseDecision{}, err
	}
	value, err := decodeWorkflowEvaluationDecision(ctx, []byte(payload))
	if err != nil {
		return WorkflowEvaluationReleaseDecision{}, errWorkflowEvaluationSuiteStoreContract
	}
	decodedCreatedAt, err := workflowEvaluationUnixNano(value.CreatedAt)
	if err != nil || tenant != ctx.TenantRef || workspace != ctx.WorkspaceID || application != ctx.ApplicationID ||
		suiteID != value.SuiteID || version != value.Version || createdAt != decodedCreatedAt {
		return WorkflowEvaluationReleaseDecision{}, errWorkflowEvaluationSuiteStoreContract
	}
	return value, nil
}

func normalizeSQLiteEvaluationSuiteStoreError(err error) error {
	if errors.Is(err, errWorkflowEvaluationSuiteStoreContract) {
		return errWorkflowEvaluationSuiteStoreContract
	}
	return err
}

const sqliteWorkflowEvaluationSuiteColumns = `tenant_ref,workspace_id,application_id,suite_id,created_at_unix_nano,current_decision_version,sanitized_suite_record`
const sqliteWorkflowEvaluationDecisionColumns = `tenant_ref,workspace_id,application_id,suite_id,version,created_at_unix_nano,sanitized_decision_record`
const sqliteWorkflowEvaluationSuiteReadSQL = `SELECT ` + sqliteWorkflowEvaluationSuiteColumns + ` FROM workflow_evaluation_suites WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND suite_id=?`
const sqliteWorkflowEvaluationSuiteListSQL = `SELECT ` + sqliteWorkflowEvaluationSuiteColumns + ` FROM workflow_evaluation_suites
WHERE tenant_ref=? AND workspace_id=? AND application_id=?
AND (? IS NULL OR created_at_unix_nano<? OR (created_at_unix_nano=? AND suite_id<?))
ORDER BY created_at_unix_nano DESC,suite_id DESC LIMIT ?`
const sqliteWorkflowEvaluationDecisionListSQL = `SELECT ` + sqliteWorkflowEvaluationDecisionColumns + ` FROM workflow_evaluation_suite_decisions
WHERE tenant_ref=? AND workspace_id=? AND application_id=? AND suite_id=? AND (?=0 OR version<?)
ORDER BY version DESC LIMIT ?`

var _ workflowEvaluationSuiteStore = (*sqliteWorkflowEvaluationSuiteStore)(nil)
