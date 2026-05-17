# `P2 Session & Tooling` 任务卡：route negative coverage matrix

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 负向 suite case 与 route / future route requirement 的覆盖矩阵。

程序化真相源为 `scripts/checks/fixtures/session-tooling-route-negative-coverage-matrix.json`，快速门禁为 `scripts/check-session-tooling-route-negative-coverage-matrix.py`。

当前状态是 `route_negative_coverage_matrix_governance_only`，仍是 governance-only：可以声明 9 个 negative suite case 与当前 checkpoint read metadata-only route、future executor / storage / confirmation route requirement 的映射已经可检查；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前覆盖矩阵

- 当前 checkpoint read metadata-only route 覆盖 14 个 denied query case。
- 当前 checkpoint read metadata-only route 覆盖 2 个 suite case：`executor-ref-checkpoint-read-denied` 与 `materialized-result-read-denied`。
- 其余 7 个 suite case 仍是 governance-only，需要未来 route / gate smoke 证明后才能进入完整 `negative_regression_suite`。
- `executor_gate_route_smoke`、`storage_materialization_gate_route_smoke`、`confirmation_gate_route_smoke` 与 `durable_store_reader_route_smoke` 当前仍全部为 `not_satisfied`。

## 当前允许声明

- `route_negative_coverage_matrix_checkable`
- `checkpoint_read_route_suite_case_mapping_checkable`
- `future_route_requirement_gaps_checkable`
- `governance_only_route_negative_coverage`

## 当前禁止声明

- `complete_negative_regression_suite`
- `P2 short close`
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
- 不执行 confirmed action。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
