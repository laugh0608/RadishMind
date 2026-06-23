# Workflow Saved Draft Database Connection Lifecycle Readiness v1 任务卡

状态：`draft_database_connection_lifecycle_readiness_defined`

## 背景

Saved Workflow Draft database connection smoke strategy 已完成，但 connection provider implementation 仍缺少可复验的 lifecycle 前置。本任务卡只定义 timeout、pool、health check、close responsibility、request id / audit ref propagation 和 sanitized diagnostics runtime 前置，不实现 connection lifecycle runtime，不创建 connection factory，不连接数据库，不创建 DB provider、SQL、schema marker 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-connection-smoke-strategy-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-driver-dsn-tls-policy-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-role-policy-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/integrations/radish-oidc-token-membership-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Connection Lifecycle Readiness v1` 细专题。
- 新增 `workflow-saved-draft-database-connection-lifecycle-readiness-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、能力矩阵、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- readiness 明确 future timeout budget、pool policy、health check boundary、close responsibility、request id / audit ref propagation 和 sanitized diagnostics runtime 前置。
- fixture 明确 `draft_database_connection_lifecycle_readiness_defined` 不是 lifecycle runtime ready、connection provider ready、database ready、repository mode ready 或 production API ready。
- no fallback、no side effects 和 artifact guard 能检测 connection lifecycle runtime、connection factory、DB provider、driver import、DSN parser、pool runtime、health check runtime、connection smoke runtime、SQL、schema marker、migration runner、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。
- checker 接入 `scripts/check-repo.py`，并位于 database connection smoke strategy 之后、product surface recheck 之前。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-lifecycle-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 connection provider implementation task card。
- 不创建 connection lifecycle runtime、connection factory、database connection provider、DB driver import、DSN parser runtime、pool runtime、health check runtime、connection smoke runtime、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command、dry-run output 或 smoke output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
