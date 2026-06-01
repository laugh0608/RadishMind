# Control Plane Read Auth/Store Transition Preconditions v1 计划

## 任务标识

- 任务：`control-plane-read-auth-store-transition-preconditions-v1`
- 主线：`Control Plane / User Workspace / Workflow v1`
- 状态：`auth_store_transition_preconditions_defined`

## 目标

在 `control-plane-read-dev-live-consumer-v1` 之后，固定从当前 dev fake auth / fixture-backed fake store 走向未来 `future Radish OIDC / auth middleware` 与 `future control plane read store repository` 的迁移准入条件。

本切片只定义 transition gates、route matrix、dual smoke plan、failure code 和禁止项，避免 dev-only live path 被误读成生产 API consumer、真实 OIDC 或数据库 read path。

## 范围

- 固定当前来源：`offline_fixture` 默认消费、显式 `dev_live_http`、`RADISHMIND_CONTROL_PLANE_READ_DEV_AUTH=1`、test-only fake auth context 和 fixture-backed fake store。
- 固定未来来源：`future Radish OIDC / auth middleware` 和 `future control plane read store repository`。
- 固定 auth middleware transition gates：issuer discovery、token validation contract、claim mapping、tenant binding、scope projection、audit context、fail-closed missing identity、dev fake auth disabled by default。
- 固定 read store transition gates：repository interface before SQL、tenant predicate from auth context、sanitized summary projection、cursor/filter/sort allowlist、failure taxonomy、no database write、no secret material。
- 固定七条 read route 的 route shape 必须保持稳定。
- 固定 dual smoke 要求：当前 offline fixture / dev live fake-store smoke 保留；未来替换前必须补 auth middleware context smoke、repository contract fake-store smoke、read store repository contract smoke 和 database read disabled guard smoke。

## 不在本次范围

- 不实现 Radish OIDC middleware。
- 不实现 token validation、login redirect 或真实 session。
- 不创建数据库 schema、migration、query 或 repository implementation。
- 不替换当前 fake store。
- 不实现 production API consumer。
- 不实现 API key lifecycle、quota enforcement、rate limit、billing 或 cost ledger。
- 不实现 workflow builder、workflow executor、confirmation、writeback 或 replay。
- 不把 `apps/radishmind-web/` 或 `apps/radishmind-console/` 写成 production admin console。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-control-plane-read-auth-store-transition-preconditions-v1.py`
- `./scripts/run-python.sh scripts/run-control-plane-read-consumer-smoke.py --check`
- `./scripts/check-repo.sh --fast`
