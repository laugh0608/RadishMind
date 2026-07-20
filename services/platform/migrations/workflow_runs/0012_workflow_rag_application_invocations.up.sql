ALTER TABLE workflow_run_records
    ADD COLUMN execution_source_kind text,
    ADD COLUMN execution_source_id text,
    ADD COLUMN execution_source_version bigint;

UPDATE workflow_run_records
SET execution_source_kind='workflow_draft',
    execution_source_id=draft_id,
    execution_source_version=draft_version;

ALTER TABLE workflow_run_records
    ALTER COLUMN execution_source_kind SET NOT NULL,
    ALTER COLUMN execution_source_id SET NOT NULL,
    ALTER COLUMN execution_source_version SET NOT NULL,
    ADD CONSTRAINT workflow_run_records_execution_source_check CHECK (
        execution_source_kind IN ('workflow_draft', 'application_configuration_draft')
        AND btrim(execution_source_id) <> ''
        AND execution_source_version > 0
        AND (
            (schema_version IN ('workflow_run_record.v0', 'workflow_run_record.v1', 'workflow_run_record.v2', 'workflow_run_record.v3')
                AND execution_source_kind = 'workflow_draft')
            OR (schema_version = 'workflow_run_record.v4'
                AND execution_source_kind = 'application_configuration_draft')
        )
    );

DROP INDEX workflow_run_records_draft_idx;
ALTER TABLE workflow_run_records DROP COLUMN draft_id, DROP COLUMN draft_version;

CREATE INDEX workflow_run_records_source_idx ON workflow_run_records
    (tenant_ref, workspace_id, application_id,
     execution_source_kind, execution_source_id, started_at DESC, run_id DESC);

CREATE TABLE workflow_rag_application_runtime_assignments (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    assignment_id text NOT NULL UNIQUE CHECK (assignment_id ~ '^wragra_[a-z2-7]{16}$'),
    record_version bigint NOT NULL CHECK (record_version > 0),
    assignment_state text NOT NULL CHECK (assignment_state IN ('active', 'revoked')),
    assignment_digest text NOT NULL CHECK (assignment_digest ~ '^sha256:[0-9a-f]{64}$'),
    updated_at timestamptz NOT NULL,
    sanitized_assignment_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_assignment_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, assignment_id)
);

CREATE TABLE workflow_rag_application_runtime_events (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    assignment_id text NOT NULL,
    event_id text NOT NULL UNIQUE CHECK (event_id ~ '^wragre_[a-z2-7]{16}$'),
    after_record_version bigint NOT NULL CHECK (after_record_version > 0),
    occurred_at timestamptz NOT NULL,
    sanitized_event_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_event_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, after_record_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, assignment_id)
        REFERENCES workflow_rag_application_runtime_assignments (
            tenant_ref, workspace_id, application_id, owner_subject_ref, assignment_id
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_application_runtime_audits (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    assignment_id text NOT NULL,
    audit_event_id text NOT NULL UNIQUE CHECK (audit_event_id ~ '^wragru_[a-z2-7]{16}$'),
    record_version bigint NOT NULL CHECK (record_version > 0),
    occurred_at timestamptz NOT NULL,
    sanitized_audit_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, audit_event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, record_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, assignment_id)
        REFERENCES workflow_rag_application_runtime_assignments (
            tenant_ref, workspace_id, application_id, owner_subject_ref, assignment_id
        ) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX workflow_rag_application_runtime_events_history_idx
    ON workflow_rag_application_runtime_events
        (tenant_ref, workspace_id, application_id, owner_subject_ref, after_record_version);
CREATE INDEX workflow_rag_application_runtime_audits_history_idx
    ON workflow_rag_application_runtime_audits
        (tenant_ref, workspace_id, application_id, owner_subject_ref, record_version);

CREATE TRIGGER workflow_rag_application_runtime_events_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_application_runtime_events
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_application_runtime_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_application_runtime_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
