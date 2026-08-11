DO $$
DECLARE constraint_name text;
BEGIN
    FOR constraint_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'application_interaction_sessions'::regclass
          AND contype = 'c'
          AND (
              pg_get_constraintdef(oid) LIKE '%workflow_definition_executor_v1%'
              OR pg_get_constraintdef(oid) LIKE '%application_session.v1%'
          )
    LOOP
        EXECUTE format('ALTER TABLE application_interaction_sessions DROP CONSTRAINT %I', constraint_name);
    END LOOP;
    FOR constraint_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'application_interaction_session_turns'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%application_session_turn.v1%'
    LOOP
        EXECUTE format('ALTER TABLE application_interaction_session_turns DROP CONSTRAINT %I', constraint_name);
    END LOOP;
END $$;

ALTER TABLE application_interaction_sessions
    ADD CONSTRAINT application_interaction_sessions_profile_v4_check
        CHECK (execution_profile IN ('workflow_definition_executor_v1','workflow_definition_executor_v2','application_rag_invocation_v1')),
    ADD CONSTRAINT application_interaction_sessions_payload_v4_check CHECK (
        jsonb_typeof(sanitized_session_payload) = 'object'
        AND sanitized_session_payload->>'schema_version' = CASE execution_profile
            WHEN 'workflow_definition_executor_v2' THEN 'application_session.v4'
            ELSE 'application_session.v1'
        END
        AND sanitized_session_payload->>'session_id' = session_id
        AND sanitized_session_payload->>'state' = session_state
        AND (sanitized_session_payload->>'record_version')::bigint = record_version
        AND sanitized_session_payload#>>'{profile_binding,execution_profile}' = execution_profile
        AND sanitized_session_payload#>>'{authority,schema_version}' = CASE execution_profile
            WHEN 'workflow_definition_executor_v2' THEN 'application_runtime_authority.v4'
            ELSE 'application_runtime_authority.v1'
        END
        AND sanitized_session_payload->>'content_retention' = 'metadata_only'
        AND (
            execution_profile <> 'workflow_definition_executor_v2'
            OR (
                jsonb_typeof(sanitized_session_payload#>'{authority,workflow_definition,input_contract}') = 'object'
                AND btrim(sanitized_session_payload#>>'{authority,workflow_definition,input_contract,contract_id}') <> ''
                AND sanitized_session_payload#>>'{authority,workflow_definition,input_contract,contract_digest}' ~ '^sha256:[a-f0-9]{64}$'
                AND jsonb_typeof(sanitized_session_payload#>'{authority,workflow_definition,input_contract,fields}') = 'array'
            )
        )
    );

ALTER TABLE application_interaction_session_turns
    ADD CONSTRAINT application_interaction_session_turns_payload_v4_check CHECK (
        jsonb_typeof(sanitized_turn_payload) = 'object'
        AND sanitized_turn_payload->>'schema_version' = CASE sanitized_turn_payload->>'execution_profile'
            WHEN 'workflow_definition_executor_v2' THEN 'application_session_turn.v4'
            ELSE 'application_session_turn.v1'
        END
        AND sanitized_turn_payload->>'execution_profile' IN ('workflow_definition_executor_v1','workflow_definition_executor_v2','application_rag_invocation_v1')
        AND sanitized_turn_payload->>'turn_id' = turn_id
        AND sanitized_turn_payload->>'session_id' = session_id
        AND (sanitized_turn_payload->>'sequence')::bigint = turn_sequence
        AND sanitized_turn_payload->>'client_turn_key' = client_turn_key
        AND sanitized_turn_payload->>'status' = turn_status
        AND NOT (sanitized_turn_payload ?| ARRAY['input_text','inputs'])
        AND (
            sanitized_turn_payload->>'execution_profile' <> 'workflow_definition_executor_v2'
            OR (
                btrim(sanitized_turn_payload->>'input_contract_id') <> ''
                AND sanitized_turn_payload->>'input_contract_digest' ~ '^sha256:[a-f0-9]{64}$'
                AND jsonb_typeof(sanitized_turn_payload->'input_fields') = 'array'
                AND sanitized_turn_payload#>>'{authority,schema_version}' = 'application_runtime_authority.v4'
            )
        )
    );
