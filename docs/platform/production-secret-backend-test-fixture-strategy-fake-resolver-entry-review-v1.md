# Production Secret Backend Test Fixture Strategy / Fake Resolver Implementation Entry Review v1

更新时间：2026-06-19

## 文档目的

本文档用于评审 production secret backend 的 `test-fixture-strategy` 与 fake resolver implementation 是否具备进入实现的条件。

结论：状态为 `test_fixture_strategy_fake_resolver_entry_review_defined`，entry decision 为 `fake_resolver_implementation_entry_not_opened`。本批只固定 test fixture strategy / fake resolver implementation entry review；后续 `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已把 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined` 固定为静态证据，但 `test-fixture-strategy` 仍为 `required_before_implementation`，不实现 resolver runtime，不实现 fake resolver runtime，不解析 secret，不连接数据库，不调用云 secret 服务，不启用 repository mode。

## 输入证据

- `production-secret-backend-implementation-readiness` 仍显示 `production_secret_backend=not_satisfied`，resolver implementation 为 `not_started`，默认 runtime state 为 `disabled_until_explicit_secret_backend_task`。
- `production-secret-reference-basic` 仍为 reference-only manifest；只保存 `ref:` 形式的 secret reference 和脱敏字段，不保存 secret value，不启用 resolver，不允许 cloud calls。
- `production-secret-backend-config-secret-ref-readiness-v1` 已固定 config 注入点和 reference-only manifest 消费边界，但不调用 resolver。
- `production-secret-backend-provider-profile-secret-binding-readiness-v1` 已固定 provider/profile 到 `secret_ref` 的绑定口径，但不把 `secret_ref_status=present` 写成 credential resolved。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 已固定 future resolver interface 的 disabled result、failure mapping 和 sanitized diagnostics，但没有 resolver runtime。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 已固定 operator runbook 和 negative gates，但不创建 operator executor 或 enablement runtime。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 已固定 rotation / audit policy 前置，但不创建 rotation runtime、audit store 或 audit writer。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 saved draft database secret resolver implementation entry 不打开，fake resolver、sanitized diagnostics runtime、connection factory handoff 和 repository mode integration 均 blocked。

## Entry Review Decision

| candidate | 本次结论 | 阻塞原因 |
| --- | --- | --- |
| placeholder secret ref fixture strategy | `blocked` | fake resolver contract 和 no secret leakage smoke strategy 已定义为静态证据，但 runtime fixture、smoke runner 和 implementation fixture shape 仍不存在 |
| fake resolver interface contract | `blocked` | 输入 / 输出 allowlist 已进入静态策略，但 resolver runtime interface 未实现，opaque credential handle contract 不存在 |
| fake resolver implementation | `blocked` | fake resolver 必须由独立 implementation task card 显式创建，disabled resolver interface 不是 runtime |
| sanitized diagnostics fixture | `blocked` | diagnostics 口径已定义，但没有 fake resolver runtime emission gate 和 no secret leakage smoke runtime |
| connection factory handoff fixture | `blocked` | database connection provider、connection factory 和 credential handle creation 均未打开 |
| repository mode fixture | `blocked` | workflow saved draft repository mode 仍 fail closed，数据库 connection provider 与 production auth runtime 集成未形成 repository mode 成功路径 |

本批不创建 `production-secret-backend-fake-resolver-implementation-v1` 任务卡。后续若要打开 fake resolver implementation，必须继续独立满足 runtime fixture、no secret leakage smoke runner、sanitized diagnostics runtime、environment binding、operator enablement gate、artifact guard 和 offline fast baseline。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `test_fixture_strategy_missing` | `test_fixture_strategy` | resolver implementation 前缺少 fake resolver test fixture strategy |
| `fake_resolver_contract_missing` | `test_fixture_strategy` | 缺少 fake resolver 输入、输出、禁止字段和 disabled default contract |
| `fake_resolver_implementation_forbidden` | `test_fixture_strategy` | 未经独立 task card 直接创建 fake resolver runtime |
| `secret_fixture_value_detected` | `artifact_guard` | committed fixture、文档或 smoke 记录包含 secret-looking value |
| `fake_resolver_fallback_forbidden` | `no_fallback` | fake resolver 被用作 production resolver、provider credential 或 DB credential fallback |
| `test_fixture_cloud_call_forbidden` | `no_side_effects` | fast baseline、fixture checker 或 fake resolver smoke 尝试联网或调用云 secret 服务 |
| `test_fixture_repository_mode_forbidden` | `repository_mode` | test fixture strategy 绕过 repository mode enablement，让 repository 成功路径打开 |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、opaque credential handle、完整用户 claim 或草案主体。

## Sanitized Diagnostics

允许输出：

- `test_fixture_strategy_status`
- `fake_resolver_entry_status`
- `resolver_state`
- `secret_ref_status`
- `secret_ref_version_status`
- `operator_gate_status`
- `rotation_gate_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential handle、resolver backend URL 或完整用户 claim。本批不允许 runtime emission；所有 diagnostics 只存在于 committed fixture / checker 的静态评审语义中。

## No Fallback

- 不从 `RADISHMIND_PLATFORM_API_KEY`、`.env.example`、fixture credential、mock provider 或 local-smoke profile 回退 production credential。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不允许 fake resolver 被解释为 production resolver、credential resolved、cloud secret backend ready 或 operator enablement ready。
- 不允许 fake query executor、reference-only manifest、disabled resolver interface、operator runbook、rotation policy 或 audit policy 代替 production secret backend。
- 不允许 resolver failure 回退到 committed secret value、developer env plaintext、sample、test auth、fake resolver 或 fake query executor。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `secret_value_read_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py`

不得新增或启用以下 artifact：

- resolver runtime
- fake resolver runtime
- fake resolver smoke fixture
- no secret leakage smoke fixture
- cloud secret SDK
- production credential file
- secret value fixture
- opaque credential handle runtime
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- public production API

## 后续推进

当前 readiness 只把 `test-fixture-strategy` 的 blocked 状态升级为可复验的 entry review 证据，不解除 fake resolver implementation、resolver runtime、cloud secret backend、database connection provider、repository mode 或 production ready 阻塞。

后续已单独完成 fake resolver implementation task card、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review、resolver backend profile selection readiness、no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness、backend health boundary readiness 和 backend health runtime implementation entry review。当前这些仍只是静态前置或 test-only runtime 证据，不得把 blocker 与 production resolver runtime 实现混成一个批次。任何 DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production secret audit store、audit writer、backend health runtime 或 public production API 都必须作为独立目标推进。

## 验证

本专题由以下证据固定：

- `docs/task-cards/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py`
- `scripts/checks/fixtures/production-ops-secret-backend-implementation-readiness.json`
- `scripts/checks/fixtures/production-secret-reference-basic.json`
- `scripts/checks/fixtures/production-secret-backend-rotation-audit-policy-readiness-v1.json`

建议验证顺序：

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
