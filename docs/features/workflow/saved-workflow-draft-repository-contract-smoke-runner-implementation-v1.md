# Saved Workflow Draft Repository Contract Smoke Runner Implementation v1 专题

更新时间：2026-06-16

## 专题定位

`Saved Workflow Draft Repository Contract Smoke Runner Implementation v1` 承接 [Saved Workflow Draft Repository Contract Smoke Runner Readiness v1](saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md)，把 future `SavedWorkflowDraftRepositoryContractSmokeRunner` 落成 dev-only static runner 和 Go 单测。

本专题只实现静态 contract smoke runner：它消费已固定的 repository operation contract、repository contract smoke matrix、auth context、schema artifact gate、selector smoke gate 和 no side effects policy，并输出 advisory-only smoke report。它不声明 `SavedWorkflowDraftRepository` interface，不创建 repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API、publish、run、executor、confirmation、writeback 或 replay。

状态：`draft_repository_contract_smoke_runner_implemented`

## 当前输入事实

- `Saved Workflow Draft Repository Contract Smoke v1` 已固定 save / read / list I/O、operation smoke matrix、failure mapping、no fallback 和 no side effects。
- `Saved Workflow Draft Repository Contract Smoke Runner Readiness v1` 已固定 runner boundary、runner I/O、operation runner matrix、failure mapping、no fallback、no side effects 和 artifact guard。
- `Saved Workflow Draft Repository Contract Preconditions v1`、`Saved Workflow Draft Auth Context Preconditions v1`、`Saved Workflow Draft Schema Artifact Evidence v1` 和 `Saved Workflow Draft Store Selector Smoke Readiness v1` 已提供 runner 需要消费的上游证据。
- 当前仍没有 repository interface、repository adapter、store selector implementation、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。

## 实现内容

新增 static runner：

- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go`
- `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner_test.go`

runner 固定以下结构：

- `SavedWorkflowDraftRepositoryActorContext`
- `SavedWorkflowDraftRepositoryContractSmokeRunner`
- `SavedWorkflowDraftRepositoryContractSmokeCase`
- `SavedWorkflowDraftRepositoryContractSmokeReport`
- `SavedWorkflowDraftRepositoryContractSmokeSideEffectReport`

runner 覆盖 `SaveWorkflowDraftRecord`、`ReadWorkflowDraftRecord` 和 `ListWorkflowDraftRecords`。成功路径只证明 operation case 与 committed smoke fixture / operation contract 对齐；失败路径保留 `draft_scope_denied`、`draft_not_found`、`draft_schema_version_unsupported`、`draft_payload_invalid`、`draft_version_conflict`、`draft_store_unavailable`、`draft_store_contract_mismatch`、`repository_store_disabled`、`invalid_draft_store_mode`、`draft_auth_context_contract_mismatch`、`draft_schema_migration_not_applied`、`draft_store_schema_version_mismatch` 和 `draft_store_migration_unavailable`。

## Advisory Boundary

该 runner 是 implementation evidence，不是执行入口：

- 不启动服务。
- 不读取或写入数据库。
- 不调用 dev HTTP route。
- 不 fallback 到 sample、fixture、memory dev store 或 test auth。
- 不产生 repository write、workflow execution、confirmation decision、business writeback 或 replay side effect。
- 不把 contract smoke report 解释为 publish ready、run ready、repository adapter ready、OIDC ready 或 production ready。

## 验收方式

- 新增 implementation fixture / checker 固定 runner 源码、Go test、operation runner matrix、failure mapping、no fallback、no side effects 和 forbidden source literals。
- 更新 repository contract smoke / runner readiness checker，让它们只在 implementation fixture 满足时放行 static runner 文件，并继续拦截 repository interface、repository adapter、selector、SQL、OIDC 和 production API artifact。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 runner readiness 之后、product surface trigger recheck 之前。
- 本批至少运行 Go 定向测试、implementation checker、runner readiness checker、repository contract smoke checker、web build 和 `./scripts/check-repo.sh --fast`。

## 后续准入

本专题完成后仍不能直接连接真实数据库。后续若继续 durable store 方向，应选择一个独立专题：

1. repository adapter implementation plan：消费本 runner、schema artifact evidence、auth context、selector readiness 和 repository contract smoke，先固定 adapter 的实现计划、迁移门禁和 rollback 证据。
2. selector implementation entry review：决定 formal config、selector 函数、selector tests 和 selector smoke fixture 的准入条件。

## 停止线

- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、SQL migration、真实数据库、Radish OIDC middleware、token validation、production API consumer 或新的 saved draft list 实现。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 `draft_repository_contract_smoke_runner_implemented` 解释为 repository adapter ready、durable store ready、database ready、selector ready、OIDC ready 或 production ready。
