DELETE FROM workflow_run_records
WHERE schema_version = 'workflow_run_record.v5';

ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_run_records_execution_source_check;

ALTER TABLE workflow_run_records
    ADD CONSTRAINT workflow_run_records_execution_source_check CHECK (
        execution_source_kind IN ('workflow_draft', 'application_configuration_draft')
        AND btrim(execution_source_id) <> ''
        AND execution_source_version > 0
        AND (
            (schema_version IN ('workflow_run_record.v0', 'workflow_run_record.v1', 'workflow_run_record.v2', 'workflow_run_record.v3')
                AND execution_source_kind = 'workflow_draft')
            OR (schema_version = 'workflow_run_record.v4'
                AND execution_source_kind = 'application_configuration_draft')
        )
    );
