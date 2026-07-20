CREATE TABLE workflow_rag_knowledge_promotion_candidates (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL UNIQUE CHECK (candidate_id GLOB 'wragp_*'),
    candidate_digest TEXT NOT NULL CHECK (length(candidate_digest) = 71 AND substr(candidate_digest, 1, 7) = 'sha256:'),
    candidate_state TEXT NOT NULL CHECK (candidate_state IN ('pending', 'deferred', 'approved', 'rejected', 'canceled')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    binding_id TEXT,
    binding_version INTEGER,
    binding_digest TEXT,
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    sanitized_candidate_payload TEXT NOT NULL CHECK (json_valid(sanitized_candidate_payload) AND json_type(sanitized_candidate_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    CHECK ((binding_id IS NULL AND binding_version IS NULL AND binding_digest IS NULL) OR
           (binding_id GLOB 'wragb_*' AND binding_version = 1 AND length(binding_digest) = 71 AND substr(binding_digest, 1, 7) = 'sha256:'))
) STRICT;

CREATE TABLE workflow_rag_knowledge_promotion_decisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    decision_id TEXT NOT NULL UNIQUE CHECK (decision_id GLOB 'wragpd_*'),
    decision TEXT NOT NULL CHECK (decision IN ('approve', 'reject', 'defer', 'cancel')),
    before_record_version INTEGER NOT NULL CHECK (before_record_version > 0),
    after_record_version INTEGER NOT NULL CHECK (after_record_version = before_record_version + 1),
    occurred_at_unix_nano INTEGER NOT NULL,
    sanitized_decision_payload TEXT NOT NULL CHECK (json_valid(sanitized_decision_payload) AND json_type(sanitized_decision_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, decision_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, after_record_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_application_bindings (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    binding_id TEXT NOT NULL UNIQUE CHECK (binding_id GLOB 'wragb_*'),
    binding_version INTEGER NOT NULL CHECK (binding_version = 1),
    binding_digest TEXT NOT NULL CHECK (length(binding_digest) = 71 AND substr(binding_digest, 1, 7) = 'sha256:'),
    approved_decision_id TEXT NOT NULL,
    approved_record_version INTEGER NOT NULL CHECK (approved_record_version > 1),
    issued_at_unix_nano INTEGER NOT NULL,
    sanitized_binding_payload TEXT NOT NULL CHECK (json_valid(sanitized_binding_payload) AND json_type(sanitized_binding_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, binding_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, approved_decision_id)
        REFERENCES workflow_rag_knowledge_promotion_decisions (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, decision_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE workflow_rag_knowledge_promotion_audits (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    event_id TEXT NOT NULL UNIQUE CHECK (event_id GLOB 'wragpa_*'),
    event_sequence INTEGER NOT NULL CHECK (event_sequence > 0),
    event_kind TEXT NOT NULL CHECK (event_kind IN (
        'promotion_candidate_created', 'promotion_decision_approve', 'promotion_decision_reject',
        'promotion_decision_defer', 'promotion_decision_cancel', 'promotion_binding_issued'
    )),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    occurred_at_unix_nano INTEGER NOT NULL,
    sanitized_audit_payload TEXT NOT NULL CHECK (json_valid(sanitized_audit_payload) AND json_type(sanitized_audit_payload) = 'object'),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_sequence),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        REFERENCES workflow_rag_knowledge_promotion_candidates (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE INDEX workflow_rag_knowledge_promotion_candidates_list_idx
    ON workflow_rag_knowledge_promotion_candidates (
        tenant_ref, workspace_id, application_id, owner_subject_ref, created_at_unix_nano DESC, candidate_id DESC
    );
CREATE INDEX workflow_rag_knowledge_promotion_decisions_history_idx
    ON workflow_rag_knowledge_promotion_decisions (
        tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, after_record_version ASC
    );
CREATE INDEX workflow_rag_knowledge_promotion_audits_history_idx
    ON workflow_rag_knowledge_promotion_audits (
        tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id, event_sequence ASC
    );

CREATE TRIGGER workflow_rag_knowledge_promotion_decisions_append_only_update
BEFORE UPDATE ON workflow_rag_knowledge_promotion_decisions BEGIN
    SELECT RAISE(ABORT, 'workflow RAG promotion decisions are append-only');
END;
CREATE TRIGGER workflow_rag_knowledge_promotion_decisions_append_only_delete
BEFORE DELETE ON workflow_rag_knowledge_promotion_decisions BEGIN
    SELECT RAISE(ABORT, 'workflow RAG promotion decisions are append-only');
END;
CREATE TRIGGER workflow_rag_application_bindings_append_only_update
BEFORE UPDATE ON workflow_rag_application_bindings BEGIN
    SELECT RAISE(ABORT, 'workflow RAG application bindings are append-only');
END;
CREATE TRIGGER workflow_rag_application_bindings_append_only_delete
BEFORE DELETE ON workflow_rag_application_bindings BEGIN
    SELECT RAISE(ABORT, 'workflow RAG application bindings are append-only');
END;
CREATE TRIGGER workflow_rag_knowledge_promotion_audits_append_only_update
BEFORE UPDATE ON workflow_rag_knowledge_promotion_audits BEGIN
    SELECT RAISE(ABORT, 'workflow RAG promotion audits are append-only');
END;
CREATE TRIGGER workflow_rag_knowledge_promotion_audits_append_only_delete
BEFORE DELETE ON workflow_rag_knowledge_promotion_audits BEGIN
    SELECT RAISE(ABORT, 'workflow RAG promotion audits are append-only');
END;
