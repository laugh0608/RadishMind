# Prompt Application 开发测试态使用指南

更新时间：2026-07-25

## 适用范围

本文说明如何在 RadishMind Platform 的开发测试态创建 Prompt Application 模板、生成不可变版本、绑定应用配置、审查发布候选、显式管理当前 Runtime Assignment，并通过 API key 或 Application Interaction Session 发起受控调用。设计边界与字段职责见[提示词应用模板版本审查与受控调用专题](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)，schema 真相源位于 `contracts/`，HTTP 实现真相源位于 `services/platform/internal/httpapi/`。

当前已提供模板与 assignment 管理、Prompt Application invocation、Application Session / Turn v2、Run v6、History / Comparison / Evaluation / Operations metadata 消费，以及 Application Development Workspace 下的 Prompt Web 工作区。Run / Session 的 memory、SQLite 与真实 PostgreSQL 行为门禁均已通过；Web 单一连续浏览器链仍在批次 E 收口中。Runtime Assignment 的 `active` 只表示某个已批准候选被显式选为当前运行 authority，不表示 provider 已被调用，也不构成生产发布。

## 资源与操作顺序

一条可用的开发测试态配置必须按以下 owner 顺序建立：

1. Application Catalog 中存在未归档、类型为 `prompt_application` 的应用。
2. Prompt Template owner 保存有效草案，并从精确 `draft_version` 生成不可变版本。
3. Application Configuration Draft owner 通过专用 binding 路由重读该模板版本，生成 `application_configuration_draft.v3`。
4. Application Publish Candidate owner 从精确配置草案生成 `application_publish_candidate.v3`，审查人读取模板源码并显式批准。
5. Prompt Runtime Assignment owner 通过 `activate` 或 `replace` 绑定当前 approved candidate；撤销使用 `revoke`。
6. 调用方使用独立 API key scope，或创建显式 Prompt profile 的 Application Session，提交变量和幂等键。
7. 用户通过 Run History、Comparison、Evaluation 与 Operations 复验 metadata-only 证据；完整输出只存在于首次同步响应中。

这些步骤不能合并为自动动作。模板版本创建不会修改配置草案，配置绑定不会创建候选，候选批准不会创建 assignment，assignment 决策不会调用 Gateway 或 provider，历史 / 评测读取不会重新执行。

## 启动与存储模式

Prompt Web 与 Platform 的完整 SQLite 链优先使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --prompt-application-local-product
```

Windows / PowerShell 使用：

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -PromptApplicationLocalProduct
```

PostgreSQL 开发测试档使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --prompt-application-postgres-dev-test
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -PromptApplicationPostgresDevTest
```

SQLite 档自动选择聚合 `sqlite_dev`，将 Prompt Template 作为第八个独立持久化组件，并在共享 Workflow Run Store 中保存 Runtime Assignment / Event、Session / Turn v2 与 Run v6 投影；PostgreSQL 档使用 `configured` profile，在启动前检查 Application Catalog、Configuration Draft、Publish Candidate、API Key、Gateway Request、Prompt Template 和共享 Workflow Run migration marker，不自动执行迁移。两档都会开启对应 strict Web consumer，但不会自动创建应用、模板、candidate、review、assignment、API key 或 invocation。

显式 `configured` 档支持以下 Template store：

| 模式 | 用途 | 关键要求 |
| --- | --- | --- |
| `memory_dev` | 单元测试、短进程验证 | 进程退出即丢失，不作为恢复证据 |
| 聚合 `sqlite_dev` | 本地产品连续开发 | 只通过 `RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev` 启用，不与组件 `*_STORE` 混用 |
| `postgres_dev_test` | migration、角色、方言与并发同构验证 | 显式 Template runtime DSN、独立 migration DSN、完整开发 gate；数据库故障不得回退内存 |

Prompt 相关显式配置键如下：

| 配置键 | 作用 |
| --- | --- |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_HTTP` | 开放 Template validate / read / list / version HTTP surface |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_WRITE` | 允许保存草案和创建版本；依赖 Template HTTP gate |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_STORE` | `memory_dev | postgres_dev_test`；聚合 SQLite 不在这里选择 |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_TEST_DATABASE_URL` | PostgreSQL runtime DML 连接，只用于服务运行 |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DEV_TEST_MIGRATION_DATABASE_URL` | PostgreSQL migration 连接，只用于 migration runner |
| `RADISHMIND_PROMPT_APPLICATION_TEMPLATE_DATABASE_TIMEOUT` | Template 数据库操作超时 |
| `RADISHMIND_PROMPT_APPLICATION_RUNTIME_DEV_HTTP` | 开放 assignment read / events / decisions；依赖 auth、draft、publish、template HTTP gate |
| `RADISHMIND_PROMPT_APPLICATION_RUNTIME_DEV_WRITE` | 允许 assignment 决策；依赖 Runtime HTTP gate |

开发 gate 默认关闭。`config-check` 和 `diagnostics` 只做启动前检查，不创建 SQLite 文件或执行 migration；只有 `serve` 进入聚合 runtime 生命周期。PostgreSQL 专项统一使用：

```bash
./scripts/run-workflow-saved-draft-postgres-dev-test.sh check
./scripts/run-workflow-saved-draft-postgres-dev-test.sh status
./scripts/run-workflow-saved-draft-postgres-dev-test.sh down
```

该入口会同时验证独立 Prompt Template `0001_prompt_application_templates` 与 Workflow Run `0016_prompt_application_runtime_projections`。验证结束后应执行 `down`；runtime DSN 与 migration DSN 不得互换、输出或提交。

## 开发身份、作用域与权限

以下示例假定 `dev_headers` 身份模式。每个请求都需要基础身份头：

```text
X-RadishMind-Dev-Read-Identity: prompt-guide
X-RadishMind-Dev-Read-Tenant: tenant_demo
X-RadishMind-Dev-Read-Subject: subject_owner
X-RadishMind-Dev-Read-Scopes: <逗号分隔的所需 scope>
X-RadishMind-Dev-Read-Audit: audit_prompt_guide
```

Template 路由还必须携带：

```text
X-RadishMind-Dev-Prompt-Template-Workspace: workspace_demo
X-RadishMind-Dev-Prompt-Template-Application: app_aaaaaaaaaaaaaaaa
```

Runtime Assignment 路由使用独立资源绑定头：

```text
X-RadishMind-Dev-Prompt-Runtime-Workspace: workspace_demo
X-RadishMind-Dev-Prompt-Runtime-Application: app_aaaaaaaaaaaaaaaa
```

Configuration Draft binding 与 Publish Candidate 继续使用各自既有的 Application Draft / Publish 资源绑定头。body、query、path 和资源绑定头中的 workspace / application 必须一致；任一不一致都按 scope denied 失败，不能只依赖客户端提交的 ID。

权限保持分离：

| scope | 允许的操作 |
| --- | --- |
| `prompt_application_templates:read` | 列出草案 / 版本脱敏摘要；不读取源码 |
| `prompt_application_templates:read_source` | 读取精确草案或不可变版本源码；候选源码审查和 approve 需要该权限 |
| `prompt_application_templates:write` | 执行只读 validate 与草案 CAS 保存 |
| `prompt_application_templates:version` | 从精确有效草案创建不可变版本 |
| `prompt_application_templates:bind` | 与 `application_drafts:write` 组合，允许配置 owner 绑定模板版本 |
| `prompt_application_runtime:read` | 读取当前 assignment 或只追加事件 |
| `prompt_application_runtime:write` | 执行 `activate | replace | revoke` 决策 |
| `prompt_application:invoke` | API key 专用调用 scope，只能调用 key 所属当前应用 |
| `application_sessions:read | write | execute` | 管理 Session v2 并通过 Prompt profile 委托同一 invocation service |

上游权限采用 `radishmind.prompt-application-templates.*` 与 `radishmind.prompt-application-runtime.*`，服务端只投影到上表的本地 scope，不隐式授予 Application Catalog、Publish Review 或更宽的写权限。

## Template API

### 路由

| 方法与路径 | scope | 说明 |
| --- | --- | --- |
| `POST /v1/user-workspace/prompt-application-templates/validate` | `prompt_application_templates:write` | 只做确定性校验，不写 owner；不要求 write gate |
| `POST /v1/user-workspace/prompt-application-templates` | `prompt_application_templates:write` | 使用 `expected_draft_version` 保存草案 |
| `GET /v1/user-workspace/prompt-application-templates` | `prompt_application_templates:read` | query 仅允许 `workspace_id`、`application_id`，返回脱敏摘要 |
| `GET /v1/user-workspace/prompt-application-templates/{template_id}` | `prompt_application_templates:read_source` | 读取草案源码 |
| `POST /v1/user-workspace/prompt-application-templates/{template_id}/versions` | `prompt_application_templates:version` | 从 `source_draft_version` 创建不可变版本 |
| `GET /v1/user-workspace/prompt-application-templates/{template_id}/versions` | `prompt_application_templates:read` | 返回版本摘要，不返回 messages |
| `GET /v1/user-workspace/prompt-application-templates/{template_id}/versions/{template_version}` | `prompt_application_templates:read_source` | 读取精确版本源码 |

所有 JSON body 拒绝未知字段。列表和详情拒绝未声明 query 参数，避免把 provider、credential 或非 owner filter 偷渡到读取面。

### 草案示例

`schema_version` 固定为 `prompt_application_template_draft.v1`。`template_id` 使用 `ptpl_` 加 16 位小写 base32 短键；创建时 `expected_draft_version=0`，后续保存必须提交当前版本。

```json
{
  "expected_draft_version": 0,
  "template": {
    "schema_version": "prompt_application_template_draft.v1",
    "template_id": "ptpl_aaaaaaaaaaaaaaaa",
    "workspace_id": "workspace_demo",
    "application_id": "app_aaaaaaaaaaaaaaaa",
    "template_name": "支持问题摘要",
    "description": "按指定语气概括支持问题",
    "messages": [
      {"role": "system", "content": "请使用 {{ tone }} 的语气。"},
      {"role": "user", "content": "问题：{{ question }}"}
    ],
    "variables": [
      {"name": "question", "type": "string", "required": true, "description": "用户问题"},
      {"name": "tone", "type": "string", "required": false, "description": "回答语气", "default_value": "清晰"}
    ],
    "output_contract": {"kind": "text", "allow_empty": false, "max_bytes": 4096}
  }
}
```

模板只支持 `{{ variable_name }}` 一轮替换。角色限制为 `system | developer | user`；变量类型限制为 `string | integer | number | boolean | string_list`；输出契约限制为 `text | json_object`。当前预算包括最多 16 条消息、64 个变量、单条消息 16 KiB、源码 64 KiB、渲染结果 128 KiB、输出 64 KiB。表达式、函数、循环、条件、include、属性访问、环境变量、文件或网络访问均不支持。

validate / save 响应使用 `validation_summary.state`、`is_valid` 和 `findings[]` 表达所有确定性问题。业务失败通常仍返回受观测 JSON envelope；调用方必须同时检查 HTTP 状态与 `failure_code`，不能把 HTTP 200 单独解释为成功。

### 创建不可变版本

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_aaaaaaaaaaaaaaaa",
  "source_draft_version": 1
}
```

只有精确草案版本仍存在、校验有效且 digest 一致时才能创建版本。同一 `source_draft_version` 只能创建一次不可变版本，重复请求返回 immutable conflict，不得覆盖已有内容；后续修改必须保存新草案版本，再创建新模板版本。

## 配置绑定与发布审查

配置绑定使用：

```text
POST /v1/user-workspace/application-configuration-drafts/{draft_id}/prompt-template-binding
```

请求需要 `application_drafts:write,prompt_application_templates:bind`，并提交：

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_aaaaaaaaaaaaaaaa",
  "expected_draft_version": 1,
  "template_id": "ptpl_aaaaaaaaaaaaaaaa",
  "template_version": 1
}
```

客户端不提交 `template_digest` 或模板正文。服务端重读精确 Template Version、计算 ref / digest，并通过既有草案 CAS 生成下一版 `application_configuration_draft.v3`。非 `prompt_application`、同时存在 Workflow RAG binding、版本漂移或源码 store 不可用都会失败关闭。

Publish Candidate 继续复用 `/v1/user-workspace/application-publish-candidates*`。创建 v3 candidate 时需要既有 `application_publish_candidates:write` 与 `prompt_application_templates:read_source`；approve 时需要既有 review scope 与源码读取权限。服务端会重读应用、精确 draft、Template Version、digest 和作用域。批准只改变 candidate 审查状态，不修改配置草案，也不创建 Runtime Assignment。

## Runtime Assignment API

| 方法与路径 | scope | 说明 |
| --- | --- | --- |
| `GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment?workspace_id=...` | `prompt_application_runtime:read` | 读取当前指针，并按当前 owner 状态重验资格 |
| `GET /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/events?workspace_id=...` | `prompt_application_runtime:read` | 读取当前指针与完整只追加事件序列 |
| `POST /v1/user-workspace/applications/{application_id}/prompt-runtime-assignment/decisions` | `prompt_application_runtime:write` | 执行显式 CAS 决策 |

首次激活示例：

```json
{
  "workspace_id": "workspace_demo",
  "expected_assignment_version": 0,
  "action": "activate",
  "candidate_id": "candidate_prompt_v1"
}
```

替换必须提交当前 assignment version 和新的 approved candidate；撤销必须提交当前 version、`action=revoke`，并省略 `candidate_id`。被撤销的 assignment 是当前 v1 状态机的终态，不能 `activate` 或 `replace`；后续如确需重新启用，必须先扩展并审查新的 owner policy / schema，不能绕过现有状态机。每个成功决策追加一个 event，事件序号、结果版本和 assignment digest 必须连续一致。

读取 assignment 不是简单返回缓存指针。服务端会重新检查当前 candidate 是否仍为 approved、是否被更新候选取代、精确 draft / template ref 与 digest 是否一致，以及应用类型、生命周期与作用域是否仍有效。任何漂移返回稳定失败码，不回退旧 candidate、旧模板版本或内存 fixture。

## 受控调用与 Session v2

API key 调用使用：

```text
POST /v1/prompt-applications/invocations
Authorization: Bearer <APPLICATION_API_KEY>
```

该 key 必须属于目标应用、处于有效状态且显式包含 `prompt_application:invoke`。应用、tenant 和 workspace scope 全部从可信 key 上下文解析，不接受 body 覆盖。请求只允许：

```json
{
  "variables": {
    "question": "如何审查本次发布？",
    "tone": "清晰"
  },
  "client_invocation_key": "prompt-guide-001"
}
```

服务端会重读当前 exact authority、确定性渲染、执行一次既有 Gateway 调用并校验输出契约。首次成功响应同时返回 `output` 与 `workflow_run_record.v6`；终态幂等重试只返回 metadata，不恢复或重放 output。相同 `client_invocation_key` 携带不同变量会失败，并发重复在首个调用运行中返回 `prompt_invocation_duplicate_running`。

Session 路径先创建显式 profile：

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_aaaaaaaaaaaaaaaa",
  "execution_profile": "prompt_application_invocation_v1"
}
```

随后向 `POST /v1/user-workspace/application-sessions/{session_id}/turns` 提交：

```json
{
  "workspace_id": "workspace_demo",
  "application_id": "app_aaaaaaaaaaaaaaaa",
  "expected_session_version": 1,
  "client_turn_key": "prompt-session-turn-001",
  "input_text": "",
  "condition_values": {},
  "model": "",
  "variables": {
    "question": "如何审查本次发布？",
    "tone": "清晰"
  }
}
```

Prompt profile 只消费 `variables`，不接受调用方用 `model`、模板、版本或 authority 改写服务端决策。Session / Turn v2 只保存 authority、input digest / bytes、变量名摘要、状态和 Run v6 引用，不保存 transcript、变量值或 `prompt_output`。

## 常见失败与处理

| failure code | 含义与处理 |
| --- | --- |
| `prompt_template_scope_denied` | 检查身份投影、scope，以及 body/query/path 与 Template 资源绑定头是否一致 |
| `prompt_template_payload_invalid` | 检查 schema version、ID、未知字段、长度预算和 UTF-8 |
| `prompt_template_syntax_invalid` / `prompt_template_variable_invalid` | 检查受限占位符、变量声明、类型、必填项和额外变量 |
| `prompt_template_secret_material_forbidden` | 删除 token、credential、Authorization、cookie、DSN 等材料；不要改用编码或摘要绕过 |
| `prompt_template_version_conflict` | 读取 `current_draft_version`，重新加载后再决定是否保存 |
| `prompt_template_write_disabled` | HTTP 可读但 write gate 未开启；不要把 validate 成功解释为可保存 |
| `prompt_template_store_unavailable` / `prompt_template_store_contract_mismatch` | 检查 selector、连接、migration marker / checksum 和持久化记录；禁止回退 memory |
| `prompt_template_application_kind_mismatch` | 只有未归档 `prompt_application` 可以拥有模板 |
| `prompt_template_binding_ineligible` | 重读草案与精确 Template Version，检查 v3 kind、作用域、digest 和 CAS |
| `prompt_runtime_assignment_not_found` | 当前没有 assignment；只能由合格 approved v3 candidate 首次 `activate` |
| `prompt_runtime_assignment_version_conflict` | 使用响应中的 `current_assignment_version` 重新加载并处理并发冲突 |
| `prompt_runtime_candidate_ineligible` | candidate 不存在、未批准、类型错误、被取代或其 exact authority 不可用 |
| `prompt_runtime_authority_changed` | assignment 指向的 application / candidate / draft / template 已漂移；必须显式审查并 replace / revoke |
| `prompt_runtime_transition_invalid` | action 与当前状态不兼容；revoked assignment 不能再次 activate 或 replace |
| `prompt_invocation_input_invalid` | 检查变量契约、预算、未知字段和 `client_invocation_key`；不要提交 provider 或 authority 字段 |
| `prompt_invocation_duplicate_running` | 同一幂等键仍在运行；只读取运行 metadata，不发起新调用 |
| `prompt_invocation_canceled` | 请求已取消；终态不 replay |
| `prompt_invocation_outcome_unknown` | provider 或终态写入结果不确定；禁止自动重试或改键绕过 |
| `prompt_invocation_output_contract_failed` | provider 响应未满足模板输出契约；Run 只保留脱敏诊断 |

## 隐私与验证边界

- Template owner 可以保存模板源码与安全默认值，但不得保存 provider credential、运行变量、渲染消息或模型输出。
- 草案 / 版本 list 只返回摘要；源码 detail 必须使用独立 `read_source` 权限。
- Configuration Draft、Publish Candidate、Runtime Assignment 和 Event 只保存精确 ref / digest，不复制模板正文。
- assignment、History、Comparison、Evaluation 和 Operations 路由不调用 Gateway、provider、工具或业务写入；只有 invocation service 允许一次计划内 Gateway 调用。
- Run v6、Session v2 和 Turn v2 不保存变量值、rendered messages、完整 output 或 provider raw response；终态重试只返回 metadata。
- 日志、错误、fixture 和 committed 文档不得出现 token、Authorization、cookie、DSN、provider raw URL / response 或真实用户输入。
- 当前能力仅用于开发测试。Prompt Web 已进入批次 E 验收，但生产认证、生产 repository、retry / fallback、replay / resume、schedule、quota 和 billing 仍未启用。

提交相关修改前至少执行：

```bash
cd services/platform
GOCACHE=/tmp/radishmind-go-cache go test ./internal/config ./internal/httpapi
cd ../..
./scripts/check-repo.sh --fast
```

触及 schema、架构、持久化边界或 API 契约时，应再执行全量 `./scripts/check-repo.sh`。真实 PostgreSQL 验证必须使用前述专项入口并在结束后关闭容器。
