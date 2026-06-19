# Control Plane Durable Read Foundation v1 任务卡

## 任务标识

- 切片：`control-plane-durable-read-foundation-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`durable_read_foundation_implemented`

## 目标

在 Image Path metadata-only response builder runtime 接线完成后，把 read-side 主线推进到受控的数据访问边界：声明 `ControlPlaneReadRepository` interface 边界，并让现有七条 fake-store-backed read handlers 通过该 interface 消费数据。

本切片只把 fake store 包进 repository interface，不引入 durable adapter。HTTP response envelope、fake auth、七条 read route 和现有产品样例必须保持行为稳定。

## 范围

1. 新增 `ControlPlaneReadRepository` interface 边界。
2. 复用已落地的 `ReadRepositoryContext`、route-specific request / result type 和静态 contract runner 证据。
3. 让 `tenant-summary-route`、`application-summary-list-route`、`api-key-summary-list-route`、`quota-summary-route`、`workflow-definition-summary-list-route`、`run-record-summary-list-route` 与 `audit-summary-list-route` 通过 repository interface 读取。
4. 当前 repository implementation 只包裹 `fixtureBackedControlPlaneReadStore`。
5. 保持现有 response fixture JSON shape、failure envelope 和 no-side-effects 口径。
6. 新增 checker 并接入仓库检查。

## 不在本次范围

- 不创建 repository adapter。
- 不创建 store selector 或正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 配置入口。
- 不创建 SQL、migration、migration runner、schema artifact manifest 或 database fixture。
- 不接真实数据库、Radish OIDC、token validation、auth middleware 或 production API consumer。
- 不实现 API key lifecycle、quota enforcement、rate limit、billing 或 cost ledger。
- 不实现 workflow builder、workflow executor、confirmation decision、business writeback、run replay、run resume 或 materialized result reader。

## 验收

- `services/platform/internal/httpapi/control_plane_read_repository.go` 声明 repository interface 边界和 fake-store repository bridge。
- `services/platform/internal/httpapi/control_plane_read.go` 的七条 read handler 通过 `s.controlPlaneReadRepository()` 读取数据，不再直接读取 fake store。
- `services/platform/internal/httpapi/control_plane_read_repository_contract.go` 的 typed summary 覆盖现有 response fixture 字段，避免 interface 化造成 JSON shape 漂移。
- `services/platform/internal/httpapi/control_plane_read_test.go` 覆盖 handler 通过 repository interface 读取的路径。
- `scripts/checks/fixtures/control-plane-durable-read-foundation-v1.json` 和 `scripts/checks/control_plane/check-control-plane-durable-read-foundation-v1.py` 固定边界、实现证据和停止线。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-durable-read-foundation-v1.py`
- `go test ./...`（在 `services/platform/` 下）
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`
