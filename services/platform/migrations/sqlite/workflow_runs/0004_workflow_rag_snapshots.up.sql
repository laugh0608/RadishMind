CREATE TABLE workflow_rag_snapshot_resources (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    snapshot_id TEXT NOT NULL CHECK (length(snapshot_id) = 21 AND substr(snapshot_id, 1, 5) = 'rags_' AND substr(snapshot_id, 6) NOT GLOB '*[^a-z2-7]*'),
    snapshot_key TEXT NOT NULL CHECK (length(snapshot_key) BETWEEN 3 AND 48 AND snapshot_key NOT GLOB '*[^a-z0-9_]*'),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    latest_version INTEGER NOT NULL CHECK (latest_version > 0),
    latest_digest TEXT NOT NULL CHECK (length(latest_digest) = 71 AND substr(latest_digest, 1, 7) = 'sha256:' AND substr(latest_digest, 8) NOT GLOB '*[^0-9a-f]*'),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    archived_at_unix_nano INTEGER,
    sanitized_resource_payload TEXT NOT NULL CHECK (json_valid(sanitized_resource_payload) AND json_type(sanitized_resource_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_key),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, latest_version),
    CHECK (
        (lifecycle_state = 'active' AND archived_at_unix_nano IS NULL)
        OR (lifecycle_state = 'archived' AND archived_at_unix_nano IS NOT NULL AND archived_at_unix_nano >= created_at_unix_nano)
    )
) STRICT;

CREATE TABLE workflow_rag_snapshot_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    snapshot_version INTEGER NOT NULL CHECK (snapshot_version > 0),
    snapshot_digest TEXT NOT NULL CHECK (length(snapshot_digest) = 71 AND substr(snapshot_digest, 1, 7) = 'sha256:' AND substr(snapshot_digest, 8) NOT GLOB '*[^0-9a-f]*'),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    sanitized_snapshot_payload TEXT NOT NULL CHECK (json_valid(sanitized_snapshot_payload) AND json_type(sanitized_snapshot_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_digest),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id)
        REFERENCES workflow_rag_snapshot_resources (tenant_ref, workspace_id, application_id, snapshot_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_snapshot_fragments (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    snapshot_version INTEGER NOT NULL CHECK (snapshot_version > 0),
    fragment_ref TEXT NOT NULL CHECK (length(fragment_ref) BETWEEN 3 AND 64 AND fragment_ref NOT GLOB '*[^a-z0-9_]*'),
    content_digest TEXT NOT NULL CHECK (length(content_digest) = 71 AND substr(content_digest, 1, 7) = 'sha256:' AND substr(content_digest, 8) NOT GLOB '*[^0-9a-f]*'),
    sanitized_fragment_payload TEXT NOT NULL CHECK (json_valid(sanitized_fragment_payload) AND json_type(sanitized_fragment_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version, fragment_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        REFERENCES workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_execution_audits (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    event_id TEXT NOT NULL CHECK (length(event_id) = 21 AND substr(event_id, 1, 5) = 'rage_' AND substr(event_id, 6) NOT GLOB '*[^a-z2-7]*'),
    event_kind TEXT NOT NULL CHECK (event_kind IN ('snapshot_created', 'snapshot_versioned', 'snapshot_archived')),
    snapshot_version INTEGER NOT NULL CHECK (snapshot_version > 0),
    snapshot_digest TEXT NOT NULL CHECK (length(snapshot_digest) = 71 AND substr(snapshot_digest, 1, 7) = 'sha256:' AND substr(snapshot_digest, 8) NOT GLOB '*[^0-9a-f]*'),
    actor_ref TEXT NOT NULL CHECK (length(trim(actor_ref)) > 0),
    request_id TEXT NOT NULL CHECK (length(trim(request_id)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_audit_payload TEXT NOT NULL CHECK (json_valid(sanitized_audit_payload) AND json_type(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, snapshot_id, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, snapshot_id, audit_ref),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        REFERENCES workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE INDEX workflow_rag_snapshot_resources_list_idx
    ON workflow_rag_snapshot_resources (tenant_ref, workspace_id, application_id, lifecycle_state, snapshot_key);
CREATE INDEX workflow_rag_snapshot_versions_history_idx
    ON workflow_rag_snapshot_versions (tenant_ref, workspace_id, application_id, snapshot_id, snapshot_version DESC);
CREATE INDEX workflow_rag_execution_audits_history_idx
    ON workflow_rag_execution_audits (tenant_ref, workspace_id, application_id, snapshot_id, occurred_at_unix_nano DESC, event_id);

CREATE TRIGGER workflow_rag_snapshot_versions_append_only_update
BEFORE UPDATE ON workflow_rag_snapshot_versions BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot versions are append-only'); END;
CREATE TRIGGER workflow_rag_snapshot_versions_append_only_delete
BEFORE DELETE ON workflow_rag_snapshot_versions BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot versions are append-only'); END;
CREATE TRIGGER workflow_rag_snapshot_fragments_append_only_update
BEFORE UPDATE ON workflow_rag_snapshot_fragments BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot fragments are append-only'); END;
CREATE TRIGGER workflow_rag_snapshot_fragments_append_only_delete
BEFORE DELETE ON workflow_rag_snapshot_fragments BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot fragments are append-only'); END;
CREATE TRIGGER workflow_rag_execution_audits_append_only_update
BEFORE UPDATE ON workflow_rag_execution_audits BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot audits are append-only'); END;
CREATE TRIGGER workflow_rag_execution_audits_append_only_delete
BEFORE DELETE ON workflow_rag_execution_audits BEGIN SELECT RAISE(ABORT, 'workflow RAG snapshot audits are append-only'); END;
