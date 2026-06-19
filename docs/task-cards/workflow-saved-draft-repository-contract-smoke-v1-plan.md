# Workflow Saved Draft Repository Contract Smoke v1 任务卡

更新时间：2026-06-16

## 任务定位

任务 ID：`workflow-saved-draft-repository-contract-smoke-v1`

状态：`draft_repository_contract_smoke_defined`

本任务承接 `Saved Workflow Draft Repository Contract Preconditions v1`、`Saved Workflow Draft Auth Context Preconditions v1`、`Saved Workflow Draft Schema Artifact Evidence v1` 和 `Saved Workflow Draft Store Selector Smoke Readiness v1`，用于固定 future repository contract smoke 的静态检查记录。它是后续 smoke runner、repository adapter 或 selector implementation 之前的准入证据，不创建任何正式实现。

## 输入

- `docs/features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-auth-context-preconditions-v1.md`
- `docs/features/workflow/saved-workflow-draft-schema-artifact-evidence-v1.md`
- `docs/features/workflow/saved-workflow-draft-store-selector-smoke-readiness-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-repository-contract-preconditions-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-schema-artifact-evidence-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-store-selector-smoke-readiness-v1.json`

## 输出

- 新增 `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-v1.md`。
- 新增 `scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json`。
- 新增 `scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-v1.py`。
- 将 checker 接入 `scripts/check-repo.py` fast baseline。
- 同步 `docs/features/README.md`、`docs/features/workflow/README.md`、`docs/features/workflow/saved-workflow-draft-v1.md`、`docs/radishmind-current-focus.md`、`scripts/README.md` 和周志。

## Contract Smoke Scope

future harness 名称固定为 `SavedWorkflowDraftRepositoryContractSmoke`。

future smoke 必须覆盖：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

future smoke 必须消费结构化 `repository_actor_context`，并覆盖 draft scope、request / result field、scope denied、not found、version conflict、store unavailable、contract mismatch、selector failure、schema artifact failure、no sample fallback、no memory dev fallback 和 no side effects。

## 不在范围

- 不创建 repository contract smoke runner、runner fixture 或 Go runner test。
- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、formal config entry、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现新的 saved draft list、durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。

## 验收方式

本任务完成后至少运行：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-store-selector-smoke-readiness-v1.py
./scripts/check-repo.sh --fast
```

若后续批次创建 smoke runner、repository adapter、SQL、OIDC 或 production API，应新增独立任务卡和验证记录，不复用本任务状态提升实现结论。

## 停止线

- `draft_repository_contract_smoke_defined` 只表示静态 smoke contract 已固定。
- 不声明 `repository_contract_smoke_runner_implemented`、`repository_interface_implemented`、`repository_adapter_ready`、`durable_store_implemented`、`database_schema_ready`、`store_selector_implemented`、`radish_oidc_ready`、`production_api_consumer_ready` 或 `production_ready`。
