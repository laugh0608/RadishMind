CREATE TABLE workflow_evaluation_cases (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    case_id text NOT NULL,
    baseline_run_id text NOT NULL,
    created_at timestamptz NOT NULL,
    sanitized_case_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, case_id)
);

CREATE INDEX workflow_evaluation_cases_history_idx ON workflow_evaluation_cases
    (tenant_ref, workspace_id, application_id, created_at DESC, case_id DESC);
CREATE INDEX workflow_evaluation_cases_baseline_idx ON workflow_evaluation_cases
    (tenant_ref, workspace_id, application_id, baseline_run_id, created_at DESC, case_id DESC);
