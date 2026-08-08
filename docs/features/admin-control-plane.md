# Admin Control Plane 设计与开发文档

更新时间：2026-08-08

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已完成 `S7 R1` Admin Control Plane 连续工作面：以 tenant / workspace / auth source / environment 为上下文，Tenant、User、Role、Audit、Provider、Profile 与 Route 七类资源只挂载一个当前 owner；旧 Operations / Deployment readiness 降为可折叠 supporting evidence，不再与主任务并列。
- Tenant / Audit 使用既有 authenticated read consumer；Audit 保持 `recorded_at_desc` 严格 cursor current window，并由当前行驱动 metadata-only detail。User / Role 没有 RadishMind list repository 或 consumer，页面明确显示 Radish-owned 真相边界和 blocked action，不用离线 fixture 或 deterministic OIDC test 伪造 membership。
- Go read handlers 已由 selector 显式路由：Tenant Summary / Audit 使用 PostgreSQL dev/test；五条 workspace operation 已共享 `WorkspaceMembershipProvider`，Applications、API Keys、Workflow Definitions 与 Runs 复用 durable owner，Quota 在缺少可信 policy owner 时明确关闭。dev / signed-test membership 已可确定性复验，OIDC integration 仍因缺少 reviewed membership contract 而失败关闭。
- 当前已有 deterministic Radish OIDC discovery / JWKS / JWT verifier，但没有 reviewed Radish upstream evidence、真实 integration token / evidence、production token / session、secret resolver、deployment preflight 或 production admin 操作。
- User Workspace 的 Application Publish Governance 已把正式 application repository、production auth / membership 和发布 owner 明确暴露为 promotion blocker；dev/test candidate approved 不会绕过这些 blocker。
- [Authenticated Read Store Transition v1](admin-control-plane/authenticated-read-store-transition-v1.md) 第一批 runtime 已完成 shared verified identity / negative auth，第二批已完成 Tenant / Audit PostgreSQL dev/test repository，第三批已完成 OIDC deterministic verifier / auth boundary / operation gate。
- [Tenant / Audit PostgreSQL Read Repository v1](admin-control-plane/tenant-audit-postgresql-read-repository-v1.md) 已完成两条 Admin operation 的 schema、manual migration、read-only role、routed selector、分页、no-fallback、真实 PostgreSQL、HTTP/Web 与浏览器验收。
- [Radish OIDC Integration Test v1](admin-control-plane/radish-oidc-integration-test-v1.md) 已完成 deterministic discovery / JWKS / JWT runtime、只开放 tenant / audit、五条 workspace route membership fail-closed、隐私和 Web 内存 token 边界；真实 Radish 联调为 `real_radish_integration_deferred`。
- [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](admin-control-plane/provider-profile-model-route-controlled-activation-dev-test-v1.md) 批次 A 至 E 已完成并关闭，覆盖原子配置领域、三模式 repository、Admin API、verified identity、四项独立权限、Gateway 不可变快照消费、Request History 谱系、Admin Web 与双数据库重启浏览器链。
- `S7 R1` 五维评分为 `2 / 2 / 2 / 2 / 2`，采用 `A / 完整 Pencil`，只建立一个桌面与一个窄屏代表面。Provider / Profile / Route 三入口继续复用 `tenant_ref + workspace_id + environment + configuration_id` 原子 owner，不拆分第二套状态机或 inventory。

## 设计边界

- RadishMind 未来作为 Radish 注册的 application/client 与 resource server 接入，不自建第二套 issuer、账号、角色或身份真相源。
- 管理端动作必须区分 read、draft、review、activate / rollback；Tenant / Audit 与身份边界当前只读，只有 Provider / Profile / Route 在 development / test 配置 owner 内开放四项独立权限和显式 generation 切换。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. `S7 R1` 页面族产品化已完成，Admin 当前没有待实现的同层 UI 切片；新的管理能力必须先进入独立功能设计专题并证明真实 owner、权限和执行边界。
2. `Radish OIDC Integration Test Runtime v1` deterministic 批次已完成；真实 Radish 联调主动 deferred。未来 Radish 注册 RadishMind application/client 与 resource audience，并提供 reviewed issuer、JWKS policy、claim / permission mapping 和短期 token 流程后，才恢复 Tenant / Audit 真实联调与 User / Role consumer 讨论。
3. OIDC 模式继续在 repository 前返回 `workspace_membership_unavailable`，直到 reviewed Radish membership owner / endpoint、撤销 / 过期语义和 claim mapping 成立；read permission 不得直接等同于 mutation authority。
4. Provider Profile assignment 与 Model Route 的开发测试态受控配置已完成；Admin 只保存既有 runtime inventory 引用，审批不自动启用，Gateway 只消费显式启用的不可变快照。不得从现有 Web 原地扩 production 配置。
5. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或新管理动作才新增专项 gate。
6. 不并行打开 quota / billing、production secret resolver、自动路由、真实 Radish OIDC 或 production deployment。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
