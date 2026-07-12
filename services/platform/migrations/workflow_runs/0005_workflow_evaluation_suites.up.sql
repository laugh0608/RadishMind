CREATE TABLE workflow_evaluation_suites (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    suite_id text NOT NULL,
    created_at timestamptz NOT NULL,
    current_decision_version integer NOT NULL DEFAULT 0 CHECK (current_decision_version >= 0),
    sanitized_suite_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, suite_id)
);

CREATE TABLE workflow_evaluation_suite_decisions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    suite_id text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    created_at timestamptz NOT NULL,
    sanitized_decision_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, suite_id, version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, suite_id)
        REFERENCES workflow_evaluation_suites (tenant_ref, workspace_id, application_id, suite_id)
        ON DELETE RESTRICT
);

CREATE INDEX workflow_evaluation_suites_history_idx ON workflow_evaluation_suites
    (tenant_ref, workspace_id, application_id, created_at DESC, suite_id DESC);
CREATE INDEX workflow_evaluation_suite_decisions_history_idx ON workflow_evaluation_suite_decisions
    (tenant_ref, workspace_id, application_id, suite_id, version DESC);
