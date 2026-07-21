CREATE TABLE prompt_application_runtime_assignments (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    assignment_id text NOT NULL CHECK (assignment_id ~ '^ptra_[a-z2-7]{16}$'), assignment_version bigint NOT NULL CHECK (assignment_version > 0),
    assignment_state text NOT NULL CHECK (assignment_state IN ('active','revoked')), assignment_digest text NOT NULL CHECK (assignment_digest ~ '^sha256:[a-f0-9]{64}$'),
    updated_at timestamptz NOT NULL, sanitized_assignment_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_assignment_payload) = 'object'
        AND sanitized_assignment_payload->>'schema_version' = 'prompt_application_runtime_assignment.v1'
        AND sanitized_assignment_payload->>'assignment_id' = assignment_id
        AND (sanitized_assignment_payload->>'assignment_version')::bigint = assignment_version
        AND sanitized_assignment_payload->>'state' = assignment_state
        AND sanitized_assignment_payload->>'assignment_digest' = assignment_digest),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
);
CREATE TABLE prompt_application_runtime_assignment_events (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    event_id text NOT NULL CHECK (event_id ~ '^ptrae_[a-z2-7]{16}$'), assignment_id text NOT NULL,
    event_sequence bigint NOT NULL CHECK (event_sequence > 0), resulting_assignment_version bigint NOT NULL CHECK (resulting_assignment_version > 0),
    occurred_at timestamptz NOT NULL, sanitized_event_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_event_payload) = 'object'
        AND sanitized_event_payload->>'schema_version' = 'prompt_application_runtime_assignment_event.v1'
        AND sanitized_event_payload->>'event_id' = event_id AND sanitized_event_payload->>'assignment_id' = assignment_id
        AND (sanitized_event_payload->>'event_sequence')::bigint = event_sequence
        AND (sanitized_event_payload->>'resulting_assignment_version')::bigint = resulting_assignment_version),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,event_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
        REFERENCES prompt_application_runtime_assignments (tenant_ref,workspace_id,application_id,owner_subject_ref) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);
CREATE TABLE prompt_application_sessions (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    session_id text NOT NULL CHECK (session_id ~ '^appsess_[a-z2-7]{16}$'), session_state text NOT NULL CHECK (session_state IN ('active','closed')),
    record_version bigint NOT NULL CHECK (record_version > 0), updated_at timestamptz NOT NULL, authority_digest text NOT NULL CHECK (authority_digest ~ '^sha256:[a-f0-9]{64}$'),
    sanitized_session_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_session_payload) = 'object' AND sanitized_session_payload->>'schema_version' = 'application_session.v2'
        AND sanitized_session_payload->>'session_id' = session_id AND sanitized_session_payload->>'state' = session_state
        AND (sanitized_session_payload->>'record_version')::bigint = record_version
        AND sanitized_session_payload#>>'{profile_binding,execution_profile}' = 'prompt_application_invocation_v1'
        AND sanitized_session_payload#>>'{authority,authority_digest}' = authority_digest
        AND sanitized_session_payload->>'content_retention' = 'metadata_only'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id)
);
CREATE TABLE prompt_application_session_turns (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    session_id text NOT NULL, turn_id text NOT NULL CHECK (turn_id ~ '^appturn_[a-z2-7]{16}$'), turn_sequence bigint NOT NULL CHECK (turn_sequence > 0),
    client_turn_key text NOT NULL CHECK (btrim(client_turn_key) <> ''), turn_status text NOT NULL CHECK (turn_status IN ('running','succeeded','failed','canceled','outcome_unknown')),
    started_at timestamptz NOT NULL, completed_at timestamptz, sanitized_turn_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_turn_payload) = 'object' AND sanitized_turn_payload->>'schema_version' = 'application_session_turn.v2'
        AND sanitized_turn_payload->>'turn_id' = turn_id AND sanitized_turn_payload->>'session_id' = session_id
        AND (sanitized_turn_payload->>'sequence')::bigint = turn_sequence AND sanitized_turn_payload->>'client_turn_key' = client_turn_key
        AND sanitized_turn_payload->>'status' = turn_status AND sanitized_turn_payload->>'execution_profile' = 'prompt_application_invocation_v1'),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_sequence),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,client_turn_key),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id)
        REFERENCES prompt_application_sessions (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK ((turn_status='running' AND completed_at IS NULL) OR (turn_status<>'running' AND completed_at>=started_at))
);
CREATE TABLE prompt_application_run_records (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, run_id text NOT NULL CHECK (run_id ~ '^run_[a-z0-9]{16,64}$'),
    record_version bigint NOT NULL CHECK (record_version > 0), run_status text NOT NULL CHECK (run_status IN ('running','succeeded','failed','canceled','outcome_unknown')),
    template_id text NOT NULL CHECK (template_id ~ '^ptpl_[a-z2-7]{16}$'), template_version bigint NOT NULL CHECK (template_version > 0),
    authority_digest text NOT NULL CHECK (authority_digest ~ '^sha256:[a-f0-9]{64}$'), started_at timestamptz NOT NULL, completed_at timestamptz,
    sanitized_run_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_run_payload) = 'object' AND sanitized_run_payload->>'schema_version' = 'workflow_run_record.v6'
        AND sanitized_run_payload->>'run_id' = run_id AND (sanitized_run_payload->>'record_version')::bigint = record_version
        AND sanitized_run_payload->>'status' = run_status AND sanitized_run_payload->>'execution_profile' = 'prompt_application_invocation_v1'
        AND sanitized_run_payload->>'execution_source_id' = template_id
        AND (sanitized_run_payload->>'execution_source_version')::bigint = template_version
        AND sanitized_run_payload#>>'{authority,authority_digest}' = authority_digest AND sanitized_run_payload->>'output' = ''
        AND NOT (sanitized_run_payload ?| ARRAY['variables','messages','raw_response','provider_api_key'])),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,run_id),
    CHECK ((run_status='running' AND completed_at IS NULL) OR (run_status<>'running' AND completed_at>=started_at))
);
CREATE INDEX prompt_application_assignment_events_scope_idx ON prompt_application_runtime_assignment_events (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence);
CREATE INDEX prompt_application_sessions_scope_idx ON prompt_application_sessions (tenant_ref,workspace_id,application_id,owner_subject_ref,session_state,updated_at DESC,session_id DESC);
CREATE INDEX prompt_application_turns_scope_idx ON prompt_application_session_turns (tenant_ref,workspace_id,application_id,owner_subject_ref,session_id,turn_sequence);
CREATE INDEX prompt_application_runs_history_idx ON prompt_application_run_records (tenant_ref,workspace_id,application_id,started_at DESC,run_id DESC);

CREATE FUNCTION enforce_prompt_application_assignment_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id OR NEW.owner_subject_ref<>OLD.owner_subject_ref
 OR NEW.assignment_id<>OLD.assignment_id OR NEW.assignment_version<>OLD.assignment_version+1 OR OLD.assignment_state='revoked' THEN RAISE EXCEPTION 'prompt application assignment transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER prompt_application_assignments_controlled_update BEFORE UPDATE ON prompt_application_runtime_assignments FOR EACH ROW EXECUTE FUNCTION enforce_prompt_application_assignment_update();
CREATE FUNCTION enforce_prompt_application_session_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id OR NEW.owner_subject_ref<>OLD.owner_subject_ref
 OR NEW.session_id<>OLD.session_id OR NEW.authority_digest<>OLD.authority_digest OR NEW.record_version<>OLD.record_version+1 OR OLD.session_state<>'active' THEN RAISE EXCEPTION 'prompt application session transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER prompt_application_sessions_controlled_update BEFORE UPDATE ON prompt_application_sessions FOR EACH ROW EXECUTE FUNCTION enforce_prompt_application_session_update();
CREATE FUNCTION enforce_prompt_application_turn_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id OR NEW.owner_subject_ref<>OLD.owner_subject_ref
 OR NEW.session_id<>OLD.session_id OR NEW.turn_id<>OLD.turn_id OR NEW.turn_sequence<>OLD.turn_sequence OR NEW.client_turn_key<>OLD.client_turn_key
 OR NEW.started_at<>OLD.started_at OR OLD.turn_status<>'running' OR NEW.turn_status='running' THEN RAISE EXCEPTION 'prompt application turn transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER prompt_application_turns_controlled_update BEFORE UPDATE ON prompt_application_session_turns FOR EACH ROW EXECUTE FUNCTION enforce_prompt_application_turn_update();
CREATE FUNCTION enforce_prompt_application_run_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.application_id<>OLD.application_id OR NEW.run_id<>OLD.run_id
 OR NEW.template_id<>OLD.template_id OR NEW.template_version<>OLD.template_version OR NEW.authority_digest<>OLD.authority_digest OR NEW.started_at<>OLD.started_at
 OR NEW.record_version<>OLD.record_version+1 OR OLD.run_status<>'running' OR NEW.run_status='running' THEN RAISE EXCEPTION 'prompt application run transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER prompt_application_runs_controlled_update BEFORE UPDATE ON prompt_application_run_records FOR EACH ROW EXECUTE FUNCTION enforce_prompt_application_run_update();
CREATE FUNCTION reject_prompt_application_runtime_mutation() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'prompt application runtime resource cannot be mutated'; END; $$;
CREATE TRIGGER prompt_application_assignments_no_delete BEFORE DELETE ON prompt_application_runtime_assignments FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
CREATE TRIGGER prompt_application_assignment_events_no_update BEFORE UPDATE ON prompt_application_runtime_assignment_events FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
CREATE TRIGGER prompt_application_assignment_events_no_delete BEFORE DELETE ON prompt_application_runtime_assignment_events FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
CREATE TRIGGER prompt_application_sessions_no_delete BEFORE DELETE ON prompt_application_sessions FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
CREATE TRIGGER prompt_application_turns_no_delete BEFORE DELETE ON prompt_application_session_turns FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
CREATE TRIGGER prompt_application_runs_no_delete BEFORE DELETE ON prompt_application_run_records FOR EACH ROW EXECUTE FUNCTION reject_prompt_application_runtime_mutation();
