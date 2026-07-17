CREATE TABLE workflow_rag_snapshot_resources (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    snapshot_id text NOT NULL CHECK (snapshot_id ~ '^rags_[a-z2-7]{16}$'),
    snapshot_key text NOT NULL CHECK (snapshot_key ~ '^[a-z][a-z0-9_]{2,47}$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    latest_version bigint NOT NULL CHECK (latest_version > 0),
    latest_digest text NOT NULL CHECK (latest_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    archived_at timestamptz,
    sanitized_resource_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_resource_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_key),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, latest_version),
    CHECK (
        (lifecycle_state = 'active' AND archived_at IS NULL)
        OR (lifecycle_state = 'archived' AND archived_at IS NOT NULL AND archived_at >= created_at)
    )
);

CREATE TABLE workflow_rag_snapshot_versions (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    snapshot_id text NOT NULL CHECK (snapshot_id ~ '^rags_[a-z2-7]{16}$'),
    snapshot_version bigint NOT NULL CHECK (snapshot_version > 0),
    snapshot_digest text NOT NULL CHECK (snapshot_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    sanitized_snapshot_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_snapshot_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_digest),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id)
        REFERENCES workflow_rag_snapshot_resources (tenant_ref, workspace_id, application_id, snapshot_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_snapshot_fragments (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    snapshot_id text NOT NULL CHECK (snapshot_id ~ '^rags_[a-z2-7]{16}$'),
    snapshot_version bigint NOT NULL CHECK (snapshot_version > 0),
    fragment_ref text NOT NULL CHECK (fragment_ref ~ '^[a-z][a-z0-9_]{2,63}$'),
    content_digest text NOT NULL CHECK (content_digest ~ '^sha256:[0-9a-f]{64}$'),
    sanitized_fragment_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_fragment_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version, fragment_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        REFERENCES workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_execution_audits (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    snapshot_id text NOT NULL CHECK (snapshot_id ~ '^rags_[a-z2-7]{16}$'),
    event_id text NOT NULL CHECK (event_id ~ '^rage_[a-z2-7]{16}$'),
    event_kind text NOT NULL CHECK (event_kind IN ('snapshot_created', 'snapshot_versioned', 'snapshot_archived')),
    snapshot_version bigint NOT NULL CHECK (snapshot_version > 0),
    snapshot_digest text NOT NULL CHECK (snapshot_digest ~ '^sha256:[0-9a-f]{64}$'),
    actor_ref text NOT NULL CHECK (btrim(actor_ref) <> ''),
    request_id text NOT NULL CHECK (btrim(request_id) <> ''),
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    occurred_at timestamptz NOT NULL,
    sanitized_audit_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, audit_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        REFERENCES workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX workflow_rag_snapshot_resources_list_idx
    ON workflow_rag_snapshot_resources (tenant_ref, workspace_id, application_id, lifecycle_state, snapshot_key);
CREATE INDEX workflow_rag_snapshot_versions_history_idx
    ON workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version DESC);
CREATE INDEX workflow_rag_execution_audits_history_idx
    ON workflow_rag_execution_audits (tenant_ref, workspace_id, application_id, snapshot_id, occurred_at DESC, event_id);

CREATE FUNCTION reject_workflow_rag_snapshot_append_only_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'workflow RAG snapshot versions, fragments and audits are append-only';
END;
$$;

CREATE TRIGGER workflow_rag_snapshot_versions_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_snapshot_versions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_snapshot_fragments_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_snapshot_fragments
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
