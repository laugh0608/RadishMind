ALTER TABLE saved_workflow_drafts
    ADD CONSTRAINT saved_workflow_drafts_payload_schema_check CHECK (
        schema_version IN ('saved_workflow_draft.v1', 'saved_workflow_draft.v2')
        AND sanitized_draft_payload ->> 'schema_version' = schema_version
        AND (
            schema_version = 'saved_workflow_draft.v1'
            OR (
                schema_version = 'saved_workflow_draft.v2'
                AND sanitized_draft_payload #>> '{input_contract,contract_digest}' ~ '^sha256:[a-f0-9]{64}$'
            )
        )
    );
