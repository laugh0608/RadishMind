# `P2 Session & Tooling` 任务卡：route smoke readiness rollup

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 当前 route smoke 覆盖与进入 `P2 short close` 前仍缺的 route / gate smoke 要求。

程序化真相源为 `scripts/checks/fixtures/session-tooling-route-smoke-readiness-rollup.json`，快速门禁为 `scripts/check-session-tooling-route-smoke-readiness-rollup.py`。

当前状态是 `route_smoke_readiness_governance_only`，也就是 governance-only：可以声明 checkpoint read metadata-only route smoke 已覆盖，并且 route smoke 缺口已可检查；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前已覆盖

- `checkpoint_read_metadata_only`：`/v1/session/recovery/checkpoints/{checkpoint_id}` 只覆盖 fixture-backed metadata-only response shape。
- 该 route 已验证 materialized result、result ref、executor ref、durable memory 与 replay 类查询会被拒绝。
- 当前 route smoke 仍不读取真实 checkpoint store，不返回 materialized result，不触发 automatic replay。

## 仍缺的 route / gate smoke

以下 requirement 当前全部为 `not_satisfied`：

- `executor_gate_route_smoke`：未来真实 executor 入口必须先证明默认拒绝发生在任何 execution / network side effect 前。
- `storage_materialization_gate_route_smoke`：未来 storage / result reader 入口必须先证明 materialized result、durable write 和业务真相源写入都被拒绝。
- `confirmation_gate_route_smoke`：未来 confirmation 接线必须先证明 missing / stale / mismatched confirmation 都在 action execution 前被拒绝。
- `durable_store_reader_route_smoke`：未来 durable reader route 必须先证明 redaction、retention、secret handling 和 no replay / no business write 边界。

## 当前允许声明

- `checkpoint_read_metadata_route_smoke_covered`
- `route_smoke_gap_checkable`
- `route_smoke_readiness_governance_only`

## 当前禁止声明

- `P2 short close`
- `complete_negative_regression_suite`
- `real_tool_executor_ready`
- `durable_storage_ready`
- `confirmation_flow_connected`
- `materialized_result_reader_ready`
- `automatic_replay_ready`

## 非目标

- 不新增真实 executor route。
- 不新增 durable store reader route。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
