# Workspace-scoped Read Transition / 工作区选择与成员资格绑定（开发 / 测试态）v1

更新时间：2026-08-19

状态：`workspace_scoped_read_transition_dev_test_v1_batch_b_complete`

## 文档目的

本专题承接 [Admin Control Plane Authenticated Read Store Transition v1](../admin-control-plane/authenticated-read-store-transition-v1.md) 留出的第四批边界，把 Applications、API keys、Workflow definitions、Runs 和 Quota 五类 User Workspace 读操作从“只有 tenant / permission、没有 workspace membership owner”的状态推进到可复验的开发 / 测试态工作区授权路径。

本专题不延续 Provider Route A–E，不创建新的用户、tenant、role、workspace 或业务数据真相源。首批只建立共享 `WorkspaceMembershipProvider`、确定性的 dev / signed-test provider、统一路由授权和既有 durable owner 的读投影适配；真实 Radish OIDC、生产 membership adapter 与 quota policy owner 后置。

## 当前事实

- `VerifiedControlPlaneIdentity` 已统一承载已验证的 issuer、subject、tenant、permission、时间和 session reference。
- 七条 read route 已共享 `ControlPlaneReadRepository` 与 sanitized response envelope。
- tenant / audit 已迁移到 PostgreSQL 开发测试态 repository；五条 workspace operation 在 OIDC integration test 模式中仍统一返回 `workspace_membership_unavailable`。
- Application Catalog、API Key Lifecycle、Workflow Definition Release 和 Workflow Run History 已分别拥有 memory / SQLite / PostgreSQL 开发测试态 owner；不能复制其表、record 或 repository。
- quota 当前没有可信 policy owner；fake quota summary 不能晋级为 authenticated workspace 数据源。

## 目标用户流程

1. 已验证身份发起 User Workspace 读请求，并通过 `X-RadishMind-Active-Workspace` 显式选择本次请求的 active workspace。
2. auth boundary 只产生 sanitized identity 与已验证的 dev / signed-test membership assertion；handler 不解析 raw token 或 membership header。
3. `WorkspaceMembershipProvider` 按 verified tenant、subject、active workspace、membership expiry 和 route permission 生成 `ControlPlaneResourceBinding`。
4. tenant、subject、workspace、permission 或 expiry 任一不成立时在 repository 前失败关闭。
5. Applications、API keys、Workflow definitions 和 Runs 只调用既有 owner 的只读投影；quota 因 policy owner 缺失返回稳定 unavailable。
6. response 继续使用现有 envelope，不返回 membership raw record、token、角色、SQL、DSN、API key material 或业务正文。

## Active Workspace 语义

### 选择

- active workspace 是逐请求资源选择，不是身份 claim，也不是授权证明。
- 五条 workspace route 必须携带单一、合法的 `X-RadishMind-Active-Workspace`；缺失返回 `workspace_selection_missing`，重复或非法值返回 `workspace_binding_mismatch`。
- tenant / audit 两条 Admin route 不读取 active workspace，不因该 header 改变 tenant binding。
- 多 workspace 身份不选择默认项，不使用“第一个 membership”，也不回退 `workspace_demo`。

### 切换

- v1 不保存服务端 active workspace session。切换表示下一请求改用另一个显式 workspace value，并重新执行完整 membership 与 permission 判断。
- cursor 必须继续绑定 tenant、subject、workspace 与过滤条件；切换 workspace 后复用旧 cursor 必须失败。
- 同一请求内不接受 query、path、cookie 或第二个 header 覆盖 active workspace。

### 失效

- identity 过期先在 auth boundary 返回 `auth_context_contract_mismatch`，repository query count 为 0。
- membership assertion 过期返回 `workspace_membership_expired`；workspace 被移除或不存在返回 `workspace_membership_denied`。
- selected workspace 与 assertion workspace 不一致返回 `workspace_binding_mismatch`。
- provider unavailable 返回 `workspace_membership_unavailable`，不得回退 dev fake store、旧 binding 或默认 workspace。

## Identity、Membership 与资源权限绑定

| 层级 | 输入 | owner | 输出 | 禁止替代 |
| --- | --- | --- | --- | --- |
| verified identity | signed test token / future OIDC token | auth boundary | issuer、subject、tenant、identity permission、expiry | 请求头、Web state、角色名称 |
| workspace selection | active workspace header | 当前请求 | selected workspace ref | membership proof |
| membership assertion | dev-test assertion / signed-test token projection | dev / signed-test issuer | tenant + subject + workspace + workspace permission + expiry | 新用户表、新 tenant 表、新 role 表 |
| membership decision | verified identity + selection + assertion | `WorkspaceMembershipProvider` | workspace verified resource binding | handler 自行比对 header |
| resource query | resource binding + route permission | 既有业务 repository / projection owner | subject-owned workspace summary | fake fixture、跨 workspace scan |

边界规则：

- identity permission 与 workspace membership permission 是两道独立 allowlist；两者都满足才允许查询。
- tenant 和 subject 只来自 verified identity；membership assertion 只能与其相等，不能覆盖。
- 首批 durable owner 均以 `tenant + workspace + owner subject` 为既有作用域，因此 v1 只返回当前 verified subject 拥有的资源，不扩 workspace 管理员跨 owner 读取。
- `ControlPlaneResourceBinding` 保存 sanitized workspace ref、permission grant、source / policy ref 和 expiry；不保存 raw token 或 raw membership record。
- `ReadRepositoryContext` 增加 `WorkspaceID`，repository 必须把它纳入查询 predicate。

## 开发 / 测试态 Provider

### `dev_headers`

- identity 继续由显式 dev auth headers 提供。
- active workspace 使用公共选择 header；membership proof 使用独立 dev membership workspace / permission headers。
- selector 与 membership proof 分离，必须精确匹配。
- dev headers 只允许显式本地开发模式，不构成 authenticated 或 production evidence。

### `signed_test_token`

- signed test token 可携带零个或多个 `workspace_memberships`，每项只包含 workspace ref 与 workspace permission allowlist。
- assertion 继承 token 已验证的 tenant、subject、issuer、policy version 和 expiry；签名、issuer、audience、time 或结构失败仍由 auth boundary 统一拒绝。
- membership provider 只消费归一化 assertion，不解析 JWT。

### `radish_oidc_integration_test`

- 本批不猜测 Radish membership claim 或 endpoint。
- 即使 OIDC token 携带同名 permission，五条 workspace route 仍返回 `workspace_membership_unavailable`，直到 reviewed upstream membership owner 与 mapping 成立。

## 五类迁移矩阵

| read operation | v1 membership permission | durable owner / projection | v1 结论 | 停止线 |
| --- | --- | --- | --- | --- |
| Applications | `applications:read` | Application Catalog repository | 迁移；按 tenant + workspace + subject 列表，输出 sanitized catalog summary | 不推导 latest definition / last run |
| API keys | `api_keys:read` | API Key Lifecycle repository | 迁移；只输出 key ID、scope、state、时间，不输出 credential / digest | 不做 key 创建、撤销或 secret 交接 |
| Workflow definitions | `applications:read` | Workflow Definition Release summary projection | 迁移；补齐 workspace predicate | 不创建第二套 definition projection |
| Runs | `runs:read` | combined Workflow Run History store | 批次 B 已迁移 workspace-wide owner projection；`application_ref` 为可选精确筛选 | 不做跨 owner 管理员读取、replay 或正文返回 |
| Quota | `usage:read` | 缺少可信 policy owner | 保持关闭，返回 `quota_policy_unavailable` | 不使用 fake quota、不推算 token / cost、不做 enforcement / billing |

未迁移或 owner 未启用不是 empty。Application / API key / Workflow / Run store unavailable 统一映射 `read_store_unavailable`；contract / cursor mismatch 映射 `read_store_contract_mismatch` 或 `invalid_filter`。

## 稳定 Failure 与 HTTP

| failure code | HTTP | repository query |
| --- | --- | --- |
| `identity_context_missing` | `401` | `0` |
| `auth_context_contract_mismatch` | `401` | `0` |
| `tenant_binding_missing` | `401 / 403` | `0` |
| `scope_denied` | `403` | `0` |
| `workspace_selection_missing` | `400` | `0` |
| `workspace_binding_mismatch` | `403` | `0` |
| `workspace_membership_denied` | `403` | `0` |
| `workspace_membership_expired` | `403` | `0` |
| `workspace_permission_denied` | `403` | `0` |
| `workspace_membership_unavailable` | `503` | `0` |
| `quota_policy_unavailable` | `503` | `0` |
| `read_store_unavailable` | `503` | attempted after allow |
| `read_store_contract_mismatch` | `500` | attempted after allow |

跨 tenant、非成员、过期身份、workspace mismatch、permission denied 和 provider unavailable 都必须证明 repository query count 为 0。公开 failure 不区分资源是否真实存在，避免 membership enumeration。

## Repository 复用与投影约束

- handler 继续只依赖 `ControlPlaneReadRepository`，不直接访问业务 repository。
- workspace adapter 只做 context translation、filter translation、failure translation 和 sanitized summary projection。
- Application Catalog / API Key cursor 继续由各自 owner 生成并绑定 tenant + workspace + subject。
- Workflow Definition summary 必须同时检查 tenant、workspace 和 subject，修正旧投影只检查 tenant + subject 的缺口。
- Run 批次 B 使用 owner 原生 workspace-wide projection 与稳定 cursor，不做 application N+1 聚合；`application_ref` 只作为可选精确筛选。
- API key projection 禁止暴露 display secret、credential token、credential digest、raw request / audit payload。
- quota 不调用 fake repository；任何成功 fixture 都不能覆盖 `quota_policy_unavailable`。

## 隐私边界

- raw access token、Authorization、dev membership headers 和 raw membership assertion 在 auth boundary 后删除。
- active workspace ref 可进入 sanitized repository context 与 audit correlation，但不进入 URL 或浏览器持久化。
- response 不返回 issuer URL、role、email、raw permission、membership source payload、API key material、run input / output、provider endpoint、SQL 或 DSN。
- denial response 只包含既有 request / tenant / failure / audit envelope；不回显 selected workspace 或 membership 候选列表。
- membership cache 不在 v1 实现；不存在 stale allow 或跨 subject / tenant 复用。

## 开发 / 测试态与生产态验收

### 开发 / 测试态

- dev headers 与 signed test token 的 positive membership、workspace switch、non-member、tenant mismatch、subject mismatch、membership expiry 和 permission denial。
- identity expired、workspace mismatch 与 denial zero-query。
- 四类 durable owner 的 tenant + workspace + subject 隔离、cursor / filter、store unavailable 和 sanitized projection。
- quota 始终 `quota_policy_unavailable` 且 repository zero-query。
- Go unit、HTTP、repository、race、vet、fast 与完整仓库检查。

### 生产态

生产验收保持 `not_satisfied`，必须另行具备：

1. reviewed Radish membership owner / endpoint 或事件投影；
2. exact tenant、subject、workspace、permission mapping 与撤销 / expiry 语义；
3. production auth / secret / cache / audit / availability policy；
4. Web server-managed session 或独立威胁评审；
5. quota policy owner 与正式用量 / 计费边界。

production 必须拒绝 dev headers、signed test token、dev membership assertion、fake store 和请求头充当授权来源。开发测试态成功不得解释为 production authorization ready。

## 实施批次

### 批次 A：共享 Membership Boundary 与 durable read projection

- 建立 `WorkspaceMembershipProvider`、request / assertion / binding 类型与 deterministic dev-test provider。
- 为 signed test token 增加可选、签名保护的 workspace membership projection。
- 五条 route 接入统一 workspace authorization；quota 在授权后明确关闭。
- `ReadRepositoryContext` 增加 `WorkspaceID`。
- 复用 Application Catalog、API Key、Workflow Definition Release 和 Run History owner。
- 补齐跨 tenant、subject、非成员、expiry、workspace mismatch、permission denied 与 zero-query 测试。

完成锚点：`workspace_scoped_read_transition_dev_test_v1_batch_a_complete`。

实施状态：已完成。Platform 已建立共享 `WorkspaceMembershipProvider`、dev / signed-test assertion、五条 route 的统一 workspace authorization、`ReadRepositoryContext.WorkspaceID`、四类既有 owner adapter 和 quota fail-closed。相邻测试覆盖跨 tenant / subject、非成员、identity / membership expiry、workspace mismatch、permission denied、quota / Run 前置停止线和 repository zero-query；Workflow Definition summary 已补齐 workspace predicate。

### 批次 B：Workspace-wide Run Projection 与 Web Selector

- 在 Run History owner 中建立稳定 workspace-wide cursor projection，移除 v1 的精确 application 前置。
- Web 增加非持久化 active workspace selector、切换失效和 sanitized denied / unavailable 状态。
- 继续不接真实 Radish OIDC 或 quota owner。

批次 B 设计确认：

- workspace-wide Run projection 仍由既有 `workflowRunStore` 的 memory / SQLite / PostgreSQL owner 承载，只新增只读 projection interface，不复制表、运行记录或写入路径。
- 查询强制绑定 verified tenant、active workspace 与 verified subject；`actor_ref` 是当前阶段可复用的 durable owner 边界，不扩 workspace 管理员跨 owner 读取。
- `application_ref` 从必填前置改为可选精确筛选；`workflow_definition_id`、`status` 与 `failure_code` 保持可选筛选。
- 稳定排序固定为 `started_at DESC, run_id DESC, application_id DESC`。cursor 同时保存最后一项三元组，并绑定 tenant、subject、workspace、完整筛选和排序；任一变化都返回 `invalid_filter`，不得静默换作用域。
- Web selector 只在当前 React 进程内保存 active workspace。切换成功后立即清空 application、workflow definition、run 与 draft 等旧 workspace 选择，并用新 workspace 重新加载七条 read route。
- workspace route 发出 `X-RadishMind-Active-Workspace`；仅 `dev_headers` 同时发出确定性的 dev membership proof。signed-test membership 继续只来自签名 token，OIDC integration test 继续明确 unavailable。
- selector 不写 URL、cookie、`localStorage`、`sessionStorage` 或服务端 session；输入不合法时不发请求，denied / unavailable 只展示 sanitized envelope failure。

完成锚点：`workspace_scoped_read_transition_dev_test_v1_batch_b_complete`。

实施状态：已完成。Run owner 新增独立 workspace-wide projection interface，memory、SQLite、PostgreSQL 及 combined Workflow / Prompt / Agent store 均复用既有运行记录；查询绑定 tenant + workspace + actor subject，稳定 cursor 绑定作用域、筛选与 `started_at DESC, run_id DESC, application_id DESC`。`application_ref` 已从必填前置改为可选精确筛选。

Admin Web 已增加进程内 active workspace selector。五条 workspace read route 携带 active workspace；仅 dev headers 模式携带逐路由最小 membership permission，signed-test 与 OIDC 不混入 dev proof。切换会清空旧 application、workflow definition、run 与 draft 选择并重新加载；离线态不发送请求，非法值不切换。

### 历史条件式批次 C：Reviewed Membership Adapter

本条件只保留 legacy Radish resource-server membership adapter 的历史语义：只有 Radish owner 提供 reviewed membership contract 后，才能设计该 integration adapter、撤销 / cache / unavailable 和对应 OIDC evidence，不得从 deterministic provider 自动晋级。2026-08-19 后，RadishMind 交互式授权改由本地 identity / membership owner 承载，browser OIDC 只做外部认证；因此不从本专题启动该批次 C。

## 停止线

- 不接真实 Radish OIDC，不读取 Radish 数据库，不猜测 membership claim / role / endpoint。
- 不创建 user、tenant、role、workspace、permission 或 quota 第二真相源。
- 不实现 production secret、membership cache、billing、cost ledger、quota enforcement 或自动模型路由。
- 不修改 Provider Route A–E，不派生同类批次。
- 不新增业务写入、API key lifecycle mutation、Workflow execute / replay / resume、外部项目集成。
- 不把请求头、signed test token、fake store 或 dev/test repository 作为生产授权来源。

## 下一实现入口

批次 A、B 已完成，开发 / 测试态 v1 的 workspace authorization、durable read projection、workspace-wide Run history 与非持久化 Web selector 已形成闭环，本专题保持关闭。RadishMind 本地 membership owner 和 Web Session actor 已转由[本地账户与 Radish OIDC 联合登录 v1](../admin-control-plane/local-account-radish-oidc-federated-login-v1.md)批次 A、B 实现；下一步只在新专题推进确定性 browser OIDC，不恢复本专题的 legacy Radish membership adapter。
