ALTER TABLE workflow_run_records
    DROP CONSTRAINT workflow_runs_action_safety_snapshot_check,
    DROP COLUMN sanitized_action_safety_snapshot,
    DROP COLUMN action_safety_projection_digest,
    DROP COLUMN action_safety_schema_version;

ALTER TABLE workflow_http_tool_action_plans
    DROP CONSTRAINT workflow_http_tool_plans_action_safety_snapshot_check,
    DROP COLUMN sanitized_action_safety_snapshot,
    DROP COLUMN action_safety_projection_digest,
    DROP COLUMN action_safety_schema_version;

ALTER TABLE agent_copilot_run_records
    DROP CONSTRAINT agent_copilot_runs_action_safety_snapshot_check,
    DROP COLUMN sanitized_action_safety_snapshot,
    DROP COLUMN action_safety_projection_digest,
    DROP COLUMN action_safety_schema_version;

ALTER TABLE agent_copilot_runtime_assignment_events
    DROP CONSTRAINT agent_copilot_assignment_events_action_safety_snapshot_check,
    DROP COLUMN sanitized_action_safety_snapshot,
    DROP COLUMN action_safety_projection_digest,
    DROP COLUMN action_safety_schema_version;

ALTER TABLE agent_copilot_runtime_assignments
    DROP CONSTRAINT agent_copilot_assignments_action_safety_snapshot_check,
    DROP COLUMN sanitized_action_safety_snapshot,
    DROP COLUMN action_safety_projection_digest,
    DROP COLUMN action_safety_schema_version;
