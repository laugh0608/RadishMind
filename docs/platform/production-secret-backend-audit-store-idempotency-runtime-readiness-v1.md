# Production Secret Backend Audit Store Idempotency Runtime Readiness v1

更新时间：2026-06-27

## 文档目的

本文档在 `Production Secret Backend Audit Store Delivery Runtime Readiness v1` 之后，固定 future production secret backend audit store idempotency runtime 的职责边界。

对应切片：`production-secret-backend-audit-store-idempotency-runtime-readiness-v1`。

结论：状态为 `audit_store_idempotency_runtime_readiness_defined`。本批只把 future idempotency runtime owner、metadata-only idempotency input envelope、idempotency result reference、idempotency key reference policy、duplicate detection policy、replay decision policy、delivery runtime dependency、runtime event schema dependency、writer runtime dependency、durable backend dependency、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard 固定为静态前置证据；不创建 audit store runtime implementation task card，不创建 idempotency runtime implementation task card，不创建 idempotency runtime、idempotency key store、duplicate detector、replay executor、retry executor、delivery runtime、runtime event schema artifact、writer runtime、audit store runtime、audit writer、audit event writer、production resolver runtime、cloud secret client、DB provider、repository mode 或 public production API。

## 输入证据

- `audit_store_runtime_implementation_entry_refresh_v3_defined` 已确认 audit store runtime task card 仍 blocked，且 durable backend、writer runtime、runtime event schema、delivery runtime、idempotency runtime、approval、credential handle、backend health 和 no leakage runtime 仍是未满足依赖。
- `audit_store_delivery_idempotency_runtime_boundary_readiness_defined` 已固定 delivery owner、idempotency key owner、duplicate handling、retry / failure semantics 和 delivery result envelope 的静态边界。
- `audit_store_durable_backend_boundary_readiness_defined` 已固定 durable backend owner 与 storage adapter responsibility，但 concrete durable backend 仍为 `not_selected`。
- `audit_store_writer_runtime_boundary_readiness_defined` 已固定 writer runtime owner、metadata-only writer input envelope 与 writer result reference，但 writer runtime 仍为 `not_created`。
- `audit_store_runtime_event_schema_materialization_readiness_defined` 已固定 runtime event schema materialization owner、schema version pin、event kind allowlist source 和 writer input compatibility，但 runtime schema artifact 仍为 `not_created`。
- `audit_store_delivery_runtime_readiness_defined` 已固定 delivery runtime owner、metadata-only delivery input / result envelope、retry policy 和 duplicate handling dependency，但 delivery runtime 仍为 `not_created`，且 idempotency runtime 仍为 separate runtime dependency。
- `implementation_readiness_defined` 当前仍保持 `production_secret_backend_status=not_satisfied`、`audit_store_runtime_task_card_status=not_created`、`audit_store_runtime_status=not_created`、`audit_writer_status=not_created`、`audit_delivery_runtime_status=not_created`、`audit_idempotency_runtime_status=not_created` 和 `audit_event_delivery_status=not_executed`。

## Idempotency Runtime Boundary

future idempotency runtime 只能消费 metadata-only references，并且必须在缺少 delivery runtime、runtime event schema、writer runtime、durable backend、operator approval runtime、credential handle runtime、backend health runtime 或 no leakage runtime 时 fail closed。

| boundary | 当前结论 | 说明 |
| --- | --- | --- |
| idempotency runtime readiness | `defined_static_only` | 只固定 future idempotency runtime contract |
| idempotency runtime task card | `not_created` | 本批不创建 idempotency runtime implementation task card |
| idempotency runtime | `not_created` | 不创建 key store、duplicate detector、replay executor 或 retry runtime |
| idempotency owner | `static_boundary_defined` | 只定义 future idempotency owner reference |
| idempotency input envelope | `metadata_only_static_envelope_defined` | 只允许 reference / status / policy version |
| idempotency result envelope | `metadata_only_static_envelope_defined` | 只定义 result reference，不持久化 decision |
| idempotency key policy | `static_reference_only_policy_defined` | 只引用 key ref，不派生 payload hash |
| duplicate detection policy | `static_fail_closed_policy_defined` | 不执行 duplicate detection |
| replay decision policy | `static_fail_closed_policy_defined` | 不执行 replay 或 delivery suppression |
| delivery runtime dependency | `static_boundary_defined_delivery_not_created` | delivery runtime 仍未创建 |
| runtime event schema dependency | `static_boundary_defined_schema_not_created` | schema artifact 仍未创建 |
| writer runtime dependency | `static_boundary_defined_writer_not_created` | writer runtime 仍未创建 |
| durable backend dependency | `static_boundary_defined_backend_not_selected` | durable backend 仍未选择 |
| audit event delivery | `not_executed` | 本批不写 event，不执行 delivery |
| audit store runtime | `not_created` | 不创建 store、writer、event、delivery 或 idempotency runtime |
| production resolver runtime | `not_created` | 不创建 resolver runtime 或 cloud client |
| database / repository / API | `blocked` | DB provider、repository mode 和 public production API 不打开 |

## Allowed Metadata

future idempotency runtime 只允许接收 metadata-only 字段：

- `idempotency_request_ref`
- `idempotency_result_ref`
- `idempotency_owner_ref`
- `idempotency_key_ref`
- `idempotency_key_policy_ref`
- `duplicate_decision_ref`
- `replay_decision_ref`
- `delivery_request_ref`
- `delivery_result_ref`
- `delivery_attempt_ref`
- `delivery_status`
- `schema_ref`
- `schema_version`
- `writer_input_ref`
- `writer_result_ref`
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
- `retry_policy_ref`
- `durable_backend_ref`
- `retention_policy_ref`
- `redaction_profile_ref`
- `policy_version`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`

这些字段都必须是 reference、short key、status 或 policy version，不得承载 secret material、payload、raw request、raw response、provider URL、DSN、cloud credential、database detail、binary payload、raw idempotency payload、raw duplicate probe、raw replay payload、raw delivery payload、raw event payload、raw writer payload 或 secret-derived hash。

## Forbidden Material

idempotency runtime readiness boundary 禁止接收、写入或诊断输出：

- raw secret、secret value、password、token、API key、authorization header、cookie
- provider raw URL、resolver backend URL、backend endpoint URL、DSN、database hostname、database error detail
- cloud credential、credential payload、full credential handle、full secret ref value
- raw operator claim、raw user claim、raw approval payload、raw ticket payload
- raw request payload、raw response payload、raw audit payload、raw writer payload、raw event payload、raw delivery payload、raw idempotency payload、raw idempotency result、raw duplicate probe、raw replay payload
- raw payload hash、delivery payload hash、event payload hash、idempotency payload hash、secret-derived hash、schema payload
- binary payload、fixture credential、developer env plaintext、committed secret value

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_idempotency_runtime_readiness_dependency_missing` | `dependency_chain` | 必需 audit store readiness / delivery readiness 证据缺失 |
| `audit_store_idempotency_runtime_readiness_task_card_forbidden` | `implementation_gate` | 本批创建 audit store runtime 或 idempotency runtime implementation task card |
| `audit_store_idempotency_runtime_readiness_idempotency_runtime_forbidden` | `idempotency_runtime` | 本批创建 idempotency runtime、key store、duplicate detector、replay executor 或 retry runtime |
| `audit_store_idempotency_runtime_readiness_duplicate_detection_forbidden` | `duplicate_detection` | 本批执行 duplicate detection 或持久化 duplicate decision |
| `audit_store_idempotency_runtime_readiness_replay_decision_forbidden` | `replay_decision` | 本批执行 replay decision、delivery suppression 或 replay executor |
| `audit_store_idempotency_runtime_readiness_delivery_runtime_forbidden` | `delivery_runtime` | 本批创建或执行 delivery runtime |
| `audit_store_idempotency_runtime_readiness_event_write_forbidden` | `audit_event_write` | 本批写 audit event 或生成 writer result |
| `audit_store_idempotency_runtime_readiness_secret_material_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret-bearing material |
| `audit_store_idempotency_runtime_readiness_fallback_forbidden` | `no_fallback` | 缺少 idempotency runtime 依赖时 fallback 到 memory / fake / sample |
| `audit_store_idempotency_runtime_readiness_side_effect_detected` | `no_side_effects` | 本批读取 secret、调用 provider / cloud / DB、写 audit event 或执行 idempotency logic |
| `audit_store_idempotency_runtime_readiness_scope_overreach` | `implementation_boundary` | 本批合入 resolver runtime、cloud client、DB provider、repository mode 或 public API |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_idempotency_runtime_readiness_status`
- `idempotency_runtime_implementation_task_card_status`
- `idempotency_runtime_status`
- `idempotency_owner_status`
- `idempotency_input_envelope_status`
- `idempotency_result_envelope_status`
- `idempotency_key_policy_status`
- `idempotency_duplicate_detection_policy_status`
- `idempotency_replay_decision_policy_status`
- `delivery_runtime_dependency_status`
- `runtime_event_schema_dependency_status`
- `writer_runtime_dependency_status`
- `durable_backend_dependency_status`
- `audit_store_runtime_task_card_status`
- `audit_store_runtime_status`
- `audit_writer_status`
- `audit_event_write_status`
- `audit_event_delivery_status`
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

禁止输出 forbidden material 中列出的任一原文、完整 payload、raw idempotency payload、raw duplicate probe、raw replay payload、raw delivery payload、raw event payload、raw writer payload 或 secret-derived material。

## No Fallback / No Side Effects

- 不允许 idempotency runtime readiness fallback 到 fake resolver runtime、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、operator runbook 文本、repository memory store、audit memory store、static handoff envelope、static contract、static ownership boundary、static delivery idempotency boundary、static durable backend boundary、static writer boundary、static runtime schema materialization readiness、static delivery readiness、historical audit event、delivery sample 或 duplicate sample。
- 不允许缺少 idempotency runtime、delivery runtime、runtime event schema、writer runtime、durable backend、operator approval runtime、credential handle runtime、backend health runtime 或 no leakage runtime 时创建 audit success。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / refresh fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 backend health check、不执行 approval runtime、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不创建 runtime event schema、不创建 audit store、不创建 audit writer runtime、不写 audit event、不创建 writer result、不执行 delivery、不持久化 delivery result、不创建 idempotency runtime、不创建 idempotency key store、不执行 duplicate detection、不创建 duplicate detector、不执行 replay decision、不创建 replay executor、不创建 retry executor、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-idempotency-runtime-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-idempotency-runtime-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py`

不得新增或启用 audit store runtime implementation task card、idempotency runtime implementation task card、delivery runtime implementation task card、audit store runtime、audit idempotency runtime、idempotency key store、duplicate detector runtime、replay executor、retry executor、audit delivery runtime、audit writer runtime、audit writer、audit event writer、idempotency result fixture、duplicate decision fixture、runtime event schema artifact、production resolver runtime、cloud secret SDK / client、secret value fixture、production credential file、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、database connection provider、DB driver、SQL migration、schema marker reader / writer、migration runner、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 audit store runtime task card 仍不能创建。后续若继续 audit store runtime 前置，应重新评审 audit store runtime implementation entry，确认 durable backend、writer runtime、runtime event schema、delivery runtime、idempotency runtime、operator approval、credential handle、backend health 和 no leakage runtime 依赖的状态，而不得直接创建 audit store runtime implementation task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-idempotency-runtime-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-delivery-runtime-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
