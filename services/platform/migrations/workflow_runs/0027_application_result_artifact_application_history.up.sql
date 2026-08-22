CREATE INDEX application_result_artifacts_application_history_idx
    ON application_result_artifacts (
        tenant_ref, workspace_id, application_id, owner_subject_ref,
        created_at DESC, artifact_id DESC
    );
