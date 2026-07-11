ALTER TABLE workflow_run_records
    ADD COLUMN failure_code text NOT NULL DEFAULT '',
    ADD COLUMN failure_boundary text NOT NULL DEFAULT '',
    ADD COLUMN selected_provider text NOT NULL DEFAULT '',
    ADD COLUMN selected_model text NOT NULL DEFAULT '';

UPDATE workflow_run_records SET
    failure_code = COALESCE(sanitized_run_record ->> 'failure_code', ''),
    failure_boundary = COALESCE(sanitized_run_record -> 'diagnostic' ->> 'failure_boundary', ''),
    selected_provider = COALESCE(sanitized_run_record ->> 'selected_provider', ''),
    selected_model = COALESCE(sanitized_run_record ->> 'selected_model', '');

CREATE INDEX workflow_run_records_failure_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, failure_code, failure_boundary, started_at DESC, run_id DESC);
CREATE INDEX workflow_run_records_provider_model_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id, selected_provider, selected_model, started_at DESC, run_id DESC);
