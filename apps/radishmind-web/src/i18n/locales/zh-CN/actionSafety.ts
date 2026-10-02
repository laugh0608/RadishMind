export const actionSafety = {
  "title": "动作安全证据",
  "owner": "所属记录",
  "projection": "投影版本",
  "decisions": "决定数量",
  "observed": "已观察调用",
  "calls": "提供方 {{provider}} · 工具 {{tool}} · 确认 {{confirmation}}",
  "notRecorded": "未记录",
  "notApplicable": "不适用",
  "legacy": "旧记录未保存动作安全快照；不会使用当前策略回填历史结论。",
  "transient": "此投影仅随当前响应展示；候选动作仍由既有审查记录承接，不创建新的变更记录。",
  "readOnly": "此只读投影仅解释既有记录，不授予执行、确认或业务写入权限。",
  "levels": "{{requested}} → 最高 {{maximum}} · {{target}} · {{method}}",
  "confirmation": "确认：{{state}} · 策略 {{policy}}",
  "blockersLabel": "阻塞项",
  "status": {
    "recorded": "已记录",
    "not_recorded_legacy": "历史记录未保存"
  },
  "ownerKind": {
    "agent_copilot_response": "Agent Copilot 响应",
    "agent_copilot_runtime_assignment": "Agent Copilot 运行绑定",
    "workflow_http_tool_action_plan": "HTTP Tool 动作计划",
    "workflow_run": "工作流运行"
  },
  "level": {
    "answer_only": "仅回答",
    "proposal_only": "仅提议",
    "handoff_ready": "可交接",
    "tool_callable": "可调用工具",
    "write_blocked": "写入受阻",
    "write_allowed_by_policy": "策略允许写入"
  },
  "target": {
    "none": "无",
    "human_review_owner": "人工审查记录",
    "workflow_http_tool": "工作流 HTTP Tool",
    "business_truth": "业务真相源",
    "shell": "Shell",
    "code": "代码",
    "sandbox": "沙箱",
    "agent_loop": "Agent 循环",
    "connector_mutation": "连接器变更",
    "automatic_apply": "自动应用"
  },
  "confirmationState": {
    "not_required": "无需确认",
    "pending": "待确认",
    "approved": "已批准",
    "rejected": "已拒绝",
    "changed": "已变化"
  },
  "blocker": {
    "action_safety_scope_denied": "权限不足",
    "action_safety_payload_invalid": "请求数据无效",
    "action_safety_source_unavailable": "来源不可用",
    "action_safety_source_changed": "来源已变化",
    "action_safety_policy_unavailable": "策略不可用",
    "action_safety_policy_changed": "策略已变化",
    "action_safety_level_escalation_denied": "不允许提高动作级别",
    "action_safety_confirmation_required": "需要确认",
    "action_safety_confirmation_changed": "确认已变化",
    "action_safety_tool_authority_unavailable": "工具授权依据不可用",
    "action_safety_write_blocked": "写入受阻",
    "action_safety_store_contract_mismatch": "记录契约不匹配"
  }
} as const;
