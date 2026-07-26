# Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1

更新时间：2026-07-26

状态：`admin_provider_route_controlled_activation_dev_test_v1_batch_b_completed_batch_c_ready`

## 当前结论

本专题把 Admin Control Plane 从 Provider / Route evidence 只读审查推进为可实际支撑 Gateway 的开发测试态配置流程。平台管理员能够在明确的租户、工作区和环境作用域内维护 Provider Profile assignment 与 Model Route，生成不可变候选版本，完成人工审查，再通过独立的显式动作切换 Gateway 后续请求消费的运行快照。

配置草案、审查结论和激活事实由 Admin Control Plane 持有；Provider 的真实 runtime inventory、credential / endpoint 解析和健康状态继续由既有运行配置与 provider owner 持有；Gateway 只消费编译后的不可变快照，不读取草案，也不成为配置真相源。

本功能只服务内部开发者预览中的 `development | test` 环境，不声明 production provider management、secret backend、自动路由、配额、计费、负载均衡或生产 Gateway 已就绪。

## 当前实现

批次 A 已完成 Go 领域、只读 inventory resolver、内存 repository 和严格完整性重验。批次 B 已补齐共享 SQLite 本地产品持久化、PostgreSQL 开发测试态持久化、独立迁移、三模式 selector 和配置投影。当前可在 `tenant + workspace + environment + configuration` 作用域内执行草案 CAS 保存、不可变候选生成、独立 review CAS、显式 activation generation CAS，以及回到曾经启用的批准候选；rollback 生成新的 generation，activation history 保持 append-only。

Profile assignment 只接受当前环境下的 `runtime_profile_ref`，candidate 创建与 activation 都解析同一个外部 inventory owner 并核对 capability、enabled 状态和 digest。approval 不创建 active snapshot；inventory 缺失、不可用或发生漂移时不产生 snapshot 和 activation record。草案、候选、snapshot 与 history 在写入返回和后续读取时都会重验 schema、规范化内容、状态关系及 digest。

SQLite 与 PostgreSQL 都持久化草案、候选、独立 review、当前快照和 append-only activation history；快照切换与 activation record 在同一事务提交。SQLite 复用聚合 migration owner 并覆盖真实文件、WAL / SHM 隐私扫描；PostgreSQL 使用独立 marker、checksum、manual migration、迁移 / 运行角色分离和事务级作用域 advisory lock。`memory_dev | sqlite_dev | postgres_dev_test` 互斥选择，配置、migration 或存储不可用时不回退。

相邻测试覆盖完整 activate / replace / rollback、两次重启恢复、draft / review / generation 并发 CAS、append-only 保护、配置拒绝、selector no fallback、数据库与 WAL 隐私扫描。真实 PostgreSQL 专项链已覆盖运行角色 DDL 拒绝、多连接并发激活 CAS、服务重开恢复和 configured profile；Platform 全量 Go 测试与 `go vet` 已通过。

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

## API 方向

后续 HTTP 批次在既有 verified identity 基础上提供：

- `GET|PUT /v1/admin/provider-route-configurations/{configuration_id}`
- `POST /v1/admin/provider-route-configurations/{configuration_id}/candidates`
- `GET /v1/admin/provider-route-candidates/{candidate_id}`
- `POST /v1/admin/provider-route-candidates/{candidate_id}/reviews`
- `POST /v1/admin/provider-route-candidates/{candidate_id}/activations`
- `GET /v1/admin/provider-route-configurations/{configuration_id}/active-snapshot`
- `GET /v1/admin/provider-route-configurations/{configuration_id}/activation-history`

具体 payload 与 HTTP 状态在 API 批次冻结。不得复用 northbound Gateway 路由承载管理动作。

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

- 接入 verified identity 与四项独立权限。
- 提供草案、候选、审查、激活、回滚和历史 API。
- 固定 HTTP failure mapping、strict payload、no-store 与脱敏诊断。
- OIDC membership 未成立时继续失败关闭。

### 批次 D：Gateway 不可变快照消费

- 引入只读 snapshot provider interface，不让 Gateway 读取 Admin repository。
- 请求开始时固定 generation / digest，历史写入脱敏 lineage。
- 验证原子切换、在途请求固定、无效路由零 bridge 调用和不回退。
- 保留现有静态配置模式作为明确选择，不在单次请求内混用两个来源。

### 批次 E：Admin Web 与连续产品验证

- 完成草案编辑、候选差异、审查、激活、回滚、当前快照和历史界面。
- 完成本地 SQLite 连续链、PostgreSQL 专项链、服务重启和浏览器连续验收。
- 更新当前焦点、路线图、能力矩阵、Admin / Gateway 专题与周志。

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
