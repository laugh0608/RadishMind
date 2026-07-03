# Production Secret Backend Cloud Secret Service Selection Readiness v1

更新时间：2026-06-27

## 文档目的

本文档固定 production secret backend 在进入真实 resolver runtime 前，对 future cloud secret service selection 的静态准入要求。

对应切片：`production-secret-backend-cloud-secret-service-selection-readiness-v1`。

结论：状态为 `cloud_secret_service_selection_readiness_defined`。本批只定义未来选择 cloud secret service 或 operator-managed secret store 时必须满足的候选范围、profile manifest、provider / environment / secret ref namespace binding、policy version、operator approval、credential handle、audit store、backend health、no leakage smoke、failure mapping、sanitized diagnostics、no fallback、no side effects 和 artifact guard；不选择具体云厂商，不绑定云 SDK，不创建 cloud secret client，不调用云 secret 服务，不读取真实 secret，不创建 production resolver runtime task card，不创建 production resolver runtime、credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、DB provider、repository mode runtime 或 public production API。

selection decision：`cloud_secret_service_selection_deferred_until_runtime_prerequisites`。

## 输入证据

- `production_resolver_runtime_blocker_consolidation_defined` 已确认 cloud secret service selection 仍是 production resolver runtime blocker。
- `resolver_backend_profile_selection_readiness_defined` 已固定 reserved backend kind、backend profile id、environment / provider profile / policy version / secret ref namespace binding，但不选择云厂商 SDK。
- `credential_handle_runtime_implementation_entry_refresh_defined` 与 `operator_approval_runtime_implementation_entry_refresh_defined` 已确认 handle / approval runtime task card 仍 blocked。
- `audit_store_runtime_implementation_entry_refresh_v3_defined`、`resolver_backend_health_runtime_implementation_entry_review_defined` 与 `real_resolver_no_secret_leakage_smoke_runtime_implementation_entry_review_defined` 已确认 audit store、backend health 和 no leakage smoke runtime 仍 blocked。
- `implementation_readiness_defined` 与 `satisfied_reference_only_resolver_disabled` 仍要求 reference-only secret manifest、offline fast baseline、no cloud call 和 no real credential。

## Selection Boundary

cloud secret service selection readiness 只允许定义选择准入，不允许执行选择或调用：

| gate | 状态 | 说明 |
| --- | --- | --- |
| candidate backend kind | `reserved_only` | 只允许 future reserved kind，不绑定 AWS / GCP / Azure / Vault 等 SDK |
| concrete cloud vendor | `not_selected` | 本批不选择具体云厂商或托管服务 |
| provider profile binding | `required_reference_only` | 只消费 provider profile readiness，不解析 credential |
| environment binding | `required_no_cross_environment` | 禁止 test / production 互相 fallback |
| secret ref namespace | `required_reference_only` | 只允许 reference-only namespace，不保存 secret value |
| policy version | `required_before_runtime` | 缺失时 fail closed |
| credential handle runtime | `blocked_before_runtime_task_card` | 不能生成或传递 credential handle |
| operator approval runtime | `blocked_before_runtime_task_card` | 不能执行 approval、ticket 或 change window gate |
| audit store runtime | `blocked_before_runtime_task_card` | 不能写 audit event、delivery result 或 idempotency record |
| backend health runtime | `blocked_before_runtime_task_card` | 不能执行 health check 或 provider call |
| no leakage smoke runtime | `blocked_before_runtime_task_card` | 不能声明 runtime leakage gate 已执行 |
| production resolver runtime | `not_created` | 不创建 resolver runtime task card 或 runtime |
| repository mode runtime | `disabled` | 不能打开 Saved Workflow Draft repository mode |

## Candidate Shape

未来 selection task card 必须只消费 metadata-only manifest：

- `candidate_kind`
- `selection_policy_version`
- `environment`
- `provider_profile_id`
- `secret_ref_namespace`
- `allowed_secret_ref_kinds`
- `operator_approval_dependency`
- `credential_handle_dependency`
- `audit_store_dependency`
- `backend_health_dependency`
- `no_leakage_smoke_dependency`
- `rotation_policy_ref`

禁止 manifest、fixture、docs 或 diagnostics 中出现 raw secret、secret value、password、token、API key、provider raw URL、resolver backend URL、DSN、cloud credential、database hostname、database error detail、credential payload、完整 secret ref value、完整 credential handle、authorization header、cookie、raw claim dump、JWKS raw dump、membership raw record 或 provider error detail。

## Failure Mapping

| code | failure boundary | 触发条件 |
| --- | --- | --- |
| `cloud_secret_service_selection_dependency_missing` | `dependency_chain` | 必需 readiness / refresh 证据缺失 |
| `cloud_secret_service_selection_vendor_selected_forbidden` | `selection_gate` | 本批选择了具体云厂商或托管服务 |
| `cloud_secret_service_selection_sdk_binding_forbidden` | `artifact_guard` | 本批绑定云 SDK 或 provider client |
| `cloud_secret_service_selection_client_created_forbidden` | `artifact_guard` | 本批创建 cloud secret client |
| `cloud_secret_service_selection_call_forbidden` | `side_effect_guard` | 本批调用 cloud secret service 或 provider |
| `cloud_secret_service_selection_secret_material_detected` | `artifact_guard` | 静态 artifact 出现 secret-bearing material |
| `cloud_secret_service_selection_credential_handle_blocked` | `credential_handle` | credential handle runtime 仍 blocked |
| `cloud_secret_service_selection_operator_approval_blocked` | `operator_approval` | operator approval runtime 仍 blocked |
| `cloud_secret_service_selection_audit_store_blocked` | `audit_store` | audit store runtime 仍 blocked |
| `cloud_secret_service_selection_backend_health_blocked` | `backend_health` | backend health runtime 仍 blocked |
| `cloud_secret_service_selection_no_leakage_blocked` | `no_secret_leakage` | no leakage smoke runtime 仍 blocked |
| `cloud_secret_service_selection_repository_mode_forbidden` | `workflow_durable_store` | 本批启用 repository mode |
| `cloud_secret_service_selection_scope_overreach` | `implementation_boundary` | 本批合入 runtime、DB、API 或 executor 能力 |

所有失败必须 fail closed，只返回 metadata-only diagnostics。

## No Fallback

- 不允许把 fake resolver、test profile、local smoke profile、developer env、repository memory store 或 committed fixture 提升为 production cloud secret service。
- 不允许从 production backend profile fallback 到 test backend profile。
- 不允许 provider profile binding 缺失时 fallback 到环境变量 credential。
- 不允许在 credential handle、operator approval、audit store、backend health 或 no leakage smoke runtime 仍 blocked 时创建 production resolver runtime task card。
- 不把 `cloud_secret_service_selection_readiness_defined` 写成 cloud secret backend ready、production resolver ready、credential resolved、repository mode ready 或 production ready。

## No Side Effects

本批 checker 只读取 committed 文档、fixture、schema、已有 readiness fixture 和 `check-repo.py` 注册顺序；不读取真实环境 secret、不连接网络、不调用云 secret 服务、不访问 provider、不调用 fake resolver、不调用 production resolver、不创建 cloud secret client、不创建 credential payload、不创建 credential handle、不连接数据库、不打开 driver、不运行 SQL、不读写 schema marker、不启用 repository mode、不写 audit store、不调用 production API。

side effect counters 必须保持：

- `real_secret_read_count=0`
- `environment_secret_read_count=0`
- `secret_resolver_call_count=0`
- `fake_resolver_call_count=0`
- `production_resolver_call_count=0`
- `cloud_secret_vendor_selected_count=0`
- `cloud_secret_sdk_import_count=0`
- `cloud_secret_client_created_count=0`
- `cloud_secret_call_count=0`
- `network_call_count=0`
- `provider_call_count=0`
- `credential_handle_runtime_created_count=0`
- `credential_handle_created_count=0`
- `operator_approval_runtime_execution_count=0`
- `audit_store_runtime_created_count=0`
- `backend_health_runtime_created_count=0`
- `no_secret_leakage_smoke_runtime_created_count=0`
- `database_connection_count=0`
- `sql_execution_count=0`
- `repository_mode_enablement_count=0`
- `production_api_call_count=0`

## Artifact Guard

本批只允许新增以下静态证据：

- `docs/platform/production-secret-backend-cloud-secret-service-selection-readiness-v1.md`
- `docs/task-cards/production-secret-backend-cloud-secret-service-selection-readiness-v1-plan.md`
- `scripts/checks/fixtures/production-secret-backend-cloud-secret-service-selection-readiness-v1.json`
- `scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py`

不得新增或启用 production resolver runtime、production resolver implementation task card、cloud secret SDK、cloud secret client、provider SDK binding、secret value fixture、production credential file、credential payload runtime、credential handle runtime、operator approval runtime、audit store runtime、backend health runtime、no leakage smoke runtime、database connection provider、DB driver / DSN parser、connection factory、SQL migration、schema marker reader / writer、migration runner、workflow saved draft repository mode runtime 或 public production API。

## 后续推进

本批完成后，cloud secret service concrete provider selection 仍必须作为后续独立任务评审。后续若继续 production resolver runtime task card，还必须先证明 credential handle runtime、operator approval runtime、audit store runtime、backend health runtime 和 no leakage smoke runtime 各自的 implementation task card 已独立评审并不再 blocked。

## 验证

```bash
./scripts/run-python.sh scripts/check-production-ops-secret-backend-cloud-secret-service-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-production-resolver-runtime-blocker-consolidation-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-resolver-backend-profile-selection-readiness-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-credential-handle-runtime-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-operator-approval-runtime-implementation-entry-refresh-v1.py
./scripts/run-python.sh scripts/check-production-ops-secret-backend-implementation-readiness.py
git diff --check
./scripts/check-repo.sh --fast
```

## 停止线

- 不选择具体云厂商、不绑定云 SDK、不创建 cloud secret client、不调用云 secret 服务。
- 不创建 production resolver runtime implementation task card。
- 不实现 production resolver runtime、production resolver backend client、credential payload、credential handle runtime、operator approval runtime、approval executor、audit store runtime、audit writer、backend health runtime、no leakage smoke runtime、database connection provider、repository mode runtime、production API、executor、confirmation、writeback 或 replay。
- 不读取真实 secret，不读取环境 secret，不访问 provider，不执行 approval / health / smoke runtime，不连接数据库，不运行 SQL，不读写 schema marker。
- 不把 `cloud_secret_service_selection_readiness_defined` 写成 cloud secret service ready、production resolver ready、secret backend ready、durable persistence ready、repository mode ready、database ready、production API ready 或 production ready。
