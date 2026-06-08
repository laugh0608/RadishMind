# `Workflow Runtime Readiness Inspector Offline` v1 计划

更新时间：2026-06-08

## 任务目标

本任务卡用于固定 `workflow-runtime-readiness-inspector-offline-v1` 的实现边界：在 `apps/radishmind-web/` 中复用 selected draft、offline validation inspector 和 execution plan preview，派生只读 runtime readiness inspector。

本切片只回答“未来进入 workflow runtime 前还缺哪些前置条件”，不实现真实 runtime、真实后端或生产接线。

## 输入事实源

- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- [`Workflow Execution Plan Preview Offline` v1 计划](workflow-execution-plan-preview-offline-v1-plan.md)
- [当前推进焦点](../radishmind-current-focus.md)
- [阶段路线图](../radishmind-roadmap.md)
- [跨项目集成契约](../radishmind-integration-contracts.md)

## 已完成范围

- 新增 `apps/radishmind-web/src/features/control-plane-read/workflowRuntimeReadinessInspector.ts`。
- 在 `App.tsx` 新增 `Runtime Readiness Inspector` 离线 panel。
- 展示 runtime prerequisite matrix、readiness blockers、implementation gates 和 route / request / audit metadata。
- readiness 覆盖 executor、provider binding、confirmation decision store、durable run/result store、audit policy、writeback policy、replay policy、auth, database, and repository gate 和 publish lifecycle gate。
- 新增 `scripts/checks/fixtures/workflow-runtime-readiness-inspector-offline-v1.json` 与 `scripts/checks/control_plane/check-workflow-runtime-readiness-inspector-offline-v1.py`。
- 已接入 `scripts/check-repo.py` fast baseline，状态固定为 `workflow_runtime_readiness_inspector_offline_defined`。

## 验收口径

- view model 必须从 offline execution plan preview 派生，不请求 live backend。
- 页面必须只展示 prerequisite、blocker、gate 和 audit metadata，不提供执行、发布、确认、写回、replay、数据库或 auth 启用控件。
- checker 必须确认 `canRequestLiveBackend`、`canStartWorkflowRuntime`、`canExecuteWorkflow`、`canSubmitConfirmationDecision`、`canWriteBusinessTruth`、`canReplayRun`、`canEnableProductionAuth`、`canAttachDatabase` 和 `canImplementRepositoryAdapter` 保持 `false`。
- 文档必须继续说明 read-side implementation trigger 当前没有 satisfied 条件。

## 非目标

- 不新增 Go route、真实 runtime API 或 production API consumer。
- 不实现 workflow executor、node executor、tool executor 或 agent loop。
- 不实现 confirmation decision、decision store、execution unlock 或确认提交。
- 不实现 durable run store、durable result store、materialized result reader、run replay 或 run resume。
- 不实现 business writeback、数据库、repository adapter、store selector、Radish OIDC、token validation 或 production auth。

## 停止线

- runtime readiness inspector 只能解释阻塞项，不能解锁任何能力。
- auth / database / repository gate 只能消费既有 readiness / trigger review 证据，不能创建实现 artifact。
- provider binding readiness 不能解释为 provider runtime 已接入 workflow executor。
- audit policy projection 不能解释为 durable audit store 或原始 payload 导出已完成。
