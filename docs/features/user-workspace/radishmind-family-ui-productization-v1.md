# RadishMind Family UI 产品化设计与迁移 v1

更新时间：2026-08-13

状态：`radishmind_family_ui_productization_v1_s9_s10_visual_r3_migrations_completed`

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
5. Application Runtime Review：受控调用、精确请求审查、应用当前窗口运行证据和跨 owner 交接。
6. Workflow Run & Evaluation Review：运行定位、兼容比较、版本化 Case / Suite 证据和摘要绑定的人工判断。
7. Admin Control Plane：管理上下文、七类资源定位、Tenant / Audit 只读证据、身份缺口和开发测试态 Provider / Profile / Route 受控配置。
8. Prompt / Agent / Copilot 类型工作区：类型源码、配置、候选、assignment、Access、受控调用和评测交接的单任务路径。
9. Admin Quota Admission：精确 application policy、quota-owner UTC usage、expected-version CAS 确认与稳定失败关闭。

### 后续设计面

- `S9 Admin Quota Admission` 已由[应用 API Key 请求配额与 Provider Attempt 准入专题](../gateway/application-api-key-request-quota-admission-dev-test-v1.md)和真实后端 owner 产生，并按五维评分 `1 / 2 / 2 / 2 / 1 = 8` 完成功能纵向切片；原 Pencil R1 的 token-only 修正没有解决页面骨架偏离，Visual R2 又因硬方形表面没有继承 S7 / S8 形态语言被退回。Desktop `C7pkb`、Narrow `x8lESc` 与 Decision R14 `tCWCW` 已修订为 Visual R3，于 2026-08-12 人工复核通过，并在 2026-08-13 完成 React 迁移与真实浏览器连续链。
- `S10 Application Evaluation Campaign` 已按五维评分 `8`、覆盖级别 `A / 完整 Pencil` 完成功能纵向切片、strict consumer、memory / SQLite exact handoff 和服务重启恢复；Visual R2 的连续 Workbench 骨架保留，但硬方形业务表面已在 Desktop `Um8Zh`、Narrow `ZxJd7` 与 Decision R15 `UNMOS` 修订为 Visual R3。2026-08-13 React 又完成 selected campaign context、campaign 主 owner、连续 evidence rows、单一 Handoff rail 与职责圆角迁移，并通过真实浏览器三视口复核。
- Provider 价格策略与应用成本审查没有建立 S11，而是在 S7 / S5 页面族新增 Desktop `wQ2t0` / `Ue7hq`、Narrow `Z5Iqv` / `i50xIV` 与 Decision R16 `VAToA`。Visual R1、React strict consumer、SQLite 重启和 `1440×900` / `720×900` / `390×844` 真实浏览器均已完成。
- Gateway Provider Attempt 受控降级同样不建立 S11，而是在 S7 Route 新增 Desktop `h41DNz` / Narrow `Q5dMjv`，在 S5 Playground 新增 Desktop `DY5HB` / Narrow `o9Btk`，在 S5 Request History 新增 Desktop `KsXpp` / Narrow `BRzOE`，并以 Decision R17 `GfqT6` 统一冻结六类代表状态。七个 Visual R1 根节点已完成结构与实际渲染复核，并于 2026-08-13 获得人工视觉批准；React 留到后续实现批次。
- 其它后续设计面必须由新的功能专题和真实使用证据产生；S1–S10 不原地派生同层页面链。

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
- 顶层 Pencil 基准面默认按产品审阅顺序横向排放：同一 surface 的 Desktop 与 Narrow 相邻，阶段 Decision 紧随对应 surface；不得再把新增基准面纵向堆叠或另起孤立画布列。该规则只约束无限画布中的审阅排布，不改变页面自身的响应式方向。
- 表面形态必须同时继承已审页面和 family-ui shape 规范：全视口壳、连续窗格、分隔线与表格事实行保持方正或发丝边界；任务、上下文、owner、boundary 等业务表面使用 `8–11px` 职责圆角，紧凑控件使用 `7–8px`，状态标签使用全圆角。禁止把所有区域统一做成硬方块，也禁止把全部容器无差别胶囊化。
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
| `S5` Application Runtime Review | `A` | 受控调用、脱敏请求审查与当前窗口证据的持续任务面；补充窄屏单 owner 顺序 | 复用既有三类 consumer，不复制执行、history 或 operations 真相源 |
| `S6` Workflow Run & Evaluation Review | `A` | 运行定位、兼容比较、exact-version Case / Suite 与 digest-bound 人工 decision；补充窄屏四任务顺序 | 复用既有五类 consumer，不复制 run、comparison、evaluation 或 release truth source |
| `S7` Admin Control Plane | `A` | 管理上下文、Tenant / User / Role / Audit / Provider / Profile / Route 七资源拓扑、单 owner 和窄屏顺序 | 五维评分为 `2 / 2 / 2 / 2 / 2`；Tenant / Audit 复用 authenticated read，User / Role 显式失败关闭，Provider / Profile / Route 复用同一原子配置 owner |
| `S8` Prompt / Agent / Copilot 类型工作区 | `A` | 类型上下文、七 / 八任务拓扑、单 owner、易失输入输出与人工评测交接；补窄屏任务先于 owner 的顺序 | 五维评分为 `1 / 2 / 2 / 2 / 2`；复用 Template / Profile、Configuration、Candidate、Assignment、Access、Session / Invocation 与 S6 owner |
| `S9` Admin Quota Admission | `A` | application policy、quota-owner UTC used / remaining、CAS 更新确认、稳定 blocked reason 与窄屏 context → path → owner → boundary 顺序 | 五维评分为 `1 / 2 / 2 / 2 / 1`；Desktop / Narrow Visual R3 与 Decision R14 已人工通过，React 已完成 selected application context、单一 quota owner、连续 policy rows、admission rail 与职责圆角迁移；不并入 Provider / Profile / Route，不冒充 production quota / billing |
| `S10` Application Evaluation Campaign | `A` | Plan、Campaign、Pair Review、Handoff、Desktop / Narrow 与共享 Decision R15 | 五维评分为 `1 / 2 / 2 / 2 / 1`；Desktop / Narrow Visual R3 与 Decision R15 已人工通过，React 已完成 selected campaign context、campaign 主 owner、连续 evidence rows、单一 Handoff rail 与职责圆角迁移；memory / SQLite exact handoff 与服务重启证据继续成立 |

当前维护十个已完成设计基准面和必要的局部 / 状态变体；新的页面族必须重新评分，不按路由、组件或状态数量扩张画板。

2026-08-10 的 Workflow Definition 结构化运行输入五维评分为 `0 / 1 / 1 / 1 / 2 = 5`，采用 `B / 局部 Pencil`：Desktop `W3O4tV` 与 Narrow `t39foq` 的首版 R1 因继承 S9 之后的独立表单看板语言被人工退回，Visual R2 因形态语言过硬再次退回；Visual R3 虽恢复职责圆角，字段列表仍像整页被硬分割成多行。Visual R4 已改为有留白层级的表单画布：Desktop 使用非对称双列字段组，Narrow 按长字段、相关短字段组、补充字段渐进重排，错误贴近精确输入，布尔值呈现为开关。Visual R4 已通过人工复核并冻结为 Session v4 与共享 React editor 的实现前置。它不是新的页面族，不建立 S11，也不回写或重画 S2 / S3 / S10 完整基准面。

2026-08-11 React 接入与真实浏览器复核已完成：Definition Direct Run 与 Application Interaction Session 复用同一 editor；Definition owner 从历史深色容器收敛为 Workbench 纸色表面，editor / Session 统一消费 `--rm-*` 语义 token。`1440×900`、`1100×900`、`721×900`、`720×900` 与 `390×844` 均无横向溢出，`720px` 断点从双列合同摘要和横向字段标签切为单列；Desktop 与窄屏实际截图确认字段、帮助、布尔选择和易失值边界可读，最终 console 无 warning / error。该结论只冻结结构化输入局部模式，不表示 S9 / S10 React 已采用 Visual R3，也不建立 S11。

2026-08-12 的 Workflow RAG 本地知识材料导入、审查与快照构建五维评分为 `1 / 1 / 1 / 1 / 1 = 5`，采用 `B / 局部 Pencil`。Desktop `U4tmEg` 与 Narrow `nI3RW` 已在同一 RadishMind 设计源中冻结并由人工通过；它们只改造 S2 内既有知识快照 owner，覆盖本地文件选择、来源 / 预算 findings、fragment list、当前 fragment inspector、提交摘要与窄屏渐进顺序，不建立 S11 或重画完整 S2。React 已按该局部基线完成单一结构化 editor，真实浏览器覆盖 `1440×900`、`720×900`、`390×844`；390px 初检发现 panel 内部 record identity 溢出，修正后 document 与 panel 均无横向溢出。create / version、CAS、SQLite 重启与隐私边界继续以真实 API record 为事实源，解析失败、duplicate、secret 和 version conflict 不冒充选中态。

2026-08-12 新选择的 Provider 价格策略版本与应用成本审查五维评分为 `1 / 2 / 2 / 2 / 2 = 9`，采用 `A / 完整 Pencil`。它只在已审 S7 增加 Pricing task、在已审 S5 增加 Request / Evidence 成本审查，冻结 CAS 更新、历史不可重算、六态 availability、当前窗口 coverage 与窄屏顺序，不建立 S11 或重画完整 S5 / S7。S9 Desktop `C7pkb` / Narrow `x8lESc` 与 S10 Desktop `Um8Zh` / Narrow `ZxJd7` 的 Visual R3 已在同日完成人工复核并通过。随后新增的 S7 Pricing Desktop `wQ2t0` / Narrow `Z5Iqv`、S5 Cost Review Desktop `Ue7hq` / Narrow `i50xIV` 与 Decision R16 `VAToA` 已形成 Visual R1：桌面冻结 review-ready 与当前窗口双栏证据，窄屏冻结 CAS conflict 与六态 coverage；五张画板全树无裁切、越界或 placeholder，并已完成人工视觉复核。

同轮人工反馈确认画布治理、信息架构与表面形态是三个独立维度。原有 26 个顶层基准面保持 S1 → S10 单行审阅带；后续 Workflow RAG 局部稿与价格专题把设计源推进到 33 个顶层基准面，Provider Attempt 又继续增加六张 S7 / S5 代表面与 R17，设计源现有 40 个顶层基准面。第一次只映射 token、第二次 Visual R2 只归位页面骨架，仍把上下文、任务、owner 与 boundary 做成同一种硬方块，因此两轮均被人工明确退回。Visual R3 保留 S8 母版的 `264px` 产品导航、薄页眉、对象路径、单一主 owner、连续事实行与一个辅助 rail，并依据 S7 / S8 和 `reference-ui` 恢复职责圆角；该规则继续作为 S9 / S10 与后续 Visual R1 的基线。结构化输入 Visual R4 进一步区分“连续事实数据”和“待编辑值”：前者可保留方正发丝边界，后者必须呈现真实控件、字段关系与局部错误，不能以表格代理表单。

2026-08-13 真实页面对照审计冻结并完成了两个独立迁移批次。S9 在 React 落实 selected application context、主 quota owner、连续 policy rows、remaining / CAS 辅助 rail、职责圆角和窄屏 context → owner → boundary 顺序；SQLite 本地产品完成 missing → create v1 → 双标签 stale CAS → reload v2。S10 随后把平均化 Campaign 列表 / 详情收敛为 selected campaign context、campaign 主 owner、连续 item evidence rows 和单一 Handoff rail；Plan 继续承担上游选择，不复制 campaign owner。S10 真实数据切换和 exact Handoff 正常，`1440×900`、`720×900`、`390×844` 均无横向溢出且控制台无 warning / error。两批均未改 Pencil、API、schema、repository、permission 或 production 边界。

2026-08-09 的 Workflow Definition 真实连续链复验五维评分为 `0 / 0 / 0 / 1 / 1 = 2`，采用 `C / 直接实现`：只补 RAG 专属 owner 交接、Definition 资格失败关闭、拓扑派生和 application 切换后的迟到 evidence 拒绝，继续复用 S2 / S3 / S6 已冻结的信息层级、任务轨、选中语义和响应式顺序。Pencil 当时正被其它项目占用，本批没有读取或修改设计源，也没有建立第十个页面族。

同日的 Workflow RAG Promotion → Configuration Draft 交接复验同样为 `0 / 0 / 0 / 1 / 1 = 2`，采用 `C / 直接实现`：复用 S2 已冻结的阶段切换、单 owner 打开和易失单引用交接，配置页复用既有 binding selector、显式恢复与失败关闭表达。`1440×900`、`900×900`、`720×900`、`390×844` 保持 context → task → owner 顺序、零横向溢出和控制台零 warning / error；本批没有新布局、交互模型或响应式策略，因此未操作正被其它项目占用的 Pencil，也没有建立第十个页面族。

2026-08-17 的 Application Result Workspace 五维评分为 `0 / 0 / 1 / 1 / 0 = 2`，采用 `C / 直接实现`：复用 S5 的 Application Context、单 owner task path 与 Run handoff，复用 S3 Saved Draft Library 的筛选列表和 Session Result Artifact Panel 的 exact inspector / lifecycle。显式 JSON 导出只增加可逆的开发测试态 digest 重校验说明，没有形成新页面拓扑、高风险确认或跨页面语义，因此未修改 Pencil、未建立 S11。真实浏览器 `1440×900`、`720×900`、`390×844` 均保持单一选中任务、零横向溢出和控制台零 warning / error。

### 设计基准面覆盖记录

| `surface_id` | 状态 | Pencil 覆盖 | 代码基线与锚点 | 设计决策与停止线 |
| --- | --- | --- | --- | --- |
| `radishmind_web_s1_product_shell_v1` | `implementation_r8_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S1 Product Shell — Desktop / Partial · R8` 与 `S1 Product Shell — Narrow / Partial · R8` | `R8` 设计基准 commit `1c537537`，实现 commit `321e9899`；`apps/radishmind-web/src/app/ProductNavigation.tsx`、`apps/radishmind-web/src/app/App.tsx`、`apps/radishmind-web/src/features/control-plane-read/workspaceProductOverviewPanel.tsx`、`workspaceOperationsInboxPanel.tsx` 与 `apps/radishmind-web/src/styles.css` | `R8` 已落地唯一导航、workspace / application 上下文、全视口桌面壳、非对称 pulse、来源分布矩阵、连续 Inbox 列表 / 详情、轻量选中轨和 evidence path。Pencil 的四项队列与 `1 / 2 / 1` 只表示设计基准；真实离线 view model 在验收时投影五项队列与 `1 / 3 / 0` 来源分布，React 未伪造计数或改写 partial 停止线。 |
| `radishmind_web_s2_application_workspace_v1` | `implementation_r6_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S2 Application Workspace — Desktop / Partial · R6`、`S2 Application Workspace — Narrow / Partial · R6` 与 `S1 + S2 Visual Language — Design Decision Record · R6` | 代码审计基线 `a545f511`，实现于 2026-08-08；`applicationDevelopmentWorkspacePanel.tsx`、`applicationDevelopmentWorkspaceSurface.tsx`、`applicationDevelopmentWorkspace.ts`、`applicationDevelopmentWorkspaceRoute.ts`、`applicationDevelopmentReadiness.ts`、`applicationDevelopmentWorkspaceControls.ts`、`ProductNavigation.tsx` 与 `styles.css` | `R6` 已在共享壳层内落地五阶段、owner contribution 与 readiness 投影；Applications 和当前阶段分别形成唯一产品导航 / 阶段选中态，普通 contribution 始终保持中性。桌面使用 `3 / 13` 窗口轨、九格来源矩阵和 authorization path；窄屏按 Application Context、当前阶段选择器、三项贡献、readiness 摘要渐进重排。React 继续使用真实 view model，当前离线 fixture 为 `1 / 9` owner references，不复制 Pencil 的代表性 `5 / 9`。 |
| `radishmind_web_s3_workflow_designer_v1` | `implementation_r2_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S3 Workflow Designer — Desktop / Partial · R2`、`S3 Workflow Designer — Narrow / Partial · R2` 与 `S1 + S2 + S3 Visual Language — Design Decision Record · R8`；三者实际渲染无裁切，`R1` 因人工评审退回，不再作为验收基准 | `App.tsx` 继续持有草案、consumer、冲突和 mutation owner，`workflowDraftDesignerPanel.tsx` 承担 Workbench 展示，`workflowNodeDesigner.tsx`、`ProductNavigation.tsx`、`productNavigationRoute.ts` 与 `styles.css` 承担画布、导航和响应式语义 | `R1` 虽通过行为测试与页面级溢出检查，但存在确定性画布裁切、节点计数与可见节点错位、连线不可读、底部三块 review 等权铺陈，以及窄屏节点裁切与重复 Inspector 层级，因此被人工退回。`R2` 已完成当前节点 / 直接邻居默认焦点、显式 `Fit graph`、紧凑 review，以及 `1440px` 右侧 / `<=1380px` 下移 / `<=760px` 折叠的 Inspector。强选中只属于当前导航或驱动 Inspector 的当前对象，finding 与 readiness / lifecycle 状态保持独立。 |
| `radishmind_web_s3_saved_draft_library_v1` | `b_level_implementation_completed` | `S3 Workflow Designer R2` 只保留紧凑活动草案引用、当前草案与完整 Library 交接，不建立第二个重复页面 | `workflowUserWorkspaceHomePanel.tsx`、saved draft list / lifecycle consumer 与 `App.tsx` 的精确打开、归档、恢复 owner；本批补齐 action-token tab 与 `tablist` / `tab` / `aria-selected` 语义 | 完整 Saved Draft Library 保持唯一挂载；Designer 只显示当前 / 活动草案引用和精确打开入口。真实浏览器已复验保存、读取、派生、两步归档、归档只读、解除归档和活动列表重新打开；S3 R2 不改变该 owner，也不新增自动打开、自动覆盖、批量 lifecycle 或生产 repository。 |
| `radishmind_web_s4_application_access_v1` | `implementation_r1_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S4 Application Access — Desktop / Dev Test · R1`、`S4 Application Access — Narrow / Dev Test · R1` 与 `S1 + S2 + S3 + S4 Visual Language — Design Decision Record · R9`；三者全树无裁切、越界或占位节点 | `applicationDevelopmentWorkspaceSurface.tsx` 编排单一 S4 task owner，`applicationApiIntegrationPanel.tsx`、`apiKeyLifecyclePanel.tsx`、`ProductNavigation.tsx`、`productNavigationRoute.ts` 与 `styles.css` 继续消费既有 integration、lifecycle、rotation、handoff 和 Playground owner | 接入、凭据、验证和退役按一个任务轨渐进呈现；普通 Key 行保持中性，只有驱动详情的行使用墨蓝细轨，active / expired / revoked / verification pending 等状态不承担选中语义。模型只开放实际声明的协议；七项 scope、一次性凭据和 `last_used_at` 验证门槛保持真实。archived application 只允许 Key metadata / detail / revoke，继续阻断 issue、rotate、integration 和 invocation；offline workspace summary 与当前应用 lifecycle 列表不合并。 |
| `radishmind_web_s5_application_runtime_review_v1` | `implementation_r1_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S5 Application Runtime Review — Desktop / Dev Test · R1`、`S5 Application Runtime Review — Narrow / Dev Test · R1` 与 `S1 + S2 + S3 + S4 + S5 Visual Language — Design Decision Record · R10`；三者全树无裁切、越界或占位节点 | `applicationRuntimeReviewWorkspace.tsx` 在 Application Workspace 后持续挂载 S5 上下文与任务轨，`modelGatewayPlaygroundPanel.tsx`、`modelGatewayRequestHistoryPanel.tsx`、`applicationOperationsPanel.tsx` 仍分别持有既有 owner；`applicationDevelopmentWorkspacePanel.tsx`、`App.tsx` 与 `styles.css` 只负责组合和响应式语义 | Run、Request、Evidence 三个任务只挂载一个当前 owner。Playground 使用目录真实模型 / 协议资格且只保存易失结果；Request History 在切换 application / workspace 时先清空并拒绝迟到响应；Operations 只汇总已加载的 Gateway / Workflow 当前窗口，不推测跨来源关联。普通行保持中性，任务 / 详情选中使用墨蓝细轨，状态继续使用独立文字和 badge。archived application 阻断 Run，history / evidence 保持真实可读边界。 |
| `radishmind_web_s6_workflow_run_evaluation_review_v1` | `implementation_r1_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S6 Workflow Run & Evaluation Review — Desktop / Dev Test · R1`、`S6 Workflow Run & Evaluation Review — Narrow / Dev Test · R1` 与 `S1 + S2 + S3 + S4 + S5 + S6 Visual Language — Design Decision Record · R11`；三者全树无裁切、越界或占位节点 | `workflowReviewWorkspace.tsx` 持续挂载 S6 上下文与四任务轨，`workflowReviewOwner.tsx`、`workflowRunComparisonPanel.tsx`、`workflowEvaluationPanel.tsx`、`workflowEvaluationSuitePanel.tsx` 继续消费既有 run、comparison、case、suite 与 decision consumer；`ProductNavigation.tsx`、`productNavigationRoute.ts`、`App.tsx` 与 `styles.css` 负责精确导航、组合和响应式语义 | Runs、Compare、Cases、Release 四个任务只挂载一个当前 owner。Run 只表示当前 cursor window；Comparison 即时派生且不持久化；Case / Suite 使用 exact refs；decision 绑定 review digest，`approved` 只形成 append-only evidence。workspace mismatch 零请求失败关闭，切换上下文先清空并拒绝迟到响应，archived application 保留历史只读但关闭诊断与 mutation。 |
| `radishmind_web_s7_admin_control_plane_v1` | `implementation_r1_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S7 Admin Control Plane — Desktop / Dev Test · R1`、`S7 Admin Control Plane — Narrow / Dev Test · R1` 与 `S1 + S2 + S3 + S4 + S5 + S6 + S7 Visual Language — Design Decision Record · R12`；三者全树无裁切、越界或占位节点 | `adminControlPlaneWorkspace.tsx` 持续挂载 context、七资源任务轨与单 owner，`adminControlPlaneRoute.ts` 固定精确 hash；`devLiveReadConsumer.ts`、`adminProviderRouteWorkspacePanel.tsx`、`ProductNavigation.tsx`、`App.tsx` 与 `styles.css` 复用既有 consumer、配置 owner 和 Workbench 壳层 | Tenant 只显示脱敏 summary；Audit 保持严格 cursor current window，行选中只驱动只读详情；User / Role 没有 RadishMind list owner，明确阻断邀请、角色变更与 production session。Provider / Profile / Route 是同一 `tenant_ref + workspace_id + environment + configuration_id` 原子 owner 的三个任务入口，不拆成伪 CRUD，也不复制 runtime inventory、credential、endpoint 或生产启用能力。 |
| `radishmind_web_s8_prompt_agent_type_workspace_v1` | `implementation_r1_completed` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 `S8 Prompt / Agent Type Workspace — Desktop / Dev Test · R1`、`S8 Prompt / Agent Type Workspace — Narrow / Dev Test · R1` 与 `S1 + S2 + S3 + S4 + S5 + S6 + S7 + S8 Visual Language — Design Decision Record · R13`；三者全树无裁切、越界或占位节点 | `promptAgentTypeWorkspace.tsx` 与 `promptAgentTypeWorkspaceModel.ts` 负责编排；Template / Profile、Configuration、Candidate、Prompt / Agent Assignment、API / Key、Invocation / Session 与 S6 既有 panel / consumer 继续持有领域真相；`applicationDevelopmentWorkspaceSurface.tsx` 只切换到一个当前 owner，`App.tsx` 在 application selection host 规范化跨类型任务 anchor | Prompt 按八任务、Agent 按七任务组织真实能力。Candidate 与 Assignment 分开定位但不复制状态；application kind 切换只替换上一类型的精确 S8 anchor，不折叠共享或 S6 深层 owner。输入 / 输出保持易失，archived source / governance 只读，controlled use 阻塞；approved、active 或 successful 均不表示 production release / enablement。 |
| `radishmind_web_s9_admin_quota_admission_v1` | `implementation_visual_r3_migration_completed_browser_verified` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 Desktop `C7pkb`、Narrow `x8lESc` 与 Decision R14 `tCWCW` 已为 Visual R3；三者全树无折叠、硬编码色、越界或占位节点，2026-08-12 人工复核通过 | `adminControlPlaneWorkspace.tsx` 继续持有第八项 Quota task；`adminGatewayRequestQuotaConsumer.ts` 契约不变，`adminGatewayRequestQuotaPanel.tsx` 与 `styles.css` 已在 2026-08-13 完成 selected application context、主 owner、连续事实行、admission rail 与职责圆角迁移，并通过 SQLite CAS 连续链和三视口复核 | Policy 与 usage 只接受 `tenant_ref + workspace_id + environment + application_id` 精确 owner；used / remaining 不从 Request History 或旧 `QuotaSummary` 推算。Visual R3 迁移不改变 CAS、权限、环境或生产停止线。 |
| `radishmind_web_s10_application_evaluation_campaign_v1` | `implementation_visual_r3_migration_completed_browser_verified` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 Desktop `Um8Zh`、Narrow `ZxJd7` 与 Decision R15 `UNMOS` 已为 Visual R3；三者全树无折叠、硬编码色、越界或占位节点，2026-08-12 人工复核通过 | 后端 A 至 D、React strict consumer、memory / SQLite exact handoff 与重启恢复继续成立；`applicationEvaluationCampaignPanel.tsx` 与 `styles.css` 已在 2026-08-13 完成 selected campaign context、campaign 主 owner、连续 evidence rows、单一 Handoff rail 与职责圆角迁移，并通过真实数据切换、exact Handoff 和三视口复核 | Plan、Campaign、Pair Review 与 Handoff 继续共用一个当前 owner；failed、quota、interrupted、partial 与 blocked 不冒充选中或 production readiness。 |
| `radishmind_web_s7_pricing_s5_cost_review_v1` | `implementation_r1_completed_browser_verified` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 S7 Desktop `wQ2t0` / Narrow `Z5Iqv`、S5 Desktop `Ue7hq` / Narrow `i50xIV` 与 Decision R16 `VAToA`；五张画板无裁切、越界或占位节点并已人工通过 | 后端 pricing owner、Admin API、Request History v2、请求级成本快照、React strict consumer、双数据库、SQLite 重启和 `1440×900` / `720×900` / `390×844` 浏览器连续链已完成 | S7 只管理精确 Provider / Profile / Model 的不可变价格 revision 与 CAS；S5 只审查当前加载 Gateway 窗口的六态 coverage 和单条请求谱系。历史不重算、`has_more` 不冒充全历史、Workflow 不与 Gateway 金额相加。 |
| `radishmind_web_s7_s5_provider_attempt_v1` | `visual_r1_approved_implementation_next` | [Family UI v1](../../designs/radishmind-web-family-ui-v1.pen) 中的 S7 Route Desktop `h41DNz` / Narrow `Q5dMjv`、S5 Playground Desktop `DY5HB` / Narrow `o9Btk`、S5 Request History Desktop `KsXpp` / Narrow `BRzOE` 与 Decision R17 `GfqT6`；七张画板无裁切、越界或占位节点，2026-08-13 人工视觉批准 | React strict consumer 尚未开始；当前实现仍是批次 D 的 unary executor、Route v2 与 Request History v3 后端事实 | Route activation 与请求级 opt-in 双门禁；stream 保持 single attempt；History 区分根 selection、terminal target、逐 attempt quota 与 partial cost，且不显示敏感 Provider 材料。 |

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

参考图产品面映射、Pencil 协作模型，以及 `S1 R8`、`S2 R6`、`S3 R2`、`S4 R1`、`S5 R1`、`S6 R1`、`S7 R1`、`S8 R1` 的桌面 / 窄屏设计和 React 实现均已完成；`S9`、`S10` 的 React 功能实现、Pencil Visual R3、React 迁移与真实浏览器复核也已完成。价格专题的 S7 / S5 Visual R1、React strict consumer、双数据库和真实浏览器连续链同样完成。`S3 R1` 因确定性画布裁切与层级重复被人工退回，S9 / S10 R1 因页面骨架偏离被退回、Visual R2 因表面形态偏离被退回的事实均保留。价格 Visual R1 只扩开发测试态价格证据，不扩 production membership、OIDC、secret、自动路由、billing 或生产启用。

### 实现批次：纵向切片迁移

按“产品壳 → Application Workspace → Saved Draft / Designer → API Integration / Key → Application Runtime Review → Workflow Run & Evaluation Review → Admin Control Plane → Prompt / Agent Type Workspace → Admin Quota Admission”的顺序逐批实施。每批同时完成：

- 对应页面和共享组件的语义 token 迁移；
- 桌面与窄屏行为；
- 键盘、焦点、状态第二通道和敏感信息边界；
- 单元测试、production build、真实浏览器路径和必要仓库门禁；
- 删除已被替代的遗留声明，避免新旧样式永久并存。

## 基础浏览器基线

- `1440x900`：真实页面 `scrollWidth` 与 viewport 均为 `1440px`，唯一桌面导航宽 `248px`；真实 view model 的五行 Operations Inbox 均为 `88px`，列表与只读 evidence detail 组成连续双窗格。
- `390x844`：闭合导航与展开菜单的页面 `scrollWidth`、body 宽度和菜单面板宽度均为 `390px`；五行 Operations Inbox 均为 `96px`，桌面详情渐进隐藏，`Open selected` 保留 owner 跳转。菜单导航后自动收起，目标锚点位于 `72px`。
- 响应式壳层在 `1101px` 保持 `248px` 桌面导航，在 `1100px` 切换 command bar；`1101px`、`1100px`、`821px` 与 `390px` 均无页面级横向溢出。
- 条件渲染的 API Key owner 不能只依赖静态 hash 滚动；S6 的 Run History、Comparison、Cases 与 Release Review，以及 S7 的七类资源任务，均由精确 hash 映射到持续工作面，并在目标 owner 挂载后滚动到 `72px`。相邻或前缀 hash 不获得隐式归属，没有新增路由、执行动作或伪造 owner。

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
- `S3 R1` 已完成 App owner / Panel presentation 职责拆分，Web `274/274` 测试、production build 和六组既有 consumer / designer checker 也已通过；这些结果继续作为行为回归证据，但不构成视觉验收。人工复核确认页面 `scrollWidth` 等于 viewport 并不能证明 React Flow 内部内容可见，`R1` 存在确定性裁切、节点计数与可见节点错位、边不可读、review 层级平均化和窄屏重复 Inspector，因此验收被撤回。
- `S3 R2` 已完成默认视口与层级修正：进入 Designer 时聚焦当前节点及直接邻居，显式 `Fit graph` 才展示全图；选择节点或 finding 会把目标带入可读视区。review 已收为紧凑入口，Inspector 在 `1440px` 位于右侧、`<=1380px` 下移、`<=760px` 默认折叠；窄屏不重复当前节点摘要与完整 Inspector。
- Pencil Desktop / Narrow R2 与 Decision R8 已完成全树和实际渲染复核，无裁切、越界或占位节点。Web `274/274` 测试、production build 与既有八项 workflow checker 全部通过；`workflowNodeDesigner` 为 `205.40kB < 220kB`，`index` 为 `468.40kB < 500kB`。
- 应用内浏览器严格复核覆盖 `1440x900`、`1381/1380px`、`1101/1100px`、`761/760px` 与 `390x844`：默认两节点邻域完整可读，显式 `Fit graph` 的八节点 / 七边全部完整，各宽度无横向溢出；移动端 review 位于 rail 之前。键盘删除保护、受保护节点、节点切换、finding 定位、Inspector 折叠 / 展开通过，全新页签控制台零 warning / error。
- S3 继续以 App 为唯一 application / workflow / run / draft / scenario、Saved Draft lifecycle、conflict 与 mutation owner；九组来源、十三项 contribution、revision `partial`、RAG authority `blocked`、readiness 只读且不可发布的事实不变。本轮不新增 API、schema、repository、task card、fixture 或专项 checker。
- `S4 R1` 已完成 Application Access 单一任务面：Connect API、Credentials、Validate、Verify / retire 只表达任务位置，不声称自动完成。API Integration 在 `api_key_dev_test` 未取得一次性凭据时显示资格阻塞；模型协议选择只开放目录真实声明的能力；workspace 配置与当前 context 不一致时零请求失败关闭。
- API Key 页面继续消费七项真实 scope 和既有 issue / read / revoke / rotation owner；应用切换先清空旧列表，避免旧应用 Key 暂挂在新标题下。archived application 保留脱敏记录、详情和 revoke，但 issue、rotate、integration 与 invocation 继续关闭；offline workspace summary 明确为独立只读 evidence，不冒充当前应用 Key。
- Pencil Desktop / Narrow R1 与 Decision R9 已显式保存，整棵设计树无 clipping / placeholder。Web `277/277` 与 production build 通过；应用内浏览器覆盖 `1440x900`、`1181/1180px`、`1101/1100px`、`761/760px`、`390x844`，所有宽度 `scrollWidth` 等于 viewport，任务选中、响应式顺序、一次性交接、签发 / 吊销、跨应用隔离和控制台零 warning / error 均通过。
- 浏览器连续链在本地产品档签发一枚开发测试 Key，只检查一次性 token 的结构与内存边界，不记录原文；随后以两步确认吊销该记录。URL、日志、文档、Pencil 和持久浏览器介质均未写入原始凭据。本轮没有新增 API、schema、repository、task card、fixture 或专项 checker。
- `S5 R1` 已把 Playground、Request History 与 Application Operations 编排为持续 Application Runtime Review 工作面：application / workspace / lifecycle context 始终位于三任务轨之前，只挂载 Run、Request、Evidence 中的当前 owner；深层 request / run handoff 精确切换 owner，不再由旧 S2 evidence stage 重复挂载 Operations。
- Playground 只开放选中模型实际声明的协议，并要求 active application、workspace 一致、catalog ready 与易失 credential；application 切换会中止请求并清空 catalog、credential 与结果。Request History 在 application / workspace 切换时先清空 list / detail，并用 generation guard 丢弃迟到响应。Operations 只合并当前已加载 Gateway / Workflow 窗口，以 source coverage 明示缺失来源，不推测跨来源关联。
- Pencil Desktop / Narrow R1 与 Decision R10 已显式保存，整棵设计树无 clipping / placeholder。Web `278/278` 与 production build 通过；应用内浏览器覆盖 `1440x900`、`1281/1280px`、`1101/1100px`、`761/760px`、`390x844`，各宽度无横向溢出，任务选中、单 owner、筛选展开、应用切换清空与控制台零 warning / error 均通过。验收数据没有 archived application 代表记录，因此未通过修改真实数据制造该状态；代码和 context 投影继续保证 archived Run disabled，history / evidence 不被伪写为空。
- S5 继续把调用结果限定在当前组件内存，把 Request History 详情限定为脱敏 envelope，把 Application Operations 限定为当前窗口 evidence；不新增 retry / fallback、replay、quota / billing、生产认证、API、schema、repository、task card、fixture 或专项 checker。
- `S6 R1` 已把 Run History、Run Comparison、Evaluation Case / Suite 与 Human Release Review 编排为四任务单 owner 工作面。Run 列表仍是严格 cursor 的当前窗口，详情仍为 metadata-only；Comparison 只消费精确且兼容的 run pair，并保持即时派生、不持久化；Case / Suite 继续消费 exact run / case version 引用；人工 decision 必须绑定当前 review digest，`approved` 只保存 append-only evidence。
- application / workspace 切换会先清空 run、detail、comparison 与 handoff 状态，并用 generation guard 拒绝迟到响应；workspace 配置不一致时所有 live owner 零请求失败关闭。archived application 保留 Run、Case、Suite 和 decision history 只读可达，诊断、创建、修订与 decision 写入关闭；offline evidence 明确独立，不冒充当前 application。
- Pencil Desktop / Narrow R1 与 Decision R11 已显式保存，整棵设计树无 clipping / placeholder。Web `281/281` 与 production build 通过；S6 chunks 分别为 `workflowReviewWorkspace 4.82 KiB`、`workflowReviewOwner 23.00 KiB`、`workflowEvaluationPanel 8.57 KiB`、`workflowEvaluationSuitePanel 15.47 KiB`、`workflowRunComparisonPanel 27.70 KiB`，主入口 `469.67 KiB` 仍低于 `500 KiB` 预算。
- 应用内浏览器覆盖 `1440x900`、`1281/1280px`、`1101/1100px`、`761/760px` 与 `390x844`：精确任务切换、单 owner、Run history 一级导航归属、窄屏 context → task → owner 顺序和深链定位正确，各宽度无横向溢出，控制台零 warning / error。默认 runtime 为 offline，因此浏览器验收如实覆盖失败关闭资格面；Pencil 的 live dev / test 数据只作为已批准代表态，不冒充本机 live owner 结果。
- S6 没有新增自动评测执行、replay / resume、自动 candidate / assignment / release、全历史聚合、业务写回、API、schema、repository、task card、fixture 或专项 checker。
- `S7 R1` 已把 Tenant、User、Role、Audit、Provider、Profile 与 Route 编排为“管理上下文 → 七资源定位 → 单一当前 owner”。Tenant / Audit 继续使用既有 `tenant:read` / `audit:read` consumer；Audit Web 补齐现有契约允许的 cursor next / previous current-window 导航和行驱动 metadata-only detail，offline next 保持禁用。User / Role 只呈现 Radish-owned 真相边界和 blocked action；Provider / Profile / Route 继续复用四项独立权限、draft CAS、不可变 candidate、独立 review 与显式 generation activation。
- Pencil Desktop / Narrow R1 与 Decision R12 已完成全树检查，结果为零裁切、零越界、零占位。Web `287/287` 与 production build 通过；`adminControlPlaneWorkspace 34.33 KiB`、`adminProviderRouteWorkspacePanel 22.47 KiB`，主入口降至 `458.43 KiB < 500 KiB`。
- 应用内浏览器覆盖 `1440x900`、`1281/1280px`、`1101/1100px`、`901/900px`、`761/760px` 与 `390x844`：七任务精确 hash、单 owner、Audit 行选中 / 只读详情、Supporting evidence 折叠、窄屏 context → task → owner 顺序和关键布局切换正确；所有宽度无横向溢出，控制台零 warning / error。默认 offline fixture 没有被改写成 live 管理事实，也没有产生 Provider / Route 管理请求。
- S7 没有新增 API、schema、repository、permission、task card、fixture、专项 checker、生产 membership、正式 OIDC、secret material、provider onboarding、自动路由或生产启用。
- `S8 R1` 已把 Prompt 的 Template → Configuration → Candidate → Assignment → Access → Invocation → Session → Evaluation，以及 Agent 的 Profile → Configuration → Candidate → Assignment → Access → Suggestion → Evaluation 编排为持续类型工作区；任一时刻只挂载一个既有 owner，Evaluation 只交接 S6，不复制证据面。
- 五维评分为 `1 / 2 / 2 / 2 / 2`，采用 `A / 完整 Pencil`。Desktop / Narrow R1 与 Decision R13 已完成全树检查，结果为零裁切、零越界、零占位；Prompt 差异进入共享决策记录，不复制第二套完整画板。
- 2026-08-09 跟进后 Web `295/295` 与 production build 通过；`promptAgentTypeWorkspace 9.38 KiB`，主入口 `462.50 KiB < 500 KiB`。应用内浏览器复验 Agent Invocation → Prompt Invocation、Prompt Session → Agent Suggestion、Profile → Template、Assignment 与 Access 双向切换，以及刷新后的等价任务身份；并覆盖 `1440×900`、`1100×900`、`900×900`、`760×844` 与 `390×844`。各宽度无横向溢出，单任务 / 单 owner、响应式顺序与控制台零 warning / error 均通过。
- 当前真实产品数据没有 archived application，浏览器没有改写数据制造该状态；纯模型测试继续保证 archived source / governance 只读和 controlled use 阻塞。S8 没有新增 API、schema、repository、permission、task card、fixture、专项 checker、自动 assignment、自动 release、agent loop、production membership、正式 OIDC、生产 secret、provider 自动接入、自动路由或生产启用。
- S8 受控使用失败交接跟进采用共享允许列表 view model：Prompt Invocation、Prompt Session 与 Agent Session 只对真实 assignment / authority 失败显示原因、零 provider 副作用和当前类型 Assignment 入口，其它失败留在原 owner。Web `298/298`、production build 和 `1440×900`、`1100×900`、`760×844`、`390×844` 浏览器复验通过，单任务 / 单 owner、窄屏顺序、零横向溢出与控制台零 warning / error 均保持成立；既有 S8 Pencil 基准面不变。
- `S9` 已把 Quota 作为 S7 Admin Control Plane 的第八项精确 owner：application 列表只让当前详情 owner 获得墨蓝选中轨，policy / exceeded / blocked 状态使用独立文字和语义色；详情只消费 quota owner 的 UTC policy / usage，不从 Request History 或旧 `QuotaSummary` 推导。
- 五维评分为 `1 / 2 / 2 / 2 / 1`，采用 `A / 完整 Pencil`。原 Desktop / Narrow R1 与 Decision R14 因页面骨架偏离被人工退回；Visual R2 虽冻结 policy / usage 层级、expected-version 确认、稳定 blocked reason 和窄屏 context → path → owner → boundary 顺序，仍因硬方形表面偏离 S1–S8 被再次退回。Visual R3 已显式保存，通过零折叠、零硬编码色、零占位检查，并于 2026-08-12 完成人工视觉复核。
- React 严格校验 exact envelope、作用域、UTC usage 算术与敏感字段；正整数创建 / 更新先 review 再 confirm，CAS 冲突保留旧 owner、禁用编辑并只提供 reload。permission、environment、missing policy、version conflict、store failure 与 disabled gate 均失败关闭，不回退旧 quota evidence。
- Web `304/304` 与 production build 通过；应用内浏览器覆盖 `1440×900`、`900×900`、`390×844`，各宽度 `scrollWidth` 等于 viewport，应用切换始终只有一个选中轨，missing → create → ready 与并发 version conflict → reload 连续链通过，控制台零 warning / error。自启动服务已关闭，`7000` / `4100` 端口已释放。
- 2026-08-13 的 Visual R3 迁移没有改 quota consumer 或更新语义：详情新增 selected application context，policy / usage / editor 收进单一主 owner，remaining 与 CAS 成为辅助 rail，scope / policy 保持连续事实行，boundary 继续独立。SQLite 本地产品再次完成 missing → create v1 → 双标签 stale CAS → reload v2；`1440×900`、`720×900`、`390×844` 的 document、workspace、detail 与 owner 均无横向溢出，控制台无 warning / error。
- S9 Web 批次复用既有后端契约、权限和聚合验证，没有新增 API、schema、migration、repository、permission、task card、fixture 或专项 checker；production quota、formal membership / OIDC、token / cost、billing、删除 / 禁用、自动提额和自动路由继续关闭。
- S9 完成后的真实 API Key 路径复核在 S5 Playground 发现 quota admission 失败恢复信息断点。五维评分为 `0 / 0 / 0 / 1 / 1 = 2`，采用 `C / 直接实现`：复用既有失败引导卡，只对允许列表内的 `quota_admission` failure 说明 UTC 日预算、零 provider 调用并打开当前 application 的 Admin Quota owner；Web `305/305`、production build 与 `1440×900`、`900×900`、`390×844` 浏览器复验通过，各宽度保持 context → task → owner 顺序、零横向溢出和控制台零 warning / error。既有 S5 / S9 Pencil 基准面继续有效，本批未操作被其它项目占用的 Pencil。
- Workflow RAG Promotion → Configuration Draft 跟进把静态 hash 收紧为 S2 workspace 单引用 handoff：配置 owner 只按精确 `candidateId` 重读并选择当前 `approved + eligible` binding，不回退、不自动恢复、不 attach。Web `308/308`、production build 与四个视口通过；来源草案 `v1` 对当前 `v2` 的恢复以既有稳定 failure 失败关闭。既有 S2 / RAG owner 设计基准继续有效，本批未操作 Pencil。
- `S10` 的功能切片已经完成；原 Desktop / Narrow R1 与 Decision R15 因 dashboard 式进度卡、稀疏工作面和另一套视觉骨架被人工退回。Visual R2 在相同根节点恢复 selected campaign 单一 owner、连续 evidence rows 和单一 handoff rail 后，仍因硬方形业务表面被再次退回；Visual R3 已按职责圆角修订，通过零折叠、零硬编码色、零占位检查，并于 2026-08-12 完成人工视觉复核。2026-08-13 React 已按该基线完成迁移：selected campaign context 驱动唯一主 owner，item 使用连续 evidence rows，Handoff 只占单一辅助 rail；真实 campaign 切换与 exact Handoff 通过，`1440×900`、`720×900`、`390×844` 无横向溢出，控制台无 warning / error。Web `346/346` 与 production build 通过，既有 memory / SQLite Plan → 两次 Campaign → Pair → exact Case / Suite Handoff 证据继续成立。

## 停止线

- 基础批次不做 `styles.css` 全量换色、组件重写或页面重新布局。
- 不因 family-ui 已提供暗色 token 就声明暗色主题切换可用。
- 不由 UI 产品化专题继续新增 API、schema、migration、repository、生产认证、production quota / billing 或执行能力；开发测试态 quota 只消费其独立功能专题的既有 owner。
- 不把开发测试态 owner、离线 evidence、首分页窗口或人工审查画成生产就绪、全量统计、自动执行或业务写回。
- 不要求每个路由、组件、功能状态或相似页面都有独立 Pencil 画板，也不让 Pencil 取代功能文档与当前代码。
- 不为普通 UI 迁移新增 task card、fixture 或专项 checker；优先复用 Web 测试、build、consumer smoke 和仓库门禁。
