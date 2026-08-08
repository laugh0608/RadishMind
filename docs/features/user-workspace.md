# 用户工作区设计与开发文档

更新时间：2026-08-08

## 功能定位

用户工作区是 RadishMind 面向终端用户和项目成员的主工作区，用于查看和管理 AI 应用、提示词应用、工作流、Agent / Copilot 应用、API 密钥、调用量、运行记录、成本摘要和人工审查入口。

## 当前状态

- `apps/radishmind-web/` 已有只读产品壳、工作区首页、应用列表、API 密钥、用量配额、工作流定义、运行历史、工作流审查区和审查交接区。
- [应用 API 接入与调用 v1](user-workspace/application-api-integration-invocation-v1.md) 已把当前选中应用、`/v1/models` 模型目录、三协议 × 三语言接入示例、现有 Gateway 调试台和脱敏请求历史串成连续的内部开发者路径；作用域不再依赖固定应用配置。
- [应用配置草案与审查 v1](user-workspace/application-configuration-draft-review-v1.md) 已为当前应用建立独立脱敏配置草案，完成模型 / 协议校验、memory / SQLite / PostgreSQL 开发测试态保存、恢复、配置比较、CAS 冲突审查，以及到 API 接入区和调试台的交接；正式应用真相源仍只读。
- [应用发布治理与晋级审查 v1](user-workspace/application-publish-governance-promotion-v1.md) 已把有效的已保存草案固定为不可变候选版本，完成服务端重读、摘要计算、审查 CAS、漂移 / 被取代检查、阻塞式晋级资格判断，以及到接入区、调试台和请求历史的交接；`approved` 仍不修改正式应用。
- [应用目录与生命周期（开发/测试态）v1](user-workspace/application-catalog-lifecycle-dev-test-v1.md) 已完成核心生命周期、memory / SQLite / PostgreSQL 开发测试持久化、Web 管理和真实浏览器纵向验收；[应用解除归档与安全重新启用 v1](user-workspace/application-unarchive-safe-reactivation-dev-test-v1.md)进一步完成 `archived -> active` 单赢家 CAS、组合权限、显式影响确认和下游重新资格判断。目录 owner 不级联改写 API Key、运行时绑定、会话、草案、候选或运行记录。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](user-workspace/api-key-lifecycle-gateway-dev-test-auth-v1.md) 已完成并关闭：活跃应用可以签发有期限、有受控作用域、只展示一次且可吊销的开发测试态密钥，五条 northbound 路由可显式启用 API 密钥认证，并记录可信调用上下文、脱敏请求历史与最近使用时间；[API 密钥引导式轮换与验证后退役 v1](user-workspace/api-key-guided-rotation-verified-retirement-dev-test-v1.md)复用这些 owner，完成同 scopes 替代、`last_used_at` 认证门槛、来源精确重读和 revoke CAS，不新增 rotate API、schema 或持久 rotation owner。
- [应用运行观测与用量归因 v1](user-workspace/application-operations-observability-usage-attribution-v1.md) 批次 A 已完成：当前应用可并列审查 Gateway 请求与 Workflow 运行的首分页窗口，分别查看状态、usage availability、受控调用计数、来源覆盖和合并时间线；两类记录不自动关联，当前窗口不冒充全量 usage、成本、配额或计费。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](user-workspace/application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md) 已完成并关闭：同一应用可显式选择 Workflow Definition v5 或 Application RAG v4 profile 建立 metadata-only session / turn，完成双数据库持久化、易失 transcript、取消、关闭、重启恢复、Run History 交接、真实浏览器和敏感信息扫描；不会从持久 metadata 恢复正文。
- [应用开发工作区与发布准备审查 v1](user-workspace/application-development-workspace-release-readiness-review-v1.md) 已完成并关闭：唯一 application context、workspace / route generation、五阶段单 surface、route-scoped evidence、精确 Draft / Run handoff、十三项 owner contribution、九个来源组和四态 readiness 投影均已进入 Web；真实浏览器已验证 Application 切换、稳定 hash、离线零 owner 请求和零页面控制台告警，缺少权威 revision 时保守显示 `incomplete / partial`。
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](user-workspace/prompt-application-template-version-review-controlled-invocation-dev-test-v1.md) 已完成并关闭：受限模板、不可变版本、Configuration Draft v3、Publish Candidate v3、Runtime Assignment、API key / Session v2、Run v6 与下游审查链均已接入；SQLite / PostgreSQL 连续链、重启恢复、CAS / authority drift / cancel 负向验收和浏览器隐私复验均已通过。
- [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](user-workspace/agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md) 已完成批次 A 至批次 E：strict contracts、policy compiler、三种 Profile owner、Configuration / Candidate v4、assignment、`agent_copilot:invoke`、Session / Turn v3、Run v7、类型专属 Web 与双数据库真实验收均已落地，专题关闭。
- [Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](user-workspace/prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md) 已完成：Prompt v6 / Comparison v5 与 Agent v7 / Comparison v6 已严格接入既有 Evaluation Case、Suite 和人工 decision；SQLite 真实浏览器完成 Agent case → suite → `approved v1`，没有新增评测 owner 或自动发布能力。
- [工作区运营收件箱（开发 / 测试态）v1](user-workspace/workspace-operations-inbox-dev-test-v1.md) 批次 A 已完成：active workspace 下的 Applications、API Keys、Workflow Definitions 与 Runs 首分页脱敏快照可投影为确定性关注队列，显式标记 partial / unavailable coverage，并跳转既有审查 surface；不新增 incident、notification、remediation 或 quota 真相源。
- [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](user-workspace/workspace-scoped-mutation-authorization-dev-test-v1.md) 已完成设计、全量 mutation inventory、唯一任务卡及批次 A 至 E；verified identity、active workspace、单项 / 原子组合与资源条件 membership permission、resource owner、稳定拒绝与副作用顺序已进入全部 47 条人类交互式 mutation，专题关闭。
- 工作区首页和工作流定义已支持创建本地工作流草案并进入草案设计器；草案保存复用仅开发的已保存草案消费端，不代表生产持久化已成立。
- `User Workspace Saved Draft Library v1` 已在工作区首页支持仅开发的活动 / 归档草案库：显示当前应用下已保存草案的脱敏摘要、组合筛选、严格 cursor 分页、空结果 / 失败状态、打开、归档只读审查和显式解除归档。默认内存、聚合 SQLite 与显式 PostgreSQL 开发测试态存储库均可承载该路径，但不代表生产持久化已成立。
- 草案设计器已支持本地节点新增、移动、删除保护、属性编辑和边重建；校验检查器、执行计划预览和运行时准入检查器使用当前活跃草案，不代表工作流可正式发布或执行。
- Family UI `S3 R1` 已完成既有草案、Saved Draft、Node Designer 与审查 owner 的 React 行为整合，但因确定性画布裁切、节点计数与实际可见节点错位、边不可读、底部 review 等权和窄屏重复 Inspector，被人工评审退回。`S3 R2` 已完成默认当前节点 / 直接邻居焦点、显式 `Fit graph`、紧凑 review、Inspector `1440px` 右侧 / `<=1380px` 下移 / `<=760px` 折叠与强选中语义修正，并通过桌面、临界断点和 `390x844` 严格复验；默认两节点邻域与全图八节点 / 七边均完整可读。完整 Saved Draft Library 仍是唯一列表 owner，归档草案只读。App owner 与九组来源、十三项 contribution、revision `partial`、RAG authority `blocked`、readiness 不可发布边界不变，不新增 API、schema、task card 或专项 checker。
- Family UI `S4 R1` 已把 Application API Integration、API Key lifecycle、一次性交接、Playground handoff 和验证后退役组织为同一 Application Access 任务面。模型协议只开放目录真实能力，Key 创建面使用七项真实 scope；workspace scope 漂移和 `api_key_dev_test` 无凭据均失败关闭。archived application 保留 Key metadata / detail / revoke，只阻断 issue、rotate、integration 与 invocation；offline workspace summary 与当前 application lifecycle 列表分离。Pencil Desktop / Narrow R1 与 Decision R9、Web `277/277`、production build、关键断点和 `390x844` 浏览器严格复验均已通过；没有新增 API、schema、repository、task card 或专项 checker。
- 工作流审查交接已把当前打开的活动草案校验、执行计划和运行时准入结果汇总为可交接审查记录，仍不保存、不导出、不发送交接内容。
- 已保存草案与运行历史具备 memory / SQLite / PostgreSQL 开发测试态存储库；受控执行器 v0、失败审查、运行比较、评测用例 / 版本管理和评测套件 / 发布审查已接入工作区运行历史。
- 应用配置草案、发布候选和显式启用的应用目录均具备 memory、SQLite 与 PostgreSQL 开发测试态存储库；应用目录未启用时，历史只读列表仍来自预置假数据存储库。
- 当前仍不具备生产认证 / 存储库、Radish 工作区成员关系、正式应用生命周期 / 晋级、生产 API 密钥、配额执行、计费、不受限工具调用、业务写回或重放；开发测试态 HTTP Tool 人工确认与两种受控应用调用已经存在，但不能外推为通用工具、自动确认或生产执行能力。
- 2026-07-30 已使用同一 SQLite 本地产品档完成三条真实开发者连续链复盘：Saved Draft 活动 → 归档 → 只读审查 → 解除归档 → 重新打开；应用活动 → 归档 → Gateway 资格拒绝 → 安全重新启用 → 资格恢复；API Key 来源选择 → 同 scopes 替代 → Gateway 认证 → `last_used_at` 验证 → 来源退役。领域 owner、双版本 CAS、失败关闭和一次性凭据边界均符合设计，确认的摩擦全部属于现有 Web 上下文交接、状态解释与信息密度。

历史已保存草案准入专题继续作为证据索引保留，不再在本入口重复展开。需要追溯时，从 [工作流专题入口](workflow/README.md) 和对应实现专题进入。

## 2026-07-30 真实路径复盘与 UI 一致性批次

本批状态为 `user_workspace_real_path_ui_coherence_v1_completed`。它是既有功能的 UI 阶段首批完整纵向切片，不改变领域协议、持久化 owner、权限、生命周期或生产边界。

| 真实目标 | 已确认摩擦 | 既有 owner | 本批处理 |
| --- | --- | --- | --- |
| 从草案库打开活动草案或审查归档草案 | 精确读取成功后页面仍停留在长列表；长名称会主导卡片高度；解除归档后的失败关闭状态直接显示 `unknown` | Saved Draft consumer 与 Workflow Draft Designer | 仅在精确打开成功后交接到 Designer；限制卡片标题的视觉行数并保留完整名称；把 `unknown` 解释为“需要重新打开” |
| 归档应用后理解 Gateway 资格为何失效 | SQLite 本地产品仍显示 `PostgreSQL dev/test`；模型目录把 canonical `api_key_application_unavailable` 降为通用 `HTTP 403` | Application Catalog 与 Gateway 模型目录 consumer | 使用不冒充具体 store 的开发测试态标签；严格消费允许列表内的 Gateway 错误信封，显示稳定错误码和固定脱敏摘要 |
| 验证替代 Key 后审查轮换状态 | 精确详情已得到 `last_used_at`，轮换面板已解锁，但同页列表仍显示未使用，直到再次加载 | API Key Lifecycle consumer 与易失轮换会话 | 用精确详情替换当前列表中的同一记录，使验证面板、详情和列表立即一致 |

验收要求：

1. Saved Draft 的“打开草案 / 只读审查”只有在精确读取和作用域检查成功后才进入 `#workflow-draft-designer`；失败时保留在草案库并展示原失败状态。
2. 长草案名称不再无限拉高卡片，完整名称仍可通过原生标题查看；解除归档后的浏览器旧快照保持只读，但向用户解释为需要重新打开，而不是暴露 `unknown` 技术状态。
3. Application Catalog 不声明无法从当前配置证明的存储实现；已归档或不可用应用的模型目录请求显示 `api_key_application_unavailable` 和固定脱敏解释，未知或非法失败正文继续收敛为通用失败，且不泄露响应正文。
4. 替代 Key 的精确验证读取成功后，同页列表立即反映最新 `last_used_at`；一次性 token 仍只保存在组件内存，不进入列表、详情、日志或持久化介质。
5. 复用现有 Web 单元测试、production build、真实浏览器链与仓库门禁；不新增 API、schema、migration、repository、任务卡、fixture 或专项 checker。

停止线：本批不修复或清理历史开发测试数据，不增加自动命名、批量生命周期、自动轮换、持久 rotation owner、跨分页运营收件箱、全历史用量聚合、生产认证或生产凭据能力。

## 设计边界

- 用户端默认只输出建议、解释、审查包和候选动作，不直接写业务真相源。
- 高风险动作必须保留 `requires_confirmation`。
- 只读侧与后续写入 / 执行侧必须分开设计；界面可展示不等于执行能力已就绪。
- API 密钥列表、详情和后续读取不得展示密钥值、摘要、`Authorization` 请求头或任何敏感材料；创建成功后由独立响应完成的一次性交接是唯一例外，且不得进入浏览器持久化介质。

## 下一批开发方向

1. `user_workspace_real_path_ui_coherence_v1` 已完成并关闭。[RadishMind Family UI 产品化设计与迁移 v1](user-workspace/radishmind-family-ui-productization-v1.md) 已对齐 family-ui `v26.7.3` 通用参考基线，并由 RadishMind 主动选择 Workbench Profile、原样 token 镜像和项目语义别名；差异附录、[27 张参考图映射](user-workspace/radishmind-family-ui-reference-mapping-v1.md)、Pencil 分级，以及 `S1 R8`、`S2 R6`、`S3 R2`、`S4 R1` 的设计、React 实现和严格浏览器验收均已完成。下一步联合复核 Gateway Playground、Request History 与 Application Operations，把既有受控调用、精确请求审查和应用当前窗口证据组织成下一条连续产品路径；先判断设计覆盖级别，再实施 React 纵向切片。Saved Draft Library 继续保持 `B` 级唯一列表 owner，当前窗口仍不冒充全历史或 quota / billing。
2. 后续批次继续要求跨 tenant / subject、非成员、过期 identity / membership、workspace mismatch、permission denied 在业务 repository 查询或副作用前失败关闭。dev header 与 signed-test assertion 只能用于开发测试，不能成为 production OIDC 授权来源。
3. 工作区运营收件箱批次 A 已完成。只有真实需要跨全部分页窗口，且四类 owner 的统一稳定 cursor 契约成立时才评审批次 B；不为扩展示例数量或页面计数启动服务端投影。
4. Prompt / Agent 继续复用 canonical Run、Comparison、Evaluation Case / Suite 与 decision owner；不复制评测算法，不把人工 `approved` 接成自动 candidate、assignment、release 或 deploy。Agent / Copilot 仍复用 canonical `CopilotRequest / CopilotResponse`，不扩 agent loop、工具执行或业务写回。
5. 本地 SQLite、应用目录与安全重新启用、API 密钥生命周期与引导式轮换、Application Interaction Session 专题均已完成并关闭；不继续扩同层页面、准入文档、检查器或证据链。应用运行观测只有在全分页统计或 quota / billing owner 成立时才评审服务端 summary；可信 reported usage 已成立，但当前仍只覆盖已加载窗口。
6. 一次性令牌继续只保存在当前 Web 组件内存；刷新、路由离开、应用 / 身份切换、组件卸载和服务重启都不得恢复原始令牌。
7. 不把开发测试态应用目录或 API 密钥解释为生产存储库与生产授权；OIDC 模式在成员关系契约未成立时继续失败关闭。后续专题不得隐式打开生产认证、成员关系适配器、正式晋级、生产 API 密钥、配额、计费、模型服务凭据或新的 Gateway 请求 / 响应 schema。

## 验收方式

- 功能展示类：`npm run build`、必要浏览器布局检查、`./scripts/check-repo.sh --fast`。
- 只读契约类：消费端冒烟验证、Go 处理器测试和只读侧契约检查。
- 写入或执行类：先补设计文档和任务卡，再补单元测试、负向测试、仓库级检查和人工确认路径。

<!--
历史检查器兼容字面量，仅供既有证据链读取，人工默认不读：
repository contract preconditions
saved draft list
durable store 迁移前置设计
owner / workspace
Workflow Draft Node Attribute Editing Model v1
Workflow Review Handoff Active Draft v1
Workspace Home / workflow definitions
创建本地 workflow 草案
dev-only saved draft consumer
User Workspace Saved Draft List v1
sanitized summary
-->
