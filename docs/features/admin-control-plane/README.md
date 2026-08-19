# Admin Control Plane 细专题入口

更新时间：2026-08-19

本目录承接 `Admin Control Plane` 中需要跨身份、权限、repository 和管理端使用路径推进的功能专题。产品面长期边界继续以 [Admin Control Plane 设计与开发文档](../admin-control-plane.md) 为准。

## 当前专题

- [RadishMind 本地账户与 Radish OIDC 联合登录 v1](local-account-radish-oidc-federated-login-v1.md)：当前活跃专题；批次 A 已完成六类本地身份 owner、PBKDF2 凭证、memory / SQLite / PostgreSQL repository、migration、版本并发、session revoke 与本地 membership adapter。下一步进入批次 B 本地注册、登录和 Web Session HTTP，不提前打开 OIDC callback。
- [Admin Control Plane 设计与开发文档](../admin-control-plane.md)：`S7 R1` 已把 Tenant、User、Role、Audit、Provider、Profile 与 Route 编排为七任务单 owner 工作面。Tenant / Audit 复用既有 authenticated read；User / Role 明确失败关闭；Provider / Profile / Route 复用同一开发测试态原子配置 owner。Pencil、React、关键断点和真实浏览器验收已完成，未扩 API、schema、repository、permission 或生产边界。
- [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](provider-profile-model-route-controlled-activation-dev-test-v1.md)：五批开发已完成并关闭，覆盖领域与三模式 repository、Admin API / Auth、Gateway 不可变快照消费、Admin Web、SQLite / PostgreSQL 产品连续验证、服务重启和真实浏览器 activation / rollback；没有创建第二套 provider inventory，也没有读取真实 secret 或启用 production。
- [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](../gateway/application-api-key-request-quota-admission-dev-test-v1.md)：批次 A 至 E 已完成独立 quota owner、三模式 repository、Admin GET / PUT、独立 read / write 权限、六条 API Key inference route 的 provider 前原子准入，以及 `S9 Admin Quota Admission` 完整 Pencil、React 严格 consumer、CAS 确认和真实浏览器验收。旧 workspace quota summary 仍保持不可用，生产 quota、rate limit、token / cost、billing、正式 membership / OIDC 与自动路由未打开。
- [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md)：第一批 verified identity / negative auth runtime 与第二批 Tenant / Audit PostgreSQL dev/test runtime 均已完成。
- [Tenant / Audit PostgreSQL Read Repository v1](tenant-audit-postgresql-read-repository-v1.md)：两条 Admin operation 已完成 projection schema、manual migration、read-only role、routed selector、keyset pagination、no-fallback、真实 PostgreSQL 与浏览器验收。
- [Radish OIDC Integration Test v1](radish-oidc-integration-test-v1.md)：deterministic discovery / JWKS / JWT verifier、两条 Admin operation gate、五条 workspace membership fail-closed 和 Web 内存 token consumer 已完成；真实 Radish 联调为 `real_radish_integration_deferred`，未来在 Radish 注册 RadishMind application/client 与 resource audience 后恢复。

## 目录停止线

- RadishMind 拥有平台本地用户、角色、权限与 workspace membership；Radish 保持自身 issuer、用户和业务授权真相。两者只通过 explicit external identity binding 联合，不复制数据库或按 email 自动合并。
- Admin read transition 不并入管理写入、application promotion、API key lifecycle、billing、secret runtime 或部署执行；开发测试态 quota 是独立 owner，不回填 Tenant / Audit read store。
- 每个实现批次只打开一个主要高风险边界；auth、membership、store 与真实 Radish 联调按顺序验收，不同时切换。
- 当前 Provider / Route 管理专题只保存既有 runtime inventory 的引用、版本、审查与激活事实；credential、endpoint 和 provider raw config 不进入 Admin repository。
