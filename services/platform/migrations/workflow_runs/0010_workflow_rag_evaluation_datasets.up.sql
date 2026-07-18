CREATE TABLE workflow_rag_evaluation_dataset_resources (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    dataset_id text NOT NULL CHECK (dataset_id ~ '^wragd_[a-z0-9_]{3,47}$'),
    dataset_key text NOT NULL CHECK (dataset_key ~ '^[a-z][a-z0-9_]{2,47}$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    latest_version bigint NOT NULL CHECK (latest_version > 0),
    latest_digest text NOT NULL CHECK (latest_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    archived_at timestamptz,
    sanitized_resource_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_resource_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id),
    UNIQUE (tenant_ref, workspace_id, application_id, dataset_key),
    CHECK (
        (lifecycle_state = 'active' AND archived_at IS NULL)
        OR (lifecycle_state = 'archived' AND archived_at IS NOT NULL AND archived_at >= created_at)
    )
);

CREATE TABLE workflow_rag_evaluation_dataset_versions (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    dataset_id text NOT NULL CHECK (dataset_id ~ '^wragd_[a-z0-9_]{3,47}$'),
    dataset_version bigint NOT NULL CHECK (dataset_version > 0),
    dataset_digest text NOT NULL CHECK (dataset_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    dataset_version_payload jsonb NOT NULL CHECK (jsonb_typeof(dataset_version_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version),
    UNIQUE (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id)
        REFERENCES workflow_rag_evaluation_dataset_resources (tenant_ref, workspace_id, application_id, dataset_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_candidate_snapshot_reviews (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    dataset_id text NOT NULL CHECK (dataset_id ~ '^wragd_[a-z0-9_]{3,47}$'),
    dataset_version bigint NOT NULL CHECK (dataset_version > 0),
    dataset_digest text NOT NULL CHECK (dataset_digest ~ '^sha256:[0-9a-f]{64}$'),
    review_id text NOT NULL CHECK (review_id ~ '^wragr_[a-z2-7]{16}$'),
    created_at timestamptz NOT NULL,
    review_payload jsonb NOT NULL CHECK (jsonb_typeof(review_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, review_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest)
        REFERENCES workflow_rag_evaluation_dataset_versions (tenant_ref, workspace_id, application_id, dataset_id, dataset_version, dataset_digest)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_evaluation_audits (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    dataset_id text NOT NULL CHECK (dataset_id ~ '^wragd_[a-z0-9_]{3,47}$'),
    dataset_version bigint NOT NULL CHECK (dataset_version > 0),
    event_id text NOT NULL CHECK (event_id ~ '^wraga_[a-z2-7]{16}$'),
    event_kind text NOT NULL CHECK (event_kind IN ('dataset_created', 'dataset_versioned', 'dataset_archived', 'candidate_review_created')),
    occurred_at timestamptz NOT NULL,
    audit_payload jsonb NOT NULL CHECK (jsonb_typeof(audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, dataset_id, event_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, dataset_id, dataset_version)
        REFERENCES workflow_rag_evaluation_dataset_versions (tenant_ref, workspace_id, application_id, dataset_id, dataset_version)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX workflow_rag_evaluation_dataset_resources_list_idx
    ON workflow_rag_evaluation_dataset_resources (tenant_ref, workspace_id, application_id, lifecycle_state, dataset_key);
CREATE INDEX workflow_rag_candidate_snapshot_reviews_history_idx
    ON workflow_rag_candidate_snapshot_reviews (tenant_ref, workspace_id, application_id, dataset_id, created_at DESC, review_id DESC);
CREATE INDEX workflow_rag_evaluation_audits_history_idx
    ON workflow_rag_evaluation_audits (tenant_ref, workspace_id, application_id, dataset_id, occurred_at DESC, event_id DESC);

CREATE TRIGGER workflow_rag_evaluation_dataset_versions_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_evaluation_dataset_versions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_candidate_snapshot_reviews_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_candidate_snapshot_reviews
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_evaluation_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_evaluation_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
