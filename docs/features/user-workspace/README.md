# 用户工作区细专题入口

更新时间：2026-09-23

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](application-evaluation-scheduled-regression-campaign-dev-test-v1.md)：状态为 `application_evaluation_scheduled_regression_campaign_dev_test_v1_completed`。受限委托 P0 与批次 A 至 D 已完成规范契约与三种存储实现、严格 HTTP 接口、逐次授权重验、显式开发测试态运行器、既有评测批次 / 运行记录交接、崩溃后不重放、完整 Pencil、单一 React 严格消费端、双数据库产品链、CAS、重启 / 重连、三视口与隐私审计；任务卡关闭，不派生批次 E。
- [应用结果资产库与受控导出（开发 / 测试态）v1](application-result-artifact-library-controlled-export-dev-test-v1.md)：状态为 `application_result_artifact_library_controlled_export_dev_test_v1_completed`；批次 A 至 C 的应用作用域严格列表、过滤游标、规范导出、独立导出权限、双数据库索引、严格 Web 消费端、三视口页面与双数据库产品连续链已完成，专题关闭且不复制结果资产 / 生命周期负责模块。
- [应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](application-session-result-artifact-explicit-retention-dev-test-v1.md)：批次 A 至 D 已完成并关闭；默认关闭的显式保存、三种存储中的不可变结果资产、版本化生命周期、共享严格 Web 消费端、SQLite 重启页面恢复与 PostgreSQL 配置化产品链均已成立，不改变运行历史 / 会话仅元数据契约，也不派生批次 E、通用结果存储或会话正文。
- [RadishMind Family UI 参考图产品面映射 v1](radishmind-family-ui-reference-mapping-v1.md)：已把 family-ui `references.md` 的 `ref-01` 至 `ref-27` 逐项映射到 S1–S8 八个产品面，固定实际查看、共享转译、禁止照搬内容、Pencil 构件与版权停止线。
- [RadishMind Family UI 产品化设计与迁移 v1](radishmind-family-ui-productization-v1.md)：family-ui `v26.7.3` 参考基线、RadishMind Workbench 选择和项目语义层已经对齐；`S1 R8` 至 `S8 R1` 已完成设计、React 与真实浏览器验收，S9 / S10 功能、Pencil Visual R3、React 迁移与真实浏览器复核也已完成。
- [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](application-evaluation-campaign-controlled-execution-dev-test-v1.md)：后端 A 至 D、S10 React 严格消费端、内存 / SQLite 三视口精确交接和重启恢复均已完成；结构化输入专题又在原负责模块上补齐定义 v2 类型化测试样例、比较 v7、三种存储的连续链与用例 / 套件精确交接。
- [Prompt / Agent / Copilot 类型工作区产品化 v1](prompt-agent-copilot-type-workspace-productization-v1.md)：S8 已完成。既有模板 / 档案、配置、候选版本、运行时绑定、接入、会话 / 调用、运行记录与评测负责模块被编排为七 / 八任务、单一负责模块的工作区；开发测试态停止线、`A` 级 Pencil 与真实浏览器证据均已关闭。
- [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](workspace-scoped-mutation-authorization-dev-test-v1.md)：批次 A 至 E 共 47 条写入操作已完成身份 / 成员资格双重权限、当前活动工作区唯一选择、原子组合与条件权限、稳定拒绝映射和零业务 / 外部副作用证据；专题关闭。
- [工作区运营收件箱（开发 / 测试态）v1](workspace-operations-inbox-dev-test-v1.md)：批次 A 已完成四类既有负责模块的首分页关注项、来源覆盖、稳定严重度、Web 既有详情跳转和工作区切换失败关闭；不新增运营或修复真相源。
- [Workspace-scoped Read Transition / 工作区选择与成员资格绑定（开发 / 测试态）v1](workspace-scoped-read-transition-dev-test-v1.md)：批次 A、B 已完成共享成员资格提供方、五类路由授权、四类持久化负责模块的读取投影、工作区范围运行记录游标与非持久化 Web 选择器；配额和生产成员资格继续关闭。
- [Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md)：Prompt v6 / Agent v7 已严格接入既有比较、评测用例、套件与人工决策；SQLite Agent 用例 → 套件 → `approved v1` 和隐私复验均已完成，专题关闭。
- [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)：批次 A 至批次 E 已完成；档案、发布与运行时绑定、唯一受控建议、会话 / 运行记录、类型专属 Web 与双数据库真实验收形成闭环，专题关闭。
- [Agent / Copilot 开发测试态使用指南](agent-copilot-dev-test-usage-guide.md)：说明档案 → 配置草案 v4 → 候选版本 v4 → 运行时绑定 → API 密钥 / 会话 v3 → 运行记录 v7 的操作顺序、启动配置、身份权限、CAS、持久化、隐私和排障边界。
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)：批次 A 至 E 均已完成，受限模板、双数据库模板负责模块、配置草案 v3、发布候选版本 v3、显式运行时绑定、受控调用、会话 / 交互轮次 v2、运行记录 v6、Web 与真实浏览器验收均有可复验证据，专题关闭。
- [提示词应用开发测试态使用指南](prompt-application-dev-test-usage-guide.md)：说明完整的模板 → 配置草案 → 候选版本审查 → 运行时绑定 → 调用 / 会话 → 运行记录审查顺序，以及启动配置、身份权限、CAS、持久化与故障处理；所有能力仅限开发测试态。
- [应用开发工作区与发布准备审查 v1](application-development-workspace-release-readiness-review-v1.md)：批次 A 至 C 已完成并关闭；路由作用域证据、草案 / 运行记录负责模块的精确重读、离线修订版本失败关闭、真实浏览器连续路径与 URL / 控制台 / 网络隐私审计均有可复验证据。
- [应用受控运行开发测试态指南](application-controlled-runtime-dev-test-guide.md)：说明 Application RAG、Workflow Definition v1 / v2、Application Interaction Session v1 / v4、v4 / v5 / v8 运行记录与 Application Operations 的启动、资源准备、作用域、恢复、失败语义和隐私边界。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)：严格契约、三种会话 / 交互轮次负责模块、权威来源精确重读、v5 / v4 单次委托、Web 易失交互工作区、双数据库启动器连续链、重启恢复、真实浏览器和敏感信息扫描均已完成，专题关闭。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：Gateway 认证、统一 `sqlite_dev` 存储库 / 聚合运行时、双数据库门禁、Web 一次性交接、真实浏览器连续路径、重启恢复与敏感信息复验均已完成，专题关闭。
- [API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](api-key-guided-rotation-verified-retirement-dev-test-v1.md)：批次 A、B 已完成；易失脱敏会话、同作用域替代、`last_used_at` 验证门槛、精确退役 CAS 与真实浏览器连续链均有可复验证据，专题关闭。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用解除归档与安全重新启用（开发 / 测试态）v1](application-unarchive-safe-reactivation-dev-test-v1.md)：批次 A 至 C、三种存储的 CAS、组合权限、显式影响确认、Gateway 资格回归、Web 与真实浏览器连续验收均已完成，专题关闭。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。
- [应用运行观测与用量归因 v1](application-operations-observability-usage-attribution-v1.md)：已完成应用作用域 Gateway Request History 与 Workflow Run History 的独立来源覆盖、当前窗口归因摘要和合并时间线；2026-08-19 后续准入评审为 `no_entry`，不启动跨页 summary，也不推测跨来源关联或估算 token、成本、配额、计费。

## 下一步

- 定时回归评测专题 P0 与批次 A 至 D 已完成并关闭，完整 Pencil 与 React 产品面继续复用 S10 表达精确评测计划、配额消费端、系统执行者与受委托用户、下次到期时间、生命周期、单次执行、评测批次交接、吊销和重启；没有建立 S11。真实提供方与生产任务执行器仍关闭，始终不得用创建者 `actor_ref` 冒充交互式请求。
- 应用运行观测后续准入已评审完成，真实页面也已贯通“受控调用或会话 → 应用运行观测 → 结果保存 → 结果工作区 → 精确运行详情 / 比较”。既有所有者的精确运行目标交接、缺失证据说明和权威来源漂移恢复引导已经完成；有效运行记录直接打开详情，缺失记录失败关闭，会话重新加载不自动切换、创建或重试。下一顺位回到上级功能设计入口，不为该普通 UI / 使用性修正启动服务端投影、新专题、Pencil 或专项门禁。
- 应用会话运行结果资产显式保存与恢复 v1 已完成批次 A 至 D 并关闭。下一产品顺位回到上级功能设计文档入口选择新的长期目标；不从已关闭的会话 / 结果资产专题扩展永久清除、会话正文、长期记忆、重放 / 继续运行或 Agent 循环。
- S9 / S10 功能实现、SQLite 重启复验、Visual R3 人工复核、React 迁移与三视口浏览器证据已完成；旧 R1 与 Visual R2 仍只保留为退回历史。Provider 价格与应用成本专题的 S7 / S5 Visual R1、React strict consumer 和产品连续链也已完成。下一顺位回到功能设计入口选择新的真实产品阻塞，不从已关闭专题派生同层页面、自动执行或生产能力。
- API 密钥引导式轮换与验证后退役已完成并关闭。下一轮先依据用户工作区与 Workflow 的真实使用证据更新对应功能设计；不从本专题扩自动轮换、持久 rotation owner 或生产凭据能力。
- 工作区运营收件箱批次 A 已完成；先以真实开发测试使用反馈判断是否需要跨全部分页窗口的服务端读取投影。没有需求与统一的所有者游标契约前不启动批次 B。
- 工作区范围读取迁移（Workspace-scoped Read Transition）开发 / 测试态批次 A、B 已完成并关闭。历史条件式批次 C 只指旧版 Radish 资源服务器成员资格适配器；本地成员资格所有者、Web 会话执行者、确定性浏览器 OIDC 与当前账户 Web 所有者已由联合身份专题批次 A 至 D 承接，S7 工作区成员 / 角色管理已由独立本地成员管理专题批次 A 至 E 承接，不从本专题恢复该适配器。
- 工作区范围写入授权（Workspace-scoped Mutation Authorization）批次 A 至 E 已完成并关闭；后续生产成员资格适配器和真实 OIDC 只在上游契约经过审查且齐备后独立恢复，不从本专题派生同层仅检查门禁批次。
- Prompt / Agent 回归评测与发布审查专题已完成并关闭；下一步先设计新的用户工作区产品能力，不继续派生本专题同层准入、更新或仅检查门禁批次。
- 提示词应用（Prompt Application）批次 A 至 E 已完成并关闭：内存 / SQLite / PostgreSQL 语义、Web、双数据库连续链、服务重启、CAS / 漂移 / 取消和敏感信息复验均已通过。
- 不继续扩“应用开发工作区与发布准备审查 v1”、Prompt Application 或当前回归评测的同层切片。Prompt / Agent 继续复用现有 Run、Comparison、Evaluation 和发布治理真相源，不另建聚合发布真相源或自治执行器。
- 不从已关闭的应用交互会话（Application Interaction Session）派生长期记忆、自动档案、重试 / 回退、定时调度、重放 / 继续运行或 Agent 循环。服务端汇总当前保持 `no_entry`；后续只有真实跨页任务、稳定的所有者 / 快照 / 游标契约、性能预算与正式配额 / 计费所有者同时成立时才重新评审。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
- 应用解除归档只恢复目录活动态；不自动创建、吊销、激活或改写 API Key、运行时绑定、会话、草案、候选、定义和运行记录。
