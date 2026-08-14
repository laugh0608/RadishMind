ALTER TABLE saved_workflow_drafts
    ADD COLUMN lifecycle_state text NOT NULL DEFAULT 'active',
    ADD COLUMN lifecycle_version bigint NOT NULL DEFAULT 1,
    ADD COLUMN archived_at timestamptz,
    ADD COLUMN library_updated_at timestamptz,
    ADD COLUMN lifecycle_updated_by_actor_ref text NOT NULL DEFAULT '',
    ADD COLUMN draft_name text NOT NULL DEFAULT '',
    ADD COLUMN validation_state text NOT NULL DEFAULT 'invalid_draft',
    ADD COLUMN provenance_kind text NOT NULL DEFAULT 'unversioned';

UPDATE saved_workflow_drafts
   SET library_updated_at = updated_at,
       draft_name = sanitized_draft_payload ->> 'name',
       validation_state = validation_summary ->> 'validation_state',
       provenance_kind = CASE
           WHEN jsonb_typeof(
               sanitized_draft_payload -> 'additional_fields' -> 'derivation_v1'
           ) = 'object'
               THEN 'saved_draft_derivation'
           WHEN COALESCE(
               (sanitized_draft_payload ->> 'base_definition_version')::bigint,
               0
           ) > 0
               THEN 'workflow_definition'
           ELSE 'unversioned'
       END;

ALTER TABLE saved_workflow_drafts
    ALTER COLUMN library_updated_at SET NOT NULL,
    ADD CONSTRAINT saved_workflow_drafts_lifecycle_state_check
        CHECK (lifecycle_state IN ('active', 'archived')),
    ADD CONSTRAINT saved_workflow_drafts_lifecycle_version_check
        CHECK (lifecycle_version > 0),
    ADD CONSTRAINT saved_workflow_drafts_archived_at_check
        CHECK (
            (lifecycle_state = 'active' AND archived_at IS NULL) OR
            (lifecycle_state = 'archived' AND archived_at IS NOT NULL)
        ),
    ADD CONSTRAINT saved_workflow_drafts_validation_state_check
        CHECK (
            validation_state IN (
                'valid_for_review',
                'invalid_draft',
                'blocked_capability',
                'schema_unsupported'
            )
        ),
    ADD CONSTRAINT saved_workflow_drafts_provenance_kind_check
        CHECK (
            provenance_kind IN (
                'unversioned',
                'workflow_definition',
                'saved_draft_derivation'
            )
        );

CREATE TABLE saved_workflow_draft_lifecycle_events (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    draft_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    lifecycle_version bigint NOT NULL CHECK (lifecycle_version > 1),
    from_state text NOT NULL CHECK (from_state IN ('active', 'archived')),
    to_state text NOT NULL CHECK (to_state IN ('active', 'archived')),
    transition_kind text NOT NULL CHECK (transition_kind IN ('archived', 'unarchived')),
    occurred_at timestamptz NOT NULL,
    actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
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
);

DROP INDEX saved_workflow_drafts_owner_list_idx;
DROP INDEX saved_workflow_drafts_status_list_idx;

CREATE INDEX saved_workflow_drafts_owner_lifecycle_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        library_updated_at DESC,
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
        library_updated_at DESC,
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
        library_updated_at DESC,
        draft_id ASC
    );

CREATE INDEX saved_workflow_drafts_name_list_idx
    ON saved_workflow_drafts (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        lifecycle_state,
        draft_name text_pattern_ops,
        library_updated_at DESC,
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

CREATE FUNCTION reject_saved_workflow_draft_lifecycle_event_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'saved workflow draft lifecycle events are append-only'
        USING ERRCODE = '42501';
END;
$$;

CREATE TRIGGER saved_workflow_draft_lifecycle_events_append_only
BEFORE UPDATE OR DELETE ON saved_workflow_draft_lifecycle_events
FOR EACH STATEMENT
EXECUTE FUNCTION reject_saved_workflow_draft_lifecycle_event_mutation();
