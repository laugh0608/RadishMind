CREATE TABLE admin_provider_route_drafts (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id text NOT NULL,
    draft_revision bigint NOT NULL CHECK (draft_revision > 0),
    draft_digest text NOT NULL,
    sanitized_draft_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_draft_payload) = 'object'
        AND sanitized_draft_payload->>'schema_version' = 'admin_provider_route_configuration_draft.v1'
        AND sanitized_draft_payload->>'tenant_ref' = tenant_ref
        AND sanitized_draft_payload->>'workspace_id' = workspace_id
        AND sanitized_draft_payload->>'environment' = environment
        AND sanitized_draft_payload->>'configuration_id' = configuration_id
        AND (sanitized_draft_payload->>'draft_revision')::bigint = draft_revision
        AND sanitized_draft_payload->>'draft_digest' = draft_digest
    ),
    updated_at text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id)
);

CREATE TABLE admin_provider_route_candidates (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id text NOT NULL,
    candidate_id text NOT NULL,
    source_draft_revision bigint NOT NULL CHECK (source_draft_revision > 0),
    source_draft_digest text NOT NULL,
    candidate_digest text NOT NULL,
    candidate_state text NOT NULL CHECK (candidate_state IN ('pending_review', 'approved', 'rejected')),
    review_version bigint NOT NULL CHECK (review_version >= 0),
    sanitized_candidate_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_candidate_payload) = 'object'
        AND sanitized_candidate_payload->>'schema_version' = 'admin_provider_route_candidate.v1'
        AND sanitized_candidate_payload->>'tenant_ref' = tenant_ref
        AND sanitized_candidate_payload->>'workspace_id' = workspace_id
        AND sanitized_candidate_payload->>'environment' = environment
        AND sanitized_candidate_payload->>'configuration_id' = configuration_id
        AND sanitized_candidate_payload->>'candidate_id' = candidate_id
        AND (sanitized_candidate_payload->>'source_draft_revision')::bigint = source_draft_revision
        AND sanitized_candidate_payload->>'source_draft_digest' = source_draft_digest
        AND sanitized_candidate_payload->>'candidate_digest' = candidate_digest
        AND sanitized_candidate_payload->>'candidate_state' = candidate_state
        AND (sanitized_candidate_payload->>'review_version')::bigint = review_version
    ),
    created_at text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
);

CREATE TABLE admin_provider_route_reviews (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id text NOT NULL,
    candidate_id text NOT NULL,
    review_version bigint NOT NULL CHECK (review_version > 0),
    decision text NOT NULL CHECK (decision IN ('approve', 'reject')),
    resulting_state text NOT NULL CHECK (resulting_state IN ('approved', 'rejected')),
    sanitized_review_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_review_payload) = 'object'
        AND sanitized_review_payload->>'schema_version' = 'admin_provider_route_review.v1'
        AND (sanitized_review_payload->>'review_version')::bigint = review_version
        AND sanitized_review_payload->>'decision' = decision
        AND sanitized_review_payload->>'resulting_state' = resulting_state
    ),
    reviewed_at text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id, review_version),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
);

CREATE TABLE admin_provider_route_active_snapshots (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id text NOT NULL,
    generation bigint NOT NULL CHECK (generation > 0),
    candidate_id text NOT NULL,
    candidate_digest text NOT NULL,
    snapshot_digest text NOT NULL,
    sanitized_snapshot_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_snapshot_payload) = 'object'
        AND sanitized_snapshot_payload->>'schema_version' = 'admin_provider_route_snapshot.v1'
        AND sanitized_snapshot_payload->>'tenant_ref' = tenant_ref
        AND sanitized_snapshot_payload->>'workspace_id' = workspace_id
        AND sanitized_snapshot_payload->>'environment' = environment
        AND sanitized_snapshot_payload->>'configuration_id' = configuration_id
        AND (sanitized_snapshot_payload->>'generation')::bigint = generation
        AND sanitized_snapshot_payload->>'candidate_id' = candidate_id
        AND sanitized_snapshot_payload->>'candidate_digest' = candidate_digest
        AND sanitized_snapshot_payload->>'snapshot_digest' = snapshot_digest
    ),
    activated_at text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
);

CREATE TABLE admin_provider_route_activation_records (
    tenant_ref text NOT NULL,
    workspace_id text NOT NULL,
    environment text NOT NULL CHECK (environment IN ('development', 'test')),
    configuration_id text NOT NULL,
    after_generation bigint NOT NULL CHECK (after_generation > 0),
    activation_id text NOT NULL,
    action text NOT NULL CHECK (action IN ('activate', 'rollback')),
    after_candidate_id text NOT NULL,
    after_snapshot_digest text NOT NULL,
    previous_record_digest text NOT NULL,
    record_digest text NOT NULL,
    sanitized_activation_payload jsonb NOT NULL CHECK (
        jsonb_typeof(sanitized_activation_payload) = 'object'
        AND sanitized_activation_payload->>'schema_version' = 'admin_provider_route_activation_record.v1'
        AND sanitized_activation_payload->>'activation_id' = activation_id
        AND sanitized_activation_payload->>'configuration_id' = configuration_id
        AND sanitized_activation_payload->>'action' = action
        AND (sanitized_activation_payload->>'after_generation')::bigint = after_generation
        AND sanitized_activation_payload->>'after_candidate_id' = after_candidate_id
        AND sanitized_activation_payload->>'after_snapshot_digest' = after_snapshot_digest
        AND sanitized_activation_payload->>'previous_record_digest' = previous_record_digest
        AND sanitized_activation_payload->>'record_digest' = record_digest
    ),
    created_at text NOT NULL,
    PRIMARY KEY (tenant_ref, workspace_id, environment, configuration_id, after_generation),
    UNIQUE (tenant_ref, workspace_id, environment, configuration_id, activation_id),
    FOREIGN KEY (tenant_ref, workspace_id, environment, configuration_id, after_candidate_id)
        REFERENCES admin_provider_route_candidates (
            tenant_ref, workspace_id, environment, configuration_id, candidate_id
        )
);

CREATE INDEX admin_provider_route_candidates_scope_idx
    ON admin_provider_route_candidates (
        tenant_ref, workspace_id, environment, configuration_id, created_at, candidate_id
    );

CREATE INDEX admin_provider_route_activation_records_scope_idx
    ON admin_provider_route_activation_records (
        tenant_ref, workspace_id, environment, configuration_id, after_generation
    );

CREATE FUNCTION enforce_admin_provider_route_draft_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
       OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
       OR NEW.draft_revision <> OLD.draft_revision + 1
       OR NEW.sanitized_draft_payload->>'created_at' <> OLD.sanitized_draft_payload->>'created_at'
       OR NEW.sanitized_draft_payload->>'created_by_actor_ref' <> OLD.sanitized_draft_payload->>'created_by_actor_ref' THEN
        RAISE EXCEPTION 'admin provider route draft transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION enforce_admin_provider_route_candidate_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
       OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
       OR NEW.candidate_id <> OLD.candidate_id OR NEW.source_draft_revision <> OLD.source_draft_revision
       OR NEW.source_draft_digest <> OLD.source_draft_digest OR NEW.candidate_digest <> OLD.candidate_digest
       OR OLD.candidate_state <> 'pending_review' OR NEW.candidate_state NOT IN ('approved', 'rejected')
       OR NEW.review_version <> OLD.review_version + 1 OR NEW.created_at <> OLD.created_at THEN
        RAISE EXCEPTION 'admin provider route candidate transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION enforce_admin_provider_route_snapshot_update() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.tenant_ref <> OLD.tenant_ref OR NEW.workspace_id <> OLD.workspace_id
       OR NEW.environment <> OLD.environment OR NEW.configuration_id <> OLD.configuration_id
       OR NEW.generation <> OLD.generation + 1 THEN
        RAISE EXCEPTION 'admin provider route snapshot transition is invalid';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION reject_admin_provider_route_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'admin provider route fact is immutable';
END;
$$;

CREATE TRIGGER admin_provider_route_drafts_controlled_update
BEFORE UPDATE ON admin_provider_route_drafts FOR EACH ROW
EXECUTE FUNCTION enforce_admin_provider_route_draft_update();
CREATE TRIGGER admin_provider_route_drafts_no_delete
BEFORE DELETE ON admin_provider_route_drafts FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();

CREATE TRIGGER admin_provider_route_candidates_controlled_update
BEFORE UPDATE ON admin_provider_route_candidates FOR EACH ROW
EXECUTE FUNCTION enforce_admin_provider_route_candidate_update();
CREATE TRIGGER admin_provider_route_candidates_no_delete
BEFORE DELETE ON admin_provider_route_candidates FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();

CREATE TRIGGER admin_provider_route_reviews_no_update
BEFORE UPDATE ON admin_provider_route_reviews FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();
CREATE TRIGGER admin_provider_route_reviews_no_delete
BEFORE DELETE ON admin_provider_route_reviews FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();

CREATE TRIGGER admin_provider_route_snapshots_controlled_update
BEFORE UPDATE ON admin_provider_route_active_snapshots FOR EACH ROW
EXECUTE FUNCTION enforce_admin_provider_route_snapshot_update();
CREATE TRIGGER admin_provider_route_snapshots_no_delete
BEFORE DELETE ON admin_provider_route_active_snapshots FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();

CREATE TRIGGER admin_provider_route_activations_no_update
BEFORE UPDATE ON admin_provider_route_activation_records FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();
CREATE TRIGGER admin_provider_route_activations_no_delete
BEFORE DELETE ON admin_provider_route_activation_records FOR EACH ROW
EXECUTE FUNCTION reject_admin_provider_route_mutation();
