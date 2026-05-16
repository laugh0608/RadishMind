# `P2 Session & Tooling` 任务卡：close-candidate readiness rollup

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` close candidate 的统一盘点口径。

程序化真相源为 `scripts/checks/fixtures/session-tooling-close-candidate-readiness-rollup.json`，快速门禁为 `scripts/check-session-tooling-close-candidate-readiness-rollup.py`。

当前 rollup 只声明 `close_candidate_governance_only`，即 governance-only：design gate 可检查，metadata smoke 可复验，负向回归已有 skeleton，但仍不是 `P2 short close`，也不是真实 executor、durable store、confirmation flow、materialized result reader 或 replay 已实现。

## 汇总范围

rollup 汇总以下已落地的 P2 治理资产：

- confirmation flow design
- independent audit records design
- result materialization policy design
- executor boundary design
- storage backend design
- negative regression skeleton
- implementation preconditions

这些资产只把当前边界收口到可检查的设计层，不解除任何实现阻塞。

## 当前允许声明

- `contract_and_metadata_smoke_ready`
- `design_gates_checkable`
- `negative_regression_skeleton_exists`
- `close_candidate_governance_only`

## 当前禁止声明

- `P2 short close`
- `real_tool_executor_ready`
- `durable_storage_ready`
- `confirmation_flow_connected`
- `materialized_result_reader_ready`
- `automatic_replay_ready`
- `complete_negative_regression_suite`

## short close 前置条件

当前仍为 `not_satisfied`：

- `upper_layer_confirmation_flow`：上层还没有真实 approve / reject / defer 接线。
- `complete_negative_regression_suite`：当前只有 skeleton，不能证明真实实现边界。
- `executor_storage_confirmation_enablement_plan`：executor、storage、confirmation 仍全部是 `not_ready`。
- `durable_store_and_result_reader_policy`：durable store 与 materialized result reader 仍未实现或启用。

## 非目标

- 不实现真实工具执行器。
- 不实现 durable session / checkpoint / audit / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
