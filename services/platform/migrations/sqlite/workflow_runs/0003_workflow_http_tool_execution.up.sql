CREATE TABLE workflow_run_records_v3 (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    store_schema_version TEXT NOT NULL CHECK (store_schema_version = 'workflow_runs_store_v3'),
    schema_version TEXT NOT NULL CHECK (schema_version IN (
        'workflow_run_record.v0', 'workflow_run_record.v1', 'workflow_run_record.v2'
    )),
    run_status TEXT NOT NULL CHECK (run_status IN (
        'running', 'succeeded', 'failed', 'canceled', 'outcome_unknown'
    )),
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
        (run_status IN ('succeeded', 'failed', 'canceled', 'outcome_unknown')
            AND completed_at_unix_nano IS NOT NULL
            AND completed_at_unix_nano >= started_at_unix_nano)
    )
) STRICT;

INSERT INTO workflow_run_records_v3 (
    tenant_ref,workspace_id,application_id,run_id,draft_id,draft_version,
    record_version,store_schema_version,schema_version,run_status,
    started_at_unix_nano,completed_at_unix_nano,actor_ref,request_id,audit_ref,
    failure_code,failure_boundary,selected_provider,selected_model,sanitized_run_record
)
SELECT
    tenant_ref,workspace_id,application_id,run_id,draft_id,draft_version,
    record_version,'workflow_runs_store_v3',schema_version,run_status,
    started_at_unix_nano,completed_at_unix_nano,actor_ref,request_id,audit_ref,
    failure_code,failure_boundary,selected_provider,selected_model,sanitized_run_record
FROM workflow_run_records;

DROP TABLE workflow_run_records;
ALTER TABLE workflow_run_records_v3 RENAME TO workflow_run_records;

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

CREATE TABLE workflow_http_tool_execution_attempts (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    confirmation_id TEXT NOT NULL CHECK (length(trim(confirmation_id)) > 0),
    attempt_id TEXT NOT NULL CHECK (
        length(attempt_id) >= 21 AND attempt_id GLOB 'wtea_[a-z0-9]*'
        AND attempt_id NOT GLOB '*[^a-z0-9_]*'
    ),
    run_id TEXT NOT NULL CHECK (
        length(run_id) >= 20 AND run_id GLOB 'run_[a-z0-9]*'
        AND run_id NOT GLOB '*[^a-z0-9_]*'
    ),
    status TEXT NOT NULL CHECK (status IN ('claimed', 'succeeded', 'failed', 'outcome_unknown')),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71
        AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    claimed_at_unix_nano INTEGER NOT NULL CHECK (claimed_at_unix_nano > 0),
    completed_at_unix_nano INTEGER,
    failure_code TEXT NOT NULL DEFAULT '',
    sanitized_execution_attempt TEXT NOT NULL CHECK (
        json_valid(sanitized_execution_attempt)
        AND json_type(sanitized_execution_attempt) = 'object'
    ),
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
        (status = 'claimed' AND completed_at_unix_nano IS NULL AND failure_code = '')
        OR
        (status IN ('succeeded', 'failed', 'outcome_unknown')
            AND completed_at_unix_nano IS NOT NULL
            AND completed_at_unix_nano >= claimed_at_unix_nano)
    )
) STRICT;

CREATE INDEX workflow_http_tool_execution_attempts_status_idx
    ON workflow_http_tool_execution_attempts (
        tenant_ref, workspace_id, application_id,
        status, claimed_at_unix_nano, attempt_id
    );
