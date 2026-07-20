DROP TABLE IF EXISTS workflow_http_tool_execution_attempts;

DELETE FROM workflow_run_records
WHERE schema_version = 'workflow_run_record.v2';

ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_run_records_run_status_check;

ALTER TABLE workflow_run_records
    ADD CONSTRAINT workflow_run_records_run_status_check
    CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled'));
