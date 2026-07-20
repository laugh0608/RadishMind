CREATE TABLE workflow_http_tool_action_plans (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    plan_id text NOT NULL CHECK (btrim(plan_id) <> ''),
    schema_version text NOT NULL CHECK (schema_version = 'workflow_http_tool_action_plan.v1'),
    status text NOT NULL CHECK (status IN (
        'pending', 'deferred', 'approved', 'rejected',
        'canceled', 'expired', 'invalidated', 'consumed'
    )),
    record_version bigint NOT NULL CHECK (record_version > 0),
    draft_id text NOT NULL CHECK (btrim(draft_id) <> ''),
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    node_id text NOT NULL CHECK (btrim(node_id) <> ''),
    tool_id text NOT NULL CHECK (
        tool_id ~ '^[a-z0-9][a-z0-9._-]*\.v[0-9]+$'
    ),
    tool_version integer NOT NULL CHECK (tool_version = 1),
    definition_digest text NOT NULL CHECK (
        definition_digest ~ '^sha256:[0-9a-f]{64}$'
    ),
    profile_id text NOT NULL CHECK (btrim(profile_id) <> ''),
    profile_version integer NOT NULL CHECK (profile_version > 0),
    profile_digest text NOT NULL CHECK (
        profile_digest ~ '^sha256:[0-9a-f]{64}$'
    ),
    target_policy_key text NOT NULL CHECK (btrim(target_policy_key) <> ''),
    tool_plan_digest text NOT NULL CHECK (
        tool_plan_digest ~ '^sha256:[0-9a-f]{64}$'
    ),
    method text NOT NULL CHECK (method = 'GET'),
    credential_policy text NOT NULL CHECK (credential_policy = 'none'),
    timeout_ms integer NOT NULL CHECK (
        timeout_ms > 0 AND timeout_ms <= 5000
    ),
    max_response_bytes integer NOT NULL CHECK (
        max_response_bytes > 0 AND max_response_bytes <= 65536
    ),
    max_output_bytes integer NOT NULL CHECK (
        max_output_bytes > 0 AND max_output_bytes <= 16384
    ),
    planned_by_actor_ref text NOT NULL CHECK (btrim(planned_by_actor_ref) <> ''),
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    sanitized_action_plan jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_action_plan) = 'object'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, plan_id),
    UNIQUE (
        tenant_ref, workspace_id, application_id, plan_id, tool_plan_digest
    ),
    CHECK (expires_at > created_at)
);

CREATE TABLE workflow_http_tool_execution_audits (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    plan_id text NOT NULL CHECK (btrim(plan_id) <> ''),
    audit_id text NOT NULL CHECK (btrim(audit_id) <> ''),
    schema_version text NOT NULL CHECK (schema_version = 'workflow_http_tool_execution_audit.v1'),
    event_kind text NOT NULL CHECK (event_kind IN (
        'confirmation_requested', 'confirmation_recorded',
        'confirmation_rejected', 'confirmation_deferred',
        'confirmation_canceled', 'confirmation_expired',
        'confirmation_invalidated', 'tool_execution_started',
        'tool_execution_succeeded', 'tool_execution_failed',
        'tool_execution_outcome_unknown'
    )),
    tool_version integer NOT NULL CHECK (tool_version = 1),
    tool_plan_digest text NOT NULL CHECK (
        tool_plan_digest ~ '^sha256:[0-9a-f]{64}$'
    ),
    actor_ref text NOT NULL CHECK (btrim(actor_ref) <> ''),
    request_id text NOT NULL CHECK (btrim(request_id) <> ''),
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    occurred_at timestamptz NOT NULL,
    sanitized_execution_audit jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_execution_audit) = 'object'
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
);

CREATE TABLE workflow_http_tool_confirmation_decisions (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    plan_id text NOT NULL CHECK (btrim(plan_id) <> ''),
    confirmation_id text NOT NULL CHECK (btrim(confirmation_id) <> ''),
    schema_version text NOT NULL CHECK (schema_version = 'workflow_http_tool_confirmation_decision.v1'),
    outcome text NOT NULL CHECK (
        outcome IN ('approve', 'reject', 'defer', 'cancel', 'expire', 'invalidate')
    ),
    draft_id text NOT NULL CHECK (btrim(draft_id) <> ''),
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    node_id text NOT NULL CHECK (btrim(node_id) <> ''),
    tool_id text NOT NULL CHECK (
        tool_id ~ '^[a-z0-9][a-z0-9._-]*\.v[0-9]+$'
    ),
    tool_version integer NOT NULL CHECK (tool_version = 1),
    tool_plan_digest text NOT NULL CHECK (
        tool_plan_digest ~ '^sha256:[0-9a-f]{64}$'
    ),
    expected_record_version bigint NOT NULL CHECK (
        expected_record_version > 0
    ),
    resulting_record_version bigint NOT NULL CHECK (
        resulting_record_version = expected_record_version + 1
    ),
    decided_by_actor_ref text NOT NULL CHECK (btrim(decided_by_actor_ref) <> ''),
    actor_source text NOT NULL CHECK (actor_source IN ('human', 'system')),
    reason_code text NOT NULL CHECK (reason_code ~ '^workflow_tool_confirmation_[a-z_]{3,64}$'),
    decided_at timestamptz NOT NULL,
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    sanitized_confirmation_decision jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_confirmation_decision) = 'object'
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
);

CREATE INDEX workflow_http_tool_action_plans_status_expiry_idx
    ON workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id,
        status, expires_at, plan_id
    );
CREATE INDEX workflow_http_tool_action_plans_draft_idx
    ON workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id,
        draft_id, draft_version, node_id, created_at DESC
    );
CREATE INDEX workflow_http_tool_confirmation_decisions_history_idx
    ON workflow_http_tool_confirmation_decisions (
        tenant_ref, workspace_id, application_id,
        plan_id, resulting_record_version DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_history_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref, workspace_id, application_id,
        plan_id, occurred_at DESC, audit_id DESC
    );
CREATE INDEX workflow_http_tool_execution_audits_event_idx
    ON workflow_http_tool_execution_audits (
        tenant_ref, workspace_id, application_id,
        event_kind, occurred_at DESC, audit_id
    );

CREATE FUNCTION reject_workflow_http_tool_append_only_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'workflow HTTP tool decision and audit records are append-only';
END;
$$;

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_confirmation_decisions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();

CREATE TRIGGER workflow_http_tool_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
