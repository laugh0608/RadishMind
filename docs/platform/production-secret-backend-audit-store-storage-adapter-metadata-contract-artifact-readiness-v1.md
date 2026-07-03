# Production Secret Backend Audit Store Storage Adapter Metadata Contract Artifact Readiness v1

更新时间：2026-06-30

## 文档目的

本文档在 `Production Secret Backend Audit Store Storage Adapter Backend Product Evidence Readiness v1` 之后，固定 future storage adapter runtime task card 前必须具备的 metadata-only adapter contract artifact readiness。

对应切片：`production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1`。

结论：状态为 `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined`，readiness decision 为 `metadata_contract_artifact_readiness_defined_without_materialized_artifact`。本批只定义 metadata contract artifact 的 reserved path、input envelope、result envelope、record identity、failure taxonomy、writer compatibility、append-only / retention / redaction reference、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建实际 contract schema 文件，不创建 storage adapter runtime task card、storage adapter runtime、DB provider、SQL、audit store runtime、repository mode 或 public production API。

## 输入证据

- `audit_store_storage_adapter_backend_product_evidence_readiness_defined` 已确认 backend product evidence readiness 已定义，但 backend product selection 仍为 `not_selected`。
- `audit_store_storage_adapter_runtime_implementation_entry_review_defined` 已确认 storage adapter runtime task card 仍 blocked before backend product evidence。
- `audit_store_concrete_durable_backend_selection_review_defined` 已把 durable backend family 静态选择为 `append_only_metadata_audit_log` / `reserved_append_only_audit_log`。
- `audit_store_runtime_blocker_matrix_defined` 已把 durable backend blocker 推进为 backend product evidence readiness 已定义但 task card 仍 blocked。
- `audit_store_runtime_event_schema_artifact_implemented` 已提供 metadata-only audit event schema artifact，但不代表 adapter contract artifact 或 runtime 已创建。
- `implementation_readiness_defined` 当前仍保持 production secret backend、audit store runtime、storage adapter runtime、writer / delivery / idempotency runtime、repository mode 与 production API 均未创建或未满足。

## Readiness Decision

| gate | 本次结论 | 说明 |
| --- | --- | --- |
| metadata contract artifact readiness | `audit_store_storage_adapter_metadata_contract_artifact_readiness_defined` | 只定义 contract artifact readiness |
| readiness decision | `metadata_contract_artifact_readiness_defined_without_materialized_artifact` | 不创建实际 schema / contract 文件 |
| contract artifact path | `reserved_static_path` | 只保留 future path，不写文件 |
| input envelope | `metadata_only_input_envelope_defined` | 只允许 reference 与审计 metadata |
| result envelope | `metadata_only_result_envelope_defined` | 只允许 record reference 与 status metadata |
| record identity | `metadata_only_record_identity_defined` | 只定义 stable record identity shape |
| failure taxonomy | `metadata_only_failure_taxonomy_defined` | 只定义 fail-closed taxonomy |
| writer compatibility | `metadata_only_writer_compatibility_defined` | 只定义 future writer 消费边界 |
| append-only reference | `static_reference_defined_without_semantics_evidence` | 仍需后续 append-only semantics evidence |
| retention / redaction reference | `static_reference_defined_without_policy_evidence` | 仍需后续 retention / redaction evidence |
| contract artifact materialization | `not_created` | 不新增 schema / contract artifact 文件 |
| storage adapter runtime | `not_created` | 不创建 runtime、client、driver 或 writer |
| DB provider / SQL | `blocked / not_created` | 不创建 driver、DSN、SQL、schema marker 或 migration |

## Reserved Contract Artifact

后续若进入 metadata contract artifact materialization task card，默认 reserved path 为：

- `contracts/production-secret-audit-storage-adapter.metadata-contract.json`

本批只保留该 path 作为 future artifact reference，不创建文件。后续 materialization 前必须先确认：

- path 命名稳定且不含 provider / profile / product / environment 长语义；
- contract artifact 只描述 metadata envelope、record identity、writer compatibility 与 failure taxonomy；
- contract artifact 不包含 endpoint、DSN、hostname、bucket、queue、topic、region resource id、credential、secret、token、provider raw URL 或 database detail；
- future checker 必须覆盖 positive fixture、forbidden field negative fixture、writer compatibility smoke 和 no secret material scan。

## Metadata Contract Shape

### Input Envelope

允许的 input metadata 字段：

- `audit_event_ref`
- `audit_event_schema_version`
- `audit_event_kind`
- `storage_adapter_contract_version`
- `storage_adapter_contract_ref`
- `backend_product_evidence_ref`
- `backend_product_class`
- `storage_record_identity_ref`
- `writer_result_ref`
- `idempotency_key_ref`
- `delivery_attempt_ref`
- `append_only_contract_ref`
- `retention_policy_ref`
- `redaction_policy_ref`
- `request_id`
- `audit_ref`
- `policy_version`

### Result Envelope

允许的 result metadata 字段：

- `storage_adapter_result_ref`
- `storage_record_ref`
- `storage_record_identity_ref`
- `write_status`
- `append_only_sequence_ref`
- `dedupe_decision_ref`
- `delivery_commit_ref`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `audit_ref`
- `policy_version`

### Record Identity

record identity 只允许由 reference 组成：

- `storage_record_identity_ref`
- `audit_event_ref`
- `append_only_sequence_ref`
- `idempotency_key_ref`
- `writer_result_ref`
- `policy_version`

禁止把 physical primary key、table name、partition key、bucket key、queue offset、topic partition、hostname、database name、DSN、provider resource id 或 credential handle raw value 写进 committed contract。

### Failure Taxonomy

metadata contract artifact 必须能表达以下 failure code families：

- `dependency_missing`
- `contract_version_mismatch`
- `forbidden_field_detected`
- `record_identity_missing`
- `append_only_reference_missing`
- `retention_redaction_reference_missing`
- `writer_compatibility_missing`
- `runtime_artifact_forbidden`

### Writer Compatibility

writer compatibility 只定义 future writer result 与 storage adapter input 的 metadata handoff：

- writer 只能提交 `writer_result_ref`、`audit_event_ref`、`audit_event_schema_version`、`audit_event_kind`、`idempotency_key_ref`、`append_only_contract_ref`、`retention_policy_ref`、`redaction_policy_ref` 和 `policy_version`。
- storage adapter 只能返回 `storage_adapter_result_ref`、`storage_record_ref`、`storage_record_identity_ref`、`write_status`、`append_only_sequence_ref`、`failure_code`、`failure_boundary` 和 `sanitized_diagnostic`。
- writer compatibility 不代表 writer runtime、storage adapter runtime、audit store runtime 或 delivery runtime 已创建。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `audit_store_storage_adapter_metadata_contract_dependency_missing` | `dependency_chain` | 缺少 backend product evidence readiness、storage adapter entry review、blocker matrix、schema artifact 或 implementation readiness |
| `audit_store_storage_adapter_metadata_contract_materialization_forbidden` | `artifact_guard` | 本批创建实际 contract schema / contract artifact 文件 |
| `audit_store_storage_adapter_metadata_contract_secret_material_detected` | `artifact_guard` | 文档、fixture 或 diagnostics 出现 secret / DSN / endpoint 等敏感材料 |
| `audit_store_storage_adapter_metadata_contract_identity_missing` | `record_identity` | 缺少 metadata-only record identity shape |
| `audit_store_storage_adapter_metadata_contract_envelope_missing` | `metadata_envelope` | 缺少 input / result envelope |
| `audit_store_storage_adapter_metadata_contract_writer_compatibility_missing` | `writer_compatibility` | 缺少 writer result 到 storage adapter input 的 metadata handoff |
| `audit_store_storage_adapter_metadata_contract_backend_selection_forbidden` | `product_selection` | 本批选择具体 backend product 或 vendor service |
| `audit_store_storage_adapter_metadata_contract_scope_overreach` | `implementation_boundary` | 本批打开 runtime、repository mode 或 public API scope |

所有 failure 必须 fail closed，只返回 metadata-only diagnostics。

## Sanitized Diagnostics

允许输出：

- `audit_store_storage_adapter_metadata_contract_artifact_status`
- `readiness_decision`
- `contract_artifact_path_status`
- `input_envelope_status`
- `result_envelope_status`
- `record_identity_status`
- `failure_taxonomy_status`
- `writer_compatibility_status`
- `append_only_reference_status`
- `retention_redaction_reference_status`
- `contract_artifact_materialization_status`
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

禁止输出 raw secret、secret value、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、product endpoint、bucket name、queue name、topic name、region-specific resource id、database name、table name、partition key、provider resource id、database error detail、cloud credential、credential payload、full secret ref value、full credential handle、raw request / response payload、raw audit payload、raw event payload、raw writer payload、payload hash、secret-derived hash 或 provider error detail。

## No Fallback / No Side Effects

- 不允许把 backend product evidence readiness、storage adapter entry review、runtime event schema artifact、memory store、fake resolver、static fixture、sample、mock provider、repository memory store、audit memory store、static writer boundary、static delivery / idempotency readiness 或 historical smoke 替代 metadata contract artifact readiness。
- 不允许把 metadata contract artifact readiness 写成 contract artifact materialized、backend product selected、DB provider ready、storage adapter runtime ready、audit store ready、production resolver ready、repository mode ready、public API ready 或 production ready。
- 本批 checker 只读取 committed 文档、fixture、已有 readiness / review fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不创建 contract schema 文件、不创建 storage adapter runtime、不创建 audit store、不创建 audit writer、不写 audit event、不执行 delivery、不执行 idempotency decision、不连接数据库、不打开 driver、不运行 SQL、不启用 repository mode、不调用 production API。

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.md`
- `docs/task-cards/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py`

不得新增或启用 `contracts/production-secret-audit-storage-adapter.metadata-contract.json`、backend product selection artifact、storage adapter runtime implementation task card、storage adapter runtime、durable audit backend runtime、database connection provider、DB driver、DSN parser、SQL migration、schema marker reader / writer、audit store runtime implementation task card、audit store runtime、audit writer runtime、audit event writer / runner、delivery runtime、idempotency runtime、duplicate detector、retry executor、replay executor、production resolver runtime、cloud secret SDK / client、credential payload runtime、credential handle runtime、approval runtime、backend health runtime、no leakage smoke runtime、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批之后 storage adapter runtime task card 仍 blocked。若继续 audit store durable backend，应先独立推进 `storage_adapter_append_only_semantics_evidence_readiness`，证明 update / delete 不进入 adapter runtime 成功路径；不直接创建 storage adapter runtime、DB provider、audit store runtime task card 或 production resolver runtime task card。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-backend-product-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-runtime-blocker-matrix-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```
