# Production Secret Backend Resolver Backend Profile Selection Readiness v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 真实 resolver runtime 之前的 resolver backend profile selection 前置边界。

对应切片：`production-secret-backend-resolver-backend-profile-selection-readiness-v1`。

结论：状态为 `resolver_backend_profile_selection_readiness_defined`。本批只定义 backend profile 的静态选择条件、provider profile / environment binding、policy version、operator approval、audit / rotation dependency、backend health reference、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不创建 backend runtime，不选择云厂商 SDK，不读取真实 secret，不创建 credential payload 或 credential handle runtime，不连接数据库，不启用 repository mode，也不创建 production resolver runtime implementation task card。

## 选择边界

backend profile selection 只允许引用静态 profile 元数据，不允许执行解析或探测：

| 字段 | 要求 |
| --- | --- |
| `backend_profile_id` | 必须存在，只能作为稳定短键或引用 ID |
| `backend_kind` | 必须属于 reserved allowlist；本批不绑定 AWS / GCP / Azure / Vault 等具体 SDK |
| `environment` | 必须显式绑定，禁止 test / production 互相 fallback |
| `provider_profile_id` | 必须与 provider profile secret binding 前置证据一致 |
| `policy_version` | 必须显式存在；缺失时 fail closed |
| `secret_ref_namespace` | 只能绑定 reference-only namespace，不包含 secret value |
| `allowed_secret_ref_kinds` | 只允许 placeholder / managed reference 类型，不允许 plaintext |
| `operator_approval_required` | production profile 必须保留 operator approval dependency |
| `audit_policy_ref` / `rotation_policy_ref` | 必须引用已有 policy evidence，不创建 runtime store |
| `health_boundary_status` | 只能记录 sanitized health boundary reference；本批不调用 backend health check |

允许的 backend kind 仅为未来实现保留名：

- `external_secret_manager_reserved`
- `operator_managed_secret_store_reserved`

禁止使用的 backend kind：

- `committed_secret_file`
- `env_plaintext`
- `fake_resolver`
- `mock_provider`
- `local_smoke_profile`
- `repository_memory_store`
- `database_dsn_resolver`

## Gate Matrix

| gate | 状态 | 说明 |
| --- | --- | --- |
| backend kind allowlist | `fixed_reference_only` | 只固定 reserved backend kind，不接具体 SDK |
| backend profile id | `required_before_runtime` | 没有 profile id 时不能进入 runtime |
| environment binding | `required_no_cross_environment` | 禁止跨 test / production fallback |
| provider profile binding | `required_reference_only` | 只消费 provider profile readiness，不解析 credential |
| secret ref namespace binding | `required_reference_only` | 只允许 reference-only secret ref namespace |
| policy version | `required_before_runtime` | 缺失时 fail closed |
| operator approval dependency | `required_before_runtime` | production profile 仍需 operator approval evidence |
| audit policy dependency | `required_policy_defined` | 只引用 policy，不写 audit store |
| rotation policy dependency | `required_policy_defined` | 只引用 policy，不执行 rotation |
| backend health boundary | `reference_only_blocked` | 只记录 future health boundary；不调用 backend |
| resolver runtime | `not_created` | 本批不创建 runtime |
| cloud secret service | `forbidden` | 不调用云 secret 服务 |
| DB / repository / API | `blocked` | 不连接 DB、不启用 repository mode、不创建 public production API |

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `resolver_backend_profile_selection_missing` | `backend_profile` | 未提供 backend profile selection evidence |
| `resolver_backend_profile_kind_unsupported` | `backend_kind` | backend kind 不在 reserved allowlist |
| `resolver_backend_profile_id_missing` | `backend_profile` | 缺少 `backend_profile_id` |
| `resolver_backend_profile_environment_mismatch` | `environment_binding` | profile environment 与请求或 provider profile 不一致 |
| `resolver_backend_profile_provider_mismatch` | `provider_profile_binding` | provider profile 不匹配 |
| `resolver_backend_profile_secret_ref_namespace_missing` | `secret_ref` | 缺少 reference-only secret ref namespace |
| `resolver_backend_profile_policy_version_missing` | `policy_gate` | 缺少 policy version |
| `resolver_backend_profile_health_boundary_missing` | `backend_health` | 缺少 backend health boundary reference |
| `resolver_backend_profile_operator_approval_missing` | `operator_gate` | production profile 缺少 operator approval dependency |
| `resolver_backend_profile_audit_policy_missing` | `audit_policy` | 缺少 audit policy reference |
| `resolver_backend_profile_rotation_policy_missing` | `rotation_policy` | 缺少 rotation policy reference |
| `resolver_backend_profile_secret_value_detected` | `artifact_guard` | 文档、fixture 或 checker 出现 secret-looking value |
| `resolver_backend_profile_cloud_sdk_forbidden` | `artifact_guard` | 本批出现云 SDK、cloud secret client 或 provider runtime |
| `resolver_backend_profile_runtime_created_forbidden` | `artifact_guard` | 本批创建 backend runtime 或 production resolver runtime |
| `resolver_backend_profile_repository_mode_forbidden` | `repository_mode` | 本批启用 repository mode |
| `resolver_backend_profile_scope_overreach` | `implementation_boundary` | 本批合入 DB provider、credential handle、audit store 或 production API |

所有失败必须 fail closed，不返回 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、完整 secret ref value、完整 credential handle、完整用户 claim、authorization header 或 cookie。

## Sanitized Diagnostics

允许输出：

- `resolver_backend_profile_selection_status`
- `backend_profile_id`
- `backend_kind`
- `environment`
- `provider_profile_id`
- `policy_version`
- `secret_ref_namespace_status`
- `health_boundary_status`
- `operator_approval_status`
- `audit_policy_status`
- `rotation_policy_status`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`

禁止输出 raw secret、secret value、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## No Fallback

- backend profile 缺失或不匹配时必须 fail closed。
- 不允许从 production backend profile fallback 到 test profile、fake resolver、mock provider、local-smoke profile、developer env plaintext、fixture credential、committed value、sample 或 repository memory store。
- 不允许 fake resolver profile 被提升成 production backend profile。
- 不允许 provider profile binding 缺失时 fallback 到 `RADISHMIND_PLATFORM_API_KEY` 或其它环境变量 credential。
- 不把 backend profile selection readiness 写成 production resolver runtime ready、credential resolved、production secret backend ready 或 public production API ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序，不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 resolver、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
- `backend_health_check_count=0`
- `credential_payload_created_count=0`
- `credential_handle_created_count=0`
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

- `docs/platform/production-secret-backend-resolver-backend-profile-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-resolver-backend-profile-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-resolver-backend-profile-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py`

不得新增或启用以下 artifact：

- production resolver runtime
- production resolver implementation task card
- production resolver backend client
- backend profile runtime
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

本批完成后，真实 resolver runtime implementation task card 仍不能创建。下一步应继续从剩余 blocker 中选择一个单独推进：

1. `no-secret-leakage-smoke-runtime-strategy`
2. `credential-handle-runtime-boundary-readiness`
3. `operator-approval-runtime-evidence-readiness`
4. `production-secret-audit-store-handoff-readiness`
5. `resolver-backend-health-boundary-readiness`

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-implementation-entry-review-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
