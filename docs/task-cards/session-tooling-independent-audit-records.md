# `P2 Session & Tooling` 任务卡：Independent Audit Records

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 的 `independent_audit_records` 设计边界。

当前目标不是实现 durable audit store、真实 executor、result materialization、长期记忆、业务写回或 replay，而是先定义独立审计记录的最小形状、事件来源，以及它与 confirmation / executor / storage 的边界。

程序化真相源为 `scripts/checks/fixtures/session-tooling-independent-audit-records-design.json`，快速门禁为 `scripts/check-session-tooling-independent-audit-records.py`。

## 当前边界

当前状态：`design_only_not_connected`。

独立审计记录当前只允许表达设计级 metadata：

- 哪个事件发生了。
- 事件来自 confirmation、executor policy gate、storage policy gate、checkpoint read route、tool registry policy 或 northbound request。
- 事件关联哪个 request、session、turn、action、confirmation、tool audit 或 checkpoint。
- 本次边界决策是否仍保持不执行、不写业务真相源、不产生 materialized result、不写 durable memory、不启动 replay。

这不是 durable audit store，也不是 checkpoint read response 的替代品。checkpoint read 可以引用 audit ref 或治理摘要，但不得返回 raw audit payload、真实工具输出或 materialized result。

## Audit Record Shape

独立审计记录必须至少包含：

- `audit_record_id`
- `event_type`
- `event_source`
- `occurred_at`
- `request_id`
- `session_id`
- `turn_id`
- `subject_ref`
- `actor_ref`
- `correlation_refs`
- `payload_hash`
- `boundary_decision`
- `side_effects`
- `redaction`
- `storage_policy`

`payload_hash` 用于稳定关联事件 payload，不把 raw tool output、credential 或其它敏感输入直接写入审计 fixture。

## Event Sources

当前固定六类事件来源：

- `confirmation_flow`：记录 `confirmation_requested / recorded / rejected / deferred / invalidated`，但 `approve` 也不执行动作。
- `executor_policy_gate`：记录未来 executor attempt 的 blocked / not enabled 决策；当前不暴露 `executor_ref`，不执行工具。
- `storage_policy_gate`：记录 materialization、durable memory 和 business truth write 的拒绝边界。
- `checkpoint_read_route`：记录 metadata read served 与 query denied；只用于关联审计引用，不返回 raw audit payload。
- `tool_registry_policy`：记录 registry / tool policy 决策和 network denied；当前 registry 仍是 `execution_enabled=false`。
- `northbound_request`：记录请求进入与 failure boundary；不写上层业务真相源。

## 与 Confirmation 的边界

confirmation record 可以携带 `audit_ref`，但 `audit_ref` 只是证据元数据，不是 approval engine。

`approve / reject / defer` 仍由上层确认流承接；在 `upper_layer_confirmation_flow` 未实现前，即使出现 `approve`，状态也只能停在 `approved_pending_execution_boundary`，不能变成 `execution_status=executed`。

## 与 Executor 的边界

当前 executor 状态仍是 `not_implemented`，`execution_enabled=false`。

后续如果实现真实 executor，必须在执行前后产生独立审计记录，并把 policy decision、confirmation、action hash 和 result materialization policy 分开记录。本任务卡不创建 executor sandbox、allowlist、真实执行入口或 retry 语义。

## 与 Storage 的边界

当前审计记录是 `fixture_only`，不实现 durable audit store。

session store、checkpoint store、audit store 和 materialized result store 必须保持职责分离。checkpoint read route 可以引用 audit summary，但不得成为 audit storage、materialized result reader 或 replay executor。

## 仍未满足的前置条件

完成本设计后，以下前置条件仍保持未满足：

- `upper_layer_confirmation_flow`
- `executor_boundary`
- `storage_backend_design`
- `result_materialization_policy`
- `negative_regression_suite`

`independent_audit_records` 只达到 `design_boundary_defined_not_implemented`，不代表真实独立审计记录、durable audit store 或 production retention policy 已完成。

## 非目标

- 不实现真实 executor。
- 不执行 confirmed action。
- 不实现 durable audit store。
- 不实现 durable session store 或 durable tool store。
- 不实现长期记忆。
- 不返回 materialized result 或 result ref。
- 不写 `RadishFlow`、`Radish` 或其它上层业务真相源。
- 不启用 automatic replay。
