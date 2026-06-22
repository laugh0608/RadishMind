# Workflow Saved Draft Schema Marker Contract Implementation Entry Review v1 任务卡

状态：`draft_schema_marker_contract_implementation_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已完成 schema marker / migration runner readiness refresh。当前只具备 applied marker、manual runner、dry-run、idempotency / lock、duplicate handling 和 rollback observability 的静态前置；repository mode runtime task card 仍 blocked。

本任务卡只承接 schema marker contract implementation entry review，用于评审 marker reader / writer contract 是否可以创建实现任务卡。本批不实现 marker runtime，不创建 schema version table，不创建 SQL、runner、DB provider 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-schema-marker-migration-runner-readiness-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Schema Marker Contract Implementation Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- entry review 明确记录 schema marker contract implementation task card 当前不创建。
- candidate matrix 覆盖 marker reader contract、marker writer contract、marker table / record shape、repository preflight consumption binding、negative marker smoke 和 implementation task card。
- failure mapping 固定 `draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch`、`draft_store_migration_unavailable`、`draft_store_unavailable`、`repository_store_disabled` 和 `invalid_draft_store_mode`。
- no fallback、no side effects 和 artifact guard 能检测 schema marker task card、marker runtime、schema version table、SQL、runner、DB provider / driver、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-contract-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-migration-runner-readiness-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源和 `check-repo.py` 注册顺序，若时间允许补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 schema marker contract implementation task card。
- 不创建 database connection provider、database secret resolver、DB driver、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
