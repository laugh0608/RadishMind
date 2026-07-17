# Workflow RAG Retrieval 与应用知识快照（开发 / 测试态）v1 实施任务卡

更新时间：2026-07-17

- 任务 ID：`workflow-rag-retrieval-application-knowledge-snapshot-dev-test-v1`
- 状态：`workflow_rag_retrieval_application_knowledge_snapshot_dev_test_v1_batch_a_completed`
- 功能设计：[Workflow RAG Retrieval 与应用知识快照（开发 / 测试态）v1](../features/workflow/rag-retrieval-application-knowledge-snapshot-dev-test-v1.md)

## 准入结论

2026-07-17 边界评审已确认功能设计中的五项决策：首版包含应用知识快照生命周期；只启用本地确定性 lexical profile；使用独立 retrieval execution 与 `workflow_run_record.v3`；fragment 正文只存 snapshot repository；全部实现由本卡按 A / B / C 三批承接。

本卡是该功能唯一高风险实现入口。批次 A 先物化契约、知识快照资源、deterministic lexical provider、三种 store 和 Web 管理，但不得注册 retrieval execution 或创建 run v3；批次 B 再接 execution / Gateway / citation / run v3；批次 C 最后完成 Draft Designer、Run History、双数据库与浏览器重启验收。

## 不可变架构边界

1. 知识快照是应用作用域的 RadishMind 运行资产，不是上层业务真相；版本 append-only，更新只创建新版本。
2. 首版 `rag_ref` 固定为 `workflow.rag.<snapshot_key>.v<version>`，不接受 alias、latest、客户端内容或客户端排名。
3. `workflow.rag.lexical-ngram-dev.v1` 只读取当前精确快照，不访问网络、文件、connector、vector database 或 embedding provider。
4. Executor v0、`POST .../runs`、`workflow_run_record.v0/v1`、HTTP Tool action plan 与 run v2 保持冻结；RAG 以后只进入独立 retrieval execution / run v3。
5. snapshot repository 复用 workflow runtime backend mode、SQLite shared database 与 PostgreSQL pool；selector / marker / migration 不匹配必须 fail closed，不回退 memory。
6. snapshot detail 是唯一正文读面；list、audit、日志、run 和普通 history 都是 metadata-only。

## 版本化契约与稳定标识

批次 A 物化：

| 契约 | 文件 | 批次 A runtime |
| --- | --- | --- |
| `workflow_rag_fragment.v1` | `contracts/workflow-rag-fragment.schema.json` | snapshot write / detail |
| `workflow_rag_snapshot.v1` | `contracts/workflow-rag-snapshot.schema.json` | create / version / list / detail / archive |
| `workflow_rag_execution_profile.v1` | `contracts/workflow-rag-execution-profile.schema.json` | deterministic lexical provider contract |
| `workflow_rag_answer.v1` | `contracts/workflow-rag-answer.schema.json` | validator only，批次 B 消费 |
| `workflow_rag_execution_audit.v1` | `contracts/workflow-rag-execution-audit.schema.json` | snapshot lifecycle audit；execution audit 留 B |
| `workflow_run_record.v3` | `contracts/workflow-run-record-v3.schema.json` | validator / compatibility only，不创建 run |

snapshot key、fragment ref 与 profile id 都使用稳定短键。snapshot digest 使用 canonical JSON + SHA-256，覆盖完整 scope、版本、classification、profile ref 与按 `fragment_ref` 排序的 fragment metadata / content digest，不覆盖 request id、actor 展示名或时间戳。

## 批次 A：契约、知识快照与确定性检索基础

状态：`completed`。

### 实现范围

1. 物化六份 schema、positive / negative contract tests 和 Go validator；run v3 只证明可解析与 v0/v1/v2 隔离。
2. 实现 snapshot create / list / detail / version / archive domain service，固定 scope、strict validation、CAS、append-only version、stable digest 与 metadata-only summary。
3. 实现 `lexical-ngram-dev.v1` provider：Latin word token、CJK bigram、稳定 BM25-like score、official source boost 和 `fragment_ref` tie-break；只提供纯函数 ranking API，不接 Workflow execution。
4. memory store 使用有界 owner lock；SQLite 在 `sqlite/workflow_runs/0004_workflow_rag_snapshots` 追加表；PostgreSQL 在 `workflow_runs/0008_workflow_rag_snapshots` 追加 migration。snapshot、fragment、audit 必须同事务提交。
5. 注册 snapshot 五条 route，独立门禁为 `RADISHMIND_WORKFLOW_RAG_SNAPSHOT_DEV=1`；沿用 verified dev auth，并要求 `workflow_rag_snapshots:write/read/archive`。
6. Web source 固定为 `VITE_RADISHMIND_WORKFLOW_RAG_SOURCE=dev-workflow-rag-http`，提供 application-scoped snapshot create / list / detail / version / archive 管理面；默认 offline 和缺 scope 时请求数为 0。
7. 批次 A 不注册 `POST .../retrieval-executions`，不调用 Gateway，不创建 run，不改变 `allow_retrieval=false`。

### API 形状

```text
POST /v1/user-workspace/workflow-retrieval-snapshots
GET  /v1/user-workspace/workflow-retrieval-snapshots
GET  /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}
POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/versions
POST /v1/user-workspace/workflow-retrieval-snapshots/{snapshot_id}/archive
```

- create：`workspace_id`、`application_id`、`snapshot_key`、`display_name`、`content_classification`、`fragments`。
- version：完整 scope、`expected_latest_version` 与完整 replacement fragments；不接受 patch。
- archive：完整 scope 与 `expected_latest_version`。
- detail：query 指定 scope 与 `snapshot_version`；不提供 latest 隐式读取。
- list：只返回 latest metadata summary，支持稳定 `snapshot_key` 顺序，不返回 fragment 正文。

### Migration 与 store marker

- PostgreSQL：`0008_workflow_rag_snapshots`，schema version 推进为 `workflow_run_store_v8`。
- SQLite：`0004_workflow_rag_snapshots`，schema version 推进为 `workflow_run_store_sqlite_v4`。
- 新表至少包含 snapshot resources、immutable versions、fragments 与 append-only audits；运行角色禁止 UPDATE / DELETE immutable version / fragment / audit，只允许 resource latest-version / archive CAS。
- migration / rollback / reapply、checksum、marker、运行角色、scope、重启与 no-fallback 必须复用现有 workflow runtime runner。

### 完成门禁

- 六份 schema 与 Go validator 覆盖 unknown field、scope drift、digest、预算、forbidden material 和 v0/v1/v2/v3 隔离。
- 三种 store 对 create / version / list / detail / archive、并发 version / archive CAS、跨 scope、restart 与 no-fallback 结果一致。
- lexical provider 覆盖 Latin、CJK、official priority、同分 tie-break、空 query、无证据和预算边界，排序不依赖数据库。
- Web consumer 严格拒绝正文出现在 list / audit，detail 只在 read scope 下显示有界授权内容。
- network、Gateway provider、workflow run、tool、confirmation、business write 与 replay 全部为 0。
- 完成锚点固定为 `workflow_rag_retrieval_application_knowledge_snapshot_dev_test_v1_batch_a_completed`。

### 批次 A 完成证据

- 六份 JSON Schema 与 Go strict validator 已物化；snapshot / fragment / profile / answer / audit / run v3 正负契约测试通过，run v3 仍为 contract-only。
- snapshot create / list / exact detail / full-replacement version / archive 已覆盖 strict JSON、三层 scope、CAS、stable digest、secret / budget 拒绝和 metadata-only list；memory、SQLite、PostgreSQL 行为一致。
- SQLite `0004_workflow_rag_snapshots` 与 `workflow_run_store_sqlite_v4`、PostgreSQL `0008_workflow_rag_snapshots` 与 `workflow_run_store_v8` 已通过 migration、append-only、运行角色、重启和 no-fallback；PostgreSQL v8 checksum 为 `sha256:e5d44074f395fb5630189f4ddc7d6b027fd941742c88334a28df9316ee2b463a`。
- Web application-scoped 管理面已覆盖 offline / 缺 scope 零请求、metadata-only list、精确正文 detail、完整 replacement CAS 和 archive；Web 102 项测试与生产构建通过。
- Batch A 的 network、Gateway provider、workflow run、tool、confirmation、business write 与 replay 保持为 0；没有注册 `retrieval-executions`，`allow_retrieval=false` 未改变。

## 批次 B：Retrieval Execution、结构化引用与 Run v3

状态：`ready_for_implementation`。

1. 注册独立 `POST /v1/user-workspace/workflow-drafts/{draft_id}/retrieval-executions` 和独立 execution gate。
2. 重读精确 Saved Draft、snapshot、profile 与 digest，接受唯一四节点三边拓扑。
3. 原子创建 run v3 running，执行一次 lexical retrieval，把有界片段交给 Gateway，再校验 `workflow_rag_answer.v1` 与 citation subset。
4. 写入 metadata-only selected refs / ranks / digests、诊断与终态；无 retry / resume / replay。
5. 扩展 Run History、comparison / evaluation 显式 unsupported、reconciliation 和三种 store。

## 批次 C：Web 纵向链与双数据库浏览器验收

状态：`blocked_by_batch_b`。

1. Draft Designer 选择精确 snapshot version，保存合法 `rag_ref`，与 offline placeholder 区分。
2. Web 显式启动 retrieval execution，并在 Run History v3 展示 snapshot / retrieval / citation metadata。
3. 完成 SQLite 创建快照 → 保存草案 → 执行 → 历史 → archive / new version → 重启恢复连续链。
4. 完成 PostgreSQL v8 migration、角色、并发、marker、重启与 no-fallback 专项。
5. 真实浏览器检查 list / run / storage / logs 不泄漏 query、fragment 正文、prompt packet、credential 或模型原始响应。

## 主要实现落点

- 契约：`contracts/workflow-rag-*.schema.json`、`contracts/workflow-run-record-v3.schema.json`。
- Platform：`services/platform/internal/httpapi/workflow_rag_*.go` 与 `server.go`。
- Store：现有 workflow runtime selector / SQLite shared database / PostgreSQL pool 与 migration runner。
- Web：`apps/radishmind-web/src/features/control-plane-read/workflowRAGSnapshotConsumer.ts`、`workflowRAGSnapshotPanel.tsx` 与对应测试。
- 文档：功能专题、当前焦点、Workflow 入口、任务卡索引和 W29 周志。

## 必要验证

- 每个批次至少执行 Go 精准测试、Web tests / build、`git diff --check` 与 `./scripts/check-repo.sh --fast`。
- 批次 A / B 的 schema、migration、store 或执行边界变化，以及批次 C 的专题关闭，都补跑完整 `./scripts/check-repo.sh`。
- PostgreSQL 专项继续使用 `./scripts/run-workflow-saved-draft-postgres-dev-test.sh check`，不得用 SQLite 代替。

## 停止线

- 不启用 crawler、任意 URL、文件扫描、attachment reader、connector、vector database、embedding provider 或 reranker。
- 不允许客户端提交 selected fragments、排名、source weight、引用白名单、profile digest 或 snapshot digest。
- 不修改上层业务真相，不创建生产知识库、production repository、production secret、公开生产 API、quota 或 billing。
- 不打开多 retrieval、condition、HTTP Tool 混合图、并行、循环、agent loop、background、replay、resume、writeback 或 publish。
- 不把批次 A 的 schema / snapshot store / lexical provider 写成 RAG execution 已完成。
