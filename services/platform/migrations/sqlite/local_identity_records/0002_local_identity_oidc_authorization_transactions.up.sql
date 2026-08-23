CREATE TABLE local_identity_oidc_authorization_transactions (
    transaction_id TEXT PRIMARY KEY,
    schema_version TEXT NOT NULL,
    intent TEXT NOT NULL CHECK (intent IN ('login', 'link')),
    user_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    session_version INTEGER NOT NULL CHECK (session_version >= 0),
    return_to TEXT NOT NULL,
    state_digest BLOB NOT NULL UNIQUE CHECK (length(state_digest) = 32),
    nonce_digest BLOB NOT NULL CHECK (length(nonce_digest) = 32),
    policy_digest BLOB NOT NULL CHECK (length(policy_digest) = 32),
    code_verifier TEXT NOT NULL CHECK (length(code_verifier) <= 128),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('pending', 'consumed', 'expired')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    created_at_unix_nano INTEGER NOT NULL,
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano >= created_at_unix_nano),
    expires_at_unix_nano INTEGER NOT NULL CHECK (
        expires_at_unix_nano > created_at_unix_nano
        AND expires_at_unix_nano - created_at_unix_nano <= 900000000000
    ),
    consumed_at_unix_nano INTEGER,
    audit_ref TEXT NOT NULL,
    CHECK (
        (intent = 'login' AND user_id = '' AND session_id = '' AND session_version = 0)
        OR
        (intent = 'link' AND user_id <> '' AND session_id <> '' AND session_version > 0)
    ),
    CHECK (
        (lifecycle_state = 'pending' AND consumed_at_unix_nano IS NULL AND length(code_verifier) >= 43)
        OR
        (lifecycle_state IN ('consumed', 'expired') AND consumed_at_unix_nano = updated_at_unix_nano AND code_verifier = '')
    )
) STRICT;

CREATE INDEX local_identity_oidc_authorization_transactions_expiry_idx
    ON local_identity_oidc_authorization_transactions(lifecycle_state, expires_at_unix_nano);
