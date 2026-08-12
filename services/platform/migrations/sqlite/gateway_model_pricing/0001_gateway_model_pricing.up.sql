CREATE TABLE gateway_model_pricing_revisions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    provider_id TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    model_id TEXT NOT NULL,
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    policy_id TEXT NOT NULL,
    policy_digest TEXT NOT NULL,
    sanitized_policy_record TEXT NOT NULL CHECK (
        json_valid(sanitized_policy_record) AND json_type(sanitized_policy_record) = 'object'
    ),
    updated_at_unix_nano INTEGER NOT NULL,
    PRIMARY KEY (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, record_version
    ),
    UNIQUE (tenant_ref, policy_id, record_version)
) STRICT;

CREATE TABLE gateway_model_pricing_current (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    provider_id TEXT NOT NULL,
    profile_id TEXT NOT NULL,
    model_id TEXT NOT NULL,
    policy_id TEXT NOT NULL,
    current_version INTEGER NOT NULL CHECK (current_version > 0),
    updated_at_unix_nano INTEGER NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, provider_id, profile_id, model_id),
    UNIQUE (tenant_ref, policy_id),
    FOREIGN KEY (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, current_version
    ) REFERENCES gateway_model_pricing_revisions (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, record_version
    )
) STRICT;

CREATE INDEX gateway_model_pricing_current_selection_idx ON gateway_model_pricing_current (
    tenant_ref, workspace_id, environment, provider_id, profile_id, model_id
);

CREATE TRIGGER gateway_model_pricing_revisions_no_update
BEFORE UPDATE ON gateway_model_pricing_revisions
BEGIN
    SELECT RAISE(ABORT, 'gateway model pricing revisions are append-only');
END;

CREATE TRIGGER gateway_model_pricing_revisions_no_delete
BEFORE DELETE ON gateway_model_pricing_revisions
BEGIN
    SELECT RAISE(ABORT, 'gateway model pricing revisions are append-only');
END;
