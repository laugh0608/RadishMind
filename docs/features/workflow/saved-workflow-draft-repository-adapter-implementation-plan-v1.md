# Saved Workflow Draft Repository Adapter Implementation Plan v1 专题

更新时间：2026-06-17

## 专题定位

`Saved Workflow Draft Repository Adapter Implementation Plan v1` 承接 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md)，用于在进入 durable persistence 前固定 future repository adapter 的实现计划、依赖证据、门禁矩阵和失败语义。

本专题只定义 implementation plan，不创建 repository interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_adapter_implementation_plan_defined`

## 当前输入事实

- `Saved Workflow Draft v1` 已完成 dev-only save / read / validate / list、saved draft list / restore、本地结构编辑、节点属性编辑和 active draft review handoff。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 future `SavedWorkflowDraftRepository` 的 actor context、save / read / list operation、request / result、failure 和 sanitized projection。
- `Saved Workflow Draft Schema / Migration Preconditions v1`、`Saved Workflow Draft Auth Context Preconditions v1`、`Saved Workflow Draft Store Selector Enablement Preconditions v1`、`Saved Workflow Draft Schema Artifact Evidence v1` 和 `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future durable store 的 schema、auth、selector 和 artifact 证据链。
- `Saved Workflow Draft Repository Contract Smoke v1` 和 `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已固定 static contract smoke runner、failure mapping、no fallback 和 no side effects。

## Adapter Gate Matrix

future adapter implementation 必须消费以下门禁：

| gate | 当前状态 | 要求 |
| --- | --- | --- |
| repository contract + static runner | `satisfied` | 消费 save / read / list operation contract、contract smoke 和 static runner |
| schema / auth / selector evidence | `satisfied` | 消费 schema migration preconditions、auth context、schema artifact evidence 和 selector smoke readiness |
| selector implementation | `not_satisfied` | 后续独立实现 formal config、selector 函数、selector tests 和 selector smoke fixture |
| schema migration artifact | `not_satisfied` | 后续独立创建 manifest、DDL review、rollback evidence 和 migration smoke |
| production auth | `not_satisfied` | 后续独立接 Radish OIDC、claim mapping、workspace membership 和 scope projection |
| durable adapter smoke | `not_satisfied` | readiness 已按 `Saved Workflow Draft Adapter Smoke Readiness v1` 定义，真实 adapter smoke 执行仍需等待 selector、schema artifact、auth 和 adapter gate |
| implementation leak guard | `required_now` | 当前不得出现 interface、adapter、selector、SQL、migration 或 OIDC artifact |

## Operation Adapter Matrix

future adapter 只允许覆盖三类 operation：

| operation | scope | success projection | future adapter checks |
| --- | --- | --- | --- |
| `SaveWorkflowDraftRecord` | `workflow_drafts:write` | sanitized saved draft record + current version metadata | scope predicate、owner / actor audit、version compare-and-set、sanitized payload、schema preflight、contract mismatch |
| `ReadWorkflowDraftRecord` | `workflow_drafts:read` | sanitized saved draft record | scope predicate、owner visibility、sanitized projection、schema preflight、not found no sample fallback、contract mismatch |
| `ListWorkflowDraftRecords` | `workflow_drafts:read` | sanitized saved draft summary list | workspace / application list predicate、owner visibility、summary projection、stable ordering、schema preflight、empty list no sample fallback |

所有 operation 当前都保持 `adapter_implementation_allowed_now=false`，且不得 fallback 到 memory dev store、sample、fixture、dev HTTP route 或 test auth。

## Failure Mapping

future adapter plan 必须保留以下 fail-closed code：

- `draft_scope_denied`
- `draft_not_found`
- `draft_schema_version_unsupported`
- `draft_payload_invalid`
- `draft_version_conflict`
- `draft_store_unavailable`
- `draft_store_contract_mismatch`
- `repository_store_disabled`
- `invalid_draft_store_mode`
- `draft_auth_context_contract_mismatch`
- `draft_schema_migration_not_applied`
- `draft_store_schema_version_mismatch`
- `draft_store_migration_unavailable`

`draft_version_conflict` 不能被改写为 adapter failure；`draft_not_found` 不能回退 sample；`draft_scope_denied` 不能返回草案主体；数据库内部错误不能泄露为 raw database detail。

## No Side Effects

本计划阶段必须保持：

- 不写 repository 或数据库。
- 不调用 workflow executor、tool executor、confirmation decision、business writeback 或 replay。
- 不读取 materialized result。
- 不启动服务、不连接数据库、不调用 OIDC。

## 后续准入

本专题完成后，已继续补齐 [Saved Workflow Draft Schema Artifact Manifest v1](saved-workflow-draft-schema-artifact-manifest-v1.md) 和 [Saved Workflow Draft Adapter Smoke Readiness v1](saved-workflow-draft-adapter-smoke-readiness-v1.md)，状态分别为 `draft_schema_artifact_manifest_defined` 和 `draft_adapter_smoke_readiness_defined`。下一步仍不能直接创建 durable adapter。后续只能选择一个独立方向：

1. `Saved Workflow Draft Store Selector Implementation Entry Review v1`：评审 formal config、selector 函数、selector tests 和 selector smoke fixture 是否进入实现。
2. `Saved Workflow Draft Schema Artifact Materialization Review v1`：另行评审是否创建 migration root、manifest、DDL review、rollback evidence 和 migration smoke artifact；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 `workflow-saved-draft-repository-adapter-implementation-plan-v1` fixture / checker，固定依赖证据、future file layout、Adapter Gate Matrix、Operation Adapter Matrix、Failure Mapping、no fallback、no side effects 和 source artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 `workflow-saved-draft-repository-contract-smoke-runner-implementation-v1` 之后。
- 本批至少运行专项 checker、static runner implementation checker、Go saved draft 相关测试和 fast repo check。

## 停止线

- 不创建 repository interface、repository adapter、store selector、formal config entry、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或新的 saved draft list 实现。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_repository_adapter_implementation_plan_defined` 解释为 repository adapter ready、durable store ready、database ready、selector ready、OIDC ready 或 production ready。
