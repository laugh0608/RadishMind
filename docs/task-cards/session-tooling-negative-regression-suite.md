# `P2 Session & Tooling` 任务卡：negative regression suite

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 负向回归 suite 的治理级消费口径。

程序化真相源为 `scripts/checks/fixtures/session-tooling-negative-regression-suite.json`，快速门禁为 `scripts/check-session-tooling-negative-regression-suite.py`。

当前状态是 `governance_suite_consumed_implementation_gates_missing`：9 个 skeleton case 已有 governance-only 消费者、audit non-write 边界断言和 forbidden side-effect absence 断言，但 executor、storage、confirmation 的 implementation gate 仍未落地，因此不能声明完整 `negative_regression_suite` 已完成。

## 当前已覆盖

- `executor_blocked`：消费 executor disabled、network disabled 和 checkpoint read 中 `executor_ref` denied。
- `storage_materialization_blocked`：消费 materialized result read、durable memory write 和 business truth write denied。
- `confirmation_blocked`：消费 missing confirmation、stale confirmation 和 mismatched confirmation payload denied。

每个 case 都必须同时固定：

- 明确治理消费者。
- audit non-write 或未来 audit event 来源。
- 不发生 execution、network、storage、materialized result、business truth write 或 replay side effect。
- 对应 implementation gate 仍为 `missing_deny_by_default_implementation_gate`。

## 当前仍然阻塞

- executor、storage、confirmation implementation gate 还不存在。
- 上层 confirmation flow 尚未接线。
- durable store 与 materialized result reader 仍未启用。
- close-candidate rollup 仍必须保持 governance-only，不能声明 `P2 short close`。

## 非目标

- 不实现真实工具执行器。
- 不实现 durable session / checkpoint / audit / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。
