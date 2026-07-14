# 用户工作区 API 密钥生命周期与 Gateway 开发测试态认证 v1 实施任务卡

更新时间：2026-07-14

状态：`api_key_lifecycle_gateway_dev_test_auth_v1_sqlite_local_product_chain_completed`

## 任务目标

按照[用户工作区 API 密钥生命周期与 Gateway 开发测试态认证 v1](../features/user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md)，把现有只读 API 密钥摘要升级为可签发、一次性交接、持久验证、使用和吊销的开发测试态凭据，并让它真实保护现有 Gateway northbound 路由。

本任务交付的是一条完整纵向链：活跃应用 → 签发 → 一次性令牌 → Gateway 作用域认证 → 脱敏请求历史 → 吊销 / 过期拒绝。仅完成可写列表或只做认证中间件都不满足完成条件。

## 前置条件

- 应用目录与生命周期已完成 `memory_dev`、`postgres_dev_test`、所有者作用域、精确 `RequireActive`、CAS、归档和重启恢复。
- Gateway 已有 `/v1/models`、Chat Completions、Responses、Messages、受控 bridge、调试台和脱敏请求历史。
- 现有 `APIKeySummary`、`api_keys:read` 和历史 `gateway-api-key-quota-readiness` 可作为兼容投影、字段与失败语义输入，但不作为实现完成证明。
- 生产认证、正式成员关系、生产凭据后端、配额、计费和公开生产 Gateway 继续关闭。

## 实现范围

1. 建立 `api_key_record.v1`、服务端密钥标识、256 位安全随机秘密、版本化令牌、内部摘要、一次性创建响应和严格敏感信息保护。
2. 建立 `APIKeyRepository`，实现 `memory_dev`、统一平台 `sqlite_dev` 与独立 `postgres_dev_test`，统一所有者 / 应用作用域、筛选分页、有效期、最近使用更新、吊销 CAS 和无回退语义。
3. 显式生命周期模式接管现有 API 密钥列表，并新增创建、详情和吊销端点；默认未启用模式继续保持历史只读摘要。
4. 新增 `api_keys:write` 与 `api_keys:revoke`，签发和验证精确检查活跃应用；OIDC 成员关系未成立时保持存储库零查询失败关闭。
5. 建立互斥的 `dev_headers` / `api_key_dev_test` Gateway 认证模式，为五条 northbound 路由执行令牌、状态、有效期、作用域和应用活跃校验。
6. 认证成功后从密钥记录生成可信 `GatewayRequestContext` 并复用请求历史；认证失败不得进入 bridge、provider 或请求历史写入。
7. 更新 Web API 密钥页和调试台，完成签发、一次性内存交接、调用、刷新清除、吊销和失败审查。
8. 完成 SQLite 本地连续链路、PostgreSQL 专属门禁、真实浏览器、重启恢复、并发与敏感信息验收，并同步文档真相源。

## 分批实施顺序

### 批次 A：领域、内存存储与管理 API

状态：已完成。

- 密钥记录、密码学安全生成、摘要比较、有效期与作用域校验；
- 内存存储库、列表 / 详情 / 创建 / 吊销和原子 CAS；
- 应用活跃检查、管理权限分离、成员关系失败关闭、一次性响应与负向 HTTP 测试。

完成标志：不依赖数据库即可通过 API 为活跃应用签发并吊销密钥；除创建成功响应外，任何输出和日志都找不到原始令牌或摘要。

实现记录：已建立 `api_key_record.v1`、256 位安全随机秘密、版本化令牌与内部摘要；完成所有者作用域的内存存储库、稳定分页、有效期投影、管理 API 和吊销 CAS。应用不可用检查先于凭据生成，OIDC 成员关系缺失时保持零存储查询；一次性创建响应、权限分离、敏感字段拒绝、并发吊销和存储故障无回退均有专项测试。完整平台 Go 回归与 API 密钥专项竞态测试通过。

### 批次 B：Gateway 认证与 PostgreSQL

状态：代码与 SQLite 本地产品链已实现，等待 PostgreSQL 专属验证。

- 显式 Gateway 认证模式、Bearer 解析、路由作用域、可信上下文和凭据冲突拒绝；
- 独立 schema、迁移清单、校验和、手动运行器、运行 / 迁移角色和存储选择器；
- 原子最近使用更新、并发认证 / 吊销、应用归档竞态、重启恢复和 no-fallback 集成验证。

完成标志：PostgreSQL 模式下有效密钥可以调用受支持 Gateway 路由并进入同作用域请求历史；吊销、过期、作用域不足、应用归档和存储故障都在 bridge / provider 前拒绝。

实现记录：五条 northbound 路由已经接入互斥且失败关闭的 `api_key_dev_test` 认证，可信上下文、作用域、应用活跃状态、最近使用更新和脱敏请求历史均由服务端记录恢复；无效凭据、开发身份头冲突、吊销、过期、作用域不足、应用不可用以及密钥 / 历史存储故障均有执行前负向测试。独立 API 密钥 PostgreSQL schema、手动迁移命令、存储选择器、摘要隔离、吊销 CAS、重启恢复与并发集成测试已经落地。配置 / 迁移测试、专项竞态和完整平台 Go 回归通过；真实 PostgreSQL 集成尚未执行，现按平台三层存储设计先补统一 SQLite 本地持久化，再完成双数据库验证。

### 批次 B2：统一 SQLite 本地持久化

状态：已完成；七组件共享 runtime、跨平台启动档和 SQLite 连续产品链均已通过。

- 消费[本地 SQLite 开发持久化 v1](../platform/local-sqlite-dev-persistence-v1.md)的共享 runtime、component migration 和聚合启动档；
- API 密钥不得单独引入私有 SQLite 连接或临时 schema，必须与七组本地运行数据统一选择、统一生命周期和统一失败语义；
- 不依赖 Docker 验证应用、签发、认证、最近使用、请求历史、吊销与平台重启恢复。

实施记录：2026-07-14 已完成 API 密钥 SQLite 组件、七组件聚合 shared runtime、默认 `local-product` 跨平台启动档与同一应用作用域 HTTP 连续链。repository 不创建私有连接；selector 同时复验应用目录与 API 密钥 migration。整数纳秒承担过期、最近使用与分页时间谓词，领域投影继续使用 RFC3339Nano；memory / SQLite 同组契约覆盖稳定分页、过期筛选、所有者隔离、最近使用单调更新、吊销 CAS 与 Gateway northbound 认证。连续链进一步覆盖服务端应用创建、配置 / 发布、签发、Bearer 调用、请求历史、工作流草案 / 运行和平台重启恢复，并确认原始令牌与输入不进入数据库、WAL 或共享内存。本批不替代 PostgreSQL 专属门禁。

完成标志：正式本地启动档只需 SQLite 文件即可恢复完整开发数据链；API 密钥原始令牌不进入数据库、WAL、日志或后续响应，存储失败不回退内存。

### 批次 B3：双数据库验证

状态：进行中；SQLite 职责证据已通过，下一步执行 PostgreSQL 专属门禁。

- SQLite 承载日常本地连续路径与浏览器重启恢复；
- PostgreSQL 承载 migration / runtime 角色、advisory lock、类型 / 索引、并发和部署同构门禁；
- 同一 repository contract cases 固定两种数据库的领域状态、版本、分页、失败码和脱敏投影一致性。

完成标志：SQLite 和 PostgreSQL 的职责证据分别成立，且没有以 SQLite 结果替代 PostgreSQL 专属验证。

### 批次 C：Web 与连续验收

- 严格 API 消费端、应用筛选、签发表单、一次性令牌面板、调试台内存交接和吊销确认；
- 刷新 / 路由离开 / 应用切换清除令牌，列表与详情始终只读脱敏字段；
- 浏览器完成应用创建、配置、签发、模型目录、单次 / 流式 / 取消调用、请求历史、吊销、再次拒绝和重启恢复。

完成标志：用户不借助开发身份头即可用新签发密钥完成一次可审查 Gateway 调用，并能在产品内确认吊销立即生效；浏览器存储、日志、URL 和后续 API 响应均无原始令牌。

## 高风险边界

- 原始令牌只允许存在于创建成功响应和短暂 Web 内存；数据库、列表、详情、错误、日志、审计、请求历史和测试产物全部禁止。
- 令牌解析、摘要比较、过期、吊销、作用域和应用状态任何一项失败都不得回退开发请求头或匿名路径。
- 管理 API 由用户身份授权，API 密钥只能调用明确列出的 Gateway 路由，不能管理自身或其他资源。
- SQLite、PostgreSQL 与内存实现必须共享同一领域规则；数据库失败不得被伪装成无记录、无效密钥或内存成功。
- API 密钥是 RadishMind northbound consumer credential，不是 provider credential，也不解除生产 secret backend 停止线。

## 必要验证

- `go test ./...` 与 API 密钥 / Gateway 认证专项竞态测试；
- Web 单元测试、严格消费端测试与生产构建；
- SQLite 临时文件 repository、聚合本地启动、重启恢复、吊销、过期与无回退集成；
- PostgreSQL 迁移、角色、并发、吊销、过期、重启与无回退破坏性集成；
- Gateway consumer smoke，覆盖五条路由的允许 / 拒绝作用域和零 bridge / provider 副作用；
- 真实浏览器连续路径与敏感信息检查；
- `./scripts/check-repo.sh --fast`，专题完成时补跑完整 `./scripts/check-repo.sh`。

## 完成条件

- 开发者可为自己的活跃应用签发有期限、有调用作用域的密钥，原始令牌只展示一次。
- API 密钥模式下 Gateway 身份、应用和权限来自密钥记录，不能由请求头覆盖，也不回退其他认证来源。
- 成功调用进入既有脱敏请求历史；吊销、过期、作用域不足和应用归档在执行前失败关闭。
- `memory_dev`、`sqlite_dev` 与 `postgres_dev_test` 领域语义一致；SQLite 服务本地连续开发，PostgreSQL 服务数据库同构门禁，两者均可迁移、可重启恢复、可审计且不回退。
- 生产密钥、正式成员关系、配额、限流、计费、生产凭据后端和公开生产 Gateway 继续保持停止线。
