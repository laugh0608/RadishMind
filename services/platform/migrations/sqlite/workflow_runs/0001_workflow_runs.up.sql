CREATE TABLE workflow_run_records (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    store_schema_version TEXT NOT NULL CHECK (store_schema_version = 'workflow_runs_store_v1'),
    schema_version TEXT NOT NULL CHECK (schema_version IN ('workflow_run_record.v0', 'workflow_run_record.v1')),
    run_status TEXT NOT NULL CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled')),
    started_at_unix_nano INTEGER NOT NULL,
    completed_at_unix_nano INTEGER,
    actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    failure_code TEXT NOT NULL DEFAULT '',
    failure_boundary TEXT NOT NULL DEFAULT '',
    selected_provider TEXT NOT NULL DEFAULT '',
    selected_model TEXT NOT NULL DEFAULT '',
    sanitized_run_record TEXT NOT NULL CHECK (
        json_valid(sanitized_run_record)
        AND json_type(sanitized_run_record) = 'object'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, run_id),
    CHECK (
        (run_status = 'running' AND completed_at_unix_nano IS NULL)
        OR
        (run_status IN ('succeeded', 'failed', 'canceled')
            AND completed_at_unix_nano IS NOT NULL
            AND completed_at_unix_nano >= started_at_unix_nano)
    )
) STRICT;

CREATE INDEX workflow_run_records_history_idx ON workflow_run_records (
    tenant_ref, workspace_id, application_id, started_at_unix_nano DESC, run_id DESC
);

CREATE INDEX workflow_run_records_status_idx ON workflow_run_records (
    tenant_ref, workspace_id, application_id, run_status, started_at_unix_nano DESC, run_id DESC
);

CREATE INDEX workflow_run_records_draft_idx ON workflow_run_records (
    tenant_ref, workspace_id, application_id, draft_id, started_at_unix_nano DESC, run_id DESC
);

CREATE INDEX workflow_run_records_failure_idx ON workflow_run_records (
    tenant_ref, workspace_id, application_id,
    failure_code, failure_boundary, started_at_unix_nano DESC, run_id DESC
);

CREATE INDEX workflow_run_records_provider_model_idx ON workflow_run_records (
    tenant_ref, workspace_id, application_id,
    selected_provider, selected_model, started_at_unix_nano DESC, run_id DESC
);

CREATE INDEX workflow_run_records_retention_idx ON workflow_run_records (
    completed_at_unix_nano, started_at_unix_nano
);
