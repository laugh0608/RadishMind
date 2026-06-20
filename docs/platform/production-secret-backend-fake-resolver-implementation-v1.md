# Production Secret Backend Fake Resolver Implementation v1

更新时间：2026-06-19

## 文档目的

本文档固定 production secret backend 的 fake resolver implementation 任务卡创建结果，明确后续 runtime 实现批次的准入、边界和验证要求。

结论：状态为 `fake_resolver_implementation_task_card_defined`。本批只创建 fake resolver implementation 的静态任务卡、fixture、checker 和文档入口同步；不实现 resolver runtime，不实现 fake resolver runtime，不创建 no secret leakage smoke runtime，不解析 secret，不创建 credential handle，不连接数据库，不调用云 secret 服务，不启用 repository mode。

## 输入证据

- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已确认 test fixture strategy 仍 required，fake resolver runtime 不打开。
- `production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1` 已固定 fake resolver input / output allowlist、禁止字段和 no secret leakage smoke strategy。
- `production-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1` 已确认下一步可以创建 fake resolver implementation task card。
- `production-secret-reference-basic` 仍为 reference-only manifest，不保存 secret value，不启用 resolver，不允许 cloud calls。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 仍确认 saved draft database secret resolver implementation 不打开。

## 本批打开范围

本批只允许打开以下静态实现准备：

- 创建 `docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md`。
- 创建 `scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json`。
- 创建 `scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py`。
- 更新 `production-ops-secret-backend-implementation-readiness`，把 `fake_resolver_implementation_task_card_status` 推进为 `created_static_task_card`。
- 更新上一张 entry readiness 的 artifact guard，让本任务卡文件成为合法后续产物。

本批不打开任何 runtime success path。fake resolver runtime、no secret leakage smoke runtime、sanitized diagnostics runtime emission、connection factory handoff 和 repository mode integration 仍必须由后续实现批次重新评审。

## Runtime Implementation 准入要求

后续若真正进入 fake resolver runtime implementation，任务卡必须先要求：

- fake resolver 只能在显式测试门禁下启用，默认 disabled。
- 输入只允许 placeholder secret ref key、environment、provider、provider profile、purpose、request id、audit ref 和 policy version。
- 输出只能返回 opaque test credential handle metadata，不得返回 raw secret、DSN、provider raw URL、database hostname 或 credential payload。
- no secret leakage runtime smoke 必须离线运行，不读取真实环境变量，不调用 provider，不调用云 secret 服务，不连接数据库，不执行 SQL。
- sanitized diagnostics 必须只输出 failure code、状态字段、request id、audit ref 和 policy version。
- side effect counters 必须保持零，除非后续 runtime smoke 批次明确打开并记录调用边界。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_implementation_task_card_missing` | `task_card` | fake resolver implementation 任务卡不存在或缺少 scope |
| `fake_resolver_implementation_disabled_gate_missing` | `implementation_gate` | 任务卡未定义 disabled-by-default runtime gate |
| `fake_resolver_implementation_no_leakage_smoke_missing` | `no_secret_leakage` | 任务卡未定义 no secret leakage runtime smoke plan |
| `fake_resolver_implementation_secret_value_detected` | `artifact_guard` | 文档、fixture、checker 或任务卡出现 secret-looking value |
| `fake_resolver_implementation_runtime_created_in_task_card` | `artifact_guard` | 本批创建 fake resolver runtime、resolver runtime 或 smoke runner |
| `fake_resolver_implementation_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或任务卡要求联网、provider call 或云 secret call |
| `fake_resolver_implementation_repository_mode_forbidden` | `repository_mode` | fake resolver 被用于打开 workflow saved draft repository mode 成功路径 |
| `fake_resolver_implementation_scope_overreach` | `implementation_boundary` | 任务卡把 DB provider、SQL、schema marker、migration runner、audit store 或 production API 合入本批 |

所有失败都必须 fail closed，不返回 secret value、DSN、provider raw URL、database hostname、database error detail、opaque credential payload、完整 secret ref value、完整用户 claim 或草案主体。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_implementation_task_card_status`
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

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential payload、resolver backend URL、full secret ref value 或完整用户 claim。本批不允许 runtime emission；所有 diagnostics 只存在于 committed fixture / checker 的静态评审语义中。

## No Fallback

- 不允许 fake resolver fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- 不允许 fake resolver implementation task card 被解释为 credential resolved、production secret backend ready、cloud secret backend ready、repository mode ready 或 production ready。

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

- `docs/platform/production-secret-backend-fake-resolver-implementation-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-implementation-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-implementation-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py`

不得新增或启用以下 artifact：

- resolver runtime
- fake resolver runtime
- no secret leakage smoke runtime
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

后续已单独完成 fake resolver runtime implementation entry review、test-only fake resolver runtime、真实 resolver runtime preconditions、真实 resolver runtime implementation entry review、resolver backend profile selection readiness、真实 resolver no leakage smoke runtime strategy、credential handle runtime boundary readiness、operator approval runtime evidence readiness、audit store handoff readiness、backend health boundary readiness 和 backend health runtime implementation entry review。下一步若继续 production secret backend，应选择 audit store runtime、operator approval runtime、credential handle runtime 或 real resolver runtime implementation entry refresh 单一方向。

`test-fixture-strategy` 后续只由 test-only fake resolver runtime 覆盖离线替身边界；它不代表 production resolver runtime、no secret leakage runtime smoke、audit store 或 production secret backend ready。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-implementation-task-card-entry-readiness-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
