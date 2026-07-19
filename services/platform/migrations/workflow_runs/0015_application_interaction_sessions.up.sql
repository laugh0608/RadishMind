CREATE TABLE application_interaction_sessions (
    tenant_ref text NOT NULL CHECK (btrim(tenant_ref) <> ''),
    workspace_id text NOT NULL CHECK (btrim(workspace_id) <> ''),
    application_id text NOT NULL CHECK (btrim(application_id) <> ''),
    owner_subject_ref text NOT NULL CHECK (btrim(owner_subject_ref) <> ''),
    session_id text NOT NULL CHECK (session_id ~ '^appsess_[a-z2-7]{16}$'),
    session_state text NOT NULL CHECK (session_state IN ('active','closed')),
    record_version bigint NOT NULL CHECK (record_version > 0),
    execution_profile text NOT NULL CHECK (execution_profile IN ('workflow_definition_executor_v1','application_rag_invocation_v1')),
    updated_at timestamptz NOT NULL,
    sanitized_session_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_session_payload) = 'object'
        AND sanitized_session_payload->>'schema_version' = 'application_session.v1'
        AND sanitized_session_payload->>'session_id' = session_id
        AND sanitized_session_payload->>'state' = session_state
        AND (sanitized_session_payload->>'record_version')::bigint = record_version
        AND sanitized_session_payload#>>'{profile_binding,execution_profile}' = execution_profile
        AND sanitized_session_payload->>'content_retention' = 'metadata_only'
    ),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id)
);

CREATE TABLE application_interaction_session_turns (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    session_id text NOT NULL, turn_id text NOT NULL CHECK (turn_id ~ '^appturn_[a-z2-7]{16}$'),
    turn_sequence bigint NOT NULL CHECK (turn_sequence > 0), client_turn_key text NOT NULL CHECK (btrim(client_turn_key) <> ''),
    turn_status text NOT NULL CHECK (turn_status IN ('running','succeeded','failed','canceled','outcome_unknown')),
    started_at timestamptz NOT NULL, completed_at timestamptz,
    sanitized_turn_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_turn_payload) = 'object'
        AND sanitized_turn_payload->>'schema_version' = 'application_session_turn.v1'
        AND sanitized_turn_payload->>'turn_id' = turn_id
        AND sanitized_turn_payload->>'session_id' = session_id
        AND (sanitized_turn_payload->>'sequence')::bigint = turn_sequence
        AND sanitized_turn_payload->>'client_turn_key' = client_turn_key
        AND sanitized_turn_payload->>'status' = turn_status
    ),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_sequence),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,client_turn_key),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id)
        REFERENCES application_interaction_sessions (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK ((turn_status='running' AND completed_at IS NULL) OR (turn_status<>'running' AND completed_at IS NOT NULL AND completed_at>=started_at))
);

CREATE INDEX application_interaction_sessions_scope_idx ON application_interaction_sessions
    (tenant_ref,workspace_id,application_id,owner_subject_ref,session_state,execution_profile,updated_at DESC,session_id DESC);
CREATE INDEX application_interaction_session_turns_scope_idx ON application_interaction_session_turns
    (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_sequence);

CREATE FUNCTION enforce_application_interaction_session_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id
       OR NEW.owner_subject_ref<>OLD.owner_subject_ref OR NEW.session_id<>OLD.session_id OR NEW.execution_profile<>OLD.execution_profile
       OR NEW.sanitized_session_payload->'profile_binding'<>OLD.sanitized_session_payload->'profile_binding'
       OR NEW.sanitized_session_payload#>>'{authority,authority_digest}'<>OLD.sanitized_session_payload#>>'{authority,authority_digest}'
       OR NEW.sanitized_session_payload->>'created_at'<>OLD.sanitized_session_payload->>'created_at'
       OR NEW.sanitized_session_payload->>'created_by_actor_ref'<>OLD.sanitized_session_payload->>'created_by_actor_ref'
       OR NEW.record_version<>OLD.record_version+1 OR OLD.session_state<>'active' OR NEW.session_state NOT IN ('active','closed') THEN
        RAISE EXCEPTION 'application interaction session transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER application_interaction_sessions_controlled_update BEFORE UPDATE ON application_interaction_sessions
FOR EACH ROW EXECUTE FUNCTION enforce_application_interaction_session_update();

CREATE FUNCTION enforce_application_interaction_turn_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id
       OR NEW.owner_subject_ref<>OLD.owner_subject_ref OR NEW.session_id<>OLD.session_id OR NEW.turn_id<>OLD.turn_id
       OR NEW.turn_sequence<>OLD.turn_sequence OR NEW.client_turn_key<>OLD.client_turn_key OR NEW.started_at<>OLD.started_at
       OR (NEW.sanitized_turn_payload-'status'-'run_ref'-'failure_code'-'failure_summary'-'completed_at'-'actor_ref'-'request_id'-'audit_ref')
          <>(OLD.sanitized_turn_payload-'status'-'run_ref'-'failure_code'-'failure_summary'-'completed_at'-'actor_ref'-'request_id'-'audit_ref')
       OR OLD.turn_status<>'running' OR NEW.turn_status='running' THEN
        RAISE EXCEPTION 'application interaction turn transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER application_interaction_turns_controlled_update BEFORE UPDATE ON application_interaction_session_turns
FOR EACH ROW EXECUTE FUNCTION enforce_application_interaction_turn_update();
CREATE TRIGGER application_interaction_sessions_no_delete BEFORE DELETE ON application_interaction_sessions
FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
CREATE TRIGGER application_interaction_turns_no_delete BEFORE DELETE ON application_interaction_session_turns
FOR EACH ROW EXECUTE FUNCTION reject_workflow_rag_snapshot_append_only_mutation();
