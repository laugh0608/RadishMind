# Control Plane Read Repository Contract Smoke Runner Implementation v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-contract-smoke-runner-implementation-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_smoke_runner_implemented`

## 目标

在不进入数据库、repository adapter 或 store selector 的前提下，实现静态 repository contract smoke runner。runner 只消费 `controlPlaneReadRepositoryRouteTypeContracts()` 和既有 `control-plane-read-repository-contract-smoke-v1` 的 route smoke 口径，校验七条 read route 的 operation、request / result type、filter / sort allowlist、failure mapping、no fake fallback 和 no side effects。

本任务卡中的 `repository contract smoke runner implementation` 只表示静态 contract/type runner 已落地，不表示 repository interface、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。

## 范围

- 创建 `services/platform/internal/httpapi/control_plane_read_repository_contract_smoke_runner.go`。
- 创建 `services/platform/internal/httpapi/control_plane_read_repository_contract_smoke_runner_test.go`。
- runner 默认从 `controlPlaneReadRepositoryRouteTypeContracts()` 读取七条 route type matrix。
- runner 静态对齐既有 repository contract smoke fixture 的 route case、failure code、no fake fallback 和 no side effects。
- 新增 implementation fixture 和 checker，并接入仓库 fast baseline。

## 不在本次范围

- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `GOCACHE=/Users/luobo/Code/RadishMind/tmp/go-build go test ./internal/httpapi`（在 `services/platform` 下执行）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-runner-implementation-v1.py`
- `./scripts/check-repo.sh --fast`
