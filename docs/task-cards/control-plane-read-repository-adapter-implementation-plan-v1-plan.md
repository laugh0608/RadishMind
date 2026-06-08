# Control Plane Read Repository Adapter Implementation Plan v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-adapter-implementation-plan-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_adapter_implementation_plan_defined`

## 目标

在 repository adapter implementation readiness refresh、store selector enablement preconditions 和 schema migration implementation preconditions 都已固定后，定义未来 repository adapter implementation plan 的可检查治理边界。该切片只固定未来文件落点、依赖证据、接口 / contract type / static runner 消费、七条 read route adapter matrix、schema migration implementation preconditions 消费、selector gate、failure mapping、no fake fallback、no side effects 和文档停止线。

本任务卡不创建 adapter 或 interface 文件，不写 SQL，不创建 migration，不实现 repository adapter、store selector、真实数据库、Radish OIDC、token validation、production API consumer、API key lifecycle、quota enforcement、workflow executor、confirmation、writeback 或 replay。

## 范围

- 新增 repository adapter implementation plan fixture 和 checker。
- 固定未来 interface、adapter、adapter test、adapter contract smoke test、selector 和 migration root 的文件落点，但不创建这些文件。
- 要求 adapter plan 消费 `control-plane-read-repository-interface-readiness-v1`、repository contract types implementation、静态 runner implementation、`control-plane-read-store-selector-enablement-preconditions-v1` 和 `control-plane-read-schema-migration-implementation-preconditions-v1`。
- 固定七条 read route 的 future adapter matrix：operation、request / result type、schema migration preconditions、selector gate、failure mapping、no fake fallback 和 no side effects。
- 将 checker 接入仓库 fast baseline，并同步 read-side contract、current focus、roadmap、capability matrix、integration contracts、脚本说明、platform README 和周志。

## 不在本次范围

- 不创建 `control_plane_read_repository_interface.go`。
- 不创建 `control_plane_read_repository_adapter.go`、adapter test 或 adapter contract smoke test。
- 不创建 `control_plane_read_store_selector.go`。
- 不创建 migration 目录、manifest、SQL 文件、schema 文件或 database fixture。
- 不声明或实现 `ControlPlaneReadRepository` interface。
- 不实现 repository adapter、store selector、database query、database connection、schema、migration runner 或 repository migration。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-adapter-implementation-plan-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-migration-implementation-preconditions-v1.py`
- `./scripts/check-repo.sh --fast`
