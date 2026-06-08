# `Workflow Confirmation Placeholder Read` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-confirmation-placeholder-read-v1`，在 `apps/radishmind-web/` 的 `workspace-run-history` 区域下方增加 confirmation placeholder 的离线只读 surface。当前状态固定为 `workflow_confirmation_placeholder_read_defined`。

本切片只展示 required action ref、risk summary、required decision shape、human review requirement、disabled reason、route / request / audit metadata 和 missing prerequisites；不新增 Go route、真实 API、数据库、Radish OIDC、workflow executor、tool executor、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 production API consumer。

## 输入事实源

- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Blocked Action Preview` v1 计划](workflow-blocked-action-preview-v1-plan.md)
- [`Workflow Application Detail Read` v1 计划](workflow-application-detail-read-v1-plan.md)
- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [路线图](../radishmind-roadmap.md)

## 实施范围

1. 前端 view model
   - 新增 `apps/radishmind-web/src/features/control-plane-read/workflowConfirmationPlaceholder.ts`。
   - 复用 `WorkflowBlockedActionPreviewViewModel` / confirmation placeholder preview，派生独立 confirmation placeholder read view model。
   - 字段覆盖 required action ref、run ref、workflow definition ref、node execution ref、tool ref、action kind、risk level、risk summary、policy reason、required decision shape、decision fields、human review marker、disabled reason、prerequisites 和 audit metadata。
2. 前端 UI
   - 在 `App.tsx` 的 `workspace-run-history` 区域下方新增只读 confirmation placeholder panel。
   - 复用 `.workflow-*` / read-only card 风格，新增必要的 `workflow-confirmation-*` 样式。
   - 不添加按钮、表单、输入框、approve / reject / defer / submit、unlock、execute、writeback 或 replay 控件。
3. 治理证据
   - 新增 `scripts/checks/fixtures/workflow-confirmation-placeholder-read-v1.json`。
   - 新增 `scripts/checks/control_plane/check-workflow-confirmation-placeholder-read-v1.py`。
   - 接入 `scripts/check-repo.py` fast baseline。

## 验收口径

- `workflow-confirmation-placeholder-read-v1` 状态固定为 `workflow_confirmation_placeholder_read_defined`。
- 页面只能展示 future confirmation decision 的数据形状、人工复核要求、disabled reason、missing prerequisites 与 audit metadata。
- `canRequestLiveBackend`、`canMutate`、`canSubmitConfirmationDecision`、`canApproveDecision`、`canRejectDecision`、`canDeferDecision`、`canPersistDecision`、`canUnlockExecution`、`canExecuteTool`、`canWriteBusinessTruth` 和 `canReplayRun` 必须保持 `false`。
- checker 必须确认前序 `workflow-function-surface-boundary-v1`、`workflow-blocked-action-preview-v1` 与 `workflow-application-detail-read-v1` 状态未漂移。

## 非目标

- 不新增 runtime API、live detail backend 或 production API consumer。
- 不实现 confirmation submit / approve / reject / defer。
- 不持久化 confirmation decision，不实现 decision store。
- 不解锁执行，不实现 workflow executor、node executor、tool executor 或 agent loop。
- 不实现 business writeback、run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不实现真实数据库、repository adapter、schema migration、store selector、Radish OIDC、token validation 或 production auth。

## 验证

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-confirmation-placeholder-read-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-blocked-action-preview-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-application-detail-read-v1.py
npm run build
./scripts/check-repo.sh --fast
```
