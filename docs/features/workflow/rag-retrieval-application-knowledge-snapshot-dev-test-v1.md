# Workflow RAG Retrieval 与应用知识快照（开发 / 测试态）v1

更新时间：2026-07-17

状态：`workflow_rag_retrieval_application_knowledge_snapshot_dev_test_v1_design_defined_pending_review`

## 设计结论

本专题定义 Workflow 首个可执行 RAG 纵向切片：内部开发者先在当前应用作用域内创建版本化、不可变的知识快照，再让精确 Saved Draft 通过一个 `rag_retrieval` 节点引用精确快照版本；操作者显式启动 retrieval execution 后，服务端以确定性的本地 lexical profile 完成一次检索，把有界片段交给既有 Gateway LLM，并要求结构化回答只引用本次实际召回的片段，最后写入 `workflow_run_record.v3` 与 Run History。

首版必须同时覆盖知识资产生命周期、检索执行、引用约束、三种开发测试态 store 和 Web 连续链，不能把预置 fixture、请求内塞入的“已召回文本”或单纯打开 `allow_retrieval=true` 当成真实 RAG 完成。

本设计当前等待边界评审。评审通过后只创建一张高风险实施任务卡，按依赖拆三个实现批次；评审前不新增 schema、route、migration、runtime provider 或 Web 写路径。

## 用户价值与适用范围

目标用户是需要为内部 AI 应用建立可审查知识问答工作流的开发者。首版解决一条完整路径：

1. 在当前 tenant / workspace / application 下提交一组已审查的知识片段，创建不可变知识快照版本。
2. 在 Draft Designer 中配置 `prompt -> rag_retrieval -> llm -> output`，`rag_ref` 精确固定快照 key 与版本。
3. 保存草案并显式启动 retrieval execution；服务端重新读取精确草案和知识快照，不信任客户端回传内容、排名或引用。
4. 本地确定性检索器只在该快照内检索一次，按预算选出片段并构造带稳定 `fragment_ref` 的上下文。
5. LLM 返回结构化回答、引用和限制项；服务端拒绝不存在、跨快照或未实际召回的引用。
6. Run History 展示快照版本、profile 版本、召回数量、来源类型、引用、失败边界和 metadata-only 诊断。

本专题只属于开发 / 测试态，不提供生产知识库、外部 crawler、在线搜索、向量数据库、embedding provider、生产认证或公开生产 API。

## 现有基础与兼容边界

以下能力可以复用，但都不等于 RAG runtime 已存在：

- Saved Draft 已允许保存 `rag_retrieval` 节点、节点级 `rag_ref` 和聚合 `rag_refs`，Draft Designer / Review Handoff 能展示这些字段。
- `CopilotRequest.artifacts[*].metadata` 已稳定支持 `source_type`、`page_slug`、`fragment_id`、`retrieval_rank` 与 `is_official`，可作为来源元数据语义参考。
- Radish docs QA 已有 official source precedence、retrieval scope 和引用评测经验，但其离线样本与 artifact 输入不是 Workflow 知识仓库或运行时检索器。
- `radish.docs.retrieval_context.v1` 属于 Session / Tooling v1 的 `contract_only` registry entry，`execution_enabled=false`；不得直接改成 Workflow executor。
- Workflow Executor v0 继续强制 `allow_retrieval=false`，只接受 Prompt / LLM / condition / output；既有 `POST .../runs`、`workflow_run_record.v0/v1` 和行为测试保持兼容冻结。
- HTTP Tool 使用独立 action plan / confirmation 与 `workflow_run_record.v2`；RAG 不伪造工具调用、确认决定或 HTTP attempt，也不复用 v2。

新增 RAG 路径必须版本化，并在功能完成后才更新现有 workflow boundary fixture。旧 `rag_ref` 占位不能自动迁移成可执行快照引用。

## 权威来源与所有权

首版将知识快照视为 RadishMind 应用运行资产，不是上层项目业务真相源：

| 层 | 唯一职责 | 不得承担 |
| --- | --- | --- |
| 上游来源 | 拥有原始文档、Wiki、FAQ、论坛内容或人工片段 | 不由 RadishMind 修改或同步 |
| 应用知识快照 | 保存操作者明确提交、已分类、不可变的检索片段版本 | 不声明自己是上游最新真相，不自动抓取 |
| retrieval profile | 固定排名算法、token / n-gram 规则、预算和输出契约 | 不保存应用内容，不选择跨 scope 数据 |
| retrieval execution | 读取精确草案与快照，执行一次检索和后续 LLM | 不写快照、不改变来源、不重放 |
| run store | 保存 metadata-only 运行、引用和诊断证据 | 不复制片段正文、原始 query 或完整 prompt packet |

上游内容变化必须由操作者显式创建新快照版本，再更新草案 `rag_ref` 并保存新草案版本。旧 run 继续绑定旧快照版本，不做静默漂移。

## 版本化契约

评审通过后的实施只物化以下契约：

- `workflow_rag_snapshot.v1`：完整 scope、snapshot key / version、状态、内容分类、fragment 数量、digest、创建者与时间。
- `workflow_rag_fragment.v1`：稳定 `fragment_ref`、来源类型、来源引用、标题、片段正文、排序辅助元数据与内容 digest。
- `workflow_rag_execution_profile.v1`：算法 id / version、规范化策略、`top_k`、query / fragment / context 预算和环境 enablement。
- `workflow_rag_answer.v1`：`answer`、`citations`、`limitations` 与 `confidence`；引用只能指向本次召回集合。
- `workflow_rag_execution_audit.v1`：快照生命周期和检索执行的 metadata-only 审计。
- `workflow_run_record.v3`：一个 retrieval attempt、零 tool / confirmation / business write / replay，以及结构化引用证据。

`rag_ref` 固定为 `workflow.rag.<snapshot_key>.v<snapshot_version>`，只允许小写稳定短键和正整数版本。草案必须同时满足：

- 恰好一个 `rag_retrieval` 节点；
- 节点 `rag_ref` 与顶层 `rag_refs` 精确一致；
- 同 scope 下存在该精确快照版本；
- 快照 digest、profile digest 和草案版本在执行时重新校验；
- 不接受 alias、latest、前缀匹配、客户端片段或客户端排名。

## 应用知识快照生命周期

快照资源使用 append-only version 语义：

1. 创建 snapshot key 时同时创建版本 1。
2. 内容更新必须提交完整新版本，并带 `expected_latest_version` 做 CAS；旧版本不可修改。
3. archive 只阻止创建新版本和新 execution；已存在 run 仍可在授权下读取其绑定版本的 fragment evidence。
4. 首版不提供 unarchive、delete、跨应用复制、自动同步或后台 re-index。
5. snapshot list 只返回 metadata summary；detail 在独立 read scope 下返回精确版本与有界 fragment 内容。

每个 fragment 至少包含：

- `fragment_ref`：快照版本内唯一稳定短键；
- `source_type`：`document / wiki / faq / forum / manual` 之一；
- `source_ref`：上游稳定引用，不允许 credential-bearing URL；
- `page_slug`、可选标题、`is_official`；
- UTF-8 正文与内容 digest；
- `content_classification`：`public` 或 `workspace_internal`。

快照不能包含 token、cookie、Authorization header、credential、DSN、私钥或被策略明确识别的密钥材料。API、日志、审计和 run record 不回显正文；只有具有 snapshot read scope 的 detail / Web 管理面可以读取授权片段。

## 首版检索器与长期适配边界

首版 profile 固定为 `workflow.rag.lexical-ngram-dev.v1`：

- 只读取当前快照已持久化片段，不访问网络、文件系统、connector 或上层项目。
- 拉丁文本使用规范化 word token；CJK 文本使用 Unicode 字符 bigram，并保留稳定停用规则版本。
- 排名使用版本化的确定性 BM25-like scorer；同分按 `fragment_ref` 升序，不能依赖数据库未定义顺序。
- query 只来自本次显式 `input_text` 经 Prompt 节点规范化后的有界文本；LLM 不能先生成检索 query。
- 无自动 query expansion、reranker、fallback、retry、embedding 或 hybrid search。
- provider interface 必须把 ranking algorithm 与 store 分离，使未来 vector / embedding adapter 可以作为独立版本加入，而不改写 v1 记录。

该 lexical profile 是可复验的开发测试态真实检索器，不是 production relevance 声明。未来接入向量数据库、embedding provider、外部搜索或 connector 必须另做凭据、网络、成本、数据驻留和删除策略评审。

## 拓扑、执行与预算

首版只接受：

```text
prompt root -> one rag_retrieval -> one llm -> output terminal
```

- 恰好四个节点和三条边，不允许 condition、HTTP Tool、并行、循环、多 retrieval、多个 LLM 或后台运行。
- `rag_retrieval` 风险固定为 `low`、`requires_confirmation=false`；内容分类或 scope 不满足时直接阻塞，不通过人工确认绕过。
- retrieval execution 使用独立 `POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions`，不放宽 executor v0 的 `/runs`。
- 服务端从精确 Saved Draft 重建执行计划，先创建 run v3 running 记录，再执行一次本地 retrieval；pre-run store 失败时 retrieval / provider 调用均为 0。
- 检索成功后，片段正文只在有界内存中组成 LLM packet；run record 只保存 fragment refs、digests、排名、来源类型与引用。

预算固定为：

| 项目 | v1 上限 |
| --- | ---: |
| 单快照 fragment 数 | 256 |
| 单 fragment 正文 | 8 KiB |
| 单快照正文总量 | 1 MiB |
| query | 4 KiB |
| `top_k` | 默认 5，最大 8 |
| 单个入模片段 | 4 KiB |
| 总 retrieval context | 32 KiB |
| retrieval 本地执行 | 2 秒 |
| 整条同步 run | 30 秒 |

超限必须稳定失败，不截断成语义不明的伪成功。只有单个入模片段可以按 profile 固定规则生成明确 `excerpt_truncated=true` 的有界摘录，原 fragment digest 保持不变。

## 结构化回答与引用约束

LLM 输出必须解析为 `workflow_rag_answer.v1`：

- `answer` 是 advisory-only 文本；
- `citations` 至少包含一个 `{fragment_ref, claim_summary}`；
- 每个 `fragment_ref` 必须属于本次 selected fragments；
- `limitations` 显式表达无匹配、证据冲突或覆盖不足；
- `confidence` 只允许 `low / medium / high`，不作为业务决策授权。

未知引用、跨快照引用、未召回引用、重复引用、空回答或无法解析的模型输出都进入稳定失败。来源优先级沿用 Radish docs QA 的可复用原则：`is_official=true` 的正式来源优先，FAQ / forum 只能作为补充；但具体来源权重属于 profile 版本，不写死在 Web。

## API 与权限设计

首版资源族固定为：

```text
POST /v1/user-workspace/workflow-retrieval-snapshots
GET  /v1/user-workspace/workflow-retrieval-snapshots
GET  /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}
POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/versions
POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/archive
POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions
```

权限拆分为：

- snapshot create / version：`workflow_rag_snapshots:write`；
- snapshot list / detail：`workflow_rag_snapshots:read`；
- archive：`workflow_rag_snapshots:archive`；
- execute：`workflow_rag:execute`、`workflow_runs:execute`、`workflow_drafts:read` 与 `workflow_rag_snapshots:read`。

所有 route 继续要求 verified actor、tenant / workspace / application binding、strict JSON、request id、audit ref 和独立开发门禁。创建与执行授权分离；能够运行 Workflow 不等于能够写知识快照。

## Store 与事务边界

snapshot、version、fragment 与 snapshot audit 归入新的 Workflow retrieval repository owner，但必须复用现有 workflow runtime backend mode、SQLite shared database、PostgreSQL pool、配置层级和 no-fallback 语义：

- `memory_dev` 使用有界 owner store；
- `sqlite_dev` 在 shared database 中追加 migration；
- `postgres_dev_test` 在 workflow runtime migration 链中追加 migration；
- selector、marker 或 migration 不匹配时 fail closed，不回退 memory。

run v3 继续写入既有 workflow run store。snapshot 创建 / 新版本 / archive 各自在单库事务内完成资源、fragment 与 audit；retrieval execution 不锁住 snapshot 内容，只读取不可变精确版本。run 创建成功但本地检索或 LLM 失败时写真实失败终态，不 retry；进程崩溃继续复用 metadata-only reconciliation，不重做 retrieval。

## Run History、比较与评测边界

`workflow_run_record.v3` 至少记录：

- snapshot id / version / digest、profile id / version / digest；
- retrieval node id、attempt status、query digest 与 query byte count；
- candidate count、selected fragment refs / digests / ranks / source types；
- retrieval latency、context bytes、citation refs；
- `retrieval_calls=1`、`tool_calls=0`、`confirmation_calls=0`、`business_writes=0`、`replay_writes=0`；
- provider / model、节点状态与结构化失败诊断。

Run History summary 不返回 fragment 正文或原始 query。detail 可在 snapshot read 授权下按 run 绑定的 fragment refs 读取最多 512 字符的片段预览，并明确其来自不可变 snapshot repository，不从 run record 读取复制内容。

首版 Run Comparison、Evaluation Cases、Baseline 和 Suite 对 v3 返回 `workflow_run_retrieval_profile_unsupported`；不得误报 store contract mismatch。实现批次必须先以确定性 ranking fixture、scope negative、引用 validator 和三种 store 行为测试建立评测基线，再决定后续独立的 RAG regression review 功能。

## 失败码与审计

稳定失败至少包括：

| failure code | 边界 |
| --- | --- |
| `workflow_rag_snapshot_not_found` | 精确快照不存在或不可见 |
| `workflow_rag_snapshot_scope_denied` | tenant / workspace / application 不匹配 |
| `workflow_rag_snapshot_archived` | archive 后禁止新 execution |
| `workflow_rag_snapshot_version_conflict` | 新版本 CAS 失败 |
| `workflow_rag_fragment_invalid` | fragment schema、分类、大小或禁止内容不合法 |
| `workflow_rag_profile_disabled` | profile 未登记、disabled 或 digest 漂移 |
| `workflow_rag_draft_ineligible` | 草案拓扑、`rag_ref` 或版本不满足 |
| `workflow_rag_query_invalid` | query 为空、超限或无法规范化 |
| `workflow_rag_budget_exceeded` | snapshot、检索或 context 预算超限 |
| `workflow_rag_no_evidence` | 没有满足阈值的片段 |
| `workflow_rag_retrieval_failed` | 本地排名器确定失败 |
| `workflow_rag_answer_invalid` | LLM 输出不满足结构化回答契约 |
| `workflow_rag_citation_invalid` | 引用未知、跨快照或未被召回 |
| `workflow_rag_store_unavailable` | repository / run store 不可用且不回退 |

诊断边界增加 `retrieval_policy / retrieval_store / retrieval_rank / retrieval_context / retrieval_citation`，并使用独立 `retrieval_failure_category`。审计事件包括 snapshot created / versioned / archived、retrieval started / succeeded / failed；审计不保存 fragment 正文、query、prompt packet 或模型原始响应。

## 实施拆分与完成定义

设计评审通过后，一张任务卡内按依赖实施：

1. 批次 A：版本化 snapshot / fragment / profile / answer / audit / run v3 契约，snapshot 生命周期 API，确定性 lexical provider，memory / SQLite / PostgreSQL repository 与 Web snapshot 管理；本批不启动 Workflow retrieval run。
2. 批次 B：独立 retrieval execution service / route、run v3、结构化 answer / citation validator、Gateway handoff、诊断、reconciliation 和三种 store 一致性。
3. 批次 C：Draft Designer 精确 `rag_ref` 选择、显式 execution、Run History v3、SQLite / PostgreSQL 连续链、真实浏览器创建快照到重启恢复，以及正文 / query / credential 泄漏审计。

只有三个批次全部通过，才能声明 `workflow_rag_retrieval_application_knowledge_snapshot_dev_test_v1_completed`。每批优先扩 Go / TypeScript 行为测试和既有聚合门禁；只有现有证据无法承载新的 schema / 执行边界时才增加一个专项 contract checker，不派生 readiness 链。

## 待评审决策

边界评审需要一次性确认：

1. 首版把“应用知识快照生命周期”纳入 RAG 纵向切片，而不是依赖预置 fixture 或客户端传已召回文本。
2. 首版只启用本地确定性 `lexical-ngram-dev.v1`，向量 / embedding / external search 延后独立评审。
3. 使用独立 retrieval execution route 与 `workflow_run_record.v3`，保持 executor v0 和 HTTP Tool v2 冻结。
4. fragment 正文只存于授权 snapshot repository；run、audit、日志和普通 history 均保持 metadata-only。
5. 设计通过后只创建一张任务卡，按 A / B / C 三批交付，不新增同层静态治理链。

## 停止线

- 不启用 crawler、任意 URL、在线搜索、文件系统扫描、attachment reader、connector 或跨应用检索。
- 不接 vector database、embedding provider、reranker、外部 credential、production secret 或自动 fallback。
- 不允许客户端、LLM 或 Web 提交排名结果、selected fragments、source 权重或最终引用白名单。
- 不修改上游文档、Wiki、FAQ、论坛或业务真相源，不把 snapshot 当上游最新真相。
- 不打开多 retrieval、并行、循环、agent loop、background run、replay、resume、writeback 或 publish。
- 不把开发测试态 lexical relevance、SQLite / PostgreSQL store 或浏览器闭环写成 production RAG ready。
- 不在边界评审通过前创建实现任务卡、contract artifact、migration、API 或 runtime provider。
