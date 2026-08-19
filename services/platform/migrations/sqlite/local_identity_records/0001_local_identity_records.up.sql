CREATE TABLE local_user_accounts (
    user_id TEXT PRIMARY KEY,
    schema_version TEXT NOT NULL,
    login_identifier TEXT NOT NULL,
    normalized_login_identifier TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'disabled')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    disabled_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND disabled_at_unix_nano IS NULL)
        OR
        (lifecycle_state = 'disabled' AND disabled_at_unix_nano = updated_at_unix_nano)
    )
) STRICT;

CREATE TABLE local_credentials (
    credential_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    iterations INTEGER NOT NULL CHECK (iterations > 0),
    key_length INTEGER NOT NULL CHECK (key_length > 0),
    salt BLOB NOT NULL CHECK (length(salt) >= 16),
    derived_key BLOB NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'superseded', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    audit_ref TEXT NOT NULL,
    CHECK (length(derived_key) = key_length)
) STRICT;

CREATE UNIQUE INDEX local_credentials_active_user_idx
    ON local_credentials(user_id) WHERE lifecycle_state = 'active';

CREATE TABLE external_identity_bindings (
    binding_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version TEXT NOT NULL,
    issuer TEXT NOT NULL,
    subject_value TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    revoked_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND revoked_at_unix_nano IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at_unix_nano = updated_at_unix_nano)
    )
) STRICT;

CREATE UNIQUE INDEX external_identity_bindings_active_subject_idx
    ON external_identity_bindings(issuer, subject_value) WHERE lifecycle_state = 'active';

CREATE TABLE local_web_sessions (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version TEXT NOT NULL,
    credential_digest BLOB NOT NULL UNIQUE CHECK (length(credential_digest) = 32),
    authentication_method TEXT NOT NULL CHECK (authentication_method IN ('local_password', 'oidc')),
    authentication_source_ref TEXT NOT NULL,
    policy_version TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    last_verified_at_unix_nano INTEGER NOT NULL CHECK (last_verified_at_unix_nano >= created_at_unix_nano),
    expires_at_unix_nano INTEGER NOT NULL CHECK (expires_at_unix_nano > created_at_unix_nano),
    revoked_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND revoked_at_unix_nano IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at_unix_nano = updated_at_unix_nano)
    )
) STRICT;

CREATE INDEX local_web_sessions_active_user_idx
    ON local_web_sessions(user_id, expires_at_unix_nano) WHERE lifecycle_state = 'active';

CREATE TABLE local_role_assignments (
    assignment_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version TEXT NOT NULL,
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    role_key TEXT NOT NULL,
    permission_grants_json TEXT NOT NULL CHECK (
        json_valid(permission_grants_json)
        AND json_type(permission_grants_json) = 'array'
        AND json_array_length(permission_grants_json) > 0
    ),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    expires_at_unix_nano INTEGER,
    revoked_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (expires_at_unix_nano IS NULL OR expires_at_unix_nano > created_at_unix_nano),
    CHECK (
        (lifecycle_state = 'active' AND revoked_at_unix_nano IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at_unix_nano = updated_at_unix_nano)
    )
) STRICT;

CREATE UNIQUE INDEX local_role_assignments_active_scope_idx
    ON local_role_assignments(user_id, tenant_ref, workspace_id, role_key)
    WHERE lifecycle_state = 'active';

CREATE INDEX local_role_assignments_authorization_idx
    ON local_role_assignments(user_id, tenant_ref, workspace_id, lifecycle_state, expires_at_unix_nano);

CREATE TABLE local_workspace_memberships (
    membership_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version TEXT NOT NULL,
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    expires_at_unix_nano INTEGER,
    revoked_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (expires_at_unix_nano IS NULL OR expires_at_unix_nano > created_at_unix_nano),
    CHECK (
        (lifecycle_state = 'active' AND revoked_at_unix_nano IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at_unix_nano = updated_at_unix_nano)
    )
) STRICT;

CREATE UNIQUE INDEX local_workspace_memberships_active_scope_idx
    ON local_workspace_memberships(user_id, tenant_ref, workspace_id)
    WHERE lifecycle_state = 'active';

CREATE INDEX local_workspace_memberships_authorization_idx
    ON local_workspace_memberships(user_id, tenant_ref, workspace_id, lifecycle_state, expires_at_unix_nano);
