CREATE TABLE workflow_run_records (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    run_id text NOT NULL,
    draft_id text NOT NULL,
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    record_version bigint NOT NULL CHECK (record_version > 0),
    schema_version text NOT NULL,
    run_status text NOT NULL CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled')),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    sanitized_run_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, run_id)
);

CREATE INDEX workflow_run_records_history_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, started_at DESC, run_id DESC);
CREATE INDEX workflow_run_records_status_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, run_status, started_at DESC, run_id DESC);
CREATE INDEX workflow_run_records_draft_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, draft_id, started_at DESC, run_id DESC);
CREATE INDEX workflow_run_records_retention_idx ON workflow_run_records
    (completed_at, started_at);
