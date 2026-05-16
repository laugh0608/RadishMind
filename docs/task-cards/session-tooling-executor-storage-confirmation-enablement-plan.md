# `P2 Session & Tooling` 任务卡：executor / storage / confirmation enablement plan

更新时间：2026-05-16

## 任务目标

本任务卡固定 executor、storage、confirmation 三类能力从 `not_ready` 进入未来 gated plan 前必须具备的证据。

程序化真相源为 `scripts/checks/fixtures/session-tooling-executor-storage-confirmation-enablement-plan.json`，快速门禁为 `scripts/check-session-tooling-executor-storage-confirmation-enablement-plan.py`。

当前状态是 `enablement_plan_defined_blocked`，仍是 governance-only：可以声明三类能力的 gated plan 入口条件已经可检查；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前结论

- executor 仍是 `blocked_not_gated_plan`，真实 implementation consumer 仍为 `missing`，默认决策仍为 `deny`。
- storage 仍是 `blocked_not_gated_plan`，真实 implementation consumer 仍为 `missing`，默认决策仍为 `deny`。
- confirmation 仍是 `blocked_not_gated_plan`，真实 implementation consumer 仍为 `missing`，默认决策仍为 `deny`。
- `upper_layer_confirmation_flow`、`complete_negative_regression_suite`、`executor_storage_confirmation_enablement_plan` 与 `durable_store_and_result_reader_policy` 仍全部为 `not_satisfied`。

## 进入 gated plan 前必须证明

- executor：执行请求必须在任何 execution / network side effect 前被默认拒绝，并且每次执行尝试的 independent audit 预期明确。
- storage：durable write 与 materialized result read 必须在任何 storage side effect 或 result payload 返回前被默认拒绝，并且 retention / redaction / secret handling 明确。
- confirmation：missing、stale、mismatched confirmation 必须在 action execution 前被拒绝，且 confirmed action handoff 必须与 advisory candidate output 分离。

## 当前允许声明

- `enablement_plan_defined`
- `gated_plan_entry_conditions_checkable`
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
