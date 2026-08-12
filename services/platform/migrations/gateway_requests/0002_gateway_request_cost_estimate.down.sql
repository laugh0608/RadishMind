DROP INDEX gateway_request_records_failure_idx;
CREATE INDEX gateway_request_records_failure_idx ON gateway_request_records
    (tenant_ref, workspace_id, consumer_ref, application_id, request_status, failure_boundary,
     usage_availability, started_at DESC, request_id DESC);

ALTER TABLE gateway_request_records DROP COLUMN cost_availability;
