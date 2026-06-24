# Workflow Node Designer Validation Overlay Navigation v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-validation-overlay-navigation-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_validation_overlay_navigation_v1_implemented`

## 目标

在 Builder interaction polish 已落地后，把已有 validation inspector finding 转成 Node Designer 内可导航的审查队列：用户可以从 structural / contract finding 定位到相关节点和连线，画布同步高亮，inspector 同步选中首个相关节点。

本任务只增加 UI-only navigation / highlight state、文档和既有专项 checker 证据；不扩 Go schema，不新增 backend route，不保存 viewport / selection / validation focus / React Flow raw object / derived edge kind / runtime order，不打开 repository mode、真实数据库、OIDC、production API、publish、run、executor、confirmation、writeback 或 replay。

## 输入事实源

- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Builder Interaction Polish v1 任务卡](workflow-node-designer-builder-interaction-polish-v1-plan.md)
- [Workflow Node Designer Controlled Edge Mutation Implementation v1 任务卡](workflow-node-designer-controlled-edge-mutation-implementation-v1-plan.md)
- [Workflow Draft Validation Inspector Offline v1 任务卡](workflow-draft-validation-inspector-offline-v1-plan.md)

## 实现范围

- `WorkflowNodeDesigner` 接收 active draft 的 `WorkflowDraftValidationInspectorViewModel`，不在画布内另建校验真相源。
- 将 structural checks 的 `evidenceRefs` 和 contract checks 的字段缺口映射成 UI-only navigation item。
- 点击 navigation item 后，选中首个相关节点、高亮相关节点 / edge，并更新交互反馈。
- 没有 graph target 的 finding 继续可见，但不伪造 fallback node / edge。
- Review Handoff、Preview Plan、Runtime Readiness 和保存链路继续消费 active draft。

## 验收口径

- navigation focus 只保存在 React state，不写入 draft、saved draft payload 或 `additional_fields`。
- structural finding 使用 `evidenceRefs` 定位节点；edge 高亮只由现有 active draft endpoint 派生。
- contract finding 只定位现有 context / output 节点和相关边，不扩 schema 或新增 validation 字段。
- 点击 finding 后 inspector 与画布 selection 保持一致，且不会触发 local edit / unsaved local mutation。
- 仍禁止保存 React Flow 原始对象、viewport、selection、validation focus、visual edge style、derived edge kind 或 runtime order。

## 验证要求

- `npm run build`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-node-designer-edge-editing-save-preconditions-v1.py`
- `git diff --check`
- `./scripts/check-repo.sh --fast`

## 停止线

- 不扩 saved draft schema、Go document type 或 backend route。
- 不保存 validation navigation state、React Flow 原始 node / edge 对象、handle id、port id、viewport、selection、connection preview、visual edge style、derived edge kind 或 runtime order。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、writeback、replay、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。
