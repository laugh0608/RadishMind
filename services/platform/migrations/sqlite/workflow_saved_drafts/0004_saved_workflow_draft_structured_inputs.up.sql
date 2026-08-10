PRAGMA defer_foreign_keys = ON;

DROP INDEX saved_workflow_draft_revisions_owner_version_idx;
DROP INDEX saved_workflow_draft_lifecycle_events_owner_version_idx;
DROP INDEX saved_workflow_drafts_owner_lifecycle_list_idx;
DROP INDEX saved_workflow_drafts_validation_list_idx;
DROP INDEX saved_workflow_drafts_provenance_list_idx;
DROP INDEX saved_workflow_drafts_name_list_idx;
DROP INDEX saved_workflow_drafts_schema_version_idx;

ALTER TABLE saved_workflow_draft_revisions RENAME TO saved_workflow_draft_revisions_pre_structured_inputs;
ALTER TABLE saved_workflow_draft_lifecycle_events RENAME TO saved_workflow_draft_lifecycle_events_pre_structured_inputs;
ALTER TABLE saved_workflow_drafts RENAME TO saved_workflow_drafts_pre_structured_inputs;

CREATE TABLE saved_workflow_drafts (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    store_schema_version TEXT NOT NULL CHECK (store_schema_version = 'saved_workflow_drafts_store_v1'),
    schema_version TEXT NOT NULL CHECK (schema_version IN ('saved_workflow_draft.v1', 'saved_workflow_draft.v2')),
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    draft_status TEXT NOT NULL CHECK (draft_status IN ('valid_for_review', 'invalid_draft', 'blocked_capability', 'schema_unsupported')),
    sanitized_draft_payload TEXT NOT NULL CHECK (json_valid(sanitized_draft_payload) AND json_type(sanitized_draft_payload) = 'object'),
    validation_summary TEXT NOT NULL CHECK (json_valid(validation_summary) AND json_type(validation_summary) = 'object'),
    blocked_capability_summary TEXT NOT NULL CHECK (json_valid(blocked_capability_summary) AND json_type(blocked_capability_summary) = 'array'),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    created_by_actor_ref TEXT NOT NULL,
    updated_by_actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL DEFAULT 'active' CHECK (lifecycle_state IN ('active', 'archived')),
    lifecycle_version INTEGER NOT NULL DEFAULT 1 CHECK (lifecycle_version > 0),
    archived_at_unix_nano INTEGER,
    library_updated_at_unix_nano INTEGER NOT NULL,
    lifecycle_updated_by_actor_ref TEXT NOT NULL DEFAULT '',
    draft_name TEXT NOT NULL DEFAULT '',
    validation_state TEXT NOT NULL DEFAULT 'invalid_draft' CHECK (validation_state IN ('valid_for_review', 'invalid_draft', 'blocked_capability', 'schema_unsupported')),
    provenance_kind TEXT NOT NULL DEFAULT 'unversioned' CHECK (provenance_kind IN ('unversioned', 'workflow_definition', 'saved_draft_derivation')),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id),
    CHECK (json_extract(sanitized_draft_payload, '$.schema_version') = schema_version),
    CHECK (
        schema_version = 'saved_workflow_draft.v1'
        OR (
            schema_version = 'saved_workflow_draft.v2'
            AND length(json_extract(sanitized_draft_payload, '$.input_contract.contract_digest')) = 71
            AND substr(json_extract(sanitized_draft_payload, '$.input_contract.contract_digest'), 1, 7) = 'sha256:'
            AND substr(json_extract(sanitized_draft_payload, '$.input_contract.contract_digest'), 8) NOT GLOB '*[^0-9a-f]*'
        )
    )
) STRICT;

INSERT INTO saved_workflow_drafts (
    tenant_ref,workspace_id,application_id,draft_id,owner_subject_ref,
    store_schema_version,schema_version,draft_version,draft_status,
    sanitized_draft_payload,validation_summary,blocked_capability_summary,
    created_at_unix_nano,updated_at_unix_nano,created_by_actor_ref,updated_by_actor_ref,request_id,audit_ref,
    lifecycle_state,lifecycle_version,archived_at_unix_nano,library_updated_at_unix_nano,
    lifecycle_updated_by_actor_ref,draft_name,validation_state,provenance_kind
)
SELECT tenant_ref,workspace_id,application_id,draft_id,owner_subject_ref,
    store_schema_version,schema_version,draft_version,draft_status,
    sanitized_draft_payload,validation_summary,blocked_capability_summary,
    created_at_unix_nano,updated_at_unix_nano,created_by_actor_ref,updated_by_actor_ref,request_id,audit_ref,
    lifecycle_state,lifecycle_version,archived_at_unix_nano,library_updated_at_unix_nano,
    lifecycle_updated_by_actor_ref,draft_name,validation_state,provenance_kind
FROM saved_workflow_drafts_pre_structured_inputs;

CREATE INDEX saved_workflow_drafts_schema_version_idx ON saved_workflow_drafts (tenant_ref, store_schema_version, schema_version);
CREATE INDEX saved_workflow_drafts_owner_lifecycle_list_idx ON saved_workflow_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, lifecycle_state, library_updated_at_unix_nano DESC, draft_id ASC);
CREATE INDEX saved_workflow_drafts_validation_list_idx ON saved_workflow_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, lifecycle_state, validation_state, library_updated_at_unix_nano DESC, draft_id ASC);
CREATE INDEX saved_workflow_drafts_provenance_list_idx ON saved_workflow_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, lifecycle_state, provenance_kind, library_updated_at_unix_nano DESC, draft_id ASC);
CREATE INDEX saved_workflow_drafts_name_list_idx ON saved_workflow_drafts (tenant_ref, workspace_id, application_id, owner_subject_ref, lifecycle_state, draft_name, library_updated_at_unix_nano DESC, draft_id ASC);

CREATE TABLE saved_workflow_draft_revisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    revision_kind TEXT NOT NULL CHECK (revision_kind IN ('saved', 'restored', 'backfilled_current')),
    restored_from_version INTEGER NOT NULL DEFAULT 0 CHECK (restored_from_version >= 0),
    sanitized_revision_record TEXT NOT NULL CHECK (json_valid(sanitized_revision_record) AND json_type(sanitized_revision_record) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id, draft_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, draft_id) REFERENCES saved_workflow_drafts (tenant_ref, workspace_id, application_id, draft_id) ON DELETE RESTRICT,
    CHECK ((revision_kind = 'restored' AND restored_from_version > 0) OR (revision_kind <> 'restored' AND restored_from_version = 0))
) STRICT;

INSERT INTO saved_workflow_draft_revisions SELECT * FROM saved_workflow_draft_revisions_pre_structured_inputs;
CREATE INDEX saved_workflow_draft_revisions_owner_version_idx ON saved_workflow_draft_revisions (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id, draft_version DESC);

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
    PRIMARY KEY (tenant_ref, workspace_id, application_id, draft_id, lifecycle_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, draft_id) REFERENCES saved_workflow_drafts (tenant_ref, workspace_id, application_id, draft_id) ON DELETE RESTRICT,
    CHECK (from_state <> to_state),
    CHECK ((transition_kind = 'archived' AND from_state = 'active' AND to_state = 'archived') OR (transition_kind = 'unarchived' AND from_state = 'archived' AND to_state = 'active'))
) STRICT;

INSERT INTO saved_workflow_draft_lifecycle_events SELECT * FROM saved_workflow_draft_lifecycle_events_pre_structured_inputs;
CREATE INDEX saved_workflow_draft_lifecycle_events_owner_version_idx ON saved_workflow_draft_lifecycle_events (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id, lifecycle_version DESC);

DROP TABLE saved_workflow_draft_revisions_pre_structured_inputs;
DROP TABLE saved_workflow_draft_lifecycle_events_pre_structured_inputs;
DROP TABLE saved_workflow_drafts_pre_structured_inputs;
