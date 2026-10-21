# 应用开发工作区与发布准备审查 v1

更新时间：2026-10-21

状态：`application_development_workspace_release_readiness_review_v1_completed`

原工作区组合批次保持关闭。2026-10-21 新增 Workflow 创建、版本迭代与回归审查走查及实施方案，见本文对应章节；本轮只形成问题证据和设计范围，产品修复与新增编辑能力尚未实施。

## 功能目标

为内部应用开发者建立一个以单一 Application 为作用域的连续工作区，把现有应用目录、配置草案、Workflow、RAG、发布候选、受控测试、运行历史、评测与运行观测组织为可理解、可交接、可失败关闭的开发路径。

本功能解决的是既有能力分散、页面入口并列、状态所有权不清晰和审查证据难以连续理解的问题，不建立新的应用、发布、运行或评测真相源。发布准备结论只是对既有权威资源和当前可读证据的只读投影，不代表正式发布资格，不触发 activation、assignment、release 或执行。

## 当前设计结论

- 首个纵向切片只需要 Web 侧的信息架构、application context、阶段导航、状态清理与只读投影；现有 API、schema、repository 和 strict consumer 足以承载，不创建实现任务卡。
- `apps/radishmind-web/src/app/App.tsx` 继续负责产品壳和一级产品面入口；应用开发流程下沉到独立 feature surface，不再由根组件直接编排全部 Application / Workflow 面板状态。
- 各领域 owner 继续拥有写入、状态机、CAS、权限、authority、失败码和持久化。工作区只保存当前 application ref、当前阶段和脱敏资源引用，不复制领域载荷。
- 发布准备投影只允许输出“尚未审查”“证据不完整”“存在阻塞”“开发测试态证据可审查”；不得输出 `production_ready`、`publish_ready` 或等价声明。
- 现有应用配置、候选审查、definition activation、RAG assignment、Session、Run History、Evaluation 和 Operations 功能保持原有停止线，不从本专题派生生产化、自动化或新的执行算法。

## 目标用户与核心任务

目标用户是正在 RadishMind 内部开发者预览环境中配置、构建、测试和审查 AI 应用的应用开发者或人工审查人。

用户需要完成六项连续任务：

1. 选择一个权威 Application，确认生命周期和当前作用域。
2. 配置应用并构建 Workflow 或 RAG 资源，理解草案、绑定和定义的当前状态。
3. 通过显式人工动作审查候选、Definition 或知识基线，并识别漂移与阻塞。
4. 选择明确的 Workflow Definition 或 Application RAG profile 进行受控测试。
5. 审查 Session、Run、Comparison、Evaluation、Gateway Request 与 Operations evidence。
6. 汇总既有来源的状态与缺失项，形成开发测试态发布准备判断和下一跳。

## 用户流程与页面阶段

| 阶段 | 用户问题 | 允许动作 | 只读证据 | 失败或缺失时 |
| --- | --- | --- | --- | --- |
| 选择应用 | 当前操作的是哪个 Application，是否仍活跃 | 选择、刷新、创建或更新开发测试态 Application | lifecycle、owner、application kind、revision | 无应用时阻塞后续阶段；归档应用只允许历史审查 |
| 配置 / 构建 | 当前配置、Workflow 与 RAG 资源是什么 | 编辑和保存既有草案、创建明确的 Workflow / RAG 候选 | draft version、validation、binding、definition summary | 校验失败、冲突或 scope 漂移时保留在本阶段 |
| 人工晋级 | 哪个不可变版本经过了什么审查 | 调用既有 approve / reject / request changes / activate / assign 等显式动作 | candidate、review、activation、assignment、digest、blockers | 任一 owner 拒绝、漂移或 CAS 冲突都不得进入成功投影 |
| 受控测试 | 当前以哪个明确 authority 运行 | 创建 Session、提交 turn，或调用既有受控测试入口 | exact Workflow v5 / Application RAG v4 authority、run ref | authority 漂移、取消、失败或 outcome unknown 显式展示，不重试或重放 |
| 运行 / 评测审查 | 本次测试发生了什么，是否有可比较证据 | 查看详情、比较、运行既有 evaluation / suite 审查 | Run History、Comparison、Evaluation、Suite、Request History、Operations | 来源失败相互隔离，局部失败不得伪装完整证据 |
| 发布准备判断 | 还缺什么，下一步回到哪里 | 打开对应既有 owner 的阶段或详情 | 各来源状态、精确引用、漂移与阻塞项 | 只输出 blocker 与下一跳，不创建发布记录或自动执行动作 |

阶段导航只表达用户流程，不改变领域状态。用户可以返回任意已开放阶段；导航本身不得把阶段标记为完成，也不得把页面访问解释为审查通过。

## 权威来源与现有页面审计

| 资源或能力 | 权威 owner | 当前 Web 入口 | 工作区处理方式 |
| --- | --- | --- | --- |
| Application lifecycle | `ApplicationCatalogRepository` 与 lifecycle service | `ApplicationCatalogPanel`、Applications 列表 | 保留唯一 application context；活跃与归档决定后续动作可用性 |
| 应用配置草案 | Application Configuration Draft repository / service | `ApplicationConfigurationDraftPanel` | 原样复用校验、保存、恢复、比较和 CAS；只交接 draft ref 与脱敏选择 |
| 应用发布候选 | Application Publish Candidate repository / governance service | `ApplicationPublishCandidatePanel` | 原样复用不可变候选、审查和 eligibility；不把 approve 映射为正式发布 |
| Saved Workflow Draft | Saved Draft repository / service | Workflow designer 与 Saved Draft consumer | 保留草案编辑和保存 owner；工作区只组织入口和当前 application scope |
| Workflow Definition | Definition candidate / version / activation owner | `WorkflowDefinitionPromotionPanel` | 保留人工 review、activation 与 v5 authority；不复制 release 或 executor 逻辑 |
| RAG 知识与绑定 | snapshot、dataset、promotion、binding owner | RAG snapshot / evaluation / promotion panels | 保留 exact snapshot、candidate、binding 和知识 authority；不复制 ranking |
| Application RAG assignment | Application RAG runtime owner | `ApplicationRAGInvocationPanel` | 保留显式 assignment 与 v4 invocation；发布候选 approve 不自动 assignment |
| Application Session / Turn | Application Session / Turn repository 与 coordinator | `ApplicationInteractionSessionPanel` | 保留 strict metadata、易失 transcript、取消和 run handoff；不恢复正文 |
| Workflow Run | Workflow Run Store | `WorkflowRunHistoryPanel` | 保留分页、详情、failure review 和 run refs；不建立聚合运行副本 |
| Comparison / Evaluation / Suite | 既有 comparison、case、baseline、suite owners | Run History 内延迟加载面板 | 复用 exact run / source compatibility；不在工作区重新计算评测结论 |
| Gateway Request | Gateway Request History | Playground 与 Request History | 只交接 `application_id + request_id`；不复制请求、响应或 provider 原文 |
| Application Operations | Gateway 与 Workflow 当前分页窗口的客户端投影 | `ApplicationOperationsPanel` | 继续标明来源覆盖与 `partial_failure`；不推测跨来源关联、成本或全量统计 |
| API key | API Key lifecycle owner | `APIKeyLifecyclePanel` | 作为受控 Gateway 测试的辅助资源；原始 token 仍只允许一次性内存交接 |

审计确认当前根组件混合承担以下职责：Application 选择、Workflow 草案编辑状态、运行刷新、多个延迟加载面板挂载，以及基于组件 `key`、回调、`CustomEvent` 和 hash 的跨面板交接。首批实现应收口这些编排职责，但不重写已通过验证的领域 consumer。

## 信息架构

用户工作区保留 Applications 一级入口。选中 Application 后展示一个 Application Development Workspace，内部由以下稳定区域组成：

1. **Application Context Header**：名称、稳定标识、生命周期、application kind、当前 revision 和环境边界。
2. **Stage Navigation**：配置 / 构建、人工晋级、受控测试、运行 / 评测审查、发布准备五个阶段；“选择应用”由上层 Applications 入口承担。
3. **Stage Surface**：挂载当前阶段的既有面板或组合视图；非当前阶段不因隐藏而自动执行请求或写入。
4. **Evidence Rail**：只展示当前阶段已选择的脱敏资源引用、来源状态和明确下一跳，不复制详情载荷。
5. **Release Readiness Review**：按来源展示状态、阻塞项、漂移和缺失，不提供正式发布按钮。

当前 URL 只允许稳定页面 / 阶段锚点，不写 token、input、answer、transcript、review reason、完整资源载荷或一次性凭据。首批实现不要求把 application selection 写入 URL 或浏览器存储。

## Application Context 与状态所有权

工作区只建立一个轻量 application context：

- `applicationId`：来自当前权威 Application 选择；
- `displayName`、`applicationKind`、`lifecycleState`、`updatedAt`：来自当前目录投影；
- `generation`：只用于拒绝旧 application / route 产生的迟到响应；
- `activeStage`：当前阶段导航状态；
- 脱敏 handoff refs：`draftId`、`candidateId`、`definitionId`、`bindingId`、`assignmentId`、`sessionId`、`runId`、`requestId` 等稳定短引用。

context 不保存领域对象、请求正文、响应正文、凭据、输入、输出或评测载荷。各子面板继续拥有自己的表单、pending、conflict、selection 和 strict response state。

以下事件必须开始新的 generation，并取消或拒绝旧 generation 的响应：

- Application 切换、归档、删除或当前 revision 失效；
- 身份 / owner context 变化；
- 离开 Application Development Workspace；
- 开始新的明确 scope restore。

切换时必须清除：一次性 API token、用户 input、answer、transcript、候选 review reason、草案冲突选择、Session / Turn 选择、Run comparison 选择、Evaluation 临时选择、RAG fragment preview、复制状态、pending handoff 和所有迟到响应。持久 metadata 只能由新 scope 下的 existing strict consumer 重新读取。

## 跨阶段交接

跨阶段交接只允许携带 application scope 与稳定脱敏引用。首批实现允许现有事件继续工作，但新 Application Development Workspace 应提供显式的 feature-scoped handoff state 和回调入口，后续按进入工作区的顺序替代散落的 `window.location.hash` 与全局 `CustomEvent`。

交接必须满足：

- 来源和目标的 `applicationId` 一致；
- 目标重新从自己的 owner 读取精确资源，不信任来源提交的完整对象；
- application generation 不一致时直接丢弃；
- 交接只改变目标选择或导航，不自动触发保存、审查、activation、assignment、provider 调用或发布；
- 一次性 token 只交给现有受控调用组件内存，不能进入通用 handoff state。

## 发布准备投影

发布准备投影是 Web 侧只读 view model，不新增服务器 contract 或持久化记录。它按来源独立展示：

| 来源组 | 可采用的既有证据 | 必须展示的缺失或阻塞 |
| --- | --- | --- |
| Application | lifecycle、revision、owner scope | 未选择、归档、scope mismatch、目录不可用 |
| Configuration / Candidate | valid saved draft、candidate review、digest、eligibility | 未保存、invalid、未审查、被取代、草案 / 基线漂移、production blockers |
| Workflow authority | approved immutable version、active definition、exact digest | 无 activation、版本漂移、eligibility 失败、owner 不可用 |
| RAG authority | approved binding、current assignment、exact snapshot authority | 无 binding / assignment、知识来源漂移、snapshot 或 store 失败 |
| Controlled test | Session / Turn metadata、v5 / v4 run ref | 无成功或可审查 run、失败、取消、outcome unknown、authority drift |
| Evaluation | compatible comparison、case / baseline / suite review | 来源不兼容、缺少 baseline / suite、评测 owner 失败 |
| Operations | Gateway / Workflow 当前窗口来源覆盖与失败摘要 | `partial_failure`、窗口不完整、usage 未报告、来源不可用 |

投影状态固定为：

- `review_not_started`：没有足够的已选择来源开始审查；
- `review_incomplete`：已有部分来源，但仍缺少必要开发测试态证据；
- `review_blocked`：存在 owner 返回的失败、漂移、生命周期或 eligibility blocker；
- `dev_test_evidence_reviewable`：当前选择的开发测试态来源可供人工审查。

`dev_test_evidence_reviewable` 只表示页面已经组织出可审查证据，不表示 production ready、publish ready、release approved 或 production authorization satisfied。若无法证明来源完整性，优先保持 `review_incomplete` 或 `review_blocked`。

## 失败语义

- Application owner 不可用或 scope 不一致时，整个工作区失败关闭，不回退 fixture 或旧 Application。
- 单个下游来源失败时只阻塞对应阶段和发布准备投影；其它已加载来源可以继续只读审查，但页面必须显示 `partial_failure` 或等价明确状态。
- CAS、漂移和过期 generation 不得覆盖当前状态；冲突只展示现有 owner 允许的脱敏 metadata。
- owner evidence 即使已经通过回调入口的 scope 检查，也必须在 React state updater 真正应用时再次核对当前 `applicationId` 与 workspace generation；切换 application 后排队中的旧 updater 直接丢弃，不得向新 workspace 抛错或覆盖 readiness。
- 2026-08-09 真实硬加载与快速 application 切换复现了上述竞争窗口：旧 evidence 已通过外层检查，但在新 application state 生效后才进入 updater，原严格断言会让 React 根节点失败。workspace owner 现以同一精确 scope predicate 在 updater 内先丢弃迟到输入；纯函数对直接误用仍保持抛错，浏览器硬加载回归无空白根节点或控制台 error。
- 取消或路由离开后，迟到响应不得重新填充 input、answer、transcript、selection、conflict 或 readiness。
- readiness projection 不定义新的外部失败码；它保留既有 owner failure code，并用来源级 UI 状态组织展示。
- 任一失败路径都不得产生额外 provider 调用、工具调用、confirmation、业务写入、replay、activation、assignment 或 release。

## 隐私与数据边界

- 不新增聚合 API、schema、repository、数据库表或 committed evidence payload。
- 不持久化 input、answer、prompt、provider response、fragment content、token、Authorization、cookie、DSN、endpoint、review reason 或完整请求 / 运行载荷。
- 不把 Gateway 与 Workflow 记录按时间、模型或文本相似度推测为同一次调用；只有既有精确引用可以建立交接。
- 不把当前分页窗口推导为全量统计，不估算 token、cost、quota 或 billing。
- 不从客户端重建 authority、digest、eligibility 或 evaluation 结论；目标 owner 必须重新读取并验证。
- 默认离线模式保持零请求；显式开发测试态模式才允许调用现有 consumer。

## 实施拆分

设计已于 2026-07-20 确认。后续按三个可分割批次实施，不新建任务卡：

### 批次 A：Application Workspace Shell

- 从 `App.tsx` 抽出 Application Development Workspace、唯一 application context、阶段导航与 generation boundary。
- 先挂载现有 Application、Configuration、Publish、Workflow / RAG、Session、Operations 和 Run History 面板，不改变 API 或领域 consumer。
- 统一 application / route switch 清理、pending cancellation 和迟到响应拒绝。

### 批次 B：Evidence Handoff 与发布准备投影

- 建立 feature-scoped 脱敏 handoff refs 和显式下一跳，逐步收口散落的 hash / 全局事件编排。
- 建立来源分组、`partial_failure`、漂移 / 缺失 blocker 和四态发布准备 view model。
- 不增加发布按钮、自动动作或聚合 store。

批次 B 的实现所有权固定如下：

- handoff state 与 readiness contribution state 由 `ApplicationDevelopmentWorkspacePanel` 持有，只在当前 application generation 内存活；stage surface 只消费控制器，不拥有跨阶段状态。
- handoff 只保存 `applicationId`、workspace generation、来源 / 目标阶段、受限 `refKind`、短 `refId` 和目标锚点；API token、协议请求体、模型响应、输入输出与完整领域对象不得进入通用 handoff。
- readiness 不重复请求 owner，也不读取面板内部可变对象；Configuration、Candidate、Workflow authority、RAG authority、Controlled test、Evaluation 与 Operations owner 只通过窄回调上报枚举状态、短引用、版本、coverage、既有 failure code 与脱敏 blocker。
- 同一来源组可以包含多个 owner contribution；Configuration / Candidate 必须分别证明草案和候选，RAG authority 必须分别证明 binding 与 assignment，不能由后写入的单一摘要覆盖另一 owner。
- Application 切换、revision / lifecycle 变化由 workspace generation 物理重建控制器；离开工作区清空 pending handoff 和 readiness contributions，返回后必须由 owner 重新上报。
- 目标 owner 消费 handoff 后重新读取精确资源；handoff 不自动触发保存、review、activation、assignment、provider / tool 调用、运行或发布。

当前 readiness contribution 固定为十三项：`application_lifecycle`、`configuration_draft`、`publish_candidate`、`workflow_definition`、`rag_binding`、`rag_assignment`、`prompt_template`、`prompt_assignment`、`agent_profile`、`agent_assignment`、`controlled_run`、`evaluation_review`、`operations_coverage`。它们归入 Application、Configuration / Candidate、Workflow authority、RAG authority、Prompt authority、Agent Copilot authority、Controlled test、Evaluation 和 Operations 九个来源组；与当前 application kind 无关的 contribution 以 `not applicable / available / complete` 收口，不形成伪 blocker。四态聚合保持保守：未选择 Application 为 `review_not_started`；任一 owner failure、漂移、生命周期 blocker 或 `partial_failure` 为 `review_blocked`；已有 Application 但适用 contribution 未齐为 `review_incomplete`；十三项均按当前 kind 完整收口时才允许 `dev_test_evidence_reviewable`。

### 批次 C：连续链与专题收口

- 覆盖 Application 切换、归档只读、route switch、late response、cancel、offline、partial source failure、v4 / v5 run handoff、Evaluation compatibility 和一次性内容清理。
- 完成真实浏览器连续路径、URL / storage / console / network 审计和现有功能不回归验证。
- 根据实际行为证据更新功能专题、当前焦点、能力矩阵和周志后关闭专题。

若任何批次发现必须新增 API、schema、repository、执行边界或高风险写入，应停止当前批次，先更新本设计并创建唯一的高风险任务卡；不得把新边界隐藏在普通 UI 重组中。

## 2026-08-03 Family UI S2 设计基准面

`radishmind_web_s2_application_workspace_v1` 已在 [Family UI Pencil 设计源](../../designs/radishmind-web-family-ui-v1.pen) 中完成 `R6` 并于 2026-08-08 落地 React，不改变本专题已经关闭的功能范围：

- 桌面 `1440x900` 继承 `S1` 单一产品导航，以 RadishFlow Copilot 这一真实 Application 作用域组织 Application Context、五阶段任务路径、唯一 Human Promotion 当前 surface 和从属 Evidence / Readiness rail；阶段只表达可进入性，不显示伪完成进度。
- 窄屏 `390x844` 固定为“Application Context → 当前阶段选择器 → 当前 surface → Readiness 摘要”，五阶段与九个来源组通过渐进展开查看，不压缩桌面双栏，也不产生横向溢出。
- 代表面使用缺少权威 revision 的离线 Application 和未证明的 RAG assignment 同时固定 `partial` 与 `blocked` 表达；状态均有文字、结构和语义色第二通道，不能被理解为生产 readiness。
- 当前代码审计确认九个来源组、十三项 contribution、generation / surfaceKey、精确 ref handoff 和目标 owner 重读仍是实现真相；Pencil 中的代表性计数、标签和短引用不进入功能契约。
- 共享设计决策记录明确消费 `ref-05`、`ref-11`、`ref-12`、`ref-13`、`ref-14`、`ref-18` 与 `ref-27`，并排除主题切换、事故语境、ETA、自动阶段完成、自动 activation / assignment、发布按钮和五阶段重复画板。
- `R1` 的功能层级通过审计，`R2` 改为连续工作面、图标分段阶段带、时间线 evidence 与贴合详情 rail；联合人工复核仍确认其尺度过小、表面过碎且主次不足。`R3` 用更大的 Application 身份、阶段 rail 和 Human Promotion 主对象，以及选择性抬升感，完成视觉根因修正。
- `S2 R6` 桌面 / 窄屏与共享设计决策记录均通过 Pencil 全树布局检查和联合人工复评，无裁切、重叠或占位节点；React 已复用 `S1 R8` 共享壳层完成纵向切片和桌面 / 窄屏真实浏览器验收。

## 当前实现进度

2026-07-20 已完成批次 A 的壳层基础切片：

- 新增纯 TypeScript application context builder，统一校验 Application 作用域、active / archived / unavailable 生命周期、revision generation 和五阶段可用性；缺少 Application 或生命周期未知时失败关闭。
- 新增独立 `ApplicationDevelopmentWorkspacePanel`，交付 Application Context Header、五阶段锚点导航、归档只读 / 受控测试阻塞语义和显式 `review_not_started` 发布准备停止线。
- `App.tsx` 中 Application 配置、候选、API 接入、API key、RAG、Session、Definition、Operations、Run History 与 Gateway Playground 已统一消费同一个 application context；Application、revision 或生命周期变化通过 generation key 重新挂载作用域组件，旧 transient state 不再跨 generation 复用。
- 新增 5 个 context / generation / lifecycle / stage mapping 精准测试；全部 Web 测试 `167 / 167` 通过，`npm run build` 通过，新模块行覆盖率 `100%`。

随后已完成批次 A 剩余实现：

- 新增 feature-owned stage surface，把 Configuration、Publish、Workflow Definition promotion、RAG、API Integration、API key、Session、Application RAG、Operations 与 Run History 从 `App.tsx` 的直接编排中收口；根组件只交付 application context 和必要的 owner props。
- route model 统一识别现有面板与 handoff hash；当前只挂载一个阶段，离开工作区时卸载 stage surface，Application / revision / lifecycle 或跨阶段变化会形成新的 route / workspace generation，旧 surface key 不再接受迟到结果。
- 跨阶段 Application API 选择改为严格校验、精确 application scope、一次消费的内存 handoff，避免目标阶段尚未挂载时丢失事件；不保存 token、input、answer 或完整领域对象。
- 归档 Application 的配置、候选和 evidence 保持只读，Controlled Test 整段失败关闭；未知或未选择 Application 不挂载写入 / 运行 owner。
- Application Workspace 精准测试增至 8 项，新增一次性 route handoff 测试；全部 Web 测试 `171 / 171`、`npm run build`、route / handoff / Workflow Definition 定向测试均通过。

批次 A 已关闭，批次 B 也已完成。通用 feature-scoped 脱敏 handoff refs、十三项 owner contribution、九个来源组与四态 readiness view model 已进入 Web；既有一次性 Application API token 交接仍保持在专用内存通道，不进入通用 handoff state。下文 2026-07-20 的“九项 / 七组”记录是 Prompt / Agent 扩展前的批次 B 历史证据，不代表当前代码规模。

2026-07-20 已完成批次 B：

- 新增 feature-scoped handoff controller，只允许当前 Application generation 内一个待消费的稳定短引用；新交接替换旧交接，离开工作区立即清空，作用域漂移、代际过期、非法引用和错误目标均失败关闭。
- 批次 B 当时由 Configuration Draft、Publish Candidate、Workflow Definition、RAG Binding、RAG Assignment、Controlled Run、Evaluation 与 Operations owner 通过窄回调上报九项脱敏 contribution；后续 Prompt Template / Assignment 与 Agent Profile / Assignment 沿用同一边界扩展为当前十三项，工作区仍不重复请求 owner，不保存领域对象、输入输出或凭据。
- Draft → Publish Review 与 Run → Run History 已改用通用 handoff；目标 owner 重新读取精确 draft / run，找不到时明确保留在 owner 视图，不信任来源提交的完整对象，也不自动触发审查或运行。
- 批次 B 当时按 Application、Configuration / Candidate、Workflow authority、RAG authority、Controlled test、Evaluation、Operations 七组展示状态、coverage、精确引用、缺失项、blocker、failure code 和下一跳；后续 Prompt / Agent 两组沿同一规则接入，当前十三项 contribution 按 application kind 全部完整收口时才输出 `dev_test_evidence_reviewable`。
- Application 切换、revision / lifecycle 变化和离开工作区会重建或清空 evidence / handoff state；`partial_failure`、owner failure、归档、漂移或 blocker 保守聚合为 `review_blocked`，不会隐藏其它已加载证据。
- Application Workspace 精准测试增至 17 项；全部 Web 测试 `180 / 180`、`npm run build`、`git diff --check` 与仓库快速门禁通过。实现没有新增 API、schema、repository、发布记录、聚合 store、执行算法或专项 checker。

批次 B 完成后随后进入批次 C，范围固定为跨阶段连续路径、归档只读、late response / cancel / offline / partial source failure、v4 / v5 run handoff、Evaluation compatibility、一次性内容清理和真实浏览器隐私审计；批次 C 不扩生产认证、正式发布或自动动作。

2026-07-20 已完成批次 C 的代码与自动验证部分：

- owner evidence report 增加当前 `surfaceKey`，workspace 通过 route state ref 校验最新 surface；同一 Application 跨阶段时保留已经接受的 contribution，但旧 surface 的迟到回调不能再覆盖当前状态。
- Draft → Publish Review 目标 owner 只接受当前 Application 下精确且有效的 draft；目标不存在、无效、scope 不符或 owner 离线时不再回退到其它草案，并明确说明没有创建候选或审查。
- Run → Run History 不再受当前分页窗口限制，目标 owner 使用精确 `runId` 读取详情；v4 / v5 继续由既有 strict parser 校验 application / workspace / authority。owner 离线、资源缺失或读取失败时明确说明没有请求或重放。
- 一次性 handoff 在目标 owner 首次处理时立即锁定，开发模式的 React effect 重放不会造成重复 owner 读取；跨 route 的旧 handoff 完成回调也不能消费后续新引用。
- 新增精确 draft 选择与分页外 v4 run 读取回归，Application Workspace 跨阶段旧 surface 拒绝断言已补齐；全部 Web 测试 `181 / 181`、`npm run build`、`git diff --check` 与仓库快速门禁通过。

2026-07-20 已完成批次 C 的真实浏览器与隐私边界收口：

- 在 `http://127.0.0.1:4100` 依次进入 Configure / Build、Human Promotion、Controlled Test、Run / Evaluation Review 与 Release Readiness；前四阶段始终只挂载一个 owner surface，Release Readiness 只展示只读汇总，不挂载执行 owner。
- 从 `app_flow_copilot` 切换到 `app_docs_assistant` 后，Application context、route generation 和 lifecycle evidence 均按新作用域重建；URL 只包含稳定阶段 hash，没有 application payload、input、answer、token 或完整资源引用。
- 浏览器启动复验发现离线 Application 没有权威 `recordVersion` 时 lifecycle evidence 会携带非法版本 `0` 并中断 React 根节点。实现已按来源真实性修正：保留可浏览 Application 作用域，但把 lifecycle evidence 降为 `incomplete / partial`，保留无版本短引用并明确缺少权威 revision，不再伪造完整证据或抛错。
- 完整页面加载与五阶段连续交互共观察到的请求仅为 Vite 页面、模块与动态 chunk；交互期没有 XHR、Fetch 或 `/v1/` owner 请求，控制台无页面 warning / error，网络无加载失败。浏览器存储不直接读取；代码扫描确认工作区只使用稳定 `location.hash`，不写 `localStorage`、`sessionStorage` 或 cookie，既有离线 consumer 测试继续固定零请求和一次性内存边界。
- Application Workspace 回归增至 `18` 项，全部 Web 测试 `182 / 182`、`npm run build` 与全量 `./scripts/check-repo.sh` 通过。批次 C 和本专题至此关闭；没有新增 API、schema、repository、fixture、checker、任务卡、发布记录或执行算法。

2026-08-08 已完成 Family UI `S2 R6` 产品化实现：

- `ApplicationDevelopmentWorkspacePanel` 复用现有 `evidenceState`、readiness view model、generation、surfaceKey 与 handoff controller，重组为 Application Context、五阶段 review path、Human Promotion 三项代表 contribution、十三段当前窗口、九组来源 readiness、authorization path 和渐进 owner surface；没有复制领域对象或增加 owner 请求。
- 桌面只让 Applications 产品导航和当前阶段形成柔底 / 墨蓝细轨选中态，普通 `missing`、`blocked`、`partial` contribution 保持中性行；窄屏按“Application Context → 当前阶段选择器 → 当前 surface → Readiness 摘要”排列，五阶段与全部九组来源可渐进展开。
- 缺少权威 revision 继续来自 Application lifecycle contribution 的 `partial`。底层 readiness 对未开始的 RAG assignment 保持 `incomplete`；S2 presentation 仅在现有 `rag_assignment` 缺失证据存在时把 RAG authority 显式表达为 `blocked` 风险，并通过 `data-rag-owner-status` 保留 owner rollup 原值，不改四态 readiness、owner contribution 或发布资格。
- 九组来源、十三项 contribution 和所有状态均由当前 view model 计算。Pencil 的 `5 / 9` 只是代表面数值；当前离线 fixture 实际只有 Application 引用，因此 React 如实显示 `1 / 9`，没有伪造来源覆盖。
- Web `272/272` 测试和 production build 通过；应用内浏览器复核 `1440x900`、`1101/1100px`、`821px`、`761/760px` 和 `390x844` 均无横向溢出，五阶段切换、九来源展开、owner review 开合和窄屏顺序正确，URL 无敏感材料，控制台零 warning / error。开发服务已在验收后关闭。
- 本轮没有新增 API、schema、repository、生产声明、task card、fixture、专项 checker、发布记录或执行算法；readiness 继续易失、只读、不可发布且不能满足 production authorization。

## 验收方式

设计评审至少确认：

- 页面阶段、资源 owner、可执行动作、只读 evidence 和下一跳一一对应；
- 唯一 application context 与 generation / 清理规则没有跨 Application 泄漏；
- 发布准备投影不会创建第二套 truth source 或生产声明；
- 首批实现不需要新增 API、schema、repository 或 checker；
- 生产认证、正式发布和自动化停止线保持关闭。

实现验证按风险分层：

- 精准 Web 测试：application / route switch、late response、cancel、offline 零请求、partial failure、归档只读、v4 / v5 handoff、一次性内容清理和零重复执行；
- 既有 consumer 测试、全部 Web 测试与 `npm run build`；
- `git diff --check` 与 `./scripts/check-repo.sh --fast`；
- 批次 C 或阶段真相源收口时补完整 `./scripts/check-repo.sh` 和真实浏览器检查；
- 只有引入后端 contract 或执行边界时才增加相称的 Go、race、vet、PostgreSQL integration 与专项验证。

## 2026-08-08 Family UI S5 组合边界

Application Runtime Review 作为 Application Workspace 的持续后置 surface 挂载，不改变 S2 五阶段、九组 source groups、十三项 contributions、revision `partial`、RAG authority `blocked` 或 readiness 只读不可发布的真相。旧 Controlled Test / Evidence 阶段只负责既有阶段 owner；Playground、Request History 与 Application Operations 由 S5 单一任务轨编排，避免同一 Operations owner 重复挂载。

S5 的 Run request、Request review、Review evidence 仅提供上下文连续性和精确 handoff。它不把调用结果写入 readiness，不把脱敏 request detail 当作可重放载荷，也不把当前窗口 operations 当作发布资格、全历史或生产健康证据。

## 2026-08-08 Family UI S6 组合边界

Workflow Run & Evaluation Review 作为 Application Workspace 的持续后置 surface 挂载，不改变 S2 五阶段、九组 source groups、十三项 contributions、revision `partial`、RAG authority `blocked` 或 readiness 只读不可发布的真相。旧 Evidence 阶段继续保留 RAG Evaluation Dataset 等阶段 owner，但不再重复挂载 Run History；Run、Comparison、Cases 与 Release Review 由 S6 四个精确 hash 编排为单一当前 owner。

S6 继续消费唯一 application / workspace / lifecycle context。workspace 配置与当前 context 不一致时 live owner 零请求失败关闭；application / workspace 切换先清空列表、详情、comparison 和 handoff，再用 generation guard 拒绝迟到响应。archived application 保留 metadata-only Run、Case、Suite 和 decision history 只读可达，诊断、创建、修订与 decision 写入关闭。

Run 列表只是严格 cursor 的当前窗口，不冒充全历史；Comparison 是请求时派生结果，不持久化也不重新执行；Case / Suite 只保存 exact version 引用和脱敏审查元数据；人工 `approved` 必须绑定当前 review digest，且只形成 append-only evidence，不创建 candidate、assignment、release、deploy、readiness contribution 或 production authorization。

## Workflow 创建、版本迭代与回归审查方案（2026-10-21）

### 目标与走查范围

目标用户是独立维护内部 Workflow 应用的开发者。完整任务为：创建应用和草案，保存并审查第一个不可变版本，显式激活及运行；派生第二版草案，重新审查和替换活动版本，取得第二次运行；比较两个精确 Run，建立版本化用例与套件，记录人工决定并在刷新后恢复证据。

本轮使用合成端口配置诊断输入、隔离 SQLite 和本地确定性提供方。主路径通过实际网页操作完成；结构化草案支线由公开开发测试态 API 准备 v2 资源，再检查网页读取、运行和评测编辑。API 准备资源不计作网页创建成功；固定回答不证明诊断质量。第二版主路径修改节点标签，验证版本来源与元数据差异，不宣称验证了提示词效果改善。

走查覆盖创建 / 保存、Definition 候选 / 审查 / 激活、派生与第二版运行、Run Comparison、Case / Suite / 人工决定、刷新恢复，以及版本族迟到响应、激活 CAS 冲突和结构化样本无效编辑。Session、RAG、HTTP Tool、模板市场与定时执行不是本轮新增走查范围；沿用各自既有证据，不合称全站或所有 Workflow 分支验收。

### 已确认可复用的路径

| 用户步骤 | 本轮观察 | 后续处理 |
| --- | --- | --- |
| 新建应用 → v1 草案编辑保存 | 实际网页可完成，后端生成应用标识，草案保留精确内容版本 | 保持现有入口和 v1 兼容 |
| 候选批准 → 独立激活 → 第一次运行 | 三个动作独立；Run v5 返回不可变 Definition 第 1 版来源 | 复用服务端权限与执行依据重验 |
| 派生草案 → 修改保存 → 第二版审查与运行 | 保留同一 Definition 版本族，需显式选择第 2 版并执行 replace；Run 指向第 2 版 | 不把批准自动等同于选择、激活或运行 |
| 两次 Run → Comparison → Case → Suite → 人工决定 | 同版本族跨版本比较返回 `comparable / changed`；用例与套件保存精确引用，人工 `needs_review` 决定可刷新恢复 | 复用现有审查模块，不创建第二套评测状态 |
| API 创建 v2 草案 → Definition → 结构化运行 | 后端资源可用，网页可消费不可变合同并产生 Run v8 | 将已有运行能力与尚缺的网页编写能力分开说明 |

### 实际问题与影响

| 编号 / 优先级 | 复现与实际影响 | 根因与代码入口 |
| --- | --- | --- |
| W1 / P2 | 网页不能创建或编辑 v2 输入合同；包含 v2 摘要的草案列表被整体拒绝，影响同页 v1 资源展示。v2 Definition 没有派生按钮，却提示使用尚不存在的 v2 合同设计入口 | `WorkflowDraftDesignerDraft` 不含版本化输入合同；`savedWorkflowDraftConsumer.ts` 的保存、摘要及详情校验固定 v1；`deriveWorkflowDraftFromDefinitionVersion` 显式拒绝 v2。不是后端缺少 v2 存储或执行能力 |
| W2 / P1 | 同一应用有 A / B 两个版本族，先选择 A 并延迟其读取，再选择 B；A 迟到后审查区仍选 B，运行区却显示 A，运行按钮可用。用户选择与执行依据失去一致性 | `workflowDefinitionPromotionPanel.tsx` 的读取代次只随应用 / 配置变化；版本族切换没有独立失效保护，`refresh` 与操作收尾也未统一绑定请求对象。走查没有执行这次不一致状态下的运行，不声称发生实际错误调用 |
| W3 / P2 | 另一操作已把激活指针推进到 v2，本页面以 v1 提交得到冲突；提示要求刷新，面板却无刷新入口，原 v1 操作仍可重复提交 | 同一面板保留旧 activation，冲突响应的当前版本只用于文案，没有明确的精确重读及重新决定流程；切换到另一版本族再切回或重载页面才能间接恢复 |
| W4 / P1 | 评测样本原整数为 `2`，输入无效文本后字段显示错误，“审查创建 / 确认创建”仍可继续；真实保存请求和持久计划继续使用旧值 `2` | `StructuredEvaluationFixtureEditor` 内部暂存无效字段，外层 `reviewPlan` 只读取最后一次有效的 `itemsJSON`；两份状态导致可见编辑、确认与提交不一致 |

以上是开发测试态可复现问题，不代表生产事故。代码入口位于 `apps/radishmind-web/src/features/control-plane-read/`。逐窗口证据、临时用例调整、验证结果和清理记录见 [W43 周志](../../devlogs/2026-W43.md#2026-10-21-workflow-完整走查与专题方案)。

### 评测路径的契约边界

Definition 计划目标冻结指针版本、Definition 版本和 digest；v2 还冻结输入合同。激活第二版后，旧计划不能自动改绑到新执行依据。更新目标必须显式保存新的计划版本。

Campaign 配对同时要求相同 `plan_id + plan_version + plan_digest`。因此当前批次配对适用于同一精确计划的重复运行，不能直接比较绑定不同 Definition 版本的两个计划修订。该限制来自 `application_evaluation_handoff.go` 和 `application_evaluation_campaign_runtime.go`，属于现有契约，不作为前端放宽校验的理由。

跨 Definition 版本回归继续使用精确 Run → Comparison → Case / Suite：v1 要求同版本族；v2 还要求同输入合同身份与 digest、同执行档案、终态记录。合同改变、v5 / v8 混用或版本族不同应说明不兼容，用户可显式建立新的基线，不能伪造可比较结论。Comparison 的 `changed` 只表示元数据变化，不等同于答案质量提升或下降。

本轮不设计跨计划版本批次配对、不自动重放旧输入、不从 Run 反推变量。若后续需要批量比较不同计划修订，须另行明确样本身份、合同兼容、输入一致性和证据归档规则。

### 实施范围与顺序

本方案作为一段连续产品开发任务推进，按依赖组织提交；每段执行必要的定向验证，最终完成两版应用的整段验收。范围确认后无需把普通实现选择拆成逐按钮审批。

1. **晋级选择与冲突恢复（W2、W3）。** 读取和操作状态绑定账户 / 工作区 / 应用 / 版本族及请求代次。切换时立即使旧版本、激活依据和易失运行输入失效；只有当前版本族的版本与激活信息均读取成功且彼此匹配，才开放相关动作。读取、写入后的刷新、异常和最终收尾全部使用同一失效规则。冲突保留用户理由，提供精确重读动作；重读成功展示当前指针、可选版本和允许动作，由用户重新决定，不自动重发。
2. **v2 草案完整编写链（W1）。** 在既有设计器增加“新建结构化输入草案”，编辑合同标识、说明及有序字段的名称、类型、必填性、标签和说明。最多 16 个字段，仅使用现有 `string / integer / number / boolean`；增删与排序复用表单交互。前端草案使用明确的 v1 / v2 联合类型，校验、保存、精确读取、列表、修订比较、恢复与派生共同保留版本和完整合同。合同 digest 以服务端为准。v1 保持原样；v2 从精确版本派生为新的可编辑草案，旧 Definition 不可修改，不通过默认值把 v2 降成 v1。
3. **评测编辑与执行依据衔接（W4）。** 在既有计划区域显式选择并读取当前活动 Definition，展示版本、指针、digest 和合同后采用为计划目标，避免必须手工拼接权威引用。重读发现变化时要求显式采用，并保留待检查的本地样本。样本字段与高级 JSON 使用一个编辑状态来源；无效原文必须保留并阻止确认，不能提交最后一次有效值。切样本、改目标和离开编辑区时明确处理未保存修改；不得借此接受服务端不支持的无效 fixture。
4. **连续交接与整体验收。** 从第二版 Run 进入现有比较和用例 / 套件审查；保留精确引用与版本，目标模块自行重读。计划界面说明“同计划重复运行”与“跨 Definition 版本 Run 审查”的适用路径。保持批准、激活、执行和评测决定相互独立，复核双语、布局、刷新恢复与异常路径，并同步使用指南。

第一段不等待新表单设计即可独立修复；第二、三段的新增编辑区域在实现前完成下述局部视觉设计。涉及的代码按设计器、保存消费端、晋级面板和评测编辑职责调整，不把合同、输入或领域记录上移到 `App.tsx` 或通用工作区上下文。仅为共享真实字段校验或版本转换引入抽取，不建立新的通用工作流管理层。

### 数据、兼容与交互约束

- 不新增 API、schema、数据库迁移、后端执行算法、依赖或生产能力；补齐前端对既有 v2 契约的读写。若现有 API 无法表达必要动作，先说明差异并重新确认范围，不在前端伪造支持。
- 本地未保存草案属于设计器；保存草案、Definition、activation、Run、Plan、Case、Suite 继续由各自服务负责。中英文切换保留用户原文和编辑状态，不触发业务请求。
- 列表和精确详情按真实 schema 分派；未知版本、未知字段、损坏合同继续拒绝，不能用忽略字段或降级读取解决 W1。
- v2 输入合同编辑保留字段顺序、类型和显式必填性。输入值仍只在运行组件内存停留；Plan 中只有用户显式提交的测试样本可以持久化，不导入历史运行原文。
- 版本族切换清除旧运行输入与输出；本地字段错误保留当前编辑。读取失败不得把旧 activation 当作最新可执行依据；传输未知不能解释为执行未发生，也不得自动重试。
- 激活候选版本的选择保持显式。批准第二版后仍显示原活动版本本身不是缺陷；界面应同时清楚标识“当前活动版本”和“待操作版本”。

### UI 设计覆盖

复用 S2 工作区、既有 Workflow 设计器、结构化运行输入 Visual R4 和 S10 计划编辑区域。W2 / W3 使用原动作和状态区域，预计为 `C / 直接实现`，五维评分 `0 / 0 / 0 / 1 / 1 = 2`。

合同编写与评测参考选择各按 `1 / 1 / 1 / 1 / 1 = 5` 采用 `B / 局部 Pencil`：只补有序字段编写、v1 / v2 来源说明、参考变更、无效值阻止提交和窄屏顺序。它们没有新产品壳、自治动作或生产确认；不得为凑分降低实际覆盖，后续出现新的复杂交接时重新判定。

本轮仅确认视觉设计需求，没有修改 Pencil 文件或取得新画板批准。实施第二、三段前完成局部设计和复核；原 Visual R4 只覆盖运行值表单，不能冒充合同编写区域的设计。使用语义 token，沿用单一主对象与连续工作区域，不增加独立页面或等权卡片看板。

### 验收标准与交付边界

| 验收路径 | 必须证明的结果 |
| --- | --- |
| v1 兼容 | 原网页创建、派生、审查、激活、两版运行、比较、Case / Suite 与刷新恢复继续可用 |
| v2 创建与迭代 | 全程网页完成合同编写、保存、精确重读、候选、运行、派生第二版；不使用 API 预置来代替网页创建，v1 / v2 混合列表正常读取 |
| 合同与版本 | 同合同两版 Run 可按既有规则比较；不同合同、混合 schema、未知字段和损坏数据失败关闭；旧版本与修订恢复不丢字段 |
| 并发与冲突 | A → B → A 迟到、读写刷新交错、账户 / 应用 / 工作区变化不污染当前选择；真实 CAS 冲突后保留理由、精确重读、再次显式决定 |
| 评测编辑一致性 | 无效整数、必填缺失、布尔未提供、JSON 错误不丢原文、不提交旧值；修改后确认摘要与实际保存一致；参考变化不静默改绑 |
| 回归审查与隐私 | 精确 Run / Case 版本和 Suite 决定可刷新恢复；界面不把比较分类解释为答案质量；变量、输出与凭据不进入不允许的存储、URL、日志 |
| 产品与工程验证 | 中英两语、桌面 / 关键断点 / 390px、真实截图及键盘操作；相关单测、覆盖率预算、构建和全量仓库检查；隔离环境关闭与产物审计 |

本轮交付为走查记录和专题方案，不含上述产品修复，不改变现有专题的历史通过证据。后续进入实现时以这四段范围为准；真实模型、团队独立使用、答案质量、生产交付与远端 CI 仍须分别取得证据。

## 停止线

- 不创建正式应用晋级、发布或 release API，不修改正式应用真相源。
- 不自动 approve、activation、assignment、profile selection、provider retry / fallback、schedule、background execution、replay / resume 或 writeback。
- 不实现长期记忆、durable transcript、自动摘要、session 内容搜索或 agent loop。
- 不进入 production OIDC、workspace membership、production secret、production API key、quota enforcement、billing、成本账本或公开生产 API。
- 不新增 Gateway 协议、provider 选择算法、Workflow 执行算法、RAG ranking、运行副本或跨 store 聚合记录。
- 不为普通页面组合新增 fixture、checker、readiness 链或平行 task card。
- 不把 `dev_test_evidence_reviewable`、已批准候选、成功 run 或 evaluation 结果解释为生产发布资格。

## 设计评审入口

功能设计、页面 / 状态所有权审计、批次 A 壳层和批次 B evidence handoff / readiness projection 均已确认。批次 C 只负责连续链、真实浏览器与隐私边界收口；若发现需要改变 API、schema、执行边界、生产声明或高风险写入，必须停止批次并先回到本专题修订设计与实施边界。
