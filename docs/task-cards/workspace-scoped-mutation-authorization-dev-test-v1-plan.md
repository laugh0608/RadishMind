# Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-27

状态：`workspace_scoped_mutation_authorization_dev_test_v1_batch_d_complete_batch_e_ready`

对应功能文档：[Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1](../features/user-workspace/workspace-scoped-mutation-authorization-dev-test-v1.md)

## 任务目标

在不创建第二套 identity、membership 或业务 owner 的前提下，让 User Workspace 中由人发起的写入、审查和执行动作复用唯一 `WorkspaceMembershipProvider`，并以 active workspace、operation permission、resource owner 和副作用顺序形成可复验的开发测试态授权边界。

本任务卡统一承载 A1 Application Catalog、A2 API Key Lifecycle 及后续既有 owner 迁移。每批必须在进入下一批前独立完成相邻测试、拒绝路径副作用证明和必要仓库门禁；不为每个 handler 派生平行任务卡、fixture 或 checker。

## 根因

- Workspace-scoped Read Transition 已有 verified identity、active workspace、membership assertion 和 durable read projection，mutation handler 尚未统一消费该 binding。
- Application Catalog 与 API Key Lifecycle 已完成 shared membership binding；后续 Workflow、Prompt、Agent、Session、Run 与 Evaluation 仍存在从 verified identity、body workspace 与旧 context builder 组合上下文的路径，body workspace 不是 membership proof。
- Workflow、Prompt、Agent、Session、Run 和 Evaluation 存在多种 dev header / body workspace 组合，若不先建立稳定共享入口，后续会继续复制授权判断和失败映射。
- API key create 还包含 credential 生成、hash、持久化和一次性交接，必须在资源根授权稳定后单独验证。

## 固定授权顺序

```text
dev/test enablement
  -> verified identity
  -> identity operation permission
  -> active workspace selection
  -> atomic workspace membership decision
  -> strict payload and resource coordinate match
  -> durable owner reload
  -> domain validation / CAS
  -> durable or external side effect
  -> sanitized response
```

- `X-RadishMind-Active-Workspace` 是人类 mutation 的唯一 workspace selector。
- body、path 和历史 dev workspace header 只做资源坐标一致性校验。
- identity permission 与 membership permission 必须同时满足。
- membership failure 允许 membership provider 完成自身判断，但业务 owner query、CAS、credential、Run、Gateway / provider 和网络副作用必须为 0。
- application API key invocation 不追加人类 membership assertion；其 workspace / application / capability 继续来自验证后的 key binding。

## 批次 A1：Application Catalog

### 范围

- `POST /v1/user-workspace/applications`
- `PUT /v1/user-workspace/applications/{application_id}`
- `POST /v1/user-workspace/applications/{application_id}/archive`
- 共享 `WorkspaceMembershipProvider` mutation permission registry 与 handler authorization 入口。

### 实现要求

- `applications:write` 与 `applications:archive` 同时进入 identity 和 membership allowlist。
- 三条 mutation 在 JSON body 解码和 Application Catalog repository 查询前完成 identity、active workspace 与 membership decision。
- body `workspace_id` 必须与 verified resource binding 中的 workspace 精确一致。
- Application Catalog context 中 tenant、actor、owner subject 和 workspace 只来自 verified binding，不接受 body 覆盖。
- 保持 Application Catalog memory / SQLite / PostgreSQL repository、record schema、CAS、soft archive、cursor 与下游归档只读语义不变。
- 旧读路由行为保持兼容；本批不借 mutation 迁移扩大 read route 范围。

### 负向矩阵

- identity 缺失、过期或结构无效；
- identity 缺少 operation permission；
- active workspace 缺失、重复或非法；
- membership 缺失、过期、tenant / subject / workspace mismatch；
- membership 缺少 operation permission；
- body workspace 与 active workspace 不一致；
- OIDC integration test membership unavailable；
- unknown field、重复字段、尾随 JSON 和 forbidden secret material；
- stale record version、archived update 与重复 archive。

### 完成条件

- 授权 failure 使用稳定 workspace failure code 和 HTTP status，不再全部压平为 Application Catalog `scope_denied`。
- identity、selection、membership 与 body binding failure 的业务 repository query / write 均为 0。
- memory、SQLite、PostgreSQL 既有 owner 行为不变；至少完成相邻 HTTP、repository 与现有 PostgreSQL 集成回归。
- 定向普通测试、race、完整 `internal/httpapi`、`go vet ./...`、fast 和 full 仓库门禁通过。

### A1 完成记录（2026-07-27）

- 三条 Application Catalog mutation 已在 body 解码前复用共享 workspace authorization；`applications:write` 与 `applications:archive` 同时进入 identity 和 membership permission 判断。
- active workspace 成为唯一 selector；body workspace 只与 verified resource binding 做精确一致性校验，tenant、actor、owner subject 与 workspace 均不接受 body 覆盖。
- identity、selection、membership 与 body binding 的稳定失败码保持原值，业务 repository spy 对全部拒绝路径记录为 0；旧 Application Catalog read route 未随 A1 扩大迁移。
- dev headers、signed-test、过期 signed identity、OIDC unavailable、memory / SQLite / PostgreSQL、定向 race、完整 Platform HTTP、`go vet`、Web 245 项测试和 production build 均通过。
- A1 没有修改 repository interface、record schema、migration、CAS、soft archive、cursor、application API key invocation 或 production enablement；下一步只进入 A2。

## 批次 A2：API Key Lifecycle

### 范围

- `POST /v1/user-workspace/api-keys`
- `POST /v1/user-workspace/api-keys/{api_key_id}/revoke`
- 三类 application invocation 与 northbound Gateway API key auth 只做回归。

### 实现要求

- `api_keys:write` 与 `api_keys:revoke` 同时进入 identity 和 membership allowlist。
- API Key owner 与 Application Catalog owner 只在 membership 成功后查询。
- token 随机数、hash、record write 与一次性交接严格位于完整授权和 active Application 校验之后。
- create 成功 response 继续 `Cache-Control: no-store`；failure、日志、history、数据库和 Web 持久介质不得出现 token。
- revoke 保持 expected-version CAS，不修改历史 request / Run。

### 完成条件

- 授权与 workspace binding failure 的 catalog query、API key query / write、credential generator 和 Gateway bridge 调用均为 0。
- memory、SQLite、PostgreSQL 连续链、重启恢复、吊销生效和敏感信息扫描通过。
- A2 完成前不推进 Workflow / Prompt / Agent / Session / Evaluation。

### A2 完成记录（2026-07-27）

- API key create / revoke 已在 body 解码前复用共享 workspace authorization；`api_keys:write` 与 `api_keys:revoke` 同时进入 identity 和 membership permission 判断，body workspace 只做精确一致性校验。
- 双操作拒绝矩阵证明 identity、selection、membership、body binding 与 OIDC unavailable 均在 Application Catalog / API Key repository 前失败，两个业务 owner 的调用数均为 0。
- 相邻 service spy 固定 Application owner reload → identifier generation → credential generation / hash → record write；inactive、missing 与 cross-owner Application 继续证明 credential generator 为 0，失败响应不交付 token。
- dev headers、signed-test、Web active workspace / dev membership 分离、memory / SQLite / PostgreSQL、吊销、重启、敏感扫描、完整 Platform HTTP、定向 race、`go vet`、Web 246 项测试和 production build 均通过。
- A2 没有修改 read route、application invocation、Gateway API key auth、repository interface、record schema、migration、expected-version CAS 或 production enablement；下一步只复核批次 B。

## 批次 B：创作 owner

### 范围

- Workflow Draft validate / save；
- Application Configuration Draft validate / save、Prompt Template binding、Agent Profile binding；
- Prompt Template validate / save / version；
- Agent Profile validate / save / version。

### 固定决策

- 单权限与组合权限统一进入 `authorizeWorkspaceScopedPermissions`；组合权限不得循环调用 membership provider。
- 12 条 mutation 均在 body 解码前完成授权；body workspace 只做 verified active workspace 精确一致性校验。
- Configuration Draft 可选 binding 能力由 verified identity 与 membership grants 交集派生；显式 binding 路由仍必须原子要求基础写权限与对应 bind 权限。
- 历史 owner dev workspace / application header 不再建立 mutation authority；既有 read route 暂不随本批迁移。

### 完成记录（2026-07-27）

- permission registry、signed-test projection 与共享 provider 已支持本批单项 / 组合权限；Prompt / Agent binding 的第二权限缺失分别在 identity 或单次 membership decision 阶段失败。
- 四类 save 拒绝矩阵证明 identity、selection、membership、body binding 与 OIDC unavailable 均在业务 owner 前失败，owner 调用为 0；双权限测试同时固定 provider 调用次数。
- 既有正向 HTTP 覆盖 validate、save、immutable version、Prompt binding 与 Agent binding；Web 四类 consumer 已发送 active workspace 和最小 dev membership permission。
- memory、SQLite、PostgreSQL 与 owner CAS / immutable / ref-only 语义保持不变；完整 Go 与 Web、PostgreSQL integration suite 已通过。
- 本批没有进入 Publish Candidate、Definition Candidate、RAG Promotion、Runtime Assignment、Session、Run、Evaluation 或 application API key invocation。

## 批次 C：审查与激活 owner

### 范围

- Application Publish Candidate create / review；
- Workflow Definition Candidate create / review 与 activation；
- Workflow RAG Promotion Candidate create / review；
- Workflow RAG、Prompt Application、Agent Copilot Runtime Assignment decision。

### 固定决策

- 本批共 10 条 mutation。可预先确定的完整权限集合必须在 body 解码与业务 owner 前由一次 membership decision 原子验证。
- RAG Promotion create 固定要求四项 permission；review、definition、activation 与三类 assignment 使用 inventory 中各自单项 permission。
- Application Publish Candidate create 先要求 `application_publish_candidates:write`，review 先要求 `application_publish_candidates:review`。RAG / Prompt / Agent source permission 由 create 最小重读的权威 Application Draft 或 `approve` 最小重读的 Publish Candidate 类型决定，并只从同一 verified identity + membership grants 交集派生；provider 不得再次调用，也不得把三项条件权限全部强加给调用者。非 `approve` review 不追加 source permission。
- 条件 source permission 缺失时，只允许 Application Draft 或 Publish Candidate 最小只读；Publish Candidate write、Application Catalog 与对应 source owner 查询必须为 0。其它授权拒绝路径要求全部业务 owner 为 0。
- 既有 candidate / review / activation / assignment repository、schema、migration、CAS、event / audit 与 API key invocation 不改变；approval 不自动 activation，assignment 不自动 invocation。

### 完成记录（2026-07-27）

- 10 条审查 / 激活 mutation 已接入共享 authorization；可预先确定的 permission 在 body 前由一次 membership decision 原子验证，RAG Promotion create 的四项 permission 不拆分为多次 provider 调用。
- 跨 10 条入口的拒绝矩阵证明 identity、selection、membership、body binding 与 OIDC unavailable 均在 primary owner 前失败；Application Publish 条件 source permission 缺失只发生一次最小 Draft / Candidate 重读，不查询 baseline、source owner 或 durable write。
- 上游 permission projection 与 Web 六类 mutation consumer 均已补齐；完整 Platform Go tests、定向 race、`go vet`、Web 246 项测试 / production build 与 PostgreSQL integration suite 通过。
- repository interface、schema、migration、CAS、append-only event / audit、approval / activation 分离和三类 application API key invocation 均未改变。
- 批次 C 已关闭，累计 27 条 mutation 完成迁移；下一步只复核批次 D。

## 批次 D：Session / Turn 与人类受控执行

### 范围

- Application Interaction Session create / close；
- Application Interaction Turn execute；
- Saved Workflow Draft run；
- Workflow Definition-bound run。

### 固定决策

- 本批共 5 条 mutation。Session write / execute 使用各自单项 permission；Saved Draft run 原子要求 `workflow_runs:execute` + `workflow_drafts:read`；Definition-bound run 原子要求 `workflow_runs:execute` + `workflow_definitions:read`。所有 permission 均在 body 前由一次 membership decision 判断。
- Session create / close 的授权与 binding failure 必须保持 Session、Application、Definition / Runtime authority owner 和外部调用为 0；Session 管理动作不创建 Run、不调用 Gateway / provider。
- Turn execute 的授权与 binding failure 必须发生在 Session read、Turn reservation、Run write 和任何 delegate 前。授权通过后仍只按 persisted Session authority 至多委托一次，不接受客户端 credential、API key 或 authority override。
- 两类 Run 的授权与 binding failure 必须发生在 Draft / Definition / Activation / Application read、planned Run write 和 Gateway / provider 前；授权通过后继续复用既有 planned → external call → terminal 顺序与失败恢复。
- repository interface、schema、migration、Session / Turn / Run CAS、idempotent replay、metadata-only persistence 和三类 application API key invocation 均不改变。

### 完成记录（2026-07-27）

- 5 条 Session / Turn / Run mutation 已接入共享 authorization；Session write / execute 与两类双权限 Run 均在 body 前由一次 membership decision 判断。
- 跨 5 条入口的拒绝矩阵证明 identity、selection、membership、body binding 与 OIDC unavailable 均在业务 owner、Turn reservation、Run write 和 Gateway / provider 前失败；双权限缺失测试同时固定 provider 0 / 1 次调用边界。
- Session 管理动作保持零 Run 与零外部调用，Turn 继续只按 persisted authority 至多委托一次，两类 Run 保持 planned → external call → terminal 顺序；客户端不能覆盖 runtime credential 或 authority。
- 上游 permission projection 与五类 Web consumer 已补齐；完整 Platform Go tests、定向 race、`go vet`、Web 246 项测试 / production build 与 PostgreSQL integration suite 通过。
- repository interface、schema、migration、Session / Turn / Run CAS、idempotent replay、metadata-only persistence 与 application API key invocation 均未改变。
- 批次 D 已关闭，累计 32 条 mutation 完成迁移；下一步只复核批次 E。

## 后续批次

1. 批次 C 已完成：Publish Candidate、Definition Candidate、RAG Promotion 与 Runtime Assignment。
2. 批次 D 已完成：Workflow Run、Application Session / Turn 与人类受控执行。
3. 下一批次 E：RAG Dataset / Snapshot、HTTP Tool 与 Evaluation。

每批开始前必须从功能文档 inventory 重读真实 permission、workspace 来源、owner 和副作用。若出现公开 API、schema、幂等协议、provider 或网络边界变化，先更新功能文档与本任务卡。

## 验证策略

- 授权核心：`workspace_membership_test.go` 与新增 mutation authorization 相邻测试。
- HTTP：每批 handler 正向、拒绝矩阵、zero-query / zero-side-effect spy 与条件 owner 重读顺序。
- 持久化：复用既有 memory、SQLite、PostgreSQL repository contract 与产品集成门禁。
- 并发：对 permission decision、关键 CAS 与当批 owner 使用定向 race。
- 仓库级：每个批次完成运行 `./scripts/check-repo.sh --fast`；A1、A2 和专题关闭运行全量 `./scripts/check-repo.sh`。

## 总停止线

- 不创建 mutation 专用 membership provider、用户表、tenant 表、role 表或聚合业务 owner。
- 不把 active workspace、body workspace、旧 dev header、resource record 或 API key 当作人类 membership proof。
- 不修改 application API key invocation 的逐请求授权模型。
- 不启用 production OIDC、production membership、quota / billing、自动发布、自动确认、replay、unrestricted tool 或业务写回。
- 不从已完成的批次 A 至 D 直接复制到 Evaluation、RAG Dataset / Snapshot 或 HTTP Tool runtime；批次 E 必须先复核组合权限、执行 credential、owner 重读和 provider / Gateway / network 副作用。
