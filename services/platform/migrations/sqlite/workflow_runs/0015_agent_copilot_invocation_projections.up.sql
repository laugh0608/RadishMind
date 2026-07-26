CREATE TABLE agent_copilot_sessions (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    session_id TEXT NOT NULL CHECK (session_id GLOB 'appsess_[a-z2-7]*' AND length(session_id) = 24 AND substr(session_id, 9) NOT GLOB '*[^a-z2-7]*'),
    session_state TEXT NOT NULL CHECK (session_state IN ('active', 'closed')),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    authority_digest TEXT NOT NULL CHECK (length(authority_digest) = 71),
    sanitized_session_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_session_payload) AND json_type(sanitized_session_payload) = 'object'
        AND json_extract(sanitized_session_payload, '$.schema_version') = 'application_session.v3'
        AND json_extract(sanitized_session_payload, '$.session_id') = session_id
        AND json_extract(sanitized_session_payload, '$.state') = session_state
        AND json_extract(sanitized_session_payload, '$.record_version') = record_version
        AND json_extract(sanitized_session_payload, '$.profile_binding.execution_profile') = 'agent_copilot_suggestion_v1'
        AND json_extract(sanitized_session_payload, '$.authority.authority_digest') = authority_digest
        AND json_extract(sanitized_session_payload, '$.content_retention') = 'metadata_only'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
) STRICT;

CREATE TABLE agent_copilot_session_turns (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    session_id TEXT NOT NULL,
    turn_id TEXT NOT NULL CHECK (turn_id GLOB 'appturn_[a-z2-7]*' AND length(turn_id) = 24 AND substr(turn_id, 9) NOT GLOB '*[^a-z2-7]*'),
    turn_sequence INTEGER NOT NULL CHECK (turn_sequence > 0),
    client_turn_key TEXT NOT NULL CHECK (length(trim(client_turn_key)) > 0),
    turn_status TEXT NOT NULL CHECK (turn_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown')),
    started_at_unix_nano INTEGER NOT NULL CHECK (started_at_unix_nano > 0),
    completed_at_unix_nano INTEGER,
    sanitized_turn_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_turn_payload) AND json_type(sanitized_turn_payload) = 'object'
        AND json_extract(sanitized_turn_payload, '$.schema_version') = 'application_session_turn.v3'
        AND json_extract(sanitized_turn_payload, '$.turn_id') = turn_id
        AND json_extract(sanitized_turn_payload, '$.session_id') = session_id
        AND json_extract(sanitized_turn_payload, '$.sequence') = turn_sequence
        AND json_extract(sanitized_turn_payload, '$.client_turn_key') = client_turn_key
        AND json_extract(sanitized_turn_payload, '$.status') = turn_status
        AND json_extract(sanitized_turn_payload, '$.execution_profile') = 'agent_copilot_suggestion_v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, client_turn_key),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        REFERENCES agent_copilot_sessions (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK ((turn_status = 'running' AND completed_at_unix_nano IS NULL) OR (turn_status <> 'running' AND completed_at_unix_nano >= started_at_unix_nano))
) STRICT;

CREATE TABLE agent_copilot_run_records (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL,
    run_id TEXT NOT NULL CHECK (run_id GLOB 'run_[a-z0-9]*' AND length(run_id) BETWEEN 20 AND 68 AND substr(run_id, 5) NOT GLOB '*[^a-z0-9]*'),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    run_status TEXT NOT NULL CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown')),
    template_id TEXT NOT NULL CHECK (template_id GLOB 'acpf_[a-z2-7]*' AND length(template_id) = 21),
    template_version INTEGER NOT NULL CHECK (template_version > 0),
    authority_digest TEXT NOT NULL CHECK (length(authority_digest) = 71),
    started_at_unix_nano INTEGER NOT NULL CHECK (started_at_unix_nano > 0),
    completed_at_unix_nano INTEGER,
    sanitized_run_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_run_payload) AND json_type(sanitized_run_payload) = 'object'
        AND json_extract(sanitized_run_payload, '$.schema_version') = 'workflow_run_record.v7'
        AND json_extract(sanitized_run_payload, '$.run_id') = run_id
        AND json_extract(sanitized_run_payload, '$.record_version') = record_version
        AND json_extract(sanitized_run_payload, '$.status') = run_status
        AND json_extract(sanitized_run_payload, '$.execution_profile') = 'agent_copilot_suggestion_v1'
        AND json_extract(sanitized_run_payload, '$.execution_source_id') = template_id
        AND json_extract(sanitized_run_payload, '$.execution_source_version') = template_version
        AND json_extract(sanitized_run_payload, '$.authority.authority_digest') = authority_digest
        AND json_extract(sanitized_run_payload, '$.output') = ''
        AND json_type(sanitized_run_payload, '$.context') IS NULL
        AND json_type(sanitized_run_payload, '$.artifacts') IS NULL
        AND json_type(sanitized_run_payload, '$.' || 'raw_' || 'response') IS NULL
        AND json_type(sanitized_run_payload, '$.' || 'provider_' || 'api_' || 'key') IS NULL
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, run_id),
    CHECK ((run_status = 'running' AND completed_at_unix_nano IS NULL) OR (run_status <> 'running' AND completed_at_unix_nano >= started_at_unix_nano))
) STRICT;

CREATE INDEX agent_copilot_sessions_scope_idx ON agent_copilot_sessions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_state, updated_at_unix_nano DESC, session_id DESC);
CREATE INDEX agent_copilot_turns_scope_idx ON agent_copilot_session_turns
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence);
CREATE INDEX agent_copilot_runs_history_idx ON agent_copilot_run_records
    (tenant_ref, workspace_id, application_id, started_at_unix_nano DESC, run_id DESC);

CREATE TRIGGER agent_copilot_sessions_controlled_update BEFORE UPDATE ON agent_copilot_sessions
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.session_id <> OLD.session_id
 OR NEW.authority_digest <> OLD.authority_digest OR NEW.record_version <> OLD.record_version + 1 OR OLD.session_state <> 'active'
BEGIN SELECT RAISE(ABORT, 'Agent Copilot session transition is invalid'); END;
CREATE TRIGGER agent_copilot_sessions_no_delete BEFORE DELETE ON agent_copilot_sessions
BEGIN SELECT RAISE(ABORT, 'Agent Copilot sessions cannot be deleted'); END;
CREATE TRIGGER agent_copilot_turns_controlled_update BEFORE UPDATE ON agent_copilot_session_turns
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.session_id <> OLD.session_id OR NEW.turn_id <> OLD.turn_id
 OR NEW.turn_sequence <> OLD.turn_sequence OR NEW.client_turn_key <> OLD.client_turn_key OR NEW.started_at_unix_nano <> OLD.started_at_unix_nano
 OR OLD.turn_status <> 'running' OR NEW.turn_status = 'running'
BEGIN SELECT RAISE(ABORT, 'Agent Copilot turn transition is invalid'); END;
CREATE TRIGGER agent_copilot_turns_no_delete BEFORE DELETE ON agent_copilot_session_turns
BEGIN SELECT RAISE(ABORT, 'Agent Copilot turns cannot be deleted'); END;
CREATE TRIGGER agent_copilot_runs_controlled_update BEFORE UPDATE ON agent_copilot_run_records
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.run_id <> OLD.run_id OR NEW.template_id <> OLD.template_id OR NEW.template_version <> OLD.template_version
 OR NEW.authority_digest <> OLD.authority_digest OR NEW.started_at_unix_nano <> OLD.started_at_unix_nano
 OR NEW.record_version <> OLD.record_version + 1 OR OLD.run_status <> 'running' OR NEW.run_status = 'running'
BEGIN SELECT RAISE(ABORT, 'Agent Copilot run transition is invalid'); END;
CREATE TRIGGER agent_copilot_runs_no_delete BEFORE DELETE ON agent_copilot_run_records
BEGIN SELECT RAISE(ABORT, 'Agent Copilot runs cannot be deleted'); END;
