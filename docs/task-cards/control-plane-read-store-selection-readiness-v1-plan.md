# Control Plane Read Store Selection Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-store-selection-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`store_selection_readiness_defined`

## 目标

在 read store selector 或真实 repository implementation 之前固定 store selection readiness：默认 source、保留 source、失败映射、no fake fallback、七条 route 的 selection matrix、no side effects 和文档停止线。该切片只定义选择策略前置条件，不创建正式配置入口，不实现 store selector、repository interface、repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

本任务卡中的 `store selection readiness` 只表示 read source 选择策略前置条件已定义，不表示 store selector implemented。

## 范围

- 固定默认 read source 仍为 `fixture-backed fake store`。
- 固定未来 `RADISHMIND_CONTROL_PLANE_READ_STORE` 只作为保留配置键，不在本切片创建正式启用入口。
- 固定 `database`、`postgres` 与 `repository` mode 当前必须 fail-closed 为 `database_read_disabled`。
- 固定未知 selector value 必须 fail-closed 为 `invalid_read_store_mode`。
- 固定七条 read route 的 store selection matrix、no fake fallback 和 no side effects。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不实现 store selector、repository interface、repository adapter、SQL query、database schema 或 migration。
- 不新增正式配置入口，也不启用 database / postgres / repository read mode。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-store-selection-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-implementation-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
