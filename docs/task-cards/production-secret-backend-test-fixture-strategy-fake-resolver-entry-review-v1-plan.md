# `Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review` v1 计划

更新时间：2026-06-19

## 任务目标

本任务卡用于评审 `production-secret-backend-implementation-readiness` 中仍 blocked 的 `test-fixture-strategy`，并判断 fake resolver implementation 是否可以进入实现批次。

结论：状态为 `test_fixture_strategy_fake_resolver_entry_review_defined`。本批只创建 entry review evidence；后续 `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined` 静态证据，但 `test-fixture-strategy` 仍为 `required_before_implementation`，fake resolver implementation entry 不打开。不实现 resolver runtime，不实现 fake resolver runtime，不解析 secret，不连接数据库，不调用云 secret 服务，不启用 repository mode。

## 输入事实源

- [Production Secret Backend Implementation v1 计划](production-secret-backend-implementation-v1-plan.md)
- [Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1](../platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md)
- [Production Secret Backend Rotation / Audit Policy Readiness v1](../platform/production-secret-backend-rotation-audit-policy-readiness-v1.md)
- [Production Secret Backend Secret Resolver Interface Disabled Readiness v1](../platform/production-secret-backend-secret-resolver-interface-disabled-readiness-v1.md)
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.json`

## 当前事实

- `rotation-and-audit-policy` 已满足，但它只固定 rotation trigger、audit event fields、secret ref version reference、rollback policy 和脱敏审计字段前置。
- `test-fixture-strategy` 仍不是 implementation-ready 证据；当前只有 fake resolver contract / no secret leakage smoke 静态策略，没有 fake resolver implementation task card、no secret leakage smoke runtime、sanitized diagnostics runtime 或 offline fake resolver smoke。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `resolver_runtime_status` 仍为 `not_created`。
- `fake_resolver_status` 仍为 `not_created`。
- `production-secret-reference-basic` 仍为 reference-only manifest，resolver disabled，cloud calls disabled。

## Entry Review Matrix

| candidate | 本批结论 | 需要后续补齐 |
| --- | --- | --- |
| placeholder secret ref fixture strategy | `blocked` | fake resolver contract、fixture shape、no secret leakage smoke |
| fake resolver interface contract | `blocked` | resolver input / result、opaque handle policy、forbidden fields checker |
| fake resolver implementation | `blocked` | 独立 implementation task card、offline smoke、artifact guard |
| sanitized diagnostics fixture | `blocked` | runtime emission gate、diagnostics no leakage check |
| connection factory handoff fixture | `blocked` | DB provider、connection factory、credential handle contract |
| repository mode fixture | `blocked` | repository mode enablement runtime、DB connection provider、production auth / membership path |

本批不创建 fake resolver implementation task card，也不创建 resolver runtime、fake resolver runtime、cloud SDK、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 production API。

## Failure Mapping

| code | 处理方式 |
| --- | --- |
| `test_fixture_strategy_missing` | fail closed；resolver implementation 不能进入 |
| `fake_resolver_contract_missing` | fail closed；不能创建 fake resolver runtime |
| `fake_resolver_implementation_forbidden` | fail closed；需要独立 implementation task card |
| `secret_fixture_value_detected` | fail closed；移除 secret-looking fixture / 文档内容 |
| `fake_resolver_fallback_forbidden` | fail closed；fake resolver 不得代替 production resolver |
| `test_fixture_cloud_call_forbidden` | fail closed；fast baseline 不联网、不调用云 secret 服务 |
| `test_fixture_repository_mode_forbidden` | fail closed；不得用 fixture 打开 repository mode 成功路径 |

failure diagnostics 只能输出脱敏状态、failure code、request id、audit ref 和 policy version，不输出 raw secret、token、API key、DSN、provider raw URL、database hostname、database error detail 或 opaque credential handle。

## Artifact Guard

允许新增：

- platform topic
- 本任务卡
- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json`
- `check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py`

禁止新增：

- fake resolver implementation task card
- resolver runtime
- fake resolver runtime
- no secret leakage smoke runtime
- cloud secret SDK
- secret value fixture
- credential handle runtime
- DB provider / driver
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- repository mode runtime
- production secret audit store
- public production API

## 验收方式

- 新增 fixture / checker 固定 entry decision、blocked candidates、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard。
- 更新 `production-ops-secret-backend-implementation-readiness.json`，让 `test-fixture-strategy` 保持 `required_before_implementation`，但引用本批 blocked entry review evidence。
- checker 接入 `./scripts/check-repo.sh --fast`，顺序位于 rotation / audit policy readiness checker 之后。
- 本批至少运行：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```

## 停止线

- 不实现 fake resolver runtime。
- 不实现 resolver runtime。
- 不解析 secret，不读取 secret value，不调用云 secret 服务。
- 不创建 DB provider、DB driver、connection factory、SQL、schema marker、migration runner 或 repository mode runtime。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `test_fixture_strategy_fake_resolver_entry_review_defined` 解释为 `test-fixture-strategy` satisfied、fake resolver ready、resolver ready、production secret backend ready、repository mode ready、database ready 或 production ready。
