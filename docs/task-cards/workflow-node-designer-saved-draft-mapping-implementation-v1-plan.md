# Workflow Node Designer Saved Draft Mapping Implementation v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-saved-draft-mapping-implementation-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_saved_draft_mapping_implementation_v1_implemented`

## 目标

在 [Workflow Node Designer Saved Draft Mapping v1](../features/workflow/node-designer-saved-draft-mapping-v1.md) 已定义后，实现不扩 persisted schema 的 UI-only layout mapping。

本任务卡只承接 active draft 内的 designer layout state、节点拖拽回写、mapping summary 和现有 Draft Designer / Node Designer 消费关系。不新增 saved draft persisted schema、backend route、Go document type、fixture、checker、repository mode、数据库、OIDC、publish、run、executor、confirmation decision、writeback 或 replay。

## 实施结果

- 已在 `WorkflowDraftDesignerDraft` 中新增 UI-only `designerLayout`，默认不进入 saved draft persisted schema。
- 已让 `WorkflowNodeDesigner` 优先消费 active draft layout，缺少坐标时继续按 lane / node order 派生默认布局。
- 已在节点拖拽结束后回写 active draft layout，并复用现有 `local_edit` / `unsaved_local` 状态。
- 已新增 mapping summary，明确 Save Draft 当前只写节点属性、edge endpoints 和 `condition_summary`，layout 与 derived edge kind 不写入 persisted schema。
- 已通过 `apps/radishmind-web` 的 `npm run build`。

## 本轮实现范围

- 在 `WorkflowDraftDesignerDraft` 中新增 UI-only `designerLayout`，用于记录当前 active draft session 的节点坐标。
- 让 `WorkflowNodeDesigner` 优先使用 `designerLayout.nodePositions`，缺失时继续按 lane / node order 派生默认布局。
- 节点拖拽结束后回写 active draft layout，并标记 `local_edit` / `unsaved_local`。
- 新增 mapping summary，明确 saved draft 当前仍只保存节点属性、edge endpoints 和 `condition_summary`；layout 与 edge kind 不进入 persisted schema。
- 保留 saved draft restore 的默认布局策略，不写 `additional_fields.designer_layout_v1`。
- 保留列表式 Draft Designer fallback。

## 验收口径

- `npm run build` 在 `apps/radishmind-web` 通过。
- `./scripts/check-repo.sh --fast` 通过。
- 拖拽节点后同一 active draft session 内布局可恢复；切换或 restore saved draft 后缺少 persisted layout 时仍按默认布局生成。
- Save / validate / read / version conflict 语义继续复用现有 saved draft consumer。
- 不新增 persisted schema、backend route、fixture、checker 或 runtime artifact。

## 停止线

- 不保存 React Flow 原始 `nodes` / `edges` JSON。
- 不写入 `additional_fields.designer_layout_v1`；若后续需要跨会话持久化 layout，必须另开 schema / fixture / checker 任务卡。
- 不把 layout、viewport、selection、visual edge style 或 edge kind 解释为运行顺序、执行计划或 runtime binding。
- 不实现 publish、run、workflow executor、node executor、tool executor、agent loop、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不接 repository mode、真实数据库、database connection provider、secret resolver、production resolver runtime、schema marker runtime、migration runner、Radish OIDC middleware、token validation、membership adapter、API key lifecycle、quota、billing 或 public production API。
