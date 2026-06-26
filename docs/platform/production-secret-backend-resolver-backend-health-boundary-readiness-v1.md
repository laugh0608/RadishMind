# Production Secret Backend Resolver Backend Health Boundary Readiness v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime implementation task card 之前的 resolver backend health boundary。

对应切片：`production-secret-backend-resolver-backend-health-boundary-readiness-v1`。

结论：状态为 `resolver_backend_health_boundary_readiness_defined`。本批只定义 future resolver backend health boundary 的 reference-only 输入、profile binding、environment binding、允许 health metadata、禁止 secret material / credential payload / provider raw URL / DSN、health lifecycle、failure mapping、sanitized diagnostics、operator / audit / rotation / no leakage / credential handle 依赖、no fallback、no side effects、artifact guard、entry review alignment 和后续 runtime implementation 拆分；不创建 backend health runtime，不执行 backend health check，不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，不创建 credential handle runtime，不执行 approval runtime，不创建 audit store / writer / event，不启用 repository mode，也不新增 public production API。

## 输入证据

- `production-secret-backend-real-resolver-runtime-implementation-entry-review-v1` 已固定真实 resolver runtime implementation task card 不在当前切片创建。
- `production-secret-backend-resolver-backend-profile-selection-readiness-v1` 已固定 backend profile selection，并要求存在 backend health reference。
- `production-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1` 已固定 no leakage smoke runtime strategy，但不创建 smoke runtime。
- `production-secret-backend-credential-handle-runtime-boundary-readiness-v1` 已固定 credential handle boundary，但不创建 credential handle runtime。
- `production-secret-backend-operator-approval-runtime-evidence-readiness-v1` 已固定 operator approval runtime evidence boundary，但不创建或执行 approval runtime。
- `production-secret-backend-audit-store-handoff-readiness-v1` 已固定 audit store handoff boundary，但不创建 audit store / writer，不写 audit event。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 与 `production-secret-backend-rotation-audit-policy-readiness-v1` 只提供策略和人工操作前置，不提供 runtime health probe。

## Health Boundary Contract

future resolver backend health boundary 只能是 metadata-only、reference-only 的 health readiness boundary。它不是 backend health runtime，不是云 secret backend client，不是 production resolver output，不是 credential payload，也不是可被重放为 secret 解析请求的结构化 payload。

允许作为 future health boundary 输入或 metadata 暴露的字段只限：

| 字段 | 要求 |
| --- | --- |
| `backend_health_ref` | opaque reference，不可从 backend URL、secret ref 或 operator claim 推导 |
| `backend_profile_ref` | 只能引用 backend profile selection readiness 中的静态 profile |
| `backend_profile_id` | 稳定短键或 reference id，不包含 provider raw URL |
| `backend_kind` | 必须来自 reserved backend kind allowlist |
| `environment` | 必须与 secret ref、provider profile、backend profile 绑定一致 |
| `provider_profile_id` | 只能引用 provider profile binding evidence |
| `secret_ref_key_status` | 只描述 reference key 是否存在，不输出 full secret ref |
| `secret_ref_version_ref_status` | 只描述版本引用状态，不输出 secret value |
| `health_policy_version` | 只允许策略版本 |
| `health_lifecycle_state` | 必须来自固定 lifecycle allowlist |
| `last_known_health_ref` | 只允许 reference id，不表示本批执行了 health check |
| `sanitized_backend_diagnostic` | 只允许脱敏摘要 |
| `failure_code` / `failure_boundary` | 只允许 fail-closed failure metadata |
| `request_id` / `audit_ref` | 只允许引用，不写 audit store |
| `policy_version` | 只允许策略版本 |

禁止任何 future health boundary、fixture、diagnostic 或文档输出：

- secret material、secret value、password、token、API key、cloud credential
- credential payload、full credential handle、credential payload hash 原文
- provider raw URL、resolver backend URL、backend endpoint URL、DSN、database hostname
- full secret ref value、full operator identity claim、authorization header、cookie、完整用户 claim
- raw health request、raw health response、raw provider error、raw backend error detail
- 可解码、可反序列化或可拼接成 credential / endpoint / DSN 的结构化 blob

## Binding Preconditions

future backend health boundary 必须在任何 runtime health check、resolver runtime 或 production resolver task card 创建前满足以下绑定：

| binding | 要求 |
| --- | --- |
| backend profile binding | 必须引用 resolver backend profile selection readiness |
| environment binding | test / production / local 不允许互相 fallback |
| provider profile binding | provider profile 必须与 provider profile secret binding evidence 对齐 |
| secret ref binding | 只能记录 secret ref key status 与 version reference status |
| operator approval dependency | 必须引用 operator approval runtime evidence readiness |
| audit handoff dependency | 必须引用 audit store handoff readiness，不写 audit store |
| rotation policy dependency | 必须引用 rotation / audit policy readiness |
| no leakage strategy dependency | 必须引用 no leakage smoke runtime strategy，runtime 未创建时不能声明 production ready |
| credential handle boundary dependency | 必须引用 credential handle boundary readiness，不创建 handle runtime |
| health policy dependency | 必须显式绑定 health policy version；缺失时 fail closed |

## Health Lifecycle

本批只固定 future lifecycle 名称，不创建 runtime 状态机：

| lifecycle state | 含义 |
| --- | --- |
| `boundary_planned` | 仅完成 backend health boundary 规划 |
| `metadata_bound` | backend profile / provider profile / environment metadata 已绑定 |
| `health_reference_prepared` | 已形成 reference-only health reference |
| `runtime_not_created` | backend health runtime 尚未创建 |
| `health_check_not_executed` | 本批未执行 health check |
| `health_status_unknown` | health 只能保持 unknown / reference-only |
| `health_status_available_metadata_only` | future runtime 只能输出脱敏 available metadata |
| `health_status_degraded_metadata_only` | future runtime 只能输出脱敏 degraded metadata |
| `health_status_unavailable_fail_closed` | unavailable 必须 fail closed |
| `diagnostics_sanitized` | 所有诊断必须脱敏 |
| `runtime_implementation_deferred` | runtime implementation 后续单独拆分 |

允许的 future health status 仅为：

- `unknown_reference_only`
- `available_metadata_only`
- `degraded_metadata_only`
- `unavailable_fail_closed`
- `blocked_dependency_missing`
- `blocked_environment_mismatch`
- `blocked_policy_missing`
- `blocked_runtime_not_created`

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| backend health reference | `defined_static_only` | 只定义 reference，不执行 health check |
| backend profile binding | `required_before_runtime` | 缺失时 fail closed |
| environment binding | `required_no_cross_environment` | 禁止跨环境 fallback |
| provider profile binding | `required_before_runtime` | 只引用 provider profile evidence |
| secret ref binding | `required_reference_only` | 不输出 full secret ref |
| health metadata allowlist | `fixed_sanitized_only` | 只允许脱敏 metadata |
| health lifecycle status allowlist | `fixed_static_only` | 只固定 future lifecycle/status |
| failure mapping | `fail_closed` | 所有失败 fail closed |
| sanitized diagnostics | `fixed_sanitized_only` | 禁止 raw backend detail |
| fallback policy | `forbidden` | 禁止 fallback 到 fake / env / mock / sample |
| side effect policy | `no_runtime_calls` | checker 只读 committed artifact |
| artifact guard | `enforced` | 禁止 runtime artifact |
| backend health runtime | `not_created` | 本批不创建 runtime |
| backend health check execution | `not_executed` | 本批不探测 backend |
| production resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| DB / repository / API | `blocked` | 不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `resolver_backend_health_boundary_missing` | `backend_health` | 缺少 backend health boundary evidence |
| `resolver_backend_health_reference_missing` | `backend_health` | 缺少 `backend_health_ref` |
| `resolver_backend_health_profile_binding_missing` | `backend_profile` | 缺少 backend profile reference |
| `resolver_backend_health_environment_mismatch` | `environment_binding` | health boundary 与 environment binding 不一致 |
| `resolver_backend_health_provider_profile_mismatch` | `provider_profile` | health boundary 与 provider profile binding 不一致 |
| `resolver_backend_health_secret_ref_binding_missing` | `secret_ref` | 缺少 secret ref key status / version reference status |
| `resolver_backend_health_policy_missing` | `policy_gate` | 缺少 health policy version |
| `resolver_backend_health_metadata_forbidden` | `metadata_contract` | 输出未允许的 health metadata |
| `resolver_backend_health_secret_material_detected` | `artifact_guard` | 出现 secret value、token、cloud credential 或 credential payload |
| `resolver_backend_health_raw_endpoint_detected` | `artifact_guard` | 出现 provider raw URL、resolver backend URL、DSN 或 backend endpoint URL |
| `resolver_backend_health_runtime_created_forbidden` | `artifact_guard` | 本批创建 backend health runtime |
| `resolver_backend_health_check_executed_forbidden` | `no_side_effects` | 本批执行 backend health check、provider call 或 cloud call |
| `resolver_backend_health_fallback_forbidden` | `no_fallback` | health boundary 缺失时 fallback 到 fake resolver、env、mock、sample 或 historical evidence |
| `resolver_backend_health_side_effect_forbidden` | `no_side_effects` | 本批读取 secret、连接 DB、写 audit store 或调用 API |
| `resolver_backend_health_repository_mode_forbidden` | `repository_mode` | 本批启用 repository mode |
| `resolver_backend_health_scope_overreach` | `implementation_boundary` | 本批合入 backend runtime、production resolver runtime、DB provider、audit writer 或 public API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、backend endpoint URL、完整 secret ref value、完整 credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw health request、raw health response 或 raw backend error detail。

## Sanitized Diagnostics

允许输出：

- `health_boundary_status`
- `health_runtime_status`
- `health_check_status`
- `backend_profile_ref_status`
- `backend_kind`
- `environment`
- `provider_profile_id`
- `secret_ref_key_status`
- `secret_ref_version_ref_status`
- `health_policy_version`
- `health_lifecycle_state`
- `last_known_health_ref_status`
- `failure_code`
- `failure_boundary`
- `sanitized_backend_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、backend endpoint URL、DSN、cloud credential、database hostname、database error detail、credential payload、full secret ref value、full credential handle、完整 operator claim、完整用户 claim、authorization header、cookie、raw health request、raw health response 或 raw backend error detail。

## No Fallback

- backend health boundary 缺失时真实 resolver runtime implementation task card 必须继续作为后续独立切片处理，不能在本批创建。
- 不允许 health boundary fallback 到 fake resolver runtime、developer env plaintext、fixture credential、committed value、sample、mock provider、local-smoke profile、previously approved test evidence、operator runbook 文本、repository memory store 或 console local record。
- 不允许 production health boundary fallback 到 test health boundary，也不允许 test health boundary fallback 到 production health boundary。
- 不允许缺少 backend profile / provider profile / environment / health policy 时创建健康成功。
- 不把 health boundary readiness 写成 backend runtime ready、backend health check executed、credential resolved、production resolver runtime ready、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不执行 backend health check、不执行 smoke runtime、不创建 credential payload、不创建 credential handle、不执行 approval runtime、不创建 audit store、不写 audit event、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `backend_health_check_count=0`
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
- `database_connection_count=0`
- `driver_open_count=0`
- `sql_execution_count=0`
- `schema_marker_read_count=0`
- `schema_marker_write_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-resolver-backend-health-boundary-readiness-v1.md`
- `docs/task-cards/production-secret-backend-resolver-backend-health-boundary-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-resolver-backend-health-boundary-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py`

不得新增或启用以下 artifact：

- backend health runtime
- backend health client
- backend health checker runtime
- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- cloud secret SDK
- secret value fixture
- production credential file
- credential payload runtime
- credential handle runtime
- approval runtime / approval executor
- no secret leakage smoke runtime
- audit store runtime
- audit writer
- database connection provider
- DB driver / DSN parser
- connection factory
- SQL migration
- schema marker reader / writer
- migration runner
- workflow saved draft repository mode runtime
- public production API

## 后续 Runtime 拆分

本批完成后，只能说明 backend health boundary 的静态输入、metadata、状态、失败语义、diagnostics 和 artifact guard 已定义。后续 `resolver-backend-health-runtime-implementation-entry-review` 已单独完成并固定为 `resolver_backend_health_runtime_implementation_entry_review_defined`，entry decision 为 blocked before runtime task card。其余 runtime 仍必须拆成独立目标推进：

1. `resolver-backend-health-runtime-implementation`
2. `production-secret-backend-real-resolver-runtime-implementation-entry-refresh`
3. `production-secret-backend-real-resolver-runtime-implementation`

任何后续 runtime 目标都必须继续满足 no secret leakage、credential handle、operator approval、audit handoff、rotation policy、health boundary、no fallback 和 no side effects 的证据链；不得用 fake resolver、developer env、mock provider、local-smoke profile、DB provider、repository memory store 或 audit memory store 替代。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-health-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-audit-store-handoff-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-evidence-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-boundary-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-no-secret-leakage-smoke-runtime-strategy-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
