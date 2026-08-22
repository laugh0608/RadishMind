CREATE TABLE application_result_artifact_lifecycles (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    artifact_id TEXT NOT NULL,
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    lifecycle_version INTEGER NOT NULL CHECK (lifecycle_version >= 1),
    archived_at_unix_nano INTEGER CHECK (
        (lifecycle_state = 'active' AND archived_at_unix_nano IS NULL)
        OR (lifecycle_state = 'archived' AND archived_at_unix_nano > 0)
    ),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    updated_by_actor_ref TEXT NOT NULL CHECK (length(trim(updated_by_actor_ref)) > 0),
    request_id TEXT NOT NULL CHECK (length(trim(request_id)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    lifecycle_payload TEXT NOT NULL CHECK (
        json_valid(lifecycle_payload) AND json_type(lifecycle_payload) = 'object'
        AND json_extract(lifecycle_payload, '$.schema_version') = 'application_result_artifact_lifecycle.v1'
        AND json_extract(lifecycle_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(lifecycle_payload, '$.workspace_id') = workspace_id
        AND json_extract(lifecycle_payload, '$.application_id') = application_id
        AND json_extract(lifecycle_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(lifecycle_payload, '$.artifact_id') = artifact_id
        AND json_extract(lifecycle_payload, '$.lifecycle_state') = lifecycle_state
        AND json_extract(lifecycle_payload, '$.lifecycle_version') = lifecycle_version
        AND json_extract(lifecycle_payload, '$.updated_by_actor_ref') = updated_by_actor_ref
        AND json_extract(lifecycle_payload, '$.request_id') = request_id
        AND json_extract(lifecycle_payload, '$.audit_ref') = audit_ref
        AND ((archived_at_unix_nano IS NULL AND json_type(lifecycle_payload, '$.archived_at') = 'null')
             OR (archived_at_unix_nano IS NOT NULL AND json_type(lifecycle_payload, '$.archived_at') = 'text'))
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
        REFERENCES application_result_artifacts (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
) STRICT;

INSERT INTO application_result_artifact_lifecycles (
    tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id,
    lifecycle_state, lifecycle_version, archived_at_unix_nano, updated_at_unix_nano,
    updated_by_actor_ref, request_id, audit_ref, lifecycle_payload
)
SELECT tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id,
       'active', 1, NULL, created_at_unix_nano,
       json_extract(artifact_payload, '$.created_by_actor_ref'),
       json_extract(artifact_payload, '$.request_id'),
       json_extract(artifact_payload, '$.audit_ref'),
       json_object(
           'schema_version', 'application_result_artifact_lifecycle.v1',
           'tenant_ref', tenant_ref,
           'workspace_id', workspace_id,
           'application_id', application_id,
           'owner_subject_ref', owner_subject_ref,
           'artifact_id', artifact_id,
           'lifecycle_state', 'active',
           'lifecycle_version', 1,
           'archived_at', NULL,
           'updated_at', json_extract(artifact_payload, '$.created_at'),
           'updated_by_actor_ref', json_extract(artifact_payload, '$.created_by_actor_ref'),
           'request_id', json_extract(artifact_payload, '$.request_id'),
           'audit_ref', json_extract(artifact_payload, '$.audit_ref')
       )
  FROM application_result_artifacts;

CREATE INDEX application_result_artifact_lifecycles_session_state_idx
    ON application_result_artifact_lifecycles (
        tenant_ref, workspace_id, application_id, owner_subject_ref,
        lifecycle_state, artifact_id
    );

CREATE TABLE application_result_artifact_lifecycle_events (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    application_id TEXT NOT NULL,
    owner_subject_ref TEXT NOT NULL,
    artifact_id TEXT NOT NULL,
    lifecycle_version INTEGER NOT NULL CHECK (lifecycle_version >= 2),
    from_state TEXT NOT NULL CHECK (from_state IN ('active', 'archived')),
    to_state TEXT NOT NULL CHECK (to_state IN ('active', 'archived') AND to_state <> from_state),
    transition_kind TEXT NOT NULL CHECK (
        (transition_kind = 'archived' AND from_state = 'active' AND to_state = 'archived')
        OR (transition_kind = 'unarchived' AND from_state = 'archived' AND to_state = 'active')
    ),
    occurred_at_unix_nano INTEGER NOT NULL CHECK (occurred_at_unix_nano > 0),
    actor_ref TEXT NOT NULL CHECK (length(trim(actor_ref)) > 0),
    request_id TEXT NOT NULL CHECK (length(trim(request_id)) > 0),
    audit_ref TEXT NOT NULL CHECK (length(trim(audit_ref)) > 0),
    event_payload TEXT NOT NULL CHECK (
        json_valid(event_payload) AND json_type(event_payload) = 'object'
        AND json_extract(event_payload, '$.schema_version') = 'application_result_artifact_lifecycle_event.v1'
        AND json_extract(event_payload, '$.tenant_ref') = tenant_ref
        AND json_extract(event_payload, '$.workspace_id') = workspace_id
        AND json_extract(event_payload, '$.application_id') = application_id
        AND json_extract(event_payload, '$.owner_subject_ref') = owner_subject_ref
        AND json_extract(event_payload, '$.artifact_id') = artifact_id
        AND json_extract(event_payload, '$.lifecycle_version') = lifecycle_version
        AND json_extract(event_payload, '$.from_state') = from_state
        AND json_extract(event_payload, '$.to_state') = to_state
        AND json_extract(event_payload, '$.transition_kind') = transition_kind
        AND json_extract(event_payload, '$.actor_ref') = actor_ref
        AND json_extract(event_payload, '$.request_id') = request_id
        AND json_extract(event_payload, '$.audit_ref') = audit_ref
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id, lifecycle_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
        REFERENCES application_result_artifact_lifecycles (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
) STRICT;

CREATE TRIGGER application_result_artifact_lifecycles_controlled_update
BEFORE UPDATE ON application_result_artifact_lifecycles
WHEN NEW.tenant_ref <> OLD.tenant_ref
  OR NEW.workspace_id <> OLD.workspace_id
  OR NEW.application_id <> OLD.application_id
  OR NEW.owner_subject_ref <> OLD.owner_subject_ref
  OR NEW.artifact_id <> OLD.artifact_id
  OR NEW.lifecycle_version <> OLD.lifecycle_version + 1
  OR NEW.lifecycle_state = OLD.lifecycle_state
BEGIN
    SELECT RAISE(ABORT, 'application result artifact lifecycle update is invalid');
END;

CREATE TRIGGER application_result_artifact_lifecycles_no_delete
BEFORE DELETE ON application_result_artifact_lifecycles
BEGIN
    SELECT RAISE(ABORT, 'application result artifact lifecycle cannot be deleted');
END;

CREATE TRIGGER application_result_artifact_lifecycle_events_no_update
BEFORE UPDATE ON application_result_artifact_lifecycle_events
BEGIN
    SELECT RAISE(ABORT, 'application result artifact lifecycle event is append-only');
END;

CREATE TRIGGER application_result_artifact_lifecycle_events_no_delete
BEFORE DELETE ON application_result_artifact_lifecycle_events
BEGIN
    SELECT RAISE(ABORT, 'application result artifact lifecycle event is append-only');
END;
