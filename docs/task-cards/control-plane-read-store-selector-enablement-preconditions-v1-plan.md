# Control Plane Read Store Selector Enablement Preconditions v1 任务卡

## 任务标识

- 切片：`control-plane-read-store-selector-enablement-preconditions-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`store_selector_enablement_preconditions_defined`

## 目标

在 store selection readiness 与 repository adapter implementation readiness refresh 已完成后，先固定未来 read store selector 启用前置条件。该切片只定义 selector enablement gates：未来配置键、默认 fake store、reserved disabled modes、unknown mode fail-closed、failure mapping、no fake fallback 和 no side effects。

本任务卡不创建正式配置入口，不实现 `SelectControlPlaneReadStore`，不创建 selector 文件，不实现 repository adapter、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

## 范围

- 新增 store selector enablement preconditions fixture 和 checker。
- 固定 `fixture_fake_store`、`database`、`postgres`、`repository` 与 `unknown` 五类 selector mode 的启用前置。
- 固定七条 read route 的 selector enablement matrix。
- 固定 schema migration、repository adapter、production auth 和 selector smoke 仍为 `not_satisfied`。
- 固定 no selector / config / adapter / SQL leak 检查，并接入仓库 fast baseline。

## 不在本次范围

- 不创建 `control_plane_read_store_selector.go`。
- 不新增正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 运行配置入口。
- 不实现 `SelectControlPlaneReadStore` 或 `ControlPlaneReadStoreSelector`。
- 不创建 `control_plane_read_repository_interface.go`、`control_plane_read_repository_adapter.go` 或 adapter test。
- 不实现 repository adapter、database query、database connection、schema 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-store-selector-enablement-preconditions-v1.py`
- `./scripts/check-repo.sh --fast`
