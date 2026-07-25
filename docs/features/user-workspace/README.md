# 用户工作区细专题入口

更新时间：2026-07-25

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [Agent / Copilot 应用档案版本审查与受控建议（开发 / 测试态）v1](agent-copilot-application-profile-version-review-controlled-suggestion-dev-test-v1.md)：批次 A 至批次 E 已完成；Profile、发布与 assignment、唯一受控建议、Session / Run、类型专属 Web 与双数据库真实验收形成闭环，专题关闭。
- [提示词应用模板版本审查与受控调用（开发 / 测试态）v1](prompt-application-template-version-review-controlled-invocation-dev-test-v1.md)：批次 A 至 E 均已完成，受限模板、双数据库 Template owner、Configuration Draft v3、Publish Candidate v3、显式 Runtime Assignment、受控 invocation、Session / Turn v2、Run v6、Web 与真实浏览器验收均有可复验证据，专题关闭。
- [Prompt Application 开发测试态使用指南](prompt-application-dev-test-usage-guide.md)：说明完整 Template → Configuration → Candidate Review → Assignment → Invocation / Session → Run Review 顺序，以及启动配置、身份权限、CAS、持久化与故障处理；所有能力仅限开发测试态。
- [应用开发工作区与发布准备审查 v1](application-development-workspace-release-readiness-review-v1.md)：批次 A 至 C 已完成并关闭；route-scoped evidence、精确 Draft / Run owner 重读、离线 revision 失败关闭、真实浏览器连续路径与 URL / console / network 隐私审计均有可复验证据。
- [应用受控运行开发测试态指南](application-controlled-runtime-dev-test-guide.md)：说明 Application RAG、Workflow Definition、Application Interaction Session、v4 / v5 运行记录与 Application Operations 的启动、资源准备、作用域、恢复、失败语义和隐私边界。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)：strict contract、三种 Session / Turn owner、exact authority reload、v5 / v4 单次委托、Web 易失交互工作区、双数据库 launcher 连续链、重启恢复、真实浏览器和敏感扫描均已完成，专题关闭。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：Gateway 认证、统一 `sqlite_dev` repository / 聚合 runtime、双数据库门禁、Web 一次性交接、真实浏览器连续路径、重启恢复与敏感信息复验均已完成，专题关闭。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。
- [应用运行观测与用量归因 v1](application-operations-observability-usage-attribution-v1.md)：已完成应用作用域 Gateway Request History 与 Workflow Run History 的独立来源覆盖、当前窗口归因摘要和合并时间线；不推测跨来源关联，不估算 token、成本、配额或计费。

## 下一步

- Agent / Copilot 专题已完成并关闭；下一步先设计新的用户工作区产品能力，不继续派生本专题同层 readiness、refresh 或 gate-only 批次。
- Prompt Application 批次 A 至 E 已完成并关闭：memory / SQLite / PostgreSQL 语义、Web、双数据库连续链、服务重启、CAS / drift / cancel 和敏感信息复验均已通过。
- 不继续扩“应用开发工作区与发布准备审查 v1”或 Prompt Application 的同层切片。Agent / Copilot 继续复用现有 workspace context、发布审查、Gateway、Session、Run History 和 Evaluation，不另建聚合发布真相源或自治执行器。
- 不从已关闭的 Application Interaction Session 派生长期记忆、自动 profile、重试 / fallback、schedule、replay / resume 或 agent loop。只有需要跨全部分页窗口的稳定统计、可信 reported usage 或正式 quota / billing owner 时，才评审服务端 summary。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
