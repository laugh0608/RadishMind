# `P2 Session & Tooling` 任务卡：negative regression suite readiness

更新时间：2026-05-16

## 任务目标

本任务卡固定完整 `negative_regression_suite` 的进入条件和验收维度。

程序化真相源为 `scripts/checks/fixtures/session-tooling-negative-regression-suite-readiness.json`，快速门禁为 `scripts/check-session-tooling-negative-regression-suite-readiness.py`。该 readiness 还消费 `scripts/checks/fixtures/session-tooling-negative-coverage-rollup.json` 与 `scripts/checks/fixtures/session-tooling-route-negative-coverage-matrix.json`，并对齐 `scripts/checks/fixtures/session-tooling-short-close-readiness-delta.json`，用于确认当前负向覆盖仍只是 governance-only rollup，完整 suite 仍是 `not_satisfied`，不是 P2 short close 或完整 suite 完成。

当前状态是 `governance_suite_consumed_deny_by_default_gates_defined`：可以声明 suite 验收口径已定义、governance-only suite 已被消费，且 executor / storage / confirmation 三类 deny-by-default implementation gate contract 已对齐；但不能声明完整 `negative_regression_suite` 已满足，也不能据此解除 executor、storage 或 confirmation 的 `not_ready` 状态。

## 必须覆盖的负向组

- `executor_blocked`：覆盖 executor disabled、network disabled 和 checkpoint read 中 `executor_ref` denied。
- `storage_materialization_blocked`：覆盖 materialized result read、durable memory write 和 business truth write denied。
- `confirmation_blocked`：覆盖 missing confirmation、stale confirmation 和 mismatched confirmation payload denied。

这些组来自现有 `scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json`，当前仍只是 skeleton。

## 完整 suite 验收要求

当前仍为 `not_satisfied`：

- `real_consumers_before_completion`：每个负向 case 必须被真实 contract、route、policy 或 implementation gate 消费，不能只存在于 fixture。
- `independent_audit_assertions`：每个 blocked action 必须验证预期 audit event，或明确验证当前不会写 audit store。
- `side_effect_absence_assertions`：每个 case 必须证明没有发生 execution、storage、materialized result、business truth write 或 replay side effect。
- `implementation_gate_alignment`：executor、storage、confirmation 的 deny-by-default gate contract 已存在且默认 deny；真实实现 consumer 仍未满足。
- `negative_coverage_rollup_alignment`：route smoke、fixture consumer、governance suite case 与 deny-by-default gate 覆盖关系已可检查；真实 implementation consumer 仍未满足。
- `route_negative_coverage_matrix_alignment`：当前 checkpoint read metadata-only route 只覆盖 2 个 suite case，其余 7 个仍是 future route required 且 requirement 为 `not_satisfied`。
- `short_close_delta_alignment`：short close delta 已把 `complete_negative_regression_suite` 继续标记为 `not_satisfied`，避免 readiness 被误写成 P2 short close。

`checkpoint_denied_query_alignment` 当前只被已有 denied query fixture 部分满足，不代表完整 suite 完成。

## 当前禁止

- 不实现真实工具执行器。
- 不实现 durable session / checkpoint / audit / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。

## 何时可以声明完整 suite

只有在以下条件同时满足后，才能把当前 readiness 推进为完整 `negative_regression_suite`：

1. 每个 skeleton case 都有明确消费者。
2. 每个 case 都验证 audit 或 audit non-write 边界。
3. 每个 case 都验证禁止 side effect。
4. executor、storage、confirmation implementation gate contract 已存在，并由 suite 覆盖 deny-by-default 行为。
5. 真实 executor、storage、confirmation consumer 能证明 deny 行为先于任何 side effect 生效。
6. close-candidate rollup 仍保持 governance-only，直到其它 short close 前置条件同时满足。
