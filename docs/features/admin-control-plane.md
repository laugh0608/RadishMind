# Admin Control Plane 设计与开发文档

更新时间：2026-08-23

## 功能定位

`Admin Control Plane` 面向平台管理员和运维，负责租户、用户、角色、权限、provider/profile、模型路由、API key、quota、price、secret backend、审计和部署状态。

## 当前状态

- `apps/radishmind-web/` 已完成 `S7 R1` Admin Control Plane 连续工作面与 `S9 R1` Admin Quota Admission：以 tenant / workspace / auth source / environment 为上下文，Tenant、User、Role、Audit、Provider、Profile、Route 与 Quota 八类资源只挂载一个当前 owner；旧 Operations / Deployment readiness 降为可折叠 supporting evidence，不再与主任务并列。
- Tenant / Audit 使用既有 authenticated read consumer；Audit 保持 `recorded_at_desc` 严格 cursor current window，并由当前行驱动 metadata-only detail。User / Role 已有本地 identity / role owner，但当前还没有 list repository / consumer，因此页面仍保持 blocked action；批次 D 前不得用离线 fixture 或 deterministic OIDC test 伪造用户、角色或 membership。
- Go read handlers 已由 selector 显式路由：Tenant Summary / Audit 使用 PostgreSQL dev/test；五条 workspace operation 已共享 `WorkspaceMembershipProvider`，Applications、API Keys、Workflow Definitions 与 Runs 复用 durable owner。旧 User Workspace `QuotaSummary` 因没有 application 选择与读权限契约继续明确关闭；独立的开发测试态 Admin quota owner 已完成 policy / usage / provider attempt admission 后端，不偷换旧投影。dev / signed-test membership 与 `local_session_dev_test` 本地 membership 已可确定性复验；legacy OIDC integration token 因没有 external identity binding 仍失败关闭，不回退本地 owner。
- 当前已有本地账户 / Web Session owner、开发测试态 session HTTP、deterministic Radish OIDC resource-server verifier，以及独立 browser Authorization Code + PKCE Relying Party；reviewed Radish client registration、真实 integration evidence、production token / session、secret resolver、deployment preflight 和 production admin 操作仍未完成。
- User Workspace 的 Application Publish Governance 已把正式 application repository、production auth / membership 和发布 owner 明确暴露为 promotion blocker；dev/test candidate approved 不会绕过这些 blocker。
- [Authenticated Read Store Transition v1](admin-control-plane/authenticated-read-store-transition-v1.md) 第一批 runtime 已完成 shared verified identity / negative auth，第二批已完成 Tenant / Audit PostgreSQL dev/test repository，第三批已完成 OIDC deterministic verifier / auth boundary / operation gate。
- [Tenant / Audit PostgreSQL Read Repository v1](admin-control-plane/tenant-audit-postgresql-read-repository-v1.md) 已完成两条 Admin operation 的 schema、manual migration、read-only role、routed selector、分页、no-fallback、真实 PostgreSQL、HTTP/Web 与浏览器验收。
- [Radish OIDC Integration Test v1](admin-control-plane/radish-oidc-integration-test-v1.md) 已完成 deterministic discovery / JWKS / JWT runtime、只开放 tenant / audit、五条 workspace route membership fail-closed、隐私和 Web 内存 token 边界；真实 Radish 联调为 `real_radish_integration_deferred`。
- [RadishMind 本地账户与 Radish OIDC 联合登录 v1](admin-control-plane/local-account-radish-oidc-federated-login-v1.md) 批次 A 至 C 已完成本地 identity owner、三种 repository、原子注册、Web Session、CSRF / Origin，以及独立 Authorization Code + PKCE、external identity resolve / create / link 和 session actor → local membership；下一步进入批次 D 的 Web 与管理面。
- [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](admin-control-plane/provider-profile-model-route-controlled-activation-dev-test-v1.md) 批次 A 至 E 已完成并关闭，覆盖原子配置领域、三模式 repository、Admin API、verified identity、四项独立权限、Gateway 不可变快照消费、Request History 谱系、Admin Web 与双数据库重启浏览器链。
- [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](gateway/application-api-key-request-quota-admission-dev-test-v1.md) 批次 A 至 E 已完成：`admin_gateway_quotas:read / write`、开发测试态环境显式绑定、三种 repository、UTC 日窗口、provider 前线性化准入，以及完整 Pencil、React 严格 consumer、CAS 确认和真实浏览器连续链均已成立；这不表示 production quota 或 billing 已就绪。
- `S7 R1` 五维评分为 `2 / 2 / 2 / 2 / 2`，采用 `A / 完整 Pencil`，只建立一个桌面与一个窄屏代表面。Provider / Profile / Route 三入口继续复用 `tenant_ref + workspace_id + environment + configuration_id` 原子 owner，不拆分第二套状态机或 inventory。

## 设计边界

- RadishMind 拥有服务自身的本地账号、角色、权限和工作区成员关系；作为 Radish OIDC client 时只把外部 `(issuer, subject)` 绑定到本地账户。当前不自建 OIDC issuer，不读取或复制 Radish 身份数据库。
- 管理端动作必须区分 read、draft、review、activate / rollback；Tenant / Audit 与身份边界当前只读，只有 Provider / Profile / Route 在 development / test 配置 owner 内开放四项独立权限和显式 generation 切换。
- 审计记录、secret reference 和 deployment evidence 只能展示脱敏摘要。
- 管理端 readiness 不等于 production ready，也不等于可以绕过人工确认。

## 下一批开发方向

1. `S7 R1` 七资源页面族与 `S9 R1` Admin Quota Admission 均已完成功能纵向切片；quota 保持独立 application policy / usage owner，不并入 Provider / Profile / Route 原子 owner。当前[Provider 价格策略版本与应用成本审查](gateway/provider-pricing-policy-version-application-cost-review-dev-test-v1.md)已落地独立 Pricing owner、两项权限、CAS、显式确认、已审 Visual R1 和 React strict consumer；价格仍不写入 Provider Route、quota 或 production billing。
2. `Radish OIDC Integration Test Runtime v1` deterministic resource-server 批次已完成；真实 bearer-token 联调主动 deferred。它不再承担浏览器登录或 User / Role owner。
3. local session、membership adapter 与 OIDC browser login 已完成；browser callback 只恢复本地 `user_id` 并签发 RadishMind Web Session，再由本地 owner 作 workspace / permission decision。批次 D 只消费这条链建设 Web 与管理面；既有 integration-token 模式继续在 repository 前返回 `workspace_membership_unavailable`，不能借本地 session 模式获得隐式 fallback。
4. Provider Profile assignment 与 Model Route 的开发测试态受控配置已完成；Admin 只保存既有 runtime inventory 引用，审批不自动启用，Gateway 只消费显式启用的不可变快照。不得从现有 Web 原地扩 production 配置。
5. 普通 evidence review 展示不再新增逐项 task card；只有真实 auth、数据库、secret、deployment 或新管理动作才新增专项 gate。
6. quota 只推进当前已批准的开发测试态 application request admission；身份专题不并行打开 billing、token / cost quota、production secret resolver、自动路由或 production deployment。

## 验收方式

- 只读展示类：web build、布局检查、fast baseline。
- read store 类：Go tests、repository contract smoke、read-side checker。
- auth / secret / deployment 类：专门 task card、负向测试、脱敏检查、no side effects 检查和全量仓库验证。
