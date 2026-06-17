# Workflow / Agent Runtime 设计与开发文档

更新时间：2026-06-17

## 功能定位

`Workflow / Agent Runtime` 承载 AI 应用执行链路，包括 Prompt、LLM、HTTP tool、RAG retrieval、condition、output、后续受控 code / sandbox 和 agent loop。

它面向 `RadishMind` 用户工作区中的 workflow 设计、审查和后续受控执行：先让用户能保存、读取、列出和校验 workflow 草案，再进入 publish、run、confirmation、writeback 或 replay 等更高风险能力。

## 当前状态

- `apps/radishmind-web/` 已有 workflow application detail、definition detail、run detail、blocked action preview、confirmation placeholder、draft designer、validation inspector、execution plan preview、runtime readiness inspector、surface overview、context selection、scenario inspector、Review Workspace、Workspace Home 和 Review Handoff。
- 当前能力全部是 offline-only / read-only / advisory-only / blocked capability 组织层。
- `workflowWorkspaceContext` 已作为本地选择和派生关系的共享入口。
- `Saved Workflow Draft v1` 已完成 platform Go domain service 实现：`SavedWorkflowDraft` v1 类型、内存 dev store boundary、save / read / validate / list 契约、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests 已落地。
- 当前已实现 `Saved Workflow Draft v1` dev-only HTTP route + web consumer 状态区分，包括 save / read / validate / list 和 restore mapping；仍未实现 builder UI durable persistence、validation result persistence、execution plan persistence、runtime readiness persistence、workflow executor、confirmation decision store、writeback、replay 或 resume。
- 当前已实现 `User Workspace Saved Draft List v1`：dev-only list route 只返回 sanitized summary，Workspace Home 可显示已保存 dev draft、empty / failure state，并通过既有 read route 恢复到 Draft Designer。
- 当前已实现 `Workflow Draft Designer Editing Model v2`：Draft Designer 可本地新增节点、移动节点、删除非受保护节点、重建边，并让 validation inspector、execution plan preview 和 runtime readiness inspector 消费当前 active draft。
- 当前已实现 `Workflow Draft Node Attribute Editing Model v1`：Draft Designer 可编辑节点属性，并把节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs 保存 / 恢复到 dev-only saved draft。
- 当前已实现 `Workflow Review Handoff Active Draft v1`：Review Handoff 新增 active draft review record，汇总 active draft validation inspector、execution plan preview 和 runtime readiness inspector 的状态、blocker count、request / audit metadata 和 reviewer question。
- 当前已定义 `Saved Workflow Draft Durable Store Preconditions v1`，固定 draft scope、owner / workspace 归属、version conflict、no sample fallback、memory dev store 与未来 repository adapter 的切换停止线；`Saved Workflow Draft Repository Contract Preconditions v1` 已进一步固定 future repository contract 的 actor context、operation matrix、request / result、failure 和 projection 边界；`Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 `draft_schema_migration_preconditions_defined`，覆盖 future durable store logical schema、index strategy、migration gate、failure mapping 和 artifact guard；`Saved Workflow Draft Auth Context Preconditions v1` 已固定 `draft_auth_context_preconditions_defined`，覆盖 future repository actor context 的身份来源、workspace membership、owner policy、scope grants 和 audit / sanitization 边界；`Saved Workflow Draft Store Selector Enablement Preconditions v1` 已固定 `draft_store_selector_enablement_preconditions_defined`，覆盖 future store mode、selector gate、failure mapping、no fallback 和 artifact guard；`Saved Workflow Draft Schema Artifact Evidence v1` 已固定 `draft_schema_artifact_evidence_defined`，覆盖 future schema artifact manifest、DDL review、rollback evidence、migration smoke、failure mapping 和 artifact guard；`Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 `draft_store_selector_smoke_readiness_defined`，覆盖 future selector smoke mode matrix、operation matrix、schema artifact failure、no fallback 和 no side effects；`Saved Workflow Draft Repository Contract Smoke v1`、`Saved Workflow Draft Repository Contract Smoke Runner Readiness v1`、`Saved Workflow Draft Repository Contract Smoke Runner Implementation v1`、`Saved Workflow Draft Repository Adapter Implementation Plan v1`、`Saved Workflow Draft Schema Artifact Manifest v1`、`Saved Workflow Draft Adapter Smoke Readiness v1`、`Saved Workflow Draft Store Selector Implementation Entry Review v1` 和 `Saved Workflow Draft Schema Artifact Materialization Review v1` 已固定 future repository contract smoke、runner readiness、static runner implementation、adapter implementation plan、`draft_schema_artifact_manifest_defined` manifest contract、`draft_adapter_smoke_readiness_defined` adapter smoke readiness、`draft_store_selector_implementation_entry_review_defined` selector implementation entry review 和 `draft_schema_artifact_materialization_review_defined` schema artifact materialization review。这些证据不实现 durable repository adapter、数据库、OIDC、production API 或执行链路。
- `Workflow Saved Draft v1 Implementation` 任务卡已进入 `saved_workflow_draft_domain_service_implemented` 状态，consumer integration 由 `workflow-saved-draft-consumer-integration-v1` 承接；后续不应跳到 executor / confirmation / writeback / replay。
- 细专题已拆到 [Workflow 细专题入口](workflow/README.md)：`Saved Workflow Draft v1`、Draft Designer、User Workspace saved draft、Review Handoff 和 durable store 迁移前置专题分别承接功能、页面、用户流和后续实现准入。

## 设计边界

- draft、validation、plan、readiness、review 和 execution 是不同阶段，不能在 UI 中混成一个“已可执行”状态。
- 功能文档不再把旧停止线作为全局禁止清单复述；每个功能应说明本轮允许打开的能力、依赖的验证证据，以及需要留到后续独立目标的能力。
- 高风险 tool/action 默认要求 confirmation。
- executor 之前必须先有 run record、audit、failure taxonomy、materialized result 和 no side effects 策略。
- 上层 `RadishFlow` / `Radish` 未提供挂载点时，不阻塞本仓库设计离线草案、审查和 readiness，但不能声明真实接入 ready。

## Saved Workflow Draft v1

### 目标用户

- `Workspace Builder`：在用户工作区中编辑 workflow 草案，需要保存未完成设计、稍后继续修改，并看到结构和契约校验结果。
- `Workflow Reviewer`：审查草案是否满足发布或后续执行前置条件，需要读取同一份草案、确认风险、blocked capability 和缺失前置条件。
- `Platform Maintainer`：维护 workflow 草案 schema、校验规则、失败语义和停止线，确保后续 executor / confirmation / writeback 不会被隐式打开。

Saved Workflow Draft v1 不面向 public production API consumer，也不面向自动执行器。它的成功状态只能表示“草案已保存且可审查”，不能表示“workflow 已发布、可执行或已接入上层项目”。

### 可打开范围

Saved Workflow Draft v1 后续实现批次允许打开：

- draft save / read / validate / list 的受控入口。
- draft schema、类型定义或等价结构化格式。
- 草案版本、冲突检测、sanitized response 和 validation summary。
- 面向当前开发阶段的存储边界，例如 local file、dev store、fake store bridge 或后续经 task card 明确选择的正式存储实现。
- 与上述能力直接相关的 unit tests、negative tests、fixture、checker 和 web build 验证。

这些能力不再被旧的 `no persistence / no mutation` 口径笼统阻挡。是否进入真实数据库、OIDC、repository adapter 或 public production API，应由对应实现批次单独判断和记录；它们不是 Saved Workflow Draft v1 功能设计本身的默认前置条件。

### 当前实现

- Platform 内部已新增 Go domain service 和 memory dev store boundary，提供 `SaveDraft`、`ReadDraft`、`ValidateDraft` 和 `ListDrafts` 四个可测试入口。
- `SavedWorkflowDraft` v1 覆盖 identity、scope、version、schema、status、editable graph、contracts、provider / tool / RAG refs、requested blocked capabilities、validation summary、blocked capability summary 和 request / audit metadata。
- Go 单元测试已覆盖成功、invalid、blocked、scope denied、not found、schema unsupported、version conflict、payload too large、store unavailable、write disabled、no sample fallback、no partial write 和 no executor / confirmation / writeback / replay side effects。
- 当前实现已注册 dev-only HTTP route，并接入 web consumer 状态区分；仍不创建 public production API，不接真实数据库、OIDC、repository adapter、schema migration 或 store selector。

### 数据边界

Saved draft 是用户工作区内的可编辑设计记录，不是 workflow definition 的发布版本，也不是 run record。v1 数据边界如下：

- 身份与作用域：`draft_id`、`workspace_id`、`application_id`、可选 `source_definition_id`、可选 `base_definition_version`、`draft_version`、`schema_version`、`draft_status`、`created_at`、`updated_at`、`created_by_actor_ref`、`updated_by_actor_ref`。
- 用户可编辑内容：草案名称、说明、节点、边、输入契约、输出契约、provider / profile 需求、tool reference、RAG reference、condition 表达式摘要和 output mapping。
- 允许的节点类型：`prompt`、`llm`、`http_tool`、`rag_retrieval`、`condition`、`output`。`code`、`sandbox`、`agent_loop` 可作为 reserved / blocked capability 标记，不在 v1 中进入可执行能力。
- 派生审查信息：risk summary、requires confirmation summary、blocked capability summary、validation state 和 validation findings 可以随 save / read / validate 返回；list 只返回 sanitized summary / metadata。它们必须能从草案主体和当前校验策略重新计算，不能成为解锁执行的真相源。
- 明确排除：secret value、API key value、OAuth / OIDC token、完整用户 claim、真实工具执行结果、materialized result、confirmation decision、execution plan persistence、runtime readiness persistence、scenario / review / handoff persistence、run input / output、business writeback payload、replay / resume state。

实现批次若需要 committed schema 或持久化格式，必须先拆 task card，并按风险新增 fixture / checker；功能文档本身只定义设计边界。

### 保存流程

1. UI 从 draft designer 生成草案提交载荷，必须包含 workspace、application、schema version、draft version 和节点 / 边主体。
2. 服务端或本地实现入口先做 normalize：排序稳定化、默认值补齐、字段 allowlist、禁止字段剔除和长度 / 数量预算检查。
3. 校验器执行结构校验：节点 id 唯一、边端点存在、图结构可遍历、输入输出契约存在、节点类型在 allowlist 内、required provider/profile/tool/reference 只保存 ref 不保存 secret。
4. 校验器执行能力校验：发现 executor、confirmation decision、writeback、replay、public API、真实数据库、OIDC、repository adapter 或 production-only 字段时，必须标记 blocked 或拒绝保存。
5. 可恢复的业务不完整状态允许保存为 `invalid` 或 `blocked`，例如缺 provider/profile ref、缺 input contract 或存在 blocked capability；不可解析、越权、schema 不兼容或结构损坏的载荷必须拒绝保存。
6. 保存必须是原子写入：成功时创建或更新同一 `draft_id`，递增 `draft_version`，返回 sanitized draft、validation summary 和 audit/request metadata；失败时不得产生部分草案，也不得回退写入 sample / fixture。

### 读取流程

1. 读取必须按 workspace + application + draft scope 查询；未来接入真实身份前，只允许使用明确标记的 dev / fake auth context，不声明 production auth ready。
2. `draft_id` 不存在、scope 不匹配或 store 不可用时 fail closed；不能静默返回离线样例作为“已保存草案”。
3. 读取结果只返回 sanitized draft、版本信息、validation summary、blocked capability summary 和 request / audit metadata。
4. 读取时应重新运行或复核校验策略，确保旧草案在策略升级后能显示 `schema_version_unsupported`、`validation_policy_changed` 或新的 blocked finding。
5. UI 可以继续展示离线样例，但必须明确标记为 sample / unsaved，不能和 saved draft record 混同。

### 校验流程

Saved Workflow Draft v1 的校验目标是“可保存、可审查、可进入后续实现前置评估”，不是“可执行”。校验分层如下：

- Schema 校验：schema version、required fields、字段类型、长度预算、节点 / 边数组、引用字段和 forbidden fields。
- Graph 校验：节点 id 唯一、边端点存在、入口 / 出口可解释、循环策略明确、孤立节点和 unreachable output 给出 finding。
- Contract 校验：input contract、output contract、LLM prompt input、tool input/output、RAG retrieval input 和 condition output 的最小结构一致。
- Capability 校验：executor、confirmation decision、writeback、replay、真实数据库、OIDC、repository adapter、production API、secret resolver 和 public delivery 字段必须保持 blocked。
- Risk 校验：高风险 tool/action 必须产生 `requires_confirmation` finding；由于 v1 不实现 confirmation decision，它只能阻止发布 / 执行，不能解锁运行。

校验结果允许返回 `valid_for_review`、`invalid_draft`、`blocked_capability` 或 `schema_unsupported`。`valid_for_review` 只表示草案可以被审查或进入下一批实现前置评估，不表示 publish / run ready。

### 失败语义

Saved Workflow Draft v1 采用 fail-closed 语义。建议固定以下失败码，并在实现批次中保持 response、日志和测试一致：

| failure code | 语义 | 行为 |
| --- | --- | --- |
| `draft_scope_denied` | workspace / application / actor scope 不匹配 | 拒绝读写，不返回草案主体 |
| `draft_not_found` | `draft_id` 不存在或不属于当前 scope | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | 草案 schema 版本不被当前实现支持 | 拒绝保存；读取时只返回可审查 metadata 和升级提示 |
| `draft_payload_invalid` | JSON / 类型 / required fields 无法解析 | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 节点、边、入口、出口或循环策略不合法 | 可保存为 `invalid` 的情况需有明确 finding；结构损坏则拒绝 |
| `draft_contract_invalid` | input / output / tool / RAG / condition 契约不一致 | 可保存为 `invalid`，不得进入 publish / run |
| `draft_blocked_capability` | 包含 executor、confirmation decision、writeback、replay 等 v1 禁止能力 | 可保存为 `blocked` 或拒绝，取决于字段是否可安全降级为标记 |
| `draft_version_conflict` | 客户端基于旧 `draft_version` 写入 | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 节点数、边数或文本字段超出预算 | 拒绝保存 |
| `draft_store_unavailable` | 保存或读取存储不可用 | fail closed，不回退 fixture，不声明成功 |
| `draft_write_disabled` | 当前环境只允许只读或 sample 模式 | 拒绝保存，允许继续查看 sample |

所有失败都必须保留 no side effects：不执行节点、不调用 provider、不调用 tool、不写上层业务真相源、不创建 run record、不提交 confirmation decision。

## 下一批开发方向

1. Saved Workflow Draft v1 consumer integration、受控本地编辑、用户工作区创建入口、User Workspace saved draft list / restore、Draft Designer 本地结构编辑、节点属性编辑、Workflow Review Handoff Active Draft v1、durable store 迁移前置设计、repository contract preconditions、schema / migration preconditions、auth context preconditions、store selector enablement preconditions、schema artifact evidence、selector smoke readiness、repository contract smoke、runner readiness、static runner implementation、repository adapter implementation plan、schema artifact manifest contract、adapter smoke readiness、selector implementation entry review 和 schema artifact materialization review 已完成。下一步若转向 durable store，应在独立 selector implementation task card 或后续 schema artifact materialization implementation task card 中选择一个方向，不直接创建 repository adapter、数据库或 OIDC 接线；若继续用户审查体验，应选择新的明确审查增强点。
2. 如果继续新增 consumer contract、fixture 或 checker，应先更新任务卡或对应细专题，明确 route 不属于 public production API，也不接真实数据库、OIDC、repository adapter、schema migration 或 store selector。
3. 若只整理现有离线审查 UI、文案或阅读路径，不再新增逐项 task card / fixture / checker，复用现有 workflow 聚合 gate、web build 和 fast baseline。
4. 可执行 run、confirmation decision、writeback 和 replay 是 Saved Workflow Draft v1 之后的独立功能目标；进入前必须重新定义 run record、audit、failure taxonomy、materialized result、no side effects、人工确认和执行存储前置条件。

## 验收方式

- 本次功能设计文档更新：确认 Saved Workflow Draft v1 已明确目标用户、数据边界、保存 / 读取 / 校验流程、失败语义、验收方式和停止线；运行必要文档 / 仓库检查。
- Saved Workflow Draft v1 domain service 实现批次：schema 或等价类型定义、保存 / 读取 / 校验单元测试、negative tests、version conflict tests、storage boundary checks、no sample fallback tests、no side effects tests 和 `./scripts/check-repo.sh --fast`。
- Saved Workflow Draft v1 consumer integration 批次：按选择补 dev-only route / web consumer / consumer contract / fixture / checker，至少运行 Go 单元测试、web build、workflow 聚合 checker 和 `./scripts/check-repo.sh --fast`。
- 若实现批次改变阶段口径、治理入口、schema 真相源或高风险执行边界，补跑全量 `./scripts/check-repo.sh`。
- executor / confirmation / writeback / replay：必须作为后续独立目标，另开 task card，补 audit tests、人工确认路径、no side effects tests 和全量仓库验证。

## 停止线

- 不实现 workflow executor、node executor、tool executor、agent loop、publish、run、confirmation decision、decision store、execution unlock、business writeback、run replay、run resume 或 materialized result reader。
- 不把 Saved Workflow Draft v1 的草案存储边界直接扩张成 repository adapter、真实 schema migration、store selector、Radish OIDC、token validation、public production API consumer、API key lifecycle、quota enforcement、billing 或 cost ledger；若后续选择其中任一方向，应回到对应功能文档和实现批次单独推进。
- 不把 saved draft 的 `valid_for_review`、validation summary、risk summary 或 readiness summary 解释为 production ready、publish ready 或 executable ready。
- 不继续新增同层只读 evidence 面板来替代 Saved Workflow Draft v1；普通只读整理只复用现有 workflow gate、web build 和 fast baseline。
- 不把 `RadishFlow` / `Radish` 上层挂载点缺失写成本功能阻塞；同时也不声明真实接入 ready。
