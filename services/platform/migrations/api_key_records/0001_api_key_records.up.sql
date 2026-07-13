CREATE TABLE api_key_records (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    api_key_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    schema_version text NOT NULL,
    display_name text NOT NULL,
    scopes text[] NOT NULL CHECK (cardinality(scopes) > 0),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    credential_digest bytea NOT NULL CHECK (octet_length(credential_digest) = 32),
    sanitized_record_payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
    last_used_at timestamptz,
    revoked_at timestamptz,
    created_by_actor_ref text NOT NULL,
    revoked_by_actor_ref text,
    request_id text NOT NULL,
    audit_ref text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, api_key_id),
    UNIQUE (api_key_id),
    UNIQUE (credential_digest)
);

CREATE INDEX api_key_records_owner_list_idx
    ON api_key_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        created_at DESC,
        api_key_id DESC
    );

CREATE INDEX api_key_records_owner_application_list_idx
    ON api_key_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        application_id,
        created_at DESC,
        api_key_id DESC
    );

CREATE INDEX api_key_records_authentication_idx
    ON api_key_records (api_key_id, lifecycle_state, expires_at);
