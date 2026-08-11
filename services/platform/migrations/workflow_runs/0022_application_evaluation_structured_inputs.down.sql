ALTER TABLE application_evaluation_campaigns
    DROP CONSTRAINT application_evaluation_campaigns_payload_v2_check;
ALTER TABLE application_evaluation_plan_versions
    DROP CONSTRAINT application_evaluation_plan_versions_payload_v2_check;
ALTER TABLE application_evaluation_plans
    DROP CONSTRAINT application_evaluation_plans_payload_v2_check;

ALTER TABLE application_evaluation_plans
    ADD CONSTRAINT application_evaluation_plans_payload_v1_check CHECK (
        jsonb_typeof(sanitized_plan_record) = 'object'
        AND sanitized_plan_record->>'schema_version' = 'application_evaluation_plan.v1'
        AND sanitized_plan_record->>'plan_id' = plan_id
        AND (sanitized_plan_record->>'record_version')::bigint = record_version
        AND (sanitized_plan_record->>'latest_plan_version')::bigint = latest_plan_version
        AND sanitized_plan_record->>'latest_plan_digest' = latest_plan_digest
        AND sanitized_plan_record->>'lifecycle_state' = lifecycle_state
    );
ALTER TABLE application_evaluation_plan_versions
    ADD CONSTRAINT application_evaluation_plan_versions_payload_v1_check CHECK (
        jsonb_typeof(sanitized_plan_version_record) = 'object'
        AND sanitized_plan_version_record->>'schema_version' = 'application_evaluation_plan_version.v1'
        AND sanitized_plan_version_record->>'plan_id' = plan_id
        AND (sanitized_plan_version_record->>'plan_version')::bigint = plan_version
    );
ALTER TABLE application_evaluation_campaigns
    ADD CONSTRAINT application_evaluation_campaigns_payload_v1_check CHECK (
        jsonb_typeof(sanitized_campaign_record) = 'object'
        AND sanitized_campaign_record->>'schema_version' = 'application_evaluation_campaign.v1'
        AND sanitized_campaign_record->>'campaign_id' = campaign_id
        AND sanitized_campaign_record->>'client_campaign_key' = client_campaign_key
        AND (sanitized_campaign_record->>'record_version')::bigint = record_version
        AND sanitized_campaign_record->>'plan_id' = plan_id
        AND (sanitized_campaign_record->>'plan_version')::bigint = plan_version
        AND sanitized_campaign_record->>'plan_digest' = plan_digest
        AND sanitized_campaign_record->>'quota_api_key_id' = quota_api_key_id
        AND sanitized_campaign_record->>'state' = campaign_state
    );

CREATE OR REPLACE FUNCTION enforce_application_evaluation_plan_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.plan_id<>OLD.plan_id OR NEW.record_version<>OLD.record_version+1
 OR OLD.lifecycle_state='archived' OR NOT (NEW.lifecycle_state='active' OR (OLD.lifecycle_state='active' AND NEW.lifecycle_state='archived'))
THEN RAISE EXCEPTION 'Application evaluation plan transition is invalid'; END IF; RETURN NEW; END; $$;
CREATE OR REPLACE FUNCTION enforce_application_evaluation_campaign_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.campaign_id<>OLD.campaign_id OR NEW.client_campaign_key<>OLD.client_campaign_key
 OR NEW.plan_id<>OLD.plan_id OR NEW.plan_version<>OLD.plan_version OR NEW.plan_digest<>OLD.plan_digest OR NEW.quota_api_key_id<>OLD.quota_api_key_id
 OR NEW.record_version<>OLD.record_version+1
 OR NOT ((OLD.campaign_state='pending' AND NEW.campaign_state='running')
     OR (OLD.campaign_state='running' AND NEW.campaign_state IN ('running','succeeded','failed','interrupted'))
     OR (OLD.campaign_state IN ('succeeded','failed','interrupted') AND NEW.campaign_state=OLD.campaign_state))
THEN RAISE EXCEPTION 'Application evaluation campaign transition is invalid'; END IF; RETURN NEW; END; $$;
