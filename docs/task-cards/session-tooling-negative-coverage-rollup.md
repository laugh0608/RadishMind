# `P2 Session & Tooling` 任务卡：negative coverage rollup

更新时间：2026-05-16

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 当前负向覆盖关系的统一盘点口径。

程序化真相源为 `scripts/checks/fixtures/session-tooling-negative-coverage-rollup.json`，快速门禁为 `scripts/check-session-tooling-negative-coverage-rollup.py`。

当前状态是 `negative_coverage_rollup_governance_only`：可以声明 checkpoint read route smoke、denied query fixture、negative consumption summary、governance-only negative suite 和 deny-by-default implementation gate contract 之间的覆盖关系已可检查；但不能声明完整 `negative_regression_suite`、`P2 short close` 或真实 executor / storage / confirmation 已实现。

## 覆盖层

当前 rollup 汇总五层：

- checkpoint read route smoke：只覆盖 fixture-backed metadata-only route 和 denied query。
- negative consumption summary：确认已有负向 fixture 与 summary 有消费者。
- governance negative suite：确认 9 个 skeleton case 已有治理消费者、audit non-write 边界和 side-effect absence 断言。
- deny-by-default implementation gates：确认 executor、storage、confirmation 三类 gate contract 默认拒绝。
- real implementation consumers：当前明确为 `missing`。

## 当前覆盖事实

- denied query fixture 覆盖 14 个 checkpoint read query case。
- negative suite 覆盖 9 个 executor / storage / confirmation skeleton case。
- 其中 2 个 suite case 已经由 checkpoint read route/denied query 侧覆盖。
- 其余 7 个 suite case 仍只属于 governance-only 与 gate contract 覆盖。
- executor、storage、confirmation 三类真实 implementation consumer 仍全部缺失。

## 当前允许声明

- `metadata_only_route_smoke_covered`
- `negative_fixture_consumers_covered`
- `governance_suite_gate_contract_alignment_covered`
- `negative_coverage_rollup_governance_only`

## 当前禁止声明

- `complete_negative_regression_suite`
- `P2 short close`
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
- 不执行 confirmed action。
- 不写入 `RadishFlow`、`Radish` 或 `RadishCatalyst` 业务真相源。

## 后续进入完整 suite 前

只有在真实 executor、storage、confirmation consumer 出现，并能证明 deny-by-default 行为先于 side effect 生效后，才允许把当前 rollup 前推为完整 `negative_regression_suite` 的实现级覆盖证据。
