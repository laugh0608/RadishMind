# Workflow Saved Draft Database Secret Resolver Runtime Dependency Refresh v1 任务卡

状态：`draft_database_secret_resolver_runtime_dependency_refresh_defined`

## 背景

Saved Workflow Draft durable store 已完成 database secret resolver readiness / implementation entry review、database connection provider entry refresh v2、schema marker runtime dependency refresh，以及 Production Secret Backend 的 real resolver、credential handle、operator approval、audit store、backend health 和 no leakage smoke runtime entry review。当前缺口集中在 secret resolver runtime 对这些生产依赖的组合消费上。

本任务卡只承接 database secret resolver runtime dependency refresh，用于把这些依赖统一成可检查的静态矩阵。本批不实现 resolver runtime，不创建 production resolver runtime、credential handle runtime、approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、DB provider、SQL、schema marker 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v2.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-runtime-dependency-refresh-v1.md`
- `docs/platform/production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.md`
- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-credential-handle-runtime-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-operator-approval-runtime-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md`
- `docs/platform/production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Secret Resolver Runtime Dependency Refresh v1` 细专题。
- 新增 `workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- dependency matrix 覆盖 saved draft secret resolver contract、secret reference manifest、production resolver runtime、credential handle、operator approval、audit store、backend health、no leakage smoke、test-only fake resolver、connection provider、schema marker、repository mode、auth membership、sanitized diagnostics 和 environment binding。
- runtime blockers 明确列出 database secret resolver runtime、production resolver runtime task card、credential handle runtime、approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、cloud secret service、connection provider、schema marker、repository mode 和 auth / membership runtime。
- failure mapping 固定 `draft_store_unavailable`、`draft_audit_context_missing`、`draft_store_migration_unavailable`、`draft_auth_context_contract_mismatch`、`repository_store_disabled` 和 `invalid_draft_store_mode`。
- no fallback、no side effects 和 artifact guard 能检测 resolver task card、production resolver runtime、credential handle、approval、audit store、backend health、no leakage smoke、DB provider、SQL、schema marker、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 database secret resolver implementation task card 或 production resolver runtime implementation task card。
- 不创建 secret resolver runtime、production resolver runtime、credential handle runtime、approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database connection provider、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
