# Workflow Saved Draft Repository Adapter Implementation Entry Review v1 任务卡

状态：`draft_repository_adapter_implementation_entry_review_defined`

## 背景

Saved Workflow Draft durable store 已完成 repository adapter implementation plan、formal store selector、schema artifact materialization、production auth readiness 和 adapter smoke readiness。进入 repository adapter implementation 前，需要先固定这些证据是否共同满足实现任务准入，并明确哪些能力仍必须留在后续独立批次中。

本任务卡只承接 implementation entry review，不创建 repository interface、adapter、database query、adapter smoke、OIDC middleware、token validation、membership adapter、production API 或执行链路。

## 输入

- `docs/features/workflow/saved-workflow-draft-repository-adapter-implementation-plan-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-implementation-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-materialization-v1.md`
- `docs/features/workflow/saved-workflow-draft-production-auth-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-adapter-smoke-readiness-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-repository-adapter-implementation-plan-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-materialization-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-adapter-smoke-readiness-v1.json`

## 输出

- 新增 `Saved Workflow Draft Repository Adapter Implementation Entry Review v1` 细专题。
- 新增 `workflow-saved-draft-repository-adapter-implementation-entry-review-v1` fixture / checker。
- 更新 workflow 入口、features 入口、Saved Workflow Draft v1、Workflow / Agent Runtime、current focus、task card index、scripts README 和周志。
- 将 checker 接入 `./scripts/check-repo.sh --fast`。

## 验收

- entry review 明确记录 repository interface、repository adapter 和 adapter unit tests 已满足后续 implementation task 的评审输入，但本批不创建任何实现 artifact。
- adapter smoke fixture / checker、repository mode enablement、production auth runtime、production API、SQL migration、migration runner、publish / run / executor / confirmation / writeback / replay 均保持关闭。
- failure mapping、no fallback 和 no side effects 与既有 saved draft durable store 证据一致。
- checker 能检测 future repository adapter / database / OIDC / production API artifact 是否被提前创建。

## 验证

本批完成后运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-repository-adapter-implementation-v1-plan.md`。
- 不创建 `services/platform/internal/httpapi/workflow_saved_draft_repository.go`、`workflow_saved_draft_repository_adapter.go`、adapter tests、database query、SQL migration、schema version table 或 migration runner。
- 不创建 OIDC middleware、token validation、membership adapter、production API、adapter smoke fixture 或 adapter smoke checker。
- 不启用 `repository` store mode。
