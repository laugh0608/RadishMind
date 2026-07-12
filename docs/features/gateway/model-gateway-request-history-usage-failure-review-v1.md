# Model Gateway Request History / Usage & Failure Review v1

更新时间：2026-07-12

状态：`model_gateway_request_history_usage_failure_review_v1_complete`

## 当前实现进度

2026-07-12 已完成首个可运行的 `memory_dev` 纵向切片：Platform 新增独立 caller context、`gateway_request_record.v1`、500 条进程内 store、三个 northbound route 的 recorder、scoped list / detail API，以及 Web 独立 consumer 和 Evidence Review 内的 lazy Request History panel。已验证 invalid JSON、缺少 caller context、三个协议、unary / stream 成功、store create failure 不改写 provider outcome、scope denied、过滤 / cursor、敏感字段拒绝、offline 零请求、Web build 和 chunk budget。

2026-07-12 第二批已完成独立 `postgres_dev_test` repository、manual migration、marker / checksum preflight、runtime DML role、selector、pool close、no-fallback 和 restart recovery。破坏性集成覆盖 fresh / repeat apply、DDL 拒绝、scope、过滤分页、并发终态 CAS、pool 重开、marker mismatch、rollback / reapply；真实浏览器以 27 条记录验证 25+2 分页、详情、过滤和 Platform / Web 重启恢复，详情显示 `postgres_dev_test`，全新会话无 error / warning。

2026-07-12 第三批已完成最终终态证据与功能关闭。审查发现 northbound request context 取消后无法继续驱动 PostgreSQL terminal update，会让 durable record 停留在 `started`；recorder 现仅为 terminal store update 派生保留 caller values、移除原取消信号并受数据库超时约束的短时 context。该 context 不传给 bridge / provider，不继续生成响应，不重试 northbound 请求，也不改变取消响应语义。

真实浏览器以单 worker / 有界 queue 的显式测试配置验证 40 路并发中 38 个 `503` queue full、`504 / BRIDGE_WORKER_TIMEOUT`、unary 与 stream `408 / BRIDGE_WORKER_CANCELED`，PostgreSQL 详情和 canceled filter 均可复验，全新浏览器会话无 error / warning。完整摘要见 [终态证据附件](evidence/request-history-terminal-evidence-2026-07-12.json)。adapter 尚无可证明的 provider token 来源，因此 usage 正确保持 `not_reported`，本功能不声明 reported usage 已验证；未来出现可信 usage contract 时应作为独立功能批次打开，不阻塞本次 dev/test request history v1 关闭。

## 功能目标

让内部开发者从现有 Model Gateway surface 审查真实 northbound 请求的路由选择、用量可用性、端到端耗时和稳定失败边界，并在显式 PostgreSQL 开发 / 测试模式下跨 Platform 重启恢复记录。该能力建立独立 Gateway 请求历史真相源，不复用 Workflow run repository，也不把现有离线 quota / cost snapshot 或 fake `/v1/user-workspace/runs` 解释为真实请求记录。

本专题只打开开发 / 测试态的 sanitized operational metadata 记录、查询和人工审查。production API key、quota enforcement、rate limit、billing、cost ledger、自动 retry / fallback、load balancing 和 production gateway 继续关闭。

## 目标用户与用户路径

目标用户是调试 Gateway 兼容面、provider/profile route 和失败行为的内部开发者或测试人员。

1. 调用方通过 `/v1/chat/completions`、`/v1/responses` 或 `/v1/messages` 发起 northbound 请求；显式开发 / 测试 caller context 把请求绑定到 tenant、workspace 和 API consumer，可选绑定 application。
2. Platform 在读取 request body 前创建 started record；解析、scope、selection、bridge、provider、response translation、stream 和 client cancellation 都更新同一 request record。
3. 请求结束后记录进入不可逆终态；请求响应语义不因 review store 写入失败而被改写，但 store 失败必须进入稳定日志、指标和响应关联信息，不能静默回退 memory 成功。
4. 用户从 Model Gateway Evidence Review 进入真实 Request History，按 route、protocol、provider、profile、model、status、failure boundary、usage availability 和时间过滤。
5. 用户打开详情审查 selection source、timing breakdown、经验证的 usage summary、稳定 failure code / boundary、request id、audit ref 和 caller scope，不查看原始输入或输出正文。
6. 默认 offline Web 模式继续使用明确标记的静态 evidence，零 HTTP；只有显式 dev/test source 才读取真实 Gateway request history API。

## Caller context 与 scope

当前 northbound compatibility route 没有 production API key lifecycle、OIDC principal 或稳定 API consumer binding。现有 `controlPlaneReadAuthContext` 只服务 control-plane 读侧，不能隐式升级为 Gateway 身份真相源，也不能把 Workflow 的 application scope 原样复制到 Gateway。

本功能新增 Platform 内部 `GatewayCallerContextProvider` 边界，输出：

- `tenant_ref`
- `workspace_id`
- `consumer_ref`
- 可选 `application_id`
- `subject_ref`
- `scope_grants`
- `audit_context`
- `source`

规则如下：

- `tenant_ref + workspace_id + consumer_ref` 是 record 最小可见性 scope；`application_id` 只在调用方已被可信绑定时保存和过滤，不作为所有 Gateway 请求的强制字段。
- caller identity 只能来自受控 provider，不从 JSON body、model alias、query、客户端 metadata 或任意 `radishmind` extension 自报。
- v1 只实现显式 opt-in 的 dev/test provider；它可以读取独立 allowlist headers，但必须由 `RADISHMIND_GATEWAY_REQUEST_HISTORY_DEV=true` 与非 production environment 双重 gate 启用。
- dev/test provider 的 header 名称与 control-plane read dev auth 分离；二者可以由本地 launcher 传入同一测试身份，但不能共享隐式全局状态或互相代替 scope 校验。
- production caller provider、API key validation 和 OIDC handoff 保留为后续独立能力；production mode 选择 request history store 或 dev header provider 必须 fail closed。
- northbound 写入要求有效 caller context；缺失或 scope 不完整时请求仍按当前 compatibility 行为执行，但不创建伪造 scoped history，并记录 `caller_context_unavailable` sanitized metric。显式 dev/test history 验收请求必须提供完整 caller context。
- list / detail 始终要求可信 caller context 和 `gateway_requests:read`；header、query、cursor 与 caller binding 不一致统一返回 scope denied，不暴露记录是否存在。

这里把“兼容 API 是否可调用”和“是否能形成可审查的 scoped history”分开，避免在 production auth 尚未成立时把开发 header 写成公开授权机制。

## Request record 与生命周期

新记录 schema 为 `gateway_request_record.v1`，以 `request_id` 标识一次 northbound 请求，并在完整 caller scope 内唯一。服务端生成或规范化 request id；客户端提供的 request id 只在通过现有长度与文本卫生校验后使用。

持久化状态固定为：

- `started`：已建立 caller context 和 record，尚未完成请求解析或 selection。
- `succeeded`：非流式响应已完整写出，或流式响应已正常发出 terminal event。
- `failed`：请求在解析、selection、bridge、provider 或 response translation 边界失败。
- `canceled`：request context、client connection 或 stream consumer 明确取消。

终态不可逆，不允许 `succeeded|failed|canceled` 回到 `started` 或互相覆盖。record 使用单调 `record_version`：create 为 1，每次合法更新原子递增。请求路径允许少量关键快照，不为每个 stream delta 写库。

生命周期要求：

- started record 必须在读取 body 之前创建，从而覆盖 invalid JSON、body too large 和 caller disconnect；caller context 不成立时不创建 unscoped record。
- selection 成功后保存 provider/profile/model 和 selection source；selection 前失败时这些字段保持 unavailable，不填配置默认值伪装已选择。
- provider timing 从经校验 Gateway envelope 的 `provider_duration_ms` 取得；Platform 端到端 timing 由 Go trace 计算，二者不得混为同一字段。
- 非流式请求在 response document 成功构建后终态化；流式请求在 terminal event 写出后终态化，首段成功不代表整个 stream 成功。
- client cancellation 与 bridge cancellation 映射为 `canceled`；queue full、timeout、worker crash、protocol、provider 和 response translation 映射为 `failed` 加稳定 boundary。
- terminal store update 必须与 provider execution context 分离：取消后只允许在数据库超时预算内保存 sanitized terminal metadata；禁止借此继续 provider 调用、重放、重试、写响应或隐藏 store failure。
- Platform 重启后遗留 `started` 记录只派生 `stale_started=true`；v1 不自动恢复、不重放、不自动改写终态。默认 stale 阈值为 5 分钟，并受 Platform request timeout 上限约束。
- started create 失败时继续执行当前 northbound 请求，但记录稳定 store observation failure；后续 update 不尝试改写 memory fallback。terminal update 失败时保留最后持久快照并记录 failure，不把未落库终态伪装成 durable success。

Request History 是开发 / 测试 review evidence，不是计费或合规审计账本，因此它的 store 故障不得改变 provider 请求结果；同时 no-fallback 和显式可观测性保证故障不会被隐藏。

## 数据白名单与 usage 语义

允许保存：

- schema / record version、request id、audit ref、created / completed time。
- tenant、workspace、consumer、可选 application、subject 的稳定 reference；不保存显示名或任意 claim payload。
- northbound route、protocol、stream flag、selection source、provider、profile、model。
- Platform `duration_ms`、Gateway `gateway_duration_ms`、provider `provider_duration_ms`，缺失值使用 availability 状态而不是 0 冒充实测。
- HTTP status class、稳定 request status、failure code、failure boundary、failure category 和 allowlist review action。
- 经响应 adapter 校验的 input / output / total token count，以及 usage source / availability。
- request / response validation state、store mode、stale_started 和零自动 retry / fallback 证据。

usage 只允许以下状态：`reported`、`not_reported`、`not_applicable`。`reported` 必须来自经校验的 provider / Gateway response contract，且 token counts 为非负整数；当前 adapter 返回零但无法证明 provider 实际报告时必须记为 `not_reported`，不能把零写成真实用量。v1 不估算 token，不推导价格、成本、quota 消耗或 billing ledger。

禁止保存或输出：

- prompt、messages、instructions、input、system text、response body、stream delta 或可逆正文摘要。
- authorization header、API key、credential、secret ref 解引用结果、cookie、完整请求头或 query dump。
- provider endpoint、base URL、DSN、本机路径、worker frame、provider raw envelope / raw error、stderr、stack trace 或 SQL。
- 任意 tool payload/result、confirmation、业务写回、checkpoint、replay / resume state。

所有 string reference 设长度上限并禁止换行、URL / endpoint 形态和常见 secret marker。HTTP detail 与 Web consumer 执行同一 forbidden-field 深度扫描，不能只依赖 UI 隐藏。

## API、过滤与分页

新增独立开发 / 测试只读资源族：

```text
GET /v1/model-gateway/requests?limit=...&cursor=...&workspace_id=...&application_id=...&consumer_ref=...&route=...&protocol=...&provider=...&profile=...&model=...&status=...&failure_boundary=...&usage_availability=...&started_from=...&started_to=...
GET /v1/model-gateway/requests/{request_id}?workspace_id=...&consumer_ref=...&application_id=...
```

- tenant 只来自 caller context，不接受 query 自报；workspace、consumer 和可选 application query 必须与 caller 授权范围一致。
- 默认 `limit=25`，允许 `1..100`；排序固定 `started_at DESC, request_id DESC`。
- cursor 为不透明、版本化、URL-safe base64 JSON，包含末项 key、完整 caller scope 摘要和过滤摘要；跨 scope、跨过滤复用或篡改返回 cursor invalid。
- 时间使用 RFC3339 UTC，区间最多 31 天，`started_from` 晚于 `started_to` 或未来偏差超过 5 分钟均非法。
- route / protocol / provider / profile / model / status / failure boundary / usage availability 均为单值精确过滤；字符串必须通过 reference hygiene。
- list 返回 sanitized summary、`next_cursor`、`has_more`、当前 request id 和查询 audit ref；detail 返回完整 sanitized record。
- 空列表是成功；store、scan 或 decode 失败不返回部分结果，也不回退离线 fixture、Workflow history 或 memory sample。

稳定失败码至少包括：

- `gateway_request_history_disabled`
- `gateway_request_scope_denied`
- `gateway_request_record_not_found`
- `gateway_request_filter_invalid`
- `gateway_request_cursor_invalid`
- `gateway_request_store_unavailable`
- `gateway_request_store_contract_mismatch`
- `gateway_request_store_mode_invalid`
- `gateway_request_store_mode_disabled`

原始数据库、bridge 或 provider 错误不得进入 HTTP failure summary。

## Repository、migration 与保留

建立独立 `gatewayRequestStore` contract，不能复用 `workflowRunStore`、Saved Draft repository 或 control-plane fake store。

默认模式：

```text
RADISHMIND_GATEWAY_REQUEST_STORE=memory_dev
```

`memory_dev` 每进程最多保留 500 条，按完整 caller scope 查询，使用 clone-on-write/read，并实现与 PostgreSQL 相同的生命周期、过滤和 cursor contract。

显式 opt-in 模式：

```text
RADISHMIND_GATEWAY_REQUEST_STORE=postgres_dev_test
RADISHMIND_GATEWAY_REQUEST_DEV_TEST_DATABASE_URL=...
RADISHMIND_GATEWAY_REQUEST_DATABASE_TIMEOUT=5s
```

PostgreSQL 使用独立 `gateway_request_schema_versions` 与 `gateway_request_records`，manual migration source 位于 `services/platform/migrations/gateway_requests/`。runtime 只做连接、marker 和 checksum preflight，不执行 DDL；migration role 与 runtime role 分离，运行角色拒绝 DDL。selector 非 allowlist、production mode、连接失败或 marker mismatch 均 fail closed，不回退 memory。

主键覆盖完整 caller scope + request id；索引覆盖 scope + started time + request id，以及常用 route / provider / status / failure / usage 过滤。record JSONB 保存 allowlist document，常用过滤字段物化并在写入时与 JSON 一致性校验。

v1 声明 dev/test 默认保留 14 天、每 scope 最多 50,000 条，但请求路径不隐式删除。schema 提供 retention 索引；显式维护命令后续单独实现，不创建后台清理 goroutine。disposable dev/test 环境由操作者清理或重建。

## Web 审查路径

- 在现有 Model Gateway Evidence Review 增加真实 Request History 入口，不创建新的一级产品面。
- 新增独立 consumer、domain mapping 和 lazy panel；`App.tsx` 只保留装配，不承载查询、过滤或详情逻辑。
- 列表显示 route/protocol、provider/profile/model、status、usage availability、total duration、provider duration、started time 和 failure boundary。
- 详情显示 caller references、selection source、timing breakdown、reported usage、稳定 failure、request / audit ref、store mode 和 stale_started；不展示原始 payload。
- offline source 零请求并明确 `offline evidence`；dev/test source 只读取新资源族。disabled、scope denied、空列表、filter/cursor invalid、store unavailable 分别呈现。
- 现有离线 quota/cost、Workflow trace 和 audit correlation 继续作为背景 evidence，不与真实 Gateway request list 合并或充当 fallback。
- 新 panel 必须保持独立 lazy chunk，并受现有 Vite build 包体预算约束。

## 可观测性与 store failure

- 请求日志记录 request id、route、status、failure code/boundary、selection references、duration、usage availability、store mode、store outcome、caller scope hash 和 audit ref。
- 不记录 caller header 原文、cursor、请求/响应正文、credential、endpoint、raw error 或数据库信息。
- store create / update failure 分别计数；terminal update failure 与 northbound request outcome 分开表达，避免把 provider success 改写成 provider failure。
- read API 记录 operation、filter names、result count、has_more、duration、store mode、stable outcome 和 scope hash。
- caller context unavailable、record stale、usage not reported 和 store failure 都必须可聚合，不通过自由文本诊断。

## 实施拆分

本设计确认需要新的 Platform domain / repository、只读 API 和 PostgreSQL schema，因此由单张 [Model Gateway Request History / Usage & Failure Review v1 任务卡](../../task-cards/model-gateway-request-history-usage-failure-review-v1-plan.md) 承接一个纵向实现批次：

1. caller context、record lifecycle、taxonomy、usage availability 和 memory store。
2. 在三个现有 northbound route 的统一 observability 边界接入 recorder，覆盖 unary / stream / cancellation。
3. list / detail API、scope、filter 和 cursor。
4. 独立 PostgreSQL migration、selector、preflight、integration 和 no-fallback。
5. Web consumer、lazy panel 与真实 Evidence Review 路径。
6. Go / race / PostgreSQL / Web / browser / repository gate 验证和文档收口。

不把这些步骤拆成平行 readiness、refresh、fixture 或 checker 链；标准测试、migration integration、Web build 和聚合门禁足以承载。

## 验收

- Go domain / memory：完整生命周期、终态不可逆、版本并发、scope、clone、FIFO、filter、cursor、stale 和 forbidden fields。
- HTTP recorder：三个 northbound route 的 parse / selection / bridge / provider / translation / stream / cancellation 路径；caller context 缺失不创建 unscoped record。
- read API：strict query、scope、空列表、list/detail 对齐、filter / cursor 篡改、store failure 和无 fixture fallback。
- PostgreSQL：fresh migration、rollback / reapply、runtime DDL 拒绝、重启恢复、并发、scope 隔离、分页、marker mismatch、连接失败与 no fallback。
- Web：offline 零请求、strict mapping、filter、pagination、detail、usage unavailable、失败状态、forbidden-field 拒绝和独立 chunk。
- 浏览器：成功、invalid request、provider failure、queue/timeout/cancel、stream complete/cancel、usage unavailable、分页过滤、详情、Platform 重启恢复和敏感字段缺失；`reported` 只在可信 provider usage contract 成立后另行验收。
- 所有路径确认自动 retry / fallback、quota write、billing write、tool、confirmation、business write 和 replay 为 0。

## 停止线

- 不实现 production API key lifecycle、OIDC Gateway auth、quota enforcement、rate limit、billing、cost ledger 或合规审计账本。
- 不启用自动 retry / fallback、load balancing、动态 provider route、production secret 或 production gateway。
- 不保存 prompt、response、credential、endpoint、provider raw material 或任意可逆内容摘要。
- 不复用 Workflow run repository、旧 fake run API、离线 cost snapshot 或 control-plane fake store 作为请求历史。
- 不在 store 故障时静默 fallback，不因 review store 故障改写 northbound provider outcome。
- 不实现后台 retention executor、replay / resume、自动修复或自动部署。
