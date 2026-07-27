# Workspace-scoped Mutation Authorization / 工作区写入与审查动作成员资格绑定（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-27

状态：`workspace_scoped_mutation_authorization_dev_test_v1_batch_a_complete_batch_b_ready`

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

## 后续批次

1. 批次 B：Workflow Draft、Application Configuration Draft、Prompt Template 与 Agent Profile。
2. 批次 C：Publish Candidate、Definition Candidate、RAG Promotion 与 Runtime Assignment。
3. 批次 D：Workflow Run、Application Session / Turn 与人类受控执行。
4. 批次 E：RAG Dataset / Snapshot、HTTP Tool 与 Evaluation。

每批开始前必须从功能文档 inventory 重读真实 permission、workspace 来源、owner 和副作用。若出现公开 API、schema、幂等协议、provider 或网络边界变化，先更新功能文档与本任务卡。

## 验证策略

- 授权核心：`workspace_membership_test.go` 与新增 mutation authorization 相邻测试。
- HTTP：Application Catalog / API Key handler 正向、负向、zero-query / zero-side-effect spy。
- 持久化：复用既有 memory、SQLite、PostgreSQL repository contract 与产品集成门禁。
- 并发：对 permission decision、Application CAS 和 API key create / revoke 使用定向 race。
- 仓库级：每个批次完成运行 `./scripts/check-repo.sh --fast`；A1、A2 和专题关闭运行全量 `./scripts/check-repo.sh`。

## 总停止线

- 不创建 mutation 专用 membership provider、用户表、tenant 表、role 表或聚合业务 owner。
- 不把 active workspace、body workspace、旧 dev header、resource record 或 API key 当作人类 membership proof。
- 不修改 application API key invocation 的逐请求授权模型。
- 不启用 production OIDC、production membership、quota / billing、自动发布、自动确认、replay、unrestricted tool 或业务写回。
- 不从已完成的 A1 / A2 直接扩审查 / 激活、Session、Run 或 Evaluation runtime；批次 B 先复核四类创作 owner 的真实权限与 context。
