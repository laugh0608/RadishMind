# Workflow Draft Designer Surface 专题

更新时间：2026-08-08

状态：`workflow_draft_designer_surface_v1_s3_r2_implemented`

## 专题定位

`Workflow Draft Designer Surface` 是 `apps/radishmind-web/` 中承载 workflow 草案选择、结构编辑、节点画布、属性检查、校验、计划预览和人工审查交接的页面专题。

它复用现有 Saved Workflow Draft owner、Node Designer、Validation Inspector、Execution Plan Preview、Runtime Readiness Inspector 与 Review Handoff，不创建第二套草案、校验、计划或审查真相源。当前 memory、SQLite 与 PostgreSQL 开发测试态 repository 已成立；production repository、production auth、公开生产 API、自动发布和自动执行继续关闭。

## 当前实现

- Workspace Home 的 Saved Draft Library 已支持活动 / 归档独立查询、服务端筛选、严格 cursor、加载更多、精确打开、两步归档和显式解除归档。
- Draft Designer 已支持草案名称、说明、节点结构、节点属性、contract fields、引用摘要、output mapping 和 edge condition summary 的受控编辑。
- Node Designer 已支持 React Flow 画布、节点选择 / 拖拽、受控连线新增 / 删除、节点删除保护、inspector、validation finding 定位和易失交互反馈。
- Save Draft 会把受控节点位置写入 `additional_fields.designer_layout_v1`；React Flow 原始对象、viewport、selection、handle、derived edge kind 与 runtime order 不进入 saved record。
- 保存、读取和校验继续复用显式 dev-only consumer；失败保留当前本地草案且不回退 sample，版本冲突不自动覆盖或合并。
- active draft 继续驱动 validation、plan preview、runtime readiness 与 Review Handoff；这些派生结果不持久化、不导出、不发送。
- 归档草案保持内容、revision 与比较只读；解除归档后旧浏览器快照保持 `unknown`，必须从活动草案库重新精确打开。

## S3 产品化目标

S3 把上述既有能力重组为共享 Workbench 壳层中的连续 Designer 任务面，不新增运行时能力。

桌面 `1440x900`：

1. `Workflows` 是唯一一级当前导航。
2. 顶部固定 workspace、application、workflow definition、draft identity、内容版本、lifecycle 和 saved state。
3. 主动作只表达 `Validate`、`Save Draft`、`Preview Plan` 与 `Review Handoff`；不得使用 production `Run` 误导当前能力。
4. 左侧承载草案范围、Saved Draft 摘要、节点类型与结构入口。
5. 中央 React Flow 画布是唯一主任务面。
6. 右侧只承载当前节点或连线 inspector。
7. validation、conflict、mapping、plan 和 handoff 使用紧凑摘要与渐进审查入口，不继续平铺为等权卡片。
8. 画布默认聚焦当前节点与直接邻居，并提供显式 `Fit graph` 查看全图；不能以缩小到不可读或裁掉节点来满足节点计数。

窄屏 `390x844` 按“草案上下文 → 主动作 → 当前节点 / 直接邻居画布 → 紧凑审查”重排。页面不得横向溢出；画布内部可以平移和缩放，Inspector 在 `<=760px` 默认折叠且不得与当前节点摘要重复。Saved Draft Library、画布和 Inspector 不在同一行压缩。

## S3 R1 人工退回与 R2 修正

- `S3 R1` 已完成 App owner / Panel presentation 的职责拆分、受控 mutation 保护和 Saved Draft Library 唯一挂载；Web 行为测试、production build、既有 checker 与页面级横向溢出检查均通过。这些结果继续作为行为回归证据，但不再代表视觉验收。
- 后续人工评审确认 `R1` 存在确定性画布裁切：页面可以没有 body 横向溢出，但 React Flow 内的节点仍被固定视口、最小缩放与隐藏溢出裁掉，导致“4 nodes”与实际可见节点错位，边也无法完整阅读。底部 Validation / Preview Plan / Review Handoff 三块等权，窄屏又重复当前节点摘要与完整 Inspector，主任务和审查层级不成立，因此 `R1` 被退回。
- `S3 R2` 已改为默认聚焦当前节点与直接邻居，保证节点、端口和连线在首屏可读；显式 `Fit graph` 才承担全图展示。选择节点或 validation finding 会把目标带入可读视区，不再只更新选中状态而让目标留在裁切区外。
- review 已改为紧凑摘要和按需展开入口。Inspector 在 `1440px` 固定于画布右侧，`<=1380px` 下移到画布之后，`<=760px` 默认折叠；窄屏不再重复同一节点的切换摘要与完整 Inspector。
- 强选中只属于当前产品导航或正在驱动 Inspector 的节点 / 连线；普通节点、Saved Draft 引用和 review 摘要保持中性，finding focus 与 lifecycle / readiness / failure 状态使用独立结构、文字和图标通道。
- `App.tsx` 继续唯一持有 application / workflow / run / draft / scenario 选择、editable draft、dirty、Saved Draft consumer / lifecycle / conflict 与全部 mutation action；`workflowDraftDesignerPanel.tsx` 只消费现有 view model 和 callbacks。九组来源、十三项 contribution、revision `partial`、RAG authority `blocked`、readiness 只读且不可发布的边界不变。
- Designer 动作仍只复用 `Validate`、`Save draft`、`Read saved`、`Preview plan` 与 `Review handoff`；没有新增发布、运行、导出或发送，也不新增 API、schema、repository、task card、fixture 或专项 checker。

## 状态模型

| 状态轴 | 状态 | 页面语义 |
| --- | --- | --- |
| 本地来源 | `sample` | 离线审查，不持久化，不冒充 saved record |
| 本地编辑 | `local_edit` | active draft 的本地编辑标记，不是 consumer 成功状态 |
| 待保存 | `unsaved_local` | 可校验或保存，仍非生产记录 |
| 请求中 | `saving` / `validating` / `reading` | 画布可查看和选择；拖拽、连线、删除和字段编辑锁定 |
| 保存成功 | `saved_dev_record` | 已写入配置的开发测试态 store，不表示发布或运行就绪 |
| 校验成功 | `validation_ready` | 当前结果可审查；validation finding 不持久化 |
| 版本冲突 | `version_conflict` | 保留本地草案并阻断继续 mutation，等待显式选择 |
| 继续本地 | `conflict_local_continued` | 显式保留本地草案，不自动覆盖、合并或打开 saved record |
| 失败 | `save_failed` / `read_failed` / `validation_failed` | 展示稳定 failure code，保留本地上下文且不回退 sample |
| 生命周期 | `active` | 无 pending、conflict 或其它既有锁时可编辑 |
| 生命周期 | `archived` | 内容、revision、比较与审查可读；编辑和下游 mutation 关闭 |
| 生命周期 | `unknown` | 解除归档后的旧只读快照；必须从活动库重新打开 |

当前产品导航和正在驱动 Inspector 的节点 / 连线才使用强选中结构；当前 validation finding 使用独立 focus 结构与文字通道。`blocked`、`review_required`、`failed` 等状态颜色不自动获得选中高亮。

## 保存与派生边界

允许进入 saved draft：

- 节点位置 `additional_fields.designer_layout_v1`。
- 节点 label、summary、provider / profile ref、tool ref、RAG ref。
- input / output contract fields 与 output mapping summary。
- edge endpoint 与 condition summary。

只允许派生或保留在 UI：

- React Flow 原始 `nodes` / `edges`、handle / port id、viewport、selection 和 drag state。
- validation focus、validation finding 与画布视觉高亮。
- derived edge kind、visual edge style 与 runtime order。
- execution plan preview、runtime readiness 与 Review Handoff record。

保存失败、读取失败和版本冲突都必须保留当前本地草案。归档态 read 不得生成可保存的 active editor state；解除归档不会自动打开、重放或恢复归档前 pending mutation。

## 验收方式

- Pencil `S3 Workflow Designer — Desktop / Partial · R2`、`S3 Workflow Designer — Narrow / Partial · R2` 与 Decision R8 已完成全树和实际渲染复核，无裁切、越界或占位节点。
- Web `274/274` 测试与 production build 通过；`workflowNodeDesigner` 为 `205.40kB < 220kB`，`index` 为 `468.40kB < 500kB`。既有八项 workflow checker 全部通过，没有新增专项 checker。
- Saved Draft consumer smoke 继续覆盖 no sample fallback、双版本 conflict、严格 cursor、归档只读和解除归档后重新打开。
- 应用内浏览器严格复核覆盖 `1440x900`、`1381/1380px`、`1101/1100px`、`761/760px` 与 `390x844`。默认两节点邻域完整可读，显式 `Fit graph` 的八节点 / 七边全部完整；各宽度无横向溢出，Inspector 在 `1440px` 右侧、`<=1380px` 下移、`<=760px` 折叠，移动端 review 位于 rail 之前。
- 键盘删除保护、受保护节点、节点切换、finding 定位、Inspector 折叠 / 展开及强选中与状态通道分离均通过；全新页签控制台零 warning / error。
- `R1` 的 Web、build 和行为检查结果继续作为回归基线，但其视觉验收因未证明画布内容完整可读而保持撤回。`R2` 已完成严格复验，下一步进入 `S4 Application API Integration / API Key` 功能事实复核、`A` 级设计与 React 纵向切片。

## 停止线

- 不新增 API、schema、migration、repository、依赖、task card、fixture 或专项 checker。
- 不新增持久化 node type、edge 字段或 React Flow 存储格式。
- 不实现自动保存、自动打开、自动覆盖、自动合并、三方合并、批量 lifecycle、永久删除或跨作用域移动。
- 不由 Designer 新增或解锁 publish、run、executor、agent loop、confirmation decision、writeback、replay、resume 或 materialized result reader；既有受控执行 owner 保持独立。
- 不保存、导出或发送 handoff，不声明 production repository、production auth、生产发布或生产执行就绪。
- 列表式字段编辑继续作为精细编辑与 fallback 路径，不与主画布争夺首屏。
