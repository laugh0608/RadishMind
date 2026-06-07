# `Workflow Function Surface Boundary` v1 计划

更新时间：2026-06-07

## 任务目标

本任务卡用于落实 `workflow-function-surface-boundary-v1`，把 `Workflow / Agent Runtime Function Surface v1` 的第一步从方向性计划推进到可检查边界。当前状态固定为 `function_surface_boundary_defined`。

本切片只定义 application detail、workflow definition detail、run detail、tool action preview 和 confirmation placeholder 的字段边界、只读状态、blocked 状态和停止线；不实现新的 runtime API、workflow builder、workflow executor、tool executor、confirmation decision、business writeback、replay、durable store、真实数据库、Radish OIDC 或 production API consumer。

## 输入事实源

- [当前推进焦点](../radishmind-current-focus.md)
- [能力矩阵](../radishmind-capability-matrix.md)
- [阶段路线图](../radishmind-roadmap.md)
- [`Workflow / Agent Runtime Function Surface` v1 计划](workflow-agent-runtime-function-surface-v1-plan.md)
- `scripts/checks/fixtures/workflow-definition-run-record-boundary.json`
- `scripts/checks/fixtures/control-plane-read-formal-ui-readiness-close-v1.json`
- `scripts/checks/fixtures/control-plane-read-implementation-trigger-review-v1.json`

## 字段边界

1. `application detail`
   - 展示 application identity、tenant / owner ref、application type、状态、provider profile、risk summary、latest run ref、route ref 和 audit ref。
   - 不提供创建、编辑、删除、发布或执行入口。
2. `workflow definition detail`
   - 展示 workflow definition id、application ref、version、definition status、nodes、edges、input / output summary、risk level、requires confirmation capability、route ref 和 audit ref。
   - 不提供 builder、节点编辑、版本发布、definition mutation 或 executor。
3. `run detail`
   - 展示 run id、definition ref、application ref、status、state timeline、input / output summary、cost / token snapshot、trace id、failure code 和 audit refs。
   - 不提供 replay、resume、materialized result reader 或 durable run store。
4. `tool action preview`
   - 展示 candidate action、tool ref、risk level、requires_confirmation、policy reason、blocked state 和 audit ref。
   - 不执行 tool，不写上层业务真相源。
5. `confirmation placeholder`
   - 展示 future confirmation decision shape、risk summary、human review requirement、disabled reason 和 audit ref。
   - 不提交、不持久化、不解锁执行，也不写回。

## 程序化证据

- `scripts/checks/fixtures/workflow-function-surface-boundary-v1.json`
- `scripts/checks/control_plane/check-workflow-function-surface-boundary-v1.py`
- `scripts/check-repo.py` fast baseline

## 后续允许切片

- `workflow-definition-detail-read-v1`：只读 workflow definition detail view，优先消费 offline fixture 或 fake-store dev path。
- `workflow-run-detail-read-v1`：只读 run detail view，展示 state timeline、trace、cost、failure 和 audit refs。
- `workflow-blocked-action-preview-v1`：blocked action explanation，只解释为什么不能执行以及未来需要什么确认条件。

## 停止线

- 不新增 runtime API。
- 不实现 workflow builder、definition mutation、workflow executor、node executor 或 tool executor。
- 不实现 confirmation decision、business writeback、run replay、run resume、durable run store 或 materialized result reader。
- 不实现真实数据库、Radish OIDC、production API consumer 或 production readiness。
