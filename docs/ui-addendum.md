# RadishMind UI 差异附录

更新时间：2026-07-30

采用基线：RadishX `docs/design/family-ui/` `v26.7.2`

Profile：`Workbench`

## 文档职责

本文件只记录 RadishMind 相对 family-ui 的项目差异、专属组件和迁移偏差。家族通用色彩、字体、间距、圆角、阴影、图标、组件和平台布局规则不在本仓库复制维护。

规则优先级：

1. RadishMind 功能设计文档中的能力边界、状态事实与停止线。
2. family-ui 当前采用版本的家族通用规范。
3. 本差异附录中的项目语义、专属组件和已登记迁移偏差。
4. 页面实现与历史设计稿。

实现与前三者冲突时，应先判断是实现偏离还是规范需要升级，再统一修正。

## 项目辨识

RadishMind 是 Radish 家族中的 AI 工具、工作流、模型网关和 Copilot 集成工作台：

- 视觉主气质：现代、克制、可信、可审计。
- 主交互强调：墨蓝，对应 `--rd-action-*`。
- 次级正向强调：玉色，对应 `--rd-accent-jade` 与 success 语义。
- 小面积辅助区分：紫色，对应 `--rd-accent-purple`。
- 品牌胭脂色不作为日常主按钮颜色，也不用于大面积工程工作台背景。
- Workbench 不使用家族纹样；`--rd-pattern-line` 仅随 token 镜像存在，不进入当前页面设计。

### 产品面表达

- 页面、区域和卡片标题只承载稳定名称，不自动生成 eyebrow、代表状态、营销句或解释性副标题。
- 状态、范围、来源数量、窗口限制和 owner 进入 badge、字段或数据行；只有直接影响判断或操作的说明才进入正文内容区。
- 常驻导航、普通主操作和边界摘要不使用 `--rd-bg-ink` 形成大面积高反差块，默认通过 `--rd-bg-*` 层级、`--rd-border-*` 和柔底状态色表达。
- 同一层级只允许一个导航 owner；没有真实跨产品层级时，不并列使用图标产品轨和文字侧栏。
- 高密度工作列表不得使用弹性行高填满视口；行高由必要内容与密度模式决定，相同页面族保持一致。

## Token 与构建映射

上游原样镜像：

- `apps/radishmind-web/src/styles/family-ui/tokens.css`
- `apps/radishmind-web/src/styles/family-ui/tokens.json`

项目别名：

- `apps/radishmind-web/src/styles/radishmind-aliases.css`

导入顺序：

1. family-ui `--rd-*` L1 token。
2. RadishMind `--rm-*` L2 alias。
3. 页面与组件样式。

项目组件优先消费 `--rm-*` 或直接消费稳定的 `--rd-*` 语义层，不直接消费 `--rd-palette-*` 调色板层。

### L2 语义约定

| RadishMind 语义 | family-ui 映射 | 用途 |
| --- | --- | --- |
| `--rm-brand-primary` | `--rd-action-primary` | 主按钮、选中态、关键交互 |
| `--rm-brand-soft` | `--rd-action-soft` | 当前上下文、轻量选中背景 |
| `--rm-accent-secondary` | `--rd-accent-jade` | 正向辅助、健康或已满足证据 |
| `--rm-accent-tertiary` | `--rd-accent-purple` | 小面积类型区分 |
| `--rm-state-*` | `--rd-state-*` | success / warning / danger / info / neutral |
| `--rm-bg-*`、`--rm-text-*`、`--rm-border-*` | 对应 `--rd-*` 语义 | 页面、面板、文本与边界 |

`--rm-brand-primary` 的名称保留现有项目词汇，但语义是 Workbench 的主交互色，不等同于 family-ui 品牌面 `--rd-brand-primary`。

## RadishMind 专属组件

以下组件可以拥有项目层结构与变量，但仍必须复用家族基础 token、交互态和无障碍规则：

- Workflow Node Designer：画布、节点、端口、连线、缩放与 inspector。
- Readiness / Stop-line / Evidence Panel：展示来源覆盖、满足度、阻塞原因和停止线。
- CAS / Drift Confirmation：显示预期版本、当前版本、漂移影响与人工确认。
- Ephemeral Credential Handoff：只展示一次的凭据交接、复制状态与离开后不可恢复说明。
- Gateway Stream / Tool Trace：流式输出、usage availability、受控工具调用与失败审查。
- Evaluation / Release Review：baseline、comparison、case / suite、人工 decision 和不可自动发布边界。

专属组件不得复制领域 owner，也不得从 UI 推导生产 readiness、全量统计或自动执行资格。

## 当前迁移偏差

### 已接受的临时偏差

- 现有 `styles.css` 仍包含大量冷灰、深海军蓝和硬编码间距；基础批次只接入 token 与字体，不做视觉突变。
- 现有组件尚未全面消费 `--rm-*` / `--rd-*`；只允许在后续页面纵向切片中迁移，不继续为新设计增加无语义硬编码颜色。
- `S1` 实现已经消除 `390px` 窄屏原有的 `421px` 横向内容宽度。根因是 Application Configuration Draft 深层标题行中的长状态 badge；移动端共享收缩规则、标题行换行与 React Flow 边界已收口，闭合导航和展开菜单的真实页面宽度均为 `390px`。
- family-ui 已包含暗色映射，但 RadishMind 尚未完成暗色页面设计、切换策略和双态视觉验收，因此当前不提供暗色主题开关。
- 旧 [UI 设计规范](radishmind-ui-design-spec.md) 暂作为历史迁移源保留；其中家族通用视觉规则已由 family-ui 取代，领域状态和产品边界逐步迁入功能专题与本附录。

### 不接受的偏差

- 颜色成为状态的唯一通道。
- 用新的私有硬编码色规避语义 token。
- Workbench 页面使用品牌纹样、装饰性大标题或营销面布局。
- 在稳定标题周围追加临时说明、代表状态或装饰性英文 kicker。
- 用大面积墨色导航、普通按钮或说明卡制造不必要的对比，或用弹性行高稀释工作列表密度。
- 用并列图标轨和文字侧栏重复表达同一组主导航入口。
- 为未来能力预留看似可执行的按钮、菜单或成功状态。
- 一次性凭据、原始请求正文、认证头或敏感 provider 信息进入日志、URL 或持久化 UI 状态。

## 维护流程

1. family-ui 版本升级时，先阅读上游 migration playbook。
2. 原样更新 CSS / JSON 镜像并核对内容一致性。
3. 审查 `--rm-*` 映射、专属组件和已登记偏差。
4. 执行 Web 测试、production build、桌面与窄屏 smoke。
5. 更新本文件、产品化专题、当前焦点与周志。
6. RadishMind 验收完成后，再由 RadishX 维护者更新上游 adoption 状态；本仓库不替代上游真相源。
