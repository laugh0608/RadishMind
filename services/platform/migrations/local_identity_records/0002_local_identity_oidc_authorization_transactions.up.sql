CREATE TABLE local_identity_oidc_authorization_transactions (
    transaction_id text PRIMARY KEY,
    schema_version text NOT NULL,
    intent text NOT NULL CHECK (intent IN ('login', 'link')),
    user_id text NOT NULL,
    session_id text NOT NULL,
    session_version bigint NOT NULL CHECK (session_version >= 0),
    return_to text NOT NULL,
    state_digest bytea NOT NULL UNIQUE CHECK (octet_length(state_digest) = 32),
    nonce_digest bytea NOT NULL CHECK (octet_length(nonce_digest) = 32),
    policy_digest bytea NOT NULL CHECK (octet_length(policy_digest) = 32),
    code_verifier text NOT NULL CHECK (char_length(code_verifier) <= 128),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('pending', 'consumed', 'expired')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL CHECK (updated_at >= created_at),
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at AND expires_at <= created_at + interval '15 minutes'),
    consumed_at timestamptz,
    audit_ref text NOT NULL,
    CHECK (
        (intent = 'login' AND user_id = '' AND session_id = '' AND session_version = 0)
        OR
        (intent = 'link' AND user_id <> '' AND session_id <> '' AND session_version > 0)
    ),
    CHECK (
        (lifecycle_state = 'pending' AND consumed_at IS NULL AND length(code_verifier) >= 43)
        OR
        (lifecycle_state IN ('consumed', 'expired') AND consumed_at = updated_at AND code_verifier = '')
    )
);

CREATE INDEX local_identity_oidc_authorization_transactions_expiry_idx
    ON local_identity_oidc_authorization_transactions(lifecycle_state, expires_at);
