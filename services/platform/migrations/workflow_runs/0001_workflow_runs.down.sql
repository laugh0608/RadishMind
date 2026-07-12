DROP TABLE IF EXISTS workflow_run_records;
DELETE FROM workflow_run_schema_versions WHERE component = 'workflow_runs';
DROP TABLE IF EXISTS workflow_run_schema_versions;
