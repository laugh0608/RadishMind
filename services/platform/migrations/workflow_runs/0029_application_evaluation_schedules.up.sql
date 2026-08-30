CREATE TABLE application_evaluation_schedules (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')),
    application_id text NOT NULL,
    schedule_id text NOT NULL CHECK (schedule_id ~ '^aesch_[a-z2-7]{16}$'),
    record_version bigint NOT NULL CHECK (record_version > 0),
    latest_schedule_version bigint NOT NULL CHECK (latest_schedule_version > 0),
    latest_schedule_digest text NOT NULL CHECK (latest_schedule_digest ~ '^sha256:[a-f0-9]{64}$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('draft','active','paused','archived')),
    updated_at timestamptz NOT NULL,
    next_due_at timestamptz,
    sanitized_schedule_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_schedule_record) = 'object'
        AND sanitized_schedule_record->>'schema_version' = 'application_evaluation_schedule.v1'
        AND sanitized_schedule_record->>'schedule_id' = schedule_id
        AND (sanitized_schedule_record->>'record_version')::bigint = record_version
        AND (sanitized_schedule_record->>'latest_schedule_version')::bigint = latest_schedule_version
        AND sanitized_schedule_record->>'latest_schedule_digest' = latest_schedule_digest
        AND sanitized_schedule_record->>'lifecycle_state' = lifecycle_state
        AND (sanitized_schedule_record->>'updated_at')::timestamptz = updated_at
        AND ((lifecycle_state = 'active' AND next_due_at IS NOT NULL
              AND (sanitized_schedule_record->>'next_due_at')::timestamptz = next_due_at)
             OR (lifecycle_state <> 'active' AND next_due_at IS NULL
                 AND sanitized_schedule_record->>'next_due_at' IS NULL))
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id)
);

CREATE TABLE application_evaluation_schedule_versions (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')),
    application_id text NOT NULL,
    schedule_id text NOT NULL,
    schedule_version bigint NOT NULL CHECK (schedule_version > 0),
    schedule_digest text NOT NULL CHECK (schedule_digest ~ '^sha256:[a-f0-9]{64}$'),
    created_at timestamptz NOT NULL,
    sanitized_schedule_version_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_schedule_version_record) = 'object'
        AND sanitized_schedule_version_record->>'schema_version' = 'application_evaluation_schedule_version.v1'
        AND sanitized_schedule_version_record->>'schedule_id' = schedule_id
        AND (sanitized_schedule_version_record->>'schedule_version')::bigint = schedule_version
        AND sanitized_schedule_version_record->>'schedule_digest' = schedule_digest
        AND (sanitized_schedule_version_record->>'created_at')::timestamptz = created_at
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,schedule_id)
        REFERENCES application_evaluation_schedules (tenant_ref,workspace_id,environment,application_id,schedule_id)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE application_evaluation_schedule_occurrences (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')),
    application_id text NOT NULL,
    schedule_id text NOT NULL,
    schedule_version bigint NOT NULL CHECK (schedule_version > 0),
    scheduled_for timestamptz NOT NULL,
    record_version bigint NOT NULL CHECK (record_version > 0),
    schedule_digest text NOT NULL CHECK (schedule_digest ~ '^sha256:[a-f0-9]{64}$'),
    occurrence_state text NOT NULL CHECK (occurrence_state IN ('claimed','campaign_created','observing','succeeded','failed','interrupted','skipped')),
    client_campaign_key text NOT NULL CHECK (client_campaign_key ~ '^scheduled_campaign_[a-f0-9]{24}$'),
    campaign_id text CHECK (campaign_id IS NULL OR campaign_id ~ '^aecamp_[a-z2-7]{16}$'),
    updated_at timestamptz NOT NULL,
    sanitized_occurrence_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_occurrence_record) = 'object'
        AND sanitized_occurrence_record->>'schema_version' = 'application_evaluation_schedule_occurrence.v1'
        AND sanitized_occurrence_record->>'schedule_id' = schedule_id
        AND (sanitized_occurrence_record->>'schedule_version')::bigint = schedule_version
        AND (sanitized_occurrence_record->>'scheduled_for_utc')::timestamptz = scheduled_for
        AND (sanitized_occurrence_record->>'record_version')::bigint = record_version
        AND sanitized_occurrence_record->>'schedule_digest' = schedule_digest
        AND sanitized_occurrence_record->>'state' = occurrence_state
        AND sanitized_occurrence_record->>'client_campaign_key' = client_campaign_key
        AND sanitized_occurrence_record->>'campaign_id' IS NOT DISTINCT FROM campaign_id
        AND (sanitized_occurrence_record->>'updated_at')::timestamptz = updated_at
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,scheduled_for),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,client_campaign_key),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest)
        REFERENCES application_evaluation_schedule_versions
            (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest)
        ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX application_evaluation_schedules_scope_idx ON application_evaluation_schedules
    (tenant_ref,workspace_id,environment,application_id,lifecycle_state,updated_at DESC,schedule_id DESC);
CREATE INDEX application_evaluation_schedules_due_idx ON application_evaluation_schedules
    (environment,lifecycle_state,next_due_at,schedule_id) WHERE lifecycle_state = 'active';
CREATE INDEX application_evaluation_schedule_occurrences_scope_idx ON application_evaluation_schedule_occurrences
    (tenant_ref,workspace_id,environment,application_id,schedule_id,scheduled_for DESC);

CREATE FUNCTION enforce_application_evaluation_schedule_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.schedule_id<>OLD.schedule_id OR NEW.record_version<>OLD.record_version+1
 OR OLD.lifecycle_state='archived'
 OR NOT (
    (NEW.latest_schedule_version=OLD.latest_schedule_version
     AND NEW.latest_schedule_digest=OLD.latest_schedule_digest
     AND ((OLD.lifecycle_state='draft' AND NEW.lifecycle_state IN ('active','archived'))
       OR (OLD.lifecycle_state='active' AND NEW.lifecycle_state IN ('active','paused','archived'))
       OR (OLD.lifecycle_state='paused' AND NEW.lifecycle_state IN ('active','archived'))))
    OR
    (NEW.latest_schedule_version=OLD.latest_schedule_version+1
     AND OLD.lifecycle_state IN ('draft','paused') AND NEW.lifecycle_state=OLD.lifecycle_state)
 )
THEN RAISE EXCEPTION 'Application evaluation schedule transition is invalid'; END IF;
RETURN NEW; END; $$;
CREATE TRIGGER application_evaluation_schedules_controlled_update BEFORE UPDATE ON application_evaluation_schedules
FOR EACH ROW EXECUTE FUNCTION enforce_application_evaluation_schedule_update();

CREATE FUNCTION enforce_application_evaluation_schedule_occurrence_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.schedule_id<>OLD.schedule_id OR NEW.schedule_version<>OLD.schedule_version
 OR NEW.scheduled_for<>OLD.scheduled_for OR NEW.schedule_digest<>OLD.schedule_digest OR NEW.client_campaign_key<>OLD.client_campaign_key
 OR NEW.record_version<>OLD.record_version+1
 OR NOT ((OLD.occurrence_state='claimed' AND NEW.occurrence_state IN ('campaign_created','failed','interrupted','skipped'))
      OR (OLD.occurrence_state='campaign_created' AND NEW.occurrence_state IN ('observing','failed','interrupted'))
      OR (OLD.occurrence_state='observing' AND NEW.occurrence_state IN ('succeeded','failed','interrupted')))
THEN RAISE EXCEPTION 'Application evaluation schedule occurrence transition is invalid'; END IF;
RETURN NEW; END; $$;
CREATE TRIGGER application_evaluation_schedule_occurrences_controlled_update
BEFORE UPDATE ON application_evaluation_schedule_occurrences
FOR EACH ROW EXECUTE FUNCTION enforce_application_evaluation_schedule_occurrence_update();

CREATE FUNCTION reject_application_evaluation_schedule_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'Application evaluation schedule record cannot be mutated'; END; $$;
CREATE TRIGGER application_evaluation_schedules_no_delete BEFORE DELETE ON application_evaluation_schedules
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_schedule_mutation();
CREATE TRIGGER application_evaluation_schedule_versions_no_update BEFORE UPDATE ON application_evaluation_schedule_versions
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_schedule_mutation();
CREATE TRIGGER application_evaluation_schedule_versions_no_delete BEFORE DELETE ON application_evaluation_schedule_versions
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_schedule_mutation();
CREATE TRIGGER application_evaluation_schedule_occurrences_no_delete BEFORE DELETE ON application_evaluation_schedule_occurrences
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_schedule_mutation();
