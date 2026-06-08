# Control Plane Read Adapter Smoke Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-adapter-smoke-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`adapter_smoke_readiness_defined`

## 目标

在创建任何 durable repository adapter、repository interface、adapter smoke fixture 或真实数据库 read path 之前，先固定 adapter smoke readiness 的可检查治理边界。该切片只定义未来 durable adapter smoke 如何消费 schema artifact manifest readiness、store selector smoke readiness、production auth readiness、static runner 和 repository adapter implementation plan。

本任务卡不表示 adapter smoke ready、repository adapter ready、database query ready、production auth ready、production API consumer ready 或 production ready。

## 范围

- 新增 adapter smoke readiness fixture 和 checker。
- 消费 `control-plane-read-repository-adapter-implementation-plan-v1`、`control-plane-read-schema-artifact-manifest-readiness-v1`、`control-plane-read-store-selector-smoke-readiness-v1`、`control-plane-read-production-auth-readiness-v1` 和 `control-plane-read-repository-contract-smoke-runner-implementation-v1`。
- 固定未来 adapter smoke 必须覆盖七条 read route，并与 adapter plan、static runner、schema artifact readiness、selector smoke readiness 和 production auth readiness 的 route matrix 对齐。
- 固定 adapter smoke 前仍未满足的 gate：schema artifact manifest、selector smoke、production auth、repository adapter implementation、adapter smoke execution 和 production API consumer。
- 固定 failure mapping、no fake fallback、no side effects 和 no implementation leak 检查。
- 将 checker 接入仓库 fast baseline，并同步 read-side contract、current focus、roadmap、capability matrix、integration contracts、脚本说明、platform README 和周志。

## 不在本次范围

- 不创建 `control-plane-read-adapter-smoke-v1.json` 或 adapter smoke checker。
- 不创建 `control_plane_read_repository_interface.go`、`control_plane_read_repository_adapter.go`、adapter test 或 adapter contract smoke test。
- 不实现 repository interface、repository adapter、store selector、database query、database connection、SQL、migration 或 migration runner。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-production-auth-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
