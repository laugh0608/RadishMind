# Production Secret Backend Credential Handle Runtime Boundary Readiness v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime implementation task card 之前的 credential handle runtime boundary。

对应切片：`production-secret-backend-credential-handle-runtime-boundary-readiness-v1`。

结论：状态为 `credential_handle_runtime_boundary_readiness_defined`。本批只定义 future opaque credential handle 的 reference 边界、允许 metadata、禁止 payload / secret material、secret ref / provider profile / environment binding 前置、operator approval / audit / rotation 依赖、handle lifecycle 状态、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 credential handle runtime，不创建 credential payload，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不启用 repository mode，不创建 no secret leakage smoke runtime，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定真实 resolver runtime implementation task card 仍 blocked。
- `production-secret-backend-real-resolver-runtime-preconditions-v1` 已要求真实 resolver implementation 前必须具备 credential handle 边界。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile static selection，但不创建 backend runtime 或 health check。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 no leakage smoke runtime strategy，但不创建 smoke runtime。
- `production-secret-backend-fake-resolver-runtime-implementation-v1` 只实现 test-only fake resolver runtime，不能替代 production credential handle boundary。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 仍确认 production resolver interface 默认 disabled。

## Opaque Reference Boundary

future credential handle 只能是不可逆、不可推导、不可解码为 credential 的 opaque reference。它不是 secret value，不是 credential payload，不是 provider token，不是 DSN，不是云 secret backend URL，也不能包含可从字符串中还原 secret material 的 path、query、claim、header 或结构化 payload。

允许作为 future success metadata 暴露的字段只限：

| 字段 | 要求 |
| --- | --- |
| `credential_handle_id` | opaque、不可逆、不可从 secret ref / provider profile / user claim 推导 |
| `credential_handle_kind` | 只允许 `opaque_reference` 或 `rotation_pending_reference` 等非 payload 类型 |
| `credential_handle_status` | 只描述 lifecycle，不表示 credential value 已暴露 |
| `credential_handle_lifecycle_state` | 必须来自固定 lifecycle allowlist |
| `environment` | 必须与 secret ref、provider profile、backend profile 绑定一致 |
| `provider` / `provider_profile` | 只允许稳定短键或 profile id |
| `secret_ref_key` | 只允许 reference key，不允许 full secret ref value |
| `secret_ref_version_ref` | 只允许版本引用，不允许 secret value 或 backend URL |
| `backend_profile_ref` | 只允许静态 profile reference |
| `operator_approval_ref` | 只允许 approval evidence reference |
| `audit_ref` | 只允许审计 reference，不写 audit store |
| `policy_version` / `rotation_policy_version` | 只允许策略版本 |
| `created_at` / `expires_at` | 只允许生命周期 metadata，不承载 credential |
| `failure_code` / `sanitized_diagnostic` | 只允许 fail-closed 脱敏诊断 |

禁止任何 future handle 或 fixture 输出：

- credential payload
- secret value
- password、token、API key、cloud credential
- provider raw URL、resolver backend URL、DSN、database hostname
- full secret ref value
- full credential handle payload
- authorization header、cookie、完整用户 claim
- 可解码、可反序列化或可拼接成 credential 的结构化 blob

## Binding Preconditions

future credential handle runtime 必须在生成任何 handle metadata 前满足以下绑定前置：

| binding | 要求 |
| --- | --- |
| secret ref binding | `secret_ref_key`、`secret_ref_version_ref` 与 reference-only secret schema 对齐 |
| provider profile binding | provider / provider profile 必须与 provider profile secret binding evidence 对齐 |
| environment binding | test / production / local 不允许互相 fallback |
| backend profile binding | 必须引用 resolver backend profile selection readiness |
| operator approval dependency | production handle issuance 必须依赖 operator approval evidence |
| audit dependency | 必须引用 audit policy / future audit store handoff，不在本批写 store |
| rotation dependency | 必须引用 rotation policy，过期 / rotation pending 时 fail closed |
| no leakage dependency | 必须引用 no leakage smoke runtime strategy，runtime 未创建时不能声明 production ready |

## Handle Lifecycle

本批只固定 future lifecycle 名称，不创建 runtime 状态机：

| lifecycle state | 含义 |
| --- | --- |
| `reference_planned` | 仅完成 reference 边界规划 |
| `metadata_bound` | secret ref / provider profile / environment metadata 已绑定 |
| `issuance_blocked_pending_operator_approval` | 等待 operator approval evidence |
| `future_issued_metadata_only` | future runtime 只返回 opaque metadata，不返回 payload |
| `rotation_pending_rebind` | rotation 后需要重新绑定版本引用 |
| `revoked` | handle 已撤销，后续使用必须 fail closed |
| `expired` | handle 已过期，禁止 fallback |
| `resolution_failed_closed` | 解析失败，只返回 sanitized failure |

任何状态都不能暴露 credential payload 或 secret material。

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| opaque reference definition | `defined_static_only` | 只定义 opaque reference，不创建 runtime |
| metadata allowlist | `fixed_sanitized_only` | 只允许 metadata 字段 |
| payload and secret material | `forbidden` | payload / secret material 一律禁止 |
| secret ref binding | `required_before_runtime` | 未绑定时 fail closed |
| provider profile binding | `required_before_runtime` | 未绑定时 fail closed |
| environment binding | `required_no_cross_environment` | 禁止跨环境 fallback |
| operator approval dependency | `required_before_runtime` | production handle issuance 依赖人工批准 |
| audit dependency | `required_before_runtime` | 只引用 audit handoff，不写 store |
| rotation dependency | `required_before_runtime` | rotation / expiry 必须 fail closed |
| lifecycle state allowlist | `defined_static_only` | 只定义 future lifecycle |
| credential handle runtime | `not_created` | 本批不创建 runtime |
| credential payload runtime | `forbidden` | 本批不创建 payload |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| DB / repository / API | `blocked` | 不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `credential_handle_boundary_missing` | `credential_boundary` | 缺少 credential handle runtime boundary evidence |
| `credential_handle_opaque_reference_missing` | `opaque_reference` | 未定义 opaque reference |
| `credential_handle_metadata_allowlist_missing` | `metadata_contract` | 未定义允许 metadata |
| `credential_handle_payload_detected` | `artifact_guard` | 文档、fixture 或 future output 出现 credential payload |
| `credential_handle_secret_material_detected` | `artifact_guard` | 出现 secret value、token、DSN 或 cloud credential |
| `credential_handle_secret_ref_binding_missing` | `secret_ref` | 缺少 secret ref key / version reference |
| `credential_handle_provider_profile_binding_missing` | `provider_profile` | 缺少 provider profile binding |
| `credential_handle_environment_binding_mismatch` | `environment_binding` | handle metadata 与 environment binding 不一致 |
| `credential_handle_operator_approval_missing` | `operator_gate` | production handle issuance 缺少 operator approval evidence |
| `credential_handle_audit_dependency_missing` | `audit_policy` | 缺少 audit dependency 或 audit handoff reference |
| `credential_handle_rotation_dependency_missing` | `rotation_policy` | 缺少 rotation policy / version dependency |
| `credential_handle_lifecycle_state_invalid` | `lifecycle` | lifecycle state 不在 allowlist |
| `credential_handle_failure_mapping_missing` | `failure_mapping` | 缺少 fail-closed failure mapping |
| `credential_handle_diagnostic_exposure_detected` | `diagnostics` | diagnostics 暴露 forbidden fields |
| `credential_handle_runtime_created_forbidden` | `artifact_guard` | 本批创建 credential handle runtime |
| `credential_handle_side_effect_forbidden` | `no_side_effects` | 本批读取 secret、调用云服务、连接 DB 或写 audit store |
| `credential_handle_fallback_forbidden` | `no_fallback` | 缺失 handle 时 fallback 到 payload、fake resolver、env 或 sample |
| `credential_handle_scope_overreach` | `implementation_boundary` | 本批合入 production resolver runtime、DB provider、audit store 或 public API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、完整 secret ref value、完整 credential handle、完整用户 claim、authorization header 或 cookie。

## Sanitized Diagnostics

允许输出：

- `credential_handle_boundary_status`
- `credential_handle_runtime_status`
- `credential_handle_metadata_status`
- `credential_handle_lifecycle_state`
- `credential_handle_kind`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_key_status`
- `secret_ref_version_ref_status`
- `backend_profile_ref_status`
- `operator_approval_status`
- `audit_policy_status`
- `rotation_policy_status`
- `failure_code`
- `failure_boundary`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## No Fallback

- credential handle boundary 缺失时真实 resolver runtime implementation task card 必须继续 blocked。
- 不允许 credential handle fallback 到 credential payload、secret value、fake resolver runtime、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许 production handle fallback 到 test handle，也不允许 test handle fallback 到 production handle。
- 不允许 secret ref / provider profile / environment binding 缺失时通过 `RADISHMIND_PLATFORM_API_KEY` 或其它环境变量 credential 生成 handle success。
- 不把 credential handle boundary readiness 写成 credential handle runtime created、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
- `credential_handle_runtime_created_count=0`
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

- `docs/platform/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.md`
- `docs/task-cards/production-secret-backend-credential-handle-runtime-boundary-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-credential-handle-runtime-boundary-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py`

不得新增或启用以下 artifact：

- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- no secret leakage smoke runtime
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

本批完成后，真实 resolver runtime implementation task card 仍不能创建。`credential-handle-runtime-boundary-readiness` 只固定 future opaque credential handle 的边界和停止线，不创建 handle runtime。后续已由 `production-secret-backend-operator-approval-runtime-evidence-readiness-v1` 固定 operator approval runtime evidence boundary，并由 `production-secret-backend-audit-store-handoff-readiness-v1` 固定 audit store handoff readiness。剩余还需单独推进：

1. `resolver-backend-health-boundary-readiness`

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
