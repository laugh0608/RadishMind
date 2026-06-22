# Workflow Saved Draft Database Role Policy Readiness v1 任务卡

状态：`draft_database_role_policy_readiness_defined`

## 背景

Saved Workflow Draft database driver / DSN / TLS policy readiness 已确认 role policy 仍是独立前置缺口。本任务卡只把 runtime DML role、migration DDL / schema marker role、least privilege review 和 cross-environment denial smoke 前置收束为可检查 readiness，不实现 role policy runtime，不创建数据库 role、role grant、DB provider、connection smoke、SQL、schema marker、migration runner 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-driver-dsn-tls-policy-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-manual-migration-runner-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-marker-contract-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Role Policy Readiness v1` 细专题。
- 新增 `workflow-saved-draft-database-role-policy-readiness-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- readiness 明确 future runtime DML role、migration DDL / schema marker role、least privilege review、cross-environment denial smoke 前置、environment binding 和 role claim / role id metadata-only shape。
- fixture 明确 `draft_database_role_policy_readiness_defined` 不是 role runtime ready、database role ready、role grant ready、connection provider ready、repository mode ready 或 production API ready。
- no fallback、no side effects 和 artifact guard 能检测 role policy runtime、database role grant、provider task card、provider runtime、driver import、DSN parser、connection factory、connection smoke、SQL、schema marker、migration runner、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。
- checker 接入 `scripts/check-repo.py`，并位于 database driver / DSN / TLS policy readiness 之后、product surface recheck 之前。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-radish-oidc-token-membership-implementation-entry-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 connection provider implementation task card。
- 不创建 role policy runtime、database role、role grant、database connection provider、DB driver import、DSN parser runtime、TLS runtime、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
