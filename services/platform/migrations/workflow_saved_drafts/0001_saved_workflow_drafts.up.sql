CREATE TABLE saved_workflow_drafts (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    draft_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    store_schema_version text NOT NULL,
    schema_version text NOT NULL,
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    draft_status text NOT NULL,
    sanitized_draft_payload jsonb NOT NULL,
    validation_summary jsonb NOT NULL,
    blocked_capability_summary jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    created_by_actor_ref text NOT NULL,
    updated_by_actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id)
);

CREATE INDEX saved_workflow_drafts_owner_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        updated_at DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_status_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        draft_status,
        updated_at DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_schema_version_idx
    ON saved_workflow_drafts (
        tenant_ref,
        store_schema_version,
        schema_version
    );
