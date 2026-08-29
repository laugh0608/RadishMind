ALTER TABLE agent_copilot_runtime_assignments
    ADD COLUMN action_safety_schema_version text,
    ADD COLUMN action_safety_projection_digest text,
    ADD COLUMN sanitized_action_safety_snapshot jsonb,
    ADD CONSTRAINT agent_copilot_assignments_action_safety_snapshot_check CHECK (
        (action_safety_schema_version IS NULL AND action_safety_projection_digest IS NULL AND sanitized_action_safety_snapshot IS NULL)
        OR (num_nonnulls(action_safety_schema_version, action_safety_projection_digest, sanitized_action_safety_snapshot) = 3
            AND
            action_safety_schema_version = 'action_safety_assignment_projection.v1'
            AND action_safety_projection_digest ~ '^sha256:[a-f0-9]{64}$'
            AND jsonb_typeof(sanitized_action_safety_snapshot) = 'object'
            AND sanitized_action_safety_snapshot->>'schema_version' = action_safety_schema_version
            AND sanitized_action_safety_snapshot->>'projection_digest' = action_safety_projection_digest
            AND sanitized_action_safety_snapshot#>>'{decision,effective_level}' = 'handoff_ready'
            AND sanitized_action_safety_snapshot#>>'{decision,requested_level}' <> 'write_allowed_by_policy'
        )
    );

ALTER TABLE agent_copilot_runtime_assignment_events
    ADD COLUMN action_safety_schema_version text,
    ADD COLUMN action_safety_projection_digest text,
    ADD COLUMN sanitized_action_safety_snapshot jsonb,
    ADD CONSTRAINT agent_copilot_assignment_events_action_safety_snapshot_check CHECK (
        (action_safety_schema_version IS NULL AND action_safety_projection_digest IS NULL AND sanitized_action_safety_snapshot IS NULL)
        OR (num_nonnulls(action_safety_schema_version, action_safety_projection_digest, sanitized_action_safety_snapshot) = 3
            AND
            action_safety_schema_version = 'action_safety_assignment_projection.v1'
            AND action_safety_projection_digest ~ '^sha256:[a-f0-9]{64}$'
            AND jsonb_typeof(sanitized_action_safety_snapshot) = 'object'
            AND sanitized_action_safety_snapshot->>'schema_version' = action_safety_schema_version
            AND sanitized_action_safety_snapshot->>'projection_digest' = action_safety_projection_digest
            AND sanitized_action_safety_snapshot#>>'{decision,effective_level}' = 'handoff_ready'
            AND sanitized_action_safety_snapshot#>>'{decision,requested_level}' <> 'write_allowed_by_policy'
        )
    );

ALTER TABLE agent_copilot_run_records
    ADD COLUMN action_safety_schema_version text,
    ADD COLUMN action_safety_projection_digest text,
    ADD COLUMN sanitized_action_safety_snapshot jsonb,
    ADD CONSTRAINT agent_copilot_runs_action_safety_snapshot_check CHECK (
        (action_safety_schema_version IS NULL AND action_safety_projection_digest IS NULL AND sanitized_action_safety_snapshot IS NULL)
        OR (num_nonnulls(action_safety_schema_version, action_safety_projection_digest, sanitized_action_safety_snapshot) = 3
            AND
            action_safety_schema_version = 'action_safety_run_projection.v1'
            AND action_safety_projection_digest ~ '^sha256:[a-f0-9]{64}$'
            AND jsonb_typeof(sanitized_action_safety_snapshot) = 'object'
            AND sanitized_action_safety_snapshot->>'schema_version' = action_safety_schema_version
            AND sanitized_action_safety_snapshot->>'projection_digest' = action_safety_projection_digest
            AND jsonb_typeof(sanitized_action_safety_snapshot->'decisions') = 'array'
            AND NOT jsonb_path_exists(sanitized_action_safety_snapshot, '$.decisions[*] ? (@.effective_level == "write_allowed_by_policy")')
            AND COALESCE((sanitized_action_safety_snapshot#>>'{side_effects,business_writes}')::bigint, 0) = 0
            AND COALESCE((sanitized_action_safety_snapshot#>>'{side_effects,replay_writes}')::bigint, 0) = 0
        )
    );

ALTER TABLE workflow_http_tool_action_plans
    ADD COLUMN action_safety_schema_version text,
    ADD COLUMN action_safety_projection_digest text,
    ADD COLUMN sanitized_action_safety_snapshot jsonb,
    ADD CONSTRAINT workflow_http_tool_plans_action_safety_snapshot_check CHECK (
        (action_safety_schema_version IS NULL AND action_safety_projection_digest IS NULL AND sanitized_action_safety_snapshot IS NULL)
        OR (num_nonnulls(action_safety_schema_version, action_safety_projection_digest, sanitized_action_safety_snapshot) = 3
            AND
            action_safety_schema_version = 'action_safety_plan_projection.v1'
            AND action_safety_projection_digest ~ '^sha256:[a-f0-9]{64}$'
            AND jsonb_typeof(sanitized_action_safety_snapshot) = 'object'
            AND sanitized_action_safety_snapshot->>'schema_version' = action_safety_schema_version
            AND sanitized_action_safety_snapshot->>'projection_digest' = action_safety_projection_digest
            AND sanitized_action_safety_snapshot->>'plan_id' = plan_id
            AND sanitized_action_safety_snapshot->>'tool_plan_digest' = tool_plan_digest
            AND sanitized_action_safety_snapshot#>>'{decision,effective_level}' NOT IN ('write_blocked', 'write_allowed_by_policy')
            AND sanitized_action_safety_snapshot#>>'{decision,requested_level}' <> 'write_allowed_by_policy'
        )
    );

ALTER TABLE workflow_run_records
    ADD COLUMN action_safety_schema_version text,
    ADD COLUMN action_safety_projection_digest text,
    ADD COLUMN sanitized_action_safety_snapshot jsonb,
    ADD CONSTRAINT workflow_runs_action_safety_snapshot_check CHECK (
        (action_safety_schema_version IS NULL AND action_safety_projection_digest IS NULL AND sanitized_action_safety_snapshot IS NULL)
        OR (num_nonnulls(action_safety_schema_version, action_safety_projection_digest, sanitized_action_safety_snapshot) = 3
            AND
            action_safety_schema_version = 'action_safety_run_projection.v1'
            AND action_safety_projection_digest ~ '^sha256:[a-f0-9]{64}$'
            AND jsonb_typeof(sanitized_action_safety_snapshot) = 'object'
            AND sanitized_action_safety_snapshot->>'schema_version' = action_safety_schema_version
            AND sanitized_action_safety_snapshot->>'projection_digest' = action_safety_projection_digest
            AND jsonb_typeof(sanitized_action_safety_snapshot->'decisions') = 'array'
            AND NOT jsonb_path_exists(sanitized_action_safety_snapshot, '$.decisions[*] ? (@.effective_level == "write_allowed_by_policy")')
            AND COALESCE((sanitized_action_safety_snapshot#>>'{side_effects,business_writes}')::bigint, 0) = 0
            AND COALESCE((sanitized_action_safety_snapshot#>>'{side_effects,replay_writes}')::bigint, 0) = 0
        )
    );
