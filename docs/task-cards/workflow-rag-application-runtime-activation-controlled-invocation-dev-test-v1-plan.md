# Workflow RAG 应用运行时激活与受控调用（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-18

状态：`workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_batch_a_ready_for_implementation`

## 目标与准入结论

按[功能设计](../features/workflow/workflow-rag-application-runtime-activation-controlled-invocation-dev-test-v1.md)交付“已批准 publish candidate v2 → 人工 runtime assignment → API key application scope → candidate snapshot retrieval → Gateway answer → metadata-only run v4 → regression review”的完整开发测试态路径。

设计、资源职责、执行边界和生产停止线已通过评审。本卡是唯一实现入口，只准入批次 A；不得先写 migration、Web 或真实运行开关，也不得绕过 assignment 直接从 publish approve / binding approve 触发调用。

## 前置基线

- `dev` 以 `d7b4a24b` 为设计起点，promotion / binding、application draft v2、publish candidate v2、API key dev/test auth、RAG snapshot / retrieval / evaluation 和 memory / SQLite / PostgreSQL store 均通过现有回归。
- runtime assignment 必须精确引用已批准且未被取代的 `application_publish_candidate.v2`；服务端从候选重读 exact draft、binding 和全部 promotion authority。
- 有效检索快照固定为 binding candidate snapshot，baseline snapshot 只作 evidence；调用方不能提交 model、protocol、resource ref、ranking 或 citation。
- API key scope 固定新增 `application_rag:invoke`，调用 route 只接受 `api_key_dev_test`；管理 assignment 使用 verified actor 和独立 management scopes。
- 所有 list / detail / event / audit / run / Gateway history / log 均 metadata-only，answer 只在同步响应中存在。

## 批次 A：contract、authority resolver 与 memory execution

状态：`ready_for_implementation`。

### 允许实现

- 物化 `workflow_rag_application_runtime_assignment.v1`、assignment event / audit、`workflow_run_record.v4` 和 `workflow_rag_application_answer.v1` strict contract。
- 实现 application-scoped current assignment、`activate / replace / revoke`、`expected_record_version` CAS、append-only event / audit 和 memory owner lock 原子提交。
- 实现服务端 authority resolver：application → publish candidate v2 → application draft v2 → immutable binding / promotion → dataset / candidate review → baseline + candidate snapshots → lexical profile。
- 实现 candidate snapshot 固定选择、assignment / authority seal、retrieval 前和 provider 前重校验，以及一次 ranker / 一次 Gateway 的同步 service。
- 新增独立 gate、assignment GET / decision API、API key `application_rag:invoke` scope 与 `POST /v1/application-rag/invocations`；调用 body 只接收有界 `input`。
- memory run v4 使用新的 execution source，不伪造 workflow draft；answer 不进入 record。

### 必须证明

- 未批准 / 撤回 / 被取代 candidate、draft / binding / dataset / review / snapshot / profile 漂移、应用归档、assignment revoked、scope / API key / store failure 全部在副作用前失败关闭。
- assignment CAS 并发单一成功；event / audit 与 current projection 无 partial write；replace / revoke 不修改任何来源资源。
- 所有前置失败的 ranker、Gateway、run write计数符合设计；running record 后的失败写真实终态，不 retry / replay。
- v0–v3、原 RAG execution、普通三协议 API key scope 和 publish production blockers 不回归。

### 批次 A 停止线

- 不创建 SQLite / PostgreSQL migration，不修改 Web / launcher，不执行真实 provider 或浏览器。
- 不打开 stream / batch / background / retry / fallback / replay / writeback，不接 connector / embedding / vector / reranker 或生产认证。

完成后推进为 `workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_batch_b_ready_for_implementation`。

## 批次 B：durable store、run source 与 evaluation

状态：`blocked_by_batch_a`。

- SQLite shared workflow database 追加 `0009_workflow_rag_application_invocations`，marker 推进为 `workflow_run_store_sqlite_v9`。
- PostgreSQL workflow migration family 在 `0011` 后追加 `0012_workflow_rag_application_invocations`，marker 推进为 `workflow_run_store_v12`。
- 重构 run store source columns：历史 v0–v3 映射 `workflow_draft`，v4 映射 `application_configuration_draft`；旧 draft filter 不返回 v4。
- assignment current projection、append-only event / audit 和 run v4 复用既有 database / pool；禁止独立 DSN、pool、database file、selector 或 fallback。
- Run History 接受 v4；Comparison / Evaluation / Baseline / Suite 增加 `workflow_rag_application_invocation.v1`，只消费 metadata-only run refs且不重新执行。
- 覆盖 migration / rollback / reapply、checksum、运行角色、并发 CAS、事务回滚、重启、损坏记录、stale running reconciliation 和 no-fallback。

完成后推进为 `workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_batch_c_ready_for_implementation`。

## 批次 C：Web、连续链与专题收口

状态：`blocked_by_batch_b`。

- 发布候选详情增加 lazy assignment 管理，approve、activate / replace / revoke 保持独立动作并保留 CAS conflict reason。
- API key 管理增加显式 `application_rag:invoke`，一次性 token 只通过内存交给 Application RAG Invocation 面板。
- 调用面板不允许选择 model / protocol / candidate / binding / snapshot，完成同步 answer、失败、Run History 与 v4 evaluation 交接。
- 应用切换清空 token、input、answer、assignment、run 和 conflict；offline 零请求，strict schema / scope / forbidden field 失败关闭。
- 完成 SQLite / PostgreSQL 连续链、重启 / corruption / no-fallback、真实浏览器成功 / 失败 / revoke / switch / recovery、URL / storage / console /敏感材料审计。
- 同步 current focus、入口、路线图、能力矩阵与周志，关闭专题和本卡；不派生第二张任务卡。

完成锚点为 `workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_completed`。

## 验证矩阵

- Go：contract、domain、HTTP、API key auth、authority reload、drift、CAS / race、execution checkpoint、answer / citation、memory / SQLite / PostgreSQL、migration、reconciliation、no-fallback、`go vet` 与全包回归。
- Web：offline、assignment decisions、one-time token handoff、invocation、v4 history / comparison / evaluation、application switch、strict schema / secret guard 和 production build。
- 连续链：publish candidate approved → assignment active → API key invoke → run v4 → evaluation → revoke blocked → replace → restart recovery。
- 停止线计数：每次成功恰好一次 retrieval / provider；tool、confirmation、business write、replay 为 0；所有前置失败 retrieval / provider 为 0。
- 仓库：每批 `git diff --check`、`./scripts/check-repo.sh --fast`；批次 B / C 按风险执行 PostgreSQL 专项和完整 `./scripts/check-repo.sh`。

## 总停止线

- 不自动 activate、baseline、promotion、release 或 publish；不修改任何来源资源。
- 不把开发测试态 assignment、API key 或 v4 run 声明为正式应用、production RAG 或可信 SLA。
- 不启用外部 connector、在线搜索、embedding、vector database、reranker、后台任务、retry、replay、resume、writeback、agent loop、production auth / secret / repository、quota 或 billing。
- 不创建平行基础设施、第二张任务卡、同层 readiness / checker 文档链。
