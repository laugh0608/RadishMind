# `Workflow Blocked Action Preview` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-blocked-action-preview-v1`，把 `Workflow / Agent Runtime Function Surface v1` 中的 tool action preview 从 definition / run detail 的局部字段推进到 `apps/radishmind-web/` 内独立的离线只读 blocked action surface。当前状态固定为 `workflow_blocked_action_preview_defined`。

本切片只解释 tool/action 为什么被阻止、当前缺哪些前置条件、未来需要什么 confirmation placeholder 和 audit trail；不新增 live backend、runtime API、workflow executor、tool executor、confirmation decision、execution unlock、business writeback、run replay、run resume、durable store、真实数据库、Radish OIDC 或 production API consumer。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [阶段路线图](../radishmind-roadmap.md)
- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Function Surface Boundary` v1 计划](workflow-function-surface-boundary-v1-plan.md)
- [`Workflow Definition Detail Read` v1 计划](workflow-definition-detail-read-v1-plan.md)
- [`Workflow Run Detail Read` v1 计划](workflow-run-detail-read-v1-plan.md)

## 实现范围

1. 在 `apps/radishmind-web/src/features/control-plane-read/workflowBlockedActionPreview.ts` 中新增离线只读 blocked action preview view model。
2. 复用 `WorkflowDefinitionBlockedActionPreview` 和 `WorkflowRunDetailGuardPreview`，默认展示 `tool_action_preview_reconnect_stream`。
3. 在 `App.tsx` 的 `workspace-run-history` 区域下方新增 blocked action panel。
4. 展示 `missingPrerequisites`、`confirmationPlaceholder`、`auditTrail`、`policyReason`、`blockedState`、route / request / audit metadata 和 related run guard。
5. 样式复用现有 `.workflow-*` 卡片风格，只补必要 class，不添加按钮、表单、执行控件、确认提交控件或写操作入口。

## 程序化证据

- `scripts/checks/fixtures/workflow-blocked-action-preview-v1.json`
- `scripts/checks/control_plane/check-workflow-blocked-action-preview-v1.py`
- `scripts/check-repo.py` fast baseline

## 验收口径

- `workflow-blocked-action-preview-v1` 状态固定为 `workflow_blocked_action_preview_defined`。
- detail surface 只从已有 definition detail / run detail 的离线 view model 派生，不请求 live backend。
- 页面只能解释 blocked action、missing prerequisites、confirmation placeholder 和 audit trail。
- `canRequestLiveBackend`、`canMutate`、`canExecuteTool`、`canSubmitConfirmationDecision`、`canUnlockExecution`、`canWriteBusinessTruth` 和 `canReplayRun` 必须保持 `false`。

## 停止线

- 不新增 Go route、runtime API 或 dev-live detail route。
- 不实现 workflow executor、node executor、tool executor 或 agent loop。
- 不执行 tool，不提交 confirmation decision，不解锁执行。
- 不实现 business writeback、run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不实现真实数据库、Radish OIDC、token validation 或 production API consumer。
