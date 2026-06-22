# Workflow Saved Draft Database Connection Provider Implementation Entry Refresh v1 任务卡

状态：`draft_database_connection_provider_implementation_entry_refresh_defined`

## 背景

Saved Workflow Draft manual migration runner 已完成 implementation entry refresh，确认 runner task card 仍依赖 database connection provider、database role、marker writer、lock / idempotency 和 rollback observability。此前 connection provider entry review 停留在 production secret backend resolver 未实现的旧状态；当前 test-only fake resolver runtime 已实现，但真实 resolver runtime、credential handle runtime 和 no leakage smoke runtime 仍 blocked。

本任务卡只刷新 connection provider implementation entry，不实现 provider，不创建 driver、DSN parser、connection factory、connection smoke、query executor、schema marker、SQL migration、runner 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-manual-migration-runner-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md`
- `docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md`
- `docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Connection Provider Implementation Entry Refresh v1` 细专题。
- 新增 `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- entry refresh 明确记录 connection provider implementation task card 当前不创建。
- candidate matrix 覆盖 secret resolver handoff、driver / DSN / TLS policy、connection lifecycle、runtime / migration role policy、offline connection smoke、repository query executor handoff 和 provider task card。
- refresh 明确 test-only fake resolver runtime 不能解锁 production resolver、credential handle、no leakage smoke、DB provider、repository mode 或 production API。
- no fallback、no side effects 和 artifact guard 能检测 provider task card、provider runtime、driver、DSN parser、connection factory、connection smoke、SQL、schema marker、migration runner、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-manual-migration-runner-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源和 `check-repo.py` 注册顺序，若时间允许补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 connection provider implementation task card。
- 不创建 database connection provider、secret resolver、DB driver、DSN parser、connection factory、role policy runtime、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
