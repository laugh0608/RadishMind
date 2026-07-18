# Workflow RAG 知识基线晋级与应用配置绑定审查 v1 实施任务卡

更新时间：2026-07-18

状态：`workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_batch_d_ready_for_implementation`

## 目标与准入结论

按 [功能设计](../features/workflow/workflow-rag-knowledge-baseline-promotion-application-binding-review-v1.md)交付“精确 dataset / candidate review / snapshot / profile / source draft → 人工 decision → 不可变 RAG binding → 应用配置草案引用 → 发布治理重校验”的完整开发测试态链路。

设计与边界评审已经通过，批次 A、批次 B 与批次 C 已完成并通过领域、HTTP、双数据库、并发与仓库门禁验证；批次 D 获得实现准入。本任务卡继续作为唯一实现入口，不派生第二张任务卡或同层 readiness 文档。

## 前置基线

- `dev` 分支保持可复验，既有 snapshot、durable evaluation dataset / candidate review、application configuration draft 与 publish governance 测试通过。
- promotion store 必须由现有 `workflowRunStore` backend selector 派生；应用草案与发布候选继续使用各自既有 repository。
- 独立 `RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV=1` gate 默认关闭；未完成所需 migration / marker 时服务启动失败，不回退 memory。
- 所有请求、响应、store payload、audit 和日志遵守 metadata-only / secret guard；不保存 query、fragment 正文、模型响应或 prompt。

## 批次 A：contract、领域与 memory 纵向链

状态：`completed`；完成锚点为 `workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_batch_a_completed`。

- 物化 `workflow_rag_knowledge_promotion_candidate.v1`、`workflow_rag_knowledge_promotion_decision.v1`、`workflow_rag_application_binding.v1` 与 audit contract。
- 实现服务端重读 dataset version / candidate review / baseline + candidate snapshots / lexical profile / source draft / application baseline。
- 实现 review eligibility、candidate digest、`pending / deferred / approved / rejected / canceled` 状态机、`expected_record_version` CAS、append-only decision / audit，以及 approve 原子创建 immutable binding。
- 注册四条 strict dev-only API、独立权限组合、稳定 failure envelope 和 list / detail metadata-only 投影。
- 精准测试固定跨 scope、secret、全部漂移 / 归档、非法 transition、并发 CAS、partial-write rollback，以及 Gateway / run / retrieval / ranker 调用为 0。

批次 A 完成前不得创建 SQLite / PostgreSQL migration，不接应用配置草案或发布候选，不实现 Web。

### 批次 A 完成证据

- 四份 JSON Schema、Go strict codec / validator 与 unknown-field 负例已经物化；candidate digest 覆盖精确 evidence、scope、创建者、创建时间和原始 request / audit ref，后续 decision 不改写这些不可变字段。
- 服务端从既有 application baseline、application draft、evaluation dataset / candidate review、snapshot repository 与 `workflowRAGLexicalProfile()` 重读全部权威来源；客户端不能提交 snapshot、profile、metrics 或 binding 推导。
- memory repository 与 workflow run / snapshot / evaluation 共用 owner lock；candidate current projection、decision、binding 和 audit 在单次 CAS append 中原子提交，注入 store failure 不留下部分状态。
- `approve / reject / defer / cancel`、终态、批准后取消、并发 CAS 单一成功、动态 eligibility、全部 dataset / review / snapshot / profile / draft / application 漂移或归档均有精准测试；非授权性决定可安全关闭漂移候选。
- 独立 `RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV=1` gate、四条 strict API、dedicated permission mapping、scope / secret guard、metadata-only list / conflict 和禁用门禁均已覆盖。
- 批次 A 未创建 SQLite / PostgreSQL migration，durable selector 明确失败且不回退 memory；未修改 application draft / publish candidate contract，未接 TypeScript / Web。
- Gateway 调用和 workflow run 创建均为 0；晋级批准只生成不可变配置 binding 资格，不修改 snapshot、dataset、review、应用配置草案、发布候选或应用状态。

## 批次 B：workflow durable store 与 migration

状态：`completed`；完成锚点为 `workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_batch_b_completed`。

- SQLite shared workflow database 追加 `0008_workflow_rag_knowledge_promotions`。
- PostgreSQL workflow run migration family 在 `0010` 后追加 `0011_workflow_rag_knowledge_promotions`，marker 推进为 `workflow_run_store_v11`。
- memory 共用 owner lock，SQLite / PostgreSQL 共用既有 database / pool 和单事务 CAS；禁止独立 DSN、pool、database file、selector 或 fallback。
- 覆盖 migration fresh / pending / rollback / reapply、checksum、运行角色 DDL 拒绝、append-only、并发单一成功、重启恢复、损坏记录、事务回滚与 no-fallback。

### 批次 B 完成证据

- SQLite shared workflow database 已追加 `0008_workflow_rag_knowledge_promotions`，总迁移序列推进到 `workflow_run_store_sqlite_v8`；PostgreSQL 已追加 `0011_workflow_rag_knowledge_promotions`，marker 推进到 `workflow_run_store_v11`。
- candidate current projection、append-only decision、immutable binding 与 append-only audit 共用既有 workflow database / pool；factory 对缺失 database / pool 和未知 backend fail closed，不存在独立 DSN、pool、selector 或 memory fallback。
- candidate create 与 decision / binding / audit append 均在单事务中提交；decision 使用 `expected_record_version` CAS，注入式 decision insert 失败会回滚 candidate projection，不留下部分 binding 或 audit。
- SQLite 与 PostgreSQL 均覆盖 exact scope、list、完整历史重建、append-only update / delete 拒绝、并发单一成功、进程重启恢复、损坏 JSON 拒绝、关闭 store no-fallback；PostgreSQL 运行角色仍无 DDL 权限。
- PostgreSQL 证据链读取固定为只读 `REPEATABLE READ`，避免 candidate projection 与 append-only history 在并发提交点形成撕裂视图；20 次专项 CAS 压力复验和完整 PostgreSQL integration 均通过。
- migration unit / configured product gate 已覆盖 fresh / pending / marker / checksum / rollback / reapply，既有聚合 SQLite runtime 迁移计数与 configured schema 断言已同步。

## 批次 C：应用配置 binding 与发布治理

状态：`completed`；完成锚点为 `workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_batch_c_completed`。

- 让 application draft v1 继续兼容读取，并以 `application_configuration_draft.v2` 保存 ref-only `workflow_rag_binding_ref`。
- 抽取服务端唯一 canonical draft digest；promotion source draft、草案保存与 publish candidate 使用同一规范化边界。
- 首次 attach / replace 只允许 binding ref 与 source draft 不同；服务端重读 binding 和全部来源后复用现有 draft CAS 保存下一版。
- `application_publish_candidate.v2` 只复制公开配置与 binding ref；create / approve / read-time eligibility 重新校验 binding，并把取消、漂移、归档和 store failure 映射为稳定 blocker。
- 证明现有 JSON payload 表无需 DDL；若测试证明需要新增列，必须先在既有 application draft / publish migration family 内评审并追加，不创建新 store。

完成证据：

- application draft v1 继续读取和保存；v2 只增加 `workflow_rag_binding_ref`，历史 v1 payload 缺少 `draft_digest` 时由服务端按同一 canonical 边界恢复。
- 首次 attach / replace 要求 `workflow_rag_promotions:bind`，服务端按 binding ref 精确读取 promotion resource，重校验当前 source draft 与全部权威知识来源，并证明除 binding ref 外没有夹带配置修改；保留同一 binding 的后续草案允许正常编辑。
- application publish candidate 仅在绑定草案上写 v2，未绑定路径继续写 v1；create 与 approve 在写入前重读草案、应用基线和 binding，read-time eligibility 将取消、dataset / snapshot / profile 漂移、应用归档与 store failure 映射为稳定 blocker。
- SQLite 与 PostgreSQL 继续复用 application draft / publish `0001` 表和 marker；JSON v2 保存、重建和重启测试通过，无 schema migration、回填、平行 selector 或 fallback。

## 批次 D：Web、连续链与收口

状态：`ready_for_implementation`。

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
