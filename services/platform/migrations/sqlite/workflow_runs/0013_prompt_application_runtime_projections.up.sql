CREATE TABLE prompt_application_runtime_assignments (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    assignment_id TEXT NOT NULL CHECK (assignment_id GLOB 'ptra_[a-z2-7]*' AND length(assignment_id) = 21 AND substr(assignment_id, 6) NOT GLOB '*[^a-z2-7]*'),
    assignment_version INTEGER NOT NULL CHECK (assignment_version > 0),
    assignment_state TEXT NOT NULL CHECK (assignment_state IN ('active', 'revoked')),
    assignment_digest TEXT NOT NULL CHECK (length(assignment_digest) = 71),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_assignment_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_assignment_payload) AND json_type(sanitized_assignment_payload) = 'object'
        AND json_extract(sanitized_assignment_payload, '$.schema_version') = 'prompt_application_runtime_assignment.v1'
        AND json_extract(sanitized_assignment_payload, '$.assignment_id') = assignment_id
        AND json_extract(sanitized_assignment_payload, '$.assignment_version') = assignment_version
        AND json_extract(sanitized_assignment_payload, '$.state') = assignment_state
        AND json_extract(sanitized_assignment_payload, '$.assignment_digest') = assignment_digest
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref)
) STRICT;

CREATE TABLE prompt_application_runtime_assignment_events (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    event_id TEXT NOT NULL CHECK (event_id GLOB 'ptrae_[a-z2-7]*' AND length(event_id) = 22 AND substr(event_id, 7) NOT GLOB '*[^a-z2-7]*'),
    assignment_id TEXT NOT NULL, event_sequence INTEGER NOT NULL CHECK (event_sequence > 0),
    resulting_assignment_version INTEGER NOT NULL CHECK (resulting_assignment_version > 0),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    sanitized_event_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_event_payload) AND json_type(sanitized_event_payload) = 'object'
        AND json_extract(sanitized_event_payload, '$.schema_version') = 'prompt_application_runtime_assignment_event.v1'
        AND json_extract(sanitized_event_payload, '$.event_id') = event_id
        AND json_extract(sanitized_event_payload, '$.assignment_id') = assignment_id
        AND json_extract(sanitized_event_payload, '$.event_sequence') = event_sequence
        AND json_extract(sanitized_event_payload, '$.resulting_assignment_version') = resulting_assignment_version
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, event_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, event_sequence),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref)
        REFERENCES prompt_application_runtime_assignments (tenant_ref, workspace_id, application_id, owner_subject_ref)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
) STRICT;

CREATE TABLE prompt_application_sessions (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    session_id TEXT NOT NULL CHECK (session_id GLOB 'appsess_[a-z2-7]*' AND length(session_id) = 24 AND substr(session_id, 9) NOT GLOB '*[^a-z2-7]*'),
    session_state TEXT NOT NULL CHECK (session_state IN ('active', 'closed')),
    record_version INTEGER NOT NULL CHECK (record_version > 0), updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    authority_digest TEXT NOT NULL CHECK (length(authority_digest) = 71),
    sanitized_session_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_session_payload) AND json_type(sanitized_session_payload) = 'object'
        AND json_extract(sanitized_session_payload, '$.schema_version') = 'application_session.v2'
        AND json_extract(sanitized_session_payload, '$.session_id') = session_id
        AND json_extract(sanitized_session_payload, '$.state') = session_state
        AND json_extract(sanitized_session_payload, '$.record_version') = record_version
        AND json_extract(sanitized_session_payload, '$.profile_binding.execution_profile') = 'prompt_application_invocation_v1'
        AND json_extract(sanitized_session_payload, '$.authority.authority_digest') = authority_digest
        AND json_extract(sanitized_session_payload, '$.content_retention') = 'metadata_only'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
) STRICT;

CREATE TABLE prompt_application_session_turns (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    session_id TEXT NOT NULL, turn_id TEXT NOT NULL CHECK (turn_id GLOB 'appturn_[a-z2-7]*' AND length(turn_id) = 24 AND substr(turn_id, 9) NOT GLOB '*[^a-z2-7]*'),
    turn_sequence INTEGER NOT NULL CHECK (turn_sequence > 0), client_turn_key TEXT NOT NULL CHECK (length(trim(client_turn_key)) > 0),
    turn_status TEXT NOT NULL CHECK (turn_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown')),
    started_at_unix_nano INTEGER NOT NULL CHECK (started_at_unix_nano > 0), completed_at_unix_nano INTEGER,
    sanitized_turn_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_turn_payload) AND json_type(sanitized_turn_payload) = 'object'
        AND json_extract(sanitized_turn_payload, '$.schema_version') = 'application_session_turn.v2'
        AND json_extract(sanitized_turn_payload, '$.turn_id') = turn_id
        AND json_extract(sanitized_turn_payload, '$.session_id') = session_id
        AND json_extract(sanitized_turn_payload, '$.sequence') = turn_sequence
        AND json_extract(sanitized_turn_payload, '$.client_turn_key') = client_turn_key
        AND json_extract(sanitized_turn_payload, '$.status') = turn_status
        AND json_extract(sanitized_turn_payload, '$.execution_profile') = 'prompt_application_invocation_v1'
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_id),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence),
    UNIQUE (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, client_turn_key),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        REFERENCES prompt_application_sessions (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    CHECK ((turn_status = 'running' AND completed_at_unix_nano IS NULL) OR (turn_status <> 'running' AND completed_at_unix_nano >= started_at_unix_nano))
) STRICT;

CREATE TABLE prompt_application_run_records (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL,
    run_id TEXT NOT NULL CHECK (run_id GLOB 'run_[a-z0-9]*' AND length(run_id) BETWEEN 20 AND 68 AND substr(run_id, 5) NOT GLOB '*[^a-z0-9]*'),
    record_version INTEGER NOT NULL CHECK (record_version > 0), run_status TEXT NOT NULL CHECK (run_status IN ('running', 'succeeded', 'failed', 'canceled', 'outcome_unknown')),
    template_id TEXT NOT NULL, template_version INTEGER NOT NULL CHECK (template_version > 0),
    authority_digest TEXT NOT NULL CHECK (length(authority_digest) = 71),
    started_at_unix_nano INTEGER NOT NULL CHECK (started_at_unix_nano > 0), completed_at_unix_nano INTEGER,
    sanitized_run_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_run_payload) AND json_type(sanitized_run_payload) = 'object'
        AND json_extract(sanitized_run_payload, '$.schema_version') = 'workflow_run_record.v6'
        AND json_extract(sanitized_run_payload, '$.run_id') = run_id
        AND json_extract(sanitized_run_payload, '$.record_version') = record_version
        AND json_extract(sanitized_run_payload, '$.status') = run_status
        AND json_extract(sanitized_run_payload, '$.execution_profile') = 'prompt_application_invocation_v1'
        AND json_extract(sanitized_run_payload, '$.execution_source_id') = template_id
        AND json_extract(sanitized_run_payload, '$.execution_source_version') = template_version
        AND json_extract(sanitized_run_payload, '$.authority.authority_digest') = authority_digest
        AND json_extract(sanitized_run_payload, '$.output') = ''
        AND json_type(sanitized_run_payload, '$.variables') IS NULL
        AND json_type(sanitized_run_payload, '$.messages') IS NULL
        AND json_type(sanitized_run_payload, '$.' || 'raw_' || 'response') IS NULL
        AND json_type(sanitized_run_payload, '$.' || 'provider_' || 'api_' || 'key') IS NULL
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, run_id),
    CHECK ((run_status = 'running' AND completed_at_unix_nano IS NULL) OR (run_status <> 'running' AND completed_at_unix_nano >= started_at_unix_nano))
) STRICT;

CREATE INDEX prompt_application_assignment_events_scope_idx ON prompt_application_runtime_assignment_events
    (tenant_ref, workspace_id, application_id, owner_subject_ref, event_sequence);
CREATE INDEX prompt_application_sessions_scope_idx ON prompt_application_sessions
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_state, updated_at_unix_nano DESC, session_id DESC);
CREATE INDEX prompt_application_turns_scope_idx ON prompt_application_session_turns
    (tenant_ref, workspace_id, application_id, owner_subject_ref, session_id, turn_sequence);
CREATE INDEX prompt_application_runs_history_idx ON prompt_application_run_records
    (tenant_ref, workspace_id, application_id, started_at_unix_nano DESC, run_id DESC);

CREATE TRIGGER prompt_application_assignments_controlled_update BEFORE UPDATE ON prompt_application_runtime_assignments
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.assignment_id <> OLD.assignment_id
 OR NEW.assignment_version <> OLD.assignment_version + 1 OR OLD.assignment_state = 'revoked'
BEGIN SELECT RAISE(ABORT, 'prompt application assignment transition is invalid'); END;
CREATE TRIGGER prompt_application_assignments_no_delete BEFORE DELETE ON prompt_application_runtime_assignments
BEGIN SELECT RAISE(ABORT, 'prompt application assignments cannot be deleted'); END;
CREATE TRIGGER prompt_application_assignment_events_no_update BEFORE UPDATE ON prompt_application_runtime_assignment_events
BEGIN SELECT RAISE(ABORT, 'prompt application assignment events are append-only'); END;
CREATE TRIGGER prompt_application_assignment_events_no_delete BEFORE DELETE ON prompt_application_runtime_assignment_events
BEGIN SELECT RAISE(ABORT, 'prompt application assignment events cannot be deleted'); END;
CREATE TRIGGER prompt_application_sessions_controlled_update BEFORE UPDATE ON prompt_application_sessions
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.session_id <> OLD.session_id
 OR NEW.authority_digest <> OLD.authority_digest OR NEW.record_version <> OLD.record_version + 1 OR OLD.session_state <> 'active'
BEGIN SELECT RAISE(ABORT, 'prompt application session transition is invalid'); END;
CREATE TRIGGER prompt_application_sessions_no_delete BEFORE DELETE ON prompt_application_sessions
BEGIN SELECT RAISE(ABORT, 'prompt application sessions cannot be deleted'); END;
CREATE TRIGGER prompt_application_turns_controlled_update BEFORE UPDATE ON prompt_application_session_turns
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.owner_subject_ref <> OLD.owner_subject_ref OR NEW.session_id <> OLD.session_id OR NEW.turn_id <> OLD.turn_id
 OR NEW.turn_sequence <> OLD.turn_sequence OR NEW.client_turn_key <> OLD.client_turn_key OR NEW.started_at_unix_nano <> OLD.started_at_unix_nano
 OR OLD.turn_status <> 'running' OR NEW.turn_status = 'running'
BEGIN SELECT RAISE(ABORT, 'prompt application turn transition is invalid'); END;
CREATE TRIGGER prompt_application_turns_no_delete BEFORE DELETE ON prompt_application_session_turns
BEGIN SELECT RAISE(ABORT, 'prompt application turns cannot be deleted'); END;
CREATE TRIGGER prompt_application_runs_controlled_update BEFORE UPDATE ON prompt_application_run_records
WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id OR NEW.application_id <> OLD.application_id
 OR NEW.run_id <> OLD.run_id OR NEW.template_id <> OLD.template_id OR NEW.template_version <> OLD.template_version
 OR NEW.authority_digest <> OLD.authority_digest OR NEW.started_at_unix_nano <> OLD.started_at_unix_nano
 OR NEW.record_version <> OLD.record_version + 1 OR OLD.run_status <> 'running' OR NEW.run_status = 'running'
BEGIN SELECT RAISE(ABORT, 'prompt application run transition is invalid'); END;
CREATE TRIGGER prompt_application_runs_no_delete BEFORE DELETE ON prompt_application_run_records
BEGIN SELECT RAISE(ABORT, 'prompt application runs cannot be deleted'); END;
