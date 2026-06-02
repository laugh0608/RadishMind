# Control Plane Read Schema Migration Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-schema-migration-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`schema_migration_readiness_defined`

## 目标

在 read store repository implementation 前固定 schema migration readiness：schema ownership、migration layout、rollback plan、tenant index strategy、read-only role policy、migration smoke、failure mapping、no side effects 和文档停止线。该切片只定义 schema / migration 前置条件，不创建 migration 目录、不写 SQL、不实现 repository interface、repository adapter、store selector、真实数据库、Radish OIDC、token validation 或 production API consumer。

本任务卡中的 `schema migration readiness` 只表示 schema / migration 前置条件已定义，不表示 database schema ready 或 migration files created。

## 范围

- 固定七类 read-side logical entity 的 owner、source of truth、tenant boundary 和 write boundary。
- 固定未来 migration root、命名规则和人工 review 口径，但不创建目录或 SQL 文件。
- 固定 rollback、backup、migration lock、schema version table 和 no startup auto-migration 策略。
- 固定 read-only role policy：runtime role 不具备 DDL、migration、insert、update 或 delete 权限。
- 固定七条 read route 的 schema migration readiness matrix。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不实现 repository interface、repository adapter、store selector、database query、database connection 或 migration runner。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-migration-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-store-selection-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
