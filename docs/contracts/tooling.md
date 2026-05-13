# RadishMind 工具框架契约

更新时间：2026-05-13

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
- `policy_decision`：本次工具调用是否被允许、阻断或需要确认。
- `execution`：v1 只能记录 `not_executed` / `blocked`，不能声称已经执行真实工具。
- `audit`：保持 advisory-only、不写业务真相源、不写 durable memory、secret 已脱敏。

## 当前停止线

- 不实现真实工具执行器。
- 不引入长期记忆或 durable session/tool store。
- 不新增 provider/model 实验。
- 不让工具直接写 `RadishFlow`、`Radish` 或其他上层业务真相源。
- 不把 candidate builder 输出当作已确认动作；高风险候选动作仍必须由上层确认流承接。
