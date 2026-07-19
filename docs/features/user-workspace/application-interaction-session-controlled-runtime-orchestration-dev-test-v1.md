# 应用交互会话与受控运行编排（开发 / 测试态）v1

更新时间：2026-07-19

状态：`application_interaction_session_controlled_runtime_orchestration_dev_test_v1_batch_c_ready_for_implementation`

## 功能定位

本专题在用户工作区建立正式的 Application 使用入口，把应用目录、人工激活的 Workflow Definition、Application RAG runtime、运行历史和诊断交接组织为显式用户会话。用户选择一个活跃应用和一个明确执行 profile，创建会话后逐次提交用户回合；服务端在每个回合执行前重新解析精确 authority，再把执行委托给现有 `workflow_definition_executor_v1` 或 Application RAG invocation service。

会话不是新的运行真相源，也不是长期自然语言记忆。它只拥有 session / turn 拓扑、状态、幂等键、精确 authority metadata 和既有 run record 引用。原始 input、answer、prompt、provider raw response、credential、token、header、检索 fragment 正文和浏览器 transcript 不进入会话存储。

## 现状与用户缺口

- Application Catalog、配置草案、发布审查、API key、Application RAG runtime 和 Application Operations 已完成，但用户仍需在多个页面或协议入口间自行判断如何使用当前应用。
- Workflow Definition v5 已具备 exact active authority、执行前 checkpoint 和严格 metadata-only run record，但其调用入口要求客户端直接携带 pointer / version / digest。
- Application RAG v4 已具备 active assignment、每次调用重校验和受控 answer，但与 definition-bound execution 没有共同的 Application-scoped 会话入口。
- 既有 `conversation_session_record` 与 recovery checkpoint 只冻结 northbound metadata / governance 契约，不是 durable Application session store，也不得被误写成长期记忆或自动 replay。

因此，本专题新增独立的 Application Interaction Session owner，并复用现有两条执行 authority 与运行证据链。

## 目标用户路径

1. 用户在当前 tenant / workspace 中选择一个 active Application。
2. 用户显式选择 `workflow_definition_executor_v1` 或 `application_rag_invocation_v1`；Workflow profile 同时选择稳定 `definition_id`。
3. 服务端读取 Application lifecycle，并解析该 profile 当前有效的 exact authority。authority 不存在、失效、漂移或 profile 不兼容时，会话创建失败。
4. 服务端创建 metadata-only session，返回 session id、record version、profile、authority snapshot 和内容保留策略。
5. 用户提交一个显式 turn。输入只存在于当前 HTTP 请求和执行调用栈；会话记录只保存 input digest / bytes。
6. turn coordinator 在 provider 调用前重新解析 authority，并与会话当前期望比较。任何 activation / assignment / digest / application lifecycle / eligibility 漂移均失败关闭。
7. coordinator 将执行委托给已有服务：Workflow profile 产生 `workflow_run_record.v5`；Application RAG profile 产生 `workflow_run_record.v4`。
8. terminal turn 只保存 run id、run schema、状态、failure metadata、authority metadata 和审计引用；answer 仅在当前响应中返回。
9. 用户可分页查看 session / turn metadata，跳转 Run History、Comparison 或 Evaluation；关闭 session 后不再允许新增 turn。
10. 服务重启后恢复 session / turn metadata 和 run 引用，不恢复 transcript，也不自动重放 provider。

## Owner 与职责边界

### Application Interaction Session owner

负责：

- `application_session.v1` 当前投影；
- `application_session_turn.v1` running projection 与 terminal evidence；
- tenant / workspace / application / owner 隔离；
- session record version CAS、turn sequence 与 client turn key 幂等；
- session / turn 分页、关闭、stale reconciliation 和既有 run 引用；
- metadata-only 持久化约束。

不负责：

- Workflow 图拓扑、预算、取消、Gateway 或 diagnostics；
- RAG ranking、citation、assignment 或 answer；
- Application lifecycle、Definition activation 或 RAG assignment 写入；
- transcript、长期记忆、业务真相源、replay / resume、schedule 或 agent loop。

### Exact Application Runtime Authority resolver

resolver 是只读、无 provider side effect 的执行前边界：

- 总是重读 Application Catalog，并要求 `lifecycle_state=active`；
- Workflow profile 重读 activation pointer、active definition version、canonical digest、source application 和 `workflow_definition_executor_v1` eligibility；
- Application RAG profile 重读 runtime assignment，并复用现有 authority resolver 校验 approved publish candidate、draft、binding、promotion、dataset、snapshot 和 lexical profile；
- 返回 strict metadata snapshot 与 canonical authority digest；
- 不接受客户端提交的 version、digest、assignment id 或 snapshot id 作为真相。

session 保存的 authority 只用于审计和漂移比较，不替代每回合重读。

## 执行 profile

首版只允许两个显式 profile：

| profile | 运行 owner | terminal run | 必要绑定 |
| --- | --- | --- | --- |
| `workflow_definition_executor_v1` | Workflow Definition execution service | `workflow_run_record.v5` | `definition_id` |
| `application_rag_invocation_v1` | Application RAG invocation service | `workflow_run_record.v4` | 当前 Application active assignment |

Application kind 不隐式决定 profile。profile 必须由用户明确选择，并由 resolver 证明当前 Application 对应 authority 可用；这样不会把 `agent`、`docs_qa` 等目录标签误当执行权限。

一个 turn 最多委托一次既有执行服务。Session coordinator 不直接调用 Gateway，不复制 executor v0 图算法，不在两个 profile 间自动 fallback。

## 数据契约

### `application_session.v1`

固定包含：

- session / scope / owner；
- `active | closed` 状态和 record version；
- explicit execution profile 与 profile binding；
- exact authority snapshot / digest；
- `metadata_only` 内容保留策略；
- turn count、last turn id、时间、actor、request / audit ref。

### `application_session_turn.v1`

固定包含：

- session id、turn id、严格递增 sequence 和 client turn key；
- execution profile 与当次 exact authority snapshot / digest；
- `running | succeeded | failed | canceled | outcome_unknown`；
- input digest / bytes；
- 可空 run reference、failure code / summary；
- started / completed、actor、request / audit ref。

turn 不保存 input、answer、prompt、模型原始响应、retrieval context、header、token 或 credential。session detail 不通过读取 run detail 拼装 transcript。

## 状态与并发

- create：仅 active Application 和可解析 profile authority 可以创建 session，初始 version 为 1。
- reserve turn：要求 session active、`expected_session_version` 精确匹配，在 provider 前原子分配 sequence、记录 running metadata 并推进 session version。
- terminal write：只允许 `running → succeeded | failed | canceled | outcome_unknown`，不再次推进 session version。
- idempotency：同一 scope / session / client turn key 只对应一个 reservation；完全相同的 reservation 或 terminal write 返回已有记录，不同内容返回冲突。
- close：CAS 把 active 变为 closed；重复 close 返回当前 closed 投影，错误 expected version 返回冲突。
- stale：持久化为 running 且超过阈值的 turn 只能改为 `outcome_unknown`，不得调用 provider 或伪造 run terminal 状态。

## API 边界

管理面使用 verified control-plane actor 与明确 scope：

- `application_sessions:read`
- `application_sessions:write`
- `application_sessions:execute`

计划路由：

- `POST /v1/user-workspace/application-sessions`
- `GET /v1/user-workspace/application-sessions`
- `GET /v1/user-workspace/application-sessions/{session_id}`
- `POST /v1/user-workspace/application-sessions/{session_id}/close`
- `GET /v1/user-workspace/application-sessions/{session_id}/turns`
- `POST /v1/user-workspace/application-sessions/{session_id}/turns`

所有路由默认关闭，仅在显式开发测试态 gate 下启用。未知字段、未知 query、scope 不匹配、owner 不匹配、跨 Application 访问和非法 cursor 均失败关闭。

## 当前实施状态

- 批次 A 已完成：strict Session / Turn contract、两个显式 profile、exact Application runtime authority resolver、memory owner、provider 前 reservation、terminal write、CAS / 幂等与默认关闭的管理 API 已落地。
- 批次 B 已完成：Session / Turn repository 已接入 memory、共享 SQLite runtime 与既有 Workflow PostgreSQL pool；SQLite `0012_application_interaction_sessions` / marker v12、PostgreSQL `0015_application_interaction_sessions` / marker v15 已落地，没有新增 DSN、pool、database file、selector 或 fallback。
- 双数据库均以事务保护 session version、turn sequence 与 client turn key；数据库约束禁止删除 session / turn，并只允许 active session 版本递增以及 `running` turn 向终态迁移。
- SQLite 已覆盖重启恢复、并发 CAS、损坏投影拒绝、关闭连接不回退与敏感内容扫描；PostgreSQL integration 已覆盖 migration / rollback / reapply、运行角色、重启、并发、受控更新、不可删除和 no-fallback。
- 当前只开放 session 管理面；批次 C 才启用 turn execution route、每回合 authority reload、既有 v5 / v4 执行委托与 stale reconciliation。Web、launcher 和真实浏览器仍由批次 D 承接。

## 实施批次

### 批次 A：设计、strict contract、exact resolver 与 memory session owner（已完成）

- 冻结 session / turn JSON Schema 和 Go codec。
- 实现两个显式 profile 的 exact authority resolver。
- 实现 memory repository、session create / read / list / close、turn reservation / terminal write 与幂等 CAS。
- 注册默认关闭 gate、管理 scopes 和 session 管理 API；turn execution route 保持关闭。
- 证明 resolver 零 provider 调用、存储 metadata-only、旧 Session metadata route 不变。

### 批次 B：SQLite / PostgreSQL durable repository（已完成）

- 在共享本地持久化与 PostgreSQL 开发测试态配置中加入 session store，不新增 DSN、pool 或 fallback。
- 增加 session projection、turn evidence、client key uniqueness、CAS 和必要索引。
- 覆盖 migration / rollback / reapply、schema marker、运行角色、重启、并发、corruption 和 no-fallback。

### 批次 C：显式 turn coordinator 与既有执行服务委托（下一批）

- 接通同步 turn route；执行前重读 exact authority。
- Workflow profile 调用 definition service 并消费 v5；Application RAG profile 调用现有 invocation service 并消费 v4。
- answer 只返回当前请求；terminal turn 只保存 run ref 和 metadata。
- 覆盖取消、运行记录终写失败、authority 漂移、stale reconciliation 和零重复 provider 调用。

### 批次 D：Web 应用交互工作区

- 增加 Application-scoped session 列表、创建、关闭、当前 transient transcript、turn 状态和 run handoff。
- 应用切换清除 transcript、session selection、输入、answer 和迟到响应。
- offline 零请求；strict consumer 拒绝 schema、scope 和敏感字段漂移。

### 批次 E：双数据库连续链与专题收口

- 完成 SQLite / PostgreSQL launcher、服务重启、CAS、取消、漂移、stale、敏感扫描与真实浏览器路径。
- 验证 Workflow v5、Application RAG v4、v0–v4、Saved Draft、HTTP Tool、Workflow RAG 和 Application Operations 不回归。
- 同步专题、任务卡、能力矩阵、current focus 和周志，推进下一项正式功能选择。

## 验收方式

- authority：两个 profile 均从服务端 owner 重读；客户端不能覆盖 version / digest / assignment / snapshot。
- provider boundary：create / list / read / close 为零 provider；turn authority 漂移在 provider 前失败。
- concurrency：同 expected version 只有一个 append / close 成功；client turn key 不产生重复执行。
- privacy：数据库、HTTP metadata、日志、fixture 和 committed evidence 不出现原始 input、answer、prompt、provider response、credential、token、header 或 fragment 正文。
- recovery：服务重启后可读 session / turn / run refs；running turn 只 reconcile 为 outcome unknown，不自动 replay。
- compatibility：既有 Session metadata、Gateway、Workflow Draft、run v0–v5、HTTP Tool、Workflow/Application RAG、Comparison、Evaluation、Baseline 和 Suite 继续通过。

## 停止线

- 不实现长期记忆、durable transcript、自动摘要或 session 内容搜索。
- 不实现 agent loop、自动 profile 选择、自动 retry / fallback、schedule、background execution、replay / resume 或 writeback。
- 不自动 activation、replace、publish 或 release。
- 不增加新的 Gateway 调用算法、Workflow 图算法、RAG ranking 或跨 store 运行副本。
- 不进入生产认证、生产 secret、public API、quota、billing 或生产能力声明。
