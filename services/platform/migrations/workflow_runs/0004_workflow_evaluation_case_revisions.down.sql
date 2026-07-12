DROP TABLE IF EXISTS workflow_evaluation_case_revisions;
ALTER TABLE workflow_evaluation_cases DROP COLUMN IF EXISTS current_version;
