# Workflow Node Designer Persisted Layout v1 任务卡

更新时间：2026-06-24

## 任务标识

- 切片：`workflow-node-designer-persisted-layout-v1`
- 轨道：`Workflow / Agent Runtime`
- 状态：`workflow_node_designer_persisted_layout_v1_implemented`

## 目标

在 `Workflow Node Designer Saved Draft Mapping v1` 已完成 UI-only layout mapping 且 `Workflow Node Designer Review Handoff v1` 已形成审查消费后，把节点坐标推进为 saved draft 的受控 persisted layout metadata。

本任务卡只承接 `additional_fields.designer_layout_v1` schema、前端保存 / 恢复映射、Go domain record 保留、HTTP 响应透传、layout 规范化、forbidden field 负向验证和专项 checker。不新增 backend route、repository mode、数据库、OIDC、public production API、publish、run、executor、confirmation decision、writeback 或 replay。

## 本轮实现

- `WorkflowDraftDesignerLayout.persistence` 支持 `saved_draft_metadata`，用于标记 layout 来源。
- Save / Validate Draft payload 写入 `additional_fields.designer_layout_v1`，只保存当前 draft 已知节点的 `node_id`、`x`、`y` 与 `pinned:false`。
- saved draft restore 从 `designer_layout_v1` 恢复节点坐标，非法 layout、未知节点或缺失节点不会阻断草案主体读取。
- Go `SavedWorkflowDraft`、memory store clone、payload document conversion 和 HTTP save / read 响应保留 `AdditionalFields`。
- Go domain 规范化 layout metadata，并递归拒绝 `additional_fields` 内的 secret、token、run output、confirmation decision、writeback payload 或 executor result。
- Node Designer mapping summary 与 Review Handoff 摘要同步 persisted layout 语义。
- 新增 `workflow-node-designer-persisted-layout-v1` fixture / checker，并接入 `scripts/check-repo.py` fast baseline。

## 输入事实源

- [Workflow Node Designer Persisted Layout v1 专题](../features/workflow/node-designer-persisted-layout-v1.md)
- [Workflow Node Designer Saved Draft Mapping v1 专题](../features/workflow/node-designer-saved-draft-mapping-v1.md)
- [Workflow Node Designer Review Handoff v1 专题](../features/workflow/node-designer-review-handoff-v1.md)
- [Saved Workflow Draft v1 专题](../features/workflow/saved-workflow-draft-v1.md)

## 验收口径

- `additional_fields.designer_layout_v1` schema 必须使用 `layout_version/source/persistence/nodes`，不得保存 React Flow 原始对象。
- 前端保存必须过滤未知节点、限幅坐标、固定 `pinned:false`。
- 前端恢复必须兼容旧草案，layout 缺失或非法时继续使用默认布局。
- Go save / read / HTTP response 必须保留 layout metadata，并且不新增 runtime side effects。
- 嵌套 forbidden field 必须被拒绝，而不是在规范化后静默通过。
- Web build、Go tests、专项 checker 和 `./scripts/check-repo.sh --fast` 通过。

## 停止线

- 不保存 React Flow 原始 `nodes` / `edges` JSON、viewport、selection、drag state、connection preview、visual edge style 或 derived edge kind。
- 不把 layout metadata 解释为运行顺序、执行计划、runtime binding、publish ready、run ready 或 production ready。
- 不新增 backend route、repository adapter、repository mode、真实数据库、schema migration、store selector、Radish OIDC middleware、token validation、membership adapter、public production API、publish、run、executor、confirmation decision、writeback、replay 或 materialized result reader。
