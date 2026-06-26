# Saved Workflow Draft Database Secret Resolver Runtime Dependency Refresh v1

更新时间：2026-06-23

## 专题定位

`Saved Workflow Draft Database Secret Resolver Runtime Dependency Refresh v1` 承接 database secret resolver implementation entry review、schema marker runtime dependency refresh、database connection provider entry refresh v2，以及 Production Secret Backend 的 real resolver、credential handle、operator approval、audit store、backend health 和 no leakage smoke runtime entry review，用于把 future saved draft database secret resolver runtime 之前的依赖矩阵重新收束。

结论：状态为 `draft_database_secret_resolver_runtime_dependency_refresh_defined`。entry decision 为 `database_secret_resolver_runtime_still_blocked_before_implementation_task_card`。本批只固定 secret resolver runtime 对 production resolver、credential handle、operator approval、audit store、backend health、no leakage smoke、connection provider、schema marker、repository mode 和 auth context 的静态依赖；不创建 database secret resolver implementation task card、secret resolver runtime、production resolver runtime、credential handle runtime、approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database connection provider、SQL、schema marker、repository mode runtime、OIDC middleware、token validation、membership adapter 或 production API。

## 输入证据

- `workflow-saved-draft-database-secret-resolver-readiness-v1` 已定义 secret ref key、resolver input / result shape、disabled runtime、sanitized diagnostics、failure taxonomy、environment binding 和 offline fake resolver strategy。
- `workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1` 已确认 secret resolver implementation entry 仍不打开。
- `workflow-saved-draft-database-connection-provider-implementation-entry-refresh-v2` 已确认 connection provider task card 仍 blocked。
- `workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1` 已确认 schema marker runtime、runner、DB provider、repository mode 和 auth runtime 依赖仍 blocked。
- `production-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1` 已确认 production resolver runtime task card 仍 blocked。
- `production-secret-backend-credential-handle-runtime-implementation-entry-review-v1`、`production-secret-backend-operator-approval-runtime-implementation-entry-review-v1`、`production-secret-backend-audit-store-runtime-implementation-entry-refresh-v3`、`production-secret-backend-resolver-backend-health-runtime-implementation-entry-review-v1` 和 `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-implementation-entry-review-v1` 均只固定 blocked-before-task-card 结论。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 只实现 test-only、默认 disabled 的 fake resolver runtime，不能解锁 production connection provider、repository mode 或 production API。

## Dependency Refresh Decision

| dependency | 本次结论 | 说明 |
| --- | --- | --- |
| saved draft secret resolver contract | `static_contract_defined_no_runtime` | secret ref、input / output、diagnostics 和 failure taxonomy 已定义 |
| secret reference manifest | `reference_only_resolver_disabled` | manifest 不保存 secret value，不启用 resolver，不允许 cloud call |
| production resolver runtime | `blocked_no_production_resolver_runtime` | production resolver runtime task card 仍未创建 |
| credential handle runtime | `blocked_no_credential_handle_runtime` | opaque handle boundary 已定义，runtime / issuer / handle 未创建 |
| operator approval runtime | `blocked_no_operator_approval_runtime` | evidence shape 已定义，approval executor 未创建或执行 |
| audit store runtime | `blocked_no_audit_store_runtime` | handoff / contract / delivery-idempotency 已定义，store / writer / event 未创建 |
| backend health runtime | `blocked_no_backend_health_runtime` | boundary 已定义，health runtime / client / check 未创建 |
| no leakage smoke runtime | `blocked_no_no_leakage_smoke_runtime` | strategy / entry review 已定义，runtime 和 smoke output 未创建 |
| test-only fake resolver | `implemented_test_only_disabled_cannot_unlock_production` | 只能作为离线替身，不能作为 production resolver |
| connection provider dependency | `blocked_no_connection_provider` | DB provider、driver、pool、lifecycle 和 smoke runtime 均未创建 |
| schema marker dependency | `blocked_no_schema_marker_runtime` | marker reader / writer、schema version table 和 runner 未创建 |
| repository mode dependency | `blocked_no_repository_mode_runtime` | repository mode 仍 fail closed |
| auth membership dependency | `static_upstream_evidence_defined_no_auth_runtime` | OIDC upstream evidence 已有，但 middleware / token validation / membership runtime 未创建 |
| sanitized diagnostics dependency | `static_contract_defined_no_runtime_emission` | diagnostics allowlist 已定义，runtime emission gate 未创建 |
| environment binding dependency | `static_contract_defined_no_runtime_enforcement` | dev / test / prod 绑定规则已定义，runtime enforcement 未创建 |

## Runtime Blockers

- `database_secret_resolver_runtime_not_created`
- `production_resolver_runtime_task_card_not_created`
- `production_resolver_runtime_not_created`
- `credential_handle_runtime_not_created`
- `operator_approval_runtime_not_created`
- `audit_store_runtime_not_created`
- `backend_health_runtime_not_created`
- `no_secret_leakage_smoke_runtime_not_created`
- `cloud_secret_service_not_selected`
- `connection_provider_not_created`
- `schema_marker_runtime_not_created`
- `repository_mode_runtime_disabled`
- `auth_middleware_not_created`
- `membership_adapter_not_created`

## Future Task Requirements

后续若要创建 saved draft database secret resolver implementation task card，必须先独立补齐：

1. secret resolver interface owner、输入输出、failure envelope 和 sanitized diagnostics runtime。
2. secret ref / environment / caller purpose / request id / audit ref / policy version 的 runtime binding。
3. production resolver runtime 与 backend profile 的可复验证据。
4. opaque credential handle runtime、issuer、lifecycle 和 payload 禁止边界。
5. operator approval runtime、operator identity、ticket / change window 和 dual-control evidence。
6. audit store runtime、writer ownership、delivery / idempotency 和 retention / redaction runtime。
7. backend health runtime、client、health check 和 failure mapping。
8. no leakage smoke runtime、artifact scan surfaces、forbidden field probe 和 smoke output。
9. connection provider handoff、schema marker preflight、repository mode fail-closed integration 和 auth membership context。
10. missing secret、disabled backend、resolution denied、credential missing、environment mismatch、connection unavailable 和 repository disabled 的负向 smoke matrix。

## Failure Mapping

| 条件 | failure code |
| --- | --- |
| repository mode reserved / runtime 未打开 | `repository_store_disabled` |
| unknown store mode | `invalid_draft_store_mode` |
| secret ref 缺失、backend disabled、resolver runtime 不可用或 resolution denied | `draft_store_unavailable` |
| credential handle runtime、operator approval runtime、audit store runtime、backend health runtime 或 no leakage smoke runtime 缺失 | `draft_store_unavailable` |
| connection provider、driver、lifecycle 或 smoke runtime 不可用 | `draft_store_unavailable` |
| request id / audit ref / purpose / policy version 缺失 | `draft_audit_context_missing` |
| marker 缺失、version mismatch 或 marker 不可读 | `draft_store_migration_unavailable` |
| repository actor context 不可信 | `draft_auth_context_contract_mismatch` |

失败时不得返回 secret value、DSN、provider URL、raw token、database host、database error detail、credential payload、credential handle、完整 claim、raw request payload 或草案主体；不得回退 `memory_dev`、sample、fixture、dev route、test auth、fake query executor、test-only fake resolver、developer env plaintext 或 local-smoke profile。

## No Fallback / No Side Effects

- 不允许 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 因本 refresh 被启用。
- 不允许 service startup、HTTP route、store selector 或 saved draft request path 读取 secret、调用 fake resolver、调用 production resolver、创建 credential handle、执行 approval runtime、写 audit event、执行 backend health check、运行 no leakage smoke、打开数据库连接、读取 schema marker 或校验 OIDC token。
- 不允许 test-only fake resolver runtime、reference-only secret manifest、operator runbook 文本、audit handoff 文档、backend health boundary、no leakage strategy 或 static schema artifact 替代 production resolver runtime。
- 本批 checker 只读取 committed 文档、fixture、源码和 `check-repo.py` 注册顺序；不启动服务、不读取 env secret、不调用 fake resolver、不调用 production resolver、不调用云 secret 服务、不连接数据库、不导入 driver、不运行 SQL、不读写 schema marker、不创建 runner、不调用 OIDC、不校验 token、不创建 membership adapter、不创建 production API、不执行 workflow。
- side effect counters 必须保持 `secret_resolver_call_count=0`、`fake_resolver_call_count=0`、`production_resolver_call_count=0`、`real_secret_read_count=0`、`environment_secret_read_count=0`、`secret_value_read_count=0`、`cloud_secret_call_count=0`、`network_call_count=0`、`credential_payload_created_count=0`、`credential_handle_created_count=0`、`credential_handle_runtime_created_count=0`、`operator_approval_runtime_execution_count=0`、`audit_store_runtime_created_count=0`、`audit_writer_created_count=0`、`audit_event_write_count=0`、`backend_health_runtime_created_count=0`、`backend_health_check_count=0`、`no_secret_leakage_smoke_runtime_created_count=0`、`database_connection_count=0`、`driver_open_count=0`、`sql_execution_count=0`、`schema_marker_read_count=0`、`schema_marker_write_count=0`、`repository_mode_enablement_count=0`、`production_api_call_count=0`、`executor_call_count=0`、`confirmation_call_count=0`、`business_writeback_count=0` 和 `replay_call_count=0`。

## 后续方向

本次 refresh 后，saved draft database secret resolver implementation task card、production resolver runtime implementation task card、connection provider implementation task card、schema marker runtime implementation task card 和 repository mode runtime task card 仍不创建。后续若继续 durable store 上游，应从以下方向选择一个独立推进：

1. `token validation schema / auth middleware runtime entry review`：消费 OIDC upstream evidence，评审 token schema、middleware、membership adapter 和 runtime smoke 是否可拆 task card。
2. `production resolver runtime blocker consolidation`：仅在 credential handle、approval、audit store、backend health 和 no leakage runtime 的 blocker 需要重新统一排序时开题。
3. `connection provider runtime dependency refresh`：在 secret resolver、schema marker 和 auth runtime 依赖出现新证据后重新评审。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-schema-marker-runtime-dependency-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-refresh-v1.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不创建 `docs/task-cards/workflow-saved-draft-database-secret-resolver-implementation-v1-plan.md` 或 `docs/task-cards/production-secret-backend-real-resolver-runtime-implementation-v1-plan.md`。
- 不创建 database secret resolver runtime、production resolver runtime、credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database connection provider、schema marker reader / writer、SQL migration、migration runner 或 repository mode runtime。
- 不启用 repository mode，不把 `RADISHMIND_WORKFLOW_SAVED_DRAFT_STORE=repository` 变成成功路径。
- 不接 OIDC middleware、token validation、membership adapter、production API、executor、confirmation、writeback 或 replay。
- 不把 `draft_database_secret_resolver_runtime_dependency_refresh_defined` 解释为 secret resolver ready、production resolver ready、database ready、repository mode ready、durable persistence ready、production API ready 或 production ready。
