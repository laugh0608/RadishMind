# Production Secret Audit Storage Adapter Metadata Contract 契约

更新时间：2026-07-08

## 定位

`contracts/production-secret-audit-storage-adapter.metadata-contract.json` 是 future production secret backend audit store storage adapter 的 metadata-only contract artifact。它只固定未来 storage adapter 可消费的 input envelope、result envelope、record identity、failure taxonomy、writer compatibility 和 forbidden field policy；artifact 本身不选择具体 backend product，不创建 storage adapter runtime、DB provider、audit store runtime、repository mode 或 production API。

对应实现切片为 `production-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1`，状态为 `audit_store_storage_adapter_metadata_contract_artifact_materialized`。

## 字段边界

- `contract_version` 固定为 `audit-storage-adapter-metadata-contract-v1`。
- `input_envelope` 只允许 reference / status / policy version 字段，覆盖 audit event、writer result、idempotency、delivery attempt、append-only contract、retention / redaction policy、request 和 audit refs。
- `result_envelope` 只允许 storage adapter result ref、record ref、record identity ref、write status、append-only sequence ref、dedupe / delivery refs、failure metadata、audit ref 和 policy version。
- `record_identity` 只描述 metadata-only stable identity，不是数据库主键、provider resource id、bucket key、queue topic 或 table / partition 设计。
- `writer_compatibility` 只证明 future writer output 的 metadata refs 可被 storage adapter input envelope 消费，不创建 writer runtime 或 storage adapter runtime。

## 正负例

| 目标 | fixture | 预期 |
| --- | --- | --- |
| positive contract candidate | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-positive-v1.json` | 通过 contract checker |
| missing required field | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-missing-required-negative-v1.json` | 被拒绝 |
| forbidden field | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-forbidden-field-negative-v1.json` | 被拒绝 |
| additional properties | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-additional-properties-negative-v1.json` | 被拒绝 |
| writer compatibility | `scripts/checks/fixtures/production-secret-audit-storage-adapter-metadata-contract-writer-compatibility-v1.json` | 验证 writer output projection 可被 input envelope 消费 |

## 禁止字段

contract artifact 和 checker 显式拒绝 secret value、raw secret、password、token、API key、authorization header、cookie、provider raw URL、resolver backend URL、DSN、database hostname、database name、table name、partition key、bucket name、queue name、topic name、product endpoint、provider resource id、cloud credential、credential payload、full secret ref、full credential handle、raw request / response / audit / event / writer / storage payload、payload hash、secret-derived hash、provider error detail、database error detail、scanner raw finding、scan output、recovery raw finding 和 recovery output。

## 当前下游消费边界

- 该 artifact 已被后续 storage adapter backend product selection、metadata-only logical table schema artifact、database / provider / driver selection、connection lifecycle readiness、connection runtime boundary、managed product profile review、reference-only provider profile review、provider account / resource / endpoint readiness 和 provider account / resource / endpoint review 消费。
- 当前只静态选择 product class `managed_database_append_only_table`、PostgreSQL-compatible 能力族、managed PostgreSQL-compatible provider candidate class、reference-only driver candidate `github.com/jackc/pgx/v5`、reference-only product profile `managed_postgresql_compatible_audit_store_profile` 与 reference-only provider profile `managed_postgresql_compatible_provider_reference`。
- 后续 readiness / review 只定义 secret-ref-only DSN handoff、TLS / role / environment binding、logical table schema boundary、connection lifecycle boundary、provider account / resource / endpoint metadata-only evidence、operator confirmation gate、rejection conditions 和 sanitized diagnostics；它不打开 driver、不解析 DSN、不创建 connection provider、DB provider、SQL、DDL、schema marker runtime、migration runner 或 storage adapter runtime。

## 消费规则

- future storage adapter runtime 只能把该 artifact 当作 metadata-only contract input，不能把 fixture 当作真实 audit record、storage record 或 runtime output。
- backend product selection 已通过独立 review 消费该 contract artifact、append-only / retention / validation / leakage / recovery 证据；当前结论只允许 static product class selection，不允许具体数据库 / vendor selection。
- `audit_event_ref`、`writer_result_ref`、`storage_record_identity_ref`、`append_only_sequence_ref`、`idempotency_key_ref` 和 `delivery_commit_ref` 都是 opaque reference，不是可解引用的生产 secret、数据库键或 provider resource id。
- 后续如果扩字段或 failure code，必须同步更新 contract artifact、positive / negative fixtures、checker、platform 专题和 blocker matrix。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-storage-adapter-metadata-contract-artifact-materialization-v1.py
./scripts/check-repo.sh --fast
```

## 停止线

本契约 artifact 本身不创建 storage adapter runtime、不创建 storage adapter client、不选择具体 backend product / vendor、不创建 DB provider、不连接数据库、不创建 SQL migration、不启用 repository mode、不创建 audit writer / delivery / idempotency runtime、不创建 production resolver runtime，也不声明 production secret backend ready。
