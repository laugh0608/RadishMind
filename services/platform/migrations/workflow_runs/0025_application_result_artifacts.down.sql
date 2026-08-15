DROP TRIGGER IF EXISTS application_result_artifacts_append_only
    ON application_result_artifacts;
DROP FUNCTION IF EXISTS reject_application_result_artifact_mutation();
DROP TABLE IF EXISTS application_result_artifacts;
