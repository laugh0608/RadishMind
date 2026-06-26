# Workflow Saved Draft Database Connection Smoke Strategy v1 任务卡

状态：`draft_database_connection_smoke_strategy_defined`

## 背景

Saved Workflow Draft database driver / DSN / TLS policy readiness 与 database role policy readiness 已完成，connection provider implementation 仍缺少可复验的 connection smoke 策略。本任务卡只定义 explicit test database、safe placeholder credential handoff、smoke input / output record shape、role denial cases、no leakage scan 和 zero production side effect 记录，不实现 smoke runner，不连接数据库，不创建 DB provider、connection factory、SQL、schema marker 或 repository mode runtime。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-driver-dsn-tls-policy-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-role-policy-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-refresh-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-runtime-boundary-review-v1.md`
- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`

## 输出

- 新增 `Saved Workflow Draft Database Connection Smoke Strategy v1` 细专题。
- 新增 `workflow-saved-draft-database-connection-smoke-strategy-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- strategy 明确 future explicit test database boundary、safe placeholder credential handoff、smoke input / output record shape、role denial cases、no leakage scan、manual-only execution boundary 和 zero production side effect。
- fixture 明确 `draft_database_connection_smoke_strategy_defined` 不是 smoke executed、connection provider ready、database ready、repository mode ready 或 production API ready。
- no fallback、no side effects 和 artifact guard 能检测 connection smoke runner、smoke runtime、DB provider、driver import、DSN parser、connection factory、SQL、schema marker、migration runner、repository mode、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。
- checker 接入 `scripts/check-repo.py`，并位于 database role policy readiness 之后、product surface recheck 之前。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 connection provider implementation task card。
- 不创建 connection smoke runner、connection smoke runtime、database connection provider、DB driver import、DSN parser runtime、TLS runtime、connection factory、role policy runtime、no leakage scan runtime、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command、dry-run output 或 smoke output。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
