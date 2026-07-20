CREATE TABLE application_catalog_records (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL,
    application_kind TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    sanitized_record_payload TEXT NOT NULL CHECK (json_valid(sanitized_record_payload)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    archived_at TEXT,
    created_by_actor_ref TEXT NOT NULL,
    updated_by_actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id)
) STRICT;

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
