CREATE TABLE workflow_definition_release_candidates (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''), workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''), owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    candidate_id text NOT NULL CHECK (char_length(candidate_id) BETWEEN 1 AND 160), definition_id text NOT NULL CHECK (char_length(definition_id) BETWEEN 1 AND 160),
    candidate_state text NOT NULL CHECK (candidate_state IN ('pending','approved','rejected','superseded')), review_version bigint NOT NULL CHECK (review_version >= 0),
    source_draft_id text NOT NULL, source_draft_version bigint NOT NULL CHECK (source_draft_version > 0), source_draft_digest text NOT NULL CHECK (source_draft_digest ~ '^sha256:[0-9a-f]{64}$'),
    definition_digest text NOT NULL CHECK (definition_digest ~ '^sha256:[0-9a-f]{64}$'), created_at timestamptz NOT NULL, updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    sanitized_candidate_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_candidate_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,definition_id)
);
CREATE TABLE workflow_definition_release_decisions (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL, candidate_id text NOT NULL,
    review_version bigint NOT NULL CHECK (review_version > 0), decision text NOT NULL CHECK (decision IN ('approve','reject')), reviewed_at timestamptz NOT NULL,
    sanitized_decision_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_decision_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,review_version),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id) REFERENCES workflow_definition_release_candidates (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE workflow_definition_versions (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL, definition_id text NOT NULL,
    definition_version bigint NOT NULL CHECK (definition_version > 0), definition_digest text NOT NULL CHECK (definition_digest ~ '^sha256:[0-9a-f]{64}$'),
    candidate_id text NOT NULL, candidate_review_version bigint NOT NULL CHECK (candidate_review_version > 0), created_at timestamptz NOT NULL,
    sanitized_version_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_version_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,definition_version),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,definition_id) REFERENCES workflow_definition_release_candidates (tenant_ref,workspace_id,application_id,owner_subject_ref,candidate_id,definition_id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE workflow_definition_activations (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL, definition_id text NOT NULL,
    pointer_version bigint NOT NULL CHECK (pointer_version > 0), activation_state text NOT NULL CHECK (activation_state IN ('active','inactive')),
    active_version bigint NOT NULL CHECK (active_version >= 0), active_definition_digest text NOT NULL CHECK (active_definition_digest = '' OR active_definition_digest ~ '^sha256:[0-9a-f]{64}$'),
    updated_at timestamptz NOT NULL, sanitized_activation_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_activation_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id),
    CHECK ((activation_state='active' AND active_version>0 AND active_definition_digest<>'') OR (activation_state='inactive' AND active_version=0 AND active_definition_digest=''))
);
CREATE TABLE workflow_definition_activation_events (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL, definition_id text NOT NULL,
    event_id text NOT NULL, after_pointer_version bigint NOT NULL CHECK (after_pointer_version > 0), occurred_at timestamptz NOT NULL,
    sanitized_event_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_event_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,event_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,after_pointer_version),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id) REFERENCES workflow_definition_activations (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE workflow_definition_release_audits (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL, audit_id text NOT NULL,
    resource_kind text NOT NULL CHECK (resource_kind IN ('candidate','version','activation')), resource_id text NOT NULL CHECK (btrim(resource_id)<>''), action text NOT NULL,
    occurred_at timestamptz NOT NULL, sanitized_audit_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,audit_id)
);
CREATE INDEX workflow_definition_candidates_scope_idx ON workflow_definition_release_candidates (tenant_ref,workspace_id,application_id,owner_subject_ref,created_at DESC,candidate_id DESC);
CREATE INDEX workflow_definition_versions_scope_idx ON workflow_definition_versions (tenant_ref,workspace_id,application_id,owner_subject_ref,definition_id,definition_version DESC);
CREATE INDEX workflow_definition_release_audits_scope_idx ON workflow_definition_release_audits (tenant_ref,workspace_id,application_id,owner_subject_ref,occurred_at,audit_id);
CREATE TRIGGER workflow_definition_release_decisions_append_only BEFORE UPDATE OR DELETE ON workflow_definition_release_decisions FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_definition_versions_append_only BEFORE UPDATE OR DELETE ON workflow_definition_versions FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_definition_activation_events_append_only BEFORE UPDATE OR DELETE ON workflow_definition_activation_events FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_definition_release_audits_append_only BEFORE UPDATE OR DELETE ON workflow_definition_release_audits FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
