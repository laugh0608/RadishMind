# Workflow Saved Draft Repository Mode Enablement v1

更新时间：2026-06-18

## 任务定位

任务 ID：`workflow-saved-draft-repository-mode-enablement-v1`

当前状态：`draft_repository_mode_enablement_review_defined`

本任务是 repository mode enablement 的准入评审。它只判断当前证据链是否足以打开 `repository` store mode，并固定结论：当前不启用 repository mode，reserved repository mode 继续 fail closed。

## 输入

- `workflow-saved-draft-store-selector-smoke-v1`
- `workflow-saved-draft-schema-artifact-materialization-v1`
- `workflow-saved-draft-repository-adapter-implementation-v1`
- `workflow-saved-draft-adapter-smoke-v1`
- `workflow-saved-draft-production-auth-runtime-v1`

## 范围

- 新增 repository mode enablement feature doc、fixture 和 checker。
- 固定 runtime boundary、config gate、schema preflight、adapter smoke / production auth runtime dependency、failure mapping、rollback、no fallback 和 no side effects。
- 固定当前准入结论：`repository` / `repository_disabled` 仍返回 `repository_store_disabled`，unknown mode 仍返回 `invalid_draft_store_mode`。
- 将 checker 接入 `check-repo.py`，并同步 workflow 专题入口、Saved Draft 主专题、current focus、product scope、task card index、scripts README、platform README 和周志。

## 不在本次范围

- 不启用 `repository` store mode。
- 不创建数据库连接、SQL migration、schema version table、migration runner 或 real query executor。
- 不接 OIDC middleware、token validation、membership adapter 或 production API。
- 不创建 executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 adapter smoke fake executor 或 production auth runtime bridge 解释为 production ready。

## 验证

- `go test ./internal/httpapi`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-enablement-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-runtime-v1.py`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`
