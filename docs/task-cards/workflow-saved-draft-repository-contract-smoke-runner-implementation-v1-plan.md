# Workflow Saved Draft Repository Contract Smoke Runner Implementation v1 任务卡

更新时间：2026-06-16

## 任务定位

任务 ID：`workflow-saved-draft-repository-contract-smoke-runner-implementation-v1`

状态：`draft_repository_contract_smoke_runner_implemented`

本任务承接 `Saved Workflow Draft Repository Contract Smoke Runner Readiness v1`，用于实现 `SavedWorkflowDraftRepositoryContractSmokeRunner` 的 dev-only static runner 和 Go 单测。它只把已冻结的 repository contract smoke、operation contract、auth context、schema artifact gate、selector smoke gate 和 side effect policy 变成可复验代码结构，不创建 repository adapter 或 durable store。

## 输入

- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-runner-readiness-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-v1.md`
- `docs/features/workflow/saved-workflow-draft-repository-contract-preconditions-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.json`
- `scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-v1.json`

## 输出

- 新增 `docs/features/workflow/saved-workflow-draft-repository-contract-smoke-runner-implementation-v1.md`。
- 新增 `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner.go`。
- 新增 `services/platform/internal/httpapi/workflow_saved_draft_repository_contract_smoke_runner_test.go`。
- 新增 `scripts/checks/fixtures/workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.json`。
- 新增 `scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py`。
- 将 checker 接入 `scripts/check-repo.py` fast baseline。
- 同步旧 smoke / readiness checker 的 implementation gate 放行规则。
- 同步 `docs/features/README.md`、`docs/features/workflow/README.md`、`docs/features/workflow/saved-workflow-draft-v1.md`、`docs/radishmind-current-focus.md`、`scripts/README.md` 和周志。

## 实现范围

runner 名称固定为 `SavedWorkflowDraftRepositoryContractSmokeRunner`，覆盖：

- `SaveWorkflowDraftRecord`
- `ReadWorkflowDraftRecord`
- `ListWorkflowDraftRecords`

runner 必须验证 operation case 与 repository contract smoke fixture 对齐，必须保留 failure code 映射，必须证明 no fallback / no side effects。Go test 必须覆盖成功 report、fixture matrix 对齐、invalid context fail-closed 和缺失 scope grant fail-closed。

## 不在范围

- 不创建 `SavedWorkflowDraftRepository` interface、repository adapter、store selector、formal config entry、SQL migration、真实数据库、Radish OIDC middleware、token validation 或 production API consumer。
- 不实现新的 saved draft list、durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
- 不把 static runner 作为 HTTP route、CLI、真实 smoke runner 或 production readiness signal 暴露。

## 验收方式

本任务完成后至少运行：

```bash
cd services/platform
go test ./internal/httpapi
cd ../..
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-implementation-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-runner-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-contract-smoke-v1.py
cd apps/radishmind-web
npm run build
cd ../..
./scripts/check-repo.sh --fast
```

若后续批次创建 repository adapter、selector、SQL、OIDC 或 production API，应新增独立专题、任务卡和验证记录，不复用本任务状态提升实现结论。

## 停止线

- `draft_repository_contract_smoke_runner_implemented` 只表示 static runner 和测试已落地。
- 不声明 `repository_interface_implemented`、`repository_adapter_ready`、`durable_store_implemented`、`database_schema_ready`、`store_selector_implemented`、`radish_oidc_ready`、`production_api_consumer_ready` 或 `production_ready`。
