# User Workspace 设计与开发文档

更新时间：2026-06-16

## 功能定位

`User Workspace` 是 RadishMind 面向终端用户和项目成员的主工作区。它用于查看和管理 AI 应用、Prompt 应用、Workflow、Agent / Copilot 应用、API key、调用量、运行记录、成本摘要和人工审查入口。

## 当前状态

- `apps/radishmind-web/` 已有 read-side shell、Workspace Home、applications、API keys、usage quota、workflow definitions、run history、Workflow Review Workspace 和 Workflow Review Handoff。
- Workspace Home 和 workflow definitions 已支持创建本地 workflow 草案并进入 Draft Designer；草案保存仍复用 dev-only saved draft consumer，不代表 production persistence。
- Workspace Home 已支持 dev-only saved draft list：显示当前 application 下已保存草案的 sanitized summary、empty / failure state、refresh 和 restore；恢复后进入 Draft Designer 继续审查和编辑，不代表 durable persistence。
- Draft Designer 已支持本地节点新增、移动、删除保护和边重建；validation inspector、execution plan preview 和 runtime readiness inspector 使用当前 active draft，不代表 workflow 可发布或可执行。
- `Workflow Review Handoff Active Draft v1` 已把恢复后的 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总为 Review Handoff 中的可交接审查记录，仍不保存、不导出、不发送 handoff。
- `Saved Workflow Draft Durable Store Preconditions v1` 已固定 durable store 迁移前置设计，明确 draft scope、owner / workspace 归属、version conflict、no sample fallback，以及 dev store 到未来 repository adapter 的切换停止线。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 repository contract preconditions，明确 future saved draft list 需要的 list operation 只能返回 sanitized summary / metadata，仍不代表 durable persistence 或 production API ready。
- `Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 `draft_schema_migration_preconditions_defined`，明确 future saved draft durable store 的 logical schema、index strategy、migration gate 和 failure mapping，仍不代表 database ready 或 migration ready。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 `draft_auth_context_preconditions_defined`，明确 future saved draft repository actor context 的身份来源、workspace membership、owner policy、scope grants 和 audit / sanitization 边界，仍不代表 Radish OIDC 或 production auth ready。
- `Saved Workflow Draft Store Selector Enablement Preconditions v1` 已固定 `draft_store_selector_enablement_preconditions_defined`，明确 future saved draft store mode、selector gate、failure mapping、no fallback 和 dev flag boundary，仍不代表 store selector ready 或 repository mode ready。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 `draft_schema_artifact_evidence_defined`，明确 future saved draft schema artifact manifest、DDL review、rollback evidence、migration smoke 和 artifact guard，仍不代表 schema artifact ready、migration ready 或 database ready。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 `draft_store_selector_smoke_readiness_defined`，明确 future saved draft selector smoke mode matrix、operation matrix、schema artifact failure、no fallback 和 no side effects，仍不代表 selector smoke ready 或 store selector ready。
- dev-only live read consumer 只能在显式 opt-in 下读取 fake-store-backed Go handlers。
- `ControlPlaneReadRepository` interface 已落地，七条 read handlers 已通过 fake-store repository bridge 消费数据。
- 当前仍不具备 production API consumer、真实数据库、Radish OIDC、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 设计边界

- 用户端默认只输出建议、解释、审查包和候选动作，不直接写业务真相源。
- 高风险动作必须保留 `requires_confirmation`。
- read-side 与未来 write/execution side 必须分开设计；展示 ready 不等于执行 ready。
- API key 页面不得展示 key value、hash、authorization header 或 secret material。

## 下一批开发方向

1. `User Workspace Saved Draft List v1`、`Workflow Draft Designer Editing Model v2`、`Workflow Draft Node Attribute Editing Model v1` 和 `Workflow Review Handoff Active Draft v1` 已落地；下一批若继续用户工作区方向，应选择新的明确用户审查增强点，不能把 handoff record 扩成保存、导出、发送或执行入口。
2. durable store 方向仍只能作为独立准入推进：repository contract smoke、repository adapter implementation plan、schema migration 或 auth contract 不能和用户工作区列表实现并行打开。
3. 在进入任何生产写入前，先补用户工作区功能设计更新，明确创建、保存、发布、执行、确认和回滚边界。
4. 若下一步只改展示、分组、文案或使用性，不新增专项 gate，复用 web build、consumer smoke 和仓库基线。
5. 若新增 API、写入、真实 auth、真实数据源或执行能力，必须新增 task card，并按风险补 fixture / checker。

## 验收方式

- 功能展示类：`npm run build`、必要浏览器布局检查、`./scripts/check-repo.sh --fast`。
- read contract 类：consumer smoke、Go handler tests、read-side contract checker。
- 写入或执行类：先补设计文档和 task card，再补单测、负向测试、仓库级检查和人工确认路径。
