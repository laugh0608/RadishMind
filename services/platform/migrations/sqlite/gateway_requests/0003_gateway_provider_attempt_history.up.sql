ALTER TABLE gateway_request_records RENAME TO gateway_request_records_v2;

CREATE TABLE gateway_request_records (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    consumer_ref TEXT NOT NULL,
    application_id TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL,
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    schema_version TEXT NOT NULL CHECK (schema_version IN (
        'gateway_request_record.v1', 'gateway_request_record.v2', 'gateway_request_record.v3'
    )),
    store_mode TEXT NOT NULL CHECK (store_mode = 'sqlite_dev'),
    request_route TEXT NOT NULL,
    protocol TEXT NOT NULL CHECK (protocol IN ('openai-models', 'openai-chat-completions', 'openai-responses', 'anthropic-messages')),
    request_status TEXT NOT NULL CHECK (request_status IN ('started', 'succeeded', 'failed', 'canceled')),
    started_at_unix_nano INTEGER NOT NULL,
    completed_at_unix_nano INTEGER,
    selected_provider TEXT NOT NULL DEFAULT '',
    selected_profile TEXT NOT NULL DEFAULT '',
    selected_model TEXT NOT NULL DEFAULT '',
    failure_boundary TEXT NOT NULL DEFAULT '',
    usage_availability TEXT NOT NULL CHECK (usage_availability IN ('reported', 'not_reported', 'not_applicable')),
    cost_availability TEXT NOT NULL CHECK (cost_availability IN (
        'estimated', 'usage_not_reported', 'price_not_configured', 'price_unavailable',
        'not_applicable', 'legacy_not_captured'
    )),
    provider_attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (provider_attempt_count BETWEEN 0 AND 2),
    fallback_used INTEGER NOT NULL DEFAULT 0 CHECK (fallback_used IN (0, 1)),
    terminal_provider TEXT NOT NULL DEFAULT '',
    terminal_profile TEXT NOT NULL DEFAULT '',
    sanitized_request_record TEXT NOT NULL CHECK (
        json_valid(sanitized_request_record)
        AND json_type(sanitized_request_record) = 'object'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, consumer_ref, application_id, request_id),
    CHECK (
        (request_status = 'started' AND completed_at_unix_nano IS NULL AND failure_boundary = '')
        OR
        (request_status = 'succeeded' AND completed_at_unix_nano >= started_at_unix_nano AND failure_boundary = '')
        OR
        (request_status IN ('failed', 'canceled') AND completed_at_unix_nano >= started_at_unix_nano AND failure_boundary <> '')
    )
) STRICT;

INSERT INTO gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id, request_id, record_version, schema_version,
    store_mode, request_route, protocol, request_status, started_at_unix_nano, completed_at_unix_nano,
    selected_provider, selected_profile, selected_model, failure_boundary, usage_availability,
    cost_availability, provider_attempt_count, fallback_used, terminal_provider, terminal_profile,
    sanitized_request_record
)
SELECT
    tenant_ref, workspace_id, consumer_ref, application_id, request_id, record_version, schema_version,
    store_mode, request_route, protocol, request_status, started_at_unix_nano, completed_at_unix_nano,
    selected_provider, selected_profile, selected_model, failure_boundary, usage_availability,
    cost_availability, 0, 0, '', '', sanitized_request_record
FROM gateway_request_records_v2;

DROP TABLE gateway_request_records_v2;

CREATE INDEX gateway_request_records_history_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id, started_at_unix_nano DESC, request_id DESC
);
CREATE INDEX gateway_request_records_route_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id,
    request_route, protocol, started_at_unix_nano DESC, request_id DESC
);
CREATE INDEX gateway_request_records_selection_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id,
    selected_provider, selected_profile, selected_model, started_at_unix_nano DESC, request_id DESC
);
CREATE INDEX gateway_request_records_failure_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id,
    request_status, failure_boundary, usage_availability, cost_availability, started_at_unix_nano DESC, request_id DESC
);
CREATE INDEX gateway_request_records_attempt_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id,
    fallback_used, terminal_provider, terminal_profile, started_at_unix_nano DESC, request_id DESC
);
CREATE INDEX gateway_request_records_retention_idx ON gateway_request_records (
    completed_at_unix_nano, started_at_unix_nano
);
