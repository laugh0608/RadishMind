# Agent / Copilot 开发测试态使用指南

更新时间：2026-07-25

## 适用范围

本文说明如何在 RadishMind Platform 的开发测试态创建 Agent / Copilot Profile、生成不可变版本、绑定应用配置、审查发布候选、显式管理 Runtime Assignment，并通过 API key 或 Application Interaction Session 发起一次受控建议。

长期设计与 owner 边界见[Agent / Copilot 应用档案版本审查与受控建议专题](agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)，schema 真相源位于 `contracts/`，HTTP 实现真相源位于 `services/platform/internal/httpapi/`。本指南只描述已经实现的开发测试态使用方式，不把 assignment、成功调用或评测批准解释为生产发布。

## 资源与操作顺序

一条可调用的 Agent / Copilot 配置必须按以下顺序建立：

1. Application Catalog 中存在未归档、类型为 `agent` 的应用。
2. Agent Copilot Profile owner 校验并保存结构化草案，再从精确 `draft_version` 创建不可变版本。
3. Application Configuration Draft owner 通过专用 binding 路由重读 Profile Version，生成 `application_configuration_draft.v4`。
4. Application Publish Candidate owner 从精确配置生成 `application_publish_candidate.v4`；审查人读取 Profile source 并显式批准。
5. Agent Copilot Runtime Assignment owner 通过 `activate` 或 `replace` 绑定当前 approved candidate；撤销使用 `revoke`。
6. 调用方签发只包含 `agent_copilot:invoke` 的开发测试态 API key，或创建 `agent_copilot_suggestion_v1` Application Session。
7. 调用方提交 canonical task、locale、artifacts、context 和幂等键；服务端重读完整 authority 后最多委托一次 Gateway。
8. 当前响应只在同步返回或当前 Web 内存中可见；Run History、Comparison、Evaluation 与 Operations 只读取 metadata-only Run v7 证据。

这些步骤不能合并为自动动作。Profile Version 不会修改配置，binding 不会创建候选，候选批准不会建立 assignment，assignment 决策不会调用 Gateway，历史与评测读取也不会重新执行。

## 启动与存储模式

SQLite 本地产品链使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --agent-copilot-local-product
```

PostgreSQL 开发测试档使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --agent-copilot-postgres-dev-test
```

两档都会启用 Application Catalog、Configuration Draft v4、Publish Candidate v4、Agent Profile / Runtime、API Key、Gateway Playground / History、Application Session v3、Run History / Comparison / Evaluation 与 Operations。launcher 只开启 owner、迁移检查和 strict Web consumer，不自动创建任何应用、Profile、候选、审查、assignment、API key 或 invocation。

当前 Agent 完整产品档只由 Shell launcher 提供；`scripts/run-radishmind-web-dev.ps1` 尚未暴露 `AgentCopilotLocalProduct` 或 `AgentCopilotPostgresDevTest` 参数。PowerShell 用户不得使用不存在的开关；在补齐对称入口前，应在支持 Shell 的环境运行上述产品档，或按 Platform `configured` 模式显式提供全部 gate 与 store 配置。

Profile store 支持以下模式：

| 模式 | 用途 | 约束 |
| --- | --- | --- |
| `memory_dev` | 单元测试、短进程验证 | 进程退出即丢失，不作为恢复证据 |
| 聚合 `sqlite_dev` | 本地连续开发 | 只通过 `RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev` 统一启用 |
| `postgres_dev_test` | migration、角色、方言与并发验证 | 显式 runtime / migration DSN；任何失败不得回退内存 |

Agent 相关配置键：

| 配置键 | 作用 |
| --- | --- |
| `RADISHMIND_AGENT_COPILOT_PROFILE_DEV_HTTP` | 开放 Profile validate / read / list / version HTTP surface |
| `RADISHMIND_AGENT_COPILOT_PROFILE_DEV_WRITE` | 允许保存草案和创建版本；依赖 Profile HTTP gate |
| `RADISHMIND_AGENT_COPILOT_PROFILE_STORE` | `memory_dev | postgres_dev_test`；聚合 SQLite 不在这里选择 |
| `RADISHMIND_AGENT_COPILOT_PROFILE_DEV_TEST_DATABASE_URL` | PostgreSQL runtime DML 连接 |
| `RADISHMIND_AGENT_COPILOT_PROFILE_DEV_TEST_MIGRATION_DATABASE_URL` | PostgreSQL migration 连接 |
| `RADISHMIND_AGENT_COPILOT_PROFILE_DATABASE_TIMEOUT` | Profile 数据库操作超时 |
| `RADISHMIND_AGENT_COPILOT_RUNTIME_DEV_HTTP` | 开放 assignment 与 invocation；依赖 auth、catalog、draft、publish、Profile owner |
| `RADISHMIND_AGENT_COPILOT_RUNTIME_DEV_WRITE` | 允许 assignment decision；依赖 Runtime HTTP gate |

Profile 使用独立 `0001_agent_copilot_profiles` migration family。Assignment / Event 与 Session / Turn / Run 投影复用共享 Workflow Run Store：SQLite 使用 `0014_agent_copilot_runtime_assignments`、`0015_agent_copilot_invocation_projections`，PostgreSQL 使用 `0017_agent_copilot_runtime_assignments`、`0018_agent_copilot_invocation_projections`。PostgreSQL runner 统一使用：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-workflow-saved-draft-postgres-dev-test.sh status
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

runner 不打印或提交 DSN。验证结束后应执行 `down`。

## 开发身份与权限

Profile 和 Runtime Assignment 管理路由使用现有开发身份头：

```text
X-RadishMind-Dev-Read-Identity: agent-guide
X-RadishMind-Dev-Read-Tenant: tenant_demo
X-RadishMind-Dev-Read-Subject: subject_owner
X-RadishMind-Dev-Read-Scopes: <逗号分隔的所需 scope>
X-RadishMind-Dev-Read-Audit: audit_agent_guide
```

Profile 路由还必须携带：

```text
X-RadishMind-Dev-Agent-Copilot-Profile-Workspace: workspace_demo
X-RadishMind-Dev-Agent-Copilot-Profile-Application: app_aaaaaaaaaaaaaaaa
```

Runtime Assignment 路由使用：

```text
X-RadishMind-Dev-Agent-Copilot-Runtime-Workspace: workspace_demo
X-RadishMind-Dev-Agent-Copilot-Runtime-Application: app_aaaaaaaaaaaaaaaa
```

body、query、path 和资源绑定头中的 workspace / application 必须一致。Configuration Draft binding 与 Publish Candidate 继续使用各自既有资源绑定头。

| scope | 允许的操作 |
| --- | --- |
| `agent_copilot_profiles:read` | 列出草案 / 版本脱敏摘要 |
| `agent_copilot_profiles:read_source` | 读取精确草案或版本源码；候选源码审查需要该权限 |
| `agent_copilot_profiles:write` | 执行确定性 validate 与草案 CAS 保存 |
| `agent_copilot_profiles:version` | 从精确有效草案创建不可变版本 |
| `agent_copilot_profiles:bind` | 与 `application_drafts:write` 组合，允许配置 owner 绑定 Profile Version |
| `agent_copilot_runtime:read` | 读取当前 assignment 和只追加事件 |
| `agent_copilot_runtime:write` | 执行 `activate | replace | revoke` |
| `agent_copilot:invoke` | API key 专用调用 scope，只能调用 key 所属当前应用 |
| `application_sessions:read | write | execute` | 管理 Session v3 并委托同一受控建议服务 |

上游权限使用 `radishmind.agent-copilot-profiles.*` 与 `radishmind.agent-copilot-runtime.*`，服务端只投影到对应本地 scope，不隐式授予应用、发布或调用权限。

## Profile API

| 方法与路径 | scope | 说明 |
| --- | --- | --- |
| `POST /v1/user-workspace/agent-copilot-profiles/validate` | `agent_copilot_profiles:write` | 只校验和编译，不写 owner |
| `POST /v1/user-workspace/agent-copilot-profiles` | `agent_copilot_profiles:write` | 使用 `expected_draft_version` 保存草案 |
| `GET /v1/user-workspace/agent-copilot-profiles` | `agent_copilot_profiles:read` | query 只允许 `workspace_id`、`application_id` |
| `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}` | `agent_copilot_profiles:read_source` | 读取完整结构化草案 |
| `POST /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions` | `agent_copilot_profiles:version` | 从 `source_draft_version` 创建不可变版本 |
| `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions` | `agent_copilot_profiles:read` | 返回版本摘要 |
| `GET /v1/user-workspace/agent-copilot-profiles/{profile_id}/versions/{profile_version}` | `agent_copilot_profiles:read_source` | 读取精确版本源码 |

草案保存示例：

```json
{
  "expected_draft_version": 0,
  "profile": {
    "schema_version": "agent_copilot_profile_draft.v1",
    "profile_id": "acpf_aaaaaaaaaaaaaaaa",
    "workspace_id": "workspace_demo",
    "application_id": "app_aaaaaaaaaaaaaaaa",
    "profile_name": "RadishFlow diagnostics advisor",
    "description": "Review diagnostics and return advisory candidate actions.",
    "project": "radishflow",
    "allowed_tasks": ["explain_diagnostics", "suggest_flowsheet_edits"],
    "default_locale": "zh-CN",
    "allowed_locales": ["en-US", "zh-CN"],
    "context_policy": {
      "allowed_fields": ["selected_unit_ids", "diagnostics"],
      "max_bytes": 65536,
      "require_task_context": true
    },
    "artifact_policy": {
      "allowed_kinds": ["json", "text"],
      "allowed_roles": ["primary", "supporting"],
      "max_count": 8,
      "max_item_bytes": 65536,
      "max_total_bytes": 131072
    },
    "response_policy": {
      "allowed_action_kinds": ["candidate_edit", "read_only_check"],
      "max_answers": 8,
      "max_issues": 16,
      "max_actions": 8,
      "max_citations": 16,
      "max_visible_text_bytes": 8192
    },
    "risk_policy": {
      "mode": "advisory",
      "requires_confirmation_for_actions": true,
      "confirmation_action_kinds": ["candidate_edit", "candidate_operation", "ghost_completion"]
    },
    "tool_hints_policy": {
      "allow_retrieval": false,
      "allow_tool_calls": false,
      "allow_image_reasoning": false
    }
  }
}
```

服务端会规范 task、locale、context field、artifact kind / role 与 action kind 顺序，并计算 `profile_digest`、`policy_digest` 和 `allowed_tasks_digest`。客户端不得提交或覆盖这些 digest。

Profile source 最大 `64 KiB`；context 最大 `128 KiB`；单个 artifact 最大 `128 KiB`，artifact 总量最大 `256 KiB`。Profile 不接受 credential、token、header、endpoint、DSN、自由 system prompt、provider / model / runtime 配置。`risk_policy.mode` 固定为 `advisory`，候选动作必须确认，三个 tool hint 必须全部为 `false`。

## 配置、候选与 Runtime Assignment

Configuration Draft v4 通过以下专用路由绑定精确版本：

```text
POST /v1/user-workspace/application-configuration-drafts/{draft_id}/agent-copilot-profile-binding
```

body 只包含 `workspace_id`、`application_id`、`expected_draft_version`、`profile_id` 和 `profile_version`。服务端重读 Profile Version 并保存 ref / digest，不复制 source。

Publish Candidate v4 继续使用既有 `/v1/user-workspace/application-publish-candidates*` 路由和人工 review 状态机。审查时必须有 `agent_copilot_profiles:read_source`；Profile 不可读、digest 漂移、应用类型变化或 task 不兼容都会阻止批准或运行资格。

Runtime Assignment 路由：

| 方法与路径 | 说明 |
| --- | --- |
| `GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment` | 读取当前 assignment 并重验 authority |
| `GET /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/events` | 读取只追加事件 |
| `POST /v1/user-workspace/applications/{application_id}/agent-copilot-runtime-assignment/decisions` | 执行 expected-version CAS decision |

`activate | replace` 必须提交 approved candidate id；`revoke` 的 `candidate_id` 必须为空。已撤销 assignment 不能原地恢复，后续启用必须重新 `activate` 并形成新的 assignment。

## 受控建议与 Session

API key 调用路由：

```text
POST /v1/agent-copilot/invocations
Authorization: Bearer <只展示一次的开发测试态 API key>
```

请求 body 只允许：

```json
{
  "task": "suggest_flowsheet_edits",
  "locale": "zh-CN",
  "conversation_id": "conversation_demo",
  "artifacts": [],
  "context": {
    "selected_unit_ids": ["unit_demo"],
    "diagnostics": []
  },
  "client_invocation_key": "agent-invocation-demo-001"
}
```

调用方不能提交 project、Profile、assignment、candidate、provider、model、credential、tool hint 或重试策略。服务端从 exact authority 恢复 project / policy，把三个 canonical tool hint 固定为 `false`，把 safety 固定为 advisory，并在 Gateway 前再次重读 Application、assignment、candidate、Configuration Draft、Profile Version 与模型资格。

Session 使用既有 Application Session 路由，创建时指定：

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_aaaaaaaaaaaaaaaa",
  "execution_profile": "agent_copilot_suggestion_v1"
}
```

Turn v3 的 `input_text` 必须为空，Agent 输入改用 `task`、`locale`、`conversation_id`、`artifacts` 与 `context`；Session 只委托同一 invocation service。相同幂等键不会恢复历史完整回答，也不会触发第二次 Gateway。

## 运行证据与排障

成功调用写入 `workflow_run_record.v7`，只保存 Profile lineage、project / task / locale、输入摘要与字节数、artifact / context 计数、模型选择、响应摘要、usage、诊断和副作用计数。`output` 固定为空；原始 context、artifact content、完整 `CopilotResponse`、transcript 和 provider raw response 不落盘。

常见失败处理：

| 现象 | 检查项 |
| --- | --- |
| Profile route disabled / write disabled | 检查 Profile HTTP / write gate，不要通过改客户端绕过 |
| `scope_denied` | 对齐身份 scope、workspace / application header、query、path 与 body |
| project / task 或 policy invalid | 从 canonical contract 选择任务和 context field；不要在 Web 单独扩枚举 |
| binding / candidate / authority changed | 重新读取 Profile Version、配置、候选和 assignment，不复用旧 digest |
| invocation duplicate running | 等待当前调用终态；不要换幂等键触发并行 provider 调用 |
| canceled / outcome unknown | 读取 Run v7 诊断；不得自动 replay |
| response contract failed | 修正 Gateway / provider 的 canonical response，不降级保存未校验文本 |

Web 在应用切换、revision 变化、阶段卸载或请求迟到时会清空未提交输入和当前回答。API key 原文、context、完整响应不得进入 URL、`localStorage`、`sessionStorage` 或日志。

## 停止线

- 不实现 agent loop、自治规划、多轮自驱、工具执行、检索执行、connector、在线搜索或业务写回。
- 不自动应用 `proposed_actions`；候选动作保持 `requires_confirmation=true`。
- 不实现 provider retry / fallback、自动 activation / release、schedule、replay / resume、长期记忆、quota、billing 或生产声明。
- 不把 `agent_copilot:invoke`、SQLite / PostgreSQL 开发测试证据或 Web 连续链解释为生产认证、生产 API key 或真实 Radish 集成。
