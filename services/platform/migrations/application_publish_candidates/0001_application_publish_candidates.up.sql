CREATE TABLE application_publish_candidates (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    candidate_id text NOT NULL,
    schema_version text NOT NULL,
    draft_id text NOT NULL,
    draft_version bigint NOT NULL CHECK (draft_version > 0),
    draft_digest text NOT NULL,
    candidate_state text NOT NULL,
    review_version bigint NOT NULL CHECK (review_version >= 0),
    sanitized_candidate_payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    created_by_actor_ref text NOT NULL,
    updated_by_actor_ref text NOT NULL,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
);

CREATE INDEX application_publish_candidates_scope_list_idx
    ON application_publish_candidates (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        created_at DESC,
        candidate_id DESC
    );

CREATE INDEX application_publish_candidates_draft_idx
    ON application_publish_candidates (
        tenant_ref,
        workspace_id,
        application_id,
        owner_subject_ref,
        draft_id,
        draft_version DESC,
        created_at DESC
    );
