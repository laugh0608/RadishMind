# Saved Workflow Draft Repository Contract Smoke Runner Readiness v1 专题

更新时间：2026-06-16

## 专题定位

`Saved Workflow Draft Repository Contract Smoke Runner Readiness v1` 承接 [Saved Workflow Draft Repository Contract Smoke v1](saved-workflow-draft-repository-contract-smoke-v1.md)，用于在创建 smoke runner、repository interface、repository adapter、store selector、数据库或 production API 前，固定 future `SavedWorkflowDraftRepositoryContractSmokeRunner` 应如何消费 repository contract smoke、operation contract、auth context、schema artifact gate、selector smoke gate 和 side effect probe。

本专题只定义 smoke runner readiness。后续 [Saved Workflow Draft Repository Contract Smoke Runner Implementation v1](saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md) 已按本专题实现 static runner 和 Go test，但本专题本身仍不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、SQL migration、Radish OIDC middleware、token validation、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_contract_smoke_runner_readiness_defined`

## 当前输入事实

- `Saved Workflow Draft Repository Contract Smoke v1` 已固定 future `SavedWorkflowDraftRepositoryContractSmoke` 的 I/O、save / read / list operation matrix、failure mapping、no fallback、no side effects 和 artifact guard。
- `Saved Workflow Draft Repository Contract Preconditions v1` 已固定 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords` 的 request / result contract。
- `Saved Workflow Draft Auth Context Preconditions v1` 已固定 future `SavedWorkflowDraftRepositoryActorContext` 的身份来源、workspace membership、owner policy、scope grants 和 audit / sanitization 边界。
- `Saved Workflow Draft Schema Artifact Evidence v1` 已固定 schema artifact failure mapping 和 migration gate 证据。
- `Saved Workflow Draft Store Selector Smoke Readiness v1` 已固定 future selector smoke 的 mode matrix、operation matrix、schema artifact failure、no fallback 和 no side effects。
- `Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 已固定 `draft_repository_contract_smoke_runner_implemented`，实现 static runner 和 Go tests。
- 当前仍没有 repository interface、repository adapter、store selector implementation、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。

## Runner Boundary

future runner 名称固定为 `SavedWorkflowDraftRepositoryContractSmokeRunner`，future 文件落点固定为：

- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner_test.go`

本专题创建时不创建上述文件；后续 implementation 专题已经按该落点实现 static runner。runner readiness 仍只记录 future runner 的输入输出、operation runner matrix、failure mapping、no fallback、no side effects 和禁止提前出现的 repository / adapter / selector / SQL / OIDC artifact。

## Runner I/O Contract

future runner 输入必须包含：

- `repository_contract_smoke_fixture`
- `repository_actor_context`
- `draft_scope`
- `operation_requests`
- `failure_injections`
- `schema_artifact_gate`
- `selector_smoke_gate`
- `side_effect_probe`

future runner 输出必须包含：

- `operation_results`
- `failure_results`
- `contract_mismatch_report`
- `fallback_report`
- `side_effect_report`
- `summary`

runner 必须以 `SavedWorkflowDraftRepositoryActorContext` 为 context source，以 `SaveWorkflowDraftRecordRequest`、`ReadWorkflowDraftRecordRequest` 和 `ListWorkflowDraftRecordsRequest` 为 request source，以对应 result type 为 result source。成功路径只能比较 sanitized saved draft record 或 summary；失败路径只能比较预期 failure code，不暴露数据库内部细节。

## Operation Runner Matrix

future runner 必须覆盖：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

每个 operation runner case 都必须消费 repository operation contract、repository contract smoke fixture、auth context contract、schema artifact gate 和 selector smoke gate。任何 failure case 都不得 fallback 到 memory dev store、sample、fixture 或 dev HTTP route，也不得产生数据库写入、workflow execution、confirmation decision、business writeback 或 replay side effect。

## 后续准入

本专题之后仍不能直接创建 repository adapter。后续若继续 durable store 方向，应选择一个独立批次：

1. repository adapter implementation plan，必须消费 smoke runner readiness / implementation、schema artifact evidence、auth context、store selector readiness 和 repository contract smoke。
2. selector implementation entry review，另行决定是否创建 formal config、selector 函数、selector tests 和 selector smoke fixture；进入该批前仍不得连接真实数据库。

## 验收方式

- 新增 task card 固定本专题为 readiness-only。
- 新增 fixture / checker 固定 runner boundary、runner I/O、operation runner matrix、failure mapping、no fallback、no side effects 和 implementation artifact guard。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 repository contract smoke 之后。
- 本批至少运行专项 checker、repository contract smoke checker、store selector smoke readiness checker、web build 和 `./scripts/check-repo.sh --fast`。

## 停止线

- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或新的 saved draft list 实现。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 runner readiness 解释为 runner implemented、repository adapter ready、database ready、selector ready、OIDC ready 或 production ready。
