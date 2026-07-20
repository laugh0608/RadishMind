# Workflow RAG 评测数据集应用资源化与候选快照审查 v1 实施任务卡

更新时间：2026-07-18

状态：`workflow_rag_evaluation_dataset_application_resource_candidate_snapshot_review_v1_completed`

## 目标

按[功能设计](../features/workflow/workflow-rag-evaluation-dataset-application-resource-candidate-snapshot-review-v1.md)交付应用作用域 durable dataset resource、精确 baseline / candidate snapshot lexical review 和 Web 审查路径。复用 workflow backend selector、shared SQLite database、PostgreSQL pool、snapshot repository 与现有 ranker；不调用 Gateway、不创建 run、不自动晋级任何资源。

## 批次 A：契约与内存纵向链

状态：已完成。

- 新增 dataset resource / candidate review 契约。
- 增加独立 `RADISHMIND_WORKFLOW_RAG_EVALUATION_DEV=1` gate、verified actor 与权限矩阵。
- 实现 strict create / list / exact read / version / archive / candidate review API。
- 实现 classification、binding、digest、CAS、metadata-only comparison 和 memory store。

## 批次 B：durable store

状态：已完成。

- SQLite shared database 新增 dataset resource / immutable version / review / audit migration。
- PostgreSQL workflow run store 新增对应 migration、manifest、checksum 与 pending upgrade。
- 覆盖事务 CAS、append-only、scope、损坏记录、重启恢复和 selector no-fallback。

## 批次 C：Web 与收口

状态：已完成。

- 新增 strict Web consumer 和 lazy panel，默认 offline 零请求。
- 覆盖 dataset lifecycle、exact content edit、candidate review、CAS conflict、敏感字段拒绝。
- 更新 current focus、Workflow 入口、能力矩阵、roadmap 和 W29 周志。
- 执行 Go 精准 / 全包、PostgreSQL 专项、Web tests / build、`git diff --check`、仓库 fast / full。

## 停止线

- 不进入 connector、online search、embedding、vector database、reranker、Gateway、run、自动 baseline / release 或 production enablement。
- 不启动真实浏览器验收；本专题 Web 以 consumer / component tests 与 build 收口。
- 不新增第二张 task card、同层 readiness 文档或平行 store / checker 基础设施。

## 完成结论

三批已按同一任务卡关闭；memory、SQLite 与 PostgreSQL 行为一致，PostgreSQL 专项、Platform 全包、Web tests / build 和仓库聚合门禁均作为提交前证据复验。完成锚点为 `workflow_rag_evaluation_dataset_application_resource_candidate_snapshot_review_v1_completed`，不派生批次 D。
