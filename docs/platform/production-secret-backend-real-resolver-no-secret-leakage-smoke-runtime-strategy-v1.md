# Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Strategy v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime implementation task card 之前的 no-secret-leakage smoke runtime strategy。

对应切片：`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1`。

结论：状态为 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`。本批只定义 future smoke runtime gate 的离线策略、扫描面、输入 / 输出 allowlist、负向 probe、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 no secret leakage smoke runtime，不执行 resolver smoke，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不创建 credential payload 或 credential handle runtime，不连接数据库，不启用 repository mode，也不创建 production resolver runtime implementation task card。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定真实 resolver runtime implementation task card 仍 blocked；本批将当时缺失的 `no-secret-leakage-smoke-runtime-strategy` 拆成独立静态前置证据。
- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已要求真实 resolver implementation task card 必须有 no leakage validation plan。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile static selection，但不创建 backend runtime 或 health check。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 已实现 test-only fake resolver runtime 的 offline no leakage Go test；它不能替代 production resolver 的 no leakage strategy。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 仍确认 production resolver interface 默认 disabled。

## Smoke Runtime Strategy

future no-secret-leakage smoke runtime gate 必须先作为离线、可复验、无副作用的测试策略存在。它只能消费 synthetic placeholder input / output / audit record，不允许调用真实 resolver、fake resolver、cloud secret backend、provider、DB driver 或 production API。

必须扫描的面：

| surface | 要求 |
| --- | --- |
| resolver input | 只允许 reference-only secret ref、provider/profile、environment、operator approval ref、audit ref、policy version 和 request id |
| resolver success output | 只允许 sanitized resolver state、credential handle metadata presence、opaque handle id、policy/audit/request metadata |
| resolver failure output | 只允许 failure code、failure boundary、sanitized diagnostic、policy/audit/request metadata |
| diagnostics | 只允许脱敏状态，不得输出 secret value、provider raw URL、DSN、cloud credential 或 backend URL |
| audit metadata | 只允许 audit ref、policy version、secret ref version ref、operator approval ref、failure code 和 sanitized status |
| smoke summary | 只允许 case id、probe category、pass/fail、failure code、sanitized diagnostic 和 zero side-effect counters |

未来 smoke runner 必须覆盖：

- success metadata no leakage
- failure diagnostics no leakage
- audit metadata no leakage
- secret-ref-value probe rejection
- credential payload probe rejection
- provider raw URL / DSN / database hostname probe rejection
- authorization header / cookie / user claim probe rejection
- cloud credential marker probe rejection
- cross-environment fallback rejection
- fake resolver substitution rejection
- repository mode / DB / production API side-effect rejection

本批不创建上述 runner、fixture output 或 executable smoke artifact，只固定策略与准入检查。

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| smoke strategy contract | `defined_static_only` | 只固定离线策略，不创建 runtime |
| smoke runtime gate | `not_created` | 本批不创建 executable smoke runtime |
| input allowlist | `reference_only_fields_fixed` | 只允许 reference-only 和审计字段 |
| success output allowlist | `sanitized_metadata_only` | 只允许 opaque handle metadata presence / id，不输出 payload |
| failure output allowlist | `fail_closed_sanitized_only` | 失败只返回 code 和 sanitized diagnostic |
| probe categories | `required_before_runtime_task` | future runtime task 必须覆盖负向 probe |
| fixture source | `synthetic_placeholder_only` | 不使用真实 secret、环境变量或 fixture credential |
| artifact scan | `required_before_runtime_task` | future smoke 必须扫描 docs / fixture / output / audit summary |
| fake resolver substitution | `forbidden` | test-only fake resolver runtime 不能替代 production no leakage gate |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| DB / repository / API | `blocked` | 不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `real_resolver_no_leakage_strategy_missing` | `no_secret_leakage` | 缺少 no leakage smoke runtime strategy evidence |
| `real_resolver_no_leakage_input_allowlist_missing` | `input_contract` | 未定义 resolver input allowlist |
| `real_resolver_no_leakage_output_allowlist_missing` | `output_contract` | 未定义 success / failure output allowlist |
| `real_resolver_no_leakage_probe_matrix_missing` | `probe_matrix` | 未定义负向 probe categories |
| `real_resolver_no_leakage_scan_surface_missing` | `scan_surface` | 未定义 input / output / diagnostics / audit / summary 扫描面 |
| `real_resolver_no_leakage_secret_value_detected` | `artifact_guard` | 文档、fixture、summary 或 future smoke output 出现 secret-looking value |
| `real_resolver_no_leakage_credential_payload_detected` | `artifact_guard` | 出现 credential payload、raw handle payload 或 secret material |
| `real_resolver_no_leakage_diagnostic_exposure_detected` | `diagnostics` | diagnostics 暴露 provider raw URL、DSN、backend URL、cloud credential 或 user claim |
| `real_resolver_no_leakage_audit_exposure_detected` | `audit_metadata` | audit metadata 暴露 secret value、full secret ref、authorization header 或 cookie |
| `real_resolver_no_leakage_runtime_created_forbidden` | `artifact_guard` | 本批创建 no leakage smoke runtime 或 production resolver runtime |
| `real_resolver_no_leakage_cloud_call_forbidden` | `no_side_effects` | checker、fixture 或 future strategy 要求联网、provider call 或云 secret call |
| `real_resolver_no_leakage_fake_resolver_substitution_forbidden` | `implementation_boundary` | 用 fake resolver runtime 代替 production no leakage gate |
| `real_resolver_no_leakage_repository_mode_forbidden` | `repository_mode` | 本批启用 repository mode 或 DB provider |
| `real_resolver_no_leakage_scope_overreach` | `implementation_boundary` | 本批合入 credential handle runtime、audit store、backend health runtime 或 production API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、完整 secret ref value、完整 credential handle、完整用户 claim、authorization header 或 cookie。

## Sanitized Diagnostics

允许输出：

- `real_resolver_no_leakage_strategy_status`
- `smoke_runtime_status`
- `scan_surface_status`
- `input_allowlist_status`
- `output_allowlist_status`
- `probe_matrix_status`
- `artifact_scan_status`
- `side_effect_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## No Fallback

- no leakage strategy 缺失时真实 resolver runtime implementation task card 必须继续 blocked。
- 不允许 production resolver no leakage gate fallback 到 fake resolver runtime、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许把 test-only fake resolver runtime 的 Go no leakage test 当成 production resolver no leakage gate。
- 不允许 provider/profile binding 缺失时通过 `RADISHMIND_PLATFORM_API_KEY` 或其它环境变量 credential 生成 smoke success。
- 不把 no leakage smoke strategy 写成 no leakage smoke runtime created、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `smoke_runtime_execution_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_audit_store_write_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.md`
- `docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.json`
- `scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py`

不得新增或启用以下 artifact：

- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- backend profile runtime
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- no secret leakage smoke runtime
- no secret leakage smoke output fixture
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- production secret audit store
- audit writer
- public production API

## 后续推进

本批完成后，真实 resolver runtime implementation task card 仍不能创建。`no-secret-leakage-smoke-runtime-strategy` 已固定为静态前置证据；后续已由 `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 固定 credential handle boundary，但 executable smoke runtime、production resolver runtime、credential handle runtime、operator approval runtime evidence、audit store handoff 和 backend health boundary 均未创建。

下一步应继续从剩余 blocker 中选择一个单独推进：

1. `operator-approval-runtime-evidence-readiness`
2. `production-secret-audit-store-handoff-readiness`
3. `resolver-backend-health-boundary-readiness`

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
