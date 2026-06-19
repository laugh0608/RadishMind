# Admin Control Plane 设计与开发文档

更新时间：2026-06-14

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已有 tenant overview、audit log、Admin Operations Review / Readiness 和 Admin Provider/Profile & Deployment Evidence Review / Readiness。
- 现有页面只整理 readiness、evidence checklist、operator risk、provider/profile readiness、secret ref readiness 和 deployment status。
- Go read handlers 仍走 fake-store-backed repository bridge。
- 当前没有 Radish OIDC、token validation、auth middleware、真实数据库、repository adapter、secret resolver、deployment preflight 或 production admin 操作。

## 设计边界

- RadishMind 未来作为 `Radish` OIDC client，不自建第二套身份真相源。
- 管理端动作必须区分 read、review、propose、confirm、apply；当前只到 read/review。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. 在进入真实管理端前，先更新本功能文档，明确 tenant / role / permission / provider profile / route / quota 的写入职责。
2. 若要接 Radish OIDC，先形成 OIDC client design、claim mapping、tenant binding、failure taxonomy 和 smoke plan。
3. 若要接真实 read store，先从 repository adapter / store selector / schema migration 的单一方向推进，不与 auth 或管理写入并行打开。
4. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或管理动作才新增专项 gate。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
