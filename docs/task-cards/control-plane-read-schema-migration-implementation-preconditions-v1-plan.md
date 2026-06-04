# Control Plane Read Schema Migration Implementation Preconditions v1 任务卡

## 任务标识

- 切片：`control-plane-read-schema-migration-implementation-preconditions-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`schema_migration_implementation_preconditions_defined`

## 目标

在创建任何 read store migration 目录、SQL、schema version 记录或 migration runner 之前，先固定 schema migration implementation preconditions：future artifact manifest、DDL review gate、rollback fixture gate、schema version smoke、tenant index smoke、read-only role smoke、failure mapping、no fake fallback、no side effects 和停止线。

本任务卡只定义 migration 实现准入，不表示 database schema ready、migration files created、migration runner ready 或 repository adapter ready。

## 范围

- 固定未来 migration root、manifest、up / down SQL artifact 和 runner 名称，但不创建这些文件。
- 固定 migration artifact manifest 必须包含的字段。
- 固定 schema version、tenant index、read-only role、projection 和 failure mapping smoke 前置条件。
- 固定七条 read route 在 migration implementation 前必须覆盖的 smoke 矩阵。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不创建 migration 目录、SQL 文件、schema 文件、database fixture 或 migration runner。
- 不实现 repository interface、repository adapter、store selector、database query、database connection 或 migration apply 流程。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-migration-implementation-preconditions-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-migration-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
