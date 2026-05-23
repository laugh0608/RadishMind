# RadishMind UI 设计参考

更新时间：2026-05-23

## 文档目的

本文档为未来 `UI Design Topic / Pencil Draft` 提供视觉参考素材和设计语言约束。这里收录的截图只作为内部灵感参考，用来学习优秀产品在布局、信息密度、颜色、层级、圆角、留白和状态表达上的处理方式。

这些参考不代表 RadishMind 要复制任何产品的界面、品牌、图标、配色或交互细节。未来 UI 必须先结合 RadishMind 的平台定位、只读/可执行边界、session/tooling metadata、readiness、错误诊断和确认流要求重新设计，再用 `pencil` 绘制 `.pen` 设计稿，定稿后才进入 React 实现。

## 参考素材清单

素材统一保存在 `docs/assets/ui-design-reference/`。文件名按来源、画面主题和原始时间戳重命名，便于后续追溯。

| 来源 | 文件 | 参考重点 |
| --- | --- | --- |
| AFFINE | [affine-docs-list-calendar-20260522210036.png](assets/ui-design-reference/affine-docs-list-calendar-20260522210036.png) | 三栏布局、文档列表密度、右侧辅助面板、温和浅色背景 |
| AFFINE | [affine-settings-modal-20260522210047.png](assets/ui-design-reference/affine-settings-modal-20260522210047.png) | 大尺寸设置弹窗、遮罩层、分组表单、圆润控件 |
| AFFINE | [affine-document-properties-20260522210158.png](assets/ui-design-reference/affine-document-properties-20260522210158.png) | 文档正文与属性栏并存、标题尺度、标签与属性行 |
| 1Panel | [onepanel-analytics-dashboard-20260522213828.png](assets/ui-design-reference/onepanel-analytics-dashboard-20260522213828.png) | 运维 dashboard 信息分区、指标卡与趋势图组合 |
| Cloudflare | [cloudflare-access-control-list-20260522214117.png](assets/ui-design-reference/cloudflare-access-control-list-20260522214117.png) | 安全/访问控制列表、表格和侧栏状态组织 |
| GitHub | [github-repository-file-list-20260522214140.png](assets/ui-design-reference/github-repository-file-list-20260522214140.png) | 仓库文件列表、右侧项目元信息、工具栏密度 |
| GitHub | [github-profile-repositories-20260522214151.png](assets/ui-design-reference/github-profile-repositories-20260522214151.png) | 个人资料与列表组合、卡片轻量边界、活动信息 |
| GitHub | [github-account-settings-20260522214216.png](assets/ui-design-reference/github-account-settings-20260522214216.png) | 设置页侧栏、表单区块、保守而清晰的分隔线 |
| Discourse | [discourse-topic-list-20260522214335.png](assets/ui-design-reference/discourse-topic-list-20260522214335.png) | 论坛主题列表、状态计数、紧凑信息表格 |
| Discourse | [discourse-chat-thread-20260522214411.png](assets/ui-design-reference/discourse-chat-thread-20260522214411.png) | 聊天/消息流、左侧频道导航、输入区位置 |
| 1Panel | [onepanel-file-manager-20260522215732.png](assets/ui-design-reference/onepanel-file-manager-20260522215732.png) | 文件管理器、密集表格、操作列和浅蓝强调色 |
| 1Panel | [onepanel-app-store-20260522215756.png](assets/ui-design-reference/onepanel-app-store-20260522215756.png) | 应用卡片网格、分类侧栏、操作按钮位置 |
| Discourse | [discourse-admin-dashboard-20260522215935.png](assets/ui-design-reference/discourse-admin-dashboard-20260522215935.png) | 管理后台概览、统计摘要、后台信息密度 |
| Discourse | [discourse-site-analytics-20260522215943.png](assets/ui-design-reference/discourse-site-analytics-20260522215943.png) | 多图表分析页、纵向信息节奏、指标切片 |
| 1Panel | [onepanel-metrics-monitoring-20260522220028.png](assets/ui-design-reference/onepanel-metrics-monitoring-20260522220028.png) | 监控图表矩阵、时间范围筛选、资源指标可读性 |
| CodexApp | [codexapp-settings-general-20260522220204.png](assets/ui-design-reference/codexapp-settings-general-20260522220204.png) | 设置中心、柔和侧栏、宽内容区、权限说明层级 |
| CodexApp | [codexapp-browser-permissions-20260522220246.png](assets/ui-design-reference/codexapp-browser-permissions-20260522220246.png) | 权限设置详情页、状态徽标、危险操作分层 |

## 重点参考方向

AFFINE 与 CodexApp 的截图是当前最重要的风格参考。它们共同的优点不是某个单独组件，而是整体节奏：浅色、圆润、克制、信息密度适中，左侧导航清楚但不抢主内容，正文区域留白充分，设置和属性内容用分组、细分隔线、弱色说明文字和少量强调色建立层级。

1Panel、GitHub、Cloudflare 和 Discourse 更适合作为信息组织参考。它们展示了后台产品常见的高密度列表、指标卡、表格、监控图表、仓库文件列表、权限/访问控制和社区消息流如何保持可扫描性。RadishMind 未来不应做营销式首页，而应更接近安静、可重复使用、面向开发者和维护者的产品工作台。

## 可吸收的设计语言

- 布局：优先采用稳定的左侧导航 + 主内容 + 可选右侧详情/诊断栏结构。右侧栏只承载上下文属性、readiness、audit note 或诊断摘要，不抢主流程。
- 信息密度：列表、表格和状态面板保持紧凑，但每个区块要有明确标题、分隔和视觉呼吸。关键状态用短标签、徽标和数值摘要表达，不用长段说明占据核心视图。
- 色彩：以温和中性色为主，保留少量蓝色作为可交互强调，绿色用于 ready / connected，红色用于危险或失败，黄色/橙色用于 warning。不要做单一蓝紫渐变或大面积装饰背景。
- 圆角和边界：整体可以偏圆润，但运维类面板仍要克制。卡片、弹窗、输入框和选择器使用轻边框、浅背景和 6-8px 左右圆角；避免过度胶囊化。
- 层级：标题、分组标题、正文、辅助说明和元信息必须有清晰字号/字重/颜色层级。主标题只用于页面或大段内容，不在紧凑面板里使用过大的展示字。
- 交互状态：ready、loading、stale、failed、blocked、requires confirmation、read-only 都要有独立视觉状态。失败态要保留上一份可读数据时，明确标记 stale，不把本地诊断误写成 production incident。
- 设置与权限：参考 CodexApp 和 AFFINE 的设置页结构，优先用分组列表、开关、下拉和状态徽标表达配置，不把复杂解释塞进按钮或卡片标题。
- 监控与诊断：参考 1Panel 和 Discourse 的图表/指标布局，先展示最关键的健康度、端口、route、provider/profile 和 stop-line，再提供细节展开。

## RadishMind 适配原则

RadishMind 的 UI 不是通用文档工具、服务器面板、社区论坛或代码托管平台。它的核心界面应围绕平台运行、模型/profile inventory、session/tooling metadata、blocked action、local readiness、错误诊断和未来确认动作组织。

未来 Pencil 设计稿至少应覆盖以下页面或状态：

1. `Local Overview`：服务状态、当前 provider/profile、可用 route、停止线摘要。
2. `Local Readiness`：healthz、overview、local-smoke、CORS、端口、no-side-effects 和诊断提示。
3. `Session & Tooling Surface`：session metadata、tool registry、blocked action、audit notes。
4. `Failure / Stale View`：overview 失败、local-smoke 失败、contract mismatch、CORS / unsafe port、保留上一份只读视图。
5. `Settings / Permissions`：未来 provider 配置、权限、确认策略和本地环境说明。
6. `Narrow Layout`：窄屏下保持导航、状态摘要和诊断信息可读，不让文本溢出或互相遮挡。

## Pencil 专题要求

- `pencil` 设计稿应先定信息架构和关键状态，不先追求装饰细节。
- 每个主要页面要同时画 ready、loading、failed / stale 和 blocked 状态。
- 设计稿中必须标明哪些控件只是只读展示，哪些未来可能进入 confirmation flow。
- 不在设计稿定稿前大面积重构 `apps/radishmind-console/` 的视觉系统。
- 定稿后再把设计拆成小的 React + Vite + TypeScript 实现任务，并补对应 visual smoke 或行为门禁。

## 不做什么

- 不直接照抄 AFFINE、CodexApp、1Panel、Discourse、GitHub 或 Cloudflare 的品牌、图标、页面结构和配色。
- 不把参考截图中的营销感、社交产品语境或通用后台能力直接搬进 RadishMind。
- 不为了“好看”牺牲平台边界：executor、durable store、confirmation、业务写回和 replay 未 ready 时，UI 必须继续明确 blocked / read-only。
- 不在没有 Pencil 定稿前，把当前本地 console 当作最终产品 UI。
