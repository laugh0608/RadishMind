ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_run_records_execution_source_check;

ALTER TABLE workflow_run_records
    ADD COLUMN input_contract_id text NOT NULL DEFAULT '',
    ADD COLUMN input_contract_digest text NOT NULL DEFAULT '';

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
    ),
    ADD CONSTRAINT workflow_run_records_structured_input_projection_check CHECK (
        (schema_version = 'workflow_run_record.v8'
            AND input_contract_id ~ '^[A-Za-z0-9][A-Za-z0-9._:-]{0,159}$'
            AND input_contract_digest ~ '^sha256:[a-f0-9]{64}$'
            AND sanitized_run_record->>'schema_version' = schema_version
            AND sanitized_run_record->>'input_contract_id' = input_contract_id
            AND sanitized_run_record->>'input_contract_digest' = input_contract_digest)
        OR
        (schema_version <> 'workflow_run_record.v8'
            AND input_contract_id = ''
            AND input_contract_digest = '')
    );
