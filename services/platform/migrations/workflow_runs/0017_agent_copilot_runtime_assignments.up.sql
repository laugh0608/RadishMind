CREATE TABLE agent_copilot_runtime_assignments (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    assignment_id text NOT NULL CHECK (assignment_id ~ '^acra_[a-z2-7]{16}$'),
    assignment_version bigint NOT NULL CHECK (assignment_version > 0),
    assignment_state text NOT NULL CHECK (assignment_state IN ('active','revoked')),
    assignment_digest text NOT NULL CHECK (assignment_digest ~ '^sha256:[a-f0-9]{64}$'),
    updated_at timestamptz NOT NULL,
    sanitized_assignment_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_assignment_payload) = 'object'
        AND sanitized_assignment_payload->>'schema_version' = 'agent_copilot_runtime_assignment.v1'
        AND sanitized_assignment_payload->>'assignment_id' = assignment_id
        AND (sanitized_assignment_payload->>'assignment_version')::bigint = assignment_version
        AND sanitized_assignment_payload->>'state' = assignment_state
        AND sanitized_assignment_payload->>'assignment_digest' = assignment_digest
        AND NOT (sanitized_assignment_payload ?| ARRAY['profile_source','credential','token','header','endpoint','dsn','system_prompt','provider'])),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
);

CREATE TABLE agent_copilot_runtime_assignment_events (
    tenant_ref text NOT NULL, workspace_id text NOT NULL, application_id text NOT NULL, owner_subject_ref text NOT NULL,
    event_id text NOT NULL CHECK (event_id ~ '^acrae_[a-z2-7]{16}$'), assignment_id text NOT NULL,
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    resulting_assignment_version bigint NOT NULL CHECK (resulting_assignment_version > 0),
    occurred_at timestamptz NOT NULL,
    sanitized_event_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_event_payload) = 'object'
        AND sanitized_event_payload->>'schema_version' = 'agent_copilot_runtime_assignment_event.v1'
        AND sanitized_event_payload->>'event_id' = event_id
        AND sanitized_event_payload->>'assignment_id' = assignment_id
        AND (sanitized_event_payload->>'event_sequence')::bigint = event_sequence
        AND (sanitized_event_payload->>'resulting_assignment_version')::bigint = resulting_assignment_version
        AND NOT (sanitized_event_payload ?| ARRAY['profile_source','credential','token','header','endpoint','dsn','system_prompt','provider'])),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,event_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
        REFERENCES agent_copilot_runtime_assignments (tenant_ref,workspace_id,application_id,owner_subject_ref)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX agent_copilot_assignment_events_scope_idx
    ON agent_copilot_runtime_assignment_events
    (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence);

CREATE FUNCTION enforce_agent_copilot_assignment_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id
 OR NEW.application_id<>OLD.application_id OR NEW.owner_subject_ref<>OLD.owner_subject_ref
 OR NEW.assignment_id<>OLD.assignment_id OR NEW.assignment_version<>OLD.assignment_version+1
 OR OLD.assignment_state='revoked' THEN
    RAISE EXCEPTION 'agent copilot assignment transition is invalid';
END IF;
RETURN NEW;
END; $$;

CREATE FUNCTION reject_agent_copilot_assignment_mutation() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
RAISE EXCEPTION 'agent copilot assignment resource cannot be mutated';
END; $$;

CREATE TRIGGER agent_copilot_assignments_controlled_update
    BEFORE UPDATE ON agent_copilot_runtime_assignments
    FOR EACH ROW EXECUTE FUNCTION enforce_agent_copilot_assignment_update();
CREATE TRIGGER agent_copilot_assignments_no_delete
    BEFORE DELETE ON agent_copilot_runtime_assignments
    FOR EACH ROW EXECUTE FUNCTION reject_agent_copilot_assignment_mutation();
CREATE TRIGGER agent_copilot_assignment_events_no_update
    BEFORE UPDATE ON agent_copilot_runtime_assignment_events
    FOR EACH ROW EXECUTE FUNCTION reject_agent_copilot_assignment_mutation();
CREATE TRIGGER agent_copilot_assignment_events_no_delete
    BEFORE DELETE ON agent_copilot_runtime_assignment_events
    FOR EACH ROW EXECUTE FUNCTION reject_agent_copilot_assignment_mutation();
