ALTER TABLE gateway_request_records
    ADD COLUMN provider_attempt_count smallint NOT NULL DEFAULT 0
    CHECK (provider_attempt_count BETWEEN 0 AND 2),
    ADD COLUMN fallback_used boolean NOT NULL DEFAULT false,
    ADD COLUMN terminal_provider text NOT NULL DEFAULT '',
    ADD COLUMN terminal_profile text NOT NULL DEFAULT '';

ALTER TABLE gateway_request_records
    ADD CONSTRAINT gateway_request_records_schema_version_check
    CHECK (schema_version IN (
        'gateway_request_record.v1',
        'gateway_request_record.v2',
        'gateway_request_record.v3'
    ));

CREATE INDEX gateway_request_records_attempt_idx ON gateway_request_records (
    tenant_ref, workspace_id, consumer_ref, application_id,
    fallback_used, terminal_provider, terminal_profile, started_at DESC, request_id DESC
);
