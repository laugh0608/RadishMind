# RadishMind Family UI 产品化设计与迁移 v1

更新时间：2026-07-30

状态：`radishmind_family_ui_productization_v1_reference_mapping_completed`

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
- 设计源：继续使用 `docs/designs/` 下的 Pencil 文件；下一批开始前先通过 Pencil 工具盘点既有设计稿，再决定更新现有源文件或创建新的 v1 产品化源文件。
- 外部参考：27 张 family-ui 参考图的逐项采用与排除边界已固定在 [Family UI 参考图产品面映射 v1](radishmind-family-ui-reference-mapping-v1.md)；Pencil 只标注 `ref-XX` 与吸收原则，不嵌入或复制外部截图。

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

参考图产品面映射已完成。下一步在 Pencil 中先完成 `S1` 产品壳的信息架构、关键页面和状态矩阵，再依次进入 Application Workspace、Saved Draft / Designer、API Integration / Key。设计评审必须逐项映射到 `ref-XX`、现有 owner、consumer 和停止线，不能用静态理想稿代替真实交互状态。

### 实现批次：纵向切片迁移

按“产品壳 → Application Workspace → Saved Draft / Designer → API Integration / Key”的顺序逐批实施。每批同时完成：

- 对应页面和共享组件的语义 token 迁移；
- 桌面与窄屏行为；
- 键盘、焦点、状态第二通道和敏感信息边界；
- 单元测试、production build、真实浏览器路径和必要仓库门禁；
- 删除已被替代的遗留声明，避免新旧样式永久并存。

## 基础浏览器基线

- `1440x900`：页面根宽度与 viewport 一致，Workbench 计算值为 `--rd-bg-app: #f7f4ee`、`--rd-bg-surface: #fffdf8`、`--rm-brand-primary: #435c74`，控制台无 warning / error。
- `390x844`：导航已经按单列重排，但完整长页面的既有 `scrollWidth` 为 `421px`。首个可见根因是 Application Configuration Draft 深层卡片中的长状态 badge 在窄容器内保持固有宽度；切回本批之前的字体栈仍为 `421px`，因此不是 token 或字体接入造成的新回归。
- 首批 Pencil 设计和对应纵向实现必须消除上述窄屏超宽，并重新组织桌面 read-shell 状态卡的信息密度；基础批次只记录基线，不混入视觉重构。

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
- Playwright `1440x900` 与 `390x844` 复验完成，控制台零 warning / error；窄屏既有超宽已用旧字体对照确认不是本批回归。
- `./scripts/check-repo.sh --fast` 与 `./scripts/check-repo.sh` 均通过；只保留 W28–W30 历史周志的既有篇幅 warning。

## 停止线

- 基础批次不做 `styles.css` 全量换色、组件重写或页面重新布局。
- 不因 family-ui 已提供暗色 token 就声明暗色主题切换可用。
- 不新增 API、schema、migration、repository、生产认证、quota / billing 或执行能力。
- 不把开发测试态 owner、离线 evidence、首分页窗口或人工审查画成生产就绪、全量统计、自动执行或业务写回。
- 不为普通 UI 迁移新增 task card、fixture 或专项 checker；优先复用 Web 测试、build、consumer smoke 和仓库门禁。
