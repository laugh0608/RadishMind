# Control Plane Read Repository Interface Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-interface-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_interface_readiness_defined`

## 目标

在静态 contract/type runner 已落地后，先固定未来 `ControlPlaneReadRepository` interface 的 method matrix 和 adapter 实现前置条件。该切片只定义 repository interface readiness：七条 read route 的 operation、request / result type、summary type、context 约束、failure mapping、no side effects 和后续 adapter gate。

本任务卡不声明 `ControlPlaneReadRepository` interface，不创建 repository interface 文件，不实现 repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

## 范围

- 新增 repository interface readiness fixture 和 checker。
- 固定未来 interface 名称、future 文件落点和 method matrix。
- 要求 method matrix 消费 `control-plane-read-repository-contract-types-implementation-v1` 与 `control-plane-read-repository-contract-smoke-runner-implementation-v1`。
- 固定 adapter implementation gate、production auth gate、failure mapping、no side effects 和 no interface / adapter leak 检查。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不创建 `control_plane_read_repository_interface.go`。
- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-interface-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
