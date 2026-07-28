DROP TABLE IF EXISTS saved_workflow_draft_lifecycle_events;
DROP FUNCTION IF EXISTS reject_saved_workflow_draft_lifecycle_event_mutation();

DROP INDEX IF EXISTS saved_workflow_drafts_name_list_idx;
DROP INDEX IF EXISTS saved_workflow_drafts_provenance_list_idx;
DROP INDEX IF EXISTS saved_workflow_drafts_validation_list_idx;
DROP INDEX IF EXISTS saved_workflow_drafts_owner_lifecycle_list_idx;

ALTER TABLE saved_workflow_drafts
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_provenance_kind_check,
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_validation_state_check,
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_archived_at_check,
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_lifecycle_version_check,
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_lifecycle_state_check,
    DROP COLUMN IF EXISTS provenance_kind,
    DROP COLUMN IF EXISTS validation_state,
    DROP COLUMN IF EXISTS draft_name,
    DROP COLUMN IF EXISTS lifecycle_updated_by_actor_ref,
    DROP COLUMN IF EXISTS library_updated_at,
    DROP COLUMN IF EXISTS archived_at,
    DROP COLUMN IF EXISTS lifecycle_version,
    DROP COLUMN IF EXISTS lifecycle_state;

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
