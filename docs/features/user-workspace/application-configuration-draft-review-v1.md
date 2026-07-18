# 用户工作区应用配置草案与审查 v1

更新时间：2026-07-15

状态：`application_configuration_draft_review_v1_complete`

## 当前实现结果

2026-07-12 已完成独立应用配置草案领域、仅开发的校验 / 保存 / 读取 / 列表 API、`memory_dev` 与 PostgreSQL 开发测试态存储库、显式迁移、作用域 / 所有者隔离、CAS 版本冲突、敏感信息失败关闭、Web 配置 / 校验 / 比较 / 恢复 / 冲突审查，以及到现有 API 接入区和调试台的应用 / 协议 / 模型交接。

Web 49 项测试与生产构建、平台全量 Go 测试和 PostgreSQL 集成套件通过。真实浏览器以 `app_docs_assistant` 加载 6 个模型，保存版本 1，在第二标签页保存版本 2，第一标签页触发冲突并显式恢复；随后完成 Responses 单次响应、流式响应、用户取消，并以同一 `request_id` 打开 `408 / BRIDGE_WORKER_CANCELED / postgres_dev_test` 请求历史详情。控制台为 0 个错误 / 0 个警告，URL 只含稳定区段锚点，`localStorage` / `sessionStorage` 为空。

2026-07-14 已补齐 SQLite 开发持久化：配置草案通过共享 runtime 使用独立 migration 和 repository，保持作用域、稳定列表、创建元数据与保存 CAS；memory / SQLite 运行同组契约，真实文件覆盖多连接并发、关闭失败、重启恢复和敏感材料禁入。该能力只服务统一 `sqlite_dev` 规划，不改变既有 PostgreSQL 手动迁移与生产停止线。

2026-07-15 已补齐 API 密钥本地产品链的模型目录复用：Playground 以 Bearer 凭据完成 `/v1/models` 严格校验后，只派发应用标识、公开模型字段和当前选择；配置面板重新校验事件并复用目录，不接收 API 密钥、`Authorization`、请求头或 Gateway 输入输出。应用切换会淘汰旧目录，避免凭据或模型选择跨应用扩散。

## 功能目标

让内部开发者从当前用户工作区应用建立独立配置草案，完成编辑、模型与协议校验、开发测试态保存、恢复、版本冲突审查和配置差异比较，并把已校验的应用、协议和模型交给现有 API 接入工作区与 Gateway 调试台。

本功能建立应用配置草案的领域与持久化边界，但不直接修改开发测试态应用目录或离线应用摘要，不创建正式应用，也不提供发布、删除或生产授权。

## 用户流程

1. 用户从应用列表选择一个应用，进入现有应用详情。
2. 应用配置工作区以当前应用的公开字段建立内存工作副本；切换应用会清除旧表单、校验、版本、冲突和交接状态。
3. 用户编辑展示名、描述、应用类型、默认协议、默认模型与允许协议集合。
4. 默认离线模式不发网络请求，允许本地编辑、校验和比较，但不保存、不恢复、不加载模型。
5. 显式开发测试态模式下，用户可由本面板的现有消费端加载当前应用作用域模型目录，也可复用 Playground 已严格校验的脱敏目录；只有通过再次校验的模型标识和协议能力可进入草案，Bearer 凭据不随目录交接。
6. 用户执行本地校验；通过后可调用仅开发的应用草案路由保存。首次保存得到版本 `1`，后续保存必须携带预期版本。
7. 用户可以刷新草案列表、恢复当前应用下自己的草案，并比较草案与只读应用基线。
8. 版本冲突时保留当前内存修改并展示服务器当前版本；用户必须显式恢复已保存版本，或以当前版本为新基线继续编辑，不能静默覆盖。
9. 已通过校验的协议与模型可以交给现有应用 API 接入区和调试台；调用与请求历史继续由既有 Gateway 消费端和同一 `request_id` 链路承载。

## 数据来源与职责

| 数据 | 真相源 / 所有者 | 本功能用途 | 禁止替代来源 |
| --- | --- | --- | --- |
| 当前应用 | 显式应用目录或默认离线应用列表 / 应用详情 | 建立比较基线和应用作用域 | 不从 URL 或浏览器存储恢复 |
| 模型目录 | 现有 `GET /v1/models` 消费端或 Playground 脱敏目录事件 | 校验默认模型和协议兼容性 | 不复制模型服务注册表，不接收 Bearer 凭据或静态伪造实时目录 |
| 应用配置草案 | 应用配置草案服务 / 开发测试态存储库 | 保存脱敏配置草案、版本和审计元数据 | 不复用工作流草案存储库，不写控制面只读 fixture |
| 测试调用 | 现有 Gateway 调试台 | 消费已校验应用 / 协议 / 模型交接 | 不新增协议适配器、SSE 解析器或请求状态模型 |
| 调用审查 | 现有 Gateway 请求历史 | 使用调用产生的同一 `request_id` | 不把草案内容补写到 Gateway 请求历史 |

## 应用作用域与身份

- 所有保存、读取和列表操作都必须绑定 `tenant_ref / workspace_id / application_id / owner_subject_ref`。
- 应用标识只能来自当前应用列表选择；请求正文、查询参数和开发作用域请求头必须一致。
- 开发测试态参与者上下文复用现有显式开发身份头，但应用配置草案使用独立权限：`application_drafts:read` 与 `application_drafts:write`。
- 列表只返回当前作用域与所有者的脱敏摘要；读取和保存不允许跨应用、工作区或所有者。
- 本批次不接 Radish OIDC、真实成员关系或生产授权；开发请求头不进入界面、草案载荷、日志或示例。

## 配置草案模型

草案 schema 固定为 `application_configuration_draft.v1`，包含：

- `draft_id`：稳定短标识；
- `workspace_id`、`application_id`、`base_application_updated_at`；
- `display_name`、`description`、`application_kind`；
- `default_protocol`、`default_model`、`allowed_protocols`；
- `draft_version`、`validation_state`、创建 / 更新时间和参与者引用；
- 脱敏校验发现与请求审计元数据。

草案不允许包含 API 密钥、密钥哈希、`Authorization`、cookie、模型服务凭据 / 端点、内部调用方请求头、提示词 / 消息输入、模型输出或业务数据。服务端必须递归检查禁止字段名，并限制字符串、数组和载荷大小。

## 校验与比较

校验规则包括：

- 展示名、描述、应用类型、草案标识和作用域字段的长度与字符约束；
- 应用类型必须来自当前受控允许列表；
- 允许协议只能包含 Chat Completions、Responses、Messages，且去重后非空；
- 默认协议必须属于允许协议；
- 开发测试态加载过模型目录时，默认模型必须存在，默认协议必须属于该模型公开支持协议；
- 载荷任意层级出现敏感信息、凭据、授权、请求头、端点或原始调用内容字段时按失败关闭处理。

比较视图只展示公开配置字段的变更前 / 变更后 / 未变化状态，不展示参与者请求头、存储库元数据、凭据或测试输入输出。模型目录刷新后模型不可用时，草案保持原值但进入阻塞校验状态，不能交给调试台。

## 状态与失败语义

前端状态固定覆盖：

- `offline`、`unsaved`、`validating`、`invalid`、`valid`；
- `saving`、`saved`、`loading`、`restored`；
- `version_conflict`、`store_failure`、`scope_denied`。

稳定服务端失败至少包括：

- `application_draft_scope_denied`；
- `application_draft_not_found`；
- `application_draft_payload_invalid`；
- `application_draft_secret_material_forbidden`；
- `application_draft_version_conflict`；
- `application_draft_store_unavailable`；
- `application_draft_write_disabled`。

存储库故障必须显式失败，`postgres_dev_test` 不得回退到内存存储；冲突响应只能返回当前版本和脱敏发现，不能返回他人的草案。

## 持久化与隐私边界

- 离线模式必须保持零请求。
- 未保存编辑、模型目录、比较选择和调试台输入输出只存在当前 React 组件内存。
- 不写入 URL 查询 / 锚点载荷、`localStorage`、`sessionStorage`、工作流草案、工作流运行、Gateway 请求历史载荷或业务真相源。
- `memory_dev` 用于当前进程内测试，聚合 `sqlite_dev` 用于本地连续产品链，`postgres_dev_test` 用于迁移、角色、方言、重启与并发验证；三者都不代表生产存储库。
- 已提交文档、测试 fixture 和浏览器证据不得保存真实敏感信息、DSN、用户输入或模型输出。

## API 与存储库边界

新增仅开发路由：

- `POST /v1/user-workspace/application-drafts/validate`
- `POST /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts`
- `GET /v1/user-workspace/application-drafts/{draft_id}`

存储库契约只提供校验之外的保存 / 读取 / 列表操作；校验属于领域服务。内存、SQLite 与 PostgreSQL 适配器必须共享同一 CAS、作用域和脱敏投影语义。SQLite migration 由聚合本地产品 runtime 管理；PostgreSQL migration 只由显式开发测试态运行器执行，平台启动不自动迁移。

## 实施拆分

1. 建立应用配置草案领域类型、校验、内存存储、HTTP 路由、稳定失败与单元测试。
2. 建立 PostgreSQL schema、标记、手动迁移运行器、CAS 存储库、存储模式选择器和集成测试。
3. 建立独立延迟加载应用配置工作区，复用模型目录消费端与调试台交接。
4. 完成草案列表、恢复、比较、版本冲突处置和应用切换隔离。
5. 完成真实浏览器连续流程、泄漏检查、文档真相源和阶段收口。

## 验收方式

单元与集成测试至少覆盖：离线零请求；本地合法 / 非法校验与敏感信息失败关闭；保存 / 读取 / 列表和脱敏摘要；租户 / 工作区 / 应用 / 所有者隔离；CAS 版本冲突且不覆盖；SQLite 重启恢复与 PostgreSQL 迁移应用 / 回滚 / 重新应用、重启恢复和不回退；`/v1/models` 兼容性校验；Playground 脱敏目录事件校验且不携带凭据；应用切换清除旧状态；草案到 API 接入区 / 调试台的应用、协议、模型交接；调试台到请求历史的同请求标识交接继续通过既有测试。

真实浏览器在显式开发测试态配置下完成：选择应用、建立草案、加载模型、编辑、校验、保存、刷新恢复、制造并审查版本冲突、比较配置、交给调试台、完成单次响应 / 流式响应 / 取消并打开精确请求历史详情。验收同时检查控制台错误、URL、`localStorage` 和 `sessionStorage`。

相称验证包括 Web 单元测试、生产构建、Go 单元与集成测试、迁移往返、`git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

## 与 Workflow RAG binding 的组合边界

- 未绑定草案继续使用 `application_configuration_draft.v1`；绑定草案使用 v2，且只新增 `workflow_rag_binding_ref={binding_id,binding_version,binding_digest}` 与服务端 canonical `draft_digest`，不复制 dataset、snapshot、query、fragment、评测指标或配置正文真相源。
- 首次 attach / replace 必须从已批准 promotion candidate 的精确 source draft 开始。服务端重读当前 binding 与全部权威知识来源，确认除 binding ref 外没有配置修改后，才通过既有 expected-version CAS 创建下一版草案。
- Web 只列出当前 eligible 的不可变 binding，并把“恢复 source draft”和“attach binding”保留为两个显式动作；批准 promotion candidate 不自动修改草案，attach 也不创建发布候选。
- memory、SQLite 与 PostgreSQL 继续复用本专题既有 repository 和 JSON payload 表，无新增草案 store、selector、DSN、pool 或 migration。双数据库连续链与真实浏览器已验证 v1 source draft → v2 binding draft、重启恢复与 no-fallback。

## 停止线

- 不创建、发布、删除或直接更新正式应用。
- 不在草案中接收或持久化 API 密钥、`Authorization` 或凭据交接；开发测试态 API 密钥生命周期由独立专题承接，生产密钥、配额、计费、成本账本、生产认证和真实成员关系仍保持关闭。
- 不保存模型服务凭据、端点、`Authorization`、内部开发请求头、测试输入输出或完整 HTTP 载荷。
- 不新增 Gateway 上行协议、模型服务注册表、协议适配器、SSE 解析器、回退或负载均衡。
- 不把开发测试态草案保存、模型校验或调用成功解释为生产就绪、真实模型服务 SLA 或正式应用已发布。
- 不扩展工作流工具、确认、写回、重放或恢复。
