CREATE TABLE application_result_artifacts (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    owner_subject_ref TEXT NOT NULL CHECK (length(trim(owner_subject_ref)) > 0),
    artifact_id TEXT NOT NULL CHECK (
        length(artifact_id) = 23 AND artifact_id GLOB 'appres_[a-z2-7]*'
        AND artifact_id NOT GLOB '*[^a-z2-7_]*'
    ),
    session_id TEXT NOT NULL CHECK (
        length(session_id) = 24 AND session_id GLOB 'appsess_[a-z2-7]*'
        AND session_id NOT GLOB '*[^a-z2-7_]*'
    ),
    turn_id TEXT NOT NULL CHECK (
        length(turn_id) = 24 AND turn_id GLOB 'appturn_[a-z2-7]*'
        AND turn_id NOT GLOB '*[^a-z2-7_]*'
    ),
    client_turn_key TEXT NOT NULL CHECK (length(trim(client_turn_key)) > 0),
    execution_profile TEXT NOT NULL CHECK (execution_profile IN (
        'workflow_definition_executor_v1',
        'workflow_definition_executor_v2',
        'application_rag_invocation_v1',
        'prompt_application_invocation_v1',
        'agent_copilot_suggestion_v1'
    )),
    run_id TEXT NOT NULL CHECK (
        length(run_id) BETWEEN 20 AND 68 AND run_id GLOB 'run_[a-z0-9]*'
        AND run_id NOT GLOB '*[^a-z0-9_]*'
    ),
    run_schema_version TEXT NOT NULL CHECK (
        (execution_profile = 'workflow_definition_executor_v1' AND run_schema_version = 'workflow_run_record.v5')
        OR (execution_profile = 'workflow_definition_executor_v2' AND run_schema_version = 'workflow_run_record.v8')
        OR (execution_profile = 'application_rag_invocation_v1' AND run_schema_version = 'workflow_run_record.v4')
        OR (execution_profile = 'prompt_application_invocation_v1' AND run_schema_version = 'workflow_run_record.v6')
        OR (execution_profile = 'agent_copilot_suggestion_v1' AND run_schema_version = 'workflow_run_record.v7')
    ),
    content_type TEXT NOT NULL CHECK (content_type IN ('text/markdown', 'application/json')),
    content_bytes INTEGER NOT NULL CHECK (content_bytes > 0 AND content_bytes <= 65536),
    content_digest TEXT NOT NULL CHECK (
        length(content_digest) = 71 AND substr(content_digest, 1, 7) = 'sha256:'
        AND substr(content_digest, 8) NOT GLOB '*[^0-9a-f]*'
    ),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    artifact_payload TEXT NOT NULL CHECK (
        json_valid(artifact_payload) AND json_type(artifact_payload) = 'object'
        AND json_extract(artifact_payload, '$.schema_version') = 'application_result_artifact.v1'
        AND json_extract(artifact_payload, '$.artifact_id') = artifact_id
        AND json_extract(artifact_payload, '$.record_version') = 1
        AND json_extract(artifact_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(artifact_payload, '$.workspace_id') = workspace_id
        AND json_extract(artifact_payload, '$.application_id') = application_id
        AND json_extract(artifact_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(artifact_payload, '$.session_id') = session_id
        AND json_extract(artifact_payload, '$.turn_id') = turn_id
        AND json_extract(artifact_payload, '$.client_turn_key') = client_turn_key
        AND json_extract(artifact_payload, '$.execution_profile') = execution_profile
        AND json_extract(artifact_payload, '$.run_ref.run_id') = run_id
        AND json_extract(artifact_payload, '$.run_ref.schema_version') = run_schema_version
        AND json_extract(artifact_payload, '$.content_type') = content_type
        AND length(CAST(json_extract(artifact_payload, '$.content') AS BLOB)) = content_bytes
        AND json_extract(artifact_payload, '$.content_bytes') = content_bytes
        AND json_extract(artifact_payload, '$.content_digest') = content_digest
        AND length(trim(json_extract(artifact_payload, '$.created_by_actor_ref'))) > 0
        AND length(trim(json_extract(artifact_payload, '$.request_id'))) > 0
        AND length(trim(json_extract(artifact_payload, '$.audit_ref'))) > 0
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_id)
) STRICT;

CREATE INDEX application_result_artifacts_session_history_idx
    ON application_result_artifacts (
        tenant_ref, workspace_id, application_id, owner_subject_ref,
        session_id, created_at_unix_nano DESC, artifact_id DESC
    );

CREATE TRIGGER application_result_artifacts_no_update
BEFORE UPDATE ON application_result_artifacts
BEGIN
    SELECT RAISE(ABORT, 'application result artifact is immutable');
END;

CREATE TRIGGER application_result_artifacts_no_delete
BEFORE DELETE ON application_result_artifacts
BEGIN
    SELECT RAISE(ABORT, 'application result artifact is immutable');
END;
