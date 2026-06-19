# Workflow Saved Draft Database Connection / Schema Marker Preconditions v1 任务卡

状态：`draft_database_connection_schema_marker_preconditions_defined`

## 背景

Saved Workflow Draft durable store 已完成 repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement、schema migration runner readiness 和 runner implementation entry review。继续推进 durable store 时，下一步不应直接写 SQL、runner 或真实 repository mode，而应先固定 database connection provider 和 schema marker contract 的实现前置条件。

本任务卡只承接 connection / marker preconditions，不创建数据库连接、secret resolver、SQL migration、schema version table、schema marker 读写、database query executor、migration runner、runner command、OIDC middleware、token validation、membership adapter、production API 或执行链路。

## 输入

- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-implementation-entry-review-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json`

## 输出

- 新增 `Saved Workflow Draft Database Connection / Schema Marker Preconditions v1` 细专题。
- 新增 `workflow-saved-draft-database-connection-schema-marker-preconditions-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- 明确 database connection provider、secret ref、query role、environment isolation 和 schema marker read / write contract 的 future 前置要求。
- 明确本批不创建 DB connection、secret resolver、SQL、schema marker、runner、query executor 或 repository mode 成功路径。
- failure mapping、no fallback 和 no side effects 与既有 saved draft durable store 证据一致。
- checker 能检测 SQL / DB / marker / runner / production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-schema-marker-preconditions-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py
./scripts/check-repo.sh --fast
```

由于本批同步阶段真相源，补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-schema-marker-implementation-v1-plan.md`。
- 不创建 database connection、secret resolver、schema marker reader / writer、database query executor、SQL migration、schema version table、migration runner 或 runner command。
- 不启用 `repository` store mode。
- 不创建 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
