# Control Plane Read Repository Contract Preconditions v1 计划

## 任务标识

- 任务：`control-plane-read-repository-contract-preconditions-v1`
- 主线：`Control Plane / User Workspace / Workflow v1`
- 状态：`repository_contract_preconditions_defined`

## 目标

在 `control-plane-read-auth-store-transition-preconditions-v1` 之后，把未来 `future control plane read store repository` 的最小契约推进到可检查状态。

本切片只固定 repository contract preconditions：interface、tenant predicate、sanitized projection、cursor/filter/sort allowlist、contract smoke 和 failure mapping。它不写 SQL、不建 migration、不实现 repository，也不接真实数据库或 Radish OIDC。

## 范围

- 固定未来 repository interface 名称：`ControlPlaneReadRepository`。
- 固定 repository context：`request_id`、`tenant_ref`、`subject_ref`、`scope_grants`、`audit_ref`、`issuer_ref` 和 `session_ref`。
- 固定七条 read route 到 repository operation 的映射。
- 固定 tenant predicate 必须来自 trusted auth context，不接受 query-string tenant override。
- 固定输出只能是 sanitized summary projection，不暴露 secret、token、API key value / hash、raw request、raw tool payload 或 prompt dump。
- 固定 cursor、filter、sort 必须按 route allowlist fail closed。
- 固定 repository contract smoke 的未来要求：fake-store contract smoke、read store repository contract smoke、database read disabled guard smoke 和 no-side-effect counters。

## 不在本次范围

- 不实现 repository interface。
- 不创建 durable adapter。
- 不写 database schema、migration、SQL 或 query。
- 不接 Radish OIDC、auth middleware 或 token validation。
- 不实现 production API consumer。
- 不实现 API key lifecycle、quota enforcement、rate limit、billing 或 cost ledger。
- 不实现 workflow executor、confirmation、writeback 或 replay。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-repository-contract-preconditions-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-auth-store-transition-preconditions-v1.py`
- `./scripts/check-repo.sh --fast`
