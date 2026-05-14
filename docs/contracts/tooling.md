# RadishMind 工具框架契约

更新时间：2026-05-14

## 文档目的

本文档固定 `Tooling Framework` 的首版最小契约。当前目标是把检索、候选生成、validator、adapter 和 policy check 这类 task-local 经验收口为可注册、可审计、可门禁的 tool contract，而不是实现真实工具执行器、长期记忆或新的 provider/model 实验。

Schema 真相源为：

- `contracts/tool.schema.json`
- `contracts/tool-registry.schema.json`
- `contracts/tool-audit-record.schema.json`

最小 fixture 为：

- `scripts/checks/fixtures/tool-registry-basic.json`
- `scripts/checks/fixtures/tool-audit-record-basic.json`

快速门禁为 `scripts/check-tooling-framework-contract.py`，并已接入 `check-repo --fast`。

## 最小结构

`ToolDefinition` 必须表达：

- `tool_id`、`tool_type`、`project_scope`：工具身份、职责类型和适用项目。
- `interface`：输入 / 输出 schema 引用与引用策略。
- `execution`：执行模式、超时和 retry policy；v1 fixture 固定为 `contract_only`，不绑定真实 executor。
- `policy`：advisory-only、不写业务真相源、不启用 durable memory、是否需要人工确认、数据访问范围和风险等级。
- `audit`：是否必须产生日志、是否记录输入输出、是否脱敏 secret。

`ToolRegistry` 必须表达：

- `registry_policy.execution_enabled=false`：当前只登记契约，不启用真实工具执行。
- `registry_policy.durable_memory_enabled=false`：当前不引入长期记忆。
- `registry_policy.network_default=disabled`：当前 registry 不默认允许联网工具。
- `tools`：内嵌的工具定义列表，门禁会逐条按 `ToolDefinition` 校验。

`ToolAuditRecord` 必须表达：

- `tool_id`、`request_id`、`session_id`、`turn_id`：调用轨与会话轨可关联。
- `session_binding`：工具审计是否进入 session recovery checkpoint，以及对应 checkpoint ref。
- `policy_decision`：本次工具调用是否被允许、阻断或需要确认。
- `execution`：v1 只能记录 `not_executed` / `blocked`，不能声称已经执行真实工具。
- `state_landing`：工具状态落点；v1 只允许 `none`、`request_local` 或 `session_recovery_checkpoint`，并强制 `durable_memory_written=false`。
- `result_cache`：工具结果缓存边界；v1 支持 `none`、`metadata_only` 和未来 `result_ref` 形态，但当前 fixture 只固定 metadata-only recovery checkpoint，不缓存真实执行结果。
- `audit`：保持 advisory-only、不写业务真相源、不写 durable memory、secret 已脱敏。

## Session 关联与缓存边界

当前最小策略是：

- session record 的 `state_policy` 决定会话状态与 tool result cache 的落点。
- tool audit record 用 `session_binding` 明确自己是否进入 recovery checkpoint。
- 进入 recovery checkpoint 的 v1 工具状态只记录 metadata、policy decision 和 audit ref，不代表工具已经执行。
- `result_cache.mode=metadata_only` 时不得写 `result_ref`；只有后续真实 executor 边界明确后，才允许引入 materialized result ref。
- 所有状态落点都必须保持 `durable_memory_written=false`。
- recovery checkpoint record / manifest 由 `contracts/session-recovery-checkpoint*.schema.json` 固定；tooling 侧只引用 checkpoint ref，不负责跨轮 replay。

## Promotion 门禁分层

当前 tooling 只能按以下层级晋级：

| 层级 | 当前门禁 | 可声明能力 | 不可声明能力 |
| --- | --- | --- | --- |
| Contract gate | `tool`、`tool-registry`、`tool-audit-record` schema 与 fixture | 工具定义、registry policy、session binding、metadata-only audit/cache 边界稳定 | 真实工具执行、真实结果缓存、durable tool store |
| Checkpoint read gate | `session-recovery-checkpoint-read` fixture 与 denied query fixture | checkpoint read 可暴露 tool audit summary，且只暴露治理元数据 | 返回 tool output、result ref、executor ref 或 materialized result |
| Platform route smoke | checkpoint metadata-only route 与负向查询参数 smoke | 平台能拒绝把 tool audit summary 升级成执行或结果读取请求 | executor、replay、长期记忆或业务真相源写入 |
| Future implementation gate | 上层确认流、executor sandbox、result materialization policy 和独立审计记录明确后再定义 | 可讨论受控工具执行器和短期 result ref | 在当前 registry v1 后直接启用 unrestricted tool calling |

## 当前停止线

- 不实现真实工具执行器。
- 不引入长期记忆或 durable session/tool store。
- 不新增 provider/model 实验。
- 不让工具直接写 `RadishFlow`、`Radish` 或其他上层业务真相源。
- 不把 candidate builder 输出当作已确认动作；高风险候选动作仍必须由上层确认流承接。
- 不让 checkpoint read route 返回 tool output、result ref、executor ref 或 materialized tool result。
