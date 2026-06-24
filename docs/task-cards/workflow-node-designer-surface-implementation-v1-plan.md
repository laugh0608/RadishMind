# Workflow Node Designer Surface Implementation v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-surface-implementation-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_surface_implementation_v1_implemented`

## 目标

在 `Workflow Node Designer Surface v1` 和 `Workflow Node Designer Library Selection v1` 已完成后，接入 `@xyflow/react`，把现有 Draft Designer 的列表式节点结构升级为可操作节点画布体验。

本任务卡只承接前端依赖引入、画布 view model、节点 / 边组件、inspector bridge、validation overlay 和现有 saved draft / Review Handoff 消费关系。不新增 backend route、saved draft schema、repository mode、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 实施结果

- 已在 `apps/radishmind-web` 引入 `@xyflow/react`，并提交 `package-lock.json` 固定依赖版本。
- 已新增 `WorkflowNodeDesigner` 前端组件，把 active draft 派生为 React Flow nodes / edges view model；React Flow 只作为画布视图与交互层，不成为业务真相源。
- 已接入 custom node、custom edge、typed handle、viewport、drag / selection、connection validation feedback、inspector bridge 和删除保护。
- 已在 Draft Designer 中保留原列表式节点编辑区域作为 fallback；画布编辑仍复用现有 active draft reducer、save / read / validate / version conflict 语义。
- 已通过 `apps/radishmind-web` 的 `npm run build`，并保持本批不新增 persisted schema、backend route、runtime artifact 或 executor 能力。

## 本轮实现范围

- 在 `apps/radishmind-web` 引入 `@xyflow/react`，提交 `package-lock.json`，并在稳定样式入口加载 React Flow 必需 CSS。
- 新增 Node Designer graph adapter，把当前 `WorkflowDraftDesignerDraft` 派生为 React Flow `nodes` / `edges` view model。
- 新增 custom node 组件，展示 node type、label、summary、typed port、provider / profile / tool / RAG ref、risk marker、blocked marker 和 protected marker。
- 新增 custom edge 或 edge data mapping，表达 `data_edge`、`control_edge`、`guard_edge` 和 `audit_edge` 的 UI derived 语义；不把 edge kind 写入 saved draft persisted schema。
- 在 Draft Designer 中加入 Node Designer 画布区域，覆盖节点展示、移动、选中、连线、删除保护和 viewport 控制。
- 通过 inspector bridge 复用现有节点属性编辑逻辑；所有编辑仍回到 active draft reducer / mapping，再由 graph adapter 派生画布状态。
- 将 validation findings、risk、blocked capability 和 Review Handoff 相关状态投射到画布 overlay，不把 `valid_for_review` 写成 publish / run ready。
- 保留现有列表式 Draft Designer 作为明确 fallback；画布不可用时不得回退 sample 或伪装保存成功。

## 输入事实源

- [Workflow Node Designer Surface v1 专题](../features/workflow/node-designer-surface-v1.md)
- [Workflow Node Designer Library Selection v1 专题](../features/workflow/node-designer-library-selection-v1.md)
- [Workflow Draft Designer Editing Model v2 专题](../features/workflow/draft-designer-editing-model-v2.md)
- [Workflow Draft Node Attribute Editing Model v1 专题](../features/workflow/draft-node-attribute-editing-model-v1.md)
- [Workflow Review Handoff Active Draft v1 专题](../features/workflow/review-handoff-active-draft-v1.md)
- [Saved Workflow Draft v1 功能专题](../features/workflow/saved-workflow-draft-v1.md)

## 验收口径

- `@xyflow/react` 只作为前端 view / interaction 依赖；RadishMind active draft 仍是业务真相源。
- 画布节点和边必须由 active draft 派生，不能把 React Flow JSON 直接保存为长期 schema。
- 新增、移动、选中、连线、删除保护和 inspector 编辑必须继续标记 `local_edit` / `unsaved_local`，并复用现有 validate / save / read / version conflict 语义。
- 连线创建必须有 typed port / edge validation feedback；不符合当前 schema 的 edge kind 只能作为 UI derived state 或 validation finding 展示。
- validation overlay 必须展示风险、blocked capability、requires confirmation 和 readiness 摘要，且不提供 production `Run` 控件。
- 画布不可用、CSS 未加载或容器尺寸异常时必须明确显示 unavailable 状态，并保留列表式 Draft Designer。
- `npm run build` 在 `apps/radishmind-web` 通过。
- 相关前端验证覆盖节点展示、移动、选中、连线、删除保护、validation overlay、save / read failure state 和 version conflict。
- `./scripts/check-repo.sh --fast` 通过；若实现中新增 schema、route、生产声明或高风险边界，补专项 fixture / checker。

## 停止线

- 不新增 saved draft persisted schema、route contract、repository adapter、database schema、migration artifact 或 auth contract。
- 不使用 React Flow Pro 组件、商业模板或未审查的外部 UI 模板作为 committed 依赖。
- 不把 React Flow `nodes` / `edges` JSON 作为 RadishMind 长期持久化格式。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
