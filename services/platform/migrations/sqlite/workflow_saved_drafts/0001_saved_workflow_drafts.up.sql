CREATE TABLE saved_workflow_drafts (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    store_schema_version TEXT NOT NULL CHECK (store_schema_version = 'saved_workflow_drafts_store_v1'),
    schema_version TEXT NOT NULL CHECK (schema_version = 'saved_workflow_draft.v1'),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    draft_status TEXT NOT NULL CHECK (
        draft_status IN ('valid_for_review', 'invalid_draft', 'blocked_capability', 'schema_unsupported')
    ),
    sanitized_draft_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_draft_payload) AND json_type(sanitized_draft_payload) = 'object'
    ),
    validation_summary TEXT NOT NULL CHECK (
        json_valid(validation_summary) AND json_type(validation_summary) = 'object'
    ),
    blocked_capability_summary TEXT NOT NULL CHECK (
        json_valid(blocked_capability_summary) AND json_type(blocked_capability_summary) = 'array'
    ),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    created_by_actor_ref TEXT NOT NULL,
    updated_by_actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id)
) STRICT;

CREATE INDEX saved_workflow_drafts_owner_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_status_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        draft_status,
        updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_schema_version_idx
    ON saved_workflow_drafts (
        tenant_ref,
        store_schema_version,
        schema_version
    );
