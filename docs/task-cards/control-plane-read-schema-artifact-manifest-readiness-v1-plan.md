# Control Plane Read Schema Artifact Manifest Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-schema-artifact-manifest-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`schema_artifact_manifest_readiness_defined`

## 目标

在创建任何 read store schema artifact、DDL、migration manifest、rollback fixture、schema smoke artifact 或 durable adapter smoke 之前，先固定 schema artifact manifest readiness 的可检查治理边界。该切片只定义未来 manifest 合同、DDL review 证据、rollback fixture 证据、schema version smoke、tenant index smoke、read-only role smoke、durable adapter smoke dependency、failure mapping、no fake fallback、no side effects 和停止线。

本任务卡不表示 schema artifact manifest ready、DDL review ready、rollback fixture ready、schema version smoke ready、tenant index smoke ready、read-only role smoke ready、database schema ready 或 repository adapter ready。

## 范围

- 新增 schema artifact manifest readiness fixture 和 checker。
- 固定未来 schema artifact manifest 必须包含的字段与证据引用，但不创建 manifest。
- 固定未来 DDL review、rollback fixture、schema version smoke、tenant index smoke、read-only role smoke 和 durable adapter smoke dependency 的 `not_satisfied` gate。
- 固定七条 read route 的 schema artifact matrix：logical entity、manifest、DDL review、rollback、schema version、tenant index、read-only role、adapter smoke dependency、no fake fallback 和 no side effects。
- 将 checker 接入仓库 fast baseline，并同步 read-side contract、current focus、roadmap、capability matrix、integration contracts、脚本说明、platform README 和周志。

## 不在本次范围

- 不创建 migration 目录、manifest、SQL 文件、DDL review artifact、rollback fixture、schema smoke artifact 或 database fixture。
- 不实现 migration runner、repository interface、repository adapter、store selector、database query、database connection 或 migration apply 流程。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-artifact-manifest-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-adapter-implementation-plan-v1.py`
- `./scripts/check-repo.sh --fast`
