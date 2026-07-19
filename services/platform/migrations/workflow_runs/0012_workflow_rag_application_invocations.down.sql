DROP TRIGGER IF EXISTS workflow_rag_application_runtime_audits_append_only
    ON workflow_rag_application_runtime_audits;
DROP TRIGGER IF EXISTS workflow_rag_application_runtime_events_append_only
    ON workflow_rag_application_runtime_events;
DROP TABLE IF EXISTS workflow_rag_application_runtime_audits;
DROP TABLE IF EXISTS workflow_rag_application_runtime_events;
DROP TABLE IF EXISTS workflow_rag_application_runtime_assignments;

DROP INDEX workflow_run_records_source_idx;
ALTER TABLE workflow_run_records
    ADD COLUMN draft_id text,
    ADD COLUMN draft_version bigint;

UPDATE workflow_run_records
SET draft_id=execution_source_id,
    draft_version=execution_source_version
WHERE execution_source_kind='workflow_draft';

ALTER TABLE workflow_run_records
    ALTER COLUMN draft_id SET NOT NULL,
    ALTER COLUMN draft_version SET NOT NULL,
    DROP CONSTRAINT workflow_run_records_execution_source_check,
    DROP COLUMN execution_source_kind,
    DROP COLUMN execution_source_id,
    DROP COLUMN execution_source_version;

CREATE INDEX workflow_run_records_draft_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, draft_id, started_at DESC, run_id DESC);
