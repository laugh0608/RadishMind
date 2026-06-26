# Workflow Saved Draft Schema Migration Runner Implementation Entry Review v1 任务卡

状态：`draft_schema_migration_runner_implementation_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已完成 repository adapter implementation、adapter smoke execution、production auth runtime bridge、repository mode enablement 准入评审和 schema migration runner readiness。继续推进 durable store 时，下一步不能直接写 SQL 或 runner，而应先评审 schema migration runner implementation 是否具备进入实现任务的条件。

本任务卡只承接 implementation entry review，不创建 SQL migration、schema version table、migration runner、runner command、database query executor、数据库连接、OIDC middleware、token validation、membership adapter、production API 或执行链路。

## 输入

- `docs/features/workflow/saved-workflow-draft-schema-migration-runner-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-mode-enablement-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-adapter-smoke-execution-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-runtime-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-schema-migration-runner-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-mode-enablement-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-production-auth-runtime-v1.json`

## 输出

- 新增 `Saved Workflow Draft Schema Migration Runner Implementation Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、产品范围、能力矩阵、task card index、scripts README、services/platform README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- entry review 明确记录 migration runner implementation 当前不打开。
- executable migration artifact、schema version marker、manual runner command、dry-run plan、apply smoke、rollback observability 和 repository mode runtime enablement 均保持 blocked。
- failure mapping、no fallback 和 no side effects 与既有 saved draft durable store 证据一致。
- checker 能检测 SQL / runner / DB / production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py
./scripts/check-repo.sh --fast
```

若本批同步了阶段真相源并且时间允许，补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-schema-migration-runner-implementation-v1-plan.md`。
- 不创建 `services/platform/internal/httpapi/workflow_saved_draft_schema_migration_runner.go`、runner command、database query executor、数据库连接、SQL migration 或 schema version table。
- 不启用 `repository` store mode。
- 不创建 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
