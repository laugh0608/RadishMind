# `Workflow Draft Designer Offline` v1 计划

更新时间：2026-06-07

## 任务目标

本任务在 `apps/radishmind-web/` 中新增离线 workflow draft designer 产品面，承接现有 `workspace-workflow-definitions`、workflow definition detail 和 confirmation placeholder 数据，展示草案模板、节点、边、readiness、风险摘要、route / request / audit metadata 与 blocked capability preview。

状态固定为：`workflow_draft_designer_offline_defined`。

## 背景判断

- `workflow-definition-detail-read-v1`、`workflow-run-detail-read-v1`、`workflow-blocked-action-preview-v1`、`workflow-application-detail-read-v1` 与 `workflow-confirmation-placeholder-read-v1` 已完成离线只读产品骨架。
- `RadishFlow` / `Radish` 现阶段没有稳定 UI、command 或 API 挂载点；这不阻塞 RadishMind 继续推进平台本体产品功能。
- 当前仍不能进入真实 workflow builder、draft persistence、publish、executor、confirmation decision、business writeback、replay、数据库、Radish OIDC 或 production API consumer。

## 本次范围

- 新增 `workflowDraftDesigner.ts`，从现有 workflow definitions summary、definition detail 和 confirmation placeholder 派生 offline draft designer view model。
- 在 `App.tsx` 的 `workspace-workflow-definitions` 区域下方渲染 draft designer panel。
- 支持本地切换当前查看的 draft template；该 `local template switch` 只影响 React 本地状态，不持久化、不发请求。
- 展示 draft identity、provider profile、节点、边、readiness、风险摘要、blocked capability 和 route / request / audit metadata。
- 复用现有 `.workflow-*` 只读卡片风格，不添加保存、发布、执行、确认提交或写回控件。

## 验收口径

- `scripts/checks/fixtures/workflow-draft-designer-offline-v1.json` 固定 slice status、依赖、字段、模板策略、draft policy、停止线和验证策略。
- `scripts/checks/control_plane/check-workflow-draft-designer-offline-v1.py` 校验 fixture、源码、样式、文档引用和 `scripts/check-repo.py` fast baseline 接入。
- `apps/radishmind-web/` 能通过 `npm run build`。
- 本次只声明 offline draft designer surface，不声明 workflow builder、draft persistence、publish、runtime API、executor、confirmation flow、writeback、replay、数据库、OIDC 或 production readiness。

## 非目标

- 不新增 Go route、真实 API、repository、migration、数据库或 Radish OIDC。
- 不实现 workflow builder mutation、draft persistence、publish、workflow executor、node executor、tool executor 或 agent loop。
- 不提交 confirmation decision、不持久化 decision、不解锁执行、不写上层业务真相源。
- 不实现 run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部仓库。

## 停止线

- 本地 template switch 只能改变当前查看的 draft，不得写入 storage、URL、API 或外部项目。
- blocked capability preview 只能说明缺少什么前置条件，不能触发真实 action。
- 后续如需真实 builder、publish 或 executor，必须另开实现任务卡，并先满足数据库、auth、repository、confirmation、audit 和 runtime gate。
