# Saved Workflow Draft Schema Artifact Manifest v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Schema Artifact Manifest v1` 承接 [Saved Workflow Draft Schema Artifact Evidence v1](saved-workflow-draft-schema-artifact-evidence-v1.md)、[Saved Workflow Draft Store Selector Smoke Readiness v1](saved-workflow-draft-store-selector-smoke-readiness-v1.md)、[Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md) 和 [Saved Workflow Draft Repository Adapter Implementation Plan v1](saved-workflow-draft-repository-adapter-implementation-plan-v1.md)，用于在创建 migration root、manifest 文件、DDL review、rollback evidence、migration smoke、SQL migration、migration runner 或 repository adapter 前，固定 future schema artifact manifest 的 shape、section matrix、operation predicate coverage、failure mapping 和停止线。

本专题只定义 manifest 准入契约，不创建 migration root、manifest 文件、DDL review artifact、rollback evidence artifact、migration smoke artifact、SQL migration、schema version table、migration runner、真实数据库、repository interface、repository adapter、store selector、selector smoke fixture、Radish OIDC middleware、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_schema_artifact_manifest_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已有 memory dev store、dev-only save / read / validate / list、saved draft restore、本地结构编辑、节点属性编辑和 active draft review handoff。
- `Saved Workflow Draft Schema / Migration Preconditions v1` 已固定 `saved_workflow_draft_record` logical schema、index strategy、migration gate、failure mapping 和 artifact guard。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 future schema artifact manifest、DDL review、rollback evidence、migration smoke、logical entity / index mapping、failure mapping、no fallback 和 no side effects。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future selector smoke mode matrix、operation matrix、schema artifact failure matrix、failure mapping、no fallback 和 no side effects。
- `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已固定 static contract smoke runner 和 save / read / list operation runner matrix。
- `Saved Workflow Draft Repository Adapter Implementation Plan v1` 已固定 future adapter file layout、operation adapter matrix、schema migration artifact gate、no fallback 和 no side effects。

## Manifest Shape Contract

future manifest 路径固定为 `services/platform/migrations/workflow_saved_drafts/manifest.json`，但本批不创建该文件。进入 materialized artifact 前，manifest shape 必须包含：

- identity：manifest id、manifest version、schema artifact id、store schema version。
- logical entity：`saved_workflow_draft_record`、`schema_version`、`store_schema_version`。
- field mapping：消费 `tenant_ref`、`workspace_id`、`application_id`、`draft_id`、`owner_subject_ref`、sanitized payload、validation summary、blocked capability summary 和 audit fields。
- index mapping：覆盖 `saved_drafts_scope_lookup`、`saved_drafts_owner_list`、`saved_drafts_version_conflict`、`saved_drafts_status_list` 和 `saved_drafts_schema_version`。
- migration plan：manual apply gate、migration lock、schema version table、no startup auto-migration。
- evidence refs：DDL review、rollback evidence、migration smoke。
- operation predicate coverage：`SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord`、`ListWorkflowDraftRecords`。
- failure mapping smoke cases 和 no side effect counters。

manifest version 固定为 `workflow_saved_draft_schema_manifest_v1`；schema artifact id 固定为 `workflow_saved_draft_schema_artifact_v1`；store schema version 固定为 `saved_workflow_drafts_store_v1`。

## Operation Manifest Matrix

| operation | scope | manifest predicate coverage | required indexes |
| --- | --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | scope lookup、owner / actor audit、`draft_version` optimistic concurrency、store / payload schema preflight | `saved_drafts_scope_lookup`、`saved_drafts_version_conflict`、`saved_drafts_schema_version` |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | scope lookup、owner visibility、sanitized projection、store / payload schema preflight | `saved_drafts_scope_lookup`、`saved_drafts_schema_version` |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | workspace / application list predicate、owner predicate、status optional filter、sanitized summary projection | `saved_drafts_owner_list`、`saved_drafts_status_list`、`saved_drafts_schema_version` |

所有 operation 当前都保持 `manifest_file_created_now=false`、`sql_migration_allowed_now=false`、`adapter_implementation_allowed_now=false`，并且不得 fallback 到 memory dev store、sample、fixture、dev HTTP route 或 test auth。

## Failure Mapping

manifest 准入必须继续保持 fail-closed failure code：

- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`

同时必须保留既有业务 failure：

- `draft_version_conflict` 不能被改写成 migration failure。
- `draft_scope_denied` 不能返回草案主体。
- `draft_not_found` 不能回退 sample。
- `draft_store_unavailable` 不能因为 memory dev store 可用而变成成功。

数据库内部错误、DDL 细节或 migration backend 原始错误不得泄露到 consumer response。

## No Side Effects

本批只能读取 committed fixture、文档和源码树。必须保持：

- 不写 repository 或数据库。
- 不创建 migration root、manifest、DDL review、rollback evidence、migration smoke 或 SQL 文件。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。
- 不读取 materialized result。
- 不启动服务、不连接数据库、不调用 OIDC。

## 后续准入

本专题完成后，durable store 方向仍必须继续按独立专题推进。后续可选方向：

1. `Saved Workflow Draft Store Selector Implementation Entry Review v1`：评审是否打开 formal config、selector 函数、selector tests 和 selector smoke fixture。
2. `Saved Workflow Draft Adapter Smoke Readiness v1`：定义 future adapter smoke 如何消费 static runner、schema artifact manifest、selector smoke 和 auth context。
3. `Saved Workflow Draft Schema Artifact Materialization Review v1`：另行评审是否创建 migration root、manifest、DDL review、rollback evidence 和 migration smoke artifact；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 `workflow-saved-draft-schema-artifact-manifest-v1` fixture / checker，固定依赖证据、manifest shape、section matrix、operation manifest matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-repository-adapter-implementation-plan-v1` 之后。
- 本批至少运行专项 checker、schema artifact evidence checker、repository adapter implementation plan checker 和 fast repo check。

## 停止线

- 不创建 migration root、manifest 文件、DDL review artifact、rollback evidence artifact、migration smoke artifact、SQL migration、schema version table、migration runner、数据库连接、repository interface、repository adapter、store selector、selector smoke fixture、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_schema_artifact_manifest_defined` 解释为 schema artifact file ready、DDL ready、migration ready、database ready、repository adapter ready、store selector ready、OIDC ready 或 production ready。
