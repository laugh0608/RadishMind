# Admin Control Plane 细专题入口

更新时间：2026-08-25

本目录承接 `Admin Control Plane` 中需要跨身份、权限、repository 和管理端使用路径推进的功能专题。产品面长期边界继续以 [Admin Control Plane 设计与开发文档](../admin-control-plane.md) 为准。

## 当前专题

- [本地账户凭证轮换与自助会话治理（开发 / 测试态）v1](local-account-credential-rotation-self-service-session-governance-dev-test-v1.md)：状态为 `local_account_credential_rotation_self_service_session_governance_dev_test_v1_design_proposed_review_required`。专题提议复用唯一 local identity owner，建立当前账户的 session directory、exact revoke、revoke others 与 credential rotation 原子链；当前等待项目所有者评审设计细节，尚未进入批次 A 代码。
- [本地用户、角色与工作区成员管理（开发 / 测试态）v1](local-user-role-workspace-membership-administration-dev-test-v1.md)：状态为 `local_user_role_workspace_membership_administration_dev_test_v1_completed`。批次 A 至 E 已在唯一 local identity owner 上完成 canonical role catalog、三存储目录 / mutation、显式 one-shot bootstrap CLI、七条 local-session-only strict Admin HTTP、批准 Pencil、React strict consumer、SQLite / PostgreSQL configured Server 产品连续链、三视口、双标签与隐私审计；专题关闭，不依赖或打开真实 Radish 与 production IAM。
- [RadishMind 本地账户与 Radish OIDC 联合登录 v1](local-account-radish-oidc-federated-login-v1.md)：批次 A 至 D 已完成六类本地身份 owner、三种 repository、原子注册、Web Session、CSRF / Origin、确定性 Authorization Code + PKCE、当前账户 / external identity revoke HTTP、完整 Pencil、strict Web、S7 当前账户 owner、no-fallback 与真实浏览器连续链。批次 E 等待真实 Radish 注册条件，不提前声明 production auth。
- [Admin Control Plane 设计与开发文档](../admin-control-plane.md)：`S7 R1` 已把 Tenant、User、Role、Audit、Provider、Profile 与 Route 编排为七任务单 owner 工作面。Tenant / Audit 复用既有 authenticated read；User / Role 复用 local identity strict consumer，提供 workspace 成员目录、exact detail、canonical 角色目录与受控 create / revoke；Provider / Profile / Route 复用同一开发测试态原子配置 owner。各 owner 保持独立权限与停止线，production IAM 未打开。
- [Provider Profile / Model Route 配置草案、版本审查与受控启用（开发 / 测试态）v1](provider-profile-model-route-controlled-activation-dev-test-v1.md)：五批开发已完成并关闭，覆盖领域与三模式 repository、Admin API / Auth、Gateway 不可变快照消费、Admin Web、SQLite / PostgreSQL 产品连续验证、服务重启和真实浏览器 activation / rollback；没有创建第二套 provider inventory，也没有读取真实 secret 或启用 production。
- [应用 API Key 请求配额与 Provider Attempt 准入（开发 / 测试态）v1](../gateway/application-api-key-request-quota-admission-dev-test-v1.md)：批次 A 至 E 已完成独立 quota owner、三模式 repository、Admin GET / PUT、独立 read / write 权限、六条 API Key inference route 的 provider 前原子准入，以及 `S9 Admin Quota Admission` 完整 Pencil、React 严格 consumer、CAS 确认和真实浏览器验收。旧 workspace quota summary 仍保持不可用，生产 quota、rate limit、token / cost、billing、正式 membership / OIDC 与自动路由未打开。
- [Authenticated Read Store Transition v1](authenticated-read-store-transition-v1.md)：第一批 verified identity / negative auth runtime 与第二批 Tenant / Audit PostgreSQL dev/test runtime 均已完成。
- [Tenant / Audit PostgreSQL Read Repository v1](tenant-audit-postgresql-read-repository-v1.md)：两条 Admin operation 已完成 projection schema、manual migration、read-only role、routed selector、keyset pagination、no-fallback、真实 PostgreSQL 与浏览器验收。
- [Radish OIDC Integration Test v1](radish-oidc-integration-test-v1.md)：deterministic discovery / JWKS / JWT verifier、两条 Admin operation gate、五条 workspace membership fail-closed 和 Web 内存 token consumer 已完成；真实 Radish 联调为 `real_radish_integration_deferred`，未来在 Radish 注册 RadishMind application/client 与 resource audience 后恢复。

## 目录停止线

- RadishMind 拥有平台本地用户、角色、权限与 workspace membership；Radish 保持自身 issuer、用户和业务授权真相。两者只通过 explicit external identity binding 联合，不复制数据库或按 email 自动合并。
- Admin read transition 不并入管理写入、application promotion、API key lifecycle、billing、secret runtime 或部署执行；开发测试态 quota 是独立 owner，不回填 Tenant / Audit read store。
- 每个实现批次只打开一个主要高风险边界；auth、membership、store 与真实 Radish 联调按顺序验收，不同时切换。
- 本地成员管理只接受 exact `user_id`、exact tenant / workspace 与 canonical built-in `role_key`；不开放全局账户搜索、客户端任意 grants、自定义角色、邀请或批量授权。新身份安全专题只处理当前账户 self-service，不反向扩大成员管理员权限。
- 当前 Provider / Route 管理专题只保存既有 runtime inventory 的引用、版本、审查与激活事实；credential、endpoint 和 provider raw config 不进入 Admin repository。
