CREATE TABLE agent_copilot_runtime_assignments (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    assignment_id TEXT NOT NULL CHECK (assignment_id GLOB 'acra_????????????????'),
    assignment_version INTEGER NOT NULL CHECK (assignment_version > 0),
    assignment_state TEXT NOT NULL CHECK (assignment_state IN ('active','revoked')),
    assignment_digest TEXT NOT NULL CHECK (assignment_digest GLOB 'sha256:*' AND length(assignment_digest)=71),
    updated_at_unix_nano INTEGER NOT NULL,
    sanitized_assignment_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_assignment_payload)
        AND json_extract(sanitized_assignment_payload,'$.schema_version')='agent_copilot_runtime_assignment.v1'
        AND json_extract(sanitized_assignment_payload,'$.assignment_id')=assignment_id
        AND json_extract(sanitized_assignment_payload,'$.assignment_version')=assignment_version
        AND json_extract(sanitized_assignment_payload,'$.state')=assignment_state
        AND json_extract(sanitized_assignment_payload,'$.assignment_digest')=assignment_digest
        AND json_type(sanitized_assignment_payload,'$.profile_source') IS NULL
        AND json_type(sanitized_assignment_payload,'$.credential') IS NULL
        AND json_type(sanitized_assignment_payload,'$.token') IS NULL
        AND json_type(sanitized_assignment_payload,'$.header') IS NULL
        AND json_type(sanitized_assignment_payload,'$.endpoint') IS NULL
        AND json_type(sanitized_assignment_payload,'$.dsn') IS NULL
        AND json_type(sanitized_assignment_payload,'$.system_prompt') IS NULL
        AND json_type(sanitized_assignment_payload,'$.provider') IS NULL),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
);

CREATE TABLE agent_copilot_runtime_assignment_events (
    tenant_ref TEXT NOT NULL, workspace_id TEXT NOT NULL, application_id TEXT NOT NULL, owner_subject_ref TEXT NOT NULL,
    event_id TEXT NOT NULL CHECK (event_id GLOB 'acrae_????????????????'), assignment_id TEXT NOT NULL,
    event_sequence INTEGER NOT NULL CHECK (event_sequence > 0),
    resulting_assignment_version INTEGER NOT NULL CHECK (resulting_assignment_version > 0),
    occurred_at_unix_nano INTEGER NOT NULL,
    sanitized_event_payload TEXT NOT NULL CHECK (
        json_valid(sanitized_event_payload)
        AND json_extract(sanitized_event_payload,'$.schema_version')='agent_copilot_runtime_assignment_event.v1'
        AND json_extract(sanitized_event_payload,'$.event_id')=event_id
        AND json_extract(sanitized_event_payload,'$.assignment_id')=assignment_id
        AND json_extract(sanitized_event_payload,'$.event_sequence')=event_sequence
        AND json_extract(sanitized_event_payload,'$.resulting_assignment_version')=resulting_assignment_version
        AND json_type(sanitized_event_payload,'$.profile_source') IS NULL
        AND json_type(sanitized_event_payload,'$.credential') IS NULL
        AND json_type(sanitized_event_payload,'$.token') IS NULL
        AND json_type(sanitized_event_payload,'$.header') IS NULL
        AND json_type(sanitized_event_payload,'$.endpoint') IS NULL
        AND json_type(sanitized_event_payload,'$.dsn') IS NULL
        AND json_type(sanitized_event_payload,'$.system_prompt') IS NULL
        AND json_type(sanitized_event_payload,'$.provider') IS NULL),
    PRIMARY KEY (tenant_ref,workspace_id,application_id,owner_subject_ref,event_id),
    UNIQUE (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence),
    FOREIGN KEY (tenant_ref,workspace_id,application_id,owner_subject_ref)
        REFERENCES agent_copilot_runtime_assignments (tenant_ref,workspace_id,application_id,owner_subject_ref)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX agent_copilot_assignment_events_scope_idx
    ON agent_copilot_runtime_assignment_events
    (tenant_ref,workspace_id,application_id,owner_subject_ref,event_sequence);

CREATE TRIGGER agent_copilot_assignments_controlled_update
BEFORE UPDATE ON agent_copilot_runtime_assignments
WHEN NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id
 OR NEW.application_id<>OLD.application_id OR NEW.owner_subject_ref<>OLD.owner_subject_ref
 OR NEW.assignment_id<>OLD.assignment_id OR NEW.assignment_version<>OLD.assignment_version+1
 OR OLD.assignment_state='revoked'
BEGIN SELECT RAISE(ABORT,'agent copilot assignment transition is invalid'); END;

CREATE TRIGGER agent_copilot_assignments_no_delete
BEFORE DELETE ON agent_copilot_runtime_assignments
BEGIN SELECT RAISE(ABORT,'agent copilot assignment resource cannot be deleted'); END;

CREATE TRIGGER agent_copilot_assignment_events_no_update
BEFORE UPDATE ON agent_copilot_runtime_assignment_events
BEGIN SELECT RAISE(ABORT,'agent copilot assignment event cannot be updated'); END;

CREATE TRIGGER agent_copilot_assignment_events_no_delete
BEFORE DELETE ON agent_copilot_runtime_assignment_events
BEGIN SELECT RAISE(ABORT,'agent copilot assignment event cannot be deleted'); END;
