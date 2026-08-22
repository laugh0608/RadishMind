PRAGMA defer_foreign_keys = ON;

DROP TRIGGER workflow_http_tool_confirmation_decisions_append_only_update;
DROP TRIGGER workflow_http_tool_confirmation_decisions_append_only_delete;
DROP TRIGGER workflow_http_tool_execution_audits_append_only_update;
DROP TRIGGER workflow_http_tool_execution_audits_append_only_delete;
DROP INDEX workflow_http_tool_execution_attempts_status_idx;
DROP INDEX workflow_http_tool_confirmation_decisions_history_idx;
DROP INDEX workflow_http_tool_execution_audits_history_idx;
DROP INDEX workflow_http_tool_execution_audits_event_idx;
DROP INDEX workflow_http_tool_action_plans_status_expiry_idx;
DROP INDEX workflow_http_tool_action_plans_draft_idx;

ALTER TABLE workflow_http_tool_execution_attempts RENAME TO workflow_http_tool_execution_attempts_pre_definition_sources;
ALTER TABLE workflow_http_tool_confirmation_decisions RENAME TO workflow_http_tool_confirmation_decisions_pre_definition_sources;
ALTER TABLE workflow_http_tool_execution_audits RENAME TO workflow_http_tool_execution_audits_pre_definition_sources;
ALTER TABLE workflow_http_tool_action_plans RENAME TO workflow_http_tool_action_plans_pre_definition_sources;

CREATE TABLE workflow_http_tool_action_plans (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    schema_version TEXT NOT NULL CHECK (schema_version IN (
        'workflow_http_tool_action_plan.v1', 'workflow_http_tool_action_plan.v2'
    )),
    status TEXT NOT NULL CHECK (status IN (
        'pending', 'deferred', 'approved', 'rejected',
        'canceled', 'expired', 'invalidated', 'consumed'
    )),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    source_kind TEXT NOT NULL CHECK (source_kind IN ('saved_workflow_draft', 'workflow_definition')),
    draft_id TEXT,
    draft_version INTEGER,
    workflow_definition_id TEXT,
    workflow_definition_version INTEGER,
    workflow_definition_digest TEXT,
    activation_pointer_version INTEGER,
    node_id TEXT NOT NULL CHECK (length(trim(node_id)) > 0),
    tool_id TEXT NOT NULL CHECK (
        length(tool_id) > 3 AND tool_id NOT GLOB '*[^a-z0-9._-]*' AND instr(tool_id, '.v') > 1
    ),
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    definition_digest TEXT NOT NULL CHECK (
        length(definition_digest) = 71 AND substr(definition_digest, 1, 7) = 'sha256:'
        AND substr(definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    profile_id TEXT NOT NULL CHECK (length(trim(profile_id)) > 0),
    profile_version INTEGER NOT NULL CHECK (profile_version > 0),
    profile_digest TEXT NOT NULL CHECK (
        length(profile_digest) = 71 AND substr(profile_digest, 1, 7) = 'sha256:'
        AND substr(profile_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    target_policy_key TEXT NOT NULL CHECK (length(trim(target_policy_key)) > 0),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71 AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    method TEXT NOT NULL CHECK (method = 'GET'),
    credential_policy TEXT NOT NULL CHECK (credential_policy = 'none'),
    timeout_ms INTEGER NOT NULL CHECK (timeout_ms > 0 AND timeout_ms <= 5000),
    max_response_bytes INTEGER NOT NULL CHECK (max_response_bytes > 0 AND max_response_bytes <= 65536),
    max_output_bytes INTEGER NOT NULL CHECK (max_output_bytes > 0 AND max_output_bytes <= 16384),
    planned_by_actor_ref TEXT NOT NULL CHECK (length(trim(planned_by_actor_ref)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    expires_at_unix_nano INTEGER NOT NULL CHECK (expires_at_unix_nano > created_at_unix_nano),
    sanitized_action_plan TEXT NOT NULL CHECK (
        json_valid(sanitized_action_plan) AND json_type(sanitized_action_plan) = 'object'
        AND json_extract(sanitized_action_plan, '$.schema_version') = schema_version
        AND (
            (schema_version = 'workflow_http_tool_action_plan.v1'
                AND json_type(sanitized_action_plan, '$.source_kind') IS NULL
                AND json_extract(sanitized_action_plan, '$.draft_id') = draft_id
                AND json_extract(sanitized_action_plan, '$.draft_version') = draft_version)
            OR
            (schema_version = 'workflow_http_tool_action_plan.v2'
                AND json_extract(sanitized_action_plan, '$.source_kind') = source_kind
                AND json_extract(sanitized_action_plan, '$.workflow_definition_id') = workflow_definition_id
                AND json_extract(sanitized_action_plan, '$.workflow_definition_version') = workflow_definition_version
                AND json_extract(sanitized_action_plan, '$.workflow_definition_digest') = workflow_definition_digest
                AND json_extract(sanitized_action_plan, '$.activation_pointer_version') = activation_pointer_version)
        )
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id),
    UNIQUE (tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest),
    CHECK (
        (schema_version = 'workflow_http_tool_action_plan.v1' AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND length(trim(draft_id)) > 0 AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_action_plan.v2' AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND length(trim(workflow_definition_id)) > 0
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest IS NOT NULL AND length(workflow_definition_digest) = 71
            AND substr(workflow_definition_digest, 1, 7) = 'sha256:'
            AND substr(workflow_definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    )
) STRICT;

CREATE TABLE workflow_http_tool_execution_audits (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    audit_id TEXT NOT NULL CHECK (length(trim(audit_id)) > 0),
    schema_version TEXT NOT NULL CHECK (schema_version IN (
        'workflow_http_tool_execution_audit.v1', 'workflow_http_tool_execution_audit.v2'
    )),
    event_kind TEXT NOT NULL CHECK (event_kind IN (
        'confirmation_requested', 'confirmation_recorded', 'confirmation_rejected',
        'confirmation_deferred', 'confirmation_canceled', 'confirmation_expired',
        'confirmation_invalidated', 'tool_execution_started', 'tool_execution_succeeded',
        'tool_execution_failed', 'tool_execution_outcome_unknown'
    )),
    source_kind TEXT NOT NULL CHECK (source_kind IN ('saved_workflow_draft', 'workflow_definition')),
    draft_id TEXT,
    draft_version INTEGER,
    workflow_definition_id TEXT,
    workflow_definition_version INTEGER,
    workflow_definition_digest TEXT,
    activation_pointer_version INTEGER,
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71 AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    actor_ref TEXT NOT NULL CHECK (length(trim(actor_ref)) > 0),
    request_id TEXT NOT NULL CHECK (length(trim(request_id)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_execution_audit TEXT NOT NULL CHECK (
        json_valid(sanitized_execution_audit) AND json_type(sanitized_execution_audit) = 'object'
        AND json_extract(sanitized_execution_audit, '$.schema_version') = schema_version
        AND (
            (schema_version = 'workflow_http_tool_execution_audit.v1'
                AND json_type(sanitized_execution_audit, '$.source_kind') IS NULL
                AND json_extract(sanitized_execution_audit, '$.draft_id') = draft_id
                AND json_extract(sanitized_execution_audit, '$.draft_version') = draft_version)
            OR
            (schema_version = 'workflow_http_tool_execution_audit.v2'
                AND json_extract(sanitized_execution_audit, '$.source_kind') = source_kind
                AND json_extract(sanitized_execution_audit, '$.workflow_definition_id') = workflow_definition_id
                AND json_extract(sanitized_execution_audit, '$.workflow_definition_version') = workflow_definition_version
                AND json_extract(sanitized_execution_audit, '$.workflow_definition_digest') = workflow_definition_digest
                AND json_extract(sanitized_execution_audit, '$.activation_pointer_version') = activation_pointer_version)
        )
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id, audit_id),
    UNIQUE (tenant_ref, workspace_id, application_id, plan_id, audit_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest)
        REFERENCES workflow_http_tool_action_plans (
            tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK (
        (schema_version = 'workflow_http_tool_execution_audit.v1' AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND length(trim(draft_id)) > 0 AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_execution_audit.v2' AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND length(trim(workflow_definition_id)) > 0
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest IS NOT NULL AND length(workflow_definition_digest) = 71
            AND substr(workflow_definition_digest, 1, 7) = 'sha256:'
            AND substr(workflow_definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    )
) STRICT;

CREATE TABLE workflow_http_tool_confirmation_decisions (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    confirmation_id TEXT NOT NULL CHECK (length(trim(confirmation_id)) > 0),
    schema_version TEXT NOT NULL CHECK (schema_version IN (
        'workflow_http_tool_confirmation_decision.v1', 'workflow_http_tool_confirmation_decision.v2'
    )),
    outcome TEXT NOT NULL CHECK (outcome IN ('approve', 'reject', 'defer', 'cancel', 'expire', 'invalidate')),
    source_kind TEXT NOT NULL CHECK (source_kind IN ('saved_workflow_draft', 'workflow_definition')),
    draft_id TEXT,
    draft_version INTEGER,
    workflow_definition_id TEXT,
    workflow_definition_version INTEGER,
    workflow_definition_digest TEXT,
    activation_pointer_version INTEGER,
    node_id TEXT NOT NULL CHECK (length(trim(node_id)) > 0),
    tool_id TEXT NOT NULL CHECK (
        length(tool_id) > 3 AND tool_id NOT GLOB '*[^a-z0-9._-]*' AND instr(tool_id, '.v') > 1
    ),
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71 AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    expected_record_version INTEGER NOT NULL CHECK (expected_record_version > 0),
    resulting_record_version INTEGER NOT NULL CHECK (resulting_record_version = expected_record_version + 1),
    decided_by_actor_ref TEXT NOT NULL CHECK (length(trim(decided_by_actor_ref)) > 0),
    actor_source TEXT NOT NULL CHECK (actor_source IN ('human', 'system')),
    reason_code TEXT NOT NULL CHECK (
        length(reason_code) >= 30 AND reason_code GLOB 'workflow_tool_confirmation_[a-z_]*'
        AND reason_code NOT GLOB '*[^a-z_]*'
    ),
    decided_at_unix_nano INTEGER NOT NULL CHECK (decided_at_unix_nano > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    sanitized_confirmation_decision TEXT NOT NULL CHECK (
        json_valid(sanitized_confirmation_decision) AND json_type(sanitized_confirmation_decision) = 'object'
        AND json_extract(sanitized_confirmation_decision, '$.schema_version') = schema_version
        AND (
            (schema_version = 'workflow_http_tool_confirmation_decision.v1'
                AND json_type(sanitized_confirmation_decision, '$.source_kind') IS NULL
                AND json_extract(sanitized_confirmation_decision, '$.draft_id') = draft_id
                AND json_extract(sanitized_confirmation_decision, '$.draft_version') = draft_version)
            OR
            (schema_version = 'workflow_http_tool_confirmation_decision.v2'
                AND json_extract(sanitized_confirmation_decision, '$.source_kind') = source_kind
                AND json_extract(sanitized_confirmation_decision, '$.workflow_definition_id') = workflow_definition_id
                AND json_extract(sanitized_confirmation_decision, '$.workflow_definition_version') = workflow_definition_version
                AND json_extract(sanitized_confirmation_decision, '$.workflow_definition_digest') = workflow_definition_digest
                AND json_extract(sanitized_confirmation_decision, '$.activation_pointer_version') = activation_pointer_version)
        )
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id, confirmation_id),
    UNIQUE (tenant_ref, workspace_id, application_id, plan_id, resulting_record_version),
    UNIQUE (tenant_ref, workspace_id, application_id, plan_id, audit_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest)
        REFERENCES workflow_http_tool_action_plans (
            tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_ref, workspace_id, application_id, plan_id, audit_ref)
        REFERENCES workflow_http_tool_execution_audits (
            tenant_ref, workspace_id, application_id, plan_id, audit_ref
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK (
        (schema_version = 'workflow_http_tool_confirmation_decision.v1' AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND length(trim(draft_id)) > 0 AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_confirmation_decision.v2' AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND length(trim(workflow_definition_id)) > 0
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest IS NOT NULL AND length(workflow_definition_digest) = 71
            AND substr(workflow_definition_digest, 1, 7) = 'sha256:'
            AND substr(workflow_definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    )
) STRICT;

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
        length(tool_plan_digest) = 71 AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    claimed_at_unix_nano INTEGER NOT NULL CHECK (claimed_at_unix_nano > 0),
    completed_at_unix_nano INTEGER,
    failure_code TEXT NOT NULL DEFAULT '',
    sanitized_execution_attempt TEXT NOT NULL CHECK (
        json_valid(sanitized_execution_attempt) AND json_type(sanitized_execution_attempt) = 'object'
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
        OR (status IN ('succeeded', 'failed', 'outcome_unknown')
            AND completed_at_unix_nano IS NOT NULL AND completed_at_unix_nano >= claimed_at_unix_nano)
    )
) STRICT;

INSERT INTO workflow_http_tool_action_plans (
    tenant_ref,workspace_id,application_id,plan_id,schema_version,status,record_version,source_kind,draft_id,draft_version,
    workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,
    node_id,tool_id,tool_version,definition_digest,profile_id,profile_version,profile_digest,target_policy_key,tool_plan_digest,
    method,credential_policy,timeout_ms,max_response_bytes,max_output_bytes,planned_by_actor_ref,audit_ref,
    created_at_unix_nano,expires_at_unix_nano,sanitized_action_plan
)
SELECT tenant_ref,workspace_id,application_id,plan_id,schema_version,status,record_version,'saved_workflow_draft',draft_id,draft_version,
    NULL,NULL,NULL,NULL,node_id,tool_id,tool_version,definition_digest,profile_id,profile_version,profile_digest,target_policy_key,tool_plan_digest,
    method,credential_policy,timeout_ms,max_response_bytes,max_output_bytes,planned_by_actor_ref,audit_ref,
    created_at_unix_nano,expires_at_unix_nano,sanitized_action_plan
FROM workflow_http_tool_action_plans_pre_definition_sources;

INSERT INTO workflow_http_tool_execution_audits (
    tenant_ref,workspace_id,application_id,plan_id,audit_id,schema_version,event_kind,source_kind,draft_id,draft_version,
    workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,
    tool_version,tool_plan_digest,actor_ref,request_id,audit_ref,occurred_at_unix_nano,sanitized_execution_audit
)
SELECT tenant_ref,workspace_id,application_id,plan_id,audit_id,schema_version,event_kind,'saved_workflow_draft',
    json_extract(sanitized_execution_audit, '$.draft_id'),json_extract(sanitized_execution_audit, '$.draft_version'),
    NULL,NULL,NULL,NULL,tool_version,tool_plan_digest,actor_ref,request_id,audit_ref,occurred_at_unix_nano,sanitized_execution_audit
FROM workflow_http_tool_execution_audits_pre_definition_sources;

INSERT INTO workflow_http_tool_confirmation_decisions (
    tenant_ref,workspace_id,application_id,plan_id,confirmation_id,schema_version,outcome,source_kind,draft_id,draft_version,
    workflow_definition_id,workflow_definition_version,workflow_definition_digest,activation_pointer_version,
    node_id,tool_id,tool_version,tool_plan_digest,expected_record_version,resulting_record_version,decided_by_actor_ref,
    actor_source,reason_code,decided_at_unix_nano,audit_ref,sanitized_confirmation_decision
)
SELECT tenant_ref,workspace_id,application_id,plan_id,confirmation_id,schema_version,outcome,'saved_workflow_draft',draft_id,draft_version,
    NULL,NULL,NULL,NULL,node_id,tool_id,tool_version,tool_plan_digest,expected_record_version,resulting_record_version,decided_by_actor_ref,
    actor_source,reason_code,decided_at_unix_nano,audit_ref,sanitized_confirmation_decision
FROM workflow_http_tool_confirmation_decisions_pre_definition_sources;

INSERT INTO workflow_http_tool_execution_attempts
SELECT * FROM workflow_http_tool_execution_attempts_pre_definition_sources;

DROP TABLE workflow_http_tool_execution_attempts_pre_definition_sources;
DROP TABLE workflow_http_tool_confirmation_decisions_pre_definition_sources;
DROP TABLE workflow_http_tool_execution_audits_pre_definition_sources;
DROP TABLE workflow_http_tool_action_plans_pre_definition_sources;

CREATE INDEX workflow_http_tool_action_plans_status_expiry_idx
    ON workflow_http_tool_action_plans (
        tenant_ref,workspace_id,application_id,status,expires_at_unix_nano,plan_id
    );
CREATE INDEX workflow_http_tool_action_plans_draft_idx
    ON workflow_http_tool_action_plans (
        tenant_ref,workspace_id,application_id,draft_id,draft_version,node_id,created_at_unix_nano DESC
    ) WHERE source_kind = 'saved_workflow_draft';
CREATE INDEX workflow_http_tool_action_plans_definition_idx
    ON workflow_http_tool_action_plans (
        tenant_ref,workspace_id,application_id,workflow_definition_id,
        workflow_definition_version,activation_pointer_version,node_id,created_at_unix_nano DESC
    ) WHERE source_kind = 'workflow_definition';
CREATE INDEX workflow_http_tool_confirmation_decisions_history_idx
    ON workflow_http_tool_confirmation_decisions (
        tenant_ref,workspace_id,application_id,plan_id,resulting_record_version DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_history_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref,workspace_id,application_id,plan_id,occurred_at_unix_nano DESC,audit_id DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_event_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref,workspace_id,application_id,event_kind,occurred_at_unix_nano DESC,audit_id
    );
CREATE INDEX workflow_http_tool_execution_attempts_status_idx
    ON workflow_http_tool_execution_attempts (
        tenant_ref,workspace_id,application_id,status,claimed_at_unix_nano,attempt_id
    );

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only_update
BEFORE UPDATE ON workflow_http_tool_confirmation_decisions
BEGIN SELECT RAISE(ABORT, 'workflow HTTP tool confirmation decisions are append-only'); END;
CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only_delete
BEFORE DELETE ON workflow_http_tool_confirmation_decisions
BEGIN SELECT RAISE(ABORT, 'workflow HTTP tool confirmation decisions are append-only'); END;
CREATE TRIGGER workflow_http_tool_execution_audits_append_only_update
BEFORE UPDATE ON workflow_http_tool_execution_audits
BEGIN SELECT RAISE(ABORT, 'workflow HTTP tool execution audits are append-only'); END;
CREATE TRIGGER workflow_http_tool_execution_audits_append_only_delete
BEFORE DELETE ON workflow_http_tool_execution_audits
BEGIN SELECT RAISE(ABORT, 'workflow HTTP tool execution audits are append-only'); END;
