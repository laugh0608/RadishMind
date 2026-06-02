# Control Plane Read Disabled Database Guard v1 任务卡

## 任务标识

- 切片：`control-plane-read-disabled-database-guard-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`disabled_database_guard_defined`

## 目标

在 repository implementation 之前固定 disabled database read guard：如果未来配置或调用尝试进入 database / postgres / repository read mode，当前阶段必须 fail-closed 为 `database_read_disabled`，不得静默回退到 fixture-backed fake store，也不得声明真实 read repository、SQL、migration 或生产 auth 已可用。

## 范围

- 固定七条 read route 的 database read guard matrix。
- 固定 reserved database mode input 的禁用状态和失败码。
- 固定 `database_read_disabled`、`read_store_unavailable` 与 `read_store_contract_mismatch` 的 guard failure mapping。
- 固定后续 smoke 必须覆盖七条 route、无 fake fallback、无副作用、无 raw secret / raw payload 输出。
- 将 checker 接入仓库 fast baseline，保证该停止线不会被后续实现误放宽。

## 不在本次范围

- 不实现真实数据库 schema、SQL query、migration 或 repository adapter。
- 不新增正式配置入口，也不启用 `RADISHMIND_CONTROL_PLANE_READ_STORE`。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-disabled-database-guard-v1.py`
- `./scripts/check-repo.sh --fast`
