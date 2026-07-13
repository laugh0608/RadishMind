# 用户工作区应用发布治理与晋级审查 v1

更新时间：2026-07-13

状态：`application_publish_governance_promotion_v1_complete`

## 功能目标

让内部开发者把当前应用下已保存、已校验的配置草案固定为不可变发布候选版本，完成版本绑定、配置摘要校验、调用证据引用、审查决定、漂移识别和晋级资格复核。

本功能只建立应用发布治理与开发测试态候选审查链。候选版本通过审查不等于正式应用已发布；在真实应用存储库、Radish OIDC / 成员关系、发布所有者和生产授权未成立时，晋级必须保持阻塞。

## 用户流程

1. 用户从应用列表选择应用，进入应用详情和应用配置草案。
2. 用户恢复一个已保存、校验状态为 `valid` 的草案版本；未保存编辑不能直接创建候选版本。
3. 发布审查工作区重新从服务端读取 `draft_id + draft_version`，不接受浏览器提交的完整配置快照。
4. 服务端重新校验草案作用域、schema、校验状态和当前应用只读基线，并计算脱敏配置摘要。
5. 用户可添加零个或多个脱敏 Gateway `request_id` 作为审查引用。候选版本只保存引用，不复制请求历史载荷；精确记录仍由 Gateway 请求历史校验和展示。
6. 创建成功后得到不可变 `application_publish_candidate.v1`。草案后续变化不会改写旧候选版本，只能创建新候选版本。
7. 审查人在显式开发测试态模式记录 `approve`、`reject`、`request_changes` 或 `withdraw`；每次决定使用预期审查版本，冲突时不得覆盖。
8. 工作区比较候选快照、当前草案版本和应用基线，展示候选版本已被取代、草案漂移与应用基线漂移。
9. 晋级资格汇总审查状态、版本漂移、正式存储库、认证 / 成员关系、发布所有者和晋级启用状态。当前生产依赖未成立时返回稳定阻塞项，不伪造发布成功。
10. 用户可以从候选版本打开现有 API 接入区、调试台和精确请求历史详情；这些交接不新增 Gateway 适配器、SSE 解析器或请求历史存储库。

## 数据来源与职责

| 数据 | 真相源 / 所有者 | 本功能用途 | 禁止替代来源 |
| --- | --- | --- | --- |
| 应用基线 | 现有控制面只读 `ApplicationSummary` 存储库 | 创建时绑定 `updated_at`，读取时检查基线漂移 | 不从 URL、浏览器存储或候选请求正文伪造 |
| 应用配置草案 | 应用配置草案存储库 | 服务端读取精确保存版本并生成不可变快照 | 不接受客户端完整快照，不复用工作流草案 |
| 发布候选版本 | 独立发布候选开发测试态存储库 | 保存不可变配置快照、摘要、审查历史和脱敏元数据 | 不写正式应用真相源 |
| Gateway 证据 | 现有请求历史 | 用同一 `request_id` 精确审查调用 | 候选版本不复制请求、响应、提示词、用量载荷或模型服务原始材料 |
| 晋级阻塞项 | 发布治理服务 | 计算晋级资格和稳定失败 | 不把 Web 文案当授权或发布结果 |

## 应用作用域与角色

- 创建 / 列表 / 读取 / 审查必须绑定 `tenant_ref / workspace_id / application_id / owner_subject_ref`。
- 请求正文、查询参数、当前应用和开发作用域请求头必须一致。
- 创建要求 `application_publish_candidates:write`；列表 / 读取要求 `application_publish_candidates:read`；审查要求独立 `application_publish_candidates:review`。
- 开发测试态审查人可以与候选版本所有者相同，用于验证状态机和 CAS；这不证明生产职责分离或真实审批身份已经成立。
- 候选列表只返回当前作用域与所有者的脱敏摘要；跨应用、工作区、租户或所有者必须失败关闭。

## 候选版本模型与不可变边界

schema 固定为 `application_publish_candidate.v1`，包含：

- `candidate_id`、`workspace_id`、`application_id`；
- `draft_id`、`draft_version`、`draft_digest`；
- `base_application_updated_at`；
- 脱敏配置快照：`display_name`、`description`、`application_kind`、`default_protocol`、`default_model`、`allowed_protocols`；
- 去重排序后的脱敏 `evidence_request_ids`；
- `candidate_state`、`review_version`、只追加审查记录；
- 创建 / 更新时间、参与者引用、请求 / 审计引用；
- 动态计算的 `promotion_eligibility`。

候选版本创建后，配置快照、草案绑定、摘要、应用基线和证据引用不允许更新。审查只追加决定记录并更新审查状态 / 版本。草案或应用基线变化时，旧候选版本仍可读取，但晋级资格必须反映漂移或被取代状态。

摘要由服务端对规范化后的脱敏配置快照进行确定性 JSON 编码后计算 SHA-256，不包含参与者、请求标识、审计引用、审查、时间戳或任何敏感信息。

## 审查状态机

候选版本初始状态为 `pending_review`，允许以下决定：

- `approve`：进入 `approved`；
- `reject`：进入 `rejected`；
- `request_changes`：进入 `changes_requested`；
- `withdraw`：进入 `withdrawn`。

`rejected`、`changes_requested` 和 `withdrawn` 为当前候选版本终态；后续修改草案并创建新候选版本。`approved` 仍可被所有者显式撤回，但不提供真实晋级。所有决定必须带 4 到 500 字符的脱敏原因，使用预期审查版本，并追加参与者、请求和审计元数据。

## 晋级资格

`promotion_eligibility` 输出 `eligible`、`status` 和有序阻塞项。当前 v1 至少检查：

- 候选版本是否为 `approved`；
- 当前应用 `updated_at` 是否与候选基线一致；
- 当前已保存草案的版本 / 摘要是否仍与候选绑定一致；
- 候选版本是否已被同一草案的后续候选版本取代；
- 正式应用存储库是否可用；
- 生产认证 / 成员关系是否已验证；
- 发布所有者是否已配置；
- 晋级运行时是否显式启用。

本批正式存储库、生产认证、发布所有者与晋级运行时均未启用，因此即使候选版本已批准，也必须返回 `status=promotion_blocked` 和对应阻塞项，不创建晋级端点，不修改应用只读模型。

## API 与存储库边界

新增仅开发路由：

- `POST /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates`
- `GET /v1/user-workspace/application-publish-candidates/{candidate_id}`
- `POST /v1/user-workspace/application-publish-candidates/{candidate_id}/reviews`

创建请求正文只允许 `candidate_id`、`draft_id`、`expected_draft_version` 和 `evidence_request_ids`。审查请求正文只允许 `expected_review_version`、`decision` 和 `reason`。存储库契约只提供创建 / 读取 / 列表 / 追加审查；应用基线、草案校验、摘要、晋级资格和失败映射属于领域服务。

内存与 PostgreSQL 开发测试态适配器必须共享相同的不可变创建、审查 CAS、作用域、排序和脱敏投影语义。PostgreSQL 使用独立 schema 标记、手动迁移运行器以及运行角色 / 迁移角色分离；连接、标记或查询失败不得回退到内存存储。

## 失败语义

稳定失败至少包括：

- `publish_candidate_scope_denied`；
- `publish_candidate_not_found`；
- `publish_candidate_payload_invalid`；
- `publish_candidate_secret_material_forbidden`；
- `publish_candidate_draft_not_found`；
- `publish_candidate_draft_version_conflict`；
- `publish_candidate_draft_invalid`；
- `application_base_revision_changed`；
- `publish_candidate_immutable_conflict`；
- `publish_candidate_review_version_conflict`；
- `publish_candidate_review_transition_invalid`；
- `publish_candidate_store_unavailable`；
- `publish_candidate_write_disabled`。

存储失败、作用域拒绝、草案不匹配、基线漂移和 CAS 冲突都不得创建或修改候选版本。冲突响应只返回当前脱敏版本 / 状态，不返回他人候选版本或草案。

## 隐私与前端状态边界

- 默认离线模式零请求；候选标识、审查原因、证据请求标识和选择状态只在当前 React 组件内存。
- 不写 URL 查询 / 锚点载荷、`localStorage`、`sessionStorage`、草稿、运行记录、Gateway 载荷或业务真相源。
- 候选版本和审查原因禁止包含 API 密钥、`Authorization`、cookie、模型服务凭据 / 端点、内部开发请求头、提示词 / 消息、模型输出、完整 HTTP 载荷或 DSN。
- Web 对创建 / 列表 / 读取 / 审查响应做严格校验和递归禁止字段扫描；作用域漂移、schema 漂移或敏感字段一律失败。
- 请求历史交接只携带 `request_id + application_id`；接入区 / 调试台交接只携带候选版本的应用、协议和模型。

## 实施拆分

1. 建立候选版本 / 审查 / 晋级资格领域、内存存储库、HTTP 路由、作用域与单元测试。
2. 建立 PostgreSQL schema、标记、手动运行器、不可变创建、审查 CAS 和集成测试。
3. 建立独立延迟加载发布审查工作区，消费已保存草案绑定、候选列表 / 详情 / 审查和漂移比较。
4. 复用现有接入区 / 调试台 / 请求历史事件，完成应用切换隔离与离线零请求。
5. 完成真实浏览器多标签冲突、漂移、审查决定、精确请求历史交接、泄漏检查和阶段文档收口。

## 验收方式

单元与集成测试至少覆盖：离线零请求、服务端草案重读、候选摘要、不可变创建、作用域 / 所有者隔离、审查 CAS、非法状态转换、敏感信息失败关闭、应用基线漂移、草案版本 / 摘要漂移、候选版本被取代、PostgreSQL 迁移 / 回滚 / 重新应用、重启恢复、运行角色 DDL 拒绝和不回退。

真实浏览器应完成：选择应用、恢复已保存草案、创建候选版本、查看摘要 / 快照 / 阻塞项、添加并打开精确请求历史标识、在第二标签页制造审查 CAS 冲突、记录批准 / 拒绝 / 要求修改 / 撤回中的主要路径、验证已批准仍为 `promotion_blocked`、切换应用后旧候选版本与原因不串入，并检查控制台、URL、`localStorage` 和 `sessionStorage`。

相称验证包括 Web 单元测试、生产构建、Go 单元测试 / 竞态测试 / 静态检查、PostgreSQL 集成测试、`git diff --check`、`./scripts/check-repo.sh --fast` 和完整 `./scripts/check-repo.sh`。

## 当前实现与复验结果

- 平台已实现候选版本 / 审查 / 晋级资格领域、`memory_dev` 与 PostgreSQL 开发测试态存储库、四条带作用域的仅开发路由、独立迁移 / 标记 / 手动运行器和失败关闭的存储模式选择器；创建只接收草案绑定，由服务端重读草案并计算摘要。
- Web 已实现独立延迟加载发布审查工作区，支持选择有效的已保存草案、候选版本创建 / 列表 / 恢复、快照比较、审查 CAS、晋级阻塞项，以及接入区 / 调试台 / 精确请求历史交接；读取候选版本与记录审查使用不同状态语义。
- 单元和 PostgreSQL 集成测试覆盖不可变创建、作用域 / 所有者、敏感信息守卫、草案 / 基线漂移、被取代状态、审查转换 / CAS、重启恢复、运行角色 DDL 拒绝、回滚 / 重新应用和不回退。
- 真实浏览器已完成 `app_flow_copilot` 草案保存、Responses 单次响应、同一 `request_id` 请求历史详情、候选版本创建、第二标签页批准、第一标签页过期审查冲突恢复，以及切换到 `app_docs_assistant` 后旧候选版本 / 原因 / 证据清空。
- 已批准候选版本仍稳定显示 `promotion_blocked`，并列出正式应用存储库、生产认证、发布所有者和晋级运行时未成立的四项阻塞；没有发生正式应用变更。
- 浏览器控制台无错误 / 警告，URL 只保留稳定区段锚点，`localStorage` 与 `sessionStorage` 均为空。

## 停止线

- 不新增正式晋级端点，不创建、更新、发布或删除正式应用。
- 不接生产 OIDC、成员关系、授权、应用存储库、凭据后端或发布所有者运行时。
- 不实现生产 API 密钥、配额执行、计费、成本账本、模型服务凭据或可信 SLA。
- 不新增 Gateway 上行协议、模型服务注册表、协议适配器、SSE 解析器、回退或负载均衡。
- 不把候选版本已批准、开发测试态审查、模型调用或请求历史记录解释为生产发布。
- 不扩展工作流工具、确认、写回、重放或恢复。

## 后续顺位

管理端认证只读、租户 / 审计 PostgreSQL 开发测试态运行时与确定性 OIDC 边界已完成，真实 Radish 联调已延后。下一产品设计转向“用户工作区应用目录与生命周期（开发/测试态）v1”，先解决应用仍依赖预置摘要、用户不能创建和管理应用的缺口，并把新应用接入现有配置草案、API 接入区与发布审查。该专题不启用正式晋级，生产认证 / 成员关系、生产 API 密钥、配额和计费继续关闭。
