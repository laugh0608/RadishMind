CREATE TABLE gateway_model_pricing_revisions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    provider_id text NOT NULL,
    profile_id text NOT NULL,
    model_id text NOT NULL,
    record_version bigint NOT NULL CHECK (record_version > 0),
    policy_id text NOT NULL,
    policy_digest text NOT NULL,
    sanitized_policy_record jsonb NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, record_version
    ),
    UNIQUE (tenant_ref, policy_id, record_version)
);

CREATE TABLE gateway_model_pricing_current (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    provider_id text NOT NULL,
    profile_id text NOT NULL,
    model_id text NOT NULL,
    policy_id text NOT NULL,
    current_version bigint NOT NULL CHECK (current_version > 0),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, provider_id, profile_id, model_id),
    UNIQUE (tenant_ref, policy_id),
    FOREIGN KEY (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, current_version
    ) REFERENCES gateway_model_pricing_revisions (
        tenant_ref, workspace_id, environment, provider_id, profile_id, model_id, record_version
    )
);

CREATE INDEX gateway_model_pricing_current_selection_idx ON gateway_model_pricing_current (
    tenant_ref, workspace_id, environment, provider_id, profile_id, model_id
);

CREATE FUNCTION gateway_model_pricing_revisions_append_only() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'gateway model pricing revisions are append-only';
END;
$$;

CREATE TRIGGER gateway_model_pricing_revisions_no_update
BEFORE UPDATE OR DELETE ON gateway_model_pricing_revisions
FOR EACH ROW EXECUTE FUNCTION gateway_model_pricing_revisions_append_only();
