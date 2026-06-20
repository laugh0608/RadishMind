# Production Secret Backend Real Resolver Runtime Preconditions v1

更新时间：2026-06-20

## 文档目的

本文档固定 production secret backend 进入真实 resolver runtime 之前必须满足的前置条件和停止线。

对应切片：`production-secret-backend-real-resolver-runtime-preconditions-v1`。

结论：状态为 `real_resolver_runtime_preconditions_defined`。本批只定义真实 resolver runtime 的启用条件、secret ref / provider profile / environment binding 前置、operator approval、audit / rotation 依赖、failure mapping、sanitized diagnostics、no fallback、no side effects、artifact guard、验证方式和后续实现拆分。不实现 production resolver runtime，不读取真实 secret，不调用云 secret 服务，不连接数据库，不创建 credential payload，不启用 repository mode，不新增 public production API。

## 输入证据

- `production-secret-backend-fake-resolver-runtime-implementation-v1` 已固定 `fake_resolver_runtime_test_only_implemented`，但只服务离线测试边界。
- `production-secret-backend-secret-resolver-interface-disabled-readiness-v1` 已固定 reference-only resolver input、disabled result、failure mapping 和 no fallback。
- `production-secret-backend-provider-profile-secret-binding-readiness-v1` 已固定 provider/profile 到 `secret_ref` 的 reference-only 绑定。
- `production-secret-backend-operator-runbook-negative-gates-readiness-v1` 已固定 operator approval evidence、negative gates 和 sanitized verification。
- `production-secret-backend-rotation-audit-policy-readiness-v1` 已固定 rotation trigger、audit event fields、secret ref version reference 和 rollback policy。
- `production-secret-reference-basic` 仍是 reference-only fixture，不包含 secret value，不代表 credential resolved。

## 当前状态

状态：`real_resolver_runtime_preconditions_defined`

结论：

- 真实 resolver runtime 的 preconditions 已定义为可检查证据。
- `production_secret_backend` 仍为 `not_satisfied`。
- `resolver_implementation_status` 仍为 `not_started`。
- `resolver_runtime_status` 仍为 `not_created`。
- `fake_resolver_runtime_status` 仍为 `implemented_test_only_disabled_by_default`。
- 本批不创建 resolver runtime、backend client、credential payload、credential handle runtime、audit store、DB provider 或 public production API。

## 启用条件

后续任何真实 resolver runtime implementation task card 打开前，必须同时满足以下条件：

1. 已有独立 implementation entry review，明确本次只打开 resolver runtime，不并行打开 DB provider、repository mode、schema marker、audit store 或 public production API。
2. `secret_ref` 只来自 reference-only manifest 或等价配置层，不允许从环境变量、`.env.example`、fixture、task card 或文档读取 secret value。
3. provider/profile binding 已声明 `credential_requirement=required`、`secret_ref_status=present`、`environment`、`provider`、`provider_profile` 和短键 profile。
4. environment binding 必须显式区分 `test` 与 `production`，禁止跨环境 secret ref fallback。
5. operator approval evidence 必须包含 `operator_id_ref`、`approval_ticket_ref`、`change_window_ref`、`runbook_version` 和 `policy_version`。
6. audit / rotation policy 必须包含 `audit_ref`、`secret_ref_version_ref`、`rotation_policy_version`、rollback / disable policy 和 sanitized verification。
7. resolver backend profile 必须通过单独任务选择并记录 backend kind、allowed environment、policy version 和 sanitized health boundary；本批不选择云厂商 SDK。
8. no secret leakage smoke 必须在实现任务中以离线方式验证输出、diagnostics、audit metadata 和 failure result 不含 secret-looking value。
9. artifact guard 必须确认 production resolver runtime、cloud client、credential payload、DB provider、repository mode 和 public API 没有提前出现。

## Resolver Contract Preconditions

真实 resolver runtime 的输入只能是 reference-only 和审计字段。允许进入 future runtime input contract 的字段为：

- `environment`
- `provider`
- `provider_profile`
- `credential_requirement`
- `secret_ref_key`
- `secret_ref_version_ref`
- `caller_purpose`
- `request_id`
- `audit_ref`
- `operator_approval_ref`
- `policy_version`
- `rotation_policy_version`
- `backend_profile_ref`

future runtime result 只能返回脱敏状态和 opaque handle metadata，不得返回 secret value 或 credential payload。允许字段为：

- `resolver_state`
- `credential_handle_present`
- `credential_handle_id`
- `credential_kind`
- `environment`
- `provider`
- `provider_profile`
- `secret_ref_version_ref`
- `request_id`
- `audit_ref`
- `policy_version`
- `failure_code`
- `sanitized_diagnostic`

`credential_handle_id` 只能是 opaque reference，不能编码 raw secret、DSN、provider raw URL、database hostname、token、API key、cloud credential 或完整 secret ref value。

## Failure Mapping

所有失败必须 fail closed，不能 fallback 到 developer env、fake resolver、mock provider、local-smoke profile、fixture credential、committed value、DB provider、repository mode 或 sample。

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `real_resolver_runtime_preconditions_missing` | `precondition_gate` | 缺少本专题定义的任一 mandatory precondition |
| `real_resolver_runtime_task_not_approved` | `implementation_gate` | 缺少独立 implementation entry review 或 task card |
| `real_resolver_backend_profile_missing` | `backend_profile` | 缺少 resolver backend profile 选择与 policy version |
| `real_resolver_secret_ref_missing` | `secret_ref` | provider/profile 需要 credential 但缺少 secret ref |
| `real_resolver_provider_profile_binding_missing` | `provider_profile_binding` | 缺少 provider/profile binding 或 credential requirement |
| `real_resolver_environment_binding_missing` | `environment_binding` | 缺少 environment binding 或 test / production 隔离 |
| `real_resolver_operator_approval_missing` | `operator_gate` | 缺少 operator approval evidence |
| `real_resolver_audit_policy_missing` | `audit_policy` | 缺少 audit policy、audit ref 或 audit event field allowlist |
| `real_resolver_rotation_policy_missing` | `rotation_policy` | 缺少 rotation trigger、secret ref version reference 或 rollback policy |
| `real_resolver_no_leakage_gate_missing` | `artifact_guard` | 缺少 no secret leakage smoke strategy / runtime validation plan |
| `real_resolver_backend_unavailable` | `backend_profile` | future backend profile 不可用或未启用 |
| `real_resolver_resolution_denied` | `policy_gate` | caller purpose、audit ref、environment 或 policy version 不允许解析 |
| `real_resolver_secret_value_exposure_detected` | `artifact_guard` | 输入、输出、diagnostics 或 audit metadata 出现 secret-looking value |
| `real_resolver_fallback_forbidden` | `no_fallback` | 出现 env / fake / mock / fixture / sample fallback |
| `real_resolver_side_effect_forbidden` | `no_side_effects` | precondition 阶段出现 secret read、cloud call、DB connection、audit store write 或 API call |
| `real_resolver_artifact_guard_violation` | `artifact_guard` | 提前新增 production resolver、cloud client、DB provider、repository mode 或 public API artifact |

## Sanitized Diagnostics

允许输出：

- `real_resolver_runtime_preconditions_status`
- `resolver_state`
- `resolver_backend_profile_status`
- `secret_ref_status`
- `secret_ref_version_status`
- `provider_profile_binding_status`
- `environment_binding_status`
- `operator_approval_status`
- `audit_policy_status`
- `rotation_policy_status`
- `no_secret_leakage_gate_status`
- `credential_handle_present`
- `failure_code`
- `sanitized_diagnostic`
- `request_id`
- `audit_ref`
- `policy_version`

禁止输出 raw secret、password、token、API key、provider raw URL、DSN、cloud credential、database hostname、database error detail、credential payload、resolver backend URL、full secret ref value、full credential handle、完整用户 claim、authorization header 或 cookie。

## No Fallback

- 不允许 production resolver fallback 到 fake resolver。
- 不允许 fake resolver fallback 到 production resolver。
- 不允许 resolver failure fallback 到 `RADISHMIND_PLATFORM_API_KEY`、developer env plaintext、fixture credential、committed value、mock provider、local-smoke profile、fake query executor、sample 或 repository memory store。
- 不允许 test `secret_ref` fallback 到 production，也不允许 production fallback 到 test。
- 不把 `secret_ref_status=present`、operator approval、audit ref、rotation policy 或 backend profile 写成 credential resolved。
- 不把 test-only fake resolver runtime 写成 production resolver runtime。

## No Side Effects

本批只写入文档、fixture 和 checker。checker 只能读取 committed 文档、fixture 和源码文本，不读取真实环境 secret，不连接网络，不调用云 secret 服务，不访问 provider，不创建 credential payload，不连接数据库，不打开 driver，不执行 SQL，不读写 schema marker，不启用 repository mode，不写 audit store，不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `cloud_secret_call_count=0`
- `provider_call_count=0`
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

本批允许新增：

- `docs/platform/production-secret-backend-real-resolver-runtime-preconditions-v1.md`
- `docs/task-cards/production-secret-backend-real-resolver-runtime-preconditions-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-real-resolver-runtime-preconditions-v1.json`
- `scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py`

不得新增或启用：

- production resolver runtime
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

## 后续实现拆分

后续若继续 production secret backend，应按以下顺序选择单一方向：

1. `real-resolver-runtime-implementation-entry-review` 已完成，结论仍为 blocked before runtime task card。
2. `resolver-backend-profile-selection-readiness` 已完成静态前置证据，但 backend runtime 仍未创建。
3. `real-resolver-no-secret-leakage-smoke-runtime-strategy` 已完成静态前置证据，但 no secret leakage smoke runtime 仍未创建。
4. 下一步应从 `credential-handle-runtime-boundary-readiness`、`operator-approval-runtime-evidence-readiness`、`production-secret-audit-store-handoff-readiness` 或 `resolver-backend-health-boundary-readiness` 中选择单一方向。
5. `real-resolver-runtime-implementation`、`database-connection-provider-entry-review` 或 `schema-marker-contract` 只能在上述 blocker 继续收敛后再评审；不得并行打开 DB provider、repository mode 或 public API。

## 验证

建议验证顺序：

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-real-resolver-runtime-preconditions-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-fake-resolver-runtime-implementation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-secret-resolver-interface-disabled-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-runbook-negative-gates-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-rotation-audit-policy-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-contract.py
./scripts/run-python.sh scripts/check-production-secret-reference-contract.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-readiness-v1.py
./scripts/run-python.sh scripts/checks/control_plane/check-workflow-saved-draft-database-secret-resolver-implementation-entry-review-v1.py
./scripts/check-repo.sh --fast
./scripts/check-repo.sh
```
