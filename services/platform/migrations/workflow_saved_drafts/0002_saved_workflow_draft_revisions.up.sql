CREATE TABLE saved_workflow_draft_revisions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    draft_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    revision_kind text NOT NULL CHECK (
        revision_kind IN ('saved', 'restored', 'backfilled_current')
    ),
    restored_from_version bigint NOT NULL DEFAULT 0 CHECK (restored_from_version >= 0),
    sanitized_revision_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id, draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, draft_id)
        REFERENCES saved_workflow_drafts (tenant_ref, workspace_id, application_id, draft_id)
        ON DELETE RESTRICT,
    CHECK (
        (revision_kind = 'restored' AND restored_from_version > 0) OR
        (revision_kind <> 'restored' AND restored_from_version = 0)
    )
);

CREATE INDEX saved_workflow_draft_revisions_owner_version_idx
    ON saved_workflow_draft_revisions (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        draft_id,
        draft_version DESC
    );

INSERT INTO saved_workflow_draft_revisions (
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    draft_version,
    revision_kind,
    restored_from_version,
    sanitized_revision_record
)
SELECT
    tenant_ref,
    workspace_id,
    application_id,
    draft_id,
    owner_subject_ref,
    draft_version,
    'backfilled_current',
    0,
    sanitized_draft_payload
FROM saved_workflow_drafts;
