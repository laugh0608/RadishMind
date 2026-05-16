# `P2 Session & Tooling` 任务卡：short close readiness delta

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 从当前 close candidate 到 `P2 short close` 之间仍缺的最小硬前置条件。

程序化真相源为 `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json`，快速门禁为 `scripts/check-session-tooling-short-close-readiness-delta.py`。

当前状态是 `short_close_blocked`：可以声明 short close delta 已经可检查，且当前 close candidate 仍是 governance-only；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前 delta

以下硬前置条件全部仍为 `not_satisfied`：

- `upper_layer_confirmation_flow`：confirmation flow design 已可检查，但还没有真实 approve / reject / defer 上层接线。
- `complete_negative_regression_suite`：governance suite、suite readiness、deny-by-default gates 和 negative coverage rollup 已存在，但真实 implementation consumer 仍缺失。
- `executor_storage_confirmation_enablement_plan`：executor、storage、confirmation 仍全部是 `not_ready`，不能被当前治理资产直接解锁。
- `durable_store_and_result_reader_policy`：storage backend 与 result materialization policy 仍停在设计门禁，没有 durable store 或 materialized result reader。

## 当前允许声明

- `close_candidate_governance_only`
- `short_close_delta_checkable`
- `hard_prerequisites_identified`
- `implementation_still_blocked`

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
