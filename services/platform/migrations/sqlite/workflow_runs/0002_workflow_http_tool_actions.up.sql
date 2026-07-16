CREATE TABLE workflow_http_tool_action_plans (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    schema_version TEXT NOT NULL CHECK (
        schema_version = 'workflow_http_tool_action_plan.v1'
    ),
    status TEXT NOT NULL CHECK (status IN (
        'pending', 'deferred', 'approved', 'rejected',
        'canceled', 'expired', 'invalidated', 'consumed'
    )),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    draft_id TEXT NOT NULL CHECK (length(trim(draft_id)) > 0),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    node_id TEXT NOT NULL CHECK (length(trim(node_id)) > 0),
    tool_id TEXT NOT NULL CHECK (
        length(tool_id) > 3
        AND tool_id NOT GLOB '*[^a-z0-9._-]*'
        AND instr(tool_id, '.v') > 1
    ),
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    definition_digest TEXT NOT NULL CHECK (
        length(definition_digest) = 71
        AND substr(definition_digest, 1, 7) = 'sha256:'
        AND substr(definition_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    profile_id TEXT NOT NULL CHECK (length(trim(profile_id)) > 0),
    profile_version INTEGER NOT NULL CHECK (profile_version > 0),
    profile_digest TEXT NOT NULL CHECK (
        length(profile_digest) = 71
        AND substr(profile_digest, 1, 7) = 'sha256:'
        AND substr(profile_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    target_policy_key TEXT NOT NULL CHECK (length(trim(target_policy_key)) > 0),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71
        AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    method TEXT NOT NULL CHECK (method = 'GET'),
    credential_policy TEXT NOT NULL CHECK (credential_policy = 'none'),
    timeout_ms INTEGER NOT NULL CHECK (
        timeout_ms > 0 AND timeout_ms <= 5000
    ),
    max_response_bytes INTEGER NOT NULL CHECK (
        max_response_bytes > 0 AND max_response_bytes <= 65536
    ),
    max_output_bytes INTEGER NOT NULL CHECK (
        max_output_bytes > 0 AND max_output_bytes <= 16384
    ),
    planned_by_actor_ref TEXT NOT NULL CHECK (
        length(trim(planned_by_actor_ref)) > 0
    ),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    expires_at_unix_nano INTEGER NOT NULL CHECK (
        expires_at_unix_nano > created_at_unix_nano
    ),
    sanitized_action_plan TEXT NOT NULL CHECK (
        json_valid(sanitized_action_plan)
        AND json_type(sanitized_action_plan) = 'object'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id),
    UNIQUE (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    )
) STRICT;

CREATE TABLE workflow_http_tool_execution_audits (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    audit_id TEXT NOT NULL CHECK (length(trim(audit_id)) > 0),
    schema_version TEXT NOT NULL CHECK (
        schema_version = 'workflow_http_tool_execution_audit.v1'
    ),
    event_kind TEXT NOT NULL CHECK (event_kind IN (
        'confirmation_requested', 'confirmation_recorded',
        'confirmation_rejected', 'confirmation_deferred',
        'confirmation_canceled', 'confirmation_expired',
        'confirmation_invalidated', 'tool_execution_started',
        'tool_execution_succeeded', 'tool_execution_failed',
        'tool_execution_outcome_unknown'
    )),
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71
        AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    actor_ref TEXT NOT NULL CHECK (length(trim(actor_ref)) > 0),
    request_id TEXT NOT NULL CHECK (length(trim(request_id)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_execution_audit TEXT NOT NULL CHECK (
        json_valid(sanitized_execution_audit)
        AND json_type(sanitized_execution_audit) = 'object'
    ),
    PRIMARY KEY (
        tenant_ref, workspace_id, application_id, plan_id, audit_id
    ),
    UNIQUE (
        tenant_ref, workspace_id, application_id, plan_id, audit_ref
    ),
    FOREIGN KEY (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    ) REFERENCES workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_http_tool_confirmation_decisions (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    plan_id TEXT NOT NULL CHECK (length(trim(plan_id)) > 0),
    confirmation_id TEXT NOT NULL CHECK (length(trim(confirmation_id)) > 0),
    schema_version TEXT NOT NULL CHECK (
        schema_version = 'workflow_http_tool_confirmation_decision.v1'
    ),
    outcome TEXT NOT NULL CHECK (
        outcome IN ('approve', 'reject', 'defer', 'cancel', 'expire', 'invalidate')
    ),
    draft_id TEXT NOT NULL CHECK (length(trim(draft_id)) > 0),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    node_id TEXT NOT NULL CHECK (length(trim(node_id)) > 0),
    tool_id TEXT NOT NULL CHECK (
        length(tool_id) > 3
        AND tool_id NOT GLOB '*[^a-z0-9._-]*'
        AND instr(tool_id, '.v') > 1
    ),
    tool_version INTEGER NOT NULL CHECK (tool_version = 1),
    tool_plan_digest TEXT NOT NULL CHECK (
        length(tool_plan_digest) = 71
        AND substr(tool_plan_digest, 1, 7) = 'sha256:'
        AND substr(tool_plan_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    expected_record_version INTEGER NOT NULL CHECK (
        expected_record_version > 0
    ),
    resulting_record_version INTEGER NOT NULL CHECK (
        resulting_record_version = expected_record_version + 1
    ),
    decided_by_actor_ref TEXT NOT NULL CHECK (
        length(trim(decided_by_actor_ref)) > 0
    ),
    actor_source TEXT NOT NULL CHECK (actor_source IN ('human', 'system')),
    reason_code TEXT NOT NULL CHECK (
        length(reason_code) >= 30
        AND reason_code GLOB 'workflow_tool_confirmation_[a-z_]*'
        AND reason_code NOT GLOB '*[^a-z_]*'
    ),
    decided_at_unix_nano INTEGER NOT NULL CHECK (decided_at_unix_nano > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    sanitized_confirmation_decision TEXT NOT NULL CHECK (
        json_valid(sanitized_confirmation_decision)
        AND json_type(sanitized_confirmation_decision) = 'object'
    ),
    PRIMARY KEY (
        tenant_ref, workspace_id, application_id, plan_id, confirmation_id
    ),
    UNIQUE (
        tenant_ref, workspace_id, application_id, plan_id, resulting_record_version
    ),
    UNIQUE (
        tenant_ref, workspace_id, application_id, plan_id, audit_ref
    ),
    FOREIGN KEY (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    ) REFERENCES workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (
        tenant_ref, workspace_id, application_id, plan_id, audit_ref
    ) REFERENCES workflow_http_tool_execution_audits (
        tenant_ref, workspace_id, application_id, plan_id, audit_ref
    ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE INDEX workflow_http_tool_action_plans_status_expiry_idx
    ON workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id,
        status, expires_at_unix_nano, plan_id
    );
CREATE INDEX workflow_http_tool_action_plans_draft_idx
    ON workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id,
        draft_id, draft_version, node_id, created_at_unix_nano DESC
    );
CREATE INDEX workflow_http_tool_confirmation_decisions_history_idx
    ON workflow_http_tool_confirmation_decisions (
        tenant_ref, workspace_id, application_id,
        plan_id, resulting_record_version DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_history_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref, workspace_id, application_id,
        plan_id, occurred_at_unix_nano DESC, audit_id DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_event_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref, workspace_id, application_id,
        event_kind, occurred_at_unix_nano DESC, audit_id
    );

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only_update
BEFORE UPDATE ON workflow_http_tool_confirmation_decisions
BEGIN
    SELECT RAISE(ABORT, 'workflow HTTP tool confirmation decisions are append-only');
END;

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only_delete
BEFORE DELETE ON workflow_http_tool_confirmation_decisions
BEGIN
    SELECT RAISE(ABORT, 'workflow HTTP tool confirmation decisions are append-only');
END;

CREATE TRIGGER workflow_http_tool_execution_audits_append_only_update
BEFORE UPDATE ON workflow_http_tool_execution_audits
BEGIN
    SELECT RAISE(ABORT, 'workflow HTTP tool execution audits are append-only');
END;

CREATE TRIGGER workflow_http_tool_execution_audits_append_only_delete
BEFORE DELETE ON workflow_http_tool_execution_audits
BEGIN
    SELECT RAISE(ABORT, 'workflow HTTP tool execution audits are append-only');
END;
