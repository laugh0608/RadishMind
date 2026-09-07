PRAGMA defer_foreign_keys = ON;

DROP INDEX saved_workflow_draft_revisions_owner_version_idx;
DROP INDEX saved_workflow_draft_lifecycle_events_owner_version_idx;
DROP INDEX saved_workflow_drafts_owner_lifecycle_list_idx;
DROP INDEX saved_workflow_drafts_validation_list_idx;
DROP INDEX saved_workflow_drafts_provenance_list_idx;
DROP INDEX saved_workflow_drafts_name_list_idx;
DROP INDEX saved_workflow_drafts_schema_version_idx;

ALTER TABLE saved_workflow_draft_revisions RENAME TO saved_workflow_draft_revisions_pre_template_catalog;
ALTER TABLE saved_workflow_draft_lifecycle_events RENAME TO saved_workflow_draft_lifecycle_events_pre_template_catalog;
ALTER TABLE saved_workflow_drafts RENAME TO saved_workflow_drafts_pre_template_catalog;

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
    provenance_kind TEXT NOT NULL DEFAULT 'unversioned' CHECK (
        provenance_kind IN ('unversioned', 'workflow_definition', 'saved_draft_derivation', 'workspace_template_derivation')
    ),
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
FROM saved_workflow_drafts_pre_template_catalog;

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
INSERT INTO saved_workflow_draft_revisions SELECT * FROM saved_workflow_draft_revisions_pre_template_catalog;
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
INSERT INTO saved_workflow_draft_lifecycle_events SELECT * FROM saved_workflow_draft_lifecycle_events_pre_template_catalog;
CREATE INDEX saved_workflow_draft_lifecycle_events_owner_version_idx ON saved_workflow_draft_lifecycle_events (tenant_ref, workspace_id, application_id, owner_subject_ref, draft_id, lifecycle_version DESC);

DROP TABLE saved_workflow_draft_revisions_pre_template_catalog;
DROP TABLE saved_workflow_draft_lifecycle_events_pre_template_catalog;
DROP TABLE saved_workflow_drafts_pre_template_catalog;

CREATE TABLE workflow_template_candidates (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    candidate_state TEXT NOT NULL CHECK (candidate_state IN ('pending', 'approved', 'rejected', 'changes_requested', 'withdrawn')),
    review_version INTEGER NOT NULL CHECK (review_version >= 0),
    source_application_id TEXT NOT NULL,
    source_owner_subject_ref TEXT NOT NULL,
    source_definition_id TEXT NOT NULL,
    source_definition_version INTEGER NOT NULL CHECK (source_definition_version > 0),
    source_definition_digest TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    sanitized_candidate_payload TEXT NOT NULL CHECK (json_valid(sanitized_candidate_payload) AND json_extract(sanitized_candidate_payload, '$.schema_version') = 'workspace_workflow_template_candidate.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, candidate_id)
) STRICT;

CREATE TABLE workflow_template_decisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    review_version INTEGER NOT NULL CHECK (review_version > 0),
    decision TEXT NOT NULL CHECK (decision IN ('approve', 'reject', 'request_changes', 'withdraw')),
    decided_at_unix_nano INTEGER NOT NULL,
    sanitized_decision_payload TEXT NOT NULL CHECK (json_valid(sanitized_decision_payload) AND json_extract(sanitized_decision_payload, '$.schema_version') = 'workspace_workflow_template_decision.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, candidate_id, review_version),
    FOREIGN KEY (tenant_ref, workspace_id, candidate_id) REFERENCES workflow_template_candidates (tenant_ref, workspace_id, candidate_id) ON DELETE RESTRICT
) STRICT;

CREATE TABLE workflow_template_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    template_version INTEGER NOT NULL CHECK (template_version > 0),
    template_digest TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    candidate_review_version INTEGER NOT NULL CHECK (candidate_review_version > 0),
    source_definition_id TEXT NOT NULL,
    source_definition_version INTEGER NOT NULL CHECK (source_definition_version > 0),
    source_definition_digest TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    sanitized_version_payload TEXT NOT NULL CHECK (json_valid(sanitized_version_payload) AND json_extract(sanitized_version_payload, '$.schema_version') = 'workspace_workflow_template_version.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, template_id, template_version),
    UNIQUE (tenant_ref, workspace_id, candidate_id),
    FOREIGN KEY (tenant_ref, workspace_id, candidate_id) REFERENCES workflow_template_candidates (tenant_ref, workspace_id, candidate_id) ON DELETE RESTRICT
) STRICT;

CREATE TABLE workflow_template_lineages (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    pointer_version INTEGER NOT NULL CHECK (pointer_version >= 0),
    lifecycle TEXT NOT NULL CHECK (lifecycle IN ('unlisted', 'listed')),
    listed_version INTEGER NOT NULL CHECK (listed_version >= 0),
    listed_digest TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    sanitized_lineage_payload TEXT NOT NULL CHECK (json_valid(sanitized_lineage_payload) AND json_extract(sanitized_lineage_payload, '$.schema_version') = 'workspace_workflow_template_lineage.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, template_id),
    CHECK ((lifecycle = 'unlisted' AND listed_version = 0 AND listed_digest = '') OR (lifecycle = 'listed' AND listed_version > 0))
) STRICT;

CREATE TABLE workflow_template_listing_events (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    after_pointer_version INTEGER NOT NULL CHECK (after_pointer_version > 0),
    decision TEXT NOT NULL CHECK (decision IN ('list', 'replace', 'unlist')),
    occurred_at_unix_nano INTEGER NOT NULL,
    sanitized_event_payload TEXT NOT NULL CHECK (json_valid(sanitized_event_payload) AND json_extract(sanitized_event_payload, '$.schema_version') = 'workspace_workflow_template_listing_event.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, template_id, after_pointer_version),
    UNIQUE (tenant_ref, workspace_id, template_id, event_id),
    FOREIGN KEY (tenant_ref, workspace_id, template_id) REFERENCES workflow_template_lineages (tenant_ref, workspace_id, template_id) ON DELETE RESTRICT
) STRICT;

CREATE TABLE workflow_template_audits (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    audit_id TEXT NOT NULL,
    audit_sequence INTEGER NOT NULL CHECK (audit_sequence > 0),
    resource_kind TEXT NOT NULL CHECK (resource_kind IN ('candidate', 'version', 'lineage')),
    resource_id TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'review_approve', 'review_reject', 'review_request_changes', 'review_withdraw', 'materialize', 'list', 'replace', 'unlist')),
    occurred_at_unix_nano INTEGER NOT NULL,
    sanitized_audit_payload TEXT NOT NULL CHECK (json_valid(sanitized_audit_payload) AND json_extract(sanitized_audit_payload, '$.schema_version') = 'workspace_workflow_template_audit.v1'),
    PRIMARY KEY (tenant_ref, workspace_id, audit_id),
    UNIQUE (tenant_ref, workspace_id, audit_sequence)
) STRICT;

CREATE INDEX workflow_template_candidates_workspace_list_idx ON workflow_template_candidates (tenant_ref, workspace_id, candidate_state, updated_at_unix_nano DESC, candidate_id DESC);
CREATE INDEX workflow_template_versions_workspace_list_idx ON workflow_template_versions (tenant_ref, workspace_id, template_id, created_at_unix_nano DESC, template_version DESC);
CREATE INDEX workflow_template_lineages_workspace_list_idx ON workflow_template_lineages (tenant_ref, workspace_id, lifecycle, updated_at_unix_nano DESC, template_id DESC);
CREATE INDEX workflow_template_audits_workspace_order_idx ON workflow_template_audits (tenant_ref, workspace_id, audit_sequence);

CREATE TRIGGER workflow_template_decisions_no_update BEFORE UPDATE ON workflow_template_decisions BEGIN SELECT RAISE(ABORT, 'workflow template decisions are append-only'); END;
CREATE TRIGGER workflow_template_decisions_no_delete BEFORE DELETE ON workflow_template_decisions BEGIN SELECT RAISE(ABORT, 'workflow template decisions are append-only'); END;
CREATE TRIGGER workflow_template_versions_no_update BEFORE UPDATE ON workflow_template_versions BEGIN SELECT RAISE(ABORT, 'workflow template versions are immutable'); END;
CREATE TRIGGER workflow_template_versions_no_delete BEFORE DELETE ON workflow_template_versions BEGIN SELECT RAISE(ABORT, 'workflow template versions are immutable'); END;
CREATE TRIGGER workflow_template_listing_events_no_update BEFORE UPDATE ON workflow_template_listing_events BEGIN SELECT RAISE(ABORT, 'workflow template listing events are append-only'); END;
CREATE TRIGGER workflow_template_listing_events_no_delete BEFORE DELETE ON workflow_template_listing_events BEGIN SELECT RAISE(ABORT, 'workflow template listing events are append-only'); END;
CREATE TRIGGER workflow_template_audits_no_update BEFORE UPDATE ON workflow_template_audits BEGIN SELECT RAISE(ABORT, 'workflow template audits are append-only'); END;
CREATE TRIGGER workflow_template_audits_no_delete BEFORE DELETE ON workflow_template_audits BEGIN SELECT RAISE(ABORT, 'workflow template audits are append-only'); END;
