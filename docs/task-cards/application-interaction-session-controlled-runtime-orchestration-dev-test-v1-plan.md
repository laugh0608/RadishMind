# 应用交互会话与受控运行编排（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-19

状态：`application_interaction_session_controlled_runtime_orchestration_dev_test_v1_batch_e_ready_for_implementation`

## 目标与准入结论

按[功能设计](../features/user-workspace/application-interaction-session-controlled-runtime-orchestration-dev-test-v1.md)交付“选择 Application → 显式 profile → 创建 metadata-only session → 提交受控 turn → exact authority reload → 复用 v5 / v4 执行 → session / run evidence 审查”的开发测试态路径。

本专题新增 Session / Turn owner 和协调边界，不新增运行真相源、Gateway 算法、Workflow 图算法、RAG 算法或长期记忆。新 API、contract、migration 和执行边界需要独立任务卡承载。

## 前置基线

- 实现基线为 `692d25aa`，工作区干净，当前分支为 `dev`。
- Application Catalog、配置 / 发布治理、API key、Application RAG v4、Application Operations 和 Workflow Definition v5 已完成。
- `workflow_definition_executor_v1` 已提供 exact active authority 与 provider 前 checkpoint；Application RAG invocation 已提供 active assignment 与 provider 前 revalidation。
- 既有 `conversation_session_record` / recovery checkpoint 为 northbound metadata governance 契约，不作为本专题 repository。
- 所有新增记录默认 metadata-only；所有开发测试态 gate 默认关闭。

## 批次 A：strict contract、exact resolver 与 memory owner

状态：`completed`。

### 允许实现

- 新增 `application_session.v1`、`application_session_turn.v1` strict JSON Schema 与 Go domain validation。
- 新增 `workflow_definition_executor_v1` / `application_rag_invocation_v1` 显式 profile binding。
- 新增只读 exact authority resolver，复用 definition / Application RAG 既有 repository 与 validator。
- 新增 memory repository，在同一 owner lock 下实现 session create / read / list / close、provider 前 turn reservation、terminal write、CAS 和 client turn key 幂等。
- 新增默认关闭的管理 gate、`application_sessions:read|write` scopes 和 session 管理 HTTP；execute scope 与 turn route 留待批次 C 启用。

### 必须证明

- Application archived、definition inactive / digest drift / ineligible、RAG assignment missing / revoked / changed 均失败关闭。
- authority resolution、session 管理、turn reservation 和 terminal write 均不会自行调用 provider。
- 同 expected version 的并发写只有一个成功；同 client turn key 不重复追加。
- session / turn 记录不包含原始 input、answer、prompt、provider raw response、credential、token、header、endpoint 或 fragment content。
- tenant / workspace / application / owner scope、strict JSON、method 和 unknown query 均失败关闭。
- 既有 `/v1/session/*` metadata route、Workflow / RAG / Gateway 行为不变。

### 完成证据

- 三份 strict JSON Schema、Go domain validation、memory repository、管理 HTTP、默认关闭 gate 与 scopes 已落地。
- exact resolver 已覆盖 Workflow active / inactive 和 Application RAG active / revoked，并证明失败路径不会调用 provider。
- reservation 在 provider 前持久化 `running` metadata；terminal write 不推进 session version，同 client turn key 的相同请求幂等、不同内容冲突。
- Platform、相关 race、vet 与仓库 fast 已通过；没有进入执行 route、Web、launcher 或浏览器。

## 批次 B：双数据库 durable repository

状态：`completed`。

- 复用共享 SQLite runtime 与 Workflow PostgreSQL pool / selector，不新增数据库连接或 fallback。
- 增加 session projection、turn evidence、scope / updated cursor / client turn key 索引与数据库 CAS。
- 覆盖迁移 / 回滚 / 重放、marker / checksum、运行角色、重启、并发、损坏、no-fallback 和敏感字段扫描。

完成证据：

- SQLite `0012` / marker v12 与 PostgreSQL `0015` / marker v15 已纳入既有 Workflow migration family；repository selector 复用 shared SQLite database / Workflow PostgreSQL pool。
- session CAS、turn reservation、terminal transition、client key uniqueness、scope cursor、删除拒绝与严格 payload projection 均由 repository 和数据库共同约束。
- SQLite 重启、并发、损坏、关闭连接不回退和敏感扫描通过；PostgreSQL integration 的 migration / rollback / reapply、运行角色、受控更新、重启、并发与 no-fallback 通过。
- v0–v5、Saved Draft、HTTP Tool、Workflow RAG、Application RAG 与 configured product integration 回归通过；没有启用 turn execution。

完成后推进批次 C。

## 批次 C：turn coordinator 与 v5 / v4 委托

状态：`completed`。

- 开放 strict turn route 与 `application_sessions:execute`。
- 每回合 provider 前重读 exact authority；不信任 session snapshot 或客户端 authority。
- Workflow profile 委托现有 definition execution service；Application RAG profile 委托现有 invocation service。
- 原子记录 running / terminal metadata 和 run ref；answer 只出现在当前响应。
- stale running 只转 `outcome_unknown`，不重放 provider。

完成证据：

- strict turn route 与 `application_sessions:execute` 已开放，RAG profile 不接受 workflow-only model / temperature / condition options，客户端不能提交 authority 字段。
- reservation 后再次比较 exact authority；Workflow 只委托 definition service 并消费 v5，RAG 只委托 invocation service 并消费 v4，没有复制图执行、Gateway 或检索算法。
- 首次成功响应可返回 transient `advisory_output` / answer；session / turn repository 只保存 digest / bytes、run ref 和失败 metadata，幂等 replay 不返回旧 answer。
- 并发相同 client turn key 只产生一次 delegate / provider；authority drift、取消、run terminal pending、session terminal write failure 和 stale running 均有失败关闭测试。
- Platform 全量、相关 race、`go vet`、PostgreSQL integration suite 与仓库 fast / full 已通过，状态推进至批次 D ready。

完成后推进批次 D。

## 批次 D：Web 产品路径

状态：`completed`。

- Application-scoped session / turn consumer、交互工作区、transient transcript、close 和 run handoff。
- application switch / late response / cancel / offline / strict schema / secret guard。
- 不把 transcript 写 URL、storage、日志或持久 repository。

完成证据：

- 默认 offline 的 strict consumer 已覆盖 session list / create / close、turn list / execute 与 read / write / execute scopes；严格拒绝 unknown field、schema / tenant / workspace / application / owner 漂移、敏感字段和错误 run schema。
- Web 交互工作区已提供显式 Workflow Definition v5 / Application RAG v4 profile、会话选择、易失 transcript、Workflow 条件 / model / temperature、取消、metadata-only turn history 与 Run History handoff。
- application / session switch、unmount 和 cancel 共同推进请求代际并 abort 旧请求；input、answer、transcript 和运行选项不会写 URL、浏览器存储、日志或 repository，迟到响应不能回填当前应用。
- Application Interaction consumer 精准测试、全部 Web 单元测试、production build 与仓库 fast / full 已通过；本批没有启动 launcher 或真实浏览器。

完成后推进批次 E。

## 批次 E：连续链与专题收口

状态：`pending`。

- SQLite / PostgreSQL 连续链、重启、并发、漂移、取消、stale、no-fallback 与敏感扫描。
- Web 测试 / build、Platform Go / race / vet、PostgreSQL integration、仓库 fast / full 和真实浏览器。
- 同步功能专题、任务卡、入口、能力矩阵、current focus 和周志。

## 总停止线

- 不自动选择 profile、activation、release、retry、fallback、schedule、replay 或 resume。
- 不实现 durable transcript、长期记忆、agent loop、业务写回、production auth / secret / repository、quota 或 billing。
- 不复制 Workflow、RAG、Gateway、History、Comparison 或 Evaluation owner。
