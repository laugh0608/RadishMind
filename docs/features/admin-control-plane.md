# Admin Control Plane 设计与开发文档

更新时间：2026-07-12

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已有 tenant overview、audit log、Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness。
- 现有页面只整理 readiness、evidence checklist、operator risk、provider/profile readiness、secret ref readiness 和 deployment status。
- Go read handlers 仍走 fake-store-backed repository bridge。
- 当前没有 Radish OIDC、token validation、auth middleware、真实数据库、repository adapter、secret resolver、deployment preflight 或 production admin 操作。
- User Workspace 的 Application Publish Governance 已把正式 application repository、production auth / membership 和发布 owner 明确暴露为 promotion blocker；dev/test candidate approved 不会绕过这些 blocker。
- [Authenticated Read Store Transition v1](admin-control-plane/authenticated-read-store-transition-v1.md) 已固定 verified identity、tenant permission binding、Admin tenant / audit durable read repository、auth / store mode compatibility、HTTP failure、隐私边界和分阶段实施顺序；当前状态仍是设计完成、runtime 未开始。

## 设计边界

- RadishMind 未来作为 `Radish` OIDC client，不自建第二套身份真相源。
- 管理端动作必须区分 read、review、propose、confirm、apply；当前只到 read/review。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. 下一实现入口创建 `Control Plane Verified Identity Context & Negative Auth Runtime v1` 高风险任务卡，先完成共享 verified identity、permission projection、signed test token、13 类负向认证和 HTTP failure 语义。
2. 第一批 repository 继续使用 fake store，禁止同时创建 PostgreSQL adapter；identity runtime 通过后，再为 Admin tenant / audit PostgreSQL read repository 创建独立任务卡。
3. Radish issuer / client registration / permission mapping evidence 未审查前，不接真实 OIDC；workspace membership contract 未成立前，不迁移五条 User Workspace read route。
4. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或管理动作才新增专项 gate。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
