# Workflow Saved Draft Production Auth Readiness v1 任务卡

状态：`draft_production_auth_readiness_defined`

## 任务目标

固定 saved workflow draft durable store 后续进入 adapter smoke / repository adapter 前必须具备的 production auth readiness 证据链。

本任务只定义 Radish OIDC issuer discovery evidence、token validation contract preconditions、claim mapping、tenant / workspace / application binding、scope projection、failure mapping、no fake fallback、no side effects 和 downstream readiness review。不创建 OIDC middleware、token validation、membership adapter、repository adapter、production API 或数据库实现。

## 输入

- `radish-oidc-client-preconditions`
- `Saved Workflow Draft Auth Context Preconditions v1`
- `Saved Workflow Draft Store Selector Implementation v1`
- `Saved Workflow Draft Schema Artifact Materialization v1`
- `Saved Workflow Draft Repository Adapter Implementation Plan v1`
- `Saved Workflow Draft Adapter Smoke Readiness v1`

## 本批输出

- `docs/features/workflow/saved-workflow-draft-production-auth-readiness-v1.md`
- `scripts/checks/fixtures/workflow-saved-draft-production-auth-readiness-v1.json`
- `scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py`
- `scripts/check-repo.py` fast baseline 注册
- workflow 专题入口、Saved Workflow Draft v1、current focus、scripts README、任务卡入口和周志同步

## 验收要求

- fixture 必须覆盖 issuer discovery evidence contract、token validation contract preconditions、claim mapping、tenant / workspace / application binding policy、operation scope projection matrix、failure mapping、no fake fallback 和 no side effects。
- checker 必须确认上游 OIDC preconditions、auth context preconditions、selector implementation、schema artifact materialization、repository adapter plan 和 adapter smoke readiness 的状态未漂移。
- checker 必须确认后续 adapter smoke / repository adapter 可以消费本 readiness evidence，但当前 adapter smoke execution、repository adapter implementation、auth middleware、token validation、membership adapter 和 production API 仍 blocked。
- checker 必须确认 production auth implementation artifact、repository adapter artifact、SQL / migration artifact 和 production API artifact 没有提前出现。
- 新 checker 必须进入 `./scripts/check-repo.sh --fast`。

## 验证

- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-production-auth-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-auth-context-preconditions-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-adapter-smoke-readiness-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-adapter-implementation-plan-v1.py`
- `./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-artifact-materialization-v1.py`
- `./scripts/check-repo.sh --fast`
- `./scripts/check-repo.sh`

## 停止线

- 不创建 OIDC middleware、OIDC client、token validation、token validation schema、membership adapter、session cookie、production auth smoke fixture / checker 或 production API consumer。
- 不创建 repository interface、repository adapter、adapter smoke fixture、adapter smoke checker、SQL migration、schema version table、migration runner、数据库连接或 database query。
- 不实现 durable persistence、publish、run、executor、confirmation decision、writeback、replay、resume 或 materialized result reader。
