CREATE TABLE application_catalog_records (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    schema_version text NOT NULL,
    display_name text NOT NULL,
    description text NOT NULL,
    application_kind text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    sanitized_record_payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    archived_at timestamptz,
    created_by_actor_ref text NOT NULL,
    updated_by_actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id)
);

CREATE INDEX application_catalog_records_owner_list_idx
    ON application_catalog_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        lifecycle_state,
        updated_at DESC,
        application_id DESC
    );

CREATE INDEX application_catalog_records_owner_kind_list_idx
    ON application_catalog_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        lifecycle_state,
        application_kind,
        updated_at DESC,
        application_id DESC
    );
