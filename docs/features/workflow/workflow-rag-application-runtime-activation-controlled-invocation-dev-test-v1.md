# Workflow RAG 应用运行时激活与受控调用（开发 / 测试态）v1

更新时间：2026-07-19

状态：`workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_completed`

## 设计结论

本专题允许内部开发者把一个已经完成人工 RAG promotion、配置草案 binding 和发布候选审查的应用，显式激活为开发测试态 RAG 运行配置，再用该应用的开发测试态 API key 发起一次受控检索调用。调用方只提交有界输入，不提交 model、protocol、publish candidate、配置版本、binding、`rag_ref`、snapshot、profile、排名或引用；服务端从可信应用作用域和唯一当前 runtime assignment 重新读取全部权威资源。

运行时固定使用 binding 中的 candidate snapshot 作为有效检索快照；baseline snapshot 只保留为晋级证据和回归比较来源，不参与本次检索。每次调用恰好执行一次既有确定性 lexical ranker 和一次既有 Gateway bridge 调用，返回结构化 advisory answer，并写入 metadata-only `workflow_run_record.v4`。输入、片段正文、prompt packet、完整回答和模型原始响应不持久化。

本专题不会把发布候选 `approved` 改写为正式发布。runtime assignment 是独立、显式、可撤销的开发测试态运行选择，只引用权威发布候选和配置，不修改应用目录、配置草案、发布候选、promotion candidate、binding、dataset 或 snapshot。批次 A 已完成 strict contract、memory assignment、authority resolver、API key scope、独立调用 API 和 metadata-only run v4；批次 B 已完成 durable store、run source migration、v4 evaluation 与真实 PostgreSQL 专项；批次 C 已完成 Web、launcher、双数据库连续链、真实浏览器与专题收口。

## 批次 A 实施结果

- 已物化 assignment、event、audit、同步 answer 与 run v4 五份 strict JSON Schema；未知字段、非法状态、错误引用和非 metadata-only 记录均失败关闭。
- memory assignment repository 与现有 workflow owner lock 共用临界区，current projection、event 与 audit 以单次 CAS 原子提交；并发激活只有一个成功，重复或无意义状态变化被拒绝。
- 管理 API 已支持显式 `activate / replace / revoke`，调用 API 只接受 API key Bearer 与 `application_rag:invoke`；普通三协议调用 scope 不会隐式获得 RAG 权限，管理权限保持独立。
- authority resolver 在激活和每次调用检查点重读 application、publish candidate v2、draft v2、immutable binding、promotion、dataset / review、两侧 snapshot 与 lexical profile；candidate snapshot 是唯一有效检索快照。
- 成功调用恰好执行一次 lexical retrieval 和一次 Gateway，使用新的 application configuration execution source 写 memory `workflow_run_record.v4`；输入、fragment 正文、prompt、完整回答、模型原始响应、token 与 credential 不进入 run、assignment、event 或 audit。
- 精准领域、HTTP、权限、隐私、漂移、CAS 与竞态测试、平台全包 Go 测试、vet 和仓库快速 / 完整门禁均通过。批次 A 未创建数据库 migration，非 memory workflow backend 在门禁开启时明确失败，不回退 memory。

## 批次 B 当前结果

- SQLite `0009` 与 PostgreSQL `0012` 已物化 assignment current projection、append-only event / audit，并把 run store 迁移到 `execution_source_kind / id / version`；v0–v3 回填为 `workflow_draft`，v4 使用 `application_configuration_draft`，旧 draft filter 不返回 v4。
- SQLite assignment current projection、event、audit 与 run v4 共用既有数据库；PostgreSQL 实现复用既有 workflow pool。两种实现均无独立 DSN、pool、selector 或 memory fallback。
- v4 已接入 Run History、Comparison、Evaluation、Baseline 与 Suite 的 `workflow_rag_application_invocation.v1` metadata-only 审查边界；stale running reconciliation 只写失败终态，不重放 retrieval 或 Gateway。
- SQLite 已覆盖迁移升级、HTTP Tool 不回归、CAS、事务回滚、append-only、重启恢复、损坏拒绝和 no-fallback；平台 `go test ./...`、定向 race、`go vet` 与 PostgreSQL build-tag 编译通过。
- 真实 PostgreSQL 专项已覆盖 migration / rollback / reapply、运行角色、事务、append-only、重启恢复和 no-fallback；首次真实运行发现 configured 产品链清理未先删除新 runtime 表，修正清理顺序后，完整 integration suite 与 `check` 配置闭环通过，容器和网络已关闭。

## 批次 C 实施结果

- Web 已接入发布候选侧 lazy assignment 管理、API key `application_rag:invoke` 签发、一次性内存交接、独立 Application RAG Invocation 面板，以及 v4 Run History、Comparison 和 Evaluation 审查。应用切换会清除 token、输入、回答、assignment、run handoff 和对比状态。
- 跨平台 launcher 新增单一 `workflow-rag-application-local-product` / `WorkflowRAGApplicationLocalProduct` 产品档，统一开启 SQLite 应用、配置 / 发布、API key、runtime assignment、受控调用、运行历史和评测链；显式组件组合仍由 `configured` 档承接。
- mock provider 已显式支持 `workflow-rag-application-invocation-v1` 结构化回答契约；Comparison v3 严格区分 application configuration execution source，并稳定返回 baseline / candidate authority、assignment 变化和零越权副作用。
- SQLite 与 PostgreSQL 连续链均覆盖 approved candidate → activate → 两次相同输入调用 → comparison v3 → evaluation → revoke 后拒绝 → restart restore；未增加平行 database、pool、selector 或 fallback。
- 真实浏览器以同一持久应用完成 activation、一次性交接、无证据失败、两次成功 v4、comparison、evaluation、revoke 后 `workflow_rag_runtime_assignment_revoked`、应用切换清理和 SQLite 重启恢复。新标签页 console 为零 error / warning，`localStorage` 与 `sessionStorage` 为空，原始 token、输入和回答未进入浏览器持久介质。

## 目标用户与完整路径

目标用户是需要验证“经过知识质量审查和配置审查的 RAG 应用是否能按绑定证据真实调用”的内部应用开发者、Workflow reviewer 和平台维护者。

1. 开发者在现有链路中完成 dataset / candidate review、人工 promotion approve、配置草案 attach，并创建 `application_publish_candidate.v2`。
2. reviewer 人工批准该发布候选；现有 production promotion blockers 继续保留，不发生正式发布。
3. 具有 runtime assignment 权限的操作者在发布审查区选择精确候选，显式执行 `activate` 或 `replace`；服务端重读候选、草案、binding 和全部知识权威来源后，写入当前应用唯一 assignment 与 append-only event / audit。
4. 开发者签发同一应用的开发测试态 API key，并显式包含 `application_rag:invoke`；原始 token 仍只在既有一次性交接内存中存在。
5. 调用方对独立应用 RAG route 提交一次有界输入。服务端从 API key 恢复 tenant / workspace / application / owner，读取 active assignment，不接受客户端资源引用。
6. 服务端在副作用前重新校验应用、发布候选、配置草案、binding、promotion、dataset、candidate review、baseline / candidate snapshot 和 lexical profile；任一漂移、归档、取消、损坏或 store failure 均失败关闭。
7. 服务端以 candidate snapshot 和绑定 profile 执行一次检索，以候选配置中的 `default_model` 和 `default_protocol` 构造现有 canonical Gateway request，验证回答与引用后返回同步结果。
8. Run History 展示运行选择、有效 snapshot / profile、排名与引用 metadata、失败边界和副作用计数；Comparison / Evaluation / Suite 可以比较两个 v4 记录，但不会重新执行 retrieval 或 Gateway。
9. 操作者可显式 `revoke` 或用新候选 `replace` assignment。变更只阻止后续调用，不改写或重放已开始和已完成记录。

## 与既有资源的职责边界

| 资源 / 服务 | 唯一职责 | 本专题不得让其承担 |
| --- | --- | --- |
| API key lifecycle | 恢复可信应用调用作用域，校验有效期、吊销和 `application_rag:invoke` | 不保存 assignment、配置、binding 或输入 |
| application publish candidate | 持有已审查的不可变配置快照、draft ref、binding ref 与 review history | 不作为 active runtime 指针，不因 approve 自动调用 |
| application configuration draft | 持有服务端 canonical 配置与 ref-only binding | 不保存 activation、query、retrieval 结果或回答 |
| promotion repository | 持有 exact evidence、decision 与 immutable binding | 不选择运行时 active candidate，不保存调用材料 |
| runtime assignment repository | 持有当前应用唯一的 ref-only 开发测试态运行选择和 append-only event / audit | 不复制配置正文、评测指标、fragment、prompt 或模型响应 |
| snapshot / lexical provider | 提供 exact candidate snapshot 和版本化确定性排名 | 不读取客户端排名，不访问网络或其它应用 |
| Workflow run store | 保存 v4 metadata-only 执行证据、诊断和 evaluation refs | 不伪造 workflow draft，不保存输入、正文或完整回答 |
| Gateway bridge / request history | 完成一次既有 provider 调用并保存既有脱敏调用元数据 | 不成为 assignment、binding 或知识真相源 |

## Runtime assignment 资源模型

新增 `workflow_rag_application_runtime_assignment.v1`，每个 tenant / workspace / application / owner 只有一个 current projection。字段至少包括：

- `assignment_id`、`record_version`、`assignment_digest`、scope、`state=active|revoked`；
- 精确 `publish_candidate_id / review_version / candidate_state`；publish candidate v2 当前没有独立 candidate digest，assignment digest 以 candidate id、review version、draft digest 和 binding ref 封印选择，不伪造新字段；
- 精确 `draft_id / draft_version / draft_digest`；
- 精确 `binding_id / binding_version / binding_digest`；
- actor、request / audit ref、created / updated time。

assignment 不复制 publish candidate 配置、promotion evidence、dataset metrics、review findings、snapshot fragments 或模型材料。以上 exact refs 只是运行选择和一致性封印；权威内容仍由原 repository 持有，调用时必须重新读取并验证一致。

新增 `workflow_rag_application_runtime_assignment_event.v1`，decision 只允许：

- `activate`：不存在 assignment 时以 `expected_record_version=0` 激活已批准候选；
- `replace`：active 或 revoked assignment 指向另一个已批准候选；
- `revoke`：把 active assignment 变为 revoked，不删除历史。

每次 decision 都要求 4–500 字符脱敏 reason 和 `expected_record_version`。current projection CAS、event append 与 audit append 必须在同一 workflow backend 事务内提交；冲突不自动 retry / merge。重复 activate、对 revoked 重复 revoke、同一 candidate 的无意义 replace 均拒绝。

## 运行记录与有效配置

新增 `workflow_run_record.v4`，`execution_kind=application_rag_invocation`。它不使用或伪造 Saved Workflow Draft，至少保存：

- runtime assignment id / version / digest；
- publish candidate id / review version / state，以及 assignment 中的 exact draft / binding seal；
- application configuration draft id / version / digest；
- RAG binding id / version / digest；
- dataset、candidate review、baseline / candidate snapshot 与 profile 的 exact metadata refs；
- `effective_snapshot_role=candidate` 和有效 snapshot id / version / digest / `rag_ref`；
- input digest / byte count、selected fragment refs / digests / ranks / source types、citation refs；
- configured protocol / model、provider、latency、status、稳定 failure、request / audit ref；
- `retrieval_calls=1`、`provider_calls=1`，以及 `tool_calls=0`、`confirmation_calls=0`、`business_writes=0`、`replay_writes=0`。

物理 run store 增加 `execution_source_kind / execution_source_id / execution_source_version`，历史 v0–v3 回填为 `workflow_draft`；v4 使用 `application_configuration_draft`。旧 draft filter 不得返回 v4，Run History 按 execution kind 显式区分。不得把 application draft id 填进旧 `draft_id` 字段伪装兼容。

同步响应使用 `workflow_rag_application_answer.v1`，只在请求内存和响应中保留 answer / citations / limitations / confidence；持久 v4 的 answer 继续为 `null`。调用 route 不支持 stream、batch、background、retry、resume 或 replay。

## 权威重读与线性化语义

assignment decision 与 invocation claim 必须按以下顺序重读：

1. active application 与 API key / actor scope；
2. exact `application_publish_candidate.v2`，状态必须为 `approved`，review version / digest 与 assignment 一致，且未被取代；
3. exact application draft v2，version / digest / validation / application base revision 与 candidate 一致；
4. exact immutable RAG binding、promotion candidate 与 approve decision；
5. exact dataset version、candidate review、baseline / candidate snapshots 和 lexical profile；
6. assignment current record version仍与 claim 一致。

只有以上检查全部通过才创建 running v4。该写入是本次调用的资格线性化点；之后 assignment replace / revoke 或权威资源归档只阻止新调用，不追溯改写或自动取消当前运行。服务端在 retrieval 前和 provider 前复核 assignment current version及不可变证据摘要；发现变化则写失败终态且对应副作用保持为 0。开发测试态仅支持单平台进程协调，不据此声明多实例生产级分布式锁语义。

## 调用协议、权限与门禁

新增独立默认关闭门禁：

```text
RADISHMIND_WORKFLOW_RAG_APPLICATION_INVOCATION_DEV=1
```

管理 API：

```text
GET  /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment
POST /v1/user-workspace/applications/{application_id}/workflow-rag-runtime-assignment/decisions
```

调用 API：

```text
POST /v1/application-rag/invocations
```

调用 route 只接受 `api_key_dev_test` Bearer，不接受开发身份头和管理端凭据。body 只允许 `input`；model、protocol、temperature、assignment、candidate、draft、binding、`rag_ref`、snapshot、profile、selected fragments 和 citations 都禁止。配置的 `default_protocol` 必须存在于 `allowed_protocols`，服务端使用既有 canonical builder / response adapter 和 `default_model`；返回固定 RadishMind answer envelope，不能伪装成现有三条兼容 route。

权限固定为：assignment 读取 `workflow_rag_runtime:read`，activate / replace / revoke `workflow_rag_runtime:write`，应用调用 API key scope 为 `application_rag:invoke`；Run History / Comparison / Evaluation 继续使用既有 read / review scopes。拥有普通 `chat:invoke`、`responses:invoke` 或 `messages:invoke` 不自动获得 RAG 调用权限，反之亦然。

## 失败关闭与漂移

稳定失败至少包括：

| failure code | 语义 |
| --- | --- |
| `workflow_rag_runtime_assignment_not_found` / `_revoked` | 当前应用没有 active assignment |
| `workflow_rag_runtime_assignment_version_conflict` | decision 或 invocation claim 的 CAS 版本过期 |
| `workflow_rag_runtime_candidate_not_approved` / `_superseded` | 发布候选未批准、已撤回或被取代 |
| `workflow_rag_runtime_configuration_changed` / `_invalid` | exact draft version / digest / validation / base revision 漂移 |
| `workflow_rag_runtime_binding_not_eligible` | binding 已取消、损坏或任一 promotion authority 失效 |
| `workflow_rag_runtime_application_archived` | 应用已归档 |
| `workflow_rag_runtime_scope_denied` | API key、actor、owner 或三层 scope 不一致 |
| `workflow_rag_runtime_payload_invalid` / `_secret_material_forbidden` | strict body、reason、预算或敏感材料非法 |
| `workflow_rag_runtime_no_evidence` / `_budget_exceeded` | lexical retrieval 无证据或超预算 |
| `workflow_rag_runtime_gateway_failed` / `_answer_invalid` / `_citation_invalid` | provider 或结构化回答失败 |
| `workflow_rag_runtime_store_unavailable` / `_store_contract_mismatch` | 任一 repository、marker、事务、decode 或 authority read 失败 |
| `workflow_rag_runtime_write_disabled` | 独立门禁未开启 |

现有 publish `promotion_eligibility` 的正式存储库、生产认证、发布 owner 与 production promotion runtime blockers 必须继续原样展示。开发测试态 assignment 使用独立 invocation eligibility，不能删除、覆盖或伪造这些 production blockers。

## Evaluation 与审查

- Run History 为 v4 增加 assignment / publish candidate / draft / binding / effective snapshot / profile refs、排名、引用、失败和副作用视图；普通 list 仍只返回 metadata。
- Run Comparison 只允许 v4 与 v4 比较，要求同 application、同 input digest 和相同 answer contract；允许 snapshot / binding / assignment 不同，以便审查知识晋级后的 ranking / citation 变化。
- 新 profile `workflow_rag_application_invocation.v1` 复用既有 Comparison、Evaluation Case、Baseline 和 Suite repository，只保存 run refs、差异摘要、decision 与 audit refs，不保存输入或回答。
- regression 至少覆盖 expected citation loss、official source demotion、selected ref / rank / digest 变化、no-evidence、citation invalid、configured model / protocol drift 和副作用越界。
- 创建 evaluation case、批准 baseline 或 suite decision 不重新执行 invocation，也不自动 activate / replace assignment。

## Store ownership 与迁移顺序

- assignment、event、audit 与 v4 run 归现有 Workflow runtime owner，继续从 workflow backend selector 派生；memory 使用既有 owner lock，SQLite 使用 shared database，PostgreSQL 使用既有 workflow pool。
- 已实现 SQLite `0009_workflow_rag_application_invocations`，marker 推进为 `workflow_run_store_sqlite_v9`；已实现并真实验证 PostgreSQL `0012_workflow_rag_application_invocations`，marker 推进为 `workflow_run_store_v12`。
- application catalog、API key、application draft 与 publish candidate 继续使用各自现有 repository；实现只能通过 service / repository interface 重读，不在 workflow migration 中复制其表或正文。
- 启用顺序为：v4 decoder / run store兼容 → workflow migration → assignment repository / authority resolver → API key scope / invocation route → Comparison / Evaluation → Web。任一 marker、pool、shared database 或依赖 repository 不兼容时服务启动失败，不回退 memory 或旧草案 execution。

## Web 路径

- 发布候选详情新增 lazy runtime assignment 区域，显示 exact candidate / draft / binding refs、current version、state 和 blockers；approve 不自动 activate。
- activate / replace / revoke 为独立确认动作；CAS conflict 保留本地 reason，刷新后由用户重新决定。
- API key 页面只允许显式签发 `application_rag:invoke`，原始 token 仍只在内存中一次性交接到独立 Application RAG Invocation 面板。
- 调用面板只显示输入框、active assignment metadata、同步 answer 和调用状态；不允许选择 model、snapshot、`rag_ref` 或排名。输入和回答不写 URL、storage、日志或历史。
- 成功 / 失败后用 run id 交接既有 Run History，并可继续进入 v4 Comparison / Evaluation；应用切换清除 token、输入、answer、assignment detail、run handoff 和 conflict state。
- 默认 offline 零请求。真实浏览器必须验证 activation、API key 交接、一次成功、一次无证据或 citation failure、Run History、comparison、revoke 后失败关闭、应用切换和服务重启恢复。

## 实施批次与准入

唯一实施入口为[实施任务卡](../../task-cards/workflow-rag-application-runtime-activation-controlled-invocation-dev-test-v1-plan.md)：

1. 批次 A：strict contracts、assignment / event / audit、authority resolver、API key scope、application invocation service、memory repository、v4 record 与管理 / 调用 API；使用 fake bridge / deterministic ranker 完成全部失败、漂移、CAS、零越权副作用测试。
2. 批次 B：SQLite `0009`、PostgreSQL `0012`、run source migration、事务 / append-only / restart / corruption / no-fallback，以及 v4 Run History、Comparison / Evaluation / Baseline / Suite 服务边界。
3. 批次 C：Web activation / API key handoff / invocation / history / evaluation 接线、launcher、SQLite 与 PostgreSQL 连续链、真实浏览器、敏感材料扫描和专题收口。

批次 A、B、C 均已完成，专题状态为 `workflow_rag_application_runtime_activation_controlled_invocation_dev_test_v1_completed`。后续不派生批次 D 或同层 readiness / checker；下一项回到总体规划和功能设计入口选择新的真实用户流程。仍不得跳过 assignment，让 publish approve 或 API key 签发直接打开 retrieval。

## 验证矩阵

- domain / HTTP / auth：strict body、API key only、scope、active application、assignment state / CAS、全部 authority reload、候选 approved / superseded、生产 blocker 不被覆盖。
- execution：candidate snapshot 选择、一次 ranker、一次 Gateway、answer / citation validator、checkpoint、cancel / interrupted、retrieval / provider 前重校验、所有前置失败调用数为 0。
- store：memory / SQLite / PostgreSQL 行为一致，migration / rollback / reapply、marker / checksum、append-only、并发、重启、损坏记录、运行角色和 no-fallback。
- privacy：list / detail / audit / run / Gateway history / logs / Web state 不保存 input、fragment、excerpt、prompt、answer、模型响应、token、credential 或 secret。
- evaluation：v4-only compatibility、同 input digest、ranking / citation / official-source regression、无自动执行、无 binding / assignment mutation。
- 不回归：v0–v3 run、Saved Draft execution、HTTP Tool、原 RAG retrieval、promotion、application draft / publish、API key 三协议调用和 request history。

每批至少执行精准 Go / Web 测试、`git diff --check` 与 `./scripts/check-repo.sh --fast`；批次 B migration 与批次 C 阶段收口补真实 PostgreSQL、race / vet、生产构建和完整 `./scripts/check-repo.sh`。

## 停止线

- 不自动 activate、replace、revoke、baseline、promotion、release 或 publish；所有状态变化都由显式人工动作触发。
- 不修改 application、draft、publish candidate、promotion、binding、dataset、review 或 snapshot；不把 assignment 写成正式应用配置。
- 不复用或放宽 executor v0、Saved Draft `/runs` 或原 retrieval execution route；v0–v3 语义保持冻结。
- 不启用 connector、crawler、文件扫描、在线搜索、embedding、vector database、reranker、query expansion、hybrid search 或跨应用检索。
- 不实现 stream、batch、background、schedule、retry、fallback、replay、resume、writeback、agent loop 或自动漂移修复。
- 不接生产 OIDC / membership、生产 repository / audit / secret、生产 API key、quota、billing、公开 SLA 或多实例生产锁；开发测试态 API key 和 PostgreSQL 证据不等于生产就绪。
- 不创建第二张任务卡、平行 selector / DSN / pool / database file、同层 readiness 文档或专项 checker 链。
