# RadishMind Family UI 参考图产品面映射 v1

更新时间：2026-08-08

状态：`radishmind_family_ui_reference_mapping_v1_s1_r8_s2_r6_s3_r2_implemented`

## 文档职责

本文件把 RadishX `docs/design/family-ui/references.md` 登记的 27 张外部 UI 参考图，逐项映射到 RadishMind 首批四个产品化设计面。它只承载“吸收什么、放到哪里、明确不学什么”的设计决策，不复制图片，不替代 family-ui 通用规范，也不改变 RadishMind 功能边界。

四个目标产品面：

| 编号 | 产品面 | 首批真实任务 |
| --- | --- | --- |
| `S1` | 产品壳与工作区导航 | workspace / application context、一级任务导航、环境、状态总览与 Operations Inbox |
| `S2` | Application Development Workspace | 五阶段路径、readiness、owner evidence、lifecycle、CAS / drift 与 handoff |
| `S3` | Saved Draft Library / Workflow Designer | 活动 / 归档草案、列表、精确打开、画布、属性、校验与审查交接 |
| `S4` | Application API Integration / API Key | 模型资格、接入向导、示例、scope、一次性凭据、验证与退役 |

## 使用与版权边界

- 外部截图只用于内部风格学习，不进入 `public/`、产品页面、设计交付包或对外材料。
- 只吸收布局、密度、组件形态、状态表达和页面节奏；不复制页面、图标、配色、品牌元素或文案。
- 进入对应设计基准面前必须实际查看相关参考图，记录“吸收什么、排除什么、如何转译”；只读取索引、标题或 token 不能替代视觉参考。
- Pencil 画板只标注 `ref-XX` 和采用原则，不嵌入外部截图。
- 全部参考图都是桌面稿；`390x844` 窄屏设计只依据 family-ui `07-layout-platforms.md` 和 RadishMind 真实任务顺序，不从截图缩放推导。
- 暗色参考只形成未来主题证据，不使当前暗色开关获得实现准入。

## S1：产品壳与工作区导航

| 参考 | 主要吸收 | RadishMind 设计落点 | 明确不采用 |
| --- | --- | --- | --- |
| `ref-03` HR 仪表盘侧栏 | 图标、文字、计数徽标的导航层级；克制图表用色 | 一级产品区、任务组和关注数；当前入口使用柔底选中态 | 高饱和蓝色整块 active 背景 |
| `ref-07` 自动化仪表盘 | KPI、主表格、模板卡的三段式节奏 | 工作区概览先给状态摘要，再给关注队列和继续任务 | 把模板推荐做成营销卡 |
| `ref-08` 通知 severity 面板 | severity tab 计数；图标、标题、摘要、时间、级别 chip | Operations Inbox 的优先级筛选、关注项结构和 partial coverage | 荧光严重度、仅靠颜色区分 |
| `ref-09` 自动化仪表盘全景 | 侧栏、主区、卡片间距比例 | `1440x900` 产品壳整体比例与首屏节奏 | 逐像素复制原页面结构 |
| `ref-15` SEO 仪表盘 | 等宽元信息、数值密度、条状分布 | route / profile / usage metadata 和紧凑指标标签 | 蓝橙图表原配色、无 owner 的趋势推测 |
| `ref-17` HR 出勤仪表盘 | 非对称数据重心、清晰大数值、分布与明细表组合 | `S1 R8` workspace pulse 与 Source evidence 的状态分布、四来源矩阵和关注明细分层，以及 `S2 R6` readiness 的 `5 / 9` 九格矩阵与次级风险 | 青绿原配色、HR 语义、把首分页冒充全量 |
| `ref-18` 律所日历工作台 | 侧栏功能分组、对象分组、标题与计数徽章 | workspace / application 分组与当前对象计数 | 日历业务语境、紫色主强调 |
| `ref-19` 文档应用 space 切换 | 账户、状态、space 子菜单和设置分组 | workspace / application context switcher | 彩虹头像、把切换状态写入持久浏览器介质 |
| `ref-20` 深色分析仪表盘 | 暗色下卡片、边框和前景层次 | 只作为未来 Workbench 暗色审查证据 | 在本批提供暗色开关或点阵地图装饰 |
| `ref-24` 告警运营台 | Critical / Active / New / Unack 快速统计、告警表与刷新提示 | Operations Inbox 严重度摘要、来源状态和显式刷新时间 | 紫底展示背景、自动刷新冒充实时 incident |

S1 组合原则：产品壳首屏使用“当前上下文 → 少量可信状态摘要 → 关注队列 → 继续任务”，不再把所有 read-shell 状态卡平铺成同等权重。

## S2：Application Development Workspace

| 参考 | 主要吸收 | RadishMind 设计落点 | 明确不采用 |
| --- | --- | --- | --- |
| `ref-05` AFFiNE 设置 | 左导航与右侧表单分区、中文排版密度、主题三态结构 | 五阶段内部的分区导航、配置表单与说明层级 | 本批实现主题切换 |
| `ref-11` 事故详情侧栏 | label 左、值右的上下文详情行 | 当前 application、revision、owner ref、readiness evidence 详情栏 | 外部集成入口直接执行动作 |
| `ref-12` 事故时间线 | 旧态到新态的迁移、时间线、状态与级别双通道 | lifecycle、candidate review、assignment 与 CAS 事件记录 | 事故语境、橙色品牌强调 |
| `ref-13` 履约进度 | 多段进度、当前步骤、引用 chip 和右对齐摘要 | Configure → Promotion → Test → Review → Readiness 五阶段路径 | ETA 承诺、订单与金额语义 |
| `ref-14` ATS 阶段管理 | 阶段 KPI 分组、排序指示、来源图标与文字 | owner contribution、阶段完成度和来源覆盖 | 紫色主按钮、把阶段数值当自动发布资格 |
| `ref-27` 客服协作 Inbox | 连续导航 / 队列 / 详情窗格、轻量选中轨和明确上下文 | `S2 R6` review path / contribution / readiness 三窗格与当前阶段导航焦点，并共享给 `S1 R8` 的 Operations Inbox 四项队列、列表 / 详情双窗格和 evidence path | 客服聊天产品照搬、纸纹背景、AI Agent 品牌和语音状态 |

S2 组合原则：五阶段是任务路径，不是装饰性 stepper。每一阶段必须显示当前 owner、已满足证据、阻塞原因和下一次人工动作；readiness 仍是不可持久化、不可发布的只读投影。

## S3：Saved Draft Library / Workflow Designer

| 参考 | 主要吸收 | RadishMind 设计落点 | 明确不采用 |
| --- | --- | --- | --- |
| `ref-04` 看板与详情浮层 | 卡片信息层级、详情字段、子任务嵌套 | 草案摘要、节点属性与校验详情的渐进展开 | 紫色 accent、把草案生命周期做成通用看板拖拽 |
| `ref-06` CRM 工作流列表 | 筛选、视图切换、搜索、行内状态 | 活动 / 归档草案库筛选、列表工具条与 lifecycle chip | 导入能力、无 owner 的 Active toggle |
| `ref-21` 邮件三栏阅读态 | 文件夹、列表、详情三栏与附件卡层次 | 草案分类、草案列表、Designer / Review 详情的桌面模式 | 邮件业务语义、把原始内容持久化为附件 |
| `ref-22` 社区发帖器 | 类型分段、标题、正文、标签与右侧线索栏 | 节点类型、属性编辑、输入字段和 validation clues | 社区帖子类型、发布动作 |
| `ref-23` 社区内容流与 toast | 卡片摘要、标签、反馈入口、成功 toast | 草案保存结果、状态标签和易失成功反馈 | 投票、反应和 Active Now 社交元素 |
| `ref-26` 邮件 Inbox tabs | Inbox / Sent / Drafts / Archive tab、勾选与附件卡 | 活动 / 归档草案分段、明确 selected row 与详情交接 | 批量生命周期、附件预览和 Spam 语义 |

S3 组合原则：完整 Saved Draft Library 保持为唯一列表 owner，Designer 只保留当前 / 活动草案引用和精确打开交接。`S3 R2` 桌面默认聚焦当前节点及直接邻居，让节点、端口和连线保持可读；全图查看由显式 `Fit graph` 触发，不用缩小到不可读来冒充四节点均可见。review 只保留紧凑摘要和展开入口；Inspector 在 `1440px` 固定于右侧，`<=1380px` 下移，`<=760px` 默认折叠。窄屏按“上下文 → 动作 → 当前节点 / 直接邻居画布 → 紧凑审查”重排，不重复同一节点摘要与完整 Inspector。强选中只归属于当前导航或正在驱动 Inspector 的对象；readiness、lifecycle 和 finding 的状态色不承担选中语义。

## S4：Application API Integration / API Key

| 参考 | 主要吸收 | RadishMind 设计落点 | 明确不采用 |
| --- | --- | --- | --- |
| `ref-01` 电商列表与 KPI | KPI 卡、软色状态 chip、pill 筛选、列表行结构 | active / expiring / revoked Key 摘要和当前应用模型资格 | 纯灰白底、趋势推测和电商指标 |
| `ref-02` 订单行内操作 | 行内操作组、状态与分页器形态 | Key detail、验证、退役和严格 cursor 列表操作 | 价格 chip、无确认的批量操作 |
| `ref-10` 可用性行式编辑器 | toggle、多行字段、行内增删复制、发丝分隔 | scope、expiry、环境变量与接入参数的紧凑编辑 | 日程语义、复制真实 secret |
| `ref-16` 数据源连接向导 | 四步 stepper、选择卡、双列表单、成功状态条 | 选择 application → 获取 Key → 选择协议 / 模型 → 验证调用 | 自动连接外部数据源、把成功测试写成生产 readiness |
| `ref-25` 分享与邀请设置 | 白色抬升表面、细边界、宽松控件和限制提示分区 | `S4` scope / 一次性交接 / 验证门槛，并作为 `S1 R8` / `S2 R6` 的共享表面质感 | 密码重置、上传、Dark Mode 和真实 token 回显 |

S4 组合原则：接入向导只编排既有 owner。原始 API Key 只在签发成功响应中出现一次；刷新、离开、application 切换和服务重启均不可恢复。

## 跨产品面设计约束

1. 统一使用纸色底、暖白表面、发丝边框和极柔阴影；身份使用灰玉、操作使用墨蓝，不把参考图的冷灰、紫蓝或高饱和品牌色带入 RadishMind。
2. 状态 chip 使用低饱和浅底与深字，并同时提供文字、图标或结构第二通道。
3. 单一主交互色使用墨蓝；玉色用于已满足和成功，紫色只做小面积类型区分，胭脂只表达需要注意且不等同危险的项目语义。
4. 桌面多栏用于降低上下文切换，不用于同时展示所有功能；不属于当前任务的 owner surface 不挂载。
5. 图表、KPI 和计数只显示已有可信数据，不估算缺失 token、cost、quota、全历史 usage 或生产健康度。
6. 外部参考只提供原则，不覆盖功能专题中的权限、CAS、生命周期、敏感信息和生产停止线。

## Pencil 使用方式

首批 Pencil 画板按以下顺序消费本映射：

1. `S1` 产品壳：`1440x900` ready / partial，`390x844` 单列导航与关注队列。
2. `S2` Application Workspace：五阶段、blocked readiness、CAS / drift confirmation。
3. `S3` Saved Draft / Designer：活动与归档库、精确打开、画布 / inspector、failed / stale。
4. `S4` API Integration / Key：接入向导、一次性交接、验证后退役和资格失败。

每张画板的说明区至少记录：

- 使用的 `ref-XX`；
- 吸收的布局或状态原则；
- 明确未采用的品牌、业务或交互元素；
- 对应 RadishMind owner、状态和停止线；
- 桌面与窄屏的结构差异。

## 验收结论

- `ref-01` 至 `ref-27` 每项恰好有一个首要产品面。
- 四个产品面都具有可追溯参考依据和明确不采用项。
- 没有复制、提交或重新托管外部参考图。
- 移动端、暗色和未实现能力没有从参考图获得隐式准入。
- 2026-08-03 已实际重新查看 `ref-05`、`ref-11`、`ref-12`、`ref-13`、`ref-14`、`ref-18` 与 `ref-27`；原有产品面映射仍成立，但 `ref-18` 的大尺度品牌 / 导航 / 搜索、抬起选中实体与宽松主对象，以及 `ref-13` 的图标任务身份、阶段 rail 与单一宽工作面，应作为 `S1` / `S2` 共享视觉语法，而不只是局部结构参考。
- 2026-08-05 联合人工评审判定 `S1 R5` / `S2 R3` 仍受米色雾感、圆角小卡和平均化信息块影响，现代产品感不足。`S1 R6` / `S2 R4` 据此重点重读 `ref-27`、`ref-25` 与 `ref-17`，分别吸收连续窗格、白色抬升表面与非对称数据重心，同时保留五阶段、九组来源、十三项 contribution、owner evidence、readiness 和 blocked / partial 停止线；该轮随后进入第二次人工复评，期间未进入 React。
- 同日第六轮人工复评确认，`R6` / `R4` 的参考方向已经成立，但 Operations Inbox、Source evidence 和 Application Workspace 仍过于规整，桌面外圈宽留白与大圆角窗口缺少产品职责。`S1 R7` / `S2 R5` 因此把 `ref-27` 的选中轨和队列节奏、`ref-17` 的主次数据关系进一步落实到真实工作面，并取消两个桌面根画板的外圈容器；该轮随后进入第七轮复评。
- 同日第七轮人工复评继续指出，`R7` / `R5` 虽已去除无职责外框，但局部仍像规整状态文字面板。`S1 R8` / `S2 R6` 因此进一步吸收 `ref-27` 的紧凑焦点队列与详情上下文，以及 `ref-17` 的矩阵化数据节奏：Operations Inbox 增加四项当前窗口和 evidence path，Source evidence 改为状态分布与四来源矩阵，Application Workspace 增加当前阶段选中轨、十三段 contribution window、九格 readiness 和 authorization path。后续局部复核又把 Inbox 选中行降为中性柔底与描边图标，并取消中央缺失 contribution 的伪选中态；选中只归属于详情 owner 或当前导航，状态仍由文字、图标和标签表达。功能真相、窄屏顺序与 React 停止线未改变。
- 2026-08-06 联合人工复评通过 `S1 R8` / `S2 R6`。参考映射、克制家族色、全视口桌面结构、窄屏渐进顺序和状态停止线均冻结为当前实现输入；后续普通实现偏差直接在 React 与浏览器复核中收口，只有结构性设计决策变化才重新回写 Pencil。
- 2026-08-08 重新审阅 `ref-04`、`ref-06`、`ref-21`、`ref-22`、`ref-23` 与 `ref-26` 的既定吸收边界，并完成 `S3 R1` 行为实现；但后续人工评审确认，Pen 全树检查与 body 无横向溢出没有覆盖 React Flow 内部确定性裁切。`R1` 的节点计数与实际可见节点错位、连线不可读、底部三块 review 等权铺陈，以及窄屏节点裁切与 Inspector 重复层级均不符合参考映射，因此视觉验收撤回。
- `S3 R2` 已按当前运行时事实修正上述结构，不改 owner 或能力边界：默认聚焦当前节点与直接邻居、提供显式 `Fit graph`、收紧 review 密度，并按 `1440px` 右侧 / `<=1380px` 下移 / `<=760px` 折叠安排 Inspector。Pencil Desktop / Narrow R2 与 Decision R8 实际渲染无裁切；桌面、临界断点和 `390x844` 的真实浏览器严格复验确认默认两节点邻域和全图八节点 / 七边完整可读、无横向溢出、移动端 review 位于 rail 之前。归档草案仍只读，九组来源、十三项 contribution、revision `partial`、RAG authority `blocked`、readiness 不可发布保持不变；不新增 API、schema、task card 或专项 checker。下一步进入 `S4 Application API Integration / API Key` 功能事实复核、`A` 级设计与 React 纵向切片。
