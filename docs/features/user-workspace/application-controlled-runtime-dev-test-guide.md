# 应用受控运行开发测试态指南

更新时间：2026-08-17

## 适用范围

本文说明如何在 `apps/radishmind-web/` 中使用 Application RAG、Workflow Definition、Application Interaction Session、应用结果资产库、Run History 与 Application Operations 的开发测试态连续链。

这是一份使用与排障说明，不是生产部署手册。所有能力都要求显式 dev/test gate、可信 application scope、对应作用域和可用 repository；默认离线产品 UI 不发出这些请求。

## 三种会话运行 profile

| Profile | 运行权威 | 运行记录 | 主要用途 |
| --- | --- | --- | --- |
| `workflow_definition_executor_v1` | active definition pointer、immutable definition version / digest、application lifecycle、profile eligibility | `workflow_run_record.v5` | 执行已人工审查并激活的不可变 Workflow Definition |
| `workflow_definition_executor_v2` | active definition pointer、immutable Definition v2、exact input contract id / digest、application lifecycle、profile eligibility | `workflow_run_record.v8` | 使用有界 typed inputs 执行已激活的结构化 Workflow Definition |
| `application_rag_invocation_v1` | current runtime assignment、approved publish candidate v2、exact RAG binding、binding eligibility、application lifecycle | `workflow_run_record.v4` | 调用已人工激活的 Application RAG runtime |

Application Session 只是这些既有运行路径之上的编排资源。v1 / RAG profile 继续使用 Session / Turn v1，结构化 Definition profile 使用 Session / Turn v4；它不复制 executor v0 的 DAG 算法，不把 RAG 伪装成 workflow draft，也不创建平行运行记录版本。

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

应用结果资产库的稳定 SQLite 产品复验使用：

```bash
./scripts/run-radishmind-web-dev.sh --mode dev-live --application-result-artifact-library-local-product
```

该 Shell 专用入口会继承完整 Application Session SQLite 产品档，并显式加入幂等双 Session fixture。fixture 只用于开发测试态 application-scoped list / filter / exact read / export / lifecycle / restart 复验，不建立真实 Provider、Session 或 Run 执行事实，也不会在重启时覆盖已推进的 lifecycle。PowerShell wrapper 当前没有对称 fixture 参数；常规 Application Session 链继续使用上方双端入口。

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
2. 创建并保存满足 executor v0 拓扑约束的 Workflow Draft；v1 使用单文本输入，v2 还必须声明最多 `16` 个扁平 `string | integer | number | boolean` 字段的有界 input contract。
3. 从精确 saved draft 创建 immutable definition candidate。
4. 人工提交 approve / reject review decision。approve 不自动激活。
5. 从 approved candidate 生成 immutable definition version。
6. 人工执行 activation CAS，建立当前 active definition pointer。
7. v1 从 Definition Promotion 面板发起 v5 运行，或在 Session 中选择 `workflow_definition_executor_v1`；v2 使用共享 typed editor 发起 v8 运行，或选择 `workflow_definition_executor_v2` 后提交 turn。

创建 session 时，Workflow Definition profile 需要明确的 `definition_id`。每轮执行前，服务端重新读取 activation pointer、definition version、definition digest、application lifecycle 与 profile eligibility；v2 还重新核对 input contract id / digest、字段类型、required、预算和 secret。Web 中显示的 ready 状态只是解释信息，不是执行授权。任一 authority 或 contract 漂移都必须在 provider 和 turn 持久化前失败，不能回退 source draft、旧 activation 或 v1 input。

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
- `GET /v1/user-workspace/application-sessions/{session_id}/result-artifacts`
- `GET /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}`
- `POST /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}/archive`
- `POST /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}/unarchive`
- `GET /v1/user-workspace/applications/{application_id}/result-artifacts`
- `GET /v1/user-workspace/applications/{application_id}/result-artifacts/{artifact_id}/export`

读取、管理和执行分别要求 `application_sessions:read`、`application_sessions:write`、`application_sessions:execute`。Session 只能从 `active` 转为 `closed`；closed session 可以查看 metadata，但不能继续创建 turn。

当前 HTTP 使用顺序：

1. 选择 application 和 profile；Workflow Definition profile 同时选择 definition。
2. 创建 session，确认服务端返回的 authority refs 与 profile。
3. 提交 turn。`save_result` 默认 false，输入与同步 answer 只存在于当前交互视图内存；只有显式设为 true，服务端才从本次成功执行的 canonical result 创建独立结果资产。
4. 保存成功时，turn 响应返回 metadata-only `result_artifact`；保存失败使用独立 `result_artifact_failure_code`，不改变已经成功的 run，也不重放 Provider。
5. 在同一 application / session 下列出结果资产 summary；默认只返回 active，可显式指定 `lifecycle_state=archived`。再以精确 `artifact_id` 读取正文；列表和 lifecycle event 不返回 content，精确读取响应使用 `Cache-Control: no-store`。
6. archive / unarchive 使用独立 `application_result_artifacts:archive` 权限和 `expected_lifecycle_version` CAS；陈旧版本或重复状态返回冲突，不修改 artifact content、digest、来源与 created at。永久 purge route 尚未开放。
7. 从 turn metadata 打开对应 v4 / v5 / v8 Run Detail、Comparison 或 Evaluation；v8 只展示合同与字段 metadata，不回显字段值。
8. 不再继续时显式关闭 session；通过 Active / Closed 过滤器区分可执行会话与历史会话。

当前 React Application Interaction、Prompt Session 与 Agent Session 已复用单一 strict artifact consumer 和共享结果资产面板。逐 turn 保存选择默认关闭；显式开启后只发送 `save_result=true`，并分别展示执行结果与保存状态。页面可读取 active / archived metadata 列表、精确正文与来源 Run，并以独立 archive 权限和 expected-version CAS 执行 archive / unarchive；列表、URL 与浏览器持久存储均不包含正文。三类 Session 不复制 schema 解析、scope guard 或迟到响应状态。

Application Runtime Review 的 `Saved results` 任务复用同一 artifact / lifecycle owner，以当前 application 和 owner 跨 Session 列出 metadata，可按 lifecycle、execution profile 与 content type 过滤。export 额外要求 `application_result_artifacts:export`；服务端重新读取 exact artifact / lifecycle 并生成 `application_result_artifact_export.v1`，Web 复核 content / export digest 后才准备一个易失 Blob URL，再由用户显式下载。export 不落库、不改变 lifecycle、不形成下载历史或公开分享；切换 application、筛选、选择或卸载页面都会撤销已准备 URL。

切换 workspace、application、profile、session、identity 或路由时，Web 会中止活动请求，并清除当前 input、answer、transcript、已读取 artifact content、一次性 credential 与冲突状态。刷新页面或重启服务后只恢复 session / turn metadata、run refs 和 SQLite / PostgreSQL 中显式保存的 artifact，不重建 transcript。结果资产使用独立 owner：memory 模式只在同一服务进程内可精确读取；SQLite / PostgreSQL 开发测试态可在服务重启后恢复显式保存的 artifact，但不得把该结论扩写为 transcript 恢复或 production durable 能力。

## 幂等、取消与不确定结果

- `client_turn_key` 在 session owner 内提供幂等。相同 key 与相同请求只返回既有 turn；冲突载荷返回 `idempotency_conflict`，不能再次调用 provider。
- 相同 key 的首次请求若已保存结果，重试只返回既有 artifact summary，不回传首次易失结果、不再次调用 Provider；首次没有保存时，重试不能事后补建已经丢失的内容。
- 用户取消会沿用既有 v4 / v5 / v8 的取消语义。取消不是 replay 或 resume，也不自动创建替代 turn。
- provider 已返回但终态持久化不确定时，记录只能进入明确的不确定状态；不得为“补结果”自动重放 provider。
- stale reconciliation 只把长期非终态记录收敛为 `outcome_unknown` 等 metadata-only 结果，不恢复答案、不重放执行。
- `version_conflict`、authority drift、session closed、application archived、profile ineligible、scope denied、store contract mismatch 与 migration failure 都必须先刷新权威状态再处理，不得回退 memory 或离线样例。

## History、Comparison、Evaluation 与 Operations

Run History、Detail、Comparison、Evaluation、Baseline 与 Suite 对 v4 / v5 / v8 都是只读消费者：

- v4 使用明确的 Application RAG execution profile 与 assignment / binding refs。
- v5 使用 `workflow_definition_executor.v1` evaluation profile 与 definition / activation refs。
- v8 使用 `workflow_definition_executor.v2`、exact input contract 和字段 metadata；Comparison v7 只接受同 Definition lineage、同 contract id / digest 的终态 v8 Run。
- 消费端只比较已持久化的 metadata、diagnostics、trace、usage availability 和 lineage，不重新调用 provider。

Application Operations 同时读取当前 application 的 Gateway Request History 与 Workflow Run History，但保留两个来源各自的加载、空结果与失败状态。合并时间线只按已有时间字段排序，不推测 request 与 run 的相关关系，不补算缺失 token、成本、quota 或 billing。

## 持久化与隐私边界

memory、SQLite 与 PostgreSQL 的 Session、Turn、Run、Comparison、Case、Suite 和 Operations 只持久化作用域、资源引用、版本 / CAS、digest、字段名 / 类型 / bytes、状态、时间、trace / usage availability 和 diagnostics 等 metadata。Application Result Artifact 是独立、显式 opt-in 的内容 owner，不改变上述契约；批次 A 至 D 固定单份正文上限 `64 KiB`，session list 和 lifecycle event 只返回 metadata，共享 Web consumer 也只在精确 read 后把正文保留于当前组件内存，SQLite / PostgreSQL 仅由精确 artifact read 恢复正文。Application Evaluation Plan v2 只持有用户显式保存并通过 secret / contract 校验的不可变 typed fixture，不从 Session 或 Run 反推输入。以下运行时材料不得进入 Session、Turn、Run History、Comparison、Case、Suite、Operations、日志或公开错误：

- 原始 input、answer 或 transcript
- prompt、provider raw response 或 fragment 正文
- Authorization、API key、credential、token、cookie 或 header
- provider secret、DSN 或异常原文

SQLite 中 Application RAG、Workflow Definition release、definition execution 与 Application Session 基线依次为 `0009`、`0010`、`0011`、`0012`，结构化 Definition、Session 与 Evaluation 扩展为 `0017`、`0018`、`0019`，Definition HTTP Tool 来源 / 执行、结果资产、结果资产生命周期与 application history 索引依次为 `0020`、`0021`、`0022`、`0023`、`0024`；PostgreSQL 基线对应 `0012`、`0013`、`0014`、`0015`，结构化扩展对应 `0020`、`0021`、`0022`，Definition HTTP Tool 来源 / 执行、结果资产、生命周期与 application history 索引对应 `0023`、`0024`、`0025`、`0026`、`0027`。两个 application history migration 只增加同一 owner 的读取索引，不新增 export 表或 repository。运行角色只授予必要 DML，migration role 与 runtime role 不得互换；旧结果资产在生命周期迁移中只回填 active v1，不改写 artifact payload；旧 Session / Run 记录不自动迁移到 v2 / v4 / v8。

## 停止线

- 不自动 review、activation、publish、assignment 或 profile 选择。
- 不增加 schedule、retry / fallback、replay / resume、agent loop、长期记忆或后台任务。
- 不从 session transcript 派生持久记忆，不把 answer 写回上层业务真相源。
- 不允许客户端事后上传 artifact content、digest 或 run ref，不从日志、缓存或 Provider 重建结果；memory artifact 不声明重启恢复，SQLite / PostgreSQL artifact 只声明开发测试态重启恢复。
- 不把本地 SQLite、PostgreSQL dev/test、mock provider、真实浏览器验收或 launcher 连续链解释为 production ready。
- 不绕过 HTTP Tool、Workflow RAG、Application RAG 或 Workflow Definition 各自的 authority owner。
