# Saved Workflow Draft PostgreSQL Dev/Test Repository v1

更新时间：2026-07-11

状态：`workflow_saved_draft_postgres_dev_test_repository_v1_in_progress`

## 功能目标

让 Radish 体系内部开发者保存的 Workflow 草案在 RadishMind 平台服务重启后仍可恢复，并继续沿用现有创建、编辑、校验、版本冲突、显式恢复和 Review Handoff 用户流程。

本专题只实现显式开发 / 测试态 PostgreSQL repository。它不启用 production `repository` mode，不把本地测试身份解释为 Radish OIDC，不打开 publish、run、executor、confirmation、业务写回或 replay。

## 用户结果

- 用户保存草案后重启平台服务，Workspace Home 仍能列出同一草案并恢复到 Draft Designer。
- 多个请求基于同一 `expected_draft_version` 保存时，只允许一个写入成功，其余请求稳定返回 `draft_version_conflict` 和当前版本。
- tenant、workspace、application、owner 任一不匹配时，调用方不能读取、列出或覆盖记录。
- 数据库、migration 或 schema marker 不可用时明确失败，不回退 `memory_dev`、sample 或 fixture。

## 模式与配置边界

新增独立 store mode：

```text
RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=postgres_dev_test
```

该模式还必须显式满足：

- `RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1`
- `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_HTTP=1`
- `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_WRITE=1`
- `RADISHMIND_WORKFLOW_SAVED_DRAFT_DEV_TEST_DATABASE_URL` 已通过进程环境提供

数据库连接超时使用 `RADISHMIND_WORKFLOW_SAVED_DRAFT_DATABASE_TIMEOUT`，默认 `5s`。数据库 URL 只允许从进程环境读取；配置摘要只输出 `configured / missing` 和字段来源，不输出 URL、用户名、密码、host、database name 或 query 参数。

`memory_dev` 继续作为默认开发模式。`repository_disabled` 与 production `repository` 继续失败关闭；`postgres_dev_test` 不得作为它们的别名，也不得隐式解锁任何生产声明。

## 请求与身份上下文

当前 `savedWorkflowDraftStore` 只接收 draft id 或 workspace / application，缺少 `context.Context`、tenant 和 owner。进入 PostgreSQL 前必须把以下信息贯穿 HTTP、domain service、store、repository 和 query executor：

- 请求 `context.Context`，用于取消、超时和连接回收。
- `request_id`、`audit_ref`。
- `tenant_ref`、`workspace_id`、`application_id`。
- `actor_subject_ref`、`owner_subject_ref`。
- `workflow_drafts:read` / `workflow_drafts:write` scope。

开发 / 测试模式只消费现有显式 dev auth headers；`tenant_ref` 来自 dev tenant binding，owner 默认等于已验证的 dev subject binding。Header 中的 workspace / application 必须与请求草案及 repository predicate 一致。

## Schema 与 migration

正式 migration source 位于：

```text
services/platform/migrations/workflow_saved_drafts/
```

本批物化：

- `workflow_saved_draft_schema_versions` marker，记录 component、migration id、store schema version、checksum 和 applied time。
- `saved_workflow_drafts` 表，保存 scope / owner / version / status / timestamp / actor / audit 列和 sanitized draft JSONB。
- `(tenant_ref, workspace_id, application_id, draft_id)` 唯一键。
- owner list 与 status list 索引。
- migration checksum、事务和 PostgreSQL advisory lock。

migration 通过独立一次性命令显式执行；平台服务启动只做连接和 schema preflight，不自动创建或升级 schema。重复执行相同 migration 必须幂等；marker、checksum、store schema 或物理表不一致时必须失败。

## Repository 行为

实现复用现有 `SavedWorkflowDraftRepository` 与 adapter，不绕过 domain validation：

- create：只允许 `expected_draft_version=0`，使用唯一键保证单次创建。
- update：在 SQL predicate 中同时比较完整 scope、owner 和 `draft_version`，不使用 Go 层读后写替代原子 CAS。
- read：只按完整 scope 与 owner 返回 sanitized record。
- list：只按 tenant / workspace / application / owner 查询，并保持 `updated_at DESC, draft_id ASC` 稳定顺序。
- query executor 只返回既有 failure code，不向 HTTP 响应、日志或测试产物透传 PostgreSQL 原始错误或连接信息。

平台 `Server` 持有连接池生命周期。启动 preflight 失败时服务不得悄悄切换 store；关闭服务时必须回收连接池。

## 验证策略

普通单元测试和 `check-repo --fast` 不依赖常驻数据库。真实 PostgreSQL 验证使用独立集成入口，并在 CI 中使用临时 PostgreSQL service：

1. 空测试数据库应用 migration，重复 apply 保持相同 marker 和 checksum。
2. 通过真实 HTTP route 保存、读取和列出草案。
3. 关闭第一个平台服务及连接池，再创建第二个服务，确认同一草案可恢复。
4. 并发提交相同 expected version，确认一个成功、其余冲突，最终版本只递增一次。
5. 覆盖 tenant / workspace / application / owner 隔离。
6. 覆盖 migration 未应用、marker mismatch、连接中断和 no fallback。
7. 用真实 Web consumer 复验保存、服务重启、恢复、二次保存、冲突、Continue / Restore 和 Review Handoff。

## 门禁调整

旧 Production Secret Backend / Audit Store / Storage Adapter readiness 脚本继续作为历史证据保留，但不再由仓库基线主动执行，也不再以“`pgx`、SQL 或 migration 文件必须不存在”约束本开发 / 测试专题。

本批不新增同类 Python checker。主要证据由 Go 单元测试、Go PostgreSQL 集成测试、Web 既有测试、浏览器记录和仓库聚合门禁承担。

## 完成条件

- 真实 migration、schema marker、连接池和 PostgreSQL query executor 已实现。
- `postgres_dev_test` 只有在完整显式配置下成功，production `repository` 仍失败关闭。
- 重启恢复、原子 CAS、scope / owner 隔离、数据库失败和 no fallback 均有真实 PostgreSQL 测试。
- Web 用户路径经过 PostgreSQL dev-live 浏览器复验，控制台无新增 error / warning。
- 功能专题、任务卡、当前焦点、架构、路线图、整改专题和周志口径一致。

## 停止线

- 不启用 production `repository` mode，不接真实 Radish OIDC、membership、production secret resolver、audit store 或公开生产 API。
- 不允许 DSN 或 PostgreSQL 原始错误进入 argv、日志、HTTP 响应、fixture、截图、trace 或 committed 文档 / 配置。
- 不自动执行 startup migration，不把 app runtime role 赋予 schema 变更权限。
- 不迁移 Control Plane Read、Gateway usage、run record 或其它领域数据。
- 不实现 publish、run、executor、unrestricted tool、confirmation decision、业务写回、replay、resume 或 materialized result reader。
