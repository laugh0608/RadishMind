# 应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1

更新时间：2026-08-09

状态：`application_api_key_request_quota_admission_dev_test_v1_backend_a_to_d_completed_pencil_blocked`

## 功能目标

为使用开发测试态 API Key 的模型调用建立应用级请求预算。管理员在明确的 tenant、workspace、application 与 `development | test` 环境中维护 UTC 自然日请求上限；每次真正进入 bridge / provider 前，平台以原子方式记录一次 admitted provider attempt。达到上限后，新 attempt 以稳定失败关闭，provider 调用数保持为零。

本专题把现有 `quota_policy_unavailable` 从诚实停止线推进为新的受控 owner，不复用离线 `QuotaSummary` fixture，也不从 Request History 分页窗口反推总量。它只服务内部开发者预览，不声明生产 quota、rate limit、价格、成本账本或 billing 已成立。

## 事实审计结论

- `workspaceScopedControlPlaneReadRepository.ReadQuotaSummary` 当前固定返回 `quota_policy_unavailable`，证明真实 policy owner 不存在。
- 既有 `QuotaSummary` 只有 tenant、period、request / token / cost limit 和离线 usage snapshot，没有 workspace、application、environment、policy version 或并发 admission 语义。
- Gateway Request History 在 API Key 认证后、请求正文校验前创建；provider reported usage 只在终态 envelope 返回后取得。History 还是分页审查 owner，memory 模式会淘汰旧记录，不能承担周期总量和原子限额。
- `/v1/chat/completions`、`/v1/responses`、`/v1/messages` 复用 `prepareGatewayRequest`，但 Application RAG、Prompt Application 与 Agent / Copilot invocation 在各自 authority、idempotency、retrieval 和 run owner 内调用 bridge。只在 HTTP middleware 扣减会把非法输入、资格失败和 idempotent replay 误算为 provider attempt。
- 当前 `Server` 没有 quota repository、migration、管理权限、稳定 `429` 映射或 provider 前 admission hook。因此实现必须扩展 API、schema、repository 与执行边界；该扩展已于 2026-08-09 获得明确授权。

## 用户与核心流程

目标用户是需要限制内部开发测试应用模型请求次数的平台管理员和应用开发者：

1. 管理员选择 workspace、application 与开发测试环境。
2. 管理员读取当前 policy、UTC 日窗口、已 admitted 次数和剩余次数。
3. 管理员以预期版本创建或更新正整数 `request_limit`。
4. API Key 请求先完成凭据、scope 与 active application 校验。
5. 请求继续完成输入校验、runtime authority、idempotency、retrieval 和 canonical request 构造。
6. bridge wrapper 在实际 provider attempt 前调用 quota owner；admitted decision 与计数增量在同一原子操作提交。
7. admitted 后才允许调用 bridge / provider。provider 失败、超时或 outcome unknown 仍消耗该次 attempt，不退款。
8. 达到上限后返回 `gateway_quota_exceeded`；不得调用 bridge / provider，不得伪造 reported usage。
9. 管理员可重新读取同一窗口的 policy 与 usage；提高或降低正整数上限必须是显式 CAS 管理动作。v1 不提供 policy 删除或禁用动作。

当前实现采用单个 Gateway 进程对应一个显式 `development | test` 环境：API Key 记录本身没有环境字段，enforcement 从受控运行配置取得环境；Admin API 仍要求逐请求携带同一开发测试态环境，以便精确管理作用域。该边界不表示跨环境复制、自动推断环境或生产环境 policy 已成立。

## v1 作用域

Policy 原子作用域固定为：

```text
tenant_ref + workspace_id + environment + application_id
```

- `environment` 只允许 `development | test`。
- 每个作用域只有一个当前 policy；`policy_id` 由服务端稳定生成，`record_version` 从 `1` 单调递增。
- 周期固定为 `calendar_day_utc`；bucket key 是 UTC `YYYY-MM-DD`，不接受客户端时区或自定义窗口。
- `request_limit` 只允许 `1..1000000`。
- 配额应用级共享；`api_key_id` 只用于 admitted decision 归因，不形成隐式的第二套 key policy。

v1 只计算由 `api_key_dev_test` 可信上下文触发、真正准备进入 provider 的以下调用：

- `POST /v1/chat/completions`
- `POST /v1/responses`
- `POST /v1/messages`
- `POST /v1/application-rag/invocations`
- `POST /v1/prompt-applications/invocations`
- `POST /v1/agent-copilot/invocations`

`GET /v1/models*`、dev header 调试、用户会话调用、Workflow executor、HTTP Tool、评测执行和离线 smoke 不进入 v1 policy。模型目录只读取 inventory，不冒充 provider inference attempt。

## Owner 与原子语义

新增唯一 `GatewayRequestQuotaRepository`，负责：

- `ReadPolicy`：精确读取 policy；
- `PutPolicy`：以 `expected_version` 原子创建或更新；
- `ReadUsage`：读取当前 UTC bucket；
- `AdmitProviderAttempt`：在同一原子操作内解析 policy、判定上限、增加计数并写 admitted decision。

`AdmitProviderAttempt` 输入必须来自可信 API Key context：tenant、workspace、application、`api_key_id`、HTTP request id、route、environment 与服务端时间。客户端不能覆盖这些字段。

admitted decision 使用 request id 做唯一约束。重复 attempt id、policy 缺失、policy / environment 不匹配、repository 不可用和计数冲突全部失败关闭。计数线性化点发生在 bridge 调用前；一旦 admitted 即表示平台允许了一次 provider attempt，后续 provider 失败不回退计数。进程在 admitted 后、调用前崩溃时保守保留计数，避免重启后超发。

memory、SQLite 与 PostgreSQL 必须具有相同语义：

- 并发 `remaining=1` 时最多一个 admitted；
- policy 与 usage 不允许跨 tenant / workspace / environment / application 混用；
- SQLite / PostgreSQL 使用事务和条件更新，不通过先查后写的非原子聚合实现；
- store、migration 或 schema 不可用时不回退 memory、fixture 或 Request History。

## API 与权限

管理 API 使用新的 owner，不并入 Provider Route 配置：

```text
GET /v1/admin/gateway-request-quotas/{application_id}
PUT /v1/admin/gateway-request-quotas/{application_id}
```

开发测试态请求显式携带 workspace 与 environment；tenant、actor、permission 和 membership 来自 verified identity。权限分离为：

- `admin_gateway_quotas:read`
- `admin_gateway_quotas:write`

PUT body 只允许：

```json
{
  "expected_version": 0,
  "request_limit": 100
}
```

创建要求 `expected_version=0`；更新要求匹配当前版本。响应只返回脱敏 policy、当前 UTC usage、剩余次数、request / audit lineage 和稳定 failure，不返回 API Key token / digest、provider credential / endpoint、输入输出、raw membership 或数据库诊断。

现有 `/v1/user-workspace/usage/quota-summary` 继续保持 `quota_policy_unavailable`，直到独立用户工作区 read projection 明确 application 选择和权限契约。本批不把 Admin policy response 偷换成旧 tenant-only fixture。

## 稳定失败与副作用

至少固定：

- `gateway_quota_disabled`
- `gateway_quota_scope_denied`
- `gateway_quota_environment_forbidden`
- `gateway_quota_payload_invalid`
- `gateway_quota_policy_not_found`
- `gateway_quota_policy_version_conflict`
- `gateway_quota_attempt_conflict`
- `gateway_quota_exceeded`
- `gateway_quota_store_unavailable`

`gateway_quota_exceeded` 在标准 northbound API 映射为 HTTP `429`；Application RAG、Prompt 与 Agent invocation 在各自严格 envelope 中返回同一稳定 code，并使用 `429`。所有 quota 拒绝都必须证明 bridge / provider 调用为零。store unavailable 与 policy missing 使用 `503`，不能按 unlimited 放行。

## Provider reported usage 边界

- v1 request limit 不依赖 token usage，因此 stream 和 `not_reported` 仍可确定 admission。
- reported input / output / total tokens 继续只进入 Request History 与 Application Operations，不能回填或改写 admitted count。
- token limit、token reservation、本地 tokenizer 估算、价格和 cost limit 继续关闭；旧 `QuotaSummary` 中的 token / cost 字段不构成本专题 owner。

## 运行配置与迁移

能力默认关闭。Admin 管理面分别由 `RADISHMIND_GATEWAY_REQUEST_QUOTA_DEV_HTTP=1` 和 `RADISHMIND_GATEWAY_REQUEST_QUOTA_DEV_WRITE=1` 开启；执行准入由 `RADISHMIND_GATEWAY_REQUEST_QUOTA_ENFORCEMENT_DEV=1` 与单一进程环境 `RADISHMIND_GATEWAY_REQUEST_QUOTA_ENVIRONMENT=development|test` 开启，并要求 API Key 开发测试态认证已启用。Admin 请求还必须携带与该进程环境相同的 `X-RadishMind-Dev-Gateway-Quota-Environment`。

store 由 `RADISHMIND_GATEWAY_REQUEST_QUOTA_STORE=memory_dev|sqlite_dev|postgres_dev_test` 选择。聚合 `RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev` 会统一投影 quota store 并应用 SQLite migration，但不会自动开启管理或 enforcement gate。PostgreSQL 模式要求 runtime DSN `RADISHMIND_GATEWAY_REQUEST_QUOTA_DEV_TEST_DATABASE_URL`，DDL 只能由独立 migration DSN `RADISHMIND_GATEWAY_REQUEST_QUOTA_DEV_TEST_MIGRATION_DATABASE_URL` 交给以下 runner：

```bash
cd services/platform
go run ./cmd/radishmind-gateway-request-quota-migrate up
go run ./cmd/radishmind-gateway-request-quota-migrate status
```

连接和 migration 检查超时可由 `RADISHMIND_GATEWAY_REQUEST_QUOTA_DATABASE_TIMEOUT` 设置。selector 只检查已应用 marker / checksum，runtime 不执行 DDL；缺失 migration、未知 store 或连接失败不回退其它实现。

## Web 与 Pencil

Web 后续在 Admin Control Plane 新增 Quota 任务 owner，读取当前 application policy、UTC window、used / remaining、record version 与 blocked reason；更新动作要求显式预期版本和影响确认。普通列表保持中性，只有驱动详情的 policy / application 才可选中；`gateway_quota_exceeded` 的注意色不得冒充选中。

五维初评为 `1 / 2 / 2 / 2 / 1 = 8`：需要新管理任务、原子阻断表达、policy / usage 主次和窄屏顺序，预期为 `A / 完整 Pencil`。Pencil 当前被其他项目占用，本专题在其空闲并通知用户前不操作设计源；后端实现不据此虚构最终 UI。

## 实施批次

唯一实施入口见[实施任务卡](../../task-cards/application-api-key-request-quota-admission-dev-test-v1-plan.md)。

1. 批次 A：领域、memory repository、并发 admission 与稳定 failure。
2. 批次 B：SQLite / PostgreSQL migration、repository、selector、重启和 no-fallback。
3. 批次 C：Admin read/write API、权限、严格 JSON 与作用域负向验证。
4. 批次 D：bridge wrapper、六条 API Key inference route、history / run failure 和零 provider 副作用。
5. 批次 E：Pencil、Admin Web、真实浏览器连续链、文档和仓库门禁；Pencil 占用期间不得开始。

截至 2026-08-09，批次 A 至 D 已完成：memory / SQLite / PostgreSQL repository、双数据库 migration、Admin GET / PUT、权限与 membership、六条 API Key inference route 的 provider 前 admission、稳定失败映射及 Run / Request History 诊断均已落地。PostgreSQL 真实门禁已验证 migration 幂等、受限 runtime role、八路并发只准入一次与重启恢复。批次 E 因 Pencil 仍被其他项目占用而保持未开始，React 不用临时结构绕过完整设计基准面。

## 验收

- memory、SQLite、PostgreSQL 在 `remaining=1` 的并发竞争中最多一个 admitted。
- expected version CAS、周期切换、跨作用域、重复 request id、missing policy、store failure 和 production environment 均失败关闭。
- 六条计费范围路由完成无 policy、admitted、exact-limit、over-limit、provider failure 与 idempotent replay 验证。
- `gateway_quota_exceeded` 的 bridge / provider 调用数为零；provider failure / timeout 在 admission 后计数保持增加。
- Models、dev headers 和非 API Key 内部调用不消耗 v1 policy。
- API 与 Web 不暴露 token、credential digest、Authorization、prompt、message、provider raw response、endpoint、SQL 或 DSN。
- Web 测试、production build、桌面 / 关键断点 / `390×844` 浏览器、SQLite 重启、PostgreSQL migration / role / 并发和仓库门禁通过。

## 停止线

- 不实现 production quota、真实 OIDC membership、production API Key、tenant-wide / key-specific policy 继承或跨环境复制。
- 不实现每秒 / 每分钟突发 rate limit、排队、优先级、自动提额、自动禁用、retry / fallback 或 load balancing。
- 不实现 token 强制配额、token 估算、价格、成本换算、billing ledger、invoice、结算或预算告警。
- 不把 Request History、Workflow Run、旧 quota fixture 或 Web 当前窗口变成 quota 真相源。
- 不把 admitted 解释为 provider success，也不把 provider reported usage 解释为计费凭证。
- 不在 Pencil 被占用时修改设计源；后续完整 Pencil 只冻结结构、交互、风险和响应式顺序，功能事实仍以本专题、API 契约和代码为准。
