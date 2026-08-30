# 用户工作区细专题入口

更新时间：2026-08-30

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [应用定时回归评测与受控 Campaign 调度（开发 / 测试态）v1](application-evaluation-scheduled-regression-campaign-dev-test-v1.md)：状态为 `application_evaluation_scheduled_regression_campaign_dev_test_v1_batch_c_completed_batch_d_awaiting_design_approval`。受限委托 P0 与 Batch A 至 C 已完成 canonical / 三存储 owner、strict HTTP、逐次授权重验、显式 dev/test runner、既有 Campaign 交接与 crash no-replay；等待 Batch D 完整 Pencil 批准。
- [应用结果资产库与受控导出（开发 / 测试态）v1](application-result-artifact-library-controlled-export-dev-test-v1.md)：状态为 `application_result_artifact_library_controlled_export_dev_test_v1_completed`；批次 A 至 C 的 application-scoped 严格列表、过滤 cursor、canonical export、独立 export 权限、双数据库索引、strict Web consumer、三视口页面与双数据库产品连续链已完成，专题关闭且不复制 artifact / lifecycle owner。
- [应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1](application-session-result-artifact-explicit-retention-dev-test-v1.md)：批次 A 至 D 已完成并关闭；默认关闭的显式保存、三存储不可变 artifact、版本化 lifecycle、共享 strict Web consumer、SQLite 重启页面恢复与 PostgreSQL 配置化产品链均已成立，不改变 Run History / Session metadata-only 契约，也不派生批次 E、通用 result store 或 transcript。
- [RadishMind Family UI 参考图产品面映射 v1](radishmind-family-ui-reference-mapping-v1.md)：已把 family-ui `references.md` 的 `ref-01` 至 `ref-27` 逐项映射到 S1–S8 八个产品面，固定实际查看、共享转译、禁止照搬内容、Pencil 构件与版权停止线。
- [RadishMind Family UI 产品化设计与迁移 v1](radishmind-family-ui-productization-v1.md)：family-ui `v26.7.3` 参考基线、RadishMind Workbench 选择和项目语义层已经对齐；`S1 R8` 至 `S8 R1` 已完成设计、React 与真实浏览器验收，S9 / S10 功能、Pencil Visual R3、React 迁移与真实浏览器复核也已完成。
- [应用评测计划、受控执行与证据归档（开发 / 测试态）v1](application-evaluation-campaign-controlled-execution-dev-test-v1.md)：后端 A 至 D、S10 React strict consumer、memory / SQLite 三视口 exact handoff 和重启恢复均已完成；结构化输入专题又在原 owner 上补齐 Definition v2 typed fixture、Comparison v7、三存储连续链与 exact Case / Suite handoff。
- [Prompt / Agent / Copilot 类型工作区产品化 v1](prompt-agent-copilot-type-workspace-productization-v1.md)：S8 已完成。既有 Template / Profile、Configuration、Candidate、Assignment、Access、Session / Invocation、Run 与 Evaluation owner 被编排为七 / 八任务单 owner 工作区；开发测试态停止线、`A` 级 Pencil 与真实浏览器证据均已关闭。
- [Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](workspace-scoped-mutation-authorization-dev-test-v1.md)：批次 A 至 E 共 47 条 mutation 已完成 identity / membership 双重权限、active workspace 唯一选择、原子组合与条件权限、稳定拒绝映射和零业务 / 外部副作用证据；专题关闭。
- [工作区运营收件箱（开发 / 测试态）v1](workspace-operations-inbox-dev-test-v1.md)：批次 A 已完成四类既有 owner 首分页关注项、来源覆盖、稳定严重度、Web 既有详情跳转和 workspace 切换失败关闭；不新增运营或修复真相源。
- [Workspace-scoped Read Transition / 工作区选择与成员资格绑定（开发 / 测试态）v1](workspace-scoped-read-transition-dev-test-v1.md)：批次 A、B 已完成共享 membership provider、五类 route 授权、四类 durable owner 读投影、workspace-wide Run cursor 与非持久化 Web selector；quota 和生产 membership 继续关闭。
- [Prompt / Agent 应用回归评测与发布审查（开发 / 测试态）v1](prompt-agent-application-regression-evaluation-release-review-dev-test-v1.md)：Prompt v6 / Agent v7 已严格接入既有 Comparison、Evaluation Case、Suite 与人工 decision；SQLite Agent case → suite → `approved v1` 和隐私复验均已完成，专题关闭。
- [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)：批次 A 至批次 E 已完成；Profile、发布与 assignment、唯一受控建议、Session / Run、类型专属 Web 与双数据库真实验收形成闭环，专题关闭。
- [Agent / Copilot 开发测试态使用指南](agent-copilot-dev-test-usage-guide.md)：说明 Profile → Configuration v4 → Candidate v4 → Assignment → API key / Session v3 → Run v7 的操作顺序、启动配置、身份权限、CAS、持久化、隐私和排障边界。
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)：批次 A 至 E 均已完成，受限模板、双数据库 Template owner、Configuration Draft v3、Publish Candidate v3、显式 Runtime Assignment、受控 invocation、Session / Turn v2、Run v6、Web 与真实浏览器验收均有可复验证据，专题关闭。
- [Prompt Application 开发测试态使用指南](prompt-application-dev-test-usage-guide.md)：说明完整 Template → Configuration → Candidate Review → Assignment → Invocation / Session → Run Review 顺序，以及启动配置、身份权限、CAS、持久化与故障处理；所有能力仅限开发测试态。
- [应用开发工作区与发布准备审查 v1](application-development-workspace-release-readiness-review-v1.md)：批次 A 至 C 已完成并关闭；route-scoped evidence、精确 Draft / Run owner 重读、离线 revision 失败关闭、真实浏览器连续路径与 URL / console / network 隐私审计均有可复验证据。
- [应用受控运行开发测试态指南](application-controlled-runtime-dev-test-guide.md)：说明 Application RAG、Workflow Definition v1 / v2、Application Interaction Session v1 / v4、v4 / v5 / v8 运行记录与 Application Operations 的启动、资源准备、作用域、恢复、失败语义和隐私边界。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)：strict contract、三种 Session / Turn owner、exact authority reload、v5 / v4 单次委托、Web 易失交互工作区、双数据库 launcher 连续链、重启恢复、真实浏览器和敏感扫描均已完成，专题关闭。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：Gateway 认证、统一 `sqlite_dev` repository / 聚合 runtime、双数据库门禁、Web 一次性交接、真实浏览器连续路径、重启恢复与敏感信息复验均已完成，专题关闭。
- [API 密钥引导式轮换与验证后退役（开发 / 测试态）v1](api-key-guided-rotation-verified-retirement-dev-test-v1.md)：批次 A、B 已完成；易失脱敏会话、同 scopes 替代、`last_used_at` 验证门槛、精确退役 CAS 与真实浏览器连续链均有可复验证据，专题关闭。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用解除归档与安全重新启用（开发 / 测试态）v1](application-unarchive-safe-reactivation-dev-test-v1.md)：批次 A 至 C、三种 store CAS、组合权限、显式影响确认、Gateway 资格回归、Web 与真实浏览器连续验收均已完成，专题关闭。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。
- [应用运行观测与用量归因 v1](application-operations-observability-usage-attribution-v1.md)：已完成应用作用域 Gateway Request History 与 Workflow Run History 的独立来源覆盖、当前窗口归因摘要和合并时间线；2026-08-19 后续准入评审为 `no_entry`，不启动跨页 summary，也不推测跨来源关联或估算 token、成本、配额、计费。

## 下一步

- 定时回归评测专题下一步只批准 Batch D 完整 Pencil：复用 S10 表达 exact Plan / quota consumer / system actor + delegated user / next due / lifecycle / Occurrence / Campaign handoff / revoke / restart，不建立 S11。未经批准不修改设计源或 React，不启动产品验收或真实 Provider，始终不得用创建者 `actor_ref` 冒充交互式请求。
- 应用运行观测后续准入已评审完成，真实页面也已贯通“受控调用或 Session → Application Operations → 结果保存 → Result Workspace → exact Run detail / Comparison”。既有 owner 的 exact Run 目标交接、缺失 evidence 说明和 authority drift 恢复引导已经完成；有效 Run 直接打开详情，缺失 Run 失败关闭，Session reload 不自动切换、创建或重试。下一顺位回到上级功能设计入口，不为该普通 UI / 使用性修正启动服务端投影、新专题、Pencil 或专项门禁。
- 应用会话运行结果资产显式保存与恢复 v1 已完成批次 A 至 D 并关闭。下一产品顺位回到上级功能设计文档入口选择新的长期目标；不从已关闭 Session / Result Artifact 专题扩永久 purge、transcript、长期记忆、replay / resume 或 agent loop。
- S9 / S10 功能实现、SQLite 重启复验、Visual R3 人工复核、React 迁移与三视口浏览器证据已完成；旧 R1 与 Visual R2 仍只保留为退回历史。Provider 价格与应用成本专题的 S7 / S5 Visual R1、React strict consumer 和产品连续链也已完成。下一顺位回到功能设计入口选择新的真实产品阻塞，不从已关闭专题派生同层页面、自动执行或生产能力。
- API 密钥引导式轮换与验证后退役已完成并关闭。下一轮先依据用户工作区与 Workflow 的真实使用证据更新对应功能设计；不从本专题扩自动轮换、持久 rotation owner 或生产凭据能力。
- 工作区运营收件箱批次 A 已完成；先以真实开发测试使用反馈判断是否需要跨全部分页窗口的服务端 read projection。没有需求与统一 owner cursor 契约前不启动批次 B。
- Workspace-scoped Read Transition 开发 / 测试态批次 A、B 已完成并关闭。历史条件式批次 C 只指 legacy Radish resource-server membership adapter；本地 membership owner、Web Session actor、确定性 browser OIDC 与当前账户 Web owner 已由联合身份专题批次 A 至 D 承接，S7 workspace 成员 / 角色管理已由独立本地成员管理专题批次 A 至 E 承接，不从本专题恢复该 adapter。
- Workspace-scoped Mutation Authorization 批次 A 至 E 已完成并关闭；后续生产 membership adapter 和真实 OIDC 只在 reviewed 上游契约齐备后独立恢复，不从本专题派生同层 gate-only 批次。
- Prompt / Agent 回归评测与发布审查专题已完成并关闭；下一步先设计新的用户工作区产品能力，不继续派生本专题同层 readiness、refresh 或 gate-only 批次。
- Prompt Application 批次 A 至 E 已完成并关闭：memory / SQLite / PostgreSQL 语义、Web、双数据库连续链、服务重启、CAS / drift / cancel 和敏感信息复验均已通过。
- 不继续扩“应用开发工作区与发布准备审查 v1”、Prompt Application 或当前回归评测的同层切片。Prompt / Agent 继续复用现有 Run、Comparison、Evaluation 和发布治理真相源，不另建聚合发布真相源或自治执行器。
- 不从已关闭的 Application Interaction Session 派生长期记忆、自动 profile、重试 / fallback、schedule、replay / resume 或 agent loop。服务端 summary 当前保持 `no_entry`；未来只有真实跨页任务、稳定 owner / snapshot / cursor、性能预算与正式 quota / billing owner 同时成立时才重新评审。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
- 应用解除归档只恢复目录活动态；不自动创建、吊销、激活或改写 API Key、运行时绑定、会话、草案、候选、定义和运行记录。
