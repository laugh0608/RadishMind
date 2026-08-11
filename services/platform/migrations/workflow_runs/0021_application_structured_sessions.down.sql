ALTER TABLE application_interaction_session_turns
    DROP CONSTRAINT application_interaction_session_turns_payload_v4_check;
ALTER TABLE application_interaction_sessions
    DROP CONSTRAINT application_interaction_sessions_payload_v4_check,
    DROP CONSTRAINT application_interaction_sessions_profile_v4_check;

ALTER TABLE application_interaction_sessions
    ADD CONSTRAINT application_interaction_sessions_profile_v1_check
        CHECK (execution_profile IN ('workflow_definition_executor_v1','application_rag_invocation_v1')),
    ADD CONSTRAINT application_interaction_sessions_payload_v1_check CHECK (
        jsonb_typeof(sanitized_session_payload) = 'object'
        AND sanitized_session_payload->>'schema_version' = 'application_session.v1'
        AND sanitized_session_payload->>'session_id' = session_id
        AND sanitized_session_payload->>'state' = session_state
        AND (sanitized_session_payload->>'record_version')::bigint = record_version
        AND sanitized_session_payload#>>'{profile_binding,execution_profile}' = execution_profile
        AND sanitized_session_payload->>'content_retention' = 'metadata_only'
    );

ALTER TABLE application_interaction_session_turns
    ADD CONSTRAINT application_interaction_session_turns_payload_v1_check CHECK (
        jsonb_typeof(sanitized_turn_payload) = 'object'
        AND sanitized_turn_payload->>'schema_version' = 'application_session_turn.v1'
        AND sanitized_turn_payload->>'turn_id' = turn_id
        AND sanitized_turn_payload->>'session_id' = session_id
        AND (sanitized_turn_payload->>'sequence')::bigint = turn_sequence
        AND sanitized_turn_payload->>'client_turn_key' = client_turn_key
        AND sanitized_turn_payload->>'status' = turn_status
    );
