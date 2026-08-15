CREATE TABLE application_result_artifacts (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    artifact_id text NOT NULL CHECK (artifact_id ~ '^appres_[a-z2-7]{16}$'),
    session_id text NOT NULL CHECK (session_id ~ '^appsess_[a-z2-7]{16}$'),
    turn_id text NOT NULL CHECK (turn_id ~ '^appturn_[a-z2-7]{16}$'),
    client_turn_key text NOT NULL CHECK (btrim(client_turn_key) <> ''),
    execution_profile text NOT NULL CHECK (execution_profile IN (
        'workflow_definition_executor_v1',
        'workflow_definition_executor_v2',
        'application_rag_invocation_v1',
        'prompt_application_invocation_v1',
        'agent_copilot_suggestion_v1'
    )),
    run_id text NOT NULL CHECK (run_id ~ '^run_[a-z0-9]{16,64}$'),
    run_schema_version text NOT NULL CHECK (
        (execution_profile = 'workflow_definition_executor_v1' AND run_schema_version = 'workflow_run_record.v5')
        OR (execution_profile = 'workflow_definition_executor_v2' AND run_schema_version = 'workflow_run_record.v8')
        OR (execution_profile = 'application_rag_invocation_v1' AND run_schema_version = 'workflow_run_record.v4')
        OR (execution_profile = 'prompt_application_invocation_v1' AND run_schema_version = 'workflow_run_record.v6')
        OR (execution_profile = 'agent_copilot_suggestion_v1' AND run_schema_version = 'workflow_run_record.v7')
    ),
    content_type text NOT NULL CHECK (content_type IN ('text/markdown', 'application/json')),
    content_bytes integer NOT NULL CHECK (content_bytes > 0 AND content_bytes <= 65536),
    content_digest text NOT NULL CHECK (content_digest ~ '^sha256:[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL,
    artifact_payload jsonb NOT NULL,
    CONSTRAINT application_result_artifacts_payload_identity_check CHECK (
        artifact_payload->>'schema_version' = 'application_result_artifact.v1'
        AND artifact_payload->>'artifact_id' = artifact_id
        AND (artifact_payload->>'record_version')::integer = 1
        AND artifact_payload->>'tenant_ref' = tenant_ref
        AND artifact_payload->>'workspace_id' = workspace_id
        AND artifact_payload->>'application_id' = application_id
        AND artifact_payload->>'owner_subject_ref' = owner_subject_ref
        AND artifact_payload->>'session_id' = session_id
        AND artifact_payload->>'turn_id' = turn_id
    ),
    CONSTRAINT application_result_artifacts_payload_provenance_check CHECK (
        artifact_payload->>'client_turn_key' = client_turn_key
        AND artifact_payload->>'execution_profile' = execution_profile
        AND artifact_payload#>>'{run_ref,run_id}' = run_id
        AND artifact_payload#>>'{run_ref,schema_version}' = run_schema_version
    ),
    CONSTRAINT application_result_artifacts_payload_content_check CHECK (
        artifact_payload->>'content_type' = content_type
        AND octet_length(artifact_payload->>'content') = content_bytes
        AND (artifact_payload->>'content_bytes')::integer = content_bytes
        AND artifact_payload->>'content_digest' = content_digest
    ),
    CONSTRAINT application_result_artifacts_payload_audit_check CHECK (
        (artifact_payload->>'created_at')::timestamptz = created_at
        AND btrim(artifact_payload->>'created_by_actor_ref') <> ''
        AND btrim(artifact_payload->>'request_id') <> ''
        AND btrim(artifact_payload->>'audit_ref') <> ''
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_id)
);

CREATE INDEX application_result_artifacts_session_history_idx
    ON application_result_artifacts (
        tenant_ref, workspace_id, application_id, owner_subject_ref,
        session_id, created_at DESC, artifact_id DESC
    );

CREATE FUNCTION reject_application_result_artifact_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'application result artifact is immutable';
END;
$$;

CREATE TRIGGER application_result_artifacts_append_only
    BEFORE UPDATE OR DELETE ON application_result_artifacts
    FOR EACH ROW EXECUTE FUNCTION reject_application_result_artifact_mutation();
