# Workflow Saved Draft Store Selector Enablement Preconditions v1 任务卡

更新时间：2026-06-15

## 任务标识

- 切片：`workflow-saved-draft-store-selector-enablement-preconditions-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_store_selector_enablement_preconditions_defined`

## 目标

在 auth context preconditions 之后，固定 saved workflow draft store mode、selector gate、failure mapping、no fallback、no side effects、dev flag boundary 和 implementation artifact guard。

本任务卡只定义 store selector enablement preconditions，不创建正式 config entry、selector 函数、selector 类型、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、production API、saved draft list、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)
- [Saved Workflow Draft Durable Store Preconditions v1 专题](../features/workflow/saved-workflow-draft-durable-store-preconditions-v1.md)
- [Saved Workflow Draft Repository Contract Preconditions v1 专题](../features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md)
- [Saved Workflow Draft Schema / Migration Preconditions v1 专题](../features/workflow/saved-workflow-draft-schema-migration-preconditions-v1.md)
- [Saved Workflow Draft Auth Context Preconditions v1 专题](../features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md)
- [Saved Workflow Draft Store Selector Enablement Preconditions v1 专题](../features/workflow/saved-workflow-draft-store-selector-enablement-preconditions-v1.md)

## 本轮交付

- 新增 store selector enablement preconditions 细专题，固定 precondition-only 状态。
- 新增 `workflow-saved-draft-store-selector-enablement-preconditions-v1` fixture / checker。
- checker 接入 fast baseline，校验 dependency、store mode contract、selector gate、failure mapping、no fallback、no side effects、dev flag boundary、文档引用和 forbidden implementation artifact。
- 同步更新 workflow 入口、Saved Draft 主专题、durable store / repository / schema / auth preconditions、User Workspace、当前焦点、任务卡索引、脚本说明和周志。

## Store Mode Contract

future selector 只允许以下 mode：

- `memory_dev`：当前默认，只用于 dev-only saved draft route。
- `repository_disabled`：reserved disabled，返回 `repository_store_disabled`。
- `repository`：future reserved，在 adapter / schema / auth / smoke gate 满足前返回 `repository_store_disabled`。
- `unknown`：fail closed，返回 `invalid_draft_store_mode`。

reserved mode 不得 fallback 到 memory dev store，也不得返回 sample / fixture 成功。

## Selector Gate

future selector 实现前必须固定：

- 正式 config key 和 mode allowlist。
- dev HTTP enablement 与 write enablement 仍由现有 dev flags 单独控制。
- repository adapter、schema / migration、auth context 和 selector smoke gate 均未满足时，`repository` mode fail closed。
- no sample fallback、no memory dev fallback、no side effects。

## 验收口径

- `workflow-saved-draft-store-selector-enablement-preconditions-v1` checker 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 文档和 fixture 均保持 precondition-only，不声明 store selector ready、repository mode ready、repository adapter ready 或 production ready。

## 停止线

- 不实现 durable persistence、正式 config entry、selector 函数、selector 类型、repository interface、repository adapter、SQL migration、真实数据库、Radish OIDC middleware、token validation、public production API、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把本任务卡、fixture 或 checker 解释为 store selector、repository adapter、saved draft list、publish、run 或 production readiness。
