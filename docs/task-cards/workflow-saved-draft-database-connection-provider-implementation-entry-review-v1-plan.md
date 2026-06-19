# Workflow Saved Draft Database Connection Provider Implementation Entry Review v1 任务卡

状态：`draft_database_connection_provider_implementation_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已完成 repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner readiness、runner implementation entry review 和 database connection / schema marker preconditions。继续推进 durable store 时，不能直接创建 DB provider、SQL、runner 或 repository mode 成功路径，应先评审 database connection provider implementation 是否具备进入实现批次的条件。

本任务卡只承接 implementation entry review，不创建 database connection provider、secret resolver、DB driver、connection factory、role policy、connection smoke、database query executor、schema marker、SQL migration、migration runner、runner command、OIDC middleware、token validation、membership adapter、production API 或执行链路。

## 输入

- `docs/features/workflow/saved-workflow-draft-database-connection-schema-marker-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-database-connection-schema-marker-preconditions-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`

## 输出

- 新增 `Saved Workflow Draft Database Connection Provider Implementation Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-database-connection-provider-implementation-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- implementation entry review 明确记录 connection provider implementation 当前不打开。
- secret ref resolver、database driver policy、connection factory、runtime query role policy、connection smoke 和 repository query executor binding 均保持 blocked。
- failure mapping、no fallback 和 no side effects 与既有 saved draft durable store 证据一致。
- checker 能检测 DB provider / driver / secret resolver / SQL / marker / runner / production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-schema-marker-preconditions-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源，补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 database connection provider、secret resolver、DB driver、connection factory、role policy、connection smoke、database query executor、schema marker reader / writer、SQL migration、migration runner 或 runner command。
- 不启用 `repository` store mode。
- 不创建 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
