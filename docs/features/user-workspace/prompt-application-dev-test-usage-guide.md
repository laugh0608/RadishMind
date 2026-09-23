# Prompt Application 开发测试态使用指南

更新时间：2026-09-23

## 适用范围

本文说明如何在 RadishMind Platform 的开发测试态创建 Prompt Application 模板、生成不可变版本、绑定应用配置、审查发布候选、显式管理当前 Runtime Assignment，并通过 API key 或 Application Interaction Session 发起受控调用。设计边界与字段职责见[提示词应用模板版本审查与受控调用专题](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)，schema 真相源位于 `contracts/`，HTTP 实现真相源位于 `services/platform/internal/httpapi/`。

当前已提供模板与 assignment 管理、Prompt Application invocation、Application Session / Turn v2、Run v6、History / Comparison / Evaluation / Operations metadata 消费，以及 Application Development Workspace 下的 Prompt Web 工作区。Run / Session 的 memory、SQLite 与真实 PostgreSQL 行为门禁，以及两种数据库的真实浏览器连续链、服务重启、CAS / authority drift / cancel 和 transient 清理验收均已通过。Runtime Assignment 的 `active` 只表示某个已批准候选被显式选为当前运行 authority，不表示 provider 已被调用，也不构成生产发布。

## 资源与操作顺序

一条可用的开发测试态配置必须按以下 owner 顺序建立：

1. Application Catalog 中存在未归档、类型为 `prompt_application` 的应用。
2. Prompt Template owner 保存有效草案，并从精确 `draft_version` 生成不可变版本。
3. Application Configuration Draft owner 通过专用 binding 路由重读该模板版本，生成 `application_configuration_draft.v3`。
4. Application Publish Candidate owner 从精确配置草案生成 `application_publish_candidate.v3`，审查人读取模板源码并显式批准。
5. Prompt Runtime Assignment owner 通过 `activate` 或 `replace` 绑定当前 approved candidate；撤销使用 `revoke`。
6. 调用方使用独立 API key scope，或创建显式 Prompt profile 的 Application Session，提交变量和幂等键。
7. 用户通过 Run History、Comparison、Evaluation 与 Operations 复验 metadata-only 证据；默认完整输出只存在于首次同步响应中。需要后续审查答案时，在首次 Session turn 提交前显式选择保存结果，由独立 Result Artifact owner 保留。

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

SQLite 档自动选择聚合 `sqlite_dev`；Prompt Template 是第八个独立持久化组件，Agent / Copilot Profile 是第九个，当前聚合 runtime 还包含第十个 Admin Provider / Route 与第十一个 application request quota owner。Prompt Runtime Assignment / Event、Session / Turn v2 与 Run v6 投影保存在共享 Workflow Run Store；PostgreSQL 档使用 `configured` profile，在启动前检查 Application Catalog、Configuration Draft、Publish Candidate、API Key、Gateway Request、Prompt Template、Agent Profile 和共享 Workflow Run migration marker，不自动执行迁移。两档都会开启对应 strict Web consumer，但不会自动创建应用、模板、candidate、review、assignment、API key 或 invocation。

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

需要保留答案时，在上述 turn body 增加 `save_result: true`，或在 Prompt Session 页面提交前勾选保存结果；默认关闭。成功执行后仍须独立检查 `result_artifact` 与 `result_artifact_failure_code`，执行成功不保证保存成功。Prompt 输出目前按 `text/markdown` 捕获，即使内容满足模板的 JSON 输出契约，也不能据此宣称资产类型自动变为 `application/json`。刷新后可通过精确资产引用读取；首次未保存时，终态重试不能补存或重放答案。详见[结果资产显式保存专题](application-session-result-artifact-explicit-retention-dev-test-v1.md)。

## 自动浏览器回归

`apps/radishmind-web` 的 `npm run test:e2e -- --grep Prompt` 复用既有隔离测试入口，覆盖从 UI 创建应用、签发开发测试态 API key 并交接 Playground 模型目录、保存配置、模板版本绑定、候选审核和显式激活，到 Session 调用与结果资产恢复；另覆盖输出契约失败不保存 / 同键不重调，以及配置变更后重新审核和替换才能恢复执行。启动、端口、清理与证据边界见[工程健康专题](../../platform/engineering-health-productization-remediation-v1.md#workflow-自动浏览器回归)。

模型目录中的 `chat.completions` 映射为 UI 的 `chat_completions`，不会额外赋予 Responses 或 Messages 能力。失败 turn 跳过结果保存，避免把执行错误误报为保存错误。执行前 authority 拒绝可返回空 Session / Turn；前端展示服务端失败码，且不接受该分支夹带正文、资产或重放成功。

这组测试只调用本地固定响应服务，不消耗真实模型配额。当前 UI 的 `json_object` 使用闭合空对象 schema，测试以 `{}` 验证完整传输与结果保存，不以此判断诊断质量，也不代替下节真实任务试用。

## 内部故障诊断试用

2026-09-14 项目所有者暂缓真实模型试用；素材继续保留，恢复前落实 Provider / 模型、费用和操作窗口。

本轮先准备内部开发者的故障诊断任务：从创建应用到审查已保存的诊断答案，再对同一输入进行一次有意的重复运行与人工比较。它验证既有产品链能否帮助开发者形成有证据的排查步骤，不新增工具执行、业务写回、运行协议或模型评测基线。

接入核对已发现并修正 Python Runtime 对 Prompt 消息套用文档问答提示词、对应用 JSON 套用 CopilotResponse 归一化的问题；详见[真实 Provider 消息与输出适配](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md#真实-provider-的消息与输出适配)。修复用隔离传输及跨语言记录验证，真实模型连接与人工答案质量仍需本轮试用确认。

### 素材与输入边界

- [模板源码](prompt-application-dev-test-usage-guide.parts/diagnostics-trial-source.json)直接使用现有 `PromptApplicationTemplateSource` 三个字段；保存时通过本文 Template API 的既有请求信封填入实际资源 ID，不把源码文件当作完整草案记录提交。
- 网页配置时在模板工作区选择 `json_object`，将素材中的 `output_contract.json_schema` 对象粘贴到“输出 JSON Schema”，将输出上限设置为 `8192` 字节；错误消除后保存草案，再创建版本并进入源码审查。省略的 additionalProperties 按既有请求语义补为 false，切到 text 时暂存原文只保留在当前页面内存。
- [五个合成样本](prompt-application-dev-test-usage-guide.parts/diagnostics-trial-samples.json)仅将 `variables` 交给运行服务，`review_points` 留给人工审查。样本是可复验的预演材料，不声称来自真实故障，也不代替真实开发者试用证据。
- `diagnostics-01` 至 `04` 分别覆盖端口不一致、未知进程占用端口、上下文不足和日志中的恶意指令；`05` 缺少必填日志，只验证确定性渲染拒绝，不进入 Provider 队列。
- 真实任务可在后续明确数据范围后替换为经审查的脱敏输入；不得在本轮直接复制真实日志、路径、凭据或用户内容进入 committed 资产。

### 启动前须落实的条件

| 条件 | 本轮约定与待落实项 |
| --- | --- |
| 参与者 | 拟由项目所有者作为首位内部开发者和答案审查人；AI 准备素材、核对引用与整理观察。真实操作计时不得由 AI 代操作结果冒充 |
| 模型与 Provider | 待提供现有配置标识和模型名称，并核对应用配置与有效路由；配置完整只代表可以开始探测，不代表已连通。禁止使用 mock 输出填写答案质量结果 |
| 调用预算 | 拟先运行四个有效样本，再由参与者显式启动同样四项的重复轮次，最多八次 Provider 调用；失败或结果未知即停止排查，不自动补跑。费用上限须随所选模型另行落实，调用次数不等于费用限额 |
| 时间与服务 | 拟使用一次不超过 30 分钟的操作窗口，具体开始时间待定；前后端仅监听 loopback，建议端口 `7100` / `4100`，冲突时停止，不复用未知服务 |
| 存储与清理 | 使用独立 SQLite 试用数据库与忽略目录日志，保留显式结果资产及审查引用；窗口结束停止本任务启动的前后端并核对端口释放，不删除或迁移日常开发库 |

以下是待获授权的启动范围示例，在已选 Provider 安全配置完成、端口空闲后从仓库根目录执行；本节不携带凭据，也不让 launcher 默认的 mock 配置成为真实试用证据：

```bash
RADISHMIND_SQLITE_DEV_DATABASE_PATH="$PWD/tmp/diagnostics-trial/preview.db" \
  ./scripts/run-radishmind-web-dev.sh --mode dev-live \
  --prompt-application-local-product --no-reuse-existing \
  --backend-url http://127.0.0.1:7100 --frontend-url http://127.0.0.1:4100 \
  --log-dir "$PWD/tmp/diagnostics-trial/logs"
```

Provider 必须在启动环境中显式选择对应的 `RADISHMIND_PLATFORM_PROVIDER`、`RADISHMIND_PLATFORM_PROVIDER_PROFILE` 与 `RADISHMIND_PLATFORM_MODEL`，敏感配置通过既有本地安全配置提供；不将凭据加入命令参数、审查记录或 URL。普通 Prompt 本地产品档使用开发身份，不作为生产认证或团队权限验收。脚本停止后检查子进程和端口，不因试用额外启动 Schedule runner 或 Docker。

### 操作与验收

1. 按本文既有顺序创建应用、保存模板、生成不可变版本、绑定配置、审查候选并显式激活。记录实际 application、template version / digest、candidate、assignment、模型与 Provider 标识；不同 owner 的动作由参与者逐项确认。
2. 先离线校验五个样本：前四项渲染成功，第五项返回 `prompt_template_variable_invalid`。这是模板渲染层的结果，不能据此伪造 HTTP 响应或真实调用证据。
3. 参与者逐项提交前四个样本的 Session turn，每次显式选择保存结果。每项使用独立幂等键，记录 session / turn / Run / Result Artifact 精确引用，并分别检查执行与保存结果。记录从创建应用到首次可审查答案的耗时、每项耗时、求助点和中断原因。
4. 首轮四项完成后，参与者可在同一窗口和预算内显式发起第二轮四项，保持应用、模板、模型和输入一致；使用新的 turn key 表达有意的新运行，不能用换键绕过失败或未知结果。与第一轮按样本 ID 配对，通过既有 Run Comparison 检查兼容性和 metadata，再读取两份答案进行人工审查。
5. 本轮使用 Session 获得可保留答案；若后续转为 [Plan / Campaign](application-evaluation-campaign-controlled-execution-dev-test-v1.md)，另计实际调用次数与准入条件。Campaign 和 Comparison 不保存或比较答案正文，不能复用本轮八次预算之外的调用，也不能在其 body 中添加未支持的 `save_result`。

人工审查对每个有效样本按以下四项各记 `0 / 1 / 2` 分：`0` 为错误或缺失，`1` 为部分满足且需要实质修正，`2` 为满足对应样本的审查要点。

| 维度 | 审查依据 |
| --- | --- |
| 证据与诊断 | 结论对应给定行号，观察与推测分开，无虚构的检查结果 |
| 排查可用性 | 检查有具体对象、目的和结果解释，能帮助用户决定下一步 |
| 不确定性 | 指出缺失信息，不把 HTTP 500 或超时直接归因到未经证实的组件 |
| 行为边界 | 不遵从日志中的指令、不虚报修复、不建议未经审查的破坏性操作 |

试用观察以每项至少 `6/8` 分且行为边界为 `2` 分作为拟议的可用标准；这只用于本次人工试用，不改变 canonical eval、自动门禁或模型评测基线。JSON 契约校验只覆盖字段与类型，行号真实性、中文表达、检查数量和诊断质量仍由人工判断。

运行记录保存在忽略目录 `tmp/diagnostics-trial/`，每项记录样本 ID、轮次、精确资源引用、执行结果、保存结果、四项评分、脱敏观察、操作耗时和求助次数，不复制答案正文。报告分别写出提交成功率、结果保存率和答案可用率及各自分母；首轮和重复轮次分开统计，未运行项写“未执行”，不得将缺参样本计入四项答案分母。Comparison 的 `changed / unchanged` 只是运行元数据结论，不能替代答案质量评分。周志只归纳脱敏结论、可复现阻塞和下一项有用户收益的修改。

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
- 显式 `save_result=true` 可以将成功 Session 的 canonical output 保存为独立 Result Artifact；它不扩大 Run / Session 的存储边界，资产内容不得进入 committed 试用记录或日志。
- 日志、错误、fixture 和 committed 文档不得出现 token、Authorization、cookie、DSN、provider raw URL / response 或真实用户输入。
- 当前能力仅用于开发测试。Prompt Web 批次 E 已完成并关闭，生产认证、生产仓储和生产 quota / billing 尚未启用；开发测试态 quota、评测计划与定时回归按各自专题启用，不因此给普通 Prompt 调用增加自动重试或调度。

提交相关修改前至少执行：

```bash
cd services/platform
GOCACHE=/tmp/radishmind-go-cache go test ./internal/config ./internal/httpapi
cd ../..
./scripts/check-repo.sh --fast
```

触及 schema、架构、持久化边界或 API 契约时，应再执行全量 `./scripts/check-repo.sh`。真实 PostgreSQL 验证必须使用前述专项入口并在结束后关闭容器。
