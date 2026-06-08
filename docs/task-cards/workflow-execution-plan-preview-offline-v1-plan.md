# `Workflow Execution Plan Preview Offline` v1 计划

更新时间：2026-06-08

## 任务目标

本任务在 `apps/radishmind-web/` 中新增 workflow execution plan preview，承接 selected draft、offline validation inspector、definition detail、run detail、blocked action preview 和 confirmation placeholder，展示只读 stage order、node-to-stage mapping、provider/profile requirements、confirmation / audit gates、blocked runtime / publish / writeback / replay reasons 和 route / request / audit metadata。

状态固定为：`workflow_execution_plan_preview_offline_defined`。

## 背景判断

- `workflow-draft-designer-offline-v1` 和 `workflow-draft-validation-inspector-offline-v1` 已能让用户查看草案与结构 / 契约校验结果。
- 下一步应解释“如果未来进入 runtime，这个草案会按什么阶段、哪些节点、哪些 gate 被组织”，但不能创建可执行计划。
- 当前仍不能进入真实 runtime API、execution plan persistence、workflow builder mutation、publish、executor、confirmation decision、business writeback、replay、数据库、Radish OIDC 或 production API consumer。

## 本次范围

- 新增 `workflowExecutionPlanPreview.ts`，从 selected draft 与 validation inspector 派生 offline execution plan preview view model。
- 在 `App.tsx` 的 draft validation inspector 下方渲染只读 execution plan preview panel。
- 展示 stage order、node-to-stage mapping、provider/profile requirements、confirmation/audit gates、blocked runtime / publish / writeback / replay reasons 和审计元数据。
- 复用现有 `.workflow-*` 只读卡片风格，不添加保存、发布、运行、确认提交、写回或 replay 控件。

## 验收口径

- `scripts/checks/fixtures/workflow-execution-plan-preview-offline-v1.json` 固定 slice status、依赖、字段、execution plan policy、停止线和验证策略。
- `scripts/checks/control_plane/check-workflow-execution-plan-preview-offline-v1.py` 校验 fixture、源码、样式、文档引用和 `scripts/check-repo.py` fast baseline 接入。
- `apps/radishmind-web/` 能通过 `npm run build`。
- 本次只声明 offline execution plan preview，不声明 execution plan persistence、workflow builder、publish、runtime API、executor、confirmation flow、writeback、replay、数据库、OIDC 或 production readiness。

## 非目标

- 不新增 Go route、真实 API、repository、migration、数据库或 Radish OIDC。
- 不实现 execution plan persistence、workflow builder mutation、draft persistence、publish、workflow executor、node executor、tool executor 或 agent loop。
- 不提交 confirmation decision、不持久化 decision、不解锁执行、不写上层业务真相源。
- 不实现 run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部仓库。

## 停止线

- execution plan preview 只能解释未来执行计划形状，不能创建、保存或提交真实 execution plan。
- provider/profile requirements 只能说明未来 runtime 依赖，不能触发 provider 调用、tool adapter 调用或 live health。
- confirmation / audit gates 只能展示 gating metadata，不能提交确认、写审计真相源或解锁执行。
- 后续如需真实 execution plan service、builder、publish 或 executor，必须另开实现任务卡，并先满足数据库、auth、repository、confirmation、audit 和 runtime gate。
