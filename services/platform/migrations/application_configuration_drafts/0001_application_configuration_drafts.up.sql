CREATE TABLE application_configuration_drafts (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    draft_id text NOT NULL,
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    schema_version text NOT NULL,
    sanitized_draft_payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    created_by_actor_ref text NOT NULL,
    updated_by_actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id)
);

CREATE INDEX application_configuration_drafts_scope_list_idx
    ON application_configuration_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        updated_at DESC,
        draft_id ASC
    );
