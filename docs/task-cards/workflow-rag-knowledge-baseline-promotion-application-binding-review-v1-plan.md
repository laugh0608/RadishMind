# Workflow RAG 知识基线晋级与应用配置绑定审查 v1 实施任务卡

更新时间：2026-07-18

状态：`workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_ready_for_implementation`

## 目标与准入结论

按 [功能设计](../features/workflow/workflow-rag-knowledge-baseline-promotion-application-binding-review-v1.md)交付“精确 dataset / candidate review / snapshot / profile / source draft → 人工 decision → 不可变 RAG binding → 应用配置草案引用 → 发布治理重校验”的完整开发测试态链路。

设计与边界评审已经通过，批次 A 获得实现准入。当前尚未实现 contract、Go / TypeScript、schema migration、API runtime 或 Web 接线；本任务卡是唯一实现入口，不派生第二张任务卡或同层 readiness 文档。

## 前置基线

- `dev` 分支保持可复验，既有 snapshot、durable evaluation dataset / candidate review、application configuration draft 与 publish governance 测试通过。
- promotion store 必须由现有 `workflowRunStore` backend selector 派生；应用草案与发布候选继续使用各自既有 repository。
- 独立 `RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV=1` gate 默认关闭；未完成所需 migration / marker 时服务启动失败，不回退 memory。
- 所有请求、响应、store payload、audit 和日志遵守 metadata-only / secret guard；不保存 query、fragment 正文、模型响应或 prompt。

## 批次 A：contract、领域与 memory 纵向链

状态：`ready_for_implementation`。

- 物化 `workflow_rag_knowledge_promotion_candidate.v1`、`workflow_rag_knowledge_promotion_decision.v1`、`workflow_rag_application_binding.v1` 与 audit contract。
- 实现服务端重读 dataset version / candidate review / baseline + candidate snapshots / lexical profile / source draft / application baseline。
- 实现 review eligibility、candidate digest、`pending / deferred / approved / rejected / canceled` 状态机、`expected_record_version` CAS、append-only decision / audit，以及 approve 原子创建 immutable binding。
- 注册四条 strict dev-only API、独立权限组合、稳定 failure envelope 和 list / detail metadata-only 投影。
- 精准测试固定跨 scope、secret、全部漂移 / 归档、非法 transition、并发 CAS、partial-write rollback，以及 Gateway / run / retrieval / ranker 调用为 0。

批次 A 完成前不得创建 SQLite / PostgreSQL migration，不接应用配置草案或发布候选，不实现 Web。

## 批次 B：workflow durable store 与 migration

状态：等待批次 A。

- SQLite shared workflow database 追加 `0008_workflow_rag_knowledge_promotions`。
- PostgreSQL workflow run migration family 在 `0010` 后追加 `0011_workflow_rag_knowledge_promotions`，marker 推进为 `workflow_run_store_v11`。
- memory 共用 owner lock，SQLite / PostgreSQL 共用既有 database / pool 和单事务 CAS；禁止独立 DSN、pool、database file、selector 或 fallback。
- 覆盖 migration fresh / pending / rollback / reapply、checksum、运行角色 DDL 拒绝、append-only、并发单一成功、重启恢复、损坏记录、事务回滚与 no-fallback。

## 批次 C：应用配置 binding 与发布治理

状态：等待批次 B。

- 让 application draft v1 继续兼容读取，并以 `application_configuration_draft.v2` 保存 ref-only `workflow_rag_binding_ref`。
- 抽取服务端唯一 canonical draft digest；promotion source draft、草案保存与 publish candidate 使用同一规范化边界。
- 首次 attach / replace 只允许 binding ref 与 source draft 不同；服务端重读 binding 和全部来源后复用现有 draft CAS 保存下一版。
- `application_publish_candidate.v2` 只复制公开配置与 binding ref；create / approve / read-time eligibility 重新校验 binding，并把取消、漂移、归档和 store failure 映射为稳定 blocker。
- 证明现有 JSON payload 表无需 DDL；若测试证明需要新增列，必须先在既有 application draft / publish migration family 内评审并追加，不创建新 store。

## 批次 D：Web、连续链与收口

状态：等待批次 C。

- 新增默认 offline 的 strict promotion consumer 与 lazy panel，完成 list / detail / decision / CAS conflict。
- 接入应用配置草案 binding selector / attach，以及发布审查 exact binding / dynamic blocker；批准、附加、发布审查保持三个显式步骤。
- 覆盖 application switch、scope / schema drift、forbidden field、offline zero request、Web tests / build。
- 完成 SQLite / PostgreSQL 连续链、重启、no-fallback、真实浏览器复验和敏感材料扫描，再同步阶段文档并关闭专题。

## 验证矩阵

- Go domain / HTTP / repository / config / migration 精准测试、全包测试、race 与 vet 按风险执行。
- SQLite shared runtime 与 PostgreSQL integration 覆盖 migration、CAS、append-only、recovery、role、corruption 和 no-fallback。
- Web consumer / component tests、生产构建和实现完成后的真实浏览器路径。
- 每个子批至少执行 `git diff --check` 与 `./scripts/check-repo.sh --fast`；改变 schema、阶段边界或完成专题时补跑完整 `./scripts/check-repo.sh`。

## 停止线

- 不自动 baseline、promotion、release 或 publish；approve 不修改任何来源资源。
- 不调用 Gateway，不创建 run，不执行 retrieval / ranker，不启用 connector、在线搜索、embedding、vector database 或 reranker。
- 不实现后台任务、retry、fallback、replay、resume、writeback、生产认证、生产 secret、生产 API key、quota、billing 或生产能力。
- 不新增第二张任务卡、平行 store / selector / DSN / pool、同层 readiness / checker 文档链。
