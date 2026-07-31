# RadishMind Family UI 产品化设计与迁移 v1

更新时间：2026-07-30

状态：`radishmind_family_ui_productization_v1_s1_implementation_completed`

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

通用 UI 规范遵循 RadishX 仓库 `docs/design/family-ui/` 的 `v26.7.2`：

- 家族通用原则、色彩、字体、间距、圆角、阴影、图标、组件、平台布局和迁移方式以上游规范为准。
- RadishMind 只在 [UI 差异附录](../../ui-addendum.md) 中维护项目差异、专属组件和暂存偏差。
- Web 构建中的上游 token 镜像必须与对应 family-ui 版本原样一致，不在镜像文件中直接修改项目差异。
- 新的家族通用 token 应先回到 family-ui 讨论；RadishMind 专属语义先使用 `--rm-*` L2 别名或项目组件变量表达。

## 产品 Profile

RadishMind 使用 `Workbench` Profile：

- 画面气质是现代、安静、紧凑、可审计的工程工作台。
- 亮色底采用家族纸色，但不使用品牌面纹样和装饰性印章。
- 主交互强调色使用墨蓝语义，成功与可继续状态使用玉色，紫色只用于小面积辅助区分。
- 颜色不是状态的唯一通道；状态必须同时具有文字、图标或结构语义。
- 信息密度允许高于品牌展示面，但主任务、当前上下文、高风险动作和失败原因必须有清晰层级。
- 页面、区域和卡片标题只显示稳定名称，不在标题上下追加 eyebrow、代表状态、营销句或临时解释句；状态、范围和来源数量进入 badge、字段或数据行。
- 常驻导航、普通主操作和边界摘要不使用大面积墨色反差块；默认用纸色层级、细边界与柔底语义色建立层次。
- 同一导航层级只有一个可见 owner；不得把图标产品轨与文字侧栏并列成两套主导航。跨产品域与当前产品内入口如果没有真实独立层级，统一收敛到一个侧栏。
- Workbench 列表按内容确定紧凑高度，不使用 `fill_container` 拉伸行高填满视口；`S1` 基线桌面任务行固定为 `68px`，窄屏三行任务条目固定为 `94px`。

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
| `radishmind_web_s1_product_shell_v1` | `implementation_completed_r4` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S1 Product Shell — Desktop / Partial · R4` 与 `S1 Product Shell — Narrow / Partial · R4` | 设计基准 commit `24655516`，实现 commit `88ec1107`；`apps/radishmind-web/src/app/ProductNavigation.tsx`、`apps/radishmind-web/src/app/App.tsx`、`apps/radishmind-web/src/features/control-plane-read/workspaceProductOverviewPanel.tsx`、`workspaceOperationsInboxPanel.tsx` 与 `apps/radishmind-web/src/styles.css` | 桌面只保留一个主导航侧栏，workspace / application 上下文持续可见，首屏以运营收件箱和来源覆盖回答“继续什么、为什么受限”；窄屏折叠侧栏并按任务顺序单列重排。代表性数量和文案不构成功能真相，不新增 production、全量统计、实时性、自动处置或 dark mode 声明。 |

`S1` 消费 `ref-03`、`ref-07`、`ref-08`、`ref-09`、`ref-15`、`ref-17`、`ref-18`、`ref-19` 与 `ref-24` 的层级、密度、上下文和状态表达；`ref-20` 只保留未来暗色证据，不在当前画板增加主题切换。两个画板均已通过 Pencil 全树布局检查，结果为 `No layout problems`；视觉复核确认没有裁切、重叠和横向溢出。

首轮视觉复核发现，仅使用纸色 token、柔底状态和标准“侧栏 + 等宽 KPI 卡 + 列表 + 右栏”骨架，虽然满足结构约束，却不足以体现参考图与 Radish 家族的可辨识气质。`R2` 据此引入双层产品导航、非均质 workspace pulse、表格式关注队列与更明确的对象语义。

第二轮视觉复核确认，`R2` 的墨色产品主轨、普通操作黑底和墨色停止线卡对比过高，自动生成的 eyebrow / 标题描述属于临时设计文案，弹性任务行也降低了信息密度。`R3` 将以下规则收敛为当前可实施基线：

- 桌面产品主轨和窄屏 command bar 改用纸色层级与柔底选中态，普通主操作与边界摘要不再使用墨色反差块；
- 页面、区域和卡片标题只保留名称；状态、范围、数量与限制进入结构化 badge、字段或数据行，不在标题下补解释性副文案；
- workspace pulse 继续使用非均质摘要，不退回四张同等权重指标卡；
- Operations Inbox 使用参考 `ref-08` / `ref-24` 的表格式关注队列、severity 筛选和软印色状态，同时保留 owner、当前窗口与 partial 停止线；
- 小面积胭脂只用于家族 seal 与关注计数，墨蓝负责交互，玉色和赭色只表达状态，雅紫只区分 application / workflow 对象；
- 桌面关注队列行高固定为 `68px`，窄屏三行任务条目固定为 `94px`；列表不再随剩余视口高度拉伸；
- 窄屏复用柔色 command bar、workspace pulse 和对象卡语法，按任务顺序单列重组，不缩小桌面导航和 KPI。

第三轮导航复核确认，`R3` 的图标产品轨与文字对象侧栏虽然配色已降低对比，但仍被感知为两套并列工具栏，且当前功能并不存在需要双轨表达的独立导航层级。`R4` 删除图标产品轨，把品牌、workspace 切换、一级入口、环境边界、账户和帮助统一归入一个桌面侧栏；窄屏继续使用单一 command bar 与折叠导航。后续只有存在真实、稳定且可独立切换的跨产品层级时，才重新评估产品轨，不为视觉效果预留第二套导航。

真实 React 实现继续以现有 view model 和动作 owner 为功能真相：导航计数、workspace、active application、source coverage、Operations Inbox 与写入边界均消费当前代码数据；Pencil 中的代表性名称、数量和动作没有进入实现。既有次级页面保留在折叠的“More surfaces”入口，不以壳层迁移删除功能入口。

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

- 引入 family-ui `v26.7.2` 的 `tokens.css` 与 `tokens.json` 原样镜像。
- 在文档根节点启用 `data-rd-profile="workbench"`。
- 建立 `--rm-* -> --rd-*` L2 兼容别名。
- 让 Web 的基础字体消费项目别名，不在同批执行全量配色或布局迁移。
- 建立项目差异附录，更新协作规范、文档入口、当前焦点和旧规范迁移状态。

### 设计批次：真实任务页面蓝图

参考图产品面映射、Pencil 协作模型与 `S1` 产品壳设计、实现和浏览器验收均已完成。下一步进入 `S2` Application Workspace `A` 级设计基准面，先以一个真实 Application 连续任务固定五阶段工作区、当前 surface、evidence / readiness 与关键 blocked / partial 表达。后续只为 `A` 级页面族或 `B` 级新决策扩稿。设计评审必须映射到 `ref-XX`、代码基线、现有 owner、consumer 和停止线，不能用静态理想稿代替真实交互状态。

### 实现批次：纵向切片迁移

按“产品壳 → Application Workspace → Saved Draft / Designer → API Integration / Key”的顺序逐批实施。每批同时完成：

- 对应页面和共享组件的语义 token 迁移；
- 桌面与窄屏行为；
- 键盘、焦点、状态第二通道和敏感信息边界；
- 单元测试、production build、真实浏览器路径和必要仓库门禁；
- 删除已被替代的遗留声明，避免新旧样式永久并存。

## 基础浏览器基线

- `1440x900`：真实页面 `scrollWidth` 与 viewport 均为 `1440px`，唯一桌面导航宽 `248px`，Operations Inbox 四行均为 `68px`，首屏概览与 inbox 在 `900px` 高度内完成展示。
- `390x844`：闭合导航与展开菜单的页面 `scrollWidth`、body 宽度和菜单面板宽度均为 `390px`；四行 Operations Inbox 均为 `94px`，菜单导航后自动收起并把目标锚点滚动到视口顶部。
- 既有 `421px` 超宽的真实根因是 Application Configuration Draft 深层标题行中的长状态 badge。实现批次通过移动端共享收缩规则、标题行换行和 React Flow 边界收口消除该偏差；应用内浏览器复验控制台无 warning / error。

## 基础批次验收

1. 上游 `tokens.css` 与 `tokens.json` 镜像内容和版本完全一致。
2. import 顺序固定为 family-ui L1 token、RadishMind L2 alias、现有组件样式。
3. HTML 使用 `zh-CN` 和 `data-rd-profile="workbench"`；构建后可计算到 Workbench 覆盖值。
4. 当前页面功能、路由、数据请求和生产边界不变。
5. Web 测试与 production build 通过，桌面和窄屏 smoke 无新增溢出或控制台错误。
6. 触及阶段与规范真相源，因此完成 fast baseline 后补跑全量 `check-repo`。

完成证据：

- 上游 CSS / JSON 镜像逐字比对通过。
- Web `272/272` 测试和 production build 通过。
- 基础批次 Playwright `1440x900` 与 `390x844` 复验完成；`S1` 实现后应用内浏览器再次确认唯一 `248px` 桌面导航、`68px` / `94px` 任务行、窄屏菜单交互和 `390px` 精确页面宽度，控制台零 warning / error。
- `./scripts/check-repo.sh --fast` 与 `./scripts/check-repo.sh` 均通过；只保留 W28–W30 历史周志的既有篇幅 warning。

## 停止线

- 基础批次不做 `styles.css` 全量换色、组件重写或页面重新布局。
- 不因 family-ui 已提供暗色 token 就声明暗色主题切换可用。
- 不新增 API、schema、migration、repository、生产认证、quota / billing 或执行能力。
- 不把开发测试态 owner、离线 evidence、首分页窗口或人工审查画成生产就绪、全量统计、自动执行或业务写回。
- 不要求每个路由、组件、功能状态或相似页面都有独立 Pencil 画板，也不让 Pencil 取代功能文档与当前代码。
- 不为普通 UI 迁移新增 task card、fixture 或专项 checker；优先复用 Web 测试、build、consumer smoke 和仓库门禁。
