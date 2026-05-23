# RadishMind UI 设计规范

更新时间：2026-05-23

## 文档目的

本文档是 RadishMind 前端 UI 的正式设计规范源，用于约束 `apps/radishmind-console/`、Pencil 设计稿和后续 React 实现任务。

[UI 设计参考](radishmind-ui-design-reference.md) 只提供灵感素材和外部产品观察；本文件定义 RadishMind 自己的视觉目标、语义 token、组件规则、状态表达、窄屏策略和设计稿治理要求。

当前规范优先服务 `UI Design Topic / Pencil Draft` 与 `P3 Local Product Shell / Ops Surface`。它不声明正式 production console 已完成，也不把 executor、durable store、confirmation、业务写回或 replay 画成当前能力。

## 设计定位

RadishMind UI 是面向开发者、维护者和本地部署者的 AI runtime ops surface。它的核心任务不是展示品牌气氛，而是让平台状态、模型/provider/profile、session/tooling metadata、readiness、错误诊断和停止线一眼可读。

风格关键词：

- 克制
- 清晰
- 可审计
- 工程化
- 本地可诊断
- 边界明确

避免方向：

- 营销式首页、夸张 hero、装饰性大图
- 单一蓝紫渐变、深色赛博风或高饱和仪表盘
- 把未实现能力画成可点击执行路径
- 用长说明文案替代清晰的状态层级
- 为了好看弱化 `read-only`、`blocked`、`stale`、`requires_confirmation` 等边界

## 设计原则

1. 状态优先：每个页面先回答“当前可读吗、可信到什么程度、哪些能力不可用”。
2. 证据可见：关键状态要有来源或证据摘要，例如 route、provider/profile、last refresh、stop-line reason。
3. 建议不执行：模型建议、工具动作和上层写回默认只是候选，执行类控件必须等确认流与门禁完成后再设计。
4. 失败不升级：本地 overview、local-smoke 或 CORS 失败时，只展示本地诊断和 stale snapshot，不把它误写成 production incident。
5. 密度有序：ops surface 可以高信息密度，但主区、辅助栏和诊断区必须有清楚层级。
6. 先 token 后颜色：页面只能消费语义 token；设计稿可标注建议色值，但实现时必须收敛到 token。

## 布局规范

桌面基准尺寸为 `16:9 / 1920x1080`。

桌面默认结构：

- 左侧导航：产品身份、主页面入口、当前只读边界提示。
- 主内容区：页面标题、refresh / retry、全局状态条、关键指标、主要详情面板。
- 右侧辅助栏：local readiness、stop-line、audit note、failure hint、stale rule。

窄屏基准尺寸为 `9:16 / 1080x1920`。

窄屏默认结构：

- 单主列，不压缩桌面三栏结构。
- 顺序为：标题与状态、关键指标、readiness、session/tooling、stop-line、failure/stale rule。
- 导航收敛为顶部或底部入口，不保留桌面侧栏。
- 文案必须按 `1080` 宽度预留英文长度冗余，不为中文短文案卡死宽度。

## 语义 Token

首版 UI token 至少覆盖以下角色：

| Token | 用途 | 当前建议 |
| --- | --- | --- |
| `--rm-bg-app` | 应用背景 | 浅冷灰或近白 |
| `--rm-bg-surface` | 卡片、面板、弹窗底 | 白色或极浅灰 |
| `--rm-bg-muted` | 次级分区、表格行底 | 浅灰 |
| `--rm-text-primary` | 主文本、标题 | 深灰黑 |
| `--rm-text-secondary` | 辅助说明、时间、metadata | 中性灰 |
| `--rm-border-soft` | 卡片和分区边框 | 低对比灰 |
| `--rm-brand-primary` | 主按钮、当前焦点 | 克制蓝或青绿 |
| `--rm-state-success` | ready、connected、passed | 低饱和绿 |
| `--rm-state-warning` | stale、warning、not ready | 低饱和橙 |
| `--rm-state-danger` | failed、danger、unsafe | 克制红 |
| `--rm-state-blocked` | blocked、disabled by policy | 橙或红的柔化底 |

实现约束：

- 不在组件 CSS 中持续新增硬编码颜色。
- 状态色只表达状态，不作为装饰色使用。
- 同屏强强调色不超过两组。
- 卡片、按钮、标签和诊断提示必须使用 token 角色，而不是按页面各自取色。

## 字体与层级

默认使用系统无衬线字体或项目已有前端字体栈。首版不引入复杂字体系统。

推荐层级：

- 页面标题：`32-40px`，仅用于页面主标题。
- 区块标题：`16-20px`，用于面板标题。
- 正文：`13-15px`，用于说明和列表。
- 元信息：`12-13px`，用于 endpoint、profile、timestamp、evidence。
- 指标数字：`24-40px`，只用于关键摘要，不在密集面板滥用。

文本规则：

- 不使用负 letter spacing。
- 英文 endpoint、provider id、profile id 允许等宽或更紧凑显示，但不得溢出容器。
- 页面文案默认短句化，长解释进入详情、tooltip、audit note 或文档链接。

## 组件规范

### 导航

导航只表达页面分组，不承担状态解释。当前 active 项必须清楚，但不使用强装饰背景。

### 状态条

状态条必须展示：

- 当前总体状态：`ready / loading / stale / failed / blocked`
- 数据来源或 route
- `last refresh` 或 `last good`
- refresh / retry 入口

### 卡片和面板

- 圆角默认 `6-8px`。
- 边框优先于重阴影。
- 卡片用于独立信息对象，不嵌套卡片。
- 面板标题、说明、主体内容和证据行必须有稳定顺序。

### 标签与徽标

标签用于短状态，不承载长解释。

必备标签语义：

- `READY`
- `LOADING`
- `STALE`
- `FAILED`
- `BLOCKED`
- `READ ONLY`
- `REQUIRES CONFIRMATION`
- `NOT READY`

### 按钮

当前 P3 只允许低风险操作：

- refresh
- retry
- copy diagnostics（后续可选）
- view details（后续可选）

不得出现：

- execute
- confirm write
- replay
- apply to upstream
- run tool

除非对应 confirmation / executor / storage / negative gate 已正式完成。

## 状态矩阵

每个核心页面或面板至少考虑以下状态：

| 状态 | UI 表达 | 边界 |
| --- | --- | --- |
| `ready` | 正常显示只读数据和证据 | 不暗示生产可用 |
| `loading` | 保留布局骨架，标记正在读取 | 不清空上一份可读数据，除非首次加载 |
| `stale` | 保留 last good snapshot，明显标记 stale | 不允许执行动作 |
| `failed` | 展示失败类别、route、hint 和可重试入口 | 不升级为业务事故 |
| `blocked` | 显示 blocked reason 和 stop-line | 不提供执行按钮 |
| `requires_confirmation` | 标出未来确认流位置 | 当前不得伪实现确认 |
| `empty` | 说明没有可展示数据或当前配置未启用 | 不用空白页面 |

## 页面范围

首批 Pencil 与 React 实现必须覆盖：

1. `Local Overview`
   - service status
   - route summary
   - provider/profile summary
   - stop-line summary
2. `Local Readiness`
   - healthz
   - overview contract
   - local-smoke
   - CORS / port / no-side-effects
3. `Provider/Profile Inventory`
   - selectable model id
   - default provider/profile/model
   - selector boundary
   - 不声明 credential readiness 或 provider health
4. `Session & Tooling Surface`
   - session metadata
   - tool registry
   - blocked action
   - audit summary
5. `Failure / Stale View`
   - last good snapshot
   - failure class
   - local diagnostic hint
6. `Settings / Permissions`
   - provider config placeholder
   - permission and confirmation policy placeholder
   - local environment note
   - 当前只做未来页面规范，不声明已实现
7. `Narrow Layout`
   - `1080x1920` 竖屏基准
   - 单列信息优先级

## Pencil 设计稿治理

`.pen` 是 UI 设计源文件，必须进入 `docs/designs/`。

设计稿要求：

- 文件名表达端点和职责。
- 主桌面稿使用 `1920x1080`。
- 窄屏稿使用 `1080x1920`。
- 每个主要页面标明 ready、failed/stale、blocked 或 loading 的至少一种对应状态。
- 设计稿中的控件必须区分当前可用、只读展示、未来确认流和明确 blocked。
- 设计稿评审通过前，不把当前 console 壳扩成正式 UI 或大面积重构。

实现映射要求：

- 设计稿定稿后，拆成小的 React + Vite + TypeScript 实现任务。
- 每个实现任务说明对应的 `.pen` 页面、涉及组件、可复用 token、验证入口和不做范围。
- 若实现偏离设计稿，应先更新设计稿或在任务文档中记录原因。

## 验收标准

一次 UI 设计或实现改动至少满足：

- 状态层级清楚，失败、过期和 blocked 不混淆。
- 无文本溢出、遮挡或错位。
- 颜色、圆角、间距和状态表达能映射到语义 token。
- 未实现能力没有被画成可点击执行路径。
- 窄屏布局不是桌面缩放，而是重新排序后的单列信息结构。
- Pencil 结构检查无 layout problems；必要时导出 PNG 进行视觉复核。
- React 实现阶段至少跑对应 console behavior / visual smoke / fast baseline。

## 与 Radish 的关系

RadishMind 可以借鉴 Radish 的 UI 治理方式：先有规范、token、设计源文件和实现映射，再进入页面重构。

但 RadishMind 不继承 Radish 的淡雅新中式视觉主题。RadishMind 的默认气质是安静、工程化、可审计的本地 AI runtime 工作台。
