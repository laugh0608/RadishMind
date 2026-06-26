# Production Secret Backend Real Resolver No Secret Leakage Smoke Runtime Implementation Entry Review v1

更新时间：2026-06-21

## 文档目的

本文档评审 production secret backend 是否可以从真实 resolver no-secret-leakage smoke runtime strategy 进入可执行 smoke runtime implementation task card。

对应切片：`production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1`。

结论：状态为 `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined`，entry decision 记录为 `real_resolver_no_secret_leakage_smoke_runtime_implementation_blocked_before_task_card`。`real_resolver_no_secret_leakage_smoke_runtime_strategy_defined` 已经固定 scan surfaces、input / output allowlist、probe categories、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard，但可执行 smoke runtime 会触碰 synthetic fixture、scan runner、artifact scanner、negative probe output、audit summary 和 side effect counter 的组合边界。当前 smoke runtime、runner、execution output 和 artifact scanner 都未创建；credential handle runtime、operator approval runtime、audit store runtime、backend health runtime 和 production resolver runtime 也仍未创建。因此本批不创建 no secret leakage smoke runtime implementation task card，也不实现或执行 smoke runtime。

本批只固定 entry review 的输入证据、阻塞语义、future runtime task card 必须覆盖的范围、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、real resolver runtime entry alignment 和 implementation readiness alignment。不读取真实 secret，不调用云 secret 服务，不访问 provider，不调用 fake resolver，不调用 production resolver，不连接数据库，不创建 credential payload，不创建 credential handle，不执行 approval runtime，不创建 audit store / writer / event，不执行 backend health check，不启用 repository mode，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 `real_resolver_no_secret_leakage_smoke_runtime_strategy_defined`，但 smoke runtime 未创建也未执行。
- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已确认真实 resolver runtime implementation task card 不在当前切片创建。
- `production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1` 已消费完整前置链，并把 `real_no_leakage_smoke_runtime_not_created` 保持为 production resolver runtime 的阻塞项。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile selection，但 backend runtime 未创建。
- `production-secret-backend-credential-handle-runtime-implementation-entry-review-v1` 已固定 credential handle runtime task card 仍 blocked。
- `production-secret-backend-operator-approval-runtime-implementation-entry-review-v1` 已固定 approval runtime task card 仍 blocked。
- `production-secret-backend-audit-store-runtime-implementation-entry-review-v1` 已固定 audit store runtime task card 仍 blocked。
- `production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` 已固定 backend health runtime task card 仍 blocked。
- `production-ops-secret-backend-implementation-readiness` 仍保持 `production_secret_backend=not_satisfied`、`real_resolver_no_secret_leakage_smoke_runtime_status=not_created`、`resolver_runtime_status=not_created` 和 `production_resolver_runtime_task_card_status=not_created`。

## Entry Review Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| no leakage strategy | `satisfied_static_strategy` | scan surfaces、allowlist、probe categories 和 failure mapping 已定义 |
| runtime task card | `blocked_before_task_card` | 本批不创建 smoke runtime implementation task card |
| smoke runtime | `not_created` | 本批不创建 runtime |
| smoke runner / execution | `not_created_not_executed` | 本批不创建 runner，不执行 smoke |
| synthetic fixture source | `required_before_runtime` | future runtime task 必须先定义 synthetic placeholder source，不得使用真实 secret、env secret 或 fixture credential |
| scan surfaces | `satisfied_static_only` | input / output / diagnostics / audit / summary 扫描面已定义，但没有 executable scanner |
| input allowlist | `reference_only_fields_fixed` | 只允许 reference-only secret ref、provider/profile、environment、approval/audit/policy/request metadata |
| output allowlist | `sanitized_metadata_only` | 只允许 sanitized state、opaque handle metadata presence、failure code 和 sanitized diagnostic |
| negative probe matrix | `satisfied_static_only` | probe categories 已定义，但未执行 |
| artifact scan | `required_before_runtime` | future runtime task 必须扫描 docs、fixtures、summary、audit metadata 和 smoke output |
| fake resolver substitution | `forbidden` | test-only fake resolver runtime 不能替代 production no leakage gate |
| credential handle runtime | `blocked_runtime_not_created` | credential handle runtime 未创建 |
| operator approval runtime | `blocked_runtime_not_created` | approval runtime 未创建也未执行 |
| audit store runtime | `blocked_store_not_created` | audit store / writer 未创建，event 未写入 |
| backend health runtime | `blocked_runtime_not_created` | backend health runtime 未创建，health check 未执行 |
| production resolver runtime | `not_created` | 本批不创建 production resolver runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| database / repository mode | `blocked` | DB provider、SQL、schema marker、repository mode 和 production API 均不打开 |

## Blocked Conditions

后续如需重新评审 no leakage smoke runtime implementation task card，至少必须单独解决以下运行时依赖或实现计划缺口；本批只把这些依赖固定为阻塞项：

- `no_secret_leakage_smoke_runtime_task_card_not_created`
- `no_secret_leakage_smoke_runtime_not_created`
- `no_secret_leakage_smoke_runner_not_created`
- `no_secret_leakage_smoke_execution_not_performed`
- `synthetic_placeholder_fixture_not_created`
- `artifact_scanner_not_created`
- `credential_handle_runtime_not_created`
- `operator_approval_runtime_not_created`
- `audit_store_runtime_not_created`
- `backend_health_runtime_not_created`
- `production_resolver_runtime_not_created`
- `cloud_secret_service_not_selected`
- `repository_mode_disabled`

这些阻塞项不能用 fake resolver runtime、developer env plaintext、fixture credential、mock provider、local-smoke profile、operator runbook 文本、audit memory store、repository memory store、static strategy 文档或历史 smoke evidence 替代。

## Future Runtime Task Card Requirements

如果后续重新评审后允许创建 no leakage smoke runtime implementation task card，该任务卡必须至少覆盖：

- disabled-by-default runtime gate，默认不启用 smoke runtime。
- synthetic placeholder fixture only，不读取真实 secret、环境变量 secret、fixture credential、provider raw URL、DSN 或 cloud credential。
- reference-only resolver input scan，覆盖 secret ref、provider/profile、environment、operator approval ref、audit ref、policy version 和 request id。
- sanitized success output scan，只允许 sanitized state、opaque credential handle metadata presence 和 policy / audit / request metadata。
- sanitized failure output scan，只允许 failure code、failure boundary、sanitized diagnostic 和 policy / audit / request metadata。
- audit metadata scan，只允许 audit ref、policy version、secret ref version ref、operator approval ref、failure code 和 sanitized status。
- smoke summary scan，只允许 case id、probe category、pass/fail、failure code、sanitized diagnostic 和 zero side-effect counters。
- negative probe matrix，覆盖 secret-ref-value、credential payload、provider raw URL / DSN / database hostname、authorization header / cookie / user claim、cloud credential marker、cross-environment fallback、fake resolver substitution、repository mode / DB / production API side effect。
- artifact scanner，扫描 docs、fixture、summary、audit metadata 和 smoke output。
- fail-closed leakage detection，任何 secret-looking value、credential payload、raw endpoint、full claim 或 auth header 都必须失败。
- no fake resolver substitution，test-only fake resolver runtime 只能作为被拒绝的 substitution probe。
- no repository mode side effect，不打开 DB、schema marker、migration runner、repository mode 或 public production API。
- offline unit test and static smoke，不调用 provider、云 secret 服务、数据库、fake resolver 或 production resolver。
- side effect counters，所有 secret read、resolver call、provider call、cloud call、DB call、audit write、health check 和 production API call 在 entry review 中必须为零。

该任务卡不得合入 production resolver runtime、production resolver backend client、cloud secret SDK、真实 credential、credential payload runtime、credential handle runtime、approval runtime、audit store runtime、backend health runtime、database connection provider、DB driver、connection factory、SQL、schema marker、migration runner、repository mode runtime 或 public production API。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `real_resolver_no_leakage_runtime_entry_strategy_missing` | `no_secret_leakage` | no leakage smoke runtime strategy 缺失或未被消费 |
| `real_resolver_no_leakage_runtime_entry_task_card_blocked` | `implementation_gate` | 当前仍不能创建 no leakage smoke runtime implementation task card |
| `real_resolver_no_leakage_runtime_entry_scan_surface_missing` | `scan_surface` | input / output / diagnostics / audit / summary 扫描面缺失 |
| `real_resolver_no_leakage_runtime_entry_input_allowlist_missing` | `input_contract` | reference-only input allowlist 缺失 |
| `real_resolver_no_leakage_runtime_entry_output_allowlist_missing` | `output_contract` | success / failure output allowlist 缺失 |
| `real_resolver_no_leakage_runtime_entry_probe_matrix_missing` | `probe_matrix` | negative probe matrix 缺失 |
| `real_resolver_no_leakage_runtime_entry_synthetic_fixture_missing` | `fixture_strategy` | synthetic placeholder fixture source 未定义 |
| `real_resolver_no_leakage_runtime_entry_secret_material_detected` | `artifact_guard` | 文档、fixture、summary 或 future smoke output 出现 secret-looking value |
| `real_resolver_no_leakage_runtime_entry_credential_payload_detected` | `artifact_guard` | 出现 credential payload、raw handle payload 或 secret material |
| `real_resolver_no_leakage_runtime_entry_diagnostic_exposure_detected` | `diagnostics` | diagnostics 暴露 provider raw URL、DSN、backend URL、cloud credential 或 user claim |
| `real_resolver_no_leakage_runtime_entry_audit_exposure_detected` | `audit_metadata` | audit metadata 暴露 secret value、full secret ref、authorization header 或 cookie |
| `real_resolver_no_leakage_smoke_runtime_created_in_entry_review` | `artifact_guard` | 本批创建 smoke runtime、runner、scanner 或 output fixture |
| `real_resolver_no_leakage_smoke_executed_in_entry_review` | `no_side_effects` | 本批执行 smoke runtime、resolver call、provider call 或 cloud call |
| `real_resolver_no_leakage_runtime_entry_fallback_forbidden` | `no_fallback` | 缺少依赖时 fallback 到 fake / env / mock / sample /历史 evidence |
| `real_resolver_no_leakage_runtime_entry_repository_mode_forbidden` | `repository_mode` | entry review 被用于打开 repository mode 成功路径 |
| `real_resolver_no_leakage_runtime_entry_scope_overreach` | `implementation_boundary` | 本批把 resolver runtime、DB provider、audit writer、approval executor、backend health client 或 public API 合入 entry review |

所有失败必须 fail closed，不返回 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、backend endpoint URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw smoke input、raw smoke output、raw request payload 或 raw response payload。

## Sanitized Diagnostics

允许输出：

- `no_leakage_runtime_entry_status`
- `strategy_status`
- `runtime_task_decision`
- `smoke_runtime_status`
- `smoke_runner_status`
- `smoke_execution_status`
- `scan_surface_status`
- `input_allowlist_status`
- `output_allowlist_status`
- `probe_matrix_status`
- `artifact_scan_status`
- `synthetic_fixture_source_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、backend endpoint URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw smoke input、raw smoke output、raw request payload 或 raw response payload。

## No Fallback

- 不允许 no leakage smoke runtime entry review fallback 到 fake resolver runtime、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、previously approved test evidence、operator runbook 文本、repository memory store 或 audit memory store。
- 不允许 production no leakage smoke fallback 到 test-only fake resolver no leakage Go test。
- 不允许缺少 synthetic placeholder fixture、artifact scanner、input / output allowlist、negative probe matrix、credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、backend profile、provider profile、secret ref 或 environment binding 时创建成功路径。
- 不把本 entry review 写成 no leakage smoke runtime ready、smoke executed、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 smoke runtime、不创建 smoke output fixture、不创建 artifact scanner、不创建 credential payload、不创建 credential handle、不执行 approval runtime、不创建 audit store、不写 audit event、不执行 backend health check、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `smoke_runtime_created_count=0`
- `smoke_runner_created_count=0`
- `smoke_runtime_execution_count=0`
- `artifact_scanner_created_count=0`
- `smoke_output_fixture_created_count=0`
- `network_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `credential_handle_runtime_created_count=0`
- `operator_approval_runtime_execution_count=0`
- `audit_store_created_count=0`
- `audit_writer_created_count=0`
- `audit_event_write_count=0`
- `backend_health_check_count=0`
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.md`
- `docs/task-cards/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.json`
- `scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py`

不得新增或启用以下 artifact：

- no secret leakage smoke runtime implementation task card
- no secret leakage smoke runtime
- no secret leakage smoke runner
- no secret leakage artifact scanner
- no secret leakage smoke output fixture
- production resolver runtime
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- approval runtime / approval executor
- audit store runtime
- audit writer
- backend health runtime
- backend health client
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- public production API

## Real Resolver Entry Alignment

本批完成后，真实 resolver runtime implementation entry refresh 中的 no leakage blocker 从“strategy defined runtime not created”推进为“no leakage smoke runtime implementation entry review 已定义但 runtime task card 仍 blocked”。这只减少静态前置不明确性，不打开 production resolver runtime implementation task card。

真实 resolver runtime 后续仍必须等待 no leakage smoke runtime、credential handle runtime、operator approval runtime、audit store / writer、backend health runtime 和生产前复核各自独立评审；不得把本 entry review 解释为 real resolver runtime ready。

## 后续推进

本批之后不应直接创建 no leakage smoke runtime implementation task card。建议下一步在以下方向中选择一个单独开题：

1. `production-secret-backend-credential-handle-runtime-implementation-entry-refresh`
2. `production-secret-backend-operator-approval-runtime-implementation-entry-refresh`
3. `production-secret-backend-audit-store-runtime-implementation-entry-refresh`
4. `production-secret-backend-resolver-backend-health-runtime-implementation-entry-refresh`
5. `production-secret-backend-real-resolver-runtime-implementation-entry-refresh`

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
