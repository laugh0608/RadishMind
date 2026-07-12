# Admin Control Plane 设计与开发文档

更新时间：2026-07-12

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已有 tenant overview、audit log、Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness。
- 现有页面只整理 readiness、evidence checklist、operator risk、provider/profile readiness、secret ref readiness 和 deployment status。
- Go read handlers 仍走 fake-store-backed repository bridge，但 shared verified identity、signed test verifier 和 route authorization 已落地。
- 当前没有真实 Radish OIDC / JWKS、production token / session、Control Plane PostgreSQL adapter、secret resolver、deployment preflight 或 production admin 操作。
- User Workspace 的 Application Publish Governance 已把正式 application repository、production auth / membership 和发布 owner 明确暴露为 promotion blocker；dev/test candidate approved 不会绕过这些 blocker。
- [Authenticated Read Store Transition v1](admin-control-plane/authenticated-read-store-transition-v1.md) 第一批 runtime 已完成：shared verified identity、RS256 signed test token、permission projection、13 类负向认证、七条 route authorization 和 Web sanitized denial state 均可复验；repository 仍为 fake store。
- [Tenant / Audit PostgreSQL Read Repository v1](admin-control-plane/tenant-audit-postgresql-read-repository-v1.md) 已完成产品设计，固定两条 Admin operation 的 projection owner、schema、manual migration、read-only role、routed selector、分页、重启恢复和 no-fallback；尚未实现数据库 runtime。

## 设计边界

- RadishMind 未来作为 `Radish` OIDC client，不自建第二套身份真相源。
- 管理端动作必须区分 read、review、propose、confirm、apply；当前只到 read/review。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. 下一实现入口创建 `Admin Tenant / Audit PostgreSQL Read Repository Runtime v1` 高风险任务卡，统一承接 migration、read-only role、adapter、routed selector、cursor 与 integration/browser evidence。
2. 实现只迁移 tenant summary 与 audit summary，继续使用 signed test token；不把五条 workspace route 一并接入数据库，也不派生同层 readiness 文档链。
3. Radish issuer / client registration / permission mapping evidence 未审查前，不接真实 OIDC；workspace membership contract 未成立前，不迁移五条 User Workspace read route。
4. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或管理动作才新增专项 gate。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
