CREATE TABLE workflow_evaluation_cases (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    case_id TEXT NOT NULL,
    baseline_run_id TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    current_version INTEGER NOT NULL CHECK (current_version > 0),
    sanitized_case_record TEXT NOT NULL CHECK (json_valid(sanitized_case_record) AND json_type(sanitized_case_record) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, case_id)
) STRICT;

CREATE INDEX workflow_evaluation_cases_history_idx ON workflow_evaluation_cases
    (tenant_ref, workspace_id, application_id, created_at_unix_nano DESC, case_id DESC);
CREATE INDEX workflow_evaluation_cases_baseline_idx ON workflow_evaluation_cases
    (tenant_ref, workspace_id, application_id, baseline_run_id, created_at_unix_nano DESC, case_id DESC);

CREATE TABLE workflow_evaluation_case_revisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    case_id TEXT NOT NULL,
    version INTEGER NOT NULL CHECK (version > 0),
    revised_at_unix_nano INTEGER NOT NULL,
    sanitized_revision_record TEXT NOT NULL CHECK (json_valid(sanitized_revision_record) AND json_type(sanitized_revision_record) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, case_id, version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, case_id)
        REFERENCES workflow_evaluation_cases (tenant_ref, workspace_id, application_id, case_id)
        ON DELETE RESTRICT
) STRICT;

CREATE INDEX workflow_evaluation_case_revisions_history_idx ON workflow_evaluation_case_revisions
    (tenant_ref, workspace_id, application_id, case_id, version DESC);

CREATE TABLE workflow_evaluation_suites (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    suite_id TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    current_decision_version INTEGER NOT NULL CHECK (current_decision_version >= 0),
    sanitized_suite_record TEXT NOT NULL CHECK (json_valid(sanitized_suite_record) AND json_type(sanitized_suite_record) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, suite_id)
) STRICT;

CREATE INDEX workflow_evaluation_suites_history_idx ON workflow_evaluation_suites
    (tenant_ref, workspace_id, application_id, created_at_unix_nano DESC, suite_id DESC);

CREATE TABLE workflow_evaluation_suite_decisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    suite_id TEXT NOT NULL,
    version INTEGER NOT NULL CHECK (version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    sanitized_decision_record TEXT NOT NULL CHECK (json_valid(sanitized_decision_record) AND json_type(sanitized_decision_record) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, suite_id, version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, suite_id)
        REFERENCES workflow_evaluation_suites (tenant_ref, workspace_id, application_id, suite_id)
        ON DELETE RESTRICT
) STRICT;

CREATE INDEX workflow_evaluation_suite_decisions_history_idx ON workflow_evaluation_suite_decisions
    (tenant_ref, workspace_id, application_id, suite_id, version DESC);

CREATE TRIGGER workflow_evaluation_case_revisions_append_only_update
BEFORE UPDATE ON workflow_evaluation_case_revisions
BEGIN
    SELECT RAISE(ABORT, 'workflow evaluation case revisions are append-only');
END;

CREATE TRIGGER workflow_evaluation_case_revisions_append_only_delete
BEFORE DELETE ON workflow_evaluation_case_revisions
BEGIN
    SELECT RAISE(ABORT, 'workflow evaluation case revisions are append-only');
END;

CREATE TRIGGER workflow_evaluation_suite_decisions_append_only_update
BEFORE UPDATE ON workflow_evaluation_suite_decisions
BEGIN
    SELECT RAISE(ABORT, 'workflow evaluation suite decisions are append-only');
END;

CREATE TRIGGER workflow_evaluation_suite_decisions_append_only_delete
BEFORE DELETE ON workflow_evaluation_suite_decisions
BEGIN
    SELECT RAISE(ABORT, 'workflow evaluation suite decisions are append-only');
END;
