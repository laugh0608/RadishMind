# Workflow RAG 知识基线晋级与应用配置绑定审查 v1

更新时间：2026-07-18

状态：`workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_ready_for_implementation`

## 功能目标

让内部 Workflow / RAG 维护者从一个当前有效的应用配置草案出发，选择精确的 durable evaluation dataset version 与已完成 candidate snapshot review，创建不可变知识基线晋级候选，经过人工 `approve / reject / defer / cancel` 决定后，获得一个可被应用配置草案引用的不可变 RAG binding 资格。发布治理随后重新读取并校验该 binding，确保发布候选只引用仍然有效的知识证据链。

本专题中的“晋级”是配置绑定资格晋级，不是 snapshot、dataset baseline、应用配置草案、发布候选或正式应用状态的自动变更。批准不会写入任何既有资源，也不会调用 Gateway、创建 workflow run 或执行 retrieval。

## 目标用户与完整用户路径

目标用户是具备应用配置草案、RAG 评测资源和人工审查权限的内部维护者。开发测试态允许同一参与者同时创建和审查候选，用于验证产品链；这不代表生产职责分离已经成立。

1. 维护者在当前活跃应用下恢复一个服务端已保存且校验为 `valid` 的应用配置草案版本。
2. 维护者从 metadata-only dataset 列表选择当前 active 且 latest 的精确 dataset version，再选择属于该 version 的 candidate review。
3. Web 只提交 dataset id / version / digest、candidate review id、draft id / expected version；不提交 metrics、snapshot 正文、query、fragment 或客户端推导的 binding。
4. 服务端重新读取应用目录、配置草案、dataset resource / version、candidate review、baseline / candidate snapshot resource / version 和当前 lexical profile，逐项校验 scope、lifecycle、latest version、digest 与 `rag_ref`。
5. 服务端确认 candidate review 的 candidate 质量状态为 `passed`，且 conclusion 为 `improved` 或 `unchanged`，再创建 `pending` 晋级候选；这些条件只产生人工审查资格，不自动批准。
6. 审查人打开候选 detail，查看精确 binding、review conclusion、漂移检查和 metadata-only evidence；dataset query 与 fragment 正文不进入该页面。
7. 审查人提交 `approve / reject / defer / cancel`、脱敏原因和 `expected_record_version`。repository 以 CAS 追加 decision 与 audit。
8. `approve` 在同一事务中追加决定、写当前状态投影，并创建一份不可变 `workflow_rag_application_binding.v1`；它只产生配置绑定资格，不更新任何来源资源。
9. `reject` 与 `cancel` 结束当前候选；`defer` 保留后续人工决定入口。非授权性决定不因来源已漂移而被阻塞，但永远不产生 binding 资格。
10. 维护者在应用配置草案中选择已批准 binding。首次附加或替换 binding 时，服务端重新读取 binding 及全部来源，并确认草案仍是候选绑定的精确 source draft；保存仍使用现有 draft CAS。
11. 服务端只把 `binding_id / binding_version / binding_digest` 写入配置草案；首次附加时除 binding 引用外的配置字段必须与 source draft 完全相同，禁止借 binding 操作夹带其它配置修改。
12. 后续草案版本可以保留同一 binding 并正常编辑其它公开配置字段；切换 binding 必须重新走精确 source draft 与已批准资格校验。
13. 创建或批准应用发布候选时，发布治理重新读取发布候选精确绑定的当前配置草案版本 / digest，以及草案中的 binding、晋级候选、decision、dataset、review、snapshots 与 profile。任一漂移、归档、取消或存储故障都形成稳定 blocker 或直接拒绝创建 / 批准。
14. 已存在的草案和发布候选始终可按原作用域读取；binding 失效只改变动态资格与 blocker，不回写历史、不自动创建新候选，也不自动发布。

## 真相源与所有权

| 数据 | 权威所有者 | 本专题读取 / 写入 | 禁止行为 |
| --- | --- | --- | --- |
| 应用 lifecycle / baseline revision | `ApplicationCatalogRepository` 或既有兼容只读基线 | 创建、批准、草案绑定、发布复核时服务端重读 | 不从 Web、URL 或候选载荷伪造 |
| 应用配置草案 | `applicationConfigurationDraftRepository` | 精确读取当前 owner 下 draft；binding 通过既有 save CAS 写下一版 | 不建立第二套草案 store，不复制完整草案到 promotion store |
| dataset resource / immutable version / candidate review | `workflowRAGEvaluationDatasetRepository` | 按 id / version / review id 精确重读 | 不接受客户端 metrics、finding、snapshot binding 推导或正文 |
| snapshot resource / immutable version | `workflowRAGSnapshotRepository` | baseline 与 candidate 均按 id / version 重读并校验 latest / active | 不复制 fragment，不创建 promotion snapshot |
| lexical profile | 既有 `workflowRAGLexicalProfile()` 服务端定义 | 每个资格检查点重新取得 version / digest | 不由客户端或 promotion store定义 profile |
| promotion candidate / decision / binding / audit | 新增 workflow-owned promotion repository | 保存引用、digest、状态与 append-only 责任链 | 不保存 dataset / snapshot / draft 正文，不建立独立 backend selector |
| application publish candidate | 既有 publish governance repository / service | 不可变候选只保存 RAG binding ref，资格动态重算 | 不复制 promotion evidence、query、fragment 或模型内容 |

## 资源模型与不可变边界

### `workflow_rag_knowledge_promotion_candidate.v1`

候选创建后不可修改的字段至少包括：

- `candidate_id`、tenant / workspace / application scope、`candidate_digest`；
- `dataset_id / dataset_version / dataset_digest`；
- `candidate_review_id`；
- baseline 与 candidate 的 `snapshot_id / snapshot_version / snapshot_digest / rag_ref`；
- `profile_id / profile_version / profile_digest`；
- source application draft 的 `draft_id / draft_version / draft_digest / base_application_updated_at`；
- 创建者、创建时间、原始 request / audit ref。

`candidate_digest` 覆盖以上不可变业务字段，不覆盖 current state、`record_version`、decision、动态 blocker 或读取请求 metadata。候选不复制 review metrics、findings、query、samples、fragment、配置正文或模型材料；detail 需要展示 review evidence 时由服务端按 `candidate_review_id` 重读权威 review。

### `workflow_rag_knowledge_promotion_decision.v1`

每条 decision 至少包含 `decision_id`、candidate id / digest、`decision`、脱敏 `reason`、from / to state、before / after `record_version`、actor、occurred_at、request / audit ref。decision append-only，禁止 update / delete。

### `workflow_rag_application_binding.v1`

只有首次有效 `approve` 才创建 version `1` 的不可变 binding。binding 包含：

- `binding_id / binding_version / binding_digest`；
- promotion candidate id / digest 与批准 decision id / record version；
- 与 candidate 完全相同的 dataset、review、baseline / candidate snapshot、profile 和 scope 引用；
- source draft id / version / digest；
- issued_at / issued_by / request / audit ref。

binding 不保存 metrics、review reason、草案配置、dataset samples 或知识正文。批准后的 `cancel` 只追加 decision 并使动态资格失效，不删除或改写 immutable binding。

### 应用配置草案与发布候选

- `application_configuration_draft.v2` 在现有公开配置上新增可选 `workflow_rag_binding_ref`，只允许 `binding_id / binding_version / binding_digest`。v1 读取兼容为无 binding。
- 草案 canonical digest 统一由服务端对规范化公开配置和 binding ref 计算，排除 validation、actor、request、audit、timestamp 与动态资格。
- `application_publish_candidate.v2` 的不可变配置快照只复制现有公开配置和同一 binding ref；不复制 promotion candidate、decision、dataset、review 或 snapshot 内容。
- 既有 PostgreSQL / SQLite 表已经以 JSON payload 保存草案和发布候选，v2 codec 不要求为了新增字段创建平行表；若实现审计证明必须新增可查询列，才在原 migration family 内追加迁移。

## 状态机、CAS 与原子性

| 当前状态 | 允许决定 / 下一状态 | 语义 |
| --- | --- | --- |
| `pending` | `approve -> approved`、`reject -> rejected`、`defer -> deferred`、`cancel -> canceled` | 等待人工审查 |
| `deferred` | `approve -> approved`、`reject -> rejected`、`cancel -> canceled` | 可继续审查；重复 defer 拒绝 |
| `approved` | `cancel -> canceled` | 已生成 binding；批准本身不改其它资源 |
| `rejected` | 无 | 终态，不产生 binding |
| `canceled` | 无 | 终态；已生成 binding 时动态资格失效 |

所有决定要求 `expected_record_version`。candidate current projection CAS、decision append、audit append，以及 approve 时 binding create 必须在同一 workflow backend 原子边界内完成。任一步失败时不得留下部分状态。冲突只返回权威 current version / state 和 metadata-only blocker，不自动 retry、merge 或 fallback。

`approve` 必须重新读取全部权威来源并满足资格；`reject / defer / cancel` 不产生资格，只需重读候选与当前 scope，因此即使知识来源已漂移也允许安全关闭或延期。来源漂移绝不能被非授权性决定转换为 binding 资格。

## 漂移、归档与失败关闭

以下检查在 candidate create、approve、首次 attach / replace binding、publish candidate create、publish approve 与动态 eligibility read 中按适用阶段执行：

- dataset resource 必须 active，且 current latest version / digest 等于候选绑定；新 dataset version 或 archive 立即使资格失效；
- candidate review 必须存在、通过 stored contract 校验，并精确绑定同一 dataset、baseline / candidate snapshots 与 profile；
- baseline / candidate snapshot resource 必须 active，且 current latest version / digest / `rag_ref` 等于候选绑定；新 version、archive、缺失或损坏都阻断；
- 当前 lexical profile version / digest 必须等于候选绑定；profile 漂移不做兼容猜测；
- promotion create / approve 前，source draft 必须仍为绑定的 current version / digest、校验为 valid，且应用 active、base revision 未漂移；
- 首次 attach / replace 前，当前草案必须仍是 source draft，且除 binding ref 外的公开配置不得变化；attach 成功后该 source draft 只作为不可变 provenance，不再要求它继续是 latest；
- attach 后若 promotion candidate 被取消、任一知识资源漂移或存储不可用，草案仍可读取，但 binding 进入 invalid，不能创建或批准发布候选；
- publish candidate create / approve 必须重读其精确绑定的配置草案 id / version / digest；草案已产生更新版本、被替换、损坏或不再 valid 时失败关闭，不能借旧发布候选绕过当前草案审查；
- application archive、跨 owner / scope、repository contract mismatch 或任何权威 read failure 都在查询其它不必要资源或写入前 fail closed。

系统不自动修改 candidate state 为 invalidated，也不改写草案或发布候选。read / eligibility 响应返回有序 blocker，维护者必须从当前资源重新创建 candidate、重新人工审查并显式替换 binding。

## API、strict JSON 与权限

新增独立开发测试态门禁：

```text
RADISHMIND_WORKFLOW_RAG_PROMOTION_DEV=1
```

该门禁不替代既有 evaluation、application draft 或 publish gate。完整 Web 路径需要相关既有门禁同时开启；任一 store / marker / repository 初始化失败时服务启动失败，不回退 memory。

新增四条 scoped API：

```text
POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates
GET  /v1/user-workspace/workflow-rag-knowledge-promotion-candidates
GET  /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}
POST /v1/user-workspace/workflow-rag-knowledge-promotion-candidates/{candidate_id}/decisions
```

create body 只允许 `workspace_id`、`application_id`、dataset id / version / digest、`candidate_review_id`、`draft_id`、`expected_draft_version`。服务端生成 candidate id / digest，并从权威资源推导 snapshots、profile 与 draft digest。decision body 只允许 scope、`expected_record_version`、`decision` 与 4–500 字符脱敏原因。未知字段、敏感材料、非法 scope 或超限载荷一律拒绝。

草案 attach / replace 不新增平行 binding API，复用既有 application configuration draft save API、scope、validation 与 CAS，只把 request / response contract 升级为可选 v2 binding ref。publish candidate create / decision 也复用既有 publish governance API，并在服务层增加 binding 重校验。

| 操作 | 权限 |
| --- | --- |
| list / detail / binding metadata read | `workflow_rag_promotions:read` |
| create candidate | `workflow_rag_promotions:write`、`workflow_rag_evaluation_datasets:read`、`workflow_rag_snapshots:read`、`application_drafts:read` |
| approve / reject / defer / cancel | `workflow_rag_promotions:review` |
| 草案 attach / replace binding | `application_drafts:write`、`workflow_rag_promotions:bind` |
| 含 RAG binding 的发布候选 create / approve | 既有 publish 权限加 `workflow_rag_promotions:read` |

开发测试态可由同一 actor 持有全部权限；生产职责分离、真实 membership 与授权继续关闭。

## 稳定失败语义

| failure code | 触发条件 |
| --- | --- |
| `workflow_rag_promotion_scope_denied` | 身份、权限、tenant / workspace / application / owner 不一致 |
| `workflow_rag_promotion_payload_invalid` | strict JSON、标识、版本、reason 或预算非法 |
| `workflow_rag_promotion_secret_material_forbidden` | 请求、reason 或存储 payload 含凭据 / secret |
| `workflow_rag_promotion_not_found` | 当前 scope 下 candidate 不存在 |
| `workflow_rag_promotion_dataset_changed` / `_dataset_archived` | dataset 非 latest 精确 binding 或已归档 |
| `workflow_rag_promotion_review_invalid` / `_review_not_eligible` | review binding 损坏，或结果未达到人工晋级前置条件 |
| `workflow_rag_promotion_snapshot_changed` / `_snapshot_archived` | 任一 snapshot 非 current exact binding 或已归档 |
| `workflow_rag_promotion_profile_changed` | lexical profile version / digest 漂移 |
| `workflow_rag_promotion_draft_changed` / `_draft_invalid` | source draft version / digest / validation / base revision 漂移 |
| `workflow_rag_promotion_application_archived` | 当前应用已归档 |
| `workflow_rag_promotion_record_version_conflict` | decision CAS 失败；返回 current version / state |
| `workflow_rag_promotion_transition_invalid` | 非法或重复状态迁移 |
| `workflow_rag_binding_not_eligible` | binding 未批准、已取消或权威来源失效 |
| `workflow_rag_promotion_store_unavailable` / `_store_contract_mismatch` | store、事务、marker、decode 或权威依赖失败 |
| `workflow_rag_promotion_write_disabled` | 独立开发写入未开启 |

所有失败都不创建候选、decision、binding、草案版本或发布候选。公开错误、日志和冲突 envelope 只返回稳定 code 与 metadata，不返回 query、fragment、草案正文、decision 详情或底层数据库错误。

## 审计、隐私与 metadata-only 边界

- candidate create、每条 decision、binding issued、draft binding attach / replace，以及 publish candidate create / decision 都记录 append-only metadata audit；纯 list / detail / eligibility read 不修改 promotion 业务资源，如既有平台记录请求审计，也只能写 metadata-only request / audit ref。当前只使用既有开发测试态 repository，不声明生产 audit store。
- list、audit、日志、普通 review record、冲突响应与 publish blocker 必须 metadata-only。
- promotion candidate / decision / binding / audit 禁止保存 query、dataset samples、review note、fragment content、selected excerpt、prompt、messages、模型响应、Gateway request / response、credential、secret、endpoint、header、cookie 或 DSN。
- Web 状态不进入 URL payload、`localStorage` 或 `sessionStorage`；应用切换清除 candidate、reason、binding selection 和 conflict state。
- candidate review 仍由既有 repository 持有 metadata-only metrics / refs；promotion store 只保存 review id 与强制 exact bindings，不复制完整 review。

## Store ownership 与 migration 顺序

- promotion repository 从现有 `workflowRunStore` backend selector 派生，不新增 mode、DSN、database file、pool 或 fallback。
- `memory_dev` 与 workflow run / snapshot / evaluation store 共用 owner lock；保存 current projection、immutable candidate、append-only decision / binding / audit。
- `sqlite_dev` 复用 shared SQLite database，在 workflow-runs migration family 追加 `0008_workflow_rag_knowledge_promotions`，所有 CAS 与 append 在同一事务完成。
- `postgres_dev_test` 复用 workflow run pool，在现有 `0010` 后追加 `0011_workflow_rag_knowledge_promotions`，schema marker 推进为 `workflow_run_store_v11`；手动 migration、checksum、rollback / reapply 和运行角色边界保持现有口径。
- 先部署兼容 v1 / v2 的 application draft 与 publish candidate decoder / validator，再允许写 v2 payload；现有 JSON payload 表首版无需 DDL 或回填。任何必须新增列的结论必须先补原 migration family 的迁移评审，不得绕过为新表或新 store。
- 启用顺序固定为：workflow migration → repository / service → v1 / v2 consumer compatibility → promotion gate → Web source。旧 schema、marker 或任一组件不兼容时启动或请求 fail closed。

## Web 使用路径

Web 在现有 RAG evaluation dataset / candidate review 面板之后增加 lazy-loaded promotion / binding review 区域，并在应用配置草案与发布审查面板显示同一 binding 的动态状态：

- 默认 offline 零请求；只有 `VITE_RADISHMIND_WORKFLOW_RAG_PROMOTION_SOURCE=dev-workflow-rag-promotion-http` 才访问 promotion API；
- candidate list 只显示 id、dataset / review refs、source draft ref、state、record version、binding ref、时间与 blocker count；
- detail 展示精确 snapshot / profile digests、权威 review metadata 与 drift blocker，不显示 query / fragment / prompt / answer；
- decision UI 明确区分 approve 与 attach，批准后不自动保存草案；CAS conflict 保留本地 reason 并要求刷新；
- 配置草案只选择 approved binding ref，首次 attach 单独保存下一 draft version；
- publish panel 展示 exact binding ref 与动态复核 blocker，不能把 `approved` 显示为 published；
- 应用切换、scope drift、非 2xx、strict schema drift 或 forbidden field 都清除旧交接并失败关闭。

## 测试矩阵

| 层级 | 必须覆盖 |
| --- | --- |
| domain / memory | server-side reload、exact bindings / digest、review eligibility、状态机、CAS、atomic approve + binding、secret guard、零自动 mutation |
| drift / scope | dataset version / archive、两侧 snapshot version / archive、review mismatch、profile drift、draft version / digest / validation、application archive、owner / scope denial |
| SQLite / PostgreSQL | migration / rollback / reapply、marker / checksum、共享 backend、并发 CAS 单一成功、append-only、重启恢复、损坏记录拒绝、事务回滚、no-fallback |
| application draft | v1 兼容、v2 ref-only binding、首次 attach 只允许 binding diff、保留 / 替换 binding、CAS、invalid binding 拒绝、list metadata-only |
| publish governance | server-side draft / binding reload、candidate digest 含 binding ref、create / approve fail-closed、read-time blockers、取消 / 漂移后旧候选可读但不可晋级 |
| HTTP / auth | 独立 gate、strict request、权限组合、三层 scope、稳定 failures、冲突 metadata-only |
| Web | offline 零请求、strict consumers、list / detail / decision / attach / publish review、application switch、scope / schema / secret rejection、tests / build |
| 不回归 | existing snapshot、evaluation dataset / review、application draft v1、publish candidate v1、RAG execution、Run History、Comparison / Evaluation / Suite、HTTP Tool |

专项测试必须证明 promotion create / decision / binding attach / publish revalidation 的 Gateway、workflow run、retrieval ranker、network、retry、replay 与 writeback 调用数均为 `0`。

## 实施批次与实现准入

唯一实施入口为[实施任务卡](../../task-cards/workflow-rag-knowledge-baseline-promotion-application-binding-review-v1-plan.md)，按以下依赖顺序推进：

1. 批次 A：版本化 contract、领域服务、memory repository、独立 gate、四条 strict API、权限、全部 server-side reload、状态机 / CAS / append-only decision / binding / audit 与精准测试。
2. 批次 B：SQLite `0008`、PostgreSQL `0011`、workflow backend 派生 repository、事务原子性、重启、损坏记录、migration 与 no-fallback。
3. 批次 C：应用配置草案 v2 ref-only binding、共享 canonical draft digest、首次 attach 规则，以及发布候选 v2 / publish governance 的 binding 重校验与兼容测试。
4. 批次 D：Web promotion panel、配置草案 / 发布审查接线、应用切换隔离、完整测试 / build、SQLite / PostgreSQL 连续链、真实浏览器和阶段收口。

边界评审结论为通过：既有 repository 和 backend ownership 足以承载本功能；资源、状态机、API、migration、Web 与测试停止线已经明确，不需要新增同层 readiness / checker 文档。状态推进到 `workflow_rag_knowledge_baseline_promotion_application_binding_review_v1_ready_for_implementation`。本次只获得批次 A 的实现准入，不代表任何 runtime、schema migration、API 或 Web 已实现。

## 停止线

- 不自动 baseline、promotion、release 或 publish；approve 只产生可复核的配置 binding 资格。
- 不修改 snapshot、dataset baseline / version、candidate review、应用配置草案、发布候选、应用目录或发布状态。
- 不调用 Gateway，不创建 workflow run，不执行 retrieval 或 lexical ranker。
- 不启用 connector、crawler、文件扫描、在线搜索、embedding、vector database、reranker 或外部 provider。
- 不实现后台任务、schedule、retry、fallback、replay、resume、writeback、agent loop 或自动修复漂移。
- 不接生产 OIDC / membership、生产 repository、production audit、生产 secret、生产 API key、quota、billing 或生产能力声明。
- 不创建第二张任务卡、平行 store / selector / DSN / pool、同层 readiness 文档或专项 checker 链；实现验证优先进入现有 Go / Web / migration / 聚合门禁。
