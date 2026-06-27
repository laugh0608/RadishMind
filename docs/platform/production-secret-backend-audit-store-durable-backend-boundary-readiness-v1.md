# Production Secret Backend Audit Store Durable Backend Boundary Readiness v1

更新时间：2026-06-27

## 文档目的

本文档在 `Production Secret Backend Audit Store Runtime Implementation Entry Refresh v3` 之后，固定 future production secret backend audit store durable backend 的职责边界。

对应切片：`production-secret-backend-audit-store-durable-backend-boundary-readiness-v1`。

结论：状态为 `audit_store_durable_backend_boundary_readiness_defined`。本批只把 durable backend owner、storage adapter responsibility、allowed metadata、forbidden material、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为静态前置证据；不选择 durable audit backend，不创建 audit store runtime implementation task card，不创建 audit store runtime、audit writer、runtime event schema、delivery runtime、idempotency runtime、production resolver runtime、cloud secret client、DB provider、repository mode 或 public production API。

## 输入证据

- `audit_store_runtime_implementation_entry_refresh_v3_defined` 已确认 audit store runtime task card 仍 blocked，且 durable backend、writer runtime、runtime event schema、delivery runtime、idempotency runtime、approval、credential handle、backend health 和 no leakage runtime 仍是未满足依赖。
- `audit_store_ownership_boundary_readiness_defined` 已固定 store / writer / schema / delivery / policy / dependency owner reference，但 durable backend owner 在 runtime task card 前仍需单独收束。
- `audit_store_delivery_idempotency_runtime_boundary_readiness_defined` 已固定 delivery owner、idempotency key owner、duplicate handling、retry / failure semantics 和 delivery result envelope。
- `audit_store_contract_event_schema_readiness_defined` 已固定 metadata-only event schema、writer input / output contract、idempotency key reference、retention / redaction binding 和 artifact guard。
- `implementation_readiness_defined` 当前仍保持 `production_secret_backend_status=not_satisfied`、`audit_store_runtime_task_card_status=not_created`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created` 和 `audit_event_delivery_status=not_executed`。
- `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_refresh_defined`、`resolver_backend_health_runtime_implementation_entry_refresh_defined`、`operator_approval_runtime_implementation_entry_refresh_defined` 和 `credential_handle_runtime_implementation_entry_refresh_defined` 均确认相关 runtime task card 仍 blocked。

## Durable Backend Boundary

| boundary | 当前结论 | 说明 |
| --- | --- | --- |
| durable backend boundary | `defined_static_only` | 只固定 future backend contract，不选择具体实现 |
| backend selection | `deferred_until_runtime_task_card` | 不选择 DB、object store、queue、log sink 或 cloud audit service |
| durable backend owner | `static_boundary_defined` | owner 只对 storage adapter contract 负责 |
| storage adapter responsibility | `metadata_only_persistence_boundary` | 只允许 future metadata-only audit event / delivery result reference |
| writer ownership | `separate_runtime_dependency` | writer runtime 仍必须后续独立评审 |
| runtime event schema | `separate_runtime_dependency` | 静态 schema 不等于 runtime schema artifact |
| delivery / idempotency runtime | `separate_runtime_dependency` | duplicate / retry / delivery result 仍不执行 |
| audit store runtime task card | `not_created` | 本批不创建 runtime task card |
| audit store runtime | `not_created` | 不创建 store、writer、event 或 delivery |
| production resolver runtime | `not_created` | 不创建 resolver runtime 或 cloud client |
| database / repository / API | `blocked` | DB provider、repository mode 和 public production API 不打开 |

## Allowed Metadata

future durable backend 只允许接收 metadata-only 字段：

- `audit_ref`
- `event_ref`
- `event_kind`
- `event_version`
- `request_id`
- `workspace_ref`
- `environment`
- `provider_profile_ref`
- `backend_profile_ref`
- `secret_ref_key`
- `credential_handle_ref`
- `approval_evidence_ref`
- `idempotency_key_ref`
- `delivery_result_ref`
- `retention_policy_ref`
- `redaction_profile_ref`
- `policy_version`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`

这些字段都必须是 reference、short key、status 或 policy version，不得承载 secret material、payload、raw claim、raw request、raw response、provider URL、DSN、cloud credential 或 secret-derived hash。

## Forbidden Material

durable backend boundary 禁止接收、持久化或诊断输出：

- raw secret、secret value、password、token、API key、authorization header、cookie
- provider raw URL、resolver backend URL、backend endpoint URL、DSN、database hostname、database error detail
- cloud credential、credential payload、full credential handle、full secret ref value
- raw operator claim、raw user claim、raw approval payload、raw ticket payload
- raw request payload、raw response payload、raw audit payload、raw payload hash、secret-derived hash
- binary payload、fixture credential、developer env plaintext、committed secret value

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_durable_backend_boundary_dependency_missing` | `dependency_chain` | 必需 audit store readiness / refresh 证据缺失 |
| `audit_store_durable_backend_boundary_task_card_forbidden` | `implementation_gate` | 本批创建 audit store runtime implementation task card |
| `audit_store_durable_backend_boundary_backend_selected_forbidden` | `durable_backend_selection` | 本批选择或启用 durable backend |
| `audit_store_durable_backend_boundary_writer_runtime_forbidden` | `audit_writer` | 本批创建 writer runtime |
| `audit_store_durable_backend_boundary_runtime_schema_forbidden` | `event_schema` | 本批创建 runtime event schema artifact |
| `audit_store_durable_backend_boundary_delivery_forbidden` | `delivery` | 本批执行 delivery、duplicate detection 或 retry |
| `audit_store_durable_backend_boundary_secret_material_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret-bearing material |
| `audit_store_durable_backend_boundary_fallback_forbidden` | `no_fallback` | 缺少 durable backend 时 fallback 到 memory / fake / sample |
| `audit_store_durable_backend_boundary_side_effect_detected` | `no_side_effects` | 本批读取 secret、调用 provider / cloud / DB、写 audit event 或启用 repository mode |
| `audit_store_durable_backend_boundary_scope_overreach` | `implementation_boundary` | 本批合入 resolver runtime、cloud client、DB provider、repository mode 或 public API |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_durable_backend_boundary_status`
- `durable_backend_selection_status`
- `durable_backend_owner_status`
- `storage_adapter_responsibility_status`
- `audit_store_runtime_task_card_status`
- `audit_store_runtime_status`
- `audit_writer_status`
- `runtime_event_schema_status`
- `delivery_runtime_status`
- `idempotency_runtime_status`
- `production_resolver_runtime_status`
- `cloud_secret_service_status`
- `database_connection_provider_status`
- `repository_mode_status`
- `production_api_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 forbidden material 中列出的任一原文、完整 payload 或 secret-derived material。

## No Fallback / No Side Effects

- 不允许 durable backend boundary fallback 到 fake resolver runtime、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、operator runbook 文本、repository memory store、audit memory store、static handoff envelope、static contract、static ownership boundary 或 static delivery / idempotency boundary。
- 不允许缺少 durable backend、writer runtime、runtime event schema、delivery runtime、idempotency runtime、operator approval runtime、credential handle runtime、backend health runtime 或 no leakage runtime 时创建 audit success。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / refresh fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 backend health check、不执行 approval runtime、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-durable-backend-boundary-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-durable-backend-boundary-readiness-v1.py`

不得新增或启用 audit store runtime implementation task card、durable audit backend、audit store runtime、audit writer、audit event writer / runner、runtime event schema artifact、audit delivery runtime、audit idempotency runtime、duplicate detector runtime、retry executor、production resolver runtime、cloud secret SDK / client、secret value fixture、production credential file、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、database connection provider、DB driver、SQL migration、schema marker reader / writer、migration runner、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 audit store runtime task card 仍不能创建。后续若继续 audit store runtime 前置，应选择 writer runtime boundary、runtime event schema materialization readiness、delivery runtime readiness、idempotency runtime readiness 或重新评审 audit store runtime implementation entry；不得直接创建 audit store runtime implementation task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-durable-backend-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-implementation-entry-refresh-v3.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
