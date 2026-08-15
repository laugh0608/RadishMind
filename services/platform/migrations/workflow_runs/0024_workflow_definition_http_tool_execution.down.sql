DROP TRIGGER workflow_http_tool_execution_audits_append_only
    ON workflow_http_tool_execution_audits;

DELETE FROM workflow_http_tool_execution_audits audit
WHERE EXISTS (
    SELECT 1
    FROM workflow_run_records run
    WHERE run.schema_version = 'workflow_run_record.v9'
      AND run.tenant_ref = audit.tenant_ref
      AND run.workspace_id = audit.workspace_id
      AND run.application_id = audit.application_id
      AND run.run_id = audit.sanitized_execution_audit->>'run_id'
);

DELETE FROM workflow_http_tool_execution_attempts attempt
WHERE EXISTS (
    SELECT 1
    FROM workflow_run_records run
    WHERE run.schema_version = 'workflow_run_record.v9'
      AND run.tenant_ref = attempt.tenant_ref
      AND run.workspace_id = attempt.workspace_id
      AND run.application_id = attempt.application_id
      AND run.run_id = attempt.run_id
);

DELETE FROM workflow_run_records
WHERE schema_version = 'workflow_run_record.v9';

ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_run_records_execution_source_check;

ALTER TABLE workflow_run_records
    ADD CONSTRAINT workflow_run_records_execution_source_check CHECK (
        execution_source_kind IN ('workflow_draft', 'application_configuration_draft', 'workflow_definition')
        AND btrim(execution_source_id) <> ''
        AND execution_source_version > 0
        AND (
            (schema_version IN ('workflow_run_record.v0', 'workflow_run_record.v1', 'workflow_run_record.v2', 'workflow_run_record.v3')
                AND execution_source_kind = 'workflow_draft')
            OR (schema_version = 'workflow_run_record.v4'
                AND execution_source_kind = 'application_configuration_draft')
            OR (schema_version IN ('workflow_run_record.v5', 'workflow_run_record.v8')
                AND execution_source_kind = 'workflow_definition')
        )
    );

CREATE TRIGGER workflow_http_tool_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_http_tool_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_http_tool_append_only_mutation();
