CREATE TABLE local_user_accounts (
    user_id text PRIMARY KEY,
    schema_version text NOT NULL,
    login_identifier text NOT NULL,
    normalized_login_identifier text NOT NULL UNIQUE,
    display_name text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'disabled')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    disabled_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND disabled_at IS NULL)
        OR
        (lifecycle_state = 'disabled' AND disabled_at = updated_at)
    )
);

CREATE TABLE local_credentials (
    credential_id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version text NOT NULL,
    algorithm text NOT NULL,
    policy_version text NOT NULL,
    iterations integer NOT NULL CHECK (iterations > 0),
    key_length integer NOT NULL CHECK (key_length > 0),
    salt bytea NOT NULL CHECK (octet_length(salt) >= 16),
    derived_key bytea NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'superseded', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    audit_ref text NOT NULL,
    CHECK (octet_length(derived_key) = key_length)
);

CREATE UNIQUE INDEX local_credentials_active_user_idx
    ON local_credentials(user_id) WHERE lifecycle_state = 'active';

CREATE TABLE external_identity_bindings (
    binding_id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version text NOT NULL,
    issuer text NOT NULL,
    subject_value text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    revoked_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND revoked_at IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at = updated_at)
    )
);

CREATE UNIQUE INDEX external_identity_bindings_active_subject_idx
    ON external_identity_bindings(issuer, subject_value) WHERE lifecycle_state = 'active';

CREATE TABLE local_web_sessions (
    session_id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version text NOT NULL,
    credential_digest bytea NOT NULL UNIQUE CHECK (octet_length(credential_digest) = 32),
    authentication_method text NOT NULL CHECK (authentication_method IN ('local_password', 'oidc')),
    authentication_source_ref text NOT NULL,
    policy_version text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    last_verified_at timestamptz NOT NULL CHECK (last_verified_at >= created_at),
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
    revoked_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (
        (lifecycle_state = 'active' AND revoked_at IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at = updated_at)
    )
);

CREATE INDEX local_web_sessions_active_user_idx
    ON local_web_sessions(user_id, expires_at) WHERE lifecycle_state = 'active';

CREATE TABLE local_role_assignments (
    assignment_id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version text NOT NULL,
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    role_key text NOT NULL,
    permission_grants text[] NOT NULL CHECK (cardinality(permission_grants) > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    expires_at timestamptz,
    revoked_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (expires_at IS NULL OR expires_at > created_at),
    CHECK (
        (lifecycle_state = 'active' AND revoked_at IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at = updated_at)
    )
);

CREATE UNIQUE INDEX local_role_assignments_active_scope_idx
    ON local_role_assignments(user_id, tenant_ref, workspace_id, role_key)
    WHERE lifecycle_state = 'active';

CREATE INDEX local_role_assignments_authorization_idx
    ON local_role_assignments(user_id, tenant_ref, workspace_id, lifecycle_state, expires_at);

CREATE TABLE local_workspace_memberships (
    membership_id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES local_user_accounts(user_id) ON DELETE RESTRICT,
    schema_version text NOT NULL,
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    expires_at timestamptz,
    revoked_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (expires_at IS NULL OR expires_at > created_at),
    CHECK (
        (lifecycle_state = 'active' AND revoked_at IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at = updated_at)
    )
);

CREATE UNIQUE INDEX local_workspace_memberships_active_scope_idx
    ON local_workspace_memberships(user_id, tenant_ref, workspace_id)
    WHERE lifecycle_state = 'active';

CREATE INDEX local_workspace_memberships_authorization_idx
    ON local_workspace_memberships(user_id, tenant_ref, workspace_id, lifecycle_state, expires_at);
