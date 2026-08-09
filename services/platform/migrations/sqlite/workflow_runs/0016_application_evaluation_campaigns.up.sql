CREATE TABLE application_evaluation_plans (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    latest_plan_version INTEGER NOT NULL CHECK (latest_plan_version > 0),
    latest_plan_digest TEXT NOT NULL CHECK (length(latest_plan_digest) = 71 AND substr(latest_plan_digest,1,7) = 'sha256:'),
    lifecycle_state TEXT NOT NULL CHECK (lifecycle_state IN ('active','archived')),
    updated_at_unix_nano INTEGER NOT NULL CHECK (updated_at_unix_nano > 0),
    sanitized_plan_record TEXT NOT NULL CHECK (
        json_valid(sanitized_plan_record)
        AND json_extract(sanitized_plan_record,'$.schema_version') = 'application_evaluation_plan.v1'
        AND json_extract(sanitized_plan_record,'$.plan_id') = plan_id
        AND json_extract(sanitized_plan_record,'$.record_version') = record_version
        AND json_extract(sanitized_plan_record,'$.latest_plan_version') = latest_plan_version
        AND json_extract(sanitized_plan_record,'$.latest_plan_digest') = latest_plan_digest
        AND json_extract(sanitized_plan_record,'$.lifecycle_state') = lifecycle_state
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,plan_id)
);

CREATE TABLE application_evaluation_plan_versions (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    plan_version INTEGER NOT NULL CHECK (plan_version > 0),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    sanitized_plan_version_record TEXT NOT NULL CHECK (
        json_valid(sanitized_plan_version_record)
        AND json_extract(sanitized_plan_version_record,'$.schema_version') = 'application_evaluation_plan_version.v1'
        AND json_extract(sanitized_plan_version_record,'$.plan_id') = plan_id
        AND json_extract(sanitized_plan_version_record,'$.plan_version') = plan_version
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,plan_id)
        REFERENCES application_evaluation_plans (tenant_ref,workspace_id,environment,application_id,plan_id)
        ON DELETE RESTRICT
);

CREATE TABLE application_evaluation_campaigns (
    tenant_ref TEXT NOT NULL,
    workspace_id TEXT NOT NULL,
    environment TEXT NOT NULL CHECK (environment IN ('development','test')),
    application_id TEXT NOT NULL,
    campaign_id TEXT NOT NULL,
    client_campaign_key TEXT NOT NULL CHECK (length(trim(client_campaign_key)) > 0),
    record_version INTEGER NOT NULL CHECK (record_version > 0),
    plan_id TEXT NOT NULL,
    plan_version INTEGER NOT NULL CHECK (plan_version > 0),
    plan_digest TEXT NOT NULL CHECK (length(plan_digest) = 71 AND substr(plan_digest,1,7) = 'sha256:'),
    quota_api_key_id TEXT NOT NULL CHECK (quota_api_key_id GLOB 'key_[a-z2-7]*' AND length(quota_api_key_id) = 20),
    campaign_state TEXT NOT NULL CHECK (campaign_state IN ('pending','running','succeeded','failed','interrupted')),
    created_at_unix_nano INTEGER NOT NULL CHECK (created_at_unix_nano > 0),
    sanitized_campaign_record TEXT NOT NULL CHECK (
        json_valid(sanitized_campaign_record)
        AND json_extract(sanitized_campaign_record,'$.schema_version') = 'application_evaluation_campaign.v1'
        AND json_extract(sanitized_campaign_record,'$.campaign_id') = campaign_id
        AND json_extract(sanitized_campaign_record,'$.client_campaign_key') = client_campaign_key
        AND json_extract(sanitized_campaign_record,'$.record_version') = record_version
        AND json_extract(sanitized_campaign_record,'$.plan_id') = plan_id
        AND json_extract(sanitized_campaign_record,'$.plan_version') = plan_version
        AND json_extract(sanitized_campaign_record,'$.plan_digest') = plan_digest
        AND json_extract(sanitized_campaign_record,'$.quota_api_key_id') = quota_api_key_id
        AND json_extract(sanitized_campaign_record,'$.state') = campaign_state
    ),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,campaign_id),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,client_campaign_key),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version)
        REFERENCES application_evaluation_plan_versions (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version)
        ON DELETE RESTRICT
);

CREATE INDEX application_evaluation_plans_scope_idx ON application_evaluation_plans
    (tenant_ref,workspace_id,environment,application_id,lifecycle_state,updated_at_unix_nano DESC,plan_id DESC);
CREATE INDEX application_evaluation_plan_versions_scope_idx ON application_evaluation_plan_versions
    (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version DESC);
CREATE INDEX application_evaluation_campaigns_scope_idx ON application_evaluation_campaigns
    (tenant_ref,workspace_id,environment,application_id,created_at_unix_nano DESC,campaign_id DESC);

CREATE TRIGGER application_evaluation_plans_controlled_update
BEFORE UPDATE ON application_evaluation_plans
FOR EACH ROW WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
 OR NEW.environment <> OLD.environment OR NEW.application_id <> OLD.application_id OR NEW.plan_id <> OLD.plan_id
 OR NEW.record_version <> OLD.record_version + 1 OR OLD.lifecycle_state = 'archived'
 OR NOT (NEW.lifecycle_state = 'active' OR (OLD.lifecycle_state = 'active' AND NEW.lifecycle_state = 'archived'))
BEGIN SELECT RAISE(ABORT,'application evaluation plan transition is invalid'); END;

CREATE TRIGGER application_evaluation_plans_no_delete
BEFORE DELETE ON application_evaluation_plans
BEGIN SELECT RAISE(ABORT,'application evaluation plan cannot be deleted'); END;

CREATE TRIGGER application_evaluation_plan_versions_no_update
BEFORE UPDATE ON application_evaluation_plan_versions
BEGIN SELECT RAISE(ABORT,'application evaluation plan version is immutable'); END;

CREATE TRIGGER application_evaluation_plan_versions_no_delete
BEFORE DELETE ON application_evaluation_plan_versions
BEGIN SELECT RAISE(ABORT,'application evaluation plan version cannot be deleted'); END;

CREATE TRIGGER application_evaluation_campaigns_controlled_update
BEFORE UPDATE ON application_evaluation_campaigns
FOR EACH ROW WHEN NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
 OR NEW.environment <> OLD.environment OR NEW.application_id <> OLD.application_id OR NEW.campaign_id <> OLD.campaign_id
 OR NEW.client_campaign_key <> OLD.client_campaign_key OR NEW.plan_id <> OLD.plan_id
 OR NEW.plan_version <> OLD.plan_version OR NEW.plan_digest <> OLD.plan_digest OR NEW.quota_api_key_id <> OLD.quota_api_key_id
 OR NEW.record_version <> OLD.record_version + 1
 OR NOT ((OLD.campaign_state = 'pending' AND NEW.campaign_state = 'running')
     OR (OLD.campaign_state = 'running' AND NEW.campaign_state IN ('running','succeeded','failed','interrupted'))
     OR (OLD.campaign_state IN ('succeeded','failed','interrupted') AND NEW.campaign_state = OLD.campaign_state))
BEGIN SELECT RAISE(ABORT,'application evaluation campaign transition is invalid'); END;

CREATE TRIGGER application_evaluation_campaigns_no_delete
BEFORE DELETE ON application_evaluation_campaigns
BEGIN SELECT RAISE(ABORT,'application evaluation campaign cannot be deleted'); END;
