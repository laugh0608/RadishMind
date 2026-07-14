# 本地 SQLite 开发持久化 v1 实施任务卡

更新时间：2026-07-14

状态：`local_sqlite_dev_persistence_v1_s2_four_repositories_completed`

## 任务目标

按照[本地 SQLite 开发持久化 v1](../platform/local-sqlite-dev-persistence-v1.md)，为 RadishMind 建立无需外部数据库服务、可重启恢复的正式本地开发持久化档，同时保留 `memory_dev` 快速测试和 `postgres_dev_test` 数据库同构门禁。

本任务必须完整覆盖七组本地运行数据和聚合启动体验。只给单个 repository 增加 SQLite、只创建数据库文件或只新增 selector 均不满足完成条件。

## 实现范围

1. 评审并引入单一 Go SQLite driver，确认许可证、纯 Go / CGO、macOS / Linux / Windows、Go 版本和 CI 兼容性。
2. 建立共享 SQLite runtime，统一数据库路径、目录 / 文件权限、foreign keys、WAL、busy timeout、连接生命周期、checkpoint 和关闭回收。
3. 建立 component migration 编排、marker、migration id、store schema version、checksum、事务和不兼容失败关闭；禁止反射建表和启动时静默重建。
4. 为工作流草案、应用配置草案、发布候选、应用目录、API 密钥、工作流运行和 Gateway 请求历史实现 SQLite repository。
5. 建立聚合 `sqlite_dev` 配置和本地启动入口，一次选择七组件；聚合模式与组件配置冲突时启动失败。
6. 把 `var/sqlite-dev/`、数据库、WAL、共享内存和临时文件纳入 Git 忽略与敏感信息边界。
7. 复用领域与 repository contract cases，覆盖重启恢复、CAS、筛选分页、终态、时间、no-fallback 和公开投影一致性。
8. 以 SQLite 完成本地浏览器连续链路，以 PostgreSQL 完成 migration / 角色 / 类型 / 索引 / 并发门禁，再恢复 API 密钥 Web 批次。

## 明确排除

- Control Plane Tenant / Audit 外部只读投影；
- Radish 身份、成员关系、业务真相和正式 OIDC 数据；
- production repository、生产审计、生产凭据、配额和计费；
- provider credential、原始 API 密钥、模型权重、数据集和完整模型输入输出；
- 用 SQLite 测试结果替代 PostgreSQL 集成或 CI 门禁。

## 分批实施

### S1：共享 runtime 与 migration

- 完成 driver 评审和依赖引入；
- 建立共享 runtime、文件治理、连接设置、component marker / checksum 和临时文件测试；
- 加入聚合配置校验，但保持正式 `sqlite_dev` 启用失败，直到七组件齐备。

完成标志：临时 SQLite 文件可可靠创建、迁移、重复检查、重开和关闭；损坏、marker mismatch、配置冲突与不可写路径均稳定失败，不产生 fallback。

实现记录：2026-07-14 已锁定 `modernc.org/sqlite v1.53.0`，完成共享 runtime、文件 / WAL / 共享内存权限、foreign keys、WAL、busy timeout、完整性检查、checkpoint、事务 migration marker、checksum、不兼容失败关闭和聚合配置。专项测试覆盖首次创建、两步 migration、重复执行、重开恢复、marker / checksum mismatch、未知 marker、migration 回滚、损坏文件、非法路径、配置冲突与启动前失败；七组件未齐备前继续以 `sqlite_dev_repository_set` 阻止正式启动。CGO 关闭条件下的 macOS / Linux / Windows 编译、专项竞态、平台完整 Go 回归、`go vet`、模块校验和仓库快速门禁均已通过。

### S2：七组件 repository

- 依次接入应用目录、配置草案、发布候选、API 密钥、Gateway 请求历史、工作流草案和工作流运行；
- 每完成一个组件即运行同领域契约、重启恢复、CAS / 终态、分页和敏感信息测试；
- 不在中途把正式本地入口暴露成混合 store mode。

实现记录：2026-07-14 已完成首个应用目录纵向切片。新增应用目录 SQLite migration 与 repository，复用 memory / SQLite 同组领域契约，覆盖所有者隔离、稳定分页、预期版本 CAS、不可逆归档、重启恢复和 no-fallback；组件 selector 必须接收共享 runtime 并复验本组件 marker，不能自行打开数据库。聚合 `sqlite_dev` 启动检查继续关闭。

实现记录：2026-07-14 已共同完成应用配置草案与发布候选。两组 repository 共享 runtime 但保持独立 schema；发布 selector 同时验证草案和候选 marker。复用契约覆盖草案创建 / 更新、候选不可变创建、审查追加 / 终态、作用域、稳定顺序和漂移语义；真实临时文件进一步覆盖两组 8 路并发单写者、跨组件重启恢复、关闭失败和敏感材料禁入。S2 当前完成 3/7，下一项为 API 密钥；正式晋级、Web、PostgreSQL 专属语义和聚合启动门禁均未修改。

实施记录：2026-07-14 已完成 API 密钥 SQLite 组件。该组件复用共享 runtime 并依赖应用目录 migration，采用整数纳秒承担过期、最近使用与分页时间谓词；验证覆盖签发、作用域 / 所有者隔离、稳定分页、最近使用单调更新、吊销 CAS、Gateway northbound 认证、认证 / 吊销并发、两次重启恢复、关闭失败和原始令牌文件扫描。S2 当前完成 4/7，下一项为 Gateway 请求历史；聚合启动、Web 和 PostgreSQL 专属门禁均未修改。

完成标志：七组件均由同一 SQLite runtime 承载，领域状态、版本、顺序、失败码和公开投影与既有模式一致。

### S3：聚合启动与双数据库验收

- 打开聚合 `sqlite_dev` 本地启动档和跨平台 wrapper；
- SQLite 浏览器覆盖应用、配置、候选、密钥、Gateway、请求历史、工作流草案 / 运行与服务重启；
- PostgreSQL 继续覆盖手动 migration、运行 / 迁移角色、advisory lock、数据库类型 / 索引、多连接并发和 no-fallback；
- 同步 API 密钥专题、当前焦点、周志和运行说明。

完成标志：日常产品联调不需要 Docker；批次级 PostgreSQL 验证仍可独立执行并保持通过；任何文档都不把 SQLite 写成生产数据库或 PostgreSQL 替代品。

## 高风险边界

- 数据库路径必须在受控本地运行目录，配置摘要不泄漏不必要的用户绝对路径。
- API 密钥原始令牌不得进入 SQLite 主文件、WAL、共享内存、备份、日志或测试输出。
- CAS 必须由 SQLite transaction、唯一约束和版本 predicate 保证，不得用进程互斥锁伪装数据库原子性。
- migration 只能前滚已评审 schema；marker 或 checksum 不匹配时停止启动，不自动删除或重建数据库。
- repository 失败必须保留稳定领域失败码，不暴露 SQL、文件内容、调用栈或凭据材料。
- shared runtime 只共享连接、migration 编排和错误归一化，不创建大一统 repository 或通用键值存储。

## 必要验证

- SQLite runtime / migration 单元测试与真实临时文件测试；
- 七组件 repository contract、重启恢复、分页、CAS、终态和 no-fallback；
- API 密钥认证 / 吊销竞态及数据库 / WAL 敏感信息扫描；
- `go test ./...`、相关 `-race` 测试、`go vet ./...`；
- 本地 SQLite Web 测试、构建与真实浏览器连续链；
- 既有 PostgreSQL 破坏性集成；
- `./scripts/check-repo.sh --fast`，专题完成时运行完整 `./scripts/check-repo.sh`。

## 完成条件

- `memory_dev / sqlite_dev / postgres_dev_test` 三层职责、配置和运行行为稳定且无隐式回退。
- 正式本地启动档一次选择七组件 SQLite，并能在平台重启后恢复完整本地开发数据。
- SQLite 与 PostgreSQL 分别通过自身职责内的验证，领域契约保持一致。
- 本地数据库文件和敏感材料不进入 Git、日志、公开响应或长期文档。
- Control Plane 外部真相、生产 repository、生产凭据、审计、配额和计费继续保持停止线。
