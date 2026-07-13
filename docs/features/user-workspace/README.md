# 用户工作区细专题入口

更新时间：2026-07-13

本目录承接用户工作区中跨应用、模型发现、接入、调用与审查的具体功能专题。产品面长期边界继续以 [用户工作区设计与开发文档](../user-workspace.md) 为准。

## 当前专题

- [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md)：批次 A 已完成领域、内存存储、管理 API、一次性凭据交接和高风险负向边界；下一批接入 Gateway 认证与 PostgreSQL，最终闭环到请求历史和吊销拒绝。
- [应用目录与生命周期（开发/测试态）v1](application-catalog-lifecycle-dev-test-v1.md)：核心生命周期、内存与 PostgreSQL 开发测试态存储、Web 管理、下游归档只读约束和真实浏览器连续验收均已完成。
- [应用 API 接入与调用 v1](application-api-integration-invocation-v1.md)：把选中应用、`/v1/models` 模型目录、三协议接入示例、现有 Gateway 调试台调用与脱敏请求历史审查串成连续的内部开发者路径。
- [应用配置草案与审查 v1](application-configuration-draft-review-v1.md)：为当前应用建立独立配置草案、校验、开发测试态持久化、版本冲突、比较和 API 接入交接。
- [应用发布治理与晋级审查 v1](application-publish-governance-promotion-v1.md)：已完成不可变候选版本、版本绑定、审查 CAS、漂移识别、阻塞式晋级资格判断，以及既有接入区、调试台和请求历史交接；不直接发布正式应用。

## 下一步

- 按 [API 密钥生命周期与 Gateway 开发测试态认证 v1](api-key-lifecycle-gateway-dev-test-auth-v1.md) 和[实施任务卡](../../task-cards/api-key-lifecycle-gateway-dev-test-auth-v1-plan.md)推进批次 B：接入互斥 Gateway 认证模式、可信调用上下文、请求历史和独立 PostgreSQL 存储；不继续扩同层应用目录面板、检查器或证据切片。OIDC 成员关系未成立时继续失败关闭。

## 目录停止线

- 应用配置草案只允许建立独立开发测试态存储库，不改变 Gateway 上行协议 schema、模型服务注册表或正式应用真相源。
- 开发测试态 API 密钥必须显式启用、只绑定活跃应用且失败关闭；不把它晋级为生产 API 密钥，也不并行打开配额、计费、自动回退、负载均衡或生产认证。
- 应用创建、发布、删除和业务写回必须由独立功能设计承接，不并入既有接入与调用工作区。
