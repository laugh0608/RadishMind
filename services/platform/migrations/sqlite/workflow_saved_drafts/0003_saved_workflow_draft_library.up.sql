ALTER TABLE saved_workflow_drafts
    ADD COLUMN lifecycle_state TEXT NOT NULL DEFAULT 'active'
    CHECK (lifecycle_state IN ('active', 'archived'));
ALTER TABLE saved_workflow_drafts
    ADD COLUMN lifecycle_version INTEGER NOT NULL DEFAULT 1
    CHECK (lifecycle_version > 0);
ALTER TABLE saved_workflow_drafts
    ADD COLUMN archived_at_unix_nano INTEGER;
ALTER TABLE saved_workflow_drafts
    ADD COLUMN library_updated_at_unix_nano INTEGER;
ALTER TABLE saved_workflow_drafts
    ADD COLUMN lifecycle_updated_by_actor_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE saved_workflow_drafts
    ADD COLUMN draft_name TEXT NOT NULL DEFAULT '';
ALTER TABLE saved_workflow_drafts
    ADD COLUMN validation_state TEXT NOT NULL DEFAULT 'invalid_draft'
    CHECK (
        validation_state IN (
            'valid_for_review',
            'invalid_draft',
            'blocked_capability',
            'schema_unsupported'
        )
    );
ALTER TABLE saved_workflow_drafts
    ADD COLUMN provenance_kind TEXT NOT NULL DEFAULT 'unversioned'
    CHECK (
        provenance_kind IN (
            'unversioned',
            'workflow_definition',
            'saved_draft_derivation'
        )
    );

UPDATE saved_workflow_drafts
   SET library_updated_at_unix_nano = updated_at_unix_nano,
       draft_name = json_extract(sanitized_draft_payload, '$.name'),
       validation_state = json_extract(validation_summary, '$.validation_state'),
       provenance_kind = CASE
           WHEN json_type(
               sanitized_draft_payload,
               '$.additional_fields.derivation_v1'
           ) = 'object'
               THEN 'saved_draft_derivation'
           WHEN COALESCE(
               json_extract(
                   sanitized_draft_payload,
                   '$.base_definition_version'
               ),
               0
           ) > 0
               THEN 'workflow_definition'
           ELSE 'unversioned'
       END;

CREATE TABLE saved_workflow_draft_lifecycle_events (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    lifecycle_version INTEGER NOT NULL CHECK (lifecycle_version > 1),
    from_state TEXT NOT NULL CHECK (from_state IN ('active', 'archived')),
    to_state TEXT NOT NULL CHECK (to_state IN ('active', 'archived')),
    transition_kind TEXT NOT NULL CHECK (transition_kind IN ('archived', 'unarchived')),
    occurred_at_unix_nano INTEGER NOT NULL,
    actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (
        tenant_ref,
        workspace_id,
        application_id,
        draft_id,
        lifecycle_version
    ),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, draft_id)
        REFERENCES saved_workflow_drafts (tenant_ref, workspace_id, application_id, draft_id)
        ON DELETE RESTRICT,
    CHECK (from_state <> to_state),
    CHECK (
        (transition_kind = 'archived' AND from_state = 'active' AND to_state = 'archived') OR
        (transition_kind = 'unarchived' AND from_state = 'archived' AND to_state = 'active')
    )
) STRICT;

DROP INDEX saved_workflow_drafts_owner_list_idx;
DROP INDEX saved_workflow_drafts_status_list_idx;

CREATE INDEX saved_workflow_drafts_owner_lifecycle_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        library_updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_validation_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        validation_state,
        library_updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_provenance_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        provenance_kind,
        library_updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_name_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        draft_name,
        library_updated_at_unix_nano DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_draft_lifecycle_events_owner_version_idx
    ON saved_workflow_draft_lifecycle_events (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        draft_id,
        lifecycle_version DESC
    );
