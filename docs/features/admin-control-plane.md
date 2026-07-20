# Admin Control Plane 设计与开发文档

更新时间：2026-07-12

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已有 tenant overview、audit log、Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness。
- 现有页面只整理 readiness、evidence checklist、operator risk、provider/profile readiness、secret ref readiness 和 deployment status。
- Go read handlers 已由 selector 显式路由：Tenant Summary / Audit 使用 PostgreSQL dev/test，五条 workspace operation 在 signed test 模式保留 fake binding、在 OIDC integration 模式 fail closed；shared verified identity、signed test verifier 和 route authorization 已落地。
- 当前已有 deterministic Radish OIDC discovery / JWKS / JWT verifier，但没有 reviewed Radish upstream evidence、真实 integration token / evidence、production token / session、secret resolver、deployment preflight 或 production admin 操作。
- User Workspace 的 Application Publish Governance 已把正式 application repository、production auth / membership 和发布 owner 明确暴露为 promotion blocker；dev/test candidate approved 不会绕过这些 blocker。
- [Authenticated Read Store Transition v1](admin-control-plane/authenticated-read-store-transition-v1.md) 第一批 runtime 已完成 shared verified identity / negative auth，第二批已完成 Tenant / Audit PostgreSQL dev/test repository，第三批已完成 OIDC deterministic verifier / auth boundary / operation gate。
- [Tenant / Audit PostgreSQL Read Repository v1](admin-control-plane/tenant-audit-postgresql-read-repository-v1.md) 已完成两条 Admin operation 的 schema、manual migration、read-only role、routed selector、分页、no-fallback、真实 PostgreSQL、HTTP/Web 与浏览器验收。
- [Radish OIDC Integration Test v1](admin-control-plane/radish-oidc-integration-test-v1.md) 已完成 deterministic discovery / JWKS / JWT runtime、只开放 tenant / audit、五条 workspace route membership fail-closed、隐私和 Web 内存 token 边界；真实 Radish 联调为 `real_radish_integration_deferred`。

## 设计边界

- RadishMind 未来作为 Radish 注册的 application/client 与 resource server 接入，不自建第二套 issuer、账号、角色或身份真相源。
- 管理端动作必须区分 read、review、propose、confirm、apply；当前只到 read/review。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. `Radish OIDC Integration Test Runtime v1` deterministic 批次已完成；真实 Radish 联调主动 deferred，不再占用 Admin 当前开发顺位。
2. 未来 Radish 注册 RadishMind application/client 与 resource audience，并提供 reviewed issuer、JWKS policy、claim / permission mapping 和短期 token 流程后，再恢复只覆盖 tenant summary 与 audit summary 的真实联调。
3. 五条 workspace route 继续返回 `workspace_membership_unavailable`，不读取 fake repository；workspace membership contract 未成立前，不迁移五条 User Workspace read route。
4. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或管理动作才新增专项 gate。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
