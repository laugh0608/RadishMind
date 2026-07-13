# 用户工作区应用目录与生命周期（开发/测试态）v1

更新时间：2026-07-13

状态：`application_catalog_lifecycle_dev_test_v1_defined`

## 当前结论

本专题建立用户可创建、查看、更新和归档应用的开发测试态生命周期，并把新应用连续交给现有配置草案、API 接入、Gateway 调试台、请求历史和发布审查。

当前实现中的 `GET /v1/user-workspace/applications` 仍从预置只读假数据存储库返回 `ApplicationSummary`。本专题不在该列表上叠加可写覆盖层，也不合并两套应用列表；显式启用应用目录模式后，应用身份、所有权、元数据和生命周期统一由独立 `ApplicationCatalogRepository` 提供。旧只读假数据模式只保留为未启用应用目录时的历史开发路径，不能作为目录存储失败时的回退来源。

本功能属于内部开发者预览。开发测试态目录可持久化、可审查和可恢复，不代表生产应用存储库、正式工作区成员授权、应用晋级、生产 API 密钥、配额或计费已成立。

## 产品缺口与目标用户

现有应用配置草案、API 接入和发布审查都要求先选择一个已经存在的应用，但应用列表仍依赖预置摘要，用户不能从产品界面建立自己的应用，也不能管理应用元数据或生命周期。

目标用户是需要在单一工作区内建立开发测试态应用，并继续完成配置、调用和审查的内部开发者。完整路径为：

1. 用户进入应用目录，查看当前所有者在当前工作区内的活跃应用。
2. 用户创建应用；应用标识由服务端生成，创建者成为不可变所有者。
3. 用户进入应用详情，查看版本、生命周期、审计引用和下游能力入口。
4. 用户通过完整元数据快照和预期版本更新应用，冲突时显式刷新或重新提交。
5. 用户选择活跃应用，进入现有配置草案、API 接入、Gateway 调试台和发布审查。
6. 用户通过预期版本归档应用；归档后仍可读取详情和既有历史证据，但不能继续修改应用或创建新的配置 / 发布审查动作。
7. PostgreSQL 开发测试态模式下，平台和 Web 重启后应用、版本与归档状态保持一致。

## 真相源与现有只读列表迁移

### 唯一应用真相源

- 应用目录模式启用后，`ApplicationCatalogRepository` 是应用身份、工作区归属、所有者、展示元数据、生命周期和记录版本的唯一真相源。
- 不把控制面假数据存储库与应用目录记录合并为一个列表，不以“目录没有记录”为理由读取预置应用。
- `memory_dev` 和 `postgres_dev_test` 是互斥存储模式；连接、schema 标记或查询失败必须显式失败，不得回退内存或假数据。
- PostgreSQL 迁移不写业务种子。真实浏览器和集成测试通过创建 API 建立应用；如开发演示需要预置应用，应使用显式、可审计的开发引导动作，不把种子写进迁移。

### `ApplicationSummary` 兼容投影

现有应用列表和下游页面继续消费 `ApplicationSummary` 兼容投影，但投影来源改为应用目录记录：

- `application_ref` 来自 `application_id`；
- `tenant_ref`、`application_kind`、`display_name`、`owner_subject_ref` 和 `updated_at` 来自目录记录；
- 新建应用尚无工作流或运行记录时，`latest_workflow_definition_ref` 为空，`last_run_status` 为 `not_available`；不得从其他预置应用复制展示数据；
- `workspace_id`、`description`、`lifecycle_state`、`record_version`、`created_at` 和可空 `archived_at` 作为目录模式下的受控扩展字段，由更新后的消费端严格校验；
- 发布治理读取应用基线时改为按完整作用域和精确应用标识读取目录记录，不再通过列出 200 条摘要后线性查找。

未启用应用目录模式时，原只读路由保持现有行为，保证历史只读证据和离线基线不被本设计提前破坏；同一运行实例不得同时把假数据与目录记录解释为可写应用真相源。

## 应用领域模型

应用记录 schema 固定为 `application_catalog_record.v1`，包含：

- `schema_version`；
- `application_id`：服务端生成的稳定短标识，建议使用 `app_` 前缀和固定长度小写安全随机串；客户端不能指定或覆盖；
- `tenant_ref`、`workspace_id`、`owner_subject_ref`；
- `display_name`、`description`、`application_kind`；
- `lifecycle_state`：`active` 或 `archived`；
- `record_version`：从 `1` 开始单调递增；
- `created_at`、`updated_at`、可空 `archived_at`；
- `created_by_actor_ref`、`updated_by_actor_ref`；
- `request_id`、`audit_ref`。

应用类型复用现有配置草案允许列表：`workflow_copilot`、`docs_qa`、`agent`、`prompt_application`。展示名长度为 2 到 120 个字符，描述最多 1000 个字符。标识、时间、所有者、作用域、生命周期、版本和审计字段均由服务端生成或投影，客户端不得修改。

记录不包含默认模型、默认协议、允许协议、API 密钥、配额、计费、模型服务凭据、端点、提示词、消息、调用输入输出、工作流定义或发布候选快照。默认模型和协议继续属于应用配置草案；运行和调用证据继续属于各自的历史存储库。

## 生命周期与并发语义

### 创建

- 创建请求只接收 `workspace_id`、`display_name`、`description` 和 `application_kind`。
- 租户、所有者、参与者、应用标识、状态、版本、时间和审计引用由服务端生成。
- 创建结果固定为 `lifecycle_state=active`、`record_version=1`。
- v1 不要求展示名唯一；应用标识承担稳定唯一性，避免并发同名创建产生隐式覆盖。

### 更新

- 更新使用完整可变元数据快照，不使用语义模糊的部分补丁。
- 请求必须携带 `workspace_id`、`expected_version`、`display_name`、`description` 和 `application_kind`。
- 只有 `active` 应用可更新；更新成功后版本加一并刷新 `updated_at`。
- 目录元数据更新属于开发测试态管理动作，不代表配置草案已发布。它会改变应用基线，使绑定旧 `updated_at` 的草案或发布候选进入既有基线漂移检查。

### 归档

- 归档是带预期版本的软归档，不物理删除记录。
- `active -> archived` 是 v1 唯一允许的生命周期转换；成功后版本加一，同时写入 `archived_at` 和新的 `updated_at`。
- v1 不提供恢复、反归档或删除端点。若后续需要恢复，必须独立设计其权限、冲突和下游重新启用语义。
- 已归档应用仍可通过显式归档筛选和精确详情读取，既有配置草案、候选版本与请求历史保持可读。
- 已归档应用禁止元数据更新、创建或保存新配置草案、创建或继续审查发布候选，以及从 Web 发起新的 API 接入 / 调试台交接。
- Gateway 原始上行 API 的生产级应用授权依赖 API 密钥和成员关系，本专题不把 Web 交接阻断解释为完整生产授权。直接调用绕过目录界面的风险必须在生产 API 分发专题中独立关闭。

### CAS 冲突

- 更新与归档都使用原子 `expected_version` CAS。
- 冲突不修改记录，只返回稳定失败码、当前脱敏版本和生命周期；不返回其他所有者的记录。
- 多标签页场景必须保留用户当前编辑，要求显式刷新当前记录后重新决定，不允许客户端自动覆盖。

## 作用域、身份与权限

所有操作绑定 `tenant_ref / workspace_id / application_id / owner_subject_ref`：

- `tenant_ref` 与 `owner_subject_ref` 只能来自已验证身份上下文，不能由请求正文、查询参数或路径覆盖。
- v1 采用所有者作用域：创建者成为不可变所有者；列表、详情、更新和归档只允许当前所有者访问，不实现所有权转移或工作区共享角色。
- 读取要求 `applications:read`；创建和更新要求 `applications:write`；归档要求独立 `applications:archive`。
- 显式开发请求头和签名测试令牌只服务开发测试态验收，不进入记录、响应正文、日志或浏览器持久化介质。
- `radish_oidc_integration_test` 模式下，工作区成员关系契约仍未成立；所有应用目录操作必须在存储库查询前返回 `workspace_membership_unavailable`，不得读取假数据或应用目录存储库。
- 身份缺失、租户不匹配、权限不足、成员关系不可用和作用域不一致都必须保持存储库零查询。

## API 边界

应用目录复用现有集合路由，并增加详情和生命周期写入：

- `GET /v1/user-workspace/applications`
- `POST /v1/user-workspace/applications`
- `GET /v1/user-workspace/applications/{application_id}`
- `PUT /v1/user-workspace/applications/{application_id}`
- `POST /v1/user-workspace/applications/{application_id}/archive`

列表和详情要求显式 `workspace_id`。列表默认只返回 `active`，允许受控 `lifecycle_state`、`application_kind`、`limit` 和 `cursor`；不允许客户端用 `owner_subject_ref` 查询其他所有者。稳定顺序为 `updated_at DESC, application_id DESC`，游标必须绑定租户、工作区、所有者和筛选摘要。

HTTP 消费端必须严格拒绝未知写入字段、作用域漂移、schema 漂移和禁止投影。写入正文不得包含 `application_id`、`tenant_ref`、`owner_subject_ref`、生命周期、版本结果、时间、审计字段或任何敏感材料。

## 稳定失败语义

至少固定以下失败码：

- `application_catalog_scope_denied`；
- `workspace_membership_unavailable`；
- `application_catalog_not_found`；
- `application_catalog_payload_invalid`；
- `application_catalog_secret_material_forbidden`；
- `application_catalog_version_conflict`；
- `application_catalog_archived`；
- `application_catalog_transition_invalid`；
- `application_catalog_cursor_invalid`；
- `application_catalog_store_unavailable`；
- `application_catalog_write_disabled`。

不存在和跨作用域读取统一返回不泄漏资源存在性的结果。存储失败、CAS 冲突、非法转换和敏感字段拒绝都不得产生部分写入或下游副作用。公开诊断只包含稳定错误码、脱敏摘要、请求标识和审计引用，不包含 SQL、DSN、端点、请求头、调用栈或其他所有者数据。

## 存储库与 PostgreSQL 开发测试态边界

`ApplicationCatalogRepository` 提供：

- `Create`；
- `Read`；
- `List`；
- `UpdateMetadata`；
- `Archive`；
- `RequireActive`，供应用配置草案与发布治理在写入前执行精确生命周期检查。

`memory_dev` 与 `postgres_dev_test` 必须共享相同的服务端标识生成、所有者作用域、排序、筛选、CAS、归档和脱敏投影语义。

PostgreSQL 使用独立 `application_catalog_records` 和 `application_catalog_schema_versions`，不复用控制面管理只读表、应用配置草案表、发布候选表或工作流表。主键固定为 `(tenant_ref, workspace_id, application_id)`；所有者、生命周期、更新时间和应用类型使用必要索引。归档只更新状态与审计元数据，不删除历史行。

迁移继续使用独立 manifest、checksum、advisory lock 和手动运行器；平台启动不自动迁移。迁移角色负责 DDL，运行角色只具必要 DML。连接、标记、checksum、查询或 CAS 失败不得回退到 `memory_dev` 或 `fake_store_dev`。

开发测试态不执行自动清理。归档记录保留到测试环境显式重建；生产保留、删除、恢复和合规策略后续独立设计。

## 与现有应用专题的交接

### 配置草案

- 创建配置草案时，应用标识、工作区和基线 `updated_at` 来自当前活跃目录记录。
- 保存新草案前由服务端 `RequireActive` 精确检查同一租户、工作区、应用和所有者；归档或不存在时不写草案。
- 既有草案 schema v1 暂时继续使用 `base_application_updated_at`，不在本专题静默升级草案 schema；目录更新或归档会触发现有基线漂移语义。
- 已归档应用的既有草案仍可读取和审查，但不能再保存新版本或交给新的调用路径。

### API 接入与 Gateway 调试台

- 活跃应用可以进入模型目录、接入示例和调试台交接；应用切换继续清除旧模型、选择、失败和请求标识。
- 已归档应用详情只显示历史审查入口，不派发新的应用 / 协议 / 模型交接事件。
- 请求历史继续按既有应用作用域读取脱敏记录；目录不复制调用输入输出或请求历史载荷。
- 本专题不改变 Gateway 上行协议、模型服务注册表、SSE 解析器、重试或回退策略。

### 发布治理

- 发布候选创建前由服务端精确读取活跃目录记录作为应用基线，不再扫描预置应用列表。
- 既有候选 schema v1 暂时继续使用 `base_application_updated_at`；目录更新保持现有 `application_base_revision_changed` 语义。
- 应用归档后，既有候选仍可读取，但晋级资格增加稳定归档阻塞项；不允许创建新候选或追加新的审查决定。
- 候选已批准仍不执行正式晋级，不把开发测试态目录更新当作候选晋级结果。

## Web 工作区

应用目录在现有应用页面形成完整管理路径：

- 默认离线模式继续读取现有 fixture，保持零网络请求和只读说明；不伪造写入成功。
- 显式开发测试态模式显示活跃 / 已归档筛选、分页、创建入口、应用详情、完整元数据编辑、版本信息、冲突处理和归档确认。
- 创建成功后选中新应用并清空旧应用上下文，再显示配置草案、API 接入和发布审查入口。
- 更新冲突保留本地编辑并显示服务器当前版本；用户显式刷新后重新决定。
- 归档确认必须展示将被关闭的新建 / 更新 / 调用交接能力；成功后从活跃列表移除并进入已归档详情。
- 已归档详情允许进入既有草案、候选版本和请求历史的只读审查，不显示可执行的新建、保存或调用按钮。
- URL 只保留稳定页面 / 区段标识；应用记录、编辑内容、请求标识和审计引用不写 `localStorage` 或 `sessionStorage`。

## 实施拆分

设计通过后只创建一张专项实现任务卡，按以下职责拆分验证，不派生同层准入文档链：

1. 建立应用领域、标识生成、校验、生命周期、内存存储库、精确读取、CAS 和稳定失败。
2. 在显式开发测试态门禁下接管现有应用集合路由，并新增详情、更新和归档路由；保持未启用模式的历史只读行为。
3. 建立独立 PostgreSQL schema、迁移、运行角色、存储模式选择器和破坏性集成测试。
4. 更新 Web 应用页面和上下文选择，完成创建、详情、更新、冲突、归档与下游交接。
5. 让配置草案保存和发布候选创建 / 审查消费精确活跃应用检查；请求历史保持只读。
6. 完成真实浏览器连续路径、服务重启恢复、泄漏检查、文档真相源和阶段收口。

## 验收方式

单元与集成测试至少覆盖：

- 服务端应用标识生成、类型 / 长度校验和敏感信息失败关闭；
- 创建、精确读取、列表、筛选、分页、完整元数据更新和软归档；
- 租户 / 工作区 / 应用 / 所有者隔离及跨作用域不泄漏；
- `applications:read`、`applications:write`、`applications:archive` 权限分离；
- OIDC 成员关系不可用时存储库零查询；
- 更新与归档 CAS、多并发单一成功者、归档后非法转换；
- 存储模式互斥、未知模式失败、PostgreSQL 连接 / 标记 / 查询失败不回退；
- PostgreSQL 迁移应用 / 重复应用 / 回滚 / 重新应用、运行角色 DDL 拒绝和服务重启恢复；
- 应用摘要兼容投影不伪造工作流或运行状态；
- 归档后配置草案保存、候选创建 / 审查和 Web 新调用交接被阻断，既有草案 / 候选 / 请求历史仍可读取；
- 应用切换清除旧编辑、模型、候选和请求标识状态。

真实浏览器在显式 PostgreSQL 开发测试态配置下完成：

1. 从空目录创建应用并进入详情；
2. 更新元数据并确认版本递增；
3. 建立配置草案，加载模型，完成单次响应 / 流式响应 / 取消和精确请求历史审查；
4. 创建发布候选并确认仍为 `promotion_blocked`；
5. 在第二标签页更新应用，使第一标签页旧版本更新冲突且不覆盖；
6. 以当前版本归档应用，确认活跃列表移除、已归档列表可读；
7. 确认新草案保存、候选创建 / 审查和调试台交接被阻断，而既有草案、候选和请求历史仍可审查；
8. 重启平台和 Web，确认应用版本、归档状态和历史交接仍一致；
9. 检查控制台、URL、`localStorage`、`sessionStorage` 和已提交产物不含敏感信息或用户输入输出。

相称验证包括 Go 单元测试 / 竞态测试 / 静态检查、Web 单元测试与生产构建、PostgreSQL 破坏性集成、`git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

## 停止线

- 不实现生产应用存储库、生产认证、Radish 工作区成员关系适配器、所有权转移或共享角色。
- 不实现反归档、物理删除、自动保留清理或正式应用晋级。
- 不实现生产 API 密钥签发 / 轮换 / 吊销、配额执行、限流、计费或成本账本。
- 不保存模型服务凭据、端点、`Authorization`、内部开发请求头、提示词、消息、调用输入输出或完整 HTTP 载荷。
- 不新增 Gateway 上行协议、模型服务注册表、协议适配器、SSE 解析器、自动重试 / 回退或负载均衡。
- 不扩展工作流工具、RAG、确认提交、业务写回、重放或恢复。
- 不把开发测试态应用目录、已批准候选版本、Web 归档阻断或本地调用成功解释为生产应用生命周期或生产授权已就绪。

## 设计评审后下一步

设计评审通过后，创建单一实现任务卡，先落地应用领域、`memory_dev`、带作用域 API 和负向测试，再完成 PostgreSQL、Web 与真实浏览器纵向闭环。实现过程中如果发现必须改变现有应用草案 / 候选 schema、Gateway 授权或工作流执行边界，应停止当前批次并回到对应功能专题更新设计，不能在应用目录实现中顺带扩张。
