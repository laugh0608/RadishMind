# RadishMind UI 差异附录

更新时间：2026-08-29

采用基线：RadishX `docs/design/family-ui/` `v26.7.3`

Profile：RadishMind 主动选择 `Workbench`

## 文档职责

本文件只记录 RadishMind 相对 family-ui 的项目选择、专属组件和迁移偏差。family-ui 提供通用参考基线，不替具体项目规定配色分工、技术接入、采用状态或迁移节奏；RadishMind 的这些决策由本文件负责。

规则优先级：

1. RadishMind 功能设计文档中的能力边界、状态事实与停止线。
2. family-ui 当前采用版本的家族通用规范。
3. 本差异附录中的项目语义、专属组件和已登记迁移偏差。
4. 页面实现与历史设计稿。

实现与前三者冲突时，应先判断是实现偏离还是规范需要升级，再统一修正。

## 项目辨识

RadishMind 是 Radish 家族中的 AI 工具、工作流、模型网关和 Copilot 集成工作台：

- 视觉主气质：现代、克制、可信、可审计。
- 身份识别：灰玉，对应 `--rd-brand-*`；只用于产品身份与明确品牌实底，不承担状态或日常主操作。
- 主交互强调：墨蓝，对应 `--rd-action-*`。
- 次级正向强调：玉色，对应 `--rd-accent-jade` 与 success 语义。
- 小面积辅助区分：紫色，对应 `--rd-accent-purple`。
- 小面积关注提示：胭脂，对应项目级 `--rm-attention-*`；不等同于品牌、danger 或自动发布资格。
- Workbench 不使用家族纹样；`--rd-pattern-line` 仅随 token 镜像存在，不进入当前页面设计。

### 产品面表达

- 页面、区域和卡片标题只承载稳定名称，不自动生成 eyebrow、代表状态、营销句或解释性副标题。
- 状态、范围、来源数量、窗口限制和 owner 进入 badge、字段或数据行；只有直接影响判断或操作的说明才进入正文内容区。
- 常驻导航、普通主操作和边界摘要不使用 `--rd-bg-ink` 形成大面积高反差块，默认通过 `--rd-bg-*` 层级、`--rd-border-*` 和柔底状态色表达。
- 同一层级只允许一个导航 owner；没有真实跨产品层级时，不并列使用图标产品轨和文字侧栏。
- 页面应有一个可直接识别的主对象；次级证据、状态和边界优先贴合主对象组织，不用等权小卡把工作流分碎。
- token 一致只是必要条件，不构成视觉继承。S8 及更早已审 Workbench 的连续产品导航、薄页眉、对象路径、唯一 owner、连续事实行和单一辅助 rail 是后续工作面的直接结构基准；不得在后续页面重新引入宽 hero、另一套侧栏尺寸、KPI 卡带或稀疏多卡骨架。
- 当前导航、当前阶段和当前任务可用语义表面、细边界和柔和阴影建立抬起实体感；阴影不应普遍施加到每个小区域。
- 可审计不等于微缩；正文、图标、控件和行高先满足可读性，再按任务密度收紧。高密度工作列表不得使用弹性行高填满视口；相同页面族保持一致。

## Token 与构建映射

RadishMind 当前选择精确镜像所采用 family-ui 版本的参考实现；这是项目接入策略，不是上游对采用方的强制要求：

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
| `--rm-identity-*` | `--rd-brand-*` | 产品身份、品牌柔底与身份悬停 |
| `--rm-text-on-identity` | `--rd-text-on-brand` | 品牌实底上的可读前景 |
| `--rm-action-*` | `--rd-action-*` | 主按钮、链接、选中态与关键交互 |
| `--rm-attention-primary` | `--rd-accent-rouge` | 小面积关注计数或需要人工留意的非状态提示 |
| `--rm-attention-soft` | 从 `--rd-accent-rouge` 派生的项目柔底 | 关注提示背景，不复用 danger 或 brand 柔底 |
| `--rm-accent-secondary` | `--rd-accent-jade` | 正向辅助、健康或已满足证据 |
| `--rm-accent-tertiary` | `--rd-accent-purple` | 小面积类型区分 |
| `--rm-state-*` | `--rd-state-*` | success / warning / danger / info / neutral |
| `--rm-bg-*`、`--rm-text-*`、`--rm-border-*` | 对应 `--rd-*` 语义 | 页面、面板、文本与边界 |

`v26.7.3` 新增 `--rd-text-on-brand` 并把参考实现中的默认品牌识别改为灰玉。RadishMind 不再用名为 brand 的项目别名承载 action；identity、action 与 attention 必须分别映射，避免上游品牌值变化静默改写交互或关注语义。

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
- `v26.7.3` token 已完成精确镜像；`S1` 产品身份改为灰玉，墨蓝 action 与胭脂 attention 通过项目别名保持原有职责，不再直接混用 `--rd-brand-primary`。
- `S1` 实现已经消除 `390px` 窄屏原有的 `421px` 横向内容宽度。根因是 Application Configuration Draft 深层标题行中的长状态 badge；移动端共享收缩规则、标题行换行与 React Flow 边界已收口，闭合导航和展开菜单的真实页面宽度均为 `390px`。
- `S1 R5` / `S2 R3` 的人工评审结论为现代产品感不足；`S1 R6` / `S2 R4` 修正了整体方向，`S1 R7` / `S2 R5` 取消了无职责的桌面外圈容器，但局部仍偏规整文字面板。`S1 R8` / `S2 R6` 已用 Operations Inbox 紧凑窗口与 evidence path、Source evidence 分布矩阵、当前阶段轻量选中轨、十三段 contribution window、九格 readiness 与 authorization path 增强信息密度和视觉焦点；选中只归属于当前导航或详情 owner，普通状态项不得因 `missing`、`blocked` 或 `partial` 获得选中底色。2026-08-06 人工复评已通过，`S1 R8` React 也已按真实 view model 落地并完成严格浏览器验收；设计中的代表计数不覆盖运行时计数。当前进入 `S2 R6`，不把尚未实现的 Application Workspace 新基准写成已落地产品行为。
- S9 / S10 原 R1 的首次返工只把独立米棕色板映射回项目 token，未改变宽 hero、等权卡片、稀疏工作面和另一套侧栏，因此被人工退回；Visual R2 虽完成连续 Workbench 结构归位，又把上下文、任务、owner 与 boundary 全部处理成硬方形，仍未继承 S7 / S8 和 `reference-ui` 的形态语言，因此再次被退回。2026-08-10 已完成 Visual R3：连续窗格与表格事实行保持方正发丝边界，业务表面使用 `8–11px` 职责圆角，紧凑控件使用 `7–8px`，标签使用全圆角。S9 与 S10 Visual R3 已于 2026-08-12 人工通过；已有 React 功能证据仍不得冒充逐项采用 Visual R3。
- 结构化输入局部稿的 Visual R3 虽满足上述圆角层级，字段区仍像用整页横线切出的静态列表，因此继续修订为 Visual R4。类型化编辑器使用带留白层级的表单画布：长短字段按任务关系非对称编排，文本、数值和布尔值呈现真实控件，错误与帮助文字归属于精确输入；不可变合同、authority 与易失值边界继续分离。Visual R4 已于 2026-08-10 通过人工复核并冻结，但不表示共享 React editor 已实现。
- Gateway Provider Attempt Visual R1 继续复用 S7 / S5 的连续 Workbench：桌面使用资源路径、单一主 owner 和一个 evidence rail，窄屏固定 `context → plan → attempts → cost / boundary`；主 / 备 target 主要由顺序、角色文字和连续 attempt 行区分，只有当前审查对象使用墨蓝选中轨。失败、quota 阻断与 partial cost 同时使用稳定文字、符号和语义状态，不依赖颜色。七个根节点已完成结构与实际渲染复核，并于 2026-08-13 获得人工视觉批准；批准不表示 React 已采用。
- Authentication Self-Service Security Visual R1 延续 Authentication Gateway 的简化 shell 与 Family UI token：Desktop `pOLcz` 把 session directory 作为唯一主对象，Narrow `LMi7H` 以纵向 bulk confirmation 重排，danger state `n2O8A5` 单独承接 credential rotation 与 forced re-login，Decision `DASE0` 固定五维评分、隐私边界和状态失效规则。四张根画板已完成结构与实际渲染复核，并于 2026-08-25 获得人工视觉批准；2026-08-26 React 已在既有 Gateway owner 内采用该结构，以单一 strict consumer、语义 token、`900px` / `620px` 响应式边界和 metadata-only invalidation 落地。同日 in-app Browser 完成 `1440×900`、`720×900`、`390×844`、双标签、console / network / URL / storage / cookie 审计；真实窄屏发现并修复单列 grid implicit row 压缩裁切 session row，以及 account trigger 被 sticky mobile navigation 覆盖两个根因。复验后三视口无横向溢出、账户入口可命中，Visual R1 的 React 采用与浏览器验收均已关闭；这些证据仍只属于开发 / 测试态，不构成 production auth、设备管理或全局 session console。
- Workflow Template Catalog & Derive Visual R1 复用既有 Workflow Workbench，不建立 S11。Catalog `wZH1x / Nb7a5`、Candidate Review `aEaOP / LdtUe`、Version / Listing `Ti8u9 / mac8B`、Derive `qa6Il / e0062U` 与 R22 Decision `bhH8F` 覆盖 pending、rejected、approved-unlisted、listed、replace conflict、unlist danger、binding unavailable 与 store unavailable。R22 五维评分为 `2 / 1 / 2 / 2 / 2 = 9`；9 个根画板、1049 个节点的 Pencil 原生静态 QA 为 0 问题，并于 2026-08-27 获项目所有者人工批准。2026-08-29 React 已在 Human Promotion 接入 Catalog / Review / Listing / Derive 四任务面板，采用 `900px` / `620px` 响应式边界与语义 token；批次 E 又以 `1440×900`、`720×900`、`390×844` 真实浏览器复验无横向溢出或控件裁切，workspace / application 切换无旧 scope 残留，最终 console 无 warning / error。
- Action Safety Ladder Visual R1 继续复用既有 Family UI / Workflow Workbench，不建立 S11。Builder Desktop `JtBu1` / write-blocked Narrow `EIWxV`、Candidate Review Desktop `l3dr1K` / policy-drift Narrow `wyqof`、Tool Plan & Confirmation Desktop `N0jBm` / confirmation-missing Narrow `CoI4i`、Run History Desktop `rHr7a` / legacy Narrow `XRRpD` 与 R23 Decision `OvCRE` 覆盖六级 ladder、scope denied、Tool unavailable、exact pre-dispatch recheck、零副作用 fail-closed 与冻结历史不反算。R23 五维评分为 `2 / 1 / 2 / 2 / 2 = 9`；9 个根画板、1055 个节点的 Pencil 原生静态 QA 对 placeholder、布局裁切、节点命名和文字填充均为 0 问题，并于 2026-08-29 获项目所有者人工视觉与边界批准。2026-08-30 React 已把代表业务标签映射到真实 Agent Session / Assignment、Workflow HTTP Tool plan / execution 与 Run History owner，使用单一 strict consumer 和共享只读面板；`1440×900`、`720×900`、`390×844` 无横向溢出，窄屏标题与状态自动换行，console 无 warning / error。该实现仍不启用 `write_allowed_by_policy`、通用 handoff 或 production execution。
- family-ui 已包含暗色映射，但 RadishMind 尚未完成暗色页面设计、切换策略和双态视觉验收，因此当前不提供暗色主题开关。
- 旧 [UI 设计规范](radishmind-ui-design-spec.md) 暂作为历史迁移源保留；其中家族通用视觉规则已由 family-ui 取代，领域状态和产品边界逐步迁入功能专题与本附录。

### 不接受的偏差

- 颜色成为状态的唯一通道。
- 用新的私有硬编码色规避语义 token。
- Workbench 页面使用品牌纹样、装饰性大标题或营销面布局。
- 在稳定标题周围追加临时说明、代表状态或装饰性英文 kicker。
- 用大面积墨色导航、普通按钮或说明卡制造不必要的对比，或用弹性行高稀释工作列表密度。
- 把正文、图标、控件和任务行压缩成微缩审计界面，或使用等权卡片墙取代一个明确的当前任务主对象。
- 用并列图标轨和文字侧栏重复表达同一组主导航入口。
- 为未来能力预留看似可执行的按钮、菜单或成功状态。
- 一次性凭据、原始请求正文、认证头或敏感 provider 信息进入日志、URL 或持久化 UI 状态。

## 维护流程

1. family-ui 版本升级时，先阅读上游 README、Changelog、规范扩展与版本兼容原则，并检查是否标记为破坏性变更。
2. 由 RadishMind 决定是否升级；采用升级时精确更新 CSS / JSON 镜像并核对内容一致性。
3. 审查 `--rm-*` 映射、品牌与操作语义、专属组件和已登记偏差，不让参考实现的值变化静默改变业务表达。
4. 执行 Web 测试、production build、桌面与窄屏 smoke。
5. 更新本文件、产品化专题、当前焦点与周志。采用状态只在 RadishMind 维护；只有新的跨界面通用语义才回到 family-ui 讨论。
