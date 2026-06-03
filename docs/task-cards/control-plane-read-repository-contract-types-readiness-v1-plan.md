# Control Plane Read Repository Contract Types Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-repository-contract-types-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_types_readiness_defined`

## 目标

在 read store repository implementation 前固定 repository contract types readiness：未来 `ReadRepositoryContext`、七条 read route 的 request / result type、failure code type、projection / filter / sort type 和 contract smoke 输入输出边界。该切片只定义 Go contract type 的字段与测试边界，不创建 Go repository contract 文件、不实现 repository interface、repository adapter、store selector、SQL、migration、真实数据库、Radish OIDC、token validation 或 production API consumer。

本任务卡中的 `repository contract types readiness` 只表示 contract type 边界已定义，不表示 repository contract types implemented。

## 范围

- 固定 `ReadRepositoryContext` 的 required fields：`request_id`、`tenant_ref`、`subject_ref`、`scope_grants`、`audit_ref`、`issuer_ref` 和 `session_ref`。
- 固定七条 read route 到未来 request / result type 的映射。
- 固定 shared request fields、result envelope fields、failure code type 和 projection / filter / sort allowlist。
- 固定 contract smoke 的未来输入输出 type 边界。
- 将 checker 接入仓库 fast baseline。

## 不在本次范围

- 不创建 `services/platform/internal/httpapi/control_plane_read_repository_contract.go`。
- 不实现 `ControlPlaneReadRepository` interface、repository adapter、store selector、database query、database connection 或 migration runner。
- 不创建 migration 目录、SQL 文件、schema 文件或 database fixture。
- 不接 Radish OIDC、auth middleware、token validation、production API consumer、API key lifecycle 或 quota enforcement。
- 不实现 workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-types-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-schema-migration-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
