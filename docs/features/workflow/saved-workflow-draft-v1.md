# Saved Workflow Draft v1 功能专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft v1` 是 `Workflow / Agent Runtime` 中第一个从只读审查面走向可保存用户草案的功能专题。它让用户工作区中的 workflow 草案可以被保存、读取、列出和校验，并让 reviewer 看到同一份可审查记录。

本专题不是 workflow publish / run / executor 专题。成功保存只表示“草案已保存且可审查”，不表示 workflow 已发布、可执行或已接入上层项目。

## 当前状态

- Platform Go domain service 已实现，文件为 `services/platform/internal/httpapi/workflow_saved_draft.go`。
- 已覆盖 `SavedWorkflowDraft` v1 类型、memory dev store、`SaveDraft` / `ReadDraft` / `ValidateDraft` / `ListDrafts`、版本冲突、失败语义、sanitized response、no sample fallback 和 no side effects tests。
- 当前已新增 dev-only HTTP route 和 web consumer 状态区分：`POST /v1/user-workspace/workflow-drafts`、`GET /v1/user-workspace/workflow-drafts/{draft_id}`、`GET /v1/user-workspace/workflow-drafts` 和 `POST /v1/user-workspace/workflow-drafts/validate` 默认关闭，只在显式 dev 配置下工作。
- 当前已补 route contract 和 consumer smoke：Go route test 固定 envelope、header、CORS、not found / store unavailable no sample fallback；前端 consumer 固定 `version_conflict` 状态，version conflict 时保留本地草案并展示当前 saved draft version metadata。
- 当前已接入 [Workflow Draft Editing Entry v1](draft-editing-entry-v1.md)：Draft Designer 可编辑草案名称、说明、节点名称和边条件摘要，validate / save / read 使用当前本地草案。
- 当前已接入 [User Workspace Draft Creation v1](user-workspace-draft-creation-v1.md)：用户可从 Workspace Home 或 workflow definitions 创建本地草案，再复用 dev-only saved draft consumer 保存。
- 当前已接入 [User Workspace Saved Draft List v1](user-workspace-saved-draft-list-v1.md)：Workspace Home 可读取当前 application 的 saved dev draft sanitized summary，并通过 read route 恢复到 Draft Designer。
- 当前已接入 [Workflow Draft Designer Editing Model v2](draft-designer-editing-model-v2.md)：Draft Designer 可本地新增节点、移动节点、删除非受保护节点并重建边，validate / save / read 继续消费当前 active draft。
- 当前已接入 [Workflow Draft Node Attribute Editing Model v1](draft-node-attribute-editing-model-v1.md)：Draft Designer 可编辑节点级 provider / profile、tool ref、RAG ref、summary、contract fields 和 output mapping，dev-only saved draft payload / response 会保存并恢复节点级 summary、contract fields、output mapping 和 refs。
- 当前已接入 [Workflow Review Handoff Active Draft v1](review-handoff-active-draft-v1.md)：Review Handoff 会把恢复后的 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总为 advisory-only 审查交接记录。
- 当前已新增 [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md)：固定 durable store 迁移前的 draft scope、`owner_subject_ref` / workspace 归属、version conflict、no sample fallback、dev store 与未来 repository adapter 的切换停止线；它只定义前置设计，不实现 durable persistence。
- 当前已新增 [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md)：固定 future `SavedWorkflowDraftRepository` 的 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords` contract preconditions；它只定义 actor context、request / result、failure 和 projection 边界，不创建 repository interface。
- 当前已新增 [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md)：固定 future durable store 的 logical schema、index strategy、migration gate、failure mapping、no sample fallback 和 artifact guard；状态为 `draft_schema_migration_preconditions_defined`，不创建真实数据库 schema 或 SQL migration。
- 当前已新增 [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md)：固定 future repository actor context 的身份来源、workspace membership、owner policy、scope grants、failure policy 和 audit / sanitization 边界；状态为 `draft_auth_context_preconditions_defined`，不创建 Radish OIDC middleware 或 token validation。
- 当前已新增 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md)：固定 future `memory_dev` / `repository_disabled` / `repository` store mode、selector gate、failure mapping、no fallback、no side effects 和 artifact guard；状态为 `draft_store_selector_enablement_preconditions_defined`，不创建正式 config entry 或 selector 代码。
- 当前已新增 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md)：固定 future schema artifact manifest、DDL review、rollback evidence、migration smoke、failure mapping 和 artifact guard；状态为 `draft_schema_artifact_evidence_defined`，不创建 migration root、manifest、DDL review artifact 或 SQL migration。
- 当前已新增 [Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)：固定 future selector smoke 的 `memory_dev` / `repository_disabled` / `repository` / unknown mode、save / read / list operation matrix、schema artifact failure、no fallback 和 no side effects；状态为 `draft_store_selector_smoke_readiness_defined`，不创建正式 config entry、selector 代码或 selector smoke fixture。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke v1](saved-workflow-draft-repository-contract-smoke-v1.md)：固定 future repository contract smoke 的 `SavedWorkflowDraftRepositoryContractSmoke` harness、save / read / list I/O、failure mapping、no sample / memory dev fallback 和 no side effects；状态为 `draft_repository_contract_smoke_defined`，不创建 smoke runner、repository interface、repository adapter、store selector 或数据库 artifact。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke Runner Readiness v1](saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md)：固定 future `SavedWorkflowDraftRepositoryContractSmokeRunner` 的 runner I/O、save / read / list operation runner matrix、failure mapping、no fallback 和 no side effects；状态为 `draft_repository_contract_smoke_runner_readiness_defined`，不创建 runner Go 文件、runner test、repository interface、repository adapter、store selector 或数据库 artifact。
- 当前已新增 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)：实现 static `SavedWorkflowDraftRepositoryContractSmokeRunner` 和 Go tests；状态为 `draft_repository_contract_smoke_runner_implemented`，仍不创建 repository interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)：固定 future repository adapter implementation plan、operation adapter matrix、failure mapping、no memory dev / sample fallback 和 no side effects；状态为 `draft_repository_adapter_implementation_plan_defined`，仍不创建 repository interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md)：固定 future schema artifact manifest shape、section matrix、operation predicate coverage、failure mapping、no memory dev / sample fallback 和 no side effects；状态为 `draft_schema_artifact_manifest_defined`，仍不创建 migration root、manifest 文件、DDL review artifact、SQL migration、真实数据库、repository adapter、store selector、Radish OIDC middleware、token validation 或 production API artifact。
- 当前已新增 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)：固定 future adapter smoke readiness、依赖 gate、operation adapter smoke matrix、failure mapping、no memory dev / sample / fixture / dev route / test auth fallback 和 no side effects；状态为 `draft_adapter_smoke_readiness_defined`，仍不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、selector、schema artifact 文件、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API artifact。
- 当前仍没有 durable persistence、repository adapter、schema artifact manifest 文件、DDL review artifact、真实 schema migration、store selector implementation、selector smoke fixture、adapter smoke fixture、adapter smoke checker、Radish OIDC middleware、token validation 或 production API。
- 当前任务卡为 [Workflow Saved Draft v1 Implementation](../../task-cards/workflow-saved-draft-v1-implementation-plan.md)，状态是 `saved_workflow_draft_domain_service_implemented`。

## 目标用户

- `Workspace Builder`：保存未完成的 workflow 设计，后续继续编辑。
- `Workflow Reviewer`：读取同一份草案，审查结构、风险和 blocked capability。
- `Platform Maintainer`：维护 schema、校验规则、失败语义和停止线，避免后续 executor 或 confirmation 被隐式打开。

## 数据边界

Saved draft 是用户工作区中的可编辑设计记录，不是 published workflow definition，也不是 run record。

允许保存：

- `draft_id`、`workspace_id`、`application_id`、`source_definition_id`、`base_definition_version`、`draft_version`、`schema_version`、`draft_status`。
- 未来 durable store 前必须补齐 `tenant_ref`、`owner_subject_ref`、`created_by_actor_ref`、`updated_by_actor_ref`、workspace membership 和 tenant predicate；当前 dev-only memory store 只作为开发态 dev store，不代表 owner / workspace production auth 已就绪。
- 草案名称、说明、节点、边、输入契约、输出契约、provider / profile ref、tool ref、RAG ref、condition 摘要和 output mapping。
- 节点级 summary、contract fields、output mapping 和 provider / tool / RAG refs；这些字段只描述草案设计，不解锁 runtime binding。
- validation summary、blocked capability summary、request / audit metadata。

明确排除：

- secret value、API key value、OAuth / OIDC token、完整用户 claim。
- 真实工具执行结果、materialized result、confirmation decision。
- execution plan persistence、runtime readiness persistence、scenario / review / handoff persistence。
- run input / output、business writeback payload、replay / resume state。

## 功能流程

保存流程：

1. UI 或 consumer 提交 sanitized draft payload，必须包含 workspace、application、schema version、draft version 和 graph 主体。
2. 服务端执行 normalize、field allowlist、forbidden field reject 和大小预算检查。
3. 校验 schema、graph、contract、capability 和 risk。
4. 成功时原子保存并递增 `draft_version`，返回 sanitized draft、validation summary、blocked capability summary 和 request / audit metadata。
5. 失败时不得产生部分写入，不得回退 sample / fixture，不得创建 run record 或 confirmation decision。

读取流程：

1. 按 workspace + application + draft scope 查询。
2. scope mismatch、not found 或 store unavailable 必须 fail closed。
3. 读取结果只返回 sanitized draft、版本信息、validation summary、blocked capability summary 和 request / audit metadata。
4. UI 可以继续展示 sample，但必须明确标记 `sample / unsaved`，不能混成 saved draft record。

校验流程：

- `valid_for_review` 只表示草案可审查，不表示 publish ready 或 run ready。
- `invalid_draft` 可用于可恢复的不完整业务草案。
- `blocked_capability` 可用于保留 code、sandbox、agent_loop、executor、confirmation decision、writeback、replay 等禁止能力的审查 finding。
- `schema_unsupported` 用于旧 schema 或不支持 schema 的可审查失败状态。

## 失败语义

必须保持以下 failure code 和行为一致：

| failure code | 行为 |
| --- | --- |
| `draft_scope_denied` | 拒绝读写，不返回草案主体 |
| `draft_not_found` | 返回 not found，不回退 sample |
| `draft_schema_version_unsupported` | 拒绝保存；读取时只返回可审查 metadata |
| `draft_payload_invalid` | 拒绝保存，无部分写入 |
| `draft_graph_invalid` | 结构损坏则拒绝；可恢复问题可保存为 invalid finding |
| `draft_contract_invalid` | 可保存为 invalid，不允许 publish / run |
| `draft_blocked_capability` | 可保存为 blocked finding 或拒绝危险字段 |
| `draft_version_conflict` | 拒绝覆盖，返回当前版本 metadata |
| `draft_payload_too_large` | 拒绝保存 |
| `draft_store_unavailable` | fail closed，不回退 fixture |
| `draft_write_disabled` | 拒绝保存，允许继续查看 sample |

## 下一批开发

dev-only consumer integration 已按 [Dev-only Saved Draft Consumer](dev-only-saved-draft-consumer.md) 落地，并已补 route contract、consumer smoke 和 version conflict UI 状态；正式草案编辑入口已按 [Workflow Draft Editing Entry v1](draft-editing-entry-v1.md) 落地；User Workspace 创建入口已按 [User Workspace Draft Creation v1](user-workspace-draft-creation-v1.md) 落地；saved dev draft list / restore 已按 [User Workspace Saved Draft List v1](user-workspace-saved-draft-list-v1.md) 落地；本地图结构编辑已按 [Workflow Draft Designer Editing Model v2](draft-designer-editing-model-v2.md) 落地；节点属性编辑已按 [Workflow Draft Node Attribute Editing Model v1](draft-node-attribute-editing-model-v1.md) 落地；恢复后的 Review Handoff 已按 [Workflow Review Handoff Active Draft v1](review-handoff-active-draft-v1.md) 汇总 active draft review record；durable store 迁移前置设计已按 [Saved Workflow Draft Durable Store Preconditions v1](saved-workflow-draft-durable-store-preconditions-v1.md) 固定；future repository contract preconditions 已按 [Saved Workflow Draft Repository Contract Preconditions v1](saved-workflow-draft-repository-contract-preconditions-v1.md) 固定；future schema / migration preconditions 已按 [Saved Workflow Draft Schema / Migration Preconditions v1](saved-workflow-draft-schema-migration-preconditions-v1.md) 固定；auth context preconditions 已按 [Saved Workflow Draft Auth Context Preconditions v1](saved-workflow-draft-auth-context-preconditions-v1.md) 固定；store selector enablement preconditions 已按 [Saved Workflow Draft Store Selector Enablement Preconditions v1](saved-workflow-draft-store-selector-enablement-preconditions-v1.md) 固定；schema artifact evidence 已按 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md) 固定；selector smoke readiness 已按 [Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md) 固定；repository contract smoke 已按 [Saved Workflow Draft Repository Contract Smoke v1](saved-workflow-draft-repository-contract-smoke-v1.md) 固定；repository contract smoke runner readiness 已按 [Saved Workflow Draft Repository Contract Smoke Runner Readiness v1](saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md) 固定；static runner implementation 已按 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md) 落地；future adapter implementation plan 已按 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md) 固定；schema artifact manifest shape 已按 [Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md) 固定；adapter smoke readiness 已按 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md) 固定，状态为 `draft_adapter_smoke_readiness_defined`。后续若转向 durable store，应在 selector implementation entry review 或 schema artifact materialization review 中选择一个独立专题；任何 durable persistence、public production API、database、OIDC、repository adapter 或 executor 仍必须作为独立专题和 task card 推进。

## 验收方式

- Go 单元测试覆盖 save / read / validate / list 成功和失败路径。
- Consumer 能区分 sample / unsaved draft 与 saved draft record。
- Workspace Home 能读取 saved dev draft sanitized summary，并通过 read route 恢复到 Draft Designer；empty / failure state 不回退 sample。
- Draft Designer 本地结构编辑后的 active draft 能继续 validate / save / read，validation inspector、execution plan preview 和 runtime readiness inspector 能消费当前 active draft。
- Draft Designer 节点属性编辑后的 active draft 能保存和恢复节点级 summary、contract fields、output mapping 和 refs。
- Review Handoff 能把 active draft validation inspector、execution plan preview 和 runtime readiness inspector 汇总为 active draft review record，且不保存、不导出、不发送 handoff。
- Repository contract smoke 定义能覆盖 save / read / list I/O、failure mapping、no sample / memory dev fallback 和 no side effects，且不创建 smoke runner、repository interface、repository adapter 或数据库 artifact。
- Repository contract smoke runner readiness 能覆盖 future runner I/O、save / read / list operation runner matrix、failure mapping、no fallback 和 no side effects，且不创建 runner Go 文件、repository interface、repository adapter 或数据库 artifact。
- Repository contract smoke runner implementation 能覆盖 static runner、Go tests、failure mapping、no fallback 和 no side effects，且不创建 repository interface、repository adapter、selector、SQL、OIDC 或 production API artifact。
- Repository adapter implementation plan 能覆盖 future adapter file layout、save / read / list operation adapter matrix、failure mapping、no memory dev / sample fallback 和 no side effects，且不创建 repository interface、repository adapter、selector、SQL、OIDC 或 production API artifact。
- Schema artifact manifest 能覆盖 future manifest shape、section matrix、operation predicate coverage、failure mapping、no memory dev / sample fallback 和 no side effects，且不创建 migration root、manifest 文件、DDL review、rollback evidence、migration smoke、SQL、repository adapter、selector、OIDC 或 production API artifact。
- Adapter smoke readiness 能覆盖 future adapter smoke dependency gate、save / read / list operation adapter smoke matrix、failure mapping、no memory dev / sample / fixture / dev route / test auth fallback 和 no side effects，且不创建 adapter smoke fixture、adapter smoke checker、repository interface、repository adapter、selector、schema artifact、SQL、OIDC 或 production API artifact。
- Consumer 能区分 `version_conflict`，并在冲突时保留本地草案、展示当前版本 metadata，不把失败回退成 sample。
- route contract 和 consumer smoke checker 进入 fast baseline。
- Web build 和 workflow 相关聚合检查通过。
- `./scripts/check-repo.sh --fast` 通过。
- 若新增 committed schema、route contract、fixture 或 checker，先更新 task card，并按风险补全专项验证。

## 停止线

- 不实现 publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不接真实数据库、repository adapter、真实 schema migration、store selector、Radish OIDC middleware、token validation、API key lifecycle、quota enforcement、billing 或 public production API。
- 不把 `valid_for_review`、validation summary、risk summary 或 readiness summary 当作执行解锁条件。
