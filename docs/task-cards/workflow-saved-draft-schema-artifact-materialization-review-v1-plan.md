# Workflow Saved Draft Schema Artifact Materialization Review v1 任务卡

## 任务标识

- 切片：`workflow-saved-draft-schema-artifact-materialization-review-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`draft_schema_artifact_materialization_review_defined`

## 目标

在创建 saved workflow draft schema artifact 文件前，评审是否打开 migration root、manifest file、DDL review artifact、rollback evidence artifact 和 migration smoke artifact。本任务卡只固定 materialization review 结论、候选项、阻塞条件、failure mapping、no fallback、no side effects 和 forbidden artifact guard。

当前结论：本批不打开 schema artifact materialization entry，不创建后续 materialization implementation task card，不写 migration、SQL、schema artifact 或 runtime migration runner。

## 范围

- 新增 schema artifact materialization review fixture 和 checker。
- 消费 `workflow-saved-draft-schema-artifact-evidence-v1`、`workflow-saved-draft-schema-artifact-manifest-v1`、`workflow-saved-draft-repository-adapter-implementation-plan-v1`、`workflow-saved-draft-adapter-smoke-readiness-v1` 和 `workflow-saved-draft-store-selector-implementation-entry-review-v1`。
- 固定五个候选项：migration root、manifest file、DDL review artifact、rollback evidence artifact、migration smoke artifact。
- 固定本批结论：五个候选项均 `blocked`，后续只能在单独 implementation task card 中打开。
- 将 checker 接入仓库 fast baseline，并同步 workflow 专题入口、Saved Workflow Draft v1、current focus、能力矩阵、脚本说明、任务卡入口和周志。

## 不在本次范围

- 不创建 `docs/task-cards/workflow-saved-draft-schema-artifact-materialization-v1-plan.md`。
- 不创建 `services/platform/migrations/workflow_saved_drafts`、`manifest.json`、`ddl-review.md`、`rollback-evidence.json` 或 `migration-smoke.json`。
- 不创建 SQL migration、schema version table、migration runner、database connection 或 database query。
- 不实现 repository interface、repository adapter、store selector、selector smoke fixture、adapter smoke fixture 或 adapter smoke checker。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle、quota enforcement 或 billing。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何执行能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-review-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-manifest-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-evidence-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-implementation-entry-review-v1.py`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`
