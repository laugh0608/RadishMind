# `Workflow Draft Validation Inspector Offline` v1 计划

更新时间：2026-06-07

## 任务目标

本任务在 `apps/radishmind-web/` 中新增 workflow draft 的 offline validation inspector，承接 `workflow-draft-designer-offline-v1` 已派生的草案视图，展示 draft identity、summary、structural checks、input / output contract checks、blocked capability checks 和 route / request / audit metadata。

状态固定为：`workflow_draft_validation_inspector_offline_defined`。

## 背景判断

- `workflow-draft-designer-offline-v1` 已完成离线草案设计面，但仍只允许本地查看和切换草案。
- `RadishFlow` / `Radish` 现阶段没有稳定 UI、command 或 API 挂载点；这不阻塞 RadishMind 继续推进成熟的离线产品功能。
- 当前仍不能进入真实 workflow builder、draft persistence、validation result persistence、publish、executor、confirmation decision、business writeback、replay、数据库、Radish OIDC 或 production API consumer。

## 本次范围

- 新增 `workflowDraftValidationInspector.ts`，从当前 selected draft 派生 offline validation inspector view model。
- 在 `App.tsx` 的 draft designer 下方渲染只读 validation inspector panel。
- 展示 validation summary、structural checks、contract checks、blocked capability checks 和审计元数据。
- 复用现有 `.workflow-*` 只读卡片风格，不添加保存、发布、执行、确认提交、写回或校验结果持久化控件。

## 验收口径

- `scripts/checks/fixtures/workflow-draft-validation-inspector-offline-v1.json` 固定 slice status、依赖、字段、validation policy、contract policy、停止线和验证策略。
- `scripts/checks/control_plane/check-workflow-draft-validation-inspector-offline-v1.py` 校验 fixture、源码、样式、文档引用和 `scripts/check-repo.py` fast baseline 接入。
- `apps/radishmind-web/` 能通过 `npm run build`。
- 本次只声明 offline validation inspector，不声明 workflow builder、draft persistence、validation result persistence、publish、runtime API、executor、confirmation flow、writeback、replay、数据库、OIDC 或 production readiness。

## 非目标

- 不新增 Go route、真实 API、repository、migration、数据库或 Radish OIDC。
- 不实现 workflow builder mutation、draft persistence、validation result persistence、publish、workflow executor、node executor、tool executor 或 agent loop。
- 不提交 confirmation decision、不持久化 decision、不解锁执行、不写上层业务真相源。
- 不实现 run replay、run resume、durable run store、durable result store 或 materialized result reader。
- 不修改 `RadishFlow`、`Radish` 或 `RadishCatalyst` 外部仓库。

## 停止线

- offline validation inspector 只能解释当前 fixture 草案的结构和契约状态，不得写入 storage、URL、API 或外部项目。
- blocked capability checks 只能说明缺少什么前置条件，不能触发真实 action。
- 后续如需真实 validation service、builder、publish 或 executor，必须另开实现任务卡，并先满足数据库、auth、repository、confirmation、audit 和 runtime gate。
