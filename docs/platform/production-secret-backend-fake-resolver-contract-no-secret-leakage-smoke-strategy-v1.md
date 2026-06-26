# Production Secret Backend Fake Resolver Contract / No Secret Leakage Smoke Strategy v1

更新时间：2026-06-19

## 文档目的

本文档固定 production secret backend 在进入 fake resolver implementation 前必须具备的静态契约和 no secret leakage smoke 策略。

结论：状态为 `fake_resolver_contract_no_secret_leakage_smoke_strategy_defined`。本批只定义 fake resolver contract、placeholder secret ref fixture shape、no secret leakage smoke strategy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不实现 resolver runtime，不实现 fake resolver runtime，不解析 secret，不创建 credential handle，不连接数据库，不调用云 secret 服务，不启用 repository mode。

## 输入证据

- `production-secret-backend-test-fixture-strategy-fake-resolver-entry-review-v1` 已确认 fake resolver implementation entry 不打开，缺口集中在 fake resolver contract、placeholder secret ref fixture、no secret leakage smoke、sanitized diagnostics runtime、environment binding 和 artifact guard。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 已固定 disabled resolver interface 的失败语义，但它不是 runtime，也不能被 fake resolver 继承为成功路径。
- `production-secret-reference-basic` 仍为 reference-only manifest，只允许 `ref:` 形式的 secret reference 和脱敏字段。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 仍确认 saved draft database secret resolver implementation 不打开。

## Fake Resolver Contract

本批只定义未来 fake resolver implementation 必须遵守的 contract shape：

| contract area | 静态要求 |
| --- | --- |
| input | 只允许 `environment`、`provider`、`provider_profile`、`secret_ref_key`、`secret_ref_version`、`purpose`、`request_id`、`audit_ref`、`policy_version` |
| input guard | `environment` 必须与 fixture 绑定；`purpose` 只能是测试 smoke；`secret_ref_key` 只能是 placeholder reference key，不是 secret value |
| success output | 只能返回 opaque test credential handle metadata；不得返回 raw credential、DSN、provider raw URL、database hostname 或完整 secret ref value |
| failure output | 必须 fail closed，返回 `failure_code`、`sanitized_diagnostic`、`request_id`、`audit_ref`、`policy_version` |
| status output | 可返回 `fake_resolver_status=contract_defined_runtime_missing`，但不得返回 `credential_resolved=true` |
| audit output | 只允许脱敏 audit ref、policy version、environment、provider/profile 和 failure code |

本契约不创建 Go interface、runtime type、resolver function、fake resolver implementation、connection factory handoff 或 smoke executor。

## No Secret Leakage Smoke Strategy

未来 no secret leakage smoke 必须先证明以下边界，再允许实现 fake resolver runtime：

- committed fixture、文档、smoke record 和 checker 不包含 secret-looking value。
- fake resolver 输入只包含 placeholder secret ref key，不包含真实 secret value。
- fake resolver 成功输出不得包含 password、token、API key、DSN、provider raw URL、database hostname、opaque credential payload 或 cloud credential。
- failure diagnostics 只能使用 failure code 和 sanitized diagnostic。
- smoke 必须离线运行，不读取真实环境变量，不调用 provider，不调用云 secret 服务，不连接数据库，不打开 driver，不执行 SQL。
- smoke 不得启用 workflow saved draft repository mode，也不得把 fake resolver 成功解释为 production secret backend ready。

本批只定义 smoke strategy，不创建 `production-secret-backend-no-secret-leakage-smoke-v1.json`，不创建 smoke runner，不执行 runtime smoke。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `fake_resolver_contract_runtime_missing` | `fake_resolver_contract` | contract 已定义但 runtime 未实现 |
| `fake_resolver_input_shape_invalid` | `fake_resolver_contract` | 输入字段超出 allowlist 或缺少 environment / provider / secret ref key |
| `fake_resolver_secret_value_detected` | `artifact_guard` | fixture、文档或 smoke record 出现 secret-looking value |
| `fake_resolver_output_secret_leakage` | `no_secret_leakage` | output 包含 raw secret、token、API key、DSN、provider raw URL 或 database hostname |
| `fake_resolver_runtime_forbidden` | `implementation_gate` | 未经独立 implementation task card 创建 fake resolver runtime |
| `fake_resolver_cloud_call_forbidden` | `no_side_effects` | checker、smoke 或 fixture 触发联网、provider call 或云 secret call |
| `fake_resolver_repository_mode_forbidden` | `repository_mode` | fake resolver 被用于打开 repository mode 成功路径 |

所有失败都必须 fail closed，不返回草案主体、secret value、credential payload、provider raw response、database error detail 或完整用户 claim。

## Sanitized Diagnostics

允许输出：

- `fake_resolver_contract_status`
- `fake_resolver_runtime_status`
- `no_secret_leakage_smoke_strategy_status`
- `secret_ref_key_status`
- `secret_ref_version_status`
- `environment_binding_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、opaque credential payload、resolver backend URL、full secret ref value 或完整用户 claim。

## No Fallback

- 不允许 fake resolver fallback 到 production resolver。
- 不允许 production resolver fallback 到 fake resolver。
- 不允许 test secret ref fallback 到 production secret ref。
- 不允许 secret resolver failure fallback 到 committed value、developer env plaintext、fixture credential、mock provider、local-smoke profile、fake query executor 或 sample。
- 不允许 fake resolver contract 被解释为 credential resolved、production secret backend ready、cloud secret backend ready 或 repository mode ready。

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

- `docs/platform/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.md`
- `docs/task-cards/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.json`
- `scripts/check-production-ops-secret-backend-fake-resolver-contract-no-secret-leakage-smoke-strategy-v1.py`

不得新增或启用以下 artifact：

- resolver runtime
- fake resolver runtime
- fake resolver implementation task card
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

当前只把 fake resolver contract 与 no secret leakage smoke strategy 固定为可复验静态证据。后续若继续推进 fake resolver implementation，必须单独创建 implementation task card，且至少先通过本策略 checker、implementation readiness checker、production secret backend contract checker、production secret reference contract checker、workflow saved draft database secret resolver readiness checker 和 entry review checker。

即便后续 fake resolver implementation 打开，也仍不得同时打开真实 resolver runtime、DB provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode、production secret audit store、cloud secret service 或 public production API。

## 验证

建议验证顺序：

```bash
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
