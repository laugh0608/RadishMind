# Production Secret Backend Audit Store Storage Adapter Negative Leakage Scan Evidence Readiness v1

更新时间：2026-07-01

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Offline Validation Evidence Readiness v1` 之后，固定 future storage adapter runtime task card 前必须具备的 negative leakage scan evidence。

对应切片：`production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined`，readiness decision 为 `negative_leakage_scan_evidence_defined_without_runtime`。本批只定义 offline validation evidence 的负向泄露扫描证据、scan manifest reference、scan target reference、forbidden material coverage、diagnostic allowlist、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 scanner、scan runner、scan output、storage adapter runtime task card、storage adapter runtime、DB provider、audit store runtime、production resolver runtime、repository mode 或 public production API。

## 输入证据

- `audit_store_storage_adapter_offline_validation_evidence_readiness_defined` 已固定 metadata-only offline validation manifest、positive / negative case reference、coverage matrix 和 backend touch forbidden policy。
- `audit_store_storage_adapter_retention_redaction_policy_evidence_readiness_defined` 已固定 retention / redaction policy 只作为 metadata-only reference，不创建 executor。
- `audit_store_storage_adapter_append_only_semantics_evidence_readiness_defined` 已固定 append-only insert-only、禁止 mutation、sequence reference、record immutability 和 duplicate / replay fail-closed policy。
- `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined` 已固定 input / result envelope、record identity、failure taxonomy 和 writer compatibility，但实际 contract artifact 仍未物化。
- `audit_store_storage_adapter_backend_product_evidence_readiness_defined` 已确认 backend product selection 仍为 `not_selected`。
- `audit_store_runtime_blocker_matrix_defined` 仍要求 durable backend blocker 阻塞 audit store runtime task card 和 production resolver runtime task card。
- `implementation_readiness_defined` 当前仍保持 production secret backend、audit store runtime、storage adapter runtime、writer / delivery / idempotency runtime、repository mode 与 production API 均未创建或未满足。

## Readiness Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| negative leakage scan evidence | `audit_store_storage_adapter_negative_leakage_scan_evidence_readiness_defined` | 只定义负向泄露扫描证据 |
| readiness decision | `negative_leakage_scan_evidence_defined_without_runtime` | 不创建 scanner、runner、runtime、DB provider 或 adapter task card |
| scan manifest | `metadata_only_negative_leakage_scan_manifest_reference_defined` | 只定义 manifest reference，不提交 scan output |
| scan target | `metadata_only_scan_target_reference_defined` | 只引用 offline validation evidence 的 metadata-only target |
| forbidden material coverage | `raw_payload_secret_credential_provider_backend_detail_coverage_defined` | 覆盖 raw payload、secret / credential material、provider detail 和 backend detail |
| diagnostic allowlist | `metadata_only_diagnostic_allowlist_defined` | diagnostics 只允许状态、failure code、audit ref 和 policy version |
| scan runner | `not_created` | 不创建 scanner、CLI、runtime smoke 或 committed output |
| rollback / recovery | `required_before_runtime_task_card` | 仍需后续独立证据 |
| next dependency | `storage_adapter_rollback_recovery_evidence_readiness` | 下一项仍是静态证据，不是 runtime |
| storage adapter runtime | `not_created` | 不创建 runtime、client、driver、executor 或 writer |

## Negative Leakage Scan Evidence

negative leakage scan evidence 只能由 metadata-only references 组成：

- `negative_leakage_scan_manifest_ref`
- `scan_target_manifest_ref`
- `offline_validation_evidence_ref`
- `forbidden_material_matrix_ref`
- `diagnostic_allowlist_ref`
- `coverage_matrix_ref`
- `failure_taxonomy_ref`
- `policy_version`
- `audit_ref`

本批不生成可执行 scanner，不写 committed scan output，不读取 raw request / response / audit / writer / storage payload，不派生 payload hash 或 secret-derived hash，不读取 provider raw URL、DSN、endpoint、bucket、queue、topic、table、partition、region resource id 或 provider resource id。后续 scanner 若要实现，必须独立证明：

- 扫描 target 只能来自 metadata-only manifest reference；
- raw payload、secret material、credential material、provider detail 和 backend detail 都只能作为 forbidden class 出现在 policy 中；
- diagnostics 不暴露 raw material、payload hash、secret-derived hash、provider resource id、backend physical location、DSN、endpoint、bucket、queue、topic、table、partition 或 provider error detail；
- 缺少任一 dependency、manifest、coverage row 或 diagnostic allowlist 时必须 fail closed；
- scanner、scan runner、scan output 和 runtime task card 必须由后续独立任务卡评审。

## Coverage Matrix

| coverage item | required evidence | 禁止解释 |
| --- | --- | --- |
| raw payload material | `forbidden_material_matrix_ref` + raw payload negative cases | 不代表 payload reader 或 scanner runtime 已存在 |
| secret / credential material | `forbidden_material_matrix_ref` + secret / credential negative cases | 不代表可以读取 secret 或 credential payload |
| provider / backend detail | `forbidden_material_matrix_ref` + provider / backend negative cases | 不代表 provider probe、DB probe 或 backend touch 已执行 |
| offline validation evidence reference | `offline_validation_evidence_ref` + `scan_target_manifest_ref` | 不代表 offline validation runner 已创建或执行 |
| diagnostic allowlist | `diagnostic_allowlist_ref` + failure taxonomy reference | 不代表 runtime diagnostic emitter 已实现 |
| artifact guard | `coverage_matrix_ref` + no side effects counters | 不代表 scanner、scan output、DB provider 或 repository mode 已启用 |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_negative_leakage_scan_dependency_missing` | `dependency_chain` | 缺少 offline validation evidence、retention / redaction policy evidence、append-only semantics evidence、metadata contract artifact readiness、backend product evidence readiness、storage adapter entry review、blocker matrix 或 implementation readiness |
| `audit_store_storage_adapter_negative_leakage_scan_manifest_reference_missing` | `scan_manifest` | 缺少 metadata-only negative leakage scan manifest reference |
| `audit_store_storage_adapter_negative_leakage_scan_target_reference_missing` | `scan_target` | 缺少 metadata-only scan target reference |
| `audit_store_storage_adapter_negative_leakage_forbidden_material_coverage_missing` | `coverage_matrix` | 未覆盖 raw payload、secret / credential material、provider detail 和 backend detail |
| `audit_store_storage_adapter_negative_leakage_diagnostic_allowlist_missing` | `diagnostic_allowlist` | 缺少 metadata-only diagnostic allowlist |
| `audit_store_storage_adapter_negative_leakage_raw_material_detected` | `artifact_guard` | 证据包含 raw payload、secret、credential、payload hash 或 provider / backend detail |
| `audit_store_storage_adapter_negative_leakage_scanner_runtime_scope_overreach` | `implementation_boundary` | 本批打开 scanner、runner、runtime、DB provider、SQL、repository mode 或 public API scope |
| `audit_store_storage_adapter_negative_leakage_fallback_detected` | `no_fallback` | 使用 sample、memory store、fake resolver、historical smoke 或 previous checker success 替代负向泄露扫描证据 |
| `audit_store_storage_adapter_negative_leakage_next_dependency_missing` | `dependency_order` | 未把下一项独立依赖固定为 rollback / recovery evidence |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_negative_leakage_scan_evidence_status`
- `readiness_decision`
- `negative_leakage_scan_manifest_status`
- `scan_target_reference_status`
- `forbidden_material_coverage_status`
- `diagnostic_allowlist_status`
- `offline_validation_evidence_status`
- `scan_runner_status`
- `scan_output_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、product endpoint、bucket name、queue name、topic name、region-specific resource id、database name、table name、partition key、provider resource id、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw event payload、raw writer payload、raw storage payload、raw retained payload、raw redacted payload、payload hash、secret-derived hash、provider error detail、scanner raw finding、scan output 或 validation runner output。

## No Fallback / No Side Effects

- 不允许把 offline validation evidence、retention / redaction policy evidence、append-only semantics evidence、metadata contract artifact readiness、backend product evidence readiness、runtime event schema artifact、memory store、fake resolver、static fixture、sample、mock provider、repository memory store、audit memory store、historical smoke 或 previous checker success 替代 negative leakage scan evidence readiness。
- 不允许把 negative leakage scan evidence readiness 写成 scanner created、scan executed、scan output committed、backend product selected、contract artifact materialized、DB provider ready、storage adapter runtime ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不读取 raw payload、不创建 scanner、不创建 scan output、不创建 storage adapter runtime、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不执行 idempotency decision、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py`

不得新增或启用 `contracts/production-secret-audit-storage-adapter.negative-leakage-scan.json`、negative leakage scanner、scan CLI、committed scan output、storage adapter runtime implementation task card、storage adapter runtime、database connection provider、DB driver、DSN parser、SQL migration、schema marker reader / writer、audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer / runner、delivery runtime、idempotency runtime、duplicate detector、retry executor、replay executor、production resolver runtime、cloud secret SDK / client、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先独立推进 `storage_adapter_rollback_recovery_evidence_readiness`，证明 future storage adapter runtime 的 rollback / recovery 证据不会破坏 append-only、retention / redaction、offline validation 和 no leakage 边界；不直接创建 storage adapter runtime、DB provider、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-negative-leakage-scan-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-offline-validation-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
