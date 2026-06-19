# Workflow Saved Draft Schema Migration Runner Readiness v1

更新时间：2026-06-18

## 任务定位

任务 ID：`workflow-saved-draft-schema-migration-runner-readiness-v1`

当前状态：`draft_schema_migration_runner_readiness_defined`

本任务定义 saved workflow draft future schema migration runner 的实现准入边界。它承接 repository mode enablement 的评审结论，固定当前仍不能创建 runner、SQL、数据库连接或 repository mode runtime。

## 输入

- `workflow-saved-draft-schema-migration-preconditions-v1`
- `workflow-saved-draft-schema-artifact-materialization-v1`
- `workflow-saved-draft-repository-adapter-implementation-v1`
- `workflow-saved-draft-adapter-smoke-v1`
- `workflow-saved-draft-production-auth-runtime-v1`
- `workflow-saved-draft-repository-mode-enablement-v1`

## 范围

- 新增 schema migration runner readiness feature doc、fixture 和 checker。
- 固定 future runner 的 runtime boundary、manual config gate、schema preflight、adapter smoke / production auth runtime dependency、failure mapping、rollback、no fallback 和 no side effects。
- 确认当前只允许静态 schema artifact 存在，`services/platform/migrations/workflow_saved_drafts/` 下不得出现 `.sql`。
- 将 checker 接入 `check-repo.py`，并同步 workflow 专题入口、Saved Draft 主专题、current focus、product scope、capability matrix、task card index、scripts README、platform README 和周志。

## 不在本次范围

- 不创建 SQL migration、schema version table、migration runner、runner command、database query executor 或数据库连接。
- 不启用 `repository` store mode。
- 不接 OIDC middleware、token validation、membership adapter 或 production API。
- 不创建 executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 static schema artifact、adapter smoke fake executor 或 production auth runtime bridge 解释为 production ready。

## 验证

- `go test ./internal/httpapi`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-migration-runner-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`
