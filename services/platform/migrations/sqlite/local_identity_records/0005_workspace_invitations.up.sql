CREATE TABLE local_workspace_invitations (
    invitation_id TEXT PRIMARY KEY,
    schema_version TEXT NOT NULL,
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    role_key TEXT NOT NULL CHECK (role_key IN ('workspace_reader', 'workspace_builder', 'workspace_reviewer')),
    role_catalog_version TEXT NOT NULL,
    role_definition_digest TEXT NOT NULL CHECK (
        length(role_definition_digest) = 71
        AND role_definition_digest LIKE 'sha256:%'
    ),
    ttl_policy TEXT NOT NULL CHECK (ttl_policy IN ('1h', '24h', '72h', '7d')),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('pending', 'claimed', 'revoked')),
    secret_digest BLOB NOT NULL CHECK (length(secret_digest) = 32),
    expires_at_unix_nano INTEGER NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    created_by_actor_ref TEXT NOT NULL,
    created_request_ref TEXT NOT NULL,
    created_audit_ref TEXT NOT NULL,
    updated_request_ref TEXT NOT NULL,
    updated_audit_ref TEXT NOT NULL,
    claimed_at_unix_nano INTEGER,
    claimed_by_user_id TEXT REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    membership_id TEXT REFERENCES local_workspace_memberships(membership_id) ON DELETE RESTRICT,
    assignment_id TEXT REFERENCES local_role_assignments(assignment_id) ON DELETE RESTRICT,
    revoked_at_unix_nano INTEGER,
    revoked_by_actor_ref TEXT,
    CHECK (
        expires_at_unix_nano = created_at_unix_nano + CASE ttl_policy
            WHEN '1h' THEN 3600000000000
            WHEN '24h' THEN 86400000000000
            WHEN '72h' THEN 259200000000000
            WHEN '7d' THEN 604800000000000
        END
    ),
    CHECK (
        (lifecycle_state = 'pending' AND record_version = 1 AND updated_at_unix_nano = created_at_unix_nano
            AND claimed_at_unix_nano IS NULL AND claimed_by_user_id IS NULL AND membership_id IS NULL
            AND assignment_id IS NULL AND revoked_at_unix_nano IS NULL AND revoked_by_actor_ref IS NULL)
        OR
        (lifecycle_state = 'claimed' AND record_version = 2 AND claimed_at_unix_nano = updated_at_unix_nano
            AND claimed_at_unix_nano < expires_at_unix_nano AND claimed_by_user_id IS NOT NULL
            AND membership_id IS NOT NULL AND assignment_id IS NOT NULL
            AND revoked_at_unix_nano IS NULL AND revoked_by_actor_ref IS NULL)
        OR
        (lifecycle_state = 'revoked' AND record_version = 2 AND revoked_at_unix_nano = updated_at_unix_nano
            AND revoked_at_unix_nano < expires_at_unix_nano AND revoked_by_actor_ref IS NOT NULL
            AND claimed_at_unix_nano IS NULL AND claimed_by_user_id IS NULL AND membership_id IS NULL
            AND assignment_id IS NULL)
    )
) STRICT;

CREATE INDEX local_workspace_invitations_directory_idx
    ON local_workspace_invitations(
        tenant_ref, workspace_id, lifecycle_state, updated_at_unix_nano DESC, invitation_id DESC
    );

CREATE INDEX local_workspace_invitations_pending_expiry_idx
    ON local_workspace_invitations(tenant_ref, workspace_id, expires_at_unix_nano)
    WHERE lifecycle_state = 'pending';
