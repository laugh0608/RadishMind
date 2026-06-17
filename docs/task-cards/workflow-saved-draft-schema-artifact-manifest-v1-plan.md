# Workflow Saved Draft Schema Artifact Manifest v1 任务卡

状态：`draft_schema_artifact_manifest_defined`

## 任务目标

在创建任何 saved workflow draft migration root、manifest 文件、DDL review、rollback evidence、migration smoke、SQL migration、migration runner、repository adapter、store selector、Radish OIDC middleware 或 production API 前，先固定 future schema artifact manifest 的可检查准入契约。

本任务卡承接：

- `workflow-saved-draft-schema-migration-preconditions-v1`
- `workflow-saved-draft-schema-artifact-evidence-v1`
- `workflow-saved-draft-store-selector-smoke-readiness-v1`
- `workflow-saved-draft-repository-contract-smoke-runner-implementation-v1`
- `workflow-saved-draft-repository-adapter-implementation-plan-v1`

## 产出

- `docs/features/workflow/saved-workflow-draft-schema-artifact-manifest-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-manifest-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py`
- `scripts/check-repo.py` fast baseline 注册
- current focus、workflow 入口、Saved Draft 主专题、脚本说明和周志同步

## 实现范围

- 固定 future manifest path：`services/platform/migrations/workflow_saved_drafts/manifest.json`。
- 固定 manifest version、schema artifact id、store schema version、schema version table 和 migration root 名称。
- 固定 manifest section matrix：identity、logical entity、field mapping、index mapping、migration plan、DDL review gate、rollback gate、migration smoke gate、operation predicates、failure mapping 和 no side effects。
- 固定 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord`、`ListWorkflowDraftRecords` 的 operation manifest matrix。
- 固定 missing manifest / DDL / rollback / migration smoke 的 fail-closed policy。
- 固定 no memory dev / sample / fixture / dev route fallback 和 no side effects。
- 固定禁止提前出现 migration root、SQL、migration runner、repository interface、repository adapter、store selector、OIDC 或 production API artifact。

## 不在范围

- 不创建 migration root、manifest 文件、DDL review artifact、rollback evidence artifact、migration smoke artifact、SQL migration、schema version table 或 migration runner。
- 不连接数据库，不引入 `database/sql`，不创建 repository interface 或 repository adapter。
- 不创建 store selector、selector smoke fixture、formal config entry、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

## 验收方式

- 运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-evidence-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py
./scripts/check-repo.sh --fast
```

- 若本批触及验证入口、阶段边界或文档真相源，再补跑：

```bash
./scripts/check-repo.sh
```

## 停止线

- `draft_schema_artifact_manifest_defined` 只表示 manifest 准入契约已固定。
- 它不代表 schema artifact file ready、DDL ready、migration ready、database ready、repository adapter ready、store selector ready、OIDC ready、production API ready 或 workflow executable ready。
