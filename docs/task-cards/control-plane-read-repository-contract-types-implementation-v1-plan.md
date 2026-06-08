# Control Plane Read Repository Contract Types Implementation v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-contract-types-implementation-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_types_implemented`

## 目标

在不进入数据库和 repository adapter 的前提下，创建未来 read store repository contract 使用的 Go 类型边界：`ReadRepositoryContext`、七条 read route 的 request / result type、failure code、projection / filter / sort type 和 route type matrix。

本任务卡中的 `repository contract types implementation` 只表示 Go contract type 已落地，不表示 repository interface、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer 已实现。

## 范围

- 创建 `services/platform/internal/httpapi/control_plane_read_repository_contract.go`。
- 创建 `services/platform/internal/httpapi/control_plane_read_repository_contract_test.go`。
- 落地 `ReadRepositoryContext`、shared request/result 类型、failure code 和七条 route request / result type。
- 落地 route type matrix，用于后续 contract smoke runner 和 repository implementation 前置校验。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `GOCACHE=/Users/luobo/Code/RadishMind/tmp/go-build go test ./internal/httpapi`（在 `services/platform` 下执行）
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-types-implementation-v1.py`
- `./scripts/check-repo.sh --fast`
