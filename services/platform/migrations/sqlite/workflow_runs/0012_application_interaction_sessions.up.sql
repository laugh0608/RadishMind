CREATE TABLE application_interaction_sessions (
    tenant_ref TEXT NOT NULL CHECK (length(trim(tenant_ref)) > 0),
    workspace_id TEXT NOT NULL CHECK (length(trim(workspace_id)) > 0),
    application_id TEXT NOT NULL CHECK (length(trim(application_id)) > 0),
    owner_subject_ref TEXT NOT NULL CHECK (length(trim(owner_subject_ref)) > 0),
    session_id TEXT NOT NULL CHECK (session_id GLOB 'appsess_[a-z2-7]*' AND length(session_id) = 24),
    session_state TEXT NOT NULL CHECK (session_state IN ('active', 'closed')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    execution_profile TEXT NOT NULL CHECK (execution_profile IN ('workflow_definition_executor_v1', 'application_rag_invocation_v1')),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_session_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_session_payload)
        AND json_type(sanitized_session_payload) = 'object'
        AND json_extract(sanitized_session_payload, '$.schema_version') = 'application_session.v1'
        AND json_extract(sanitized_session_payload, '$.session_id') = session_id
        AND json_extract(sanitized_session_payload, '$.state') = session_state
        AND json_extract(sanitized_session_payload, '$.record_version') = record_version
        AND json_extract(sanitized_session_payload, '$.profile_binding.execution_profile') = execution_profile
        AND json_extract(sanitized_session_payload, '$.content_retention') = 'metadata_only'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
) STRICT;

CREATE TABLE application_interaction_session_turns (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    session_id TEXT NOT NULL,
    turn_id TEXT NOT NULL CHECK (turn_id GLOB 'appturn_[a-z2-7]*' AND length(turn_id) = 24),
    turn_sequence INTEGER NOT NULL CHECK (turn_sequence > 0),
    client_turn_key TEXT NOT NULL CHECK (length(trim(client_turn_key)) > 0),
    turn_status TEXT NOT NULL CHECK (turn_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown')),
    started_at_unix_nano INTEGER NOT NULL CHECK (started_at_unix_nano > 0),
    completed_at_unix_nano INTEGER CHECK (completed_at_unix_nano IS NULL OR completed_at_unix_nano >= started_at_unix_nano),
    sanitized_turn_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_turn_payload)
        AND json_type(sanitized_turn_payload) = 'object'
        AND json_extract(sanitized_turn_payload, '$.schema_version') = 'application_session_turn.v1'
        AND json_extract(sanitized_turn_payload, '$.turn_id') = turn_id
        AND json_extract(sanitized_turn_payload, '$.session_id') = session_id
        AND json_extract(sanitized_turn_payload, '$.sequence') = turn_sequence
        AND json_extract(sanitized_turn_payload, '$.client_turn_key') = client_turn_key
        AND json_extract(sanitized_turn_payload, '$.status') = turn_status
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, client_turn_key),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        REFERENCES application_interaction_sessions (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK ((turn_status = 'running' AND completed_at_unix_nano IS NULL) OR (turn_status <> 'running' AND completed_at_unix_nano IS NOT NULL))
) STRICT;

CREATE INDEX application_interaction_sessions_scope_idx ON application_interaction_sessions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_state, execution_profile, updated_at_unix_nano DESC, session_id DESC);
CREATE INDEX application_interaction_session_turns_scope_idx ON application_interaction_session_turns
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence);

CREATE TRIGGER application_interaction_sessions_controlled_update
BEFORE UPDATE ON application_interaction_sessions
WHEN NEW.tenant_ref <> OLD.tenant_ref
  OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.application_id <> OLD.application_id
  OR NEW.owner_subject_ref <> OLD.owner_subject_ref
  OR NEW.session_id <> OLD.session_id
  OR NEW.execution_profile <> OLD.execution_profile
  OR json_extract(NEW.sanitized_session_payload, '$.profile_binding') <> json_extract(OLD.sanitized_session_payload, '$.profile_binding')
  OR json_extract(NEW.sanitized_session_payload, '$.authority.authority_digest') <> json_extract(OLD.sanitized_session_payload, '$.authority.authority_digest')
  OR json_extract(NEW.sanitized_session_payload, '$.created_at') <> json_extract(OLD.sanitized_session_payload, '$.created_at')
  OR json_extract(NEW.sanitized_session_payload, '$.created_by_actor_ref') <> json_extract(OLD.sanitized_session_payload, '$.created_by_actor_ref')
  OR NEW.record_version <> OLD.record_version + 1
  OR OLD.session_state <> 'active'
  OR NEW.session_state NOT IN ('active', 'closed')
BEGIN SELECT RAISE(ABORT, 'application interaction session transition is invalid'); END;
CREATE TRIGGER application_interaction_sessions_no_delete BEFORE DELETE ON application_interaction_sessions
BEGIN SELECT RAISE(ABORT, 'application interaction sessions cannot be deleted'); END;

CREATE TRIGGER application_interaction_turns_controlled_update
BEFORE UPDATE ON application_interaction_session_turns
WHEN NEW.tenant_ref <> OLD.tenant_ref
  OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.application_id <> OLD.application_id
  OR NEW.owner_subject_ref <> OLD.owner_subject_ref
  OR NEW.session_id <> OLD.session_id
  OR NEW.turn_id <> OLD.turn_id
  OR NEW.turn_sequence <> OLD.turn_sequence
  OR NEW.client_turn_key <> OLD.client_turn_key
  OR NEW.started_at_unix_nano <> OLD.started_at_unix_nano
  OR json_extract(NEW.sanitized_turn_payload, '$.execution_profile') <> json_extract(OLD.sanitized_turn_payload, '$.execution_profile')
  OR json_extract(NEW.sanitized_turn_payload, '$.authority.authority_digest') <> json_extract(OLD.sanitized_turn_payload, '$.authority.authority_digest')
  OR json_extract(NEW.sanitized_turn_payload, '$.input_digest') <> json_extract(OLD.sanitized_turn_payload, '$.input_digest')
  OR json_extract(NEW.sanitized_turn_payload, '$.input_bytes') <> json_extract(OLD.sanitized_turn_payload, '$.input_bytes')
  OR OLD.turn_status <> 'running'
  OR NEW.turn_status = 'running'
BEGIN SELECT RAISE(ABORT, 'application interaction turn transition is invalid'); END;
CREATE TRIGGER application_interaction_turns_no_delete BEFORE DELETE ON application_interaction_session_turns
BEGIN SELECT RAISE(ABORT, 'application interaction turns cannot be deleted'); END;
