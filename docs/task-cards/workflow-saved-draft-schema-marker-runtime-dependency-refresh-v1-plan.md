# Workflow Saved Draft Schema Marker Runtime Dependency Refresh v1 任务卡

状态：`draft_schema_marker_runtime_dependency_refresh_defined`

## 背景

Saved Workflow Draft durable store 已完成 schema marker contract entry review、manual migration runner entry refresh、database connection provider entry refresh v2 和 Radish OIDC upstream evidence refresh。当前缺口不再只是单一 marker contract，而是 marker runtime 对 runner、connection provider、repository mode、auth / membership、lifecycle / smoke 和 rollback evidence 的组合依赖。

本任务卡只承接 schema marker runtime dependency refresh，用于把这些依赖统一成可检查的静态矩阵。本批不实现 marker runtime，不创建 schema version table，不创建 SQL、runner、DB provider、repository mode runtime、OIDC middleware 或 membership adapter。

## 输入

- `docs/features/workflow/saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-manual-migration-runner-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v2.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-lifecycle-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-smoke-strategy-v1.md`
- `docs/integrations/radish-oidc-token-membership-upstream-evidence-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Schema Marker Runtime Dependency Refresh v1` 细专题。
- 新增 `workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- dependency matrix 覆盖 marker contract、applied marker source、read path、write path、manual runner、connection provider、connection lifecycle、connection smoke、repository mode、auth membership、idempotency / lock 和 rollback observability。
- runtime blockers 明确列出 schema marker reader / writer、schema version table、manual runner、connection provider、repository mode runtime、token validation schema、auth middleware、membership adapter 和 connection smoke runtime。
- failure mapping 固定 `draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch`、`draft_store_migration_unavailable`、`draft_store_unavailable`、`draft_audit_context_missing`、`repository_store_disabled` 和 `invalid_draft_store_mode`。
- no fallback、no side effects 和 artifact guard 能检测 schema marker implementation task card、marker runtime、schema version table、SQL、runner、DB provider / driver、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-upstream-evidence-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 schema marker implementation task card 或 schema marker contract implementation task card。
- 不创建 database connection provider、database secret resolver、DB driver、connection factory、connection lifecycle runtime、connection smoke runner、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
