ALTER TABLE agent_copilot_runtime_assignments ADD COLUMN action_safety_schema_version TEXT;
ALTER TABLE agent_copilot_runtime_assignments ADD COLUMN action_safety_projection_digest TEXT;
ALTER TABLE agent_copilot_runtime_assignments ADD COLUMN sanitized_action_safety_snapshot TEXT;

ALTER TABLE agent_copilot_runtime_assignment_events ADD COLUMN action_safety_schema_version TEXT;
ALTER TABLE agent_copilot_runtime_assignment_events ADD COLUMN action_safety_projection_digest TEXT;
ALTER TABLE agent_copilot_runtime_assignment_events ADD COLUMN sanitized_action_safety_snapshot TEXT;

ALTER TABLE agent_copilot_run_records ADD COLUMN action_safety_schema_version TEXT;
ALTER TABLE agent_copilot_run_records ADD COLUMN action_safety_projection_digest TEXT;
ALTER TABLE agent_copilot_run_records ADD COLUMN sanitized_action_safety_snapshot TEXT;

ALTER TABLE workflow_http_tool_action_plans ADD COLUMN action_safety_schema_version TEXT;
ALTER TABLE workflow_http_tool_action_plans ADD COLUMN action_safety_projection_digest TEXT;
ALTER TABLE workflow_http_tool_action_plans ADD COLUMN sanitized_action_safety_snapshot TEXT;

ALTER TABLE workflow_run_records ADD COLUMN action_safety_schema_version TEXT;
ALTER TABLE workflow_run_records ADD COLUMN action_safety_projection_digest TEXT;
ALTER TABLE workflow_run_records ADD COLUMN sanitized_action_safety_snapshot TEXT;

CREATE TRIGGER agent_copilot_assignments_action_safety_insert
BEFORE INSERT ON agent_copilot_runtime_assignments
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_assignment_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.decision.effective_level') = 'handoff_ready'
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0)
)
BEGIN SELECT RAISE(ABORT, 'Agent Copilot assignment Action Safety snapshot is invalid'); END;

CREATE TRIGGER agent_copilot_assignments_action_safety_update
BEFORE UPDATE ON agent_copilot_runtime_assignments
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_assignment_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.decision.effective_level') = 'handoff_ready'
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0)
)
BEGIN SELECT RAISE(ABORT, 'Agent Copilot assignment Action Safety snapshot is invalid'); END;

CREATE TRIGGER agent_copilot_assignment_events_action_safety_insert
BEFORE INSERT ON agent_copilot_runtime_assignment_events
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_assignment_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.decision.effective_level') = 'handoff_ready'
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0)
)
BEGIN SELECT RAISE(ABORT, 'Agent Copilot assignment event Action Safety snapshot is invalid'); END;

CREATE TRIGGER agent_copilot_runs_action_safety_insert
BEFORE INSERT ON agent_copilot_run_records
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_run_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.business_writes'), 0) = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.replay_writes'), 0) = 0)
)
BEGIN SELECT RAISE(ABORT, 'Agent Copilot run Action Safety snapshot is invalid'); END;

CREATE TRIGGER agent_copilot_runs_action_safety_update
BEFORE UPDATE ON agent_copilot_run_records
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_run_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.business_writes'), 0) = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.replay_writes'), 0) = 0)
)
BEGIN SELECT RAISE(ABORT, 'Agent Copilot run Action Safety snapshot is invalid'); END;

CREATE TRIGGER workflow_http_tool_plans_action_safety_insert
BEFORE INSERT ON workflow_http_tool_action_plans
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_plan_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.plan_id') = NEW.plan_id
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.tool_plan_digest') = NEW.tool_plan_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.decision.effective_level') <> 'write_blocked')
)
BEGIN SELECT RAISE(ABORT, 'Workflow HTTP Tool plan Action Safety snapshot is invalid'); END;

CREATE TRIGGER workflow_http_tool_plans_action_safety_update
BEFORE UPDATE ON workflow_http_tool_action_plans
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_plan_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.plan_id') = NEW.plan_id
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.tool_plan_digest') = NEW.tool_plan_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.decision.effective_level') <> 'write_blocked')
)
BEGIN SELECT RAISE(ABORT, 'Workflow HTTP Tool plan Action Safety snapshot is invalid'); END;

CREATE TRIGGER workflow_runs_action_safety_insert
BEFORE INSERT ON workflow_run_records
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_run_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.business_writes'), 0) = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.replay_writes'), 0) = 0)
)
BEGIN SELECT RAISE(ABORT, 'Workflow run Action Safety snapshot is invalid'); END;

CREATE TRIGGER workflow_runs_action_safety_update
BEFORE UPDATE ON workflow_run_records
WHEN NOT (
    (NEW.action_safety_schema_version IS NULL AND NEW.action_safety_projection_digest IS NULL AND NEW.sanitized_action_safety_snapshot IS NULL)
    OR (NEW.action_safety_schema_version IS NOT NULL
        AND NEW.action_safety_projection_digest IS NOT NULL
        AND NEW.sanitized_action_safety_snapshot IS NOT NULL
        AND NEW.action_safety_schema_version = 'action_safety_run_projection.v1'
        AND length(NEW.action_safety_projection_digest) = 71
        AND substr(NEW.action_safety_projection_digest, 1, 7) = 'sha256:'
        AND json_valid(NEW.sanitized_action_safety_snapshot)
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.schema_version') = NEW.action_safety_schema_version
        AND json_extract(NEW.sanitized_action_safety_snapshot, '$.projection_digest') = NEW.action_safety_projection_digest
        AND instr(NEW.sanitized_action_safety_snapshot, 'write_allowed_by_policy') = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.business_writes'), 0) = 0
        AND coalesce(json_extract(NEW.sanitized_action_safety_snapshot, '$.side_effects.replay_writes'), 0) = 0)
)
BEGIN SELECT RAISE(ABORT, 'Workflow run Action Safety snapshot is invalid'); END;
