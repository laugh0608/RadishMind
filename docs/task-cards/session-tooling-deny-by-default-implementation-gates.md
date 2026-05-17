# `P2 Session & Tooling` 任务卡：deny-by-default implementation gates

更新时间：2026-05-16

## 任务目标

本任务卡固定 executor、storage、confirmation 三类真实实现入口之前的最小 deny-by-default gate contract。

程序化真相源为 `scripts/checks/fixtures/session-tooling-deny-by-default-implementation-gates.json`，快速门禁为 `scripts/check-session-tooling-deny-by-default-implementation-gates.py`。

当前状态是 `deny_by_default_gates_defined_implementation_blocked`：可以声明三类 implementation gate 的最小契约已可检查，默认决策均为 deny，并且已被 governance-only negative regression suite 对齐；但这不是 executor、durable storage 或 confirmation flow 的真实实现。

## Gate 范围

### executor

默认拒绝：

- registered tool execution request
- network tool execution request
- checkpoint read 中暴露 `executor_ref`

当前必须返回或对齐的拒绝码包括：

- `TOOL_EXECUTOR_DISABLED`
- `TOOL_NETWORK_DISABLED`
- `CHECKPOINT_MATERIALIZED_RESULTS_DISABLED`

### storage

默认拒绝：

- materialized result read
- durable memory write
- business truth write

当前必须返回或对齐的拒绝码包括：

- `CHECKPOINT_MATERIALIZED_RESULTS_DISABLED`
- `CHECKPOINT_DURABLE_MEMORY_DISABLED`
- `BUSINESS_TRUTH_WRITE_DISABLED`

### confirmation

默认拒绝：

- missing confirmation
- stale confirmation
- mismatched confirmation payload

当前必须返回或对齐的拒绝码包括：

- `CONFIRMATION_REQUIRED`
- `CONFIRMATION_STALE`
- `CONFIRMATION_PAYLOAD_MISMATCH`

## 共同不变量

- fail closed：缺 policy、缺 confirmation 或缺实现时默认拒绝。
- 不在显式 enablement 前产生 execution、network、storage、business truth write 或 replay side effect。
- checkpoint read 不返回 materialized result、result ref、output ref 或 executor ref。
- 模型层和 tooling 层都不直接写 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。

## 与负向回归的关系

`session-tooling-negative-regression-suite.json` 现在可以声明每个负向 case 已对齐到对应 deny-by-default gate contract。

这只说明 gate contract 可检查，不说明完整 `negative_regression_suite` 完成。完整 suite 仍需要真实 contract、route、policy 或 implementation consumer 证明 deny 行为在真实执行入口前生效。

## 当前禁止

- 不实现真实工具执行器。
- 不实现 durable store。
- 不实现 durable audit / checkpoint / result store。
- 不实现 materialized result reader。
- 不接上层 confirmation flow。
- 不启用 automatic replay。
- 不执行 confirmed action。
- 不写上层业务真相源。

## 后续进入真实实现前

只有以下条件同时满足，才允许把 gate contract 前推到真实实现：

1. 上层确认流或等价只读承接边界存在。
2. executor、storage、confirmation 的真实入口都有独立 consumer。
3. negative regression suite 能验证真实入口仍默认拒绝。
4. audit 记录与 side-effect absence 断言都不依赖 checkpoint read response 伪装。
5. close-candidate rollup 仍不声明 `P2 short close`，直到其它前置条件也满足。
