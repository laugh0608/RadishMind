# Workflow Saved Draft Database Secret Resolver Implementation Entry Review v1 任务卡

状态：`draft_database_secret_resolver_implementation_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已推进到 database secret resolver readiness。上一批只定义了 secret ref key、resolver input / result shape、sanitized diagnostics、failure taxonomy、disabled runtime 和 offline fake resolver strategy，不代表 resolver implementation、fake resolver、production secret backend、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode 或 production API ready。

本任务卡只承接 implementation entry review，评审 secret resolver implementation 是否打开。当前结论为不打开 implementation entry，不创建 resolver implementation task card，也不创建任何 runtime artifact。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`

## 输出

- 新增 `Saved Workflow Draft Database Secret Resolver Implementation Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- implementation entry review 明确记录 secret resolver implementation 当前不打开。
- blocked reason 覆盖 production secret backend resolver not started、reference-only secret manifest、fake resolver task 未创建、sanitized diagnostics runtime 未实现、connection provider entry blocked 和 repository mode disabled。
- failure mapping 固定 `draft_store_unavailable`、`repository_store_disabled` 和 `invalid_draft_store_mode`，且 diagnostics 只允许脱敏状态。
- no fallback、no side effects 和 artifact guard 能检测 resolver、fake resolver、DB provider / driver、SQL、schema marker、migration runner、repository mode、OIDC 和 production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源，补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-secret-resolver-implementation-v1-plan.md`。
- 不创建 secret resolver implementation、fake resolver、database connection provider、DB driver、connection factory、role policy、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner 或 runner command。
- 不启用 `repository` store mode。
- 不创建 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
