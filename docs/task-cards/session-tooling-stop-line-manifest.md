# `P2 Session & Tooling` 任务卡：stop-line manifest

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` close candidate 后仍不能放宽的停止线。

程序化真相源为 `scripts/checks/fixtures/session-tooling-stop-line-manifest.json`，快速门禁为 `scripts/check-session-tooling-stop-line-manifest.py`。

当前状态是 `stop_lines_governance_only`，仍是 governance-only：可以声明 hard prerequisite blocker、stop-line capability 和证据来源已经可检查；但不能声明 `P2 short close`、完整 `negative_regression_suite`、真实 executor、durable store、confirmation flow 接线、materialized result reader 或 replay 已完成。

## 当前 hard prerequisite blocker

以下 blocker 仍全部是 `not_satisfied`：

- `upper_layer_confirmation_flow`
- `complete_negative_regression_suite`
- `executor_storage_confirmation_enablement_plan`
- `durable_store_and_result_reader_policy`

## 当前停止线

- 不实现真实工具执行器。
- 不实现 durable session / checkpoint / audit / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不引入长期记忆。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。

## 当前证据来源

- `scripts/checks/fixtures/session-tooling-close-candidate-readiness-rollup.json`
- `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json`
- `scripts/checks/fixtures/session-tooling-executor-storage-confirmation-enablement-plan.json`
- `scripts/checks/fixtures/session-tooling-negative-regression-suite-readiness.json`
- `scripts/checks/fixtures/session-tooling-route-negative-coverage-matrix.json`

## 放宽停止线前必须同步

如果后续要放宽任何停止线，至少必须同步更新对应 fixture、再生成检查、任务卡和入口文档；不能只改实现或只改一份文档。

当前必须联动的检查包括：

- `check-session-tooling-close-candidate-readiness-rollup.py`
- `check-session-tooling-short-close-readiness-delta.py`
- `check-session-tooling-executor-storage-confirmation-enablement-plan.py`
- `check-session-tooling-negative-regression-suite-readiness.py`
- `check-session-tooling-route-negative-coverage-matrix.py`
- `check-session-tooling-stop-line-manifest.py`
