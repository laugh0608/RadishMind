# RadishMind Family UI 产品化设计与迁移 v1

更新时间：2026-08-08

状态：`radishmind_family_ui_productization_v1_s1_r8_s2_r6_implemented`

## 目标

本专题把 `apps/radishmind-web/` 从已有功能完整、但视觉语言分散的内部开发者预览，逐步治理为 Radish 家族中可持续演进的 Workbench 产品面。

当前已经可以开始 UI 产品化，不要求先完成全部未来功能。开始条件是：拟设计页面的主要用户任务、真实状态和执行边界已经稳定，设计不会把未实现能力画成可用能力。后续新增功能仍必须先进入对应功能设计文档，不能因为进入 UI 阶段而跳过设计。

首批产品化不追求一次性改写全部样式。推进顺序固定为：

1. 建立 family-ui 规范、token 和项目差异基座。
2. 用真实任务和真实数据状态完成页面级产品设计。
3. 按纵向切片逐页实现、复验和迁移遗留样式。

## 现状依据

- 用户工作区、Application Development Workspace、Saved Draft、API Key、Gateway 调试与 Run / Evaluation 审查已有连续产品路径。
- 2026-07-30 的 SQLite 真实路径复盘确认领域行为正确，首批交接、状态解释和信息密度问题已经修正。
- Web 已有单元测试、production build、consumer smoke、宽屏与窄屏浏览器复验入口。
- 当前 `styles.css` 仍集中承载大量硬编码颜色、间距和组件样式，旧 UI 规范也混合了家族通用规则与 RadishMind 领域规则；继续直接叠加页面样式会扩大迁移成本。

因此，本阶段适合先固定产品设计基座，再做页面级视觉收敛；不适合等待所有长期功能完备，也不适合直接全量换色。

## 规范真相源

通用 UI 参考基线采用 RadishX 仓库 `docs/design/family-ui/` 的 `v26.7.3`：

- 家族通用原则、稳定语义、字体、间距、圆角、阴影、图标、组件和平台布局以上游规范为参考；上游不替项目决定产品配色、Profile、接入方式、采用状态或迁移顺序。
- RadishMind 只在 [UI 差异附录](../../ui-addendum.md) 中维护项目差异、专属组件和暂存偏差。
- RadishMind 主动选择 `Workbench` Profile，并把上游 token 原样镜像作为本项目的接入策略；镜像文件必须与对应 family-ui 版本一致，不在其中直接修改项目差异。
- 新的家族通用 token 应先回到 family-ui 讨论；RadishMind 专属语义先使用 `--rm-*` L2 别名或项目组件变量表达。

## 产品 Profile

RadishMind 项目主动选择 `Workbench` Profile：

- 画面气质是现代、安静、具象、可审计的中等密度工程工作台。
- 亮色底采用家族纸色，但不使用品牌面纹样和装饰性印章。
- 身份与家族归属使用灰玉语义，主交互使用墨蓝语义，成功与可继续状态使用玉色，紫色只用于小面积类型区分，胭脂只用于需要注意但不等同危险的项目语义。
- 颜色不是状态的唯一通道；状态必须同时具有文字、图标或结构语义。
- 信息密度允许高于品牌展示面，但主任务、当前上下文、高风险动作和失败原因必须有清晰层级。
- 页面、区域和卡片标题只显示稳定名称，不在标题上下追加 eyebrow、代表状态、营销句或临时解释句；状态、范围和来源数量进入 badge、字段或数据行。
- 常驻导航、普通主操作和边界摘要不使用大面积墨色反差块；默认用纸色层级、细边界与柔底语义色建立层次。
- 同一导航层级只有一个可见 owner；不得把图标产品轨与文字侧栏并列成两套主导航。跨产品域与当前产品内入口如果没有真实独立层级，统一收敛到一个侧栏。
- 页面优先建立一个能直接识别的主对象，再让证据、状态和边界贴合其组织；不把所有信息分成等权小卡。
- 视觉层级优先使用尺度、留白、表面材质和少量抬升感；当前导航、当前阶段或当前任务可成为柔和抬起的实体，其余区域保持平静。
- Workbench 不因可审计信息较多就把正文、图标和行高压成微缩密码；列表按内容确定可读高度，不使用 `fill_container` 拉伸行高填满视口。已实现的 `S1 React R8` 桌面 / 窄屏任务行分别为 `88px` / `96px`；`S2 R6` 继续复用这套可读尺度，不回退到 `R4` 的微缩任务行。

## 产品设计范围

### 第一组设计面

第一组以真实开发者连续任务为主，不按现有长页面的 DOM 顺序照搬：

1. 产品壳与工作区导航：当前 workspace、application context、一级任务入口、环境与能力边界。
2. Application Development Workspace：五阶段任务路径、readiness、精确 owner evidence 与 handoff。
3. Saved Draft Library 与 Workflow Designer：活动 / 归档草案、精确打开、画布、校验、执行计划和审查交接。
4. Application API Integration 与 API Key：模型资格、接入示例、一次性凭据交接、验证后退役和失败解释。

### 后续设计面

- Gateway Playground、Request History 与 Application Operations。
- Workflow Run History、Comparison、Evaluation Case / Suite 与人工发布审查。
- Admin / Control Plane 的只读证据、blocked action 和开发测试态管理面。
- Prompt 与 Agent / Copilot 的类型专属配置、受控调用和回归评测路径。

后续设计面不因列入范围而自动获得实现准入；每一批仍要以对应功能专题中的当前能力为边界。

## 页面与状态矩阵

每个页面设计至少覆盖以下状态中的适用集合：

| 状态 | 必须表达的内容 | 不允许的表达 |
| --- | --- | --- |
| ready | 当前上下文、主任务、下一动作和证据来源 | 只用绿色表示可用 |
| loading | 正在读取的 owner、保留的页面骨架和不可执行边界 | 用空白页面代替 |
| empty | 为什么为空、如何建立首个对象、是否缺少权限或前置条件 | 把空数据误写成系统故障 |
| partial | 已覆盖来源、未覆盖来源和当前窗口限制 | 把首分页或已加载窗口冒充全量 |
| failed / stale | 稳定错误语义、重试或重新打开路径、保留的安全上下文 | 展示原始响应正文或敏感信息 |
| blocked | 阻塞原因、需要谁确认或补齐什么前置条件 | 画成可点击但无真实 owner 的动作 |
| confirmation | 影响范围、CAS / drift、明确确认和取消 | 自动确认或弱化高风险影响 |

## 设计基线

- Web 主设计画板：`1440x900`。
- 宽屏复验：至少覆盖当前常用桌面尺寸；现有 `1920x1080` 设计证据可作为补充，不作为唯一基线。
- 窄屏主设计画板：`390x844`，按任务顺序重排为单列，不缩放桌面三栏。
- 交互状态：至少包含键盘 focus、hover、disabled、loading、error 和 confirmation。
- 设计源：Family UI 产品化设计使用 [radishmind-web-family-ui-v1.pen](../../designs/radishmind-web-family-ui-v1.pen)；既有 `radishmind-console-ops-surface-v0.pen` 保留为历史 Ops Surface 证据，不继续承载新的页面族。
- 外部参考：27 张 family-ui 参考图的逐项采用与排除边界已固定在 [Family UI 参考图产品面映射 v1](radishmind-family-ui-reference-mapping-v1.md)；Pencil 只标注 `ref-XX` 与吸收原则，不嵌入或复制外部截图。

## Pencil 协作模型

Pencil 只承载稳定的设计决策，不承载完整功能清单。功能、文案、按钮是否存在、可用条件、数据来源和权限边界继续以功能文档、API 契约与当前代码为准；Pencil 负责信息层级、布局结构、组件关系、交互语义、关键状态表达和响应式顺序。

“主要页面”统一改称“设计基准面”：

> 为一个页面族冻结信息架构、交互模型、风险表达或响应式策略的最小代表面。

路由重要、功能很多或页面很长，不自动构成设计基准面。只有页面产生了下游实现无法从既有模式安全推导的新设计决策，才进入完整 Pencil 设计。

### 判定维度

每个候选页面族按五个维度记 `0` 至 `2` 分：

| 维度 | `0` 分 | `1` 分 | `2` 分 |
| --- | --- | --- | --- |
| 结构新颖度 | 复用既有布局 | 既有结构的明显变体 | 新产品壳、导航、上下文或页面拓扑 |
| 交互新颖度 | 纯展示或标准控件 | 多个既有交互的组合 | 新画布、多阶段工作区、对比审查或一次性交接 |
| 边界风险 | 无新增风险表达 | 可逆的开发测试态动作 | 凭据、确认、高风险动作或生产语义误导风险 |
| 状态与窄屏复杂度 | 一至两个可推导状态 | 多状态或高密度内容 | 多来源 partial / blocked，或窄屏必须改变信息顺序 |
| 复用杠杆 | 单页面局部 | 可供两个页面复用 | 约束三个以上页面或全局产品壳 |

覆盖级别：

- `A / 完整 Pencil`：总分不少于 `6`，或结构新颖度、交互新颖度、边界风险任一项为 `2`。设计一个桌面代表面，只补无法从代表面推导的窄屏或关键风险状态。
- `B / 局部 Pencil`：总分 `3–5` 且没有上述强触发项。只设计新的区域、组件或状态，不建立完整重复页面。
- `C / 直接实现`：总分 `0–2` 且没有强触发项。复用已评审模式直接实现，以真实浏览器复核为验收。

评分用于统一讨论，不替代边界判断。页面在实现中暴露出新的结构、交互或风险决策时，应提升覆盖级别；普通文案、字段和按钮条件变化不因此升级。

### 首批覆盖矩阵

| 产品面 / 页面族 | 级别 | Pencil 覆盖 | 实现方式 |
| --- | --- | --- | --- |
| `S1` 产品壳与工作区导航 | `A` | `1440x900` partial 代表面冻结 ready 共用层级，`390x844` 补单列重排 | 先实现基准面，再让后续一级入口继承壳层 |
| `S2` Application Development Workspace | `A` | Application Context、五阶段导航、当前 surface、evidence / readiness 及关键 blocked / partial 表达 | 以一个真实 Application 连续任务作为代表，不为五个阶段各建完整画板 |
| `S3` Saved Draft Library | `B` | 仅在生命周期筛选、只读归档提示或交接结构无法从既有列表模式推导时补局部稿 | 沿列表、状态标签和精确打开模式直接实现并做浏览器复核 |
| `S3` Workflow Designer | `A` | 画布、节点 / 连线、inspector、校验与审查交接代表面 | 设计器作为新交互原型；相似检查面板直接复用 |
| `S4` API Integration / API Key | `A` | 接入任务主面与一次性凭据交接、验证后退役的关键风险状态 | API Key 风险状态并入同一页面族，不为每个生命周期状态复制完整页面 |

首批预计维护四个完整设计基准面和必要的局部 / 状态变体，不按路由、组件或状态数量扩张画板。

### 设计基准面覆盖记录

| `surface_id` | 状态 | Pencil 覆盖 | 代码基线与锚点 | 设计决策与停止线 |
| --- | --- | --- | --- | --- |
| `radishmind_web_s1_product_shell_v1` | `implementation_r8_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S1 Product Shell — Desktop / Partial · R8` 与 `S1 Product Shell — Narrow / Partial · R8` | `R8` 设计基准 commit `1c537537`，实现 commit `321e9899`；`apps/radishmind-web/src/app/ProductNavigation.tsx`、`apps/radishmind-web/src/app/App.tsx`、`apps/radishmind-web/src/features/control-plane-read/workspaceProductOverviewPanel.tsx`、`workspaceOperationsInboxPanel.tsx` 与 `apps/radishmind-web/src/styles.css` | `R8` 已落地唯一导航、workspace / application 上下文、全视口桌面壳、非对称 pulse、来源分布矩阵、连续 Inbox 列表 / 详情、轻量选中轨和 evidence path。Pencil 的四项队列与 `1 / 2 / 1` 只表示设计基准；真实离线 view model 在验收时投影五项队列与 `1 / 3 / 0` 来源分布，React 未伪造计数或改写 partial 停止线。 |
| `radishmind_web_s2_application_workspace_v1` | `implementation_r6_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S2 Application Workspace — Desktop / Partial · R6`、`S2 Application Workspace — Narrow / Partial · R6` 与 `S1 + S2 Visual Language — Design Decision Record · R6` | 代码审计基线 `a545f511`，实现于 2026-08-08；`applicationDevelopmentWorkspacePanel.tsx`、`applicationDevelopmentWorkspaceSurface.tsx`、`applicationDevelopmentWorkspace.ts`、`applicationDevelopmentWorkspaceRoute.ts`、`applicationDevelopmentReadiness.ts`、`applicationDevelopmentWorkspaceControls.ts`、`ProductNavigation.tsx` 与 `styles.css` | `R6` 已在共享壳层内落地五阶段、owner contribution 与 readiness 投影；Applications 和当前阶段分别形成唯一产品导航 / 阶段选中态，普通 contribution 始终保持中性。桌面使用 `3 / 13` 窗口轨、九格来源矩阵和 authorization path；窄屏按 Application Context、当前阶段选择器、三项贡献、readiness 摘要渐进重排。React 继续使用真实 view model，当前离线 fixture 为 `1 / 9` owner references，不复制 Pencil 的代表性 `5 / 9`。 |

`S1` 消费 `ref-03`、`ref-07`、`ref-08`、`ref-09`、`ref-15`、`ref-17`、`ref-18`、`ref-19` 与 `ref-24` 的产品层级，并共享 `ref-25` 的白色抬升表面和 `ref-27` 的连续窗格语法；`ref-20` 只保留未来暗色证据，不在当前画板增加主题切换。两个画板均已通过 Pencil 全树布局检查，结果为 `No layout problems`；视觉复核确认没有裁切、重叠和横向溢出。

首轮视觉复核发现，仅使用纸色 token、柔底状态和标准“侧栏 + 等宽 KPI 卡 + 列表 + 右栏”骨架，虽然满足结构约束，却不足以体现参考图与 Radish 家族的可辨识气质。`R2` 据此引入双层产品导航、非均质 workspace pulse、表格式关注队列与更明确的对象语义。

第二轮视觉复核确认，`R2` 的墨色产品主轨、普通操作黑底和墨色停止线卡对比过高，自动生成的 eyebrow / 标题描述属于临时设计文案，弹性任务行也降低了信息密度。`R3` 将以下规则收敛为当前可实施基线：

- 桌面产品主轨和窄屏 command bar 改用纸色层级与柔底选中态，普通主操作与边界摘要不再使用墨色反差块；
- 页面、区域和卡片标题只保留名称；状态、范围、数量与限制进入结构化 badge、字段或数据行，不在标题下补解释性副文案；
- workspace pulse 继续使用非均质摘要，不退回四张同等权重指标卡；
- Operations Inbox 使用参考 `ref-08` / `ref-24` 的表格式关注队列、severity 筛选和软印色状态，同时保留 owner、当前窗口与 partial 停止线；
- 小面积灰玉只用于家族 seal 等身份表达，墨蓝负责交互，胭脂只用于关注计数等注意语义，玉色和赭色只表达状态，雅紫只区分 application / workflow 对象；
- 桌面关注队列行高固定为 `68px`，窄屏三行任务条目固定为 `94px`；列表不再随剩余视口高度拉伸；
- 窄屏复用柔色 command bar、workspace pulse 和对象卡语法，按任务顺序单列重组，不缩小桌面导航和 KPI。

第三轮导航复核确认，`R3` 的图标产品轨与文字对象侧栏虽然配色已降低对比，但仍被感知为两套并列工具栏，且当前功能并不存在需要双轨表达的独立导航层级。`R4` 删除图标产品轨，把品牌、workspace 切换、一级入口、环境边界、账户和帮助统一归入一个桌面侧栏；窄屏继续使用单一 command bar 与折叠导航。后续只有存在真实、稳定且可独立切换的跨产品层级时，才重新评估产品轨，不为视觉效果预留第二套导航。

第四轮 `S1` / `S2` 联合视觉复核确认，`S1 R4` 和 `S2 R2` 的信息架构、真实状态与停止线没有错，但尺度过小、表面过碎、层级过于平均，整体更像“微缩审计原型”，没有形成参考 UI 的产品实体感。`R5` / `R3` 共同消费 `ref-18` 与 `ref-13` 的核心语法：

- 放大品牌、导航、搜索、对象身份和主标题，当前导航以纸白抬起实体和墨蓝强调形成双通道。
- 每个页面只保留一个主对象：`S1` 为 workspace pulse 与运营任务面，`S2` 为当前 Human Promotion 连续工作面；证据和边界贴合主面，不再四处分裂。
- 正文、图标、控件和行高回到可读尺度，背景保持清爽中性，阴影只用于导航当前态、主对象和必要分层。
- 墨蓝、玉色、紫色和风险色只在当前态、对象类型和必要状态中小面积出现，不用柔底色把页面切成彩色碎片。
- 窄屏不缩小桌面信息，而是保留大尺度当前阶段和连续任务面，把次要来源收入渐进披露。

2026-08-05 的第五轮联合人工评审进一步确认，`S1 R5` / `S2 R3` 虽然修正了尺度和主对象，整体仍受米色雾感、圆角小卡、弱边界和平均化信息块影响，没有达到参考 UI 的现代产品感。`S1 R6` / `S2 R4` 因此重点重读并转译 `ref-27`、`ref-25` 与 `ref-17`：

- 以近白产品窗口、清晰墨色文字、极细边界和少量低位阴影替代全局米色柔雾；家族色只服务身份、当前态、数据与必要状态。
- `S1` 桌面把 workspace pulse 做成非对称大尺度数据主面，并把 Operations Inbox 重组为任务列表与选中详情的连续双窗格；不再使用横向概览小卡和独立 continuation 小卡。
- `S2` 桌面把五阶段 rail、三项代表 contribution 和 readiness 详情组织为连续三窗格；当前 Human Promotion 是唯一任务主面，九组来源和十三项 contribution 仍来自现有 view model。
- `390x844` 不缩放桌面窗格，而是保留大标题、当前阶段和主任务列表，把来源详情与次级动作收入渐进披露。
- `ref-27` 只提供连续窗格、选中行与详情层次，`ref-25` 只提供白色抬升表面、细边界和宽松控件，`ref-17` 只提供非对称数据重心；客服、邀请、HR 业务语义和参考图原配色均不进入 RadishMind。

第六轮人工复评确认，`S1 R6` / `S2 R4` 的整体方向已接近参考 UI，但 Operations Inbox、Source evidence 与 Application Workspace 三窗格仍显得过于规整，桌面最外层 `24px` 留白、大圆角和阴影也形成了没有产品职责的宽外框。`S1 R7` / `S2 R5` 据此完成局部聚焦修正：

- 两个桌面根画板取消外圈留白、圆角窗口、外框与投影，导航和主工作面直接贴合 `1440x900` 视口；内容区域仍保留自身必要内边距，不把“全视口”误解成元素贴边。
- Operations Inbox 使用分段筛选容器、选中行墨蓝轨、实体图标面和更清晰的标题 / 描述 / 元数据层级；Source evidence 取消四个等权胶囊，改为 `01 / 04` 主覆盖数值、四段覆盖轨和细分状态行。
- Application Workspace 左栏改为带 `02 / 05` 位置与当前轨的 review path；中央 contribution 保持唯一主任务面；右栏把三个等权 KPI 改为 `5 / 9` 主覆盖指标和 blocked / missing 次级风险，并保持五个代表来源与“查看全部九组”入口。
- 窄屏信息结构未发生变化，只随设计包版本前移；本轮没有新增路由、动作、API、schema、生产声明或 React 实现。

第七轮人工复评认为，`S1 R7` / `S2 R5` 已消除桌面外框并建立正确连续窗格，但 Operations Inbox、Source evidence 与 readiness 仍偏向规整文字面板，尚未形成参考 UI 所体现的紧凑信息密度和清晰焦点。`S1 R8` / `S2 R6` 因此只在现有五个根画板内继续收敛：

- Operations Inbox 将当前窗口完整表达为四项紧凑注意队列，只为正在驱动右侧详情的条目保留中性柔底、墨蓝细轨和描边图标，并用三段 evidence path 填补详情区空白；没有扩展到全分页投影。
- Source evidence 使用 `01 Ready / 02 Partial / 01 Blocked` 分布和四来源矩阵，把覆盖关系从普通状态行提升为可扫读的数据构件；状态仍同时保留文字通道。
- Application Workspace 以柔底、墨蓝细轨和编号实体建立唯一当前阶段选中态；中央 contribution 保持中性状态列表，并使用十三段 contribution window 表达 `3 shown / 10 additional`，右栏用 `5 / 9` 九格覆盖矩阵和 authorization path 明确 human review 与 production closed 的关系。
- 两个桌面根画板继续贴合 `1440x900` 视口；窄屏结构不变，只随设计包版本前移。没有新增路由、动作、API、schema、生产声明或 React 实现。

同日针对高亮与列表协调的局部复核继续归入 `S1 R8` / `S2 R6`，不另增版本：选中态只表示当前正在驱动详情或导航的对象，`missing`、`blocked`、`partial` 等状态不自动获得选中高亮。`S1` 桌面与窄屏的首项 Inbox 因承担当前详情 owner，使用轻量选中轨；`S2` 只保留当前阶段的导航选中态，Application candidate 等 contribution 即使缺失也回归普通状态行。共享设计决策记录同步固定该规则。

真实 React 实现继续以现有 view model 和动作 owner 为功能真相：导航计数、workspace、active application、source coverage、Operations Inbox 与写入边界均消费当前代码数据；Pencil 中的代表性名称、数量和动作没有进入实现。既有次级页面保留在折叠的“More surfaces”入口，不以壳层迁移删除功能入口。

`S2 R6` 已完成以下设计收敛：

- 继承 `S1` 的唯一桌面导航和窄屏 command bar，不为 Application Workspace 建立第二套产品壳；
- 五阶段在桌面完整展示可进入性，在窄屏收敛为当前阶段选择器和五段位置 rail；阶段本身不显示完成勾选或自动晋级含义；
- 当前 Human Promotion surface 是桌面主视觉区域，owner evidence 以 Application Candidate、Workflow Definition 与 RAG authority 三条代表贡献说明精确 owner、状态和下一跳；
- 选中态遵循交互归属：当前阶段可以使用轻量导航选中轨，普通 contribution 只以图标、文字和状态标签表达 `available`、`missing` 或 `blocked`，不与导航争夺焦点；
- 九个来源组和十三项 contribution 仍以当前 TypeScript view model 为真相。桌面显式展示五个代表来源并保留“查看全部九组”入口；窄屏只展示 Application 与 RAG authority 两组最高优先状态，并提供同一渐进入口；
- 缺少权威 Application revision 保持 `partial`，未证明 RAG assignment 保持 `blocked`；readiness 明确为易失、只读、不可发布且不能满足 production authorization；
- `ref-05`、`ref-11`、`ref-12`、`ref-13`、`ref-14`、`ref-17`、`ref-18`、`ref-25` 与 `ref-27` 的分区、上下文、迁移、阶段、贡献、尺度和表面原则已经转译；主题切换、事故语境、ETA、聊天产品结构、邀请设置和自动发布均明确排除；
- `S1` 桌面 / 窄屏、`S2` 桌面 / 窄屏与共享设计决策记录的 Pencil 全树布局检查均无裁切、重叠或占位节点。
- `R1` 人工评审确认功能层级和停止线正确，但等权小卡、低对比和弱字号只形成通用管理页；`R2` 改为连续工作面；`R3` 放大主对象但仍保留柔雾小卡气质；`R4` 以近白连续窗格收敛现代产品感；`R5` 取消桌面外圈容器；`R6` 再以当前阶段轻量选中轨、九格 readiness、authorization path 与 contribution window 修正规整文字面板感，并把状态行与选中态彻底分离。

### 设计与实现闭环

1. 打开 Pencil 前先读取当前功能文档和代码，记录页面族、代码锚点、基线 commit、真实动作、状态、owner 与停止线。
2. 完成覆盖分级，只把 `A` 级页面族和 `B` 级新决策放入 Pencil；`C` 级页面不建稿。
3. Pencil 评审只确认设计决策、真实能力边界和 `ref-XX` 吸收原则，不把代表性文案当作功能真相。
4. 先实现设计基准面，再让同族页面直接消费已落地组件与模式；实现使用代码中的真实文案、条件和 API 状态。
5. 以运行中的 React 页面、桌面 / 窄屏截图、键盘路径、测试和 consumer smoke 完成最终复核。
6. 只有信息架构、交互模型、风险表达、共享组件解剖或响应式顺序变化时回写 Pencil；普通文案和功能字段变化只更新其真实 owner。

每个 Pencil 画板在产品化专题的覆盖记录中关联 `surface_id`、代码锚点和基线 commit。Pencil 与实现发生差异时，先判断变化属于设计决策还是功能事实：前者更新设计基准面，后者更新文档 / 代码并保留实现为真相，不机械复制回设计稿。

## 实施批次

### 基础批次：规范与 token 基座

状态：已完成。

交付：

- 初始引入 family-ui `v26.7.2` 的 token，并于 2026-08-03 升级为 `v26.7.3` 的 `tokens.css` 与 `tokens.json` 原样镜像。
- 在文档根节点启用 `data-rd-profile="workbench"`。
- 建立 `--rm-* -> --rd-*` L2 语义别名，并明确区分身份、操作、正向、类型与注意语义。
- 让 Web 的基础字体消费项目别名，不在同批执行全量配色或布局迁移。
- 建立项目差异附录，更新协作规范、文档入口、当前焦点和旧规范迁移状态。

### 设计批次：真实任务页面蓝图

参考图产品面映射、Pencil 协作模型、`S1 R8` 与 `S2 R6` 的桌面 / 窄屏设计、React 实现和真实浏览器验收均已完成。当前按既定顺序回到 `S3` Workflow Designer 对应功能设计文档核对画布交互和代码事实，不新增同层理想稿、兄弟画板或普通 UI 专项 task card。

### 实现批次：纵向切片迁移

按“产品壳 → Application Workspace → Saved Draft / Designer → API Integration / Key”的顺序逐批实施。每批同时完成：

- 对应页面和共享组件的语义 token 迁移；
- 桌面与窄屏行为；
- 键盘、焦点、状态第二通道和敏感信息边界；
- 单元测试、production build、真实浏览器路径和必要仓库门禁；
- 删除已被替代的遗留声明，避免新旧样式永久并存。

## 基础浏览器基线

- `1440x900`：真实页面 `scrollWidth` 与 viewport 均为 `1440px`，唯一桌面导航宽 `248px`；真实 view model 的五行 Operations Inbox 均为 `88px`，列表与只读 evidence detail 组成连续双窗格。
- `390x844`：闭合导航与展开菜单的页面 `scrollWidth`、body 宽度和菜单面板宽度均为 `390px`；五行 Operations Inbox 均为 `96px`，桌面详情渐进隐藏，`Open selected` 保留 owner 跳转。菜单导航后自动收起，目标锚点位于 `72px`。
- 响应式壳层在 `1101px` 保持 `248px` 桌面导航，在 `1100px` 切换 command bar；`1101px`、`1100px`、`821px` 与 `390px` 均无页面级横向溢出。
- 条件渲染的 API Key 与 Run History owner 不能只依赖静态 hash 滚动。实现会先让现有 application route owner 消费 `#workspace-api-keys` / `#workspace-run-history`，再等待真实目标挂载并滚动到 `72px`；没有新增路由、执行动作或伪造 owner。

## 基础批次验收

1. 上游 `tokens.css` 与 `tokens.json` 镜像内容和版本完全一致。
2. import 顺序固定为 family-ui L1 token、RadishMind L2 alias、现有组件样式。
3. HTML 使用 `zh-CN` 和 `data-rd-profile="workbench"`；构建后可计算到 Workbench 覆盖值。
4. 当前页面功能、路由、数据请求和生产边界不变。
5. Web 测试与 production build 通过，桌面和窄屏 smoke 无新增溢出或控制台错误。
6. 触及阶段与规范真相源，因此完成 fast baseline 后补跑全量 `check-repo`。

完成证据：

- 上游 `v26.7.3` CSS / JSON 镜像逐字比对通过，项目别名不再混用身份色与操作色。
- Web `272/272` 测试和 production build 通过。
- 基础批次 Playwright `1440x900` 与 `390x844` 复验完成；`S1 R8` 实现后应用内浏览器再次确认唯一 `248px` 桌面导航、`88px` / `96px` 任务行、窄屏菜单交互、条件渲染 owner 跳转和 `390px` 精确页面宽度，控制台零 warning / error。
- `v26.7.3` 对齐后应用内浏览器再次确认：灰玉 identity、墨蓝 action 与胭脂 attention 分工正确；`1440x900` 和 `390x844` 均无横向溢出；窄屏菜单关闭后目标标题在 `72px` 处保持可见；控制台零 warning / error。
- `S1 R8`、`S2 R6` 的桌面 / 窄屏与共享设计决策记录已显式保存到 Family UI Pencil 设计源；五个根画板全树布局检查无裁切、重叠或占位节点，并以 `2x` PNG 导出复核全视口桌面壳层、Operations Inbox 四项窗口与 evidence path、Source evidence 分布矩阵、九格 readiness、authorization path、窄屏渐进顺序、九组来源和 blocked / partial 停止线。
- `321e9899 feat(ui): 落地 S1 R8 产品壳` 已完成 Workbench 导航、workspace pulse、Source evidence、Operations Inbox 与响应式交互迁移；Web `272/272` 测试和 production build 通过，严格应用内浏览器复核覆盖桌面、临界宽度、窄屏菜单、API Key / Run History 跨阶段锚点和零 warning / error。
- `S2 R6` 已完成 Application Context、五阶段 review path、Human Promotion contribution 主面、九组来源 readiness、authorization path 和 owner surface 渐进展开；真实离线 view model 保持 revision `partial`、十三项 contribution、九组来源和不可发布边界。未证明的 RAG assignment 由 presentation 层根据现有缺失证据明确表达为 `blocked`，同时保留底层 owner rollup 的 `incomplete` 供审计，不改写四态 readiness 聚合。
- S2 应用内浏览器严格复核覆盖 `1440x900`、`1101/1100px`、`821px`、`761/760px` 与 `390x844`；所有宽度的页面 `scrollWidth` 均等于 viewport，窄屏三条 contribution 均为 `90px`，五阶段菜单、九来源展开与 owner review 开合正常，URL 无敏感材料，控制台零 warning / error。Pencil 的 `5 / 9` 保持代表数值，当前 fixture 实际显示 `1 / 9` owner references。
- `./scripts/check-repo.sh --fast` 与 `./scripts/check-repo.sh` 均通过；只保留 W28–W30 历史周志的既有篇幅 warning。

## 停止线

- 基础批次不做 `styles.css` 全量换色、组件重写或页面重新布局。
- 不因 family-ui 已提供暗色 token 就声明暗色主题切换可用。
- 不新增 API、schema、migration、repository、生产认证、quota / billing 或执行能力。
- 不把开发测试态 owner、离线 evidence、首分页窗口或人工审查画成生产就绪、全量统计、自动执行或业务写回。
- 不要求每个路由、组件、功能状态或相似页面都有独立 Pencil 画板，也不让 Pencil 取代功能文档与当前代码。
- 不为普通 UI 迁移新增 task card、fixture 或专项 checker；优先复用 Web 测试、build、consumer smoke 和仓库门禁。
