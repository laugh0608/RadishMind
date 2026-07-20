ALTER TABLE workflow_rag_execution_audits
    DROP CONSTRAINT IF EXISTS workflow_rag_execution_audits_event_kind_check;

DROP TRIGGER IF EXISTS workflow_rag_execution_audits_append_only
    ON workflow_rag_execution_audits;

DELETE FROM workflow_rag_execution_audits
WHERE event_kind IN ('retrieval_started', 'retrieval_succeeded', 'retrieval_failed');

CREATE TRIGGER workflow_rag_execution_audits_append_only
    BEFORE UPDATE OR DELETE ON workflow_rag_execution_audits
    FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();

ALTER TABLE workflow_rag_execution_audits
    ADD CONSTRAINT workflow_rag_execution_audits_event_kind_check
    CHECK (event_kind IN ('snapshot_created', 'snapshot_versioned', 'snapshot_archived'));
