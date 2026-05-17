# `P2 Session & Tooling` 任务卡：upper-layer confirmation flow readiness

更新时间：2026-05-17

## 任务目标

本任务卡固定 `upper_layer_confirmation_flow` 从 design-only 走向可接线前必须补齐的证据清单。

程序化真相源为 `scripts/checks/fixtures/session-tooling-upper-layer-confirmation-flow-readiness.json`，快速门禁为 `scripts/check-session-tooling-upper-layer-confirmation-flow-readiness.py`。

当前状态是 `readiness_defined_not_connected` / governance-only：可以声明上层 confirmation flow 的接线前证据已经可检查，但不能声明 `confirmation_flow_connected`、`confirmed_action_execution`、`P2 short close` 或完整 `negative_regression_suite` 已满足。

## 输入事实源

本 readiness 只聚合现有治理事实，不实现真实上层接线：

- `scripts/checks/fixtures/session-tooling-confirmation-flow-design.json`
- `scripts/checks/fixtures/session-tooling-implementation-preconditions.json`
- `scripts/checks/fixtures/session-tooling-negative-regression-suite.json`
- `scripts/checks/fixtures/session-tooling-short-close-entry-checklist.json`

## 当前 readiness gates

以下 gate 全部仍为 `not_satisfied`：

- `confirmation_request_handoff_contract`：还没有真实上层 request handoff shape。
- `confirmation_decision_binding`：还没有把 approve / reject / defer 决策绑定到独立 audit record。
- `confirmation_negative_gate_consumers`：missing / stale / mismatched confirmation 仍只有 governance consumer，没有真实 gate consumer。
- `confirmed_action_handoff_boundary`：approve outcome 仍只能停在 `approved_pending_execution_boundary`，不能执行 confirmed action。

## 当前允许声明

- `upper_layer_confirmation_readiness_checkable`
- `confirmation_handoff_required_evidence_identified`
- `confirmation_negative_gate_consumers_identified`
- `governance_only_readiness_status`

## 当前禁止声明

- `P2 short close`
- `confirmation_flow_connected`
- `confirmed_action_execution`
- `business_truth_write`
- `automatic_replay`
- `complete_negative_regression_suite`

## 非目标

- 不接真实上层 confirmation flow。
- 不实现 confirmed action execution。
- 不写 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
- 不实现 durable confirmation store。
- 不实现 durable audit store。
- 不启用 automatic replay。
