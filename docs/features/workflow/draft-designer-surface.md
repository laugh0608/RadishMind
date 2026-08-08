# Workflow Draft Designer Surface 专题

更新时间：2026-08-08

状态：`workflow_draft_designer_surface_v1_implemented`

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
7. validation、conflict、mapping、plan 和 handoff 使用渐进审查面，不继续平铺为等权卡片。

窄屏 `390x844` 按“草案上下文 → 主动作 → 当前节点 → 有界画布 → inspector → validation / handoff”重排。页面不得横向溢出；画布内部可以平移和缩放。Saved Draft Library、画布和 inspector 不在同一行压缩。

## S3 产品化实现结果

- `App.tsx` 继续唯一持有 application / workflow / run / draft / scenario 选择、editable draft、dirty、Saved Draft consumer / lifecycle / conflict 与全部 mutation action；`workflowDraftDesignerPanel.tsx` 只消费现有 view model 和 callbacks，并以 lazy chunk 承担 Workbench 展示。
- Designer 顶部显示真实 application、draft、content / lifecycle version、lifecycle 与 edit state；动作只复用 `Validate`、`Save draft`、`Read saved`、`Preview plan` 与 `Review handoff`，没有新增发布、运行、导出或发送。
- 完整 Saved Draft Library 仍只挂载在 Workspace Home；Designer 左轨只承载紧凑草案引用、节点类型和精确打开交接。Library 的活动 / 归档 tab 使用 action token，并补齐 `tablist` / `tab` / `aria-selected` 语义。
- `WorkflowNodeDesigner` 以草案 id 作为组件 key，切换草案时重置 viewport、selection 与 validation focus；Delete / Backspace、React Flow remove change 和锁定态 mutation 都不能绕过 App 的受控动作与 protected-node guard。
- `1280px` 起 Inspector 下移，`760px` 起页面单列且隐藏 MiniMap；`390x844` 保持当前节点 → 有界画布 → Inspector → validation / review 的真实 DOM 顺序，画布内部平移不扩张 body。

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

当前节点、当前 validation finding 和当前产品导航使用不同的选中结构与文字通道。`blocked`、`review_required`、`failed` 等状态颜色不自动获得选中高亮。

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

- Web 测试和 production build 通过。
- Saved Draft consumer smoke 继续覆盖 no sample fallback、双版本 conflict、严格 cursor、归档只读和解除归档后重新打开。
- 画布交互覆盖节点选择 / 拖拽、合法与非法连线、edge 删除、节点删除保护、validation focus、pending lock 和 inspector 编辑。
- 应用内浏览器复核 `1440x900`、共享壳层临界宽度、Designer 自身重排断点与 `390x844`；页面无横向溢出，画布内部平移不扩张 body，控制台无新增错误。
- S3 只复用现有 Web 测试、consumer smoke、build 和仓库门禁，不为普通 UI 产品化新增专项 checker。
- 2026-08-08 实际证据为 Web `274/274`、production build、六组既有 checker，以及应用内浏览器 `1440x900`、`1281/1280px`、`1101/1100px`、`761/760px`、`390x844`；保存 / 读取 / 派生 / 归档 / 解除归档与只读审查链通过，页面级横向溢出为零，控制台零 warning / error。

## 停止线

- 不新增 API、schema、migration、repository、依赖、task card、fixture 或专项 checker。
- 不新增持久化 node type、edge 字段或 React Flow 存储格式。
- 不实现自动保存、自动打开、自动覆盖、自动合并、三方合并、批量 lifecycle、永久删除或跨作用域移动。
- 不由 Designer 新增或解锁 publish、run、executor、agent loop、confirmation decision、writeback、replay、resume 或 materialized result reader；既有受控执行 owner 保持独立。
- 不保存、导出或发送 handoff，不声明 production repository、production auth、生产发布或生产执行就绪。
- 列表式字段编辑继续作为精细编辑与 fallback 路径，不与主画布争夺首屏。
