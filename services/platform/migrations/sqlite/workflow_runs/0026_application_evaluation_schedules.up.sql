CREATE TABLE application_evaluation_schedules (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    schedule_id TEXT NOT NULL CHECK (schedule_id GLOB 'aesch_[a-z2-7]*' AND length(schedule_id) = 22),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    latest_schedule_version INTEGER NOT NULL CHECK (latest_schedule_version > 0),
    latest_schedule_digest TEXT NOT NULL CHECK (length(latest_schedule_digest) = 71 AND substr(latest_schedule_digest,1,7) = 'sha256:'),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('draft','active','paused','archived')),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    next_due_at_unix_nano INTEGER,
    sanitized_schedule_record TEXT NOT NULL CHECK (
        json_valid(sanitized_schedule_record)
        AND json_extract(sanitized_schedule_record,'$.schema_version') = 'application_evaluation_schedule.v1'
        AND json_extract(sanitized_schedule_record,'$.schedule_id') = schedule_id
        AND json_extract(sanitized_schedule_record,'$.record_version') = record_version
        AND json_extract(sanitized_schedule_record,'$.latest_schedule_version') = latest_schedule_version
        AND json_extract(sanitized_schedule_record,'$.latest_schedule_digest') = latest_schedule_digest
        AND json_extract(sanitized_schedule_record,'$.lifecycle_state') = lifecycle_state
        AND ((lifecycle_state = 'active' AND next_due_at_unix_nano IS NOT NULL
              AND json_type(sanitized_schedule_record,'$.next_due_at') = 'text')
             OR (lifecycle_state <> 'active' AND next_due_at_unix_nano IS NULL
                 AND json_type(sanitized_schedule_record,'$.next_due_at') = 'null'))
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id)
);

CREATE TABLE application_evaluation_schedule_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    schedule_id TEXT NOT NULL,
    schedule_version INTEGER NOT NULL CHECK (schedule_version > 0),
    schedule_digest TEXT NOT NULL CHECK (length(schedule_digest) = 71 AND substr(schedule_digest,1,7) = 'sha256:'),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    sanitized_schedule_version_record TEXT NOT NULL CHECK (
        json_valid(sanitized_schedule_version_record)
        AND json_extract(sanitized_schedule_version_record,'$.schema_version') = 'application_evaluation_schedule_version.v1'
        AND json_extract(sanitized_schedule_version_record,'$.schedule_id') = schedule_id
        AND json_extract(sanitized_schedule_version_record,'$.schedule_version') = schedule_version
        AND json_extract(sanitized_schedule_version_record,'$.schedule_digest') = schedule_digest
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,schedule_id)
        REFERENCES application_evaluation_schedules (tenant_ref,workspace_id,environment,application_id,schedule_id)
        ON DELETE RESTRICT
);

CREATE TABLE application_evaluation_schedule_occurrences (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    schedule_id TEXT NOT NULL,
    schedule_version INTEGER NOT NULL CHECK (schedule_version > 0),
    scheduled_for_unix_nano INTEGER NOT NULL CHECK (scheduled_for_unix_nano > 0),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    schedule_digest TEXT NOT NULL CHECK (length(schedule_digest) = 71 AND substr(schedule_digest,1,7) = 'sha256:'),
    occurrence_state TEXT NOT NULL CHECK (occurrence_state IN ('claimed','campaign_created','observing','succeeded','failed','interrupted','skipped')),
    client_campaign_key TEXT NOT NULL CHECK (client_campaign_key GLOB 'scheduled_campaign_[a-f0-9]*' AND length(client_campaign_key) = 43),
    campaign_id TEXT CHECK (campaign_id IS NULL OR (campaign_id GLOB 'aecamp_[a-z2-7]*' AND length(campaign_id) = 23)),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_occurrence_record TEXT NOT NULL CHECK (
        json_valid(sanitized_occurrence_record)
        AND json_extract(sanitized_occurrence_record,'$.schema_version') = 'application_evaluation_schedule_occurrence.v1'
        AND json_extract(sanitized_occurrence_record,'$.schedule_id') = schedule_id
        AND json_extract(sanitized_occurrence_record,'$.schedule_version') = schedule_version
        AND json_extract(sanitized_occurrence_record,'$.record_version') = record_version
        AND json_extract(sanitized_occurrence_record,'$.schedule_digest') = schedule_digest
        AND json_extract(sanitized_occurrence_record,'$.state') = occurrence_state
        AND json_extract(sanitized_occurrence_record,'$.client_campaign_key') = client_campaign_key
        AND json_extract(sanitized_occurrence_record,'$.campaign_id') IS campaign_id
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,scheduled_for_unix_nano),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,client_campaign_key),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest)
        REFERENCES application_evaluation_schedule_versions
            (tenant_ref,workspace_id,environment,application_id,schedule_id,schedule_version,schedule_digest)
        ON DELETE RESTRICT
);

CREATE INDEX application_evaluation_schedules_scope_idx ON application_evaluation_schedules
    (tenant_ref,workspace_id,environment,application_id,lifecycle_state,updated_at_unix_nano DESC,schedule_id DESC);
CREATE INDEX application_evaluation_schedules_due_idx ON application_evaluation_schedules
    (environment,lifecycle_state,next_due_at_unix_nano,schedule_id) WHERE lifecycle_state = 'active';
CREATE INDEX application_evaluation_schedule_occurrences_scope_idx ON application_evaluation_schedule_occurrences
    (tenant_ref,workspace_id,environment,application_id,schedule_id,scheduled_for_unix_nano DESC);

CREATE TRIGGER application_evaluation_schedules_controlled_update
BEFORE UPDATE ON application_evaluation_schedules
FOR EACH ROW WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
 OR NEW.environment <> OLD.environment OR NEW.application_id <> OLD.application_id OR NEW.schedule_id <> OLD.schedule_id
 OR NEW.record_version <> OLD.record_version + 1 OR OLD.lifecycle_state = 'archived'
 OR NOT (
    (NEW.latest_schedule_version = OLD.latest_schedule_version
     AND NEW.latest_schedule_digest = OLD.latest_schedule_digest
     AND ((OLD.lifecycle_state = 'draft' AND NEW.lifecycle_state IN ('active','archived'))
       OR (OLD.lifecycle_state = 'active' AND NEW.lifecycle_state IN ('active','paused','archived'))
       OR (OLD.lifecycle_state = 'paused' AND NEW.lifecycle_state IN ('active','archived'))))
    OR
    (NEW.latest_schedule_version = OLD.latest_schedule_version + 1
     AND OLD.lifecycle_state IN ('draft','paused') AND NEW.lifecycle_state = OLD.lifecycle_state)
 )
BEGIN SELECT RAISE(ABORT,'application evaluation schedule transition is invalid'); END;

CREATE TRIGGER application_evaluation_schedule_occurrences_controlled_update
BEFORE UPDATE ON application_evaluation_schedule_occurrences
FOR EACH ROW WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
 OR NEW.environment <> OLD.environment OR NEW.application_id <> OLD.application_id OR NEW.schedule_id <> OLD.schedule_id
 OR NEW.schedule_version <> OLD.schedule_version OR NEW.scheduled_for_unix_nano <> OLD.scheduled_for_unix_nano
 OR NEW.schedule_digest <> OLD.schedule_digest OR NEW.client_campaign_key <> OLD.client_campaign_key
 OR NEW.record_version <> OLD.record_version + 1
 OR NOT ((OLD.occurrence_state = 'claimed' AND NEW.occurrence_state IN ('campaign_created','failed','interrupted','skipped'))
      OR (OLD.occurrence_state = 'campaign_created' AND NEW.occurrence_state IN ('observing','failed','interrupted'))
      OR (OLD.occurrence_state = 'observing' AND NEW.occurrence_state IN ('succeeded','failed','interrupted')))
BEGIN SELECT RAISE(ABORT,'application evaluation schedule occurrence transition is invalid'); END;

CREATE TRIGGER application_evaluation_schedules_no_delete
BEFORE DELETE ON application_evaluation_schedules
BEGIN SELECT RAISE(ABORT,'application evaluation schedule cannot be deleted'); END;

CREATE TRIGGER application_evaluation_schedule_versions_no_update
BEFORE UPDATE ON application_evaluation_schedule_versions
BEGIN SELECT RAISE(ABORT,'application evaluation schedule version is immutable'); END;

CREATE TRIGGER application_evaluation_schedule_versions_no_delete
BEFORE DELETE ON application_evaluation_schedule_versions
BEGIN SELECT RAISE(ABORT,'application evaluation schedule version cannot be deleted'); END;

CREATE TRIGGER application_evaluation_schedule_occurrences_no_delete
BEFORE DELETE ON application_evaluation_schedule_occurrences
BEGIN SELECT RAISE(ABORT,'application evaluation schedule occurrence cannot be deleted'); END;
