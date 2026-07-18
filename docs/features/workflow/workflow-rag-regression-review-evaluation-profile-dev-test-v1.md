# Workflow RAG Regression Review 与 Evaluation Profile（开发 / 测试态）v1

更新时间：2026-07-18

状态：`workflow_rag_regression_review_evaluation_profile_dev_test_v1_completed`

## 功能目标

让内部开发者把已经完成的 `workflow_run_record.v3` 纳入 Run Comparison、Evaluation Cases、Baseline / Case Versioning 与 Evaluation Suite / Release Review，形成可复验的 RAG 回归审查链。该能力只读取既有不可变运行记录和 metadata-only retrieval evidence，不重新执行 retrieval、不调用 Gateway，也不把 query、fragment 正文、answer 或模型原始响应复制进比较、评测或 decision evidence。

本专题是独立功能，不是“Workflow RAG Retrieval 与应用知识快照 v1”的后续批次，也不改变该专题已完成状态。

## 用户路径

1. 用户在 Run History 中选择一个 RAG baseline run 和一个满足同一 retrieval binding 的 candidate run。
2. 服务端从同一 scoped run store 重读两份 v3 记录，校验不可变 snapshot、profile、query 和 retrieval node binding。
3. Comparison 展示状态变化、draft / provider / model 变化、selected fragment rank diff、citation 增减、候选数、context bytes 和 retrieval latency delta。
4. 用户把一个 baseline 和 1–20 个兼容 candidate 保存为既有 evaluation case，设置预期分类并按版本审查。
5. exact case version 可进入既有 suite；suite review 和 decision digest 消费 retrieval profile 的即时审查结果，不触发运行、发布或晋级。
6. SQLite 与 PostgreSQL 重启后，case、revision、suite 和 decision 仍可恢复并重新派生同一 RAG review。

## 可比性与分类

首版只接受：

- `workflow_run_record.v3` 与 `workflow_run_record.v3`；
- 相同 tenant / workspace / application scope；
- 相同 snapshot id / version / digest 与 `rag_ref`；
- 相同 retrieval profile id / version / digest；
- 相同 query digest / byte count；
- 相同 retrieval node id。

draft id / version / digest、provider、model 和运行状态允许变化，并作为审查字段返回。v3 与 v0 / v1 / v2 混合、不同 snapshot、profile、query 或 retrieval node 返回 `workflow_run_retrieval_profile_incompatible`，不得降级为 store contract mismatch。

确定性分类规则：

- baseline 成功而 candidate 非成功：`regression`；
- baseline 非成功而 candidate 成功：`improvement`；
- 双方均成功但 candidate 丢失 baseline citation ref：`regression`；
- 没有成功性反转或 citation loss，但 draft、provider、model、failure、fragment、rank、citation、候选数、context 或稳定节点证据变化：`changed`；
- 稳定字段一致：`unchanged`；
- 任一 run 仍为未 stale running：`inconclusive`。

retrieval latency 只作为观察证据，不单独判定回归。baseline citation 仅表示用户选中的基线证据，不被解释为业务真相或语义正确性裁决。

## Comparison 契约

旧 v0 / v1 比较继续返回 `workflow_run_comparison.v1`。RAG 比较返回 `workflow_run_comparison.v2`，在既有安全摘要外增加 `retrieval`：

- snapshot / profile / query / retrieval node binding；
- baseline / candidate retrieval attempt status；
- candidate count、selected count、citation count、context bytes 与 latency delta；
- fragment ref / digest / source type / official / baseline rank / candidate rank / change；
- citation added / removed refs；
- `evidence_changed`、`ranking_changed` 与 `citation_changed`。

响应不得包含原始 query、fragment 正文、fragment preview、完整 prompt、answer、credential、endpoint 或 provider raw envelope。Comparison 不调用 snapshot repository；授权 fragment preview 继续只属于普通 Run History detail。

## Evaluation、Baseline 与 Suite

既有 `workflow_evaluation_case.v2`、case revision、suite 和 release decision 持久化形状保持不变：它们已经只保存 run refs、预期分类和审计元数据。服务端在 create / revise 时按 baseline-candidate 对执行 v3 可比性校验；review 即时复用 Comparison v2。

review item 增加派生的 `run_profile=workflow_rag_retrieval.v1` 与 `comparison_schema_version=workflow_run_comparison.v2`。suite item 同样返回 run profile，review digest 将其纳入 canonical JSON，防止把 retrieval review 与普通零副作用 review 混为同一证据。

一个 case 内只能使用同一种可比 profile；一个 suite 可以聚合普通 case 与 retrieval case，因为每个 exact case version 的 profile 都显式进入 review item 和 digest。HTTP Tool v2 继续返回 `workflow_run_side_effect_profile_unsupported`。

## Store 与 migration

- `memory_dev`：复用现有 run / evaluation / suite store，保持有界 family 和完整 revision / decision trail。
- `sqlite_dev`：在 workflow run shared database 增加 `0006_workflow_evaluation_resources`，补齐 case、revision、suite 和 decision repository；selector 必须选择 SQLite 实现，不能回退进程内 memory。
- `postgres_dev_test`：复用既有 0003–0005 evaluation tables、同一 pool 和 selector；本专题不改变持久化 shape，因此不新增空迁移。

SQLite revision 与 decision 为 append-only；case / suite current projection 通过事务和 expected-version CAS 更新。任何 marker、schema、decode 或数据库错误都 fail closed，不返回部分列表，也不回退 memory。

## Web

Run History 允许选择 v3 baseline；candidate 选择按前端已读取的 metadata 预筛同 profile / snapshot / query binding，服务端仍是最终权威。Comparison 面板展示 retrieval binding、fragment rank diff、citation 增减和 metadata-only 停止线。

Evaluation Cases、Baseline revision 和 Suite 继续复用现有 lazy panel；RAG run 不再被过滤。创建与 revision 的不兼容组合在服务端拒绝，review 显示 profile、comparison schema 与 retrieval finding codes。默认 offline 保持零请求。

## 验收

- Go domain：v3 compatible / incompatible、citation loss、rank / evidence / context / latency diff、状态反转、running 和 metadata-only JSON。
- HTTP：scope、strict query、稳定 incompatible failure、v1 / v2 schema、禁止字段和旧 profile 回归。
- Evaluation：create、revision、baseline promotion、historical review、mixed case profile、suite digest 和 HTTP Tool v2 unsupported。
- Store：memory、SQLite、PostgreSQL 的 create / revise / review / suite / decision、重启恢复、scope、CAS、损坏记录和 no-fallback。
- Web：strict v2 consumer、v3 selection、retrieval diff、case / baseline / suite profile、offline 零请求、tests 与 build。
- 最终执行 Platform Go 精准 / 全包、PostgreSQL 专项门禁、Web test / build、仓库 fast / full，并关闭本批启动的服务和容器。

## 完成结果

- `workflow_run_comparison.v2` 已支持相同不可变 retrieval binding 的 v3 / v3 比较；citation loss 稳定归类为回归，fragment evidence、rank、citation、候选数、context 和 latency 只返回 metadata-only 差异。
- Evaluation create / revise / historical review 与 Suite review 已消费 `workflow_rag_retrieval.v1`，并把 comparison schema 和 run profile 纳入即时审查与 suite digest；HTTP Tool v2 仍明确 unsupported。
- SQLite workflow run shared database 已增加 `0006_workflow_evaluation_resources`，case / revision / suite / decision 具备事务 CAS、append-only 历史、重启恢复、损坏记录拒绝和 no-fallback；PostgreSQL 复用既有 0003–0005 表和同一 pool。
- Web Run History 已允许 v3 baseline，按 snapshot / profile / query / retrieval node 预筛 candidate；Comparison、Evaluation 与 Suite 面板显示 retrieval profile 和审查证据，严格 consumer 拒绝正文、query、prompt、answer、credential 和模型原始响应。
- 浏览器连续验收创建了同一 binding 的两条 v3 run、Evaluation case 和 exact-version Suite；Comparison 为 `unchanged`，case / suite review 为 `passed`。开发服务重启后仍恢复 5 条原有 run、case 与 suite，没有创建新 run，也没有重放 retrieval 或 Gateway。
- 验收中发现 Saved Draft digest 曾包含每次读取都会变化的 request audit 投影，导致同一精确草案被误判为 `draft_changed`；现已改为只散列 `draft_version` 与不可变 draft payload，并以单元测试固定读时 audit / validation 投影不得影响 digest。修正前的开发态历史 run 保持不可变，不进行记录重写。
- Go 精准测试、PostgreSQL 专项门禁、Web tests / build 与仓库级门禁均作为本专题完成证据；生产能力、外部检索源、自动 baseline 和自动 release 保持关闭。

## 停止线

- 不重新执行 retrieval 或 Gateway，不实现 retry、replay、resume、scheduled execution 或 background evaluation。
- 不读取或保存 fragment 正文、query、answer、prompt packet、credential 或模型原始响应。
- 不支持 HTTP Tool v2 副作用比较；该能力需要独立设计。
- 不接 crawler、online search、connector、vector database、embedding、reranker 或外部 provider。
- 不开放多 retrieval、复杂图、agent loop、writeback、publish、自动 baseline、自动 release 或 production enablement。
