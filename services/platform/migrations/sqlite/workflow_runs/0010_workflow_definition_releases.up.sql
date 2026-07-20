CREATE TABLE workflow_definition_release_candidates (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    owner_subject_ref TEXT NOT NULL CHECK (length(trim(owner_subject_ref)) > 0),
    candidate_id TEXT NOT NULL CHECK (length(candidate_id) BETWEEN 1 AND 160),
    definition_id TEXT NOT NULL CHECK (length(definition_id) BETWEEN 1 AND 160),
    candidate_state TEXT NOT NULL CHECK (candidate_state IN ('pending', 'approved', 'rejected', 'superseded')),
    review_version INTEGER NOT NULL CHECK (review_version >= 0),
    source_draft_id TEXT NOT NULL CHECK (length(source_draft_id) BETWEEN 1 AND 160),
    source_draft_version INTEGER NOT NULL CHECK (source_draft_version > 0),
    source_draft_digest TEXT NOT NULL CHECK (length(source_draft_digest) = 71 AND substr(source_draft_digest, 1, 7) = 'sha256:'),
    definition_digest TEXT NOT NULL CHECK (length(definition_digest) = 71 AND substr(definition_digest, 1, 7) = 'sha256:'),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    sanitized_candidate_payload TEXT NOT NULL CHECK (json_valid(sanitized_candidate_payload) AND json_type(sanitized_candidate_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, definition_id)
) STRICT;

CREATE TABLE workflow_definition_release_decisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    review_version INTEGER NOT NULL CHECK (review_version > 0),
    decision TEXT NOT NULL CHECK (decision IN ('approve', 'reject')),
    reviewed_at_unix_nano INTEGER NOT NULL CHECK (reviewed_at_unix_nano > 0),
    sanitized_decision_payload TEXT NOT NULL CHECK (json_valid(sanitized_decision_payload) AND json_type(sanitized_decision_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, review_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_definition_release_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_definition_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    definition_id TEXT NOT NULL,
    definition_version INTEGER NOT NULL CHECK (definition_version > 0),
    definition_digest TEXT NOT NULL CHECK (length(definition_digest) = 71 AND substr(definition_digest, 1, 7) = 'sha256:'),
    candidate_id TEXT NOT NULL,
    candidate_review_version INTEGER NOT NULL CHECK (candidate_review_version > 0),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    sanitized_version_payload TEXT NOT NULL CHECK (json_valid(sanitized_version_payload) AND json_type(sanitized_version_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id, definition_version),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, definition_id)
        REFERENCES workflow_definition_release_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, definition_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_definition_activations (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    definition_id TEXT NOT NULL,
    pointer_version INTEGER NOT NULL CHECK (pointer_version > 0),
    activation_state TEXT NOT NULL CHECK (activation_state IN ('active', 'inactive')),
    active_version INTEGER NOT NULL CHECK (active_version >= 0),
    active_definition_digest TEXT NOT NULL CHECK (active_definition_digest = '' OR (length(active_definition_digest) = 71 AND substr(active_definition_digest, 1, 7) = 'sha256:')),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_activation_payload TEXT NOT NULL CHECK (json_valid(sanitized_activation_payload) AND json_type(sanitized_activation_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id),
    CHECK ((activation_state = 'active' AND active_version > 0 AND active_definition_digest <> '') OR (activation_state = 'inactive' AND active_version = 0 AND active_definition_digest = ''))
) STRICT;

CREATE TABLE workflow_definition_activation_events (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    definition_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    after_pointer_version INTEGER NOT NULL CHECK (after_pointer_version > 0),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_event_payload TEXT NOT NULL CHECK (json_valid(sanitized_event_payload) AND json_type(sanitized_event_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id, after_pointer_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id)
        REFERENCES workflow_definition_activations (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_definition_release_audits (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    audit_id TEXT NOT NULL,
    resource_kind TEXT NOT NULL CHECK (resource_kind IN ('candidate', 'version', 'activation')),
    resource_id TEXT NOT NULL CHECK (length(trim(resource_id)) > 0),
    action TEXT NOT NULL,
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_audit_payload TEXT NOT NULL CHECK (json_valid(sanitized_audit_payload) AND json_type(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, audit_id)
) STRICT;

CREATE INDEX workflow_definition_candidates_scope_idx ON workflow_definition_release_candidates
    (tenant_ref, workspace_id, application_id, owner_subject_ref, created_at_unix_nano DESC, candidate_id DESC);
CREATE INDEX workflow_definition_versions_scope_idx ON workflow_definition_versions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, definition_id, definition_version DESC);
CREATE INDEX workflow_definition_release_audits_scope_idx ON workflow_definition_release_audits
    (tenant_ref, workspace_id, application_id, owner_subject_ref, occurred_at_unix_nano, audit_id);

CREATE TRIGGER workflow_definition_release_decisions_append_only_update BEFORE UPDATE ON workflow_definition_release_decisions BEGIN SELECT RAISE(ABORT, 'workflow definition decisions are append-only'); END;
CREATE TRIGGER workflow_definition_release_decisions_append_only_delete BEFORE DELETE ON workflow_definition_release_decisions BEGIN SELECT RAISE(ABORT, 'workflow definition decisions are append-only'); END;
CREATE TRIGGER workflow_definition_versions_append_only_update BEFORE UPDATE ON workflow_definition_versions BEGIN SELECT RAISE(ABORT, 'workflow definition versions are immutable'); END;
CREATE TRIGGER workflow_definition_versions_append_only_delete BEFORE DELETE ON workflow_definition_versions BEGIN SELECT RAISE(ABORT, 'workflow definition versions are immutable'); END;
CREATE TRIGGER workflow_definition_activation_events_append_only_update BEFORE UPDATE ON workflow_definition_activation_events BEGIN SELECT RAISE(ABORT, 'workflow definition activation events are append-only'); END;
CREATE TRIGGER workflow_definition_activation_events_append_only_delete BEFORE DELETE ON workflow_definition_activation_events BEGIN SELECT RAISE(ABORT, 'workflow definition activation events are append-only'); END;
CREATE TRIGGER workflow_definition_release_audits_append_only_update BEFORE UPDATE ON workflow_definition_release_audits BEGIN SELECT RAISE(ABORT, 'workflow definition release audits are append-only'); END;
CREATE TRIGGER workflow_definition_release_audits_append_only_delete BEFORE DELETE ON workflow_definition_release_audits BEGIN SELECT RAISE(ABORT, 'workflow definition release audits are append-only'); END;
