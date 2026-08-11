DO $$
DECLARE constraint_name text;
BEGIN
    FOR constraint_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'application_evaluation_plans'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%application_evaluation_plan.v1%'
    LOOP
        EXECUTE format('ALTER TABLE application_evaluation_plans DROP CONSTRAINT %I', constraint_name);
    END LOOP;
    FOR constraint_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'application_evaluation_plan_versions'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%application_evaluation_plan_version.v1%'
    LOOP
        EXECUTE format('ALTER TABLE application_evaluation_plan_versions DROP CONSTRAINT %I', constraint_name);
    END LOOP;
    FOR constraint_name IN
        SELECT conname FROM pg_constraint
        WHERE conrelid = 'application_evaluation_campaigns'::regclass
          AND contype = 'c'
          AND pg_get_constraintdef(oid) LIKE '%application_evaluation_campaign.v1%'
    LOOP
        EXECUTE format('ALTER TABLE application_evaluation_campaigns DROP CONSTRAINT %I', constraint_name);
    END LOOP;
END $$;

ALTER TABLE application_evaluation_plans
    ADD CONSTRAINT application_evaluation_plans_payload_v2_check CHECK (
        jsonb_typeof(sanitized_plan_record) = 'object'
        AND sanitized_plan_record->>'schema_version' = CASE sanitized_plan_record->>'execution_profile'
            WHEN 'workflow_definition_executor_v2' THEN 'application_evaluation_plan.v2'
            ELSE 'application_evaluation_plan.v1'
        END
        AND sanitized_plan_record->>'execution_profile' IN (
            'workflow_definition_executor_v1','workflow_definition_executor_v2','application_rag_invocation_v1',
            'prompt_application_invocation_v1','agent_copilot_suggestion_v1'
        )
        AND sanitized_plan_record->>'plan_id' = plan_id
        AND (sanitized_plan_record->>'record_version')::bigint = record_version
        AND (sanitized_plan_record->>'latest_plan_version')::bigint = latest_plan_version
        AND sanitized_plan_record->>'latest_plan_digest' = latest_plan_digest
        AND sanitized_plan_record->>'lifecycle_state' = lifecycle_state
    );

ALTER TABLE application_evaluation_plan_versions
    ADD CONSTRAINT application_evaluation_plan_versions_payload_v2_check CHECK (
        jsonb_typeof(sanitized_plan_version_record) = 'object'
        AND sanitized_plan_version_record->>'schema_version' = CASE sanitized_plan_version_record->>'execution_profile'
            WHEN 'workflow_definition_executor_v2' THEN 'application_evaluation_plan_version.v2'
            ELSE 'application_evaluation_plan_version.v1'
        END
        AND sanitized_plan_version_record->>'execution_profile' IN (
            'workflow_definition_executor_v1','workflow_definition_executor_v2','application_rag_invocation_v1',
            'prompt_application_invocation_v1','agent_copilot_suggestion_v1'
        )
        AND sanitized_plan_version_record->>'plan_id' = plan_id
        AND (sanitized_plan_version_record->>'plan_version')::bigint = plan_version
        AND (
            sanitized_plan_version_record->>'execution_profile' <> 'workflow_definition_executor_v2'
            OR (
                jsonb_typeof(sanitized_plan_version_record#>'{target,workflow_definition,input_contract}') = 'object'
                AND btrim(sanitized_plan_version_record#>>'{target,workflow_definition,input_contract,contract_id}') <> ''
                AND sanitized_plan_version_record#>>'{target,workflow_definition,input_contract,contract_digest}' ~ '^sha256:[a-f0-9]{64}$'
                AND jsonb_typeof(sanitized_plan_version_record#>'{target,workflow_definition,input_contract,fields}') = 'array'
            )
        )
    );

ALTER TABLE application_evaluation_campaigns
    ADD CONSTRAINT application_evaluation_campaigns_payload_v2_check CHECK (
        jsonb_typeof(sanitized_campaign_record) = 'object'
        AND sanitized_campaign_record->>'schema_version' = CASE sanitized_campaign_record->>'execution_profile'
            WHEN 'workflow_definition_executor_v2' THEN 'application_evaluation_campaign.v2'
            ELSE 'application_evaluation_campaign.v1'
        END
        AND sanitized_campaign_record->>'execution_profile' IN (
            'workflow_definition_executor_v1','workflow_definition_executor_v2','application_rag_invocation_v1',
            'prompt_application_invocation_v1','agent_copilot_suggestion_v1'
        )
        AND sanitized_campaign_record->>'campaign_id' = campaign_id
        AND sanitized_campaign_record->>'client_campaign_key' = client_campaign_key
        AND (sanitized_campaign_record->>'record_version')::bigint = record_version
        AND sanitized_campaign_record->>'plan_id' = plan_id
        AND (sanitized_campaign_record->>'plan_version')::bigint = plan_version
        AND sanitized_campaign_record->>'plan_digest' = plan_digest
        AND sanitized_campaign_record->>'quota_api_key_id' = quota_api_key_id
        AND sanitized_campaign_record->>'state' = campaign_state
        AND NOT (sanitized_campaign_record ?| ARRAY['input_text','inputs'])
    );

CREATE OR REPLACE FUNCTION enforce_application_evaluation_plan_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.plan_id<>OLD.plan_id OR NEW.record_version<>OLD.record_version+1
 OR NEW.sanitized_plan_record->>'schema_version'<>OLD.sanitized_plan_record->>'schema_version'
 OR OLD.lifecycle_state='archived' OR NOT (NEW.lifecycle_state='active' OR (OLD.lifecycle_state='active' AND NEW.lifecycle_state='archived'))
THEN RAISE EXCEPTION 'Application evaluation plan transition is invalid'; END IF; RETURN NEW; END; $$;

CREATE OR REPLACE FUNCTION enforce_application_evaluation_campaign_update() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN
IF NEW.tenant_ref<>OLD.tenant_ref OR NEW.workspace_id<>OLD.workspace_id OR NEW.environment<>OLD.environment
 OR NEW.application_id<>OLD.application_id OR NEW.campaign_id<>OLD.campaign_id OR NEW.client_campaign_key<>OLD.client_campaign_key
 OR NEW.plan_id<>OLD.plan_id OR NEW.plan_version<>OLD.plan_version OR NEW.plan_digest<>OLD.plan_digest OR NEW.quota_api_key_id<>OLD.quota_api_key_id
 OR NEW.record_version<>OLD.record_version+1
 OR NEW.sanitized_campaign_record->>'schema_version'<>OLD.sanitized_campaign_record->>'schema_version'
 OR NEW.sanitized_campaign_record->>'execution_profile'<>OLD.sanitized_campaign_record->>'execution_profile'
 OR NOT ((OLD.campaign_state='pending' AND NEW.campaign_state='running')
     OR (OLD.campaign_state='running' AND NEW.campaign_state IN ('running','succeeded','failed','interrupted'))
     OR (OLD.campaign_state IN ('succeeded','failed','interrupted') AND NEW.campaign_state=OLD.campaign_state))
THEN RAISE EXCEPTION 'Application evaluation campaign transition is invalid'; END IF; RETURN NEW; END; $$;
