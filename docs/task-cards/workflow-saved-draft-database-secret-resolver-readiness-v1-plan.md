# Workflow Saved Draft Database Secret Resolver Readiness v1 任务卡

状态：`draft_database_secret_resolver_readiness_defined`

## 背景

Saved Workflow Draft durable store 已完成 repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner readiness、runner implementation entry review、database connection / schema marker preconditions 和 database connection provider implementation entry review。继续推进 database connection provider 前，必须先固定 saved draft database secret resolver 的 readiness 边界，避免把 production secret backend 的 reference-only 状态误写成 resolver runtime 已可用。

本任务卡只承接 readiness，不创建 secret resolver implementation、fake resolver、DB driver、database connection provider、connection factory、role policy、connection smoke、database query executor、schema marker、SQL migration、migration runner、OIDC middleware、token validation、membership adapter、production API 或执行链路。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-connection-provider-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-database-connection-schema-marker-preconditions-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`

## 输出

- 新增 `Saved Workflow Draft Database Secret Resolver Readiness v1` 细专题。
- 新增 `workflow-saved-draft-database-secret-resolver-readiness-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- readiness 明确记录 secret resolver implementation 当前不打开。
- secret ref key、resolver input / result shape、sanitized diagnostics、environment binding、failure taxonomy 和 offline fake resolver strategy 已固定。
- production secret backend 仍保持 `not_started`，`production-secret-reference-basic` 仍保持 reference-only、resolver disabled 和 no cloud calls。
- checker 能检测 secret resolver、DB provider / driver、SQL、marker、runner、repository mode、OIDC 和 production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
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
