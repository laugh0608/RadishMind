CREATE TABLE gateway_request_records (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    consumer_ref text NOT NULL,
    application_id text NOT NULL DEFAULT '',
    request_id text NOT NULL,
    record_version bigint NOT NULL CHECK (record_version > 0),
    schema_version text NOT NULL,
    store_mode text NOT NULL CHECK (store_mode = 'postgres_dev_test'),
    request_route text NOT NULL,
    protocol text NOT NULL,
    request_status text NOT NULL CHECK (request_status IN ('started', 'succeeded', 'failed', 'canceled')),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    selected_provider text NOT NULL DEFAULT '',
    selected_profile text NOT NULL DEFAULT '',
    selected_model text NOT NULL DEFAULT '',
    failure_boundary text NOT NULL DEFAULT '',
    usage_availability text NOT NULL CHECK (usage_availability IN ('reported', 'not_reported', 'not_applicable')),
    sanitized_request_record jsonb NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, consumer_ref, application_id, request_id)
);

CREATE INDEX gateway_request_records_history_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, started_at DESC, request_id DESC);
CREATE INDEX gateway_request_records_route_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, request_route, protocol, started_at DESC, request_id DESC);
CREATE INDEX gateway_request_records_selection_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, selected_provider, selected_profile, selected_model, started_at DESC, request_id DESC);
CREATE INDEX gateway_request_records_failure_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, request_status, failure_boundary, usage_availability, started_at DESC, request_id DESC);
CREATE INDEX gateway_request_records_retention_idx ON gateway_request_records
    (completed_at, started_at);
