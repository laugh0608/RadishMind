CREATE TABLE gateway_request_quota_policies (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    application_id text NOT NULL,
    policy_id text NOT NULL,
    record_version bigint NOT NULL CHECK (record_version > 0),
    request_limit bigint NOT NULL CHECK (request_limit BETWEEN 1 AND 1000000),
    sanitized_policy_record jsonb NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id),
    UNIQUE (tenant_ref, policy_id)
);

CREATE TABLE gateway_request_quota_usage (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    application_id text NOT NULL,
    period_start date NOT NULL,
    admitted_request_count bigint NOT NULL CHECK (admitted_request_count >= 0),
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id, period_start),
    FOREIGN KEY (tenant_ref, workspace_id, environment, application_id)
        REFERENCES gateway_request_quota_policies (tenant_ref, workspace_id, environment, application_id)
);

CREATE TABLE gateway_request_quota_admissions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    application_id text NOT NULL,
    request_id text NOT NULL,
    admission_id text NOT NULL,
    api_key_id text NOT NULL,
    request_route text NOT NULL,
    period_start date NOT NULL,
    policy_id text NOT NULL,
    policy_version bigint NOT NULL CHECK (policy_version > 0),
    admitted_at timestamptz NOT NULL,
    sanitized_admission_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id, request_id),
    UNIQUE (tenant_ref, admission_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, application_id)
        REFERENCES gateway_request_quota_policies (tenant_ref, workspace_id, environment, application_id)
);

CREATE INDEX gateway_request_quota_admissions_period_idx ON gateway_request_quota_admissions (
    tenant_ref, workspace_id, environment, application_id, period_start, admitted_at
);
