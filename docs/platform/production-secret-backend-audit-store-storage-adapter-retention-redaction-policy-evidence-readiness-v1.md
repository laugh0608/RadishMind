# Production Secret Backend Audit Store Storage Adapter Retention / Redaction Policy Evidence Readiness v1

更新时间：2026-06-30

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Append-Only Semantics Evidence Readiness v1` 之后，固定 future storage adapter runtime task card 前必须具备的 retention / redaction policy evidence。

对应切片：`production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined`，readiness decision 为 `retention_redaction_policy_evidence_defined_without_runtime`。本批只定义 metadata-only retention window reference、metadata-only redaction policy reference、append-only immutability compatibility、forbidden erase / overwrite policy、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 storage adapter runtime task card、storage adapter runtime、retention executor、redaction executor、DB provider、SQL、audit store runtime、writer runtime、delivery runtime、idempotency runtime、repository mode 或 public production API。

## 输入证据

- `audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined` 已固定 append-only insert-only 操作、禁止 mutation、sequence reference、record immutability、duplicate / replay fail-closed policy 和 writer append compatibility。
- `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined` 已固定 metadata-only input / result envelope、record identity、failure taxonomy 和 writer compatibility，但未创建实际 contract artifact。
- `audit_store_storage_adapter_backend_product_evidence_readiness_defined` 已确认 backend product evidence readiness 已定义，但 backend product selection 仍为 `not_selected`。
- `audit_store_storage_adapter_runtime_implementation_entry_review_defined` 已确认 storage adapter runtime task card 仍 blocked before backend product evidence。
- `audit_store_runtime_blocker_matrix_defined` 已提供 durable backend blocker 与依赖顺序；本批把 durable backend blocker 推进为 retention / redaction policy evidence readiness 已定义但 task card 仍 blocked。
- `implementation_readiness_defined` 当前仍保持 production secret backend、audit store runtime、storage adapter runtime、writer / delivery / idempotency runtime、repository mode 与 production API 均未创建或未满足。

## Readiness Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| retention / redaction policy evidence | `audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined` | 只定义 policy evidence |
| readiness decision | `retention_redaction_policy_evidence_defined_without_runtime` | 不创建 runtime、executor 或 adapter task card |
| retention window | `metadata_only_retention_window_reference_defined` | 只允许 retention policy reference，不定义物理 TTL、partition 或 deletion job |
| redaction reference | `metadata_only_redaction_policy_reference_defined` | 只允许 redaction policy reference，不写 raw payload 或 payload hash |
| append-only compatibility | `append_only_immutability_compatible_policy_defined` | retention / redaction 不得改写既有 record identity、sequence 或 payload |
| erase / overwrite policy | `delete_overwrite_inline_redaction_forbidden` | delete / overwrite / inline redact 不得进入 adapter 成功路径 |
| offline validation | `not_created` | 不创建 runtime smoke、DB smoke 或 adapter validation runner |
| negative leakage scan | `not_created` | 不创建 scanner 或 committed scan output |
| rollback / recovery | `required_before_runtime_task_card` | 仍需后续独立证据 |
| next dependency | `storage_adapter_offline_validation_evidence_readiness` | 下一项仍是静态证据，不是 runtime |
| storage adapter runtime | `not_created` | 不创建 runtime、client、driver、executor 或 writer |

## Retention Policy Evidence

retention evidence 只能是 metadata-only reference：

- `retention_policy_ref`
- `retention_window_ref`
- `retention_anchor_ref`
- `retention_decision_ref`
- `append_only_contract_ref`
- `storage_record_identity_ref`
- `policy_version`
- `audit_ref`

本批不定义具体 TTL、cron、job、物理 partition、bucket lifecycle、topic retention、table purge、queue cleanup、log compaction、object deletion、database deletion 或 provider rule。后续 runtime 若要实现 retention，必须独立证明：

- retention 不通过 delete / overwrite / truncate / compact existing record 达成；
- retention 不重写 append-only sequence、record identity 或 writer result reference；
- retention diagnostic 不暴露 backend physical location、DSN、endpoint、topic、bucket、queue、table、partition 或 credential；
- retention failure 必须 fail closed，不能被解释为 audit write success 或 idempotent success。

## Redaction Policy Evidence

redaction evidence 只能是 metadata-only reference：

- `redaction_policy_ref`
- `redaction_decision_ref`
- `redaction_reason_ref`
- `redaction_actor_ref`
- `append_only_contract_ref`
- `storage_record_identity_ref`
- `policy_version`
- `audit_ref`

本批不执行 redaction，不写 redaction marker，不读取、输出或派生 raw payload，不提交 payload hash、secret-derived hash 或 provider-specific identifier。后续 runtime 若要实现 redaction，必须独立证明：

- redaction 不通过 overwrite existing record、inline redact payload、delete payload 或 mutate identity 达成；
- redaction 不要求 storage adapter 读取 raw event payload、raw writer payload、credential payload 或 secret material；
- redaction result 只能以 reference 表示，不暴露原始值、hash、provider error detail 或 backend error detail；
- redaction failure 必须 fail closed，不能回退到 unredacted output、sample、memory store 或 fake resolver。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_retention_redaction_dependency_missing` | `dependency_chain` | 缺少 append-only semantics evidence、metadata contract artifact readiness、backend product evidence readiness、storage adapter entry review、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_retention_window_reference_missing` | `retention_policy` | 缺少 metadata-only retention policy / window / anchor reference |
| `audit_store_storage_adapter_redaction_reference_missing` | `redaction_policy` | 缺少 metadata-only redaction policy / decision / actor reference |
| `audit_store_storage_adapter_retention_redaction_mutation_forbidden` | `append_only_compatibility` | retention / redaction 被声明为 delete、overwrite、truncate、compact、inline redact 或 mutate identity |
| `audit_store_storage_adapter_retention_redaction_payload_material_detected` | `artifact_guard` | 文档、fixture 或 diagnostics 出现 raw payload、payload hash、secret material、credential payload、DSN 或 provider detail |
| `audit_store_storage_adapter_retention_redaction_runtime_scope_overreach` | `implementation_boundary` | 本批打开 runtime、executor、DB provider、SQL、repository mode 或 public API scope |
| `audit_store_storage_adapter_retention_redaction_fallback_detected` | `no_fallback` | 使用 sample、memory store、fake resolver、historical smoke 或 static fixture 替代 policy evidence |
| `audit_store_storage_adapter_retention_redaction_next_dependency_missing` | `dependency_order` | 未把下一项独立依赖固定为 offline validation evidence |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_retention_redaction_policy_evidence_status`
- `readiness_decision`
- `retention_window_reference_status`
- `redaction_policy_reference_status`
- `append_only_compatibility_status`
- `forbidden_erasure_policy_status`
- `offline_validation_status`
- `negative_leakage_scan_status`
- `rollback_recovery_status`
- `next_dependency`
- `storage_adapter_runtime_status`
- `audit_store_runtime_status`
- `production_resolver_runtime_status`
- `repository_mode_status`
- `production_api_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、product endpoint、bucket name、queue name、topic name、region-specific resource id、database name、table name、partition key、provider resource id、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw event payload、raw writer payload、raw storage payload、raw retained payload、raw redacted payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 append-only semantics evidence、metadata contract artifact readiness、backend product evidence readiness、runtime event schema artifact、memory store、fake resolver、static fixture、sample、mock provider、repository memory store、audit memory store、static writer boundary、static delivery / idempotency readiness 或 historical smoke 替代 retention / redaction policy evidence readiness。
- 不允许把 retention / redaction policy evidence readiness 写成 backend product selected、DB provider ready、storage adapter runtime ready、retention runtime ready、redaction runtime ready、audit store ready、writer runtime ready、delivery runtime ready、idempotency runtime ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不创建 contract schema 文件、不创建 storage adapter runtime、不创建 retention / redaction executor、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不执行 idempotency decision、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py`

不得新增或启用 `contracts/production-secret-audit-storage-adapter.metadata-contract.json`、backend product selection artifact、storage adapter runtime implementation task card、storage adapter runtime、retention executor、redaction executor、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、schema marker reader / writer、audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer / runner、delivery runtime、idempotency runtime、duplicate detector、retry executor、replay executor、production resolver runtime、cloud secret SDK / client、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先独立推进 `storage_adapter_offline_validation_evidence_readiness`，证明已定义的 metadata contract、append-only semantics 与 retention / redaction policy 可以离线验证且不会触碰真实 backend；不直接创建 storage adapter runtime、DB provider、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-retention-redaction-policy-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-append-only-semantics-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
