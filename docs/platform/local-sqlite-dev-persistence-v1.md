# 本地 SQLite 开发持久化 v1

更新时间：2026-07-14

状态：`local_sqlite_dev_persistence_v1_s3_local_product_chain_completed`

## 当前结论

RadishMind 增加 `sqlite_dev`，用于无需 PostgreSQL 服务即可完成本地产品开发、前后端联调和服务重启恢复。它与现有 `memory_dev`、`postgres_dev_test` 并行，不替代任何一层：

| 模式 | 默认用途 | 持久化 | 外部服务 | 能证明什么 |
| --- | --- | --- | --- | --- |
| `memory_dev` | 单元测试、离线页面、快速回归 | 否 | 不需要 | 领域规则、HTTP 契约和无副作用边界 |
| `sqlite_dev` | 本地产品启动、日常连续开发、浏览器联调 | 是 | 不需要 | 本地完整链路、重启恢复和开发数据连续性 |
| `postgres_dev_test` | 批次验收、CI、部署同构验证 | 是 | PostgreSQL | PostgreSQL migration、角色权限、并发和方言语义 |

基础配置和普通测试继续以 `memory_dev` 为安全默认，避免测试进程隐式写文件。仓库提供的本地产品启动档默认选择 `sqlite_dev`；只有显式数据库验收才选择 `postgres_dev_test`。未知模式、文件不可用、schema 不兼容、迁移失败或查询失败都必须显式失败，不得回退其他模式。

本专题只定义开发态持久化。它不声明 production repository、生产数据库、正式 Radish 身份 / 成员关系、生产凭据后端、生产审计账本或公开生产 API 已就绪。

2026-07-14 已完成 S1：锁定 BSD-3-Clause、无 CGO 且与仓库 Go 1.25 对齐的 `modernc.org/sqlite v1.53.0`；新增共享 SQLite runtime、受控文件权限、逐连接 foreign keys / WAL / busy timeout、完整性检查、checkpoint / 关闭回收，以及带 component、migration id、store schema version、checksum 和应用时间的事务 migration marker。聚合配置已经可识别 `sqlite_dev` 并输出脱敏摘要；七组 repository 未齐备前，配置检查保持 `sqlite_dev_repository_set` 未满足，平台启动在任何 store 构造前明确失败。

2026-07-14 已完成 S2 的应用目录纵向切片：应用目录拥有独立 SQLite migration 与严格 JSON 文本表，repository 复用既有领域服务、所有者作用域、稳定游标、不可逆归档和预期版本 CAS；组件 selector 只接受已应用对应 migration 的共享 runtime，缺少 runtime、缺少或不匹配 marker、数据库关闭和查询故障均失败关闭，不创建私有连接或回退内存。

同日继续完成应用配置草案与发布候选：两组件分别拥有独立 migration 和 repository，但共享一个 runtime；草案保存保持创建元数据与版本 CAS，候选保持不可变创建、只追加审查、审查 CAS、终态和晋级阻塞语义。发布 selector 同时复验草案与候选 migration，避免依赖组件缺失时构造半成品。memory / SQLite 复用同组领域契约，真实文件测试覆盖跨组件重启、并发单写者、作用域隔离、稳定排序、关闭失败和敏感材料禁入。S2 当前完成 3/7，下一项固定为 API 密钥；聚合 `sqlite_dev` 继续保持关闭。

2026-07-14 已完成 S2 的 API 密钥组件。该组件同时复验应用目录依赖，使用整数纳秒承担过期、最近使用和分页时间谓词，并保持原始令牌只出现在创建成功响应。签发、筛选分页、Gateway 认证、最近使用单调更新、吊销 CAS、认证 / 吊销竞态、两次重启恢复、关闭失败和数据库 / WAL / 共享内存令牌扫描均已形成组件证据。

同日完成 Gateway 请求历史 SQLite repository。该组件保持完整 caller scope、终态不可逆、筛选游标、脱敏记录和 review store / provider outcome 边界；整数纳秒承担时间范围与游标谓词，同一时刻用 request id 稳定排序。真实文件证据覆盖取消终态、重启恢复、关闭不回退、损坏文档失败关闭、模型输入输出禁入，以及与应用目录和 API 密钥共享 runtime 的可信 northbound 调用。该批结束时 S2 完成 5/7，下一项固定为工作流草案。

随后完成工作流草案 SQLite repository。实现复用既有 domain service、`SavedWorkflowDraftRepositoryAdapter`、actor scope 与 schema preflight，以独立 STRICT 表保存脱敏 document，并用整数纳秒承担列表顺序和物理时间复验。数据库条件更新保证预期版本 CAS，损坏时间或 document 会拒绝整次读取 / 列表。专项证据覆盖 16 路并发单写者、tenant / workspace / application / owner 隔离、稳定等时刻排序、HTTP 保存 / 读取、重启恢复、关闭不回退、marker mismatch 和敏感内容禁入。S2 当前完成 6/7，下一项固定为工作流运行；七组件齐备前不开放聚合 `sqlite_dev`。

最后完成工作流运行 SQLite repository。独立 STRICT 表同时保存脱敏 JSON 与筛选列，整数纳秒承担时间区间、stale running、游标和排序；读取复验 JSON 与物理列，create / update 使用完整 scope、预期版本、原始开始时间和 `running` 状态谓词保证终态单写者。专项证据覆盖跨模式筛选、等时刻排序、16 路终态竞争、完整 scope 隔离、真实 executor、重启恢复、关闭不回退、marker mismatch、损坏记录无部分列表和原始输入禁入。S2 七组 repository 已全部完成；聚合启动仍保持关闭，下一批统一接入共享 runtime 和启动生命周期。

S2 聚合 runtime 批次已经完成：聚合 `sqlite_dev` 一次性派生七个组件的有效 store mode，按应用目录、应用配置草案、应用发布候选、API 密钥、Gateway 请求历史、工作流草案、工作流运行的固定顺序应用 migration，再把同一数据库句柄注入全部 repository。平台在构造任何 SQLite repository 前完成配置与 migration 校验；后续 repository 或 bridge 构造失败时，已构造资源和共享 runtime 按逆序回收；正常关闭时七个 repository 先停止使用连接，共享 runtime 再执行 checkpoint 和连接关闭，并由 `Server` 的单次关闭语义保证不会重复回收。真实文件聚合测试已覆盖七个 marker、七种 selector、七类记录跨 `Server` 重启恢复、关闭后连接不可用、migration marker 不兼容和 bridge 启动失败后的干净重启。

S3 本地产品档已完成：Shell / PowerShell wrapper 默认使用 `local-product`，统一注入聚合 `sqlite_dev`、仓库根受控数据库路径和七组件开发门禁；`configured` 只承接调用方显式配置，供 PostgreSQL 专项验收和故障注入使用。现有部署 smoke 已验证两个启动档、七组件配置投影、路径脱敏、校验阶段不创建数据库、组件配置冲突和未知档退出码。真实 HTTP 连续链在同一应用作用域内完成应用创建、配置草案、发布候选审查、API 密钥签发、Bearer Gateway 调用、脱敏请求历史、工作流草案和运行，并在 `Server` 重启后逐项恢复；原始令牌、Gateway 输入和工作流输入未进入数据库、WAL 或共享内存。随后真实 PostgreSQL 专项门禁已独立通过，不以 SQLite 结果替代 migration、角色、方言或并发证据；下一步进入 API 密钥 Web 一次性交接与浏览器连续验收。

Driver 评审证据以 [`modernc.org/sqlite` 官方 package 页面](https://pkg.go.dev/modernc.org/sqlite)、[`v1.53.0` 模块声明](https://gitlab.com/cznic/sqlite/-/raw/v1.53.0/go.mod)和[许可证原文](https://gitlab.com/cznic/sqlite/-/raw/v1.53.0/LICENSE)为准。`github.com/mattn/go-sqlite3` 因明确要求 CGO 与 GCC，且 Windows / 交叉编译需要额外工具链，本阶段不采用；该结论只服务当前本地开发 runtime，不构成永久排除其它 driver 的平台政策。

具体实现统一由[本地 SQLite 开发持久化 v1 实施任务卡](../task-cards/local-sqlite-dev-persistence-v1-plan.md)承接，不再为 driver、单个 repository 或 migration 派生同层任务卡与检查器。

## 为什么现在补齐

现有 `memory_dev` 适合快速测试，却不能保留应用、配置、密钥、草案、运行和请求历史；现有 `postgres_dev_test` 能验证真实数据库语义，但日常联调需要额外启动 PostgreSQL。API 密钥专题已经把“应用 → 密钥 → Gateway → 请求历史”推进到持久化边界，此时若继续只依赖 PostgreSQL，本地开发体验会长期缺少零外部服务的连续链路。

`sqlite_dev` 解决的是本地产品开发体验，不是数据库兼容性替代。PostgreSQL 专属能力仍必须由真实 PostgreSQL 验证，不能用 SQLite 结果代替。

## 数据范围

v1 覆盖 RadishMind 自身拥有的七组开发运行数据：

1. 已保存工作流草案 `workflow_saved_drafts`；
2. 应用配置草案 `application_configuration_drafts`；
3. 应用发布候选 `application_publish_candidates`；
4. 应用目录 `application_catalog_records`；
5. API 密钥记录 `api_key_records`；
6. 工作流运行记录 `workflow_runs`；
7. Gateway 脱敏请求历史 `gateway_requests`。

这些组件必须在同一个本地启动档中整体选择 `sqlite_dev`，不允许正式入口长期混用 SQLite、内存和 PostgreSQL。组件级 selector 仍可供单元测试和专项故障注入使用，但用户不需要逐项设置七个环境变量。

以下数据不进入本地 SQLite：

- Radish 身份、成员关系、租户和上层业务真相；
- Control Plane Tenant / Audit 的外部只读投影；
- provider credential、生产 secret、DSN、OIDC token、cookie 或 Authorization；
- 模型权重、数据集、图片二进制、完整提示词 / 响应正文；
- 生产审计账本、配额、计费和成本台账。

Control Plane Read 继续使用 `fake_store_dev` 或显式 `postgres_dev_test`。本专题不新增 SQLite 身份真相，也不把测试投影持久化成正式数据。

## 本地数据库与生命周期

正式本地路径固定为仓库运行数据目录下的单文件数据库：

```text
var/sqlite-dev/radishmind.db
```

`var/sqlite-dev/` 必须进入 `.gitignore`，数据库、WAL、共享内存和临时文件均不得提交。路径可通过单一显式配置覆盖，但配置摘要只报告模式、路径是否配置和 schema 状态，不输出用户主目录或其它不必要的绝对路径。

平台进程持有一个共享 SQLite runtime，由它统一负责：

- 创建受控父目录和数据库文件；
- 设置本地文件权限；
- 打开连接、启用 foreign keys、WAL 和受控 busy timeout；
- 顺序应用已评审的本地 schema migration；
- 向七个 repository 注入共享数据库句柄；
- 在服务关闭时完成 checkpoint 和连接回收。

聚合启动采用单一所有权：组件 factory 在 `sqlite_dev` 下只验证共享 runtime 和自己的 migration marker，不创建也不关闭私有数据库连接；`Server` 是共享 runtime 的唯一生命周期所有者。启动阶段任一 migration、marker、repository 或 bridge 失败都必须关闭已经打开的资源并返回原始失败，不留下仍持有数据库文件的半启动服务。

SQLite 不是独立服务，不增加后台守护进程。正式本地启动入口可以自动应用兼容的前滚 migration，以保持开箱即用；marker、checksum 或 schema 版本不兼容时必须停止启动。数据库重置是独立显式动作，不得由普通启动、测试失败或版本不匹配自动删除文件。

## Schema 与迁移边界

七个组件继续拥有各自的表、索引、schema 版本和迁移记录。共享 SQLite runtime 只负责连接与编排，不把领域表合并成通用键值表，也不引入无法审查的自动建表反射。

SQLite migration 与 PostgreSQL migration 使用同一领域版本目标，但允许物理 SQL 不同：

- PostgreSQL `jsonb` 在 SQLite 中保存为严格校验的 JSON 文本；
- PostgreSQL `text[]` 在 SQLite 中使用受控 JSON 数组或明确关系表，领域层继续负责排序、去重和允许值校验；
- 时间统一保存为 UTC `RFC3339Nano` 文本，比较和游标必须有跨模式契约测试；
- PostgreSQL tuple comparison 改写为等价的显式条件；
- PostgreSQL advisory lock、角色和 DDL 权限边界不在 SQLite 中伪造，仍由 PostgreSQL 集成门禁验证；
- SQLite 写并发使用事务、唯一约束和版本谓词实现 CAS，不以进程锁替代数据库原子性。

本地 migration 必须可重复检查、可从空文件建立完整 schema，并保存 component、migration id、store schema version、checksum 和 applied time。已应用 migration 不允许被静默改写；需要演进时新增 migration。

## Repository 一致性

三种模式共享既有领域服务、验证、失败码、所有者作用域和响应投影。SQLite repository 不能复制或放宽领域规则，也不能直接向 HTTP 层暴露 SQL 错误。

每个组件至少保持以下一致性：

- tenant、workspace、application、owner predicate 一致；
- create / update / archive / revoke 的唯一键与预期版本语义一致；
- list 的过滤、顺序、游标和分页边界一致；
- 时间、有效期、最近使用时间和终态投影一致；
- sanitized payload 与敏感字段禁入边界一致；
- 数据库关闭、锁等待、损坏或 schema 不兼容时失败关闭；
- 任何存储故障都不得回退到 `memory_dev`、fixture 或另一数据库。

实现应优先复用 repository contract 与领域类型，不以统一大而全 repository、动态 SQL factory 或多层转发包装追求表面复用。只有连接生命周期、migration 编排、错误归一化和通用 SQLite 设置进入共享平台层。

## API 密钥安全边界

SQLite API 密钥表只保存固定长度摘要和脱敏记录，绝不保存原始令牌。签发响应仍是原始令牌唯一出现位置；数据库文件、WAL、备份、日志、错误、请求历史和后续读取均不得包含令牌。

本地数据库是开发者工作数据，不是生产凭据后端。文件权限、目录忽略和敏感信息扫描必须进入验收，但这些措施不能被解释为生产 secret storage、pepper、轮换或集中吊销基础设施。

## 配置与启动体验

平台新增聚合本地持久化配置，用于一次选择完整本地链路：

```text
RADISHMIND_LOCAL_PERSISTENCE_MODE=sqlite_dev
RADISHMIND_SQLITE_DEV_DATABASE_PATH=var/sqlite-dev/radishmind.db
```

本地启动 wrapper 默认注入 `sqlite_dev` 及现有明确的 dev-only HTTP / write gates；直接运行测试和未选择本地产品档时仍使用 `memory_dev`。现有各组件 `*_STORE` 配置继续存在：

- 聚合模式负责正式本地启动的一致选择；
- 组件配置只允许在测试或明确诊断时覆盖；
- 聚合模式与组件配置冲突时启动失败，不允许猜测优先级；
- `postgres_dev_test` 继续使用独立 DSN、手动 PostgreSQL migration 和运行 / 迁移角色。

跨平台 wrapper 固定提供两个启动档：

- `local-product` 是默认档，注入聚合 `sqlite_dev`、仓库根 `var/sqlite-dev/radishmind.db` 的受控路径、七组件所需开发门禁；调用方显式提供数据库路径时保留该覆盖，但不得同时提供组件 `*_STORE` 形成第二配置源；
- `configured` 不注入聚合持久化模式和组件门禁，只承接调用方的文件 / 环境配置，用于 PostgreSQL 专项验收、组件故障注入和历史兼容诊断；它不代表生产启动档。

`config-summary` 与 `config-check` 在两个启动档下都只加载和校验配置，不创建数据库文件或应用 migration。只有 `serve` 进入 `Server` 生命周期后才打开 SQLite。未知启动档必须返回退出码 `2`；Shell 与 PowerShell 的默认值、命令集合、错误语义和受控路径必须一致。

配置摘要新增本地持久化模式、数据库文件是否配置、schema 状态和组件选择一致性，不输出数据库内容、API 密钥摘要或用户不需要的绝对路径。

聚合 `sqlite_dev` 的有效配置必须把七个组件 store mode 全部投影为 `sqlite_dev`，并复用现有 dev-only HTTP / write / executor / history gates；摘要按这组有效配置报告所需门禁。用户若显式设置任一组件 `*_STORE`，即使值同为 `sqlite_dev`，也视为与聚合档冲突并停止启动，避免形成两个配置真相源。

## 验证矩阵

### 普通开发门禁

- 领域与 HTTP 单元测试继续使用 `memory_dev`；
- SQLite repository 测试使用临时目录和真实 SQLite 文件，不依赖 Docker；
- 每个组件执行 create / read / list / update 或终态操作、重开连接恢复、分页和 no-fallback；
- API 密钥增加摘要不出公开投影、认证 / 吊销并发和 WAL 敏感信息扫描；
- 聚合本地启动档验证七个组件没有混用 store mode。

### 跨模式契约

同一组 repository contract cases 至少在 `memory_dev` 与 `sqlite_dev` 上运行；需要验证数据库方言与并发的 cases 再在 `postgres_dev_test` 上运行。失败码、状态、版本、顺序和公开投影必须一致，数据库实现细节不进入契约。

### PostgreSQL 门禁

真实 PostgreSQL 继续验证：

- up / repeat / rollback / reapply migration；
- migration / runtime 角色分离和 DDL 拒绝；
- PostgreSQL 类型、索引、tuple pagination 和 advisory lock；
- 多连接池重启恢复与高并发 CAS；
- marker / checksum mismatch、连接关闭和 no fallback。

SQLite 通过不构成上述门禁通过。

### 浏览器连续验收

本地浏览器默认使用 `sqlite_dev`，覆盖应用创建、配置保存、发布候选、API 密钥签发与 Gateway 调用、请求历史、工作流草案与运行记录、平台 / Web 重启恢复。批次晋级前再以 `postgres_dev_test` 复验数据库专属和关键纵向路径。

## 实施顺序

### 批次 S1：共享 runtime 与迁移骨架

- 选择并评审 Go SQLite driver、许可证、纯 Go / CGO、macOS / Linux / Windows 和 CI 兼容性；
- 建立共享连接生命周期、文件权限、WAL、busy timeout、marker、checksum 和显式失败语义；
- 建立聚合配置、组件一致性检查和临时文件测试，不开放不完整的正式本地启动档。

状态：已于 2026-07-14 完成。driver、共享 runtime、migration marker、重复迁移、重开、损坏文件、marker / checksum mismatch、失败回滚、路径错误、文件权限和聚合配置冲突均已有真实临时文件或配置测试；正式 `sqlite_dev` 启动继续关闭。

### 批次 S2：七组件 repository

- 按应用目录、配置草案、发布候选、API 密钥、请求历史、工作流草案、运行记录的依赖顺序接入 SQLite；
- 复用领域与 repository contract cases，逐组件完成重启恢复、CAS、分页和敏感信息验证；
- 所有组件就绪后再打开聚合 `sqlite_dev` 本地启动档，避免长期混合存储。

进度：七组 repository 与聚合 runtime 已于 2026-07-14 全部完成。各组件均使用独立 migration、共享 runtime 注入和组件 marker 复验；聚合入口统一投影七个 store mode，按固定顺序应用 migration，并由 `Server` 独占 shared runtime 的启动、失败回滚和关闭生命周期。跨模式契约覆盖作用域、顺序、CAS / 终态、筛选游标、关闭失败、内容禁入和重启恢复；聚合真实文件测试进一步证明七类记录可在平台重启后恢复，marker 不兼容不会回退内存。跨平台本地启动档、SQLite HTTP 连续产品链和真实 PostgreSQL 专项门禁随后完成；Web 与浏览器仍未完成。

### 批次 S3：双数据库与浏览器收口

- 使用 SQLite 完成本地连续产品链和重启恢复；
- 使用 PostgreSQL 完成 migration、角色、并发和部署同构门禁；
- 更新 API 密钥专题、当前焦点和周志，再进入 API 密钥 Web 批次。

进度：默认 `local-product` / 显式 `configured` 跨平台启动档和 SQLite HTTP 连续产品链已于 2026-07-14 完成；配置校验不创建数据库，正式 `serve` 才进入 migration 与 shared runtime 生命周期。同日真实 PostgreSQL 已完成 migration、角色、类型 / 索引、advisory lock、多连接并发、竞态、重启恢复和 no-fallback 门禁。双数据库证据现已成立，下一项固定为 API 密钥 Web 一次性交接与浏览器连续验收。

## 停止线

- 不以 SQLite 替代 PostgreSQL 集成、CI 或部署同构验证。
- 不把 `sqlite_dev` 作为 production repository 或正式审计 / 凭据后端。
- 不持久化外部身份真相、provider credential、原始 API 密钥、完整模型输入输出或业务写回。
- 不自动删除、降级或重建不兼容数据库；不在存储失败时回退内存。
- 不为了跨数据库复用引入通用键值表、动态 ORM、晦涩 SQL builder 或大一统 repository。
- 不在七组件 shared runtime、selector、启动 / 关闭生命周期和连续链路完整接入并验证前，把聚合 `sqlite_dev` 宣称为正式本地产品启动档。
