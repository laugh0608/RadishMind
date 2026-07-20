# 用户工作区细专题入口

更新时间：2026-07-20

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [应用开发工作区与发布准备审查 v1](application-development-workspace-release-readiness-review-v1.md)：批次 A 已完成；唯一 application context、workspace / route generation、五阶段导航、生命周期失败关闭、当前阶段单独挂载、一次性跨阶段 Application API handoff 和发布准备停止线已进入 Web。当前为 `batch_a_completed_batch_b_ready`。
- [应用受控运行开发测试态指南](application-controlled-runtime-dev-test-guide.md)：说明 Application RAG、Workflow Definition、Application Interaction Session、v4 / v5 运行记录与 Application Operations 的启动、资源准备、作用域、恢复、失败语义和隐私边界。
- [应用交互会话与受控运行编排（开发 / 测试态）v1](application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)：strict contract、三种 Session / Turn owner、exact authority reload、v5 / v4 单次委托、Web 易失交互工作区、双数据库 launcher 连续链、重启恢复、真实浏览器和敏感扫描均已完成，专题关闭。
- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：Gateway 认证、统一 `sqlite_dev` repository / 聚合 runtime、双数据库门禁、Web 一次性交接、真实浏览器连续路径、重启恢复与敏感信息复验均已完成，专题关闭。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。
- [应用运行观测与用量归因 v1](application-operations-observability-usage-attribution-v1.md)：已完成应用作用域 Gateway Request History 与 Workflow Run History 的独立来源覆盖、当前窗口归因摘要和合并时间线；不推测跨来源关联，不估算 token、成本、配额或计费。

## 下一步

- 进入“应用开发工作区与发布准备审查 v1”批次 B：建立通用 feature-scoped 脱敏 handoff refs、来源分组、`partial_failure` / 漂移 / 缺失 blocker 与四态 readiness view model。不新增 API、schema、repository、发布记录或执行算法。
- 不从已关闭的 Application Interaction Session 派生长期记忆、自动 profile、重试 / fallback、schedule、replay / resume 或 agent loop。只有需要跨全部分页窗口的稳定统计、可信 reported usage 或正式 quota / billing owner 时，才评审服务端 summary。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
