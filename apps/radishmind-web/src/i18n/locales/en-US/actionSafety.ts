export const actionSafety = {
  "title": "Action Safety evidence",
  "owner": "Owner",
  "projection": "Projection",
  "decisions": "Decisions",
  "observed": "Observed",
  "calls": "provider {{provider}} · tool {{tool}} · confirmation {{confirmation}}",
  "notRecorded": "Not recorded",
  "notApplicable": "Not applicable",
  "legacy": "This legacy owner has no Action Safety snapshot. Current policy does not backfill historical conclusions.",
  "transient": "This projection accompanies the current response only. Candidate actions remain with the existing review owner and create no new mutation owner.",
  "readOnly": "This read-only projection explains the existing owner. It grants no execution, confirmation or business write permissions.",
  "levels": "{{requested}} → maximum {{maximum}} · {{target}} · {{method}}",
  "confirmation": "Confirmation: {{state}} · policy {{policy}}",
  "blockersLabel": "Blockers",
  "status": {
    "recorded": "Recorded",
    "not_recorded_legacy": "Not recorded in legacy record"
  },
  "ownerKind": {
    "agent_copilot_response": "Agent Copilot response",
    "agent_copilot_runtime_assignment": "Agent Copilot runtime assignment",
    "workflow_http_tool_action_plan": "HTTP Tool action plan",
    "workflow_run": "Workflow run"
  },
  "level": {
    "answer_only": "Answer only",
    "proposal_only": "Proposal only",
    "handoff_ready": "Ready for handoff",
    "tool_callable": "Tool callable",
    "write_blocked": "Writes blocked",
    "write_allowed_by_policy": "Writes allowed by policy"
  },
  "target": {
    "none": "None",
    "human_review_owner": "Human review owner",
    "workflow_http_tool": "Workflow HTTP Tool",
    "business_truth": "Business truth",
    "shell": "Shell",
    "code": "Code",
    "sandbox": "Sandbox",
    "agent_loop": "Agent loop",
    "connector_mutation": "Connector mutation",
    "automatic_apply": "Automatic apply"
  },
  "confirmationState": {
    "not_required": "Not required",
    "pending": "Pending",
    "approved": "Approved",
    "rejected": "Rejected",
    "changed": "Changed"
  },
  "blocker": {
    "action_safety_scope_denied": "Scope denied",
    "action_safety_payload_invalid": "Invalid payload",
    "action_safety_source_unavailable": "Source unavailable",
    "action_safety_source_changed": "Source changed",
    "action_safety_policy_unavailable": "Policy unavailable",
    "action_safety_policy_changed": "Policy changed",
    "action_safety_level_escalation_denied": "Level escalation denied",
    "action_safety_confirmation_required": "Confirmation required",
    "action_safety_confirmation_changed": "Confirmation changed",
    "action_safety_tool_authority_unavailable": "Tool authority unavailable",
    "action_safety_write_blocked": "Writes blocked",
    "action_safety_store_contract_mismatch": "Record contract mismatch"
  }
} as const;
