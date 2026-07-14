# 用户工作区 API 密钥生命周期与 Gateway 开发测试态认证 v1

更新时间：2026-07-14

状态：`api_key_lifecycle_gateway_dev_test_auth_v1_sqlite_local_product_chain_completed`

## 当前结论

本专题把现有只读 API 密钥摘要升级为显式启用、可持久化、可吊销且能真实调用 Gateway 的开发测试态凭据闭环。目标用户是已经在用户工作区创建应用、完成配置并需要从代码或调试台验证模型 API 的内部开发者。

完整路径必须同时成立：用户为自己的活跃应用签发一枚有明确调用作用域和有效期的密钥；原始令牌只在创建成功后展示一次；Gateway 通过令牌恢复租户、工作区、应用、所有者和调用作用域；成功调用进入既有脱敏请求历史；吊销、过期、作用域不足、应用归档和存储故障都在 bridge / provider 之前失败关闭。

本功能只服务内部开发者预览。它不声明生产 API 密钥、正式工作区成员授权、配额、限流、计费、生产凭据后端或公开生产 Gateway 已就绪。

## 当前实现

批次 A 已完成密钥领域、密码学安全生成、一次性凭据交接、内存存储库和管理 API。显式生命周期模式可为当前所有者的活跃应用签发有期限且有受控作用域的密钥，支持脱敏列表、详情、筛选分页和预期版本吊销；创建成功响应是唯一返回原始令牌的位置，并设置 `Cache-Control: no-store`。

管理权限已经拆为 `api_keys:read`、`api_keys:write` 和 `api_keys:revoke`。应用缺失、已归档或跨所有者统一失败关闭；OIDC 成员关系未成立时保持存储库零查询；并发吊销只有一个版本能够成功。专项竞态测试和完整平台 Go 回归已经通过。

批次 B 的代码实现已经完成：新增互斥的 `dev_headers` / `api_key_dev_test` Gateway 认证模式，为五条 northbound 路由建立 Bearer 解析、路由作用域、活跃应用检查和可信 `GatewayRequestContext`；成功认证原子更新 `last_used_at` 并写入既有脱敏请求历史，凭据冲突、无效、吊销、过期、作用域不足、应用不可用以及任一存储故障都在 bridge / provider 前失败关闭。管理 API 明确拒绝开发测试态 API 密钥凭据，control-plane 认证中间件不再消费 northbound Authorization。

独立 PostgreSQL `api_key_records` / `api_key_schema_versions`、迁移清单与校验和、手动迁移命令、存储选择器、摘要隔离、稳定分页、吊销 CAS 和最近使用更新也已实现。配置 / 迁移单测、Gateway 负向与防回退测试、专项竞态以及完整平台 Go 回归已经通过；带 PostgreSQL 标签的测试已完成编译，但真实迁移、运行角色、重启恢复和并发认证 / 吊销尚未执行。

本专题现已消费[本地 SQLite 开发持久化 v1](../../platform/local-sqlite-dev-persistence-v1.md)：应用目录、API 密钥和请求历史所在的完整七组件本地数据链已统一接入 `sqlite_dev`，形成 `memory_dev / sqlite_dev / postgres_dev_test` 三层存储。现有 PostgreSQL 实现继续保留，不被 SQLite 替代；SQLite 连续链已经通过，待 PostgreSQL 专属门禁通过后关闭后端持久化批次并进入 Web。

2026-07-14 已完成 API 密钥 SQLite repository。该组件只接入共享 runtime、独立 component migration、领域 repository 与组件 selector；selector 同时复验应用目录与 API 密钥 migration。SQLite 的创建 / 过期 / 最近使用 / 吊销时间使用整数纳秒列参与比较和排序，公开记录继续保留 RFC3339Nano，避免可变小数位文本产生时间字典序偏差。memory / SQLite 复用签发、筛选分页、所有者隔离、吊销 CAS 和 Gateway 认证契约；真实文件测试覆盖最近使用单调更新、认证 / 吊销竞态、两次重启恢复、关闭失败以及原始令牌不进入数据库、WAL 或共享内存。Gateway 请求历史、工作流草案和工作流运行 SQLite repository 随后均已完成，聚合 `sqlite_dev` 已把七组件注入同一 shared runtime。

同日完成跨平台本地产品档与 SQLite 连续产品链。默认 `local-product` wrapper 统一选择七组件 SQLite 和开发门禁，显式 `configured` 档保留 PostgreSQL / 故障注入入口。真实 HTTP 测试以同一服务端生成应用为作用域，完成配置保存、发布候选审查、API 密钥签发、Bearer Gateway 调用、最近使用更新和脱敏请求历史，并把同应用工作流草案与运行一起写入共享数据库；平台重启后应用、草案、候选、密钥、历史、工作流和运行均可恢复，后续读取不返回原始令牌，数据库物理文件不包含令牌或原始输入。下一批只剩 PostgreSQL 专属门禁，之后才进入 Web。

当前尚未实现 Web 一次性交接和浏览器连续验收。开发请求头仍是默认认证模式；只有显式设置 `RADISHMIND_GATEWAY_AUTH_MODE=api_key_dev_test` 才由 API 密钥保护 Gateway，不能从开发测试态实现反推生产凭据能力成立。

## 产品缺口与用户流程

当前 `workspace-api-keys` 只展示预置摘要，不能签发、使用或吊销真实密钥；Gateway 调试路径依赖显式开发请求头，外部代码无法以应用凭据完成可审查调用。应用目录完成后，这已经成为“应用创建 → 配置 → API 调用”链上的首个独立缺口。

目标流程为：

1. 用户从活跃应用进入 API 密钥页，查看自己在当前工作区和应用下的脱敏密钥记录。
2. 用户选择允许的 Gateway 调用作用域、有效期和展示名称，服务端生成密钥标识与高熵令牌。
3. 创建响应通过独立的一次性凭据交接返回原始令牌；列表、详情、日志和后续读取永不返回令牌或摘要。
4. 用户复制令牌，或在当前页面内存仍持有令牌时交给既有 Gateway 调试台完成一次验证调用。
5. Gateway 从令牌恢复可信调用上下文，校验有效期、吊销状态、作用域和绑定应用活跃状态，再进入模型目录或推理处理器。
6. 成功认证的调用继续进入既有脱敏请求历史，`consumer_ref` 稳定投影为 `api_key:<api_key_id>`，不保存原始令牌。
7. 用户以预期版本吊销密钥；吊销后新的调用在 bridge / provider 和请求历史写入之前被拒绝，记录仍可只读审查。
8. PostgreSQL 开发测试态模式下，平台重启后密钥元数据、吊销状态、最近使用时间和验证结果保持一致。

## 启用条件与真相源

- API 密钥生命周期默认关闭；只有显式启用后才接管 `GET /v1/user-workspace/api-keys` 并开放写入端点。
- 显式密钥模式要求应用目录同时启用。签发与 Gateway 验证都必须精确读取同一 `ApplicationCatalogRepository` 中的活跃应用，不允许绑定预置假应用。
- `APIKeyRepository` 是密钥标识、应用绑定、所有者、作用域、有效期、吊销状态、记录版本、凭据摘要和最近使用时间的唯一真相源。
- `memory_dev`、`sqlite_dev` 与 `postgres_dev_test` 是互斥存储模式；正式本地启动档统一选择 `sqlite_dev`，批次 / CI 数据库门禁选择 `postgres_dev_test`。连接、文件、schema 标记、查询、凭据验证或写入失败必须显式失败，不得回退内存、预置摘要或开发请求头。
- 未启用时，现有只读假数据路由和离线页面保持原状；同一运行实例不得把假数据摘要和可验证密钥合并成一个可写列表。

历史 `gateway-api-key-quota-readiness` 只作为字段、失败语义和脱敏边界输入，不再派生同层准入链。本专题以功能设计、单一实施任务卡、单元 / 集成测试和现有聚合门禁承载实现证据。

## 密钥领域模型

持久化记录固定为 `api_key_record.v1`，包含：

- `schema_version`；
- `api_key_id`：服务端生成的稳定短标识；
- `tenant_ref`、`workspace_id`、`application_id`、`owner_subject_ref`；
- `display_name`；
- `scopes`：排序去重后的受控调用作用域；
- `lifecycle_state`：持久化状态只允许 `active` 或 `revoked`；
- `effective_state`：响应投影为 `active`、`expired` 或 `revoked`，过期由服务端时间计算，不伪造状态写入；
- `record_version`：从 `1` 开始单调递增；
- `credential_digest`：仅存储库和认证服务内部可见的固定长度摘要，不进入公开领域投影；
- `created_at`、`expires_at`、可空 `last_used_at`、可空 `revoked_at`；
- `created_by_actor_ref`、可空 `revoked_by_actor_ref`；
- `request_id`、`audit_ref`。

展示名长度为 2 到 80 个字符。创建请求不直接接收绝对过期时间，只接收 1 到 90 天的 `expires_in_days`，由服务端根据同一时钟计算 `expires_at`。v1 不签发永不过期密钥。

允许的 Gateway 作用域固定为：

| 作用域 | 允许路由 |
| --- | --- |
| `models:read` | `GET /v1/models`、`GET /v1/models/{id}` |
| `chat:invoke` | `POST /v1/chat/completions` |
| `responses:invoke` | `POST /v1/responses` |
| `messages:invoke` | `POST /v1/messages` |

创建时至少选择一个作用域。`usage:read`、`runs:read`、管理 API、密钥管理 API、配额和计费能力不包含在本批密钥授权中。

## 凭据生成、存储与一次性交接

- 令牌格式固定为版本化前缀、公开密钥标识和随机秘密三段，例如 `rmd_dev_<api_key_id>.<secret>`；解析器只接受当前版本和固定字符集。
- `secret` 使用系统密码学安全随机源生成至少 256 位随机值，并以无填充 URL 安全编码承载；不得使用时间、UUID、伪随机数或用户输入派生。
- 开发测试态存储只保存完整令牌的 `SHA-256` 摘要。高熵随机输入使直接摘要足以承担本阶段离线猜测边界；比较必须使用常量时间函数。生产环境是否需要服务端 pepper、专用凭据后端或密钥轮换策略必须重新设计，不能从本阶段直接晋级。
- 创建响应使用独立 `api_key_issue_response.v1`：包含脱敏记录和一次性 `credential.token`。该字段只存在于成功创建响应，响应必须设置 `Cache-Control: no-store`；列表、详情、错误、审计、请求历史和数据库查询结果不得包含它。
- Web 只在当前组件内存持有一次性令牌；刷新、路由离开、明确关闭或完成交接后立即清除。不得写入 URL、localStorage、sessionStorage、cookie、日志、遥测、错误状态或 committed fixture。
- 常规 API 密钥列表和详情继续遵守 forbidden output guard。一次性创建响应使用独立严格消费器，不复用允许任意敏感字段的通用响应类型。
- v1 不提供“再次查看”、下载、导出或恢复原始令牌。需要替换时先签发新密钥，验证后再吊销旧密钥。

## 生命周期与并发语义

### 签发

- 创建请求只接收 `workspace_id`、`application_id`、`display_name`、`scopes` 和 `expires_in_days`。
- 租户、所有者、参与者、密钥标识、状态、版本、时间和审计字段由服务端生成。
- 创建前必须按租户、工作区、应用和所有者执行精确 `RequireActive`；应用不存在、已归档或不属于当前所有者时不生成随机令牌、不写存储库。
- 同一应用允许多枚活跃密钥，展示名不承担唯一性；用户以密钥标识、作用域和创建时间区分记录。

### 验证与最近使用时间

- Gateway 先严格解析 Bearer 令牌并定位密钥记录，再以常量时间比较摘要；不存在与摘要不匹配统一返回 `api_key_invalid`，不泄漏标识是否存在。
- 通过摘要后检查持久化状态、服务端过期时间、路由作用域和绑定应用活跃状态。任一步失败都不得调用 bridge / provider，也不得创建 Gateway 请求历史。
- 认证服务以“成功校验并取得调用准入”为线性化点。应用归档与认证并发时，归档在线性化点前完成则拒绝；认证先完成则该次请求可继续，归档后的新请求必须拒绝。
- `last_used_at` 只在完整认证成功后更新；更新失败必须在进入 bridge / provider 前返回 `api_key_store_unavailable`，不能以过期的最近使用记录继续调用。
- 成功认证生成 `GatewayRequestContext`，其租户、工作区、应用、所有者和作用域全部来自密钥记录；请求头不能覆盖。`Source` 为 `api_key_dev_test`，`ConsumerRef` 为 `api_key:<api_key_id>`。

### 吊销

- 吊销使用 `expected_version` 原子 CAS，成功后写入 `revoked`、`revoked_at`、操作人和新版本。
- 过期但尚未吊销的记录仍可吊销，以形成明确审计终态；已经吊销的记录再次吊销返回稳定状态冲突，不静默成功。
- v1 不提供恢复、反吊销、物理删除、原地换密、修改作用域或延长有效期。

## 身份、作用域与 Gateway 认证模式

密钥管理 API 继续使用现有已验证用户身份上下文，而不是用待管理的 API 密钥授权：

- 列表和详情要求 `api_keys:read`；
- 签发要求 `api_keys:write`；
- 吊销要求独立 `api_keys:revoke`；
- 所有操作采用所有者作用域，不实现工作区共享、代管或所有权转移；
- `radish_oidc_integration_test` 模式下成员关系契约未成立时，在应用与密钥存储库查询前返回 `workspace_membership_unavailable`。

Gateway 新增显式认证模式选择：

- `dev_headers` 保留现有内部调试路径；
- `api_key_dev_test` 要求所有受支持 northbound 路由携带一枚有效 Bearer 密钥；
- 两种模式互斥，配置未知必须启动失败；`api_key_dev_test` 不得在密钥无效时回退开发请求头或匿名调用；
- 请求同时携带 Bearer 密钥与 `X-RadishMind-Dev-Gateway-*` 身份头时返回凭据冲突；
- `/healthz` 和平台运维路由不纳入 API 密钥认证，密钥管理路由也不接受 API 密钥作为用户身份。

## API 边界

密钥生命周期复用现有集合路由，并增加详情与吊销：

- `GET /v1/user-workspace/api-keys`
- `POST /v1/user-workspace/api-keys`
- `GET /v1/user-workspace/api-keys/{api_key_id}`
- `POST /v1/user-workspace/api-keys/{api_key_id}/revoke`

列表和详情要求显式 `workspace_id`，列表允许受控 `application_id`、`effective_state`、`limit` 和 `cursor`。稳定顺序为 `created_at DESC, api_key_id DESC`，游标绑定租户、工作区、所有者和筛选摘要。

生命周期模式下，现有 `APIKeySummary` 继续作为列表兼容投影；允许增加 `workspace_id`、`application_id`、`display_name`、`effective_state`、`record_version` 和 `revoked_at` 等受控字段。消费端必须严格校验扩展字段，不允许摘要、令牌、请求头或内部认证材料进入投影。

## 稳定失败语义

至少固定以下失败码：

- `api_key_lifecycle_disabled`；
- `api_key_application_catalog_required`；
- `api_key_scope_denied`；
- `api_key_payload_invalid`；
- `api_key_secret_material_forbidden`；
- `api_key_not_found`；
- `api_key_version_conflict`；
- `api_key_revoked`；
- `api_key_expired`；
- `api_key_missing`；
- `api_key_invalid`；
- `api_key_credential_conflict`；
- `api_key_store_unavailable`；
- `workspace_membership_unavailable`。

`api_key_missing` 与 `api_key_invalid` 返回 `401`；已可信识别但被吊销、过期、作用域不足或应用不可用返回 `403`；管理端版本冲突返回 `409`。公开诊断只包含稳定错误码、请求标识和脱敏审计引用，不包含令牌片段、摘要、Authorization 内容、SQL、DSN、调用栈或其他所有者数据。

## 存储库与 PostgreSQL 开发测试态边界

`APIKeyRepository` 至少提供：

- `Create`；
- `Read`；
- `List`；
- `FindCredential`，只向认证服务返回内部验证材料；
- `RecordSuccessfulAuthentication`；
- `Revoke`。

PostgreSQL 使用独立 `api_key_records` 和 `api_key_schema_versions`，不复用控制面只读表、应用目录表、配置草案表或请求历史表。主键固定为 `(tenant_ref, workspace_id, api_key_id)`；凭据摘要建立唯一约束，应用、所有者、状态、有效期和创建时间使用必要索引。

迁移继续使用独立 manifest、checksum、advisory lock 和手动运行器；平台启动不自动迁移。迁移角色负责 DDL，运行角色只具必要 DML。连接、标记、checksum、查询、认证更新或 CAS 失败不得回退到 `memory_dev`、预置摘要或开发请求头。

## Web 与现有产品链交接

- API 密钥页从只读摘要升级为应用作用域列表、签发表单、一次性令牌交接、详情和吊销确认；默认离线模式继续不发网络请求。
- 一次性交接明确说明令牌无法再次查看，提供单独复制动作；不拼接或展示完整 `Authorization` 命令行，避免令牌进入终端历史模板。
- 在令牌仍存在于组件内存时，用户可把它交给既有 Gateway 调试台；调试台不得持久化令牌，应用切换、认证模式切换和页面刷新都必须清除。
- API key 模式下，模型目录、单次 / 流式调用与取消继续复用现有 Gateway 请求和响应契约；不新增第二套 northbound schema。
- 成功调用的请求历史显示 `api_key_id` 派生的 consumer reference、应用作用域和既有脱敏运行证据，不显示令牌、摘要或认证请求头。
- 应用归档后密钥记录和既有请求历史仍可读取并可吊销，但不能再签发密钥，现有密钥也不能发起新调用。

## 分批实施顺序

### 批次 A：领域、内存存储与管理 API

状态：已完成。

- 实现记录、密码学安全生成、一次性交接、作用域 / 有效期校验、所有者作用域和敏感信息保护；
- 实现内存存储库、列表 / 详情 / 创建 / 吊销、CAS 与应用活跃检查；
- 用单元和 HTTP 负向测试固定一次性返回、零泄漏、权限分离、成员关系失败关闭和无回退。

### 批次 B：Gateway 认证与 PostgreSQL 实现

状态：代码与 SQLite 本地产品链已实现，等待 PostgreSQL 专属验证。

- 实现显式 Gateway 认证模式、路由作用域映射、可信 `GatewayRequestContext` 和认证失败零 bridge / provider / history 副作用；
- 实现独立 PostgreSQL schema、迁移、运行器、存储选择器、原子最近使用更新、吊销 CAS 和重启恢复；
- 完成迁移、运行角色、摘要不可逆读取、并发认证 / 吊销、应用归档竞态和 no-fallback 集成验证。

### 批次 B2：统一 SQLite 本地持久化

状态：已完成；七组件共享 runtime、跨平台启动档和 SQLite 连续产品链均已通过。

- 消费平台级 `sqlite_dev` 共享 runtime、schema migration 和聚合本地启动档；
- API 密钥、应用目录和请求历史与其余四组本地运行数据整体接入，避免正式本地入口长期混用存储；
- 验证签发、重启恢复、Gateway 认证、最近使用更新、请求历史、吊销和敏感信息边界，全程不需要 Docker。

### 批次 B3：双数据库门禁

状态：进行中；SQLite 职责证据已通过，下一步执行 PostgreSQL 专属门禁。

- SQLite 验证本地连续产品链、重启恢复、分页、CAS 和 no-fallback；
- PostgreSQL 验证 migration、运行 / 迁移角色、类型 / 索引语义、多连接并发和部署同构；
- 两种数据库均通过后关闭后端持久化批次，再进入 Web。

### 批次 C：Web 与连续验收

- 实现严格 Web 消费端、签发 / 一次性交接 / 调试台内存交接 / 吊销界面；
- 覆盖应用创建与配置、密钥签发、模型目录、单次与流式调用、请求历史、吊销、拒绝和重启恢复；
- 收口功能文档、任务卡、当前焦点和周志，不派生同层 readiness / checker 链。

## 验收方式

- Go 单元测试：令牌格式与熵源、摘要与常量时间比较、字段校验、有效期、作用域、所有者隔离、CAS、吊销、过期、应用归档和敏感信息扫描。
- HTTP / Gateway 测试：管理权限分离、未知字段、一次性响应、凭据冲突、路由作用域、可信上下文、无回退，以及失败时零 bridge / provider / 请求历史调用。
- PostgreSQL：迁移应用 / 重复应用 / 回滚 / 重应用、运行角色 DDL 拒绝、重启恢复、并发认证 / 吊销、最近使用更新和摘要不出库。
- Web：默认离线零请求、严格响应校验、一次性令牌只存在内存、刷新清除、应用切换清除、吊销确认和失败状态。
- 浏览器：空目录创建应用、保存配置、签发密钥、以密钥完成 Gateway 调用、查看同请求历史、刷新确认令牌不可恢复、吊销、再次调用失败、重启后状态保持。
- 门禁：平台 Go 回归与竞态、Web 单测 / 构建、Gateway smoke、快速仓库检查；本专题涉及认证、schema 和阶段真相源，完成实现批次时补跑完整仓库检查。

## 停止线

- 不声明 production API key、公开生产 Gateway、正式工作区成员授权或生产凭据存储已就绪。
- 不并行实现 quota enforcement、rate limit、billing、cost ledger、自动 retry/fallback、load balancing 或生产 secret backend。
- 不把用户 API 密钥与 provider credential 混用；密钥记录不得包含 provider profile secret、模型服务令牌、DSN 或外部凭据。
- 不允许密钥调用管理端、密钥生命周期、应用生命周期、配置草案或发布审查 API。
- 不提供原始令牌再次查看、恢复、导出、物理删除、反吊销、原地换密、作用域修改或有效期延长。
- 不为本功能逐项新增 readiness、refresh、fixture 或 checker 链；新增 schema 和高风险认证边界由单一实施任务卡、代码测试、PostgreSQL 集成和现有聚合门禁承载。
