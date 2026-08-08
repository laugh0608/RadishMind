# Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1

更新时间：2026-07-26

状态：`admin_provider_route_controlled_activation_dev_test_v1_completed`

## 当前结论

本专题把 Admin Control Plane 从 Provider / Route evidence 只读审查推进为可实际支撑 Gateway 的开发测试态配置流程。平台管理员能够在明确的租户、工作区和环境作用域内维护 Provider Profile assignment 与 Model Route，生成不可变候选版本，完成人工审查，再通过独立的显式动作切换 Gateway 后续请求消费的运行快照。

配置草案、审查结论和激活事实由 Admin Control Plane 持有；Provider 的真实 runtime inventory、credential / endpoint 解析和健康状态继续由既有运行配置与 provider owner 持有；Gateway 只消费编译后的不可变快照，不读取草案，也不成为配置真相源。

本功能只服务内部开发者预览中的 `development | test` 环境，不声明 production provider management、secret backend、自动路由、配额、计费、负载均衡或生产 Gateway 已就绪。

## 当前实现

批次 A 已完成 Go 领域、只读 inventory resolver、内存 repository 和严格完整性重验。批次 B 已补齐共享 SQLite 本地产品持久化、PostgreSQL 开发测试态持久化、独立迁移、三模式 selector 和配置投影。批次 C 已完成 Admin HTTP、verified identity 投影、四项独立权限、显式开发测试态 HTTP / write 门禁和稳定状态映射。当前可在 `tenant + workspace + environment + configuration` 作用域内通过 API 执行草案 CAS 保存、不可变候选生成、独立 review CAS、显式 activation generation CAS，以及回到曾经启用的批准候选；rollback 生成新的 generation，activation history 保持 append-only。

Profile assignment 只接受当前环境下的 `runtime_profile_ref`，candidate 创建与 activation 都解析同一个外部 inventory owner 并核对 capability、enabled 状态和 digest。approval 不创建 active snapshot；inventory 缺失、不可用或发生漂移时不产生 snapshot 和 activation record。草案、候选、snapshot 与 history 在写入返回和后续读取时都会重验 schema、规范化内容、状态关系及 digest。

SQLite 与 PostgreSQL 都持久化草案、候选、独立 review、当前快照和 append-only activation history；快照切换与 activation record 在同一事务提交。SQLite 复用聚合 migration owner 并覆盖真实文件、WAL / SHM 隐私扫描；PostgreSQL 使用独立 marker、checksum、manual migration、迁移 / 运行角色分离和事务级作用域 advisory lock。`memory_dev | sqlite_dev | postgres_dev_test` 互斥选择，配置、migration 或存储不可用时不回退。

管理路由使用既有 verified identity owner，`tenant_ref` 不接受请求覆盖；workspace / environment 请求头、权限、真实 OIDC membership blocker 和写入门禁均在 repository 前校验。写入 body 拒绝未知字段、重复字段、尾随文档和超限载荷；所有管理响应均为 `no-store`，只返回脱敏资源、稳定 failure code、request / audit lineage 与冲突所需的当前版本。Bridge inventory resolver 只摘要化消费既有 inventory，不复制 endpoint 或 credential。

相邻测试覆盖完整 HTTP activate 流程、权限独立投影、身份 / membership / 环境拒绝、strict JSON、HTTP 状态、inventory 摘要确定性、完整 activate / replace / rollback、两次重启恢复、draft / review / generation 并发 CAS、append-only 保护、配置拒绝、selector no fallback、数据库与 WAL 隐私扫描。真实 PostgreSQL 专项链已覆盖 Admin HTTP 写入 / 读取、运行角色 DDL 拒绝、多连接并发激活 CAS、服务重开恢复和 configured profile；Platform 全量 Go 测试与 `go vet` 已通过。

批次 E 已在既有 Provider Deployment Review 中加入受控配置工作区：页面能够维护草案、创建不可变候选、查看候选与当前快照差异、记录独立审查、显式激活、选择历史候选回滚，并读取当前快照与 append-only generation history。严格 Web consumer 固定四项权限、作用域、CAS 元数据、响应 schema 和敏感字段拒绝；offline 模式保持零请求。

SQLite 本地产品浏览器连续完成 `draft → candidate-one → approve → generation 1 → Gateway invocation → Request History lineage → candidate-two → generation 2 → rollback candidate-one → generation 3`，服务重启后恢复 generation 3 与三条历史。Gateway 调用在没有真实 Provider credential / endpoint 的停止线下命中精确快照后返回 `GATEWAY_INFERENCE_FAILED`；Request History 仍准确展示 `configuration_id`、generation 与 snapshot digest，不把失败误写成 route miss，也不保存请求正文或凭据。

同一启动器已提供 PostgreSQL 开发测试态产品模式，统一预检 Admin Provider Route、API key、Gateway Request History migration 和 API key auth。真实 PostgreSQL 浏览器完成草案、候选、审批与 generation 1 激活，服务重启后恢复当前快照和一条历史；完整 PostgreSQL integration、migration 恢复、Web 237 项测试与生产构建均通过。

## 用户与真实需求

目标用户是需要为内部应用配置模型调用路径的平台管理员和运维人员。现有 Admin 页面可以查看 provider/profile readiness、route evidence 和部署阻塞项，但不能形成受控修改；Gateway 当前主要依赖进程配置中的单一 provider/profile/model，缺少可审查、可追溯、可回滚的配置版本。

用户需要完成以下连续流程：

1. 选择明确的租户、工作区和开发测试环境。
2. 创建或修改一份 Provider Profile / Model Route 配置草案。
3. 由服务端校验 Profile 引用、capability、模型路由和既有 provider inventory。
4. 从确定的草案 revision 生成不可变候选版本。
5. 由具备独立权限的管理员批准或拒绝候选。
6. 批准只形成可激活资格，不改变 Gateway 行为。
7. 具备激活权限的管理员以预期 generation 显式启用候选。
8. Gateway 后续请求固定消费新的不可变快照；在途请求继续使用启动时取得的旧快照。
9. 管理员能够把运行配置原子回滚到曾经启用且仍然有效的历史批准版本。
10. 草案、候选、审查、激活、回滚和 Gateway 请求都能按版本与摘要建立脱敏追踪。

## 真相源与职责边界

### Admin Control Plane

- 保存配置草案及其单调 revision。
- 保存不可变候选、人工审查记录和 append-only 激活记录。
- 保存每个作用域当前激活的不可变快照与 generation。
- 校验 Provider Profile assignment 是否精确引用既有 runtime inventory。
- 不保存 credential value、原始 endpoint、Authorization header、环境变量值或 provider raw config。

### Provider runtime inventory

- 继续拥有 provider identity、runtime profile identity、支持的 capability、环境和可用状态。
- 由现有配置 / bridge owner 暴露只读解析边界。
- Admin 不复制 inventory，不允许以草案声明覆盖 inventory 事实。
- inventory 不存在、已禁用、环境不匹配或摘要漂移时，候选生成或激活失败关闭。

### Model Gateway

- 只读取已激活的 `ProviderRouteSnapshot`。
- 单个请求取得快照后固定 generation 与 digest，不在执行中途重新选取。
- 明确路由无效时失败，不回退到全局默认 provider/profile。
- 请求历史只记录脱敏的配置标识、generation 和 digest，不保存草案或 runtime 原始配置。

## 配置作用域

v1 的原子配置作用域固定为：

- `tenant_ref`
- `workspace_id`
- `environment`，只允许 `development | test`
- `configuration_id`

Provider Profile assignment 与 Model Route 在同一作用域内一起版本化和激活。v1 不实现跨工作区 Profile 共享、跨环境复制或全局默认配置；后续若需要 tenant-wide profile catalog，必须单独设计继承、授权和冲突规则。

## 领域模型

### 配置草案

`admin_provider_route_configuration_draft.v1` 至少包含：

- 作用域字段和 `configuration_id`
- `draft_revision`，首次保存为 `1`，后续 CAS 更新单调递增
- `display_name`
- 排序稳定的 `provider_profiles`
- 排序稳定的 `model_routes`
- `draft_digest`
- 创建 / 更新时间、操作者、request id 与 audit ref

Provider Profile assignment 包含：

- `profile_id`
- `display_name`
- `provider_id`
- `runtime_profile_ref`
- 排序去重后的 `capabilities`

`runtime_profile_ref` 只允许：

```text
ref:radishmind/<environment>/provider-profiles/<profile-key>
```

该引用只定位既有 inventory 项，不承载 URL、路径、credential 或 provider config。

Model Route 包含：

- `route_id`
- `protocol`：`chat_completions | responses | messages`
- `model_id`
- `provider_profile_id`

协议与所需 capability 固定映射；Route 引用的 Profile 必须声明该 capability，inventory 也必须确认支持。相同 `protocol + model_id` 在同一配置中只能有一条路由。

### 不可变候选

`admin_provider_route_candidate.v1` 固定保存：

- `candidate_id`
- 来源 `configuration_id`、`draft_revision` 与 `draft_digest`
- 完整、规范化的配置快照
- 每个 Profile 对应的 inventory digest
- `candidate_digest`
- `candidate_state`
- `review_version`
- 创建事实和最多一条终态审查记录

状态只允许：

```text
pending_review -> approved
pending_review -> rejected
```

候选创建后不可原地修改。草案继续变化不会改变既有候选；需要纳入新草案内容时创建新的候选。

### 激活快照与记录

`admin_provider_route_snapshot.v1` 包含：

- 配置作用域
- `generation`
- 来源 candidate id / digest
- 完整规范化 Profile / Route 配置
- inventory bindings / digests
- `snapshot_digest`
- 激活时间、操作者、request id 与 audit ref

`admin_provider_route_activation_record.v1` 是 append-only 记录，包含 action、来源和目标 generation / candidate、前一快照引用、前一记录摘要、当前记录摘要及操作事实。记录摘要形成连续 hash chain；action 只允许 `activate | rollback`。

批准和激活严格分离：

- `approve` 不产生 snapshot，不增加 generation。
- 激活必须携带 `expected_generation`；首次激活要求 `0`。
- 并发激活只有一个请求能够成功。
- 回滚目标必须是历史上成功启用过、当前仍为 approved 且 inventory binding 未漂移的候选。
- 回滚生成新的 generation，不把 generation 倒退，也不修改历史记录。

## 严格校验与敏感材料边界

- 标识符、展示名、模型名、列表数量和规范 JSON 都有明确预算。
- Profile / Route 标识符必须唯一；Route binding 必须完整且无歧义。
- 所有输入拒绝 URL、Authorization、Bearer、cookie、token、API key、password、DSN、私钥和环境变量赋值形态。
- `runtime_profile_ref` 必须匹配当前环境，不能包含路径穿越、查询字符串或 fragment。
- candidate 创建与 activation 都重新解析 inventory；activation 要求解析结果与候选保存的 digest 精确一致。
- 任何校验、repository 或 inventory 失败都不得生成部分 candidate、review、snapshot 或 activation record。
- 公开结果、错误、日志和审计不得返回 inventory 原始载荷或敏感配置。

## 权限与动作分离

开发测试态权限固定拆分为：

- `admin_provider_routes:read`
- `admin_provider_routes:draft`
- `admin_provider_routes:review`
- `admin_provider_routes:activate`

草案写入、审查和激活是三个独立授权面。v1 允许同一测试身份同时具备多项权限，但服务端不能把权限合并为一个布尔写入开关，也不能在创建候选时自动批准或启用。

真实 Radish workspace membership 尚未成立时，相关 HTTP API 继续失败关闭；不得用开发请求头结果反推 production authorization 已就绪。

## 稳定失败语义

至少固定：

- `admin_provider_route_disabled`
- `admin_provider_route_scope_denied`
- `admin_provider_route_payload_invalid`
- `admin_provider_route_sensitive_material_forbidden`
- `admin_provider_route_environment_forbidden`
- `admin_provider_route_draft_not_found`
- `admin_provider_route_draft_revision_conflict`
- `admin_provider_route_candidate_not_found`
- `admin_provider_route_candidate_conflict`
- `admin_provider_route_review_version_conflict`
- `admin_provider_route_review_transition_invalid`
- `admin_provider_route_candidate_not_approved`
- `admin_provider_route_inventory_not_found`
- `admin_provider_route_inventory_mismatch`
- `admin_provider_route_inventory_unavailable`
- `admin_provider_route_generation_conflict`
- `admin_provider_route_already_active`
- `admin_provider_route_rollback_target_invalid`
- `admin_provider_route_store_unavailable`

版本冲突返回当前 revision / review version / generation，但不返回其他作用域数据。inventory not found 与 mismatch 只给稳定码，不泄露相邻 inventory 项。

## Admin HTTP 契约

批次 C 在既有 verified identity 基础上固定以下管理路由：

- `GET|PUT /v1/admin/provider-route-configurations/{configuration_id}`
- `POST /v1/admin/provider-route-configurations/{configuration_id}/candidates`
- `GET /v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}`
- `POST /v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}/reviews`
- `POST /v1/admin/provider-route-configurations/{configuration_id}/candidates/{candidate_id}/activations`
- `GET /v1/admin/provider-route-configurations/{configuration_id}/active-snapshot`
- `GET /v1/admin/provider-route-configurations/{configuration_id}/activation-history`

`tenant_ref` 只来自 verified identity；开发测试态 workspace 与 environment 分别由
`X-RadishMind-Dev-Admin-Provider-Route-Workspace` 和
`X-RadishMind-Dev-Admin-Provider-Route-Environment` 显式传入，并在进入 repository 前校验。路由不接受额外 query 参数，也不允许 body 重写 tenant、workspace、environment、actor、request id 或 audit ref。

写入 body 固定为：

- 草案：`expected_revision`、`display_name`、`provider_profiles`、`model_routes`
- 候选：`candidate_id`、`expected_draft_revision`
- 审查：`expected_review_version`、`decision`、`reason`
- 激活 / 回滚：`expected_generation`、`action`、`reason`

所有 body 使用有界 strict JSON，未知字段、重复对象、尾随内容和超限载荷在领域与 repository 前拒绝。管理响应统一返回 request / audit lineage、作用域、对应领域资源、稳定 `failure_code` 和冲突时的当前 revision / review version / generation；activation history 缺省为空数组。所有成功与失败响应都设置 `Cache-Control: no-store`。

HTTP 状态固定分层：缺失 / 无效身份为 `401`，权限或环境禁止为 `403`，资源不存在为 `404`，CAS / 状态冲突为 `409`，领域资格不满足为 `422`，inventory / store / membership unavailable 为 `503`，strict payload 错误为 `400`。真实 OIDC workspace membership 未成立时必须在 repository 前返回 `workspace_membership_unavailable`。不得复用 northbound Gateway 路由承载管理动作。

## 持久化方向

按顺序支持：

1. `memory_dev`：领域行为、CAS、并发和失败语义基线。
2. `sqlite_dev`：本地产品档、重启恢复和真实文件隐私扫描。
3. `postgres_dev_test`：迁移 / 角色分离、稳定读取、事务 CAS、并发激活和服务重启。

三种模式互斥；存储不可用时不得回退到其他模式。PostgreSQL 继续采用独立 schema marker、manual migration、checksum 和 advisory lock，平台启动不自动执行 DDL。

## 实施批次

### 批次 A：领域、内存 repository 与状态转换

- 已完成本文件和单一专项任务卡。
- 已实现严格领域类型、规范化、digest、inventory resolver 边界。
- 已实现草案 CAS、不可变候选、独立 review、generation CAS、activate / rollback。
- 已实现 append-only activation history 和 active snapshot。
- 已覆盖权限、敏感材料、作用域隔离、inventory 漂移、并发和无部分写入测试。

### 批次 B：SQLite 与 PostgreSQL 开发测试态持久化

- 已完成独立表、索引、migration、selector 和三模式 repository parity。
- 已验证重启恢复、并发 CAS、运行 / 迁移角色分离、DDL 拒绝、no fallback。
- 已扫描数据库、WAL、错误和日志，确认不持久化敏感材料。

### 批次 C：Admin API 与身份权限

- 已接入 verified identity 与四项独立权限，权限不会隐含应用或 Gateway 写权限。
- 已提供草案、候选、审查、激活、回滚、当前快照和历史 API。
- 已固定 HTTP failure mapping、strict payload、no-store、request / audit lineage 与脱敏诊断。
- `RADISHMIND_ADMIN_PROVIDER_ROUTE_DEV_HTTP` 与 `RADISHMIND_ADMIN_PROVIDER_ROUTE_DEV_WRITE` 默认关闭且具有启动依赖；OIDC membership 未成立时在 repository 前失败关闭。
- 已覆盖内存 HTTP 连续链和真实 PostgreSQL Admin HTTP 持久化链；下一步只进入批次 D Gateway snapshot consumer，不提前打开 Web。

### 批次 D：Gateway 不可变快照消费

- 已引入只读 snapshot provider interface，Gateway 不读取 Admin 草案、候选、审查或历史 repository。
- `static_config` 保持默认；显式 `admin_snapshot_dev_test` 要求 API key 开发测试态认证、Request History、环境和配置键完整，两个来源不在单次请求内混用。
- 三协议已按 `protocol + model` 精确匹配已激活 route，并在请求开始时固定 configuration、generation、digest、runtime profile 与 resolved model。
- snapshot / route / profile 缺失、override 和 inventory digest 漂移均在 bridge 前失败，不回退静态配置；后续 activation 只改变下一请求。
- Request History 已在 memory、SQLite 与 PostgreSQL 保存脱敏 configuration / generation / digest；三协议、原子切换、在途固定、重启恢复、零 bridge 调用和 race 验证通过。

### 批次 E：Admin Web 与连续产品验证

- 已完成草案编辑、候选差异、审查、激活、回滚、当前快照和历史界面。
- `S7 R1` 已把 Provider、Profile 与 Route 作为同一原子 owner 的三个资源入口纳入 Admin Control Plane 连续工作面；入口只改变当前任务焦点，不复制 draft、candidate、review、activation、history 或 runtime inventory 状态。
- 已完成 SQLite 与 PostgreSQL 开发测试态启动模式、专项门禁、服务重启和真实浏览器连续验收。
- 已让 Gateway Playground 与 Request History 展示精确的激活快照 generation / digest 谱系。
- 已更新当前焦点、路线图、能力矩阵、Admin / Gateway 专题与周志。

## 验收

- approval 对 Gateway 零影响，activation 才改变后续请求。
- 相同草案产生确定性 digest；候选和快照不可变。
- stale draft、review 和 generation 写入均失败且不产生部分事实。
- inventory 缺失、禁用、环境 / capability / digest 漂移在 side effect 前失败。
- rollback 生成新 generation，并能追溯到历史批准候选。
- 并发 activation 只有一个成功；读操作始终看到完整旧快照或完整新快照。
- memory / SQLite / PostgreSQL 领域行为一致。
- Gateway 调用能够关联准确 snapshot generation / digest，且不保存敏感配置。
- Admin Web 能连续完成编辑、候选、审查、激活、调用验证和回滚。
- 相邻测试、Go race、Web 测试 / build、聚合产品 smoke、fast 与最终 full 仓库检查通过。

## 停止线

- 不实现 production credential resolver、secret backend、真实 endpoint 管理或 Provider 账号生命周期。
- 不实现 production environment、自动 activation、定时发布、动态发现、retry / fallback、负载均衡或跨环境复制。
- 不实现 quota、rate limit、billing、price 或 cost ledger。
- 不修改 northbound Gateway schema，不创建第二套 provider inventory 或 selection policy。
- 不接真实 Radish OIDC，不放开 workspace membership。
- 不把开发测试态激活解释为生产部署、production Gateway 或 provider SLA。
- 不为普通页面和既有测试可承载的行为新增同层 checker / readiness 链。
