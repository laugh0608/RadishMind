DROP TRIGGER workflow_http_tool_confirmation_decisions_append_only
    ON workflow_http_tool_confirmation_decisions;
DROP TRIGGER workflow_http_tool_execution_audits_append_only
    ON workflow_http_tool_execution_audits;

DELETE FROM workflow_http_tool_confirmation_decisions
WHERE schema_version = 'workflow_http_tool_confirmation_decision.v2';
DELETE FROM workflow_http_tool_execution_audits
WHERE schema_version = 'workflow_http_tool_execution_audit.v2';
DELETE FROM workflow_http_tool_action_plans
WHERE schema_version = 'workflow_http_tool_action_plan.v2';

DROP INDEX workflow_http_tool_action_plans_definition_idx;

ALTER TABLE workflow_http_tool_confirmation_decisions
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_source_payload_check,
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_source_union_check,
    DROP CONSTRAINT workflow_http_tool_confirmation_decisions_schema_version_v2_check,
    ALTER COLUMN draft_id SET NOT NULL,
    ALTER COLUMN draft_version SET NOT NULL,
    DROP COLUMN activation_pointer_version,
    DROP COLUMN workflow_definition_digest,
    DROP COLUMN workflow_definition_version,
    DROP COLUMN workflow_definition_id,
    DROP COLUMN source_kind,
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_schema_version_check CHECK (
        schema_version = 'workflow_http_tool_confirmation_decision.v1'
    ),
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_draft_id_check CHECK (btrim(draft_id) <> ''),
    ADD CONSTRAINT workflow_http_tool_confirmation_decisions_draft_version_check CHECK (draft_version > 0);

ALTER TABLE workflow_http_tool_execution_audits
    DROP CONSTRAINT workflow_http_tool_execution_audits_source_payload_check,
    DROP CONSTRAINT workflow_http_tool_execution_audits_source_union_check,
    DROP CONSTRAINT workflow_http_tool_execution_audits_schema_version_v2_check,
    DROP COLUMN activation_pointer_version,
    DROP COLUMN workflow_definition_digest,
    DROP COLUMN workflow_definition_version,
    DROP COLUMN workflow_definition_id,
    DROP COLUMN draft_version,
    DROP COLUMN draft_id,
    DROP COLUMN source_kind,
    ADD CONSTRAINT workflow_http_tool_execution_audits_schema_version_check CHECK (
        schema_version = 'workflow_http_tool_execution_audit.v1'
    );

ALTER TABLE workflow_http_tool_action_plans
    DROP CONSTRAINT workflow_http_tool_action_plans_source_payload_check,
    DROP CONSTRAINT workflow_http_tool_action_plans_source_union_check,
    DROP CONSTRAINT workflow_http_tool_action_plans_schema_version_v2_check,
    ALTER COLUMN draft_id SET NOT NULL,
    ALTER COLUMN draft_version SET NOT NULL,
    DROP COLUMN activation_pointer_version,
    DROP COLUMN workflow_definition_digest,
    DROP COLUMN workflow_definition_version,
    DROP COLUMN workflow_definition_id,
    DROP COLUMN source_kind,
    ADD CONSTRAINT workflow_http_tool_action_plans_schema_version_check CHECK (
        schema_version = 'workflow_http_tool_action_plan.v1'
    ),
    ADD CONSTRAINT workflow_http_tool_action_plans_draft_id_check CHECK (btrim(draft_id) <> ''),
    ADD CONSTRAINT workflow_http_tool_action_plans_draft_version_check CHECK (draft_version > 0);

CREATE TRIGGER workflow_http_tool_confirmation_decisions_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_confirmation_decisions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
CREATE TRIGGER workflow_http_tool_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
