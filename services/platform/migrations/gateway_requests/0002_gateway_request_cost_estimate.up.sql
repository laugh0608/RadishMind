ALTER TABLE gateway_request_records
    ADD COLUMN cost_availability text NOT NULL DEFAULT 'legacy_not_captured'
    CHECK (cost_availability IN (
        'estimated', 'usage_not_reported', 'price_not_configured', 'price_unavailable',
        'not_applicable', 'legacy_not_captured'
    ));

DROP INDEX gateway_request_records_failure_idx;
CREATE INDEX gateway_request_records_failure_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, request_status, failure_boundary,
     usage_availability, cost_availability, started_at DESC, request_id DESC);
