# `P2 Session & Tooling` 任务卡：实现前置条件

更新时间：2026-05-14

## 任务目标

本任务卡固定 `P2 Session & Tooling Foundation` 从 contract / metadata smoke 进入真实实现前必须满足的条件。

当前目标不是实现 executor、durable store、长期记忆或 replay，而是把 `executor`、`storage`、`confirmation` 三类缺口拆成可审计、可检查、可阻断的前置条件。

程序化真相源为 `scripts/checks/fixtures/session-tooling-implementation-preconditions.json`，快速门禁为 `scripts/check-session-tooling-implementation-preconditions.py`。

负向回归 skeleton 为 `scripts/checks/fixtures/session-tooling-negative-regression-skeleton.json`，快速门禁为 `scripts/check-session-tooling-negative-regression-skeleton.py`。它只固定 blocked executor、blocked storage 和 blocked confirmation 的失败样本骨架，不代表 negative_regression_suite 已满足。

Confirmation flow design 为 `scripts/checks/fixtures/session-tooling-confirmation-flow-design.json`，快速门禁为 `scripts/check-session-tooling-confirmation-flow-design.py`。它只定义 approve / reject / defer、stale / mismatched / missing confirmation、审计事件和禁止绕过规则，不代表上层确认流已经接入。

Independent audit records design 为 `scripts/checks/fixtures/session-tooling-independent-audit-records-design.json`，快速门禁为 `scripts/check-session-tooling-independent-audit-records.py`。它只定义审计记录形状、事件来源和 confirmation / executor / storage 职责分离，不代表 durable audit store 已经存在。

Result materialization policy design 为 `scripts/checks/fixtures/session-tooling-result-materialization-policy-design.json`，快速门禁为 `scripts/check-session-tooling-result-materialization-policy.py`。它只定义 `metadata_only`、未来 `result_ref` 和未来 materialized result 的分层边界，不代表 materialized result reader 或 durable result store 已经存在。

## 当前已完成门禁

当前可以声明的能力只到以下层级：

- session record、tool registry、tool audit record 和 checkpoint read schema 已有 fixture 与 contract check。
- checkpoint read route 已有 fixture-backed metadata-only smoke。
- denied query fixture 已覆盖 materialized result、result ref、executor ref、durable memory 和 replay 类请求。
- promotion gate、negative consumption summary、route smoke coverage summary 和 readiness summary 已进入 `check-repo --fast`。
- confirmation flow、independent audit records 和 result materialization policy 已有设计级 fixture 与 check，但仍未接入真实实现。

这些门禁只说明 contract 和 metadata smoke ready，不说明真实 executor、durable storage、confirmation flow 或 replay 已经存在。

## Executor 前置条件

当前状态：`not_ready`。

当前边界：`contract_only_no_execution`。

启用真实 executor 前必须满足：

- `executor_boundary`：执行沙箱、allowlist、输入输出 envelope、timeout/retry 和失败边界明确。
- `result_materialization_policy`：真实执行结果是否可保存、如何引用、何时脱敏和何时禁止返回明确。
- `independent_audit_records`：每一次执行尝试都能生成独立 audit record，不依赖 checkpoint read response 伪装审计。
- `negative_regression_suite`：覆盖 blocked execution、blocked network、blocked business truth writes 和 blocked replay。

当前禁止：

- 不实现 `real_tool_executor`。
- 不启用 unrestricted tool calling。
- 不在 checkpoint read response 中返回 `executor_ref`。
- 不让工具写入 `RadishFlow`、`Radish` 或其它上层业务真相源。

## Storage 前置条件

当前状态：`not_ready`。

当前边界：`metadata_refs_only_no_durable_store`。

启用 durable storage 前必须满足：

- `storage_backend_design`：session、checkpoint、audit 和 materialized result 的职责边界分开。
- `result_materialization_policy`：默认 metadata-only，任何 result ref 都必须有显式策略。
- `independent_audit_records`：durable write 与 audit write 分离，避免用单一持久化记录同时冒充状态和审计。
- `negative_regression_suite`：覆盖 durable memory disabled、result ref disabled 和 business truth write disabled。

当前禁止：

- 不新增 durable session store。
- 不新增 durable tool store。
- 不新增长期记忆。
- 不实现 materialized result reader。
- 不把 fixture-backed checkpoint route 写成真实 checkpoint storage backend。

## Confirmation 前置条件

当前状态：`not_ready`。

当前边界：`requires_confirmation_metadata_only`。

启用确认后动作承接前必须满足：

- `upper_layer_confirmation_flow`：上层明确拥有 approve、reject、defer 三类结果。
- `result_materialization_policy`：确认后的动作、执行结果和引用策略明确。
- `independent_audit_records`：确认事件、执行尝试和状态变化分别可审计。
- `negative_regression_suite`：覆盖 missing confirmation、stale confirmation 和 mismatched action payload。

当前禁止：

- 不做隐式确认。
- 不启用 automatic replay。
- 不执行 confirmed action。
- 不写上层业务真相源。

## 何时进入实现设计

只有以下条件同时满足，才允许从本任务卡进入 executor / storage / confirmation 的真实设计：

1. 上层项目给出真实确认流或等价的只读承接边界。
2. executor sandbox、storage backend、materialization policy 和 independent audit record 均有明确任务卡或契约草案。
3. 负向回归先于实现落地，并能证明越界执行、越界读取、越界写入和 replay 都会失败。
4. `scripts/checks/fixtures/session-tooling-readiness-summary.json` 仍保持当前 metadata-only 门禁通过，且不把 readiness 改写成实现完成。

## 负向回归 skeleton

当前已经固定三组 skeleton case，但仍只属于治理骨架：

- blocked executor：覆盖 executor disabled、network disabled 和 checkpoint read 中 `executor_ref` denied。
- blocked storage：覆盖 materialized result read、durable memory write 和 business truth write denied。
- blocked confirmation：覆盖 missing confirmation、stale confirmation 和 mismatched confirmation payload denied。

这些 case 只用于提前固定失败边界、错误码和 forbidden outputs。它们不创建真实执行入口、不创建持久化后端、不读取 materialized result，也不执行确认后的动作。

## 非目标

- 不实现真实工具执行器。
- 不实现 durable session store、durable tool store 或长期记忆。
- 不实现 materialized result reader。
- 不实现 replay executor。
- 不新增真实 provider/model 实验。
- 不修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部工作区。
