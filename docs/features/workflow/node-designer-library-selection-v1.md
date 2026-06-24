# Workflow Node Designer Library Selection v1 专题

更新时间：2026-06-24

状态：`workflow_node_designer_library_selection_v1_defined`

## 专题定位

`Workflow Node Designer Library Selection v1` 承接 [Workflow Node Designer Surface v1](node-designer-surface-v1.md)，固定首批节点画布实现前的库选型、状态模型、依赖引入方式、验证方式和停止线。

本专题定义阶段不安装依赖、不提交 UI 实现、不新增 task card / fixture / checker、不改变 saved draft schema，也不打开 workflow publish、run、executor、confirmation、writeback、replay、repository mode、真实数据库、OIDC middleware、token validation、membership adapter 或 public production API。

## 选型结论

`Workflow Node Designer Surface Implementation v1` 已按本专题选型选择 `@xyflow/react` 作为 React 画布库。

选型依据：

- 与当前 `apps/radishmind-web` 的 `React + Vite + TypeScript` 技术栈直接匹配。
- 原生围绕 nodes、edges、handles、custom nodes、custom edges、viewport、selection 和 callbacks 组织，能承载 RadishMind 的 typed port / edge / validation overlay。
- 支持受控 nodes / edges 状态模型，适合让 RadishMind 的 active draft 继续做业务真相源，画布库只做交互和视图层。
- 官方文档包含 TypeScript、testing、performance、state management、custom nodes、handles、edges 和 Vite 模板路径，后续实现验证边界清晰。
- GitHub 仓库声明 React Flow / Svelte Flow 为 MIT license，且 `@xyflow/react` 是 xyflow monorepo 中的 React Flow 12 包。

实现批次已通过 `npm install @xyflow/react` 产生的 `package-lock.json` 固定实际版本；本专题仍只作为选型和依赖边界记录，不把画布依赖解释为 saved draft schema、runtime artifact、publish / run 或 executor 已打开。

## 候选对比

| 候选 | 结论 | 理由 |
| --- | --- | --- |
| `@xyflow/react` | 选定 | React / Vite / TypeScript 直接匹配，节点、边、handle 和视口能力贴合 Builder Surface，能让项目自有 active draft 状态继续做真相源 |
| `Rete.js` | 暂不采用 | TypeScript-first 且能力强，但官方定位偏 visual programming / processing-oriented node editor，包含 engine / processing / code generation 方向；当前阶段不应把 UI 画布和执行语义耦合 |
| `BaklavaJS` | 暂不采用 | 官方定位为 VueJS browser graph / node editor，生态方向与当前 React 产品 UI 不匹配 |
| 手写 SVG / canvas / D3 | 暂不采用 | 会把拖拽、缩放、选区、连线、键盘交互、无障碍和测试成本提前压进本仓库，不符合当前“先推进真实 Builder 体验、保持实现边界清晰”的原则 |

## 状态模型

画布实现必须把 `@xyflow/react` 视为 view-model 层，而不是 workflow domain truth source。

推荐拆分：

- `active draft graph`：沿用现有 Draft Designer / Saved Workflow Draft v1 的节点、边、contract、refs、validation 和 version metadata。
- `canvas view model`：由 active draft 派生 `nodes`、`edges`、`viewport`、`selection`、`connection draft` 和 `validation overlay`。
- `node data`：只保存渲染所需的 `draft_node_id`、node type、label、summary、contract summary、provider / profile / tool / RAG ref、risk marker、blocked marker 和 protected marker。
- `handle id`：由 `draft_node_id + direction + port_id` 组合生成，表达 typed port，不把 handle id 写成新的持久化 schema 真相源。
- `edge data`：可携带 `data_edge`、`control_edge`、`guard_edge` 或 `audit_edge` 的 UI derived kind；若 saved draft schema 尚不能表达 edge kind，只能作为 UI derived state 或 validation finding。

所有新增、移动、选中、连线、删除保护和 inspector 编辑，都应回到现有 active draft reducer / mapping 层处理，再派生给画布组件。不得让画布组件直接绕过 Draft Designer 的 save / validate / read / version conflict 语义。

## 依赖引入边界

实现任务卡已按以下边界落地：

- 在 `apps/radishmind-web` 引入 `@xyflow/react` 依赖，并提交 `package-lock.json`。
- 在全局样式入口引入 React Flow 必需 CSS，并确认父容器有稳定宽高。
- 新增 Node Designer 专用 graph adapter、node component、edge component、inspector bridge 和 validation overlay 组件。
- 复用现有 workflow 聚合 gate、web build、saved draft consumer smoke 和 fast baseline。

实现任务卡仍不允许：

- 使用 React Flow Pro 组件或商业模板作为 committed 依赖。
- 直接复制官方模板成为产品页面结构。
- 在同一批里修改 saved draft persisted schema、route contract、repository mode、database schema 或 auth runtime。
- 把 React Flow 的 `nodes` / `edges` JSON 直接作为 RadishMind 的长期持久化格式。

## 降级与失败策略

画布依赖不可用、CSS 未加载、容器尺寸异常或渲染失败时，不允许静默回退为 sample draft，也不允许伪装为保存成功。

允许的降级方式：

- 保留现有列表式 Draft Designer 编辑入口。
- 显示明确的 node designer unavailable 状态，并保留 active draft / saved draft 的 read、validate、save 和 version conflict 语义。
- 后续实现可把画布 mount failure 纳入前端单元测试或 smoke 记录，但不为文档选型阶段新增 checker。

## 实现任务边界

[Workflow Node Designer Surface Implementation v1 任务卡](../../task-cards/workflow-node-designer-surface-implementation-v1-plan.md) 已实现，范围限于：

- 画布区域接入。
- 节点展示、移动、选中、删除保护。
- 端口展示、连线创建和 edge validation feedback。
- 当前节点 inspector 与 active draft 的双向映射。
- validation findings / risk / blocked capability overlay。
- `Preview Plan`、`Save Draft`、`Review Handoff` 等现有动作入口的状态消费。

`Workflow Node Designer Saved Draft Mapping v1` 已定义 layout metadata 与 edge kind 的 saved draft 映射边界。后续如果要保存 layout metadata 或扩展 edge kind persisted schema，必须按该专题结论补实现 task card、fixture、checker 和对应验证。

## 验收方式

本专题定义阶段：

- 本文档进入 workflow 细专题入口。
- `Workflow Node Designer Surface v1` 的实现拆分更新为“库选型已完成，surface implementation 任务卡已实现”。
- 当前焦点能说明 Builder 体验下一步转为 persisted layout schema 评审或 durable store 上游前置，而不是进入 executor。
- 不产生依赖、代码、schema、fixture、checker 或 runtime artifact。
- `./scripts/check-repo.sh --fast` 通过。

实现阶段：

- `npm run build` 在 `apps/radishmind-web` 通过。
- 节点展示、移动、选中、连线校验反馈、删除保护、validation overlay、save / read failure state 和 version conflict 至少有对应前端验证。
- Saved draft consumer smoke 继续覆盖 no sample fallback、version conflict 和 failure code。
- 若新增 schema、route、dependency boundary 或高风险能力，补实现任务卡、专项 fixture / checker 和仓库基线。

## 停止线

- 本专题定义阶段不安装 `@xyflow/react` 或修改 `package.json`；依赖引入只允许由 implementation 任务卡承接。
- 不提交画布 UI、node component、edge component、inspector bridge 或 runtime 代码。
- 不修改 saved draft persisted schema、route contract、repository adapter、database schema、migration artifact 或 auth contract。
- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。

## 参考资料

- `@xyflow/react` / React Flow 官方文档：https://reactflow.dev/learn
- xyflow GitHub 仓库：https://github.com/xyflow/xyflow
- Rete.js 官方文档：https://retejs.org/
- BaklavaJS 官方文档：https://baklava.tech/
