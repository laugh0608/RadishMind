# Saved Workflow Draft Durable Store Preconditions v1 专题

更新时间：2026-06-15

## 专题定位

`Saved Workflow Draft Durable Store Preconditions v1` 是 `Saved Workflow Draft v1` 从 memory dev store 迁向未来 durable store 前的设计专题。它只固定迁移前置条件、scope / ownership 契约、冲突语义、no sample fallback 和 dev store 到未来 repository adapter 的切换停止线。

本专题不是 durable persistence 实现，不创建数据库 schema、migration、store selector、repository adapter、Radish OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_durable_store_preconditions_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 platform Go domain service、memory dev store、save / read / validate、版本冲突、失败语义和 no sample fallback。
- `Dev-only Saved Draft Consumer` 已有默认关闭的 dev-only HTTP route 与 web consumer；保存仍要求 dev auth、workspace / application header 和 write enablement。
- `Workflow Draft Editing Entry v1` 与 `User Workspace Draft Creation v1` 已让用户从 Workspace Home / workflow definitions 创建本地草案并进入 Draft Designer。
- [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md) 已固定 `draft_repository_contract_preconditions_defined`，覆盖 future `SavedWorkflowDraftRepository` actor context、operation matrix、request / result contract、failure policy 和 sanitized projection policy。
- [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md) 已固定 `draft_schema_migration_preconditions_defined`，覆盖 future durable store logical schema、index strategy、migration gate、failure mapping 和 artifact guard。
- [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) 已固定 `draft_auth_context_preconditions_defined`，覆盖 future repository actor context 的身份来源、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization policy。
- [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) 已固定 `draft_store_selector_enablement_preconditions_defined`，覆盖 future store mode、selector gate、failure mapping、no fallback、no side effects 和 artifact guard。
- 当前没有 durable repository、数据库、schema migration、store selector、Radish OIDC、production API consumer 或执行链路。

## Draft Scope

durable store 迁移前必须把草案 scope 固定为不可混淆的复合边界：

- `tenant_ref`：未来来自 Radish OIDC / platform auth context；当前 dev-only 路径只能使用 fake / dev auth，不声明 production auth ready。
- `workspace_id`：用户工作区归属，必须属于当前 tenant。
- `application_id`：草案所在 AI application / workflow application，必须属于当前 workspace。
- `draft_id`：草案主键，只在 `tenant_ref + workspace_id + application_id` 内唯一解释。
- `source_definition_id` 与 `base_definition_version`：只表示草案来源，不表示草案已发布或覆盖 workflow definition。
- `draft_version`：乐观并发版本，只用于保存冲突检测。
- `schema_version`：草案 payload schema 版本，不能替代数据库 schema migration 版本。

读取、保存和校验都必须显式携带 `workspace_id` 与 `application_id`。缺少 scope、scope mismatch、跨 workspace / application 访问或 actor 不具备 scope 时，必须返回 `draft_scope_denied`，不得返回草案主体。

## Owner / Workspace 归属

durable store 迁移前必须区分 owner、actor 和 workspace：

- `owner_subject_ref`：草案归属主体；未来来自 Radish OIDC subject 或团队 ownership 规则。
- `created_by_actor_ref` / `updated_by_actor_ref`：具体操作人；当前 dev-only 路径只能来自 fake / dev auth subject binding。
- `workspace_id` / `application_id`：访问与列表过滤的主 scope，不能由请求 query 覆盖 auth context。
- owner 不是跨 workspace 读写授权；共享、转让、团队可见性需要后续独立权限专题。

当前 memory dev store 可以继续只保留 `CreatedByActorRef` 和 `UpdatedByActorRef`，但进入 durable repository adapter 前，必须补齐 owner 字段、workspace membership 判断、tenant predicate 和 actor audit 字段的 contract / fixture / test。

## Version Conflict

durable store 必须沿用现有 `draft_version_conflict` 语义：

1. 保存请求必须携带客户端基于的 `expected_draft_version`。
2. 当前 store 中 `draft_version` 与 expected 不一致时，拒绝覆盖。
3. 响应只返回当前版本 metadata，例如 `current_draft_version`、`current_updated_at`、`current_updated_by_actor_ref` 和 request / audit metadata。
4. UI 必须保留用户本地草案，不回退 sample，也不静默用 server 版本覆盖本地编辑。
5. 重新保存必须由用户或明确交互基于当前版本再次触发。

冲突不是 store unavailable，也不是 schema unsupported；后续 repository adapter 不能把冲突吞成普通 save failure。

## No Sample Fallback

durable store 迁移必须继续 fail closed。以下情况都不得回退 offline sample 或 fixture：

- `draft_scope_denied`
- `draft_not_found`
- `draft_store_unavailable`
- `draft_schema_version_unsupported`
- `draft_version_conflict`
- 未来 repository / database mode 禁用、未配置或 contract mismatch

UI 可以继续展示 sample / unsaved local draft，但必须标记为 sample / local-only，不能把它包装成 saved durable record。

## Store 切换停止线

当前唯一可用 store 是 memory dev store。未来 repository adapter 只能在下列条件全部满足后进入实现任务：

1. [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md) 明确 `SavedWorkflowDraftRepository` contract 输入、输出、scope predicate、owner 字段、version conflict、failure code 和 sanitized projection。
2. committed schema / migration preconditions 已固定，并覆盖 tenant / workspace / application / draft / owner / version 索引；真实 schema artifact、DDL review evidence 和 SQL migration 仍需后续独立批次。
3. store selector 设计已固定，默认仍为 dev store；repository / database mode 不允许隐式 fallback 到 dev store。
4. Radish OIDC / auth context contract 已固定；没有 token validation 前不得声明 production auth ready。
5. repository contract smoke、adapter smoke、no sample fallback、scope denied、version conflict、store unavailable 和 no side effects 测试已进入 fast 或专项验证链路。
6. public production API、publish、run、executor、confirmation、writeback、replay 仍作为独立专题，不随 durable store adapter 一起打开。

未来可以预留 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` 这类配置键，但本专题不得创建正式 config entry，也不得实现 selector。

## 验收方式

- 新增 task card 固定本专题为 precondition-only。
- 新增 fixture / checker 固定 draft scope、owner / workspace 归属、version conflict、no sample fallback、store 切换停止线和 forbidden implementation artifact。
- checker 进入 `./scripts/check-repo.sh --fast`。
- 本批至少运行专项 checker 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建真实数据库、migration、SQL、repository adapter、store selector、Radish OIDC、token validation 或 production API consumer。
- 不实现 durable persistence、saved draft list、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 memory dev store、dev-only route、validation summary、`valid_for_review` 或本专题 fixture 解释为 durable store ready。
