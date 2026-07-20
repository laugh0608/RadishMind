ALTER TABLE workflow_rag_execution_audits
    DROP CONSTRAINT workflow_rag_execution_audits_event_kind_check;

ALTER TABLE workflow_rag_execution_audits
    ADD CONSTRAINT workflow_rag_execution_audits_event_kind_check
    CHECK (event_kind IN ('snapshot_created', 'snapshot_versioned', 'snapshot_archived', 'retrieval_started', 'retrieval_succeeded', 'retrieval_failed'));
