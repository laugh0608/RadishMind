ALTER TABLE saved_workflow_drafts
    DROP CONSTRAINT saved_workflow_drafts_provenance_kind_check;

ALTER TABLE saved_workflow_drafts
    ADD CONSTRAINT saved_workflow_drafts_provenance_kind_check
        CHECK (
            provenance_kind IN (
                'unversioned',
                'workflow_definition',
                'saved_draft_derivation',
                'workspace_template_derivation'
            )
        );

CREATE TABLE workflow_template_candidates (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    candidate_id text NOT NULL,
    template_id text NOT NULL,
    candidate_state text NOT NULL CHECK (
        candidate_state IN ('pending', 'approved', 'rejected', 'changes_requested', 'withdrawn')
    ),
    review_version bigint NOT NULL CHECK (review_version >= 0),
    source_application_id text NOT NULL,
    source_owner_subject_ref text NOT NULL,
    source_definition_id text NOT NULL,
    source_definition_version bigint NOT NULL CHECK (source_definition_version > 0),
    source_definition_digest text NOT NULL CHECK (
        source_definition_digest ~ '^sha256:[a-f0-9]{64}$'
    ),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    sanitized_candidate_payload jsonb NOT NULL CHECK (
        sanitized_candidate_payload ->> 'schema_version' = 'workspace_workflow_template_candidate.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, candidate_id)
);

CREATE TABLE workflow_template_decisions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    candidate_id text NOT NULL,
    review_version bigint NOT NULL CHECK (review_version > 0),
    decision text NOT NULL CHECK (
        decision IN ('approve', 'reject', 'request_changes', 'withdraw')
    ),
    decided_at timestamptz NOT NULL,
    sanitized_decision_payload jsonb NOT NULL CHECK (
        sanitized_decision_payload ->> 'schema_version' = 'workspace_workflow_template_decision.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, candidate_id, review_version),
    FOREIGN KEY (tenant_ref, workspace_id, candidate_id)
        REFERENCES workflow_template_candidates (tenant_ref, workspace_id, candidate_id)
        ON DELETE RESTRICT
);

CREATE TABLE workflow_template_versions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    template_id text NOT NULL,
    template_version bigint NOT NULL CHECK (template_version > 0),
    template_digest text NOT NULL CHECK (template_digest ~ '^sha256:[a-f0-9]{64}$'),
    candidate_id text NOT NULL,
    candidate_review_version bigint NOT NULL CHECK (candidate_review_version > 0),
    source_definition_id text NOT NULL,
    source_definition_version bigint NOT NULL CHECK (source_definition_version > 0),
    source_definition_digest text NOT NULL CHECK (
        source_definition_digest ~ '^sha256:[a-f0-9]{64}$'
    ),
    created_at timestamptz NOT NULL,
    sanitized_version_payload jsonb NOT NULL CHECK (
        sanitized_version_payload ->> 'schema_version' = 'workspace_workflow_template_version.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, template_id, template_version),
    UNIQUE (tenant_ref, workspace_id, candidate_id),
    FOREIGN KEY (tenant_ref, workspace_id, candidate_id)
        REFERENCES workflow_template_candidates (tenant_ref, workspace_id, candidate_id)
        ON DELETE RESTRICT
);

CREATE TABLE workflow_template_lineages (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    template_id text NOT NULL,
    pointer_version bigint NOT NULL CHECK (pointer_version >= 0),
    lifecycle text NOT NULL CHECK (lifecycle IN ('unlisted', 'listed')),
    listed_version bigint NOT NULL CHECK (listed_version >= 0),
    listed_digest text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    sanitized_lineage_payload jsonb NOT NULL CHECK (
        sanitized_lineage_payload ->> 'schema_version' = 'workspace_workflow_template_lineage.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, template_id),
    CHECK (
        (lifecycle = 'unlisted' AND listed_version = 0 AND listed_digest = '') OR
        (lifecycle = 'listed' AND listed_version > 0 AND listed_digest ~ '^sha256:[a-f0-9]{64}$')
    )
);

CREATE TABLE workflow_template_listing_events (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    template_id text NOT NULL,
    event_id text NOT NULL,
    after_pointer_version bigint NOT NULL CHECK (after_pointer_version > 0),
    decision text NOT NULL CHECK (decision IN ('list', 'replace', 'unlist')),
    occurred_at timestamptz NOT NULL,
    sanitized_event_payload jsonb NOT NULL CHECK (
        sanitized_event_payload ->> 'schema_version' = 'workspace_workflow_template_listing_event.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, template_id, after_pointer_version),
    UNIQUE (tenant_ref, workspace_id, template_id, event_id),
    FOREIGN KEY (tenant_ref, workspace_id, template_id)
        REFERENCES workflow_template_lineages (tenant_ref, workspace_id, template_id)
        ON DELETE RESTRICT
);

CREATE TABLE workflow_template_audits (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    audit_id text NOT NULL,
    audit_sequence bigint NOT NULL CHECK (audit_sequence > 0),
    resource_kind text NOT NULL CHECK (resource_kind IN ('candidate', 'version', 'lineage')),
    resource_id text NOT NULL,
    action text NOT NULL CHECK (
        action IN (
            'create', 'review_approve', 'review_reject', 'review_request_changes',
            'review_withdraw', 'materialize', 'list', 'replace', 'unlist'
        )
    ),
    occurred_at timestamptz NOT NULL,
    sanitized_audit_payload jsonb NOT NULL CHECK (
        sanitized_audit_payload ->> 'schema_version' = 'workspace_workflow_template_audit.v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, audit_id),
    UNIQUE (tenant_ref, workspace_id, audit_sequence)
);

CREATE INDEX workflow_template_candidates_workspace_list_idx
    ON workflow_template_candidates (
        tenant_ref, workspace_id, candidate_state, updated_at DESC, candidate_id DESC
    );
CREATE INDEX workflow_template_versions_workspace_list_idx
    ON workflow_template_versions (
        tenant_ref, workspace_id, template_id, created_at DESC, template_version DESC
    );
CREATE INDEX workflow_template_lineages_workspace_list_idx
    ON workflow_template_lineages (
        tenant_ref, workspace_id, lifecycle, updated_at DESC, template_id DESC
    );
CREATE INDEX workflow_template_audits_workspace_order_idx
    ON workflow_template_audits (tenant_ref, workspace_id, audit_sequence);

CREATE FUNCTION reject_workflow_template_history_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'workflow template history is append-only'
        USING ERRCODE = '42501';
END;
$$;

CREATE TRIGGER workflow_template_decisions_append_only
BEFORE UPDATE OR DELETE ON workflow_template_decisions
FOR EACH STATEMENT EXECUTE FUNCTION reject_workflow_template_history_mutation();
CREATE TRIGGER workflow_template_versions_append_only
BEFORE UPDATE OR DELETE ON workflow_template_versions
FOR EACH STATEMENT EXECUTE FUNCTION reject_workflow_template_history_mutation();
CREATE TRIGGER workflow_template_listing_events_append_only
BEFORE UPDATE OR DELETE ON workflow_template_listing_events
FOR EACH STATEMENT EXECUTE FUNCTION reject_workflow_template_history_mutation();
CREATE TRIGGER workflow_template_audits_append_only
BEFORE UPDATE OR DELETE ON workflow_template_audits
FOR EACH STATEMENT EXECUTE FUNCTION reject_workflow_template_history_mutation();
