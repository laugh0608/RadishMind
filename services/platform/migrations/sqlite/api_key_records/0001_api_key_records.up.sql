CREATE TABLE api_key_records (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    api_key_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    schema_version TEXT NOT NULL,
    display_name TEXT NOT NULL,
    scopes_json TEXT NOT NULL CHECK (
        json_valid(scopes_json)
        AND json_type(scopes_json) = 'array'
        AND json_array_length(scopes_json) > 0
    ),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'revoked')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    credential_digest BLOB NOT NULL CHECK (length(credential_digest) = 32),
    sanitized_record_payload TEXT NOT NULL CHECK (json_valid(sanitized_record_payload)),
    created_at_unix_nano INTEGER NOT NULL,
    expires_at_unix_nano INTEGER NOT NULL CHECK (expires_at_unix_nano > created_at_unix_nano),
    last_used_at_unix_nano INTEGER CHECK (
        last_used_at_unix_nano IS NULL OR last_used_at_unix_nano >= created_at_unix_nano
    ),
    revoked_at_unix_nano INTEGER CHECK (
        revoked_at_unix_nano IS NULL OR revoked_at_unix_nano >= created_at_unix_nano
    ),
    created_by_actor_ref TEXT NOT NULL,
    revoked_by_actor_ref TEXT,
    request_id TEXT NOT NULL,
    audit_ref TEXT NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, api_key_id),
    UNIQUE (api_key_id),
    UNIQUE (credential_digest),
    CHECK (
        (lifecycle_state = 'active' AND revoked_at_unix_nano IS NULL AND revoked_by_actor_ref IS NULL)
        OR
        (lifecycle_state = 'revoked' AND revoked_at_unix_nano IS NOT NULL AND revoked_by_actor_ref IS NOT NULL)
    )
) STRICT;

CREATE INDEX api_key_records_owner_list_idx
    ON api_key_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        created_at_unix_nano DESC,
        api_key_id DESC
    );

CREATE INDEX api_key_records_owner_application_list_idx
    ON api_key_records (
        tenant_ref,
        workspace_id,
        owner_subject_ref,
        application_id,
        created_at_unix_nano DESC,
        api_key_id DESC
    );

CREATE INDEX api_key_records_authentication_idx
    ON api_key_records (api_key_id, lifecycle_state, expires_at_unix_nano);
