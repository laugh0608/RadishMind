DROP INDEX IF EXISTS workflow_run_records_provider_model_idx;
DROP INDEX IF EXISTS workflow_run_records_failure_idx;
ALTER TABLE workflow_run_records
    DROP COLUMN IF EXISTS selected_model,
    DROP COLUMN IF EXISTS selected_provider,
    DROP COLUMN IF EXISTS failure_boundary,
    DROP COLUMN IF EXISTS failure_code;
