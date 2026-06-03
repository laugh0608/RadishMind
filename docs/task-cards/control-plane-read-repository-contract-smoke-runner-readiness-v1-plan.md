# Control Plane Read Repository Contract Smoke Runner Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-contract-smoke-runner-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_smoke_runner_readiness_defined`

## 目标

在 repository contract types 已落地后，先固定未来 smoke runner 如何消费 `controlPlaneReadRepositoryRouteTypeContracts()` 与既有 `ControlPlaneReadRepositoryContractSmoke` 定义。该切片只定义 runner readiness：输入、输出、route matrix、failure mapping、no fake fallback 和 no side effects。

本任务卡不实现 smoke runner，也不声明 repository interface、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已可用。

## 范围

- 新增 smoke runner readiness fixture 和 checker。
- 固定 future runner 名称、future 文件落点和 type catalog 消费点。
- 固定七条 read route 的 runner matrix，要求 request / result type 与 repository contract smoke operation 对齐。
- 固定 failure mapping、no fake fallback、side effect counter 和停止线。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不创建 `control_plane_read_repository_contract_smoke_runner.go`。
- 不实现 `ControlPlaneReadRepositoryContractSmokeRunner` 或任何 runner 函数。
- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-smoke-runner-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
