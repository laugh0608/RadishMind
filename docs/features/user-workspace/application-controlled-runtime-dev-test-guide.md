# 应用受控运行开发测试态指南

更新时间：2026-07-20

## 适用范围

本文说明如何在 `apps/radishmind-web/` 中使用 Application RAG、Workflow Definition、Application Interaction Session、Run History 与 Application Operations 的开发测试态连续链。

这是一份使用与排障说明，不是生产部署手册。所有能力都要求显式 dev/test gate、可信 application scope、对应作用域和可用 repository；默认离线产品 UI 不发出这些请求。

## 两种会话运行 profile

| Profile | 运行权威 | 运行记录 | 主要用途 |
| --- | --- | --- | --- |
| `workflow_definition_executor_v1` | active definition pointer、immutable definition version / digest、application lifecycle、profile eligibility | `workflow_run_record.v5` | 执行已人工审查并激活的不可变 Workflow Definition |
| `application_rag_invocation_v1` | current runtime assignment、approved publish candidate v2、exact RAG binding、binding eligibility、application lifecycle | `workflow_run_record.v4` | 调用已人工激活的 Application RAG runtime |

Application Session 只是这两条既有运行路径之上的编排资源。它不复制 executor v0 的 DAG 算法，不把 RAG 伪装成 workflow draft，也不创建新的运行记录版本。

## 启动方式

共享 SQLite 本地产品链：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-session-local-product
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -ApplicationSessionLocalProduct
```

PostgreSQL 开发测试态连续链：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-session-postgres-dev-test
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -ApplicationSessionPostgresDevTest
```

只审查 Workflow Definition 晋级与 v5 运行时，可使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --workflow-definition-local-product
```

```powershell
pwsh ./scripts/run-radishmind-web-dev.ps1 -Mode dev-live -WorkflowDefinitionLocalProduct
```

PostgreSQL 对应参数为 `--workflow-definition-postgres-dev-test` / `-WorkflowDefinitionPostgresDevTest`。Application RAG 的完整本地产品链使用 `--workflow-rag-application-local-product` / `-WorkflowRAGApplicationLocalProduct`。

launcher 会组合必要的 Platform gate、Web source、workspace / application scope、共享 store 与 migration preflight。不要再手工组合不完整 selector；`configured` 档只用于显式 PostgreSQL组件组合和故障注入，缺失 marker、checksum、角色或连接时必须失败关闭。

## Application Development Workspace 使用方式

进入 User Workspace 并选中 Application 后，页面会以同一个 Application context 组织五个开发阶段。Application 仍通过 Applications 入口创建和选择，工作区不复制目录记录，也不创建另一份应用真相源。

| 阶段 | 主要工作 | 关键边界 |
| --- | --- | --- |
| Configure / Build | 配置草案、Workflow Draft、知识绑定等构建工作 | 保存与校验仍由各资源 owner 负责 |
| Human Promotion | 发布候选、Workflow Definition 和 RAG 晋级审查 | 审查、activation 与 assignment 都要求人工显式动作 |
| Controlled Test | API key、Application API、RAG 调用与 Application Session | 只允许显式开发测试 source，不把 UI ready 当作执行授权 |
| Run / Evaluation Review | Run History、Comparison、Evaluation 与 Application Operations | 只读消费持久化 metadata，不重新调用 provider |
| Release Readiness | 汇总 Application、配置、Workflow、RAG、受控测试、评测和运维证据 | 只做开发测试态证据投影，不提供发布或 production ready 动作 |

每次只挂载当前阶段的功能 surface。切换 Application、Application revision、lifecycle 或阶段时，页面会创建新的 workspace / route generation 和 `surfaceKey`；旧 surface 的迟到请求结果会被拒绝，不能覆盖当前选择或证据。未显示的阶段不会作为隐藏面板继续发请求。

跨阶段跳转只传递当前 Application generation 内的稳定短引用。例如从配置草案前往发布审查，或从一次受控运行前往历史 / 评测时，目标 owner 会按精确 draft、candidate、definition、binding、assignment、session、run、request 或 evaluation 引用重新读取数据。handoff 不携带来源完整对象，也不自动保存、审查、激活、分配、调用 provider 或发布；切换 Application 后，旧 handoff 不再有效。

Release Readiness 固定使用四种状态：

- `review_not_started`：尚未形成可供汇总的 owner evidence。
- `review_incomplete`：已有部分 evidence，但来源不全或无法证明当前 revision。
- `review_blocked`：存在 lifecycle、authority drift、owner failure 或其它明确阻塞。
- `dev_test_evidence_reviewable`：当前开发测试态 evidence 已可人工复核；不等于可发布或 production ready。

`active` Application 可使用各阶段允许的开发测试动作；`archived` Application 的 Controlled Test 被阻塞，其余阶段仅保留只读审查；Application 不可用时全部阶段阻塞。默认离线来源缺少权威 revision 时，页面保留可浏览的稳定引用并显示 `incomplete / partial`，不会伪造 revision，也不会因此让整个页面崩溃。

稳定 URL 只记录当前阶段，不保存 Application payload、资源完整对象、input、answer、token、review reason 或一次性凭据。刷新或重新选择 Application 后，应以各 owner 的权威读取结果为准，不依赖浏览器恢复易失状态。

### 常见问题定位

- readiness 长期为 `review_incomplete`：逐项检查对应 owner 是否加载成功、是否属于当前 Application revision，以及精确引用是否仍存在；不要用其它 draft 或 run 自动补位。
- readiness 为 `review_blocked`：先处理 Application archived、authority drift、作用域不足、repository / migration 失败或 owner `partial_failure`，再刷新当前阶段。
- handoff 后找不到目标资源：返回来源阶段确认资源是否仍属于当前 Application generation；目标 owner 不允许退回最近一次 draft、candidate 或 run。
- 切换阶段后出现旧数据：记录当前 `surfaceKey` 并检查迟到回调是否经过 workspace controls；不要在 panel 外维护第二份可写选择状态。
- 离线模式显示证据不完整：这是缺少权威 Application revision 时的预期行为；需要连续链时改用本节启动参数对应的显式开发测试 source。

## Workflow Definition 准备与运行

1. 在 Application Catalog 中选择一个 `active` application。
2. 创建并保存满足 executor v0 拓扑约束的 Workflow Draft；执行时只支持既有 `prompt`、`llm`、`condition`、`output` 图算法。
3. 从精确 saved draft 创建 immutable definition candidate。
4. 人工提交 approve / reject review decision。approve 不自动激活。
5. 从 approved candidate 生成 immutable definition version。
6. 人工执行 activation CAS，建立当前 active definition pointer。
7. 从 Definition Promotion 面板发起 v5 运行，或在 Application Interaction Session 中选择 `workflow_definition_executor_v1` 后提交 turn。

创建 session 时，Workflow Definition profile 需要明确的 `definition_id`。每轮执行前，服务端重新读取 activation pointer、definition version、definition digest、application lifecycle 与 profile eligibility；Web 中显示的 ready 状态只是解释信息，不是执行授权。任一 authority 漂移都必须在 provider 调用前失败，不能回退 source draft 或旧 activation。

Workflow Definition 的主要路由族为：

- `/v1/user-workspace/workflow-definition-candidates*`
- `/v1/user-workspace/workflow-definitions/{definition_id}/versions*`
- `/v1/user-workspace/workflow-definitions/{definition_id}/activation`
- `/v1/user-workspace/workflow-definitions/{definition_id}/activation-decisions`
- `POST /v1/user-workspace/workflow-definition-runs`

管理动作使用 `workflow_definitions:write`、`workflow_definitions:read`、`workflow_definitions:review`、`workflow_definitions:activate`；运行还要求 `workflow_runs:execute`。缺少任一作用域都应保留稳定失败语义，不进入 provider。

## Application RAG 准备与调用

Application RAG 复用既有知识治理链，操作顺序固定为：

1. 准备 immutable knowledge snapshot、evaluation dataset 与 candidate review。
2. 人工批准 promotion candidate，并在 Application Configuration Draft v2 中显式 attach exact binding。
3. 创建并人工批准 Application Publish Candidate v2；approve 不自动激活 runtime。
4. 在 Application RAG Runtime Assignment 中人工执行 `activate` 或 `replace` CAS。
5. 使用具备 `application_rag:invoke` 的当前 application API key 调用，或在 Application Interaction Session 中选择 `application_rag_invocation_v1` 后提交 turn。

主要路由为：

- `GET /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment`
- `POST /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment/decisions`
- `POST /v1/application-rag/invocations`

服务端在每次调用前重读 current assignment、approved publish candidate、exact binding、binding eligibility 与 application lifecycle。客户端不能提交或覆盖 authority 摘要，也不能用已签发 API key 绕过 application archived、assignment revoked 或 binding drift。

完整 snapshot、dataset、promotion、binding 与发布重校验步骤见 [Workflow RAG 开发测试态使用与资源治理指南](../workflow/workflow-rag-dev-test-usage-governance-guide.md)。API key 的签发、一次性交接与吊销见 [应用目录与 API 密钥开发测试指南](application-catalog-api-key-dev-test-guide.md)。

## Application Interaction Session

Session 路由族为：

- `POST /v1/user-workspace/application-sessions`
- `GET /v1/user-workspace/application-sessions`
- `GET /v1/user-workspace/application-sessions/{session_id}`
- `POST /v1/user-workspace/application-sessions/{session_id}/close`
- `GET /v1/user-workspace/application-sessions/{session_id}/turns`
- `POST /v1/user-workspace/application-sessions/{session_id}/turns`

读取、管理和执行分别要求 `application_sessions:read`、`application_sessions:write`、`application_sessions:execute`。Session 只能从 `active` 转为 `closed`；closed session 可以查看 metadata，但不能继续创建 turn。

Web 使用顺序：

1. 选择 application 和 profile；Workflow Definition profile 同时选择 definition。
2. 创建 session，确认服务端返回的 authority refs 与 profile。
3. 提交 turn。输入与同步 answer 只存在于当前交互视图内存。
4. 从 turn metadata 打开对应 v4 / v5 Run Detail、Comparison 或 Evaluation。
5. 不再继续时显式关闭 session；通过 Active / Closed 过滤器区分可执行会话与历史会话。

切换 workspace、application、profile 或路由时，Web 会中止活动请求，并清除当前 input、answer、transcript、一次性 credential 与冲突状态。刷新页面或重启服务后只恢复 session / turn metadata 和 run refs，不重建 transcript。

## 幂等、取消与不确定结果

- `client_turn_key` 在 session owner 内提供幂等。相同 key 与相同请求只返回既有 turn；冲突载荷返回 `idempotency_conflict`，不能再次调用 provider。
- 用户取消会沿用既有 v5 / v4 的取消语义。取消不是 replay 或 resume，也不自动创建替代 turn。
- provider 已返回但终态持久化不确定时，记录只能进入明确的不确定状态；不得为“补结果”自动重放 provider。
- stale reconciliation 只把长期非终态记录收敛为 `outcome_unknown` 等 metadata-only 结果，不恢复答案、不重放执行。
- `version_conflict`、authority drift、session closed、application archived、profile ineligible、scope denied、store contract mismatch 与 migration failure 都必须先刷新权威状态再处理，不得回退 memory 或离线样例。

## History、Comparison、Evaluation 与 Operations

Run History、Detail、Comparison、Evaluation、Baseline 与 Suite 对 v4 / v5 都是只读消费者：

- v4 使用明确的 Application RAG execution profile 与 assignment / binding refs。
- v5 使用 `workflow_definition_executor.v1` evaluation profile 与 definition / activation refs。
- 消费端只比较已持久化的 metadata、diagnostics、trace、usage availability 和 lineage，不重新调用 provider。

Application Operations 同时读取当前 application 的 Gateway Request History 与 Workflow Run History，但保留两个来源各自的加载、空结果与失败状态。合并时间线只按已有时间字段排序，不推测 request 与 run 的相关关系，不补算缺失 token、成本、quota 或 billing。

## 持久化与隐私边界

memory、SQLite 与 PostgreSQL 只持久化作用域、资源引用、版本 / CAS、digest、状态、时间、trace / usage availability 和 diagnostics 等 metadata。以下内容不得进入 Session、Turn、Run History、Comparison、Evaluation、Operations、日志、fixture 或公开错误：

- 原始 input、answer 或 transcript
- prompt、provider raw response 或 fragment 正文
- Authorization、API key、credential、token、cookie 或 header
- provider secret、DSN 或异常原文

SQLite migration 中 Application RAG、Workflow Definition release、definition execution 与 Application Session 依次为 `0009`、`0010`、`0011`、`0012`；PostgreSQL 对应为 `0012`、`0013`、`0014`、`0015`。运行角色只授予必要 DML，migration role 与 runtime role 不得互换。

## 停止线

- 不自动 review、activation、publish、assignment 或 profile 选择。
- 不增加 schedule、retry / fallback、replay / resume、agent loop、长期记忆或后台任务。
- 不从 session transcript 派生持久记忆，不把 answer 写回上层业务真相源。
- 不把本地 SQLite、PostgreSQL dev/test、mock provider、真实浏览器验收或 launcher 连续链解释为 production ready。
- 不绕过 HTTP Tool、Workflow RAG、Application RAG 或 Workflow Definition 各自的 authority owner。
