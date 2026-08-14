DO $$
DECLARE payload_constraint text;
BEGIN
    SELECT conname INTO payload_constraint
    FROM pg_constraint
    WHERE conrelid = 'admin_provider_route_drafts'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) LIKE '%schema_version%';
    EXECUTE format('ALTER TABLE admin_provider_route_drafts DROP CONSTRAINT %I', payload_constraint);
END;
$$;

ALTER TABLE admin_provider_route_drafts
    ADD CONSTRAINT admin_provider_route_draft_payload_v2_check CHECK (
        jsonb_typeof(sanitized_draft_payload) = 'object'
        AND sanitized_draft_payload->>'schema_version' IN (
            'admin_provider_route_configuration_draft.v1',
            'admin_provider_route_configuration_draft.v2'
        )
        AND sanitized_draft_payload->>'tenant_ref' = tenant_ref
        AND sanitized_draft_payload->>'workspace_id' = workspace_id
        AND sanitized_draft_payload->>'environment' = environment
        AND sanitized_draft_payload->>'configuration_id' = configuration_id
        AND (sanitized_draft_payload->>'draft_revision')::bigint = draft_revision
        AND sanitized_draft_payload->>'draft_digest' = draft_digest
    );

DO $$
DECLARE payload_constraint text;
BEGIN
    SELECT conname INTO payload_constraint
    FROM pg_constraint
    WHERE conrelid = 'admin_provider_route_candidates'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) LIKE '%schema_version%';
    EXECUTE format('ALTER TABLE admin_provider_route_candidates DROP CONSTRAINT %I', payload_constraint);
END;
$$;

ALTER TABLE admin_provider_route_candidates
    ADD CONSTRAINT admin_provider_route_candidate_payload_v2_check CHECK (
        jsonb_typeof(sanitized_candidate_payload) = 'object'
        AND sanitized_candidate_payload->>'schema_version' IN (
            'admin_provider_route_candidate.v1',
            'admin_provider_route_candidate.v2'
        )
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
    );

DO $$
DECLARE payload_constraint text;
BEGIN
    SELECT conname INTO payload_constraint
    FROM pg_constraint
    WHERE conrelid = 'admin_provider_route_active_snapshots'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) LIKE '%schema_version%';
    EXECUTE format('ALTER TABLE admin_provider_route_active_snapshots DROP CONSTRAINT %I', payload_constraint);
END;
$$;

ALTER TABLE admin_provider_route_active_snapshots
    ADD CONSTRAINT admin_provider_route_snapshot_payload_v2_check CHECK (
        jsonb_typeof(sanitized_snapshot_payload) = 'object'
        AND sanitized_snapshot_payload->>'schema_version' IN (
            'admin_provider_route_snapshot.v1',
            'admin_provider_route_snapshot.v2'
        )
        AND sanitized_snapshot_payload->>'tenant_ref' = tenant_ref
        AND sanitized_snapshot_payload->>'workspace_id' = workspace_id
        AND sanitized_snapshot_payload->>'environment' = environment
        AND sanitized_snapshot_payload->>'configuration_id' = configuration_id
        AND (sanitized_snapshot_payload->>'generation')::bigint = generation
        AND sanitized_snapshot_payload->>'candidate_id' = candidate_id
        AND sanitized_snapshot_payload->>'candidate_digest' = candidate_digest
        AND sanitized_snapshot_payload->>'snapshot_digest' = snapshot_digest
    );
