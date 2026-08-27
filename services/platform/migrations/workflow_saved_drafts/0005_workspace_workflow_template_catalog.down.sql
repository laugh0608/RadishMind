DROP TABLE IF EXISTS workflow_template_listing_events;
DROP TABLE IF EXISTS workflow_template_audits;
DROP TABLE IF EXISTS workflow_template_lineages;
DROP TABLE IF EXISTS workflow_template_versions;
DROP TABLE IF EXISTS workflow_template_decisions;
DROP TABLE IF EXISTS workflow_template_candidates;
DROP FUNCTION IF EXISTS reject_workflow_template_history_mutation();

ALTER TABLE saved_workflow_drafts
    DROP CONSTRAINT IF EXISTS saved_workflow_drafts_provenance_kind_check;
ALTER TABLE saved_workflow_drafts
    ADD CONSTRAINT saved_workflow_drafts_provenance_kind_check
        CHECK (
            provenance_kind IN (
                'unversioned',
                'workflow_definition',
                'saved_draft_derivation'
            )
        ) NOT VALID;
