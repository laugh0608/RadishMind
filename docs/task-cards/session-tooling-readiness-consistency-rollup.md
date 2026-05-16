# `P2 Session & Tooling` 任务卡：readiness consistency rollup

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 多个 readiness / rollup fixture 之间的横向一致性，避免 close-candidate readiness rollup、route smoke readiness rollup、short close readiness delta、negative coverage rollup 和 `negative_regression_suite` readiness 各自更新后发生口径漂移。

程序化真相源为 `scripts/checks/fixtures/session-tooling-readiness-consistency-rollup.json`，快速门禁为 `scripts/check-session-tooling-readiness-consistency-rollup.py`。

当前状态是 `no_readiness_drift_governance_only`：可以声明这些治理层 rollup 的 status、source reference、hard prerequisite、覆盖计数和禁止声明口径当前一致；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前一致性断言

- `foundation status`、`close-candidate readiness rollup`、`route smoke readiness rollup`、`short close readiness delta`、`negative coverage rollup` 与 `negative_regression_suite` readiness 的状态仍保持 governance-only / blocked。
- `upper_layer_confirmation_flow`、`complete_negative_regression_suite`、`executor_storage_confirmation_enablement_plan` 与 `durable_store_and_result_reader_policy` 仍全部为 `not_satisfied`。
- executor、storage、confirmation 三类 implementation area 仍是 `not_ready`，真实 implementation consumer 仍为 `missing`，默认决策仍为 `deny`。
- negative suite 仍是 9 个 case，其中 2 个 route-consumed case、7 个 governance-only case。
- future route smoke requirement 仍是 4 个，满足数量为 0。

## 当前允许声明

- `readiness_drift_checkable`
- `rollup_statuses_aligned`
- `short_close_prerequisites_still_not_satisfied`
- `governance_only_consistency_verified`

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
