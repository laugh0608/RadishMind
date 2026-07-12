ALTER TABLE workflow_evaluation_cases
    ADD COLUMN current_version integer NOT NULL DEFAULT 1 CHECK (current_version > 0);

CREATE TABLE workflow_evaluation_case_revisions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    case_id text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    created_at timestamptz NOT NULL,
    sanitized_revision_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, case_id, version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, case_id)
        REFERENCES workflow_evaluation_cases (tenant_ref, workspace_id, application_id, case_id)
        ON DELETE RESTRICT
);

INSERT INTO workflow_evaluation_case_revisions
    (tenant_ref, workspace_id, application_id, case_id, version, created_at, sanitized_revision_record)
SELECT tenant_ref, workspace_id, application_id, case_id, 1, created_at, sanitized_case_record
FROM workflow_evaluation_cases;

CREATE INDEX workflow_evaluation_case_revisions_history_idx
    ON workflow_evaluation_case_revisions
    (tenant_ref, workspace_id, application_id, case_id, version DESC);
