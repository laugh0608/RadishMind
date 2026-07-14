CREATE TABLE application_configuration_drafts (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    schema_version TEXT NOT NULL,
    sanitized_draft_payload TEXT NOT NULL CHECK (json_valid(sanitized_draft_payload)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    created_by_actor_ref TEXT NOT NULL,
    updated_by_actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id)
) STRICT;

CREATE INDEX application_configuration_drafts_scope_list_idx
    ON application_configuration_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        updated_at DESC,
        draft_id ASC
    );
