CREATE TABLE application_publish_candidates (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    candidate_id TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    draft_id TEXT NOT NULL,
    draft_version INTEGER NOT NULL CHECK (draft_version > 0),
    draft_digest TEXT NOT NULL,
    candidate_state TEXT NOT NULL CHECK (candidate_state IN ('pending_review', 'approved', 'rejected', 'changes_requested', 'withdrawn')),
    review_version INTEGER NOT NULL CHECK (review_version >= 0),
    sanitized_candidate_payload TEXT NOT NULL CHECK (json_valid(sanitized_candidate_payload)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    created_by_actor_ref TEXT NOT NULL,
    updated_by_actor_ref TEXT NOT NULL,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, candidate_id)
) STRICT;

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
