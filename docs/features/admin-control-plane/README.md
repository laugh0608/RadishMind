# Admin Control Plane 细专题入口

更新时间：2026-07-26

本目录承接 `Admin Control Plane` 中需要跨身份、权限、repository 和管理端使用路径推进的功能专题。产品面长期边界继续以 [Admin Control Plane 设计与开发文档](../admin-control-plane.md) 为准。

## 当前专题

- [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](provider-profile-model-route-controlled-activation-dev-test-v1.md)：当前在制产品专题，按领域与内存基线、开发测试持久化、Admin API、Gateway 快照消费和 Admin Web 五批推进；不会创建第二套 provider inventory，也不会读取真实 secret 或启用 production。
- [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md)：第一批 verified identity / negative auth runtime 与第二批 Tenant / Audit PostgreSQL dev/test runtime 均已完成。
- [Tenant / Audit PostgreSQL Read Repository v1](tenant-audit-postgresql-read-repository-v1.md)：两条 Admin operation 已完成 projection schema、manual migration、read-only role、routed selector、keyset pagination、no-fallback、真实 PostgreSQL 与浏览器验收。
- [Radish OIDC Integration Test v1](radish-oidc-integration-test-v1.md)：deterministic discovery / JWKS / JWT verifier、两条 Admin operation gate、五条 workspace membership fail-closed 和 Web 内存 token consumer 已完成；真实 Radish 联调为 `real_radish_integration_deferred`，未来在 Radish 注册 RadishMind application/client 与 resource audience 后恢复。

## 目录停止线

- Radish 继续拥有用户、tenant、role、permission 与 OIDC 真相；RadishMind 不复制身份系统。
- Admin read transition 不并入管理写入、application promotion、API key lifecycle、quota enforcement、billing、secret runtime 或部署执行。
- 每个实现批次只打开一个主要高风险边界；auth、membership、store 与真实 Radish 联调按顺序验收，不同时切换。
- 当前 Provider / Route 管理专题只保存既有 runtime inventory 的引用、版本、审查与激活事实；credential、endpoint 和 provider raw config 不进入 Admin repository。
