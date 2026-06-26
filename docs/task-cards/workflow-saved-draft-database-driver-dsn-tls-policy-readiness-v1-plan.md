# Workflow Saved Draft Database Driver / DSN / TLS Policy Readiness v1 任务卡

状态：`draft_database_driver_dsn_tls_policy_readiness_defined`

## 背景

Saved Workflow Draft database connection provider implementation entry refresh 已确认 provider task card 仍 blocked，其中 driver / DSN / TLS policy 是独立缺口。本任务卡只把该缺口拆成可检查 readiness，不实现 provider，不选择 driver，不创建 DSN parser、TLS runtime、connection factory、connection smoke、SQL、schema marker 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-manual-migration-runner-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Driver / DSN / TLS Policy Readiness v1` 细专题。
- 新增 `workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- readiness 明确 future DB driver selection、DSN construction / redaction boundary、TLS policy、environment binding、forbidden diagnostics scan、role policy dependency 和 connection smoke 前置。
- fixture 明确 `draft_database_driver_dsn_tls_policy_readiness_defined` 不是 driver selected、driver imported、DSN parser ready、TLS runtime ready、connection provider ready 或 repository mode ready。
- no fallback、no side effects 和 artifact guard 能检测 provider task card、provider runtime、driver import、DSN parser、TLS runtime、connection factory、connection smoke、SQL、schema marker、migration runner、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。
- checker 接入 `scripts/check-repo.py`，并位于 database connection provider entry refresh 之后、product surface recheck 之前。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 connection provider implementation task card。
- 不创建 database connection provider、DB driver import、DSN parser runtime、TLS runtime、connection factory、role policy runtime、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
