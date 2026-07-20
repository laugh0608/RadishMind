# 应用开发工作区与发布准备审查 v1

更新时间：2026-07-20

状态：`application_development_workspace_release_readiness_review_v1_batch_a_completed_batch_b_ready`

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

### 批次 C：连续链与专题收口

- 覆盖 Application 切换、归档只读、route switch、late response、cancel、offline、partial source failure、v4 / v5 run handoff、Evaluation compatibility 和一次性内容清理。
- 完成真实浏览器连续路径、URL / storage / console / network 审计和现有功能不回归验证。
- 根据实际行为证据更新功能专题、当前焦点、能力矩阵和周志后关闭专题。

若任何批次发现必须新增 API、schema、repository、执行边界或高风险写入，应停止当前批次，先更新本设计并创建唯一的高风险任务卡；不得把新边界隐藏在普通 UI 重组中。

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

批次 A 已关闭，批次 B 可以开始。下一步建立通用 feature-scoped 脱敏 handoff refs、来源分组与四态 readiness view model；本批次的一次性 Application API handoff 只是维持现有跨阶段路径的窄边界，不提前等同于完整 Batch B handoff state。

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

## 停止线

- 不创建正式应用晋级、发布或 release API，不修改正式应用真相源。
- 不自动 approve、activation、assignment、profile selection、provider retry / fallback、schedule、background execution、replay / resume 或 writeback。
- 不实现长期记忆、durable transcript、自动摘要、session 内容搜索或 agent loop。
- 不进入 production OIDC、workspace membership、production secret、production API key、quota enforcement、billing、成本账本或公开生产 API。
- 不新增 Gateway 协议、provider 选择算法、Workflow 执行算法、RAG ranking、运行副本或跨 store 聚合记录。
- 不为普通页面组合新增 fixture、checker、readiness 链或平行 task card。
- 不把 `dev_test_evidence_reviewable`、已批准候选、成功 run 或 evaluation 结果解释为生产发布资格。

## 设计评审入口

功能设计和现有页面 / 状态所有权审计已确认，批次 A 获准进入产品代码实现。若实现发现需要改变 API、schema、执行边界、生产声明或高风险写入，必须停止批次并先回到本专题修订设计与实施边界。
