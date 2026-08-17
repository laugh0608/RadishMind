CREATE TABLE application_result_artifact_lifecycles (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    artifact_id text NOT NULL,
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active', 'archived')),
    lifecycle_version bigint NOT NULL CHECK (lifecycle_version >= 1),
    archived_at timestamptz,
    updated_at timestamptz NOT NULL,
    updated_by_actor_ref text NOT NULL CHECK (btrim(updated_by_actor_ref) <> ''),
    request_id text NOT NULL CHECK (btrim(request_id) <> ''),
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    lifecycle_payload jsonb NOT NULL,
    CONSTRAINT application_result_artifact_lifecycle_archive_check CHECK (
        (lifecycle_state = 'active' AND archived_at IS NULL)
        OR (lifecycle_state = 'archived' AND archived_at IS NOT NULL)
    ),
    CONSTRAINT application_result_artifact_lifecycle_payload_check CHECK (
        lifecycle_payload->>'schema_version' = 'application_result_artifact_lifecycle.v1'
        AND lifecycle_payload->>'tenant_ref' = tenant_ref
        AND lifecycle_payload->>'workspace_id' = workspace_id
        AND lifecycle_payload->>'application_id' = application_id
        AND lifecycle_payload->>'owner_subject_ref' = owner_subject_ref
        AND lifecycle_payload->>'artifact_id' = artifact_id
        AND lifecycle_payload->>'lifecycle_state' = lifecycle_state
        AND (lifecycle_payload->>'lifecycle_version')::bigint = lifecycle_version
        AND (lifecycle_payload->>'updated_at')::timestamptz = updated_at
        AND lifecycle_payload->>'updated_by_actor_ref' = updated_by_actor_ref
        AND lifecycle_payload->>'request_id' = request_id
        AND lifecycle_payload->>'audit_ref' = audit_ref
        AND ((archived_at IS NULL AND lifecycle_payload->'archived_at' = 'null'::jsonb)
             OR (archived_at IS NOT NULL AND (lifecycle_payload->>'archived_at')::timestamptz = archived_at))
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
        REFERENCES application_result_artifacts (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
);

INSERT INTO application_result_artifact_lifecycles (
    tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id,
    lifecycle_state, lifecycle_version, archived_at, updated_at,
    updated_by_actor_ref, request_id, audit_ref, lifecycle_payload
)
SELECT tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id,
       'active', 1, NULL, created_at,
       artifact_payload->>'created_by_actor_ref', artifact_payload->>'request_id', artifact_payload->>'audit_ref',
       jsonb_build_object(
           'schema_version', 'application_result_artifact_lifecycle.v1',
           'tenant_ref', tenant_ref,
           'workspace_id', workspace_id,
           'application_id', application_id,
           'owner_subject_ref', owner_subject_ref,
           'artifact_id', artifact_id,
           'lifecycle_state', 'active',
           'lifecycle_version', 1,
           'archived_at', NULL,
           'updated_at', artifact_payload->>'created_at',
           'updated_by_actor_ref', artifact_payload->>'created_by_actor_ref',
           'request_id', artifact_payload->>'request_id',
           'audit_ref', artifact_payload->>'audit_ref'
       )
  FROM application_result_artifacts;

CREATE INDEX application_result_artifact_lifecycles_state_idx
    ON application_result_artifact_lifecycles (
        tenant_ref, workspace_id, application_id, owner_subject_ref,
        lifecycle_state, artifact_id
    );

CREATE TABLE application_result_artifact_lifecycle_events (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    application_id text NOT NULL,
    owner_subject_ref text NOT NULL,
    artifact_id text NOT NULL,
    lifecycle_version bigint NOT NULL CHECK (lifecycle_version >= 2),
    from_state text NOT NULL CHECK (from_state IN ('active', 'archived')),
    to_state text NOT NULL CHECK (to_state IN ('active', 'archived') AND to_state <> from_state),
    transition_kind text NOT NULL CHECK (
        (transition_kind = 'archived' AND from_state = 'active' AND to_state = 'archived')
        OR (transition_kind = 'unarchived' AND from_state = 'archived' AND to_state = 'active')
    ),
    occurred_at timestamptz NOT NULL,
    actor_ref text NOT NULL CHECK (btrim(actor_ref) <> ''),
    request_id text NOT NULL CHECK (btrim(request_id) <> ''),
    audit_ref text NOT NULL CHECK (btrim(audit_ref) <> ''),
    event_payload jsonb NOT NULL,
    CONSTRAINT application_result_artifact_lifecycle_event_payload_check CHECK (
        event_payload->>'schema_version' = 'application_result_artifact_lifecycle_event.v1'
        AND event_payload->>'tenant_ref' = tenant_ref
        AND event_payload->>'workspace_id' = workspace_id
        AND event_payload->>'application_id' = application_id
        AND event_payload->>'owner_subject_ref' = owner_subject_ref
        AND event_payload->>'artifact_id' = artifact_id
        AND (event_payload->>'lifecycle_version')::bigint = lifecycle_version
        AND event_payload->>'from_state' = from_state
        AND event_payload->>'to_state' = to_state
        AND event_payload->>'transition_kind' = transition_kind
        AND (event_payload->>'occurred_at')::timestamptz = occurred_at
        AND event_payload->>'actor_ref' = actor_ref
        AND event_payload->>'request_id' = request_id
        AND event_payload->>'audit_ref' = audit_ref
    ),
    PRIMARY KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id, lifecycle_version),
    FOREIGN KEY (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
        REFERENCES application_result_artifact_lifecycles (tenant_ref, workspace_id, application_id, owner_subject_ref, artifact_id)
);

CREATE FUNCTION validate_application_result_artifact_lifecycle_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'application result artifact lifecycle cannot be deleted';
    END IF;
    IF NEW.tenant_ref <> OLD.tenant_ref
       OR NEW.workspace_id <> OLD.workspace_id
       OR NEW.application_id <> OLD.application_id
       OR NEW.owner_subject_ref <> OLD.owner_subject_ref
       OR NEW.artifact_id <> OLD.artifact_id
       OR NEW.lifecycle_version <> OLD.lifecycle_version + 1
       OR NEW.lifecycle_state = OLD.lifecycle_state THEN
        RAISE EXCEPTION 'application result artifact lifecycle update is invalid';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER application_result_artifact_lifecycles_controlled_mutation
    BEFORE UPDATE OR DELETE ON application_result_artifact_lifecycles
    FOR EACH ROW EXECUTE FUNCTION validate_application_result_artifact_lifecycle_mutation();

CREATE FUNCTION reject_application_result_artifact_lifecycle_event_mutation()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'application result artifact lifecycle event is append-only';
END;
$$;

CREATE TRIGGER application_result_artifact_lifecycle_events_append_only
    BEFORE UPDATE OR DELETE ON application_result_artifact_lifecycle_events
    FOR EACH ROW EXECUTE FUNCTION reject_application_result_artifact_lifecycle_event_mutation();
