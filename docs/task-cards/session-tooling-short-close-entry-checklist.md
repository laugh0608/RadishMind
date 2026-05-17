# `P2 Session & Tooling` 任务卡：short close entry checklist

更新时间：2026-05-17

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 进入 `P2 short close` 前的只读预检清单。

程序化真相源为 `scripts/checks/fixtures/session-tooling-short-close-entry-checklist.json`，快速门禁为 `scripts/check-session-tooling-short-close-entry-checklist.py`。

当前状态是 `entry_blocked_governance_only`：可以声明 entry checklist 已可检查，且当前 close candidate 仍是 governance-only；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 输入事实源

entry checklist 只从既有治理事实源聚合，不新增真实能力：

- `scripts/checks/fixtures/session-tooling-stop-line-manifest.json`
- `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json`
- `scripts/checks/fixtures/session-tooling-route-smoke-readiness-rollup.json`
- `scripts/checks/fixtures/session-tooling-negative-regression-suite-readiness.json`

## 当前进入条件

以下进入条件全部仍为 `not_satisfied`：

- `upper_layer_confirmation_flow`：confirmation flow design 已可检查，但还没有真实 approve / reject / defer 上层接线。
- `complete_negative_regression_suite`：governance suite、suite readiness、deny-by-default gates 和 route coverage matrix 已存在，但真实 implementation consumer 仍缺失。
- `executor_storage_confirmation_enablement_plan`：executor、storage、confirmation 仍是 `blocked_not_gated_plan`，还不能进入 gated plan。
- `durable_store_and_result_reader_policy`：storage backend 与 result materialization policy 仍停在设计门禁，没有 durable store 或 materialized result reader。

## 当前 route smoke 缺口

以下 future route smoke requirement 全部仍为 `not_satisfied`：

- `executor_gate_route_smoke`
- `storage_materialization_gate_route_smoke`
- `confirmation_gate_route_smoke`
- `durable_store_reader_route_smoke`

## 当前允许声明

- `short_close_entry_checklist_checkable`
- `entry_conditions_identified`
- `entry_still_blocked`
- `governance_only_entry_status`

## 当前禁止声明

- `P2 short close`
- `complete_negative_regression_suite`
- `real_tool_executor_ready`
- `durable_storage_ready`
- `confirmation_flow_connected`
- `materialized_result_reader_ready`
- `automatic_replay_ready`

## 非目标

- 不实现真实工具执行器。
- 不实现 durable session / checkpoint / audit / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不引入长期记忆。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
