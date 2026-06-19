# Workflow Saved Draft Durable Store Preconditions v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-durable-store-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_durable_store_preconditions_defined`

## 目标

在 `User Workspace Draft Creation v1` 完成后，为 `Saved Workflow Draft v1` 后续迁向 durable store 固定前置设计：草案 scope、owner / workspace 归属、version conflict、no sample fallback，以及 memory dev store 到未来 repository adapter 的切换停止线。

本任务卡只定义迁移前置条件和静态检查，不实现 durable persistence、真实数据库、Radish OIDC、production API、store selector、repository adapter、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Durable Store Preconditions v1 专题](../features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
- [Dev-only Saved Draft Consumer 实现专题](../features/workflow/dev-only-saved-draft-consumer.md)
- [User Workspace Draft Creation v1 专题](../features/workflow/user-workspace-draft-creation-v1.md)
- [User Workspace 设计与开发文档](../features/user-workspace.md)

## 本轮交付

- 新增 durable store migration preconditions 细专题，固定 precondition-only 状态。
- 新增 `workflow-saved-draft-durable-store-preconditions-v1` fixture / checker。
- checker 接入 fast baseline，校验文档引用、task card 状态、scope / owner / conflict / no fallback / store switch 停止线和禁止提前出现的实现 artifact。
- 同步更新 workflow 入口、Saved Workflow Draft 主专题、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Scope Contract

进入 durable repository adapter 前，草案必须按以下复合 scope 解释：

- `tenant_ref`
- `workspace_id`
- `application_id`
- `draft_id`
- `source_definition_id`
- `base_definition_version`
- `draft_version`
- `schema_version`

要求：

- 读、写、校验都必须显式携带 `workspace_id` 和 `application_id`。
- `draft_id` 只在 `tenant_ref + workspace_id + application_id` 内可解释。
- scope mismatch、缺少 scope、跨 workspace / application 访问和 actor scope 不满足，均返回 `draft_scope_denied`。
- scope denial 不返回草案主体，不回退 sample。

## Owner / Workspace Contract

durable store 迁移前必须明确：

- `owner_subject_ref` 是草案归属主体。
- `created_by_actor_ref` / `updated_by_actor_ref` 是操作人审计字段。
- `workspace_id` / `application_id` 是访问和列表过滤主边界。
- owner 不等于跨 workspace 授权；共享、转让和团队可见性后续另开权限专题。
- 当前 dev auth subject binding 只能作为开发态 actor source，不声明 Radish OIDC / production auth ready。

## Version Conflict Contract

- save 必须使用 `expected_draft_version`。
- 当前版本不同返回 `draft_version_conflict`。
- 冲突不得覆盖当前 store 记录。
- 响应只返回 current version metadata、request / audit metadata 和 failure code。
- UI 必须保留本地草案，不回退 sample，也不自动覆盖本地编辑。

## No Sample Fallback Contract

以下失败必须 fail closed：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_write_disabled`
- 未来 `repository_store_disabled` / `invalid_draft_store_mode` / `draft_store_contract_mismatch`

UI 允许展示 sample 或 local-only 草案，但必须继续标记为 `sample` / `unsaved_local`，不能包装成 saved durable record。

## Store Switch Stop Line

本批不得创建或实现：

- `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE` 正式 config entry
- `SavedWorkflowDraftRepository` interface / adapter
- `workflow_saved_draft_store_selector`
- SQL / migration / database schema
- Radish OIDC token validation
- production API consumer

未来打开 repository adapter 前，必须先满足 repository contract、schema / migration、store selector、production auth contract、adapter smoke、no sample fallback、version conflict 和 no side effects 测试。

## 验收口径

- `workflow-saved-draft-durable-store-preconditions-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 precondition-only，不声明 durable store ready。

## 停止线

- 不实现 durable persistence、repository adapter、schema migration、store selector、真实数据库、Radish OIDC、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把当前 memory dev store、dev-only HTTP route、本地创建草案、validation summary 或 `valid_for_review` 解释为 durable persistence、publish、run 或 production readiness。
