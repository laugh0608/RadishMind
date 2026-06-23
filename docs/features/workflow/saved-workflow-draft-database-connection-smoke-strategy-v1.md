# Saved Workflow Draft Database Connection Smoke Strategy v1

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Database Connection Smoke Strategy v1` 承接 [Saved Workflow Draft Database Role Policy Readiness v1](saved-workflow-draft-database-role-policy-readiness-v1.md)，只定义 future saved draft database connection provider 之前的 connection smoke 策略、记录形态和停止线。

结论：状态为 `draft_database_connection_smoke_strategy_defined`。本批只固定 future explicit test database boundary、safe placeholder credential handoff、smoke input / output record shape、role denial cases、no leakage scan、environment binding、manual-only execution boundary、zero production side effect 和 artifact guard；不创建 connection smoke runner、connection smoke runtime、DB provider、DB driver import、DSN parser runtime、connection factory、SQL、schema marker、migration runner、repository mode runtime、OIDC middleware、token validator、membership adapter、production API、executor、confirmation、writeback 或 replay。

## 输入证据

- `workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1` 已定义 driver / DSN / TLS policy、DSN redaction、environment binding 和 forbidden diagnostics scan 的 future 边界。
- `workflow-saved-draft-database-role-policy-readiness-v1` 已定义 runtime DML role、migration DDL / marker role、least privilege review 和 cross-environment denial smoke 前置。
- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1` 已确认 connection provider implementation task card 仍为 `blocked_before_implementation_task_card`。
- `workflow-saved-draft-database-secret-resolver-readiness-v1` 与 `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已定义 resolver input / result、disabled runtime 和 failure taxonomy，但 resolver implementation 仍不创建。
- `workflow-saved-draft-repository-mode-runtime-boundary-review-v1` 已确认 repository mode runtime task card 仍 blocked。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 已确认 no leakage smoke runtime 仍 blocked；本批只能定义 scan surfaces 和 record requirement，不执行 scanner。

## Strategy Decision

| strategy area | 本次结论 | 阻塞或停止线 |
| --- | --- | --- |
| explicit test database boundary | `defined_no_database_connection` | 后续 smoke 只能在显式 test 环境或人工运行记录中执行；本批不连接数据库 |
| placeholder credential handoff | `defined_metadata_only_no_secret_resolution` | 只允许 placeholder credential reference / opaque handle metadata；本批不读取 secret、不调用 resolver |
| smoke input shape | `defined_static_contract_only` | 输入只包含 environment、driver / TLS / role policy refs、placeholder credential ref、request / audit refs 和 expected denial cases |
| smoke output shape | `defined_static_record_only` | 输出只能是 sanitized status、policy ids、redacted fingerprints、failure codes、role denial result 和 no leakage scan summary |
| role denial cases | `defined_no_runtime_denial_execution` | 定义 runtime role DDL denied、migration role repository denied、cross-environment role denied；本批不执行 denial smoke |
| no leakage scan | `defined_static_scan_requirement` | 定义 scan surfaces 和 forbidden terms；本批不创建 scanner、不读取真实 smoke output |
| manual-only execution boundary | `defined_no_fast_baseline_runtime` | future connection smoke 不能进入 fast baseline 的联网 / DB 连接路径 |
| connection provider dependency | `blocked_before_implementation_task_card` | provider task card 仍需等待 smoke strategy、secret resolver、schema marker、repository mode runtime 和 auth upstream evidence 可复验 |

## Smoke Contract Boundary

future connection smoke 必须遵守以下边界：

- smoke target 只能是显式 test database 或人工审查过的 test-prod smoke target，且必须绑定 `environment_id`、`policy_version`、`driver_policy_ref`、`tls_policy_ref`、`role_policy_ref` 和 `smoke_profile_id`。
- smoke credential 只能来自 safe placeholder credential handoff 或后续明确限定的 test-only opaque credential handle reference；不得使用 production secret、developer plaintext env、raw DSN、完整 secret ref 或 credential payload。
- smoke input 不得包含 raw host、raw port、database name、username、password、raw DSN、authorization header、cookie、full user claim、secret value 或 SQL text。
- smoke output 不得包含 raw DSN、database host、username、database error detail、credential payload、SQL detail、schema marker payload、full role grant 或草案主体。
- 允许输出 metadata-only `smoke_record_id`、`environment_id`、policy ids、redacted target fingerprint、redacted role fingerprint、sanitized status、failure code、role denial case id、no leakage scan summary 和 side effect summary。
- role denial cases 至少覆盖 runtime DML role DDL denied、runtime DML role marker write denied、migration DDL / marker role repository adapter denied、cross-environment role binding denied 和 unknown role class denied。
- future smoke 只可由独立任务卡创建 runner 或人工 run record；本策略本身不创建 runner、不执行 smoke、不连接数据库、不导入 driver、不运行 SQL。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| smoke target 未显式声明、环境不可信或 smoke profile 缺失 | `draft_store_unavailable` |
| placeholder credential handoff 缺失、包含 secret material 或环境不一致 | `draft_store_unavailable` |
| driver / DSN / TLS / role policy ref 缺失或不一致 | `draft_store_unavailable` |
| no leakage scan strategy 缺失或 scan surface 不完整 | `draft_store_unavailable` |
| role denial case 缺失或未覆盖 cross-environment denial | `draft_store_unavailable` |
| future smoke result 不可观测或 sanitized record contract mismatch | `draft_store_unavailable` |
| marker 缺失或 migration 未应用 | `draft_schema_migration_not_applied` |
| marker version mismatch | `draft_store_schema_version_mismatch` |
| marker 不可读、runner 状态不可观测或 lock 状态不可信 | `draft_store_migration_unavailable` |
| stored record contract mismatch | `draft_store_contract_mismatch` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回 raw DSN、database host、database error detail、secret value、credential payload、完整 secret ref、完整 credential handle、完整 user claim、authorization header、cookie、SQL detail、schema marker payload、role grant 或草案主体。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 strategy 被启用。
- 不允许 fast baseline、service startup、HTTP route、store selector 或 saved draft request path 自动执行 smoke、解析 secret、打开 driver、连接数据库、读取 marker 或写 repository。
- 不允许 test-only fake resolver runtime、fake query executor smoke、static schema artifact、rollback evidence、driver / DSN / TLS policy、role policy 或 container smoke 替代 future connection smoke strategy 的后续实现前置。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不导入 driver、不创建 DSN parser、不创建 role policy runtime、不连接数据库、不运行 connection smoke、不运行 no leakage scan runtime、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`secret_value_read_count=0`、`credential_handle_created_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`database_connection_count=0`、`driver_import_count=0`、`driver_open_count=0`、`dsn_parse_count=0`、`dsn_build_count=0`、`tls_config_create_count=0`、`role_policy_evaluation_count=0`、`role_claim_parse_count=0`、`role_grant_count=0`、`connection_smoke_count=0`、`no_leakage_scan_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`repository_read_count=0`、`repository_write_count=0`、`oidc_call_count=0`、`token_validation_count=0`、`membership_lookup_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本批后，connection provider implementation task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `Radish OIDC upstream evidence refresh` 已固定为 `radish_oidc_token_membership_upstream_evidence_refresh_defined`；后续 auth runtime 仍需独立 entry review。
2. `connection lifecycle readiness` 已由 `workflow-saved-draft-database-connection-lifecycle-readiness-v1` 固定，状态为 `draft_database_connection_lifecycle_readiness_defined`。
3. `connection provider implementation entry refresh v2` 已固定为 `draft_database_connection_provider_implementation_entry_refresh_v2_defined`；connection smoke 只满足静态 strategy，不打开 provider task card。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-smoke-strategy-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-role-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-driver-dsn-tls-policy-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-repository-mode-runtime-boundary-review-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-connection-provider-implementation-v1-plan.md`。
- 不创建 connection smoke runner、connection smoke runtime、database connection provider、DB driver import、DSN parser runtime、connection factory、role policy runtime、no leakage scan runtime、database query executor、schema marker reader / writer、SQL migration、migration runner、runner command、dry-run output 或 smoke output。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_connection_smoke_strategy_defined` 解释为 connection smoke executed、connection provider ready、database ready、secret resolver ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
