# Production Secret Backend Fake Resolver Runtime Implementation v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 的 fake resolver runtime implementation 任务卡创建结果，明确后续 test-only runtime 实现必须满足的边界、输入输出、诊断、泄漏检查和验证方式。

结论：状态为 `fake_resolver_runtime_implementation_task_card_defined`。本批只创建 runtime implementation 的静态任务卡、fixture、checker 和文档入口同步；不实现 fake resolver runtime，不创建 no secret leakage smoke runtime，不实现 production resolver runtime，不解析 secret，不读取真实环境 secret，不创建 credential handle runtime，不连接数据库，不调用云 secret 服务，不启用 workflow saved draft repository mode 或 production API。

## 输入证据

- `production-secret-backend-fake-resolver-runtime-implementation-entry-review-v1` 已固定 `fake_resolver_runtime_implementation_entry_review_defined`，并确认下一步可以单独创建 runtime implementation 任务卡。
- `production-secret-backend-fake-resolver-implementation-v1` 已创建 fake resolver implementation 静态任务卡，要求后续 runtime 默认 disabled、test-only、offline no leakage smoke。
- `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 fake resolver input / output allowlist、禁止字段和 no secret leakage smoke strategy。
- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已固定 `test-fixture-strategy` 仍为 `required_before_implementation`。
- `production-ops-secret-backend-implementation-readiness` 显示 resolver runtime、fake resolver runtime 和 no secret leakage smoke runtime 仍为 `not_created`。

## 本批打开范围

本批只允许打开以下静态实现准备：

- 创建 `docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md`。
- 创建 `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json`。
- 创建 `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py`。
- 更新 `production-ops-secret-backend-implementation-readiness`，把 `fake_resolver_runtime_implementation_task_card_status` 推进为 `created_static_task_card`，并把 `fake_resolver_runtime_implementation_status` 推进为 `task_card_defined_runtime_not_started`。
- 更新入口文档和 `check-repo.py` 注册顺序，让后续实现只能在本任务卡约束下进入。

本批不创建任何 runtime success path。fake resolver runtime、no secret leakage smoke runtime、credential handle runtime、sanitized diagnostics runtime emission、connection factory handoff 和 repository mode integration 仍必须由后续代码实现批次单独落地并验证。

## Runtime Implementation 任务卡要求

后续代码实现批次必须把以下项目作为验收项：

- fake resolver 默认 disabled，只能在显式 test fixture / smoke gate 下启用。
- 输入字段只允许 `environment`、`provider`、`provider_profile`、`secret_ref_key`、`secret_ref_version`、`purpose`、`request_id`、`audit_ref`、`policy_version`。
- 成功输出只能返回 opaque test credential handle metadata，不返回 credential payload、raw secret、DSN、provider raw URL 或 database hostname。
- 失败输出必须包含 `failure_code`、`sanitized_diagnostic`、`request_id`、`audit_ref` 和 `policy_version`，并 fail closed。
- no secret leakage smoke 必须扫描 committed fixture、runtime response、diagnostics 和 smoke record，且离线执行。
- side effect counters 必须保持零，不允许 provider call、cloud secret call、database connection、driver open、SQL execution 或 repository mode enablement。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_runtime_task_card_missing` | `task_card` | runtime implementation 任务卡不存在或缺少 scope |
| `fake_resolver_runtime_task_card_entry_review_missing` | `implementation_entry` | runtime implementation entry review 未被消费 |
| `fake_resolver_runtime_disabled_gate_missing` | `implementation_gate` | 任务卡未定义 disabled-by-default runtime gate |
| `fake_resolver_runtime_placeholder_fixture_missing` | `test_fixture` | 任务卡未定义 placeholder secret ref fixture |
| `fake_resolver_runtime_environment_binding_missing` | `environment_binding` | 任务卡未定义 environment binding |
| `fake_resolver_runtime_opaque_handle_metadata_missing` | `credential_boundary` | 任务卡未定义 opaque test credential handle metadata |
| `fake_resolver_runtime_diagnostics_boundary_missing` | `sanitized_diagnostics` | 任务卡未定义 runtime diagnostics allowlist |
| `fake_resolver_runtime_no_leakage_smoke_missing` | `no_secret_leakage` | 任务卡未定义离线 no secret leakage smoke |
| `fake_resolver_runtime_secret_value_detected` | `artifact_guard` | 文档、fixture、checker 或任务卡出现 secret-looking value |
| `fake_resolver_runtime_created_in_task_card` | `artifact_guard` | 本批创建 fake resolver runtime、resolver runtime 或 smoke runner |
| `fake_resolver_runtime_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或任务卡要求联网、provider call 或云 secret call |
| `fake_resolver_runtime_repository_mode_forbidden` | `repository_mode` | fake resolver 被用于打开 workflow saved draft repository mode 成功路径 |
| `fake_resolver_runtime_scope_overreach` | `implementation_boundary` | 任务卡把 DB provider、SQL、schema marker、migration runner、audit store 或 production API 合入本批 |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、credential payload、完整 secret ref value、完整用户 claim 或草案主体。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_runtime_implementation_task_card_status`
- `fake_resolver_runtime_implementation_status`
- `fake_resolver_contract_status`
- `no_secret_leakage_smoke_strategy_status`
- `test_fixture_strategy_status`
- `resolver_state`
- `fake_resolver_runtime_status`
- `no_secret_leakage_smoke_runtime_status`
- `secret_ref_status`
- `environment_binding_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value 或完整用户 claim。本批不允许 runtime emission；所有 diagnostics 只存在于 committed fixture / checker 的静态任务卡语义中。

## No Fallback

- 不允许 fake resolver fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- 不允许 runtime implementation task card 被解释为 credential resolved、production secret backend ready、cloud secret backend ready、repository mode ready 或 production ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境变量、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

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

- `docs/platform/production-secret-backend-fake-resolver-runtime-implementation-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-runtime-implementation-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-runtime-implementation-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py`

不得新增或启用以下 artifact：

- fake resolver runtime
- resolver runtime
- no secret leakage smoke runtime
- no secret leakage smoke fixture
- cloud secret SDK
- production credential file
- secret value fixture
- credential handle runtime
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

下一步若继续 production secret backend，可以进入 test-only fake resolver runtime 的代码实现批次。该实现必须直接消费本任务卡，并且只落地 disabled-by-default fake resolver runtime、placeholder secret ref fixture、opaque metadata、sanitized diagnostics、离线 no secret leakage smoke 和 side effect counters。

真实 resolver runtime、DB provider、cloud secret service、repository mode、audit store 或 production API 都仍需独立评审。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
