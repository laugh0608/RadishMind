# Saved Workflow Draft PostgreSQL DDL Review v1

状态：`postgresql_dev_test_migration_reviewed_and_executed`

## 审查范围

- up migration：`0001_saved_workflow_drafts.up.sql`、`0002_saved_workflow_draft_revisions.up.sql`
- down migration：`0002_saved_workflow_draft_revisions.down.sql`、`0001_saved_workflow_drafts.down.sql`
- marker：`workflow_saved_draft_schema_versions`
- store schema：`saved_workflow_drafts_store_v1`
- runner：`cmd/radishmind-workflow-draft-migrate`

## 审查结论

- up migration 创建 `saved_workflow_drafts`，以 `tenant_ref + workspace_id + application_id + draft_id` 为主键。
- `0002` 创建 `saved_workflow_draft_revisions`，以完整草案作用域和 `draft_version` 为主键，只允许 `saved / restored / backfilled_current` 三种来源。
- 升级既有数据库时只把当前主表快照回填为 `backfilled_current`，不根据版本号、时间或审计引用推测迁移前历史。
- 普通保存与显式恢复都在同一事务更新 current record 并插入 revision；revision 冲突或校验失败时 current record 不得提交。
- owner list、status list 与 schema compatibility 索引覆盖当前 read/list/preflight 路径。
- expected-version 更新在 SQL predicate 内完成，不能退回 Go 层读后写。
- apply 使用事务、checksum 和 PostgreSQL advisory lock；重复 apply 保持幂等。
- 服务启动只执行 marker / table preflight，不自动 apply migration。
- 本地与 CI 使用独立 migration role 和 runtime role；runtime role 没有 schema `CREATE` 权限，对 current 表保留既有 DML，对 revision 表只允许 `SELECT / INSERT` 并明确拒绝 `UPDATE / DELETE`。
- down migration 只允许 disposable dev/test 集成测试调用，日常 migration CLI 不暴露 `down`。

## 失败与脱敏

未应用、checksum mismatch、store schema mismatch、物理表缺失或连接不可用都失败关闭，不回退 memory、sample 或 fixture。数据库连接信息和原始 PostgreSQL 错误不得进入配置摘要、HTTP 响应、日志或运行证据；允许公开的数据库细节仅限 SQLSTATE。

## 生产停止线

本审查只批准显式 `postgres_dev_test`。production `repository` 仍需 Radish OIDC、membership、production secret resolver、正式数据库资源、审计、备份和部署复核。
