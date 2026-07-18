CREATE TABLE workflow_rag_evaluation_dataset_resources (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    dataset_id TEXT NOT NULL CHECK (dataset_id GLOB 'wragd_*'),
    dataset_key TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    latest_version INTEGER NOT NULL CHECK (latest_version > 0),
    latest_digest TEXT NOT NULL CHECK (length(latest_digest) = 71 AND substr(latest_digest, 1, 7) = 'sha256:'),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    archived_at_unix_nano INTEGER,
    sanitized_resource_payload TEXT NOT NULL CHECK (json_valid(sanitized_resource_payload) AND json_type(sanitized_resource_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id),
    UNIQUE (tenant_ref, workspace_id, application_id, dataset_key),
    UNIQUE (tenant_ref, workspace_id, application_id, dataset_id, latest_version),
    CHECK ((lifecycle_state = 'active' AND archived_at_unix_nano IS NULL) OR
           (lifecycle_state = 'archived' AND archived_at_unix_nano IS NOT NULL AND archived_at_unix_nano >= created_at_unix_nano))
) STRICT;

CREATE TABLE workflow_rag_evaluation_dataset_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    dataset_id TEXT NOT NULL,
    dataset_version INTEGER NOT NULL CHECK (dataset_version > 0),
    dataset_digest TEXT NOT NULL CHECK (length(dataset_digest) = 71 AND substr(dataset_digest, 1, 7) = 'sha256:'),
    created_at_unix_nano INTEGER NOT NULL,
    dataset_version_payload TEXT NOT NULL CHECK (json_valid(dataset_version_payload) AND json_type(dataset_version_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version),
    UNIQUE (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id)
        REFERENCES workflow_rag_evaluation_dataset_resources (tenant_ref, workspace_id, application_id, dataset_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_candidate_snapshot_reviews (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    dataset_id TEXT NOT NULL,
    dataset_version INTEGER NOT NULL CHECK (dataset_version > 0),
    dataset_digest TEXT NOT NULL,
    review_id TEXT NOT NULL CHECK (review_id GLOB 'wragr_*'),
    created_at_unix_nano INTEGER NOT NULL,
    review_payload TEXT NOT NULL CHECK (json_valid(review_payload) AND json_type(review_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, review_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest)
        REFERENCES workflow_rag_evaluation_dataset_versions (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_evaluation_audits (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    dataset_id TEXT NOT NULL,
    dataset_version INTEGER NOT NULL CHECK (dataset_version > 0),
    event_id TEXT NOT NULL CHECK (event_id GLOB 'wraga_*'),
    event_kind TEXT NOT NULL CHECK (event_kind IN ('dataset_created', 'dataset_versioned', 'dataset_archived', 'candidate_review_created')),
    occurred_at_unix_nano INTEGER NOT NULL,
    audit_payload TEXT NOT NULL CHECK (json_valid(audit_payload) AND json_type(audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, event_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version)
        REFERENCES workflow_rag_evaluation_dataset_versions (tenant_ref, workspace_id, application_id, dataset_id, dataset_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE INDEX workflow_rag_evaluation_dataset_resources_list_idx
    ON workflow_rag_evaluation_dataset_resources (tenant_ref, workspace_id, application_id, lifecycle_state, dataset_key);
CREATE INDEX workflow_rag_candidate_snapshot_reviews_history_idx
    ON workflow_rag_candidate_snapshot_reviews (tenant_ref, workspace_id, application_id, dataset_id, created_at_unix_nano DESC, review_id DESC);
CREATE INDEX workflow_rag_evaluation_audits_history_idx
    ON workflow_rag_evaluation_audits (tenant_ref, workspace_id, application_id, dataset_id, occurred_at_unix_nano DESC, event_id DESC);

CREATE TRIGGER workflow_rag_evaluation_dataset_versions_append_only_update
BEFORE UPDATE ON workflow_rag_evaluation_dataset_versions BEGIN
    SELECT RAISE(ABORT, 'workflow RAG evaluation dataset versions are append-only');
END;
CREATE TRIGGER workflow_rag_evaluation_dataset_versions_append_only_delete
BEFORE DELETE ON workflow_rag_evaluation_dataset_versions BEGIN
    SELECT RAISE(ABORT, 'workflow RAG evaluation dataset versions are append-only');
END;
CREATE TRIGGER workflow_rag_candidate_snapshot_reviews_append_only_update
BEFORE UPDATE ON workflow_rag_candidate_snapshot_reviews BEGIN
    SELECT RAISE(ABORT, 'workflow RAG candidate reviews are append-only');
END;
CREATE TRIGGER workflow_rag_candidate_snapshot_reviews_append_only_delete
BEFORE DELETE ON workflow_rag_candidate_snapshot_reviews BEGIN
    SELECT RAISE(ABORT, 'workflow RAG candidate reviews are append-only');
END;
CREATE TRIGGER workflow_rag_evaluation_audits_append_only_update
BEFORE UPDATE ON workflow_rag_evaluation_audits BEGIN
    SELECT RAISE(ABORT, 'workflow RAG evaluation audits are append-only');
END;
CREATE TRIGGER workflow_rag_evaluation_audits_append_only_delete
BEFORE DELETE ON workflow_rag_evaluation_audits BEGIN
    SELECT RAISE(ABORT, 'workflow RAG evaluation audits are append-only');
END;
