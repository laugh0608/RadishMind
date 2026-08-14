CREATE TABLE application_evaluation_plans (
    tenant_ref text NOT NULL, workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')), application_id text NOT NULL,
    plan_id text NOT NULL CHECK (plan_id ~ '^aeplan_[a-z2-7]{16}$'),
    record_version bigint NOT NULL CHECK (record_version > 0), latest_plan_version bigint NOT NULL CHECK (latest_plan_version > 0),
    latest_plan_digest text NOT NULL CHECK (latest_plan_digest ~ '^sha256:[a-f0-9]{64}$'),
    lifecycle_state text NOT NULL CHECK (lifecycle_state IN ('active','archived')), updated_at timestamptz NOT NULL,
    sanitized_plan_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_plan_record) = 'object'
        AND sanitized_plan_record->>'schema_version' = 'application_evaluation_plan.v1'
        AND sanitized_plan_record->>'plan_id' = plan_id
        AND (sanitized_plan_record->>'record_version')::bigint = record_version
        AND (sanitized_plan_record->>'latest_plan_version')::bigint = latest_plan_version
        AND sanitized_plan_record->>'latest_plan_digest' = latest_plan_digest
        AND sanitized_plan_record->>'lifecycle_state' = lifecycle_state),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,plan_id)
);

CREATE TABLE application_evaluation_plan_versions (
    tenant_ref text NOT NULL, workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')), application_id text NOT NULL,
    plan_id text NOT NULL, plan_version bigint NOT NULL CHECK (plan_version > 0), created_at timestamptz NOT NULL,
    sanitized_plan_version_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_plan_version_record) = 'object'
        AND sanitized_plan_version_record->>'schema_version' = 'application_evaluation_plan_version.v1'
        AND sanitized_plan_version_record->>'plan_id' = plan_id
        AND (sanitized_plan_version_record->>'plan_version')::bigint = plan_version),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,plan_id)
        REFERENCES application_evaluation_plans (tenant_ref,workspace_id,environment,application_id,plan_id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE application_evaluation_campaigns (
    tenant_ref text NOT NULL, workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development','test')), application_id text NOT NULL,
    campaign_id text NOT NULL CHECK (campaign_id ~ '^aecamp_[a-z2-7]{16}$'), client_campaign_key text NOT NULL CHECK (btrim(client_campaign_key) <> ''),
    record_version bigint NOT NULL CHECK (record_version > 0), plan_id text NOT NULL, plan_version bigint NOT NULL CHECK (plan_version > 0),
    plan_digest text NOT NULL CHECK (plan_digest ~ '^sha256:[a-f0-9]{64}$'),
    quota_api_key_id text NOT NULL CHECK (quota_api_key_id ~ '^key_[a-z2-7]{16}$'),
    campaign_state text NOT NULL CHECK (campaign_state IN ('pending','running','succeeded','failed','interrupted')), created_at timestamptz NOT NULL,
    sanitized_campaign_record jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_campaign_record) = 'object'
        AND sanitized_campaign_record->>'schema_version' = 'application_evaluation_campaign.v1'
        AND sanitized_campaign_record->>'campaign_id' = campaign_id
        AND sanitized_campaign_record->>'client_campaign_key' = client_campaign_key
        AND (sanitized_campaign_record->>'record_version')::bigint = record_version
        AND sanitized_campaign_record->>'plan_id' = plan_id
        AND (sanitized_campaign_record->>'plan_version')::bigint = plan_version
        AND sanitized_campaign_record->>'plan_digest' = plan_digest
        AND sanitized_campaign_record->>'quota_api_key_id' = quota_api_key_id
        AND sanitized_campaign_record->>'state' = campaign_state),
    PRIMARY KEY (tenant_ref,workspace_id,environment,application_id,campaign_id),
    UNIQUE (tenant_ref,workspace_id,environment,application_id,client_campaign_key),
    FOREIGN KEY (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version)
        REFERENCES application_evaluation_plan_versions (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX application_evaluation_plans_scope_idx ON application_evaluation_plans
    (tenant_ref,workspace_id,environment,application_id,lifecycle_state,updated_at DESC,plan_id DESC);
CREATE INDEX application_evaluation_plan_versions_scope_idx ON application_evaluation_plan_versions
    (tenant_ref,workspace_id,environment,application_id,plan_id,plan_version DESC);
CREATE INDEX application_evaluation_campaigns_scope_idx ON application_evaluation_campaigns
    (tenant_ref,workspace_id,environment,application_id,created_at DESC,campaign_id DESC);

CREATE FUNCTION enforce_application_evaluation_plan_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.plan_id<>OLD.plan_id OR NEW.record_version<>OLD.record_version+1
 OR OLD.lifecycle_state='archived' OR NOT (NEW.lifecycle_state='active' OR (OLD.lifecycle_state='active' AND NEW.lifecycle_state='archived'))
THEN RAISE EXCEPTION 'Application evaluation plan transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER application_evaluation_plans_controlled_update BEFORE UPDATE ON application_evaluation_plans
FOR EACH ROW EXECUTE FUNCTION enforce_application_evaluation_plan_update();

CREATE FUNCTION enforce_application_evaluation_campaign_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.campaign_id<>OLD.campaign_id OR NEW.client_campaign_key<>OLD.client_campaign_key
 OR NEW.plan_id<>OLD.plan_id OR NEW.plan_version<>OLD.plan_version OR NEW.plan_digest<>OLD.plan_digest OR NEW.quota_api_key_id<>OLD.quota_api_key_id
 OR NEW.record_version<>OLD.record_version+1
 OR NOT ((OLD.campaign_state='pending' AND NEW.campaign_state='running')
     OR (OLD.campaign_state='running' AND NEW.campaign_state IN ('running','succeeded','failed','interrupted'))
     OR (OLD.campaign_state IN ('succeeded','failed','interrupted') AND NEW.campaign_state=OLD.campaign_state))
THEN RAISE EXCEPTION 'Application evaluation campaign transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE TRIGGER application_evaluation_campaigns_controlled_update BEFORE UPDATE ON application_evaluation_campaigns
FOR EACH ROW EXECUTE FUNCTION enforce_application_evaluation_campaign_update();

CREATE FUNCTION reject_application_evaluation_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'Application evaluation record cannot be mutated'; END; $$;
CREATE TRIGGER application_evaluation_plans_no_delete BEFORE DELETE ON application_evaluation_plans
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_mutation();
CREATE TRIGGER application_evaluation_plan_versions_no_update BEFORE UPDATE ON application_evaluation_plan_versions
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_mutation();
CREATE TRIGGER application_evaluation_plan_versions_no_delete BEFORE DELETE ON application_evaluation_plan_versions
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_mutation();
CREATE TRIGGER application_evaluation_campaigns_no_delete BEFORE DELETE ON application_evaluation_campaigns
FOR EACH ROW EXECUTE FUNCTION reject_application_evaluation_mutation();
