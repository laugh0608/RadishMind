CREATE TABLE workflow_rag_knowledge_promotion_candidates (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    candidate_id text NOT NULL UNIQUE CHECK (candidate_id ~ '^wragp_[a-z2-7]{16}$'),
    candidate_digest text NOT NULL CHECK (candidate_digest ~ '^sha256:[0-9a-f]{64}$'),
    candidate_state text NOT NULL CHECK (candidate_state IN ('pending', 'deferred', 'approved', 'rejected', 'canceled')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    binding_id text,
    binding_version bigint,
    binding_digest text,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    sanitized_candidate_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_candidate_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    CHECK ((binding_id IS NULL AND binding_version IS NULL AND binding_digest IS NULL) OR
           (binding_id ~ '^wragb_[a-z2-7]{16}$' AND binding_version = 1 AND binding_digest ~ '^sha256:[0-9a-f]{64}$'))
);

CREATE TABLE workflow_rag_knowledge_promotion_decisions (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    candidate_id text NOT NULL CHECK (candidate_id ~ '^wragp_[a-z2-7]{16}$'),
    decision_id text NOT NULL UNIQUE CHECK (decision_id ~ '^wragpd_[a-z2-7]{16}$'),
    decision text NOT NULL CHECK (decision IN ('approve', 'reject', 'defer', 'cancel')),
    before_record_version bigint NOT NULL CHECK (before_record_version > 0),
    after_record_version bigint NOT NULL CHECK (after_record_version = before_record_version + 1),
    occurred_at timestamptz NOT NULL,
    sanitized_decision_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_decision_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, decision_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, after_record_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_application_bindings (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    candidate_id text NOT NULL CHECK (candidate_id ~ '^wragp_[a-z2-7]{16}$'),
    binding_id text NOT NULL UNIQUE CHECK (binding_id ~ '^wragb_[a-z2-7]{16}$'),
    binding_version bigint NOT NULL CHECK (binding_version = 1),
    binding_digest text NOT NULL CHECK (binding_digest ~ '^sha256:[0-9a-f]{64}$'),
    approved_decision_id text NOT NULL CHECK (approved_decision_id ~ '^wragpd_[a-z2-7]{16}$'),
    approved_record_version bigint NOT NULL CHECK (approved_record_version > 1),
    issued_at timestamptz NOT NULL,
    sanitized_binding_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_binding_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, binding_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, approved_decision_id)
        REFERENCES workflow_rag_knowledge_promotion_decisions (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, decision_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE workflow_rag_knowledge_promotion_audits (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    candidate_id text NOT NULL CHECK (candidate_id ~ '^wragp_[a-z2-7]{16}$'),
    event_id text NOT NULL UNIQUE CHECK (event_id ~ '^wragpa_[a-z2-7]{16}$'),
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    event_kind text NOT NULL CHECK (event_kind IN (
        'promotion_candidate_created', 'promotion_decision_approve', 'promotion_decision_reject',
        'promotion_decision_defer', 'promotion_decision_cancel', 'promotion_binding_issued'
    )),
    record_version bigint NOT NULL CHECK (record_version > 0),
    occurred_at timestamptz NOT NULL,
    sanitized_audit_payload jsonb NOT NULL CHECK (jsonb_typeof(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_sequence),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX workflow_rag_knowledge_promotion_candidates_list_idx
    ON workflow_rag_knowledge_promotion_candidates (
        tenant_ref, workspace_id, application_id, owner_subject_ref, created_at DESC, candidate_id DESC
    );
CREATE INDEX workflow_rag_knowledge_promotion_decisions_history_idx
    ON workflow_rag_knowledge_promotion_decisions (
        tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, after_record_version ASC
    );
CREATE INDEX workflow_rag_knowledge_promotion_audits_history_idx
    ON workflow_rag_knowledge_promotion_audits (
        tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_sequence ASC
    );

CREATE TRIGGER workflow_rag_knowledge_promotion_decisions_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_knowledge_promotion_decisions
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_application_bindings_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_application_bindings
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER workflow_rag_knowledge_promotion_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_knowledge_promotion_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
