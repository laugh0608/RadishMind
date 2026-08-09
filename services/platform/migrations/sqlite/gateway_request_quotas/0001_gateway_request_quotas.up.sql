CREATE TABLE gateway_request_quota_policies (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    application_id TEXT NOT NULL,
    policy_id TEXT NOT NULL,
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    request_limit INTEGER NOT NULL CHECK (request_limit BETWEEN 1 AND 1000000),
    sanitized_policy_record TEXT NOT NULL CHECK (
        json_valid(sanitized_policy_record) AND json_type(sanitized_policy_record) = 'object'
    ),
    updated_at_unix_nano INTEGER NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id),
    UNIQUE (tenant_ref, policy_id)
) STRICT;

CREATE TABLE gateway_request_quota_usage (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    application_id TEXT NOT NULL,
    period_start TEXT NOT NULL,
    admitted_request_count INTEGER NOT NULL CHECK (admitted_request_count >= 0),
    updated_at_unix_nano INTEGER NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id, period_start),
    FOREIGN KEY (tenant_ref, workspace_id, environment, application_id)
        REFERENCES gateway_request_quota_policies (tenant_ref, workspace_id, environment, application_id)
) STRICT;

CREATE TABLE gateway_request_quota_admissions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development', 'test')),
    application_id TEXT NOT NULL,
    request_id TEXT NOT NULL,
    admission_id TEXT NOT NULL,
    api_key_id TEXT NOT NULL,
    request_route TEXT NOT NULL,
    period_start TEXT NOT NULL,
    policy_id TEXT NOT NULL,
    policy_version INTEGER NOT NULL CHECK (policy_version > 0),
    admitted_at_unix_nano INTEGER NOT NULL,
    sanitized_admission_record TEXT NOT NULL CHECK (
        json_valid(sanitized_admission_record) AND json_type(sanitized_admission_record) = 'object'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, environment, application_id, request_id),
    UNIQUE (tenant_ref, admission_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, application_id)
        REFERENCES gateway_request_quota_policies (tenant_ref, workspace_id, environment, application_id)
) STRICT;

CREATE INDEX gateway_request_quota_admissions_period_idx ON gateway_request_quota_admissions (
    tenant_ref, workspace_id, environment, application_id, period_start, admitted_at_unix_nano
);
