# Control Plane Read Repository Adapter Implementation Readiness Refresh v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-adapter-implementation-readiness-refresh-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_adapter_implementation_readiness_refreshed`

## 目标

在 repository contract types、静态 contract smoke runner 和 repository interface readiness 都已固定后，刷新未来 repository adapter implementation 的准入条件。该切片只把未来 adapter 的文件落点、依赖证据、七条 read route 覆盖、failure mapping、no fake fallback 和 no side effects 对齐为可检查 gate。

本任务卡不创建 `ControlPlaneReadRepository` interface，不创建 repository adapter 文件，不实现 store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

## 范围

- 新增 repository adapter implementation readiness refresh fixture 和 checker。
- 要求 adapter gate 消费 `control-plane-read-repository-interface-readiness-v1`、repository contract types implementation 和静态 runner implementation。
- 固定七条 read route 到未来 adapter 检查项的矩阵。
- 固定 schema migration、store selector enablement、production auth 和 adapter smoke 仍为 `not_satisfied`。
- 固定 no adapter / interface / SQL / selector leak 检查，并接入仓库 fast baseline。

## 不在本次范围

- 不创建 `control_plane_read_repository_interface.go`。
- 不创建 `control_plane_read_repository_adapter.go` 或 adapter test。
- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection、schema 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-adapter-implementation-readiness-refresh-v1.py`
- `./scripts/check-repo.sh --fast`
