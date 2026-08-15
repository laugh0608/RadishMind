DROP TRIGGER workflow_http_tool_confirmation_decisions_append_only
    ON workflow_http_tool_confirmation_decisions;
DROP TRIGGER workflow_http_tool_execution_audits_append_only
    ON workflow_http_tool_execution_audits;

ALTER TABLE workflow_http_tool_action_plans
    DROP CONSTRAINT workflow_http_tool_action_plans_schema_version_check,
    DROP CONSTRAINT workflow_http_tool_action_plans_draft_id_check,
    DROP CONSTRAINT workflow_http_tool_action_plans_draft_version_check,
    ALTER COLUMN draft_id DROP NOT NULL,
    ALTER COLUMN draft_version DROP NOT NULL,
    ADD COLUMN source_kind text,
    ADD COLUMN workflow_definition_id text,
    ADD COLUMN workflow_definition_version bigint,
    ADD COLUMN workflow_definition_digest text,
    ADD COLUMN activation_pointer_version bigint;

UPDATE workflow_http_tool_action_plans
SET source_kind = 'saved_workflow_draft';

ALTER TABLE workflow_http_tool_action_plans
    ALTER COLUMN source_kind SET NOT NULL,
    ADD CONSTRAINT workflow_http_tool_action_plans_schema_version_v2_check CHECK (
        schema_version IN ('workflow_http_tool_action_plan.v1', 'workflow_http_tool_action_plan.v2')
    ),
    ADD CONSTRAINT workflow_http_tool_action_plans_source_union_check CHECK (
        (schema_version = 'workflow_http_tool_action_plan.v1'
            AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND btrim(draft_id) <> ''
            AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_action_plan.v2'
            AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND btrim(workflow_definition_id) <> ''
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest ~ '^sha256:[0-9a-f]{64}$'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    ),
    ADD CONSTRAINT workflow_http_tool_action_plans_source_payload_check CHECK (
        sanitized_action_plan->>'schema_version' = schema_version
        AND (
            (schema_version = 'workflow_http_tool_action_plan.v1'
                AND NOT (sanitized_action_plan ? 'source_kind')
                AND sanitized_action_plan->>'draft_id' = draft_id
                AND (sanitized_action_plan->>'draft_version')::bigint = draft_version)
            OR
            (schema_version = 'workflow_http_tool_action_plan.v2'
                AND sanitized_action_plan->>'source_kind' = source_kind
                AND sanitized_action_plan->>'workflow_definition_id' = workflow_definition_id
                AND (sanitized_action_plan->>'workflow_definition_version')::bigint = workflow_definition_version
                AND sanitized_action_plan->>'workflow_definition_digest' = workflow_definition_digest
                AND (sanitized_action_plan->>'activation_pointer_version')::bigint = activation_pointer_version)
        )
    );

CREATE INDEX workflow_http_tool_action_plans_definition_idx
    ON workflow_http_tool_action_plans (
        tenant_ref, workspace_id, application_id, workflow_definition_id,
        workflow_definition_version, activation_pointer_version, node_id, created_at DESC
    ) WHERE source_kind = 'workflow_definition';

ALTER TABLE workflow_http_tool_execution_audits
    DROP CONSTRAINT workflow_http_tool_execution_audits_schema_version_check,
    ADD COLUMN source_kind text,
    ADD COLUMN draft_id text,
    ADD COLUMN draft_version bigint,
    ADD COLUMN workflow_definition_id text,
    ADD COLUMN workflow_definition_version bigint,
    ADD COLUMN workflow_definition_digest text,
    ADD COLUMN activation_pointer_version bigint;

UPDATE workflow_http_tool_execution_audits
SET source_kind = 'saved_workflow_draft',
    draft_id = sanitized_execution_audit->>'draft_id',
    draft_version = (sanitized_execution_audit->>'draft_version')::bigint;

ALTER TABLE workflow_http_tool_execution_audits
    ALTER COLUMN source_kind SET NOT NULL,
    ADD CONSTRAINT workflow_http_tool_execution_audits_schema_version_v2_check CHECK (
        schema_version IN ('workflow_http_tool_execution_audit.v1', 'workflow_http_tool_execution_audit.v2')
    ),
    ADD CONSTRAINT workflow_http_tool_execution_audits_source_union_check CHECK (
        (schema_version = 'workflow_http_tool_execution_audit.v1'
            AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND btrim(draft_id) <> ''
            AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_execution_audit.v2'
            AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND btrim(workflow_definition_id) <> ''
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest ~ '^sha256:[0-9a-f]{64}$'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    ),
    ADD CONSTRAINT workflow_http_tool_execution_audits_source_payload_check CHECK (
        sanitized_execution_audit->>'schema_version' = schema_version
        AND (
            (schema_version = 'workflow_http_tool_execution_audit.v1'
                AND NOT (sanitized_execution_audit ? 'source_kind')
                AND sanitized_execution_audit->>'draft_id' = draft_id
                AND (sanitized_execution_audit->>'draft_version')::bigint = draft_version)
            OR
            (schema_version = 'workflow_http_tool_execution_audit.v2'
                AND sanitized_execution_audit->>'source_kind' = source_kind
                AND sanitized_execution_audit->>'workflow_definition_id' = workflow_definition_id
                AND (sanitized_execution_audit->>'workflow_definition_version')::bigint = workflow_definition_version
                AND sanitized_execution_audit->>'workflow_definition_digest' = workflow_definition_digest
                AND (sanitized_execution_audit->>'activation_pointer_version')::bigint = activation_pointer_version)
        )
    );

ALTER TABLE workflow_http_tool_confirmation_decisions
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_schema_version_check,
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_draft_id_check,
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_draft_version_check,
    ALTER COLUMN draft_id DROP NOT NULL,
    ALTER COLUMN draft_version DROP NOT NULL,
    ADD COLUMN source_kind text,
    ADD COLUMN workflow_definition_id text,
    ADD COLUMN workflow_definition_version bigint,
    ADD COLUMN workflow_definition_digest text,
    ADD COLUMN activation_pointer_version bigint;

UPDATE workflow_http_tool_confirmation_decisions
SET source_kind = 'saved_workflow_draft';

ALTER TABLE workflow_http_tool_confirmation_decisions
    ALTER COLUMN source_kind SET NOT NULL,
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_schema_version_v2_check CHECK (
        schema_version IN ('workflow_http_tool_confirmation_decision.v1', 'workflow_http_tool_confirmation_decision.v2')
    ),
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_source_union_check CHECK (
        (schema_version = 'workflow_http_tool_confirmation_decision.v1'
            AND source_kind = 'saved_workflow_draft'
            AND draft_id IS NOT NULL AND btrim(draft_id) <> ''
            AND draft_version IS NOT NULL AND draft_version > 0
            AND workflow_definition_id IS NULL AND workflow_definition_version IS NULL
            AND workflow_definition_digest IS NULL AND activation_pointer_version IS NULL)
        OR
        (schema_version = 'workflow_http_tool_confirmation_decision.v2'
            AND source_kind = 'workflow_definition'
            AND draft_id IS NULL AND draft_version IS NULL
            AND workflow_definition_id IS NOT NULL AND btrim(workflow_definition_id) <> ''
            AND workflow_definition_version IS NOT NULL AND workflow_definition_version > 0
            AND workflow_definition_digest ~ '^sha256:[0-9a-f]{64}$'
            AND activation_pointer_version IS NOT NULL AND activation_pointer_version > 0)
    ),
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_source_payload_check CHECK (
        sanitized_confirmation_decision->>'schema_version' = schema_version
        AND (
            (schema_version = 'workflow_http_tool_confirmation_decision.v1'
                AND NOT (sanitized_confirmation_decision ? 'source_kind')
                AND sanitized_confirmation_decision->>'draft_id' = draft_id
                AND (sanitized_confirmation_decision->>'draft_version')::bigint = draft_version)
            OR
            (schema_version = 'workflow_http_tool_confirmation_decision.v2'
                AND sanitized_confirmation_decision->>'source_kind' = source_kind
                AND sanitized_confirmation_decision->>'workflow_definition_id' = workflow_definition_id
                AND (sanitized_confirmation_decision->>'workflow_definition_version')::bigint = workflow_definition_version
                AND sanitized_confirmation_decision->>'workflow_definition_digest' = workflow_definition_digest
                AND (sanitized_confirmation_decision->>'activation_pointer_version')::bigint = activation_pointer_version)
        )
    );

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_confirmation_decisions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
CREATE TRIGGER workflow_http_tool_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
