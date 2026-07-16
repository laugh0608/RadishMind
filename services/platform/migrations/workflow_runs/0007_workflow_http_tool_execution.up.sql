ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_run_records_run_status_check;

ALTER TABLE workflow_run_records
    ADD CONSTRAINT workflow_run_records_run_status_check
    CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown'));

CREATE TABLE workflow_http_tool_execution_attempts (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    plan_id text NOT NULL CHECK (btrim(plan_id) <> ''),
    confirmation_id text NOT NULL CHECK (btrim(confirmation_id) <> ''),
    attempt_id text NOT NULL CHECK (attempt_id ~ '^wtea_[a-z0-9]{16,64}$'),
    run_id text NOT NULL CHECK (run_id ~ '^run_[a-z0-9]{16,64}$'),
    status text NOT NULL CHECK (status IN ('claimed', 'succeeded', 'failed', 'outcome_unknown')),
    tool_plan_digest text NOT NULL CHECK (tool_plan_digest ~ '^sha256:[0-9a-f]{64}$'),
    claimed_at timestamptz NOT NULL,
    completed_at timestamptz,
    failure_code text NOT NULL DEFAULT '',
    sanitized_execution_attempt jsonb NOT NULL CHECK (jsonb_typeof(sanitized_execution_attempt) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id, attempt_id),
    UNIQUE (tenant_ref, workspace_id, application_id, plan_id),
    UNIQUE (tenant_ref, workspace_id, application_id, run_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, plan_id, confirmation_id)
        REFERENCES workflow_http_tool_confirmation_decisions (
            tenant_ref, workspace_id, application_id, plan_id, confirmation_id
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest)
        REFERENCES workflow_http_tool_action_plans (
            tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_ref, workspace_id, application_id, run_id)
        REFERENCES workflow_run_records (
            tenant_ref, workspace_id, application_id, run_id
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK (
        (status = 'claimed' AND completed_at IS NULL AND failure_code = '')
        OR
        (status IN ('succeeded', 'failed', 'outcome_unknown')
            AND completed_at IS NOT NULL AND completed_at >= claimed_at)
    )
);

CREATE INDEX workflow_http_tool_execution_attempts_status_idx
    ON workflow_http_tool_execution_attempts (
        tenant_ref, workspace_id, application_id,
        status, claimed_at, attempt_id
    );
