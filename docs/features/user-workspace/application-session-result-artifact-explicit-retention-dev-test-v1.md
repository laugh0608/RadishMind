# 应用会话运行结果资产显式保存与恢复（开发 / 测试态）v1

更新时间：2026-08-17

状态：`application_session_result_artifact_explicit_retention_dev_test_v1_completed`

## 功能定位

本专题为 Application Interaction Session 增加用户显式选择的运行结果保存能力。默认回合继续只返回易失结果；只有用户在提交新 turn 时明确要求保存，服务端才允许从同一次受控执行的 canonical result 捕获内容，创建独立的 `application_result_artifact.v1`，并绑定真实 session、turn 与 terminal run。

结果资产不是 transcript、Run History 扩展或 provider response archive。它是用户主动保留的一份应用结果，服务端拥有来源绑定、内容摘要、读取边界和后续生命周期；既有 `workflow_run_record.*`、`application_session.v*` 与 `application_session_turn.v*` 继续保持 metadata-only。

## 为什么曾是当前最高优先级

- Application Session 已能执行 Workflow Definition、结构化 Definition、Application RAG、Prompt 与 Agent / Copilot，但刷新或重启后只恢复 turn metadata 和 run ref，首次响应中的结果无法再次读取。
- 能力矩阵已把 materialized result reader 列为 `Conversation & Session` 的真实缺口；继续扩 HTTP Tool、Provider Attempt、只读 evidence 或 readiness 无法解决用户完成一次运行后保存成果的问题。
- 生产身份、secret backend、真实 Provider、billing、agent loop、自动执行与外部接入仍有明确前置阻塞，不适合作为当前产品批次。
- 通过独立 owner 和显式 opt-in，可以在不污染 Run History、不建立 durable transcript、不重放 provider 的前提下先形成开发测试态真实路径。

## 目标用户

- `Application Builder`：运行应用后，选择保存值得继续审查、比较或交付的结果，并在离开当前页面后重新读取。
- `Workflow Reviewer`：从 session / turn / run 精确来源确认结果由哪次受控执行产生，而不是读取客户端自行上传的无来源文本。
- `Platform Maintainer`：维持默认不保存、作用域隔离、内容大小上限、幂等、no replay 和 memory / SQLite / PostgreSQL 同构边界，并为后续生命周期与 Web 消费保持内容不可变和 no-fallback。

## 用户路径

1. 用户在现有 Application Interaction Session 中创建或选择 active session。
2. 用户提交新 turn；默认 `save_result=false`，结果仍只随首次响应返回。
3. 用户显式选择保存时，turn route 要求 `save_result=true`，并沿现有 `application_sessions:execute` 身份、成员资格和 application scope 执行。
4. coordinator 仍先完成 authority 重读、provider 前 reservation、单次既有 runtime 委托和 terminal turn 写入。
5. 只有 terminal turn 为 `succeeded`、run ref 与 profile 严格匹配且 canonical result 非空时，Result Artifact owner 才捕获内容；客户端不能提交 artifact content、digest、run ref 或 source metadata。
6. 首次响应返回原结果和 metadata-only artifact summary。保存失败不得伪造成功，也不得重新调用 provider；执行成功与保存失败必须分别表达。
7. 用户可按同一 application / session 列出 metadata-only artifact summary，并用精确 artifact id 读取内容。
8. 相同 client turn key 的重试不得重复 provider 调用或创建第二份资产；若首次未选择保存，重试不能从已丢失的易失结果补建资产。
9. SQLite / PostgreSQL 开发测试态现在可在服务重启后恢复 artifact；memory owner 仍明确不声明重启恢复。

## Owner 与职责边界

### Application Result Artifact owner

负责：

- `application_result_artifact.v1` 当前记录；
- tenant / workspace / application / owner / session / turn / run 精确绑定；
- 每个 turn 最多一个 artifact 的幂等唯一性；
- canonical content、content type、bytes 与 `sha256:` digest；
- metadata-only list 与精确 content read；
- memory / SQLite / PostgreSQL repository，以及后续 archive / unarchive lifecycle 和 Web consumer。

不负责：

- 创建或执行 session turn；
- 解析 runtime authority、调用 Gateway / Provider、重试或恢复执行；
- 修改 Run History、Comparison、Evaluation、Application Operations 或业务真相源；
- 保存用户 input、prompt、provider raw response、header、token、credential、retrieval fragment 正文或完整 transcript。

### Application Interaction Session owner

继续拥有 session / turn 拓扑、状态、幂等键、authority metadata 和 run ref。它只把首次执行产生的 canonical result 与 terminal turn 交给 Result Artifact owner，不保存 artifact content，也不把 artifact 成功作为 run 成功条件。

### 既有 runtime owner

Workflow Definition、Application RAG、Prompt 与 Agent / Copilot runtime 继续拥有各自执行、输出验证和 run record。Result Artifact owner 不复制其算法或重新读取 provider；v1 只接入已经由 runtime 校验成功的 canonical response。

## 数据合同

`application_result_artifact.v1` 固定包含：

- `artifact_id`、`record_version`；
- tenant / workspace / application / owner；
- session id、turn id、client turn key；
- execution profile；
- run id 与 run schema version；
- `text/markdown` 或 `application/json` content type；
- canonical content、content bytes 与 `sha256:` digest；
- created at / actor / request / audit ref。

列表只返回带生命周期投影的 `application_result_artifact_summary.v2`，不返回 content。精确 read 才返回 content，并同时返回 `application_result_artifact_lifecycle.v1`；两者都要求与记录一致的 application scope 和 owner。

v1 单份 content 上限固定为 `64 KiB`，要求 UTF-8、非空和 canonical serialization。内容被视为用户内容，可能包含业务敏感信息，因此不得进入日志、错误摘要、trace、URL、cursor、committed fixture 或 run record。

## 状态、并发与失败语义

- capture 只接受 `succeeded` terminal turn 和非空 run ref；running / failed / canceled / outcome unknown 均拒绝。
- 同一 scope / session / turn 只能创建一份 artifact；相同 run ref、profile、content type 与 digest 重试返回原记录，不同内容返回冲突。
- artifact 写入发生在 terminal turn 已成功落下之后。artifact store 失败不回滚 run，也不重放 provider；响应同时保留成功 turn 和稳定 artifact failure。
- v1 不允许客户端事后上传内容。首次未保存的易失结果在幂等重试时保持不可恢复。
- list cursor 绑定 tenant / workspace / application / owner / session 与最后一个排序键；scope 或 filter 漂移失败关闭。
- store unavailable、contract mismatch、not found、scope denied、payload invalid、content too large、source unavailable 与 conflict 使用稳定 failure code，不透传底层错误正文。

## 权限边界

当前读取和保存路径复用 parent session 权限，生命周期 mutation 使用独立权限：

- 新 turn 的显式 capture 复用 `application_sessions:execute`；只有执行者能在同一次请求中选择保存。
- list / read 复用 `application_sessions:read`，并继续要求 verified identity、active workspace membership 和精确 application scope。
- archive / unarchive 同时要求 `application_sessions:read` 与 `application_result_artifacts:archive`；`application_sessions:read|write|execute` 均不能替代后者。

Result Artifact 是独立数据 owner，但当前不新增平行的成员资格语义。若后续需要跨 session 分享、导出、删除或委派，必须独立评审权限，不得把 `read` 自动提升为分享或删除权限。

## 实施批次

### 批次 A：strict contract、memory owner 与 HTTP 纵向链（已完成）

- 实现 record / summary、校验、memory repository、每 turn 唯一性和严格 cursor。
- turn body 增加可选 `save_result`，默认 false；仅从 server-side canonical result capture。
- 覆盖 Workflow Definition、结构化 Definition、Application RAG、Prompt 与 Agent / Copilot 五类现有 session profile 的 canonical result serialization。
- 增加 session-scoped artifact list 与精确 read route；列表不返回 content。
- 响应区分 turn failure 与 artifact failure；幂等 retry 不执行 provider、不补建已丢失结果。
- 精准测试覆盖 scope、owner、大小、UTF-8、重复 capture、store unavailable、未知字段和内容不进入既有 session / run store。

当前实现已新增 `application_result_artifact.v1` / summary schema、独立 memory repository 和 service；五类 session profile 都只在成功 terminal turn 后从服务端 canonical response 捕获内容。turn body 的 `save_result` 默认 false；响应可分别表达 `result_artifact` 和 `result_artifact_failure_code`。session-scoped list 不返回 content，精确 read 设置 `Cache-Control: no-store`。幂等重试只读取已经存在的 artifact summary，不重新调用 Provider，也不从易失结果补建。

### 批次 B：SQLite / PostgreSQL 开发测试态 durable repository（已完成）

- 复用现有 local persistence runtime、Workflow PostgreSQL pool、selector 与 migration family，不新增 DSN 或连接池。
- 增加不可变 artifact 表、scope / session / turn 唯一键、严格 cursor 索引与运行角色权限。
- 覆盖 migration / rollback / reapply、并发、重启恢复、损坏 payload、no fallback 和敏感内容不进入诊断。

当前实现已在共享 Workflow Run Store migration family 顺序追加 SQLite `0022_application_result_artifacts` 与 PostgreSQL `0025_application_result_artifacts`，并由现有 run store backend selector 注入对应 repository。物理表使用完整 scope、artifact id 主键、每 session / turn 唯一键、session history cursor 索引、JSON / 关系列一致性约束和数据库级 update / delete 拒绝；没有新增组件、DSN、连接池或 memory fallback。五类 profile 的 session / turn 分属通用、Prompt 与 Agent 三组物理表，因此 artifact 表不伪造指向单表的外键；capture 前仍由 combined session repository 精确读取 terminal turn、profile 与 run ref，数据库只承接已经通过来源校验的不可变记录。

SQLite 真实文件证据覆盖创建、相同 turn 幂等、冲突、并发单创建、重启、跨 owner 隔离、不可变触发器、损坏 payload 和关闭后失败关闭。PostgreSQL 17 证据覆盖运行角色 DML / DDL 边界、并发收敛、重连恢复、不可变触发器、跨 owner 隔离及完整 migration rollback / reapply；`timestamptz` 投影按数据库微秒精度写入，读取仍与 JSON 中的 RFC3339Nano 时间做小于一微秒的严格一致性复验。所有 repository 错误继续只映射稳定 failure code，不包含正文或底层数据库详情。

### 批次 C1：生命周期后端纵向切片（已完成）

- 先冻结生命周期合同：不可变 artifact content / provenance 与可变生命周期状态必须分离；archive / unarchive 使用独立 `lifecycle_version`、expected-version CAS 和 append-only event，不修改 artifact payload，也不拆除现有数据库 update / delete 拒绝。
- list 默认只返回 active artifact；显式 archived filter、排序和 cursor 必须绑定 tenant / workspace / application / owner / session / lifecycle state。archive 不改变来源、digest、created at 或 read 权限，unarchive 只能恢复同一精确 owner 的原 artifact。
- 当前生命周期合同固定为 `application_result_artifact_lifecycle.v1`、`application_result_artifact_lifecycle_event.v1` 与带生命周期投影的 `application_result_artifact_summary.v2`。新资产与 `active v1` current state 必须在同一 repository 原子操作中创建；SQLite / PostgreSQL 迁移需为既有资产回填同一状态，但不得改写 `application_result_artifact.v1` payload。archive / unarchive 分别产生 `archived` / `unarchived` event，成功后生命周期版本加一。
- archive / unarchive 使用独立 `application_result_artifacts:archive` 权限，并继续要求 verified identity、active workspace membership、精确 application 与 owner scope；`application_sessions:read|write|execute` 均不能替代该权限。永久 purge 执行、自动 retention job、级联删除与批量清理继续关闭，只记录未来启用所需的权限、保留期、显式确认和关联资源规则。
- 生命周期路由固定为 `POST /v1/user-workspace/application-sessions/{session_id}/result-artifacts/{artifact_id}/archive` 与对应 `/unarchive`；body 只接受 `workspace_id`、`application_id` 和 `expected_lifecycle_version`。精确 read 继续允许读取同 owner 的 archived artifact，并同时返回生命周期投影；不存在 purge route。
- 生命周期 API、schema 或 migration 明确后更新具体任务卡，并在 memory、SQLite、PostgreSQL 同时落地 repository、route、CAS、并发、重启、no-fallback 与隐私证据，不能只交付 memory 或 Web 状态。

当前实现保持 `application_result_artifacts` 不可变表和 `application_result_artifact.v1` payload 原样，新增独立 current lifecycle 与 append-only event。新资产在同一 repository transaction 中创建 active v1；SQLite `0023_application_result_artifact_lifecycle` 与 PostgreSQL `0026_application_result_artifact_lifecycle` 会为既有资产回填 active v1。默认 list 只返回 active，显式 `lifecycle_state=archived` 才列出归档资产，cursor v2 绑定完整 scope、session、filter 与 limit；精确 read 仍允许同 owner 读取 archived content。

archive / unarchive route 已按 expected lifecycle version 执行 CAS，每次成功写入一条 `archived` / `unarchived` event；重复状态返回 state conflict，陈旧版本和并发失败返回 version conflict，并给出脱敏后的当前 lifecycle version / state。memory、SQLite 与 PostgreSQL 已覆盖原子创建、旧数据升级、并发单胜者、重启、跨 owner、损坏投影、运行角色、rollback / reapply 与 no-fallback。永久 purge route、自动 retention、级联删除和 production capability 均未建立。

### 批次 C2：共享应用工作区消费（已完成）

- 为现有 Application Interaction、Prompt Session 与 Agent Session 三类 surface 建立共享 strict artifact consumer，再接入保存选择、保存结果状态、active / archived metadata 列表、精确读取、session / run handoff 和 archive / unarchive；不得在三块页面复制 schema 解析、scope guard 或迟到响应状态。先按页面族判断是否需要局部 Pencil，不创建 S11。
- application、session、workspace、identity、profile 或路由切换必须清除已读取 content、一次性状态和迟到响应；不写 localStorage、sessionStorage、IndexedDB 或 URL。

当前实现以单一 `applicationResultArtifactConsumer.ts` 严格解析 summary v2、artifact v1、lifecycle v1 与 event v1，并以共享 `ApplicationResultArtifactPanel` 承接逐 turn 默认关闭的保存选择、active / archived metadata 列表、cursor、精确正文读取、Run handoff 与 expected-version archive / unarchive。三类 Session consumer 只负责各自 turn contract，并仅在用户明确选择时发送 `save_result=true`；执行成功与 artifact 保存失败继续独立表达。application、session、filter、artifact、generation 或 route scope 漂移会中止请求、清除正文并拒绝迟到响应。

页面设计审查五维评分为 `0 / 0 / 0 / 0 / 2 = 2`，采用 `C / 直接实现`：信息层级、风险表达和响应式模式均复用既有 S3 生命周期库与 S8 Session owner，只有三页共享复用带来新增杠杆，因此未修改 Pencil、未建立 S11。SQLite 本地产品已在真实浏览器完成 Prompt 与结构化 Workflow Application 的显式保存、metadata list、exact read 和 archive / unarchive；Agent 旧 assignment 继续以 `application_session_authority_changed` 在 Provider 调用前失败关闭。`1440×900`、`720×900`、`390×844` 无页面级横向溢出，控制台无 warning / error；Web `371/371` 与 production build 已通过。

2026-08-19 owner-local 跟进修正了共享 artifact 面板向 Run Review 的 exact handoff：目标 Run 存在时直接选中并读取 canonical detail；目标不存在或不在当前 Application scope 时显示明确的 unavailable 状态，不回退其它 Run，不 replay，也不为 fixture 合成 Run。根因是 `React.StrictMode` 重放 mount effect 时重复推进请求代际并清空首次有效 handoff；当前按稳定 owner-scope key 只初始化一次，仍保留 application / workspace / refresh / live scope 变化时的正常失效。

### 批次 D：双数据库产品连续链与专题收口（已完成）

- memory / SQLite / PostgreSQL 验证同一 profile matrix、幂等、权限、cursor、archive / unarchive、purge route 不存在与 no-fallback。
- SQLite 本地产品验证保存 → 刷新 → 精确读取 → 服务重启恢复 → archive → unarchive，并复核永久 purge 仍不可调用及桌面 / 窄屏行为。
- 同步 current focus、功能索引、能力矩阵、路线图与周志；专题关闭后不派生通用 result store 或 transcript 批次。

当前三种 store 已复用同一五 profile fixture 验证创建、幂等重放、精确读取与 active v1 初始生命周期；既有 HTTP 测试继续覆盖 strict cursor、组合权限、archive / unarchive 和 metadata-only list。PostgreSQL 17 聚合集成新增配置化整机产品链：平台通过共享 Workflow Run Store pool 创建 terminal source 与显式 artifact，验证相同 capture 幂等、归档、永久 `DELETE` route 返回 `405`、关闭连接后 store unavailable 且不回退 memory，并在重新构造完整 Server 后恢复原正文、digest 与 archived v2，再解除归档为 active v3。migration / runtime role 分离、并发、rollback / reapply 和五 profile repository 证据继续由既有聚合集成承载。

SQLite 本地产品在服务重新启动后恢复上一轮显式保存的 Prompt artifact `appres_oz5tcssa3ysdvdya`；metadata list 先恢复 active v3，精确读取返回同一正文与 `sha256:8c2a089ddbd14b8729422ffb86a7098785839d26d57c9566760f785f960a70bf`，随后完成 archive v4 → unarchive v5。当前 `1280×720` 页面无横向溢出且页面日志无 warning / error；批次 C2 的 `1440×900`、`720×900`、`390×844` 响应式证据继续成立。批次 D 未新增 API、schema、migration、Pencil、task card 或 checker。

## 验收方式

- provenance：客户端不能提交 content、digest、run ref 或 source profile；artifact 必须由同一次成功执行产生。
- default-off：`save_result` 省略或 false 时不访问 artifact repository，现有 response 和 metadata-only stores 不变。
- idempotency：同 client turn key 最多一次 provider、一个 terminal turn 和一个 artifact；重试不返回首次 transient result，但可返回已有 artifact summary。
- privacy：list、turn、run、日志、错误、cursor 和 committed 资产不包含 content；只有精确 read 返回 content。
- authorization：identity / membership / workspace / application / owner 任一不匹配均在 artifact read 或 capture 前失败关闭。
- compatibility：既有五类 session profile、Run History、Comparison、Evaluation、Gateway、RAG 与 HTTP Tool 测试不回归。
- repository：批次 A 已通过 Go 单元 / HTTP / race 精准测试、`go vet` 和仓库快速 / 全量门禁；批次 B 已通过 SQLite 真实文件专项、PostgreSQL 17 聚合集成、运行角色、并发、重启、rollback / reapply 与 no-fallback 证据。
- lifecycle：批次 C1 已通过 memory / SQLite / PostgreSQL CAS、旧资产 active v1 回填、HTTP 组合权限、metadata-only list、archived exact read、append-only event、配置化 PostgreSQL profile 与仓库门禁；批次 C2 已通过共享 strict consumer、三类 Session 接入、SQLite 真实保存 / exact read / archive / unarchive 和三视口浏览器复核；批次 D 已通过三存储五 profile matrix、PostgreSQL 配置化 Server 关闭 / 重启产品链、purge route 不存在、no-fallback，以及 SQLite Web 服务重启后的同一 artifact 恢复与 lifecycle 往返。专题证据链已闭合。

## 停止线

- 不默认保存，不持久化完整 transcript、用户 input、prompt、provider raw response、retrieval fragment 正文或 HTTP Tool response body。
- 不修改 `workflow_run_record.*`、`application_session.*` 的 metadata-only 内容策略，不把 artifact content 塞回 run / turn。
- 不允许客户端事后上传结果并绑定 run，不从日志、缓存或 provider 重建结果。
- 不实现 replay / resume、自动 retry / fallback、background execution、schedule、agent loop、业务写回或自动发布。
- 不打开真实 Provider、production secret、production auth、public sharing、public URL、跨 workspace 分享、billing 或 production capability。
- memory owner 不声明 durable 或重启恢复；SQLite / PostgreSQL 的 durable 结论只限开发测试态。所有后续工作必须保留 artifact payload 不可变，永久 purge、自动清理和 production lifecycle 继续关闭。
