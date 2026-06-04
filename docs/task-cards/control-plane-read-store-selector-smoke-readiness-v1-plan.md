# Control Plane Read Store Selector Smoke Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-store-selector-smoke-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`store_selector_smoke_readiness_defined`

## 目标

在实现任何 read store selector、正式 read-store 配置入口或 selector smoke fixture 之前，先固定 store selector smoke readiness 的可检查治理边界。该切片只定义未来 selector smoke contract、模式矩阵、七条 route selector smoke matrix、reserved mode fail-closed、unknown mode fail-closed、no fake fallback、no side effects 和停止线。

本任务卡不表示 store selector smoke ready、store selector implemented、formal read store config ready、repository adapter ready、database schema ready 或 production API consumer ready。

## 范围

- 新增 store selector smoke readiness fixture 和 checker。
- 消费 `control-plane-read-store-selector-enablement-preconditions-v1`、`control-plane-read-schema-artifact-manifest-readiness-v1` 和 `control-plane-read-repository-adapter-implementation-plan-v1`。
- 固定未来 selector smoke 必须覆盖 unset / `fixture_fake_store`、`database`、`postgres`、`repository` 和 unknown mode。
- 固定七条 read route 在 selector smoke 中必须覆盖 default mode、reserved mode failure、unknown mode failure、adapter smoke dependency、no fake fallback 和 no side effects。
- 将 checker 接入仓库 fast baseline，并同步 read-side contract、current focus、roadmap、capability matrix、integration contracts、脚本说明、platform README 和周志。

## 不在本次范围

- 不创建 `control_plane_read_store_selector.go`、selector test、selector smoke fixture 或 selector smoke checker。
- 不新增正式 `RADISHMIND_CONTROL_PLANE_READ_STORE` 配置入口。
- 不实现 store selector、repository adapter、database query、database connection、SQL、migration 或 migration runner。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-store-selector-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-artifact-manifest-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
