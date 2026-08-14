DROP INDEX gateway_request_records_attempt_idx;

ALTER TABLE gateway_request_records
    DROP CONSTRAINT gateway_request_records_schema_version_check,
    DROP COLUMN terminal_profile,
    DROP COLUMN terminal_provider,
    DROP COLUMN fallback_used,
    DROP COLUMN provider_attempt_count;
