# Control Plane Read Production Auth Readiness v1 任务卡

## 任务标识

- 切片：`control-plane-read-production-auth-readiness-v1`
- 轨道：`Control Plane / User Workspace / Workflow v1`
- 状态：`production_auth_readiness_defined`

## 目标

在实现任何 Radish OIDC client、token validation、auth middleware、production API consumer 或正式 read-side durable adapter smoke 之前，先固定 production auth readiness 的可检查治理边界。该切片只定义未来 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant binding、scope projection、failure mapping、no fake fallback、no side effects 和停止线。

本任务卡不表示 production auth ready、Radish OIDC client ready、token validation ready、auth middleware ready、production API consumer ready、repository adapter ready、database ready 或 production ready。

## 范围

- 新增 production auth readiness fixture 和 checker。
- 消费 `radish-oidc-client-preconditions`、`control-plane-read-auth-store-transition-preconditions-v1`、`control-plane-read-route-contract-v1`、`control-plane-read-negative-contract-v1` 与 `control-plane-read-store-selector-smoke-readiness-v1`。
- 固定未来 issuer discovery evidence 必须包含 issuer、discovery document、JWKS、signing algorithm、scope、有效期、环境、operator review 和 sanitized evidence ref。
- 固定 token validation contract preconditions：issuer/audience、JWKS refresh、algorithm allowlist、时间窗口、required claims 和 sanitized failure envelope。
- 固定七条 read route 的 scope projection matrix，确保 route scope 来自 trusted auth context，不允许 production fake auth header。
- 固定 tenant binding 必须 fail-closed，禁止 query/path tenant override、default tenant fallback 和 local admin bypass。
- 将 checker 接入仓库 fast baseline，并同步 read-side contract、current focus、roadmap、capability matrix、integration contracts、脚本说明、platform README 和周志。

## 不在本次范围

- 不创建 `contracts/radish-oidc-token-validation.schema.json`。
- 不创建 `control_plane_read_auth_middleware.go`、middleware test、production auth smoke fixture 或 production auth smoke checker。
- 不实现 Radish OIDC client、issuer network call、token validation、auth middleware、login / logout flow、session cookie 或 production API consumer。
- 不实现 repository interface、repository adapter、store selector、database query、SQL、migration 或 migration runner。
- 不实现 API key lifecycle、quota enforcement、rate limit、billing、cost ledger、workflow executor、confirmation、business writeback、replay 或任何写能力。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-store-selector-smoke-readiness-v1.py`
- `./scripts/check-repo.sh --fast`
