# `P2 Session & Tooling` 任务卡：Executor Boundary

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 的 `executor_boundary` 设计边界。

当前目标不是实现真实工具执行器、shell 执行、网络工具、durable store、materialized result reader、长期记忆、业务写回或 replay，而是先定义未来执行器必须满足的 sandbox、allowlist、execution envelope、timeout / retry 和失败边界。

程序化真相源为 `scripts/checks/fixtures/session-tooling-executor-boundary-design.json`，快速门禁为 `scripts/check-session-tooling-executor-boundary.py`。

## 当前边界

当前状态：`design_only_not_connected`。

当前 executor 只允许表达设计级 metadata：

- `execution_enabled=false`
- `sandbox_enabled=false`
- `allowlist_enforced=false`
- `network_enabled=false`
- `result_materialization_enabled=false`
- `writes_business_truth=false`
- `replay_enabled=false`

因此本任务卡不会创建真实执行入口，不会启动后台进程，不会运行工具，不会写上层业务真相源。

## Execution Envelope

未来执行请求 envelope 至少需要：

- `execution_request_id`
- `tool_id`
- `request_id`
- `session_id`
- `turn_id`
- `action_ref`
- `action_hash`
- `confirmation_ref`
- `policy_decision_ref`
- `input_ref`
- `output_policy`
- `audit_record_ref`

当前明确禁止在 envelope 或 checkpoint read 中出现：

- `executor_ref`
- `process_id`
- `stdout`
- `stderr`
- `raw_tool_output`
- `result_ref`
- `materialized_result_uri`

执行输入输出 envelope 不能混同于模型 prompt 文本，checkpoint read response 也不能冒充 execution envelope。

## Sandbox Policy

当前 sandbox 仍是 `not_implemented`。

启用前至少需要：

- `process_or_worker_isolation`
- `filesystem_scope_policy`
- `network_policy`
- `secret_injection_policy`
- `resource_limits`
- `kill_timeout_policy`

当前禁止 shell execution、workspace 外写入、网络访问、secret 访问、GPU / 长时间任务、后台进程和业务真相源写入。

## Allowlist Policy

当前 allowlist 仍是 `not_implemented`，默认决策是 `deny`。

启用前至少需要：

- `tool_id_allowlist`
- `project_scope_allowlist`
- `risk_level_policy`
- `network_scope_policy`
- `write_scope_policy`
- `confirmation_requirement_policy`

当前没有任何工具被允许真实执行。unregistered tool、network tool、business truth write tool、durable store write tool、replay tool 和 long running job 全部保持 blocked。

## Timeout / Retry

当前 timeout / retry 只做设计级约束。

启用前至少需要：

- `per_tool_timeout_ms`
- `cancellation_boundary`
- `idempotency_key_policy`
- `retry_budget_policy`
- `partial_failure_audit_policy`

当前禁止 automatic retry、background retry、未经重新校验的确认后 retry 和 replay retry。

## 失败边界

当前必须保留三类失败边界：

- `executor-disabled-tool-run`：返回 `TOOL_EXECUTOR_DISABLED`，不执行工具，不返回 `executor_ref` 或 `result_ref`。
- `executor-network-disabled`：返回 `TOOL_NETWORK_DISABLED`，不发送网络请求，不执行工具。
- `executor-ref-checkpoint-read-denied`：返回 `CHECKPOINT_MATERIALIZED_RESULTS_DISABLED`，checkpoint read 不返回 executor ref、raw output、result ref 或 materialized result URI。

这些 case 仍是 design-level gate，不代表完整 `negative_regression_suite` 已满足。

## 与 Confirmation 的边界

未来高风险执行 envelope 必须携带 `confirmation_ref`，但确认批准不能绕过 executor policy gate 或 sandbox checks。

missing、stale 或 mismatched confirmation 必须先于执行失败。

## 与 Result Materialization 的边界

当前 `result_ref_enabled=false`。

executor output 不能在当前范围内创建 `result_ref`、raw tool output 或 `materialized_result_uri`。未来结果物化必须先通过独立的 `result_materialization_policy` 门禁。

## 与 Audit 的边界

未来每次执行尝试都必须在 policy decision 前后产生独立审计记录。

当前只允许记录 blocked / not_executed metadata。审计记录不保存 raw tool output 或 credential。

## 与 Storage 的边界

executor boundary 不创建 durable session store、durable tool store、durable audit store 或 durable result store。

业务真相源写入仍然禁止。任何持久化结果都必须等待 storage backend design 明确。

## 非目标

- 不实现真实 executor。
- 不运行 shell 或其它本地命令。
- 不启用网络工具。
- 不创建 `executor_ref`、`result_ref` 或 materialized result。
- 不实现 durable store。
- 不实现长期记忆。
- 不写 `RadishFlow`、`Radish` 或其它上层业务真相源。
- 不启用 automatic replay。
