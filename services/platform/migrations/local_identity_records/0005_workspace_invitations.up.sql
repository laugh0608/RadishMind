CREATE TABLE local_workspace_invitations (
    invitation_id text PRIMARY KEY,
    schema_version text NOT NULL,
    record_version bigint NOT NULL CHECK (record_version > 0),
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    role_key text NOT NULL CHECK (role_key IN ('workspace_reader', 'workspace_builder', 'workspace_reviewer')),
    role_catalog_version text NOT NULL,
    role_definition_digest text NOT NULL CHECK (
        length(role_definition_digest) = 71
        AND role_definition_digest LIKE 'sha256:%'
    ),
    ttl_policy text NOT NULL CHECK (ttl_policy IN ('1h', '24h', '72h', '7d')),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('pending', 'claimed', 'revoked')),
    secret_digest bytea NOT NULL CHECK (octet_length(secret_digest) = 32),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    created_by_actor_ref text NOT NULL,
    created_request_ref text NOT NULL,
    created_audit_ref text NOT NULL,
    updated_request_ref text NOT NULL,
    updated_audit_ref text NOT NULL,
    claimed_at timestamptz,
    claimed_by_user_id text REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    membership_id text REFERENCES local_workspace_memberships(membership_id) ON DELETE RESTRICT,
    assignment_id text REFERENCES local_role_assignments(assignment_id) ON DELETE RESTRICT,
    revoked_at timestamptz,
    revoked_by_actor_ref text,
    CHECK (
        expires_at = created_at + CASE ttl_policy
            WHEN '1h' THEN interval '1 hour'
            WHEN '24h' THEN interval '24 hours'
            WHEN '72h' THEN interval '72 hours'
            WHEN '7d' THEN interval '7 days'
        END
    ),
    CHECK (
        (lifecycle_state = 'pending' AND record_version = 1 AND updated_at = created_at
            AND claimed_at IS NULL AND claimed_by_user_id IS NULL AND membership_id IS NULL
            AND assignment_id IS NULL AND revoked_at IS NULL AND revoked_by_actor_ref IS NULL)
        OR
        (lifecycle_state = 'claimed' AND record_version = 2 AND claimed_at = updated_at
            AND claimed_at < expires_at AND claimed_by_user_id IS NOT NULL AND membership_id IS NOT NULL
            AND assignment_id IS NOT NULL AND revoked_at IS NULL AND revoked_by_actor_ref IS NULL)
        OR
        (lifecycle_state = 'revoked' AND record_version = 2 AND revoked_at = updated_at
            AND revoked_at < expires_at AND revoked_by_actor_ref IS NOT NULL
            AND claimed_at IS NULL AND claimed_by_user_id IS NULL AND membership_id IS NULL
            AND assignment_id IS NULL)
    )
);

CREATE INDEX local_workspace_invitations_directory_idx
    ON local_workspace_invitations(
        tenant_ref, workspace_id, lifecycle_state, updated_at DESC, invitation_id DESC
    );

CREATE INDEX local_workspace_invitations_pending_expiry_idx
    ON local_workspace_invitations(tenant_ref, workspace_id, expires_at)
    WHERE lifecycle_state = 'pending';
