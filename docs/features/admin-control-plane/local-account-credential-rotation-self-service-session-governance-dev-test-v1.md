# 本地账户凭证轮换与自助会话治理（开发 / 测试态）v1

更新时间：2026-08-26

状态：`local_account_credential_rotation_self_service_session_governance_dev_test_v1_batch_d_react_completed_batch_e_ready`

## 功能定位

本功能在 RadishMind 已有本地账户、密码凭证、Web Session、三种开发测试态 repository 与 Authentication / Account surface 之上，为当前已登录账户提供自己的会话审查、受控会话撤销和本地密码凭证轮换。

它补齐的是联合登录批次 A 至 D 与本地成员管理专题关闭后的身份安全缺口：专题启动时页面只能看到当前会话，原 `POST /v1/auth/sessions/{session_id}/revoke` 也只允许撤销当前会话；repository 虽已有 `ReplaceCredential`、`ReadWebSession` 与 `RevokeWebSession` 原语，却没有 user-scoped session directory，也没有“凭证替换与来源会话失效”的单一原子 owner。批次 A 至 D 已依次补齐 owner、双数据库、strict HTTP、完整 Pencil 与 React strict consumer；双数据库产品连续链仍由批次 E 承接。

本功能只处理当前账户自助安全操作。账户禁用、管理员代重置、账户恢复、MFA、生产 session store、真实 Radish 联调和 production auth 继续由后续独立目标承接。

## 选择依据

- 当前阶段要求选择一个用户可感知、复用 canonical owner、可在本仓库确定性验证的新长期目标。
- `UserAccount`、`LocalCredential` 与 `WebSession` 已由同一 `localIdentityRepository` 在 memory、SQLite、PostgreSQL 中拥有，不需要第二套账户、凭证或会话真相源。
- 密码凭证轮换、其它会话撤销与跨标签失效是注册 / 登录之后的直接用户闭环，不依赖 reviewed Radish client registration、真实 Provider、production secret 或上层项目挂载点。
- 该能力涉及密码、会话、并发、API、数据库和浏览器隐私，必须先完成独立功能设计与高风险任务卡，不能作为现有成员管理批次 F 或普通 UI 小切片进入。

## 已批准产品决策

- 当前本地 `user_id` 是 self-service scope 的唯一账户身份；session cookie 只恢复 actor，客户端不能提交或覆盖目标 `user_id`。
- self-service session directory 只列出当前账户自己的 session。它不是管理员全局 session console，也不接受 login identifier、email、OIDC issuer / subject 或 workspace 作为账户搜索条件。
- session summary 只允许返回 `session_id`、authentication method、effective state、record version、current-session 标记以及创建、最近验证、过期、撤销时间。`AuthenticationSourceRef`、cookie、credential digest、IP、原始 User-Agent、raw claim 和 audit body 不进入响应。
- 本地凭证轮换要求有效 Web Session、同源 Origin、CSRF、当前密码验证、新密码策略校验和显式 session 影响确认。OIDC-only 账户不能借此隐式创建本地密码登录方式。
- 凭证轮换必须在同一 repository transaction 中 supersede 当前 active `LocalCredential`，并撤销全部由该旧 credential 建立且仍 active 的 `local_password` Web Session；任一步失败都不得替换凭证或部分撤销会话。
- 当前请求若使用将被 supersede 的本地密码 session，成功后该 session 同样失效，服务端清理当前认证 cookie，用户必须使用新密码重新登录。当前请求若使用 OIDC session，OIDC session 不因本地密码轮换自动失效；用户可通过会话治理显式撤销它。
- 单会话撤销允许把当前账户的任意 exact session 作为 target；撤销当前 session 时清理 cookie。批量操作只允许“撤销其它全部 active session”，保留当前 session，且必须在单一 transaction 中完成。
- session 状态由 `lifecycle_state + expires_at + snapshot_at` 确定。列表 cursor 必须绑定 exact `user_id`、state filter、limit、`snapshot_at`、`created_at DESC + session_id DESC` 位置，不能在过期边界漂移后混页或回退上次结果。
- 所有成功 mutation 返回 canonical server result 或 `204`；Web 不自行推算 credential、session lifecycle、版本或撤销数量。

## 目标用户与核心流程

目标用户是使用 RadishMind `local_session_dev_test` 登录的内部开发者和团队成员。

### 会话审查与撤销

1. 用户打开 Authentication / Account surface，读取当前账户 profile 与 user-scoped session page。
2. 页面明确标记当前 session、认证方式、effective state 与时间边界，不展示设备指纹或推断位置。
3. 用户选择一个自己的 session，确认 exact session、认证方式、过期时间和撤销影响。
4. 服务端重读当前 actor、目标 session ownership、record version 和 recent authentication 后执行 CAS revoke。
5. 用户也可以显式撤销其它全部 active session；服务端在单一 transaction 中重读并撤销，当前 session 保持 active。
6. mutation 成功后 Web 丢弃旧 page、selection、confirmation 和迟到响应，重新读取 canonical session page；跨标签只广播 metadata-only `session_changed` 信号。

### 本地密码凭证轮换

1. 用户提交当前密码、新密码和显式 session 影响确认；浏览器不预读或持久化 credential metadata。
2. 服务端从当前 session 恢复 exact `user_id`，在 repository transaction 中读取 active local credential 并校验当前密码。
3. 新密码必须通过当前 allowlist policy，且不能与当前密码等价；原始密码只存在于单次请求内存。
4. 服务端创建新的 versioned credential、supersede 旧 credential，并撤销全部 source-bound active local-password sessions。
5. 成功后若当前 session 被撤销，响应清理认证 cookie，并要求重新登录；若当前 session 来自 OIDC，则保留该 session，但仍重新读取账户与 session 状态。
6. store、并发、验证或撤销任一步失败时，旧 credential 和全部 session 保持原状，不回退 memory、fixture 或上次成功结果。

## Owner 与数据边界

| 资源 | 唯一 owner | 本功能允许 | 明确排除 |
| --- | --- | --- | --- |
| `UserAccount` | RadishMind Identity | active actor 与 self-service scope 校验 | 禁用、删除、全局账户目录 |
| `LocalCredential` | RadishMind Credential | 当前密码验证、versioned replacement、supersede | 管理员代重置、恢复 token、密码提示、raw password persistence |
| `WebSession` | RadishMind Session | user-scoped list、exact revoke、revoke others、source-bound atomic revoke | production session store、上游 token revoke、设备追踪、地理位置推断 |
| `ExternalIdentityBinding` | RadishMind Identity | 仅判断当前认证方式与保留登录方式 | 新增 / 解绑、Radish 目录同步、claim 授权 |
| `WorkspaceMembership` | RadishMind Authorization | 不参与 self-service ownership | 把 workspace admin 当作全局账户安全管理员 |

首版不新增独立 credential history、session audit store、device table 或安全事件 owner。旧 credential 与 revoked session 继续保留在既有 local identity repository 中作为 lifecycle fact；生产 retention、purge 与合规策略不在本功能内定义。

## 开发测试态 HTTP 边界

批次 C 已固定四条入口的 exact contract：

- `GET /v1/auth/sessions`：当前账户 session page，支持严格 state filter、limit 与 filter-bound cursor。
- `POST /v1/auth/sessions/{session_id}/revoke`：撤销当前账户的 exact session，要求 expected version、recent authentication 与显式确认。
- `POST /v1/auth/sessions/revoke-others`：原子撤销当前账户除当前 session 外的全部 active session。
- `POST /v1/auth/local/credential/rotate`：校验当前密码，原子替换 active credential 并撤销旧 credential 的 source-bound local-password sessions。

四条入口只在显式 `local_session_dev_test` 下启用。Bearer、dev header、signed-test token、OIDC resource-server token、payload `user_id`、active workspace 或 membership grant 都不能替代当前本地 Web Session。所有 mutation 使用 exact Origin、double-submit CSRF、strict JSON、大小预算、未知 / 重复字段拒绝、request ref 与脱敏 audit ref。

## 稳定失败分类

- `local_identity_session_cursor_invalid`
- `local_identity_session_scope_denied`
- `local_identity_session_version_conflict`
- `local_identity_session_recent_authentication_required`
- `local_identity_session_bulk_revoke_conflict`
- `local_identity_credential_unavailable`
- `local_identity_credential_current_invalid`
- `local_identity_credential_policy_rejected`
- `local_identity_credential_reuse_denied`
- `local_identity_credential_rotation_conflict`
- `local_identity_service_unavailable`

无权、缺失或跨账户 target 使用同一脱敏 scope failure，不证明其它账户或 session 是否存在。公开错误不包含 login identifier、credential id、password policy internals、cookie、digest、SQL、raw auth detail 或被撤销 session 的认证来源引用。

## Web 与 Family UI

- 本功能继续使用现有 Authentication / Account surface，不建立 S11，也不复制 S7 User / Role 的 workspace member owner。
- 五维评分提议为 `0 / 2 / 2 / 1 / 2 = 7`，由交互新颖度、状态复杂度和身份风险触发 `A / 完整 Pencil`；只在现有 Authentication Gateway 页面族补一个 Desktop、一个无法从 Desktop 推导的 Narrow 和一个关键危险操作状态。
- session directory 是主 owner；credential rotation 是从属安全动作。当前 session、其它 active session、expired / revoked history、单项确认和 bulk confirmation 必须保持清楚层级。
- 密码只存在受控 input state；提交完成、取消、账户 / session 改变、组件卸载和路由离开时立即清空。不得进入 URL、Web Storage、IndexedDB、service worker、日志、截图或跨标签消息。
- 页面复用 Family UI semantic token、Workbench 连续窗格、单一主对象和职责圆角；危险状态同时使用文字、符号和语义色，不以颜色作为唯一通道。

## 实施拆分

### 批次 A：领域合同与 memory 原子链

- 定义 session summary/page/cursor、self-service actor、exact revoke、revoke others 与 credential rotation result。
- 在唯一 local identity owner 中增加 user-scoped session query，以及 credential replacement + source-bound session revoke 的 aggregate transaction。
- memory 覆盖超过 `100` 条同时间戳分页、snapshot-bound cursor、ownership、recent-auth、current / other session、bulk atomicity、password reuse、并发单胜者与零部分写入。
- 不注册 HTTP，不修改 config、migration、Pencil 或 Web。

完成证据（2026-08-25）：

- 项目所有者已批准 owner、原子语义、HTTP 提议、Pencil 覆盖和停止线；代码采用独立 `localIdentitySelfServiceSecurityRepository` capability interface 接入既有 `memoryLocalIdentityRepository`，没有创建平行 identity、credential 或 session owner。
- 已落地 `local_identity_self_service_session_summary.v1`、page、`snapshot_at`、filter-bound cursor、self-service actor，以及 exact revoke、revoke others、credential rotation 三类 canonical result。
- memory session page 以 exact `user_id`、state、limit、`snapshot_at` 和 `created_at DESC + session_id DESC` 绑定 cursor；effective state 按快照计算，`121` 条同时间戳、多页过期边界、后插入 session 排除、owner / filter 漂移和敏感字段扫描已通过。
- exact revoke 以 owner + expected version 执行；revoke others 在单锁中预构建全部目标后一次提交；credential rotation 在同一锁中验证 active account、active credential、当前密码与复用边界，再一起 supersede 旧 credential 并撤销全部 source-bound local-password session。当前 local session 会进入撤销集，当前 OIDC session 保持 active。
- bulk 与 rotation 均覆盖坏目标导致的零部分写入；credential rotation 并发四争用者只有一个赢家。新增精准测试、完整 `internal/httpapi` race、完整 Platform 测试、`go vet`、PostgreSQL tagged compile 及仓库 fast / full 门禁已通过。
- 批次 A 没有注册 HTTP、修改 config / migration / Pencil / React / CSS、增加 fixture / checker 或打开 production 能力；下一准入只进入批次 B 的双数据库 durable owner 与 query-plan 审查。

### 批次 B：SQLite / PostgreSQL durable owner

- 实现与 memory 同构的稳定列表与原子 mutation；先通过 query plan 证明是否需要最小顺序索引。
- 覆盖 migration / rollback / reapply、受限 runtime role、CAS、并发、服务重启、store unavailable 与 no-fallback。
- 不创建第二套 session table、credential history table、ORM 或 production repository mode。

完成证据（2026-08-25）：

- SQLite / PostgreSQL 已由同一 `localIdentitySelfServiceSecurityRepository` capability 提供 durable owner；session page 直接读取既有 `local_web_sessions`，mutation 在单一 transaction 中重读 account、credential 与 user-scoped sessions，再提交 exact revoke、revoke others 或 credential replacement + source-bound revoke。没有扩张通用 repository interface，也没有增加第二套 table、ORM 或 fallback owner。
- 原有 `local_web_sessions_active_user_idx` 只覆盖 active session 的过期查询，不能满足 `created_at DESC + session_id DESC` 的稳定顺序。双库因此只追加 `0004_local_identity_self_service_sessions` / `local_identity_records_store_v4` 与 `local_web_sessions_self_service_list_idx`；SQLite `EXPLAIN QUERY PLAN` 和 PostgreSQL `ANALYZE + EXPLAIN` 均证明列表使用该 ordered index，未改写既有表或历史 migration。
- durable contract 已覆盖 snapshot-bound 多页、同时间戳顺序、后插入排除、post-snapshot revoke、exact CAS、并发双争用单胜者、bulk 保留 current / effective-expired session、OIDC current session 保留，以及 active / expired source-bound local-password session 随 credential rotation 一起撤销。
- 以另一账户的 credential id 制造数据库唯一键冲突后，旧 active credential 与全部 source-bound session 均保持原状，证明 credential supersede 后续失败会整事务回滚；成功轮换后新密码可验证、旧 credential superseded，三条旧来源 session 同时 revoked。
- SQLite v3 → v4 升级、重放、重启、关闭连接 no-fallback 已通过；PostgreSQL v1 / v3 → v4、既有数据保留、受限 runtime role DDL 拒绝、rollback / reapply、pool 重连和回滚后 no-fallback 已在 PostgreSQL 17 聚合集成中通过。测试容器、网络和临时 volume 已清理。
- 批次 B 没有注册 HTTP、修改 config / Server startup / Pencil / React / CSS 或打开 production store；下一准入只进入批次 C 的四条 strict HTTP 与 local Web Session 授权。

### 批次 C：strict HTTP 与本地会话授权

- 固定四条 route、strict request / response、CSRF / Origin、recent authentication、current-password proof、confirmation 与稳定失败映射。
- 扩展既有 session revoke 语义时保持 self ownership；旧 current-session logout 继续走 `POST /v1/auth/logout`。
- 覆盖响应敏感字段扫描、错误 method、跨账户、stale version、重复提交和业务副作用为零。

完成证据（2026-08-25）：

- `GET /v1/auth/sessions`、exact revoke、revoke others 与 local credential rotate 已注册为四条 exact route；mutation body 分别固定为 `expected_record_version + confirmed`、`confirmed` 和 `current_password + new_password + session_impact_confirmed`，未知 / 重复字段、多份 JSON、非法或重复 query 与错误 method 均失败关闭。
- Server 只从现有本地 session middleware 恢复 `local_session_dev_test` actor，并再次绑定 exact `user_id + current_session_id`；dev header、Bearer、active workspace 与 membership 不能形成 fallback。列表成功刷新 CSRF cookie，但不返回 credential id、authentication source ref、digest、audit ref、login identifier、cookie 或 password。
- 三类 mutation 复用 exact Origin 与 double-submit CSRF；exact / bulk 要求十分钟近期认证，credential rotation 同时要求近期认证、当前密码 proof 与显式 session impact confirmation。跨账户或缺失 target 统一返回 `local_identity_session_scope_denied`，stale CAS、bulk conflict、credential unavailable / invalid / policy / reuse / conflict 均映射到稳定脱敏失败码。
- exact revoke 只调用 owned-session repository operation；bulk 与 rotation 只调用各自 aggregate operation。撤销当前 session 或 local-password 当前 session 随 credential rotation 失效时同时清理 session / CSRF cookie；撤销其它 session、revoke others 或 OIDC 当前 session 轮换本地密码时不改写当前 session cookie。
- HTTP 专项覆盖分页与 tamper cursor、跨账户、stale / duplicate CAS、bulk 重复提交零影响、current / other / OIDC cookie 分流、错误 / 复用 / 新密码、Origin / CSRF、recent-auth、strict JSON、响应与日志敏感字段扫描及失败后零业务副作用。定向 race、完整 Platform、`go vet` 与 PostgreSQL tagged compile 已通过。
- 批次 C 未修改 config gate、migration、Pencil、React 或 CSS，未新增 MFA、恢复、管理员代重置、refresh token、OIDC upstream logout 或 production capability；下一准入只进入批次 D 的完整 Pencil 与 React strict consumer。

### 批次 D：Pencil 与 React strict consumer

- 完成设计提议的 Authentication Gateway Desktop / Narrow / danger state，并在人工批准后实施。
- 建立单一 strict consumer、session directory、exact selection、credential rotation 和跨标签 metadata-only invalidation。
- 覆盖 loading、empty、expired、revoked、denied、unavailable、conflict、success、forced re-login 与窄屏顺序。

Pencil 完成证据（2026-08-25）：

- Family UI 正确设计源已新增 Desktop `pOLcz`、Narrow `LMi7H`、credential rotation danger state `n2O8A5` 与 Decision Record `DASE0`；四张根画板均无 placeholder，Pencil 全树布局检查无裁切、越界或循环尺寸问题。
- Desktop 保持 session directory 单一主 owner，覆盖 current / other、active / expired / revoked、exact selected target，并把 local credential rotation 限制为右侧从属动作。
- Narrow 固定 `current → other active → ended history → credential disclosure → bulk confirmation` 顺序；bulk danger sheet 显式列出 exact target set，并说明当前 session 保留。
- danger state 使用空 input placeholder，展示当前 local-password session 会进入撤销集、OIDC session 保留、事务失败时旧 credential 与全部 session 不变，以及成功后的 cookie 清理和 forced re-login。
- R21 记录五维评分 `0 / 2 / 2 / 1 / 2 = 7`、全状态矩阵、selection / cursor / confirmation / late response 失效、密码组件内存清理和敏感材料禁入边界。项目所有者已人工批准四张代表板与 R21；React / CSS 尚未实施。

React 完成证据（2026-08-26）：

- Authentication Gateway 已接入单一 `LocalIdentitySelfServiceSecurityPanel`，没有建立 S11 或平行 session / credential owner。页面以 current session、其它 active session 和 ended history 三组投影 canonical directory，并提供 exact revoke、aggregate revoke others 与从属 credential rotation。
- 单一 strict consumer 已覆盖四条 route，严格解析 canonical schema、cursor / filter / state / actor scope 和稳定错误边界；响应中任意未知字段、敏感字段、scope 漂移或当前 actor 不一致均失败关闭。
- actor、current session 或 invalidation generation 变化会替换整个 security scope；请求使用 `generation + AbortController` 丢弃迟到响应，mutation 成功和跨标签 metadata-only `session_changed` 会统一失效旧 cursor、selection 与 confirmation。
- revoke others 和 credential rotation 只在已加载完整 directory 时展示 exact target set，不用客户端单项循环冒充服务端原子操作。password 与 confirmation 只留在组件内存，进入危险复核前就从可见 input state 清空，取消、成功、失败、scope 切换、路由离开和卸载均会清理；不进入 URL、Storage、日志或跨标签 payload。
- Desktop / Narrow 使用 Family UI 语义 token 和连续 Workbench；`900px` 以下改为单列，`620px` 以下将操作与危险确认改为纵向布局。实现对齐已批准的 `pOLcz` / `LMi7H` / `n2O8A5`，但本批不冒充批次 E 的真实三视口浏览器审计。
- Web strict consumer / state 测试与全量 `398/398` 通过，production build 成功；Platform `internal/httpapi` 普通测试和 race 也已通过，确认前端接入没有放宽既有 HTTP 安全边界。仓库 fast / full 门禁均通过。

### 批次 E：双数据库产品连续链与收口

- SQLite 完成双标签、多个 local / OIDC session、单项撤销、revoke others、密码轮换、旧密码失败、新密码登录、服务重启恢复与隐私审计。
- PostgreSQL configured Server 完成同构链、停止 no-fallback、重连恢复、受限角色和 migration 证据。
- 完成 `1440×900`、`720×900`、`390×844`、console / network / URL / storage / cookie 审计，并回写真相源后关闭专题。

## 验收方式

- memory、SQLite、PostgreSQL 的 session scope、稳定 cursor、effective state snapshot、CAS、并发和原子多 session revoke。
- credential replacement 与 source-bound session revoke 要么全部成功，要么全部不发生；旧密码、旧 credential 与已撤销 session 不能继续认证。
- 当前 local-password session 成功轮换后必须失效并清理 cookie；OIDC current session 不被本地密码轮换隐式撤销。
- session / account / credential / repository 任一失败在业务 repository、Gateway、Provider、Tool 或其它副作用前结束。
- Web 的密码清理、selection invalidation、迟到响应拒绝、跨标签同步、三视口和敏感材料扫描通过。

## 停止线

- 不实现账户禁用、删除、管理员代重置、邀请、全局账户或 session 搜索。
- 不实现 MFA、账户恢复、reset token、refresh token、密码历史服务、泄漏密码在线查询或生产速率限制。
- 不存储 IP、原始 User-Agent、设备指纹、地理位置或风险评分，不把 metadata-only session row 冒充安全设备识别。
- 不撤销或持久化上游 OIDC token，不把本地 session revoke 写成 Radish 全局登出。
- 不打开 production auth、production session store、production secret、真实 Radish 批次 E 或 production IAM。
- 不新增平行 identity / credential / session owner，不从本专题派生 S11 或普通只读 console。

## 下一实现入口

[本地账户凭证轮换与自助会话治理 v1 高风险任务卡](../../task-cards/local-account-credential-rotation-self-service-session-governance-dev-test-v1-plan.md)承接批次 A 至 E。批次 A 至 D 已完成；下一实现入口是批次 E 的 SQLite / PostgreSQL configured Server 产品连续链、三视口与隐私审计。本次已停在批次 E 之前，不打开 production auth 或任何已有停止线外的能力。
