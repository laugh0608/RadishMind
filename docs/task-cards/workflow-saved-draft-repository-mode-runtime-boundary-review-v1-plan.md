# Workflow Saved Draft Repository Mode Runtime Boundary Review v1 任务卡

状态：`draft_repository_mode_runtime_boundary_review_defined`

## 背景

Saved Workflow Draft durable store 已经具备 repository adapter、adapter smoke execution 和 production auth runtime bridge；同时也已经完成 repository mode enablement、schema migration runner implementation entry review、database connection / schema marker preconditions、connection provider implementation entry review、database secret resolver readiness / implementation entry review，以及 production secret backend audit store runtime entry refresh v3。

这些证据说明 repository mode runtime 的若干静态输入已经可复验，但不表示 repository mode 可以启用。真实 schema migration runner、database connection provider、database secret resolver、schema marker、真实 query executor、OIDC middleware、token validation、membership adapter、production API consumer 和 production secret backend 仍未满足。

本任务卡只承接 runtime boundary review，不创建 repository mode runtime implementation task card，也不创建任何 runtime artifact。

## 输入

- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-secret-resolver-implementation-entry-review-v1.md`
- `docs/platform/production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.md`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`

## 输出

- 新增 `Saved Workflow Draft Repository Mode Runtime Boundary Review v1` 细专题。
- 新增 `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- runtime boundary review 明确记录 repository mode runtime 当前不打开。
- blocked conditions 覆盖 schema migration runner、database connection provider、database secret resolver、schema marker、real query executor、OIDC/token/membership、production API consumer、production secret backend 和 audit store runtime。
- 下一步依赖选择明确指向 migration runner / schema marker、connection provider、secret resolver、OIDC membership 或 production API readiness 中的独立方向。
- failure mapping 固定 `repository_store_disabled`、`invalid_draft_store_mode`、schema / auth / store / scope failure code，且失败不返回草案主体或敏感材料。
- no fallback、no side effects 和 artifact guard 能检测 repository mode runtime、DB provider / driver、SQL、schema marker、runner、secret resolver、OIDC、membership、production API、executor、confirmation、writeback 或 replay artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源和 `check-repo.py` 注册顺序，补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-repository-mode-runtime-implementation-v1-plan.md`。
- 不启用 `repository` store mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不创建 database connection provider、database secret resolver、fake database secret resolver、DB driver、connection factory、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command 或 dry-run output。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
